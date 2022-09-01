package terraform

import (
	"context"

	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/config/logger"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/github/markdown"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform/state"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type receiverActivities interface {
	UpdateCheckRun(ctx context.Context, request activities.UpdateCheckRunRequest) (activities.UpdateCheckRunResponse, error)
}

type StateReceiver struct {
	Repo     github.Repo
	Activity receiverActivities
}

func (n *StateReceiver) Receive(ctx workflow.Context, c workflow.ReceiveChannel, checkRunID int64) {
	var workflowState *state.Workflow
	c.Receive(ctx, &workflowState)

	// TODO: if we never created a check run, there was likely some issue, we should attempt to create it again.
	if checkRunID == 0 {
		logger.Error(ctx, "check run id is 0, skipping update of check run")
		return
	}

	// this shouldn't be possible
	if workflowState.Plan == nil {
		logger.Error(ctx, "Plan job state is nil, This is likely a bug. Unable to update checks")
		return
	}

	summary := markdown.RenderWorkflowStateTmpl(workflowState)
	checkRunState, checkRunConclusion := determineCheckRunStateAndConclusion(workflowState)

	// cap our retries for non-terminal states to allow for at least some progress
	if checkRunState != github.CheckRunComplete {
		ctx = workflow.WithRetryPolicy(ctx, temporal.RetryPolicy{
			MaximumAttempts: 3,
		})
	}

	// TODO: should we block here? maybe we can just make this async
	var resp activities.UpdateCheckRunResponse
	err := workflow.ExecuteActivity(ctx, n.Activity.UpdateCheckRun, activities.UpdateCheckRunRequest{
		Title:      "atlantis/deploy",
		State:      checkRunState,
		Repo:       n.Repo,
		ID:         checkRunID,
		Summary:    summary,
		Conclusion: checkRunConclusion,
	}).Get(ctx, &resp)

	if err != nil {
		logger.Error(ctx, "updating check run", "err", err)
	}
}

func determineCheckRunStateAndConclusionFromJob(jobState *state.Job) (github.CheckRunState, github.CheckRunConclusion) {
	var checkRunState github.CheckRunState
	var checkRunConclusion github.CheckRunConclusion

	switch jobState.Status {
	case state.InProgressJobStatus:
		checkRunState = github.CheckRunPending
	case state.SuccessJobStatus:
		checkRunState = github.CheckRunComplete
		checkRunConclusion = github.CheckRunSuccess
	// applies can be rejected manually and atm we consider these failures
	case state.FailedJobStatus, state.RejectedJobStatus:
		checkRunState = github.CheckRunComplete
		checkRunConclusion = github.CheckRunFailure
	}
	return checkRunState, checkRunConclusion
}

func determineCheckRunStateAndConclusion(workflowState *state.Workflow) (github.CheckRunState, github.CheckRunConclusion) {
	if workflowState.Apply == nil {
		return determineCheckRunStateAndConclusionFromJob(workflowState.Plan)
	}

	return determineCheckRunStateAndConclusionFromJob(workflowState.Apply)
}
