package github_test

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs/fixtures"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/vcs/provider/github"
	"github.com/stretchr/testify/assert"
)

type testTokenGetter struct {
	token string
	err   error
}

func (t *testTokenGetter) GetToken() (string, error) {
	return t.token, t.err
}

// Fetch from provided branch
func TestFetchSimple(t *testing.T) {
	// Initialize the git repo.
	repoDir, cleanupRepo := initRepo(t)
	defer cleanupRepo()
	sha1 := appendCommit(t, repoDir, ".gitkeep", "initial commit")
	_ = runCmd(t, repoDir, "git", "rev-parse", "HEAD")
	sha2 := appendCommit(t, repoDir, ".gitignore", "second commit")
	expCommit2 := runCmd(t, repoDir, "git", "rev-parse", "HEAD")

	dataDir, cleanupDataDir := tempDir(t)
	defer cleanupDataDir()
	defer disableSSLVerification()()
	testServer, err := fixtures.GithubAppTestServer(t)
	assert.NoError(t, err)
	logger := logging.NewNoopCtxLogger(t)
	fetcher := &github.RepoFetcher{
		DataDir:           dataDir,
		GithubHostname:    testServer,
		Logger:            logger,
		GithubCredentials: &testTokenGetter{},
	}
	repo := newBaseRepo(repoDir)
	options := github.RepoFetcherOptions{ShallowClone: false}
	destinationPath, _, err := fetcher.Fetch(context.Background(), repo, repo.DefaultBranch, sha2, options)
	assert.NoError(t, err)

	// Use rev-parse to verify at correct commit.
	actCommit := runCmd(t, destinationPath, "git", "rev-parse", "HEAD")
	assert.Equal(t, expCommit2, actCommit)

	cpCmd := exec.Command("git", "checkout", sha1)
	cpCmd.Dir = destinationPath
	_, err = cpCmd.CombinedOutput()
	assert.NoError(t, err)
}

// Fetch from provided branch
func TestFetchSimple_Shallow(t *testing.T) {
	// Initialize the git repo.
	repoDir, cleanupRepo := initRepo(t)
	defer cleanupRepo()
	sha1 := appendCommit(t, repoDir, ".gitkeep", "initial commit")
	_ = runCmd(t, repoDir, "git", "rev-parse", "HEAD")
	sha2 := appendCommit(t, repoDir, ".gitignore", "second commit")
	expCommit2 := runCmd(t, repoDir, "git", "rev-parse", "HEAD")

	dataDir, cleanupDataDir := tempDir(t)
	defer cleanupDataDir()
	defer disableSSLVerification()()
	testServer, err := fixtures.GithubAppTestServer(t)
	assert.NoError(t, err)
	logger := logging.NewNoopCtxLogger(t)
	fetcher := &github.RepoFetcher{
		DataDir:           dataDir,
		GithubHostname:    testServer,
		Logger:            logger,
		GithubCredentials: &testTokenGetter{},
	}
	repo := newBaseRepo(repoDir)
	options := github.RepoFetcherOptions{ShallowClone: true}
	destinationPath, _, err := fetcher.Fetch(context.Background(), repo, repo.DefaultBranch, sha2, options)
	assert.NoError(t, err)

	// Use rev-parse to verify at correct commit.
	actCommit := runCmd(t, destinationPath, "git", "rev-parse", "HEAD")
	assert.Equal(t, expCommit2, actCommit)

	cpCmd := exec.Command("git", "checkout", sha1)
	cpCmd.Dir = destinationPath
	_, err = cpCmd.CombinedOutput()
	assert.Error(t, err)
}

// Fetch with checkout to correct sha needed
func TestFetchCheckout(t *testing.T) {
	// Initialize the git repo.
	repoDir, cleanupRepo := initRepo(t)
	defer cleanupRepo()

	// Add first commit
	sha1 := appendCommit(t, repoDir, ".gitkeep", "initial commit")
	expCommit := runCmd(t, repoDir, "git", "rev-parse", "HEAD")

	// Add second commit
	_ = appendCommit(t, repoDir, ".gitignore", "second commit")

	dataDir, cleanupDataDir := tempDir(t)
	defer cleanupDataDir()
	defer disableSSLVerification()()
	testServer, err := fixtures.GithubAppTestServer(t)
	assert.NoError(t, err)
	logger := logging.NewNoopCtxLogger(t)
	fetcher := &github.RepoFetcher{
		DataDir:           dataDir,
		GithubHostname:    testServer,
		Logger:            logger,
		GithubCredentials: &testTokenGetter{},
	}
	repo := newBaseRepo(repoDir)
	options := github.RepoFetcherOptions{ShallowClone: false}
	destinationPath, _, err := fetcher.Fetch(context.Background(), repo, repo.DefaultBranch, sha1, options)
	assert.NoError(t, err)

	// Use rev-parse to verify at correct commit.
	actCommit := runCmd(t, destinationPath, "git", "rev-parse", "HEAD")
	assert.Equal(t, expCommit, actCommit)
}

