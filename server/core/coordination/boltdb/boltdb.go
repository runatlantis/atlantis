// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

// Package boltdb implements the coordination store on a local BoltDB file.
package boltdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	bolt "go.etcd.io/bbolt"
)

// Store implements coordination.Store on BoltDB.
type Store struct {
	db                    *bolt.DB
	locksBucketName       []byte
	pullsBucketName       []byte
	globalLocksBucketName []byte
}

const (
	locksBucketName       = "runLocks"
	pullsBucketName       = "pulls"
	globalLocksBucketName = "globalLocks"
	pullKeySeparator      = "::"
)

// New opens (creating if needed) the atlantis.db file under dataDir and
// builds the store on it. Production wiring opens the file through the
// backends registry and calls NewStore; this remains for tests.
func New(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, fmt.Errorf("creating data dir: %w", err)
	}
	db, err := bolt.Open(path.Join(dataDir, "atlantis.db"), 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		if err.Error() == "timeout" {
			return nil, errors.New("starting BoltDB: timeout (a possible cause is another Atlantis instance already running)")
		}
		return nil, fmt.Errorf("starting BoltDB: %w", err)
	}
	return NewStore(db)
}

// NewStore builds the coordination store on an opened bolt database,
// creating its buckets and migrating old lock keys.
func NewStore(db *bolt.DB) (*Store, error) {
	// Create the buckets.
	err := db.Update(func(tx *bolt.Tx) error {
		var err error
		if _, err = tx.CreateBucketIfNotExists([]byte(locksBucketName)); err != nil {
			return fmt.Errorf("creating bucket %q: %w", locksBucketName, err)
		}
		if _, err = tx.CreateBucketIfNotExists([]byte(pullsBucketName)); err != nil {
			return fmt.Errorf("creating bucket %q: %w", pullsBucketName, err)
		}
		if _, err = tx.CreateBucketIfNotExists([]byte(globalLocksBucketName)); err != nil {
			return fmt.Errorf("creating bucket %q: %w", globalLocksBucketName, err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("starting BoltDB: %w", err)
	}

	// Migrate old lock keys to new format.
	// Old format: {repoFullName}/{path}/{workspace}
	// New format: {repoFullName}/{path}/{workspace}/{projectName}
	// We scan all keys and for those that don't match the new format,
	// we read their value, create a new key with the new format and
	// delete the old key.
	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(locksBucketName))

		// Phase 1: Collect keys that need migration
		type migration struct {
			oldKey   []byte
			newKey   string
			oldValue []byte
		}
		var migrations []migration

		if err := bucket.ForEach(func(oldKey, oldValue []byte) error {
			_, _, err := models.ParseLockKey(string(oldKey))
			if err != nil {
				var currLock models.ProjectLock
				if err := json.Unmarshal(oldValue, &currLock); err != nil {
					return errors.Wrap(err, "failed to deserialize current lock")
				}
				newKey := models.GenerateLockKey(currLock.Project, currLock.Workspace)
				migrations = append(migrations, migration{
					oldKey:   append([]byte(nil), oldKey...),
					newKey:   newKey,
					oldValue: append([]byte(nil), oldValue...),
				})
			}
			return nil
		}); err != nil {
			return err
		}

		for _, m := range migrations {
			if err := bucket.Put([]byte(m.newKey), m.oldValue); err != nil {
				return err
			}
			if err := bucket.Delete(m.oldKey); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		log.Printf("warning: failed to migrate BoltDB lock keys: %v", err)
	}

	return &Store{
		db:                    db,
		locksBucketName:       []byte(locksBucketName),
		pullsBucketName:       []byte(pullsBucketName),
		globalLocksBucketName: []byte(globalLocksBucketName),
	}, nil
}

// NewWithDB is used for testing.
func NewWithDB(db *bolt.DB, bucket string, globalBucket string) (*Store, error) {
	return &Store{
		db:                    db,
		locksBucketName:       []byte(bucket),
		pullsBucketName:       []byte(pullsBucketName),
		globalLocksBucketName: []byte(globalBucket),
	}, nil
}

// TryLock attempts to create a new lock. If the lock is
// acquired, it will return true and the lock returned will be newLock.
// If the lock is not acquired, it will return false and the current
// lock that is preventing this lock from being acquired.
func (b *Store) TryLock(newLock models.ProjectLock) (bool, models.ProjectLock, error) {
	var lockAcquired bool
	var currLock models.ProjectLock
	key := b.lockKey(newLock.Project, newLock.Workspace)
	newLockSerialized, _ := json.Marshal(newLock)
	transactionErr := b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(b.locksBucketName)

		// if there is no run at that key then we're free to create the lock
		currLockSerialized := bucket.Get([]byte(key))
		if currLockSerialized == nil {
			// This will only error on readonly buckets, it's okay to ignore.
			bucket.Put([]byte(key), newLockSerialized) // nolint: errcheck
			lockAcquired = true
			currLock = newLock
			return nil
		}

		// otherwise the lock fails, return to caller the run that's holding the lock
		if err := json.Unmarshal(currLockSerialized, &currLock); err != nil {
			return fmt.Errorf("failed to deserialize current lock: %w", err)
		}
		lockAcquired = false
		return nil
	})

	if transactionErr != nil {
		return false, currLock, fmt.Errorf("DB transaction failed: %w", transactionErr)
	}

	return lockAcquired, currLock, nil
}

