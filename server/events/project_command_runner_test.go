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

package events_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/runtime"
	tmocks "github.com/runatlantis/atlantis/server/core/terraform/mocks"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/models/testdata"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	jobmocks "github.com/runatlantis/atlantis/server/jobs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

// Test that it runs the expected plan steps.
func TestDefaultProjectCommandRunner_Plan(t *testing.T) {
	RegisterMockTestingT(t)
	mockInit := mocks.NewMockStepRunner()
	mockPlan := mocks.NewMockStepRunner()
	mockApply := mocks.NewMockStepRunner()
	mockRun := mocks.NewMockCustomStepRunner()
	realEnv := runtime.EnvStepRunner{}
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockLocker := mocks.NewMockProjectLocker()
	mockCommandRequirementHandler := mocks.NewMockCommandRequirementHandler()

	runner := events.DefaultProjectCommandRunner{
		Locker:                    mockLocker,
		LockURLGenerator:          mockURLGenerator{},
		InitStepRunner:            mockInit,
		PlanStepRunner:            mockPlan,
		ApplyStepRunner:           mockApply,
		RunStepRunner:             mockRun,
		EnvStepRunner:             &realEnv,
		PullApprovedChecker:       nil,
		WorkingDir:                mockWorkingDir,
		Webhooks:                  nil,
		WorkingDirLocker:          events.NewDefaultWorkingDirLocker(),
		CommandRequirementHandler: mockCommandRequirementHandler,
	}

	repoDir := t.TempDir()
	When(mockWorkingDir.Clone(
		Any[models.Repo](),
		Any[models.PullRequest](),
		Any[string](),
	)).ThenReturn(repoDir, false, nil)
	When(mockLocker.TryLock(
		Any[logging.SimpleLogging](),
		Any[models.PullRequest](),
		Any[models.User](),
		Any[string](),
		Any[models.Project](),
		AnyBool(),
	)).ThenReturn(&events.TryLockResponse{
		LockAcquired: true,
		LockKey:      "lock-key",
	}, nil)

	expEnvs := map[string]string{
		"name": "value",
	}
	ctx := command.ProjectContext{
		Log: logging.NewNoopLogger(t),
		Steps: []valid.Step{
			{
				StepName:    "env",
				EnvVarName:  "name",
				EnvVarValue: "value",
			},
			{
				StepName: "run",
			},
			{
				StepName: "apply",
			},
			{
				StepName: "plan",
			},
			{
				StepName: "init",
			},
		},
		Workspace:  "default",
		RepoRelDir: ".",
	}
	// Each step will output its step name.
	When(mockInit.Run(ctx, nil, repoDir, expEnvs)).ThenReturn("init", nil)
	When(mockPlan.Run(ctx, nil, repoDir, expEnvs)).ThenReturn("plan", nil)
	When(mockApply.Run(ctx, nil, repoDir, expEnvs)).ThenReturn("apply", nil)
	When(mockRun.Run(ctx, "", repoDir, expEnvs, true, "")).ThenReturn("run", nil)
	res := runner.Plan(ctx)

	Assert(t, res.PlanSuccess != nil, "exp plan success")
	Equals(t, "https://lock-key", res.PlanSuccess.LockURL)
	Equals(t, "run\napply\nplan\ninit", res.PlanSuccess.TerraformOutput)
	expSteps := []string{"run", "apply", "plan", "init", "env"}
	for _, step := range expSteps {
		switch step {
		case "init":
			mockInit.VerifyWasCalledOnce().Run(ctx, nil, repoDir, expEnvs)
		case "plan":
			mockPlan.VerifyWasCalledOnce().Run(ctx, nil, repoDir, expEnvs)
		case "apply":
			mockApply.VerifyWasCalledOnce().Run(ctx, nil, repoDir, expEnvs)
		case "run":
			mockRun.VerifyWasCalledOnce().Run(ctx, "", repoDir, expEnvs, true, "")
		}
	}
}

