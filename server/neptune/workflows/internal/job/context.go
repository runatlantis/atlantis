package job

import (
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities/terraform"
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

type PlanApprovalType string

const (
	ManualApproval PlanApprovalType = "manual"
	AutoApproval   PlanApprovalType = "auto"
)

type PlanMode string

func NewDestroyPlanMode() *PlanMode {
	m := PlanMode("destroy")
	return &m
}

type Plan struct {
	Mode     *PlanMode
	Approval PlanApprovalType
	Terraform
}

func (m PlanMode) ToFlag() terraform.Flag {
	return terraform.Flag{
		Value: string(m),
	}
}

type Terraform struct {
	Steps []Step
}

func (j Terraform) GetSteps() []Step {
	return j.Steps
}
