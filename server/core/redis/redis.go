// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

// Package redis handles our remote database layer.
package redis

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

var ctx = context.Background()
var errPullStatusMissing = errors.New("pull status is missing")

// RedisDB is a database using Redis 6+
type RedisDB struct { // nolint: revive
	client redis.Cmdable
}

const (
	pullKeySeparator = "::"
)

const unlockIfOwnedByPullScript = "" +
	"local value = redis.call(\"GET\", KEYS[1])\n" +
	"if not value then\n" +
	"  return nil\n" +
	"end\n" +
	"\n" +
	"local ok, lock = pcall(cjson.decode, value)\n" +
	"if not ok then\n" +
	"  return redis.error_reply(\"failed to deserialize current lock\")\n" +
	"end\n" +
	"\n" +
	"if not lock[\"Pull\"] or tonumber(lock[\"Pull\"][\"Num\"]) ~= tonumber(ARGV[1]) then\n" +
	"  return nil\n" +
	"end\n" +
	"\n" +
	"redis.call(\"DEL\", KEYS[1])\n" +
	"return value\n"

const compareAndSwapPullScript = "" +
	"local current = redis.call(\"GET\", KEYS[1])\n" +
	"if ARGV[1] == \"0\" then\n" +
	"  if current then return 0 end\n" +
	"elseif not current or current ~= ARGV[2] then\n" +
	"  return 0\n" +
	"end\n" +
	"redis.call(\"SET\", KEYS[1], ARGV[3])\n" +
	"return 1\n"

// Config holds configuration for Redis connections.
type Config struct {
	Hostname           string
	Port               int
	Password           string
	Username           string
	TLSEnabled         bool
	InsecureSkipVerify bool
	DB                 int
	// ClusterAddresses is a list of cluster node addresses. When set, cluster mode is used.
	ClusterAddresses []string
}

// New creates a new RedisDB for client interactions with redis.
// Deprecated: Use NewWithConfig for new code.
func New(hostname string, port int, password string, tlsEnabled bool, insecureSkipVerify bool, db int) (*RedisDB, error) {
	return NewWithConfig(Config{
		Hostname:           hostname,
		Port:               port,
		Password:           password,
		TLSEnabled:         tlsEnabled,
		InsecureSkipVerify: insecureSkipVerify,
		DB:                 db,
	})
}

// NewWithConfig creates a new RedisDB based on the provided configuration.
// It automatically selects the appropriate Redis client type:
// - If ClusterAddresses is set, uses Redis Cluster mode
// - Otherwise, uses single-node mode
func NewWithConfig(cfg Config) (*RedisDB, error) {
	var rdb redis.Cmdable

	var tlsConfig *tls.Config
	if cfg.TLSEnabled {
		tlsConfig = &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: cfg.InsecureSkipVerify, //nolint:gosec // In some cases, users may want to use this at their own caution
		}
	}

	// Determine which Redis client to use based on configuration
	var connDesc string
	switch {
	case len(cfg.ClusterAddresses) > 0:
		// Filter out empty addresses
		var addrs []string
		for _, addr := range cfg.ClusterAddresses {
			trimmed := strings.TrimSpace(addr)
			if trimmed != "" {
				addrs = append(addrs, trimmed)
			}
		}
		if len(addrs) == 0 {
			return nil, errors.New("redis cluster addresses provided but all are empty")
		}
		rdb = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:     addrs,
			Username:  cfg.Username,
			Password:  cfg.Password,
			TLSConfig: tlsConfig,
		})
		connDesc = fmt.Sprintf("cluster nodes %s", strings.Join(addrs, ", "))
	default:
		address := fmt.Sprintf("%s:%d", cfg.Hostname, cfg.Port)
		rdb = redis.NewClient(&redis.Options{
			Addr:      address,
			Username:  cfg.Username,
			Password:  cfg.Password,
			DB:        cfg.DB,
			TLSConfig: tlsConfig,
		})
		connDesc = address
	}

	// Check if connection is valid
	err := rdb.Ping(ctx).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis at %s: %w", connDesc, err)
	}

	// Migrate old lock keys to new format with a bounded timeout.
	// Non-fatal: if migration times out or fails, remaining keys will be
	// retried on next startup. This avoids blocking boot on large key sets.
	migrateCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if err := migrateOldLockKeys(migrateCtx, rdb); err != nil {
		log.Printf("WARN: lock key migration incomplete (will retry next startup): %v", err)
	}

	return &RedisDB{
		client: rdb,
	}, nil
}

