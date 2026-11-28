// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"errors"
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/core/boltdb"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/models/testdata"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics/metricstest"
	. "github.com/runatlantis/atlantis/testing"
)

type unlockedApplyLockChecker struct{}

func (unlockedApplyLockChecker) CheckApplyLock() (locking.ApplyCommandLock, error) {
	return locking.ApplyCommandLock{}, nil
}

type noopPullReqStatusFetcher struct{}

func (noopPullReqStatusFetcher) FetchPullStatus(logging.SimpleLogging, models.PullRequest) (models.PullReqStatus, error) {
	return models.PullReqStatus{}, nil
}

type recordingApplyStatusUpdater struct {
	combined      []models.CommitStatus
	combinedCount []models.CommitStatus
}

func (u *recordingApplyStatusUpdater) UpdateCombined(_ logging.SimpleLogging, _ models.Repo, _ models.PullRequest, status models.CommitStatus, _ command.Name) error {
	u.combined = append(u.combined, status)
	return nil
}

func (u *recordingApplyStatusUpdater) UpdateCombinedCount(_ logging.SimpleLogging, _ models.Repo, _ models.PullRequest, status models.CommitStatus, _ command.Name, _ models.ProjectCounts) error {
	u.combinedCount = append(u.combinedCount, status)
	return nil
}

func (*recordingApplyStatusUpdater) UpdatePreWorkflowHook(logging.SimpleLogging, models.PullRequest, models.CommitStatus, string, string, string) error {
	return nil
}

func (*recordingApplyStatusUpdater) UpdatePostWorkflowHook(logging.SimpleLogging, models.PullRequest, models.CommitStatus, string, string, string) error {
	return nil
}

type staticApplyCommandBuilder struct {
	commands []command.ProjectContext
}

func (b staticApplyCommandBuilder) BuildApplyCommands(*command.Context, *CommentCommand) ([]command.ProjectContext, error) {
	return b.commands, nil
}

func (staticApplyCommandBuilder) ShouldIgnoreTargetedDir(*command.Context, *CommentCommand) bool {
	return false
}

type staticProjectApplyRunner struct {
	output command.ProjectCommandOutput
}

func (r staticProjectApplyRunner) Apply(command.ProjectContext) command.ProjectCommandOutput {
	return r.output
}

type recordingDeferredApplyRunner struct {
	output   command.ProjectCommandOutput
	statuses []models.CommitStatus
}

func (r *recordingDeferredApplyRunner) Apply(command.ProjectContext) command.ProjectCommandOutput {
	return r.output
}

func (r *recordingDeferredApplyRunner) PublishDeferredApplyStatuses(_ []command.ProjectContext, _ command.Result, status models.CommitStatus) {
	r.statuses = append(r.statuses, status)
}

type sequenceApplyIdentityFetcher struct {
	identities []models.PullRequest
	calls      int
}

func (f *sequenceApplyIdentityFetcher) GetLivePullIdentity(command.ProjectContext) (models.PullRequest, error) {
	if len(f.identities) == 0 {
		return models.PullRequest{}, nil
	}
	idx := f.calls
	if idx >= len(f.identities) {
		idx = len(f.identities) - 1
	}
	f.calls++
	return f.identities[idx], nil
}

