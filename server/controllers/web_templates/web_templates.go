// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package web_templates

import (
	"embed"
	"html/template"
	"io"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/runatlantis/atlantis/server/jobs"
)

//go:generate go tool pegomock generate --package mocks -o mocks/mock_template_writer.go TemplateWriter

//go:embed templates/*
var templatesFS embed.FS

// Read all the templates from the embedded filesystem
var templates, _ = template.New("").Funcs(sprig.TxtFuncMap()).ParseFS(templatesFS, "templates/*.tmpl")

var templateFileNames = map[string]string{
	"index":              "index.html.tmpl",
	"lock":               "lock.html.tmpl",
	"project-jobs":       "project-jobs.html.tmpl",
	"project-jobs-error": "project-jobs-error.html.tmpl",
	"github-app":         "github-app.html.tmpl",
}

// TemplateWriter is an interface over html/template that's used to enable
// mocking.
type TemplateWriter interface {
	// Execute applies a parsed template to the specified data object,
	// writing the output to wr.
	Execute(wr io.Writer, data any) error
}

// LockIndexData holds the fields needed to display the index view for locks.
type LockIndexData struct {
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

// IndexData holds the data for rendering the index page
type IndexData struct {
	Locks            []LockIndexData
	PullToJobMapping []jobs.PullInfoWithJobIDs

	ApplyLock       ApplyLockData
	AtlantisVersion string
	// CleanedBasePath is the path Atlantis is accessible at externally. If
	// not using a path-based proxy, this will be an empty string. Never ends
	// in a '/' (hence "cleaned").
	CleanedBasePath string
}

var IndexTemplate = templates.Lookup(templateFileNames["index"])

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

var LockTemplate = templates.Lookup(templateFileNames["lock"])

// ProjectJobData holds the data needed to stream the current PR information
type ProjectJobData struct {
	AtlantisVersion string
	ProjectPath     string
	CleanedBasePath string
}

var ProjectJobsTemplate = templates.Lookup(templateFileNames["project-jobs"])

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
