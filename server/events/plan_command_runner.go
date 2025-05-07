package events

import "strconv"

import (
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
)

func NewPlanCommandRunner(
	silenceVCSStatusNoPlans bool,
	silenceVCSStatusNoProjects bool,
	vcsClient vcs.Client,
	pendingPlanFinder PendingPlanFinder,
	workingDir WorkingDir,
	commitStatusUpdater CommitStatusUpdater,
	projectCommandBuilder ProjectPlanCommandBuilder,
	projectCommandRunner ProjectPlanCommandRunner,
	dbUpdater *DBUpdater,
	pullUpdater *PullUpdater,
	policyCheckCommandRunner *PolicyCheckCommandRunner,
	autoMerger *AutoMerger,
	parallelPoolSize int,
	SilenceNoProjects bool,
	pullStatusFetcher PullStatusFetcher,
	lockingLocker locking.Locker,
	discardApprovalOnPlan bool,
	pullReqStatusFetcher vcs.PullReqStatusFetcher,
) *PlanCommandRunner {
	return &PlanCommandRunner{
		silenceVCSStatusNoPlans:    silenceVCSStatusNoPlans,
		silenceVCSStatusNoProjects: silenceVCSStatusNoProjects,
		vcsClient:                  vcsClient,
		pendingPlanFinder:          pendingPlanFinder,
		workingDir:                 workingDir,
		commitStatusUpdater:        commitStatusUpdater,
		prjCmdBuilder:              projectCommandBuilder,
		prjCmdRunner:               projectCommandRunner,
		dbUpdater:                  dbUpdater,
		pullUpdater:                pullUpdater,
		policyCheckCommandRunner:   policyCheckCommandRunner,
		autoMerger:                 autoMerger,
		parallelPoolSize:           parallelPoolSize,
		SilenceNoProjects:          SilenceNoProjects,
		pullStatusFetcher:          pullStatusFetcher,
		lockingLocker:              lockingLocker,
		DiscardApprovalOnPlan:      discardApprovalOnPlan,
		pullReqStatusFetcher:       pullReqStatusFetcher,
	}
}

type PlanCommandRunner struct {
	vcsClient vcs.Client
	// SilenceNoProjects is whether Atlantis should respond to PRs if no projects
	// are found
	SilenceNoProjects bool
	// SilenceVCSStatusNoPlans is whether autoplan should set commit status if no plans
	// are found
	silenceVCSStatusNoPlans bool
	// SilenceVCSStatusNoPlans is whether any plan should set commit status if no projects
	// are found
	silenceVCSStatusNoProjects bool
	commitStatusUpdater        CommitStatusUpdater
	pendingPlanFinder          PendingPlanFinder
	workingDir                 WorkingDir
	prjCmdBuilder              ProjectPlanCommandBuilder
	prjCmdRunner               ProjectPlanCommandRunner
	dbUpdater                  *DBUpdater
	pullUpdater                *PullUpdater
	policyCheckCommandRunner   *PolicyCheckCommandRunner
	autoMerger                 *AutoMerger
	parallelPoolSize           int
	pullStatusFetcher          PullStatusFetcher
	lockingLocker              locking.Locker
	// DiscardApprovalOnPlan controls if all already existing approvals should be removed/dismissed before executing
	// a plan.
	DiscardApprovalOnPlan bool
	pullReqStatusFetcher  vcs.PullReqStatusFetcher
	SilencePRComments     []string
}

