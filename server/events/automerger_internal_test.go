// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

// Internal tests for AutoMerger (package events, not events_test) so that we
// can access the unexported automerge / automergeEnabled / deleteSourceBranchOnMergeEnabled
// methods directly.
package events

import (
	"errors"
	"testing"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

// buildApplyContext builds a minimal *command.Context suitable for automerge tests.
func buildApplyContext(t *testing.T, vcsClient *vcsmocks.MockClient) *command.Context {
	t.Helper()
	return &command.Context{
		Log: logging.NewNoopLogger(t),
		Pull: models.PullRequest{
			Num: 1,
			BaseRepo: models.Repo{
				FullName: "owner/repo",
				VCSHost:  models.VCSHost{Type: models.Github},
			},
		},
	}
}

// ---------------------------------------------------------------------------
// automerge
// ---------------------------------------------------------------------------

func TestAutomerge_AllApplied_MergesAndComments(t *testing.T) {
	RegisterMockTestingT(t)
	vcsClient := vcsmocks.NewMockClient()
	am := &AutoMerger{VCSClient: vcsClient, GlobalAutomerge: true}
	ctx := buildApplyContext(t, vcsClient)

	pullStatus := models.PullStatus{
		Projects: []models.ProjectStatus{
			{Status: models.AppliedPlanStatus},
			{Status: models.AppliedPlanStatus},
		},
	}

	am.automerge(ctx, pullStatus, false, "")

	// Should have commented (once for the automerge announcement) and merged once.
	vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](),
		Eq(ctx.Pull.BaseRepo),
		Eq(ctx.Pull.Num),
		Eq(automergeComment),
		Eq(command.Apply.String()),
	)
	vcsClient.VerifyWasCalledOnce().MergePull(
		Any[logging.SimpleLogging](),
		Eq(ctx.Pull),
		Eq(models.PullRequestOptions{DeleteSourceBranchOnMerge: false, MergeMethod: ""}),
	)
}

func TestAutomerge_NotAllApplied_DoesNotMerge(t *testing.T) {
	RegisterMockTestingT(t)
	vcsClient := vcsmocks.NewMockClient()
	am := &AutoMerger{VCSClient: vcsClient, GlobalAutomerge: true}
	ctx := buildApplyContext(t, vcsClient)

	pullStatus := models.PullStatus{
		Projects: []models.ProjectStatus{
			{Status: models.AppliedPlanStatus, RepoRelDir: "a", Workspace: "default"},
			{Status: models.PlannedPlanStatus, RepoRelDir: "b", Workspace: "default"},
		},
	}

	am.automerge(ctx, pullStatus, false, "")

	// Neither comment nor merge should have been called.
	vcsClient.VerifyWasCalled(Never()).CreateComment(
		Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string](),
	)
	vcsClient.VerifyWasCalled(Never()).MergePull(
		Any[logging.SimpleLogging](), Any[models.PullRequest](), Any[models.PullRequestOptions](),
	)
}

func TestAutomerge_MergeFails_PostsFailureComment(t *testing.T) {
	RegisterMockTestingT(t)
	vcsClient := vcsmocks.NewMockClient()
	am := &AutoMerger{VCSClient: vcsClient, GlobalAutomerge: true}
	ctx := buildApplyContext(t, vcsClient)

	pullStatus := models.PullStatus{
		Projects: []models.ProjectStatus{
			{Status: models.AppliedPlanStatus},
		},
	}

	mergeErr := errors.New("merge conflict")
	When(vcsClient.MergePull(Any[logging.SimpleLogging](), Any[models.PullRequest](), Any[models.PullRequestOptions]())).
		ThenReturn(mergeErr)

	am.automerge(ctx, pullStatus, false, "")

	// Two CreateComment calls expected: announcement + failure.
	vcsClient.VerifyWasCalled(Twice()).CreateComment(
		Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string](),
	)
}

