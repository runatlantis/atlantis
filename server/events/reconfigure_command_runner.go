// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"fmt"

	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/jobs"
)

// ReconfigureCommandRunner deletes a pull request's local Atlantis workspace
// state before DefaultCommandRunner continues with a normal plan command.
type ReconfigureCommandRunner struct {
	WorkingDir               WorkingDir       `validate:"required"`
	WorkingDirLocker         WorkingDirLocker `validate:"required"`
	Locker                   locking.Locker   `validate:"required"`
	Database                 db.Database      `validate:"required"`
	PullUpdater              *PullUpdater     `validate:"required"`
	LogStreamResourceCleaner ResourceCleaner
	CancellationTracker      CancellationTracker
}

func NewReconfigureCommandRunner(
	workingDir WorkingDir,
	workingDirLocker WorkingDirLocker,
	locker locking.Locker,
	database db.Database,
	pullUpdater *PullUpdater,
	logStreamResourceCleaner ResourceCleaner,
	cancellationTracker CancellationTracker,
) *ReconfigureCommandRunner {
	return &ReconfigureCommandRunner{
		WorkingDir:               workingDir,
		WorkingDirLocker:         workingDirLocker,
		Locker:                   locker,
		Database:                 database,
		PullUpdater:              pullUpdater,
		LogStreamResourceCleaner: logStreamResourceCleaner,
		CancellationTracker:      cancellationTracker,
	}
}

func (r *ReconfigureCommandRunner) Run(ctx *command.Context, cmd *CommentCommand) {
	unlockFn, err := r.WorkingDirLocker.TryLock(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, "", "", "", command.Reconfigure)
	if err != nil {
		r.fail(ctx, cmd, err)
		return
	}
	defer unlockFn()

	pullStatus, err := r.Database.GetPullStatus(ctx.Pull)
	if err != nil {
		ctx.Log.Err("retrieving pull status: %s", err)
	}
	r.cleanupLogStreams(ctx.Pull, pullStatus)

	if err := r.WorkingDir.Delete(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull); err != nil {
		r.fail(ctx, cmd, fmt.Errorf("deleting pull request workspace: %w", err))
		return
	}

	if _, err := r.Locker.UnlockByPull(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num); err != nil {
		r.fail(ctx, cmd, fmt.Errorf("deleting locks: %w", err))
		return
	}

	if err := r.Database.DeletePullStatus(ctx.Pull); err != nil {
		r.fail(ctx, cmd, fmt.Errorf("deleting pull status from db: %w", err))
		return
	}

	if r.CancellationTracker != nil {
		r.CancellationTracker.Clear(ctx.Pull)
	}
	ctx.PullStatus = nil
}

func (r *ReconfigureCommandRunner) cleanupLogStreams(pull models.PullRequest, pullStatus *models.PullStatus) {
	if r.LogStreamResourceCleaner == nil || pullStatus == nil {
		return
	}

	for _, project := range pullStatus.Projects {
		r.LogStreamResourceCleaner.CleanUp(jobs.PullInfo{
			PullNum:      pull.Num,
			Repo:         pull.BaseRepo.Name,
			RepoFullName: pull.BaseRepo.FullName,
			ProjectName:  project.ProjectName,
			Path:         project.RepoRelDir,
			Workspace:    project.Workspace,
		})
	}
}

func (r *ReconfigureCommandRunner) fail(ctx *command.Context, cmd *CommentCommand, err error) {
	ctx.CommandHasErrors = true
	r.PullUpdater.updatePull(ctx, cmd, command.Result{Error: err})
}
