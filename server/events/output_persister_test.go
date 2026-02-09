// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"errors"
	"reflect"
	"testing"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/db/mocks"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

func TestOutputPersister_PersistPlanSuccess(t *testing.T) {
	RegisterMockTestingT(t)
	mockDB := mocks.NewMockDatabase()
	persister := events.NewOutputPersister(mockDB, nil)

	ctx := command.ProjectContext{
		BaseRepo: models.Repo{FullName: "owner/repo"},
		Pull: models.PullRequest{
			Num:   123,
			URL:   "https://github.com/owner/repo/pull/123",
			Title: "Add new infrastructure",
		},
		RepoRelDir:  "terraform/staging",
		Workspace:   "default",
		ProjectName: "myproject",
		User:        models.User{Username: "testuser"},
		CommandName: command.Plan,
		JobID:       "job-123",
	}

	result := command.ProjectResult{
		Command:     command.Plan,
		RepoRelDir:  "terraform/staging",
		Workspace:   "default",
		ProjectName: "myproject",
		ProjectCommandOutput: command.ProjectCommandOutput{
			PlanSuccess: &models.PlanSuccess{
				TerraformOutput: "Plan: 1 to add, 2 to change, 3 to destroy.",
			},
		},
	}

	err := persister.PersistResult(ctx, result)
	Ok(t, err)

	// Verify SaveProjectOutput was called
	mockDB.VerifyWasCalledOnce().SaveProjectOutput(AnyProjectOutput())
}

func TestOutputPersister_PersistFailure(t *testing.T) {
	RegisterMockTestingT(t)
	mockDB := mocks.NewMockDatabase()
	persister := events.NewOutputPersister(mockDB, nil)

	ctx := command.ProjectContext{
		BaseRepo:    models.Repo{FullName: "owner/repo"},
		Pull:        models.PullRequest{Num: 123},
		RepoRelDir:  "terraform/staging",
		Workspace:   "default",
		ProjectName: "myproject",
		User:        models.User{Username: "testuser"},
		CommandName: command.Plan,
	}

	result := command.ProjectResult{
		Command:     command.Plan,
		RepoRelDir:  "terraform/staging",
		Workspace:   "default",
		ProjectName: "myproject",
		ProjectCommandOutput: command.ProjectCommandOutput{
			Failure: "terraform init failed",
		},
	}

	err := persister.PersistResult(ctx, result)
	Ok(t, err)

	mockDB.VerifyWasCalledOnce().SaveProjectOutput(AnyProjectOutput())
}

func TestOutputPersister_PersistApplySuccess(t *testing.T) {
	RegisterMockTestingT(t)
	mockDB := mocks.NewMockDatabase()
	persister := events.NewOutputPersister(mockDB, nil)

	ctx := command.ProjectContext{
		BaseRepo:    models.Repo{FullName: "owner/repo"},
		Pull:        models.PullRequest{Num: 123},
		RepoRelDir:  "terraform/staging",
		Workspace:   "default",
		ProjectName: "myproject",
		User:        models.User{Username: "testuser"},
		CommandName: command.Apply,
	}

	result := command.ProjectResult{
		Command:     command.Apply,
		RepoRelDir:  "terraform/staging",
		Workspace:   "default",
		ProjectName: "myproject",
		ProjectCommandOutput: command.ProjectCommandOutput{
			ApplySuccess: "Apply complete! Resources: 1 added, 0 changed, 0 destroyed.",
		},
	}

	err := persister.PersistResult(ctx, result)
	Ok(t, err)

	mockDB.VerifyWasCalledOnce().SaveProjectOutput(AnyProjectOutput())
}

func TestOutputPersister_PersistError(t *testing.T) {
	RegisterMockTestingT(t)
	mockDB := mocks.NewMockDatabase()
	persister := events.NewOutputPersister(mockDB, nil)

	ctx := command.ProjectContext{
		BaseRepo:    models.Repo{FullName: "owner/repo"},
		Pull:        models.PullRequest{Num: 123},
		RepoRelDir:  "terraform/staging",
		Workspace:   "default",
		ProjectName: "myproject",
		User:        models.User{Username: "testuser"},
		CommandName: command.Plan,
	}

	result := command.ProjectResult{
		Command:     command.Plan,
		RepoRelDir:  "terraform/staging",
		Workspace:   "default",
		ProjectName: "myproject",
		ProjectCommandOutput: command.ProjectCommandOutput{
			Error: errors.New("unexpected error"),
		},
	}

	err := persister.PersistResult(ctx, result)
	Ok(t, err)

	mockDB.VerifyWasCalledOnce().SaveProjectOutput(AnyProjectOutput())
}

