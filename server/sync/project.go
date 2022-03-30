package sync

import (
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

// ProjectSyncer implements project locks.
type ProjectSyncer struct {
	events.ProjectCommandRunner
	Locker           events.ProjectLocker
	LockURLGenerator events.LockURLGenerator
}

func (p *ProjectSyncer) Plan(ctx command.ProjectContext) command.ProjectResult {
	result, lockResponse := p.sync(ctx, command.Plan, p.ProjectCommandRunner.Plan)
	lockUrl := p.LockURLGenerator.GenerateLockURL(lockResponse.LockKey)

	if result.PlanSuccess != nil {
		result.PlanSuccess.LockURL = lockUrl
	}

	return result
}

func (p *ProjectSyncer) PolicyCheck(ctx command.ProjectContext) command.ProjectResult {
	// Acquire Atlantis lock for this repo/dir/workspace.
	// This should already be acquired from the prior plan operation.
	// if for some reason an unlock happens between the plan and policy check step
	// we will attempt to capture the lock here but fail to get the working directory
	// at which point we will unlock again to preserve functionality
	// If we fail to capture the lock here (super unlikely) then we error out and the user is forced to replan
	result, lockResponse := p.sync(ctx, command.PolicyCheck, p.ProjectCommandRunner.PolicyCheck)
	lockUrl := p.LockURLGenerator.GenerateLockURL(lockResponse.LockKey)

	if result.PolicyCheckSuccess != nil {
		result.PolicyCheckSuccess.LockURL = lockUrl
	}

	return result
}

func (p *ProjectSyncer) sync(
	ctx command.ProjectContext,
	cmdName command.Name,
	execute func(ctx command.ProjectContext) command.ProjectResult,
) (
	result command.ProjectResult,
	lockResponse *events.TryLockResponse,
) {
	result = command.ProjectResult{
		Command:     cmdName,
		RepoRelDir:  ctx.RepoRelDir,
		Workspace:   ctx.Workspace,
		ProjectName: ctx.ProjectName,
	}

	// Acquire Atlantis lock for this repo/dir/workspace.
	lockResponse, err := p.Locker.TryLock(ctx.Log, ctx.Pull, ctx.User, ctx.Workspace, models.NewProject(ctx.Pull.BaseRepo.FullName, ctx.RepoRelDir))
	if err != nil {
		result.Error = errors.Wrap(err, "acquiring lock")
		return
	}

	if !lockResponse.LockAcquired {
		result.Failure = lockResponse.LockFailureReason
		return
	}
	ctx.Log.Debugf("acquired lock for project")

	result = execute(ctx)
	if result.Error != nil {
		if unlockErr := lockResponse.UnlockFn(); unlockErr != nil {
			ctx.Log.Errorf("error unlocking state after %s error: %v", cmdName, unlockErr)
		}
		return
	}

	return
}
