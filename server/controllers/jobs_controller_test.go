// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
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

func TestJobsController_SSEConnectionLimit(t *testing.T) {
	t.Run("rejects connections beyond limit", func(t *testing.T) {
		outputHandler := jobs.NewTestOutputHandler()
		outputHandler.SetJobExists("job-1", true)
		outputHandler.SetJobExists("job-2", true)
		outputHandler.SetJobExists("job-3", true)

		logger := logging.NewNoopLogger(t)
		scope := tally.NewTestScope("test", nil)

		jc := controllers.JobsController{
			Logger:            logger,
			OutputHandler:     outputHandler,
			StatsScope:        scope,
			SSEMaxConnections: 2,
		}

		// Start two connections that will block (they hold slots)
		var wg sync.WaitGroup
		wg.Add(2)
		for i := 1; i <= 2; i++ {
			go func(jobNum int) {
				defer wg.Done()
				jobID := "job-" + string(rune('0'+jobNum))
				req, _ := http.NewRequest("GET", "/jobs/"+jobID+"/stream", bytes.NewBuffer(nil))
				req = mux.SetURLVars(req, map[string]string{"job-id": jobID})
				w := httptest.NewRecorder()
				jc.GetProjectJobsSSE(w, req)
			}(i)
		}

		// Wait for both connections to register with the output handler
		time.Sleep(50 * time.Millisecond)

		// Third connection should be rejected with 503
		req, _ := http.NewRequest("GET", "/jobs/job-3/stream", bytes.NewBuffer(nil))
		req = mux.SetURLVars(req, map[string]string{"job-id": "job-3"})
		w := httptest.NewRecorder()
		jc.GetProjectJobsSSE(w, req)

		ResponseContains(t, w, http.StatusServiceUnavailable, "Too many active connections")

		// Clean up: complete the blocking jobs so goroutines finish
		outputHandler.CompleteJob("job-1")
		outputHandler.CompleteJob("job-2")
		wg.Wait()
	})

	t.Run("allows connections when limit is zero (disabled)", func(t *testing.T) {
		outputHandler := jobs.NewTestOutputHandler()
		outputHandler.SetJobExists("job-1", true)

		logger := logging.NewNoopLogger(t)
		scope := tally.NewTestScope("test", nil)

		jc := controllers.JobsController{
			Logger:            logger,
			OutputHandler:     outputHandler,
			StatsScope:        scope,
			SSEMaxConnections: 0, // Disabled
		}

		// Complete the job so the handler returns quickly
		go func() {
			time.Sleep(10 * time.Millisecond)
			outputHandler.CompleteJob("job-1")
		}()

		req, _ := http.NewRequest("GET", "/jobs/job-1/stream", bytes.NewBuffer(nil))
		req = mux.SetURLVars(req, map[string]string{"job-id": "job-1"})
		w := httptest.NewRecorder()
		jc.GetProjectJobsSSE(w, req)

		// Should succeed, not get 503
		result := w.Result()
		Assert(t, result.StatusCode != http.StatusServiceUnavailable, "expected connection to be allowed when limit is 0, got %d", result.StatusCode)
	})

	t.Run("frees slot after connection ends", func(t *testing.T) {
		outputHandler := jobs.NewTestOutputHandler()
		outputHandler.SetJobExists("job-1", true)
		outputHandler.SetJobExists("job-2", true)

		logger := logging.NewNoopLogger(t)
		scope := tally.NewTestScope("test", nil)

		jc := controllers.JobsController{
			Logger:            logger,
			OutputHandler:     outputHandler,
			StatsScope:        scope,
			SSEMaxConnections: 1,
		}

		// First connection: complete it so it frees its slot
		go func() {
			time.Sleep(10 * time.Millisecond)
			outputHandler.CompleteJob("job-1")
		}()

		req1, _ := http.NewRequest("GET", "/jobs/job-1/stream", bytes.NewBuffer(nil))
		req1 = mux.SetURLVars(req1, map[string]string{"job-id": "job-1"})
		w1 := httptest.NewRecorder()
		jc.GetProjectJobsSSE(w1, req1)

		// First connection is done; the slot should be freed.
		// Second connection should succeed.
		go func() {
			time.Sleep(10 * time.Millisecond)
			outputHandler.CompleteJob("job-2")
		}()

		req2, _ := http.NewRequest("GET", "/jobs/job-2/stream", bytes.NewBuffer(nil))
		req2 = mux.SetURLVars(req2, map[string]string{"job-id": "job-2"})
		w2 := httptest.NewRecorder()
		jc.GetProjectJobsSSE(w2, req2)

		result := w2.Result()
		Assert(t, result.StatusCode != http.StatusServiceUnavailable, "expected connection to succeed after slot freed, got %d", result.StatusCode)
	})
}

