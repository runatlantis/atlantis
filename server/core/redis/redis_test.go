// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package redis_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/pkg/errors"
	redisLib "github.com/redis/go-redis/v9"
	"github.com/runatlantis/atlantis/server/core/redis"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"

	. "github.com/runatlantis/atlantis/testing"
)

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

var (
	cert   tls.Certificate
	caPath string
)

func TestRedisWithTLS(t *testing.T) {
	t.Log("connecting to redis over TLS")

	// Setup the Miniredis Server for TLS
	certBytes, keyBytes, err := generateLocalhostCert()
	Ok(t, err)
	certOut := new(bytes.Buffer)
	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	Ok(t, err)
	certData := certOut.Bytes()
	keyOut := new(bytes.Buffer)
	err = pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: keyBytes})
	Ok(t, err)
	cert, err = tls.X509KeyPair(certData, keyOut.Bytes())
	Ok(t, err)
	certFile, err := os.CreateTemp("", "cert.*.pem")
	Ok(t, err)
	caPath = certFile.Name()
	_, err = certFile.Write(certData)
	Ok(t, err)
	defer certFile.Close()
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true, //nolint:gosec // This is purely for testing
	}

	// Start Server and Connect
	s := miniredis.NewMiniRedis()
	if err := s.StartTLS(tlsConfig); err != nil {
		t.Fatalf("could not start miniredis: %s", err)
		// not reached
	}
	t.Cleanup(s.Close)
	_ = newTestRedisTLS(s)
}

func TestLockCommandNotSet(t *testing.T) {
	t.Log("retrieving apply lock when there are none should return empty LockCommand")
	s := miniredis.RunT(t)
	r := newTestRedis(s)
	exists, err := r.CheckCommandLock(command.Apply)
	Ok(t, err)
	Assert(t, exists == nil, "exp nil")
}

func TestLockCommandEnabled(t *testing.T) {
	t.Log("setting the apply lock")
	s := miniredis.RunT(t)
	r := newTestRedis(s)
	timeNow := time.Now()
	_, err := r.LockCommand(command.Apply, timeNow)
	Ok(t, err)

	config, err := r.CheckCommandLock(command.Apply)
	Ok(t, err)
	Equals(t, true, config.IsLocked())
}

func TestLockCommandFail(t *testing.T) {
	t.Log("setting the apply lock")
	s := miniredis.RunT(t)
	r := newTestRedis(s)
	timeNow := time.Now()
	_, err := r.LockCommand(command.Apply, timeNow)
	Ok(t, err)

	_, err = r.LockCommand(command.Apply, timeNow)
	ErrEquals(t, "db transaction failed: lock already exists", err)
}

func TestUnlockCommandDisabled(t *testing.T) {
	t.Log("unsetting the apply lock")
	s := miniredis.RunT(t)
	r := newTestRedis(s)
	timeNow := time.Now()
	_, err := r.LockCommand(command.Apply, timeNow)
	Ok(t, err)

	config, err := r.CheckCommandLock(command.Apply)
	Ok(t, err)
	Equals(t, true, config.IsLocked())

	err = r.UnlockCommand(command.Apply)
	Ok(t, err)

	config, err = r.CheckCommandLock(command.Apply)
	Ok(t, err)
	Assert(t, config == nil, "exp nil object")
}

func TestUnlockCommandFail(t *testing.T) {
	t.Log("setting the apply lock")
	s := miniredis.RunT(t)
	r := newTestRedis(s)
	err := r.UnlockCommand(command.Apply)
	ErrEquals(t, "db transaction failed: no lock exists", err)
}

