// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package events_test

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/boltdb"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/core/terraform"
	tmocks "github.com/runatlantis/atlantis/server/core/terraform/mocks"
	tfclientmocks "github.com/runatlantis/atlantis/server/core/terraform/tfclient/mocks"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/models/testdata"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/events/webhooks"
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
	When(mockWorkingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
		Any[string]())).ThenReturn(repoDir, nil)
	When(mockWorkingDir.GitReadLock(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(func() {})
	When(mockLocker.TryLock(Any[logging.SimpleLogging](), Any[models.PullRequest](), Any[models.User](), Any[string](),
		Any[models.Project](), AnyBool())).ThenReturn(&events.TryLockResponse{LockAcquired: true, LockKey: "lock-key"}, nil)

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
	When(mockRun.Run(ctx, nil, "", repoDir, expEnvs, true, nil, nil)).ThenReturn("run", nil)
	res := runner.Plan(ctx)

	Assert(t, res.PlanSuccess != nil, "exp plan success")
	Equals(t, "https://lock-key", res.PlanSuccess.LockURL)
	t.Logf("output is %s", res.PlanSuccess.TerraformOutput)
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
			mockRun.VerifyWasCalledOnce().Run(ctx, nil, "", repoDir, expEnvs, true, nil, nil)
		}
	}
}

func TestDefaultProjectCommandRunner_PlanSuppressesCustomRunStepStreaming(t *testing.T) {
	RegisterMockTestingT(t)
	mockRun := mocks.NewMockCustomStepRunner()
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockLocker := mocks.NewMockProjectLocker()
	mockCommandRequirementHandler := mocks.NewMockCommandRequirementHandler()

	runner := events.DefaultProjectCommandRunner{
		Locker:                    mockLocker,
		LockURLGenerator:          mockURLGenerator{},
		RunStepRunner:             mockRun,
		WorkingDir:                mockWorkingDir,
		WorkingDirLocker:          events.NewDefaultWorkingDirLocker(),
		CommandRequirementHandler: mockCommandRequirementHandler,
	}

	repoDir := t.TempDir()
	When(mockWorkingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[string]())).
		ThenReturn(repoDir, nil)
	When(mockWorkingDir.GitReadLock(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(func() {})
	When(mockLocker.TryLock(Any[logging.SimpleLogging](), Any[models.PullRequest](), Any[models.User](), Any[string](), Any[models.Project](), AnyBool())).
		ThenReturn(&events.TryLockResponse{LockAcquired: true, LockKey: "lock-key"}, nil)

	ctx := command.ProjectContext{
		Log:               logging.NewNoopLogger(t),
		Steps:             []valid.Step{{StepName: "run"}},
		Workspace:         "default",
		RepoRelDir:        ".",
		SuppressJobOutput: true,
	}
	When(mockRun.Run(ctx, nil, "", repoDir, map[string]string{}, false, nil, nil)).ThenReturn("run", nil)

	res := runner.Plan(ctx)

	Assert(t, res.PlanSuccess != nil, "exp plan success")
	mockRun.VerifyWasCalledOnce().Run(ctx, nil, "", repoDir, map[string]string{}, false, nil, nil)
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

	const (
		expErrorBanner          = "\r\nError:\r\nerror\r\n"
		expMultilineErrorBanner = "\r\nError:\r\nerror\r\nmore detail\r\n"
		expFailureBanner        = "\r\nFailure:\r\nfailure\r\n"
		expMultilineFailure     = "\r\nFailure:\r\nfailure\r\nmore detail\r\n"
	)

	cases := []struct {
		Description      string
		Failure          bool
		Error            bool
		Success          bool
		CommandName      command.Name
		ErrorMessage     string
		FailureMessage   string
		ExpErrorBanner   string
		ExpFailureBanner string
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
			Description:      "plan multiline failure",
			Failure:          true,
			CommandName:      command.Plan,
			FailureMessage:   "failure\nmore detail",
			ExpFailureBanner: expMultilineFailure,
		},
		{
			Description: "plan error",
			Error:       true,
			CommandName: command.Plan,
		},
		{
			Description:    "plan multiline error",
			Error:          true,
			CommandName:    command.Plan,
			ErrorMessage:   "error\nmore detail",
			ExpErrorBanner: expMultilineErrorBanner,
		},
		{
			Description: "plan error and failure",
			Error:       true,
			Failure:     true,
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
		{
			Description: "apply error and failure",
			Error:       true,
			Failure:     true,
			CommandName: command.Apply,
		},
	}

	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			var prjResult command.ProjectCommandOutput
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
				prjResult = command.ProjectCommandOutput{
					PlanSuccess:  &models.PlanSuccess{},
					ApplySuccess: "exists",
				}
				expCommitStatus = models.SuccessCommitStatus
			} else {
				expCommitStatus = models.FailedCommitStatus
				if c.Error {
					errorMessage := "error"
					if c.ErrorMessage != "" {
						errorMessage = c.ErrorMessage
					}
					prjResult.Error = errors.New(errorMessage)
				}
				if c.Failure {
					failureMessage := "failure"
					if c.FailureMessage != "" {
						failureMessage = c.FailureMessage
					}
					prjResult.Failure = failureMessage
				}
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
			if c.CommandName == command.Apply && c.Success {
				mockJobURLSetter.VerifyWasCalled(Never()).SetJobURLWithStatus(ctx, c.CommandName, models.SuccessCommitStatus, &prjResult)
			} else {
				mockJobURLSetter.VerifyWasCalled(Once()).SetJobURLWithStatus(ctx, c.CommandName, expCommitStatus, &prjResult)
			}

			switch c.CommandName {
			case command.Plan:
				mockProjectCommandRunner.VerifyWasCalledOnce().Plan(ctx)
			case command.Apply:
				mockProjectCommandRunner.VerifyWasCalledOnce().Apply(ctx)
			}

			// Assert the ordering and content of JobMessageSender.Send calls.
			// Banners (if any) must be streamed before the OperationComplete signal
			// so the xterm-based job page renders the final status.
			inOrder := new(InOrderContext)
			expectedSends := 0
			if c.Error {
				expectedBanner := c.ExpErrorBanner
				if expectedBanner == "" {
					expectedBanner = expErrorBanner
				}
				mockJobMessageSender.VerifyWasCalledInOrder(Once(), inOrder).Send(ctx, expectedBanner, false)
				expectedSends++
			}
			if c.Failure {
				expectedBanner := c.ExpFailureBanner
				if expectedBanner == "" {
					expectedBanner = expFailureBanner
				}
				mockJobMessageSender.VerifyWasCalledInOrder(Once(), inOrder).Send(ctx, expectedBanner, false)
				expectedSends++
			}
			mockJobMessageSender.VerifyWasCalledInOrder(Once(), inOrder).Send(ctx, "", true)
			expectedSends++
			mockJobMessageSender.VerifyWasCalled(Times(expectedSends)).Send(Any[command.ProjectContext](), Any[string](), Any[bool]())
		})
	}
}

func TestProjectOutputWrapper_DefersRemoteApplyURLSuccessStatus(t *testing.T) {
	RegisterMockTestingT(t)
	ctx := command.ProjectContext{
		Log:        logging.NewNoopLogger(t),
		Workspace:  "default",
		RepoRelDir: ".",
	}
	prjResult := command.ProjectCommandOutput{
		ApplySuccess:    "apply complete",
		ApplySuccessURL: "https://app.terraform.io/app/org/workspace/runs/run-123",
	}
	mockJobURLSetter := mocks.NewMockJobURLSetter()
	mockJobMessageSender := mocks.NewMockJobMessageSender()
	mockProjectCommandRunner := mocks.NewMockProjectCommandRunner()
	runner := &events.ProjectOutputWrapper{
		JobURLSetter:         mockJobURLSetter,
		JobMessageSender:     mockJobMessageSender,
		ProjectCommandRunner: mockProjectCommandRunner,
	}
	When(mockProjectCommandRunner.Apply(Any[command.ProjectContext]())).ThenReturn(prjResult)

	runner.Apply(ctx)

	mockJobURLSetter.VerifyWasCalled(Once()).SetJobURLWithStatus(ctx, command.Apply, models.PendingCommitStatus, nil)
	mockJobURLSetter.VerifyWasCalled(Never()).SetJobURLWithStatus(ctx, command.Apply, models.SuccessCommitStatus, &prjResult)
	mockJobMessageSender.VerifyWasCalledOnce().Send(ctx, "", true)

	runner.PublishDeferredApplyStatuses([]command.ProjectContext{ctx}, command.Result{ProjectResults: []command.ProjectResult{{
		ProjectCommandOutput: prjResult,
		Command:              command.Apply,
		RepoRelDir:           ctx.RepoRelDir,
		Workspace:            ctx.Workspace,
	}}}, models.SuccessCommitStatus)

	mockJobURLSetter.VerifyWasCalled(Once()).SetJobURLWithStatus(ctx, command.Apply, models.SuccessCommitStatus, &prjResult)
}

func TestProjectOutputWrapperSuppressesJobOutput(t *testing.T) {
	RegisterMockTestingT(t)
	mockProjectCommandRunner := mocks.NewMockProjectCommandRunner()
	mockJobURLSetter := mocks.NewMockJobURLSetter()
	mockJobMessageSender := mocks.NewMockJobMessageSender()
	ctx := command.ProjectContext{
		Log:               logging.NewNoopLogger(t),
		SuppressVCSStatus: true,
		SuppressJobOutput: true,
	}
	expected := command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}}
	When(mockProjectCommandRunner.Plan(ctx)).ThenReturn(expected)

	runner := &events.ProjectOutputWrapper{
		ProjectCommandRunner: mockProjectCommandRunner,
		JobURLSetter:         mockJobURLSetter,
		JobMessageSender:     mockJobMessageSender,
	}
	result := runner.Plan(ctx)

	Equals(t, expected, result)
	mockProjectCommandRunner.VerifyWasCalledOnce().Plan(ctx)
	mockJobURLSetter.VerifyWasCalled(Never()).SetJobURLWithStatus(Any[command.ProjectContext](), Any[command.Name](), Any[models.CommitStatus](), Any[*command.ProjectCommandOutput]())
	mockJobMessageSender.VerifyWasCalled(Never()).Send(Any[command.ProjectContext](), Any[string](), Any[bool]())
}

