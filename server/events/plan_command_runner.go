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
}

func (p *PlanCommandRunner) runAutoplan(ctx *command.Context) {
	baseRepo := ctx.Pull.BaseRepo
	pull := ctx.Pull
	publicationToken, err := p.acquirePlanPublicationClaim(ctx)
	if err != nil {
		return
	}
	unlockPullPlan, lockErr := p.lockPullForPlan(ctx)
	if lockErr != nil {
		p.finishPlanPublicationClaim(ctx, publicationToken, p.handleNoProjectPlanStateError(ctx, AutoplanCommand{}, lockErr))
		return
	}
	defer unlockPullPlan()

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
		p.finishPlanPublicationClaim(ctx, publicationToken, errors.Join(
			p.commitStatusUpdater.UpdateCombined(ctx.Log, baseRepo, pull, models.FailedCommitStatus, command.Plan),
			p.pullUpdater.updatePull(ctx, AutoplanCommand{}, command.Result{Error: err}),
		))
		return
	}

	projectCmds, policyCheckCmds := p.partitionProjectCmds(ctx, projectCmds)

	if len(projectCmds) == 0 {
		ctx.Log.Info("determined there was no project to run plan in")
		beginResult, err := p.clearPlansAndPullStatusForNoProjects(ctx, pull, publicationToken)
		if err != nil {
			p.finishPlanPublicationClaim(ctx, publicationToken, p.handleNoProjectPlanStateError(ctx, AutoplanCommand{}, err))
			return
		}
		ctx.PullStatus = &beginResult.PullStatus
		var publicationErrors []error
		if !p.silenceVCSStatusNoPlans && !p.silenceVCSStatusNoProjects {
			publicationErrors = append(publicationErrors,
				p.publishCancelledPlanStatuses(cancelledPlanProjectContexts(ctx, beginResult.Canceled)),
				p.updateNoProjectCommitStatuses(ctx, beginResult.PullStatus),
			)
		} else {
			// When silence is enabled and no projects are found, don't set any status
			ctx.Log.Debug("silence enabled and no projects found - not setting any VCS status")
		}
		p.finishPlanPublicationClaim(ctx, publicationToken, errors.Join(publicationErrors...))
		return
	}
	planGeneration, beginResult, err := p.dbUpdater.beginPlanGeneration(pull, projectCmds, true, publicationToken)
	if err != nil {
		publicationErr := errors.Join(
			p.publishPlanGenerationStartFailureStatuses(projectCmds, err),
			p.handlePlanGenerationStartError(ctx, AutoplanCommand{}, err),
		)
		p.finishPlanPublicationClaim(ctx, publicationToken, publicationErr)
		return
	}
	publicationErr := errors.Join(
		p.publishPendingPlanStatuses(projectCmds),
		p.publishCancelledPlanStatuses(cancelledPlanProjectContexts(ctx, beginResult.Canceled)),
		p.updatePendingCommitStatus(ctx, command.Plan),
	)
	if !p.finishPlanPublicationClaim(ctx, publicationToken, publicationErr) {
		return
	}
	prepareProjectCommandsForPlanGeneration(projectCmds, planGeneration)
	prepareProjectCommandsForPlanGeneration(policyCheckCmds, planGeneration)

	// Do not delete deterministic plan-store artifacts or same-pull plan locks
	// after BeginPlanGeneration. A newer generation on another replica may
	// already own them; durable status and the accepted artifact hash invalidate
	// the previous generation without destructive cleanup.
	result := runProjectCmdsWithCancellationTracker(ctx, projectCmds, p.cancellationTracker, p.parallelPoolSize, p.isParallelEnabled(projectCmds), p.prjCmdRunner.Plan)
	result = completePlanGenerationResults(projectCmds, result)
	result = invalidateAutomergePlanResults(result, p.autoMerger.automergeEnabled(projectCmds))

	publicationToken, err = p.waitForPlanPublicationClaim(ctx)
	if err != nil {
		p.completeUnpublishedPlanJobs(projectCmds, result, "Atlantis could not claim the durable publication boundary for this completed plan. Run `atlantis plan` again.")
		return
	}
	pullStatus, err := p.dbUpdater.completePlanGeneration(ctx.Pull, planGeneration, result.ProjectResults, publicationToken)
	if err != nil {
		if db.IsPlanGenerationObsolete(err) {
			p.finishPlanPublicationClaim(ctx, publicationToken, nil)
			p.handleSupersededPlanGeneration(ctx, planGeneration, projectCmds, result, err)
			return
		}
		publicationErr = errors.Join(
			p.publishDeferredPlanStatuses(projectCmds, result, models.FailedCommitStatus),
			p.handlePlanGenerationCompletionError(ctx, AutoplanCommand{}, err),
		)
		p.finishPlanPublicationClaim(ctx, publicationToken, publicationErr)
		return
	}
	ctx.PullStatus = &pullStatus
	publicationErr = errors.Join(
		p.publishDeferredPlanStatuses(projectCmds, result, models.SuccessCommitStatus),
		p.pullUpdater.updatePull(ctx, AutoplanCommand{}, result),
		p.updateCommitStatus(ctx, pullStatus, command.Plan),
		p.updateCommitStatus(ctx, pullStatus, command.Apply),
	)
	if !p.finishPlanPublicationClaim(ctx, publicationToken, publicationErr) {
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

		p.policyCheckCommandRunner.Run(ctx, policyCheckCmds)
	}
}

