// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

func TestResolveNonPRHeadCommitSkipsNonSyntheticPull(t *testing.T) {
	repoDir := initReachabilityRepo(t)
	ctx := &command.Context{
		Pull: models.PullRequest{
			Num:        0,
			HeadCommit: "main",
		},
	}

	Ok(t, resolveNonPRHeadCommit(ctx, repoDir))
	Equals(t, "main", ctx.Pull.HeadCommit)
}

func TestResolveNonPRHeadCommitUpdatesSyntheticNegativePull(t *testing.T) {
	repoDir := initReachabilityRepo(t)
	headCommit := strings.TrimSpace(runControllerGit(t, repoDir, "rev-parse", "HEAD"))
	ctx := &command.Context{
		Pull: models.PullRequest{
			Num:        -1,
			HeadCommit: "main",
		},
	}

	Ok(t, resolveNonPRHeadCommit(ctx, repoDir))
	Equals(t, headCommit, ctx.Pull.HeadCommit)
}

func TestResolveAPIHeadCommitUpdatesPRBranchRef(t *testing.T) {
	repoDir := initReachabilityRepo(t)
	headCommit := strings.TrimSpace(runControllerGit(t, repoDir, "rev-parse", "HEAD"))
	ctx := &command.Context{
		Pull: models.PullRequest{
			Num:        123,
			HeadCommit: "main",
		},
	}

	Ok(t, resolveAPIHeadCommit(ctx, repoDir, false))
	Equals(t, headCommit, ctx.Pull.HeadCommit)
}

func TestAPIResolvePRHead_BranchCheckoutUsesHEADWhenHEADIsMergeCommit(t *testing.T) {
	repoDir := initReachabilityRepo(t)
	runControllerGit(t, repoDir, "checkout", "-b", "feature", "old")
	runControllerGit(t, repoDir, "commit", "--allow-empty", "-m", "feature")
	runControllerGit(t, repoDir, "checkout", "main")
	runControllerGit(t, repoDir, "merge", "--no-ff", "feature", "-m", "merge feature")
	mergeHead := strings.TrimSpace(runControllerGit(t, repoDir, "rev-parse", "HEAD"))
	secondParent := strings.TrimSpace(runControllerGit(t, repoDir, "rev-parse", "HEAD^2"))
	ctx := &command.Context{
		Pull: models.PullRequest{
			Num:        123,
			HeadCommit: "main",
		},
	}

	Ok(t, resolveAPIHeadCommit(ctx, repoDir, false))
	Equals(t, mergeHead, ctx.Pull.HeadCommit)
	Assert(t, ctx.Pull.HeadCommit != secondParent, "expected branch checkout to use HEAD, not HEAD^2")
}

func TestAPIResolvePRHead_MergeCheckoutUsesHEADSecondParent(t *testing.T) {
	repoDir := initReachabilityRepo(t)
	runControllerGit(t, repoDir, "checkout", "-b", "feature", "old")
	runControllerGit(t, repoDir, "commit", "--allow-empty", "-m", "feature")
	featureHead := strings.TrimSpace(runControllerGit(t, repoDir, "rev-parse", "HEAD"))
	runControllerGit(t, repoDir, "checkout", "main")
	runControllerGit(t, repoDir, "merge", "--no-ff", "feature", "-m", "merge feature")
	mergeHead := strings.TrimSpace(runControllerGit(t, repoDir, "rev-parse", "HEAD"))
	ctx := &command.Context{
		Pull: models.PullRequest{
			Num:        123,
			HeadCommit: "main",
		},
	}

	Ok(t, resolveAPIHeadCommit(ctx, repoDir, true))
	Equals(t, featureHead, ctx.Pull.HeadCommit)
	Assert(t, ctx.Pull.HeadCommit != mergeHead, "expected PR head, not merge commit")
}

func initReachabilityRepo(t *testing.T) string {
	t.Helper()
	repoDir := newReachabilityGitTempDir(t, "reachability-origin-*")
	runControllerGit(t, repoDir, "init", "--initial-branch=main")
	disableReachabilityGitMaintenance(t, repoDir)
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
	cloneDir := filepath.Join(newReachabilityGitTempDir(t, "reachability-clone-*"), "clone")
	cmd := exec.Command("git", "clone", "--depth=1", "--branch", ref, "file://"+origin, cloneDir) //nolint:gosec
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("running git clone: %s: %v", output, err)
	}
	disableReachabilityGitMaintenance(t, cloneDir)
	return cloneDir
}

func newReachabilityGitTempDir(t *testing.T, pattern string) string {
	t.Helper()
	dir, err := os.MkdirTemp("", pattern)
	Ok(t, err)
	t.Cleanup(func() {
		for range 5 {
			if err := os.RemoveAll(dir); err == nil {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	})
	return dir
}

func disableReachabilityGitMaintenance(t *testing.T, repoDir string) {
	t.Helper()
	runControllerGit(t, repoDir, "config", "--local", "gc.auto", "0")
	runControllerGit(t, repoDir, "config", "--local", "maintenance.auto", "false")
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