func TestMigrationOldLockKeysToNewFormat(t *testing.T) {
	t.Log("migration should convert old format keys to new format with project name")

	s := miniredis.RunT(t)

	// Create a direct redis client to set up old format locks
	client := redisLib.NewClient(&redisLib.Options{
		Addr: s.Addr(),
	})
	defer client.Close()

	// Create a lock in old format: {repoFullName}/{path}/{workspace}
	oldKey := "pr/owner/repo/path/default"
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

	// Insert old format lock directly
	err = client.Set(context.Background(), oldKey, oldLockSerialized, 0).Err()
	Ok(t, err)

	// Verify old key exists before migration
	val, err := client.Get(context.Background(), oldKey).Result()
	Ok(t, err)
	Assert(t, val != "", "old key should exist before migration")

	// Now create a new Redis instance which should trigger the migration
	r, err := redis.New(s.Host(), s.Server().Addr().Port, "", false, false, 0)
	Ok(t, err)

	// Verify the old key no longer exists
	_, err = client.Get(context.Background(), oldKey).Result()
	Assert(t, err != nil, "old key should be deleted after migration")

	// Verify the new key exists with correct format
	retrievedLock, err := r.GetLock(oldProject, "default")
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

	s := miniredis.RunT(t)
	r := newTestRedis(s)

	// Create a lock with the new format (includes project name)
	projectWithName := models.NewProject("owner/repo", "path", "projectName")
	newLock := models.ProjectLock{
		Pull:      models.PullRequest{Num: 1},
		User:      models.User{Username: "testuser"},
		Workspace: "default",
		Project:   projectWithName,
		Time:      time.Now(),
	}

	// Try to lock using the new format
	acquired, _, err := r.TryLock(newLock)
	Ok(t, err)
	Assert(t, acquired, "should acquire lock")

	// Verify the lock was created and can be retrieved with the correct key format
	retrievedLock, err := r.GetLock(projectWithName, "default")
	Ok(t, err)
	Assert(t, retrievedLock != nil, "lock should exist")
	Equals(t, "projectName", retrievedLock.Project.ProjectName)
	Equals(t, "testuser", retrievedLock.User.Username)

	// Close the current Redis connection and create a new one
	// This simulates a restart which would trigger the migration logic
	r.Close()
	r = newTestRedis(s)
	defer r.Close()
	// Verify lock still exists after "migration"
	retrievedLock, err = r.GetLock(projectWithName, "default")
	Ok(t, err)
	Assert(t, retrievedLock != nil, "lock should exist")
	Equals(t, "projectName", retrievedLock.Project.ProjectName)
	Equals(t, "testuser", retrievedLock.User.Username)

}

func TestMixedLocksPresent(t *testing.T) {
	s := miniredis.RunT(t)
	r := newTestRedis(s)
	timeNow := time.Now()
	_, err := r.LockCommand(command.Apply, timeNow)
	Ok(t, err)

	_, _, err = r.TryLock(lock)
	Ok(t, err)

	ls, err := r.List()
	Ok(t, err)
	Equals(t, 1, len(ls))
}

func TestListNoLocks(t *testing.T) {
	t.Log("listing locks when there are none should return an empty list")
	s := miniredis.RunT(t)
	r := newTestRedis(s)
	ls, err := r.List()
	Ok(t, err)
	Equals(t, 0, len(ls))
}

func TestListOneLock(t *testing.T) {
	t.Log("listing locks when there is one should return it")
	s := miniredis.RunT(t)
	r := newTestRedis(s)
	_, _, err := r.TryLock(lock)
	Ok(t, err)
	ls, err := r.List()
	Ok(t, err)
	Equals(t, 1, len(ls))
}

func TestListMultipleLocks(t *testing.T) {
	t.Log("listing locks when there are multiple should return them")
	s := miniredis.RunT(t)
	rdb := newTestRedis(s)

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
		_, _, err := rdb.TryLock(newLock)
		Ok(t, err)
	}
	ls, err := rdb.List()
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
	s := miniredis.RunT(t)
	rdb := newTestRedis(s)
	_, _, err := rdb.TryLock(lock)
	Ok(t, err)
	_, err = rdb.Unlock(project, workspace)
	Ok(t, err)

	ls, err := rdb.List()
	Ok(t, err)
	Equals(t, 0, len(ls))
}

func TestLockingNoLocks(t *testing.T) {
	t.Log("with no locks yet, lock should succeed")
	s := miniredis.RunT(t)
	rdb := newTestRedis(s)
	acquired, currLock, err := rdb.TryLock(lock)
	Ok(t, err)
	Equals(t, true, acquired)
	Equals(t, lock, currLock)
}

