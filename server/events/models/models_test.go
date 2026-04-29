// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package models_test

import (
	"fmt"
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs/azuredevops"
	. "github.com/runatlantis/atlantis/testing"
)

func TestNewRepo_EmptyRepoFullName(t *testing.T) {
	_, err := models.NewRepo(models.Github, "", "https://github.com/notowner/repo.git", "u", "p", "")
	ErrEquals(t, "repoFullName can't be empty", err)
}

func TestNewRepo_EmptyCloneURL(t *testing.T) {
	_, err := models.NewRepo(models.Github, "owner/repo", "", "u", "p", "")
	ErrEquals(t, "cloneURL can't be empty", err)
}

func TestNewRepo_InvalidCloneURL(t *testing.T) {
	_, err := models.NewRepo(models.Github, "owner/repo", ":", "u", "p", "")
	ErrEquals(t, "invalid clone url: parse \":.git\": missing protocol scheme", err)
}

func TestNewRepo_CloneURLWrongRepo(t *testing.T) {
	_, err := models.NewRepo(models.Github, "owner/repo", "https://github.com/notowner/repo.git", "u", "p", "")
	ErrEquals(t, `expected clone url to have path "/owner/repo.git" but had "/notowner/repo.git"`, err)
}

func TestNewRepo_EmptyAzureDevopsProject(t *testing.T) {
	_, err := models.NewRepo(models.AzureDevops, "", "https://dev.azure.com/notowner/project/_git/repo", "u", "p", "")
	ErrEquals(t, "repoFullName can't be empty", err)
}

// For bitbucket server we don't validate the clone URL because the callers
// are actually constructing it.
func TestNewRepo_CloneURLBitbucketServer(t *testing.T) {
	repo, err := models.NewRepo(models.BitbucketServer, "owner/repo", "http://mycorp.com:7990/scm/at/atlantis-example.git", "u", "p", "")
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
	repo, err := models.NewRepo(models.AzureDevops, "owner/project space/repo", "https://dev.azure.com/owner/project space/repo", "u", "p", "")
	Ok(t, err)
	Equals(t, repo.CloneURL, "https://u:p@dev.azure.com/owner/project%20space/repo")
	Equals(t, repo.SanitizedCloneURL, "https://u:<redacted>@dev.azure.com/owner/project%20space/repo")

	repo, err = models.NewRepo(models.BitbucketCloud, "owner/repo space", "https://bitbucket.org/owner/repo space", "u", "p", "")
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
			_, err := models.NewRepo(models.Github, c.repoFullName, cloneURL, "u", "p", "")
			ErrEquals(t, c.expErr, err)
		})
	}
}

// If the clone url doesn't end with .git, and VCS is not Azure DevOps, it is appended
func TestNewRepo_MissingDotGit(t *testing.T) {
	repo, err := models.NewRepo(models.BitbucketCloud, "owner/repo", "https://bitbucket.org/owner/repo", "u", "p", "")
	Ok(t, err)
	Equals(t, repo.CloneURL, "https://u:p@bitbucket.org/owner/repo.git")
	Equals(t, repo.SanitizedCloneURL, "https://u:<redacted>@bitbucket.org/owner/repo.git")
}

