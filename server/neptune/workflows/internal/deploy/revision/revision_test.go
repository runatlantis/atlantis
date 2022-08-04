package revision_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/revision"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/revision/queue"
	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

type testQueue struct {
	Queue []string
}

func (q *testQueue) Push(msg queue.Message) {
	q.Queue = append(q.Queue, msg.Revision)
}

type response struct {
	Queue   []string
	Timeout bool
}

func testWorkflow(ctx workflow.Context) (response, error) {
	var timeout bool
	queue := &testQueue{}

	receiver := revision.NewReceiver(ctx, queue)
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
	env.ExecuteWorkflow(testWorkflow)
	assert.True(t, env.IsWorkflowCompleted())

	var resp response
	err := env.GetWorkflowResult(&resp)
	assert.NoError(t, err)

	assert.Equal(t, []string{rev}, resp.Queue)
	assert.False(t, resp.Timeout)
}
