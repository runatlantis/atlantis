package events

import (
	"fmt"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

type commandOutputGenerator interface {
	GeneratePolicyCheckOutputStore(ctx *command.Context, cmd *command.Comment) (command.PolicyCheckOutputStore, error)
}

func NewApprovePoliciesCommandRunner(
	commitStatusUpdater CommitStatusUpdater,
	prjCommandBuilder ProjectApprovePoliciesCommandBuilder,
	prjCommandRunner ProjectApprovePoliciesCommandRunner,
	outputUpdater OutputUpdater,
	dbUpdater *DBUpdater,
	policyCheckOutputGenerator commandOutputGenerator,
) *ApprovePoliciesCommandRunner {
	return &ApprovePoliciesCommandRunner{
		commitStatusUpdater:        commitStatusUpdater,
		prjCmdBuilder:              prjCommandBuilder,
		prjCmdRunner:               prjCommandRunner,
		outputUpdater:              outputUpdater,
		dbUpdater:                  dbUpdater,
		policyCheckOutputGenerator: policyCheckOutputGenerator,
	}
}

type ApprovePoliciesCommandRunner struct {
	commitStatusUpdater        CommitStatusUpdater
	outputUpdater              OutputUpdater
	dbUpdater                  *DBUpdater
	prjCmdBuilder              ProjectApprovePoliciesCommandBuilder
	prjCmdRunner               ProjectApprovePoliciesCommandRunner
	policyCheckOutputGenerator commandOutputGenerator
}

func (a *ApprovePoliciesCommandRunner) Run(ctx *command.Context, cmd *command.Comment) {
	baseRepo := ctx.Pull.BaseRepo
	pull := ctx.Pull

	statusID, err := a.commitStatusUpdater.UpdateCombined(ctx.RequestCtx, baseRepo, pull, models.PendingCommitStatus, command.PolicyCheck, "", "")
	if err != nil {
		ctx.Log.WarnContext(ctx.RequestCtx, fmt.Sprintf("unable to update commit status: %s", err))
	}

	projectCmds, err := a.prjCmdBuilder.BuildApprovePoliciesCommands(ctx, cmd)
	if err != nil {
		if _, statusErr := a.commitStatusUpdater.UpdateCombined(ctx.RequestCtx, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, command.PolicyCheck, statusID, ""); statusErr != nil {
			ctx.Log.WarnContext(ctx.RequestCtx, fmt.Sprintf("unable to update commit status: %s", statusErr))
		}
		a.outputUpdater.UpdateOutput(ctx, cmd, command.Result{Error: err})
		return
	}

	if len(projectCmds) == 0 {
		ctx.Log.InfoContext(ctx.RequestCtx, "determined there was no project to run approve_policies in")
		// If there were no projects modified, we set successful commit statuses
		// with 0/0 projects approve_policies successfully because some users require
		// the Atlantis status to be passing for all pull requests.
		if _, err := a.commitStatusUpdater.UpdateCombinedCount(ctx.RequestCtx, ctx.Pull.BaseRepo, ctx.Pull, models.SuccessCommitStatus, command.PolicyCheck, 0, 0, statusID); err != nil {
			ctx.Log.WarnContext(ctx.RequestCtx, fmt.Sprintf("unable to update commit status: %s", err))
		}
		return
	}

	result := a.buildApprovePolicyCommandResults(ctx, projectCmds)

	// Adds the policy check output for failing policies which needs to be populated when using github checks
	// Noop when github checks is not enabled.
	policyCheckOutputStore, err := a.policyCheckOutputGenerator.GeneratePolicyCheckOutputStore(ctx, cmd)
	if err != nil {
		if _, statusErr := a.commitStatusUpdater.UpdateCombined(ctx.RequestCtx, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, command.PolicyCheck, statusID, ""); statusErr != nil {
			ctx.Log.WarnContext(ctx.RequestCtx, fmt.Sprintf("unable to update commit status: %s", statusErr))
		}
		a.outputUpdater.UpdateOutput(ctx, cmd, command.Result{Error: err})
		return
	}

	for i, prjResult := range result.ProjectResults {
		policyCheckOutput := policyCheckOutputStore.Get(prjResult.ProjectName, prjResult.Workspace)
		if policyCheckOutput != nil {
			result.ProjectResults[i].PolicyCheckSuccess = policyCheckOutput
		}
	}

	a.outputUpdater.UpdateOutput(
		ctx,
		cmd,
		result,
	)

	pullStatus, err := a.dbUpdater.updateDB(ctx, pull, result.ProjectResults)
	if err != nil {
		ctx.Log.ErrorContext(ctx.RequestCtx, fmt.Sprintf("writing results: %s", err))
		return
	}

	a.updateCommitStatus(ctx, pullStatus, statusID)
}

func (a *ApprovePoliciesCommandRunner) buildApprovePolicyCommandResults(ctx *command.Context, prjCmds []command.ProjectContext) (result command.Result) {
	// Check if vcs user is in the owner list of the PolicySets. All projects
	// share the same Owners list at this time so no reason to iterate over each
	// project.
	if len(prjCmds) > 0 && !prjCmds[0].PolicySets.IsOwner(ctx.User.Username) {
		result.Error = fmt.Errorf("contact policy owners to approve failing policies")
		return
	}

	var prjResults []command.ProjectResult

	for _, prjCmd := range prjCmds {
		prjResult := a.prjCmdRunner.ApprovePolicies(prjCmd)
		prjResults = append(prjResults, prjResult)
	}
	result.ProjectResults = prjResults
	return
}

func (a *ApprovePoliciesCommandRunner) updateCommitStatus(ctx *command.Context, pullStatus models.PullStatus, statusID string) {
	var numSuccess int
	var numErrored int
	status := models.SuccessCommitStatus

	numSuccess = pullStatus.StatusCount(models.PassedPolicyCheckStatus)
	numErrored = pullStatus.StatusCount(models.ErroredPolicyCheckStatus)

	if numErrored > 0 {
		status = models.FailedCommitStatus
	}

	if _, err := a.commitStatusUpdater.UpdateCombinedCount(ctx.RequestCtx, ctx.Pull.BaseRepo, ctx.Pull, status, command.PolicyCheck, numSuccess, len(pullStatus.Projects), statusID); err != nil {
		ctx.Log.WarnContext(ctx.RequestCtx, fmt.Sprintf("unable to update commit status: %s", err))
	}
}
