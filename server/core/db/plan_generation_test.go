// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package db_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

func generationProject(name string) command.ProjectContext {
	return command.ProjectContext{Workspace: "default", RepoRelDir: ".", ProjectName: name, RequiresAtlantisManagedPlanFile: true}
}
func generationResult(name, generation string) command.ProjectResult {
	return command.ProjectResult{Command: command.Plan, Workspace: "default", RepoRelDir: ".", ProjectName: name, PlanGeneration: generation, ManagedPlanHash: strings.Repeat("a", 64), ProjectCommandOutput: command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}}}
}
func TestPlanGeneration_SameHeadSupersession(t *testing.T) {
	pull := models.PullRequest{HeadCommit: "same-head", BaseBranch: "main"}
	g1, err := db.BeginPlanGeneration(nil, pull, "G1", []command.ProjectContext{generationProject("a")}, true)
	Ok(t, err)
	Equals(t, models.ErroredPlanStatus, g1.Projects[0].Status)
	g2, err := db.BeginPlanGeneration(&g1.PullStatus, pull, "G2", []command.ProjectContext{generationProject("a")}, true)
	Ok(t, err)
	accepted, err := db.MergePullResults(&g2.PullStatus, pull, []command.ProjectResult{generationResult("a", "G2")})
	Ok(t, err)
	Equals(t, "G2", accepted.Projects[0].AcceptedPlanGeneration)
	after, err := db.MergePullResults(&accepted, pull, []command.ProjectResult{generationResult("a", "G1")})
	Assert(t, errors.Is(err, db.ErrPlanGenerationSuperseded), "expected stale generation rejection: %v", err)
	Equals(t, accepted, after)
}
func TestPlanGeneration_OrdinaryWritersPreserveActiveGeneration(t *testing.T) {
	pull := models.PullRequest{HeadCommit: "head"}
	active, err := db.BeginPlanGeneration(nil, pull, "G1", []command.ProjectContext{generationProject("a")}, true)
	Ok(t, err)
	for _, head := range []string{"head", "old-head"} {
		writePull := pull
		writePull.HeadCommit = head
		result := generationResult("a", "")
		after, err := db.MergePullResults(&active.PullStatus, writePull, []command.ProjectResult{result})
		Assert(t, errors.Is(err, db.ErrPlanGenerationSuperseded), "expected unfenced writer rejection: %v", err)
		Equals(t, active.PullStatus, after)
	}
}
func TestPlanGeneration_CompletionRequiresSavedDigest(t *testing.T) {
	pull := models.PullRequest{HeadCommit: "head"}
	active, err := db.BeginPlanGeneration(nil, pull, "G1", []command.ProjectContext{generationProject("a")}, true)
	Ok(t, err)
	result := generationResult("a", "G1")
	result.ManagedPlanHash = ""
	after, err := db.MergePullResults(&active.PullStatus, pull, []command.ProjectResult{result})
	Assert(t, errors.Is(err, db.ErrPlanGenerationInvalid), "missing saved digest must fail: %v", err)
	Equals(t, active.PullStatus, after)
	result.Error = errors.New("upload failed")
	after, err = db.MergePullResults(&active.PullStatus, pull, []command.ProjectResult{result})
	Ok(t, err)
	Equals(t, "", after.Projects[0].AcceptedPlanGeneration)
	Equals(t, "", after.Projects[0].ManagedPlanHash)
	Equals(t, models.ErroredPlanStatus, after.Projects[0].Status)
}
func TestPlanGeneration_MixedObsoleteApplyInvalidatesSuccessfulExecution(t *testing.T) {
	pull := models.PullRequest{HeadCommit: "head"}
	active, err := db.BeginPlanGeneration(nil, pull, "G2", []command.ProjectContext{generationProject("a"), generationProject("b")}, true)
	Ok(t, err)
	accepted, err := db.MergePullResults(&active.PullStatus, pull, []command.ProjectResult{generationResult("a", "G2"), generationResult("b", "G2")})
	Ok(t, err)
	a := generationResult("a", "G2")
	a.Command = command.Apply
	a.ApplySuccess = "infrastructure changed"
	b := generationResult("b", "G1")
	b.Command = command.Apply
	b.Error = errors.New("policy check failed")
	for _, results := range [][]command.ProjectResult{{a, b}, {b, a}} {
		after, err := db.MergePullResults(&accepted, pull, results)
		Assert(t, errors.Is(err, db.ErrApplyExecutionAmbiguous), "must surface already executed apply: %v", err)
		Equals(t, models.ErroredApplyStatus, after.Projects[0].Status)
		Equals(t, "", after.Projects[0].AcceptedPlanGeneration)
		Equals(t, "", after.Projects[0].ManagedPlanHash)
		Equals(t, accepted.Projects[1], after.Projects[1])
		Equals(t, "G2", accepted.Projects[0].AcceptedPlanGeneration)
	}
}
func TestPlanGeneration_AdmissionPreservesPoliciesAndUntargetedProjects(t *testing.T) {
	pull := models.PullRequest{HeadCommit: "head"}
	policies := []models.PolicySetStatus{{PolicySetName: "sticky"}}
	before := models.PullStatus{Pull: pull, Projects: []models.ProjectStatus{
		{Workspace: "default", RepoRelDir: ".", ProjectName: "a", PolicyStatus: policies, Status: models.AppliedPlanStatus},
		{Workspace: "default", RepoRelDir: ".", ProjectName: "sibling", Status: models.PlannedPlanStatus},
	}}
	after, err := db.BeginPlanGeneration(&before, pull, "G1", []command.ProjectContext{generationProject("a")}, false)
	Ok(t, err)
	Equals(t, policies, after.Projects[0].PolicyStatus)
	Equals(t, before.Projects[1], after.Projects[1])
	Equals(t, before.Projects, after.Previous)
}

