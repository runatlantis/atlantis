// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"fmt"

	"github.com/google/uuid"
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
	begun, err := a.PlanGenerationDB.BeginPlanGeneration(ctx.Pull, generation, projects, false)
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