func TestJobsController_SSEEventIDs(t *testing.T) {
	t.Run("backfill lines include sequential id fields", func(t *testing.T) {
		outputHandler := jobs.NewTestOutputHandler()
		outputHandler.SetJobExists("test-job", true)
		outputHandler.SetBufferedLines("test-job", []string{"line-a", "line-b", "line-c"})

		logger := logging.NewNoopLogger(t)
		scope := tally.NewTestScope("test", nil)

		jc := controllers.JobsController{
			Logger:        logger,
			OutputHandler: outputHandler,
			StatsScope:    scope,
		}

		req, _ := http.NewRequest("GET", "/jobs/test-job/stream", bytes.NewBuffer(nil))
		req = mux.SetURLVars(req, map[string]string{"job-id": "test-job"})
		w := httptest.NewRecorder()

		// Complete the job so the handler returns
		go func() {
			time.Sleep(10 * time.Millisecond)
			outputHandler.CompleteJob("test-job")
		}()

		jc.GetProjectJobsSSE(w, req)

		body := w.Body.String()

		// Verify sequential id fields: id: 0, id: 1, id: 2
		for i := 0; i < 3; i++ {
			expected := fmt.Sprintf("id: %d\n", i)
			Assert(t, strings.Contains(body, expected),
				"expected %q in body, got:\n%s", expected, body)
		}
	})

	t.Run("live lines include sequential id fields continuing from backfill", func(t *testing.T) {
		outputHandler := jobs.NewTestOutputHandler()
		outputHandler.SetJobExists("test-job", true)
		outputHandler.SetBufferedLines("test-job", []string{"buffered-0", "buffered-1"})

		logger := logging.NewNoopLogger(t)
		scope := tally.NewTestScope("test", nil)

		jc := controllers.JobsController{
			Logger:        logger,
			OutputHandler: outputHandler,
			StatsScope:    scope,
		}

		req, _ := http.NewRequest("GET", "/jobs/test-job/stream", bytes.NewBuffer(nil))
		req = mux.SetURLVars(req, map[string]string{"job-id": "test-job"})
		w := httptest.NewRecorder()

		go func() {
			time.Sleep(10 * time.Millisecond)
			outputHandler.SendTestLine("test-job", "live-line")
			time.Sleep(5 * time.Millisecond)
			outputHandler.CompleteJob("test-job")
		}()

		jc.GetProjectJobsSSE(w, req)

		body := w.Body.String()

		// Backfill had 2 lines (id: 0, id: 1), so live line should be id: 2
		Assert(t, strings.Contains(body, "id: 0\n"), "expected id: 0 for first buffered line")
		Assert(t, strings.Contains(body, "id: 1\n"), "expected id: 1 for second buffered line")
		Assert(t, strings.Contains(body, "id: 2\n"), "expected id: 2 for live line")
	})
}