func TestLockingExistingLock(t *testing.T) {
	t.Log("if there is an existing lock, lock should...")
	s := miniredis.RunT(t)
	rdb := newTestRedis(s)
	_, _, err := rdb.TryLock(lock)
	Ok(t, err)

	t.Log("...succeed if the new project has a different path")
	{
		newLock := lock
		newLock.Project = models.NewProject(project.RepoFullName, "different/path", "")
		acquired, currLock, err := rdb.TryLock(newLock)
		Ok(t, err)
		Equals(t, true, acquired)
		Equals(t, pullNum, currLock.Pull.Num)
	}

	t.Log("...succeed if the new project has a different workspace")
	{
		newLock := lock
		newLock.Workspace = "different-workspace"
		acquired, currLock, err := rdb.TryLock(newLock)
		Ok(t, err)
		Equals(t, true, acquired)
		Equals(t, newLock, currLock)
	}

	t.Log("...succeed if the new project has a different repoName")
	{
		newLock := lock
		newLock.Project = models.NewProject("different/repo", project.Path, "")
		acquired, currLock, err := rdb.TryLock(newLock)
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
			acquired, currLock, err := rdb.TryLock(newLock)
			Ok(t, err)
			Equals(t, true, acquired)
			Equals(t, newLock, currLock)
		}
	*/

	t.Log("...not succeed if the new project only has a different pullNum")
	{
		newLock := lock
		newLock.Pull.Num = lock.Pull.Num + 1
		acquired, currLock, err := rdb.TryLock(newLock)
		Ok(t, err)
		Equals(t, false, acquired)
		Equals(t, currLock.Pull.Num, pullNum)
	}
}

func TestUnlockingNoLocks(t *testing.T) {
	t.Log("unlocking with no locks should succeed")
	s := miniredis.RunT(t)
	rdb := newTestRedis(s)
	_, err := rdb.Unlock(project, workspace)

	Ok(t, err)
}

func TestUnlocking(t *testing.T) {
	t.Log("unlocking with an existing lock should succeed")
	s := miniredis.RunT(t)
	rdb := newTestRedis(s)

	_, _, err := rdb.TryLock(lock)
	Ok(t, err)
	_, err = rdb.Unlock(project, workspace)
	Ok(t, err)

	// should be no locks listed
	ls, err := rdb.List()
	Ok(t, err)
	Equals(t, 0, len(ls))

	// should be able to re-lock that repo with a new pull num
	newLock := lock
	newLock.Pull.Num = lock.Pull.Num + 1
	acquired, currLock, err := rdb.TryLock(newLock)
	Ok(t, err)
	Equals(t, true, acquired)
	Equals(t, newLock, currLock)
}

func TestUnlockingMultiple(t *testing.T) {
	t.Log("unlocking and locking multiple locks should succeed")
	s := miniredis.RunT(t)
	rdb := newTestRedis(s)

	_, _, err := rdb.TryLock(lock)
	Ok(t, err)

	new1 := lock
	new1.Project.RepoFullName = "new/repo"
	_, _, err = rdb.TryLock(new1)
	Ok(t, err)

	new2 := lock
	new2.Project.Path = "new/path"
	_, _, err = rdb.TryLock(new2)
	Ok(t, err)

	new3 := lock
	new3.Workspace = "new-workspace"
	_, _, err = rdb.TryLock(new3)
	Ok(t, err)

	// now try and unlock them
	_, err = rdb.Unlock(new3.Project, new3.Workspace)
	Ok(t, err)
	_, err = rdb.Unlock(new2.Project, workspace)
	Ok(t, err)
	_, err = rdb.Unlock(new1.Project, workspace)
	Ok(t, err)
	_, err = rdb.Unlock(project, workspace)
	Ok(t, err)

	// should be none left
	ls, err := rdb.List()
	Ok(t, err)
	Equals(t, 0, len(ls))
}

func TestUnlockByPullNone(t *testing.T) {
	t.Log("UnlockByPull should be successful when there are no locks")
	s := miniredis.RunT(t)
	rdb := newTestRedis(s)

	_, err := rdb.UnlockByPull("any/repo", 1)
	Ok(t, err)
}

func TestUnlockByPullOne(t *testing.T) {
	t.Log("with one lock, UnlockByPull should...")
	s := miniredis.RunT(t)
	rdb := newTestRedis(s)
	_, _, err := rdb.TryLock(lock)
	Ok(t, err)

	t.Log("...delete nothing when its the same repo but a different pull")
	{
		_, err := rdb.UnlockByPull(project.RepoFullName, pullNum+1)
		Ok(t, err)
		ls, err := rdb.List()
		Ok(t, err)
		Equals(t, 1, len(ls))
	}
	t.Log("...delete nothing when its the same pull but a different repo")
	{
		_, err := rdb.UnlockByPull("different/repo", pullNum)
		Ok(t, err)
		ls, err := rdb.List()
		Ok(t, err)
		Equals(t, 1, len(ls))
	}
	t.Log("...delete the lock when its the same repo and pull")
	{
		_, err := rdb.UnlockByPull(project.RepoFullName, pullNum)
		Ok(t, err)
		ls, err := rdb.List()
		Ok(t, err)
		Equals(t, 0, len(ls))
	}
}

