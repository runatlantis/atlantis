// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/petergtz/pegomock/v4"
	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/db/mocks"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	eventmocks "github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestOutputPersistingProjectCommandRunner_Plan(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	mockDB := mocks.NewMockDatabase()
	mockRunner := eventmocks.NewMockProjectCommandRunner()
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

	When(mockRunner.Plan(AnyProjectContext())).ThenReturn(expectedOutput)

	runner := events.NewOutputPersistingProjectCommandRunner(mockRunner, persister)
	result := runner.Plan(ctx)

	// Verify result is passed through
	Assert(t, result.PlanSuccess != nil, "expected PlanSuccess to be set")
	Equals(t, "Plan: 1 to add, 0 to change, 0 to destroy.", result.PlanSuccess.TerraformOutput)

	// Verify inner runner was called
	mockRunner.VerifyWasCalledOnce().Plan(AnyProjectContext())

	// Verify output was persisted
	mockDB.VerifyWasCalledOnce().SaveProjectOutput(AnyProjectOutput())
}

func TestOutputPersistingProjectCommandRunner_Apply(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	mockDB := mocks.NewMockDatabase()
	mockRunner := eventmocks.NewMockProjectCommandRunner()
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

	When(mockRunner.Apply(AnyProjectContext())).ThenReturn(expectedOutput)

	runner := events.NewOutputPersistingProjectCommandRunner(mockRunner, persister)
	result := runner.Apply(ctx)

	// Verify result is passed through
	Equals(t, "Apply complete! Resources: 1 added, 0 changed, 0 destroyed.", result.ApplySuccess)

	// Verify output was persisted
	mockDB.VerifyWasCalledOnce().SaveProjectOutput(AnyProjectOutput())
}

func TestOutputPersistingProjectCommandRunner_Error(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	mockDB := mocks.NewMockDatabase()
	mockRunner := eventmocks.NewMockProjectCommandRunner()
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

	When(mockRunner.Plan(AnyProjectContext())).ThenReturn(expectedOutput)

	runner := events.NewOutputPersistingProjectCommandRunner(mockRunner, persister)
	result := runner.Plan(ctx)

	// Verify error is passed through
	Assert(t, result.Error != nil, "expected error to be set")
	Equals(t, "terraform init failed", result.Error.Error())

	// Verify output was persisted (even on error)
	mockDB.VerifyWasCalledOnce().SaveProjectOutput(AnyProjectOutput())
}

func TestOutputPersistingProjectCommandRunner_NilPersister(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	mockRunner := eventmocks.NewMockProjectCommandRunner()

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

	When(mockRunner.Plan(AnyProjectContext())).ThenReturn(expectedOutput)

	// Create runner with nil persister - should not panic
	runner := events.NewOutputPersistingProjectCommandRunner(mockRunner, nil)
	result := runner.Plan(ctx)

	// Verify result is still passed through
	Assert(t, result.PlanSuccess != nil, "expected PlanSuccess to be set")
}

func TestOutputPersistingProjectCommandRunner_PolicyCheck(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	mockDB := mocks.NewMockDatabase()
	mockRunner := eventmocks.NewMockProjectCommandRunner()
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

	When(mockRunner.PolicyCheck(AnyProjectContext())).ThenReturn(expectedOutput)

	runner := events.NewOutputPersistingProjectCommandRunner(mockRunner, persister)
	result := runner.PolicyCheck(ctx)

	// Verify result is passed through
	Assert(t, result.PolicyCheckResults != nil, "expected PolicyCheckResults to be set")
	Equals(t, 1, len(result.PolicyCheckResults.PolicySetResults))

	// Verify output was persisted
	mockDB.VerifyWasCalledOnce().SaveProjectOutput(AnyProjectOutput())
}

func TestOutputPersistingProjectCommandRunner_Version(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	mockDB := mocks.NewMockDatabase()
	mockRunner := eventmocks.NewMockProjectCommandRunner()
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

	When(mockRunner.Version(AnyProjectContext())).ThenReturn(expectedOutput)

	runner := events.NewOutputPersistingProjectCommandRunner(mockRunner, persister)
	result := runner.Version(ctx)

	// Verify result is passed through
	Equals(t, "Terraform v1.5.0", result.VersionSuccess)

	// Verify output was persisted
	mockDB.VerifyWasCalledOnce().SaveProjectOutput(AnyProjectOutput())
}

func TestOutputPersistingProjectCommandRunner_Import(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	mockDB := mocks.NewMockDatabase()
	mockRunner := eventmocks.NewMockProjectCommandRunner()
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

	When(mockRunner.Import(AnyProjectContext())).ThenReturn(expectedOutput)

	runner := events.NewOutputPersistingProjectCommandRunner(mockRunner, persister)
	result := runner.Import(ctx)

	// Verify result is passed through
	Assert(t, result.ImportSuccess != nil, "expected ImportSuccess to be set")

	// Verify output was persisted
	mockDB.VerifyWasCalledOnce().SaveProjectOutput(AnyProjectOutput())
}

func TestOutputPersistingProjectCommandRunner_StateRm(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	mockDB := mocks.NewMockDatabase()
	mockRunner := eventmocks.NewMockProjectCommandRunner()
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

	When(mockRunner.StateRm(AnyProjectContext())).ThenReturn(expectedOutput)

	runner := events.NewOutputPersistingProjectCommandRunner(mockRunner, persister)
	result := runner.StateRm(ctx)

	// Verify result is passed through
	Assert(t, result.StateRmSuccess != nil, "expected StateRmSuccess to be set")

	// Verify output was persisted
	mockDB.VerifyWasCalledOnce().SaveProjectOutput(AnyProjectOutput())
}

func TestOutputPersistingProjectCommandRunner_ApprovePolicies(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	mockDB := mocks.NewMockDatabase()
	mockRunner := eventmocks.NewMockProjectCommandRunner()
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

	When(mockRunner.ApprovePolicies(AnyProjectContext())).ThenReturn(expectedOutput)

	runner := events.NewOutputPersistingProjectCommandRunner(mockRunner, persister)
	result := runner.ApprovePolicies(ctx)

	// Verify result is passed through
	Assert(t, result.PolicyCheckResults != nil, "expected PolicyCheckResults to be set")

	// Verify output was persisted
	mockDB.VerifyWasCalledOnce().SaveProjectOutput(AnyProjectOutput())
}

func TestOutputPersistingProjectCommandRunner_DatabaseError(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	mockDB := mocks.NewMockDatabase()
	mockRunner := eventmocks.NewMockProjectCommandRunner()
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

	When(mockRunner.Plan(AnyProjectContext())).ThenReturn(expectedOutput)
	When(mockDB.SaveProjectOutput(AnyProjectOutput())).ThenReturn(errors.New("database connection lost"))

	runner := events.NewOutputPersistingProjectCommandRunner(mockRunner, persister)

	// Should not panic even if database save fails
	result := runner.Plan(ctx)

	// Verify result is still passed through
	Assert(t, result.PlanSuccess != nil, "expected PlanSuccess to be set")
}

// AnyProjectContext returns a pegomock matcher for command.ProjectContext
func AnyProjectContext() command.ProjectContext {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf(command.ProjectContext{})))
	return command.ProjectContext{}
}
