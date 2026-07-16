// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"strings"
	"testing"
	"time"
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

func TestAtlantisStatusContexts(t *testing.T) {
	tests := []struct {
		name    string
		command string
		want    string
	}{
		{name: "plan", command: "plan", want: "atlantis/plan"},
		{name: "apply", command: "apply", want: "atlantis/apply"},
		{name: "unknown command keeps current permissive mapping", command: "unlock", want: "atlantis/unlock"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := atlantisCommandStatusContext(tt.command)
			if got != tt.want {
				t.Fatalf("atlantisCommandStatusContext(%q) = %q, want %q", tt.command, got, tt.want)
			}
		})
	}
}

func TestAtlantisProjectStatusPrefixes(t *testing.T) {
	tests := []struct {
		command string
		want    string
	}{
		{command: "plan", want: "atlantis/plan: "},
		{command: "apply", want: "atlantis/apply: "},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			if got := atlantisProjectStatusPrefix(tt.command); got != tt.want {
				t.Fatalf("atlantisProjectStatusPrefix(%q) = %q, want %q", tt.command, got, tt.want)
			}
		})
	}
}

func TestNewCommentContains(t *testing.T) {
	tests := []struct {
		name     string
		comments []string
		baseline []string
		expected string
		want     bool
	}{
		{
			name:     "new marker is accepted",
			comments: []string{"plan marker", "apply ATLANTIS_E2E_OK"},
			baseline: []string{"plan marker"},
			expected: "ATLANTIS_E2E_OK",
			want:     true,
		},
		{
			name:     "stale marker is rejected",
			comments: []string{"old ATLANTIS_E2E_OK", "new unrelated comment"},
			baseline: []string{"old ATLANTIS_E2E_OK"},
			expected: "ATLANTIS_E2E_OK",
			want:     false,
		},
		{
			name:     "duplicate body added after baseline is new",
			comments: []string{"ATLANTIS_E2E_OK", "ATLANTIS_E2E_OK"},
			baseline: []string{"ATLANTIS_E2E_OK"},
			expected: "ATLANTIS_E2E_OK",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newCommentContains(tt.comments, tt.baseline, tt.expected); got != tt.want {
				t.Fatalf("newCommentContains(%v, %v, %q) = %v, want %v", tt.comments, tt.baseline, tt.expected, got, tt.want)
			}
		})
	}
}

func TestPlanThenApplyCasesAreExplicit(t *testing.T) {
	var got []string
	for _, tc := range testCases {
		if tc.Scenario == ScenarioPlanThenApply {
			got = append(got, tc.Name)
			if tc.ApplyCommand == "" {
				t.Errorf("plan-then-apply case %q has no ApplyCommand", tc.Name)
			}
		}
		if tc.Name == "standalone" && tc.Scenario != ScenarioPlanOnly {
			t.Errorf("legacy standalone case scenario = %d, want plan-only", tc.Scenario)
		}
	}

	want := []string{"builtin-autoplan-apply", "custom-plan-path-apply"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("plan-then-apply cases = %v, want %v", got, want)
	}
}

func TestPlanGenerationRegressionCasesAreExplicit(t *testing.T) {
	var replanCases []string
	var expectedFailureCases []string
	for _, tc := range testCases {
		switch tc.Scenario {
		case ScenarioPlanThenReplanThenApply:
			replanCases = append(replanCases, tc.Name)
			if tc.ReplanMutateFile == "" || tc.ReplanMutateContent == "" || tc.ExpectedReplanCommentSubstring == "" {
				t.Errorf("replan case %q is missing second-generation mutation or marker", tc.Name)
			}
			if tc.ExpectedReplanCommentSubstring == tc.ExpectedApplyCommentSubstring {
				t.Errorf("replan case %q reuses its post-apply marker for the pre-apply assertion", tc.Name)
			}
		case ScenarioPlanThenApplyExpectFailure:
			expectedFailureCases = append(expectedFailureCases, tc.Name)
			if len(tc.ExpectedFailedApplyStatusContexts) == 0 || tc.ExpectedApplyCommentSubstring == "" || tc.ForbiddenApplyCommentSubstring == "" {
				t.Errorf("expected-failure case %q is missing failure or forbidden-marker assertions", tc.Name)
			}
		}
	}

	wantReplan := []string{"builtin-replan-apply", "custom-plan-replan-apply"}
	if strings.Join(replanCases, ",") != strings.Join(wantReplan, ",") {
		t.Fatalf("plan-replan-apply cases = %v, want %v", replanCases, wantReplan)
	}
	wantFailure := []string{"mixed-managed-plan-mutation"}
	if strings.Join(expectedFailureCases, ",") != strings.Join(wantFailure, ",") {
		t.Fatalf("expected-apply-failure cases = %v, want %v", expectedFailureCases, wantFailure)
	}
}

