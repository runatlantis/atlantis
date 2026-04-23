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

func TestCommandResult_MarshalJSON(t *testing.T) {
	cases := map[string]struct {
		cr          command.Result
		expContains string
	}{
		"error is serialized": {
			cr: command.Result{
				Error: errors.New("test error message"),
			},
			expContains: `"Error":"test error message"`,
		},
		"nil error": {
			cr: command.Result{
				Error: nil,
			},
			expContains: `"Error":null`,
		},
		"project result with error": {
			cr: command.Result{
				ProjectResults: []command.ProjectResult{
					{
						ProjectCommandOutput: command.ProjectCommandOutput{
							Error: errors.New("project error"),
						},
						RepoRelDir:  ".",
						Workspace:   "default",
						ProjectName: "myproject",
						Command:     command.Plan,
					},
				},
			},
			expContains: `"Error":"project error"`,
		},
	}

	for descrip, c := range cases {
		t.Run(descrip, func(t *testing.T) {
			jsonBytes, err := json.Marshal(c.cr)
			Ok(t, err)
			jsonStr := string(jsonBytes)
			Assert(t, len(jsonStr) > 0, "JSON should not be empty")
			Assert(t, jsonStr != "{}", "JSON should not be empty object")
			// Check that the expected string is contained in the JSON
			if c.expContains != "" {
				Assert(t, len(jsonStr) > len(c.expContains), "JSON should contain expected string")
				contains := false
				for i := 0; i <= len(jsonStr)-len(c.expContains); i++ {
					if jsonStr[i:i+len(c.expContains)] == c.expContains {
						contains = true
						break
					}
				}
				Assert(t, contains, "JSON should contain: %s, got: %s", c.expContains, jsonStr)
			}
		})
	}
}

func TestProjectResultMarshalJSON(t *testing.T) {
	cases := map[string]struct {
		pr          command.ProjectResult
		expContains []string
	}{
		"all fields with error": {
			pr: command.ProjectResult{
				ProjectCommandOutput: command.ProjectCommandOutput{
					Error:          errors.New("terraform error"),
					Failure:        "",
					ApplySuccess:   "",
					VersionSuccess: "",
				},
				Command:     command.Plan,
				SubCommand:  "",
				RepoRelDir:  "terraform/infrastructure",
				Workspace:   "default",
				ProjectName: "myproject",
			},
			expContains: []string{
				`"Error":"terraform error"`,
				`"Command":1`,
				`"RepoRelDir":"terraform/infrastructure"`,
				`"Workspace":"default"`,
				`"ProjectName":"myproject"`,
			},
		},
		"nil error": {
			pr: command.ProjectResult{
				ProjectCommandOutput: command.ProjectCommandOutput{
					Error: nil,
				},
				Command:     command.Apply,
				RepoRelDir:  ".",
				Workspace:   "staging",
				ProjectName: "test-project",
			},
			expContains: []string{
				`"Error":null`,
				`"Command":0`,
				`"RepoRelDir":"."`,
				`"Workspace":"staging"`,
				`"ProjectName":"test-project"`,
			},
		},
	}

	for descrip, c := range cases {
		t.Run(descrip, func(t *testing.T) {
			jsonBytes, err := json.Marshal(c.pr)
			Ok(t, err)
			jsonStr := string(jsonBytes)
			Assert(t, len(jsonStr) > 0, "JSON should not be empty")
			Assert(t, jsonStr != "{}", "JSON should not be empty object")
			// Check that all expected strings are contained in the JSON
			for _, expected := range c.expContains {
				Assert(t, len(jsonStr) > len(expected), "JSON should contain expected string")
				contains := false
				for i := 0; i <= len(jsonStr)-len(expected); i++ {
					if jsonStr[i:i+len(expected)] == expected {
						contains = true
						break
					}
				}
				Assert(t, contains, "JSON should contain: %s, got: %s", expected, jsonStr)
			}
		})
	}
}
