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
	dbUpdater *DBUpdater,
	SilenceNoProjects bool,
	DisableUnlockLabel string,
) *UnlockCommandRunner {
	return &UnlockCommandRunner{
		deleteLockCommand:  deleteLockCommand,
		vcsClient:          vcsClient,
		dbUpdater:          dbUpdater,
		SilenceNoProjects:  SilenceNoProjects,
		DisableUnlockLabel: DisableUnlockLabel,
	}
}

type UnlockCommandRunner struct {
	vcsClient         vcs.Client
	deleteLockCommand DeleteLockCommand
	dbUpdater         *DBUpdater
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

	var numLocks int
	if err == nil && !hasLabel {
		publicationToken, claimErr := u.dbUpdater.acquirePlanPublicationClaim(ctx.Pull)
		if claimErr != nil {
			ctx.CommandHasErrors = true
			ctx.Log.Err("acquiring plan publication claim for unlock: %s", claimErr)
			return
		}
		claimActive := true
		finishClaim := func(publicationErr error) {
			if !claimActive || publicationErr != nil {
				return
			}
			if releaseErr := u.dbUpdater.releasePlanPublicationClaim(ctx.Pull, publicationToken); releaseErr != nil {
				ctx.CommandHasErrors = true
				ctx.Log.Err("releasing plan publication claim for unlock: %s", releaseErr)
				return
			}
			claimActive = false
		}
		defer func() {
			if claimActive {
				ctx.Log.Warn("retaining plan publication claim after ambiguous unlock publication")
			}
		}()

		_, beginResult, replaceErr := u.dbUpdater.beginPlanGeneration(ctx.Pull, nil, true, publicationToken)
		if replaceErr != nil {
			ctx.CommandHasErrors = true
			ctx.Log.Err("discarding durable plans before unlock: %s", replaceErr)
			finishClaim(nil)
			return
		}
		ctx.PullStatus = &beginResult.PullStatus
		numLocks, err = u.deleteLockCommand.DeleteLocksByPull(ctx.Log, baseRepo.FullName, pullNum)
		if err != nil {
			vcsMessage = "Failed to delete PR locks"
			ctx.Log.Err("failed to delete locks by pull %s", err.Error())
		}

		if err == nil && numLocks == 0 {
			ctx.Log.Info("No locks to delete")
			if u.SilenceNoProjects {
				finishClaim(nil)
				return
			}
		}

		commentErr := u.vcsClient.CreateComment(ctx.Log, baseRepo, pullNum, vcsMessage, command.Unlock.String())
		if commentErr != nil {
			ctx.Log.Err("unable to comment: %s", commentErr)
		}
		finishClaim(commentErr)
		return
	}

	// if there are no locks to delete, no errors, and SilenceNoProjects is enabled, don't comment
	if err == nil && numLocks == 0 {
		ctx.Log.Info("No locks to delete")
		if u.SilenceNoProjects {
			return
		}
	}

	if commentErr := u.vcsClient.CreateComment(ctx.Log, baseRepo, pullNum, vcsMessage, command.Unlock.String()); commentErr != nil {
		ctx.Log.Err("unable to comment: %s", commentErr)
	}
}
