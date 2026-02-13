// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

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

// Test that if we don't have any existing files, we check out the repo.
func TestClone_NoneExisting(t *testing.T) {
	// Initialize the git repo.
	repoDir := initRepo(t)
	expCommit := runCmd(t, repoDir, "git", "rev-parse", "HEAD")

	dataDir := t.TempDir()

	logger := logging.NewNoopLogger(t)

	wd := &events.FileWorkspace{
		DataDir:                     dataDir,
		CheckoutMerge:               false,
		TestingOverrideHeadCloneURL: fmt.Sprintf("file://%s", repoDir),
		GpgNoSigningEnabled:         true,
	}

	cloneDir, err := wd.Clone(logger, models.Repo{}, models.PullRequest{
		BaseRepo:   models.Repo{},
		HeadBranch: "branch",
	}, "default")
	Ok(t, err)

	// Use rev-parse to verify at correct commit.
	actCommit := runCmd(t, cloneDir, "git", "rev-parse", "HEAD")
	Equals(t, expCommit, actCommit)
}

// Test running on main branch with merge strategy
func TestClone_MainBranchWithMergeStrategy(t *testing.T) {
	// Initialize the git repo.
	repoDir := initRepo(t)
	expCommit := runCmd(t, repoDir, "git", "rev-parse", "main")

	dataDir := t.TempDir()

	logger := logging.NewNoopLogger(t)

	overrideURL := fmt.Sprintf("file://%s", repoDir)
	wd := &events.FileWorkspace{
		DataDir:                     dataDir,
		CheckoutMerge:               true,
		TestingOverrideHeadCloneURL: overrideURL,
		TestingOverrideBaseCloneURL: overrideURL,
		GpgNoSigningEnabled:         true,
	}

	_, err := wd.Clone(logger, models.Repo{}, models.PullRequest{
		Num:        0,
		BaseRepo:   models.Repo{},
		HeadBranch: "branch",
		BaseBranch: "main",
	}, "default")
	Ok(t, err)

	// Create a file that we can use to check if the repo was recloned.
	runCmd(t, dataDir, "touch", "repos/0/default/proof")

	// re-clone to make sure we don't try to merge main into itself
	cloneDir, err := wd.Clone(logger, models.Repo{}, models.PullRequest{
		BaseRepo:   models.Repo{},
		HeadBranch: "branch",
		BaseBranch: "main",
	}, "default")
	Ok(t, err)

	// Use rev-parse to verify at correct commit.
	actCommit := runCmd(t, cloneDir, "git", "rev-parse", "HEAD")
	Equals(t, expCommit, actCommit)

	// Check that our proof file is still there, proving that we didn't reclone.
	_, err = os.Stat(filepath.Join(cloneDir, "proof"))
	Ok(t, err)
}

// Test that if we don't have any existing files, we check out the repo
// successfully when we're using the merge method.
func TestClone_CheckoutMergeNoneExisting(t *testing.T) {
	// Initialize the git repo.
	repoDir := initRepo(t)

	// Add a commit to branch 'branch' that's not on main.
	runCmd(t, repoDir, "git", "checkout", "branch")
	runCmd(t, repoDir, "touch", "branch-file")
	runCmd(t, repoDir, "git", "add", "branch-file")
	runCmd(t, repoDir, "git", "commit", "-m", "branch-commit")
	branchCommit := runCmd(t, repoDir, "git", "rev-parse", "HEAD")

	// Now switch back to main and advance the main branch by another
	// commit.
	runCmd(t, repoDir, "git", "checkout", "main")
	runCmd(t, repoDir, "touch", "main-file")
	runCmd(t, repoDir, "git", "add", "main-file")
	runCmd(t, repoDir, "git", "commit", "-m", "main-commit")
	mainCommit := runCmd(t, repoDir, "git", "rev-parse", "HEAD")

	// Finally, perform a merge in another branch ourselves, just so we know
	// what the final state of the repo should be.
	runCmd(t, repoDir, "git", "checkout", "-b", "mergetest")
	runCmd(t, repoDir, "git", "merge", "-m", "atlantis-merge", "branch")
	expLsOutput := runCmd(t, repoDir, "ls")

	logger := logging.NewNoopLogger(t)

	dataDir := t.TempDir()

	overrideURL := fmt.Sprintf("file://%s", repoDir)
	wd := &events.FileWorkspace{
		DataDir:                     dataDir,
		CheckoutMerge:               true,
		CheckoutDepth:               50,
		TestingOverrideHeadCloneURL: overrideURL,
		TestingOverrideBaseCloneURL: overrideURL,
		GpgNoSigningEnabled:         true,
	}

	cloneDir, err := wd.Clone(logger, models.Repo{}, models.PullRequest{
		BaseRepo:   models.Repo{},
		HeadBranch: "branch",
		BaseBranch: "main",
	}, "default")
	Ok(t, err)

	// Check the commits.
	actBaseCommit := runCmd(t, cloneDir, "git", "rev-parse", "HEAD~1")
	actHeadCommit := runCmd(t, cloneDir, "git", "rev-parse", "HEAD^2")
	Equals(t, mainCommit, actBaseCommit)
	Equals(t, branchCommit, actHeadCommit)

	// Use ls to verify the repo looks good.
	actLsOutput := runCmd(t, cloneDir, "ls")
	Equals(t, expLsOutput, actLsOutput)
}

// Test that if we're using the merge method and the repo is already cloned at
// the right commit, then we don't reclone.
func TestClone_CheckoutMergeNoReclone(t *testing.T) {
	// Initialize the git repo.
	repoDir := initRepo(t)

	// Add a commit to branch 'branch' that's not on main.
	runCmd(t, repoDir, "git", "checkout", "branch")
	runCmd(t, repoDir, "touch", "branch-file")
	runCmd(t, repoDir, "git", "add", "branch-file")
	runCmd(t, repoDir, "git", "commit", "-m", "branch-commit")

	// Now switch back to main and advance the main branch by another commit.
	runCmd(t, repoDir, "git", "checkout", "main")
	runCmd(t, repoDir, "touch", "main-file")
	runCmd(t, repoDir, "git", "add", "main-file")
	runCmd(t, repoDir, "git", "commit", "-m", "main-commit")

	logger := logging.NewNoopLogger(t)

	// Run the clone for the first time.
	dataDir := t.TempDir()
	overrideURL := fmt.Sprintf("file://%s", repoDir)
	wd := &events.FileWorkspace{
		DataDir:                     dataDir,
		CheckoutMerge:               true,
		CheckoutDepth:               50,
		TestingOverrideHeadCloneURL: overrideURL,
		TestingOverrideBaseCloneURL: overrideURL,
		GpgNoSigningEnabled:         true,
	}

	_, err := wd.Clone(logger, models.Repo{}, models.PullRequest{
		BaseRepo:   models.Repo{},
		HeadBranch: "branch",
		BaseBranch: "main",
	}, "default")
	Ok(t, err)

	// Create a file that we can use to check if the repo was recloned.
	runCmd(t, dataDir, "touch", "repos/0/default/proof")

	// Now run the clone again.
	cloneDir, err := wd.Clone(logger, models.Repo{}, models.PullRequest{
		BaseRepo:   models.Repo{},
		HeadBranch: "branch",
		BaseBranch: "main",
	}, "default")
	Ok(t, err)

	// Check that our proof file is still there, proving that we didn't reclone.
	_, err = os.Stat(filepath.Join(cloneDir, "proof"))
	Ok(t, err)
}

// Same as TestClone_CheckoutMergeNoReclone however the branch that gets
// merged is a fast-forward merge. See #584.
func TestClone_CheckoutMergeNoRecloneFastForward(t *testing.T) {
	// Initialize the git repo.
	repoDir := initRepo(t)

	// Add a commit to branch 'branch' that's not on main.
	// This will result in a fast-forwardable merge.
	runCmd(t, repoDir, "git", "checkout", "branch")
	runCmd(t, repoDir, "touch", "branch-file")
	runCmd(t, repoDir, "git", "add", "branch-file")
	runCmd(t, repoDir, "git", "commit", "-m", "branch-commit")

	logger := logging.NewNoopLogger(t)

	// Run the clone for the first time.
	dataDir := t.TempDir()
	overrideURL := fmt.Sprintf("file://%s", repoDir)
	wd := &events.FileWorkspace{
		DataDir:                     dataDir,
		CheckoutMerge:               true,
		CheckoutDepth:               50,
		TestingOverrideHeadCloneURL: overrideURL,
		TestingOverrideBaseCloneURL: overrideURL,
		GpgNoSigningEnabled:         true,
	}

	_, err := wd.Clone(logger, models.Repo{}, models.PullRequest{
		BaseRepo:   models.Repo{},
		HeadBranch: "branch",
		BaseBranch: "main",
	}, "default")
	Ok(t, err)

	// Create a file that we can use to check if the repo was recloned.
	runCmd(t, dataDir, "touch", "repos/0/default/proof")

	// Now run the clone again.
	cloneDir, err := wd.Clone(logger, models.Repo{}, models.PullRequest{
		BaseRepo:   models.Repo{},
		HeadBranch: "branch",
		BaseBranch: "main",
	}, "default")
	Ok(t, err)

	// Check that our proof file is still there, proving that we didn't reclone.
	_, err = os.Stat(filepath.Join(cloneDir, "proof"))
	Ok(t, err)
}

