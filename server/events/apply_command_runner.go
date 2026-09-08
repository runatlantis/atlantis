// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"errors"
	"fmt"
	"slices"

	"github.com/google/uuid"

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
	Publication           *PublicationCoordinator
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
		a.reportApplyStartFailure(ctx, cmd, err, applyLockCheckFailedComment)
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
		unlockPullApply, err = a.workingDirLocker.TryLockPull(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, command.Apply, WorkingDirLockMetadataForPull(ctx.Pull))
		if err != nil {
			ctx.CommandHasErrors = true
			a.reportApplyStartFailure(ctx, cmd, err, "")
			return
		}
		defer unlockPullApply()
	}

	if err := a.refreshPullStatus(ctx, pull); err != nil {
		ctx.Log.Err("fetching current plan status: %s", err)
		ctx.CommandHasErrors = true
		a.reportApplyStartFailure(ctx, cmd, fmt.Errorf("fetching current plan status: %w", err), "")
		return
	}
	livePull, err := a.refreshLivePullIdentity(ctx)
	if err != nil {
		ctx.Log.Err("fetching live pull request: %s", err)
		ctx.CommandHasErrors = true
		a.reportApplyStartFailure(ctx, cmd, fmt.Errorf("fetching live pull request: %w", err), "")
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
		a.reportApplyStartFailure(ctx, cmd, projectCmdsErr, "")
		return
	}

	// Preserve the no-project/silence behavior while refreshing its durable
	// status and publishing inside the same short lease.
	if len(projectCmds) == 0 && a.SilenceNoProjects {
		err := a.Publication.RunCommand(ctx, func() error {
			ctx.Log.Info("determined there was no project to run apply in")
			if a.silenceVCSStatusNoProjects {
				return nil
			}
			if ctx.TerminalPublisher != nil {
				if err := a.refreshPullStatus(ctx, pull); err != nil {
					return err
				}
				var err error
				livePull, err = a.refreshLivePullIdentity(ctx)
				if err != nil {
					return err
				}
			}
			currentPull := applyPullWithLiveIdentity(pull, livePull)
			pullStatus, err := a.currentNoProjectApplyPullStatus(ctx, pull, currentPull)
			if err != nil {
				return err
			}
			if cmd.IsForSpecificProject() {
				return a.updateCommitStatus(ctx, *pullStatus)
			}
			return publishTerminal(ctx, func() error {
				return a.commitStatusUpdater.UpdateCombinedCount(ctx.Log, baseRepo, pull, models.SuccessCommitStatus, command.Apply, models.ProjectCounts{})
			})
		})
		if err != nil {
			ctx.CommandHasErrors = true
			ctx.Log.Warn("not publishing no-project apply success status because %s", err)
		}
		return
	}
	if !a.prepareApply(ctx, cmd, projectCmds) {
		return
	}

	preApplyPullStatus := ctx.PullStatus
	result := runProjectCmdsWithCancellationTracker(ctx, projectCmds, a.cancellationTracker, a.parallelPoolSize, a.isParallelEnabled(projectCmds), a.prjCmdRunner.Apply)
	completionErr := a.Publication.RunCommand(ctx, func() error {
		finalLivePull, err := a.refreshLivePullIdentity(ctx)
		if err != nil {
			return a.failApplyResult(ctx, cmd, projectCmds, result, fmt.Errorf("fetching live pull request after apply: %w", err), true)
		}
		if err := livePullIdentityChangedDuringApply(livePull, finalLivePull); err != nil {
			return a.failApplyResult(ctx, cmd, projectCmds, result, err, true)
		}
		livePull = finalLivePull
		ctx.CommandHasErrors = result.HasErrors()

		pullStatus, err := a.dbUpdater.updateDB(ctx, pull, result.ProjectResults)
		if err != nil {
			return a.failApplyResult(ctx, cmd, projectCmds, result, fmt.Errorf("persisting apply results: %w", err), false)
		}
		if err := a.pullUpdater.updatePull(ctx, cmd, result); err != nil {
			return err
		}

		currentPull := applyPullWithLiveIdentity(pull, livePull)
		if err := applyResultStatusUpdateError(result, pullStatus, pull, currentPull, preApplyPullStatus); err != nil {
			ctx.Log.Warn("not publishing apply success status because %s", err)
			ctx.CommandHasErrors = true
			if err := publishTerminal(ctx, func() error { return a.publishDeferredApplyStatuses(projectCmds, result, models.FailedCommitStatus) }); err != nil {
				return err
			}
			return publishTerminal(ctx, func() error {
				return a.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, cmd.CommandName())
			})
		}

		if err := publishTerminal(ctx, func() error { return a.publishDeferredApplyStatuses(projectCmds, result, models.SuccessCommitStatus) }); err != nil {
			return err
		}
		if err := a.updateCommitStatus(ctx, pullStatus); err != nil {
			return err
		}

		if result.HasErrors() {
			return nil
		}
		if err := pullStatusFreshnessError(currentPull, pullStatus.Pull, "recorded apply status"); err != nil {
			ctx.Log.Warn("not automerging because %s", err)
			return nil
		}

		if a.autoMerger.automergeEnabled(projectCmds) && !cmd.AutoMergeDisabled {
			if len(a.disableAutomergeLabel) > 0 {
				labels, err := a.vcsClient.GetPullLabels(ctx.Log, baseRepo, pull)
				if err != nil {
					ctx.Log.Err("unable to get pull request labels so not automerging, error %s", err)
					return nil
				} else if slices.Contains(labels, a.disableAutomergeLabel) {
					ctx.Log.Info("pull/merge request has disable automerge label %q so not automerging", a.disableAutomergeLabel)
					return nil
				}
			}
			return a.autoMerger.automerge(ctx, pullStatus, a.autoMerger.deleteSourceBranchOnMergeEnabled(projectCmds), cmd.AutoMergeMethod)
		}
		return nil
	})
	if completionErr != nil {
		ctx.CommandHasErrors = true
		if reportErr := a.pullUpdater.updatePull(ctx, cmd, command.Result{Error: fmt.Errorf("persisting or publishing apply completion: %w; do not repeat apply until its outcome is reconciled", completionErr)}); reportErr != nil {
			ctx.Log.Err("reporting command result: %s", reportErr)
		}
	}

}