// Fetch a non default branch
func TestFetchNonDefaultBranch(t *testing.T) {
	// Initialize the git repo.
	repoDir, cleanupRepo := initRepo(t)
	defer cleanupRepo()

	// Create branch and append commit
	runCmd(t, repoDir, "git", "checkout", "-b", "test-branch")
	sha := appendCommit(t, repoDir, ".gitkeep", "initial commit")
	expCommit := runCmd(t, repoDir, "git", "rev-parse", "HEAD")

	dataDir, cleanupDataDir := tempDir(t)
	defer cleanupDataDir()
	defer disableSSLVerification()()
	testServer, err := fixtures.GithubAppTestServer(t)
	assert.NoError(t, err)
	logger := logging.NewNoopCtxLogger(t)
	fetcher := &github.RepoFetcher{
		DataDir:           dataDir,
		GithubHostname:    testServer,
		Logger:            logger,
		GithubCredentials: &testTokenGetter{},
	}
	repo := newBaseRepo(repoDir)
	options := github.RepoFetcherOptions{ShallowClone: false}
	destinationPath, _, err := fetcher.Fetch(context.Background(), repo, "test-branch", sha, options)
	assert.NoError(t, err)

	// Use rev-parse to verify at correct commit.
	actCommit := runCmd(t, destinationPath, "git", "rev-parse", "HEAD")
	assert.Equal(t, expCommit, actCommit)
}

// Validate if clone is not possible, we return a failure
func TestFetchSimpleFailure(t *testing.T) {
	// Initialize the git repo.
	repoDir, cleanupRepo := initRepo(t)
	defer cleanupRepo()
	sha := appendCommit(t, repoDir, ".gitkeep", "initial commit")

	dataDir, cleanupDataDir := tempDir(t)
	defer cleanupDataDir()
	defer disableSSLVerification()()
	testServer, err := fixtures.GithubAppTestServer(t)
	assert.NoError(t, err)
	logger := logging.NewNoopCtxLogger(t)
	fetcher := &github.RepoFetcher{
		DataDir:           dataDir,
		GithubHostname:    testServer,
		Logger:            logger,
		GithubCredentials: &testTokenGetter{},
	}
	repo := newBaseRepo(repoDir)
	repo.DefaultBranch = "invalid-branch"
	options := github.RepoFetcherOptions{ShallowClone: false}
	_, _, err = fetcher.Fetch(context.Background(), repo, repo.DefaultBranch, sha, options)
	assert.Error(t, err)
}

// Validate if we do not find requested sha, we don't checkout
func TestFetchCheckoutFailure(t *testing.T) {
	// Initialize the git repo.
	repoDir, cleanupRepo := initRepo(t)
	defer cleanupRepo()

	// Add first commit
	_ = appendCommit(t, repoDir, ".gitkeep", "initial commit")

	// Add second commit
	_ = appendCommit(t, repoDir, ".gitignore", "second commit")

	dataDir, cleanupDataDir := tempDir(t)
	defer cleanupDataDir()
	defer disableSSLVerification()()
	testServer, err := fixtures.GithubAppTestServer(t)
	assert.NoError(t, err)
	logger := logging.NewNoopCtxLogger(t)
	fetcher := &github.RepoFetcher{
		DataDir:           dataDir,
		GithubHostname:    testServer,
		Logger:            logger,
		GithubCredentials: &testTokenGetter{},
	}
	repo := newBaseRepo(repoDir)
	options := github.RepoFetcherOptions{ShallowClone: false}
	_, _, err = fetcher.Fetch(context.Background(), repo, repo.DefaultBranch, "invalidsha", options)
	assert.Error(t, err)
}

