package events

import (
	"fmt"

	"github.com/runatlantis/atlantis/server/events/models"
)

func NewApprovePoliciesCommandRunner(
	commitStatusUpdater CommitStatusUpdater,
	prjCommandBuilder ProjectApprovePoliciesCommandBuilder,
	prjCommandRunner ProjectApprovePoliciesCommandRunner,
	pullUpdater *PullUpdater,
	dbUpdater *DBUpdater,
	silenceNoProjects bool,
) *ApprovePoliciesCommandRunner {
	return &ApprovePoliciesCommandRunner{
		commitStatusUpdater: commitStatusUpdater,
		prjCmdBuilder:       prjCommandBuilder,
		prjCmdRunner:        prjCommandRunner,
		pullUpdater:         pullUpdater,
		dbUpdater:           dbUpdater,
		silenceNoProjects:   silenceNoProjects,
	}
}

type ApprovePoliciesCommandRunner struct {
	commitStatusUpdater CommitStatusUpdater
	pullUpdater         *PullUpdater
	dbUpdater           *DBUpdater
	prjCmdBuilder       ProjectApprovePoliciesCommandBuilder
	prjCmdRunner        ProjectApprovePoliciesCommandRunner
	silenceNoProjects   bool
}

func (a *ApprovePoliciesCommandRunner) Run(ctx *CommandContext, cmd *CommentCommand) {
	baseRepo := ctx.Pull.BaseRepo
	pull := ctx.Pull

	if err := a.commitStatusUpdater.UpdateCombined(baseRepo, pull, models.PendingCommitStatus, models.PolicyCheckCommand); err != nil {
		ctx.Log.Warn("unable to update commit status: %s", err)
	}

	projectCmds, err := a.prjCmdBuilder.BuildApprovePoliciesCommands(ctx, cmd)
	if err != nil {
		if statusErr := a.commitStatusUpdater.UpdateCombined(ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, models.PolicyCheckCommand); statusErr != nil {
			ctx.Log.Warn("unable to update commit status: %s", statusErr)
		}
		a.pullUpdater.updatePull(ctx, cmd, CommandResult{Error: err})
		return
	}

	if len(projectCmds) == 0 && a.silenceNoProjects {
		ctx.Log.Info("determined there was no project to run approve_policies in")
		return
	}

	result := a.buildApprovePolicyCommandResults(ctx, projectCmds)

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

func (a *ApprovePoliciesCommandRunner) buildApprovePolicyCommandResults(ctx *CommandContext, prjCmds []models.ProjectCommandContext) (result CommandResult) {
	// Check if vcs user is in the owner list of the PolicySets. All projects
	// share the same Owners list at this time so no reason to iterate over each
	// project.
	if len(prjCmds) > 0 && !prjCmds[0].PolicySets.IsOwner(ctx.User.Username) {
		result.Error = fmt.Errorf("contact policy owners to approve failing policies")
		return
	}

	var prjResults []models.ProjectResult

	for _, prjCmd := range prjCmds {
		prjResult := a.prjCmdRunner.ApprovePolicies(prjCmd)
		prjResults = append(prjResults, prjResult)
	}
	result.ProjectResults = prjResults
	return
}

func (a *ApprovePoliciesCommandRunner) updateCommitStatus(ctx *CommandContext, pullStatus models.PullStatus) {
	// Don't updateCommitStatus either!
	if a.silenceNoProjects {
		return
	}

	var numSuccess int
	var numErrored int
	status := models.SuccessCommitStatus

	numSuccess = pullStatus.StatusCount(models.PassedPolicyCheckStatus)
	numErrored = pullStatus.StatusCount(models.ErroredPolicyCheckStatus)

	if numErrored > 0 {
		status = models.FailedCommitStatus
	}

	if err := a.commitStatusUpdater.UpdateCombinedCount(ctx.Pull.BaseRepo, ctx.Pull, status, models.PolicyCheckCommand, numSuccess, len(pullStatus.Projects)); err != nil {
		ctx.Log.Warn("unable to update commit status: %s", err)
	}
}
