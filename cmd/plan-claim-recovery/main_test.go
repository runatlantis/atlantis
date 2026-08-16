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
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

func TestRunClearsOfflineBoltDBPlanPublicationClaim(t *testing.T) {
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
	Ok(t, database.AcquirePlanPublicationClaim(pull, "stopped-publisher"))
	Ok(t, database.Close())

	var stdout bytes.Buffer
	Ok(t, run([]string{
		"--data-dir", dataDir,
		"--vcs-hostname", "github.com",
		"--repo", "runatlantis/atlantis",
		"--pull", "42",
		"--confirm-all-replicas-stopped",
	}, &stdout, &bytes.Buffer{}))
	Assert(t, strings.Contains(stdout.String(), "github.com/runatlantis/atlantis#42"), "unexpected output: %s", stdout.String())

	database, err = boltdb.New(dataDir)
	Ok(t, err)
	t.Cleanup(func() { Ok(t, database.Close()) })
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

func TestRunRefusesOpenBoltDB(t *testing.T) {
	dataDir := t.TempDir()
	database, err := boltdb.New(dataDir)
	Ok(t, err)
	t.Cleanup(func() { Ok(t, database.Close()) })
	pull := models.PullRequest{Num: 42, BaseRepo: models.Repo{FullName: "runatlantis/atlantis", VCSHost: models.VCSHost{Hostname: "github.com"}}}
	Ok(t, database.AcquirePlanPublicationClaim(pull, "live-publisher"))

	err = run([]string{
		"--data-dir", dataDir,
		"--vcs-hostname", "github.com",
		"--repo", "runatlantis/atlantis",
		"--pull", "42",
		"--confirm-all-replicas-stopped",
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
	}, &bytes.Buffer{}, &bytes.Buffer{})
	Assert(t, err != nil, "expected a missing BoltDB error")
	Assert(t, strings.Contains(err.Error(), "finding existing BoltDB"), "unexpected error: %v", err)
	_, statErr := os.Stat(filepath.Join(dataDir, "mistyped", "atlantis.db"))
	Assert(t, errors.Is(statErr, os.ErrNotExist), "recovery must not create a database at a mistyped path")
}
