// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package jobs_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/jobs"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
	"github.com/stretchr/testify/assert"
)

func createTestProjectCmdContext(t *testing.T) command.ProjectContext {
	logger := logging.NewNoopLogger(t)
	return command.ProjectContext{
		BaseRepo: models.Repo{
			Name:  "test-repo",
			Owner: "test-org",
		},
		HeadRepo: models.Repo{
			Name:  "test-repo",
			Owner: "test-org",
		},
		Pull: models.PullRequest{
			Num:        1,
			HeadBranch: "main",
			BaseBranch: "main",
			Author:     "test-user",
			HeadCommit: "234r232432",
		},
		User: models.User{
			Username: "test-user",
		},
		Log:         logger,
		Workspace:   "myworkspace",
		RepoRelDir:  "test-dir",
		ProjectName: "test-project",
		JobID:       "1234",
	}
}

func createProjectCommandOutputHandler(t *testing.T) jobs.ProjectCommandOutputHandler {
	logger := logging.NewNoopLogger(t)
	prjCmdOutputChan := make(chan *jobs.ProjectCmdOutputLine)
	prjCmdOutputHandler := jobs.NewAsyncProjectCommandOutputHandler(
		prjCmdOutputChan,
		logger,
	)

	go func() {
		prjCmdOutputHandler.Handle()
	}()

	return prjCmdOutputHandler
}

