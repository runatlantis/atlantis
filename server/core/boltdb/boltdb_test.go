// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

package boltdb_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/core/boltdb"

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
								PolicySetName: "policy1",
								ReqApprovals:  1,
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
								PolicySetName: "policy1",
								ReqApprovals:  1,
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
								PolicySetName: "policy1",
								ReqApprovals:  1,
								CurApprovals:  1,
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
						Approvals:     1,
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
						Approvals:     0,
					},
				},
			},
		}, updateStatus.Projects)
	}
	b.Close()
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

func TestBoltDB_SaveAndGetProjectOutput(t *testing.T) {
	db := newTestDB2(t)
	defer db.Close()

	now := time.Now()
	output := models.ProjectOutput{
		RepoFullName:  "owner/repo",
		PullNum:       123,
		ProjectName:   "myproject",
		Workspace:     "default",
		Path:          "terraform/staging",
		CommandName:   "plan",
		JobID:         "job-123",
		RunTimestamp:  now.UnixMilli(),
		Output:        "Plan: 1 to add, 0 to change, 0 to destroy.",
		Status:        models.SuccessOutputStatus,
		ResourceStats: models.ResourceStats{Add: 1},
		TriggeredBy:   "testuser",
		StartedAt:     now.Add(-time.Minute),
		CompletedAt:   now,
	}

	err := db.SaveProjectOutput(output)
	Ok(t, err)

	// Test GetProjectOutputHistory
	history, err := db.GetProjectOutputHistory("owner/repo", 123, "terraform/staging", "default", "myproject")
	Ok(t, err)
	Assert(t, len(history) == 1, "expected 1 result")
	Equals(t, output.Output, history[0].Output)
	Equals(t, output.Status, history[0].Status)
	Equals(t, output.ResourceStats.Add, history[0].ResourceStats.Add)

	// Test GetProjectOutputRun
	retrieved, err := db.GetProjectOutputRun("owner/repo", 123, "terraform/staging", "default", "myproject", "plan", now.UnixMilli())
	Ok(t, err)
	Assert(t, retrieved != nil, "expected non-nil result")
	Equals(t, output.Output, retrieved.Output)
}

func TestBoltDB_GetProjectOutputHistory_NotFound(t *testing.T) {
	db := newTestDB2(t)
	defer db.Close()

	history, err := db.GetProjectOutputHistory("owner/repo", 999, "nonexistent", "default", "")
	Ok(t, err)
	Assert(t, len(history) == 0, "expected empty history for non-existent output")
}

func TestBoltDB_GetProjectOutputsByPull(t *testing.T) {
	db := newTestDB2(t)
	defer db.Close()

	// Save two outputs for the same PR
	output1 := models.ProjectOutput{
		RepoFullName: "owner/repo",
		PullNum:      123,
		ProjectName:  "project1",
		Workspace:    "default",
		Path:         "terraform/staging",
		CommandName:  "plan",
		Status:       models.SuccessOutputStatus,
	}
	output2 := models.ProjectOutput{
		RepoFullName: "owner/repo",
		PullNum:      123,
		ProjectName:  "project2",
		Workspace:    "default",
		Path:         "terraform/prod",
		CommandName:  "plan",
		Status:       models.FailedOutputStatus,
	}

	err := db.SaveProjectOutput(output1)
	Ok(t, err)
	err = db.SaveProjectOutput(output2)
	Ok(t, err)

	outputs, err := db.GetProjectOutputsByPull("owner/repo", 123)
	Ok(t, err)
	Equals(t, 2, len(outputs))
}

func TestBoltDB_GetProjectOutputsByPull_SkipsPolicyCheck(t *testing.T) {
	db := newTestDB2(t)
	defer db.Close()

	now := time.Now()

	// Save a plan record and a policy_check record for the same project
	planOutput := models.ProjectOutput{
		RepoFullName: "owner/repo",
		PullNum:      123,
		ProjectName:  "project1",
		Workspace:    "default",
		Path:         "terraform/staging",
		CommandName:  "plan",
		RunTimestamp: now.Add(-1 * time.Minute).UnixMilli(),
		Status:       models.SuccessOutputStatus,
		Output:       "Plan: 1 to add",
	}
	policyOutput := models.ProjectOutput{
		RepoFullName: "owner/repo",
		PullNum:      123,
		ProjectName:  "project1",
		Workspace:    "default",
		Path:         "terraform/staging",
		CommandName:  "policy_check",
		RunTimestamp: now.UnixMilli(),
		Status:       models.SuccessOutputStatus,
		PolicyPassed: true,
		PolicyOutput: "Policies passed: 2/2",
	}

	err := db.SaveProjectOutput(planOutput)
	Ok(t, err)
	err = db.SaveProjectOutput(policyOutput)
	Ok(t, err)

	outputs, err := db.GetProjectOutputsByPull("owner/repo", 123)
	Ok(t, err)

	// Should return only the plan record, not the policy_check
	Equals(t, 1, len(outputs))
	Equals(t, "plan", outputs[0].CommandName)

	// Policy data should be merged into the plan record
	Equals(t, true, outputs[0].PolicyPassed)
	Equals(t, "Policies passed: 2/2", outputs[0].PolicyOutput)
}