// Test that if there's a conflict when merging we return a good error.
func TestClone_CheckoutMergeConflict(t *testing.T) {
	// Initialize the git repo.
	repoDir := initRepo(t)

	// Add a commit to branch 'branch' that's not on main.
	runCmd(t, repoDir, "git", "checkout", "branch")
	runCmd(t, repoDir, "sh", "-c", "echo hi >> file")
	runCmd(t, repoDir, "git", "add", "file")
	runCmd(t, repoDir, "git", "commit", "-m", "branch-commit")

	// Add a new commit to main that will cause a conflict if branch was
	// merged.
	runCmd(t, repoDir, "git", "checkout", "main")
	runCmd(t, repoDir, "sh", "-c", "echo conflict >> file")
	runCmd(t, repoDir, "git", "add", "file")
	runCmd(t, repoDir, "git", "commit", "-m", "commit")

	logger := logging.NewNoopLogger(t)

	// We're set up, now trigger the Atlantis clone.
	dataDir := t.TempDir()
	overrideURL := fmt.Sprintf("file://%s", repoDir)
	wd := &events.FileWorkspace{
		DataDir:                     dataDir,
		CheckoutMerge:               true,
		CheckoutDepth:               50,
		TestingOverrideHeadCloneURL: overrideURL,
		TestingOverrideBaseCloneURL: overrideURL,
		GpgNoSigningEnabled:         true,
	}

	_, err := wd.Clone(logger, models.Repo{}, models.PullRequest{
		BaseRepo:   models.Repo{},
		HeadBranch: "branch",
		BaseBranch: "main",
	}, "default")

	ErrContains(t, "running git merge -q --no-ff -m atlantis-merge FETCH_HEAD", err)
	ErrContains(t, "Auto-merging file", err)
	ErrContains(t, "CONFLICT (add/add)", err)
	ErrContains(t, "Merge conflict in file", err)
	ErrContains(t, "Automatic merge failed; fix conflicts and then commit the result.", err)
	ErrContains(t, "exit status 1", err)
}

func TestClone_CheckoutMergeShallow(t *testing.T) {
	// Initialize the git repo.
	repoDir := initRepo(t)

	runCmd(t, repoDir, "git", "commit", "--allow-empty", "-m", "should not be cloned")
	oldCommit := strings.TrimSpace(runCmd(t, repoDir, "git", "rev-parse", "HEAD"))

	runCmd(t, repoDir, "git", "commit", "--allow-empty", "-m", "merge-base")
	baseCommit := strings.TrimSpace(runCmd(t, repoDir, "git", "rev-parse", "HEAD"))

	runCmd(t, repoDir, "git", "branch", "-f", "branch", "HEAD")

	// Add a commit to branch 'branch' that's not on master.
	runCmd(t, repoDir, "git", "checkout", "branch")
	runCmd(t, repoDir, "touch", "branch-file")
	runCmd(t, repoDir, "git", "add", "branch-file")
	runCmd(t, repoDir, "git", "commit", "-m", "branch-commit")

	// Now switch back to master and advance the master branch by another
	// commit.
	runCmd(t, repoDir, "git", "checkout", "main")
	runCmd(t, repoDir, "touch", "main-file")
	runCmd(t, repoDir, "git", "add", "main-file")
	runCmd(t, repoDir, "git", "commit", "-m", "main-commit")

	overrideURL := fmt.Sprintf("file://%s", repoDir)

	// Test that we don't check out full repo if using CheckoutMerge strategy
	t.Run("Shallow", func(t *testing.T) {
		logger := logging.NewNoopLogger(t)

		dataDir := t.TempDir()

		wd := &events.FileWorkspace{
			DataDir:       dataDir,
			CheckoutMerge: true,
			// retrieve two commits in each branch:
			// master: master-commit, merge-base
			// branch: branch-commit, merge-base
			CheckoutDepth:               2,
			TestingOverrideHeadCloneURL: overrideURL,
			TestingOverrideBaseCloneURL: overrideURL,
			GpgNoSigningEnabled:         true,
		}

		cloneDir, err := wd.Clone(logger, models.Repo{}, models.PullRequest{
			BaseRepo:   models.Repo{},
			HeadBranch: "branch",
			BaseBranch: "main",
		}, "default")
		Ok(t, err)

		gotBaseCommitType := runCmd(t, cloneDir, "git", "cat-file", "-t", baseCommit)
		Assert(t, gotBaseCommitType == "commit\n", "should have merge-base in shallow repo")
		gotOldCommitType := runCmdErrCode(t, cloneDir, 128, "git", "cat-file", "-t", oldCommit)
		Assert(t, strings.Contains(gotOldCommitType, "could not get object info"), "should not have old commit in shallow repo")
	})

	// Test that we will check out full repo if CheckoutDepth is too small
	t.Run("FullClone", func(t *testing.T) {
		logger := logging.NewNoopLogger(t)

		dataDir := t.TempDir()

		wd := &events.FileWorkspace{
			DataDir:       dataDir,
			CheckoutMerge: true,
			// 1 is not enough to retrieve merge-base, so full clone should be performed
			CheckoutDepth:               1,
			TestingOverrideHeadCloneURL: overrideURL,
			TestingOverrideBaseCloneURL: overrideURL,
			GpgNoSigningEnabled:         true,
		}

		cloneDir, err := wd.Clone(logger, models.Repo{}, models.PullRequest{
			BaseRepo:   models.Repo{},
			HeadBranch: "branch",
			BaseBranch: "main",
		}, "default")
		Ok(t, err)

		gotBaseCommitType := runCmd(t, cloneDir, "git", "cat-file", "-t", baseCommit)
		Assert(t, gotBaseCommitType == "commit\n", "should have merge-base in full repo")
		gotOldCommitType := runCmd(t, cloneDir, "git", "cat-file", "-t", oldCommit)
		Assert(t, gotOldCommitType == "commit\n", "should have old commit in full repo")
	})

}

// Test that if the repo is already cloned and is at the right commit, we
// don't reclone.
func TestClone_NoReclone(t *testing.T) {
	repoDir := initRepo(t)
	dataDir := t.TempDir()

	runCmd(t, dataDir, "mkdir", "-p", "repos/0/")
	runCmd(t, dataDir, "mv", repoDir, "repos/0/default")
	// Create a file that we can use later to check if the repo was recloned.
	runCmd(t, dataDir, "touch", "repos/0/default/proof")

	logger := logging.NewNoopLogger(t)

	wd := &events.FileWorkspace{
		DataDir:                     dataDir,
		CheckoutMerge:               false,
		TestingOverrideHeadCloneURL: fmt.Sprintf("file://%s", repoDir),
		GpgNoSigningEnabled:         true,
	}
	cloneDir, err := wd.Clone(logger, models.Repo{}, models.PullRequest{
		BaseRepo:   models.Repo{},
		HeadBranch: "branch",
	}, "default")
	Ok(t, err)

	// Check that our proof file is still there.
	_, err = os.Stat(filepath.Join(cloneDir, "proof"))
	Ok(t, err)
}

// Test that if the repo is already cloned but is at the wrong commit, we
// fetch and reset
func TestClone_ResetOnWrongCommit(t *testing.T) {
	repoDir := initRepo(t)
	dataDir := t.TempDir()

	// Copy the repo to our data dir.
	runCmd(t, dataDir, "mkdir", "-p", "repos/0/")
	runCmd(t, dataDir, "git", "clone", repoDir, "repos/0/default")

	// Now add a commit to the repo, so the one in the data dir is out of date.
	runCmd(t, repoDir, "git", "checkout", "branch")
	runCmd(t, repoDir, "touch", "newfile")
	runCmd(t, repoDir, "git", "add", "newfile")
	runCmd(t, repoDir, "git", "commit", "-m", "newfile")
	expCommit := strings.TrimSpace(runCmd(t, repoDir, "git", "rev-parse", "HEAD"))

	// Pretend that terraform has created a plan file, we'll check for it later
	planFile := filepath.Join(dataDir, "repos/0/default/default.tfplan")
	assert.NoFileExists(t, planFile)
	_, err := os.Create(planFile)
	Assert(t, err == nil, "creating plan file: %v", err)
	assert.FileExists(t, planFile)

	logger := logging.NewNoopLogger(t)

	wd := &events.FileWorkspace{
		DataDir:                     dataDir,
		CheckoutMerge:               false,
		TestingOverrideHeadCloneURL: fmt.Sprintf("file://%s", repoDir),
		GpgNoSigningEnabled:         true,
	}
	cloneDir, err := wd.Clone(logger, models.Repo{}, models.PullRequest{
		BaseRepo:   models.Repo{},
		HeadBranch: "branch",
		HeadCommit: expCommit,
		BaseBranch: "main",
	}, "default")
	Ok(t, err)
	assert.FileExists(t, planFile, "Plan file should not been wiped out by reset")

	// Use rev-parse to verify at correct commit.
	actCommit := strings.TrimSpace(runCmd(t, cloneDir, "git", "rev-parse", "HEAD"))
	Equals(t, expCommit, actCommit)
}

// Test that if the repo is already cloned but is at the wrong commit, but the base has changed
// we need to reclone
func TestClone_ReCloneOnBaseChange(t *testing.T) {
	repoDir := initRepo(t)
	dataDir := t.TempDir()

	// Copy the repo to our data dir.
	runCmd(t, dataDir, "mkdir", "-p", "repos/0/")
	runCmd(t, dataDir, "git", "clone", repoDir, "repos/0/default")

	// Now add a commit to the repo, so the one in the data dir is out of date.
	runCmd(t, repoDir, "git", "checkout", "branch")
	runCmd(t, repoDir, "touch", "newfile")
	runCmd(t, repoDir, "git", "add", "newfile")
	runCmd(t, repoDir, "git", "commit", "-m", "newfile")
	expCommit := strings.TrimSpace(runCmd(t, repoDir, "git", "rev-parse", "HEAD"))

	// Pretend that terraform has created a plan file, we'll check for it later
	planFile := filepath.Join(dataDir, "repos/0/default/default.tfplan")
	assert.NoFileExists(t, planFile)
	_, err := os.Create(planFile)
	Assert(t, err == nil, "creating plan file: %v", err)
	assert.FileExists(t, planFile)

	logger := logging.NewNoopLogger(t)

	wd := &events.FileWorkspace{
		DataDir:                     dataDir,
		CheckoutMerge:               false,
		TestingOverrideHeadCloneURL: fmt.Sprintf("file://%s", repoDir),
		GpgNoSigningEnabled:         true,
	}
	cloneDir, err := wd.Clone(logger, models.Repo{}, models.PullRequest{
		BaseRepo:   models.Repo{},
		HeadBranch: "branch",
		HeadCommit: expCommit,
		BaseBranch: "some-other-base-branch",
	}, "default")
	Ok(t, err)
	assert.NoFileExists(t, planFile, "Plan file should been wiped out by the reclone")

	// Use rev-parse to verify at correct commit.
	actCommit := strings.TrimSpace(runCmd(t, cloneDir, "git", "rev-parse", "HEAD"))
	Equals(t, expCommit, actCommit)
}

