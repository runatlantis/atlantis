// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package boltdb_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/core/boltdb"
	coredb "github.com/runatlantis/atlantis/server/core/db"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
	bolt "go.etcd.io/bbolt"
)

var lockBucket = "bucket"
var configBucket = "configBucket"
var project = models.NewProject("owner/repo", "parent/child", "")
var workspace = "default"
var pullNum = 1
var lock = models.ProjectLock{
	Pull: models.PullRequest{
		Num: pullNum,
	},
	User: models.User{
		Username: "lkysow",
	},
	Workspace: workspace,
	Project:   project,
	Time:      time.Now(),
}

func TestLockCommandNotSet(t *testing.T) {
	t.Log("retrieving apply lock when there are none should return empty LockCommand")
	db, b := newTestDB()
	defer cleanupDB(db)
	exists, err := b.CheckCommandLock(command.Apply)
	Ok(t, err)
	Assert(t, exists == nil, "exp nil")
}

func TestLockCommandEnabled(t *testing.T) {
	t.Log("setting the apply lock")
	db, b := newTestDB()
	defer cleanupDB(db)
	timeNow := time.Now()
	_, err := b.LockCommand(command.Apply, timeNow)
	Ok(t, err)

	config, err := b.CheckCommandLock(command.Apply)
	Ok(t, err)
	Equals(t, true, config.IsLocked())
}

func TestLockCommandFail(t *testing.T) {
	t.Log("setting the apply lock")
	db, b := newTestDB()
	defer cleanupDB(db)
	timeNow := time.Now()
	_, err := b.LockCommand(command.Apply, timeNow)
	Ok(t, err)

	_, err = b.LockCommand(command.Apply, timeNow)
	ErrEquals(t, "db transaction failed: lock already exists", err)
}

func TestUnlockCommandDisabled(t *testing.T) {
	t.Log("unsetting the apply lock")
	db, b := newTestDB()
	defer cleanupDB(db)
	timeNow := time.Now()
	_, err := b.LockCommand(command.Apply, timeNow)
	Ok(t, err)

	config, err := b.CheckCommandLock(command.Apply)
	Ok(t, err)
	Equals(t, true, config.IsLocked())

	err = b.UnlockCommand(command.Apply)
	Ok(t, err)

	config, err = b.CheckCommandLock(command.Apply)
	Ok(t, err)
	Assert(t, config == nil, "exp nil object")
}

func TestMigrationOldLockKeysToNewFormat(t *testing.T) {
	t.Log("migration should convert old format keys to new format with project name")

	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create a database file manually with an old format key
	dbPath := tmpDir + "/atlantis.db"
	boltDB, err := bolt.Open(dbPath, 0600, nil)
	Ok(t, err)

	// Create buckets
	err = boltDB.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte("runLocks")); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists([]byte("pulls")); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists([]byte("globalLocks")); err != nil {
			return err
		}
		return nil
	})
	Ok(t, err)

	// Create a lock in old format: {repoFullName}/{path}/{workspace}
	oldKey := "owner/repo/path/default"
	oldProject := models.NewProject("owner/repo", "path", "myproject")
	oldLock := models.ProjectLock{
		Pull:      models.PullRequest{Num: 1},
		User:      models.User{Username: "testuser"},
		Workspace: "default",
		Project:   oldProject,
		Time:      time.Now(),
	}

	oldLockSerialized, err := json.Marshal(oldLock)
	Ok(t, err)

	// Insert old format lock
	err = boltDB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("runLocks"))
		return bucket.Put([]byte(oldKey), oldLockSerialized)
	})
	Ok(t, err)

	// Verify old key exists
	err = boltDB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("runLocks"))
		val := bucket.Get([]byte(oldKey))
		Assert(t, val != nil, "old key should exist before migration")
		return nil
	})
	Ok(t, err)

	// Close the database
	boltDB.Close()

	// Now open with boltdb.New which should trigger the migration
	b, err := boltdb.New(tmpDir)
	Ok(t, err)
	defer b.Close()

	// List all locks
	allLocks, err := b.List()
	Ok(t, err)
	Assert(t, len(allLocks) == 1, "should have 1 lock after migration")

	// Verify the lock can be retrieved using the GetLock method
	// which uses the new key format internally
	projectWithName := models.NewProject("owner/repo", "path", "myproject")
	retrievedLock, err := b.GetLock(projectWithName, "default")
	Ok(t, err)
	Assert(t, retrievedLock != nil, "lock should exist with new key format")
	Equals(t, "owner/repo", retrievedLock.Project.RepoFullName)
	Equals(t, "path", retrievedLock.Project.Path)
	Equals(t, "myproject", retrievedLock.Project.ProjectName)
	Equals(t, "default", retrievedLock.Workspace)
	Equals(t, "testuser", retrievedLock.User.Username)
}

func TestNoMigrationNeededForNewFormatKeys(t *testing.T) {
	t.Log("migration should not affect keys already in new format")

	// Create a temporary directory for the test database
	tmp := t.TempDir()
	db, err := boltdb.New(tmp)
	Ok(t, err)

	// Create a lock with the new format (includes project name)
	projectWithName := models.NewProject("owner/repo", "path", "projectName")
	newLock := models.ProjectLock{
		Pull:      models.PullRequest{Num: 1},
		User:      models.User{Username: "testuser"},
		Workspace: "default",
		Project:   projectWithName,
		Time:      time.Now(),
	}

	// Acquire lock using the new format
	acquired, _, err := db.TryLock(newLock)
	Ok(t, err)
	Assert(t, acquired, "should acquire lock")

	// Verify the lock can be retrieved immediately after creation
	retrievedLock, err := db.GetLock(projectWithName, "default")
	Ok(t, err)
	Assert(t, retrievedLock != nil, "lock should exist")
	Equals(t, "projectName", retrievedLock.Project.ProjectName)
	Equals(t, "testuser", retrievedLock.User.Username)

	// Close and reopen the database to trigger any migration logic
	db.Close()
	db, err = boltdb.New(tmp)
	Ok(t, err)
	defer db.Close()

	// Verify lock still exists after reopening (no migration should have changed it)
	retrievedLock, err = db.GetLock(projectWithName, "default")
	Ok(t, err)
	Assert(t, retrievedLock != nil, "lock should exist after migration")
	Equals(t, "projectName", retrievedLock.Project.ProjectName)
	Equals(t, "testuser", retrievedLock.User.Username)
}

func TestUnlockCommandFail(t *testing.T) {
	t.Log("setting the apply lock")
	db, b := newTestDB()
	defer cleanupDB(db)
	err := b.UnlockCommand(command.Apply)
	ErrEquals(t, "db transaction failed: no lock exists", err)
}

func TestMixedLocksPresent(t *testing.T) {
	db, b := newTestDB()
	defer cleanupDB(db)
	timeNow := time.Now()
	_, err := b.LockCommand(command.Apply, timeNow)
	Ok(t, err)

	_, _, err = b.TryLock(lock)
	Ok(t, err)
	ls, err := b.List()
	Ok(t, err)
	Equals(t, 1, len(ls))
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
		newLock.Project = models.NewProject(r, "path", "")
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
		Assert(t, found, "expected %s in %v", r, ls)
	}
}

func TestListAddRemove(t *testing.T) {
	t.Log("listing after adding and removing should return none")
	db, b := newTestDB()
	defer cleanupDB(db)
	_, _, err := b.TryLock(lock)
	Ok(t, err)
	_, err = b.Unlock(project, workspace)
	Ok(t, err)

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
		newLock.Project = models.NewProject(project.RepoFullName, "different/path", "")
		acquired, currLock, err := b.TryLock(newLock)
		Ok(t, err)
		Equals(t, true, acquired)
		Equals(t, pullNum, currLock.Pull.Num)
	}

	t.Log("...succeed if the new project has a different workspace")
	{
		newLock := lock
		newLock.Workspace = "different-workspace"
		acquired, currLock, err := b.TryLock(newLock)
		Ok(t, err)
		Equals(t, true, acquired)
		Equals(t, newLock, currLock)
	}

	t.Log("...succeed if the new project has a different repoName")
	{
		newLock := lock
		newLock.Project = models.NewProject("different/repo", project.Path, "")
		acquired, currLock, err := b.TryLock(newLock)
		Ok(t, err)
		Equals(t, true, acquired)
		Equals(t, newLock, currLock)
	}
	// TODO: How should we handle different name?
	/*
		t.Log("...succeed if the new project has a different name")
		{
			newLock := lock
			newLock.Project = models.NewProject(project.RepoFullName, project.Path, "different-name")
			acquired, currLock, err := b.TryLock(newLock)
			Ok(t, err)
			Equals(t, true, acquired)
			Equals(t, newLock, currLock)
		}
	*/

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
	_, err := b.Unlock(project, workspace)

	Ok(t, err)
}

func TestUnlocking(t *testing.T) {
	t.Log("unlocking with an existing lock should succeed")
	db, b := newTestDB()
	defer cleanupDB(db)

	_, _, err := b.TryLock(lock)
	Ok(t, err)
	_, err = b.Unlock(project, workspace)
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

	_, _, err := b.TryLock(lock)
	Ok(t, err)

	new1 := lock
	new1.Project.RepoFullName = "new/repo"
	_, _, err = b.TryLock(new1)
	Ok(t, err)

	new2 := lock
	new2.Project.Path = "new/path"
	_, _, err = b.TryLock(new2)
	Ok(t, err)

	new3 := lock
	new3.Workspace = "new-workspace"
	_, _, err = b.TryLock(new3)
	Ok(t, err)

	// now try and unlock them
	_, err = b.Unlock(new3.Project, new3.Workspace)
	Ok(t, err)
	_, err = b.Unlock(new2.Project, workspace)
	Ok(t, err)
	_, err = b.Unlock(new1.Project, workspace)
	Ok(t, err)
	_, err = b.Unlock(project, workspace)
	Ok(t, err)

	// should be none left
	ls, err := b.List()
	Ok(t, err)
	Equals(t, 0, len(ls))
}

func TestUnlockIfOwnedByPullMissingLock(t *testing.T) {
	t.Log("UnlockIfOwnedByPull should ignore missing locks")
	db, b := newTestDB()
	defer cleanupDB(db)

	deleted, err := b.UnlockIfOwnedByPull(project, workspace, pullNum)
	Ok(t, err)
	Equals(t, (*models.ProjectLock)(nil), deleted)
}

func TestUnlockIfOwnedByPullOtherPull(t *testing.T) {
	t.Log("UnlockIfOwnedByPull should not delete another pull's lock")
	db, b := newTestDB()
	defer cleanupDB(db)

	_, _, err := b.TryLock(lock)
	Ok(t, err)

	deleted, err := b.UnlockIfOwnedByPull(project, workspace, pullNum+1)
	Ok(t, err)
	Equals(t, (*models.ProjectLock)(nil), deleted)

	existing, err := b.GetLock(project, workspace)
	Ok(t, err)
	Assert(t, existing != nil, "expected lock to remain")
	Equals(t, pullNum, existing.Pull.Num)
}

func TestUnlockIfOwnedByPullCurrentPull(t *testing.T) {
	t.Log("UnlockIfOwnedByPull should delete the current pull's lock")
	db, b := newTestDB()
	defer cleanupDB(db)

	_, _, err := b.TryLock(lock)
	Ok(t, err)

	deleted, err := b.UnlockIfOwnedByPull(project, workspace, pullNum)
	Ok(t, err)
	Assert(t, deleted != nil, "expected deleted lock")
	Equals(t, lock.Project, deleted.Project)
	Equals(t, lock.Workspace, deleted.Workspace)
	Equals(t, lock.Pull, deleted.Pull)
	Equals(t, lock.User, deleted.User)

	existing, err := b.GetLock(project, workspace)
	Ok(t, err)
	Equals(t, (*models.ProjectLock)(nil), existing)
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
	_, err = b.Unlock(project, workspace)
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

	// add additional locks with the same repo and pull num but different paths/workspaces
	new1 := lock
	new1.Project.Path = "dif/path"
	_, _, err = b.TryLock(new1)
	Ok(t, err)
	new2 := lock
	new2.Workspace = "new-workspace"
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
	l, err := b.GetLock(project, workspace)
	Ok(t, err)
	Equals(t, (*models.ProjectLock)(nil), l)
}

func TestGetLock(t *testing.T) {
	t.Log("getting a lock should return the lock")
	db, b := newTestDB()
	defer cleanupDB(db)
	_, _, err := b.TryLock(lock)
	Ok(t, err)

	l, err := b.GetLock(project, workspace)
	Ok(t, err)
	// can't compare against time so doing each field
	Equals(t, lock.Project, l.Project)
	Equals(t, lock.Workspace, l.Workspace)
	Equals(t, lock.Pull, l.Pull)
	Equals(t, lock.User, l.User)
}

// Test we can create a status and then getCommandLock it.
func TestPullStatus_UpdateGet(t *testing.T) {
	b := newTestDB2(t)

	pull := models.PullRequest{
		Num:        1,
		HeadCommit: "sha",
		URL:        "url",
		HeadBranch: "head",
		BaseBranch: "base",
		Author:     "lkysow",
		State:      models.OpenPullState,
		BaseRepo: models.Repo{
			FullName:          "runatlantis/atlantis",
			Owner:             "runatlantis",
			Name:              "atlantis",
			CloneURL:          "clone-url",
			SanitizedCloneURL: "clone-url",
			VCSHost: models.VCSHost{
				Hostname: "github.com",
				Type:     models.Github,
			},
		},
	}
	status, err := b.UpdatePullWithResults(
		pull,
		[]command.ProjectResult{
			{
				Command:    command.Plan,
				RepoRelDir: ".",
				Workspace:  "default",
				ProjectCommandOutput: command.ProjectCommandOutput{
					Failure: "failure",
				},
			},
		})
	Ok(t, err)

	maybeStatus, err := b.GetPullStatus(pull)
	Ok(t, err)
	Equals(t, pull, maybeStatus.Pull) // nolint: staticcheck
	Equals(t, []models.ProjectStatus{
		{
			Workspace:   "default",
			RepoRelDir:  ".",
			ProjectName: "",
			Status:      models.ErroredPlanStatus,
		},
	}, status.Projects)
	b.Close()
}