func TestNewRepo_HTTPAuth(t *testing.T) {
	// When the url has http the auth should be added.
	repo, err := models.NewRepo(models.Github, "owner/repo", "http://github.com/owner/repo.git", "u", "p", "")
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
	repo, err := models.NewRepo(models.Github, "owner/repo", "https://github.com/owner/repo.git", "u", "p", "")
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

// TestNewRepo_Gitlab_HostAndSubpath exercises the GitLab-specific host and
// subpath validation introduced for subpath-hosted GitLab instances.
func TestNewRepo_Gitlab_SaaS_Valid(t *testing.T) {
	repo, err := models.NewRepo(
		models.Gitlab, "owner/repo",
		"https://gitlab.com/owner/repo.git",
		"u", "p", "gitlab.com")
	Ok(t, err)
	Equals(t, "owner/repo", repo.FullName)
	Equals(t, "gitlab.com", repo.VCSHost.Hostname)
}

func TestNewRepo_Gitlab_SaaS_WrongHost(t *testing.T) {
	_, err := models.NewRepo(
		models.Gitlab, "owner/repo",
		"https://evil.com/owner/repo.git",
		"u", "p", "gitlab.com")
	ErrEquals(t, `expected clone url host "gitlab.com" but had "evil.com"`, err)
}

func TestNewRepo_Gitlab_SaaS_UnexpectedSubpath(t *testing.T) {
	// gitlab.com configured with no subpath — a URL carrying a subpath must
	// still be rejected (regression guard against the suffix-match approach).
	_, err := models.NewRepo(
		models.Gitlab, "owner/repo",
		"https://gitlab.com/x/owner/repo.git",
		"u", "p", "gitlab.com")
	ErrEquals(t, `expected clone url to have path "/owner/repo.git" but had "/x/owner/repo.git"`, err)
}

func TestNewRepo_Gitlab_Subpath_Valid(t *testing.T) {
	repo, err := models.NewRepo(
		models.Gitlab, "owner/repo",
		"https://acme.com/gitlab/owner/repo.git",
		"u", "p", "acme.com/gitlab")
	Ok(t, err)
	Equals(t, "owner/repo", repo.FullName)
}

func TestNewRepo_Gitlab_Subpath_MissingSubpath(t *testing.T) {
	_, err := models.NewRepo(
		models.Gitlab, "owner/repo",
		"https://acme.com/owner/repo.git",
		"u", "p", "acme.com/gitlab")
	ErrEquals(t, `expected clone url to have path "/gitlab/owner/repo.git" but had "/owner/repo.git"`, err)
}

func TestNewRepo_Gitlab_Subpath_WrongSubpath(t *testing.T) {
	_, err := models.NewRepo(
		models.Gitlab, "owner/repo",
		"https://acme.com/other/owner/repo.git",
		"u", "p", "acme.com/gitlab")
	ErrEquals(t, `expected clone url to have path "/gitlab/owner/repo.git" but had "/other/owner/repo.git"`, err)
}

func TestNewRepo_Gitlab_Subpath_WrongHost(t *testing.T) {
	_, err := models.NewRepo(
		models.Gitlab, "owner/repo",
		"https://evil.com/gitlab/owner/repo.git",
		"u", "p", "acme.com/gitlab")
	ErrEquals(t, `expected clone url host "acme.com" but had "evil.com"`, err)
}

func TestNewRepo_Gitlab_Host_CaseInsensitive(t *testing.T) {
	_, err := models.NewRepo(
		models.Gitlab, "owner/repo",
		"https://acme.com/gitlab/owner/repo.git",
		"u", "p", "ACME.com/gitlab")
	Ok(t, err)
}

func TestNewRepo_Gitlab_EmptyHostname_SkipsEnhancedCheck(t *testing.T) {
	// When vcsHostname is empty, NewRepo falls back to the pre-change
	// strict-path-equality behavior. cmd/server.go defaults GitlabHostname to
	// "gitlab.com" in production, so this path guards direct-API callers and
	// test helpers that construct EventParser without setting a hostname.
	_, err := models.NewRepo(
		models.Gitlab, "owner/repo",
		"https://acme.com/gitlab/owner/repo.git",
		"u", "p", "")
	ErrEquals(t, `expected clone url to have path "/owner/repo.git" but had "/gitlab/owner/repo.git"`, err)
}

func TestNewRepo_Gitlab_InvalidConfiguredHostname(t *testing.T) {
	// If the configured hostname is itself malformed, NewRepo must surface a
	// wrapping error rather than silently skipping validation.
	_, err := models.NewRepo(
		models.Gitlab, "owner/repo",
		"https://acme.com/gitlab/owner/repo.git",
		"u", "p", "https:///no/host/here")
	ErrContains(t, "parsing configured gitlab hostname", err)
}

func TestProject_String(t *testing.T) {
	Equals(t, "repofullname=owner/repo path=my/path", (models.Project{
		RepoFullName: "owner/repo",
		Path:         "my/path",
	}).String())
}

func TestNewProject(t *testing.T) {
	cases := []struct {
		repo       string
		path       string
		name       string
		expProject models.Project
	}{
		{
			repo: "foo/bar",
			path: "/",
			name: "",
			expProject: models.Project{
				ProjectName:  "",
				RepoFullName: "foo/bar",
				Path:         ".",
			},
		},
		{
			repo: "baz/foo",
			path: "./another/path",
			name: "somename",
			expProject: models.Project{
				ProjectName:  "somename",
				RepoFullName: "baz/foo",
				Path:         "another/path",
			},
		},
		{
			repo: "baz/foo",
			path: ".",
			name: "somename",
			expProject: models.Project{
				ProjectName:  "somename",
				RepoFullName: "baz/foo",
				Path:         ".",
			},
		},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%s_%s", c.name, c.path), func(t *testing.T) {
			p := models.NewProject(c.repo, c.path, c.name)
			Equals(t, c.expProject, p)
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
			owner, project, repo := azuredevops.SplitAzureDevopsRepoFullName(c.input)
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
					PolicySetName: "policy1",
					PolicyOutput:  "20 tests, 19 passed, 2 warnings, 0 failures, 0 exceptions",
				},
			},
			exp: "policy set: policy1: 20 tests, 19 passed, 2 warnings, 0 failures, 0 exceptions",
		},
		{
			description: "test multiple formats with multiple policy sets",
			policysetResults: []models.PolicySetResult{
				{
					PolicySetName: "policy1",
					PolicyOutput:  "20 tests, 19 passed, 2 warnings, 0 failures, 0 exceptions",
				},
				{
					PolicySetName: "policy2",
					PolicyOutput:  "3 tests, 0 passed, 1 warning, 1 failure, 0 exceptions, 1 skipped",
				},
				{
					PolicySetName: "policy3",
					PolicyOutput:  "1 test, 0 passed, 1 warning, 1 failure, 1 exception",
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
					PolicySetName:    "policy1",
					Passed:           false,
					ReqApprovalCount: 1,
				},
			},
			policyClearedExp: false,
			policySummaryExp: "policy set: policy1: requires: 1 approval(s), have: 0.",
		},
		{
			description: "single policy set, passed",
			policysetResults: []models.PolicySetResult{
				{
					PolicySetName:    "policy1",
					Passed:           true,
					ReqApprovalCount: 1,
				},
			},
			policyClearedExp: true,
			policySummaryExp: "policy set: policy1: passed.",
		},
		{
			description: "single policy set, fully approved",
			policysetResults: []models.PolicySetResult{
				{
					PolicySetName:    "policy1",
					Passed:           false,
					ReqApprovalCount: 1,
					Approvals:        make([]models.PolicySetApproval, 1),
				},
			},
			policyClearedExp: true,
			policySummaryExp: "policy set: policy1: approved.",
		},
		{
			description: "multiple policy sets, different states.",
			policysetResults: []models.PolicySetResult{
				{
					PolicySetName:    "policy1",
					Passed:           false,
					ReqApprovalCount: 2,
				},
				{
					PolicySetName:    "policy2",
					Passed:           false,
					ReqApprovalCount: 1,
					Approvals:        make([]models.PolicySetApproval, 1),
				},
				{
					PolicySetName:    "policy3",
					Passed:           true,
					ReqApprovalCount: 1,
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
					PolicySetName:    "policy1",
					Passed:           false,
					ReqApprovalCount: 2,
					Approvals:        make([]models.PolicySetApproval, 2),
				},
				{
					PolicySetName:    "policy2",
					Passed:           false,
					ReqApprovalCount: 1,
					Approvals:        make([]models.PolicySetApproval, 1),
				},
				{
					PolicySetName:    "policy3",
					Passed:           true,
					ReqApprovalCount: 1,
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

func TestApprovalCoversAllHashes(t *testing.T) {
	cases := []struct {
		description    string
		approvalHashes []string
		required       []string
		expected       bool
	}{
		{
			description:    "empty required always covered",
			approvalHashes: nil,
			required:       nil,
			expected:       true,
		},
		{
			description:    "empty required with non-empty approval",
			approvalHashes: []string{"a", "b"},
			required:       nil,
			expected:       true,
		},
		{
			description:    "exact match",
			approvalHashes: []string{"hash1", "hash2"},
			required:       []string{"hash1", "hash2"},
			expected:       true,
		},
		{
			description:    "superset covers required",
			approvalHashes: []string{"hash1", "hash2", "hash3"},
			required:       []string{"hash1", "hash2"},
			expected:       true,
		},
		{
			description:    "missing hash fails",
			approvalHashes: []string{"hash1"},
			required:       []string{"hash1", "hash2"},
			expected:       false,
		},
		{
			description:    "empty approval with non-empty required fails",
			approvalHashes: nil,
			required:       []string{"hash1"},
			expected:       false,
		},
		{
			description:    "completely disjoint fails",
			approvalHashes: []string{"a", "b"},
			required:       []string{"c", "d"},
			expected:       false,
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			Equals(t, c.expected, models.ApprovalCoversAllHashes(c.approvalHashes, c.required))
		})
	}
}

func TestPolicySetStatus_GetCurApprovals(t *testing.T) {
	cases := []struct {
		description string
		status      models.PolicySetStatus
		expected    int
	}{
		{
			description: "nil approvals returns 0",
			status:      models.PolicySetStatus{Hashes: []string{"h1"}},
			expected:    0,
		},
		{
			description: "empty approvals returns 0",
			status:      models.PolicySetStatus{Approvals: []models.PolicySetApproval{}, Hashes: []string{"h1"}},
			expected:    0,
		},
		{
			description: "all approvals cover hashes",
			status: models.PolicySetStatus{
				Hashes: []string{"h1", "h2"},
				Approvals: []models.PolicySetApproval{
					{Approver: "user1", Hashes: []string{"h1", "h2"}},
					{Approver: "user2", Hashes: []string{"h1", "h2", "h3"}},
				},
			},
			expected: 2,
		},
		{
			description: "only matching approvals counted",
			status: models.PolicySetStatus{
				Hashes: []string{"h1", "h2"},
				Approvals: []models.PolicySetApproval{
					{Approver: "user1", Hashes: []string{"h1", "h2"}},
					{Approver: "user2", Hashes: []string{"h1"}},
				},
			},
			expected: 1,
		},
		{
			description: "nil hashes means all approvals count",
			status: models.PolicySetStatus{
				Hashes: nil,
				Approvals: []models.PolicySetApproval{
					{Approver: "user1", Hashes: nil},
					{Approver: "user2", Hashes: []string{"h1"}},
				},
			},
			expected: 2,
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			Equals(t, c.expected, c.status.GetCurApprovals())
		})
	}
}

func TestPolicySetStatus_OwnerHasFullyApproved(t *testing.T) {
	cases := []struct {
		description string
		status      models.PolicySetStatus
		owner       string
		expected    bool
	}{
		{
			description: "owner not in approvals",
			status: models.PolicySetStatus{
				Hashes:    []string{"h1"},
				Approvals: []models.PolicySetApproval{{Approver: "other", Hashes: []string{"h1"}}},
			},
			owner:    "user1",
			expected: false,
		},
		{
			description: "owner approved but hashes stale",
			status: models.PolicySetStatus{
				Hashes:    []string{"h1", "h2"},
				Approvals: []models.PolicySetApproval{{Approver: "user1", Hashes: []string{"h1"}}},
			},
			owner:    "user1",
			expected: false,
		},
		{
			description: "owner approved with matching hashes",
			status: models.PolicySetStatus{
				Hashes:    []string{"h1", "h2"},
				Approvals: []models.PolicySetApproval{{Approver: "user1", Hashes: []string{"h1", "h2"}}},
			},
			owner:    "user1",
			expected: true,
		},
		{
			description: "empty approvals returns false",
			status: models.PolicySetStatus{
				Hashes:    []string{"h1"},
				Approvals: nil,
			},
			owner:    "user1",
			expected: false,
		},
		{
			description: "nil hashes means any approval covers",
			status: models.PolicySetStatus{
				Hashes:    nil,
				Approvals: []models.PolicySetApproval{{Approver: "user1", Hashes: nil}},
			},
			owner:    "user1",
			expected: true,
		},
		{
			description: "stale approval followed by valid approval from same owner",
			status: models.PolicySetStatus{
				Hashes: []string{"h1", "h2"},
				Approvals: []models.PolicySetApproval{
					{Approver: "user1", Hashes: []string{"h1"}},
					{Approver: "user1", Hashes: []string{"h1", "h2"}},
				},
			},
			owner:    "user1",
			expected: true,
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			Equals(t, c.expected, c.status.OwnerHasFullyApproved(c.owner))
		})
	}
}

