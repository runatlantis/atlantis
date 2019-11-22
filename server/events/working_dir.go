// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

package events

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

const workingDirPrefix = "repos"

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_working_dir.go WorkingDir
//go:generate pegomock generate -m --use-experimental-model-gen --package events WorkingDir

// WorkingDir handles the workspace on disk for running commands.
type WorkingDir interface {
	// Clone git clones headRepo, checks out the branch and then returns the
	// absolute path to the root of the cloned repo.
	Clone(log *logging.SimpleLogger, baseRepo models.Repo, headRepo models.Repo, p models.PullRequest, workspace string) (string, error)
	// GetWorkingDir returns the path to the workspace for this repo and pull.
	// If workspace does not exist on disk, error will be of type os.IsNotExist.
	GetWorkingDir(r models.Repo, p models.PullRequest, workspace string) (string, error)
	GetPullDir(r models.Repo, p models.PullRequest) (string, error)
	// Delete deletes the workspace for this repo and pull.
	Delete(r models.Repo, p models.PullRequest) error
	DeleteForWorkspace(r models.Repo, p models.PullRequest, workspace string) error
}

// FileWorkspace implements WorkingDir with the file system.
type FileWorkspace struct {
	DataDir string
	// CheckoutMerge is true if we should check out the branch that corresponds
	// to what the base branch will look like *after* the pull request is merged.
	// If this is false, then we will check out the head branch from the pull
	// request.
	CheckoutMerge bool
	// TestingOverrideHeadCloneURL can be used during testing to override the
	// URL of the head repo to be cloned. If it's empty then we clone normally.
	TestingOverrideHeadCloneURL string
	// TestingOverrideBaseCloneURL can be used during testing to override the
	// URL of the base repo to be cloned. If it's empty then we clone normally.
	TestingOverrideBaseCloneURL string
}

// Clone git clones headRepo, checks out the branch and then returns the absolute
// path to the root of the cloned repo. If the repo already exists and is at
// the right commit it does nothing. This is to support running commands in
// multiple dirs of the same repo without deleting existing plans.
func (w *FileWorkspace) Clone(
	log *logging.SimpleLogger,
	baseRepo models.Repo,
	headRepo models.Repo,
	p models.PullRequest,
	workspace string) (string, error) {
	cloneDir := w.cloneDir(baseRepo, p, workspace)

	// If the directory already exists, check if it's at the right commit.
	// If so, then we do nothing.
	if _, err := os.Stat(cloneDir); err == nil {
		log.Debug("clone directory %q already exists, checking if it's at the right commit", cloneDir)

		// We use git rev-parse to see if our repo is at the right commit.
		// If just checking out the pull request branch, we can use HEAD.
		// If doing a merge, then HEAD won't be at the pull request's HEAD
		// because we'll already have performed a merge. Instead, we'll check
		// HEAD^2 since that will be the commit before our merge.
		pullHead := "HEAD"
		if w.CheckoutMerge {
			pullHead = "HEAD^2"
		}
		revParseCmd := exec.Command("git", "rev-parse", pullHead) // #nosec
		revParseCmd.Dir = cloneDir
		output, err := revParseCmd.CombinedOutput()
		if err != nil {
			log.Warn("will re-clone repo, could not determine if was at correct commit: %s: %s: %s", strings.Join(revParseCmd.Args, " "), err, string(output))
			return w.forceClone(log, cloneDir, headRepo, p)
		}
		currCommit := strings.Trim(string(output), "\n")
		// We're prefix matching here because BitBucket doesn't give us the full
		// commit, only a 12 character prefix.
		if strings.HasPrefix(currCommit, p.HeadCommit) {
			log.Debug("repo is at correct commit %q so will not re-clone", p.HeadCommit)
			return cloneDir, nil
		}
		log.Debug("repo was already cloned but is not at correct commit, wanted %q got %q", p.HeadCommit, currCommit)
		// We'll fall through to re-clone.
	}

	// Otherwise we clone the repo.
	return w.forceClone(log, cloneDir, headRepo, p)
}