// Test we can create a status, delete it, and then we shouldn't be able to getCommandLock
// it.
func TestPullStatus_UpdateDeleteGet(t *testing.T) {
	b := newTestDB2(t)

	pull := models.PullRequest{
		Num:        1,
		HeadCommit: "sha",
		URL:        "url",
		HeadBranch: "head",
		BaseBranch: "base",
		Author:     "lkysow",
		State:      models.OpenPullState,
		BaseRepo: models.Repo{
			FullName:          "runatlantis/atlantis",
			Owner:             "runatlantis",
			Name:              "atlantis",
			CloneURL:          "clone-url",
			SanitizedCloneURL: "clone-url",
			VCSHost: models.VCSHost{
				Hostname: "github.com",
				Type:     models.Github,
			},
		},
	}
	_, err := b.UpdatePullWithResults(
		pull,
		[]command.ProjectResult{
			{
				RepoRelDir: ".",
				Workspace:  "default",
				ProjectCommandOutput: command.ProjectCommandOutput{
					Failure: "failure",
				},
			},
		})
	Ok(t, err)

	err = b.DeletePullStatus(pull)
	Ok(t, err)

	maybeStatus, err := b.GetPullStatus(pull)
	Ok(t, err)
	Assert(t, maybeStatus == nil, "exp nil")
	b.Close()
}

// Test we can create a status, update a specific project's status within that
// pull status, and when we getCommandLock all the project statuses, that specific project
// should be updated.
func TestPullStatus_UpdateProject(t *testing.T) {
	b := newTestDB2(t)

	pull := models.PullRequest{
		Num:        1,
		HeadCommit: "sha",
		URL:        "url",
		HeadBranch: "head",
		BaseBranch: "base",
		Author:     "lkysow",
		State:      models.OpenPullState,
		BaseRepo: models.Repo{
			FullName:          "runatlantis/atlantis",
			Owner:             "runatlantis",
			Name:              "atlantis",
			CloneURL:          "clone-url",
			SanitizedCloneURL: "clone-url",
			VCSHost: models.VCSHost{
				Hostname: "github.com",
				Type:     models.Github,
			},
		},
	}
	_, err := b.UpdatePullWithResults(
		pull,
		[]command.ProjectResult{
			{
				RepoRelDir: ".",
				Workspace:  "default",
				ProjectCommandOutput: command.ProjectCommandOutput{
					Failure: "failure",
				},
			},
			{
				RepoRelDir: ".",
				Workspace:  "staging",
				ProjectCommandOutput: command.ProjectCommandOutput{
					ApplySuccess: "success!",
				},
			},
		})
	Ok(t, err)

	err = b.UpdateProjectStatus(pull, "default", ".", models.DiscardedPlanStatus)
	Ok(t, err)

	status, err := b.GetPullStatus(pull)
	Ok(t, err)
	Equals(t, pull, status.Pull) // nolint: staticcheck
	Equals(t, []models.ProjectStatus{
		{
			Workspace:   "default",
			RepoRelDir:  ".",
			ProjectName: "",
			Status:      models.DiscardedPlanStatus,
		},
		{
			Workspace:   "staging",
			RepoRelDir:  ".",
			ProjectName: "",
			Status:      models.AppliedPlanStatus,
		},
	}, status.Projects) // nolint: staticcheck
	b.Close()
}

// Test that if we update an existing pull status and our new status is for a
// different HeadSHA, that we just overwrite the old status.
func TestPullStatus_UpdateNewCommit(t *testing.T) {
	b := newTestDB2(t)

	pull := models.PullRequest{
		Num:        1,
		HeadCommit: "sha",
		URL:        "url",
		HeadBranch: "head",
		BaseBranch: "base",
		Author:     "lkysow",
		State:      models.OpenPullState,
		BaseRepo: models.Repo{
			FullName:          "runatlantis/atlantis",
			Owner:             "runatlantis",
			Name:              "atlantis",
			CloneURL:          "clone-url",
			SanitizedCloneURL: "clone-url",
			VCSHost: models.VCSHost{
				Hostname: "github.com",
				Type:     models.Github,
			},
		},
	}
	_, err := b.UpdatePullWithResults(
		pull,
		[]command.ProjectResult{
			{
				RepoRelDir: ".",
				Workspace:  "default",
				ProjectCommandOutput: command.ProjectCommandOutput{
					Failure: "failure",
				},
			},
		})
	Ok(t, err)

	pull.HeadCommit = "newsha"
	status, err := b.UpdatePullWithResults(pull,
		[]command.ProjectResult{
			{
				RepoRelDir: ".",
				Workspace:  "staging",
				ProjectCommandOutput: command.ProjectCommandOutput{
					ApplySuccess: "success!",
				},
			},
		})

	Ok(t, err)
	Equals(t, 1, len(status.Projects))

	maybeStatus, err := b.GetPullStatus(pull)
	Ok(t, err)
	Equals(t, pull, maybeStatus.Pull)
	Equals(t, []models.ProjectStatus{
		{
			Workspace:   "staging",
			RepoRelDir:  ".",
			ProjectName: "",
			Status:      models.AppliedPlanStatus,
		},
	}, maybeStatus.Projects)
	b.Close()
}

func TestPullStatus_UpdateSameCommitNewBaseBranch(t *testing.T) {
	b := newTestDB2(t)

	pull := models.PullRequest{
		Num:        1,
		HeadCommit: "sha",
		URL:        "url",
		HeadBranch: "head",
		BaseBranch: "base",
		Author:     "lkysow",
		State:      models.OpenPullState,
		BaseRepo: models.Repo{
			FullName:          "runatlantis/atlantis",
			Owner:             "runatlantis",
			Name:              "atlantis",
			CloneURL:          "clone-url",
			SanitizedCloneURL: "clone-url",
			VCSHost: models.VCSHost{
				Hostname: "github.com",
				Type:     models.Github,
			},
		},
	}
	_, err := b.UpdatePullWithResults(
		pull,
		[]command.ProjectResult{
			{
				Command:    command.Plan,
				RepoRelDir: "old-base",
				Workspace:  "default",
				ProjectCommandOutput: command.ProjectCommandOutput{
					PlanSuccess: &models.PlanSuccess{},
				},
			},
		})
	Ok(t, err)

	pull.BaseBranch = "release"
	status, err := b.UpdatePullWithResults(pull,
		[]command.ProjectResult{
			{
				Command:    command.Plan,
				RepoRelDir: ".",
				Workspace:  "staging",
				ProjectCommandOutput: command.ProjectCommandOutput{
					PlanSuccess: &models.PlanSuccess{},
				},
			},
		})

	Ok(t, err)
	Equals(t, 1, len(status.Projects))

	maybeStatus, err := b.GetPullStatus(pull)
	Ok(t, err)
	Equals(t, pull, maybeStatus.Pull)
	Equals(t, []models.ProjectStatus{
		{
			Workspace:   "staging",
			RepoRelDir:  ".",
			ProjectName: "",
			Status:      models.PlannedPlanStatus,
		},
	}, maybeStatus.Projects)
	b.Close()
}

func TestBoltDB_SameCommitBackfillBaseDoesNotPromoteLegacyOldBaseProjects(t *testing.T) {
	b := newTestDB2(t)

	pull := models.PullRequest{
		Num:        1,
		HeadCommit: "sha",
		URL:        "url",
		HeadBranch: "head",
		Author:     "lkysow",
		State:      models.OpenPullState,
		BaseRepo: models.Repo{
			FullName:          "runatlantis/atlantis",
			Owner:             "runatlantis",
			Name:              "atlantis",
			CloneURL:          "clone-url",
			SanitizedCloneURL: "clone-url",
			VCSHost: models.VCSHost{
				Hostname: "github.com",
				Type:     models.Github,
			},
		},
	}
	_, err := b.UpdatePullWithResults(
		pull,
		[]command.ProjectResult{
			{
				Command:    command.Plan,
				RepoRelDir: "old-base",
				Workspace:  "default",
				ProjectCommandOutput: command.ProjectCommandOutput{
					PlanSuccess: &models.PlanSuccess{},
				},
			},
		})
	Ok(t, err)

	pull.BaseBranch = "main"
	status, err := b.UpdatePullWithResults(pull,
		[]command.ProjectResult{
			{
				Command:    command.Plan,
				RepoRelDir: ".",
				Workspace:  "staging",
				ProjectCommandOutput: command.ProjectCommandOutput{
					PlanSuccess: &models.PlanSuccess{},
				},
			},
		})

	Ok(t, err)
	Equals(t, "main", status.Pull.BaseBranch)
	Equals(t, []models.ProjectStatus{
		{
			Workspace:   "staging",
			RepoRelDir:  ".",
			ProjectName: "",
			Status:      models.PlannedPlanStatus,
		},
	}, status.Projects)

	maybeStatus, err := b.GetPullStatus(pull)
	Ok(t, err)
	Equals(t, "main", maybeStatus.Pull.BaseBranch)
	Equals(t, status.Projects, maybeStatus.Projects)
	b.Close()
}

// Test that if we update an existing pull status via Apply and our new status is for a
// the same commit, that we merge the statuses.
func TestPullStatus_UpdateMerge_Apply(t *testing.T) {
	b := newTestDB2(t)

	pull := models.PullRequest{
		Num:        1,
		HeadCommit: "sha",
		URL:        "url",
		HeadBranch: "head",
		BaseBranch: "base",
		Author:     "lkysow",
		State:      models.OpenPullState,
		BaseRepo: models.Repo{
			FullName:          "runatlantis/atlantis",
			Owner:             "runatlantis",
			Name:              "atlantis",
			CloneURL:          "clone-url",
			SanitizedCloneURL: "clone-url",
			VCSHost: models.VCSHost{
				Hostname: "github.com",
				Type:     models.Github,
			},
		},
	}
	_, err := b.UpdatePullWithResults(
		pull,
		[]command.ProjectResult{
			{
				Command:    command.Plan,
				RepoRelDir: "mergeme",
				Workspace:  "default",
				ProjectCommandOutput: command.ProjectCommandOutput{
					Failure: "failure",
				},
			},
			{
				Command:     command.Plan,
				RepoRelDir:  "projectname",
				Workspace:   "default",
				ProjectName: "projectname",
				ProjectCommandOutput: command.ProjectCommandOutput{
					Failure: "failure",
				},
			},
			{
				Command:    command.Plan,
				RepoRelDir: "staythesame",
				Workspace:  "default",
				ProjectCommandOutput: command.ProjectCommandOutput{
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "tf out",
						LockURL:         "lock-url",
						RePlanCmd:       "plan command",
						ApplyCmd:        "apply command",
					},
				},
			},
		})
	Ok(t, err)

	updateStatus, err := b.UpdatePullWithResults(pull,
		[]command.ProjectResult{
			{
				Command:    command.Apply,
				RepoRelDir: "mergeme",
				Workspace:  "default",
				ProjectCommandOutput: command.ProjectCommandOutput{
					ApplySuccess: "applied!",
				},
			},
			{
				Command:     command.Apply,
				RepoRelDir:  "projectname",
				Workspace:   "default",
				ProjectName: "projectname",
				ProjectCommandOutput: command.ProjectCommandOutput{
					Error: errors.New("apply error"),
				},
			},
			{
				Command:    command.Apply,
				RepoRelDir: "newresult",
				Workspace:  "default",
				ProjectCommandOutput: command.ProjectCommandOutput{
					ApplySuccess: "success!",
				},
			},
		})
	Ok(t, err)

	getStatus, err := b.GetPullStatus(pull)
	Ok(t, err)

	// Test both the pull state returned from the update call *and* the getCommandLock
	// call.
	for _, s := range []models.PullStatus{updateStatus, *getStatus} {
		Equals(t, pull, s.Pull)
		Equals(t, []models.ProjectStatus{
			{
				RepoRelDir: "mergeme",
				Workspace:  "default",
				Status:     models.AppliedPlanStatus,
			},
			{
				RepoRelDir:  "projectname",
				Workspace:   "default",
				ProjectName: "projectname",
				Status:      models.ErroredApplyStatus,
			},
			{
				RepoRelDir: "staythesame",
				Workspace:  "default",
				Status:     models.PlannedPlanStatus,
			},
			{
				RepoRelDir: "newresult",
				Workspace:  "default",
				Status:     models.AppliedPlanStatus,
			},
		}, updateStatus.Projects)
	}
	b.Close()
}

