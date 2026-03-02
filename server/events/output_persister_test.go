// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"errors"
	"testing"

	"github.com/runatlantis/atlantis/server/core/db/mocks"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
	"go.uber.org/mock/gomock"
)

func TestOutputPersister_PersistPlanSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockDB := mocks.NewMockDatabase(ctrl)
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

	// Expect lookup for existing stub's RunTimestamp (returns nil = no stub found)
	mockDB.EXPECT().GetProjectOutputByJobID("job-123").Return(nil, nil)
	mockDB.EXPECT().SaveProjectOutput(gomock.Any()).Return(nil)

	err := persister.PersistResult(ctx, result)
	Ok(t, err)
}

func TestOutputPersister_PersistFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockDB := mocks.NewMockDatabase(ctrl)
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

	mockDB.EXPECT().SaveProjectOutput(gomock.Any()).Return(nil)

	err := persister.PersistResult(ctx, result)
	Ok(t, err)
}

func TestOutputPersister_PersistApplySuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockDB := mocks.NewMockDatabase(ctrl)
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

	mockDB.EXPECT().SaveProjectOutput(gomock.Any()).Return(nil)

	err := persister.PersistResult(ctx, result)
	Ok(t, err)
}

func TestOutputPersister_PersistError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockDB := mocks.NewMockDatabase(ctrl)
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

	mockDB.EXPECT().SaveProjectOutput(gomock.Any()).Return(nil)

	err := persister.PersistResult(ctx, result)
	Ok(t, err)
}

func TestOutputPersister_PersistPolicyCheck(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockDB := mocks.NewMockDatabase(ctrl)
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

	mockDB.EXPECT().SaveProjectOutput(gomock.Any()).Return(nil)

	err := persister.PersistResult(ctx, result)
	Ok(t, err)
}

func TestOutputPersister_DatabaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockDB := mocks.NewMockDatabase(ctrl)
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

	mockDB.EXPECT().SaveProjectOutput(gomock.Any()).Return(errors.New("database error"))

	err := persister.PersistResult(ctx, result)
	Assert(t, err != nil, "expected error from database")
	Assert(t, err.Error() == "database error", "expected database error message")
}

func TestOutputPersister_PersistsPullURLAndTitle(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockDB := mocks.NewMockDatabase(ctrl)
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

	// Expect lookup for existing stub's RunTimestamp (returns nil = no stub found)
	mockDB.EXPECT().GetProjectOutputByJobID("job-456").Return(nil, nil)

	// Capture the output that gets passed to SaveProjectOutput
	var capturedOutput models.ProjectOutput
	mockDB.EXPECT().SaveProjectOutput(gomock.Any()).DoAndReturn(func(output models.ProjectOutput) error {
		capturedOutput = output
		return nil
	})

	err := persister.PersistResult(ctx, result)
	Ok(t, err)

	// Verify PullURL and PullTitle are captured from ctx.Pull
	Assert(t, capturedOutput.PullURL == "https://gitlab.com/owner/repo/-/merge_requests/456",
		"expected PullURL to be captured from ctx.Pull.URL")
	Assert(t, capturedOutput.PullTitle == "Upgrade database to version 2.0",
		"expected PullTitle to be captured from ctx.Pull.Title")
}

func TestOutputPersister_ReusesStubTimestamp(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockDB := mocks.NewMockDatabase(ctrl)
	persister := events.NewOutputPersister(mockDB, nil)

	ctx := command.ProjectContext{
		BaseRepo:    models.Repo{FullName: "owner/repo"},
		Pull:        models.PullRequest{Num: 123},
		RepoRelDir:  "terraform/staging",
		Workspace:   "default",
		ProjectName: "myproject",
		User:        models.User{Username: "testuser"},
		CommandName: command.Plan,
		JobID:       "job-789",
	}

	result := command.ProjectResult{
		Command:     command.Plan,
		RepoRelDir:  "terraform/staging",
		Workspace:   "default",
		ProjectName: "myproject",
		ProjectCommandOutput: command.ProjectCommandOutput{
			PlanSuccess: &models.PlanSuccess{
				TerraformOutput: "Plan: 1 to add.",
			},
		},
	}

	// Simulate existing stub with a specific timestamp
	stubTimestamp := int64(1709400000000) // Fixed timestamp from stub
	existingStub := &models.ProjectOutput{
		JobID:        "job-789",
		RunTimestamp: stubTimestamp,
	}

	// Expect lookup to return the existing stub
	mockDB.EXPECT().GetProjectOutputByJobID("job-789").Return(existingStub, nil)

	// Capture the output and verify it uses the stub's timestamp
	var capturedOutput models.ProjectOutput
	mockDB.EXPECT().SaveProjectOutput(gomock.Any()).DoAndReturn(func(output models.ProjectOutput) error {
		capturedOutput = output
		return nil
	})

	err := persister.PersistResult(ctx, result)
	Ok(t, err)

	// CRITICAL: RunTimestamp must match the stub's timestamp for database lookups
	Assert(t, capturedOutput.RunTimestamp == stubTimestamp,
		"expected RunTimestamp to be reused from stub for consistent database lookups")
}
