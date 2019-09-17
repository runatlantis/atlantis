package events_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

// Test that if we don't have any existing files, we check out the repo.
func TestClone_NoneExisting(t *testing.T) {
	// Initialize the git repo.
	repoDir, cleanup := initRepo(t)
	defer cleanup()
	expCommit := runCmd(t, repoDir, "git", "rev-parse", "HEAD")

	dataDir, cleanup2 := TempDir(t)
	defer cleanup2()

	wd := &events.FileWorkspace{
		DataDir:                     dataDir,
		CheckoutMerge:               false,
		TestingOverrideHeadCloneURL: fmt.Sprintf("file://%s", repoDir),
	}

	cloneDir, err := wd.Clone(nil, models.Repo{}, models.Repo{}, models.PullRequest{
		HeadBranch: "branch",
	}, "default", []string{})
	Ok(t, err)

	// Use rev-parse to verify at correct commit.
	actCommit := runCmd(t, cloneDir, "git", "rev-parse", "HEAD")
	Equals(t, expCommit, actCommit)
}

// Test that if we don't have any existing files, we check out the repo
// successfully when we're using the merge method.
func TestClone_CheckoutMergeNoneExisting(t *testing.T) {
	// Initialize the git repo.
	repoDir, cleanup := initRepo(t)
	defer cleanup()

	// Add a commit to branch 'branch' that's not on master.
	runCmd(t, repoDir, "git", "checkout", "branch")
	runCmd(t, repoDir, "touch", "branch-file")
	runCmd(t, repoDir, "git", "add", "branch-file")
	runCmd(t, repoDir, "git", "commit", "-m", "branch-commit")
	branchCommit := runCmd(t, repoDir, "git", "rev-parse", "HEAD")

	// Now switch back to master and advance the master branch by another
	// commit.
	runCmd(t, repoDir, "git", "checkout", "master")
	runCmd(t, repoDir, "touch", "master-file")
	runCmd(t, repoDir, "git", "add", "master-file")
	runCmd(t, repoDir, "git", "commit", "-m", "master-commit")
	masterCommit := runCmd(t, repoDir, "git", "rev-parse", "HEAD")

	// Finally, perform a merge in another branch ourselves, just so we know
	// what the final state of the repo should be.
	runCmd(t, repoDir, "git", "checkout", "-b", "mergetest")
	runCmd(t, repoDir, "git", "merge", "-m", "atlantis-merge", "branch")
	expLsOutput := runCmd(t, repoDir, "ls")

	dataDir, cleanup2 := TempDir(t)
	defer cleanup2()

	overrideURL := fmt.Sprintf("file://%s", repoDir)
	wd := &events.FileWorkspace{
		DataDir:                     dataDir,
		CheckoutMerge:               true,
		TestingOverrideHeadCloneURL: overrideURL,
		TestingOverrideBaseCloneURL: overrideURL,
	}

	cloneDir, err := wd.Clone(nil, models.Repo{}, models.Repo{}, models.PullRequest{
		HeadBranch: "branch",
		BaseBranch: "master",
	}, "default", []string{})
	Ok(t, err)

	// Check the commits.
	actBaseCommit := runCmd(t, cloneDir, "git", "rev-parse", "HEAD~1")
	actHeadCommit := runCmd(t, cloneDir, "git", "rev-parse", "HEAD^2")
	Equals(t, masterCommit, actBaseCommit)
	Equals(t, branchCommit, actHeadCommit)

	// Use ls to verify the repo looks good.
	actLsOutput := runCmd(t, cloneDir, "ls")
	Equals(t, expLsOutput, actLsOutput)
}

