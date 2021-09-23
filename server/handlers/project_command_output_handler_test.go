package handlers_test

import (
	"sync"
	"testing"
	"time"

	"github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/events/models"
	featuremocks "github.com/runatlantis/atlantis/server/feature/mocks"
	featurematchers "github.com/runatlantis/atlantis/server/feature/mocks/matchers"
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
		wg.Wait()

		close(ch)

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
		wg.Wait()

		// Send a clear msg
		wg.Add(1)
		projectOutputHandler.Clear(ctx)
		wg.Wait()

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

		When(projectJobURLGenerator.GenerateProjectJobURL(matchers.EqModelsProjectCommandContext(ctx))).ThenReturn("url-to-project-jobs")
		err := prjCmdOutputHandler.SetJobURLWithStatus(ctx, models.PlanCommand, models.PendingCommitStatus)
		Ok(t, err)

		projectStatusUpdater.VerifyWasCalledOnce().UpdateProject(ctx, models.PlanCommand, models.PendingCommitStatus, "url-to-project-jobs")
	})
}

func TestFeatureAwareOutputHandler(t *testing.T) {
	ctx := createTestProjectCmdContext(t)
	RegisterMockTestingT(t)
	projectOutputHandler := mocks.NewMockProjectCommandOutputHandler()

	featureAllocator := featuremocks.NewMockAllocator()
	featureAwareOutputHandler := handlers.FeatureAwareOutputHandler{
		FeatureAllocator:            featureAllocator,
		ProjectCommandOutputHandler: projectOutputHandler,
	}

	cases := []struct {
		Description        string
		FeatureFlagEnabled bool
	}{
		{
			Description:        "noop when feature is disabled",
			FeatureFlagEnabled: false,
		},
		{
			Description:        "delegate when feature is enabled",
			FeatureFlagEnabled: true,
		},
	}

	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			var expectedWasCalled func() *EqMatcher

			if c.FeatureFlagEnabled {
				expectedWasCalled = Once
			} else {
				expectedWasCalled = Never
			}
			When(featureAllocator.ShouldAllocate(featurematchers.AnyFeatureName(), pegomock.AnyString())).ThenReturn(c.FeatureFlagEnabled, nil)

			err := featureAwareOutputHandler.SetJobURLWithStatus(ctx, models.PlanCommand, models.PendingCommitStatus)
			Ok(t, err)
			projectOutputHandler.VerifyWasCalled(expectedWasCalled()).SetJobURLWithStatus(matchers.AnyModelsProjectCommandContext(), matchers.AnyModelsCommandName(), matchers.AnyModelsCommitStatus())

			featureAwareOutputHandler.Clear(ctx)
			projectOutputHandler.VerifyWasCalled(expectedWasCalled()).Clear(matchers.AnyModelsProjectCommandContext())

			featureAwareOutputHandler.Send(ctx, "test")
			projectOutputHandler.VerifyWasCalled(expectedWasCalled()).Send(matchers.AnyModelsProjectCommandContext(), pegomock.AnyString())
		})
	}
}
