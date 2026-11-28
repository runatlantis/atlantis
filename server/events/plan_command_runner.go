// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/runatlantis/atlantis/server/core/config/valid"
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
	// LivePullHeadFetcher fetches the pull request's current head commit. An
	// autoplan that waited behind a plan of an older commit uses it to plan
	// the live head.
	LivePullHeadFetcher LivePullHeadFetcher
	// Drainer reports whether Atlantis is shutting down, so that an autoplan
	// that waited doesn't start a new plan during shutdown. Optional.
	Drainer *Drainer
}

func (p *PlanCommandRunner) runAutoplan(ctx *command.Context, isCancelled func() bool) {
	baseRepo := ctx.Pull.BaseRepo
	pull := ctx.Pull
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
		if statusErr := p.commitStatusUpdater.UpdateCombined(ctx.Log, baseRepo, pull, models.FailedCommitStatus, command.Plan); statusErr != nil {
			ctx.Log.Warn("unable to update commit status: %s", statusErr)
		}
		p.pullUpdater.updatePull(ctx, AutoplanCommand{}, command.Result{Error: err})
		return
	}

	projectCmds, policyCheckCmds := p.partitionProjectCmds(ctx, projectCmds)

	if len(projectCmds) == 0 {
		ctx.Log.Info("determined there was no project to run plan in")
		if _, err := p.clearPlansAndPullStatusForNoProjects(ctx, pull); err != nil {
			p.handleNoProjectPlanStateError(ctx, AutoplanCommand{}, err)
			return
		}
		if !p.silenceVCSStatusNoPlans && !p.silenceVCSStatusNoProjects {
			// If there were no projects modified, we set successful commit statuses
			// with 0/0 projects planned/policy_checked/applied successfully because some users require
			// the Atlantis status to be passing for all pull requests.
			ctx.Log.Debug("setting VCS status to success with no projects found")
			if err := p.commitStatusUpdater.UpdateCombinedCount(ctx.Log, baseRepo, pull, models.SuccessCommitStatus, command.Plan, models.ProjectCounts{}); err != nil {
				ctx.Log.Warn("unable to update commit status: %s", err)
			}
			if err := p.commitStatusUpdater.UpdateCombinedCount(ctx.Log, baseRepo, pull, models.SuccessCommitStatus, command.PolicyCheck, models.ProjectCounts{}); err != nil {
				ctx.Log.Warn("unable to update commit status: %s", err)
			}
			if err := p.commitStatusUpdater.UpdateCombinedCount(ctx.Log, baseRepo, pull, models.SuccessCommitStatus, command.Apply, models.ProjectCounts{}); err != nil {
				ctx.Log.Warn("unable to update commit status: %s", err)
			}
		} else {
			// When silence is enabled and no projects are found, don't set any status
			ctx.Log.Debug("silence enabled and no projects found - not setting any VCS status")
		}
		return
	}
	p.updatePendingCommitStatus(ctx, command.Plan)

	// discard previous plans that might not be relevant anymore
	ctx.Log.Debug("deleting previous plans and locks")
	if err := p.deletePlansAndPlanLocks(ctx, projectCmds); err != nil {
		if statusErr := p.commitStatusUpdater.UpdateCombined(ctx.Log, baseRepo, pull, models.FailedCommitStatus, command.Plan); statusErr != nil {
			ctx.Log.Warn("unable to update commit status: %s", statusErr)
		}
		p.pullUpdater.updatePull(ctx, AutoplanCommand{}, command.Result{Error: err})
		return
	}

	result := runProjectCmdsWithCancellation(ctx, projectCmds, isCancelled, p.parallelPoolSize, p.isParallelEnabled(projectCmds), p.prjCmdRunner.Plan)

	if p.autoMerger.automergeEnabled(projectCmds) && result.HasErrors() {
		ctx.Log.Info("deleting plans because there were errors and automerge requires all plans succeed")
		if err := p.deletePlansAndPlanLocks(ctx, projectCmds); err != nil {
			ctx.Log.Err("deleting pending plans: %s", err)
		}
		result.PlansDeleted = true
	}

	pullStatus, err := p.dbUpdater.updateDB(ctx, ctx.Pull, result.ProjectResults)
	if err != nil {
		p.planPersistenceFailed(ctx, AutoplanCommand{}, projectCmds, result, err)
		return
	}

	p.publishPlanStatuses(projectCmds, result, models.SuccessCommitStatus)
	p.pullUpdater.updatePull(ctx, AutoplanCommand{}, result)
	p.updateCommitStatus(ctx, pullStatus, command.Plan)
	p.updateCommitStatus(ctx, pullStatus, command.Apply)

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

		p.policyCheckCommandRunner.Run(ctx, policyCheckCmds)
	}
}

