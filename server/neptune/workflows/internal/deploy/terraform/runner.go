package terraform

import (
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform/state"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type PlanRejectionError struct {
	msg string
}

func (e PlanRejectionError) Error() string {
	return e.msg
}

type Workflow func(ctx workflow.Context, request terraform.Request) error

type stateReceiver interface {
	Receive(ctx workflow.Context, c workflow.ReceiveChannel, deploymentInfo DeploymentInfo)
}

func NewWorkflowRunner(a receiverActivities, w Workflow) *WorkflowRunner {
	return &WorkflowRunner{
		Workflow: w,
		StateReceiver: &StateReceiver{
			Activity: a,
		},
	}
}

type WorkflowRunner struct {
	StateReceiver stateReceiver
	Workflow      Workflow
}

func (r *WorkflowRunner) Run(ctx workflow.Context, deploymentInfo DeploymentInfo) error {
	id := deploymentInfo.ID
	ctx = workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID: id.String(),
		// we shouldn't ever percolate failures up unless they are user errors (aka. Terraform specific)
		// retrying indefinitely allows us to fix whatever issue that comes about without involving the user to redeploy
		RetryPolicy: &temporal.RetryPolicy{
			NonRetryableErrorTypes: []string{terraform.TerraformClientErrorType, terraform.PlanRejectedErrorType},
		},
		SearchAttributes: map[string]interface{}{
			"Repository": deploymentInfo.Repo.GetFullName(),
			"Root":       deploymentInfo.Root.Name,
			"Trigger":    deploymentInfo.Root.Trigger,
			"Revision":   deploymentInfo.Revision,
		},
	})
	terraformWorkflowRequest := terraform.Request{
		Root:         deploymentInfo.Root,
		Repo:         deploymentInfo.Repo,
		DeploymentID: id.String(),
		Revision:     deploymentInfo.Revision,
	}

	future := workflow.ExecuteChildWorkflow(ctx, r.Workflow, terraformWorkflowRequest)
	return r.awaitWorkflow(ctx, future, deploymentInfo)
}

func (r *WorkflowRunner) awaitWorkflow(ctx workflow.Context, future workflow.ChildWorkflowFuture, deploymentInfo DeploymentInfo) error {
	selector := workflow.NewNamedSelector(ctx, "TerraformChildWorkflow")

	// our child workflow will signal us when there is a state change which we will handle accordingly.
	// if for some reason the workflow is orphaned or we are retrying it independently, we have no way
	// to really update the state since we won't be listening for that signal anymore
	// we could have moved this to the main selector in the worker however we wouldn't always have this deployment info
	// which is necessary for knowing which check run id to update.
	// TODO: figure out how to solve this
	ch := workflow.GetSignalChannel(ctx, state.WorkflowStateChangeSignal)
	selector.AddReceive(ch, func(c workflow.ReceiveChannel, _ bool) {
		r.StateReceiver.Receive(ctx, c, deploymentInfo)
	})
	var workflowComplete bool
	var err error
	selector.AddFuture(future, func(f workflow.Future) {
		workflowComplete = true
		err = f.Get(ctx, nil)
	})

	for {
		selector.Select(ctx)

		if workflowComplete {
			break
		}
	}

	// if we have an app error we should attempt to unwrap it's details into our own
	// application error and act accordingly
	var appErr *temporal.ApplicationError
	if errors.As(err, &appErr) {
		var underlyingErr terraform.ApplicationError
		detailsErr := appErr.Details(&underlyingErr)

		if detailsErr == nil && underlyingErr.ErrType == terraform.PlanRejectedErrorType {
			return PlanRejectionError{msg: underlyingErr.Msg}
		}
	}

	return errors.Wrap(err, "executing terraform workflow")
}
