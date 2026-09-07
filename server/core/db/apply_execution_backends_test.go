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

func TestApplyExecution_BackendParity(t *testing.T) {
	for _, backend := range []string{"bolt", "redis"} {
		t.Run(backend, func(t *testing.T) {
			var storage db.Database
			if backend == "bolt" {
				b, err := boltdb.New(t.TempDir())
				Ok(t, err)
				storage = b
			} else {
				server := miniredis.RunT(t)
				port, err := strconv.Atoi(server.Port())
				Ok(t, err)
				b, err := redisdb.New(server.Host(), port, "", false, false, 0)
				Ok(t, err)
				storage = b
			}
			t.Cleanup(func() { Ok(t, storage.Close()) })
			pull := models.PullRequest{Num: 1, HeadCommit: "same-head", BaseRepo: models.Repo{FullName: "owner/repo"}}
			project := generationProject("a")
			_, err := storage.BeginPlanGeneration(pull, "G1", []command.ProjectContext{project}, true, command.NoClaim{})
			Ok(t, err)
			accepted, err := storage.UpdatePullWithResults(pull, []command.ProjectResult{generationResult("a", "G1")}, command.NoClaim{})
			Ok(t, err)
			project.PlanGeneration, project.AcceptedPlanGeneration, project.ExpectedPlanHash = "G1", "G1", accepted.Projects[0].ManagedPlanHash
			running, err := storage.BeginApplyExecution(pull, []command.ProjectContext{project}, "first", command.NoClaim{})
			Ok(t, err)
			_, err = storage.BeginApplyExecution(pull, []command.ProjectContext{project}, "second", command.NoClaim{})
			Assert(t, errors.Is(err, db.ErrApplyAlreadyStarted), "duplicate admission: %v", err)
			// PR updates removing projects cannot erase an unresolved execution, even
			// when the update changes the head and normally replaces the entire status.
			for _, replace := range []bool{false, true} {
				nextPull := pull
				nextPull.HeadCommit = "new-head"
				_, err = storage.BeginPlanGeneration(nextPull, "removed", nil, replace, command.NoClaim{})
				Assert(t, errors.Is(err, db.ErrApplyAlreadyStarted), "removed reservation: %v", err)
			}
			unchanged, err := storage.GetPullStatus(pull)
			Ok(t, err)
			Equals(t, running, *unchanged)
			excluded := command.ProjectResult{Command: command.Apply, Workspace: project.Workspace, RepoRelDir: project.RepoRelDir, ProjectName: project.ProjectName, PlanGeneration: "G1", ApplyExecutionID: "wrong", ExcludeFromPlanStatus: true, ProjectCommandOutput: command.ProjectCommandOutput{Error: errors.New("directory missing")}}
			_, err = storage.UpdatePullWithResults(pull, []command.ProjectResult{excluded}, command.NoClaim{})
			Assert(t, errors.Is(err, db.ErrApplyAlreadyStarted), "wrong owner released reservation: %v", err)
			excluded.ApplyExecutionID = "first"
			restored, err := storage.UpdatePullWithResults(pull, []command.ProjectResult{excluded}, command.NoClaim{})
			Ok(t, err)
			Equals(t, accepted, restored)
			// A fresh command can now start; a failed execution cannot be retried even
			// after a subsequent successful plan replaces the accepted generation.
			_, err = storage.BeginApplyExecution(pull, []command.ProjectContext{project}, "second", command.NoClaim{})
			Ok(t, err)
			failure := excluded
			failure.ExcludeFromPlanStatus = false
			failure.ApplyExecutionID = "second"
			failure.ApplyAttempted = true
			_, err = storage.UpdatePullWithResults(pull, []command.ProjectResult{failure}, command.NoClaim{})
			Ok(t, err)
			_, err = storage.BeginPlanGeneration(pull, "G2", []command.ProjectContext{project}, true, command.NoClaim{})
			Ok(t, err)
			_, err = storage.UpdatePullWithResults(pull, []command.ProjectResult{generationResult("a", "G2")}, command.NoClaim{})
			Ok(t, err)
			project.PlanGeneration, project.AcceptedPlanGeneration = "G2", "G2"
			_, err = storage.BeginApplyExecution(pull, []command.ProjectContext{project}, "third", command.NoClaim{})
			Assert(t, errors.Is(err, db.ErrApplyAlreadyStarted), "replan replayed unknown execution: %v", err)
			// An obsolete owner that never executed may release its own reservation
			// without overwriting the newer accepted plan or returning false success.
			beforeRelease, err := storage.GetPullStatus(pull)
			Ok(t, err)
			failure.ApplyAttempted = false
			_, err = storage.UpdatePullWithResults(pull, []command.ProjectResult{failure}, command.NoClaim{})
			Assert(t, errors.Is(err, db.ErrApplyExecutionSuperseded), "obsolete release must remain visible: %v", err)
			afterRelease, err := storage.GetPullStatus(pull)
			Ok(t, err)
			beforeRelease.Projects[0].ApplyExecutionID = ""
			Equals(t, *beforeRelease, *afterRelease)
			_, err = storage.BeginApplyExecution(pull, []command.ProjectContext{project}, "third", command.NoClaim{})
			Ok(t, err)
			_, err = storage.BeginPlanGeneration(pull, "G3", []command.ProjectContext{project}, true, command.NoClaim{})
			Ok(t, err)
			missingPlan := generationResult("a", "G3")
			missingPlan.ExcludeFromPlanStatus, missingPlan.Error = true, errors.New("directory removed")
			_, err = storage.UpdatePullWithResults(pull, []command.ProjectResult{missingPlan}, command.NoClaim{})
			Assert(t, errors.Is(err, db.ErrApplyAlreadyStarted), "excluded plan erased running execution: %v", err)
			retained, err := storage.GetPullStatus(pull)
			Ok(t, err)
			Equals(t, "third", retained.Projects[0].ApplyExecutionID)
			_, err = storage.DiscardPlanStatus(pull, retained.Projects[0], command.NoClaim{})
			Ok(t, err)
			_, err = storage.BeginPlanGeneration(pull, "G4", []command.ProjectContext{project}, true, command.NoClaim{})
			Ok(t, err)
			_, err = storage.UpdatePullWithResults(pull, []command.ProjectResult{generationResult("a", "G4")}, command.NoClaim{})
			Ok(t, err)
			project.PlanGeneration, project.AcceptedPlanGeneration = "G4", "G4"
			start, outcomes := make(chan struct{}), make(chan error, 2)
			for _, owner := range []string{"replica-a", "replica-b"} {
				go func() {
					<-start
					_, err := storage.BeginApplyExecution(pull, []command.ProjectContext{project}, owner, command.NoClaim{})
					outcomes <- err
				}()
			}
			close(start)
			winners := 0
			for range 2 {
				if err := <-outcomes; err == nil {
					winners++
				} else {
					Assert(t, errors.Is(err, db.ErrApplyAlreadyStarted), "unexpected admission failure: %v", err)
				}
			}
			Equals(t, 1, winners)
		})
	}
}
