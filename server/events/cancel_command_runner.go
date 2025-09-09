package events

import (
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/vcs"
)

const cancelComment = "Marked pull request as cancelled. Future plan/apply operations will be skipped until new commits or manual reset."

func NewCancelCommandRunner(
	vcsClient vcs.Client,
	projectCmdRunner ProjectCommandRunner,
	pullUpdater *PullUpdater,
	silenceNoProjects bool,
) *CancelCommandRunner {
	return &CancelCommandRunner{
		VCSClient:         vcsClient,
		ProjectCmdRunner:  projectCmdRunner,
		PullUpdater:       pullUpdater,
		SilenceNoProjects: silenceNoProjects,
	}
}

type CancelCommandRunner struct {
	VCSClient         vcs.Client
	ProjectCmdRunner  ProjectCommandRunner
	PullUpdater       *PullUpdater
	SilenceNoProjects bool
}

func (c *CancelCommandRunner) Run(ctx *command.Context, cmd *CommentCommand) {
	if c.ProjectCmdRunner == nil {
		ctx.Log.Err("ProjectCmdRunner is nil")
		return
	}

	// Get the DefaultProjectCommandRunner to access the process tracker
	defaultRunner, ok := c.ProjectCmdRunner.(*DefaultProjectCommandRunner)
	if !ok {
		ctx.Log.Err("ProjectCmdRunner is not a DefaultProjectCommandRunner")
		return
	}

	if defaultRunner.ProcessTracker == nil {
		ctx.Log.Err("ProcessTracker is nil")
		return
	}

	// Mark the pull as cancelled
	defaultRunner.ProcessTracker.CancelPull(ctx.Pull)
	ctx.Log.Info("Pull request marked as cancelled; future operations will be skipped")
	if err := c.VCSClient.CreateComment(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull.Num, cancelComment, ""); err != nil {
		ctx.Log.Err("unable to comment: %s", err)
	}
}
