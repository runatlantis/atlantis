// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

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

// TestCase defines a single E2E fixture exercise.
type TestCase struct {
	// Name identifies the test case in logs and results.
	Name string

	// Dir is the fixture directory relative to the repo root (e.g. "standalone").
	Dir string

	// Workspace to use. Empty means default workspace.
	Workspace string

	// MutateFile is the file path relative to Dir to create/modify.
	// Defaults to "{Name}.tf" if empty.
	MutateFile string

	// MutateContent is written to MutateFile. Defaults to a null_resource.
	MutateContent string

	// StatusPrefix is the commit status context prefix to poll.
	// Defaults to "atlantis/plan". For multi-project cases expecting
	// per-project statuses, use "atlantis/plan" to match all plan statuses.
	StatusPrefix string

	// ExpectedStatusCount is how many matching statuses to expect.
	// 0 or 1 means "at least one matching status must succeed."
	// >1 means "exactly N matching statuses must all succeed."
	ExpectedStatusCount int

	// ExpectFailure means we expect the plan to fail (status = failure/error).
	ExpectFailure bool

	// ApplyCommand to post after plan succeeds. Empty = skip apply.
	ApplyCommand string

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

// testCases is the fixture test matrix.
// Cases with Status=CaseActive run in every E2E invocation.
var testCases = []TestCase{
	// ─── Original smoke-test fixtures (must not regress) ───
	{
		Name:         "standalone",
		Dir:          "standalone",
		ApplyCommand: "atlantis apply -d standalone",
		VCS:          VCSBoth,
		Status:       CaseActive,
	},
	{
		Name:         "standalone-with-workspace",
		Dir:          "standalone-with-workspace",
		Workspace:    "staging",
		ApplyCommand: "atlantis apply -d standalone-with-workspace -w staging",
		VCS:          VCSBoth,
		Status:       CaseActive,
	},

	// ─── Multi-project: single project change ───
	{
		Name:       "multi-project-single",
		Dir:        "multi-projects/project1",
		MutateFile: "main.tf",
		VCS:        VCSBoth,
		Status:     CaseActive,
	},

	// ─── Multi-project: shared-module fan-out ───
	{
		Name:                "multi-project-fanout",
		Dir:                 "multi-projects/shared-module",
		MutateFile:          "main.tf",
		StatusPrefix:        "atlantis/plan",
		ExpectedStatusCount: 2,
		VCS:                 VCSBoth,
		Status:              CaseActive,
	},

	// ─── Detection: .tf.json format ───
	{
		Name:          "detection-terraform-json",
		Dir:           "detection/terraform-json",
		MutateFile:    "extra.tf",
		MutateContent: "resource \"null_resource\" \"json_test\" {}\n",
		VCS:           VCSBoth,
		Status:        CaseActive,
	},

	// ─── Autodiscovery: included project ───
	{
		Name:       "autodiscovery-included",
		Dir:        "autodiscovery/included-a",
		MutateFile: "main.tf",
		VCS:        VCSBoth,
		Status:     CaseActive,
	},

	// ─── Custom workflow: PROJECT_NAME env ───
	{
		Name:       "custom-workflow-project-name",
		Dir:        "custom-workflows/project-name-env",
		MutateFile: "main.tf",
		VCS:        VCSBoth,
		Status:     CaseActive,
	},

	// ─── Output: long-line rendering ───
	{
		Name:       "output-long-line",
		Dir:        "output/long-line",
		MutateFile: "main.tf",
		VCS:        VCSBoth,
		Status:     CaseActive,
	},

	// ─── Disabled/opt-in fixtures with reasons ───
	{
		Name:       "output-failure",
		Dir:        "output/failure",
		MutateFile: "main.tf",
		Status:     CaseDisabled,
		SkipReason: "Intentional failure fixture; autoplan disabled in atlantis.yaml. Needs manual trigger support.",
	},
	{
		Name:       "detection-opentofu",
		Dir:        "detection/opentofu-basic",
		MutateFile: "main.tf",
		Status:     CaseDisabled,
		SkipReason: "Requires tofu binary in E2E environment. Not guaranteed in CI.",
	},
	{
		Name:       "drift-local-file",
		Dir:        "drift/local-file",
		MutateFile: "main.tf",
		Status:     CaseDisabled,
		SkipReason: "Requires --enable-drift-detection server flag. Alpha feature.",
	},
	{
		Name:       "multi-project-workspace",
		Dir:        "multi-projects/project-with-workspace",
		Workspace:  "dev",
		MutateFile: "main.tf",
		Status:     CaseOptIn,
		SkipReason: "Workspace fixture works but adds CI time. Enable with E2E_OPT_IN=1.",
	},
	{
		Name:       "autodiscovery-explicit",
		Dir:        "autodiscovery/explicit",
		MutateFile: "main.tf",
		Status:     CaseOptIn,
		SkipReason: "Validates explicit-over-autodiscovery precedence. Enable with E2E_OPT_IN=1.",
	},
}
