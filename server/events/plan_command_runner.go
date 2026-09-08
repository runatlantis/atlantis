// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/google/uuid"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/utils"
)

// GenerateLockID creates a consistent lock ID for a project context.
// This ensures the same format is used for both locking and unlocking operations.
func GenerateLockID(projCtx command.ProjectContext) string {
	// Use models.NewProject to ensure consistent path cleaning
	project := models.NewProject(projCtx.BaseRepo.FullName, projCtx.RepoRelDir, projCtx.ProjectName)
	return models.GenerateLockKey(project, projCtx.Workspace)
}

func NewPlanCommandRunner(
	silenceVCSStatusNoPlans bool,
	silenceVCSStatusNoProjects bool,
	vcsClient vcs.Client,
	pendingPlanFinder PendingPlanFinder,
	workingDir WorkingDir,
	workingDirLocker WorkingDirLocker,
	commitStatusUpdater CommitStatusUpdater,
	projectCommandBuilder ProjectPlanCommandBuilder,
	projectCommandRunner ProjectPlanCommandRunner,
	cancellationTracker CancellationTracker,
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
	PendingApplyStatus bool,

) *PlanCommandRunner {
	return &PlanCommandRunner{
		silenceVCSStatusNoPlans:    silenceVCSStatusNoPlans,
		silenceVCSStatusNoProjects: silenceVCSStatusNoProjects,
		vcsClient:                  vcsClient,
		pendingPlanFinder:          pendingPlanFinder,
		workingDir:                 workingDir,
		workingDirLocker:           workingDirLocker,
		commitStatusUpdater:        commitStatusUpdater,
		prjCmdBuilder:              projectCommandBuilder,
		prjCmdRunner:               projectCommandRunner,
		cancellationTracker:        cancellationTracker,
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
		PendingApplyStatus:         PendingApplyStatus,
	}
}

type PlanCommandRunner struct {
	Publication         *PublicationCoordinator
	LivePullHeadFetcher LivePullHeadFetcher
	PlanReaper          PlanArtifactReaper
	vcsClient           vcs.Client
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
	workingDirLocker           WorkingDirLocker
	prjCmdBuilder              ProjectPlanCommandBuilder
	prjCmdRunner               ProjectPlanCommandRunner
	cancellationTracker        CancellationTracker
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
	PendingApplyStatus    bool
}

func (p *PlanCommandRunner) runAutoplan(ctx *command.Context) {
	pull := ctx.Pull
	unlockPullPlan, ok := p.lockPullForPlan(ctx, AutoplanCommand{})
	if !ok {
		return
	}
	defer unlockPullPlan()

	var err error
	ctx.PullRequestStatus, err = p.pullReqStatusFetcher.FetchPullStatus(ctx.Log, pull)
	if err != nil {
		// On error we continue the request with mergeable assumed false.
		// We want to continue because not all plan's will need this status,
		// only if they rely on the mergeability requirement.
		// All PullRequestStatus fields are set to false by default when error.
		ctx.Log.Warn("unable to get pull request status: %s. Continuing with mergeable and approved assumed false", err)
	}

	projectCmds, err := p.prjCmdBuilder.BuildAutoplanCommands(ctx)
	if err != nil {
		p.reportPlanStartFailure(ctx, AutoplanCommand{}, err)
		return
	}

	projectCmds, policyCheckCmds := p.partitionProjectCmds(ctx, projectCmds)

	if len(projectCmds) == 0 {
		p.handleNoProjectPlan(ctx, AutoplanCommand{})
		return
	}

	if !p.preparePlan(ctx, AutoplanCommand{}, projectCmds, true) {
		return
	}

	result := runProjectCmdsWithCancellationTracker(ctx, projectCmds, p.cancellationTracker, p.parallelPoolSize, p.isParallelEnabled(projectCmds), p.prjCmdRunner.Plan)

	if p.autoMerger.automergeEnabled(projectCmds) && result.HasErrors() {
		ctx.Log.Info("deleting plans because there were errors and automerge requires all plans succeed")
		if err := p.deletePlansAndPlanLocks(ctx, projectCmds); err != nil {
			ctx.Log.Err("deleting pending plans: %s", err)
		}
		result.PlansDeleted = true
		for i := range result.ProjectResults {
			project := &result.ProjectResults[i]
			if project.PlanGeneration != "" && project.Error == nil && project.Failure == "" {
				project.Failure = "plan discarded because automerge requires all plans to succeed; run `atlantis plan` again"
			}
		}
	}

	pullStatus, completed := p.completePlan(ctx, AutoplanCommand{}, projectCmds, result)
	if !completed {
		return
	}

	// Check if there are any planned projects and if there are any errors or if plans are being deleted
	if len(policyCheckCmds) > 0 &&
		(!result.HasErrors() && !result.PlansDeleted) {
		// Run policy_check command
		ctx.Log.Info("Running policy_checks for all plans")

		// refresh ctx's view of pull status since we just wrote to it.
		// realistically each command should refresh this at the start,
		// however, policy checking is weird since it's called within the plan command itself
		// we need to better structure how this command works.
		ctx.PullStatus = &pullStatus
		bindAcceptedGenerations(policyCheckCmds, &pullStatus)
		p.policyCheckCommandRunner.Run(ctx, policyCheckCmds)
	}
}

