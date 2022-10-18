package state

import (
	"net/url"

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

type JobOutput struct {
	URL *url.URL

	// populated for plan jobs
	Summary terraform.PlanSummary
}

type Job struct {
	Output *JobOutput
	Status JobStatus
}

type Workflow struct {
	Plan  *Job
	Apply *Job
}
