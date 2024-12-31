package events

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
)

//go:generate pegomock generate --package mocks -o mocks/mock_post_workflow_hook_url_generator.go PostWorkflowHookURLGenerator

// PostWorkflowHookURLGenerator generates urls to view the post workflow progress.
type PostWorkflowHookURLGenerator interface {
	GenerateProjectWorkflowHookURL(hookID string) (string, error)
}

//go:generate pegomock generate --package mocks -o mocks/mock_post_workflows_hooks_command_runner.go PostWorkflowHooksCommandRunner

type PostWorkflowHooksCommandRunner interface {
	RunPostHooks(ctx *command.Context, cmd *CommentCommand) error
}

// DefaultPostWorkflowHooksCommandRunner is the first step when processing a workflow hook commands.
type DefaultPostWorkflowHooksCommandRunner struct {
	VCSClient              vcs.Client
	WorkingDirLocker       WorkingDirLocker
	WorkingDir             WorkingDir
	GlobalCfg              valid.GlobalCfg
	PostWorkflowHookRunner runtime.PostWorkflowHookRunner
	CommitStatusUpdater    CommitStatusUpdater
	Router                 PostWorkflowHookURLGenerator
}

// RunPostHooks runs post_workflow_hooks after a plan/apply has completed
func (w *DefaultPostWorkflowHooksCommandRunner) RunPostHooks(ctx *command.Context, cmd *CommentCommand) error {
	postWorkflowHooks := make([]*valid.WorkflowHook, 0)
	for _, repo := range w.GlobalCfg.Repos {
		if repo.IDMatches(ctx.Pull.BaseRepo.ID()) && repo.BranchMatches(ctx.Pull.BaseBranch) && len(repo.PostWorkflowHooks) > 0 {
			postWorkflowHooks = append(postWorkflowHooks, repo.PostWorkflowHooks...)
		}
	}

	// short circuit any other calls if there are no post-hooks configured
	if len(postWorkflowHooks) == 0 {
		return nil
	}

	ctx.Log.Info("Post-workflow hooks configured, running...")

	unlockFn, err := w.WorkingDirLocker.TryLock(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, DefaultWorkspace, DefaultRepoRelDir)
	if err != nil {
		return err
	}
	ctx.Log.Debug("got workspace lock")
	defer unlockFn()

	repoDir, _, err := w.WorkingDir.Clone(ctx.Log, ctx.HeadRepo, ctx.Pull, DefaultWorkspace)
	if err != nil {
		return err
	}

	var escapedArgs []string
	if cmd != nil {
		escapedArgs = escapeArgs(cmd.Flags)
	}

	err = w.runHooks(
		models.WorkflowHookCommandContext{
			BaseRepo:           ctx.Pull.BaseRepo,
			HeadRepo:           ctx.HeadRepo,
			Log:                ctx.Log,
			Pull:               ctx.Pull,
			User:               ctx.User,
			Verbose:            false,
			EscapedCommentArgs: escapedArgs,
			CommandName:        cmd.Name.String(),
			API:                ctx.API,
		},
		postWorkflowHooks, repoDir)

	if err != nil {
		ctx.Log.Err("Error running post-workflow hooks %s.", err)
		return err
	}

	return nil
}

func (w *DefaultPostWorkflowHooksCommandRunner) runHooks(
	ctx models.WorkflowHookCommandContext,
	postWorkflowHooks []*valid.WorkflowHook,
	repoDir string,
) error {

	for i, hook := range postWorkflowHooks {
		ctx.HookDescription = hook.StepDescription
		if ctx.HookDescription == "" {
			ctx.HookDescription = fmt.Sprintf("Post workflow hook #%d", i)
		}

		ctx.HookStepName = fmt.Sprintf("post %s #%d", ctx.CommandName, i)

		ctx.Log.Debug("Processing post workflow hook '%s', Command '%s', Target commands [%s]",
			ctx.HookDescription, ctx.CommandName, hook.Commands)
		if hook.Commands != "" && !strings.Contains(hook.Commands, ctx.CommandName) {
			ctx.Log.Debug("Skipping post workflow hook '%s' as command '%s' is not in Commands [%s]",
				ctx.HookDescription, ctx.CommandName, hook.Commands)
			continue
		}

		ctx.Log.Debug("Running post workflow hook: '%s'", ctx.HookDescription)
		ctx.HookID = uuid.NewString()
		shell := hook.Shell
		if shell == "" {
			ctx.Log.Debug("Setting shell to default: '%s'", shell)
			shell = "sh"
		}
		shellArgs := hook.ShellArgs
		if shellArgs == "" {
			ctx.Log.Debug("Setting shellArgs to default: '%s'", shellArgs)
			shellArgs = "-c"
		}
		url, err := w.Router.GenerateProjectWorkflowHookURL(ctx.HookID)
		if err != nil && !ctx.API {
			return err
		}

		if err := w.CommitStatusUpdater.UpdatePostWorkflowHook(ctx.Log, ctx.Pull, models.PendingCommitStatus, ctx.HookDescription, "", url); err != nil {
			ctx.Log.Warn("unable to update post workflow hook status: %s", err)
		}

		_, runtimeDesc, err := w.PostWorkflowHookRunner.Run(ctx, hook.RunCommand, shell, shellArgs, repoDir)

		if err != nil {
			if err := w.CommitStatusUpdater.UpdatePostWorkflowHook(ctx.Log, ctx.Pull, models.FailedCommitStatus, ctx.HookDescription, runtimeDesc, url); err != nil {
				ctx.Log.Warn("unable to update post workflow hook status: %s", err)
			}
			return err
		}

		if err := w.CommitStatusUpdater.UpdatePostWorkflowHook(ctx.Log, ctx.Pull, models.SuccessCommitStatus, ctx.HookDescription, runtimeDesc, url); err != nil {
			ctx.Log.Warn("unable to update post workflow hook status: %s", err)
		}
	}

	ctx.Log.Info("Post-workflow hooks completed")

	return nil
}
