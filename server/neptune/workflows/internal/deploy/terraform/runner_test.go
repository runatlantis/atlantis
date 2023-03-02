package terraform_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
	internalTerraform "github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/metrics"
	terraformWorkflow "github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

type testStateReceiver struct {
	payloads []testSignalPayload
}

func (r *testStateReceiver) Receive(ctx workflow.Context, c workflow.ReceiveChannel, deploymentInfo internalTerraform.DeploymentInfo) {

	var payload testSignalPayload
	c.Receive(ctx, &payload)

	r.payloads = append(r.payloads, payload)
}

type testSignalPayload struct {
	S string
}

func testTerraformWorklfowWithPlanRejectionError(ctx workflow.Context, request terraformWorkflow.Request) error {
	return temporal.NewApplicationError("some message", terraformWorkflow.PlanRejectedErrorType, terraformWorkflow.ApplicationError{ErrType: terraformWorkflow.PlanRejectedErrorType, Msg: "something"})
}

// signals parent twice with a sleep in between to mimic what our real terraform workflow would be like
func testTerraformWorkflow(ctx workflow.Context, request terraformWorkflow.Request) error {
	info := workflow.GetInfo(ctx)
	parentExecution := info.ParentWorkflowExecution

	payload := testSignalPayload{
		S: "hello",
	}

	if err := workflow.SignalExternalWorkflow(ctx, parentExecution.ID, parentExecution.RunID, state.WorkflowStateChangeSignal, payload).Get(ctx, nil); err != nil {
		return err
	}

	if err := workflow.Sleep(ctx, 5*time.Second); err != nil {
		return err
	}

	return workflow.SignalExternalWorkflow(ctx, parentExecution.ID, parentExecution.RunID, state.WorkflowStateChangeSignal, payload).Get(ctx, nil)
}

type request struct {
	PlanRejectionErr bool
	DiffDirection    activities.DiffDirection
	Info             internalTerraform.DeploymentInfo
}

type response struct {
	Payloads      []testSignalPayload
	PlanRejection bool
}

func parentWorkflow(ctx workflow.Context, r request) (response, error) {
	receiver := &testStateReceiver{}
	runner := &internalTerraform.WorkflowRunner{
		StateReceiver: receiver,
	}

	if r.PlanRejectionErr == true {
		runner.Workflow = testTerraformWorklfowWithPlanRejectionError
	} else {
		runner.Workflow = testTerraformWorkflow
	}

	if err := runner.Run(ctx, r.Info, r.DiffDirection, metrics.NewNullableScope()); err != nil {
		if _, ok := err.(internalTerraform.PlanRejectionError); ok {
			return response{
				PlanRejection: true,
			}, nil
		}
		return response{}, err
	}

	return response{
		Payloads: receiver.payloads,
	}, nil
}

func buildDeploymentInfo(t *testing.T) internalTerraform.DeploymentInfo {
	uuid := uuid.New()

	return internalTerraform.DeploymentInfo{
		ID:         uuid,
		Revision:   "1234",
		CheckRunID: 1,
		Root: terraform.Root{
			Name: "some-root",
			Plan: terraform.PlanJob{},
		},
		Repo: github.Repo{},
	}
}

func TestWorkflowRunner_Run(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	env.RegisterWorkflow(testTerraformWorkflow)

	env.ExecuteWorkflow(parentWorkflow, request{
		Info: buildDeploymentInfo(t),
	})

	var resp response
	err := env.GetWorkflowResult(&resp)
	assert.NoError(t, err)

	assert.Len(t, resp.Payloads, 2)

	for _, p := range resp.Payloads {
		assert.Equal(t, testSignalPayload{
			S: "hello",
		}, p)
	}
}

func TestWorkflowRunner_RunWithDivergedCommit(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	env.RegisterWorkflow(testTerraformWorkflow)

	r := request{
		Info:          buildDeploymentInfo(t),
		DiffDirection: activities.DirectionDiverged,
	}

	env.OnWorkflow(testTerraformWorkflow, mock.Anything, terraformWorkflow.Request{
		Root: terraform.Root{
			Name: r.Info.Root.Name,
			Plan: terraform.PlanJob{
				Approval: terraform.PlanApproval{
					Type:   terraform.ManualApproval,
					Reason: ":warning: Requested Revision is not ahead of deployed revision, please confirm the changes described in the plan.",
				},
			},
		},
		Repo:         r.Info.Repo,
		DeploymentID: r.Info.ID.String(),
		Revision:     r.Info.Revision,
	}).Return(func(ctx workflow.Context, request terraformWorkflow.Request) error {
		return nil
	})

	env.ExecuteWorkflow(parentWorkflow, r)

	env.AssertExpectations(t)

	var resp response
	err := env.GetWorkflowResult(&resp)
	assert.NoError(t, err)
}

func TestWorkflowRunner_RunWithManuallyTriggeredRoot(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	env.RegisterWorkflow(testTerraformWorkflow)

	r := request{
		Info:          buildDeploymentInfo(t),
		DiffDirection: activities.DirectionAhead,
	}

	r.Info.Root.Trigger = terraform.ManualTrigger

	env.OnWorkflow(testTerraformWorkflow, mock.Anything, terraformWorkflow.Request{
		Root: terraform.Root{
			Name:    r.Info.Root.Name,
			Trigger: terraform.ManualTrigger,
			Plan: terraform.PlanJob{
				Approval: terraform.PlanApproval{
					Type:   terraform.ManualApproval,
					Reason: ":warning: Manually Triggered Deploys must be confirmed before proceeding.",
				},
			},
		},
		Repo:         r.Info.Repo,
		DeploymentID: r.Info.ID.String(),
		Revision:     r.Info.Revision,
	}).Return(func(ctx workflow.Context, request terraformWorkflow.Request) error {
		return nil
	})

	env.ExecuteWorkflow(parentWorkflow, r)

	env.AssertExpectations(t)

	var resp response
	err := env.GetWorkflowResult(&resp)
	assert.NoError(t, err)
}

func TestWorkflowRunner_PlanRejected(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	env.RegisterWorkflow(testTerraformWorklfowWithPlanRejectionError)

	env.ExecuteWorkflow(parentWorkflow, request{
		PlanRejectionErr: true,
		Info:             buildDeploymentInfo(t),
	})

	var resp response
	err := env.GetWorkflowResult(&resp)
	assert.NoError(t, err)

	assert.True(t, resp.PlanRejection)
}
