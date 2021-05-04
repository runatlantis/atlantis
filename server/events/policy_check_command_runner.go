package events

import "github.com/runatlantis/atlantis/server/events/models"

func NewPolicyCheckCommandRunner(
	dbUpdater *DBUpdater,
	pullUpdater *PullUpdater,
	commitStatusUpdater CommitStatusUpdater,
	projectCommandRunner ProjectPolicyCheckCommandRunner,
	parallelPoolSize int,
	silenceVCSStatusNoProjects bool,
) *PolicyCheckCommandRunner {
	return &PolicyCheckCommandRunner{
		dbUpdater:                  dbUpdater,
		pullUpdater:                pullUpdater,
		commitStatusUpdater:        commitStatusUpdater,
		prjCmdRunner:               projectCommandRunner,
		parallelPoolSize:           parallelPoolSize,
		silenceVCSStatusNoProjects: silenceVCSStatusNoProjects,
	}
}

type PolicyCheckCommandRunner struct {
	dbUpdater           *DBUpdater
	pullUpdater         *PullUpdater
	commitStatusUpdater CommitStatusUpdater
	prjCmdRunner        ProjectPolicyCheckCommandRunner
	parallelPoolSize    int
	// SilenceVCSStatusNoProjects is whether any plan should set commit status if no projects
	// are found
	silenceVCSStatusNoProjects bool
}

func (p *PolicyCheckCommandRunner) Run(ctx *CommandContext, cmds []models.ProjectCommandContext) {
	if len(cmds) == 0 {
		ctx.Log.Info("no projects to run policy_check in")
		if !p.silenceVCSStatusNoProjects {
			// If there were no projects modified, we set successful commit statuses
			// with 0/0 projects policy_checked successfully because some users require
			// the Atlantis status to be passing for all pull requests.
			ctx.Log.Debug("setting VCS status to success with no projects found")
			if err := p.commitStatusUpdater.UpdateCombinedCount(ctx.Pull.BaseRepo, ctx.Pull, models.SuccessCommitStatus, models.PolicyCheckCommand, 0, 0); err != nil {
				ctx.Log.Warn("unable to update commit status: %s", err)
			}
		}
		return
	}

	// So set policy_check commit status to pending
	if err := p.commitStatusUpdater.UpdateCombined(ctx.Pull.BaseRepo, ctx.Pull, models.PendingCommitStatus, models.PolicyCheckCommand); err != nil {
		ctx.Log.Warn("unable to update commit status: %s", err)
	}

	var result CommandResult
	if p.isParallelEnabled(cmds) {
		ctx.Log.Info("Running policy_checks in parallel")
		result = runProjectCmdsParallel(cmds, p.prjCmdRunner.PolicyCheck, p.parallelPoolSize)
	} else {
		result = runProjectCmds(cmds, p.prjCmdRunner.PolicyCheck)
	}

	p.pullUpdater.updatePull(ctx, PolicyCheckCommand{}, result)

	pullStatus, err := p.dbUpdater.updateDB(ctx, ctx.Pull, result.ProjectResults)
	if err != nil {
		ctx.Log.Err("writing results: %s", err)
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

func (p *PolicyCheckCommandRunner) isParallelEnabled(cmds []models.ProjectCommandContext) bool {
	return len(cmds) > 0 && cmds[0].ParallelPolicyCheckEnabled
}
