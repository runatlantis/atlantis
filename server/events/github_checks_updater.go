// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v71/github"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// maxCheckRunTextLength is the maximum number of characters allowed in a
// GitHub check run output text field.
const maxCheckRunTextLength = 65535

// checkRunApplyActionIdentifier is the identifier used for the "Apply" button
// action in a plan check run. The GitHub API requires this to be max 20 chars.
const checkRunApplyActionIdentifier = "apply"

// checkRunExternalIDSep is the separator used when encoding workspace and relDir
// into an ExternalID field when there is no project name.
const checkRunExternalIDSep = "::"

// GithubChecksClientInterface is the subset of the GitHub client used by
// GithubChecksUpdater. It is satisfied by *github.Client (via the Atlantis
// wrapper) so that it can be mocked in tests.
//
//go:generate pegomock generate --package mocks -o mocks/mock_github_checks_client.go github.com/runatlantis/atlantis/server/events GithubChecksClientInterface
type GithubChecksClientInterface interface {
	// CreateCheckRun creates a new GitHub check run and returns its ID.
	CreateCheckRun(logger logging.SimpleLogging, repo models.Repo, opts github.CreateCheckRunOptions) (int64, error)
	// UpdateCheckRun updates an existing check run by its ID.
	UpdateCheckRun(logger logging.SimpleLogging, repo models.Repo, checkRunID int64, opts github.UpdateCheckRunOptions) error
	// FindCheckRunID finds an existing check run by commit ref and name.
	// Returns 0 if not found.
	FindCheckRunID(logger logging.SimpleLogging, repo models.Repo, ref, name string) (int64, error)
}

// GithubChecksUpdater implements CommitStatusUpdater using the GitHub Checks
// API (https://docs.github.com/en/rest/checks). It is an experimental
// alternative to the default commit-status-based updater and requires a GitHub
// App (personal-access tokens do not have write access to the Checks API).
//
// For each Terraform project it creates or updates a check run whose output
// contains the full plan/apply output. Successful plan check runs also include
// an "Apply" action button; when the user clicks it, GitHub sends a
// check_run webhook with action="requested_action" which Atlantis handles in
// VCSEventsController to trigger the apply command.
type GithubChecksUpdater struct {
	// Client is the GitHub client subset used to create/update check runs.
	Client GithubChecksClientInterface
	// StatusName is prepended to every check run name, e.g. "atlantis".
	StatusName string

	// checkRunIDs is an in-memory cache mapping
	// "{owner}/{repo}/{sha}/{checkRunName}" → checkRunID.
	// It avoids an extra ListCheckRunsForRef API call on subsequent updates
	// (pending → completed) within the same Atlantis process lifetime.
	// A sync.RWMutex guards concurrent access since plan/apply can run in
	// parallel goroutines.
	mu          sync.RWMutex
	checkRunIDs map[string]int64
}

// NewGithubChecksUpdater returns a ready-to-use GithubChecksUpdater.
func NewGithubChecksUpdater(client GithubChecksClientInterface, statusName string) *GithubChecksUpdater {
	return &GithubChecksUpdater{
		Client:      client,
		StatusName:  statusName,
		checkRunIDs: make(map[string]int64),
	}
}

// checkRunCacheKey returns the map key for the in-memory cache.
func checkRunCacheKey(repo models.Repo, sha, name string) string {
	return fmt.Sprintf("%s/%s/%s/%s", repo.Owner, repo.Name, sha, name)
}

// checkRunExternalID returns the external_id value encoded into the check run.
// This is used when a "requested_action" webhook arrives so we can route the
// apply command to the right project without storing server-side state.
//
// Format (project name takes precedence):
//
//	projectName          – when ctx.ProjectName != ""
//	workspace::relDir    – otherwise
func checkRunExternalID(ctx command.ProjectContext) string {
	if ctx.ProjectName != "" {
		return ctx.ProjectName
	}
	return ctx.Workspace + checkRunExternalIDSep + ctx.RepoRelDir
}

// ParseCheckRunExternalID parses an external_id back into workspace + relDir or
// projectName so that the "Apply" button webhook can reconstruct the command.
// Returns (projectName, workspace, relDir).
func ParseCheckRunExternalID(externalID string) (projectName, workspace, relDir string) {
	if idx := strings.Index(externalID, checkRunExternalIDSep); idx >= 0 {
		return "", externalID[:idx], externalID[idx+len(checkRunExternalIDSep):]
	}
	return externalID, "", ""
}

