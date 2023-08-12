package events

import (
	"fmt"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"strings"
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
		planVcsMessage := buildCommentOnDequeuedPullRequest(projectLocks)
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

func buildCommentOnDequeuedPullRequest(projectLocks []models.ProjectLock) string {
	var releasedLocksMessages []string
	for _, lock := range projectLocks {
		releasedLocksMessages = append(releasedLocksMessages, fmt.Sprintf("* dir: `%s` workspace: `%s`", lock.Project.Path, lock.Workspace))
	}

	// stick to the first User for now, if needed, create a list of unique users and mention them all
	lockCreatorMention := "@" + projectLocks[0].User.Username
	releasedLocksMessage := strings.Join(releasedLocksMessages, "\n")

	return fmt.Sprintf("%s\nThe following locks have been aquired by this PR and can now be planned:\n%s",
		lockCreatorMention, releasedLocksMessage)
}
