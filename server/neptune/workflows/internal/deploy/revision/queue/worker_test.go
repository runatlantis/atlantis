package queue_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/revision/queue"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/root"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

type testTerraformWorkflowRunner struct{}

func (r testTerraformWorkflowRunner) Run(ctx workflow.Context, deploymentInfo terraform.DeploymentInfo) error {
	return nil
}

type request struct {
	Queue []terraform.DeploymentInfo
}

type response struct {
	QueueIsEmpty bool
	EndState     queue.WorkerState
}

type queueAndState struct {
	QueueIsEmpty bool
	State        queue.WorkerState
}

func testWorkflow(ctx workflow.Context, r request) (response, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToCloseTimeout: 5 * time.Second,
	})

	q := queue.NewQueue()

	for _, s := range r.Queue {
		q.Push(s)
	}

	var wa *testDeployActivity
	worker := queue.Worker{
		Queue:                   q,
		TerraformWorkflowRunner: &testTerraformWorkflowRunner{},
		DbActivities:            wa,
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
		EndState:     worker.GetState(),
	}, nil
}

func TestWorker(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	da := testDeployActivity{}
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

	deploymentInfoList := []terraform.DeploymentInfo{
		{
			ID:         uuid.UUID{},
			Revision:   "1",
			CheckRunID: 1234,
			Root: root.Root{
				Name: "root_1",
			},
			Repo: github.Repo{
				Owner: "owner",
				Name:  "test",
			},
		},
		{
			ID:         uuid.UUID{},
			Revision:   "2",
			CheckRunID: 5678,
			Root: root.Root{
				Name: "root_2",
			},
			Repo: github.Repo{
				Owner: "owner",
				Name:  "test",
			},
		},
	}

	repo := github.Repo{
		Owner: "owner",
		Name:  "test",
	}

	// Mock FetchLatestDeploymentRequest for both roots
	env.OnActivity(da.FetchLatestDeployment, mock.Anything, activities.FetchLatestDeploymentRequest{
		FullRepositoryName: "owner/test",
		RootName:           deploymentInfoList[0].Root.Name,
	}).Return(activities.FetchLatestDeploymentResponse{}, nil)
	env.OnActivity(da.FetchLatestDeployment, mock.Anything, activities.FetchLatestDeploymentRequest{
		FullRepositoryName: "owner/test",
		RootName:           deploymentInfoList[1].Root.Name,
	}).Return(activities.FetchLatestDeploymentResponse{}, nil)

	// Mock StoreLatestDeploymentRequest for both roots
	env.OnActivity(da.StoreLatestDeployment, mock.Anything, activities.StoreLatestDeploymentRequest{
		DeploymentInfo: root.DeploymentInfo{
			Version:    queue.DeploymentInfoVersion,
			ID:         uuid.UUID{}.String(),
			CheckRunID: deploymentInfoList[0].CheckRunID,
			Revision:   deploymentInfoList[0].Revision,
			Repo:       repo,
			Root: root.Root{
				Name: deploymentInfoList[0].Root.Name,
			},
		},
	}).Return(nil)
	env.OnActivity(da.StoreLatestDeployment, mock.Anything, activities.StoreLatestDeploymentRequest{
		DeploymentInfo: root.DeploymentInfo{
			Version:    queue.DeploymentInfoVersion,
			ID:         uuid.UUID{}.String(),
			CheckRunID: deploymentInfoList[1].CheckRunID,
			Revision:   deploymentInfoList[1].Revision,
			Repo:       repo,
			Root: root.Root{
				Name: deploymentInfoList[1].Root.Name,
			},
		},
	}).Return(nil)

	env.ExecuteWorkflow(testWorkflow, request{
		Queue: []terraform.DeploymentInfo{
			{
				ID:         uuid.UUID{},
				Revision:   "1",
				CheckRunID: 1234,
				Root: root.Root{
					Name: "root_1",
				},
				Repo: github.Repo{
					Owner: "owner",
					Name:  "test",
				},
			},
			{
				ID:         uuid.UUID{},
				Revision:   "2",
				CheckRunID: 5678,
				Root: root.Root{
					Name: "root_2",
				},
				Repo: github.Repo{
					Owner: "owner",
					Name:  "test",
				},
			},
		},
	})

	var resp response
	err := env.GetWorkflowResult(&resp)
	assert.NoError(t, err)

	assert.Equal(t, queue.CompleteWorkerState, resp.EndState)
	assert.True(t, resp.QueueIsEmpty)
}

type testDeployActivity struct{}

func (t *testDeployActivity) FetchLatestDeployment(ctx context.Context, request activities.FetchLatestDeploymentRequest) (activities.FetchLatestDeploymentResponse, error) {
	return activities.FetchLatestDeploymentResponse{}, nil
}

func (t *testDeployActivity) StoreLatestDeployment(ctx context.Context, request activities.StoreLatestDeploymentRequest) error {
	return nil
}
