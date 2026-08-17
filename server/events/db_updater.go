// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

type DBUpdater struct {
	Database db.Database
}

func (c *DBUpdater) acquirePlanPublicationClaim(pull models.PullRequest) (string, error) {
	token := uuid.NewString()
	if err := c.Database.AcquirePlanPublicationClaim(pull, token); err != nil {
		return "", err
	}
	return token, nil
}

func (c *DBUpdater) waitForPlanPublicationClaim(pull models.PullRequest) (string, error) {
	token := uuid.NewString()
	for {
		err := c.Database.AcquirePlanPublicationClaim(pull, token)
		if err == nil {
			return token, nil
		}
		if !errors.Is(err, db.ErrPlanPublicationBusy) {
			return "", err
		}
		time.Sleep(25 * time.Millisecond)
	}
}

func (c *DBUpdater) releasePlanPublicationClaim(pull models.PullRequest, token string) error {
	return c.Database.ReleasePlanPublicationClaim(pull, token)
}

func (c *DBUpdater) beginPlanGeneration(pull models.PullRequest, projectCmds []command.ProjectContext, replace bool, claimToken string) (string, db.PlanGenerationBeginResult, error) {
	generation := uuid.NewString()
	projects := make([]models.ProjectStatus, 0, len(projectCmds))
	seen := make(map[applyPlanKey]struct{}, len(projectCmds))
	for _, projectCtx := range projectCmds {
		key := newApplyPlanKey(projectCtx.Workspace, projectCtx.RepoRelDir, projectCtx.ProjectName)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		projects = append(projects, models.ProjectStatus{
			Workspace:   projectCtx.Workspace,
			RepoRelDir:  projectCtx.RepoRelDir,
			ProjectName: projectCtx.ProjectName,
		})
	}
	var result db.PlanGenerationBeginResult
	var err error
	if replace {
		result, err = c.Database.BeginPlanGenerationReplacing(pull, projects, generation, claimToken)
	} else {
		result, err = c.Database.BeginPlanGeneration(pull, projects, generation, claimToken)
	}
	return generation, result, err
}

func (c *DBUpdater) completePlanGeneration(pull models.PullRequest, generation string, results []command.ProjectResult, claimToken string) (models.PullStatus, error) {
	return c.Database.CompletePlanGeneration(pull, generation, results, claimToken)
}

func (c *DBUpdater) updatePolicyResultsForPlanGeneration(ctx *command.Context, pull models.PullRequest, results []command.ProjectResult, claimToken string) (models.PullStatus, error) {
	filtered := filterProjectResultsForDB(ctx, results)
	if len(filtered) == 0 {
		return c.currentPullStatus(pull)
	}
	return c.Database.UpdatePolicyResultsForPlanGeneration(pull, filtered, claimToken)
}

func (c *DBUpdater) updateApplyResultsForPlanGeneration(ctx *command.Context, pull models.PullRequest, results []command.ProjectResult, claimToken string) (models.PullStatus, error) {
	filtered := filterProjectResultsForDB(ctx, results)
	if len(filtered) == 0 {
		return c.currentPullStatus(pull)
	}
	return c.Database.UpdateApplyResultsForPlanGeneration(pull, filtered, claimToken)
}

func (c *DBUpdater) updateDiscardResultsForPlanGeneration(ctx *command.Context, pull models.PullRequest, results []command.ProjectResult, claimToken string) (models.PullStatus, error) {
	filtered := filterProjectResultsForDB(ctx, results)
	if len(filtered) == 0 {
		return c.currentPullStatus(pull)
	}
	return c.Database.UpdateDiscardResultsForPlanGeneration(pull, filtered, claimToken)
}

func (c *DBUpdater) updateDB(ctx *command.Context, pull models.PullRequest, results []command.ProjectResult) (models.PullStatus, error) {
	if len(results) == 0 && pull.HeadCommit != "" {
		pullStatus, err := c.Database.GetPullStatus(pull)
		if err != nil {
			return models.PullStatus{}, err
		}
		if pullStatus != nil && !pullStatusFreshForPull(pull, pullStatus.Pull) {
			ctx.Log.Debug("ignoring empty result from pull head %q base %q because current plan status is for head %q base %q", pull.HeadCommit, pull.BaseBranch, pullStatus.Pull.HeadCommit, pullStatus.Pull.BaseBranch)
			return *pullStatus, nil
		}
	}

	if staleApplyResultForCurrentPull(pull, results) {
		pullStatus, err := c.Database.GetPullStatus(pull)
		if err != nil {
			return models.PullStatus{}, err
		}
		if pullStatus != nil && !pullStatusFreshForPull(pull, pullStatus.Pull) {
			ctx.Log.Debug("ignoring stale apply result from pull head %q base %q because current plan status is for head %q base %q", pull.HeadCommit, pull.BaseBranch, pullStatus.Pull.HeadCommit, pullStatus.Pull.BaseBranch)
			return *pullStatus, nil
		}
	}

	// Filter out results that errored due to the directory not existing. We
	// don't store these in the database because they would never be "apply-able"
	// and so the pull request would always have errors.
	filtered := filterProjectResultsForDB(ctx, results)
	if len(results) > 0 && len(filtered) == 0 {
		return c.currentPullStatus(pull)
	}
	ctx.Log.Debug("updating DB with pull results")
	return c.Database.UpdatePullWithResults(pull, filtered)
}

