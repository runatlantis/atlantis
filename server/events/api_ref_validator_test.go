// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

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
)

func TestValidateNonPRAPIRefUnchangedFailsWhenMutableBranchMoves(t *testing.T) {
	repoDir, initialCommit := initAPIRefValidatorGitRepo(t)
	advanceAPIRefValidatorGitMain(t, repoDir)
	ctx := command.ProjectContext{
		Log: logging.NewNoopLogger(t),
		API: true,
		Pull: models.PullRequest{
			Num:        -1,
			HeadBranch: "main",
			HeadCommit: initialCommit,
		},
	}

	err := ValidateNonPRAPIRefUnchanged(ctx, repoDir)

	if err == nil {
		t.Fatal("expected mutable API ref change to fail")
	}
	if !strings.Contains(err.Error(), "changed") {
		t.Fatalf("expected changed-ref error, got %v", err)
	}
}

func TestValidateNonPRAPIRefUnchangedAllowsImmutableSHA(t *testing.T) {
	repoDir, initialCommit := initAPIRefValidatorGitRepo(t)
	advanceAPIRefValidatorGitMain(t, repoDir)
	ctx := command.ProjectContext{
		Log: logging.NewNoopLogger(t),
		API: true,
		Pull: models.PullRequest{
			Num:        -1,
			HeadBranch: initialCommit,
			HeadCommit: initialCommit,
		},
	}

	if err := ValidateNonPRAPIRefUnchanged(ctx, repoDir); err != nil {
		t.Fatalf("expected immutable API SHA to pass, got %v", err)
	}
}

func TestValidateNonPRAPIRefUnchangedAllowsBareTagLikeRef(t *testing.T) {
	repoDir, initialCommit := initAPIRefValidatorGitRepo(t)
	createAPIRefValidatorGitBranch(t, repoDir, "v1.0.0", initialCommit)
	advanceAPIRefValidatorGitMain(t, repoDir)
	ctx := command.ProjectContext{
		Log: logging.NewNoopLogger(t),
		API: true,
		Pull: models.PullRequest{
			Num:        -1,
			HeadBranch: "v1.0.0",
			HeadCommit: initialCommit,
		},
	}

	if err := ValidateNonPRAPIRefUnchanged(ctx, repoDir); err != nil {
		t.Fatalf("expected bare tag-like API ref to pass, got %v", err)
	}
}

func TestValidateNonPRAPIRefUnchanged_BranchNamedReleaseIsRevalidated(t *testing.T) {
	repoDir, initialCommit := initAPIRefValidatorGitRepo(t)
	createAPIRefValidatorGitBranch(t, repoDir, "release", initialCommit)
	advanceAPIRefValidatorGitBranch(t, repoDir, "release")
	ctx := command.ProjectContext{
		Log: logging.NewNoopLogger(t),
		API: true,
		Pull: models.PullRequest{
			Num:        -1,
			HeadBranch: "release",
			HeadCommit: initialCommit,
		},
	}

	err := ValidateNonPRAPIRefUnchanged(ctx, repoDir)

	if err == nil {
		t.Fatal("expected moved release branch to fail")
	}
	if !strings.Contains(err.Error(), "changed") {
		t.Fatalf("expected changed-ref error, got %v", err)
	}
}

func TestValidateNonPRAPIRefUnchanged_BranchNamedStableIsRevalidated(t *testing.T) {
	repoDir, initialCommit := initAPIRefValidatorGitRepo(t)
	createAPIRefValidatorGitBranch(t, repoDir, "stable", initialCommit)
	advanceAPIRefValidatorGitBranch(t, repoDir, "stable")
	ctx := command.ProjectContext{
		Log: logging.NewNoopLogger(t),
		API: true,
		Pull: models.PullRequest{
			Num:        -1,
			HeadBranch: "stable",
			HeadCommit: initialCommit,
		},
	}

	err := ValidateNonPRAPIRefUnchanged(ctx, repoDir)

	if err == nil {
		t.Fatal("expected moved stable branch to fail")
	}
	if !strings.Contains(err.Error(), "changed") {
		t.Fatalf("expected changed-ref error, got %v", err)
	}
}

func TestValidateNonPRAPIRefUnchanged_TagLikeBranchNameIsRevalidated(t *testing.T) {
	repoDir, initialCommit := initAPIRefValidatorGitRepo(t)
	createAPIRefValidatorGitBranch(t, repoDir, "v1.2.3", initialCommit)
	advanceAPIRefValidatorGitBranch(t, repoDir, "v1.2.3")
	ctx := command.ProjectContext{
		Log: logging.NewNoopLogger(t),
		API: true,
		Pull: models.PullRequest{
			Num:        -1,
			HeadBranch: "v1.2.3",
			HeadCommit: initialCommit,
		},
	}

	err := ValidateNonPRAPIRefUnchanged(ctx, repoDir)

	if err == nil {
		t.Fatal("expected moved tag-like branch to fail")
	}
	if !strings.Contains(err.Error(), "changed") {
		t.Fatalf("expected changed-ref error, got %v", err)
	}
}

