package locking

import (
	"github.com/boltdb/bolt"
	"fmt"
	"encoding/json"
	"github.com/pkg/errors"
	"encoding/hex"
)

type BoltDBLockManager struct {
	db          *bolt.DB
	locksBucket []byte
}

func NewBoltDBLockManager(db *bolt.DB, locksBucket string) *BoltDBLockManager {
	return &BoltDBLockManager{db, []byte(locksBucket)}
}

func (b BoltDBLockManager) TryLock(run Run) (TryLockResponse, error) {
	var response TryLockResponse
	newRunSerialized, err := b.serialize(run)
	if err != nil {
		return response, errors.Wrap(err, "failed to serialize run")
	}

	lockId := run.StateKey()
	transactionErr := b.db.Update(func(tx *bolt.Tx) error {
		locksBucket := tx.Bucket(b.locksBucket)

		// if there is no run at that key then we're free to create the lock
		lockingRunSerialized := locksBucket.Get(lockId)
		if lockingRunSerialized == nil {
			locksBucket.Put(lockId, newRunSerialized) // not a readonly bucket so okay to ignore error
			response = TryLockResponse{
				LockAcquired: true,
				LockingRun: run,
				LockID: b.lockIDToString(lockId),
			}
			return nil
		}

		// otherwise the lock fails, return to caller the run that's holding the lock
		var lockingRun Run
		if err := b.deserialize(lockingRunSerialized, &lockingRun); err != nil {
			return errors.Wrap(err, "failed to deserialize run")
		}
		response = TryLockResponse{
			LockAcquired: false,
			LockingRun: lockingRun,
			LockID: b.lockIDToString(lockId),
		}
		return nil
	})

	if transactionErr != nil {
		return response, errors.Wrap(transactionErr, "DB transaction failed")
	}

	return response, nil
}

func (b BoltDBLockManager) Unlock(lockID string) error {
	idAsHex, err := hex.DecodeString(lockID)
	if err != nil {
		return errors.Wrap(err, "id was not in correct format")
	}
	err = b.db.Update(func(tx *bolt.Tx) error {
		locks := tx.Bucket(b.locksBucket)
		return locks.Delete(idAsHex)
	})
	return errors.Wrap(err, "DB transaction failed")
}

func (b BoltDBLockManager) ListLocks() (map[string]Run, error) {
	m := make(map[string]Run)
	bytes := make(map[string][]byte)

	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(b.locksBucket)
		c := bucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			bytes[b.lockIDToString(k)] = v
		}
		return nil
	})
	if err != nil {
		return m, errors.Wrap(err, "DB transaction failed")
	}

	// deserialize bytes into the proper objects
	for k, v := range bytes {
		var run Run
		if err := b.deserialize(v, &run); err != nil {
			return m, errors.Wrap(err, fmt.Sprintf("failed to deserialize run at key %q", string(k)))
		}
		m[k] = run
	}

	return m, nil
}

func (b BoltDBLockManager) lockIDToString(key []byte) string {
	return string(hex.EncodeToString(key))
}

func (b BoltDBLockManager) deserialize(bs []byte, run *Run) error {
	return json.Unmarshal(bs, run)
}

func (b BoltDBLockManager) serialize(run Run) ([]byte, error) {
	return json.Marshal(run)
}