func TestProjectOutputWrapperDoesNotReplayStreamedStepOutput(t *testing.T) {
	RegisterMockTestingT(t)

	mockLocker := mocks.NewMockProjectLocker()
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockPlan := mocks.NewMockStepRunner()
	mockJobURLSetter := mocks.NewMockJobURLSetter()
	mockJobMessageSender := mocks.NewMockJobMessageSender()
	repoDir := t.TempDir()
	ctx := command.ProjectContext{
		Log:       logging.NewNoopLogger(t),
		Steps:     []valid.Step{{StepName: "plan"}},
		Workspace: "default",
		Pull: models.PullRequest{
			BaseRepo: models.Repo{FullName: "runatlantis/atlantis"},
		},
		RepoRelDir: ".",
	}

	When(mockLocker.TryLock(Any[logging.SimpleLogging](), Any[models.PullRequest](), Any[models.User](),
		Any[string](), Any[models.Project](), AnyBool())).
		ThenReturn(&events.TryLockResponse{LockAcquired: true, LockKey: "lock-key", UnlockFn: func() error { return nil }}, nil)
	When(mockWorkingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
		Any[string]())).ThenReturn(repoDir, nil)
	When(mockWorkingDir.GitReadLock(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(func() {})
	When(mockPlan.Run(ctx, nil, repoDir, map[string]string{})).
		ThenReturn("already streamed output", errors.New("error\nmore detail"))

	runner := &events.ProjectOutputWrapper{
		JobURLSetter:     mockJobURLSetter,
		JobMessageSender: mockJobMessageSender,
		ProjectCommandRunner: &events.DefaultProjectCommandRunner{
			Locker:                    mockLocker,
			LockURLGenerator:          mockURLGenerator{},
			PlanStepRunner:            mockPlan,
			WorkingDir:                mockWorkingDir,
			WorkingDirLocker:          events.NewDefaultWorkingDirLocker(),
			CommandRequirementHandler: &events.DefaultCommandRequirementHandler{WorkingDir: mockWorkingDir},
		},
	}

	result := runner.Plan(ctx)

	ErrEquals(t, "error\nmore detail\nalready streamed output", result.Error)
	inOrder := new(InOrderContext)
	mockJobMessageSender.VerifyWasCalledInOrder(Once(), inOrder).Send(ctx, "\r\nError:\r\nerror\r\nmore detail\r\n", false)
	mockJobMessageSender.VerifyWasCalledInOrder(Once(), inOrder).Send(ctx, "", true)
	mockJobMessageSender.VerifyWasCalled(Times(2)).Send(Any[command.ProjectContext](), Any[string](), Any[bool]())
}

func TestProjectOutputWrapperDoesNotReplayCustomRunStepOutput(t *testing.T) {
	RegisterMockTestingT(t)

	tfClient := tfclientmocks.NewMockClient()
	tfDistribution := terraform.NewDistributionTerraformWithDownloader(tmocks.NewMockDownloader())
	tfVersion, err := version.NewVersion("0.12.0")
	Ok(t, err)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	run := runtime.RunStepRunner{
		TerraformExecutor:       tfClient,
		DefaultTFDistribution:   tfDistribution,
		DefaultTFVersion:        tfVersion,
		ProjectCmdOutputHandler: projectCmdOutputHandler,
	}
	mockLocker := mocks.NewMockProjectLocker()
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockJobURLSetter := mocks.NewMockJobURLSetter()
	mockJobMessageSender := mocks.NewMockJobMessageSender()
	repoDir := t.TempDir()
	Ok(t, os.WriteFile(filepath.Join(repoDir, "output.txt"), []byte("already streamed output\n"), 0o600))
	ctx := command.ProjectContext{
		Log: logging.NewNoopLogger(t),
		Steps: []valid.Step{
			{
				StepName:   "run",
				RunCommand: "cat output.txt; exit 1",
			},
		},
		Workspace: "default",
		Pull: models.PullRequest{
			BaseRepo: models.Repo{FullName: "runatlantis/atlantis"},
		},
		RepoRelDir: ".",
	}

	When(tfClient.EnsureVersion(Any[logging.SimpleLogging](), Any[terraform.Distribution](), Any[*version.Version]())).
		ThenReturn(nil)
	When(mockLocker.TryLock(Any[logging.SimpleLogging](), Any[models.PullRequest](), Any[models.User](),
		Any[string](), Any[models.Project](), AnyBool())).
		ThenReturn(&events.TryLockResponse{LockAcquired: true, LockKey: "lock-key", UnlockFn: func() error { return nil }}, nil)
	When(mockWorkingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
		Any[string]())).ThenReturn(repoDir, nil)
	When(mockWorkingDir.GitReadLock(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(func() {})

	runner := &events.ProjectOutputWrapper{
		JobURLSetter:     mockJobURLSetter,
		JobMessageSender: mockJobMessageSender,
		ProjectCommandRunner: &events.DefaultProjectCommandRunner{
			Locker:                    mockLocker,
			LockURLGenerator:          mockURLGenerator{},
			RunStepRunner:             &run,
			WorkingDir:                mockWorkingDir,
			WorkingDirLocker:          events.NewDefaultWorkingDirLocker(),
			CommandRequirementHandler: &events.DefaultCommandRequirementHandler{WorkingDir: mockWorkingDir},
		},
	}

	result := runner.Plan(ctx)

	ErrContains(t, "already streamed output", result.Error)
	_, messages, operationComplete := mockJobMessageSender.VerifyWasCalled(Times(2)).
		Send(Any[command.ProjectContext](), Any[string](), Any[bool]()).GetAllCapturedArguments()
	Assert(t, strings.Contains(messages[0], "\r\nError:\r\nrunning 'sh -c' 'cat output.txt; exit 1'"), fmt.Sprintf("expected error summary banner, got %q", messages[0]))
	Assert(t, !strings.Contains(messages[0], "already streamed output"), fmt.Sprintf("expected banner not to replay run output, got %q", messages[0]))
	Equals(t, "", messages[1])
	Equals(t, false, operationComplete[0])
	Equals(t, true, operationComplete[1])
}

func TestProjectOutputWrapperPreservesNonStreamedEnvStepOutput(t *testing.T) {
	RegisterMockTestingT(t)

	tfClient := tfclientmocks.NewMockClient()
	tfDistribution := terraform.NewDistributionTerraformWithDownloader(tmocks.NewMockDownloader())
	tfVersion, err := version.NewVersion("0.12.0")
	Ok(t, err)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	run := runtime.RunStepRunner{
		TerraformExecutor:       tfClient,
		DefaultTFDistribution:   tfDistribution,
		DefaultTFVersion:        tfVersion,
		ProjectCmdOutputHandler: projectCmdOutputHandler,
	}
	env := runtime.EnvStepRunner{
		RunStepRunner: &run,
	}
	mockLocker := mocks.NewMockProjectLocker()
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockJobURLSetter := mocks.NewMockJobURLSetter()
	mockJobMessageSender := mocks.NewMockJobMessageSender()
	repoDir := t.TempDir()
	Ok(t, os.WriteFile(filepath.Join(repoDir, "output.txt"), []byte("not streamed output\n"), 0o600))
	ctx := command.ProjectContext{
		Log: logging.NewNoopLogger(t),
		Steps: []valid.Step{
			{
				StepName:   "env",
				EnvVarName: "dynamic_var",
				RunCommand: "cat output.txt; exit 1",
			},
		},
		Workspace: "default",
		Pull: models.PullRequest{
			BaseRepo: models.Repo{FullName: "runatlantis/atlantis"},
		},
		RepoRelDir: ".",
	}

	When(tfClient.EnsureVersion(Any[logging.SimpleLogging](), Any[terraform.Distribution](), Any[*version.Version]())).
		ThenReturn(nil)
	When(mockLocker.TryLock(Any[logging.SimpleLogging](), Any[models.PullRequest](), Any[models.User](),
		Any[string](), Any[models.Project](), AnyBool())).
		ThenReturn(&events.TryLockResponse{LockAcquired: true, LockKey: "lock-key", UnlockFn: func() error { return nil }}, nil)
	When(mockWorkingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
		Any[string]())).ThenReturn(repoDir, nil)
	When(mockWorkingDir.GitReadLock(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(func() {})

	runner := &events.ProjectOutputWrapper{
		JobURLSetter:     mockJobURLSetter,
		JobMessageSender: mockJobMessageSender,
		ProjectCommandRunner: &events.DefaultProjectCommandRunner{
			Locker:                    mockLocker,
			LockURLGenerator:          mockURLGenerator{},
			RunStepRunner:             &run,
			EnvStepRunner:             &env,
			WorkingDir:                mockWorkingDir,
			WorkingDirLocker:          events.NewDefaultWorkingDirLocker(),
			CommandRequirementHandler: &events.DefaultCommandRequirementHandler{WorkingDir: mockWorkingDir},
		},
	}

	result := runner.Plan(ctx)

	ErrContains(t, "not streamed output", result.Error)
	_, messages, operationComplete := mockJobMessageSender.VerifyWasCalled(Times(2)).
		Send(Any[command.ProjectContext](), Any[string](), Any[bool]()).GetAllCapturedArguments()
	Assert(t, strings.Contains(messages[0], "\r\nError:\r\n"), fmt.Sprintf("expected error banner, got %q", messages[0]))
	Assert(t, strings.Contains(messages[0], "\r\nnot streamed output\r\n"), fmt.Sprintf("expected banner to include non-streamed output, got %q", messages[0]))
	Equals(t, "", messages[1])
	Equals(t, false, operationComplete[0])
	Equals(t, true, operationComplete[1])
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
			MergeableStatus: models.MergeableStatus{IsMergeable: false},
		},
		ApplyRequirements: []string{"mergeable"},
	}
	tmp := t.TempDir()
	When(mockWorkingDir.GetWorkingDir(ctx.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(tmp, nil)

	res := runner.Apply(ctx)
	Equals(t, "Pull request must be mergeable before running apply.", res.Failure)
}

// Regression test for: a pre-plan requirement failure must stop before plan
// execution side effects after the merge checkout has been refreshed.
func TestDefaultProjectCommandRunner_PlanUndivergedBlocksAfterMergeAgain(t *testing.T) {
	RegisterMockTestingT(t)
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockLocker := mocks.NewMockProjectLocker()
	runner := &events.DefaultProjectCommandRunner{
		Locker:           mockLocker,
		LockURLGenerator: mockURLGenerator{},
		WorkingDir:       mockWorkingDir,
		WorkingDirLocker: events.NewDefaultWorkingDirLocker(),
		CommandRequirementHandler: &events.DefaultCommandRequirementHandler{
			WorkingDir: mockWorkingDir,
		},
	}
	log := logging.NewNoopLogger(t)
	ctx := command.ProjectContext{
		Log:              log,
		PlanRequirements: []string{"undiverged"},
		RepoRelDir:       ".",
		Workspace:        "default",
	}
	repoDir := t.TempDir()
	When(mockLocker.TryLock(Any[logging.SimpleLogging](), Any[models.PullRequest](), Any[models.User](),
		Any[string](), Any[models.Project](), AnyBool())).
		ThenReturn(&events.TryLockResponse{LockAcquired: true, LockKey: "lock-key", UnlockFn: func() error { return nil }}, nil)
	When(mockWorkingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
		Any[string]())).ThenReturn(repoDir, nil)
	When(mockWorkingDir.HasDivergedFromPullHead(Any[logging.SimpleLogging](), Any[string](), Any[string](),
		Any[[]string](), Any[models.PullRequest]())).ThenReturn(true)

	res := runner.Plan(ctx)

	Equals(t, "Default branch must be rebased onto pull request before running plan.", res.Failure)
	mockWorkingDir.VerifyWasCalledOnce().MergeAgain(
		Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[string]())
}

func TestDefaultProjectCommandRunner_PlanChecksProjectPathAfterMergeAgain(t *testing.T) {
	RegisterMockTestingT(t)
	mockLocker := mocks.NewMockProjectLocker()
	repoDir := t.TempDir()
	workingDir := &mergeCreatesProjectWorkingDir{
		repoDir:     repoDir,
		projectPath: "project1",
	}
	runner := &events.DefaultProjectCommandRunner{
		Locker:           mockLocker,
		LockURLGenerator: mockURLGenerator{},
		WorkingDir:       workingDir,
		WorkingDirLocker: events.NewDefaultWorkingDirLocker(),
		CommandRequirementHandler: &events.DefaultCommandRequirementHandler{
			WorkingDir: workingDir,
		},
	}
	ctx := command.ProjectContext{
		Log:        logging.NewNoopLogger(t),
		RepoRelDir: "project1",
		Workspace:  "default",
		Pull: models.PullRequest{
			Num:      1,
			BaseRepo: models.Repo{FullName: "runatlantis/atlantis"},
		},
	}
	When(mockLocker.TryLock(Any[logging.SimpleLogging](), Any[models.PullRequest](), Any[models.User](),
		Any[string](), Any[models.Project](), AnyBool())).
		ThenReturn(&events.TryLockResponse{LockAcquired: true, LockKey: "lock-key", UnlockFn: func() error { return nil }}, nil)

	res := runner.Plan(ctx)

	Equals(t, "", res.Failure)
	Ok(t, res.Error)
	Assert(t, res.PlanSuccess != nil, "expected plan success")
	Assert(t, workingDir.mergeCalled, "expected merge checkout refresh")
}

func TestDefaultProjectCommandRunner_PlanValidationFailureKeepsLockWhenDeletePlanFails(t *testing.T) {
	RegisterMockTestingT(t)
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockLocker := mocks.NewMockProjectLocker()
	runner := &events.DefaultProjectCommandRunner{
		Locker:           mockLocker,
		LockURLGenerator: mockURLGenerator{},
		WorkingDir:       mockWorkingDir,
		WorkingDirLocker: events.NewDefaultWorkingDirLocker(),
		CommandRequirementHandler: &events.DefaultCommandRequirementHandler{
			WorkingDir: mockWorkingDir,
		},
	}
	log := logging.NewNoopLogger(t)
	ctx := command.ProjectContext{
		Log:              log,
		PlanRequirements: []string{"undiverged"},
		RepoRelDir:       ".",
		Workspace:        "default",
		ProjectName:      "project",
		Pull: models.PullRequest{
			Num:      1,
			BaseRepo: models.Repo{FullName: "runatlantis/atlantis"},
		},
	}
	repoDir := t.TempDir()
	unlockCalls := 0
	When(mockLocker.TryLock(Any[logging.SimpleLogging](), Any[models.PullRequest](), Any[models.User](),
		Any[string](), Any[models.Project](), AnyBool())).
		ThenReturn(&events.TryLockResponse{LockAcquired: true, LockKey: "lock-key", UnlockFn: func() error {
			unlockCalls++
			return nil
		}}, nil)
	When(mockWorkingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
		Any[string]())).ThenReturn(repoDir, nil)
	When(mockWorkingDir.HasDivergedFromPullHead(Any[logging.SimpleLogging](), Any[string](), Any[string](),
		Any[[]string](), Any[models.PullRequest]())).ThenReturn(true)
	When(mockWorkingDir.DeletePlan(Any[logging.SimpleLogging](), Eq(ctx.Pull.BaseRepo), Eq(ctx.Pull),
		Eq(ctx.Workspace), Eq(ctx.RepoRelDir), Eq(ctx.ProjectName))).ThenReturn(errors.New("delete failed"))

	res := runner.Plan(ctx)

	Equals(t, "Default branch must be rebased onto pull request before running plan.", res.Failure)
	Assert(t, res.Error != nil, "expected delete plan error")
	Equals(t, "deleting stale plan after plan validation failure: delete failed", res.Error.Error())
	Equals(t, 0, unlockCalls)
	mockWorkingDir.VerifyWasCalledOnce().DeletePlan(Any[logging.SimpleLogging](), Eq(ctx.Pull.BaseRepo), Eq(ctx.Pull),
		Eq(ctx.Workspace), Eq(ctx.RepoRelDir), Eq(ctx.ProjectName))
	mockWorkingDir.VerifyWasCalledOnce().MergeAgain(
		Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[string]())
}

type mergeCreatesProjectWorkingDir struct {
	repoDir     string
	projectPath string
	mergeCalled bool
}

func (m *mergeCreatesProjectWorkingDir) Clone(logging.SimpleLogging, models.Repo, models.PullRequest, string) (string, error) {
	return m.repoDir, nil
}

func (m *mergeCreatesProjectWorkingDir) MergeAgain(logging.SimpleLogging, models.Repo, models.PullRequest, string) (bool, error) {
	m.mergeCalled = true
	return true, os.MkdirAll(filepath.Join(m.repoDir, m.projectPath), 0o755)
}

func (m *mergeCreatesProjectWorkingDir) GetWorkingDir(models.Repo, models.PullRequest, string) (string, error) {
	return m.repoDir, nil
}

func (m *mergeCreatesProjectWorkingDir) HasDiverged(logging.SimpleLogging, string, string, []string, models.PullRequest) bool {
	return false
}

func (m *mergeCreatesProjectWorkingDir) HasDivergedFromPullHead(logging.SimpleLogging, string, string, []string, models.PullRequest) bool {
	return false
}

func (m *mergeCreatesProjectWorkingDir) GetDivergedFiles(logging.SimpleLogging, string, models.PullRequest) ([]string, error) {
	return nil, nil
}

func (m *mergeCreatesProjectWorkingDir) GetDivergedFilesFromPullHead(logging.SimpleLogging, string, models.PullRequest) ([]string, error) {
	return nil, nil
}

func (m *mergeCreatesProjectWorkingDir) GetPullDir(models.Repo, models.PullRequest) (string, error) {
	return m.repoDir, nil
}

func (m *mergeCreatesProjectWorkingDir) Delete(logging.SimpleLogging, models.Repo, models.PullRequest) error {
	return nil
}

func (m *mergeCreatesProjectWorkingDir) DeleteForWorkspace(logging.SimpleLogging, models.Repo, models.PullRequest, string) error {
	return nil
}

func (m *mergeCreatesProjectWorkingDir) DeletePlan(logging.SimpleLogging, models.Repo, models.PullRequest, string, string, string) error {
	return nil
}

func (m *mergeCreatesProjectWorkingDir) GetGitUntrackedFiles(logging.SimpleLogging, models.Repo, models.PullRequest, string) ([]string, error) {
	return nil, nil
}

func (m *mergeCreatesProjectWorkingDir) GitReadLock(models.Repo, models.PullRequest, string) func() {
	return func() {}
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
	When(mockWorkingDir.HasDiverged(ctx.Log, tmp, ctx.RepoRelDir, ctx.AutoplanWhenModified, ctx.Pull)).ThenReturn(true)

	res := runner.Apply(ctx)
	Equals(t, "Default branch must be rebased onto pull request before running apply.", res.Failure)
}

func TestProjectCommandRunner_ApplyRevalidatesPlanUnderLock(t *testing.T) {
	RegisterMockTestingT(t)
	mockApply := mocks.NewMockStepRunner()
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockLocker := mocks.NewMockProjectLocker()
	workingDirLocker := &trackingWorkingDirLocker{}
	validator := &assertLockedApplyPlanValidator{t: t, locker: workingDirLocker}
	runner := &events.DefaultProjectCommandRunner{
		Locker:                    mockLocker,
		LockURLGenerator:          mockURLGenerator{},
		ApplyStepRunner:           mockApply,
		WorkingDir:                mockWorkingDir,
		WorkingDirLocker:          workingDirLocker,
		CommandRequirementHandler: &events.DefaultCommandRequirementHandler{WorkingDir: mockWorkingDir},
		ApplyPlanValidator:        validator,
	}
	repoDir := t.TempDir()
	ctx := command.ProjectContext{
		Log:         logging.NewNoopLogger(t),
		CommandName: command.Apply,
		Steps:       valid.DefaultApplyStage.Steps,
		Workspace:   "default",
		RepoRelDir:  ".",
		Pull: models.PullRequest{
			Num:      1,
			BaseRepo: models.Repo{FullName: "runatlantis/atlantis"},
		},
	}
	When(mockWorkingDir.GetWorkingDir(ctx.Pull.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(repoDir, nil)
	When(mockWorkingDir.GitReadLock(ctx.Pull.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(func() {})
	When(mockLocker.TryLock(Any[logging.SimpleLogging](), Eq(ctx.Pull), Any[models.User](), Eq(ctx.Workspace), Any[models.Project](), AnyBool())).
		ThenReturn(&events.TryLockResponse{LockAcquired: true, LockKey: "lock-key"}, nil)
	When(mockApply.Run(ctx, nil, repoDir, map[string]string{})).ThenReturn("apply", nil)

	res := runner.Apply(ctx)

	Ok(t, res.Error)
	Equals(t, "apply", res.ApplySuccess)
	Assert(t, validator.called, "expected apply plan validator to run")
	mockApply.VerifyWasCalledOnce().Run(ctx, nil, repoDir, map[string]string{})
}

func TestProjectCommandRunner_ApplyRevalidatesImmediatelyBeforeApplyStep(t *testing.T) {
	RegisterMockTestingT(t)
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockLocker := mocks.NewMockProjectLocker()
	workingDirLocker := &trackingWorkingDirLocker{}
	calls := []string{}
	runner := &events.DefaultProjectCommandRunner{
		Locker:                    mockLocker,
		LockURLGenerator:          mockURLGenerator{},
		ApplyStepRunner:           &recordingStepRunner{name: "apply", calls: &calls},
		RunStepRunner:             &mutatingCustomStepRunner{calls: &calls},
		WorkingDir:                mockWorkingDir,
		WorkingDirLocker:          workingDirLocker,
		CommandRequirementHandler: &events.DefaultCommandRequirementHandler{WorkingDir: mockWorkingDir},
		ApplyPlanValidator:        &sequencedApplyPlanValidator{t: t, locker: workingDirLocker, calls: &calls},
	}
	repoDir := t.TempDir()
	ctx := command.ProjectContext{
		Log:         logging.NewNoopLogger(t),
		CommandName: command.Apply,
		Steps: []valid.Step{
			{StepName: "run"},
			{StepName: "apply"},
		},
		Workspace:  "default",
		RepoRelDir: ".",
		Pull: models.PullRequest{
			Num:      1,
			BaseRepo: models.Repo{FullName: "runatlantis/atlantis"},
		},
	}
	When(mockWorkingDir.GetWorkingDir(ctx.Pull.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(repoDir, nil)
	When(mockWorkingDir.GitReadLock(ctx.Pull.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(func() {})
	When(mockLocker.TryLock(Any[logging.SimpleLogging](), Eq(ctx.Pull), Any[models.User](), Eq(ctx.Workspace), Any[models.Project](), AnyBool())).
		ThenReturn(&events.TryLockResponse{LockAcquired: true, LockKey: "lock-key"}, nil)

	res := runner.Apply(ctx)

	Ok(t, res.Error)
	Equals(t, "run\napply", res.ApplySuccess)
	Equals(t, []string{"validate", "run", "validate", "apply"}, calls)
}

func TestProjectCommandRunner_ApplyDoesNotCallApplyStepWhenFinalValidationFails(t *testing.T) {
	RegisterMockTestingT(t)
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockLocker := mocks.NewMockProjectLocker()
	calls := []string{}
	runner := &events.DefaultProjectCommandRunner{
		Locker:                    mockLocker,
		LockURLGenerator:          mockURLGenerator{},
		ApplyStepRunner:           &recordingStepRunner{name: "apply", calls: &calls},
		RunStepRunner:             &mutatingCustomStepRunner{calls: &calls},
		WorkingDir:                mockWorkingDir,
		WorkingDirLocker:          events.NewDefaultWorkingDirLocker(),
		CommandRequirementHandler: &events.DefaultCommandRequirementHandler{WorkingDir: mockWorkingDir},
		ApplyPlanValidator: &sequencedApplyPlanValidator{
			calls: &calls,
			errs:  []error{nil, errors.New("final validation failed; rerun plan")},
		},
	}
	repoDir := t.TempDir()
	ctx := command.ProjectContext{
		Log:         logging.NewNoopLogger(t),
		CommandName: command.Apply,
		Steps: []valid.Step{
			{StepName: "run"},
			{StepName: "apply"},
		},
		Workspace:  "default",
		RepoRelDir: ".",
		Pull: models.PullRequest{
			Num:      1,
			BaseRepo: models.Repo{FullName: "runatlantis/atlantis"},
		},
	}
	When(mockWorkingDir.GetWorkingDir(ctx.Pull.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(repoDir, nil)
	When(mockWorkingDir.GitReadLock(ctx.Pull.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(func() {})
	When(mockLocker.TryLock(Any[logging.SimpleLogging](), Eq(ctx.Pull), Any[models.User](), Eq(ctx.Workspace), Any[models.Project](), AnyBool())).
		ThenReturn(&events.TryLockResponse{LockAcquired: true, LockKey: "lock-key"}, nil)

	res := runner.Apply(ctx)

	Assert(t, res.Error != nil, "expected final validation error")
	Assert(t, strings.Contains(res.Error.Error(), "final validation failed"), "got: %s", res.Error)
	Equals(t, []string{"validate", "run", "validate"}, calls)
}

func TestApplyPlanValidator_RejectsWhenLivePullHeadChanged(t *testing.T) {
	db := newTestBoltDB(t)
	repoDir := t.TempDir()
	oldHead := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	newHead := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	ctx := command.ProjectContext{
		Log:         logging.NewNoopLogger(t),
		CommandName: command.Apply,
		Workspace:   "default",
		RepoRelDir:  ".",
		ProjectName: "projA",
		Pull: models.PullRequest{
			Num:        1,
			HeadCommit: oldHead,
			BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
		},
	}
	_, err := db.UpdatePullWithResults(ctx.Pull, []command.ProjectResult{plannedProjectResult(ctx.RepoRelDir, ctx.Workspace, ctx.ProjectName)})
	Ok(t, err)
	planPath := filepath.Join(repoDir, runtime.GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	Ok(t, os.WriteFile(planPath, []byte("plan"), 0600))
	validator := &events.DefaultApplyPlanValidator{PullStatusFetcher: db, LivePullHeadFetcher: fakeLivePullHeadFetcher{head: newHead}}

	err = validator.ValidateProjectPlan(ctx, repoDir)

	Assert(t, err != nil, "expected live head change error")
	Assert(t, strings.Contains(err.Error(), "recorded plan status is from commit"), "got: %s", err)
	_, err = os.Stat(planPath)
	Ok(t, err)
}

func TestApplyPlanValidator_StaleCommandHeadDoesNotDeleteCurrentLivePlan(t *testing.T) {
	db := newTestBoltDB(t)
	repoDir := t.TempDir()
	oldHead := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	newHead := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	currentPull := models.PullRequest{
		Num:        1,
		HeadCommit: newHead,
		BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
	}
	ctx := command.ProjectContext{
		Log:         logging.NewNoopLogger(t),
		CommandName: command.Apply,
		Workspace:   "default",
		RepoRelDir:  ".",
		ProjectName: "projA",
		Pull: models.PullRequest{
			Num:        currentPull.Num,
			HeadCommit: oldHead,
			BaseRepo:   currentPull.BaseRepo,
		},
	}
	_, err := db.UpdatePullWithResults(currentPull, []command.ProjectResult{plannedProjectResult(ctx.RepoRelDir, ctx.Workspace, ctx.ProjectName)})
	Ok(t, err)
	planPath := filepath.Join(repoDir, runtime.GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	Ok(t, os.WriteFile(planPath, []byte("current plan"), 0600))
	validator := &events.DefaultApplyPlanValidator{PullStatusFetcher: db, LivePullHeadFetcher: fakeLivePullHeadFetcher{head: currentPull.HeadCommit}}

	err = validator.ValidateProjectPlan(ctx, repoDir)

	Assert(t, err != nil, "expected stale command head to be rejected")
	Assert(t, strings.Contains(err.Error(), "pull request head changed"), "got: %s", err)
	contents, readErr := os.ReadFile(planPath)
	Ok(t, readErr)
	Equals(t, "current plan", string(contents))
}

func TestApplyPlanValidator_RejectsPullStatusFromDifferentBaseBranch(t *testing.T) {
	db := newTestBoltDB(t)
	repoDir := t.TempDir()
	ctx := command.ProjectContext{
		Log:         logging.NewNoopLogger(t),
		CommandName: command.Apply,
		Workspace:   "default",
		RepoRelDir:  ".",
		ProjectName: "projA",
		Pull: models.PullRequest{
			Num:        1,
			HeadCommit: "abc123",
			BaseBranch: "main",
			BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
		},
	}
	statusPull := ctx.Pull
	statusPull.BaseBranch = "release"
	_, err := db.UpdatePullWithResults(statusPull, []command.ProjectResult{plannedProjectResult(ctx.RepoRelDir, ctx.Workspace, ctx.ProjectName)})
	Ok(t, err)
	planPath := filepath.Join(repoDir, runtime.GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	Ok(t, os.WriteFile(planPath, []byte("plan"), 0600))
	validator := &events.DefaultApplyPlanValidator{PullStatusFetcher: db}

	err = validator.ValidateProjectPlan(ctx, repoDir)

	Assert(t, err != nil, "expected base branch mismatch")
	Assert(t, strings.Contains(err.Error(), "base branch"), "got: %s", err)
	_, err = os.Stat(planPath)
	Assert(t, os.IsNotExist(err), "expected stale-base plan file to be deleted")
}

func TestApplyPlanValidator_RejectsRecordedPlanStatusWithEmptyBaseWhenLiveBaseKnown(t *testing.T) {
	db := newTestBoltDB(t)
	repoDir := t.TempDir()
	head := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	ctx := command.ProjectContext{
		Log:         logging.NewNoopLogger(t),
		CommandName: command.Apply,
		Workspace:   "default",
		RepoRelDir:  ".",
		ProjectName: "projA",
		Pull: models.PullRequest{
			Num:        1,
			HeadCommit: head,
			BaseBranch: "release",
			BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
		},
	}
	statusPull := ctx.Pull
	statusPull.BaseBranch = ""
	_, err := db.UpdatePullWithResults(statusPull, []command.ProjectResult{plannedProjectResult(ctx.RepoRelDir, ctx.Workspace, ctx.ProjectName)})
	Ok(t, err)
	planPath := filepath.Join(repoDir, runtime.GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	Ok(t, os.WriteFile(planPath, []byte("legacy-base plan"), 0600))
	validator := &events.DefaultApplyPlanValidator{
		PullStatusFetcher:   db,
		LivePullHeadFetcher: fakeLivePullHeadFetcher{head: head, base: ctx.Pull.BaseBranch},
	}

	err = validator.ValidateProjectPlan(ctx, repoDir)

	Assert(t, err != nil, "expected empty recorded base to be rejected")
	Assert(t, strings.Contains(err.Error(), "missing a recorded base branch"), "got: %s", err)
	_, err = os.Stat(planPath)
	Ok(t, err)
}

func TestApplyPlanValidator_RejectsRecordedPlanStatusWithEmptyHeadWhenLiveHeadKnown(t *testing.T) {
	db := newTestBoltDB(t)
	repoDir := t.TempDir()
	liveHead := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	ctx := command.ProjectContext{
		Log:         logging.NewNoopLogger(t),
		CommandName: command.Apply,
		Workspace:   "default",
		RepoRelDir:  ".",
		ProjectName: "projA",
		Pull: models.PullRequest{
			Num:        1,
			HeadCommit: liveHead,
			BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
		},
	}
	statusPull := ctx.Pull
	statusPull.HeadCommit = ""
	_, err := db.UpdatePullWithResults(statusPull, []command.ProjectResult{plannedProjectResult(ctx.RepoRelDir, ctx.Workspace, ctx.ProjectName)})
	Ok(t, err)
	planPath := filepath.Join(repoDir, runtime.GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	Ok(t, os.WriteFile(planPath, []byte("legacy-head plan"), 0600))
	validator := &events.DefaultApplyPlanValidator{
		PullStatusFetcher:   db,
		LivePullHeadFetcher: fakeLivePullHeadFetcher{head: liveHead},
	}

	err = validator.ValidateProjectPlan(ctx, repoDir)

	Assert(t, err != nil, "expected empty recorded head to be rejected")
	Assert(t, strings.Contains(err.Error(), "missing a recorded head commit"), "got: %s", err)
	_, err = os.Stat(planPath)
	Ok(t, err)
}

func TestApplyPlanValidator_RejectsWhenLiveBaseChangedSinceCommandStart(t *testing.T) {
	db := newTestBoltDB(t)
	repoDir := t.TempDir()
	head := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	ctx := command.ProjectContext{
		Log:         logging.NewNoopLogger(t),
		CommandName: command.Apply,
		Workspace:   "default",
		RepoRelDir:  ".",
		ProjectName: "projA",
		Pull: models.PullRequest{
			Num:        1,
			HeadCommit: head,
			BaseBranch: "main",
			BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
		},
	}
	_, err := db.UpdatePullWithResults(ctx.Pull, []command.ProjectResult{plannedProjectResult(ctx.RepoRelDir, ctx.Workspace, ctx.ProjectName)})
	Ok(t, err)
	planPath := filepath.Join(repoDir, runtime.GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	Ok(t, os.WriteFile(planPath, []byte("old-base plan"), 0600))
	validator := &events.DefaultApplyPlanValidator{
		PullStatusFetcher:   db,
		LivePullHeadFetcher: fakeLivePullHeadFetcher{head: head, base: "release"},
	}

	err = validator.ValidateProjectPlan(ctx, repoDir)

	Assert(t, err != nil, "expected live base change to reject apply")
	Assert(t, strings.Contains(err.Error(), "base branch"), "got: %s", err)
	_, err = os.Stat(planPath)
	Ok(t, err)
}

func TestApplyPlanValidator_TargetedApplySameHeadDifferentBaseReturnsStaleCommandError(t *testing.T) {
	db := newTestBoltDB(t)
	repoDir := t.TempDir()
	head := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	commandStartPull := models.PullRequest{
		Num:        1,
		HeadCommit: head,
		BaseBranch: "main",
		BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
	}
	livePull := commandStartPull
	livePull.BaseBranch = "release"
	_, err := db.UpdatePullWithResults(livePull, []command.ProjectResult{plannedProjectResult(".", "default", "projA")})
	Ok(t, err)
	planPath := filepath.Join(repoDir, runtime.GetPlanFilename("default", "projA"))
	Ok(t, os.WriteFile(planPath, []byte("current-base plan"), 0600))
	validator := &events.DefaultApplyPlanValidator{
		PullStatusFetcher:   db,
		LivePullHeadFetcher: fakeLivePullHeadFetcher{head: livePull.HeadCommit, base: livePull.BaseBranch},
	}
	ctx := command.ProjectContext{
		Log:         logging.NewNoopLogger(t),
		CommandName: command.Apply,
		Workspace:   "default",
		RepoRelDir:  ".",
		ProjectName: "projA",
		Pull:        commandStartPull,
	}

	err = validator.ValidateProjectPlan(ctx, repoDir)

	Assert(t, err != nil, "expected same-head base retarget to be stale")
	Assert(t, strings.Contains(err.Error(), "pull request base branch changed"), "got: %s", err)
	contents, readErr := os.ReadFile(planPath)
	Ok(t, readErr)
	Equals(t, "current-base plan", string(contents))
}

func TestProjectCommandRunner_TargetedStaleCommandClassifiedBeforeApplyRequirements(t *testing.T) {
	assertTargetedStaleCommandClassifiedBeforeApplyValidation(t)
}

func TestProjectCommandRunner_NonPRMutableRefChangedBeforeApplyDoesNotRunApply(t *testing.T) {
	repoDir, initialCommit := initProjectRunnerAPIRefGitRepo(t)
	advanceProjectRunnerAPIRefGitMain(t, repoDir)
	runner, mockApply := newNonPRAPIApplyRunner(t, repoDir)
	ctx := nonPRAPIApplyContext(t, initialCommit)

	res := runner.Apply(ctx)

	Assert(t, res.Error != nil, "expected changed mutable ref to fail before apply")
	Assert(t, strings.Contains(res.Error.Error(), "changed"), "got: %s", res.Error)
	mockApply.VerifyWasCalled(Never()).Run(Any[command.ProjectContext](), Any[[]string](), Any[string](), Any[map[string]string]())
}

func TestProjectCommandRunner_NonPRReleaseBranchChangedBeforeApplyDoesNotRunApply(t *testing.T) {
	repoDir, initialCommit := initProjectRunnerAPIRefGitRepo(t)
	createProjectRunnerAPIRefGitBranch(t, repoDir, "release", initialCommit)
	advanceProjectRunnerAPIRefGitBranch(t, repoDir, "release")
	runner, mockApply := newNonPRAPIApplyRunner(t, repoDir)
	ctx := nonPRAPIApplyContext(t, initialCommit)
	ctx.Pull.HeadBranch = "release"

	res := runner.Apply(ctx)

	Assert(t, res.Error != nil, "expected changed release branch to fail before apply")
	Assert(t, strings.Contains(res.Error.Error(), "changed"), "got: %s", res.Error)
	mockApply.VerifyWasCalled(Never()).Run(Any[command.ProjectContext](), Any[[]string](), Any[string](), Any[map[string]string]())
}

func TestProjectCommandRunner_NonPRStableBranchChangedBeforeApplyDoesNotRunApply(t *testing.T) {
	repoDir, initialCommit := initProjectRunnerAPIRefGitRepo(t)
	createProjectRunnerAPIRefGitBranch(t, repoDir, "stable", initialCommit)
	advanceProjectRunnerAPIRefGitBranch(t, repoDir, "stable")
	runner, mockApply := newNonPRAPIApplyRunner(t, repoDir)
	ctx := nonPRAPIApplyContext(t, initialCommit)
	ctx.Pull.HeadBranch = "stable"

	res := runner.Apply(ctx)

	Assert(t, res.Error != nil, "expected changed stable branch to fail before apply")
	Assert(t, strings.Contains(res.Error.Error(), "changed"), "got: %s", res.Error)
	mockApply.VerifyWasCalled(Never()).Run(Any[command.ProjectContext](), Any[[]string](), Any[string](), Any[map[string]string]())
}

func TestProjectCommandRunner_NonPRMutableRefChangedBetweenRunStepAndApplyDoesNotRunApply(t *testing.T) {
	repoDir, initialCommit := initProjectRunnerAPIRefGitRepo(t)
	runner, mockApply := newNonPRAPIApplyRunner(t, repoDir)
	runner.RunStepRunner = &mutatingCustomStepRunner{mutate: func() error {
		advanceProjectRunnerAPIRefGitMain(t, repoDir)
		return nil
	}}
	ctx := nonPRAPIApplyContext(t, initialCommit)
	ctx.Steps = []valid.Step{{StepName: "run"}, {StepName: "apply"}}

	res := runner.Apply(ctx)

	Assert(t, res.Error != nil, "expected changed mutable ref to fail before built-in apply")
	Assert(t, strings.Contains(res.Error.Error(), "changed"), "got: %s", res.Error)
	mockApply.VerifyWasCalled(Never()).Run(Any[command.ProjectContext](), Any[[]string](), Any[string](), Any[map[string]string]())
}

func TestProjectCommandRunner_NonPRMutableRefChangedDuringApplyFailsClosed(t *testing.T) {
	repoDir, initialCommit := initProjectRunnerAPIRefGitRepo(t)
	runner, mockApply := newNonPRAPIApplyRunner(t, repoDir)
	When(mockApply.Run(Any[command.ProjectContext](), Any[[]string](), Any[string](), Any[map[string]string]())).
		Then(func([]Param) ReturnValues {
			advanceProjectRunnerAPIRefGitMain(t, repoDir)
			return ReturnValues{"applied", nil}
		})
	ctx := nonPRAPIApplyContext(t, initialCommit)

	res := runner.Apply(ctx)

	Assert(t, res.Error != nil, "expected changed mutable ref during apply to fail closed")
	Assert(t, strings.Contains(res.Error.Error(), "changed"), "got: %s", res.Error)
	mockApply.VerifyWasCalledOnce().Run(Any[command.ProjectContext](), Any[[]string](), Any[string](), Any[map[string]string]())
}

func TestProjectCommandRunner_TargetedStaleCommandClassifiedBeforeDependencyValidation(t *testing.T) {
	assertTargetedStaleCommandClassifiedBeforeApplyValidation(t)
}

func TestProjectCommandRunner_TargetedBaseRetargetClassifiedBeforeRequirements(t *testing.T) {
	RegisterMockTestingT(t)
	mockApply := mocks.NewMockStepRunner()
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockRequirementHandler := mocks.NewMockCommandRequirementHandler()
	head := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	runner := &events.DefaultProjectCommandRunner{
		ApplyStepRunner:           mockApply,
		WorkingDir:                mockWorkingDir,
		WorkingDirLocker:          events.NewDefaultWorkingDirLocker(),
		CommandRequirementHandler: mockRequirementHandler,
		ApplyPlanValidator:        &events.DefaultApplyPlanValidator{LivePullHeadFetcher: fakeLivePullHeadFetcher{head: head, base: "release"}},
	}
	ctx := command.ProjectContext{
		Log:         logging.NewNoopLogger(t),
		CommandName: command.Apply,
		Steps:       valid.DefaultApplyStage.Steps,
		Workspace:   "default",
		RepoRelDir:  ".",
		ProjectName: "projA",
		Pull: models.PullRequest{
			Num:        1,
			HeadCommit: head,
			BaseBranch: "main",
			BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
		},
	}

	res := runner.Apply(ctx)

	Assert(t, res.Error != nil, "expected stale base command error")
	Assert(t, strings.Contains(res.Error.Error(), "pull request base branch changed"), "got: %s", res.Error)
	mockWorkingDir.VerifyWasCalled(Never()).GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Any[string]())
	mockRequirementHandler.VerifyWasCalled(Never()).ValidateApplyProject(Any[string](), Any[command.ProjectContext]())
	mockRequirementHandler.VerifyWasCalled(Never()).ValidateProjectDependencies(Any[command.ProjectContext]())
	mockApply.VerifyWasCalled(Never()).Run(Any[command.ProjectContext](), Any[[]string](), Any[string](), Any[map[string]string]())
}

func TestProjectCommandRunner_TargetedStaleCommandClassifiedBeforeDirNotExist(t *testing.T) {
	RegisterMockTestingT(t)
	mockApply := mocks.NewMockStepRunner()
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockRequirementHandler := mocks.NewMockCommandRequirementHandler()
	oldHead := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	liveHead := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	runner := &events.DefaultProjectCommandRunner{
		ApplyStepRunner:           mockApply,
		WorkingDir:                mockWorkingDir,
		WorkingDirLocker:          events.NewDefaultWorkingDirLocker(),
		CommandRequirementHandler: mockRequirementHandler,
		ApplyPlanValidator:        &events.DefaultApplyPlanValidator{LivePullHeadFetcher: fakeLivePullHeadFetcher{head: liveHead}},
	}
	ctx := command.ProjectContext{
		Log:         logging.NewNoopLogger(t),
		CommandName: command.Apply,
		Steps:       valid.DefaultApplyStage.Steps,
		Workspace:   "default",
		RepoRelDir:  "deleted-dir",
		ProjectName: "projA",
		Pull: models.PullRequest{
			Num:        1,
			HeadCommit: oldHead,
			BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
		},
	}

	res := runner.Apply(ctx)

	Assert(t, res.Error != nil, "expected stale command-head error")
	Assert(t, strings.Contains(res.Error.Error(), "pull request head changed"), "got: %s", res.Error)
	mockWorkingDir.VerifyWasCalled(Never()).GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Any[string]())
	mockRequirementHandler.VerifyWasCalled(Never()).ValidateApplyProject(Any[string](), Any[command.ProjectContext]())
	mockRequirementHandler.VerifyWasCalled(Never()).ValidateProjectDependencies(Any[command.ProjectContext]())
	mockApply.VerifyWasCalled(Never()).Run(Any[command.ProjectContext](), Any[[]string](), Any[string](), Any[map[string]string]())
}

func assertTargetedStaleCommandClassifiedBeforeApplyValidation(t *testing.T) {
	t.Helper()
	RegisterMockTestingT(t)
	mockApply := mocks.NewMockStepRunner()
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockRequirementHandler := mocks.NewMockCommandRequirementHandler()
	oldHead := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	liveHead := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	runner := &events.DefaultProjectCommandRunner{
		ApplyStepRunner:           mockApply,
		WorkingDir:                mockWorkingDir,
		WorkingDirLocker:          events.NewDefaultWorkingDirLocker(),
		CommandRequirementHandler: mockRequirementHandler,
		ApplyPlanValidator:        &events.DefaultApplyPlanValidator{LivePullHeadFetcher: fakeLivePullHeadFetcher{head: liveHead}},
	}
	repoDir := t.TempDir()
	ctx := command.ProjectContext{
		Log:         logging.NewNoopLogger(t),
		CommandName: command.Apply,
		Steps:       valid.DefaultApplyStage.Steps,
		Workspace:   "default",
		RepoRelDir:  ".",
		ProjectName: "projA",
		Pull: models.PullRequest{
			Num:        1,
			HeadCommit: oldHead,
			BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
		},
	}
	When(mockWorkingDir.GetWorkingDir(ctx.Pull.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(repoDir, nil)

	res := runner.Apply(ctx)

	Assert(t, res.Error != nil, "expected stale command-head error")
	Assert(t, strings.Contains(res.Error.Error(), "pull request head changed"), "got: %s", res.Error)
	mockRequirementHandler.VerifyWasCalled(Never()).ValidateApplyProject(Any[string](), Any[command.ProjectContext]())
	mockRequirementHandler.VerifyWasCalled(Never()).ValidateProjectDependencies(Any[command.ProjectContext]())
	mockApply.VerifyWasCalled(Never()).Run(Any[command.ProjectContext](), Any[[]string](), Any[string](), Any[map[string]string]())
}

func TestValidatePlansForApply_CurrentPlanStillApplyableAfterStaleTargetedApplyFails(t *testing.T) {
	db := newTestBoltDB(t)
	repoDir := t.TempDir()
	oldHead := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	newHead := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	currentPull := models.PullRequest{
		Num:        1,
		HeadCommit: newHead,
		BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
	}
	project := command.ProjectContext{
		Log:         logging.NewNoopLogger(t),
		CommandName: command.Apply,
		Workspace:   "default",
		RepoRelDir:  ".",
		ProjectName: "projA",
		Pull: models.PullRequest{
			Num:        currentPull.Num,
			HeadCommit: oldHead,
			BaseRepo:   currentPull.BaseRepo,
		},
	}
	_, err := db.UpdatePullWithResults(currentPull, []command.ProjectResult{plannedProjectResult(project.RepoRelDir, project.Workspace, project.ProjectName)})
	Ok(t, err)
	planPath := filepath.Join(repoDir, runtime.GetPlanFilename(project.Workspace, project.ProjectName))
	Ok(t, os.WriteFile(planPath, []byte("current plan"), 0600))
	validator := &events.DefaultApplyPlanValidator{PullStatusFetcher: db, LivePullHeadFetcher: fakeLivePullHeadFetcher{head: currentPull.HeadCommit}}

	err = validator.ValidateProjectPlan(project, repoDir)
	Assert(t, err != nil, "expected stale targeted apply to fail")

	_, err = os.Stat(planPath)
	Ok(t, err)
	pullStatus, err := db.GetPullStatus(currentPull)
	Ok(t, err)
	err = events.ValidatePlansForApply(&command.Context{
		Log:        logging.NewNoopLogger(t),
		Pull:       currentPull,
		PullStatus: pullStatus,
	}, []events.PendingPlan{{
		RepoDir:     repoDir,
		RepoRelDir:  project.RepoRelDir,
		Workspace:   project.Workspace,
		ProjectName: project.ProjectName,
	}})
	Ok(t, err)
}

func TestApplyPlanValidator_FailsClosedWhenLiveHeadFetchFails(t *testing.T) {
	db := newTestBoltDB(t)
	repoDir := t.TempDir()
	ctx := command.ProjectContext{
		Log:         logging.NewNoopLogger(t),
		CommandName: command.Apply,
		Workspace:   "default",
		RepoRelDir:  ".",
		ProjectName: "projA",
		Pull: models.PullRequest{
			Num:        1,
			HeadCommit: "abc123",
			BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
		},
	}
	_, err := db.UpdatePullWithResults(ctx.Pull, []command.ProjectResult{plannedProjectResult(ctx.RepoRelDir, ctx.Workspace, ctx.ProjectName)})
	Ok(t, err)
	planPath := filepath.Join(repoDir, runtime.GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	Ok(t, os.WriteFile(planPath, []byte("plan"), 0600))
	validator := &events.DefaultApplyPlanValidator{PullStatusFetcher: db, LivePullHeadFetcher: fakeLivePullHeadFetcher{err: errors.New("vcs unavailable")}}

	err = validator.ValidateProjectPlan(ctx, repoDir)

	Assert(t, err != nil, "expected live head fetch failure")
	Assert(t, strings.Contains(err.Error(), "fetching live pull request"), "got: %s", err)
	_, err = os.Stat(planPath)
	Ok(t, err)
}

func TestApplyPlanValidator_UsesInMemoryPullStatusForAPIApply(t *testing.T) {
	db := newTestBoltDB(t)
	repoDir := t.TempDir()
	planContents := []byte("api plan")
	ctx := command.ProjectContext{
		Log:              logging.NewNoopLogger(t),
		CommandName:      command.Apply,
		API:              true,
		Workspace:        "default",
		RepoRelDir:       ".",
		ProjectName:      "projA",
		ExpectedPlanHash: planHashForContent(planContents),
		Pull: models.PullRequest{
			Num:        1,
			HeadCommit: "abc123",
			BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
		},
	}
	ctx.PullStatus = &models.PullStatus{
		Pull: ctx.Pull,
		Projects: []models.ProjectStatus{{
			Workspace:   ctx.Workspace,
			RepoRelDir:  ctx.RepoRelDir,
			ProjectName: ctx.ProjectName,
			Status:      models.PlannedPlanStatus,
		}},
	}
	planPath := filepath.Join(repoDir, runtime.GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	Ok(t, os.WriteFile(planPath, planContents, 0600))
	validator := &events.DefaultApplyPlanValidator{
		PullStatusFetcher:   db,
		LivePullHeadFetcher: fakeLivePullHeadFetcher{head: ctx.Pull.HeadCommit},
	}

	err := validator.ValidateProjectPlan(ctx, repoDir)

	Ok(t, err)
}

func TestApplyPlanValidator_APIApplyWithBranchRefDoesNotCompareRefStringToLiveSHA(t *testing.T) {
	db := newTestBoltDB(t)
	repoDir := t.TempDir()
	liveHead := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	planContents := []byte("api branch plan")
	ctx := command.ProjectContext{
		Log:              logging.NewNoopLogger(t),
		CommandName:      command.Apply,
		API:              true,
		Workspace:        "default",
		RepoRelDir:       ".",
		ProjectName:      "projA",
		ExpectedPlanHash: planHashForContent(planContents),
		Pull: models.PullRequest{
			Num:        1,
			HeadCommit: "main",
			BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
		},
	}
	ctx.PullStatus = &models.PullStatus{
		Pull: models.PullRequest{
			Num:        ctx.Pull.Num,
			HeadCommit: liveHead,
			BaseRepo:   ctx.Pull.BaseRepo,
		},
		Projects: []models.ProjectStatus{{
			Workspace:   ctx.Workspace,
			RepoRelDir:  ctx.RepoRelDir,
			ProjectName: ctx.ProjectName,
			Status:      models.PlannedPlanStatus,
		}},
	}
	planPath := filepath.Join(repoDir, runtime.GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	Ok(t, os.WriteFile(planPath, planContents, 0600))
	validator := &events.DefaultApplyPlanValidator{
		PullStatusFetcher:   db,
		LivePullHeadFetcher: fakeLivePullHeadFetcher{head: liveHead},
	}

	err := validator.ValidateProjectPlan(ctx, repoDir)

	Ok(t, err)
}

func TestApplyPlanValidator_APIApplyWithResolvedCommitComparesToLiveSHA(t *testing.T) {
	db := newTestBoltDB(t)
	repoDir := t.TempDir()
	oldHead := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	liveHead := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	planContents := []byte("api resolved plan")
	ctx := command.ProjectContext{
		Log:              logging.NewNoopLogger(t),
		CommandName:      command.Apply,
		API:              true,
		Workspace:        "default",
		RepoRelDir:       ".",
		ProjectName:      "projA",
		ExpectedPlanHash: planHashForContent(planContents),
		Pull: models.PullRequest{
			Num:        1,
			HeadCommit: oldHead,
			BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
		},
	}
	ctx.PullStatus = &models.PullStatus{
		Pull: models.PullRequest{
			Num:        ctx.Pull.Num,
			HeadCommit: liveHead,
			BaseRepo:   ctx.Pull.BaseRepo,
		},
		Projects: []models.ProjectStatus{{
			Workspace:   ctx.Workspace,
			RepoRelDir:  ctx.RepoRelDir,
			ProjectName: ctx.ProjectName,
			Status:      models.PlannedPlanStatus,
		}},
	}
	planPath := filepath.Join(repoDir, runtime.GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	Ok(t, os.WriteFile(planPath, planContents, 0600))
	validator := &events.DefaultApplyPlanValidator{
		PullStatusFetcher:   db,
		LivePullHeadFetcher: fakeLivePullHeadFetcher{head: liveHead},
	}

	err := validator.ValidateProjectPlan(ctx, repoDir)

	Assert(t, err != nil, "expected stale resolved API commit to fail")
	Assert(t, strings.Contains(err.Error(), "pull request head changed"), "got: %s", err)
	contents, readErr := os.ReadFile(planPath)
	Ok(t, readErr)
	Equals(t, string(planContents), string(contents))
}

func TestApplyPlanValidator_APIApplyPrefersSeededPullStatusOverStaleDB(t *testing.T) {
	db := newTestBoltDB(t)
	repoDir := t.TempDir()
	staleHead := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	liveHead := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	planContents := []byte("fresh api plan")
	stalePull := models.PullRequest{
		Num:        1,
		HeadCommit: staleHead,
		BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
	}
	_, err := db.UpdatePullWithResults(stalePull, []command.ProjectResult{plannedProjectResult(".", "default", "projA")})
	Ok(t, err)
	ctx := command.ProjectContext{
		Log:              logging.NewNoopLogger(t),
		CommandName:      command.Apply,
		API:              true,
		Workspace:        "default",
		RepoRelDir:       ".",
		ProjectName:      "projA",
		ExpectedPlanHash: planHashForContent(planContents),
		Pull: models.PullRequest{
			Num:        stalePull.Num,
			HeadCommit: liveHead,
			BaseRepo:   stalePull.BaseRepo,
		},
	}
	ctx.PullStatus = &models.PullStatus{
		Pull: ctx.Pull,
		Projects: []models.ProjectStatus{{
			Workspace:   ctx.Workspace,
			RepoRelDir:  ctx.RepoRelDir,
			ProjectName: ctx.ProjectName,
			Status:      models.PlannedPlanStatus,
		}},
	}
	planPath := filepath.Join(repoDir, runtime.GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	Ok(t, os.WriteFile(planPath, planContents, 0600))
	validator := &events.DefaultApplyPlanValidator{
		PullStatusFetcher:   db,
		LivePullHeadFetcher: fakeLivePullHeadFetcher{head: liveHead},
	}

	err = validator.ValidateProjectPlan(ctx, repoDir)

	Ok(t, err)
}

func TestApplyPlanValidator_APIInMemoryErroredPolicyCheckStatusRejectsApply(t *testing.T) {
	db := newTestBoltDB(t)
	repoDir := t.TempDir()
	liveHead := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	planContents := []byte("api policy errored plan")
	ctx := command.ProjectContext{
		Log:              logging.NewNoopLogger(t),
		CommandName:      command.Apply,
		API:              true,
		Workspace:        "default",
		RepoRelDir:       ".",
		ProjectName:      "projA",
		ExpectedPlanHash: planHashForContent(planContents),
		Pull: models.PullRequest{
			Num:        1,
			HeadCommit: liveHead,
			BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
		},
	}
	ctx.PullStatus = &models.PullStatus{
		Pull: ctx.Pull,
		Projects: []models.ProjectStatus{{
			Workspace:   ctx.Workspace,
			RepoRelDir:  ctx.RepoRelDir,
			ProjectName: ctx.ProjectName,
			Status:      models.ErroredPolicyCheckStatus,
		}},
	}
	planPath := filepath.Join(repoDir, runtime.GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	Ok(t, os.WriteFile(planPath, planContents, 0600))
	validator := &events.DefaultApplyPlanValidator{
		PullStatusFetcher:   db,
		LivePullHeadFetcher: fakeLivePullHeadFetcher{head: liveHead},
	}

	err := validator.ValidateProjectPlan(ctx, repoDir)

	Assert(t, err != nil, "expected policy-check errored status to reject API apply")
	Assert(t, strings.Contains(err.Error(), "policy checks have errored"), "got: %s", err)
}

func TestApplyPlanValidator_APIApplyUsesDBWhenNoSeededPullStatusExists(t *testing.T) {
	db := newTestBoltDB(t)
	repoDir := t.TempDir()
	liveHead := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	planContents := []byte("api db plan")
	ctx := command.ProjectContext{
		Log:              logging.NewNoopLogger(t),
		CommandName:      command.Apply,
		API:              true,
		Workspace:        "default",
		RepoRelDir:       ".",
		ProjectName:      "projA",
		ExpectedPlanHash: planHashForContent(planContents),
		Pull: models.PullRequest{
			Num:        1,
			HeadCommit: liveHead,
			BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
		},
	}
	_, err := db.UpdatePullWithResults(ctx.Pull, []command.ProjectResult{plannedProjectResult(ctx.RepoRelDir, ctx.Workspace, ctx.ProjectName)})
	Ok(t, err)
	planPath := filepath.Join(repoDir, runtime.GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	Ok(t, os.WriteFile(planPath, planContents, 0600))
	validator := &events.DefaultApplyPlanValidator{
		PullStatusFetcher:   db,
		LivePullHeadFetcher: fakeLivePullHeadFetcher{head: liveHead},
	}

	err = validator.ValidateProjectPlan(ctx, repoDir)

	Ok(t, err)
}

func TestApplyPlanValidator_CommentApplyStillFailsWhenDBPullStatusMissing(t *testing.T) {
	db := newTestBoltDB(t)
	repoDir := t.TempDir()
	ctx := command.ProjectContext{
		Log:         logging.NewNoopLogger(t),
		CommandName: command.Apply,
		Workspace:   "default",
		RepoRelDir:  ".",
		ProjectName: "projA",
		Pull: models.PullRequest{
			Num:        1,
			HeadCommit: "abc123",
			BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
		},
	}
	ctx.PullStatus = &models.PullStatus{
		Pull: ctx.Pull,
		Projects: []models.ProjectStatus{{
			Workspace:   ctx.Workspace,
			RepoRelDir:  ctx.RepoRelDir,
			ProjectName: ctx.ProjectName,
			Status:      models.PlannedPlanStatus,
		}},
	}
	planPath := filepath.Join(repoDir, runtime.GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	Ok(t, os.WriteFile(planPath, []byte("comment plan"), 0600))
	validator := &events.DefaultApplyPlanValidator{PullStatusFetcher: db}

	err := validator.ValidateProjectPlan(ctx, repoDir)

	Assert(t, err != nil, "expected comment apply to require DB pull status")
	Assert(t, strings.Contains(err.Error(), "no current plan status found"), "got: %s", err)
}

func TestApplyPlanValidator_DriftApplyUsesInMemoryPullStatusWithoutLiveHeadFetch(t *testing.T) {
	db := newTestBoltDB(t)
	repoDir := t.TempDir()
	planContents := []byte("drift plan")
	ctx := command.ProjectContext{
		Log:              logging.NewNoopLogger(t),
		CommandName:      command.Apply,
		API:              true,
		Workspace:        "default",
		RepoRelDir:       ".",
		ProjectName:      "projA",
		ExpectedPlanHash: planHashForContent(planContents),
		Pull: models.PullRequest{
			Num:        -1,
			HeadCommit: "abc123",
			BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
		},
	}
	ctx.PullStatus = &models.PullStatus{
		Pull: ctx.Pull,
		Projects: []models.ProjectStatus{{
			Workspace:   ctx.Workspace,
			RepoRelDir:  ctx.RepoRelDir,
			ProjectName: ctx.ProjectName,
			Status:      models.PlannedPlanStatus,
		}},
	}
	planPath := filepath.Join(repoDir, runtime.GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	Ok(t, os.WriteFile(planPath, planContents, 0600))
	validator := &events.DefaultApplyPlanValidator{
		PullStatusFetcher:   db,
		LivePullHeadFetcher: fakeLivePullHeadFetcher{err: errors.New("live head should not be fetched")},
	}

	err := validator.ValidateProjectPlan(ctx, repoDir)

	Ok(t, err)
}

func TestApplyPlanValidator_RejectsPlanPathOutsideProjectDir(t *testing.T) {
	db := newTestBoltDB(t)
	repoDir := t.TempDir()
	ctx := command.ProjectContext{
		Log:              logging.NewNoopLogger(t),
		CommandName:      command.Apply,
		Workspace:        "../escape",
		RepoRelDir:       ".",
		ProjectName:      "",
		ExpectedPlanHash: planHashForContent([]byte("plan")),
		Pull: models.PullRequest{
			Num:        1,
			HeadCommit: "abc123",
			BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
		},
	}
	validator := &events.DefaultApplyPlanValidator{PullStatusFetcher: db}

	err := validator.ValidateProjectPlan(ctx, repoDir)

	Assert(t, err != nil, "expected plan path traversal error")
	Assert(t, strings.Contains(err.Error(), "plan path traversal detected"), "got: %s", err)
}

func TestApplyPlanValidator_RejectsTraversalPlanPathBeforeHashing(t *testing.T) {
	db := newTestBoltDB(t)
	repoDir := t.TempDir()
	ctx := command.ProjectContext{
		Log:              logging.NewNoopLogger(t),
		CommandName:      command.Apply,
		Workspace:        "../escape",
		RepoRelDir:       ".",
		ProjectName:      "",
		ExpectedPlanHash: planHashForContent([]byte("plan")),
		Pull: models.PullRequest{
			Num:        1,
			HeadCommit: "abc123",
			BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
		},
	}
	validator := &events.DefaultApplyPlanValidator{PullStatusFetcher: db}

	err := validator.ValidateProjectPlan(ctx, repoDir)

	Assert(t, err != nil, "expected traversal to fail before hashing")
	Assert(t, strings.Contains(err.Error(), "plan path traversal detected"), "got: %s", err)
}

func TestApplyPlanValidator_HashesOnlyContainedPlanPath(t *testing.T) {
	db := newTestBoltDB(t)
	repoDir := t.TempDir()
	planContents := []byte("plan")
	ctx := command.ProjectContext{
		Log:              logging.NewNoopLogger(t),
		CommandName:      command.Apply,
		Workspace:        "default",
		RepoRelDir:       ".",
		ProjectName:      "projA",
		ExpectedPlanHash: planHashForContent(planContents),
		Pull: models.PullRequest{
			Num:        1,
			HeadCommit: "abc123",
			BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
		},
	}
	_, err := db.UpdatePullWithResults(ctx.Pull, []command.ProjectResult{plannedProjectResult(ctx.RepoRelDir, ctx.Workspace, ctx.ProjectName)})
	Ok(t, err)
	planPath := filepath.Join(repoDir, runtime.GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	Ok(t, os.WriteFile(planPath, planContents, 0600))
	validator := &events.DefaultApplyPlanValidator{PullStatusFetcher: db}

	err = validator.ValidateProjectPlan(ctx, repoDir)

	Ok(t, err)
}

func TestApplyPlanValidator_HashMismatchDoesNotDeleteNewerCurrentPlan(t *testing.T) {
	db := newTestBoltDB(t)
	repoDir := t.TempDir()
	ctx := command.ProjectContext{
		Log:              logging.NewNoopLogger(t),
		CommandName:      command.Apply,
		Workspace:        "default",
		RepoRelDir:       ".",
		ProjectName:      "projA",
		ExpectedPlanHash: planHashForContent([]byte("old plan")),
		Pull: models.PullRequest{
			Num:        1,
			HeadCommit: "abc123",
			BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
		},
	}
	_, err := db.UpdatePullWithResults(ctx.Pull, []command.ProjectResult{plannedProjectResult(ctx.RepoRelDir, ctx.Workspace, ctx.ProjectName)})
	Ok(t, err)
	planPath := filepath.Join(repoDir, runtime.GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	Ok(t, os.WriteFile(planPath, []byte("new plan"), 0600))
	validator := &events.DefaultApplyPlanValidator{PullStatusFetcher: db}

	err = validator.ValidateProjectPlan(ctx, repoDir)

	Assert(t, err != nil, "expected plan hash mismatch")
	Assert(t, strings.Contains(err.Error(), "plan file changed"), "got: %s", err)
	contents, readErr := os.ReadFile(planPath)
	Ok(t, readErr)
	Equals(t, "new plan", string(contents))
}

func TestProjectCommandRunner_ApplyRejectsPlanDeletedAfterBuilderValidation(t *testing.T) {
	RegisterMockTestingT(t)
	mockApply := mocks.NewMockStepRunner()
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockLocker := mocks.NewMockProjectLocker()
	db := newTestBoltDB(t)
	runner := &events.DefaultProjectCommandRunner{
		Locker:                    mockLocker,
		LockURLGenerator:          mockURLGenerator{},
		ApplyStepRunner:           mockApply,
		WorkingDir:                mockWorkingDir,
		WorkingDirLocker:          events.NewDefaultWorkingDirLocker(),
		CommandRequirementHandler: &events.DefaultCommandRequirementHandler{WorkingDir: mockWorkingDir},
		ApplyPlanValidator:        &events.DefaultApplyPlanValidator{PullStatusFetcher: db},
	}
	repoDir := t.TempDir()
	ctx := command.ProjectContext{
		Log:         logging.NewNoopLogger(t),
		CommandName: command.Apply,
		Steps:       valid.DefaultApplyStage.Steps,
		Workspace:   "default",
		RepoRelDir:  ".",
		ProjectName: "projA",
		Pull: models.PullRequest{
			Num:        1,
			HeadCommit: "abc123",
			BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
		},
		PullStatus: &models.PullStatus{
			Pull: models.PullRequest{HeadCommit: "abc123"},
			Projects: []models.ProjectStatus{
				{Workspace: "default", RepoRelDir: ".", ProjectName: "projA", Status: models.PlannedPlanStatus},
			},
		},
	}
	_, err := db.UpdatePullWithResults(ctx.Pull, []command.ProjectResult{
		{
			Command:     command.Plan,
			Workspace:   ctx.Workspace,
			RepoRelDir:  ctx.RepoRelDir,
			ProjectName: ctx.ProjectName,
			ProjectCommandOutput: command.ProjectCommandOutput{
				PlanSuccess: &models.PlanSuccess{},
			},
		},
	})
	Ok(t, err)
	When(mockWorkingDir.GetWorkingDir(ctx.Pull.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(repoDir, nil)
	When(mockLocker.TryLock(Any[logging.SimpleLogging](), Eq(ctx.Pull), Any[models.User](), Eq(ctx.Workspace), Any[models.Project](), AnyBool())).
		ThenReturn(&events.TryLockResponse{LockAcquired: true, LockKey: "lock-key"}, nil)

	res := runner.Apply(ctx)

	Assert(t, res.Error != nil, "expected missing plan file error")
	Assert(t, strings.Contains(res.Error.Error(), "plan file is missing"), "got: %s", res.Error)
	mockApply.VerifyWasCalled(Never()).Run(Any[command.ProjectContext](), Any[[]string](), Any[string](), Any[map[string]string]())
}

func TestProjectCommandRunner_ApplyDoesNotRunTerraformWhenLiveHeadChangedAfterCommandStart(t *testing.T) {
	RegisterMockTestingT(t)
	mockApply := mocks.NewMockStepRunner()
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockLocker := mocks.NewMockProjectLocker()
	db := newTestBoltDB(t)
	oldHead := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	newHead := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	runner := &events.DefaultProjectCommandRunner{
		Locker:                    mockLocker,
		LockURLGenerator:          mockURLGenerator{},
		ApplyStepRunner:           mockApply,
		WorkingDir:                mockWorkingDir,
		WorkingDirLocker:          events.NewDefaultWorkingDirLocker(),
		CommandRequirementHandler: &events.DefaultCommandRequirementHandler{WorkingDir: mockWorkingDir},
		ApplyPlanValidator:        &events.DefaultApplyPlanValidator{PullStatusFetcher: db, LivePullHeadFetcher: fakeLivePullHeadFetcher{head: newHead}},
	}
	repoDir := t.TempDir()
	ctx := command.ProjectContext{
		Log:         logging.NewNoopLogger(t),
		CommandName: command.Apply,
		Steps:       valid.DefaultApplyStage.Steps,
		Workspace:   "default",
		RepoRelDir:  ".",
		ProjectName: "projA",
		Pull: models.PullRequest{
			Num:        1,
			HeadCommit: oldHead,
			BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
		},
	}
	_, err := db.UpdatePullWithResults(ctx.Pull, []command.ProjectResult{plannedProjectResult(ctx.RepoRelDir, ctx.Workspace, ctx.ProjectName)})
	Ok(t, err)
	planPath := filepath.Join(repoDir, runtime.GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	Ok(t, os.WriteFile(planPath, []byte("plan"), 0600))
	When(mockWorkingDir.GetWorkingDir(ctx.Pull.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(repoDir, nil)
	When(mockLocker.TryLock(Any[logging.SimpleLogging](), Eq(ctx.Pull), Any[models.User](), Eq(ctx.Workspace), Any[models.Project](), AnyBool())).
		ThenReturn(&events.TryLockResponse{LockAcquired: true, LockKey: "lock-key"}, nil)

	res := runner.Apply(ctx)

	Assert(t, res.Error != nil, "expected live head change error")
	Assert(t, strings.Contains(res.Error.Error(), "pull request head changed"), "got: %s", res.Error)
	_, err = os.Stat(planPath)
	Ok(t, err)
	mockApply.VerifyWasCalled(Never()).Run(Any[command.ProjectContext](), Any[[]string](), Any[string](), Any[map[string]string]())
}

func TestProjectCommandRunner_ApplyRejectsPlanMutatedByPreApplyRunStep(t *testing.T) {
	RegisterMockTestingT(t)
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockLocker := mocks.NewMockProjectLocker()
	db := newTestBoltDB(t)
	calls := []string{}
	runner := &events.DefaultProjectCommandRunner{
		Locker:                    mockLocker,
		LockURLGenerator:          mockURLGenerator{},
		ApplyStepRunner:           &recordingStepRunner{name: "apply", calls: &calls},
		WorkingDir:                mockWorkingDir,
		WorkingDirLocker:          events.NewDefaultWorkingDirLocker(),
		CommandRequirementHandler: &events.DefaultCommandRequirementHandler{WorkingDir: mockWorkingDir},
		ApplyPlanValidator:        &events.DefaultApplyPlanValidator{PullStatusFetcher: db},
	}
	repoDir := t.TempDir()
	ctx := command.ProjectContext{
		Log:         logging.NewNoopLogger(t),
		CommandName: command.Apply,
		Steps: []valid.Step{
			{StepName: "run"},
			{StepName: "apply"},
		},
		Workspace:   "default",
		RepoRelDir:  ".",
		ProjectName: "projA",
		Pull: models.PullRequest{
			Num:        1,
			HeadCommit: "abc123",
			BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
		},
		PullStatus: &models.PullStatus{
			Pull: models.PullRequest{HeadCommit: "abc123"},
			Projects: []models.ProjectStatus{
				{Workspace: "default", RepoRelDir: ".", ProjectName: "projA", Status: models.PlannedPlanStatus},
			},
		},
	}
	_, err := db.UpdatePullWithResults(ctx.Pull, []command.ProjectResult{plannedProjectResult(ctx.RepoRelDir, ctx.Workspace, ctx.ProjectName)})
	Ok(t, err)
	planPath := filepath.Join(repoDir, runtime.GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	Ok(t, os.WriteFile(planPath, []byte("plan"), 0600))
	runner.RunStepRunner = &mutatingCustomStepRunner{
		calls: &calls,
		mutate: func() error {
			return os.Remove(planPath)
		},
	}
	When(mockWorkingDir.GetWorkingDir(ctx.Pull.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(repoDir, nil)
	When(mockWorkingDir.GitReadLock(ctx.Pull.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(func() {})
	When(mockLocker.TryLock(Any[logging.SimpleLogging](), Eq(ctx.Pull), Any[models.User](), Eq(ctx.Workspace), Any[models.Project](), AnyBool())).
		ThenReturn(&events.TryLockResponse{LockAcquired: true, LockKey: "lock-key"}, nil)

	res := runner.Apply(ctx)

	Assert(t, res.Error != nil, "expected final missing plan error")
	Assert(t, strings.Contains(res.Error.Error(), "plan file is missing"), "got: %s", res.Error)
	Equals(t, []string{"run"}, calls)
}

func TestProjectCommandRunner_ApplyRejectsPlanOverwrittenByPreApplyRunStep(t *testing.T) {
	testProjectCommandRunnerRejectsPlanContentMutation(t, []byte("changed plan"), "plan file changed")
}

func TestProjectCommandRunner_ApplyRejectsPlanMutatedBeforeBuiltInApplyStep(t *testing.T) {
	testProjectCommandRunnerRejectsPlanContentMutation(t, []byte("changed plan"), "plan file changed")
}

func TestProjectCommandRunner_PreApplyRunStepPlanfileExposureIsDocumented(t *testing.T) {
	testProjectCommandRunnerRejectsPlanContentMutation(t, []byte("changed plan"), "plan file changed")
}

func TestProjectCommandRunner_ApplyRejectsPlanTruncatedByPreApplyRunStep(t *testing.T) {
	testProjectCommandRunnerRejectsPlanContentMutation(t, nil, "plan file changed")
}

func testProjectCommandRunnerRejectsPlanContentMutation(t *testing.T, newContents []byte, expectedErr string) {
	RegisterMockTestingT(t)
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockLocker := mocks.NewMockProjectLocker()
	db := newTestBoltDB(t)
	calls := []string{}
	runner := &events.DefaultProjectCommandRunner{
		Locker:                    mockLocker,
		LockURLGenerator:          mockURLGenerator{},
		ApplyStepRunner:           &recordingStepRunner{name: "apply", calls: &calls},
		WorkingDir:                mockWorkingDir,
		WorkingDirLocker:          events.NewDefaultWorkingDirLocker(),
		CommandRequirementHandler: &events.DefaultCommandRequirementHandler{WorkingDir: mockWorkingDir},
		ApplyPlanValidator:        &events.DefaultApplyPlanValidator{PullStatusFetcher: db},
	}
	repoDir := t.TempDir()
	ctx := command.ProjectContext{
		Log:         logging.NewNoopLogger(t),
		CommandName: command.Apply,
		Steps: []valid.Step{
			{StepName: "run"},
			{StepName: "apply"},
		},
		Workspace:   "default",
		RepoRelDir:  ".",
		ProjectName: "projA",
		Pull: models.PullRequest{
			Num:        1,
			HeadCommit: "abc123",
			BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
		},
	}
	_, err := db.UpdatePullWithResults(ctx.Pull, []command.ProjectResult{plannedProjectResult(ctx.RepoRelDir, ctx.Workspace, ctx.ProjectName)})
	Ok(t, err)
	planPath := filepath.Join(repoDir, runtime.GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	Ok(t, os.WriteFile(planPath, []byte("original plan"), 0600))
	runner.RunStepRunner = &mutatingCustomStepRunner{
		calls: &calls,
		mutate: func() error {
			return os.WriteFile(planPath, newContents, 0600)
		},
	}
	When(mockWorkingDir.GetWorkingDir(ctx.Pull.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(repoDir, nil)
	When(mockWorkingDir.GitReadLock(ctx.Pull.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(func() {})
	When(mockLocker.TryLock(Any[logging.SimpleLogging](), Eq(ctx.Pull), Any[models.User](), Eq(ctx.Workspace), Any[models.Project](), AnyBool())).
		ThenReturn(&events.TryLockResponse{LockAcquired: true, LockKey: "lock-key"}, nil)

	res := runner.Apply(ctx)

	Assert(t, res.Error != nil, "expected plan hash mismatch")
	Assert(t, strings.Contains(res.Error.Error(), expectedErr), "got: %s", res.Error)
	contents, err := os.ReadFile(planPath)
	Ok(t, err)
	Equals(t, string(newContents), string(contents))
	Equals(t, []string{"run"}, calls)
}

func TestProjectCommandRunner_ApplyUsesExpectedPlanHash(t *testing.T) {
	RegisterMockTestingT(t)
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockLocker := mocks.NewMockProjectLocker()
	db := newTestBoltDB(t)
	calls := []string{}
	runner := &events.DefaultProjectCommandRunner{
		Locker:                    mockLocker,
		LockURLGenerator:          mockURLGenerator{},
		ApplyStepRunner:           &recordingStepRunner{name: "apply", calls: &calls},
		WorkingDir:                mockWorkingDir,
		WorkingDirLocker:          events.NewDefaultWorkingDirLocker(),
		CommandRequirementHandler: &events.DefaultCommandRequirementHandler{WorkingDir: mockWorkingDir},
		ApplyPlanValidator:        &events.DefaultApplyPlanValidator{PullStatusFetcher: db},
	}
	repoDir := t.TempDir()
	planContents := []byte("original plan")
	ctx := command.ProjectContext{
		Log:              logging.NewNoopLogger(t),
		CommandName:      command.Apply,
		Steps:            valid.DefaultApplyStage.Steps,
		Workspace:        "default",
		RepoRelDir:       ".",
		ProjectName:      "projA",
		ExpectedPlanHash: planHashForContent(planContents),
		Pull: models.PullRequest{
			Num:        1,
			HeadCommit: "abc123",
			BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
		},
	}
	_, err := db.UpdatePullWithResults(ctx.Pull, []command.ProjectResult{plannedProjectResult(ctx.RepoRelDir, ctx.Workspace, ctx.ProjectName)})
	Ok(t, err)
	planPath := filepath.Join(repoDir, runtime.GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	Ok(t, os.WriteFile(planPath, planContents, 0600))
	When(mockWorkingDir.GetWorkingDir(ctx.Pull.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(repoDir, nil)
	When(mockWorkingDir.GitReadLock(ctx.Pull.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(func() {})
	When(mockLocker.TryLock(Any[logging.SimpleLogging](), Eq(ctx.Pull), Any[models.User](), Eq(ctx.Workspace), Any[models.Project](), AnyBool())).
		ThenReturn(&events.TryLockResponse{LockAcquired: true, LockKey: "lock-key"}, nil)

	res := runner.Apply(ctx)

	Ok(t, res.Error)
	Equals(t, []string{"apply"}, calls)
}

func TestProjectCommandRunner_RechecksLiveIdentityAfterApplyStep(t *testing.T) {
	res, _, calls := runProjectApplyWithBaseChangeAfterApplyStep(t)

	Assert(t, res.Error != nil, "expected post-apply live identity error")
	Assert(t, strings.Contains(res.Error.Error(), "pull request base branch changed"), "got: %s", res.Error)
	Equals(t, []string{"apply"}, calls)
}

func TestProjectCommandRunner_BaseRetargetDuringApplyFailsBeforeSuccessOutput(t *testing.T) {
	res, _, _ := runProjectApplyWithBaseChangeAfterApplyStep(t)

	Assert(t, res.Error != nil, "expected post-apply live identity error")
	Equals(t, "", res.ApplySuccess)
}

func TestProjectCommandRunner_BaseRetargetDuringApplyDoesNotRunSuccessWebhooks(t *testing.T) {
	_, mockSender, _ := runProjectApplyWithBaseChangeAfterApplyStep(t)

	_, results := mockSender.VerifyWasCalledOnce().Send(Any[logging.SimpleLogging](), Any[webhooks.ApplyResult]()).GetCapturedArguments()
	Assert(t, !results.Success, "expected stale post-apply webhook to be unsuccessful")
}

func TestProjectCommandRunner_UnchangedIdentityApplyStillSucceeds(t *testing.T) {
	RegisterMockTestingT(t)
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockLocker := mocks.NewMockProjectLocker()
	mockSender := mocks.NewMockWebhooksSender()
	initialPull := models.PullRequest{
		Num:        1,
		HeadCommit: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		BaseBranch: "main",
		BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
	}
	fetcher := &sequenceLivePullIdentityFetcher{identities: []models.PullRequest{initialPull, initialPull}}
	calls := []string{}
	runner := &events.DefaultProjectCommandRunner{
		Locker:                    mockLocker,
		LockURLGenerator:          mockURLGenerator{},
		ApplyStepRunner:           &recordingStepRunner{name: "apply", calls: &calls},
		WorkingDir:                mockWorkingDir,
		WorkingDirLocker:          events.NewDefaultWorkingDirLocker(),
		CommandRequirementHandler: &events.DefaultCommandRequirementHandler{WorkingDir: mockWorkingDir},
		ApplyPlanValidator:        &events.DefaultApplyPlanValidator{LivePullHeadFetcher: fetcher},
		Webhooks:                  mockSender,
	}
	repoDir := t.TempDir()
	ctx := command.ProjectContext{
		Log:              logging.NewNoopLogger(t),
		CommandName:      command.Apply,
		Steps:            valid.DefaultApplyStage.Steps,
		Workspace:        "default",
		RepoRelDir:       ".",
		ProjectName:      "projA",
		ExpectedPlanHash: "already-captured",
		Pull:             initialPull,
	}
	When(mockWorkingDir.GetWorkingDir(ctx.Pull.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(repoDir, nil)
	When(mockWorkingDir.GitReadLock(ctx.Pull.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(func() {})
	When(mockLocker.TryLock(Any[logging.SimpleLogging](), Eq(ctx.Pull), Any[models.User](), Eq(ctx.Workspace), Any[models.Project](), AnyBool())).
		ThenReturn(&events.TryLockResponse{LockAcquired: true, LockKey: "lock-key"}, nil)

	res := runner.Apply(ctx)

	Ok(t, res.Error)
	Equals(t, "apply", res.ApplySuccess)
	Equals(t, []string{"apply"}, calls)
	_, results := mockSender.VerifyWasCalledOnce().Send(Any[logging.SimpleLogging](), Any[webhooks.ApplyResult]()).GetCapturedArguments()
	Assert(t, results.Success, "expected unchanged identity apply webhook to be successful")
}

func runProjectApplyWithBaseChangeAfterApplyStep(t *testing.T) (command.ProjectCommandOutput, *mocks.MockWebhooksSender, []string) {
	t.Helper()
	RegisterMockTestingT(t)
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockLocker := mocks.NewMockProjectLocker()
	mockSender := mocks.NewMockWebhooksSender()
	initialPull := models.PullRequest{
		Num:        1,
		HeadCommit: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		BaseBranch: "main",
		BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
	}
	finalPull := initialPull
	finalPull.BaseBranch = "release"
	fetcher := &sequenceLivePullIdentityFetcher{identities: []models.PullRequest{initialPull, finalPull}}
	calls := []string{}
	runner := &events.DefaultProjectCommandRunner{
		Locker:                    mockLocker,
		LockURLGenerator:          mockURLGenerator{},
		ApplyStepRunner:           &recordingStepRunner{name: "apply", calls: &calls},
		WorkingDir:                mockWorkingDir,
		WorkingDirLocker:          events.NewDefaultWorkingDirLocker(),
		CommandRequirementHandler: &events.DefaultCommandRequirementHandler{WorkingDir: mockWorkingDir},
		ApplyPlanValidator:        &events.DefaultApplyPlanValidator{LivePullHeadFetcher: fetcher},
		Webhooks:                  mockSender,
	}
	repoDir := t.TempDir()
	ctx := command.ProjectContext{
		Log:              logging.NewNoopLogger(t),
		CommandName:      command.Apply,
		Steps:            valid.DefaultApplyStage.Steps,
		Workspace:        "default",
		RepoRelDir:       ".",
		ProjectName:      "projA",
		ExpectedPlanHash: "already-captured",
		Pull:             initialPull,
	}
	When(mockWorkingDir.GetWorkingDir(ctx.Pull.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(repoDir, nil)
	When(mockWorkingDir.GitReadLock(ctx.Pull.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(func() {})
	When(mockLocker.TryLock(Any[logging.SimpleLogging](), Eq(ctx.Pull), Any[models.User](), Eq(ctx.Workspace), Any[models.Project](), AnyBool())).
		ThenReturn(&events.TryLockResponse{LockAcquired: true, LockKey: "lock-key"}, nil)

	return runner.Apply(ctx), mockSender, calls
}

func TestProjectCommandRunner_ApplyRejectsFreshErroredPolicyCheckStatus(t *testing.T) {
	RegisterMockTestingT(t)
	mockApply := mocks.NewMockStepRunner()
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockLocker := mocks.NewMockProjectLocker()
	db := newTestBoltDB(t)
	runner := &events.DefaultProjectCommandRunner{
		Locker:                    mockLocker,
		LockURLGenerator:          mockURLGenerator{},
		ApplyStepRunner:           mockApply,
		WorkingDir:                mockWorkingDir,
		WorkingDirLocker:          events.NewDefaultWorkingDirLocker(),
		CommandRequirementHandler: &events.DefaultCommandRequirementHandler{WorkingDir: mockWorkingDir},
		ApplyPlanValidator:        &events.DefaultApplyPlanValidator{PullStatusFetcher: db},
	}
	repoDir := t.TempDir()
	ctx := command.ProjectContext{
		Log:         logging.NewNoopLogger(t),
		CommandName: command.Apply,
		Steps:       valid.DefaultApplyStage.Steps,
		Workspace:   "default",
		RepoRelDir:  ".",
		ProjectName: "projA",
		Pull: models.PullRequest{
			Num:        1,
			HeadCommit: "abc123",
			BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
		},
	}
	_, err := db.UpdatePullWithResults(ctx.Pull, []command.ProjectResult{plannedProjectResult(ctx.RepoRelDir, ctx.Workspace, ctx.ProjectName)})
	Ok(t, err)
	_, err = db.UpdatePullWithResults(ctx.Pull, []command.ProjectResult{erroredPolicyProjectResult(ctx)})
	Ok(t, err)
	planPath := filepath.Join(repoDir, runtime.GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	Ok(t, os.WriteFile(planPath, []byte("plan"), 0600))
	When(mockWorkingDir.GetWorkingDir(ctx.Pull.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(repoDir, nil)
	When(mockLocker.TryLock(Any[logging.SimpleLogging](), Eq(ctx.Pull), Any[models.User](), Eq(ctx.Workspace), Any[models.Project](), AnyBool())).
		ThenReturn(&events.TryLockResponse{LockAcquired: true, LockKey: "lock-key"}, nil)

	res := runner.Apply(ctx)

	Assert(t, res.Error != nil, "expected policy-check status error")
	Assert(t, strings.Contains(res.Error.Error(), "policy checks have errored"), "got: %s", res.Error)
	_, err = os.Stat(planPath)
	Assert(t, os.IsNotExist(err), "expected rejected plan file to be deleted")
	mockApply.VerifyWasCalled(Never()).Run(Any[command.ProjectContext](), Any[[]string](), Any[string](), Any[map[string]string]())
}

func TestProjectCommandRunner_ApplyDoesNotRunTerraformWhenPolicyStatusChangesAfterBuild(t *testing.T) {
	RegisterMockTestingT(t)
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockLocker := mocks.NewMockProjectLocker()
	db := newTestBoltDB(t)
	calls := []string{}
	runner := &events.DefaultProjectCommandRunner{
		Locker:                    mockLocker,
		LockURLGenerator:          mockURLGenerator{},
		ApplyStepRunner:           &recordingStepRunner{name: "apply", calls: &calls},
		WorkingDir:                mockWorkingDir,
		WorkingDirLocker:          events.NewDefaultWorkingDirLocker(),
		CommandRequirementHandler: &events.DefaultCommandRequirementHandler{WorkingDir: mockWorkingDir},
		ApplyPlanValidator:        &events.DefaultApplyPlanValidator{PullStatusFetcher: db},
	}
	repoDir := t.TempDir()
	ctx := command.ProjectContext{
		Log:         logging.NewNoopLogger(t),
		CommandName: command.Apply,
		Steps: []valid.Step{
			{StepName: "run"},
			{StepName: "apply"},
		},
		Workspace:   "default",
		RepoRelDir:  ".",
		ProjectName: "projA",
		Pull: models.PullRequest{
			Num:        1,
			HeadCommit: "abc123",
			BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
		},
		PullStatus: &models.PullStatus{
			Pull: models.PullRequest{HeadCommit: "abc123"},
			Projects: []models.ProjectStatus{
				{Workspace: "default", RepoRelDir: ".", ProjectName: "projA", Status: models.PlannedPlanStatus},
			},
		},
	}
	_, err := db.UpdatePullWithResults(ctx.Pull, []command.ProjectResult{plannedProjectResult(ctx.RepoRelDir, ctx.Workspace, ctx.ProjectName)})
	Ok(t, err)
	planPath := filepath.Join(repoDir, runtime.GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	Ok(t, os.WriteFile(planPath, []byte("plan"), 0600))
	runner.RunStepRunner = &mutatingCustomStepRunner{
		calls: &calls,
		mutate: func() error {
			_, err := db.UpdatePullWithResults(ctx.Pull, []command.ProjectResult{erroredPolicyProjectResult(ctx)})
			return err
		},
	}
	When(mockWorkingDir.GetWorkingDir(ctx.Pull.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(repoDir, nil)
	When(mockWorkingDir.GitReadLock(ctx.Pull.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(func() {})
	When(mockLocker.TryLock(Any[logging.SimpleLogging](), Eq(ctx.Pull), Any[models.User](), Eq(ctx.Workspace), Any[models.Project](), AnyBool())).
		ThenReturn(&events.TryLockResponse{LockAcquired: true, LockKey: "lock-key"}, nil)

	res := runner.Apply(ctx)

	Assert(t, res.Error != nil, "expected policy-check status error")
	Assert(t, strings.Contains(res.Error.Error(), "policy checks have errored"), "got: %s", res.Error)
	Equals(t, []string{"run"}, calls)
}

func TestProjectCommandRunner_ApplyRejectsStalePullStatusAfterBuilderValidation(t *testing.T) {
	RegisterMockTestingT(t)
	mockApply := mocks.NewMockStepRunner()
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockLocker := mocks.NewMockProjectLocker()
	db := newTestBoltDB(t)
	runner := &events.DefaultProjectCommandRunner{
		Locker:                    mockLocker,
		LockURLGenerator:          mockURLGenerator{},
		ApplyStepRunner:           mockApply,
		WorkingDir:                mockWorkingDir,
		WorkingDirLocker:          events.NewDefaultWorkingDirLocker(),
		CommandRequirementHandler: &events.DefaultCommandRequirementHandler{WorkingDir: mockWorkingDir},
		ApplyPlanValidator:        &events.DefaultApplyPlanValidator{PullStatusFetcher: db},
	}
	repoDir := t.TempDir()
	ctx := command.ProjectContext{
		Log:         logging.NewNoopLogger(t),
		CommandName: command.Apply,
		Steps:       valid.DefaultApplyStage.Steps,
		Workspace:   "default",
		RepoRelDir:  ".",
		ProjectName: "projA",
		Pull: models.PullRequest{
			Num:        1,
			HeadCommit: "new123",
			BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
		},
		PullStatus: &models.PullStatus{
			Pull: models.PullRequest{HeadCommit: "new123"},
			Projects: []models.ProjectStatus{
				{Workspace: "default", RepoRelDir: ".", ProjectName: "projA", Status: models.PlannedPlanStatus},
			},
		},
	}
	stalePull := ctx.Pull
	stalePull.HeadCommit = "old123"
	_, err := db.UpdatePullWithResults(stalePull, []command.ProjectResult{
		{
			Command:     command.Plan,
			Workspace:   ctx.Workspace,
			RepoRelDir:  ctx.RepoRelDir,
			ProjectName: ctx.ProjectName,
			ProjectCommandOutput: command.ProjectCommandOutput{
				PlanSuccess: &models.PlanSuccess{},
			},
		},
	})
	Ok(t, err)
	planPath := filepath.Join(repoDir, runtime.GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	Ok(t, os.WriteFile(planPath, []byte("plan"), 0600))
	When(mockWorkingDir.GetWorkingDir(ctx.Pull.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(repoDir, nil)
	When(mockLocker.TryLock(Any[logging.SimpleLogging](), Eq(ctx.Pull), Any[models.User](), Eq(ctx.Workspace), Any[models.Project](), AnyBool())).
		ThenReturn(&events.TryLockResponse{LockAcquired: true, LockKey: "lock-key"}, nil)

	res := runner.Apply(ctx)

	Assert(t, res.Error != nil, "expected stale PullStatus error")
	Assert(t, strings.Contains(res.Error.Error(), "recorded plan status is from commit"), "got: %s", res.Error)
	_, err = os.Stat(planPath)
	Assert(t, os.IsNotExist(err), "expected stale plan file to be deleted")
	mockApply.VerifyWasCalled(Never()).Run(Any[command.ProjectContext](), Any[[]string](), Any[string](), Any[map[string]string]())
}

func TestProjectCommandRunner_ApplyValidationFailureDoesNotLaunderStalePlanAsErroredApply(t *testing.T) {
	RegisterMockTestingT(t)
	mockApply := mocks.NewMockStepRunner()
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockLocker := mocks.NewMockProjectLocker()
	db := newTestBoltDB(t)
	runner := &events.DefaultProjectCommandRunner{
		Locker:                    mockLocker,
		LockURLGenerator:          mockURLGenerator{},
		ApplyStepRunner:           mockApply,
		WorkingDir:                mockWorkingDir,
		WorkingDirLocker:          events.NewDefaultWorkingDirLocker(),
		CommandRequirementHandler: &events.DefaultCommandRequirementHandler{WorkingDir: mockWorkingDir},
		ApplyPlanValidator:        &events.DefaultApplyPlanValidator{PullStatusFetcher: db},
	}
	repoDir := t.TempDir()
	ctx := command.ProjectContext{
		Log:         logging.NewNoopLogger(t),
		CommandName: command.Apply,
		Steps:       valid.DefaultApplyStage.Steps,
		Workspace:   "default",
		RepoRelDir:  ".",
		ProjectName: "projA",
		Pull: models.PullRequest{
			Num:        1,
			HeadCommit: "new123",
			BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
		},
	}
	stalePull := ctx.Pull
	stalePull.HeadCommit = "old123"
	_, err := db.UpdatePullWithResults(stalePull, []command.ProjectResult{plannedProjectResult(ctx.RepoRelDir, ctx.Workspace, ctx.ProjectName)})
	Ok(t, err)
	planPath := filepath.Join(repoDir, runtime.GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	Ok(t, os.WriteFile(planPath, []byte("plan"), 0600))
	When(mockWorkingDir.GetWorkingDir(ctx.Pull.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(repoDir, nil)
	When(mockLocker.TryLock(Any[logging.SimpleLogging](), Eq(ctx.Pull), Any[models.User](), Eq(ctx.Workspace), Any[models.Project](), AnyBool())).
		ThenReturn(&events.TryLockResponse{LockAcquired: true, LockKey: "lock-key"}, nil)

	res := runner.Apply(ctx)

	Assert(t, res.Error != nil, "expected stale PullStatus error")
	_, err = os.Stat(planPath)
	Assert(t, os.IsNotExist(err), "expected stale plan file to be deleted")
	_, err = db.UpdatePullWithResults(ctx.Pull, []command.ProjectResult{
		{
			Command:     command.Apply,
			Workspace:   ctx.Workspace,
			RepoRelDir:  ctx.RepoRelDir,
			ProjectName: ctx.ProjectName,
			ProjectCommandOutput: command.ProjectCommandOutput{
				Error: errors.New("validation rejected plan"),
			},
		},
	})
	Ok(t, err)
	pullStatus, err := db.GetPullStatus(ctx.Pull)
	Ok(t, err)
	err = events.ValidatePlansForApply(&command.Context{Log: logging.NewNoopLogger(t), Pull: ctx.Pull, PullStatus: pullStatus}, nil)
	Assert(t, err != nil, "expected later generic apply to reject errored apply status without plan file")
	Assert(t, strings.Contains(err.Error(), "plan file is missing"), "got: %s", err)
	mockApply.VerifyWasCalled(Never()).Run(Any[command.ProjectContext](), Any[[]string](), Any[string](), Any[map[string]string]())
}

type trackingWorkingDirLocker struct {
	locked bool
}

func planHashForContent(contents []byte) string {
	hash := sha256.Sum256(contents)
	return hex.EncodeToString(hash[:])
}

type fakeLivePullHeadFetcher struct {
	head string
	base string
	err  error
}

func (f fakeLivePullHeadFetcher) GetLivePullIdentity(command.ProjectContext) (models.PullRequest, error) {
	return models.PullRequest{HeadCommit: f.head, BaseBranch: f.base}, f.err
}

func (l *trackingWorkingDirLocker) TryLock(string, int, string, string, string, command.Name) (func(), error) {
	l.locked = true
	return func() { l.locked = false }, nil
}

func (l *trackingWorkingDirLocker) TryLockPull(string, int, command.Name) (func(), error) {
	l.locked = true
	return func() { l.locked = false }, nil
}

func (l *trackingWorkingDirLocker) HasCommandLock(string, int, command.Name) bool {
	return l.locked
}

func (l *trackingWorkingDirLocker) UnlockByPull(string, int) {
	l.locked = false
}

type assertLockedApplyPlanValidator struct {
	t      *testing.T
	locker *trackingWorkingDirLocker
	called bool
}

func (v *assertLockedApplyPlanValidator) ValidateProjectPlan(command.ProjectContext, string) error {
	v.called = true
	Assert(v.t, v.locker.locked, "expected apply plan validation to run while working dir lock is held")
	return nil
}

type sequencedApplyPlanValidator struct {
	t      *testing.T
	locker *trackingWorkingDirLocker
	calls  *[]string
	errs   []error
	count  int
}

func (v *sequencedApplyPlanValidator) ValidateProjectPlan(command.ProjectContext, string) error {
	if v.calls != nil {
		*v.calls = append(*v.calls, "validate")
	}
	if v.locker != nil {
		Assert(v.t, v.locker.locked, "expected apply plan validation to run while working dir lock is held")
	}
	defer func() { v.count++ }()
	if v.count < len(v.errs) {
		return v.errs[v.count]
	}
	return nil
}

type recordingStepRunner struct {
	name  string
	calls *[]string
	err   error
}

func (r *recordingStepRunner) Run(command.ProjectContext, []string, string, map[string]string) (string, error) {
	if r.calls != nil {
		*r.calls = append(*r.calls, r.name)
	}
	return r.name, r.err
}

type mutatingCustomStepRunner struct {
	calls  *[]string
	mutate func() error
}

func (r *mutatingCustomStepRunner) Run(
	command.ProjectContext,
	*valid.CommandShell,
	string,
	string,
	map[string]string,
	bool,
	[]valid.PostProcessRunOutputOption,
	[]*regexp.Regexp,
) (string, error) {
	if r.calls != nil {
		*r.calls = append(*r.calls, "run")
	}
	if r.mutate != nil {
		if err := r.mutate(); err != nil {
			return "", err
		}
	}
	return "run", nil
}

func erroredPolicyProjectResult(ctx command.ProjectContext) command.ProjectResult {
	return command.ProjectResult{
		Command:     command.PolicyCheck,
		Workspace:   ctx.Workspace,
		RepoRelDir:  ctx.RepoRelDir,
		ProjectName: ctx.ProjectName,
		ProjectCommandOutput: command.ProjectCommandOutput{
			Error: errors.New("policy check errored"),
		},
	}
}

func newTestBoltDB(t *testing.T) *boltdb.BoltDB {
	t.Helper()
	db, err := boltdb.New(t.TempDir())
	Ok(t, err)
	t.Cleanup(func() {
		_ = db.Close()
	})
	return db
}

func nonPRAPIApplyContext(t *testing.T, headCommit string) command.ProjectContext {
	t.Helper()
	return command.ProjectContext{
		Log:         logging.NewNoopLogger(t),
		CommandName: command.Apply,
		API:         true,
		Steps:       []valid.Step{{StepName: "apply"}},
		Workspace:   "default",
		RepoRelDir:  ".",
		ProjectName: "projA",
		Pull: models.PullRequest{
			Num:        -1,
			HeadBranch: "main",
			HeadCommit: headCommit,
			BaseRepo:   models.Repo{FullName: "runatlantis/atlantis"},
		},
	}
}

func newNonPRAPIApplyRunner(t *testing.T, repoDir string) (*events.DefaultProjectCommandRunner, *mocks.MockStepRunner) {
	t.Helper()
	RegisterMockTestingT(t)
	mockApply := mocks.NewMockStepRunner()
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockLocker := mocks.NewMockProjectLocker()
	mockRequirementHandler := mocks.NewMockCommandRequirementHandler()
	When(mockWorkingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(repoDir, nil)
	When(mockWorkingDir.GitReadLock(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(func() {})
	When(mockLocker.TryLock(Any[logging.SimpleLogging](), Any[models.PullRequest](), Any[models.User](), Any[string](), Any[models.Project](), AnyBool())).
		ThenReturn(&events.TryLockResponse{LockAcquired: true, LockKey: "lock-key"}, nil)
	When(mockRequirementHandler.ValidateApplyProject(Any[string](), Any[command.ProjectContext]())).ThenReturn("", nil)
	When(mockRequirementHandler.ValidateProjectDependencies(Any[command.ProjectContext]())).ThenReturn("", nil)
	When(mockApply.Run(Any[command.ProjectContext](), Any[[]string](), Any[string](), Any[map[string]string]())).ThenReturn("applied", nil)
	return &events.DefaultProjectCommandRunner{
		Locker:                    mockLocker,
		ApplyStepRunner:           mockApply,
		WorkingDir:                mockWorkingDir,
		WorkingDirLocker:          events.NewDefaultWorkingDirLocker(),
		CommandRequirementHandler: mockRequirementHandler,
	}, mockApply
}

func initProjectRunnerAPIRefGitRepo(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	originDir := filepath.Join(root, "origin.git")
	repoDir := filepath.Join(root, "work")
	runProjectRunnerAPIRefGit(t, "", "init", "--bare", originDir)
	runProjectRunnerAPIRefGit(t, originDir, "config", "gc.auto", "0")
	runProjectRunnerAPIRefGit(t, originDir, "config", "maintenance.auto", "false")
	runProjectRunnerAPIRefGit(t, originDir, "config", "receive.autogc", "false")
	runProjectRunnerAPIRefGit(t, "", "init", "--initial-branch=main", repoDir)
	runProjectRunnerAPIRefGit(t, repoDir, "config", "gc.auto", "0")
	runProjectRunnerAPIRefGit(t, repoDir, "config", "maintenance.auto", "false")
	runProjectRunnerAPIRefGit(t, repoDir, "config", "user.email", "atlantisbot@runatlantis.io")
	runProjectRunnerAPIRefGit(t, repoDir, "config", "user.name", "atlantisbot")
	runProjectRunnerAPIRefGit(t, repoDir, "config", "commit.gpgsign", "false")
	Ok(t, os.WriteFile(filepath.Join(repoDir, "main.tf"), []byte("resource \"null_resource\" \"main\" {}\n"), 0600))
	runProjectRunnerAPIRefGit(t, repoDir, "add", "main.tf")
	runProjectRunnerAPIRefGit(t, repoDir, "commit", "-q", "-m", "initial")
	initialCommit := strings.TrimSpace(runProjectRunnerAPIRefGit(t, repoDir, "rev-parse", "HEAD"))
	runProjectRunnerAPIRefGit(t, repoDir, "remote", "add", "origin", "file://"+originDir)
	runProjectRunnerAPIRefGit(t, repoDir, "push", "-q", "-u", "origin", "main")
	return repoDir, initialCommit
}

func advanceProjectRunnerAPIRefGitMain(t *testing.T, repoDir string) string {
	t.Helper()
	runProjectRunnerAPIRefGit(t, repoDir, "checkout", "-q", "main")
	Ok(t, os.WriteFile(filepath.Join(repoDir, "main.tf"), []byte("resource \"null_resource\" \"changed\" {}\n"), 0600))
	runProjectRunnerAPIRefGit(t, repoDir, "add", "main.tf")
	runProjectRunnerAPIRefGit(t, repoDir, "commit", "-q", "-m", "advance main")
	runProjectRunnerAPIRefGit(t, repoDir, "push", "-q", "origin", "HEAD:main")
	return strings.TrimSpace(runProjectRunnerAPIRefGit(t, repoDir, "rev-parse", "HEAD"))
}

func createProjectRunnerAPIRefGitBranch(t *testing.T, repoDir string, branch string, commit string) {
	t.Helper()
	runProjectRunnerAPIRefGit(t, repoDir, "checkout", "-q", "-B", branch, commit)
	runProjectRunnerAPIRefGit(t, repoDir, "push", "-q", "-u", "origin", "HEAD:"+branch)
	runProjectRunnerAPIRefGit(t, repoDir, "checkout", "-q", "main")
}

func advanceProjectRunnerAPIRefGitBranch(t *testing.T, repoDir string, branch string) string {
	t.Helper()
	runProjectRunnerAPIRefGit(t, repoDir, "checkout", "-q", branch)
	Ok(t, os.WriteFile(filepath.Join(repoDir, "main.tf"), []byte("resource \"null_resource\" \""+branch+"\" {}\n"), 0600))
	runProjectRunnerAPIRefGit(t, repoDir, "add", "main.tf")
	runProjectRunnerAPIRefGit(t, repoDir, "commit", "-q", "-m", "advance "+branch)
	runProjectRunnerAPIRefGit(t, repoDir, "push", "-q", "origin", "HEAD:"+branch)
	return strings.TrimSpace(runProjectRunnerAPIRefGit(t, repoDir, "rev-parse", "HEAD"))
}

func runProjectRunnerAPIRefGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	args = append([]string{
		"-c", "protocol.file.allow=always",
		"-c", "gc.auto=0",
		"-c", "maintenance.auto=false",
		"-c", "receive.autogc=false",
	}, args...)
	cmd := exec.Command("git", args...) //nolint:gosec
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null", "GIT_TERMINAL_PROMPT=0")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("running git %s: %s: %v", strings.Join(args, " "), output, err)
	}
	return string(output)
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
			When(mockWorkingDir.GitReadLock(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(func() {})
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
				Log:               logging.NewNoopLogger(t),
				Steps:             c.steps,
				Workspace:         "default",
				ApplyRequirements: c.applyReqs,
				RepoRelDir:        ".",
				PullReqStatus: models.PullReqStatus{
					ApprovalStatus: models.ApprovalStatus{
						IsApproved: true,
					},
					MergeableStatus: models.MergeableStatus{IsMergeable: false},
				},
			}
			expEnvs := map[string]string{
				"key": "value",
			}
			When(mockInit.Run(ctx, nil, repoDir, expEnvs)).ThenReturn("init", nil)
			When(mockPlan.Run(ctx, nil, repoDir, expEnvs)).ThenReturn("plan", nil)
			When(mockApply.Run(ctx, nil, repoDir, expEnvs)).ThenReturn("apply", nil)
			When(mockRun.Run(ctx, nil, "", repoDir, expEnvs, true, nil, nil)).ThenReturn("run", nil)
			When(mockEnv.Run(ctx, nil, "", "value", repoDir, make(map[string]string))).ThenReturn("value", nil)

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
					mockRun.VerifyWasCalledOnce().Run(ctx, nil, "", repoDir, expEnvs, true, nil, nil)
				case "env":
					mockEnv.VerifyWasCalledOnce().Run(ctx, nil, "", "value", repoDir, expEnvs)
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
	When(mockWorkingDir.GitReadLock(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(func() {})
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

func TestDefaultProjectCommandRunner_ApplySuppressesApplyWebhooks(t *testing.T) {
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
	When(mockWorkingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(repoDir, nil)
	When(mockWorkingDir.GitReadLock(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(func() {})
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
		Log:                   logging.NewNoopLogger(t),
		Steps:                 []valid.Step{{StepName: "apply"}},
		Workspace:             "default",
		RepoRelDir:            ".",
		SuppressApplyWebhooks: true,
	}
	expEnvs := map[string]string{}
	When(mockApply.Run(ctx, nil, repoDir, expEnvs)).ThenReturn("apply", nil)

	res := runner.Apply(ctx)

	Equals(t, "apply", res.ApplySuccess)
	mockApply.VerifyWasCalledOnce().Run(ctx, nil, repoDir, expEnvs)
	mockSender.VerifyWasCalled(Never()).Send(Any[logging.SimpleLogging](), Any[webhooks.ApplyResult]())
}

// Test run and env steps. We don't use mocks for this test since we're
// not running any Terraform.
func TestDefaultProjectCommandRunner_RunEnvSteps(t *testing.T) {
	RegisterMockTestingT(t)
	tfClient := tfclientmocks.NewMockClient()
	tfDistribution := terraform.NewDistributionTerraformWithDownloader(tmocks.NewMockDownloader())
	tfVersion, err := version.NewVersion("0.12.0")
	Ok(t, err)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	run := runtime.RunStepRunner{
		TerraformExecutor:       tfClient,
		DefaultTFDistribution:   tfDistribution,
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
	When(mockWorkingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
		Any[string]())).ThenReturn(repoDir, nil)
	When(mockWorkingDir.GitReadLock(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(func() {})
	When(mockLocker.TryLock(Any[logging.SimpleLogging](), Any[models.PullRequest](), Any[models.User](), Any[string](),
		Any[models.Project](), AnyBool())).ThenReturn(&events.TryLockResponse{LockAcquired: true, LockKey: "lock-key"}, nil)

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
			When(mockWorkingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
				Any[string]())).ThenReturn(repoDir, nil)
			When(mockWorkingDir.GitReadLock(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(func() {})
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

// Test that custom policy checks use configured policy set names instead of defaulting to "Custom".
// This is a regression test for https://github.com/runatlantis/atlantis/pull/5331
// where custom policy sets defaulting to "Custom" allowed any user to approve policies.
func TestDefaultProjectCommandRunner_CustomPolicyCheckNames(t *testing.T) {
	RegisterMockTestingT(t)

	cases := []struct {
		description       string
		customPolicyCheck bool
		policySets        []valid.PolicySet
		policyOutputs     []string
		expectedNames     []string
	}{
		{
			description:       "Custom policy check with single named policy set",
			customPolicyCheck: true,
			policySets: []valid.PolicySet{
				{
					Name:         "security_policy",
					ApproveCount: 1,
					Owners: valid.PolicyOwners{
						Users: []string{"security-team"},
					},
				},
			},
			policyOutputs: []string{"Policy check passed"},
			expectedNames: []string{"security_policy"},
		},
		{
			description:       "Custom policy check with multiple named policy sets",
			customPolicyCheck: true,
			policySets: []valid.PolicySet{
				{
					Name:         "security_policy",
					ApproveCount: 1,
					Owners: valid.PolicyOwners{
						Users: []string{"security-team"},
					},
				},
				{
					Name:         "compliance_policy",
					ApproveCount: 2,
					Owners: valid.PolicyOwners{
						Users: []string{"compliance-team"},
					},
				},
			},
			policyOutputs: []string{"Security check passed", "Compliance check FAIL"},
			expectedNames: []string{"security_policy", "compliance_policy"},
		},
		{
			description:       "Custom policy check defaults to 'Custom' when no policy set configured",
			customPolicyCheck: true,
			policySets:        []valid.PolicySet{},
			policyOutputs:     []string{"Policy check passed"},
			expectedNames:     []string{"Custom"},
		},
		{
			description:       "More outputs than policy sets - excess use 'Custom'",
			customPolicyCheck: true,
			policySets: []valid.PolicySet{
				{
					Name:         "security_policy",
					ApproveCount: 1,
					Owners: valid.PolicyOwners{
						Users: []string{"security-team"},
					},
				},
			},
			policyOutputs: []string{"Security check passed", "Extra check passed"},
			expectedNames: []string{"security_policy", "Custom"},
		},
		{
			description:       "More policy sets than outputs - only received outputs processed",
			customPolicyCheck: true,
			policySets: []valid.PolicySet{
				{
					Name:         "security_policy",
					ApproveCount: 1,
					Owners: valid.PolicyOwners{
						Users: []string{"security-team"},
					},
				},
				{
					Name:         "compliance_policy",
					ApproveCount: 1,
					Owners: valid.PolicyOwners{
						Users: []string{"compliance-team"},
					},
				},
				{
					Name:         "audit_policy",
					ApproveCount: 1,
					Owners: valid.PolicyOwners{
						Users: []string{"audit-team"},
					},
				},
			},
			policyOutputs: []string{"Security check passed"},
			expectedNames: []string{"security_policy"},
		},
		{
			description:       "Empty output is preserved and marked as failed",
			customPolicyCheck: true,
			policySets: []valid.PolicySet{
				{
					Name:         "security_policy",
					ApproveCount: 1,
					Owners: valid.PolicyOwners{
						Users: []string{"security-team"},
					},
				},
				{
					Name:         "compliance_policy",
					ApproveCount: 1,
					Owners: valid.PolicyOwners{
						Users: []string{"compliance-team"},
					},
				},
			},
			policyOutputs: []string{"Security check passed", ""},
			expectedNames: []string{"security_policy", "compliance_policy"},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			mockPolicyCheck := mocks.NewMockStepRunner()
			mockWorkingDir := mocks.NewMockWorkingDir()
			mockLocker := mocks.NewMockProjectLocker()

			runner := events.DefaultProjectCommandRunner{
				Locker:                mockLocker,
				LockURLGenerator:      mockURLGenerator{},
				PolicyCheckStepRunner: mockPolicyCheck,
				WorkingDir:            mockWorkingDir,
				WorkingDirLocker:      events.NewDefaultWorkingDirLocker(),
			}

			repoDir := t.TempDir()
			When(mockWorkingDir.GetWorkingDir(
				Any[models.Repo](),
				Any[models.PullRequest](),
				Any[string](),
			)).ThenReturn(repoDir, nil)
			When(mockWorkingDir.GitReadLock(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(func() {})

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

			// Setup policy check steps - one step per policy output
			var steps []valid.Step
			for range c.policyOutputs {
				steps = append(steps, valid.Step{
					StepName: "policy_check",
				})
			}

			// Setup mock to return outputs in sequence
			// Note: pegomock will return these in order for successive calls
			for _, output := range c.policyOutputs {
				When(mockPolicyCheck.Run(
					Any[command.ProjectContext](),
					Any[[]string](),
					Any[string](),
					Any[map[string]string](),
				)).ThenReturn(output, nil)
			}

			ctx := command.ProjectContext{
				Log:               logging.NewNoopLogger(t),
				Workspace:         "default",
				RepoRelDir:        ".",
				CustomPolicyCheck: c.customPolicyCheck,
				PolicySets: valid.PolicySets{
					PolicySets: c.policySets,
				},
				Steps: steps,
			}

			res := runner.PolicyCheck(ctx)

			Assert(t, res.Error == nil, "not expecting error: %v", res.Error)
			Assert(t, res.PolicyCheckResults != nil, "expecting policy check results")

			// Verify that the policy set names match the configured names
			policyResults := res.PolicyCheckResults.PolicySetResults
			Equals(t, len(c.expectedNames), len(policyResults))

			for i, expectedName := range c.expectedNames {
				Equals(t, expectedName, policyResults[i].PolicySetName)
			}
		})
	}
}

// Test that when custom policy check has configured policy sets but no outputs are generated,
// it does NOT trigger "unable to unmarshal conftest output" error.
// This test reproduces the bug where policySetResults remains nil when the outputs
// array is empty, which would incorrectly trigger the nil check error.
func TestDefaultProjectCommandRunner_CustomPolicyCheck_EmptyOutputsArray(t *testing.T) {
	RegisterMockTestingT(t)

	cases := []struct {
		description       string
		customPolicyCheck bool
		policySets        []valid.PolicySet
		steps             []valid.Step
		expectError       bool
		expectedErrorMsg  string
	}{
		{
			description:       "Custom policy check with configured policy set but no steps (empty outputs array)",
			customPolicyCheck: true,
			policySets: []valid.PolicySet{
				{
					Name:         "test_policy",
					ApproveCount: 1,
					Owners: valid.PolicyOwners{
						Users: []string{"test-user"},
					},
				},
			},
			steps:            []valid.Step{}, // No steps - outputs array will be empty
			expectError:      true,           // Should error when policy sets configured but no results
			expectedErrorMsg: "custom policy check produced no results despite configured policy sets",
		},
		{
			description:       "Custom policy check with no configured policy sets and no steps",
			customPolicyCheck: true,
			policySets:        []valid.PolicySet{}, // No policy sets configured
			steps:             []valid.Step{},      // No steps
			expectError:       false,               // Should NOT error when no policy sets configured
			expectedErrorMsg:  "",
		},
		{
			description:       "Non-custom (conftest) policy check with no steps",
			customPolicyCheck: false,
			policySets:        []valid.PolicySet{},
			steps:             []valid.Step{},
			expectError:       true,
			expectedErrorMsg:  "unable to unmarshal conftest output",
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			mockPolicyCheck := mocks.NewMockStepRunner()
			mockWorkingDir := mocks.NewMockWorkingDir()
			mockLocker := mocks.NewMockProjectLocker()

			runner := events.DefaultProjectCommandRunner{
				Locker:                mockLocker,
				LockURLGenerator:      mockURLGenerator{},
				PolicyCheckStepRunner: mockPolicyCheck,
				WorkingDir:            mockWorkingDir,
				WorkingDirLocker:      events.NewDefaultWorkingDirLocker(),
			}

			repoDir := t.TempDir()
			When(mockWorkingDir.GetWorkingDir(
				Any[models.Repo](),
				Any[models.PullRequest](),
				Any[string](),
			)).ThenReturn(repoDir, nil)
			When(mockWorkingDir.GitReadLock(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(func() {})

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
				UnlockFn:     func() error { return nil },
			}, nil)

			ctx := command.ProjectContext{
				Log:               logging.NewNoopLogger(t),
				Workspace:         "default",
				RepoRelDir:        ".",
				CustomPolicyCheck: c.customPolicyCheck,
				PolicySets: valid.PolicySets{
					PolicySets: c.policySets,
				},
				Steps: c.steps,
			}

			res := runner.PolicyCheck(ctx)

			if c.expectError {
				Assert(t, res.Error != nil, "expecting error but got nil")
				if c.expectedErrorMsg != "" {
					Assert(t, res.Error.Error() == c.expectedErrorMsg,
						"expected error message '%s' but got '%s'",
						c.expectedErrorMsg, res.Error.Error())
				}
			} else {
				Assert(t, res.Error == nil, "not expecting error but got: %v", res.Error)
				Assert(t, res.PolicyCheckResults != nil, "expecting policy check results")
			}
		})
	}
}

// Test custom policy check failure detection logic (regex and FAIL prefix).
func TestDefaultProjectCommandRunner_CustomPolicyCheckFailureDetection(t *testing.T) {
	RegisterMockTestingT(t)

	cases := []struct {
		description    string
		policyOutput   string
		expectedPassed bool
		expectedOutput string
	}{
		{
			description:    "Output with '1 failure' pattern should fail",
			policyOutput:   "Policy check found 1 failure in the code",
			expectedPassed: false,
			expectedOutput: "Policy check found 1 failure in the code",
		},
		{
			description:    "Output with '2 failures' pattern should fail",
			policyOutput:   "Found 2 failures in security scan",
			expectedPassed: false,
			expectedOutput: "Found 2 failures in security scan",
		},
		{
			description:    "Output with '10 failures' pattern should fail",
			policyOutput:   "Total: 10 failures detected",
			expectedPassed: false,
			expectedOutput: "Total: 10 failures detected",
		},
		{
			description:    "Output with JSON 'failures': [...] pattern should fail",
			policyOutput:   `{"result": "failures": [{"rule": "test"}]}`,
			expectedPassed: false,
			expectedOutput: `{"result": "failures": [{"rule": "test"}]}`,
		},
		{
			description:    "Output starting with 'FAIL' prefix should fail",
			policyOutput:   "FAIL: Policy validation failed",
			expectedPassed: false,
			expectedOutput: "FAIL: Policy validation failed",
		},
		{
			description:    "Output starting with 'FAIL' after whitespace should fail",
			policyOutput:   "  FAIL: Something went wrong",
			expectedPassed: false,
			expectedOutput: "  FAIL: Something went wrong",
		},
		{
			description:    "Output with 'FAIL' in middle should pass",
			policyOutput:   "The check did not FAIL completely",
			expectedPassed: true,
			expectedOutput: "The check did not FAIL completely",
		},
		{
			description:    "Output with '0 failure' should pass (regex only matches 1-9)",
			policyOutput:   "Found 0 failure in the scan",
			expectedPassed: true,
			expectedOutput: "Found 0 failure in the scan",
		},
		{
			description:    "Output with word 'failure' but not pattern should pass",
			policyOutput:   "This is a failure message but not a failure count",
			expectedPassed: true,
			expectedOutput: "This is a failure message but not a failure count",
		},
		{
			description:    "Output with 'fail' word should pass (not matching pattern)",
			policyOutput:   "The test might fail if conditions are not met",
			expectedPassed: true,
			expectedOutput: "The test might fail if conditions are not met",
		},
		{
			description:    "Output with 'failures' word but not JSON pattern should pass",
			policyOutput:   "Checking for potential failures in the system",
			expectedPassed: true,
			expectedOutput: "Checking for potential failures in the system",
		},
		{
			description:    "Output with '99 failures' should fail",
			policyOutput:   "Detected 99 failures in compliance check",
			expectedPassed: false,
			expectedOutput: "Detected 99 failures in compliance check",
		},
		{
			description:    "Output with '100 failures' should fail",
			policyOutput:   "Total: 100 failures found",
			expectedPassed: false,
			expectedOutput: "Total: 100 failures found",
		},
		{
			description:    "Normal success output should pass",
			policyOutput:   "Policy check passed successfully",
			expectedPassed: true,
			expectedOutput: "Policy check passed successfully",
		},
		{
			description:    "Empty output should fail (handled separately but included for completeness)",
			policyOutput:   "",
			expectedPassed: false,
			expectedOutput: "WARNING: Policy check produced no output. This may indicate a misconfiguration.",
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			mockPolicyCheck := mocks.NewMockStepRunner()
			mockWorkingDir := mocks.NewMockWorkingDir()
			mockLocker := mocks.NewMockProjectLocker()

			runner := events.DefaultProjectCommandRunner{
				Locker:                mockLocker,
				LockURLGenerator:      mockURLGenerator{},
				PolicyCheckStepRunner: mockPolicyCheck,
				WorkingDir:            mockWorkingDir,
				WorkingDirLocker:      events.NewDefaultWorkingDirLocker(),
			}

			repoDir := t.TempDir()
			When(mockWorkingDir.GetWorkingDir(
				Any[models.Repo](),
				Any[models.PullRequest](),
				Any[string](),
			)).ThenReturn(repoDir, nil)
			When(mockWorkingDir.GitReadLock(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(func() {})

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

			// Setup policy check step
			steps := []valid.Step{
				{
					StepName: "policy_check",
				},
			}

			// Setup mock to return the test output
			When(mockPolicyCheck.Run(
				Any[command.ProjectContext](),
				Any[[]string](),
				Any[string](),
				Any[map[string]string](),
			)).ThenReturn(c.policyOutput, nil)

			ctx := command.ProjectContext{
				Log:               logging.NewNoopLogger(t),
				Workspace:         "default",
				RepoRelDir:        ".",
				CustomPolicyCheck: true,
				PolicySets: valid.PolicySets{
					PolicySets: []valid.PolicySet{
						{
							Name:         "test_policy",
							ApproveCount: 1,
							Owners: valid.PolicyOwners{
								Users: []string{"test-user"},
							},
						},
					},
				},
				Steps: steps,
			}

			res := runner.PolicyCheck(ctx)

			Assert(t, res.Error == nil, "not expecting error: %v", res.Error)
			Assert(t, res.PolicyCheckResults != nil, "expecting policy check results")

			// Verify the result
			policyResults := res.PolicyCheckResults.PolicySetResults
			Equals(t, 1, len(policyResults))
			Equals(t, c.expectedPassed, policyResults[0].Passed)
			Equals(t, c.expectedOutput, policyResults[0].PolicyOutput)
			Equals(t, "test_policy", policyResults[0].PolicySetName)
		})
	}
}

// Test that custom policy checks do not produce any additional output in the PreConftestOutput or PostConfTestOutput blocks
// This is a regression test for two bugs:
// 1. A blank code block would appear in custom policy check comments due PreConftestOutput being present.
// 2. The last policy output would appear both in the PolicySetResults and in PostConftestOutput.
func TestDefaultProjectCommandRunner_CustomPolicyCheck_NoPreOrPostConftestOutput(t *testing.T) {
	RegisterMockTestingT(t)

	cases := []struct {
		description   string
		policyOutputs []string
		policySets    []valid.PolicySet
	}{
		{
			description:   "Single custom policy check",
			policyOutputs: []string{"Custom policy check - 0 failures, 5 passed"},
			policySets: []valid.PolicySet{
				{
					Name:         "policy1",
					ApproveCount: 1,
					Owners: valid.PolicyOwners{
						Teams: []string{"team-1"},
					},
				},
			},
		},
		{
			description: "Multiple custom policy checks",
			policyOutputs: []string{
				"Custom policy check 1 - 0 failures, 5 passed",
				"Custom policy check 2 - 1 failures, 3 passed",
			},
			policySets: []valid.PolicySet{
				{
					Name:         "policy1",
					ApproveCount: 1,
					Owners: valid.PolicyOwners{
						Teams: []string{"team-1"},
					},
				},
				{
					Name:         "policy2",
					ApproveCount: 1,
					Owners: valid.PolicyOwners{
						Teams: []string{"team-2"},
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			mockPolicyCheck := mocks.NewMockStepRunner()
			mockWorkingDir := mocks.NewMockWorkingDir()
			mockLocker := mocks.NewMockProjectLocker()

			runner := events.DefaultProjectCommandRunner{
				Locker:                mockLocker,
				LockURLGenerator:      mockURLGenerator{},
				PolicyCheckStepRunner: mockPolicyCheck,
				WorkingDir:            mockWorkingDir,
				WorkingDirLocker:      events.NewDefaultWorkingDirLocker(),
			}

			repoDir := t.TempDir()
			When(mockWorkingDir.GetWorkingDir(
				Any[models.Repo](),
				Any[models.PullRequest](),
				Any[string](),
			)).ThenReturn(repoDir, nil)
			When(mockWorkingDir.GitReadLock(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(func() {})

			When(mockLocker.TryLock(
				Any[logging.SimpleLogging](),
				Any[models.PullRequest](),
				Any[models.User](),
				Any[string](),
				Any[models.Project](),
				Any[bool](),
			)).ThenReturn(&events.TryLockResponse{
				LockAcquired: true,
				LockKey:      "lock-key",
			}, nil)

			var steps []valid.Step
			for range c.policyOutputs {
				steps = append(steps, valid.Step{StepName: "policy_check"})
			}

			mockCall := When(mockPolicyCheck.Run(
				Any[command.ProjectContext](),
				Any[[]string](),
				Any[string](),
				Any[map[string]string](),
			))
			for _, output := range c.policyOutputs {
				mockCall = mockCall.ThenReturn(output, nil)
			}

			ctx := command.ProjectContext{
				Log:               logging.NewNoopLogger(t),
				Workspace:         "default",
				RepoRelDir:        ".",
				CustomPolicyCheck: true,
				PolicySets: valid.PolicySets{
					PolicySets: c.policySets,
				},
				Steps: steps,
			}

			res := runner.PolicyCheck(ctx)

			Assert(t, res.Error == nil, "not expecting error: %v", res.Error)
			Assert(t, res.PolicyCheckResults != nil, "expecting policy check results")

			policyResults := res.PolicyCheckResults.PolicySetResults
			Equals(t, len(c.policyOutputs), len(policyResults))

			for i, expectedOutput := range c.policyOutputs {
				Equals(t, c.policySets[i].Name, policyResults[i].PolicySetName)
				Equals(t, expectedOutput, policyResults[i].PolicyOutput)
			}

			// All outputs should be in PolicySetResults, not in PreConftestOutput or PostConftestOutput
			Equals(t, "", res.PolicyCheckResults.PreConftestOutput)
			Equals(t, "", res.PolicyCheckResults.PostConftestOutput)
		})
	}
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
					PolicySetName:    "policy1",
					ReqApprovalCount: 1,
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
					PolicySetName:    "policy1",
					ReqApprovalCount: 1,
					Approvals:        []models.PolicySetApproval{{Approver: testdata.User.Username, Hashes: nil}},
				},
				{
					PolicySetName:    "policy2",
					ReqApprovalCount: 2,
					Approvals:        []models.PolicySetApproval{{Approver: testdata.User.Username, Hashes: nil}},
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
					PolicySetName:    "policy1",
					ReqApprovalCount: 1,
					Approvals:        []models.PolicySetApproval{{Approver: testdata.User.Username, Hashes: nil}},
				},
				{
					PolicySetName:    "policy2",
					ReqApprovalCount: 2,
				},
			},
			expFailure: "One or more policy sets require additional approval.",
		},
		{
			description: "When user is a top-level owner through membership, increment approval on all policies.",
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
					PolicySetName:    "policy1",
					ReqApprovalCount: 1,
					Approvals:        []models.PolicySetApproval{{Approver: testdata.User.Username, Hashes: nil}},
				},
				{
					PolicySetName:    "policy2",
					ReqApprovalCount: 1,
					Approvals:        []models.PolicySetApproval{{Approver: testdata.User.Username, Hashes: nil}},
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
					PolicySetName:    "policy1",
					ReqApprovalCount: 1,
					Approvals:        []models.PolicySetApproval{{Approver: testdata.User.Username, Hashes: nil}},
				},
				{
					PolicySetName:    "policy2",
					ReqApprovalCount: 1,
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
					Approvals:     []models.PolicySetApproval{{Approver: "approver1"}, {Approver: "approver2"}},
				},
			},
			expOut: []models.PolicySetResult{
				{
					PolicySetName:    "policy1",
					ReqApprovalCount: 2,
					Approvals:        []models.PolicySetApproval{{Approver: "approver1"}, {Approver: "approver2"}},
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
					Approvals:     nil,
				},
			},
			expOut: []models.PolicySetResult{
				{
					PolicySetName:    "policy1",
					ReqApprovalCount: 2,
					Passed:           true,
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
					Approvals:     nil,
					Passed:        false,
				},
				{
					PolicySetName: "policy2",
					Approvals:     nil,
					Passed:        false,
				},
			},
			expOut: []models.PolicySetResult{
				{
					PolicySetName:    "policy1",
					ReqApprovalCount: 1,
					Approvals:        []models.PolicySetApproval{{Approver: testdata.User.Username, Hashes: nil}},
				},
				{
					PolicySetName:    "policy2",
					ReqApprovalCount: 1,
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
					Approvals:     []models.PolicySetApproval{{Approver: "approver1"}},
					Passed:        false,
				},
				{
					PolicySetName: "policy2",
					Approvals:     []models.PolicySetApproval{{Approver: "approver1"}},
					Passed:        false,
				},
			},
			expOut: []models.PolicySetResult{
				{
					PolicySetName:    "policy1",
					ReqApprovalCount: 1,
					Approvals:        []models.PolicySetApproval{},
				},
				{
					PolicySetName:    "policy2",
					ReqApprovalCount: 2,
					Approvals:        []models.PolicySetApproval{},
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
					Approvals:     []models.PolicySetApproval{{Approver: "approver1"}},
					Passed:        false,
				},
				{
					PolicySetName: "policy2",
					Approvals:     []models.PolicySetApproval{{Approver: "approver1"}},
					Passed:        false,
				},
			},
			expOut: []models.PolicySetResult{
				{
					PolicySetName:    "policy1",
					ReqApprovalCount: 1,
					Approvals:        []models.PolicySetApproval{},
				},
				{
					PolicySetName:    "policy2",
					ReqApprovalCount: 2,
					Approvals:        []models.PolicySetApproval{{Approver: "approver1"}},
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
					Approvals:     []models.PolicySetApproval{{Approver: "approver1"}},
					Passed:        false,
				},
				{
					PolicySetName: "policy2",
					Approvals:     []models.PolicySetApproval{{Approver: "approver1"}},
					Passed:        false,
				},
			},
			expOut: []models.PolicySetResult{
				{
					PolicySetName:    "policy1",
					ReqApprovalCount: 1,
					Approvals:        []models.PolicySetApproval{{Approver: "approver1"}},
				},
				{
					PolicySetName:    "policy2",
					ReqApprovalCount: 2,
					Approvals:        []models.PolicySetApproval{},
				},
			},
			expFailure: `One or more policy sets require additional approval.`,
			hasErr:     false,
		},
		{
			description:         "Policy Approval should not be the Author of the PR",
			userTeams:           []string{"someuserteam"},
			clearPolicyApproval: false,
			policySetCfg: valid.PolicySets{
				PolicySets: []valid.PolicySet{
					{
						Owners: valid.PolicyOwners{
							Users: []string{"lkysow"},
						},
						Name:         "policy1",
						ApproveCount: 1,
					},
					{
						Owners: valid.PolicyOwners{
							Users: []string{"lkysow"},
						},
						Name:               "policy2",
						ApproveCount:       1,
						PreventSelfApprove: true,
					},
				},
			},
			policySetStatus: []models.PolicySetStatus{
				{
					PolicySetName: "policy1",
					Approvals:     nil,
					Passed:        false,
				},
				{
					PolicySetName: "policy2",
					Approvals:     nil,
					Passed:        false,
				},
			},
			expOut: []models.PolicySetResult{
				{
					PolicySetName:    "policy1",
					ReqApprovalCount: 1,
					Approvals:        []models.PolicySetApproval{{Approver: testdata.User.Username, Hashes: nil}},
				},
				{
					PolicySetName:    "policy2",
					ReqApprovalCount: 1,
				},
			},
			expFailure: `One or more policy sets require additional approval.`,
			hasErr:     true,
		},
		{
			// Regression: prior to the explicit break in doApprovePolicies'
			// inner loop, two ProjectPolicyStatus entries that share a name
			// would each receive an approval and emit a duplicate
			// PolicySetResult. The break ensures exactly one match per
			// configured policy set.
			description: "Duplicate PolicySetName entries in ProjectPolicyStatus only produce one approval and one result.",
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
				},
			},
			policySetStatus: []models.PolicySetStatus{
				{PolicySetName: "policy1"},
				{PolicySetName: "policy1"},
			},
			expOut: []models.PolicySetResult{
				{
					PolicySetName:    "policy1",
					ReqApprovalCount: 1,
					Approvals:        []models.PolicySetApproval{{Approver: testdata.User.Username, Hashes: nil}},
				},
			},
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
			When(mockWorkingDir.GitReadLock(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(func() {})
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

			modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, Author: testdata.User.Username}
			When(runner.VcsClient.GetTeamNamesForUser(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.User))).ThenReturn(c.userTeams, nil)
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

func TestDefaultProjectCommandRunner_ApprovePolicies_DuplicateApproval(t *testing.T) {
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

	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, Author: testdata.User.Username}
	When(runner.VcsClient.GetTeamNamesForUser(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.User))).ThenReturn(nil, nil)

	ctx := command.ProjectContext{
		User:       testdata.User,
		Log:        logging.NewNoopLogger(t),
		Workspace:  "default",
		RepoRelDir: ".",
		PolicySets: valid.PolicySets{
			Owners: valid.PolicyOwners{
				Users: []string{testdata.User.Username},
			},
			PolicySets: []valid.PolicySet{
				{
					Name:         "policy1",
					ApproveCount: 2,
				},
			},
		},
		ProjectPolicyStatus: []models.PolicySetStatus{
			{
				PolicySetName: "policy1",
				Hashes:        []string{"h1", "h2"},
				Approvals: []models.PolicySetApproval{
					{Approver: testdata.User.Username, Hashes: []string{"h1", "h2"}},
				},
			},
		},
		Pull: modelPull,
	}

	res := runner.ApprovePolicies(ctx)
	Assert(t, res.Error != nil, "expected error for duplicate approval by same user")
	Equals(t, 1, len(res.PolicyCheckResults.PolicySetResults))
	result := res.PolicyCheckResults.PolicySetResults[0]
	Equals(t, 1, result.GetCurApprovals())
}

// Test that sticky carry-over preserves all approvals (including dormant ones
// for non-current hashes) so they can reactivate if code is reverted.
func TestDefaultProjectCommandRunner_PolicyCheck_StickyCarryOverPreservesDormantApprovals(t *testing.T) {
	RegisterMockTestingT(t)
	mockPolicyCheck := mocks.NewMockStepRunner()
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockLocker := mocks.NewMockProjectLocker()

	runner := events.DefaultProjectCommandRunner{
		Locker:                mockLocker,
		LockURLGenerator:      mockURLGenerator{},
		PolicyCheckStepRunner: mockPolicyCheck,
		WorkingDir:            mockWorkingDir,
		WorkingDirLocker:      events.NewDefaultWorkingDirLocker(),
	}

	repoDir := t.TempDir()
	When(mockWorkingDir.GetWorkingDir(
		Any[models.Repo](),
		Any[models.PullRequest](),
		Any[string](),
	)).ThenReturn(repoDir, nil)
	When(mockWorkingDir.GitReadLock(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(func() {})
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

	// Custom policy check returning NEW hashes (simulating changed code).
	When(mockPolicyCheck.Run(
		Any[command.ProjectContext](),
		Any[[]string](),
		Any[string](),
		Any[map[string]string](),
	)).ThenReturn("1 failure found\nnew-violation-line", nil)

	ctx := command.ProjectContext{
		Log:               logging.NewNoopLogger(t),
		Workspace:         "default",
		RepoRelDir:        ".",
		CustomPolicyCheck: true,
		PolicySets: valid.PolicySets{
			PolicySets: []valid.PolicySet{
				{
					Name:            "policy1",
					ApproveCount:    1,
					StickyApprovals: true,
					PolicyItemRegex: ".*",
				},
			},
		},
		// Previous status has an approval for OLD hashes (from a prior commit).
		ProjectPolicyStatus: []models.PolicySetStatus{
			{
				PolicySetName: "policy1",
				Hashes:        []string{"old-violation-line"},
				Approvals: []models.PolicySetApproval{
					{Approver: "boss", Hashes: []string{"old-violation-line"}},
				},
			},
		},
		Steps: []valid.Step{{StepName: "policy_check"}},
	}

	res := runner.PolicyCheck(ctx)

	Assert(t, res.Error == nil, "not expecting error: %v", res.Error)
	Assert(t, res.PolicyCheckResults != nil, "expecting policy check results")

	result := res.PolicyCheckResults.PolicySetResults[0]

	// The old approval should be preserved even though it doesn't match new hashes.
	Equals(t, 1, len(result.Approvals))
	Equals(t, "boss", result.Approvals[0].Approver)
	Equals(t, []string{"old-violation-line"}, result.Approvals[0].Hashes)

	// But GetCurApprovals should return 0 because the approval doesn't cover new hashes.
	status := models.PolicySetStatus{
		PolicySetName: result.PolicySetName,
		Hashes:        result.Hashes,
		Approvals:     result.Approvals,
	}
	Equals(t, 0, status.GetCurApprovals())
}

func TestDefaultProjectCommandRunner_PolicyCheck_StickyCarryOverBehavior(t *testing.T) {
	cases := []struct {
		description       string
		stickyApprovals   bool
		configRegex       string
		storedRegex       string
		storedApprovals   []models.PolicySetApproval
		expectCarriedOver bool
	}{
		{
			description:     "non-sticky policy set drops all previous approvals",
			stickyApprovals: false,
			configRegex:     ".*",
			storedRegex:     ".*",
			storedApprovals: []models.PolicySetApproval{
				{Approver: "boss", Hashes: []string{"old-hash"}},
			},
			expectCarriedOver: false,
		},
		{
			description:     "sticky with matching regex preserves approvals",
			stickyApprovals: true,
			configRegex:     ".+",
			storedRegex:     ".+",
			storedApprovals: []models.PolicySetApproval{
				{Approver: "boss", Hashes: []string{"old-hash"}},
			},
			expectCarriedOver: true,
		},
		{
			description:     "sticky with changed regex drops approvals",
			stickyApprovals: true,
			configRegex:     `(?m)^FAIL.*`,
			storedRegex:     ".+",
			storedApprovals: []models.PolicySetApproval{
				{Approver: "boss", Hashes: []string{"old-hash"}},
			},
			expectCarriedOver: false,
		},
		{
			description:     "sticky with empty stored regex (legacy) preserves approvals",
			stickyApprovals: true,
			configRegex:     ".+",
			storedRegex:     "",
			storedApprovals: []models.PolicySetApproval{
				{Approver: "boss", Hashes: []string{"old-hash"}},
			},
			expectCarriedOver: true,
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			RegisterMockTestingT(t)
			mockPolicyCheck := mocks.NewMockStepRunner()
			mockWorkingDir := mocks.NewMockWorkingDir()
			mockLocker := mocks.NewMockProjectLocker()

			runner := events.DefaultProjectCommandRunner{
				Locker:                mockLocker,
				LockURLGenerator:      mockURLGenerator{},
				PolicyCheckStepRunner: mockPolicyCheck,
				WorkingDir:            mockWorkingDir,
				WorkingDirLocker:      events.NewDefaultWorkingDirLocker(),
			}

			repoDir := t.TempDir()
			When(mockWorkingDir.GetWorkingDir(
				Any[models.Repo](),
				Any[models.PullRequest](),
				Any[string](),
			)).ThenReturn(repoDir, nil)
			When(mockWorkingDir.GitReadLock(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(func() {})
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

			When(mockPolicyCheck.Run(
				Any[command.ProjectContext](),
				Any[[]string](),
				Any[string](),
				Any[map[string]string](),
			)).ThenReturn("new-violation-line", nil)

			ctx := command.ProjectContext{
				Log:               logging.NewNoopLogger(t),
				Workspace:         "default",
				RepoRelDir:        ".",
				CustomPolicyCheck: true,
				PolicySets: valid.PolicySets{
					PolicySets: []valid.PolicySet{
						{
							Name:            "policy1",
							ApproveCount:    1,
							StickyApprovals: c.stickyApprovals,
							PolicyItemRegex: c.configRegex,
						},
					},
				},
				ProjectPolicyStatus: []models.PolicySetStatus{
					{
						PolicySetName:   "policy1",
						Hashes:          []string{"old-hash"},
						PolicyItemRegex: c.storedRegex,
						Approvals:       c.storedApprovals,
					},
				},
				Steps: []valid.Step{{StepName: "policy_check"}},
			}

			res := runner.PolicyCheck(ctx)

			Assert(t, res.Error == nil, "not expecting error: %v", res.Error)
			Assert(t, res.PolicyCheckResults != nil, "expecting policy check results")

			result := res.PolicyCheckResults.PolicySetResults[0]
			if c.expectCarriedOver {
				Equals(t, len(c.storedApprovals), len(result.Approvals))
				Equals(t, c.storedApprovals[0].Approver, result.Approvals[0].Approver)
			} else {
				Assert(t, len(result.Approvals) == 0 || result.Approvals == nil,
					"expected no approvals carried over but got %d", len(result.Approvals))
			}
		})
	}
}

func TestDefaultProjectCommandRunner_ApprovePolicies_HashAwareApproval(t *testing.T) {
	cases := []struct {
		description       string
		policySetStatus   []models.PolicySetStatus
		policySetCfg      valid.PolicySets
		expTotalApprovals int
		expValidApprovals int
		expFailure        string
		hasErr            bool
	}{
		{
			description: "approval with matching hashes counts",
			policySetCfg: valid.PolicySets{
				Owners: valid.PolicyOwners{
					Users: []string{testdata.User.Username},
				},
				PolicySets: []valid.PolicySet{
					{Name: "policy1", ApproveCount: 1},
				},
			},
			policySetStatus: []models.PolicySetStatus{
				{
					PolicySetName: "policy1",
					Hashes:        []string{"h1", "h2"},
				},
			},
			expTotalApprovals: 1,
			expValidApprovals: 1,
			expFailure:        "",
		},
		{
			description: "stale approval preserved alongside fresh approval, only fresh one valid",
			policySetCfg: valid.PolicySets{
				Owners: valid.PolicyOwners{
					Users: []string{testdata.User.Username},
				},
				PolicySets: []valid.PolicySet{
					{Name: "policy1", ApproveCount: 2},
				},
			},
			policySetStatus: []models.PolicySetStatus{
				{
					PolicySetName: "policy1",
					Hashes:        []string{"h_new"},
					Approvals: []models.PolicySetApproval{
						{Approver: "other-user", Hashes: []string{"h_old"}},
					},
				},
			},
			expTotalApprovals: 2,
			expValidApprovals: 1,
			expFailure:        "One or more policy sets require additional approval.",
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

			modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, Author: testdata.User.Username}
			When(runner.VcsClient.GetTeamNamesForUser(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.User))).ThenReturn(nil, nil)

			ctx := command.ProjectContext{
				User:                testdata.User,
				Log:                 logging.NewNoopLogger(t),
				Workspace:           "default",
				RepoRelDir:          ".",
				PolicySets:          c.policySetCfg,
				ProjectPolicyStatus: c.policySetStatus,
				Pull:                modelPull,
			}

			res := runner.ApprovePolicies(ctx)
			Equals(t, c.expFailure, res.Failure)
			result := res.PolicyCheckResults.PolicySetResults[0]
			Equals(t, c.expTotalApprovals, len(result.Approvals))
			Equals(t, c.expValidApprovals, result.GetCurApprovals())
			if c.hasErr {
				Assert(t, res.Error != nil, "expecting error")
			}
		})
	}
}

func TestDefaultProjectCommandRunner_PathTraversal(t *testing.T) {
	const defaultTraversalPattern = "../../../../etc"

	cases := []struct {
		name              string
		traversalPatterns []string
		expectUnlock      bool
		runFn             func(runner *events.DefaultProjectCommandRunner, ctx command.ProjectContext) error
		setupWorkingDir   func(mockWorkingDir *mocks.MockWorkingDir, repoDir string)
	}{
		{
			name: "Plan",
			traversalPatterns: []string{
				"../../../../etc",
				"../etc",
				"sub/../../etc",
			},
			expectUnlock: true,
			setupWorkingDir: func(mockWorkingDir *mocks.MockWorkingDir, repoDir string) {
				When(mockWorkingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[string]())).
					ThenReturn(repoDir, nil)
				When(mockWorkingDir.MergeAgain(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[string]())).
					ThenReturn(false, nil)
			},
			runFn: func(runner *events.DefaultProjectCommandRunner, ctx command.ProjectContext) error {
				return runner.Plan(ctx).Error
			},
		},
		{
			name:              "Apply",
			traversalPatterns: []string{defaultTraversalPattern},
			setupWorkingDir: func(mockWorkingDir *mocks.MockWorkingDir, repoDir string) {
				When(mockWorkingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).
					ThenReturn(repoDir, nil)
			},
			runFn: func(runner *events.DefaultProjectCommandRunner, ctx command.ProjectContext) error {
				return runner.Apply(ctx).Error
			},
		},
		{
			name:              "PolicyCheck",
			traversalPatterns: []string{defaultTraversalPattern},
			expectUnlock:      true,
			setupWorkingDir: func(mockWorkingDir *mocks.MockWorkingDir, repoDir string) {
				When(mockWorkingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).
					ThenReturn(repoDir, nil)
			},
			runFn: func(runner *events.DefaultProjectCommandRunner, ctx command.ProjectContext) error {
				return runner.PolicyCheck(ctx).Error
			},
		},
		{
			name:              "Version",
			traversalPatterns: []string{defaultTraversalPattern},
			setupWorkingDir: func(mockWorkingDir *mocks.MockWorkingDir, repoDir string) {
				When(mockWorkingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).
					ThenReturn(repoDir, nil)
			},
			runFn: func(runner *events.DefaultProjectCommandRunner, ctx command.ProjectContext) error {
				return runner.Version(ctx).Error
			},
		},
		{
			name:              "Import",
			traversalPatterns: []string{defaultTraversalPattern},
			setupWorkingDir: func(mockWorkingDir *mocks.MockWorkingDir, repoDir string) {
				When(mockWorkingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[string]())).
					ThenReturn(repoDir, nil)
			},
			runFn: func(runner *events.DefaultProjectCommandRunner, ctx command.ProjectContext) error {
				return runner.Import(ctx).Error
			},
		},
		{
			name:              "StateRm",
			traversalPatterns: []string{defaultTraversalPattern},
			setupWorkingDir: func(mockWorkingDir *mocks.MockWorkingDir, repoDir string) {
				When(mockWorkingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[string]())).
					ThenReturn(repoDir, nil)
			},
			runFn: func(runner *events.DefaultProjectCommandRunner, ctx command.ProjectContext) error {
				return runner.StateRm(ctx).Error
			},
		},
	}

	for _, tc := range cases {
		for _, pattern := range tc.traversalPatterns {
			t.Run(tc.name+" rejects traversal pattern "+pattern, func(t *testing.T) {
				RegisterMockTestingT(t)
				mockWorkingDir := mocks.NewMockWorkingDir()
				mockLocker := mocks.NewMockProjectLocker()
				runner := &events.DefaultProjectCommandRunner{
					Locker:           mockLocker,
					LockURLGenerator: mockURLGenerator{},
					WorkingDir:       mockWorkingDir,
					WorkingDirLocker: events.NewDefaultWorkingDirLocker(),
				}
				repoDir := t.TempDir()
				tc.setupWorkingDir(mockWorkingDir, repoDir)
				unlockCalled := false
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
					UnlockFn: func() error {
						unlockCalled = true
						return nil
					},
				}, nil)
				ctx := command.ProjectContext{
					Log:        logging.NewNoopLogger(t),
					Workspace:  "default",
					RepoRelDir: pattern,
					RePlanCmd:  "atlantis plan -d .",
				}
				err := tc.runFn(runner, ctx)
				Assert(t, err != nil, "expected error for RepoRelDir %q in runner %q", pattern, tc.name)
				Assert(t,
					strings.Contains(err.Error(), "project path traversal detected"),
					"expected traversal error for runner %q with RepoRelDir %q, got: %s", tc.name, pattern, err,
				)
				if tc.expectUnlock {
					Assert(t, unlockCalled, "expected runner %q with RepoRelDir %q to release project lock", tc.name, pattern)
				}
			})
		}
	}
}