func TestProjectCommandOutputHandler(t *testing.T) {
	Msg := "Test Terraform Output"
	ctx := createTestProjectCmdContext(t)

	t.Run("receive message from main channel", func(t *testing.T) {
		var wg sync.WaitGroup
		var expectedMsg string
		projectOutputHandler := createProjectCommandOutputHandler(t)

		ch := make(chan string, 1)

		// Register returns buffered lines (none yet) and registers the channel
		buffered, complete := projectOutputHandler.Register(ctx.JobID, ch)
		assert.Empty(t, buffered)
		assert.False(t, complete)

		wg.Add(1)

		// read from channel
		go func() {
			for msg := range ch {
				expectedMsg = msg
				wg.Done()
			}
		}()

		projectOutputHandler.Send(ctx, Msg, false)
		wg.Wait()
		close(ch)

		// Wait for the msg to be read.
		wg.Wait()
		Equals(t, expectedMsg, Msg)
	})

	t.Run("copies buffer to new channels", func(t *testing.T) {
		var wg sync.WaitGroup

		projectOutputHandler := createProjectCommandOutputHandler(t)

		// send first message to populate the buffer
		projectOutputHandler.Send(ctx, Msg, false)

		// Give time for async processing so the buffer is populated
		time.Sleep(10 * time.Millisecond)

		ch := make(chan string, 2)

		// Register returns the buffered line and registers the channel for live output
		buffered, complete := projectOutputHandler.Register(ctx.JobID, ch)
		assert.False(t, complete)
		assert.Equal(t, []string{Msg}, buffered)

		// Now collect live messages from the channel
		receivedMsgs := []string{}
		wg.Add(1)
		go func() {
			for msg := range ch {
				receivedMsgs = append(receivedMsgs, msg)
				if len(receivedMsgs) >= 1 {
					wg.Done()
				}
			}
		}()

		projectOutputHandler.Send(ctx, Msg, false)
		wg.Wait()
		close(ch)

		// The buffered line came from Register's return value, the live line from the channel
		assert.Len(t, receivedMsgs, 1)
		assert.Equal(t, Msg, receivedMsgs[0])
	})

	t.Run("clean up all jobs when PR is closed", func(t *testing.T) {
		var wg sync.WaitGroup
		projectOutputHandler := createProjectCommandOutputHandler(t)

		ch := make(chan string, 2)

		// Register returns buffered lines (none yet) and registers the channel
		buffered, complete := projectOutputHandler.Register(ctx.JobID, ch)
		assert.Empty(t, buffered)
		assert.False(t, complete)

		wg.Add(1)

		// read from channel
		go func() {
			for msg := range ch {
				if msg == "Complete" {
					wg.Done()
				}
			}
		}()

		projectOutputHandler.Send(ctx, Msg, false)
		projectOutputHandler.Send(ctx, "Complete", false)

		pullContext := jobs.PullInfo{
			PullNum:      ctx.Pull.Num,
			Repo:         ctx.BaseRepo.Name,
			RepoFullName: ctx.BaseRepo.FullName,
			ProjectName:  ctx.ProjectName,
			Path:         ctx.RepoRelDir,
			Workspace:    ctx.Workspace,
		}
		wg.Wait() // Must finish reading messages before cleaning up
		projectOutputHandler.CleanUp(pullContext)

		// Check all the resources are cleaned up.
		dfProjectOutputHandler, ok := projectOutputHandler.(*jobs.AsyncProjectCommandOutputHandler)
		assert.True(t, ok)

		assert.Empty(t, dfProjectOutputHandler.GetProjectOutputBuffer(ctx.JobID))
		assert.Empty(t, dfProjectOutputHandler.GetReceiverBufferForPull(ctx.JobID))
		assert.Empty(t, dfProjectOutputHandler.GetJobIDMapForPull(pullContext))
	})

	t.Run("mark operation status complete and close conn buffers for the job", func(t *testing.T) {
		projectOutputHandler := createProjectCommandOutputHandler(t)

		ch := make(chan string, 2)

		// Register returns buffered lines (none yet) and registers the channel
		buffered, complete := projectOutputHandler.Register(ctx.JobID, ch)
		assert.Empty(t, buffered)
		assert.False(t, complete)

		// read from channel
		go func() {
			for range ch { //revive:disable-line:empty-block
			}
		}()

		projectOutputHandler.Send(ctx, Msg, false)
		projectOutputHandler.Send(ctx, "", true)

		// Wait for the handler to process the message
		time.Sleep(10 * time.Millisecond)

		dfProjectOutputHandler, ok := projectOutputHandler.(*jobs.AsyncProjectCommandOutputHandler)
		assert.True(t, ok)

		outputBuffer := dfProjectOutputHandler.GetProjectOutputBuffer(ctx.JobID)
		assert.True(t, outputBuffer.OperationComplete)

		_, ok = (<-ch)
		assert.False(t, ok)

	})

	t.Run("close conn buffer after streaming logs for completed operation", func(t *testing.T) {
		projectOutputHandler := createProjectCommandOutputHandler(t)

		ch := make(chan string)

		// Register returns buffered lines (none yet) and registers the channel
		buffered, complete := projectOutputHandler.Register(ctx.JobID, ch)
		assert.Empty(t, buffered)
		assert.False(t, complete)

		// read from channel
		go func() {
			for range ch { //revive:disable-line:empty-block
			}
		}()

		projectOutputHandler.Send(ctx, Msg, false)
		projectOutputHandler.Send(ctx, "", true)

		// Wait for the handler to process the message
		time.Sleep(10 * time.Millisecond)

		ch2 := make(chan string, 2)

		// For a completed job, Register returns the buffered lines and complete=true.
		// The channel is closed immediately (not registered for live output).
		buffered2, complete2 := projectOutputHandler.Register(ctx.JobID, ch2)
		assert.True(t, complete2)
		assert.Equal(t, []string{Msg}, buffered2)

		// Channel should be closed since job is complete
		_, ok := <-ch2
		assert.False(t, ok)
	})
}

