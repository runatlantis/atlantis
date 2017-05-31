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
	RepoOwner: "repoOwner",
	RepoName: "repo",
	Path: "parent/child",
	Env: "default",
	PullID: 1,
	User: "user",
	Timestamp: time.Now(),
}

func keyWithNewField(fieldName string, value string) locking.Run {
	copy := run
	switch fieldName {
	case "RepoOwner":
		copy.RepoOwner = value
		return copy
	case "RepoName":
		copy.RepoName = value
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
func NewTestDB() *bolt.DB {
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
	return db
}

func cleanupDB(db *bolt.DB) {
	os.Remove(db.Path())
	db.Close()
}

func TestListWhenNoLocks(t *testing.T) {
	db := NewTestDB()
	defer cleanupDB(db)
	b := locking.NewBoltDBLockManager(db, LockBucket)
	m, err := b.ListLocks()
	Ok(t, err)
	Equals(t, 0, len(m))
}

func TestListLocks(t *testing.T) {
	db := NewTestDB()
	defer cleanupDB(db)
	b := locking.NewBoltDBLockManager(db, LockBucket)
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
		LockID: "f13c9d87dc7156638cc075f0164d628f1338d23b17296dd391d922ea6fe6e9f8",
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
				"f13c9d87dc7156638cc075f0164d628f1338d23b17296dd391d922ea6fe6e9f8",
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
		description: "lock when existing lock is for different repo owner should succeed",
		sequence: []interface{}{
			regularLock,
			LockCommand{
				keyWithNewField("RepoOwner", "different"),
				locking.TryLockResponse{
					LockAcquired: true,
					LockingRun: keyWithNewField("RepoOwner", "different"),
				},
			},
		},
	},
	{
		description: "lock when existing lock is for different repo should succeed",
		sequence: []interface{}{
			regularLock,
			LockCommand{
				keyWithNewField("RepoName", "different"),
				locking.TryLockResponse{
					LockAcquired: true,
					LockingRun: keyWithNewField("RepoName", "different"),
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
				"f13c9d87dc7156638cc075f0164d628f1338d23b17296dd391d922ea6fe6e9f8",
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

		db := NewTestDB()
		b := locking.NewBoltDBLockManager(db, LockBucket)

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