func (p *PlanCommandRunner) run(ctx *command.Context, cmd *CommentCommand) {
	var err error
	baseRepo := ctx.Pull.BaseRepo
	pull := ctx.Pull
	publicationToken, err := p.acquirePlanPublicationClaim(ctx)
	if err != nil {
		return
	}
	unlockPullPlan, lockErr := p.lockPullForPlan(ctx)
	if lockErr != nil {
		p.finishPlanPublicationClaim(ctx, publicationToken, p.handleNoProjectPlanStateError(ctx, cmd, lockErr))
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
		p.finishPlanPublicationClaim(ctx, publicationToken, nil)
		return
	}

	if p.DiscardApprovalOnPlan {
		if discardErr := p.pullUpdater.VCSClient.DiscardReviews(ctx.Log, baseRepo, pull); discardErr != nil {
			ctx.Log.Err("failed to remove approvals: %s", discardErr)
			ctx.CommandHasErrors = true
			p.finishPlanPublicationClaim(ctx, publicationToken, fmt.Errorf("discarding pull request approvals: %w", discardErr))
			return
		}
	}

	if err != nil {
		p.finishPlanPublicationClaim(ctx, publicationToken, errors.Join(
			p.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, command.Plan),
			p.pullUpdater.updatePull(ctx, cmd, command.Result{Error: err}),
		))
		return
	}

	var noProjectBeginResult *db.PlanGenerationBeginResult
	if len(projectCmds) == 0 && !cmd.IsForSpecificProject() {
		ctx.Log.Info("determined there was no project to run plan in")
		beginResult, err := p.clearPlansAndPullStatusForNoProjects(ctx, pull, publicationToken)
		if err != nil {
			p.finishPlanPublicationClaim(ctx, publicationToken, p.handleNoProjectPlanStateError(ctx, cmd, err))
			return
		}
		ctx.PullStatus = &beginResult.PullStatus
		noProjectBeginResult = &beginResult
	}
	if len(projectCmds) == 0 && p.SilenceNoProjects {
		var publicationErrors []error
		if cmd.IsForSpecificProject() {
			ctx.Log.Info("determined there was no project to run plan in")
		}
		if !p.silenceVCSStatusNoProjects {
			if cmd.IsForSpecificProject() {
				// With a specific plan, just reset the status so it's not stuck in pending state
				pullStatus, err := p.pullStatusFetcher.GetPullStatus(pull)
				if err != nil {
					ctx.Log.Warn("unable to fetch pull status: %s", err)
					p.finishPlanPublicationClaim(ctx, publicationToken, err)
					return
				}
				if pullStatus == nil {
					// default to 0/0
					ctx.Log.Debug("setting VCS status to 0/0 success as no previous state was found")
					if err := p.commitStatusUpdater.UpdateCombinedCount(ctx.Log, baseRepo, pull, models.SuccessCommitStatus, command.Plan, models.ProjectCounts{}); err != nil {
						ctx.Log.Warn("unable to update commit status: %s", err)
						publicationErrors = append(publicationErrors, err)
					}
					p.finishPlanPublicationClaim(ctx, publicationToken, errors.Join(publicationErrors...))
					return
				}
				ctx.Log.Debug("resetting VCS status")
				publicationErrors = append(publicationErrors, p.updateCommitStatus(ctx, *pullStatus, command.Plan))
			} else {
				publicationErrors = append(publicationErrors,
					p.publishCancelledPlanStatuses(cancelledPlanProjectContexts(ctx, noProjectBeginResult.Canceled)),
					p.updateNoProjectCommitStatuses(ctx, noProjectBeginResult.PullStatus),
				)
			}
		} else {
			// When silence is enabled and no projects are found, don't set any status
			ctx.Log.Debug("silence enabled and no projects found - not setting any VCS status")
		}
		p.finishPlanPublicationClaim(ctx, publicationToken, errors.Join(publicationErrors...))
		return
	}
	projectCmds, policyCheckCmds := p.partitionProjectCmds(ctx, projectCmds)
	var planGeneration string
	if len(projectCmds) > 0 {
		var beginResult db.PlanGenerationBeginResult
		planGeneration, beginResult, err = p.dbUpdater.beginPlanGeneration(pull, projectCmds, !cmd.IsForSpecificProject(), publicationToken)
		if err != nil {
			publicationErr := errors.Join(
				p.publishPlanGenerationStartFailureStatuses(projectCmds, err),
				p.handlePlanGenerationStartError(ctx, cmd, err),
			)
			p.finishPlanPublicationClaim(ctx, publicationToken, publicationErr)
			return
		}
		publicationErr := errors.Join(
			p.publishPendingPlanStatuses(projectCmds),
			p.publishCancelledPlanStatuses(cancelledPlanProjectContexts(ctx, beginResult.Canceled)),
			p.updatePendingCommitStatus(ctx, command.Plan),
		)
		if !p.finishPlanPublicationClaim(ctx, publicationToken, publicationErr) {
			return
		}
		publicationToken = ""
		prepareProjectCommandsForPlanGeneration(projectCmds, planGeneration)
		prepareProjectCommandsForPlanGeneration(policyCheckCmds, planGeneration)
	}

	// Cleanup after generation begin is intentionally non-destructive. Another
	// replica can supersede this command while sharing its plan-store key and
	// same-pull project lock.
	result := runProjectCmdsWithCancellationTracker(ctx, projectCmds, p.cancellationTracker, p.parallelPoolSize, p.isParallelEnabled(projectCmds), p.prjCmdRunner.Plan)
	result = completePlanGenerationResults(projectCmds, result)
	result = invalidateAutomergePlanResults(result, p.autoMerger.automergeEnabled(projectCmds))
	ctx.CommandHasErrors = result.HasErrors()

	var pullStatus models.PullStatus
	if noProjectBeginResult != nil {
		pullStatus = noProjectBeginResult.PullStatus
	} else if len(projectCmds) == 0 && !cmd.IsForSpecificProject() {
		pullStatus, err = p.dbUpdater.replaceDB(ctx, pull, result.ProjectResults)
	} else if len(projectCmds) == 0 {
		pullStatus, err = p.dbUpdater.updateDB(ctx, pull, result.ProjectResults)
	} else {
		publicationToken, err = p.waitForPlanPublicationClaim(ctx)
		if err != nil {
			p.completeUnpublishedPlanJobs(projectCmds, result, "Atlantis could not claim the durable publication boundary for this completed plan. Run `atlantis plan` again.")
			return
		}
		pullStatus, err = p.dbUpdater.completePlanGeneration(pull, planGeneration, result.ProjectResults, publicationToken)
	}
	if err != nil {
		if db.IsPlanGenerationObsolete(err) {
			if publicationToken != "" {
				p.finishPlanPublicationClaim(ctx, publicationToken, nil)
			}
			p.handleSupersededPlanGeneration(ctx, planGeneration, projectCmds, result, err)
			return
		}
		publicationErr := errors.Join(
			p.publishDeferredPlanStatuses(projectCmds, result, models.FailedCommitStatus),
			p.handlePlanGenerationCompletionError(ctx, cmd, err),
		)
		if publicationToken != "" {
			p.finishPlanPublicationClaim(ctx, publicationToken, publicationErr)
		}
		return
	}
	ctx.PullStatus = &pullStatus
	commitStatusErr := errors.Join(
		p.updateCommitStatus(ctx, pullStatus, command.Plan),
		p.updateCommitStatus(ctx, pullStatus, command.Apply),
	)
	if len(projectCmds) == 0 && !cmd.IsForSpecificProject() {
		commitStatusErr = errors.Join(
			p.publishCancelledPlanStatuses(cancelledPlanProjectContexts(ctx, noProjectBeginResult.Canceled)),
			p.updateNoProjectCommitStatuses(ctx, pullStatus),
		)
	}
	publicationErr := errors.Join(
		p.publishDeferredPlanStatuses(projectCmds, result, models.SuccessCommitStatus),
		p.pullUpdater.updatePull(ctx, cmd, result),
		commitStatusErr,
	)
	if publicationToken != "" && !p.finishPlanPublicationClaim(ctx, publicationToken, publicationErr) {
		return
	}

	// Runs policy checks step after all plans are successful.
	// This step does not approve any policies that require approval.
	if len(result.ProjectResults) > 0 &&
		len(policyCheckCmds) > 0 && (!result.HasErrors() && !result.PlansDeleted) {
		ctx.Log.Info("Running policy check for '%s'", cmd.CommandName())
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

func (p *PlanCommandRunner) clearPlansAndPullStatusForNoProjects(ctx *command.Context, pull models.PullRequest, claimToken string) (db.PlanGenerationBeginResult, error) {
	_, beginResult, err := p.dbUpdater.beginPlanGeneration(pull, nil, true, claimToken)
	if err != nil {
		return db.PlanGenerationBeginResult{}, fmt.Errorf("writing empty plan status: %w", err)
	}
	// Retain discarded project tombstones along with any filesystem artifacts.
	// Another replica can begin a generation immediately after this durable
	// replacement, and this process has no atomic way to prove a discovered plan
	// still belongs to the replaced state. Generic apply deliberately ignores
	// artifacts whose matching durable project status is already non-applyable.
	return beginResult, nil
}

func (p *PlanCommandRunner) lockPullForPlan(ctx *command.Context) (func(), error) {
	if p.workingDirLocker == nil {
		return func() {}, nil
	}
	unlockFn, err := p.workingDirLocker.TryLockPull(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, command.Plan, WorkingDirLockMetadataForPull(ctx.Pull))
	if err != nil {
		return nil, err
	}
	return unlockFn, nil
}

func (p *PlanCommandRunner) handleNoProjectPlanStateError(ctx *command.Context, cmd PullCommand, err error) error {
	ctx.CommandHasErrors = true
	var publicationErrors []error
	if statusErr := p.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, command.Plan); statusErr != nil {
		ctx.Log.Warn("unable to update commit status: %s", statusErr)
		publicationErrors = append(publicationErrors, statusErr)
	}
	publicationErrors = append(publicationErrors, p.pullUpdater.updatePull(ctx, cmd, command.Result{Error: err}))
	return errors.Join(publicationErrors...)
}

