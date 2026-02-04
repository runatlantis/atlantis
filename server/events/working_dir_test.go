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

// Test that if the base branch is missing from the remote, we re-clone.
func TestClone_MissingRemoteBranch(t *testing.T) {
	// 1. Initialize remote repo
	repoDir := initRepo(t)
	dataDir := t.TempDir()

	logger := logging.NewNoopLogger(t)

	wd := &events.FileWorkspace{
		DataDir:                     dataDir,
		CheckoutMerge:               false,
		TestingOverrideHeadCloneURL: fmt.Sprintf("file://%s", repoDir),
		GpgNoSigningEnabled:         true,
	}

	// 2. Initial successful clone
	expCommit := runCmd(t, repoDir, "git", "rev-parse", "HEAD")
	cloneDir, err := wd.Clone(logger, models.Repo{}, models.PullRequest{
		BaseRepo:   models.Repo{},
		HeadBranch: "branch",
		HeadCommit: expCommit,
		BaseBranch: "main",
	}, "default")
	Ok(t, err)

	// Create a plan file that should be deleted on re-clone.
	planFile := filepath.Join(cloneDir, "default.tfplan")
	_, err = os.Create(planFile)
	Ok(t, err)
	assert.FileExists(t, planFile)

	// 3. Create a new commit on the head branch and delete 'main' branch from remote.
	runCmd(t, repoDir, "git", "checkout", "branch")
	newFile := filepath.Join(repoDir, "new-file.txt")
	err = os.WriteFile(newFile, []byte("new content"), 0o600)
	Ok(t, err)
	runCmd(t, repoDir, "git", "add", "new-file.txt")
	runCmd(t, repoDir, "git", "commit", "-m", "new commit for head branch")
	newHeadCommit := runCmd(t, repoDir, "git", "rev-parse", "HEAD")
	runCmd(t, repoDir, "git", "branch", "-D", "main")

	// 4. Trigger another clone, expecting re-clone because the base branch 'main'
	// is missing from the remote and the head commit has changed.
	_, err = wd.Clone(logger, models.Repo{}, models.PullRequest{
		BaseRepo:   models.Repo{},
		HeadBranch: "branch",
		HeadCommit: newHeadCommit,
		BaseBranch: "main",
	}, "default")
	Ok(t, err)

	// After re-clone, the plan file should be gone.
	assert.NoFileExists(t, planFile, "Plan file should have been wiped out by the reclone")
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