// Test that if we update one existing policy status via approve_policies and our new status is for a
// the same commit, that we merge the statuses.
func TestPullStatus_UpdateMerge_ApprovePolicies(t *testing.T) {
	b := newTestDB2(t)

	pull := models.PullRequest{
		Num:        1,
		HeadCommit: "sha",
		URL:        "url",
		HeadBranch: "head",
		BaseBranch: "base",
		Author:     "lkysow",
		State:      models.OpenPullState,
		BaseRepo: models.Repo{
			FullName:          "runatlantis/atlantis",
			Owner:             "runatlantis",
			Name:              "atlantis",
			CloneURL:          "clone-url",
			SanitizedCloneURL: "clone-url",
			VCSHost: models.VCSHost{
				Hostname: "github.com",
				Type:     models.Github,
			},
		},
	}
	_, err := b.UpdatePullWithResults(
		pull,
		[]command.ProjectResult{
			{
				Command:    command.PolicyCheck,
				RepoRelDir: "mergeme",
				Workspace:  "default",
				ProjectCommandOutput: command.ProjectCommandOutput{
					Failure: "policy failure",
					PolicyCheckResults: &models.PolicyCheckResults{
						PolicySetResults: []models.PolicySetResult{
							{
								PolicySetName:    "policy1",
								ReqApprovalCount: 1,
							},
						},
					},
				},
			},
			{
				Command:     command.PolicyCheck,
				RepoRelDir:  "projectname",
				Workspace:   "default",
				ProjectName: "projectname",
				ProjectCommandOutput: command.ProjectCommandOutput{
					Failure: "policy failure",
					PolicyCheckResults: &models.PolicyCheckResults{
						PolicySetResults: []models.PolicySetResult{
							{
								PolicySetName:    "policy1",
								ReqApprovalCount: 1,
							},
						},
					},
				},
			},
		})
	Ok(t, err)

	updateStatus, err := b.UpdatePullWithResults(pull,
		[]command.ProjectResult{
			{
				Command:    command.ApprovePolicies,
				RepoRelDir: "mergeme",
				Workspace:  "default",
				ProjectCommandOutput: command.ProjectCommandOutput{
					PolicyCheckResults: &models.PolicyCheckResults{
						PolicySetResults: []models.PolicySetResult{
							{
								PolicySetName:    "policy1",
								ReqApprovalCount: 1,
								Approvals:        []models.PolicySetApproval{{Approver: "approver1"}},
							},
						},
					},
				},
			},
		})
	Ok(t, err)

	getStatus, err := b.GetPullStatus(pull)
	Ok(t, err)

	// Test both the pull state returned from the update call *and* the getCommandLock
	// call.
	for _, s := range []models.PullStatus{updateStatus, *getStatus} {
		Equals(t, pull, s.Pull)
		Equals(t, []models.ProjectStatus{
			{
				RepoRelDir: "mergeme",
				Workspace:  "default",
				Status:     models.PassedPolicyCheckStatus,
				PolicyStatus: []models.PolicySetStatus{
					{
						PolicySetName: "policy1",
						Approvals:     []models.PolicySetApproval{{Approver: "approver1"}},
					},
				},
			},
			{
				RepoRelDir:  "projectname",
				Workspace:   "default",
				ProjectName: "projectname",
				Status:      models.ErroredPolicyCheckStatus,
				PolicyStatus: []models.PolicySetStatus{
					{
						PolicySetName: "policy1",
						Approvals:     nil,
					},
				},
			},
		}, updateStatus.Projects)
	}
	b.Close()
}

// Test that policy approvals are preserved when HeadCommit changes,
// so sticky approvals can survive across code pushes.
func TestPullStatus_UpdateNewCommit_PreservesPolicyApprovals(t *testing.T) {
	b := newTestDB2(t)

	pull := models.PullRequest{
		Num:        1,
		HeadCommit: "sha-A",
		URL:        "url",
		HeadBranch: "head",
		BaseBranch: "base",
		Author:     "lkysow",
		State:      models.OpenPullState,
		BaseRepo: models.Repo{
			FullName:          "runatlantis/atlantis",
			Owner:             "runatlantis",
			Name:              "atlantis",
			CloneURL:          "clone-url",
			SanitizedCloneURL: "clone-url",
			VCSHost: models.VCSHost{
				Hostname: "github.com",
				Type:     models.Github,
			},
		},
	}

	// Write initial policy check results with an approval at commit A.
	_, err := b.UpdatePullWithResults(pull, []command.ProjectResult{
		{
			Command:    command.PolicyCheck,
			RepoRelDir: "mydir",
			Workspace:  "default",
			ProjectCommandOutput: command.ProjectCommandOutput{
				Failure: "policy failure",
				PolicyCheckResults: &models.PolicyCheckResults{
					PolicySetResults: []models.PolicySetResult{
						{
							PolicySetName:    "policy1",
							ReqApprovalCount: 1,
							Hashes:           []string{"h1", "h2"},
							Approvals: []models.PolicySetApproval{
								{Approver: "boss", Hashes: []string{"h1", "h2"}},
							},
						},
					},
				},
			},
		},
	})
	Ok(t, err)

	// Push new commit B with a plan result (no PolicyCheckResults).
	// This simulates what happens when autoplan writes the plan to DB
	// before the policy check runs.
	pull.HeadCommit = "sha-B"
	status, err := b.UpdatePullWithResults(pull, []command.ProjectResult{
		{
			Command:    command.Plan,
			RepoRelDir: "mydir",
			Workspace:  "default",
			ProjectCommandOutput: command.ProjectCommandOutput{
				PlanSuccess: &models.PlanSuccess{
					TerraformOutput: "plan output",
				},
			},
		},
	})
	Ok(t, err)

	// The policy approvals from commit A should be preserved.
	Equals(t, 1, len(status.Projects))
	Equals(t, "mydir", status.Projects[0].RepoRelDir)
	Assert(t, len(status.Projects[0].PolicyStatus) > 0, "expected policy status to be preserved across commit change")
	Equals(t, "policy1", status.Projects[0].PolicyStatus[0].PolicySetName)
	Equals(t, 1, len(status.Projects[0].PolicyStatus[0].Approvals))
	Equals(t, "boss", status.Projects[0].PolicyStatus[0].Approvals[0].Approver)

	// Verify via GetPullStatus too.
	getStatus, err := b.GetPullStatus(pull)
	Ok(t, err)
	Equals(t, 1, len(getStatus.Projects[0].PolicyStatus[0].Approvals))
	b.Close()
}

// TestPullStatus_UpdateOverwritesCorruptData verifies that
// UpdatePullWithResults tolerates a pre-existing pull-status blob whose JSON
// no longer matches the current Go shape (e.g. after upgrading across a
// PullStatus schema change). The corrupt entry should be logged and
// overwritten rather than causing every subsequent plan to fail.
func TestPullStatus_UpdateOverwritesCorruptData(t *testing.T) {
	tmp := t.TempDir()

	pull := models.PullRequest{
		Num:        1,
		HeadCommit: "sha-A",
		URL:        "url",
		HeadBranch: "head",
		BaseBranch: "base",
		Author:     "lkysow",
		State:      models.OpenPullState,
		BaseRepo: models.Repo{
			FullName:          "runatlantis/atlantis",
			Owner:             "runatlantis",
			Name:              "atlantis",
			CloneURL:          "clone-url",
			SanitizedCloneURL: "clone-url",
			VCSHost: models.VCSHost{
				Hostname: "github.com",
				Type:     models.Github,
			},
		},
	}

	// First, let boltdb.New create the file and buckets.
	b, err := boltdb.New(tmp)
	Ok(t, err)
	Ok(t, b.Close())

	// Inject a corrupt pull-status blob simulating a legacy on-disk shape
	// that the current Go types cannot unmarshal.
	key := fmt.Appendf(nil, "%s::%s::%d",
		pull.BaseRepo.VCSHost.Hostname, pull.BaseRepo.FullName, pull.Num)
	corrupt := []byte(`{"Projects":[{"Workspace":"default","RepoRelDir":"mydir","ProjectName":"","PolicyStatus":[{"PolicySetName":"policy1","Passed":false,"Approvals":2}],"Status":0}],"Pull":{"Num":1}}`)
	raw, err := bolt.Open(filepath.Join(tmp, "atlantis.db"), 0600, nil)
	Ok(t, err)
	err = raw.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("pulls")).Put(key, corrupt)
	})
	Ok(t, err)
	Ok(t, raw.Close())

	// Reopen and write fresh results. This must succeed despite the
	// unreadable prior entry.
	b, err = boltdb.New(tmp)
	Ok(t, err)
	defer b.Close()

	status, err := b.UpdatePullWithResults(pull, []command.ProjectResult{
		{
			Command:    command.Plan,
			RepoRelDir: "mydir",
			Workspace:  "default",
			ProjectCommandOutput: command.ProjectCommandOutput{
				PlanSuccess: &models.PlanSuccess{TerraformOutput: "plan output"},
			},
		},
	})
	Ok(t, err)
	Equals(t, 1, len(status.Projects))
	Equals(t, "mydir", status.Projects[0].RepoRelDir)
	// Prior in-flight approvals are lost; this is the documented trade-off.
	Equals(t, 0, len(status.Projects[0].PolicyStatus))

	// The corrupt entry is gone: reading it back returns clean data.
	got, err := b.GetPullStatus(pull)
	Ok(t, err)
	Assert(t, got != nil, "expected non-nil pull status")
	Equals(t, 1, len(got.Projects))
	Equals(t, models.PlannedPlanStatus, got.Projects[0].Status)
}

func TestBeginPlanGenerationOverwritesCorruptData(t *testing.T) {
	tmp := t.TempDir()
	pull := models.PullRequest{
		Num:        1,
		HeadCommit: "sha-A",
		BaseBranch: "main",
		BaseRepo: models.Repo{
			FullName: "runatlantis/atlantis",
			VCSHost:  models.VCSHost{Hostname: "github.com", Type: models.Github},
		},
	}

	database, err := boltdb.New(tmp)
	Ok(t, err)
	Ok(t, database.Close())

	key := fmt.Appendf(nil, "%s::%s::%d", pull.BaseRepo.VCSHost.Hostname, pull.BaseRepo.FullName, pull.Num)
	raw, err := bolt.Open(filepath.Join(tmp, "atlantis.db"), 0600, nil)
	Ok(t, err)
	err = raw.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("pulls")).Put(key, []byte(`{"Projects":[{"PolicyStatus":[{"Approvals":2}]}]}`))
	})
	Ok(t, err)
	Ok(t, raw.Close())

	database, err = boltdb.New(tmp)
	Ok(t, err)
	t.Cleanup(func() { database.Close() })
	selected := []models.ProjectStatus{{Workspace: "default", RepoRelDir: "project-a", ProjectName: "a"}}
	status, err := database.BeginPlanGeneration(pull, selected, "generation-1", "")
	Ok(t, err)
	Equals(t, pull, status.Pull)
	project := findPlanGenerationProject(t, status, "project-a", "a")
	Equals(t, models.ErroredPlanStatus, project.Status)
	Equals(t, "generation-1", project.PlanGeneration)

	status.PullStatus, err = database.CompletePlanGeneration(pull, "generation-1", []command.ProjectResult{{
		Command: command.Plan, Workspace: "default", RepoRelDir: "project-a", ProjectName: "a",
		ProjectCommandOutput: command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}},
	}}, "")

	Ok(t, err)
	project = findPlanGenerationProject(t, status, "project-a", "a")
	Equals(t, models.PlannedPlanStatus, project.Status)
	Equals(t, "", project.PlanGeneration)
}

func TestPlanPublicationClaimLifecycleAndTransitionFence(t *testing.T) {
	database, err := boltdb.New(t.TempDir())
	Ok(t, err)
	t.Cleanup(func() { database.Close() })
	pull := planGenerationTestPull()
	selected := []models.ProjectStatus{{Workspace: "default", RepoRelDir: "project-a", ProjectName: "a"}}

	Ok(t, database.AcquirePlanPublicationClaim(pull, "owner-a"))
	Ok(t, database.AcquirePlanPublicationClaim(pull, "owner-a"))
	err = database.AcquirePlanPublicationClaim(pull, "owner-b")
	assertPlanPublicationClaimError(t, err, coredb.ErrPlanPublicationBusy)
	_, err = database.BeginPlanGeneration(pull, selected, "generation-a", "")
	assertPlanPublicationClaimError(t, err, coredb.ErrPlanPublicationNotOwned)
	_, err = database.BeginPlanGeneration(pull, selected, "generation-a", "owner-b")
	assertPlanPublicationClaimError(t, err, coredb.ErrPlanPublicationNotOwned)
	_, err = database.BeginPlanGeneration(pull, selected, "generation-a", "owner-a", "owner-b")
	assertPlanPublicationClaimError(t, err, coredb.ErrPlanPublicationNotOwned)
	_, err = database.BeginPlanGenerationReplacing(pull, selected, "generation-a")
	assertPlanPublicationClaimError(t, err, coredb.ErrPlanPublicationNotOwned)
	_, err = database.BeginPlanGeneration(pull, selected, "generation-a", "owner-a")
	Ok(t, err)
	_, err = database.CompletePlanGeneration(pull, "generation-a", []command.ProjectResult{simplePlanGenerationResult("project-a", "a")}, "")
	assertPlanPublicationClaimError(t, err, coredb.ErrPlanPublicationNotOwned)
	_, err = database.CompletePlanGeneration(pull, "generation-a", []command.ProjectResult{simplePlanGenerationResult("project-a", "a")}, "owner-a")
	Ok(t, err)
	_, err = database.UpdatePolicyResultsForPlanGeneration(pull, []command.ProjectResult{policyGenerationResult("project-a", "a", "generation-a", "policy")}, "")
	assertPlanPublicationClaimError(t, err, coredb.ErrPlanPublicationNotOwned)
	_, err = database.UpdateApplyResultsForPlanGeneration(pull, []command.ProjectResult{applyGenerationResult("project-a", "a", "generation-a", nil)}, "")
	assertPlanPublicationClaimError(t, err, coredb.ErrPlanPublicationNotOwned)
	_, err = database.UpdateDiscardResultsForPlanGeneration(pull, []command.ProjectResult{discardGenerationResult(command.Import, "project-a", "a", "generation-a")}, "")
	assertPlanPublicationClaimError(t, err, coredb.ErrPlanPublicationNotOwned)
	err = database.ReleasePlanPublicationClaim(pull, "owner-b")
	assertPlanPublicationClaimError(t, err, coredb.ErrPlanPublicationNotOwned)
	Ok(t, database.ReleasePlanPublicationClaim(pull, "owner-a"))
	Ok(t, database.AcquirePlanPublicationClaim(pull, "owner-b"))
	Ok(t, database.ForceClearPlanPublicationClaim(pull))
	_, err = database.BeginPlanGeneration(pull, selected, "generation-b")
	Ok(t, err)
}

