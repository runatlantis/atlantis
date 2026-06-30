// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

// ValidateNonPRAPIRefUnchanged fails closed if a synthetic non-PR API apply was
// planned against a mutable branch ref that has advanced since setup.
func ValidateNonPRAPIRefUnchanged(ctx command.ProjectContext, repoDir string) error {
	if !ctx.API || ctx.Pull.Num > 0 || repoDir == "" {
		return nil
	}
	if _, err := os.Stat(filepath.Join(repoDir, ".git")); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("checking API checkout git metadata: %w", err)
	}
	headRef := strings.TrimSpace(ctx.Pull.HeadBranch)
	if headRef == "" || isImmutableAPICommitRef(headRef) {
		return nil
	}
	if !strings.HasPrefix(headRef, "refs/tags/") && models.RequiresBaseBranchForRef(headRef) {
		return nil
	}
	resolved, err := resolveMutableAPIRef(repoDir, headRef)
	if err != nil {
		return err
	}
	if ctx.Pull.HeadCommit != "" && resolved != ctx.Pull.HeadCommit {
		return fmt.Errorf("API ref %q changed from %s to %s while apply was running; rerun plan/apply for the current ref", headRef, shortAPICommit(ctx.Pull.HeadCommit), shortAPICommit(resolved))
	}
	return nil
}

func resolveMutableAPIRef(repoDir, ref string) (string, error) {
	ref = strings.TrimSpace(ref)
	if strings.HasPrefix(ref, "refs/tags/") {
		if models.IsUnsafeAPIRef(ref) {
			return "", fmt.Errorf("invalid API ref %q", ref)
		}
		remoteRef := ref
		fetchRef := fmt.Sprintf("+%s:%s", ref, remoteRef)
		cmd := exec.Command("git", "fetch", "origin", "--", fetchRef) //nolint:gosec // fetchRef is built from a refs/tags ref checked for unsafe forms.
		cmd.Dir = repoDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("resolving API ref %q: %s: %w", ref, strings.TrimSpace(string(output)), err)
		}
		return checkedOutCommit(repoDir, remoteRef)
	}

	branchRef, ok := models.CheckedBaseBranchRef(ref)
	if !ok {
		return "", fmt.Errorf("invalid API ref %q", ref)
	}
	remoteRef := "refs/remotes/origin/" + branchRef
	fetchRef := fmt.Sprintf("+refs/heads/%s:%s", branchRef, remoteRef)
	cmd := exec.Command("git", "fetch", "origin", "--", fetchRef) //nolint:gosec // fetchRef is built from a validated branch ref.
	cmd.Dir = repoDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("resolving API ref %q: %s: %w", ref, strings.TrimSpace(string(output)), err)
	}
	return checkedOutCommit(repoDir, remoteRef)
}

func checkedOutCommit(repoDir string, ref string) (string, error) {
	cmd := exec.Command("git", "rev-parse", ref) //nolint:gosec // ref is either a local git ref built by Atlantis or a constant.
	cmd.Dir = repoDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("resolving git ref %q: %s: %w", ref, strings.TrimSpace(string(output)), err)
	}
	commit := strings.TrimSpace(string(output))
	if commit == "" {
		return "", fmt.Errorf("resolving git ref %q: empty commit", ref)
	}
	return commit, nil
}

func isImmutableAPICommitRef(ref string) bool {
	ref = strings.TrimSpace(ref)
	if len(ref) < 7 || len(ref) > 40 {
		return false
	}
	for _, r := range ref {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') {
			continue
		}
		return false
	}
	return true
}

func shortAPICommit(commit string) string {
	if len(commit) > 12 {
		return commit[:12]
	}
	return commit
}
