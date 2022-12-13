package event_test

import (
	"context"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/neptune/gateway/event"
	"github.com/runatlantis/atlantis/server/neptune/sync"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckSuiteHandler(t *testing.T) {
	branch := "branch"
	t.Run("success", func(t *testing.T) {
		logger := logging.NewNoopCtxLogger(t)
		subject := event.CheckSuiteHandler{
			RootDeployer: &mockRootDeployer{},
			Logger:       logging.NewNoopCtxLogger(t),
			Scheduler:    &sync.SynchronousScheduler{Logger: logger},
		}
		e := event.CheckSuite{
			Action: event.WrappedCheckRunAction(event.ReRequestedActionType),
			Branch: branch,
			Repo:   models.Repo{DefaultBranch: branch},
		}
		err := subject.Handle(context.Background(), e)
		assert.NoError(t, err)
	})
	t.Run("unsupported action", func(t *testing.T) {
		logger := logging.NewNoopCtxLogger(t)
		subject := event.CheckSuiteHandler{
			Logger:       logging.NewNoopCtxLogger(t),
			Scheduler:    &sync.SynchronousScheduler{Logger: logger},
			RootDeployer: &mockRootDeployer{},
		}
		e := event.CheckSuite{
			Action: event.WrappedCheckRunAction("something"),
		}
		err := subject.Handle(context.Background(), e)
		assert.NoError(t, err)
	})

	t.Run("invalid branch", func(t *testing.T) {
		logger := logging.NewNoopCtxLogger(t)
		subject := event.CheckSuiteHandler{
			Logger:       logging.NewNoopCtxLogger(t),
			Scheduler:    &sync.SynchronousScheduler{Logger: logger},
			RootDeployer: &mockRootDeployer{},
		}
		e := event.CheckSuite{
			Action: event.WrappedCheckRunAction(event.ReRequestedActionType),
			Branch: "something",
			Repo:   models.Repo{DefaultBranch: branch},
		}
		err := subject.Handle(context.Background(), e)
		assert.NoError(t, err)
	})
	t.Run("failed root deployer", func(t *testing.T) {
		logger := logging.NewNoopCtxLogger(t)
		subject := event.CheckSuiteHandler{
			Logger:       logging.NewNoopCtxLogger(t),
			Scheduler:    &sync.SynchronousScheduler{Logger: logger},
			RootDeployer: &mockRootDeployer{error: assert.AnError},
		}
		e := event.CheckSuite{
			Action: event.WrappedCheckRunAction(event.ReRequestedActionType),
			Branch: branch,
			Repo:   models.Repo{DefaultBranch: branch},
		}
		err := subject.Handle(context.Background(), e)
		assert.Error(t, err)
	})
}
