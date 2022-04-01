package events

import (
	"context"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_pre_workflows_hooks_command_runner.go PreWorkflowHooksCommandRunner

type PreWorkflowHooksCommandRunner interface {
	RunPreHooks(ctx context.Context, cmdCtx *command.Context) error
}

// DefaultPreWorkflowHooksCommandRunner is the first step when processing a workflow hook commands.
type DefaultPreWorkflowHooksCommandRunner struct {
	VCSClient             vcs.Client
	WorkingDirLocker      WorkingDirLocker
	WorkingDir            WorkingDir
	GlobalCfg             valid.GlobalCfg
	PreWorkflowHookRunner runtime.PreWorkflowHookRunner
}

// RunPreHooks runs pre_workflow_hooks when PR is opened or updated.
func (w *DefaultPreWorkflowHooksCommandRunner) RunPreHooks(
	ctx context.Context,
	cmdCtx *command.Context,
) error {
	pull := cmdCtx.Pull
	baseRepo := pull.BaseRepo
	headRepo := cmdCtx.HeadRepo
	user := cmdCtx.User
	log := cmdCtx.Log

	preWorkflowHooks := make([]*valid.PreWorkflowHook, 0)
	for _, repo := range w.GlobalCfg.Repos {
		if repo.IDMatches(baseRepo.ID()) && len(repo.PreWorkflowHooks) > 0 {
			preWorkflowHooks = append(preWorkflowHooks, repo.PreWorkflowHooks...)
		}
	}

	// short circuit any other calls if there are no pre-hooks configured
	if len(preWorkflowHooks) == 0 {
		return nil
	}

	log.Debugf("pre-hooks configured, running...")

	unlockFn, err := w.WorkingDirLocker.TryLock(baseRepo.FullName, pull.Num, DefaultWorkspace)
	if err != nil {
		return errors.Wrap(err, "locking working dir")
	}
	log.Debugf("got workspace lock")
	defer unlockFn()

	repoDir, _, err := w.WorkingDir.Clone(log, headRepo, pull, DefaultWorkspace)
	if err != nil {
		return errors.Wrap(err, "cloning repository")
	}

	err = w.runHooks(
		models.PreWorkflowHookCommandContext{
			BaseRepo: baseRepo,
			HeadRepo: headRepo,
			Log:      log,
			Pull:     pull,
			User:     user,
			Verbose:  false,
		},
		preWorkflowHooks, repoDir)

	if err != nil {
		return errors.Wrap(err, "running pre workflow hooks")
	}

	return nil
}

func (w *DefaultPreWorkflowHooksCommandRunner) runHooks(
	cmdCtx models.PreWorkflowHookCommandContext,
	preWorkflowHooks []*valid.PreWorkflowHook,
	repoDir string,
) error {

	for _, hook := range preWorkflowHooks {
		_, err := w.PreWorkflowHookRunner.Run(cmdCtx, hook.RunCommand, repoDir)

		if err != nil {
			return err
		}
	}

	return nil
}
