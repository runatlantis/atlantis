// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package command_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

func TestProjectResult_IsSuccessful(t *testing.T) {
	cases := map[string]struct {
		pr  command.ProjectResult
		exp bool
	}{
		"plan success": {
			command.ProjectResult{
				ProjectCommandOutput: command.ProjectCommandOutput{
					PlanSuccess: &models.PlanSuccess{},
				},
			},
			true,
		},
		"policy_check success": {
			command.ProjectResult{
				ProjectCommandOutput: command.ProjectCommandOutput{
					PolicyCheckResults: &models.PolicyCheckResults{},
				},
			},
			true,
		},
		"apply success": {
			command.ProjectResult{
				ProjectCommandOutput: command.ProjectCommandOutput{
					ApplySuccess: "success",
				},
			},
			true,
		},
		"failure": {
			command.ProjectResult{
				ProjectCommandOutput: command.ProjectCommandOutput{
					Failure: "failure",
				},
			},
			false,
		},
		"error": {
			command.ProjectResult{
				ProjectCommandOutput: command.ProjectCommandOutput{
					Error: errors.New("error"),
				},
			},
			false,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			Equals(t, c.exp, c.pr.IsSuccessful())
		})
	}
}

func TestProjectResult_PlanStatus(t *testing.T) {
	cases := []struct {
		p         command.ProjectResult
		expStatus models.ProjectPlanStatus
	}{
		{
			p: command.ProjectResult{
				Command: command.Plan,
				ProjectCommandOutput: command.ProjectCommandOutput{
					Error: errors.New("err"),
				},
			},
			expStatus: models.ErroredPlanStatus,
		},
		{
			p: command.ProjectResult{
				Command: command.Plan,
				ProjectCommandOutput: command.ProjectCommandOutput{
					Failure: "failure",
				},
			},
			expStatus: models.ErroredPlanStatus,
		},
		{
			p: command.ProjectResult{
				Command: command.Plan,
				ProjectCommandOutput: command.ProjectCommandOutput{
					PlanSuccess: &models.PlanSuccess{},
				},
			},
			expStatus: models.PlannedPlanStatus,
		},
		{
			p: command.ProjectResult{
				Command: command.Plan,
				ProjectCommandOutput: command.ProjectCommandOutput{
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "No changes. Infrastructure is up-to-date.",
					},
				},
			},
			expStatus: models.PlannedNoChangesPlanStatus,
		},
		{
			p: command.ProjectResult{
				Command: command.Apply,
				ProjectCommandOutput: command.ProjectCommandOutput{
					Error: errors.New("err"),
				},
			},
			expStatus: models.ErroredApplyStatus,
		},
		{
			p: command.ProjectResult{
				Command: command.Apply,
				ProjectCommandOutput: command.ProjectCommandOutput{
					Failure: "failure",
				},
			},
			expStatus: models.ErroredApplyStatus,
		},
		{
			p: command.ProjectResult{
				Command: command.Apply,
				ProjectCommandOutput: command.ProjectCommandOutput{
					ApplySuccess: "success",
				},
			},
			expStatus: models.AppliedPlanStatus,
		},
		{
			p: command.ProjectResult{
				Command: command.PolicyCheck,
				ProjectCommandOutput: command.ProjectCommandOutput{
					PolicyCheckResults: &models.PolicyCheckResults{},
				},
			},
			expStatus: models.PassedPolicyCheckStatus,
		},
		{
			p: command.ProjectResult{
				Command: command.PolicyCheck,
				ProjectCommandOutput: command.ProjectCommandOutput{
					Failure: "failure",
				},
			},
			expStatus: models.ErroredPolicyCheckStatus,
		},
		{
			p: command.ProjectResult{
				Command: command.ApprovePolicies,
				ProjectCommandOutput: command.ProjectCommandOutput{
					PolicyCheckResults: &models.PolicyCheckResults{},
				},
			},
			expStatus: models.PassedPolicyCheckStatus,
		},
		{
			p: command.ProjectResult{
				Command: command.ApprovePolicies,
				ProjectCommandOutput: command.ProjectCommandOutput{
					Failure: "failure",
				},
			},
			expStatus: models.ErroredPolicyCheckStatus,
		},
	}

	for _, c := range cases {
		t.Run(c.expStatus.String(), func(t *testing.T) {
			Equals(t, c.expStatus, c.p.PlanStatus())
		})
	}
}

