package event_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/lyft/feature"
	"github.com/runatlantis/atlantis/server/neptune/gateway/event"
	"github.com/runatlantis/atlantis/server/neptune/gateway/sync"
	"github.com/runatlantis/atlantis/server/neptune/workflows"
	"github.com/runatlantis/atlantis/server/vcs"
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

const testRoot = "testroot"

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

func TestHandlePushEvent_FiltersEvents(t *testing.T) {
	logger := logging.NewNoopCtxLogger(t)
	repoFullName := "nish/repo"
	repoOwner := "nish"
	repoName := "repo"
	repoURL := "www.nish.com"
	sha := "12345"

	t.Run("filters non branch types", func(t *testing.T) {
		e := event.Push{
			Repo: models.Repo{
				FullName: repoFullName,
				Name:     repoName,
				Owner:    repoOwner,
				CloneURL: repoURL,
			},
			Sha: sha,
			Ref: vcs.Ref{
				Type: vcs.TagRef,
				Name: "blah",
			},
		}
		testSignaler := &testSignaler{t: t}
		allocator := &testAllocator{
			expectedAllocation: true,
			expectedFeatureID:  feature.PlatformMode,
			expectedFeatureCtx: feature.FeatureContext{
				RepoName: repoFullName,
			},
			t: t,
		}

		handler := event.PushHandler{
			Allocator:         allocator,
			Scheduler:         &sync.SynchronousScheduler{Logger: logger},
			TemporalClient:    testSignaler,
			Logger:            logger,
			RootConfigBuilder: &mockRootConfigBuilder{},
		}

		err := handler.Handle(context.Background(), e)
		assert.NoError(t, err)

		assert.False(t, testSignaler.called)
	})

	t.Run("filters non-default branch types", func(t *testing.T) {
		e := event.Push{
			Repo: models.Repo{
				FullName:      repoFullName,
				Name:          repoName,
				Owner:         repoOwner,
				CloneURL:      repoURL,
				DefaultBranch: "main",
			},
			Sha: sha,
			Ref: vcs.Ref{
				Type: vcs.BranchRef,
				Name: "random",
			},
		}
		testSignaler := &testSignaler{t: t}
		allocator := &testAllocator{
			expectedAllocation: true,
			expectedFeatureID:  feature.PlatformMode,
			expectedFeatureCtx: feature.FeatureContext{
				RepoName: repoFullName,
			},
			t: t,
		}

		handler := event.PushHandler{
			Allocator:         allocator,
			Scheduler:         &sync.SynchronousScheduler{Logger: logger},
			TemporalClient:    testSignaler,
			Logger:            logger,
			RootConfigBuilder: &mockRootConfigBuilder{},
		}

		err := handler.Handle(context.Background(), e)
		assert.NoError(t, err)

		assert.False(t, testSignaler.called)
	})

	t.Run("filters deleted branches", func(t *testing.T) {
		e := event.Push{
			Repo: models.Repo{
				FullName:      repoFullName,
				Name:          repoName,
				Owner:         repoOwner,
				CloneURL:      repoURL,
				DefaultBranch: "main",
			},
			Sha:    sha,
			Action: event.DeletedAction,
			Ref: vcs.Ref{
				Type: vcs.BranchRef,
				Name: "main",
			},
		}
		testSignaler := &testSignaler{t: t}
		allocator := &testAllocator{
			expectedAllocation: true,
			expectedFeatureID:  feature.PlatformMode,
			expectedFeatureCtx: feature.FeatureContext{
				RepoName: repoFullName,
			},
			t: t,
		}

		handler := event.PushHandler{
			Allocator:         allocator,
			Scheduler:         &sync.SynchronousScheduler{Logger: logger},
			TemporalClient:    testSignaler,
			Logger:            logger,
			RootConfigBuilder: &mockRootConfigBuilder{},
		}

		err := handler.Handle(context.Background(), e)
		assert.NoError(t, err)

		assert.False(t, testSignaler.called)
	})

}

