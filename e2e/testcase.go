// SPDX-License-Identifier: Apache-2.0

package main

// VCSProvider controls which VCS backends a test case runs on.
type VCSProvider int

const (
	VCSBoth   VCSProvider = iota // Run on both GitHub and GitLab
	VCSGitHub                    // GitHub only
	VCSGitLab                    // GitLab only
)

// CaseStatus controls whether a test case is run by default.
type CaseStatus int

const (
	CaseActive   CaseStatus = iota // Run in default E2E
	CaseOptIn                      // Run only with E2E_OPT_IN=1
	CaseDisabled                   // Never run, documented for future
)

// Scenario controls the runner path for complex lifecycle tests.
type Scenario int

const (
	ScenarioPlanOnly Scenario = iota
	ScenarioOnApplyLockPreservation
)

// TestCase defines a single E2E fixture exercise.
type TestCase struct {
	// Name identifies the test case in logs and results.
	Name string

	// Dir is the fixture directory relative to the repo root.
	Dir string

	// Workspace for documentation. Workspace-aware execution is follow-up.
	Workspace string

	// MutateFile is the file path relative to Dir to create/modify.
	// Defaults to "{Name}.tf" if empty.
	MutateFile string

	// MutateContent is written to MutateFile. Defaults to a null_resource.
	MutateContent string

	// ExpectedStatusContexts lists the exact per-project commit status contexts
	// that must appear on GitHub (e.g. "atlantis/plan: project1").
	// Empty means no project-level assertion (only aggregate success).
	// Cases using this field should be VCSGitHub since GitLab does not expose
	// per-project commit statuses.
	ExpectedStatusContexts []string

	// ForbidExtraProjectStatuses, when true, fails if any "atlantis/plan: *"
	// status appears that is not in ExpectedStatusContexts.
	ForbidExtraProjectStatuses bool

	// ExpectFailure means we expect the plan to fail (status = failure/error).
	ExpectFailure bool

	// ExpectedCommentSubstring, if non-empty, requires this string to appear
	// in at least one Atlantis PR/MR comment after plan succeeds.
	ExpectedCommentSubstring string

	// ApplyCommand to post after plan succeeds. Empty = skip apply.
	// Reserved; apply execution is not yet implemented in the harness.
	ApplyCommand string

	// Scenario selects a specialized runner path. Zero keeps the default plan-only behavior.
	Scenario Scenario

	// VCS controls which providers run this case.
	VCS VCSProvider

	// Status controls whether this case is active by default.
	Status CaseStatus

	// SkipReason documents why a case is opt-in or disabled.
	SkipReason string
}

// defaultMutateContent is a minimal Terraform resource for triggering a plan.
const defaultMutateContent = `resource "null_resource" "e2e" {
  triggers = {
    run = timestamp()
  }
}
`

// projectStatusPrefix is the prefix for per-project Atlantis statuses.
// The aggregate status is "atlantis/plan" (no colon/space).
const projectStatusPrefix = "atlantis/plan: "

func atlantisCommandStatusContext(command string) string {
	return "atlantis/" + command
}

