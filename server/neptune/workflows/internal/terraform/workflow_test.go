package terraform_test

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"testing"
	"time"

	"go.temporal.io/sdk/client"

	"github.com/runatlantis/atlantis/server/neptune/workflows/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/execute"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
	terraformModel "github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform/gate"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

const (
	testRepoName     = "testrepo"
	testRootName     = "testroot"
	testDeploymentID = "123"
	testPath         = "rel/path"
	DeployDir        = "deployments/123"
)

var testGithubRepo = github.Repo{
	Name: testRepoName,
}

var approvalReason = "Because I want"

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
			Approval: terraformModel.PlanApproval{
				Type:   terraformModel.ManualApproval,
				Reason: approvalReason,
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
		LocalRoot:       testLocalRoot,
		DeployDirectory: DeployDir,
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
	return activities.TerraformPlanResponse{
		Summary: terraformModel.PlanSummary{
			Updates: []terraformModel.ResourceSummary{
				{
					Address: "addr",
				},
			},
		},
	}, r.expectedError
}

type request struct {
	// bool since our errors are not serializable using json
	ShouldErrorDuringJobUpdate bool
}

type response struct {
	States           []state.Workflow
	PlanRejected     bool
	UpdateJobErrored bool
	ClientErrored    bool
}

func testTerraformWorkflow(ctx workflow.Context, req request) (*response, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToCloseTimeout: 30 * time.Second,
	})

	var gAct *githubActivities
	var tAct *terraformActivities

	var expectedError error
	runner := &jobRunner{
		expectedError: expectedError,
	}

	var s []state.Workflow

	store := state.NewWorkflowStoreWithGenerator(
		// add a notifier which just appends to a list which allows us to
		// test every state change
		func(st *state.Workflow) error {

			if st.Plan.Status == state.InProgressJobStatus && req.ShouldErrorDuringJobUpdate {
				return fmt.Errorf("some error")
			}
			// need to copy since its a pointer and will get mutated
			s = append(s, copy(st))
			return nil
		},
		&testURLGenerator{},
	)

	runnerReq := terraform.Request{
		Root:         testLocalRoot.Root,
		Repo:         testGithubRepo,
		DeploymentID: testDeploymentID,
	}

	subject := &terraform.Runner{
		ReviewGate: &gate.Review{
			Timeout:        30 * time.Second,
			MetricsHandler: client.MetricsNopHandler,
			Client:         store,
		},
		GithubActivities:    gAct,
		TerraformActivities: tAct,
		Request:             runnerReq,
		RootFetcher: &terraform.RootFetcher{
			Request: runnerReq,
			Ta:      tAct,
			Ga:      gAct,
		},
		JobRunner:      runner,
		MetricsHandler: client.MetricsNopHandler,
		Store:          store,
	}

	var planRejected bool
	var updateJobErr bool
	if err := subject.Run(ctx); err != nil {

		var appErr *temporal.ApplicationError
		if errors.As(err, &appErr) {
			var internalAppErr terraform.ApplicationError
			e := appErr.Details(&internalAppErr)

			if e != nil {
				return nil, err
			}

			switch internalAppErr.ErrType {
			case terraform.PlanRejectedErrorType:
				planRejected = true
			case terraform.UpdateJobErrorType:
				updateJobErr = true
			default:
				return nil, err
			}

		}
	}

	return &response{
		States: s,

		// doing this so that we can still check states when we get this type of error
		PlanRejected:     planRejected,
		UpdateJobErrored: updateJobErr,
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
			OnWaitingActions: s.Apply.OnWaitingActions,
		}
	}
	copy.Result = s.Result
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
		LocalRoot:       testLocalRoot,
		DeployDirectory: DeployDir,
	}, nil)
	env.OnActivity(ta.Cleanup, mock.Anything, activities.CleanupRequest{
		DeployDirectory: DeployDir,
	}).Return(activities.CleanupResponse{}, nil)

	// send approval of plan
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("planreview", gate.PlanReviewSignalRequest{
			Status: gate.Approved,
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
				Status: state.WaitingJobStatus,
				Output: &state.JobOutput{
					URL: outputURL,
				},
				OnWaitingActions: state.JobActions{
					Actions: []state.JobAction{
						{
							ID:   state.ConfirmAction,
							Info: "Confirm this plan to proceed to apply",
						},
						{
							ID:   state.RejectAction,
							Info: "Reject this plan to prevent the apply",
						},
					},
					Summary: approvalReason,
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
				OnWaitingActions: state.JobActions{
					Actions: []state.JobAction{
						{
							ID:   state.ConfirmAction,
							Info: "Confirm this plan to proceed to apply",
						},
						{
							ID:   state.RejectAction,
							Info: "Reject this plan to prevent the apply",
						},
					},
					Summary: approvalReason,
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
				OnWaitingActions: state.JobActions{
					Actions: []state.JobAction{
						{
							ID:   state.ConfirmAction,
							Info: "Confirm this plan to proceed to apply",
						},
						{
							ID:   state.RejectAction,
							Info: "Reject this plan to prevent the apply",
						},
					},
					Summary: approvalReason,
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
				OnWaitingActions: state.JobActions{
					Actions: []state.JobAction{
						{
							ID:   state.ConfirmAction,
							Info: "Confirm this plan to proceed to apply",
						},
						{
							ID:   state.RejectAction,
							Info: "Reject this plan to prevent the apply",
						},
					},
					Summary: approvalReason,
				},
			},
			Result: state.WorkflowResult{
				Reason: state.SuccessfulCompletionReason,
				Status: state.CompleteWorkflowStatus,
			},
		},
	}, resp.States)
}