func TestApplyCommandRunner_DeferredApplySuccessPublishesAfterFinalFreshness(t *testing.T) {
	database, err := boltdb.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	pull := models.PullRequest{
		BaseRepo:   testdata.GithubRepo,
		State:      models.OpenPullState,
		Num:        testdata.Pull.Num,
		HeadCommit: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		BaseBranch: "main",
	}
	_, err = database.UpdatePullWithResults(pull, []command.ProjectResult{internalPlannedProjectResult("dirA", DefaultWorkspace, "projA")})
	if err != nil {
		t.Fatal(err)
	}
	deferredRunner := &recordingDeferredApplyRunner{output: command.ProjectCommandOutput{
		ApplySuccess:    "applied",
		ApplySuccessURL: "https://app.terraform.io/app/org/workspace/runs/run-123",
	}}
	runner := newInternalApplyCommandRunner(t, database, staticApplyCommandBuilder{commands: []command.ProjectContext{{
		CommandName:       command.Apply,
		RepoRelDir:        "dirA",
		Workspace:         DefaultWorkspace,
		ProjectName:       "projA",
		ProjectPlanStatus: models.PlannedPlanStatus,
		Pull:              pull,
	}}}, deferredRunner, &sequenceApplyIdentityFetcher{identities: []models.PullRequest{pull, pull}})
	ctx := newInternalApplyContext(t, pull)

	runner.Run(ctx, &CommentCommand{Name: command.Apply, ProjectName: "projA"})

	if ctx.CommandHasErrors {
		t.Fatal("expected current apply to succeed")
	}
	if !containsCommitStatus(deferredRunner.statuses, models.SuccessCommitStatus) {
		t.Fatalf("expected deferred project apply success after final freshness, got %v", deferredRunner.statuses)
	}
	if containsCommitStatus(deferredRunner.statuses, models.FailedCommitStatus) {
		t.Fatalf("did not expect deferred failed project status for current apply, got %v", deferredRunner.statuses)
	}
}

func TestApplyCommandRunner_DeferredApplySuccessFailsWhenFinalFreshnessFails(t *testing.T) {
	database, err := boltdb.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	initialPull := models.PullRequest{
		BaseRepo:   testdata.GithubRepo,
		State:      models.OpenPullState,
		Num:        testdata.Pull.Num,
		HeadCommit: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		BaseBranch: "main",
	}
	finalPull := initialPull
	finalPull.BaseBranch = "release"
	_, err = database.UpdatePullWithResults(initialPull, []command.ProjectResult{internalPlannedProjectResult("dirA", DefaultWorkspace, "projA")})
	if err != nil {
		t.Fatal(err)
	}
	deferredRunner := &recordingDeferredApplyRunner{output: command.ProjectCommandOutput{
		ApplySuccess:    "applied",
		ApplySuccessURL: "https://app.terraform.io/app/org/workspace/runs/run-123",
	}}
	runner := newInternalApplyCommandRunner(t, database, staticApplyCommandBuilder{commands: []command.ProjectContext{{
		CommandName:       command.Apply,
		RepoRelDir:        "dirA",
		Workspace:         DefaultWorkspace,
		ProjectName:       "projA",
		ProjectPlanStatus: models.PlannedPlanStatus,
		Pull:              initialPull,
	}}}, deferredRunner, &sequenceApplyIdentityFetcher{identities: []models.PullRequest{initialPull, finalPull}})
	ctx := newInternalApplyContext(t, initialPull)

	runner.Run(ctx, &CommentCommand{Name: command.Apply, ProjectName: "projA"})

	if !ctx.CommandHasErrors {
		t.Fatal("expected final freshness failure to mark the command as errored")
	}
	if containsCommitStatus(deferredRunner.statuses, models.SuccessCommitStatus) {
		t.Fatalf("expected no deferred project apply success for stale final identity, got %v", deferredRunner.statuses)
	}
	if !containsCommitStatus(deferredRunner.statuses, models.FailedCommitStatus) {
		t.Fatalf("expected deferred project apply failure for stale final identity, got %v", deferredRunner.statuses)
	}
}

func newInternalApplyCommandRunner(t *testing.T, database *boltdb.BoltDB, builder ProjectApplyCommandBuilder, projectRunner ProjectApplyCommandRunner, liveFetcher LivePullHeadFetcher) *ApplyCommandRunner {
	t.Helper()
	vcsClient := &vcs.NotConfiguredVCSClient{Host: models.Github}
	pullUpdater := &PullUpdater{
		VCSClient:        vcsClient,
		MarkdownRenderer: NewMarkdownRenderer(false, false, false, false, false, false, "", "atlantis", false, false),
	}
	return NewApplyCommandRunner(
		vcsClient,
		false,
		unlockedApplyLockChecker{},
		&recordingApplyStatusUpdater{},
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
		nil,
		noopPullReqStatusFetcher{},
		liveFetcher,
		"",
	)
}

func newInternalApplyContext(t *testing.T, pull models.PullRequest) *command.Context {
	t.Helper()
	return &command.Context{
		User:     testdata.User,
		Log:      logging.NewNoopLogger(t),
		Scope:    metricstest.NewLoggingScope(t, logging.NewNoopLogger(t), "atlantis"),
		Pull:     pull,
		HeadRepo: testdata.GithubRepo,
		Trigger:  command.CommentTrigger,
	}
}