func TestBoltDB_GetProjectOutputsByPull_Empty(t *testing.T) {
	db := newTestDB2(t)
	defer db.Close()

	outputs, err := db.GetProjectOutputsByPull("owner/repo", 999)
	Ok(t, err)
	Equals(t, 0, len(outputs))
}

func TestBoltDB_DeleteProjectOutputsByPull(t *testing.T) {
	db := newTestDB2(t)
	defer db.Close()

	output := models.ProjectOutput{
		RepoFullName: "owner/repo",
		PullNum:      123,
		ProjectName:  "myproject",
		Workspace:    "default",
		Path:         "terraform/staging",
	}
	err := db.SaveProjectOutput(output)
	Ok(t, err)

	err = db.DeleteProjectOutputsByPull("owner/repo", 123)
	Ok(t, err)

	outputs, err := db.GetProjectOutputsByPull("owner/repo", 123)
	Ok(t, err)
	Equals(t, 0, len(outputs))
}

func TestBoltDB_GetActivePullRequests(t *testing.T) {
	db := newTestDB2(t)
	defer db.Close()

	// Save outputs for two different PRs
	output1 := models.ProjectOutput{
		RepoFullName: "owner/repo",
		PullNum:      123,
		ProjectName:  "project1",
		Workspace:    "default",
		Path:         "terraform/staging",
	}
	output2 := models.ProjectOutput{
		RepoFullName: "owner/repo",
		PullNum:      456,
		ProjectName:  "project2",
		Workspace:    "default",
		Path:         "terraform/prod",
	}

	err := db.SaveProjectOutput(output1)
	Ok(t, err)
	err = db.SaveProjectOutput(output2)
	Ok(t, err)

	pulls, err := db.GetActivePullRequests()
	Ok(t, err)
	Equals(t, 2, len(pulls))
}

func TestBoltDB_GetActivePullRequests_Empty(t *testing.T) {
	db := newTestDB2(t)
	defer db.Close()

	pulls, err := db.GetActivePullRequests()
	Ok(t, err)
	Equals(t, 0, len(pulls))
}

func TestBoltDB_GetActivePullRequests_PullIndexPreservesMetadata(t *testing.T) {
	db := newTestDB2(t)
	defer db.Close()

	// Save output with PR metadata
	output := models.ProjectOutput{
		RepoFullName: "owner/repo",
		PullNum:      42,
		ProjectName:  "myproject",
		Workspace:    "default",
		Path:         "terraform/staging",
		PullURL:      "https://github.com/owner/repo/pull/42",
		PullTitle:    "Add new feature",
	}

	err := db.SaveProjectOutput(output)
	Ok(t, err)

	pulls, err := db.GetActivePullRequests()
	Ok(t, err)
	Equals(t, 1, len(pulls))
	Equals(t, 42, pulls[0].Num)
	Equals(t, "https://github.com/owner/repo/pull/42", pulls[0].URL)
	Equals(t, "Add new feature", pulls[0].Title)
	Equals(t, "owner/repo", pulls[0].BaseRepo.FullName)
}

func TestBoltDB_GetActivePullRequests_PullIndexUpdatedOnSave(t *testing.T) {
	db := newTestDB2(t)
	defer db.Close()

	// Save first output without PR metadata
	output1 := models.ProjectOutput{
		RepoFullName: "owner/repo",
		PullNum:      42,
		ProjectName:  "project1",
		Workspace:    "default",
		Path:         "terraform/staging",
	}
	err := db.SaveProjectOutput(output1)
	Ok(t, err)

	// Save second output for the same PR with metadata
	output2 := models.ProjectOutput{
		RepoFullName: "owner/repo",
		PullNum:      42,
		ProjectName:  "project2",
		Workspace:    "default",
		Path:         "terraform/prod",
		PullURL:      "https://github.com/owner/repo/pull/42",
		PullTitle:    "Updated title",
	}
	err = db.SaveProjectOutput(output2)
	Ok(t, err)

	// The pull index should reflect the latest save's metadata
	pulls, err := db.GetActivePullRequests()
	Ok(t, err)
	Equals(t, 1, len(pulls))
	Equals(t, 42, pulls[0].Num)
	Equals(t, "https://github.com/owner/repo/pull/42", pulls[0].URL)
	Equals(t, "Updated title", pulls[0].Title)
}

