// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

var (
	ErrApplyAlreadyStarted      = errors.New("this plan has an unresolved apply execution; confirm the previous execution has stopped, reconcile the state, discard the plan, and run `atlantis plan` before another apply")
	ErrPlanStatusNotFound       = errors.New("no durable plan status exists")
	ErrPlanGenerationSuperseded = errors.New("plan generation was superseded; run `atlantis plan` again")
	ErrPlanGenerationInvalid    = errors.New("invalid plan generation transition")
	// ErrApplyExecutionSuperseded commits only the release of an obsolete
	// reservation whose owner confirms no apply steps were attempted.
	ErrApplyExecutionSuperseded = fmt.Errorf("%w: pre-execution reservation released", ErrPlanGenerationSuperseded)
	ErrApplyExecutionAmbiguous  = errors.New("apply changed infrastructure after its plan generation was superseded; reconcile the state and run `atlantis plan` before applying again")
)

// PlanGenerationBeginResult includes the previous identities for exact cleanup.
// Cleanup must never discover a mutable canonical object and assume it still
// belongs to one of these identities.
type PlanGenerationBeginResult struct {
	models.PullStatus
	Previous []models.ProjectStatus
}

// BeginPlanGeneration is a pure transition; backends commit it atomically.
// A generation is admitted before any of its plan steps can replace artifacts.
// Policy approvals survive admission and are filtered by the existing policy
// runner according to sticky-approval configuration.
func BeginPlanGeneration(current *models.PullStatus, pull models.PullRequest, generation string, projects []command.ProjectContext, replace bool) (PlanGenerationBeginResult, error) {
	if generation == "" {
		return PlanGenerationBeginResult{}, fmt.Errorf("%w: empty generation", ErrPlanGenerationInvalid)
	}
	result := PlanGenerationBeginResult{PullStatus: models.PullStatus{Pull: pull}}
	if current != nil {
		for _, project := range current.Projects {
			if project.PlanGeneration == generation {
				return PlanGenerationBeginResult{}, fmt.Errorf("%w: generation already admitted", ErrPlanGenerationInvalid)
			}
		}
		result.Previous = slices.Clone(current.Projects)
		if !replace && samePullIdentity(current.Pull, pull) {
			result.Projects = slices.Clone(current.Projects)
		}
	}
	// Superseding one member cancels the unfinished command as a whole.
	// Its other projects retain fail-closed rows until explicitly replanned.
	superseded := make(map[string]bool)
	for _, ctx := range projects {
		if old := findProject(result.Previous, projectIdentity{ctx.Workspace, ctx.RepoRelDir, ctx.ProjectName}); old != nil && old.PlanGenerationActive {
			superseded[old.PlanGeneration] = true
		}
	}
	for i := range result.Projects {
		if superseded[result.Projects[i].PlanGeneration] {
			result.Projects[i].PlanGenerationActive = false
		}
	}
	seen := make(map[projectIdentity]bool)
	for _, ctx := range projects {
		key := projectIdentity{ctx.Workspace, ctx.RepoRelDir, ctx.ProjectName}
		if seen[key] {
			return PlanGenerationBeginResult{}, fmt.Errorf("%w: duplicate project %q", ErrPlanGenerationInvalid, ctx.ProjectName)
		}
		seen[key] = true
		project := models.ProjectStatus{Workspace: ctx.Workspace, RepoRelDir: ctx.RepoRelDir, ProjectName: ctx.ProjectName, Status: models.ErroredPlanStatus, PlanGeneration: generation, PlanGenerationActive: true, ManagedPlan: ctx.RequiresAtlantisManagedPlanFile || slices.ContainsFunc(ctx.Steps, func(step valid.Step) bool { return step.StepName == "plan" || step.StepName == "apply" })}
		if old := findProject(result.Previous, key); old != nil {
			project.PolicyStatus = slices.Clone(old.PolicyStatus)
			project.ApplyExecutionID = old.ApplyExecutionID
		}
		if old := findProject(result.Projects, key); old != nil {
			*old = project
		} else {
			result.Projects = append(result.Projects, project)
		}
	}
	// Removing a project must not erase an unresolved infrastructure operation.
	// Ordinary no-project updates retain their existing cleanup behavior.
	for _, old := range result.Previous {
		if old.ApplyExecutionID != "" {
			kept := findProject(result.Projects, projectIdentity{old.Workspace, old.RepoRelDir, old.ProjectName})
			if kept == nil || kept.ApplyExecutionID != old.ApplyExecutionID {
				return PlanGenerationBeginResult{}, ErrApplyAlreadyStarted
			}
		}
	}
	return result, nil
}

