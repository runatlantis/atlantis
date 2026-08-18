// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package models_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

func TestRemediationResultCompletedAtJSON(t *testing.T) {
	result := models.NewRemediationResult("id", "owner/repo", "main", models.RemediationPlanOnly)

	body, err := json.Marshal(result)
	Ok(t, err)
	Assert(t, !strings.Contains(string(body), "completed_at"), "expected incomplete remediation to omit completed_at: %s", body)

	result.Complete()
	body, err = json.Marshal(result)
	Ok(t, err)
	Assert(t, strings.Contains(string(body), "completed_at"), "expected completed remediation to include completed_at: %s", body)
}

func TestRemediationRequestValidateRequiresBaseBranchForSHAAndTagRefs(t *testing.T) {
	for _, c := range []struct {
		name string
		ref  string
	}{
		{name: "sha", ref: "0123456789abcdef0123456789abcdef01234567"},
		{name: "tag ref", ref: "refs/tags/v1.0.0"},
		{name: "tag ref without semver", ref: "refs/tags/prod"},
		{name: "bare semver tag", ref: "v1.0.0"},
		{name: "ambiguous prod ref", ref: "prod"},
		{name: "ambiguous latest ref", ref: "latest"},
		{name: "ambiguous stable ref", ref: "stable"},
	} {
		t.Run(c.name, func(t *testing.T) {
			request := models.RemediationRequest{
				Repository: "owner/repo",
				Ref:        c.ref,
				Type:       "Github",
			}
			errs := request.Validate()
			Assert(t, len(errs) > 0, "expected validation error")
			Equals(t, "base_branch", errs[0].Field)

			request.BaseBranch = "main"
			Equals(t, 0, len(request.Validate()))
		})
	}
}

func TestRemediationRequestValidateBranchRefDoesNotRequireBaseBranch(t *testing.T) {
	for _, ref := range []string{"main", "master", "develop", "feature/foo", "release/2026-06", "refs/heads/release/1.2", "refs/heads/feature/foo"} {
		t.Run(ref, func(t *testing.T) {
			request := models.RemediationRequest{
				Repository: "owner/repo",
				Ref:        ref,
				Type:       "Github",
			}

			Equals(t, 0, len(request.Validate()))
		})
	}
}

func TestRemediationRequestValidateRejectsMalformedBaseBranch(t *testing.T) {
	for _, baseBranch := range []string{
		"refs/pull/1/head",
		"pull/1/head",
		"refs/merge-requests/1/head",
		"refs/tags/prod",
		"-main",
		"--upload-pack=/tmp/x",
		"foo:bar",
		"   ",
	} {
		t.Run(baseBranch, func(t *testing.T) {
			request := models.RemediationRequest{
				Repository: "owner/repo",
				Ref:        "main",
				BaseBranch: baseBranch,
				Type:       "Github",
			}
			errs := request.Validate()
			Assert(t, len(errs) > 0, "expected validation error")
			Equals(t, "base_branch", errs[0].Field)
		})
	}

	request := models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		BaseBranch: "release/1.2",
		Type:       "Github",
	}
	Equals(t, 0, len(request.Validate()))
}

func TestRemediationRequestValidateRejectsUnsafeRefs(t *testing.T) {
	for _, ref := range []string{
		"refs/pull/123/head",
		"refs/pull/123/merge",
		"pull/123/head",
		"pull/123/merge",
		"+refs/pull/123/head",
		"refs/merge-requests/123/head",
		"refs/merge-requests/123/merge",
		"merge-requests/123/head",
		"merge-requests/123/merge",
		"+refs/merge-requests/123/head",
		"main:refs/tmp/main",
		"--upload-pack=/tmp/x",
	} {
		t.Run(ref, func(t *testing.T) {
			request := models.RemediationRequest{
				Repository: "owner/repo",
				Ref:        ref,
				BaseBranch: "main",
				Type:       "Github",
			}
			errs := request.Validate()
			Assert(t, len(errs) > 0, "expected validation error")
			Equals(t, "ref", errs[0].Field)
		})
	}
}

func TestRemediationRequestValidateRejectsInvalidWorkspaces(t *testing.T) {
	for _, workspace := range []string{"../../tmp/plan", "prod/stage", "..", ".", " prod", "prod "} {
		t.Run("path "+workspace, func(t *testing.T) {
			request := models.RemediationRequest{
				Repository: "owner/repo",
				Ref:        "main",
				Type:       "Github",
				Paths:      []models.DriftDetectionPath{{Directory: "env", Workspace: workspace}},
			}
			errs := request.Validate()
			Assert(t, len(errs) > 0, "expected validation error")
			Equals(t, "paths", errs[0].Field)
		})

		t.Run("top-level "+workspace, func(t *testing.T) {
			request := models.RemediationRequest{
				Repository: "owner/repo",
				Ref:        "main",
				Type:       "Github",
				Workspaces: []string{workspace},
			}
			errs := request.Validate()
			Assert(t, len(errs) > 0, "expected validation error")
			Equals(t, "workspaces", errs[0].Field)
		})
	}

	request := models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
		Paths:      []models.DriftDetectionPath{{Directory: "env", Workspace: "prod"}},
		Workspaces: []string{"prod"},
	}
	Equals(t, 0, len(request.Validate()))
}

