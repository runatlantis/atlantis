package queue_test

import (
	"go.temporal.io/sdk/client"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
	tfActivity "github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/revision/queue"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

func TestLockStateUpdater_unlocked(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	a := &testDeployActivity{}
	env.RegisterActivity(a)

	info := terraform.DeploymentInfo{
		CheckRunID: 123,
		ID:         uuid.New(),
		Revision:   "1",
		Root: tfActivity.Root{
			Name:    "root",
			Trigger: tfActivity.MergeTrigger,
		},
		Repo: github.Repo{
			Name: "repo",
		},
	}

	updateCheckRunRequest := activities.UpdateCheckRunRequest{
		Title: terraform.BuildCheckRunTitle(info.Root.Name),
		State: github.CheckRunQueued,
		Repo:  info.Repo,
		ID:    info.CheckRunID,
	}

	updateCheckRunResponse := activities.UpdateCheckRunResponse{
		ID: updateCheckRunRequest.ID,
	}

	env.OnActivity(a.UpdateCheckRun, mock.Anything, updateCheckRunRequest).Return(updateCheckRunResponse, nil)

	env.ExecuteWorkflow(testUpdaterWorkflow, updaterReq{
		Queue: []terraform.DeploymentInfo{info},
	})

	err := env.GetWorkflowResult(nil)
	env.AssertExpectations(t)

	assert.NoError(t, err)
}

func TestLockStateUpdater_locked(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	a := &testDeployActivity{}
	env.RegisterActivity(a)

	info := terraform.DeploymentInfo{
		CheckRunID: 123,
		ID:         uuid.New(),
		Revision:   "1",
		Root: tfActivity.Root{
			Name:    "root",
			Trigger: tfActivity.MergeTrigger,
		},
		Repo: github.Repo{
			Name: "repo",
		},
	}

	updateCheckRunRequest := activities.UpdateCheckRunRequest{
		Title:   terraform.BuildCheckRunTitle(info.Root.Name),
		State:   github.CheckRunQueued,
		Repo:    info.Repo,
		ID:      info.CheckRunID,
		Summary: "This deploy is locked from a manual deployment for revision 1234.  Unlock to proceed.",
		Actions: []github.CheckRunAction{
			github.CreateUnlockAction(),
		},
	}

	updateCheckRunResponse := activities.UpdateCheckRunResponse{
		ID: updateCheckRunRequest.ID,
	}

	env.OnActivity(a.UpdateCheckRun, mock.Anything, updateCheckRunRequest).Return(updateCheckRunResponse, nil)

	env.ExecuteWorkflow(testUpdaterWorkflow, updaterReq{
		Queue: []terraform.DeploymentInfo{info},
		Lock: queue.LockState{
			Status:   queue.LockedStatus,
			Revision: "1234",
		},
	})

	err := env.GetWorkflowResult(nil)
	env.AssertExpectations(t)

	assert.NoError(t, err)
}

type updaterReq struct {
	Queue []terraform.DeploymentInfo
	Lock  queue.LockState
}

func testUpdaterWorkflow(ctx workflow.Context, r updaterReq) error {
	options := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 5 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, options)
	var a *testDeployActivity
	subject := &queue.LockStateUpdater{
		Activities: a,
	}

	q := queue.NewQueue(noopCallback, client.MetricsNopHandler)
	for _, i := range r.Queue {
		q.Push(i)
	}

	q.SetLockForMergedItems(ctx, r.Lock)
	subject.UpdateQueuedRevisions(ctx, q)

	return nil
}
