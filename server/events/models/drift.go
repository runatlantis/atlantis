// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// DriftSummary represents detected infrastructure drift from a plan output.
// It captures the changes detected between the Terraform state and actual infrastructure.
type DriftSummary struct {
	// HasDrift indicates whether any drift was detected.
	HasDrift bool `json:"has_drift"`
	// ToAdd is the number of resources to add.
	ToAdd int `json:"to_add"`
	// ToChange is the number of resources to change.
	ToChange int `json:"to_change"`
	// ToDestroy is the number of resources to destroy.
	ToDestroy int `json:"to_destroy"`
	// ToImport is the number of resources to import (Terraform 1.5+).
	ToImport int `json:"to_import"`
	// ToForget is the number of resources to forget.
	ToForget int `json:"to_forget"`
	// Summary is a human-readable summary of the drift.
	Summary string `json:"summary"`
	// ChangesOutside indicates if Terraform detected changes made outside of Terraform.
	ChangesOutside bool `json:"changes_outside"`
}

// TotalChanges returns the total number of resource changes.
func (d DriftSummary) TotalChanges() int {
	return d.ToAdd + d.ToChange + d.ToDestroy + d.ToImport + d.ToForget
}

// RequiresBaseBranchForRef reports whether a drift API ref needs an explicit branch context.
func RequiresBaseBranchForRef(ref string) bool {
	ref = strings.TrimSpace(ref)
	if strings.HasPrefix(ref, "refs/tags/") {
		return true
	}
	if strings.HasPrefix(ref, "refs/heads/") {
		return false
	}
	if isLikelyBareTagRef(ref) || isAmbiguousBareRef(ref) {
		return true
	}
	if len(ref) < 7 || len(ref) > 40 {
		return false
	}
	for _, r := range ref {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') {
			continue
		}
		return false
	}
	return true
}

func requiresBaseBranch(ref string) bool {
	return RequiresBaseBranchForRef(ref)
}

// IsSupportedDriftVCSHostType reports whether the drift API can parse and
// execute requests for a VCS provider today.
func IsSupportedDriftVCSHostType(vcsType string) bool {
	switch vcsType {
	case "Github", "Gitlab", "Gitea":
		return true
	default:
		return false
	}
}

// IsValidAPIRepositoryForType reports whether repository has a shape that can
// be safely handed to the configured VCS client for drift APIs.
func IsValidAPIRepositoryForType(repository, vcsType string) bool {
	repository = strings.TrimSpace(repository)
	if repository == "" || strings.ContainsAny(repository, " \t\r\n\\") {
		return false
	}
	if strings.HasPrefix(repository, "/") || strings.HasSuffix(repository, "/") || strings.Contains(repository, "//") {
		return false
	}
	parts := strings.Split(repository, "/")
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return false
		}
	}
	switch vcsType {
	case "Github", "Gitea":
		return len(parts) == 2
	case "Gitlab":
		return len(parts) >= 1
	default:
		return false
	}
}

func supportedDriftVCSTypeMessage() string {
	return "type must be one of: Github, Gitlab, Gitea"
}

// IsUnsafeAPIRef reports whether a caller-controlled non-PR API ref can target
// provider pull/merge request namespaces or alter git fetch behavior.
func IsUnsafeAPIRef(ref string) bool {
	normalized := strings.ToLower(strings.TrimSpace(ref))
	if strings.HasPrefix(normalized, "-") || strings.HasPrefix(normalized, "+") {
		return true
	}
	if strings.Contains(normalized, ":") {
		return true
	}
	normalized = strings.TrimPrefix(normalized, "refs/")

	for _, namespace := range []string{"pull", "merge-requests", "pull-requests"} {
		if hasUnsafePRNamespace(normalized, namespace) {
			return true
		}
	}
	return false
}

func hasUnsafePRNamespace(ref, namespace string) bool {
	rest, ok := strings.CutPrefix(ref, namespace+"/")
	if !ok {
		return false
	}
	id, rest, ok := strings.Cut(rest, "/")
	if !ok {
		return false
	}
	if _, err := strconv.Atoi(id); err != nil {
		return false
	}
	action, _, _ := strings.Cut(rest, "/")
	switch action {
	case "head", "merge", "from", "to":
		return true
	default:
		return false
	}
}

