// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"errors"
	"fmt"
	"slices"

	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
)

func NewApplyCommandRunner(
	vcsClient vcs.Client,
	disableApplyAll bool,
	applyCommandLocker locking.ApplyLockChecker,
	commitStatusUpdater CommitStatusUpdater,
	prjCommandBuilder ProjectApplyCommandBuilder,
	prjCmdRunner ProjectApplyCommandRunner,
	cancellationTracker CancellationTracker,
	autoMerger *AutoMerger,
	pullUpdater *PullUpdater,
	dbUpdater *DBUpdater,
	database db.Database,
	parallelPoolSize int,
	SilenceNoProjects bool,
	silenceVCSStatusNoProjects bool,
	workingDirLocker WorkingDirLocker,
	pullReqStatusFetcher vcs.PullReqStatusFetcher,
	livePullHeadFetcher LivePullHeadFetcher,
	disableAutomergeLabel string,
) *ApplyCommandRunner {
	return &ApplyCommandRunner{
		vcsClient:                  vcsClient,
		DisableApplyAll:            disableApplyAll,
		locker:                     applyCommandLocker,
		commitStatusUpdater:        commitStatusUpdater,
		prjCmdBuilder:              prjCommandBuilder,
		prjCmdRunner:               prjCmdRunner,
		cancellationTracker:        cancellationTracker,
		autoMerger:                 autoMerger,
		pullUpdater:                pullUpdater,
		dbUpdater:                  dbUpdater,
		Database:                   database,
		parallelPoolSize:           parallelPoolSize,
		SilenceNoProjects:          SilenceNoProjects,
		silenceVCSStatusNoProjects: silenceVCSStatusNoProjects,
		workingDirLocker:           workingDirLocker,
		pullReqStatusFetcher:       pullReqStatusFetcher,
		livePullHeadFetcher:        livePullHeadFetcher,
		disableAutomergeLabel:      disableAutomergeLabel,
	}
}

type ApplyCommandRunner struct {
	DisableApplyAll       bool
	Database              db.Database
	locker                locking.ApplyLockChecker
	vcsClient             vcs.Client
	commitStatusUpdater   CommitStatusUpdater
	prjCmdBuilder         ProjectApplyCommandBuilder
	prjCmdRunner          ProjectApplyCommandRunner
	cancellationTracker   CancellationTracker
	autoMerger            *AutoMerger
	pullUpdater           *PullUpdater
	dbUpdater             *DBUpdater
	parallelPoolSize      int
	workingDirLocker      WorkingDirLocker
	pullReqStatusFetcher  vcs.PullReqStatusFetcher
	livePullHeadFetcher   LivePullHeadFetcher
	disableAutomergeLabel string
	// SilenceNoProjects is whether Atlantis should respond to PRs if no projects
	// are found
	SilenceNoProjects bool
	// SilenceVCSStatusNoPlans is whether any plan should set commit status if no projects
	// are found
	silenceVCSStatusNoProjects bool
	SilencePRComments          []string
}