func (p *PlanCommandRunner) run(ctx *command.Context, cmd *CommentCommand, isCancelled func() bool) {
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
		if statusErr := p.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, command.Plan); statusErr != nil {
			ctx.Log.Warn("unable to update commit status: %s", statusErr)
		}
		p.pullUpdater.updatePull(ctx, cmd, command.Result{Error: err})
		return
	}

	var noProjectPullStatus *models.PullStatus
	if len(projectCmds) == 0 && !cmd.IsForSpecificProject() {
		ctx.Log.Info("determined there was no project to run plan in")
		pullStatus, err := p.clearPlansAndPullStatusForNoProjects(ctx, pull)
		if err != nil {
			p.handleNoProjectPlanStateError(ctx, cmd, err)
			return
		}
		noProjectPullStatus = &pullStatus
	}
	if len(projectCmds) == 0 && p.SilenceNoProjects {
		if cmd.IsForSpecificProject() {
			ctx.Log.Info("determined there was no project to run plan in")
		}
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
					if err := p.commitStatusUpdater.UpdateCombinedCount(ctx.Log, baseRepo, pull, models.SuccessCommitStatus, command.Plan, models.ProjectCounts{}); err != nil {
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
				if err := p.commitStatusUpdater.UpdateCombinedCount(ctx.Log, baseRepo, pull, models.SuccessCommitStatus, command.Plan, models.ProjectCounts{}); err != nil {
					ctx.Log.Warn("unable to update commit status: %s", err)
				}
				if err := p.commitStatusUpdater.UpdateCombinedCount(ctx.Log, baseRepo, pull, models.SuccessCommitStatus, command.PolicyCheck, models.ProjectCounts{}); err != nil {
					ctx.Log.Warn("unable to update commit status: %s", err)
				}
				if err := p.commitStatusUpdater.UpdateCombinedCount(ctx.Log, baseRepo, pull, models.SuccessCommitStatus, command.Apply, models.ProjectCounts{}); err != nil {
					ctx.Log.Warn("unable to update commit status: %s", err)
				}
			}
		} else {
			// When silence is enabled and no projects are found, don't set any status
			ctx.Log.Debug("silence enabled and no projects found - not setting any VCS status")
		}
		return
	}
	projectCmds, policyCheckCmds := p.partitionProjectCmds(ctx, projectCmds)
	if len(projectCmds) > 0 {
		p.updatePendingCommitStatus(ctx, command.Plan)
	}

	// if the plan is generic, new plans will be generated based on changes
	// discard previous plans that might not be relevant anymore
	if !cmd.IsForSpecificProject() && len(projectCmds) > 0 {
		ctx.Log.Debug("deleting previous plans and locks")
		if err := p.deletePlansAndPlanLocks(ctx, projectCmds); err != nil {
			if statusErr := p.commitStatusUpdater.UpdateCombined(ctx.Log, baseRepo, pull, models.FailedCommitStatus, command.Plan); statusErr != nil {
				ctx.Log.Warn("unable to update commit status: %s", statusErr)
			}
			p.pullUpdater.updatePull(ctx, cmd, command.Result{Error: err})
			return
		}
	}

	result := runProjectCmdsWithCancellation(ctx, projectCmds, isCancelled, p.parallelPoolSize, p.isParallelEnabled(projectCmds), p.prjCmdRunner.Plan)
	ctx.CommandHasErrors = result.HasErrors()

	if p.autoMerger.automergeEnabled(projectCmds) && result.HasErrors() {
		ctx.Log.Info("deleting plans because there were errors and automerge requires all plans succeed")
		if err := p.deletePlansAndPlanLocks(ctx, projectCmds); err != nil {
			ctx.Log.Err("deleting pending plans: %s", err)
		}
		result.PlansDeleted = true
	}

	var pullStatus models.PullStatus
	if noProjectPullStatus != nil {
		pullStatus = *noProjectPullStatus
	} else if len(projectCmds) == 0 && !cmd.IsForSpecificProject() {
		pullStatus, err = p.dbUpdater.replaceDB(ctx, pull, result.ProjectResults)
	} else {
		pullStatus, err = p.dbUpdater.updateDB(ctx, pull, result.ProjectResults)
	}
	if err != nil {
		p.planPersistenceFailed(ctx, cmd, projectCmds, result, err)
		return
	}

	p.publishPlanStatuses(projectCmds, result, models.SuccessCommitStatus)
	p.pullUpdater.updatePull(ctx, cmd, result)
	p.updateCommitStatus(ctx, pullStatus, command.Plan)
	p.updateCommitStatus(ctx, pullStatus, command.Apply)

	// Runs policy checks step after all plans are successful.
	// This step does not approve any policies that require approval.
	if len(result.ProjectResults) > 0 &&
		(!result.HasErrors() && !result.PlansDeleted) {
		ctx.Log.Info("Running policy check for '%s'", cmd.CommandName())
		p.policyCheckCommandRunner.Run(ctx, policyCheckCmds)
	} else if len(projectCmds) == 0 && !cmd.IsForSpecificProject() {
		// If there were no projects modified, we set successful commit statuses
		// with 0/0 projects planned/policy_checked/applied successfully because some users require
		// the Atlantis status to be passing for all pull requests.
		ctx.Log.Debug("setting VCS status to success with no projects found")
		if err := p.commitStatusUpdater.UpdateCombinedCount(ctx.Log, baseRepo, pull, models.SuccessCommitStatus, command.PolicyCheck, models.ProjectCounts{}); err != nil {
			ctx.Log.Warn("unable to update commit status: %s", err)
		}
	}
}

