package events

import (
	"fmt"
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

	u.commentOnDequeuedPullRequests(ctx, dequeueStatus, baseRepo)
}

// TODO(Ghais) extract to a common interface and use when:
//  * Pull request is closed: server/events/pull_closed_executor.go:106
//  * Unlocked from the UI
func (u *UnlockCommandRunner) commentOnDequeuedPullRequests(ctx *CommandContext, dequeueStatus models.DequeueStatus, baseRepo models.Repo) {
	locksByPullRequest := groupByPullRequests(dequeueStatus.ProjectLocks)
	for pullRequestNumber, projectLocks := range locksByPullRequest {
		planVcsMessage := buildCommentOnDequeuedPullRequest(projectLocks)
		if commentErr := u.vcsClient.CreateComment(baseRepo, pullRequestNumber, planVcsMessage, ""); commentErr != nil {
			ctx.Log.Err("unable to comment on PR %s: %s", pullRequestNumber, commentErr)
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