func internalPlannedProjectResult(repoRelDir, workspace, projectName string) command.ProjectResult {
	return command.ProjectResult{
		Command:     command.Plan,
		RepoRelDir:  repoRelDir,
		Workspace:   workspace,
		ProjectName: projectName,
		ProjectCommandOutput: command.ProjectCommandOutput{
			PlanSuccess: &models.PlanSuccess{TerraformOutput: "Plan: 1 to add, 0 to change, 0 to destroy."},
		},
	}
}

func TestApplyCommandRunner_StaleCommandResultWithEmptyPullStatusDoesNotPublishZeroZeroSuccess(t *testing.T) {
	database, err := boltdb.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	vcsClient := &vcs.NotConfiguredVCSClient{Host: models.Github}
	commitUpdater := &recordingApplyStatusUpdater{}
	pullUpdater := &PullUpdater{
		VCSClient:        vcsClient,
		MarkdownRenderer: NewMarkdownRenderer(false, false, false, false, false, false, "", "atlantis", false, false),
	}
	runner := NewApplyCommandRunner(
		vcsClient,
		false,
		unlockedApplyLockChecker{},
		commitUpdater,
		staticApplyCommandBuilder{commands: []command.ProjectContext{{
			CommandName: command.Apply,
			RepoRelDir:  "dirA",
			Workspace:   DefaultWorkspace,
			ProjectName: "projA",
			Pull: models.PullRequest{
				BaseRepo:   testdata.GithubRepo,
				State:      models.OpenPullState,
				Num:        testdata.Pull.Num,
				HeadCommit: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
				BaseBranch: "main",
			},
		}}},
		staticProjectApplyRunner{output: command.ProjectCommandOutput{
			Error: fmt.Errorf("%w: pull request base branch changed", errStaleCommandHead),
		}},
		NewCancellationTracker(),
		&AutoMerger{VCSClient: vcsClient},
		pullUpdater,
		&DBUpdater{Database: database},
		database,
		1,
		false,
		false,
		nil,
		noopPullReqStatusFetcher{},
		nil,
		"",
	)
	pull := models.PullRequest{
		BaseRepo:   testdata.GithubRepo,
		State:      models.OpenPullState,
		Num:        testdata.Pull.Num,
		HeadCommit: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		BaseBranch: "main",
	}
	cmd := &CommentCommand{Name: command.Apply, ProjectName: "projA"}
	ctx := &command.Context{
		User:     testdata.User,
		Log:      logging.NewNoopLogger(t),
		Scope:    metricstest.NewLoggingScope(t, logging.NewNoopLogger(t), "atlantis"),
		Pull:     pull,
		HeadRepo: testdata.GithubRepo,
		Trigger:  command.CommentTrigger,
	}

	runner.Run(ctx, cmd)

	if !ctx.CommandHasErrors {
		t.Fatal("expected stale apply to mark the command as errored")
	}
	if containsCommitStatus(commitUpdater.combinedCount, models.SuccessCommitStatus) {
		t.Fatal("expected no successful apply count status for stale command result")
	}
	if containsCommitStatus(commitUpdater.combined, models.SuccessCommitStatus) {
		t.Fatal("expected no successful apply status for stale command result")
	}
	if !containsCommitStatus(commitUpdater.combined, models.FailedCommitStatus) {
		t.Fatal("expected failed apply status for stale command result")
	}
}

