// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"encoding/json"

	"github.com/runatlantis/atlantis/server/events/models"
)

// ProjectResult is the result of executing a plan/policy_check/apply for a specific project.
type ProjectResult struct {
	ProjectCommandOutput ProjectCommandOutput
	Command              Name
	SubCommand           string
	RepoRelDir           string
	Workspace            string
	ProjectName          string
	SilencePRComments    []string
}

// ProjectCommandOutput is the output of a plan/policy_check/apply for a specific project.
type ProjectCommandOutput struct {
	Error              error
	Failure            string
	PlanSuccess        *models.PlanSuccess
	PolicyCheckResults *models.PolicyCheckResults
	ApplySuccess       string
	VersionSuccess     string
	ImportSuccess      *models.ImportSuccess
	StateRmSuccess     *models.StateRmSuccess
}

// CommitStatus returns the vcs commit status of this project result.
func (p ProjectResult) CommitStatus() models.CommitStatus {
	if p.ProjectCommandOutput.Error != nil {
		return models.FailedCommitStatus
	}
	if p.ProjectCommandOutput.Failure != "" {
		return models.FailedCommitStatus
	}
	return models.SuccessCommitStatus
}

// PolicyStatus returns the approval status of policy sets of this project result.
func (p ProjectResult) PolicyStatus() []models.PolicySetStatus {
	var policyStatuses []models.PolicySetStatus
	if p.ProjectCommandOutput.PolicyCheckResults != nil {
		for _, policySet := range p.ProjectCommandOutput.PolicyCheckResults.PolicySetResults {
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
		if p.ProjectCommandOutput.Error != nil {
			return models.ErroredPlanStatus
		} else if p.ProjectCommandOutput.Failure != "" {
			return models.ErroredPlanStatus
		} else if p.ProjectCommandOutput.PlanSuccess.NoChanges() {
			return models.PlannedNoChangesPlanStatus
		}
		return models.PlannedPlanStatus
	case PolicyCheck, ApprovePolicies:
		if p.ProjectCommandOutput.Error != nil {
			return models.ErroredPolicyCheckStatus
		} else if p.ProjectCommandOutput.Failure != "" {
			return models.ErroredPolicyCheckStatus
		}
		return models.PassedPolicyCheckStatus
	case Apply:
		if p.ProjectCommandOutput.Error != nil {
			return models.ErroredApplyStatus
		} else if p.ProjectCommandOutput.Failure != "" {
			return models.ErroredApplyStatus
		}
		return models.AppliedPlanStatus
	}

	panic("PlanStatus() missing a combination")
}

// IsSuccessful returns true if this project result had no errors.
func (p ProjectResult) IsSuccessful() bool {
	return p.ProjectCommandOutput.PlanSuccess != nil || (p.ProjectCommandOutput.PolicyCheckResults != nil && p.ProjectCommandOutput.Error == nil && p.ProjectCommandOutput.Failure == "") || p.ProjectCommandOutput.ApplySuccess != ""
}

// MarshalJSON implements custom JSON marshaling for ProjectResult to properly serialize all fields.
func (p ProjectResult) MarshalJSON() ([]byte, error) {
	type Alias ProjectResult
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(&p),
	})
}

// MarshalJSON implements custom JSON marshaling to properly serialize the Error field.
func (p ProjectCommandOutput) MarshalJSON() ([]byte, error) {
	type Alias ProjectCommandOutput
	var errMsg *string
	if p.Error != nil {
		msg := p.Error.Error()
		errMsg = &msg
	}
	return json.Marshal(&struct {
		Error *string `json:"Error"`
		*Alias
	}{
		Error: errMsg,
		Alias: (*Alias)(&p),
	})
}
