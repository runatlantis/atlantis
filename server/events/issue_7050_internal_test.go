// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/core/boltdb"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/models/testdata"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

// Regression tests for races found while testing the #7050 fix: a late
// webhook for a head an autoplan already adopted, a panic while fetching the
// live head, and overlapping targeted applies publishing statuses out of order.

type issue7050StatusUpdate struct {
	head   string
	status models.CommitStatus
	cmd    command.Name
}

type issue7050StatusRecorder struct {
	mu      sync.Mutex
	updates []issue7050StatusUpdate
}

func (s *issue7050StatusRecorder) UpdateCombined(_ logging.SimpleLogging, _ models.Repo, pull models.PullRequest, status models.CommitStatus, cmd command.Name) error {
	s.mu.Lock()
	s.updates = append(s.updates, issue7050StatusUpdate{head: pull.HeadCommit, status: status, cmd: cmd})
	s.mu.Unlock()
	return nil
}

func (s *issue7050StatusRecorder) UpdateCombinedCount(logging.SimpleLogging, models.Repo, models.PullRequest, models.CommitStatus, command.Name, models.ProjectCounts) error {
	return nil
}

func (s *issue7050StatusRecorder) UpdatePreWorkflowHook(logging.SimpleLogging, models.PullRequest, models.CommitStatus, string, string, string) error {
	return nil
}

func (s *issue7050StatusRecorder) UpdatePostWorkflowHook(logging.SimpleLogging, models.PullRequest, models.CommitStatus, string, string, string) error {
	return nil
}

func (s *issue7050StatusRecorder) snapshot() []issue7050StatusUpdate {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]issue7050StatusUpdate, len(s.updates))
	copy(out, s.updates)
	return out
}

type issue7050CommentClient struct {
	*vcs.NotConfiguredVCSClient
	mu       sync.Mutex
	comments []string
}

func (c *issue7050CommentClient) CreateComment(_ logging.SimpleLogging, _ models.Repo, _ int, comment string, _ string) error {
	c.mu.Lock()
	c.comments = append(c.comments, comment)
	c.mu.Unlock()
	return nil
}

func (c *issue7050CommentClient) snapshot() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]string, len(c.comments))
	copy(out, c.comments)
	return out
}

// TestAutoplan_LateWebhookForAdoptedLiveHeadExitsQuietly adopts a newer live head while an
// older plan lock is outstanding, then delivers the webhook for that live
// head. That webhook is the first plan for the commit, so it must not post a
// failed status or a lock comment while the adopted plan still holds the lock.
func TestAutoplan_LateWebhookForAdoptedLiveHeadExitsQuietly(t *testing.T) {
	locker := NewDefaultWorkingDirLocker()
	statuses := &issue7050StatusRecorder{}
	comments := &issue7050CommentClient{NotConfiguredVCSClient: &vcs.NotConfiguredVCSClient{}}
	runner := &PlanCommandRunner{
		workingDirLocker:    locker,
		LivePullHeadFetcher: &stubLivePullHeadFetcher{head: "live"},
		commitStatusUpdater: statuses,
		pullUpdater: &PullUpdater{
			VCSClient:        comments,
			MarkdownRenderer: NewMarkdownRenderer(false, false, false, false, false, false, "", "atlantis", false, false),
		},
	}

	ctx, unlock, ok := runAutoplanLockBehindStalePlan(t, runner, locker)
	Assert(t, ok, "expected the waiting autoplan to adopt the live head and take the lock")
	defer unlock()
	Equals(t, "live", ctx.Pull.HeadCommit)
	beforeStatus := len(statuses.snapshot())
	beforeComments := len(comments.snapshot())

	webhook := &command.Context{
		Log: logging.NewNoopLogger(t),
		Pull: models.PullRequest{
			BaseRepo:   ctx.Pull.BaseRepo,
			Num:        ctx.Pull.Num,
			HeadCommit: ctx.Pull.HeadCommit,
		},
		Trigger: command.AutoTrigger,
	}
	_, webhookOK := runner.lockPullForPlan(webhook, AutoplanCommand{})
	Assert(t, !webhookOK, "the live head is already being planned, so its webhook must not start a second plan")

	var failed []issue7050StatusUpdate
	for _, update := range statuses.snapshot()[beforeStatus:] {
		if update.status == models.FailedCommitStatus {
			failed = append(failed, update)
		}
	}
	posted := comments.snapshot()[beforeComments:]
	if len(failed) > 0 || len(posted) > 0 {
		t.Fatalf("webhook for the commit the waiter adopted posted a failure while that plan holds the lock: statuses=%+v comments=%q", failed, posted)
	}
}