func TestNewPolicySetResult(t *testing.T) {
	cases := []struct {
		description      string
		name             string
		output           string
		passed           bool
		reqApprovalCount int
		regex            string
		expHashes        []string
	}{
		{
			description:      "dot-star regex includes empty-string hash",
			name:             "policy1",
			output:           "FAIL - plan.json - deny\n\n2 tests, 1 passed, 0 warnings, 1 failure, 0 exceptions\n",
			passed:           false,
			reqApprovalCount: 1,
			regex:            `.*`,
			expHashes: []string{
				models.HashPolicyItem("FAIL - plan.json - deny"),
				models.HashPolicyItem(""),
				models.HashPolicyItem("2 tests, 1 passed, 0 warnings, 1 failure, 0 exceptions"),
			},
		},
		{
			description:      "custom multiline regex extracts only FAIL lines",
			name:             "policy1",
			output:           "FAIL - plan.json - deny\nOK - plan.json - allow\nFAIL - plan.json - other",
			passed:           false,
			reqApprovalCount: 1,
			regex:            `(?m)^FAIL.*`,
			expHashes: []string{
				models.HashPolicyItem("FAIL - plan.json - deny"),
				models.HashPolicyItem("FAIL - plan.json - other"),
			},
		},
		{
			description:      "empty output with dot-star regex yields hash of empty string",
			name:             "policy1",
			output:           "",
			passed:           true,
			reqApprovalCount: 1,
			regex:            `.*`,
			expHashes:        []string{models.HashPolicyItem("")},
		},
		{
			description:      "default regex matches entire output as one item",
			name:             "policy1",
			output:           "FAIL - plan.json - deny\n\n2 tests, 1 passed, 0 warnings, 1 failure, 0 exceptions\n",
			passed:           false,
			reqApprovalCount: 1,
			regex:            `(?s).+`,
			expHashes: []string{
				models.HashPolicyItem("FAIL - plan.json - deny\n\n2 tests, 1 passed, 0 warnings, 1 failure, 0 exceptions\n"),
			},
		},
		{
			description:      "default regex on empty output yields hash of empty string",
			name:             "policy1",
			output:           "",
			passed:           true,
			reqApprovalCount: 1,
			regex:            `(?s).+`,
			expHashes:        []string{models.HashPolicyItem("")},
		},
		{
			description:      "no matches with specific regex falls back to hash of raw output",
			name:             "policy1",
			output:           "all good here",
			passed:           true,
			reqApprovalCount: 0,
			regex:            `^FAIL.*`,
			expHashes:        []string{models.HashPolicyItem("all good here")},
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			result, err := models.NewPolicySetResult(c.name, c.output, c.passed, c.reqApprovalCount, c.regex)
			Ok(t, err)
			Equals(t, c.name, result.PolicySetName)
			Equals(t, c.output, result.PolicyOutput)
			Equals(t, c.passed, result.Passed)
			Equals(t, c.reqApprovalCount, result.ReqApprovalCount)
			Equals(t, c.expHashes, result.Hashes)
			Assert(t, result.Approvals == nil, "new result should have nil approvals")
		})
	}
}

