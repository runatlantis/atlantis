package event_test

import (
	"context"
	"fmt"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/lyft/feature"
	"github.com/runatlantis/atlantis/server/neptune/gateway/event"
	"github.com/runatlantis/atlantis/server/neptune/sync"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

var testRepo = models.Repo{
	FullName: repoFullName,
}
var testPull = models.PullRequest{
	BaseRepo: testRepo,
}

func TestCommentEventWorkerProxy_HandleAllocationError(t *testing.T) {
	logger := logging.NewNoopCtxLogger(t)
	writer := &mockSnsWriter{}
	allocator := &testAllocator{
		t:                 t,
		expectedFeatureID: feature.PlatformMode,
		expectedFeatureCtx: feature.FeatureContext{
			RepoName: repoFullName,
		},
		expectedError: assert.AnError,
	}
	scheduler := &sync.SynchronousScheduler{Logger: logger}
	rootDeployer := &mockRootDeployer{}
	commentCreator := &mockCommentCreator{}
	statusUpdater := &mockStatusUpdater{}
	cfg := valid.NewGlobalCfg("somedir")
	commentEventWorkerProxy := event.NewCommentEventWorkerProxy(logger, writer, allocator, scheduler, rootDeployer, commentCreator, statusUpdater, cfg)
	bufReq := buildRequest(t)
	commentEvent := event.Comment{
		Pull:     testPull,
		BaseRepo: testRepo,
	}
	cmd := &command.Comment{
		Name: command.Plan,
	}
	err := commentEventWorkerProxy.Handle(context.Background(), bufReq, commentEvent, cmd)
	assert.NoError(t, err)
	assert.False(t, rootDeployer.isCalled)
	assert.False(t, commentCreator.isCalled)
	assert.True(t, statusUpdater.isCalled)
	assert.True(t, writer.isCalled)
}

func TestCommentEventWorkerProxy_HandleForceApply(t *testing.T) {
	logger := logging.NewNoopCtxLogger(t)
	writer := &mockSnsWriter{}
	allocator := &testAllocator{
		t:                 t,
		expectedFeatureID: feature.PlatformMode,
		expectedFeatureCtx: feature.FeatureContext{
			RepoName: repoFullName,
		},
	}
	scheduler := &sync.SynchronousScheduler{Logger: logger}
	rootDeployer := &mockRootDeployer{}
	commentCreator := &mockCommentCreator{}
	statusUpdater := &mockStatusUpdater{}
	cfg := valid.NewGlobalCfg("somedir")
	commentEventWorkerProxy := event.NewCommentEventWorkerProxy(logger, writer, allocator, scheduler, rootDeployer, commentCreator, statusUpdater, cfg)
	bufReq := buildRequest(t)
	commentEvent := event.Comment{
		Pull:     testPull,
		BaseRepo: testRepo,
	}
	cmd := &command.Comment{
		Name:       command.Apply,
		ForceApply: true,
	}
	err := commentEventWorkerProxy.Handle(context.Background(), bufReq, commentEvent, cmd)
	assert.NoError(t, err)
	assert.True(t, statusUpdater.isCalled)
	assert.False(t, commentCreator.isCalled)
	assert.False(t, rootDeployer.isCalled)
	assert.True(t, writer.isCalled)
}

func TestCommentEventWorkerProxy_HandlePlatformModeForceApply(t *testing.T) {
	logger := logging.NewNoopCtxLogger(t)
	writer := &mockSnsWriter{}
	allocator := &testAllocator{
		t:                 t,
		expectedFeatureID: feature.PlatformMode,
		expectedFeatureCtx: feature.FeatureContext{
			RepoName: repoFullName,
		},
		expectedAllocation: true,
	}
	scheduler := &sync.SynchronousScheduler{Logger: logger}
	rootDeployer := &mockRootDeployer{}
	commentCreator := &mockCommentCreator{}
	statusUpdater := &mockStatusUpdater{}
	cfg := valid.NewGlobalCfg("somedir")
	commentEventWorkerProxy := event.NewCommentEventWorkerProxy(logger, writer, allocator, scheduler, rootDeployer, commentCreator, statusUpdater, cfg)
	bufReq := buildRequest(t)
	testPull := models.PullRequest{
		BaseRepo: testRepo,
	}
	commentEvent := event.Comment{
		Pull:     testPull,
		BaseRepo: testRepo,
	}
	cmd := &command.Comment{
		Name:       command.Apply,
		ForceApply: true,
	}
	err := commentEventWorkerProxy.Handle(context.Background(), bufReq, commentEvent, cmd)
	assert.NoError(t, err)
	assert.True(t, commentCreator.isCalled)
	assert.True(t, rootDeployer.isCalled)
	assert.False(t, writer.isCalled)
	assert.False(t, statusUpdater.isCalled)
}

func TestCommentEventWorkerProxy_HandlePlanComment(t *testing.T) {
	logger := logging.NewNoopCtxLogger(t)
	writer := &mockSnsWriter{}
	allocator := &testAllocator{
		t:                 t,
		expectedFeatureID: feature.PlatformMode,
		expectedFeatureCtx: feature.FeatureContext{
			RepoName: repoFullName,
		},
	}
	scheduler := &sync.SynchronousScheduler{Logger: logger}
	rootDeployer := &mockRootDeployer{}
	commentCreator := &mockCommentCreator{}
	statusUpdater := &mockStatusUpdater{}
	cfg := valid.NewGlobalCfg("somedir")
	commentEventWorkerProxy := event.NewCommentEventWorkerProxy(logger, writer, allocator, scheduler, rootDeployer, commentCreator, statusUpdater, cfg)
	bufReq := buildRequest(t)
	commentEvent := event.Comment{
		Pull:     testPull,
		BaseRepo: testRepo,
	}
	cmd := &command.Comment{
		Name: command.Plan,
	}
	err := commentEventWorkerProxy.Handle(context.Background(), bufReq, commentEvent, cmd)
	assert.NoError(t, err)
	assert.True(t, statusUpdater.isCalled)
	assert.False(t, commentCreator.isCalled)
	assert.False(t, rootDeployer.isCalled)
	assert.True(t, writer.isCalled)
}

func TestCommentEventWorkerProxy_WriteError(t *testing.T) {
	logger := logging.NewNoopCtxLogger(t)
	writer := &mockSnsWriter{
		err: assert.AnError,
	}
	allocator := &testAllocator{
		t:                 t,
		expectedFeatureID: feature.PlatformMode,
		expectedFeatureCtx: feature.FeatureContext{
			RepoName: repoFullName,
		},
	}
	scheduler := &sync.SynchronousScheduler{Logger: logger}
	rootDeployer := &mockRootDeployer{}
	commentCreator := &mockCommentCreator{}
	statusUpdater := &mockStatusUpdater{}
	cfg := valid.NewGlobalCfg("somedir")
	commentEventWorkerProxy := event.NewCommentEventWorkerProxy(logger, writer, allocator, scheduler, rootDeployer, commentCreator, statusUpdater, cfg)
	bufReq := buildRequest(t)
	commentEvent := event.Comment{
		Pull:     testPull,
		BaseRepo: testRepo,
	}
	cmd := &command.Comment{
		Name: command.Plan,
	}
	err := commentEventWorkerProxy.Handle(context.Background(), bufReq, commentEvent, cmd)
	assert.Error(t, err)
	assert.True(t, statusUpdater.isCalled)
	assert.False(t, commentCreator.isCalled)
	assert.False(t, rootDeployer.isCalled)
	assert.True(t, writer.isCalled)
}

func TestCommentEventWorkerProxy_HandleNoQueuedStatus(t *testing.T) {
	logger := logging.NewNoopCtxLogger(t)
	writer := &mockSnsWriter{}
	scheduler := &sync.SynchronousScheduler{Logger: logger}
	rootDeployer := &mockRootDeployer{}
	commentCreator := &mockCommentCreator{}
	statusUpdater := &mockStatusUpdater{}
	cfg := valid.NewGlobalCfg("somedir")
	// add branch regex
	cfg.Repos = []valid.Repo{
		{
			ID:          "/repo",
			BranchRegex: regexp.MustCompile("regex"),
		},
	}
	bufReq := buildRequest(t)
	allocator := &testAllocator{
		t:                 t,
		expectedFeatureID: feature.PlatformMode,
		expectedFeatureCtx: feature.FeatureContext{
			RepoName: repoFullName,
		},
	}

	forkedPull := models.PullRequest{
		BaseRepo: testRepo,
		HeadRepo: models.Repo{
			Owner: "new-owner",
		},
	}
	closedPull := models.PullRequest{
		BaseRepo: testRepo,
		State:    models.ClosedPullState,
	}
	cases := []struct {
		descriptor string
		allocator  *testAllocator
		command    *command.Comment
		event      event.Comment
	}{
		{
			descriptor: "non-plan/apply comment",
			allocator:  allocator,
			command:    &command.Comment{Name: command.Unlock},
			event: event.Comment{
				Pull:     testPull,
				BaseRepo: testRepo,
			},
		},
		{
			descriptor: "apply comment but platform mode enabled",
			allocator: &testAllocator{
				t:                 t,
				expectedFeatureID: feature.PlatformMode,
				expectedFeatureCtx: feature.FeatureContext{
					RepoName: repoFullName,
				},
				expectedAllocation: true,
			},
			command: &command.Comment{Name: command.Apply},
			event: event.Comment{
				Pull:     testPull,
				BaseRepo: testRepo,
			},
		},
		{
			descriptor: "forked PR",
			allocator:  allocator,
			command:    &command.Comment{Name: command.Plan},
			event: event.Comment{
				Pull:     forkedPull,
				BaseRepo: testRepo,
			},
		},
		{
			descriptor: "closed PR",
			allocator:  allocator,
			command:    &command.Comment{Name: command.Plan},
			event: event.Comment{
				Pull:     closedPull,
				BaseRepo: testRepo,
			},
		},
		{
			descriptor: "invalid base branch",
			allocator:  allocator,
			command:    &command.Comment{Name: command.Plan},
			event: event.Comment{
				Pull:     testPull,
				BaseRepo: testRepo,
			},
		},
	}
	for _, c := range cases {
		t.Run(c.descriptor, func(t *testing.T) {
			commentEventWorkerProxy := event.NewCommentEventWorkerProxy(logger, writer, c.allocator, scheduler, rootDeployer, commentCreator, statusUpdater, cfg)
			err := commentEventWorkerProxy.Handle(context.Background(), bufReq, c.event, c.command)
			assert.NoError(t, err)
			assert.False(t, statusUpdater.isCalled)
			assert.False(t, commentCreator.isCalled)
			assert.False(t, rootDeployer.isCalled)
			assert.True(t, writer.isCalled)
		})
	}
}

type mockCommentCreator struct {
	isCalled bool
	err      error
}

func (c *mockCommentCreator) CreateComment(models.Repo, int, string, string) error {
	c.isCalled = true
	return c.err
}

type mockStatusUpdater struct {
	isCalled bool
	err      error
}

func (s *mockStatusUpdater) UpdateCombined(context.Context, models.Repo, models.PullRequest, models.VCSStatus, fmt.Stringer, string, string) (string, error) {
	s.isCalled = true
	return "", s.err
}
