// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"errors"
	"testing"

	"github.com/runatlantis/atlantis/server/core/db/mocks"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	eventmocks "github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
	"go.uber.org/mock/gomock"
)

func TestOutputPersistingProjectCommandRunner_Plan(t *testing.T) {
	ctrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	mockDB := mocks.NewMockDatabase(ctrl)
	mockRunner := eventmocks.NewMockProjectCommandRunner(ctrl)
	persister := events.NewOutputPersister(mockDB, nil)

	ctx := command.ProjectContext{
		Log:         logger,
		CommandName: command.Plan,
		BaseRepo:    models.Repo{FullName: "owner/repo"},
		Pull:        models.PullRequest{Num: 1},
		ProjectName: "test-project",
		Workspace:   "default",
		RepoRelDir:  ".",
	}

	expectedOutput := command.ProjectCommandOutput{
		PlanSuccess: &models.PlanSuccess{
			TerraformOutput: "Plan: 1 to add, 0 to change, 0 to destroy.",
		},
	}

	mockRunner.EXPECT().Plan(gomock.Any()).Return(expectedOutput)

	// First call is PersistStub (running status), second is PersistResult (final status)
	var capturedOutput models.ProjectOutput
	gomock.InOrder(
		mockDB.EXPECT().SaveProjectOutput(gomock.Any()).Return(nil),
		mockDB.EXPECT().SaveProjectOutput(gomock.Any()).DoAndReturn(func(output models.ProjectOutput) error {
			capturedOutput = output
			return nil
		}),
	)

	runner := events.NewOutputPersistingProjectCommandRunner(mockRunner, persister)
	result := runner.Plan(ctx)

	Assert(t, result.PlanSuccess != nil, "expected PlanSuccess to be set")
	Equals(t, "Plan: 1 to add, 0 to change, 0 to destroy.", result.PlanSuccess.TerraformOutput)

	Equals(t, "owner/repo", capturedOutput.RepoFullName)
	Equals(t, 1, capturedOutput.PullNum)
	Equals(t, "plan", capturedOutput.CommandName)
	Equals(t, models.SuccessOutputStatus, capturedOutput.Status)
}

func TestOutputPersistingProjectCommandRunner_Apply(t *testing.T) {
	ctrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	mockDB := mocks.NewMockDatabase(ctrl)
	mockRunner := eventmocks.NewMockProjectCommandRunner(ctrl)
	persister := events.NewOutputPersister(mockDB, nil)

	ctx := command.ProjectContext{
		Log:         logger,
		CommandName: command.Apply,
		BaseRepo:    models.Repo{FullName: "owner/repo"},
		Pull:        models.PullRequest{Num: 2},
		ProjectName: "test-project",
		Workspace:   "default",
		RepoRelDir:  ".",
	}

	expectedOutput := command.ProjectCommandOutput{
		ApplySuccess: "Apply complete! Resources: 1 added, 0 changed, 0 destroyed.",
	}

	mockRunner.EXPECT().Apply(gomock.Any()).Return(expectedOutput)

	// First call is PersistStub (running status), second is PersistResult (final status)
	var capturedOutput models.ProjectOutput
	gomock.InOrder(
		mockDB.EXPECT().SaveProjectOutput(gomock.Any()).Return(nil),
		mockDB.EXPECT().SaveProjectOutput(gomock.Any()).DoAndReturn(func(output models.ProjectOutput) error {
			capturedOutput = output
			return nil
		}),
	)

	runner := events.NewOutputPersistingProjectCommandRunner(mockRunner, persister)
	result := runner.Apply(ctx)

	Equals(t, "Apply complete! Resources: 1 added, 0 changed, 0 destroyed.", result.ApplySuccess)

	Equals(t, "owner/repo", capturedOutput.RepoFullName)
	Equals(t, 2, capturedOutput.PullNum)
	Equals(t, "apply", capturedOutput.CommandName)
	Equals(t, models.SuccessOutputStatus, capturedOutput.Status)
}