// NormalizeAPIRef normalizes branch refs used by drift and API command requests.
func NormalizeAPIRef(ref string) string {
	return strings.TrimPrefix(strings.TrimSpace(ref), "refs/heads/")
}

// CheckedBaseBranchRef returns the normalized branch name for a safe API base_branch.
func CheckedBaseBranchRef(baseBranch string) (string, bool) {
	if IsUnsafeAPIRef(baseBranch) {
		return "", false
	}
	baseRef := NormalizeAPIRef(baseBranch)
	if baseRef == "" || strings.HasPrefix(baseRef, "-") || strings.Contains(baseRef, ":") {
		return "", false
	}
	if strings.HasPrefix(baseRef, "refs/") {
		return "", false
	}
	lower := strings.ToLower(strings.TrimPrefix(baseRef, "refs/"))
	for _, namespace := range []string{"pull", "merge-requests", "pull-requests"} {
		if strings.HasPrefix(lower, namespace+"/") {
			return "", false
		}
	}
	if !isSafeBaseBranchRef(baseRef) {
		return "", false
	}
	return baseRef, true
}

// NormalizeAPIPath returns a clean repo-relative selector path for drift APIs.
func NormalizeAPIPath(directory string) (string, bool) {
	trimmed := strings.TrimSpace(directory)
	if trimmed == "" || strings.Contains(trimmed, "\\") || strings.ContainsAny(trimmed, "*?[]{}") {
		return "", false
	}
	cleaned := path.Clean(trimmed)
	if cleaned == "" || path.IsAbs(cleaned) || cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.ContainsAny(cleaned, "*?[]{}") {
		return "", false
	}
	return cleaned, true
}

// IsValidAPIWorkspace reports whether an explicit drift/remediation workspace
// selector is safe to use for plan-file path construction.
func IsValidAPIWorkspace(workspace string) bool {
	trimmed := strings.TrimSpace(workspace)
	if trimmed == "" || trimmed != workspace {
		return false
	}
	if workspace == "." || workspace == ".." {
		return false
	}
	if strings.ContainsAny(workspace, "/\\") || strings.Contains(workspace, "..") {
		return false
	}
	return true
}

func isSafeBaseBranchRef(baseRef string) bool {
	if baseRef == "" || strings.HasPrefix(baseRef, "-") || strings.HasPrefix(baseRef, "/") || strings.HasSuffix(baseRef, "/") {
		return false
	}
	if strings.ContainsAny(baseRef, " \t\r\n\\~^:?*[]") || strings.Contains(baseRef, "..") || strings.Contains(baseRef, "@{") || strings.Contains(baseRef, "//") {
		return false
	}
	if strings.HasSuffix(baseRef, ".") || strings.HasSuffix(baseRef, ".lock") {
		return false
	}
	for component := range strings.SplitSeq(baseRef, "/") {
		if component == "" || strings.HasPrefix(component, ".") || strings.HasSuffix(component, ".lock") {
			return false
		}
	}
	return true
}

func isLikelyBareTagRef(ref string) bool {
	if strings.Contains(ref, "/") || !strings.Contains(ref, ".") {
		return false
	}
	trimmed := strings.TrimPrefix(strings.TrimPrefix(ref, "v"), "V")
	if trimmed == ref && (ref == "" || ref[0] < '0' || ref[0] > '9') {
		return false
	}
	if trimmed == "" || trimmed[0] < '0' || trimmed[0] > '9' {
		return false
	}
	for _, r := range trimmed {
		if (r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_' {
			continue
		}
		return false
	}
	return true
}

func isAmbiguousBareRef(ref string) bool {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return false
	}
	if strings.Contains(ref, "/") {
		return false
	}
	switch strings.ToLower(ref) {
	case "main", "master", "develop", "development", "dev", "trunk":
		return false
	case "prod", "production", "latest", "stable", "release":
		return true
	default:
		return false
	}
}

