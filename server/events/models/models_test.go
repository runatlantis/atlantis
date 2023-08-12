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

package models_test

import (
	"fmt"
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	. "github.com/runatlantis/atlantis/testing"
)

func TestNewRepo_EmptyRepoFullName(t *testing.T) {
	_, err := models.NewRepo(models.Github, "", "https://github.com/notowner/repo.git", "u", "p")
	ErrEquals(t, "repoFullName can't be empty", err)
}

func TestNewRepo_EmptyCloneURL(t *testing.T) {
	_, err := models.NewRepo(models.Github, "owner/repo", "", "u", "p")
	ErrEquals(t, "cloneURL can't be empty", err)
}

func TestNewRepo_InvalidCloneURL(t *testing.T) {
	_, err := models.NewRepo(models.Github, "owner/repo", ":", "u", "p")
	ErrEquals(t, "invalid clone url: parse \":.git\": missing protocol scheme", err)
}

func TestNewRepo_CloneURLWrongRepo(t *testing.T) {
	_, err := models.NewRepo(models.Github, "owner/repo", "https://github.com/notowner/repo.git", "u", "p")
	ErrEquals(t, `expected clone url to have path "/owner/repo.git" but had "/notowner/repo.git"`, err)
}

func TestNewRepo_EmptyAzureDevopsProject(t *testing.T) {
	_, err := models.NewRepo(models.AzureDevops, "", "https://dev.azure.com/notowner/project/_git/repo", "u", "p")
	ErrEquals(t, "repoFullName can't be empty", err)
}

// For bitbucket server we don't validate the clone URL because the callers
// are actually constructing it.
func TestNewRepo_CloneURLBitbucketServer(t *testing.T) {
	repo, err := models.NewRepo(models.BitbucketServer, "owner/repo", "http://mycorp.com:7990/scm/at/atlantis-example.git", "u", "p")
	Ok(t, err)
	Equals(t, models.Repo{
		FullName:          "owner/repo",
		Owner:             "owner",
		Name:              "repo",
		CloneURL:          "http://u:p@mycorp.com:7990/scm/at/atlantis-example.git",
		SanitizedCloneURL: "http://u:<redacted>@mycorp.com:7990/scm/at/atlantis-example.git",
		VCSHost: models.VCSHost{
			Hostname: "mycorp.com",
			Type:     models.BitbucketServer,
		},
	}, repo)
}

// If the clone URL contains a space, NewRepo() should encode it
func TestNewRepo_CloneURLContainsSpace(t *testing.T) {
	repo, err := models.NewRepo(models.AzureDevops, "owner/project space/repo", "https://dev.azure.com/owner/project space/repo", "u", "p")
	Ok(t, err)
	Equals(t, repo.CloneURL, "https://u:p@dev.azure.com/owner/project%20space/repo")
	Equals(t, repo.SanitizedCloneURL, "https://u:<redacted>@dev.azure.com/owner/project%20space/repo")

	repo, err = models.NewRepo(models.BitbucketCloud, "owner/repo space", "https://bitbucket.org/owner/repo space", "u", "p")
	Ok(t, err)
	Equals(t, repo.CloneURL, "https://u:p@bitbucket.org/owner/repo%20space.git")
	Equals(t, repo.SanitizedCloneURL, "https://u:<redacted>@bitbucket.org/owner/repo%20space.git")
}

func TestNewRepo_FullNameWrongFormat(t *testing.T) {
	cases := []struct {
		repoFullName string
		expErr       string
	}{
		{
			"owner/repo/extra",
			`invalid repo format "owner/repo/extra", owner "owner/repo" should not contain any /'s`,
		},
		{
			"/",
			`invalid repo format "/", owner "" or repo "" was empty`,
		},
		{
			"//",
			`invalid repo format "//", owner "" or repo "" was empty`,
		},
		{
			"///",
			`invalid repo format "///", owner "" or repo "" was empty`,
		},
		{
			"a/",
			`invalid repo format "a/", owner "" or repo "" was empty`,
		},
		{
			"/b",
			`invalid repo format "/b", owner "" or repo "b" was empty`,
		},
	}
	for _, c := range cases {
		t.Run(c.repoFullName, func(t *testing.T) {
			cloneURL := fmt.Sprintf("https://github.com/%s.git", c.repoFullName)
			_, err := models.NewRepo(models.Github, c.repoFullName, cloneURL, "u", "p")
			ErrEquals(t, c.expErr, err)
		})
	}
}

