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
	PolicyCheckResults *models.PolicyCheckResults
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

// PolicyStatus returns the approval status of policy sets of this project result.
func (p ProjectResult) PolicyStatus() []models.PolicySetStatus {
	var policyStatuses []models.PolicySetStatus
	if p.PolicyCheckResults != nil {
		for _, policySet := range p.PolicyCheckResults.PolicySetResults {
			policyStatus := models.PolicySetStatus{
				PolicySetName: policySet.PolicySetName,
				Passed:        policySet.Passed,
				Approvals:     policySet.CurApprovals,
			}
			policyStatuses = append(policyStatuses, policyStatus)
		}
	}
	return policyStatuses
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
	return p.PlanSuccess != nil || (p.PolicyCheckResults != nil && p.Error == nil && p.Failure == "") || p.ApplySuccess != ""
}
