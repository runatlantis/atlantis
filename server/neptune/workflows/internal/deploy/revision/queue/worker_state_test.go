package queue_test

import (
	"container/list"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/deployment"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"

	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/revision/queue"
	internalTerraform "github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/metrics"
)

type testBlockingDeployer struct{}

func (d *testBlockingDeployer) Deploy(ctx workflow.Context, requestedDeployment internalTerraform.DeploymentInfo, latestDeployment *deployment.Info, _ metrics.Scope) (*deployment.Info, error) {
	ch := workflow.GetSignalChannel(ctx, "unblock-deploy")
	selector := workflow.NewSelector(ctx)

	var v string
	selector.AddReceive(ch, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &v)
	})

	for {
		selector.Select(ctx)

		if v != "" {
			return requestedDeployment.BuildPersistableInfo(), nil
		}
	}
}

func testStateWorkflow(ctx workflow.Context, r workerRequest) (workerResponse, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToCloseTimeout: 30 * time.Second,
	})

	q := &testQueue{
		Queue: list.New(),
	}

	q.SetLockForMergedItems(ctx, queue.LockState{Status: r.InitialLockStatus})

	workflow.Go(ctx, func(ctx workflow.Context) {
		ch := workflow.GetSignalChannel(ctx, "add-to-queue")

		for {
			var info internalTerraform.DeploymentInfo
			ch.Receive(ctx, &info)
			q.Push(info)
		}
	})

	worker := queue.Worker{
		Queue:    q,
		Deployer: &testBlockingDeployer{},
	}

	err := workflow.SetQueryHandler(ctx, "queue", func() (queueAndState, error) {
		return queueAndState{
			QueueIsEmpty:      q.IsEmpty(),
			State:             worker.GetState(),
			Lock:              q.Lock,
			CurrentDeployment: worker.GetCurrentDeploymentState(),
		}, nil
	})
	if err != nil {
		return workerResponse{}, err
	}

	worker.Work(ctx)

	return workerResponse{
		QueueIsEmpty: q.IsEmpty(),
		EndState:     worker.GetState(),
	}, nil
}

// Here we want to make sure that the worker state is reflective
// of the actual queue + deploy
// 1. start with an unlocked queue with no items
// 2. query the workflow to ensure the state is waiting
// 3. add something to queue via a signal
// 4. query the workflow again to ensure the state is working
// 5. unblock the deployment via a signal
// 6. Cancel worker
// 7. ensure ending workflow state is complete

func TestWorker_StartsWithEmptyQueue(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	env.RegisterDelayedCallback(func() {
		encoded, err := env.QueryWorkflow("queue")

		assert.NoError(t, err)

		var q queueAndState
		err = encoded.Get(&q)

		assert.NoError(t, err)

		assert.True(t, q.QueueIsEmpty)
		assert.Equal(t, queue.LockState{
			Status: queue.UnlockedStatus,
		}, q.Lock)
		assert.Equal(t, queue.WaitingWorkerState, q.State)

	}, 2*time.Second)

	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("add-to-queue", internalTerraform.DeploymentInfo{
			Revision: "1234",
			ID:       uuid.New(),
			Root: terraform.Root{
				Name:    "some-name",
				Trigger: terraform.MergeTrigger,
			},
			Repo: github.Repo{
				Name:  "some-name",
				Owner: "some-owner",
			},
		})
	}, 4*time.Second)

	env.RegisterDelayedCallback(func() {
		encoded, err := env.QueryWorkflow("queue")

		assert.NoError(t, err)

		var q queueAndState
		err = encoded.Get(&q)

		assert.NoError(t, err)

		assert.True(t, q.QueueIsEmpty)
		assert.Equal(t, queue.LockState{
			Status: queue.UnlockedStatus,
		}, q.Lock)
		assert.Equal(t, queue.WorkingWorkerState, q.State)

	}, 6*time.Second)

	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("unblock-deploy", "hi")
	}, 8*time.Second)

	env.RegisterDelayedCallback(func() {
		env.CancelWorkflow()
	}, 20*time.Second)

	env.ExecuteWorkflow(testStateWorkflow, workerRequest{})

	var resp workerResponse
	err := env.GetWorkflowResult(&resp)
	assert.NoError(t, err)

	assert.Equal(t, queue.CompleteWorkerState, resp.EndState)
	assert.True(t, resp.QueueIsEmpty)
}