func (p *PlanCommandRunner) run(ctx *command.Context, cmd *CommentCommand) {
	var err error
	baseRepo := ctx.Pull.BaseRepo
	pull := ctx.Pull
	unlockPullPlan, ok := p.lockPullForPlan(ctx, cmd)
	if !ok {
		return
	}
	defer unlockPullPlan()

	ctx.PullRequestStatus, err = p.pullReqStatusFetcher.FetchPullStatus(ctx.Log, pull)
	if err != nil {
		// On error we continue the request with mergeable assumed false.
		// We want to continue because not all apply's will need this status,
		// only if they rely on the mergeability requirement.
		// All PullRequestStatus fields are set to false by default when error.
		ctx.Log.Warn("unable to get pull request status: %s. Continuing with mergeable and approved assumed false", err)
	}

	projectCmds, err := p.prjCmdBuilder.BuildPlanCommands(ctx, cmd)
	if MarkCommandSkippedIfIgnoredTargetedDir(ctx, command.Plan, err) {
		return
	}

	if p.DiscardApprovalOnPlan {
		if discardErr := p.pullUpdater.VCSClient.DiscardReviews(ctx.Log, baseRepo, pull); discardErr != nil {
			ctx.Log.Err("failed to remove approvals: %s", discardErr)
		}
	}

	if err != nil {
		p.reportPlanStartFailure(ctx, cmd, err)
		return
	}

	projectCmds, policyCheckCmds := p.partitionProjectCmds(ctx, projectCmds)
	if len(projectCmds) == 0 {
		p.handleNoProjectPlan(ctx, cmd)
		return
	}

	if !p.preparePlan(ctx, cmd, projectCmds, !cmd.IsForSpecificProject()) {
		return
	}

	result := runProjectCmdsWithCancellationTracker(ctx, projectCmds, p.cancellationTracker, p.parallelPoolSize, p.isParallelEnabled(projectCmds), p.prjCmdRunner.Plan)
	ctx.CommandHasErrors = result.HasErrors()

	if p.autoMerger.automergeEnabled(projectCmds) && result.HasErrors() {
		ctx.Log.Info("deleting plans because there were errors and automerge requires all plans succeed")
		if err := p.deletePlansAndPlanLocks(ctx, projectCmds); err != nil {
			ctx.Log.Err("deleting pending plans: %s", err)
		}
		result.PlansDeleted = true
		for i := range result.ProjectResults {
			project := &result.ProjectResults[i]
			if project.PlanGeneration != "" && project.Error == nil && project.Failure == "" {
				project.Failure = "plan discarded because automerge requires all plans to succeed; run `atlantis plan` again"
			}
		}
	}

	pullStatus, completed := p.completePlan(ctx, cmd, projectCmds, result)
	if !completed {
		return
	}

	// Runs policy checks step after all plans are successful.
	// This step does not approve any policies that require approval.
	if len(result.ProjectResults) > 0 &&
		(!result.HasErrors() && !result.PlansDeleted) {
		ctx.Log.Info("Running policy check for '%s'", cmd.CommandName())
		ctx.PullStatus = &pullStatus
		bindAcceptedGenerations(policyCheckCmds, &pullStatus)
		p.policyCheckCommandRunner.Run(ctx, policyCheckCmds)
	}
}

