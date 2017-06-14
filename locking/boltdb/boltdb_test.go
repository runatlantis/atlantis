package boltdb_test

import (
	"github.com/hootsuite/atlantis/locking"
	. "github.com/hootsuite/atlantis/testing_util"

	"github.com/boltdb/bolt"
	"os"
	"testing"
	"io/ioutil"
	"github.com/pkg/errors"

	"time"
	"github.com/hootsuite/atlantis/locking/boltdb"
)

var lockBucket = "bucket"
var run = locking.Run{
	RepoFullName: "owner/repo",
	Path: "parent/child",
	Env: "default",
	PullNum: 1,
	User: "user",
	Timestamp: time.Now(),
}

func TestListNoLocks(t *testing.T) {
	t.Log("listing locks when there are none should return an empty list")
	db, b := newTestDB()
	defer cleanupDB(db)
	ls, err := b.ListLocks()
	Ok(t, err)
	Equals(t, 0, len(ls))
}

func TestListOneLock(t *testing.T) {
	t.Log("listing locks when there is one should return it")
	db, b := newTestDB()
	defer cleanupDB(db)
	_, err := b.TryLock(run)
	Ok(t, err)
	ls, err := b.ListLocks()
	Ok(t, err)
	Equals(t, 1, len(ls))
	Equals(t, run, ls[run.StateKey()])
}

func TestListMultipleLocks(t *testing.T) {
	t.Log("listing locks when there are multiple should return them")
	db, b := newTestDB()
	defer cleanupDB(db)

	// add multiple locks
	_, err := b.TryLock(run)
	Ok(t, err)

	run2 := run
	run2.RepoFullName = "different"
	_, err = b.TryLock(run2)
	Ok(t, err)

	run3 := run
	run3.RepoFullName = "different again"
	_, err = b.TryLock(run3)
	Ok(t, err)

	run4 := run
	run4.RepoFullName = "different again again"
	_, err = b.TryLock(run4)
	Ok(t, err)

	ls, err := b.ListLocks()
	Ok(t, err)
	Equals(t, 4, len(ls))
	Equals(t, run, ls[run.StateKey()])
	Equals(t, run2, ls[run2.StateKey()])
	Equals(t, run3, ls[run3.StateKey()])
	Equals(t, run4, ls[run4.StateKey()])
}

func TestListAddRemove(t *testing.T) {
	t.Log("listing after adding and removing should return none")
	db, b := newTestDB()
	defer cleanupDB(db)
	_, err := b.TryLock(run)
	Ok(t, err)
	b.Unlock(run.StateKey())

	ls, err := b.ListLocks()
	Ok(t, err)
	Equals(t, 0, len(ls))
}

func TestLockingNoLocks(t *testing.T) {
	t.Log("with no locks yet, lock should succeed")
	db, b := newTestDB()
	defer cleanupDB(db)
	r, err := b.TryLock(run)
	Ok(t, err)
	Equals(t, true, r.LockAcquired)
	Equals(t, run, r.LockingRun)
	Equals(t, run.StateKey(), r.LockID)
}

func TestLockingExistingLock(t *testing.T) {
	t.Log("if there is an existing lock, lock should...")
	db, b := newTestDB()
	defer cleanupDB(db)
	_, err := b.TryLock(run)
	Ok(t, err)

	t.Log("...succeed if the new run has a different path")
	{
		new := run
		new.Path = "different/path"
		r, err := b.TryLock(new)
		Ok(t, err)
		Equals(t, true, r.LockAcquired)
		Equals(t, new, r.LockingRun)
		Equals(t, new.StateKey(), r.LockID)
	}

	t.Log("...succeed if the new run has a different environment")
	{
		new := run
		new.Env = "different-env"
		r, err := b.TryLock(new)
		Ok(t, err)
		Equals(t, true, r.LockAcquired)
		Equals(t, new, r.LockingRun)
		Equals(t, new.StateKey(), r.LockID)
	}

	t.Log("...succeed if the new run has a different repoName")
	{
		new := run
		new.RepoFullName = "new/repo"
		r, err := b.TryLock(new)
		Ok(t, err)
		Equals(t, true, r.LockAcquired)
		Equals(t, new, r.LockingRun)
		Equals(t, new.StateKey(), r.LockID)
	}

	t.Log("...not succeed if the new run only has a different pullNum")
	{
		new := run
		new.PullNum = run.PullNum + 1
		r, err := b.TryLock(new)
		Ok(t, err)
		Equals(t, false, r.LockAcquired)
		Equals(t, run, r.LockingRun)
		Equals(t, run.StateKey(), r.LockID)
	}
}