// TestAutoplan_PanicDuringLiveHeadFetchReleasesPullLock panics after the waiter has
// the pull lock and before Start. The lock and waiter slot must be released so
// a later autoplan can acquire the pull.
func TestAutoplan_PanicDuringLiveHeadFetchReleasesPullLock(t *testing.T) {
	repo := models.Repo{FullName: "owner/repo"}
	locker := NewDefaultWorkingDirLocker()
	unlockOld, err := locker.TryLockPull(repo.FullName, 1, command.Plan, WorkingDirLockMetadata{HeadCommit: "old"})
	Ok(t, err)

	runner := &PlanCommandRunner{
		workingDirLocker: locker,
		LivePullHeadFetcher: &stubLivePullHeadFetcher{
			head:    "live",
			onFetch: func() { panic("live head fetch panicked") },
		},
	}
	ctx := &command.Context{
		Log:     logging.NewNoopLogger(t),
		Pull:    models.PullRequest{BaseRepo: repo, Num: 1, HeadCommit: "event"},
		Trigger: command.AutoTrigger,
	}

	type outcome struct {
		unlock   func()
		ok       bool
		panicked any
	}
	result := make(chan outcome, 1)
	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				result <- outcome{panicked: recovered}
			}
		}()
		unlock, ok := runner.lockPullForPlan(ctx, AutoplanCommand{})
		result <- outcome{unlock: unlock, ok: ok}
	}()
	waitForPullLockWaiter(t, locker, repo.FullName, 1)
	unlockOld()

	var got outcome
	select {
	case got = <-result:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for the autoplan that panics during live head fetch")
	}
	if got.panicked == nil {
		if got.ok && got.unlock != nil {
			got.unlock()
		}
		t.Fatal("expected the live head fetch to panic")
	}

	later := make(chan error, 1)
	go func() {
		lock, lockErr := locker.LockPullAfterStalePlan(repo.FullName, 1, command.Plan, WorkingDirLockMetadata{HeadCommit: "later"}, nil)
		if lockErr == nil {
			if !lock.Start() {
				lock.Unlock()
				later <- errors.New("later autoplan acquired the lock but Start failed")
				return
			}
			lock.Unlock()
		}
		later <- lockErr
	}()
	select {
	case lockErr := <-later:
		Ok(t, lockErr)
	case <-time.After(500 * time.Millisecond):
		locker.UnlockByPull(repo.FullName, 1)
		<-later
		t.Fatal("panic during live-head fetch leaked the pull lock; a later autoplan stayed blocked until UnlockByPull")
	}
}

type issue7050CountPost struct {
	status models.CommitStatus
	counts models.ProjectCounts
}

// issue7050GatedApplyStatus lets the first partial apply snapshot sit inside the
// status post until the other apply has posted a complete success. That is the
// order UpdateCombinedCount can observe once two targeted applies overlap.
type issue7050GatedApplyStatus struct {
	mu          sync.Mutex
	posts       []issue7050CountPost
	partialSeen chan struct{}
	fullPosted  chan struct{}
	oncePartial sync.Once
	onceFull    sync.Once
}

func (s *issue7050GatedApplyStatus) UpdateCombined(logging.SimpleLogging, models.Repo, models.PullRequest, models.CommitStatus, command.Name) error {
	return nil
}