func (a *ApplyCommandRunner) publishDeferredApplyStatuses(projectCmds []command.ProjectContext, result command.Result, status models.CommitStatus) error {
	publisher, ok := a.prjCmdRunner.(DeferredApplyStatusPublisher)
	if !ok {
		return nil
	}
	return publisher.PublishDeferredApplyStatuses(projectCmds, result, status)
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
		if _, ok := errors.AsType[DirNotExistErr](projectResult.Error); ok {
			continue
		}
		if pullStatus.StatusCount(models.ErroredApplyStatus) == 0 {
			return errors.New("apply result has errors but no errored apply status was recorded")
		}
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

func (a *ApplyCommandRunner) updatePendingCommitStatus(ctx *command.Context) error {
	if a.silenceVCSStatusNoProjects {
		ctx.Log.Debug("silence enabled - not setting pending VCS status")
		return nil
	}
	if err := publishTerminal(ctx, func() error {
		return a.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.PendingCommitStatus, command.Apply)
	}); err != nil {
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

	if numErrored > 0 {
		status = models.FailedCommitStatus
	} else if numSuccess < len(pullStatus.Projects) {
		// If there are plans that haven't been applied yet, we'll use a pending
		// status.
		status = models.PendingCommitStatus
	}

	if err := publishTerminal(ctx, func() error {
		return a.commitStatusUpdater.UpdateCombinedCount(
			ctx.Log,
			ctx.Pull.BaseRepo,
			ctx.Pull,
			status,
			command.Apply,
			models.ProjectCounts{Success: numSuccess, Total: len(pullStatus.Projects), Errored: numErrored, NoChanges: numNoChanges},
		)
	}); err != nil {
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

// failApplyResult makes the outcome visible even when persistence or the final
// live-head refresh fails. Successful execution is never presented as safe to
// retry after its authorization becomes ambiguous.
func (a *ApplyCommandRunner) failApplyResult(ctx *command.Context, cmd *CommentCommand, projects []command.ProjectContext, result command.Result, failure error, recordAmbiguousExecution bool) error {
	ctx.CommandHasErrors = true
	publication := result
	result.ProjectResults = slices.Clone(result.ProjectResults)
	if recordAmbiguousExecution {
		changed := false
		for i := range result.ProjectResults {
			project := &result.ProjectResults[i]
			if (project.PlanGeneration != "" || project.ApplyExecutionID != "") && project.ApplyExecuted {
				project.Error = fmt.Errorf("%w: %w", db.ErrApplyExecutionAmbiguous, failure)
				changed = true
			}
		}
		if changed {
			_, persistErr := a.dbUpdater.updateDB(ctx, ctx.Pull, result.ProjectResults)
			failure = errors.Join(db.ErrApplyExecutionAmbiguous, failure, persistErr)
		}
	}
	result.Error = failure
	if err := publishTerminal(ctx, func() error { return a.publishDeferredApplyStatuses(projects, publication, models.FailedCommitStatus) }); err != nil {
		return err
	}
	if err := publishTerminal(ctx, func() error {
		return a.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, command.Apply)
	}); err != nil {
		return err
	}
	return a.pullUpdater.updatePull(ctx, cmd, result)
}

func (a *ApplyCommandRunner) prepareApply(ctx *command.Context, cmd *CommentCommand, projects []command.ProjectContext) bool {
	err := a.Publication.RunCommand(ctx, func() error {
		if len(projects) == 0 {
			return nil
		}
		if ctx.TerminalPublisher != nil {
			live, err := a.refreshLivePullIdentity(ctx)
			if err != nil {
				return err
			}
			if a.livePullHeadFetcher != nil {
				if err := validateCommandStartIdentity(command.ProjectContext{Pull: ctx.Pull}, live); err != nil {
					return err
				}
			}
		}
		if ctx.TerminalPublisher != nil {
			// Reject obsolete/consumed plans before overwriting a newer command's
			// VCS status with Pending. The lease keeps this read valid until the
			// admission write after pending publication succeeds.
			current, err := a.dbUpdater.Database.GetPullStatus(ctx.Pull)
			if err != nil {
				return err
			}
			if _, err := db.BeginApplyExecution(current, ctx.Pull, projects, ctx.PublicationFence.Owner); err != nil {
				return err
			}
		}
		if err := a.updatePendingCommitStatus(ctx); err != nil {
			return err
		}
		if ctx.TerminalPublisher == nil {
			return a.beginDurableApply(ctx, projects)
		}
		for i := range projects {
			projects[i].PublicationDeferred = true
		}
		if publisher, ok := a.prjCmdRunner.(PendingProjectStatusPublisher); ok {
			if err := publishTerminal(ctx, func() error { return publisher.PublishPendingProjectStatuses(projects) }); err != nil {
				return err
			}
		}
		executionID := ctx.PublicationFence.Owner
		status, err := a.dbUpdater.Database.BeginApplyExecution(ctx.Pull, projects, executionID, ctx.PublicationMode())
		if err != nil {
			return err
		}
		ctx.PullStatus = &status
		for i := range projects {
			projects[i].ApplyExecutionID = executionID
		}
		return nil
	})
	if err != nil {
		ctx.CommandHasErrors = true
		if reportErr := a.pullUpdater.updatePull(ctx, cmd, command.Result{Error: fmt.Errorf("starting apply: %w", err)}); reportErr != nil {
			ctx.Log.Err("reporting command result: %s", reportErr)
		}
		return false
	}
	return true
}

func (a *ApplyCommandRunner) reportApplyStartFailure(ctx *command.Context, cmd *CommentCommand, failure error, comment string) {
	ctx.CommandHasErrors = true
	err := a.Publication.RunWithObservedStatus(ctx, a.Database, a.livePullHeadFetcher, func() error {
		if err := publishTerminal(ctx, func() error {
			return a.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, cmd.CommandName())
		}); err != nil {
			return err
		}
		if comment != "" {
			return publishTerminal(ctx, func() error {
				return a.vcsClient.CreateComment(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull.Num, comment, command.Apply.String())
			})
		}
		return a.pullUpdater.updatePull(ctx, cmd, command.Result{Error: failure})
	})
	if errors.Is(err, db.ErrPlanGenerationSuperseded) {
		ctx.Log.Warn("suppressing obsolete apply failure publication %v", err)
	} else if err != nil {
		if reportErr := a.pullUpdater.updatePull(ctx, cmd, command.Result{Error: errors.Join(failure, err)}); reportErr != nil {
			ctx.Log.Err("reporting command result: %s", reportErr)
		}
	}
}

func (a *ApplyCommandRunner) beginDurableApply(ctx *command.Context, projects []command.ProjectContext) error {
	var durable []command.ProjectContext
	for _, project := range projects {
		if project.PlanGeneration != "" {
			durable = append(durable, project)
		}
	}
	if len(durable) == 0 {
		return nil
	}
	executionID := uuid.NewString()
	status, err := a.dbUpdater.Database.BeginApplyExecution(ctx.Pull, durable, executionID, command.NoClaim{})
	if err != nil {
		return err
	}
	ctx.PullStatus = &status
	for i := range projects {
		if projects[i].PlanGeneration != "" {
			projects[i].ApplyExecutionID = executionID
		}
	}
	return nil
}
