package events_test

import (
	"fmt"
	"testing"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/events"
	eventMocks "github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	vcsMocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/events/vcs/testdata"
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
		Logger:                      logger,
	}

	defer disableSSLVerification()()
	testServer, err := testdata.GithubAppTestServer(t)
	Ok(t, err)

	gwd := &events.GithubAppWorkingDir{
		WorkingDir: wd,
		Credentials: &vcs.GithubAppCredentials{
			Key:      []byte(testdata.GithubPrivateKey),
			AppID:    1,
			Hostname: testServer,
		},
		GithubHostname: testServer,
	}

	cloneDir, _, err := gwd.Clone(models.Repo{}, models.PullRequest{
		BaseRepo:   models.Repo{},
		HeadBranch: "branch",
	}, "default")
	Ok(t, err)

	// Use rev-parse to verify at correct commit.
	actCommit := runCmd(t, cloneDir, "git", "rev-parse", "HEAD")
	Equals(t, expCommit, actCommit)
}

func TestClone_GithubAppSetsCorrectUrl(t *testing.T) {
	RegisterMockTestingT(t)

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
	// remove credentials from both urls since we want to use the credential store
	modifiedBaseRepo.CloneURL = "https://github.com/runatlantis/atlantis.git"
	modifiedBaseRepo.SanitizedCloneURL = "https://github.com/runatlantis/atlantis.git"

	When(credentials.GetToken()).ThenReturn("token", nil)
	When(workingDir.Clone(modifiedBaseRepo, models.PullRequest{BaseRepo: modifiedBaseRepo}, "default")).ThenReturn(
		"", true, nil,
	)

	_, success, _ := ghAppWorkingDir.Clone(headRepo, models.PullRequest{BaseRepo: baseRepo}, "default")

	workingDir.VerifyWasCalledOnce().Clone(modifiedBaseRepo, models.PullRequest{BaseRepo: modifiedBaseRepo}, "default")

	Assert(t, success == true, "clone url mutation error")
}
