package queue_test

import (
	"container/list"
	"errors"
	"fmt"
	"testing"
	"time"

	"go.temporal.io/sdk/client"

	"github.com/google/uuid"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/deployment"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/revision/queue"
	internalTerraform "github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/terraform"
	terraformWorkflow "github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

type testQueue struct {
	Queue *list.List
	Lock  queue.LockState
}

func (q *testQueue) IsEmpty() bool {
	return q.Queue.Len() == 0
}
func (q *testQueue) CanPop() bool {
	return !q.IsEmpty()
}
func (q *testQueue) Pop() (internalTerraform.DeploymentInfo, error) {
	if q.IsEmpty() {
		return internalTerraform.DeploymentInfo{}, fmt.Errorf("calling pop on empty queue")
	}

	result := q.Queue.Remove(q.Queue.Front())
	return result.(internalTerraform.DeploymentInfo), nil
}

func (q *testQueue) Push(msg internalTerraform.DeploymentInfo) {
	q.Queue.PushBack(msg)
}

func (q *testQueue) SetLockForMergedItems(ctx workflow.Context, state queue.LockState) {
	q.Lock = state
}

type workerRequest struct {
	Queue                         []internalTerraform.DeploymentInfo
	ExpectedValidationErrors      []*queue.ValidationError
	ExpectedPlanRejectionErrros   []*internalTerraform.PlanRejectionError
	ExpectedTerraformClientErrors []*activities.TerraformClientError
	InitialLockStatus             queue.LockStatus
}

type workerResponse struct {
	QueueIsEmpty bool
	EndState     queue.WorkerState
	CapturedArgs []*deployment.Info
}

type queueAndState struct {
	QueueIsEmpty      bool
	State             queue.WorkerState
	Lock              queue.LockState
	CurrentDeployment queue.CurrentDeployment
	LatestDeployment  *deployment.Info
}

type testDeployer struct {
	expectedRevisions             []*deployment.Info
	expectedValidationErrors      []*queue.ValidationError
	expectedPlanRejectionErrros   []*internalTerraform.PlanRejectionError
	ExpectedTerraformClientErrors []*activities.TerraformClientError

	capturedLatestDeployments []*deployment.Info

	// keeps track of where we are in expected revisions, panics
	// if we go to far
	count int
}

func (d *testDeployer) Deploy(ctx workflow.Context, requestedDeployment internalTerraform.DeploymentInfo, latestDeployment *deployment.Info) (*deployment.Info, error) {
	d.capturedLatestDeployments = append(d.capturedLatestDeployments, latestDeployment)
	info := d.expectedRevisions[d.count]

	// Checking if the error is nil instead of returning the nil error directly since Temporal wraps the error somehow
	var err error
	if d.expectedValidationErrors[d.count] != nil {
		err = d.expectedValidationErrors[d.count]
	} else if d.expectedPlanRejectionErrros[d.count] != nil {
		err = d.expectedPlanRejectionErrros[d.count]
	} else if d.ExpectedTerraformClientErrors[d.count] != nil {
		err = d.ExpectedTerraformClientErrors[d.count]
	}
	d.count++

	return info, err
}

func testWorkerWorkflow(ctx workflow.Context, r workerRequest) (workerResponse, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToCloseTimeout: 5 * time.Second,
	})

	q := &testQueue{
		Queue: list.New(),
	}

	q.SetLockForMergedItems(ctx, queue.LockState{Status: r.InitialLockStatus})

	var infos []*deployment.Info
	for _, s := range r.Queue {
		infos = append(infos, &deployment.Info{
			Revision: s.Revision,
		})
		q.Push(s)
	}

	deployer := &testDeployer{
		expectedRevisions:             infos,
		expectedValidationErrors:      r.ExpectedValidationErrors,
		expectedPlanRejectionErrros:   r.ExpectedPlanRejectionErrros,
		ExpectedTerraformClientErrors: r.ExpectedTerraformClientErrors,
	}

	worker := queue.Worker{
		Queue:          q,
		Deployer:       deployer,
		MetricsHandler: client.MetricsNopHandler,
	}

	err := workflow.SetQueryHandler(ctx, "queue", func() (queueAndState, error) {

		return queueAndState{
			QueueIsEmpty:      q.IsEmpty(),
			State:             worker.GetState(),
			Lock:              q.Lock,
			CurrentDeployment: worker.GetCurrentDeploymentState(),
			LatestDeployment:  worker.GetLatestDeployment(),
		}, nil
	})
	if err != nil {
		return workerResponse{}, err
	}

	worker.Work(ctx)

	return workerResponse{
		QueueIsEmpty: q.IsEmpty(),
		EndState:     worker.GetState(),
		CapturedArgs: deployer.capturedLatestDeployments,
	}, nil
}

