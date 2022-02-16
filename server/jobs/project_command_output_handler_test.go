package jobs_test

import (
	"sync"
	"testing"
	"time"

	. "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/jobs"
	"github.com/runatlantis/atlantis/server/jobs/mocks"
	"github.com/runatlantis/atlantis/server/logging"

	. "github.com/runatlantis/atlantis/testing"
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

func createProjectCommandOutputHandler(t *testing.T) (jobs.ProjectCommandOutputHandler, *mocks.MockJobStore) {
	logger := logging.NewNoopLogger(t)
	prjCmdOutputChan := make(chan *jobs.ProjectCmdOutputLine)
	jobStore := mocks.NewMockJobStore()
	prjCmdOutputHandler := jobs.NewAsyncProjectCommandOutputHandler(
		prjCmdOutputChan,
		logger,
		jobStore,
	)

	go func() {
		prjCmdOutputHandler.Handle()
	}()

	return prjCmdOutputHandler, jobStore
}

func TestProjectCommandOutputHandler(t *testing.T) {
	Msg := "Test Terraform Output"
	ctx := createTestProjectCmdContext(t)

	t.Run("receive message from main channel", func(t *testing.T) {
		var wg sync.WaitGroup
		var expectedMsg string
		projectOutputHandler, jobStore := createProjectCommandOutputHandler(t)

		When(jobStore.Get(AnyString())).ThenReturn(jobs.Job{}, nil)
		ch := make(chan string)

		// read from channel
		go func() {
			for msg := range ch {
				expectedMsg = msg
				wg.Done()
			}
		}()

		// register channel and backfill from buffer
		// Note: We call this synchronously because otherwise
		// there could be a race where we are unable to register the channel
		// before sending messages due to the way we lock our buffer memory cache
		projectOutputHandler.Register(ctx.JobID, ch)

		wg.Add(1)
		projectOutputHandler.Send(ctx, Msg)
		wg.Wait()
		close(ch)

		Equals(t, expectedMsg, Msg)
	})

	t.Run("copies buffer to new channels", func(t *testing.T) {
		var wg sync.WaitGroup

		projectOutputHandler, jobStore := createProjectCommandOutputHandler(t)
		When(jobStore.Get(AnyString())).ThenReturn(jobs.Job{
			Output: []string{Msg},
			Status: jobs.Processing,
		}, nil)

		// send first message to populate the buffer
		projectOutputHandler.Send(ctx, Msg)
		time.Sleep(10 * time.Millisecond)

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

		projectOutputHandler.Send(ctx, Msg)
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
		projectOutputHandler, jobStore := createProjectCommandOutputHandler(t)
		When(jobStore.Get(AnyString())).ThenReturn(jobs.Job{}, nil)

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

		projectOutputHandler.Send(ctx, Msg)
		projectOutputHandler.Send(ctx, "Complete")

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

		assert.Empty(t, dfProjectOutputHandler.GetJob(ctx.JobID).Output)
		assert.Empty(t, dfProjectOutputHandler.GetReceiverBufferForPull(ctx.JobID))
		assert.Empty(t, dfProjectOutputHandler.GetJobIdMapForPull(pullContext))
	})

	t.Run("close conn buffer after streaming logs for completed operation", func(t *testing.T) {
		projectOutputHandler, jobStore := createProjectCommandOutputHandler(t)
		job := jobs.Job{
			Output: []string{"a", "b"},
			Status: jobs.Complete,
		}
		When(jobStore.Get(AnyString())).ThenReturn(job, nil)

		ch := make(chan string)

		opComplete := make(chan bool)
		// buffer channel will be closed immediately after logs are streamed
		go func() {
			for range ch {
			}
			opComplete <- true
		}()

		// register channel and backfill from buffer
		// Note: We call this synchronously because otherwise
		// there could be a race where we are unable to register the channel
		// before sending messages due to the way we lock our buffer memory cache
		projectOutputHandler.Register(ctx.JobID, ch)

		assert.True(t, <-opComplete)
	})
}
