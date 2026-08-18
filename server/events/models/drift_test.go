// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package models_test

import (
	"strings"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

func TestNewDriftSummaryFromPlanStats(t *testing.T) {
	cases := []struct {
		name     string
		stats    models.PlanSuccessStats
		summary  string
		expected models.DriftSummary
	}{
		{
			name: "no drift",
			stats: models.PlanSuccessStats{
				Add:     0,
				Change:  0,
				Destroy: 0,
				Import:  0,
			},
			summary: "No changes",
			expected: models.DriftSummary{
				HasDrift:  false,
				ToAdd:     0,
				ToChange:  0,
				ToDestroy: 0,
				ToImport:  0,
				Summary:   "No changes",
			},
		},
		{
			name: "with additions",
			stats: models.PlanSuccessStats{
				Add:     3,
				Change:  0,
				Destroy: 0,
				Import:  0,
			},
			summary: "Plan: 3 to add",
			expected: models.DriftSummary{
				HasDrift:  true,
				ToAdd:     3,
				ToChange:  0,
				ToDestroy: 0,
				ToImport:  0,
				Summary:   "Plan: 3 to add",
			},
		},
		{
			name: "with changes",
			stats: models.PlanSuccessStats{
				Add:     0,
				Change:  5,
				Destroy: 0,
				Import:  0,
			},
			summary: "Plan: 5 to change",
			expected: models.DriftSummary{
				HasDrift:  true,
				ToAdd:     0,
				ToChange:  5,
				ToDestroy: 0,
				ToImport:  0,
				Summary:   "Plan: 5 to change",
			},
		},
		{
			name: "with destroys",
			stats: models.PlanSuccessStats{
				Add:     0,
				Change:  0,
				Destroy: 2,
				Import:  0,
			},
			summary: "Plan: 2 to destroy",
			expected: models.DriftSummary{
				HasDrift:  true,
				ToAdd:     0,
				ToChange:  0,
				ToDestroy: 2,
				ToImport:  0,
				Summary:   "Plan: 2 to destroy",
			},
		},
		{
			name: "with imports",
			stats: models.PlanSuccessStats{
				Add:     0,
				Change:  0,
				Destroy: 0,
				Import:  1,
			},
			summary: "Plan: 1 to import",
			expected: models.DriftSummary{
				HasDrift:  true,
				ToAdd:     0,
				ToChange:  0,
				ToDestroy: 0,
				ToImport:  1,
				Summary:   "Plan: 1 to import",
			},
		},
		{
			name: "with forgets",
			stats: models.PlanSuccessStats{
				Add:     0,
				Change:  0,
				Destroy: 0,
				Import:  0,
				Forget:  2,
			},
			summary: "Plan: 0 to add, 0 to change, 0 to destroy, 2 to forget",
			expected: models.DriftSummary{
				HasDrift:  true,
				ToAdd:     0,
				ToChange:  0,
				ToDestroy: 0,
				ToImport:  0,
				ToForget:  2,
				Summary:   "Plan: 0 to add, 0 to change, 0 to destroy, 2 to forget",
			},
		},
		{
			name: "with changes outside terraform",
			stats: models.PlanSuccessStats{
				Add:            0,
				Change:         1,
				Destroy:        0,
				Import:         0,
				ChangesOutside: true,
			},
			summary: "Plan: 1 to change (changes outside Terraform)",
			expected: models.DriftSummary{
				HasDrift:       true,
				ToAdd:          0,
				ToChange:       1,
				ToDestroy:      0,
				ToImport:       0,
				Summary:        "Plan: 1 to change (changes outside Terraform)",
				ChangesOutside: true,
			},
		},
		{
			name: "with only changes outside terraform",
			stats: models.PlanSuccessStats{
				ChangesOutside: true,
			},
			summary: "Objects have changed outside of Terraform",
			expected: models.DriftSummary{
				HasDrift:       true,
				Summary:        "Objects have changed outside of Terraform",
				ChangesOutside: true,
			},
		},
		{
			name: "mixed drift",
			stats: models.PlanSuccessStats{
				Add:     2,
				Change:  3,
				Destroy: 1,
				Import:  1,
				Forget:  1,
			},
			summary: "Plan: 1 to import, 2 to add, 3 to change, 1 to destroy, 1 to forget",
			expected: models.DriftSummary{
				HasDrift:  true,
				ToAdd:     2,
				ToChange:  3,
				ToDestroy: 1,
				ToImport:  1,
				ToForget:  1,
				Summary:   "Plan: 1 to import, 2 to add, 3 to change, 1 to destroy, 1 to forget",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := models.NewDriftSummaryFromPlanStats(c.stats, c.summary)
			Equals(t, c.expected, result)
		})
	}
}

func TestNewDriftStatusResponseUsesMostRecentLastChecked(t *testing.T) {
	oldCheck := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	recentCheck := oldCheck.Add(2 * time.Hour)

	response := models.NewDriftStatusResponse("owner/repo", []models.ProjectDrift{
		{ProjectName: "old", LastChecked: oldCheck},
		{ProjectName: "recent", LastChecked: recentCheck},
	})

	Equals(t, recentCheck, response.CheckedAt)
}

func TestNewDriftStatusResponseEmptyHasZeroCheckedAt(t *testing.T) {
	response := models.NewDriftStatusResponse("owner/repo", nil)

	Assert(t, response.CheckedAt.IsZero(), "expected empty status response to have zero checked_at")
}

func TestNewDriftSummaryFromPlanSuccess_Nil(t *testing.T) {
	result := models.NewDriftSummaryFromPlanSuccess(nil)
	Equals(t, models.DriftSummary{}, result)
}

func TestNewDriftSummaryFromPlanSuccess_OutsideOnlyDriftIncludesNote(t *testing.T) {
	result := models.NewDriftSummaryFromPlanSuccess(&models.PlanSuccess{
		TerraformOutput: "Note: Objects have changed outside of Terraform\ndummy\nNo changes. Infrastructure is up-to-date.",
	})

	Equals(t, true, result.HasDrift)
	Equals(t, true, result.ChangesOutside)
	Equals(t, 0, result.TotalChanges())
	Equals(t, "\n**Note: Objects have changed outside of Terraform**\nNo changes. Infrastructure is up-to-date.", result.Summary)
}

func TestNewDriftStatusResponse(t *testing.T) {
	projects := []models.ProjectDrift{
		{
			ProjectName: "project1",
			Path:        "path1",
			Workspace:   "default",
			Ref:         "main",
			Drift: models.DriftSummary{
				HasDrift: true,
				ToAdd:    2,
			},
			LastChecked: time.Now(),
		},
		{
			ProjectName: "project2",
			Path:        "path2",
			Workspace:   "default",
			Ref:         "main",
			Drift: models.DriftSummary{
				HasDrift: false,
			},
			LastChecked: time.Now(),
		},
		{
			ProjectName: "project3",
			Path:        "path3",
			Workspace:   "staging",
			Ref:         "main",
			Drift: models.DriftSummary{
				HasDrift: true,
				ToChange: 1,
			},
			LastChecked: time.Now(),
		},
	}

	result := models.NewDriftStatusResponse("owner/repo", projects)

	Equals(t, "owner/repo", result.Repository)
	Equals(t, 3, result.TotalProjects)
	Equals(t, 2, result.ProjectsWithDrift)
	Equals(t, projects, result.Projects)
	Assert(t, !result.CheckedAt.IsZero(), "CheckedAt should be set")
}

func TestNewDriftStatusResponse_Empty(t *testing.T) {
	result := models.NewDriftStatusResponse("owner/repo", []models.ProjectDrift{})

	Equals(t, "owner/repo", result.Repository)
	Equals(t, 0, result.TotalProjects)
	Equals(t, 0, result.ProjectsWithDrift)
}

func TestDriftDetectionRequestValidateRejectsEmptySelectors(t *testing.T) {
	for _, c := range []struct {
		name    string
		request models.DriftDetectionRequest
		field   string
	}{
		{
			name: "empty project",
			request: models.DriftDetectionRequest{
				Repository: "owner/repo",
				Ref:        "main",
				Type:       "Github",
				Projects:   []string{""},
			},
			field: "projects",
		},
		{
			name: "space-only project",
			request: models.DriftDetectionRequest{
				Repository: "owner/repo",
				Ref:        "main",
				Type:       "Github",
				Projects:   []string{"   "},
			},
			field: "projects",
		},
		{
			name: "empty path",
			request: models.DriftDetectionRequest{
				Repository: "owner/repo",
				Ref:        "main",
				Type:       "Github",
				Paths:      []models.DriftDetectionPath{{Directory: ""}},
			},
			field: "paths",
		},
		{
			name: "space-only path",
			request: models.DriftDetectionRequest{
				Repository: "owner/repo",
				Ref:        "main",
				Type:       "Github",
				Paths:      []models.DriftDetectionPath{{Directory: "   "}},
			},
			field: "paths",
		},
		{
			name: "parent path",
			request: models.DriftDetectionRequest{
				Repository: "owner/repo",
				Ref:        "main",
				Type:       "Github",
				Paths:      []models.DriftDetectionPath{{Directory: "../env"}},
			},
			field: "paths",
		},
		{
			name: "absolute path",
			request: models.DriftDetectionRequest{
				Repository: "owner/repo",
				Ref:        "main",
				Type:       "Github",
				Paths:      []models.DriftDetectionPath{{Directory: "/env"}},
			},
			field: "paths",
		},
		{
			name: "mixed projects and paths",
			request: models.DriftDetectionRequest{
				Repository: "owner/repo",
				Ref:        "main",
				Type:       "Github",
				Projects:   []string{"app"},
				Paths:      []models.DriftDetectionPath{{Directory: "env"}},
			},
			field: "paths",
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			errs := c.request.Validate()
			Assert(t, len(errs) > 0, "expected validation error")
			Equals(t, c.field, errs[0].Field)
		})
	}
}

