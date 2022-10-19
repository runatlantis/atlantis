package events

import (
	"fmt"

	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
)

func NewApplyCommandRunner(
	vcsClient vcs.Client,
	disableApplyAll bool,
	applyCommandLocker locking.ApplyLockChecker,
	vcsStatusUpdater VCSStatusUpdater,
	prjCommandBuilder ProjectApplyCommandBuilder,
	prjCmdRunner ProjectApplyCommandRunner,
	outputUpdater OutputUpdater,
	dbUpdater *DBUpdater,
	parallelPoolSize int,
	pullReqStatusFetcher vcs.PullReqStatusFetcher,
) *ApplyCommandRunner {
	return &ApplyCommandRunner{
		vcsClient:            vcsClient,
		DisableApplyAll:      disableApplyAll,
		locker:               applyCommandLocker,
		vcsStatusUpdater:     vcsStatusUpdater,
		prjCmdBuilder:        prjCommandBuilder,
		prjCmdRunner:         prjCmdRunner,
		outputUpdater:        outputUpdater,
		dbUpdater:            dbUpdater,
		parallelPoolSize:     parallelPoolSize,
		pullReqStatusFetcher: pullReqStatusFetcher,
	}
}

type ApplyCommandRunner struct {
	DisableApplyAll      bool
	locker               locking.ApplyLockChecker
	vcsClient            vcs.Client
	vcsStatusUpdater     VCSStatusUpdater
	prjCmdBuilder        ProjectApplyCommandBuilder
	prjCmdRunner         ProjectApplyCommandRunner
	outputUpdater        OutputUpdater
	dbUpdater            *DBUpdater
	parallelPoolSize     int
	pullReqStatusFetcher vcs.PullReqStatusFetcher
}

func (a *ApplyCommandRunner) Run(ctx *command.Context, cmd *command.Comment) {
	var err error
	baseRepo := ctx.Pull.BaseRepo
	pull := ctx.Pull

	locked, err := a.IsLocked()
	// CheckApplyLock falls back to DisableApply flag if fetching the lock
	// raises an error
	// We will log failure as warning
	if err != nil {
		ctx.Log.WarnContext(ctx.RequestCtx, fmt.Sprintf("checking global apply lock: %s", err))
	}

	if locked {
		ctx.Log.InfoContext(ctx.RequestCtx, "ignoring apply command since apply disabled globally")
		if err := a.vcsClient.CreateComment(baseRepo, pull.Num, applyDisabledComment, command.Apply.String()); err != nil {
			ctx.Log.ErrorContext(ctx.RequestCtx, fmt.Sprintf("unable to comment on pull request: %s", err))
		}

		return
	}

	if a.DisableApplyAll && !cmd.IsForSpecificProject() {
		ctx.Log.InfoContext(ctx.RequestCtx, "ignoring apply command without flags since apply all is disabled")
		if err := a.vcsClient.CreateComment(baseRepo, pull.Num, applyAllDisabledComment, command.Apply.String()); err != nil {
			ctx.Log.ErrorContext(ctx.RequestCtx, fmt.Sprintf("unable to comment on pull request: %s", err))
		}

		return
	}

	statusID, err := a.vcsStatusUpdater.UpdateCombined(ctx.RequestCtx, baseRepo, pull, models.PendingVCSStatus, cmd.CommandName(), "", "")
	if err != nil {
		ctx.Log.WarnContext(ctx.RequestCtx, fmt.Sprintf("unable to update commit status: %s", err))
	}

	// Get the mergeable status before we set any build statuses of our own.
	// We do this here because when we set a "Pending" status, if users have
	// required the Atlantis status checks to pass, then we've now changed
	// the mergeability status of the pull request.
	// This sets the approved, mergeable, and sqlocked status in the context.
	ctx.PullRequestStatus, err = a.pullReqStatusFetcher.FetchPullStatus(baseRepo, pull)
	if err != nil {
		// On error we continue the request with mergeable assumed false.
		// We want to continue because not all apply's will need this status,
		// only if they rely on the mergeability requirement.
		// All PullRequestStatus fields are set to false by default when error.
		ctx.Log.WarnContext(ctx.RequestCtx, fmt.Sprintf("unable to get pull request status: %s. Continuing with mergeable and approved assumed false", err))
	}

	var projectCmds []command.ProjectContext
	projectCmds, err = a.prjCmdBuilder.BuildApplyCommands(ctx, cmd)

	if err != nil {
		if _, statusErr := a.vcsStatusUpdater.UpdateCombined(ctx.RequestCtx, ctx.Pull.BaseRepo, ctx.Pull, models.FailedVCSStatus, cmd.CommandName(), statusID, ""); statusErr != nil {
			ctx.Log.WarnContext(ctx.RequestCtx, fmt.Sprintf("unable to update commit status: %s", statusErr))
		}
		a.outputUpdater.UpdateOutput(ctx, cmd, command.Result{Error: err})
		return
	}

	// Only run commands in parallel if enabled
	var result command.Result
	if a.isParallelEnabled(projectCmds) {
		ctx.Log.InfoContext(ctx.RequestCtx, "Running applies in parallel")
		result = runProjectCmdsParallel(projectCmds, a.prjCmdRunner.Apply, a.parallelPoolSize)
	} else {
		result = runProjectCmds(projectCmds, a.prjCmdRunner.Apply)
	}

	a.outputUpdater.UpdateOutput(
		ctx,
		cmd,
		result)

	pullStatus, err := a.dbUpdater.updateDB(ctx, pull, result.ProjectResults)
	if err != nil {
		ctx.Log.ErrorContext(ctx.RequestCtx, fmt.Sprintf("writing results: %s", err))
		return
	}

	a.updateVcsStatus(ctx, pullStatus, statusID)
}

func (a *ApplyCommandRunner) IsLocked() (bool, error) {
	lock, err := a.locker.CheckApplyLock()

	return lock.Locked, err
}

func (a *ApplyCommandRunner) isParallelEnabled(projectCmds []command.ProjectContext) bool {
	return len(projectCmds) > 0 && projectCmds[0].ParallelApplyEnabled
}

func (a *ApplyCommandRunner) updateVcsStatus(ctx *command.Context, pullStatus models.PullStatus, statusID string) {
	var numSuccess int
	var numErrored int
	status := models.SuccessVCSStatus

	numSuccess = pullStatus.StatusCount(models.AppliedPlanStatus)
	numErrored = pullStatus.StatusCount(models.ErroredApplyStatus)

	if numErrored > 0 {
		status = models.FailedVCSStatus
	} else if numSuccess < len(pullStatus.Projects) {
		// If there are plans that haven't been applied yet, we'll use a pending
		// status.
		status = models.PendingVCSStatus
	}

	if _, err := a.vcsStatusUpdater.UpdateCombinedCount(
		ctx.RequestCtx,
		ctx.Pull.BaseRepo,
		ctx.Pull,
		status,
		command.Apply,
		numSuccess,
		len(pullStatus.Projects),
		statusID,
	); err != nil {
		ctx.Log.WarnContext(ctx.RequestCtx, fmt.Sprintf("unable to update commit status: %s", err))
	}
}

// applyAllDisabledComment is posted when apply all commands (i.e. "atlantis apply")
// are disabled and an apply all command is issued.
var applyAllDisabledComment = "**Error:** Running `atlantis apply` without flags is disabled." +
	" You must specify which project to apply via the `-d <dir>`, `-w <workspace>` or `-p <project name>` flags."

// applyDisabledComment is posted when apply commands are disabled globally and an apply command is issued.
var applyDisabledComment = "**Error:** Running `atlantis apply` is disabled."
