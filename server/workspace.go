package server

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/hootsuite/atlantis/models"
	"github.com/pkg/errors"
)

const workspacePrefix = "repos"

type Workspace struct {
	dataDir string
	sshKey  string
}

func (w *Workspace) Clone(ctx *CommandContext) (string, error) {
	cloneDir := w.cloneDir(ctx)

	// this is safe to do because we lock runs on repo/pull/env so no one else is using this workspace
	ctx.Log.Info("cleaning clone directory %q", cloneDir)
	if err := os.RemoveAll(cloneDir); err != nil {
		return "", errors.Wrap(err, "deleting old workspace")
	}

	// create the directory and parents if necessary
	ctx.Log.Info("creating dir %q", cloneDir)
	if err := os.MkdirAll(cloneDir, 0755); err != nil {
		return "", errors.Wrap(err, "creating new workspace")
	}

	// Check if ssh key is set and create git ssh wrapper
	cloneCmd := exec.Command("git", "clone", ctx.Repo.SSHURL, cloneDir)

	// clone the repo
	ctx.Log.Info("git cloning %q into %q", ctx.Repo.SSHURL, cloneDir)
	if output, err := cloneCmd.CombinedOutput(); err != nil {
		return "", errors.Wrapf(err, "cloning %s: %s", ctx.Repo.SSHURL, string(output))
	}

	// check out the branch for this PR
	ctx.Log.Info("checking out branch %q", ctx.Pull.Branch)
	checkoutCmd := exec.Command("git", "checkout", ctx.Pull.Branch)
	checkoutCmd.Dir = cloneDir
	if err := checkoutCmd.Run(); err != nil {
		return "", errors.Wrapf(err, "checking out branch %s", ctx.Pull.Branch)
	}
	return cloneDir, nil
}

func (w *Workspace) GetWorkspace(ctx *CommandContext) (string, error) {
	repoDir := w.cloneDir(ctx)
	if _, err := os.Stat(repoDir); err != nil {
		return "", errors.Wrap(err, "checking if workspace exists")
	}
	return repoDir, nil
}

// Delete deletes the workspace for this repo and pull
func (w *Workspace) Delete(repo models.Repo, pull models.PullRequest) error {
	return os.RemoveAll(w.repoPullDir(repo, pull))
}

func (w *Workspace) repoPullDir(repo models.Repo, pull models.PullRequest) string {
	return filepath.Join(w.dataDir, workspacePrefix, repo.FullName, strconv.Itoa(pull.Num))
}

func (w *Workspace) cloneDir(ctx *CommandContext) string {
	return filepath.Join(w.repoPullDir(ctx.Repo, ctx.Pull), ctx.Command.environment)
}
