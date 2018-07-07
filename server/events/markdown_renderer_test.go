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
	"strings"
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
			events.ApplyCommand,
			err,
			"**Apply Error**\n```\nerr\n```\n\n",
		},
		{
			"plan error",
			events.PlanCommand,
			err,
			"**Plan Error**\n```\nerr\n```\n\n",
		},
	}

	r := events.MarkdownRenderer{}
	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			res := events.CommandResult{
				Error: c.Error,
			}
			for _, verbose := range []bool{true, false} {
				t.Log("testing " + c.Description)
				s := r.Render(res, c.Command, "log", verbose)
				if !verbose {
					Equals(t, c.Expected, s)
				} else {
					Equals(t, c.Expected+"<details><summary>Log</summary>\n  <p>\n\n```\nlog```\n</p></details>\n", s)
				}
			}
		})
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
			events.ApplyCommand,
			"failure",
			"**Apply Failed**: failure\n\n",
		},
		{
			"plan failure",
			events.PlanCommand,
			"failure",
			"**Plan Failed**: failure\n\n",
		},
	}

	r := events.MarkdownRenderer{}
	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			res := events.CommandResult{
				Failure: c.Failure,
			}
			for _, verbose := range []bool{true, false} {
				t.Log("testing " + c.Description)
				s := r.Render(res, c.Command, "log", verbose)
				if !verbose {
					Equals(t, c.Expected, s)
				} else {
					Equals(t, c.Expected+"<details><summary>Log</summary>\n  <p>\n\n```\nlog```\n</p></details>\n", s)
				}
			}
		})
	}
}