func TestOutputPersistingProjectCommandRunner_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	mockDB := mocks.NewMockDatabase(ctrl)
	mockRunner := eventmocks.NewMockProjectCommandRunner(ctrl)
	persister := events.NewOutputPersister(mockDB, nil)

	ctx := command.ProjectContext{
		Log:         logger,
		CommandName: command.Plan,
		BaseRepo:    models.Repo{FullName: "owner/repo"},
		Pull:        models.PullRequest{Num: 1},
		ProjectName: "test-project",
		Workspace:   "default",
		RepoRelDir:  ".",
	}

	expectedOutput := command.ProjectCommandOutput{
		Error: errors.New("terraform init failed"),
	}

	mockRunner.EXPECT().Plan(gomock.Any()).Return(expectedOutput)

	// First call is PersistStub (running status), second is PersistResult (final status)
	var capturedOutput models.ProjectOutput
	gomock.InOrder(
		mockDB.EXPECT().SaveProjectOutput(gomock.Any()).Return(nil),
		mockDB.EXPECT().SaveProjectOutput(gomock.Any()).DoAndReturn(func(output models.ProjectOutput) error {
			capturedOutput = output
			return nil
		}),
	)

	runner := events.NewOutputPersistingProjectCommandRunner(mockRunner, persister)
	result := runner.Plan(ctx)

	Assert(t, result.Error != nil, "expected error to be set")
	Equals(t, "terraform init failed", result.Error.Error())

	Equals(t, models.FailedOutputStatus, capturedOutput.Status)
	Equals(t, "terraform init failed", capturedOutput.Error)
}

func TestOutputPersistingProjectCommandRunner_NilPersister(t *testing.T) {
	ctrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	mockRunner := eventmocks.NewMockProjectCommandRunner(ctrl)

	ctx := command.ProjectContext{
		Log:         logger,
		CommandName: command.Plan,
		BaseRepo:    models.Repo{FullName: "owner/repo"},
		Pull:        models.PullRequest{Num: 1},
	}

	expectedOutput := command.ProjectCommandOutput{
		PlanSuccess: &models.PlanSuccess{
			TerraformOutput: "Plan output",
		},
	}

	mockRunner.EXPECT().Plan(gomock.Any()).Return(expectedOutput)

	// Create runner with nil persister - should not panic
	runner := events.NewOutputPersistingProjectCommandRunner(mockRunner, nil)
	result := runner.Plan(ctx)

	Assert(t, result.PlanSuccess != nil, "expected PlanSuccess to be set")
}

func TestOutputPersistingProjectCommandRunner_PolicyCheck(t *testing.T) {
	ctrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	mockDB := mocks.NewMockDatabase(ctrl)
	mockRunner := eventmocks.NewMockProjectCommandRunner(ctrl)
	persister := events.NewOutputPersister(mockDB, nil)

	ctx := command.ProjectContext{
		Log:         logger,
		CommandName: command.PolicyCheck,
		BaseRepo:    models.Repo{FullName: "owner/repo"},
		Pull:        models.PullRequest{Num: 1},
		ProjectName: "test-project",
		Workspace:   "default",
		RepoRelDir:  ".",
	}

	expectedOutput := command.ProjectCommandOutput{
		PolicyCheckResults: &models.PolicyCheckResults{
			PolicySetResults: []models.PolicySetResult{
				{PolicySetName: "test-policy", Passed: true, PolicyOutput: "All policies passed"},
			},
		},
	}

	mockRunner.EXPECT().PolicyCheck(gomock.Any()).Return(expectedOutput)

	// First call is PersistStub (running status), second is PersistResult (final status)
	var capturedOutput models.ProjectOutput
	gomock.InOrder(
		mockDB.EXPECT().SaveProjectOutput(gomock.Any()).Return(nil),
		mockDB.EXPECT().SaveProjectOutput(gomock.Any()).DoAndReturn(func(output models.ProjectOutput) error {
			capturedOutput = output
			return nil
		}),
	)

	runner := events.NewOutputPersistingProjectCommandRunner(mockRunner, persister)
	result := runner.PolicyCheck(ctx)

	Assert(t, result.PolicyCheckResults != nil, "expected PolicyCheckResults to be set")
	Equals(t, 1, len(result.PolicyCheckResults.PolicySetResults))

	Equals(t, "policy_check", capturedOutput.CommandName)
	Equals(t, models.SuccessOutputStatus, capturedOutput.Status)
	Equals(t, true, capturedOutput.PolicyPassed)
}

