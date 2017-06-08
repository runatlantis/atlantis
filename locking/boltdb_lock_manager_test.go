package locking_test

import (
	"testing"
	"github.com/hootsuite/atlantis/locking"
	"time"
	. "github.com/hootsuite/atlantis/testing_util"
	"os"
	"github.com/boltdb/bolt"
	"io/ioutil"
	"github.com/pkg/errors"
)

const LockBucket = "locks"

var run = locking.Run{
	RepoFullName: "owner/repo",
	Path: "parent/child",
	Env: "default",
	PullNum: 1,
	User: "user",
	Timestamp: time.Now(),
}

func keyWithNewField(fieldName string, value string) locking.Run {
	copy := run
	switch fieldName {
	case "RepoFullName":
		copy.RepoFullName = value
		return copy
	case "Path":
		copy.Path = value
		return copy
	case "Env":
		copy.Env = value
		return copy
	default:
		return copy
	}
}

// NewTestDB returns a TestDB using a temporary path.
func NewTestDB() (*bolt.DB, *locking.BoltDBLockManager) {
	// Retrieve a temporary path.
	f, err := ioutil.TempFile("", "")
	if err != nil {
		panic(errors.Wrap(err, "failed to create temp file"))
	}
	path := f.Name()
	f.Close()

	// Open the database.
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		panic(errors.Wrap(err, "could not start bolt DB"))
	}
	if err := db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(LockBucket)); err != nil {
			return errors.Wrap(err, "failed to create bucket")
		}
		return nil
	}); err != nil {
		panic(errors.Wrap(err, "could not create bucket"))
	}
	b, _ := locking.NewBoltDBLockManagerWithDB(db, LockBucket)
	return db, b
}

func cleanupDB(db *bolt.DB) {
	os.Remove(db.Path())
	db.Close()
}

func TestListWhenNoLocks(t *testing.T) {
	db, b := NewTestDB()
	defer cleanupDB(db)
	m, err := b.ListLocks()
	Ok(t, err)
	Equals(t, 0, len(m))
}

func TestListLocks(t *testing.T) {
	db, b := NewTestDB()
	defer cleanupDB(db)
	_, err := b.TryLock(run)
	Ok(t, err)

	locks, err := b.ListLocks()
	Assert(t, len(locks) == 1, "should have 1 lock")
	for _, v := range locks {
		Equals(t, run, v)
	}
}

type LockCommand struct {
	run      locking.Run
	expected locking.TryLockResponse
}

type UnlockCommand struct {
	runKey   string
	expected error
}

var regularLock = LockCommand{
	run,
	locking.TryLockResponse{
		LockAcquired: true,
		LockingRun: run,
		LockID: "cae47fa9dfe223b0d8fefd51f2fbbf9849047c9d048d659d1ba906acc2913534",
	},
}

var tableTestData = []struct {
	sequence    []interface{}
	description string
}{
	{
		description: "regular lock should succeed",
		sequence: []interface{}{regularLock},
	},
	{
		description: "unlock when there are no locks should succeed",
		sequence: []interface{}{
			UnlockCommand{
				"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				nil,
			},
		},
	},
	{
		description: "lock and unlock should succeed",
		sequence: []interface{}{
			regularLock,
			UnlockCommand{
				"cae47fa9dfe223b0d8fefd51f2fbbf9849047c9d048d659d1ba906acc2913534",
				nil,
			},
		},
	},
	{
		description: "lock for same run path should fail",
		sequence: []interface{}{
			regularLock,
			LockCommand{
				run,
				locking.TryLockResponse{
					LockAcquired: false,
					LockingRun: regularLock.run,
					LockID: regularLock.expected.LockID,
				},
			},
		},
	},
	{
		description: "lock when existing lock is for different repo name should succeed",
		sequence: []interface{}{
			regularLock,
			LockCommand{
				keyWithNewField("RepoFullName", "repo/different"),
				locking.TryLockResponse{
					LockAcquired: true,
					LockingRun: keyWithNewField("RepoFullName", "repo/different"),
				},
			},
		},
	},
	{
		description: "lock when existing lock is for different path should succeed",
		sequence: []interface{}{
			regularLock,
			LockCommand{
				keyWithNewField("Path", "different"),
				locking.TryLockResponse{
					LockAcquired: true,
					LockingRun: keyWithNewField("Path", "different"),
				},
			},
		},
	},
	{
		description: "lock when existing lock is for different env should succeed",
		sequence: []interface{}{
			regularLock,
			LockCommand{
				keyWithNewField("Env", "different"),
				locking.TryLockResponse{
					LockAcquired: true,
					LockingRun: keyWithNewField("Env", "different"),
				},
			},
		},
	},
	{
		description: "lock, unlock, lock should succeed",
		sequence: []interface{}{
			regularLock,
			UnlockCommand{
				"cae47fa9dfe223b0d8fefd51f2fbbf9849047c9d048d659d1ba906acc2913534",
				nil,
			},
			LockCommand{
				run,
				locking.TryLockResponse{
					LockAcquired: true,
					LockingRun: run,
					LockID: regularLock.expected.LockID,
				},
			},
		},
	},
}

func TestAllCases(t *testing.T) {
	for _, testCase := range tableTestData {
		t.Logf("testing: %q", testCase.description)

		db, b := NewTestDB()
		for i, cmd := range testCase.sequence {
			t.Logf("index: %d", i)
			switch cmd.(type) {
			case LockCommand:
				lockCmd := cmd.(LockCommand)
				r, err := b.TryLock(lockCmd.run)
				Ok(t, err)
				Equals(t, lockCmd.expected.LockAcquired, r.LockAcquired)
				Equals(t, lockCmd.expected.LockingRun, r.LockingRun)
				if lockCmd.expected.LockID != "" {
					Equals(t, lockCmd.expected.LockID, r.LockID)
				}
			case UnlockCommand:
				unlockCmd := cmd.(UnlockCommand)
				err := b.Unlock(unlockCmd.runKey)
				Equals(t, unlockCmd.expected, err)
			}
		}
		cleanupDB(db)
	}
}

