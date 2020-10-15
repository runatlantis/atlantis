package events_test

import (
	"fmt"
	"testing"

	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/events"
	eventMocks "github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/vcs/fixtures"
	vcsMocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	. "github.com/runatlantis/atlantis/testing"
)

// Test that if we don't have any existing files, we check out the repo with a github app.
func TestClone_GithubAppNoneExisting(t *testing.T) {
	// Initialize the git repo.
	repoDir, cleanup := initRepo(t)
	defer cleanup()
	expCommit := runCmd(t, repoDir, "git", "rev-parse", "HEAD")

	dataDir, cleanup2 := TempDir(t)
	defer cleanup2()

	wd := &events.FileWorkspace{
		DataDir:                     dataDir,
		CheckoutMerge:               false,
		TestingOverrideHeadCloneURL: fmt.Sprintf("file://%s", repoDir),
	}

	tmpDir, cleanup3 := DirStructure(t, map[string]interface{}{
		"key.pem": fixtures.GithubPrivateKey,
	})
	defer cleanup3()

	defer disableSSLVerification()()
	testServer, err := fixtures.GithubAppTestServer(t)
	Ok(t, err)

	gwd := &events.GithubAppWorkingDir{
		WorkingDir: wd,
		Credentials: &vcs.GithubAppCredentials{
			KeyPath:  fmt.Sprintf("%v/key.pem", tmpDir),
			AppID:    1,
			Hostname: testServer,
		},
		GithubHostname: testServer,
	}

	cloneDir, _, err := gwd.Clone(nil, models.Repo{}, models.PullRequest{
		BaseRepo:   models.Repo{},
		HeadBranch: "branch",
	}, "default")
	Ok(t, err)

	// Use rev-parse to verify at correct commit.
	actCommit := runCmd(t, cloneDir, "git", "rev-parse", "HEAD")
	Equals(t, expCommit, actCommit)
}

func TestClone_GithubAppSetsCorrectUrl(t *testing.T) {
	workingDir := eventMocks.NewMockWorkingDir()

	credentials := vcsMocks.NewMockGithubCredentials()

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
	modifiedBaseRepo.CloneURL = "https://x-access-token:token@github.com/runatlantis/atlantis.git"
	modifiedBaseRepo.SanitizedCloneURL = "https://x-access-token:<redacted>@github.com/runatlantis/atlantis.git"

	When(credentials.GetToken()).ThenReturn("token", nil)
	When(workingDir.Clone(nil, modifiedBaseRepo, models.PullRequest{BaseRepo: modifiedBaseRepo}, "default")).ThenReturn(
		"", true, nil,
	)

	_, success, _ := ghAppWorkingDir.Clone(nil, headRepo, models.PullRequest{BaseRepo: baseRepo}, "default")

	Assert(t, success == true, "clone url mutation error")
}