func (p *PlanCommandRunner) Run(ctx *command.Context, cmd *CommentCommand) {
	if ctx.Trigger == command.AutoTrigger {
		p.runAutoplan(ctx)
	} else {
		p.run(ctx, cmd)
	}
}

func (p *PlanCommandRunner) clearPlansAndPullStatusForNoProjects(ctx *command.Context, pull models.PullRequest) (models.PullStatus, error) {
	// One atomic replacement supersedes old generations. Delete + an ordinary
	// empty update could otherwise erase a generation admitted in between.
	admitted, err := p.dbUpdater.Database.BeginPlanGeneration(pull, uuid.NewString(), nil, true, ctx.PublicationMode())
	if err != nil {
		return models.PullStatus{}, fmt.Errorf("writing empty plan status: %w", err)
	}
	if _, err := p.deletePlansAndPendingPlanLocks(ctx); err != nil {
		return models.PullStatus{}, err
	}
	// A replica can have durable project locks even after losing its checkout.
	// Release the captured old projects as well as locally discovered plans.
	for _, old := range admitted.Previous {
		if p.PlanReaper != nil {
			if err := p.PlanReaper.ReapPlan(ctx.Log, pull, old); err != nil {
				return models.PullStatus{}, err
			}
		}
		project := models.NewProject(pull.BaseRepo.FullName, old.RepoRelDir, old.ProjectName)
		if err := p.unlockPlanLockIfOwnedByPull(ctx, project, old.Workspace, models.GenerateLockKey(project, old.Workspace)); err != nil {
			return models.PullStatus{}, err
		}
	}
	return admitted.PullStatus, nil
}

func (p *PlanCommandRunner) lockPullForPlan(ctx *command.Context, cmd PullCommand) (func(), bool) {
	if p.workingDirLocker == nil {
		return func() {}, true
	}
	unlockFn, err := p.workingDirLocker.TryLockPull(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, command.Plan, WorkingDirLockMetadataForPull(ctx.Pull))
	if err != nil {
		p.handleNoProjectPlanStateError(ctx, cmd, err)
		return nil, false
	}
	return unlockFn, true
}

func (p *PlanCommandRunner) handleNoProjectPlanStateError(ctx *command.Context, cmd PullCommand, err error) {
	p.reportPlanStartFailure(ctx, cmd, err)
}

func (p *PlanCommandRunner) reportPlanStartFailure(ctx *command.Context, cmd PullCommand, failure error) {
	ctx.CommandHasErrors = true
	var database db.Database
	if p.dbUpdater != nil {
		database = p.dbUpdater.Database
	}
	err := p.Publication.RunWithObservedStatus(ctx, database, p.LivePullHeadFetcher, func() error {
		if err := publishTerminal(ctx, func() error {
			return p.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, command.Plan)
		}); err != nil {
			return err
		}
		return p.pullUpdater.updatePull(ctx, cmd, command.Result{Error: failure})
	})
	if err != nil && !p.obsoletePlan(ctx, err) {
		if reportErr := p.pullUpdater.updatePull(ctx, cmd, command.Result{Error: errors.Join(failure, err)}); reportErr != nil {
			ctx.Log.Err("reporting command result: %s", reportErr)
		}
	}
}

func (p *PlanCommandRunner) ShouldSkipPreWorkflowHooks(ctx *command.Context, cmd *CommentCommand) bool {
	return MarkCommandSkippedIfIgnoredTarget(ctx, command.Plan, cmd, p.prjCmdBuilder)
}

func (p *PlanCommandRunner) updatePendingCommitStatus(ctx *command.Context, commandName command.Name) error {
	if p.silenceVCSStatusNoProjects {
		ctx.Log.Debug("silence enabled - not setting pending VCS status")
		return nil
	}
	if err := publishTerminal(ctx, func() error {
		return p.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.PendingCommitStatus, commandName)
	}); err != nil {
		ctx.Log.Warn("unable to update commit status: %s", err)
		return err
	}
	return nil
}