// MergePullResults preserves ordinary legacy behavior while preventing writers
// from erasing or completing a generation they did not observe. The returned
// status must be committed even when ErrApplyExecutionAmbiguous is returned:
// successful infrastructure changes cannot be rolled back with a DB transaction.
func MergePullResults(current *models.PullStatus, pull models.PullRequest, results []command.ProjectResult) (models.PullStatus, error) {
	next := models.PullStatus{Pull: pull}
	if current != nil && samePullIdentity(current.Pull, pull) {
		next = *current
		next.Projects = slices.Clone(current.Projects)
	}
	// Validate the whole batch before installing any ordinary results. A
	// conflict in one project can arrive after another project's apply has
	// succeeded, so conflict handling must examine every executed result.
	var conflict error
	completed := make(map[string]map[projectIdentity]bool)
	for _, result := range results {
		if result.Command != command.Plan || result.PlanGeneration == "" {
			continue
		}
		if completed[result.PlanGeneration] == nil {
			completed[result.PlanGeneration] = make(map[projectIdentity]bool)
		}
		if completed[result.PlanGeneration][resultIdentity(result)] {
			conflict = fmt.Errorf("%w: duplicate project result", ErrPlanGenerationInvalid)
		}
		completed[result.PlanGeneration][resultIdentity(result)] = true
	}
	if current != nil {
		for _, project := range current.Projects {
			if selected := completed[project.PlanGeneration]; selected != nil && !selected[projectIdentity{project.Workspace, project.RepoRelDir, project.ProjectName}] {
				conflict = fmt.Errorf("%w: generation is incomplete", ErrPlanGenerationInvalid)
			}
		}
	}
	if current != nil && !samePullIdentity(current.Pull, pull) && slices.ContainsFunc(current.Projects, func(p models.ProjectStatus) bool { return p.PlanGeneration != "" || p.ApplyExecutionID != "" }) {
		conflict = ErrPlanGenerationSuperseded
	}
	for _, result := range results {
		if result.ApplyExecuted && errors.Is(result.Error, ErrApplyExecutionAmbiguous) {
			conflict = ErrApplyExecutionAmbiguous
		}
		if err := validateGenerationResult(current, pull, result); err != nil {
			conflict = err
		}
	}
	if conflict != nil {
		if current != nil {
			next = *current
			next.Projects = slices.Clone(current.Projects)
		}
		ambiguous, released := false, false
		for _, result := range results {
			if result.Command == command.Apply && result.ApplyExecutionID != "" && !result.ApplyAttempted && !result.ApplyExecuted && result.ApplySuccess == "" && (result.Error != nil || result.Failure != "") {
				if project := findProject(next.Projects, resultIdentity(result)); project != nil && project.ApplyExecutionID == result.ApplyExecutionID {
					project.ApplyExecutionID = ""
					released = true
				}
			}

			if result.Command != command.Apply || (!result.ApplyExecuted && (result.ApplySuccess == "" || result.Error != nil || result.Failure != "")) {
				continue
			}
			ambiguous = true
			if project := findProject(next.Projects, resultIdentity(result)); project != nil {
				project.Status = models.ErroredApplyStatus
				project.PlanGenerationActive = false
				project.AcceptedPlanGeneration = ""
				project.ManagedPlanHash = ""
				if result.ApplyExecutionID != "" && project.ApplyExecutionID == result.ApplyExecutionID {
					project.ApplyExecutionID = ""
				}
			}
		}
		if ambiguous {
			return next, ErrApplyExecutionAmbiguous
		}
		if released {
			return next, ErrApplyExecutionSuperseded
		}
		return next, conflict
	}

	for _, result := range results {
		key := resultIdentity(result)
		if result.ExcludeFromPlanStatus && result.Command == command.Apply {
			// A directory error precedes execution: release only this admission,
			// without replacing the existing status with an errored apply row.
			findProject(next.Projects, key).ApplyExecutionID = ""
			continue
		}
		if result.ExcludeFromPlanStatus {
			next.Projects = slices.DeleteFunc(next.Projects, func(project models.ProjectStatus) bool {
				return projectIdentity{project.Workspace, project.RepoRelDir, project.ProjectName} == key
			})
			continue
		}
		project := findProject(next.Projects, key)
		if project == nil {
			next.Projects = append(next.Projects, models.ProjectStatus{Workspace: result.Workspace, RepoRelDir: result.RepoRelDir, ProjectName: result.ProjectName})
			project = &next.Projects[len(next.Projects)-1]
			if current != nil {
				if old := findProject(current.Projects, key); old != nil {
					project.PolicyStatus = slices.Clone(old.PolicyStatus)
				}
			}
		}
		// Unlock is a cleanup operation, not evidence that a completed apply was
		// discarded. Preserve terminal apply history.
		if result.Command == command.Unlock {
			if project.Status != models.AppliedPlanStatus {
				project.Status = models.DiscardedPlanStatus
			}
		} else {
			project.Status = result.PlanStatus()
		}
		if result.Command == command.Plan && result.PlanGeneration != "" {
			project.PlanGenerationActive = false
		}
		if result.Command == command.Unlock || ((result.Command == command.Import || result.Command == command.State) && result.Error == nil && result.Failure == "") {
			project.PlanGenerationActive = false
			project.AcceptedPlanGeneration = ""
			project.ManagedPlanHash = ""
		}
		if result.Command == command.Plan && result.PlanGeneration != "" && result.Error == nil && result.Failure == "" {
			project.AcceptedPlanGeneration = result.PlanGeneration
			if project.ManagedPlan {
				project.ManagedPlanHash = result.ManagedPlanHash
			}
		}
		if result.Command == command.Apply && result.ApplyExecutionID != "" {
			if (result.Error == nil && result.Failure == "") || !result.ApplyAttempted {
				project.ApplyExecutionID = ""
			} else {
				project.AcceptedPlanGeneration = ""
				project.ManagedPlanHash = ""
			}
		}
		mergePolicies(project, result.PolicyStatus())
	}
	return next, nil
}

