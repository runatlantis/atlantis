// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
	"github.com/runatlantis/atlantis/server/controllers/web_templates"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/jobs"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
	tally "github.com/uber-go/tally/v4"
)

// sseConnectionCount tracks active SSE connections
var sseConnectionCount atomic.Int32

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
	KeyGenerator             JobIDKeyGenerator
	StatsScope               tally.Scope                      `validate:"required"`
	OutputHandler            jobs.ProjectCommandOutputHandler `validate:"required"`
	ApplyLockChecker         func() bool
	SSEMaxConnections        int32
	SSEIdleTimeout           time.Duration
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
	case status == "interrupted":
		return "Interrupted", "failed", "alert-triangle"
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

	// Check if job exists in memory
	if !j.OutputHandler.IsKeyExists(jobID) {
		// Check if job exists in the database (completed or interrupted jobs)
		if dbOutput, err := j.Database.GetProjectOutputByJobID(jobID); err == nil && dbOutput != nil {
			// Map DB status to UI status
			var status string
			switch dbOutput.Status {
			case models.InterruptedOutputStatus, models.RunningOutputStatus:
				status = "interrupted"
			case models.SuccessOutputStatus:
				status = "complete"
			case models.FailedOutputStatus:
				status = "error"
			default:
				status = "interrupted"
			}

			// Extract repo owner/name
			var repoOwner, repoName string
			if parts := strings.Split(dbOutput.RepoFullName, "/"); len(parts) == 2 {
				repoOwner = parts[0]
				repoName = parts[1]
			}

			badgeText, badgeStyle, _ := computeJobBadge(dbOutput.CommandName, status)

			var applyLockActive bool
			if j.ApplyLockChecker != nil {
				applyLockActive = j.ApplyLockChecker()
			}

			viewData := web_templates.JobDetailData{
				LayoutData: web_templates.LayoutData{
					AtlantisVersion: j.AtlantisVersion,
					CleanedBasePath: j.AtlantisURL.Path,
					ActiveNav:       "jobs",
					ApplyLockActive: applyLockActive,
				},
				JobID:          jobID,
				JobStep:        dbOutput.CommandName,
				RepoFullName:   dbOutput.RepoFullName,
				RepoOwner:      repoOwner,
				RepoName:       repoName,
				PullNum:        dbOutput.PullNum,
				ProjectPath:    dbOutput.Path,
				Workspace:      dbOutput.Workspace,
				Status:         status,
				Output:         dbOutput.Output,
				TriggeredBy:    dbOutput.TriggeredBy,
				BadgeText:      badgeText,
				BadgeStyle:     badgeStyle,
				AddCount:       dbOutput.ResourceStats.Add,
				ChangeCount:    dbOutput.ResourceStats.Change,
				DestroyCount:   dbOutput.ResourceStats.Destroy,
				PolicyPassed:   dbOutput.PolicyPassed,
				HasPolicyCheck: dbOutput.CommandName == "policy_check" || dbOutput.PolicyOutput != "",
			}

			if !dbOutput.StartedAt.IsZero() {
				viewData.StartTimeUnix = dbOutput.StartedAt.UnixMilli()
			}
			if !dbOutput.CompletedAt.IsZero() {
				viewData.EndTimeUnix = dbOutput.CompletedAt.UnixMilli()
			}
			viewData.TerminalScriptData = web_templates.MustEncodeScriptData(map[string]any{
				"output":    dbOutput.Output,
				"badgeText": badgeText,
				"jobStep":   dbOutput.CommandName,
				"status":    status,
				"startTime": viewData.StartTimeUnix,
				"endTime":   viewData.EndTimeUnix,
			})

			return j.ProjectJobsTemplate.Execute(w, viewData)
		}

		// Not found in memory or database
		var applyLockActive bool
		if j.ApplyLockChecker != nil {
			applyLockActive = j.ApplyLockChecker()
		}
		return j.ProjectJobsErrorTemplate.Execute(w, web_templates.ProjectJobsError{
			LayoutData: web_templates.LayoutData{
				AtlantisVersion: j.AtlantisVersion,
				CleanedBasePath: j.AtlantisURL.Path,
				ActiveNav:       "jobs",
				ApplyLockActive: applyLockActive,
			},
		})
	}

	// Try to find job info from pull mapping
	var repoFullName, repoOwner, repoName, projectPath, workspace, jobStep, triggeredBy string
	var pullNum int
	var startTimeUnix, endTimeUnix int64

	pullMappings := j.OutputHandler.GetPullToJobMapping()