func TestPlanPublicationClaimSurvivesRestart(t *testing.T) {
	dataDir := t.TempDir()
	pull := planGenerationTestPull()
	database, err := boltdb.New(dataDir)
	Ok(t, err)
	Ok(t, database.AcquirePlanPublicationClaim(pull, "owner-a"))
	Ok(t, database.Close())

	database, err = boltdb.New(dataDir)
	Ok(t, err)
	t.Cleanup(func() { database.Close() })
	err = database.AcquirePlanPublicationClaim(pull, "owner-b")
	assertPlanPublicationClaimError(t, err, coredb.ErrPlanPublicationBusy)
	Ok(t, database.ReleasePlanPublicationClaim(pull, "owner-a"))
}

func TestDeletePullStatusClearsPlanPublicationClaim(t *testing.T) {
	database, err := boltdb.New(t.TempDir())
	Ok(t, err)
	t.Cleanup(func() { database.Close() })
	pull := planGenerationTestPull()
	Ok(t, database.AcquirePlanPublicationClaim(pull, "owner-a"))
	Ok(t, database.DeletePullStatus(pull))
	Ok(t, database.AcquirePlanPublicationClaim(pull, "owner-b"))
}

func TestPlanPublicationClaimBackendErrorIsOrdinary(t *testing.T) {
	database, err := boltdb.New(t.TempDir())
	Ok(t, err)
	Ok(t, database.Close())
	err = database.AcquirePlanPublicationClaim(planGenerationTestPull(), "owner-a")
	Assert(t, err != nil, "expected closed backend error")
	Assert(t, !errors.Is(err, coredb.ErrPlanPublicationBusy), "backend error must not be busy")
	Assert(t, !errors.Is(err, coredb.ErrPlanPublicationNotOwned), "backend error must not be ownership")
}

func TestBeginPlanGenerationReplacingDiscardsPriorProjects(t *testing.T) {
	database, err := boltdb.New(t.TempDir())
	Ok(t, err)
	t.Cleanup(func() { database.Close() })
	pull := planGenerationTestPull()
	projectA := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-a", ProjectName: "a"}
	projectB := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-b", ProjectName: "b"}
	policyStatus := []models.PolicySetStatus{{PolicySetName: "required-policy", Passed: true}}

	_, err = database.UpdatePullWithResults(pull, []command.ProjectResult{{
		Command: command.PolicyCheck, Workspace: "default", RepoRelDir: "project-a", ProjectName: "a",
		ProjectCommandOutput: command.ProjectCommandOutput{PolicyCheckResults: &models.PolicyCheckResults{
			PolicySetResults: []models.PolicySetResult{{PolicySetName: "required-policy", Passed: true}},
		}},
	}, {
		Command: command.PolicyCheck, Workspace: "default", RepoRelDir: "project-b", ProjectName: "b",
		ProjectCommandOutput: command.ProjectCommandOutput{PolicyCheckResults: &models.PolicyCheckResults{
			PolicySetResults: []models.PolicySetResult{{PolicySetName: "required-policy", Passed: true}},
		}},
	}})
	Ok(t, err)
	_, err = database.BeginPlanGeneration(pull, []models.ProjectStatus{projectA, projectB}, "generation-old", "")
	Ok(t, err)
	prior, err := database.CompletePlanGeneration(pull, "generation-old", []command.ProjectResult{
		managedPlanGenerationResult("project-a", "a", "hash-a"),
		managedPlanGenerationResult("project-b", "b", "hash-b"),
	}, "")

	Ok(t, err)
	Equals(t, models.PlannedPlanStatus, findPlanGenerationProject(t, prior, "project-a", "a").Status)
	Equals(t, "hash-a", findPlanGenerationProject(t, prior, "project-a", "a").ManagedPlanHash)
	Equals(t, "generation-old", findPlanGenerationProject(t, prior, "project-a", "a").AcceptedPlanGeneration)
	Equals(t, models.PlannedPlanStatus, findPlanGenerationProject(t, prior, "project-b", "b").Status)
	Equals(t, "hash-b", findPlanGenerationProject(t, prior, "project-b", "b").ManagedPlanHash)
	Equals(t, "generation-old", findPlanGenerationProject(t, prior, "project-b", "b").AcceptedPlanGeneration)

	status, err := database.BeginPlanGenerationReplacing(pull, []models.ProjectStatus{projectA}, "generation-new", "")
	Ok(t, err)
	selected := findPlanGenerationProject(t, status, "project-a", "a")
	Equals(t, models.ErroredPlanStatus, selected.Status)
	Equals(t, "generation-new", selected.PlanGeneration)
	Equals(t, "", selected.ManagedPlanHash)
	Equals(t, "", selected.AcceptedPlanGeneration)
	Equals(t, policyStatus, selected.PolicyStatus)
	unrelated := findPlanGenerationProject(t, status, "project-b", "b")
	Equals(t, models.DiscardedPlanStatus, unrelated.Status)
	Equals(t, "", unrelated.PlanGeneration)
	Equals(t, "", unrelated.ManagedPlanHash)
	Equals(t, "", unrelated.AcceptedPlanGeneration)
	Equals(t, policyStatus, unrelated.PolicyStatus)
}

func TestBeginPlanGenerationReplacingKeepsActiveProjectsErrored(t *testing.T) {
	database, err := boltdb.New(t.TempDir())
	Ok(t, err)
	t.Cleanup(func() { database.Close() })
	pull := planGenerationTestPull()
	projectX := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-x", ProjectName: "x"}
	projectY := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-y", ProjectName: "y"}
	_, err = database.BeginPlanGeneration(pull, []models.ProjectStatus{projectX}, "generation-active", "")
	Ok(t, err)

	status, err := database.BeginPlanGenerationReplacing(pull, []models.ProjectStatus{projectY}, "generation-new", "")
	Ok(t, err)
	Equals(t, []coredb.PlanGenerationProject{{Workspace: "default", RepoRelDir: "project-x", ProjectName: "x", CurrentGeneration: "generation-active", CurrentStatus: models.ErroredPlanStatus}}, status.Canceled)
	Equals(t, 2, len(status.Projects))
	cancelled := findPlanGenerationProject(t, status, "project-x", "x")
	Equals(t, models.ErroredPlanStatus, cancelled.Status)
	Equals(t, "", cancelled.PlanGeneration)
	Equals(t, "", cancelled.ManagedPlanHash)
	Equals(t, "", cancelled.AcceptedPlanGeneration)
	selected := findPlanGenerationProject(t, status, "project-y", "y")
	Equals(t, models.ErroredPlanStatus, selected.Status)
	Equals(t, "generation-new", selected.PlanGeneration)

	_, err = database.CompletePlanGeneration(pull, "generation-active", []command.ProjectResult{
		managedPlanGenerationResult("project-x", "x", "hash-stale"),
	}, "")

	assertPlanGenerationError(t, err, coredb.ErrPlanGenerationSuperseded)
	unchanged, err := database.GetPullStatus(pull)
	Ok(t, err)
	Equals(t, 2, len(unchanged.Projects))
	Equals(t, models.ErroredPlanStatus, findPlanGenerationProject(t, *unchanged, "project-x", "x").Status)
	Equals(t, "generation-new", findPlanGenerationProject(t, *unchanged, "project-y", "y").PlanGeneration)
}

func TestBeginPlanGenerationReplacingOverwritesCorruptData(t *testing.T) {
	dataDir := t.TempDir()
	pull := planGenerationTestPull()
	database, err := boltdb.New(dataDir)
	Ok(t, err)
	Ok(t, database.Close())
	key := fmt.Appendf(nil, "%s::%s::%d", pull.BaseRepo.VCSHost.Hostname, pull.BaseRepo.FullName, pull.Num)
	raw, err := bolt.Open(filepath.Join(dataDir, "atlantis.db"), 0600, nil)
	Ok(t, err)
	Ok(t, raw.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("pulls")).Put(key, []byte(`{"Projects":[{"PolicyStatus":[{"Approvals":2}]}]}`))
	}))
	Ok(t, raw.Close())
	database, err = boltdb.New(dataDir)
	Ok(t, err)
	t.Cleanup(func() { database.Close() })

	status, err := database.BeginPlanGenerationReplacing(pull, []models.ProjectStatus{{
		Workspace: "default", RepoRelDir: "project-a", ProjectName: "a",
	}}, "generation-1", "")

	Ok(t, err)
	Equals(t, 1, len(status.Projects))
	project := findPlanGenerationProject(t, status, "project-a", "a")
	Equals(t, models.ErroredPlanStatus, project.Status)
	Equals(t, "generation-1", project.PlanGeneration)
}

func TestBeginPlanGenerationReplacingReturnsBoltErrors(t *testing.T) {
	database, err := boltdb.New(t.TempDir())
	Ok(t, err)
	Ok(t, database.Close())
	_, err = database.BeginPlanGenerationReplacing(planGenerationTestPull(), []models.ProjectStatus{{
		Workspace: "default", RepoRelDir: "project-a", ProjectName: "a",
	}}, "generation-1", "")

	Assert(t, err != nil, "expected closed Bolt database error")
	var completionErr *coredb.PlanGenerationCompletionError
	Assert(t, !errors.As(err, &completionErr), "expected backend error to remain ordinary, got %v", err)
}

func TestPlanGenerationInvalidationPreservesUnrelatedProjectsAndRejectsStaleCompletion(t *testing.T) {
	database, err := boltdb.New(t.TempDir())
	Ok(t, err)
	t.Cleanup(func() { database.Close() })
	pull := models.PullRequest{
		Num:        1,
		HeadCommit: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		BaseBranch: "main",
		BaseRepo: models.Repo{
			FullName: "runatlantis/atlantis",
			VCSHost:  models.VCSHost{Hostname: "github.com", Type: models.Github},
		},
	}
	planResult := func(dir, projectName string) command.ProjectResult {
		return command.ProjectResult{
			Command:     command.Plan,
			Workspace:   "default",
			RepoRelDir:  dir,
			ProjectName: projectName,
			ProjectCommandOutput: command.ProjectCommandOutput{
				PlanSuccess: &models.PlanSuccess{},
			},
		}
	}
	_, err = database.UpdatePullWithResults(pull, []command.ProjectResult{
		managedPlanGenerationResult("project-a", "a", "hash-a"),
		planResult("project-b", "b"),
	})
	Ok(t, err)
	_, err = database.UpdatePullWithResults(pull, []command.ProjectResult{{
		Command:     command.PolicyCheck,
		Workspace:   "default",
		RepoRelDir:  "project-a",
		ProjectName: "a",
		ProjectCommandOutput: command.ProjectCommandOutput{PolicyCheckResults: &models.PolicyCheckResults{
			PolicySetResults: []models.PolicySetResult{{
				PolicySetName: "required-policy",
				Passed:        true,
				Approvals:     []models.PolicySetApproval{{Approver: "reviewer"}},
			}},
		}},
	}})
	Ok(t, err)
	selected := []models.ProjectStatus{{Workspace: "default", RepoRelDir: "project-a", ProjectName: "a"}}

	status, err := database.BeginPlanGeneration(pull, selected, "generation-1", "")
	Ok(t, err)
	projectA := findPlanGenerationProject(t, status, "project-a", "a")
	projectB := findPlanGenerationProject(t, status, "project-b", "b")
	Equals(t, models.ErroredPlanStatus, projectA.Status)
	Equals(t, "generation-1", projectA.PlanGeneration)
	Equals(t, 1, len(projectA.PolicyStatus))
	Equals(t, models.PlannedPlanStatus, projectB.Status)
	Equals(t, "", projectB.PlanGeneration)

	_, err = database.BeginPlanGeneration(pull, selected, "generation-2", "")
	Ok(t, err)
	_, err = database.CompletePlanGeneration(pull, "generation-1", []command.ProjectResult{planResult("project-a", "a")}, "")
	Assert(t, err != nil, "expected stale plan generation completion to fail")
	Assert(t, errors.Is(err, coredb.ErrPlanGenerationSuperseded), "expected superseded classification, got %v", err)
	var completionErr *coredb.PlanGenerationCompletionError
	Assert(t, errors.As(err, &completionErr), "expected typed completion error, got %T", err)

	status.PullStatus, err = database.CompletePlanGeneration(pull, "generation-2", []command.ProjectResult{planResult("project-a", "a")}, "")
	Ok(t, err)
	projectA = findPlanGenerationProject(t, status, "project-a", "a")
	projectB = findPlanGenerationProject(t, status, "project-b", "b")
	Equals(t, models.PlannedPlanStatus, projectA.Status)
	Equals(t, "", projectA.PlanGeneration)
	Equals(t, models.PlannedPlanStatus, projectB.Status)
}

