package runners

import (
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/models"
)

func WithLocking(
	prjCmdRunner ProjectCommandRunner,
	locker events.ProjectLocker,
	lockUrlGenerator LockURLGenerator,
	workginDirLocker events.WorkingDirLocker,
) ProjectCommandRunner {
	return &LockingProjectCommandRunner{
		ProjectCommandRunner: prjCmdRunner,
		Locker:               locker,
		LockURLGenerator:     lockUrlGenerator,
		WorkingDirLocker:     workginDirLocker,
	}
}

// LockingProjectCommandRunner implements project locks.
type LockingProjectCommandRunner struct {
	ProjectCommandRunner
	Locker           events.ProjectLocker
	LockURLGenerator LockURLGenerator
	WorkingDirLocker events.WorkingDirLocker
}

func (p *LockingProjectCommandRunner) Plan(ctx models.ProjectCommandContext) models.ProjectResult {
	return p.runWithLocks(ctx, p.ProjectCommandRunner.Plan)
}

func (p *LockingProjectCommandRunner) PolicyCheck(ctx models.ProjectCommandContext) models.ProjectResult {
	// Acquire Atlantis lock for this repo/dir/workspace.
	// This should already be acquired from the prior plan operation.
	// if for some reason an unlock happens between the plan and policy check step
	// we will attempt to capture the lock here but fail to get the working directory
	// at which point we will unlock again to preserve functionality
	// If we fail to capture the lock here (super unlikely) then we error out and the user is forced to replan
	return p.runWithLocks(ctx, p.ProjectCommandRunner.PolicyCheck)
}

func (p *LockingProjectCommandRunner) Apply(ctx models.ProjectCommandContext) (result models.ProjectResult) {
	result = models.ProjectResult{
		Command:     models.ApplyCommand,
		RepoRelDir:  ctx.RepoRelDir,
		Workspace:   ctx.Workspace,
		ProjectName: ctx.ProjectName,
	}

	// Acquire internal lock for the directory we're going to operate in.
	unlockFn, err := p.WorkingDirLocker.TryLock(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, ctx.Workspace)
	if err != nil {
		result.Error = err
		return
	}
	defer unlockFn()

	return p.ProjectCommandRunner.Apply(ctx)
}

func (p *LockingProjectCommandRunner) runWithLocks(
	ctx models.ProjectCommandContext,
	execute func(ctx models.ProjectCommandContext) models.ProjectResult,
) (result models.ProjectResult) {
	result = models.ProjectResult{
		Command:     models.PlanCommand,
		RepoRelDir:  ctx.RepoRelDir,
		Workspace:   ctx.Workspace,
		ProjectName: ctx.ProjectName,
	}

	// Acquire Atlantis lock for this repo/dir/workspace.
	lockAttempt, err := p.Locker.TryLock(ctx.Log, ctx.Pull, ctx.User, ctx.Workspace, models.NewProject(ctx.Pull.BaseRepo.FullName, ctx.RepoRelDir))
	if err != nil {
		result.Error = errors.Wrap(err, "acquiring lock")
		return
	}

	if !lockAttempt.LockAcquired {
		result.Failure = lockAttempt.LockFailureReason
		return
	}
	ctx.Log.Debug("acquired lock for project")

	// Acquire internal lock for the directory we're going to operate in.
	unlockFn, err := p.WorkingDirLocker.TryLock(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, ctx.Workspace)
	if err != nil {
		result.Error = err
		return
	}
	defer unlockFn()

	result = execute(ctx)
	if result.Error != nil {
		if unlockErr := lockAttempt.UnlockFn(); unlockErr != nil {
			ctx.Log.Err("error unlocking state after plan error: %v", unlockErr)
		}
		return
	}

	lockUrl := p.LockURLGenerator.GenerateLockURL(lockAttempt.LockKey)

	if result.PlanSuccess != nil {
		result.PlanSuccess.LockURL = lockUrl
	}

	if result.PolicyCheckSuccess != nil {
		result.PolicyCheckSuccess.LockURL = lockUrl
	}

	return
}