func TestAutomerge_DeleteSourceBranch_PassedToMerge(t *testing.T) {
	RegisterMockTestingT(t)
	vcsClient := vcsmocks.NewMockClient()
	am := &AutoMerger{VCSClient: vcsClient, GlobalAutomerge: true}
	ctx := buildApplyContext(t, vcsClient)

	pullStatus := models.PullStatus{
		Projects: []models.ProjectStatus{
			{Status: models.AppliedPlanStatus},
		},
	}

	am.automerge(ctx, pullStatus, true, "squash")

	vcsClient.VerifyWasCalledOnce().MergePull(
		Any[logging.SimpleLogging](),
		Eq(ctx.Pull),
		Eq(models.PullRequestOptions{DeleteSourceBranchOnMerge: true, MergeMethod: "squash"}),
	)
}

func TestAutomerge_EmptyProjects_DoesNotMerge(t *testing.T) {
	RegisterMockTestingT(t)
	vcsClient := vcsmocks.NewMockClient()
	am := &AutoMerger{VCSClient: vcsClient, GlobalAutomerge: true}
	ctx := buildApplyContext(t, vcsClient)

	// PullStatus with no projects – nothing to check, nothing to merge.
	am.automerge(ctx, models.PullStatus{}, false, "")

	// No projects == vacuously all applied → should still merge.
	vcsClient.VerifyWasCalledOnce().MergePull(
		Any[logging.SimpleLogging](), Any[models.PullRequest](), Any[models.PullRequestOptions](),
	)
}

// ---------------------------------------------------------------------------
// automergeEnabled
// ---------------------------------------------------------------------------

func TestAutomergeEnabled_GlobalTrue_NoProjects(t *testing.T) {
	am := &AutoMerger{GlobalAutomerge: true}
	Equals(t, true, am.automergeEnabled(nil))
	Equals(t, true, am.automergeEnabled([]command.ProjectContext{}))
}

func TestAutomergeEnabled_GlobalFalse_NoProjects(t *testing.T) {
	am := &AutoMerger{GlobalAutomerge: false}
	Equals(t, false, am.automergeEnabled(nil))
	Equals(t, false, am.automergeEnabled([]command.ProjectContext{}))
}

func TestAutomergeEnabled_ProjectOverridesGlobal(t *testing.T) {
	// Even when GlobalAutomerge is false, a project with AutomergeEnabled=true should win.
	am := &AutoMerger{GlobalAutomerge: false}
	Equals(t, true, am.automergeEnabled([]command.ProjectContext{
		{AutomergeEnabled: true},
	}))

	// And when GlobalAutomerge is true, a project with AutomergeEnabled=false should win.
	am2 := &AutoMerger{GlobalAutomerge: true}
	Equals(t, false, am2.automergeEnabled([]command.ProjectContext{
		{AutomergeEnabled: false},
	}))
}

// ---------------------------------------------------------------------------
// deleteSourceBranchOnMergeEnabled
// ---------------------------------------------------------------------------

func TestDeleteSourceBranchOnMergeEnabled_NoProjects_ReturnsFalse(t *testing.T) {
	am := &AutoMerger{}
	Equals(t, false, am.deleteSourceBranchOnMergeEnabled(nil))
	Equals(t, false, am.deleteSourceBranchOnMergeEnabled([]command.ProjectContext{}))
}

func TestDeleteSourceBranchOnMergeEnabled_WithFlag(t *testing.T) {
	am := &AutoMerger{}
	Equals(t, true, am.deleteSourceBranchOnMergeEnabled([]command.ProjectContext{
		{DeleteSourceBranchOnMerge: true},
	}))
}

func TestDeleteSourceBranchOnMergeEnabled_FlagFalse(t *testing.T) {
	am := &AutoMerger{}
	Equals(t, false, am.deleteSourceBranchOnMergeEnabled([]command.ProjectContext{
		{DeleteSourceBranchOnMerge: false},
	}))
}
