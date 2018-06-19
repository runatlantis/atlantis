// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.
//
package events_test

import (
	"errors"
	"testing"

	"github.com/runatlantis/atlantis/server/events"
	. "github.com/runatlantis/atlantis/testing"
)

func TestRenderErr(t *testing.T) {
	err := errors.New("err")
	cases := []struct {
		Description string
		Command     events.CommandName
		Error       error
		Expected    string
	}{
		{
			"apply error",
			events.Apply,
			err,
			"**Apply Error**\n```\nerr\n```\n\n",
		},
		{
			"plan error",
			events.Plan,
			err,
			"**Plan Error**\n```\nerr\n```\n\n",
		},
	}

	r := events.MarkdownRenderer{}
	for _, c := range cases {
		res := events.CommandResponse{
			Error: c.Error,
		}
		for _, verbose := range []bool{true, false} {
			t.Log("testing " + c.Description)
			s := r.Render(res, c.Command, "log", verbose, false)
			if !verbose {
				Equals(t, c.Expected, s)
			} else {
				Equals(t, c.Expected+"<details><summary>Log</summary>\n  <p>\n\n```\nlog```\n</p></details>\n", s)
			}
		}
	}
}

func TestRenderFailure(t *testing.T) {
	cases := []struct {
		Description string
		Command     events.CommandName
		Failure     string
		Expected    string
	}{
		{
			"apply failure",
			events.Apply,
			"failure",
			"**Apply Failed**: failure\n\n",
		},
		{
			"plan failure",
			events.Plan,
			"failure",
			"**Plan Failed**: failure\n\n",
		},
	}

	r := events.MarkdownRenderer{}
	for _, c := range cases {
		res := events.CommandResponse{
			Failure: c.Failure,
		}
		for _, verbose := range []bool{true, false} {
			t.Log("testing " + c.Description)
			s := r.Render(res, c.Command, "log", verbose, false)
			if !verbose {
				Equals(t, c.Expected, s)
			} else {
				Equals(t, c.Expected+"<details><summary>Log</summary>\n  <p>\n\n```\nlog```\n</p></details>\n", s)
			}
		}
	}
}

func TestRenderErrAndFailure(t *testing.T) {
	t.Log("if there is an error and a failure, the error should be printed")
	r := events.MarkdownRenderer{}
	res := events.CommandResponse{
		Error:   errors.New("error"),
		Failure: "failure",
	}
	s := r.Render(res, events.Plan, "", false, false)
	Equals(t, "**Plan Error**\n```\nerror\n```\n\n", s)
}

