// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/runatlantis/atlantis/server/controllers/web_templates"
	"github.com/runatlantis/atlantis/server/events/models"
)

// PRDetailDatabase defines the database methods needed by PRDetailController
type PRDetailDatabase interface {
	GetProjectOutputsByPull(repoFullName string, pullNum int) ([]models.ProjectOutput, error)
}

// PRDetailController handles the PR detail page
type PRDetailController struct {
	db                       PRDetailDatabase
	prDetailTemplate         web_templates.TemplateWriter
	prDetailProjectsTemplate web_templates.TemplateWriter
	atlantisVersion          string
	cleanedBasePath          string
	applyLockChecker         func() bool
}

// NewPRDetailController creates a new PRDetailController
func NewPRDetailController(
	database PRDetailDatabase,
	prDetailTemplate web_templates.TemplateWriter,
	prDetailProjectsTemplate web_templates.TemplateWriter,
	atlantisVersion string,
	cleanedBasePath string,
	applyLockChecker func() bool,
) *PRDetailController {
	return &PRDetailController{
		db:                       database,
		prDetailTemplate:         prDetailTemplate,
		prDetailProjectsTemplate: prDetailProjectsTemplate,
		atlantisVersion:          atlantisVersion,
		cleanedBasePath:          cleanedBasePath,
		applyLockChecker:         applyLockChecker,
	}
}

// PRDetail renders the full PR detail page
func (c *PRDetailController) PRDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	owner := vars["owner"]
	repo := vars["repo"]
	pullNumStr := vars["pull_num"]

	pullNum, err := strconv.Atoi(pullNumStr)
	if err != nil {
		http.Error(w, "invalid pull number", http.StatusBadRequest)
		return
	}

	// Filtering is done client-side, so always load all projects
	data, err := c.buildPRDetailData(owner, repo, pullNum)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error loading PR data: %s", err)
		return
	}

	if err := c.prDetailTemplate.Execute(w, data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error rendering template: %s", err)
	}
}

// PRDetailProjects renders just the project list for htmx filter updates
func (c *PRDetailController) PRDetailProjects(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	owner := vars["owner"]
	repo := vars["repo"]
	pullNumStr := vars["pull_num"]

	pullNum, err := strconv.Atoi(pullNumStr)
	if err != nil {
		http.Error(w, "invalid pull number", http.StatusBadRequest)
		return
	}

	// Filtering is done client-side, so always load all projects
	data, err := c.buildPRDetailData(owner, repo, pullNum)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error loading PR data: %s", err)
		return
	}

	if err := c.prDetailProjectsTemplate.Execute(w, data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error rendering template: %s", err)
	}
}

