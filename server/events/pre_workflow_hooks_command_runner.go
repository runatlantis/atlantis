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

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_pre_workflows_hooks_command_runner.go PreWorkflowHooksCommandRunner

type PreWorkflowHooksCommandRunner interface {
	RunPreHooks(
		baseRepo models.Repo,
		headRepo models.Repo,
		pull models.PullRequest,
		user models.User,
	)
}

// DefaultPreWorkflowHooksCommandRunner is the first step when processing a workflow hook commands.
type DefaultPreWorkflowHooksCommandRunner struct {
	VCSClient             vcs.Client
	Logger                logging.SimpleLogging
	WorkingDirLocker      WorkingDirLocker
	WorkingDir            WorkingDir
	GlobalCfg             valid.GlobalCfg
	Drainer               *Drainer
	PreWorkflowHookRunner *runtime.PreWorkflowHookRunner
}

// RunPreHooks runs pre_workflow_hooks when PR is opened or updated.
func (w *DefaultPreWorkflowHooksCommandRunner) RunPreHooks(
	baseRepo models.Repo,
	headRepo models.Repo,
	pull models.PullRequest,
	user models.User,
) {
	if opStarted := w.Drainer.StartOp(); !opStarted {
		if commentErr := w.VCSClient.CreateComment(baseRepo, pull.Num, ShutdownComment, "pre_workflow_hooks"); commentErr != nil {
			w.Logger.Log(logging.Error, "unable to comment that Atlantis is shutting down: %s", commentErr)
		}
		return
	}
	defer w.Drainer.OpDone()

	log := w.buildLogger(baseRepo.FullName, pull.Num)
	defer w.logPanics(baseRepo, pull.Num, log)

	log.Info("running pre hooks")

	unlockFn, err := w.WorkingDirLocker.TryLock(baseRepo.FullName, pull.Num, DefaultWorkspace)
	if err != nil {
		log.Warn("workspace is locked")
		return
	}
	log.Debug("got workspace lock")
	defer unlockFn()

	repoDir, _, err := w.WorkingDir.Clone(log, headRepo, pull, DefaultWorkspace)
	if err != nil {
		log.Err("unable to run pre workflow hooks: %s", err)
		return
	}

	preWorkflowHooks := make([]*valid.PreWorkflowHook, 0)
	for _, repo := range w.GlobalCfg.Repos {
		if repo.IDMatches(baseRepo.ID()) && len(repo.PreWorkflowHooks) > 0 {
			preWorkflowHooks = append(preWorkflowHooks, repo.PreWorkflowHooks...)
		}
	}

	ctx := models.PreWorkflowHookCommandContext{
		BaseRepo: baseRepo,
		HeadRepo: headRepo,
		Log:      log,
		Pull:     pull,
		User:     user,
		Verbose:  false,
	}

	err = w.runHooks(ctx, preWorkflowHooks, repoDir)

	if err != nil {
		log.Err("pre workflow hook run error results: %s", err)
	}
}

func (w *DefaultPreWorkflowHooksCommandRunner) runHooks(
	ctx models.PreWorkflowHookCommandContext,
	preWorkflowHooks []*valid.PreWorkflowHook,
	repoDir string,
) error {

	for _, hook := range preWorkflowHooks {
		_, err := w.PreWorkflowHookRunner.Run(ctx, hook.RunCommand, repoDir)

		if err != nil {
			return nil
		}
	}

	return nil
}

func (w *DefaultPreWorkflowHooksCommandRunner) buildLogger(repoFullName string, pullNum int) *logging.SimpleLogger {
	src := fmt.Sprintf("%s#%d", repoFullName, pullNum)
	return w.Logger.NewLogger(src, true, w.Logger.GetLevel())
}

// logPanics logs and creates a comment on the pull request for panics.
func (w *DefaultPreWorkflowHooksCommandRunner) logPanics(baseRepo models.Repo, pullNum int, logger logging.SimpleLogging) {
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