func TestRenderProjectResults(t *testing.T) {
	cases := []struct {
		Description    string
		Command        events.CommandName
		ProjectResults []events.ProjectResult
		Expected       string
	}{
		{
			"single successful plan",
			events.Plan,
			[]events.ProjectResult{
				{
					PlanSuccess: &events.PlanSuccess{
						TerraformOutput: "terraform-output",
						LockURL:         "lock-url",
					},
					Workspace: "workspace",
					Path:      "path",
				},
			},
			"Ran Plan in dir: `path` workspace: `workspace`\n```diff\nterraform-output\n```\n\n* To **discard** this plan click [here](lock-url).\n\n",
		},
		{
			"single successful apply",
			events.Apply,
			[]events.ProjectResult{
				{
					ApplySuccess: "success",
					Workspace:    "workspace",
					Path:         "path",
				},
			},
			"Ran Apply in dir: `path` workspace: `workspace`\n```diff\nsuccess\n```\n\n",
		},
		{
			"multiple successful plans",
			events.Plan,
			[]events.ProjectResult{
				{
					Workspace: "workspace",
					Path:      "path",
					PlanSuccess: &events.PlanSuccess{
						TerraformOutput: "terraform-output",
						LockURL:         "lock-url",
					},
				},
				{
					Workspace: "workspace",
					Path:      "path2",
					PlanSuccess: &events.PlanSuccess{
						TerraformOutput: "terraform-output2",
						LockURL:         "lock-url2",
					},
				},
			},
			"Ran Plan for 2 projects:\n1. workspace: `workspace` path: `path`\n1. workspace: `workspace` path: `path2`\n\n### 1. workspace: `workspace` path: `path`\n```diff\nterraform-output\n```\n\n* To **discard** this plan click [here](lock-url).\n---\n### 2. workspace: `workspace` path: `path2`\n```diff\nterraform-output2\n```\n\n* To **discard** this plan click [here](lock-url2).\n---\n\n",
		},
		{
			"multiple successful applies",
			events.Apply,
			[]events.ProjectResult{
				{
					Path:         "path",
					Workspace:    "workspace",
					ApplySuccess: "success",
				},
				{
					Path:         "path2",
					Workspace:    "workspace",
					ApplySuccess: "success2",
				},
			},
			"Ran Apply for 2 projects:\n1. workspace: `workspace` path: `path`\n1. workspace: `workspace` path: `path2`\n\n### 1. workspace: `workspace` path: `path`\n```diff\nsuccess\n```\n---\n### 2. workspace: `workspace` path: `path2`\n```diff\nsuccess2\n```\n---\n\n",
		},
		{
			"single errored plan",
			events.Plan,
			[]events.ProjectResult{
				{
					Error:     errors.New("error"),
					Path:      "path",
					Workspace: "workspace",
				},
			},
			"Ran Plan in dir: `path` workspace: `workspace`\n**Plan Error**\n```\nerror\n```\n\n\n",
		},
		{
			"single failed plan",
			events.Plan,
			[]events.ProjectResult{
				{
					Path:      "path",
					Workspace: "workspace",
					Failure:   "failure",
				},
			},
			"Ran Plan in dir: `path` workspace: `workspace`\n**Plan Failed**: failure\n\n\n",
		},
		{
			"successful, failed, and errored plan",
			events.Plan,
			[]events.ProjectResult{
				{
					Workspace: "workspace",
					Path:      "path",
					PlanSuccess: &events.PlanSuccess{
						TerraformOutput: "terraform-output",
						LockURL:         "lock-url",
					},
				},
				{
					Workspace: "workspace",
					Path:      "path2",
					Failure:   "failure",
				},
				{
					Workspace: "workspace",
					Path:      "path3",
					Error:     errors.New("error"),
				},
			},
			"Ran Plan for 3 projects:\n1. workspace: `workspace` path: `path`\n1. workspace: `workspace` path: `path2`\n1. workspace: `workspace` path: `path3`\n\n### 1. workspace: `workspace` path: `path`\n```diff\nterraform-output\n```\n\n* To **discard** this plan click [here](lock-url).\n---\n### 2. workspace: `workspace` path: `path2`\n**Plan Failed**: failure\n\n---\n### 3. workspace: `workspace` path: `path3`\n**Plan Error**\n```\nerror\n```\n\n---\n\n",
		},
		{
			"successful, failed, and errored apply",
			events.Apply,
			[]events.ProjectResult{
				{
					Workspace:    "workspace",
					Path:         "path",
					ApplySuccess: "success",
				},
				{
					Workspace: "workspace",
					Path:      "path2",
					Failure:   "failure",
				},
				{
					Workspace: "workspace",
					Path:      "path3",
					Error:     errors.New("error"),
				},
			},
			"Ran Apply for 3 projects:\n1. workspace: `workspace` path: `path`\n1. workspace: `workspace` path: `path2`\n1. workspace: `workspace` path: `path3`\n\n### 1. workspace: `workspace` path: `path`\n```diff\nsuccess\n```\n---\n### 2. workspace: `workspace` path: `path2`\n**Apply Failed**: failure\n\n---\n### 3. workspace: `workspace` path: `path3`\n**Apply Error**\n```\nerror\n```\n\n---\n\n",
		},
	}

	r := events.MarkdownRenderer{}
	for _, c := range cases {
		res := events.CommandResponse{
			ProjectResults: c.ProjectResults,
		}
		for _, verbose := range []bool{true, false} {
			t.Run(c.Description, func(t *testing.T) {
				s := r.Render(res, c.Command, "log", verbose, false)
				if !verbose {
					Equals(t, c.Expected, s)
				} else {
					Equals(t, c.Expected+"<details><summary>Log</summary>\n  <p>\n\n```\nlog```\n</p></details>\n", s)
				}
			})
		}
	}
}