func (w *FileWorkspace) forceClone(log *logging.SimpleLogger,
	cloneDir string,
	headRepo models.Repo,
	p models.PullRequest) (string, error) {

	err := os.RemoveAll(cloneDir)
	if err != nil {
		return "", errors.Wrapf(err, "deleting dir %q before cloning", cloneDir)
	}

	// Create the directory and parents if necessary.
	log.Info("creating dir %q", cloneDir)
	if err := os.MkdirAll(cloneDir, 0700); err != nil {
		return "", errors.Wrap(err, "creating new workspace")
	}

	// During testing, we mock some of this out.
	headCloneURL := headRepo.CloneURL
	if w.TestingOverrideHeadCloneURL != "" {
		headCloneURL = w.TestingOverrideHeadCloneURL
	}
	baseCloneURL := p.BaseRepo.CloneURL
	if w.TestingOverrideBaseCloneURL != "" {
		baseCloneURL = w.TestingOverrideBaseCloneURL
	}

	var cmds [][]string
	if w.CheckoutMerge {
		// NOTE: We can't do a shallow clone when we're merging because we'll
		// get merge conflicts if our clone doesn't have the commits that the
		// branch we're merging branched off at.
		// See https://groups.google.com/forum/#!topic/git-users/v3MkuuiDJ98.
		cmds = [][]string{
			{
				"git", "clone", "--branch", p.BaseBranch, "--single-branch", baseCloneURL, cloneDir,
			},
			{
				"git", "remote", "add", "head", headCloneURL,
			},
			{
				"git", "fetch", "head", fmt.Sprintf("+refs/heads/%s:", p.HeadBranch),
			},
			// We use --no-ff because we always want there to be a merge commit.
			// This way, our branch will look the same regardless if the merge
			// could be fast forwarded. This is useful later when we run
			// git rev-parse HEAD^2 to get the head commit because it will
			// always succeed whereas without --no-ff, if the merge was fast
			// forwarded then git rev-parse HEAD^2 would fail.
			{
				"git", "merge", "-q", "--no-ff", "-m", "atlantis-merge", "FETCH_HEAD",
			},
		}
	} else {
		cmds = [][]string{
			{
				"git", "clone", "--branch", p.HeadBranch, "--depth=1", "--single-branch", headCloneURL, cloneDir,
			},
		}
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...) // nolint: gosec
		cmd.Dir = cloneDir
		// The git merge command requires these env vars are set.
		cmd.Env = append(os.Environ(), []string{
			"EMAIL=atlantis@runatlants.io",
			"GIT_AUTHOR_NAME=atlantis",
			"GIT_COMMITTER_NAME=atlantis",
		}...)

		cmdStr := w.sanitizeGitCredentials(strings.Join(cmd.Args, " "), p.BaseRepo, headRepo)
		output, err := cmd.CombinedOutput()
		sanitizedOutput := w.sanitizeGitCredentials(string(output), p.BaseRepo, headRepo)
		if err != nil {
			sanitizedErrMsg := w.sanitizeGitCredentials(err.Error(), p.BaseRepo, headRepo)
			return "", fmt.Errorf("running %s: %s: %s", cmdStr, sanitizedOutput, sanitizedErrMsg)
		}
		log.Debug("ran: %s. Output: %s", cmdStr, strings.TrimSuffix(sanitizedOutput, "\n"))
	}
	return cloneDir, nil
}

// GetWorkingDir returns the path to the workspace for this repo and pull.
func (w *FileWorkspace) GetWorkingDir(r models.Repo, p models.PullRequest, workspace string) (string, error) {
	repoDir := w.cloneDir(r, p, workspace)
	if _, err := os.Stat(repoDir); err != nil {
		return "", errors.Wrap(err, "checking if workspace exists")
	}
	return repoDir, nil
}

// GetPullDir returns the dir where the workspaces for this pull are cloned.
// If the dir doesn't exist it will return an error.
func (w *FileWorkspace) GetPullDir(r models.Repo, p models.PullRequest) (string, error) {
	dir := w.repoPullDir(r, p)
	if _, err := os.Stat(dir); err != nil {
		return "", err
	}
	return dir, nil
}

// Delete deletes the workspace for this repo and pull.
func (w *FileWorkspace) Delete(r models.Repo, p models.PullRequest) error {
	return os.RemoveAll(w.repoPullDir(r, p))
}

// DeleteForWorkspace deletes the working dir for this workspace.
func (w *FileWorkspace) DeleteForWorkspace(r models.Repo, p models.PullRequest, workspace string) error {
	return os.RemoveAll(w.cloneDir(r, p, workspace))
}

func (w *FileWorkspace) repoPullDir(r models.Repo, p models.PullRequest) string {
	return filepath.Join(w.DataDir, workingDirPrefix, r.FullName, strconv.Itoa(p.Num))
}

func (w *FileWorkspace) cloneDir(r models.Repo, p models.PullRequest, workspace string) string {
	return filepath.Join(w.repoPullDir(r, p), workspace)
}

// sanitizeGitCredentials replaces any git clone urls that contain credentials
// in s with the sanitized versions.
func (w *FileWorkspace) sanitizeGitCredentials(s string, base models.Repo, head models.Repo) string {
	baseReplaced := strings.Replace(s, base.CloneURL, base.SanitizedCloneURL, -1)
	return strings.Replace(baseReplaced, head.CloneURL, head.SanitizedCloneURL, -1)
}
