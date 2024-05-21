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
	"sync"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/utils"
)

const workingDirPrefix = "repos"

var cloneLocks sync.Map

//go:generate pegomock generate github.com/runatlantis/atlantis/server/events --package mocks -o mocks/mock_working_dir.go WorkingDir
//go:generate pegomock generate github.com/runatlantis/atlantis/server/events --package events WorkingDir

// WorkingDir handles the workspace on disk for running commands.
type WorkingDir interface {
	// Clone git clones headRepo, checks out the branch and then returns the
	// absolute path to the root of the cloned repo. It also returns
	// a boolean indicating if we should warn users that the branch we're
	// merging into has been updated since we cloned it.
	Clone(logger logging.SimpleLogging, headRepo models.Repo, p models.PullRequest, workspace string) (string, bool, error)
	// GetWorkingDir returns the path to the workspace for this repo and pull.
	// If workspace does not exist on disk, error will be of type os.IsNotExist.
	GetWorkingDir(r models.Repo, p models.PullRequest, workspace string) (string, error)
	HasDiverged(logger logging.SimpleLogging, cloneDir string) bool
	GetPullDir(r models.Repo, p models.PullRequest) (string, error)
	// Delete deletes the workspace for this repo and pull.
	Delete(logger logging.SimpleLogging, r models.Repo, p models.PullRequest) error
	DeleteForWorkspace(logger logging.SimpleLogging, r models.Repo, p models.PullRequest, workspace string) error
	// Set a flag in the workingdir so Clone() can know that it is safe to re-clone the workingdir if
	// the upstream branch has been modified. This is only safe after grabbing the project lock
	// and before running any plans
	SetCheckForUpstreamChanges()
	// DeletePlan deletes the plan for this repo, pull, workspace path and project name
	DeletePlan(logger logging.SimpleLogging, r models.Repo, p models.PullRequest, workspace string, path string, projectName string) error
	// GetGitUntrackedFiles returns a list of Git untracked files in the working dir.
	GetGitUntrackedFiles(logger logging.SimpleLogging, r models.Repo, p models.PullRequest, workspace string) ([]string, error)
}

// FileWorkspace implements WorkingDir with the file system.
type FileWorkspace struct {
	DataDir string
	// CheckoutMerge is true if we should check out the branch that corresponds
	// to what the base branch will look like *after* the pull request is merged.
	// If this is false, then we will check out the head branch from the pull
	// request.
	CheckoutMerge bool
	// CheckoutDepth is how many commits of feature branch and main branch we'll
	// retrieve by default. If their merge base is not retrieved with this depth,
	// full fetch will be performed. Only matters if CheckoutMerge=true.
	CheckoutDepth int
	// TestingOverrideHeadCloneURL can be used during testing to override the
	// URL of the head repo to be cloned. If it's empty then we clone normally.
	TestingOverrideHeadCloneURL string
	// TestingOverrideBaseCloneURL can be used during testing to override the
	// URL of the base repo to be cloned. If it's empty then we clone normally.
	TestingOverrideBaseCloneURL string
	// GithubAppEnabled is true if we should fetch the ref "pull/PR_NUMBER/head"
	// from the "origin" remote. If this is false, we fetch "+refs/heads/$HEAD_BRANCH"
	// from the "head" remote.
	GithubAppEnabled bool
	// use the global setting without overriding
	GpgNoSigningEnabled bool
	// flag indicating if we have to merge with potential new changes upstream (directly after grabbing project lock)
	CheckForUpstreamChanges bool
}

