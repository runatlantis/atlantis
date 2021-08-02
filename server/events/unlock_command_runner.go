package events

import (
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
)

func NewUnlockCommandRunner(
	deleteLockCommand DeleteLockCommand,
	vcsClient vcs.Client,
	SilenceNoProjects bool,
) *UnlockCommandRunner {
	return &UnlockCommandRunner{
		deleteLockCommand: deleteLockCommand,
		vcsClient:         vcsClient,
		SilenceNoProjects: SilenceNoProjects,
	}
}

type UnlockCommandRunner struct {
	vcsClient         vcs.Client
	deleteLockCommand DeleteLockCommand
	// SilenceNoProjects is whether Atlantis should respond to PRs if no projects
	// are found
	SilenceNoProjects bool
}

func (u *UnlockCommandRunner) Run(
	ctx *CommandContext,
	cmd *CommentCommand,
) {
	baseRepo := ctx.Pull.BaseRepo
	pullNum := ctx.Pull.Num

	numLocks, dequeueStatus, err := u.deleteLockCommand.DeleteLocksByPull(baseRepo.FullName, pullNum)
	vcsMessage := prepareUnlockedVcsMessage(dequeueStatus, err, ctx)

	// if there are no locks to delete, no errors, and SilenceNoProjects is enabled, don't comment
	if err == nil && numLocks == 0 && u.SilenceNoProjects {
		return
	}

	if commentErr := u.vcsClient.CreateComment(baseRepo, pullNum, vcsMessage, models.UnlockCommand.String()); commentErr != nil {
		ctx.Log.Err("unable to comment on PR %s: %s", pullNum, commentErr)
	}

	u.triggerPlansForDequeuedPRs(ctx, dequeueStatus, baseRepo)
}

func (u *UnlockCommandRunner) triggerPlansForDequeuedPRs(ctx *CommandContext, dequeueStatus models.DequeueStatus, baseRepo models.Repo) {
	for _, lock := range dequeueStatus.ProjectLocks {
		// TODO monikma #4 use exact dequeued comment instead of hardcoding it
		planVcsMessage := "atlantis plan -d " + lock.Project.Path
		if commentErr := u.vcsClient.CreateComment(baseRepo, lock.Pull.Num, planVcsMessage, ""); commentErr != nil {
			// TODO monikma at this point planning queue will be interrupted, how to resolve from this?
			ctx.Log.Err("unable to comment on PR %s: %s", lock.Pull.Num, commentErr)
		}
	}
}

func prepareUnlockedVcsMessage(dequeueStatus models.DequeueStatus, err error, ctx *CommandContext) string {
	vcsMessage := "All Atlantis locks for this PR have been unlocked and plans discarded"
	if len(dequeueStatus.ProjectLocks) > 0 {
		vcsMessage = vcsMessage + dequeueStatus.String()
	}
	if err != nil {
		vcsMessage = "Failed to delete PR locks"
		ctx.Log.Err("failed to delete locks by pull %s", err.Error())
	}
	return vcsMessage
}
