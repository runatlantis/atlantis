package boltdb_test

import (
	. "github.com/hootsuite/atlantis/testing_util"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/hootsuite/atlantis/locking/boltdb"
	"github.com/hootsuite/atlantis/models"
)

var lockBucket = "bucket"
var project = models.NewProject("owner/repo", "parent/child")
var env = "default"
var pullNum = 1

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
	_, _, err := b.TryLock(project, env, pullNum)
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
		_, _, err := b.TryLock(models.NewProject(r, "path"), env, pullNum)
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
	_, _, err := b.TryLock(project, env, pullNum)
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
	acquired, lockingPullNum, err := b.TryLock(project, env, pullNum)
	Ok(t, err)
	Equals(t, true, acquired)
	Equals(t, pullNum, lockingPullNum)
}

func TestLockingExistingLock(t *testing.T) {
	t.Log("if there is an existing lock, lock should...")
	db, b := newTestDB()
	defer cleanupDB(db)
	_, _, err := b.TryLock(project, env, pullNum)
	Ok(t, err)

	t.Log("...succeed if the new project has a different path")
	{
		acquired, lockingPull, err := b.TryLock(models.NewProject(project.RepoFullName, "different/path"), env, pullNum)
		Ok(t, err)
		Equals(t, true, acquired)
		Equals(t, pullNum, lockingPull)
	}

	t.Log("...succeed if the new project has a different environment")
	{
		acquired, lockingPull, err := b.TryLock(project, "different-env", pullNum)
		Ok(t, err)
		Equals(t, true, acquired)
		Equals(t, pullNum, lockingPull)
	}

	t.Log("...succeed if the new project has a different repoName")
	{
		acquired, lockingPull, err := b.TryLock(models.NewProject("different/repo", project.Path), env, pullNum)
		Ok(t, err)
		Equals(t, true, acquired)
		Equals(t, pullNum, lockingPull)
	}

	t.Log("...not succeed if the new project only has a different pullNum")
	{
		newPull := pullNum + 1
		acquired, lockingPull, err := b.TryLock(project, env, newPull)
		Ok(t, err)
		Equals(t, false, acquired)
		Equals(t, lockingPull, pullNum)
	}
}

func TestUnlockingNoLocks(t *testing.T) {
	t.Log("unlocking with no locks should succeed")
	db, b := newTestDB()
	defer cleanupDB(db)

	Ok(t, b.Unlock(project, env))
}

func TestUnlocking(t *testing.T) {
	t.Log("unlocking with an existing lock should succeed")
	db, b := newTestDB()
	defer cleanupDB(db)

	b.TryLock(project, env, pullNum)
	Ok(t, b.Unlock(project, env))

	// should be no locks listed
	ls, err := b.List()
	Ok(t, err)
	Equals(t, 0, len(ls))

	// should be able to re-lock that repo with a new pull num
	newPull := pullNum + 1
	acquired, lockingPull, err := b.TryLock(project, env, newPull)
	Ok(t, err)
	Equals(t, true, acquired)
	Equals(t, newPull, lockingPull)
}

func TestUnlockingMultiple(t *testing.T) {
	t.Log("unlocking and locking multiple locks should succeed")
	db, b := newTestDB()
	defer cleanupDB(db)

	b.TryLock(project, env, pullNum)

	new := project
	new.RepoFullName = "new/repo"
	b.TryLock(project, env, pullNum)

	new2 := project
	new2.Path = "new/path"
	b.TryLock(project, env, pullNum)

	new3Env := "new-env"
	b.TryLock(project, new3Env, pullNum)

	// now try and unlock them
	Ok(t, b.Unlock(project, new3Env))
	Ok(t, b.Unlock(new2, env))
	Ok(t, b.Unlock(new, env))
	Ok(t, b.Unlock(project, env))

	// should be none left
	ls, err := b.List()
	Ok(t, err)
	Equals(t, 0, len(ls))
}

func TestUnlockByPullNone(t *testing.T) {
	t.Log("UnlockByPull should be successful when there are no locks")
	db, b := newTestDB()
	defer cleanupDB(db)

	err := b.UnlockByPull("any/repo", 1)
	Ok(t, err)
}

func TestUnlockByPullOne(t *testing.T) {
	t.Log("with one lock, UnlockByPull should...")
	db, b := newTestDB()
	defer cleanupDB(db)
	_, _, err := b.TryLock(project, env, pullNum)
	Ok(t, err)

	t.Log("...delete nothing when its the same repo but a different pull")
	{
		err := b.UnlockByPull(project.RepoFullName, pullNum+1)
		Ok(t, err)
		ls, err := b.List()
		Ok(t, err)
		Equals(t, 1, len(ls))
	}
	t.Log("...delete nothing when its the same pull but a different repo")
	{
		err := b.UnlockByPull("different/repo", pullNum)
		Ok(t, err)
		ls, err := b.List()
		Ok(t, err)
		Equals(t, 1, len(ls))
	}
	t.Log("...delete the lock when its the same repo and pull")
	{
		err := b.UnlockByPull(project.RepoFullName, pullNum)
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
	_, _, err := b.TryLock(project, env, pullNum)
	Ok(t, err)
	Ok(t, b.Unlock(project, env))

	err = b.UnlockByPull(project.RepoFullName, pullNum)
	Ok(t, err)
	ls, err := b.List()
	Ok(t, err)
	Equals(t, 0, len(ls))
}

func TestUnlockByPullMatching(t *testing.T) {
	t.Log("UnlockByPull should delete all locks in that repo and pull num")
	db, b := newTestDB()
	defer cleanupDB(db)
	_, _, err := b.TryLock(project, env, pullNum)
	Ok(t, err)

	// add additional locks with the same repo and pull num but different paths/envs
	new := project
	new.Path = "dif/path"
	_, _, err = b.TryLock(new, env, pullNum)
	Ok(t, err)
	_, _, err = b.TryLock(project, "new-env", pullNum)
	Ok(t, err)

	// there should now be 3
	ls, err := b.List()
	Ok(t, err)
	Equals(t, 3, len(ls))

	// should all be unlocked
	err = b.UnlockByPull(project.RepoFullName, pullNum)
	Ok(t, err)
	ls, err = b.List()
	Ok(t, err)
	Equals(t, 0, len(ls))
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