func TestUnlockingNoLocks(t *testing.T) {
	t.Log("unlocking with no locks should succeed")
	db, b := newTestDB()
	defer cleanupDB(db)

	Ok(t, b.Unlock("any-lock-id"))
}

func TestUnlocking(t *testing.T) {
	t.Log("unlocking with an existing lock should succeed")
	db, b := newTestDB()
	defer cleanupDB(db)

	b.TryLock(run)
	Ok(t, b.Unlock(run.StateKey()))

	// should be no locks listed
	ls, err := b.ListLocks()
	Ok(t, err)
	Equals(t, 0, len(ls))

	// should be able to re-lock that repo with a new pull num
	new := run
	new.PullNum = run.PullNum + 1
	r, err := b.TryLock(new)
	Ok(t, err)
	Equals(t, true, r.LockAcquired)
}

func TestUnlockingMultiple(t *testing.T) {
	t.Log("unlocking and locking multiple locks should succeed")
	db, b := newTestDB()
	defer cleanupDB(db)

	b.TryLock(run)

	new := run
	new.RepoFullName = "new/repo"
	b.TryLock(new)

	new2 := run
	new2.Path = "new/path"
	b.TryLock(new2)

	new3 := run
	new3.Env = "new/env"
	b.TryLock(new3)

	// now try and unlock them
	Ok(t, b.Unlock(new3.StateKey()))
	Ok(t, b.Unlock(new2.StateKey()))
	Ok(t, b.Unlock(new.StateKey()))
	Ok(t, b.Unlock(run.StateKey()))

	// should be none left
	ls, err := b.ListLocks()
	Ok(t, err)
	Equals(t, 0, len(ls))
}

func TestFindLocksNone(t *testing.T) {
	t.Log("find should return no locks when there are none")
	db, b := newTestDB()
	defer cleanupDB(db)

	ls, err := b.FindLocksForPull("any/repo", 1)
	Ok(t, err)
	Equals(t, 0, len(ls))
}

func TestFindLocksOne(t *testing.T) {
	t.Log("with one lock find should...")
	db, b := newTestDB()
	defer cleanupDB(db)
	_, err := b.TryLock(run)
	Ok(t, err)

	t.Log("...return no locks from the same repo but different pull num")
	{
		ls, err := b.FindLocksForPull(run.RepoFullName, run.PullNum + 1)
		Ok(t, err)
		Equals(t, 0, len(ls))
	}
	t.Log("...return no locks from a different repo but the same pull num")
	{
		ls, err := b.FindLocksForPull(run.RepoFullName + "dif", run.PullNum)
		Ok(t, err)
		Equals(t, 0, len(ls))
	}
	t.Log("...return the one lock when called with that repo and pull num")
	{
		ls, err := b.FindLocksForPull(run.RepoFullName, run.PullNum)
		Ok(t, err)
		Equals(t, 1, len(ls))
		Equals(t, run.StateKey(), ls[0])
	}
}

func TestFindLocksAfterUnlock(t *testing.T) {
	t.Log("after locking and unlocking find should return no locks")
	db, b := newTestDB()
	defer cleanupDB(db)
	_, err := b.TryLock(run)
	Ok(t, err)
	Ok(t, b.Unlock(run.StateKey()))

	ls, err := b.FindLocksForPull(run.RepoFullName, run.PullNum)
	Ok(t, err)
	Equals(t, 0, len(ls))
}

func TestFindMultipleMatching(t *testing.T) {
	t.Log("find should return all matching lock ids")
	db, b := newTestDB()
	defer cleanupDB(db)
	_, err := b.TryLock(run)
	Ok(t, err)

	// add additional locks with the same repo and pull num but different paths/envs
	new := run
	new.Path = "dif/path"
	_, err = b.TryLock(new)
	Ok(t, err)
	new2 := run
	new2.Env = "new-env"
	_, err = b.TryLock(new2)
	Ok(t, err)

	// should get all of them back
	ls, err := b.FindLocksForPull(run.RepoFullName, run.PullNum)
	Ok(t, err)
	Equals(t, 3, len(ls))
	ContainsStr(t, run.StateKey(), ls)
	ContainsStr(t, new.StateKey(), ls)
	ContainsStr(t, new2.StateKey(), ls)
}

// newTestDB returns a TestDB using a temporary path.
func newTestDB() (*bolt.DB, *boltdb.Backend) {
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
		if _, err := tx.CreateBucketIfNotExists([]byte(lockBucket)); err != nil {
			return errors.Wrap(err, "failed to create bucket")
		}
		return nil
	}); err != nil {
		panic(errors.Wrap(err, "could not create bucket"))
	}
	b, _ := boltdb.NewWithDB(db, lockBucket)
	return db, b
}

func cleanupDB(db *bolt.DB) {
	os.Remove(db.Path())
	db.Close()
}

