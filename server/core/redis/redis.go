// Package redis handles our remote database layer.
package redis

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

var ctx = context.Background()

// Redis is a database using Redis 6
type RedisDB struct { // nolint: revive
	client       *redis.Client
	queueEnabled bool
}

const (
	pullKeySeparator = "::"
)

func New(hostname string, port int, password string, tlsEnabled bool, insecureSkipVerify bool, db int, queueEnabled bool) (*RedisDB, error) {
	var rdb *redis.Client

	var tlsConfig *tls.Config
	if tlsEnabled {
		tlsConfig = &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: insecureSkipVerify, //nolint:gosec // In some cases, users may want to use this at their own caution
		}
	}

	rdb = redis.NewClient(&redis.Options{
		Addr:      fmt.Sprintf("%s:%d", hostname, port),
		Password:  password,
		DB:        db,
		TLSConfig: tlsConfig,
	})

	// Check if connection is valid
	err := rdb.Ping(ctx).Err()
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to connect to redis instance at %s:%d", hostname, port))
	}

	return &RedisDB{
		client:       rdb,
		queueEnabled: queueEnabled,
	}, nil
}

// NewWithClient is used for testing.
func NewWithClient(client *redis.Client, bucket string, globalBucket string) (*RedisDB, error) {
	return &RedisDB{
		client: client,
	}, nil
}

// TryLock attempts to create a new lock. If the lock is
// acquired, it will return true and the lock returned will be newLock.
// If the lock is not acquired, it will return false and the current
// lock that is preventing this lock from being acquired.
func (r *RedisDB) TryLock(newLock models.ProjectLock) (bool, models.ProjectLock, *models.EnqueueStatus, error) {
	var currLock models.ProjectLock
	lockKey := r.lockKey(newLock.Project, newLock.Workspace)
	newLockSerialized, _ := json.Marshal(newLock)
	var enqueueStatus *models.EnqueueStatus

	val, err := r.client.Get(ctx, lockKey).Result()
	// if there is no run at that key then we're free to create the lock
	if err == redis.Nil {
		err := r.client.Set(ctx, lockKey, newLockSerialized, 0).Err()
		if err != nil {
			return false, currLock, enqueueStatus, errors.Wrap(err, "db transaction failed")
		}
		return true, newLock, enqueueStatus, nil
	} else if err != nil {
		// otherwise the lock fails, return to caller the run that's holding the lock
		return false, currLock, enqueueStatus, errors.Wrap(err, "db transaction failed")
	} else {
		if err := json.Unmarshal([]byte(val), &currLock); err != nil {
			return false, currLock, enqueueStatus, errors.Wrap(err, "failed to deserialize current lock")
		}
		// checking if current lock is with the same PR or if queue is disabled
		if currLock.Pull.Num == newLock.Pull.Num || !r.queueEnabled {
			return false, currLock, enqueueStatus, nil
		}
		enqueueStatus, err = r.enqueue(newLock)
		return false, currLock, enqueueStatus, err
	}
}

func (r *RedisDB) enqueue(newLock models.ProjectLock) (*models.EnqueueStatus, error) {
	queueKey := r.queueKey(newLock.Project, newLock.Workspace)
	currQueueSerialized, err := r.client.Get(ctx, queueKey).Result()
	var queue models.ProjectLockQueue
	if err == redis.Nil {
		queue = models.ProjectLockQueue{}
	} else if err != nil {
		return nil, errors.Wrap(err, "db transaction failed")
	} else {
		if err := json.Unmarshal([]byte(currQueueSerialized), &queue); err != nil {
			return nil, errors.Wrap(err, "failed to deserialize current queue")
		}
	}
	// Lock is already in the queue
	if indexInQueue := queue.FindPullRequest(newLock.Pull.Num); indexInQueue > -1 {
		enqueueStatus := &models.EnqueueStatus{
			Status:     models.AlreadyInTheQueue,
			QueueDepth: indexInQueue + 1,
		}
		return enqueueStatus, nil
	}

	// Not in the queue, add it
	newQueue := append(queue, newLock)
	newQueueSerialized, err := json.Marshal(newQueue)
	if err != nil {
		return nil, errors.Wrap(err, "serializing")
	}
	err = r.client.Set(ctx, queueKey, newQueueSerialized, 0).Err()
	if err != nil {
		return nil, errors.Wrap(err, "db transaction failed")
	}
	enqueueStatus := &models.EnqueueStatus{
		Status:     models.Enqueued,
		QueueDepth: len(newQueue),
	}
	return enqueueStatus, nil
}