// Unlock attempts to unlock the project and workspace.
// If there is no lock, then it will return a nil pointer.
// If there is a lock, then it will delete it, and then return a pointer
// to the deleted lock.
func (b *Store) Unlock(p models.Project, workspace string) (*models.ProjectLock, error) {
	var lock models.ProjectLock
	foundLock := false
	key := b.lockKey(p, workspace)
	err := b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(b.locksBucketName)
		serialized := bucket.Get([]byte(key))
		if serialized != nil {
			if err := json.Unmarshal(serialized, &lock); err != nil {
				return fmt.Errorf("failed to deserialize lock: %w", err)
			}
			foundLock = true
		}
		return bucket.Delete([]byte(key))
	})
	if err != nil {
		err = fmt.Errorf("DB transaction failed: %w", err)
	}
	if foundLock {
		return &lock, err
	}
	return nil, err
}

// UnlockIfOwnedByPull deletes a lock only if it is still owned by pullNum.
func (b *Store) UnlockIfOwnedByPull(p models.Project, workspace string, pullNum int) (*models.ProjectLock, error) {
	var lock models.ProjectLock
	foundLock := false
	key := b.lockKey(p, workspace)
	err := b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(b.locksBucketName)
		serialized := bucket.Get([]byte(key))
		if serialized == nil {
			return nil
		}
		if err := json.Unmarshal(serialized, &lock); err != nil {
			return fmt.Errorf("failed to deserialize lock: %w", err)
		}
		if lock.Pull.Num != pullNum {
			return nil
		}
		if err := bucket.Delete([]byte(key)); err != nil {
			return err
		}
		foundLock = true
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("DB transaction failed: %w", err)
	}
	if foundLock {
		return &lock, nil
	}
	return nil, nil
}

// List lists all current locks.
func (b *Store) List() ([]models.ProjectLock, error) {
	var locks []models.ProjectLock
	var locksBytes [][]byte
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(b.locksBucketName)
		c := bucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			locksBytes = append(locksBytes, v)
		}
		return nil
	})
	if err != nil {
		return locks, fmt.Errorf("DB transaction failed: %w", err)
	}

	// deserialize bytes into the proper objects
	for k, v := range locksBytes {
		var lock models.ProjectLock
		if err := json.Unmarshal(v, &lock); err != nil {
			return locks, fmt.Errorf("failed to deserialize lock at key '%d': %w", k, err)
		}
		locks = append(locks, lock)
	}

	return locks, nil
}

// LockCommand attempts to create a new lock for a CommandName.
// If the lock doesn't exists, it will create a lock and return a pointer to it.
// If the lock already exists, it will return an "lock already exists" error
func (b *Store) LockCommand(cmdName command.Name, lockTime time.Time) (*command.Lock, error) {
	lock := command.Lock{
		CommandName: cmdName,
		LockMetadata: command.LockMetadata{
			UnixTime: lockTime.Unix(),
		},
	}

	newLockSerialized, _ := json.Marshal(lock)
	transactionErr := b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(b.globalLocksBucketName)

		currLockSerialized := bucket.Get([]byte(b.commandLockKey(cmdName)))
		if currLockSerialized != nil {
			return errors.New("lock already exists")
		}

		// This will only error on readonly buckets, it's okay to ignore.
		bucket.Put([]byte(b.commandLockKey(cmdName)), newLockSerialized) // nolint: errcheck
		return nil
	})

	if transactionErr != nil {
		return nil, fmt.Errorf("db transaction failed: %w", transactionErr)
	}

	return &lock, nil
}

