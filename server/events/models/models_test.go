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
	"errors"
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

func TestProjectResult_IsSuccessful(t *testing.T) {
	cases := map[string]struct {
		pr  models.ProjectResult
		exp bool
	}{
		"plan success": {
			models.ProjectResult{
				PlanSuccess: &models.PlanSuccess{},
			},
			true,
		},
		"policy_check success": {
			models.ProjectResult{
				PolicyCheckSuccess: &models.PolicyCheckSuccess{},
			},
			true,
		},
		"apply success": {
			models.ProjectResult{
				ApplySuccess: "success",
			},
			true,
		},
		"failure": {
			models.ProjectResult{
				Failure: "failure",
			},
			false,
		},
		"error": {
			models.ProjectResult{
				Error: errors.New("error"),
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
		p         models.ProjectResult
		expStatus models.ProjectPlanStatus
	}{
		{
			p: models.ProjectResult{
				Command: models.PlanCommand,
				Error:   errors.New("err"),
			},
			expStatus: models.ErroredPlanStatus,
		},
		{
			p: models.ProjectResult{
				Command: models.PlanCommand,
				Failure: "failure",
			},
			expStatus: models.ErroredPlanStatus,
		},
		{
			p: models.ProjectResult{
				Command:     models.PlanCommand,
				PlanSuccess: &models.PlanSuccess{},
			},
			expStatus: models.PlannedPlanStatus,
		},
		{
			p: models.ProjectResult{
				Command: models.ApplyCommand,
				Error:   errors.New("err"),
			},
			expStatus: models.ErroredApplyStatus,
		},
		{
			p: models.ProjectResult{
				Command: models.ApplyCommand,
				Failure: "failure",
			},
			expStatus: models.ErroredApplyStatus,
		},
		{
			p: models.ProjectResult{
				Command:      models.ApplyCommand,
				ApplySuccess: "success",
			},
			expStatus: models.AppliedPlanStatus,
		},
		{
			p: models.ProjectResult{
				Command:            models.PolicyCheckCommand,
				PolicyCheckSuccess: &models.PolicyCheckSuccess{},
			},
			expStatus: models.PassedPolicyCheckStatus,
		},
		{
			p: models.ProjectResult{
				Command: models.PolicyCheckCommand,
				Failure: "failure",
			},
			expStatus: models.ErroredPolicyCheckStatus,
		},
		{
			p: models.ProjectResult{
				Command:            models.ApprovePoliciesCommand,
				PolicyCheckSuccess: &models.PolicyCheckSuccess{},
			},
			expStatus: models.PassedPolicyCheckStatus,
		},
		{
			p: models.ProjectResult{
				Command: models.ApprovePoliciesCommand,
				Failure: "failure",
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
		p         models.ProjectResult
		expResult string
	}{
		{
			p: models.ProjectResult{
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
			expResult: "Plan: 0 to add, 0 to change, 1 to destroy.",
		},
		{
			p: models.ProjectResult{
				PlanSuccess: &models.PlanSuccess{
					TerraformOutput: `
					An execution plan has been generated and is shown below.
					Resource actions are indicated with the following symbols:

					No changes. Infrastructure is up-to-date.`,
				},
			},
			expResult: "No changes. Infrastructure is up-to-date.",
		},
		{
			p: models.ProjectResult{
				PlanSuccess: &models.PlanSuccess{
					TerraformOutput: `
					Note: Objects have changed outside of Terraform

					Terraform detected the following changes made outside of Terraform since the
					last "terraform apply":

					No changes. Your infrastructure matches the configuration.`,
				},
			},
			expResult: "\n**Note: Objects have changed outside of Terraform**\nNo changes. Your infrastructure matches the configuration.",
		},
		{
			p: models.ProjectResult{
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
			expResult: "\n**Note: Objects have changed outside of Terraform**\nPlan: 0 to add, 0 to change, 1 to destroy.",
		},
		{
			p: models.ProjectResult{
				PlanSuccess: &models.PlanSuccess{
					TerraformOutput: `No match, expect empty`,
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

func TestApplyCommand_String(t *testing.T) {
	uc := models.ApplyCommand

	Equals(t, "apply", uc.String())
}

func TestPlanCommand_String(t *testing.T) {
	uc := models.PlanCommand

	Equals(t, "plan", uc.String())
}

func TestPolicyCheckCommand_String(t *testing.T) {
	uc := models.PolicyCheckCommand

	Equals(t, "policy_check", uc.String())
}

func TestUnlockCommand_String(t *testing.T) {
	uc := models.UnlockCommand

	Equals(t, "unlock", uc.String())
}