// Fetch with correct branch
func TestFetchNonDefaultBranchFailure(t *testing.T) {
	// Initialize the git repo.
	repoDir, cleanupRepo := initRepo(t)
	defer cleanupRepo()

	// Create branch and append commit
	runCmd(t, repoDir, "git", "checkout", "-b", "test-branch")
	sha := appendCommit(t, repoDir, ".gitkeep", "initial commit")

	dataDir, cleanupDataDir := tempDir(t)
	defer cleanupDataDir()
	defer disableSSLVerification()()
	testServer, err := fixtures.GithubAppTestServer(t)
	assert.NoError(t, err)
	logger := logging.NewNoopCtxLogger(t)
	fetcher := &github.RepoFetcher{
		DataDir:           dataDir,
		GithubHostname:    testServer,
		Logger:            logger,
		GithubCredentials: &testTokenGetter{},
	}
	repo := newBaseRepo(repoDir)
	options := github.RepoFetcherOptions{ShallowClone: false}
	_, _, err = fetcher.Fetch(context.Background(), repo, "invalid-branch", sha, options)
	assert.Error(t, err)
}

func TestFetch_ErrorGettingGHToken(t *testing.T) {
	// Initialize the git repo.
	repoDir, cleanupRepo := initRepo(t)
	defer cleanupRepo()
	sha := appendCommit(t, repoDir, ".gitkeep", "initial commit")

	dataDir, cleanupDataDir := tempDir(t)
	defer cleanupDataDir()
	defer disableSSLVerification()()
	testServer, err := fixtures.GithubAppTestServer(t)
	assert.NoError(t, err)
	logger := logging.NewNoopCtxLogger(t)
	fetcher := &github.RepoFetcher{
		DataDir:        dataDir,
		GithubHostname: testServer,
		Logger:         logger,
		GithubCredentials: &testTokenGetter{
			err: errors.New("error"),
		},
	}
	repo := newBaseRepo(repoDir)
	options := github.RepoFetcherOptions{ShallowClone: false}
	_, _, err = fetcher.Fetch(context.Background(), repo, repo.DefaultBranch, sha, options)
	assert.ErrorContains(t, err, "error")
}

func newBaseRepo(repoDir string) models.Repo {
	return models.Repo{
		VCSHost: models.VCSHost{
			Hostname: "github.com",
		},
		FullName:      "nish/repo",
		DefaultBranch: "branch",
		CloneURL:      fmt.Sprintf("file://%s", repoDir),
	}
}

func initRepo(t *testing.T) (string, func()) {
	repoDir, cleanup := tempDir(t)
	runCmd(t, repoDir, "git", "init")
	return repoDir, cleanup
}

func appendCommit(t *testing.T, repoDir string, fileName string, commitMessage string) string {
	runCmd(t, repoDir, "touch", fileName)
	runCmd(t, repoDir, "git", "add", fileName)
	runCmd(t, repoDir, "git", "config", "--local", "user.email", "atlantisbot@runatlantis.io")
	runCmd(t, repoDir, "git", "config", "--local", "user.name", "atlantisbot")
	output := runCmd(t, repoDir, "git", "commit", "-m", commitMessage)
	// hacky regex to fetch commit sha
	re := regexp.MustCompile(`\w+]`)
	commitResult := re.FindString(output)
	sha := commitResult[:len(commitResult)-1]
	runCmd(t, repoDir, "git", "branch", "branch", "-f")
	return sha
}

func runCmd(t *testing.T, dir string, name string, args ...string) string {
	t.Helper()
	cpCmd := exec.Command(name, args...)
	cpCmd.Dir = dir
	cpOut, err := cpCmd.CombinedOutput()
	assert.NoError(t, err, "err running %q: %s", strings.Join(append([]string{name}, args...), " "), cpOut)
	return string(cpOut)
}

func tempDir(t *testing.T) (string, func()) {
	tmpDir := t.TempDir()
	return tmpDir, func() {
		os.RemoveAll(tmpDir) // nolint: errcheck
	}
}

// disableSSLVerification disables ssl verification for the global http client
// and returns a function to be called in a defer that will re-enable it.
func disableSSLVerification() func() {
	orig := http.DefaultTransport.(*http.Transport).TLSClientConfig
	// nolint: gosec
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	return func() {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = orig
	}
}
