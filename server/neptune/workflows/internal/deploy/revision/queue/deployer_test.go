package queue_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/deployment"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
	model "github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/revision/queue"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

type ErrorType string

const (
	PlanRejectionError   ErrorType = "PlanRejectionError"
	TerraformClientError ErrorType = "TerraformClientError"
)

type testTerraformWorkflowRunner struct {
	expectedDeployment terraform.DeploymentInfo
	expectedErrorType  ErrorType
}

func (r testTerraformWorkflowRunner) Run(ctx workflow.Context, deploymentInfo terraform.DeploymentInfo, diffDirection activities.DiffDirection) error {
	if r.expectedErrorType == PlanRejectionError {
		return terraform.NewPlanRejectionError("plan rejected")
	} else if r.expectedErrorType == TerraformClientError {
		return activities.NewTerraformClientError(errors.New("error"))
	}
	return nil
}

type testDeployActivity struct{}

func (t *testDeployActivity) FetchLatestDeployment(ctx context.Context, deployerRequest activities.FetchLatestDeploymentRequest) (activities.FetchLatestDeploymentResponse, error) {
	return activities.FetchLatestDeploymentResponse{}, nil
}

func (t *testDeployActivity) StoreLatestDeployment(ctx context.Context, deployerRequest activities.StoreLatestDeploymentRequest) error {
	return nil
}

func (t *testDeployActivity) GithubCompareCommit(ctx context.Context, deployerRequest activities.CompareCommitRequest) (activities.CompareCommitResponse, error) {
	return activities.CompareCommitResponse{}, nil
}

func (t *testDeployActivity) GithubUpdateCheckRun(ctx context.Context, deployerRequest activities.UpdateCheckRunRequest) (activities.UpdateCheckRunResponse, error) {
	return activities.UpdateCheckRunResponse{}, nil
}

func (t *testDeployActivity) AuditJob(ctx context.Context, request activities.AuditJobRequest) error {
	return nil
}

type deployerRequest struct {
	Info         terraform.DeploymentInfo
	LatestDeploy *deployment.Info
	ErrType      ErrorType
}

func testDeployerWorkflow(ctx workflow.Context, r deployerRequest) (*deployment.Info, error) {
	options := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 5 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, options)
	var a *testDeployActivity

	deployer := &queue.Deployer{
		Activities: a,
		TerraformWorkflowRunner: &testTerraformWorkflowRunner{
			expectedDeployment: r.Info,
			expectedErrorType:  r.ErrType,
		},
	}

	return deployer.Deploy(ctx, r.Info, r.LatestDeploy, metrics.NewNullableScope())
}

func TestDeployer_FirstDeploy(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	da := &testDeployActivity{}
	env.RegisterActivity(da)

	repo := github.Repo{
		Owner: "owner",
		Name:  "test",
	}

	root := model.Root{
		Name: "root_1",
	}

	deploymentInfo := terraform.DeploymentInfo{
		ID:         uuid.UUID{},
		Revision:   "3455",
		CheckRunID: 1234,
		Root:       root,
		Repo:       repo,
	}

	latestDeployedRevision := &deployment.Info{
		ID:       deploymentInfo.ID.String(),
		Version:  1.0,
		Revision: "3455",
		Root: deployment.Root{
			Name: deploymentInfo.Root.Name,
		},
		Repo: deployment.Repo{
			Owner: deploymentInfo.Repo.Owner,
			Name:  deploymentInfo.Repo.Name,
		},
	}

	storeDeploymentRequest := activities.StoreLatestDeploymentRequest{
		DeploymentInfo: &deployment.Info{
			Version:  deployment.InfoSchemaVersion,
			ID:       deploymentInfo.ID.String(),
			Revision: deploymentInfo.Revision,
			Root: deployment.Root{
				Name: deploymentInfo.Root.Name,
			},
			Repo: deployment.Repo{
				Owner: deploymentInfo.Repo.Owner,
				Name:  deploymentInfo.Repo.Name,
			},
		},
	}

	env.OnActivity(da.StoreLatestDeployment, mock.Anything, storeDeploymentRequest).Return(nil)

	env.ExecuteWorkflow(testDeployerWorkflow, deployerRequest{
		Info: deploymentInfo,
	})

	env.AssertExpectations(t)

	var resp *deployment.Info
	err := env.GetWorkflowResult(&resp)
	assert.NoError(t, err)

	assert.Equal(t, latestDeployedRevision, resp)
}

