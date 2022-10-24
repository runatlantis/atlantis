package revision_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/request"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/revision"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/revision/queue"
	terraformWorkflow "github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

type testQueue struct {
	Queue      []terraformWorkflow.DeploymentInfo
	LockStatus queue.LockStatus
}

func (q *testQueue) Push(msg terraformWorkflow.DeploymentInfo) {
	q.Queue = append(q.Queue, msg)
}

func (q *testQueue) SetLockStatusForMergedTrigger(status queue.LockStatus) {
	q.LockStatus = status
}

type req struct {
	ID uuid.UUID
}

type response struct {
	Queue      []terraformWorkflow.DeploymentInfo
	LockStatus queue.LockStatus
	Timeout    bool
}

type testActivities struct{}

func (a *testActivities) CreateCheckRun(ctx context.Context, request activities.CreateCheckRunRequest) (activities.CreateCheckRunResponse, error) {
	return activities.CreateCheckRunResponse{}, nil
}

func testWorkflow(ctx workflow.Context, r req) (response, error) {

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToCloseTimeout: 5 * time.Second,
	})
	var timeout bool
	queue := &testQueue{}

	var a *testActivities

	receiver := revision.NewReceiver(ctx, queue, a, func(ctx workflow.Context) (uuid.UUID, error) {
		return r.ID, nil
	})
	selector := workflow.NewSelector(ctx)

	selector.AddReceive(workflow.GetSignalChannel(ctx, "test-signal"), receiver.Receive)

	for {
		selector.Select(ctx)

		if !selector.HasPending() {
			break
		}
	}

	return response{
		Queue:      queue.Queue,
		LockStatus: queue.LockStatus,
		Timeout:    timeout,
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
				Name: "root",
			},
			Repo: request.Repo{Name: "nish"},
		})
	}, 0)

	a := &testActivities{}

	env.RegisterActivity(a)

	id := uuid.Must(uuid.NewUUID())

	env.OnActivity(a.CreateCheckRun, mock.Anything, activities.CreateCheckRunRequest{
		Title:      "atlantis/deploy: root",
		Sha:        rev,
		Repo:       github.Repo{Name: "nish"},
		ExternalID: id.String(),
	}).Return(activities.CreateCheckRunResponse{ID: 1}, nil)

	env.ExecuteWorkflow(testWorkflow, req{
		ID: id,
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
			Root:       terraform.Root{Name: "root"},
			ID:         id,
			Repo:       github.Repo{Name: "nish"},
		},
	}, resp.Queue)
	assert.Equal(t, queue.UnlockedStatus, resp.LockStatus)
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

	a := &testActivities{}

	env.RegisterActivity(a)

	id := uuid.Must(uuid.NewUUID())

	env.OnActivity(a.CreateCheckRun, mock.Anything, activities.CreateCheckRunRequest{
		Title:      "atlantis/deploy: root",
		Sha:        rev,
		Repo:       github.Repo{Name: "nish"},
		ExternalID: id.String(),
	}).Return(activities.CreateCheckRunResponse{ID: 1}, nil)

	env.ExecuteWorkflow(testWorkflow, req{
		ID: id,
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
	assert.Equal(t, queue.LockedStatus, resp.LockStatus)
	assert.False(t, resp.Timeout)
}