func (a *ApplyCommandRunner) Run(ctx *command.Context, cmd *CommentCommand) {
	var err error
	baseRepo := ctx.Pull.BaseRepo
	pull := ctx.Pull

	var projectCmds []command.ProjectContext
	var projectCmdsErr error

	if a.ShouldSkipPreWorkflowHooks(ctx, cmd) {
		return
	}
	if a.skipIgnoredTargetedDirBeforeApplyLocks(ctx, cmd) {
		return
	}
	publicationToken, claimErr := a.acquirePlanPublicationClaim(ctx)
	if claimErr != nil {
		return
	}
	claimActive := true
	finishClaim := func(publicationErr error) {
		if !claimActive {
			return
		}
		claimActive = false
		if !a.finishPlanPublicationClaim(ctx, publicationToken, publicationErr) {
			return
		}
	}
	defer func() { finishClaim(nil) }()

	locked, err := a.IsLocked()
	if err != nil {
		ctx.Log.Err("checking global apply lock: %s", err)
		ctx.CommandHasErrors = true
		finishClaim(errors.Join(
			a.commitStatusUpdater.UpdateCombined(ctx.Log, baseRepo, pull, models.FailedCommitStatus, cmd.CommandName()),
			a.vcsClient.CreateComment(ctx.Log, baseRepo, pull.Num, applyLockCheckFailedComment, command.Apply.String()),
		))
		return
	}

	if locked {
		ctx.Log.Info("ignoring apply command since apply disabled globally")
		finishClaim(a.vcsClient.CreateComment(ctx.Log, baseRepo, pull.Num, applyDisabledComment, command.Apply.String()))
		return
	}

	if a.DisableApplyAll && !cmd.IsForSpecificProject() {
		ctx.Log.Info("ignoring apply command without flags since apply all is disabled")
		finishClaim(a.vcsClient.CreateComment(ctx.Log, baseRepo, pull.Num, applyAllDisabledComment, command.Apply.String()))
		return
	}

	var unlockPullApply func()
	if a.workingDirLocker != nil {
		unlockPullApply, err = a.workingDirLocker.TryLockPull(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, command.Apply, WorkingDirLockMetadataForPull(ctx.Pull))
		if err != nil {
			ctx.CommandHasErrors = true
			finishClaim(errors.Join(
				a.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, cmd.CommandName()),
				a.pullUpdater.updatePull(ctx, cmd, command.Result{Error: err}),
			))
			return
		}
		defer unlockPullApply()
	}

	if err := a.refreshPullStatus(ctx, pull); err != nil {
		ctx.Log.Err("fetching current plan status: %s", err)
		ctx.CommandHasErrors = true
		finishClaim(errors.Join(
			a.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, cmd.CommandName()),
			a.pullUpdater.updatePull(ctx, cmd, command.Result{Error: fmt.Errorf("fetching current plan status: %w", err)}),
		))
		return
	}
	livePull, err := a.refreshLivePullIdentity(ctx)
	if err != nil {
		ctx.Log.Err("fetching live pull request: %s", err)
		ctx.CommandHasErrors = true
		finishClaim(errors.Join(
			a.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, cmd.CommandName()),
			a.pullUpdater.updatePull(ctx, cmd, command.Result{Error: fmt.Errorf("fetching live pull request: %w", err)}),
		))
		return
	}
	if livePull.HeadCommit != "" && !cmd.IsForSpecificProject() {
		ctx.Pull.HeadCommit = livePull.HeadCommit
		if livePull.BaseBranch != "" {
			ctx.Pull.BaseBranch = livePull.BaseBranch
		}
		pull = ctx.Pull
	}

	// Get the mergeable status before we set any build statuses of our own.
	// We do this here because when we set a "Pending" status, if users have
	// required the Atlantis status checks to pass, then we've now changed
	// the mergeability status of the pull request.
	// This sets the approved, mergeable, and sqlocked status in the context.
	ctx.PullRequestStatus, err = a.pullReqStatusFetcher.FetchPullStatus(ctx.Log, pull)
	if err != nil {
		// On error we continue the request with mergeable assumed false.
		// We want to continue because not all apply's will need this status,
		// only if they rely on the mergeability requirement.
		// All PullRequestStatus fields are set to false by default when error.
		ctx.Log.Warn("unable to get pull request status: %s. Continuing with mergeable and approved assumed false", err)
	}
	projectCmds, projectCmdsErr = a.prjCmdBuilder.BuildApplyCommands(ctx, cmd)
	if MarkCommandSkippedIfIgnoredTargetedDir(ctx, cmd.CommandName(), projectCmdsErr) {
		return
	}
	if projectCmdsErr != nil {
		finishClaim(errors.Join(
			a.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, cmd.CommandName()),
			a.pullUpdater.updatePull(ctx, cmd, command.Result{Error: projectCmdsErr}),
		))
		return
	}

	// If there are no projects to apply, don't respond to the PR and ignore
	if len(projectCmds) == 0 && a.SilenceNoProjects {
		ctx.Log.Info("determined there was no project to run plan in")
		if !a.silenceVCSStatusNoProjects {
			currentPull := applyPullWithLiveIdentity(pull, livePull)
			pullStatus, err := a.currentNoProjectApplyPullStatus(ctx, pull, currentPull)
			if err != nil {
				ctx.Log.Warn("not publishing no-project apply success status because %s", err)
				ctx.CommandHasErrors = true
				return
			}
			if cmd.IsForSpecificProject() {
				// With a specific apply, just reset the status so it's not stuck in pending state
				ctx.Log.Debug("resetting VCS status")
				finishClaim(a.updateCommitStatus(ctx, *pullStatus))
			} else {
				// With a generic apply, we set successful commit statuses
				// with 0/0 projects planned successfully because some users require
				// the Atlantis status to be passing for all pull requests.
				// Does not apply to skipped runs for specific projects
				ctx.Log.Debug("setting VCS status to success with no projects found")
				finishClaim(a.commitStatusUpdater.UpdateCombinedCount(ctx.Log, baseRepo, pull, models.SuccessCommitStatus, command.Apply, models.ProjectCounts{}))
			}
		}
		return
	}
	if len(projectCmds) > 0 {
		if statusErr := a.updatePendingCommitStatus(ctx); statusErr != nil {
			finishClaim(statusErr)
			return
		}
	}

	preApplyPullStatus := copyPullStatus(ctx.PullStatus)
	result := runProjectCmdsWithCancellationTracker(ctx, projectCmds, a.cancellationTracker, a.parallelPoolSize, a.isParallelEnabled(projectCmds), a.prjCmdRunner.Apply)
	finalLivePull, err := a.refreshLivePullIdentity(ctx)
	if err != nil {
		ctx.Log.Err("fetching live pull request after apply: %s", err)
		ctx.CommandHasErrors = true
		result.Error = fmt.Errorf("fetching live pull request after apply: %w", err)
		finishClaim(errors.Join(
			a.publishDeferredApplyStatuses(projectCmds, result, models.FailedCommitStatus),
			a.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, cmd.CommandName()),
			a.pullUpdater.updatePull(ctx, cmd, result),
		))
		return
	}
	if err := livePullIdentityChangedDuringApply(livePull, finalLivePull); err != nil {
		ctx.Log.Warn("apply result is stale because %s", err)
		ctx.CommandHasErrors = true
		result.Error = err
		publicationErr := errors.Join(
			a.publishDeferredApplyStatuses(projectCmds, result, models.FailedCommitStatus),
			a.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, cmd.CommandName()),
			a.pullUpdater.updatePull(ctx, cmd, result),
		)
		if publicationErr != nil {
			finishClaim(publicationErr)
		}
		return
	}
	livePull = finalLivePull
	ctx.CommandHasErrors = result.HasErrors()

	var pullStatus models.PullStatus
	if len(result.ProjectResults) == 0 {
		pullStatus, err = a.dbUpdater.updateDB(ctx, pull, nil)
	} else {
		persistedResults := applyResultsForDurableUpdate(result.ProjectResults, preApplyPullStatus)
		pullStatus, err = a.dbUpdater.updateApplyResultsForPlanGeneration(ctx, pull, persistedResults, publicationToken)
	}
	if err != nil {
		if db.IsPlanGenerationObsolete(err) {
			ctx.Log.Warn("apply results were superseded by a newer plan generation after execution; durable plan state was not changed")
			ctx.CommandHasErrors = true
			finishClaim(nil)
			return
		}
		ctx.Log.Err("writing results: %s", err)
		ctx.CommandHasErrors = true
		result.Error = fmt.Errorf("writing apply results after execution; one or more apply steps may have completed, so verify infrastructure state before retrying: %w", err)
		publicationErr := errors.Join(
			a.publishDeferredApplyStatuses(projectCmds, result, models.FailedCommitStatus),
			a.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, cmd.CommandName()),
			a.pullUpdater.updatePull(ctx, cmd, result),
		)
		finishClaim(errors.Join(publicationErr, err))
		return
	}
	ctx.PullStatus = &pullStatus

	currentPull := applyPullWithLiveIdentity(pull, livePull)
	if err := applyResultStatusUpdateError(result, pullStatus, pull, currentPull, preApplyPullStatus); err != nil {
		ctx.Log.Warn("not publishing apply success status because %s", err)
		ctx.CommandHasErrors = true
		result.Error = err
		publicationErr := errors.Join(
			a.publishDeferredApplyStatuses(projectCmds, result, models.FailedCommitStatus),
			a.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, cmd.CommandName()),
			a.pullUpdater.updatePull(ctx, cmd, result),
		)
		finishClaim(publicationErr)
		return
	}

	publicationErr := errors.Join(
		a.pullUpdater.updatePull(ctx, cmd, result),
		a.publishDeferredApplyStatuses(projectCmds, result, models.SuccessCommitStatus),
		a.updateCommitStatus(ctx, pullStatus),
	)
	if publicationErr != nil {
		finishClaim(publicationErr)
		return
	}

	if result.HasErrors() {
		finishClaim(nil)
		return
	}
	if err := pullStatusFreshnessError(currentPull, pullStatus.Pull, "recorded apply status"); err != nil {
		ctx.Log.Warn("not automerging because %s", err)
		finishClaim(nil)
		return
	}

	if a.autoMerger.automergeEnabled(projectCmds) && !cmd.AutoMergeDisabled {
		if len(a.disableAutomergeLabel) > 0 {
			labels, err := a.vcsClient.GetPullLabels(ctx.Log, baseRepo, pull)
			if err != nil {
				ctx.Log.Err("unable to get pull request labels so not automerging, error %s", err)
				finishClaim(nil)
				return
			} else if slices.Contains(labels, a.disableAutomergeLabel) {
				ctx.Log.Info("pull/merge request has disable automerge label %q so not automerging", a.disableAutomergeLabel)
				finishClaim(nil)
				return
			}
		}
		if err := a.autoMerger.automerge(ctx, pullStatus, a.autoMerger.deleteSourceBranchOnMergeEnabled(projectCmds), cmd.AutoMergeMethod); err != nil {
			finishClaim(err)
			return
		}
	}
	finishClaim(nil)
}

