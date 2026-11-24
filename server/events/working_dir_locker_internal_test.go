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
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
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

func TestPlanCommandRunnerAutoplanWaitsForPlanOfOlderCommit(t *testing.T) {
	repo := models.Repo{FullName: "owner/repo"}
	locker := NewDefaultWorkingDirLocker()
	unlockOld, err := locker.TryLockPull(repo.FullName, 1, command.Plan, WorkingDirLockMetadata{HeadCommit: "old"})
	Ok(t, err)

	// commitStatusUpdater and pullUpdater are nil: waiting must not publish a
	// failed status or comment for the new head.
	runner := &PlanCommandRunner{workingDirLocker: locker, LivePullHeadFetcher: &stubLivePullHeadFetcher{head: "new"}}
	ctx := &command.Context{
		Log:     logging.NewNoopLogger(t),
		Pull:    models.PullRequest{BaseRepo: repo, Num: 1, HeadCommit: "new"},
		Trigger: command.AutoTrigger,
	}

	type lockResult struct {
		unlock func()
		ok     bool
	}
	result := make(chan lockResult, 1)
	go func() {
		unlock, ok := runner.lockPullForPlan(ctx, AutoplanCommand{})
		result <- lockResult{unlock: unlock, ok: ok}
	}()
	waitForPullLockWaiter(t, locker, repo.FullName, 1)
	unlockOld()

	select {
	case r := <-result:
		Assert(t, r.ok, "expected autoplan to get the pull lock")
		locker.mutex.Lock()
		Equals(t, "new", locker.locks[locker.pullKey(repo.FullName, 1)].HeadCommit)
		locker.mutex.Unlock()
		r.unlock()
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for autoplan to take the pull lock")
	}
}

func TestPlanCommandRunnerSupersededAutoplanExitsQuietly(t *testing.T) {
	repo := models.Repo{FullName: "owner/repo"}
	locker := NewDefaultWorkingDirLocker()
	unlockOld, err := locker.TryLockPull(repo.FullName, 1, command.Plan, WorkingDirLockMetadata{HeadCommit: "old"})
	Ok(t, err)
	defer unlockOld()

	runner := &PlanCommandRunner{workingDirLocker: locker}
	newCtx := func(sha string) *command.Context {
		return &command.Context{
			Log:     logging.NewNoopLogger(t),
			Pull:    models.PullRequest{BaseRepo: repo, Num: 1, HeadCommit: sha},
			Trigger: command.AutoTrigger,
		}
	}

	first := make(chan bool, 1)
	go func() {
		_, ok := runner.lockPullForPlan(newCtx("new"), AutoplanCommand{})
		first <- ok
	}()
	waitForPullLockWaiter(t, locker, repo.FullName, 1)

	second := make(chan bool, 1)
	go func() {
		unlock, ok := runner.lockPullForPlan(newCtx("newer"), AutoplanCommand{})
		if ok {
			unlock()
		}
		second <- ok
	}()

	select {
	case ok := <-first:
		Assert(t, !ok, "expected superseded autoplan to stop")
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for superseded autoplan to stop")
	}

	locker.UnlockByPull(repo.FullName, 1)
	select {
	case ok := <-second:
		Assert(t, !ok, "expected cancelled autoplan to stop")
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for cancelled autoplan to stop")
	}
}

type stubLivePullHeadFetcher struct {
	head    string
	err     error
	calls   int
	onFetch func()
}

func (f *stubLivePullHeadFetcher) GetLivePullIdentity(command.ProjectContext) (models.PullRequest, error) {
	f.calls++
	if f.onFetch != nil {
		f.onFetch()
	}
	return models.PullRequest{HeadCommit: f.head}, f.err
}

