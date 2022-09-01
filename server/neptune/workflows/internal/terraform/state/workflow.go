package state

import "net/url"

type JobStatus string

const WorkflowStateChangeSignal = "terraform-workflow-state-change"

const (
	InProgressJobStatus JobStatus = "in-progress"
	RejectedJobStatus   JobStatus = "rejected"
	FailedJobStatus     JobStatus = "failed"
	SuccessJobStatus    JobStatus = "success"
)

type JobOutput struct {
	URL *url.URL
}

type Job struct {
	Output *JobOutput
	Status JobStatus
}

type Workflow struct {
	Plan  *Job
	Apply *Job
}