func TestBoltDB_GetActivePullRequests_PullIndexCleanedOnDelete(t *testing.T) {
	db := newTestDB2(t)
	defer db.Close()

	output := models.ProjectOutput{
		RepoFullName: "owner/repo",
		PullNum:      42,
		ProjectName:  "myproject",
		Workspace:    "default",
		Path:         "terraform/staging",
		PullURL:      "https://github.com/owner/repo/pull/42",
		PullTitle:    "My PR",
	}

	err := db.SaveProjectOutput(output)
	Ok(t, err)

	// Verify pull is active
	pulls, err := db.GetActivePullRequests()
	Ok(t, err)
	Equals(t, 1, len(pulls))

	// Delete the outputs for that pull
	err = db.DeleteProjectOutputsByPull("owner/repo", 42)
	Ok(t, err)

	// Pull index should be empty now
	pulls, err = db.GetActivePullRequests()
	Ok(t, err)
	Equals(t, 0, len(pulls))
}

func TestBoltDB_GetActivePullRequests_MultiplePRsAcrossRepos(t *testing.T) {
	db := newTestDB2(t)
	defer db.Close()

	outputs := []models.ProjectOutput{
		{
			RepoFullName: "owner/repo1",
			PullNum:      1,
			ProjectName:  "project",
			Workspace:    "default",
			Path:         "terraform",
			PullURL:      "https://github.com/owner/repo1/pull/1",
		},
		{
			RepoFullName: "owner/repo1",
			PullNum:      2,
			ProjectName:  "project",
			Workspace:    "default",
			Path:         "terraform",
			PullURL:      "https://github.com/owner/repo1/pull/2",
		},
		{
			RepoFullName: "owner/repo2",
			PullNum:      1,
			ProjectName:  "project",
			Workspace:    "default",
			Path:         "terraform",
			PullURL:      "https://github.com/owner/repo2/pull/1",
		},
	}

	for _, o := range outputs {
		err := db.SaveProjectOutput(o)
		Ok(t, err)
	}

	pulls, err := db.GetActivePullRequests()
	Ok(t, err)
	Equals(t, 3, len(pulls))

	// Delete one PR's outputs
	err = db.DeleteProjectOutputsByPull("owner/repo1", 1)
	Ok(t, err)

	pulls, err = db.GetActivePullRequests()
	Ok(t, err)
	Equals(t, 2, len(pulls))
}

func TestBoltDB_GetProjectOutputHistory_SortsDescending(t *testing.T) {
	db := newTestDB2(t)
	defer db.Close()

	base := time.Now()

	// Save outputs with different timestamps (out of order)
	timestamps := []int64{
		base.Add(1 * time.Minute).UnixMilli(),
		base.Add(3 * time.Minute).UnixMilli(),
		base.Add(2 * time.Minute).UnixMilli(),
	}

	for _, ts := range timestamps {
		output := models.ProjectOutput{
			RepoFullName: "owner/repo",
			PullNum:      123,
			ProjectName:  "project",
			Workspace:    "default",
			Path:         "terraform",
			CommandName:  "plan",
			RunTimestamp: ts,
			Status:       models.SuccessOutputStatus,
		}
		err := db.SaveProjectOutput(output)
		Ok(t, err)
	}

	// Get history and verify it's sorted descending (newest first)
	history, err := db.GetProjectOutputHistory("owner/repo", 123, "terraform", "default", "project")
	Ok(t, err)
	Equals(t, 3, len(history))

	// Verify descending order
	Assert(t, history[0].RunTimestamp > history[1].RunTimestamp,
		"expected %d > %d", history[0].RunTimestamp, history[1].RunTimestamp)
	Assert(t, history[1].RunTimestamp > history[2].RunTimestamp,
		"expected %d > %d", history[1].RunTimestamp, history[2].RunTimestamp)
}