// Test that if the repo is already cloned but is at the wrong commit, but the base has changed
// we need to reclone
func TestClone_ReCloneOnErrorAttemptingReuse(t *testing.T) {
	repoDir := initRepo(t)
	dataDir := t.TempDir()

	// Copy the repo to our data dir.
	runCmd(t, dataDir, "mkdir", "-p", "repos/0/")
	runCmd(t, dataDir, "git", "clone", repoDir, "repos/0/default")

	// Now add a commit to the repo, so the one in the data dir is out of date.
	runCmd(t, repoDir, "git", "checkout", "branch")
	runCmd(t, repoDir, "touch", "newfile")
	runCmd(t, repoDir, "git", "add", "newfile")
	runCmd(t, repoDir, "git", "commit", "-m", "newfile")
	expCommit := strings.TrimSpace(runCmd(t, repoDir, "git", "rev-parse", "HEAD"))

	// Now intentionally break the remote so it's unable to fetch
	runCmd(t, dataDir, "rm", "repos/0/default/.git/HEAD")

	// Pretend that terraform has created a plan file, we'll check for it later
	planFile := filepath.Join(dataDir, "repos/0/default/default.tfplan")
	assert.NoFileExists(t, planFile)
	_, err := os.Create(planFile)
	Assert(t, err == nil, "creating plan file: %v", err)
	assert.FileExists(t, planFile)

	logger := logging.NewNoopLogger(t)

	wd := &events.FileWorkspace{
		DataDir:                     dataDir,
		CheckoutMerge:               false,
		TestingOverrideHeadCloneURL: fmt.Sprintf("file://%s", repoDir),
		GpgNoSigningEnabled:         true,
	}
	cloneDir, err := wd.Clone(logger, models.Repo{}, models.PullRequest{
		BaseRepo:   models.Repo{},
		HeadBranch: "branch",
		HeadCommit: expCommit,
		BaseBranch: "main",
	}, "default")
	Ok(t, err)
	assert.NoFileExists(t, planFile, "Plan file should been wiped out by the reclone")

	// Use rev-parse to verify at correct commit.
	actCommit := strings.TrimSpace(runCmd(t, cloneDir, "git", "rev-parse", "HEAD"))
	Equals(t, expCommit, actCommit)
}

func TestClone_ResetOnWrongCommitWithMergeStrategy(t *testing.T) {
	repoDir := initRepo(t)
	dataDir := t.TempDir()

	// In the repo, make sure there's a commit on the branch and main so we have something to merge
	runCmd(t, repoDir, "touch", "newfile")
	runCmd(t, repoDir, "git", "add", "newfile")
	runCmd(t, repoDir, "git", "commit", "-m", "newfile")

	runCmd(t, repoDir, "git", "checkout", "branch")
	runCmd(t, repoDir, "touch", "branchfile")
	runCmd(t, repoDir, "git", "add", "branchfile")
	runCmd(t, repoDir, "git", "commit", "-m", "branchfile")

	// Copy the repo to our data dir.
	checkoutDir := filepath.Join(dataDir, "repos/1/default")

	// "clone" the repoDir, instead of just copying it, because we're going to track it
	runCmd(t, dataDir, "mkdir", "-p", "repos/1/")
	runCmd(t, dataDir, "git", "clone", repoDir, checkoutDir)
	runCmd(t, checkoutDir, "git", "config", "--local", "user.email", "atlantisbot@runatlantis.io")
	runCmd(t, checkoutDir, "git", "config", "--local", "user.name", "atlantisbot")
	runCmd(t, checkoutDir, "git", "config", "--local", "commit.gpgsign", "false")

	runCmd(t, checkoutDir, "git", "checkout", "main")
	runCmd(t, checkoutDir, "git", "remote", "add", "source", repoDir)

	// Simulate the merge strategy
	runCmd(t, checkoutDir, "git", "checkout", "branch")
	runCmd(t, checkoutDir, "git", "config", "--local", "user.email", "atlantisbot@runatlantis.io")
	runCmd(t, checkoutDir, "git", "config", "--local", "user.name", "atlantisbot")
	runCmd(t, checkoutDir, "git", "merge", "-q", "--no-ff", "-m", "atlantis-merge", "origin/main")

	// Now add a commit to the repo, so the one in the data dir is out of date.
	runCmd(t, repoDir, "touch", "anotherfile")
	runCmd(t, repoDir, "git", "add", "anotherfile")
	runCmd(t, repoDir, "git", "commit", "-m", "anotherfile")
	expCommit := strings.TrimSpace(runCmd(t, repoDir, "git", "rev-parse", "HEAD"))

	// Pretend that terraform has created a plan file, we'll check for it later
	planFile := filepath.Join(dataDir, "repos/1/default/default.tfplan")
	assert.NoFileExists(t, planFile)
	_, err := os.Create(planFile)
	Assert(t, err == nil, "creating plan file: %v", err)
	assert.FileExists(t, planFile)

	logger := logging.NewNoopLogger(t)

	wd := &events.FileWorkspace{
		DataDir:                     dataDir,
		CheckoutMerge:               true,
		TestingOverrideHeadCloneURL: fmt.Sprintf("file://%s", repoDir),
		GpgNoSigningEnabled:         true,
	}
	fmt.Println(repoDir)
	cloneDir, err := wd.Clone(logger, models.Repo{}, models.PullRequest{
		BaseRepo:   models.Repo{},
		HeadBranch: "branch",
		HeadCommit: expCommit,
		BaseBranch: "main",
		Num:        1,
	}, "default")
	Ok(t, err)
	assert.FileExists(t, planFile, "Plan file should not been wiped out by reset")

	// Use rev-parse to verify at correct commit.
	actCommit := strings.TrimSpace(runCmd(t, cloneDir, "git", "rev-parse", "HEAD^2"))
	Equals(t, expCommit, actCommit)
}

// Test that if the branch we're merging into has diverged and we're using
// checkout-strategy=merge, we actually merge the branch.
// Also check that we do not merge if we are not using the merge strategy.
func TestClone_MasterHasDiverged(t *testing.T) {
	// Initialize the git repo.
	repoDir := initRepo(t)

	// Simulate first PR.
	runCmd(t, repoDir, "git", "checkout", "-b", "first-pr")
	runCmd(t, repoDir, "touch", "file1")
	runCmd(t, repoDir, "git", "add", "file1")
	runCmd(t, repoDir, "git", "commit", "-m", "file1")

	// Atlantis checkout first PR.
	firstPRDir := repoDir + "/first-pr"
	runCmd(t, repoDir, "mkdir", "-p", "first-pr")
	runCmd(t, firstPRDir, "git", "clone", "--branch", "main", "--single-branch", repoDir, ".")
	runCmd(t, firstPRDir, "git", "remote", "add", "source", repoDir)
	runCmd(t, firstPRDir, "git", "fetch", "source", "+refs/heads/first-pr")
	runCmd(t, firstPRDir, "git", "config", "--local", "user.email", "atlantisbot@runatlantis.io")
	runCmd(t, firstPRDir, "git", "config", "--local", "user.name", "atlantisbot")
	runCmd(t, firstPRDir, "git", "config", "--local", "commit.gpgsign", "false")
	runCmd(t, firstPRDir, "git", "merge", "-q", "--no-ff", "-m", "atlantis-merge", "FETCH_HEAD")

	// Simulate second PR.
	runCmd(t, repoDir, "git", "checkout", "main")
	runCmd(t, repoDir, "git", "checkout", "-b", "second-pr")
	runCmd(t, repoDir, "touch", "file2")
	runCmd(t, repoDir, "git", "add", "file2")
	runCmd(t, repoDir, "git", "commit", "-m", "file2")

	// Atlantis checkout second PR.
	secondPRDir := repoDir + "/second-pr"
	runCmd(t, repoDir, "mkdir", "-p", "second-pr")
	runCmd(t, secondPRDir, "git", "clone", "--branch", "main", "--single-branch", repoDir, ".")
	runCmd(t, secondPRDir, "git", "remote", "add", "source", repoDir)
	runCmd(t, secondPRDir, "git", "fetch", "source", "+refs/heads/second-pr")
	runCmd(t, secondPRDir, "git", "config", "--local", "user.email", "atlantisbot@runatlantis.io")
	runCmd(t, secondPRDir, "git", "config", "--local", "user.name", "atlantisbot")
	runCmd(t, secondPRDir, "git", "config", "--local", "commit.gpgsign", "false")
	runCmd(t, secondPRDir, "git", "merge", "-q", "--no-ff", "-m", "atlantis-merge", "FETCH_HEAD")

	// Merge first PR
	runCmd(t, repoDir, "git", "checkout", "main")
	runCmd(t, repoDir, "git", "merge", "first-pr")

	// Copy the second-pr repo to our data dir which has diverged remote main
	runCmd(t, repoDir, "mkdir", "-p", "repos/0/")
	runCmd(t, repoDir, "cp", "-R", secondPRDir, "repos/0/default")

	logger := logging.NewNoopLogger(t)

	// Run the clone.
	wd := &events.FileWorkspace{
		DataDir:             repoDir,
		CheckoutMerge:       false,
		CheckoutDepth:       50,
		GpgNoSigningEnabled: true,
	}

	// Pretend terraform has created a plan file, we'll check for it later
	planFile := filepath.Join(repoDir, "repos/0/default/default.tfplan")
	assert.NoFileExists(t, planFile)
	_, err := os.Create(planFile)
	Assert(t, err == nil, "creating plan file: %v", err)
	assert.FileExists(t, planFile)

	// Run MergeAgain without the checkout merge strategy. It should return
	// false for mergedAgain
	_, err = wd.Clone(logger, models.Repo{}, models.PullRequest{
		BaseRepo:   models.Repo{},
		HeadBranch: "second-pr",
		BaseBranch: "main",
	}, "default")
	Ok(t, err)
	mergedAgain, err := wd.MergeAgain(logger, models.Repo{CloneURL: repoDir}, models.PullRequest{
		BaseRepo:   models.Repo{CloneURL: repoDir},
		HeadBranch: "second-pr",
		BaseBranch: "main",
	}, "default")
	Ok(t, err)
	assert.FileExists(t, planFile, "Existing plan file should not be deleted by merging again")
	Assert(t, mergedAgain == false, "MergeAgain with CheckoutMerge=false should not merge")

	wd.CheckoutMerge = true
	// Run the clone twice with the merge strategy, the first run should
	// return true for mergedAgain, subsequent runs should
	// return false since the first call is supposed to merge.
	mergedAgain, err = wd.MergeAgain(logger, models.Repo{CloneURL: repoDir}, models.PullRequest{
		BaseRepo:   models.Repo{CloneURL: repoDir},
		HeadBranch: "second-pr",
		BaseBranch: "main",
	}, "default")
	Ok(t, err)
	assert.FileExists(t, planFile, "Existing plan file should not be deleted by merging again")
	Assert(t, mergedAgain == true, "First clone with CheckoutMerge=true with diverged base should have merged")

	mergedAgain, err = wd.MergeAgain(logger, models.Repo{CloneURL: repoDir}, models.PullRequest{
		BaseRepo:   models.Repo{CloneURL: repoDir},
		HeadBranch: "second-pr",
		BaseBranch: "main",
	}, "default")
	Ok(t, err)
	Assert(t, mergedAgain == false, "Second clone with CheckoutMerge=true and initially diverged base should not merge again")
	assert.FileExists(t, planFile, "Existing plan file should not have been deleted")
}

