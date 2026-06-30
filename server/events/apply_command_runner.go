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

	locked, err := a.IsLocked()
	if err != nil {
		ctx.Log.Err("checking global apply lock: %s", err)
		ctx.CommandHasErrors = true
		if statusErr := a.commitStatusUpdater.UpdateCombined(ctx.Log, baseRepo, pull, models.FailedCommitStatus, cmd.CommandName()); statusErr != nil {
			ctx.Log.Warn("unable to update commit status: %s", statusErr)
		}
		if err := a.vcsClient.CreateComment(ctx.Log, baseRepo, pull.Num, applyLockCheckFailedComment, command.Apply.String()); err != nil {
			ctx.Log.Err("unable to comment on pull request: %s", err)
		}
		return
	}

	if locked {
		ctx.Log.Info("ignoring apply command since apply disabled globally")
		if err := a.vcsClient.CreateComment(ctx.Log, baseRepo, pull.Num, applyDisabledComment, command.Apply.String()); err != nil {
			ctx.Log.Err("unable to comment on pull request: %s", err)
		}

		return
	}

	if a.DisableApplyAll && !cmd.IsForSpecificProject() {
		ctx.Log.Info("ignoring apply command without flags since apply all is disabled")
		if err := a.vcsClient.CreateComment(ctx.Log, baseRepo, pull.Num, applyAllDisabledComment, command.Apply.String()); err != nil {
			ctx.Log.Err("unable to comment on pull request: %s", err)
		}

		return
	}

	var unlockPullApply func()
	if a.workingDirLocker != nil {
		unlockPullApply, err = a.workingDirLocker.TryLockPull(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, command.Apply)
		if err != nil {
			ctx.CommandHasErrors = true
			if statusErr := a.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, cmd.CommandName()); statusErr != nil {
				ctx.Log.Warn("unable to update commit status: %s", statusErr)
			}
			a.pullUpdater.updatePull(ctx, cmd, command.Result{Error: err})
			return
		}
		defer unlockPullApply()
	}

	if err := a.refreshPullStatus(ctx, pull); err != nil {
		ctx.Log.Err("fetching current plan status: %s", err)
		ctx.CommandHasErrors = true
		if statusErr := a.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, cmd.CommandName()); statusErr != nil {
			ctx.Log.Warn("unable to update commit status: %s", statusErr)
		}
		a.pullUpdater.updatePull(ctx, cmd, command.Result{Error: fmt.Errorf("fetching current plan status: %w", err)})
		return
	}
	livePull, err := a.refreshLivePullIdentity(ctx)
	if err != nil {
		ctx.Log.Err("fetching live pull request: %s", err)
		ctx.CommandHasErrors = true
		if statusErr := a.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, cmd.CommandName()); statusErr != nil {
			ctx.Log.Warn("unable to update commit status: %s", statusErr)
		}
		a.pullUpdater.updatePull(ctx, cmd, command.Result{Error: fmt.Errorf("fetching live pull request: %w", err)})
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
		if statusErr := a.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, cmd.CommandName()); statusErr != nil {
			ctx.Log.Warn("unable to update commit status: %s", statusErr)
		}
		a.pullUpdater.updatePull(ctx, cmd, command.Result{Error: projectCmdsErr})
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
				a.updateCommitStatus(ctx, *pullStatus)
			} else {
				// With a generic apply, we set successful commit statuses
				// with 0/0 projects planned successfully because some users require
				// the Atlantis status to be passing for all pull requests.
				// Does not apply to skipped runs for specific projects
				ctx.Log.Debug("setting VCS status to success with no projects found")
				if err := a.commitStatusUpdater.UpdateCombinedCount(ctx.Log, baseRepo, pull, models.SuccessCommitStatus, command.Apply, models.ProjectCounts{}); err != nil {
					ctx.Log.Warn("unable to update commit status: %s", err)
				}
			}
		}
		return
	}
	if len(projectCmds) > 0 {
		a.updatePendingCommitStatus(ctx)
	}

	preApplyPullStatus := ctx.PullStatus
	result := runProjectCmdsWithCancellationTracker(ctx, projectCmds, a.cancellationTracker, a.parallelPoolSize, a.isParallelEnabled(projectCmds), a.prjCmdRunner.Apply)
	finalLivePull, err := a.refreshLivePullIdentity(ctx)
	if err != nil {
		ctx.Log.Err("fetching live pull request after apply: %s", err)
		ctx.CommandHasErrors = true
		result.Error = fmt.Errorf("fetching live pull request after apply: %w", err)
		if statusErr := a.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, cmd.CommandName()); statusErr != nil {
			ctx.Log.Warn("unable to update commit status: %s", statusErr)
		}
		a.pullUpdater.updatePull(ctx, cmd, result)
		return
	}
	if err := livePullIdentityChangedDuringApply(livePull, finalLivePull); err != nil {
		ctx.Log.Warn("apply result is stale because %s", err)
		ctx.CommandHasErrors = true
		result.Error = err
		a.publishDeferredApplyStatuses(projectCmds, result, models.FailedCommitStatus)
		if statusErr := a.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, cmd.CommandName()); statusErr != nil {
			ctx.Log.Warn("unable to update commit status: %s", statusErr)
		}
		a.pullUpdater.updatePull(ctx, cmd, result)
		return
	}
	livePull = finalLivePull
	ctx.CommandHasErrors = result.HasErrors()

	a.pullUpdater.updatePull(
		ctx,
		cmd,
		result)

	pullStatus, err := a.dbUpdater.updateDB(ctx, pull, result.ProjectResults)
	if err != nil {
		ctx.Log.Err("writing results: %s", err)
		return
	}

	currentPull := applyPullWithLiveIdentity(pull, livePull)
	if err := applyResultStatusUpdateError(result, pullStatus, pull, currentPull, preApplyPullStatus); err != nil {
		ctx.Log.Warn("not publishing apply success status because %s", err)
		ctx.CommandHasErrors = true
		a.publishDeferredApplyStatuses(projectCmds, result, models.FailedCommitStatus)
		if statusErr := a.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, cmd.CommandName()); statusErr != nil {
			ctx.Log.Warn("unable to update commit status: %s", statusErr)
		}
		return
	}

	a.publishDeferredApplyStatuses(projectCmds, result, models.SuccessCommitStatus)
	a.updateCommitStatus(ctx, pullStatus)

	if result.HasErrors() {
		return
	}
	if err := pullStatusFreshnessError(currentPull, pullStatus.Pull, "recorded apply status"); err != nil {
		ctx.Log.Warn("not automerging because %s", err)
		return
	}

	if a.autoMerger.automergeEnabled(projectCmds) && !cmd.AutoMergeDisabled {
		if len(a.disableAutomergeLabel) > 0 {
			labels, err := a.vcsClient.GetPullLabels(ctx.Log, baseRepo, pull)
			if err != nil {
				ctx.Log.Err("unable to get pull request labels so not automerging, error %s", err)
				return
			} else if slices.Contains(labels, a.disableAutomergeLabel) {
				ctx.Log.Info("pull/merge request has disable automerge label %q so not automerging", a.disableAutomergeLabel)
				return
			}
		}
		a.autoMerger.automerge(ctx, pullStatus, a.autoMerger.deleteSourceBranchOnMergeEnabled(projectCmds), cmd.AutoMergeMethod)
	}
}