func TestWorker_ReceivesUnlockSignal(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(queue.UnlockSignalName, queue.UnlockSignalRequest{User: "username"})
	}, 5*time.Second)

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
		assert.Equal(t, queue.LockState{
			Status: queue.UnlockedStatus,
		}, q.Lock)
		assert.Equal(t, queue.WaitingWorkerState, q.State)

		env.CancelWorkflow()

	}, 10*time.Second)

	env.ExecuteWorkflow(testWorkerWorkflow, workerRequest{
		// start locked and ensure we can unlock it.
		InitialLockStatus: queue.LockedStatus,
	})

	env.AssertExpectations(t)

	var resp workerResponse
	err := env.GetWorkflowResult(&resp)
	assert.NoError(t, err)

	// deploy should never be called
	assert.Len(t, resp.CapturedArgs, 0)

	assert.Equal(t, queue.CompleteWorkerState, resp.EndState)
	assert.True(t, resp.QueueIsEmpty)
}

func TestWorker_DeploysItems(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	repo := github.Repo{
		Owner: "owner",
		Name:  "test",
	}

	deploymentInfoList := []internalTerraform.DeploymentInfo{
		{
			ID:         uuid.UUID{},
			Revision:   "1",
			CheckRunID: 1234,
			Root: terraform.Root{
				Name: "root_1",
			},
			Repo: repo,
		},
		{
			ID:         uuid.UUID{},
			Revision:   "2",
			CheckRunID: 5678,
			Root: terraform.Root{
				Name: "root_2",
			},
			Repo: repo,
		},
	}

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
		assert.Equal(t, q.CurrentDeployment, queue.CurrentDeployment{
			Deployment: deploymentInfoList[len(deploymentInfoList)-1],
			Status:     queue.CompleteStatus,
		})

		env.CancelWorkflow()

	}, 10*time.Second)

	env.ExecuteWorkflow(testWorkerWorkflow, workerRequest{
		Queue: deploymentInfoList,
		ExpectedValidationErrors: []*queue.ValidationError{
			nil, nil,
		},
		ExpectedPlanRejectionErrros: []*internalTerraform.PlanRejectionError{
			nil, nil,
		},
		ExpectedTerraformClientErrors: []*activities.TerraformClientError{
			nil, nil,
		},
	})

	env.AssertExpectations(t)

	var resp workerResponse
	err := env.GetWorkflowResult(&resp)
	assert.NoError(t, err)

	assert.Equal(t, []*deployment.Info{
		nil, {
			Revision: "1",
		},
	}, resp.CapturedArgs)
	assert.Equal(t, queue.CompleteWorkerState, resp.EndState)
	assert.True(t, resp.QueueIsEmpty)
}