// Unlock attempts to unlock the project and workspace.
// If there is no lock, then it will return a nil pointer.
// If there is a lock, then it will delete it, and then return a pointer
// to the deleted lock. If updateQueue is true, it will also grant the
// lock to the next PR in the queue, update the queue and return the dequeued lock.
func (r *RedisDB) Unlock(project models.Project, workspace string) (*models.ProjectLock, *models.ProjectLock, error) {
	var lock models.ProjectLock
	lockKey := r.lockKey(project, workspace)

	val, err := r.client.Get(ctx, lockKey).Result()
	if err == redis.Nil {
		return nil, nil, nil
	} else if err != nil {
		return nil, nil, errors.Wrap(err, "db transaction failed")
	} else {
		if err := json.Unmarshal([]byte(val), &lock); err != nil {
			return nil, nil, errors.Wrap(err, "failed to deserialize current lock")
		}
		r.client.Del(ctx, lockKey)
		// Dequeue next item
		if r.queueEnabled {
			dequeuedLock, err := r.dequeue(project, workspace, lockKey)
			return &lock, dequeuedLock, err
		}
		return &lock, nil, nil
	}
}

func (r *RedisDB) dequeue(project models.Project, workspace string, lockKey string) (*models.ProjectLock, error) {
	queueKey := r.queueKey(project, workspace)
	currQueueSerialized, err := r.client.Get(ctx, queueKey).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "db transaction failed")
	} else {
		var currQueue models.ProjectLockQueue
		if err := json.Unmarshal([]byte(currQueueSerialized), &currQueue); err != nil {
			return nil, errors.Wrap(err, "failed to deserialize queue for current lock")
		}

		dequeuedLock, newQueue := currQueue.Dequeue()

		// A lock was dequeued - update current lock holder
		if dequeuedLock != nil {
			dequeuedLockSerialized, err := json.Marshal(*dequeuedLock)
			if err != nil {
				return dequeuedLock, errors.Wrap(err, "serializing")
			}
			err = r.client.Set(ctx, lockKey, dequeuedLockSerialized, 0).Err()
			if err != nil {
				return dequeuedLock, errors.Wrap(err, "db transaction failed")
			}
		}

		// New queue is empty and can be deleted
		if len(newQueue) == 0 {
			r.client.Del(ctx, queueKey)
			return dequeuedLock, nil
		}

		newQueueSerialized, err := json.Marshal(newQueue)
		if err != nil {
			return dequeuedLock, errors.Wrap(err, "serializing")
		}
		err = r.client.Set(ctx, queueKey, newQueueSerialized, 0).Err()
		if err != nil {
			return dequeuedLock, errors.Wrap(err, "db transaction failed")
		}
		return dequeuedLock, nil
	}
}

// List lists all current locks.
func (r *RedisDB) List() ([]models.ProjectLock, error) {
	var locks []models.ProjectLock
	iter := r.client.Scan(ctx, 0, "pr*", 0).Iterator()
	for iter.Next(ctx) {
		var lock models.ProjectLock
		val, err := r.client.Get(ctx, iter.Val()).Result()
		if err != nil {
			return nil, errors.Wrap(err, "db transaction failed")
		}
		if err := json.Unmarshal([]byte(val), &lock); err != nil {
			return locks, errors.Wrap(err, fmt.Sprintf("failed to deserialize lock at key '%s'", iter.Val()))
		}
		locks = append(locks, lock)
	}
	if err := iter.Err(); err != nil {
		return locks, errors.Wrap(err, "db transaction failed")
	}

	return locks, nil
}

// GetLock returns a pointer to the lock for that project and workspace.
// If there is no lock, it returns a nil pointer.
func (r *RedisDB) GetLock(project models.Project, workspace string) (*models.ProjectLock, error) {
	key := r.lockKey(project, workspace)

	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "db transaction failed")
	} else {
		var lock models.ProjectLock
		if err := json.Unmarshal([]byte(val), &lock); err != nil {
			return nil, errors.Wrapf(err, "deserializing lock at key %q", key)
		}
		// need to set it to Local after deserialization due to https://github.com/golang/go/issues/19486
		lock.Time = lock.Time.Local()
		return &lock, nil
	}
}

// UnlockByPull deletes all locks associated with that pull request and returns them.
func (r *RedisDB) UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, *models.DequeueStatus, error) {
	var locks []models.ProjectLock
	var dequeuedLocks = make([]models.ProjectLock, 0, len(locks))

	iter := r.client.Scan(ctx, 0, fmt.Sprintf("pr/%s*", repoFullName), 0).Iterator()
	for iter.Next(ctx) {
		var lock models.ProjectLock
		val, err := r.client.Get(ctx, iter.Val()).Result()
		if err != nil {
			return nil, nil, errors.Wrap(err, "db transaction failed")
		}
		if err := json.Unmarshal([]byte(val), &lock); err != nil {
			return locks, nil, errors.Wrap(err, fmt.Sprintf("failed to deserialize lock at key '%s'", iter.Val()))
		}
		if lock.Pull.Num == pullNum {
			locks = append(locks, lock)
			_, dequeuedLock, err := r.Unlock(lock.Project, lock.Workspace)
			if err != nil {
				return locks, nil, errors.Wrapf(err, "unlocking repo %s, path %s, workspace %s", lock.Project.RepoFullName, lock.Project.Path, lock.Workspace)
			}
			if dequeuedLock != nil {
				dequeuedLocks = append(dequeuedLocks, *dequeuedLock)
			}
		}
	}

	if err := iter.Err(); err != nil {
		return locks, &models.DequeueStatus{ProjectLocks: dequeuedLocks}, errors.Wrap(err, "db transaction failed")
	}

	return locks, &models.DequeueStatus{ProjectLocks: dequeuedLocks}, nil
}

