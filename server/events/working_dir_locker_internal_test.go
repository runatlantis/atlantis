// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package events

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

func TestWorkingDirLockMetadata(t *testing.T) {
	const sha = "0123456789abcdef0123456789abcdef01234567"
	tests := []struct {
		name     string
		hostType models.VCSHostType
		pullURL  string
		wantURL  string
	}{
		{
			name:     "GitHub uses the fork-safe pull commit route",
			hostType: models.Github,
			pullURL:  "https://github.com/owner/repo/pull/123",
			wantURL:  "https://github.com/owner/repo/pull/123/commits/" + sha,
		},
		{
			name:     "GitLab uses the merge request diff route",
			hostType: models.Gitlab,
			pullURL:  "https://gitlab.example.com/gitlab/group/repo/-/merge_requests/123",
			wantURL:  "https://gitlab.example.com/gitlab/group/repo/-/merge_requests/123/diffs?commit_id=" + sha,
		},
		{
			name:     "Gitea uses the pull commit route",
			hostType: models.Gitea,
			pullURL:  "https://gitea.example.com/owner/repo/pulls/123/",
			wantURL:  "https://gitea.example.com/owner/repo/pulls/123/commits/" + sha,
		},
		{
			name:     "unsupported provider falls back to SHA",
			hostType: models.BitbucketCloud,
			pullURL:  "https://bitbucket.org/owner/repo/pull-requests/123",
		},
		{
			name:     "missing pull URL falls back to SHA",
			hostType: models.Github,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := WorkingDirLockMetadataForPull(models.PullRequest{
				HeadCommit: sha,
				URL:        tt.pullURL,
				BaseRepo:   models.Repo{VCSHost: models.VCSHost{Type: tt.hostType}},
			})
			if metadata.HeadCommit != sha {
				t.Fatalf("expected head commit %q, got %q", sha, metadata.HeadCommit)
			}
			if metadata.CommitURL != tt.wantURL {
				t.Fatalf("expected commit URL %q, got %q", tt.wantURL, metadata.CommitURL)
			}
		})
	}
}

func TestWorkingDirLockMetadataWithoutHeadCommit(t *testing.T) {
	metadata := WorkingDirLockMetadataForPull(models.PullRequest{
		URL:      "https://github.com/owner/repo/pull/123",
		BaseRepo: models.Repo{VCSHost: models.VCSHost{Type: models.Github}},
	})
	if metadata != (WorkingDirLockMetadata{}) {
		t.Fatalf("expected empty metadata, got %#v", metadata)
	}
}

func TestWorkingDirLockMetadataForProject(t *testing.T) {
	ctx := command.ProjectContext{Pull: models.PullRequest{HeadCommit: "sha"}}
	metadata := WorkingDirLockMetadataForProject(ctx, "https://atlantis.example.com/jobs/job-id")
	if metadata.HeadCommit != "sha" || metadata.JobURL != "https://atlantis.example.com/jobs/job-id" {
		t.Fatalf("unexpected metadata: %#v", metadata)
	}
}

func TestWorkingDirLockErrorUsesOwnerMetadata(t *testing.T) {
	locker := NewDefaultWorkingDirLocker()
	owner := WorkingDirLockMetadata{HeadCommit: "owner-sha", CommitURL: "owner-commit", JobURL: "owner-job"}
	_, err := locker.TryLock("owner/repo", 1, "default", ".", "project", command.Plan, owner)
	if err != nil {
		t.Fatal(err)
	}
	_, err = locker.TryLock("owner/repo", 1, "default", ".", "project", command.Apply, WorkingDirLockMetadata{HeadCommit: "loser-sha", CommitURL: "loser-commit", JobURL: "loser-job"})
	lockErr, ok := err.(*workingDirLockError)
	if !ok {
		t.Fatalf("expected typed lock error, got %T", err)
	}
	if lockErr.metadata != owner {
		t.Fatalf("expected owner metadata %#v, got %#v", owner, lockErr.metadata)
	}
}

func TestTryLockPullUsesBlockingPlanProjectMetadata(t *testing.T) {
	locker := NewDefaultWorkingDirLocker()
	owner := WorkingDirLockMetadata{HeadCommit: "owner-sha", CommitURL: "owner-commit", JobURL: "owner-job"}
	_, err := locker.TryLock("owner/repo", 1, "default", ".", "project", command.Plan, owner)
	if err != nil {
		t.Fatal(err)
	}
	_, err = locker.TryLockPull("owner/repo", 1, command.Apply, WorkingDirLockMetadata{})
	lockErr, ok := err.(*workingDirLockError)
	if !ok {
		t.Fatalf("expected typed lock error, got %T", err)
	}
	if lockErr.metadata != owner || lockErr.multipleJobs {
		t.Fatalf("unexpected lock error: %#v", lockErr)
	}
}

