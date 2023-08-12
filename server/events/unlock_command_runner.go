package events

import (
	"github.com/runatlantis/atlantis/server/events/command"
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
	ctx *command.Context,
	cmd *CommentCommand,
) {
	baseRepo := ctx.Pull.BaseRepo
	pullNum := ctx.Pull.Num

	vcsMessage := "All Atlantis locks for this PR have been unlocked and plans discarded"
	numLocks, dequeueStatus, err := u.deleteLockCommand.DeleteLocksByPull(baseRepo.FullName, pullNum)
	if err != nil {
		vcsMessage = "Failed to delete PR locks"
		ctx.Log.Err("failed to delete locks by pull %s", err.Error())
	}

	// if there are no locks to delete, no errors, and SilenceNoProjects is enabled, don't comment
	if err == nil && numLocks == 0 && u.SilenceNoProjects {
		return
	}

	if commentErr := u.vcsClient.CreateComment(baseRepo, pullNum, vcsMessage, command.Unlock.String()); commentErr != nil {
		ctx.Log.Err("unable to comment on PR %s: %s", pullNum, commentErr)
	}

	if dequeueStatus != nil {
		u.commentOnDequeuedPullRequests(ctx, *dequeueStatus)
	}
}

func (u *UnlockCommandRunner) commentOnDequeuedPullRequests(ctx *command.Context, dequeueStatus models.DequeueStatus) {
	locksByPullRequest := groupByPullRequests(dequeueStatus.ProjectLocks)
	for pullRequestNumber, projectLocks := range locksByPullRequest {
		planVcsMessage := models.BuildCommentOnDequeuedPullRequest(projectLocks)
		if commentErr := u.vcsClient.CreateComment(projectLocks[0].Pull.BaseRepo, pullRequestNumber, planVcsMessage, ""); commentErr != nil {
			ctx.Log.Err("unable to comment on PR %d: %s", pullRequestNumber, commentErr)
		}
	}
}

func groupByPullRequests(projectLocks []models.ProjectLock) map[int][]models.ProjectLock {
	result := make(map[int][]models.ProjectLock)
	for _, lock := range projectLocks {
		result[lock.Pull.Num] = append(result[lock.Pull.Num], lock)
	}
	return result
}
