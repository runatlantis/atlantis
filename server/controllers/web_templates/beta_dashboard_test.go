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
			{RepoFullName: "lock-only/repo", PullNum: 1, Workspace: "default", Path: "infra", LockedBy: "lock-owner"},
		},
		PullToJobMapping: []jobs.PullInfoWithJobIDs{
			{Pull: jobs.PullInfo{RepoFullName: "a/repo", PullNum: 9, Workspace: "production", Path: "storage", ProjectName: "storage"}, JobIDInfos: []jobs.JobIDInfo{
				{JobID: "old", JobIDUrl: "/jobs/old", Time: now.Add(-time.Hour)},
				{JobID: "new", JobIDUrl: "/jobs/new", Time: now},
			}},
			{Pull: jobs.PullInfo{RepoFullName: "a/repo", PullNum: 9, Workspace: "production", Path: "database", ProjectName: "database"}},
			{Pull: jobs.PullInfo{RepoFullName: "a/repo", PullNum: 9, Workspace: "staging", Path: "network"}, JobIDInfos: []jobs.JobIDInfo{{JobID: "staging", JobIDUrl: "/jobs/staging"}}},
			{Pull: jobs.PullInfo{RepoFullName: "b/repo", PullNum: 9, Workspace: "default", Path: "iam"}, JobIDInfos: []jobs.JobIDInfo{{JobID: "other", JobIDUrl: "/jobs/other"}}},
			{Pull: jobs.PullInfo{RepoFullName: "empty/repo", PullNum: 1}},
			{Pull: jobs.PullInfo{RepoFullName: "a/repo", PullNum: 10}, JobIDInfos: []jobs.JobIDInfo{{JobID: "hook", JobIDUrl: "/jobs/hook"}}},
			{Pull: jobs.PullInfo{}, JobIDInfos: []jobs.JobIDInfo{{JobID: "untracked-hook"}}},
		},
	}
	grouped := web_templates.GroupBetaDashboard(data)
	Equals(t, 6, grouped.Total) // Counts jobs, including repeated requests and workflow hooks.
	Equals(t, 3, len(grouped.Repositories))
	Equals(t, "", grouped.Repositories[0].Name)
	Equals(t, "", grouped.Repositories[0].Workspaces[0].Pulls[0].ID) // Hook context isn't a request.
	repo := grouped.Repositories[1]
	Equals(t, "a/repo", repo.Name)
	Equals(t, "", repo.Workspaces[0].Name) // Missing metadata isn't assigned to default.
	Equals(t, "production", repo.Workspaces[1].Name)
	Equals(t, "staging", repo.Workspaces[2].Name)
	production := repo.Workspaces[1].Pulls[0]
	staging := repo.Workspaces[2].Pulls[0]
	Equals(t, production.ID, staging.ID)
	Equals(t, "owner", production.Author)
	Equals(t, 1, len(production.Projects)) // Locks and empty mappings don't create projects.
	Equals(t, "storage", production.Projects[0].Name)
	Equals(t, "/jobs/new", production.Projects[0].Jobs[0].JobIDUrl)
	Equals(t, "/jobs/old", production.Projects[0].Jobs[1].JobIDUrl)
	Equals(t, 2, repo.Workspaces[1].JobCount())
	Equals(t, 1, repo.Workspaces[2].JobCount())
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
			{Pull: jobs.PullInfo{RepoFullName: "org/repo", PullNum: 1, Workspace: "prod", Path: "infra"}, JobIDInfos: []jobs.JobIDInfo{{JobID: "github", JobIDUrl: "/jobs/github"}}},
			{Pull: jobs.PullInfo{RepoFullName: "group/repo", PullNum: 2}, JobIDInfos: []jobs.JobIDInfo{
				{PullRequestURL: "https://gitlab.example.com/old-link", Time: now.Add(-time.Hour)},
				{PullRequestURL: gitlabURL, Time: now},
			}},
			{Pull: jobs.PullInfo{RepoFullName: "no-url/repo", PullNum: 3}, JobIDInfos: []jobs.JobIDInfo{{JobID: "no-url", JobIDUrl: "/jobs/no-url"}}},
			{Pull: jobs.PullInfo{RepoFullName: "unsafe/repo", PullNum: 4}, JobIDInfos: []jobs.JobIDInfo{{PullRequestURL: "javascript:alert(1)"}}},
		},
	}
	grouped := web_templates.GroupBetaDashboard(data)
	Equals(t, gitlabURL, grouped.Repositories[0].Workspaces[0].Pulls[0].URL)
	Equals(t, "", grouped.Repositories[1].Workspaces[0].Pulls[0].URL)
	Equals(t, githubURL, grouped.Repositories[2].Workspaces[0].Pulls[0].URL)
	Equals(t, 1, len(grouped.Repositories[2].Workspaces)) // The staging lock doesn't create a Jobs group.
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
					{Pull: jobs.PullInfo{RepoFullName: "org/no-lock", PullNum: 2}, JobIDInfos: []jobs.JobIDInfo{{JobID: "no-lock", JobIDUrl: "/jobs/no-lock"}}},
				},
			}
			var output bytes.Buffer
			Ok(t, web_templates.BetaDashboardTemplate.Execute(&output, data))
			html := output.String()
			Assert(t, strings.Contains(html, `<h1 class="visually-hidden">Atlantis dashboard</h1>`), "must retain an accessible page heading without using results space")
			Assert(t, !strings.Contains(html, "Dashboard · Beta"), "must not render the removed page header")
			Assert(t, !strings.Contains(html, `<h1>Jobs</h1>`), "must not repeat the jobs section heading in the page header")
			Assert(t, strings.Contains(html, `<title>Jobs · Atlantis</title>`), "must use the Jobs page title")
			Assert(t, strings.Contains(html, `<h2 id="pull-heading">Jobs</h2>`), "must use the Jobs section label")
			Assert(t, strings.Contains(html, `Jobs <span id="pull-count">2</span>`), "navigation must count tracked jobs")
			for _, expected := range []string{"/atlantis/static/css/beta-dashboard.css?v=3", "/atlantis/static/js/beta-dashboard-filters.js?v=2", `/atlantis/?view=classic`, `/atlantis/jobs/job-id`, `/atlantis/lock?id=encoded%252Fid`, "Workspace unavailable", "User unavailable", "test-version", "&lt;script&gt;", `class="dashboard-content" role="region" aria-label="Dashboard results" tabindex="0"`} {
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
	Assert(t, strings.Contains(empty.String(), "No jobs"), "must render jobs empty state")
	Assert(t, strings.Contains(empty.String(), "No locks found."), "must render lock empty state")
	Assert(t, web_templates.BetaDashboardTemplate.Execute(&empty, nil) != nil, "invalid data must return an error")
}

