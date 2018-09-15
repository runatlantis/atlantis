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

package events_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/models"
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
				s := r.Render(res, c.Command, "log", verbose, models.Github)
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
				s := r.Render(res, c.Command, "log", verbose, models.Github)
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
	s := r.Render(res, events.PlanCommand, "", false, models.Github)
	Equals(t, "**Plan Error**\n```\nerror\n```\n\n", s)
}

func TestRenderProjectResults(t *testing.T) {
	cases := []struct {
		Description    string
		Command        events.CommandName
		ProjectResults []events.ProjectResult
		VCSHost        models.VCSHostType
		Expected       string
	}{
		{
			"no projects",
			events.PlanCommand,
			[]events.ProjectResult{},
			models.Github,
			"Ran Plan for 0 projects:\n\n\n",
		},
		{
			"single successful plan",
			events.PlanCommand,
			[]events.ProjectResult{
				{
					PlanSuccess: &events.PlanSuccess{
						TerraformOutput: "terraform-output",
						LockURL:         "lock-url",
						RePlanCmd:       "atlantis plan -d path -w workspace",
						ApplyCmd:        "atlantis apply -d path -w workspace",
					},
					Workspace:  "workspace",
					RepoRelDir: "path",
				},
			},
			models.Github,
			`Ran Plan in dir: $path$ workspace: $workspace$

$$$diff
terraform-output
$$$

* :arrow_forward: To **apply** this plan, comment:
    * $atlantis apply -d path -w workspace$
* :put_litter_in_its_place: To **delete** this plan click [here](lock-url)
* :repeat: To **plan** this project again, comment:
    * $atlantis plan -d path -w workspace$

---
* :fast_forward: To **apply** all unapplied plans from this pull request, comment:
    * $atlantis apply$
`,
		},
		{
			"single successful apply",
			events.ApplyCommand,
			[]events.ProjectResult{
				{
					ApplySuccess: "success",
					Workspace:    "workspace",
					RepoRelDir:   "path",
				},
			},
			models.Github,
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
					PlanSuccess: &events.PlanSuccess{
						TerraformOutput: "terraform-output",
						LockURL:         "lock-url",
						ApplyCmd:        "atlantis apply -d path -w workspace",
						RePlanCmd:       "atlantis plan -d path -w workspace",
					},
				},
				{
					Workspace:  "workspace",
					RepoRelDir: "path2",
					PlanSuccess: &events.PlanSuccess{
						TerraformOutput: "terraform-output2",
						LockURL:         "lock-url2",
						ApplyCmd:        "atlantis apply -d path2 -w workspace",
						RePlanCmd:       "atlantis plan -d path2 -w workspace",
					},
				},
			},
			models.Github,
			`Ran Plan for 2 projects:
1. workspace: $workspace$ dir: $path$
1. workspace: $workspace$ dir: $path2$

### 1. workspace: $workspace$ dir: $path$
$$$diff
terraform-output
$$$

* :arrow_forward: To **apply** this plan, comment:
    * $atlantis apply -d path -w workspace$
* :put_litter_in_its_place: To **delete** this plan click [here](lock-url)
* :repeat: To **plan** this project again, comment:
    * $atlantis plan -d path -w workspace$
---
### 2. workspace: $workspace$ dir: $path2$
$$$diff
terraform-output2
$$$

* :arrow_forward: To **apply** this plan, comment:
    * $atlantis apply -d path2 -w workspace$
* :put_litter_in_its_place: To **delete** this plan click [here](lock-url2)
* :repeat: To **plan** this project again, comment:
    * $atlantis plan -d path2 -w workspace$
---
* :fast_forward: To **apply** all unapplied plans from this pull request, comment:
    * $atlantis apply$
`,
		},
		{
			"multiple successful applies",
			events.ApplyCommand,
			[]events.ProjectResult{
				{
					RepoRelDir:   "path",
					Workspace:    "workspace",
					ApplySuccess: "success",
				},
				{
					RepoRelDir:   "path2",
					Workspace:    "workspace",
					ApplySuccess: "success2",
				},
			},
			models.Github,
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
					Error:      errors.New("error"),
					RepoRelDir: "path",
					Workspace:  "workspace",
				},
			},
			models.Github,
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
					Failure:    "failure",
				},
			},
			models.Github,
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
					PlanSuccess: &events.PlanSuccess{
						TerraformOutput: "terraform-output",
						LockURL:         "lock-url",
						ApplyCmd:        "atlantis apply -d path -w workspace",
						RePlanCmd:       "atlantis plan -d path -w workspace",
					},
				},
				{
					Workspace:  "workspace",
					RepoRelDir: "path2",
					Failure:    "failure",
				},
				{
					Workspace:  "workspace",
					RepoRelDir: "path3",
					Error:      errors.New("error"),
				},
			},
			models.Github,
			`Ran Plan for 3 projects:
1. workspace: $workspace$ dir: $path$
1. workspace: $workspace$ dir: $path2$
1. workspace: $workspace$ dir: $path3$

### 1. workspace: $workspace$ dir: $path$
$$$diff
terraform-output
$$$

* :arrow_forward: To **apply** this plan, comment:
    * $atlantis apply -d path -w workspace$
* :put_litter_in_its_place: To **delete** this plan click [here](lock-url)
* :repeat: To **plan** this project again, comment:
    * $atlantis plan -d path -w workspace$
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
* :fast_forward: To **apply** all unapplied plans from this pull request, comment:
    * $atlantis apply$
`,
		},
		{
			"successful, failed, and errored apply",
			events.ApplyCommand,
			[]events.ProjectResult{
				{
					Workspace:    "workspace",
					RepoRelDir:   "path",
					ApplySuccess: "success",
				},
				{
					Workspace:  "workspace",
					RepoRelDir: "path2",
					Failure:    "failure",
				},
				{
					Workspace:  "workspace",
					RepoRelDir: "path3",
					Error:      errors.New("error"),
				},
			},
			models.Github,
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
		{
			"successful, failed, and errored apply",
			events.ApplyCommand,
			[]events.ProjectResult{
				{
					Workspace:    "workspace",
					RepoRelDir:   "path",
					ApplySuccess: "success",
				},
				{
					Workspace:  "workspace",
					RepoRelDir: "path2",
					Failure:    "failure",
				},
				{
					Workspace:  "workspace",
					RepoRelDir: "path3",
					Error:      errors.New("error"),
				},
			},
			models.Github,
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
					s := r.Render(res, c.Command, "log", verbose, c.VCSHost)
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

// Test that we format the terraform plan output so it shows up properly
// when using diff syntax highlighting.
func TestRenderProjectResults_DiffSyntax(t *testing.T) {
	r := events.MarkdownRenderer{}
	result := r.Render(
		events.CommandResult{
			ProjectResults: []events.ProjectResult{
				{
					RepoRelDir: ".",
					Workspace:  "default",
					PlanSuccess: &events.PlanSuccess{
						TerraformOutput: `Refreshing Terraform state in-memory prior to plan...
The refreshed state will be used to calculate this plan, but will not be
persisted to local or remote state storage.


------------------------------------------------------------------------

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create
  ~ update in-place
  - destroy

Terraform will perform the following actions:

+ null_resource.test[0]
      id: <computed>

  + null_resource.test[1]
      id: <computed>

  ~ aws_security_group_rule.allow_all
      description: "" => "test3"

  - aws_security_group_rule.allow_all
`,
						LockURL:   "lock-url",
						RePlanCmd: "atlantis plan -d .",
						ApplyCmd:  "atlantis apply -d .",
					},
				},
			},
		},
		events.PlanCommand,
		"log",
		false,
		models.Github,
	)

	exp := `Ran Plan in dir: $.$ workspace: $default$

<details><summary>Show Output</summary>

$$$diff
Refreshing Terraform state in-memory prior to plan...
The refreshed state will be used to calculate this plan, but will not be
persisted to local or remote state storage.


------------------------------------------------------------------------

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
+ create
~ update in-place
- destroy

Terraform will perform the following actions:

+ null_resource.test[0]
      id: <computed>

+ null_resource.test[1]
      id: <computed>

~ aws_security_group_rule.allow_all
      description: "" => "test3"

- aws_security_group_rule.allow_all

$$$

* :arrow_forward: To **apply** this plan, comment:
    * $atlantis apply -d .$
* :put_litter_in_its_place: To **delete** this plan click [here](lock-url)
* :repeat: To **plan** this project again, comment:
    * $atlantis plan -d .$</details>

---
* :fast_forward: To **apply** all unapplied plans from this pull request, comment:
    * $atlantis apply$
`
	expWithBackticks := strings.Replace(exp, "$", "`", -1)
	Equals(t, expWithBackticks, result)
}

// Test that if the output is longer than 12 lines, it gets wrapped on the right
// VCS hosts during an error.
func TestRenderProjectResults_WrappedErr(t *testing.T) {
	cases := []struct {
		VCSHost    models.VCSHostType
		Output     string
		ShouldWrap bool
	}{
		{
			models.Github,
			strings.Repeat("line\n", 1),
			false,
		},
		{
			models.Github,
			strings.Repeat("line\n", 13),
			true,
		},
		{
			models.Gitlab,
			strings.Repeat("line\n", 1),
			false,
		},
		{
			models.Gitlab,
			strings.Repeat("line\n", 13),
			true,
		},
		{
			models.BitbucketCloud,
			strings.Repeat("line\n", 1),
			false,
		},
		{
			models.BitbucketCloud,
			strings.Repeat("line\n", 13),
			false,
		},
		{
			models.BitbucketServer,
			strings.Repeat("line\n", 1),
			false,
		},
		{
			models.BitbucketServer,
			strings.Repeat("line\n", 13),
			false,
		},
	}

	mr := events.MarkdownRenderer{}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%s_%v", c.VCSHost.String(), c.ShouldWrap),
			func(t *testing.T) {
				rendered := mr.Render(events.CommandResult{
					ProjectResults: []events.ProjectResult{
						{
							RepoRelDir: ".",
							Workspace:  "default",
							Error:      errors.New(c.Output),
						},
					},
				}, events.PlanCommand, "log", false, c.VCSHost)
				var exp string
				if c.ShouldWrap {
					exp = `Ran Plan in dir: $.$ workspace: $default$

**Plan Error**
<details><summary>Show Output</summary>

$$$
` + c.Output + `
$$$
</details>

`
				} else {
					exp = `Ran Plan in dir: $.$ workspace: $default$

**Plan Error**
$$$
` + c.Output + `
$$$


`
				}

				expWithBackticks := strings.Replace(exp, "$", "`", -1)
				Equals(t, expWithBackticks, rendered)
			})
	}
}