func TestNewPolicySetResult_InvalidRegex(t *testing.T) {
	result, err := models.NewPolicySetResult("policy1", "output", true, 1, `[invalid`)
	Assert(t, err != nil, "expected error for invalid regex")
	Assert(t, result == nil, "expected nil result for invalid regex")
}

func TestPolicySetResult_GetCurApprovals(t *testing.T) {
	cases := []struct {
		description string
		result      models.PolicySetResult
		expected    int
	}{
		{
			description: "nil approvals",
			result:      models.PolicySetResult{Hashes: []string{"h1"}},
			expected:    0,
		},
		{
			description: "approval covers all hashes",
			result: models.PolicySetResult{
				Hashes:    []string{"h1", "h2"},
				Approvals: []models.PolicySetApproval{{Approver: "u", Hashes: []string{"h1", "h2"}}},
			},
			expected: 1,
		},
		{
			description: "stale approval not counted",
			result: models.PolicySetResult{
				Hashes:    []string{"h1", "h2"},
				Approvals: []models.PolicySetApproval{{Approver: "u", Hashes: []string{"h1"}}},
			},
			expected: 0,
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			Equals(t, c.expected, c.result.GetCurApprovals())
		})
	}
}

func TestPolicySetResult_ApprovedHashes(t *testing.T) {
	cases := []struct {
		description string
		result      models.PolicySetResult
		expected    []string
	}{
		{
			description: "no approvals returns empty",
			result:      models.PolicySetResult{},
			expected:    nil,
		},
		{
			description: "single approval returns its hashes",
			result: models.PolicySetResult{
				Approvals: []models.PolicySetApproval{
					{Approver: "u1", Hashes: []string{"a", "b"}},
				},
			},
			expected: []string{"a", "b"},
		},
		{
			description: "multiple approvals deduplicates",
			result: models.PolicySetResult{
				Approvals: []models.PolicySetApproval{
					{Approver: "u1", Hashes: []string{"a", "b"}},
					{Approver: "u2", Hashes: []string{"b", "c"}},
				},
			},
			expected: []string{"a", "b", "c"},
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			Equals(t, c.expected, c.result.ApprovedHashes())
		})
	}
}

