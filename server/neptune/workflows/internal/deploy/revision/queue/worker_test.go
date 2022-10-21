package queue_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/deployment"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/revision/queue"
	terraformWorkflow "github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

type testTerraformWorkflowRunner struct {
}

func (r testTerraformWorkflowRunner) Run(ctx workflow.Context, deploymentInfo terraformWorkflow.DeploymentInfo) error {
	return nil
}

type request struct {
	Queue []terraformWorkflow.DeploymentInfo
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
	info := workflow.GetInfo(ctx)
	execution := info.WorkflowExecution

	payload := queue.UnlockSignalRequest{User: "username"}
	if err := workflow.SignalExternalWorkflow(ctx, execution.ID, execution.RunID, queue.UnlockSignalName, payload).Get(ctx, nil); err != nil {
		return response{}, err
	}

	if err := workflow.Sleep(ctx, 1*time.Second); err != nil {
		return response{}, err
	}

	q := queue.NewQueue()

	for _, s := range r.Queue {
		q.Push(s)
	}

	var wa *testDeployActivity
	worker := queue.Worker{
		Queue: q,
		RevisionProcessor: &queue.RevisionProcessor{
			Activities:              wa,
			TerraformWorkflowRunner: &testTerraformWorkflowRunner{},
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
		EndState:     worker.GetState(),
	}, nil
}

func TestWorker_FetchLatestDeploymentOnStartupOnly(t *testing.T) {
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

	repo := github.Repo{
		Owner: "owner",
		Name:  "test",
	}

	deploymentInfoList := []terraformWorkflow.DeploymentInfo{
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

	latestDeployment := deployment.Info{
		Revision: "2345",
	}

	// Mock FetchLatestDeploymentRequest only once to assert it's only being called for the first deploy
	env.OnActivity(da.FetchLatestDeployment, mock.Anything, activities.FetchLatestDeploymentRequest{
		FullRepositoryName: repo.GetFullName(),
		RootName:           deploymentInfoList[0].Root.Name,
	}).Return(activities.FetchLatestDeploymentResponse{
		DeploymentInfo: &latestDeployment,
	}, nil)

	// Mock CompareCommits for both the deploy request
	env.OnActivity(da.CompareCommit, mock.Anything, activities.CompareCommitRequest{
		Repo:                   repo,
		DeployRequestRevision:  deploymentInfoList[0].Revision,
		LatestDeployedRevision: latestDeployment.Revision,
	}).Return(activities.CompareCommitResponse{
		CommitComparison: activities.DirectionAhead,
	}, nil)
	env.OnActivity(da.CompareCommit, mock.Anything, activities.CompareCommitRequest{
		Repo:                   repo,
		DeployRequestRevision:  deploymentInfoList[1].Revision,
		LatestDeployedRevision: deploymentInfoList[0].Revision,
	}).Return(activities.CompareCommitResponse{
		CommitComparison: activities.DirectionAhead,
	}, nil)

	// Mock StoreLatestDeploymentRequest for both requests
	env.OnActivity(da.StoreLatestDeployment, mock.Anything, activities.StoreLatestDeploymentRequest{
		DeploymentInfo: &deployment.Info{
			Version:  deployment.InfoSchemaVersion,
			ID:       uuid.UUID{}.String(),
			Revision: deploymentInfoList[0].Revision,
			Repo: deployment.Repo{
				Owner: "owner",
				Name:  "test",
			},
			Root: deployment.Root{
				Name: deploymentInfoList[0].Root.Name,
			},
		},
	}).Return(nil)
	env.OnActivity(da.StoreLatestDeployment, mock.Anything, activities.StoreLatestDeploymentRequest{
		DeploymentInfo: &deployment.Info{
			Version:  deployment.InfoSchemaVersion,
			ID:       uuid.UUID{}.String(),
			Revision: deploymentInfoList[1].Revision,
			Repo: deployment.Repo{
				Owner: "owner",
				Name:  "test",
			},
			Root: deployment.Root{
				Name: deploymentInfoList[1].Root.Name,
			},
		},
	}).Return(nil)

	env.ExecuteWorkflow(testWorkflow, request{
		Queue: deploymentInfoList,
	})

	env.AssertExpectations(t)

	var resp response
	err := env.GetWorkflowResult(&resp)
	assert.NoError(t, err)

	assert.Equal(t, queue.CompleteWorkerState, resp.EndState)
	assert.True(t, resp.QueueIsEmpty)
}

func TestWorker_CompareCommit_SkipDeploy(t *testing.T) {
	deploymentInfo, _, repo, fetchDeploymentRequest, fetchDeploymentResponse, compareCommitRequest, _ := getTestArtifacts()

	cases := []struct {
		description           string
		compareCommitResponse activities.CompareCommitResponse
		updateCheckRunRequest activities.UpdateCheckRunRequest
	}{
		{
			description: "behind",
			compareCommitResponse: activities.CompareCommitResponse{
				CommitComparison: activities.DirectionBehind,
			},
			updateCheckRunRequest: activities.UpdateCheckRunRequest{
				Title:   terraformWorkflow.BuildCheckRunTitle(deploymentInfo.Root.Name),
				State:   github.CheckRunFailure,
				Repo:    repo,
				ID:      deploymentInfo.CheckRunID,
				Summary: queue.DirectionBehindSummary,
			},
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
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

			deploymentInfoList := []terraformWorkflow.DeploymentInfo{
				deploymentInfo,
			}

			env.OnActivity(da.FetchLatestDeployment, mock.Anything, fetchDeploymentRequest).Return(fetchDeploymentResponse, nil)
			env.OnActivity(da.UpdateCheckRun, mock.Anything, c.updateCheckRunRequest).Return(activities.UpdateCheckRunResponse{
				ID: c.updateCheckRunRequest.ID,
			}, nil)
			env.OnActivity(da.CompareCommit, mock.Anything, compareCommitRequest).Return(c.compareCommitResponse, nil)

			env.ExecuteWorkflow(testWorkflow, request{
				Queue: deploymentInfoList,
			})

			env.AssertExpectations(t)

			var resp response
			err := env.GetWorkflowResult(&resp)
			assert.NoError(t, err)

			assert.Equal(t, queue.CompleteWorkerState, resp.EndState)
			assert.True(t, resp.QueueIsEmpty)
		})
	}

}

func TestWorker_CompareCommit_DeployAhead(t *testing.T) {
	deploymentInfo, _, _, fetchDeploymentRequest, fetchDeploymentResponse, compareCommitRequest, storeLatestDeploymentReq := getTestArtifacts()
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	compareCommitResponse := activities.CompareCommitResponse{
		CommitComparison: activities.DirectionAhead,
	}

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

	deploymentInfoList := []terraformWorkflow.DeploymentInfo{
		deploymentInfo,
	}

	env.OnActivity(da.FetchLatestDeployment, mock.Anything, fetchDeploymentRequest).Return(fetchDeploymentResponse, nil)
	env.OnActivity(da.CompareCommit, mock.Anything, compareCommitRequest).Return(compareCommitResponse, nil)
	env.OnActivity(da.StoreLatestDeployment, mock.Anything, storeLatestDeploymentReq).Return(nil)

	env.ExecuteWorkflow(testWorkflow, request{
		Queue: deploymentInfoList,
	})

	env.AssertExpectations(t)

	var resp response
	err := env.GetWorkflowResult(&resp)
	assert.NoError(t, err)

	assert.Equal(t, queue.CompleteWorkerState, resp.EndState)
	assert.True(t, resp.QueueIsEmpty)

}

func TestWorker_CompareCommit_DeployDiverged(t *testing.T) {
	deploymentInfo, _, _, fetchDeploymentRequest, fetchDeploymentResponse, compareCommitRequest, storeLatestDeploymentReq := getTestArtifacts()
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	compareCommitResponse := activities.CompareCommitResponse{
		CommitComparison: activities.DirectionDiverged,
	}

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

	deploymentInfoList := []terraformWorkflow.DeploymentInfo{
		deploymentInfo,
	}
	updateCheckRunRequest := activities.UpdateCheckRunRequest{
		Title:   terraformWorkflow.BuildCheckRunTitle(deploymentInfo.Root.Name),
		State:   github.CheckRunPending,
		Repo:    deploymentInfo.Repo,
		ID:      deploymentInfo.CheckRunID,
		Summary: queue.DivergedCommitsSummary,
		Actions: []github.CheckRunAction{github.CreateUnlockAction()},
	}

	env.OnActivity(da.FetchLatestDeployment, mock.Anything, fetchDeploymentRequest).Return(fetchDeploymentResponse, nil)
	env.OnActivity(da.CompareCommit, mock.Anything, compareCommitRequest).Return(compareCommitResponse, nil)
	env.OnActivity(da.StoreLatestDeployment, mock.Anything, storeLatestDeploymentReq).Return(nil)
	env.OnActivity(da.UpdateCheckRun, mock.Anything, updateCheckRunRequest).Return(activities.UpdateCheckRunResponse{
		ID: int64(1),
	}, nil)

	env.ExecuteWorkflow(testWorkflow, request{
		Queue: deploymentInfoList,
	})

	env.AssertExpectations(t)

	var resp response
	err := env.GetWorkflowResult(&resp)
	assert.NoError(t, err)

	assert.Equal(t, queue.CompleteWorkerState, resp.EndState)
	assert.True(t, resp.QueueIsEmpty)

}

func TestWorker_FirstDeploy(t *testing.T) {
	deploymentInfo, _, _, fetchDeploymentRequest, _, _, storeLatestDeploymentReq := getTestArtifacts()
	fetchDeploymentResponse := activities.FetchLatestDeploymentResponse{
		DeploymentInfo: nil,
	}

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

	deploymentInfoList := []terraformWorkflow.DeploymentInfo{
		deploymentInfo,
	}

	env.OnActivity(da.FetchLatestDeployment, mock.Anything, fetchDeploymentRequest).Return(fetchDeploymentResponse, nil)
	env.OnActivity(da.StoreLatestDeployment, mock.Anything, storeLatestDeploymentReq).Return(nil)

	env.ExecuteWorkflow(testWorkflow, request{
		Queue: deploymentInfoList,
	})

	env.AssertExpectations(t)

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

func (t *testDeployActivity) CompareCommit(ctx context.Context, request activities.CompareCommitRequest) (activities.CompareCommitResponse, error) {
	return activities.CompareCommitResponse{}, nil
}

func (t *testDeployActivity) UpdateCheckRun(ctx context.Context, request activities.UpdateCheckRunRequest) (activities.UpdateCheckRunResponse, error) {
	return activities.UpdateCheckRunResponse{}, nil
}

// Setup test artifacts for compare commit tests
func getTestArtifacts() (
	deploymentInfo terraformWorkflow.DeploymentInfo,
	latestDeployedRevision deployment.Info,
	repo github.Repo,
	fetchDeploymentRequest activities.FetchLatestDeploymentRequest,
	fetchDeploymentResponse activities.FetchLatestDeploymentResponse,
	compareCommitRequest activities.CompareCommitRequest,
	storeDeploymentRequest activities.StoreLatestDeploymentRequest,
) {
	repo = github.Repo{
		Owner: "owner",
		Name:  "test",
	}

	deploymentInfo = terraformWorkflow.DeploymentInfo{
		ID:         uuid.UUID{},
		Revision:   "1",
		CheckRunID: 1234,
		Root: terraform.Root{
			Name: "root_1",
		},
		Repo: repo,
	}

	latestDeployedRevision = deployment.Info{
		Revision: "3455",
	}

	fetchDeploymentRequest = activities.FetchLatestDeploymentRequest{
		FullRepositoryName: repo.GetFullName(),
		RootName:           deploymentInfo.Root.Name,
	}

	fetchDeploymentResponse = activities.FetchLatestDeploymentResponse{
		DeploymentInfo: &deployment.Info{
			Revision: latestDeployedRevision.Revision,
			Repo: deployment.Repo{
				Owner: "owner",
				Name:  "test",
			},
		},
	}

	compareCommitRequest = activities.CompareCommitRequest{
		Repo:                   repo,
		DeployRequestRevision:  deploymentInfo.Revision,
		LatestDeployedRevision: latestDeployedRevision.Revision,
	}

	storeDeploymentRequest = activities.StoreLatestDeploymentRequest{
		DeploymentInfo: &deployment.Info{
			Version:  deployment.InfoSchemaVersion,
			ID:       deploymentInfo.ID.String(),
			Revision: deploymentInfo.Revision,
			Root: deployment.Root{
				Name: deploymentInfo.Root.Name,
			},
			Repo: deployment.Repo{
				Owner: "owner",
				Name:  "test",
			},
		},
	}
	return
}