func TestRemediationRequestValidateRejectsRegexProjectSelectors(t *testing.T) {
	for _, project := range []string{"app-.*", "^app$", "app[0-9]", "app|db"} {
		t.Run(project, func(t *testing.T) {
			request := models.RemediationRequest{
				Repository: "owner/repo",
				Ref:        "main",
				Type:       "Github",
				Projects:   []string{project},
			}
			errs := request.Validate()
			Assert(t, len(errs) > 0, "expected validation error")
			Equals(t, "projects", errs[0].Field)
		})
	}

	request := models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
		Projects:   []string{"app.prod"},
	}
	Equals(t, 0, len(request.Validate()))
}

func TestRemediationRequestValidateRejectsUnsupportedVCSTypes(t *testing.T) {
	for _, vcsType := range []string{"BitbucketCloud", "BitbucketServer", "AzureDevops"} {
		t.Run(vcsType, func(t *testing.T) {
			request := models.RemediationRequest{
				Repository: "owner/repo",
				Ref:        "main",
				Type:       vcsType,
			}

			errs := request.Validate()
			Assert(t, len(errs) > 0, "expected validation error")
			Equals(t, "type", errs[0].Field)
			Assert(t, strings.Contains(errs[0].Message, "Github, Gitlab, Gitea"), "unexpected message: %q", errs[0].Message)
		})
	}
	for _, vcsType := range []string{"Github", "Gitlab", "Gitea"} {
		t.Run(vcsType, func(t *testing.T) {
			request := models.RemediationRequest{
				Repository: "owner/repo",
				Ref:        "main",
				Type:       vcsType,
			}

			Equals(t, 0, len(request.Validate()))
		})
	}
}

func TestRemediationRequestValidateRejectsEmptyPathSelectors(t *testing.T) {
	for _, directory := range []string{"   ", "../env", "/env"} {
		request := models.RemediationRequest{
			Repository: "owner/repo",
			Ref:        "main",
			Type:       "Github",
			Paths:      []models.DriftDetectionPath{{Directory: directory}},
		}

		errs := request.Validate()
		Assert(t, len(errs) > 0, "expected validation error")
		Equals(t, "paths", errs[0].Field)
	}
}

func TestRemediationRequestValidateRejectsConflictingPathAndTopLevelWorkspaces(t *testing.T) {
	request := models.RemediationRequest{
		Repository: "owner/repo",
		Ref:        "main",
		Type:       "Github",
		Paths: []models.DriftDetectionPath{{
			Directory: "env",
			Workspace: "dev",
		}},
		Workspaces: []string{"prod"},
	}

	errs := request.Validate()
	Assert(t, len(errs) > 0, "expected validation error")
	Equals(t, "paths", errs[0].Field)
}

func TestRemediationRequestValidateRejectsGlobPathSelectors(t *testing.T) {
	for _, directory := range []string{"envs/*", "envs/pro?", "envs/[prod]", "envs/{prod}"} {
		t.Run(directory, func(t *testing.T) {
			request := models.RemediationRequest{
				Repository: "owner/repo",
				Ref:        "main",
				Type:       "Github",
				Paths:      []models.DriftDetectionPath{{Directory: directory}},
			}
			errs := request.Validate()
			Assert(t, len(errs) > 0, "expected validation error")
			Equals(t, "paths", errs[0].Field)
		})
	}
}

func TestRemediationRequestValidateRejectsMalformedRepositories(t *testing.T) {
	for _, repository := range []string{"owner", "/repo", "owner/", "owner//repo", "owner/../repo", "owner repo/name"} {
		t.Run(repository, func(t *testing.T) {
			request := models.RemediationRequest{
				Repository: repository,
				Ref:        "main",
				Type:       "Github",
			}
			errs := request.Validate()
			Assert(t, len(errs) > 0, "expected validation error")
			Equals(t, "repository", errs[0].Field)
		})
	}

	github := models.RemediationRequest{Repository: "owner/repo/extra", Ref: "main", Type: "Github"}
	errs := github.Validate()
	Assert(t, len(errs) > 0, "expected GitHub repository validation error")
	Equals(t, "repository", errs[0].Field)

	gitlab := models.RemediationRequest{Repository: "group/subgroup/repo", Ref: "main", Type: "Gitlab"}
	Equals(t, 0, len(gitlab.Validate()))
}
