package gateway

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/metrics"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/lyft/feature"
	"github.com/runatlantis/atlantis/server/recovery"
	"github.com/uber-go/tally/v4"
)

const PlatformModeApplyStatusMessage = "Bypassed for platform mode"

// AutoplanValidator handles setting up repo cloning and checking to verify of any terraform files have changed
type AutoplanValidator struct {
	Scope                         tally.Scope
	VCSClient                     vcs.Client
	PreWorkflowHooksCommandRunner events.PreWorkflowHooksCommandRunner
	Drainer                       *events.Drainer
	GlobalCfg                     valid.GlobalCfg
	VCSStatusUpdater              events.VCSStatusUpdater
	PrjCmdBuilder                 events.ProjectPlanCommandBuilder
	OutputUpdater                 events.OutputUpdater
	WorkingDir                    events.WorkingDir
	WorkingDirLocker              events.WorkingDirLocker
	Allocator                     feature.Allocator
}

const DefaultWorkspace = "default"

func (r *AutoplanValidator) isValid(ctx context.Context, logger logging.Logger, baseRepo models.Repo, headRepo models.Repo, pull models.PullRequest, user models.User) (bool, error) {
	if opStarted := r.Drainer.StartOp(); !opStarted {
		return false, errors.New("atlantis is shutting down, cannot process current event")
	}
	defer r.Drainer.OpDone()
	defer r.logPanics(ctx, logger)

	cmdCtx := &command.Context{
		User:       user,
		Log:        logger,
		Scope:      r.Scope,
		Pull:       pull,
		HeadRepo:   headRepo,
		Trigger:    command.AutoTrigger,
		RequestCtx: ctx,
	}
	if !r.validateCtxAndComment(cmdCtx) {
		return false, errors.New("invalid command context")
	}

	// Update status and fail the req when preworkflow hook fails since this step is critical in determining if this req needs to be forwarded to the worker
	err := r.PreWorkflowHooksCommandRunner.RunPreHooks(ctx, cmdCtx)
	if err != nil {
		if _, statusErr := r.VCSStatusUpdater.UpdateCombined(ctx, cmdCtx.HeadRepo, cmdCtx.Pull, models.FailedVCSStatus, command.Plan, "", err.Error()); statusErr != nil {
			cmdCtx.Log.WarnContext(cmdCtx.RequestCtx, fmt.Sprintf("unable to update commit status: %v", statusErr))
		}
		return false, errors.Wrap(err, "running preworkflow hook")
	}

	projectCmds, err := r.PrjCmdBuilder.BuildAutoplanCommands(cmdCtx)
	if err != nil {
		if _, statusErr := r.VCSStatusUpdater.UpdateCombined(ctx, baseRepo, pull, models.FailedVCSStatus, command.Plan, "", ""); statusErr != nil {
			cmdCtx.Log.WarnContext(cmdCtx.RequestCtx, fmt.Sprintf("unable to update commit status: %v", statusErr))
		}
		// If error happened after clone was made, we should clean it up here too
		unlockFn, lockErr := r.WorkingDirLocker.TryLock(baseRepo.FullName, pull.Num, DefaultWorkspace)
		if lockErr != nil {
			cmdCtx.Log.WarnContext(cmdCtx.RequestCtx, "workspace was locked")
			return false, errors.Wrap(err, lockErr.Error())
		}
		defer unlockFn()
		if cloneErr := r.WorkingDir.Delete(baseRepo, pull); cloneErr != nil {
			cmdCtx.Log.WarnContext(cmdCtx.RequestCtx, "unable to delete clone after autoplan failed", map[string]interface{}{"err": cloneErr})
		}
		r.OutputUpdater.UpdateOutput(cmdCtx, events.AutoplanCommand{}, command.Result{Error: err})
		return false, errors.Wrap(err, "Failed building autoplan commands")
	}
	unlockFn, err := r.WorkingDirLocker.TryLock(baseRepo.FullName, pull.Num, DefaultWorkspace)
	if err != nil {
		cmdCtx.Log.WarnContext(cmdCtx.RequestCtx, "workspace was locked")
		return false, err
	}
	defer unlockFn()
	// Delete repo clone generated to validate plan
	if err := r.WorkingDir.Delete(baseRepo, pull); err != nil {
		return false, errors.Wrap(err, "Failed deleting cloned repo")
	}
	if len(projectCmds) == 0 {
		cmdCtx.Log.InfoContext(cmdCtx.RequestCtx, "no modified projects have been found")
		for _, cmd := range []command.Name{command.Plan, command.Apply, command.PolicyCheck} {
			if _, err := r.VCSStatusUpdater.UpdateCombinedCount(ctx, baseRepo, pull, models.SuccessVCSStatus, cmd, 0, 0, ""); err != nil {
				cmdCtx.Log.WarnContext(cmdCtx.RequestCtx, fmt.Sprintf("unable to update commit status: %s", err))
			}
		}
		return false, nil
	}

	// WorkflowModeType can be configured per project root, so we need to ensure all roots are in platform mode
	// before we set atlantis/apply status to success and allow PR to merge into the default branch
	// TODO: Remove this after we remove the required atlantis/apply status check
	if allProjectsInPlatformMode(projectCmds) {
		if err := r.updateAtlantisApplyChecks(cmdCtx, baseRepo, projectCmds); err != nil {
			cmdCtx.Log.ErrorContext(cmdCtx.RequestCtx, errors.Wrap(err, "updating atlantis apply status").Error())
		}
	}
	if _, err := r.VCSStatusUpdater.UpdateCombined(ctx, baseRepo, pull, models.QueuedVCSStatus, command.Plan, "", "Request received. Adding to the queue..."); err != nil {
		cmdCtx.Log.WarnContext(cmdCtx.RequestCtx, fmt.Sprintf("unable to update commit status: %s", err))
	}
	return true, nil
}

