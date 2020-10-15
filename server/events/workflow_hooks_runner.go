package events

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/runatlantis/atlantis/server/events/db"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/recovery"
)

type WorkflowHooksCommandRunner interface {
	RunPreHooks(baseRepo models.Repo, headRepo models.Repo, pull models.PullRequest, user models.User) error
}

// DefaultWorkflowHooksCommandRunner is the first step when processing a workflow hook commands.
type DefaultWorkflowHooksCommandRunner struct {
	VCSClient                vcs.Client
	GithubPullGetter         GithubPullGetter
	AzureDevopsPullGetter    AzureDevopsPullGetter
	GitlabMergeRequestGetter GitlabMergeRequestGetter
	Logger                   *logging.SimpleLogger
	WorkingDirLocker         WorkingDirLocker
	WorkingDir               WorkingDir
	GlobalCfg                valid.GlobalCfg
	DB                       *db.BoltDB
	Drainer                  *Drainer
	DeleteLockCommand        DeleteLockCommand
}

func (w *DefaultWorkflowHooksCommandRunner) RunPreHooks(
	baseRepo models.Repo,
	headRepo models.Repo,
	pull models.PullRequest,
	user models.User,
) (*WorkflowHookCommandResult, error) {
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
	return &WorkflowHookCommandResult{
		WorkflowHookResults: outputs,
	}, nil
}

func (w *DefaultWorkflowHooksCommandRunner) runHooks(
	ctx models.WorkflowHookCommandContext,
	workflowHooks []*valid.WorkflowHook,
	repoDir string,
) (outputs []models.WorkflowHookResult) {
	for _, hook := range workflowHooks {
		out, err := w.runCmd(ctx, hook.RunCommand, repoDir)

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

func (w *DefaultWorkflowHooksCommandRunner) runCmd(ctx models.WorkflowHookCommandContext, command string, path string) (string, error) {
	cmd := exec.Command("sh", "-c", command) // #nosec
	cmd.Dir = path

	baseEnvVars := os.Environ()
	customEnvVars := map[string]string{
		"BASE_BRANCH_NAME": ctx.Pull.BaseBranch,
		"BASE_REPO_NAME":   ctx.BaseRepo.Name,
		"BASE_REPO_OWNER":  ctx.BaseRepo.Owner,
		"DIR":              path,
		"HEAD_BRANCH_NAME": ctx.Pull.HeadBranch,
		"HEAD_REPO_NAME":   ctx.HeadRepo.Name,
		"HEAD_REPO_OWNER":  ctx.HeadRepo.Owner,
		"PULL_AUTHOR":      ctx.Pull.Author,
		"PULL_NUM":         fmt.Sprintf("%d", ctx.Pull.Num),
		"USER_NAME":        ctx.User.Username,
	}

	finalEnvVars := baseEnvVars
	for key, val := range customEnvVars {
		finalEnvVars = append(finalEnvVars, fmt.Sprintf("%s=%s", key, val))
	}

	cmd.Env = finalEnvVars
	out, err := cmd.CombinedOutput()

	if err != nil {
		err = fmt.Errorf("%s: running %q in %q: \n%s", err, command, path, out)
		ctx.Log.Debug("error: %s", err)
		return "", err
	}
	ctx.Log.Info("successfully ran %q in %q", command, path)
	return string(out), nil
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
