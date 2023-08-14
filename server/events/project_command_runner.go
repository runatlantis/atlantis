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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
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

//go:generate pegomock generate --package mocks -o mocks/mock_lock_url_generator.go LockURLGenerator

// LockURLGenerator generates urls to locks.
type LockURLGenerator interface {
	// GenerateLockURL returns the full URL to the lock at lockID.
	GenerateLockURL(lockID string) string
}

//go:generate pegomock generate --package mocks -o mocks/mock_step_runner.go StepRunner

// StepRunner runs steps. Steps are individual pieces of execution like
// `terraform plan`.
type StepRunner interface {
	// Run runs the step.
	Run(ctx command.ProjectContext, extraArgs []string, path string, envs map[string]string) (string, error)
}

//go:generate pegomock generate --package mocks -o mocks/mock_custom_step_runner.go CustomStepRunner

// CustomStepRunner runs custom run steps.
type CustomStepRunner interface {
	// Run cmd in path.
	Run(ctx command.ProjectContext, cmd string, path string, envs map[string]string, streamOutput bool, postProcessOutput valid.PostProcessRunOutputOption) (string, error)
}

//go:generate pegomock generate --package mocks -o mocks/mock_env_step_runner.go EnvStepRunner

// EnvStepRunner runs env steps.
type EnvStepRunner interface {
	Run(ctx command.ProjectContext, cmd string, value string, path string, envs map[string]string) (string, error)
}

// MultiEnvStepRunner runs multienv steps.
type MultiEnvStepRunner interface {
	// Run cmd in path.
	Run(ctx command.ProjectContext, cmd string, path string, envs map[string]string) (string, error)
}

//go:generate pegomock generate --package mocks -o mocks/mock_webhooks_sender.go WebhooksSender

// WebhooksSender sends webhook.
type WebhooksSender interface {
	// Send sends the webhook.
	Send(log logging.SimpleLogging, res webhooks.ApplyResult) error
}

//go:generate pegomock generate --package mocks -o mocks/mock_project_command_runner.go ProjectCommandRunner

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

type ProjectImportCommandRunner interface {
	// Import runs terraform import for the project described by ctx.
	Import(ctx command.ProjectContext) command.ProjectResult
}

type ProjectStateCommandRunner interface {
	// StateRm runs terraform state rm for the project described by ctx.
	StateRm(ctx command.ProjectContext) command.ProjectResult
}

// ProjectCommandRunner runs project commands. A project command is a command
// for a specific TF project.
type ProjectCommandRunner interface {
	ProjectPlanCommandRunner
	ProjectApplyCommandRunner
	ProjectPolicyCheckCommandRunner
	ProjectApprovePoliciesCommandRunner
	ProjectVersionCommandRunner
	ProjectImportCommandRunner
	ProjectStateCommandRunner
}

//go:generate pegomock generate --package mocks -o mocks/mock_job_url_setter.go JobURLSetter

type JobURLSetter interface {
	// SetJobURLWithStatus sets the commit status for the project represented by
	// ctx and updates the status with and url to a job.
	SetJobURLWithStatus(ctx command.ProjectContext, cmdName command.Name, status models.CommitStatus, result *command.ProjectResult) error
}

//go:generate pegomock generate --package mocks -o mocks/mock_job_message_sender.go JobMessageSender

type JobMessageSender interface {
	Send(ctx command.ProjectContext, msg string, operationComplete bool)
}

// ProjectOutputWrapper is a decorator that creates a new PR status check per project.
// The status contains a url that outputs current progress of the terraform plan/apply command.
type ProjectOutputWrapper struct {
	ProjectCommandRunner
	JobMessageSender JobMessageSender
	JobURLSetter     JobURLSetter
}

func (p *ProjectOutputWrapper) Plan(ctx command.ProjectContext) command.ProjectResult {
	result := p.updateProjectPRStatus(command.Plan, ctx, p.ProjectCommandRunner.Plan)
	p.JobMessageSender.Send(ctx, "", OperationComplete)
	return result
}

func (p *ProjectOutputWrapper) Apply(ctx command.ProjectContext) command.ProjectResult {
	result := p.updateProjectPRStatus(command.Apply, ctx, p.ProjectCommandRunner.Apply)
	p.JobMessageSender.Send(ctx, "", OperationComplete)
	return result
}