func TestPlanGenerationPartialSupersessionCancelsWholePriorGeneration(t *testing.T) {
	database, err := boltdb.New(t.TempDir())
	Ok(t, err)
	t.Cleanup(func() { database.Close() })
	pull := planGenerationTestPull()
	projectA := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-a", ProjectName: "a"}
	projectB := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-b", ProjectName: "b"}

	_, err = database.BeginPlanGeneration(pull, []models.ProjectStatus{projectA, projectB}, "generation-a", "")
	Ok(t, err)
	status, err := database.BeginPlanGeneration(pull, []models.ProjectStatus{projectA}, "generation-b", "")
	Ok(t, err)
	Equals(t, []coredb.PlanGenerationProject{{Workspace: "default", RepoRelDir: "project-b", ProjectName: "b", CurrentGeneration: "generation-a", CurrentStatus: models.ErroredPlanStatus}}, status.Canceled)
	activeA := findPlanGenerationProject(t, status, "project-a", "a")
	cancelledB := findPlanGenerationProject(t, status, "project-b", "b")
	Equals(t, "generation-b", activeA.PlanGeneration)
	Equals(t, models.ErroredPlanStatus, cancelledB.Status)
	Equals(t, "", cancelledB.PlanGeneration)

	oldResults := []command.ProjectResult{
		managedPlanGenerationResult("project-a", "a", "hash-a"),
		managedPlanGenerationResult("project-b", "b", "hash-b"),
	}
	_, err = database.CompletePlanGeneration(pull, "generation-a", oldResults, "")
	Assert(t, errors.Is(err, coredb.ErrPlanGenerationSuperseded), "expected superseded classification, got %v", err)
	var completionErr *coredb.PlanGenerationCompletionError
	Assert(t, errors.As(err, &completionErr), "expected typed completion error, got %T", err)
	Equals(t, 2, len(completionErr.Projects))

	unchanged, err := database.GetPullStatus(pull)
	Ok(t, err)
	Equals(t, "generation-b", findPlanGenerationProject(t, *unchanged, "project-a", "a").PlanGeneration)
	Equals(t, "", findPlanGenerationProject(t, *unchanged, "project-b", "b").PlanGeneration)
	Equals(t, "", findPlanGenerationProject(t, *unchanged, "project-a", "a").ManagedPlanHash)
	Equals(t, "", findPlanGenerationProject(t, *unchanged, "project-b", "b").ManagedPlanHash)

	_, err = database.CompletePlanGeneration(pull, "generation-b", []command.ProjectResult{
		managedPlanGenerationResult("project-a", "a", "hash-new"),
	}, "")

	Ok(t, err)
	_, err = database.CompletePlanGeneration(pull, "generation-a", oldResults, "")
	Assert(t, errors.Is(err, coredb.ErrPlanGenerationSuperseded), "expected completed newer generation to supersede old completion, got %v", err)
}

func TestCompletePlanGenerationClassifiesStateErrors(t *testing.T) {
	t.Run("backend", func(t *testing.T) {
		database, err := boltdb.New(t.TempDir())
		Ok(t, err)
		pull := planGenerationTestPull()
		project := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-a", ProjectName: "a"}
		_, err = database.BeginPlanGeneration(pull, []models.ProjectStatus{project}, "generation-1", "")
		Ok(t, err)
		Ok(t, database.Close())
		_, err = database.CompletePlanGeneration(pull, "generation-1", []command.ProjectResult{managedPlanGenerationResult("project-a", "a", "hash")}, "")
		Assert(t, err != nil, "expected closed database completion to fail")
		var completionErr *coredb.PlanGenerationCompletionError
		Assert(t, !errors.As(err, &completionErr), "expected backend error to remain ordinary, got %v", err)
	})

	t.Run("missing", func(t *testing.T) {
		database, err := boltdb.New(t.TempDir())
		Ok(t, err)
		t.Cleanup(func() { database.Close() })
		_, err = database.CompletePlanGeneration(planGenerationTestPull(), "generation-1", nil, "")
		assertPlanGenerationError(t, err, coredb.ErrPlanGenerationStateInvalid)
	})

	t.Run("empty results without owner", func(t *testing.T) {
		database, err := boltdb.New(t.TempDir())
		Ok(t, err)
		t.Cleanup(func() { database.Close() })
		pull := planGenerationTestPull()
		_, err = database.UpdatePullWithResults(pull, []command.ProjectResult{{
			Command: command.Plan, Workspace: "default", RepoRelDir: "project-a", ProjectName: "a",
			ProjectCommandOutput: command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}},
		}})
		Ok(t, err)
		_, err = database.CompletePlanGeneration(pull, "generation-1", nil, "")
		assertPlanGenerationError(t, err, coredb.ErrPlanGenerationStateInvalid)
	})

	t.Run("corrupt", func(t *testing.T) {
		dataDir := t.TempDir()
		pull := planGenerationTestPull()
		database, err := boltdb.New(dataDir)
		Ok(t, err)
		Ok(t, database.Close())
		key := fmt.Appendf(nil, "%s::%s::%d", pull.BaseRepo.VCSHost.Hostname, pull.BaseRepo.FullName, pull.Num)
		raw, err := bolt.Open(filepath.Join(dataDir, "atlantis.db"), 0600, nil)
		Ok(t, err)
		Ok(t, raw.Update(func(tx *bolt.Tx) error {
			return tx.Bucket([]byte("pulls")).Put(key, []byte(`{"Projects":`))
		}))
		Ok(t, raw.Close())
		database, err = boltdb.New(dataDir)
		Ok(t, err)
		t.Cleanup(func() { database.Close() })
		_, err = database.CompletePlanGeneration(pull, "generation-1", nil, "")
		assertPlanGenerationError(t, err, coredb.ErrPlanGenerationStateInvalid)
	})

	t.Run("pull identity changed", func(t *testing.T) {
		database, err := boltdb.New(t.TempDir())
		Ok(t, err)
		t.Cleanup(func() { database.Close() })
		pull := planGenerationTestPull()
		project := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-a", ProjectName: "a"}
		_, err = database.BeginPlanGeneration(pull, []models.ProjectStatus{project}, "generation-1", "")
		Ok(t, err)
		changedPull := pull
		changedPull.HeadCommit = "new-head"
		_, err = database.CompletePlanGeneration(changedPull, "generation-1", []command.ProjectResult{managedPlanGenerationResult("project-a", "a", "hash")}, "")
		assertPlanGenerationError(t, err, coredb.ErrPlanGenerationPullChanged)
	})

	t.Run("incomplete", func(t *testing.T) {
		database, err := boltdb.New(t.TempDir())
		Ok(t, err)
		t.Cleanup(func() { database.Close() })
		pull := planGenerationTestPull()
		projects := []models.ProjectStatus{
			{Workspace: "default", RepoRelDir: "project-a", ProjectName: "a"},
			{Workspace: "default", RepoRelDir: "project-b", ProjectName: "b"},
		}
		_, err = database.BeginPlanGeneration(pull, projects, "generation-1", "")
		Ok(t, err)
		_, err = database.CompletePlanGeneration(pull, "generation-1", []command.ProjectResult{managedPlanGenerationResult("project-a", "a", "hash")}, "")
		assertPlanGenerationError(t, err, coredb.ErrPlanGenerationIncomplete)
	})

	t.Run("managed hash missing", func(t *testing.T) {
		database, err := boltdb.New(t.TempDir())
		Ok(t, err)
		t.Cleanup(func() { database.Close() })
		pull := planGenerationTestPull()
		project := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-a", ProjectName: "a"}
		_, err = database.BeginPlanGeneration(pull, []models.ProjectStatus{project}, "generation-1", "")
		Ok(t, err)
		result := managedPlanGenerationResult("project-a", "a", "")
		_, err = database.CompletePlanGeneration(pull, "generation-1", []command.ProjectResult{result}, "")
		assertPlanGenerationError(t, err, coredb.ErrPlanGenerationStateInvalid)
	})
}

func TestUpdatePullWithResultsRejectsActivePlanGeneration(t *testing.T) {
	database, err := boltdb.New(t.TempDir())
	Ok(t, err)
	t.Cleanup(func() { database.Close() })
	pull := models.PullRequest{
		Num:        1,
		HeadCommit: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		BaseBranch: "main",
		BaseRepo: models.Repo{
			FullName: "runatlantis/atlantis",
			VCSHost:  models.VCSHost{Hostname: "github.com", Type: models.Github},
		},
	}
	planResult := func(dir, projectName string) command.ProjectResult {
		return command.ProjectResult{
			Command:     command.Plan,
			Workspace:   "default",
			RepoRelDir:  dir,
			ProjectName: projectName,
			ProjectCommandOutput: command.ProjectCommandOutput{
				PlanSuccess: &models.PlanSuccess{},
			},
		}
	}
	policyResult := func(commandName command.Name) command.ProjectResult {
		return command.ProjectResult{
			Command:     commandName,
			Workspace:   "default",
			RepoRelDir:  "project-a",
			ProjectName: "a",
			ProjectCommandOutput: command.ProjectCommandOutput{
				PolicyCheckResults: &models.PolicyCheckResults{PolicySetResults: []models.PolicySetResult{{
					PolicySetName: "required-policy",
					Passed:        true,
					Approvals:     []models.PolicySetApproval{{Approver: "reviewer"}},
				}}},
			},
		}
	}
	_, err = database.UpdatePullWithResults(pull, []command.ProjectResult{
		managedPlanGenerationResult("project-a", "a", "hash-a"),
		planResult("project-b", "b"),
	})
	Ok(t, err)
	_, err = database.UpdatePullWithResults(pull, []command.ProjectResult{policyResult(command.PolicyCheck)})
	Ok(t, err)

	status, err := database.BeginPlanGeneration(pull, []models.ProjectStatus{{
		Workspace: "default", RepoRelDir: "project-a", ProjectName: "a",
	}}, "generation-1", "")

	Ok(t, err)
	active := findPlanGenerationProject(t, status, "project-a", "a")
	Equals(t, models.ErroredPlanStatus, active.Status)
	Equals(t, "generation-1", active.PlanGeneration)
	Equals(t, 1, len(active.PolicyStatus))

	ordinaryWrites := []struct {
		name   string
		result command.ProjectResult
	}{
		{name: "approve policies", result: policyResult(command.ApprovePolicies)},
		{name: "policy check", result: policyResult(command.PolicyCheck)},
		{name: "apply", result: command.ProjectResult{
			Command: command.Apply, Workspace: "default", RepoRelDir: "project-a", ProjectName: "a",
			ProjectCommandOutput: command.ProjectCommandOutput{ApplySuccess: "applied"},
		}},
		{name: "import", result: command.ProjectResult{
			Command: command.Import, Workspace: "default", RepoRelDir: "project-a", ProjectName: "a",
			ProjectCommandOutput: command.ProjectCommandOutput{ImportSuccess: &models.ImportSuccess{}},
		}},
		{name: "state", result: command.ProjectResult{
			Command: command.State, Workspace: "default", RepoRelDir: "project-a", ProjectName: "a",
			ProjectCommandOutput: command.ProjectCommandOutput{StateRmSuccess: &models.StateRmSuccess{}},
		}},
	}
	for _, test := range ordinaryWrites {
		t.Run(test.name, func(t *testing.T) {
			_, err := database.UpdatePullWithResults(pull, []command.ProjectResult{test.result})
			Assert(t, err != nil, "expected ordinary write to reject an active plan generation")
			Assert(t, strings.Contains(err.Error(), "project has an active plan generation"), "unexpected error: %s", err)

			got, err := database.GetPullStatus(pull)
			Ok(t, err)
			active := findPlanGenerationProject(t, *got, "project-a", "a")
			Equals(t, models.ErroredPlanStatus, active.Status)
			Equals(t, "generation-1", active.PlanGeneration)
			Equals(t, 1, len(active.PolicyStatus))
		})
	}

	_, err = database.UpdatePullWithResults(pull, []command.ProjectResult{
		{
			Command: command.Apply, Workspace: "default", RepoRelDir: "project-b", ProjectName: "b",
			ProjectCommandOutput: command.ProjectCommandOutput{ApplySuccess: "applied"},
		},
		policyResult(command.ApprovePolicies),
	})
	Assert(t, err != nil, "expected the whole update to reject an active plan generation")
	got, err := database.GetPullStatus(pull)
	Ok(t, err)
	Equals(t, models.PlannedPlanStatus, findPlanGenerationProject(t, *got, "project-b", "b").Status)
	stalePull := pull
	stalePull.HeadCommit = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	_, err = database.UpdatePullWithResults(stalePull, []command.ProjectResult{{
		Command: command.Apply, Workspace: "default", RepoRelDir: "project-b", ProjectName: "b",
		ProjectCommandOutput: command.ProjectCommandOutput{ApplySuccess: "applied"},
	}})
	Assert(t, err != nil, "expected a different-head result to preserve every active plan generation")
	Assert(t, strings.Contains(err.Error(), "project has an active plan generation"), "unexpected error: %s", err)
	got, err = database.GetPullStatus(pull)
	Ok(t, err)
	active = findPlanGenerationProject(t, *got, "project-a", "a")
	Equals(t, models.ErroredPlanStatus, active.Status)
	Equals(t, "generation-1", active.PlanGeneration)
	Equals(t, models.PlannedPlanStatus, findPlanGenerationProject(t, *got, "project-b", "b").Status)
	err = database.UpdateProjectStatus(pull, "default", "project-a", models.DiscardedPlanStatus)
	Assert(t, err != nil, "expected lock deletion status update to reject an active plan generation")
	Assert(t, strings.Contains(err.Error(), "project has an active plan generation"), "unexpected error: %s", err)
	_, err = database.ReplacePullWithResults(pull, []command.ProjectResult{planResult("project-b", "b")})
	Assert(t, err != nil, "expected atomic replacement to reject an active plan generation")
	Assert(t, strings.Contains(err.Error(), "project has an active plan generation"), "unexpected error: %s", err)
	got, err = database.GetPullStatus(pull)
	Ok(t, err)
	active = findPlanGenerationProject(t, *got, "project-a", "a")
	Equals(t, models.ErroredPlanStatus, active.Status)
	Equals(t, "generation-1", active.PlanGeneration)
	Equals(t, models.PlannedPlanStatus, findPlanGenerationProject(t, *got, "project-b", "b").Status)

	status.PullStatus, err = database.CompletePlanGeneration(pull, "generation-1", []command.ProjectResult{planResult("project-a", "a")}, "")
	Ok(t, err)
	completed := findPlanGenerationProject(t, status, "project-a", "a")
	Equals(t, models.PlannedPlanStatus, completed.Status)
	Equals(t, 1, len(completed.PolicyStatus))
	_, err = database.CompletePlanGeneration(pull, "stale-generation", []command.ProjectResult{planResult("project-a", "a")}, "")
	Assert(t, err != nil, "expected stale plan generation completion to fail")
}