func TestOutputPersistingProjectCommandRunner_Version(t *testing.T) {
	ctrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	mockDB := mocks.NewMockDatabase(ctrl)
	mockRunner := eventmocks.NewMockProjectCommandRunner(ctrl)
	persister := events.NewOutputPersister(mockDB, nil)

	ctx := command.ProjectContext{
		Log:         logger,
		CommandName: command.Version,
		BaseRepo:    models.Repo{FullName: "owner/repo"},
		Pull:        models.PullRequest{Num: 1},
		ProjectName: "test-project",
		Workspace:   "default",
		RepoRelDir:  ".",
	}

	expectedOutput := command.ProjectCommandOutput{
		VersionSuccess: "Terraform v1.5.0",
	}

	mockRunner.EXPECT().Version(gomock.Any()).Return(expectedOutput)

	// First call is PersistStub (running status), second is PersistResult (final status)
	var capturedOutput models.ProjectOutput
	gomock.InOrder(
		mockDB.EXPECT().SaveProjectOutput(gomock.Any()).Return(nil),
		mockDB.EXPECT().SaveProjectOutput(gomock.Any()).DoAndReturn(func(output models.ProjectOutput) error {
			capturedOutput = output
			return nil
		}),
	)

	runner := events.NewOutputPersistingProjectCommandRunner(mockRunner, persister)
	result := runner.Version(ctx)

	Equals(t, "Terraform v1.5.0", result.VersionSuccess)

	Equals(t, "version", capturedOutput.CommandName)
	Equals(t, models.SuccessOutputStatus, capturedOutput.Status)
}

func TestOutputPersistingProjectCommandRunner_Import(t *testing.T) {
	ctrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	mockDB := mocks.NewMockDatabase(ctrl)
	mockRunner := eventmocks.NewMockProjectCommandRunner(ctrl)
	persister := events.NewOutputPersister(mockDB, nil)

	ctx := command.ProjectContext{
		Log:         logger,
		CommandName: command.Import,
		BaseRepo:    models.Repo{FullName: "owner/repo"},
		Pull:        models.PullRequest{Num: 1},
		ProjectName: "test-project",
		Workspace:   "default",
		RepoRelDir:  ".",
	}

	expectedOutput := command.ProjectCommandOutput{
		ImportSuccess: &models.ImportSuccess{
			Output: "Import successful",
		},
	}

	mockRunner.EXPECT().Import(gomock.Any()).Return(expectedOutput)

	// First call is PersistStub (running status), second is PersistResult (final status)
	var capturedOutput models.ProjectOutput
	gomock.InOrder(
		mockDB.EXPECT().SaveProjectOutput(gomock.Any()).Return(nil),
		mockDB.EXPECT().SaveProjectOutput(gomock.Any()).DoAndReturn(func(output models.ProjectOutput) error {
			capturedOutput = output
			return nil
		}),
	)

	runner := events.NewOutputPersistingProjectCommandRunner(mockRunner, persister)
	result := runner.Import(ctx)

	Assert(t, result.ImportSuccess != nil, "expected ImportSuccess to be set")

	Equals(t, "import", capturedOutput.CommandName)
	Equals(t, models.SuccessOutputStatus, capturedOutput.Status)
}