// If the clone url doesn't end with .git, and VCS is not Azure DevOps, it is appended
func TestNewRepo_MissingDotGit(t *testing.T) {
	repo, err := models.NewRepo(models.BitbucketCloud, "owner/repo", "https://bitbucket.org/owner/repo", "u", "p")
	Ok(t, err)
	Equals(t, repo.CloneURL, "https://u:p@bitbucket.org/owner/repo.git")
	Equals(t, repo.SanitizedCloneURL, "https://u:<redacted>@bitbucket.org/owner/repo.git")
}

func TestNewRepo_HTTPAuth(t *testing.T) {
	// When the url has http the auth should be added.
	repo, err := models.NewRepo(models.Github, "owner/repo", "http://github.com/owner/repo.git", "u", "p")
	Ok(t, err)
	Equals(t, models.Repo{
		VCSHost: models.VCSHost{
			Hostname: "github.com",
			Type:     models.Github,
		},
		SanitizedCloneURL: "http://u:<redacted>@github.com/owner/repo.git",
		CloneURL:          "http://u:p@github.com/owner/repo.git",
		FullName:          "owner/repo",
		Owner:             "owner",
		Name:              "repo",
	}, repo)
}

func TestNewRepo_HTTPSAuth(t *testing.T) {
	// When the url has https the auth should be added.
	repo, err := models.NewRepo(models.Github, "owner/repo", "https://github.com/owner/repo.git", "u", "p")
	Ok(t, err)
	Equals(t, models.Repo{
		VCSHost: models.VCSHost{
			Hostname: "github.com",
			Type:     models.Github,
		},
		SanitizedCloneURL: "https://u:<redacted>@github.com/owner/repo.git",
		CloneURL:          "https://u:p@github.com/owner/repo.git",
		FullName:          "owner/repo",
		Owner:             "owner",
		Name:              "repo",
	}, repo)
}

func TestProject_String(t *testing.T) {
	Equals(t, "repofullname=owner/repo path=my/path", (models.Project{
		RepoFullName: "owner/repo",
		Path:         "my/path",
	}).String())
}

func TestNewProject(t *testing.T) {
	cases := []struct {
		path    string
		expPath string
	}{
		{
			"/",
			".",
		},
		{
			"./another/path",
			"another/path",
		},
		{
			".",
			".",
		},
	}

	for _, c := range cases {
		t.Run(c.path, func(t *testing.T) {
			p := models.NewProject("repo/owner", c.path)
			Equals(t, c.expPath, p.Path)
		})
	}
}

func TestVCSHostType_ToString(t *testing.T) {
	cases := []struct {
		vcsType models.VCSHostType
		exp     string
	}{
		{
			models.Github,
			"Github",
		},
		{
			models.Gitlab,
			"Gitlab",
		},
		{
			models.BitbucketCloud,
			"BitbucketCloud",
		},
		{
			models.BitbucketServer,
			"BitbucketServer",
		},
		{
			models.AzureDevops,
			"AzureDevops",
		},
	}

	for _, c := range cases {
		t.Run(c.exp, func(t *testing.T) {
			Equals(t, c.exp, c.vcsType.String())
		})
	}
}

func TestSplitRepoFullName(t *testing.T) {
	cases := []struct {
		input    string
		expOwner string
		expRepo  string
	}{
		{
			"owner/repo",
			"owner",
			"repo",
		},
		{
			"group/subgroup/owner/repo",
			"group/subgroup/owner",
			"repo",
		},
		{
			"",
			"",
			"",
		},
		{
			"/",
			"",
			"",
		},
		{
			"owner/",
			"",
			"",
		},
		{
			"/repo",
			"",
			"repo",
		},
		{
			"group/subgroup/",
			"",
			"",
		},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			owner, repo := models.SplitRepoFullName(c.input)
			Equals(t, c.expOwner, owner)
			Equals(t, c.expRepo, repo)
		})
	}
}

