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

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/webhooks"
	"github.com/runatlantis/atlantis/server/logging"
)

const OperationComplete = true

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

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_webhooks_sender.go WebhooksSender

// WebhooksSender sends webhook.
type WebhooksSender interface {
	// Send sends the webhook.
	Send(log logging.Logger, res webhooks.ApplyResult) error
}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_project_command_runner.go ProjectCommandRunner

type ProjectPlanCommandRunner interface {
	// Plan runs terraform plan for the project described by ctx.
	Plan(ctx command.ProjectContext) command.ProjectResult
}

type ProjectApplyCommandRunner interface {
	// Apply runs terraform apply for the project described by ctx.
	Apply(ctx command.ProjectContext) command.ProjectResult
}

type ProjectPolicyCheckCommandRunner interface {
	// PolicyCheck runs OPA defined policies for the project desribed by ctx.
	PolicyCheck(ctx command.ProjectContext) command.ProjectResult
}

type ProjectApprovePoliciesCommandRunner interface {
	// Approves any failing OPA policies.
	ApprovePolicies(ctx command.ProjectContext) command.ProjectResult
}

type ProjectVersionCommandRunner interface {
	// Version runs terraform version for the project described by ctx.
	Version(ctx command.ProjectContext) command.ProjectResult
}

// ProjectCommandRunner runs project commands. A project command is a command
// for a specific TF project.
type ProjectCommandRunner interface {
	ProjectPlanCommandRunner
	ProjectApplyCommandRunner
	ProjectPolicyCheckCommandRunner
	ProjectApprovePoliciesCommandRunner
	ProjectVersionCommandRunner
}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_job_closer.go JobCloser

// Job Closer closes a job by marking op complete and clearing up buffers if logs are successfully persisted
type JobCloser interface {
	CloseJob(jobID string, repo models.Repo)
}

func NewProjectCommandRunner(
	stepsRunner runtime.StepsRunner,
	workingDir WorkingDir,
	webhooks WebhooksSender,
	workingDirLocker WorkingDirLocker,
	aggregateApplyRequirements ApplyRequirement,
) *DefaultProjectCommandRunner {
	return &DefaultProjectCommandRunner{
		StepsRunner:                stepsRunner,
		WorkingDir:                 workingDir,
		Webhooks:                   webhooks,
		WorkingDirLocker:           workingDirLocker,
		AggregateApplyRequirements: aggregateApplyRequirements,
	}
}

// DefaultProjectCommandRunner implements ProjectCommandRunner.
type DefaultProjectCommandRunner struct { //create object and test
	StepsRunner                runtime.StepsRunner
	WorkingDir                 WorkingDir
	Webhooks                   WebhooksSender
	WorkingDirLocker           WorkingDirLocker
	AggregateApplyRequirements ApplyRequirement
}

// Plan runs terraform plan for the project described by ctx.
func (p *DefaultProjectCommandRunner) Plan(ctx command.ProjectContext) command.ProjectResult {
	planSuccess, failure, err := p.doPlan(ctx)
	return command.ProjectResult{
		Command:     command.Plan,
		PlanSuccess: planSuccess,
		Error:       err,
		Failure:     failure,
		RepoRelDir:  ctx.RepoRelDir,
		Workspace:   ctx.Workspace,
		ProjectName: ctx.ProjectName,
		StatusID:    ctx.StatusID,
		JobID:       ctx.JobID,
	}
}

// PolicyCheck evaluates policies defined with Rego for the project described by ctx.
func (p *DefaultProjectCommandRunner) PolicyCheck(ctx command.ProjectContext) command.ProjectResult {
	policySuccess, failure, err := p.doPolicyCheck(ctx)
	return command.ProjectResult{
		Command:            command.PolicyCheck,
		PolicyCheckSuccess: policySuccess,
		Error:              err,
		Failure:            failure,
		RepoRelDir:         ctx.RepoRelDir,
		Workspace:          ctx.Workspace,
		ProjectName:        ctx.ProjectName,
		StatusID:           ctx.StatusID,
	}
}

// Apply runs terraform apply for the project described by ctx.
func (p *DefaultProjectCommandRunner) Apply(ctx command.ProjectContext) command.ProjectResult {
	applyOut, failure, err := p.doApply(ctx)
	return command.ProjectResult{
		Command:      command.Apply,
		Failure:      failure,
		Error:        err,
		ApplySuccess: applyOut,
		RepoRelDir:   ctx.RepoRelDir,
		Workspace:    ctx.Workspace,
		ProjectName:  ctx.ProjectName,
		StatusID:     ctx.StatusID,
		JobID:        ctx.JobID,
	}
}

func (p *DefaultProjectCommandRunner) ApprovePolicies(ctx command.ProjectContext) command.ProjectResult {
	approvedOut, failure, err := p.doApprovePolicies(ctx)
	return command.ProjectResult{
		Command:            command.PolicyCheck,
		Failure:            failure,
		Error:              err,
		PolicyCheckSuccess: approvedOut,
		RepoRelDir:         ctx.RepoRelDir,
		Workspace:          ctx.Workspace,
		ProjectName:        ctx.ProjectName,
		StatusID:           ctx.StatusID,
	}
}

func (p *DefaultProjectCommandRunner) Version(ctx command.ProjectContext) command.ProjectResult {
	versionOut, failure, err := p.doVersion(ctx)
	return command.ProjectResult{
		Command:        command.Version,
		Failure:        failure,
		Error:          err,
		VersionSuccess: versionOut,
		RepoRelDir:     ctx.RepoRelDir,
		Workspace:      ctx.Workspace,
		ProjectName:    ctx.ProjectName,
	}
}