// GetQueueByLock returns the queue for a given lock.
// If queue is not enabled or if no such queue exists, it returns a nil pointer.
func (r *RedisDB) GetQueueByLock(project models.Project, workspace string) (models.ProjectLockQueue, error) {
	if !r.queueEnabled {
		return nil, nil
	}

	key := r.queueKey(project, workspace)

	queueSerialized, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		// Queue not found
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "db transaction failed")
	} else {
		// Queue is found, deserialize and return
		var queue models.ProjectLockQueue
		if err := json.Unmarshal([]byte(queueSerialized), &queue); err != nil {
			return nil, errors.Wrapf(err, "deserializing queue at key %q", key)
		}
		for _, lock := range queue {
			// need to set it to Local after deserialization due to https://github.com/golang/go/issues/19486
			lock.Time = lock.Time.Local()
		}
		return queue, nil
	}
}

func (r *RedisDB) LockCommand(cmdName command.Name, lockTime time.Time) (*command.Lock, error) {

	lock := command.Lock{
		CommandName: cmdName,
		LockMetadata: command.LockMetadata{
			UnixTime: lockTime.Unix(),
		},
	}

	cmdLockKey := r.commandLockKey(cmdName)

	newLockSerialized, _ := json.Marshal(lock)

	_, err := r.client.Get(ctx, cmdLockKey).Result()
	if err == redis.Nil {
		err = r.client.Set(ctx, cmdLockKey, newLockSerialized, 0).Err()
		return &lock, errors.Wrap(err, "db transaction failed")
	} else if err != nil {
		return nil, errors.Wrap(err, "db transaction failed")
	} else {
		return nil, errors.New("db transaction failed: lock already exists")
	}
}

func (r *RedisDB) UnlockCommand(cmdName command.Name) error {
	cmdLockKey := r.commandLockKey(cmdName)
	_, err := r.client.Get(ctx, cmdLockKey).Result()
	if err == redis.Nil {
		return errors.New("db transaction failed: no lock exists")
	} else if err != nil {
		return errors.Wrap(err, "db transaction failed")
	} else {
		return r.client.Del(ctx, cmdLockKey).Err()
	}
}

func (r *RedisDB) CheckCommandLock(cmdName command.Name) (*command.Lock, error) {
	cmdLock := command.Lock{}

	cmdLockKey := r.commandLockKey(cmdName)
	val, err := r.client.Get(ctx, cmdLockKey).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "db transaction failed")
	} else {
		if err := json.Unmarshal([]byte(val), &cmdLock); err != nil {
			return nil, errors.Wrap(err, "failed to deserialize Lock")
		}
		return &cmdLock, err
	}
}

// UpdatePullWithResults updates pull's status with the latest project results.
// It returns the new PullStatus object.
func (r *RedisDB) UpdateProjectStatus(pull models.PullRequest, workspace string, repoRelDir string, newStatus models.ProjectPlanStatus) error {
	key, err := r.pullKey(pull)
	if err != nil {
		return err
	}

	currStatusPtr, err := r.getPull(key)
	if err != nil {
		return err
	}
	if currStatusPtr == nil {
		return nil
	}
	currStatus := *currStatusPtr

	// Update the status.
	for i := range currStatus.Projects {
		// NOTE: We're using a reference here because we are
		// in-place updating its Status field.
		proj := &currStatus.Projects[i]
		if proj.Workspace == workspace && proj.RepoRelDir == repoRelDir {
			proj.Status = newStatus
			break
		}
	}

	err = r.writePull(key, currStatus)
	return errors.Wrap(err, "db transaction failed")
}

func (r *RedisDB) GetPullStatus(pull models.PullRequest) (*models.PullStatus, error) {
	key, err := r.pullKey(pull)
	if err != nil {
		return nil, err
	}

	pullStatus, err := r.getPull(key)

	return pullStatus, errors.Wrap(err, "db transaction failed")
}

func (r *RedisDB) DeletePullStatus(pull models.PullRequest) error {
	key, err := r.pullKey(pull)
	if err != nil {
		return err
	}
	return errors.Wrap(r.deletePull(key), "db transaction failed")
}

