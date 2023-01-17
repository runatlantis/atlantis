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
	Status gate.PlanStatus
}

type req struct {
	PlanSummary terraform.PlanSummary
}

func testReviewWorkflow(ctx workflow.Context, r req) (res, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToCloseTimeout: 30 * time.Second,
	})
	review := gate.Review{
		Timeout:        10 * time.Second,
		MetricsHandler: client.MetricsNopHandler,
	}

	status := review.Await(ctx, terraform.Root{
		Plan: terraform.PlanJob{
			Approval: terraform.PlanApproval{
				Type: terraform.ManualApproval,
			},
		},
	}, r.PlanSummary)

	return res{
		Status: status,
	}, nil
}

func TestAwait_timesOut(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()

	env.ExecuteWorkflow(testReviewWorkflow, req{PlanSummary: terraform.PlanSummary{
		Updates: []terraform.ResourceSummary{
			{
				Address: "addr",
			},
		},
	}})

	var r res
	err := env.GetWorkflowResult(&r)
	assert.NoError(t, err)

	assert.Equal(t, r, res{
		Status: gate.Rejected,
	})
}

func TestAwait_approvesEmptyPlan(t *testing.T) {
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