func (p *DefaultProjectCommandRunner) doApprovePolicies(ctx command.ProjectContext) (*models.PolicyCheckSuccess, string, error) {

	// TODO: Make this a bit smarter
	// without checking some sort of state that the policy check has indeed passed this is likely to cause issues

	return &models.PolicyCheckSuccess{
		PolicyCheckOutput: "Policies approved",
	}, "", nil
}

func (p *DefaultProjectCommandRunner) doPolicyCheck(ctx command.ProjectContext) (*models.PolicyCheckSuccess, string, error) {
	// Acquire internal lock for the directory we're going to operate in.
	// We should refactor this to keep the lock for the duration of plan and policy check since as of now
	// there is a small gap where we don't have the lock and if we can't get this here, we should just unlock the PR.
	unlockFn, err := p.WorkingDirLocker.TryLock(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, ctx.ProjectCloneDir())
	if err != nil {
		return nil, "", err
	}
	defer unlockFn()

	// we shouldn't attempt to clone this again. If changes occur to the pull request while the plan is happening
	// that shouldn't affect this particular operation.
	repoDir, err := p.WorkingDir.GetWorkingDir(ctx.Pull.BaseRepo, ctx.Pull, ctx.Workspace)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", errors.New("project has not been cloned–did you run plan?")
		}
		return nil, "", err
	}
	absPath := filepath.Join(repoDir, ctx.RepoRelDir)
	if _, err = os.Stat(absPath); os.IsNotExist(err) {
		return nil, "", DirNotExistErr{RepoRelDir: ctx.RepoRelDir}
	}

	outputs, err := p.StepsRunner.Run(ctx.RequestCtx, ctx, absPath)
	if err != nil {
		// Note: we are explicitly not unlocking the pr here since a failing
		// policy check will require approval. This is a bit tricky and hacky
		// solution because we will be missing legitimate failures and assume
		// any failure is a policy check failure.
		return nil, fmt.Sprintf("%s\n%s", err, outputs), nil
	}

	return &models.PolicyCheckSuccess{
		PolicyCheckOutput: outputs,

		// set this to false right now because we don't have this information
		// TODO: refactor the templates in a sane way so we don't need this
		HasDiverged: false,
	}, "", nil
}

func (p *DefaultProjectCommandRunner) doPlan(ctx command.ProjectContext) (*models.PlanSuccess, string, error) {
	unlockFn, err := p.WorkingDirLocker.TryLock(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, ctx.ProjectCloneDir())
	if err != nil {
		return nil, "", err
	}
	defer unlockFn()

	// Clone is idempotent so okay to run even if the repo was already cloned.
	repoDir, hasDiverged, cloneErr := p.WorkingDir.Clone(ctx.Log, ctx.HeadRepo, ctx.Pull, ctx.ProjectCloneDir())
	if cloneErr != nil {
		return nil, "", cloneErr
	}
	projAbsPath := filepath.Join(repoDir, ctx.RepoRelDir)
	if _, err = os.Stat(projAbsPath); os.IsNotExist(err) {
		return nil, "", DirNotExistErr{RepoRelDir: ctx.RepoRelDir}
	}

	outputs, err := p.StepsRunner.Run(ctx.RequestCtx, ctx, projAbsPath)

	if err != nil {
		return nil, "", fmt.Errorf("%s\n%s", err, outputs)
	}

	return &models.PlanSuccess{
		TerraformOutput: outputs,
		RePlanCmd:       ctx.RePlanCmd,
		ApplyCmd:        ctx.ApplyCmd,
		HasDiverged:     hasDiverged,
	}, "", nil
}

func (p *DefaultProjectCommandRunner) doApply(ctx command.ProjectContext) (applyOut string, failure string, err error) {
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

	if !ctx.ForceApply && ctx.WorkflowModeType != valid.PlatformWorkflowMode {
		failure, err = p.AggregateApplyRequirements.ValidateProject(repoDir, ctx)
		if failure != "" || err != nil {
			return "", failure, err
		}
	}
	// Acquire internal lock for the directory we're going to operate in.
	unlockFn, err := p.WorkingDirLocker.TryLock(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, ctx.Workspace)
	if err != nil {
		return "", "", err
	}
	defer unlockFn()

	outputs, err := p.StepsRunner.Run(ctx.RequestCtx, ctx, absPath)

	p.Webhooks.Send(ctx.Log, webhooks.ApplyResult{ // nolint: errcheck
		Workspace: ctx.Workspace,
		User:      ctx.User,
		Repo:      ctx.Pull.BaseRepo,
		Pull:      ctx.Pull,
		Success:   err == nil,
		Directory: ctx.RepoRelDir,
	})

	if err != nil {
		return "", "", fmt.Errorf("%s\n%s", err, outputs)
	}

	return outputs, "", nil
}

func (p *DefaultProjectCommandRunner) doVersion(ctx command.ProjectContext) (versionOut string, failure string, err error) {
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

	// Acquire internal lock for the directory we're going to operate in.
	unlockFn, err := p.WorkingDirLocker.TryLock(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, ctx.Workspace)
	if err != nil {
		return "", "", err
	}
	defer unlockFn()

	outputs, err := p.StepsRunner.Run(ctx.RequestCtx, ctx, absPath)
	if err != nil {
		return "", "", fmt.Errorf("%s\n%s", err, outputs)
	}

	return outputs, "", nil
}