func TestApplyCommandRunner_NoPlanLegacyEmptyIdentityDoesNotPublishZeroZeroSuccess(t *testing.T) {
	database, err := boltdb.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	legacyPull := models.PullRequest{
		BaseRepo: testdata.GithubRepo,
		State:    models.OpenPullState,
		Num:      testdata.Pull.Num,
	}
	if _, err := database.UpdatePullWithResults(legacyPull, nil); err != nil {
		t.Fatal(err)
	}
	vcsClient := &vcs.NotConfiguredVCSClient{Host: models.Github}
	commitUpdater := &recordingApplyStatusUpdater{}
	pullUpdater := &PullUpdater{
		VCSClient:        vcsClient,
		MarkdownRenderer: NewMarkdownRenderer(false, false, false, false, false, false, "", "atlantis", false, false),
	}
	runner := NewApplyCommandRunner(
		vcsClient,
		false,
		unlockedApplyLockChecker{},
		commitUpdater,
		staticApplyCommandBuilder{},
		staticProjectApplyRunner{},
		NewCancellationTracker(),
		&AutoMerger{VCSClient: vcsClient},
		pullUpdater,
		&DBUpdater{Database: database},
		database,
		1,
		false,
		false,
		nil,
		noopPullReqStatusFetcher{},
		nil,
		"",
	)
	pull := models.PullRequest{
		BaseRepo:   testdata.GithubRepo,
		State:      models.OpenPullState,
		Num:        testdata.Pull.Num,
		HeadCommit: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		BaseBranch: "release",
	}
	ctx := &command.Context{
		User:     testdata.User,
		Log:      logging.NewNoopLogger(t),
		Scope:    metricstest.NewLoggingScope(t, logging.NewNoopLogger(t), "atlantis"),
		Pull:     pull,
		HeadRepo: testdata.GithubRepo,
		Trigger:  command.CommentTrigger,
	}

	runner.Run(ctx, &CommentCommand{Name: command.Apply})

	if !ctx.CommandHasErrors {
		t.Fatal("expected legacy empty identity apply to mark the command as errored")
	}
	if containsCommitStatus(commitUpdater.combinedCount, models.SuccessCommitStatus) {
		t.Fatal("expected no successful 0/0 apply count status for legacy empty identity")
	}
	if containsCommitStatus(commitUpdater.combined, models.SuccessCommitStatus) {
		t.Fatal("expected no successful apply status for legacy empty identity")
	}
	if !containsCommitStatus(commitUpdater.combined, models.FailedCommitStatus) {
		t.Fatal("expected failed apply status for legacy empty identity")
	}
}

func TestApplyCommandRunner_SilencedNoProjectLegacyEmptyIdentityDoesNotPublishZeroZeroSuccess(t *testing.T) {
	database, err := boltdb.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	legacyPull := models.PullRequest{
		BaseRepo: testdata.GithubRepo,
		State:    models.OpenPullState,
		Num:      testdata.Pull.Num,
	}
	if _, err := database.UpdatePullWithResults(legacyPull, nil); err != nil {
		t.Fatal(err)
	}
	vcsClient := &vcs.NotConfiguredVCSClient{Host: models.Github}
	commitUpdater := &recordingApplyStatusUpdater{}
	pullUpdater := &PullUpdater{
		VCSClient:        vcsClient,
		MarkdownRenderer: NewMarkdownRenderer(false, false, false, false, false, false, "", "atlantis", false, false),
	}
	runner := NewApplyCommandRunner(
		vcsClient,
		false,
		unlockedApplyLockChecker{},
		commitUpdater,
		staticApplyCommandBuilder{},
		staticProjectApplyRunner{},
		NewCancellationTracker(),
		&AutoMerger{VCSClient: vcsClient},
		pullUpdater,
		&DBUpdater{Database: database},
		database,
		1,
		true,
		false,
		nil,
		noopPullReqStatusFetcher{},
		nil,
		"",
	)
	pull := models.PullRequest{
		BaseRepo:   testdata.GithubRepo,
		State:      models.OpenPullState,
		Num:        testdata.Pull.Num,
		HeadCommit: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		BaseBranch: "release",
	}
	ctx := &command.Context{
		User:     testdata.User,
		Log:      logging.NewNoopLogger(t),
		Scope:    metricstest.NewLoggingScope(t, logging.NewNoopLogger(t), "atlantis"),
		Pull:     pull,
		HeadRepo: testdata.GithubRepo,
		Trigger:  command.CommentTrigger,
	}

	runner.Run(ctx, &CommentCommand{Name: command.Apply, RepoRelDir: "missing"})

	if !ctx.CommandHasErrors {
		t.Fatal("expected legacy empty identity apply to mark the command as errored")
	}
	if containsCommitStatus(commitUpdater.combinedCount, models.SuccessCommitStatus) {
		t.Fatal("expected no successful 0/0 apply count status for silenced legacy empty identity")
	}
	if containsCommitStatus(commitUpdater.combined, models.SuccessCommitStatus) {
		t.Fatal("expected no successful apply status for silenced legacy empty identity")
	}
}

