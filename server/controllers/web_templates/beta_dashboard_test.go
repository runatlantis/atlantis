// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package web_templates_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/controllers/web_templates"
	"github.com/runatlantis/atlantis/server/jobs"
	. "github.com/runatlantis/atlantis/testing"
)

func TestGroupBetaDashboard(t *testing.T) {
	now := time.Now()
	data := web_templates.IndexData{
		Locks: []web_templates.LockIndexData{
			{RepoFullName: "a/repo", PullNum: 9, Workspace: "production", Path: "storage", LockedBy: "old", Time: now.Add(-time.Hour), LockPath: "/lock?id=storage"},
			{RepoFullName: "a/repo", PullNum: 9, Workspace: "staging", Path: "network", LockedBy: "owner", Time: now, LockPath: "/lock?id=network"},
			{RepoFullName: "b/repo", PullNum: 9, Workspace: "default", Path: "iam", LockedBy: "other"},
		},
		PullToJobMapping: []jobs.PullInfoWithJobIDs{
			{Pull: jobs.PullInfo{RepoFullName: "a/repo", PullNum: 9, Workspace: "production", Path: "storage", ProjectName: "storage"}, JobIDInfos: []jobs.JobIDInfo{
				{JobID: "old", JobIDUrl: "/jobs/old", Time: now.Add(-time.Hour)},
				{JobID: "new", JobIDUrl: "/jobs/new", Time: now},
			}},
			{Pull: jobs.PullInfo{RepoFullName: "a/repo", PullNum: 9, Workspace: "production", Path: "database", ProjectName: "database"}},
			{Pull: jobs.PullInfo{RepoFullName: "a/repo", PullNum: 10}, JobIDInfos: []jobs.JobIDInfo{{JobID: "hook", JobIDUrl: "/jobs/hook"}}},
			{Pull: jobs.PullInfo{}, JobIDInfos: []jobs.JobIDInfo{{JobID: "untracked-hook"}}},
		},
	}
	grouped := web_templates.GroupBetaDashboard(data)
	Equals(t, 3, grouped.Total) // Same number in another repository is another request.
	Equals(t, 3, len(grouped.Repositories))
	Equals(t, "", grouped.Repositories[0].Name)
	Equals(t, "", grouped.Repositories[0].Workspaces[0].Pulls[0].ID) // Hooks aren't counted as requests.
	repo := grouped.Repositories[1]
	Equals(t, "a/repo", repo.Name)
	Equals(t, "", repo.Workspaces[0].Name) // Missing metadata isn't assigned to default.
	Equals(t, "production", repo.Workspaces[1].Name)
	Equals(t, "staging", repo.Workspaces[2].Name)
	production := repo.Workspaces[1].Pulls[0]
	staging := repo.Workspaces[2].Pulls[0]
	Equals(t, production.ID, staging.ID)
	Equals(t, "owner", production.Author)
	Equals(t, 2, len(production.Projects)) // The matching lock doesn't duplicate a project.
	Equals(t, "database", production.Projects[0].Name)
	Equals(t, "/jobs/new", production.Projects[1].Jobs[0].JobIDUrl)
	Equals(t, "/jobs/old", production.Projects[1].Jobs[1].JobIDUrl)
	Equals(t, "old", data.PullToJobMapping[0].JobIDInfos[0].JobID) // Input remains unchanged.
	Equals(t, "/lock?id=storage", grouped.Locks[0].LockPath)
	for i := 0; i < 20; i++ {
		Equals(t, grouped, web_templates.GroupBetaDashboard(data))
	}
	Equals(t, 0, web_templates.GroupBetaDashboard(web_templates.IndexData{}).Total)
}

