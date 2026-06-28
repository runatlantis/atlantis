// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestVerifyNonPRBaseBranchReachabilityUnshallowsReachableCommit(t *testing.T) {
	origin := initReachabilityRepo(t)
	oldCommit := strings.TrimSpace(runControllerGit(t, origin, "rev-parse", "old"))
	cloneDir := cloneShallowRef(t, origin, "main")
	runControllerGit(t, cloneDir, "fetch", "--depth=1", "origin", "tag", "old")
	runControllerGit(t, cloneDir, "checkout", "--detach", oldCommit)

	_, err := runControllerGitAllowError(cloneDir, "merge-base", "--is-ancestor", oldCommit, "refs/remotes/origin/main")
	Assert(t, err != nil, "expected shallow merge-base check to fail before unshallowing")

	ctx := &command.Context{
		Log: logging.NewNoopLogger(t),
		Pull: models.PullRequest{
			Num:        -1,
			BaseBranch: "main",
			HeadBranch: "refs/tags/old",
			HeadCommit: oldCommit,
		},
	}
	Ok(t, verifyNonPRBaseBranchReachability(ctx, cloneDir))
	_, err = os.Stat(filepath.Join(cloneDir, ".git", "shallow"))
	Assert(t, os.IsNotExist(err), "expected verification to unshallow the checkout")
}

func TestVerifyNonPRBaseBranchReachabilityRejectsUnreachableCommitAfterUnshallow(t *testing.T) {
	origin := initReachabilityRepo(t)
	runControllerGit(t, origin, "checkout", "--orphan", "unrelated")
	runControllerGit(t, origin, "commit", "--allow-empty", "-m", "unrelated")
	unrelatedCommit := strings.TrimSpace(runControllerGit(t, origin, "rev-parse", "HEAD"))
	runControllerGit(t, origin, "tag", "unrelated")
	cloneDir := cloneShallowRef(t, origin, "unrelated")

	ctx := &command.Context{
		Log: logging.NewNoopLogger(t),
		Pull: models.PullRequest{
			Num:        -1,
			BaseBranch: "main",
			HeadBranch: "refs/tags/unrelated",
			HeadCommit: unrelatedCommit,
		},
	}
	err := verifyNonPRBaseBranchReachability(ctx, cloneDir)
	ErrContains(t, "is not reachable from base_branch", err)
}

func initReachabilityRepo(t *testing.T) string {
	t.Helper()
	repoDir := t.TempDir()
	runControllerGit(t, repoDir, "init", "--initial-branch=main")
	runControllerGit(t, repoDir, "config", "user.email", "atlantisbot@runatlantis.io")
	runControllerGit(t, repoDir, "config", "user.name", "atlantisbot")
	runControllerGit(t, repoDir, "config", "commit.gpgsign", "false")
	runControllerGit(t, repoDir, "commit", "--allow-empty", "-m", "old")
	runControllerGit(t, repoDir, "tag", "old")
	runControllerGit(t, repoDir, "commit", "--allow-empty", "-m", "new")
	return repoDir
}

func cloneShallowRef(t *testing.T, origin string, ref string) string {
	t.Helper()
	cloneDir := filepath.Join(t.TempDir(), "clone")
	cmd := exec.Command("git", "clone", "--depth=1", "--branch", ref, "file://"+origin, cloneDir) //nolint:gosec
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("running git clone: %s: %v", output, err)
	}
	return cloneDir
}

func runControllerGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	output, err := runControllerGitAllowError(dir, args...)
	if err != nil {
		t.Fatalf("running git %s: %s: %v", strings.Join(args, " "), output, err)
	}
	return output
}

func runControllerGitAllowError(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...) //nolint:gosec
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	return string(output), err
}