func TestProjectOutputWrapper(t *testing.T) {
	RegisterMockTestingT(t)
	ctx := command.ProjectContext{
		Log: logging.NewNoopLogger(t),
		Steps: []valid.Step{
			{
				StepName: "plan",
			},
		},
		Workspace:  "default",
		RepoRelDir: ".",
	}

	cases := []struct {
		Description string
		Failure     bool
		Error       bool
		Success     bool
		CommandName command.Name
	}{
		{
			Description: "plan success",
			Success:     true,
			CommandName: command.Plan,
		},
		{
			Description: "plan failure",
			Failure:     true,
			CommandName: command.Plan,
		},
		{
			Description: "plan error",
			Error:       true,
			CommandName: command.Plan,
		},
		{
			Description: "apply success",
			Success:     true,
			CommandName: command.Apply,
		},
		{
			Description: "apply failure",
			Failure:     true,
			CommandName: command.Apply,
		},
		{
			Description: "apply error",
			Error:       true,
			CommandName: command.Apply,
		},
	}

	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			var prjResult command.ProjectResult
			var expCommitStatus models.CommitStatus

			mockJobURLSetter := mocks.NewMockJobURLSetter()
			mockJobMessageSender := mocks.NewMockJobMessageSender()
			mockProjectCommandRunner := mocks.NewMockProjectCommandRunner()

			runner := &events.ProjectOutputWrapper{
				JobURLSetter:         mockJobURLSetter,
				JobMessageSender:     mockJobMessageSender,
				ProjectCommandRunner: mockProjectCommandRunner,
			}

			if c.Success {
				prjResult = command.ProjectResult{
					PlanSuccess:  &models.PlanSuccess{},
					ApplySuccess: "exists",
				}
				expCommitStatus = models.SuccessCommitStatus
			} else if c.Failure {
				prjResult = command.ProjectResult{
					Failure: "failure",
				}
				expCommitStatus = models.FailedCommitStatus
			} else if c.Error {
				prjResult = command.ProjectResult{
					Error: errors.New("error"),
				}
				expCommitStatus = models.FailedCommitStatus
			}

			When(mockProjectCommandRunner.Plan(Any[command.ProjectContext]())).ThenReturn(prjResult)
			When(mockProjectCommandRunner.Apply(Any[command.ProjectContext]())).ThenReturn(prjResult)

			switch c.CommandName {
			case command.Plan:
				runner.Plan(ctx)
			case command.Apply:
				runner.Apply(ctx)
			}

			mockJobURLSetter.VerifyWasCalled(Once()).SetJobURLWithStatus(ctx, c.CommandName, models.PendingCommitStatus, nil)
			mockJobURLSetter.VerifyWasCalled(Once()).SetJobURLWithStatus(ctx, c.CommandName, expCommitStatus, &prjResult)

			switch c.CommandName {
			case command.Plan:
				mockProjectCommandRunner.VerifyWasCalledOnce().Plan(ctx)
			case command.Apply:
				mockProjectCommandRunner.VerifyWasCalledOnce().Apply(ctx)
			}
		})
	}
}

