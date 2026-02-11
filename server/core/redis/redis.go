// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

// Package redis handles our remote database layer.
package redis

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

var ctx = context.Background()

// Redis is a database using Redis 6
type RedisDB struct { // nolint: revive
	client *redis.Client
}

const (
	pullKeySeparator       = "::"
	projectOutputKeyPrefix = "output/"
	pullOutputIndexPrefix  = "pull-outputs/"
	jobIDIndexPrefix       = "job-id-index/"
)

func New(hostname string, port int, password string, tlsEnabled bool, insecureSkipVerify bool, db int) (*RedisDB, error) {
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
		return nil, fmt.Errorf("failed to connect to redis instance at %s:%d: %w", hostname, port, err)
	}

	// Migrate old lock keys to new format.
	// Old format: pr/{repoFullName}/{path}/{workspace}
	// New format: pr/{repoFullName}/{path}/{workspace}/{projectName}
	// We scan all keys and for those that don't match the new format,
	// we read their value, create a new key with the new format and
	// delete the old key.
	allKeys := rdb.Keys(ctx, "pr/*")
	for _, oldKey := range allKeys.Val() {
		// Remove the "pr/" prefix to validate the key format
		keyWithoutPrefix := strings.TrimPrefix(oldKey, "pr/")

		_, err := locking.IsCurrentLocking(keyWithoutPrefix)
		if err != nil {
			var currLock models.ProjectLock
			oldValue, err := rdb.Get(ctx, oldKey).Result()
			if err != nil {
				return nil, errors.Wrap(err, "failed to get current lock")
			}
			if err := json.Unmarshal([]byte(oldValue), &currLock); err != nil {
				return nil, errors.Wrap(err, "failed to deserialize current lock")
			}
			newKey := fmt.Sprintf("pr/%s", models.GenerateLockKey(currLock.Project, currLock.Workspace))
			rdb.Set(ctx, newKey, oldValue, 0)
			rdb.Del(ctx, oldKey)
		}
	}

	return &RedisDB{
		client: rdb,
	}, nil
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
	if err != nil {
		return fmt.Errorf("db transaction failed: %w", err)
	}
	return nil
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

	var newStatus models.PullStatus
	currStatus, err := r.getPull(key)
	if err != nil {
		return newStatus, fmt.Errorf("db transaction failed: %w", err)
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
	err = r.writePull(key, newStatus)
	if err != nil {
		return models.PullStatus{}, fmt.Errorf("db transaction failed: %w", err)
	}
	return newStatus, nil
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

func (r *RedisDB) writePull(key string, pull models.PullStatus) error {
	serialized, err := json.Marshal(pull)
	if err != nil {
		return fmt.Errorf("serializing: %w", err)
	}
	err = r.client.Set(ctx, key, serialized, 0).Err()
	if err != nil {
		return fmt.Errorf("DB Transaction failed: %w", err)
	}
	return nil
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

func (r *RedisDB) projectOutputKey(key string) string {
	return projectOutputKeyPrefix + key
}

func (r *RedisDB) pullOutputIndexKey(repoFullName string, pullNum int) string {
	return fmt.Sprintf("%s%s::%d", pullOutputIndexPrefix, repoFullName, pullNum)
}

func (r *RedisDB) jobIDIndexKey(jobID string) string {
	return jobIDIndexPrefix + jobID
}

// SaveProjectOutput saves a project output to Redis atomically.
func (r *RedisDB) SaveProjectOutput(output models.ProjectOutput) error {
	key := r.projectOutputKey(output.Key())

	bytes, err := json.Marshal(output)
	if err != nil {
		return fmt.Errorf("marshaling project output: %w", err)
	}

	// Use transaction to ensure atomicity of save + index updates
	indexKey := r.pullOutputIndexKey(output.RepoFullName, output.PullNum)
	pipe := r.client.TxPipeline()
	pipe.Set(ctx, key, bytes, 0)
	pipe.SAdd(ctx, indexKey, key)

	// Also save the job ID index for O(1) lookups by job ID
	if output.JobID != "" {
		pipe.Set(ctx, r.jobIDIndexKey(output.JobID), key, 0)
	}

	_, err = pipe.Exec(ctx)
	return err
}

// GetProjectOutputRun retrieves a specific project output run.
func (r *RedisDB) GetProjectOutputRun(repoFullName string, pullNum int, path string, workspace string, projectName string, command string, runTimestamp int64) (*models.ProjectOutput, error) {
	key := fmt.Sprintf("%s::%d::%s::%s::%s::%s::%d", repoFullName, pullNum, path, workspace, projectName, command, runTimestamp)
	return r.getProjectOutputByKey(r.projectOutputKey(key))
}

func (r *RedisDB) getProjectOutputByKey(key string) (*models.ProjectOutput, error) {
	bytes, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var output models.ProjectOutput
	if err := json.Unmarshal(bytes, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

// GetProjectOutputHistory retrieves all runs for a project, sorted by timestamp descending.
func (r *RedisDB) GetProjectOutputHistory(repoFullName string, pullNum int, path string, workspace string, projectName string) ([]models.ProjectOutput, error) {
	pattern := fmt.Sprintf("%s::%d::%s::%s::%s::*", repoFullName, pullNum, path, workspace, projectName)
	return r.getProjectOutputsByPattern(r.projectOutputKey(pattern), true)
}

func (r *RedisDB) getProjectOutputsByPattern(pattern string, sortDescending bool) ([]models.ProjectOutput, error) {
	var outputs []models.ProjectOutput

	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		bytes, err := r.client.Get(ctx, iter.Val()).Bytes()
		if err != nil {
			continue
		}

		var output models.ProjectOutput
		if err := json.Unmarshal(bytes, &output); err != nil {
			continue
		}
		outputs = append(outputs, output)
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	if outputs == nil {
		outputs = []models.ProjectOutput{}
	}

	if sortDescending {
		sort.Slice(outputs, func(i, j int) bool {
			return outputs[i].RunTimestamp > outputs[j].RunTimestamp
		})
	}

	return outputs, nil
}

// GetProjectOutputsByPull retrieves the latest output per project for a pull request.
// Uses the pull-level index set (SMEMBERS) instead of SCAN for O(n) key lookup
// instead of iterating over the entire keyspace.
func (r *RedisDB) GetProjectOutputsByPull(repoFullName string, pullNum int) ([]models.ProjectOutput, error) {
	indexKey := r.pullOutputIndexKey(repoFullName, pullNum)
	keys, err := r.client.SMembers(ctx, indexKey).Result()
	if err != nil {
		return nil, err
	}

	// Group by project key and keep only the latest per project.
	// Errors from individual key lookups are intentionally skipped — index sets
	// may contain stale references to expired or deleted keys.
	latestByProject := make(map[string]models.ProjectOutput)
	for _, key := range keys {
		output, err := r.getProjectOutputByKey(key)
		if err != nil || output == nil {
			continue
		}
		projectKey := output.ProjectKey()
		if existing, ok := latestByProject[projectKey]; !ok || output.RunTimestamp > existing.RunTimestamp {
			latestByProject[projectKey] = *output
		}
	}

	outputs := make([]models.ProjectOutput, 0, len(latestByProject))
	for _, output := range latestByProject {
		outputs = append(outputs, output)
	}

	return outputs, nil
}

// DeleteProjectOutputsByPull deletes all project outputs for a pull request
// atomically using a Redis pipeline, including job-id-index entries.
func (r *RedisDB) DeleteProjectOutputsByPull(repoFullName string, pullNum int) error {
	indexKey := r.pullOutputIndexKey(repoFullName, pullNum)

	keys, err := r.client.SMembers(ctx, indexKey).Result()
	if err != nil {
		return err
	}

	// Collect job-id-index keys to clean up.
	// Errors from individual key lookups are intentionally skipped — index sets
	// may contain stale references to expired or deleted keys.
	var jobIDIndexKeys []string
	for _, key := range keys {
		output, err := r.getProjectOutputByKey(key)
		if err == nil && output != nil && output.JobID != "" {
			jobIDIndexKeys = append(jobIDIndexKeys, r.jobIDIndexKey(output.JobID))
		}
	}

	pipe := r.client.Pipeline()
	if len(keys) > 0 {
		pipe.Del(ctx, keys...)
	}
	for _, jk := range jobIDIndexKeys {
		pipe.Del(ctx, jk)
	}
	pipe.Del(ctx, indexKey)
	_, err = pipe.Exec(ctx)
	return err
}

// GetProjectOutputByJobID retrieves a project output by its job ID.
// Uses an index for O(1) lookups, with fallback to full scan for backwards compatibility.
func (r *RedisDB) GetProjectOutputByJobID(jobID string) (*models.ProjectOutput, error) {
	// First, try to use the job ID index for O(1) lookup
	outputKey, err := r.client.Get(ctx, r.jobIDIndexKey(jobID)).Result()
	if err == nil && outputKey != "" {
		output, err := r.getProjectOutputByKey(outputKey)
		if err != nil {
			return nil, err
		}
		if output != nil {
			return output, nil
		}
	} else if err != nil && err != redis.Nil {
		return nil, err
	}

	// Fallback to full scan for backwards compatibility with existing data
	// that doesn't have an index entry yet
	pattern := r.projectOutputKey("*")
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		bytes, err := r.client.Get(ctx, iter.Val()).Bytes()
		if err != nil {
			continue
		}

		var output models.ProjectOutput
		if err := json.Unmarshal(bytes, &output); err != nil {
			continue
		}
		if output.JobID == jobID {
			return &output, nil
		}
	}
	return nil, iter.Err()
}

// GetActivePullRequests returns all pull requests that have stored project outputs.
func (r *RedisDB) GetActivePullRequests() ([]models.PullRequest, error) {
	// Scan for all pull index keys (avoids blocking KEYS command)
	var keys []string
	iter := r.client.Scan(ctx, 0, pullOutputIndexPrefix+"*", 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}

	pullSet := make(map[string]models.PullRequest)
	for _, indexKey := range keys {
		// Parse the key to extract repo and pull number
		// Format: pull-outputs/{repo}::{pullNum}
		suffix := strings.TrimPrefix(indexKey, pullOutputIndexPrefix)
		parts := strings.SplitN(suffix, "::", 2)
		if len(parts) != 2 {
			continue
		}
		repoFullName := parts[0]
		pullNum, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}

		pullKey := fmt.Sprintf("%s::%d", repoFullName, pullNum)
		if _, exists := pullSet[pullKey]; !exists {
			pullSet[pullKey] = models.PullRequest{
				Num: pullNum,
				BaseRepo: models.Repo{
					FullName: repoFullName,
				},
			}
		}

		// Try to get URL and Title from one of the outputs
		outputKeys, err := r.client.SMembers(ctx, indexKey).Result()
		if err != nil || len(outputKeys) == 0 {
			continue
		}

		// Check first output for URL/Title
		for _, outputKey := range outputKeys {
			output, err := r.getProjectOutputByKey(outputKey)
			if err != nil || output == nil {
				continue
			}
			if output.PullURL != "" || output.PullTitle != "" {
				existing := pullSet[pullKey]
				if output.PullURL != "" {
					existing.URL = output.PullURL
				}
				if output.PullTitle != "" {
					existing.Title = output.PullTitle
				}
				pullSet[pullKey] = existing
				break // Found URL/Title, no need to check more outputs
			}
		}
	}

	pulls := make([]models.PullRequest, 0, len(pullSet))
	for _, pr := range pullSet {
		pulls = append(pulls, pr)
	}
	return pulls, nil
}

func (r *RedisDB) Close() error {
	return r.client.Close()
}