func TestDeployer_CompareCommit_DeployAhead(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	da := &testDeployActivity{}
	env.RegisterActivity(da)

	repo := github.Repo{
		Owner: "owner",
		Name:  "test",
	}

	root := model.Root{
		Name: "root_1",
	}

	deploymentInfo := terraform.DeploymentInfo{
		ID:         uuid.UUID{},
		Revision:   "3455",
		CheckRunID: 1234,
		Root:       root,
		Repo:       repo,
	}

	latestDeployedRevision := &deployment.Info{
		ID:       deploymentInfo.ID.String(),
		Version:  1.0,
		Revision: "3255",
		Root: deployment.Root{
			Name: deploymentInfo.Root.Name,
		},
		Repo: deployment.Repo{
			Owner: deploymentInfo.Repo.Owner,
			Name:  deploymentInfo.Repo.Name,
		},
	}

	storeDeploymentRequest := activities.StoreLatestDeploymentRequest{
		DeploymentInfo: &deployment.Info{
			Version:  deployment.InfoSchemaVersion,
			ID:       deploymentInfo.ID.String(),
			Revision: deploymentInfo.Revision,
			Root: deployment.Root{
				Name: deploymentInfo.Root.Name,
			},
			Repo: deployment.Repo{
				Owner: deploymentInfo.Repo.Owner,
				Name:  deploymentInfo.Repo.Name,
			},
		},
	}

	compareCommitRequest := activities.CompareCommitRequest{
		Repo:                   repo,
		DeployRequestRevision:  deploymentInfo.Revision,
		LatestDeployedRevision: latestDeployedRevision.Revision,
	}

	compareCommitResponse := activities.CompareCommitResponse{
		CommitComparison: activities.DirectionAhead,
	}

	env.OnActivity(da.GithubCompareCommit, mock.Anything, compareCommitRequest).Return(compareCommitResponse, nil)
	env.OnActivity(da.StoreLatestDeployment, mock.Anything, storeDeploymentRequest).Return(nil)

	env.ExecuteWorkflow(testDeployerWorkflow, deployerRequest{
		Info:         deploymentInfo,
		LatestDeploy: latestDeployedRevision,
	})

	env.AssertExpectations(t)

	var resp *deployment.Info
	err := env.GetWorkflowResult(&resp)
	assert.NoError(t, err)

	assert.Equal(t, &deployment.Info{
		ID:       deploymentInfo.ID.String(),
		Version:  1.0,
		Revision: "3455",
		Root: deployment.Root{
			Name: deploymentInfo.Root.Name,
		},
		Repo: deployment.Repo{
			Owner: deploymentInfo.Repo.Owner,
			Name:  deploymentInfo.Repo.Name,
		},
	}, resp)
}

func TestDeployer_CompareCommit_Identical(t *testing.T) {
	repo := github.Repo{
		Owner: "owner",
		Name:  "test",
	}
	root := model.Root{
		Name:  "root_1",
		Rerun: true,
	}
	deploymentInfo := terraform.DeploymentInfo{
		ID:         uuid.UUID{},
		Revision:   "3455",
		CheckRunID: 1234,
		Root:       root,
		Repo:       repo,
	}

	latestDeployedRevision := &deployment.Info{
		ID:       deploymentInfo.ID.String(),
		Version:  1.0,
		Revision: "3255",
		Root: deployment.Root{
			Name: deploymentInfo.Root.Name,
		},
		Repo: deployment.Repo{
			Owner: deploymentInfo.Repo.Owner,
			Name:  deploymentInfo.Repo.Name,
		},
	}
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	da := &testDeployActivity{}
	env.RegisterActivity(da)
	compareCommitRequest := activities.CompareCommitRequest{
		Repo:                   repo,
		DeployRequestRevision:  deploymentInfo.Revision,
		LatestDeployedRevision: latestDeployedRevision.Revision,
	}

	compareCommitResponse := activities.CompareCommitResponse{
		CommitComparison: activities.DirectionIdentical,
	}

	env.OnActivity(da.GithubCompareCommit, mock.Anything, compareCommitRequest).Return(compareCommitResponse, nil)
	env.ExecuteWorkflow(testDeployerWorkflow, deployerRequest{
		Info:         deploymentInfo,
		LatestDeploy: latestDeployedRevision,
	})
	env.AssertExpectations(t)
	var resp *deployment.Info
	err := env.GetWorkflowResult(&resp)
	assert.NoError(t, err)
	assert.Equal(t, &deployment.Info{
		ID:       deploymentInfo.ID.String(),
		Version:  1.0,
		Revision: "3455",
		Root: deployment.Root{
			Name: deploymentInfo.Root.Name,
		},
		Repo: deployment.Repo{
			Owner: deploymentInfo.Repo.Owner,
			Name:  deploymentInfo.Repo.Name,
		},
	}, resp)

}