func (r *RedisDB) UpdatePullWithResults(pull models.PullRequest, newResults []command.ProjectResult) (models.PullStatus, error) {
	key, err := r.pullKey(pull)
	if err != nil {
		return models.PullStatus{}, err
	}

	var newStatus models.PullStatus
	currStatus, err := r.getPull(key)
	if err != nil {
		return newStatus, errors.Wrap(err, "db transaction failed")
	}

	// If there is no pull OR if the pull we have is out of date, we
	// just write a new pull.
	if currStatus == nil || currStatus.Pull.HeadCommit != pull.HeadCommit {
		var statuses []models.ProjectStatus
		for _, res := range newResults {
			statuses = append(statuses, r.projectResultToProject(res))
		}
		newStatus = models.PullStatus{
			Pull:     pull,
			Projects: statuses,
		}
	} else {
		// If there's an existing pull at the right commit then we have to
		// merge our project results with the existing ones. We do a merge
		// because it's possible a user is just applying a single project
		// in this command and so we don't want to delete our data about
		// other projects that aren't affected by this command.
		newStatus = *currStatus
		for _, res := range newResults {
			// First, check if we should update any existing projects.
			updatedExisting := false
			for i := range newStatus.Projects {
				// NOTE: We're using a reference here because we are
				// in-place updating its Status field.
				proj := &newStatus.Projects[i]
				if res.Workspace == proj.Workspace &&
					res.RepoRelDir == proj.RepoRelDir &&
					res.ProjectName == proj.ProjectName {

					proj.Status = res.PlanStatus()

					// Updating only policy sets which are included in results; keeping the rest.
					if len(proj.PolicyStatus) > 0 {
						for i, oldPolicySet := range proj.PolicyStatus {
							for _, newPolicySet := range res.PolicyStatus() {
								if oldPolicySet.PolicySetName == newPolicySet.PolicySetName {
									proj.PolicyStatus[i] = newPolicySet
								}
							}
						}
					} else {
						proj.PolicyStatus = res.PolicyStatus()
					}

					updatedExisting = true
					break
				}
			}

			if !updatedExisting {
				// If we didn't update an existing project, then we need to
				// add this because it's a new one.
				newStatus.Projects = append(newStatus.Projects, r.projectResultToProject(res))
			}
		}
	}

	// Now, we overwrite the key with our new status.
	return newStatus, errors.Wrap(r.writePull(key, newStatus), "db transaction failed")
}

func (r *RedisDB) getPull(key string) (*models.PullStatus, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "db transaction failed")
	} else {
		var p models.PullStatus
		if err := json.Unmarshal([]byte(val), &p); err != nil {
			return nil, errors.Wrapf(err, "deserializing pull at %q with contents %q", key, val)
		}
		return &p, nil
	}
}

func (r *RedisDB) writePull(key string, pull models.PullStatus) error {
	serialized, err := json.Marshal(pull)
	if err != nil {
		return errors.Wrap(err, "serializing")
	}
	err = r.client.Set(ctx, key, serialized, 0).Err()
	return errors.Wrap(err, "DB Transaction failed")
}

func (r *RedisDB) deletePull(key string) error {
	err := r.client.Del(ctx, key).Err()
	return errors.Wrap(err, "DB Transaction failed")
}

func (r *RedisDB) lockKey(p models.Project, workspace string) string {
	return fmt.Sprintf("pr/%s/%s/%s", p.RepoFullName, p.Path, workspace)
}

func (r *RedisDB) queueKey(p models.Project, workspace string) string {
	return fmt.Sprintf("queue/%s/%s/%s", p.RepoFullName, p.Path, workspace)
}

func (r *RedisDB) commandLockKey(cmdName command.Name) string {
	return fmt.Sprintf("global/%s/lock", cmdName)
}

func (r *RedisDB) pullKey(pull models.PullRequest) (string, error) {
	hostname := pull.BaseRepo.VCSHost.Hostname
	if strings.Contains(hostname, pullKeySeparator) {
		return "", fmt.Errorf("vcs hostname %q contains illegal string %q", hostname, pullKeySeparator)
	}
	repo := pull.BaseRepo.FullName
	if strings.Contains(repo, pullKeySeparator) {
		return "", fmt.Errorf("repo name %q contains illegal string %q", hostname, pullKeySeparator)
	}

	return fmt.Sprintf("%s::%s::%d", hostname, repo, pull.Num), nil
}

func (r *RedisDB) projectResultToProject(p command.ProjectResult) models.ProjectStatus {
	return models.ProjectStatus{
		Workspace:    p.Workspace,
		RepoRelDir:   p.RepoRelDir,
		ProjectName:  p.ProjectName,
		PolicyStatus: p.PolicyStatus(),
		Status:       p.PlanStatus(),
	}
}