// UnlockCommand removes CommandName lock if present.
// If there are no lock it returns an error.
func (b *Store) UnlockCommand(cmdName command.Name) error {
	transactionErr := b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(b.globalLocksBucketName)

		if l := bucket.Get([]byte(b.commandLockKey(cmdName))); l == nil {
			return errors.New("no lock exists")
		}

		return bucket.Delete([]byte(b.commandLockKey(cmdName)))
	})

	if transactionErr != nil {
		return fmt.Errorf("db transaction failed: %w", transactionErr)
	}

	return nil
}

// CheckCommandLock checks if CommandName lock was set.
// If the lock exists return the pointer to the lock object, otherwise return nil
func (b *Store) CheckCommandLock(cmdName command.Name) (*command.Lock, error) {
	cmdLock := command.Lock{}

	found := false

	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(b.globalLocksBucketName)

		serializedLock := bucket.Get([]byte(b.commandLockKey(cmdName)))

		if serializedLock != nil {
			if err := json.Unmarshal(serializedLock, &cmdLock); err != nil {
				return fmt.Errorf("failed to deserialize UserConfig: %w", err)
			}
			found = true
		}

		return nil
	})

	if found {
		return &cmdLock, err
	}

	return nil, err
}

// UnlockByPull deletes all locks associated with that pull request and returns them.
func (b *Store) UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error) {
	var locks []models.ProjectLock
	err := b.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(b.locksBucketName).Cursor()

		// we can use the repoFullName as a prefix search since that's the first part of the key
		for k, v := c.Seek([]byte(repoFullName)); k != nil && bytes.HasPrefix(k, []byte(repoFullName)); k, v = c.Next() {
			var lock models.ProjectLock
			if err := json.Unmarshal(v, &lock); err != nil {
				return fmt.Errorf("deserializing lock at key %q: %w", string(k), err)
			}
			if lock.Pull.Num == pullNum {
				locks = append(locks, lock)
			}
		}
		return nil
	})
	if err != nil {
		return locks, err
	}

	// delete the locks
	for _, lock := range locks {
		if _, err = b.Unlock(lock.Project, lock.Workspace); err != nil {
			return locks, fmt.Errorf("unlocking repo %s, path %s, workspace %s: %w", lock.Project.RepoFullName, lock.Project.Path, lock.Workspace, err)
		}
	}
	return locks, nil
}

// GetLock returns a pointer to the lock for that project and workspace.
// If there is no lock, it returns a nil pointer.
func (b *Store) GetLock(p models.Project, workspace string) (*models.ProjectLock, error) {
	key := b.lockKey(p, workspace)
	var lockBytes []byte
	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(b.locksBucketName)
		lockBytes = b.Get([]byte(key))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("getting lock data: %w", err)
	}
	// lockBytes will be nil if there was no data at that key
	if lockBytes == nil {
		return nil, nil
	}

	var lock models.ProjectLock
	if err := json.Unmarshal(lockBytes, &lock); err != nil {
		return nil, fmt.Errorf("deserializing lock at key %q: %w", key, err)
	}

	// need to set it to Local after deserialization due to https://github.com/golang/go/issues/19486
	lock.Time = lock.Time.Local()
	return &lock, nil
}

// UpdatePullWithResults updates pull's status with the latest project results.
// It returns the new PullStatus object.
func (b *Store) UpdatePullWithResults(pull models.PullRequest, newResults []command.ProjectResult) (models.PullStatus, error) {
	key, err := b.pullKey(pull)
	if err != nil {
		return models.PullStatus{}, err
	}

	var newStatus models.PullStatus
	err = b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(b.pullsBucketName)
		currStatus, err := b.getPullFromBucket(bucket, key)
		if err != nil {
			// Tolerate an unreadable prior entry (e.g. after a PullStatus
			// schema change). It will be overwritten with fresh data below;
			// in-flight policy approvals captured in the old blob are lost.
			log.Printf("warning: discarding unreadable pull status at %q: %v", key, err)
			currStatus = nil
		}

		// If there is no pull OR if the pull we have is out of date, we
		// just write a new pull.
		if currStatus == nil || pullStatusOutdatedForPull(currStatus.Pull, pull) {
			var statuses []models.ProjectStatus
			for _, r := range newResults {
				statuses = append(statuses, b.projectResultToProject(r))
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
					newStatus.Projects = append(newStatus.Projects, b.projectResultToProject(res))
				}
			}
		}

		// Now, we overwrite the key with our new status.
		return b.writePullToBucket(bucket, key, newStatus)
	})
	if err != nil {
		return models.PullStatus{}, fmt.Errorf("DB transaction failed: %w", err)
	}
	return newStatus, nil
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