func TestUnlockByPullAfterUnlock(t *testing.T) {
	t.Log("after locking and unlocking, UnlockByPull should be successful")
	s := miniredis.RunT(t)
	rdb := newTestRedis(s)
	_, _, err := rdb.TryLock(lock)
	Ok(t, err)
	_, err = rdb.Unlock(project, workspace)
	Ok(t, err)

	_, err = rdb.UnlockByPull(project.RepoFullName, pullNum)
	Ok(t, err)
	ls, err := rdb.List()
	Ok(t, err)
	Equals(t, 0, len(ls))
}

func TestUnlockByPullMatching(t *testing.T) {
	t.Log("UnlockByPull should delete all locks in that repo and pull num")
	s := miniredis.RunT(t)
	rdb := newTestRedis(s)
	_, _, err := rdb.TryLock(lock)
	Ok(t, err)

	// add additional locks with the same repo and pull num but different paths/workspaces
	new1 := lock
	new1.Project.Path = "dif/path"
	_, _, err = rdb.TryLock(new1)
	Ok(t, err)
	new2 := lock
	new2.Workspace = "new-workspace"
	_, _, err = rdb.TryLock(new2)
	Ok(t, err)

	// there should now be 3
	ls, err := rdb.List()
	Ok(t, err)
	Equals(t, 3, len(ls))

	// should all be unlocked
	_, err = rdb.UnlockByPull(project.RepoFullName, pullNum)
	Ok(t, err)
	ls, err = rdb.List()
	Ok(t, err)
	Equals(t, 0, len(ls))
}

func TestGetLockNotThere(t *testing.T) {
	t.Log("getting a lock that doesn't exist should return a nil pointer")
	s := miniredis.RunT(t)
	rdb := newTestRedis(s)
	l, err := rdb.GetLock(project, workspace)
	Ok(t, err)
	Equals(t, (*models.ProjectLock)(nil), l)
}

func TestGetLock(t *testing.T) {
	t.Log("getting a lock should return the lock")
	s := miniredis.RunT(t)
	rdb := newTestRedis(s)
	_, _, err := rdb.TryLock(lock)
	Ok(t, err)

	l, err := rdb.GetLock(project, workspace)
	Ok(t, err)
	// can't compare against time so doing each field
	Equals(t, lock.Project, l.Project)
	Equals(t, lock.Workspace, l.Workspace)
	Equals(t, lock.Pull, l.Pull)
	Equals(t, lock.User, l.User)
}