// migrateOldLockKeys migrates old lock key format to new format.
// Old format: pr/{repoFullName}/{path}/{workspace}
// New format: pr/{repoFullName}/{path}/{workspace}/{projectName}
// Uses Scan instead of Keys for compatibility with Redis Cluster (Scan fans out
// across all nodes via go-redis ClusterClient, whereas Keys does not).
func migrateOldLockKeys(ctx context.Context, rdb redis.Cmdable) error {
	iter := rdb.Scan(ctx, 0, "pr/*", 0).Iterator()
	for iter.Next(ctx) {
		oldKey := iter.Val()
		// Remove the "pr/" prefix to validate the key format
		keyWithoutPrefix := strings.TrimPrefix(oldKey, "pr/")

		_, err := locking.IsCurrentLocking(keyWithoutPrefix)
		if err != nil {
			var currLock models.ProjectLock
			oldValue, err := rdb.Get(ctx, oldKey).Result()
			if err != nil {
				return errors.Wrap(err, "failed to get current lock")
			}
			if err := json.Unmarshal([]byte(oldValue), &currLock); err != nil {
				return errors.Wrap(err, "failed to deserialize current lock")
			}
			newKey := fmt.Sprintf("pr/%s", models.GenerateLockKey(currLock.Project, currLock.Workspace))

			// Skip if the new key already exists (idempotent on retry).
			if _, err := rdb.Get(ctx, newKey).Result(); err == nil {
				// New key exists — just clean up the old one.
				rdb.Del(ctx, oldKey)
				continue
			}

			if err := rdb.Set(ctx, newKey, oldValue, 0).Err(); err != nil {
				return errors.Wrapf(err, "failed to set new lock key %s", newKey)
			}
			if err := rdb.Del(ctx, oldKey).Err(); err != nil {
				return errors.Wrapf(err, "failed to delete old lock key %s", oldKey)
			}
		}
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed scanning for old lock keys: %w", err)
	}
	return nil
}

// NewWithClient is used for testing.
func NewWithClient(client *redis.Client, _ string, _ string) (*RedisDB, error) {
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
			return false, currLock, fmt.Errorf("db transaction failed: %w", err)
		}
		return true, newLock, nil
	} else if err != nil {
		// otherwise the lock fails, return to caller the run that's holding the lock
		return false, currLock, fmt.Errorf("db transaction failed: %w", err)
	}

	if err := json.Unmarshal([]byte(val), &currLock); err != nil {
		return false, currLock, fmt.Errorf("failed to deserialize current lock: %w", err)
	}
	return false, currLock, nil
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
		return nil, fmt.Errorf("db transaction failed: %w", err)
	}

	if err := json.Unmarshal([]byte(val), &lock); err != nil {
		return nil, fmt.Errorf("failed to deserialize current lock: %w", err)
	}
	r.client.Del(ctx, key)
	return &lock, nil
}

// UnlockIfOwnedByPull deletes a lock only if it is still owned by pullNum.
func (r *RedisDB) UnlockIfOwnedByPull(project models.Project, workspace string, pullNum int) (*models.ProjectLock, error) {
	key := r.lockKey(project, workspace)
	val, err := r.client.Eval(ctx, unlockIfOwnedByPullScript, []string{key}, pullNum).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("db transaction failed: %w", err)
	}
	if val == nil {
		return nil, nil
	}

	serializedLock, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected unlock script result type %T", val)
	}

	var lock models.ProjectLock
	if err := json.Unmarshal([]byte(serializedLock), &lock); err != nil {
		return nil, fmt.Errorf("failed to deserialize current lock: %w", err)
	}
	return &lock, nil
}

