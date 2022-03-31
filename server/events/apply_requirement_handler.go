package events

import (
	"github.com/runatlantis/atlantis/server/core/config/raw"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/models"
)

//go:generate pegomock generate -m --package mocks -o mocks/mock_apply_handler.go ApplyRequirement
type ApplyRequirement interface {
	ValidateProject(repoDir string, ctx models.ProjectCommandContext) (string, error)
}

type AggregateApplyRequirements struct {
	WorkingDir WorkingDir
}

func (a *AggregateApplyRequirements) ValidateProject(repoDir string, ctx models.ProjectCommandContext) (failure string, err error) {
	for _, req := range ctx.ApplyRequirements {
		switch req {
		case raw.ApprovedApplyRequirement:
			if !ctx.PullReqStatus.ApprovalStatus.IsApproved {
				return "Pull request must be approved by at least one person other than the author before running apply.", nil
			}
		// this should come before mergeability check since mergeability is a superset of this check.
		case valid.PoliciesPassedApplyReq:
			if ctx.ProjectPlanStatus == models.ErroredPolicyCheckStatus {
				return "All policies must pass for project before running apply", nil
			}
		case raw.MergeableApplyRequirement:
			if !ctx.PullReqStatus.Mergeable {
				return "Pull request must be mergeable before running apply.", nil
			}
		case raw.UnDivergedApplyRequirement:
			if a.WorkingDir.HasDiverged(ctx.Log, repoDir) {
				return "Default branch must be rebased onto pull request before running apply.", nil
			}
		}
	}
	// Passed all apply requirements configured.
	return "", nil
}