func (p *PlanCommandRunner) handlePlanResultPersistenceError(ctx *command.Context, cmd PullCommand, err error) error {
	persistenceErr := fmt.Errorf("persisting plan results: %w", err)
	ctx.CommandHasErrors = true
	var publicationErrors []error
	if statusErr := p.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, command.Plan); statusErr != nil {
		ctx.Log.Warn("unable to update commit status: %s", statusErr)
		publicationErrors = append(publicationErrors, statusErr)
	}
	if statusErr := p.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, command.Apply); statusErr != nil {
		ctx.Log.Warn("unable to update commit status: %s", statusErr)
		publicationErrors = append(publicationErrors, statusErr)
	}
	publicationErrors = append(publicationErrors, p.pullUpdater.updatePull(ctx, cmd, command.Result{Error: persistenceErr}))
	return errors.Join(publicationErrors...)
}

func (p *PlanCommandRunner) handlePlanGenerationStartError(ctx *command.Context, cmd PullCommand, err error) error {
	return p.handlePlanResultPersistenceError(ctx, cmd, fmt.Errorf("invalidating previous plan state before planning: %w", err))
}

func (p *PlanCommandRunner) handlePlanGenerationCompletionError(
	ctx *command.Context,
	cmd PullCommand,
	err error,
) error {
	return p.handlePlanResultPersistenceError(ctx, cmd, err)
}

