// Package boltdb provides a locking implementation using Bolt.
// Bolt is a key/value store that writes all data to a file.
// See https://github.com/boltdb/bolt for more information.
package boltdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/boltdb/bolt"
	"github.com/hootsuite/atlantis/server/events/models"
	"github.com/pkg/errors"
)

// BoltLocker is a locking backend using BoltDB
type BoltLocker struct {
	db     *bolt.DB
	bucket []byte
}

const bucketName = "runLocks"

// New returns a valid locker. We need to be able to write to dataDir
// since bolt stores its data as a file
func New(dataDir string) (*BoltLocker, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, errors.Wrap(err, "creating data dir")
	}
	db, err := bolt.Open(path.Join(dataDir, "atlantis.db"), 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		if err.Error() == "timeout" {
			return nil, errors.New("starting BoltDB: timeout (a possible cause is another Atlantis instance already running)")
		}
		return nil, errors.Wrap(err, "starting BoltDB")
	}
	err = db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(bucketName)); err != nil {
			return errors.Wrapf(err, "creating %q bucketName", bucketName)
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "starting BoltDB")
	}
	// todo: close BoltDB when server is sigtermed
	return &BoltLocker{db, []byte(bucketName)}, nil
}

// NewWithDB is used for testing.
func NewWithDB(db *bolt.DB, bucket string) (*BoltLocker, error) {
	return &BoltLocker{db, []byte(bucket)}, nil
}

// TryLock attempts to create a new lock. If the lock is
// acquired, it will return true and the lock returned will be newLock.
// If the lock is not acquired, it will return false and the current
// lock that is preventing this lock from being acquired.
func (b *BoltLocker) TryLock(newLock models.ProjectLock) (bool, models.ProjectLock, error) {
	var lockAcquired bool
	var currLock models.ProjectLock
	key := b.key(newLock.Project, newLock.Env)
	newLockSerialized, _ := json.Marshal(newLock)
	transactionErr := b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(b.bucket)

		// if there is no run at that key then we're free to create the lock
		currLockSerialized := bucket.Get([]byte(key))
		if currLockSerialized == nil {
			bucket.Put([]byte(key), newLockSerialized) // not a readonly bucketName so okay to ignore error
			lockAcquired = true
			currLock = newLock
			return nil
		}

		// otherwise the lock fails, return to caller the run that's holding the lock
		if err := json.Unmarshal(currLockSerialized, &currLock); err != nil {
			return errors.Wrap(err, "failed to deserialize current lock")
		}
		lockAcquired = false
		return nil
	})

	if transactionErr != nil {
		return false, currLock, errors.Wrap(transactionErr, "DB transaction failed")
	}

	return lockAcquired, currLock, nil
}

// Unlock attempts to unlock the project and environment.
// If there is no lock, then it will return a nil pointer.
// If there is a lock, then it will delete it, and then return a pointer
// to the deleted lock.
func (b BoltLocker) Unlock(p models.Project, env string) (*models.ProjectLock, error) {
	var lock models.ProjectLock
	foundLock := false
	key := b.key(p, env)
	err := b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(b.bucket)
		serialized := bucket.Get([]byte(key))
		if serialized != nil {
			if err := json.Unmarshal(serialized, &lock); err != nil {
				return errors.Wrap(err, "failed to deserialize lock")
			}
			foundLock = true
		}
		return bucket.Delete([]byte(key))
	})
	err = errors.Wrap(err, "DB transaction failed")
	if foundLock {
		return &lock, err
	}
	return nil, err
}

// List lists all current locks.
func (b BoltLocker) List() ([]models.ProjectLock, error) {
	var locks []models.ProjectLock
	var locksBytes [][]byte
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(b.bucket)
		c := bucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			locksBytes = append(locksBytes, v)
		}
		return nil
	})
	if err != nil {
		return locks, errors.Wrap(err, "DB transaction failed")
	}

	// deserialize bytes into the proper objects
	for k, v := range locksBytes {
		var lock models.ProjectLock
		if err := json.Unmarshal(v, &lock); err != nil {
			return locks, errors.Wrap(err, fmt.Sprintf("failed to deserialize lock at key %q", string(k)))
		}
		locks = append(locks, lock)
	}

	return locks, nil
}

// UnlockByPull deletes all locks associated with that pull request and returns them.
func (b BoltLocker) UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error) {
	var locks []models.ProjectLock
	err := b.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(b.bucket).Cursor()

		// we can use the repoFullName as a prefix search since that's the first part of the key
		for k, v := c.Seek([]byte(repoFullName)); k != nil && bytes.HasPrefix(k, []byte(repoFullName)); k, v = c.Next() {
			var lock models.ProjectLock
			if err := json.Unmarshal(v, &lock); err != nil {
				return errors.Wrapf(err, "deserializing lock at key %q", string(k))
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
		if _, err = b.Unlock(lock.Project, lock.Env); err != nil {
			return locks, errors.Wrapf(err, "unlocking repo %s, path %s, env %s", lock.Project.RepoFullName, lock.Project.Path, lock.Env)
		}
	}
	return locks, nil
}

// GetLock returns a pointer to the lock for that project and env.
// If there is no lock, it returns a nil pointer.
func (b BoltLocker) GetLock(p models.Project, env string) (*models.ProjectLock, error) {
	key := b.key(p, env)
	var lockBytes []byte
	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(b.bucket)
		lockBytes = b.Get([]byte(key))
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "getting lock data")
	}
	// lockBytes will be nil if there was no data at that key
	if lockBytes == nil {
		return nil, nil
	}

	var lock models.ProjectLock
	if err := json.Unmarshal(lockBytes, &lock); err != nil {
		return nil, errors.Wrapf(err, "deserializing lock at key %q", key)
	}

	// need to set it to Local after deserialization due to https://github.com/golang/go/issues/19486
	lock.Time = lock.Time.Local()
	return &lock, nil
}

func (b BoltLocker) key(p models.Project, env string) string {
	return fmt.Sprintf("%s/%s/%s", p.RepoFullName, p.Path, env)
}