func TestHasDiverged_MasterHasDiverged(t *testing.T) {
	// Initialize the git repo.
	repoDir := initRepo(t)

	// Simulate first PR.
	runCmd(t, repoDir, "git", "checkout", "-b", "first-pr")
	runCmd(t, repoDir, "touch", "file1")
	runCmd(t, repoDir, "git", "add", "file1")
	runCmd(t, repoDir, "git", "commit", "-m", "file1")

	// Atlantis checkout first PR.
	firstPRDir := repoDir + "/first-pr"
	runCmd(t, repoDir, "mkdir", "-p", "first-pr")
	runCmd(t, firstPRDir, "git", "clone", "--branch", "main", "--single-branch", repoDir, ".")
	runCmd(t, firstPRDir, "git", "remote", "add", "source", repoDir)
	runCmd(t, firstPRDir, "git", "fetch", "source", "+refs/heads/first-pr")
	runCmd(t, firstPRDir, "git", "config", "--local", "user.email", "atlantisbot@runatlantis.io")
	runCmd(t, firstPRDir, "git", "config", "--local", "user.name", "atlantisbot")
	runCmd(t, firstPRDir, "git", "config", "--local", "commit.gpgsign", "false")
	runCmd(t, firstPRDir, "git", "merge", "-q", "--no-ff", "-m", "atlantis-merge", "FETCH_HEAD")

	// Simulate second PR.
	runCmd(t, repoDir, "git", "checkout", "main")
	runCmd(t, repoDir, "git", "checkout", "-b", "second-pr")
	runCmd(t, repoDir, "touch", "file2")
	runCmd(t, repoDir, "git", "add", "file2")
	runCmd(t, repoDir, "git", "commit", "-m", "file2")

	// Atlantis checkout second PR.
	secondPRDir := repoDir + "/second-pr"
	runCmd(t, repoDir, "mkdir", "-p", "second-pr")
	runCmd(t, secondPRDir, "git", "clone", "--branch", "main", "--single-branch", repoDir, ".")
	runCmd(t, secondPRDir, "git", "remote", "add", "source", repoDir)
	runCmd(t, secondPRDir, "git", "fetch", "source", "+refs/heads/second-pr")
	runCmd(t, secondPRDir, "git", "config", "--local", "user.email", "atlantisbot@runatlantis.io")
	runCmd(t, secondPRDir, "git", "config", "--local", "user.name", "atlantisbot")
	runCmd(t, secondPRDir, "git", "config", "--local", "commit.gpgsign", "false")
	runCmd(t, secondPRDir, "git", "merge", "-q", "--no-ff", "-m", "atlantis-merge", "FETCH_HEAD")

	// Merge first PR
	runCmd(t, repoDir, "git", "checkout", "main")
	runCmd(t, repoDir, "git", "merge", "first-pr")

	// Copy the second-pr repo to our data dir which has diverged remote main
	runCmd(t, repoDir, "mkdir", "-p", "repos/0/")
	runCmd(t, repoDir, "cp", "-R", secondPRDir, "repos/0/default")

	// "git", "remote", "set-url", "origin", p.BaseRepo.CloneURL,
	runCmd(t, repoDir+"/repos/0/default", "git", "remote", "update")

	logger := logging.NewNoopLogger(t)

	// Run the clone.
	wd := &events.FileWorkspace{
		DataDir:             repoDir,
		CheckoutMerge:       true,
		CheckoutDepth:       50,
		GpgNoSigningEnabled: true,
	}
	hasDiverged := wd.HasDiverged(logger, repoDir+"/repos/0/default")
	Equals(t, hasDiverged, true)

	// Run it again but without the checkout merge strategy. It should return
	// false.
	wd.CheckoutMerge = false
	hasDiverged = wd.HasDiverged(logger, repoDir+"/repos/0/default")
	Equals(t, hasDiverged, false)
}

func TestHasDivergedWhenModified_CheckoutMergeDisabled(t *testing.T) {
	// Initialize the git repo.
	repoDir := initRepo(t)

	// Simulate first PR with file changes.
	runCmd(t, repoDir, "git", "checkout", "-b", "first-pr")
	runCmd(t, repoDir, "mkdir", "-p", "project1")
	runCmd(t, repoDir, "touch", "project1/main.tf")
	runCmd(t, repoDir, "git", "add", "project1/main.tf")
	runCmd(t, repoDir, "git", "commit", "-m", "add project1")
	headCommit := strings.TrimSpace(runCmd(t, repoDir, "git", "rev-parse", "HEAD"))

	// Atlantis checkout first PR.
	firstPRDir := repoDir + "/first-pr"
	runCmd(t, repoDir, "mkdir", "-p", "first-pr")
	runCmd(t, firstPRDir, "git", "clone", "--branch", "main", "--single-branch", repoDir, ".")
	runCmd(t, firstPRDir, "git", "remote", "add", "source", repoDir)
	runCmd(t, firstPRDir, "git", "fetch", "source", "+refs/heads/first-pr")
	runCmd(t, firstPRDir, "git", "config", "--local", "user.email", "atlantisbot@runatlantis.io")
	runCmd(t, firstPRDir, "git", "config", "--local", "user.name", "atlantisbot")
	runCmd(t, firstPRDir, "git", "config", "--local", "commit.gpgsign", "false")
	runCmd(t, firstPRDir, "git", "merge", "-q", "--no-ff", "-m", "atlantis-merge", "FETCH_HEAD")

	logger := logging.NewNoopLogger(t)

	// Create the workspace without CheckoutMerge.
	wd := &events.FileWorkspace{
		DataDir:             repoDir,
		CheckoutMerge:       false,
		CheckoutDepth:       50,
		GpgNoSigningEnabled: true,
	}

	pullRequest := models.PullRequest{
		BaseRepo:   models.Repo{CloneURL: repoDir},
		HeadBranch: "first-pr",
		BaseBranch: "main",
		HeadCommit: headCommit,
	}

	// Should return false when CheckoutMerge is disabled.
	hasDiverged := wd.HasDivergedWhenModified(logger, firstPRDir, []string{"project1/**"}, pullRequest)
	Equals(t, false, hasDiverged)
}