func (p *ProjectOutputWrapper) updateProjectPRStatus(commandName command.Name, ctx command.ProjectContext, execute func(ctx command.ProjectContext) command.ProjectResult) command.ProjectResult {
	// Create a PR status to track project's plan status. The status will
	// include a link to view the progress of atlantis plan command in real
	// time
	if err := p.JobURLSetter.SetJobURLWithStatus(ctx, commandName, models.PendingCommitStatus, nil); err != nil {
		ctx.Log.Err("updating project PR status", err)
	}

	// ensures we are differentiating between project level command and overall command
	result := execute(ctx)

	if result.Error != nil || result.Failure != "" {
		if err := p.JobURLSetter.SetJobURLWithStatus(ctx, commandName, models.FailedCommitStatus, &result); err != nil {
			ctx.Log.Err("updating project PR status", err)
		}

		return result
	}

	if err := p.JobURLSetter.SetJobURLWithStatus(ctx, commandName, models.SuccessCommitStatus, &result); err != nil {
		ctx.Log.Err("updating project PR status", err)
	}

	return result
}

// DefaultProjectCommandRunner implements ProjectCommandRunner.
type DefaultProjectCommandRunner struct {
	VcsClient                 vcs.Client
	Locker                    ProjectLocker
	LockURLGenerator          LockURLGenerator
	InitStepRunner            StepRunner
	PlanStepRunner            StepRunner
	ShowStepRunner            StepRunner
	ApplyStepRunner           StepRunner
	PolicyCheckStepRunner     StepRunner
	VersionStepRunner         StepRunner
	ImportStepRunner          StepRunner
	StateRmStepRunner         StepRunner
	RunStepRunner             CustomStepRunner
	EnvStepRunner             EnvStepRunner
	MultiEnvStepRunner        MultiEnvStepRunner
	PullApprovedChecker       runtime.PullApprovedChecker
	WorkingDir                WorkingDir
	Webhooks                  WebhooksSender
	WorkingDirLocker          WorkingDirLocker
	CommandRequirementHandler CommandRequirementHandler
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
	}
}

// PolicyCheck evaluates policies defined with Rego for the project described by ctx.
func (p *DefaultProjectCommandRunner) PolicyCheck(ctx command.ProjectContext) command.ProjectResult {
	policySuccess, failure, err := p.doPolicyCheck(ctx)
	return command.ProjectResult{
		Command:            command.PolicyCheck,
		PolicyCheckResults: policySuccess,
		Error:              err,
		Failure:            failure,
		RepoRelDir:         ctx.RepoRelDir,
		Workspace:          ctx.Workspace,
		ProjectName:        ctx.ProjectName,
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
	}
}

