// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	"bufio"
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/runatlantis/atlantis/server/controllers"
	"github.com/runatlantis/atlantis/server/jobs"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
	tally "github.com/uber-go/tally/v4"
)

func TestJobsController_GetProjectJobsSSE(t *testing.T) {
	t.Run("returns 404 for unknown job", func(t *testing.T) {
		outputHandler := jobs.NewTestOutputHandler()
		logger := logging.NewNoopLogger(t)
		scope := tally.NewTestScope("test", nil)

		jc := controllers.JobsController{
			Logger:        logger,
			OutputHandler: outputHandler,
			StatsScope:    scope,
		}

		req, _ := http.NewRequest("GET", "/jobs/unknown-job-id/stream", bytes.NewBuffer(nil))
		req = mux.SetURLVars(req, map[string]string{"job-id": "unknown-job-id"})
		w := httptest.NewRecorder()

		jc.GetProjectJobsSSE(w, req)

		ResponseContains(t, w, http.StatusNotFound, "Job not found")
	})

	t.Run("sets correct SSE headers", func(t *testing.T) {
		outputHandler := jobs.NewTestOutputHandler()
		// Pre-register a job so it exists
		outputHandler.SetJobExists("test-job-id", true)

		logger := logging.NewNoopLogger(t)
		scope := tally.NewTestScope("test", nil)

		jc := controllers.JobsController{
			Logger:        logger,
			OutputHandler: outputHandler,
			StatsScope:    scope,
		}

		req, _ := http.NewRequest("GET", "/jobs/test-job-id/stream", bytes.NewBuffer(nil))
		req = mux.SetURLVars(req, map[string]string{"job-id": "test-job-id"})
		w := httptest.NewRecorder()

		// Complete the job immediately so the handler returns
		go func() {
			time.Sleep(10 * time.Millisecond)
			outputHandler.CompleteJob("test-job-id")
		}()

		jc.GetProjectJobsSSE(w, req)

		result := w.Result()
		Equals(t, "text/event-stream", result.Header.Get("Content-Type"))
		Equals(t, "no-cache", result.Header.Get("Cache-Control"))
		Equals(t, "keep-alive", result.Header.Get("Connection"))
		Equals(t, "no", result.Header.Get("X-Accel-Buffering"))
	})

	t.Run("streams output as SSE events", func(t *testing.T) {
		outputHandler := jobs.NewTestOutputHandler()
		outputHandler.SetJobExists("test-job-id", true)

		logger := logging.NewNoopLogger(t)
		scope := tally.NewTestScope("test", nil)

		jc := controllers.JobsController{
			Logger:        logger,
			OutputHandler: outputHandler,
			StatsScope:    scope,
		}

		req, _ := http.NewRequest("GET", "/jobs/test-job-id/stream", bytes.NewBuffer(nil))
		req = mux.SetURLVars(req, map[string]string{"job-id": "test-job-id"})
		w := httptest.NewRecorder()

		// Send test lines and complete the job in a goroutine
		go func() {
			time.Sleep(10 * time.Millisecond)
			outputHandler.SendTestLine("test-job-id", "line 1")
			time.Sleep(5 * time.Millisecond)
			outputHandler.SendTestLine("test-job-id", "line 2")
			time.Sleep(5 * time.Millisecond)
			outputHandler.CompleteJob("test-job-id")
		}()

		jc.GetProjectJobsSSE(w, req)

		body := w.Body.String()

		// Verify SSE format: each line should be "data: <content>\n\n"
		scanner := bufio.NewScanner(strings.NewReader(body))
		var dataLines []string
		var hasComplete bool

		for scanner.Scan() {
			line := scanner.Text()
			if data, found := strings.CutPrefix(line, "data: "); found {
				dataLines = append(dataLines, data)
			}
			if line == "event: complete" {
				hasComplete = true
			}
		}

		Assert(t, len(dataLines) >= 2, "expected at least 2 data lines, got %d", len(dataLines))
		Assert(t, hasComplete, "expected 'event: complete' in response")
		Contains(t, "line 1", dataLines)
		Contains(t, "line 2", dataLines)
	})
}
