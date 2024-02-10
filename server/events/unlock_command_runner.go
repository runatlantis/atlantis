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
		labels, err = u.vcsClient.GetPullLabels(baseRepo, ctx.Pull)
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
		numLocks, err = u.deleteLockCommand.DeleteLocksByPull(baseRepo.FullName, pullNum)
		if err != nil {
			vcsMessage = "Failed to delete PR locks"
			ctx.Log.Err("failed to delete locks by pull %s", err.Error())
		}
	}

	// if there are no locks to delete, no errors, and SilenceNoProjects is enabled, don't comment
	if err == nil && numLocks == 0 {
		ctx.Log.Info("No locks to delete")
		if u.SilenceNoProjects {
			return
		}
	}

	if commentErr := u.vcsClient.CreateComment(baseRepo, pullNum, vcsMessage, command.Unlock.String()); commentErr != nil {
		ctx.Log.Err("unable to comment: %s", commentErr)
	}
}