func (p *PlanCommandRunner) updateCommitStatus(ctx *command.Context, pullStatus models.PullStatus, commandName command.Name) error {
	var numSuccess int
	var numErrored int
	var numNoChanges int
	status := models.SuccessCommitStatus

	switch commandName {
	case command.Plan:
		numErrored = pullStatus.StatusCount(models.ErroredPlanStatus)
		// We consider anything that isn't a plan error as a plan success.
		// For example, if there is an apply error, that means that at least a
		// plan was generated successfully.
		numSuccess = len(pullStatus.Projects) - numErrored

		if numErrored > 0 {
			status = models.FailedCommitStatus
		}
	case command.Apply:
		numNoChanges = pullStatus.StatusCount(models.PlannedNoChangesPlanStatus)
		numSuccess = pullStatus.StatusCount(models.AppliedPlanStatus) + numNoChanges
		numErrored = pullStatus.StatusCount(models.ErroredApplyStatus)

		if numErrored > 0 {
			status = models.FailedCommitStatus
		} else if numSuccess < len(pullStatus.Projects) {
			// When there are planned changes that haven't been applied yet:
			// - GitLab: Set status to pending if PendingApplyStatus is enabled
			//           This prevents MR merging until all applies complete
			// - Other VCS: Leave status unchanged (existing behavior)
			if ctx.Pull.BaseRepo.VCSHost.Type == models.Gitlab && p.PendingApplyStatus {
				ctx.Log.Debug("Pending Apply Status is set. Pipeline status will be marked as pending since there are changes to apply")
				status = models.PendingCommitStatus
			} else {
				if p.PendingApplyStatus {
					// If a VCS uses this flag other than Gitlab, we log the warning to the user
					ctx.Log.Warn("Flag --pending-apply-status is not yet supported by your VCS. Pipeline status will not be marked as pending")
				}
				// Otherwise, status remains SuccessCommitStatus (no update needed)
				return nil
			}
		}
	}

	if err := publishTerminal(ctx, func() error {
		return p.commitStatusUpdater.UpdateCombinedCount(
			ctx.Log,
			ctx.Pull.BaseRepo,
			ctx.Pull,
			status,
			commandName,
			models.ProjectCounts{Success: numSuccess, Total: len(pullStatus.Projects), Errored: numErrored, NoChanges: numNoChanges},
		)
	}); err != nil {
		ctx.Log.Warn("unable to update commit status: %s", err)
		return err
	}
	return nil
}

// deletePlans deletes all plans generated in this ctx.
func (p *PlanCommandRunner) deletePlans(ctx *command.Context) ([]PendingPlan, error) {
	return p.deletePlansWithPostDelete(ctx, nil)
}

func (p *PlanCommandRunner) deletePlansAndPendingPlanLocks(ctx *command.Context) ([]PendingPlan, error) {
	return p.deletePlansWithPostDelete(ctx, p.deletePlanLocksForPendingPlans)
}

func (p *PlanCommandRunner) deletePlansWithPostDelete(ctx *command.Context, postDelete func(*command.Context, []PendingPlan) error) ([]PendingPlan, error) {
	pullDir, err := p.workingDir.GetPullDir(ctx.Pull.BaseRepo, ctx.Pull)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting pull dir: %w", err)
	}
	plans, err := p.pendingPlanFinder.Find(pullDir)
	if err != nil {
		return nil, fmt.Errorf("finding pending plans: %w", err)
	}

	var unlocks []func()
	defer func() {
		for _, unlock := range slices.Backward(unlocks) {
			unlock()
		}
	}()
	for _, plan := range plans {
		unlockFn, err := p.workingDirLocker.TryLock(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, plan.Workspace, plan.RepoRelDir, plan.ProjectName, command.Plan, WorkingDirLockMetadataForPull(ctx.Pull))
		if err != nil {
			return nil, fmt.Errorf("locking pending plan for dir %q workspace %q project %q before deleting: %w", plan.RepoRelDir, plan.Workspace, plan.ProjectName, err)
		}
		unlocks = append(unlocks, unlockFn)
	}

	for _, plan := range plans {
		planPath := filepath.Join(plan.planRepoDir(), plan.RepoRelDir, runtime.GetPlanFilename(plan.Workspace, plan.ProjectName))
		if err := utils.RemoveIgnoreNonExistent(planPath); err != nil {
			return nil, fmt.Errorf("deleting plan at %s: %w", planPath, err)
		}
	}
	if postDelete != nil {
		if err := postDelete(ctx, plans); err != nil {
			return nil, err
		}
	}
	return plans, nil
}

