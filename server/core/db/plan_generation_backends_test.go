// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package db_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/runatlantis/atlantis/server/core/boltdb"
	"github.com/runatlantis/atlantis/server/core/db"
	redisdb "github.com/runatlantis/atlantis/server/core/redis"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

type generationBackend interface {
	DiscardPlanStatus(models.PullRequest, models.ProjectStatus) (bool, error)
	UpdateProjectStatus(models.PullRequest, string, string, models.ProjectPlanStatus) error

	BeginPlanGeneration(models.PullRequest, string, []command.ProjectContext, bool) (db.PlanGenerationBeginResult, error)
	UpdatePullWithResults(models.PullRequest, []command.ProjectResult) (models.PullStatus, error)
	GetPullStatus(models.PullRequest) (*models.PullStatus, error)
	DeletePullStatus(models.PullRequest) error
	Close() error
}

func TestPlanGeneration_BackendParity(t *testing.T) {
	backends := map[string]func(*testing.T) generationBackend{
		"bolt": func(t *testing.T) generationBackend { b, err := boltdb.New(t.TempDir()); Ok(t, err); return b },
		"redis": func(t *testing.T) generationBackend {
			s := miniredis.RunT(t)
			port, err := strconv.Atoi(s.Port())
			Ok(t, err)
			b, err := redisdb.New(s.Host(), port, "", false, false, 0)
			Ok(t, err)
			return b
		},
	}
	for name, newBackend := range backends {
		t.Run(name, func(t *testing.T) {
			b := newBackend(t)
			t.Cleanup(func() { Ok(t, b.Close()) })
			pull := models.PullRequest{Num: 1, HeadCommit: "same-head", BaseRepo: models.Repo{FullName: "owner/repo"}}
			_, err := b.BeginPlanGeneration(pull, "G1", []command.ProjectContext{generationProject("a")}, true)
			Ok(t, err)
			_, err = b.BeginPlanGeneration(pull, "G2", []command.ProjectContext{generationProject("a")}, true)
			Ok(t, err)
			accepted, err := b.UpdatePullWithResults(pull, []command.ProjectResult{generationResult("a", "G2")})
			Ok(t, err)
			_, err = b.UpdatePullWithResults(pull, []command.ProjectResult{generationResult("a", "G1")})
			Assert(t, errors.Is(err, db.ErrPlanGenerationSuperseded), "expected old completion rejection: %v", err)
			actual, err := b.GetPullStatus(pull)
			Ok(t, err)
			Equals(t, accepted, *actual)
			// An old ordinary writer must neither complete nor erase this generation.
			_, err = b.UpdatePullWithResults(pull, []command.ProjectResult{generationResult("a", "")})
			Assert(t, errors.Is(err, db.ErrPlanGenerationSuperseded), "expected legacy writer rejection: %v", err)
			actual, err = b.GetPullStatus(pull)
			Ok(t, err)
			Equals(t, accepted, *actual)
			// Unfenced legacy status mutations cannot bypass the generation.
			err = b.UpdateProjectStatus(pull, "default", ".", models.DiscardedPlanStatus)
			Assert(t, errors.Is(err, db.ErrPlanGenerationSuperseded), "legacy status mutation must reject: %v", err)
			wrong := accepted.Projects[0]
			wrong.PlanGeneration = "G1"
			_, err = b.DiscardPlanStatus(pull, wrong)
			Assert(t, errors.Is(err, db.ErrPlanGenerationSuperseded), "discard requires exact identity: %v", err)
			// Infrastructure changes from an obsolete execution invalidate acceptance
			// durably, even though the adapter returns a terminal error to the caller.
			applied := generationResult("a", "G1")
			applied.Command = command.Apply
			applied.ApplySuccess = "changed infrastructure"
			_, err = b.UpdatePullWithResults(pull, []command.ProjectResult{applied})
			Assert(t, errors.Is(err, db.ErrApplyExecutionAmbiguous), "expected ambiguous execution: %v", err)
			actual, err = b.GetPullStatus(pull)
			Ok(t, err)
			Equals(t, "", actual.Projects[0].AcceptedPlanGeneration)
			Equals(t, "", actual.Projects[0].ManagedPlanHash)
			Equals(t, models.ErroredApplyStatus, actual.Projects[0].Status)
			// Replanning is explicit reconciliation. Applied terminal history is
			// preserved when a later UI unlock requests a discard.
			_, err = b.BeginPlanGeneration(pull, "G3", []command.ProjectContext{generationProject("a")}, true)
			Ok(t, err)
			_, err = b.UpdatePullWithResults(pull, []command.ProjectResult{generationResult("a", "G3")})
			Ok(t, err)
			applied.PlanGeneration = "G3"
			_, err = b.UpdatePullWithResults(pull, []command.ProjectResult{applied})
			Ok(t, err)
			actual, err = b.GetPullStatus(pull)
			Ok(t, err)
			discarded, err := b.DiscardPlanStatus(pull, actual.Projects[0])
			Ok(t, err)
			Assert(t, !discarded, "unlock after apply is not a discard")
			actual, err = b.GetPullStatus(pull)
			Ok(t, err)
			Equals(t, models.AppliedPlanStatus, actual.Projects[0].Status)
			// Close cleanup removes status. A late generation cannot resurrect it.
			Ok(t, b.DeletePullStatus(pull))
			_, err = b.UpdatePullWithResults(pull, []command.ProjectResult{generationResult("a", "G2")})
			Assert(t, errors.Is(err, db.ErrPlanGenerationSuperseded), "expected closed pull rejection: %v", err)
			actual, err = b.GetPullStatus(pull)
			Ok(t, err)
			Assert(t, actual == nil, "late completion must not recreate deleted status")
			_, err = b.DiscardPlanStatus(pull, models.ProjectStatus{Workspace: "default", RepoRelDir: ".", ProjectName: "a"})
			Assert(t, errors.Is(err, db.ErrPlanGenerationSuperseded), "missing durable plan must not report discarded: %v", err)
		})
	}
}
