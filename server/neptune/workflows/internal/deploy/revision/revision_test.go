package revision_test

import (
	"context"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/revision"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/revision/queue"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

type testQueue struct {
	Queue []queue.Message
}

func (q *testQueue) Push(msg queue.Message) {
	q.Queue = append(q.Queue, msg)
}

type response struct {
	Queue   []queue.Message
	Timeout bool
}

type testActivities struct{}

func (a *testActivities) CreateCheckRun(ctx context.Context, request activities.CreateCheckRunRequest) (activities.CreateCheckRunResponse, error) {
	return activities.CreateCheckRunResponse{}, nil
}

func testWorkflow(ctx workflow.Context) (response, error) {

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToCloseTimeout: 5 * time.Second,
	})
	var timeout bool
	queue := &testQueue{}

	var a *testActivities

	receiver := revision.NewReceiver(ctx, queue, github.Repo{Name: "nish"}, a)
	selector := workflow.NewSelector(ctx)

	selector.AddReceive(workflow.GetSignalChannel(ctx, "test-signal"), receiver.Receive)

	for {
		selector.Select(ctx)

		if !selector.HasPending() {
			break
		}
	}

	return response{
		Queue:   queue.Queue,
		Timeout: timeout,
	}, nil
}

func TestEnqueue(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	rev := "1234"

	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("test-signal", revision.NewRevisionRequest{
			Revision: rev,
		})
	}, 0)

	a := &testActivities{}

	env.RegisterActivity(a)

	env.OnActivity(a.CreateCheckRun, mock.Anything, activities.CreateCheckRunRequest{
		Title: "atlantis/deploy",
		Sha:   rev,
		Repo:  github.Repo{Name: "nish"},
	}).Return(activities.CreateCheckRunResponse{ID: 1}, nil)

	env.ExecuteWorkflow(testWorkflow)
	env.AssertExpectations(t)
	assert.True(t, env.IsWorkflowCompleted())

	var resp response
	err := env.GetWorkflowResult(&resp)
	assert.NoError(t, err)

	assert.Equal(t, []queue.Message{
		{
			Revision:   rev,
			CheckRunID: 1,
		},
	}, resp.Queue)
	assert.False(t, resp.Timeout)
}