// These test cases should cover the same behavior as TestSplitRepoFullName
// and only produce different output in the AzureDevops case of
// owner/project/repo.
func TestAzureDevopsSplitRepoFullName(t *testing.T) {
	cases := []struct {
		input      string
		expOwner   string
		expRepo    string
		expProject string
	}{
		{
			"owner/repo",
			"owner",
			"repo",
			"",
		},
		{
			"group/subgroup/owner/repo",
			"group/subgroup/owner",
			"repo",
			"",
		},
		{
			"group/subgroup/owner/project/repo",
			"group/subgroup/owner/project",
			"repo",
			"",
		},
		{
			"",
			"",
			"",
			"",
		},
		{
			"/",
			"",
			"",
			"",
		},
		{
			"owner/",
			"",
			"",
			"",
		},
		{
			"/repo",
			"",
			"repo",
			"",
		},
		{
			"group/subgroup/",
			"",
			"",
			"",
		},
		{
			"owner/project/repo",
			"owner",
			"repo",
			"project",
		},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			owner, project, repo := vcs.SplitAzureDevopsRepoFullName(c.input)
			Equals(t, c.expOwner, owner)
			Equals(t, c.expProject, project)
			Equals(t, c.expRepo, repo)
		})
	}
}

func TestPlanSuccess_Summary(t *testing.T) {
	cases := []struct {
		input string
		exp   string
	}{
		{
			"Note: Objects have changed outside of Terraform\ndummy\nPlan: 0 to add, 1 to change, 2 to destroy.",
			"\n**Note: Objects have changed outside of Terraform**\nPlan: 0 to add, 1 to change, 2 to destroy.",
		},
		{
			"dummy\nPlan: 100 to add, 111 to change, 222 to destroy.",
			"Plan: 100 to add, 111 to change, 222 to destroy.",
		},
		{
			"dummy\nPlan: 42 to import, 53 to add, 64 to change, 75 to destroy.",
			"Plan: 42 to import, 53 to add, 64 to change, 75 to destroy.",
		},
		{
			"Note: Objects have changed outside of Terraform\ndummy\nNo changes. Infrastructure is up-to-date.",
			"\n**Note: Objects have changed outside of Terraform**\nNo changes. Infrastructure is up-to-date.",
		},
		{
			"dummy\nNo changes. Your infrastructure matches the configuration.",
			"No changes. Your infrastructure matches the configuration.",
		},
	}
	for i, c := range cases {
		t.Run(fmt.Sprintf("summary %d", i), func(t *testing.T) {
			pcs := models.PlanSuccess{
				TerraformOutput: c.input,
			}
			Equals(t, c.exp, pcs.Summary())
		})
	}
}

func TestPlanSuccess_DiffSummary(t *testing.T) {
	cases := []struct {
		input string
		exp   string
	}{
		{
			"Note: Objects have changed outside of Terraform\ndummy\nPlan: 0 to add, 1 to change, 2 to destroy.",
			"Plan: 0 to add, 1 to change, 2 to destroy.",
		},
		{
			"dummy\nPlan: 100 to add, 111 to change, 222 to destroy.",
			"Plan: 100 to add, 111 to change, 222 to destroy.",
		},
		{
			"Note: Objects have changed outside of Terraform\ndummy\nNo changes. Infrastructure is up-to-date.",
			"No changes. Infrastructure is up-to-date.",
		},
		{
			"dummy\nNo changes. Your infrastructure matches the configuration.",
			"No changes. Your infrastructure matches the configuration.",
		},
	}
	for i, c := range cases {
		t.Run(fmt.Sprintf("summary %d", i), func(t *testing.T) {
			pcs := models.PlanSuccess{
				TerraformOutput: c.input,
			}
			Equals(t, c.exp, pcs.DiffSummary())
		})
	}
}

func TestPolicyCheckResults_Summary(t *testing.T) {
	cases := []struct {
		description      string
		policysetResults []models.PolicySetResult
		exp              string
	}{
		{
			description: "test single format with single policy set",
			policysetResults: []models.PolicySetResult{
				{
					PolicySetName:  "policy1",
					ConftestOutput: "20 tests, 19 passed, 2 warnings, 0 failures, 0 exceptions",
				},
			},
			exp: "policy set: policy1: 20 tests, 19 passed, 2 warnings, 0 failures, 0 exceptions",
		},
		{
			description: "test multiple formats with multiple policy sets",
			policysetResults: []models.PolicySetResult{
				{
					PolicySetName:  "policy1",
					ConftestOutput: "20 tests, 19 passed, 2 warnings, 0 failures, 0 exceptions",
				},
				{
					PolicySetName:  "policy2",
					ConftestOutput: "3 tests, 0 passed, 1 warning, 1 failure, 0 exceptions, 1 skipped",
				},
				{
					PolicySetName:  "policy3",
					ConftestOutput: "1 test, 0 passed, 1 warning, 1 failure, 1 exception",
				},
			},
			exp: `policy set: policy1: 20 tests, 19 passed, 2 warnings, 0 failures, 0 exceptions
policy set: policy2: 3 tests, 0 passed, 1 warning, 1 failure, 0 exceptions, 1 skipped
policy set: policy3: 1 test, 0 passed, 1 warning, 1 failure, 1 exception`,
		},
	}
	for _, summary := range cases {
		t.Run(summary.description, func(t *testing.T) {
			pcs := models.PolicyCheckResults{
				PolicySetResults: summary.policysetResults,
			}
			Equals(t, summary.exp, pcs.Summary())
		})
	}
}

