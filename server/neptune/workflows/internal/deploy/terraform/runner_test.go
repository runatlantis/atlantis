package terraform_test

import (
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
	internalTerraform "github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/sideeffect"
	terraformWorkflow "github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform/state"
	"github.com/stretchr/testify/assert"
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

	uuid, err := sideeffect.GenerateUUID(ctx)

	if err != nil {
		return response{}, nil
	}

	if err := runner.Run(ctx, internalTerraform.DeploymentInfo{
		ID:         uuid,
		Revision:   "1234",
		CheckRunID: 1,
		Root:       terraform.Root{},
	}); err != nil {
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

func TestWorkflowRunner_Run(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	env.RegisterWorkflow(testTerraformWorkflow)

	env.ExecuteWorkflow(parentWorkflow, request{})

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

func TestWorkflowRunner_PlanRejected(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	env.RegisterWorkflow(testTerraformWorklfowWithPlanRejectionError)

	env.ExecuteWorkflow(parentWorkflow, request{
		PlanRejectionErr: true,
	})

	var resp response
	err := env.GetWorkflowResult(&resp)
	assert.NoError(t, err)

	assert.True(t, resp.PlanRejection)
}