func (p *PlanCommandRunner) acquirePlanPublicationClaim(ctx *command.Context) (string, error) {
	token, err := p.dbUpdater.acquirePlanPublicationClaim(ctx.Pull)
	if err == nil {
		return token, nil
	}
	ctx.CommandHasErrors = true
	ctx.Log.Err("acquiring plan publication claim: %s", err)
	return "", err
}

func (p *PlanCommandRunner) waitForPlanPublicationClaim(ctx *command.Context) (string, error) {
	token, err := p.dbUpdater.waitForPlanPublicationClaim(ctx.Pull)
	if err == nil {
		return token, nil
	}
	ctx.CommandHasErrors = true
	ctx.Log.Err("waiting for plan publication claim: %s", err)
	return "", err
}

// finishPlanPublicationClaim releases only after every synchronous VCS write
// returned success. An error may be ambiguous at the provider boundary, so
// retaining the non-expiring claim is the fail-closed choice until pull close
// recovery confirms the old publisher has stopped.
func (p *PlanCommandRunner) finishPlanPublicationClaim(ctx *command.Context, token string, publicationErr error) bool {
	if publicationErr != nil {
		ctx.CommandHasErrors = true
		ctx.Log.Err("publishing plan state; publication claim retained for offline recovery: %s", publicationErr)
		return false
	}
	if err := p.dbUpdater.releasePlanPublicationClaim(ctx.Pull, token); err != nil {
		ctx.CommandHasErrors = true
		ctx.Log.Err("releasing plan publication claim: %s", err)
		return false
	}
	return true
}