func TestWorker_DeploysItems_ValidationError_SkipLatestDeploymentUpdate(t *testing.T) {
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

		// 1st deploy is the latest deployment since 2nd deploy resulted in a validation error
		assert.Equal(t, deployment.Info{
			Revision: "1",
		}, *q.LatestDeployment)

		env.CancelWorkflow()

	}, 10*time.Second)

	repo := github.Repo{
		Owner: "owner",
		Name:  "test",
	}

	deploymentInfoList := []internalTerraform.DeploymentInfo{
		{
			ID:         uuid.UUID{},
			Revision:   "1",
			CheckRunID: 1234,
			Root: terraform.Root{
				Name: "root_1",
			},
			Repo: repo,
		},
		{
			ID:         uuid.UUID{},
			Revision:   "2",
			CheckRunID: 5678,
			Root: terraform.Root{
				Name: "root_2",
			},
			Repo: repo,
		},
	}

	expectedValidationErrors := []*queue.ValidationError{
		nil, queue.NewValidationError("err"),
	}

	// 2nd deploy failed due to validation error so the 1st deploy should be the expected latest deployment
	expectedLatestDeployment := deployment.Info{
		Revision: "1",
	}

	env.ExecuteWorkflow(testWorkerWorkflow, workerRequest{
		Queue:                    deploymentInfoList,
		ExpectedValidationErrors: expectedValidationErrors,
		ExpectedPlanRejectionErrros: []*internalTerraform.PlanRejectionError{
			nil, nil,
		},
		ExpectedTerraformClientErrors: []*activities.TerraformClientError{
			nil, nil,
		},
	})

	env.AssertExpectations(t)

	var resp workerResponse
	err := env.GetWorkflowResult(&resp)
	assert.NoError(t, err)

	assert.Equal(t, queue.CompleteWorkerState, resp.EndState)

	assert.Equal(t, []*deployment.Info{
		nil, &expectedLatestDeployment,
	}, resp.CapturedArgs)
	assert.True(t, resp.QueueIsEmpty)
}

func TestNewWorker(t *testing.T) {
	emptyWorkflow := func(ctx workflow.Context, request terraformWorkflow.Request) error {
		return nil
	}

	type res struct {
		Lock queue.LockState
	}

	testWorkflow := func(ctx workflow.Context) (res, error) {
		ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			ScheduleToCloseTimeout: 5 * time.Second,
		})
		q := queue.NewQueue(noopCallback, client.MetricsNopHandler)
		_, err := queue.NewWorker(ctx, q, client.MetricsNopHandler, &testDeployActivity{}, emptyWorkflow, "nish/repo", "root")
		return res{
			Lock: q.GetLockState(),
		}, err
	}

	fetchDeployRequest := activities.FetchLatestDeploymentRequest{
		FullRepositoryName: "nish/repo",
		RootName:           "root",
	}

	t.Run("last deploy was manual", func(t *testing.T) {
		ts := testsuite.WorkflowTestSuite{}
		env := ts.NewTestWorkflowEnvironment()

		a := &testDeployActivity{}
		env.RegisterActivity(a)

		env.OnActivity(a.FetchLatestDeployment, mock.Anything, fetchDeployRequest).Return(activities.FetchLatestDeploymentResponse{
			DeploymentInfo: &deployment.Info{
				Root: deployment.Root{
					Name:    "root",
					Trigger: "manual",
				},
				Revision: "1234",
			},
		}, nil)

		env.ExecuteWorkflow(testWorkflow)

		env.AssertExpectations(t)

		var r res
		err := env.GetWorkflowResult(&r)
		assert.NoError(t, err)

		assert.Equal(t, queue.LockState{
			Revision: "1234",
			Status:   queue.LockedStatus,
		}, r.Lock)

	})

	t.Run("last deploy was merged", func(t *testing.T) {
		ts := testsuite.WorkflowTestSuite{}
		env := ts.NewTestWorkflowEnvironment()

		a := &testDeployActivity{}
		env.RegisterActivity(a)

		env.OnActivity(a.FetchLatestDeployment, mock.Anything, fetchDeployRequest).Return(activities.FetchLatestDeploymentResponse{
			DeploymentInfo: &deployment.Info{
				Root: deployment.Root{
					Name:    "root",
					Trigger: "merged",
				},
				Revision: "1234",
			},
		}, nil)

		env.ExecuteWorkflow(testWorkflow)

		env.AssertExpectations(t)

		var r res
		err := env.GetWorkflowResult(&r)
		assert.NoError(t, err)

		assert.Equal(t, queue.LockState{}, r.Lock)
	})

	t.Run("first deploy", func(t *testing.T) {
		ts := testsuite.WorkflowTestSuite{}
		env := ts.NewTestWorkflowEnvironment()

		a := &testDeployActivity{}
		env.RegisterActivity(a)

		env.OnActivity(a.FetchLatestDeployment, mock.Anything, fetchDeployRequest).Return(activities.FetchLatestDeploymentResponse{
			DeploymentInfo: nil,
		}, nil)

		env.ExecuteWorkflow(testWorkflow)

		env.AssertExpectations(t)

		var r res
		err := env.GetWorkflowResult(&r)
		assert.NoError(t, err)

		assert.Equal(t, queue.LockState{}, r.Lock)
	})
}