func (a *ApplyCommandRunner) publishDeferredApplyStatuses(projectCmds []command.ProjectContext, result command.Result, status models.CommitStatus) error {
	publisher, ok := a.prjCmdRunner.(DeferredApplyStatusPublisher)
	if !ok {
		return nil
	}
	return publisher.PublishDeferredApplyStatuses(projectCmds, result, status)
}

func (a *ApplyCommandRunner) acquirePlanPublicationClaim(ctx *command.Context) (string, error) {
	token, err := a.dbUpdater.acquirePlanPublicationClaim(ctx.Pull)
	if err == nil {
		return token, nil
	}
	ctx.CommandHasErrors = true
	ctx.Log.Err("acquiring plan publication claim for apply: %s", err)
	return "", err
}

func (a *ApplyCommandRunner) finishPlanPublicationClaim(ctx *command.Context, token string, publicationErr error) bool {
	if publicationErr != nil {
		ctx.CommandHasErrors = true
		ctx.Log.Err("publishing apply state; publication claim retained for offline recovery: %s", publicationErr)
		return false
	}
	if err := a.dbUpdater.releasePlanPublicationClaim(ctx.Pull, token); err != nil {
		ctx.CommandHasErrors = true
		ctx.Log.Err("releasing plan publication claim after apply: %s", err)
		return false
	}
	return true
}

