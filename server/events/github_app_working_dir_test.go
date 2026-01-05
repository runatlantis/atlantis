// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"fmt"
	"testing"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/events"
	eventMocks "github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs/github"
	githubMocks "github.com/runatlantis/atlantis/server/events/vcs/github/mocks"
	githubtestdata "github.com/runatlantis/atlantis/server/events/vcs/github/testdata"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

// Test that if we don't have any existing files, we check out the repo with a github app.
func TestClone_GithubAppNoneExisting(t *testing.T) {
	// Initialize the git repo.
	repoDir := initRepo(t)
	expCommit := runCmd(t, repoDir, "git", "rev-parse", "HEAD")

	dataDir := t.TempDir()

	logger := logging.NewNoopLogger(t)

	wd := &events.FileWorkspace{
		DataDir:                     dataDir,
		CheckoutMerge:               false,
		TestingOverrideHeadCloneURL: fmt.Sprintf("file://%s", repoDir),
	}

	defer disableSSLVerification()()
	testServer, err := githubtestdata.GithubAppTestServer(t)
	Ok(t, err)

	gwd := &events.GithubAppWorkingDir{
		WorkingDir: wd,
		Credentials: &github.AppCredentials{
			Key:      []byte(githubtestdata.PrivateKey),
			AppID:    1,
			Hostname: testServer,
		},
		GithubHostname: testServer,
	}

	cloneDir, err := gwd.Clone(logger, models.Repo{}, models.PullRequest{
		BaseRepo:   models.Repo{},
		HeadBranch: "branch",
	}, "default")
	Ok(t, err)

	// Use rev-parse to verify at correct commit.
	actCommit := runCmd(t, cloneDir, "git", "rev-parse", "HEAD")
	Equals(t, expCommit, actCommit)
}

func TestClone_GithubAppSetsCorrectUrl(t *testing.T) {
	logger := logging.NewNoopLogger(t)

	RegisterMockTestingT(t)

	workingDir := eventMocks.NewMockWorkingDir()

	credentials := githubMocks.NewMockCredentials()

	ghAppWorkingDir := events.GithubAppWorkingDir{
		WorkingDir:     workingDir,
		Credentials:    credentials,
		GithubHostname: "some-host",
	}

	baseRepo, _ := models.NewRepo(
		models.Github,
		"runatlantis/atlantis",
		"https://github.com/runatlantis/atlantis.git",

		// user and token have to be blank otherwise this proxy wouldn't be invoked to begin with
		"",
		"",
	)

	headRepo := baseRepo

	modifiedBaseRepo := baseRepo
	// remove credentials from both urls since we want to use the credential store
	modifiedBaseRepo.CloneURL = "https://github.com/runatlantis/atlantis.git"
	modifiedBaseRepo.SanitizedCloneURL = "https://github.com/runatlantis/atlantis.git"

	When(credentials.GetToken()).ThenReturn("token", nil)
	When(workingDir.Clone(Any[logging.SimpleLogging](), Eq(modifiedBaseRepo), Eq(models.PullRequest{BaseRepo: modifiedBaseRepo}),
		Eq("default"))).ThenReturn("", nil)

	_, err := ghAppWorkingDir.Clone(logger, headRepo, models.PullRequest{BaseRepo: baseRepo}, "default")

	workingDir.VerifyWasCalledOnce().Clone(logger, modifiedBaseRepo, models.PullRequest{BaseRepo: modifiedBaseRepo}, "default")

	Ok(t, err)
}

// Similar to `Clone()`
// `MergeAgain()` should set the repo URL correctly
func TestMergeAgain_GithubAppSetsCorrectUrl(t *testing.T) {
	logger := logging.NewNoopLogger(t)

	RegisterMockTestingT(t)

	workingDir := eventMocks.NewMockWorkingDir()

	credentials := githubMocks.NewMockCredentials()

	ghAppWorkingDir := events.GithubAppWorkingDir{
		WorkingDir:     workingDir,
		Credentials:    credentials,
		GithubHostname: "some-host",
	}

	baseRepo, _ := models.NewRepo(
		models.Github,
		"runatlantis/atlantis",
		"https://github.com/runatlantis/atlantis.git",

		// user and token have to be blank otherwise this proxy wouldn't be invoked to begin with
		"",
		"",
	)

	headRepo := baseRepo

	modifiedBaseRepo := baseRepo
	// remove credentials from both urls since we want to use the credential store
	modifiedBaseRepo.CloneURL = "https://github.com/runatlantis/atlantis.git"
	modifiedBaseRepo.SanitizedCloneURL = "https://github.com/runatlantis/atlantis.git"

	When(credentials.GetToken()).ThenReturn("token", nil)
	When(workingDir.MergeAgain(Any[logging.SimpleLogging](), Eq(modifiedBaseRepo), Eq(models.PullRequest{BaseRepo: modifiedBaseRepo}),
		Eq("default"))).ThenReturn(false, nil)

	_, err := ghAppWorkingDir.MergeAgain(logger, headRepo, models.PullRequest{BaseRepo: baseRepo}, "default")

	// MergeAgain
	workingDir.VerifyWasCalledOnce().MergeAgain(logger, modifiedBaseRepo, models.PullRequest{BaseRepo: modifiedBaseRepo}, "default")

	Ok(t, err)
}