func TestPlanGenerationPreservesStickyPolicyStatusAcrossRestartAndPullHeadChange(t *testing.T) {
	dataDir := t.TempDir()
	database, err := boltdb.New(dataDir)
	Ok(t, err)
	pull := models.PullRequest{
		Num:        1,
		HeadCommit: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		BaseBranch: "main",
		BaseRepo: models.Repo{
			FullName: "runatlantis/atlantis",
			VCSHost:  models.VCSHost{Hostname: "github.com", Type: models.Github},
		},
	}
	planResult := func(dir, projectName string) command.ProjectResult {
		return command.ProjectResult{
			Command:     command.Plan,
			Workspace:   "default",
			RepoRelDir:  dir,
			ProjectName: projectName,
			ProjectCommandOutput: command.ProjectCommandOutput{
				PlanSuccess: &models.PlanSuccess{},
			},
		}
	}
	policyResult := func(dir, projectName, hash string, approvals []models.PolicySetApproval) command.ProjectResult {
		return command.ProjectResult{
			Command:     command.PolicyCheck,
			Workspace:   "default",
			RepoRelDir:  dir,
			ProjectName: projectName,
			ProjectCommandOutput: command.ProjectCommandOutput{
				PolicyCheckResults: &models.PolicyCheckResults{PolicySetResults: []models.PolicySetResult{{
					PolicySetName: "required-policy",
					Passed:        true,
					Hashes:        []string{hash},
					Approvals:     approvals,
				}}},
			},
		}
	}
	approval := models.PolicySetApproval{Approver: "reviewer", Hashes: []string{"hash-1"}}
	_, err = database.UpdatePullWithResults(pull, []command.ProjectResult{
		planResult("project-a", "a"),
		planResult("project-b", "b"),
	})
	Ok(t, err)
	_, err = database.UpdatePullWithResults(pull, []command.ProjectResult{
		policyResult("project-a", "a", "hash-1", []models.PolicySetApproval{approval}),
		policyResult("project-b", "b", "hash-b", []models.PolicySetApproval{{Approver: "other-reviewer", Hashes: []string{"hash-b"}}}),
	})
	Ok(t, err)

	selected := []models.ProjectStatus{{Workspace: "default", RepoRelDir: "project-a", ProjectName: "a"}}
	status, err := database.BeginPlanGeneration(pull, selected, "generation-1", "")
	Ok(t, err)
	projectA := findPlanGenerationProject(t, status, "project-a", "a")
	projectB := findPlanGenerationProject(t, status, "project-b", "b")
	Equals(t, models.ErroredPlanStatus, projectA.Status)
	Equals(t, "generation-1", projectA.PlanGeneration)
	Equals(t, "reviewer", projectA.PolicyStatus[0].Approvals[0].Approver)
	Equals(t, "other-reviewer", projectB.PolicyStatus[0].Approvals[0].Approver)

	failedPlan := planResult("project-a", "a")
	failedPlan.PlanSuccess = nil
	failedPlan.Error = errors.New("plan failed")
	status.PullStatus, err = database.CompletePlanGeneration(pull, "generation-1", []command.ProjectResult{failedPlan}, "")
	Ok(t, err)
	projectA = findPlanGenerationProject(t, status, "project-a", "a")
	Equals(t, models.ErroredPlanStatus, projectA.Status)
	Equals(t, "", projectA.PlanGeneration)
	Equals(t, "reviewer", projectA.PolicyStatus[0].Approvals[0].Approver)

	_, err = database.BeginPlanGeneration(pull, selected, "generation-2", "")
	Ok(t, err)
	_, err = database.CompletePlanGeneration(pull, "stale-generation", []command.ProjectResult{planResult("project-a", "a")}, "")
	Assert(t, err != nil, "expected failed generation completion")
	Ok(t, database.Close())

	database, err = boltdb.New(dataDir)
	Ok(t, err)
	t.Cleanup(func() { database.Close() })
	restartedStatus, err := database.GetPullStatus(pull)
	Ok(t, err)
	projectA = findPlanGenerationProject(t, *restartedStatus, "project-a", "a")
	Equals(t, models.ErroredPlanStatus, projectA.Status)
	Equals(t, "generation-2", projectA.PlanGeneration)
	Equals(t, "reviewer", projectA.PolicyStatus[0].Approvals[0].Approver)
	projectB = findPlanGenerationProject(t, *restartedStatus, "project-b", "b")
	Equals(t, "other-reviewer", projectB.PolicyStatus[0].Approvals[0].Approver)

	newHeadPull := pull
	newHeadPull.HeadCommit = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	status, err = database.BeginPlanGeneration(newHeadPull, selected, "generation-3", "")
	Ok(t, err)
	projectA = findPlanGenerationProject(t, status, "project-a", "a")
	Equals(t, models.ErroredPlanStatus, projectA.Status)
	Equals(t, "generation-3", projectA.PlanGeneration)
	Equals(t, "reviewer", projectA.PolicyStatus[0].Approvals[0].Approver)
	status.PullStatus, err = database.CompletePlanGeneration(newHeadPull, "generation-3", []command.ProjectResult{planResult("project-a", "a")}, "")
	Ok(t, err)
	projectA = findPlanGenerationProject(t, status, "project-a", "a")
	Equals(t, models.PlannedPlanStatus, projectA.Status)
	Equals(t, "reviewer", projectA.PolicyStatus[0].Approvals[0].Approver)
	Ok(t, database.Close())
	database, err = boltdb.New(dataDir)
	Ok(t, err)
	postPlanRestartStatus, err := database.GetPullStatus(newHeadPull)
	Ok(t, err)
	projectA = findPlanGenerationProject(t, *postPlanRestartStatus, "project-a", "a")
	Equals(t, models.PlannedPlanStatus, projectA.Status)
	Equals(t, "reviewer", projectA.PolicyStatus[0].Approvals[0].Approver)

	status.PullStatus, err = database.UpdatePullWithResults(newHeadPull, []command.ProjectResult{
		policyResult("project-a", "a", "hash-2", nil),
	})
	Ok(t, err)
	projectA = findPlanGenerationProject(t, status, "project-a", "a")
	Equals(t, "hash-2", projectA.PolicyStatus[0].Hashes[0])
	Equals(t, 0, len(projectA.PolicyStatus[0].Approvals))
}

func TestCompletePlanGenerationRejectsMatchingTokenWithNonActiveStatus(t *testing.T) {
	dataDir := t.TempDir()
	database, err := boltdb.New(dataDir)
	Ok(t, err)
	pull := models.PullRequest{
		Num:        1,
		HeadCommit: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		BaseBranch: "main",
		BaseRepo: models.Repo{
			FullName: "runatlantis/atlantis",
			VCSHost:  models.VCSHost{Hostname: "github.com", Type: models.Github},
		},
	}
	_, err = database.BeginPlanGeneration(pull, []models.ProjectStatus{{
		Workspace: "default", RepoRelDir: "project-a", ProjectName: "a",
	}}, "generation-1", "")

	Ok(t, err)
	Ok(t, database.Close())

	key := fmt.Appendf(nil, "%s::%s::%d", pull.BaseRepo.VCSHost.Hostname, pull.BaseRepo.FullName, pull.Num)
	raw, err := bolt.Open(filepath.Join(dataDir, "atlantis.db"), 0600, nil)
	Ok(t, err)
	err = raw.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("pulls"))
		var status models.PullStatus
		if err := json.Unmarshal(bucket.Get(key), &status); err != nil {
			return err
		}
		status.Projects[0].Status = models.PlannedPlanStatus
		serialized, err := json.Marshal(status)
		if err != nil {
			return err
		}
		return bucket.Put(key, serialized)
	})
	Ok(t, err)
	Ok(t, raw.Close())

	database, err = boltdb.New(dataDir)
	Ok(t, err)
	t.Cleanup(func() { database.Close() })
	_, err = database.CompletePlanGeneration(pull, "generation-1", []command.ProjectResult{{
		Command: command.Plan, Workspace: "default", RepoRelDir: "project-a", ProjectName: "a",
		ProjectCommandOutput: command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}},
	}}, "")

	assertPlanGenerationError(t, err, coredb.ErrPlanGenerationStateInvalid)
	status, err := database.GetPullStatus(pull)
	Ok(t, err)
	Equals(t, models.PlannedPlanStatus, status.Projects[0].Status)
	Equals(t, "generation-1", status.Projects[0].PlanGeneration)
}

func TestUpdatePolicyResultsForPlanGenerationIsGenerationFenced(t *testing.T) {
	database, err := boltdb.New(t.TempDir())
	Ok(t, err)
	t.Cleanup(func() { database.Close() })
	pull := planGenerationTestPull()
	projectA := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-a", ProjectName: "a"}
	projectB := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-b", ProjectName: "b"}
	_, err = database.BeginPlanGeneration(pull, []models.ProjectStatus{projectA}, "generation-a", "")
	Ok(t, err)
	_, err = database.CompletePlanGeneration(pull, "generation-a", []command.ProjectResult{
		managedPlanGenerationResult("project-a", "a", "hash-a"),
	}, "")

	Ok(t, err)
	_, err = database.BeginPlanGeneration(pull, []models.ProjectStatus{projectB}, "generation-b", "")
	Ok(t, err)
	_, err = database.CompletePlanGeneration(pull, "generation-b", []command.ProjectResult{
		managedPlanGenerationResult("project-b", "b", "hash-b"),
	}, "")

	Ok(t, err)

	_, err = database.UpdatePolicyResultsForPlanGeneration(pull, []command.ProjectResult{
		policyGenerationResult("project-a", "a", "generation-a", "policy-a"),
		policyGenerationResult("project-b", "b", "generation-a", "policy-b"),
	}, "")

	assertPlanGenerationError(t, err, coredb.ErrPlanGenerationSuperseded)
	unchanged, err := database.GetPullStatus(pull)
	Ok(t, err)
	Equals(t, 0, len(findPlanGenerationProject(t, *unchanged, "project-a", "a").PolicyStatus))
	Equals(t, 0, len(findPlanGenerationProject(t, *unchanged, "project-b", "b").PolicyStatus))

	status, err := database.UpdatePolicyResultsForPlanGeneration(pull, []command.ProjectResult{
		policyGenerationResult("project-a", "a", "generation-a", "policy-a"),
		policyGenerationResult("project-b", "b", "generation-b", "policy-b"),
	}, "")

	Ok(t, err)
	project := findPlanGenerationProject(t, status, "project-a", "a")
	Equals(t, models.PassedPolicyCheckStatus, project.Status)
	Equals(t, "hash-a", project.ManagedPlanHash)
	Equals(t, "generation-a", project.AcceptedPlanGeneration)
	Equals(t, "policy-a", project.PolicyStatus[0].PolicySetName)
	project = findPlanGenerationProject(t, status, "project-b", "b")
	Equals(t, "hash-b", project.ManagedPlanHash)
	Equals(t, "generation-b", project.AcceptedPlanGeneration)
	Equals(t, "policy-b", project.PolicyStatus[0].PolicySetName)

	_, err = database.UpdatePullWithResults(pull, []command.ProjectResult{{
		Command: command.Plan, Workspace: "default", RepoRelDir: "legacy-policy", ProjectName: "legacy-policy",
		ProjectCommandOutput: command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}},
	}})
	Ok(t, err)
	status, err = database.UpdatePolicyResultsForPlanGeneration(pull, []command.ProjectResult{
		policyGenerationResult("legacy-policy", "legacy-policy", "", "legacy-policy"),
	}, "")

	Ok(t, err)
	Equals(t, models.PassedPolicyCheckStatus, findPlanGenerationProject(t, status, "legacy-policy", "legacy-policy").Status)

	_, err = database.BeginPlanGeneration(pull, []models.ProjectStatus{projectA}, "generation-new", "")
	Ok(t, err)
	_, err = database.UpdatePolicyResultsForPlanGeneration(pull, []command.ProjectResult{
		policyGenerationResult("project-a", "a", "generation-a", "stale-policy"),
	}, "")

	assertPlanGenerationError(t, err, coredb.ErrPlanGenerationSuperseded)
}

func TestUpdatePolicyResultsForPlanGenerationReturnsBoltErrors(t *testing.T) {
	database, err := boltdb.New(t.TempDir())
	Ok(t, err)
	Ok(t, database.Close())
	_, err = database.UpdatePolicyResultsForPlanGeneration(planGenerationTestPull(), []command.ProjectResult{
		policyGenerationResult("project-a", "a", "generation-1", "policy-a"),
	}, "")

	Assert(t, err != nil, "expected closed Bolt database error")
	var completionErr *coredb.PlanGenerationCompletionError
	Assert(t, !errors.As(err, &completionErr), "expected backend error to remain ordinary, got %v", err)
}

