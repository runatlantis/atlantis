// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

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
	runAPIRefValidatorGit(t, repoDir, "tag", "v1.0.0", initialCommit)
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

func initAPIRefValidatorGitRepo(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	originDir := filepath.Join(root, "origin.git")
	repoDir := filepath.Join(root, "work")
	runAPIRefValidatorGit(t, "", "init", "--bare", originDir)
	runAPIRefValidatorGit(t, "", "init", "--initial-branch=main", repoDir)
	runAPIRefValidatorGit(t, repoDir, "config", "user.email", "atlantisbot@runatlantis.io")
	runAPIRefValidatorGit(t, repoDir, "config", "user.name", "atlantisbot")
	runAPIRefValidatorGit(t, repoDir, "config", "commit.gpgsign", "false")
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
	if err := os.WriteFile(filepath.Join(repoDir, "main.tf"), []byte("resource \"null_resource\" \"changed\" {}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	runAPIRefValidatorGit(t, repoDir, "add", "main.tf")
	runAPIRefValidatorGit(t, repoDir, "commit", "-q", "-m", "advance main")
	runAPIRefValidatorGit(t, repoDir, "push", "-q", "origin", "HEAD:main")
	return strings.TrimSpace(runAPIRefValidatorGit(t, repoDir, "rev-parse", "HEAD"))
}

func runAPIRefValidatorGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...) //nolint:gosec
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("running git %s: %s: %v", strings.Join(args, " "), output, err)
	}
	return string(output)
}
