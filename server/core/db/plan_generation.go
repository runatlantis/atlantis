// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"errors"
	"fmt"
	"strings"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

var (
	ErrPlanGenerationSuperseded   = errors.New("plan generation was superseded")
	ErrPlanGenerationPullChanged  = errors.New("plan generation pull identity changed")
	ErrPlanGenerationStateInvalid = errors.New("plan generation state is invalid")
	ErrPlanGenerationIncomplete   = errors.New("plan generation is incomplete")
	ErrPlanPublicationBusy        = errors.New("plan publication claim is busy")
	ErrPlanPublicationNotOwned    = errors.New("plan publication claim is not owned")
)

// IsPlanGenerationObsolete reports whether a command's generation no longer
// owns the current durable pull identity or project generation.
func IsPlanGenerationObsolete(err error) bool {
	return errors.Is(err, ErrPlanGenerationSuperseded) || errors.Is(err, ErrPlanGenerationPullChanged)
}

// PlanPublicationClaimError classifies deterministic claim conflicts separately
// from storage/backend failures.
type PlanPublicationClaimError struct {
	Kind   error
	Detail string
	Cause  error
}

func (e *PlanPublicationClaimError) Error() string {
	if e == nil {
		return ""
	}
	message := "plan publication claim failed"
	if e.Kind != nil {
		message = e.Kind.Error()
	}
	if e.Detail != "" {
		message += ": " + e.Detail
	}
	if e.Cause != nil {
		message += ": " + e.Cause.Error()
	}
	return message
}

func (e *PlanPublicationClaimError) Unwrap() []error {
	if e == nil {
		return nil
	}
	var unwrapped []error
	if e.Kind != nil {
		unwrapped = append(unwrapped, e.Kind)
	}
	if e.Cause != nil {
		unwrapped = append(unwrapped, e.Cause)
	}
	return unwrapped
}

// ValidatePlanPublicationClaim requires exact ownership when a claim exists.
// Empty tokens remain valid only for legacy/internal callers when no claim exists.
func ValidatePlanPublicationClaim(currentToken, providedToken string) error {
	if currentToken == "" && providedToken == "" {
		return nil
	}
	if currentToken == "" {
		return &PlanPublicationClaimError{Kind: ErrPlanPublicationNotOwned, Detail: "no claim exists"}
	}
	if providedToken == "" {
		return &PlanPublicationClaimError{Kind: ErrPlanPublicationNotOwned, Detail: "claim token is empty"}
	}
	if currentToken != providedToken {
		return &PlanPublicationClaimError{Kind: ErrPlanPublicationNotOwned, Detail: "claim token does not match"}
	}
	return nil
}

// PlanPublicationClaimToken normalizes the optional compatibility argument.
// Production publication paths pass exactly one token; zero preserves legacy
// direct calls when no durable claim exists.
func PlanPublicationClaimToken(claimTokens []string) (string, error) {
	if len(claimTokens) == 0 {
		return "", nil
	}
	if len(claimTokens) == 1 {
		return claimTokens[0], nil
	}
	return "", &PlanPublicationClaimError{Kind: ErrPlanPublicationNotOwned, Detail: "multiple claim tokens provided"}
}

// PlanGenerationProject identifies one project involved in a plan generation
// completion conflict.
type PlanGenerationProject struct {
	Workspace         string
	RepoRelDir        string
	ProjectName       string
	CurrentGeneration string
	CurrentStatus     models.ProjectPlanStatus
}

// PlanGenerationBeginResult is the exact state transition committed by Begin.
// Canceled contains prior active-generation projects that became terminal and
// excludes selected projects immediately reactivated under the new generation.
type PlanGenerationBeginResult struct {
	models.PullStatus
	Canceled []PlanGenerationProject
}

// PlanGenerationCompletionError classifies deterministic state conflicts
// separately from storage/backend failures. Kind is one of the exported
// ErrPlanGeneration* sentinels.
type PlanGenerationCompletionError struct {
	Kind       error
	Generation string
	Projects   []PlanGenerationProject
	Detail     string
	Cause      error
}