func (s *issue7050GatedApplyStatus) UpdateCombinedCount(_ logging.SimpleLogging, _ models.Repo, _ models.PullRequest, status models.CommitStatus, _ command.Name, counts models.ProjectCounts) error {
	partial := status == models.PendingCommitStatus && counts.Total == 2 && counts.Success == 1
	if partial {
		s.oncePartial.Do(func() { close(s.partialSeen) })
		// Act like a slow status request. Without serialization the other
		// apply posts its complete count meanwhile and this older count
		// lands after it.
		select {
		case <-s.fullPosted:
		case <-time.After(300 * time.Millisecond):
		}
	}
	s.mu.Lock()
	s.posts = append(s.posts, issue7050CountPost{status: status, counts: counts})
	full := status == models.SuccessCommitStatus && counts.Success == counts.Total && counts.Total == 2
	s.mu.Unlock()
	if full {
		s.onceFull.Do(func() { close(s.fullPosted) })
	}
	return nil
}

func (s *issue7050GatedApplyStatus) UpdatePreWorkflowHook(logging.SimpleLogging, models.PullRequest, models.CommitStatus, string, string, string) error {
	return nil
}

func (s *issue7050GatedApplyStatus) UpdatePostWorkflowHook(logging.SimpleLogging, models.PullRequest, models.CommitStatus, string, string, string) error {
	return nil
}

func (s *issue7050GatedApplyStatus) snapshot() []issue7050CountPost {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]issue7050CountPost, len(s.posts))
	copy(out, s.posts)
	return out
}

type issue7050ProjectBuilder struct {
	byName map[string]command.ProjectContext
}

func (b issue7050ProjectBuilder) BuildApplyCommands(_ *command.Context, cmd *CommentCommand) ([]command.ProjectContext, error) {
	project, ok := b.byName[cmd.ProjectName]
	if !ok {
		return nil, fmt.Errorf("no project %q", cmd.ProjectName)
	}
	return []command.ProjectContext{project}, nil
}

func (issue7050ProjectBuilder) ShouldIgnoreTargetedDir(*command.Context, *CommentCommand) bool {
	return false
}

// issue7050PhasedApplyRunner releases both applies from the same point, then holds
// one until the other has published its database snapshot. The held apply
// therefore commits the complete 2/2 result while the first status post is
// still in flight.
type issue7050PhasedApplyRunner struct {
	entered  *sync.WaitGroup
	release  <-chan struct{}
	holdName string
}

func (r *issue7050PhasedApplyRunner) Apply(ctx command.ProjectContext) command.ProjectCommandOutput {
	r.entered.Done()
	r.entered.Wait()
	if ctx.ProjectName == r.holdName {
		<-r.release
	}
	return command.ProjectCommandOutput{ApplySuccess: "applied"}
}