func TestHasDivergedWhenModified_EmptyPatterns(t *testing.T) {
	// Initialize the git repo.
	repoDir := initRepo(t)

	// Simulate first PR.
	runCmd(t, repoDir, "git", "checkout", "-b", "first-pr")
	runCmd(t, repoDir, "touch", "file1")
	runCmd(t, repoDir, "git", "add", "file1")
	runCmd(t, repoDir, "git", "commit", "-m", "file1")

	// Atlantis checkout first PR.
	firstPRDir := repoDir + "/first-pr"
	runCmd(t, repoDir, "mkdir", "-p", "first-pr")
	runCmd(t, firstPRDir, "git", "clone", "--branch", "main", "--single-branch", repoDir, ".")
	runCmd(t, firstPRDir, "git", "remote", "add", "source", repoDir)
	runCmd(t, firstPRDir, "git", "fetch", "source", "+refs/heads/first-pr")
	runCmd(t, firstPRDir, "git", "config", "--local", "user.email", "atlantisbot@runatlantis.io")
	runCmd(t, firstPRDir, "git", "config", "--local", "user.name", "atlantisbot")
	runCmd(t, firstPRDir, "git", "config", "--local", "commit.gpgsign", "false")
	runCmd(t, firstPRDir, "git", "merge", "-q", "--no-ff", "-m", "atlantis-merge", "FETCH_HEAD")

	// Simulate second PR.
	runCmd(t, repoDir, "git", "checkout", "main")
	runCmd(t, repoDir, "git", "checkout", "-b", "second-pr")
	runCmd(t, repoDir, "touch", "file2")
	runCmd(t, repoDir, "git", "add", "file2")
	runCmd(t, repoDir, "git", "commit", "-m", "file2")
	secondHeadCommit := strings.TrimSpace(runCmd(t, repoDir, "git", "rev-parse", "HEAD"))

	// Atlantis checkout second PR.
	secondPRDir := repoDir + "/second-pr"
	runCmd(t, repoDir, "mkdir", "-p", "second-pr")
	runCmd(t, secondPRDir, "git", "clone", "--branch", "main", "--single-branch", repoDir, ".")
	runCmd(t, secondPRDir, "git", "remote", "add", "source", repoDir)
	runCmd(t, secondPRDir, "git", "fetch", "source", "+refs/heads/second-pr")
	runCmd(t, secondPRDir, "git", "config", "--local", "user.email", "atlantisbot@runatlantis.io")
	runCmd(t, secondPRDir, "git", "config", "--local", "user.name", "atlantisbot")
	runCmd(t, secondPRDir, "git", "config", "--local", "commit.gpgsign", "false")
	runCmd(t, secondPRDir, "git", "merge", "-q", "--no-ff", "-m", "atlantis-merge", "FETCH_HEAD")

	// Merge first PR to make main diverge from second PR.
	runCmd(t, repoDir, "git", "checkout", "main")
	runCmd(t, repoDir, "git", "merge", "first-pr")

	// Copy the second-pr repo to our data dir which has diverged remote main.
	runCmd(t, repoDir, "mkdir", "-p", "repos/0/")
	runCmd(t, repoDir, "cp", "-R", secondPRDir, "repos/0/default")

	runCmd(t, repoDir+"/repos/0/default", "git", "remote", "update")

	logger := logging.NewNoopLogger(t)

	wd := &events.FileWorkspace{
		DataDir:             repoDir,
		CheckoutMerge:       true,
		CheckoutDepth:       50,
		GpgNoSigningEnabled: true,
	}

	pullRequest := models.PullRequest{
		BaseRepo:   models.Repo{CloneURL: repoDir},
		HeadBranch: "second-pr",
		BaseBranch: "main",
		HeadCommit: secondHeadCommit,
	}

	// With empty patterns, should fall back to regular HasDiverged which should return true.
	hasDiverged := wd.HasDivergedWhenModified(logger, repoDir+"/repos/0/default", []string{}, pullRequest)
	Equals(t, true, hasDiverged)
}

func TestHasDivergedWhenModified_PatternHasDiverged(t *testing.T) {
	// Initialize the git repo.
	repoDir := initRepo(t)

	// Simulate first PR with changes to project1.
	runCmd(t, repoDir, "git", "checkout", "-b", "first-pr")
	runCmd(t, repoDir, "mkdir", "-p", "project1")
	runCmd(t, repoDir, "touch", "project1/main.tf")
	runCmd(t, repoDir, "git", "add", "project1/main.tf")
	runCmd(t, repoDir, "git", "commit", "-m", "add project1")

	// Atlantis checkout first PR.
	firstPRDir := repoDir + "/first-pr"
	runCmd(t, repoDir, "mkdir", "-p", "first-pr")
	runCmd(t, firstPRDir, "git", "clone", "--branch", "main", "--single-branch", repoDir, ".")
	runCmd(t, firstPRDir, "git", "remote", "add", "source", repoDir)
	runCmd(t, firstPRDir, "git", "fetch", "source", "+refs/heads/first-pr")
	runCmd(t, firstPRDir, "git", "config", "--local", "user.email", "atlantisbot@runatlantis.io")
	runCmd(t, firstPRDir, "git", "config", "--local", "user.name", "atlantisbot")
	runCmd(t, firstPRDir, "git", "config", "--local", "commit.gpgsign", "false")
	runCmd(t, firstPRDir, "git", "merge", "-q", "--no-ff", "-m", "atlantis-merge", "FETCH_HEAD")

	// Simulate second PR with changes to project1.
	runCmd(t, repoDir, "git", "checkout", "main")
	runCmd(t, repoDir, "git", "checkout", "-b", "second-pr")
	runCmd(t, repoDir, "mkdir", "-p", "project1")
	runCmd(t, repoDir, "touch", "project1/variables.tf")
	runCmd(t, repoDir, "git", "add", "project1/variables.tf")
	runCmd(t, repoDir, "git", "commit", "-m", "add variables to project1")
	secondHeadCommit := strings.TrimSpace(runCmd(t, repoDir, "git", "rev-parse", "HEAD"))

	// Atlantis checkout second PR.
	secondPRDir := repoDir + "/second-pr"
	runCmd(t, repoDir, "mkdir", "-p", "second-pr")
	runCmd(t, secondPRDir, "git", "clone", "--branch", "main", "--single-branch", repoDir, ".")
	runCmd(t, secondPRDir, "git", "remote", "add", "source", repoDir)
	runCmd(t, secondPRDir, "git", "fetch", "source", "+refs/heads/second-pr")
	runCmd(t, secondPRDir, "git", "config", "--local", "user.email", "atlantisbot@runatlantis.io")
	runCmd(t, secondPRDir, "git", "config", "--local", "user.name", "atlantisbot")
	runCmd(t, secondPRDir, "git", "config", "--local", "commit.gpgsign", "false")
	runCmd(t, secondPRDir, "git", "merge", "-q", "--no-ff", "-m", "atlantis-merge", "FETCH_HEAD")

	// Merge first PR to create divergence.
	runCmd(t, repoDir, "git", "checkout", "main")
	runCmd(t, repoDir, "git", "merge", "first-pr")

	// Copy the second-pr repo to our data dir which has diverged remote main.
	runCmd(t, repoDir, "mkdir", "-p", "repos/0/")
	runCmd(t, repoDir, "cp", "-R", secondPRDir, "repos/0/default")

	runCmd(t, repoDir+"/repos/0/default", "git", "remote", "update")

	logger := logging.NewNoopLogger(t)

	wd := &events.FileWorkspace{
		DataDir:             repoDir,
		CheckoutMerge:       true,
		CheckoutDepth:       50,
		GpgNoSigningEnabled: true,
	}

	pullRequest := models.PullRequest{
		BaseRepo:   models.Repo{CloneURL: repoDir},
		HeadBranch: "second-pr",
		BaseBranch: "main",
		HeadCommit: secondHeadCommit,
	}

	// Pattern matches files that have diverged (project1 was modified in first-pr and merged to main).
	hasDiverged := wd.HasDivergedWhenModified(logger, repoDir+"/repos/0/default", []string{"project1/**"}, pullRequest)
	Equals(t, true, hasDiverged)
}

func TestHasDivergedWhenModified_PatternHasNotDiverged(t *testing.T) {
	// Initialize the git repo.
	repoDir := initRepo(t)

	// Simulate first PR with changes to project1.
	runCmd(t, repoDir, "git", "checkout", "-b", "first-pr")
	runCmd(t, repoDir, "mkdir", "-p", "project1")
	runCmd(t, repoDir, "touch", "project1/main.tf")
	runCmd(t, repoDir, "git", "add", "project1/main.tf")
	runCmd(t, repoDir, "git", "commit", "-m", "add project1")

	// Merge first PR to main.
	runCmd(t, repoDir, "git", "checkout", "main")
	runCmd(t, repoDir, "git", "merge", "first-pr")

	// Simulate second PR with changes to project1 after main already has project1.
	runCmd(t, repoDir, "git", "checkout", "-b", "second-pr")
	runCmd(t, repoDir, "touch", "project1/variables.tf")
	runCmd(t, repoDir, "git", "add", "project1/variables.tf")
	runCmd(t, repoDir, "git", "commit", "-m", "add variables to project1")
	secondHeadCommit := strings.TrimSpace(runCmd(t, repoDir, "git", "rev-parse", "HEAD"))

	// Atlantis checkout second PR (which includes main's changes since it branched from main).
	secondPRDir := repoDir + "/second-pr"
	runCmd(t, repoDir, "mkdir", "-p", "second-pr")
	runCmd(t, secondPRDir, "git", "clone", "--branch", "main", "--single-branch", repoDir, ".")
	runCmd(t, secondPRDir, "git", "remote", "add", "source", repoDir)
	runCmd(t, secondPRDir, "git", "fetch", "source", "+refs/heads/second-pr")
	runCmd(t, secondPRDir, "git", "config", "--local", "user.email", "atlantisbot@runatlantis.io")
	runCmd(t, secondPRDir, "git", "config", "--local", "user.name", "atlantisbot")
	runCmd(t, secondPRDir, "git", "config", "--local", "commit.gpgsign", "false")
	runCmd(t, secondPRDir, "git", "merge", "-q", "--no-ff", "-m", "atlantis-merge", "FETCH_HEAD")

	// Copy the second-pr repo to our data dir.
	runCmd(t, repoDir, "mkdir", "-p", "repos/0/")
	runCmd(t, repoDir, "cp", "-R", secondPRDir, "repos/0/default")

	runCmd(t, repoDir+"/repos/0/default", "git", "remote", "update")

	logger := logging.NewNoopLogger(t)

	wd := &events.FileWorkspace{
		DataDir:             repoDir,
		CheckoutMerge:       true,
		CheckoutDepth:       50,
		GpgNoSigningEnabled: true,
	}

	pullRequest := models.PullRequest{
		BaseRepo:   models.Repo{CloneURL: repoDir},
		HeadBranch: "second-pr",
		BaseBranch: "main",
		HeadCommit: secondHeadCommit,
	}

	// Pattern matches project1 files. Main's last commit for project1 is in second-pr's history (since it branched from main).
	// So it has not diverged.
	hasDiverged := wd.HasDivergedWhenModified(logger, repoDir+"/repos/0/default", []string{"project1/**"}, pullRequest)
	Equals(t, false, hasDiverged)
}

