package terraform

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities"
	terraformActivities "github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform/state"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type PlanRejectionError struct {
	msg string
}

func NewPlanRejectionError(msg string) *PlanRejectionError {
	return &PlanRejectionError{
		msg: msg,
	}
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

func (r *WorkflowRunner) Run(ctx workflow.Context, deploymentInfo DeploymentInfo, diffDirection activities.DiffDirection) error {
	id := deploymentInfo.ID
	ctx = workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID: id.String(),

		// allows all signals to be received even in a cancellation state
		WaitForCancellation: true,
		SearchAttributes: map[string]interface{}{
			"atlantis_repository": deploymentInfo.Repo.GetFullName(),
			"atlantis_root":       deploymentInfo.Root.Name,
			"atlantis_trigger":    deploymentInfo.Root.Trigger,
			"atlantis_revision":   deploymentInfo.Revision,
		},
	})

	request := terraform.Request{
		Repo:         deploymentInfo.Repo,
		Root:         r.buildRequestRoot(deploymentInfo.Root, diffDirection),
		DeploymentID: id.String(),
		Revision:     deploymentInfo.Revision,
	}

	future := workflow.ExecuteChildWorkflow(ctx, r.Workflow, request)
	return r.awaitWorkflow(ctx, future, deploymentInfo)
}

func (r *WorkflowRunner) buildRequestRoot(root terraformActivities.Root, diffDirection activities.DiffDirection) terraformActivities.Root {
	var approvalType terraformActivities.PlanApprovalType
	var reasons []string

	if diffDirection == activities.DirectionDiverged || (root.Trigger == terraformActivities.ManualTrigger && !root.Rerun) {
		approvalType = terraformActivities.ManualApproval
	}

	if diffDirection == activities.DirectionDiverged {
		reasons = append(reasons, ":warning: Requested Revision is not ahead of deployed revision, please confirm the changes described in the plan.")
	}

	if root.Trigger == terraformActivities.ManualTrigger && !root.Rerun {
		reasons = append(reasons, ":warning: Manually Triggered Deploys must be confirmed before proceeding.")
	}

	return root.WithPlanApprovalOverride(
		terraformActivities.PlanApproval{
			Type:   approvalType,
			Reason: strings.Join(reasons, "\n"),
		},
	)
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
