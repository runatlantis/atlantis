package redis_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/redis"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"

	. "github.com/runatlantis/atlantis/testing"
)

var project = models.NewProject("owner/repo", "parent/child")
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

func TestGetQueueByLock(t *testing.T) {
	t.Log("Getting Queue By Lock")
	s := miniredis.RunT(t)
	r := newTestRedisQueue(s)

	// queue doesn't exist -> should return nil
	queue, err := r.GetQueueByLock(lock.Project, lock.Workspace)
	Ok(t, err)
	Assert(t, queue == nil, "exp nil")

	_, _, _, err = r.TryLock(lock)
	Ok(t, err)

	lock1 := lock
	lock1.Pull.Num = 2
	_, _, _, err = r.TryLock(lock1) // this lock should be queued
	Ok(t, err)

	lock2 := lock
	lock2.Pull.Num = 3
	_, _, _, err = r.TryLock(lock2) // this lock should be queued
	Ok(t, err)

	queue, _ = r.GetQueueByLock(lock.Project, lock.Workspace)
	Equals(t, 2, len(queue))
}

func TestSingleQueue(t *testing.T) {
	t.Log("locking should return correct EnqueueStatus for a single queue")
	s := miniredis.RunT(t)
	r := newTestRedisQueue(s)

	lockAcquired, _, _, err := r.TryLock(lock)
	Ok(t, err)
	Equals(t, true, lockAcquired)

	secondLock := lock
	secondLock.Pull.Num = pullNum + 1
	lockAcquired, _, enqueueStatus, err := r.TryLock(secondLock)
	Ok(t, err)
	Equals(t, false, lockAcquired)
	Equals(t, models.Enqueued, enqueueStatus.Status)
	Equals(t, 1, enqueueStatus.QueueDepth)

	lockAcquired, _, enqueueStatus, err = r.TryLock(secondLock)
	Ok(t, err)
	Equals(t, false, lockAcquired)
	Equals(t, models.AlreadyInTheQueue, enqueueStatus.Status)
	Equals(t, 1, enqueueStatus.QueueDepth)

	thirdLock := lock
	thirdLock.Pull.Num = pullNum + 2
	lockAcquired, _, enqueueStatus, err = r.TryLock(thirdLock)
	Ok(t, err)
	Equals(t, false, lockAcquired)
	Equals(t, models.Enqueued, enqueueStatus.Status)
	Equals(t, 2, enqueueStatus.QueueDepth)
}

