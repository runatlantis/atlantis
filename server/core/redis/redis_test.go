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
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	miniredisserver "github.com/alicebob/miniredis/v2/server"
	"github.com/pkg/errors"
	redisLib "github.com/redis/go-redis/v9"
	coredb "github.com/runatlantis/atlantis/server/core/db"
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

func TestNewWithConfig_SingleNode(t *testing.T) {
	t.Log("creating redis client with single-node mode")
	s := miniredis.RunT(t)
	r, err := redis.NewWithConfig(redis.Config{
		Hostname: s.Host(),
		Port:     s.Server().Addr().Port,
	})
	Ok(t, err)
	Assert(t, r != nil, "expected redis client to be created")
	r.Close()
}

func TestNewWithConfig_WithUsername(t *testing.T) {
	t.Log("creating redis client with username")
	s := miniredis.RunT(t)
	r, err := redis.NewWithConfig(redis.Config{
		Hostname: s.Host(),
		Port:     s.Server().Addr().Port,
		Username: "testuser",
	})
	Ok(t, err)
	Assert(t, r != nil, "expected redis client to be created")
	r.Close()
}

func TestNewWithConfig_ClusterEmptyAddresses(t *testing.T) {
	t.Log("cluster mode with empty addresses should fail")
	_, err := redis.NewWithConfig(redis.Config{
		ClusterAddresses: []string{"", ""},
	})
	Assert(t, err != nil, "expected error when cluster addresses are all empty")
	Assert(t, err.Error() == "redis cluster addresses provided but all are empty", "unexpected error: %v", err)
}

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

func TestUnlockIfOwnedByPullMissingLock(t *testing.T) {
	t.Log("UnlockIfOwnedByPull should ignore missing locks")
	s := miniredis.RunT(t)
	rdb := newTestRedis(s)

	deleted, err := rdb.UnlockIfOwnedByPull(project, workspace, pullNum)
	Ok(t, err)
	Equals(t, (*models.ProjectLock)(nil), deleted)
}

func TestUnlockIfOwnedByPullOtherPull(t *testing.T) {
	t.Log("UnlockIfOwnedByPull should not delete another pull's lock")
	s := miniredis.RunT(t)
	rdb := newTestRedis(s)

	_, _, err := rdb.TryLock(lock)
	Ok(t, err)

	deleted, err := rdb.UnlockIfOwnedByPull(project, workspace, pullNum+1)
	Ok(t, err)
	Equals(t, (*models.ProjectLock)(nil), deleted)

	existing, err := rdb.GetLock(project, workspace)
	Ok(t, err)
	Assert(t, existing != nil, "expected lock to remain")
	Equals(t, pullNum, existing.Pull.Num)
}

func TestUnlockIfOwnedByPullCurrentPull(t *testing.T) {
	t.Log("UnlockIfOwnedByPull should delete the current pull's lock")
	s := miniredis.RunT(t)
	rdb := newTestRedis(s)

	_, _, err := rdb.TryLock(lock)
	Ok(t, err)

	deleted, err := rdb.UnlockIfOwnedByPull(project, workspace, pullNum)
	Ok(t, err)
	Assert(t, deleted != nil, "expected deleted lock")
	Equals(t, lock.Project, deleted.Project)
	Equals(t, lock.Workspace, deleted.Workspace)
	Equals(t, lock.Pull, deleted.Pull)
	Equals(t, lock.User, deleted.User)

	existing, err := rdb.GetLock(project, workspace)
	Ok(t, err)
	Equals(t, (*models.ProjectLock)(nil), existing)
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

func TestPullStatus_UpdateSameCommitNewBaseBranch(t *testing.T) {
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
				RepoRelDir: "old-base",
				Workspace:  "default",
				ProjectCommandOutput: command.ProjectCommandOutput{
					PlanSuccess: &models.PlanSuccess{},
				},
			},
		})
	Ok(t, err)

	pull.BaseBranch = "release"
	status, err := rdb.UpdatePullWithResults(pull,
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

	maybeStatus, err := rdb.GetPullStatus(pull)
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
}