func TestRegressionGenerationContent(t *testing.T) {
	if got, want := regressionGenerationContent("GENERATION_2"), "e2e_generation = \"GENERATION_2\"\n"; got != want {
		t.Fatalf("regressionGenerationContent() = %q, want %q", got, want)
	}
}

func TestOnApplyLockProjectPlanStatusContext(t *testing.T) {
	got := onApplyLockProjectPlanStatusContext()
	want := "atlantis/plan: locking-on-apply-preservation"
	if got != want {
		t.Fatalf("onApplyLockProjectPlanStatusContext() = %q, want %q", got, want)
	}
}

func TestIsNewCommitStatus(t *testing.T) {
	baseTime := time.Date(2026, 6, 29, 1, 2, 3, 0, time.UTC)
	newerTime := baseTime.Add(time.Second)
	olderTime := baseTime.Add(-time.Second)

	tests := []struct {
		name     string
		status   CommitStatus
		baseline CommitStatus
		want     bool
	}{
		{
			name:     "empty status is never new",
			status:   CommitStatus{},
			baseline: CommitStatus{},
			want:     false,
		},
		{
			name:     "no baseline means non-empty status is new",
			status:   CommitStatus{State: "success", ID: 1, UpdatedAt: baseTime},
			baseline: CommitStatus{},
			want:     true,
		},
		{
			name:     "same ID and same timestamp is not new",
			status:   CommitStatus{State: "success", ID: 1, UpdatedAt: baseTime},
			baseline: CommitStatus{State: "success", ID: 1, UpdatedAt: baseTime},
			want:     false,
		},
		{
			name:     "different ID is new",
			status:   CommitStatus{State: "success", ID: 2, UpdatedAt: baseTime},
			baseline: CommitStatus{State: "success", ID: 1, UpdatedAt: baseTime},
			want:     true,
		},
		{
			name:     "same ID but newer timestamp is new",
			status:   CommitStatus{State: "success", ID: 1, UpdatedAt: newerTime},
			baseline: CommitStatus{State: "success", ID: 1, UpdatedAt: baseTime},
			want:     true,
		},
		{
			name:     "older timestamp is not new",
			status:   CommitStatus{State: "success", ID: 1, UpdatedAt: olderTime},
			baseline: CommitStatus{State: "success", ID: 1, UpdatedAt: baseTime},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isNewCommitStatus(tt.status, tt.baseline)
			if got != tt.want {
				t.Fatalf("isNewCommitStatus(%+v, %+v) = %v, want %v", tt.status, tt.baseline, got, tt.want)
			}
		})
	}
}

func TestShouldReturnCommitStatusWaitsForExpectedTerminalState(t *testing.T) {
	success := func(state string) bool { return state == "success" }
	failure := func(state string) bool { return state == "failure" }
	tests := []struct {
		name          string
		state         string
		inProgress    bool
		expectedState func(string) bool
		want          bool
	}{
		{
			name:       "legacy polling returns any terminal status",
			state:      "failure",
			inProgress: false,
			want:       true,
		},
		{
			name:          "transient failure does not satisfy expected success",
			state:         "failure",
			inProgress:    false,
			expectedState: success,
			want:          false,
		},
		{
			name:          "pending never returns",
			state:         "pending",
			inProgress:    true,
			expectedState: success,
			want:          false,
		},
		{
			name:          "success satisfies expected success",
			state:         "success",
			inProgress:    false,
			expectedState: success,
			want:          true,
		},
		{
			name:          "failure satisfies expected failure",
			state:         "failure",
			inProgress:    false,
			expectedState: failure,
			want:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldReturnCommitStatus(tt.state, tt.inProgress, tt.expectedState); got != tt.want {
				t.Fatalf("shouldReturnCommitStatus(%q, %v) = %v, want %v", tt.state, tt.inProgress, got, tt.want)
			}
		})
	}
}

