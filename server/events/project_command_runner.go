// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

package events

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/runtime"
	"github.com/runatlantis/atlantis/server/events/webhooks"
	"github.com/runatlantis/atlantis/server/events/yaml/raw"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
	"github.com/runatlantis/atlantis/server/logging"
)

// DirNotExistErr is an error caused by the directory not existing.
type DirNotExistErr struct {
	RepoRelDir string
}

// Error implements the error interface.
func (d DirNotExistErr) Error() string {
	return fmt.Sprintf("dir %q does not exist", d.RepoRelDir)
}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_lock_url_generator.go LockURLGenerator

// LockURLGenerator generates urls to locks.
type LockURLGenerator interface {
	// GenerateLockURL returns the full URL to the lock at lockID.
	GenerateLockURL(lockID string) string
}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_step_runner.go StepRunner

// StepRunner runs steps. Steps are individual pieces of execution like
// `terraform plan`.
type StepRunner interface {
	// Run runs the step.
	Run(ctx models.ProjectCommandContext, extraArgs []string, path string, envs map[string]string) (string, error)
}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_custom_step_runner.go CustomStepRunner

// CustomStepRunner runs custom run steps.
type CustomStepRunner interface {
	// Run cmd in path.
	Run(ctx models.ProjectCommandContext, cmd string, path string, envs map[string]string) (string, error)
}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_env_step_runner.go EnvStepRunner

// EnvStepRunner runs env steps.
type EnvStepRunner interface {
	Run(ctx models.ProjectCommandContext, cmd string, value string, path string, envs map[string]string) (string, error)
}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_webhooks_sender.go WebhooksSender

// WebhooksSender sends webhook.
type WebhooksSender interface {
	// Send sends the webhook.
	Send(log *logging.SimpleLogger, res webhooks.ApplyResult) error
}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_project_command_runner.go ProjectCommandRunner

// ProjectCommandRunner runs project commands. A project command is a command
// for a specific TF project.
type ProjectCommandRunner interface {
	// Plan runs terraform plan for the project described by ctx.
	Plan(ctx models.ProjectCommandContext) models.ProjectResult
	// Apply runs terraform apply for the project described by ctx.
	Apply(ctx models.ProjectCommandContext) models.ProjectResult
}

// DefaultProjectCommandRunner implements ProjectCommandRunner.
type DefaultProjectCommandRunner struct {
	Locker              ProjectLocker
	LockURLGenerator    LockURLGenerator
	InitStepRunner      StepRunner
	PlanStepRunner      StepRunner
	ApplyStepRunner     StepRunner
	RunStepRunner       CustomStepRunner
	EnvStepRunner       EnvStepRunner
	PullApprovedChecker runtime.PullApprovedChecker
	WorkingDir          WorkingDir
	Webhooks            WebhooksSender
	WorkingDirLocker    WorkingDirLocker
}

// Plan runs terraform plan for the project described by ctx.
func (p *DefaultProjectCommandRunner) Plan(ctx models.ProjectCommandContext) models.ProjectResult {
	planSuccess, failure, err := p.doPlan(ctx)
	return models.ProjectResult{
		Command:     models.PlanCommand,
		PlanSuccess: planSuccess,
		Error:       err,
		Failure:     failure,
		RepoRelDir:  ctx.RepoRelDir,
		Workspace:   ctx.Workspace,
		ProjectName: ctx.ProjectName,
	}
}

// Apply runs terraform apply for the project described by ctx.
func (p *DefaultProjectCommandRunner) Apply(ctx models.ProjectCommandContext) models.ProjectResult {
	applyOut, failure, err := p.doApply(ctx)
	return models.ProjectResult{
		Command:      models.ApplyCommand,
		Failure:      failure,
		Error:        err,
		ApplySuccess: applyOut,
		RepoRelDir:   ctx.RepoRelDir,
		Workspace:    ctx.Workspace,
		ProjectName:  ctx.ProjectName,
	}
}

