// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

package web_templates

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/runatlantis/atlantis/server/jobs"
)

//go:generate pegomock generate --package mocks -o mocks/mock_template_writer.go TemplateWriter

//go:embed templates/*
var templatesFS embed.FS

// devMode controls whether templates are loaded from disk (true) or embedded (false)
var devMode bool

// devTemplatesDir is the path to the templates directory when in dev mode
var devTemplatesDir string

// SetDevMode enables development mode where templates are loaded from disk on each request.
// templatesDir should be the path to the templates directory (e.g., "server/controllers/web_templates/templates")
func SetDevMode(enabled bool, templatesDir string) {
	devMode = enabled
	devTemplatesDir = templatesDir
	if enabled {
		fmt.Printf("[DEV MODE] Templates will be loaded from disk: %s\n", templatesDir)
	}
}

// IsDevMode returns whether dev mode is enabled
func IsDevMode() bool {
	return devMode
}

// Read all the templates from the embedded filesystem (for non-layout templates)
var templates, _ = template.New("").Funcs(sprig.TxtFuncMap()).ParseFS(templatesFS, "templates/*.tmpl")

// mustParseLayoutTemplate creates a template set with the layout and a specific page template.
// This is needed because Go templates use global definitions for blocks, so each page
// that extends the layout needs its own template set.
func mustParseLayoutTemplate(pageTemplate string) *template.Template {
	t, err := template.New("").Funcs(sprig.TxtFuncMap()).ParseFS(templatesFS,
		"templates/layout.html.tmpl",
		"templates/"+pageTemplate,
	)
	if err != nil {
		panic("failed to parse template " + pageTemplate + ": " + err.Error())
	}
	// Return the page template, which contains {{ template "layout" . }} at the end
	// This ensures Execute() runs the page template which invokes the layout
	return t.Lookup(pageTemplate)
}

// parseLayoutTemplateFromDisk parses a layout template from disk for dev mode
func parseLayoutTemplateFromDisk(pageTemplate string) (*template.Template, error) {
	layoutPath := filepath.Join(devTemplatesDir, "layout.html.tmpl")
	pagePath := filepath.Join(devTemplatesDir, pageTemplate)

	t, err := template.New("").Funcs(sprig.TxtFuncMap()).ParseFiles(layoutPath, pagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s from disk: %w", pageTemplate, err)
	}
	return t.Lookup(pageTemplate), nil
}

// parseTemplateFromDisk parses a single template from disk for dev mode
func parseTemplateFromDisk(templateName string) (*template.Template, error) {
	// Parse all templates to get shared definitions
	pattern := filepath.Join(devTemplatesDir, "*.tmpl")
	t, err := template.New("").Funcs(sprig.TxtFuncMap()).ParseGlob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates from disk: %w", err)
	}
	return t.Lookup(templateName), nil
}

// DevTemplateWriter wraps a template name and reloads from disk on each Execute call
type DevTemplateWriter struct {
	templateName string
	isLayout     bool
}

// Execute reloads the template from disk and executes it
func (d *DevTemplateWriter) Execute(wr io.Writer, data any) error {
	var t *template.Template
	var err error

	if d.isLayout {
		t, err = parseLayoutTemplateFromDisk(d.templateName)
	} else {
		t, err = parseTemplateFromDisk(d.templateName)
	}

	if err != nil {
		return err
	}
	if t == nil {
		return fmt.Errorf("template %s not found", d.templateName)
	}
	return t.Execute(wr, data)
}

// NewDevTemplateWriter creates a new DevTemplateWriter for the given template
func NewDevTemplateWriter(templateName string, isLayout bool) *DevTemplateWriter {
	return &DevTemplateWriter{
		templateName: templateName,
		isLayout:     isLayout,
	}
}

// GetStaticDir returns the path to the static directory for dev mode
func GetStaticDir() string {
	if devTemplatesDir == "" {
		return ""
	}
	// Go up from templates dir to get to server/static
	return filepath.Join(filepath.Dir(filepath.Dir(devTemplatesDir)), "static")
}

// StaticDirExists checks if the static directory exists
func StaticDirExists() bool {
	staticDir := GetStaticDir()
	if staticDir == "" {
		return false
	}
	info, err := os.Stat(staticDir)
	return err == nil && info.IsDir()
}