// Clone git clones headRepo, checks out the branch and then returns the absolute
// path to the root of the cloned repo. It also returns
// a boolean indicating whether we had to merge with upstream again.
// If the repo already exists and is at
// the right commit it does nothing. This is to support running commands in
// multiple dirs of the same repo without deleting existing plans.
func (w *FileWorkspace) Clone(logger logging.SimpleLogging, headRepo models.Repo, p models.PullRequest, workspace string) (string, bool, error) {
	cloneDir := w.cloneDir(p.BaseRepo, p, workspace)
	defer func() { w.CheckForUpstreamChanges = false }()

	c := wrappedGitContext{cloneDir, headRepo, p}
	// If the directory already exists, check if it's at the right commit.
	// If so, then we do nothing.
	if _, err := os.Stat(cloneDir); err == nil {
		logger.Debug("clone directory '%s' already exists, checking if it's at the right commit", cloneDir)

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
		outputRevParseCmd, err := revParseCmd.CombinedOutput()
		if err != nil {
			logger.Warn("will re-clone repo, could not determine if was at correct commit: %s: %s: %s", strings.Join(revParseCmd.Args, " "), err, string(outputRevParseCmd))
			return cloneDir, false, w.forceClone(logger, c)
		}
		currCommit := strings.Trim(string(outputRevParseCmd), "\n")

		// We're prefix matching here because BitBucket doesn't give us the full
		// commit, only a 12 character prefix.
		if strings.HasPrefix(currCommit, p.HeadCommit) {
			if w.CheckForUpstreamChanges && w.CheckoutMerge && w.recheckDiverged(logger, p, headRepo, cloneDir) {
				logger.Info("base branch has been updated, using merge strategy and will clone again")
				return cloneDir, true, w.mergeAgain(logger, c)
			}
			logger.Debug("repo is at correct commit '%s' so will not re-clone", p.HeadCommit)
			return cloneDir, false, nil
		} else {
			logger.Debug("repo was already cloned but is not at correct commit, wanted '%s' got '%s'", p.HeadCommit, currCommit)
		}
		// We'll fall through to re-clone.
	}

	// Otherwise we clone the repo.
	return cloneDir, false, w.forceClone(logger, c)
}

// recheckDiverged returns true if the branch we're merging into has diverged
// from what we currently have checked out.
// This matters in the case of the merge checkout strategy because after
// cloning the repo and doing the merge, it's possible main was updated
// and we have to perform a new merge.
// If there are any errors we return false since we prefer things to succeed
// vs. stopping the plan/apply.
func (w *FileWorkspace) recheckDiverged(logger logging.SimpleLogging, p models.PullRequest, headRepo models.Repo, cloneDir string) bool {
	if !w.CheckoutMerge {
		// It only makes sense to warn that main has diverged if we're using
		// the checkout merge strategy. If we're just checking out the branch,
		// then it doesn't matter what's going on with main because we've
		// decided to always run off the branch.
		return false
	}

	// Bring our remote refs up to date.
	// Reset the URL in case we are using github app credentials since these might have
	// expired and refreshed and the URL would now be different.
	// In this case, we should be using a proxy URL which substitutes the credentials in
	// as a long term fix, but something like that requires more e2e testing/time
	cmds := [][]string{
		{
			"git", "remote", "set-url", "origin", p.BaseRepo.CloneURL,
		},
		{
			"git", "remote", "set-url", "head", headRepo.CloneURL,
		},
		{
			"git", "remote", "update",
		},
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...) // nolint: gosec
		cmd.Dir = cloneDir

		output, err := cmd.CombinedOutput()
		if err != nil {
			logger.Warn("getting remote update failed: %s", string(output))
			return false
		}
	}

	return w.HasDiverged(logger, cloneDir)
}

func (w *FileWorkspace) HasDiverged(logger logging.SimpleLogging, cloneDir string) bool {
	if !w.CheckoutMerge {
		// Both the diverged warning and the UnDiverged apply requirement only apply to merge checkout strategy so
		// we assume false here for 'branch' strategy.
		return false
	}

	statusFetchCmd := exec.Command("git", "fetch")
	statusFetchCmd.Dir = cloneDir
	outputStatusFetch, err := statusFetchCmd.CombinedOutput()
	if err != nil {
		logger.Warn("fetching repo has failed: %s", string(outputStatusFetch))
		return false
	}

	// Check if remote main branch has diverged.
	statusUnoCmd := exec.Command("git", "status", "--untracked-files=no")
	statusUnoCmd.Dir = cloneDir
	outputStatusUno, err := statusUnoCmd.CombinedOutput()
	if err != nil {
		logger.Warn("getting repo status has failed: %s", string(outputStatusUno))
		return false
	}
	hasDiverged := strings.Contains(string(outputStatusUno), "have diverged")
	return hasDiverged
}

