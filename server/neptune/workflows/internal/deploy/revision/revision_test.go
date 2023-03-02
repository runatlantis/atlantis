package revision_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/notifier"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/request"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/revision"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/revision/queue"
	terraformWorkflow "github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/terraform"
	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

type testCheckRunClient struct {
	expectedRequest notifier.GithubCheckRunRequest
	expectedT       *testing.T
}

func (t *testCheckRunClient) CreateOrUpdate(ctx workflow.Context, deploymentID string, request notifier.GithubCheckRunRequest) (int64, error) {
	assert.Equal(t.expectedT, t.expectedRequest, request)

	return 1, nil
}

type testQueue struct {
	Queue []terraformWorkflow.DeploymentInfo
	Lock  queue.LockState
}

func (q *testQueue) Scan() []terraformWorkflow.DeploymentInfo {
	return q.Queue
}

func (q *testQueue) Push(msg terraformWorkflow.DeploymentInfo) {
	q.Queue = append(q.Queue, msg)
}

func (q *testQueue) GetLockState() queue.LockState {
	return q.Lock
}

func (q *testQueue) SetLockForMergedItems(ctx workflow.Context, state queue.LockState) {
	q.Lock = state
}

type testWorker struct {
	Current queue.CurrentDeployment
}

func (t testWorker) GetCurrentDeploymentState() queue.CurrentDeployment {
	return t.Current
}

type req struct {
	ID              uuid.UUID
	Lock            queue.LockState
	Current         queue.CurrentDeployment
	InitialElements []terraformWorkflow.DeploymentInfo
	ExpectedRequest notifier.GithubCheckRunRequest
	ExpectedT       *testing.T
}

type response struct {
	Queue   []terraformWorkflow.DeploymentInfo
	Lock    queue.LockState
	Timeout bool
}

func testWorkflow(ctx workflow.Context, r req) (response, error) {

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToCloseTimeout: 5 * time.Second,
	})
	var timeout bool
	queue := &testQueue{
		Lock:  r.Lock,
		Queue: r.InitialElements,
	}

	worker := &testWorker{
		Current: r.Current,
	}

	receiver := revision.NewReceiver(ctx, queue, &testCheckRunClient{
		expectedRequest: r.ExpectedRequest,
		expectedT:       r.ExpectedT,
	}, func(ctx workflow.Context) (uuid.UUID, error) {
		return r.ID, nil
	}, worker)
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
		Lock:    queue.Lock,
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
			Root: request.Root{
				Name:    "root",
				Trigger: request.MergeTrigger,
			},
			Repo: request.Repo{Name: "nish"},
		})
	}, 0)

	id := uuid.Must(uuid.NewUUID())

	env.ExecuteWorkflow(testWorkflow, req{
		ID: id,
		ExpectedRequest: notifier.GithubCheckRunRequest{
			Title: "atlantis/deploy: root",
			Sha:   rev,
			Repo:  github.Repo{Name: "nish"},
			State: github.CheckRunQueued,
		},
		ExpectedT: t,
	})
	env.AssertExpectations(t)
	assert.True(t, env.IsWorkflowCompleted())

	var resp response
	err := env.GetWorkflowResult(&resp)
	assert.NoError(t, err)

	assert.Equal(t, []terraformWorkflow.DeploymentInfo{
		{
			Revision:   rev,
			CheckRunID: 1,
			Root:       terraform.Root{Name: "root", Trigger: terraform.MergeTrigger},
			ID:         id,
			Repo:       github.Repo{Name: "nish"},
		},
	}, resp.Queue)
	assert.Equal(t, queue.LockState{
		Status: queue.UnlockedStatus,
	}, resp.Lock)
	assert.False(t, resp.Timeout)
}

func TestEnqueue_ManualTrigger(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	rev := "1234"

	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("test-signal", revision.NewRevisionRequest{
			Revision: rev,
			Root: request.Root{
				Name:    "root",
				Trigger: request.ManualTrigger,
			},
			Repo: request.Repo{Name: "nish"},
		})
	}, 0)

	id := uuid.Must(uuid.NewUUID())

	env.ExecuteWorkflow(testWorkflow, req{
		ID: id,
		ExpectedRequest: notifier.GithubCheckRunRequest{
			Title: "atlantis/deploy: root",
			Sha:   rev,
			Repo:  github.Repo{Name: "nish"},
			State: github.CheckRunQueued,
		},
		ExpectedT: t,
	})
	env.AssertExpectations(t)
	assert.True(t, env.IsWorkflowCompleted())

	var resp response
	err := env.GetWorkflowResult(&resp)
	assert.NoError(t, err)

	assert.Equal(t, []terraformWorkflow.DeploymentInfo{
		{
			Revision:   rev,
			CheckRunID: 1,
			Root:       terraform.Root{Name: "root", Trigger: terraform.ManualTrigger},
			ID:         id,
			Repo:       github.Repo{Name: "nish"},
		},
	}, resp.Queue)
	assert.Equal(t, queue.LockState{
		Status:   queue.LockedStatus,
		Revision: "1234",
	}, resp.Lock)
	assert.False(t, resp.Timeout)
}

