// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/runatlantis/atlantis/server/controllers/web_templates"
	"github.com/runatlantis/atlantis/server/controllers/websocket"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/jobs"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
	tally "github.com/uber-go/tally/v4"
)

type JobIDKeyGenerator struct{}

func (g JobIDKeyGenerator) Generate(r *http.Request) (string, error) {
	jobID, ok := mux.Vars(r)["job-id"]
	if !ok {
		return "", fmt.Errorf("internal error: no job-id in route")
	}

	return jobID, nil
}

type JobsController struct {
	AtlantisVersion          string                       `validate:"required"`
	AtlantisURL              *url.URL                     `validate:"required"`
	Logger                   logging.SimpleLogging        `validate:"required"`
	ProjectJobsTemplate      web_templates.TemplateWriter `validate:"required"`
	ProjectJobsErrorTemplate web_templates.TemplateWriter `validate:"required"`
	Database                 db.Database                  `validate:"required"`
	WsMux                    *websocket.Multiplexor       `validate:"required"`
	KeyGenerator             JobIDKeyGenerator
	StatsScope               tally.Scope                      `validate:"required"`
	OutputHandler            jobs.ProjectCommandOutputHandler // For SSE streaming
}

// computeJobBadge returns badge text, style, and icon based on job step and status
func computeJobBadge(jobStep, status string) (text, style, icon string) {
	switch {
	case status == "running" && jobStep == "plan":
		return "Planning", "pending", "loader"
	case status == "running" && jobStep == "apply":
		return "Applying", "pending", "loader"
	case status == "complete" && jobStep == "plan":
		return "Planned", "success", "check-circle"
	case status == "complete" && jobStep == "apply":
		return "Applied", "success", "check-circle"
	case status == "error" && jobStep == "plan":
		return "Plan Failed", "failed", "x-circle"
	case status == "error" && jobStep == "apply":
		return "Apply Failed", "failed", "x-circle"
	default:
		return "Running", "pending", "loader"
	}
}

func (j *JobsController) getProjectJobs(w http.ResponseWriter, r *http.Request) error {
	jobID, err := j.KeyGenerator.Generate(r)
	if err != nil {
		j.respond(w, logging.Error, http.StatusBadRequest, "%s", err.Error())
		return err
	}

	// Check if job exists
	if !j.OutputHandler.IsKeyExists(jobID) {
		return j.ProjectJobsErrorTemplate.Execute(w, web_templates.ProjectJobData{
			AtlantisVersion: j.AtlantisVersion,
			ProjectPath:     jobID,
			CleanedBasePath: j.AtlantisURL.Path,
		})
	}

	// Try to find job info from pull mapping
	var repoFullName, repoOwner, repoName, projectPath, workspace, jobStep, triggeredBy string
	var pullNum int
	var startTimeUnix, endTimeUnix int64

	pullMappings := j.OutputHandler.GetPullToJobMapping()
	for _, pm := range pullMappings {
		for _, jobInfo := range pm.JobIDInfos {
			if jobInfo.JobID == jobID {
				repoFullName = pm.Pull.RepoFullName
				if parts := strings.Split(repoFullName, "/"); len(parts) == 2 {
					repoOwner = parts[0]
					repoName = parts[1]
				}
				pullNum = pm.Pull.PullNum
				projectPath = pm.Pull.Path
				workspace = pm.Pull.Workspace
				jobStep = jobInfo.JobStep
				triggeredBy = jobInfo.TriggeredBy
				startTimeUnix = jobInfo.Time.UnixMilli()
				// If job has completed, set end time
				if !jobInfo.CompletedAt.IsZero() {
					endTimeUnix = jobInfo.CompletedAt.UnixMilli()
				}
				break
			}
		}
	}

	if jobStep == "" {
		jobStep = "plan" // Default
	}

	// Determine status and load output if job is complete
	status := "running"
	var output string
	var addCount, changeCount, destroyCount int
	var policyPassed bool
	if endTimeUnix > 0 {
		status = "complete"
		// First try to get output from in-memory buffer (for recently completed jobs)
		if asyncHandler, ok := j.OutputHandler.(*jobs.AsyncProjectCommandOutputHandler); ok {
			buffer := asyncHandler.GetProjectOutputBuffer(jobID)
			if len(buffer.Buffer) > 0 {
				output = strings.Join(buffer.Buffer, "\n")
			}
		}
		// Fall back to database if buffer is empty
		if output == "" {
			if projectOutput, err := j.Database.GetProjectOutputByJobID(jobID); err == nil && projectOutput != nil {
				output = projectOutput.Output
				// Also load completion stats
				addCount = projectOutput.ResourceStats.Add
				changeCount = projectOutput.ResourceStats.Change
				destroyCount = projectOutput.ResourceStats.Destroy
				policyPassed = projectOutput.PolicyPassed
				// Use TriggeredBy from DB if not available from live job
				if triggeredBy == "" {
					triggeredBy = projectOutput.TriggeredBy
				}
			}
		}
	}

	// Compute badge based on job step and status
	badgeText, badgeStyle, badgeIcon := computeJobBadge(jobStep, status)

	viewData := web_templates.JobDetailData{
		LayoutData: web_templates.LayoutData{
			AtlantisVersion: j.AtlantisVersion,
			CleanedBasePath: j.AtlantisURL.Path,
			ActiveNav:       "jobs",
		},
		JobID:         jobID,
		JobStep:       jobStep,
		RepoFullName:  repoFullName,
		RepoOwner:     repoOwner,
		RepoName:      repoName,
		PullNum:       pullNum,
		ProjectPath:   projectPath,
		Workspace:     workspace,
		Status:        status,
		StartTimeUnix: startTimeUnix,
		EndTimeUnix:   endTimeUnix,
		Output:        output,
		StreamURL:     j.AtlantisURL.Path + "/jobs/" + jobID + "/stream",
		// Status panel fields
		TriggeredBy:  triggeredBy,
		BadgeText:    badgeText,
		BadgeStyle:   badgeStyle,
		BadgeIcon:    badgeIcon,
		AddCount:     addCount,
		ChangeCount:  changeCount,
		DestroyCount: destroyCount,
		PolicyPassed: policyPassed,
	}

	return j.ProjectJobsTemplate.Execute(w, viewData)
}