func TestTryLockPullKeepsBlockingCommitMetadataTogether(t *testing.T) {
	const repo = "owner/repo"
	locker := NewDefaultWorkingDirLocker()
	_, err := locker.TryLockPull(repo, 1, command.Plan, WorkingDirLockMetadata{HeadCommit: "old", CommitURL: "old-commit"})
	if err != nil {
		t.Fatal(err)
	}
	for i, metadata := range []WorkingDirLockMetadata{
		{HeadCommit: "old", JobURL: "old-job"},
		{HeadCommit: "new", JobURL: "new-job"},
	} {
		_, err = locker.TryLock(repo, 1, "workspace", ".", []string{"old", "new"}[i], command.Plan, metadata)
		if err != nil {
			t.Fatal(err)
		}
	}

	_, err = locker.TryLockPull(repo, 1, command.Apply, WorkingDirLockMetadata{HeadCommit: "new"})
	lockErr := err.(*workingDirLockError)
	if lockErr.metadata.HeadCommit != "old" || lockErr.metadata.CommitURL != "old-commit" || lockErr.metadata.JobURL != "old-job" || lockErr.multipleJobs {
		t.Fatalf("mixed blocking commits: %#v", lockErr)
	}
}

func TestTryLockPullFindsPlanForCurrentCommit(t *testing.T) {
	const repo = "owner/repo"
	locker := NewDefaultWorkingDirLocker()
	for i, metadata := range []WorkingDirLockMetadata{
		{HeadCommit: "old", JobURL: "old-job"},
		{HeadCommit: "new", JobURL: "new-job"},
	} {
		_, err := locker.TryLock(repo, 1, "workspace", ".", []string{"old", "new"}[i], command.Plan, metadata)
		if err != nil {
			t.Fatal(err)
		}
	}

	_, err := locker.TryLockPull(repo, 1, command.Apply, WorkingDirLockMetadata{HeadCommit: "new"})
	lockErr := err.(*workingDirLockError)
	if lockErr.metadata.HeadCommit != "new" || lockErr.metadata.JobURL != "new-job" || lockErr.multipleJobs {
		t.Fatalf("selected wrong blocking commit: %#v", lockErr)
	}
}

func TestTryLockPullSelectsActiveProjectJobs(t *testing.T) {
	const repo = "owner/repo"
	const jobURL = "https://atlantis.example.com/jobs/job-id"
	tests := []struct {
		name         string
		jobURLs      []string
		wantJobURL   string
		multipleJobs bool
	}{
		{name: "zero"},
		{name: "one", jobURLs: []string{jobURL}, wantJobURL: jobURL},
		{name: "duplicate URL is one job", jobURLs: []string{jobURL, jobURL}, wantJobURL: jobURL},
		{name: "multiple", jobURLs: []string{jobURL + "-1", jobURL + "-2"}, multipleJobs: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			locker := NewDefaultWorkingDirLocker()
			_, err := locker.TryLockPull(repo, 1, command.Plan, WorkingDirLockMetadata{HeadCommit: "owner-sha", JobURL: "pull-job-is-ignored"})
			if err != nil {
				t.Fatal(err)
			}
			for i, url := range tt.jobURLs {
				_, err = locker.TryLock(repo, 1, "workspace", ".", []string{"a", "b"}[i], command.Plan, WorkingDirLockMetadata{HeadCommit: "owner-sha", JobURL: url})
				if err != nil {
					t.Fatal(err)
				}
			}
			_, err = locker.TryLockPull(repo, 1, command.Apply, WorkingDirLockMetadata{JobURL: "loser-job"})
			lockErr, ok := err.(*workingDirLockError)
			if !ok {
				t.Fatalf("expected typed lock error, got %T", err)
			}
			if lockErr.metadata.HeadCommit != "owner-sha" || lockErr.metadata.JobURL != tt.wantJobURL || lockErr.multipleJobs != tt.multipleJobs {
				t.Fatalf("unexpected lock error: %#v", lockErr)
			}
			if lockErr.Error() != "cannot run \"apply\": pull request 1 is currently locked by \"plan\" for commit owner-s.\nWait until the previous command is complete and try again" {
				t.Fatalf("unexpected lock error message: %s", lockErr)
			}
		})
	}
}