func (p *DefaultProjectCommandRunner) doPlan(ctx models.ProjectCommandContext) (*models.PlanSuccess, string, error) {
	// Acquire Atlantis lock for this repo/dir/workspace.
	lockAttempt, err := p.Locker.TryLock(ctx.Log, ctx.Pull, ctx.User, ctx.Workspace, models.NewProject(ctx.Pull.BaseRepo.FullName, ctx.RepoRelDir))
	if err != nil {
		return nil, "", errors.Wrap(err, "acquiring lock")
	}
	if !lockAttempt.LockAcquired {
		return nil, lockAttempt.LockFailureReason, nil
	}
	ctx.Log.Debug("acquired lock for project")

	// Acquire internal lock for the directory we're going to operate in.
	unlockFn, err := p.WorkingDirLocker.TryLock(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, ctx.Workspace)
	if err != nil {
		return nil, "", err
	}
	defer unlockFn()

	// Clone is idempotent so okay to run even if the repo was already cloned.
	repoDir, hasDiverged, cloneErr := p.WorkingDir.Clone(ctx.Log, ctx.HeadRepo, ctx.Pull, ctx.Workspace)
	if cloneErr != nil {
		if unlockErr := lockAttempt.UnlockFn(); unlockErr != nil {
			ctx.Log.Err("error unlocking state after plan error: %v", unlockErr)
		}
		return nil, "", cloneErr
	}
	projAbsPath := filepath.Join(repoDir, ctx.RepoRelDir)
	if _, err = os.Stat(projAbsPath); os.IsNotExist(err) {
		return nil, "", DirNotExistErr{RepoRelDir: ctx.RepoRelDir}
	}

	outputs, err := p.runSteps(ctx.Steps, ctx, projAbsPath)
	if err != nil {
		if unlockErr := lockAttempt.UnlockFn(); unlockErr != nil {
			ctx.Log.Err("error unlocking state after plan error: %v", unlockErr)
		}
		return nil, "", fmt.Errorf("%s\n%s", err, strings.Join(outputs, "\n"))
	}

	return &models.PlanSuccess{
		LockURL:         p.LockURLGenerator.GenerateLockURL(lockAttempt.LockKey),
		TerraformOutput: strings.Join(outputs, "\n"),
		RePlanCmd:       ctx.RePlanCmd,
		ApplyCmd:        ctx.ApplyCmd,
		HasDiverged:     hasDiverged,
	}, "", nil
}

func (p *DefaultProjectCommandRunner) runSteps(steps []valid.Step, ctx models.ProjectCommandContext, absPath string) ([]string, error) {
	var outputs []string
	envs := make(map[string]string)
	for _, step := range steps {
		var out string
		var err error
		switch step.StepName {
		case "init":
			out, err = p.InitStepRunner.Run(ctx, step.ExtraArgs, absPath, envs)
		case "plan":
			out, err = p.PlanStepRunner.Run(ctx, step.ExtraArgs, absPath, envs)
		case "apply":
			out, err = p.ApplyStepRunner.Run(ctx, step.ExtraArgs, absPath, envs)
		case "run":
			out, err = p.RunStepRunner.Run(ctx, step.RunCommand, absPath, envs)
		case "env":
			out, err = p.EnvStepRunner.Run(ctx, step.RunCommand, step.EnvVarValue, absPath, envs)
			envs[step.EnvVarName] = out
			// We reset out to the empty string because we don't want it to
			// be printed to the PR, it's solely to set the environment variable.
			out = ""
		}

		if out != "" {
			outputs = append(outputs, out)
		}
		if err != nil {
			return outputs, err
		}
	}
	return outputs, nil
}

func (p *DefaultProjectCommandRunner) doApply(ctx models.ProjectCommandContext) (applyOut string, failure string, err error) {
	repoDir, err := p.WorkingDir.GetWorkingDir(ctx.Pull.BaseRepo, ctx.Pull, ctx.Workspace)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", errors.New("project has not been cloned–did you run plan?")
		}
		return "", "", err
	}
	absPath := filepath.Join(repoDir, ctx.RepoRelDir)
	if _, err = os.Stat(absPath); os.IsNotExist(err) {
		return "", "", DirNotExistErr{RepoRelDir: ctx.RepoRelDir}
	}

	for _, req := range ctx.ApplyRequirements {
		switch req {
		case raw.ApprovedApplyRequirement:
			approved, err := p.PullApprovedChecker.PullIsApproved(ctx.Pull.BaseRepo, ctx.Pull) // nolint: vetshadow
			if err != nil {
				return "", "", errors.Wrap(err, "checking if pull request was approved")
			}
			if !approved {
				return "", "Pull request must be approved by at least one person other than the author before running apply.", nil
			}
		case raw.MergeableApplyRequirement:
			if !ctx.PullMergeable {
				return "", "Pull request must be mergeable before running apply.", nil
			}
		}
	}
	// Acquire internal lock for the directory we're going to operate in.
	unlockFn, err := p.WorkingDirLocker.TryLock(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, ctx.Workspace)
	if err != nil {
		return "", "", err
	}
	defer unlockFn()

	outputs, err := p.runSteps(ctx.Steps, ctx, absPath)
	p.Webhooks.Send(ctx.Log, webhooks.ApplyResult{ // nolint: errcheck
		Workspace: ctx.Workspace,
		User:      ctx.User,
		Repo:      ctx.Pull.BaseRepo,
		Pull:      ctx.Pull,
		Success:   err == nil,
		Directory: ctx.RepoRelDir,
	})
	if err != nil {
		return "", "", fmt.Errorf("%s\n%s", err, strings.Join(outputs, "\n"))
	}
	return strings.Join(outputs, "\n"), "", nil
}