func TestBetaDashboardLockOnlyEntries(t *testing.T) {
	data := web_templates.IndexData{
		Locks:            []web_templates.LockIndexData{{RepoFullName: "lock-only/repo", PullNum: 42, Workspace: "staging", Path: "infra", LockedBy: "owner", LockPath: "/lock?id=lock-only", PullRequestURL: "https://github.com/lock-only/repo/pull/42"}},
		PullToJobMapping: []jobs.PullInfoWithJobIDs{{Pull: jobs.PullInfo{RepoFullName: "lock-only/repo", PullNum: 42, Workspace: "staging", Path: "infra"}}},
	}
	grouped := web_templates.GroupBetaDashboard(data)
	Equals(t, 0, grouped.Total)
	Equals(t, 0, len(grouped.Repositories))
	Equals(t, data.Locks, grouped.Locks)
	var output bytes.Buffer
	Ok(t, web_templates.BetaDashboardTemplate.Execute(&output, data))
	html := output.String()
	sections := strings.SplitN(html, `<section id="locks"`, 2)
	Assert(t, strings.Contains(sections[0], "No jobs"), "lock-only data must leave Jobs empty")
	Assert(t, !strings.Contains(sections[0], "lock-only/repo"), "locks must not synthesize job entries")
	for _, expected := range []string{`class="lock-detail-link" href="/lock?id=lock-only"`, `class="lock-request-link" href="https://github.com/lock-only/repo/pull/42"`, `<code>infra</code>`, `data-repo="lock-only/repo" data-workspace="staging" data-user="owner"`, `class="lock-label">Locked</span>`} {
		Assert(t, strings.Contains(html, expected), "Locks must retain %q", expected)
	}
}