func TestPlanGeneration_FailedCompletionCannotLaterSucceed(t *testing.T) {
	pull := models.PullRequest{HeadCommit: "head"}
	active, err := db.BeginPlanGeneration(nil, pull, "G1", []command.ProjectContext{generationProject("a")}, true)
	Ok(t, err)
	failed := generationResult("a", "G1")
	failed.Error = errors.New("saving artifact")
	completed, err := db.MergePullResults(&active.PullStatus, pull, []command.ProjectResult{failed})
	Ok(t, err)
	Assert(t, !completed.Projects[0].PlanGenerationActive, "failed generation must be terminal")
	after, err := db.MergePullResults(&completed, pull, []command.ProjectResult{generationResult("a", "G1")})
	Assert(t, errors.Is(err, db.ErrPlanGenerationInvalid), "late success must not revive failed generation: %v", err)
	Equals(t, completed, after)
}

func TestPlanGeneration_TargetedSupersessionCancelsUnfinishedCommand(t *testing.T) {
	pull := models.PullRequest{HeadCommit: "same-head"}
	active, err := db.BeginPlanGeneration(nil, pull, "G1", []command.ProjectContext{generationProject("a"), generationProject("b")}, true)
	Ok(t, err)
	next, err := db.BeginPlanGeneration(&active.PullStatus, pull, "G2", []command.ProjectContext{generationProject("a")}, false)
	Ok(t, err)
	Assert(t, next.Projects[0].PlanGenerationActive, "new generation must be active")
	Assert(t, !next.Projects[1].PlanGenerationActive, "old command's sibling must be cancelled")
	Equals(t, models.ErroredPlanStatus, next.Projects[1].Status)
	Equals(t, "", next.Projects[1].AcceptedPlanGeneration)
	accepted, err := db.MergePullResults(&next.PullStatus, pull, []command.ProjectResult{generationResult("a", "G2")})
	Ok(t, err)
	after, err := db.MergePullResults(&accepted, pull, []command.ProjectResult{generationResult("b", "G1")})
	Assert(t, errors.Is(err, db.ErrPlanGenerationInvalid), "cancelled sibling must not complete: %v", err)
	Equals(t, accepted, after)
}

func TestPlanGeneration_RemediationAfterFailedPlan(t *testing.T) {
	for _, cmd := range []command.Name{command.Import, command.State, command.Unlock} {
		t.Run(cmd.String(), func(t *testing.T) {
			pull := models.PullRequest{HeadCommit: "head"}
			active, err := db.BeginPlanGeneration(nil, pull, "G1", []command.ProjectContext{generationProject("a")}, true)
			Ok(t, err)
			remediation := generationResult("a", "G1")
			remediation.Command = cmd
			_, err = db.MergePullResults(&active.PullStatus, pull, []command.ProjectResult{remediation})
			Assert(t, errors.Is(err, db.ErrPlanGenerationSuperseded), "remediation must not race active plan: %v", err)
			failed := generationResult("a", "G1")
			failed.Error = errors.New("plan failed")
			completed, err := db.MergePullResults(&active.PullStatus, pull, []command.ProjectResult{failed})
			Ok(t, err)
			after, err := db.MergePullResults(&completed, pull, []command.ProjectResult{remediation})
			Ok(t, err)
			Equals(t, models.DiscardedPlanStatus, after.Projects[0].Status)
			Equals(t, "", after.Projects[0].AcceptedPlanGeneration)
		})
	}
}

func TestPlanGeneration_LegacyRecoveryCannotErasePartialIdentity(t *testing.T) {
	for _, raw := range []string{
		`{"Projects":[{"PlanGeneration":"G1"}]}`,
		`{"Projects":[{"AcceptedPlanGeneration":"G1"}]}`,
		`{"Projects":[{"ManagedPlanHash":"digest"}]}`,
		`{"Projects":[{"PlanGenerationActive":true}]}`,
		`{"Projects":`,
	} {
		Assert(t, !db.LegacyPullStatus([]byte(raw)), "must not recover generation data as legacy: %s", raw)
	}
	Assert(t, db.LegacyPullStatus([]byte(`{"Projects":[{"Status":"old-schema"}]}`)), "legacy schema recovery remains supported")
}

func TestPlanGeneration_PostExecutionFailureInvalidatesMatchingGeneration(t *testing.T) {
	pull := models.PullRequest{HeadCommit: "head"}
	active, err := db.BeginPlanGeneration(nil, pull, "G1", []command.ProjectContext{generationProject("a")}, true)
	Ok(t, err)
	accepted, err := db.MergePullResults(&active.PullStatus, pull, []command.ProjectResult{generationResult("a", "G1")})
	Ok(t, err)
	result := generationResult("a", "G1")
	result.Command = command.Apply
	result.ApplyExecuted = true
	result.ApplySuccess = ""
	result.Error = db.ErrApplyExecutionAmbiguous
	after, err := db.MergePullResults(&accepted, pull, []command.ProjectResult{result})
	Assert(t, errors.Is(err, db.ErrApplyExecutionAmbiguous), "execution with failed final validation must be explicit: %v", err)
	Equals(t, models.ErroredApplyStatus, after.Projects[0].Status)
	Equals(t, "", after.Projects[0].AcceptedPlanGeneration)
	Equals(t, "", after.Projects[0].ManagedPlanHash)
}
