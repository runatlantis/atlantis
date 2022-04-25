package events

import (
	"context"
	"fmt"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

func NewPolicyCheckCommandRunner(
	dbUpdater *DBUpdater,
	commitOutputUpdater CommitOutputUpdater,
	commitStatusUpdater CommitStatusUpdater,
	projectCommandRunner ProjectPolicyCheckCommandRunner,
	parallelPoolSize int,
) *PolicyCheckCommandRunner {
	return &PolicyCheckCommandRunner{
		dbUpdater:           dbUpdater,
		commitOutputUpdater: commitOutputUpdater,
		commitStatusUpdater: commitStatusUpdater,
		prjCmdRunner:        projectCommandRunner,
		parallelPoolSize:    parallelPoolSize,
	}
}

type PolicyCheckCommandRunner struct {
	dbUpdater           *DBUpdater
	commitOutputUpdater CommitOutputUpdater
	commitStatusUpdater CommitStatusUpdater
	prjCmdRunner        ProjectPolicyCheckCommandRunner
	parallelPoolSize    int
}

func (p *PolicyCheckCommandRunner) updateCombinedCommitStatusAndLogError(ctx *command.Context, status models.CommitStatus, cmdName fmt.Stringer) {
	baseRepo := ctx.Pull.BaseRepo
	pull := ctx.Pull
	statusID, err := p.commitStatusUpdater.UpdateCombined(context.TODO(), baseRepo, pull, status, cmdName, ctx.StatusID)
	if err != nil {
		ctx.Log.Warnf("unable to update commit status: %s", err)
	}

	// Assign the status ID to the commandContext for updates later in the workflow
	ctx.StatusID = statusID
}

func (p *PolicyCheckCommandRunner) updateCombinedCountCommitStatusAndLogError(ctx *command.Context, status models.CommitStatus, cmdName fmt.Stringer, numSuccess int, numTotal int) {
	baseRepo := ctx.Pull.BaseRepo
	pull := ctx.Pull
	statusID, err := p.commitStatusUpdater.UpdateCombinedCount(context.TODO(), baseRepo, pull, status, cmdName, numSuccess, numTotal, ctx.StatusID)
	if err != nil {
		ctx.Log.Warnf("unable to update commit status: %s", err)
	}

	// Assign the status ID to the commandContext for updates later in the workflow
	ctx.StatusID = statusID
}

func (p *PolicyCheckCommandRunner) Run(ctx *command.Context, cmds []command.ProjectContext) {
	if len(cmds) == 0 {
		ctx.Log.Infof("no projects to run policy_check in")
		// If there were no projects modified, we set successful commit statuses
		// with 0/0 projects policy_checked successfully because some users require
		// the Atlantis status to be passing for all pull requests.
		ctx.Log.Debugf("setting VCS status to success with no projects found")
		p.updateCombinedCountCommitStatusAndLogError(ctx, models.SuccessCommitStatus, command.PolicyCheck, 0, 0)
		return
	}

	// So set policy_check commit status to pending
	p.updateCombinedCommitStatusAndLogError(ctx, models.PendingCommitStatus, command.PolicyCheck)

	var result command.Result
	if p.isParallelEnabled(cmds) {
		ctx.Log.Infof("Running policy_checks in parallel")
		result = runProjectCmdsParallel(cmds, p.prjCmdRunner.PolicyCheck, p.parallelPoolSize)
	} else {
		result = runProjectCmds(cmds, p.prjCmdRunner.PolicyCheck)
	}

	p.commitOutputUpdater.Update(ctx, PolicyCheckCommand{}, result)

	pullStatus, err := p.dbUpdater.updateDB(ctx, ctx.Pull, result.ProjectResults)
	if err != nil {
		ctx.Log.Errorf("writing results: %s", err)
	}

	p.updateCommitStatus(ctx, pullStatus)
}

func (p *PolicyCheckCommandRunner) updateCommitStatus(ctx *command.Context, pullStatus models.PullStatus) {
	var numSuccess int
	var numErrored int
	status := models.SuccessCommitStatus

	numSuccess = pullStatus.StatusCount(models.PassedPolicyCheckStatus)
	numErrored = pullStatus.StatusCount(models.ErroredPolicyCheckStatus)

	if numErrored > 0 {
		status = models.FailedCommitStatus
	}

	p.updateCombinedCountCommitStatusAndLogError(ctx, status, command.PolicyCheck, numSuccess, len(pullStatus.Projects))
}

func (p *PolicyCheckCommandRunner) isParallelEnabled(cmds []command.ProjectContext) bool {
	return len(cmds) > 0 && cmds[0].ParallelPolicyCheckEnabled
}
