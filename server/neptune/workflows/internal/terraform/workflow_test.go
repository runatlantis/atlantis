package terraform_test

import (
	"context"
	"errors"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/root"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"testing"
	"time"
)

const (
	testRepoName     = "testrepo"
	testRootName     = "testroot"
	testDeploymentID = "123"
	testPath         = "rel/path"
)

type testActivities struct{}

func (a *testActivities) FetchRoot(_ context.Context, _ activities.FetchRootRequest) (activities.FetchRootResponse, error) {
	return activities.FetchRootResponse{}, nil
}

func (a *testActivities) Cleanup(_ context.Context, _ activities.CleanupRequest) (activities.CleanupResponse, error) {
	return activities.CleanupResponse{}, nil
}

func testTerraformWorkflow(ctx workflow.Context) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToCloseTimeout: 5 * time.Second,
	})
	ch := workflow.NewChannel(ctx)
	act := &testActivities{}
	testRepo := github.Repo{
		Name: testRepoName,
	}
	testRoot := root.Root{
		Name: testRootName,
	}

	// Run download step
	var downloadResponse activities.FetchRootResponse
	err := workflow.ExecuteActivity(ctx, act.FetchRoot, activities.FetchRootRequest{
		Repo:         testRepo,
		Root:         testRoot,
		DeploymentId: testDeploymentID,
	}).Get(ctx, &downloadResponse)
	if err != nil {
		return err
	}

	// TODO: run plan steps

	// Send plan approval signal
	approval := terraform.PlanReview{
		Status: terraform.Approved,
	}
	workflow.Go(ctx, func(ctx workflow.Context) {
		ch.Send(ctx, approval)
	})

	// Receive signal
	var planReview terraform.PlanReview
	ch.Receive(ctx, &planReview)
	if planReview.Status != terraform.Approved {
		return errors.New("failed to receive approval")
	}

	// TODO: run apply steps

	// Cleanup
	var cleanupResponse activities.CleanupResponse
	err = workflow.ExecuteActivity(ctx, act.Cleanup, activities.CleanupRequest{
		LocalRoot: downloadResponse.LocalRoot,
	}).Get(ctx, &cleanupResponse)
	if err != nil {
		return err
	}

	return nil
}

func Test_TerraformWorkflowSuccess(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.SetWorkerOptions(worker.Options{
		BackgroundActivityContext: context.Background(),
	})
	a := &testActivities{}
	env.RegisterActivity(a)

	testRepo := github.Repo{
		Name: testRepoName,
	}
	testRoot := root.Root{
		Name: testRootName,
	}
	testLocalRoot := &root.LocalRoot{
		Root: testRoot,
		Path: testPath,
		Repo: testRepo,
	}
	env.OnActivity(a.FetchRoot, mock.Anything, activities.FetchRootRequest{
		Repo:         testRepo,
		Root:         testRoot,
		DeploymentId: testDeploymentID,
	}).Return(activities.FetchRootResponse{
		LocalRoot: testLocalRoot,
	}, nil)
	env.OnActivity(a.Cleanup, mock.Anything, activities.CleanupRequest{
		LocalRoot: testLocalRoot,
	}).Return(activities.CleanupResponse{}, nil)

	env.ExecuteWorkflow(testTerraformWorkflow)
	env.AssertExpectations(t)
	assert.True(t, env.IsWorkflowCompleted())
	assert.NoError(t, env.GetWorkflowError())
}

func Test_TerraformWorkflow_CloneFailure(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.SetWorkerOptions(worker.Options{
		BackgroundActivityContext: context.Background(),
	})
	a := &testActivities{}
	env.RegisterActivity(a)

	testRepo := github.Repo{
		Name: testRepoName,
	}
	testRoot := root.Root{
		Name: testRootName,
	}
	env.OnActivity(a.FetchRoot, mock.Anything, activities.FetchRootRequest{
		Repo:         testRepo,
		Root:         testRoot,
		DeploymentId: testDeploymentID,
	}).Return(activities.FetchRootResponse{}, errors.New("CloneActivityError"))

	env.ExecuteWorkflow(testTerraformWorkflow)
	assert.True(t, env.IsWorkflowCompleted())
	err := env.GetWorkflowError()
	var applicationErr *temporal.ApplicationError
	assert.True(t, errors.As(err, &applicationErr))
	assert.Equal(t, "CloneActivityError", applicationErr.Error())
}

func Test_TerraformWorkflow_CleanupFailure(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.SetWorkerOptions(worker.Options{
		BackgroundActivityContext: context.Background(),
	})
	a := &testActivities{}
	env.RegisterActivity(a)

	testRepo := github.Repo{
		Name: testRepoName,
	}
	testRoot := root.Root{
		Name: testRootName,
	}
	testLocalRoot := &root.LocalRoot{
		Root: testRoot,
		Path: testPath,
		Repo: testRepo,
	}
	env.OnActivity(a.FetchRoot, mock.Anything, activities.FetchRootRequest{
		Repo:         testRepo,
		Root:         testRoot,
		DeploymentId: testDeploymentID,
	}).Return(activities.FetchRootResponse{
		LocalRoot: testLocalRoot,
	}, nil)
	env.OnActivity(a.Cleanup, mock.Anything, activities.CleanupRequest{
		LocalRoot: testLocalRoot,
	}).Return(activities.CleanupResponse{}, errors.New("CleanupActivityError"))

	env.ExecuteWorkflow(testTerraformWorkflow)
	assert.True(t, env.IsWorkflowCompleted())
	err := env.GetWorkflowError()
	var applicationErr *temporal.ApplicationError
	assert.True(t, errors.As(err, &applicationErr))
	assert.Equal(t, "CleanupActivityError", applicationErr.Error())
}