func mergePolicies(project *models.ProjectStatus, results []models.PolicySetStatus) {
	if len(project.PolicyStatus) == 0 {
		project.PolicyStatus = slices.Clone(results)
		return
	}
	project.PolicyStatus = slices.Clone(project.PolicyStatus)
	for i, old := range project.PolicyStatus {
		for _, policy := range results {
			if old.PolicySetName == policy.PolicySetName {
				project.PolicyStatus[i] = policy
			}
		}
	}
}

type projectIdentity struct{ workspace, dir, name string }

func resultIdentity(result command.ProjectResult) projectIdentity {
	return projectIdentity{result.Workspace, result.RepoRelDir, result.ProjectName}
}
func findProject(projects []models.ProjectStatus, key projectIdentity) *models.ProjectStatus {
	for i := range projects {
		if projects[i].Workspace == key.workspace && projects[i].RepoRelDir == key.dir && projects[i].ProjectName == key.name {
			return &projects[i]
		}
	}
	return nil
}
func samePullIdentity(left, right models.PullRequest) bool {
	return left.HeadCommit == right.HeadCommit && (right.BaseBranch == "" || left.BaseBranch == right.BaseBranch)
}

func validateGenerationResult(current *models.PullStatus, pull models.PullRequest, result command.ProjectResult) error {
	if result.ExcludeFromPlanStatus {
		planError := result.Command == command.Plan && result.PlanGeneration != ""
		applyError := result.Command == command.Apply && result.ApplyExecutionID != "" && !result.ApplyAttempted && !result.ApplyExecuted
		if result.Error == nil || (!planError && !applyError) {
			return fmt.Errorf("%w: invalid excluded result", ErrPlanGenerationInvalid)
		}
	}
	var prior *models.ProjectStatus
	if current != nil {
		prior = findProject(current.Projects, resultIdentity(result))
	}
	if result.Command == command.Apply && (result.ApplyExecutionID != "" || (prior != nil && prior.ApplyExecutionID != "")) {
		if prior == nil || result.ApplyExecutionID == "" || prior.ApplyExecutionID != result.ApplyExecutionID || prior.Status == models.DiscardedPlanStatus {
			return ErrApplyAlreadyStarted
		}
	}
	if result.PlanGeneration == "" && (prior == nil || prior.PlanGeneration == "") {
		return nil
	}
	if current == nil || !samePullIdentity(current.Pull, pull) || prior == nil || prior.PlanGeneration != result.PlanGeneration || result.PlanGeneration == "" {
		return ErrPlanGenerationSuperseded
	}
	if result.Command == command.Plan {
		if result.ExcludeFromPlanStatus && prior.ApplyExecutionID != "" {
			return ErrApplyAlreadyStarted
		}
		if !prior.PlanGenerationActive || prior.AcceptedPlanGeneration != "" || prior.Status != models.ErroredPlanStatus {
			return fmt.Errorf("%w: %w: generation already completed", ErrPlanGenerationInvalid, ErrPlanGenerationSuperseded)
		}
		if result.Error == nil && result.Failure == "" {
			if result.PlanSuccess == nil {
				return fmt.Errorf("%w: missing plan result", ErrPlanGenerationInvalid)
			}
			if prior.ManagedPlan {
				digest, err := hex.DecodeString(result.ManagedPlanHash)
				if err != nil || len(digest) != 32 {
					return fmt.Errorf("%w: successful managed plan has no valid saved artifact digest", ErrPlanGenerationInvalid)
				}
			}
		}
	} else if result.Command == command.Import || result.Command == command.State || result.Command == command.Unlock {
		// Remediation and cleanup may invalidate a failed generation too, but
		// must not race a still-executing plan under that same identity.
		if prior.PlanGenerationActive {
			return ErrPlanGenerationSuperseded
		}
	} else if prior.AcceptedPlanGeneration != result.PlanGeneration {
		return ErrPlanGenerationSuperseded
	}
	return nil
}