func (p *DefaultProjectCommandRunner) ApprovePolicies(ctx command.ProjectContext) command.ProjectResult {
	approvedOut, failure, err := p.doApprovePolicies(ctx)
	return command.ProjectResult{
		Command:            command.PolicyCheck,
		Failure:            failure,
		Error:              err,
		PolicyCheckResults: approvedOut,
		RepoRelDir:         ctx.RepoRelDir,
		Workspace:          ctx.Workspace,
		ProjectName:        ctx.ProjectName,
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

// Import runs terraform import for the project described by ctx.
func (p *DefaultProjectCommandRunner) Import(ctx command.ProjectContext) command.ProjectResult {
	importSuccess, failure, err := p.doImport(ctx)
	return command.ProjectResult{
		Command:       command.Import,
		ImportSuccess: importSuccess,
		Error:         err,
		Failure:       failure,
		RepoRelDir:    ctx.RepoRelDir,
		Workspace:     ctx.Workspace,
		ProjectName:   ctx.ProjectName,
	}
}

// StateRm runs terraform state rm for the project described by ctx.
func (p *DefaultProjectCommandRunner) StateRm(ctx command.ProjectContext) command.ProjectResult {
	stateRmSuccess, failure, err := p.doStateRm(ctx)
	return command.ProjectResult{
		Command:        command.State,
		SubCommand:     "rm",
		StateRmSuccess: stateRmSuccess,
		Error:          err,
		Failure:        failure,
		RepoRelDir:     ctx.RepoRelDir,
		Workspace:      ctx.Workspace,
		ProjectName:    ctx.ProjectName,
	}
}

func (p *DefaultProjectCommandRunner) doApprovePolicies(ctx command.ProjectContext) (*models.PolicyCheckResults, string, error) {
	// Acquire Atlantis lock for this repo/dir/workspace.
	lockAttempt, err := p.Locker.TryLock(ctx.Log, ctx.Pull, ctx.User, ctx.Workspace, models.NewProject(ctx.Pull.BaseRepo.FullName, ctx.RepoRelDir), ctx.RepoLocking)
	if err != nil {
		return nil, "", errors.Wrap(err, "acquiring lock")
	}
	if !lockAttempt.LockAcquired {
		return nil, lockAttempt.LockFailureReason, nil
	}
	ctx.Log.Debug("acquired lock for project")

	// Acquire internal lock for the directory we're going to operate in.
	unlockFn, err := p.WorkingDirLocker.TryLock(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, ctx.Workspace, ctx.RepoRelDir)
	if err != nil {
		return nil, "", err
	}
	defer unlockFn()

	teams := []string{}

	policySetCfg := ctx.PolicySets

	// Only query the users team membership if any teams have been configured as owners on any policy set(s).
	if policySetCfg.HasTeamOwners() {
		// A convenient way to access vcsClient. Not sure if best way.
		userTeams, err := p.VcsClient.GetTeamNamesForUser(ctx.Pull.BaseRepo, ctx.User)
		if err != nil {
			ctx.Log.Err("unable to get team membership for user: %s", err)
			return nil, "", err
		}
		teams = append(teams, userTeams...)
	}
	isAdmin := policySetCfg.Owners.IsOwner(ctx.User.Username, teams)

	var failure string

	// Run over each policy set for the project and perform appropriate approval.
	var prjPolicySetResults []models.PolicySetResult
	var prjErr error
	allPassed := true
	for _, policySet := range policySetCfg.PolicySets {
		isOwner := policySet.Owners.IsOwner(ctx.User.Username, teams) || isAdmin
		prjPolicyStatus := ctx.ProjectPolicyStatus
		for i, policyStatus := range prjPolicyStatus {
			ignorePolicy := false
			if policySet.Name == policyStatus.PolicySetName {
				// Policy set either passed or has sufficient approvals. Move on.
				if policyStatus.Passed || (policyStatus.Approvals == policySet.ApproveCount) {
					if !ctx.ClearPolicyApproval {
						ignorePolicy = true
					}
				}
				// Set ignore flag if targeted policy does not match.
				if ctx.PolicySetTarget != "" && (ctx.PolicySetTarget != policySet.Name) {
					ignorePolicy = true
				}
				// Increment approval if user is owner.
				if isOwner && !ignorePolicy {
					if !ctx.ClearPolicyApproval {
						prjPolicyStatus[i].Approvals = policyStatus.Approvals + 1
					} else {
						prjPolicyStatus[i].Approvals = 0
					}
					// User is not authorized to approve policy set.
				} else if !ignorePolicy {
					prjErr = multierror.Append(prjErr, fmt.Errorf("policy set: %s user %s is not a policy owner - please contact policy owners to approve failing policies", policySet.Name, ctx.User.Username))
				}
				// Still bubble up this failure, even if policy set is not targeted.
				if !policyStatus.Passed && (prjPolicyStatus[i].Approvals != policySet.ApproveCount) {
					allPassed = false
				}
				prjPolicySetResults = append(prjPolicySetResults, models.PolicySetResult{
					PolicySetName: policySet.Name,
					Passed:        policyStatus.Passed,
					CurApprovals:  prjPolicyStatus[i].Approvals,
					ReqApprovals:  policySet.ApproveCount,
				})
			}
		}
	}
	if !allPassed {
		failure = `One or more policy sets require additional approval.`
	}
	return &models.PolicyCheckResults{
		LockURL:            p.LockURLGenerator.GenerateLockURL(lockAttempt.LockKey),
		PolicySetResults:   prjPolicySetResults,
		ApplyCmd:           ctx.ApplyCmd,
		RePlanCmd:          ctx.RePlanCmd,
		ApprovePoliciesCmd: ctx.ApprovePoliciesCmd,
	}, failure, prjErr
}

func (p *DefaultProjectCommandRunner) doPolicyCheck(ctx command.ProjectContext) (*models.PolicyCheckResults, string, error) {
	// Acquire Atlantis lock for this repo/dir/workspace.
	// This should already be acquired from the prior plan operation.
	// if for some reason an unlock happens between the plan and policy check step
	// we will attempt to capture the lock here but fail to get the working directory
	// at which point we will unlock again to preserve functionality
	// If we fail to capture the lock here (super unlikely) then we error out and the user is forced to replan
	lockAttempt, err := p.Locker.TryLock(ctx.Log, ctx.Pull, ctx.User, ctx.Workspace, models.NewProject(ctx.Pull.BaseRepo.FullName, ctx.RepoRelDir), ctx.RepoLocking)

	if err != nil {
		return nil, "", errors.Wrap(err, "acquiring lock")
	}
	if !lockAttempt.LockAcquired {
		return nil, lockAttempt.LockFailureReason, nil
	}
	ctx.Log.Debug("acquired lock for project")

	// Acquire internal lock for the directory we're going to operate in.
	// We should refactor this to keep the lock for the duration of plan and policy check since as of now
	// there is a small gap where we don't have the lock and if we can't get this here, we should just unlock the PR.
	unlockFn, err := p.WorkingDirLocker.TryLock(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, ctx.Workspace, ctx.RepoRelDir)
	if err != nil {
		return nil, "", err
	}
	defer unlockFn()

	// we shouldn't attempt to clone this again. If changes occur to the pull request while the plan is happening
	// that shouldn't affect this particular operation.
	repoDir, err := p.WorkingDir.GetWorkingDir(ctx.Pull.BaseRepo, ctx.Pull, ctx.Workspace)
	if err != nil {

		// let's unlock here since something probably nuked our directory between the plan and policy check phase
		if unlockErr := lockAttempt.UnlockFn(); unlockErr != nil {
			ctx.Log.Err("error unlocking state after plan error: %v", unlockErr)
		}

		if os.IsNotExist(err) {
			return nil, "", errors.New("project has not been cloned–did you run plan?")
		}
		return nil, "", err
	}
	absPath := filepath.Join(repoDir, ctx.RepoRelDir)
	if _, err = os.Stat(absPath); os.IsNotExist(err) {

		// let's unlock here since something probably nuked our directory between the plan and policy check phase
		if unlockErr := lockAttempt.UnlockFn(); unlockErr != nil {
			ctx.Log.Err("error unlocking state after plan error: %v", unlockErr)
		}

		return nil, "", DirNotExistErr{RepoRelDir: ctx.RepoRelDir}
	}

	var failure string
	outputs, err := p.runSteps(ctx.Steps, ctx, absPath)
	var errs error
	if err != nil {
		for {
			err = errors.Unwrap(err)
			if err == nil {
				break
			}
			// Exclude errors for failed policies
			if !strings.Contains(err.Error(), "some policies failed") {
				errs = multierror.Append(errs, err)
			}
		}

		if errs != nil {
			// Note: we are explicitly not unlocking the pr here since a failing policy check will require
			// approval
			return nil, "", errs
		}
	}

	// Separate output from custom run steps
	var index int
	var preConftestOutput []string
	var postConftestOutput []string
	var policySetResults []models.PolicySetResult
	for i, output := range outputs {
		index = i
		err = json.Unmarshal([]byte(strings.Join([]string{output}, "\n")), &policySetResults)
		if err == nil {
			break
		}
		preConftestOutput = append(preConftestOutput, output)
	}
	if policySetResults == nil {
		return nil, "", errors.New("unable to unmarshal conftest output")
	}
	if len(outputs) > 0 {
		postConftestOutput = outputs[(index + 1):]
	}

	result := &models.PolicyCheckResults{
		LockURL:            p.LockURLGenerator.GenerateLockURL(lockAttempt.LockKey),
		PreConftestOutput:  strings.Join(preConftestOutput, "\n"),
		PostConftestOutput: strings.Join(postConftestOutput, "\n"),
		PolicySetResults:   policySetResults,
		RePlanCmd:          ctx.RePlanCmd,
		ApplyCmd:           ctx.ApplyCmd,
		ApprovePoliciesCmd: ctx.ApprovePoliciesCmd,
	}

	// Using this function instead of catching failed policy runs with errors, for cases when '--no-fail' is passed to conftest.
	// One reason to pass such an arg to conftest would be to prevent workflow termination so custom run scripts
	// can be run after the conftest step.
	ctx.Log.Err(strings.Join(outputs, "\n"))
	if !result.PolicyCleared() {
		failure = "Some policy sets did not pass."
	}

	return result, failure, nil
}

func (p *DefaultProjectCommandRunner) doPlan(ctx command.ProjectContext) (*models.PlanSuccess, string, error) {
	// Acquire Atlantis lock for this repo/dir/workspace.
	lockAttempt, err := p.Locker.TryLock(ctx.Log, ctx.Pull, ctx.User, ctx.Workspace, models.NewProject(ctx.Pull.BaseRepo.FullName, ctx.RepoRelDir), ctx.RepoLocking)
	if err != nil {
		return nil, "", errors.Wrap(err, "acquiring lock")
	}
	if !lockAttempt.LockAcquired {
		return nil, lockAttempt.LockFailureReason, nil
	}
	ctx.Log.Debug("acquired lock for project")

	// Acquire internal lock for the directory we're going to operate in.
	unlockFn, err := p.WorkingDirLocker.TryLock(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, ctx.Workspace, ctx.RepoRelDir)
	if err != nil {
		return nil, "", err
	}
	defer unlockFn()

	p.WorkingDir.SetSafeToReClone()
	// Clone is idempotent so okay to run even if the repo was already cloned.
	repoDir, hasDiverged, cloneErr := p.WorkingDir.Clone(ctx.HeadRepo, ctx.Pull, ctx.Workspace)
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

	failure, err := p.CommandRequirementHandler.ValidatePlanProject(repoDir, ctx)
	if failure != "" || err != nil {
		return nil, failure, err
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

	failure, err = p.CommandRequirementHandler.ValidateApplyProject(repoDir, ctx)
	if failure != "" || err != nil {
		return "", failure, err
	}

	// Acquire internal lock for the directory we're going to operate in.
	unlockFn, err := p.WorkingDirLocker.TryLock(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, ctx.Workspace, ctx.RepoRelDir)
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
	unlockFn, err := p.WorkingDirLocker.TryLock(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, ctx.Workspace, ctx.RepoRelDir)
	if err != nil {
		return "", "", err
	}
	defer unlockFn()

	outputs, err := p.runSteps(ctx.Steps, ctx, absPath)
	if err != nil {
		return "", "", fmt.Errorf("%s\n%s", err, strings.Join(outputs, "\n"))
	}

	return strings.Join(outputs, "\n"), "", nil
}

func (p *DefaultProjectCommandRunner) doImport(ctx command.ProjectContext) (out *models.ImportSuccess, failure string, err error) {
	// Clone is idempotent so okay to run even if the repo was already cloned.
	repoDir, _, cloneErr := p.WorkingDir.Clone(ctx.HeadRepo, ctx.Pull, ctx.Workspace)
	if cloneErr != nil {
		return nil, "", cloneErr
	}
	projAbsPath := filepath.Join(repoDir, ctx.RepoRelDir)
	if _, err = os.Stat(projAbsPath); os.IsNotExist(err) {
		return nil, "", DirNotExistErr{RepoRelDir: ctx.RepoRelDir}
	}

	failure, err = p.CommandRequirementHandler.ValidateImportProject(repoDir, ctx)
	if failure != "" || err != nil {
		return nil, failure, err
	}

	// Acquire Atlantis lock for this repo/dir/workspace.
	lockAttempt, err := p.Locker.TryLock(ctx.Log, ctx.Pull, ctx.User, ctx.Workspace, models.NewProject(ctx.Pull.BaseRepo.FullName, ctx.RepoRelDir), ctx.RepoLocking)
	if err != nil {
		return nil, "", errors.Wrap(err, "acquiring lock")
	}
	if !lockAttempt.LockAcquired {
		return nil, lockAttempt.LockFailureReason, nil
	}
	ctx.Log.Debug("acquired lock for project")

	// Acquire internal lock for the directory we're going to operate in.
	unlockFn, err := p.WorkingDirLocker.TryLock(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, ctx.Workspace, ctx.RepoRelDir)
	if err != nil {
		return nil, "", err
	}
	defer unlockFn()

	outputs, err := p.runSteps(ctx.Steps, ctx, projAbsPath)
	if err != nil {
		return nil, "", fmt.Errorf("%s\n%s", err, strings.Join(outputs, "\n"))
	}

	// after import, re-plan command is required without import args
	rePlanCmd := strings.TrimSpace(strings.Split(ctx.RePlanCmd, "--")[0])
	return &models.ImportSuccess{
		Output:    strings.Join(outputs, "\n"),
		RePlanCmd: rePlanCmd,
	}, "", nil
}

func (p *DefaultProjectCommandRunner) doStateRm(ctx command.ProjectContext) (out *models.StateRmSuccess, failure string, err error) {
	// Clone is idempotent so okay to run even if the repo was already cloned.
	repoDir, _, cloneErr := p.WorkingDir.Clone(ctx.HeadRepo, ctx.Pull, ctx.Workspace)
	if cloneErr != nil {
		return nil, "", cloneErr
	}
	projAbsPath := filepath.Join(repoDir, ctx.RepoRelDir)
	if _, err = os.Stat(projAbsPath); os.IsNotExist(err) {
		return nil, "", DirNotExistErr{RepoRelDir: ctx.RepoRelDir}
	}

	// Acquire Atlantis lock for this repo/dir/workspace.
	lockAttempt, err := p.Locker.TryLock(ctx.Log, ctx.Pull, ctx.User, ctx.Workspace, models.NewProject(ctx.Pull.BaseRepo.FullName, ctx.RepoRelDir), ctx.RepoLocking)
	if err != nil {
		return nil, "", errors.Wrap(err, "acquiring lock")
	}
	if !lockAttempt.LockAcquired {
		return nil, lockAttempt.LockFailureReason, nil
	}
	ctx.Log.Debug("acquired lock for project")

	// Acquire internal lock for the directory we're going to operate in.
	unlockFn, err := p.WorkingDirLocker.TryLock(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, ctx.Workspace, ctx.RepoRelDir)
	if err != nil {
		return nil, "", err
	}
	defer unlockFn()

	outputs, err := p.runSteps(ctx.Steps, ctx, projAbsPath)
	if err != nil {
		return nil, "", fmt.Errorf("%s\n%s", err, strings.Join(outputs, "\n"))
	}

	// after state rm, re-plan command is required without state rm args
	rePlanCmd := strings.TrimSpace(strings.Split(ctx.RePlanCmd, "--")[0])
	return &models.StateRmSuccess{
		Output:    strings.Join(outputs, "\n"),
		RePlanCmd: rePlanCmd,
	}, "", nil
}

func (p *DefaultProjectCommandRunner) runSteps(steps []valid.Step, ctx command.ProjectContext, absPath string) ([]string, error) {
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
		case "show":
			_, err = p.ShowStepRunner.Run(ctx, step.ExtraArgs, absPath, envs)
		case "policy_check":
			out, err = p.PolicyCheckStepRunner.Run(ctx, step.ExtraArgs, absPath, envs)
		case "apply":
			out, err = p.ApplyStepRunner.Run(ctx, step.ExtraArgs, absPath, envs)
		case "version":
			out, err = p.VersionStepRunner.Run(ctx, step.ExtraArgs, absPath, envs)
		case "import":
			out, err = p.ImportStepRunner.Run(ctx, step.ExtraArgs, absPath, envs)
		case "state_rm":
			out, err = p.StateRmStepRunner.Run(ctx, step.ExtraArgs, absPath, envs)
		case "run":
			out, err = p.RunStepRunner.Run(ctx, step.RunCommand, absPath, envs, true, step.Output)
		case "env":
			out, err = p.EnvStepRunner.Run(ctx, step.RunCommand, step.EnvVarValue, absPath, envs)
			envs[step.EnvVarName] = out
			// We reset out to the empty string because we don't want it to
			// be printed to the PR, it's solely to set the environment variable.
			out = ""
		case "multienv":
			out, err = p.MultiEnvStepRunner.Run(ctx, step.RunCommand, absPath, envs)
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
