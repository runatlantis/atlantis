package events

import (
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"slices"
)

func NewUnlockCommandRunner(
	deleteLockCommand DeleteLockCommand,
	vcsClient vcs.Client,
	SilenceNoProjects bool,
	disableUnlockLabel string,
) *UnlockCommandRunner {
	return &UnlockCommandRunner{
		deleteLockCommand:  deleteLockCommand,
		vcsClient:          vcsClient,
		SilenceNoProjects:  SilenceNoProjects,
		disableUnlockLabel: disableUnlockLabel,
	}
}

type UnlockCommandRunner struct {
	vcsClient         vcs.Client
	deleteLockCommand DeleteLockCommand
	// SilenceNoProjects is whether Atlantis should respond to PRs if no projects
	// are found
	SilenceNoProjects  bool
	disableUnlockLabel string
}

func (u *UnlockCommandRunner) Run(
	ctx *command.Context,
	cmd *CommentCommand,
) {
	baseRepo := ctx.Pull.BaseRepo
	pullNum := ctx.Pull.Num
	disableUnlockLabel := u.disableUnlockLabel

	ctx.Log.Info("Unlocking all locks")
	vcsMessage := "All Atlantis locks for this PR have been unlocked and plans discarded"

	labels, err := u.vcsClient.GetPullLabels(baseRepo, ctx.Pull)
	if err != nil {
		vcsMessage = "Failed to retrieve PR labels"
		ctx.Log.Err("faield to retrieve PR labels for pull %s", err.Error())
	}
	if slices.Contains(labels, disableUnlockLabel) {
		vcsMessage = "Not allowed to unlock PR with " + disableUnlockLabel + " label"
		ctx.Log.Info("Not allowed to unlock PR with %v label", disableUnlockLabel)
	}

	var numLocks int
	if err == nil {
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