// TestApply_OverlappingTargetedAppliesKeepLatestStatus runs apply -p projA and
// apply -p projB for the same commit. Both are allowed through the shared
// pull lock. The combined commit status must not move from "2/2 applied"
// back to a pending 1/2 snapshot.
func TestApply_OverlappingTargetedAppliesKeepLatestStatus(t *testing.T) {
	database, err := boltdb.New(t.TempDir())
	Ok(t, err)
	t.Cleanup(func() { database.Close() })

	pull := models.PullRequest{
		BaseRepo:   testdata.GithubRepo,
		State:      models.OpenPullState,
		Num:        testdata.Pull.Num,
		HeadCommit: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		BaseBranch: "main",
	}
	_, err = database.UpdatePullWithResults(pull, []command.ProjectResult{
		internalPlannedProjectResult("dirA", DefaultWorkspace, "projA"),
		internalPlannedProjectResult("dirB", DefaultWorkspace, "projB"),
	})
	Ok(t, err)

	project := func(dir, name string) command.ProjectContext {
		return command.ProjectContext{
			CommandName:       command.Apply,
			RepoRelDir:        dir,
			Workspace:         DefaultWorkspace,
			ProjectName:       name,
			ProjectPlanStatus: models.PlannedPlanStatus,
			Pull:              pull,
			SilencePRComments: []string{command.Apply.String()},
		}
	}
	statuses := &issue7050GatedApplyStatus{
		partialSeen: make(chan struct{}),
		fullPosted:  make(chan struct{}),
	}
	entered := &sync.WaitGroup{}
	entered.Add(2)
	projectRunner := &issue7050PhasedApplyRunner{entered: entered, release: statuses.partialSeen, holdName: "projB"}
	vcsClient := &vcs.NotConfiguredVCSClient{Host: models.Github}
	pullUpdater := &PullUpdater{
		VCSClient:        vcsClient,
		MarkdownRenderer: NewMarkdownRenderer(false, false, false, false, false, false, "", "atlantis", false, false),
	}
	builder := issue7050ProjectBuilder{byName: map[string]command.ProjectContext{
		"projA": project("dirA", "projA"),
		"projB": project("dirB", "projB"),
	}}
	locker := NewDefaultWorkingDirLocker()
	newRunner := func() *ApplyCommandRunner {
		return NewApplyCommandRunner(
			vcsClient,
			false,
			unlockedApplyLockChecker{},
			statuses,
			builder,
			projectRunner,
			NewCancellationTracker(),
			&AutoMerger{VCSClient: vcsClient},
			pullUpdater,
			&DBUpdater{Database: database},
			database,
			1,
			false,
			false,
			locker,
			noopPullReqStatusFetcher{},
			nil,
			"",
		)
	}
	runnerA := newRunner()
	runnerB := newRunner()
	ctxA := newInternalApplyContext(t, pull)
	ctxB := newInternalApplyContext(t, pull)

	errA := make(chan error, 1)
	errB := make(chan error, 1)
	go func() {
		runnerA.Run(ctxA, &CommentCommand{Name: command.Apply, ProjectName: "projA"})
		if ctxA.CommandHasErrors {
			errA <- errors.New("projA apply recorded command errors")
			return
		}
		errA <- nil
	}()
	go func() {
		runnerB.Run(ctxB, &CommentCommand{Name: command.Apply, ProjectName: "projB"})
		if ctxB.CommandHasErrors {
			errB <- errors.New("projB apply recorded command errors")
			return
		}
		errB <- nil
	}()

	timeout := time.After(8 * time.Second)
	select {
	case applyErr := <-errA:
		Ok(t, applyErr)
	case <-timeout:
		t.Fatal("timed out waiting for projA apply")
	}
	select {
	case applyErr := <-errB:
		Ok(t, applyErr)
	case <-timeout:
		t.Fatal("timed out waiting for projB apply")
	}

	stored, err := database.GetPullStatus(pull)
	Ok(t, err)
	if stored == nil || len(stored.Projects) != 2 {
		t.Fatalf("expected both projects to stay in the database, got %#v", stored)
	}
	for _, projectStatus := range stored.Projects {
		Equals(t, models.AppliedPlanStatus, projectStatus.Status)
	}

	posts := statuses.snapshot()
	sawFullSuccess := false
	regressed := false
	for _, post := range posts {
		if post.status == models.SuccessCommitStatus && post.counts.Success == post.counts.Total && post.counts.Total == 2 {
			sawFullSuccess = true
		}
		if sawFullSuccess && post.status == models.PendingCommitStatus && post.counts.Success < post.counts.Total {
			regressed = true
		}
	}
	if !sawFullSuccess || regressed {
		t.Fatalf("combined apply status regressed after a complete success, database has both projects applied, posts=%+v", posts)
	}
}

// TestWorkingDirLocker_StressAutoplanWaitersAndSharedApplies exercises the waiter and shared-apply paths.
// It is expected to pass; a leftover lock or a blocked waiter is a failure.
func TestWorkingDirLocker_StressAutoplanWaitersAndSharedApplies(t *testing.T) {
	const repo = "owner/repo"
	for i := 0; i < 40; i++ {
		locker := NewDefaultWorkingDirLocker()
		unlock, err := locker.TryLockPull(repo, 1, command.Plan, WorkingDirLockMetadata{HeadCommit: "base"})
		Ok(t, err)

		var wg sync.WaitGroup
		wg.Add(8)
		for n := 0; n < 8; n++ {
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
