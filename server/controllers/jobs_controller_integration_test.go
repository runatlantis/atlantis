// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

//go:build integration

package controllers_test

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func TestSSEIntegration(t *testing.T) {
	t.Run("reconnecting client receives backfilled output", func(t *testing.T) {
		logger := logging.NewNoopLogger(t)
		prjCmdOutputChan := make(chan *jobs.ProjectCmdOutputLine)
		outputHandler := jobs.NewAsyncProjectCommandOutputHandler(prjCmdOutputChan, logger)
		go outputHandler.Handle()

		// Type assert to get access to test methods
		asyncHandler := outputHandler.(*jobs.AsyncProjectCommandOutputHandler)

		router := mux.NewRouter()
		scope := tally.NewTestScope("test", nil)
		ctrl := &controllers.JobsController{
			AtlantisVersion: "test",
			AtlantisURL:     &url.URL{Path: ""},
			Logger:          logger,
			OutputHandler:   outputHandler,
			KeyGenerator:    controllers.JobIDKeyGenerator{},
			StatsScope:      scope,
		}
		router.HandleFunc("/jobs/{job-id}/stream", ctrl.GetProjectJobsSSE).Methods("GET")

		server := httptest.NewServer(router)
		defer server.Close()

		jobID := "backfill-test"

		// Send some output to create the job
		asyncHandler.SendTestLine(jobID, "line 1")
		asyncHandler.SendTestLine(jobID, "line 2")
		asyncHandler.SendTestLine(jobID, "line 3")

		// Give time for buffering
		time.Sleep(50 * time.Millisecond)

		// Connect first client
		ctx1, cancel1 := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel1()

		req1, _ := http.NewRequestWithContext(ctx1, "GET", server.URL+"/jobs/"+jobID+"/stream", nil)
		resp1, err := http.DefaultClient.Do(req1)
		Ok(t, err)

		// Read first 3 lines
		reader := bufio.NewReader(resp1.Body)
		var lines1 []string
		for i := 0; i < 6; i++ { // 6 because each data line is followed by empty line
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			if strings.HasPrefix(line, "data: ") {
				lines1 = append(lines1, strings.TrimPrefix(strings.TrimSpace(line), "data: "))
			}
		}
		resp1.Body.Close()
		cancel1() // Disconnect

		// Verify we got 3 lines
		Equals(t, 3, len(lines1))

		// Send more output while disconnected
		asyncHandler.SendTestLine(jobID, "line 4")
		time.Sleep(50 * time.Millisecond)

		// Reconnect - should get ALL lines (backfill)
		ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel2()

		req2, _ := http.NewRequestWithContext(ctx2, "GET", server.URL+"/jobs/"+jobID+"/stream", nil)
		resp2, err := http.DefaultClient.Do(req2)
		Ok(t, err)

		// Read all lines
		reader2 := bufio.NewReader(resp2.Body)
		var lines2 []string
		for i := 0; i < 10; i++ { // Read up to 10 lines
			line, err := reader2.ReadString('\n')
			if err != nil {
				break
			}
			if strings.HasPrefix(line, "data: ") {
				lines2 = append(lines2, strings.TrimPrefix(strings.TrimSpace(line), "data: "))
			}
		}
		resp2.Body.Close()

		// Should have received all 4 lines on reconnect (backfill)
		Equals(t, 4, len(lines2))
		Equals(t, "line 1", lines2[0])
		Equals(t, "line 4", lines2[3])
	})
}
