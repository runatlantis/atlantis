package events

import (
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/vcs"
)

func NewImportCommandRunner(
	pullUpdater *PullUpdater,
	pullReqStatusFetcher vcs.PullReqStatusFetcher,
	prjCmdBuilder ProjectImportCommandBuilder,
	prjCmdRunner ProjectImportCommandRunner,
) *ImportCommandRunner {
	return &ImportCommandRunner{
		pullUpdater:          pullUpdater,
		pullReqStatusFetcher: pullReqStatusFetcher,
		prjCmdBuilder:        prjCmdBuilder,
		prjCmdRunner:         prjCmdRunner,
	}
}

type ImportCommandRunner struct {
	pullUpdater          *PullUpdater
	pullReqStatusFetcher vcs.PullReqStatusFetcher
	prjCmdBuilder        ProjectImportCommandBuilder
	prjCmdRunner         ProjectImportCommandRunner
}

func (v *ImportCommandRunner) Run(ctx *command.Context, cmd *CommentCommand) {
	var err error
	// Get the mergeable status before we set any build statuses of our own.
	// We do this here because when we set a "Pending" status, if users have
	// required the Atlantis status checks to pass, then we've now changed
	// the mergeability status of the pull request.
	// This sets the approved, mergeable, and sqlocked status in the context.
	ctx.PullRequestStatus, err = v.pullReqStatusFetcher.FetchPullStatus(ctx.Pull)
	if err != nil {
		// On error we continue the request with mergeable assumed false.
		// We want to continue because not all import will need this status,
		// only if they rely on the mergeability requirement.
		// All PullRequestStatus fields are set to false by default when error.
		ctx.Log.Warn("unable to get pull request status: %s. Continuing with mergeable and approved assumed false", err)
	}

	var projectCmds []command.ProjectContext
	projectCmds, err = v.prjCmdBuilder.BuildImportCommands(ctx, cmd)
	if err != nil {
		ctx.Log.Warn("Error %s", err)
	}

	var result command.Result
	if len(projectCmds) > 1 {
		// There is no usecase to kick terraform import into multiple projects.
		// To avoid incorrect import, suppress to execute terraform import in multiple projects.
		result = command.Result{
			Failure: "import cannot run on multiple projects. please specify one project.",
		}
	} else {
		result = runProjectCmds(projectCmds, v.prjCmdRunner.Import)
	}
	v.pullUpdater.updatePull(ctx, cmd, result)
}