// TestRaceConditionPrevention tests that our fixes prevent the specific race conditions
func TestRaceConditionPrevention(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	prjCmdOutputChan := make(chan *jobs.ProjectCmdOutputLine)
	handler := jobs.NewAsyncProjectCommandOutputHandler(prjCmdOutputChan, logger)

	// Start the handler
	go handler.Handle()

	ctx := createTestProjectCmdContext(t)
	pullInfo := jobs.PullInfo{
		PullNum:      ctx.Pull.Num,
		Repo:         ctx.BaseRepo.Name,
		RepoFullName: ctx.BaseRepo.FullName,
		ProjectName:  ctx.ProjectName,
		Path:         ctx.RepoRelDir,
		Workspace:    ctx.Workspace,
	}

	t.Run("concurrent pullToJobMapping access", func(t *testing.T) {
		var wg sync.WaitGroup
		numGoroutines := 50

		// This test specifically targets the original race condition
		// that was fixed by using sync.Map for pullToJobMapping

		// Concurrent writers (Handle() method updates the mapping)
		for i := range numGoroutines {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				// Send message which triggers Handle() to update pullToJobMapping
				handler.Send(ctx, fmt.Sprintf("message-%d", id), false)
			}(i)
		}

		// Concurrent readers (GetPullToJobMapping() method reads the mapping)
		for range numGoroutines {
			wg.Go(func() {
				// This would race with Handle() before the sync.Map fix
				mappings := handler.GetPullToJobMapping()
				_ = mappings
			})
		}

		// Concurrent readers of GetJobIDMapForPull
		for range numGoroutines {
			wg.Go(func() {
				// This would also race with Handle() before the fix
				jobMap := handler.(*jobs.AsyncProjectCommandOutputHandler).GetJobIDMapForPull(pullInfo)
				_ = jobMap
			})
		}

		wg.Wait()
	})

	t.Run("concurrent buffer access", func(t *testing.T) {
		var wg sync.WaitGroup
		numGoroutines := 30

		// First populate some data
		handler.Send(ctx, "initial", false)
		time.Sleep(5 * time.Millisecond)

		// Test the race condition we fixed in GetProjectOutputBuffer
		for range numGoroutines {
			wg.Go(func() {
				// This would race with completeJob() before the RLock fix
				buffer := handler.(*jobs.AsyncProjectCommandOutputHandler).GetProjectOutputBuffer(ctx.JobID)
				_ = buffer
			})
		}

		// Concurrent operations that modify the buffer
		for i := range numGoroutines {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				if id%10 == 0 {
					// Occasionally complete a job to test completeJob() race
					handler.Send(ctx, "", true)
				} else {
					handler.Send(ctx, "test", false)
				}
			}(i)
		}

		wg.Wait()
	})

	// Clean up
	close(prjCmdOutputChan)
}

// TestHighConcurrencyStress performs stress testing with many concurrent operations
func TestProjectCommandOutputHandler_GracefulClose(t *testing.T) {
	t.Run("closes channel when buffer full instead of silent delete", func(t *testing.T) {
		logger := logging.NewNoopLogger(t)
		prjCmdOutputChan := make(chan *jobs.ProjectCmdOutputLine)
		prjCmdOutputHandler := jobs.NewAsyncProjectCommandOutputHandler(
			prjCmdOutputChan,
			logger,
		)

		go prjCmdOutputHandler.Handle()

		ctx := createTestProjectCmdContext(t)

		// Register a channel with size 1 that we won't read from
		slowCh := make(chan string, 1)
		buffered, complete := prjCmdOutputHandler.Register(ctx.JobID, slowCh)
		assert.Empty(t, buffered)
		assert.False(t, complete)

		// Send messages until buffer would overflow
		// First message fills the channel
		prjCmdOutputHandler.Send(ctx, "msg1", false)
		// Second message should trigger close (not silent delete)
		prjCmdOutputHandler.Send(ctx, "msg2", false)

		// Give time for async processing
		time.Sleep(100 * time.Millisecond)

		// Channel should be closed, not just deleted
		select {
		case _, ok := <-slowCh:
			if ok {
				// Read the first message, try again
				_, ok = <-slowCh
			}
			Assert(t, !ok, "channel should be closed after buffer overflow")
		case <-time.After(500 * time.Millisecond):
			t.Fatal("channel was not closed")
		}
	})
}

// TestRegisterLargeBufferNoDeadlock verifies that Register with >1000 buffered lines
// does not deadlock. Before the fix, addChan sent buffered lines through the channel
// while holding locks, which would block if the buffer exceeded channel capacity.
func TestRegisterLargeBufferNoDeadlock(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	prjCmdOutputChan := make(chan *jobs.ProjectCmdOutputLine)
	handler := jobs.NewAsyncProjectCommandOutputHandler(prjCmdOutputChan, logger)

	go handler.Handle()

	ctx := createTestProjectCmdContext(t)

	// Write 1500 lines (exceeding channel capacity of 1000)
	numLines := 1500
	for i := range numLines {
		handler.Send(ctx, fmt.Sprintf("line-%d", i), false)
	}

	// Give time for async processing
	time.Sleep(50 * time.Millisecond)

	// Register should return immediately with all buffered lines (no deadlock)
	done := make(chan struct{})
	go func() {
		ch := make(chan string, 1000)
		buffered, complete := handler.Register(ctx.JobID, ch)
		assert.Len(t, buffered, numLines)
		assert.False(t, complete)
		assert.Equal(t, "line-0", buffered[0])
		assert.Equal(t, fmt.Sprintf("line-%d", numLines-1), buffered[numLines-1])

		// Channel should still be open for live output
		handler.Deregister(ctx.JobID, ch)
		close(done)
	}()

	select {
	case <-done:
		// Success - no deadlock
	case <-time.After(5 * time.Second):
		t.Fatal("Register deadlocked with large buffer")
	}

	close(prjCmdOutputChan)
}

