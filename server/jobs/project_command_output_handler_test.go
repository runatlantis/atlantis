package jobs_test

import (
	"sync"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/jobs"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
	"github.com/stretchr/testify/assert"
)

func createTestProjectCmdContext(t *testing.T) models.ProjectCommandContext {
	logger := logging.NewNoopLogger(t)
	return models.ProjectCommandContext{
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
			HeadBranch: "master",
			BaseBranch: "master",
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

		ch := make(chan string)

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

		ch := make(chan string)

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
		assert.Equal(t, len(expectedMsgs), len(receivedMsgs))
		for i := range expectedMsgs {
			assert.Equal(t, expectedMsgs[i], receivedMsgs[i])
		}
	})

	t.Run("clean up all jobs when PR is closed", func(t *testing.T) {
		var wg sync.WaitGroup
		projectOutputHandler := createProjectCommandOutputHandler(t)

		ch := make(chan string)

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
			PullNum:     ctx.Pull.Num,
			Repo:        ctx.BaseRepo.Name,
			ProjectName: ctx.ProjectName,
			Workspace:   ctx.Workspace,
		}
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

		ch := make(chan string)

		// register channel and backfill from buffer
		// Note: We call this synchronously because otherwise
		// there could be a race where we are unable to register the channel
		// before sending messages due to the way we lock our buffer memory cache
		projectOutputHandler.Register(ctx.JobID, ch)

		// read from channel
		go func() {
			for range ch {
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
			for range ch {
			}
		}()

		projectOutputHandler.Send(ctx, Msg, false)
		projectOutputHandler.Send(ctx, "", true)

		// Wait for the handler to process the message
		time.Sleep(10 * time.Millisecond)

		ch2 := make(chan string)
		opComplete := make(chan bool)

		// buffer channel will be closed immediately after logs are streamed
		go func() {
			for range ch2 {
			}
			opComplete <- true
		}()

		projectOutputHandler.Register(ctx.JobID, ch2)

		assert.True(t, <-opComplete)
	})
}