func (w *FileWorkspace) forceClone(logger logging.SimpleLogging, c wrappedGitContext) error {
	value, _ := cloneLocks.LoadOrStore(c.dir, new(sync.Mutex))
	mutex := value.(*sync.Mutex)

	defer mutex.Unlock()
	if locked := mutex.TryLock(); !locked {
		mutex.Lock()
		return nil
	}

	err := os.RemoveAll(c.dir)
	if err != nil {
		return errors.Wrapf(err, "deleting dir '%s' before cloning", c.dir)
	}

	// Create the directory and parents if necessary.
	logger.Info("creating dir '%s'", c.dir)
	if err := os.MkdirAll(c.dir, 0700); err != nil {
		return errors.Wrap(err, "creating new workspace")
	}

	// During testing, we mock some of this out.
	headCloneURL := c.head.CloneURL
	if w.TestingOverrideHeadCloneURL != "" {
		headCloneURL = w.TestingOverrideHeadCloneURL
	}
	baseCloneURL := c.pr.BaseRepo.CloneURL
	if w.TestingOverrideBaseCloneURL != "" {
		baseCloneURL = w.TestingOverrideBaseCloneURL
	}

	// if branch strategy, use depth=1
	if !w.CheckoutMerge {
		return w.wrappedGit(logger, c, "clone", "--depth=1", "--branch", c.pr.HeadBranch, "--single-branch", headCloneURL, c.dir)
	}

	// if merge strategy...

	// if no checkout depth, omit depth arg
	if w.CheckoutDepth == 0 {
		if err := w.wrappedGit(logger, c, "clone", "--branch", c.pr.BaseBranch, "--single-branch", baseCloneURL, c.dir); err != nil {
			return err
		}
	} else {
		if err := w.wrappedGit(logger, c, "clone", "--depth", fmt.Sprint(w.CheckoutDepth), "--branch", c.pr.BaseBranch, "--single-branch", baseCloneURL, c.dir); err != nil {
			return err
		}
	}

	if err := w.wrappedGit(logger, c, "remote", "add", "head", headCloneURL); err != nil {
		return err
	}
	if w.GpgNoSigningEnabled {
		if err := w.wrappedGit(logger, c, "config", "--local", "commit.gpgsign", "false"); err != nil {
			return err
		}
	}

	return w.mergeToBaseBranch(logger, c)
}

// There is a new upstream update that we need, and we want to update to it
// without deleting any existing plans
func (w *FileWorkspace) mergeAgain(logger logging.SimpleLogging, c wrappedGitContext) error {
	value, _ := cloneLocks.LoadOrStore(c.dir, new(sync.Mutex))
	mutex := value.(*sync.Mutex)

	defer mutex.Unlock()
	if locked := mutex.TryLock(); !locked {
		mutex.Lock()
		return nil
	}

	// Reset branch as if it was cloned again
	if err := w.wrappedGit(logger, c, "reset", "--hard", fmt.Sprintf("refs/remotes/origin/%s", c.pr.BaseBranch)); err != nil {
		return err
	}

	return w.mergeToBaseBranch(logger, c)
}

// wrappedGitContext is the configuration for wrappedGit that is typically unchanged
// for a series of calls to wrappedGit
type wrappedGitContext struct {
	dir  string
	head models.Repo
	pr   models.PullRequest
}

// wrappedGit runs git with additional environment settings required for git merge,
// and with sanitized error logging to avoid leaking git credentials
func (w *FileWorkspace) wrappedGit(logger logging.SimpleLogging, c wrappedGitContext, args ...string) error {
	cmd := exec.Command("git", args...) // nolint: gosec
	cmd.Dir = c.dir
	// The git merge command requires these env vars are set.
	cmd.Env = append(os.Environ(), []string{
		"EMAIL=atlantis@runatlantis.io",
		"GIT_AUTHOR_NAME=atlantis",
		"GIT_COMMITTER_NAME=atlantis",
	}...)
	cmdStr := w.sanitizeGitCredentials(strings.Join(cmd.Args, " "), c.pr.BaseRepo, c.head)
	output, err := cmd.CombinedOutput()
	sanitizedOutput := w.sanitizeGitCredentials(string(output), c.pr.BaseRepo, c.head)
	if err != nil {
		sanitizedErrMsg := w.sanitizeGitCredentials(err.Error(), c.pr.BaseRepo, c.head)
		return fmt.Errorf("running %s: %s: %s", cmdStr, sanitizedOutput, sanitizedErrMsg)
	}
	logger.Debug("ran: %s. Output: %s", cmdStr, strings.TrimSuffix(sanitizedOutput, "\n"))
	return nil
}