func livePullIdentityChangedDuringApply(before models.PullRequest, after models.PullRequest) error {
	if before.HeadCommit != "" && after.HeadCommit != "" && before.HeadCommit != after.HeadCommit {
		return fmt.Errorf(
			"%w: pull request head changed from %s to %s while apply was running; run `atlantis plan` before apply",
			errStaleCommandHead,
			shortSHA(before.HeadCommit),
			shortSHA(after.HeadCommit),
		)
	}
	if before.BaseBranch != "" && after.BaseBranch != "" && before.BaseBranch != after.BaseBranch {
		return fmt.Errorf(
			"%w: pull request base branch changed from %q to %q while apply was running; run `atlantis plan` before apply",
			errStaleCommandHead,
			before.BaseBranch,
			after.BaseBranch,
		)
	}
	return nil
}

func applyPullWithLiveIdentity(pull models.PullRequest, livePull models.PullRequest) models.PullRequest {
	currentPull := pull
	if livePull.HeadCommit != "" {
		currentPull.HeadCommit = livePull.HeadCommit
	}
	if livePull.BaseBranch != "" {
		currentPull.BaseBranch = livePull.BaseBranch
	}
	return currentPull
}

func applyResultStatusUpdateError(result command.Result, pullStatus models.PullStatus, commandPull models.PullRequest, currentPull models.PullRequest, preApplyPullStatus *models.PullStatus) error {
	if len(result.ProjectResults) == 0 {
		if preApplyPullStatus == nil {
			return errors.New("apply produced no project results and no recorded plan status was available")
		}
		if err := pullStatusApplyEligibilityError(currentPull, preApplyPullStatus.Pull, "recorded plan status"); err != nil {
			return err
		}
	}
	if staleApplyResultForCurrentPull(commandPull, result.ProjectResults) && !pullStatusFreshForPull(commandPull, pullStatus.Pull) {
		return fmt.Errorf(
			"%w: apply result was for head %s base %q but recorded apply status is for head %s base %q",
			errStaleCommandHead,
			shortSHA(commandPull.HeadCommit),
			commandPull.BaseBranch,
			shortSHA(pullStatus.Pull.HeadCommit),
			pullStatus.Pull.BaseBranch,
		)
	}
	if applyResultHasStaleCommandHead(result.ProjectResults) {
		return fmt.Errorf("%w: apply result is stale", errStaleCommandHead)
	}
	if err := pullStatusApplyEligibilityError(currentPull, pullStatus.Pull, "recorded apply status"); err != nil {
		return err
	}
	for _, projectResult := range result.ProjectResults {
		if projectResult.Error == nil && projectResult.Failure == "" {
			continue
		}
		if isPolicyBlockedApplyResult(projectResult, preApplyPullStatus) {
			continue
		}
		projectStatus := findProjectInPullStatus(&pullStatus, projectResult.Workspace, projectResult.RepoRelDir, projectResult.ProjectName)
		if projectStatus == nil || projectStatus.Status != models.ErroredApplyStatus ||
			projectStatus.PlanGeneration != "" || projectStatus.AcceptedPlanGeneration != projectResult.AcceptedPlanGeneration {
			return fmt.Errorf(
				"apply result for dir %q workspace %q project %q has errors but no matching errored apply status was recorded",
				projectResult.RepoRelDir, projectResult.Workspace, projectResult.ProjectName,
			)
		}
	}
	return nil
}