// getOrCreateCheckRunID returns the cached ID for a check run, or looks it up
// via the API (falling back to create on not-found). When a new check run is
// created its ID is cached.
func (g *GithubChecksUpdater) getOrCreateCheckRunID(
	logger logging.SimpleLogging,
	repo models.Repo,
	pull models.PullRequest,
	name string,
	createOpts github.CreateCheckRunOptions,
) (int64, error) {
	key := checkRunCacheKey(repo, pull.HeadCommit, name)

	// Fast path: check cache under read lock.
	g.mu.RLock()
	id, ok := g.checkRunIDs[key]
	g.mu.RUnlock()
	if ok && id > 0 {
		return id, nil
	}

	// Slow path: look up via the API.
	id, err := g.Client.FindCheckRunID(logger, repo, pull.HeadCommit, name)
	if err != nil {
		return 0, fmt.Errorf("finding check run: %w", err)
	}
	if id > 0 {
		g.mu.Lock()
		g.checkRunIDs[key] = id
		g.mu.Unlock()
		return id, nil
	}

	// Not found – create a new check run.
	id, err = g.Client.CreateCheckRun(logger, repo, createOpts)
	if err != nil {
		return 0, fmt.Errorf("creating check run: %w", err)
	}
	g.mu.Lock()
	g.checkRunIDs[key] = id
	g.mu.Unlock()
	return id, nil
}

// UpdateCombined creates or updates a single "combined" check run that
// represents the overall status for a plan/apply command across all projects.
func (g *GithubChecksUpdater) UpdateCombined(
	logger logging.SimpleLogging,
	repo models.Repo,
	pull models.PullRequest,
	status models.CommitStatus,
	cmdName command.Name,
) error {
	name := fmt.Sprintf("%s/%s", g.StatusName, cmdName.String())
	ghStatus, ghConclusion := commitStatusToCheckRun(status)
	title := cases.Title(language.English).String(cmdName.String())

	var summary string
	switch status {
	case models.PendingCommitStatus:
		summary = fmt.Sprintf("%s in progress…", title)
	case models.SuccessCommitStatus:
		summary = fmt.Sprintf("%s succeeded.", title)
	case models.FailedCommitStatus:
		summary = fmt.Sprintf("%s failed.", title)
	}

	now := github.Timestamp{Time: time.Now()}
	createOpts := github.CreateCheckRunOptions{
		Name:      name,
		HeadSHA:   pull.HeadCommit,
		Status:    github.Ptr(ghStatus),
		StartedAt: &now,
		Output: &github.CheckRunOutput{
			Title:   github.Ptr(title),
			Summary: github.Ptr(summary),
		},
	}
	if ghConclusion != "" {
		createOpts.Conclusion = github.Ptr(ghConclusion)
		createOpts.CompletedAt = &now
	}

	id, err := g.getOrCreateCheckRunID(logger, repo, pull, name, createOpts)
	if err != nil {
		return err
	}

	updateOpts := github.UpdateCheckRunOptions{
		Name:   name,
		Status: github.Ptr(ghStatus),
		Output: &github.CheckRunOutput{
			Title:   github.Ptr(title),
			Summary: github.Ptr(summary),
		},
	}
	if ghConclusion != "" {
		updateOpts.Conclusion = github.Ptr(ghConclusion)
		updateOpts.CompletedAt = &now
	}
	return g.Client.UpdateCheckRun(logger, repo, id, updateOpts)
}

// UpdateCombinedCount creates or updates a combined check run with a count
// summary, e.g. "2/3 projects planned successfully."
func (g *GithubChecksUpdater) UpdateCombinedCount(
	logger logging.SimpleLogging,
	repo models.Repo,
	pull models.PullRequest,
	status models.CommitStatus,
	cmdName command.Name,
	numSuccess int,
	numTotal int,
) error {
	name := fmt.Sprintf("%s/%s", g.StatusName, cmdName.String())
	ghStatus, ghConclusion := commitStatusToCheckRun(status)

	var cmdVerb string
	switch cmdName {
	case command.Plan:
		cmdVerb = "planned"
	case command.PolicyCheck:
		cmdVerb = "policies checked"
	case command.Apply:
		cmdVerb = "applied"
	default:
		cmdVerb = cmdName.String()
	}

	title := cases.Title(language.English).String(cmdName.String())
	summary := fmt.Sprintf("%d/%d projects %s successfully.", numSuccess, numTotal, cmdVerb)
	now := github.Timestamp{Time: time.Now()}

	createOpts := github.CreateCheckRunOptions{
		Name:      name,
		HeadSHA:   pull.HeadCommit,
		Status:    github.Ptr(ghStatus),
		StartedAt: &now,
		Output: &github.CheckRunOutput{
			Title:   github.Ptr(title),
			Summary: github.Ptr(summary),
		},
	}
	if ghConclusion != "" {
		createOpts.Conclusion = github.Ptr(ghConclusion)
		createOpts.CompletedAt = &now
	}

	id, err := g.getOrCreateCheckRunID(logger, repo, pull, name, createOpts)
	if err != nil {
		return err
	}

	updateOpts := github.UpdateCheckRunOptions{
		Name:   name,
		Status: github.Ptr(ghStatus),
		Output: &github.CheckRunOutput{
			Title:   github.Ptr(title),
			Summary: github.Ptr(summary),
		},
	}
	if ghConclusion != "" {
		updateOpts.Conclusion = github.Ptr(ghConclusion)
		updateOpts.CompletedAt = &now
	}
	return g.Client.UpdateCheckRun(logger, repo, id, updateOpts)
}

