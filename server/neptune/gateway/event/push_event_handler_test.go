package event_test

import (
	"context"
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/lyft/feature"
	"github.com/runatlantis/atlantis/server/neptune/gateway/event"
	"github.com/runatlantis/atlantis/server/neptune/gateway/sync"
	"github.com/runatlantis/atlantis/server/neptune/workflows"
	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/client"
)

type testRun struct{}

func (r testRun) GetID() string {
	return "123"
}

func (r testRun) GetRunID() string {
	return "456"
}

func (r testRun) Get(ctx context.Context, valuePtr interface{}) error {
	return nil
}

func (r testRun) GetWithOptions(ctx context.Context, valuePtr interface{}, options client.WorkflowRunGetOptions) error {
	return nil
}

type testSignaler struct {
	t                    *testing.T
	expectedWorkflowID   string
	expectedSignalName   string
	expectedSignalArg    interface{}
	expectedOptions      client.StartWorkflowOptions
	expectedWorkflow     interface{}
	expectedWorkflowArgs interface{}
	expectedErr          error

	called bool
}

func (s *testSignaler) SignalWithStartWorkflow(ctx context.Context, workflowID string, signalName string, signalArg interface{},
	options client.StartWorkflowOptions, workflow interface{}, workflowArgs ...interface{}) (client.WorkflowRun, error) {

	s.called = true

	assert.Equal(s.t, s.expectedWorkflowID, workflowID)
	assert.Equal(s.t, s.expectedSignalName, signalName)
	assert.Equal(s.t, s.expectedSignalArg, signalArg)
	assert.Equal(s.t, s.expectedOptions, options)
	assert.IsType(s.t, s.expectedWorkflow, workflow)
	assert.Equal(s.t, []interface{}{s.expectedWorkflowArgs}, workflowArgs)

	return testRun{}, s.expectedErr
}

type testAllocator struct {
	t                  *testing.T
	expectedFeatureID  feature.Name
	expectedFeatureCtx feature.FeatureContext
	expectedAllocation bool
	expectedError      error
}

func (a *testAllocator) ShouldAllocate(featureID feature.Name, featureCtx feature.FeatureContext) (bool, error) {
	assert.Equal(a.t, a.expectedFeatureID, featureID)
	assert.Equal(a.t, a.expectedFeatureCtx, featureCtx)

	return a.expectedAllocation, a.expectedError
}

func TestHandlePushEvent(t *testing.T) {
	logger := logging.NewNoopCtxLogger(t)

	repoFullName := "nish/repo"
	repoOwner := "nish"
	repoName := "repo"
	repoURL := "www.nish.com"
	sha := "12345"

	e := event.Push{
		Repo: models.Repo{
			FullName: repoFullName,
			Name:     repoName,
			Owner:    repoOwner,
			CloneURL: repoURL,
		},
		Sha: sha,
	}

	t.Run("allocation result false", func(t *testing.T) {
		testSignaler := &testSignaler{t: t}
		allocator := &testAllocator{
			expectedAllocation: false,
			expectedFeatureID:  feature.PlatformMode,
			expectedFeatureCtx: feature.FeatureContext{
				RepoName: repoFullName,
			},
			t: t,
		}

		handler := event.PushHandler{
			Scheduler:      &sync.SynchronousScheduler{Logger: logger},
			TemporalClient: testSignaler,
			Allocator:      allocator,
			Logger:         logger,
		}

		err := handler.Handle(context.Background(), e)
		assert.NoError(t, err)

		assert.False(t, testSignaler.called)
	})

	t.Run("allocation error", func(t *testing.T) {
		testSignaler := &testSignaler{t: t}
		allocator := &testAllocator{
			expectedError:     assert.AnError,
			expectedFeatureID: feature.PlatformMode,
			expectedFeatureCtx: feature.FeatureContext{
				RepoName: repoFullName,
			},
			t: t,
		}

		handler := event.PushHandler{
			Scheduler:      &sync.SynchronousScheduler{Logger: logger},
			TemporalClient: testSignaler,
			Allocator:      allocator,
			Logger:         logger,
		}

		err := handler.Handle(context.Background(), e)
		assert.NoError(t, err)

		assert.False(t, testSignaler.called)
	})

	t.Run("signal success", func(t *testing.T) {
		testSignaler := &testSignaler{
			t:                  t,
			expectedWorkflowID: repoFullName,
			expectedSignalName: workflows.DeployNewRevisionSignalID,
			expectedSignalArg: workflows.DeployNewRevisionSignalRequest{
				Revision: sha,
			},
			expectedWorkflow: workflows.Deploy,
			expectedOptions: client.StartWorkflowOptions{
				TaskQueue: workflows.DeployTaskQueue,
			},
			expectedWorkflowArgs: workflows.DeployRequest{
				Repository: workflows.Repo{
					FullName: repoFullName,
					Name:     repoName,
					Owner:    repoOwner,
					URL:      repoURL,
				},
			},
		}
		allocator := &testAllocator{
			expectedAllocation: true,
			expectedFeatureID:  feature.PlatformMode,
			expectedFeatureCtx: feature.FeatureContext{
				RepoName: repoFullName,
			},
			t: t,
		}

		handler := event.PushHandler{
			Scheduler:      &sync.SynchronousScheduler{Logger: logger},
			TemporalClient: testSignaler,
			Allocator:      allocator,
			Logger:         logger,
		}

		err := handler.Handle(context.Background(), e)
		assert.NoError(t, err)

		assert.True(t, testSignaler.called)
	})

	t.Run("signal error", func(t *testing.T) {
		testSignaler := &testSignaler{
			t:                  t,
			expectedWorkflowID: repoFullName,
			expectedSignalName: workflows.DeployNewRevisionSignalID,
			expectedSignalArg: workflows.DeployNewRevisionSignalRequest{
				Revision: sha,
			},
			expectedWorkflow: workflows.Deploy,
			expectedOptions: client.StartWorkflowOptions{
				TaskQueue: workflows.DeployTaskQueue,
			},
			expectedWorkflowArgs: workflows.DeployRequest{
				Repository: workflows.Repo{
					FullName: repoFullName,
					Name:     repoName,
					Owner:    repoOwner,
					URL:      repoURL,
				},
			},
			expectedErr: assert.AnError,
		}
		allocator := &testAllocator{
			expectedAllocation: true,
			expectedFeatureID:  feature.PlatformMode,
			expectedFeatureCtx: feature.FeatureContext{
				RepoName: repoFullName,
			},
			t: t,
		}

		handler := event.PushHandler{
			Scheduler:      &sync.SynchronousScheduler{Logger: logger},
			TemporalClient: testSignaler,
			Allocator:      allocator,
			Logger:         logger,
		}

		err := handler.Handle(context.Background(), e)
		assert.Error(t, err)

		assert.True(t, testSignaler.called)
	})
}