func TestJobsController_SSEReconnection(t *testing.T) {
	t.Run("Last-Event-ID skips already sent lines", func(t *testing.T) {
		outputHandler := jobs.NewTestOutputHandler()
		outputHandler.SetJobExists("test-job", true)
		outputHandler.SetBufferedLines("test-job", []string{"line-0", "line-1", "line-2", "line-3"})

		logger := logging.NewNoopLogger(t)
		scope := tally.NewTestScope("test", nil)

		jc := controllers.JobsController{
			Logger:        logger,
			OutputHandler: outputHandler,
			StatsScope:    scope,
		}

		// Client reconnects having received through id 1 (lines 0 and 1)
		req, _ := http.NewRequest("GET", "/jobs/test-job/stream", bytes.NewBuffer(nil))
		req = mux.SetURLVars(req, map[string]string{"job-id": "test-job"})
		req.Header.Set("Last-Event-ID", "1")
		w := httptest.NewRecorder()

		go func() {
			time.Sleep(10 * time.Millisecond)
			outputHandler.CompleteJob("test-job")
		}()

		jc.GetProjectJobsSSE(w, req)

		body := w.Body.String()

		// Lines 0 and 1 should NOT appear in the output
		Assert(t, !strings.Contains(body, "data: line-0\n"),
			"line-0 should be skipped on reconnect, got:\n%s", body)
		Assert(t, !strings.Contains(body, "data: line-1\n"),
			"line-1 should be skipped on reconnect, got:\n%s", body)

		// Lines 2 and 3 SHOULD appear
		Assert(t, strings.Contains(body, "id: 2\n"),
			"expected id: 2 in output, got:\n%s", body)
		Assert(t, strings.Contains(body, "data: line-2\n"),
			"expected line-2 in output, got:\n%s", body)
		Assert(t, strings.Contains(body, "id: 3\n"),
			"expected id: 3 in output, got:\n%s", body)
		Assert(t, strings.Contains(body, "data: line-3\n"),
			"expected line-3 in output, got:\n%s", body)
	})

	t.Run("invalid Last-Event-ID sends full backfill", func(t *testing.T) {
		outputHandler := jobs.NewTestOutputHandler()
		outputHandler.SetJobExists("test-job", true)
		outputHandler.SetBufferedLines("test-job", []string{"line-0", "line-1"})

		logger := logging.NewNoopLogger(t)
		scope := tally.NewTestScope("test", nil)

		jc := controllers.JobsController{
			Logger:        logger,
			OutputHandler: outputHandler,
			StatsScope:    scope,
		}

		for _, lastID := range []string{"not-a-number", "-1", "abc123"} {
			req, _ := http.NewRequest("GET", "/jobs/test-job/stream", bytes.NewBuffer(nil))
			req = mux.SetURLVars(req, map[string]string{"job-id": "test-job"})
			req.Header.Set("Last-Event-ID", lastID)
			w := httptest.NewRecorder()

			go func() {
				time.Sleep(10 * time.Millisecond)
				outputHandler.CompleteJob("test-job")
			}()

			jc.GetProjectJobsSSE(w, req)

			body := w.Body.String()

			// Full backfill should be sent
			Assert(t, strings.Contains(body, "id: 0\n"),
				"Last-Event-ID=%q: expected id: 0 in output, got:\n%s", lastID, body)
			Assert(t, strings.Contains(body, "data: line-0\n"),
				"Last-Event-ID=%q: expected line-0 in output, got:\n%s", lastID, body)
			Assert(t, strings.Contains(body, "id: 1\n"),
				"Last-Event-ID=%q: expected id: 1 in output, got:\n%s", lastID, body)
			Assert(t, strings.Contains(body, "data: line-1\n"),
				"Last-Event-ID=%q: expected line-1 in output, got:\n%s", lastID, body)

			// Re-register job for next iteration (CompleteJob clears it)
			outputHandler.SetJobExists("test-job", true)
			outputHandler.SetBufferedLines("test-job", []string{"line-0", "line-1"})
		}
	})

	t.Run("Last-Event-ID exceeding buffer sends no backfill but streams live", func(t *testing.T) {
		outputHandler := jobs.NewTestOutputHandler()
		outputHandler.SetJobExists("test-job", true)
		outputHandler.SetBufferedLines("test-job", []string{"line-0", "line-1"})

		logger := logging.NewNoopLogger(t)
		scope := tally.NewTestScope("test", nil)

		jc := controllers.JobsController{
			Logger:        logger,
			OutputHandler: outputHandler,
			StatsScope:    scope,
		}

		// Client says it received through id 99, but buffer only has 2 lines
		req, _ := http.NewRequest("GET", "/jobs/test-job/stream", bytes.NewBuffer(nil))
		req = mux.SetURLVars(req, map[string]string{"job-id": "test-job"})
		req.Header.Set("Last-Event-ID", "99")
		w := httptest.NewRecorder()

		go func() {
			time.Sleep(10 * time.Millisecond)
			outputHandler.SendTestLine("test-job", "live-line")
			time.Sleep(5 * time.Millisecond)
			outputHandler.CompleteJob("test-job")
		}()

		jc.GetProjectJobsSSE(w, req)

		body := w.Body.String()

		// No backfill lines should be sent
		Assert(t, !strings.Contains(body, "data: line-0\n"),
			"expected no backfill line-0, got:\n%s", body)
		Assert(t, !strings.Contains(body, "data: line-1\n"),
			"expected no backfill line-1, got:\n%s", body)

		// Live line should still be streamed (with lineNum = len(buffered) = 2)
		Assert(t, strings.Contains(body, "id: 2\n"),
			"expected live line with id: 2, got:\n%s", body)
		Assert(t, strings.Contains(body, "data: live-line\n"),
			"expected live-line in output, got:\n%s", body)
	})
}