// TestRegisterThenWriteNoDuplicates verifies that after Register returns buffered lines,
// subsequent writeLogLine calls send only new lines through the channel (no duplication).
func TestRegisterThenWriteNoDuplicates(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	prjCmdOutputChan := make(chan *jobs.ProjectCmdOutputLine)
	handler := jobs.NewAsyncProjectCommandOutputHandler(prjCmdOutputChan, logger)

	go handler.Handle()

	ctx := createTestProjectCmdContext(t)

	// Write initial lines
	handler.Send(ctx, "pre-register-1", false)
	handler.Send(ctx, "pre-register-2", false)
	time.Sleep(10 * time.Millisecond)

	ch := make(chan string, 100)
	buffered, complete := handler.Register(ctx.JobID, ch)
	assert.False(t, complete)
	assert.Equal(t, []string{"pre-register-1", "pre-register-2"}, buffered)

	// Now send more lines after registration
	handler.Send(ctx, "post-register-1", false)
	handler.Send(ctx, "post-register-2", false)

	// Read from channel -- should only get post-register lines
	var received []string
	timeout := time.After(2 * time.Second)
	for i := 0; i < 2; i++ {
		select {
		case line := <-ch:
			received = append(received, line)
		case <-timeout:
			t.Fatalf("timed out waiting for line %d", i)
		}
	}
	assert.Equal(t, []string{"post-register-1", "post-register-2"}, received)

	handler.Deregister(ctx.JobID, ch)
	close(prjCmdOutputChan)
}

func TestHighConcurrencyStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	logger := logging.NewNoopLogger(t)
	prjCmdOutputChan := make(chan *jobs.ProjectCmdOutputLine)
	handler := jobs.NewAsyncProjectCommandOutputHandler(prjCmdOutputChan, logger)

	// Start the handler
	go handler.Handle()

	var wg sync.WaitGroup
	numWorkers := 20
	operationsPerWorker := 100

	// Multiple workers performing mixed operations
	wg.Add(numWorkers)
	for worker := range numWorkers {
		go func(workerID int) {
			defer wg.Done()

			ctx := createTestProjectCmdContext(t)
			ctx.JobID = "worker-job-" + fmt.Sprintf("%d", workerID)
			ctx.Pull.Num = workerID

			pullInfo := jobs.PullInfo{
				PullNum:      ctx.Pull.Num,
				Repo:         ctx.BaseRepo.Name,
				RepoFullName: ctx.BaseRepo.FullName,
				ProjectName:  ctx.ProjectName,
				Path:         ctx.RepoRelDir,
				Workspace:    ctx.Workspace,
			}

			for op := range operationsPerWorker {
				switch op % 6 {
				case 0:
					// Send messages
					handler.Send(ctx, "stress test message", false)
				case 1:
					// Read pull to job mapping
					mappings := handler.GetPullToJobMapping()
					_ = mappings
				case 2:
					// Read job ID map for pull
					jobMap := handler.(*jobs.AsyncProjectCommandOutputHandler).GetJobIDMapForPull(pullInfo)
					_ = jobMap
				case 3:
					// Read project output buffer
					buffer := handler.(*jobs.AsyncProjectCommandOutputHandler).GetProjectOutputBuffer(ctx.JobID)
					_ = buffer
				case 4:
					// Read receiver buffer
					receivers := handler.(*jobs.AsyncProjectCommandOutputHandler).GetReceiverBufferForPull(ctx.JobID)
					_ = receivers
				case 5:
					// Occasional cleanup
					if op%20 == 0 {
						handler.CleanUp(pullInfo)
					}
				}
			}
		}(worker)
	}

	wg.Wait()
	close(prjCmdOutputChan)
}
