package deploy_test

import (
	"go.temporal.io/sdk/client"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/revision/queue"
	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

const testSignalID = "test-signal"

type queueWorker struct {
	state queue.WorkerState
	ctx   workflow.Context
}

func (w *queueWorker) GetState() queue.WorkerState {
	return w.state
}

func (w *queueWorker) Work(ctx workflow.Context) {
	w.state = queue.WorkingWorkerState
	// sleep and then flip to waiting
	_ = workflow.Sleep(ctx, 60*time.Second)

	w.state = queue.WaitingWorkerState

	// do this so we can check for cancellation status
	w.ctx = ctx
}

type receiver struct {
	receiveCalled bool
	ctx           workflow.Context
}

func (n *receiver) Receive(c workflow.ReceiveChannel, more bool) {

	var s string
	c.Receive(n.ctx, &s)
	n.receiveCalled = true
}

type response struct {
	WorkerCtxCancelled bool
	ReceiverCalled     bool
}

type request struct {
	WorkerState queue.WorkerState
}

func testWorkflow(ctx workflow.Context, r request) (response, error) {
	receiver := &receiver{ctx: ctx}

	worker := &queueWorker{state: r.WorkerState}

	runner := &deploy.Runner{
		QueueWorker:              worker,
		RevisionReceiver:         receiver,
		NewRevisionSignalChannel: workflow.GetSignalChannel(ctx, testSignalID),
		MetricsHandler:           client.MetricsNopHandler,
	}

	err := runner.Run(ctx)

	return response{
		WorkerCtxCancelled: worker.ctx.Err() == workflow.ErrCanceled,
		ReceiverCalled:     receiver.receiveCalled,
	}, err
}

func TestRunner(t *testing.T) {
	t.Run("cancels waiting worker", func(t *testing.T) {
		ts := testsuite.WorkflowTestSuite{}
		env := ts.NewTestWorkflowEnvironment()

		// should timeout since we're not sending any signal
		env.ExecuteWorkflow(testWorkflow, request{})

		var resp response
		err := env.GetWorkflowResult(&resp)
		assert.NoError(t, err)
		assert.Equal(t, response{WorkerCtxCancelled: true, ReceiverCalled: false}, resp)
	})

	t.Run("receives signal and then times out", func(t *testing.T) {
		ts := testsuite.WorkflowTestSuite{}
		env := ts.NewTestWorkflowEnvironment()

		env.RegisterDelayedCallback(func() {
			env.SignalWorkflow(testSignalID, "")
		}, 7*time.Second)

		// should timeout after sending the first signal
		env.ExecuteWorkflow(testWorkflow, request{})

		var resp response
		err := env.GetWorkflowResult(&resp)
		assert.NoError(t, err)
		assert.Equal(t, response{WorkerCtxCancelled: true, ReceiverCalled: true}, resp)
	})
}