// Test we can create a status and then getCommandLock it.
func TestPullStatus_UpdateGet(t *testing.T) {
	s := miniredis.RunT(t)
	rdb := newTestRedis(s)

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
	status, err := rdb.UpdatePullWithResults(
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

	maybeStatus, err := rdb.GetPullStatus(pull)
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
}

// Test we can create a status, delete it, and then we shouldn't be able to getCommandLock
// it.
func TestPullStatus_UpdateDeleteGet(t *testing.T) {
	s := miniredis.RunT(t)
	rdb := newTestRedis(s)

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
	_, err := rdb.UpdatePullWithResults(
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

	err = rdb.DeletePullStatus(pull)
	Ok(t, err)

	maybeStatus, err := rdb.GetPullStatus(pull)
	Ok(t, err)
	Assert(t, maybeStatus == nil, "exp nil")
}

// Test we can create a status, update a specific project's status within that
// pull status, and when we getCommandLock all the project statuses, that specific project
// should be updated.
func TestPullStatus_UpdateProject(t *testing.T) {
	s := miniredis.RunT(t)
	rdb := newTestRedis(s)

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
	_, err := rdb.UpdatePullWithResults(
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

	err = rdb.UpdateProjectStatus(pull, "default", ".", models.DiscardedPlanStatus)
	Ok(t, err)

	status, err := rdb.GetPullStatus(pull)
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
}

// Test that if we update an existing pull status and our new status is for a
// different HeadSHA, that we just overwrite the old status.
func TestPullStatus_UpdateNewCommit(t *testing.T) {
	s := miniredis.RunT(t)
	rdb := newTestRedis(s)

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
	_, err := rdb.UpdatePullWithResults(
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
	status, err := rdb.UpdatePullWithResults(pull,
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

	maybeStatus, err := rdb.GetPullStatus(pull)
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
}

// Test that if we update an existing pull status via Apply and our new status is for a
// the same commit, that we merge the statuses.
func TestPullStatus_UpdateMerge_Apply(t *testing.T) {
	s := miniredis.RunT(t)
	rdb := newTestRedis(s)

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
	_, err := rdb.UpdatePullWithResults(
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

	updateStatus, err := rdb.UpdatePullWithResults(pull,
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

	getStatus, err := rdb.GetPullStatus(pull)
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
}

// Test that if we update one existing policy status via approve_policies and our new status is for a
// the same commit, that we merge the statuses.
func TestPullStatus_UpdateMerge_ApprovePolicies(t *testing.T) {
	s := miniredis.RunT(t)
	rdb := newTestRedis(s)

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
	_, err := rdb.UpdatePullWithResults(
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

	updateStatus, err := rdb.UpdatePullWithResults(pull,
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

	getStatus, err := rdb.GetPullStatus(pull)
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
}

func newTestRedis(mr *miniredis.Miniredis) *redis.RedisDB {
	r, err := redis.New(mr.Host(), mr.Server().Addr().Port, "", false, false, 0)
	if err != nil {
		panic(fmt.Errorf("failed to create test redis client: %w", err))
	}
	return r
}

func newTestRedisTLS(mr *miniredis.Miniredis) *redis.RedisDB {
	r, err := redis.New(mr.Host(), mr.Server().Addr().Port, "", true, true, 0)
	if err != nil {
		panic(fmt.Errorf("failed to create test redis client: %w", err))
	}
	return r
}

func TestRedis_SaveAndGetProjectOutput(t *testing.T) {
	mr := miniredis.RunT(t)
	r := newTestRedis(mr)

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

	err := r.SaveProjectOutput(output)
	Ok(t, err)

	// Test GetProjectOutputHistory
	history, err := r.GetProjectOutputHistory("owner/repo", 123, "terraform/staging", "default", "myproject")
	Ok(t, err)
	Assert(t, len(history) == 1, "expected 1 result")
	Equals(t, output.Output, history[0].Output)
	Equals(t, output.Status, history[0].Status)

	// Test GetProjectOutputRun
	retrieved, err := r.GetProjectOutputRun("owner/repo", 123, "terraform/staging", "default", "myproject", "plan", now.UnixMilli())
	Ok(t, err)
	Assert(t, retrieved != nil, "expected non-nil result")
	Equals(t, output.Output, retrieved.Output)
}

func TestRedis_GetProjectOutputHistory_NotFound(t *testing.T) {
	mr := miniredis.RunT(t)
	r := newTestRedis(mr)

	history, err := r.GetProjectOutputHistory("owner/repo", 999, "nonexistent", "default", "")
	Ok(t, err)
	Assert(t, len(history) == 0, "expected empty history for non-existent output")
}

func TestRedis_GetProjectOutputsByPull(t *testing.T) {
	mr := miniredis.RunT(t)
	r := newTestRedis(mr)

	output1 := models.ProjectOutput{
		RepoFullName: "owner/repo",
		PullNum:      123,
		ProjectName:  "project1",
		Workspace:    "default",
		Path:         "terraform/staging",
	}
	output2 := models.ProjectOutput{
		RepoFullName: "owner/repo",
		PullNum:      123,
		ProjectName:  "project2",
		Workspace:    "default",
		Path:         "terraform/prod",
	}

	err := r.SaveProjectOutput(output1)
	Ok(t, err)
	err = r.SaveProjectOutput(output2)
	Ok(t, err)

	outputs, err := r.GetProjectOutputsByPull("owner/repo", 123)
	Ok(t, err)
	Equals(t, 2, len(outputs))
}

func TestRedis_GetProjectOutputsByPull_Empty(t *testing.T) {
	mr := miniredis.RunT(t)
	r := newTestRedis(mr)

	outputs, err := r.GetProjectOutputsByPull("owner/repo", 999)
	Ok(t, err)
	Equals(t, 0, len(outputs))
}

func TestRedis_DeleteProjectOutputsByPull(t *testing.T) {
	mr := miniredis.RunT(t)
	r := newTestRedis(mr)

	output := models.ProjectOutput{
		RepoFullName: "owner/repo",
		PullNum:      123,
		ProjectName:  "myproject",
		Workspace:    "default",
		Path:         "terraform/staging",
	}
	err := r.SaveProjectOutput(output)
	Ok(t, err)

	err = r.DeleteProjectOutputsByPull("owner/repo", 123)
	Ok(t, err)

	outputs, err := r.GetProjectOutputsByPull("owner/repo", 123)
	Ok(t, err)
	Equals(t, 0, len(outputs))
}

func TestRedis_GetActivePullRequests(t *testing.T) {
	mr := miniredis.RunT(t)
	r := newTestRedis(mr)

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

	err := r.SaveProjectOutput(output1)
	Ok(t, err)
	err = r.SaveProjectOutput(output2)
	Ok(t, err)

	pulls, err := r.GetActivePullRequests()
	Ok(t, err)
	Equals(t, 2, len(pulls))
}

func TestRedis_GetActivePullRequests_Empty(t *testing.T) {
	mr := miniredis.RunT(t)
	r := newTestRedis(mr)

	pulls, err := r.GetActivePullRequests()
	Ok(t, err)
	Equals(t, 0, len(pulls))
}

func TestRedisDB_SaveProjectOutput_Atomic(t *testing.T) {
	mr := miniredis.RunT(t)
	r := newTestRedis(mr)

	output := models.ProjectOutput{
		RepoFullName: "owner/repo",
		PullNum:      123,
		ProjectName:  "myproject",
		Workspace:    "default",
		Path:         "terraform/app",
		CommandName:  "plan",
		RunTimestamp: time.Now().UnixMilli(),
		Status:       models.SuccessOutputStatus,
	}

	// Save should succeed
	err := r.SaveProjectOutput(output)
	Ok(t, err)

	// Verify both the output and index were saved
	retrieved, err := r.GetProjectOutputsByPull("owner/repo", 123)
	Ok(t, err)
	Equals(t, 1, len(retrieved))
	Equals(t, output.ProjectName, retrieved[0].ProjectName)
}

func TestRedis_GetProjectOutputByJobID(t *testing.T) {
	mr := miniredis.RunT(t)
	r := newTestRedis(mr)

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

	err := r.SaveProjectOutput(output)
	Ok(t, err)

	// Test retrieval by job ID (should use index for O(1) lookup)
	retrieved, err := r.GetProjectOutputByJobID("job-abc-123")
	Ok(t, err)
	Assert(t, retrieved != nil, "expected non-nil result")
	Equals(t, output.JobID, retrieved.JobID)
	Equals(t, output.Output, retrieved.Output)
	Equals(t, output.RepoFullName, retrieved.RepoFullName)
}

func TestRedis_GetProjectOutputByJobID_NotFound(t *testing.T) {
	mr := miniredis.RunT(t)
	r := newTestRedis(mr)

	// Test retrieval of non-existent job ID
	retrieved, err := r.GetProjectOutputByJobID("nonexistent-job-id")
	Ok(t, err)
	Assert(t, retrieved == nil, "expected nil result for non-existent job ID")
}

func TestRedis_GetProjectOutputByJobID_EmptyJobID(t *testing.T) {
	mr := miniredis.RunT(t)
	r := newTestRedis(mr)

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

	err := r.SaveProjectOutput(output)
	Ok(t, err)

	// Empty job IDs are not indexed, but the fallback scan will still find them
	// This verifies backwards compatibility - outputs without job IDs can still be found
	retrieved, err := r.GetProjectOutputByJobID("")
	Ok(t, err)
	// The scan finds outputs with matching (empty) job ID
	Assert(t, retrieved != nil, "fallback scan should find output with empty job ID")
}

func generateLocalhostCert() ([]byte, []byte, error) {
	var err error

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	keyBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, keyBytes, err
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, keyBytes, err
	}

	notBefore := time.Now()
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Atlantis Test Suite"},
		},
		NotBefore: notBefore,
		NotAfter:  notBefore.Add(time.Hour),
		KeyUsage:  x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,

		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},

		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	return certBytes, keyBytes, err
}