func (p *PlanCommandRunner) handleSupersededPlanGeneration(
	ctx *command.Context,
	generation string,
	projectCmds []command.ProjectContext,
	result command.Result,
	err error,
) {
	ctx.Log.Info("plan generation %q was superseded by a newer generation; ignoring obsolete results", generation)
	ctx.CommandHasErrors = false
	// Never mutate project VCS contexts from the obsolete command. Any
	// read-before-publish classification has another race with a still newer
	// generation. Durable cancellation is reflected by the current runner's
	// aggregate status; this stale runner only closes its job stream.
	p.completeSupersededPlanJobs(projectCmds, result)
}

func completePlanGenerationResults(projectCmds []command.ProjectContext, result command.Result) command.Result {
	completed := make(map[applyPlanKey]struct{}, len(result.ProjectResults))
	for i := range result.ProjectResults {
		projectResult := &result.ProjectResults[i]
		if projectResult.Command == command.Plan && projectResult.Error == nil && projectResult.Failure == "" && projectResult.PlanSuccess == nil {
			projectResult.Error = errors.New("plan command completed without a plan result")
		}
		completed[newApplyPlanKey(projectResult.Workspace, projectResult.RepoRelDir, projectResult.ProjectName)] = struct{}{}
	}

	for _, projectCtx := range projectCmds {
		key := newApplyPlanKey(projectCtx.Workspace, projectCtx.RepoRelDir, projectCtx.ProjectName)
		if _, ok := completed[key]; ok {
			continue
		}
		completed[key] = struct{}{}
		result.ProjectResults = append(result.ProjectResults, command.ProjectResult{
			ProjectCommandOutput: command.ProjectCommandOutput{
				Error: errors.New("plan skipped because an earlier execution-order group failed"),
			},
			Command:           command.Plan,
			SubCommand:        projectCtx.SubCommand,
			RepoRelDir:        projectCtx.RepoRelDir,
			Workspace:         projectCtx.Workspace,
			ProjectName:       projectCtx.ProjectName,
			SilencePRComments: projectCtx.SilencePRComments,
		})
	}
	return result
}