func TestE2ERunNonceIncludesPerRunEntropy(t *testing.T) {
	t.Setenv("GITHUB_RUN_ID", "12345")
	t.Setenv("GITHUB_RUN_ATTEMPT", "2")

	first := e2eRunNonce()
	second := e2eRunNonce()
	if first == second {
		t.Fatalf("e2eRunNonce() returned duplicate values: %q", first)
	}
	if !strings.HasPrefix(first, "12345-2-") || !strings.HasPrefix(second, "12345-2-") {
		t.Fatalf("e2eRunNonce() did not include GitHub run metadata: first=%q second=%q", first, second)
	}
}

func TestNewLifecycleCleanupContextIgnoresParentCancellation(t *testing.T) {
	parent, cancelParent := context.WithCancel(context.Background())
	cancelParent()

	ctx, cleanup := newLifecycleCleanupContext(parent)
	defer cleanup()

	select {
	case <-ctx.Done():
		t.Fatalf("cleanup context was already canceled: %v", ctx.Err())
	default:
	}
}

func TestContainsExactPullRef(t *testing.T) {
	tests := []struct {
		name       string
		comment    string
		pullNumber int
		want       bool
	}{
		{name: "end of string", comment: "#123", pullNumber: 123, want: true},
		{name: "period", comment: "#123.", pullNumber: 123, want: true},
		{name: "comma", comment: "#123,", pullNumber: 123, want: true},
		{name: "space", comment: "#123 ", pullNumber: 123, want: true},
		{name: "newline", comment: "#123\n", pullNumber: 123, want: true},
		{name: "superset number", comment: "#1234", pullNumber: 123, want: false},
		{name: "embedded superset number", comment: "abc#1234", pullNumber: 123, want: false},
		{name: "numeric boundary only permits letters", comment: "#123abc", pullNumber: 123, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsExactPullRef(tt.comment, tt.pullNumber)
			if got != tt.want {
				t.Fatalf("containsExactPullRef(%q, %d) = %v, want %v", tt.comment, tt.pullNumber, got, tt.want)
			}
		})
	}
}

func TestFindLockConflictComment(t *testing.T) {
	tests := []struct {
		name            string
		comments        []string
		ownerPullNumber int
		want            bool
	}{
		{
			name: "repo lock owned by expected PR",
			comments: []string{
				"plan succeeded",
				"This project is currently locked by an unapplied plan from pull #123. To continue, delete the lock from #123 or apply that plan and merge the pull request.",
			},
			ownerPullNumber: 123,
			want:            true,
		},
		{
			name:            "does not match owner pull prefix",
			comments:        []string{"This project is currently locked by an unapplied plan from pull #1234. To continue, delete the lock from #1234 or apply that plan and merge the pull request."},
			ownerPullNumber: 123,
			want:            false,
		},
		{
			name:            "does not match shorter wrong owner",
			comments:        []string{"This project is currently locked by an unapplied plan from pull #123. To continue, delete the lock from #123 or apply that plan and merge the pull request."},
			ownerPullNumber: 1234,
			want:            false,
		},
		{
			name:            "matches exact owner before punctuation",
			comments:        []string{"This project is currently locked by an unapplied plan from pull #123. To continue, delete the lock from #123 or apply that plan and merge the pull request."},
			ownerPullNumber: 123,
			want:            true,
		},
		{
			name:            "normal plan success comment",
			comments:        []string{"Plan succeeded for locking-on-apply-preservation."},
			ownerPullNumber: 123,
			want:            false,
		},
		{
			name:            "normal apply failure comment",
			comments:        []string{"Apply failed: terraform exited with status 1."},
			ownerPullNumber: 123,
			want:            false,
		},
		{
			name:            "working dir lock comment",
			comments:        []string{"The working directory is currently locked."},
			ownerPullNumber: 123,
			want:            false,
		},
		{
			name:            "repo lock owned by another PR",
			comments:        []string{"This project is currently locked by an unapplied plan from pull #999. To continue, delete the lock from #999 or apply that plan and merge the pull request."},
			ownerPullNumber: 123,
			want:            false,
		},
		{
			name:            "generic fragments only",
			comments:        []string{"currently locked delete the lock apply that plan"},
			ownerPullNumber: 123,
			want:            false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got := findLockConflictComment(tt.comments, tt.ownerPullNumber)
			if got != tt.want {
				t.Fatalf("findLockConflictComment(%v, %d) = %v, want %v", tt.comments, tt.ownerPullNumber, got, tt.want)
			}
		})
	}
}
