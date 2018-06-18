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
//
package events

import (
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/runtime"
	"github.com/runatlantis/atlantis/server/events/webhooks"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
	"github.com/runatlantis/atlantis/server/logging"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_lock_url_generator.go LockURLGenerator

type LockURLGenerator interface {
	GenerateLockURL(lockID string) string
}

type WebhooksSender interface {
	Send(log *logging.SimpleLogger, result webhooks.ApplyResult) error
}

// PlanSuccess is the result of a successful plan.
type PlanSuccess struct {
	TerraformOutput string
	LockURL         string
}

type ProjectOperator struct {
	Locker            ProjectLocker
	LockURLGenerator  LockURLGenerator
	InitStepOperator  runtime.InitStepOperator
	PlanStepOperator  runtime.PlanStepOperator
	ApplyStepOperator runtime.ApplyStepOperator
	RunStepOperator   runtime.RunStepOperator
	ApprovalOperator  runtime.ApprovalOperator
	Workspace         AtlantisWorkspace
	Webhooks          WebhooksSender
}

func (p *ProjectOperator) Plan(ctx models.ProjectCommandContext, projAbsPathPtr *string) ProjectResult {
	// Acquire Atlantis lock for this repo/dir/workspace.
	lockAttempt, err := p.Locker.TryLock(ctx.Log, ctx.Pull, ctx.User, ctx.Workspace, models.NewProject(ctx.BaseRepo.FullName, ctx.RepoRelPath))
	if err != nil {
		return ProjectResult{Error: errors.Wrap(err, "acquiring lock")}
	}
	if !lockAttempt.LockAcquired {
		return ProjectResult{Failure: lockAttempt.LockFailureReason}
	}
	ctx.Log.Debug("acquired lock for project")

	// Ensure project has been cloned.
	var projAbsPath string
	if projAbsPathPtr == nil {
		ctx.Log.Debug("project has not yet been cloned")
		repoDir, err := p.Workspace.Clone(ctx.Log, ctx.BaseRepo, ctx.HeadRepo, ctx.Pull, ctx.Workspace)
		if err != nil {
			if unlockErr := lockAttempt.UnlockFn(); unlockErr != nil {
				ctx.Log.Err("error unlocking state after plan error: %v", unlockErr)
			}
			return ProjectResult{Error: err}
		}
		projAbsPath = filepath.Join(repoDir, ctx.RepoRelPath)
		ctx.Log.Debug("project successfully cloned to %q", projAbsPath)
	} else {
		projAbsPath = *projAbsPathPtr
		ctx.Log.Debug("project was already cloned to %q", projAbsPath)
	}

	// Use default stage unless another workflow is defined in config
	stage := p.defaultPlanStage()
	if ctx.ProjectConfig != nil && ctx.ProjectConfig.Workflow != nil {
		ctx.Log.Debug("project configured to use workflow %q", *ctx.ProjectConfig.Workflow)
		configuredStage := ctx.GlobalConfig.GetPlanStage(*ctx.ProjectConfig.Workflow)
		if configuredStage != nil {
			ctx.Log.Debug("project will use the configured stage for that workflow")
			stage = *configuredStage
		}
	}
	outputs, err := p.runSteps(stage.Steps, ctx, projAbsPath)
	if err != nil {
		if unlockErr := lockAttempt.UnlockFn(); unlockErr != nil {
			ctx.Log.Err("error unlocking state after plan error: %v", unlockErr)
		}
		// todo: include output from other steps.
		return ProjectResult{Error: err}
	}

	return ProjectResult{
		PlanSuccess: &PlanSuccess{
			LockURL:         p.LockURLGenerator.GenerateLockURL(lockAttempt.LockKey),
			TerraformOutput: strings.Join(outputs, "\n"),
		},
	}
}

func (p *ProjectOperator) runSteps(steps []valid.Step, ctx models.ProjectCommandContext, absPath string) ([]string, error) {
	var outputs []string
	for _, step := range steps {
		var out string
		var err error
		switch step.StepName {
		case "init":
			out, err = p.InitStepOperator.Run(ctx, step.ExtraArgs, absPath)
		case "plan":
			out, err = p.PlanStepOperator.Run(ctx, step.ExtraArgs, absPath)
		case "apply":
			out, err = p.ApplyStepOperator.Run(ctx, step.ExtraArgs, absPath)
		case "run":
			out, err = p.RunStepOperator.Run(ctx, step.RunCommand, absPath)
		}

		if err != nil {
			// todo: include output from other steps.
			return nil, err
		}
		if out != "" {
			outputs = append(outputs, out)
		}
	}
	return outputs, nil
}

func (p *ProjectOperator) Apply(ctx models.ProjectCommandContext, absPath string) ProjectResult {
	if ctx.ProjectConfig != nil {
		for _, req := range ctx.ProjectConfig.ApplyRequirements {
			switch req {
			case "approved":
				approved, err := p.ApprovalOperator.IsApproved(ctx.BaseRepo, ctx.Pull)
				if err != nil {
					return ProjectResult{Error: errors.Wrap(err, "checking if pull request was approved")}
				}
				if !approved {
					return ProjectResult{Failure: "Pull request must be approved before running apply."}
				}
			}
		}
	}

	// Use default stage unless another workflow is defined in config
	stage := p.defaultApplyStage()
	if ctx.ProjectConfig != nil && ctx.ProjectConfig.Workflow != nil {
		configuredStage := ctx.GlobalConfig.GetApplyStage(*ctx.ProjectConfig.Workflow)
		if configuredStage != nil {
			stage = *configuredStage
		}
	}
	outputs, err := p.runSteps(stage.Steps, ctx, absPath)
	p.Webhooks.Send(ctx.Log, webhooks.ApplyResult{ // nolint: errcheck
		Workspace: ctx.Workspace,
		User:      ctx.User,
		Repo:      ctx.BaseRepo,
		Pull:      ctx.Pull,
		Success:   err == nil,
	})
	if err != nil {
		// todo: include output from other steps.
		return ProjectResult{Error: err}
	}
	return ProjectResult{
		ApplySuccess: strings.Join(outputs, "\n"),
	}
}

func (p ProjectOperator) defaultPlanStage() valid.Stage {
	return valid.Stage{
		Steps: []valid.Step{
			{
				StepName: "init",
			},
			{
				StepName: "plan",
			},
		},
	}
}

func (p ProjectOperator) defaultApplyStage() valid.Stage {
	return valid.Stage{
		Steps: []valid.Step{
			{
				StepName: "apply",
			},
		},
	}
}
