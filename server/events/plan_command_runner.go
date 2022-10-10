package events

import (
	"fmt"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
)

func NewPlanCommandRunner(
	vcsClient vcs.Client,
	pendingPlanFinder PendingPlanFinder,
	workingDir WorkingDir,
	commitStatusUpdater CommitStatusUpdater,
	projectCommandBuilder ProjectPlanCommandBuilder,
	projectCommandRunner ProjectPlanCommandRunner,
	dbUpdater *DBUpdater,
	outputUpdater OutputUpdater,
	policyCheckCommandRunner *PolicyCheckCommandRunner,
	parallelPoolSize int,
) *PlanCommandRunner {
	return &PlanCommandRunner{
		vcsClient:                vcsClient,
		pendingPlanFinder:        pendingPlanFinder,
		workingDir:               workingDir,
		commitStatusUpdater:      commitStatusUpdater,
		prjCmdBuilder:            projectCommandBuilder,
		prjCmdRunner:             projectCommandRunner,
		dbUpdater:                dbUpdater,
		outputUpdater:            outputUpdater,
		policyCheckCommandRunner: policyCheckCommandRunner,
		parallelPoolSize:         parallelPoolSize,
	}
}

type PlanCommandRunner struct {
	vcsClient                vcs.Client
	commitStatusUpdater      CommitStatusUpdater
	pendingPlanFinder        PendingPlanFinder
	workingDir               WorkingDir
	prjCmdBuilder            ProjectPlanCommandBuilder
	prjCmdRunner             ProjectPlanCommandRunner
	dbUpdater                *DBUpdater
	outputUpdater            OutputUpdater
	policyCheckCommandRunner *PolicyCheckCommandRunner
	parallelPoolSize         int
}

func (p *PlanCommandRunner) runAutoplan(ctx *command.Context) {
	baseRepo := ctx.Pull.BaseRepo
	pull := ctx.Pull

	projectCmds, err := p.prjCmdBuilder.BuildAutoplanCommands(ctx)
	if err != nil {
		if _, statusErr := p.commitStatusUpdater.UpdateCombined(ctx.RequestCtx, baseRepo, pull, models.FailedCommitStatus, command.Plan, "", ""); statusErr != nil {
			ctx.Log.WarnContext(ctx.RequestCtx, fmt.Sprintf("unable to update commit status: %s", statusErr))
		}
		p.outputUpdater.UpdateOutput(ctx, AutoplanCommand{}, command.Result{Error: err})
		return
	}

	projectCmds, policyCheckCmds := p.partitionProjectCmds(ctx, projectCmds)

	if len(projectCmds) == 0 {
		ctx.Log.InfoContext(ctx.RequestCtx, "determined there was no project to run plan in")
		// If there were no projects modified, we set successful commit statuses
		// with 0/0 projects planned/policy_checked/applied successfully because some users require
		// the Atlantis status to be passing for all pull requests.
		if _, err := p.commitStatusUpdater.UpdateCombinedCount(ctx.RequestCtx, baseRepo, pull, models.SuccessCommitStatus, command.Plan, 0, 0, ""); err != nil {
			ctx.Log.WarnContext(ctx.RequestCtx, fmt.Sprintf("unable to update commit status: %s", err))
		}
		if _, err := p.commitStatusUpdater.UpdateCombinedCount(ctx.RequestCtx, baseRepo, pull, models.SuccessCommitStatus, command.PolicyCheck, 0, 0, ""); err != nil {
			ctx.Log.WarnContext(ctx.RequestCtx, fmt.Sprintf("unable to update commit status: %s", err))
		}
		if _, err := p.commitStatusUpdater.UpdateCombinedCount(ctx.RequestCtx, baseRepo, pull, models.SuccessCommitStatus, command.Apply, 0, 0, ""); err != nil {
			ctx.Log.WarnContext(ctx.RequestCtx, fmt.Sprintf("unable to update commit status: %s", err))
		}
		return
	}

	// At this point we are sure Atlantis has work to do, so set commit status to pending
	statusID, err := p.commitStatusUpdater.UpdateCombined(ctx.RequestCtx, ctx.Pull.BaseRepo, ctx.Pull, models.PendingCommitStatus, command.Plan, "", "")
	if err != nil {
		ctx.Log.WarnContext(ctx.RequestCtx, fmt.Sprintf("unable to update commit status: %s", err))
	}

	// Only run commands in parallel if enabled
	var result command.Result
	if p.isParallelEnabled(projectCmds) {
		ctx.Log.InfoContext(ctx.RequestCtx, "Running plans in parallel")
		result = runProjectCmdsParallel(projectCmds, p.prjCmdRunner.Plan, p.parallelPoolSize)
	} else {
		result = runProjectCmds(projectCmds, p.prjCmdRunner.Plan)
	}

	p.outputUpdater.UpdateOutput(ctx, AutoplanCommand{}, result)

	pullStatus, err := p.dbUpdater.updateDB(ctx, ctx.Pull, result.ProjectResults)
	if err != nil {
		ctx.Log.ErrorContext(ctx.RequestCtx, fmt.Sprintf("writing results: %s", err))
	}

	p.updateCommitStatus(ctx, pullStatus, statusID)

	// Check if there are any planned projects and if there are any errors or if plans are being deleted
	if len(policyCheckCmds) > 0 && !result.HasErrors() {
		// Run policy_check command
		ctx.Log.InfoContext(ctx.RequestCtx, "Running policy_checks for all plans")

		// refresh ctx's view of pull status since we just wrote to it.
		// realistically each command should refresh this at the start,
		// however, policy checking is weird since it's called within the plan command itself
		// we need to better structure how this command works.
		ctx.PullStatus = &pullStatus

		p.policyCheckCommandRunner.Run(ctx, policyCheckCmds)
	}
}

