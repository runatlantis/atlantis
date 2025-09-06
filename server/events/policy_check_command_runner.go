package events

import (
	"errors"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/status"
)

func NewPolicyCheckCommandRunner(
	dbUpdater *DBUpdater,
	pullUpdater *PullUpdater,
	commitStatusUpdater CommitStatusUpdater,
	projectCommandRunner ProjectPolicyCheckCommandRunner,
	parallelPoolSize int,
	silenceVCSStatusNoProjects bool,
	quietPolicyChecks bool,
	statusManager status.StatusManager,
) *PolicyCheckCommandRunner {
	return &PolicyCheckCommandRunner{
		dbUpdater:                  dbUpdater,
		pullUpdater:                pullUpdater,
		commitStatusUpdater:        commitStatusUpdater,
		prjCmdRunner:               projectCommandRunner,
		parallelPoolSize:           parallelPoolSize,
		silenceVCSStatusNoProjects: silenceVCSStatusNoProjects,
		quietPolicyChecks:          quietPolicyChecks,
		StatusManager:              statusManager,
	}
}

type PolicyCheckCommandRunner struct {
	dbUpdater           *DBUpdater
	pullUpdater         *PullUpdater
	commitStatusUpdater CommitStatusUpdater
	prjCmdRunner        ProjectPolicyCheckCommandRunner
	parallelPoolSize    int
	StatusManager       status.StatusManager
	// SilenceVCSStatusNoProjects is whether any plan should set commit status if no projects
	// are found
	silenceVCSStatusNoProjects bool
	quietPolicyChecks          bool
}

func (p *PolicyCheckCommandRunner) Run(ctx *command.Context, cmds []command.ProjectContext) {
	if len(cmds) == 0 {
		ctx.Log.Info("no projects to run policy_check in")
		// Use StatusManager to handle no projects found with policy-aware decisions
		if err := p.StatusManager.HandleNoProjectsFound(ctx, command.PolicyCheck); err != nil {
			ctx.Log.Warn("unable to handle no projects status: %s", err)
		}
		return
	}

	// Set policy_check commit status to pending
	if err := p.StatusManager.SetPending(ctx, command.PolicyCheck); err != nil {
		ctx.Log.Warn("unable to update commit status: %s", err)
	}

	var result command.Result
	if p.isParallelEnabled(cmds) {
		ctx.Log.Info("Running policy_checks in parallel")
		result = runProjectCmdsParallel(cmds, p.prjCmdRunner.PolicyCheck, p.parallelPoolSize)
	} else {
		result = runProjectCmds(cmds, p.prjCmdRunner.PolicyCheck)
	}

	// Quiet policy checks unless there's an error
	if result.HasErrors() || !p.quietPolicyChecks {
		p.pullUpdater.updatePull(ctx, PolicyCheckCommand{}, result)
	}

	pullStatus, err := p.dbUpdater.updateDB(ctx, ctx.Pull, result.ProjectResults)
	if err != nil {
		ctx.Log.Err("writing results: %s", err)
	}

	p.updateCommitStatus(ctx, pullStatus)
}

func (p *PolicyCheckCommandRunner) updateCommitStatus(ctx *command.Context, pullStatus models.PullStatus) {
	var numSuccess int
	var numErrored int

	numSuccess = pullStatus.StatusCount(models.PassedPolicyCheckStatus)
	numErrored = pullStatus.StatusCount(models.ErroredPolicyCheckStatus)

	if numErrored > 0 {
		// Use a fake error for the failure - StatusManager will handle the status setting
		err := errors.New("policy check failed for one or more projects")
		if statusErr := p.StatusManager.SetFailure(ctx, command.PolicyCheck, err); statusErr != nil {
			ctx.Log.Warn("unable to update commit status: %s", statusErr)
		}
	} else {
		// All successful
		if statusErr := p.StatusManager.SetSuccess(ctx, command.PolicyCheck, numSuccess, len(pullStatus.Projects)); statusErr != nil {
			ctx.Log.Warn("unable to update commit status: %s", statusErr)
		}
	}
}

func (p *PolicyCheckCommandRunner) isParallelEnabled(cmds []command.ProjectContext) bool {
	return len(cmds) > 0 && cmds[0].ParallelPolicyCheckEnabled
}