func TestRedis_SameCommitBackfillBaseDoesNotPromoteLegacyOldBaseProjects(t *testing.T) {
	s := miniredis.RunT(t)
	rdb := newTestRedis(s)

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
	_, err := rdb.UpdatePullWithResults(
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
	status, err := rdb.UpdatePullWithResults(pull,
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

	maybeStatus, err := rdb.GetPullStatus(pull)
	Ok(t, err)
	Equals(t, "main", maybeStatus.Pull.BaseBranch)
	Equals(t, status.Projects, maybeStatus.Projects)
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
}

// Test that policy approvals are preserved when HeadCommit changes,
// so sticky approvals can survive across code pushes.
func TestPullStatus_UpdateNewCommit_PreservesPolicyApprovals(t *testing.T) {
	s := miniredis.RunT(t)
	rdb := newTestRedis(s)

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
	_, err := rdb.UpdatePullWithResults(pull, []command.ProjectResult{
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
	pull.HeadCommit = "sha-B"
	status, err := rdb.UpdatePullWithResults(pull, []command.ProjectResult{
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
	getStatus, err := rdb.GetPullStatus(pull)
	Ok(t, err)
	Equals(t, 1, len(getStatus.Projects[0].PolicyStatus[0].Approvals))
}

// TestPullStatus_UpdateOverwritesCorruptData verifies that
// UpdatePullWithResults tolerates a pre-existing pull-status blob whose JSON
// no longer matches the current Go shape (e.g. after upgrading across a
// PullStatus schema change). The corrupt entry should be logged and
// overwritten rather than causing every subsequent plan to fail.
func TestPullStatus_UpdateOverwritesCorruptData(t *testing.T) {
	s := miniredis.RunT(t)
	rdb := newTestRedis(s)

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

	// Inject a corrupt pull-status blob simulating a legacy on-disk shape
	// that the current Go types cannot unmarshal (Approvals used to be int).
	key := fmt.Sprintf("%s::%s::%d",
		pull.BaseRepo.VCSHost.Hostname, pull.BaseRepo.FullName, pull.Num)
	corrupt := `{"Projects":[{"Workspace":"default","RepoRelDir":"mydir","ProjectName":"","PolicyStatus":[{"PolicySetName":"policy1","Passed":false,"Approvals":2}],"Status":0}],"Pull":{"Num":1}}`
	Ok(t, s.Set(key, corrupt))

	// Write fresh results. This must succeed despite the unreadable prior entry.
	status, err := rdb.UpdatePullWithResults(pull, []command.ProjectResult{
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
	got, err := rdb.GetPullStatus(pull)
	Ok(t, err)
	Assert(t, got != nil, "expected non-nil pull status")
	Equals(t, 1, len(got.Projects))
	Equals(t, models.PlannedPlanStatus, got.Projects[0].Status)
}

func TestBeginPlanGenerationOverwritesCorruptData(t *testing.T) {
	mr := miniredis.RunT(t)
	database := newTestRedis(mr)
	t.Cleanup(func() { database.Close() })
	pull := models.PullRequest{
		Num:        1,
		HeadCommit: "sha-A",
		BaseBranch: "main",
		BaseRepo: models.Repo{
			FullName: "runatlantis/atlantis",
			VCSHost:  models.VCSHost{Hostname: "github.com", Type: models.Github},
		},
	}
	key := fmt.Sprintf("%s::%s::%d", pull.BaseRepo.VCSHost.Hostname, pull.BaseRepo.FullName, pull.Num)
	Ok(t, mr.Set(key, `{"Projects":[{"PolicyStatus":[{"Approvals":2}]}]}`))

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
	server := miniredis.RunT(t)
	database := newTestRedis(server)
	t.Cleanup(func() { database.Close() })
	pull := planGenerationTestPull()
	selected := []models.ProjectStatus{{Workspace: "default", RepoRelDir: "project-a", ProjectName: "a"}}

	Ok(t, database.AcquirePlanPublicationClaim(pull, "owner-a"))
	Ok(t, database.AcquirePlanPublicationClaim(pull, "owner-a"))
	claimToken, err := database.GetPlanPublicationClaim(pull)
	Ok(t, err)
	Equals(t, "owner-a", claimToken)
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
	_, err = database.GetPlanPublicationClaim(pull)
	assertPlanPublicationClaimError(t, err, coredb.ErrPlanPublicationNotOwned)
	Ok(t, database.AcquirePlanPublicationClaim(pull, "owner-b"))
	err = database.ReleasePlanPublicationClaim(pull, "owner-a")
	assertPlanPublicationClaimError(t, err, coredb.ErrPlanPublicationNotOwned)
	err = database.AcquirePlanPublicationClaim(pull, "owner-c")
	assertPlanPublicationClaimError(t, err, coredb.ErrPlanPublicationBusy)
	Ok(t, database.ForceClearPlanPublicationClaim(pull))
	_, err = database.BeginPlanGeneration(pull, selected, "generation-b")
	Ok(t, err)
}

func TestPlanPublicationClaimSurvivesClientRestart(t *testing.T) {
	server := miniredis.RunT(t)
	pull := planGenerationTestPull()
	database := newTestRedis(server)
	Ok(t, database.AcquirePlanPublicationClaim(pull, "owner-a"))
	Ok(t, database.Close())

	database = newTestRedis(server)
	t.Cleanup(func() { database.Close() })
	err := database.AcquirePlanPublicationClaim(pull, "owner-b")
	assertPlanPublicationClaimError(t, err, coredb.ErrPlanPublicationBusy)
	Ok(t, database.ReleasePlanPublicationClaim(pull, "owner-a"))
}

func TestAcquirePlanPublicationClaimHandlesReleaseAtAtomicBoundary(t *testing.T) {
	server := miniredis.RunT(t)
	database := newTestRedis(server)
	t.Cleanup(func() { database.Close() })
	pull := planGenerationTestPull()
	Ok(t, database.AcquirePlanPublicationClaim(pull, "owner-a"))
	pullKey := fmt.Sprintf("%s::%s::%d", pull.BaseRepo.VCSHost.Hostname, pull.BaseRepo.FullName, pull.Num)
	claimKey := "plan-publication/{" + pullKey + "}"
	var released atomic.Bool
	server.Server().SetPreHook(func(peer *miniredisserver.Peer, cmd string, _ ...string) bool {
		if cmd != "EVAL" || !released.CompareAndSwap(false, true) {
			return false
		}
		claim, err := server.Get(claimKey)
		if err != nil || claim != "owner-a" {
			peer.WriteError(fmt.Sprintf("unexpected old claim %q: %v", claim, err))
			return true
		}
		server.Del(claimKey) // Previous owner releases immediately before atomic acquire.
		return false
	})

	Ok(t, database.AcquirePlanPublicationClaim(pull, "owner-b"))
	Assert(t, released.Load(), "expected release at atomic acquire boundary")
	err := database.AcquirePlanPublicationClaim(pull, "owner-c")
	assertPlanPublicationClaimError(t, err, coredb.ErrPlanPublicationBusy)
	Ok(t, database.ReleasePlanPublicationClaim(pull, "owner-b"))
}

func TestBeginPlanGenerationRejectsClaimTurnoverAtCAS(t *testing.T) {
	server := miniredis.RunT(t)
	database := newTestRedis(server)
	t.Cleanup(func() { database.Close() })
	pull := planGenerationTestPull()
	Ok(t, database.AcquirePlanPublicationClaim(pull, "owner-a"))
	turnedOver := installPlanPublicationClaimTurnoverBarrier(t, server, pull, "owner-a", "owner-b")

	_, err := database.BeginPlanGeneration(pull, []models.ProjectStatus{{
		Workspace: "default", RepoRelDir: "project-a", ProjectName: "a",
	}}, "generation-a", "owner-a")

	Assert(t, turnedOver.Load(), "expected claim turnover before begin CAS")
	assertPlanPublicationClaimError(t, err, coredb.ErrPlanPublicationNotOwned)
	status, err := database.GetPullStatus(pull)
	Ok(t, err)
	Assert(t, status == nil, "old owner must not create pull status after claim turnover")
	Ok(t, database.AcquirePlanPublicationClaim(pull, "owner-b"))
}

func TestCompletePlanGenerationRejectsClaimTurnoverAtCAS(t *testing.T) {
	server := miniredis.RunT(t)
	database := newTestRedis(server)
	t.Cleanup(func() { database.Close() })
	pull := planGenerationTestPull()
	project := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-a", ProjectName: "a"}
	_, err := database.BeginPlanGeneration(pull, []models.ProjectStatus{project}, "generation-a")
	Ok(t, err)
	Ok(t, database.AcquirePlanPublicationClaim(pull, "owner-a"))
	turnedOver := installPlanPublicationClaimTurnoverBarrier(t, server, pull, "owner-a", "owner-b")

	_, err = database.CompletePlanGeneration(pull, "generation-a", []command.ProjectResult{
		simplePlanGenerationResult("project-a", "a"),
	}, "owner-a")

	Assert(t, turnedOver.Load(), "expected claim turnover before completion CAS")
	assertPlanPublicationClaimError(t, err, coredb.ErrPlanPublicationNotOwned)
	status, err := database.GetPullStatus(pull)
	Ok(t, err)
	project = findPlanGenerationProject(t, *status, "project-a", "a")
	Equals(t, models.ErroredPlanStatus, project.Status)
	Equals(t, "generation-a", project.PlanGeneration)
}

func TestUpdatePolicyResultsRejectsClaimTurnoverAtCAS(t *testing.T) {
	server := miniredis.RunT(t)
	database := newTestRedis(server)
	t.Cleanup(func() { database.Close() })
	pull := planGenerationTestPull()
	project := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-a", ProjectName: "a"}
	_, err := database.BeginPlanGeneration(pull, []models.ProjectStatus{project}, "generation-a")
	Ok(t, err)
	_, err = database.CompletePlanGeneration(pull, "generation-a", []command.ProjectResult{
		simplePlanGenerationResult("project-a", "a"),
	})
	Ok(t, err)
	Ok(t, database.AcquirePlanPublicationClaim(pull, "owner-a"))
	turnedOver := installPlanPublicationClaimTurnoverBarrier(t, server, pull, "owner-a", "owner-b")

	_, err = database.UpdatePolicyResultsForPlanGeneration(pull, []command.ProjectResult{
		policyGenerationResult("project-a", "a", "generation-a", "policy-a"),
	}, "owner-a")

	Assert(t, turnedOver.Load(), "expected claim turnover before policy CAS")
	assertPlanPublicationClaimError(t, err, coredb.ErrPlanPublicationNotOwned)
	status, err := database.GetPullStatus(pull)
	Ok(t, err)
	project = findPlanGenerationProject(t, *status, "project-a", "a")
	Equals(t, models.PlannedPlanStatus, project.Status)
	Equals(t, 0, len(project.PolicyStatus))
}

func TestUpdateApplyResultsRejectsClaimTurnoverAtCAS(t *testing.T) {
	server := miniredis.RunT(t)
	database := newTestRedis(server)
	t.Cleanup(func() { database.Close() })
	pull := planGenerationTestPull()
	project := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-a", ProjectName: "a"}
	_, err := database.BeginPlanGeneration(pull, []models.ProjectStatus{project}, "generation-a")
	Ok(t, err)
	_, err = database.CompletePlanGeneration(pull, "generation-a", []command.ProjectResult{
		simplePlanGenerationResult("project-a", "a"),
	})
	Ok(t, err)
	Ok(t, database.AcquirePlanPublicationClaim(pull, "owner-a"))
	turnedOver := installPlanPublicationClaimTurnoverBarrier(t, server, pull, "owner-a", "owner-b")

	_, err = database.UpdateApplyResultsForPlanGeneration(pull, []command.ProjectResult{
		applyGenerationResult("project-a", "a", "generation-a", nil),
	}, "owner-a")

	Assert(t, turnedOver.Load(), "expected claim turnover before apply CAS")
	assertPlanPublicationClaimError(t, err, coredb.ErrPlanPublicationNotOwned)
	status, err := database.GetPullStatus(pull)
	Ok(t, err)
	project = findPlanGenerationProject(t, *status, "project-a", "a")
	Equals(t, models.PlannedPlanStatus, project.Status)
	Equals(t, "generation-a", project.AcceptedPlanGeneration)
}

func TestDeletePullStatusClearsPlanPublicationClaim(t *testing.T) {
	server := miniredis.RunT(t)
	database := newTestRedis(server)
	t.Cleanup(func() { database.Close() })
	pull := planGenerationTestPull()
	pullKey := fmt.Sprintf("%s::%s::%d", pull.BaseRepo.VCSHost.Hostname, pull.BaseRepo.FullName, pull.Num)
	claimKey := "plan-publication/{" + pullKey + "}"
	var claimHeldDuringPullDelete atomic.Bool
	Ok(t, database.AcquirePlanPublicationClaim(pull, "owner-a"))
	server.Server().SetPreHook(func(_ *miniredisserver.Peer, cmd string, args ...string) bool {
		if cmd == "DEL" && len(args) == 1 && args[0] == pullKey {
			claim, err := server.Get(claimKey)
			claimHeldDuringPullDelete.Store(err == nil && claim == "owner-a")
		}
		return false
	})
	Ok(t, database.DeletePullStatus(pull))
	Assert(t, claimHeldDuringPullDelete.Load(), "expected claim to fence transitions until pull deletion")
	Ok(t, database.AcquirePlanPublicationClaim(pull, "owner-b"))
}

func TestPlanPublicationClaimBackendErrorIsOrdinary(t *testing.T) {
	server := miniredis.RunT(t)
	database := newTestRedis(server)
	server.Close()
	err := database.AcquirePlanPublicationClaim(planGenerationTestPull(), "owner-a")
	Assert(t, err != nil, "expected closed backend error")
	Assert(t, !errors.Is(err, coredb.ErrPlanPublicationBusy), "backend error must not be busy")
	Assert(t, !errors.Is(err, coredb.ErrPlanPublicationNotOwned), "backend error must not be ownership")
	_, err = database.GetPlanPublicationClaim(planGenerationTestPull())
	Assert(t, err != nil, "expected closed backend error")
	Assert(t, !errors.Is(err, coredb.ErrPlanPublicationNotOwned), "backend error must not be ownership")
}

func TestBeginPlanGenerationReplacingDiscardsPriorProjects(t *testing.T) {
	server := miniredis.RunT(t)
	database := newTestRedis(server)
	t.Cleanup(func() { database.Close() })
	pull := planGenerationTestPull()
	projectA := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-a", ProjectName: "a"}
	projectB := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-b", ProjectName: "b"}
	policyStatus := []models.PolicySetStatus{{PolicySetName: "required-policy", Passed: true}}

	_, err := database.UpdatePullWithResults(pull, []command.ProjectResult{{
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
	server := miniredis.RunT(t)
	database := newTestRedis(server)
	t.Cleanup(func() { database.Close() })
	pull := planGenerationTestPull()
	projectX := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-x", ProjectName: "x"}
	projectY := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-y", ProjectName: "y"}
	_, err := database.BeginPlanGeneration(pull, []models.ProjectStatus{projectX}, "generation-active", "")
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
	server := miniredis.RunT(t)
	database := newTestRedis(server)
	t.Cleanup(func() { database.Close() })
	pull := planGenerationTestPull()
	key := fmt.Sprintf("%s::%s::%d", pull.BaseRepo.VCSHost.Hostname, pull.BaseRepo.FullName, pull.Num)
	Ok(t, server.Set(key, `{"Projects":[{"PolicyStatus":[{"Approvals":2}]}]}`))

	status, err := database.BeginPlanGenerationReplacing(pull, []models.ProjectStatus{{
		Workspace: "default", RepoRelDir: "project-a", ProjectName: "a",
	}}, "generation-1", "")

	Ok(t, err)
	Equals(t, 1, len(status.Projects))
	project := findPlanGenerationProject(t, status, "project-a", "a")
	Equals(t, models.ErroredPlanStatus, project.Status)
	Equals(t, "generation-1", project.PlanGeneration)
}

func TestBeginPlanGenerationRetriesWhenCorruptRawValueChangesBeforeCAS(t *testing.T) {
	mr := miniredis.RunT(t)
	database := newTestRedis(mr)
	t.Cleanup(func() { database.Close() })
	pull := models.PullRequest{
		Num:        1,
		HeadCommit: "sha-A",
		BaseBranch: "main",
		BaseRepo: models.Repo{
			FullName: "runatlantis/atlantis",
			VCSHost:  models.VCSHost{Hostname: "github.com", Type: models.Github},
		},
	}
	key := fmt.Sprintf("%s::%s::%d", pull.BaseRepo.VCSHost.Hostname, pull.BaseRepo.FullName, pull.Num)
	Ok(t, mr.Set(key, `{"Projects":[{"PolicyStatus":[{"Approvals":2}]}]}`))
	concurrentStatus := models.PullStatus{
		Pull: pull,
		Projects: []models.ProjectStatus{{
			Workspace:      "default",
			RepoRelDir:     "project-a",
			ProjectName:    "a",
			Status:         models.ErroredPlanStatus,
			PlanGeneration: "generation-old",
		}, {
			Workspace:      "default",
			RepoRelDir:     "project-b",
			ProjectName:    "b",
			Status:         models.ErroredPlanStatus,
			PlanGeneration: "generation-old",
		}},
	}
	serializedConcurrent, err := json.Marshal(concurrentStatus)
	Ok(t, err)
	var changed atomic.Bool
	mr.Server().SetPreHook(func(peer *miniredisserver.Peer, cmd string, _ ...string) bool {
		if cmd != "EVAL" || !changed.CompareAndSwap(false, true) {
			return false
		}
		if err := mr.Set(key, string(serializedConcurrent)); err != nil {
			peer.WriteError(err.Error())
			return true
		}
		return false
	})

	status, err := database.BeginPlanGeneration(pull, []models.ProjectStatus{{
		Workspace:   "default",
		RepoRelDir:  "project-a",
		ProjectName: "a",
	}}, "generation-1", "")

	Ok(t, err)
	Assert(t, changed.Load(), "expected the first CAS attempt to race with a raw-value change")
	Equals(t, []coredb.PlanGenerationProject{{Workspace: "default", RepoRelDir: "project-b", ProjectName: "b", CurrentGeneration: "generation-old", CurrentStatus: models.ErroredPlanStatus}}, status.Canceled)
	projectA := findPlanGenerationProject(t, status, "project-a", "a")
	Equals(t, models.ErroredPlanStatus, projectA.Status)
	Equals(t, "generation-1", projectA.PlanGeneration)
	projectB := findPlanGenerationProject(t, status, "project-b", "b")
	Equals(t, models.ErroredPlanStatus, projectB.Status)
	Equals(t, "", projectB.PlanGeneration)
}

func TestBeginPlanGenerationReplacingRetriesWhenRawValueChangesBeforeCAS(t *testing.T) {
	server := miniredis.RunT(t)
	database := newTestRedis(server)
	t.Cleanup(func() { database.Close() })
	pull := planGenerationTestPull()
	key := fmt.Sprintf("%s::%s::%d", pull.BaseRepo.VCSHost.Hostname, pull.BaseRepo.FullName, pull.Num)
	Ok(t, server.Set(key, `{"Projects":[{"PolicyStatus":[{"Approvals":2}]}]}`))
	policyStatus := []models.PolicySetStatus{{PolicySetName: "required-policy", Passed: true}}
	concurrentStatus := models.PullStatus{
		Pull: pull,
		Projects: []models.ProjectStatus{{
			Workspace: "default", RepoRelDir: "project-a", ProjectName: "a",
			Status: models.PlannedPlanStatus, ManagedPlanHash: "hash-a", PolicyStatus: policyStatus,
		}, {
			Workspace: "default", RepoRelDir: "project-b", ProjectName: "b",
			Status: models.PlannedPlanStatus, ManagedPlanHash: "hash-b", PolicyStatus: policyStatus,
		}},
	}
	serializedConcurrent, err := json.Marshal(concurrentStatus)
	Ok(t, err)
	var changed atomic.Bool
	server.Server().SetPreHook(func(peer *miniredisserver.Peer, cmd string, _ ...string) bool {
		if cmd != "EVAL" || !changed.CompareAndSwap(false, true) {
			return false
		}
		if err := server.Set(key, string(serializedConcurrent)); err != nil {
			peer.WriteError(err.Error())
			return true
		}
		return false
	})

	status, err := database.BeginPlanGenerationReplacing(pull, []models.ProjectStatus{{
		Workspace: "default", RepoRelDir: "project-a", ProjectName: "a",
	}}, "generation-1", "")

	Ok(t, err)
	Assert(t, changed.Load(), "expected the first CAS attempt to race with a raw-value change")
	selected := findPlanGenerationProject(t, status, "project-a", "a")
	Equals(t, models.ErroredPlanStatus, selected.Status)
	Equals(t, "generation-1", selected.PlanGeneration)
	Equals(t, "", selected.ManagedPlanHash)
	Equals(t, policyStatus, selected.PolicyStatus)
	unrelated := findPlanGenerationProject(t, status, "project-b", "b")
	Equals(t, models.DiscardedPlanStatus, unrelated.Status)
	Equals(t, "", unrelated.PlanGeneration)
	Equals(t, "", unrelated.ManagedPlanHash)
	Equals(t, policyStatus, unrelated.PolicyStatus)
}

func TestBeginPlanGenerationReturnsRedisErrors(t *testing.T) {
	mr := miniredis.RunT(t)
	database := newTestRedis(mr)
	t.Cleanup(func() { database.Close() })
	mr.SetError("LOADING backend unavailable")

	_, err := database.BeginPlanGeneration(models.PullRequest{
		Num: 1,
		BaseRepo: models.Repo{
			FullName: "runatlantis/atlantis",
			VCSHost:  models.VCSHost{Hostname: "github.com", Type: models.Github},
		},
	}, []models.ProjectStatus{{Workspace: "default", RepoRelDir: "project-a"}}, "generation-1", "")

	Assert(t, err != nil, "expected Redis backend error")
	Assert(t, strings.Contains(err.Error(), "backend unavailable"), "expected backend error to be retained, got %v", err)

	_, err = database.BeginPlanGenerationReplacing(models.PullRequest{
		Num: 1,
		BaseRepo: models.Repo{
			FullName: "runatlantis/atlantis",
			VCSHost:  models.VCSHost{Hostname: "github.com", Type: models.Github},
		},
	}, []models.ProjectStatus{{Workspace: "default", RepoRelDir: "project-a"}}, "generation-1", "")

	Assert(t, err != nil, "expected Redis backend error")
	Assert(t, strings.Contains(err.Error(), "backend unavailable"), "expected backend error to be retained, got %v", err)
}

func TestPlanGenerationInvalidationPreservesUnrelatedProjectsAndRejectsStaleCompletion(t *testing.T) {
	server := miniredis.RunT(t)
	database := newTestRedis(server)
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
	_, err := database.UpdatePullWithResults(pull, []command.ProjectResult{
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
	server := miniredis.RunT(t)
	database := newTestRedis(server)
	t.Cleanup(func() { database.Close() })
	pull := planGenerationTestPull()
	projectA := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-a", ProjectName: "a"}
	projectB := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-b", ProjectName: "b"}

	_, err := database.BeginPlanGeneration(pull, []models.ProjectStatus{projectA, projectB}, "generation-a", "")
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
		server := miniredis.RunT(t)
		database := newTestRedis(server)
		t.Cleanup(func() { database.Close() })
		pull := planGenerationTestPull()
		project := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-a", ProjectName: "a"}
		_, err := database.BeginPlanGeneration(pull, []models.ProjectStatus{project}, "generation-1", "")
		Ok(t, err)
		server.Close()
		_, err = database.CompletePlanGeneration(pull, "generation-1", []command.ProjectResult{managedPlanGenerationResult("project-a", "a", "hash")}, "")
		Assert(t, err != nil, "expected stopped redis completion to fail")
		var completionErr *coredb.PlanGenerationCompletionError
		Assert(t, !errors.As(err, &completionErr), "expected backend error to remain ordinary, got %v", err)
	})

	t.Run("missing", func(t *testing.T) {
		server := miniredis.RunT(t)
		database := newTestRedis(server)
		t.Cleanup(func() { database.Close() })
		_, err := database.CompletePlanGeneration(planGenerationTestPull(), "generation-1", nil, "")
		assertPlanGenerationError(t, err, coredb.ErrPlanGenerationStateInvalid)
	})

	t.Run("empty results without owner", func(t *testing.T) {
		server := miniredis.RunT(t)
		database := newTestRedis(server)
		t.Cleanup(func() { database.Close() })
		pull := planGenerationTestPull()
		_, err := database.UpdatePullWithResults(pull, []command.ProjectResult{{
			Command: command.Plan, Workspace: "default", RepoRelDir: "project-a", ProjectName: "a",
			ProjectCommandOutput: command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}},
		}})
		Ok(t, err)
		_, err = database.CompletePlanGeneration(pull, "generation-1", nil, "")
		assertPlanGenerationError(t, err, coredb.ErrPlanGenerationStateInvalid)
	})

	t.Run("corrupt", func(t *testing.T) {
		server := miniredis.RunT(t)
		database := newTestRedis(server)
		t.Cleanup(func() { database.Close() })
		pull := planGenerationTestPull()
		key := fmt.Sprintf("%s::%s::%d", pull.BaseRepo.VCSHost.Hostname, pull.BaseRepo.FullName, pull.Num)
		Ok(t, server.Set(key, `{"Projects":`))
		_, err := database.CompletePlanGeneration(pull, "generation-1", nil, "")
		assertPlanGenerationError(t, err, coredb.ErrPlanGenerationStateInvalid)
	})

	t.Run("pull identity changed", func(t *testing.T) {
		server := miniredis.RunT(t)
		database := newTestRedis(server)
		t.Cleanup(func() { database.Close() })
		pull := planGenerationTestPull()
		project := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-a", ProjectName: "a"}
		_, err := database.BeginPlanGeneration(pull, []models.ProjectStatus{project}, "generation-1", "")
		Ok(t, err)
		changedPull := pull
		changedPull.HeadCommit = "new-head"
		_, err = database.CompletePlanGeneration(changedPull, "generation-1", []command.ProjectResult{managedPlanGenerationResult("project-a", "a", "hash")}, "")
		assertPlanGenerationError(t, err, coredb.ErrPlanGenerationPullChanged)
	})

	t.Run("incomplete", func(t *testing.T) {
		server := miniredis.RunT(t)
		database := newTestRedis(server)
		t.Cleanup(func() { database.Close() })
		pull := planGenerationTestPull()
		projects := []models.ProjectStatus{
			{Workspace: "default", RepoRelDir: "project-a", ProjectName: "a"},
			{Workspace: "default", RepoRelDir: "project-b", ProjectName: "b"},
		}
		_, err := database.BeginPlanGeneration(pull, projects, "generation-1", "")
		Ok(t, err)
		_, err = database.CompletePlanGeneration(pull, "generation-1", []command.ProjectResult{managedPlanGenerationResult("project-a", "a", "hash")}, "")
		assertPlanGenerationError(t, err, coredb.ErrPlanGenerationIncomplete)
	})

	t.Run("managed hash missing", func(t *testing.T) {
		server := miniredis.RunT(t)
		database := newTestRedis(server)
		t.Cleanup(func() { database.Close() })
		pull := planGenerationTestPull()
		project := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-a", ProjectName: "a"}
		_, err := database.BeginPlanGeneration(pull, []models.ProjectStatus{project}, "generation-1", "")
		Ok(t, err)
		result := managedPlanGenerationResult("project-a", "a", "")
		_, err = database.CompletePlanGeneration(pull, "generation-1", []command.ProjectResult{result}, "")
		assertPlanGenerationError(t, err, coredb.ErrPlanGenerationStateInvalid)
	})
}

func TestUpdatePullWithResultsRejectsActivePlanGeneration(t *testing.T) {
	mr := miniredis.RunT(t)
	database := newTestRedis(mr)
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
	_, err := database.UpdatePullWithResults(pull, []command.ProjectResult{
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
	mr := miniredis.RunT(t)
	database := newTestRedis(mr)
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
	_, err := database.UpdatePullWithResults(pull, []command.ProjectResult{
		managedPlanGenerationResult("project-a", "a", "hash-a"),
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

	database = newTestRedis(mr)
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
	database = newTestRedis(mr)
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
	mr := miniredis.RunT(t)
	database := newTestRedis(mr)
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
	_, err := database.BeginPlanGeneration(pull, []models.ProjectStatus{{
		Workspace: "default", RepoRelDir: "project-a", ProjectName: "a",
	}}, "generation-1", "")

	Ok(t, err)

	key := fmt.Sprintf("%s::%s::%d", pull.BaseRepo.VCSHost.Hostname, pull.BaseRepo.FullName, pull.Num)
	serialized, err := mr.Get(key)
	Ok(t, err)
	var status models.PullStatus
	Ok(t, json.Unmarshal([]byte(serialized), &status))
	status.Projects[0].Status = models.PlannedPlanStatus
	serializedStatus, err := json.Marshal(status)
	Ok(t, err)
	Ok(t, mr.Set(key, string(serializedStatus)))

	_, err = database.CompletePlanGeneration(pull, "generation-1", []command.ProjectResult{{
		Command: command.Plan, Workspace: "default", RepoRelDir: "project-a", ProjectName: "a",
		ProjectCommandOutput: command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}},
	}}, "")

	assertPlanGenerationError(t, err, coredb.ErrPlanGenerationStateInvalid)
	got, err := database.GetPullStatus(pull)
	Ok(t, err)
	Equals(t, models.PlannedPlanStatus, got.Projects[0].Status)
	Equals(t, "generation-1", got.Projects[0].PlanGeneration)
}

func TestUpdatePolicyResultsForPlanGenerationIsGenerationFenced(t *testing.T) {
	server := miniredis.RunT(t)
	database := newTestRedis(server)
	t.Cleanup(func() { database.Close() })
	pull := planGenerationTestPull()
	projectA := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-a", ProjectName: "a"}
	projectB := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-b", ProjectName: "b"}
	_, err := database.BeginPlanGeneration(pull, []models.ProjectStatus{projectA}, "generation-a", "")
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

func TestUpdatePolicyResultsForPlanGenerationRetriesCAS(t *testing.T) {
	server := miniredis.RunT(t)
	database := newTestRedis(server)
	t.Cleanup(func() { database.Close() })
	pull := planGenerationTestPull()
	project := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-a", ProjectName: "a"}
	_, err := database.BeginPlanGeneration(pull, []models.ProjectStatus{project}, "generation-1", "")
	Ok(t, err)
	_, err = database.CompletePlanGeneration(pull, "generation-1", []command.ProjectResult{
		managedPlanGenerationResult("project-a", "a", "hash-a"),
	}, "")

	Ok(t, err)
	key := fmt.Sprintf("%s::%s::%d", pull.BaseRepo.VCSHost.Hostname, pull.BaseRepo.FullName, pull.Num)
	serialized, err := server.Get(key)
	Ok(t, err)
	var concurrentStatus models.PullStatus
	Ok(t, json.Unmarshal([]byte(serialized), &concurrentStatus))
	concurrentStatus.Projects[0].AcceptedPlanGeneration = "generation-2"
	serializedConcurrent, err := json.Marshal(concurrentStatus)
	Ok(t, err)
	var changed atomic.Bool
	server.Server().SetPreHook(func(peer *miniredisserver.Peer, cmd string, _ ...string) bool {
		if cmd != "EVAL" || !changed.CompareAndSwap(false, true) {
			return false
		}
		if err := server.Set(key, string(serializedConcurrent)); err != nil {
			peer.WriteError(err.Error())
			return true
		}
		return false
	})

	_, err = database.UpdatePolicyResultsForPlanGeneration(pull, []command.ProjectResult{
		policyGenerationResult("project-a", "a", "generation-1", "policy-a"),
	}, "")

	Assert(t, changed.Load(), "expected the first CAS attempt to race with a raw-value change")
	assertPlanGenerationError(t, err, coredb.ErrPlanGenerationSuperseded)
	got, err := database.GetPullStatus(pull)
	Ok(t, err)
	Equals(t, "generation-2", got.Projects[0].AcceptedPlanGeneration)
	Equals(t, 0, len(got.Projects[0].PolicyStatus))
}

func TestUpdatePolicyResultsForPlanGenerationReturnsRedisErrors(t *testing.T) {
	server := miniredis.RunT(t)
	database := newTestRedis(server)
	t.Cleanup(func() { database.Close() })
	server.SetError("LOADING backend unavailable")
	_, err := database.UpdatePolicyResultsForPlanGeneration(planGenerationTestPull(), []command.ProjectResult{
		policyGenerationResult("project-a", "a", "generation-1", "policy-a"),
	}, "")

	Assert(t, err != nil, "expected Redis backend error")
	Assert(t, strings.Contains(err.Error(), "backend unavailable"), "expected backend error to be retained, got %v", err)
	var completionErr *coredb.PlanGenerationCompletionError
	Assert(t, !errors.As(err, &completionErr), "expected backend error to remain ordinary, got %v", err)
}

func TestUpdateApplyResultsForPlanGenerationIsGenerationFenced(t *testing.T) {
	server := miniredis.RunT(t)
	database := newTestRedis(server)
	t.Cleanup(func() { database.Close() })
	pull := planGenerationTestPull()
	projectA := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-a", ProjectName: "a"}
	projectB := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-b", ProjectName: "b"}
	_, err := database.BeginPlanGeneration(pull, []models.ProjectStatus{projectA}, "generation-a", "")
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

func TestUpdateApplyResultsForPlanGenerationRetriesCASAfterNewerCompletion(t *testing.T) {
	server := miniredis.RunT(t)
	database := newTestRedis(server)
	t.Cleanup(func() { database.Close() })
	pull := planGenerationTestPull()
	project := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-a", ProjectName: "a"}
	_, err := database.BeginPlanGeneration(pull, []models.ProjectStatus{project}, "generation-a", "")
	Ok(t, err)
	_, err = database.CompletePlanGeneration(pull, "generation-a", []command.ProjectResult{
		managedPlanGenerationResult("project-a", "a", "hash-a"),
	}, "")

	Ok(t, err)
	key := fmt.Sprintf("%s::%s::%d", pull.BaseRepo.VCSHost.Hostname, pull.BaseRepo.FullName, pull.Num)
	serialized, err := server.Get(key)
	Ok(t, err)
	var newerStatus models.PullStatus
	Ok(t, json.Unmarshal([]byte(serialized), &newerStatus))
	newerStatus.Projects[0].Status = models.PlannedPlanStatus
	newerStatus.Projects[0].ManagedPlanHash = "hash-b"
	newerStatus.Projects[0].AcceptedPlanGeneration = "generation-b"
	serializedNewer, err := json.Marshal(newerStatus)
	Ok(t, err)
	var completedNewer atomic.Bool
	server.Server().SetPreHook(func(peer *miniredisserver.Peer, cmd string, _ ...string) bool {
		if cmd != "EVAL" || !completedNewer.CompareAndSwap(false, true) {
			return false
		}
		if err := server.Set(key, string(serializedNewer)); err != nil {
			peer.WriteError(err.Error())
			return true
		}
		return false
	})

	_, err = database.UpdateApplyResultsForPlanGeneration(pull, []command.ProjectResult{
		applyGenerationResult("project-a", "a", "generation-a", nil),
	}, "")

	Assert(t, completedNewer.Load(), "expected newer generation completion before apply CAS")
	assertPlanGenerationError(t, err, coredb.ErrPlanGenerationSuperseded)
	got, err := database.GetPullStatus(pull)
	Ok(t, err)
	Equals(t, models.PlannedPlanStatus, got.Projects[0].Status)
	Equals(t, "hash-b", got.Projects[0].ManagedPlanHash)
	Equals(t, "generation-b", got.Projects[0].AcceptedPlanGeneration)
}

func TestUpdateApplyResultsForPlanGenerationReturnsRedisErrors(t *testing.T) {
	server := miniredis.RunT(t)
	database := newTestRedis(server)
	t.Cleanup(func() { database.Close() })
	server.SetError("LOADING backend unavailable")
	_, err := database.UpdateApplyResultsForPlanGeneration(planGenerationTestPull(), []command.ProjectResult{
		applyGenerationResult("project-a", "a", "generation-a", nil),
	}, "")

	Assert(t, err != nil, "expected Redis backend error")
	Assert(t, strings.Contains(err.Error(), "backend unavailable"), "expected backend error to be retained, got %v", err)
	var completionErr *coredb.PlanGenerationCompletionError
	Assert(t, !errors.As(err, &completionErr), "expected backend error to remain ordinary, got %v", err)
}

func TestUpdateDiscardResultsForPlanGenerationIsGenerationFenced(t *testing.T) {
	server := miniredis.RunT(t)
	database := newTestRedis(server)
	t.Cleanup(func() { database.Close() })
	pull := planGenerationTestPull()
	projectA := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-a", ProjectName: "a"}
	projectB := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-b", ProjectName: "b"}
	_, err := database.BeginPlanGeneration(pull, []models.ProjectStatus{projectA, projectB}, "generation-a")
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
	server := miniredis.RunT(t)
	database := newTestRedis(server)
	t.Cleanup(func() { database.Close() })
	pull := planGenerationTestPull()
	_, err := database.UpdatePullWithResults(pull, []command.ProjectResult{
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
	server := miniredis.RunT(t)
	database := newTestRedis(server)
	t.Cleanup(func() { database.Close() })
	pull := planGenerationTestPull()
	projectA := models.ProjectStatus{Workspace: "default", RepoRelDir: "shared-dir", ProjectName: "a"}
	projectB := models.ProjectStatus{Workspace: "default", RepoRelDir: "shared-dir", ProjectName: "b"}
	_, err := database.BeginPlanGeneration(pull, []models.ProjectStatus{projectA}, "generation-a")
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

func TestUpdateDiscardResultsForPlanGenerationRejectsClaimTurnoverAtCAS(t *testing.T) {
	server := miniredis.RunT(t)
	database := newTestRedis(server)
	t.Cleanup(func() { database.Close() })
	pull := planGenerationTestPull()
	project := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-a", ProjectName: "a"}
	_, err := database.BeginPlanGeneration(pull, []models.ProjectStatus{project}, "generation-a")
	Ok(t, err)
	_, err = database.CompletePlanGeneration(pull, "generation-a", []command.ProjectResult{
		managedPlanGenerationResult("project-a", "a", "hash-a"),
	})
	Ok(t, err)
	Ok(t, database.AcquirePlanPublicationClaim(pull, "owner-a"))
	turnedOver := installPlanPublicationClaimTurnoverBarrier(t, server, pull, "owner-a", "owner-b")

	_, err = database.UpdateDiscardResultsForPlanGeneration(pull, []command.ProjectResult{
		discardGenerationResult(command.Import, "project-a", "a", "generation-a"),
	}, "owner-a")

	Assert(t, turnedOver.Load(), "expected claim turnover before discard CAS")
	assertPlanPublicationClaimError(t, err, coredb.ErrPlanPublicationNotOwned)
	unchanged, err := database.GetPullStatus(pull)
	Ok(t, err)
	project = findPlanGenerationProject(t, *unchanged, "project-a", "a")
	Equals(t, models.PlannedPlanStatus, project.Status)
	Equals(t, "hash-a", project.ManagedPlanHash)
	Equals(t, "generation-a", project.AcceptedPlanGeneration)
}

func TestUpdateDiscardResultsForPlanGenerationRetriesCASAfterNewerCompletion(t *testing.T) {
	server := miniredis.RunT(t)
	database := newTestRedis(server)
	t.Cleanup(func() { database.Close() })
	pull := planGenerationTestPull()
	project := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-a", ProjectName: "a"}
	_, err := database.BeginPlanGeneration(pull, []models.ProjectStatus{project}, "generation-a")
	Ok(t, err)
	_, err = database.CompletePlanGeneration(pull, "generation-a", []command.ProjectResult{
		managedPlanGenerationResult("project-a", "a", "hash-a"),
	})
	Ok(t, err)
	Ok(t, database.AcquirePlanPublicationClaim(pull, "owner-a"))
	defer func() { database.ForceClearPlanPublicationClaim(pull) }() //nolint:errcheck

	key := fmt.Sprintf("%s::%s::%d", pull.BaseRepo.VCSHost.Hostname, pull.BaseRepo.FullName, pull.Num)
	serialized, err := server.Get(key)
	Ok(t, err)
	var newerStatus models.PullStatus
	Ok(t, json.Unmarshal([]byte(serialized), &newerStatus))
	newerStatus.Projects[0].ManagedPlanHash = "hash-b"
	newerStatus.Projects[0].AcceptedPlanGeneration = "generation-b"
	serializedNewer, err := json.Marshal(newerStatus)
	Ok(t, err)
	var completedNewer atomic.Bool
	server.Server().SetPreHook(func(peer *miniredisserver.Peer, cmd string, _ ...string) bool {
		if cmd != "EVAL" || !completedNewer.CompareAndSwap(false, true) {
			return false
		}
		if err := server.Set(key, string(serializedNewer)); err != nil {
			peer.WriteError(err.Error())
			return true
		}
		return false
	})

	_, err = database.UpdateDiscardResultsForPlanGeneration(pull, []command.ProjectResult{
		discardGenerationResult(command.Import, "project-a", "a", "generation-a"),
	}, "owner-a")

	Assert(t, completedNewer.Load(), "expected newer completion before discard CAS")
	assertPlanGenerationError(t, err, coredb.ErrPlanGenerationSuperseded)
	unchanged, err := database.GetPullStatus(pull)
	Ok(t, err)
	project = findPlanGenerationProject(t, *unchanged, "project-a", "a")
	Equals(t, models.PlannedPlanStatus, project.Status)
	Equals(t, "hash-b", project.ManagedPlanHash)
	Equals(t, "generation-b", project.AcceptedPlanGeneration)
}

func TestUpdateDiscardResultsForPlanGenerationReturnsRedisErrors(t *testing.T) {
	server := miniredis.RunT(t)
	database := newTestRedis(server)
	t.Cleanup(func() { database.Close() })
	server.SetError("LOADING backend unavailable")
	_, err := database.UpdateDiscardResultsForPlanGeneration(planGenerationTestPull(), []command.ProjectResult{
		discardGenerationResult(command.Import, "project-a", "a", "generation-a"),
	})
	Assert(t, err != nil, "expected Redis backend error")
	Assert(t, strings.Contains(err.Error(), "backend unavailable"), "expected backend error to be retained, got %v", err)
	Assert(t, !errors.Is(err, coredb.ErrPlanGenerationSuperseded), "backend error must remain ordinary")
	Assert(t, !errors.Is(err, coredb.ErrPlanPublicationNotOwned), "backend error must remain ordinary")
}

func TestManagedPlanHashLifecycle(t *testing.T) {
	server := miniredis.RunT(t)
	database := newTestRedis(server)
	pull := planGenerationTestPull()
	project := models.ProjectStatus{Workspace: "default", RepoRelDir: "project-a", ProjectName: "a"}

	_, err := database.BeginPlanGeneration(pull, []models.ProjectStatus{project}, "generation-1", "")
	Ok(t, err)
	status, err := database.CompletePlanGeneration(pull, "generation-1", []command.ProjectResult{
		managedPlanGenerationResult("project-a", "a", "hash-1"),
	}, "")

	Ok(t, err)
	Equals(t, "hash-1", findPlanGenerationProject(t, status, "project-a", "a").ManagedPlanHash)
	Equals(t, "generation-1", findPlanGenerationProject(t, status, "project-a", "a").AcceptedPlanGeneration)
	Ok(t, database.Close())

	database = newTestRedis(server)
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

func installPlanPublicationClaimTurnoverBarrier(t *testing.T, server *miniredis.Miniredis, pull models.PullRequest, oldToken, newToken string) *atomic.Bool {
	t.Helper()
	pullKey := fmt.Sprintf("%s::%s::%d", pull.BaseRepo.VCSHost.Hostname, pull.BaseRepo.FullName, pull.Num)
	claimKey := "plan-publication/{" + pullKey + "}"
	var turnedOver atomic.Bool
	server.Server().SetPreHook(func(peer *miniredisserver.Peer, cmd string, _ ...string) bool {
		if cmd != "EVAL" || !turnedOver.CompareAndSwap(false, true) {
			return false
		}
		claim, err := server.Get(claimKey)
		if err != nil || claim != oldToken {
			peer.WriteError(fmt.Sprintf("unexpected old claim %q: %v", claim, err))
			return true
		}
		server.Del(claimKey)                                   // ForceClearPlanPublicationClaim.
		if err := server.Set(claimKey, newToken); err != nil { // AcquirePlanPublicationClaim by the new owner.
			peer.WriteError(err.Error())
			return true
		}
		return false
	})
	return &turnedOver
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