func (p *PlanCommandRunner) run(ctx *command.Context, cmd *command.Comment) {
	var err error
	baseRepo := ctx.Pull.BaseRepo
	pull := ctx.Pull

	// creating status for the first time
	statusID, err := p.commitStatusUpdater.UpdateCombined(ctx.RequestCtx, baseRepo, pull, models.PendingCommitStatus, command.Plan, "", "")
	if err != nil {
		ctx.Log.WarnContext(ctx.RequestCtx, fmt.Sprintf("unable to update commit status: %s", err))
	}

	projectCmds, err := p.prjCmdBuilder.BuildPlanCommands(ctx, cmd)
	if err != nil {
		if _, statusErr := p.commitStatusUpdater.UpdateCombined(ctx.RequestCtx, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, command.Plan, statusID, ""); statusErr != nil {
			ctx.Log.WarnContext(ctx.RequestCtx, fmt.Sprintf("unable to update commit status: %s", statusErr))
		}
		p.outputUpdater.UpdateOutput(ctx, cmd, command.Result{Error: err})
		return
	}

	projectCmds, policyCheckCmds := p.partitionProjectCmds(ctx, projectCmds)

	// Only run commands in parallel if enabled
	var result command.Result
	if p.isParallelEnabled(projectCmds) {
		ctx.Log.InfoContext(ctx.RequestCtx, "Running applies in parallel")
		result = runProjectCmdsParallel(projectCmds, p.prjCmdRunner.Plan, p.parallelPoolSize)
	} else {
		result = runProjectCmds(projectCmds, p.prjCmdRunner.Plan)
	}

	p.outputUpdater.UpdateOutput(
		ctx,
		cmd,
		result)

	pullStatus, err := p.dbUpdater.updateDB(ctx, pull, result.ProjectResults)
	if err != nil {
		ctx.Log.ErrorContext(ctx.RequestCtx, fmt.Sprintf("writing results: %s", err))
		return
	}

	p.updateCommitStatus(ctx, pullStatus, statusID)

	// Runs policy checks step after all plans are successful.
	// This step does not approve any policies that require approval.
	if len(result.ProjectResults) > 0 && !result.HasErrors() {
		ctx.Log.InfoContext(ctx.RequestCtx, fmt.Sprintf("Running policy check for %s", cmd.String()))
		p.policyCheckCommandRunner.Run(ctx, policyCheckCmds)
	}
}

func (p *PlanCommandRunner) Run(ctx *command.Context, cmd *command.Comment) {
	if ctx.Trigger == command.AutoTrigger {
		p.runAutoplan(ctx)
	} else {
		p.run(ctx, cmd)
	}
}

func (p *PlanCommandRunner) updateCommitStatus(ctx *command.Context, pullStatus models.PullStatus, statusID string) {
	var numSuccess int
	var numErrored int
	status := models.SuccessCommitStatus

	numErrored = pullStatus.StatusCount(models.ErroredPlanStatus)
	// We consider anything that isn't a plan error as a plan success.
	// For example, if there is an apply error, that means that at least a
	// plan was generated successfully.
	numSuccess = len(pullStatus.Projects) - numErrored

	if numErrored > 0 {
		status = models.FailedCommitStatus
	}

	if _, err := p.commitStatusUpdater.UpdateCombinedCount(
		ctx.RequestCtx,
		ctx.Pull.BaseRepo,
		ctx.Pull,
		status,
		command.Plan,
		numSuccess,
		len(pullStatus.Projects),
		statusID,
	); err != nil {
		ctx.Log.WarnContext(ctx.RequestCtx, fmt.Sprintf("unable to update commit status: %s", err))
	}
}

func (p *PlanCommandRunner) partitionProjectCmds(
	ctx *command.Context,
	cmds []command.ProjectContext,
) (
	projectCmds []command.ProjectContext,
	policyCheckCmds []command.ProjectContext,
) {
	for _, cmd := range cmds {
		switch cmd.CommandName {
		case command.Plan:
			projectCmds = append(projectCmds, cmd)
		case command.PolicyCheck:
			policyCheckCmds = append(policyCheckCmds, cmd)
		default:
			ctx.Log.ErrorContext(ctx.RequestCtx, fmt.Sprintf("%s is not supported", cmd.CommandName))
		}
	}
	return
}

func (p *PlanCommandRunner) isParallelEnabled(projectCmds []command.ProjectContext) bool {
	return len(projectCmds) > 0 && projectCmds[0].ParallelPlanEnabled
}
