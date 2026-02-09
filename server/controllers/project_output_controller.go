// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"fmt"
	"html"
	"html/template"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/runatlantis/atlantis/server/controllers/web_templates"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/jobs"
)

// Terraform output syntax highlighting patterns
var (
	tfAddPattern      = regexp.MustCompile(`^(\s*)\+`)
	tfDestroyPattern  = regexp.MustCompile(`^(\s*)-`)
	tfChangePattern   = regexp.MustCompile(`^(\s*)~`)
	tfResourcePattern = regexp.MustCompile(`# ([\w_\.]+)`)
)

// ProjectOutputController handles web page requests for project output views
type ProjectOutputController struct {
	db                           db.Database
	projectOutputTemplate        web_templates.TemplateWriter
	projectOutputPartialTemplate web_templates.TemplateWriter
	atlantisVersion              string
	cleanedBasePath              string
	applyLockChecker             func() bool
	outputHandler                jobs.ProjectCommandOutputHandler
}

// NewProjectOutputController creates a new ProjectOutputController
func NewProjectOutputController(
	database db.Database,
	projectOutputTemplate web_templates.TemplateWriter,
	projectOutputPartialTemplate web_templates.TemplateWriter,
	atlantisVersion string,
	cleanedBasePath string,
	applyLockChecker func() bool,
	outputHandler jobs.ProjectCommandOutputHandler,
) *ProjectOutputController {
	return &ProjectOutputController{
		db:                           database,
		projectOutputTemplate:        projectOutputTemplate,
		projectOutputPartialTemplate: projectOutputPartialTemplate,
		atlantisVersion:              atlantisVersion,
		cleanedBasePath:              cleanedBasePath,
		applyLockChecker:             applyLockChecker,
		outputHandler:                outputHandler,
	}
}

// findActiveJob checks if there's a running job for this project
func (c *ProjectOutputController) findActiveJob(repoFullName string, pullNum int, path, workspace string) *jobs.JobIDInfo {
	if c.outputHandler == nil {
		return nil
	}
	pullMappings := c.outputHandler.GetPullToJobMapping()
	for _, pm := range pullMappings {
		if pm.Pull.RepoFullName == repoFullName &&
			pm.Pull.PullNum == pullNum &&
			pm.Pull.Path == path &&
			pm.Pull.Workspace == workspace {
			// Return first (most recent) job
			if len(pm.JobIDInfos) > 0 {
				job := pm.JobIDInfos[0]
				return &job
			}
		}
	}
	return nil
}

// ProjectOutput renders the project output page
func (c *ProjectOutputController) ProjectOutput(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	owner := vars["owner"]
	repo := vars["repo"]
	repoFullName := owner + "/" + repo
	pullNumStr := vars["pull_num"]
	path := vars["path"]
	workspace := r.URL.Query().Get("workspace")
	projectName := r.URL.Query().Get("project")
	runParam := r.URL.Query().Get("run")

	if workspace == "" {
		workspace = "default"
	}

	pullNum, err := strconv.Atoi(pullNumStr)
	if err != nil {
		http.Error(w, "invalid pull number", http.StatusBadRequest)
		return
	}

	// Fetch history for this project
	history, err := c.db.GetProjectOutputHistory(repoFullName, pullNum, path, workspace, projectName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error fetching project history: %s", err)
		return
	}

	if len(history) == 0 {
		http.Error(w, "project output not found", http.StatusNotFound)
		return
	}

	// Determine which run to display
	var output *models.ProjectOutput
	var isHistorical bool

	if runParam != "" {
		// Parse run timestamp and find specific run
		runTimestamp, err := strconv.ParseInt(runParam, 10, 64)
		if err != nil {
			http.Error(w, "invalid run parameter", http.StatusBadRequest)
			return
		}

		// Find the run in history
		for i := range history {
			if history[i].RunTimestamp == runTimestamp {
				output = &history[i]
				isHistorical = (i != 0) // Not the latest
				break
			}
		}

		if output == nil {
			http.Error(w, "run not found", http.StatusNotFound)
			return
		}
	} else {
		// Show latest (first in sorted history)
		output = &history[0]
		isHistorical = false
	}

	data := c.buildProjectOutputData(output, owner, repo, history, isHistorical)

	if err := c.projectOutputTemplate.Execute(w, data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error rendering template: %s", err)
	}
}