func TestWorker_DeploysItems_PlanRejectionError_SkipLatestDeploymentUpdate(t *testing.T) {
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

		// 1st deploy is the latest deployment since 2nd deploy resulted in a validation error
		assert.Equal(t, deployment.Info{
			Revision: "1",
		}, *q.LatestDeployment)

		env.CancelWorkflow()

	}, 10*time.Second)

	repo := github.Repo{
		Owner: "owner",
		Name:  "test",
	}

	deploymentInfoList := []internalTerraform.DeploymentInfo{
		{
			ID:         uuid.UUID{},
			Revision:   "1",
			CheckRunID: 1234,
			Root: terraform.Root{
				Name: "root_1",
			},
			Repo: repo,
		},
		{
			ID:         uuid.UUID{},
			Revision:   "2",
			CheckRunID: 5678,
			Root: terraform.Root{
				Name: "root_2",
			},
			Repo: repo,
		},
	}

	expectedPlanRejectionErrors := []*internalTerraform.PlanRejectionError{
		nil, internalTerraform.NewPlanRejectionError("plan rejected"),
	}

	// 2nd deploy failed due to plan rejection error so the 1st deploy should be the expected latest deployment
	expectedLatestDeployment := deployment.Info{
		Revision: "1",
	}

	env.ExecuteWorkflow(testWorkerWorkflow, workerRequest{
		Queue:                       deploymentInfoList,
		ExpectedPlanRejectionErrros: expectedPlanRejectionErrors,
		ExpectedValidationErrors: []*queue.ValidationError{
			nil, nil,
		},
		ExpectedTerraformClientErrors: []*activities.TerraformClientError{
			nil, nil,
		},
	})

	env.AssertExpectations(t)

	var resp workerResponse
	err := env.GetWorkflowResult(&resp)
	assert.NoError(t, err)

	assert.Equal(t, queue.CompleteWorkerState, resp.EndState)

	assert.Equal(t, []*deployment.Info{
		nil, &expectedLatestDeployment,
	}, resp.CapturedArgs)
	assert.True(t, resp.QueueIsEmpty)
}

func TestWorker_DeploysItems_TerraformClientError_UpdateLatestDeployment(t *testing.T) {
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

		//  2nd deploy since it resulted in a TFClient error
		assert.Equal(t, deployment.Info{
			Revision: "2",
		}, *q.LatestDeployment)

		env.CancelWorkflow()

	}, 10*time.Second)

	repo := github.Repo{
		Owner: "owner",
		Name:  "test",
	}

	deploymentInfoList := []internalTerraform.DeploymentInfo{
		{
			ID:         uuid.UUID{},
			Revision:   "1",
			CheckRunID: 1234,
			Root: terraform.Root{
				Name: "root_1",
			},
			Repo: repo,
		},
		{
			ID:         uuid.UUID{},
			Revision:   "2",
			CheckRunID: 5678,
			Root: terraform.Root{
				Name: "root_2",
			},
			Repo: repo,
		},
	}

	firstDeployment := deployment.Info{
		Revision: "1",
	}

	env.ExecuteWorkflow(testWorkerWorkflow, workerRequest{
		Queue: deploymentInfoList,
		ExpectedPlanRejectionErrros: []*internalTerraform.PlanRejectionError{
			nil, nil,
		},
		ExpectedValidationErrors: []*queue.ValidationError{
			nil, nil,
		},
		ExpectedTerraformClientErrors: []*activities.TerraformClientError{
			nil, activities.NewTerraformClientError(errors.New("tf client error")),
		},
	})

	env.AssertExpectations(t)

	var resp workerResponse
	err := env.GetWorkflowResult(&resp)
	assert.NoError(t, err)

	assert.Equal(t, queue.CompleteWorkerState, resp.EndState)

	assert.Equal(t, []*deployment.Info{
		nil, &firstDeployment,
	}, resp.CapturedArgs)
	assert.True(t, resp.QueueIsEmpty)
}
