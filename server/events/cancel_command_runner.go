package events

import (
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/vcs"
)

const cancelComment = "Cancelled all queued operations and future execution groups for this pull request. Currently running operations will continue to completion."

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

	if defaultRunner.CancellationTracker == nil {
		ctx.Log.Err("CancellationTracker is nil")
		return
	}

	// Cancel the entire pull request to prevent future execution order groups from running
	defaultRunner.CancellationTracker.Cancel(ctx.Pull)
	ctx.Log.Info("Cancelled all queued operations and future execution groups for pull request; currently running operations will continue to completion")
	if err := c.VCSClient.CreateComment(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull.Num, cancelComment, ""); err != nil {
		ctx.Log.Err("unable to comment: %s", err)
	}
}