func TestNewPolicySetResult_StoresPolicyItemRegex(t *testing.T) {
	result, err := models.NewPolicySetResult("test", "line1\nline2", true, 1, `line\d`)
	Ok(t, err)
	Equals(t, `line\d`, result.PolicyItemRegex)
	Equals(t, []string{models.HashPolicyItem("line1"), models.HashPolicyItem("line2")}, result.Hashes)
}

func TestPolicySetResult_PolicySetHashes(t *testing.T) {
	result := models.PolicySetResult{
		PolicyOutput: "FAIL - plan.json - deny\nOK - plan.json - allow\nFAIL - plan.json - other",
	}

	allLines, err := result.PolicySetHashes(`.*`)
	Ok(t, err)
	Equals(t, 3, len(allLines))

	failOnly, err := result.PolicySetHashes(`(?m)^FAIL.*`)
	Ok(t, err)
	Equals(t, 2, len(failOnly))
	Equals(t, models.HashPolicyItem("FAIL - plan.json - deny"), failOnly[0])
	Equals(t, models.HashPolicyItem("FAIL - plan.json - other"), failOnly[1])

	noMatch, err := result.PolicySetHashes(`(?m)^WARN.*`)
	Ok(t, err)
	Equals(t, 1, len(noMatch))
	Equals(t, models.HashPolicyItem(result.PolicyOutput), noMatch[0])

	_, err = result.PolicySetHashes(`[invalid`)
	Assert(t, err != nil, "expected error for invalid regex")
}

