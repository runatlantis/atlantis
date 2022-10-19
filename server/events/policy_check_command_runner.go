package events

import (
	"fmt"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

func NewPolicyCheckCommandRunner(
	dbUpdater *DBUpdater,
	outputUpdater OutputUpdater,
	vcsStatusUpdater VCSStatusUpdater,
	projectCommandRunner ProjectPolicyCheckCommandRunner,
	parallelPoolSize int,
) *PolicyCheckCommandRunner {
	return &PolicyCheckCommandRunner{
		dbUpdater:        dbUpdater,
		outputUpdater:    outputUpdater,
		vcsStatusUpdater: vcsStatusUpdater,
		prjCmdRunner:     projectCommandRunner,
		parallelPoolSize: parallelPoolSize,
	}
}

type PolicyCheckCommandRunner struct {
	dbUpdater        *DBUpdater
	outputUpdater    OutputUpdater
	vcsStatusUpdater VCSStatusUpdater
	prjCmdRunner     ProjectPolicyCheckCommandRunner
	parallelPoolSize int
}

func (p *PolicyCheckCommandRunner) Run(ctx *command.Context, cmds []command.ProjectContext) {
	if len(cmds) == 0 {
		ctx.Log.InfoContext(ctx.RequestCtx, "no projects to run policy_check in")
		// If there were no projects modified, we set successful commit statuses
		// with 0/0 projects policy_checked successfully because some users require
		// the Atlantis status to be passing for all pull requests.
		if _, err := p.vcsStatusUpdater.UpdateCombinedCount(ctx.RequestCtx, ctx.Pull.BaseRepo, ctx.Pull, models.SuccessVCSStatus, command.PolicyCheck, 0, 0, ""); err != nil {
			ctx.Log.WarnContext(ctx.RequestCtx, fmt.Sprintf("unable to update commit status: %s", err))
		}
		return
	}

	// So set policy_check commit status to pending
	statusID, err := p.vcsStatusUpdater.UpdateCombined(ctx.RequestCtx, ctx.Pull.BaseRepo, ctx.Pull, models.PendingVCSStatus, command.PolicyCheck, "", "")
	if err != nil {
		ctx.Log.WarnContext(ctx.RequestCtx, fmt.Sprintf("unable to update commit status: %s", err))
	}

	var result command.Result
	if p.isParallelEnabled(cmds) {
		ctx.Log.InfoContext(ctx.RequestCtx, "Running policy_checks in parallel")
		result = runProjectCmdsParallel(cmds, p.prjCmdRunner.PolicyCheck, p.parallelPoolSize)
	} else {
		result = runProjectCmds(cmds, p.prjCmdRunner.PolicyCheck)
	}

	p.outputUpdater.UpdateOutput(ctx, PolicyCheckCommand{}, result)

	pullStatus, err := p.dbUpdater.updateDB(ctx, ctx.Pull, result.ProjectResults)
	if err != nil {
		ctx.Log.ErrorContext(ctx.RequestCtx, fmt.Sprintf("writing results: %s", err))
	}

	p.updateVcsStatus(ctx, pullStatus, statusID)
}

func (p *PolicyCheckCommandRunner) updateVcsStatus(ctx *command.Context, pullStatus models.PullStatus, statusID string) {
	var numSuccess int
	var numErrored int
	status := models.SuccessVCSStatus

	numSuccess = pullStatus.StatusCount(models.PassedPolicyCheckStatus)
	numErrored = pullStatus.StatusCount(models.ErroredPolicyCheckStatus)

	if numErrored > 0 {
		status = models.FailedVCSStatus
	}

	if _, err := p.vcsStatusUpdater.UpdateCombinedCount(ctx.RequestCtx, ctx.Pull.BaseRepo, ctx.Pull, status, command.PolicyCheck, numSuccess, len(pullStatus.Projects), statusID); err != nil {
		ctx.Log.WarnContext(ctx.RequestCtx, fmt.Sprintf("unable to update commit status: %s", err))
	}
}

func (p *PolicyCheckCommandRunner) isParallelEnabled(cmds []command.ProjectContext) bool {
	return len(cmds) > 0 && cmds[0].ParallelPolicyCheckEnabled
}