func (p *PlanCommandRunner) deletePlansAndPlanLocks(ctx *command.Context, projectCmds []command.ProjectContext) error {
	if _, err := p.deletePlans(ctx); err != nil {
		return err
	}
	return p.deletePlanLocks(ctx, projectCmds)
}

func (p *PlanCommandRunner) deletePlanLocks(ctx *command.Context, projectCmds []command.ProjectContext) error {
	unlocked := make(map[string]bool)
	for _, projCtx := range projectCmds {
		if projCtx.RepoLocksMode != valid.RepoLocksOnPlanMode {
			continue
		}

		lockKey := GenerateLockID(projCtx)
		if unlocked[lockKey] {
			continue
		}
		unlocked[lockKey] = true

		project := models.NewProject(projCtx.BaseRepo.FullName, projCtx.RepoRelDir, projCtx.ProjectName)
		if err := p.unlockPlanLockIfOwnedByPull(ctx, project, projCtx.Workspace, lockKey); err != nil {
			return err
		}
	}
	return nil
}

func (p *PlanCommandRunner) deletePlanLocksForPendingPlans(ctx *command.Context, plans []PendingPlan) error {
	unlocked := make(map[string]bool)
	for _, plan := range plans {
		project := models.NewProject(ctx.Pull.BaseRepo.FullName, plan.RepoRelDir, plan.ProjectName)
		lockKey := models.GenerateLockKey(project, plan.Workspace)
		if unlocked[lockKey] {
			continue
		}
		unlocked[lockKey] = true
		if err := p.unlockPlanLockIfOwnedByPull(ctx, project, plan.Workspace, lockKey); err != nil {
			return err
		}
	}
	return nil
}

func (p *PlanCommandRunner) unlockPlanLockIfOwnedByPull(ctx *command.Context, project models.Project, workspace string, lockKey string) error {
	if _, err := p.lockingLocker.UnlockIfOwnedByPull(project, workspace, ctx.Pull.Num); err != nil {
		return fmt.Errorf("deleting lock %q for pull %d: %w", lockKey, ctx.Pull.Num, err)
	}
	return nil
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

// Successful project checks are deferred until the command's results are durable.
func (p *PlanCommandRunner) publishPlanStatuses(projectCmds []command.ProjectContext, result command.Result, status models.CommitStatus) error {
	if publisher, ok := p.prjCmdRunner.(DeferredPlanStatusPublisher); ok {
		return publisher.PublishDeferredPlanStatuses(projectCmds, result, status)
	}
	return nil
}

func (p *PlanCommandRunner) planPersistenceFailed(ctx *command.Context, cmd PullCommand, projectCmds []command.ProjectContext, result command.Result, err error) error {
	ctx.CommandHasErrors = true
	result.Error = fmt.Errorf("persisting plan results: %w; restore database connectivity and run `atlantis plan` again before applying", err)
	if err := publishTerminal(ctx, func() error { return p.publishPlanStatuses(projectCmds, result, models.FailedCommitStatus) }); err != nil {
		return err
	}
	for _, name := range []command.Name{command.Plan, command.Apply} {
		if err := publishTerminal(ctx, func() error {
			return p.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, name)
		}); err != nil {
			return err
		}
	}
	return p.pullUpdater.updatePull(ctx, cmd, result)
}