// Test PolicyCleared and PolicySummary
func TestPolicyCheckResults_PolicyFuncs(t *testing.T) {
	cases := []struct {
		description      string
		policysetResults []models.PolicySetResult
		policyClearedExp bool
		policySummaryExp string
	}{
		{
			description: "single policy set, not passed",
			policysetResults: []models.PolicySetResult{
				{
					PolicySetName: "policy1",
					Passed:        false,
					ReqApprovals:  1,
				},
			},
			policyClearedExp: false,
			policySummaryExp: "policy set: policy1: requires: 1 approval(s), have: 0.",
		},
		{
			description: "single policy set, passed",
			policysetResults: []models.PolicySetResult{
				{
					PolicySetName: "policy1",
					Passed:        true,
					ReqApprovals:  1,
				},
			},
			policyClearedExp: true,
			policySummaryExp: "policy set: policy1: passed.",
		},
		{
			description: "single policy set, fully approved",
			policysetResults: []models.PolicySetResult{
				{
					PolicySetName: "policy1",
					Passed:        false,
					ReqApprovals:  1,
					CurApprovals:  1,
				},
			},
			policyClearedExp: true,
			policySummaryExp: "policy set: policy1: approved.",
		},
		{
			description: "multiple policy sets, different states.",
			policysetResults: []models.PolicySetResult{
				{
					PolicySetName: "policy1",
					Passed:        false,
					ReqApprovals:  2,
					CurApprovals:  0,
				},
				{
					PolicySetName: "policy2",
					Passed:        false,
					ReqApprovals:  1,
					CurApprovals:  1,
				},
				{
					PolicySetName: "policy3",
					Passed:        true,
					ReqApprovals:  1,
					CurApprovals:  0,
				},
			},
			policyClearedExp: false,
			policySummaryExp: `policy set: policy1: requires: 2 approval(s), have: 0.
policy set: policy2: approved.
policy set: policy3: passed.`,
		},
		{
			description: "multiple policy sets, all cleared.",
			policysetResults: []models.PolicySetResult{
				{
					PolicySetName: "policy1",
					Passed:        false,
					ReqApprovals:  2,
					CurApprovals:  2,
				},
				{
					PolicySetName: "policy2",
					Passed:        false,
					ReqApprovals:  1,
					CurApprovals:  1,
				},
				{
					PolicySetName: "policy3",
					Passed:        true,
					ReqApprovals:  1,
					CurApprovals:  0,
				},
			},
			policyClearedExp: true,
			policySummaryExp: `policy set: policy1: approved.
policy set: policy2: approved.
policy set: policy3: passed.`,
		},
	}
	for _, summary := range cases {
		t.Run(summary.description, func(t *testing.T) {
			pcs := models.PolicyCheckResults{
				PolicySetResults: summary.policysetResults,
			}
			Equals(t, summary.policyClearedExp, pcs.PolicyCleared())
			Equals(t, summary.policySummaryExp, pcs.PolicySummary())
		})
	}
}

func TestPullStatus_StatusCount(t *testing.T) {
	ps := models.PullStatus{
		Projects: []models.ProjectStatus{
			{
				Status: models.PlannedPlanStatus,
			},
			{
				Status: models.PlannedPlanStatus,
			},
			{
				Status: models.AppliedPlanStatus,
			},
			{
				Status: models.ErroredApplyStatus,
			},
			{
				Status: models.DiscardedPlanStatus,
			},
			{
				Status: models.ErroredPolicyCheckStatus,
			},
			{
				Status: models.PassedPolicyCheckStatus,
			},
		},
	}

	Equals(t, 2, ps.StatusCount(models.PlannedPlanStatus))
	Equals(t, 1, ps.StatusCount(models.AppliedPlanStatus))
	Equals(t, 1, ps.StatusCount(models.ErroredApplyStatus))
	Equals(t, 0, ps.StatusCount(models.ErroredPlanStatus))
	Equals(t, 1, ps.StatusCount(models.DiscardedPlanStatus))
	Equals(t, 1, ps.StatusCount(models.ErroredPolicyCheckStatus))
	Equals(t, 1, ps.StatusCount(models.PassedPolicyCheckStatus))
}