// runAutoplanLockBehindStalePlan holds a plan lock for commit "old", starts
// an autoplan lock for commit "new" with runner, releases the old lock once
// the autoplan is waiting, and returns what the autoplan got.
func runAutoplanLockBehindStalePlan(t *testing.T, runner *PlanCommandRunner, locker *DefaultWorkingDirLocker) (*command.Context, func(), bool) {
	t.Helper()
	repo := models.Repo{FullName: "owner/repo"}
	unlockOld, err := locker.TryLockPull(repo.FullName, 1, command.Plan, WorkingDirLockMetadata{HeadCommit: "old"})
	Ok(t, err)

	ctx := &command.Context{
		Log:     logging.NewNoopLogger(t),
		Pull:    models.PullRequest{BaseRepo: repo, Num: 1, HeadCommit: "new"},
		Trigger: command.AutoTrigger,
	}
	type lockResult struct {
		unlock func()
		ok     bool
	}
	result := make(chan lockResult, 1)
	go func() {
		unlock, ok := runner.lockPullForPlan(ctx, AutoplanCommand{})
		result <- lockResult{unlock: unlock, ok: ok}
	}()
	waitForPullLockWaiter(t, locker, repo.FullName, 1)
	unlockOld()

	select {
	case r := <-result:
		return ctx, r.unlock, r.ok
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for autoplan lock")
	}
	return nil, nil, false
}

func pullLockHeadCommit(locker *DefaultWorkingDirLocker, repoFullName string, pullNum int) (string, bool) {
	locker.mutex.Lock()
	defer locker.mutex.Unlock()
	lock, ok := locker.locks[locker.pullKey(repoFullName, pullNum)]
	return lock.HeadCommit, ok
}

func TestPlanCommandRunnerAutoplanPlansLiveHeadAfterWaiting(t *testing.T) {
	locker := NewDefaultWorkingDirLocker()
	fetcher := &stubLivePullHeadFetcher{head: "newest"}
	runner := &PlanCommandRunner{workingDirLocker: locker, LivePullHeadFetcher: fetcher}

	ctx, unlock, ok := runAutoplanLockBehindStalePlan(t, runner, locker)
	Assert(t, ok, "expected autoplan to get the pull lock")
	defer unlock()

	Equals(t, "newest", ctx.Pull.HeadCommit)
	head, locked := pullLockHeadCommit(locker, "owner/repo", 1)
	Assert(t, locked, "expected pull lock to be held")
	Equals(t, "newest", head)
	Equals(t, 1, fetcher.calls)
}

func TestPlanCommandRunnerAutoplanReportsUnavailableLiveHead(t *testing.T) {
	for name, fetcher := range map[string]*stubLivePullHeadFetcher{
		"not configured": nil,
		"fetch error":    {err: errors.New("boom")},
		"empty head":     {},
	} {
		t.Run(name, func(t *testing.T) {
			locker := NewDefaultWorkingDirLocker()
			statuses := &issue7050StatusRecorder{}
			comments := &issue7050CommentClient{NotConfiguredVCSClient: &vcs.NotConfiguredVCSClient{}}
			runner := &PlanCommandRunner{
				workingDirLocker:    locker,
				commitStatusUpdater: statuses,
				pullUpdater: &PullUpdater{
					VCSClient:        comments,
					MarkdownRenderer: NewMarkdownRenderer(false, false, false, false, false, false, "", "atlantis", false, false),
				},
			}
			if fetcher != nil {
				runner.LivePullHeadFetcher = fetcher
			}
			_, _, ok := runAutoplanLockBehindStalePlan(t, runner, locker)
			Assert(t, !ok, "must not plan an unverified event head")
			_, locked := pullLockHeadCommit(locker, "owner/repo", 1)
			Assert(t, !locked, "expected pull lock to be released")
			updates := statuses.snapshot()
			Equals(t, 1, len(updates))
			Equals(t, models.FailedCommitStatus, updates[0].status)
			Equals(t, 1, len(comments.snapshot()))
		})
	}
}