func allProjectsInPlatformMode(cmds []command.ProjectContext) bool {
	for _, cmd := range cmds {
		if cmd.WorkflowModeType != valid.PlatformWorkflowMode {
			return false
		}
	}
	return true
}

func (r *AutoplanValidator) InstrumentedIsValid(ctx context.Context, logger logging.Logger, baseRepo models.Repo, headRepo models.Repo, pull models.PullRequest, user models.User) bool {
	timer := r.Scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer timer.Stop()
	isValid, err := r.isValid(ctx, logger, baseRepo, headRepo, pull, user)

	if err != nil {
		logger.ErrorContext(ctx, err.Error())
		r.Scope.Counter(metrics.ExecutionErrorMetric).Inc(1)
		return false
	}
	if !isValid {
		r.Scope.Counter(metrics.ExecutionFailureMetric).Inc(1)
		return false
	}
	r.Scope.Counter(metrics.ExecutionSuccessMetric).Inc(1)
	return true
}

func (r *AutoplanValidator) logPanics(ctx context.Context, logger logging.Logger) {
	if err := recover(); err != nil {
		stack := recovery.Stack(3)
		logger.ErrorContext(ctx, fmt.Sprintf("PANIC: %s\n%s", err, stack))
	}
}

func (r *AutoplanValidator) validateCtxAndComment(cmdCtx *command.Context) bool {
	if cmdCtx.HeadRepo.Owner != cmdCtx.Pull.BaseRepo.Owner {
		cmdCtx.Log.InfoContext(cmdCtx.RequestCtx, "command was run on a fork pull request which is disallowed")
		return false
	}

	if cmdCtx.Pull.State != models.OpenPullState {
		cmdCtx.Log.InfoContext(cmdCtx.RequestCtx, "command was run on closed pull request")
		return false
	}

	repo := r.GlobalCfg.MatchingRepo(cmdCtx.Pull.BaseRepo.ID())
	if !repo.BranchMatches(cmdCtx.Pull.BaseBranch) {
		cmdCtx.Log.InfoContext(cmdCtx.RequestCtx, "command was run on a pull request which doesn't match base branches")
		// just ignore it to allow us to use any git workflows without malicious intentions.
		return false
	}
	return true
}

func (r *AutoplanValidator) updateAtlantisApplyChecks(cmdCtx *command.Context, repo models.Repo, prjCmds []command.ProjectContext) error {
	shouldAllocate, err := r.Allocator.ShouldAllocate(feature.PlatformMode, feature.FeatureContext{
		RepoName: repo.FullName,
	})
	if err != nil {
		return errors.Wrap(err, "allocating platform mode")
	}

	if !shouldAllocate {
		return nil
	}

	if _, statusErr := r.VCSStatusUpdater.UpdateCombined(cmdCtx.RequestCtx, cmdCtx.HeadRepo, cmdCtx.Pull, models.SuccessVCSStatus, command.Apply, "", PlatformModeApplyStatusMessage); statusErr != nil {
		return errors.Wrap(statusErr, "updating atlantis apply to success")
	}

	return nil
}
