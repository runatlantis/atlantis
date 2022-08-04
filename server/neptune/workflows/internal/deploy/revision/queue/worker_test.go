package queue_test

import (
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/revision/queue"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/github"
	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

type request struct {
	Queue []string
}

type response struct {
	QueueIsEmpty bool
	EndState queue.WorkerState
}

type queueAndState struct {
	QueueIsEmpty bool
	State        queue.WorkerState
}

func testWorkflow(ctx workflow.Context, r request) (response, error) {
	q := queue.NewQueue()

	for _, s := range r.Queue {
		q.Push(wrap(s))
	}

	var a *activities.Deploy
	worker := queue.Worker{
		Queue:      q,
		Activities: a,
		Repo: github.Repo{
			Name:     "hello",
			Owner:    "nish",
			URL:      "git@github.com/nish/hello.git",
		},
	}

	err := workflow.SetQueryHandler(ctx, "queue", func() (queueAndState, error) {

		return queueAndState{
			QueueIsEmpty: q.IsEmpty(),
			State:        worker.GetState(),
		}, nil
	})
	if err != nil {
		return response{}, err
	}

	worker.Work(ctx)

	return response{
		QueueIsEmpty: q.IsEmpty(),
		EndState: worker.GetState(),
	}, nil
}

func TestWorker(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	// we set this callback so we can query the state of the queue
	// after all processing has complete to determine whether we should 
	// shutdown the worker
	env.RegisterDelayedCallback(func() {
		encoded, err := env.QueryWorkflow("queue")

		assert.NoError(t, err)

		var q queueAndState
		err = encoded.Get(&q)

		assert.NoError(t, err)

		assert.True(t, q.QueueIsEmpty)
		assert.Equal(t, queue.WaitingWorkerState, q.State)

		env.CancelWorkflow()

	}, 10*time.Second)

	env.ExecuteWorkflow(testWorkflow, request{
		Queue: []string{"1", "2"},
	})

	var resp response
	err := env.GetWorkflowResult(&resp)
	assert.NoError(t, err)

	assert.Equal(t, queue.CompleteWorkerState, resp.EndState)
	assert.True(t, resp.QueueIsEmpty)

}