func (e *PlanGenerationCompletionError) Error() string {
	if e == nil {
		return ""
	}
	message := "plan generation completion failed"
	if e.Kind != nil {
		message = e.Kind.Error()
	}
	if e.Generation != "" {
		message = fmt.Sprintf("plan generation %q: %s", e.Generation, message)
	}
	if e.Detail != "" {
		message += ": " + e.Detail
	}
	if e.Cause != nil {
		message += ": " + e.Cause.Error()
	}
	return message
}

func (e *PlanGenerationCompletionError) Unwrap() []error {
	if e == nil {
		return nil
	}
	var unwrapped []error
	if e.Kind != nil {
		unwrapped = append(unwrapped, e.Kind)
	}
	if e.Cause != nil {
		unwrapped = append(unwrapped, e.Cause)
	}
	return unwrapped
}

// ValidatePlanGenerationCompletion preflights a completion without mutating
// status. Callers must perform this validation and the subsequent write within
// the same transaction or compare-and-swap attempt.
func ValidatePlanGenerationCompletion(status *models.PullStatus, pull models.PullRequest, generation string, results []command.ProjectResult) error {
	if generation == "" {
		return &PlanGenerationCompletionError{
			Kind:   ErrPlanGenerationStateInvalid,
			Detail: "generation is empty",
		}
	}
	if status == nil {
		return &PlanGenerationCompletionError{
			Kind:       ErrPlanGenerationStateInvalid,
			Generation: generation,
			Detail:     "pull status is missing",
		}
	}
	if pullStatusOutdatedForPull(status.Pull, pull) {
		return &PlanGenerationCompletionError{
			Kind:       ErrPlanGenerationPullChanged,
			Generation: generation,
		}
	}

	resultProjects := make(map[planGenerationProjectKey]struct{}, len(results))
	var superseded []PlanGenerationProject
	var invalid []PlanGenerationProject
	for _, result := range results {
		key := newPlanGenerationProjectKey(result.Workspace, result.RepoRelDir, result.ProjectName)
		if _, duplicate := resultProjects[key]; duplicate {
			invalid = append(invalid, planGenerationProjectFromKey(key, "", models.ProjectPlanStatus(0)))
			continue
		}
		resultProjects[key] = struct{}{}
		if result.Command != command.Plan {
			invalid = append(invalid, planGenerationProjectFromKey(key, "", models.ProjectPlanStatus(0)))
			continue
		}

		project := findProjectStatus(status.Projects, key)
		if project == nil {
			invalid = append(invalid, planGenerationProjectFromKey(key, "", models.ProjectPlanStatus(0)))
			continue
		}
		if project.PlanGeneration != generation {
			conflict := planGenerationProjectFromKey(key, project.PlanGeneration, project.Status)
			superseded = append(superseded, conflict)
			continue
		}
		if project.Status != models.ErroredPlanStatus {
			invalid = append(invalid, planGenerationProjectFromKey(key, project.PlanGeneration, project.Status))
			continue
		}
		if _, err := ManagedPlanHashAfterResult(project.ManagedPlanHash, result); err != nil {
			invalid = append(invalid, planGenerationProjectFromKey(key, project.PlanGeneration, project.Status))
		}
	}

	// Supersession takes priority over other state conflicts. Completion remains
	// all-or-nothing, so callers persist none of the results in this case.
	if len(superseded) > 0 {
		return &PlanGenerationCompletionError{
			Kind:       ErrPlanGenerationSuperseded,
			Generation: generation,
			Projects:   superseded,
			Detail:     formatPlanGenerationProjects("no longer current", superseded),
		}
	}
	if len(invalid) > 0 {
		return &PlanGenerationCompletionError{
			Kind:       ErrPlanGenerationStateInvalid,
			Generation: generation,
			Projects:   invalid,
			Detail:     formatPlanGenerationProjects("invalid", invalid),
		}
	}

	var incomplete []PlanGenerationProject
	generationOwned := false
	for _, project := range status.Projects {
		if project.PlanGeneration != generation {
			continue
		}
		generationOwned = true
		key := newPlanGenerationProjectKey(project.Workspace, project.RepoRelDir, project.ProjectName)
		if _, completed := resultProjects[key]; !completed {
			incomplete = append(incomplete, planGenerationProjectFromKey(key, generation, project.Status))
		}
	}
	if len(results) == 0 && !generationOwned {
		return &PlanGenerationCompletionError{
			Kind:       ErrPlanGenerationStateInvalid,
			Generation: generation,
			Detail:     "no project owns generation",
		}
	}
	if len(incomplete) > 0 {
		return &PlanGenerationCompletionError{
			Kind:       ErrPlanGenerationIncomplete,
			Generation: generation,
			Projects:   incomplete,
			Detail:     formatPlanGenerationProjects("missing result", incomplete),
		}
	}
	return nil
}

