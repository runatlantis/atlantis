package events

import (
	"fmt"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/vcs"
)

const (
	cancelNoOperationsComment = "No running operations found for this pull request."
	cancelFailedComment       = "Failed to cancel any operations for this pull request."
	cancelSuccessComment      = "Cancelled %d running operation(s) for this pull request."
)

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

	// Get running processes for this pull request
	runningProcesses := defaultRunner.ProcessTracker.GetRunningProcesses(ctx.Pull)

	if len(runningProcesses) == 0 {
		// Even if there are no current processes, mark the pull as cancelled so
		// any future operations will immediately observe cancellation.
		defaultRunner.ProcessTracker.CancelPull(ctx.Pull)
		if err := c.VCSClient.CreateComment(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull.Num, cancelNoOperationsComment, ""); err != nil {
			ctx.Log.Err("unable to comment: %s", err)
		}
		return
	}

	// Cancel all running operations and mark the pull as cancelled for any
	// subsequent operations.
	cancelledCount := 0
	for _, process := range runningProcesses {
		if err := defaultRunner.ProcessTracker.CancelOperation(process.PID); err != nil {
			ctx.Log.Warn("Failed to cancel operation (%s): %s", process.Command, err)
		} else {
			ctx.Log.Info("Cancelled operation (%s) for project %s", process.Command, process.Project)
			cancelledCount++
		}
	}
	defaultRunner.ProcessTracker.CancelPull(ctx.Pull)

	// Create a comment with the results
	var comment string
	if cancelledCount > 0 {
		comment = fmt.Sprintf(cancelSuccessComment, cancelledCount)
	} else {
		comment = cancelFailedComment
	}

	if err := c.VCSClient.CreateComment(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull.Num, comment, ""); err != nil {
		ctx.Log.Err("unable to comment: %s", err)
	}
}
