package terraform_test

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/github/markdown"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/root"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

type testActivities struct {
}

func (a *testActivities) UpdateCheckRun(ctx context.Context, request activities.UpdateCheckRunRequest) (activities.UpdateCheckRunResponse, error) {
	return activities.UpdateCheckRunResponse{}, nil
}

type stateReceiveRequest struct {
	StatesToSend []*state.Workflow
}

func testStateReceiveWorkflow(ctx workflow.Context, r stateReceiveRequest) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToCloseTimeout: 5 * time.Second,
	})
	ch := workflow.NewChannel(ctx)

	receiver := &terraform.StateReceiver{
		Repo:     github.Repo{Name: "hello"},
		Activity: &testActivities{},
	}

	workflow.Go(ctx, func(ctx workflow.Context) {
		for _, s := range r.StatesToSend {
			ch.Send(ctx, s)
		}
	})

	receiver.Receive(ctx, ch, terraform.DeploymentInfo{
		CheckRunID: 1,
		Root:       root.Root{Name: "root"},
	})

	return nil
}

func TestStateReceive(t *testing.T) {
	outputURL, err := url.Parse("www.nish.com")
	assert.NoError(t, err)

	jobOutput := &state.JobOutput{
		URL: outputURL,
	}

	cases := []struct {
		State                 *state.Workflow
		ExpectedCheckRunState github.CheckRunState
		ExpectedActions       []github.CheckRunAction
	}{
		{
			State: &state.Workflow{
				Plan: &state.Job{
					Output: jobOutput,
					Status: state.WaitingJobStatus,
				},
			},
			ExpectedCheckRunState: github.CheckRunPending,
		},
		{
			State: &state.Workflow{
				Plan: &state.Job{
					Output: jobOutput,
					Status: state.InProgressJobStatus,
				},
			},
			ExpectedCheckRunState: github.CheckRunPending,
		},
		{
			State: &state.Workflow{
				Plan: &state.Job{
					Output: jobOutput,
					Status: state.FailedJobStatus,
				},
			},
			ExpectedCheckRunState: github.CheckRunFailure,
		},
		{
			State: &state.Workflow{
				Plan: &state.Job{
					Output: jobOutput,
					Status: state.SuccessJobStatus,
				},
			},
			ExpectedCheckRunState: github.CheckRunPending,
		},
		{
			State: &state.Workflow{
				Plan: &state.Job{
					Output: jobOutput,
					Status: state.SuccessJobStatus,
				},
				Apply: &state.Job{
					Output: jobOutput,
					Status: state.WaitingJobStatus,
				},
			},
			ExpectedCheckRunState: github.CheckRunPending,
			ExpectedActions: []github.CheckRunAction{
				github.CreatePlanReviewAction(github.Approve),
				github.CreatePlanReviewAction(github.Reject),
			},
		},
		{
			State: &state.Workflow{
				Plan: &state.Job{
					Output: jobOutput,
					Status: state.SuccessJobStatus,
				},
				Apply: &state.Job{
					Output: jobOutput,
					Status: state.InProgressJobStatus,
				},
			},
			ExpectedCheckRunState: github.CheckRunPending,
		},
		{
			State: &state.Workflow{
				Plan: &state.Job{
					Output: jobOutput,
					Status: state.SuccessJobStatus,
				},
				Apply: &state.Job{
					Output: jobOutput,
					Status: state.FailedJobStatus,
				},
			},
			ExpectedCheckRunState: github.CheckRunFailure,
		},
		{
			State: &state.Workflow{
				Plan: &state.Job{
					Output: jobOutput,
					Status: state.SuccessJobStatus,
				},
				Apply: &state.Job{
					Output: jobOutput,
					Status: state.SuccessJobStatus,
				},
			},
			ExpectedCheckRunState: github.CheckRunSuccess,
		},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			ts := testsuite.WorkflowTestSuite{}
			env := ts.NewTestWorkflowEnvironment()

			var a = &testActivities{}
			env.RegisterActivity(a)

			env.OnActivity(a.UpdateCheckRun, mock.Anything, activities.UpdateCheckRunRequest{
				Title: "atlantis/deploy: root",
				State: c.ExpectedCheckRunState,
				Repo: github.Repo{
					Name: "hello",
				},
				Summary: markdown.RenderWorkflowStateTmpl(c.State),
				ID:      1,
				Actions: c.ExpectedActions,
			}).Return(activities.UpdateCheckRunResponse{}, nil)

			env.ExecuteWorkflow(testStateReceiveWorkflow, stateReceiveRequest{
				StatesToSend: []*state.Workflow{c.State},
			})

			env.AssertExpectations(t)

			err = env.GetWorkflowResult(nil)
			assert.NoError(t, err)

		})
	}
}