outer:
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
				break outer
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
	var policyPassed, hasPolicyCheck bool
	if endTimeUnix > 0 {
		status = "complete"
		// First try to get output from in-memory buffer (for recently completed jobs)
		if asyncHandler, ok := j.OutputHandler.(*jobs.AsyncProjectCommandOutputHandler); ok {
			buffer := asyncHandler.GetProjectOutputBuffer(jobID)
			if len(buffer.Buffer) > 0 {
				output = strings.Join(buffer.Buffer, "\n")
			}
		}
		// Always try database for stats and metadata (buffer only has raw output lines)
		if projectOutput, err := j.Database.GetProjectOutputByJobID(jobID); err == nil && projectOutput != nil {
			if output == "" {
				output = projectOutput.Output
			}
			addCount = projectOutput.ResourceStats.Add
			changeCount = projectOutput.ResourceStats.Change
			destroyCount = projectOutput.ResourceStats.Destroy
			policyPassed = projectOutput.PolicyPassed
			hasPolicyCheck = projectOutput.CommandName == "policy_check" || projectOutput.PolicyOutput != ""
			if triggeredBy == "" {
				triggeredBy = projectOutput.TriggeredBy
			}
			// Detect failed status from DB record
			if projectOutput.Status == models.FailedOutputStatus || projectOutput.Error != "" {
				status = "error"
			}
		}
	}

	// Compute badge based on job step and status
	badgeText, badgeStyle, _ := computeJobBadge(jobStep, status)

	var applyLockActive bool
	if j.ApplyLockChecker != nil {
		applyLockActive = j.ApplyLockChecker()
	}

	viewData := web_templates.JobDetailData{
		LayoutData: web_templates.LayoutData{
			AtlantisVersion: j.AtlantisVersion,
			CleanedBasePath: j.AtlantisURL.Path,
			ActiveNav:       "jobs",
			ApplyLockActive: applyLockActive,
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
		TriggeredBy:    triggeredBy,
		BadgeText:      badgeText,
		BadgeStyle:     badgeStyle,
		AddCount:       addCount,
		ChangeCount:    changeCount,
		DestroyCount:   destroyCount,
		PolicyPassed:   policyPassed,
		HasPolicyCheck: hasPolicyCheck,
	}
	viewData.TerminalScriptData = web_templates.MustEncodeScriptData(map[string]any{
		"output":    output,
		"badgeText": badgeText,
		"jobStep":   jobStep,
		"status":    status,
		"startTime": startTimeUnix,
		"endTime":   endTimeUnix,
	})

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

func (j *JobsController) respond(w http.ResponseWriter, lvl logging.LogLevel, responseCode int, format string, args ...any) {
	response := fmt.Sprintf(format, args...)
	j.Logger.Log(lvl, response)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(responseCode)
	fmt.Fprintln(w, response)
}

func (j *JobsController) getProjectJobsSSE(w http.ResponseWriter, r *http.Request) error {
	// Check connection limit
	if j.SSEMaxConnections > 0 && sseConnectionCount.Load() >= j.SSEMaxConnections {
		j.respond(w, logging.Warn, http.StatusServiceUnavailable, "Too many active connections")
		return fmt.Errorf("SSE connection limit reached: %d", j.SSEMaxConnections)
	}
	sseConnectionCount.Add(1)
	defer sseConnectionCount.Add(-1)

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

	// Parse Last-Event-ID for SSE reconnection support.
	// On reconnect, the browser sends this header with the last received event ID.
	var resumeFrom int
	if lastID := r.Header.Get("Last-Event-ID"); lastID != "" {
		if n, err := strconv.Atoi(lastID); err == nil && n >= 0 {
			resumeFrom = n + 1 // Resume from the line after the last received
		}
	}

	// Register for live output and get buffered lines.
	// Register is synchronous -- it snapshots the buffer under locks and
	// returns without sending anything through the channel.
	receiver := make(chan string, 1000)
	buffered, complete := j.OutputHandler.Register(jobID, receiver)
	defer j.OutputHandler.Deregister(jobID, receiver)

	// Write buffered lines directly to the response (not through the channel).
	// Skip lines already received by the client (SSE reconnection).
	for i, line := range buffered {
		if i < resumeFrom {
			continue
		}
		fmt.Fprintf(w, "id: %d\n", i)
		for part := range strings.SplitSeq(line, "\n") {
			fmt.Fprintf(w, "data: %s\n", part)
		}
		fmt.Fprint(w, "\n")
	}
	if len(buffered) > resumeFrom {
		flusher.Flush()
	}

	// If job was already complete when we registered, send completion and return.
	if complete {
		if info := j.OutputHandler.GetJobInfo(jobID); info != nil && !info.CompletedAt.IsZero() {
			fmt.Fprintf(w, "event: complete\ndata: done\n\n")
			flusher.Flush()
		}
		return nil
	}

	// Stream live output as SSE events
	lineNum := len(buffered)
	idleTimeout := j.SSEIdleTimeout
	if idleTimeout <= 0 {
		idleTimeout = 30 * time.Minute
	}
	idleTimer := time.NewTimer(idleTimeout)
	defer idleTimer.Stop()

streamLoop:
	for {
		select {
		case line, ok := <-receiver:
			if !ok {
				// Channel closed -- could be job completion or buffer overflow
				break streamLoop
			}
			idleTimer.Reset(idleTimeout)
			// SSE spec: multi-line data must use separate "data:" fields
			fmt.Fprintf(w, "id: %d\n", lineNum)
			for part := range strings.SplitSeq(line, "\n") {
				fmt.Fprintf(w, "data: %s\n", part)
			}
			fmt.Fprint(w, "\n")
			flusher.Flush()
			lineNum++
		case <-idleTimer.C:
			j.Logger.Warn("SSE idle timeout for job %s", jobID)
			break streamLoop
		case <-r.Context().Done():
			// Client disconnected
			return nil
		}
	}

	// Only send completion event if the job actually completed.
	// Channel closure can also happen on buffer overflow (slow client),
	// in which case the job is still running.
	if info := j.OutputHandler.GetJobInfo(jobID); info != nil && !info.CompletedAt.IsZero() {
		fmt.Fprintf(w, "event: complete\ndata: done\n\n")
		flusher.Flush()
	}

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
