// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
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
	Publication       *PublicationCoordinator
	vcsClient         vcs.Client
	deleteLockCommand DeleteLockCommand
	// SilenceNoProjects is whether Atlantis should respond to PRs if no projects
	// are found
	SilenceNoProjects  bool
	DisableUnlockLabel string
}

func (u *UnlockCommandRunner) Run(ctx *command.Context, _ *CommentCommand) {
	baseRepo := ctx.Pull.BaseRepo
	pullNum := ctx.Pull.Num
	disableUnlockLabel := u.DisableUnlockLabel

	ctx.Log.Info("Unlocking all locks")
	vcsMessage := "All Atlantis locks for this PR have been unlocked and plans discarded"

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

	publicationErr := u.Publication.RunCommand(ctx, func() error {
		var numLocks int
		if err == nil && !hasLabel {
			numLocks, err = u.deleteLockCommand.DeleteLocksByPull(ctx.Log, ctx.Pull, ctx.PublicationMode())
			if err != nil {
				return err
			}
		}
		if err == nil && numLocks == 0 {
			ctx.Log.Info("No locks to delete")
			if u.SilenceNoProjects {
				return nil
			}
		}
		return publishTerminal(ctx, func() error {
			return u.vcsClient.CreateComment(ctx.Log, baseRepo, pullNum, vcsMessage, command.Unlock.String())
		})
	})
	if publicationErr != nil {
		ctx.CommandHasErrors = true
		ctx.Log.Err("unlocking pull request: %v", publicationErr)
		message := "Failed to delete PR locks"
		if u.Publication != nil {
			message = "Unable to complete unlock command: " + publicationErr.Error()
		}
		if commentErr := u.vcsClient.CreateComment(ctx.Log, baseRepo, pullNum, message, command.Unlock.String()); commentErr != nil {
			ctx.Log.Err("unable to comment: %s", commentErr)
		}
	}
}