// LegacyPullStatus can retain main's schema-upgrade recovery only when a
// projection proves there is no generation to erase. Syntactically unreadable
// data requires an explicit fail-closed Begin instead of an ordinary writer.
func LegacyPullStatus(data []byte) bool {
	var projection struct {
		Projects []struct {
			ApplyExecutionID       string
			PlanGeneration         string
			AcceptedPlanGeneration string
			ManagedPlanHash        string
			PlanGenerationActive   bool
		}
	}
	if json.Unmarshal(data, &projection) != nil {
		return false
	}
	for _, project := range projection.Projects {
		if project.PlanGeneration != "" || project.AcceptedPlanGeneration != "" || project.ManagedPlanHash != "" || project.PlanGenerationActive || project.ApplyExecutionID != "" {
			return false
		}
	}
	return true
}

// UpdateLegacyProjectStatus cannot mutate a generation-backed project without
// an observed identity. UI discard uses DiscardPlanStatus instead.
func UpdateLegacyProjectStatus(current *models.PullStatus, pull models.PullRequest, workspace, dir string, status models.ProjectPlanStatus) (models.PullStatus, error) {
	if current == nil {
		return models.PullStatus{}, ErrPlanStatusNotFound
	}
	next := *current
	next.Projects = slices.Clone(current.Projects)
	for i := range next.Projects {
		project := &next.Projects[i]
		if project.Workspace != workspace || project.RepoRelDir != dir {
			continue
		}
		if project.PlanGeneration != "" || project.ApplyExecutionID != "" {
			return *current, ErrPlanGenerationSuperseded
		}
		if project.Status != models.AppliedPlanStatus || status != models.DiscardedPlanStatus {
			project.Status = status
		}
		break
	}
	return next, nil
}