// ProjectOutputPartial returns just the output content for HTMX swaps
func (c *ProjectOutputController) ProjectOutputPartial(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	owner := vars["owner"]
	repo := vars["repo"]
	repoFullName := owner + "/" + repo
	pullNumStr := vars["pull_num"]
	path := vars["path"]
	workspace := r.URL.Query().Get("workspace")
	projectName := r.URL.Query().Get("project")
	runParam := r.URL.Query().Get("run")

	if workspace == "" {
		workspace = "default"
	}

	pullNum, err := strconv.Atoi(pullNumStr)
	if err != nil {
		http.Error(w, "invalid pull number", http.StatusBadRequest)
		return
	}

	if runParam == "" {
		http.Error(w, "run parameter required", http.StatusBadRequest)
		return
	}

	runTimestamp, err := strconv.ParseInt(runParam, 10, 64)
	if err != nil {
		http.Error(w, "invalid run parameter", http.StatusBadRequest)
		return
	}

	// Find the run - we need to search history since we don't know the command
	history, err := c.db.GetProjectOutputHistory(repoFullName, pullNum, path, workspace, projectName)
	if err != nil {
		http.Error(w, "error fetching history", http.StatusInternalServerError)
		return
	}

	var output *models.ProjectOutput
	for i := range history {
		if history[i].RunTimestamp == runTimestamp {
			output = &history[i]
			break
		}
	}

	if output == nil {
		http.Error(w, "run not found", http.StatusNotFound)
		return
	}

	// Build minimal data for partial
	status, _, statusLabel := c.mapStatus(output.Status, output.CommandName)
	var duration string
	if !output.StartedAt.IsZero() && !output.CompletedAt.IsZero() {
		duration = FormatDuration(output.CompletedAt.Sub(output.StartedAt))
	}

	data := web_templates.ProjectOutputData{
		Status:       status,
		StatusLabel:  statusLabel,
		CommandName:  output.CommandName,
		TriggeredBy:  output.TriggeredBy,
		Duration:     duration,
		Workspace:    output.Workspace,
		AddCount:     output.ResourceStats.Add,
		ChangeCount:  output.ResourceStats.Change,
		DestroyCount: output.ResourceStats.Destroy,
		ImportCount:  output.ResourceStats.Import,
		Output:       output.Output,
		OutputHTML:   HighlightTerraformOutput(output.Output),
		Error:        output.Error,
		RunTimestamp: output.RunTimestamp,
		PolicyPassed: output.PolicyPassed,
	}

	if err := c.projectOutputPartialTemplate.Execute(w, data); err != nil {
		http.Error(w, "error rendering partial", http.StatusInternalServerError)
	}
}

