// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"slices"
	"testing"
	"time"

	"github.com/google/go-github/v88/github"

	"github.com/runatlantis/atlantis/server/core/boltdb"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/models/testdata"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics/metricstest"
)

// These tests exercise persistence and status ordering with successful comments.
// Unexpected VCS operations still use the failing embedded client.
type internalApplyVCSClient struct {
	vcs.NotConfiguredVCSClient
}

func (*internalApplyVCSClient) CreateComment(logging.SimpleLogging, models.Repo, int, string, string) error {
	return nil
}

type unlockedApplyLockChecker struct{}

func (unlockedApplyLockChecker) CheckApplyLock() (locking.ApplyCommandLock, error) {
	return locking.ApplyCommandLock{}, nil
}

type noopPullReqStatusFetcher struct{}

func (noopPullReqStatusFetcher) FetchPullStatus(logging.SimpleLogging, models.PullRequest) (models.PullReqStatus, error) {
	return models.PullReqStatus{}, nil
}

type recordingApplyStatusUpdater struct {
	pendingError  error
	combined      []models.CommitStatus
	combinedCount []models.CommitStatus
}

func (u *recordingApplyStatusUpdater) UpdateCombined(_ logging.SimpleLogging, _ models.Repo, _ models.PullRequest, status models.CommitStatus, _ command.Name) error {
	u.combined = append(u.combined, status)
	if status == models.PendingCommitStatus {
		return u.pendingError
	}
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

func (r *recordingDeferredApplyRunner) PublishDeferredApplyStatuses(_ []command.ProjectContext, _ command.Result, status models.CommitStatus) error {
	r.statuses = append(r.statuses, status)
	return nil
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
	_, err = database.UpdatePullWithResults(pull, []command.ProjectResult{internalPlannedProjectResult("dirA", DefaultWorkspace, "projA")}, command.NoClaim{})
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
	_, err = database.UpdatePullWithResults(initialPull, []command.ProjectResult{internalPlannedProjectResult("dirA", DefaultWorkspace, "projA")}, command.NoClaim{})
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
	vcsClient := &internalApplyVCSClient{NotConfiguredVCSClient: vcs.NotConfiguredVCSClient{Host: models.Github}}
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
	vcsClient := &internalApplyVCSClient{NotConfiguredVCSClient: vcs.NotConfiguredVCSClient{Host: models.Github}}
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
	if _, err := database.UpdatePullWithResults(legacyPull, nil, command.NoClaim{}); err != nil {
		t.Fatal(err)
	}
	vcsClient := &internalApplyVCSClient{NotConfiguredVCSClient: vcs.NotConfiguredVCSClient{Host: models.Github}}
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
	if _, err := database.UpdatePullWithResults(legacyPull, nil, command.NoClaim{}); err != nil {
		t.Fatal(err)
	}
	vcsClient := &internalApplyVCSClient{NotConfiguredVCSClient: vcs.NotConfiguredVCSClient{Host: models.Github}}
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

type publicationApplyRunner struct {
	run func(command.ProjectContext) command.ProjectCommandOutput
}

func (r publicationApplyRunner) Apply(ctx command.ProjectContext) command.ProjectCommandOutput {
	return r.run(ctx)
}

func TestApplyCommandRunner_ExecutionRunsOutsideLeaseAndCannotRepeat(t *testing.T) {
	for _, outcome := range []string{"success", "pre-execution failure", "unknown execution outcome"} {
		t.Run(outcome, func(t *testing.T) {
			database, err := boltdb.New(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { database.Close() })
			pull := testdata.Pull
			pull.BaseRepo = testdata.GithubRepo
			_, err = database.UpdatePullWithResults(pull, []command.ProjectResult{internalPlannedProjectResult("dirA", DefaultWorkspace, "projA")}, command.NoClaim{})
			if err != nil {
				t.Fatal(err)
			}
			project := command.ProjectContext{CommandName: command.Apply, RepoRelDir: "dirA", Workspace: DefaultWorkspace, ProjectName: "projA", ProjectPlanStatus: models.PlannedPlanStatus, Pull: pull}
			clock := newPublicationTestClock()
			publication := NewPublicationCoordinator(database, context.Background())
			publication.Clock = clock
			executed := false
			projectRunner := publicationApplyRunner{run: func(ctx command.ProjectContext) command.ProjectCommandOutput {
				executed = true
				if ctx.ApplyExecutionID == "" {
					t.Fatal("apply must have a durable execution identity")
				}
				// Time spent in Terraform does not require lease renewal. A different
				// replica can publish, but cannot execute this same plan again.
				clock.advance(10 * publication.LeaseTTL)
				_, err := database.AcquirePublicationLease(context.Background(), pull, "another replica", time.Minute)
				if err != nil {
					t.Fatalf("Terraform must not hold publication lease: %v", err)
				}
				fence := command.PublicationFence{Owner: "another replica"}
				_, err = database.BeginApplyExecution(pull, []command.ProjectContext{project}, "duplicate apply", fence)
				if !errors.Is(err, db.ErrApplyAlreadyStarted) {
					t.Fatalf("duplicate apply must fail closed: %v", err)
				}
				if err := database.ReleasePublicationLease(context.Background(), pull, fence); err != nil {
					t.Fatal(err)
				}
				switch outcome {
				case "pre-execution failure":
					return command.ProjectCommandOutput{Error: errors.New("validation rejected")}
				case "unknown execution outcome":
					return command.ProjectCommandOutput{ApplyAttempted: true, Error: errors.New("execution connection lost")}
				default:
					return command.ProjectCommandOutput{ApplyAttempted: true, ApplyExecuted: true, ApplySuccess: "applied"}
				}
			}}
			runner := newInternalApplyCommandRunner(t, database, staticApplyCommandBuilder{commands: []command.ProjectContext{project}}, projectRunner, &sequenceApplyIdentityFetcher{identities: []models.PullRequest{pull}})
			runner.Publication = publication
			ctx := newInternalApplyContext(t, pull)
			runner.Run(ctx, &CommentCommand{Name: command.Apply, ProjectName: "projA"})
			if !executed {
				t.Fatal("expected apply runner to execute")
			}
			if ctx.CommandHasErrors != (outcome != "success") {
				t.Fatalf("unexpected command error flag for %s", outcome)
			}
			status, err := database.GetPullStatus(pull)
			if err != nil {
				t.Fatal(err)
			}
			if (status.Projects[0].ApplyExecutionID != "") != (outcome == "unknown execution outcome") {
				t.Fatalf("unexpected execution ownership after %s: %+v", outcome, status.Projects[0])
			}
			if outcome == "success" && status.Projects[0].Status != models.AppliedPlanStatus {
				t.Fatal("successful apply must be durably applied")
			}
			_, err = database.AcquirePublicationLease(context.Background(), pull, "after completion", time.Minute)
			if err != nil {
				t.Fatalf("completion must release publication lease: %v", err)
			}
		})
	}
}

func TestApplyCommandRunner_AdmissionBeforePendingAndPendingFailureReleases(t *testing.T) {
	for _, reason := range []string{"obsolete", "already executing", "pending failure", "unknown pending outcome"} {
		t.Run(reason, func(t *testing.T) {
			database, err := boltdb.New(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { database.Close() })
			pull := testdata.Pull
			pull.BaseRepo = testdata.GithubRepo
			_, err = database.UpdatePullWithResults(pull, []command.ProjectResult{internalPlannedProjectResult("dirA", DefaultWorkspace, "projA")}, command.NoClaim{})
			if err != nil {
				t.Fatal(err)
			}
			project := command.ProjectContext{CommandName: command.Apply, RepoRelDir: "dirA", Workspace: DefaultWorkspace, ProjectName: "projA", ProjectPlanStatus: models.PlannedPlanStatus, Pull: pull}
			if reason == "obsolete" {
				project.PlanGeneration = "obsolete"
			}
			if reason == "already executing" {
				_, err = database.BeginApplyExecution(pull, []command.ProjectContext{project}, "running", command.NoClaim{})
				if err != nil {
					t.Fatal(err)
				}
			}
			runner := newInternalApplyCommandRunner(t, database, staticApplyCommandBuilder{commands: []command.ProjectContext{project}}, publicationApplyRunner{run: func(command.ProjectContext) command.ProjectCommandOutput {
				t.Error("rejected admission must not execute Terraform")
				return command.ProjectCommandOutput{}
			}}, &sequenceApplyIdentityFetcher{identities: []models.PullRequest{pull}})
			runner.Publication = NewPublicationCoordinator(database, context.Background())
			statuses := &recordingApplyStatusUpdater{}
			if reason == "pending failure" {
				statuses.pendingError = &github.ErrorResponse{Response: &http.Response{StatusCode: http.StatusTooManyRequests}, Message: "temporary pending status rejection"}
			}
			if reason == "unknown pending outcome" {
				statuses.pendingError = errors.New("request timed out after sending pending status")
			}
			runner.commitStatusUpdater = statuses
			ctx := newInternalApplyContext(t, pull)
			runner.Run(ctx, &CommentCommand{Name: command.Apply, ProjectName: "projA"})
			if !ctx.CommandHasErrors {
				t.Fatal("rejected apply must be visible as an error")
			}
			if containsCommitStatus(statuses.combined, models.PendingCommitStatus) != (reason == "pending failure" || reason == "unknown pending outcome") {
				t.Fatalf("an obsolete/consumed command must not overwrite Pending: %v", statuses.combined)
			}
			status, err := database.GetPullStatus(pull)
			if err != nil {
				t.Fatal(err)
			}
			expectedExecution := ""
			if reason == "already executing" {
				expectedExecution = "running"
			}
			if status.Projects[0].ApplyExecutionID != expectedExecution {
				t.Fatal("pre-execution failure must not leave its own execution reservation")
			}
			if reason == "unknown pending outcome" {
				lease, err := database.GetPublicationLease(context.Background(), pull)
				if err != nil || lease == nil || !lease.Publishing {
					t.Fatalf("unknown remote status must remain quarantined: %+v, %v", lease, err)
				}
				return
			}
			_, err = database.AcquirePublicationLease(context.Background(), pull, "next command", time.Minute)
			if err != nil {
				t.Fatalf("pre-execution error must release lease: %v", err)
			}
		})
	}
}