func TestEnqueue_ManualTrigger_QueueAlreadyLocked(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	rev := "1234"

	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("test-signal", revision.NewRevisionRequest{
			Revision: rev,
			Root: request.Root{
				Name:    "root",
				Trigger: request.ManualTrigger,
			},
			Repo: request.Repo{Name: "nish"},
		})
	}, 0)

	id := uuid.Must(uuid.NewUUID())

	env.ExecuteWorkflow(testWorkflow, req{
		ID: id,
		Lock: queue.LockState{
			// ensure that the lock gets updated
			Status:   queue.LockedStatus,
			Revision: "123334444555",
		},
		ExpectedRequest: notifier.GithubCheckRunRequest{
			Title: "atlantis/deploy: root",
			Sha:   rev,
			Repo:  github.Repo{Name: "nish"},
			State: github.CheckRunQueued,
		},
		ExpectedT: t,
	})
	env.AssertExpectations(t)
	assert.True(t, env.IsWorkflowCompleted())

	var resp response
	err := env.GetWorkflowResult(&resp)
	assert.NoError(t, err)

	assert.Equal(t, []terraformWorkflow.DeploymentInfo{
		{
			Revision:   rev,
			CheckRunID: 1,
			Root:       terraform.Root{Name: "root", Trigger: terraform.ManualTrigger},
			ID:         id,
			Repo:       github.Repo{Name: "nish"},
		},
	}, resp.Queue)
	assert.Equal(t, queue.LockState{
		Status:   queue.LockedStatus,
		Revision: "1234",
	}, resp.Lock)
	assert.False(t, resp.Timeout)
}

func TestEnqueue_MergeTrigger_QueueAlreadyLocked(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	rev := "1234"

	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("test-signal", revision.NewRevisionRequest{
			Revision: rev,
			Root: request.Root{
				Name:    "root",
				Trigger: request.MergeTrigger,
			},
			Repo: request.Repo{Name: "nish"},
		})
	}, 0)

	id := uuid.Must(uuid.NewUUID())

	env.ExecuteWorkflow(testWorkflow, req{
		ID: id,
		Lock: queue.LockState{
			// ensure that the lock gets updated
			Status:   queue.LockedStatus,
			Revision: "123334444555",
		},
		ExpectedRequest: notifier.GithubCheckRunRequest{
			Title:   "atlantis/deploy: root",
			Sha:     rev,
			Repo:    github.Repo{Name: "nish"},
			Summary: "This deploy is locked from a manual deployment for revision 123334444555.  Unlock to proceed.",
			Actions: []github.CheckRunAction{github.CreateUnlockAction()},
			State:   github.CheckRunActionRequired,
		},
		ExpectedT: t,
	})
	env.AssertExpectations(t)
	assert.True(t, env.IsWorkflowCompleted())

	var resp response
	err := env.GetWorkflowResult(&resp)
	assert.NoError(t, err)

	assert.Equal(t, []terraformWorkflow.DeploymentInfo{
		{
			Revision:   rev,
			CheckRunID: 1,
			Root:       terraform.Root{Name: "root", Trigger: terraform.MergeTrigger},
			ID:         id,
			Repo:       github.Repo{Name: "nish"},
		},
	}, resp.Queue)
	assert.Equal(t, queue.LockState{
		Status:   queue.LockedStatus,
		Revision: "123334444555",
	}, resp.Lock)
	assert.False(t, resp.Timeout)
}

func TestEnqueue_ManualTrigger_RequestAlreadyInQueue(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	rev := "1234"

	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("test-signal", revision.NewRevisionRequest{
			Revision: rev,
			Root: request.Root{
				Name:    "root",
				Trigger: request.ManualTrigger,
			},
			Repo: request.Repo{Name: "nish"},
		})
	}, 0)

	id := uuid.Must(uuid.NewUUID())

	deploymentInfo := terraformWorkflow.DeploymentInfo{
		Revision:   rev,
		CheckRunID: 1,
		Root:       terraform.Root{Name: "root", Trigger: terraform.ManualTrigger},
		ID:         id,
		Repo:       github.Repo{Name: "nish"},
	}
	env.ExecuteWorkflow(testWorkflow, req{
		ID:              id,
		InitialElements: []terraformWorkflow.DeploymentInfo{deploymentInfo},
	})
	env.AssertExpectations(t)
	assert.True(t, env.IsWorkflowCompleted())

	var resp response
	err := env.GetWorkflowResult(&resp)
	assert.NoError(t, err)
	// should not another of the same to the queue
	assert.Equal(t, []terraformWorkflow.DeploymentInfo{deploymentInfo}, resp.Queue)
}

func TestEnqueue_ManualTrigger_RequestAlreadyInProgress(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	rev := "1234"

	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("test-signal", revision.NewRevisionRequest{
			Revision: rev,
			Root: request.Root{
				Name:    "root",
				Trigger: request.ManualTrigger,
			},
			Repo: request.Repo{Name: "nish"},
		})
	}, 0)

	id := uuid.Must(uuid.NewUUID())

	deploymentInfo := terraformWorkflow.DeploymentInfo{
		Revision:   rev,
		CheckRunID: 1,
		Root:       terraform.Root{Name: "root", Trigger: terraform.ManualTrigger},
		ID:         id,
		Repo:       github.Repo{Name: "nish"},
	}
	env.ExecuteWorkflow(testWorkflow, req{
		ID: id,
		Current: queue.CurrentDeployment{
			Deployment: deploymentInfo,
			Status:     queue.InProgressStatus,
		},
	})
	env.AssertExpectations(t)
	assert.True(t, env.IsWorkflowCompleted())

	var resp response
	err := env.GetWorkflowResult(&resp)
	assert.NoError(t, err)
	// should not add in progress to the queue
	assert.Empty(t, resp.Queue)
}
