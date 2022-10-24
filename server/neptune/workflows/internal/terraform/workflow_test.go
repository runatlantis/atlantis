package terraform_test

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/neptune/workflows/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/execute"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
	terraformModel "github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

const (
	testRepoName     = "testrepo"
	testRootName     = "testroot"
	testDeploymentID = "123"
	testPath         = "rel/path"
)

var testGithubRepo = github.Repo{
	Name: testRepoName,
}

var testLocalRoot = &terraformModel.LocalRoot{
	Root: terraformModel.Root{
		Name: testRootName,
		Plan: terraformModel.PlanJob{
			Job: execute.Job{
				Steps: []execute.Step{
					{
						StepName: "step1",
					},
				},
			},
		},
		Apply: execute.Job{
			Steps: []execute.Step{
				{
					StepName: "step2",
				},
			},
		},
	},
	Path: testPath,
	Repo: testGithubRepo,
}

type testURLGenerator struct{}

func (g *testURLGenerator) Generate(jobID fmt.Stringer, BaseURL fmt.Stringer) (*url.URL, error) {
	return url.Parse("www.test.com/jobs/1235")
}

type githubActivities struct{}

func (a *githubActivities) FetchRoot(_ context.Context, _ activities.FetchRootRequest) (activities.FetchRootResponse, error) {
	return activities.FetchRootResponse{
		LocalRoot: testLocalRoot,
	}, nil
}

type terraformActivities struct{}

func (a *terraformActivities) Cleanup(ctx context.Context, request activities.CleanupRequest) (activities.CleanupResponse, error) {
	return activities.CleanupResponse{}, nil
}
func (a *terraformActivities) GetWorkerInfo(ctx context.Context) (*activities.GetWorkerInfoResponse, error) {
	u, err := url.Parse("www.test.com")
	return &activities.GetWorkerInfoResponse{
		ServerURL: u,
		TaskQueue: "taskqueue",
	}, err
}

type jobRunner struct {
	expectedError error
}

func (r *jobRunner) Apply(ctx workflow.Context, localRoot *terraformModel.LocalRoot, jobID string, planFile string) error {
	return r.expectedError
}

func (r *jobRunner) Plan(ctx workflow.Context, localRoot *terraformModel.LocalRoot, jobID string) (activities.TerraformPlanResponse, error) {
	return activities.TerraformPlanResponse{}, r.expectedError
}

type request struct{}

type response struct {
	States []state.Workflow
}

func testTerraformWorkflow(ctx workflow.Context, req request) (*response, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToCloseTimeout: 30 * time.Second,
	})

	var gAct *githubActivities
	var tAct *terraformActivities
	runner := &jobRunner{}

	var s []state.Workflow

	runnerReq := terraform.Request{
		Root:         testLocalRoot.Root,
		Repo:         testGithubRepo,
		DeploymentID: testDeploymentID,
	}

	subject := &terraform.Runner{
		GithubActivities:    gAct,
		TerraformActivities: tAct,
		Request:             runnerReq,
		RootFetcher: &terraform.RootFetcher{
			Request: runnerReq,
			Ta:      tAct,
			Ga:      gAct,
		},
		JobRunner: runner,
		Store: state.NewWorkflowStoreWithGenerator(
			// add a notifier which just appends to a list which allows us to
			// test every state change
			func(st *state.Workflow) error {
				// need to copy since its a pointer and will get mutated
				s = append(s, copy(st))
				return nil
			},
			&testURLGenerator{},
		),
	}

	if err := subject.Run(ctx); err != nil {
		return nil, err
	}

	return &response{
		States: s,
	}, nil
}

func copy(s *state.Workflow) state.Workflow {
	var copy state.Workflow
	if s.Plan != nil {
		copy.Plan = &state.Job{
			Status: s.Plan.Status,
			Output: &state.JobOutput{
				URL: s.Plan.Output.URL,
			},
		}
	}

	if s.Apply != nil {
		copy.Apply = &state.Job{
			Status: s.Apply.Status,
			Output: &state.JobOutput{
				URL: s.Apply.Output.URL,
			},
		}
	}
	return copy
}