func TestHandlePushEvent(t *testing.T) {
	version, err := version.NewVersion("1.0.3")
	assert.NoError(t, err)

	logger := logging.NewNoopCtxLogger(t)

	repoFullName := "nish/repo"
	repoOwner := "nish"
	repoName := "repo"
	repoURL := "www.nish.com"
	repoRefName := "main"
	repoRefType := "branch"
	sha := "12345"
	repo := models.Repo{
		FullName:      repoFullName,
		Name:          repoName,
		Owner:         repoOwner,
		CloneURL:      repoURL,
		DefaultBranch: "main",
	}

	e := event.Push{
		Repo: repo,
		Ref: vcs.Ref{
			Type: vcs.BranchRef,
			Name: "main",
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
			Allocator:         allocator,
			Scheduler:         &sync.SynchronousScheduler{Logger: logger},
			TemporalClient:    testSignaler,
			Logger:            logger,
			RootConfigBuilder: &mockRootConfigBuilder{},
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
			Allocator:         allocator,
			Scheduler:         &sync.SynchronousScheduler{Logger: logger},
			TemporalClient:    testSignaler,
			Logger:            logger,
			RootConfigBuilder: &mockRootConfigBuilder{},
		}

		err := handler.Handle(context.Background(), e)
		assert.NoError(t, err)

		assert.False(t, testSignaler.called)
	})

	t.Run("signal success", func(t *testing.T) {
		testSignaler := &testSignaler{
			t:                  t,
			expectedWorkflowID: fmt.Sprintf("%s||%s", repoFullName, testRoot),
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
					HeadCommit: workflows.HeadCommit{
						Ref: workflows.Ref{
							Name: repoRefName,
							Type: repoRefType,
						},
					},
				},
				Root: workflows.Root{
					Name: testRoot,
					Plan: workflows.Job{
						Steps: convertTestSteps(valid.DefaultPlanStage.Steps),
					},
					Apply: workflows.Job{
						Steps: convertTestSteps(valid.DefaultApplyStage.Steps),
					},
					TfVersion: version.String(),
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
		ctx := context.Background()
		rootCfg := valid.MergedProjectCfg{
			Name: testRoot,
			DeploymentWorkflow: valid.Workflow{
				Plan:  valid.DefaultPlanStage,
				Apply: valid.DefaultApplyStage,
			},
			TerraformVersion: version,
		}
		rootCfgs := []*valid.MergedProjectCfg{
			&rootCfg,
		}
		rootConfigBuilder := &mockRootConfigBuilder{
			rootConfigs: rootCfgs,
		}
		handler := event.PushHandler{
			Allocator:         allocator,
			Scheduler:         &sync.SynchronousScheduler{Logger: logger},
			TemporalClient:    testSignaler,
			Logger:            logger,
			RootConfigBuilder: rootConfigBuilder,
		}

		err := handler.Handle(ctx, e)
		assert.NoError(t, err)

		assert.True(t, testSignaler.called)
	})

	t.Run("signal error", func(t *testing.T) {
		testSignaler := &testSignaler{
			t:                  t,
			expectedWorkflowID: fmt.Sprintf("%s||%s", repoFullName, testRoot),
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
					HeadCommit: workflows.HeadCommit{
						Ref: workflows.Ref{
							Name: repoRefName,
							Type: repoRefType,
						},
					},
				},
				Root: workflows.Root{
					Name: testRoot,
					Plan: workflows.Job{
						Steps: convertTestSteps(valid.DefaultPlanStage.Steps),
					},
					Apply: workflows.Job{
						Steps: convertTestSteps(valid.DefaultApplyStage.Steps),
					},
					TfVersion: version.String(),
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

		ctx := context.Background()
		rootCfg := valid.MergedProjectCfg{
			Name: testRoot,
			DeploymentWorkflow: valid.Workflow{
				Plan:  valid.DefaultPlanStage,
				Apply: valid.DefaultApplyStage,
			},
			TerraformVersion: version,
		}
		rootCfgs := []*valid.MergedProjectCfg{
			&rootCfg,
		}
		rootConfigBuilder := &mockRootConfigBuilder{
			rootConfigs: rootCfgs,
		}
		handler := event.PushHandler{
			Allocator:         allocator,
			Scheduler:         &sync.SynchronousScheduler{Logger: logger},
			TemporalClient:    testSignaler,
			Logger:            logger,
			RootConfigBuilder: rootConfigBuilder,
		}

		err := handler.Handle(ctx, e)
		assert.Error(t, err)

		assert.True(t, testSignaler.called)
	})
}

func convertTestSteps(steps []valid.Step) []workflows.Step {
	var convertedSteps []workflows.Step
	for _, step := range steps {
		convertedSteps = append(convertedSteps, workflows.Step{
			StepName:    step.StepName,
			ExtraArgs:   step.ExtraArgs,
			RunCommand:  step.RunCommand,
			EnvVarName:  step.EnvVarName,
			EnvVarValue: step.EnvVarValue,
		})
	}
	return convertedSteps
}

type mockRootConfigBuilder struct {
	rootConfigs []*valid.MergedProjectCfg
	error       error
}

func (r *mockRootConfigBuilder) Build(_ context.Context, _ event.Push) ([]*valid.MergedProjectCfg, error) {
	return r.rootConfigs, r.error
}