// List lists all current locks.
func (r *RedisDB) List() ([]models.ProjectLock, error) {
	var locks []models.ProjectLock
	iter := r.client.Scan(ctx, 0, "pr*", 0).Iterator()
	for iter.Next(ctx) {
		var lock models.ProjectLock
		val, err := r.client.Get(ctx, iter.Val()).Result()
		if err != nil {
			return nil, fmt.Errorf("db transaction failed: %w", err)
		}
		if err := json.Unmarshal([]byte(val), &lock); err != nil {
			return locks, fmt.Errorf("failed to deserialize lock at key '%s': %w", iter.Val(), err)
		}
		locks = append(locks, lock)
	}
	if err := iter.Err(); err != nil {
		return locks, fmt.Errorf("db transaction failed: %w", err)
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
		return nil, fmt.Errorf("db transaction failed: %w", err)
	}

	var lock models.ProjectLock
	if err := json.Unmarshal([]byte(val), &lock); err != nil {
		return nil, fmt.Errorf("deserializing lock at key %q: %w", key, err)
	}
	// need to set it to Local after deserialization due to https://github.com/golang/go/issues/19486
	lock.Time = lock.Time.Local()
	return &lock, nil
}

// UnlockByPull deletes all locks associated with that pull request and returns them.
func (r *RedisDB) UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error) {
	var locks []models.ProjectLock

	iter := r.client.Scan(ctx, 0, fmt.Sprintf("pr/%s*", repoFullName), 0).Iterator()
	for iter.Next(ctx) {
		var lock models.ProjectLock
		val, err := r.client.Get(ctx, iter.Val()).Result()
		if err != nil {
			return nil, fmt.Errorf("db transaction failed: %w", err)
		}
		if err := json.Unmarshal([]byte(val), &lock); err != nil {
			return locks, fmt.Errorf("failed to deserialize lock at key '%s': %w", iter.Val(), err)
		}
		if lock.Pull.Num == pullNum {
			locks = append(locks, lock)
			if _, err := r.Unlock(lock.Project, lock.Workspace); err != nil {
				return locks, fmt.Errorf("unlocking repo %s, path %s, workspace %s: %w", lock.Project.RepoFullName, lock.Project.Path, lock.Workspace, err)
			}
		}
	}

	if err := iter.Err(); err != nil {
		return locks, fmt.Errorf("db transaction failed: %w", err)
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
		if err != nil {
			return nil, fmt.Errorf("db transaction failed: %w", err)
		}
		return &lock, nil
	} else if err != nil {
		return nil, fmt.Errorf("db transaction failed: %w", err)
	}

	return nil, errors.New("db transaction failed: lock already exists")
}

func (r *RedisDB) UnlockCommand(cmdName command.Name) error {
	cmdLockKey := r.commandLockKey(cmdName)
	_, err := r.client.Get(ctx, cmdLockKey).Result()
	if err == redis.Nil {
		return errors.New("db transaction failed: no lock exists")
	} else if err != nil {
		return fmt.Errorf("db transaction failed: %w", err)
	}

	return r.client.Del(ctx, cmdLockKey).Err()

}

func (r *RedisDB) CheckCommandLock(cmdName command.Name) (*command.Lock, error) {
	cmdLock := command.Lock{}

	cmdLockKey := r.commandLockKey(cmdName)
	val, err := r.client.Get(ctx, cmdLockKey).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("db transaction failed: %w", err)
	}

	if err := json.Unmarshal([]byte(val), &cmdLock); err != nil {
		return nil, fmt.Errorf("failed to deserialize Lock: %w", err)
	}
	return &cmdLock, err
}

// UpdateProjectStatus updates pull's status with the latest project results.
// It returns the new PullStatus object.
func (r *RedisDB) UpdateProjectStatus(pull models.PullRequest, workspace string, repoRelDir string, newStatus models.ProjectPlanStatus) error {
	key, err := r.pullKey(pull)
	if err != nil {
		return err
	}

	_, err = r.updatePullAtomically(key, false, func(currStatus *models.PullStatus) (models.PullStatus, error) {
		if currStatus == nil {
			return models.PullStatus{}, errPullStatusMissing
		}
		newPullStatus := *currStatus
		for i := range newPullStatus.Projects {
			proj := &newPullStatus.Projects[i]
			if proj.Workspace == workspace && proj.RepoRelDir == repoRelDir {
				if proj.PlanGeneration != "" {
					return models.PullStatus{}, fmt.Errorf("project has an active plan generation for dir %q workspace %q project %q", proj.RepoRelDir, proj.Workspace, proj.ProjectName)
				}
				proj.Status = newStatus
				proj.PlanGeneration = ""
				break
			}
		}
		return newPullStatus, nil
	})
	if errors.Is(err, errPullStatusMissing) {
		return nil
	}
	return err
}