func TestBetaDashboardVersionInSidebar(t *testing.T) {
	version := "dev (commit: none) (build date: unknown)"
	var output bytes.Buffer
	Ok(t, web_templates.BetaDashboardTemplate.Execute(&output, web_templates.IndexData{AtlantisVersion: version}))
	sidebar := strings.SplitN(output.String(), "</aside>", 2)[0]
	Assert(t, strings.Contains(sidebar, `<span>Atlantis</span>`), "sidebar brand must use the capitalized Atlantis name")
	Assert(t, strings.Contains(sidebar, `class="sidebar-footer"`), "version must be in the sidebar footer")
	Assert(t, strings.Contains(sidebar, `<span class="version">`+version+`</span>`), "sidebar must retain the complete version and build information")
	Assert(t, strings.Contains(sidebar, `href="/?view=classic">Classic dashboard</a>`), "classic dashboard must remain available in the sidebar")
	Assert(t, strings.Index(sidebar, "Classic dashboard") < strings.Index(sidebar, `class="sidebar-build"`), "classic dashboard link must precede build information")
	Equals(t, 1, strings.Count(output.String(), "Classic dashboard"))
	Equals(t, 1, strings.Count(output.String(), version))
}

func TestBetaDashboardCompactJobDetails(t *testing.T) {
	var output bytes.Buffer
	data := web_templates.IndexData{
		CleanedBasePath: "/atlantis",
		PullToJobMapping: []jobs.PullInfoWithJobIDs{
			{Pull: jobs.PullInfo{RepoFullName: "org/repo", PullNum: 1, Workspace: "production", Path: "terraform/networking", ProjectName: "networking-production"}, JobIDInfos: []jobs.JobIDInfo{
				{JobIDUrl: "/jobs/first-plan", JobDescription: "plan", JobStep: "plan", TimeFormatted: "2026-10-04 14:00:47"},
				{JobIDUrl: "/jobs/second-plan", JobStep: "plan", TimeFormatted: "2026-10-04 13:59:29"},
			}},
			{Pull: jobs.PullInfo{RepoFullName: "org/repo", PullNum: 1, Workspace: "production", Path: "terraform/storage", ProjectName: "storage"}, JobIDInfos: []jobs.JobIDInfo{
				{JobIDUrl: "/jobs/storage", JobDescription: "Planning storage", JobStep: "plan"},
			}},
		},
	}
	Ok(t, web_templates.BetaDashboardTemplate.Execute(&output, data))
	html := output.String()
	for _, expected := range []string{`class="request-info"`, `class="project-list"`, "networking-production", "terraform/networking", "terraform/storage", `href="/atlantis/jobs/first-plan">plan</a>`, `href="/atlantis/jobs/second-plan">plan</a>`, `href="/atlantis/jobs/storage">Planning storage</a>`, "2026-10-04 14:00:47", "2026-10-04 13:59:29"} {
		Assert(t, strings.Contains(html, expected), "compact job rows must retain %q", expected)
	}
	Equals(t, 3, strings.Count(html, `class="job-link"`))
	Equals(t, 1, strings.Count(html, `<span class="step">plan</span>`)) // Only the distinct description needs a separate step label.
}

func TestClassicDashboardBetaLink(t *testing.T) {
	for _, basePath := range []string{"", "/atlantis"} {
		var output bytes.Buffer
		Ok(t, web_templates.IndexTemplate.Execute(&output, web_templates.IndexData{CleanedBasePath: basePath}))
		Assert(t, strings.Contains(output.String(), `href="`+basePath+`/beta">Try the beta dashboard</a>`), "classic beta link must honor the base path")
	}
}
