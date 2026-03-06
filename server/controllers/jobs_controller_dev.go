//go:build dev

package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/runatlantis/atlantis/server/jobs"
	"github.com/runatlantis/atlantis/server/logging"
)

// testOutputGenerators holds registered test output generators, keyed by pattern name.
// Populated via init() in jobs_test_patterns.go.
var testOutputGenerators map[string]func(chan<- string)

// registerTestPattern registers a test output generator for dev mode.
func registerTestPattern(name string, gen func(chan<- string)) {
	if testOutputGenerators == nil {
		testOutputGenerators = make(map[string]func(chan<- string))
	}
	testOutputGenerators[name] = gen
}

// CreateTestJob creates a test job for development/testing purposes.
// Query params:
//   - pattern: output pattern (default, slow, burst, colors, error, long)
//   - repo: repository full name (e.g., "acme/infrastructure")
//   - pr: pull request number
//   - project: project path
//   - step: job step (plan, apply)
func (j *JobsController) CreateTestJob(w http.ResponseWriter, r *http.Request) {
	patternName := r.URL.Query().Get("pattern")
	if patternName == "" {
		patternName = "default"
	}

	gen, ok := testOutputGenerators[patternName]
	if !ok {
		gen, ok = testOutputGenerators["default"]
		if !ok {
			http.Error(w, "No test patterns registered", http.StatusNotFound)
			return
		}
	}

	// Parse PR params with defaults
	repoFullName := r.URL.Query().Get("repo")
	if repoFullName == "" {
		repoFullName = "test/repository"
	}

	prNumStr := r.URL.Query().Get("pr")
	prNum := 0
	if prNumStr != "" {
		if n, err := strconv.Atoi(prNumStr); err == nil {
			prNum = n
		}
	}

	projectPath := r.URL.Query().Get("project")
	if projectPath == "" {
		projectPath = "terraform/test"
	}

	jobStep := r.URL.Query().Get("step")
	if jobStep == "" {
		jobStep = "plan"
	}

	// Extract repo name from full name
	repoName := repoFullName
	if parts := strings.Split(repoFullName, "/"); len(parts) == 2 {
		repoName = parts[1]
	}

	// Generate unique job ID
	jobID := fmt.Sprintf("test-%d", time.Now().UnixNano())

	// Build PullInfo for job mapping
	pullInfo := jobs.PullInfo{
		PullNum:      prNum,
		Repo:         repoName,
		RepoFullName: repoFullName,
		ProjectName:  "",
		Path:         projectPath,
		Workspace:    "default",
	}

	handler, ok := j.OutputHandler.(*jobs.AsyncProjectCommandOutputHandler)
	if !ok {
		j.respond(w, logging.Warn, http.StatusInternalServerError, "Dev test jobs not supported with this output handler")
		return
	}

	// Register the job with PR association
	handler.RegisterTestJob(jobID, pullInfo, jobStep)

	// Create output channel
	outputChan := make(chan string, 1000)

	// Start generating output in background
	go func() {
		// Small delay to allow redirect to complete
		time.Sleep(100 * time.Millisecond)
		gen(outputChan)
		close(outputChan)
	}()

	// Feed output to handler
	go func() {
		for line := range outputChan {
			handler.SendTestLine(jobID, line)
		}
		handler.MarkComplete(jobID)
	}()

	// Redirect to job page
	redirectURL := fmt.Sprintf("%s/jobs/%s", j.AtlantisURL.Path, jobID)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}