func TestUpdateApplyResultsForPlanGenerationIsGenerationFenced(t *testing.T) {
	database, err := boltdb.New(t.TempDir())
	Ok(t, err)
	t.Cleanup(func() { database.Close() })
	pull := planGenerationTestPull()
	projectA := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-a", ProjectName: "a"}
	projectB := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-b", ProjectName: "b"}
	_, err = database.BeginPlanGeneration(pull, []models.ProjectStatus{projectA}, "generation-a", "")
	Ok(t, err)
	_, err = database.CompletePlanGeneration(pull, "generation-a", []command.ProjectResult{
		managedPlanGenerationResult("project-a", "a", "hash-a"),
	}, "")

	Ok(t, err)
	_, err = database.BeginPlanGeneration(pull, []models.ProjectStatus{projectB}, "generation-b", "")
	Ok(t, err)
	_, err = database.CompletePlanGeneration(pull, "generation-b", []command.ProjectResult{
		managedPlanGenerationResult("project-b", "b", "hash-b"),
	}, "")

	Ok(t, err)

	status, err := database.UpdateApplyResultsForPlanGeneration(pull, []command.ProjectResult{
		applyGenerationResult("project-a", "a", "generation-a", errors.New("apply failed")),
		applyGenerationResult("project-b", "b", "generation-b", nil),
	}, "")

	Ok(t, err)
	failed := findPlanGenerationProject(t, status, "project-a", "a")
	Equals(t, models.ErroredApplyStatus, failed.Status)
	Equals(t, "hash-a", failed.ManagedPlanHash)
	Equals(t, "generation-a", failed.AcceptedPlanGeneration)
	applied := findPlanGenerationProject(t, status, "project-b", "b")
	Equals(t, models.AppliedPlanStatus, applied.Status)
	Equals(t, "", applied.ManagedPlanHash)
	Equals(t, "", applied.AcceptedPlanGeneration)

	_, err = database.BeginPlanGeneration(pull, []models.ProjectStatus{projectA}, "generation-new", "")
	Ok(t, err)
	_, err = database.CompletePlanGeneration(pull, "generation-new", []command.ProjectResult{
		managedPlanGenerationResult("project-a", "a", "hash-new"),
	}, "")

	Ok(t, err)
	_, err = database.UpdateApplyResultsForPlanGeneration(pull, []command.ProjectResult{
		applyGenerationResult("project-a", "a", "generation-a", nil),
	}, "")

	assertPlanGenerationError(t, err, coredb.ErrPlanGenerationSuperseded)
	unchanged, err := database.GetPullStatus(pull)
	Ok(t, err)
	current := findPlanGenerationProject(t, *unchanged, "project-a", "a")
	Equals(t, models.PlannedPlanStatus, current.Status)
	Equals(t, "hash-new", current.ManagedPlanHash)
	Equals(t, "generation-new", current.AcceptedPlanGeneration)

	policyFailure := policyGenerationResult("project-a", "a", "generation-new", "policy")
	policyFailure.Error = errors.New("policy denied")
	policyFailure.PolicyCheckResults = nil
	status, err = database.UpdatePolicyResultsForPlanGeneration(pull, []command.ProjectResult{policyFailure}, "")
	Ok(t, err)
	Equals(t, models.ErroredPolicyCheckStatus, findPlanGenerationProject(t, status, "project-a", "a").Status)
	_, err = database.UpdateApplyResultsForPlanGeneration(pull, []command.ProjectResult{
		applyGenerationResult("project-a", "a", "generation-new", nil),
	}, "")

	Assert(t, errors.Is(err, coredb.ErrPlanGenerationSuperseded), "apply must not overwrite a same-generation policy blocker")

	_, err = database.UpdatePullWithResults(pull, []command.ProjectResult{{
		Command: command.Plan, Workspace: "default", RepoRelDir: "legacy", ProjectName: "legacy",
		ProjectCommandOutput: command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}},
	}})
	Ok(t, err)
	status, err = database.UpdateApplyResultsForPlanGeneration(pull, []command.ProjectResult{
		applyGenerationResult("legacy", "legacy", "", nil),
	}, "")

	Ok(t, err)
	Equals(t, models.AppliedPlanStatus, findPlanGenerationProject(t, status, "legacy", "legacy").Status)

	_, err = database.UpdatePullWithResults(pull, []command.ProjectResult{{
		Command: command.Plan, Workspace: "default", RepoRelDir: "no-changes", ProjectName: "no-changes",
		ProjectCommandOutput: command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{TerraformOutput: "No changes. Infrastructure is up-to-date."}},
	}})
	Ok(t, err)
	status, err = database.UpdateApplyResultsForPlanGeneration(pull, []command.ProjectResult{
		applyGenerationResult("no-changes", "no-changes", "", nil),
	}, "")

	Ok(t, err)
	Equals(t, models.AppliedPlanStatus, findPlanGenerationProject(t, status, "no-changes", "no-changes").Status)

	_, err = database.BeginPlanGeneration(pull, []models.ProjectStatus{{
		Workspace: "default", RepoRelDir: "legacy", ProjectName: "legacy",
	}}, "generation-active", "")

	Ok(t, err)
	_, err = database.UpdateApplyResultsForPlanGeneration(pull, []command.ProjectResult{
		applyGenerationResult("legacy", "legacy", "", nil),
	}, "")

	Assert(t, errors.Is(err, coredb.ErrPlanGenerationSuperseded), "legacy empty generation must not overwrite an active generation")
}

func TestUpdateApplyResultsForPlanGenerationReturnsBoltErrors(t *testing.T) {
	database, err := boltdb.New(t.TempDir())
	Ok(t, err)
	Ok(t, database.Close())
	_, err = database.UpdateApplyResultsForPlanGeneration(planGenerationTestPull(), []command.ProjectResult{
		applyGenerationResult("project-a", "a", "generation-a", nil),
	}, "")

	Assert(t, err != nil, "expected closed Bolt database error")
	var completionErr *coredb.PlanGenerationCompletionError
	Assert(t, !errors.As(err, &completionErr), "expected backend error to remain ordinary, got %v", err)
}

func TestUpdateDiscardResultsForPlanGenerationIsGenerationFenced(t *testing.T) {
	database, err := boltdb.New(t.TempDir())
	Ok(t, err)
	t.Cleanup(func() { database.Close() })
	pull := planGenerationTestPull()
	projectA := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-a", ProjectName: "a"}
	projectB := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-b", ProjectName: "b"}
	_, err = database.BeginPlanGeneration(pull, []models.ProjectStatus{projectA, projectB}, "generation-a")
	Ok(t, err)
	_, err = database.CompletePlanGeneration(pull, "generation-a", []command.ProjectResult{
		managedPlanGenerationResult("project-a", "a", "hash-a"),
		managedPlanGenerationResult("project-b", "b", "hash-b"),
	})
	Ok(t, err)
	_, err = database.BeginPlanGeneration(pull, []models.ProjectStatus{projectA}, "generation-b")
	Ok(t, err)
	_, err = database.CompletePlanGeneration(pull, "generation-b", []command.ProjectResult{
		managedPlanGenerationResult("project-a", "a", "hash-new"),
	})
	Ok(t, err)

	_, err = database.UpdateDiscardResultsForPlanGeneration(pull, []command.ProjectResult{
		discardGenerationResult(command.State, "project-b", "b", "generation-a"),
		discardGenerationResult(command.Import, "project-a", "a", "generation-a"),
	})
	assertPlanGenerationError(t, err, coredb.ErrPlanGenerationSuperseded)
	unchanged, err := database.GetPullStatus(pull)
	Ok(t, err)
	Equals(t, models.PlannedPlanStatus, findPlanGenerationProject(t, *unchanged, "project-a", "a").Status)
	Equals(t, "hash-new", findPlanGenerationProject(t, *unchanged, "project-a", "a").ManagedPlanHash)
	Equals(t, "generation-b", findPlanGenerationProject(t, *unchanged, "project-a", "a").AcceptedPlanGeneration)
	Equals(t, models.PlannedPlanStatus, findPlanGenerationProject(t, *unchanged, "project-b", "b").Status)
	Equals(t, "hash-b", findPlanGenerationProject(t, *unchanged, "project-b", "b").ManagedPlanHash)
	Equals(t, "generation-a", findPlanGenerationProject(t, *unchanged, "project-b", "b").AcceptedPlanGeneration)

	status, err := database.UpdateDiscardResultsForPlanGeneration(pull, []command.ProjectResult{
		discardGenerationResult(command.Import, "project-a", "a", "generation-b"),
		discardGenerationResult(command.State, "project-b", "b", "generation-a"),
	})
	Ok(t, err)
	for _, project := range status.Projects {
		Equals(t, models.DiscardedPlanStatus, project.Status)
		Equals(t, "", project.ManagedPlanHash)
		Equals(t, "", project.AcceptedPlanGeneration)
	}
}

func TestUpdateDiscardResultsForPlanGenerationLegacyAndFailure(t *testing.T) {
	database, err := boltdb.New(t.TempDir())
	Ok(t, err)
	t.Cleanup(func() { database.Close() })
	pull := planGenerationTestPull()
	_, err = database.UpdatePullWithResults(pull, []command.ProjectResult{
		managedPlanGenerationResult("project-a", "a", "legacy-hash"),
	})
	Ok(t, err)
	failure := discardGenerationResult(command.Import, "project-a", "a", "")
	failure.Error = errors.New("import failed")
	_, err = database.UpdateDiscardResultsForPlanGeneration(pull, []command.ProjectResult{failure})
	assertPlanGenerationError(t, err, coredb.ErrPlanGenerationStateInvalid)
	unchanged, err := database.GetPullStatus(pull)
	Ok(t, err)
	Equals(t, models.PlannedPlanStatus, unchanged.Projects[0].Status)
	Equals(t, "legacy-hash", unchanged.Projects[0].ManagedPlanHash)
	_, err = database.UpdateDiscardResultsForPlanGeneration(pull, []command.ProjectResult{
		discardGenerationResult(command.State, "missing-project", "missing", ""),
	})
	assertPlanGenerationError(t, err, coredb.ErrPlanGenerationStateInvalid)

	status, err := database.UpdateDiscardResultsForPlanGeneration(pull, []command.ProjectResult{
		discardGenerationResult(command.Import, "project-a", "a", ""),
	})
	Ok(t, err)
	Equals(t, models.DiscardedPlanStatus, status.Projects[0].Status)
	Equals(t, "", status.Projects[0].ManagedPlanHash)
	Equals(t, "", status.Projects[0].AcceptedPlanGeneration)
	_, err = database.BeginPlanGeneration(pull, []models.ProjectStatus{{
		Workspace: "default", RepoRelDir: "project-a", ProjectName: "a",
	}}, "generation-active")
	Ok(t, err)
	_, err = database.UpdateDiscardResultsForPlanGeneration(pull, []command.ProjectResult{
		discardGenerationResult(command.State, "project-a", "a", ""),
	})
	assertPlanGenerationError(t, err, coredb.ErrPlanGenerationSuperseded)
}

func TestUpdateDiscardResultsForPlanGenerationUnlockUsesExactNamedProject(t *testing.T) {
	database, err := boltdb.New(t.TempDir())
	Ok(t, err)
	t.Cleanup(func() { database.Close() })
	pull := planGenerationTestPull()
	projectA := models.ProjectStatus{Workspace: "default", RepoRelDir: "shared-dir", ProjectName: "a"}
	projectB := models.ProjectStatus{Workspace: "default", RepoRelDir: "shared-dir", ProjectName: "b"}
	_, err = database.BeginPlanGeneration(pull, []models.ProjectStatus{projectA}, "generation-a")
	Ok(t, err)
	_, err = database.CompletePlanGeneration(pull, "generation-a", []command.ProjectResult{
		managedPlanGenerationResult("shared-dir", "a", "hash-a"),
	})
	Ok(t, err)
	_, err = database.BeginPlanGeneration(pull, []models.ProjectStatus{projectB}, "generation-b")
	Ok(t, err)
	_, err = database.CompletePlanGeneration(pull, "generation-b", []command.ProjectResult{
		managedPlanGenerationResult("shared-dir", "b", "hash-b"),
	})
	Ok(t, err)

	_, err = database.UpdateDiscardResultsForPlanGeneration(pull, []command.ProjectResult{
		discardGenerationResult(command.Unlock, "shared-dir", "b", "generation-b"),
		discardGenerationResult(command.Unlock, "shared-dir", "a", "generation-b"),
	})
	assertPlanGenerationError(t, err, coredb.ErrPlanGenerationSuperseded)
	unchanged, err := database.GetPullStatus(pull)
	Ok(t, err)
	Equals(t, models.PlannedPlanStatus, findPlanGenerationProject(t, *unchanged, "shared-dir", "a").Status)
	Equals(t, "hash-a", findPlanGenerationProject(t, *unchanged, "shared-dir", "a").ManagedPlanHash)
	Equals(t, models.PlannedPlanStatus, findPlanGenerationProject(t, *unchanged, "shared-dir", "b").Status)
	Equals(t, "hash-b", findPlanGenerationProject(t, *unchanged, "shared-dir", "b").ManagedPlanHash)

	status, err := database.UpdateDiscardResultsForPlanGeneration(pull, []command.ProjectResult{
		discardGenerationResult(command.Unlock, "shared-dir", "a", "generation-a"),
	})
	Ok(t, err)
	Equals(t, models.DiscardedPlanStatus, findPlanGenerationProject(t, status, "shared-dir", "a").Status)
	Equals(t, "", findPlanGenerationProject(t, status, "shared-dir", "a").ManagedPlanHash)
	Equals(t, models.PlannedPlanStatus, findPlanGenerationProject(t, status, "shared-dir", "b").Status)
	Equals(t, "hash-b", findPlanGenerationProject(t, status, "shared-dir", "b").ManagedPlanHash)
	Equals(t, "generation-b", findPlanGenerationProject(t, status, "shared-dir", "b").AcceptedPlanGeneration)

	_, err = database.BeginPlanGeneration(pull, []models.ProjectStatus{projectA}, "generation-active")
	Ok(t, err)
	_, err = database.UpdateDiscardResultsForPlanGeneration(pull, []command.ProjectResult{
		discardGenerationResult(command.Unlock, "shared-dir", "a", ""),
	})
	assertPlanGenerationError(t, err, coredb.ErrPlanGenerationSuperseded)
}

