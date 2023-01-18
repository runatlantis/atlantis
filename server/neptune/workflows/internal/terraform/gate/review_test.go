package gate_test

import (
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform/gate"
	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

type res struct {
	Status                        gate.PlanStatus
	ActionsClientCalled           bool
	ActionsClientCapturedApproval terraform.PlanApproval
}

type req struct {
	PlanSummary      terraform.PlanSummary
	ApprovalOverride terraform.PlanApproval
}

type testClient struct {
	called           bool
	capturedApproval terraform.PlanApproval
}

func (c *testClient) UpdateApprovalActions(approval terraform.PlanApproval) error {
	c.called = true
	c.capturedApproval = approval

	return nil
}

func testReviewWorkflow(ctx workflow.Context, r req) (res, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToCloseTimeout: 30 * time.Second,
	})

	c := &testClient{}
	review := gate.Review{
		Timeout:        10 * time.Second,
		MetricsHandler: client.MetricsNopHandler,
		Client:         c,
	}

	status, err := review.Await(ctx, terraform.Root{
		Plan: terraform.PlanJob{
			Approval: r.ApprovalOverride,
		},
	}, r.PlanSummary)

	return res{
		Status:                        status,
		ActionsClientCalled:           c.called,
		ActionsClientCapturedApproval: c.capturedApproval,
	}, err
}

func TestAwait_timesOut(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()

	approvalOverride := terraform.PlanApproval{
		Type: terraform.ManualApproval,
	}

	env.ExecuteWorkflow(testReviewWorkflow, req{PlanSummary: terraform.PlanSummary{
		Updates: []terraform.ResourceSummary{
			{
				Address: "addr",
			},
		},
	}, ApprovalOverride: approvalOverride})

	var r res
	err := env.GetWorkflowResult(&r)
	assert.NoError(t, err)

	assert.Equal(t, r, res{
		Status:                        gate.Rejected,
		ActionsClientCalled:           true,
		ActionsClientCapturedApproval: approvalOverride,
	})
}

func TestAwait_approvesEmptyPlan(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()

	env.ExecuteWorkflow(testReviewWorkflow, req{
		ApprovalOverride: terraform.PlanApproval{
			Type: terraform.ManualApproval,
		},
	})

	var r res
	err := env.GetWorkflowResult(&r)
	assert.NoError(t, err)

	assert.Equal(t, r, res{
		Status: gate.Approved,
	})
}
func TestAwait_autoApprove(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()

	env.ExecuteWorkflow(testReviewWorkflow, req{})

	var r res
	err := env.GetWorkflowResult(&r)
	assert.NoError(t, err)

	assert.Equal(t, r, res{
		Status: gate.Approved,
	})

}