// UpdateProject creates or updates a per-project check run.
// For successful plan check runs, the full plan output is included in the check
// run text field (truncated to the GitHub limit of 65 535 chars). An "Apply"
// action button is attached only when the plan reports actual resource changes
// (i.e. PlanSuccess.NoChanges() == false), so that plans with no changes do
// not present an unnecessary Apply button.
func (g *GithubChecksUpdater) UpdateProject(
	ctx command.ProjectContext,
	cmdName command.Name,
	status models.CommitStatus,
	url string,
	result *command.ProjectCommandOutput,
) error {
	projectID := ctx.ProjectName
	if projectID == "" {
		projectID = fmt.Sprintf("%s/%s", ctx.RepoRelDir, ctx.Workspace)
	}
	name := fmt.Sprintf("%s/%s: %s", g.StatusName, cmdName.String(), projectID)
	externalID := checkRunExternalID(ctx)
	ghStatus, ghConclusion := commitStatusToCheckRun(status)

	title := fmt.Sprintf("%s: %s", cases.Title(language.English).String(cmdName.String()), projectID)

	var summary, text string
	var actions []*github.CheckRunAction

	switch status {
	case models.PendingCommitStatus:
		summary = fmt.Sprintf("%s in progress…", cases.Title(language.English).String(cmdName.String()))
	case models.SuccessCommitStatus:
		if result != nil && result.PlanSuccess != nil {
			summary = result.PlanSuccess.DiffSummary()
			text = truncateString(result.PlanSuccess.TerraformOutput, maxCheckRunTextLength)
			// Only add the Apply button when there are actual changes.
			if !result.PlanSuccess.NoChanges() {
				actions = []*github.CheckRunAction{
					{
						Label:       "Apply",
						Description: "Apply this Terraform plan",
						Identifier:  checkRunApplyActionIdentifier,
					},
				}
			}
		} else if result != nil && result.ApplySuccess != "" {
			summary = "Apply succeeded."
			text = truncateString(result.ApplySuccess, maxCheckRunTextLength)
		} else {
			summary = fmt.Sprintf("%s succeeded.", cases.Title(language.English).String(cmdName.String()))
		}
	case models.FailedCommitStatus:
		if result != nil && result.Error != nil {
			summary = fmt.Sprintf("%s failed.", cases.Title(language.English).String(cmdName.String()))
			text = truncateString(result.Error.Error(), maxCheckRunTextLength)
		} else if result != nil && result.Failure != "" {
			summary = fmt.Sprintf("%s failed.", cases.Title(language.English).String(cmdName.String()))
			text = truncateString(result.Failure, maxCheckRunTextLength)
		} else {
			summary = fmt.Sprintf("%s failed.", cases.Title(language.English).String(cmdName.String()))
		}
	}

	now := github.Timestamp{Time: time.Now()}
	output := &github.CheckRunOutput{
		Title:   github.Ptr(title),
		Summary: github.Ptr(summary),
	}
	if text != "" {
		output.Text = github.Ptr(text)
	}

	createOpts := github.CreateCheckRunOptions{
		Name:       name,
		HeadSHA:    ctx.Pull.HeadCommit,
		ExternalID: github.Ptr(externalID),
		Status:     github.Ptr(ghStatus),
		StartedAt:  &now,
		Output:     output,
	}
	if url != "" {
		createOpts.DetailsURL = github.Ptr(url)
	}
	if ghConclusion != "" {
		createOpts.Conclusion = github.Ptr(ghConclusion)
		createOpts.CompletedAt = &now
	}
	if len(actions) > 0 {
		createOpts.Actions = actions
	}

	id, err := g.getOrCreateCheckRunID(ctx.Log, ctx.BaseRepo, ctx.Pull, name, createOpts)
	if err != nil {
		return err
	}

	updateOpts := github.UpdateCheckRunOptions{
		Name:       name,
		ExternalID: github.Ptr(externalID),
		Status:     github.Ptr(ghStatus),
		Output:     output,
	}
	if url != "" {
		updateOpts.DetailsURL = github.Ptr(url)
	}
	if ghConclusion != "" {
		updateOpts.Conclusion = github.Ptr(ghConclusion)
		updateOpts.CompletedAt = &now
	}
	if len(actions) > 0 {
		updateOpts.Actions = actions
	}

	return g.Client.UpdateCheckRun(ctx.Log, ctx.BaseRepo, id, updateOpts)
}

