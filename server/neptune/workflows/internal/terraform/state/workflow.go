package state

import (
	"net/url"
	"time"

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
	TimedOutError
)

type JobOutput struct {
	URL *url.URL

	// populated for plan jobs
	Summary terraform.PlanSummary
}

type Job struct {
	ID        string
	Output    *JobOutput
	Status    JobStatus
	StartTime time.Time
	EndTime   time.Time
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