func (p *PlanCommandRunner) runAutoplan(ctx *command.Context) {
	baseRepo := ctx.Pull.BaseRepo
	pull := ctx.Pull

	projectCmds, err := p.prjCmdBuilder.BuildAutoplanCommands(ctx)
	if err != nil {
		if statusErr := p.commitStatusUpdater.UpdateCombined(ctx.Log, baseRepo, pull, models.FailedCommitStatus, command.Plan); statusErr != nil {
			ctx.Log.Warn("unable to update commit status: %s", statusErr)
		}
		p.pullUpdater.updatePull(ctx, AutoplanCommand{}, command.Result{Error: err})
		return
	}

	projectCmds, policyCheckCmds := p.partitionProjectCmds(ctx, projectCmds)

	if len(projectCmds) == 0 {
		ctx.Log.Info("determined there was no project to run plan in")
		if !(p.silenceVCSStatusNoPlans || p.silenceVCSStatusNoProjects) {
			// If there were no projects modified, we set successful commit statuses
			// with 0/0 projects planned/policy_checked/applied successfully because some users require
			// the Atlantis status to be passing for all pull requests.
			ctx.Log.Debug("setting VCS status to success with no projects found")
			if err := p.commitStatusUpdater.UpdateCombinedCount(ctx.Log, baseRepo, pull, models.SuccessCommitStatus, command.Plan, 0, 0); err != nil {
				ctx.Log.Warn("unable to update commit status: %s", err)
			}
			if err := p.commitStatusUpdater.UpdateCombinedCount(ctx.Log, baseRepo, pull, models.SuccessCommitStatus, command.PolicyCheck, 0, 0); err != nil {
				ctx.Log.Warn("unable to update commit status: %s", err)
			}
			if err := p.commitStatusUpdater.UpdateCombinedCount(ctx.Log, baseRepo, pull, models.SuccessCommitStatus, command.Apply, 0, 0); err != nil {
				ctx.Log.Warn("unable to update commit status: %s", err)
			}
		}
		return
	}

	// discard previous plans that might not be relevant anymore
	ctx.Log.Debug("deleting previous plans and locks")
	p.deletePlans(ctx)
	_, err = p.lockingLocker.UnlockByPull(baseRepo.FullName, pull.Num)
	if err != nil {
		ctx.Log.Err("deleting locks: %s", err)
	}

	// Only run commands in parallel if enabled
	var result command.Result
	if p.isParallelEnabled(projectCmds) {
		ctx.Log.Info("Running plans in parallel")
		result = runProjectCmdsParallelGroups(ctx, projectCmds, p.prjCmdRunner.Plan, p.parallelPoolSize)
	} else {
		result = runProjectCmds(projectCmds, p.prjCmdRunner.Plan)
	}

	if p.autoMerger.automergeEnabled(projectCmds) && result.HasErrors() {
		ctx.Log.Info("deleting plans because there were errors and automerge requires all plans succeed")
		p.deletePlans(ctx)
		result.PlansDeleted = true
	}

	p.pullUpdater.updatePull(ctx, AutoplanCommand{}, result)

	pullStatus, err := p.dbUpdater.updateDB(ctx, ctx.Pull, result.ProjectResults)
	if err != nil {
		ctx.Log.Err("writing results: %s", err)
	}

	p.updateCommitStatus(ctx, pullStatus, command.Plan)
	p.updateCommitStatus(ctx, pullStatus, command.Apply)

	// Check if there are any planned projects and if there are any errors or if plans are being deleted
	if len(policyCheckCmds) > 0 &&
		!(result.HasErrors() || result.PlansDeleted) {
		// Run policy_check command
		ctx.Log.Info("Running policy_checks for all plans")

		// refresh ctx's view of pull status since we just wrote to it.
		// realistically each command should refresh this at the start,
		// however, policy checking is weird since it's called within the plan command itself
		// we need to better structure how this command works.
		ctx.PullStatus = &pullStatus

		p.policyCheckCommandRunner.Run(ctx, policyCheckCmds)
	}
}