func TestOutputPersistingProjectCommandRunner_StateRm(t *testing.T) {
	ctrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	mockDB := mocks.NewMockDatabase(ctrl)
	mockRunner := eventmocks.NewMockProjectCommandRunner(ctrl)
	persister := events.NewOutputPersister(mockDB, nil)

	ctx := command.ProjectContext{
		Log:         logger,
		CommandName: command.State,
		BaseRepo:    models.Repo{FullName: "owner/repo"},
		Pull:        models.PullRequest{Num: 1},
		ProjectName: "test-project",
		Workspace:   "default",
		RepoRelDir:  ".",
	}

	expectedOutput := command.ProjectCommandOutput{
		StateRmSuccess: &models.StateRmSuccess{
			Output: "Successfully removed 1 resource",
		},
	}

	mockRunner.EXPECT().StateRm(gomock.Any()).Return(expectedOutput)

	// First call is PersistStub (running status), second is PersistResult (final status)
	var capturedOutput models.ProjectOutput
	gomock.InOrder(
		mockDB.EXPECT().SaveProjectOutput(gomock.Any()).Return(nil),
		mockDB.EXPECT().SaveProjectOutput(gomock.Any()).DoAndReturn(func(output models.ProjectOutput) error {
			capturedOutput = output
			return nil
		}),
	)

	runner := events.NewOutputPersistingProjectCommandRunner(mockRunner, persister)
	result := runner.StateRm(ctx)

	Assert(t, result.StateRmSuccess != nil, "expected StateRmSuccess to be set")

	Equals(t, "state", capturedOutput.CommandName)
	Equals(t, models.SuccessOutputStatus, capturedOutput.Status)
}

func TestOutputPersistingProjectCommandRunner_ApprovePolicies(t *testing.T) {
	ctrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	mockDB := mocks.NewMockDatabase(ctrl)
	mockRunner := eventmocks.NewMockProjectCommandRunner(ctrl)
	persister := events.NewOutputPersister(mockDB, nil)

	ctx := command.ProjectContext{
		Log:         logger,
		CommandName: command.ApprovePolicies,
		BaseRepo:    models.Repo{FullName: "owner/repo"},
		Pull:        models.PullRequest{Num: 1},
		ProjectName: "test-project",
		Workspace:   "default",
		RepoRelDir:  ".",
	}

	expectedOutput := command.ProjectCommandOutput{
		PolicyCheckResults: &models.PolicyCheckResults{
			PolicySetResults: []models.PolicySetResult{
				{PolicySetName: "test-policy", Passed: true, CurApprovals: 1, ReqApprovals: 1},
			},
		},
	}

	mockRunner.EXPECT().ApprovePolicies(gomock.Any()).Return(expectedOutput)

	// First call is PersistStub (running status), second is PersistResult (final status)
	var capturedOutput models.ProjectOutput
	gomock.InOrder(
		mockDB.EXPECT().SaveProjectOutput(gomock.Any()).Return(nil),
		mockDB.EXPECT().SaveProjectOutput(gomock.Any()).DoAndReturn(func(output models.ProjectOutput) error {
			capturedOutput = output
			return nil
		}),
	)

	runner := events.NewOutputPersistingProjectCommandRunner(mockRunner, persister)
	result := runner.ApprovePolicies(ctx)

	Assert(t, result.PolicyCheckResults != nil, "expected PolicyCheckResults to be set")

	Equals(t, "approve_policies", capturedOutput.CommandName)
	Equals(t, models.SuccessOutputStatus, capturedOutput.Status)
}

func TestOutputPersistingProjectCommandRunner_DatabaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	mockDB := mocks.NewMockDatabase(ctrl)
	mockRunner := eventmocks.NewMockProjectCommandRunner(ctrl)
	persister := events.NewOutputPersister(mockDB, nil)

	ctx := command.ProjectContext{
		Log:         logger,
		CommandName: command.Plan,
		BaseRepo:    models.Repo{FullName: "owner/repo"},
		Pull:        models.PullRequest{Num: 1},
	}

	expectedOutput := command.ProjectCommandOutput{
		PlanSuccess: &models.PlanSuccess{
			TerraformOutput: "Plan output",
		},
	}

	mockRunner.EXPECT().Plan(gomock.Any()).Return(expectedOutput)

	// First call (stub) fails, second call (result) also fails
	gomock.InOrder(
		mockDB.EXPECT().SaveProjectOutput(gomock.Any()).Return(errors.New("database connection lost")),
		mockDB.EXPECT().SaveProjectOutput(gomock.Any()).Return(errors.New("database connection lost")),
	)

	runner := events.NewOutputPersistingProjectCommandRunner(mockRunner, persister)

	// Should not panic even if database save fails
	result := runner.Plan(ctx)

	// Verify result is still passed through
	Assert(t, result.PlanSuccess != nil, "expected PlanSuccess to be set")
}