func (a *ApplyCommandRunner) publishDeferredApplyStatuses(projectCmds []command.ProjectContext, result command.Result, status models.CommitStatus) {
	publisher, ok := a.prjCmdRunner.(DeferredApplyStatusPublisher)
	if !ok {
		return
	}
	publisher.PublishDeferredApplyStatuses(projectCmds, result, status)
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
	if result.HasErrors() && pullStatus.StatusCount(models.ErroredApplyStatus) == 0 {
		return errors.New("apply result has errors but no errored apply status was recorded")
	}
	return nil
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

func (a *ApplyCommandRunner) updatePendingCommitStatus(ctx *command.Context) {
	if a.silenceVCSStatusNoProjects {
		ctx.Log.Debug("silence enabled - not setting pending VCS status")
		return
	}
	if err := a.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.PendingCommitStatus, command.Apply); err != nil {
		ctx.Log.Warn("unable to update commit status: %s", err)
	}
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

func (a *ApplyCommandRunner) updateCommitStatus(ctx *command.Context, pullStatus models.PullStatus) {
	var numSuccess int
	var numErrored int
	var numNoChanges int
	status := models.SuccessCommitStatus

	numNoChanges = pullStatus.StatusCount(models.PlannedNoChangesPlanStatus)
	numSuccess = pullStatus.StatusCount(models.AppliedPlanStatus) + numNoChanges
	numErrored = pullStatus.StatusCount(models.ErroredApplyStatus)

	if numErrored > 0 {
		status = models.FailedCommitStatus
	} else if numSuccess < len(pullStatus.Projects) {
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
		models.ProjectCounts{Success: numSuccess, Total: len(pullStatus.Projects), Errored: numErrored, NoChanges: numNoChanges},
	); err != nil {
		ctx.Log.Warn("unable to update commit status: %s", err)
	}
}

// applyAllDisabledComment is posted when apply all commands (i.e. "atlantis apply")
// are disabled and an apply all command is issued.
var applyAllDisabledComment = "**Error:** Running `atlantis apply` without flags is disabled." +
	" You must specify which project to apply via the `-d <dir>`, `-w <workspace>` or `-p <project name>` flags."

// applyDisabledComment is posted when apply commands are disabled globally and an apply command is issued.
var applyDisabledComment = "**Error:** Running `atlantis apply` is disabled."

// applyLockCheckFailedComment is posted when the global apply lock check fails (e.g. database unreachable).
var applyLockCheckFailedComment = "**Error:** Failed to check global apply lock. Running `atlantis apply` is not allowed until the lock backend is reachable."