func copyPullStatus(status *models.PullStatus) *models.PullStatus {
	if status == nil {
		return nil
	}
	copy := *status
	copy.Projects = append([]models.ProjectStatus(nil), status.Projects...)
	return &copy
}

func applyResultsForDurableUpdate(results []command.ProjectResult, preApplyPullStatus *models.PullStatus) []command.ProjectResult {
	filtered := make([]command.ProjectResult, 0, len(results))
	for _, result := range results {
		if isPolicyBlockedApplyResult(result, preApplyPullStatus) {
			continue
		}
		filtered = append(filtered, result)
	}
	return filtered
}

func isPolicyBlockedApplyResult(result command.ProjectResult, preApplyPullStatus *models.PullStatus) bool {
	if preApplyPullStatus == nil || result.Command != command.Apply || (result.Error == nil && result.Failure == "") {
		return false
	}
	projectStatus := findProjectInPullStatus(preApplyPullStatus, result.Workspace, result.RepoRelDir, result.ProjectName)
	return projectStatus != nil &&
		projectStatus.Status == models.ErroredPolicyCheckStatus &&
		projectStatus.PlanGeneration == "" &&
		projectStatus.AcceptedPlanGeneration == result.AcceptedPlanGeneration
}

func applyResultHasStaleCommandHead(results []command.ProjectResult) bool {
	for _, result := range results {
		if errors.Is(result.Error, errStaleCommandHead) {
			return true
		}
	}
	return false
}

func (a *ApplyCommandRunner) currentNoProjectApplyPullStatus(ctx *command.Context, pull models.PullRequest, currentPull models.PullRequest) (*models.PullStatus, error) {
	pullStatus := ctx.PullStatus
	if pullStatus == nil && a.Database != nil {
		var err error
		pullStatus, err = a.Database.GetPullStatus(pull)
		if err != nil {
			return nil, fmt.Errorf("fetching recorded plan status: %w", err)
		}
	}
	if pullStatus == nil {
		return nil, errors.New("no recorded plan status found")
	}
	if err := pullStatusApplyEligibilityError(currentPull, pullStatus.Pull, "recorded plan status"); err != nil {
		return nil, err
	}
	return pullStatus, nil
}

func (a *ApplyCommandRunner) refreshPullStatus(ctx *command.Context, pull models.PullRequest) error {
	if a.Database == nil {
		return nil
	}
	pullStatus, err := a.Database.GetPullStatus(pull)
	if err != nil {
		return err
	}
	ctx.PullStatus = pullStatus
	return nil
}