func TestOutputPersister_PersistPolicyCheck(t *testing.T) {
	RegisterMockTestingT(t)
	mockDB := mocks.NewMockDatabase()
	persister := events.NewOutputPersister(mockDB, nil)

	ctx := command.ProjectContext{
		BaseRepo:    models.Repo{FullName: "owner/repo"},
		Pull:        models.PullRequest{Num: 123},
		RepoRelDir:  "terraform/staging",
		Workspace:   "default",
		ProjectName: "myproject",
		User:        models.User{Username: "testuser"},
		CommandName: command.PolicyCheck,
	}

	result := command.ProjectResult{
		Command:     command.PolicyCheck,
		RepoRelDir:  "terraform/staging",
		Workspace:   "default",
		ProjectName: "myproject",
		ProjectCommandOutput: command.ProjectCommandOutput{
			PolicyCheckResults: &models.PolicyCheckResults{
				PolicySetResults: []models.PolicySetResult{
					{
						PolicySetName: "policy1",
						Passed:        true,
						PolicyOutput:  "All policies passed",
					},
				},
			},
		},
	}

	err := persister.PersistResult(ctx, result)
	Ok(t, err)

	mockDB.VerifyWasCalledOnce().SaveProjectOutput(AnyProjectOutput())
}

func TestOutputPersister_DatabaseError(t *testing.T) {
	RegisterMockTestingT(t)
	mockDB := mocks.NewMockDatabase()
	persister := events.NewOutputPersister(mockDB, nil)

	ctx := command.ProjectContext{
		BaseRepo:    models.Repo{FullName: "owner/repo"},
		Pull:        models.PullRequest{Num: 123},
		RepoRelDir:  "terraform/staging",
		Workspace:   "default",
		ProjectName: "myproject",
		User:        models.User{Username: "testuser"},
		CommandName: command.Plan,
	}

	result := command.ProjectResult{
		Command:     command.Plan,
		RepoRelDir:  "terraform/staging",
		Workspace:   "default",
		ProjectName: "myproject",
		ProjectCommandOutput: command.ProjectCommandOutput{
			PlanSuccess: &models.PlanSuccess{
				TerraformOutput: "Plan: 0 to add, 0 to change, 0 to destroy.",
			},
		},
	}

	// Set up mock to return an error
	When(mockDB.SaveProjectOutput(AnyProjectOutput())).ThenReturn(errors.New("database error"))

	err := persister.PersistResult(ctx, result)
	Assert(t, err != nil, "expected error from database")
	Assert(t, err.Error() == "database error", "expected database error message")
}

func TestOutputPersister_PersistsPullURLAndTitle(t *testing.T) {
	RegisterMockTestingT(t)
	mockDB := mocks.NewMockDatabase()
	persister := events.NewOutputPersister(mockDB, nil)

	ctx := command.ProjectContext{
		BaseRepo: models.Repo{FullName: "owner/repo"},
		Pull: models.PullRequest{
			Num:   456,
			URL:   "https://gitlab.com/owner/repo/-/merge_requests/456",
			Title: "Upgrade database to version 2.0",
		},
		RepoRelDir:  "terraform/prod",
		Workspace:   "production",
		ProjectName: "database",
		User:        models.User{Username: "devuser"},
		CommandName: command.Plan,
		JobID:       "job-456",
	}

	result := command.ProjectResult{
		Command:     command.Plan,
		RepoRelDir:  "terraform/prod",
		Workspace:   "production",
		ProjectName: "database",
		ProjectCommandOutput: command.ProjectCommandOutput{
			PlanSuccess: &models.PlanSuccess{
				TerraformOutput: "Plan: 5 to change.",
			},
		},
	}

	// Capture the output that gets passed to SaveProjectOutput
	var capturedOutput models.ProjectOutput
	When(mockDB.SaveProjectOutput(AnyProjectOutput())).Then(func(params []Param) ReturnValues {
		capturedOutput = params[0].(models.ProjectOutput)
		return []ReturnValue{nil}
	})

	err := persister.PersistResult(ctx, result)
	Ok(t, err)

	// Verify PullURL and PullTitle are captured from ctx.Pull
	Assert(t, capturedOutput.PullURL == "https://gitlab.com/owner/repo/-/merge_requests/456",
		"expected PullURL to be captured from ctx.Pull.URL")
	Assert(t, capturedOutput.PullTitle == "Upgrade database to version 2.0",
		"expected PullTitle to be captured from ctx.Pull.Title")
}

// AnyProjectOutput returns a pegomock matcher for models.ProjectOutput
func AnyProjectOutput() models.ProjectOutput {
	RegisterMatcher(NewAnyMatcher(reflect.TypeFor[models.ProjectOutput]()))
	return models.ProjectOutput{}
}
