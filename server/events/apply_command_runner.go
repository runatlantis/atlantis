package events

import (
	"errors"
	
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/status"
	"github.com/runatlantis/atlantis/server/events/vcs"
)

func NewApplyCommandRunner(
	vcsClient vcs.Client,
	disableApplyAll bool,
	applyCommandLocker locking.ApplyLockChecker,
	commitStatusUpdater CommitStatusUpdater,
	prjCommandBuilder ProjectApplyCommandBuilder,
	prjCmdRunner ProjectApplyCommandRunner,
	autoMerger *AutoMerger,
	pullUpdater *PullUpdater,
	dbUpdater *DBUpdater,
	backend locking.Backend,
	parallelPoolSize int,
	SilenceNoProjects bool,
	silenceVCSStatusNoProjects bool,
	pullReqStatusFetcher vcs.PullReqStatusFetcher,
	statusManager status.StatusManager,
) *ApplyCommandRunner {
	return &ApplyCommandRunner{
		vcsClient:                  vcsClient,
		DisableApplyAll:            disableApplyAll,
		locker:                     applyCommandLocker,
		commitStatusUpdater:        commitStatusUpdater,
		prjCmdBuilder:              prjCommandBuilder,
		prjCmdRunner:               prjCmdRunner,
		autoMerger:                 autoMerger,
		pullUpdater:                pullUpdater,
		dbUpdater:                  dbUpdater,
		Backend:                    backend,
		parallelPoolSize:           parallelPoolSize,
		SilenceNoProjects:          SilenceNoProjects,
		silenceVCSStatusNoProjects: silenceVCSStatusNoProjects,
		pullReqStatusFetcher:       pullReqStatusFetcher,
		StatusManager:              statusManager,
	}
}

type ApplyCommandRunner struct {
	DisableApplyAll      bool
	Backend              locking.Backend
	locker               locking.ApplyLockChecker
	vcsClient            vcs.Client
	commitStatusUpdater  CommitStatusUpdater
	prjCmdBuilder        ProjectApplyCommandBuilder
	prjCmdRunner         ProjectApplyCommandRunner
	autoMerger           *AutoMerger
	pullUpdater          *PullUpdater
	dbUpdater            *DBUpdater
	parallelPoolSize     int
	pullReqStatusFetcher vcs.PullReqStatusFetcher
	StatusManager        status.StatusManager
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

	locked, err := a.IsLocked()
	// CheckApplyLock falls back to AllowedCommand flag if fetching the lock
	// raises an error
	// We will log failure as warning
	if err != nil {
		ctx.Log.Warn("checking global apply lock: %s", err)
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

	var projectCmds []command.ProjectContext
	projectCmds, err = a.prjCmdBuilder.BuildApplyCommands(ctx, cmd)

	if err != nil {
		if statusErr := a.StatusManager.SetFailure(ctx, cmd.CommandName(), err); statusErr != nil {
			ctx.Log.Warn("unable to update commit status: %s", statusErr)
		}
		a.pullUpdater.updatePull(ctx, cmd, command.Result{Error: err})
		return
	}

	// If there are no projects to apply, don't respond to the PR and ignore
	if len(projectCmds) == 0 && a.SilenceNoProjects {
		ctx.Log.Info("determined there was no project to run plan in")
		// Use StatusManager to handle no projects found with policy-aware decisions
		if err := a.StatusManager.HandleNoProjectsFound(ctx, cmd.CommandName()); err != nil {
			ctx.Log.Warn("unable to handle no projects status: %s", err)
		}
		return
	}

	// Only run commands in parallel if enabled
	var result command.Result
	if a.isParallelEnabled(projectCmds) {
		ctx.Log.Info("Running applies in parallel")
		result = runProjectCmdsParallelGroups(ctx, projectCmds, a.prjCmdRunner.Apply, a.parallelPoolSize)
	} else {
		result = runProjectCmds(projectCmds, a.prjCmdRunner.Apply)
	}
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

	a.updateCommitStatus(ctx, pullStatus)

	if a.autoMerger.automergeEnabled(projectCmds) && !cmd.AutoMergeDisabled {
		a.autoMerger.automerge(ctx, pullStatus, a.autoMerger.deleteSourceBranchOnMergeEnabled(projectCmds), cmd.AutoMergeMethod)
	}
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
	
	numSuccess = pullStatus.StatusCount(models.AppliedPlanStatus) + pullStatus.StatusCount(models.PlannedNoChangesPlanStatus)
	numErrored = pullStatus.StatusCount(models.ErroredApplyStatus)

	if numErrored > 0 {
		// Use a fake error for the failure - StatusManager will handle the status setting
		err := errors.New("apply failed for one or more projects")
		if statusErr := a.StatusManager.SetFailure(ctx, command.Apply, err); statusErr != nil {
			ctx.Log.Warn("unable to update commit status: %s", statusErr)
		}
	} else if numSuccess < len(pullStatus.Projects) {
		// If there are plans that haven't been applied yet, set pending
		if statusErr := a.StatusManager.SetPending(ctx, command.Apply); statusErr != nil {
			ctx.Log.Warn("unable to update commit status: %s", statusErr)
		}
	} else {
		// All successful
		if statusErr := a.StatusManager.SetSuccess(ctx, command.Apply, numSuccess, len(pullStatus.Projects)); statusErr != nil {
			ctx.Log.Warn("unable to update commit status: %s", statusErr)
		}
	}
}

// applyAllDisabledComment is posted when apply all commands (i.e. "atlantis apply")
// are disabled and an apply all command is issued.
var applyAllDisabledComment = "**Error:** Running `atlantis apply` without flags is disabled." +
	" You must specify which project to apply via the `-d <dir>`, `-w <workspace>` or `-p <project name>` flags."

// applyDisabledComment is posted when apply commands are disabled globally and an apply command is issued.
var applyDisabledComment = "**Error:** Running `atlantis apply` is disabled."
