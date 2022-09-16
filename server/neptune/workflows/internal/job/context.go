package job

import (
	"go.temporal.io/sdk/workflow"
)

// ExecutionContext wraps the workflow context with other info needed to execute a step
type ExecutionContext struct {
	Path      string
	Envs      map[string]string
	TfVersion string
	workflow.Context
	JobID string
}

type Job struct {
	Steps []Step
}

type JobInstance struct {
	Job
	JobID string
}
