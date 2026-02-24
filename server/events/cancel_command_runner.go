package events

import (
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/workingdir"
)

const cancelComment = "Cancelled all queued operations and released working directory locks for this pull request.\n" +
	"New operations can now be started. Currently running operations will continue to completion."

func NewCancelCommandRunner(
	vcsClient vcs.Client,
	projectCmdRunner ProjectCommandRunner,
	pullUpdater *PullUpdater,
	workingDirLocker workingdir.Locker,
	silenceNoProjects bool,
) *CancelCommandRunner {
	return &CancelCommandRunner{
		VCSClient:         vcsClient,
		ProjectCmdRunner:  projectCmdRunner,
		PullUpdater:       pullUpdater,
		WorkingDirLocker:  workingDirLocker,
		SilenceNoProjects: silenceNoProjects,
	}
}

type CancelCommandRunner struct {
	VCSClient         vcs.Client
	ProjectCmdRunner  ProjectCommandRunner
	PullUpdater       *PullUpdater
	WorkingDirLocker  workingdir.Locker
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

	// Clean up working directory locks for this pull request
	if defaultRunner.WorkingDirLocker != nil {
		defaultRunner.WorkingDirLocker.UnlockByPull(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num)
		ctx.Log.Debug("Released working directory locks for pull request")
	}

	ctx.Log.Info("Cancelled all queued operations and future execution groups for pull request; currently running operations will continue to completion")
	if err := c.VCSClient.CreateComment(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull.Num, cancelComment, ""); err != nil {
		ctx.Log.Err("unable to comment: %s", err)
	}
}