// testCases is the fixture test matrix.
// Cases with Status=CaseActive run in every E2E invocation.
var testCases = []TestCase{
	// ─── Original smoke-test fixtures (must not regress) ───
	// These use per-project context assertion on GitHub to confirm the
	// named project was actually planned (not just aggregate success).
	{
		Name:                       "standalone",
		Dir:                        "standalone",
		MutateFile:                 "e2e_trigger.tf",
		ExpectedStatusContexts:     []string{"atlantis/plan: standalone"},
		ForbidExtraProjectStatuses: true,
		ApplyCommand:               "atlantis apply -d standalone",
		VCS:                        VCSGitHub,
		Status:                     CaseActive,
	},
	{
		Name:                       "standalone-with-workspace",
		Dir:                        "standalone-with-workspace",
		Workspace:                  "staging",
		MutateFile:                 "e2e_trigger.tf",
		ExpectedStatusContexts:     []string{"atlantis/plan: standalone-with-workspace"},
		ForbidExtraProjectStatuses: true,
		ApplyCommand:               "atlantis apply -d standalone-with-workspace -w staging",
		VCS:                        VCSGitHub,
		Status:                     CaseActive,
	},
	// GitLab smoke: asserts aggregate pipeline success plus plan comment marker.
	// Cannot assert per-project status contexts on GitLab, but verifying the
	// Atlantis MR note contains the project header prevents 0/0 false positives.
	{
		Name:                     "standalone-gitlab",
		Dir:                      "standalone",
		MutateFile:               "e2e_trigger.tf",
		ExpectedCommentSubstring: "dir: `standalone` workspace: `default`",
		VCS:                      VCSGitLab,
		Status:                   CaseActive,
	},

	// ─── Multi-project: single project change ───
	// GitHub-only. Asserts exactly project1 planned; forbids extra project statuses.
	{
		Name:                       "multi-project-single",
		Dir:                        "multi-projects/project1",
		MutateFile:                 "e2e_trigger.tf",
		ExpectedStatusContexts:     []string{"atlantis/plan: project1"},
		ForbidExtraProjectStatuses: true,
		VCS:                        VCSGitHub,
		Status:                     CaseActive,
	},

	// ─── Multi-project: shared-module fan-out ───
	// GitHub-only. Asserts exactly project1 + project2 via when_modified.
	// Fails if shared-module is autodiscovered as a third project.
	{
		Name:                       "multi-project-fanout",
		Dir:                        "multi-projects/shared-module",
		MutateFile:                 "e2e_trigger.tf",
		ExpectedStatusContexts:     []string{"atlantis/plan: project1", "atlantis/plan: project2"},
		ForbidExtraProjectStatuses: true,
		VCS:                        VCSGitHub,
		Status:                     CaseActive,
	},

	// ─── Configured .tf.json project planning ───
	// Proves Atlantis can plan a configured project containing only .tf.json files.
	// Note: this tests configured JSON project planning, not autodiscovery/detection
	// of unconfigured .tf.json directories (the project is explicit in atlantis.yaml).
	{
		Name:                   "configured-terraform-json",
		Dir:                    "detection/terraform-json",
		MutateFile:             "extra.tf.json",
		ExpectedStatusContexts: []string{"atlantis/plan: detection-json"},
		MutateContent: `{
  "resource": {
    "null_resource": {
      "json_e2e": {
        "triggers": {
          "run": "e2e"
        }
      }
    }
  }
}
`,
		VCS:    VCSGitHub,
		Status: CaseActive,
	},

	// ─── Autodiscovery: included project ───
	// GitHub-only. Proves autodiscovery selected included-a.
	// ProjectID for unnamed autodiscovered projects is "dir/workspace".
	{
		Name:                   "autodiscovery-included",
		Dir:                    "autodiscovery/included-a",
		MutateFile:             "e2e_trigger.tf",
		ExpectedStatusContexts: []string{"atlantis/plan: autodiscovery/included-a/default"},
		VCS:                    VCSGitHub,
		Status:                 CaseActive,
	},

	// ─── Custom workflow: PROJECT_NAME env ───
	// Asserts both: project status context proves the project ran, and
	// comment marker proves the custom workflow script executed.
	{
		Name:                     "custom-workflow-project-name",
		Dir:                      "custom-workflows/project-name-env",
		MutateFile:               "e2e_trigger.tf",
		ExpectedStatusContexts:   []string{"atlantis/plan: custom-workflow-env"},
		ExpectedCommentSubstring: "PASS: PROJECT_NAME=custom-workflow-env",
		VCS:                      VCSGitHub,
		Status:                   CaseActive,
	},

	// ─── Output: long-line rendering ───
	// Writes a separate trigger file to preserve the >64KiB long-line expression.
	// Asserts project status context proves the project was planned.
	// TODO: Add ExpectedCommentSubstring for a post-long-line marker once
	// atlantis-tests fixture includes one (requires fixture-side follow-up).
	{
		Name:                   "output-long-line",
		Dir:                    "output/long-line",
		MutateFile:             "e2e_trigger.tf",
		ExpectedStatusContexts: []string{"atlantis/plan: output-long-line"},
		VCS:                    VCSGitHub,
		Status:                 CaseActive,
	},

	// ─── Disabled/opt-in fixtures with reasons ───
	{
		Name:       "output-failure",
		Dir:        "output/failure",
		MutateFile: "e2e_trigger.tf",
		Status:     CaseDisabled,
		SkipReason: "Intentional failure fixture; autoplan disabled in atlantis.yaml. Needs manual trigger support.",
	},
	{
		Name:       "detection-opentofu",
		Dir:        "detection/opentofu-basic",
		MutateFile: "e2e_trigger.tf",
		Status:     CaseDisabled,
		SkipReason: "Requires tofu binary in E2E environment. Not guaranteed in CI.",
	},
	{
		Name:       "drift-local-file",
		Dir:        "drift/local-file",
		MutateFile: "e2e_trigger.tf",
		Status:     CaseDisabled,
		SkipReason: "Requires --enable-drift-detection server flag. Alpha feature.",
	},
	{
		Name:       "multi-project-workspace",
		Dir:        "multi-projects/project-with-workspace",
		Workspace:  "dev",
		MutateFile: "e2e_trigger.tf",
		Status:     CaseOptIn,
		SkipReason: "Workspace fixture works but adds CI time. Enable with E2E_OPT_IN=1.",
	},
	{
		Name:       "autodiscovery-explicit",
		Dir:        "autodiscovery/explicit",
		MutateFile: "e2e_trigger.tf",
		Status:     CaseOptIn,
		SkipReason: "Validates explicit-over-autodiscovery precedence. Enable with E2E_OPT_IN=1.",
	},
	{
		Name:                   "locking-on-apply-preservation",
		Dir:                    "locking/on-apply-lock-preservation",
		MutateFile:             "e2e_pr1_trigger.tf",
		ExpectedStatusContexts: []string{"atlantis/plan: locking-on-apply-preservation"},
		VCS:                    VCSGitHub,
		Status:                 CaseOptIn,
		SkipReason:             "Exercises apply plus two-PR repo-lock lifecycle; enable with E2E_OPT_IN=1.",
		Scenario:               ScenarioOnApplyLockPreservation,
	},
}