func TestPlanSuccess_Summary(t *testing.T) {
	cases := []struct {
		p         command.ProjectResult
		expResult string
	}{
		{
			p: command.ProjectResult{
				ProjectCommandOutput: command.ProjectCommandOutput{
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: `
					An execution plan has been generated and is shown below.
					Resource actions are indicated with the following symbols:
					  - destroy

					Terraform will perform the following actions:

					  - null_resource.hi[1]


					Plan: 0 to add, 0 to change, 1 to destroy.`,
					},
				},
			},
			expResult: "Plan: 0 to add, 0 to change, 1 to destroy.",
		},
		{
			p: command.ProjectResult{
				ProjectCommandOutput: command.ProjectCommandOutput{
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: `
					An execution plan has been generated and is shown below.
					Resource actions are indicated with the following symbols:

					No changes. Infrastructure is up-to-date.`,
					},
				},
			},
			expResult: "No changes. Infrastructure is up-to-date.",
		},
		{
			p: command.ProjectResult{
				ProjectCommandOutput: command.ProjectCommandOutput{
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: `
					Note: Objects have changed outside of Terraform

					Terraform detected the following changes made outside of Terraform since the
					last "terraform apply":

					No changes. Your infrastructure matches the configuration.`,
					},
				},
			},
			expResult: "\n**Note: Objects have changed outside of Terraform**\nNo changes. Your infrastructure matches the configuration.",
		},
		{
			p: command.ProjectResult{
				ProjectCommandOutput: command.ProjectCommandOutput{
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: `
					Note: Objects have changed outside of Terraform

					Terraform detected the following changes made outside of Terraform since the
					last "terraform apply":

					An execution plan has been generated and is shown below.
					Resource actions are indicated with the following symbols:
					  - destroy

					Terraform will perform the following actions:

					  - null_resource.hi[1]


					Plan: 0 to add, 0 to change, 1 to destroy.`,
					},
				},
			},
			expResult: "\n**Note: Objects have changed outside of Terraform**\nPlan: 0 to add, 0 to change, 1 to destroy.",
		},
		{
			p: command.ProjectResult{
				ProjectCommandOutput: command.ProjectCommandOutput{
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: `No match, expect empty`,
					},
				},
			},
			expResult: "",
		},
	}

	for _, c := range cases {
		t.Run(c.expResult, func(t *testing.T) {
			Equals(t, c.expResult, c.p.PlanSuccess.Summary())
		})
	}
}

// TestProjectResult_MarshalJSON verifies that ProjectResult serializes errors properly
// and maintains backwards-compatible flat JSON structure (no ProjectCommandOutput wrapper).
func TestProjectResult_MarshalJSON(t *testing.T) {
	cases := map[string]struct {
		pr          command.ProjectResult
		checkFields map[string]any // Fields to verify in JSON output
	}{
		"nil error serializes as null": {
			pr: command.ProjectResult{
				ProjectCommandOutput: command.ProjectCommandOutput{
					PlanSuccess: &models.PlanSuccess{
						TerraformOutput: "Plan: 1 to add",
					},
				},
				Command:    command.Plan,
				RepoRelDir: ".",
				Workspace:  "default",
			},
			checkFields: map[string]any{
				"Error":      nil,
				"Failure":    "",
				"RepoRelDir": ".",
				"Workspace":  "default",
			},
		},
		"error serializes as string not empty object": {
			pr: command.ProjectResult{
				ProjectCommandOutput: command.ProjectCommandOutput{
					Error:   errors.New("terraform plan failed: resource not found"),
					Failure: "plan execution error",
				},
				Command:    command.Plan,
				RepoRelDir: "modules/vpc",
				Workspace:  "production",
			},
			checkFields: map[string]any{
				"Error":      "terraform plan failed: resource not found",
				"Failure":    "plan execution error",
				"RepoRelDir": "modules/vpc",
				"Workspace":  "production",
			},
		},
		"all fields present in flat structure": {
			pr: command.ProjectResult{
				ProjectCommandOutput: command.ProjectCommandOutput{
					ApplySuccess: "Apply complete!",
				},
				Command:           command.Apply,
				SubCommand:        "",
				RepoRelDir:        ".",
				Workspace:         "default",
				ProjectName:       "my-project",
				SilencePRComments: []string{"comment1"},
			},
			checkFields: map[string]any{
				"ApplySuccess": "Apply complete!",
				"Command":      float64(command.Apply), // JSON numbers are float64
				"RepoRelDir":   ".",
				"Workspace":    "default",
				"ProjectName":  "my-project",
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			jsonBytes, err := json.Marshal(tc.pr)
			Ok(t, err)

			// Parse into generic map to check structure
			var parsed map[string]any
			err = json.Unmarshal(jsonBytes, &parsed)
			Ok(t, err)

			// Verify expected fields
			for field, expected := range tc.checkFields {
				actual, exists := parsed[field]
				Assert(t, exists, "expected field %q to exist in JSON output", field)
				Equals(t, expected, actual)
			}

			// CRITICAL: Verify NO ProjectCommandOutput wrapper exists
			// This is the key backwards-compatibility check - embedded structs
			// should produce flat fields, not nested objects
			_, hasWrapper := parsed["ProjectCommandOutput"]
			Assert(t, !hasWrapper, "JSON should NOT have ProjectCommandOutput wrapper - must maintain flat structure")
		})
	}
}

