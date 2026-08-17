// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

// plan-claim-recovery clears one durable BoltDB plan-publication claim while
// Atlantis is offline. It is intentionally a separate administrative utility:
// clearing a claim while its publisher is alive can allow stale VCS writes.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/runatlantis/atlantis/server/core/boltdb"
	"github.com/runatlantis/atlantis/server/events/models"
)

type recoveryOptions struct {
	dataDir                   string
	vcsHostname               string
	repoFullName              string
	pullNum                   int
	confirmAllReplicasStopped bool
	inspect                   bool
	claimToken                string
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) (retErr error) {
	flags := flag.NewFlagSet("plan-claim-recovery", flag.ContinueOnError)
	flags.SetOutput(stderr)
	var opts recoveryOptions
	flags.StringVar(&opts.dataDir, "data-dir", "~/.atlantis", "Atlantis data directory containing atlantis.db")
	flags.StringVar(&opts.vcsHostname, "vcs-hostname", "", "VCS hostname, for example github.com")
	flags.StringVar(&opts.repoFullName, "repo", "", "repository full name, for example runatlantis/atlantis")
	flags.IntVar(&opts.pullNum, "pull", 0, "pull request number")
	flags.BoolVar(&opts.confirmAllReplicasStopped, "confirm-all-replicas-stopped", false, "confirm no Atlantis replica can still publish for this pull")
	flags.BoolVar(&opts.inspect, "inspect", false, "print the current offline claim token without clearing it")
	flags.StringVar(&opts.claimToken, "claim-token", "", "exact claim token previously obtained with --inspect")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %s", strings.Join(flags.Args(), " "))
	}
	if !opts.confirmAllReplicasStopped {
		return errors.New("refusing recovery without --confirm-all-replicas-stopped")
	}
	if opts.vcsHostname == "" {
		return errors.New("--vcs-hostname is required")
	}
	if opts.repoFullName == "" {
		return errors.New("--repo is required")
	}
	if opts.pullNum <= 0 {
		return errors.New("--pull must be greater than zero")
	}
	if opts.inspect && opts.claimToken != "" {
		return errors.New("--inspect and --claim-token are mutually exclusive")
	}
	if !opts.inspect && opts.claimToken == "" {
		return errors.New("refusing recovery without --claim-token; run with --inspect first")
	}
	if opts.claimToken != "" {
		if err := validateClaimToken(opts.claimToken); err != nil {
			return fmt.Errorf("--claim-token is not a valid canonical UUID: %w", err)
		}
	}

	dataDir, err := expandHome(opts.dataDir)
	if err != nil {
		return err
	}
	databasePath := filepath.Join(dataDir, "atlantis.db")
	databaseInfo, err := os.Stat(databasePath)
	if err != nil {
		return fmt.Errorf("finding existing BoltDB at %q: %w", databasePath, err)
	}
	if !databaseInfo.Mode().IsRegular() {
		return fmt.Errorf("existing BoltDB path %q is not a regular file", databasePath)
	}
	database, err := boltdb.New(dataDir)
	if err != nil {
		return fmt.Errorf("opening BoltDB; verify all Atlantis replicas using this database are stopped: %w", err)
	}
	defer func() {
		retErr = errors.Join(retErr, database.Close())
	}()

	pull := models.PullRequest{
		Num: opts.pullNum,
		BaseRepo: models.Repo{
			FullName: opts.repoFullName,
			VCSHost:  models.VCSHost{Hostname: opts.vcsHostname},
		},
	}
	currentToken, err := database.GetPlanPublicationClaim(pull)
	if err != nil {
		return fmt.Errorf("inspecting plan publication claim: %w", err)
	}
	if err := validateClaimToken(currentToken); err != nil {
		return fmt.Errorf("stored plan publication claim is malformed; refusing recovery: %w", err)
	}
	if opts.inspect {
		_, err = fmt.Fprintf(stdout, "offline plan publication claim for %s/%s#%d: %s\n", opts.vcsHostname, opts.repoFullName, opts.pullNum, currentToken)
		return err
	}
	if err := database.ReleasePlanPublicationClaim(pull, opts.claimToken); err != nil {
		return fmt.Errorf("clearing exact plan publication claim: %w", err)
	}
	_, err = fmt.Fprintf(stdout, "cleared offline plan publication claim for %s/%s#%d\n", opts.vcsHostname, opts.repoFullName, opts.pullNum)
	return err
}

func validateClaimToken(token string) error {
	parsed, err := uuid.Parse(token)
	if err != nil {
		return err
	}
	if parsed == uuid.Nil || parsed.String() != token {
		return errors.New("claim token must be a non-zero lowercase UUID in canonical form")
	}
	return nil
}

func expandHome(path string) (string, error) {
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return filepath.Abs(path)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	if path == "~" {
		return home, nil
	}
	return filepath.Join(home, strings.TrimPrefix(path, "~/")), nil
}