// Template name constants
const (
	TemplateName_Layout               = "layout"
	TemplateName_ProjectJobsError     = "project-jobs-error"
	TemplateName_GithubApp            = "github-app"
	TemplateName_PRList               = "pr-list"
	TemplateName_PRListRows           = "pr-list-rows"
	TemplateName_PRDetail             = "pr-detail"
	TemplateName_PRDetailProjects     = "pr-detail-projects"
	TemplateName_ProjectOutput        = "project-output"
	TemplateName_ProjectOutputPartial = "project-output-partial"
	TemplateName_Settings             = "settings"
	TemplateName_LocksPage            = "locks-page"
	TemplateName_JobsPage             = "jobs-page"
	TemplateName_JobsPartial          = "jobs-partial"
	TemplateName_JobDetail            = "job-detail"
)

// layoutTemplates tracks which templates use the layout wrapper
var layoutTemplates = map[string]bool{
	TemplateName_PRList:        true,
	TemplateName_PRDetail:      true,
	TemplateName_ProjectOutput: true,
	TemplateName_Settings:      true,
	TemplateName_LocksPage:     true,
	TemplateName_JobsPage:      true,
	TemplateName_JobDetail:     true,
}

// cachedTemplates holds pre-parsed templates for production mode
var cachedTemplates = map[string]TemplateWriter{}

var templateFileNames = map[string]string{
	"layout":                 "layout.html.tmpl",
	"project-jobs-error":     "project-jobs-error.html.tmpl",
	"github-app":             "github-app.html.tmpl",
	"pr-list":                "pr-list.html.tmpl",
	"pr-list-rows":           "pr-list-rows.html.tmpl",
	"pr-detail":              "pr-detail.html.tmpl",
	"pr-detail-projects":     "pr-detail-projects.html.tmpl",
	"project-output":         "project-output.html.tmpl",
	"project-output-partial": "project-output-partial.html.tmpl",
	"settings":               "settings.html.tmpl",
	"locks-page":             "locks.html.tmpl",
	"jobs-page":              "jobs.html.tmpl",
	"jobs-partial":           "jobs-partial.html.tmpl",
	"job-detail":             "job-detail.html.tmpl",
}

// GetTemplate returns a TemplateWriter for the given template name.
// In dev mode, templates are reloaded from disk on each request.
// In production mode, pre-parsed embedded templates are returned.
func GetTemplate(name string) TemplateWriter {
	fileName, ok := templateFileNames[name]
	if !ok {
		panic("unknown template: " + name)
	}

	if devMode {
		isLayout := layoutTemplates[name]
		return NewDevTemplateWriter(fileName, isLayout)
	}

	// Return cached template for production
	if t, ok := cachedTemplates[name]; ok {
		return t
	}

	// Fallback to lookup (shouldn't happen if init ran)
	if layoutTemplates[name] {
		return mustParseLayoutTemplate(fileName)
	}
	return templates.Lookup(fileName)
}

// init caches all templates at startup for production mode
//
//nolint:gochecknoinits // intentional: cache templates at startup for performance
func init() {
	for name, fileName := range templateFileNames {
		if layoutTemplates[name] {
			cachedTemplates[name] = mustParseLayoutTemplate(fileName)
		} else {
			cachedTemplates[name] = templates.Lookup(fileName)
		}
	}
}

// TemplateWriter is an interface over html/template that's used to enable
// mocking.
type TemplateWriter interface {
	// Execute applies a parsed template to the specified data object,
	// writing the output to wr.
	Execute(wr io.Writer, data any) error
}

// LayoutData contains common fields for the sidebar layout
type LayoutData struct {
	AtlantisVersion string
	CleanedBasePath string
	ActiveNav       string // "prs", "locks", "jobs", "settings"
	ApplyLockActive bool
}

// LockIndexData holds the fields needed to display the index view for locks.
type LockIndexData struct {
	LockID        string
	LockPath      string
	RepoFullName  string
	PullNum       int
	Path          string
	Workspace     string
	LockedBy      string
	Time          time.Time
	TimeFormatted string
}