func TestBetaDashboardPullRequestURLs(t *testing.T) {
	now := time.Now()
	githubURL := "https://github.example.com/org/repo/pull/1"
	gitlabURL := "https://gitlab.example.com/group/repo/-/merge_requests/2?view=one&mode=two"
	data := web_templates.IndexData{
		CleanedBasePath: "/atlantis",
		Locks: []web_templates.LockIndexData{
			{RepoFullName: "org/repo", PullNum: 1, Workspace: "prod", Path: "infra", LockPath: "/lock?id=lock-id", PullRequestURL: githubURL, Time: now},
			{RepoFullName: "org/repo", PullNum: 1, Workspace: "staging", Path: "infra"},
		},
		PullToJobMapping: []jobs.PullInfoWithJobIDs{
			{Pull: jobs.PullInfo{RepoFullName: "org/repo", PullNum: 1, Workspace: "prod", Path: "infra"}},
			{Pull: jobs.PullInfo{RepoFullName: "group/repo", PullNum: 2}, JobIDInfos: []jobs.JobIDInfo{
				{PullRequestURL: "https://gitlab.example.com/old-link", Time: now.Add(-time.Hour)},
				{PullRequestURL: gitlabURL, Time: now},
			}},
			{Pull: jobs.PullInfo{RepoFullName: "no-url/repo", PullNum: 3}},
			{Pull: jobs.PullInfo{RepoFullName: "unsafe/repo", PullNum: 4}, JobIDInfos: []jobs.JobIDInfo{{PullRequestURL: "javascript:alert(1)"}}},
		},
	}
	grouped := web_templates.GroupBetaDashboard(data)
	Equals(t, gitlabURL, grouped.Repositories[0].Workspaces[0].Pulls[0].URL)
	Equals(t, "", grouped.Repositories[1].Workspaces[0].Pulls[0].URL)
	Equals(t, githubURL, grouped.Repositories[2].Workspaces[0].Pulls[0].URL)
	Equals(t, githubURL, grouped.Repositories[2].Workspaces[1].Pulls[0].URL)
	var output bytes.Buffer
	Ok(t, web_templates.BetaDashboardTemplate.Execute(&output, data))
	html := output.String()
	for _, expected := range []string{
		`href="` + githubURL + `"`,
		`href="https://gitlab.example.com/group/repo/-/merge_requests/2?view=one&amp;mode=two"`,
		`href="/atlantis/lock?id=lock-id"`,
		`Open pull / merge request #1`,
		`<h4 class="pull-title">Pull / merge request <span class="pull-number">#3</span></h4>`,
	} {
		Assert(t, strings.Contains(html, expected), "missing %q", expected)
	}
	Assert(t, !strings.Contains(html, `href="/atlantis/https://`), "external request URLs must not get an Atlantis base path")
	Assert(t, !strings.Contains(html, `href="javascript:`), "unsafe URL protocols must be escaped")
}

func TestBetaDashboardTemplate(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		for _, locked := range []bool{false, true} {
			data := web_templates.IndexData{
				CleanedBasePath: "/atlantis", AtlantisVersion: "test-version",
				ApplyLock: web_templates.ApplyLockData{GlobalApplyLockEnabled: enabled, Locked: locked},
				Locks:     []web_templates.LockIndexData{{RepoFullName: "org/<script>", Workspace: "prod", PullNum: 1, LockedBy: "<img onerror=alert(1)>", LockPath: "/lock?id=encoded%252Fid"}},
				PullToJobMapping: []jobs.PullInfoWithJobIDs{
					{Pull: jobs.PullInfo{RepoFullName: "org/<script>", PullNum: 1}, JobIDInfos: []jobs.JobIDInfo{{JobIDUrl: "/jobs/job-id", JobDescription: "<script>alert(1)</script>"}}},
					{Pull: jobs.PullInfo{RepoFullName: "org/no-lock", PullNum: 2}},
				},
			}
			var output bytes.Buffer
			Ok(t, web_templates.BetaDashboardTemplate.Execute(&output, data))
			html := output.String()
			for _, expected := range []string{"/atlantis/static/css/beta-dashboard.css", "/atlantis/static/js/beta-dashboard-filters.js", `/atlantis/jobs/job-id`, `/atlantis/lock?id=encoded%252Fid`, "Workspace unavailable", "User unavailable", "test-version", "&lt;script&gt;"} {
				Assert(t, strings.Contains(html, expected), "missing %q", expected)
			}
			Assert(t, !strings.Contains(html, "<script>alert(1)</script>"), "metadata must be escaped")
			Assert(t, !strings.Contains(html, "data-status="), "must not fabricate operation results")
			Equals(t, enabled, strings.Contains(html, `id="apply-toggle"`))
			Equals(t, enabled, strings.Contains(html, `id="apply-dialog"`))
			Equals(t, enabled, strings.Contains(html, `href="#apply-controls"`))
			if enabled {
				endpoint := "/atlantis/apply/lock"
				if locked {
					endpoint = "/atlantis/apply/unlock"
				}
				Assert(t, strings.Contains(html, `data-endpoint="`+endpoint+`"`), "apply endpoint must honor base path")
			}
		}
	}
	var empty bytes.Buffer
	Ok(t, web_templates.BetaDashboardTemplate.Execute(&empty, web_templates.IndexData{}))
	Assert(t, strings.Contains(empty.String(), "No pull requests"), "must render request empty state")
	Assert(t, strings.Contains(empty.String(), "No locks found."), "must render lock empty state")
	Assert(t, web_templates.BetaDashboardTemplate.Execute(&empty, nil) != nil, "invalid data must return an error")
}

func TestClassicDashboardBetaLink(t *testing.T) {
	for _, basePath := range []string{"", "/atlantis"} {
		var output bytes.Buffer
		Ok(t, web_templates.IndexTemplate.Execute(&output, web_templates.IndexData{CleanedBasePath: basePath}))
		Assert(t, strings.Contains(output.String(), `href="`+basePath+`/beta">Try the beta dashboard</a>`), "classic beta link must honor the base path")
	}
}