func (p *PlanCommandRunner) Run(ctx *command.Context, cmd *CommentCommand) {
	if ctx.Trigger == command.AutoTrigger {
		p.RunAutoplanWithHooks(ctx, nil, nil)
	} else {
		p.RunCommentPlanWithHooks(ctx, cmd, nil, nil)
	}
}

// RunCommentPlanWithHooks holds the pull lock while the plan and both hooks use the workspace.
func (p *PlanCommandRunner) RunCommentPlanWithHooks(ctx *command.Context, cmd *CommentCommand, before func() bool, after func()) {
	unlock, ok := p.lockPullForPlan(ctx, cmd)
	if !ok {
		return
	}
	defer unlock()
	isCancelled, finish := trackProjectCmds(p.cancellationTracker, ctx.Pull)
	defer finish()

	if before != nil && !before() {
		return
	}
	p.run(ctx, cmd, isCancelled)
	if after != nil {
		after()
	}
}

// RunAutoplanWithHooks takes the pull-level plan lock, then runs before
// (pre-workflow hooks), the autoplan, and after (post-workflow hooks) while
// holding it. If the autoplan waited and was superseded or cancelled, neither
// hook runs.
func (p *PlanCommandRunner) RunAutoplanWithHooks(ctx *command.Context, before func() bool, after func()) {
	unlock, ok := p.lockPullForPlan(ctx, AutoplanCommand{})
	if !ok {
		return
	}
	defer unlock()
	isCancelled, finish := trackProjectCmds(p.cancellationTracker, ctx.Pull)
	defer finish()
	if before != nil && !before() {
		return
	}
	p.runAutoplan(ctx, isCancelled)
	if after != nil {
		after()
	}
}

