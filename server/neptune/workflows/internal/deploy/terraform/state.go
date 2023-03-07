package terraform

import (
	"context"
	"strconv"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/metrics"
	key "github.com/runatlantis/atlantis/server/neptune/context"

	"github.com/runatlantis/atlantis/server/neptune/workflows/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github/markdown"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/config/logger"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/notifier"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/version"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform/state"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type auditActivities interface {
	AuditJob(ctx context.Context, request activities.AuditJobRequest) error
}

type receiverActivities interface {
	auditActivities
	GithubUpdateCheckRun(ctx context.Context, request activities.UpdateCheckRunRequest) (activities.UpdateCheckRunResponse, error)
}

type StateReceiver struct {
	Activity             receiverActivities
	CheckRunSessionCache CheckRunClient
}

func (n *StateReceiver) Receive(ctx workflow.Context, c workflow.ReceiveChannel, deploymentInfo DeploymentInfo) {
	var workflowState *state.Workflow
	c.Receive(ctx, &workflowState)

	workflow.GetMetricsHandler(ctx).WithTags(map[string]string{
		metrics.SignalNameTag: state.WorkflowStateChangeSignal,
	}).Counter(metrics.SignalReceive).Inc(1)

	// TODO: if we never created a check run, there was likely some issue, we should attempt to create it again.
	if deploymentInfo.CheckRunID == 0 {
		logger.Error(ctx, "check run id is 0, skipping update of check run")
		return
	}

	// emit audit events when Apply operation is run
	if workflowState.Apply != nil {
		if err := n.emitApplyEvents(ctx, workflowState.Apply, deploymentInfo); err != nil {
			logger.Error(ctx, errors.Wrap(err, "auditing apply job event").Error())
		}
	}

	if err := n.updateCheckRun(ctx, workflowState, deploymentInfo); err != nil {
		logger.Error(ctx, "updating check run", key.ErrKey, err)
	}
}

func (n *StateReceiver) updateCheckRun(ctx workflow.Context, workflowState *state.Workflow, deploymentInfo DeploymentInfo) error {
	summary := markdown.RenderWorkflowStateTmpl(workflowState)
	checkRunState := determineCheckRunState(workflowState)

	request := notifier.GithubCheckRunRequest{
		Title:   BuildCheckRunTitle(deploymentInfo.Root.Name),
		Sha:     deploymentInfo.Revision,
		State:   checkRunState,
		Repo:    deploymentInfo.Repo,
		Summary: summary,
	}

	if workflowState.Apply != nil {
		// add any actions pertaining to the apply job
		for _, a := range workflowState.Apply.GetActions().Actions {
			request.Actions = append(request.Actions, a.ToGithubCheckRunAction())
		}
	}

	// cap our retries for non-terminal states to allow for at least some progress
	if checkRunState != github.CheckRunFailure && checkRunState != github.CheckRunSuccess {
		ctx = workflow.WithRetryPolicy(ctx, temporal.RetryPolicy{
			MaximumAttempts: 3,
		})
	}

	version := workflow.GetVersion(ctx, version.CacheCheckRunSessions, workflow.DefaultVersion, 1)

	if version == workflow.DefaultVersion {
		return workflow.ExecuteActivity(ctx, n.Activity.GithubUpdateCheckRun, activities.UpdateCheckRunRequest{
			Title:   request.Title,
			State:   request.State,
			Repo:    request.Repo,
			ID:      deploymentInfo.CheckRunID,
			Summary: request.Summary,
			Actions: request.Actions,
		}).Get(ctx, nil)
	}

	_, err := n.CheckRunSessionCache.CreateOrUpdate(ctx, deploymentInfo.ID.String(), request)
	return err
}

func (n *StateReceiver) emitApplyEvents(ctx workflow.Context, jobState *state.Job, deploymentInfo DeploymentInfo) error {
	var atlantisJobState activities.AtlantisJobState
	startTime := strconv.FormatInt(jobState.StartTime.Unix(), 10)

	var endTime string
	switch jobState.Status {
	case state.InProgressJobStatus:
		atlantisJobState = activities.AtlantisJobStateRunning
	case state.SuccessJobStatus:
		atlantisJobState = activities.AtlantisJobStateSuccess
		endTime = strconv.FormatInt(jobState.EndTime.Unix(), 10)
	case state.FailedJobStatus:
		atlantisJobState = activities.AtlantisJobStateFailure
		endTime = strconv.FormatInt(jobState.EndTime.Unix(), 10)

	// no need to emit events on other states
	default:
		return nil
	}

	auditJobReq := activities.AuditJobRequest{
		Repo:           deploymentInfo.Repo,
		Root:           deploymentInfo.Root,
		JobID:          jobState.ID,
		InitiatingUser: deploymentInfo.InitiatingUser,
		Tags:           deploymentInfo.Tags,
		Revision:       deploymentInfo.Revision,
		State:          atlantisJobState,
		StartTime:      startTime,
		EndTime:        endTime,
		IsForceApply:   deploymentInfo.Root.Trigger == terraform.ManualTrigger && !deploymentInfo.Root.Rerun,
	}

	return workflow.ExecuteActivity(ctx, n.Activity.AuditJob, auditJobReq).Get(ctx, nil)
}

func determineCheckRunState(workflowState *state.Workflow) github.CheckRunState {
	if waitingForActionOn(workflowState.Plan) || waitingForActionOn(workflowState.Apply) {
		return github.CheckRunActionRequired
	}

	if workflowState.Result.Status != state.CompleteWorkflowStatus {
		return github.CheckRunPending
	}

	if workflowState.Result.Reason == state.SuccessfulCompletionReason {
		return github.CheckRunSuccess
	}

	timeouts := []state.WorkflowCompletionReason{
		state.TimeoutError,
		state.ActivityDurationTimeoutError,
		state.HeartbeatTimeoutError,
		state.SchedulingTimeoutError,
	}

	for _, t := range timeouts {
		if workflowState.Result.Reason == t {
			return github.CheckRunTimeout
		}
	}

	return github.CheckRunFailure
}

func waitingForActionOn(job *state.Job) bool {
	return job != nil && job.Status == state.WaitingJobStatus && len(job.OnWaitingActions.Actions) > 0
}