// waitForPullLockWaiter blocks until a LockPullAfterStalePlan caller is queued
// for repoFullName/pullNum.
func waitForPullLockWaiter(t *testing.T, d *DefaultWorkingDirLocker, repoFullName string, pullNum int) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		d.mutex.Lock()
		_, queued := d.waiters[d.pullLockPrefix(repoFullName, pullNum)]
		d.mutex.Unlock()
		if queued {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("timed out waiting for a queued pull lock waiter")
}

func TestLockPullAfterStalePlanGivesUpWhenPlanForSameCommitTakesOver(t *testing.T) {
	const repo = "owner/repo"
	locker := NewDefaultWorkingDirLocker()
	_, err := locker.TryLockPull(repo, 1, command.Plan, WorkingDirLockMetadata{HeadCommit: "old"})
	Ok(t, err)

	result := make(chan error, 1)
	go func() {
		_, err := locker.LockPullAfterStalePlan(repo, 1, command.Plan, WorkingDirLockMetadata{HeadCommit: "new"}, nil)
		result <- err
	}()
	waitForPullLockWaiter(t, locker, repo, 1)

	// A comment-triggered plan for the new commit takes the pull lock as the
	// old plan releases it, before the waiter re-checks.
	locker.mutex.Lock()
	locker.locks[locker.pullKey(repo, 1)] = workingDirLock{Command: command.Plan, WorkingDirLockMetadata: WorkingDirLockMetadata{HeadCommit: "new"}}
	locker.released.Broadcast()
	locker.mutex.Unlock()

	select {
	case err := <-result:
		Assert(t, errors.Is(err, ErrPullLockSuperseded), "expected superseded, got %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for LockPullAfterStalePlan to return")
	}
	locker.mutex.Lock()
	defer locker.mutex.Unlock()
	Equals(t, 0, len(locker.waiters))
}

func pullLockHeadCommit(locker *DefaultWorkingDirLocker, repoFullName string, pullNum int) (string, bool) {
	locker.mutex.Lock()
	defer locker.mutex.Unlock()
	lock, ok := locker.locks[locker.pullKey(repoFullName, pullNum)]
	return lock.HeadCommit, ok
}

type pausedWaiterLock struct {
	mutex  *sync.Mutex
	resume <-chan struct{}
}

func (l pausedWaiterLock) Lock() {
	<-l.resume
	l.mutex.Lock()
}

func (l pausedWaiterLock) Unlock() { l.mutex.Unlock() }

func TestAutoplanFreeLockSupersedesPreviousWaiter(t *testing.T) {
	locker := NewDefaultWorkingDirLocker()
	resume := make(chan struct{})
	locker.released.L = pausedWaiterLock{mutex: &locker.mutex, resume: resume}
	unlockOld, err := locker.TryLockPull("owner/repo", 1, command.Plan, WorkingDirLockMetadata{HeadCommit: "old"})
	Ok(t, err)
	waiting := make(chan struct{})
	result := make(chan error, 1)
	go func() {
		lock, err := locker.LockPullAfterStalePlan("owner/repo", 1, command.Plan, WorkingDirLockMetadata{HeadCommit: "previous"}, func(WorkingDirLockMetadata) { close(waiting) })
		if err == nil {
			lock.Unlock()
		}
		result <- err
	}()
	select {
	case <-waiting:
	case <-time.After(5 * time.Second):
		t.Fatal("previous autoplan did not wait")
	}
	unlockOld()
	latest, err := locker.LockPullAfterStalePlan("owner/repo", 1, command.Plan, WorkingDirLockMetadata{HeadCommit: "latest"}, nil)
	Ok(t, err)
	latest.Unlock()
	close(resume)
	select {
	case err := <-result:
		Assert(t, errors.Is(err, ErrPullLockSuperseded), "later arrival must discard the previous waiter even when it takes a free lock")
	case <-time.After(5 * time.Second):
		t.Fatal("previous waiter did not stop")
	}
}

func TestWorkingDirLockReleasePreservesReplacement(t *testing.T) {
	for _, pullLock := range []bool{true, false} {
		t.Run(fmt.Sprint(pullLock), func(t *testing.T) {
			locker := NewDefaultWorkingDirLocker()
			acquire := func() (func(), error) {
				if pullLock {
					return locker.TryLockPull("owner/repo", 1, command.Plan, WorkingDirLockMetadata{HeadCommit: "same"})
				}
				return locker.TryLock("owner/repo", 1, "default", ".", "", command.Plan, WorkingDirLockMetadata{HeadCommit: "same"})
			}
			staleUnlock, err := acquire()
			Ok(t, err)
			locker.UnlockByPull("owner/repo", 1)
			currentUnlock, err := acquire()
			Ok(t, err)
			staleUnlock()
			staleUnlock()
			_, err = acquire()
			Assert(t, err != nil, "stale release must not unlock the replacement with identical metadata")
			currentUnlock()
			unlock, err := acquire()
			Ok(t, err)
			unlock()
		})
	}
}

