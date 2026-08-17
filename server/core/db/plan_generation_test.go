// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package db_test

import (
	"errors"
	"testing"

	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

func TestValidatePlanGenerationCompletionRejectsNonPlanResult(t *testing.T) {
	pull := models.PullRequest{HeadCommit: "abc123", BaseBranch: "main"}
	status := &models.PullStatus{
		Pull: pull,
		Projects: []models.ProjectStatus{{
			Workspace:      "default",
			RepoRelDir:     "infra",
			ProjectName:    "app",
			Status:         models.ErroredPlanStatus,
			PlanGeneration: "generation-a",
		}},
	}

	err := db.ValidatePlanGenerationCompletion(status, pull, "generation-a", []command.ProjectResult{{
		Command:     command.PolicyCheck,
		Workspace:   "default",
		RepoRelDir:  "infra",
		ProjectName: "app",
	}})

	Assert(t, errors.Is(err, db.ErrPlanGenerationStateInvalid), "got: %v", err)
}

func TestPlanPublicationClaimToken(t *testing.T) {
	tests := []struct {
		name        string
		claimTokens []string
		want        string
		wantErr     bool
	}{
		{name: "legacy caller", want: ""},
		{name: "one token", claimTokens: []string{"owner-a"}, want: "owner-a"},
		{name: "explicit empty token", claimTokens: []string{""}, want: ""},
		{name: "multiple tokens", claimTokens: []string{"owner-a", "owner-b"}, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := db.PlanPublicationClaimToken(test.claimTokens)
			if test.wantErr {
				Assert(t, errors.Is(err, db.ErrPlanPublicationNotOwned), "got: %v", err)
				var claimErr *db.PlanPublicationClaimError
				Assert(t, errors.As(err, &claimErr), "expected PlanPublicationClaimError, got: %T", err)
				return
			}

			Ok(t, err)
			Equals(t, test.want, got)
		})
	}
}