func (p *PlanCommandRunner) run(ctx *command.Context, cmd *CommentCommand) {
	var err error
	baseRepo := ctx.Pull.BaseRepo
	pull := ctx.Pull

	ctx.PullRequestStatus, err = p.pullReqStatusFetcher.FetchPullStatus(ctx.Log, pull)
	if err != nil {
		// On error we continue the request with mergeable assumed false.
		// We want to continue because not all apply's will need this status,
		// only if they rely on the mergeability requirement.
		// All PullRequestStatus fields are set to false by default when error.
		ctx.Log.Warn("unable to get pull request status: %s. Continuing with mergeable and approved assumed false", err)
	}

	if p.DiscardApprovalOnPlan {
		if err = p.pullUpdater.VCSClient.DiscardReviews(ctx.Log, baseRepo, pull); err != nil {
			ctx.Log.Err("failed to remove approvals: %s", err)
		}
	}

	projectCmds, err := p.prjCmdBuilder.BuildPlanCommands(ctx, cmd)
	if err != nil {
		if statusErr := p.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, command.Plan); statusErr != nil {
			ctx.Log.Warn("unable to update commit status: %s", statusErr)
		}
		p.pullUpdater.updatePull(ctx, cmd, command.Result{Error: err})
		return
	}

	if len(projectCmds) == 0 && p.SilenceNoProjects {
		ctx.Log.Info("determined there was no project to run plan in")
		if !p.silenceVCSStatusNoProjects {
			if cmd.IsForSpecificProject() {
				// With a specific plan, just reset the status so it's not stuck in pending state
				pullStatus, err := p.pullStatusFetcher.GetPullStatus(pull)
				if err != nil {
					ctx.Log.Warn("unable to fetch pull status: %s", err)
					return
				}
				if pullStatus == nil {
					// default to 0/0
					ctx.Log.Debug("setting VCS status to 0/0 success as no previous state was found")
					if err := p.commitStatusUpdater.UpdateCombinedCount(ctx.Log, baseRepo, pull, models.SuccessCommitStatus, command.Plan, 0, 0); err != nil {
						ctx.Log.Warn("unable to update commit status: %s", err)
					}
					return
				}
				ctx.Log.Debug("resetting VCS status")
				p.updateCommitStatus(ctx, *pullStatus, command.Plan)
			} else {
				// With a generic plan, we set successful commit statuses
				// with 0/0 projects planned successfully because some users require
				// the Atlantis status to be passing for all pull requests.
				// Does not apply to skipped runs for specific projects
				ctx.Log.Debug("setting VCS status to success with no projects found")
				if err := p.commitStatusUpdater.UpdateCombinedCount(ctx.Log, baseRepo, pull, models.SuccessCommitStatus, command.Plan, 0, 0); err != nil {
					ctx.Log.Warn("unable to update commit status: %s", err)
				}
				if err := p.commitStatusUpdater.UpdateCombinedCount(ctx.Log, baseRepo, pull, models.SuccessCommitStatus, command.PolicyCheck, 0, 0); err != nil {
					ctx.Log.Warn("unable to update commit status: %s", err)
				}
				if err := p.commitStatusUpdater.UpdateCombinedCount(ctx.Log, baseRepo, pull, models.SuccessCommitStatus, command.Apply, 0, 0); err != nil {
					ctx.Log.Warn("unable to update commit status: %s", err)
				}
			}
		}
		return
	}

	projectCmds, policyCheckCmds := p.partitionProjectCmds(ctx, projectCmds)

	// if the plan is generic, new plans will be generated based on changes
	// discard previous plans that might not be relevant anymore
	if !cmd.IsForSpecificProject() {
		ctx.Log.Debug("deleting previous plans and locks")
		p.deletePlans(ctx)
	}

	// Only run commands in parallel if enabled
	var result command.Result
	if p.isParallelEnabled(projectCmds) {
		ctx.Log.Info("Running plans in parallel")
		result = runProjectCmdsParallelGroups(ctx, projectCmds, p.prjCmdRunner.Plan, p.parallelPoolSize)
	} else {
		result = runProjectCmds(projectCmds, p.prjCmdRunner.Plan)
	}
	ctx.CommandHasErrors = result.HasErrors()

	for i, projResult := range result.ProjectResults {
		projCtx := projectCmds[i]

		if projResult.PlanStatus() == models.PlannedNoChangesPlanStatus || projResult.PlanStatus() == models.ErroredPlanStatus {
			ctx.Log.Info("Keeping lock for project '%s' (no changes or error)", projCtx.ProjectName)
			continue
		}

		// delete lock only if there are changes
		ctx.Log.Info("Deleting lock for project '%s' (changes detected)", projCtx.ProjectName)
		lockID := projCtx.BaseRepo.FullName + "/" + strconv.Itoa(projCtx.Pull.Num) + "/" + projCtx.ProjectName + "/" + projCtx.Workspace

		_, err := p.lockingLocker.Unlock(lockID)
		if err != nil {
			ctx.Log.Err("failed unlocking project '%s': %s", projCtx.ProjectName, err)
		}
	}

	if p.autoMerger.automergeEnabled(projectCmds) && result.HasErrors() {
		ctx.Log.Info("deleting plans because there were errors and automerge requires all plans succeed")
		p.deletePlans(ctx)
		result.PlansDeleted = true
	}

	p.pullUpdater.updatePull(
		ctx,
		cmd,
		result)

	pullStatus, err := p.dbUpdater.updateDB(ctx, pull, result.ProjectResults)
	if err != nil {
		ctx.Log.Err("writing results: %s", err)
		return
	}

	p.updateCommitStatus(ctx, pullStatus, command.Plan)
	p.updateCommitStatus(ctx, pullStatus, command.Apply)

	// Runs policy checks step after all plans are successful.
	// This step does not approve any policies that require approval.
	if len(result.ProjectResults) > 0 &&
		!(result.HasErrors() || result.PlansDeleted) {
		ctx.Log.Info("Running policy check for '%s'", cmd.CommandName())
		p.policyCheckCommandRunner.Run(ctx, policyCheckCmds)
	} else if len(projectCmds) == 0 && !cmd.IsForSpecificProject() {
		// If there were no projects modified, we set successful commit statuses
		// with 0/0 projects planned/policy_checked/applied successfully because some users require
		// the Atlantis status to be passing for all pull requests.
		ctx.Log.Debug("setting VCS status to success with no projects found")
		if err := p.commitStatusUpdater.UpdateCombinedCount(ctx.Log, baseRepo, pull, models.SuccessCommitStatus, command.PolicyCheck, 0, 0); err != nil {
			ctx.Log.Warn("unable to update commit status: %s", err)
		}
	}
}