func TestPlanCommandRunnerAutoplanSkipsAfterWaitingWhenShuttingDown(t *testing.T) {
	locker := NewDefaultWorkingDirLocker()
	drainer := &Drainer{}
	drainer.ShutdownBlocking()
	fetcher := &stubLivePullHeadFetcher{head: "newest"}
	// commitStatusUpdater and pullUpdater are nil: skipping must not publish anything.
	runner := &PlanCommandRunner{workingDirLocker: locker, LivePullHeadFetcher: fetcher, Drainer: drainer}

	_, _, ok := runAutoplanLockBehindStalePlan(t, runner, locker)
	Assert(t, !ok, "expected autoplan to stop during shutdown")
	_, locked := pullLockHeadCommit(locker, "owner/repo", 1)
	Assert(t, !locked, "expected pull lock to be released")
	Equals(t, 0, fetcher.calls)
}

func TestPlanCommandRunnerAutoplanDoesNotFetchLiveHeadWithoutWaiting(t *testing.T) {
	locker := NewDefaultWorkingDirLocker()
	fetcher := &stubLivePullHeadFetcher{head: "newest"}
	runner := &PlanCommandRunner{workingDirLocker: locker, LivePullHeadFetcher: fetcher}
	ctx := &command.Context{
		Log:     logging.NewNoopLogger(t),
		Pull:    models.PullRequest{BaseRepo: models.Repo{FullName: "owner/repo"}, Num: 1, HeadCommit: "new"},
		Trigger: command.AutoTrigger,
	}

	unlock, ok := runner.lockPullForPlan(ctx, AutoplanCommand{})
	Assert(t, ok, "expected autoplan to get the pull lock")
	defer unlock()
	Equals(t, "new", ctx.Pull.HeadCommit)
	Equals(t, 0, fetcher.calls)
}

func TestPlanCommandRunnerAutoplanShutdownDuringLiveHeadFetch(t *testing.T) {
	locker := NewDefaultWorkingDirLocker()
	drainer := &Drainer{}
	fetcher := &stubLivePullHeadFetcher{head: "newest", onFetch: drainer.ShutdownBlocking}
	runner := &PlanCommandRunner{workingDirLocker: locker, LivePullHeadFetcher: fetcher, Drainer: drainer}
	_, _, ok := runAutoplanLockBehindStalePlan(t, runner, locker)
	Assert(t, !ok, "must not start planning after shutdown begins during lookup")
	_, locked := pullLockHeadCommit(locker, "owner/repo", 1)
	Assert(t, !locked, "expected pull lock to be released")
}

