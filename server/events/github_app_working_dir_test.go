package events_test

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"testing"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	vcsMocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/events/vcs/testdata"
	"github.com/runatlantis/atlantis/server/events/workspace"
	workspaceMocks "github.com/runatlantis/atlantis/server/events/workspace/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

// Helper functions for Github app tests
func runCmdGithubApp(t *testing.T, dir string, name string, args ...string) string {
	t.Helper()
	cpCmd := exec.Command(name, args...)
	cpCmd.Dir = dir
	cpOut, err := cpCmd.CombinedOutput()
	Assert(t, err == nil, "err running %q: %s", strings.Join(append([]string{name}, args...), " "), cpOut)
	return string(cpOut)
}

func initRepoGithubApp(t *testing.T) string {
	repoDir := t.TempDir()
	runCmdGithubApp(t, repoDir, "git", "init", "--initial-branch=main")
	runCmdGithubApp(t, repoDir, "touch", ".gitkeep")
	runCmdGithubApp(t, repoDir, "git", "add", ".gitkeep")
	runCmdGithubApp(t, repoDir, "git", "config", "--local", "user.email", "atlantisbot@runatlantis.io")
	runCmdGithubApp(t, repoDir, "git", "config", "--local", "user.name", "atlantisbot")
	runCmdGithubApp(t, repoDir, "git", "config", "--local", "commit.gpgsign", "false")
	runCmdGithubApp(t, repoDir, "git", "commit", "-m", "initial commit")
	runCmdGithubApp(t, repoDir, "git", "checkout", "-b", "branch")
	return repoDir
}

func disableSSLVerificationGithubApp() func() {
	orig := http.DefaultTransport.(*http.Transport).TLSClientConfig
	// nolint: gosec
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	return func() {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = orig
	}
}

// Test that if we don't have any existing files, we check out the repo with a github app.
func TestClone_GithubAppNoneExisting(t *testing.T) {
	// Initialize the git repo.
	repoDir := initRepoGithubApp(t)
	expCommit := runCmdGithubApp(t, repoDir, "git", "rev-parse", "HEAD")

	dataDir := t.TempDir()

	logger := logging.NewNoopLogger(t)

	wd := &workspace.FileWorkspace{
		DataDir:                     dataDir,
		CheckoutMerge:               false,
		TestingOverrideHeadCloneURL: fmt.Sprintf("file://%s", repoDir),
	}

	defer disableSSLVerificationGithubApp()()
	testServer, err := testdata.GithubAppTestServer(t)
	Ok(t, err)

	gwd := &workspace.GithubAppWorkingDir{
		WorkingDir: wd,
		Credentials: &vcs.GithubAppCredentials{
			Key:      []byte(testdata.GithubPrivateKey),
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

	workingDir := workspaceMocks.NewMockWorkingDir()

	credentials := vcsMocks.NewMockGithubCredentials()

	ghAppWorkingDir := workspace.GithubAppWorkingDir{
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

	workingDir := workspaceMocks.NewMockWorkingDir()

	credentials := vcsMocks.NewMockGithubCredentials()

	ghAppWorkingDir := workspace.GithubAppWorkingDir{
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
