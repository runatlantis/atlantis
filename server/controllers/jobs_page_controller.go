// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/runatlantis/atlantis/server/controllers/web_templates"
	"github.com/runatlantis/atlantis/server/jobs"
)

// JobsPageController handles the jobs page
type JobsPageController struct {
	template        web_templates.TemplateWriter
	partialTemplate web_templates.TemplateWriter
	getJobs         func() []jobs.PullInfoWithJobIDs
	isApplyLocked   func() bool
	atlantisVersion string
	cleanedBasePath string
}

// NewJobsPageController creates a new JobsPageController
func NewJobsPageController(
	template web_templates.TemplateWriter,
	partialTemplate web_templates.TemplateWriter,
	getJobs func() []jobs.PullInfoWithJobIDs,
	isApplyLocked func() bool,
	atlantisVersion string,
	cleanedBasePath string,
) *JobsPageController {
	return &JobsPageController{
		template:        template,
		partialTemplate: partialTemplate,
		getJobs:         getJobs,
		isApplyLocked:   isApplyLocked,
		atlantisVersion: atlantisVersion,
		cleanedBasePath: cleanedBasePath,
	}
}

// buildJobsPageData builds the page data from current jobs
func (c *JobsPageController) buildJobsPageData(r *http.Request) web_templates.JobsPageData {
	jobsList := c.getJobs()

	// Apply PR filter if specified
	prFilter := r.URL.Query().Get("pr")
	if prFilter != "" {
		var filteredJobs []jobs.PullInfoWithJobIDs
		for _, pull := range jobsList {
			pullKey := fmt.Sprintf("%s/%d", pull.Pull.RepoFullName, pull.Pull.PullNum)
			if pullKey == prFilter || pull.Pull.RepoFullName == prFilter {
				filteredJobs = append(filteredJobs, pull)
			}
		}
		jobsList = filteredJobs
	}

	// Count total jobs and extract unique repositories
	totalJobs := 0
	repoSet := make(map[string]struct{})
	for _, pull := range jobsList {
		totalJobs += len(pull.JobIDInfos)
		if pull.Pull.RepoFullName != "" {
			repoSet[pull.Pull.RepoFullName] = struct{}{}
		}
	}

	// Sort repositories alphabetically
	repositories := make([]string, 0, len(repoSet))
	for repo := range repoSet {
		repositories = append(repositories, repo)
	}
	sort.Strings(repositories)

	return web_templates.JobsPageData{
		LayoutData: web_templates.LayoutData{
			AtlantisVersion: c.atlantisVersion,
			CleanedBasePath: c.cleanedBasePath,
			ActiveNav:       "jobs",
			ApplyLockActive: c.isApplyLocked(),
		},
		Jobs:         jobsList,
		TotalCount:   totalJobs,
		Repositories: repositories,
	}
}

// Get renders the jobs page
func (c *JobsPageController) Get(w http.ResponseWriter, r *http.Request) {
	data := c.buildJobsPageData(r)
	if err := c.template.Execute(w, data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// GetPartial renders just the jobs list partial for HTMX refresh
func (c *JobsPageController) GetPartial(w http.ResponseWriter, r *http.Request) {
	data := c.buildJobsPageData(r)
	if err := c.partialTemplate.Execute(w, data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