func TestDriftDetectionRequestValidateRejectsUnsupportedVCSTypes(t *testing.T) {
	for _, vcsType := range []string{"BitbucketCloud", "BitbucketServer", "AzureDevops"} {
		t.Run(vcsType, func(t *testing.T) {
			request := models.DriftDetectionRequest{
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
			request := models.DriftDetectionRequest{
				Repository: "owner/repo",
				Ref:        "main",
				Type:       vcsType,
			}

			Equals(t, 0, len(request.Validate()))
		})
	}
}

func TestDriftDetectionRequestValidateRequiresBaseBranchForSHAAndTagRefs(t *testing.T) {
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
			request := models.DriftDetectionRequest{
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

func TestDriftDetectionRequestValidateBranchRefDoesNotRequireBaseBranch(t *testing.T) {
	for _, ref := range []string{"main", "master", "develop", "feature/foo", "release/2026-06", "refs/heads/release/1.2", "refs/heads/feature/foo"} {
		t.Run(ref, func(t *testing.T) {
			request := models.DriftDetectionRequest{
				Repository: "owner/repo",
				Ref:        ref,
				Type:       "Github",
			}

			Equals(t, 0, len(request.Validate()))
		})
	}
}

func TestDriftDetectionRequestValidateRejectsMalformedBaseBranch(t *testing.T) {
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
			request := models.DriftDetectionRequest{
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

	request := models.DriftDetectionRequest{
		Repository: "owner/repo",
		Ref:        "main",
		BaseBranch: "release/1.2",
		Type:       "Github",
	}
	Equals(t, 0, len(request.Validate()))
}

func TestDriftDetectionRequestValidateRejectsUnsafeRefs(t *testing.T) {
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
			request := models.DriftDetectionRequest{
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

func TestDriftDetectionRequestValidateRejectsInvalidPathWorkspaces(t *testing.T) {
	for _, workspace := range []string{"../../tmp/plan", "prod/stage", "..", ".", " prod", "prod "} {
		t.Run(workspace, func(t *testing.T) {
			request := models.DriftDetectionRequest{
				Repository: "owner/repo",
				Ref:        "main",
				Type:       "Github",
				Paths:      []models.DriftDetectionPath{{Directory: "env", Workspace: workspace}},
			}
			errs := request.Validate()
			Assert(t, len(errs) > 0, "expected validation error")
			Equals(t, "paths", errs[0].Field)
		})
	}

	for _, workspace := range []string{"default", "prod", "dev-us-east-1"} {
		t.Run("valid "+workspace, func(t *testing.T) {
			request := models.DriftDetectionRequest{
				Repository: "owner/repo",
				Ref:        "main",
				Type:       "Github",
				Paths:      []models.DriftDetectionPath{{Directory: "env", Workspace: workspace}},
			}
			Equals(t, 0, len(request.Validate()))
		})
	}
}

func TestDriftDetectionRequestValidateRejectsGlobPathSelectors(t *testing.T) {
	for _, directory := range []string{"envs/*", "envs/pro?", "envs/[prod]", "envs/{prod}"} {
		t.Run(directory, func(t *testing.T) {
			request := models.DriftDetectionRequest{
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

func TestDriftDetectionRequestValidateRejectsMalformedRepositories(t *testing.T) {
	for _, repository := range []string{"owner", "/repo", "owner/", "owner//repo", "owner/../repo", "owner repo/name"} {
		t.Run(repository, func(t *testing.T) {
			request := models.DriftDetectionRequest{
				Repository: repository,
				Ref:        "main",
				Type:       "Github",
			}
			errs := request.Validate()
			Assert(t, len(errs) > 0, "expected validation error")
			Equals(t, "repository", errs[0].Field)
		})
	}

	github := models.DriftDetectionRequest{Repository: "owner/repo/extra", Ref: "main", Type: "Github"}
	errs := github.Validate()
	Assert(t, len(errs) > 0, "expected GitHub repository validation error")
	Equals(t, "repository", errs[0].Field)

	gitlab := models.DriftDetectionRequest{Repository: "group/subgroup/repo", Ref: "main", Type: "Gitlab"}
	Equals(t, 0, len(gitlab.Validate()))
}
