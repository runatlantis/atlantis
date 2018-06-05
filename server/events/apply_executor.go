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
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/repoconfig"
	"github.com/runatlantis/atlantis/server/events/run"
	"github.com/runatlantis/atlantis/server/events/terraform"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/webhooks"
)

// ApplyExecutor handles executing terraform apply.
type ApplyExecutor struct {
	VCSClient         vcs.ClientProxy
	Terraform         *terraform.DefaultClient
	RequireApproval   bool
	Run               *run.Run
	AtlantisWorkspace AtlantisWorkspace
	ProjectLocker     *DefaultProjectLocker
	Webhooks          webhooks.Sender
	ExecutionPlanner  *repoconfig.ExecutionPlanner
}

// Execute executes apply for the ctx.
func (a *ApplyExecutor) Execute(ctx *CommandContext) CommandResponse {
	//if a.RequireApproval {
	//	approved, err := a.VCSClient.PullIsApproved(ctx.BaseRepo, ctx.Pull)
	//	if err != nil {
	//		return CommandResponse{Error: errors.Wrap(err, "checking if pull request was approved")}
	//	}
	//	if !approved {
	//		return CommandResponse{Failure: "Pull request must be approved before running apply."}
	//	}
	//	ctx.Log.Info("confirmed pull request was approved")
	//}

	repoDir, err := a.AtlantisWorkspace.GetWorkspace(ctx.BaseRepo, ctx.Pull, ctx.Command.Workspace)
	if err != nil {
		return CommandResponse{Failure: "No workspace found. Did you run plan?"}
	}
	ctx.Log.Info("found workspace in %q", repoDir)

	stage, err := a.ExecutionPlanner.BuildApplyStage(ctx.Log, repoDir, ctx.Command.Workspace, ctx.Command.Dir, ctx.Command.Flags, ctx.User.Username)
	if err != nil {
		return CommandResponse{Error: err}
	}

	// check if we have the lock
	tryLockResponse, err := a.ProjectLocker.TryLock(ctx, models.NewProject(ctx.BaseRepo.FullName, ctx.Command.Dir))
	if err != nil {
		return CommandResponse{ProjectResults: []ProjectResult{{Error: err}}}
	}
	if !tryLockResponse.LockAcquired {
		return CommandResponse{ProjectResults: []ProjectResult{{Failure: tryLockResponse.LockFailureReason}}}
	}

	// Check apply requirements.
	for _, req := range stage.ApplyRequirements {
		isMet, reason := req.IsMet()
		if !isMet {
			return CommandResponse{Failure: reason}
		}
	}

	out, err := stage.Run()

	// Send webhooks even if there's an error.
	a.Webhooks.Send(ctx.Log, webhooks.ApplyResult{ // nolint: errcheck
		Workspace: ctx.Command.Workspace,
		User:      ctx.User,
		Repo:      ctx.BaseRepo,
		Pull:      ctx.Pull,
		Success:   err == nil,
	})

	if err != nil {
		return CommandResponse{Error: err}
	}
	return CommandResponse{ProjectResults: []ProjectResult{{ApplySuccess: out}}}
}