func TestHasDivergedWhenModified_MultiplePatterns(t *testing.T) {
	// Initialize the git repo.
	repoDir := initRepo(t)

	// Add initial files to project2 and project3 on main.
	runCmd(t, repoDir, "mkdir", "-p", "project2")
	runCmd(t, repoDir, "mkdir", "-p", "project3")
	runCmd(t, repoDir, "touch", "project2/main.tf")
	runCmd(t, repoDir, "touch", "project3/main.tf")
	runCmd(t, repoDir, "git", "add", "project2/main.tf")
	runCmd(t, repoDir, "git", "add", "project3/main.tf")
	runCmd(t, repoDir, "git", "commit", "-m", "add project2 and project3")

	// Simulate PR with additional changes to project2 and project3.
	runCmd(t, repoDir, "git", "checkout", "-b", "feature-pr")
	runCmd(t, repoDir, "touch", "project2/variables.tf")
	runCmd(t, repoDir, "touch", "project3/variables.tf")
	runCmd(t, repoDir, "git", "add", "project2/variables.tf")
	runCmd(t, repoDir, "git", "add", "project3/variables.tf")
	runCmd(t, repoDir, "git", "commit", "-m", "add variables to project2 and project3")
	featureHeadCommit := strings.TrimSpace(runCmd(t, repoDir, "git", "rev-parse", "HEAD"))

	// Atlantis checkout feature PR.
	featurePRDir := repoDir + "/feature-pr"
	runCmd(t, repoDir, "mkdir", "-p", "feature-pr")
	runCmd(t, featurePRDir, "git", "clone", "--branch", "main", "--single-branch", repoDir, ".")
	runCmd(t, featurePRDir, "git", "remote", "add", "source", repoDir)
	runCmd(t, featurePRDir, "git", "fetch", "source", "+refs/heads/feature-pr")
	runCmd(t, featurePRDir, "git", "config", "--local", "user.email", "atlantisbot@runatlantis.io")
	runCmd(t, featurePRDir, "git", "config", "--local", "user.name", "atlantisbot")
	runCmd(t, featurePRDir, "git", "config", "--local", "commit.gpgsign", "false")
	runCmd(t, featurePRDir, "git", "merge", "-q", "--no-ff", "-m", "atlantis-merge", "FETCH_HEAD")

	// Copy the feature-pr repo to our data dir.
	runCmd(t, repoDir, "mkdir", "-p", "repos/0/")
	runCmd(t, repoDir, "cp", "-R", featurePRDir, "repos/0/default")

	runCmd(t, repoDir+"/repos/0/default", "git", "remote", "update")

	logger := logging.NewNoopLogger(t)

	wd := &events.FileWorkspace{
		DataDir:             repoDir,
		CheckoutMerge:       true,
		CheckoutDepth:       50,
		GpgNoSigningEnabled: true,
	}

	pullRequest := models.PullRequest{
		BaseRepo:   models.Repo{CloneURL: repoDir},
		HeadBranch: "feature-pr",
		BaseBranch: "main",
		HeadCommit: featureHeadCommit,
	}

	// Both patterns match files. Main's commits for these patterns are in feature-pr's history.
	// So neither has diverged.
	hasDiverged := wd.HasDivergedWhenModified(logger, repoDir+"/repos/0/default", []string{"project2/**", "project3/**"}, pullRequest)
	Equals(t, false, hasDiverged)
}

func TestHasDivergedWhenModified_PatternMatchesNothing(t *testing.T) {
	// Initialize the git repo.
	repoDir := initRepo(t)

	// Simulate first PR with changes to project1.
	runCmd(t, repoDir, "git", "checkout", "-b", "first-pr")
	runCmd(t, repoDir, "mkdir", "-p", "project1")
	runCmd(t, repoDir, "touch", "project1/main.tf")
	runCmd(t, repoDir, "git", "add", "project1/main.tf")
	runCmd(t, repoDir, "git", "commit", "-m", "add project1")
	headCommit := strings.TrimSpace(runCmd(t, repoDir, "git", "rev-parse", "HEAD"))

	// Atlantis checkout first PR.
	firstPRDir := repoDir + "/first-pr"
	runCmd(t, repoDir, "mkdir", "-p", "first-pr")
	runCmd(t, firstPRDir, "git", "clone", "--branch", "main", "--single-branch", repoDir, ".")
	runCmd(t, firstPRDir, "git", "remote", "add", "source", repoDir)
	runCmd(t, firstPRDir, "git", "fetch", "source", "+refs/heads/first-pr")
	runCmd(t, firstPRDir, "git", "config", "--local", "user.email", "atlantisbot@runatlantis.io")
	runCmd(t, firstPRDir, "git", "config", "--local", "user.name", "atlantisbot")
	runCmd(t, firstPRDir, "git", "config", "--local", "commit.gpgsign", "false")
	runCmd(t, firstPRDir, "git", "merge", "-q", "--no-ff", "-m", "atlantis-merge", "FETCH_HEAD")

	logger := logging.NewNoopLogger(t)

	wd := &events.FileWorkspace{
		DataDir:             repoDir,
		CheckoutMerge:       true,
		CheckoutDepth:       50,
		GpgNoSigningEnabled: true,
	}

	pullRequest := models.PullRequest{
		BaseRepo:   models.Repo{CloneURL: repoDir},
		HeadBranch: "first-pr",
		BaseBranch: "main",
		HeadCommit: headCommit,
	}

	// Pattern matches no files (project2 doesn't exist), should not be considered diverged.
	hasDiverged := wd.HasDivergedWhenModified(logger, firstPRDir, []string{"project2/**"}, pullRequest)
	Equals(t, false, hasDiverged)
}

func TestHasDivergedWhenModified_BrandNewFiles(t *testing.T) {
	// When a PR adds brand new files matching a pattern that main has never touched,
	// this should NOT be considered diverged - main hasn't moved, we're just adding new content.

	// Initialize the git repo.
	repoDir := initRepo(t)

	// Simulate PR that adds brand new project1 files (main has never had project1).
	runCmd(t, repoDir, "git", "checkout", "-b", "new-project-pr")
	runCmd(t, repoDir, "mkdir", "-p", "project1")
	runCmd(t, repoDir, "touch", "project1/main.tf")
	runCmd(t, repoDir, "git", "add", "project1/main.tf")
	runCmd(t, repoDir, "git", "commit", "-m", "add brand new project1")
	headCommit := strings.TrimSpace(runCmd(t, repoDir, "git", "rev-parse", "HEAD"))

	// Atlantis checkout the PR.
	prDir := repoDir + "/new-project-pr"
	runCmd(t, repoDir, "mkdir", "-p", "new-project-pr")
	runCmd(t, prDir, "git", "clone", "--branch", "main", "--single-branch", repoDir, ".")
	runCmd(t, prDir, "git", "remote", "add", "source", repoDir)
	runCmd(t, prDir, "git", "fetch", "source", "+refs/heads/new-project-pr")
	runCmd(t, prDir, "git", "config", "--local", "user.email", "atlantisbot@runatlantis.io")
	runCmd(t, prDir, "git", "config", "--local", "user.name", "atlantisbot")
	runCmd(t, prDir, "git", "config", "--local", "commit.gpgsign", "false")
	runCmd(t, prDir, "git", "merge", "-q", "--no-ff", "-m", "atlantis-merge", "FETCH_HEAD")

	logger := logging.NewNoopLogger(t)

	wd := &events.FileWorkspace{
		DataDir:             repoDir,
		CheckoutMerge:       true,
		CheckoutDepth:       50,
		GpgNoSigningEnabled: true,
	}

	pullRequest := models.PullRequest{
		BaseRepo:   models.Repo{CloneURL: repoDir},
		HeadBranch: "new-project-pr",
		BaseBranch: "main",
		HeadCommit: headCommit,
	}

	// Pattern matches brand new files that main has never touched.
	// Main hasn't "diverged" - we're just adding new content.
	hasDiverged := wd.HasDivergedWhenModified(logger, prDir, []string{"project1/**"}, pullRequest)
	Equals(t, false, hasDiverged)
}

