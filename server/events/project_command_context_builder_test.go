// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	tfclientmocks "github.com/runatlantis/atlantis/server/core/terraform/tfclient/mocks"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestProjectCommandContextBuilder_PullStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCommentBuilder := mocks.NewMockCommentBuilder(ctrl)
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
		mockCommentBuilder.EXPECT().BuildPlanComment(projRepoRelDir, projWorkspace, projName, []string{}).Return(expectedPlanCmt).AnyTimes()
		mockCommentBuilder.EXPECT().BuildApplyComment(projRepoRelDir, projWorkspace, projName, false, "").Return(expectedApplyCmt).AnyTimes()
		mockCommentBuilder.EXPECT().BuildApprovePoliciesComment(projRepoRelDir, projWorkspace, projName).Return("Approve Policies Comment").AnyTimes()

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
		mockCommentBuilder.EXPECT().BuildPlanComment(projRepoRelDir, projWorkspace, "", []string{}).Return(expectedPlanCmt).AnyTimes()
		mockCommentBuilder.EXPECT().BuildApplyComment(projRepoRelDir, projWorkspace, "", false, "").Return(expectedApplyCmt).AnyTimes()
		mockCommentBuilder.EXPECT().BuildApprovePoliciesComment(projRepoRelDir, projWorkspace, "").Return("Approve Policies Comment").AnyTimes()
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
		mockCommentBuilder.EXPECT().BuildPlanComment(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedPlanCmt).AnyTimes()
		mockCommentBuilder.EXPECT().BuildApplyComment(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedApplyCmt).AnyTimes()
		mockCommentBuilder.EXPECT().BuildApprovePoliciesComment(gomock.Any(), gomock.Any(), gomock.Any()).Return("Approve Policies Comment").AnyTimes()
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
		mockCommentBuilder.EXPECT().BuildPlanComment(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedPlanCmt).AnyTimes()
		mockCommentBuilder.EXPECT().BuildApplyComment(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedApplyCmt).AnyTimes()
		mockCommentBuilder.EXPECT().BuildApprovePoliciesComment(gomock.Any(), gomock.Any(), gomock.Any()).Return("Approve Policies Comment").AnyTimes()
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