func containsCommitStatus(statuses []models.CommitStatus, expected models.CommitStatus) bool {
	return slices.Contains(statuses, expected)
}

type applyCountPost struct {
	status models.CommitStatus
	counts models.ProjectCounts
}

// Holds the first 1/2 status post until the other apply has posted 2/2.
type gatedApplyStatus struct {
	mu          sync.Mutex
	posts       []applyCountPost
	partialSeen chan struct{}
	fullPosted  chan struct{}
	oncePartial sync.Once
	onceFull    sync.Once
}

func (s *gatedApplyStatus) UpdateCombined(logging.SimpleLogging, models.Repo, models.PullRequest, models.CommitStatus, command.Name) error {
	return nil
}

func (s *gatedApplyStatus) UpdateCombinedCount(_ logging.SimpleLogging, _ models.Repo, _ models.PullRequest, status models.CommitStatus, _ command.Name, counts models.ProjectCounts) error {
	partial := status == models.PendingCommitStatus && counts.Total == 2 && counts.Success == 1
	if partial {
		s.oncePartial.Do(func() { close(s.partialSeen) })
		// Let this 1/2 post land after the 2/2 post.
		select {
		case <-s.fullPosted:
		case <-time.After(300 * time.Millisecond):
		}
	}
	s.mu.Lock()
	s.posts = append(s.posts, applyCountPost{status: status, counts: counts})
	full := status == models.SuccessCommitStatus && counts.Success == counts.Total && counts.Total == 2
	s.mu.Unlock()
	if full {
		s.onceFull.Do(func() { close(s.fullPosted) })
	}
	return nil
}

func (s *gatedApplyStatus) UpdatePreWorkflowHook(logging.SimpleLogging, models.PullRequest, models.CommitStatus, string, string, string) error {
	return nil
}

func (s *gatedApplyStatus) UpdatePostWorkflowHook(logging.SimpleLogging, models.PullRequest, models.CommitStatus, string, string, string) error {
	return nil
}

func (s *gatedApplyStatus) snapshot() []applyCountPost {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]applyCountPost, len(s.posts))
	copy(out, s.posts)
	return out
}

type projectByNameBuilder struct {
	byName map[string]command.ProjectContext
}

func (b projectByNameBuilder) BuildApplyCommands(_ *command.Context, cmd *CommentCommand) ([]command.ProjectContext, error) {
	project, ok := b.byName[cmd.ProjectName]
	if !ok {
		return nil, fmt.Errorf("no project %q", cmd.ProjectName)
	}
	return []command.ProjectContext{project}, nil
}

func (projectByNameBuilder) ShouldIgnoreTargetedDir(*command.Context, *CommentCommand) bool {
	return false
}

// Starts both applies together, then holds one until the other has written its snapshot.
type phasedApplyRunner struct {
	entered  *sync.WaitGroup
	release  <-chan struct{}
	holdName string
}

func (r *phasedApplyRunner) Apply(ctx command.ProjectContext) command.ProjectCommandOutput {
	r.entered.Done()
	r.entered.Wait()
	if ctx.ProjectName == r.holdName {
		<-r.release
	}
	return command.ProjectCommandOutput{ApplySuccess: "applied"}
}

// Overlapping apply -p must not publish a 1/2 status after 2/2.
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
	statuses := &gatedApplyStatus{
		partialSeen: make(chan struct{}),
		fullPosted:  make(chan struct{}),
	}
	entered := &sync.WaitGroup{}
	entered.Add(2)
	projectRunner := &phasedApplyRunner{entered: entered, release: statuses.partialSeen, holdName: "projB"}
	vcsClient := &vcs.NotConfiguredVCSClient{Host: models.Github}
	pullUpdater := &PullUpdater{
		VCSClient:        vcsClient,
		MarkdownRenderer: NewMarkdownRenderer(false, false, false, false, false, false, "", "atlantis", false, false),
	}
	builder := projectByNameBuilder{byName: map[string]command.ProjectContext{
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