// GetPullStatus returns the status for pull.
// If there is no status, returns a nil pointer.
func (b *Store) GetPullStatus(pull models.PullRequest) (*models.PullStatus, error) {
	key, err := b.pullKey(pull)
	if err != nil {
		return nil, err
	}
	var s *models.PullStatus
	err = b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(b.pullsBucketName)
		var txErr error
		s, txErr = b.getPullFromBucket(bucket, key)
		return txErr
	})
	if err != nil {
		return nil, fmt.Errorf("DB transaction failed: %w", err)
	}
	return s, nil
}

// DeletePullStatus deletes the status for pull.
func (b *Store) DeletePullStatus(pull models.PullRequest) error {
	key, err := b.pullKey(pull)
	if err != nil {
		return err
	}
	err = b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(b.pullsBucketName)
		return bucket.Delete(key)
	})
	if err != nil {
		return fmt.Errorf("DB transaction failed: %w", err)
	}
	return nil
}

// UpdateProjectStatus updates project status.
func (b *Store) UpdateProjectStatus(pull models.PullRequest, workspace string, repoRelDir string, newStatus models.ProjectPlanStatus) error {
	key, err := b.pullKey(pull)
	if err != nil {
		return err
	}
	err = b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(b.pullsBucketName)
		currStatusPtr, err := b.getPullFromBucket(bucket, key)
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
		return b.writePullToBucket(bucket, key, currStatus)
	})
	if err != nil {
		return fmt.Errorf("DB transaction failed: %w", err)
	}
	return nil
}

func (b *Store) pullKey(pull models.PullRequest) ([]byte, error) {
	hostname := pull.BaseRepo.VCSHost.Hostname
	if strings.Contains(hostname, pullKeySeparator) {
		return nil, fmt.Errorf("vcs hostname %q contains illegal string %q", hostname, pullKeySeparator)
	}
	repo := pull.BaseRepo.FullName
	if strings.Contains(repo, pullKeySeparator) {
		return nil, fmt.Errorf("repo name %q contains illegal string %q", hostname, pullKeySeparator)
	}

	return fmt.Appendf(nil, "%s::%s::%d", hostname, repo, pull.Num),
		nil
}

func (b *Store) commandLockKey(cmdName command.Name) string {
	return fmt.Sprintf("%s/lock", cmdName)
}

func (b *Store) lockKey(p models.Project, workspace string) string {
	return models.GenerateLockKey(p, workspace)
}

func (b *Store) getPullFromBucket(bucket *bolt.Bucket, key []byte) (*models.PullStatus, error) {
	serialized := bucket.Get(key)
	if serialized == nil {
		return nil, nil
	}

	var p models.PullStatus
	if err := json.Unmarshal(serialized, &p); err != nil {
		return nil, fmt.Errorf("deserializing pull at %q with contents %q: %w", key, serialized, err)
	}
	return &p, nil
}

func (b *Store) writePullToBucket(bucket *bolt.Bucket, key []byte, pull models.PullStatus) error {
	serialized, err := json.Marshal(pull)
	if err != nil {
		return fmt.Errorf("serializing: %w", err)
	}
	return bucket.Put(key, serialized)
}

func (b *Store) projectResultToProject(p command.ProjectResult) models.ProjectStatus {
	return models.ProjectStatus{
		Workspace:    p.Workspace,
		RepoRelDir:   p.RepoRelDir,
		ProjectName:  p.ProjectName,
		PolicyStatus: p.PolicyStatus(),
		Status:       p.PlanStatus(),
	}
}

// Ping checks the database connection health.
func (b *Store) Ping() error {
	return b.db.View(func(tx *bolt.Tx) error {
		return nil
	})
}

func (b *Store) Close() error {
	return b.db.Close()
}
