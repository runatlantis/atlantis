package events

import (
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
)

func NewApprovePoliciesCommandRunner(
	commitStatusUpdater CommitStatusUpdater,
	prjCommandBuilder ProjectApprovePoliciesCommandBuilder,
	prjCommandRunner ProjectApprovePoliciesCommandRunner,
	pullUpdater *PullUpdater,
	dbUpdater *DBUpdater,
	SilenceNoProjects bool,
	silenceVCSStatusNoProjects bool,
	vcsClient vcs.Client,
) *ApprovePoliciesCommandRunner {
	return &ApprovePoliciesCommandRunner{
		commitStatusUpdater:        commitStatusUpdater,
		prjCmdBuilder:              prjCommandBuilder,
		prjCmdRunner:               prjCommandRunner,
		pullUpdater:                pullUpdater,
		dbUpdater:                  dbUpdater,
		SilenceNoProjects:          SilenceNoProjects,
		silenceVCSStatusNoProjects: silenceVCSStatusNoProjects,
		vcsClient:                  vcsClient,
	}
}

type ApprovePoliciesCommandRunner struct {
	commitStatusUpdater CommitStatusUpdater
	pullUpdater         *PullUpdater
	dbUpdater           *DBUpdater
	prjCmdBuilder       ProjectApprovePoliciesCommandBuilder
	prjCmdRunner        ProjectApprovePoliciesCommandRunner
	// SilenceNoProjects is whether Atlantis should respond to PRs if no projects
	// are found
	SilenceNoProjects          bool
	silenceVCSStatusNoProjects bool
	vcsClient                  vcs.Client
}

func (a *ApprovePoliciesCommandRunner) Run(ctx *command.Context, cmd *CommentCommand) {
	baseRepo := ctx.Pull.BaseRepo
	pull := ctx.Pull

	if err := a.commitStatusUpdater.UpdateCombined(ctx.Log, baseRepo, pull, models.PendingCommitStatus, command.PolicyCheck); err != nil {
		ctx.Log.Warn("unable to update commit status: %s", err)
	}

	projectCmds, err := a.prjCmdBuilder.BuildApprovePoliciesCommands(ctx, cmd)
	if err != nil {
		if statusErr := a.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, command.PolicyCheck); statusErr != nil {
			ctx.Log.Warn("unable to update commit status: %s", statusErr)
		}
		a.pullUpdater.updatePull(ctx, cmd, command.Result{Error: err})
		return
	}

	if len(projectCmds) == 0 && a.SilenceNoProjects {
		ctx.Log.Info("determined there was no project to run approve_policies in")
		if !a.silenceVCSStatusNoProjects {
			// If there were no projects modified, we set successful commit statuses
			// with 0/0 projects approve_policies successfully because some users require
			// the Atlantis status to be passing for all pull requests.
			ctx.Log.Debug("setting VCS status to success with no projects found")
			if err := a.commitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.SuccessCommitStatus, command.PolicyCheck, 0, 0); err != nil {
				ctx.Log.Warn("unable to update commit status: %s", err)
			}
		}
		return
	}

	result := runProjectCmds(projectCmds, a.prjCmdRunner.ApprovePolicies)

	a.pullUpdater.updatePull(
		ctx,
		cmd,
		result,
	)

	pullStatus, err := a.dbUpdater.updateDB(ctx, pull, result.ProjectResults)
	if err != nil {
		ctx.Log.Err("writing results: %s", err)
		return
	}

	a.updateCommitStatus(ctx, pullStatus)
}

func (a *ApprovePoliciesCommandRunner) updateCommitStatus(ctx *command.Context, pullStatus models.PullStatus) {
	var numSuccess int
	var numErrored int
	status := models.SuccessCommitStatus

	numSuccess = pullStatus.StatusCount(models.PassedPolicyCheckStatus)
	numErrored = pullStatus.StatusCount(models.ErroredPolicyCheckStatus)

	if numErrored > 0 {
		status = models.FailedCommitStatus
	}

	if err := a.commitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, status, command.PolicyCheck, numSuccess, len(pullStatus.Projects)); err != nil {
		ctx.Log.Warn("unable to update commit status: %s", err)
	}
}
