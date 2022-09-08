// Package redis handles our remote database layer.
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

var ctx = context.Background()

// Redis is a database using Redis 6
type RedisDB struct { // nolint: revive
	client *redis.Client
}

const (
	pullKeySeparator = "::"
)

func New(hostname string, port int, password string) (*RedisDB, error) {

	fmt.Println(hostname, port)
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", hostname, port),
		Password: password,
		DB:       0, // use default DB
	})

	return &RedisDB{
		client: rdb,
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
func (r *RedisDB) TryLock(newLock models.ProjectLock) (bool, models.ProjectLock, error) {
	var currLock models.ProjectLock
	key := r.lockKey(newLock.Project, newLock.Workspace)
	newLockSerialized, _ := json.Marshal(newLock)

	val, err := r.client.Get(ctx, key).Result()
	// if there is no run at that key then we're free to create the lock
	if err == redis.Nil {
		err := r.client.Set(ctx, key, newLockSerialized, 0).Err()
		if err != nil {
			return false, currLock, errors.Wrap(err, "db transaction failed")
		}
		return true, newLock, nil
	} else if err != nil {
		// otherwise the lock fails, return to caller the run that's holding the lock
		return false, currLock, errors.Wrap(err, "db transaction failed")
	} else {
		if err := json.Unmarshal([]byte(val), &currLock); err != nil {
			return false, currLock, errors.Wrap(err, "failed to deserialize current lock")
		}
		return false, currLock, nil
	}
}

// Unlock attempts to unlock the project and workspace.
// If there is no lock, then it will return a nil pointer.
// If there is a lock, then it will delete it, and then return a pointer
// to the deleted lock.
func (r *RedisDB) Unlock(project models.Project, workspace string) (*models.ProjectLock, error) {
	var lock models.ProjectLock
	key := r.lockKey(project, workspace)

	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "db transaction failed")
	} else {
		if err := json.Unmarshal([]byte(val), &lock); err != nil {
			return nil, errors.Wrap(err, "failed to deserialize current lock")
		}
		r.client.Del(ctx, key)
		return &lock, nil
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
func (r *RedisDB) UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error) {
	var locks []models.ProjectLock

	iter := r.client.Scan(ctx, 0, fmt.Sprintf("pr/%s*", repoFullName), 0).Iterator()
	for iter.Next(ctx) {
		var lock models.ProjectLock
		val, err := r.client.Get(ctx, iter.Val()).Result()
		if err != nil {
			return nil, errors.Wrap(err, "db transaction failed")
		}
		if err := json.Unmarshal([]byte(val), &lock); err != nil {
			return locks, errors.Wrap(err, fmt.Sprintf("failed to deserialize lock at key '%s'", iter.Val()))
		}
		if lock.Pull.Num == pullNum {
			locks = append(locks, lock)
			if _, err := r.Unlock(lock.Project, lock.Workspace); err != nil {
				return locks, errors.Wrapf(err, "unlocking repo %s, path %s, workspace %s", lock.Project.RepoFullName, lock.Project.Path, lock.Workspace)
			}
		}
	}

	if err := iter.Err(); err != nil {
		return locks, errors.Wrap(err, "db transaction failed")
	}

	return locks, nil
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
		Workspace:   p.Workspace,
		RepoRelDir:  p.RepoRelDir,
		ProjectName: p.ProjectName,
		Status:      p.PlanStatus(),
	}
}