func TestJobsController_SSEIdleTimeout(t *testing.T) {
	t.Run("disconnects after idle timeout", func(t *testing.T) {
		outputHandler := jobs.NewTestOutputHandler()
		outputHandler.SetJobExists("idle-job", true)

		logger := logging.NewNoopLogger(t)
		scope := tally.NewTestScope("test", nil)

		jc := controllers.JobsController{
			Logger:         logger,
			OutputHandler:  outputHandler,
			StatsScope:     scope,
			SSEIdleTimeout: 50 * time.Millisecond,
		}

		req, _ := http.NewRequest("GET", "/jobs/idle-job/stream", bytes.NewBuffer(nil))
		req = mux.SetURLVars(req, map[string]string{"job-id": "idle-job"})
		w := httptest.NewRecorder()

		start := time.Now()
		jc.GetProjectJobsSSE(w, req)
		elapsed := time.Since(start)

		// Should have returned after roughly the idle timeout, not hanging forever
		Assert(t, elapsed < 2*time.Second, "expected handler to return within 2s due to idle timeout, took %s", elapsed)
		// And it should have taken at least the idle timeout duration
		Assert(t, elapsed >= 50*time.Millisecond, "expected handler to wait at least 50ms, took %s", elapsed)

		// Clean up -- mark job complete so test output handler is tidy
		outputHandler.CompleteJob("idle-job")
	})

	t.Run("idle timer resets on activity", func(t *testing.T) {
		outputHandler := jobs.NewTestOutputHandler()
		outputHandler.SetJobExists("active-job", true)

		logger := logging.NewNoopLogger(t)
		scope := tally.NewTestScope("test", nil)

		jc := controllers.JobsController{
			Logger:         logger,
			OutputHandler:  outputHandler,
			StatsScope:     scope,
			SSEIdleTimeout: 100 * time.Millisecond,
		}

		req, _ := http.NewRequest("GET", "/jobs/active-job/stream", bytes.NewBuffer(nil))
		req = mux.SetURLVars(req, map[string]string{"job-id": "active-job"})
		w := httptest.NewRecorder()

		// Send a line every 60ms (within the 100ms timeout) three times, then stop
		go func() {
			time.Sleep(60 * time.Millisecond)
			outputHandler.SendTestLine("active-job", "keep alive 1")
			time.Sleep(60 * time.Millisecond)
			outputHandler.SendTestLine("active-job", "keep alive 2")
			time.Sleep(60 * time.Millisecond)
			outputHandler.SendTestLine("active-job", "keep alive 3")
			// Now stop sending -- idle timeout should fire after 100ms
		}()

		start := time.Now()
		jc.GetProjectJobsSSE(w, req)
		elapsed := time.Since(start)

		// The handler should have stayed alive through all three keep-alive messages
		// (3 * 60ms = 180ms) plus the final idle timeout (100ms) = ~280ms total
		Assert(t, elapsed >= 250*time.Millisecond, "expected handler to stay alive through keep-alive messages, took %s", elapsed)

		body := w.Body.String()
		Assert(t, strings.Contains(body, "keep alive 1"), "expected 'keep alive 1' in output")
		Assert(t, strings.Contains(body, "keep alive 3"), "expected 'keep alive 3' in output")

		// Clean up
		outputHandler.CompleteJob("active-job")
	})
}