func TestBoltDB_GetProjectOutputByJobID(t *testing.T) {
	db := newTestDB2(t)
	defer db.Close()

	now := time.Now()
	output := models.ProjectOutput{
		RepoFullName: "owner/repo",
		PullNum:      123,
		ProjectName:  "myproject",
		Workspace:    "default",
		Path:         "terraform/staging",
		CommandName:  "plan",
		JobID:        "job-abc-123",
		RunTimestamp: now.UnixMilli(),
		Output:       "Plan: 1 to add",
		Status:       models.SuccessOutputStatus,
	}

	err := db.SaveProjectOutput(output)
	Ok(t, err)

	// Test retrieval by job ID (should use index for O(1) lookup)
	retrieved, err := db.GetProjectOutputByJobID("job-abc-123")
	Ok(t, err)
	Assert(t, retrieved != nil, "expected non-nil result")
	Equals(t, output.JobID, retrieved.JobID)
	Equals(t, output.Output, retrieved.Output)
	Equals(t, output.RepoFullName, retrieved.RepoFullName)
}

func TestBoltDB_GetProjectOutputByJobID_NotFound(t *testing.T) {
	db := newTestDB2(t)
	defer db.Close()

	// Test retrieval of non-existent job ID
	retrieved, err := db.GetProjectOutputByJobID("nonexistent-job-id")
	Ok(t, err)
	Assert(t, retrieved == nil, "expected nil result for non-existent job ID")
}

func TestBoltDB_MarkInterruptedOutputs_FlipsRunning(t *testing.T) {
	db := newTestDB2(t)
	defer db.Close()

	now := time.Now()
	// Save a running output
	running := models.ProjectOutput{
		RepoFullName: "owner/repo",
		PullNum:      123,
		ProjectName:  "myproject",
		Workspace:    "default",
		Path:         "terraform/staging",
		CommandName:  "plan",
		JobID:        "job-running",
		RunTimestamp: now.UnixMilli(),
		Status:       models.RunningOutputStatus,
	}
	err := db.SaveProjectOutput(running)
	Ok(t, err)

	// Mark interrupted
	err = db.MarkInterruptedOutputs()
	Ok(t, err)

	// Verify it was flipped
	retrieved, err := db.GetProjectOutputByJobID("job-running")
	Ok(t, err)
	Assert(t, retrieved != nil, "expected non-nil result")
	Equals(t, models.InterruptedOutputStatus, retrieved.Status)
}

func TestBoltDB_MarkInterruptedOutputs_LeavesOtherStatuses(t *testing.T) {
	db := newTestDB2(t)
	defer db.Close()

	now := time.Now()
	// Save a success output
	success := models.ProjectOutput{
		RepoFullName: "owner/repo",
		PullNum:      123,
		ProjectName:  "project1",
		Workspace:    "default",
		Path:         "terraform/staging",
		CommandName:  "plan",
		JobID:        "job-success",
		RunTimestamp: now.UnixMilli(),
		Status:       models.SuccessOutputStatus,
		Output:       "Plan: 1 to add",
	}
	// Save a failed output
	failed := models.ProjectOutput{
		RepoFullName: "owner/repo",
		PullNum:      123,
		ProjectName:  "project2",
		Workspace:    "default",
		Path:         "terraform/prod",
		CommandName:  "plan",
		JobID:        "job-failed",
		RunTimestamp: now.Add(time.Millisecond).UnixMilli(),
		Status:       models.FailedOutputStatus,
		Output:       "Error: something went wrong",
	}

	err := db.SaveProjectOutput(success)
	Ok(t, err)
	err = db.SaveProjectOutput(failed)
	Ok(t, err)

	// Mark interrupted
	err = db.MarkInterruptedOutputs()
	Ok(t, err)

	// Verify success is untouched
	retrievedSuccess, err := db.GetProjectOutputByJobID("job-success")
	Ok(t, err)
	Assert(t, retrievedSuccess != nil, "expected non-nil result")
	Equals(t, models.SuccessOutputStatus, retrievedSuccess.Status)

	// Verify failed is untouched
	retrievedFailed, err := db.GetProjectOutputByJobID("job-failed")
	Ok(t, err)
	Assert(t, retrievedFailed != nil, "expected non-nil result")
	Equals(t, models.FailedOutputStatus, retrievedFailed.Status)
}