func TestMultipleQueues(t *testing.T) {
	t.Log("locking should return correct EnqueueStatus for multiple queues")
	s := miniredis.RunT(t)
	r := newTestRedisQueue(s)

	lockAcquired, _, _, err := r.TryLock(lock)
	Ok(t, err)
	Equals(t, true, lockAcquired)

	lockInDifferentWorkspace := lock
	lockInDifferentWorkspace.Workspace = "different-workspace"
	lockAcquired, _, _, err = r.TryLock(lockInDifferentWorkspace)
	Ok(t, err)
	Equals(t, true, lockAcquired)

	secondLock := lock
	secondLock.Pull.Num = pullNum + 1
	lockAcquired, _, enqueueStatus, err := r.TryLock(secondLock)
	Ok(t, err)
	Equals(t, false, lockAcquired)
	Equals(t, 1, enqueueStatus.QueueDepth)
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

func TestMixedLocksPresent(t *testing.T) {
	s := miniredis.RunT(t)
	r := newTestRedis(s)
	timeNow := time.Now()
	_, err := r.LockCommand(command.Apply, timeNow)
	Ok(t, err)

	_, _, _, err = r.TryLock(lock)
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
	_, _, _, err := r.TryLock(lock)
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
		newLock.Project = models.NewProject(r, "path")
		_, _, _, err := rdb.TryLock(newLock)
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
	_, _, _, err := rdb.TryLock(lock)
	Ok(t, err)
	_, _, err = rdb.Unlock(project, workspace)
	Ok(t, err)

	ls, err := rdb.List()
	Ok(t, err)
	Equals(t, 0, len(ls))
}

func TestLockingNoLocks(t *testing.T) {
	t.Log("with no locks yet, lock should succeed")
	s := miniredis.RunT(t)
	rdb := newTestRedis(s)
	acquired, currLock, _, err := rdb.TryLock(lock)
	Ok(t, err)
	Equals(t, true, acquired)
	Equals(t, lock, currLock)
}

func TestLockingExistingLock(t *testing.T) {
	t.Log("if there is an existing lock, lock should...")
	s := miniredis.RunT(t)
	rdb := newTestRedisQueue(s)
	_, _, _, err := rdb.TryLock(lock)
	Ok(t, err)

	t.Log("...succeed if the new project has a different path")
	{
		newLock := lock
		newLock.Project = models.NewProject(project.RepoFullName, "different/path")
		acquired, currLock, _, err := rdb.TryLock(newLock)
		Ok(t, err)
		Equals(t, true, acquired)
		Equals(t, pullNum, currLock.Pull.Num)
	}

	t.Log("...succeed if the new project has a different workspace")
	{
		newLock := lock
		newLock.Workspace = "different-workspace"
		acquired, currLock, _, err := rdb.TryLock(newLock)
		Ok(t, err)
		Equals(t, true, acquired)
		Equals(t, newLock, currLock)
	}

	t.Log("...succeed if the new project has a different repoName")
	{
		newLock := lock
		newLock.Project = models.NewProject("different/repo", project.Path)
		acquired, currLock, _, err := rdb.TryLock(newLock)
		Ok(t, err)
		Equals(t, true, acquired)
		Equals(t, newLock, currLock)
	}

	t.Log("...succeed if the new project has a different pullNum, the locking attempt will be queued")
	{
		newLock := lock
		newLock.Pull.Num = lock.Pull.Num + 1
		acquired, _, enqueueStatus, err := rdb.TryLock(newLock)
		Ok(t, err)
		Equals(t, false, acquired)
		Equals(t, 1, enqueueStatus.QueueDepth)
	}
}

func TestUnlockingNoLocks(t *testing.T) {
	t.Log("unlocking with no locks should succeed")
	s := miniredis.RunT(t)
	rdb := newTestRedis(s)
	_, _, err := rdb.Unlock(project, workspace)

	Ok(t, err)
}

func TestUnlocking(t *testing.T) {
	t.Log("unlocking with an existing lock should succeed")
	s := miniredis.RunT(t)
	rdb := newTestRedis(s)

	_, _, _, err := rdb.TryLock(lock)
	Ok(t, err)
	_, _, err = rdb.Unlock(project, workspace)
	Ok(t, err)

	// should be no locks listed
	ls, err := rdb.List()
	Ok(t, err)
	Equals(t, 0, len(ls))

	// should be able to re-lock that repo with a new pull num
	newLock := lock
	newLock.Pull.Num = lock.Pull.Num + 1
	acquired, currLock, _, err := rdb.TryLock(newLock)
	Ok(t, err)
	Equals(t, true, acquired)
	Equals(t, newLock, currLock)
}

func TestUnlockingMultiple(t *testing.T) {
	t.Log("unlocking and locking multiple locks should succeed")
	s := miniredis.RunT(t)
	rdb := newTestRedis(s)

	_, _, _, err := rdb.TryLock(lock)
	Ok(t, err)

	new := lock
	new.Project.RepoFullName = "new/repo"
	_, _, _, err = rdb.TryLock(new)
	Ok(t, err)

	new2 := lock
	new2.Project.Path = "new/path"
	_, _, _, err = rdb.TryLock(new2)
	Ok(t, err)

	new3 := lock
	new3.Workspace = "new-workspace"
	_, _, _, err = rdb.TryLock(new3)
	Ok(t, err)

	// now try and unlock them
	_, _, err = rdb.Unlock(new3.Project, new3.Workspace)
	Ok(t, err)
	_, _, err = rdb.Unlock(new2.Project, workspace)
	Ok(t, err)
	_, _, err = rdb.Unlock(new.Project, workspace)
	Ok(t, err)
	_, _, err = rdb.Unlock(project, workspace)
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

	_, _, err := rdb.UnlockByPull("any/repo", 1)
	Ok(t, err)
}

func TestUnlockByPullOne(t *testing.T) {
	t.Log("with one lock, UnlockByPull should...")
	s := miniredis.RunT(t)
	rdb := newTestRedis(s)
	_, _, _, err := rdb.TryLock(lock)
	Ok(t, err)

	t.Log("...delete nothing when its the same repo but a different pull")
	{
		_, _, err := rdb.UnlockByPull(project.RepoFullName, pullNum+1)
		Ok(t, err)
		ls, err := rdb.List()
		Ok(t, err)
		Equals(t, 1, len(ls))
	}
	t.Log("...delete nothing when its the same pull but a different repo")
	{
		_, _, err := rdb.UnlockByPull("different/repo", pullNum)
		Ok(t, err)
		ls, err := rdb.List()
		Ok(t, err)
		Equals(t, 1, len(ls))
	}
	t.Log("...delete the lock when its the same repo and pull")
	{
		_, _, err := rdb.UnlockByPull(project.RepoFullName, pullNum)
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
	_, _, _, err := rdb.TryLock(lock)
	Ok(t, err)
	_, _, err = rdb.Unlock(project, workspace)
	Ok(t, err)

	_, _, err = rdb.UnlockByPull(project.RepoFullName, pullNum)
	Ok(t, err)
	ls, err := rdb.List()
	Ok(t, err)
	Equals(t, 0, len(ls))
}

func TestUnlockByPullMatching(t *testing.T) {
	t.Log("UnlockByPull should delete all locks in that repo and pull num")
	s := miniredis.RunT(t)
	rdb := newTestRedis(s)
	_, _, _, err := rdb.TryLock(lock)
	Ok(t, err)

	// add additional locks with the same repo and pull num but different paths/workspaces
	new := lock
	new.Project.Path = "dif/path"
	_, _, _, err = rdb.TryLock(new)
	Ok(t, err)
	new2 := lock
	new2.Workspace = "new-workspace"
	_, _, _, err = rdb.TryLock(new2)
	Ok(t, err)

	// there should now be 3
	ls, err := rdb.List()
	Ok(t, err)
	Equals(t, 3, len(ls))

	// should all be unlocked
	_, _, err = rdb.UnlockByPull(project.RepoFullName, pullNum)
	Ok(t, err)
	ls, err = rdb.List()
	Ok(t, err)
	Equals(t, 0, len(ls))
}

func TestDequeueAfterUnlock(t *testing.T) {
	t.Log("unlocking should dequeue and grant lock to the next ProjectLock")
	s := miniredis.RunT(t)
	r := newTestRedisQueue(s)

	// first lock acquired
	firstLock := lock
	_, _, _, err := r.TryLock(firstLock)
	Ok(t, err)

	// second lock enqueued
	secondLock := firstLock
	secondLock.Pull.Num = pullNum + 1
	_, _, _, err = r.TryLock(secondLock)
	Ok(t, err)

	// third lock enqueued
	thirdLock := firstLock
	thirdLock.Pull.Num = pullNum + 2
	_, _, _, err = r.TryLock(thirdLock)
	Ok(t, err)
	queue, err := r.GetQueueByLock(firstLock.Project, firstLock.Workspace)
	Ok(t, err)
	Equals(t, 2, len(queue))
	Equals(t, secondLock.Pull, queue[0].Pull)
	Equals(t, thirdLock.Pull, queue[1].Pull)

	// first lock unlocked -> second lock dequeued and lock acquired
	_, dequeuedLock, err := r.Unlock(firstLock.Project, firstLock.Workspace)
	Ok(t, err)
	queue, err = r.GetQueueByLock(firstLock.Project, firstLock.Workspace)
	Ok(t, err)
	Equals(t, secondLock, *dequeuedLock)
	Equals(t, 1, len(queue))
	Equals(t, thirdLock.Pull, queue[0].Pull)

	// second lock unlocked -> third lock dequeued and lock acquired
	_, dequeuedLock, err = r.Unlock(secondLock.Project, secondLock.Workspace)
	Ok(t, err)
	Equals(t, thirdLock, *dequeuedLock)
	queue, err = r.GetQueueByLock(firstLock.Project, firstLock.Workspace)
	Ok(t, err)
	Equals(t, 0, len(queue))

	l, err := r.GetLock(project, workspace)
	Ok(t, err)
	Equals(t, thirdLock, *l)

	// Queue is deleted when empty
	queue, err = r.GetQueueByLock(thirdLock.Project, thirdLock.Workspace)
	Ok(t, err)
	Assert(t, queue == nil, "exp nil")

	// third lock unlocked -> no more locks in the queue
	_, dequeuedLock, err = r.Unlock(thirdLock.Project, thirdLock.Workspace)
	Ok(t, err)
	Equals(t, (*models.ProjectLock)(nil), dequeuedLock)
}

func TestDequeueAfterUnlockByPull(t *testing.T) {
	t.Log("unlocking by pull should dequeue and grant lock to all dequeued ProjectLocks")
	s := miniredis.RunT(t)
	r := newTestRedisQueue(s)

	_, _, _, err := r.TryLock(lock)
	Ok(t, err)

	lock2 := lock
	lock2.Workspace = "different-workspace"
	_, _, _, err = r.TryLock(lock2)
	Ok(t, err)

	lock3 := lock
	lock3.Pull.Num = pullNum + 1
	_, _, _, err = r.TryLock(lock3)
	Ok(t, err)

	lock4 := lock
	lock4.Workspace = "different-workspace"
	lock4.Pull.Num = pullNum + 1
	_, _, _, err = r.TryLock(lock4)
	Ok(t, err)

	_, dequeueStatus, err := r.UnlockByPull(project.RepoFullName, pullNum)
	Ok(t, err)

	Equals(t, 2, len(dequeueStatus.ProjectLocks))
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
	_, _, _, err := rdb.TryLock(lock)
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
				Failure:    "failure",
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
				Failure:    "failure",
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
				Failure:    "failure",
			},
			{
				RepoRelDir:   ".",
				Workspace:    "staging",
				ApplySuccess: "success!",
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
				Failure:    "failure",
			},
		})
	Ok(t, err)

	pull.HeadCommit = "newsha"
	status, err := rdb.UpdatePullWithResults(pull,
		[]command.ProjectResult{
			{
				RepoRelDir:   ".",
				Workspace:    "staging",
				ApplySuccess: "success!",
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
				Failure:    "failure",
			},
			{
				Command:     command.Plan,
				RepoRelDir:  "projectname",
				Workspace:   "default",
				ProjectName: "projectname",
				Failure:     "failure",
			},
			{
				Command:    command.Plan,
				RepoRelDir: "staythesame",
				Workspace:  "default",
				PlanSuccess: &models.PlanSuccess{
					TerraformOutput: "tf out",
					LockURL:         "lock-url",
					RePlanCmd:       "plan command",
					ApplyCmd:        "apply command",
				},
			},
		})
	Ok(t, err)

	updateStatus, err := rdb.UpdatePullWithResults(pull,
		[]command.ProjectResult{
			{
				Command:      command.Apply,
				RepoRelDir:   "mergeme",
				Workspace:    "default",
				ApplySuccess: "applied!",
			},
			{
				Command:     command.Apply,
				RepoRelDir:  "projectname",
				Workspace:   "default",
				ProjectName: "projectname",
				Error:       errors.New("apply error"),
			},
			{
				Command:      command.Apply,
				RepoRelDir:   "newresult",
				Workspace:    "default",
				ApplySuccess: "success!",
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
				Failure:    "policy failure",
				PolicyCheckResults: &models.PolicyCheckResults{
					PolicySetResults: []models.PolicySetResult{
						{
							PolicySetName: "policy1",
							ReqApprovals:  1,
						},
					},
				},
			},
			{
				Command:     command.PolicyCheck,
				RepoRelDir:  "projectname",
				Workspace:   "default",
				ProjectName: "projectname",
				Failure:     "policy failure",
				PolicyCheckResults: &models.PolicyCheckResults{
					PolicySetResults: []models.PolicySetResult{
						{
							PolicySetName: "policy1",
							ReqApprovals:  1,
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
	r, err := redis.New(mr.Host(), mr.Server().Addr().Port, "", false, false, 0, false)
	if err != nil {
		panic(errors.Wrap(err, "failed to create test redis client"))
	}
	return r
}

func newTestRedisTLS(mr *miniredis.Miniredis) *redis.RedisDB {
	r, err := redis.New(mr.Host(), mr.Server().Addr().Port, "", true, true, 0, false)
	if err != nil {
		panic(errors.Wrap(err, "failed to create test redis client"))
	}
	return r
}

func newTestRedisQueue(mr *miniredis.Miniredis) *redis.RedisDB {
	r, err := redis.New(mr.Host(), mr.Server().Addr().Port, "", false, false, 0, true)
	if err != nil {
		panic(errors.Wrap(err, "failed to create test redis client"))
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
