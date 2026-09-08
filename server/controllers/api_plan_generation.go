// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

// beginAPIPlan gives PR-backed API plans their own generation. Synthetic API
// requests keep request-local status and do not enter the PR generation protocol.
func (a *APIController) beginAPIPlan(ctx *command.Context, projects []command.ProjectContext) error {
	if ctx.Pull.Num <= 0 || len(projects) == 0 {
		return nil
	}
	if a.PlanGenerationDB == nil {
		// Legacy embedded controllers may not supply a durable writer, but cannot
		// reuse a generation installed by another command as their new API plan.
		for _, project := range projects {
			if project.PlanGeneration != "" {
				return fmt.Errorf("persisting PR API plan requires a generation database")
			}
		}
		return nil
	}
	generation := uuid.NewString()
	begun, err := a.PlanGenerationDB.BeginPlanGeneration(ctx.Pull, generation, projects, false, ctx.PublicationMode())
	if err != nil {
		return fmt.Errorf("starting API plan generation: %w", err)
	}
	ctx.PullStatus = &begun.PullStatus
	if a.PlanReaper != nil {
		for _, old := range begun.Previous {
			for _, current := range begun.Projects {
				if current.Workspace == old.Workspace && current.RepoRelDir == old.RepoRelDir && current.ProjectName == old.ProjectName && current.PlanGeneration != old.PlanGeneration {
					if err := a.PlanReaper.ReapPlan(ctx.Log, ctx.Pull, old); err != nil {
						return fmt.Errorf("reaping superseded API plan: %w", err)
					}
					break
				}
			}
		}
	}
	for i := range projects {
		project := &projects[i]
		project.PlanGeneration = generation
		project.AcceptedPlanGeneration, project.ExpectedPlanHash = "", ""
		project.SavedPlanHash = new(string)
		project.PullStatus = ctx.PullStatus
	}
	return nil
}

func (a *APIController) persistAPIResults(ctx *command.Context, results []command.ProjectResult) error {
	if ctx.Pull.Num <= 0 || a.PlanGenerationDB == nil || len(results) == 0 {
		return nil
	}
	updater := &events.DBUpdater{Database: a.PlanGenerationDB}
	status, err := updater.UpdateAPIResults(ctx, results)
	if err != nil {
		return fmt.Errorf("persisting API results: %w", err)
	}
	ctx.PullStatus = &status
	return nil
}

func refreshAPIGeneration(projects []command.ProjectContext, status *models.PullStatus) {
	if status == nil {
		return
	}
	for i := range projects {
		project := &projects[i]
		for _, stored := range status.Projects {
			if project.Workspace == stored.Workspace && project.RepoRelDir == stored.RepoRelDir && project.ProjectName == stored.ProjectName {
				project.PullStatus = status
				project.PlanGeneration, project.AcceptedPlanGeneration, project.ExpectedPlanHash = stored.PlanGeneration, stored.AcceptedPlanGeneration, stored.ManagedPlanHash
				project.ProjectPlanStatus = stored.Status
				break
			}
		}
	}
}

// beginAPIApply reserves this project after pre-workflow hooks and before any
// execution. Each API project persists its result before the next one starts.
func (a *APIController) beginAPIApply(ctx *command.Context, project *command.ProjectContext) error {
	if ctx.Pull.Num <= 0 || project.PlanGeneration == "" {
		return nil
	}
	if a.PlanGenerationDB == nil {
		return fmt.Errorf("reserving PR API apply requires a generation database")
	}
	executionID := uuid.NewString()
	status, err := a.PlanGenerationDB.BeginApplyExecution(ctx.Pull, []command.ProjectContext{*project}, executionID, ctx.PublicationMode())
	if err != nil {
		return fmt.Errorf("reserving API apply execution: %w", err)
	}
	ctx.PullStatus = &status
	project.PullStatus = ctx.PullStatus
	project.ApplyExecutionID = executionID
	return nil
}

// runAPIPublication checks the live identity under short ownership. Execution
// happens outside this callback; completion persists before publishing results.
func (a *APIController) runAPIPublication(ctx *command.Context, operation func() error) error {
	return a.Publication.RunCommand(ctx, func() error {
		if ctx.Pull.Num > 0 && a.LivePullHeadFetcher != nil {
			live, err := a.LivePullHeadFetcher.GetLivePullIdentity(command.ProjectContext{Log: ctx.Log, Pull: ctx.Pull, PullStatus: ctx.PullStatus, API: ctx.API})
			if err != nil {
				return fmt.Errorf("refreshing API publication identity: %w", err)
			}
			if live.HeadCommit == "" || live.HeadCommit != ctx.Pull.HeadCommit || live.BaseBranch != ctx.Pull.BaseBranch {
				return fmt.Errorf("%w: API pull identity changed; replan before applying", db.ErrPlanGenerationSuperseded)
			}
		}
		return operation()
	})
}

func publishAPIStatus(ctx *command.Context, remote func() error) error {
	if ctx.TerminalPublisher != nil {
		return ctx.TerminalPublisher.Publish(remote)
	}
	if ctx.PublicationRequired {
		return db.ErrPublicationOwnerLost
	}
	return remote()
}

func (a *APIController) publishAPIEmptyResult(ctx *command.Context) error {
	if a.SilenceVCSStatusNoProjects || ctx.SuppressVCSStatus {
		return nil
	}
	return a.Publication.RunWithObservedStatus(ctx, a.PlanGenerationDB, a.LivePullHeadFetcher, func() error {
		for _, name := range []command.Name{command.Plan, command.PolicyCheck, command.Apply} {
			if err := publishAPIStatus(ctx, func() error {
				return a.CommitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.SuccessCommitStatus, name, models.ProjectCounts{})
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

// cleanupAPIPlanLocks preserves legacy terminal cleanup while refusing to
// remove a newer command's locks after a PR-backed API request is superseded.
func (a *APIController) cleanupAPIPlanLocks(ctx *command.Context) {
	if err := a.Publication.RunWithObservedStatus(ctx, a.PlanGenerationDB, a.LivePullHeadFetcher, func() error {
		_, err := a.Locker.UnlockByPull(ctx.HeadRepo.FullName, ctx.Pull.Num)
		return err
	}); err != nil {
		ctx.Log.Warn("cleaning API plan locks: %s", err)
	}
}
