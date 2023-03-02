package queue_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
	tfActivity "github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/notifier"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/revision/queue"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/version"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

type testCheckRunClient struct {
	expectedRequest      notifier.GithubCheckRunRequest
	expectedDeploymentID string
	expectedT            *testing.T
}

func (t *testCheckRunClient) CreateOrUpdate(ctx workflow.Context, deploymentID string, request notifier.GithubCheckRunRequest) (int64, error) {
	assert.Equal(t.expectedT, t.expectedRequest, request)
	assert.Equal(t.expectedT, t.expectedDeploymentID, deploymentID)

	return 1, nil
}

func TestLockStateUpdater_unlocked_old_version(t *testing.T) {
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

	env.OnActivity(a.GithubUpdateCheckRun, mock.Anything, updateCheckRunRequest).Return(updateCheckRunResponse, nil)
	env.OnGetVersion(version.CacheCheckRunSessions, workflow.DefaultVersion, 1).Return(workflow.DefaultVersion)

	env.ExecuteWorkflow(testUpdaterWorkflow, updaterReq{
		Queue: []terraform.DeploymentInfo{info},
	})

	err := env.GetWorkflowResult(nil)
	env.AssertExpectations(t)

	assert.NoError(t, err)
}

func TestLockStateUpdater_locked_old_version(t *testing.T) {
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
		State:   github.CheckRunActionRequired,
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

	env.OnGetVersion(version.CacheCheckRunSessions, workflow.DefaultVersion, 1).Return(workflow.DefaultVersion)
	env.OnActivity(a.GithubUpdateCheckRun, mock.Anything, updateCheckRunRequest).Return(updateCheckRunResponse, nil)

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

func TestLockStateUpdater_unlocked_new_version(t *testing.T) {
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

	env.OnGetVersion(version.CacheCheckRunSessions, workflow.DefaultVersion, 1).Return(workflow.Version(1))
	env.AssertNotCalled(t, "GithubUpdateCheckRun", mock.Anything, updateCheckRunRequest)

	env.ExecuteWorkflow(testUpdaterWorkflow, updaterReq{
		Queue: []terraform.DeploymentInfo{info},
		ExpectedRequest: notifier.GithubCheckRunRequest{
			Title: terraform.BuildCheckRunTitle(info.Root.Name),
			State: github.CheckRunQueued,
			Repo:  info.Repo,
			Sha:   info.Revision,
		},
		ExpectedDeploymentID: info.ID.String(),
		ExpectedT:            t,
	})

	err := env.GetWorkflowResult(nil)
	env.AssertExpectations(t)

	assert.NoError(t, err)
}

func TestLockStateUpdater_locked_new_version(t *testing.T) {
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
		State:   github.CheckRunActionRequired,
		Repo:    info.Repo,
		ID:      info.CheckRunID,
		Summary: "This deploy is locked from a manual deployment for revision 1234.  Unlock to proceed.",
		Actions: []github.CheckRunAction{
			github.CreateUnlockAction(),
		},
	}

	env.AssertNotCalled(t, "GithubUpdateCheckRun", mock.Anything, updateCheckRunRequest)
	env.OnGetVersion(version.CacheCheckRunSessions, workflow.DefaultVersion, 1).Return(workflow.Version(1))

	env.ExecuteWorkflow(testUpdaterWorkflow, updaterReq{
		Queue: []terraform.DeploymentInfo{info},
		Lock: queue.LockState{
			Status:   queue.LockedStatus,
			Revision: "1234",
		},
		ExpectedRequest: notifier.GithubCheckRunRequest{
			Title:   terraform.BuildCheckRunTitle(info.Root.Name),
			State:   github.CheckRunActionRequired,
			Repo:    info.Repo,
			Summary: "This deploy is locked from a manual deployment for revision 1234.  Unlock to proceed.",
			Actions: []github.CheckRunAction{
				github.CreateUnlockAction(),
			},
			Sha: info.Revision,
		},
		ExpectedDeploymentID: info.ID.String(),
		ExpectedT:            t,
	})

	err := env.GetWorkflowResult(nil)
	env.AssertExpectations(t)

	assert.NoError(t, err)
}

type updaterReq struct {
	Queue                []terraform.DeploymentInfo
	Lock                 queue.LockState
	ExpectedRequest      notifier.GithubCheckRunRequest
	ExpectedDeploymentID string
	ExpectedT            *testing.T
}

func testUpdaterWorkflow(ctx workflow.Context, r updaterReq) error {
	options := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 5 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, options)
	var a *testDeployActivity
	subject := &queue.LockStateUpdater{
		Activities: a,
		GithubCheckRunCache: &testCheckRunClient{
			expectedRequest:      r.ExpectedRequest,
			expectedDeploymentID: r.ExpectedDeploymentID,
			expectedT:            r.ExpectedT,
		},
	}

	q := queue.NewQueue(noopCallback, metrics.NewNullableScope())
	for _, i := range r.Queue {
		q.Push(i)
	}

	q.SetLockForMergedItems(ctx, r.Lock)
	subject.UpdateQueuedRevisions(ctx, q)

	return nil
}
