// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/runatlantis/atlantis/server/controllers/web_templates"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

// PRController handles web page requests for PR views
type PRController struct {
	db                 db.Database
	prListTemplate     web_templates.TemplateWriter
	prListRowsTemplate web_templates.TemplateWriter
	atlantisVersion    string
	cleanedBasePath    string
	applyLockChecker   func() bool
	getJobsForPull     func() map[string]int
	logger             logging.SimpleLogging
}

// NewPRController creates a new PRController
func NewPRController(
	database db.Database,
	prListTemplate web_templates.TemplateWriter,
	prListRowsTemplate web_templates.TemplateWriter,
	atlantisVersion string,
	cleanedBasePath string,
	applyLockChecker func() bool,
	getJobsForPull func() map[string]int,
	logger logging.SimpleLogging,
) *PRController {
	return &PRController{
		db:                 database,
		prListTemplate:     prListTemplate,
		prListRowsTemplate: prListRowsTemplate,
		atlantisVersion:    atlantisVersion,
		cleanedBasePath:    cleanedBasePath,
		applyLockChecker:   applyLockChecker,
		getJobsForPull:     getJobsForPull,
		logger:             logger,
	}
}

// PRList renders the full PR list page
func (c *PRController) PRList(w http.ResponseWriter, r *http.Request) {
	data, err := c.buildPRListData()
	if err != nil {
		if c.logger != nil {
			c.logger.Err("error loading PR data: %s", err)
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	renderTemplate(w, c.prListTemplate, data, c.logger)
}

// PRListPartial renders just the table rows for htmx refresh
func (c *PRController) PRListPartial(w http.ResponseWriter, r *http.Request) {
	data, err := c.buildPRListData()
	if err != nil {
		if c.logger != nil {
			c.logger.Err("error loading PR data: %s", err)
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	renderTemplate(w, c.prListRowsTemplate, data, c.logger)
}

// buildPRListData loads all PR data - status filtering is done client-side
func (c *PRController) buildPRListData() (web_templates.PRListData, error) {
	pulls, err := c.db.GetActivePullRequests()
	if err != nil {
		return web_templates.PRListData{}, err
	}

	var jobCounts map[string]int
	if c.getJobsForPull != nil {
		jobCounts = c.getJobsForPull()
	}

	items := make([]web_templates.PRListItem, 0, len(pulls))
	repoSet := make(map[string]bool)

	for _, pull := range pulls {
		outputs, err := c.db.GetProjectOutputsByPull(pull.BaseRepo.FullName, pull.Num)
		if err != nil {
			// Show PR with error state instead of hiding it
			// Use current time since we don't have timestamp info for the error case
			now := time.Now()
			items = append(items, web_templates.PRListItem{
				RepoFullName:   pull.BaseRepo.FullName,
				PullNum:        pull.Num,
				Status:         "error",
				ErrorMessage:   "Failed to load project data. The database may be temporarily unavailable.",
				LastActivityTS: now,
				LastActivity:   FormatRelativeTime(now),
			})
			repoSet[pull.BaseRepo.FullName] = true
			continue
		}

		item := c.buildPRListItem(pull, outputs)
		if jobCounts != nil {
			key := fmt.Sprintf("%s/%d", pull.BaseRepo.FullName, pull.Num)
			item.ActiveJobCount = jobCounts[key]
		}
		items = append(items, item)
		repoSet[pull.BaseRepo.FullName] = true
	}

	// Sort by last activity (most recent first)
	sort.Slice(items, func(i, j int) bool {
		return items[i].LastActivityTS.After(items[j].LastActivityTS)
	})

	// Extract unique repositories sorted
	repos := make([]string, 0, len(repoSet))
	for repo := range repoSet {
		repos = append(repos, repo)
	}
	sort.Strings(repos)

	// Build JSON data for Alpine.js
	jsonItems := make([]map[string]any, 0, len(items))
	for _, item := range items {
		jsonItems = append(jsonItems, map[string]any{
			"repo":             item.RepoFullName,
			"pullNum":          item.PullNum,
			"groupKey":         item.RepoFullName,
			"groupDisplayName": item.RepoFullName,
			"title":            item.Title,
			"status":           item.Status,
			"projectCount":     item.ProjectCount,
			"successCount":     item.SuccessCount,
			"addCount":         item.AddCount,
			"changeCount":      item.ChangeCount,
			"destroyCount":     item.DestroyCount,
			"lastActivity":     item.LastActivity,
			"url":              fmt.Sprintf("%s/pr/%s/pulls/%d", c.cleanedBasePath, item.RepoFullName, item.PullNum),
			"errorMessage":     item.ErrorMessage,
			"activeJobCount":   item.ActiveJobCount,
			"jobsUrl":          fmt.Sprintf("%s/jobs?pr=%s/%d", c.cleanedBasePath, item.RepoFullName, item.PullNum),
		})
	}

	// Status filtering is now done client-side via Alpine.js
	return web_templates.PRListData{
		LayoutData: web_templates.LayoutData{
			AtlantisVersion: c.atlantisVersion,
			CleanedBasePath: c.cleanedBasePath,
			ActiveNav:       "prs",
			ApplyLockActive: c.applyLockChecker(),
		},
		PullRequests: items, // All PRs - filtering done client-side
		TotalCount:   len(items),
		Repositories: repos,
		ActiveRepo:   "",
		ScriptData:   web_templates.MustEncodeScriptData(jsonItems),
	}, nil
}

func (c *PRController) buildPRListItem(pull models.PullRequest, outputs []models.ProjectOutput) web_templates.PRListItem {
	item := web_templates.PRListItem{
		RepoFullName: pull.BaseRepo.FullName,
		PullNum:      pull.Num,
		Title:        pull.Title, // Now populated from ProjectOutput.PullTitle
		ProjectCount: len(outputs),
	}

	var latestActivity time.Time

	for _, output := range outputs {
		switch output.Status {
		case models.SuccessOutputStatus:
			item.SuccessCount++
		case models.FailedOutputStatus:
			item.FailedCount++
		case models.RunningOutputStatus:
			item.PendingCount++
		case models.InterruptedOutputStatus:
			item.FailedCount++
		}

		item.AddCount += output.ResourceStats.Add
		item.ChangeCount += output.ResourceStats.Change
		item.DestroyCount += output.ResourceStats.Destroy

		if output.CompletedAt.After(latestActivity) {
			latestActivity = output.CompletedAt
		}
	}

	item.LastActivityTS = latestActivity
	item.LastActivity = FormatRelativeTime(latestActivity)
	item.Status = DetermineStatus(item.SuccessCount, item.FailedCount, item.PendingCount)

	return item
}

// DetermineStatus returns the overall status for a PR based on project counts
func DetermineStatus(success, failed, pending int) string {
	total := success + failed + pending

	if total == 0 {
		return "pending"
	}

	if failed > 0 {
		return "failed"
	}

	if pending > 0 {
		return "pending"
	}

	return "passed"
}

// FormatRelativeTime formats a timestamp as a human-readable relative time
func FormatRelativeTime(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}

	duration := time.Since(t)

	if duration < time.Minute {
		return "just now"
	}

	if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	}

	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}

	days := int(duration.Hours() / 24)
	if days == 1 {
		return "1 day ago"
	}
	return fmt.Sprintf("%d days ago", days)
}