func (c *PRDetailController) buildPRDetailData(owner, repo string, pullNum int) (web_templates.PRDetailData, error) {
	repoFullName := owner + "/" + repo

	outputs, err := c.db.GetProjectOutputsByPull(repoFullName, pullNum)
	if err != nil {
		// Return page with error message instead of 500
		return web_templates.PRDetailData{
			LayoutData: web_templates.LayoutData{
				AtlantisVersion: c.atlantisVersion,
				CleanedBasePath: c.cleanedBasePath,
				ActiveNav:       "prs",
				ApplyLockActive: c.applyLockChecker(),
			},
			RepoFullName: repoFullName,
			RepoOwner:    owner,
			RepoName:     repo,
			PullNum:      pullNum,
			PullURL:      fmt.Sprintf("https://github.com/%s/pull/%d", repoFullName, pullNum), // Fallback for error case
			ErrorMessage: "Unable to load project data for this pull request. This may be due to a temporary database issue. Please try refreshing the page.",
		}, nil
	}

	allProjects := make([]web_templates.PRDetailProject, 0, len(outputs))
	failedProjects := make([]web_templates.PRDetailProject, 0)

	var successCount, failedCount, pendingCount, policyFailCount int
	var addCount, changeCount, destroyCount int
	var latestActivity time.Time

	for _, output := range outputs {
		project := BuildDetailProject(output)
		allProjects = append(allProjects, project)

		switch output.Status {
		case models.SuccessOutputStatus:
			successCount++
		case models.FailedOutputStatus:
			failedCount++
			failedProjects = append(failedProjects, project)
		case models.PendingOutputStatus:
			pendingCount++
		}

		if !output.PolicyPassed {
			policyFailCount++
		}

		addCount += output.ResourceStats.Add
		changeCount += output.ResourceStats.Change
		destroyCount += output.ResourceStats.Destroy

		if output.CompletedAt.After(latestActivity) {
			latestActivity = output.CompletedAt
		}
	}

	// Sort by status (failed first), then by path
	sort.Slice(allProjects, func(i, j int) bool {
		if allProjects[i].Status != allProjects[j].Status {
			return statusPriority(allProjects[i].Status) < statusPriority(allProjects[j].Status)
		}
		return allProjects[i].Path < allProjects[j].Path
	})

	// Build PR URL from stored data, with GitHub URL as fallback
	pullURL := ""
	if len(outputs) > 0 && outputs[0].PullURL != "" {
		pullURL = outputs[0].PullURL
	} else {
		pullURL = fmt.Sprintf("https://github.com/%s/pull/%d", repoFullName, pullNum)
	}

	return web_templates.PRDetailData{
		LayoutData: web_templates.LayoutData{
			AtlantisVersion: c.atlantisVersion,
			CleanedBasePath: c.cleanedBasePath,
			ActiveNav:       "prs",
			ApplyLockActive: c.applyLockChecker(),
		},
		RepoFullName:    repoFullName,
		RepoOwner:       owner,
		RepoName:        repo,
		PullNum:         pullNum,
		PullURL:         pullURL,
		Projects:        allProjects, // All projects - filtering done client-side
		FailedProjects:  failedProjects,
		TotalCount:      len(allProjects),
		SuccessCount:    successCount,
		FailedCount:     failedCount,
		PendingCount:    pendingCount,
		PolicyFailCount: policyFailCount,
		AddCount:        addCount,
		ChangeCount:     changeCount,
		DestroyCount:    destroyCount,
		LastActivity:    FormatRelativeTime(latestActivity),
	}, nil
}

// BuildDetailProject converts a ProjectOutput to PRDetailProject
func BuildDetailProject(output models.ProjectOutput) web_templates.PRDetailProject {
	status, statusLabel := DetermineProjectStatus(output.CommandName, output.Status)

	policyIcon := "checkmark"
	if !output.PolicyPassed {
		policyIcon = "x"
	}

	return web_templates.PRDetailProject{
		ProjectName:   output.ProjectName,
		Path:          output.Path,
		Workspace:     output.Workspace,
		Status:        status,
		StatusLabel:   statusLabel,
		PolicyPassed:  output.PolicyPassed,
		PolicyIcon:    policyIcon,
		AddCount:      output.ResourceStats.Add,
		ChangeCount:   output.ResourceStats.Change,
		DestroyCount:  output.ResourceStats.Destroy,
		Error:         output.Error,
		LastUpdated:   FormatRelativeTime(output.CompletedAt),
		LastUpdatedTS: output.CompletedAt,
	}
}

// DetermineProjectStatus returns the CSS status class and human-readable label
// based on the command name and execution status
func DetermineProjectStatus(commandName string, status models.ProjectOutputStatus) (statusClass string, statusLabel string) {
	switch commandName {
	case "plan":
		switch status {
		case models.PendingOutputStatus:
			return "pending", "Planning"
		case models.SuccessOutputStatus:
			return "success", "Planned"
		case models.FailedOutputStatus:
			return "failed", "Plan Failed"
		}
	case "apply":
		switch status {
		case models.PendingOutputStatus:
			return "pending", "Applying"
		case models.SuccessOutputStatus:
			return "applied", "Applied"
		case models.FailedOutputStatus:
			return "failed", "Apply Failed"
		}
	case "policy_check":
		switch status {
		case models.PendingOutputStatus:
			return "pending", "Checking Policy"
		case models.SuccessOutputStatus:
			return "success", "Policy Checked"
		case models.FailedOutputStatus:
			return "failed", "Policy Check Failed"
		}
	default:
		// Handle empty or unknown command name - fall back to simple status
		switch status {
		case models.PendingOutputStatus:
			return "pending", "Pending"
		case models.SuccessOutputStatus:
			return "success", "Success"
		case models.FailedOutputStatus:
			return "failed", "Failed"
		}
	}
	// Default fallback
	return "pending", "Pending"
}

func statusPriority(status string) int {
	switch status {
	case "failed":
		return 0
	case "pending":
		return 1
	case "success":
		return 2
	default:
		return 3
	}
}