// Test that if we're using the merge method and the repo is already cloned at
// the right commit, then we don't reclone.
func TestClone_CheckoutMergeNoReclone(t *testing.T) {
	// Initialize the git repo.
	repoDir, cleanup := initRepo(t)
	defer cleanup()

	// Add a commit to branch 'branch' that's not on master.
	runCmd(t, repoDir, "git", "checkout", "branch")
	runCmd(t, repoDir, "touch", "branch-file")
	runCmd(t, repoDir, "git", "add", "branch-file")
	runCmd(t, repoDir, "git", "commit", "-m", "branch-commit")

	// Now switch back to master and advance the master branch by another commit.
	runCmd(t, repoDir, "git", "checkout", "master")
	runCmd(t, repoDir, "touch", "master-file")
	runCmd(t, repoDir, "git", "add", "master-file")
	runCmd(t, repoDir, "git", "commit", "-m", "master-commit")

	// Run the clone for the first time.
	dataDir, cleanup2 := TempDir(t)
	defer cleanup2()
	overrideURL := fmt.Sprintf("file://%s", repoDir)
	wd := &events.FileWorkspace{
		DataDir:                     dataDir,
		CheckoutMerge:               true,
		TestingOverrideHeadCloneURL: overrideURL,
		TestingOverrideBaseCloneURL: overrideURL,
	}

	_, err := wd.Clone(nil, models.Repo{}, models.Repo{}, models.PullRequest{
		HeadBranch: "branch",
		BaseBranch: "master",
	}, "default", []string{})
	Ok(t, err)

	// Create a file that we can use to check if the repo was recloned.
	runCmd(t, dataDir, "touch", "repos/0/default/proof")

	// Now run the clone again.
	cloneDir, err := wd.Clone(nil, models.Repo{}, models.Repo{}, models.PullRequest{
		HeadBranch: "branch",
		BaseBranch: "master",
	}, "default", []string{})
	Ok(t, err)

	// Check that our proof file is still there, proving that we didn't reclone.
	_, err = os.Stat(filepath.Join(cloneDir, "proof"))
	Ok(t, err)
}

// Same as TestClone_CheckoutMergeNoReclone however the branch that gets
// merged is a fast-forward merge. See #584.
func TestClone_CheckoutMergeNoRecloneFastForward(t *testing.T) {
	// Initialize the git repo.
	repoDir, cleanup := initRepo(t)
	defer cleanup()

	// Add a commit to branch 'branch' that's not on master.
	// This will result in a fast-forwardable merge.
	runCmd(t, repoDir, "git", "checkout", "branch")
	runCmd(t, repoDir, "touch", "branch-file")
	runCmd(t, repoDir, "git", "add", "branch-file")
	runCmd(t, repoDir, "git", "commit", "-m", "branch-commit")

	// Run the clone for the first time.
	dataDir, cleanup2 := TempDir(t)
	defer cleanup2()
	overrideURL := fmt.Sprintf("file://%s", repoDir)
	wd := &events.FileWorkspace{
		DataDir:                     dataDir,
		CheckoutMerge:               true,
		TestingOverrideHeadCloneURL: overrideURL,
		TestingOverrideBaseCloneURL: overrideURL,
	}

	_, err := wd.Clone(nil, models.Repo{}, models.Repo{}, models.PullRequest{
		HeadBranch: "branch",
		BaseBranch: "master",
	}, "default", []string{})
	Ok(t, err)

	// Create a file that we can use to check if the repo was recloned.
	runCmd(t, dataDir, "touch", "repos/0/default/proof")

	// Now run the clone again.
	cloneDir, err := wd.Clone(nil, models.Repo{}, models.Repo{}, models.PullRequest{
		HeadBranch: "branch",
		BaseBranch: "master",
	}, "default", []string{})
	Ok(t, err)

	// Check that our proof file is still there, proving that we didn't reclone.
	_, err = os.Stat(filepath.Join(cloneDir, "proof"))
	Ok(t, err)
}