func TestUpdateJobError(t *testing.T) {
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
		DeployDirectory: DeployDir,
		LocalRoot:       testLocalRoot,
	}, nil)

	// execute workflow
	env.ExecuteWorkflow(testTerraformWorkflow, request{
		ShouldErrorDuringJobUpdate: true,
	})
	assert.True(t, env.IsWorkflowCompleted())

	var resp response
	err = env.GetWorkflowResult(&resp)
	assert.NoError(t, err)

	// assert results are expected
	env.AssertExpectations(t)
	assert.True(t, resp.UpdateJobErrored)
	assert.Equal(t, []state.Workflow{
		{
			Plan: &state.Job{
				Status: state.WaitingJobStatus,
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
		DeployDirectory: DeployDir,
		LocalRoot:       testLocalRoot,
	}, nil)

	// send rejection of plan
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("planreview", gate.PlanReviewSignalRequest{
			Status: gate.Rejected,
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
	assert.True(t, resp.PlanRejected)
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
				Status: state.WaitingJobStatus,
				Output: &state.JobOutput{
					URL: outputURL,
				},
				OnWaitingActions: state.JobActions{
					Actions: []state.JobAction{
						{
							ID:   state.ConfirmAction,
							Info: "Confirm this plan to proceed to apply",
						},
						{
							ID:   state.RejectAction,
							Info: "Reject this plan to prevent the apply",
						},
					},
					Summary: approvalReason,
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
				OnWaitingActions: state.JobActions{
					Actions: []state.JobAction{
						{
							ID:   state.ConfirmAction,
							Info: "Confirm this plan to proceed to apply",
						},
						{
							ID:   state.RejectAction,
							Info: "Reject this plan to prevent the apply",
						},
					},
					Summary: approvalReason,
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
				OnWaitingActions: state.JobActions{
					Actions: []state.JobAction{
						{
							ID:   state.ConfirmAction,
							Info: "Confirm this plan to proceed to apply",
						},
						{
							ID:   state.RejectAction,
							Info: "Reject this plan to prevent the apply",
						},
					},
					Summary: approvalReason,
				},
			},
			Result: state.WorkflowResult{
				Reason: state.InternalServiceError,
				Status: state.CompleteWorkflowStatus,
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
		DeployDirectory: DeployDir,
		LocalRoot:       testLocalRoot,
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
		DeployDirectory: DeployDir,
	}).Return(activities.CleanupResponse{}, assert.AnError)

	// send approval of plan
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("planreview", gate.PlanReviewSignalRequest{
			Status: gate.Approved,
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