// NewDriftSummaryFromPlanStats creates a DriftSummary from PlanSuccessStats.
// This leverages the existing plan output parsing infrastructure.
func NewDriftSummaryFromPlanStats(stats PlanSuccessStats, summary string) DriftSummary {
	hasDrift := stats.Add > 0 || stats.Change > 0 || stats.Destroy > 0 || stats.Import > 0 || stats.Forget > 0 || stats.ChangesOutside
	return DriftSummary{
		HasDrift:       hasDrift,
		ToAdd:          stats.Add,
		ToChange:       stats.Change,
		ToDestroy:      stats.Destroy,
		ToImport:       stats.Import,
		ToForget:       stats.Forget,
		Summary:        summary,
		ChangesOutside: stats.ChangesOutside,
	}
}

// NewDriftSummaryFromPlanSuccess creates a DriftSummary from a PlanSuccess result.
func NewDriftSummaryFromPlanSuccess(plan *PlanSuccess) DriftSummary {
	if plan == nil {
		return DriftSummary{}
	}
	stats := plan.Stats()
	summary := plan.DiffSummary()
	if stats.ChangesOutside {
		summary = plan.Summary()
	}
	return NewDriftSummaryFromPlanStats(stats, summary)
}

// ProjectDrift represents the drift status for a specific project.
type ProjectDrift struct {
	// ProjectName is the name of the project (from atlantis.yaml).
	ProjectName string `json:"project_name"`
	// Path is the relative path to the project root within the repository.
	Path string `json:"path"`
	// Workspace is the Terraform workspace.
	Workspace string `json:"workspace"`
	// Ref is the git reference (branch/tag/commit) that was checked.
	Ref string `json:"ref"`
	// BaseBranch is the branch context used for repo config branch filters and
	// requirements when Ref is a tag or commit SHA.
	BaseBranch string `json:"base_branch,omitempty"`
	// ResolvedCommit is the checked-out commit SHA that produced this drift
	// record. Remediation uses it to avoid applying stale drift after a moving ref
	// changes.
	ResolvedCommit string `json:"resolved_commit,omitempty"`
	// DetectionID links this drift result to the detection run that produced it.
	DetectionID string `json:"detection_id,omitempty"`
	// Drift contains the drift summary for this project.
	Drift DriftSummary `json:"drift"`
	// LastChecked is when the drift was last detected.
	LastChecked time.Time `json:"last_checked"`
	// Error contains any error message if drift detection failed for this project.
	Error string `json:"error,omitempty"`
}

// DriftStatusResponse is the API response for GET /api/drift/status.
type DriftStatusResponse struct {
	// Repository is the full repository name (owner/repo).
	Repository string `json:"repository"`
	// Projects contains the drift status for each project.
	Projects []ProjectDrift `json:"projects"`
	// CheckedAt is when the drift check was performed.
	CheckedAt time.Time `json:"checked_at"`
	// TotalProjects is the total number of projects checked.
	TotalProjects int `json:"total_projects"`
	// ProjectsWithDrift is the number of projects that have drift.
	ProjectsWithDrift int `json:"projects_with_drift"`
}

// NewDriftStatusResponse creates a DriftStatusResponse from a list of ProjectDrift.
func NewDriftStatusResponse(repository string, projects []ProjectDrift) DriftStatusResponse {
	driftCount := 0
	var checkedAt time.Time
	for _, p := range projects {
		if p.Drift.HasDrift {
			driftCount++
		}
		if p.LastChecked.After(checkedAt) {
			checkedAt = p.LastChecked
		}
	}
	return DriftStatusResponse{
		Repository:        repository,
		Projects:          projects,
		CheckedAt:         checkedAt,
		TotalProjects:     len(projects),
		ProjectsWithDrift: driftCount,
	}
}

// DriftDetectionPath represents a project path for drift detection.
type DriftDetectionPath struct {
	// Directory is the relative path to the Terraform directory.
	Directory string `json:"directory"`
	// Workspace is the Terraform workspace (optional).
	Workspace string `json:"workspace,omitempty"`
}

