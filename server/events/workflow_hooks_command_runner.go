package events

import (
	"fmt"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/runtime"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/recovery"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_workflows_hooks_command_runner.go WorkflowHooksCommandRunner

type WorkflowHooksCommandRunner interface {
	RunPreHooks(baseRepo models.Repo, headRepo models.Repo, pull models.PullRequest, user models.User) (*WorkflowHooksCommandResult, error)
}

// DefaultWorkflowHooksCommandRunner is the first step when processing a workflow hook commands.
type DefaultWorkflowHooksCommandRunner struct {
	VCSClient          vcs.Client
	Logger             logging.SimpleLogging
	WorkingDirLocker   WorkingDirLocker
	WorkingDir         WorkingDir
	GlobalCfg          valid.GlobalCfg
	Drainer            *Drainer
	WorkflowHookRunner runtime.WorkflowHookRunner
}

func (w *DefaultWorkflowHooksCommandRunner) RunPreHooks(
	baseRepo models.Repo,
	headRepo models.Repo,
	pull models.PullRequest,
	user models.User,
) (*WorkflowHooksCommandResult, error) {
	if opStarted := w.Drainer.StartOp(); !opStarted {
		if commentErr := w.VCSClient.CreateComment(baseRepo, pull.Num, ShutdownComment, models.WorkflowHooksCommand.String()); commentErr != nil {
			w.Logger.Log(logging.Error, "unable to comment that Atlantis is shutting down: %s", commentErr)
		}
		return nil, nil
	}
	defer w.Drainer.OpDone()

	log := w.buildLogger(baseRepo.FullName, pull.Num)
	defer w.logPanics(baseRepo, pull.Num, log)

	log.Info("Running Pre Hooks for repo: ")

	unlockFn, err := w.WorkingDirLocker.TryLock(baseRepo.FullName, pull.Num, DefaultWorkspace)
	if err != nil {
		log.Warn("workspace was locked")
		return nil, err
	}
	log.Debug("got workspace lock")
	defer unlockFn()

	repoDir, _, err := w.WorkingDir.Clone(log, baseRepo, headRepo, pull, DefaultWorkspace)
	if err != nil {
		return nil, err
	}

	workflowHooks := make([]*valid.WorkflowHook, 0)
	for _, repo := range w.GlobalCfg.Repos {
		if repo.IDMatches(baseRepo.ID()) && len(repo.WorkflowHooks) > 0 {
			workflowHooks = append(workflowHooks, repo.WorkflowHooks...)
		}
	}

	ctx := models.WorkflowHookCommandContext{
		BaseRepo: baseRepo,
		HeadRepo: headRepo,
		Log:      log,
		Pull:     pull,
		User:     user,
		Verbose:  false,
	}

	outputs := w.runHooks(ctx, workflowHooks, repoDir)
	return &WorkflowHooksCommandResult{
		WorkflowHookResults: outputs,
	}, nil
}

func (w *DefaultWorkflowHooksCommandRunner) runHooks(
	ctx models.WorkflowHookCommandContext,
	workflowHooks []*valid.WorkflowHook,
	repoDir string,
) (outputs []models.WorkflowHookResult) {
	for _, hook := range workflowHooks {
		out, err := w.WorkflowHookRunner.Run(ctx, hook.RunCommand, repoDir)

		res := models.WorkflowHookResult{
			Command: models.WorkflowHooksCommand,
			Output:  out,
		}

		if err != nil {
			res.Error = err
			res.Success = false
		} else {
			res.Success = true
		}

		outputs = append(outputs, res)

		if !res.IsSuccessful() {
			return
		}
	}

	return
}

func (w *DefaultWorkflowHooksCommandRunner) buildLogger(repoFullName string, pullNum int) *logging.SimpleLogger {
	src := fmt.Sprintf("%s#%d", repoFullName, pullNum)
	return w.Logger.NewLogger(src, true, w.Logger.GetLevel())
}

// logPanics logs and creates a comment on the pull request for panics.
func (w *DefaultWorkflowHooksCommandRunner) logPanics(baseRepo models.Repo, pullNum int, logger logging.SimpleLogging) {
	if err := recover(); err != nil {
		stack := recovery.Stack(3)
		logger.Err("PANIC: %s\n%s", err, stack)
		if commentErr := w.VCSClient.CreateComment(
			baseRepo,
			pullNum,
			fmt.Sprintf("**Error: goroutine panic. This is a bug.**\n```\n%s\n%s```", err, stack),
			"",
		); commentErr != nil {
			logger.Err("unable to comment: %s", commentErr)
		}
	}
}
