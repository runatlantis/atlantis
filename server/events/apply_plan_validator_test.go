// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

// plannedProjectResultWithHash is a successful plan result whose plan step
// recorded planHash for the plan file it wrote.
func plannedProjectResultWithHash(repoRelDir, workspace, projectName, planHash string) command.ProjectResult {
	res := plannedProjectResult(repoRelDir, workspace, projectName)
	res.PlanSuccess.PlanHash = planHash
	return res
}

func recordedHashApplyContext(t *testing.T) command.ProjectContext {
	t.Helper()
	return command.ProjectContext{
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
}

// A plan file replaced after `atlantis plan` is not the plan that was reviewed,
// even when the apply-time hash taken from disk agrees with the file.
func TestApplyPlanValidator_RejectsPlanDifferingFromRecordedHashAndDeletesIt(t *testing.T) {
	db := newTestBoltDB(t)
	repoDir := t.TempDir()
	ctx := recordedHashApplyContext(t)
	replaced := []byte("replaced plan")
	ctx.ExpectedPlanHash = planHashForContent(replaced)
	_, err := db.UpdatePullWithResults(ctx.Pull, []command.ProjectResult{
		plannedProjectResultWithHash(ctx.RepoRelDir, ctx.Workspace, ctx.ProjectName, planHashForContent([]byte("reviewed plan"))),
	})
	Ok(t, err)
	planPath := filepath.Join(repoDir, runtime.GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	Ok(t, os.WriteFile(planPath, replaced, 0600))
	validator := &events.DefaultApplyPlanValidator{PullStatusFetcher: db}

	err = validator.ValidateProjectPlan(ctx, repoDir)

	Assert(t, err != nil, "expected recorded plan hash mismatch")
	Assert(t, strings.Contains(err.Error(), "does not match the plan that was recorded"), "got: %s", err)
	_, statErr := os.Stat(planPath)
	Assert(t, os.IsNotExist(statErr), "expected rejected plan file to be deleted, stat err: %v", statErr)
}

func TestApplyPlanValidator_AcceptsPlanMatchingRecordedHash(t *testing.T) {
	db := newTestBoltDB(t)
	repoDir := t.TempDir()
	ctx := recordedHashApplyContext(t)
	reviewed := []byte("reviewed plan")
	_, err := db.UpdatePullWithResults(ctx.Pull, []command.ProjectResult{
		plannedProjectResultWithHash(ctx.RepoRelDir, ctx.Workspace, ctx.ProjectName, planHashForContent(reviewed)),
	})
	Ok(t, err)
	planPath := filepath.Join(repoDir, runtime.GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	Ok(t, os.WriteFile(planPath, reviewed, 0600))
	validator := &events.DefaultApplyPlanValidator{PullStatusFetcher: db}

	err = validator.ValidateProjectPlan(ctx, repoDir)

	Ok(t, err)
	_, statErr := os.Stat(planPath)
	Ok(t, statErr)
}

// Statuses written before plan hashes were recorded, and workflows whose plan
// stage manages its own plan file, carry no hash; they are still accepted.
func TestApplyPlanValidator_SkipsRecordedHashCheckWhenNoneRecorded(t *testing.T) {
	db := newTestBoltDB(t)
	repoDir := t.TempDir()
	ctx := recordedHashApplyContext(t)
	_, err := db.UpdatePullWithResults(ctx.Pull, []command.ProjectResult{
		plannedProjectResult(ctx.RepoRelDir, ctx.Workspace, ctx.ProjectName),
	})
	Ok(t, err)
	planPath := filepath.Join(repoDir, runtime.GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	Ok(t, os.WriteFile(planPath, []byte("unrecorded plan"), 0600))
	validator := &events.DefaultApplyPlanValidator{PullStatusFetcher: db}

	err = validator.ValidateProjectPlan(ctx, repoDir)

	Ok(t, err)
	recorded, err := validator.RecordedPlanHash(ctx)
	Ok(t, err)
	Equals(t, "", recorded)
}

func TestApplyPlanValidator_RecordedPlanHashReturnsStoredHash(t *testing.T) {
	db := newTestBoltDB(t)
	ctx := recordedHashApplyContext(t)
	want := planHashForContent([]byte("reviewed plan"))
	_, err := db.UpdatePullWithResults(ctx.Pull, []command.ProjectResult{
		plannedProjectResultWithHash(ctx.RepoRelDir, ctx.Workspace, ctx.ProjectName, want),
	})
	Ok(t, err)
	validator := &events.DefaultApplyPlanValidator{PullStatusFetcher: db}

	got, err := validator.RecordedPlanHash(ctx)

	Ok(t, err)
	Equals(t, want, got)
}
