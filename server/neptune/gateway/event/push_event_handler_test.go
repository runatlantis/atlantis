package event_test

import (
	"context"
	"testing"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/lyft/feature"
	"github.com/runatlantis/atlantis/server/neptune/gateway/event"
	"github.com/runatlantis/atlantis/server/neptune/sync"
	"github.com/runatlantis/atlantis/server/neptune/workflows"
	"github.com/runatlantis/atlantis/server/vcs"
	"github.com/stretchr/testify/assert"
)

const testRoot = "testroot"

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
		allocator := &testAllocator{
			expectedAllocation: true,
			expectedFeatureID:  feature.PlatformMode,
			expectedFeatureCtx: feature.FeatureContext{
				RepoName: repoFullName,
			},
			t: t,
		}

		handler := event.PushHandler{
			Allocator:    allocator,
			Scheduler:    &sync.SynchronousScheduler{Logger: logger},
			Logger:       logger,
			RootDeployer: &mockRootDeployer{},
		}

		err := handler.Handle(context.Background(), e)
		assert.NoError(t, err)
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

		allocator := &testAllocator{
			expectedAllocation: true,
			expectedFeatureID:  feature.PlatformMode,
			expectedFeatureCtx: feature.FeatureContext{
				RepoName: repoFullName,
			},
			t: t,
		}

		handler := event.PushHandler{
			Allocator:    allocator,
			Scheduler:    &sync.SynchronousScheduler{Logger: logger},
			Logger:       logger,
			RootDeployer: &mockRootDeployer{},
		}

		err := handler.Handle(context.Background(), e)
		assert.NoError(t, err)
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
		allocator := &testAllocator{
			expectedAllocation: true,
			expectedFeatureID:  feature.PlatformMode,
			expectedFeatureCtx: feature.FeatureContext{
				RepoName: repoFullName,
			},
			t: t,
		}

		handler := event.PushHandler{
			Allocator:    allocator,
			Scheduler:    &sync.SynchronousScheduler{Logger: logger},
			Logger:       logger,
			RootDeployer: &mockRootDeployer{},
		}

		err := handler.Handle(context.Background(), e)
		assert.NoError(t, err)
	})

}

func TestHandlePushEvent(t *testing.T) {
	logger := logging.NewNoopCtxLogger(t)

	repoFullName := "nish/repo"
	repoOwner := "nish"
	repoName := "repo"
	repoURL := "www.nish.com"
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
		allocator := &testAllocator{
			expectedAllocation: false,
			expectedFeatureID:  feature.PlatformMode,
			expectedFeatureCtx: feature.FeatureContext{
				RepoName: repoFullName,
			},
			t: t,
		}

		handler := event.PushHandler{
			Allocator:    allocator,
			Scheduler:    &sync.SynchronousScheduler{Logger: logger},
			Logger:       logger,
			RootDeployer: &mockRootDeployer{},
		}

		err := handler.Handle(context.Background(), e)
		assert.NoError(t, err)
	})

	t.Run("allocation error", func(t *testing.T) {
		allocator := &testAllocator{
			expectedError:     assert.AnError,
			expectedFeatureID: feature.PlatformMode,
			expectedFeatureCtx: feature.FeatureContext{
				RepoName: repoFullName,
			},
			t: t,
		}

		handler := event.PushHandler{
			Allocator:    allocator,
			Scheduler:    &sync.SynchronousScheduler{Logger: logger},
			Logger:       logger,
			RootDeployer: &mockRootDeployer{},
		}

		err := handler.Handle(context.Background(), e)
		assert.NoError(t, err)
	})

	t.Run("success", func(t *testing.T) {
		allocator := &testAllocator{
			expectedAllocation: true,
			expectedFeatureID:  feature.PlatformMode,
			expectedFeatureCtx: feature.FeatureContext{
				RepoName: repoFullName,
			},
			t: t,
		}
		ctx := context.Background()
		handler := event.PushHandler{
			Allocator:    allocator,
			Scheduler:    &sync.SynchronousScheduler{Logger: logger},
			Logger:       logger,
			RootDeployer: &mockRootDeployer{},
		}

		err := handler.Handle(ctx, e)
		assert.NoError(t, err)
	})

	t.Run("root deployer error", func(t *testing.T) {
		allocator := &testAllocator{
			expectedAllocation: true,
			expectedFeatureID:  feature.PlatformMode,
			expectedFeatureCtx: feature.FeatureContext{
				RepoName: repoFullName,
			},
			t: t,
		}

		ctx := context.Background()
		handler := event.PushHandler{
			Allocator:    allocator,
			Scheduler:    &sync.SynchronousScheduler{Logger: logger},
			Logger:       logger,
			RootDeployer: &mockRootDeployer{error: assert.AnError},
		}

		err := handler.Handle(ctx, e)
		assert.Error(t, err)
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

type mockRootDeployer struct {
	isCalled bool
	error    error
}

func (m *mockRootDeployer) Deploy(_ context.Context, _ event.RootDeployOptions) error {
	m.isCalled = true
	return m.error
}