func TestUpdateDiscardResultsForPlanGenerationReturnsBoltErrors(t *testing.T) {
	database, err := boltdb.New(t.TempDir())
	Ok(t, err)
	Ok(t, database.Close())
	_, err = database.UpdateDiscardResultsForPlanGeneration(planGenerationTestPull(), []command.ProjectResult{
		discardGenerationResult(command.Import, "project-a", "a", "generation-a"),
	})
	Assert(t, err != nil, "expected Bolt backend error")
	Assert(t, !errors.Is(err, coredb.ErrPlanGenerationSuperseded), "backend error must remain ordinary")
	Assert(t, !errors.Is(err, coredb.ErrPlanPublicationNotOwned), "backend error must remain ordinary")
}

func TestManagedPlanHashLifecycle(t *testing.T) {
	dataDir := t.TempDir()
	database, err := boltdb.New(dataDir)
	Ok(t, err)
	pull := planGenerationTestPull()
	project := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-a", ProjectName: "a"}

	_, err = database.BeginPlanGeneration(pull, []models.ProjectStatus{project}, "generation-1", "")
	Ok(t, err)
	status, err := database.CompletePlanGeneration(pull, "generation-1", []command.ProjectResult{
		managedPlanGenerationResult("project-a", "a", "hash-1"),
	}, "")

	Ok(t, err)
	Equals(t, "hash-1", findPlanGenerationProject(t, status, "project-a", "a").ManagedPlanHash)
	Equals(t, "generation-1", findPlanGenerationProject(t, status, "project-a", "a").AcceptedPlanGeneration)
	Ok(t, database.Close())

	database, err = boltdb.New(dataDir)
	Ok(t, err)
	t.Cleanup(func() { database.Close() })
	restarted, err := database.GetPullStatus(pull)
	Ok(t, err)
	Equals(t, "hash-1", findPlanGenerationProject(t, *restarted, "project-a", "a").ManagedPlanHash)
	Equals(t, "generation-1", findPlanGenerationProject(t, *restarted, "project-a", "a").AcceptedPlanGeneration)

	_, err = database.UpdatePullWithResults(pull, []command.ProjectResult{{
		Command: command.Plan, Workspace: "default", RepoRelDir: "project-b", ProjectName: "b",
		ProjectCommandOutput: command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}},
	}})
	Ok(t, err)
	status, err = database.UpdatePullWithResults(pull, []command.ProjectResult{{
		Command: command.PolicyCheck, Workspace: "default", RepoRelDir: "project-a", ProjectName: "a",
		ProjectCommandOutput: command.ProjectCommandOutput{PolicyCheckResults: &models.PolicyCheckResults{}},
	}})
	Ok(t, err)
	Equals(t, "hash-1", findPlanGenerationProject(t, status, "project-a", "a").ManagedPlanHash)
	Equals(t, "generation-1", findPlanGenerationProject(t, status, "project-a", "a").AcceptedPlanGeneration)
	Equals(t, "", findPlanGenerationProject(t, status, "project-b", "b").ManagedPlanHash)
	Equals(t, "", findPlanGenerationProject(t, status, "project-b", "b").AcceptedPlanGeneration)

	status, err = database.UpdatePullWithResults(pull, []command.ProjectResult{{
		Command: command.ApprovePolicies, Workspace: "default", RepoRelDir: "project-a", ProjectName: "a",
		ProjectCommandOutput: command.ProjectCommandOutput{PolicyCheckResults: &models.PolicyCheckResults{}},
	}})
	Ok(t, err)
	Equals(t, "hash-1", findPlanGenerationProject(t, status, "project-a", "a").ManagedPlanHash)
	Equals(t, "generation-1", findPlanGenerationProject(t, status, "project-a", "a").AcceptedPlanGeneration)
	status, err = database.UpdatePullWithResults(pull, []command.ProjectResult{{
		Command: command.Apply, Workspace: "default", RepoRelDir: "project-a", ProjectName: "a",
		ProjectCommandOutput: command.ProjectCommandOutput{Error: errors.New("apply failed")},
	}})
	Ok(t, err)
	Equals(t, "hash-1", findPlanGenerationProject(t, status, "project-a", "a").ManagedPlanHash)
	Equals(t, "generation-1", findPlanGenerationProject(t, status, "project-a", "a").AcceptedPlanGeneration)
	status, err = database.UpdatePullWithResults(pull, []command.ProjectResult{{
		Command: command.Apply, Workspace: "default", RepoRelDir: "project-a", ProjectName: "a",
		ProjectCommandOutput: command.ProjectCommandOutput{ApplySuccess: "applied"},
	}})
	Ok(t, err)
	Equals(t, "", findPlanGenerationProject(t, status, "project-a", "a").ManagedPlanHash)
	Equals(t, "", findPlanGenerationProject(t, status, "project-a", "a").AcceptedPlanGeneration)

	_, err = database.BeginPlanGeneration(pull, []models.ProjectStatus{project}, "generation-2", "")
	Ok(t, err)
	status, err = database.CompletePlanGeneration(pull, "generation-2", []command.ProjectResult{
		managedPlanGenerationResult("project-a", "a", "hash-2"),
	}, "")

	Ok(t, err)
	Equals(t, "hash-2", findPlanGenerationProject(t, status, "project-a", "a").ManagedPlanHash)
	Equals(t, "generation-2", findPlanGenerationProject(t, status, "project-a", "a").AcceptedPlanGeneration)
	beginStatus, err := database.BeginPlanGeneration(pull, []models.ProjectStatus{project}, "generation-3", "")
	Ok(t, err)
	Equals(t, "", findPlanGenerationProject(t, beginStatus, "project-a", "a").ManagedPlanHash)
	Equals(t, "", findPlanGenerationProject(t, beginStatus, "project-a", "a").AcceptedPlanGeneration)
	status, err = database.CompletePlanGeneration(pull, "generation-3", []command.ProjectResult{{
		Command: command.Plan, Workspace: "default", RepoRelDir: "project-a", ProjectName: "a",
		ProjectCommandOutput: command.ProjectCommandOutput{Error: errors.New("plan failed"), AtlantisManagedPlan: true},
	}}, "")

	Ok(t, err)
	Equals(t, "", findPlanGenerationProject(t, status, "project-a", "a").ManagedPlanHash)
	Equals(t, "", findPlanGenerationProject(t, status, "project-a", "a").AcceptedPlanGeneration)

	_, err = database.BeginPlanGeneration(pull, []models.ProjectStatus{project}, "generation-4", "")
	Ok(t, err)
	_, err = database.CompletePlanGeneration(pull, "generation-4", []command.ProjectResult{
		managedPlanGenerationResult("project-a", "a", "hash-4"),
	}, "")

	Ok(t, err)
	status, err = database.UpdatePullWithResults(pull, []command.ProjectResult{{
		Command: command.Import, Workspace: "default", RepoRelDir: "project-a", ProjectName: "a",
		ProjectCommandOutput: command.ProjectCommandOutput{ImportSuccess: &models.ImportSuccess{}},
	}})
	Ok(t, err)
	Equals(t, "", findPlanGenerationProject(t, status, "project-a", "a").ManagedPlanHash)
	Equals(t, "", findPlanGenerationProject(t, status, "project-a", "a").AcceptedPlanGeneration)

	_, err = database.BeginPlanGeneration(pull, []models.ProjectStatus{project}, "generation-5", "")
	Ok(t, err)
	_, err = database.CompletePlanGeneration(pull, "generation-5", []command.ProjectResult{
		managedPlanGenerationResult("project-a", "a", "hash-5"),
	}, "")

	Ok(t, err)
	Ok(t, database.UpdateProjectStatus(pull, "default", "project-a", models.DiscardedPlanStatus))
	statusPtr, err := database.GetPullStatus(pull)
	Ok(t, err)
	Equals(t, "", findPlanGenerationProject(t, *statusPtr, "project-a", "a").ManagedPlanHash)
	Equals(t, "", findPlanGenerationProject(t, *statusPtr, "project-a", "a").AcceptedPlanGeneration)
}

func planGenerationTestPull() models.PullRequest {
	return models.PullRequest{
		Num:        1,
		HeadCommit: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		BaseBranch: "main",
		BaseRepo: models.Repo{
			FullName: "runatlantis/atlantis",
			VCSHost:  models.VCSHost{Hostname: "github.com", Type: models.Github},
		},
	}
}

func managedPlanGenerationResult(dir, projectName, hash string) command.ProjectResult {
	return command.ProjectResult{
		Command:     command.Plan,
		Workspace:   "default",
		RepoRelDir:  dir,
		ProjectName: projectName,
		ProjectCommandOutput: command.ProjectCommandOutput{
			PlanSuccess:         &models.PlanSuccess{},
			AtlantisManagedPlan: true,
			ManagedPlanHash:     hash,
		},
	}
}

func simplePlanGenerationResult(dir, projectName string) command.ProjectResult {
	return command.ProjectResult{
		Command: command.Plan, Workspace: "default", RepoRelDir: dir, ProjectName: projectName,
		ProjectCommandOutput: command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}},
	}
}

func policyGenerationResult(dir, projectName, acceptedGeneration, policyName string) command.ProjectResult {
	return command.ProjectResult{
		Command:                command.PolicyCheck,
		Workspace:              "default",
		RepoRelDir:             dir,
		ProjectName:            projectName,
		AcceptedPlanGeneration: acceptedGeneration,
		ProjectCommandOutput: command.ProjectCommandOutput{PolicyCheckResults: &models.PolicyCheckResults{
			PolicySetResults: []models.PolicySetResult{{PolicySetName: policyName, Passed: true}},
		}},
	}
}

func applyGenerationResult(dir, projectName, acceptedGeneration string, applyErr error) command.ProjectResult {
	result := command.ProjectResult{
		Command:                command.Apply,
		Workspace:              "default",
		RepoRelDir:             dir,
		ProjectName:            projectName,
		AcceptedPlanGeneration: acceptedGeneration,
		ProjectCommandOutput:   command.ProjectCommandOutput{ApplySuccess: "applied"},
	}
	if applyErr != nil {
		result.Error = applyErr
		result.ApplySuccess = ""
	}
	return result
}

func discardGenerationResult(commandName command.Name, dir, projectName, acceptedGeneration string) command.ProjectResult {
	result := command.ProjectResult{
		Command:                commandName,
		Workspace:              "default",
		RepoRelDir:             dir,
		ProjectName:            projectName,
		AcceptedPlanGeneration: acceptedGeneration,
	}
	switch commandName {
	case command.Import:
		result.ImportSuccess = &models.ImportSuccess{}
	case command.State:
		result.StateRmSuccess = &models.StateRmSuccess{}
	}
	return result
}

func assertPlanGenerationError(t *testing.T, err, target error) {
	t.Helper()
	Assert(t, errors.Is(err, target), "expected %v classification, got %v", target, err)
	var completionErr *coredb.PlanGenerationCompletionError
	Assert(t, errors.As(err, &completionErr), "expected typed completion error, got %T", err)
}

func assertPlanPublicationClaimError(t *testing.T, err, target error) {
	t.Helper()
	Assert(t, errors.Is(err, target), "expected %v classification, got %v", target, err)
	var claimErr *coredb.PlanPublicationClaimError
	Assert(t, errors.As(err, &claimErr), "expected typed claim error, got %T", err)
}

func findPlanGenerationProject(t *testing.T, statusValue any, dir, projectName string) models.ProjectStatus {
	t.Helper()
	var status models.PullStatus
	switch value := statusValue.(type) {
	case models.PullStatus:
		status = value
	case coredb.PlanGenerationBeginResult:
		status = value.PullStatus
	default:
		t.Fatalf("unexpected pull status type %T", statusValue)
	}
	for _, project := range status.Projects {
		if project.RepoRelDir == dir && project.ProjectName == projectName {
			return project
		}
	}
	t.Fatalf("project %q at %q not found", projectName, dir)
	return models.ProjectStatus{}
}

// newTestDB returns a TestDB using a temporary path.
func newTestDB() (*bolt.DB, *boltdb.BoltDB) {
	// Retrieve a temporary path.
	f, err := os.CreateTemp("", "")
	if err != nil {
		panic(fmt.Errorf("failed to create temp file: %w", err))
	}
	path := f.Name()
	f.Close() // nolint: errcheck

	// Open the database.
	boltDB, err := bolt.Open(path, 0600, nil)
	if err != nil {
		panic(fmt.Errorf("could not start bolt DB: %w", err))
	}
	if err := boltDB.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(lockBucket)); err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		if _, err := tx.CreateBucketIfNotExists([]byte(configBucket)); err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		return nil
	}); err != nil {
		panic(fmt.Errorf("could not create bucket: %w", err))
	}
	b, _ := boltdb.NewWithDB(boltDB, lockBucket, configBucket)
	return boltDB, b
}

func newTestDB2(t *testing.T) *boltdb.BoltDB {
	tmp := t.TempDir()
	boltDB, err := boltdb.New(tmp)
	Ok(t, err)
	return boltDB
}

func cleanupDB(db *bolt.DB) {
	db.Close()           // nolint: errcheck
	os.Remove(db.Path()) // nolint: errcheck
}