func (j *JobsController) GetProjectJobs(w http.ResponseWriter, r *http.Request) {
	errorCounter := j.StatsScope.SubScope("getprojectjobs").Counter(metrics.ExecutionErrorMetric)
	err := j.getProjectJobs(w, r)
	if err != nil {
		j.Logger.Err(err.Error())
		errorCounter.Inc(1)
	}
}

func (j *JobsController) getProjectJobsWS(w http.ResponseWriter, r *http.Request) error {
	err := j.WsMux.Handle(w, r)

	if err != nil {
		j.respond(w, logging.Error, http.StatusInternalServerError, "%s", err.Error())
		return err
	}

	return nil
}

func (j *JobsController) GetProjectJobsWS(w http.ResponseWriter, r *http.Request) {
	jobsMetric := j.StatsScope.SubScope("getprojectjobs")
	errorCounter := jobsMetric.Counter(metrics.ExecutionErrorMetric)
	executionTime := jobsMetric.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	err := j.getProjectJobsWS(w, r)

	if err != nil {
		errorCounter.Inc(1)
	}
}

func (j *JobsController) respond(w http.ResponseWriter, lvl logging.LogLevel, responseCode int, format string, args ...any) {
	response := fmt.Sprintf(format, args...)
	j.Logger.Log(lvl, response)
	w.WriteHeader(responseCode)
	fmt.Fprintln(w, response)
}

func (j *JobsController) getProjectJobsSSE(w http.ResponseWriter, r *http.Request) error {
	jobID, err := j.KeyGenerator.Generate(r)
	if err != nil {
		j.respond(w, logging.Error, http.StatusBadRequest, "%s", err.Error())
		return err
	}

	// Check if job exists
	if !j.OutputHandler.IsKeyExists(jobID) {
		j.respond(w, logging.Warn, http.StatusNotFound, "Job not found: %s", jobID)
		return fmt.Errorf("job not found: %s", jobID)
	}

	// Get flusher for SSE
	flusher, ok := w.(http.Flusher)
	if !ok {
		j.respond(w, logging.Error, http.StatusInternalServerError, "Streaming not supported")
		return fmt.Errorf("streaming not supported")
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// Create channel and register with output handler
	receiver := make(chan string, 1000)
	j.OutputHandler.Register(jobID, receiver)
	defer j.OutputHandler.Deregister(jobID, receiver)

	// Stream output as SSE events
streamLoop:
	for {
		select {
		case line, ok := <-receiver:
			if !ok {
				// Channel closed - job complete
				break streamLoop
			}
			fmt.Fprintf(w, "data: %s\n\n", line)
			flusher.Flush()
		case <-r.Context().Done():
			// Client disconnected
			return nil
		}
	}

	// Send completion event
	fmt.Fprintf(w, "event: complete\ndata: done\n\n")
	flusher.Flush()

	return nil
}

// GetProjectJobsSSE handles SSE streaming for job output.
func (j *JobsController) GetProjectJobsSSE(w http.ResponseWriter, r *http.Request) {
	jobsMetric := j.StatsScope.SubScope("getprojectjobssse")
	errorCounter := jobsMetric.Counter(metrics.ExecutionErrorMetric)
	executionTime := jobsMetric.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	err := j.getProjectJobsSSE(w, r)
	if err != nil {
		j.Logger.Err(err.Error())
		errorCounter.Inc(1)
	}
}

// CreateTestJob creates a test job for development/testing purposes.
// Only available in dev mode.
// Query params:
//   - pattern: output pattern (default, slow, burst, colors, error, long)
//   - repo: repository full name (e.g., "acme/infrastructure")
//   - pr: pull request number
//   - project: project path
//   - step: job step (plan, apply)
func (j *JobsController) CreateTestJob(w http.ResponseWriter, r *http.Request) {
	pattern := TestPattern(r.URL.Query().Get("pattern"))
	if pattern == "" {
		pattern = TestPatternDefault
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

	// Register the job with PR association
	j.OutputHandler.(*jobs.AsyncProjectCommandOutputHandler).RegisterTestJob(jobID, pullInfo, jobStep)

	// Create output channel
	outputChan := make(chan string, 1000)

	// Start generating output in background
	go func() {
		// Small delay to allow redirect to complete
		time.Sleep(100 * time.Millisecond)
		GenerateTestOutput(pattern, outputChan)
	}()

	// Feed output to handler
	go func() {
		for line := range outputChan {
			j.OutputHandler.(*jobs.AsyncProjectCommandOutputHandler).SendTestLine(jobID, line)
		}
		j.OutputHandler.(*jobs.AsyncProjectCommandOutputHandler).MarkComplete(jobID)
	}()

	// Redirect to job page
	redirectURL := fmt.Sprintf("%s/jobs/%s", j.AtlantisURL.Path, jobID)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}
