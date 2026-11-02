// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"fmt"
	"slices"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/vcs"
)

func NewUnlockCommandRunner(
	deleteLockCommand DeleteLockCommand,
	vcsClient vcs.Client,
	SilenceNoProjects bool,
	DisableUnlockLabel string,
) *UnlockCommandRunner {
	return &UnlockCommandRunner{
		deleteLockCommand:  deleteLockCommand,
		vcsClient:          vcsClient,
		SilenceNoProjects:  SilenceNoProjects,
		DisableUnlockLabel: DisableUnlockLabel,
	}
}

type UnlockCommandRunner struct {
	vcsClient         vcs.Client
	deleteLockCommand DeleteLockCommand
	// SilenceNoProjects is whether Atlantis should respond to PRs if no projects
	// are found
	SilenceNoProjects  bool
	DisableUnlockLabel string
}

func (u *UnlockCommandRunner) Run(ctx *command.Context, cmd *CommentCommand) {
	baseRepo := ctx.Pull.BaseRepo
	pullNum := ctx.Pull.Num
	disableUnlockLabel := u.DisableUnlockLabel

	vcsMessage := "All Atlantis locks for this PR have been unlocked and plans discarded"
	if cmd.ProjectName != "" {
		ctx.Log.Info("unlocking project %q", cmd.ProjectName)
		vcsMessage = fmt.Sprintf("Atlantis locks for project %q in this PR have been unlocked and its plans discarded. Run plan again before applying.", cmd.ProjectName)
	} else {
		ctx.Log.Info("unlocking all locks")
	}

	var hasLabel bool
	var err error
	if disableUnlockLabel != "" {
		var labels []string
		labels, err = u.vcsClient.GetPullLabels(ctx.Log, baseRepo, ctx.Pull)
		if err != nil {
			vcsMessage = "Failed to retrieve PR labels... Not unlocking"
			ctx.Log.Err("Failed to retrieve PR labels for pull %s", err.Error())
		}
		hasLabel = slices.Contains(labels, disableUnlockLabel)
		if hasLabel {
			vcsMessage = "Not allowed to unlock PR with " + disableUnlockLabel + " label"
			ctx.Log.Info("Not allowed to unlock PR with %v label", disableUnlockLabel)
		}
	}

	var numLocks int
	if err == nil && !hasLabel {
		if cmd.ProjectName != "" {
			numLocks, err = u.deleteLockCommand.DeleteLocksByProject(ctx.Log, baseRepo.FullName, pullNum, cmd.ProjectName)
			if err != nil {
				vcsMessage = fmt.Sprintf("Could not unlock project %q: %s", cmd.ProjectName, err)
			}
		} else {
			numLocks, err = u.deleteLockCommand.DeleteLocksByPull(ctx.Log, baseRepo.FullName, pullNum)
			if err != nil {
				vcsMessage = "Failed to delete PR locks"
			}
		}
		if err != nil {
			ctx.Log.Err("deleting locks for pull %d - %s", pullNum, err)
		}
	}

	// if there are no locks to delete, no errors, and SilenceNoProjects is enabled, don't comment
	if err == nil && !hasLabel && numLocks == 0 {
		ctx.Log.Info("No locks to delete")
		if u.SilenceNoProjects {
			return
		}
	}

	if commentErr := u.vcsClient.CreateComment(ctx.Log, baseRepo, pullNum, vcsMessage, command.Unlock.String()); commentErr != nil {
		ctx.Log.Err("unable to comment: %s", commentErr)
	}
}
