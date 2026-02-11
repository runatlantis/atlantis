// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package jobs

import (
	"sync"
	"time"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

// TestOutputHandler is a test implementation of ProjectCommandOutputHandler
// that allows injecting test output and controlling job completion.
type TestOutputHandler struct {
	mu sync.RWMutex

	// jobExists tracks which job IDs exist
	jobExists map[string]bool

	// receivers tracks registered channels per job ID
	receivers map[string][]chan string

	// completedJobs tracks completion times for GetJobInfo
	completedJobs map[string]time.Time

	// bufferedLines holds pre-set buffered output returned by Register
	bufferedLines map[string][]string
}

// NewTestOutputHandler creates a new TestOutputHandler for testing.
func NewTestOutputHandler() *TestOutputHandler {
	return &TestOutputHandler{
		jobExists:     make(map[string]bool),
		receivers:     make(map[string][]chan string),
		completedJobs: make(map[string]time.Time),
	}
}

// SetBufferedLines sets buffered lines that Register will return for a job.
// This simulates a job that already has output in its buffer.
func (t *TestOutputHandler) SetBufferedLines(jobID string, lines []string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.bufferedLines == nil {
		t.bufferedLines = make(map[string][]string)
	}
	t.bufferedLines[jobID] = lines
}

// SetJobExists sets whether a job ID exists (for IsKeyExists).
func (t *TestOutputHandler) SetJobExists(jobID string, exists bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.jobExists[jobID] = exists
}

// SendTestLine sends a line to all registered receivers for a job.
func (t *TestOutputHandler) SendTestLine(jobID string, line string) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, ch := range t.receivers[jobID] {
		select {
		case ch <- line:
		default:
			// Channel full or blocked, skip
		}
	}
}

// CompleteJob marks a job as complete and closes all registered channels.
func (t *TestOutputHandler) CompleteJob(jobID string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.completedJobs[jobID] = time.Now()
	for _, ch := range t.receivers[jobID] {
		close(ch)
	}
	delete(t.receivers, jobID)
}

// Register registers a channel to receive output for a job.
// If the job is already complete, the channel is closed immediately and complete=true is returned.
func (t *TestOutputHandler) Register(jobID string, receiver chan string) ([]string, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	buffered := t.bufferedLines[jobID]
	if _, completed := t.completedJobs[jobID]; completed {
		close(receiver)
		return buffered, true
	}
	t.receivers[jobID] = append(t.receivers[jobID], receiver)
	return buffered, false
}

// Deregister removes a channel from receiving output for a job.
func (t *TestOutputHandler) Deregister(jobID string, receiver chan string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	channels := t.receivers[jobID]
	for i, ch := range channels {
		if ch == receiver {
			t.receivers[jobID] = append(channels[:i], channels[i+1:]...)
			break
		}
	}
}

// IsKeyExists returns whether a job ID exists.
func (t *TestOutputHandler) IsKeyExists(key string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.jobExists[key]
}

// Send is a no-op for testing (implements interface).
func (t *TestOutputHandler) Send(_ command.ProjectContext, _ string, _ bool) {}

// SendWorkflowHook is a no-op for testing (implements interface).
func (t *TestOutputHandler) SendWorkflowHook(_ models.WorkflowHookCommandContext, _ string, _ bool) {}

// Handle is a no-op for testing (implements interface).
func (t *TestOutputHandler) Handle() {}

// CleanUp is a no-op for testing (implements interface).
func (t *TestOutputHandler) CleanUp(_ PullInfo) {}

// GetPullToJobMapping returns an empty slice for testing (implements interface).
func (t *TestOutputHandler) GetPullToJobMapping() []PullInfoWithJobIDs {
	return []PullInfoWithJobIDs{}
}

// GetProjectOutputBuffer returns an empty buffer for testing (implements interface).
func (t *TestOutputHandler) GetProjectOutputBuffer(_ string) OutputBuffer {
	return OutputBuffer{}
}

// GetJobInfo returns job info if the job has been completed.
func (t *TestOutputHandler) GetJobInfo(jobID string) *JobIDInfo {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if completedAt, ok := t.completedJobs[jobID]; ok {
		return &JobIDInfo{
			JobID:       jobID,
			CompletedAt: completedAt,
		}
	}
	return nil
}