func TestDeployer_CompareCommit_SkipDeploy(t *testing.T) {
	repo := github.Repo{
		Owner: "owner",
		Name:  "test",
	}
	root := model.Root{
		Name:  "root_1",
		Rerun: true,
	}
	deploymentInfo := terraform.DeploymentInfo{
		ID:         uuid.UUID{},
		Revision:   "3455",
		CheckRunID: 1234,
		Root:       root,
		Repo:       repo,
	}

	latestDeployedRevision := &deployment.Info{
		ID:       deploymentInfo.ID.String(),
		Version:  1.0,
		Revision: "3255",
		Root: deployment.Root{
			Name: deploymentInfo.Root.Name,
		},
		Repo: deployment.Repo{
			Owner: deploymentInfo.Repo.Owner,
			Name:  deploymentInfo.Repo.Name,
		},
	}

	t.Run("behind deploy", func(t *testing.T) {
		ts := testsuite.WorkflowTestSuite{}
		env := ts.NewTestWorkflowEnvironment()

		da := &testDeployActivity{}
		env.RegisterActivity(da)
		compareCommitRequest := activities.CompareCommitRequest{
			Repo:                   repo,
			DeployRequestRevision:  deploymentInfo.Revision,
			LatestDeployedRevision: latestDeployedRevision.Revision,
		}

		compareCommitResponse := activities.CompareCommitResponse{
			CommitComparison: activities.DirectionBehind,
		}

		updateCheckRunRequest := activities.UpdateCheckRunRequest{
			Title:   terraform.BuildCheckRunTitle(deploymentInfo.Root.Name),
			State:   github.CheckRunFailure,
			Repo:    repo,
			ID:      deploymentInfo.CheckRunID,
			Summary: queue.DirectionBehindSummary,
		}

		updateCheckRunResponse := activities.UpdateCheckRunResponse{
			ID: updateCheckRunRequest.ID,
		}

		env.OnActivity(da.GithubUpdateCheckRun, mock.Anything, updateCheckRunRequest).Return(updateCheckRunResponse, nil)
		env.OnActivity(da.GithubCompareCommit, mock.Anything, compareCommitRequest).Return(compareCommitResponse, nil)

		env.ExecuteWorkflow(testDeployerWorkflow, deployerRequest{
			Info:         deploymentInfo,
			LatestDeploy: latestDeployedRevision,
		})
		env.AssertExpectations(t)
		var resp *deployment.Info
		err := env.GetWorkflowResult(&resp)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	cases := []activities.DiffDirection{activities.DirectionAhead, activities.DirectionDiverged}
	for _, diffDirection := range cases {
		t.Run(fmt.Sprintf("rerun deploy %s", diffDirection), func(t *testing.T) {
			ts := testsuite.WorkflowTestSuite{}
			env := ts.NewTestWorkflowEnvironment()

			da := &testDeployActivity{}
			env.RegisterActivity(da)
			compareCommitRequest := activities.CompareCommitRequest{
				Repo:                   repo,
				DeployRequestRevision:  deploymentInfo.Revision,
				LatestDeployedRevision: latestDeployedRevision.Revision,
			}

			compareCommitResponse := activities.CompareCommitResponse{
				CommitComparison: diffDirection,
			}

			updateCheckRunRequest := activities.UpdateCheckRunRequest{
				Title:   terraform.BuildCheckRunTitle(deploymentInfo.Root.Name),
				State:   github.CheckRunFailure,
				Repo:    repo,
				ID:      deploymentInfo.CheckRunID,
				Summary: queue.RerunNotIdenticalSummary,
			}

			updateCheckRunResponse := activities.UpdateCheckRunResponse{
				ID: updateCheckRunRequest.ID,
			}

			env.OnActivity(da.GithubUpdateCheckRun, mock.Anything, updateCheckRunRequest).Return(updateCheckRunResponse, nil)
			env.OnActivity(da.GithubCompareCommit, mock.Anything, compareCommitRequest).Return(compareCommitResponse, nil)

			env.ExecuteWorkflow(testDeployerWorkflow, deployerRequest{
				Info:         deploymentInfo,
				LatestDeploy: latestDeployedRevision,
			})
			env.AssertExpectations(t)
			var resp *deployment.Info
			err := env.GetWorkflowResult(&resp)
			assert.Error(t, err)
			assert.Nil(t, resp)
		})
	}
}

func TestDeployer_CompareCommit_DeployDiverged(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	da := &testDeployActivity{}
	env.RegisterActivity(da)

	repo := github.Repo{
		Owner: "owner",
		Name:  "test",
	}

	root := model.Root{
		Name: "root_1",
	}

	deploymentInfo := terraform.DeploymentInfo{
		ID:         uuid.UUID{},
		Revision:   "3455",
		CheckRunID: 1234,
		Root:       root,
		Repo:       repo,
	}

	latestDeployedRevision := &deployment.Info{
		ID:       deploymentInfo.ID.String(),
		Version:  1.0,
		Revision: "3255",
		Root: deployment.Root{
			Name: deploymentInfo.Root.Name,
		},
		Repo: deployment.Repo{
			Owner: deploymentInfo.Repo.Owner,
			Name:  deploymentInfo.Repo.Name,
		},
	}

	storeDeploymentRequest := activities.StoreLatestDeploymentRequest{
		DeploymentInfo: &deployment.Info{
			Version:  deployment.InfoSchemaVersion,
			ID:       deploymentInfo.ID.String(),
			Revision: deploymentInfo.Revision,
			Root: deployment.Root{
				Name: deploymentInfo.Root.Name,
			},
			Repo: deployment.Repo{
				Owner: deploymentInfo.Repo.Owner,
				Name:  deploymentInfo.Repo.Name,
			},
		},
	}

	compareCommitRequest := activities.CompareCommitRequest{
		Repo:                   repo,
		DeployRequestRevision:  deploymentInfo.Revision,
		LatestDeployedRevision: latestDeployedRevision.Revision,
	}

	compareCommitResponse := activities.CompareCommitResponse{
		CommitComparison: activities.DirectionDiverged,
	}

	env.OnActivity(da.GithubCompareCommit, mock.Anything, compareCommitRequest).Return(compareCommitResponse, nil)
	env.OnActivity(da.StoreLatestDeployment, mock.Anything, storeDeploymentRequest).Return(nil)

	env.ExecuteWorkflow(testDeployerWorkflow, deployerRequest{
		Info:         deploymentInfo,
		LatestDeploy: latestDeployedRevision,
	})

	env.AssertExpectations(t)

	var resp *deployment.Info
	err := env.GetWorkflowResult(&resp)
	assert.NoError(t, err)

	assert.Equal(t, &deployment.Info{
		ID:       deploymentInfo.ID.String(),
		Version:  1.0,
		Revision: "3455",
		Root: deployment.Root{
			Name: deploymentInfo.Root.Name,
		},
		Repo: deployment.Repo{
			Owner: deploymentInfo.Repo.Owner,
			Name:  deploymentInfo.Repo.Name,
		},
	}, resp)
}

func TestDeployer_WorkflowFailure_PlanRejection_SkipUpdateLatestDeployment(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	da := &testDeployActivity{}
	env.RegisterActivity(da)

	repo := github.Repo{
		Owner: "owner",
		Name:  "test",
	}

	root := model.Root{
		Name: "root_1",
	}

	deploymentInfo := terraform.DeploymentInfo{
		ID:         uuid.UUID{},
		Revision:   "3455",
		CheckRunID: 1234,
		Root:       root,
		Repo:       repo,
	}

	latestDeployedRevision := &deployment.Info{
		ID:       deploymentInfo.ID.String(),
		Version:  1.0,
		Revision: "3255",
		Root: deployment.Root{
			Name: deploymentInfo.Root.Name,
		},
		Repo: deployment.Repo{
			Owner: deploymentInfo.Repo.Owner,
			Name:  deploymentInfo.Repo.Name,
		},
	}

	compareCommitRequest := activities.CompareCommitRequest{
		Repo:                   repo,
		DeployRequestRevision:  deploymentInfo.Revision,
		LatestDeployedRevision: latestDeployedRevision.Revision,
	}

	compareCommitResponse := activities.CompareCommitResponse{
		CommitComparison: activities.DirectionAhead,
	}

	env.OnActivity(da.GithubCompareCommit, mock.Anything, compareCommitRequest).Return(compareCommitResponse, nil)

	env.ExecuteWorkflow(testDeployerWorkflow, deployerRequest{
		Info:         deploymentInfo,
		LatestDeploy: latestDeployedRevision,
		ErrType:      PlanRejectionError,
	})

	env.AssertExpectations(t)

	var resp *deployment.Info
	err := env.GetWorkflowResult(&resp)

	wfErr, ok := err.(*temporal.WorkflowExecutionError)
	assert.True(t, ok)

	appErr, ok := wfErr.Unwrap().(*temporal.ApplicationError)
	assert.True(t, ok)

	receivedErrType := appErr.Type()

	assert.Equal(t, "PlanRejectionError", receivedErrType)
}

func TestDeployer_TerraformClientError_UpdateLatestDeployment(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	da := &testDeployActivity{}
	env.RegisterActivity(da)

	repo := github.Repo{
		Owner: "owner",
		Name:  "test",
	}

	root := model.Root{
		Name: "root_1",
	}

	deploymentInfo := terraform.DeploymentInfo{
		ID:         uuid.UUID{},
		Revision:   "3455",
		CheckRunID: 1234,
		Root:       root,
		Repo:       repo,
	}

	latestDeployedRevision := &deployment.Info{
		ID:       deploymentInfo.ID.String(),
		Version:  1.0,
		Revision: "3255",
		Root: deployment.Root{
			Name: deploymentInfo.Root.Name,
		},
		Repo: deployment.Repo{
			Owner: deploymentInfo.Repo.Owner,
			Name:  deploymentInfo.Repo.Name,
		},
	}

	compareCommitRequest := activities.CompareCommitRequest{
		Repo:                   repo,
		DeployRequestRevision:  deploymentInfo.Revision,
		LatestDeployedRevision: latestDeployedRevision.Revision,
	}

	compareCommitResponse := activities.CompareCommitResponse{
		CommitComparison: activities.DirectionAhead,
	}

	storeDeploymentRequest := activities.StoreLatestDeploymentRequest{
		DeploymentInfo: &deployment.Info{
			Version:  deployment.InfoSchemaVersion,
			ID:       deploymentInfo.ID.String(),
			Revision: deploymentInfo.Revision,
			Root: deployment.Root{
				Name: deploymentInfo.Root.Name,
			},
			Repo: deployment.Repo{
				Owner: deploymentInfo.Repo.Owner,
				Name:  deploymentInfo.Repo.Name,
			},
		},
	}

	env.OnActivity(da.GithubCompareCommit, mock.Anything, compareCommitRequest).Return(compareCommitResponse, nil)
	env.OnActivity(da.StoreLatestDeployment, mock.Anything, storeDeploymentRequest).Return(nil)

	env.ExecuteWorkflow(testDeployerWorkflow, deployerRequest{
		Info:         deploymentInfo,
		LatestDeploy: latestDeployedRevision,
		ErrType:      TerraformClientError,
	})

	env.AssertExpectations(t)

	var resp *deployment.Info
	err := env.GetWorkflowResult(&resp)

	wfErr, ok := err.(*temporal.WorkflowExecutionError)
	assert.True(t, ok)

	appErr, ok := wfErr.Unwrap().(*temporal.ApplicationError)
	assert.True(t, ok)

	assert.Equal(t, "TerraformClientError", appErr.Type())
}
