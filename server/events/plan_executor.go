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
	"fmt"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/locking"
	"github.com/runatlantis/atlantis/server/events/run"
	"github.com/runatlantis/atlantis/server/events/runtime"
	"github.com/runatlantis/atlantis/server/events/terraform"
	"github.com/runatlantis/atlantis/server/events/vcs"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_lock_url_generator.go LockURLGenerator

type LockURLGenerator interface {
	GenerateLockURL(lockID string) string
}

// PlanExecutor handles everything related to running terraform plan.
type PlanExecutor struct {
	VCSClient        vcs.ClientProxy
	Terraform        terraform.Client
	Locker           locking.Locker
	Run              run.Runner
	Workspace        AtlantisWorkspace
	ProjectFinder    ProjectFinder
	ProjectLocker    ProjectLocker
	ExecutionPlanner *ExecutionPlanner
	LockURLGenerator LockURLGenerator
}

// PlanSuccess is the result of a successful plan.
type PlanSuccess struct {
	TerraformOutput string
	LockURL         string
}

// Execute executes terraform plan for the ctx.
func (p *PlanExecutor) Execute(ctx *CommandContext) CommandResponse {
	cloneDir, err := p.Workspace.Clone(ctx.Log, ctx.BaseRepo, ctx.HeadRepo, ctx.Pull, ctx.Command.Workspace)
	if err != nil {
		return CommandResponse{Error: err}
	}

	var stages []runtime.PlanStage
	if ctx.Command.Autoplan {
		modifiedFiles, err := p.VCSClient.GetModifiedFiles(ctx.BaseRepo, ctx.Pull)
		if err != nil {
			return CommandResponse{Error: errors.Wrap(err, "getting modified files")}
		}
		stages, err = p.ExecutionPlanner.BuildAutoplanStages(ctx.Log, ctx.BaseRepo.FullName, cloneDir, ctx.User.Username, modifiedFiles)
		if err != nil {
			return CommandResponse{Error: err}
		}
	} else {
		stage, err := p.ExecutionPlanner.BuildPlanStage(ctx.Log, cloneDir, ctx.Command.Workspace, ctx.Command.Dir, ctx.Command.Flags, ctx.User.Username)
		if err != nil {
			return CommandResponse{Error: err}
		}
		stages = append(stages, stage)
	}

	var projectResults []ProjectResult
	for _, stage := range stages {
		projectResult := ProjectResult{
			Path:      stage.ProjectPath,
			Workspace: stage.Workspace,
		}

		// todo: this should be moved into the plan stage
		//tryLockResponse, err := p.ProjectLocker.TryLock(ctx, models.NewProject(ctx.BaseRepo.FullName, ctx.Command.Dir))
		//if err != nil {
		//	return CommandResponse{ProjectResults: []ProjectResult{{Error: err}}}
		//}
		//if !tryLockResponse.LockAcquired {
		//	return CommandResponse{ProjectResults: []ProjectResult{{Failure: tryLockResponse.LockFailureReason}}}
		//}
		// todo: endtodo

		out, err := stage.Run()
		if err != nil {
			//if unlockErr := tryLockResponse.UnlockFn(); unlockErr != nil {
			//	ctx.Log.Err("error unlocking state after plan error: %s", unlockErr)
			//}
			projectResult.Error = fmt.Errorf("%s\n%s", err.Error(), out)
		} else {
			projectResult.PlanSuccess = &PlanSuccess{
				TerraformOutput: out,
				//LockURL:         p.LockURLGenerator.GenerateLockURL(tryLockResponse.LockKey),
			}
		}
		projectResults = append(projectResults, projectResult)
	}
	return CommandResponse{ProjectResults: projectResults}
}
