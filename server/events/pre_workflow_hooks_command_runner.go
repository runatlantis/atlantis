package events

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
)

//go:generate pegomock generate -m --package mocks -o mocks/mock_pre_workflow_hook_url_generator.go PreWorkflowHookURLGenerator

// PreWorkflowHookURLGenerator generates urls to view the pre workflow progress.
type PreWorkflowHookURLGenerator interface {
	GenerateProjectWorkflowHookURL(hookID string) (string, error)
}

//go:generate pegomock generate -m --package mocks -o mocks/mock_pre_workflows_hooks_command_runner.go PreWorkflowHooksCommandRunner

type PreWorkflowHooksCommandRunner interface {
	RunPreHooks(ctx *command.Context, cmd *CommentCommand) error
}

// DefaultPreWorkflowHooksCommandRunner is the first step when processing a workflow hook commands.
type DefaultPreWorkflowHooksCommandRunner struct {
	VCSClient             vcs.Client
	WorkingDirLocker      WorkingDirLocker
	WorkingDir            WorkingDir
	GlobalCfg             valid.GlobalCfg
	PreWorkflowHookRunner runtime.PreWorkflowHookRunner
	CommitStatusUpdater   CommitStatusUpdater
	Router                PreWorkflowHookURLGenerator
}

// RunPreHooks runs pre_workflow_hooks when PR is opened or updated.
func (w *DefaultPreWorkflowHooksCommandRunner) RunPreHooks(ctx *command.Context, cmd *CommentCommand) error {
	pull := ctx.Pull
	baseRepo := pull.BaseRepo
	headRepo := ctx.HeadRepo
	user := ctx.User
	log := ctx.Log

	preWorkflowHooks := make([]*valid.WorkflowHook, 0)
	for _, repo := range w.GlobalCfg.Repos {
		if repo.IDMatches(baseRepo.ID()) && len(repo.PreWorkflowHooks) > 0 {
			preWorkflowHooks = append(preWorkflowHooks, repo.PreWorkflowHooks...)
		}
	}

	// short circuit any other calls if there are no pre-hooks configured
	if len(preWorkflowHooks) == 0 {
		return nil
	}

	log.Debug("pre-hooks configured, running...")

	unlockFn, err := w.WorkingDirLocker.TryLock(baseRepo.FullName, pull.Num, DefaultWorkspace, DefaultRepoRelDir)
	if err != nil {
		return err
	}
	log.Debug("got workspace lock")
	defer unlockFn()

	repoDir, _, err := w.WorkingDir.Clone(log, headRepo, pull, DefaultWorkspace)
	if err != nil {
		return err
	}

	var escapedArgs []string
	if cmd != nil {
		escapedArgs = escapeArgs(cmd.Flags)
	}

	err = w.runHooks(
		models.WorkflowHookCommandContext{
			BaseRepo:           baseRepo,
			HeadRepo:           headRepo,
			Log:                log,
			Pull:               pull,
			User:               user,
			Verbose:            false,
			EscapedCommentArgs: escapedArgs,
			HookID:             uuid.NewString(),
		},
		preWorkflowHooks, repoDir)

	if err != nil {
		return err
	}

	return nil
}

func (w *DefaultPreWorkflowHooksCommandRunner) runHooks(
	ctx models.WorkflowHookCommandContext,
	preWorkflowHooks []*valid.WorkflowHook,
	repoDir string,
) error {
	for i, hook := range preWorkflowHooks {
		hookDescription := hook.StepDescription
		if hookDescription == "" {
			hookDescription = fmt.Sprintf("Pre workflow hook #%d", i)
		}

		url, err := w.Router.GenerateProjectWorkflowHookURL(ctx.HookID)
		if err != nil {
			return err
		}

		if err := w.CommitStatusUpdater.UpdatePreWorkflowHook(ctx.Pull, models.PendingCommitStatus, hookDescription, "", url); err != nil {
			ctx.Log.Warn("unable to pre workflow hook status: %s", err)
			return err
		}

		_, runtimeDesc, err := w.PreWorkflowHookRunner.Run(ctx, hook.RunCommand, repoDir)

		if err != nil {
			if err := w.CommitStatusUpdater.UpdatePreWorkflowHook(ctx.Pull, models.FailedCommitStatus, hookDescription, runtimeDesc, url); err != nil {
				ctx.Log.Warn("unable to pre workflow hook status: %s", err)
			}
			return err
		}

		if err := w.CommitStatusUpdater.UpdatePreWorkflowHook(ctx.Pull, models.SuccessCommitStatus, hookDescription, runtimeDesc, url); err != nil {
			ctx.Log.Warn("unable to pre workflow hook status: %s", err)
			return err
		}
	}

	return nil
}
