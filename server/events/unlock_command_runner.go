package events

import (
	"fmt"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/vcs"
)

func NewUnlockCommandRunner(
	deleteLockCommand DeleteLockCommand,
	vcsClient vcs.Client,
) *UnlockCommandRunner {
	return &UnlockCommandRunner{
		deleteLockCommand: deleteLockCommand,
		vcsClient:         vcsClient,
	}
}

type UnlockCommandRunner struct {
	vcsClient         vcs.Client
	deleteLockCommand DeleteLockCommand
}

func (u *UnlockCommandRunner) Run(
	ctx *command.Context,
	cmd *command.Comment,
) {
	baseRepo := ctx.Pull.BaseRepo
	pullNum := ctx.Pull.Num

	vcsMessage := "All Atlantis locks for this PR have been unlocked and plans discarded"
	_, err := u.deleteLockCommand.DeleteLocksByPull(baseRepo.FullName, pullNum)
	if err != nil {
		vcsMessage = "Failed to delete PR locks"
		ctx.Log.Error(fmt.Sprintf("failed to delete locks by pull %s", err.Error()))
	}

	if commentErr := u.vcsClient.CreateComment(baseRepo, pullNum, vcsMessage, command.Unlock.String()); commentErr != nil {
		ctx.Log.Error(fmt.Sprintf("unable to comment: %s", commentErr))
	}
}
