package boltdb_test

import (
	. "github.com/hootsuite/atlantis/testing_util"

	"io/ioutil"
	"os"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"

	"time"

	"github.com/hootsuite/atlantis/server/events/locking/boltdb"
	"github.com/hootsuite/atlantis/server/events/models"
)

var lockBucket = "bucket"
var project = models.NewProject("owner/repo", "parent/child")
var env = "default"
var pullNum = 1
var lock = models.ProjectLock{
	Pull: models.PullRequest{
		Num: pullNum,
	},
	User: models.User{
		Username: "lkysow",
	},
	Env:     env,
	Project: project,
	Time:    time.Now(),
}

func TestListNoLocks(t *testing.T) {
	t.Log("listing locks when there are none should return an empty list")
	db, b := newTestDB()
	defer cleanupDB(db)
	ls, err := b.List()
	Ok(t, err)
	Equals(t, 0, len(ls))
}

func TestListOneLock(t *testing.T) {
	t.Log("listing locks when there is one should return it")
	db, b := newTestDB()
	defer cleanupDB(db)
	_, _, err := b.TryLock(lock)
	Ok(t, err)
	ls, err := b.List()
	Ok(t, err)
	Equals(t, 1, len(ls))
}

func TestListMultipleLocks(t *testing.T) {
	t.Log("listing locks when there are multiple should return them")
	db, b := newTestDB()
	defer cleanupDB(db)

	// add multiple locks
	repos := []string{
		"owner/repo1",
		"owner/repo2",
		"owner/repo3",
		"owner/repo4",
	}

	for _, r := range repos {
		newLock := lock
		newLock.Project = models.NewProject(r, "path")
		_, _, err := b.TryLock(newLock)
		Ok(t, err)
	}
	ls, err := b.List()
	Ok(t, err)
	Equals(t, 4, len(ls))
	for _, r := range repos {
		found := false
		for _, l := range ls {
			if l.Project.RepoFullName == r {
				found = true
			}
		}
		Assert(t, found == true, "expected %s in %v", r, ls)
	}
}

func TestListAddRemove(t *testing.T) {
	t.Log("listing after adding and removing should return none")
	db, b := newTestDB()
	defer cleanupDB(db)
	_, _, err := b.TryLock(lock)
	Ok(t, err)
	b.Unlock(project, env)

	ls, err := b.List()
	Ok(t, err)
	Equals(t, 0, len(ls))
}

func TestLockingNoLocks(t *testing.T) {
	t.Log("with no locks yet, lock should succeed")
	db, b := newTestDB()
	defer cleanupDB(db)
	acquired, currLock, err := b.TryLock(lock)
	Ok(t, err)
	Equals(t, true, acquired)
	Equals(t, lock, currLock)
}

func TestLockingExistingLock(t *testing.T) {
	t.Log("if there is an existing lock, lock should...")
	db, b := newTestDB()
	defer cleanupDB(db)
	_, _, err := b.TryLock(lock)
	Ok(t, err)

	t.Log("...succeed if the new project has a different path")
	{
		newLock := lock
		newLock.Project = models.NewProject(project.RepoFullName, "different/path")
		acquired, currLock, err := b.TryLock(newLock)
		Ok(t, err)
		Equals(t, true, acquired)
		Equals(t, pullNum, currLock.Pull.Num)
	}

	t.Log("...succeed if the new project has a different environment")
	{
		newLock := lock
		newLock.Env = "different-env"
		acquired, currLock, err := b.TryLock(newLock)
		Ok(t, err)
		Equals(t, true, acquired)
		Equals(t, newLock, currLock)
	}

	t.Log("...succeed if the new project has a different repoName")
	{
		newLock := lock
		newLock.Project = models.NewProject("different/repo", project.Path)
		acquired, currLock, err := b.TryLock(newLock)
		Ok(t, err)
		Equals(t, true, acquired)
		Equals(t, newLock, currLock)
	}

	t.Log("...not succeed if the new project only has a different pullNum")
	{
		newLock := lock
		newLock.Pull.Num = lock.Pull.Num + 1
		acquired, currLock, err := b.TryLock(newLock)
		Ok(t, err)
		Equals(t, false, acquired)
		Equals(t, currLock.Pull.Num, pullNum)
	}
}

func TestUnlockingNoLocks(t *testing.T) {
	t.Log("unlocking with no locks should succeed")
	db, b := newTestDB()
	defer cleanupDB(db)
	_, err := b.Unlock(project, env)

	Ok(t, err)
}

func TestUnlocking(t *testing.T) {
	t.Log("unlocking with an existing lock should succeed")
	db, b := newTestDB()
	defer cleanupDB(db)

	b.TryLock(lock)
	_, err := b.Unlock(project, env)
	Ok(t, err)

	// should be no locks listed
	ls, err := b.List()
	Ok(t, err)
	Equals(t, 0, len(ls))

	// should be able to re-lock that repo with a new pull num
	newLock := lock
	newLock.Pull.Num = lock.Pull.Num + 1
	acquired, currLock, err := b.TryLock(newLock)
	Ok(t, err)
	Equals(t, true, acquired)
	Equals(t, newLock, currLock)
}

