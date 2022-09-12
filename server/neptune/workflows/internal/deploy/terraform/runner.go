package terraform

import (
	"github.com/pkg/errors"
	internalContext "github.com/runatlantis/atlantis/server/neptune/context"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/config/logger"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/root"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/sideeffect"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform/state"
	"go.temporal.io/sdk/workflow"
)

type Workflow func(ctx workflow.Context, request terraform.Request) error

type stateReceiver interface {
	Receive(ctx workflow.Context, c workflow.ReceiveChannel, checkRunID int64)
}

func NewWorkflowRunner(repo github.Repo, a receiverActivities, w Workflow) *WorkflowRunner {
	return &WorkflowRunner{
		Repo:     repo,
		Workflow: w,
		StateReceiver: &StateReceiver{
			Repo:     repo,
			Activity: a,
		},
	}
}

type WorkflowRunner struct {
	StateReceiver stateReceiver
	Repo          github.Repo
	Workflow      Workflow
}

func (r *WorkflowRunner) Run(ctx workflow.Context, checkRunID int64, revision string, root root.Root) error {
	id, err := sideeffect.GenerateUUID(ctx)

	ctx = workflow.WithValue(ctx, internalContext.DeploymentIDKey, id)

	logger.Info(ctx, "Generated id")

	if err != nil {
		return errors.Wrap(err, "generating id")
	}

	ctx = workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID: id.String(),
	})
	terraformWorkflowRequest := terraform.Request{
		Repo:         r.Repo,
		Root:         root,
		DeploymentId: id.String(),
		Revision:     revision,
	}

	future := workflow.ExecuteChildWorkflow(ctx, r.Workflow, terraformWorkflowRequest)
	return r.awaitWorkflow(ctx, future, checkRunID)
}

func (r *WorkflowRunner) awaitWorkflow(ctx workflow.Context, future workflow.ChildWorkflowFuture, checkRunID int64) error {
	var childWE workflow.Execution
	if err := future.GetChildWorkflowExecution().Get(ctx, &childWE); err != nil {
		return errors.Wrap(err, "getting child workflow execution")
	}

	selector := workflow.NewSelector(ctx)

	// our child workflow will signal us when there is a state change which we will
	// handle accordingly
	ch := workflow.GetSignalChannel(ctx, state.WorkflowStateChangeSignal)
	selector.AddReceive(ch, func(c workflow.ReceiveChannel, _ bool) {
		r.StateReceiver.Receive(ctx, c, checkRunID)
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

	if err != nil {
		return errors.Wrap(err, "executing terraform workflow")
	}
	return nil
}