func (a *ApplyCommandRunner) refreshLivePullIdentity(ctx *command.Context) (models.PullRequest, error) {
	if a.livePullHeadFetcher == nil {
		return models.PullRequest{}, nil
	}
	livePull, err := a.livePullHeadFetcher.GetLivePullIdentity(command.ProjectContext{
		Log:        ctx.Log,
		Pull:       ctx.Pull,
		PullStatus: ctx.PullStatus,
		API:        ctx.API,
	})
	if err != nil {
		return models.PullRequest{}, err
	}
	if livePull.HeadCommit == "" {
		return models.PullRequest{}, fmt.Errorf("live pull request head is empty")
	}
	return livePull, nil
}

func (a *ApplyCommandRunner) updatePendingCommitStatus(ctx *command.Context) error {
	if a.silenceVCSStatusNoProjects {
		ctx.Log.Debug("silence enabled - not setting pending VCS status")
		return nil
	}
	if err := a.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.PendingCommitStatus, command.Apply); err != nil {
		ctx.Log.Warn("unable to update commit status: %s", err)
		return err
	}
	return nil
}

func (a *ApplyCommandRunner) ShouldSkipPreWorkflowHooks(ctx *command.Context, cmd *CommentCommand) bool {
	return MarkCommandSkippedIfIgnoredTarget(ctx, command.Apply, cmd, a.prjCmdBuilder)
}

func (a *ApplyCommandRunner) skipIgnoredTargetedDirBeforeApplyLocks(ctx *command.Context, cmd *CommentCommand) bool {
	if cmd.ProjectName != "" || cmd.RepoRelDir == "" {
		return false
	}
	_, err := a.prjCmdBuilder.BuildApplyCommands(ctx, cmd)
	return MarkCommandSkippedIfIgnoredTargetedDir(ctx, cmd.CommandName(), err)
}

func (a *ApplyCommandRunner) IsLocked() (bool, error) {
	lock, err := a.locker.CheckApplyLock()

	return lock.Locked, err
}

func (a *ApplyCommandRunner) isParallelEnabled(projectCmds []command.ProjectContext) bool {
	return len(projectCmds) > 0 && projectCmds[0].ParallelApplyEnabled
}

func (a *ApplyCommandRunner) updateCommitStatus(ctx *command.Context, pullStatus models.PullStatus) error {
	var numSuccess int
	var numErrored int
	var numNoChanges int
	status := models.SuccessCommitStatus

	numNoChanges = pullStatus.StatusCount(models.PlannedNoChangesPlanStatus)
	numSuccess = pullStatus.StatusCount(models.AppliedPlanStatus) + numNoChanges
	numErrored = pullStatus.StatusCount(models.ErroredApplyStatus)
	numProjects := countActivePullStatusProjects(pullStatus)

	if numErrored > 0 {
		status = models.FailedCommitStatus
	} else if numSuccess < numProjects {
		// If there are plans that haven't been applied yet, we'll use a pending
		// status.
		status = models.PendingCommitStatus
	}

	if err := a.commitStatusUpdater.UpdateCombinedCount(
		ctx.Log,
		ctx.Pull.BaseRepo,
		ctx.Pull,
		status,
		command.Apply,
		models.ProjectCounts{Success: numSuccess, Total: numProjects, Errored: numErrored, NoChanges: numNoChanges},
	); err != nil {
		ctx.Log.Warn("unable to update commit status: %s", err)
		return err
	}
	return nil
}

// applyAllDisabledComment is posted when apply all commands (i.e. "atlantis apply")
// are disabled and an apply all command is issued.
var applyAllDisabledComment = "**Error:** Running `atlantis apply` without flags is disabled." +
	" You must specify which project to apply via the `-d <dir>`, `-w <workspace>` or `-p <project name>` flags."

// applyDisabledComment is posted when apply commands are disabled globally and an apply command is issued.
var applyDisabledComment = "**Error:** Running `atlantis apply` is disabled."

// applyLockCheckFailedComment is posted when the global apply lock check fails (e.g. database unreachable).
var applyLockCheckFailedComment = "**Error:** Failed to check global apply lock. Running `atlantis apply` is not allowed until the lock backend is reachable."