// ValidatePolicyResultsForPlanGeneration preflights follow-on policy results
// against the completed plan generation consumed by each result.
func ValidatePolicyResultsForPlanGeneration(status *models.PullStatus, pull models.PullRequest, results []command.ProjectResult) error {
	return validateResultsForAcceptedPlanGeneration(status, pull, results, "policy", func(result command.ProjectResult) bool {
		return result.Command == command.PolicyCheck || result.Command == command.ApprovePolicies
	}, func(status models.ProjectPlanStatus) bool {
		return status == models.PlannedPlanStatus ||
			status == models.PlannedNoChangesPlanStatus ||
			status == models.ErroredPolicyCheckStatus ||
			status == models.PassedPolicyCheckStatus ||
			status == models.ErroredApplyStatus
	})
}

// ValidateApplyResultsForPlanGeneration preflights apply results against the
// completed plan generation consumed by each result.
func ValidateApplyResultsForPlanGeneration(status *models.PullStatus, pull models.PullRequest, results []command.ProjectResult) error {
	return validateResultsForAcceptedPlanGeneration(status, pull, results, "apply", func(result command.ProjectResult) bool {
		return result.Command == command.Apply
	}, func(status models.ProjectPlanStatus) bool {
		return status == models.PlannedPlanStatus ||
			status == models.PlannedNoChangesPlanStatus ||
			status == models.PassedPolicyCheckStatus ||
			status == models.ErroredApplyStatus
	})
}

// ValidateDiscardResultsForPlanGeneration preflights successful import/state
// results that consume and invalidate a completed plan generation.
func ValidateDiscardResultsForPlanGeneration(status *models.PullStatus, pull models.PullRequest, results []command.ProjectResult) error {
	return validateResultsForAcceptedPlanGeneration(status, pull, results, "discard", func(result command.ProjectResult) bool {
		if result.Error != nil || result.Failure != "" {
			return false
		}
		switch result.Command {
		case command.Import:
			return result.ImportSuccess != nil
		case command.State:
			return result.StateRmSuccess != nil
		case command.Unlock:
			return true
		default:
			return false
		}
	}, func(status models.ProjectPlanStatus) bool {
		return status == models.PlannedPlanStatus ||
			status == models.PlannedNoChangesPlanStatus ||
			status == models.PassedPolicyCheckStatus ||
			status == models.ErroredPolicyCheckStatus ||
			status == models.ErroredApplyStatus ||
			status == models.AppliedPlanStatus ||
			status == models.DiscardedPlanStatus
	})
}