func TestHasDivergedWhenModified_FilesDeletedFromMain(t *testing.T) {
	// When files are deleted from main after PR was created, this IS divergence.
	// Even though remoteRefHash == "", this is different from brand new files because
	// the files existed on main when the PR branched off.

	// Initialize the git repo with project1.
	repoDir := initRepo(t)
	runCmd(t, repoDir, "mkdir", "-p", "project1")
	runCmd(t, repoDir, "touch", "project1/main.tf")
	runCmd(t, repoDir, "git", "add", "project1/main.tf")
	runCmd(t, repoDir, "git", "commit", "-m", "add project1 to main")

	// Create PR that modifies project1.
	runCmd(t, repoDir, "git", "checkout", "-b", "feature-pr")
	runCmd(t, repoDir, "touch", "project1/variables.tf")
	runCmd(t, repoDir, "git", "add", "project1/variables.tf")
	runCmd(t, repoDir, "git", "commit", "-m", "add variables to project1")
	featureHeadCommit := strings.TrimSpace(runCmd(t, repoDir, "git", "rev-parse", "HEAD"))

	// Atlantis checkout feature PR.
	featurePRDir := repoDir + "/feature-pr"
	runCmd(t, repoDir, "mkdir", "-p", "feature-pr")
	runCmd(t, featurePRDir, "git", "clone", "--branch", "main", "--single-branch", repoDir, ".")
	runCmd(t, featurePRDir, "git", "remote", "add", "source", repoDir)
	runCmd(t, featurePRDir, "git", "fetch", "source", "+refs/heads/feature-pr")
	runCmd(t, featurePRDir, "git", "config", "--local", "user.email", "atlantisbot@runatlantis.io")
	runCmd(t, featurePRDir, "git", "config", "--local", "user.name", "atlantisbot")
	runCmd(t, featurePRDir, "git", "config", "--local", "commit.gpgsign", "false")
	runCmd(t, featurePRDir, "git", "merge", "-q", "--no-ff", "-m", "atlantis-merge", "FETCH_HEAD")

	// After PR created, delete project1 from main.
	runCmd(t, repoDir, "git", "checkout", "main")
	runCmd(t, repoDir, "git", "rm", "-r", "project1")
	runCmd(t, repoDir, "git", "commit", "-m", "delete project1")

	// Copy feature-pr repo to our data dir.
	runCmd(t, repoDir, "mkdir", "-p", "repos/0/")
	runCmd(t, repoDir, "cp", "-R", featurePRDir, "repos/0/default")

	runCmd(t, repoDir+"/repos/0/default", "git", "remote", "update")

	logger := logging.NewNoopLogger(t)

	wd := &events.FileWorkspace{
		DataDir:             repoDir,
		CheckoutMerge:       true,
		CheckoutDepth:       50,
		GpgNoSigningEnabled: true,
	}

	pullRequest := models.PullRequest{
		BaseRepo:   models.Repo{CloneURL: repoDir},
		HeadBranch: "feature-pr",
		BaseBranch: "main",
		HeadCommit: featureHeadCommit,
	}

	// Pattern matches project1 which has been deleted from main.
	// This IS divergence - main has moved forward by deleting the files.
	hasDiverged := wd.HasDivergedWhenModified(logger, repoDir+"/repos/0/default", []string{"project1/**"}, pullRequest)
	Equals(t, true, hasDiverged)
}

func TestHasDivergedWhenModified_MixedPatternsOneDiverged(t *testing.T) {
	// Test that when multiple patterns are provided and ONE has diverged,
	// the function correctly returns true (diverged).

	// Initialize the git repo.
	repoDir := initRepo(t)

	// Add project1 and project2 to main.
	runCmd(t, repoDir, "mkdir", "-p", "project1")
	runCmd(t, repoDir, "mkdir", "-p", "project2")
	runCmd(t, repoDir, "touch", "project1/main.tf")
	runCmd(t, repoDir, "touch", "project2/main.tf")
	runCmd(t, repoDir, "git", "add", "project1/main.tf")
	runCmd(t, repoDir, "git", "add", "project2/main.tf")
	runCmd(t, repoDir, "git", "commit", "-m", "add project1 and project2")

	// Create PR that modifies both projects.
	runCmd(t, repoDir, "git", "checkout", "-b", "feature-pr")
	runCmd(t, repoDir, "touch", "project1/variables.tf")
	runCmd(t, repoDir, "touch", "project2/variables.tf")
	runCmd(t, repoDir, "git", "add", "project1/variables.tf")
	runCmd(t, repoDir, "git", "add", "project2/variables.tf")
	runCmd(t, repoDir, "git", "commit", "-m", "modify both projects")
	featureHeadCommit := strings.TrimSpace(runCmd(t, repoDir, "git", "rev-parse", "HEAD"))

	// Atlantis checkout feature PR.
	featurePRDir := repoDir + "/feature-pr"
	runCmd(t, repoDir, "mkdir", "-p", "feature-pr")
	runCmd(t, featurePRDir, "git", "clone", "--branch", "main", "--single-branch", repoDir, ".")
	runCmd(t, featurePRDir, "git", "remote", "add", "source", repoDir)
	runCmd(t, featurePRDir, "git", "fetch", "source", "+refs/heads/feature-pr")
	runCmd(t, featurePRDir, "git", "config", "--local", "user.email", "atlantisbot@runatlantis.io")
	runCmd(t, featurePRDir, "git", "config", "--local", "user.name", "atlantisbot")
	runCmd(t, featurePRDir, "git", "config", "--local", "commit.gpgsign", "false")
	runCmd(t, featurePRDir, "git", "merge", "-q", "--no-ff", "-m", "atlantis-merge", "FETCH_HEAD")

	// After PR created, modify ONLY project1 on main (not project2).
	runCmd(t, repoDir, "git", "checkout", "main")
	runCmd(t, repoDir, "touch", "project1/outputs.tf")
	runCmd(t, repoDir, "git", "add", "project1/outputs.tf")
	runCmd(t, repoDir, "git", "commit", "-m", "modify project1 on main")

	// Copy feature-pr repo to our data dir.
	runCmd(t, repoDir, "mkdir", "-p", "repos/0/")
	runCmd(t, repoDir, "cp", "-R", featurePRDir, "repos/0/default")

	runCmd(t, repoDir+"/repos/0/default", "git", "remote", "update")

	logger := logging.NewNoopLogger(t)

	wd := &events.FileWorkspace{
		DataDir:             repoDir,
		CheckoutMerge:       true,
		CheckoutDepth:       50,
		GpgNoSigningEnabled: true,
	}

	pullRequest := models.PullRequest{
		BaseRepo:   models.Repo{CloneURL: repoDir},
		HeadBranch: "feature-pr",
		BaseBranch: "main",
		HeadCommit: featureHeadCommit,
	}

	// Check with both patterns: project1 (diverged) and project2 (not diverged).
	// Should return true because at least one pattern has diverged.
	hasDiverged := wd.HasDivergedWhenModified(logger, repoDir+"/repos/0/default", []string{"project1/**", "project2/**"}, pullRequest)
	Equals(t, true, hasDiverged)

	// Check with only project2 pattern (not diverged).
	hasDiverged = wd.HasDivergedWhenModified(logger, repoDir+"/repos/0/default", []string{"project2/**"}, pullRequest)
	Equals(t, false, hasDiverged)
}

func TestHasDivergedWhenModified_PRDidNotTouchPattern(t *testing.T) {
	// This test shows the intended behavior: even if the PR didn't modify files matching the pattern,
	// if main did (after PR was created), we consider it diverged. This ensures plans stay current.

	// Initialize the git repo with project1 already existing.
	repoDir := initRepo(t)
	runCmd(t, repoDir, "mkdir", "-p", "project1")
	runCmd(t, repoDir, "touch", "project1/main.tf")
	runCmd(t, repoDir, "git", "add", "project1/main.tf")
	runCmd(t, repoDir, "git", "commit", "-m", "add project1 to main")

	// Create first PR that modifies project2 (not project1).
	runCmd(t, repoDir, "git", "checkout", "-b", "first-pr")
	runCmd(t, repoDir, "mkdir", "-p", "project2")
	runCmd(t, repoDir, "touch", "project2/main.tf")
	runCmd(t, repoDir, "git", "add", "project2/main.tf")
	runCmd(t, repoDir, "git", "commit", "-m", "add project2")
	firstHeadCommit := strings.TrimSpace(runCmd(t, repoDir, "git", "rev-parse", "HEAD"))

	// Atlantis checkout first PR.
	firstPRDir := repoDir + "/first-pr"
	runCmd(t, repoDir, "mkdir", "-p", "first-pr")
	runCmd(t, firstPRDir, "git", "clone", "--branch", "main", "--single-branch", repoDir, ".")
	runCmd(t, firstPRDir, "git", "remote", "add", "source", repoDir)
	runCmd(t, firstPRDir, "git", "fetch", "source", "+refs/heads/first-pr")
	runCmd(t, firstPRDir, "git", "config", "--local", "user.email", "atlantisbot@runatlantis.io")
	runCmd(t, firstPRDir, "git", "config", "--local", "user.name", "atlantisbot")
	runCmd(t, firstPRDir, "git", "config", "--local", "commit.gpgsign", "false")
	runCmd(t, firstPRDir, "git", "merge", "-q", "--no-ff", "-m", "atlantis-merge", "FETCH_HEAD")

	// Create second PR that modifies project1.
	runCmd(t, repoDir, "git", "checkout", "main")
	runCmd(t, repoDir, "git", "checkout", "-b", "second-pr")
	runCmd(t, repoDir, "touch", "project1/variables.tf")
	runCmd(t, repoDir, "git", "add", "project1/variables.tf")
	runCmd(t, repoDir, "git", "commit", "-m", "modify project1")

	// Merge second PR to main (so main now has new project1 changes).
	runCmd(t, repoDir, "git", "checkout", "main")
	runCmd(t, repoDir, "git", "merge", "second-pr")

	// Copy first-pr repo to our data dir.
	runCmd(t, repoDir, "mkdir", "-p", "repos/0/")
	runCmd(t, repoDir, "cp", "-R", firstPRDir, "repos/0/default")

	runCmd(t, repoDir+"/repos/0/default", "git", "remote", "update")

	logger := logging.NewNoopLogger(t)

	wd := &events.FileWorkspace{
		DataDir:             repoDir,
		CheckoutMerge:       true,
		CheckoutDepth:       50,
		GpgNoSigningEnabled: true,
	}

	pullRequest := models.PullRequest{
		BaseRepo:   models.Repo{CloneURL: repoDir},
		HeadBranch: "first-pr",
		BaseBranch: "main",
		HeadCommit: firstHeadCommit,
	}

	// Check divergence for project1 pattern.
	// first-pr never touched project1, but main has moved forward on it.
	// This IS considered diverged - the pattern is in autoplanWhenModified,
	// and we want to ensure plans stay current with main's changes.
	hasDiverged := wd.HasDivergedWhenModified(logger, repoDir+"/repos/0/default", []string{"project1/**"}, pullRequest)
	Equals(t, true, hasDiverged)
}

