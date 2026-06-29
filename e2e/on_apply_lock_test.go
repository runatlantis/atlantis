// SPDX-License-Identifier: Apache-2.0

package main

import (
	"strings"
	"testing"
)

const sampleAtlantisYAML = `version: 3
projects:
  - dir: standalone
    name: standalone
    workspace: default

  # --- Locking lifecycle fixtures ---
  - dir: locking/on-apply-lock-preservation
    name: locking-on-apply-preservation
    workspace: default
    # NOTE: The future runner scenario must run this project with
    # repo_locks.mode: on_apply.

workflows:
  default:
    plan:
      steps:
        - init
        - plan
`

func TestEnableOnApplyRepoLocksForFixtureContent(t *testing.T) {
	got, err := enableOnApplyRepoLocksForFixtureContent(sampleAtlantisYAML)
	if err != nil {
		t.Fatalf("enableOnApplyRepoLocksForFixtureContent() error = %v", err)
	}

	wantBlock := `  - dir: locking/on-apply-lock-preservation
    name: locking-on-apply-preservation
    workspace: default
    repo_locks:
      mode: on_apply
    # NOTE:`
	if !strings.Contains(got, wantBlock) {
		t.Fatalf("patched YAML missing repo_locks block under fixture project:\n%s", got)
	}
	if !strings.Contains(got, "  - dir: standalone\n    name: standalone\n    workspace: default") {
		t.Fatalf("patched YAML changed unrelated project:\n%s", got)
	}
}

func TestEnableOnApplyRepoLocksForFixtureContentMissingProject(t *testing.T) {
	_, err := enableOnApplyRepoLocksForFixtureContent(`version: 3
projects:
  - dir: standalone
    name: standalone
`)
	if err == nil {
		t.Fatal("enableOnApplyRepoLocksForFixtureContent() error = nil, want missing project error")
	}
}

func TestEnableOnApplyRepoLocksForFixtureContentAlreadyOnApply(t *testing.T) {
	input := strings.Replace(sampleAtlantisYAML, "    name: locking-on-apply-preservation\n    workspace: default\n    # NOTE:", "    name: locking-on-apply-preservation\n    workspace: default\n    repo_locks:\n      mode: on_apply\n    # NOTE:", 1)
	got, err := enableOnApplyRepoLocksForFixtureContent(input)
	if err != nil {
		t.Fatalf("enableOnApplyRepoLocksForFixtureContent() error = %v", err)
	}
	if got != input {
		t.Fatalf("already-active repo_locks should be deterministic and unchanged\nwant:\n%s\ngot:\n%s", input, got)
	}
}

func TestEnableOnApplyRepoLocksForFixtureContentUnexpectedRepoLocks(t *testing.T) {
	input := strings.Replace(sampleAtlantisYAML, "    name: locking-on-apply-preservation\n    workspace: default\n    # NOTE:", "    name: locking-on-apply-preservation\n    workspace: default\n    repo_locks:\n      mode: on_plan\n    # NOTE:", 1)
	_, err := enableOnApplyRepoLocksForFixtureContent(input)
	if err == nil {
		t.Fatal("enableOnApplyRepoLocksForFixtureContent() error = nil, want unexpected repo_locks error")
	}
}

func TestFindLockConflictComment(t *testing.T) {
	comment, ok := findLockConflictComment([]string{
		"plan succeeded",
		"This project is currently locked by pull request #1.",
	})
	if !ok {
		t.Fatal("findLockConflictComment() ok = false, want true")
	}
	if !strings.Contains(comment, "currently locked") {
		t.Fatalf("findLockConflictComment() comment = %q", comment)
	}
}