func validateResultsForAcceptedPlanGeneration(status *models.PullStatus, pull models.PullRequest, results []command.ProjectResult, resultKind string, validCommand func(command.ProjectResult) bool, validStatus func(models.ProjectPlanStatus) bool) error {
	if status == nil {
		return &PlanGenerationCompletionError{
			Kind:   ErrPlanGenerationStateInvalid,
			Detail: "pull status is missing",
		}
	}
	if pullStatusOutdatedForPull(status.Pull, pull) {
		return &PlanGenerationCompletionError{Kind: ErrPlanGenerationPullChanged}
	}
	if len(results) == 0 {
		return &PlanGenerationCompletionError{
			Kind:   ErrPlanGenerationStateInvalid,
			Detail: resultKind + " results are empty",
		}
	}

	seen := make(map[planGenerationProjectKey]struct{}, len(results))
	var superseded []PlanGenerationProject
	var invalid []PlanGenerationProject
	for _, result := range results {
		key := newPlanGenerationProjectKey(result.Workspace, result.RepoRelDir, result.ProjectName)
		if _, duplicate := seen[key]; duplicate {
			invalid = append(invalid, planGenerationProjectFromKey(key, "", models.ProjectPlanStatus(0)))
			continue
		}
		seen[key] = struct{}{}
		project := findProjectStatus(status.Projects, key)
		if project == nil {
			invalid = append(invalid, planGenerationProjectFromKey(key, "", models.ProjectPlanStatus(0)))
			continue
		}
		if !validCommand(result) {
			invalid = append(invalid, planGenerationProjectFromKey(key, project.AcceptedPlanGeneration, project.Status))
			continue
		}
		if project.PlanGeneration != "" || project.AcceptedPlanGeneration != result.AcceptedPlanGeneration || !validStatus(project.Status) {
			superseded = append(superseded, planGenerationProjectFromKey(key, project.AcceptedPlanGeneration, project.Status))
		}
	}
	if len(superseded) > 0 {
		return &PlanGenerationCompletionError{
			Kind:     ErrPlanGenerationSuperseded,
			Projects: superseded,
			Detail:   formatPlanGenerationProjects("no longer accepted", superseded),
		}
	}
	if len(invalid) > 0 {
		return &PlanGenerationCompletionError{
			Kind:     ErrPlanGenerationStateInvalid,
			Projects: invalid,
			Detail:   formatPlanGenerationProjects("invalid", invalid),
		}
	}
	return nil
}

// SupersedePlanGenerations cancels every project still owned by a generation
// replaced by one of selected. Callers then mark selected as active for the new
// generation in the same atomic write.
func SupersedePlanGenerations(status *models.PullStatus, selected []models.ProjectStatus, generation string) {
	if status == nil {
		return
	}
	superseded := make(map[string]struct{})
	for _, selectedProject := range selected {
		key := newPlanGenerationProjectKey(selectedProject.Workspace, selectedProject.RepoRelDir, selectedProject.ProjectName)
		current := findProjectStatus(status.Projects, key)
		if current == nil || current.PlanGeneration == "" || current.PlanGeneration == generation {
			continue
		}
		superseded[current.PlanGeneration] = struct{}{}
	}
	for i := range status.Projects {
		project := &status.Projects[i]
		if _, ok := superseded[project.PlanGeneration]; !ok {
			continue
		}
		project.Status = models.ErroredPlanStatus
		project.PlanGeneration = ""
		project.ManagedPlanHash = ""
		project.AcceptedPlanGeneration = ""
	}
}