func TestPlanSuccessStats(t *testing.T) {
	tests := []struct {
		name   string
		output string
		exp    models.PlanSuccessStats
	}{
		{
			"has changes",
			`An execution plan has been generated and is shown below.
					Resource actions are indicated with the following symbols:
					  - destroy
					Terraform will perform the following actions:
					  - null_resource.hi[1]
					Plan: 1 to add, 3 to change, 2 to destroy.`,
			models.PlanSuccessStats{
				Changes: true,

				Add:     1,
				Change:  3,
				Destroy: 2,
			},
		},
		{
			"no changes",
			`An execution plan has been generated and is shown below.
					Resource actions are indicated with the following symbols:
					No changes. Infrastructure is up-to-date.`,
			models.PlanSuccessStats{},
		},
		{
			"changes outside",
			`Note: Objects have changed outside of Terraform
					Terraform detected the following changes made outside of Terraform since the
					last "terraform apply":
					No changes. Your infrastructure matches the configuration.`,
			models.PlanSuccessStats{
				ChangesOutside: true,
			},
		},
		{
			"with imports",
			`Terraform used the selected providers to generate the following execution
			plan. Resource actions are indicated with the following symbols:	
			  + create
			  ~ update in-place
			  - destroy
			Terraform will perform the following actions:
			  - null_resource.hi[1]
			Plan: 42 to import, 31 to add, 20 to change, 1 to destroy.`,
			models.PlanSuccessStats{
				Changes: true,

				Import:  42,
				Add:     31,
				Change:  20,
				Destroy: 1,
			},
		},
		{
			"changes and changes outside",
			`Note: Objects have changed outside of Terraform
					Terraform detected the following changes made outside of Terraform since the
					last "terraform apply":
					An execution plan has been generated and is shown below.
					Resource actions are indicated with the following symbols:
					  - destroy
					Terraform will perform the following actions:
					  - null_resource.hi[1]
					Plan: 3 to add, 0 to change, 1 to destroy.`,
			models.PlanSuccessStats{
				Changes:        true,
				ChangesOutside: true,

				Add:     3,
				Change:  0,
				Destroy: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := models.NewPlanSuccessStats(tt.output)
			if s != tt.exp {
				t.Errorf("\nexp: %#v\ngot: %#v", tt.exp, s)
			}
		})
	}
}

func TestProjectQueue_FindPullRequest(t *testing.T) {
	queue := models.ProjectLockQueue{
		{Pull: models.PullRequest{Num: 15}},
		{Pull: models.PullRequest{Num: 16}},
		{Pull: models.PullRequest{Num: 17}},
	}
	Equals(t, 0, queue.FindPullRequest(15))
	Equals(t, 1, queue.FindPullRequest(16))
	Equals(t, 2, queue.FindPullRequest(17))
	Equals(t, -1, queue.FindPullRequest(20))

	emptyQueue := models.ProjectLockQueue{}
	Equals(t, -1, emptyQueue.FindPullRequest(15))
}

func TestProjectQueue_Dequeue(t *testing.T) {
	queue := models.ProjectLockQueue{
		{Pull: models.PullRequest{Num: 15}},
		{Pull: models.PullRequest{Num: 16}},
		{Pull: models.PullRequest{Num: 17}},
	}

	dequeuedLock, newQueue := queue.Dequeue()
	Equals(t, 15, dequeuedLock.Pull.Num)
	Equals(t, 2, len(newQueue))

	dequeuedLock, newQueue = newQueue.Dequeue()
	Equals(t, 16, dequeuedLock.Pull.Num)
	Equals(t, 1, len(newQueue))

	dequeuedLock, newQueue = newQueue.Dequeue()
	Equals(t, 17, dequeuedLock.Pull.Num)
	Equals(t, 0, len(newQueue))

	dequeuedLock, newQueue = newQueue.Dequeue()
	Assert(t, dequeuedLock == nil, "dequeued lock was not nil")
	Equals(t, 0, len(newQueue))
}