func (p *PlanCommandRunner) clearPlansAndPullStatusForNoProjects(ctx *command.Context, pull models.PullRequest) (models.PullStatus, error) {
	if _, err := p.deletePlansAndPendingPlanLocks(ctx); err != nil {
		return models.PullStatus{}, err
	}
	pullStatus, err := p.dbUpdater.replaceDB(ctx, pull, nil)
	if err != nil {
		return models.PullStatus{}, fmt.Errorf("writing empty plan status: %w", err)
	}
	return pullStatus, nil
}

func (p *PlanCommandRunner) lockPullForPlan(ctx *command.Context, cmd PullCommand) (func(), bool) {
	if p.workingDirLocker == nil {
		return func() {}, true
	}
	repoFullName, pullNum, metadata := ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, WorkingDirLockMetadataForPull(ctx.Pull)
	if ctx.Trigger != command.AutoTrigger {
		unlockFn, err := p.workingDirLocker.TryLockPull(repoFullName, pullNum, command.Plan, metadata)
		if err != nil {
			p.handleNoProjectPlanStateError(ctx, cmd, err)
			return nil, false
		}
		return unlockFn, true
	}

	return p.lockPullForAutoplan(ctx, cmd)
}

// lockPullForAutoplan takes the pull-level plan lock for an autoplan. Autoplan
// runs because a new commit was pushed, so if this pull request is still
// planning an older commit it waits for that plan to finish instead of
// failing the new head (#7050).
//
// Waiters are ordered by arrival, not by commit, and webhooks can arrive out
// of order. After waiting, the pull request's live head is fetched so the plan
// runs for the commit that is HEAD now, not the one the event carried.
func (p *PlanCommandRunner) lockPullForAutoplan(ctx *command.Context, cmd PullCommand) (func(), bool) {
	repoFullName, pullNum := ctx.Pull.BaseRepo.FullName, ctx.Pull.Num
	metadata := WorkingDirLockMetadataForPull(ctx.Pull)
	waited := false
	lock, err := p.workingDirLocker.LockPullAfterStalePlan(repoFullName, pullNum, command.Plan, metadata, func(holder WorkingDirLockMetadata) {
		waited = true
		ctx.Log.Info("pull request is still planning commit %s; autoplan for commit %s will start when it finishes", shortSHA(holder.HeadCommit), shortSHA(metadata.HeadCommit))
	})
	switch {
	case errors.Is(err, ErrPullLockSuperseded):
		ctx.Log.Info("skipping autoplan for commit %s because a newer autoplan for this pull request is waiting or already running", shortSHA(metadata.HeadCommit))
		return nil, false
	case errors.Is(err, ErrPullLockCancelled):
		ctx.Log.Info("skipping autoplan for commit %s because the pull request was closed or its commands were cancelled while it waited", shortSHA(metadata.HeadCommit))
		return nil, false
	case err != nil:
		p.handleNoProjectPlanStateError(ctx, cmd, err)
		return nil, false
	}
	// Release the lock on every path that doesn't hand it to the caller,
	// including a panic while fetching the live head.
	handedOff := false
	defer func() {
		if !handedOff {
			lock.Unlock()
		}
	}()
	if waited {
		if p.Drainer != nil && p.Drainer.GetStatus().ShuttingDown {
			return nil, false
		}
		livePull, err := p.livePullIdentity(ctx)
		if err != nil {
			if p.Drainer != nil && p.Drainer.GetStatus().ShuttingDown {
				return nil, false
			}
			if !lock.Start() {
				return nil, false
			}
			p.handleNoProjectPlanStateError(ctx, cmd, fmt.Errorf("fetching live pull request: %w", err))
			return nil, false
		}
		if livePull.State == models.ClosedPullState {
			return nil, false
		}
		ctx.Pull.HeadCommit = livePull.HeadCommit
		if livePull.BaseBranch != "" {
			ctx.Pull.BaseBranch = livePull.BaseBranch
		}
		if !lock.UpdateMetadata(WorkingDirLockMetadataForPull(ctx.Pull)) {
			return nil, false
		}
		if p.pullStatusFetcher != nil {
			pullStatus, err := p.pullStatusFetcher.GetPullStatus(ctx.Pull)
			if err != nil {
				if p.Drainer != nil && p.Drainer.GetStatus().ShuttingDown {
					return nil, false
				}
				if !lock.Start() {
					return nil, false
				}
				p.handleNoProjectPlanStateError(ctx, cmd, fmt.Errorf("fetching current pull status: %w", err))
				return nil, false
			}
			ctx.PullStatus = pullStatus
		}
		if p.Drainer != nil && p.Drainer.GetStatus().ShuttingDown {
			return nil, false
		}
	}
	if !lock.Start() {
		return nil, false
	}
	handedOff = true
	return lock.Unlock, true
}