func TestSuccess(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	ga := &githubActivities{}
	ta := &terraformActivities{}
	env.RegisterActivity(ga)
	env.RegisterActivity(ta)

	outputURL, err := url.Parse("www.test.com/jobs/1235")
	assert.NoError(t, err)

	// set activity expectations
	env.OnActivity(ga.FetchRoot, mock.Anything, activities.FetchRootRequest{
		Repo:         testGithubRepo,
		Root:         testLocalRoot.Root,
		DeploymentID: testDeploymentID,
	}).Return(activities.FetchRootResponse{
		LocalRoot: testLocalRoot,
	}, nil)
	env.OnActivity(ta.Cleanup, mock.Anything, activities.CleanupRequest{
		LocalRoot: testLocalRoot,
	}).Return(activities.CleanupResponse{}, nil)

	// send approval of plan
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("planreview", terraform.PlanReviewSignalRequest{
			Status: terraform.Approved,
		})
	}, 5*time.Second)

	// execute workflow
	env.ExecuteWorkflow(testTerraformWorkflow, request{})
	assert.True(t, env.IsWorkflowCompleted())

	var resp response
	err = env.GetWorkflowResult(&resp)
	assert.NoError(t, err)

	// assert results are expected
	env.AssertExpectations(t)
	assert.Equal(t, []state.Workflow{
		{
			Plan: &state.Job{
				Status: state.WaitingJobStatus,
				Output: &state.JobOutput{
					URL: outputURL,
				},
			},
		},
		{
			Plan: &state.Job{
				Status: state.InProgressJobStatus,
				Output: &state.JobOutput{
					URL: outputURL,
				},
			},
		},
		{
			Plan: &state.Job{
				Status: state.SuccessJobStatus,
				Output: &state.JobOutput{
					URL: outputURL,
				},
			},
		},
		{
			Plan: &state.Job{
				Status: state.SuccessJobStatus,
				Output: &state.JobOutput{
					URL: outputURL,
				},
			},
			Apply: &state.Job{
				Status: state.WaitingJobStatus,
				Output: &state.JobOutput{
					URL: outputURL,
				},
			},
		},
		{
			Plan: &state.Job{
				Status: state.SuccessJobStatus,
				Output: &state.JobOutput{
					URL: outputURL,
				},
			},
			Apply: &state.Job{
				Status: state.InProgressJobStatus,
				Output: &state.JobOutput{
					URL: outputURL,
				},
			},
		},
		{
			Plan: &state.Job{
				Status: state.SuccessJobStatus,
				Output: &state.JobOutput{
					URL: outputURL,
				},
			},
			Apply: &state.Job{
				Status: state.SuccessJobStatus,
				Output: &state.JobOutput{
					URL: outputURL,
				},
			},
		},
	}, resp.States)
}

func TestPlanRejection(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	ga := &githubActivities{}
	ta := &terraformActivities{}
	env.RegisterActivity(ga)
	env.RegisterActivity(ta)

	outputURL, err := url.Parse("www.test.com/jobs/1235")
	assert.NoError(t, err)

	// set activity expectations
	env.OnActivity(ga.FetchRoot, mock.Anything, activities.FetchRootRequest{
		Repo:         testGithubRepo,
		Root:         testLocalRoot.Root,
		DeploymentID: testDeploymentID,
	}).Return(activities.FetchRootResponse{
		LocalRoot: testLocalRoot,
	}, nil)

	// send rejection of plan
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("planreview", terraform.PlanReviewSignalRequest{
			Status: terraform.Rejected,
		})
	}, 5*time.Second)

	// execute workflow
	env.ExecuteWorkflow(testTerraformWorkflow, request{})
	assert.True(t, env.IsWorkflowCompleted())

	var resp response
	err = env.GetWorkflowResult(&resp)
	assert.NoError(t, err)

	// assert results are expected
	env.AssertExpectations(t)
	assert.Equal(t, []state.Workflow{
		{
			Plan: &state.Job{
				Status: state.WaitingJobStatus,
				Output: &state.JobOutput{
					URL: outputURL,
				},
			},
		},
		{
			Plan: &state.Job{
				Status: state.InProgressJobStatus,
				Output: &state.JobOutput{
					URL: outputURL,
				},
			},
		},
		{
			Plan: &state.Job{
				Status: state.SuccessJobStatus,
				Output: &state.JobOutput{
					URL: outputURL,
				},
			},
		},
		{
			Plan: &state.Job{
				Status: state.SuccessJobStatus,
				Output: &state.JobOutput{
					URL: outputURL,
				},
			},
			Apply: &state.Job{
				Status: state.WaitingJobStatus,
				Output: &state.JobOutput{
					URL: outputURL,
				},
			},
		},
		{
			Plan: &state.Job{
				Status: state.SuccessJobStatus,
				Output: &state.JobOutput{
					URL: outputURL,
				},
			},
			Apply: &state.Job{
				Status: state.RejectedJobStatus,
				Output: &state.JobOutput{
					URL: outputURL,
				},
			},
		},
	}, resp.States)
}

func TestFetchRootError(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	ga := &githubActivities{}
	ta := &terraformActivities{}
	env.RegisterActivity(ga)
	env.RegisterActivity(ta)

	// set activity expectations
	env.OnActivity(ga.FetchRoot, mock.Anything, activities.FetchRootRequest{
		Repo:         testGithubRepo,
		Root:         testLocalRoot.Root,
		DeploymentID: testDeploymentID,
	}).Return(activities.FetchRootResponse{
		LocalRoot: testLocalRoot,
	}, assert.AnError)

	// execute workflow
	env.ExecuteWorkflow(testTerraformWorkflow, request{})
	assert.True(t, env.IsWorkflowCompleted())

	var resp response
	err := env.GetWorkflowResult(&resp)
	assert.Error(t, err)

	// assert results are expected
	env.AssertExpectations(t)
}

func TestCleanupErrorReturnsNoError(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	ga := &githubActivities{}
	ta := &terraformActivities{}
	env.RegisterActivity(ga)
	env.RegisterActivity(ta)

	// set activity expectations
	env.OnActivity(ta.Cleanup, mock.Anything, activities.CleanupRequest{
		LocalRoot: testLocalRoot,
	}).Return(activities.CleanupResponse{}, assert.AnError)

	// send approval of plan
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("planreview", terraform.PlanReviewSignalRequest{
			Status: terraform.Approved,
		})
	}, 5*time.Second)

	// execute workflow
	env.ExecuteWorkflow(testTerraformWorkflow, request{})
	assert.True(t, env.IsWorkflowCompleted())

	// assert results are expected
	env.AssertExpectations(t)
	var resp response
	err := env.GetWorkflowResult(&resp)
	assert.NoError(t, err)
}