// Test that if the output is longer than 12 lines, it gets wrapped on the right
// VCS hosts for a single project.
func TestRenderProjectResults_WrapSingleProject(t *testing.T) {
	cases := []struct {
		VCSHost    models.VCSHostType
		Output     string
		ShouldWrap bool
	}{
		{
			models.Github,
			strings.Repeat("line\n", 1),
			false,
		},
		{
			models.Github,
			strings.Repeat("line\n", 13),
			true,
		},
		{
			models.Gitlab,
			strings.Repeat("line\n", 1),
			false,
		},
		{
			models.Gitlab,
			strings.Repeat("line\n", 13),
			true,
		},
		{
			models.BitbucketCloud,
			strings.Repeat("line\n", 1),
			false,
		},
		{
			models.BitbucketCloud,
			strings.Repeat("line\n", 13),
			false,
		},
		{
			models.BitbucketServer,
			strings.Repeat("line\n", 1),
			false,
		},
		{
			models.BitbucketServer,
			strings.Repeat("line\n", 13),
			false,
		},
	}

	mr := events.MarkdownRenderer{}
	for _, c := range cases {
		for _, cmd := range []events.CommandName{events.PlanCommand, events.ApplyCommand} {
			t.Run(fmt.Sprintf("%s_%s_%v", c.VCSHost.String(), cmd.String(), c.ShouldWrap),
				func(t *testing.T) {
					var pr events.ProjectResult
					switch cmd {
					case events.PlanCommand:
						pr = events.ProjectResult{
							RepoRelDir: ".",
							Workspace:  "default",
							PlanSuccess: &events.PlanSuccess{
								TerraformOutput: c.Output,
								LockURL:         "lock-url",
								RePlanCmd:       "replancmd",
								ApplyCmd:        "applycmd",
							},
						}
					case events.ApplyCommand:
						pr = events.ProjectResult{
							RepoRelDir:   ".",
							Workspace:    "default",
							ApplySuccess: c.Output,
						}
					}
					rendered := mr.Render(events.CommandResult{
						ProjectResults: []events.ProjectResult{pr},
					}, cmd, "log", false, c.VCSHost)

					// Check result.
					var exp string
					switch cmd {
					case events.PlanCommand:
						if c.ShouldWrap {
							exp = `Ran Plan in dir: $.$ workspace: $default$

<details><summary>Show Output</summary>

$$$diff
` + c.Output + `
$$$

* :arrow_forward: To **apply** this plan, comment:
    * $applycmd$
* :put_litter_in_its_place: To **delete** this plan click [here](lock-url)
* :repeat: To **plan** this project again, comment:
    * $replancmd$</details>

---
* :fast_forward: To **apply** all unapplied plans from this pull request, comment:
    * $atlantis apply$
`
						} else {
							exp = `Ran Plan in dir: $.$ workspace: $default$

$$$diff
` + c.Output + `
$$$

* :arrow_forward: To **apply** this plan, comment:
    * $applycmd$
* :put_litter_in_its_place: To **delete** this plan click [here](lock-url)
* :repeat: To **plan** this project again, comment:
    * $replancmd$

---
* :fast_forward: To **apply** all unapplied plans from this pull request, comment:
    * $atlantis apply$
`
						}
					case events.ApplyCommand:
						if c.ShouldWrap {
							exp = `Ran Apply in dir: $.$ workspace: $default$

<details><summary>Show Output</summary>

$$$diff
` + c.Output + `
$$$
</details>

`
						} else {
							exp = `Ran Apply in dir: $.$ workspace: $default$

$$$diff
` + c.Output + `
$$$

`
						}
					}

					expWithBackticks := strings.Replace(exp, "$", "`", -1)
					Equals(t, expWithBackticks, rendered)
				})
		}
	}
}

func TestRenderProjectResults_MultiProjectWrapped(t *testing.T) {
	mr := events.MarkdownRenderer{}
	tfOut := strings.Repeat("line\n", 13)
	rendered := mr.Render(events.CommandResult{
		ProjectResults: []events.ProjectResult{
			{
				RepoRelDir:   ".",
				Workspace:    "staging",
				ApplySuccess: tfOut,
			},
			{
				RepoRelDir:   ".",
				Workspace:    "production",
				ApplySuccess: tfOut,
			},
		},
	}, events.ApplyCommand, "log", false, models.Github)
	exp := `Ran Apply for 2 projects:
1. workspace: $staging$ dir: $.$
1. workspace: $production$ dir: $.$

### 1. workspace: $staging$ dir: $.$
<details><summary>Show Output</summary>

$$$diff
` + tfOut + `
$$$
</details>
---
### 2. workspace: $production$ dir: $.$
<details><summary>Show Output</summary>

$$$diff
` + tfOut + `
$$$
</details>
---

`
	expWithBackticks := strings.Replace(exp, "$", "`", -1)
	Equals(t, expWithBackticks, rendered)
}