func TestValidateNonPRAPIRefUnchanged_AnnotatedTagPeelsToCommit(t *testing.T) {
	repoDir, initialCommit := initAPIRefValidatorGitRepo(t)
	runAPIRefValidatorGit(t, repoDir, "tag", "-a", "v1.0.0", "-m", "release", initialCommit)
	runAPIRefValidatorGit(t, repoDir, "push", "-q", "origin", "refs/tags/v1.0.0")
	ctx := command.ProjectContext{
		Log: logging.NewNoopLogger(t),
		API: true,
		Pull: models.PullRequest{
			Num:        -1,
			HeadBranch: "refs/tags/v1.0.0",
			HeadCommit: initialCommit,
		},
	}

	if err := ValidateNonPRAPIRefUnchanged(ctx, repoDir); err != nil {
		t.Fatalf("expected annotated tag to peel to commit, got %v", err)
	}
}

func TestValidateNonPRAPIRefUnchanged_LightweightTagStillMatches(t *testing.T) {
	repoDir, initialCommit := initAPIRefValidatorGitRepo(t)
	runAPIRefValidatorGit(t, repoDir, "tag", "lightweight", initialCommit)
	runAPIRefValidatorGit(t, repoDir, "push", "-q", "origin", "refs/tags/lightweight")
	ctx := command.ProjectContext{
		Log: logging.NewNoopLogger(t),
		API: true,
		Pull: models.PullRequest{
			Num:        -1,
			HeadBranch: "refs/tags/lightweight",
			HeadCommit: initialCommit,
		},
	}

	if err := ValidateNonPRAPIRefUnchanged(ctx, repoDir); err != nil {
		t.Fatalf("expected lightweight tag to match, got %v", err)
	}
}

func TestValidateNonPRAPIRefUnchanged_AnnotatedTagChangedCommitFails(t *testing.T) {
	repoDir, initialCommit := initAPIRefValidatorGitRepo(t)
	runAPIRefValidatorGit(t, repoDir, "tag", "-a", "v1.0.0", "-m", "release", initialCommit)
	runAPIRefValidatorGit(t, repoDir, "push", "-q", "origin", "refs/tags/v1.0.0")
	advanceAPIRefValidatorGitMain(t, repoDir)
	runAPIRefValidatorGit(t, repoDir, "tag", "-f", "-a", "v1.0.0", "-m", "release 2")
	runAPIRefValidatorGit(t, repoDir, "push", "-q", "-f", "origin", "refs/tags/v1.0.0")
	ctx := command.ProjectContext{
		Log: logging.NewNoopLogger(t),
		API: true,
		Pull: models.PullRequest{
			Num:        -1,
			HeadBranch: "refs/tags/v1.0.0",
			HeadCommit: initialCommit,
		},
	}

	err := ValidateNonPRAPIRefUnchanged(ctx, repoDir)

	if err == nil {
		t.Fatal("expected moved annotated tag to fail")
	}
	if !strings.Contains(err.Error(), "changed") {
		t.Fatalf("expected changed-ref error, got %v", err)
	}
}

func TestValidateNonPRAPIRefUnchanged_UnsafeTagRefRejected(t *testing.T) {
	repoDir, initialCommit := initAPIRefValidatorGitRepo(t)
	ctx := command.ProjectContext{
		Log: logging.NewNoopLogger(t),
		API: true,
		Pull: models.PullRequest{
			Num:        -1,
			HeadBranch: "refs/tags/../bad",
			HeadCommit: initialCommit,
		},
	}

	err := ValidateNonPRAPIRefUnchanged(ctx, repoDir)

	if err == nil {
		t.Fatal("expected unsafe tag ref to fail")
	}
	if !strings.Contains(err.Error(), "invalid API ref") {
		t.Fatalf("expected invalid API ref error, got %v", err)
	}
}

func TestValidateNonPRAPIRefUnchangedAllowsNonGitDir(t *testing.T) {
	ctx := command.ProjectContext{
		Log: logging.NewNoopLogger(t),
		API: true,
		Pull: models.PullRequest{
			Num:        -1,
			HeadBranch: "main",
			HeadCommit: strings.Repeat("a", 40),
		},
	}

	if err := ValidateNonPRAPIRefUnchanged(ctx, t.TempDir()); err != nil {
		t.Fatalf("expected non-git API dir to pass, got %v", err)
	}
}