// CanceledPlanGenerationProjects returns the active projects that a Begin
// transition makes terminal. Callers must invoke it from the same transaction
// or compare-and-swap attempt that performs the transition.
func CanceledPlanGenerationProjects(status *models.PullStatus, selected []models.ProjectStatus, generation string, replace bool) []PlanGenerationProject {
	if status == nil {
		return nil
	}
	selectedKeys := make(map[planGenerationProjectKey]struct{}, len(selected))
	superseded := make(map[string]struct{})
	for _, selectedProject := range selected {
		key := newPlanGenerationProjectKey(selectedProject.Workspace, selectedProject.RepoRelDir, selectedProject.ProjectName)
		selectedKeys[key] = struct{}{}
		if replace {
			continue
		}
		current := findProjectStatus(status.Projects, key)
		if current != nil && current.PlanGeneration != "" && current.PlanGeneration != generation {
			superseded[current.PlanGeneration] = struct{}{}
		}
	}

	var canceled []PlanGenerationProject
	for _, project := range status.Projects {
		if project.PlanGeneration == "" || project.PlanGeneration == generation {
			continue
		}
		key := newPlanGenerationProjectKey(project.Workspace, project.RepoRelDir, project.ProjectName)
		if _, reactivated := selectedKeys[key]; reactivated {
			continue
		}
		if !replace {
			if _, ok := superseded[project.PlanGeneration]; !ok {
				continue
			}
		}
		canceled = append(canceled, planGenerationProjectFromKey(key, project.PlanGeneration, project.Status))
	}
	return canceled
}

// ManagedPlanHashAfterResult returns the durable convention-plan authority
// after result is applied to a project status.
func ManagedPlanHashAfterResult(current string, result command.ProjectResult) (string, error) {
	switch result.Command {
	case command.Plan:
		if result.Error != nil || result.Failure != "" {
			return "", nil
		}
		if result.AtlantisManagedPlan {
			if result.ManagedPlanHash == "" {
				return "", errors.New("successful managed plan result has an empty hash")
			}
			return result.ManagedPlanHash, nil
		}
		return "", nil
	case command.PolicyCheck, command.ApprovePolicies:
		return current, nil
	case command.Apply:
		if result.Error != nil || result.Failure != "" {
			return current, nil
		}
		return "", nil
	case command.Import, command.State:
		return "", nil
	default:
		return current, nil
	}
}

// AcceptedPlanGenerationAfterResult returns the durable completed-generation
// identity after a non-generation-fenced result is applied.
func AcceptedPlanGenerationAfterResult(current string, result command.ProjectResult) string {
	switch result.Command {
	case command.Plan, command.Import, command.State:
		return ""
	case command.PolicyCheck, command.ApprovePolicies:
		return current
	case command.Apply:
		if result.Error != nil || result.Failure != "" {
			return current
		}
		return ""
	default:
		return current
	}
}

type planGenerationProjectKey struct {
	workspace   string
	repoRelDir  string
	projectName string
}

func newPlanGenerationProjectKey(workspace, repoRelDir, projectName string) planGenerationProjectKey {
	return planGenerationProjectKey{workspace: workspace, repoRelDir: repoRelDir, projectName: projectName}
}

func findProjectStatus(projects []models.ProjectStatus, key planGenerationProjectKey) *models.ProjectStatus {
	for i := range projects {
		project := &projects[i]
		if project.Workspace == key.workspace && project.RepoRelDir == key.repoRelDir && project.ProjectName == key.projectName {
			return project
		}
	}
	return nil
}

func planGenerationProjectFromKey(key planGenerationProjectKey, currentGeneration string, currentStatus models.ProjectPlanStatus) PlanGenerationProject {
	return PlanGenerationProject{
		Workspace:         key.workspace,
		RepoRelDir:        key.repoRelDir,
		ProjectName:       key.projectName,
		CurrentGeneration: currentGeneration,
		CurrentStatus:     currentStatus,
	}
}

func formatPlanGenerationProjects(prefix string, projects []PlanGenerationProject) string {
	formatted := make([]string, 0, len(projects))
	for _, project := range projects {
		formatted = append(formatted, fmt.Sprintf("%s for dir %q workspace %q project %q", prefix, project.RepoRelDir, project.Workspace, project.ProjectName))
	}
	return strings.Join(formatted, "; ")
}

func pullStatusOutdatedForPull(statusPull models.PullRequest, pull models.PullRequest) bool {
	if statusPull.HeadCommit != pull.HeadCommit {
		return true
	}
	if pull.BaseBranch == "" {
		return false
	}
	return statusPull.BaseBranch == "" || statusPull.BaseBranch != pull.BaseBranch
}