func (c *ProjectOutputController) buildProjectOutputData(output *models.ProjectOutput, owner, repo string, history []models.ProjectOutput, isHistorical bool) web_templates.ProjectOutputData {
	status, statusIcon, statusLabel := c.mapStatus(output.Status, output.CommandName)

	// Get latest run status (first in history, which is sorted by timestamp desc)
	var latestStatus, latestStatusLabel string
	if len(history) > 0 {
		latestStatus, _, latestStatusLabel = c.mapStatus(history[0].Status, history[0].CommandName)
	} else {
		latestStatus, latestStatusLabel = status, statusLabel
	}

	var duration string
	if !output.StartedAt.IsZero() && !output.CompletedAt.IsZero() {
		duration = FormatDuration(output.CompletedAt.Sub(output.StartedAt))
	}

	// Build PR URL from stored data, with GitHub URL as fallback
	pullURL := output.PullURL
	if pullURL == "" {
		pullURL = fmt.Sprintf("https://github.com/%s/%s/pull/%d", owner, repo, output.PullNum)
	}

	// Build history items
	historyItems := make([]web_templates.ProjectOutputHistoryItem, len(history))
	for i, h := range history {
		_, _, hStatusLabel := c.mapStatus(h.Status, h.CommandName)
		var hDuration string
		if !h.StartedAt.IsZero() && !h.CompletedAt.IsZero() {
			hDuration = FormatDuration(h.CompletedAt.Sub(h.StartedAt))
		}

		historyItems[i] = web_templates.ProjectOutputHistoryItem{
			RunTimestamp:    h.RunTimestamp,
			RunTimestampFmt: formatTime(h.CompletedAt),
			CommandName:     h.CommandName,
			Status:          c.mapStatusClass(h.Status, h.CommandName),
			StatusLabel:     hStatusLabel,
			TriggeredBy:     h.TriggeredBy,
			Duration:        hDuration,
			IsCurrent:       h.RunTimestamp == output.RunTimestamp,
		}
	}

	// Check for active job
	var activeJob *web_templates.ProjectOutputActiveJob
	if job := c.findActiveJob(output.RepoFullName, output.PullNum, output.Path, output.Workspace); job != nil {
		activeJob = &web_templates.ProjectOutputActiveJob{
			JobID:     job.JobID,
			JobStep:   job.JobStep,
			StartedAt: job.TimeFormatted,
			StreamURL: c.cleanedBasePath + "/jobs/" + job.JobID + "/stream",
		}
	}

	return web_templates.ProjectOutputData{
		LayoutData: web_templates.LayoutData{
			AtlantisVersion: c.atlantisVersion,
			CleanedBasePath: c.cleanedBasePath,
			ActiveNav:       "prs",
			ApplyLockActive: c.applyLockChecker(),
		},

		// Navigation
		RepoFullName: output.RepoFullName,
		RepoOwner:    owner,
		RepoName:     repo,
		PullNum:      output.PullNum,
		PullURL:      pullURL,

		// Project identification
		ProjectName: output.ProjectName,
		Path:        output.Path,
		Workspace:   output.Workspace,

		// Status
		Status:            status,
		StatusIcon:        statusIcon,
		StatusLabel:       statusLabel,
		LatestStatus:      latestStatus,
		LatestStatusLabel: latestStatusLabel,

		// Resource changes
		AddCount:     output.ResourceStats.Add,
		ChangeCount:  output.ResourceStats.Change,
		DestroyCount: output.ResourceStats.Destroy,
		ImportCount:  output.ResourceStats.Import,

		// Metadata
		CommandName: output.CommandName,
		TriggeredBy: output.TriggeredBy,
		StartedAt:   formatTime(output.StartedAt),
		CompletedAt: formatTime(output.CompletedAt),
		Duration:    duration,

		// Output
		Output:     output.Output,
		OutputHTML: HighlightTerraformOutput(output.Output),

		// Policy
		PolicyPassed:     output.PolicyPassed,
		PolicyOutput:     output.PolicyOutput,
		PolicyOutputHTML: template.HTML(html.EscapeString(output.PolicyOutput)), //nolint:gosec // G203: input is escaped

		// Error
		Error: output.Error,

		// History
		RunTimestamp:    output.RunTimestamp,
		RunTimestampFmt: formatTime(output.CompletedAt),
		History:         historyItems,
		IsHistorical:    isHistorical,

		// Live job
		ActiveJob: activeJob,
	}
}

func (c *ProjectOutputController) mapStatusClass(status models.ProjectOutputStatus, commandName string) string {
	statusClass, _ := DetermineProjectStatus(commandName, status)
	return statusClass
}

func (c *ProjectOutputController) mapStatus(status models.ProjectOutputStatus, commandName string) (statusStr, icon, label string) {
	statusClass, statusLabel := DetermineProjectStatus(commandName, status)

	// Map status class to icon
	switch statusClass {
	case "success":
		icon = "✓"
	case "applied":
		icon = "✓✓"
	case "failed":
		icon = "✗"
	case "pending":
		icon = "⏳"
	default:
		icon = "?"
	}

	return statusClass, icon, statusLabel
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("Jan 2, 2006 3:04 PM")
}

// FormatDuration formats a duration as a human-readable string
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}

	if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
}

// HighlightTerraformOutput applies syntax highlighting to terraform output
func HighlightTerraformOutput(output string) template.HTML {
	if output == "" {
		return ""
	}

	// Escape HTML first
	escaped := html.EscapeString(output)

	// Split into lines for processing
	lines := strings.Split(escaped, "\n")
	var highlighted []string

	for _, line := range lines {
		switch {
		case tfAddPattern.MatchString(line):
			highlighted = append(highlighted, `<span class="tf-add">`+line+`</span>`)
		case tfDestroyPattern.MatchString(line):
			highlighted = append(highlighted, `<span class="tf-destroy">`+line+`</span>`)
		case tfChangePattern.MatchString(line):
			highlighted = append(highlighted, `<span class="tf-change">`+line+`</span>`)
		case tfResourcePattern.MatchString(line):
			// Highlight resource names in comments
			highlighted = append(highlighted, tfResourcePattern.ReplaceAllString(line, `# <span class="tf-resource">$1</span>`))
		default:
			highlighted = append(highlighted, line)
		}
	}

	return template.HTML(strings.Join(highlighted, "\n")) //nolint:gosec // G203: content is escaped via html.EscapeString above
}
