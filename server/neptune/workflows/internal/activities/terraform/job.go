package terraform

import (
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities/execute"
)

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

type PlanJob struct {
	Mode     *PlanMode
	Approval PlanApprovalType
	execute.Job
}

func (m PlanMode) ToFlag() Flag {
	return Flag{
		Value: string(m),
	}
}
