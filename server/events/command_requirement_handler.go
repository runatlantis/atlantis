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

//go:generate go tool pegomock generate --package mocks -o mocks/mock_command_requirement_handler.go CommandRequirementHandler
type CommandRequirementHandler interface {
	ValidateProjectDependencies(ctx command.ProjectContext) (string, error)
	ValidatePlanProject(repoDir string, ctx command.ProjectContext) (string, error)
	ValidateApplyProject(repoDir string, ctx command.ProjectContext) (string, error)
	ValidateImportProject(repoDir string, ctx command.ProjectContext) (string, error)
}

type UndivergedProjectImpactResolver interface {
	HasUndivergedImpact(ctx command.ProjectContext, repoDir string, workingDir WorkingDir) (handled bool, impacted bool, err error)
}

type DefaultCommandRequirementHandler struct {
	WorkingDir            WorkingDir
	ProjectImpactResolver UndivergedProjectImpactResolver
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
	// Check if this is a non-PR API call (e.g., drift detection)
	// PR-specific requirements (approved, mergeable) should be skipped for these
	isNonPRAPICall := ctx.API && ctx.Pull.Num == 0

	for _, req := range requirements {
		switch req {
		case raw.ApprovedRequirement:
			// Skip approval check for non-PR API calls - approval is a PR-specific concept
			if isNonPRAPICall {
				ctx.Log.Info("skipping approval requirement for API call without PR number (e.g., drift detection)")
				continue
			}
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
			// Skip mergeability check for non-PR API calls - mergeability is a PR-specific concept
			if isNonPRAPICall {
				ctx.Log.Info("skipping mergeable requirement for API call without PR number (e.g., drift detection)")
				continue
			}
			if !ctx.PullReqStatus.MergeableStatus.IsMergeable {
				suffix := ""
				if ctx.PullReqStatus.MergeableStatus.Reason != "" {
					suffix = fmt.Sprintf(" (%s)", ctx.PullReqStatus.MergeableStatus.Reason)
				}
				return fmt.Sprintf("Pull request must be mergeable before running %s%s.", cmd, suffix), nil
			}
		case raw.UnDivergedRequirement:
			diverged, err := a.hasUndivergedImpact(repoDir, ctx)
			if err != nil {
				ctx.Log.Warn("evaluating undiverged requirement has failed, falling back to full divergence check: %s", err)
				diverged = a.WorkingDir.HasDiverged(ctx.Log, repoDir, ctx.RepoRelDir, nil, ctx.Pull)
			}
			if diverged {
				return fmt.Sprintf("Default branch must be rebased onto pull request before running %s.", cmd), nil
			}
		}
	}
	// Passed all requirements configured.
	return "", nil
}

func (a *DefaultCommandRequirementHandler) hasUndivergedImpact(repoDir string, ctx command.ProjectContext) (bool, error) {
	if a.ProjectImpactResolver == nil {
		return a.WorkingDir.HasDiverged(ctx.Log, repoDir, ctx.RepoRelDir, ctx.AutoplanWhenModified, ctx.Pull), nil
	}

	handled, impacted, err := a.ProjectImpactResolver.HasUndivergedImpact(ctx, repoDir, a.WorkingDir)
	if err != nil {
		return false, err
	}
	if !handled {
		return a.WorkingDir.HasDiverged(ctx.Log, repoDir, ctx.RepoRelDir, ctx.AutoplanWhenModified, ctx.Pull), nil
	}
	return impacted, nil
}
