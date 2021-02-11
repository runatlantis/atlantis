package events

import (
	"github.com/runatlantis/atlantis/server/events/db"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
)

func NewApplyCommandRunner(
	vcsClient vcs.Client,
	disableApplyAll bool,
	disableApply bool,
	commitStatusUpdater CommitStatusUpdater,
	prjCommandBuilder ProjectApplyCommandBuilder,
	prjCmdRunner ProjectApplyCommandRunner,
	autoMerger *AutoMerger,
	pullUpdater *PullUpdater,
	dbUpdater *DBUpdater,
	db *db.BoltDB,
	parallelPoolSize int,
) *ApplyCommandRunner {
	return &ApplyCommandRunner{
		vcsClient:           vcsClient,
		DisableApplyAll:     disableApplyAll,
		DisableApply:        disableApply,
		commitStatusUpdater: commitStatusUpdater,
		prjCmdBuilder:       prjCommandBuilder,
		prjCmdRunner:        prjCmdRunner,
		autoMerger:          autoMerger,
		pullUpdater:         pullUpdater,
		dbUpdater:           dbUpdater,
		DB:                  db,
		parallelPoolSize:    parallelPoolSize,
	}
}

type ApplyCommandRunner struct {
	DisableApplyAll     bool
	DisableApply        bool
	DB                  *db.BoltDB
	vcsClient           vcs.Client
	commitStatusUpdater CommitStatusUpdater
	prjCmdBuilder       ProjectApplyCommandBuilder
	prjCmdRunner        ProjectApplyCommandRunner
	autoMerger          *AutoMerger
	pullUpdater         *PullUpdater
	dbUpdater           *DBUpdater
	parallelPoolSize    int
}

func (a *ApplyCommandRunner) Run(ctx *CommandContext, cmd *CommentCommand) {
	var err error
	baseRepo := ctx.Pull.BaseRepo
	pull := ctx.Pull

	if a.DisableApply {
		ctx.Log.Info("ignoring apply command since apply disabled globally")
		if err := a.vcsClient.CreateComment(baseRepo, pull.Num, applyDisabledComment, models.ApplyCommand.String()); err != nil {
			ctx.Log.Err("unable to comment on pull request: %s", err)
		}

		return
	}

	if a.DisableApplyAll && !cmd.IsForSpecificProject() {
		ctx.Log.Info("ignoring apply command without flags since apply all is disabled")
		if err := a.vcsClient.CreateComment(baseRepo, pull.Num, applyAllDisabledComment, models.ApplyCommand.String()); err != nil {
			ctx.Log.Err("unable to comment on pull request: %s", err)
		}

		return
	}

	// Get the mergeable status before we set any build statuses of our own.
	// We do this here because when we set a "Pending" status, if users have
	// required the Atlantis status checks to pass, then we've now changed
	// the mergeability status of the pull request.
	ctx.PullMergeable, err = a.vcsClient.PullIsMergeable(baseRepo, pull)
	if err != nil {
		// On error we continue the request with mergeable assumed false.
		// We want to continue because not all apply's will need this status,
		// only if they rely on the mergeability requirement.
		ctx.PullMergeable = false
		ctx.Log.Warn("unable to get mergeable status: %s. Continuing with mergeable assumed false", err)
	}

	// TODO: This needs to be revisited and new PullMergeable like conditions should
	// be added to check against it.
	if a.anyFailedPolicyChecks(pull) {
		ctx.PullMergeable = false
		ctx.Log.Warn("when using policy checks all policies have to be approved or pass. Continuing with mergeable assumed false")
	}

	ctx.Log.Info("pull request mergeable status: %t", ctx.PullMergeable)

	if err = a.commitStatusUpdater.UpdateCombined(baseRepo, pull, models.PendingCommitStatus, cmd.CommandName()); err != nil {
		ctx.Log.Warn("unable to update commit status: %s", err)
	}

	var projectCmds []models.ProjectCommandContext
	projectCmds, err = a.prjCmdBuilder.BuildApplyCommands(ctx, cmd)

	if err != nil {
		if statusErr := a.commitStatusUpdater.UpdateCombined(ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, cmd.CommandName()); statusErr != nil {
			ctx.Log.Warn("unable to update commit status: %s", statusErr)
		}
		a.pullUpdater.updatePull(ctx, cmd, CommandResult{Error: err})
		return
	}

	// Only run commands in parallel if enabled
	var result CommandResult
	if a.isParallelEnabled(projectCmds) {
		ctx.Log.Info("Running applies in parallel")
		result = runProjectCmdsParallel(projectCmds, a.prjCmdRunner.Apply, a.parallelPoolSize)
	} else {
		result = runProjectCmds(projectCmds, a.prjCmdRunner.Apply)
	}

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

	if a.autoMerger.automergeEnabled(projectCmds) {
		a.autoMerger.automerge(ctx, pullStatus)
	}
}

func (a *ApplyCommandRunner) isParallelEnabled(projectCmds []models.ProjectCommandContext) bool {
	return len(projectCmds) > 0 && projectCmds[0].ParallelApplyEnabled
}

func (a *ApplyCommandRunner) updateCommitStatus(ctx *CommandContext, pullStatus models.PullStatus) {
	var numSuccess int
	var numErrored int
	status := models.SuccessCommitStatus

	numSuccess = pullStatus.StatusCount(models.AppliedPlanStatus)
	numErrored = pullStatus.StatusCount(models.ErroredApplyStatus)

	if numErrored > 0 {
		status = models.FailedCommitStatus
	} else if numSuccess < len(pullStatus.Projects) {
		// If there are plans that haven't been applied yet, we'll use a pending
		// status.
		status = models.PendingCommitStatus
	}

	if err := a.commitStatusUpdater.UpdateCombinedCount(
		ctx.Pull.BaseRepo,
		ctx.Pull,
		status,
		models.ApplyCommand,
		numSuccess,
		len(pullStatus.Projects),
	); err != nil {
		ctx.Log.Warn("unable to update commit status: %s", err)
	}
}

func (a *ApplyCommandRunner) anyFailedPolicyChecks(pull models.PullRequest) bool {
	policyCheckPullStatus, _ := a.DB.GetPullStatus(pull)
	if policyCheckPullStatus != nil && policyCheckPullStatus.StatusCount(models.ErroredPolicyCheckStatus) > 0 {
		return true
	}

	return false

}

// applyAllDisabledComment is posted when apply all commands (i.e. "atlantis apply")
// are disabled and an apply all command is issued.
var applyAllDisabledComment = "**Error:** Running `atlantis apply` without flags is disabled." +
	" You must specify which project to apply via the `-d <dir>`, `-w <workspace>` or `-p <project name>` flags."

// applyDisabledComment is posted when apply commands are disabled globally and an apply command is issued.
var applyDisabledComment = "**Error:** Running `atlantis apply` is disabled."