// TestProjectResult_MarshalJSON_FlatStructure is a specific test ensuring that
// the JSON output maintains backwards compatibility by having a flat structure
// rather than nesting fields under ProjectCommandOutput.
func TestProjectResult_MarshalJSON_FlatStructure(t *testing.T) {
	pr := command.ProjectResult{
		ProjectCommandOutput: command.ProjectCommandOutput{
			Error:   errors.New("test error"),
			Failure: "test failure",
			PlanSuccess: &models.PlanSuccess{
				TerraformOutput: "output",
				LockURL:         "http://lock",
				RePlanCmd:       "atlantis plan",
				ApplyCmd:        "atlantis apply",
			},
		},
		Command:     command.Plan,
		RepoRelDir:  ".",
		Workspace:   "default",
		ProjectName: "test-project",
	}

	jsonBytes, err := json.Marshal(pr)
	Ok(t, err)

	var parsed map[string]any
	err = json.Unmarshal(jsonBytes, &parsed)
	Ok(t, err)

	// All these fields should be at the top level, not nested
	topLevelFields := []string{
		"Error", "Failure", "PlanSuccess", "PolicyCheckResults",
		"ApplySuccess", "VersionSuccess", "ImportSuccess", "StateRmSuccess",
		"Command", "SubCommand", "RepoRelDir", "Workspace", "ProjectName", "SilencePRComments",
	}

	for _, field := range topLevelFields {
		_, exists := parsed[field]
		Assert(t, exists, "field %q should exist at top level of JSON", field)
	}

	// Verify error is string, not empty object
	errorVal := parsed["Error"]
	Assert(t, errorVal != nil, "Error should not be nil")
	errorStr, ok := errorVal.(string)
	Assert(t, ok, "Error should be string, got %T: %v", errorVal, errorVal)
	Equals(t, "test error", errorStr)

	// Verify PlanSuccess is an object with expected fields
	planSuccess, ok := parsed["PlanSuccess"].(map[string]any)
	Assert(t, ok, "PlanSuccess should be an object")
	Equals(t, "output", planSuccess["TerraformOutput"])
}

var Summary string

func BenchmarkPlanSuccess_Summary(b *testing.B) {
	var s string

	fixtures := map[string]string{
		"changes": `
					An execution plan has been generated and is shown below.
					Resource actions are indicated with the following symbols:
					  - destroy

					Terraform will perform the following actions:

					  - null_resource.hi[1]


					Plan: 0 to add, 0 to change, 1 to destroy.`,
		"no changes": `
					An execution plan has been generated and is shown below.
					Resource actions are indicated with the following symbols:

					No changes. Infrastructure is up-to-date.`,
		"changes outside Terraform": `
					Note: Objects have changed outside of Terraform

					Terraform detected the following changes made outside of Terraform since the
					last "terraform apply":

					No changes. Your infrastructure matches the configuration.`,
		"changes and changes outside": `
					Note: Objects have changed outside of Terraform

					Terraform detected the following changes made outside of Terraform since the
					last "terraform apply":

					An execution plan has been generated and is shown below.
					Resource actions are indicated with the following symbols:
					  - destroy

					Terraform will perform the following actions:

					  - null_resource.hi[1]


					Plan: 0 to add, 0 to change, 1 to destroy.`,
		"empty summary, no matches": `No match, expect empty`,
	}

	for name, output := range fixtures {
		p := &models.PlanSuccess{
			TerraformOutput: output,
		}

		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				s = p.Summary()
			}

			Summary = s
		})
	}
}
