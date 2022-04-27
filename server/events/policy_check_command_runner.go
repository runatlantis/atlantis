package events

import (
	"context"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

func NewPolicyCheckCommandRunner(
	dbUpdater *DBUpdater,
	pullUpdater *PullUpdater,
	commitStatusUpdater CommitStatusUpdater,
	projectCommandRunner ProjectPolicyCheckCommandRunner,
	parallelPoolSize int,
) *PolicyCheckCommandRunner {
	return &PolicyCheckCommandRunner{
		dbUpdater:           dbUpdater,
		pullUpdater:         pullUpdater,
		commitStatusUpdater: commitStatusUpdater,
		prjCmdRunner:        projectCommandRunner,
		parallelPoolSize:    parallelPoolSize,
	}
}

type PolicyCheckCommandRunner struct {
	dbUpdater           *DBUpdater
	pullUpdater         *PullUpdater
	commitStatusUpdater CommitStatusUpdater
	prjCmdRunner        ProjectPolicyCheckCommandRunner
	parallelPoolSize    int
}

func (p *PolicyCheckCommandRunner) Run(ctx *command.Context, cmds []command.ProjectContext) {
	if len(cmds) == 0 {
		ctx.Log.Infof("no projects to run policy_check in")
		// If there were no projects modified, we set successful commit statuses
		// with 0/0 projects policy_checked successfully because some users require
		// the Atlantis status to be passing for all pull requests.
		if err := p.commitStatusUpdater.UpdateCombinedCount(context.TODO(), ctx.Pull.BaseRepo, ctx.Pull, models.SuccessCommitStatus, command.PolicyCheck, 0, 0); err != nil {
			ctx.Log.Warnf("unable to update commit status: %s", err)
		}
		return
	}

	// So set policy_check commit status to pending
	if err := p.commitStatusUpdater.UpdateCombined(context.TODO(), ctx.Pull.BaseRepo, ctx.Pull, models.PendingCommitStatus, command.PolicyCheck); err != nil {
		ctx.Log.Warnf("unable to update commit status: %s", err)
	}

	var result command.Result
	if p.isParallelEnabled(cmds) {
		ctx.Log.Infof("Running policy_checks in parallel")
		result = runProjectCmdsParallel(cmds, p.prjCmdRunner.PolicyCheck, p.parallelPoolSize)
	} else {
		result = runProjectCmds(cmds, p.prjCmdRunner.PolicyCheck)
	}

	p.pullUpdater.UpdatePull(ctx, PolicyCheckCommand{}, result)

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

	if err := p.commitStatusUpdater.UpdateCombinedCount(context.TODO(), ctx.Pull.BaseRepo, ctx.Pull, status, command.PolicyCheck, numSuccess, len(pullStatus.Projects)); err != nil {
		ctx.Log.Warnf("unable to update commit status: %s", err)
	}
}

func (p *PolicyCheckCommandRunner) isParallelEnabled(cmds []command.ProjectContext) bool {
	return len(cmds) > 0 && cmds[0].ParallelPolicyCheckEnabled
}