func TestHashPolicyItem(t *testing.T) {
	h := models.HashPolicyItem("FAIL - plan.json - deny")
	Equals(t, 64, len(h))
	Assert(t, h == models.HashPolicyItem("FAIL - plan.json - deny"), "same input should produce same hash")
	Assert(t, h != models.HashPolicyItem("FAIL - plan.json - other"), "different input should produce different hash")
	Assert(t, models.HashPolicyItem("") != "", "empty string should still produce a hash")
}

func TestPolicySetResult_PolicySetHashes_DefaultRegex(t *testing.T) {
	result := models.PolicySetResult{
		PolicyOutput: "FAIL - plan.json - deny\nOK\nFAIL - other",
	}
	hashes, err := result.PolicySetHashes(`(?s).+`)
	Ok(t, err)
	Equals(t, 1, len(hashes))
	Equals(t, models.HashPolicyItem("FAIL - plan.json - deny\nOK\nFAIL - other"), hashes[0])
}

func TestPolicyCheckResults_PolicyFuncs_WithHashes(t *testing.T) {
	cases := []struct {
		description      string
		policysetResults []models.PolicySetResult
		policyClearedExp bool
		policySummaryExp string
	}{
		{
			description: "approval with matching hashes clears policy",
			policysetResults: []models.PolicySetResult{
				{
					PolicySetName:    "policy1",
					Passed:           false,
					ReqApprovalCount: 1,
					Hashes:           []string{"h1", "h2"},
					Approvals: []models.PolicySetApproval{
						{Approver: "user1", Hashes: []string{"h1", "h2"}},
					},
				},
			},
			policyClearedExp: true,
			policySummaryExp: "policy set: policy1: approved.",
		},
		{
			description: "approval with stale hashes does not clear",
			policysetResults: []models.PolicySetResult{
				{
					PolicySetName:    "policy1",
					Passed:           false,
					ReqApprovalCount: 1,
					Hashes:           []string{"h1", "h2", "h3"},
					Approvals: []models.PolicySetApproval{
						{Approver: "user1", Hashes: []string{"h1", "h2"}},
					},
				},
			},
			policyClearedExp: false,
			policySummaryExp: "policy set: policy1: requires: 1 approval(s), have: 0.",
		},
		{
			description: "mixed: one stale and one fresh among two required",
			policysetResults: []models.PolicySetResult{
				{
					PolicySetName:    "policy1",
					Passed:           false,
					ReqApprovalCount: 2,
					Hashes:           []string{"h1"},
					Approvals: []models.PolicySetApproval{
						{Approver: "user1", Hashes: []string{"h1"}},
						{Approver: "user2", Hashes: []string{"old_h1"}},
					},
				},
			},
			policyClearedExp: false,
			policySummaryExp: "policy set: policy1: requires: 2 approval(s), have: 1.",
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			pcs := models.PolicyCheckResults{
				PolicySetResults: c.policysetResults,
			}
			Equals(t, c.policyClearedExp, pcs.PolicyCleared())
			Equals(t, c.policySummaryExp, pcs.PolicySummary())
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

func TestDiffMarkdownFormattedTerraformOutput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		exp   string
	}{
		{
			"cloudformation stack with heredoc template",
			`Terraform will perform the following actions:

  # aws_cloudformation_stack.example will be created
  + resource "aws_cloudformation_stack" "example" {
      + id            = (known after apply)
      + name          = "my-stack"
      + template_body = <<-EOT
            {
              "AWSTemplateFormatVersion": "2010-09-09",
              "Resources": {
                "MyBucket": {
                  "Type": "AWS::S3::Bucket"
                }
              }
            }
        EOT
    }

Plan: 1 to add, 0 to change, 0 to destroy.`,
			`Terraform will perform the following actions:

  # aws_cloudformation_stack.example will be created
+   resource "aws_cloudformation_stack" "example" {
+       id            = (known after apply)
+       name          = "my-stack"
+       template_body = <<-EOT
            {
              "AWSTemplateFormatVersion": "2010-09-09",
              "Resources": {
                "MyBucket": {
                  "Type": "AWS::S3::Bucket"
                }
              }
            }
        EOT
    }

Plan: 1 to add, 0 to change, 0 to destroy.`,
		},
		{
			"cloudformation stack with parameters and tags",
			`  # aws_cloudformation_stack.example will be updated in-place
  ~ resource "aws_cloudformation_stack" "example" {
        id              = "arn:aws:cloudformation:..."
      ~ name            -> "updated-stack-name"
      ~ parameters      = {
          - "OldParam" = "old-value" -> null
          + "NewParam" = "new-value"
        }
      + tags            = {
          + "Environment" = "production"
        }
        template_body   = <<-EOT
            {
              "AWSTemplateFormatVersion": "2010-09-09"
            }
        EOT
    }`,
			`# aws_cloudformation_stack.example will be updated in-place
!   resource "aws_cloudformation_stack" "example" {
        id              = "arn:aws:cloudformation:..."
!       name            -> "updated-stack-name"
!       parameters      = {
          - "OldParam" = "old-value" -> null
          + "NewParam" = "new-value"
        }
+       tags            = {
          + "Environment" = "production"
        }
        template_body   = <<-EOT
            {
              "AWSTemplateFormatVersion": "2010-09-09"
            }
        EOT
    }`,
		},
		{
			"cloudformation stack with nested attributes",
			`  # aws_cloudformation_stack.example will be created
  + resource "aws_cloudformation_stack" "example" {
      + capabilities  = ["CAPABILITY_IAM"]
      + id            = (known after apply)
      + outputs       = (known after apply)
      + policy_body   = jsonencode({
          + Statement = [
              + {
                  + Effect = "Allow"
                },
            ]
        })
      + template_url  = "https://s3.amazonaws.com/bucket/template.json"
      + timeouts      {
          + create = "30m"
        }
    }`,
			`# aws_cloudformation_stack.example will be created
+   resource "aws_cloudformation_stack" "example" {
+       capabilities  = ["CAPABILITY_IAM"]
+       id            = (known after apply)
+       outputs       = (known after apply)
+       policy_body   = jsonencode({
+           Statement = [
              + {
+                   Effect = "Allow"
                },
            ]
        })
+       template_url  = "https://s3.amazonaws.com/bucket/template.json"
+       timeouts      {
+           create = "30m"
        }
    }`,
		},
		{
			"cloudformation stack with complex template body",
			`  + resource "aws_cloudformation_stack" "vpc" {
      + arn           = (known after apply)
      + name          = "vpc-stack"
      + parameters    = {
          + "CidrBlock" = "10.0.0.0/16"
        }
      + template_body = <<-EOT
            Resources:
              VPC:
                Type: AWS::EC2::VPC
                Properties:
                  CidrBlock: !Ref CidrBlock
                  EnableDnsSupport: true
              InternetGateway:
                Type: AWS::EC2::InternetGateway
        EOT
    }`,
			`+   resource "aws_cloudformation_stack" "vpc" {
+       arn           = (known after apply)
+       name          = "vpc-stack"
+       parameters    = {
          + "CidrBlock" = "10.0.0.0/16"
        }
+       template_body = <<-EOT
            Resources:
              VPC:
                Type: AWS::EC2::VPC
                Properties:
                  CidrBlock: !Ref CidrBlock
                  EnableDnsSupport: true
              InternetGateway:
                Type: AWS::EC2::InternetGateway
        EOT
    }`,
		},
		{
			"multiple resource types with various operators",
			`Terraform will perform the following actions:

  # aws_instance.example will be updated
  ~ resource "aws_instance" "example" {
        id                = "i-1234567890"
      ~ instance_type     -> "t3.medium"
      + monitoring        = true
      - user_data         = "old-data" -> null
        ami                = "ami-12345"
    }

  # aws_s3_bucket.data will be created
  + resource "aws_s3_bucket" "data" {
      + bucket          = "my-bucket"
      + id              = (known after apply)
      + versioning      {
          + enabled = true
        }
    }`,
			`Terraform will perform the following actions:

  # aws_instance.example will be updated
!   resource "aws_instance" "example" {
        id                = "i-1234567890"
!       instance_type     -> "t3.medium"
+       monitoring        = true
-       user_data         = "old-data" -> null
        ami                = "ami-12345"
    }

  # aws_s3_bucket.data will be created
+   resource "aws_s3_bucket" "data" {
+       bucket          = "my-bucket"
+       id              = (known after apply)
+       versioning      {
+           enabled = true
        }
    }`,
		},
		{
			"cloudformation stack with yaml template and fn sub intrinsics",
			`  # aws_cloudformation_stack.iam_role will be created
  + resource "aws_cloudformation_stack" "iam_role" {
      + id            = (known after apply)
      + name          = "audit-role-stack"
      + template_body = <<-EOT
            AWSTemplateFormatVersion: '2010-09-09'
            Description: IAM Role with managed policies
            Resources:
              AuditRole:
                Type: AWS::IAM::Role
                Properties:
                  RoleName: AuditRole
                  AssumeRolePolicyDocument:
                    Version: '2012-10-17'
                    Statement:
                      - Effect: Allow
                        Principal:
                          Service: ec2.amazonaws.com
                        Action: sts:AssumeRole
                  ManagedPolicyArns:
                    - Fn::Sub: arn:${AWS::Partition}:iam::aws:policy/job-function/ViewOnlyAccess
                    - Fn::Sub: arn:${AWS::Partition}:iam::aws:policy/SecurityAudit
        EOT
    }`,
			`# aws_cloudformation_stack.iam_role will be created
+   resource "aws_cloudformation_stack" "iam_role" {
+       id            = (known after apply)
+       name          = "audit-role-stack"
+       template_body = <<-EOT
            AWSTemplateFormatVersion: '2010-09-09'
            Description: IAM Role with managed policies
            Resources:
              AuditRole:
                Type: AWS::IAM::Role
                Properties:
                  RoleName: AuditRole
                  AssumeRolePolicyDocument:
                    Version: '2012-10-17'
                    Statement:
                      - Effect: Allow
                        Principal:
                          Service: ec2.amazonaws.com
                        Action: sts:AssumeRole
                  ManagedPolicyArns:
                    - Fn::Sub: arn:${AWS::Partition}:iam::aws:policy/job-function/ViewOnlyAccess
                    - Fn::Sub: arn:${AWS::Partition}:iam::aws:policy/SecurityAudit
        EOT
    }`,
		},
		{
			// Regression test for https://github.com/runatlantis/atlantis/issues/6419
			// YAML list items using key=value format (e.g. ArgoCD syncOptions) inside a heredoc
			// must not be mistaken for Terraform diff markers.
			"argocd application with yaml key=value list items in heredoc",
			`  # argocd_application.example will be updated in-place
  ~ resource "argocd_application" "example" {
        id   = "my-app"
      ~ metadata {
          ~ resource_version = "12345" -> (known after apply)
        }
        spec = <<-EOT
            destination:
              namespace: default
              server: https://kubernetes.default.svc
            source:
              repoURL: https://github.com/example/repo
            syncPolicy:
              automated:
                prune: true
                selfHeal: true
              syncOptions:
              - ServerSideApply=true
        EOT
    }`,
			`# argocd_application.example will be updated in-place
!   resource "argocd_application" "example" {
        id   = "my-app"
!       metadata {
!           resource_version = "12345" -> (known after apply)
        }
        spec = <<-EOT
            destination:
              namespace: default
              server: https://kubernetes.default.svc
            source:
              repoURL: https://github.com/example/repo
            syncPolicy:
              automated:
                prune: true
                selfHeal: true
              syncOptions:
              - ServerSideApply=true
        EOT
    }`,
		},
		{
			"cloudformation stack with extra-spaced yaml list items in heredoc",
			`  + resource "aws_cloudformation_stack" "conditions" {
      + id            = (known after apply)
      + name          = "conditions-stack"
      + template_body = <<-EOT
            AWSTemplateFormatVersion: '2010-09-09'
            Conditions:
              isOrg:
                Fn::Not:
                  -   Fn::Equals:
                        -   !Ref ManagementAccountId
                        -   ""
              isProd:
                Fn::Equals:
                  -   !Ref Environment
                  -   production
        EOT
    }`,
			`+   resource "aws_cloudformation_stack" "conditions" {
+       id            = (known after apply)
+       name          = "conditions-stack"
+       template_body = <<-EOT
            AWSTemplateFormatVersion: '2010-09-09'
            Conditions:
              isOrg:
                Fn::Not:
                  -   Fn::Equals:
                        -   !Ref ManagementAccountId
                        -   ""
              isProd:
                Fn::Equals:
                  -   !Ref Environment
                  -   production
        EOT
    }`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := models.PlanSuccess{
				TerraformOutput: tt.input,
			}
			result := ps.DiffMarkdownFormattedTerraformOutput()
			Equals(t, tt.exp, result)
		})
	}
}
