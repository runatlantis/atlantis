package events_test

import (
	"fmt"
	"testing"

	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/vcs/fixtures"
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

	cloneDir, _, err := gwd.Clone(nil, models.Repo{}, models.Repo{}, models.PullRequest{
		HeadBranch: "branch",
	}, "default")
	Ok(t, err)

	// Use rev-parse to verify at correct commit.
	actCommit := runCmd(t, cloneDir, "git", "rev-parse", "HEAD")
	Equals(t, expCommit, actCommit)
}
