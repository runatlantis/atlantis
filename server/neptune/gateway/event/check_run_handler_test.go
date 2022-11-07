package event_test

import (
	"context"
	"github.com/runatlantis/atlantis/server/events/models"
	"testing"

	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/neptune/gateway/event"
	"github.com/runatlantis/atlantis/server/neptune/workflows"
	"github.com/stretchr/testify/assert"
)

func TestCheckRunHandler(t *testing.T) {
	t.Run("unrelated check run", func(t *testing.T) {
		signaler := &testSignaler{}
		subject := event.CheckRunHandler{
			Logger:         logging.NewNoopCtxLogger(t),
			TemporalClient: signaler,
		}
		e := event.CheckRun{
			Name: "something",
		}
		err := subject.Handle(context.Background(), e)
		assert.NoError(t, err)
		assert.False(t, signaler.called)
	})

	t.Run("unsupported action", func(t *testing.T) {
		signaler := &testSignaler{}
		subject := event.CheckRunHandler{
			Logger:         logging.NewNoopCtxLogger(t),
			TemporalClient: signaler,
		}
		e := event.CheckRun{
			Action: event.WrappedCheckRunAction("test"),
			Name:   "atlantis/deploy: testroot",
		}
		err := subject.Handle(context.Background(), e)
		assert.NoError(t, err)
		assert.False(t, signaler.called)
	})

	t.Run("wrong requested actions object", func(t *testing.T) {
		signaler := &testSignaler{}
		subject := event.CheckRunHandler{
			Logger:         logging.NewNoopCtxLogger(t),
			TemporalClient: signaler,
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
		signaler := &testSignaler{}
		subject := event.CheckRunHandler{
			Logger:         logging.NewNoopCtxLogger(t),
			TemporalClient: signaler,
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
		user := "nish"
		workflowID := "wfid"
		signaler := &testSignaler{
			t:                  t,
			expectedWorkflowID: workflowID,
			expectedSignalName: workflows.TerraformPlanReviewSignalName,
			expectedSignalArg: workflows.TerraformPlanReviewSignalRequest{
				Status: workflows.ApprovedPlanReviewStatus,
				User:   user,
			},
		}
		subject := event.CheckRunHandler{
			Logger:         logging.NewNoopCtxLogger(t),
			TemporalClient: signaler,
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
		user := "nish"
		workflowID := "testrepo||testroot"
		signaler := &testSignaler{
			t:                  t,
			expectedWorkflowID: workflowID,
			expectedSignalName: workflows.DeployUnlockSignalName,
			expectedSignalArg: workflows.DeployUnlockSignalRequest{
				User: user,
			},
		}
		subject := event.CheckRunHandler{
			Logger:         logging.NewNoopCtxLogger(t),
			TemporalClient: signaler,
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
		user := "nish"
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
		user := "nish"
		workflowID := "wfid"
		signaler := &testSignaler{
			t:                  t,
			expectedWorkflowID: workflowID,
			expectedSignalName: workflows.TerraformPlanReviewSignalName,
			expectedSignalArg: workflows.TerraformPlanReviewSignalRequest{
				Status: workflows.ApprovedPlanReviewStatus,
				User:   user,
			},
			expectedErr: assert.AnError,
		}
		subject := event.CheckRunHandler{
			Logger:         logging.NewNoopCtxLogger(t),
			TemporalClient: signaler,
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
