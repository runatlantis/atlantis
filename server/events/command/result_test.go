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

func TestCommandResult_HasErrors(t *testing.T) {
	cases := map[string]struct {
		cr  command.Result
		exp bool
	}{
		"error": {
			cr: command.Result{
				Error: errors.New("err"),
			},
			exp: true,
		},
		"failure": {
			cr: command.Result{
				Failure: "failure",
			},
			exp: true,
		},
		"empty results list": {
			cr: command.Result{
				ProjectResults: []command.ProjectResult{},
			},
			exp: false,
		},
		"successful plan": {
			cr: command.Result{
				ProjectResults: []command.ProjectResult{
					{
						ProjectCommandOutput: command.ProjectCommandOutput{
							PlanSuccess: &models.PlanSuccess{},
						},
					},
				},
			},
			exp: false,
		},
		"successful apply": {
			cr: command.Result{
				ProjectResults: []command.ProjectResult{
					{
						ProjectCommandOutput: command.ProjectCommandOutput{
							ApplySuccess: "success",
						},
					},
				},
			},
			exp: false,
		},
		"single errored project": {
			cr: command.Result{
				ProjectResults: []command.ProjectResult{
					{
						ProjectCommandOutput: command.ProjectCommandOutput{
							Error: errors.New("err"),
						},
					},
				},
			},
			exp: true,
		},
		"single failed project": {
			cr: command.Result{
				ProjectResults: []command.ProjectResult{
					{
						ProjectCommandOutput: command.ProjectCommandOutput{
							Failure: "failure",
						},
					},
				},
			},
			exp: true,
		},
		"two successful projects": {
			cr: command.Result{
				ProjectResults: []command.ProjectResult{
					{
						ProjectCommandOutput: command.ProjectCommandOutput{
							PlanSuccess: &models.PlanSuccess{},
						},
					},
					{
						ProjectCommandOutput: command.ProjectCommandOutput{
							ApplySuccess: "success",
						},
					},
				},
			},
			exp: false,
		},
		"one successful, one failed project": {
			cr: command.Result{
				ProjectResults: []command.ProjectResult{
					{
						ProjectCommandOutput: command.ProjectCommandOutput{
							PlanSuccess: &models.PlanSuccess{},
						},
					},
					{
						ProjectCommandOutput: command.ProjectCommandOutput{
							Failure: "failed",
						},
					},
				},
			},
			exp: true,
		},
	}

	for descrip, c := range cases {
		t.Run(descrip, func(t *testing.T) {
			Equals(t, c.exp, c.cr.HasErrors())
		})
	}
}

// TestResult_MarshalJSON verifies that Result serializes errors properly
// and maintains backwards-compatible JSON structure.
func TestResult_MarshalJSON(t *testing.T) {
	cases := map[string]struct {
		result      command.Result
		checkFields map[string]any // Fields to verify in JSON output
	}{
		"nil error serializes as null": {
			result: command.Result{
				Failure:        "",
				ProjectResults: []command.ProjectResult{},
				PlansDeleted:   false,
			},
			checkFields: map[string]any{
				"Error":        nil,
				"Failure":      "",
				"PlansDeleted": false,
			},
		},
		"error serializes as string": {
			result: command.Result{
				Error:          errors.New("something went wrong"),
				Failure:        "deployment failed",
				ProjectResults: []command.ProjectResult{},
				PlansDeleted:   true,
			},
			checkFields: map[string]any{
				"Error":        "something went wrong",
				"Failure":      "deployment failed",
				"PlansDeleted": true,
			},
		},
		"nested project errors serialize correctly": {
			result: command.Result{
				Error: nil,
				ProjectResults: []command.ProjectResult{
					{
						ProjectCommandOutput: command.ProjectCommandOutput{
							Error:   errors.New("plan failed"),
							Failure: "",
						},
						RepoRelDir: ".",
						Workspace:  "default",
					},
				},
			},
			checkFields: map[string]any{
				"Error": nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			jsonBytes, err := json.Marshal(tc.result)
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

			// Verify ProjectResults is present and is an array
			_, exists := parsed["ProjectResults"]
			Assert(t, exists, "expected ProjectResults field in JSON output")
		})
	}
}

// TestResult_MarshalJSON_ProjectErrorAsString verifies that nested ProjectResult
// errors within Result are serialized as strings, not empty objects.
func TestResult_MarshalJSON_ProjectErrorAsString(t *testing.T) {
	result := command.Result{
		ProjectResults: []command.ProjectResult{
			{
				ProjectCommandOutput: command.ProjectCommandOutput{
					Error: errors.New("terraform init failed"),
				},
				Command:    command.Plan,
				RepoRelDir: "modules/vpc",
				Workspace:  "production",
			},
		},
	}

	jsonBytes, err := json.Marshal(result)
	Ok(t, err)

	// Parse to check nested structure
	var parsed map[string]any
	err = json.Unmarshal(jsonBytes, &parsed)
	Ok(t, err)

	projectResults, ok := parsed["ProjectResults"].([]any)
	Assert(t, ok, "ProjectResults should be an array")
	Assert(t, len(projectResults) == 1, "expected 1 project result")

	project, ok := projectResults[0].(map[string]any)
	Assert(t, ok, "project result should be an object")

	// The error should be a string, not {} or null
	errorVal := project["Error"]
	Assert(t, errorVal != nil, "Error field should not be nil when error exists")
	errorStr, ok := errorVal.(string)
	Assert(t, ok, "Error field should be a string, got %T", errorVal)
	Equals(t, "terraform init failed", errorStr)
}
