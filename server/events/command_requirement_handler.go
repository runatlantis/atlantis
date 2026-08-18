// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"fmt"
	"strings"

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
	HasUndivergedImpactFromPullHead(ctx command.ProjectContext, repoDir string, workingDir WorkingDir) (handled bool, impacted bool, err error)
}

type DefaultCommandRequirementHandler struct {
	WorkingDir            WorkingDir
	ProjectImpactResolver UndivergedProjectImpactResolver
	// VCSStatusName is the prefix Atlantis uses for its commit statuses
	// (the --vcs-status-name flag, "atlantis" by default). It is used to
	// recognise Atlantis plan statuses when scoping the mergeable requirement
	// to a single project.
	VCSStatusName string
}

func (a *DefaultCommandRequirementHandler) ValidateProjectDependencies(ctx command.ProjectContext) (failure string, err error) {
	for _, dependOnProject := range ctx.DependsOn {
		dependencyFound := false

		for _, project := range ctx.PullStatus.Projects {

			if project.ProjectName != dependOnProject {
				continue
			}
			dependencyFound = true
			if project.Status != models.AppliedPlanStatus && project.Status != models.PlannedNoChangesPlanStatus {
				return fmt.Sprintf("Can't apply your project unless you apply its dependencies: [%s]", project.ProjectName), nil
			}
		}
		if ctx.FailOnMissingDependencies && !dependencyFound {
			return fmt.Sprintf("Can't apply your project unless you apply its dependencies: [%s]", dependOnProject), nil
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
	// Only explicitly opted-in non-PR API workflows, such as drift detection and
	// remediation, may skip PR-only requirements.
	skipPRRequirements := ctx.API && ctx.Pull.Num <= 0 && ctx.SkipPRRequirements

	for _, req := range requirements {
		switch req {
		case raw.ApprovedRequirement:
			if skipPRRequirements {
				ctx.Log.Info("skipping approval requirement for opted-in API call without PR number")
				continue
			}
			if !ctx.PullReqStatus.ApprovalStatus.IsApproved {
				return fmt.Sprintf("Pull request must be approved according to the project's approval rules before running %s.", cmd), nil
			}
		// this should come before mergeability check since mergeability is a superset of this check.
		case valid.PoliciesPassedCommandReq:
			// We should rely on this function instead of plan status, since plan status after a failed apply will not carry the policy error over.
			if ctx.API && ctx.Pull.Num <= 0 && ctx.RunPolicyChecks && len(ctx.ProjectPolicyStatus) == 0 {
				return fmt.Sprintf("All policies must pass for project before running %s.", cmd), nil
			}
			if !ctx.PolicyCleared() {
				return fmt.Sprintf("All policies must pass for project before running %s.", cmd), nil
			}
		case raw.MergeableRequirement:
			if skipPRRequirements {
				ctx.Log.Info("skipping mergeable requirement for opted-in API call without PR number")
				continue
			}
			mergeableStatus := ctx.PullReqStatus.MergeableStatus
			if !mergeableStatus.IsMergeable && (cmd != command.Apply || !a.mergeableIgnoringOtherProjectPlans(ctx, mergeableStatus)) {
				suffix := ""
				if mergeableStatus.Reason != "" {
					suffix = fmt.Sprintf(" (%s)", mergeableStatus.Reason)
				}
				return fmt.Sprintf("Pull request must be mergeable before running %s%s.", cmd, suffix), nil
			}
		case raw.UnDivergedRequirement:
			diverged, err := a.hasUndivergedImpact(repoDir, ctx, cmd)
			if err != nil {
				ctx.Log.Warn("evaluating undiverged requirement has failed, falling back to full divergence check: %s", err)
				diverged = a.hasDiverged(repoDir, ctx, cmd, nil)
			}
			if diverged {
				return fmt.Sprintf("Default branch must be rebased onto pull request before running %s.", cmd), nil
			}
		}
	}
	// Passed all requirements configured.
	return "", nil
}

// mergeableIgnoringOtherProjectPlans reports whether the merge request should be
// treated as mergeable for THIS project's apply even though the MR-wide
// mergeable check failed, because every blocking commit status is either an
// Atlantis per-project plan status that belongs to another project, or the
// aggregate Atlantis plan status caused by one of those other project blockers.
//
// A per-project apply only depends on that project's own plan. A failing plan
// in an unrelated project in the same merge request must not block it. The
// project's own plan is still enforced: its own plan status keeps blocking
// here, and the plan file's existence/freshness is verified at apply time by
// the project command runner.
//
// This is conservative: if the MR is blocked for any reason other than other
// projects' plan statuses (approvals, conflicts, divergence, external CI, a
// lone aggregate plan status, or a provider that does not populate
// BlockingStatuses), it stays blocked.
func (a *DefaultCommandRequirementHandler) mergeableIgnoringOtherProjectPlans(ctx command.ProjectContext, status models.MergeableStatus) bool {
	// Only the GitLab client populates BlockingStatuses. An empty list means the
	// MR is blocked for a reason we cannot (or must not) scope per project.
	if len(status.BlockingStatuses) == 0 || a.VCSStatusName == "" {
		return false
	}
	ownPlanStatus := truncateContext(fmt.Sprintf("%s/plan: %s", a.VCSStatusName, ctx.ProjectID()))
	combinedPlanStatus := a.VCSStatusName + "/plan"
	otherProjectPlanBlocked := false
	for _, name := range status.BlockingStatuses {
		switch {
		case name == combinedPlanStatus:
			// The combined plan status may be GitLab's aggregate of other project
			// plan failures, but it can also represent failures before per-project
			// statuses exist. Only ignore it if another blocking status proves this
			// is an unrelated project failure.
			continue
		case name == ownPlanStatus:
			// This project's own plan is failing or incomplete: must block.
			return false
		case isAtlantisProjectPlanStatus(name, a.VCSStatusName):
			// Another project's per-project plan status: irrelevant to this
			// project's apply, ignore it.
			otherProjectPlanBlocked = true
			continue
		default:
			// External CI or any other non-plan blocker: must block.
			return false
		}
	}
	return otherProjectPlanBlocked
}

// isAtlantisProjectPlanStatus reports whether statusName is an Atlantis
// per-project plan commit status, "<vcsStatusName>/plan: <project>".
func isAtlantisProjectPlanStatus(statusName, vcsStatusName string) bool {
	return strings.HasPrefix(statusName, vcsStatusName+"/plan: ")
}

func (a *DefaultCommandRequirementHandler) hasUndivergedImpact(repoDir string, ctx command.ProjectContext, cmd command.Name) (bool, error) {
	if a.ProjectImpactResolver == nil {
		return a.hasDiverged(repoDir, ctx, cmd, ctx.AutoplanWhenModified), nil
	}

	var handled bool
	var impacted bool
	var err error
	if cmd == command.Plan {
		handled, impacted, err = a.ProjectImpactResolver.HasUndivergedImpactFromPullHead(ctx, repoDir, a.WorkingDir)
	} else {
		handled, impacted, err = a.ProjectImpactResolver.HasUndivergedImpact(ctx, repoDir, a.WorkingDir)
	}
	if err != nil {
		return false, err
	}
	if !handled {
		return a.hasDiverged(repoDir, ctx, cmd, ctx.AutoplanWhenModified), nil
	}
	return impacted, nil
}

func (a *DefaultCommandRequirementHandler) hasDiverged(repoDir string, ctx command.ProjectContext, cmd command.Name, autoplanWhenModified []string) bool {
	if cmd == command.Plan {
		return a.WorkingDir.HasDivergedFromPullHead(ctx.Log, repoDir, ctx.RepoRelDir, autoplanWhenModified, ctx.Pull)
	}
	return a.WorkingDir.HasDiverged(ctx.Log, repoDir, ctx.RepoRelDir, autoplanWhenModified, ctx.Pull)
}
