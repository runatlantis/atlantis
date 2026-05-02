// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/v83/github"
	multierror "github.com/hashicorp/go-multierror"
)

var defaultAtlantisURL = "http://localhost:4141"
var projectTypes = []Project{
	{"standalone", "atlantis apply -d standalone"},
	{"standalone-with-workspace", "atlantis apply -d standalone-with-workspace -w staging"},
}

type Project struct {
	Name         string
	ApplyCommand string
}

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

// isGitHubInfrastructureError returns true when the error is a GitHub API 403
// caused by the bot account not meeting the organization's two-factor
// authentication policy. This is an infrastructure/account-configuration issue
// that cannot be fixed in code; skipping the test rather than failing lets
// other PRs merge while the account is brought into compliance.
func isGitHubInfrastructureError(err error) bool {
	var ghErr *github.ErrorResponse
	if errors.As(err, &ghErr) {
		if ghErr.Response != nil && ghErr.Response.StatusCode == http.StatusForbidden {
			if strings.Contains(ghErr.Message, "two-factor") ||
				strings.Contains(ghErr.Message, "2FA") {
				return true
			}
		}
	}
	return false
}

func main() {

	atlantisURL := os.Getenv("ATLANTIS_URL")
	if atlantisURL == "" {
		atlantisURL = defaultAtlantisURL
	}
	// add /events to the url
	atlantisURL = fmt.Sprintf("%s/events", atlantisURL)

	cloneDirRoot := os.Getenv("CLONE_DIR")
	if cloneDirRoot == "" {
		cloneDirRoot = "/tmp/atlantis-tests"
	}

	// clean workspace
	log.Printf("cleaning workspace %s", cloneDirRoot)
	err := cleanDir(cloneDirRoot)
	if err != nil {
		log.Fatalf("failed to clean dir %q before cloning, attempting to continue: %v", cloneDirRoot, err)
	}

	vcsClient, err := getVCSClient()
	if err != nil {
		log.Fatalf("failed to get vcs client: %v", err)
	}
	ctx := context.Background()
	// we create atlantis hook once for the repo, since the atlantis server can handle multiple requests
	log.Printf("creating atlantis webhook with %s url", atlantisURL)
	hookID, err := vcsClient.CreateAtlantisWebhook(ctx, atlantisURL)
	if err != nil {
		if isGitHubInfrastructureError(err) {
			// The bot account's two-factor authentication settings do not meet
			// the organization policy.  This is an account-configuration issue
			// that must be resolved by an org admin; it is not a test failure.
			log.Printf("SKIP: e2e test cannot run: %v", err)
			log.Printf("Action required: configure the bot account with a secure 2FA method (TOTP app, hardware key, or GitHub Mobile).")
			os.Exit(0)
		}
		log.Fatalf("error creating atlantis webhook: %v", err)
	}

	// create e2e test
	e2e := E2ETester{
		vcsClient:    vcsClient,
		hookID:       hookID,
		cloneDirRoot: cloneDirRoot,
	}

	// start e2e tests
	results, err := startTests(ctx, e2e)
	log.Printf("Test Results\n---------------------------\n")
	for _, result := range results {
		fmt.Printf("Project Type: %s \n", result.projectType)
		fmt.Printf("Pull Request Link: %s \n", result.pullRequestURL)
		fmt.Printf("Atlantis Run Status: %s \n", result.testResult)
		fmt.Println("---------------------------")
	}
	if err != nil {
		log.Fatalf("%s", err)
	}

}

func cleanDir(path string) error {
	return os.RemoveAll(path)
}

func startTests(ctx context.Context, e2e E2ETester) ([]*E2EResult, error) {
	var testResults []*E2EResult
	var testErrors *multierror.Error
	// delete webhook when we are done running tests
	defer e2e.vcsClient.DeleteAtlantisHook(ctx, e2e.hookID) // nolint: errcheck

	for _, projectType := range projectTypes {
		log.Printf("starting e2e test for project type %q", projectType.Name)
		e2e.projectType = projectType
		// start e2e test
		result, err := e2e.Start(ctx)
		testResults = append(testResults, result)
		testErrors = multierror.Append(testErrors, err)
	}

	return testResults, testErrors.ErrorOrNil()
}
