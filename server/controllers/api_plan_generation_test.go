// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/controllers"
	"github.com/runatlantis/atlantis/server/core/boltdb"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestAPIController_PRApplyAfterDurablePlan(t *testing.T) {
	for _, outcome := range []string{"success", "persistence failure", "missing saved hash"} {
		t.Run(outcome, func(t *testing.T) {

			ac, builder, runner := setup(t)
			storage, err := boltdb.New(t.TempDir())
			Ok(t, err)
			t.Cleanup(func() { Ok(t, storage.Close()) })
			pull := models.PullRequest{Num: 42, HeadBranch: "current-head", HeadCommit: "current-head", BaseBranch: "main", BaseRepo: models.Repo{FullName: "owner/repo", VCSHost: models.VCSHost{Hostname: "gitlab.com", Type: models.Gitlab}}}
			project := command.ProjectContext{Workspace: events.DefaultWorkspace, RepoRelDir: ".", ProjectName: "app", RequiresAtlantisManagedPlanFile: true}
			_, err = storage.BeginPlanGeneration(pull, "previous-plan", []command.ProjectContext{project}, true, command.NoClaim{})
			Ok(t, err)
			_, err = storage.UpdatePullWithResults(pull, []command.ProjectResult{{Command: command.Plan, PlanGeneration: "previous-plan", ManagedPlanHash: strings.Repeat("a", 64), Workspace: project.Workspace, RepoRelDir: project.RepoRelDir, ProjectName: project.ProjectName, ProjectCommandOutput: command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}}}}, command.NoClaim{})
			Ok(t, err)
			ac.PullStatusFetcher = storage
			ac.PlanGenerationDB = storage
			reaper := &recordingAPIPlanReaper{}
			ac.PlanReaper = reaper
			if outcome == "persistence failure" {
				ac.PlanGenerationDB = rejectingAPIResultDatabase{Database: storage}
			}
			build := func(name command.Name) func([]Param) ReturnValues {
				return func(args []Param) ReturnValues {
					ctx := args[0].(*command.Context)
					cmd := project
					cmd.CommandName, cmd.Pull, cmd.PullStatus, cmd.Log, cmd.API = name, ctx.Pull, ctx.PullStatus, ctx.Log, true
					if ctx.PullStatus != nil {
						status := ctx.PullStatus.Projects[0]
						cmd.PlanGeneration, cmd.AcceptedPlanGeneration, cmd.ExpectedPlanHash = status.PlanGeneration, status.AcceptedPlanGeneration, status.ManagedPlanHash
					}
					return ReturnValues{[]command.ProjectContext{cmd}, nil}
				}
			}
			When(builder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).Then(build(command.Plan))
			When(builder.BuildApplyCommands(Any[*command.Context](), Any[*events.CommentCommand]())).Then(build(command.Apply))
			When(runner.Plan(Any[command.ProjectContext]())).Then(func(args []Param) ReturnValues {
				ctx := args[0].(command.ProjectContext)
				if ctx.SavedPlanHash != nil && outcome != "missing saved hash" {
					*ctx.SavedPlanHash = strings.Repeat("b", 64)
				}
				return ReturnValues{command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}}}
			})
			validator := &events.DefaultApplyPlanValidator{PullStatusFetcher: storage}
			applied := false
			When(runner.Apply(Any[command.ProjectContext]())).Then(func(args []Param) ReturnValues {
				applied = true
				ctx := args[0].(command.ProjectContext)
				Assert(t, ctx.ApplyExecutionID != "", "API apply must own a durable execution reservation")
				_, admissionErr := storage.BeginApplyExecution(ctx.Pull, []command.ProjectContext{ctx}, "another-api-worker", command.NoClaim{})
				Assert(t, errors.Is(admissionErr, db.ErrApplyAlreadyStarted), "concurrent API apply must be rejected")
				if ctx.ExpectedPlanHash != strings.Repeat("b", 64) || ctx.PlanGeneration == "previous-plan" {
					return ReturnValues{command.ProjectCommandOutput{Error: fmt.Errorf("API apply selected the previous PR plan instead of the plan it just created")}}
				}
				return ReturnValues{command.ProjectCommandOutput{Error: validator.ValidateProjectPlanStatus(ctx), ApplySuccess: "success"}}
			})
			body, err := json.Marshal(controllers.APIRequest{Repository: "owner/repo", Ref: "current-head", BaseBranch: "main", Type: "Gitlab", PR: 42, Projects: []string{"app"}})
			Ok(t, err)
			request := httptest.NewRequest(http.MethodPost, "/api/apply", bytes.NewReader(body))
			request.Header.Set(atlantisTokenHeader, atlantisToken)
			response := httptest.NewRecorder()
			ac.Apply(response, request)
			if outcome == "success" {
				ResponseContains(t, response, http.StatusOK, "")
				Equals(t, 1, len(reaper.projects))
				Equals(t, "previous-plan", reaper.projects[0].AcceptedPlanGeneration)
				Equals(t, strings.Repeat("a", 64), reaper.projects[0].ManagedPlanHash)
				Assert(t, applied, "apply must consume the newly accepted API plan")
				current, err := storage.GetPullStatus(pull)
				Ok(t, err)
				Equals(t, models.AppliedPlanStatus, current.Projects[0].Status)
				Equals(t, strings.Repeat("b", 64), current.Projects[0].ManagedPlanHash)
				Assert(t, current.Projects[0].AcceptedPlanGeneration != "previous-plan", "API must own a new generation")
			} else {
				ResponseContains(t, response, http.StatusInternalServerError, "")
				Assert(t, !applied, "failed plan persistence must prevent apply")
				current, err := storage.GetPullStatus(pull)
				Ok(t, err)
				Equals(t, "", current.Projects[0].AcceptedPlanGeneration)
				Equals(t, models.ErroredPlanStatus, current.Projects[0].Status)
			}

		})
	}
}

type rejectingAPIResultDatabase struct{ db.Database }

func (r rejectingAPIResultDatabase) UpdatePullWithResults(models.PullRequest, []command.ProjectResult, command.PublicationWriteMode) (models.PullStatus, error) {
	return models.PullStatus{}, errors.New("database write unavailable")
}

type recordingAPIPlanReaper struct{ projects []models.ProjectStatus }

func (r *recordingAPIPlanReaper) ReapPlan(_ logging.SimpleLogging, _ models.PullRequest, project models.ProjectStatus) error {
	r.projects = append(r.projects, project)
	return nil
}