func (p *PlanCommandRunner) livePullIdentity(ctx *command.Context) (models.PullRequest, error) {
	if p.LivePullHeadFetcher == nil {
		return models.PullRequest{}, errors.New("live pull request fetcher is not configured")
	}
	livePull, err := p.LivePullHeadFetcher.GetLivePullIdentity(command.ProjectContext{Log: ctx.Log, Pull: ctx.Pull, PullStatus: ctx.PullStatus})
	if err != nil {
		return models.PullRequest{}, err
	}
	if livePull.State != models.ClosedPullState && livePull.HeadCommit == "" {
		return models.PullRequest{}, errors.New("live pull request head is empty")
	}
	return livePull, nil
}

func (p *PlanCommandRunner) handleNoProjectPlanStateError(ctx *command.Context, cmd PullCommand, err error) {
	ctx.CommandHasErrors = true
	if statusErr := p.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, command.Plan); statusErr != nil {
		ctx.Log.Warn("unable to update commit status: %s", statusErr)
	}
	p.pullUpdater.updatePull(ctx, cmd, command.Result{Error: err})
}

func (p *PlanCommandRunner) ShouldSkipPreWorkflowHooks(ctx *command.Context, cmd *CommentCommand) bool {
	return MarkCommandSkippedIfIgnoredTarget(ctx, command.Plan, cmd, p.prjCmdBuilder)
}

func (p *PlanCommandRunner) updatePendingCommitStatus(ctx *command.Context, commandName command.Name) {
	if p.silenceVCSStatusNoProjects {
		ctx.Log.Debug("silence enabled - not setting pending VCS status")
		return
	}
	if err := p.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.PendingCommitStatus, commandName); err != nil {
		ctx.Log.Warn("unable to update commit status: %s", err)
	}
}

func (p *PlanCommandRunner) updateCommitStatus(ctx *command.Context, pullStatus models.PullStatus, commandName command.Name) {
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
				return
			}
		}
	}

	if err := p.commitStatusUpdater.UpdateCombinedCount(
		ctx.Log,
		ctx.Pull.BaseRepo,
		ctx.Pull,
		status,
		commandName,
		models.ProjectCounts{Success: numSuccess, Total: len(pullStatus.Projects), Errored: numErrored, NoChanges: numNoChanges},
	); err != nil {
		ctx.Log.Warn("unable to update commit status: %s", err)
	}
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
func (p *PlanCommandRunner) publishPlanStatuses(projectCmds []command.ProjectContext, result command.Result, status models.CommitStatus) {
	if publisher, ok := p.prjCmdRunner.(DeferredPlanStatusPublisher); ok {
		publisher.PublishDeferredPlanStatuses(projectCmds, result, status)
	}
}

func (p *PlanCommandRunner) planPersistenceFailed(ctx *command.Context, cmd PullCommand, projectCmds []command.ProjectContext, result command.Result, err error) {
	ctx.CommandHasErrors = true
	result.Error = fmt.Errorf("persisting plan results: %w; restore database connectivity and run `atlantis plan` again before applying", err)
	p.publishPlanStatuses(projectCmds, result, models.FailedCommitStatus)
	for _, name := range []command.Name{command.Plan, command.Apply} {
		if statusErr := p.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, name); statusErr != nil {
			ctx.Log.Warn("unable to update commit status: %s", statusErr)
		}
	}
	p.pullUpdater.updatePull(ctx, cmd, result)
}
