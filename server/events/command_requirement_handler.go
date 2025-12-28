// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"fmt"

	"github.com/runatlantis/atlantis/server/core/config/raw"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

//go:generate pegomock generate --package mocks -o mocks/mock_command_requirement_handler.go CommandRequirementHandler
type CommandRequirementHandler interface {
	ValidateProjectDependencies(ctx command.ProjectContext) (string, error)
	ValidatePlanProject(repoDir string, ctx command.ProjectContext) (string, error)
	ValidateApplyProject(repoDir string, ctx command.ProjectContext) (string, error)
	ValidateImportProject(repoDir string, ctx command.ProjectContext) (string, error)
}

type DefaultCommandRequirementHandler struct {
	WorkingDir WorkingDir
}

func (a *DefaultCommandRequirementHandler) ValidateProjectDependencies(ctx command.ProjectContext) (failure string, err error) {
	for _, dependOnProject := range ctx.DependsOn {

		for _, project := range ctx.PullStatus.Projects {

			if project.ProjectName == dependOnProject && project.Status != models.AppliedPlanStatus && project.Status != models.PlannedNoChangesPlanStatus {
				return fmt.Sprintf("Can't apply your project unless you apply its dependencies: [%s]", project.ProjectName), nil
			}
		}
	}

	return "", nil
}

func (a *DefaultCommandRequirementHandler) ValidatePlanProject(repoDir string, ctx command.ProjectContext) (failure string, err error) {
	return a.validateCommandRequirement(repoDir, ctx, command.Plan, ctx.PlanRequirements)
}

func (a *DefaultCommandRequirementHandler) ValidateApplyProject(repoDir string, ctx command.ProjectContext) (failure string, err error) {
	return a.validateCommandRequirement(repoDir, ctx, command.Apply, ctx.ApplyRequirements)
}

func (a *DefaultCommandRequirementHandler) ValidateImportProject(repoDir string, ctx command.ProjectContext) (failure string, err error) {
	return a.validateCommandRequirement(repoDir, ctx, command.Import, ctx.ImportRequirements)
}

func (a *DefaultCommandRequirementHandler) validateCommandRequirement(repoDir string, ctx command.ProjectContext, cmd command.Name, requirements []string) (failure string, err error) {
	for _, req := range requirements {
		switch req {
		case raw.ApprovedRequirement:
			if !ctx.PullReqStatus.ApprovalStatus.IsApproved {
				return fmt.Sprintf("Pull request must be approved according to the project's approval rules before running %s.", cmd), nil
			}
		// this should come before mergeability check since mergeability is a superset of this check.
		case valid.PoliciesPassedCommandReq:
			// We should rely on this function instead of plan status, since plan status after a failed apply will not carry the policy error over.
			if !ctx.PolicyCleared() {
				return fmt.Sprintf("All policies must pass for project before running %s.", cmd), nil
			}
		case raw.MergeableRequirement:
			if !ctx.PullReqStatus.MergeableStatus.IsMergeable {
				suffix := ""
				if ctx.PullReqStatus.MergeableStatus.Reason != "" {
					suffix = fmt.Sprintf(" (%s)", ctx.PullReqStatus.MergeableStatus.Reason)
				}
				return fmt.Sprintf("Pull request must be mergeable before running %s%s.", cmd, suffix), nil
			}
		case raw.UnDivergedRequirement:
			if a.WorkingDir.HasDiverged(ctx.Log, repoDir) {
				return fmt.Sprintf("Default branch must be rebased onto pull request before running %s.", cmd), nil
			}
		}
	}
	// Passed all requirements configured.
	return "", nil
}
