// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	multierror "github.com/hashicorp/go-multierror"
)

var defaultAtlantisURL = "http://localhost:4141"

func getVCSClient() (VCSClient, error) {
	if os.Getenv("ATLANTIS_GH_USER") != "" || os.Getenv("ATLANTIS_GH_APP_ID") != "" {
		log.Print("Running tests for github")
		return NewGithubClient(), nil
	}
	if os.Getenv("ATLANTIS_GITLAB_USER") != "" {
		log.Print("Running tests for gitlab")
		return NewGitlabClient(), nil
	}

	return nil, errors.New("could not determine which vcs client")
}

func isGitHub() bool {
	return os.Getenv("ATLANTIS_GH_USER") != "" || os.Getenv("ATLANTIS_GH_APP_ID") != ""
}

func main() {
	atlantisURL := os.Getenv("ATLANTIS_URL")
	if atlantisURL == "" {
		atlantisURL = defaultAtlantisURL
	}
	atlantisURL = fmt.Sprintf("%s/events", atlantisURL)

	cloneDirRoot := os.Getenv("CLONE_DIR")
	if cloneDirRoot == "" {
		cloneDirRoot = "/tmp/atlantis-tests"
	}

	log.Printf("cleaning workspace %s", cloneDirRoot)
	if err := os.RemoveAll(cloneDirRoot); err != nil {
		log.Fatalf("failed to clean dir %q: %v", cloneDirRoot, err)
	}

	vcsClient, err := getVCSClient()
	if err != nil {
		log.Fatalf("failed to get vcs client: %v", err)
	}

	ctx := context.Background()
	log.Printf("creating atlantis webhook with %s url", atlantisURL)
	hookID, err := vcsClient.CreateAtlantisWebhook(ctx, atlantisURL)
	if err != nil {
		log.Fatalf("error creating atlantis webhook: %v", err)
	}

	cases := activeCases()
	log.Printf("running %d test cases", len(cases))

	results, err := runCases(ctx, vcsClient, hookID, cloneDirRoot, cases)

	// Print results summary.
	log.Printf("\nTest Results\n---------------------------")
	for _, r := range results {
		status := r.testResult
		if r.err != nil {
			status = fmt.Sprintf("FAIL: %v", r.err)
		}
		fmt.Printf("  %-35s %s\n", r.testCase, status)
		if r.pullRequestURL != "" {
			fmt.Printf("  %-35s PR: %s\n", "", r.pullRequestURL)
		}
	}
	fmt.Println("---------------------------")

	if err != nil {
		log.Fatalf("%v", err)
	}
}

func activeCases() []TestCase {
	optIn := os.Getenv("E2E_OPT_IN") == "1"
	gh := isGitHub()

	var active []TestCase
	for _, tc := range testCases {
		switch tc.Status {
		case CaseDisabled:
			log.Printf("skipping disabled case %q: %s", tc.Name, tc.SkipReason)
			continue
		case CaseOptIn:
			if !optIn {
				log.Printf("skipping opt-in case %q (set E2E_OPT_IN=1): %s", tc.Name, tc.SkipReason)
				continue
			}
		}

		switch tc.VCS {
		case VCSGitHub:
			if !gh {
				continue
			}
		case VCSGitLab:
			if gh {
				continue
			}
		}

		active = append(active, tc)
	}
	return active
}

func runCases(ctx context.Context, vcsClient VCSClient, hookID int64, cloneDirRoot string, cases []TestCase) ([]*E2EResult, error) {
	var results []*E2EResult
	var testErrors *multierror.Error

	defer vcsClient.DeleteAtlantisHook(ctx, hookID) // nolint: errcheck

	for _, tc := range cases {
		log.Printf("━━━ starting: %s ━━━", tc.Name)
		e2e := &E2ETester{
			vcsClient:    vcsClient,
			hookID:       hookID,
			cloneDirRoot: cloneDirRoot,
			testCase:     tc,
		}
		result, err := e2e.Start(ctx)
		if err != nil {
			result.err = err
			log.Printf("━━━ FAILED: %s — %v ━━━", tc.Name, err)
		} else {
			log.Printf("━━━ passed: %s ━━━", tc.Name)
		}
		results = append(results, result)
		testErrors = multierror.Append(testErrors, err)
	}

	return results, testErrors.ErrorOrNil()
}
