// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"
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
	VCSClient              vcs.Client                     `validate:"required"`
	WorkingDirLocker       WorkingDirLocker               `validate:"required"`
	WorkingDir             WorkingDir                     `validate:"required"`
	GlobalCfg              valid.GlobalCfg                `validate:"required"`
	PostWorkflowHookRunner runtime.PostWorkflowHookRunner `validate:"required"`
	CommitStatusUpdater    CommitStatusUpdater            `validate:"required"`
	Router                 PostWorkflowHookURLGenerator   `validate:"required"`
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

	unlockFn, err := w.WorkingDirLocker.TryLock(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, DefaultWorkspace, DefaultRepoRelDir, "", cmd.Name)
	if err != nil {
		return err
	}
	ctx.Log.Debug("got workspace lock")
	defer unlockFn()

	// check if the MR is closed (merged). if so, we need to handle the case where
	// the source branch might have been deleted
	var repoDir string
	if ctx.Pull.State == models.ClosedPullState {
		ctx.Log.Info("mr is closed (merged), using base branch for post-workflow hooks")
		// for closed MRs, we'll clone the base repo and checkout the base branch
		// instead of trying to merge the head branch which might be deleted
		repoDir, err = w.cloneForClosedMR(ctx)
	} else {
		repoDir, err = w.WorkingDir.Clone(ctx.Log, ctx.HeadRepo, ctx.Pull, DefaultWorkspace)
	}

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
			CommandHasErrors:   ctx.CommandHasErrors,
			API:                ctx.API,
		},
		postWorkflowHooks, repoDir)

	if err != nil {
		ctx.Log.Err("running post-workflow hooks: %s", err)
		return err
	}

	return nil
}

// cloneForClosedMR clones the repository for a closed MR without trying to fetch the deleted head branch
func (w *DefaultPostWorkflowHooksCommandRunner) cloneForClosedMR(ctx *command.Context) (string, error) {
	// for closed MRs, we'll use a simpler approach: clone the base repo directly
	// and checkout the base branch. this avoids the merge strategy that tries to fetch the deleted head branch

	// first, try to get the existing working directory if it exists
	existingDir, err := w.WorkingDir.GetWorkingDir(ctx.Pull.BaseRepo, ctx.Pull, DefaultWorkspace)
	if err == nil {
		// if the directory exists, we can use it
		ctx.Log.Info("Using existing working directory for closed MR %q", existingDir)
		return existingDir, nil
	}

	// if the directory doesn't exist, we need to create it
	// we'll use a temporary approach: create a minimal clone of the base repo
	ctx.Log.Info("Creating new working directory for closed MR")

	// get the pull directory path
	pullDir, err := w.WorkingDir.GetPullDir(ctx.Pull.BaseRepo, ctx.Pull)
	if err != nil {
		// if GetPullDir fails, we'll create the directory structure manually
		ctx.Log.Warn("Could not get pull directory, creating manually")
		// this is a fallback - we'll create a temporary directory
		tempDir, err := os.MkdirTemp("", "atlantis-closed-mr-*")
		if err != nil {
			return "", errors.Wrap(err, "creating temporary directory for closed MR")
		}
		cloneDir := filepath.Join(tempDir, DefaultWorkspace)

		// clone the base repo with the base branch
		baseCloneURL := ctx.Pull.BaseRepo.CloneURL
		cmd := exec.Command("git", "clone", "--depth=1", "--branch", ctx.Pull.BaseBranch, "--single-branch", baseCloneURL, cloneDir) // #nosec

		_, err = cmd.CombinedOutput()
		if err != nil {
			return "", errors.Wrap(err, "Running git clone for closed MR")
		}

		ctx.Log.Info("Successfully cloned base repo for closed MR in temp directory")
		return cloneDir, nil
	}

	// if we can get the pull directory, create the workspace directory
	cloneDir := filepath.Join(pullDir, DefaultWorkspace)
	if err := os.MkdirAll(cloneDir, 0700); err != nil {
		return "", errors.Wrap(err, "creating workspace directory for closed MR")
	}

	// clone the base repo with the base branch
	baseCloneURL := ctx.Pull.BaseRepo.CloneURL
	cmd := exec.Command("git", "clone", "--depth=1", "--branch", ctx.Pull.BaseBranch, "--single-branch", baseCloneURL, cloneDir) // #nosec

	_, err = cmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrap(err, "Running git clone for closed MR")
	}

	ctx.Log.Info("Successfully cloned base repo for closed MR")
	return cloneDir, nil
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
