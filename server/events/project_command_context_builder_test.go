// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"testing"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	tfclientmocks "github.com/runatlantis/atlantis/server/core/terraform/tfclient/mocks"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/stretchr/testify/assert"
)

func TestProjectCommandContextBuilder_PullStatus(t *testing.T) {

	mockCommentBuilder := mocks.NewMockCommentBuilder()
	subject := events.DefaultProjectCommandContextBuilder{
		CommentBuilder: mockCommentBuilder,
	}

	projRepoRelDir := "dir1"
	projWorkspace := "default"
	projName := "project1"

	projCfg := valid.MergedProjectCfg{
		RepoRelDir: projRepoRelDir,
		Workspace:  projWorkspace,
		Name:       projName,
		Workflow: valid.Workflow{
			Name:  valid.DefaultWorkflowName,
			Apply: valid.DefaultApplyStage,
		},
	}

	pullStatus := &models.PullStatus{
		Projects: []models.ProjectStatus{},
	}

	commandCtx := &command.Context{
		Log:        logging.NewNoopLogger(t),
		PullStatus: pullStatus,
	}

	expectedApplyCmt := "Apply Comment"
	expectedPlanCmt := "Plan Comment"

	terraformClient := tfclientmocks.NewMockClient()

	t.Run("with project name defined", func(t *testing.T) {
		When(mockCommentBuilder.BuildPlanComment(projRepoRelDir, projWorkspace, projName, []string{})).ThenReturn(expectedPlanCmt)
		When(mockCommentBuilder.BuildApplyComment(projRepoRelDir, projWorkspace, projName, false, "")).ThenReturn(expectedApplyCmt)

		pullStatus.Projects = []models.ProjectStatus{
			{
				Status:      models.ErroredPolicyCheckStatus,
				ProjectName: "project1",
				RepoRelDir:  "dir1",
			},
		}

		result := subject.BuildProjectContext(commandCtx, command.Plan, "", projCfg, []string{}, "some/dir", false, false, false, false, false, terraformClient)
		assert.Equal(t, models.ErroredPolicyCheckStatus, result[0].ProjectPlanStatus)
	})

	t.Run("with no project name defined", func(t *testing.T) {
		projCfg.Name = ""
		When(mockCommentBuilder.BuildPlanComment(projRepoRelDir, projWorkspace, "", []string{})).ThenReturn(expectedPlanCmt)
		When(mockCommentBuilder.BuildApplyComment(projRepoRelDir, projWorkspace, "", false, "")).ThenReturn(expectedApplyCmt)
		pullStatus.Projects = []models.ProjectStatus{
			{
				Status:     models.ErroredPlanStatus,
				RepoRelDir: "dir2",
			},
			{
				Status:     models.ErroredPolicyCheckStatus,
				RepoRelDir: "dir1",
			},
		}

		result := subject.BuildProjectContext(commandCtx, command.Plan, "", projCfg, []string{}, "some/dir", false, false, false, false, false, terraformClient)

		assert.Equal(t, models.ErroredPolicyCheckStatus, result[0].ProjectPlanStatus)
	})

	t.Run("when ParallelApply is set to true", func(t *testing.T) {
		projCfg.Name = "Apply Comment"
		When(mockCommentBuilder.BuildPlanComment(projRepoRelDir, projWorkspace, "", []string{})).ThenReturn(expectedPlanCmt)
		When(mockCommentBuilder.BuildApplyComment(projRepoRelDir, projWorkspace, "", false, "")).ThenReturn(expectedApplyCmt)
		pullStatus.Projects = []models.ProjectStatus{
			{
				Status:     models.ErroredPlanStatus,
				RepoRelDir: "dir2",
			},
			{
				Status:     models.ErroredPolicyCheckStatus,
				RepoRelDir: "dir1",
			},
		}

		result := subject.BuildProjectContext(commandCtx, command.Plan, "", projCfg, []string{}, "some/dir", false, true, false, false, false, terraformClient)

		assert.True(t, result[0].ParallelApplyEnabled)
		assert.False(t, result[0].ParallelPlanEnabled)
	})

	t.Run("when AbortOnExecutionOrderFail is set to true", func(t *testing.T) {
		projCfg.Name = "Apply Comment"
		When(mockCommentBuilder.BuildPlanComment(projRepoRelDir, projWorkspace, "", []string{})).ThenReturn(expectedPlanCmt)
		When(mockCommentBuilder.BuildApplyComment(projRepoRelDir, projWorkspace, "", false, "")).ThenReturn(expectedApplyCmt)
		pullStatus.Projects = []models.ProjectStatus{
			{
				Status:     models.ErroredPlanStatus,
				RepoRelDir: "dir2",
			},
			{
				Status:     models.ErroredPolicyCheckStatus,
				RepoRelDir: "dir1",
			},
		}

		result := subject.BuildProjectContext(commandCtx, command.Plan, "", projCfg, []string{}, "some/dir", false, false, false, false, true, terraformClient)

		assert.True(t, result[0].AbortOnExecutionOrderFail)
	})
}

func TestProjectCommandContextBuilder_PropagatesAPIWorkflowFlags(t *testing.T) {
	RegisterMockTestingT(t)
	mockCommentBuilder := mocks.NewMockCommentBuilder()
	subject := events.DefaultProjectCommandContextBuilder{
		CommentBuilder: mockCommentBuilder,
	}
	terraformClient := tfclientmocks.NewMockClient()
	projCfg := valid.MergedProjectCfg{
		RepoRelDir: "env",
		Workspace:  "prod",
		Name:       "app",
		Workflow: valid.Workflow{
			Name: valid.DefaultWorkflowName,
			Plan: valid.DefaultPlanStage,
		},
	}
	When(mockCommentBuilder.BuildPlanComment("env", "prod", "app", []string{})).ThenReturn("plan comment")

	apiCtx := &command.Context{
		Log:                   logging.NewNoopLogger(t),
		API:                   true,
		SkipPRRequirements:    true,
		SkipPRModifiedFiles:   true,
		SuppressVCSStatus:     true,
		SuppressJobOutput:     true,
		SuppressApplyWebhooks: true,
	}
	apiResult := subject.BuildProjectContext(apiCtx, command.Plan, "", projCfg, []string{}, "repo", false, false, false, false, false, terraformClient)
	assert.True(t, apiResult[0].API)
	assert.True(t, apiResult[0].SkipPRRequirements)
	assert.True(t, apiResult[0].SuppressVCSStatus)
	assert.True(t, apiResult[0].SuppressJobOutput)
	assert.True(t, apiResult[0].SuppressApplyWebhooks)

	When(mockCommentBuilder.BuildPlanComment("env", "prod", "app", []string{})).ThenReturn("plan comment")
	normalCtx := &command.Context{Log: logging.NewNoopLogger(t)}
	normalResult := subject.BuildProjectContext(normalCtx, command.Plan, "", projCfg, []string{}, "repo", false, false, false, false, false, terraformClient)
	assert.False(t, normalResult[0].API)
	assert.False(t, normalResult[0].SkipPRRequirements)
	assert.False(t, normalResult[0].SuppressVCSStatus)
	assert.False(t, normalResult[0].SuppressJobOutput)
	assert.False(t, normalResult[0].SuppressApplyWebhooks)
}