// beginGeneration makes every selected project non-applyable before a plan step
// can replace its artifact. Synthetic non-PR API operations keep their legacy
// request-local status rather than entering the durable PR generation protocol.
func (p *PlanCommandRunner) beginGeneration(ctx *command.Context, cmd PullCommand, projects []command.ProjectContext, replace bool) bool {
	if ctx.Pull.Num <= 0 || len(projects) == 0 {
		return true
	}
	generation := uuid.NewString()
	admitted, err := p.dbUpdater.Database.BeginPlanGeneration(ctx.Pull, generation, projects, replace, ctx.PublicationMode())
	if err != nil {
		if reportErr := p.planPersistenceFailed(ctx, cmd, projects, command.Result{}, fmt.Errorf("admitting plan generation: %w", err)); reportErr != nil {
			ctx.Log.Err("reporting plan persistence error: %s", reportErr)
		}
		return false
	}
	ctx.PullStatus = &admitted.PullStatus
	for i := range projects {
		projects[i].PlanGeneration = generation
		projects[i].AcceptedPlanGeneration = ""
		projects[i].ExpectedPlanHash = ""
		projects[i].SavedPlanHash = new(string)
		projects[i].PullStatus = &admitted.PullStatus
	}
	if p.PlanReaper != nil {
		for _, old := range admitted.Previous {
			current := findProjectInPullStatus(&admitted.PullStatus, old.Workspace, old.RepoRelDir, old.ProjectName)
			if current != nil && current.PlanGeneration == old.PlanGeneration {
				continue
			}
			if err := p.PlanReaper.ReapPlan(ctx.Log, ctx.Pull, old); err != nil {
				if reportErr := p.planPersistenceFailed(ctx, cmd, projects, command.Result{}, fmt.Errorf("reaping superseded plan: %w", err)); reportErr != nil {
					ctx.Log.Err("reporting plan persistence error: %s", reportErr)
				}
				return false
			}
		}
	}
	return true
}

func bindAcceptedGenerations(projects []command.ProjectContext, status *models.PullStatus) {
	for i := range projects {
		project := findProjectInPullStatus(status, projects[i].Workspace, projects[i].RepoRelDir, projects[i].ProjectName)
		if project != nil {
			projects[i].PlanGeneration = project.PlanGeneration
			projects[i].AcceptedPlanGeneration = project.AcceptedPlanGeneration
			projects[i].ExpectedPlanHash = project.ManagedPlanHash
			projects[i].PullStatus = status
		}
	}
}

func (p *PlanCommandRunner) obsoletePlan(ctx *command.Context, err error) bool {
	if !errors.Is(err, db.ErrPlanGenerationSuperseded) {
		return false
	}
	ctx.CommandHasErrors = true
	ctx.Log.Warn("suppressing obsolete plan publication %v", err)
	return true
}

// preparePlan releases its lease before the caller invokes any plan steps.
func (p *PlanCommandRunner) preparePlan(ctx *command.Context, cmd PullCommand, projects []command.ProjectContext, discardPrevious bool) bool {
	started := false
	err := p.Publication.RunCommand(ctx, func() error {
		if err := p.refreshPublicationHead(ctx); err != nil {
			return err
		}
		if !p.beginGeneration(ctx, cmd, projects, false) {
			return nil
		}
		if len(projects) > 0 {
			if err := p.updatePendingCommitStatus(ctx, command.Plan); err != nil {
				return err
			}
		}
		if discardPrevious && len(projects) > 0 {
			if err := p.deletePlansAndPlanLocks(ctx, projects); err != nil {
				return err
			}
		}
		if ctx.TerminalPublisher != nil {
			for i := range projects {
				projects[i].PublicationDeferred = true
			}
			if publisher, ok := p.prjCmdRunner.(PendingProjectStatusPublisher); ok {
				if err := publishTerminal(ctx, func() error { return publisher.PublishPendingProjectStatuses(projects) }); err != nil {
					return err
				}
			}
		}
		started = true
		return nil
	})
	if err != nil {
		ctx.CommandHasErrors = true
		if reportErr := p.pullUpdater.updatePull(ctx, cmd, command.Result{Error: fmt.Errorf("starting plan publication: %w", err)}); reportErr != nil {
			ctx.Log.Err("reporting command result: %s", reportErr)
		}
		return false
	}
	return started
}

