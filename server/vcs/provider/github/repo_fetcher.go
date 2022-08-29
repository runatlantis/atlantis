package github

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const workingDirPrefix = "repos"

// RepoFetcher implements repoFetcher through git clone operations
type RepoFetcher struct {
	DataDir        string
	Token          string
	GithubHostname string
	Logger         logging.Logger
}

func (g *RepoFetcher) Fetch(baseRepo models.Repo, sha string, destinationPath string) error {
	home, err := homedir.Dir()
	if err != nil {
		return errors.Wrap(err, "getting home dir to write ~/.git-credentials file")
	}

	// https://developer.github.com/apps/building-github-apps/authenticating-with-github-apps/#http-based-git-access-by-an-installation
	if err := events.WriteGitCreds("x-access-token", g.Token, g.GithubHostname, home, g.Logger, true); err != nil {
		return err
	}
	// Realistically, this is a super brittle way of supporting clones using gh app installation tokens
	// This URL should be built during Repo creation and the struct should be immutable going forward.
	// Doing this requires a larger refactor however, and can probably be coupled with supporting > 1 installation
	authURL := fmt.Sprintf("://x-access-token:%s", g.Token)
	baseRepo.CloneURL = strings.Replace(baseRepo.CloneURL, "://:", authURL, 1)
	baseRepo.SanitizedCloneURL = strings.Replace(baseRepo.SanitizedCloneURL, "://:", "://x-access-token:", 1)
	return g.clone(baseRepo, sha, destinationPath)
}

func (g *RepoFetcher) clone(baseRepo models.Repo, sha string, destinationPath string) error {
	// Create the directory and parents if necessary.
	if err := os.MkdirAll(destinationPath, 0700); err != nil {
		return errors.Wrap(err, "creating new directory")
	}

	// Fetch default branch into clone directory
	cloneCmd := []string{"git", "clone", "--branch", baseRepo.DefaultBranch, "--single-branch", baseRepo.CloneURL, destinationPath}
	_, err := g.run(cloneCmd, destinationPath)
	if err != nil {
		return errors.New("failed to clone directory")
	}

	// Return immediately if commit at HEAD of clone matches request commit
	revParseCmd := []string{"git", "rev-parse", "HEAD"}
	revParseOutput, err := g.run(revParseCmd, destinationPath)
	currCommit := strings.Trim(string(revParseOutput), "\n")
	if strings.HasPrefix(currCommit, sha) {
		return nil
	}

	// Otherwise, checkout the correct sha
	checkoutCmd := []string{"git", "checkout", sha}
	_, err = g.run(checkoutCmd, destinationPath)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to checkout to sha: %s", sha))
	}
	return nil
}

func (g *RepoFetcher) run(args []string, destinationPath string) ([]byte, error) {
	cmd := exec.Command(args[0], args[1:]...) // nolint: gosec
	cmd.Dir = destinationPath
	// The repo merge command requires these env vars are set.
	cmd.Env = append(os.Environ(), []string{
		"EMAIL=atlantis@runatlantis.io",
		"GIT_AUTHOR_NAME=atlantis",
		"GIT_COMMITTER_NAME=atlantis",
	}...)
	return cmd.CombinedOutput()
}

func (g *RepoFetcher) Cleanup(filePath string) error {
	return os.RemoveAll(filePath)
}

func (g *RepoFetcher) GenerateDirPath(repoName string) string {
	return filepath.Join(g.DataDir, workingDirPrefix, repoName, uuid.New().String())
}
