package events

import "github.com/runatlantis/atlantis/server/events/models"

func NewPolicyCheckCommandRunner(
	cmdRunner *DefaultCommandRunner,
	prjCmds []models.ProjectCommandContext,
) *PolicyCheckCommandRunner {
	return &PolicyCheckCommandRunner{
		cmdRunner:           cmdRunner,
		cmds:                prjCmds,
		commitStatusUpdater: cmdRunner.CommitStatusUpdater,
		prjCmdRunner:        cmdRunner.ProjectCommandRunner,
	}
}

type PolicyCheckCommandRunner struct {
	cmdRunner           *DefaultCommandRunner
	cmds                []models.ProjectCommandContext
	commitStatusUpdater CommitStatusUpdater
	prjCmdRunner        ProjectPolicyCheckCommandRunner
}

func (p *PolicyCheckCommandRunner) Run(ctx *CommandContext) {
	if len(p.cmds) == 0 {
		return
	}

	// So set policy_check commit status to pending
	if err := p.commitStatusUpdater.UpdateCombined(ctx.Pull.BaseRepo, ctx.Pull, models.PendingCommitStatus, models.PolicyCheckCommand); err != nil {
		ctx.Log.Warn("unable to update commit status: %s", err)
	}

	var result CommandResult
	if p.isParallelEnabled() {
		ctx.Log.Info("Running policy_checks in parallel")
		result = runProjectCmdsParallel(p.cmds, p.prjCmdRunner.PolicyCheck)
	} else {
		result = runProjectCmds(p.cmds, p.prjCmdRunner.PolicyCheck)
	}

	p.cmdRunner.updatePull(ctx, PolicyCheckCommand{}, result)

	pullStatus, err := p.cmdRunner.updateDB(ctx, ctx.Pull, result.ProjectResults)
	if err != nil {
		p.cmdRunner.Logger.Err("writing results: %s", err)
	}

	p.updateCommitStatus(ctx, pullStatus)
}

func (p *PolicyCheckCommandRunner) updateCommitStatus(ctx *CommandContext, pullStatus models.PullStatus) {
	var numSuccess int
	var numErrored int
	status := models.SuccessCommitStatus

	numSuccess = pullStatus.StatusCount(models.PassedPolicyCheckStatus)
	numErrored = pullStatus.StatusCount(models.ErroredPolicyCheckStatus)

	if numErrored > 0 {
		status = models.FailedCommitStatus
	}

	if err := p.commitStatusUpdater.UpdateCombinedCount(ctx.Pull.BaseRepo, ctx.Pull, status, models.PolicyCheckCommand, numSuccess, len(pullStatus.Projects)); err != nil {
		ctx.Log.Warn("unable to update commit status: %s", err)
	}
}

func (p *PolicyCheckCommandRunner) isParallelEnabled() bool {
	return len(p.cmds) > 0 && p.cmds[0].ParallelPolicyCheckEnabled
}