// Test that if there's a conflict when merging we return a good error.
func TestClone_CheckoutMergeConflict(t *testing.T) {
	// Initialize the git repo.
	repoDir, cleanup := initRepo(t)
	defer cleanup()

	// Add a commit to branch 'branch' that's not on master.
	runCmd(t, repoDir, "git", "checkout", "branch")
	runCmd(t, repoDir, "sh", "-c", "echo hi >> file")
	runCmd(t, repoDir, "git", "add", "file")
	runCmd(t, repoDir, "git", "commit", "-m", "branch-commit")

	// Add a new commit to master that will cause a conflict if branch was
	// merged.
	runCmd(t, repoDir, "git", "checkout", "master")
	runCmd(t, repoDir, "sh", "-c", "echo conflict >> file")
	runCmd(t, repoDir, "git", "add", "file")
	runCmd(t, repoDir, "git", "commit", "-m", "commit")

	// We're set up, now trigger the Atlantis clone.
	dataDir, cleanup2 := TempDir(t)
	defer cleanup2()
	overrideURL := fmt.Sprintf("file://%s", repoDir)
	wd := &events.FileWorkspace{
		DataDir:                     dataDir,
		CheckoutMerge:               true,
		TestingOverrideHeadCloneURL: overrideURL,
		TestingOverrideBaseCloneURL: overrideURL,
	}

	_, err := wd.Clone(nil, models.Repo{}, models.Repo{}, models.PullRequest{
		HeadBranch: "branch",
		BaseBranch: "master",
	}, "default", []string{})
	ErrEquals(t, `running git merge -q --no-ff -m atlantis-merge FETCH_HEAD: Auto-merging file
CONFLICT (add/add): Merge conflict in file
Automatic merge failed; fix conflicts and then commit the result.
: exit status 1`, err)
}

// Test that if the repo is already cloned and is at the right commit, we
// don't reclone.
func TestClone_NoReclone(t *testing.T) {
	repoDir, _ := initRepo(t)
	dataDir, cleanup2 := TempDir(t)
	defer cleanup2()

	runCmd(t, dataDir, "mkdir", "-p", "repos/0/")
	runCmd(t, dataDir, "mv", repoDir, "repos/0/default")
	// Create a file that we can use later to check if the repo was recloned.
	runCmd(t, dataDir, "touch", "repos/0/default/proof")

	wd := &events.FileWorkspace{
		DataDir:                     dataDir,
		CheckoutMerge:               false,
		TestingOverrideHeadCloneURL: fmt.Sprintf("file://%s", repoDir),
	}
	cloneDir, err := wd.Clone(nil, models.Repo{}, models.Repo{}, models.PullRequest{
		HeadBranch: "branch",
	}, "default", []string{})
	Ok(t, err)

	// Check that our proof file is still there.
	_, err = os.Stat(filepath.Join(cloneDir, "proof"))
	Ok(t, err)
}

// Test that if the repo is already cloned but is at the wrong commit, we
// reclone.
func TestClone_RecloneWrongCommit(t *testing.T) {
	repoDir, cleanup := initRepo(t)
	defer cleanup()
	dataDir, cleanup2 := TempDir(t)
	defer cleanup2()

	// Copy the repo to our data dir.
	runCmd(t, dataDir, "mkdir", "-p", "repos/0/")
	runCmd(t, dataDir, "cp", "-R", repoDir, "repos/0/default")

	// Now add a commit to the repo, so the one in the data dir is out of date.
	runCmd(t, repoDir, "git", "checkout", "branch")
	runCmd(t, repoDir, "touch", "newfile")
	runCmd(t, repoDir, "git", "add", "newfile")
	runCmd(t, repoDir, "git", "commit", "-m", "newfile")
	expCommit := runCmd(t, repoDir, "git", "rev-parse", "HEAD")

	wd := &events.FileWorkspace{
		DataDir:                     dataDir,
		CheckoutMerge:               false,
		TestingOverrideHeadCloneURL: fmt.Sprintf("file://%s", repoDir),
	}
	cloneDir, err := wd.Clone(nil, models.Repo{}, models.Repo{}, models.PullRequest{
		HeadBranch: "branch",
		HeadCommit: expCommit,
	}, "default", []string{})
	Ok(t, err)

	// Use rev-parse to verify at correct commit.
	actCommit := runCmd(t, cloneDir, "git", "rev-parse", "HEAD")
	Equals(t, expCommit, actCommit)
}

func initRepo(t *testing.T) (string, func()) {
	repoDir, cleanup := TempDir(t)
	runCmd(t, repoDir, "git", "init")
	runCmd(t, repoDir, "touch", ".gitkeep")
	runCmd(t, repoDir, "git", "add", ".gitkeep")
	runCmd(t, repoDir, "git", "config", "--local", "user.email", "atlantisbot@runatlantis.io")
	runCmd(t, repoDir, "git", "config", "--local", "user.name", "atlantisbot")
	runCmd(t, repoDir, "git", "commit", "-m", "initial commit")
	runCmd(t, repoDir, "git", "branch", "branch")
	return repoDir, cleanup
}