func prepareProjectCommandsForPlanGeneration(projectCmds []command.ProjectContext, generation string) {
	for i := range projectCmds {
		projectCmds[i].PlanGeneration = generation
		projectCmds[i].AcceptedPlanGeneration = generation
		if !projectCmds[i].RequiresAtlantisManagedPlanFile {
			continue
		}
		managedPlanHash := ""
		projectCmds[i].GeneratedPlanHash = &managedPlanHash
	}
}

func invalidateAutomergePlanResults(result command.Result, automergeEnabled bool) command.Result {
	if !automergeEnabled || !result.HasErrors() {
		return result
	}
	for i := range result.ProjectResults {
		projectResult := &result.ProjectResults[i]
		if projectResult.Command != command.Plan || projectResult.Error != nil || projectResult.Failure != "" {
			continue
		}
		projectResult.Failure = "plan invalidated because one or more projects failed and automerge requires all plans to succeed"
	}
	result.PlansDeleted = true
	return result
}

func (p *PlanCommandRunner) publishDeferredPlanStatuses(projectCmds []command.ProjectContext, result command.Result, status models.CommitStatus) error {
	publisher, ok := p.prjCmdRunner.(DeferredPlanStatusPublisher)
	if !ok {
		return nil
	}
	return publisher.PublishDeferredPlanStatuses(projectCmds, result, status)
}

func (p *PlanCommandRunner) publishPendingPlanStatuses(projectCmds []command.ProjectContext) error {
	publisher, ok := p.prjCmdRunner.(DeferredPlanStatusPublisher)
	if !ok {
		return nil
	}
	return publisher.PublishPendingPlanStatuses(projectCmds)
}

func (p *PlanCommandRunner) publishCancelledPlanStatuses(projectCmds []command.ProjectContext) error {
	if len(projectCmds) == 0 {
		return nil
	}
	publisher, ok := p.prjCmdRunner.(DeferredPlanStatusPublisher)
	if !ok {
		return nil
	}
	return publisher.PublishCancelledPlanStatuses(projectCmds)
}

func (p *PlanCommandRunner) publishPlanGenerationStartFailureStatuses(projectCmds []command.ProjectContext, err error) error {
	publisher, ok := p.prjCmdRunner.(DeferredPlanStatusPublisher)
	if !ok {
		return nil
	}
	return publisher.PublishPlanGenerationStartFailureStatuses(projectCmds, err)
}

func (p *PlanCommandRunner) completeSupersededPlanJobs(projectCmds []command.ProjectContext, result command.Result) {
	publisher, ok := p.prjCmdRunner.(DeferredPlanStatusPublisher)
	if !ok {
		return
	}
	publisher.CompleteSupersededPlanJobs(projectCmds, result)
}

func (p *PlanCommandRunner) completeUnpublishedPlanJobs(projectCmds []command.ProjectContext, result command.Result, message string) {
	publisher, ok := p.prjCmdRunner.(DeferredPlanStatusPublisher)
	if !ok {
		return
	}
	publisher.CompleteUnpublishedPlanJobs(projectCmds, result, message)
}

func cancelledPlanProjectContexts(ctx *command.Context, projects []db.PlanGenerationProject) []command.ProjectContext {
	contexts := make([]command.ProjectContext, 0, len(projects))
	for _, project := range projects {
		contexts = append(contexts, command.ProjectContext{
			Log:         ctx.Log,
			CommandName: command.Plan,
			BaseRepo:    ctx.Pull.BaseRepo,
			Pull:        ctx.Pull,
			Workspace:   project.Workspace,
			RepoRelDir:  project.RepoRelDir,
			ProjectName: project.ProjectName,
		})
	}
	return contexts
}