// TestHasDivergedWhenModified_WithStaleOriginMain tests the actual Atlantis behavior
// where origin/main can become stale and needs to be fetched.
// This simulates:
// 1. Atlantis clones main and merges PR branch (creating merge commit at HEAD)
// 2. Main moves forward with new commits touching the pattern
// 3. HasDivergedWhenModified should detect this by fetching fresh refs
func TestHasDivergedWhenModified_WithStaleOriginMain(t *testing.T) {
	// Initialize the git repo.
	repoDir := initRepo(t)

	// Add initial project1 to main.
	runCmd(t, repoDir, "mkdir", "-p", "project1")
	runCmd(t, repoDir, "touch", "project1/main.tf")
	runCmd(t, repoDir, "git", "add", "project1/main.tf")
	runCmd(t, repoDir, "git", "commit", "-m", "add project1 initial")

	// Create a PR branch from main.
	runCmd(t, repoDir, "git", "checkout", "-b", "feature-pr")
	runCmd(t, repoDir, "mkdir", "-p", "project2")
	runCmd(t, repoDir, "touch", "project2/main.tf")
	runCmd(t, repoDir, "git", "add", "project2/main.tf")
	runCmd(t, repoDir, "git", "commit", "-m", "add project2")
	prHeadCommit := strings.TrimSpace(runCmd(t, repoDir, "git", "rev-parse", "HEAD"))

	// Simulate Atlantis checkout: clone main and merge PR (mimicking forceClone + mergeToBaseBranch).
	atlantisDir := repoDir + "/atlantis-workspace"
	runCmd(t, repoDir, "mkdir", "-p", "atlantis-workspace")
	runCmd(t, atlantisDir, "git", "clone", "--branch", "main", "--single-branch", repoDir, ".")
	runCmd(t, atlantisDir, "git", "remote", "add", "head", repoDir)
	runCmd(t, atlantisDir, "git", "fetch", "head", "+refs/heads/feature-pr:")
	runCmd(t, atlantisDir, "git", "config", "--local", "user.email", "atlantisbot@runatlantis.io")
	runCmd(t, atlantisDir, "git", "config", "--local", "user.name", "atlantisbot")
	runCmd(t, atlantisDir, "git", "config", "--local", "commit.gpgsign", "false")
	runCmd(t, atlantisDir, "git", "merge", "-q", "--no-ff", "-m", "atlantis-merge", "FETCH_HEAD")

	// At this point, atlantisDir has HEAD = merge commit (main + feature-pr).
	// origin/main points to main at the time of clone.

	// Now simulate main moving forward with changes to project1.
	runCmd(t, repoDir, "git", "checkout", "main")
	runCmd(t, repoDir, "touch", "project1/variables.tf")
	runCmd(t, repoDir, "git", "add", "project1/variables.tf")
	runCmd(t, repoDir, "git", "commit", "-m", "update project1 on main")

	// At this point:
	// - atlantisDir's HEAD includes old main + feature-pr
	// - atlantisDir's origin/main is stale (before the "update project1 on main" commit)
	// - repoDir's main has moved forward with a commit touching project1/**
	// - HasDivergedWhenModified should fetch and detect this divergence

	// Copy atlantis workspace to data dir to test.
	runCmd(t, repoDir, "mkdir", "-p", "repos/0/")
	runCmd(t, repoDir, "cp", "-R", atlantisDir, "repos/0/default")

	logger := logging.NewNoopLogger(t)

	wd := &events.FileWorkspace{
		DataDir:             repoDir,
		CheckoutMerge:       true,
		CheckoutDepth:       50,
		GpgNoSigningEnabled: true,
	}

	pullRequest := models.PullRequest{
		BaseRepo:   models.Repo{CloneURL: repoDir},
		HeadBranch: "feature-pr",
		BaseBranch: "main",
		HeadCommit: prHeadCommit,
	}

	// Test 1: Pattern matching project1 should detect divergence.
	// Even though HEAD has a merge commit, origin/main has a new commit touching project1/**
	// that isn't in HEAD. The fetch should pick this up.
	hasDiverged := wd.HasDivergedWhenModified(logger, repoDir+"/repos/0/default", []string{"project1/**"}, pullRequest)
	Equals(t, true, hasDiverged)

	// Test 2: Pattern matching project2 should NOT detect divergence.
	// Main has no new commits touching project2/**, so it hasn't diverged for this pattern.
	hasDiverged = wd.HasDivergedWhenModified(logger, repoDir+"/repos/0/default", []string{"project2/**"}, pullRequest)
	Equals(t, false, hasDiverged)
}

// TestHasDivergedWhenModified_PatternMatching tests that dockerignore-style patterns
// are correctly handled, including *.tf, **/*.tf, and exclusions.
func TestHasDivergedWhenModified_PatternMatching(t *testing.T) {
	tests := []struct {
		name             string
		changedFiles     []string
		patterns         []string
		expectDivergence bool
	}{
		{
			name:             "simple wildcard *.tf matches root level",
			changedFiles:     []string{"main.tf"},
			patterns:         []string{"*.tf"},
			expectDivergence: true,
		},
		{
			name:             "simple wildcard *.tf does not match subdirectory",
			changedFiles:     []string{"project/main.tf"},
			patterns:         []string{"*.tf"},
			expectDivergence: false,
		},
		{
			name:             "recursive **/*.tf matches any level",
			changedFiles:     []string{"project/sub/main.tf"},
			patterns:         []string{"**/*.tf"},
			expectDivergence: true,
		},
		{
			name:             "directory pattern matches files inside",
			changedFiles:     []string{"project/main.tf"},
			patterns:         []string{"project/**"},
			expectDivergence: true,
		},
		{
			name:             "exclusion pattern",
			changedFiles:     []string{"project/excluded.tf"},
			patterns:         []string{"**/*.tf", "!project/excluded.tf"},
			expectDivergence: false,
		},
		{
			name:             "exclusion pattern with directory",
			changedFiles:     []string{"project/excluded/main.tf"},
			patterns:         []string{"**/*.tf", "!project/excluded/**"},
			expectDivergence: false,
		},
		{
			name:             "multiple patterns, one matches",
			changedFiles:     []string{"other/file.hcl"},
			patterns:         []string{"*.tf", "**/*.hcl"},
			expectDivergence: true,
		},
		{
			name:             "exact file match",
			changedFiles:     []string{"terragrunt.hcl"},
			patterns:         []string{"terragrunt.hcl"},
			expectDivergence: true,
		},
		{
			name:             "no pattern match",
			changedFiles:     []string{"README.md"},
			patterns:         []string{"**/*.tf", "**/*.hcl"},
			expectDivergence: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize the git repo.
			repoDir := initRepo(t)

			// Create a PR branch.
			runCmd(t, repoDir, "git", "checkout", "-b", "pr-branch")
			runCmd(t, repoDir, "touch", "pr-file.txt")
			runCmd(t, repoDir, "git", "add", "pr-file.txt")
			runCmd(t, repoDir, "git", "commit", "-m", "pr change")
			prHeadCommit := strings.TrimSpace(runCmd(t, repoDir, "git", "rev-parse", "HEAD"))

			// Atlantis checkout PR.
			prDir := repoDir + "/pr-workspace"
			runCmd(t, repoDir, "mkdir", "-p", "pr-workspace")
			runCmd(t, prDir, "git", "clone", "--branch", "main", "--single-branch", repoDir, ".")
			runCmd(t, prDir, "git", "remote", "add", "head", repoDir)
			runCmd(t, prDir, "git", "fetch", "head", "+refs/heads/pr-branch:")
			runCmd(t, prDir, "git", "config", "--local", "user.email", "atlantisbot@runatlantis.io")
			runCmd(t, prDir, "git", "config", "--local", "user.name", "atlantisbot")
			runCmd(t, prDir, "git", "config", "--local", "commit.gpgsign", "false")
			runCmd(t, prDir, "git", "merge", "-q", "--no-ff", "-m", "atlantis-merge", "FETCH_HEAD")

			// Now add changed files to main to create divergence.
			runCmd(t, repoDir, "git", "checkout", "main")
			for _, file := range tt.changedFiles {
				dir := filepath.Dir(file)
				if dir != "." {
					runCmd(t, repoDir, "mkdir", "-p", dir)
				}
				runCmd(t, repoDir, "touch", file)
				runCmd(t, repoDir, "git", "add", file)
			}
			runCmd(t, repoDir, "git", "commit", "-m", "add test files")

			// Copy to data dir for testing.
			runCmd(t, repoDir, "mkdir", "-p", "repos/0/")
			runCmd(t, repoDir, "cp", "-R", prDir, "repos/0/default")

			logger := logging.NewNoopLogger(t)
			wd := &events.FileWorkspace{
				DataDir:             repoDir,
				CheckoutMerge:       true,
				CheckoutDepth:       50,
				GpgNoSigningEnabled: true,
			}

			pullRequest := models.PullRequest{
				BaseRepo:   models.Repo{CloneURL: repoDir},
				HeadBranch: "pr-branch",
				BaseBranch: "main",
				HeadCommit: prHeadCommit,
			}

			hasDiverged := wd.HasDivergedWhenModified(logger, repoDir+"/repos/0/default", tt.patterns, pullRequest)
			if hasDiverged != tt.expectDivergence {
				t.Errorf("expected divergence=%v, got=%v for patterns=%v and changedFiles=%v",
					tt.expectDivergence, hasDiverged, tt.patterns, tt.changedFiles)
			}
		})
	}
}

func initRepo(t *testing.T) string {
	repoDir := t.TempDir()
	runCmd(t, repoDir, "git", "init", "--initial-branch=main")
	runCmd(t, repoDir, "touch", ".gitkeep")
	runCmd(t, repoDir, "git", "add", ".gitkeep")
	runCmd(t, repoDir, "git", "config", "--local", "user.email", "atlantisbot@runatlantis.io")
	runCmd(t, repoDir, "git", "config", "--local", "user.name", "atlantisbot")
	runCmd(t, repoDir, "git", "config", "--local", "commit.gpgsign", "false")
	runCmd(t, repoDir, "git", "commit", "-m", "initial commit")
	runCmd(t, repoDir, "git", "branch", "branch")
	return repoDir
}