// completePlan reacquires ownership after execution; generation checks and all
// terminal publication happen before this short lease is released.
func (p *PlanCommandRunner) completePlan(ctx *command.Context, cmd PullCommand, projects []command.ProjectContext, result command.Result) (models.PullStatus, bool) {
	var status models.PullStatus
	reported := false
	err := p.Publication.RunCommand(ctx, func() error {
		if err := p.refreshPublicationHead(ctx); err != nil {
			return err
		}
		var err error
		status, err = p.dbUpdater.updateDB(ctx, ctx.Pull, result.ProjectResults)
		if err != nil {
			if !errors.Is(err, db.ErrPlanGenerationSuperseded) {
				publicationErr := p.planPersistenceFailed(ctx, cmd, projects, result, err)
				reported = true
				return errors.Join(err, publicationErr)
			}
			return err
		}
		ctx.PullStatus = &status
		if err := publishTerminal(ctx, func() error { return p.publishPlanStatuses(projects, result, models.SuccessCommitStatus) }); err != nil {
			return err
		}
		if err := p.pullUpdater.updatePull(ctx, cmd, result); err != nil {
			return err
		}
		if err := p.updateCommitStatus(ctx, status, command.Plan); err != nil {
			return err
		}
		return p.updateCommitStatus(ctx, status, command.Apply)
	})
	if err != nil {
		if p.obsoletePlan(ctx, err) {
			return status, false
		}
		ctx.CommandHasErrors = true
		if !reported {
			if reportErr := p.pullUpdater.updatePull(ctx, cmd, command.Result{Error: fmt.Errorf("finishing plan publication: %w", err)}); reportErr != nil {
				ctx.Log.Err("reporting command result: %s", reportErr)
			}
		}
		return status, false
	}
	return status, true
}

func (p *PlanCommandRunner) refreshPublicationHead(ctx *command.Context) error {
	if p.LivePullHeadFetcher == nil || ctx.Pull.Num <= 0 {
		return nil
	}
	project := command.ProjectContext{Log: ctx.Log, Pull: ctx.Pull, PullStatus: ctx.PullStatus, API: ctx.API}
	live, err := p.LivePullHeadFetcher.GetLivePullIdentity(project)
	if err != nil {
		return fmt.Errorf("refreshing pull identity before publication: %w", err)
	}
	if live.HeadCommit == "" {
		return fmt.Errorf("live pull request head is empty")
	}
	if err := validateCommandStartIdentity(project, live); err != nil {
		return fmt.Errorf("%w: %w", db.ErrPlanGenerationSuperseded, err)
	}
	return nil
}

// Empty plans still need one atomic cleanup/publication section. A targeted
// empty plan retains previous durable state, matching main's existing UX.
func (p *PlanCommandRunner) handleNoProjectPlan(ctx *command.Context, cmd PullCommand) {
	targeted := false
	if comment, ok := cmd.(*CommentCommand); ok {
		targeted = comment.IsForSpecificProject()
	}
	err := p.Publication.RunCommand(ctx, func() error {
		if err := p.refreshPublicationHead(ctx); err != nil {
			return err
		}
		var status models.PullStatus
		if !targeted {
			var err error
			status, err = p.clearPlansAndPullStatusForNoProjects(ctx, ctx.Pull)
			if err != nil {
				return err
			}
		}
		publishEmpty := func(names ...command.Name) error {
			for _, name := range names {
				if err := publishTerminal(ctx, func() error {
					return p.commitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.SuccessCommitStatus, name, models.ProjectCounts{})
				}); err != nil {
					return err
				}
			}
			return nil
		}
		if cmd.IsAutoplan() {
			if p.silenceVCSStatusNoPlans || p.silenceVCSStatusNoProjects {
				return nil
			}
			return publishEmpty(command.Plan, command.PolicyCheck, command.Apply)
		}
		if p.SilenceNoProjects {
			if p.silenceVCSStatusNoProjects {
				return nil
			}
			if !targeted {
				return publishEmpty(command.Plan, command.PolicyCheck, command.Apply)
			}
			previous, err := p.pullStatusFetcher.GetPullStatus(ctx.Pull)
			if err != nil {
				return err
			}
			if previous == nil {
				return publishEmpty(command.Plan)
			}
			return p.updateCommitStatus(ctx, *previous, command.Plan)
		}
		if targeted {
			var err error
			status, err = p.dbUpdater.updateDB(ctx, ctx.Pull, nil)
			if err != nil {
				return err
			}
		}
		if err := p.pullUpdater.updatePull(ctx, cmd, command.Result{}); err != nil {
			return err
		}
		if err := p.updateCommitStatus(ctx, status, command.Plan); err != nil {
			return err
		}
		if err := p.updateCommitStatus(ctx, status, command.Apply); err != nil {
			return err
		}
		if !targeted {
			return publishEmpty(command.PolicyCheck)
		}
		return nil
	})
	if err != nil {
		if p.obsoletePlan(ctx, err) {
			return
		}
		p.handleNoProjectPlanStateError(ctx, cmd, err)
	}
}
