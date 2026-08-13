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

		// register channel and backfill from buffer
		// Note: We call this synchronously because otherwise
		// there could be a race where we are unable to register the channel
		// before sending messages due to the way we lock our buffer memory cache
		projectOutputHandler.Register(ctx.JobID, ch)

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

		// send first message to populated the buffer
		projectOutputHandler.Send(ctx, Msg, false)

		ch := make(chan string, 2)

		receivedMsgs := []string{}

		wg.Add(1)
		// read from channel asynchronously
		go func() {
			for msg := range ch {
				receivedMsgs = append(receivedMsgs, msg)

				// we're only expecting two messages here.
				if len(receivedMsgs) >= 2 {
					wg.Done()
				}
			}
		}()
		// register channel and backfill from buffer
		// Note: We call this synchronously because otherwise
		// there could be a race where we are unable to register the channel
		// before sending messages due to the way we lock our buffer memory cache
		projectOutputHandler.Register(ctx.JobID, ch)

		projectOutputHandler.Send(ctx, Msg, false)
		wg.Wait()
		close(ch)

		expectedMsgs := []string{Msg, Msg}
		assert.Len(t, receivedMsgs, len(expectedMsgs))
		for i := range expectedMsgs {
			assert.Equal(t, expectedMsgs[i], receivedMsgs[i])
		}
	})

	t.Run("clean up all jobs when PR is closed", func(t *testing.T) {
		var wg sync.WaitGroup
		projectOutputHandler := createProjectCommandOutputHandler(t)

		ch := make(chan string, 2)

		// register channel and backfill from buffer
		// Note: We call this synchronously because otherwise
		// there could be a race where we are unable to register the channel
		// before sending messages due to the way we lock our buffer memory cache
		projectOutputHandler.Register(ctx.JobID, ch)

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

		// register channel and backfill from buffer
		// Note: We call this synchronously because otherwise
		// there could be a race where we are unable to register the channel
		// before sending messages due to the way we lock our buffer memory cache
		projectOutputHandler.Register(ctx.JobID, ch)

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

		// register channel and backfill from buffer
		// Note: We call this synchronously because otherwise
		// there could be a race where we are unable to register the channel
		// before sending messages due to the way we lock our buffer memory cache
		projectOutputHandler.Register(ctx.JobID, ch)

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
		opComplete := make(chan bool)

		// buffer channel will be closed immediately after logs are streamed
		go func() {
			for range ch2 { //revive:disable-line:empty-block
			}
			opComplete <- true
		}()

		projectOutputHandler.Register(ctx.JobID, ch2)

		assert.True(t, <-opComplete)
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