// ApplyLockData holds the fields to display in the index view
type ApplyLockData struct {
	Locked                 bool
	GlobalApplyLockEnabled bool
	Time                   time.Time
	TimeFormatted          string
}

// LockDetailData holds the fields needed to display the lock detail view.
type LockDetailData struct {
	LockKeyEncoded  string
	LockKey         string
	RepoOwner       string
	RepoName        string
	PullRequestLink string
	LockedBy        string
	Workspace       string
	AtlantisVersion string
	// CleanedBasePath is the path Atlantis is accessible at externally. If
	// not using a path-based proxy, this will be an empty string. Never ends
	// in a '/' (hence "cleaned").
	CleanedBasePath string
}

// ProjectJobData holds the data needed to stream the current PR information
type ProjectJobData struct {
	AtlantisVersion string
	ProjectPath     string
	CleanedBasePath string
}

type ProjectJobsError struct {
	AtlantisVersion string
	ProjectPath     string
	CleanedBasePath string
}

var ProjectJobsErrorTemplate = templates.Lookup(templateFileNames["project-jobs-error"])

// GithubSetupData holds the data for rendering the github app setup page
type GithubSetupData struct {
	Target          string
	Manifest        string
	ID              int64
	Key             string
	WebhookSecret   string
	URL             string
	CleanedBasePath string
}

var GithubAppSetupTemplate = templates.Lookup(templateFileNames["github-app"])

// PRListData holds data for the PR list page template
type PRListData struct {
	LayoutData
	PullRequests []PRListItem
	TotalCount   int
	Repositories []string // For repo filter dropdown
	ActiveRepo   string   // Selected repo filter
	// ActiveStatuses removed - filtering is now done client-side
}

// PRListItem represents a single PR in the list
type PRListItem struct {
	RepoFullName   string
	PullNum        int
	Title          string // Future: from VCS
	Status         string // "passed", "failed", "pending", "mixed", "error"
	StatusIcon     string // Emoji/icon for status
	ProjectCount   int
	SuccessCount   int
	FailedCount    int
	PendingCount   int
	AddCount       int
	ChangeCount    int
	DestroyCount   int
	LastActivity   string    // Human-readable relative time
	LastActivityTS time.Time // For sorting
	ErrorMessage   string    // Non-empty if we failed to load PR data
	ActiveJobCount int       // Number of active jobs for this PR
}

var PRListTemplate = mustParseLayoutTemplate(templateFileNames["pr-list"])
var PRListRowsTemplate = templates.Lookup(templateFileNames["pr-list-rows"])

// PRDetailData holds data for the PR detail page template
type PRDetailData struct {
	LayoutData
	RepoFullName    string
	RepoOwner       string
	RepoName        string
	PullNum         int
	Title           string // Future: from VCS
	PullURL         string // Link to GitHub/GitLab PR
	Projects        []PRDetailProject
	FailedProjects  []PRDetailProject
	TotalCount      int
	SuccessCount    int
	FailedCount     int
	PendingCount    int
	PolicyFailCount int
	AddCount        int
	ChangeCount     int
	DestroyCount    int
	LastActivity    string
	ErrorMessage    string // Non-empty if we failed to load PR data
	// ActiveFilters removed - filtering is now done client-side
}

// PRDetailProject represents a single project in the PR detail view
type PRDetailProject struct {
	ProjectName   string
	Path          string
	Workspace     string
	Status        string // "success", "failed", "pending", "applied"
	StatusLabel   string // Human-readable: "Planned", "Applied", "Plan Failed", etc.
	PolicyPassed  bool
	PolicyIcon    string
	AddCount      int
	ChangeCount   int
	DestroyCount  int
	Error         string
	LastUpdated   string
	LastUpdatedTS time.Time
}

var PRDetailTemplate = mustParseLayoutTemplate(templateFileNames["pr-detail"])
var PRDetailProjectsTemplate = templates.Lookup(templateFileNames["pr-detail-projects"])

// ProjectOutputHistoryItem represents a single run in the history list
type ProjectOutputHistoryItem struct {
	RunTimestamp    int64
	RunTimestampFmt string // "Feb 4, 2:30 PM"
	CommandName     string
	Status          string
	StatusLabel     string
	TriggeredBy     string
	Duration        string
	IsCurrent       bool // True if this is the currently displayed run
}