func (r *RedisDB) GetPullStatus(pull models.PullRequest) (*models.PullStatus, error) {
	key, err := r.pullKey(pull)
	if err != nil {
		return nil, err
	}

	pullStatus, err := r.getPull(key)
	if err != nil {
		return nil, fmt.Errorf("db transaction failed: %w", err)
	}
	return pullStatus, nil
}

func (r *RedisDB) DeletePullStatus(pull models.PullRequest) error {
	key, err := r.pullKey(pull)
	if err != nil {
		return err
	}
	err = r.deletePull(key)
	if err != nil {
		return fmt.Errorf("db transaction failed: %w", err)
	}
	return nil
}

func (r *RedisDB) UpdatePullWithResults(pull models.PullRequest, newResults []command.ProjectResult) (models.PullStatus, error) {
	key, err := r.pullKey(pull)
	if err != nil {
		return models.PullStatus{}, err
	}

	return r.updatePullAtomically(key, true, func(currStatus *models.PullStatus) (models.PullStatus, error) {
		if currStatus != nil && pullStatusOutdatedForPull(currStatus.Pull, pull) {
			if err := rejectAnyActivePlanGeneration(currStatus); err != nil {
				return models.PullStatus{}, err
			}
		}
		if err := rejectResultsForActivePlanGeneration(currStatus, newResults); err != nil {
			return models.PullStatus{}, err
		}
		var newStatus models.PullStatus
		// If there is no pull OR if the pull we have is out of date, we
		// just write a new pull.
		if currStatus == nil || pullStatusOutdatedForPull(currStatus.Pull, pull) {
			var statuses []models.ProjectStatus
			for _, res := range newResults {
				statuses = append(statuses, r.projectResultToProject(res))
			}
			// Preserve policy status from the previous commit so approvals
			// survive between the plan DB write and the subsequent policy
			// check DB write. doPolicyCheck applies sticky filtering and
			// overwrites these when it writes its own results.
			if currStatus != nil {
				for i := range statuses {
					for _, old := range currStatus.Projects {
						if statuses[i].Workspace == old.Workspace &&
							statuses[i].RepoRelDir == old.RepoRelDir &&
							statuses[i].ProjectName == old.ProjectName &&
							len(old.PolicyStatus) > 0 {
							statuses[i].PolicyStatus = old.PolicyStatus
							break
						}
					}
				}
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
				updatedExisting := false
				for i := range newStatus.Projects {
					proj := &newStatus.Projects[i]
					if res.Workspace == proj.Workspace &&
						res.RepoRelDir == proj.RepoRelDir &&
						res.ProjectName == proj.ProjectName {
						proj.Status = res.PlanStatus()
						proj.PlanGeneration = ""

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
					newStatus.Projects = append(newStatus.Projects, r.projectResultToProject(res))
				}
			}
		}
		return newStatus, nil
	})
}

// ReplacePullWithResults atomically replaces a pull status unless a plan
// generation is active. Only CompletePlanGeneration may finalize that state.
func (r *RedisDB) ReplacePullWithResults(pull models.PullRequest, newResults []command.ProjectResult) (models.PullStatus, error) {
	key, err := r.pullKey(pull)
	if err != nil {
		return models.PullStatus{}, err
	}

	return r.updatePullAtomically(key, true, func(currStatus *models.PullStatus) (models.PullStatus, error) {
		if err := rejectAnyActivePlanGeneration(currStatus); err != nil {
			return models.PullStatus{}, err
		}
		newStatus := models.PullStatus{Pull: pull}
		for _, result := range newResults {
			newStatus.Projects = append(newStatus.Projects, r.projectResultToProject(result))
		}
		return newStatus, nil
	})
}

// BeginPlanGeneration atomically makes the selected projects non-applyable
// before their plan steps can replace any plan artifacts.
func (r *RedisDB) BeginPlanGeneration(pull models.PullRequest, projects []models.ProjectStatus, generation string) (models.PullStatus, error) {
	if generation == "" {
		return models.PullStatus{}, errors.New("plan generation is empty")
	}
	key, err := r.pullKey(pull)
	if err != nil {
		return models.PullStatus{}, err
	}

	return r.updatePullAtomically(key, false, func(currStatus *models.PullStatus) (models.PullStatus, error) {
		var newStatus models.PullStatus
		if currStatus == nil || pullStatusOutdatedForPull(currStatus.Pull, pull) {
			newStatus = models.PullStatus{Pull: pull}
		} else {
			newStatus = *currStatus
		}

		for _, project := range projects {
			updated := false
			for i := range newStatus.Projects {
				current := &newStatus.Projects[i]
				if sameProjectStatus(*current, project) {
					current.Status = models.ErroredPlanStatus
					current.PlanGeneration = generation
					updated = true
					break
				}
			}
			if !updated {
				policyStatus := project.PolicyStatus
				if currStatus != nil {
					if previous := findProjectStatus(currStatus.Projects, project.Workspace, project.RepoRelDir, project.ProjectName); previous != nil {
						policyStatus = previous.PolicyStatus
					}
				}
				newStatus.Projects = append(newStatus.Projects, models.ProjectStatus{
					Workspace:      project.Workspace,
					RepoRelDir:     project.RepoRelDir,
					ProjectName:    project.ProjectName,
					PlanGeneration: generation,
					PolicyStatus:   policyStatus,
					Status:         models.ErroredPlanStatus,
				})
			}
		}
		return newStatus, nil
	})
}

// CompletePlanGeneration atomically persists final plan results only while the
// selected projects still belong to the plan generation that produced them.
func (r *RedisDB) CompletePlanGeneration(pull models.PullRequest, generation string, newResults []command.ProjectResult) (models.PullStatus, error) {
	if generation == "" {
		return models.PullStatus{}, errors.New("plan generation is empty")
	}
	key, err := r.pullKey(pull)
	if err != nil {
		return models.PullStatus{}, err
	}

	return r.updatePullAtomically(key, false, func(currStatus *models.PullStatus) (models.PullStatus, error) {
		if currStatus == nil {
			return models.PullStatus{}, errors.New("plan generation status is missing")
		}
		if pullStatusOutdatedForPull(currStatus.Pull, pull) {
			return models.PullStatus{}, errors.New("plan generation pull identity changed")
		}

		newStatus := *currStatus
		for _, result := range newResults {
			project := findProjectStatus(newStatus.Projects, result.Workspace, result.RepoRelDir, result.ProjectName)
			if project == nil || project.Status != models.ErroredPlanStatus || project.PlanGeneration != generation {
				return models.PullStatus{}, fmt.Errorf("plan generation %q is no longer current for dir %q workspace %q project %q", generation, result.RepoRelDir, result.Workspace, result.ProjectName)
			}
			project.Status = result.PlanStatus()
			project.PlanGeneration = ""
		}
		for _, project := range newStatus.Projects {
			if project.PlanGeneration == generation {
				return models.PullStatus{}, fmt.Errorf("plan generation %q is incomplete for dir %q workspace %q project %q", generation, project.RepoRelDir, project.Workspace, project.ProjectName)
			}
		}
		return newStatus, nil
	})
}

func (r *RedisDB) updatePullAtomically(key string, tolerateUnreadable bool, update func(*models.PullStatus) (models.PullStatus, error)) (models.PullStatus, error) {
	const maxAttempts = 32
	for attempt := range maxAttempts {
		serializedCurrent, err := r.client.Get(ctx, key).Result()
		exists := true
		if err == redis.Nil {
			exists = false
			serializedCurrent = ""
		} else if err != nil {
			return models.PullStatus{}, fmt.Errorf("db transaction failed: %w", err)
		}

		var current *models.PullStatus
		if exists {
			var decoded models.PullStatus
			if err := json.Unmarshal([]byte(serializedCurrent), &decoded); err != nil {
				if !tolerateUnreadable {
					return models.PullStatus{}, fmt.Errorf("deserializing pull at %q with contents %q: %w", key, serializedCurrent, err)
				}
				log.Printf("warning: discarding unreadable pull status at %q: %v", key, err)
			} else {
				current = &decoded
			}
		}

		newStatus, err := update(current)
		if err != nil {
			return models.PullStatus{}, err
		}
		serializedNew, err := json.Marshal(newStatus)
		if err != nil {
			return models.PullStatus{}, fmt.Errorf("serializing: %w", err)
		}
		existsArg := "0"
		if exists {
			existsArg = "1"
		}
		swapped, err := r.client.Eval(ctx, compareAndSwapPullScript, []string{key}, existsArg, serializedCurrent, serializedNew).Int()
		if err != nil {
			return models.PullStatus{}, fmt.Errorf("db transaction failed: %w", err)
		}
		if swapped == 1 {
			return newStatus, nil
		}
		if attempt+1 < maxAttempts {
			time.Sleep(time.Duration(rand.IntN(10)+1) * time.Millisecond)
		}
	}
	return models.PullStatus{}, errors.New("db transaction failed: pull status changed too many times")
}

func sameProjectStatus(left, right models.ProjectStatus) bool {
	return left.Workspace == right.Workspace &&
		left.RepoRelDir == right.RepoRelDir &&
		left.ProjectName == right.ProjectName
}

func findProjectStatus(projects []models.ProjectStatus, workspace, repoRelDir, projectName string) *models.ProjectStatus {
	for i := range projects {
		project := &projects[i]
		if project.Workspace == workspace && project.RepoRelDir == repoRelDir && project.ProjectName == projectName {
			return project
		}
	}
	return nil
}

func rejectResultsForActivePlanGeneration(status *models.PullStatus, results []command.ProjectResult) error {
	if status == nil {
		return nil
	}
	for _, result := range results {
		project := findProjectStatus(status.Projects, result.Workspace, result.RepoRelDir, result.ProjectName)
		if project != nil && project.PlanGeneration != "" {
			return fmt.Errorf("project has an active plan generation for dir %q workspace %q project %q", result.RepoRelDir, result.Workspace, result.ProjectName)
		}
	}
	return nil
}

func rejectAnyActivePlanGeneration(status *models.PullStatus) error {
	if status == nil {
		return nil
	}
	for _, project := range status.Projects {
		if project.PlanGeneration != "" {
			return fmt.Errorf("project has an active plan generation for dir %q workspace %q project %q", project.RepoRelDir, project.Workspace, project.ProjectName)
		}
	}
	return nil
}

func pullStatusOutdatedForPull(statusPull models.PullRequest, pull models.PullRequest) bool {
	if statusPull.HeadCommit != pull.HeadCommit {
		return true
	}
	if pull.BaseBranch == "" {
		return false
	}
	return statusPull.BaseBranch == "" || statusPull.BaseBranch != pull.BaseBranch
}

func (r *RedisDB) getPull(key string) (*models.PullStatus, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("db transaction failed: %w", err)
	}

	var p models.PullStatus
	if err := json.Unmarshal([]byte(val), &p); err != nil {
		return nil, fmt.Errorf("deserializing pull at %q with contents %q: %w", key, val, err)
	}
	return &p, nil
}

func (r *RedisDB) deletePull(key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("DB Transaction failed: %w", err)
	}
	return nil
}

func (r *RedisDB) lockKey(p models.Project, workspace string) string {
	return fmt.Sprintf("pr/%s", models.GenerateLockKey(p, workspace))
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

// Ping checks the Redis connection health.
func (r *RedisDB) Ping() error {
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return r.client.Ping(pingCtx).Err()
}

func (r *RedisDB) Close() error {
	// Prefer a narrower interface and return an explicit error for unsupported client types.
	if closer, ok := r.client.(interface{ Close() error }); ok {
		return closer.Close()
	}

	return fmt.Errorf("redis: unsupported client type %T does not implement Close() error", r.client)
}
