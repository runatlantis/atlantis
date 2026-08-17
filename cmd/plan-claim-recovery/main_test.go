// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/core/boltdb"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

func TestRunClearsOfflineBoltDBPlanPublicationClaim(t *testing.T) {
	const claimToken = "11111111-1111-4111-8111-111111111111"
	dataDir := t.TempDir()
	pull := models.PullRequest{
		Num: 42,
		BaseRepo: models.Repo{
			FullName: "runatlantis/atlantis",
			VCSHost:  models.VCSHost{Hostname: "github.com"},
		},
	}
	database, err := boltdb.New(dataDir)
	Ok(t, err)
	Ok(t, database.AcquirePlanPublicationClaim(pull, claimToken))
	_, err = database.BeginPlanGeneration(pull, []models.ProjectStatus{{
		Workspace: "default", RepoRelDir: "project-a", ProjectName: "a",
	}}, "generation-a", claimToken)
	Ok(t, err)
	_, err = database.CompletePlanGeneration(pull, "generation-a", []command.ProjectResult{{
		Command: command.Plan, Workspace: "default", RepoRelDir: "project-a", ProjectName: "a",
		ProjectCommandOutput: command.ProjectCommandOutput{
			PlanSuccess: &models.PlanSuccess{}, AtlantisManagedPlan: true, ManagedPlanHash: "hash-a",
		},
	}}, claimToken)
	Ok(t, err)
	Ok(t, database.Close())

	var stdout bytes.Buffer
	Ok(t, run([]string{
		"--data-dir", dataDir,
		"--vcs-hostname", "github.com",
		"--repo", "runatlantis/atlantis",
		"--pull", "42",
		"--confirm-all-replicas-stopped",
		"--inspect",
	}, &stdout, &bytes.Buffer{}))
	Assert(t, strings.Contains(stdout.String(), claimToken), "inspection did not return current token: %s", stdout.String())
	stdout.Reset()
	Ok(t, run([]string{
		"--data-dir", dataDir,
		"--vcs-hostname", "github.com",
		"--repo", "runatlantis/atlantis",
		"--pull", "42",
		"--confirm-all-replicas-stopped",
		"--claim-token", claimToken,
	}, &stdout, &bytes.Buffer{}))
	Assert(t, strings.Contains(stdout.String(), "github.com/runatlantis/atlantis#42"), "unexpected output: %s", stdout.String())

	database, err = boltdb.New(dataDir)
	Ok(t, err)
	t.Cleanup(func() { Ok(t, database.Close()) })
	status, err := database.GetPullStatus(pull)
	Ok(t, err)
	Assert(t, status != nil, "recovery must preserve accepted pull status")
	Equals(t, "generation-a", status.Projects[0].AcceptedPlanGeneration)
	Equals(t, "hash-a", status.Projects[0].ManagedPlanHash)
	Ok(t, database.AcquirePlanPublicationClaim(pull, "recovered-publisher"))
	Ok(t, database.ReleasePlanPublicationClaim(pull, "recovered-publisher"))
}