// UpdatePreWorkflowHook creates or updates a check run for a pre-workflow hook.
func (g *GithubChecksUpdater) UpdatePreWorkflowHook(
	log logging.SimpleLogging,
	pull models.PullRequest,
	status models.CommitStatus,
	hookDescription string,
	runtimeDescription string,
	url string,
) error {
	return g.updateWorkflowHook(log, pull, status, hookDescription, runtimeDescription, "pre_workflow_hook", url)
}

// UpdatePostWorkflowHook creates or updates a check run for a post-workflow hook.
func (g *GithubChecksUpdater) UpdatePostWorkflowHook(
	log logging.SimpleLogging,
	pull models.PullRequest,
	status models.CommitStatus,
	hookDescription string,
	runtimeDescription string,
	url string,
) error {
	return g.updateWorkflowHook(log, pull, status, hookDescription, runtimeDescription, "post_workflow_hook", url)
}

func (g *GithubChecksUpdater) updateWorkflowHook(
	log logging.SimpleLogging,
	pull models.PullRequest,
	status models.CommitStatus,
	hookDescription string,
	runtimeDescription string,
	workflowType string,
	url string,
) error {
	name := fmt.Sprintf("%s/%s: %s", g.StatusName, workflowType, hookDescription)
	ghStatus, ghConclusion := commitStatusToCheckRun(status)
	title := hookDescription
	var summary string
	if runtimeDescription != "" {
		summary = runtimeDescription
	} else {
		switch status {
		case models.PendingCommitStatus:
			summary = "in progress…"
		case models.SuccessCommitStatus:
			summary = "succeeded."
		case models.FailedCommitStatus:
			summary = "failed."
		}
	}

	now := github.Timestamp{Time: time.Now()}
	createOpts := github.CreateCheckRunOptions{
		Name:      name,
		HeadSHA:   pull.HeadCommit,
		Status:    github.Ptr(ghStatus),
		StartedAt: &now,
		Output: &github.CheckRunOutput{
			Title:   github.Ptr(title),
			Summary: github.Ptr(summary),
		},
	}
	if url != "" {
		createOpts.DetailsURL = github.Ptr(url)
	}
	if ghConclusion != "" {
		createOpts.Conclusion = github.Ptr(ghConclusion)
		createOpts.CompletedAt = &now
	}

	id, err := g.getOrCreateCheckRunID(log, pull.BaseRepo, pull, name, createOpts)
	if err != nil {
		return err
	}

	updateOpts := github.UpdateCheckRunOptions{
		Name:   name,
		Status: github.Ptr(ghStatus),
		Output: &github.CheckRunOutput{
			Title:   github.Ptr(title),
			Summary: github.Ptr(summary),
		},
	}
	if url != "" {
		updateOpts.DetailsURL = github.Ptr(url)
	}
	if ghConclusion != "" {
		updateOpts.Conclusion = github.Ptr(ghConclusion)
		updateOpts.CompletedAt = &now
	}
	return g.Client.UpdateCheckRun(log, pull.BaseRepo, id, updateOpts)
}

// commitStatusToCheckRun converts an Atlantis CommitStatus into the GitHub
// Checks API status and conclusion strings.
//
// GitHub check run status values: "queued", "in_progress", "completed".
// GitHub check run conclusion values (only when status="completed"):
// "success", "failure", "neutral", "cancelled", "skipped", "timed_out", "action_required".
func commitStatusToCheckRun(s models.CommitStatus) (status, conclusion string) {
	switch s {
	case models.PendingCommitStatus:
		return "in_progress", ""
	case models.SuccessCommitStatus:
		return "completed", "success"
	case models.FailedCommitStatus:
		return "completed", "failure"
	default:
		return "in_progress", ""
	}
}

// truncateString truncates s to at most maxLen characters, appending an
// ellipsis when truncation occurs so the reader knows content was cut.
// When maxLen is smaller than the ellipsis itself, only the first maxLen bytes
// of s are returned without an ellipsis.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	const ellipsis = "\n\n[Output truncated due to GitHub check run size limit]"
	cutAt := maxLen - len(ellipsis)
	if cutAt <= 0 {
		// maxLen is too small to include both content and the ellipsis; just
		// return as many bytes of the original string as the limit allows.
		if maxLen <= 0 {
			return ""
		}
		return s[:maxLen]
	}
	return s[:cutAt] + ellipsis
}