func TestBoltDB_SaveProjectOutput_UpsertByJobID(t *testing.T) {
	db := newTestDB2(t)
	defer db.Close()

	now := time.Now()
	// Save initial stub with running status
	stub := models.ProjectOutput{
		RepoFullName: "owner/repo",
		PullNum:      123,
		ProjectName:  "myproject",
		Workspace:    "default",
		Path:         "terraform/staging",
		CommandName:  "plan",
		JobID:        "job-upsert-123",
		RunTimestamp: now.UnixMilli(),
		Status:       models.RunningOutputStatus,
		TriggeredBy:  "testuser",
		StartedAt:    now,
	}
	err := db.SaveProjectOutput(stub)
	Ok(t, err)

	// Now upsert with completed output (same JobID, different timestamp in Key)
	completed := models.ProjectOutput{
		RepoFullName:  "owner/repo",
		PullNum:       123,
		ProjectName:   "myproject",
		Workspace:     "default",
		Path:          "terraform/staging",
		CommandName:   "plan",
		JobID:         "job-upsert-123",
		RunTimestamp:  now.UnixMilli(),
		Status:        models.SuccessOutputStatus,
		Output:        "Plan: 1 to add, 0 to change, 0 to destroy.",
		ResourceStats: models.ResourceStats{Add: 1},
		TriggeredBy:   "testuser",
		StartedAt:     now,
		CompletedAt:   now.Add(30 * time.Second),
	}
	err = db.SaveProjectOutput(completed)
	Ok(t, err)

	// Verify only one record exists (not two)
	history, err := db.GetProjectOutputHistory("owner/repo", 123, "terraform/staging", "default", "myproject")
	Ok(t, err)
	Equals(t, 1, len(history))
	Equals(t, models.SuccessOutputStatus, history[0].Status)
	Equals(t, "Plan: 1 to add, 0 to change, 0 to destroy.", history[0].Output)

	// Verify lookup by job ID returns the updated record
	retrieved, err := db.GetProjectOutputByJobID("job-upsert-123")
	Ok(t, err)
	Assert(t, retrieved != nil, "expected non-nil result")
	Equals(t, models.SuccessOutputStatus, retrieved.Status)
	Equals(t, completed.Output, retrieved.Output)
}

func TestBoltDB_SaveProjectOutput_InsertNewRecord(t *testing.T) {
	db := newTestDB2(t)
	defer db.Close()

	now := time.Now()
	// Save output without a matching job ID in the index
	output := models.ProjectOutput{
		RepoFullName: "owner/repo",
		PullNum:      123,
		ProjectName:  "myproject",
		Workspace:    "default",
		Path:         "terraform/staging",
		CommandName:  "plan",
		JobID:        "job-new-456",
		RunTimestamp: now.UnixMilli(),
		Status:       models.SuccessOutputStatus,
		Output:       "Plan output",
	}
	err := db.SaveProjectOutput(output)
	Ok(t, err)

	// Save a second output with a different job ID
	output2 := models.ProjectOutput{
		RepoFullName: "owner/repo",
		PullNum:      123,
		ProjectName:  "myproject",
		Workspace:    "default",
		Path:         "terraform/staging",
		CommandName:  "plan",
		JobID:        "job-new-789",
		RunTimestamp: now.Add(time.Minute).UnixMilli(),
		Status:       models.SuccessOutputStatus,
		Output:       "Plan output 2",
	}
	err = db.SaveProjectOutput(output2)
	Ok(t, err)

	// Verify two separate records exist
	history, err := db.GetProjectOutputHistory("owner/repo", 123, "terraform/staging", "default", "myproject")
	Ok(t, err)
	Equals(t, 2, len(history))

	// Verify each can be found by job ID
	r1, err := db.GetProjectOutputByJobID("job-new-456")
	Ok(t, err)
	Assert(t, r1 != nil, "expected non-nil result for job-new-456")
	Equals(t, "Plan output", r1.Output)

	r2, err := db.GetProjectOutputByJobID("job-new-789")
	Ok(t, err)
	Assert(t, r2 != nil, "expected non-nil result for job-new-789")
	Equals(t, "Plan output 2", r2.Output)
}

func TestBoltDB_GetProjectOutputByJobID_EmptyJobID(t *testing.T) {
	db := newTestDB2(t)
	defer db.Close()

	// Save output with empty job ID - should not be indexed
	output := models.ProjectOutput{
		RepoFullName: "owner/repo",
		PullNum:      123,
		ProjectName:  "myproject",
		Workspace:    "default",
		Path:         "terraform/staging",
		CommandName:  "plan",
		JobID:        "", // Empty job ID
		RunTimestamp: time.Now().UnixMilli(),
	}

	err := db.SaveProjectOutput(output)
	Ok(t, err)

	// Empty job IDs are not indexed, but the fallback scan will still find them
	// This verifies backwards compatibility - outputs without job IDs can still be found
	retrieved, err := db.GetProjectOutputByJobID("")
	Ok(t, err)
	// The scan finds outputs with matching (empty) job ID
	Assert(t, retrieved != nil, "fallback scan should find output with empty job ID")
}
