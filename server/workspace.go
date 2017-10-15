package server

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/hootsuite/atlantis/server/models"
	"github.com/pkg/errors"
	"github.com/hootsuite/atlantis/server/logging"
	"os/exec"
)

const workspacePrefix = "repos"

//go:generate pegomock generate --use-experimental-model-gen --package mocks -o mocks/mock_workspace.go Workspace

type Workspace interface {
	Clone(log *logging.SimpleLogger, baseRepo models.Repo, headRepo models.Repo, p models.PullRequest, env string) (string, error)
	GetWorkspace(r models.Repo, p models.PullRequest, env string) (string, error)
	Delete(r models.Repo, p models.PullRequest) error
}

type FileWorkspace struct {
	dataDir string
	sshKey  string
}

func (w *FileWorkspace) Clone(
	log *logging.SimpleLogger,
	baseRepo models.Repo,
	headRepo models.Repo,
	p models.PullRequest,
	env string) (string, error) {
	cloneDir := w.cloneDir(baseRepo, p, env)

	// this is safe to do because we lock runs on repo/pull/env so no one else is using this workspace
	log.Info("cleaning clone directory %q", cloneDir)
	if err := os.RemoveAll(cloneDir); err != nil {
		return "", errors.Wrap(err, "deleting old workspace")
	}

	// create the directory and parents if necessary
	log.Info("creating dir %q", cloneDir)
	if err := os.MkdirAll(cloneDir, 0755); err != nil {
		return "", errors.Wrap(err, "creating new workspace")
	}

	log.Info("git cloning %q into %q", headRepo.SanitizedCloneURL, cloneDir)
	cloneCmd := exec.Command("git", "clone", headRepo.CloneURL, cloneDir)
	if output, err := cloneCmd.CombinedOutput(); err != nil {
		return "", errors.Wrapf(err, "cloning %s: %s", headRepo.SanitizedCloneURL, string(output))
	}

	// check out the branch for this PR
	log.Info("checking out branch %q", p.Branch)
	checkoutCmd := exec.Command("git", "checkout", p.Branch)
	checkoutCmd.Dir = cloneDir
	if err := checkoutCmd.Run(); err != nil {
		return "", errors.Wrapf(err, "checking out branch %s", p.Branch)
	}
	return cloneDir, nil
}

func (w *FileWorkspace) GetWorkspace(r models.Repo, p models.PullRequest, env string) (string, error) {
	repoDir := w.cloneDir(r, p, env)
	if _, err := os.Stat(repoDir); err != nil {
		return "", errors.Wrap(err, "checking if workspace exists")
	}
	return repoDir, nil
}

// Delete deletes the workspace for this repo and pull
func (w *FileWorkspace) Delete(r models.Repo, p models.PullRequest) error {
	return os.RemoveAll(w.repoPullDir(r, p))
}

func (w *FileWorkspace) repoPullDir(r models.Repo, p models.PullRequest) string {
	return filepath.Join(w.dataDir, workspacePrefix, r.FullName, strconv.Itoa(p.Num))
}

func (w *FileWorkspace) cloneDir(r models.Repo, p models.PullRequest, env string) string {
	return filepath.Join(w.repoPullDir(r, p), env)
}