// Test what happens if there's no working dir. This signals that the project
// was never planned.
func TestDefaultProjectCommandRunner_ApplyNotCloned(t *testing.T) {
	mockWorkingDir := mocks.NewMockWorkingDir()
	runner := &events.DefaultProjectCommandRunner{
		WorkingDir: mockWorkingDir,
	}
	ctx := command.ProjectContext{}
	When(mockWorkingDir.GetWorkingDir(ctx.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn("", os.ErrNotExist)

	res := runner.Apply(ctx)
	ErrEquals(t, "project has not been cloned–did you run plan?", res.Error)
}

// Test that if approval is required and the PR isn't approved we give an error.
func TestDefaultProjectCommandRunner_ApplyNotApproved(t *testing.T) {
	RegisterMockTestingT(t)
	mockWorkingDir := mocks.NewMockWorkingDir()
	runner := &events.DefaultProjectCommandRunner{
		WorkingDir:       mockWorkingDir,
		WorkingDirLocker: events.NewDefaultWorkingDirLocker(),
		CommandRequirementHandler: &events.DefaultCommandRequirementHandler{
			WorkingDir: mockWorkingDir,
		},
	}
	ctx := command.ProjectContext{
		ApplyRequirements: []string{"approved"},
	}
	tmp := t.TempDir()
	When(mockWorkingDir.GetWorkingDir(ctx.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(tmp, nil)

	res := runner.Apply(ctx)
	Equals(t, "Pull request must be approved according to the project's approval rules before running apply.", res.Failure)
}

// Test that if mergeable is required and the PR isn't mergeable we give an error.
func TestDefaultProjectCommandRunner_ApplyNotMergeable(t *testing.T) {
	RegisterMockTestingT(t)
	mockWorkingDir := mocks.NewMockWorkingDir()
	runner := &events.DefaultProjectCommandRunner{
		WorkingDir:       mockWorkingDir,
		WorkingDirLocker: events.NewDefaultWorkingDirLocker(),
		CommandRequirementHandler: &events.DefaultCommandRequirementHandler{
			WorkingDir: mockWorkingDir,
		},
	}
	ctx := command.ProjectContext{
		PullReqStatus: models.PullReqStatus{
			Mergeable: false,
		},
		ApplyRequirements: []string{"mergeable"},
	}
	tmp := t.TempDir()
	When(mockWorkingDir.GetWorkingDir(ctx.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(tmp, nil)

	res := runner.Apply(ctx)
	Equals(t, "Pull request must be mergeable before running apply.", res.Failure)
}

// Test that if undiverged is required and the PR is diverged we give an error.
func TestDefaultProjectCommandRunner_ApplyDiverged(t *testing.T) {
	RegisterMockTestingT(t)
	mockWorkingDir := mocks.NewMockWorkingDir()
	runner := &events.DefaultProjectCommandRunner{
		WorkingDir:       mockWorkingDir,
		WorkingDirLocker: events.NewDefaultWorkingDirLocker(),
		CommandRequirementHandler: &events.DefaultCommandRequirementHandler{
			WorkingDir: mockWorkingDir,
		},
	}
	log := logging.NewNoopLogger(t)
	ctx := command.ProjectContext{
		Log:               log,
		ApplyRequirements: []string{"undiverged"},
	}
	tmp := t.TempDir()
	When(mockWorkingDir.GetWorkingDir(ctx.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(tmp, nil)
	When(mockWorkingDir.HasDiverged(tmp)).ThenReturn(true)

	res := runner.Apply(ctx)
	Equals(t, "Default branch must be rebased onto pull request before running apply.", res.Failure)
}

// Test that it runs the expected apply steps.
func TestDefaultProjectCommandRunner_Apply(t *testing.T) {
	cases := []struct {
		description string
		steps       []valid.Step
		applyReqs   []string

		expSteps      []string
		expOut        string
		expFailure    string
		pullMergeable bool
	}{
		{
			description: "normal workflow",
			steps:       valid.DefaultApplyStage.Steps,
			expSteps:    []string{"apply"},
			expOut:      "apply",
		},
		{
			description: "approval required",
			steps:       valid.DefaultApplyStage.Steps,
			applyReqs:   []string{"approved"},
			expSteps:    []string{"approve", "apply"},
			expOut:      "apply",
		},
		{
			description:   "mergeable required",
			steps:         valid.DefaultApplyStage.Steps,
			pullMergeable: true,
			applyReqs:     []string{"mergeable"},
			expSteps:      []string{"apply"},
			expOut:        "apply",
		},
		{
			description:   "mergeable required, pull not mergeable",
			steps:         valid.DefaultApplyStage.Steps,
			pullMergeable: false,
			applyReqs:     []string{"mergeable"},
			expSteps:      []string{""},
			expOut:        "",
			expFailure:    "Pull request must be mergeable before running apply.",
		},
		{
			description:   "mergeable and approved required",
			steps:         valid.DefaultApplyStage.Steps,
			pullMergeable: true,
			applyReqs:     []string{"mergeable", "approved"},
			expSteps:      []string{"approved", "apply"},
			expOut:        "apply",
		},
		{
			description: "workflow with custom apply stage",
			steps: []valid.Step{
				{
					StepName:    "env",
					EnvVarName:  "key",
					EnvVarValue: "value",
				},
				{
					StepName: "run",
				},
				{
					StepName: "apply",
				},
				{
					StepName: "plan",
				},
				{
					StepName: "init",
				},
			},
			expSteps: []string{"env", "run", "apply", "plan", "init"},
			expOut:   "run\napply\nplan\ninit",
		},
	}

	for _, c := range cases {
		if c.description != "workflow with custom apply stage" {
			continue
		}
		t.Run(c.description, func(t *testing.T) {
			RegisterMockTestingT(t)
			mockInit := mocks.NewMockStepRunner()
			mockPlan := mocks.NewMockStepRunner()
			mockApply := mocks.NewMockStepRunner()
			mockRun := mocks.NewMockCustomStepRunner()
			mockEnv := mocks.NewMockEnvStepRunner()
			mockWorkingDir := mocks.NewMockWorkingDir()
			mockLocker := mocks.NewMockProjectLocker()
			mockSender := mocks.NewMockWebhooksSender()
			applyReqHandler := &events.DefaultCommandRequirementHandler{
				WorkingDir: mockWorkingDir,
			}

			runner := events.DefaultProjectCommandRunner{
				Locker:                    mockLocker,
				LockURLGenerator:          mockURLGenerator{},
				InitStepRunner:            mockInit,
				PlanStepRunner:            mockPlan,
				ApplyStepRunner:           mockApply,
				RunStepRunner:             mockRun,
				EnvStepRunner:             mockEnv,
				WorkingDir:                mockWorkingDir,
				Webhooks:                  mockSender,
				WorkingDirLocker:          events.NewDefaultWorkingDirLocker(),
				CommandRequirementHandler: applyReqHandler,
			}
			repoDir := t.TempDir()
			When(mockWorkingDir.GetWorkingDir(
				Any[models.Repo](),
				Any[models.PullRequest](),
				Any[string](),
			)).ThenReturn(repoDir, nil)

			ctx := command.ProjectContext{
				Log:               logging.NewNoopLogger(t),
				Steps:             c.steps,
				Workspace:         "default",
				ApplyRequirements: c.applyReqs,
				RepoRelDir:        ".",
				PullReqStatus: models.PullReqStatus{
					ApprovalStatus: models.ApprovalStatus{
						IsApproved: true,
					},
					Mergeable: true,
				},
			}
			expEnvs := map[string]string{
				"key": "value",
			}
			When(mockInit.Run(ctx, nil, repoDir, expEnvs)).ThenReturn("init", nil)
			When(mockPlan.Run(ctx, nil, repoDir, expEnvs)).ThenReturn("plan", nil)
			When(mockApply.Run(ctx, nil, repoDir, expEnvs)).ThenReturn("apply", nil)
			When(mockRun.Run(ctx, "", repoDir, expEnvs, true, "")).ThenReturn("run", nil)
			When(mockEnv.Run(ctx, "", "value", repoDir, make(map[string]string))).ThenReturn("value", nil)

			res := runner.Apply(ctx)
			Equals(t, c.expOut, res.ApplySuccess)
			Equals(t, c.expFailure, res.Failure)

			for _, step := range c.expSteps {
				switch step {
				case "init":
					mockInit.VerifyWasCalledOnce().Run(ctx, nil, repoDir, expEnvs)
				case "plan":
					mockPlan.VerifyWasCalledOnce().Run(ctx, nil, repoDir, expEnvs)
				case "apply":
					mockApply.VerifyWasCalledOnce().Run(ctx, nil, repoDir, expEnvs)
				case "run":
					mockRun.VerifyWasCalledOnce().Run(ctx, "", repoDir, expEnvs, true, "")
				case "env":
					mockEnv.VerifyWasCalledOnce().Run(ctx, "", "value", repoDir, expEnvs)
				}
			}
		})
	}
}

// Test that it runs the expected apply steps.
func TestDefaultProjectCommandRunner_ApplyRunStepFailure(t *testing.T) {
	RegisterMockTestingT(t)
	mockApply := mocks.NewMockStepRunner()
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockLocker := mocks.NewMockProjectLocker()
	mockSender := mocks.NewMockWebhooksSender()
	applyReqHandler := &events.DefaultCommandRequirementHandler{
		WorkingDir: mockWorkingDir,
	}

	runner := events.DefaultProjectCommandRunner{
		Locker:                    mockLocker,
		LockURLGenerator:          mockURLGenerator{},
		ApplyStepRunner:           mockApply,
		WorkingDir:                mockWorkingDir,
		WorkingDirLocker:          events.NewDefaultWorkingDirLocker(),
		CommandRequirementHandler: applyReqHandler,
		Webhooks:                  mockSender,
	}
	repoDir := t.TempDir()
	When(mockWorkingDir.GetWorkingDir(
		Any[models.Repo](),
		Any[models.PullRequest](),
		Any[string](),
	)).ThenReturn(repoDir, nil)

	ctx := command.ProjectContext{
		Log: logging.NewNoopLogger(t),
		Steps: []valid.Step{
			{
				StepName: "apply",
			},
		},
		Workspace:         "default",
		ApplyRequirements: []string{},
		RepoRelDir:        ".",
	}
	expEnvs := map[string]string{}
	When(mockApply.Run(ctx, nil, repoDir, expEnvs)).ThenReturn("apply", fmt.Errorf("something went wrong"))

	res := runner.Apply(ctx)
	Assert(t, res.ApplySuccess == "", "exp apply failure")

	mockApply.VerifyWasCalledOnce().Run(ctx, nil, repoDir, expEnvs)
}

// Test run and env steps. We don't use mocks for this test since we're
// not running any Terraform.
func TestDefaultProjectCommandRunner_RunEnvSteps(t *testing.T) {
	RegisterMockTestingT(t)
	tfClient := tmocks.NewMockClient()
	tfVersion, err := version.NewVersion("0.12.0")
	Ok(t, err)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	run := runtime.RunStepRunner{
		TerraformExecutor:       tfClient,
		DefaultTFVersion:        tfVersion,
		ProjectCmdOutputHandler: projectCmdOutputHandler,
	}
	env := runtime.EnvStepRunner{
		RunStepRunner: &run,
	}
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockLocker := mocks.NewMockProjectLocker()
	mockCommandRequirementHandler := mocks.NewMockCommandRequirementHandler()

	runner := events.DefaultProjectCommandRunner{
		Locker:                    mockLocker,
		LockURLGenerator:          mockURLGenerator{},
		RunStepRunner:             &run,
		EnvStepRunner:             &env,
		WorkingDir:                mockWorkingDir,
		Webhooks:                  nil,
		WorkingDirLocker:          events.NewDefaultWorkingDirLocker(),
		CommandRequirementHandler: mockCommandRequirementHandler,
	}

	repoDir := t.TempDir()
	When(mockWorkingDir.Clone(
		Any[models.Repo](),
		Any[models.PullRequest](),
		Any[string](),
	)).ThenReturn(repoDir, false, nil)
	When(mockLocker.TryLock(
		Any[logging.SimpleLogging](),
		Any[models.PullRequest](),
		Any[models.User](),
		Any[string](),
		Any[models.Project](),
		AnyBool(),
	)).ThenReturn(&events.TryLockResponse{
		LockAcquired: true,
		LockKey:      "lock-key",
	}, nil)

	ctx := command.ProjectContext{
		Log: logging.NewNoopLogger(t),
		Steps: []valid.Step{
			{
				StepName:   "run",
				RunCommand: "echo var=$var",
			},
			{
				StepName:    "env",
				EnvVarName:  "var",
				EnvVarValue: "value",
			},
			{
				StepName:   "run",
				RunCommand: "echo var=$var",
			},
			{
				StepName:   "env",
				EnvVarName: "dynamic_var",
				RunCommand: "echo dynamic_value",
			},
			{
				StepName:   "run",
				RunCommand: "echo dynamic_var=$dynamic_var",
			},
			// Test overriding the variable
			{
				StepName:    "env",
				EnvVarName:  "dynamic_var",
				EnvVarValue: "overridden",
			},
			{
				StepName:   "run",
				RunCommand: "echo dynamic_var=$dynamic_var",
			},
		},
		Workspace:  "default",
		RepoRelDir: ".",
	}
	res := runner.Plan(ctx)
	Assert(t, res.PlanSuccess != nil, "exp plan success")
	Equals(t, "https://lock-key", res.PlanSuccess.LockURL)
	Equals(t, "var=\n\nvar=value\n\ndynamic_var=dynamic_value\n\ndynamic_var=overridden\n", res.PlanSuccess.TerraformOutput)
}

// Test that it runs the expected import steps.
func TestDefaultProjectCommandRunner_Import(t *testing.T) {
	expEnvs := map[string]string{}
	cases := []struct {
		description   string
		steps         []valid.Step
		importReqs    []string
		pullReqStatus models.PullReqStatus
		setup         func(repoDir string, ctx command.ProjectContext, mockLocker *mocks.MockProjectLocker, mockInit *mocks.MockStepRunner, mockImport *mocks.MockStepRunner)

		expSteps   []string
		expOut     *models.ImportSuccess
		expFailure string
	}{
		{
			description: "normal workflow",
			steps:       valid.DefaultImportStage.Steps,
			importReqs:  []string{"approved"},
			pullReqStatus: models.PullReqStatus{
				ApprovalStatus: models.ApprovalStatus{
					IsApproved: true,
				},
			},
			setup: func(repoDir string, ctx command.ProjectContext, mockLocker *mocks.MockProjectLocker, mockInit *mocks.MockStepRunner, mockImport *mocks.MockStepRunner) {
				When(mockLocker.TryLock(
					Any[logging.SimpleLogging](),
					Any[models.PullRequest](),
					Any[models.User](),
					Any[string](),
					Any[models.Project](),
					AnyBool(),
				)).ThenReturn(&events.TryLockResponse{
					LockAcquired: true,
					LockKey:      "lock-key",
				}, nil)

				When(mockInit.Run(ctx, nil, repoDir, expEnvs)).ThenReturn("init", nil)
				When(mockImport.Run(ctx, nil, repoDir, expEnvs)).ThenReturn("import", nil)
			},
			expSteps: []string{"import"},
			expOut: &models.ImportSuccess{
				Output:    "init\nimport",
				RePlanCmd: "atlantis plan -d .",
			},
		},
		{
			description: "approval required",
			steps:       valid.DefaultImportStage.Steps,
			importReqs:  []string{"approved"},
			pullReqStatus: models.PullReqStatus{
				ApprovalStatus: models.ApprovalStatus{
					IsApproved: false,
				},
			},
			expFailure: "Pull request must be approved according to the project's approval rules before running import.",
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			RegisterMockTestingT(t)
			mockInit := mocks.NewMockStepRunner()
			mockImport := mocks.NewMockStepRunner()
			mockStateRm := mocks.NewMockStepRunner()
			mockWorkingDir := mocks.NewMockWorkingDir()
			mockLocker := mocks.NewMockProjectLocker()
			mockSender := mocks.NewMockWebhooksSender()
			applyReqHandler := &events.DefaultCommandRequirementHandler{
				WorkingDir: mockWorkingDir,
			}

			runner := events.DefaultProjectCommandRunner{
				Locker:                    mockLocker,
				LockURLGenerator:          mockURLGenerator{},
				InitStepRunner:            mockInit,
				ImportStepRunner:          mockImport,
				StateRmStepRunner:         mockStateRm,
				WorkingDir:                mockWorkingDir,
				Webhooks:                  mockSender,
				WorkingDirLocker:          events.NewDefaultWorkingDirLocker(),
				CommandRequirementHandler: applyReqHandler,
			}
			ctx := command.ProjectContext{
				Log:                logging.NewNoopLogger(t),
				Steps:              c.steps,
				Workspace:          "default",
				ImportRequirements: c.importReqs,
				RepoRelDir:         ".",
				PullReqStatus:      c.pullReqStatus,
				RePlanCmd:          "atlantis plan -d . -- addr id",
			}
			repoDir := t.TempDir()
			When(mockWorkingDir.Clone(
				Any[models.Repo](),
				Any[models.PullRequest](),
				Any[string](),
			)).ThenReturn(repoDir, false, nil)
			if c.setup != nil {
				c.setup(repoDir, ctx, mockLocker, mockInit, mockImport)
			}

			res := runner.Import(ctx)
			Equals(t, c.expOut, res.ImportSuccess)
			Equals(t, c.expFailure, res.Failure)

			for _, step := range c.expSteps {
				switch step {
				case "init":
					mockInit.VerifyWasCalledOnce().Run(ctx, nil, repoDir, expEnvs)
				case "import":
					mockImport.VerifyWasCalledOnce().Run(ctx, nil, repoDir, expEnvs)
				}
			}
		})
	}
}

type mockURLGenerator struct{}

func (m mockURLGenerator) GenerateLockURL(lockID string) string {
	return "https://" + lockID
}

// Test approve policies logic.
func TestDefaultProjectCommandRunner_ApprovePolicies(t *testing.T) {
	cases := []struct {
		description string

		policySetCfg        valid.PolicySets
		policySetStatus     []models.PolicySetStatus
		userTeams           []string // Teams the user is a member of
		targetedPolicy      string   // Policy to target when running approvals
		clearPolicyApproval bool

		expOut     []models.PolicySetResult
		expFailure string
		hasErr     bool
	}{
		{
			description: "When user is not an owner at any level, approve policy fails.",
			hasErr:      true,
			policySetCfg: valid.PolicySets{
				Owners: valid.PolicyOwners{
					Users: []string{"someotheruser1"},
				},
				PolicySets: []valid.PolicySet{
					{
						Name:         "policy1",
						ApproveCount: 1,
						Owners: valid.PolicyOwners{
							Teams: []string{"someotherteam"},
						},
					},
				},
			},
			expOut: []models.PolicySetResult{
				{
					PolicySetName: "policy1",
					ReqApprovals:  1,
				},
			},
			expFailure: "One or more policy sets require additional approval.",
		},
		{
			description: "When user is a top-level owner, increment approval count on all policies.",
			hasErr:      false,
			policySetCfg: valid.PolicySets{
				Owners: valid.PolicyOwners{
					Users: []string{testdata.User.Username},
				},
				PolicySets: []valid.PolicySet{
					{
						Name:         "policy1",
						ApproveCount: 1,
					},
					{
						Name:         "policy2",
						ApproveCount: 2,
					},
				},
			},
			expOut: []models.PolicySetResult{
				{
					PolicySetName: "policy1",
					ReqApprovals:  1,
					CurApprovals:  1,
				},
				{
					PolicySetName: "policy2",
					ReqApprovals:  2,
					CurApprovals:  1,
				},
			},
			expFailure: "One or more policy sets require additional approval.",
		},
		{
			description: "When user is not a top-level owner, but an owner of a policy set, increment approval count only the policy set they are an owner of.",
			hasErr:      true,
			policySetCfg: valid.PolicySets{
				PolicySets: []valid.PolicySet{
					{
						Owners: valid.PolicyOwners{
							Users: []string{testdata.User.Username},
						},
						Name:         "policy1",
						ApproveCount: 1,
					},
					{
						Name:         "policy2",
						ApproveCount: 2,
					},
				},
			},
			expOut: []models.PolicySetResult{
				{
					PolicySetName: "policy1",
					ReqApprovals:  1,
					CurApprovals:  1,
				},
				{
					PolicySetName: "policy2",
					ReqApprovals:  2,
					CurApprovals:  0,
				},
			},
			expFailure: "One or more policy sets require additional approval.",
		},
		{
			description: "When user is a top-level ownner through membership, increment approval on all policies.",
			userTeams:   []string{"someuserteam"},
			policySetCfg: valid.PolicySets{
				Owners: valid.PolicyOwners{
					Teams: []string{"someuserteam"},
				},
				PolicySets: []valid.PolicySet{
					{
						Name:         "policy1",
						ApproveCount: 1,
					},
					{
						Name:         "policy2",
						ApproveCount: 1,
					},
				},
			},
			expOut: []models.PolicySetResult{
				{
					PolicySetName: "policy1",
					ReqApprovals:  1,
					CurApprovals:  1,
				},
				{
					PolicySetName: "policy2",
					ReqApprovals:  1,
					CurApprovals:  1,
				},
			},
			expFailure: "",
		},
		{
			description: "When user is not a top-level owner, but is an owner of one policy set through nembership, increment approval only the policy to which they are an owner.",
			hasErr:      true,
			userTeams:   []string{"someuserteam"},
			policySetCfg: valid.PolicySets{
				PolicySets: []valid.PolicySet{
					{
						Owners: valid.PolicyOwners{
							Teams: []string{"someuserteam"},
						},
						Name:         "policy1",
						ApproveCount: 1,
					},
					{
						Name:         "policy2",
						ApproveCount: 1,
					},
				},
			},
			expOut: []models.PolicySetResult{
				{
					PolicySetName: "policy1",
					ReqApprovals:  1,
					CurApprovals:  1,
				},
				{
					PolicySetName: "policy2",
					ReqApprovals:  1,
					CurApprovals:  0,
				},
			},
			expFailure: "One or more policy sets require additional approval.",
		},
		{
			description: "Do not increment or error on passing or fully-approved policy sets.",
			userTeams:   []string{"someuserteam"},
			policySetCfg: valid.PolicySets{
				PolicySets: []valid.PolicySet{
					{
						Owners: valid.PolicyOwners{
							Teams: []string{"someuserteam"},
						},
						Name:         "policy1",
						ApproveCount: 2,
					},
				},
			},
			policySetStatus: []models.PolicySetStatus{
				{
					PolicySetName: "policy1",
					Approvals:     2,
				},
			},
			expOut: []models.PolicySetResult{
				{
					PolicySetName: "policy1",
					ReqApprovals:  2,
					CurApprovals:  2,
				},
			},
			expFailure: ``,
			hasErr:     false,
		},
		{
			description: "Policies should not fail if they pass.",
			userTeams:   []string{"someuserteam"},
			policySetCfg: valid.PolicySets{
				PolicySets: []valid.PolicySet{
					{
						Owners: valid.PolicyOwners{
							Teams: []string{"someuserteam"},
						},
						Name:         "policy1",
						ApproveCount: 2,
					},
				},
			},
			policySetStatus: []models.PolicySetStatus{
				{
					PolicySetName: "policy1",
					Passed:        true,
					Approvals:     0,
				},
			},
			expOut: []models.PolicySetResult{
				{
					PolicySetName: "policy1",
					ReqApprovals:  2,
					CurApprovals:  0,
					Passed:        true,
				},
			},
			expFailure: ``,
			hasErr:     false,
		},
		{
			description:    "Non-targeted failing policies should still trigger failure when a targeted policy is cleared.",
			userTeams:      []string{"someuserteam"},
			targetedPolicy: "policy1",
			policySetCfg: valid.PolicySets{
				PolicySets: []valid.PolicySet{
					{
						Owners: valid.PolicyOwners{
							Teams: []string{"someuserteam"},
						},
						Name:         "policy1",
						ApproveCount: 1,
					},
					{
						Owners: valid.PolicyOwners{
							Teams: []string{"someuserteam"},
						},
						Name:         "policy2",
						ApproveCount: 1,
					},
				},
			},
			policySetStatus: []models.PolicySetStatus{
				{
					PolicySetName: "policy1",
					Approvals:     0,
					Passed:        false,
				},
				{
					PolicySetName: "policy2",
					Approvals:     0,
					Passed:        false,
				},
			},
			expOut: []models.PolicySetResult{
				{
					PolicySetName: "policy1",
					ReqApprovals:  1,
					CurApprovals:  1,
				},
				{
					PolicySetName: "policy2",
					ReqApprovals:  1,
					CurApprovals:  0,
				},
			},
			expFailure: `One or more policy sets require additional approval.`,
			hasErr:     false,
		},
		{
			description:         "Approval count should be zero if ClearPolicyApproval is set.",
			userTeams:           []string{"someuserteam"},
			clearPolicyApproval: true,
			policySetCfg: valid.PolicySets{
				PolicySets: []valid.PolicySet{
					{
						Owners: valid.PolicyOwners{
							Teams: []string{"someuserteam"},
						},
						Name:         "policy1",
						ApproveCount: 1,
					},
					{
						Owners: valid.PolicyOwners{
							Teams: []string{"someuserteam"},
						},
						Name:         "policy2",
						ApproveCount: 2,
					},
				},
			},
			policySetStatus: []models.PolicySetStatus{
				{
					PolicySetName: "policy1",
					Approvals:     1,
					Passed:        false,
				},
				{
					PolicySetName: "policy2",
					Approvals:     1,
					Passed:        false,
				},
			},
			expOut: []models.PolicySetResult{
				{
					PolicySetName: "policy1",
					ReqApprovals:  1,
					CurApprovals:  0,
				},
				{
					PolicySetName: "policy2",
					ReqApprovals:  2,
					CurApprovals:  0,
				},
			},
			expFailure: `One or more policy sets require additional approval.`,
			hasErr:     false,
		},
		{
			description:         "Approval count should not clear if user is not owner and ClearPolicyApproval is set.",
			userTeams:           []string{"someuserteam"},
			clearPolicyApproval: true,
			policySetCfg: valid.PolicySets{
				PolicySets: []valid.PolicySet{
					{
						Owners: valid.PolicyOwners{
							Teams: []string{"someuserteam"},
						},
						Name:         "policy1",
						ApproveCount: 1,
					},
					{
						Owners: valid.PolicyOwners{
							Teams: []string{"someotheruserteam"},
						},
						Name:         "policy2",
						ApproveCount: 2,
					},
				},
			},
			policySetStatus: []models.PolicySetStatus{
				{
					PolicySetName: "policy1",
					Approvals:     1,
					Passed:        false,
				},
				{
					PolicySetName: "policy2",
					Approvals:     1,
					Passed:        false,
				},
			},
			expOut: []models.PolicySetResult{
				{
					PolicySetName: "policy1",
					ReqApprovals:  1,
					CurApprovals:  0,
				},
				{
					PolicySetName: "policy2",
					ReqApprovals:  2,
					CurApprovals:  1,
				},
			},
			expFailure: `One or more policy sets require additional approval.`,
			hasErr:     true,
		},
		{
			description:         "Approval count should only clear targeted policies when ClearPolicyApproval is set.",
			userTeams:           []string{"someuserteam"},
			targetedPolicy:      "policy2",
			clearPolicyApproval: true,
			policySetCfg: valid.PolicySets{
				PolicySets: []valid.PolicySet{
					{
						Owners: valid.PolicyOwners{
							Teams: []string{"someuserteam"},
						},
						Name:         "policy1",
						ApproveCount: 1,
					},
					{
						Owners: valid.PolicyOwners{
							Teams: []string{"someuserteam"},
						},
						Name:         "policy2",
						ApproveCount: 2,
					},
				},
			},
			policySetStatus: []models.PolicySetStatus{
				{
					PolicySetName: "policy1",
					Approvals:     1,
					Passed:        false,
				},
				{
					PolicySetName: "policy2",
					Approvals:     1,
					Passed:        false,
				},
			},
			expOut: []models.PolicySetResult{
				{
					PolicySetName: "policy1",
					ReqApprovals:  1,
					CurApprovals:  1,
				},
				{
					PolicySetName: "policy2",
					ReqApprovals:  2,
					CurApprovals:  0,
				},
			},
			expFailure: `One or more policy sets require additional approval.`,
			hasErr:     false,
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			RegisterMockTestingT(t)
			mockVcsClient := vcsmocks.NewMockClient()
			mockInit := mocks.NewMockStepRunner()
			mockPlan := mocks.NewMockStepRunner()
			mockApply := mocks.NewMockStepRunner()
			mockRun := mocks.NewMockCustomStepRunner()
			mockEnv := mocks.NewMockEnvStepRunner()
			mockWorkingDir := mocks.NewMockWorkingDir()
			mockLocker := mocks.NewMockProjectLocker()
			mockSender := mocks.NewMockWebhooksSender()

			runner := events.DefaultProjectCommandRunner{
				Locker:           mockLocker,
				VcsClient:        mockVcsClient,
				LockURLGenerator: mockURLGenerator{},
				InitStepRunner:   mockInit,
				PlanStepRunner:   mockPlan,
				ApplyStepRunner:  mockApply,
				RunStepRunner:    mockRun,
				EnvStepRunner:    mockEnv,
				WorkingDir:       mockWorkingDir,
				Webhooks:         mockSender,
				WorkingDirLocker: events.NewDefaultWorkingDirLocker(),
			}
			repoDir := t.TempDir()
			When(mockWorkingDir.GetWorkingDir(
				Any[models.Repo](),
				Any[models.PullRequest](),
				Any[string](),
			)).ThenReturn(repoDir, nil)
			When(mockLocker.TryLock(
				Any[logging.SimpleLogging](),
				Any[models.PullRequest](),
				Any[models.User](),
				Any[string](),
				Any[models.Project](),
				AnyBool(),
			)).ThenReturn(&events.TryLockResponse{
				LockAcquired: true,
				LockKey:      "lock-key",
			}, nil)

			var projPolicyStatus []models.PolicySetStatus
			if c.policySetStatus == nil {
				for _, p := range c.policySetCfg.PolicySets {
					projPolicyStatus = append(projPolicyStatus, models.PolicySetStatus{
						PolicySetName: p.Name,
					})
				}
			} else {
				projPolicyStatus = c.policySetStatus
			}

			modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}
			When(runner.VcsClient.GetTeamNamesForUser(testdata.GithubRepo, testdata.User)).ThenReturn(c.userTeams, nil)
			ctx := command.ProjectContext{
				User:                testdata.User,
				Log:                 logging.NewNoopLogger(t),
				Workspace:           "default",
				RepoRelDir:          ".",
				PolicySets:          c.policySetCfg,
				ProjectPolicyStatus: projPolicyStatus,
				Pull:                modelPull,
				PolicySetTarget:     c.targetedPolicy,
				ClearPolicyApproval: c.clearPolicyApproval,
			}

			res := runner.ApprovePolicies(ctx)
			Equals(t, c.expOut, res.PolicyCheckResults.PolicySetResults)
			Equals(t, c.expFailure, res.Failure)
			if c.hasErr == true {
				Assert(t, res.Error != nil, "expecting error.")
			} else {
				Assert(t, res.Error == nil, "not expecting error.")
			}
		})
	}
}