// ProjectOutputActiveJob represents a currently running job for this project
type ProjectOutputActiveJob struct {
	JobID     string
	JobStep   string
	StartedAt string
	StreamURL string
}

// ProjectOutputData holds data for the project output page template
type ProjectOutputData struct {
	LayoutData

	// Navigation
	RepoFullName string
	RepoOwner    string
	RepoName     string
	PullNum      int
	PullURL      string // Link to GitHub/GitLab PR

	// Project identification
	ProjectName string
	Path        string
	Workspace   string

	// Status (of current run being viewed)
	Status      string // "success", "failed", "pending"
	StatusIcon  string
	StatusLabel string // "Plan succeeded", "Plan failed", etc.

	// Latest run status (always shows latest, even when viewing historical)
	LatestStatus      string
	LatestStatusLabel string

	// Resource changes
	AddCount     int
	ChangeCount  int
	DestroyCount int
	ImportCount  int

	// Metadata
	CommandName string // "plan", "apply", "policy_check"
	TriggeredBy string
	StartedAt   string // Formatted time
	CompletedAt string // Formatted time
	Duration    string // e.g., "2m 34s"

	// Output
	Output     string
	OutputHTML template.HTML // Pre-formatted with highlighting

	// Policy
	PolicyPassed     bool
	PolicyOutput     string
	PolicyOutputHTML template.HTML

	// Error
	Error string

	// History
	RunTimestamp    int64  // Current run timestamp
	RunTimestampFmt string // Formatted for display
	History         []ProjectOutputHistoryItem
	IsHistorical    bool // True if viewing a past run (not latest)

	// Live job (if any)
	ActiveJob *ProjectOutputActiveJob
}

var ProjectOutputTemplate = mustParseLayoutTemplate(templateFileNames["project-output"])
var ProjectOutputPartialTemplate = templates.Lookup(templateFileNames["project-output-partial"])

// SettingsData holds data for the settings page template
type SettingsData struct {
	LayoutData
	GlobalApplyLockEnabled bool
	ApplyLockTime          string
	Version                string
}

var SettingsTemplate = mustParseLayoutTemplate(templateFileNames["settings"])

// LocksPageData holds data for the locks page template
type LocksPageData struct {
	LayoutData
	Locks        []LockIndexData
	TotalCount   int
	Repositories []string
}

var LocksPageTemplate = mustParseLayoutTemplate(templateFileNames["locks-page"])

// JobsPageData holds data for the jobs page template
type JobsPageData struct {
	LayoutData
	Jobs         []jobs.PullInfoWithJobIDs
	TotalCount   int
	Repositories []string // Unique repos for filter dropdown
}

var JobsPageTemplate = mustParseLayoutTemplate(templateFileNames["jobs-page"])
var JobsPartialTemplate = templates.Lookup(templateFileNames["jobs-partial"])

// JobDetailData holds data for the job detail page template
type JobDetailData struct {
	LayoutData

	// Job identification
	JobID   string
	JobStep string

	// PR context (for breadcrumbs)
	RepoFullName string
	RepoOwner    string
	RepoName     string
	PullNum      int
	ProjectPath  string
	Workspace    string

	// Status
	Status        string // "running", "complete", "error"
	StartedAt     string
	StartTimeUnix int64 // Unix timestamp in milliseconds for JS
	EndTimeUnix   int64 // Unix timestamp in milliseconds for JS (0 if still running)
	ElapsedTime   string

	// Status panel fields
	TriggeredBy string // Username who triggered the job
	BadgeText   string // "Planning", "Applying", "Planned", "Applied", "Plan Failed", "Apply Failed"
	BadgeStyle  string // "pending", "success", "failed"
	BadgeIcon   string // "loader", "check-circle", "x-circle"

	// Completion stats (progressive disclosure)
	AddCount     int
	ChangeCount  int
	DestroyCount int
	PolicyPassed bool

	// Output (for completed jobs loaded from database)
	Output string

	// Streaming (for live jobs)
	StreamURL string
}

var JobDetailTemplate = mustParseLayoutTemplate(templateFileNames["job-detail"])
