package jobs_test

import (
	"regexp"
	"sync"
	"testing"

	"github.com/runatlantis/atlantis/server/events/terraform/filter"
	"github.com/stretchr/testify/assert"

	. "github.com/petergtz/pegomock"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/jobs"
	"github.com/runatlantis/atlantis/server/jobs/mocks"
	"github.com/runatlantis/atlantis/server/jobs/mocks/matchers"
	"github.com/runatlantis/atlantis/server/logging"

	. "github.com/runatlantis/atlantis/testing"
)

func createTestProjectCmdContext(t *testing.T) command.ProjectContext {
	logger := logging.NewNoopCtxLogger(t)
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
	logger := logging.NewNoopCtxLogger(t)
	prjCmdOutputChan := make(chan *jobs.ProjectCmdOutputLine)
	jobStore := mocks.NewMockJobStore()
	prjCmdOutputHandler := jobs.NewAsyncProjectCommandOutputHandler(
		prjCmdOutputChan,
		logger,
		jobStore,
		filter.LogFilter{
			Regexes: []*regexp.Regexp{regexp.MustCompile("InvalidMessage")},
		},
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

		When(jobStore.Get(matchers.AnyContextContext(), AnyString())).ThenReturn(&jobs.Job{}, nil)
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
		projectOutputHandler.Register(ctx.RequestCtx, ctx.JobID, ch)

		wg.Add(1)
		projectOutputHandler.Send(ctx, Msg)
		wg.Wait()
		close(ch)

		Equals(t, expectedMsg, Msg)
	})

	t.Run("strip message from main channel", func(t *testing.T) {
		var wg sync.WaitGroup
		var expectedMsg string
		projectOutputHandler, jobStore := createProjectCommandOutputHandler(t)
		strippedMessage := "InvalidMessage test"

		When(jobStore.Get(matchers.AnyContextContext(), AnyString())).ThenReturn(&jobs.Job{}, nil)
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
		projectOutputHandler.Register(ctx.RequestCtx, ctx.JobID, ch)

		wg.Add(1)
		// even if stripped message is sent first, registered channel will never receive it, making expectedMsg == Msg
		projectOutputHandler.Send(ctx, strippedMessage)
		projectOutputHandler.Send(ctx, Msg)
		wg.Wait()
		close(ch)

		Equals(t, expectedMsg, Msg)
	})

	t.Run("copies buffer to new channels", func(t *testing.T) {
		var wg sync.WaitGroup
		var receivedMsg string

		projectOutputHandler, jobStore := createProjectCommandOutputHandler(t)

		// Mocking the job store acts like populating the buffer
		When(jobStore.Get(matchers.AnyContextContext(), AnyString())).ThenReturn(&jobs.Job{
			Output: []string{Msg},
			Status: jobs.Processing,
		}, nil)

		ch := make(chan string)
		go func() {
			for msg := range ch {
				receivedMsg = msg
				wg.Done()
			}
		}()

		wg.Add(1)

		// Register the channel and wait for msg in the buffer to be read
		projectOutputHandler.Register(ctx.RequestCtx, ctx.JobID, ch)
		wg.Wait()

		close(ch)

		// Assert received msg is copied from the buffer
		assert.Equal(t, receivedMsg, Msg)
	})

	t.Run("clean up all jobs when PR is closed", func(t *testing.T) {
		projectOutputHandler, jobStore := createProjectCommandOutputHandler(t)
		When(jobStore.Get(matchers.AnyContextContext(), AnyString())).ThenReturn(&jobs.Job{}, nil)

		ch := make(chan string)

		// read from channel
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			<-ch
			wg.Done()
		}()

		// register channel and backfill from buffer
		// Note: We call this synchronously because otherwise
		// there could be a race where we are unable to register the channel
		// before sending messages due to the way we lock our buffer memory cache
		projectOutputHandler.Register(ctx.RequestCtx, ctx.JobID, ch)
		projectOutputHandler.Send(ctx, Msg)

		wg.Wait()

		pullContext := jobs.PullInfo{
			PullNum:     ctx.Pull.Num,
			Repo:        ctx.BaseRepo.Name,
			ProjectName: ctx.ProjectName,
			Workspace:   ctx.Workspace,
		}

		// Cleanup is called when a PR is closed
		projectOutputHandler.CleanUp(pullContext)

		// Check all the resources are cleaned up.
		dfProjectOutputHandler, ok := projectOutputHandler.(*jobs.AsyncProjectCommandOutputHandler)
		assert.True(t, ok)

		job, err := dfProjectOutputHandler.JobStore.Get(ctx.RequestCtx, ctx.JobID)
		Ok(t, err)

		assert.Empty(t, job.Output)
		assert.Empty(t, dfProjectOutputHandler.GetReceiverBufferForPull(ctx.JobID))
		assert.Empty(t, dfProjectOutputHandler.GetJobIDMapForPull(pullContext))
	})

	t.Run("close conn buffer after streaming logs for completed operation", func(t *testing.T) {
		projectOutputHandler, jobStore := createProjectCommandOutputHandler(t)
		job := jobs.Job{
			Output: []string{"a", "b"},
			Status: jobs.Complete,
		}
		When(jobStore.Get(matchers.AnyContextContext(), AnyString())).ThenReturn(&job, nil)

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
		projectOutputHandler.Register(ctx.RequestCtx, ctx.JobID, ch)

		assert.True(t, <-opComplete)
	})
}