// Merge the PR into the base branch.
func (w *FileWorkspace) mergeToBaseBranch(logger logging.SimpleLogging, c wrappedGitContext) error {
	fetchRef := fmt.Sprintf("+refs/heads/%s:", c.pr.HeadBranch)
	fetchRemote := "head"
	if w.GithubAppEnabled {
		fetchRef = fmt.Sprintf("pull/%d/head:", c.pr.Num)
		fetchRemote = "origin"
	}

	// if no checkout depth, omit depth arg
	if w.CheckoutDepth == 0 {
		if err := w.wrappedGit(logger, c, "fetch", fetchRemote, fetchRef); err != nil {
			return err
		}
	} else {
		if err := w.wrappedGit(logger, c, "fetch", "--depth", fmt.Sprint(w.CheckoutDepth), fetchRemote, fetchRef); err != nil {
			return err
		}
	}

	if err := w.wrappedGit(logger, c, "merge-base", c.pr.BaseBranch, "FETCH_HEAD"); err != nil {
		// git merge-base returning error means that we did not receive enough commits in shallow clone.
		// Fall back to retrieving full repo history.
		if err := w.wrappedGit(logger, c, "fetch", "--unshallow"); err != nil {
			return err
		}
	}

	// We use --no-ff because we always want there to be a merge commit.
	// This way, our branch will look the same regardless if the merge
	// could be fast forwarded. This is useful later when we run
	// git rev-parse HEAD^2 to get the head commit because it will
	// always succeed whereas without --no-ff, if the merge was fast
	// forwarded then git rev-parse HEAD^2 would fail.
	return w.wrappedGit(logger, c, "merge", "-q", "--no-ff", "-m", "atlantis-merge", "FETCH_HEAD")
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
func (w *FileWorkspace) Delete(logger logging.SimpleLogging, r models.Repo, p models.PullRequest) error {
	repoPullDir := w.repoPullDir(r, p)
	logger.Info("Deleting repo pull directory: " + repoPullDir)
	return os.RemoveAll(repoPullDir)
}

// DeleteForWorkspace deletes the working dir for this workspace.
func (w *FileWorkspace) DeleteForWorkspace(logger logging.SimpleLogging, r models.Repo, p models.PullRequest, workspace string) error {
	workspaceDir := w.cloneDir(r, p, workspace)
	logger.Info("Deleting workspace directory: " + workspaceDir)
	return os.RemoveAll(workspaceDir)
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

// Set the flag that indicates we need to check for upstream changes (if using merge checkout strategy)
func (w *FileWorkspace) SetCheckForUpstreamChanges() {
	w.CheckForUpstreamChanges = true
}

func (w *FileWorkspace) DeletePlan(logger logging.SimpleLogging, r models.Repo, p models.PullRequest, workspace string, projectPath string, projectName string) error {
	planPath := filepath.Join(w.cloneDir(r, p, workspace), projectPath, runtime.GetPlanFilename(workspace, projectName))
	logger.Info("Deleting plan: " + planPath)
	return utils.RemoveIgnoreNonExistent(planPath)
}

// getGitUntrackedFiles returns a list of Git untracked files in the working dir.
func (w *FileWorkspace) GetGitUntrackedFiles(logger logging.SimpleLogging, r models.Repo, p models.PullRequest, workspace string) ([]string, error) {
	workingDir, err := w.GetWorkingDir(r, p, workspace)
	if err != nil {
		return nil, err
	}

	logger.Debug("Checking for Git untracked files in directory: '%s'", workingDir)
	cmd := exec.Command("git", "ls-files", "--others", "--exclude-standard")
	cmd.Dir = workingDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	untrackedFiles := strings.Split(string(output), "\n")[:]
	logger.Debug("Untracked files: '%s'", strings.Join(untrackedFiles, ","))
	return untrackedFiles, nil
}