// DriftDetectionRequest is the API request for POST /api/drift/detect.
type DriftDetectionRequest struct {
	// Repository is the full repository name (owner/repo). Required.
	Repository string `json:"repository"`
	// Ref is the git reference (branch/tag/commit) to check for drift. Required.
	Ref string `json:"ref"`
	// BaseBranch is required when ref is a raw commit SHA or tag value.
	// It preserves Atlantis repo-config branch filtering and undiverged checks.
	BaseBranch string `json:"base_branch,omitempty"`
	// Type is the VCS provider type (Github/Gitlab). Required.
	Type string `json:"type"`
	// Projects is a list of project names to check. If empty, all projects are checked.
	Projects []string `json:"projects,omitempty"`
	// Paths is a list of paths to check. If empty, project names are used.
	Paths []DriftDetectionPath `json:"paths,omitempty"`
}

// Validate checks the request and returns any validation errors.
func (r *DriftDetectionRequest) Validate() []FieldError {
	var errors []FieldError

	if r.Repository == "" {
		errors = append(errors, FieldError{Field: "repository", Message: "repository is required"})
	}
	if r.Ref == "" {
		errors = append(errors, FieldError{Field: "ref", Message: "ref is required"})
	} else if IsUnsafeAPIRef(r.Ref) {
		errors = append(errors, FieldError{Field: "ref", Message: "ref is invalid"})
	} else if requiresBaseBranch(r.Ref) && strings.TrimSpace(r.BaseBranch) == "" {
		errors = append(errors, FieldError{Field: "base_branch", Message: "base_branch is required when ref is a commit SHA, tag, or ambiguous bare ref"})
	}
	if r.BaseBranch != "" {
		if _, ok := CheckedBaseBranchRef(r.BaseBranch); !ok {
			errors = append(errors, FieldError{Field: "base_branch", Message: "base_branch is invalid"})
		}
	}
	if r.Type == "" {
		errors = append(errors, FieldError{Field: "type", Message: "type is required"})
	} else if !IsSupportedDriftVCSHostType(r.Type) {
		errors = append(errors, FieldError{Field: "type", Message: supportedDriftVCSTypeMessage()})
	} else if r.Repository != "" && !IsValidAPIRepositoryForType(r.Repository, r.Type) {
		errors = append(errors, FieldError{Field: "repository", Message: "repository must be a valid repository path for the VCS type"})
	}
	if len(r.Projects) > 0 && len(r.Paths) > 0 {
		errors = append(errors, FieldError{Field: "paths", Message: "projects and paths cannot both be set"})
	}
	for _, project := range r.Projects {
		if strings.TrimSpace(project) == "" {
			errors = append(errors, FieldError{Field: "projects", Message: "project names cannot be empty"})
			break
		}
	}
	for _, path := range r.Paths {
		if _, ok := NormalizeAPIPath(path.Directory); !ok {
			errors = append(errors, FieldError{Field: "paths", Message: "path directories must be clean repo-relative paths"})
			break
		}
		if path.Workspace != "" && !IsValidAPIWorkspace(path.Workspace) {
			errors = append(errors, FieldError{Field: "paths", Message: "path workspaces are invalid"})
			break
		}
	}
	return errors
}

// DriftDetectionResult is the API response for POST /api/drift/detect.
type DriftDetectionResult struct {
	// ID is a unique identifier for this detection run.
	ID string `json:"id"`
	// Repository is the full repository name.
	Repository string `json:"repository"`
	// Projects contains the drift status for each project checked.
	Projects []ProjectDrift `json:"projects"`
	// DetectedAt is when the drift detection was performed.
	DetectedAt time.Time `json:"detected_at"`
	// TotalProjects is the total number of projects checked.
	TotalProjects int `json:"total_projects"`
	// ProjectsWithDrift is the number of projects that have drift.
	ProjectsWithDrift int `json:"projects_with_drift"`
}

// NewDriftDetectionResult creates a new DriftDetectionResult with a generated ID.
func NewDriftDetectionResult(repository string) *DriftDetectionResult {
	return &DriftDetectionResult{
		ID:         uuid.New().String(),
		Repository: repository,
		Projects:   []ProjectDrift{},
		DetectedAt: time.Now(),
	}
}

// AddProject adds a project drift result and updates counts.
func (r *DriftDetectionResult) AddProject(project ProjectDrift) {
	r.Projects = append(r.Projects, project)
	r.TotalProjects = len(r.Projects)
	driftCount := 0
	for _, p := range r.Projects {
		if p.Drift.HasDrift {
			driftCount++
		}
	}
	r.ProjectsWithDrift = driftCount
}
