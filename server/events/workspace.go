package events

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/hootsuite/atlantis/server/events/models"
	"github.com/hootsuite/atlantis/server/logging"
	"github.com/pkg/errors"
)

const workspacePrefix = "repos"

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_workspace.go Workspace

// Workspace handles the workspace on disk for running commands.
type Workspace interface {
	// Clone git clones headRepo, checks out the branch and then returns the
	// absolute path to the root of the cloned repo.
	Clone(log *logging.SimpleLogger, baseRepo models.Repo, headRepo models.Repo, p models.PullRequest, env string) (string, error)
	// GetWorkspace returns the path to the workspace for this repo and pull.
	GetWorkspace(r models.Repo, p models.PullRequest, env string) (string, error)
	// Delete deletes the workspace for this repo and pull.
	Delete(r models.Repo, p models.PullRequest) error
}

// FileWorkspace implements Workspace with the file system.
type FileWorkspace struct {
	DataDir string
}

// Clone git clones headRepo, checks out the branch and then returns the absolute
// path to the root of the cloned repo.
func (w *FileWorkspace) Clone(
	log *logging.SimpleLogger,
	baseRepo models.Repo,
	headRepo models.Repo,
	p models.PullRequest,
	env string) (string, error) {
	cloneDir := w.cloneDir(baseRepo, p, env)

	// This is safe to do because we lock runs on repo/pull/env so no one else
	// is using this workspace.
	log.Info("cleaning clone directory %q", cloneDir)
	if err := os.RemoveAll(cloneDir); err != nil {
		return "", errors.Wrap(err, "deleting old workspace")
	}

	// Create the directory and parents if necessary.
	log.Info("creating dir %q", cloneDir)
	if err := os.MkdirAll(cloneDir, 0700); err != nil {
		return "", errors.Wrap(err, "creating new workspace")
	}

	log.Info("git cloning %q into %q", headRepo.SanitizedCloneURL, cloneDir)
	cloneCmd := exec.Command("git", "clone", headRepo.CloneURL, cloneDir) // #nosec
	if output, err := cloneCmd.CombinedOutput(); err != nil {
		return "", errors.Wrapf(err, "cloning %s: %s", headRepo.SanitizedCloneURL, string(output))
	}

	// Check out the branch for this PR.
	log.Info("checking out branch %q", p.Branch)
	checkoutCmd := exec.Command("git", "checkout", p.Branch) // #nosec
	checkoutCmd.Dir = cloneDir
	if err := checkoutCmd.Run(); err != nil {
		return "", errors.Wrapf(err, "checking out branch %s", p.Branch)
	}
	return cloneDir, nil
}

// GetWorkspace returns the path to the workspace for this repo and pull.
func (w *FileWorkspace) GetWorkspace(r models.Repo, p models.PullRequest, env string) (string, error) {
	repoDir := w.cloneDir(r, p, env)
	if _, err := os.Stat(repoDir); err != nil {
		return "", errors.Wrap(err, "checking if workspace exists")
	}
	return repoDir, nil
}

// Delete deletes the workspace for this repo and pull.
func (w *FileWorkspace) Delete(r models.Repo, p models.PullRequest) error {
	return os.RemoveAll(w.repoPullDir(r, p))
}

func (w *FileWorkspace) repoPullDir(r models.Repo, p models.PullRequest) string {
	return filepath.Join(w.DataDir, workspacePrefix, r.FullName, strconv.Itoa(p.Num))
}

func (w *FileWorkspace) cloneDir(r models.Repo, p models.PullRequest, env string) string {
	return filepath.Join(w.repoPullDir(r, p), env)
}