func initAPIRefValidatorGitRepo(t *testing.T) (string, string) {
	t.Helper()
	root, err := os.MkdirTemp("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		for range 10 {
			if err := os.RemoveAll(root); err == nil || os.IsNotExist(err) {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
		if err := os.RemoveAll(root); err != nil && !os.IsNotExist(err) {
			t.Fatalf("removing git fixture: %v", err)
		}
	})
	originDir := filepath.Join(root, "origin.git")
	repoDir := filepath.Join(root, "work")
	runAPIRefValidatorGit(t, "", "init", "--bare", originDir)
	runAPIRefValidatorGit(t, originDir, "config", "gc.auto", "0")
	runAPIRefValidatorGit(t, originDir, "config", "maintenance.auto", "false")
	runAPIRefValidatorGit(t, originDir, "config", "receive.autogc", "false")
	runAPIRefValidatorGit(t, "", "init", "--initial-branch=main", repoDir)
	runAPIRefValidatorGit(t, repoDir, "config", "gc.auto", "0")
	runAPIRefValidatorGit(t, repoDir, "config", "maintenance.auto", "false")
	runAPIRefValidatorGit(t, repoDir, "config", "user.email", "atlantisbot@runatlantis.io")
	runAPIRefValidatorGit(t, repoDir, "config", "user.name", "atlantisbot")
	runAPIRefValidatorGit(t, repoDir, "config", "commit.gpgsign", "false")
	runAPIRefValidatorGit(t, repoDir, "config", "tag.gpgsign", "false")
	if err := os.WriteFile(filepath.Join(repoDir, "main.tf"), []byte("resource \"null_resource\" \"main\" {}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	runAPIRefValidatorGit(t, repoDir, "add", "main.tf")
	runAPIRefValidatorGit(t, repoDir, "commit", "-q", "-m", "initial")
	initialCommit := strings.TrimSpace(runAPIRefValidatorGit(t, repoDir, "rev-parse", "HEAD"))
	runAPIRefValidatorGit(t, repoDir, "remote", "add", "origin", "file://"+originDir)
	runAPIRefValidatorGit(t, repoDir, "push", "-q", "-u", "origin", "main")
	return repoDir, initialCommit
}

func advanceAPIRefValidatorGitMain(t *testing.T, repoDir string) string {
	t.Helper()
	runAPIRefValidatorGit(t, repoDir, "checkout", "-q", "main")
	if err := os.WriteFile(filepath.Join(repoDir, "main.tf"), []byte("resource \"null_resource\" \"changed\" {}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	runAPIRefValidatorGit(t, repoDir, "add", "main.tf")
	runAPIRefValidatorGit(t, repoDir, "commit", "-q", "-m", "advance main")
	runAPIRefValidatorGit(t, repoDir, "push", "-q", "origin", "HEAD:main")
	return strings.TrimSpace(runAPIRefValidatorGit(t, repoDir, "rev-parse", "HEAD"))
}

func createAPIRefValidatorGitBranch(t *testing.T, repoDir string, branch string, commit string) {
	t.Helper()
	runAPIRefValidatorGit(t, repoDir, "checkout", "-q", "-B", branch, commit)
	runAPIRefValidatorGit(t, repoDir, "push", "-q", "-u", "origin", "HEAD:"+branch)
	runAPIRefValidatorGit(t, repoDir, "checkout", "-q", "main")
}

func advanceAPIRefValidatorGitBranch(t *testing.T, repoDir string, branch string) string {
	t.Helper()
	runAPIRefValidatorGit(t, repoDir, "checkout", "-q", branch)
	if err := os.WriteFile(filepath.Join(repoDir, "main.tf"), []byte("resource \"null_resource\" \""+branch+"\" {}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	runAPIRefValidatorGit(t, repoDir, "add", "main.tf")
	runAPIRefValidatorGit(t, repoDir, "commit", "-q", "-m", "advance "+branch)
	runAPIRefValidatorGit(t, repoDir, "push", "-q", "origin", "HEAD:"+branch)
	return strings.TrimSpace(runAPIRefValidatorGit(t, repoDir, "rev-parse", "HEAD"))
}

func runAPIRefValidatorGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	args = append([]string{
		"-c", "protocol.file.allow=always",
		"-c", "gc.auto=0",
		"-c", "maintenance.auto=false",
		"-c", "receive.autogc=false",
	}, args...)
	cmd := exec.Command("git", args...) //nolint:gosec
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null", "GIT_TERMINAL_PROMPT=0")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("running git %s: %s: %v", strings.Join(args, " "), output, err)
	}
	return string(output)
}