func (p *PlanCommandRunner) ShouldSkipPreWorkflowHooks(ctx *command.Context, cmd *CommentCommand) bool {
	return MarkCommandSkippedIfIgnoredTarget(ctx, command.Plan, cmd, p.prjCmdBuilder)
}

func (p *PlanCommandRunner) updatePendingCommitStatus(ctx *command.Context, commandName command.Name) error {
	if p.silenceVCSStatusNoProjects {
		ctx.Log.Debug("silence enabled - not setting pending VCS status")
		return nil
	}
	if err := p.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.PendingCommitStatus, commandName); err != nil {
		ctx.Log.Warn("unable to update commit status: %s", err)
		return err
	}
	return nil
}

func (p *PlanCommandRunner) updateNoProjectCommitStatuses(ctx *command.Context, pullStatus models.PullStatus) error {
	activeProjects := countActivePullStatusProjects(pullStatus)
	if activeProjects == 0 {
		ctx.Log.Debug("setting VCS status to success with no active projects found")
		return errors.Join(
			p.commitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.SuccessCommitStatus, command.Plan, models.ProjectCounts{}),
			p.commitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.SuccessCommitStatus, command.PolicyCheck, models.ProjectCounts{}),
			p.commitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.SuccessCommitStatus, command.Apply, models.ProjectCounts{}),
		)
	}

	// Replacing with no projects can cancel a plan that is still running on
	// another replica. Those canceled projects are terminal plan errors, not a
	// successful 0/0 plan.
	counts := models.ProjectCounts{Total: activeProjects, Errored: activeProjects}
	return errors.Join(
		p.updateCommitStatus(ctx, pullStatus, command.Plan),
		p.commitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, command.PolicyCheck, counts),
		p.commitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, command.Apply, counts),
	)
}

func (p *PlanCommandRunner) updateCommitStatus(ctx *command.Context, pullStatus models.PullStatus, commandName command.Name) error {
	var numSuccess int
	var numErrored int
	var numNoChanges int
	numProjects := countActivePullStatusProjects(pullStatus)
	status := models.SuccessCommitStatus

	switch commandName {
	case command.Plan:
		var numPending int
		for _, project := range pullStatus.Projects {
			if project.PlanGeneration != "" {
				numPending++
			} else if project.Status == models.ErroredPlanStatus {
				numErrored++
			}
		}
		// We consider anything that isn't a plan error or an incomplete plan
		// generation as a plan success.
		// For example, if there is an apply error, that means that at least a
		// plan was generated successfully.
		numSuccess = numProjects - numErrored - numPending

		if numErrored > 0 {
			status = models.FailedCommitStatus
		} else if numPending > 0 {
			status = models.PendingCommitStatus
		}
	case command.Apply:
		numNoChanges = pullStatus.StatusCount(models.PlannedNoChangesPlanStatus)
		numSuccess = pullStatus.StatusCount(models.AppliedPlanStatus) + numNoChanges
		numErrored = pullStatus.StatusCount(models.ErroredApplyStatus)

		if numErrored > 0 {
			status = models.FailedCommitStatus
		} else if numSuccess < numProjects {
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

	if err := p.commitStatusUpdater.UpdateCombinedCount(
		ctx.Log,
		ctx.Pull.BaseRepo,
		ctx.Pull,
		status,
		commandName,
		models.ProjectCounts{Success: numSuccess, Total: numProjects, Errored: numErrored, NoChanges: numNoChanges},
	); err != nil {
		ctx.Log.Warn("unable to update commit status: %s", err)
		return err
	}
	return nil
}

func countActivePullStatusProjects(pullStatus models.PullStatus) int {
	return len(pullStatus.Projects) - pullStatus.StatusCount(models.DiscardedPlanStatus)
}

// deletePlans deletes all plans generated in this ctx.
func (p *PlanCommandRunner) deletePlans(ctx *command.Context) ([]PendingPlan, error) {
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
		planPath := filepath.Join(plan.RepoDir, plan.RepoRelDir, runtime.GetPlanFilename(plan.Workspace, plan.ProjectName))
		if err := utils.RemoveIgnoreNonExistent(planPath); err != nil {
			return nil, fmt.Errorf("deleting plan at %s: %w", planPath, err)
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
