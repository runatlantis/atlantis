package terraform

import (
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/execute"
)

type PlanApproval struct {
	Type   PlanApprovalType
	Reason string
}

type PlanApprovalType int

const (
	AutoApproval PlanApprovalType = iota
	ManualApproval
)

type PlanMode string

func NewDestroyPlanMode() *PlanMode {
	m := PlanMode("destroy")
	return &m
}

type PlanJob struct {
	Mode     *PlanMode
	Approval PlanApproval
	execute.Job
}

func (m PlanJob) GetPlanMode() PlanMode {
	if m.Mode != nil {
		return *m.Mode
	}

	return PlanMode("default")
}

func (m PlanMode) ToFlag() Flag {
	return Flag{
		Value: string(m),
	}
}