func TestPlanCommandRunnerAutoplanRefreshCanBeSuperseded(t *testing.T) {
	locker := NewDefaultWorkingDirLocker()
	result := make(chan PullLock, 1)
	queued := make(chan struct{})
	fetcher := &stubLivePullHeadFetcher{head: "newest", onFetch: func() {
		go func() {
			lock, err := locker.LockPullAfterStalePlan("owner/repo", 1, command.Plan, WorkingDirLockMetadata{HeadCommit: "newest"}, func(WorkingDirLockMetadata) { close(queued) })
			Ok(t, err)
			result <- lock
		}()
		select {
		case <-queued:
		case <-time.After(5 * time.Second):
			t.Fatal("later autoplan did not queue")
		}
	}}
	runner := &PlanCommandRunner{workingDirLocker: locker, LivePullHeadFetcher: fetcher}
	_, _, ok := runAutoplanLockBehindStalePlan(t, runner, locker)
	Assert(t, !ok, "later arrival must supersede autoplan still fetching live head")
	select {
	case lock := <-result:
		defer lock.Unlock()
		Assert(t, lock.Start(), "latest attempt must be authorized")
	case <-time.After(5 * time.Second):
		t.Fatal("latest attempt did not acquire lock")
	}
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

func TestPlanCommandRunnerAdmittedAutoplanCanLockDuringShutdown(t *testing.T) {
	locker := NewDefaultWorkingDirLocker()
	drainer := &Drainer{}
	drainer.ShutdownBlocking()
	runner := &PlanCommandRunner{workingDirLocker: locker, Drainer: drainer}
	ctx := &command.Context{Log: logging.NewNoopLogger(t), Pull: models.PullRequest{BaseRepo: models.Repo{FullName: "owner/repo"}, Num: 1, HeadCommit: "new"}, Trigger: command.AutoTrigger}
	unlock, ok := runner.lockPullForPlan(ctx, AutoplanCommand{})
	Assert(t, ok, "an autoplan already admitted by the drainer must continue")
	unlock()
	_, locked := pullLockHeadCommit(locker, "owner/repo", 1)
	Assert(t, !locked, "completed plan must release its lock")
}

func TestPlanCommandRunnerAutoplanCancellationDuringLiveHeadFetch(t *testing.T) {
	for _, replacement := range []command.Name{command.Plan, command.Apply} {
		for _, fetchErr := range []error{nil, errors.New("lookup unavailable")} {
			t.Run(fmt.Sprintf("%s/%v", replacement, fetchErr), func(t *testing.T) {
				locker := NewDefaultWorkingDirLocker()
				started, resume := make(chan struct{}), make(chan struct{})
				fetcher := &stubLivePullHeadFetcher{head: "live", err: fetchErr, onFetch: func() {
					close(started)
					<-resume
				}}
				runner := &PlanCommandRunner{workingDirLocker: locker, LivePullHeadFetcher: fetcher}
				ctx := &command.Context{Log: logging.NewNoopLogger(t), Pull: models.PullRequest{BaseRepo: models.Repo{FullName: "owner/repo"}, Num: 1, HeadCommit: "new"}, Trigger: command.AutoTrigger}
				unlockOld, err := locker.TryLockPull("owner/repo", 1, command.Plan, WorkingDirLockMetadata{HeadCommit: "old"})
				Ok(t, err)
				result := make(chan bool, 1)
				go func() {
					unlock, ok := runner.lockPullForPlan(ctx, AutoplanCommand{})
					if ok {
						unlock()
					}
					result <- ok
				}()
				waitForPullLockWaiter(t, locker, "owner/repo", 1)
				unlockOld()
				select {
				case <-started:
				case <-time.After(5 * time.Second):
					t.Fatal("live head fetch did not start")
				}
				locker.UnlockByPull("owner/repo", 1)
				unlockReplacement, err := locker.TryLockPull("owner/repo", 1, replacement, WorkingDirLockMetadata{HeadCommit: "new"})
				Ok(t, err)
				defer unlockReplacement()
				close(resume)
				select {
				case ok := <-result:
					Assert(t, !ok, "cancelled autoplan must not continue under a replacement lock")
				case <-time.After(5 * time.Second):
					t.Fatal("cancelled autoplan did not stop")
				}
				head, held := pullLockHeadCommit(locker, "owner/repo", 1)
				Assert(t, held, "stale release must preserve the replacement lock")
				Equals(t, "new", head)
				Assert(t, locker.HasCommandLock("owner/repo", 1, replacement), "replacement command must retain ownership")
			})
		}
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

func TestPlanCommandRunnerRejectsClosedLivePull(t *testing.T) {
	locker := NewDefaultWorkingDirLocker()
	fetcher := &closedLivePullFetcher{}
	runner := &PlanCommandRunner{workingDirLocker: locker, LivePullHeadFetcher: fetcher}
	_, _, ok := runAutoplanLockBehindStalePlan(t, runner, locker)
	Assert(t, !ok, "closed pull must not be planned even if its head is nonempty")
	_, held := pullLockHeadCommit(locker, "owner/repo", 1)
	Assert(t, !held, "closed pull must release pending lock")
}

type closedLivePullFetcher struct{}

func (*closedLivePullFetcher) GetLivePullIdentity(command.ProjectContext) (models.PullRequest, error) {
	return models.PullRequest{HeadCommit: "live", State: models.ClosedPullState}, nil
}
