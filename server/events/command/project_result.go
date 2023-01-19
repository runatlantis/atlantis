package command

import (
	"github.com/runatlantis/atlantis/server/events/models"
)

// ProjectResult is the result of executing a plan/policy_check/apply for a specific project.
type ProjectResult struct {
	Command            Name
	SubCommand         string
	RepoRelDir         string
	Workspace          string
	Error              error
	Failure            string
	PlanSuccess        *models.PlanSuccess
	PolicyCheckSuccess *models.PolicyCheckSuccess
	ApplySuccess       string
	VersionSuccess     string
	ImportSuccess      *models.ImportSuccess
	StateRmSuccess     *models.StateRmSuccess
	ProjectName        string
}

// CommitStatus returns the vcs commit status of this project result.
func (p ProjectResult) CommitStatus() models.CommitStatus {
	if p.Error != nil {
		return models.FailedCommitStatus
	}
	if p.Failure != "" {
		return models.FailedCommitStatus
	}
	return models.SuccessCommitStatus
}

// PlanStatus returns the plan status.
func (p ProjectResult) PlanStatus() models.ProjectPlanStatus {
	switch p.Command {

	case Plan:
		if p.Error != nil {
			return models.ErroredPlanStatus
		} else if p.Failure != "" {
			return models.ErroredPlanStatus
		}
		return models.PlannedPlanStatus
	case PolicyCheck, ApprovePolicies:
		if p.Error != nil {
			return models.ErroredPolicyCheckStatus
		} else if p.Failure != "" {
			return models.ErroredPolicyCheckStatus
		}
		return models.PassedPolicyCheckStatus
	case Apply:
		if p.Error != nil {
			return models.ErroredApplyStatus
		} else if p.Failure != "" {
			return models.ErroredApplyStatus
		}
		return models.AppliedPlanStatus
	}

	panic("PlanStatus() missing a combination")
}

// IsSuccessful returns true if this project result had no errors.
func (p ProjectResult) IsSuccessful() bool {
	return p.PlanSuccess != nil || p.PolicyCheckSuccess != nil || p.ApplySuccess != ""
}