func filterProjectResultsForDB(ctx *command.Context, results []command.ProjectResult) []command.ProjectResult {
	filtered := make([]command.ProjectResult, 0, len(results))
	for _, result := range results {
		if _, ok := result.Error.(DirNotExistErr); ok {
			ctx.Log.Debug("ignoring error result from project at dir %q workspace %q because it is dir not exist error", result.RepoRelDir, result.Workspace)
			continue
		}
		if errors.Is(result.Error, errStaleCommandHead) {
			ctx.Log.Debug("ignoring stale command-head result from project at dir %q workspace %q project %q", result.RepoRelDir, result.Workspace, result.ProjectName)
			continue
		}
		filtered = append(filtered, result)
	}
	return filtered
}

func successfulDiscardMutationProjectResults(result command.Result) []command.ProjectResult {
	successful := make([]command.ProjectResult, 0, len(result.ProjectResults))
	for _, projectResult := range result.ProjectResults {
		if projectResult.Error != nil || projectResult.Failure != "" {
			continue
		}
		switch projectResult.Command {
		case command.Import:
			if projectResult.ImportSuccess == nil {
				continue
			}
		case command.State:
			if projectResult.StateRmSuccess == nil {
				continue
			}
		default:
			continue
		}
		successful = append(successful, projectResult)
	}
	return successful
}

func discardTargetsInPullStatus(pullStatus *models.PullStatus, results []command.ProjectResult) []command.ProjectResult {
	if pullStatus == nil || len(results) == 0 {
		return nil
	}
	projects := make(map[applyPlanKey]struct{}, len(pullStatus.Projects))
	for _, project := range pullStatus.Projects {
		if !statusAllowedForDiscoveredPlan(project.Status) {
			continue
		}
		projects[newApplyPlanKey(project.Workspace, project.RepoRelDir, project.ProjectName)] = struct{}{}
	}
	filtered := make([]command.ProjectResult, 0, len(results))
	for _, result := range results {
		if _, ok := projects[newApplyPlanKey(result.Workspace, result.RepoRelDir, result.ProjectName)]; ok {
			filtered = append(filtered, result)
		}
	}
	return filtered
}

func activePlanGenerationTarget(pullStatus *models.PullStatus, projectCmds []command.ProjectContext) (models.ProjectStatus, bool) {
	if pullStatus == nil || len(projectCmds) == 0 {
		return models.ProjectStatus{}, false
	}
	targets := make(map[applyPlanKey]struct{}, len(projectCmds))
	for _, projectCmd := range projectCmds {
		targets[newApplyPlanKey(projectCmd.Workspace, projectCmd.RepoRelDir, projectCmd.ProjectName)] = struct{}{}
	}
	for _, project := range pullStatus.Projects {
		if project.PlanGeneration == "" {
			continue
		}
		if _, ok := targets[newApplyPlanKey(project.Workspace, project.RepoRelDir, project.ProjectName)]; ok {
			return project, true
		}
	}
	return models.ProjectStatus{}, false
}

func (c *DBUpdater) currentPullStatus(pull models.PullRequest) (models.PullStatus, error) {
	pullStatus, err := c.Database.GetPullStatus(pull)
	if err != nil {
		return models.PullStatus{}, err
	}
	if pullStatus != nil {
		return *pullStatus, nil
	}
	return models.PullStatus{Pull: pull}, nil
}

func staleApplyResultForCurrentPull(pull models.PullRequest, results []command.ProjectResult) bool {
	if pull.HeadCommit == "" {
		return false
	}
	for _, result := range results {
		if result.Command == command.Apply {
			return true
		}
	}
	return false
}

func (c *DBUpdater) replaceDB(ctx *command.Context, pull models.PullRequest, results []command.ProjectResult) (models.PullStatus, error) {
	ctx.Log.Debug("replacing DB pull results")
	return c.Database.ReplacePullWithResults(pull, results)
}