func TestUnlockingMultiple(t *testing.T) {
	t.Log("unlocking and locking multiple locks should succeed")
	db, b := newTestDB()
	defer cleanupDB(db)

	b.TryLock(lock)

	new := lock
	new.Project.RepoFullName = "new/repo"
	b.TryLock(new)

	new2 := lock
	new2.Project.Path = "new/path"
	b.TryLock(new2)

	new3 := lock
	new3.Env = "new-env"
	b.TryLock(new3)

	// now try and unlock them
	_, err := b.Unlock(new3.Project, new3.Env)
	Ok(t, err)
	_, err = b.Unlock(new2.Project, env)
	Ok(t, err)
	_, err = b.Unlock(new.Project, env)
	Ok(t, err)
	_, err = b.Unlock(project, env)
	Ok(t, err)

	// should be none left
	ls, err := b.List()
	Ok(t, err)
	Equals(t, 0, len(ls))
}

func TestUnlockByPullNone(t *testing.T) {
	t.Log("UnlockByPull should be successful when there are no locks")
	db, b := newTestDB()
	defer cleanupDB(db)

	_, err := b.UnlockByPull("any/repo", 1)
	Ok(t, err)
}

func TestUnlockByPullOne(t *testing.T) {
	t.Log("with one lock, UnlockByPull should...")
	db, b := newTestDB()
	defer cleanupDB(db)
	_, _, err := b.TryLock(lock)
	Ok(t, err)

	t.Log("...delete nothing when its the same repo but a different pull")
	{
		_, err := b.UnlockByPull(project.RepoFullName, pullNum+1)
		Ok(t, err)
		ls, err := b.List()
		Ok(t, err)
		Equals(t, 1, len(ls))
	}
	t.Log("...delete nothing when its the same pull but a different repo")
	{
		_, err := b.UnlockByPull("different/repo", pullNum)
		Ok(t, err)
		ls, err := b.List()
		Ok(t, err)
		Equals(t, 1, len(ls))
	}
	t.Log("...delete the lock when its the same repo and pull")
	{
		_, err := b.UnlockByPull(project.RepoFullName, pullNum)
		Ok(t, err)
		ls, err := b.List()
		Ok(t, err)
		Equals(t, 0, len(ls))
	}
}

func TestUnlockByPullAfterUnlock(t *testing.T) {
	t.Log("after locking and unlocking, UnlockByPull should be successful")
	db, b := newTestDB()
	defer cleanupDB(db)
	_, _, err := b.TryLock(lock)
	Ok(t, err)
	_, err = b.Unlock(project, env)
	Ok(t, err)

	_, err = b.UnlockByPull(project.RepoFullName, pullNum)
	Ok(t, err)
	ls, err := b.List()
	Ok(t, err)
	Equals(t, 0, len(ls))
}

func TestUnlockByPullMatching(t *testing.T) {
	t.Log("UnlockByPull should delete all locks in that repo and pull num")
	db, b := newTestDB()
	defer cleanupDB(db)
	_, _, err := b.TryLock(lock)
	Ok(t, err)

	// add additional locks with the same repo and pull num but different paths/envs
	new := lock
	new.Project.Path = "dif/path"
	_, _, err = b.TryLock(new)
	Ok(t, err)
	new2 := lock
	new2.Env = "new-env"
	_, _, err = b.TryLock(new2)
	Ok(t, err)

	// there should now be 3
	ls, err := b.List()
	Ok(t, err)
	Equals(t, 3, len(ls))

	// should all be unlocked
	_, err = b.UnlockByPull(project.RepoFullName, pullNum)
	Ok(t, err)
	ls, err = b.List()
	Ok(t, err)
	Equals(t, 0, len(ls))
}

func TestGetLockNotThere(t *testing.T) {
	t.Log("getting a lock that doesn't exist should return a nil pointer")
	db, b := newTestDB()
	defer cleanupDB(db)
	l, err := b.GetLock(project, env)
	Ok(t, err)
	Equals(t, (*models.ProjectLock)(nil), l)
}

func TestGetLock(t *testing.T) {
	t.Log("getting a lock should return the lock")
	db, b := newTestDB()
	defer cleanupDB(db)
	_, _, err := b.TryLock(lock)
	Ok(t, err)

	l, err := b.GetLock(project, env)
	Ok(t, err)
	// can't compare against time so doing each field
	Equals(t, lock.Project, l.Project)
	Equals(t, lock.Env, l.Env)
	Equals(t, lock.Pull, l.Pull)
	Equals(t, lock.User, l.User)
}

// newTestDB returns a TestDB using a temporary path.
func newTestDB() (*bolt.DB, *boltdb.BoltLocker) {
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
