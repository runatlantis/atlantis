// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"fmt"
	"slices"
	"testing"

	"github.com/runatlantis/atlantis/server/core/boltdb"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/models/testdata"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics/metricstest"
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
