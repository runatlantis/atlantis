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
func (w *DefaultPostWorkflowHooksCommandRunner) RunPostHooks(
	ctx *command.Context, cmd *CommentCommand,
) error {
	pull := ctx.Pull
	baseRepo := pull.BaseRepo
	headRepo := ctx.HeadRepo
	user := ctx.User
	log := ctx.Log

	postWorkflowHooks := make([]*valid.WorkflowHook, 0)
	for _, repo := range w.GlobalCfg.Repos {
		if repo.IDMatches(baseRepo.ID()) && repo.BranchMatches(pull.BaseBranch) && len(repo.PostWorkflowHooks) > 0 {
			postWorkflowHooks = append(postWorkflowHooks, repo.PostWorkflowHooks...)
		}
	}

	// short circuit any other calls if there are no post-hooks configured
	if len(postWorkflowHooks) == 0 {
		return nil
	}

	log.Debug("post-hooks configured, running...")

	unlockFn, err := w.WorkingDirLocker.TryLock(baseRepo.FullName, pull.Num, DefaultWorkspace, DefaultRepoRelDir)
	if err != nil {
		return err
	}
	log.Debug("got workspace lock")
	defer unlockFn()

	repoDir, _, err := w.WorkingDir.Clone(headRepo, pull, DefaultWorkspace)
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
			CommandName:        cmd.Name.String(),
		},
		postWorkflowHooks, repoDir)

	if err != nil {
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
		hookDescription := hook.StepDescription
		if hookDescription == "" {
			hookDescription = fmt.Sprintf("Post workflow hook #%d", i)
		}

		ctx.HookID = uuid.NewString()
		shell := hook.Shell
		if shell == "" {
			ctx.Log.Debug("Setting shell to default: %q", shell)
			shell = "sh"
		}
		shellArgs := hook.ShellArgs
		if shellArgs == "" {
			ctx.Log.Debug("Setting shellArgs to default: %q", shellArgs)
			shellArgs = "-c"
		}
		url, err := w.Router.GenerateProjectWorkflowHookURL(ctx.HookID)
		if err != nil {
			return err
		}

		if err := w.CommitStatusUpdater.UpdatePostWorkflowHook(ctx.Pull, models.PendingCommitStatus, hookDescription, "", url); err != nil {
			ctx.Log.Warn("unable to update post workflow hook status: %s", err)
		}

		_, runtimeDesc, err := w.PostWorkflowHookRunner.Run(ctx, hook.RunCommand, shell, shellArgs, repoDir)

		if err != nil {
			if err := w.CommitStatusUpdater.UpdatePostWorkflowHook(ctx.Pull, models.FailedCommitStatus, hookDescription, runtimeDesc, url); err != nil {
				ctx.Log.Warn("unable to update post workflow hook status: %s", err)
			}
			return err
		}

		if err := w.CommitStatusUpdater.UpdatePostWorkflowHook(ctx.Pull, models.SuccessCommitStatus, hookDescription, runtimeDesc, url); err != nil {
			ctx.Log.Warn("unable to update post workflow hook status: %s", err)
		}
	}
	return nil
}