func TestRunRequiresOfflineRecoveryConfirmation(t *testing.T) {
	err := run([]string{
		"--data-dir", t.TempDir(),
		"--vcs-hostname", "github.com",
		"--repo", "runatlantis/atlantis",
		"--pull", "42",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	Assert(t, err != nil, "expected a recovery error")
	Assert(t, strings.Contains(err.Error(), "--confirm-all-replicas-stopped"), "unexpected error: %v", err)
}

func TestRunRequiresValidExactClaimToken(t *testing.T) {
	tests := []struct {
		name  string
		extra []string
		want  string
	}{
		{name: "missing", want: "--claim-token"},
		{name: "malformed", extra: []string{"--claim-token", "not-a-uuid"}, want: "canonical UUID"},
		{name: "nil uuid", extra: []string{"--claim-token", "00000000-0000-0000-0000-000000000000"}, want: "non-zero"},
		{name: "inspect and clear", extra: []string{"--inspect", "--claim-token", "11111111-1111-4111-8111-111111111111"}, want: "mutually exclusive"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			args := []string{
				"--data-dir", t.TempDir(),
				"--vcs-hostname", "github.com",
				"--repo", "runatlantis/atlantis",
				"--pull", "42",
				"--confirm-all-replicas-stopped",
			}
			args = append(args, test.extra...)
			err := run(args, &bytes.Buffer{}, &bytes.Buffer{})
			Assert(t, err != nil, "expected recovery input to fail")
			Assert(t, strings.Contains(err.Error(), test.want), "unexpected error: %v", err)
		})
	}
}

func TestRunRefusesOpenBoltDB(t *testing.T) {
	const claimToken = "11111111-1111-4111-8111-111111111111"
	dataDir := t.TempDir()
	database, err := boltdb.New(dataDir)
	Ok(t, err)
	t.Cleanup(func() { Ok(t, database.Close()) })
	pull := models.PullRequest{Num: 42, BaseRepo: models.Repo{FullName: "runatlantis/atlantis", VCSHost: models.VCSHost{Hostname: "github.com"}}}
	Ok(t, database.AcquirePlanPublicationClaim(pull, claimToken))

	err = run([]string{
		"--data-dir", dataDir,
		"--vcs-hostname", "github.com",
		"--repo", "runatlantis/atlantis",
		"--pull", "42",
		"--confirm-all-replicas-stopped",
		"--claim-token", claimToken,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	Assert(t, err != nil, "expected recovery to refuse an open BoltDB")
	Assert(t, strings.Contains(err.Error(), "verify all Atlantis replicas"), "unexpected error: %v", err)

	Assert(t, errors.Is(database.AcquirePlanPublicationClaim(pull, "competitor"), db.ErrPlanPublicationBusy), "expected live claim to remain")
}

func TestRunRefusesMissingBoltDB(t *testing.T) {
	dataDir := t.TempDir()
	err := run([]string{
		"--data-dir", filepath.Join(dataDir, "mistyped"),
		"--vcs-hostname", "github.com",
		"--repo", "runatlantis/atlantis",
		"--pull", "42",
		"--confirm-all-replicas-stopped",
		"--claim-token", "11111111-1111-4111-8111-111111111111",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	Assert(t, err != nil, "expected a missing BoltDB error")
	Assert(t, strings.Contains(err.Error(), "finding existing BoltDB"), "unexpected error: %v", err)
	_, statErr := os.Stat(filepath.Join(dataDir, "mistyped", "atlantis.db"))
	Assert(t, errors.Is(statErr, os.ErrNotExist), "recovery must not create a database at a mistyped path")
}

func TestRunDoesNotClearNewerClaimDuringStaleRecovery(t *testing.T) {
	dataDir := t.TempDir()
	pull := models.PullRequest{
		Num: 42,
		BaseRepo: models.Repo{
			FullName: "runatlantis/atlantis",
			VCSHost:  models.VCSHost{Hostname: "github.com"},
		},
	}
	const staleToken = "11111111-1111-4111-8111-111111111111"
	const newerToken = "22222222-2222-4222-8222-222222222222"
	database, err := boltdb.New(dataDir)
	Ok(t, err)
	Ok(t, database.AcquirePlanPublicationClaim(pull, staleToken))
	Ok(t, database.ForceClearPlanPublicationClaim(pull))
	Ok(t, database.AcquirePlanPublicationClaim(pull, newerToken))
	_, err = database.BeginPlanGeneration(pull, []models.ProjectStatus{{
		Workspace: "default", RepoRelDir: "project-a", ProjectName: "a",
	}}, "generation-b", newerToken)
	Ok(t, err)
	_, err = database.CompletePlanGeneration(pull, "generation-b", []command.ProjectResult{{
		Command: command.Plan, Workspace: "default", RepoRelDir: "project-a", ProjectName: "a",
		ProjectCommandOutput: command.ProjectCommandOutput{
			PlanSuccess: &models.PlanSuccess{}, AtlantisManagedPlan: true, ManagedPlanHash: "hash-b",
		},
	}}, newerToken)
	Ok(t, err)
	Ok(t, database.Close())

	err = run([]string{
		"--data-dir", dataDir,
		"--vcs-hostname", "github.com",
		"--repo", "runatlantis/atlantis",
		"--pull", "42",
		"--confirm-all-replicas-stopped",
		"--claim-token", staleToken,
	}, &bytes.Buffer{}, &bytes.Buffer{})

	Assert(t, err != nil, "stale recovery must not clear an unrecognized newer claim")
	database, err = boltdb.New(dataDir)
	Ok(t, err)
	t.Cleanup(func() { Ok(t, database.Close()) })
	Assert(t, errors.Is(database.AcquirePlanPublicationClaim(pull, "competitor"), db.ErrPlanPublicationBusy), "newer claim must remain held")
	status, err := database.GetPullStatus(pull)
	Ok(t, err)
	Assert(t, status != nil, "accepted pull state must remain")
	Equals(t, "generation-b", status.Projects[0].AcceptedPlanGeneration)
	Equals(t, "hash-b", status.Projects[0].ManagedPlanHash)
}

func TestRunRefusesMissingPlanPublicationClaim(t *testing.T) {
	dataDir := t.TempDir()
	database, err := boltdb.New(dataDir)
	Ok(t, err)
	Ok(t, database.Close())

	err = run([]string{
		"--data-dir", dataDir,
		"--vcs-hostname", "github.com",
		"--repo", "runatlantis/atlantis",
		"--pull", "42",
		"--confirm-all-replicas-stopped",
		"--inspect",
	}, &bytes.Buffer{}, &bytes.Buffer{})

	Assert(t, errors.Is(err, db.ErrPlanPublicationNotOwned), "missing claim must fail closed, got %v", err)
}

func TestRunRefusesMalformedPlanPublicationClaim(t *testing.T) {
	dataDir := t.TempDir()
	pull := models.PullRequest{Num: 42, BaseRepo: models.Repo{FullName: "runatlantis/atlantis", VCSHost: models.VCSHost{Hostname: "github.com"}}}
	database, err := boltdb.New(dataDir)
	Ok(t, err)
	Ok(t, database.AcquirePlanPublicationClaim(pull, "not-a-uuid"))
	Ok(t, database.Close())

	err = run([]string{
		"--data-dir", dataDir,
		"--vcs-hostname", "github.com",
		"--repo", "runatlantis/atlantis",
		"--pull", "42",
		"--confirm-all-replicas-stopped",
		"--inspect",
	}, &bytes.Buffer{}, &bytes.Buffer{})

	Assert(t, err != nil, "malformed stored claim must fail closed")
	Assert(t, strings.Contains(err.Error(), "stored plan publication claim is malformed"), "unexpected error: %v", err)
	database, err = boltdb.New(dataDir)
	Ok(t, err)
	t.Cleanup(func() { Ok(t, database.Close()) })
	Assert(t, errors.Is(database.AcquirePlanPublicationClaim(pull, "competitor"), db.ErrPlanPublicationBusy), "malformed claim must remain held")
}