// DiscardPlanStatus requires the exact durable project and identity observed
// by the caller. The boolean distinguishes a discard from unlocking an already
// applied plan; no real ProjectPlanStatus value is used as a missing sentinel.
func DiscardPlanStatus(current *models.PullStatus, pull models.PullRequest, expected models.ProjectStatus) (models.PullStatus, bool, error) {
	if current == nil || !samePullIdentity(current.Pull, pull) {
		return models.PullStatus{}, false, ErrPlanGenerationSuperseded
	}
	next := *current
	next.Projects = slices.Clone(current.Projects)
	project := findProject(next.Projects, projectIdentity{expected.Workspace, expected.RepoRelDir, expected.ProjectName})
	if project == nil || project.PlanGeneration != expected.PlanGeneration || project.AcceptedPlanGeneration != expected.AcceptedPlanGeneration || project.ManagedPlanHash != expected.ManagedPlanHash || project.Status != expected.Status || project.PlanGenerationActive != expected.PlanGenerationActive || project.ManagedPlan != expected.ManagedPlan || project.ApplyExecutionID != expected.ApplyExecutionID {
		return *current, false, ErrPlanGenerationSuperseded
	}
	if project.Status == models.AppliedPlanStatus {
		return next, false, nil
	}
	project.Status = models.DiscardedPlanStatus
	project.PlanGenerationActive = false
	project.AcceptedPlanGeneration = ""
	project.ManagedPlanHash = ""
	project.ApplyExecutionID = ""
	return next, true, nil
}

// BeginApplyExecution consumes the right to start this plan before Terraform
// runs. A crash requires reconciliation and exact discard before replanning,
// not automatic replay of an operation that may have changed infrastructure.
func BeginApplyExecution(current *models.PullStatus, pull models.PullRequest, projects []command.ProjectContext, executionID string) (models.PullStatus, error) {
	if executionID == "" || current == nil || !samePullIdentity(current.Pull, pull) {
		return models.PullStatus{}, ErrPlanGenerationSuperseded
	}
	next := *current
	next.Projects = slices.Clone(current.Projects)
	for _, ctx := range projects {
		project := findProject(next.Projects, projectIdentity{ctx.Workspace, ctx.RepoRelDir, ctx.ProjectName})
		if project == nil || project.PlanGenerationActive || project.PlanGeneration != ctx.PlanGeneration {
			return models.PullStatus{}, ErrPlanGenerationSuperseded
		}
		if project.ApplyExecutionID != "" {
			return models.PullStatus{}, ErrApplyAlreadyStarted
		}
		if project.PlanGeneration != "" {
			if project.AcceptedPlanGeneration != project.PlanGeneration || project.AcceptedPlanGeneration != ctx.AcceptedPlanGeneration {
				return models.PullStatus{}, ErrPlanGenerationSuperseded
			}
			managed := project.ManagedPlan || ctx.RequiresAtlantisManagedPlanFile || slices.ContainsFunc(ctx.Steps, func(step valid.Step) bool { return step.StepName == "apply" })
			if managed {
				digest, err := hex.DecodeString(project.ManagedPlanHash)
				if err != nil || len(digest) != 32 || project.ManagedPlanHash != ctx.ExpectedPlanHash {
					return models.PullStatus{}, ErrPlanGenerationSuperseded
				}
			}
		}
		switch project.Status {
		case models.PlannedPlanStatus, models.PlannedNoChangesPlanStatus, models.PassedPolicyCheckStatus, models.ErroredApplyStatus:
		default:
			return models.PullStatus{}, ErrPlanGenerationSuperseded
		}
		project.ApplyExecutionID = executionID
	}
	return next, nil
}
