package handlers_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/handlers"
	"github.com/runatlantis/atlantis/server/handlers/mocks"
	"github.com/runatlantis/atlantis/server/handlers/mocks/matchers"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/stretchr/testify/assert"

	. "github.com/petergtz/pegomock"
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
		},
		User: models.User{
			Username: "test-user",
		},
		Log:         logger,
		Workspace:   "myworkspace",
		RepoRelDir:  "test-dir",
		ProjectName: "test-project",
	}
}

func createProjectCommandOutputHandler(t *testing.T) handlers.ProjectCommandOutputHandler {
	logger := logging.NewNoopLogger(t)
	prjCmdOutputChan := make(chan *models.ProjectCmdOutputLine)
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

	t.Run("Should Receive Message Sent in the ProjectCmdOutput channel", func(t *testing.T) {
		var wg sync.WaitGroup
		var expectedMsg string
		projectOutputHandler := createProjectCommandOutputHandler(t)

		wg.Add(1)
		ch := make(chan string)
		go func() {
			err := projectOutputHandler.Receive(ctx.PullInfo(), ch, func(msg string) error {
				expectedMsg = msg
				wg.Done()
				return nil
			})
			Ok(t, err)
		}()

		projectOutputHandler.Send(ctx, Msg)
		close(ch)

		// Wait for the msg to be read.
		wg.Wait()
		Equals(t, expectedMsg, Msg)
	})

	t.Run("Should Clear ProjectOutputBuffer when new Plan", func(t *testing.T) {
		var wg sync.WaitGroup

		projectOutputHandler := createProjectCommandOutputHandler(t)

		wg.Add(1)
		ch := make(chan string)
		go func() {
			err := projectOutputHandler.Receive(ctx.PullInfo(), ch, func(msg string) error {
				wg.Done()
				return nil
			})
			Ok(t, err)
		}()

		projectOutputHandler.Send(ctx, Msg)

		// Wait for the msg to be read.
		wg.Wait()

		// Send a clear msg
		projectOutputHandler.Clear(ctx)
		close(ch)

		dfProjectOutputHandler, ok := projectOutputHandler.(*handlers.AsyncProjectCommandOutputHandler)
		assert.True(t, ok)

		// Wait for the clear msg to be received by handle()
		time.Sleep(1 * time.Second)
		assert.Empty(t, dfProjectOutputHandler.GetProjectOutputBuffer(ctx.PullInfo()))
	})

	t.Run("Should Cleanup receiverBuffers receiving WS channel closed", func(t *testing.T) {
		var wg sync.WaitGroup

		projectOutputHandler := createProjectCommandOutputHandler(t)

		wg.Add(1)
		ch := make(chan string)
		go func() {
			err := projectOutputHandler.Receive(ctx.PullInfo(), ch, func(msg string) error {
				wg.Done()
				return nil
			})
			Ok(t, err)
		}()

		projectOutputHandler.Send(ctx, Msg)

		// Wait for the msg to be read.
		wg.Wait()

		// Close chan to execute cleanup.
		close(ch)
		time.Sleep(1 * time.Second)

		dfProjectOutputHandler, ok := projectOutputHandler.(*handlers.AsyncProjectCommandOutputHandler)
		assert.True(t, ok)

		x := dfProjectOutputHandler.GetReceiverBufferForPull(ctx.PullInfo())
		assert.Empty(t, x)
	})

	t.Run("Should copy over existing log messages to new WS channels", func(t *testing.T) {
		var wg sync.WaitGroup

		projectOutputHandler := createProjectCommandOutputHandler(t)

		wg.Add(1)
		ch := make(chan string)
		go func() {
			err := projectOutputHandler.Receive(ctx.PullInfo(), ch, func(msg string) error {
				wg.Done()
				return nil
			})
			Ok(t, err)
		}()

		projectOutputHandler.Send(ctx, Msg)

		// Wait for the msg to be read.
		wg.Wait()

		// Close channel to close prev connection.
		// This should close the first go routine with receive call.
		close(ch)

		ch = make(chan string)

		// Expecting two calls to callback.
		wg.Add(2)

		receivedMsgs := []string{}
		go func() {
			err := projectOutputHandler.Receive(ctx.PullInfo(), ch, func(msg string) error {
				receivedMsgs = append(receivedMsgs, msg)
				wg.Done()
				return nil
			})
			Ok(t, err)
		}()

		// Make sure addChan gets the buffer lock and adds ch to the map.
		time.Sleep(1 * time.Second)

		projectOutputHandler.Send(ctx, Msg)

		// Wait for the message to be read.
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
		prjCmdOutputChan := make(chan *models.ProjectCmdOutputLine)
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
		prjCmdOutputChan := make(chan *models.ProjectCmdOutputLine)
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
}