func TestRenderErrAndFailure(t *testing.T) {
	t.Log("if there is an error and a failure, the error should be printed")
	r := events.MarkdownRenderer{}
	res := events.CommandResult{
		Error:   errors.New("error"),
		Failure: "failure",
	}
	s := r.Render(res, events.PlanCommand, "", false)
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
			"no projects",
			events.PlanCommand,
			[]events.ProjectResult{},
			"Ran Plan for 0 projects:\n\n\n",
		},
		{
			"single successful plan",
			events.PlanCommand,
			[]events.ProjectResult{
				{
					ProjectCommandResult: events.ProjectCommandResult{
						PlanSuccess: &events.PlanSuccess{
							TerraformOutput: "terraform-output",
							LockURL:         "lock-url",
						},
					},
					Workspace:  "workspace",
					RepoRelDir: "path",
				},
			},
			`Ran Plan in dir: $path$ workspace: $workspace$

$$$diff
terraform-output
$$$

* To **delete** this plan click [here](lock-url)
* To apply all unapplied plans comment $atlantis apply$
`,
		},
		{
			"single successful apply",
			events.ApplyCommand,
			[]events.ProjectResult{
				{
					ProjectCommandResult: events.ProjectCommandResult{
						ApplySuccess: "success",
					},
					Workspace:  "workspace",
					RepoRelDir: "path",
				},
			},
			`Ran Apply in dir: $path$ workspace: $workspace$

$$$diff
success
$$$

`,
		},
		{
			"multiple successful plans",
			events.PlanCommand,
			[]events.ProjectResult{
				{
					Workspace:  "workspace",
					RepoRelDir: "path",
					ProjectCommandResult: events.ProjectCommandResult{
						PlanSuccess: &events.PlanSuccess{
							TerraformOutput: "terraform-output",
							LockURL:         "lock-url",
						},
					},
				},
				{
					Workspace:  "workspace",
					RepoRelDir: "path2",
					ProjectCommandResult: events.ProjectCommandResult{
						PlanSuccess: &events.PlanSuccess{
							TerraformOutput: "terraform-output2",
							LockURL:         "lock-url2",
						},
					},
				},
			},
			`Ran Plan for 2 projects:
1. workspace: $workspace$ dir: $path$
1. workspace: $workspace$ dir: $path2$

### 1. workspace: $workspace$ dir: $path$
$$$diff
terraform-output
$$$

* To **delete** this plan click [here](lock-url)
---
### 2. workspace: $workspace$ dir: $path2$
$$$diff
terraform-output2
$$$

* To **delete** this plan click [here](lock-url2)
---
* To apply all unapplied plans comment $atlantis apply$
`,
		},
		{
			"multiple successful applies",
			events.ApplyCommand,
			[]events.ProjectResult{
				{
					RepoRelDir: "path",
					Workspace:  "workspace",
					ProjectCommandResult: events.ProjectCommandResult{
						ApplySuccess: "success",
					},
				},
				{
					RepoRelDir: "path2",
					Workspace:  "workspace",
					ProjectCommandResult: events.ProjectCommandResult{
						ApplySuccess: "success2",
					},
				},
			},
			`Ran Apply for 2 projects:
1. workspace: $workspace$ dir: $path$
1. workspace: $workspace$ dir: $path2$

### 1. workspace: $workspace$ dir: $path$
$$$diff
success
$$$
---
### 2. workspace: $workspace$ dir: $path2$
$$$diff
success2
$$$
---

`,
		},
		{
			"single errored plan",
			events.PlanCommand,
			[]events.ProjectResult{
				{
					ProjectCommandResult: events.ProjectCommandResult{
						Error: errors.New("error"),
					},
					RepoRelDir: "path",
					Workspace:  "workspace",
				},
			},
			`Ran Plan in dir: $path$ workspace: $workspace$

**Plan Error**
$$$
error
$$$


`,
		},
		{
			"single failed plan",
			events.PlanCommand,
			[]events.ProjectResult{
				{
					RepoRelDir: "path",
					Workspace:  "workspace",
					ProjectCommandResult: events.ProjectCommandResult{
						Failure: "failure",
					},
				},
			},
			`Ran Plan in dir: $path$ workspace: $workspace$

**Plan Failed**: failure


`,
		},
		{
			"successful, failed, and errored plan",
			events.PlanCommand,
			[]events.ProjectResult{
				{
					Workspace:  "workspace",
					RepoRelDir: "path",
					ProjectCommandResult: events.ProjectCommandResult{
						PlanSuccess: &events.PlanSuccess{
							TerraformOutput: "terraform-output",
							LockURL:         "lock-url",
						},
					},
				},
				{
					Workspace:  "workspace",
					RepoRelDir: "path2",
					ProjectCommandResult: events.ProjectCommandResult{
						Failure: "failure",
					},
				},
				{
					Workspace:  "workspace",
					RepoRelDir: "path3",
					ProjectCommandResult: events.ProjectCommandResult{
						Error: errors.New("error"),
					},
				},
			},
			`Ran Plan for 3 projects:
1. workspace: $workspace$ dir: $path$
1. workspace: $workspace$ dir: $path2$
1. workspace: $workspace$ dir: $path3$

### 1. workspace: $workspace$ dir: $path$
$$$diff
terraform-output
$$$

* To **delete** this plan click [here](lock-url)
---
### 2. workspace: $workspace$ dir: $path2$
**Plan Failed**: failure

---
### 3. workspace: $workspace$ dir: $path3$
**Plan Error**
$$$
error
$$$

---
* To apply all unapplied plans comment $atlantis apply$
`,
		},
		{
			"successful, failed, and errored apply",
			events.ApplyCommand,
			[]events.ProjectResult{
				{
					Workspace:  "workspace",
					RepoRelDir: "path",
					ProjectCommandResult: events.ProjectCommandResult{
						ApplySuccess: "success",
					},
				},
				{
					Workspace:  "workspace",
					RepoRelDir: "path2",
					ProjectCommandResult: events.ProjectCommandResult{
						Failure: "failure",
					},
				},
				{
					Workspace:  "workspace",
					RepoRelDir: "path3",
					ProjectCommandResult: events.ProjectCommandResult{
						Error: errors.New("error"),
					},
				},
			},
			`Ran Apply for 3 projects:
1. workspace: $workspace$ dir: $path$
1. workspace: $workspace$ dir: $path2$
1. workspace: $workspace$ dir: $path3$

### 1. workspace: $workspace$ dir: $path$
$$$diff
success
$$$
---
### 2. workspace: $workspace$ dir: $path2$
**Apply Failed**: failure

---
### 3. workspace: $workspace$ dir: $path3$
**Apply Error**
$$$
error
$$$

---

`,
		},
	}

	r := events.MarkdownRenderer{}
	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			res := events.CommandResult{
				ProjectResults: c.ProjectResults,
			}
			for _, verbose := range []bool{true, false} {
				t.Run(c.Description, func(t *testing.T) {
					s := r.Render(res, c.Command, "log", verbose)
					expWithBackticks := strings.Replace(c.Expected, "$", "`", -1)
					if !verbose {
						Equals(t, expWithBackticks, s)
					} else {
						Equals(t, expWithBackticks+"<details><summary>Log</summary>\n  <p>\n\n```\nlog```\n</p></details>\n", s)
					}
				})
			}
		})
	}
}
