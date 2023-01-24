package events

import (
	"github.com/runatlantis/atlantis/server/core/config/raw"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

//go:generate pegomock generate -m --package mocks -o mocks/mock_command_requirement_handler.go CommandRequirementHandler
type CommandRequirementHandler interface {
	ValidatePlanProject(repoDir string, ctx command.ProjectContext) (string, error)
	ValidateApplyProject(repoDir string, ctx command.ProjectContext) (string, error)
	ValidateImportProject(repoDir string, ctx command.ProjectContext) (string, error)
}

type DefaultCommandRequirementHandler struct {
	WorkingDir WorkingDir
}

func (a *DefaultCommandRequirementHandler) ValidatePlanProject(repoDir string, ctx command.ProjectContext) (failure string, err error) {
	for _, req := range ctx.PlanRequirements {
		switch req {
		case raw.ApprovedRequirement:
			if !ctx.PullReqStatus.ApprovalStatus.IsApproved {
				return "Pull request must be approved by at least one person other than the author before running plan.", nil
			}
		case raw.MergeableRequirement:
			if !ctx.PullReqStatus.Mergeable {
				return "Pull request must be mergeable before running plan.", nil
			}
		case raw.UnDivergedRequirement:
			if a.WorkingDir.HasDiverged(ctx.Log, repoDir) {
				return "Default branch must be rebased onto pull request before running plan.", nil
			}
		}
	}
	// Passed all plan requirements configured.
	return "", nil
}

func (a *DefaultCommandRequirementHandler) ValidateApplyProject(repoDir string, ctx command.ProjectContext) (failure string, err error) {
	for _, req := range ctx.ApplyRequirements {
		switch req {
		case raw.ApprovedRequirement:
			if !ctx.PullReqStatus.ApprovalStatus.IsApproved {
				return "Pull request must be approved by at least one person other than the author before running apply.", nil
			}
		// this should come before mergeability check since mergeability is a superset of this check.
		case valid.PoliciesPassedCommandReq:
			if ctx.ProjectPlanStatus == models.ErroredPolicyCheckStatus {
				return "All policies must pass for project before running apply.", nil
			}
		case raw.MergeableRequirement:
			if !ctx.PullReqStatus.Mergeable {
				return "Pull request must be mergeable before running apply.", nil
			}
		case raw.UnDivergedRequirement:
			if a.WorkingDir.HasDiverged(ctx.Log, repoDir) {
				return "Default branch must be rebased onto pull request before running apply.", nil
			}
		}
	}
	// Passed all apply requirements configured.
	return "", nil
}

func (a *DefaultCommandRequirementHandler) ValidateImportProject(repoDir string, ctx command.ProjectContext) (failure string, err error) {
	for _, req := range ctx.ImportRequirements {
		switch req {
		case raw.ApprovedRequirement:
			if !ctx.PullReqStatus.ApprovalStatus.IsApproved {
				return "Pull request must be approved by at least one person other than the author before running import.", nil
			}
		case raw.MergeableRequirement:
			if !ctx.PullReqStatus.Mergeable {
				return "Pull request must be mergeable before running import.", nil
			}
		case raw.UnDivergedRequirement:
			if a.WorkingDir.HasDiverged(ctx.Log, repoDir) {
				return "Default branch must be rebased onto pull request before running import.", nil
			}
		}
	}
	// Passed all import requirements configured.
	return "", nil
}
