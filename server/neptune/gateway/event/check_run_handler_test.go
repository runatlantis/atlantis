package event_test

import (
	"context"
	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/neptune/sync"
	"testing"

	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/neptune/gateway/event"
	"github.com/stretchr/testify/assert"
)

func TestCheckRunHandler(t *testing.T) {
	t.Run("unrelated check run", func(t *testing.T) {
		signaler := &testSignaler{}
		logger := logging.NewNoopCtxLogger(t)
		subject := event.CheckRunHandler{
			Logger:            logging.NewNoopCtxLogger(t),
			RootConfigBuilder: &mockRootConfigBuilder{},
			Scheduler:         &sync.SynchronousScheduler{Logger: logger},
			DeploySignaler:    &mockDeploySignaler{},
		}
		e := event.CheckRun{
			Name: "something",
		}
		err := subject.Handle(context.Background(), e)
		assert.NoError(t, err)
		assert.False(t, signaler.called)
	})

	t.Run("unsupported action", func(t *testing.T) {
		signaler := &mockDeploySignaler{}
		logger := logging.NewNoopCtxLogger(t)
		subject := event.CheckRunHandler{
			Logger:            logging.NewNoopCtxLogger(t),
			RootConfigBuilder: &mockRootConfigBuilder{},
			Scheduler:         &sync.SynchronousScheduler{Logger: logger},
			DeploySignaler:    signaler,
		}
		e := event.CheckRun{
			Action: event.WrappedCheckRunAction("test"),
			Name:   "atlantis/deploy: testroot",
		}
		err := subject.Handle(context.Background(), e)
		assert.NoError(t, err)
		assert.False(t, signaler.called)
	})

	t.Run("invalid rerequested branch", func(t *testing.T) {
		version, err := version.NewVersion("1.0.3")
		assert.NoError(t, err)
		signaler := &mockDeploySignaler{}
		logger := logging.NewNoopCtxLogger(t)
		rootCfg := valid.MergedProjectCfg{
			Name: testRoot,
			DeploymentWorkflow: valid.Workflow{
				Plan:  valid.DefaultPlanStage,
				Apply: valid.DefaultApplyStage,
			},
			TerraformVersion: version,
			WorkflowMode:     valid.DefaultWorkflowMode,
		}
		rootCfgs := []*valid.MergedProjectCfg{
			&rootCfg,
		}
		rootConfigBuilder := &mockRootConfigBuilder{
			rootConfigs: rootCfgs,
		}
		subject := event.CheckRunHandler{
			Logger:            logging.NewNoopCtxLogger(t),
			RootConfigBuilder: rootConfigBuilder,
			Scheduler:         &sync.SynchronousScheduler{Logger: logger},
			DeploySignaler:    signaler,
		}
		e := event.CheckRun{
			Action: event.WrappedCheckRunAction("test"),
			Name:   "atlantis/deploy: testroot",
			Repo:   models.Repo{DefaultBranch: "main"},
			Branch: "something",
		}
		err = subject.Handle(context.Background(), e)
		assert.NoError(t, err)
		assert.False(t, signaler.called)
	})

	t.Run("invalid rerequested branch", func(t *testing.T) {
		signaler := &mockDeploySignaler{}
		logger := logging.NewNoopCtxLogger(t)
		subject := event.CheckRunHandler{
			Logger:            logging.NewNoopCtxLogger(t),
			RootConfigBuilder: &mockRootConfigBuilder{},
			Scheduler:         &sync.SynchronousScheduler{Logger: logger},
			DeploySignaler:    signaler,
		}
		e := event.CheckRun{
			Action: event.WrappedCheckRunAction("test"),
			Name:   "atlantis/deploy: testroot",
			Repo:   models.Repo{DefaultBranch: "main"},
			Branch: "something",
		}
		err := subject.Handle(context.Background(), e)
		assert.NoError(t, err)
		assert.False(t, signaler.called)
	})

	t.Run("wrong requested actions object", func(t *testing.T) {
		signaler := &mockDeploySignaler{}
		logger := logging.NewNoopCtxLogger(t)
		subject := event.CheckRunHandler{
			Logger:            logging.NewNoopCtxLogger(t),
			RootConfigBuilder: &mockRootConfigBuilder{},
			Scheduler:         &sync.SynchronousScheduler{Logger: logger},
			DeploySignaler:    signaler,
		}
		e := event.CheckRun{
			Action: event.WrappedCheckRunAction("requested_action"),
			Name:   "atlantis/deploy: testroot",
		}
		err := subject.Handle(context.Background(), e)
		assert.Error(t, err)
		assert.False(t, signaler.called)
	})

	t.Run("unsupported action id", func(t *testing.T) {
		signaler := &mockDeploySignaler{}
		logger := logging.NewNoopCtxLogger(t)
		subject := event.CheckRunHandler{
			Logger:            logging.NewNoopCtxLogger(t),
			RootConfigBuilder: &mockRootConfigBuilder{},
			Scheduler:         &sync.SynchronousScheduler{Logger: logger},
			DeploySignaler:    signaler,
		}
		e := event.CheckRun{
			Action: event.RequestedActionChecksAction{
				Identifier: "some random thing",
			},
			Name: "atlantis/deploy: testroot",
		}
		err := subject.Handle(context.Background(), e)
		assert.Error(t, err)
		assert.False(t, signaler.called)
	})

	t.Run("plan signal success", func(t *testing.T) {
		signaler := &mockDeploySignaler{}
		user := models.User{Username: "nish"}
		workflowID := "wfid"
		logger := logging.NewNoopCtxLogger(t)
		subject := event.CheckRunHandler{
			Logger:            logging.NewNoopCtxLogger(t),
			RootConfigBuilder: &mockRootConfigBuilder{},
			Scheduler:         &sync.SynchronousScheduler{Logger: logger},
			DeploySignaler:    signaler,
		}
		e := event.CheckRun{
			Action: event.RequestedActionChecksAction{
				Identifier: "Approve",
			},
			ExternalID: workflowID,
			User:       user,
			Name:       "atlantis/deploy: testroot",
		}
		err := subject.Handle(context.Background(), e)
		assert.NoError(t, err)
		assert.True(t, signaler.called)
	})

	t.Run("unlock signal success", func(t *testing.T) {
		user := models.User{Username: "nish"}
		workflowID := "testrepo||testroot"
		signaler := &mockDeploySignaler{}
		logger := logging.NewNoopCtxLogger(t)
		subject := event.CheckRunHandler{
			Logger:            logging.NewNoopCtxLogger(t),
			RootConfigBuilder: &mockRootConfigBuilder{},
			Scheduler:         &sync.SynchronousScheduler{Logger: logger},
			DeploySignaler:    signaler,
		}
		e := event.CheckRun{
			Action: event.RequestedActionChecksAction{
				Identifier: "Unlock",
			},
			ExternalID: workflowID,
			User:       user,
			Repo:       models.Repo{FullName: "testrepo"},
			Name:       "atlantis/deploy: testroot",
		}
		err := subject.Handle(context.Background(), e)
		assert.NoError(t, err)
		assert.True(t, signaler.called)
	})

	t.Run("invalid root name", func(t *testing.T) {
		user := models.User{Username: "nish"}
		workflowID := "testrepo||testroot"
		subject := event.CheckRunHandler{
			Logger: logging.NewNoopCtxLogger(t),
		}
		e := event.CheckRun{
			Action: event.RequestedActionChecksAction{
				Identifier: "Unlock",
			},
			ExternalID: workflowID,
			User:       user,
			Repo:       models.Repo{FullName: "testrepo"},
			Name:       "atlantis/invalid: testroot",
		}
		err := subject.Handle(context.Background(), e)
		assert.ErrorContains(t, err, "unable to determine root name")
	})

	t.Run("signal error", func(t *testing.T) {
		signaler := &mockDeploySignaler{error: assert.AnError}
		user := models.User{Username: "nish"}
		workflowID := "wfid"
		logger := logging.NewNoopCtxLogger(t)
		subject := event.CheckRunHandler{
			Logger:            logging.NewNoopCtxLogger(t),
			RootConfigBuilder: &mockRootConfigBuilder{},
			Scheduler:         &sync.SynchronousScheduler{Logger: logger},
			DeploySignaler:    signaler,
		}
		e := event.CheckRun{
			Action: event.RequestedActionChecksAction{
				Identifier: "Approve",
			},
			ExternalID: workflowID,
			User:       user,
			Name:       "atlantis/deploy: testroot",
		}
		err := subject.Handle(context.Background(), e)
		assert.Error(t, err)
		assert.True(t, signaler.called)
	})

}
