package events

import (
	"context"
	"fmt"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

func NewApprovePoliciesCommandRunner(
	commitStatusUpdater CommitStatusUpdater,
	prjCommandBuilder ProjectApprovePoliciesCommandBuilder,
	prjCommandRunner ProjectApprovePoliciesCommandRunner,
	commitOutputUpdater CommitOutputUpdater,
	dbUpdater *DBUpdater,
) *ApprovePoliciesCommandRunner {
	return &ApprovePoliciesCommandRunner{
		commitStatusUpdater: commitStatusUpdater,
		prjCmdBuilder:       prjCommandBuilder,
		prjCmdRunner:        prjCommandRunner,
		commitOutputUpdater: commitOutputUpdater,
		dbUpdater:           dbUpdater,
	}
}

type ApprovePoliciesCommandRunner struct {
	commitStatusUpdater CommitStatusUpdater
	commitOutputUpdater CommitOutputUpdater
	dbUpdater           *DBUpdater
	prjCmdBuilder       ProjectApprovePoliciesCommandBuilder
	prjCmdRunner        ProjectApprovePoliciesCommandRunner
}

func (p *ApprovePoliciesCommandRunner) updateCombinedCommitStatusAndLogError(ctx *command.Context, status models.CommitStatus, cmdName fmt.Stringer) {
	baseRepo := ctx.Pull.BaseRepo
	pull := ctx.Pull
	statusID, err := p.commitStatusUpdater.UpdateCombined(context.TODO(), baseRepo, pull, status, cmdName, ctx.StatusID)
	if err != nil {
		ctx.Log.Warnf("unable to update commit status: %s", err)
	}

	// Assign the status ID to the commandContext for updates later in the workflow
	ctx.StatusID = statusID
}

func (p *ApprovePoliciesCommandRunner) updateCombinedCountCommitStatusAndLogError(ctx *command.Context, status models.CommitStatus, cmdName fmt.Stringer, numSuccess int, numTotal int) {
	baseRepo := ctx.Pull.BaseRepo
	pull := ctx.Pull
	statusID, err := p.commitStatusUpdater.UpdateCombinedCount(context.TODO(), baseRepo, pull, status, cmdName, numSuccess, numTotal, ctx.StatusID)
	if err != nil {
		ctx.Log.Warnf("unable to update commit status: %s", err)
	}

	// Assign the status ID to the commandContext for updates later in the workflow
	ctx.StatusID = statusID
}

func (a *ApprovePoliciesCommandRunner) Run(ctx *command.Context, cmd *command.Comment) {
	pull := ctx.Pull
	a.updateCombinedCommitStatusAndLogError(ctx, models.PendingCommitStatus, command.PolicyCheck)

	projectCmds, err := a.prjCmdBuilder.BuildApprovePoliciesCommands(ctx, cmd)
	if err != nil {
		a.updateCombinedCommitStatusAndLogError(ctx, models.FailedCommitStatus, command.PolicyCheck)
		a.commitOutputUpdater.Update(ctx, cmd, command.Result{Error: err})
		return
	}

	if len(projectCmds) == 0 {
		ctx.Log.Infof("determined there was no project to run approve_policies in")
		// If there were no projects modified, we set successful commit statuses
		// with 0/0 projects approve_policies successfully because some users require
		// the Atlantis status to be passing for all pull requests.
		ctx.Log.Debugf("setting VCS status to success with no projects found")
		a.updateCombinedCountCommitStatusAndLogError(ctx, models.SuccessCommitStatus, command.PolicyCheck, 0, 0)
		return
	}

	result := a.buildApprovePolicyCommandResults(ctx, projectCmds)

	a.commitOutputUpdater.Update(
		ctx,
		cmd,
		result,
	)

	pullStatus, err := a.dbUpdater.updateDB(ctx, pull, result.ProjectResults)
	if err != nil {
		ctx.Log.Errorf("writing results: %s", err)
		return
	}

	a.updateCommitStatus(ctx, pullStatus)
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

func (a *ApprovePoliciesCommandRunner) updateCommitStatus(ctx *command.Context, pullStatus models.PullStatus) {
	var numSuccess int
	var numErrored int
	status := models.SuccessCommitStatus

	numSuccess = pullStatus.StatusCount(models.PassedPolicyCheckStatus)
	numErrored = pullStatus.StatusCount(models.ErroredPolicyCheckStatus)

	if numErrored > 0 {
		status = models.FailedCommitStatus
	}
	a.updateCombinedCountCommitStatusAndLogError(ctx, status, command.PolicyCheck, numSuccess, len(pullStatus.Projects))
}
