// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"fmt"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
)

type AutoMerger struct {
	VCSClient       vcs.Client
	GlobalAutomerge bool
	VCSStatusName   string // VCS status name prefix (e.g., "atlantis")
}

func (c *AutoMerger) automerge(ctx *command.Context, pullStatus models.PullStatus, deleteSourceBranchOnMerge bool, mergeMethod string) {
	// We only automerge if all projects have been successfully applied.
	for _, p := range pullStatus.Projects {
		if p.Status != models.AppliedPlanStatus {
			ctx.Log.Info("not automerging because project at dir %q, workspace %q has status %q", p.RepoRelDir, p.Workspace, p.Status.String())
			return
		}
	}

	// Check bypass merge validation if the VCS client supports it.
	// This validates team permissions and generates audit comments for bypass merges.
	if bypassChecker, ok := c.VCSClient.(vcs.BypassMergeChecker); ok {
		allowed, auditMessage, err := bypassChecker.ValidateBypassMerge(
			ctx.Log,
			ctx.Pull.BaseRepo,
			ctx.Pull,
			ctx.User,
			c.VCSStatusName,
		)
		if err != nil {
			ctx.Log.Err("bypass merge validation failed: %s", err)
			failureComment := fmt.Sprintf("**Automerge blocked**: Failed to validate bypass merge permissions:\n```\n%s\n```", err)
			if commentErr := c.VCSClient.CreateComment(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull.Num, failureComment, command.Apply.String()); commentErr != nil {
				ctx.Log.Err("failed to comment about bypass validation failure: %s", commentErr)
			}
			return
		}

		if !allowed {
			ctx.Log.Info("user %s is not authorized to perform bypass merge", ctx.User.Username)
			notAllowedComment := fmt.Sprintf(":no_entry: **Automerge blocked**: User **@%s** is not authorized to merge with bypass.\n\n"+
				"The **%s/apply** status check is not passing, and bypass merge is restricted to specific teams.\n\n"+
				"Please contact a team member with bypass permissions or ensure all apply operations complete successfully.",
				ctx.User.Username, c.VCSStatusName)
			if commentErr := c.VCSClient.CreateComment(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull.Num, notAllowedComment, command.Apply.String()); commentErr != nil {
				ctx.Log.Err("failed to comment about bypass not allowed: %s", commentErr)
			}
			return
		}

		// Add audit comment if provided
		if auditMessage != "" {
			if commentErr := c.VCSClient.CreateComment(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull.Num, auditMessage, command.Apply.String()); commentErr != nil {
				ctx.Log.Err("failed to add bypass audit comment: %s", commentErr)
				// Continue with merge even if audit comment fails
			}
		}
	}

	// Comment that we're automerging the pull request.
	if err := c.VCSClient.CreateComment(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull.Num, automergeComment, command.Apply.String()); err != nil {
		ctx.Log.Err("failed to comment about automerge: %s", err)
		// Commenting isn't required so continue.
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
