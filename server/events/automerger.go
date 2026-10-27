// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"errors"
	"fmt"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
)

type AutoMerger struct {
	VCSClient             vcs.Client
	GlobalAutomerge       bool
	GlobalAutomergeMethod string
	// GithubStackGetter and PullStatusFetcher are used to check that the pull
	// requests below a stacked pull request have been applied, because GitHub
	// merges them together with it.
	GithubStackGetter GithubStackGetter
	PullStatusFetcher PullStatusFetcher
}

func (c *AutoMerger) automerge(ctx *command.Context, pullStatus models.PullStatus, deleteSourceBranchOnMerge bool, mergeMethod string) {
	// We only automerge if all projects have been successfully applied.
	for _, p := range pullStatus.Projects {
		if p.Status != models.AppliedPlanStatus {
			ctx.Log.Info("not automerging because project at dir %q, workspace %q has status %q", p.RepoRelDir, p.Workspace, p.Status.String())
			return
		}
	}

	comment := automergeComment
	if ctx.Pull.Stack != nil {
		mergedWith, err := c.stackedPullsMergedWith(ctx)
		if err != nil {
			ctx.Log.Info("not automerging stacked pull request: %s", err)
			skipComment := fmt.Sprintf("Automerge skipped: %s", err)
			if commentErr := c.VCSClient.CreateComment(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull.Num, skipComment, command.Apply.String()); commentErr != nil {
				ctx.Log.Err("failed to comment about automerge being skipped: %s", commentErr)
			}
			return
		}
		if len(mergedWith) > 0 {
			comment += fmt.Sprintf(" GitHub will also merge %s below it in the stack.", formatStackedPulls(mergedWith))
		}
	}

	// Comment that we're automerging the pull request.
	if err := c.VCSClient.CreateComment(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull.Num, comment, command.Apply.String()); err != nil {
		ctx.Log.Err("failed to comment about automerge: %s", err)
		// Commenting isn't required so continue.
	}

	// Fall back to the server-side default merge method when the comment
	// command didn't specify one with --auto-merge-method.
	if mergeMethod == "" {
		mergeMethod = c.GlobalAutomergeMethod
	}

	// Make the API call to perform the merge.
	ctx.Log.Info("automerging pull request")
	var pullOptions models.PullRequestOptions
	pullOptions.DeleteSourceBranchOnMerge = deleteSourceBranchOnMerge
	pullOptions.MergeMethod = mergeMethod
	err := c.VCSClient.MergePull(ctx.Log, ctx.Pull, pullOptions)

	if err != nil {
		ctx.Log.Err("automerging failed: %s", err)

		failureComment := fmt.Sprintf("Automerging failed:\n```\n%s\n```", err)
		if commentErr := c.VCSClient.CreateComment(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull.Num, failureComment, command.Apply.String()); commentErr != nil {
			ctx.Log.Err("failed to comment about automerge failing: %s", err)
		}
	}
}

// stackedPullsMergedWith returns the open pull requests below the stacked
// pull request of ctx, which GitHub merges together with it. It returns an
// error if any of them has not been fully applied.
func (c *AutoMerger) stackedPullsMergedWith(ctx *command.Context) ([]models.StackedPull, error) {
	if ctx.Pull.Stack.Position <= 1 {
		return nil, nil
	}
	if c.GithubStackGetter == nil || c.PullStatusFetcher == nil {
		return nil, errors.New("unable to check the pull requests below this one in its stack")
	}
	stack, err := c.GithubStackGetter.GetPullRequestStack(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull.Num)
	if err != nil {
		return nil, fmt.Errorf("fetching pull request stack: %w", err)
	}
	below, err := pullsBelowInStack(stack, ctx.Pull)
	if err != nil {
		return nil, err
	}
	open := openStackedPulls(below)

	var unapplied []models.StackedPull
	for _, p := range open {
		status, err := c.PullStatusFetcher.GetPullStatus(models.PullRequest{Num: p.Num, BaseRepo: ctx.Pull.BaseRepo})
		if err != nil {
			return nil, fmt.Errorf("fetching status of pull request #%d: %w", p.Num, err)
		}
		if !isFullyAppliedAtCommit(status, p.HeadCommit) {
			unapplied = append(unapplied, p)
		}
	}
	if len(unapplied) > 0 {
		return nil, fmt.Errorf("GitHub merges stacked pull requests together with the ones below them, and %s below this one in the stack %s not been applied", formatStackedPulls(unapplied), hasOrHave(len(unapplied)))
	}
	return open, nil
}

// isFullyAppliedAtCommit returns true if Atlantis applied all projects of the
// pull request of status at commit.
func isFullyAppliedAtCommit(status *models.PullStatus, commit string) bool {
	if status == nil || status.Pull.HeadCommit != commit {
		return false
	}
	for _, p := range status.Projects {
		if p.Status != models.AppliedPlanStatus {
			return false
		}
	}
	return true
}

func hasOrHave(n int) string {
	if n == 1 {
		return "has"
	}
	return "have"
}

// automergeEnabled returns true if automerging is enabled in this context.
func (c *AutoMerger) automergeEnabled(projectCmds []command.ProjectContext) bool {
	// Use project automerge settings if projects exist; otherwise, use global automerge settings.
	automerge := c.GlobalAutomerge
	if len(projectCmds) > 0 {
		automerge = projectCmds[0].AutomergeEnabled
	}
	return automerge
}

// deleteSourceBranchOnMergeEnabled returns true if we should delete the source branch on merge in this context.
func (c *AutoMerger) deleteSourceBranchOnMergeEnabled(projectCmds []command.ProjectContext) bool {
	//check if this repo is configured for automerging.
	return (len(projectCmds) > 0 && projectCmds[0].DeleteSourceBranchOnMerge)
}
