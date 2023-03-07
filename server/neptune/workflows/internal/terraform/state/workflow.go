package state

import (
	"net/url"
	"time"

	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
)

type JobStatus string

const WorkflowStateChangeSignal = "terraform-workflow-state-change"

const (
	WaitingJobStatus    JobStatus = "waiting"
	InProgressJobStatus JobStatus = "in-progress"
	RejectedJobStatus   JobStatus = "rejected"
	FailedJobStatus     JobStatus = "failed"
	SuccessJobStatus    JobStatus = "success"
)

type WorkflowStatus int
type WorkflowCompletionReason int

const (
	InProgressWorkflowStatus WorkflowStatus = iota
	CompleteWorkflowStatus
)

const (
	UnknownCompletionReason WorkflowCompletionReason = iota
	SuccessfulCompletionReason
	InternalServiceError
	TimeoutError
	SchedulingTimeoutError
	HeartbeatTimeoutError
	ActivityDurationTimeoutError
)

type JobOutput struct {
	URL *url.URL

	// populated for plan jobs
	Summary terraform.PlanSummary
}

type JobAction struct {
	ID string

	// Info gives additional info about the action that can be used in things like tooltips
	Info string
}

func (a JobAction) ToGithubCheckRunAction() github.CheckRunAction {
	return github.CheckRunAction{
		Description: a.Info,
		Label:       a.ID,
	}
}

type JobActions struct {
	Actions []JobAction

	// Provides a form for messaging around the set actions
	Summary string
}

type Job struct {
	ID               string
	Output           *JobOutput
	OnWaitingActions JobActions
	Status           JobStatus
	StartTime        time.Time
	EndTime          time.Time
}

func (j Job) GetActions() JobActions {
	if j.Status == WaitingJobStatus {
		return j.OnWaitingActions
	}

	return JobActions{}
}

type WorkflowResult struct {
	Status WorkflowStatus
	Reason WorkflowCompletionReason
}

type Workflow struct {
	Plan   *Job
	Apply  *Job
	Result WorkflowResult
}