func (p *PlanCommandRunner) Run(ctx *command.Context, cmd *CommentCommand) {
	if ctx.Trigger == command.AutoTrigger {
		p.runAutoplan(ctx)
	} else {
		p.run(ctx, cmd)
	}
}

func (p *PlanCommandRunner) updateCommitStatus(ctx *command.Context, pullStatus models.PullStatus, commandName command.Name) {
	var numSuccess int
	var numErrored int
	status := models.SuccessCommitStatus

	if commandName == command.Plan {
		numErrored = pullStatus.StatusCount(models.ErroredPlanStatus)
		// We consider anything that isn't a plan error as a plan success.
		// For example, if there is an apply error, that means that at least a
		// plan was generated successfully.
		numSuccess = len(pullStatus.Projects) - numErrored

		if numErrored > 0 {
			status = models.FailedCommitStatus
		}
	} else if commandName == command.Apply {
		numSuccess = pullStatus.StatusCount(models.AppliedPlanStatus) + pullStatus.StatusCount(models.PlannedNoChangesPlanStatus)
		numErrored = pullStatus.StatusCount(models.ErroredApplyStatus)

		if numErrored > 0 {
			status = models.FailedCommitStatus
		} else if numSuccess < len(pullStatus.Projects) {
			// If there are plans that haven't been applied yet, no need to update the status
			return
		}
	}

	if err := p.commitStatusUpdater.UpdateCombinedCount(
		ctx.Log,
		ctx.Pull.BaseRepo,
		ctx.Pull,
		status,
		commandName,
		numSuccess,
		len(pullStatus.Projects),
	); err != nil {
		ctx.Log.Warn("unable to update commit status: %s", err)
	}
}

// deletePlans deletes all plans generated in this ctx.
func (p *PlanCommandRunner) deletePlans(ctx *command.Context) {
	pullDir, err := p.workingDir.GetPullDir(ctx.Pull.BaseRepo, ctx.Pull)
	if err != nil {
		ctx.Log.Err("getting pull dir: %s", err)
	}
	if err := p.pendingPlanFinder.DeletePlans(pullDir); err != nil {
		ctx.Log.Err("deleting pending plans: %s", err)
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
			ctx.Log.Err("%s is not supported", cmd.CommandName)
		}
	}
	return
}

func (p *PlanCommandRunner) isParallelEnabled(projectCmds []command.ProjectContext) bool {
	return len(projectCmds) > 0 && projectCmds[0].ParallelPlanEnabled
}