func TestAutoplanLockHandleRejectsReplacementMetadataUpdate(t *testing.T) {
	locker := NewDefaultWorkingDirLocker()
	stale, err := locker.LockPullAfterStalePlan("owner/repo", 1, command.Plan, WorkingDirLockMetadata{HeadCommit: "same"}, nil)
	Ok(t, err)
	Assert(t, stale.UpdateMetadata(WorkingDirLockMetadata{HeadCommit: "same"}), "owner may refresh its metadata")
	locker.UnlockByPull("owner/repo", 1)
	current, err := locker.LockPullAfterStalePlan("owner/repo", 1, command.Plan, WorkingDirLockMetadata{HeadCommit: "same"}, nil)
	Ok(t, err)
	defer current.Unlock()
	Assert(t, !stale.UpdateMetadata(WorkingDirLockMetadata{HeadCommit: "wrong"}), "revoked owner may not refresh replacement metadata")
	stale.Unlock()
	head, held := pullLockHeadCommit(locker, "owner/repo", 1)
	Assert(t, held, "stale autoplan release must preserve replacement")
	Equals(t, "same", head)
}

func TestAutoplanAuthorizationSeparatesPendingAndRunningPlans(t *testing.T) {
	for _, start := range []bool{false, true} {
		t.Run(fmt.Sprint(start), func(t *testing.T) {
			locker := NewDefaultWorkingDirLocker()
			lock, err := locker.LockPullAfterStalePlan("owner/repo", 1, command.Plan, WorkingDirLockMetadata{HeadCommit: "event"}, nil)
			Ok(t, err)
			defer lock.Unlock()
			Assert(t, lock.UpdateMetadata(WorkingDirLockMetadata{HeadCommit: "live"}), "owner may refresh before authorization")
			if start {
				Assert(t, lock.Start(), "uncancelled attempt must be authorized")
			}
			locker.CancelAutoplanWaiter("owner/repo", 1)
			if start {
				Assert(t, lock.UpdateMetadata(WorkingDirLockMetadata{HeadCommit: "live"}), "waiter cancellation must not revoke a running plan")
				Assert(t, locker.HasCommandLock("owner/repo", 1, command.Plan), "running plan must keep its lock")
			} else {
				Assert(t, !lock.Start(), "cancellation after refresh must prevent authorization")
			}
		})
	}
}

func TestWorkingDirLocker_StressAutoplanWaitersAndSharedApplies(t *testing.T) {
	const repo = "owner/repo"
	for i := range 40 {
		locker := NewDefaultWorkingDirLocker()
		unlock, err := locker.TryLockPull(repo, 1, command.Plan, WorkingDirLockMetadata{HeadCommit: "base"})
		Ok(t, err)

		var wg sync.WaitGroup
		wg.Add(8)
		for n := range 8 {
			sha := fmt.Sprintf("sha-%d-%d", i, n)
			go func() {
				defer wg.Done()
				lock, lockErr := locker.LockPullAfterStalePlan(repo, 1, command.Plan, WorkingDirLockMetadata{HeadCommit: sha}, nil)
				if lockErr != nil {
					return
				}
				if !lock.Start() {
					lock.Unlock()
					return
				}
				lock.Unlock()
			}()
		}
		time.Sleep(time.Millisecond)
		unlock()

		finished := make(chan struct{})
		go func() {
			wg.Wait()
			close(finished)
		}()
		select {
		case <-finished:
		case <-time.After(5 * time.Second):
			t.Fatalf("iteration %d: autoplan waiters did not finish", i)
		}

		sharedA, err := locker.TryLockPullShared(repo, 2, command.Apply, WorkingDirLockMetadata{HeadCommit: "same"})
		Ok(t, err)
		sharedB, err := locker.TryLockPullShared(repo, 2, command.Apply, WorkingDirLockMetadata{HeadCommit: "same"})
		Ok(t, err)
		sharedA()
		sharedB()

		locker.mutex.Lock()
		leftoverLocks := len(locker.locks)
		leftoverWaiters := len(locker.waiters)
		locker.mutex.Unlock()
		if leftoverLocks != 0 || leftoverWaiters != 0 {
			t.Fatalf("iteration %d: leftover locks=%d waiters=%d", i, leftoverLocks, leftoverWaiters)
		}
	}
}
