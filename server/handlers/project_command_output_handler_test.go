package handlers_test

import (
	"errors"
	"sync"
	"testing"

	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/handlers"
	"github.com/runatlantis/atlantis/server/handlers/mocks"
	"github.com/runatlantis/atlantis/server/handlers/mocks/matchers"
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

func createProjectCommandOutputHandler(t *testing.T) handlers.ProjectCommandOutputHandler {
	logger := logging.NewNoopLogger(t)
	prjCmdOutputChan := make(chan *handlers.ProjectCmdOutputLine)
	projectStatusUpdater := mocks.NewMockProjectStatusUpdater()
	projectJobURLGenerator := mocks.NewMockProjectJobURLGenerator()
	prjCmdOutputHandler := handlers.NewAsyncProjectCommandOutputHandler(
		prjCmdOutputChan,
		projectStatusUpdater,
		projectJobURLGenerator,
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

		projectOutputHandler.Send(ctx, Msg)
		wg.Wait()
		close(ch)

		Equals(t, expectedMsg, Msg)
	})

	t.Run("copies buffer to new channels", func(t *testing.T) {
		var wg sync.WaitGroup

		projectOutputHandler := createProjectCommandOutputHandler(t)

		// send first message to populated the buffer
		projectOutputHandler.Send(ctx, Msg)

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

	t.Run("update project status with project jobs url", func(t *testing.T) {
		RegisterMockTestingT(t)
		logger := logging.NewNoopLogger(t)
		prjCmdOutputChan := make(chan *handlers.ProjectCmdOutputLine)
		projectStatusUpdater := mocks.NewMockProjectStatusUpdater()
		projectJobURLGenerator := mocks.NewMockProjectJobURLGenerator()
		prjCmdOutputHandler := handlers.NewAsyncProjectCommandOutputHandler(
			prjCmdOutputChan,
			projectStatusUpdater,
			projectJobURLGenerator,
			logger,
		)

		When(projectJobURLGenerator.GenerateProjectJobURL(matchers.EqModelsProjectCommandContext(ctx))).ThenReturn("url-to-project-jobs", nil)
		err := prjCmdOutputHandler.SetJobURLWithStatus(ctx, models.PlanCommand, models.PendingCommitStatus)
		Ok(t, err)

		projectStatusUpdater.VerifyWasCalledOnce().UpdateProject(ctx, models.PlanCommand, models.PendingCommitStatus, "url-to-project-jobs")
	})

	t.Run("update project status with project jobs url error", func(t *testing.T) {
		RegisterMockTestingT(t)
		logger := logging.NewNoopLogger(t)
		prjCmdOutputChan := make(chan *handlers.ProjectCmdOutputLine)
		projectStatusUpdater := mocks.NewMockProjectStatusUpdater()
		projectJobURLGenerator := mocks.NewMockProjectJobURLGenerator()
		prjCmdOutputHandler := handlers.NewAsyncProjectCommandOutputHandler(
			prjCmdOutputChan,
			projectStatusUpdater,
			projectJobURLGenerator,
			logger,
		)

		When(projectJobURLGenerator.GenerateProjectJobURL(matchers.EqModelsProjectCommandContext(ctx))).ThenReturn("url-to-project-jobs", errors.New("some error"))
		err := prjCmdOutputHandler.SetJobURLWithStatus(ctx, models.PlanCommand, models.PendingCommitStatus)
		assert.Error(t, err)
	})

	// Close all jobs for a PR when clean up
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

		projectOutputHandler.Send(ctx, Msg)
		projectOutputHandler.Send(ctx, "Complete")

		pullContext := handlers.PullContext{
			PullNum:     ctx.Pull.Num,
			Repo:        ctx.BaseRepo.Name,
			ProjectName: ctx.ProjectName,
			Workspace:   ctx.Workspace,
		}
		projectOutputHandler.CleanUp(pullContext)

		// Check all the resources are cleaned up.
		dfProjectOutputHandler, ok := projectOutputHandler.(*handlers.AsyncProjectCommandOutputHandler)
		assert.True(t, ok)

		assert.Empty(t, dfProjectOutputHandler.GetProjectOutputBuffer(ctx.JobID))
		assert.Empty(t, dfProjectOutputHandler.GetReceiverBufferForPull(ctx.JobID))
		assert.Empty(t, dfProjectOutputHandler.GetJobIdMapForPullContext(pullContext))
	})
}
