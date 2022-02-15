package events_test

import (
	"testing"

	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events"
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

	commandCtx := &events.CommandContext{
		Log:        logging.NewNoopLogger(t),
		PullStatus: pullStatus,
	}

	expectedApplyCmt := "Apply Comment"
	expectedPlanCmt := "Plan Comment"

	t.Run("with project name defined", func(t *testing.T) {
		When(mockCommentBuilder.BuildPlanComment(projRepoRelDir, projWorkspace, projName, []string{})).ThenReturn(expectedPlanCmt)
		When(mockCommentBuilder.BuildApplyComment(projRepoRelDir, projWorkspace, projName, false)).ThenReturn(expectedApplyCmt)

		pullStatus.Projects = []models.ProjectStatus{
			{
				Status:      models.ErroredPolicyCheckStatus,
				ProjectName: "project1",
				RepoRelDir:  "dir1",
			},
		}

		result := subject.BuildProjectContext(commandCtx, models.PlanCommand, projCfg, []string{}, "some/dir", false, false, false, false, false)

		assert.Equal(t, models.ErroredPolicyCheckStatus, result[0].ProjectPlanStatus)
	})

	t.Run("with no project name defined", func(t *testing.T) {
		projCfg.Name = ""
		When(mockCommentBuilder.BuildPlanComment(projRepoRelDir, projWorkspace, "", []string{})).ThenReturn(expectedPlanCmt)
		When(mockCommentBuilder.BuildApplyComment(projRepoRelDir, projWorkspace, "", false)).ThenReturn(expectedApplyCmt)
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

		result := subject.BuildProjectContext(commandCtx, models.PlanCommand, projCfg, []string{}, "some/dir", false, false, false, false, false)

		assert.Equal(t, models.ErroredPolicyCheckStatus, result[0].ProjectPlanStatus)
	})

	t.Run("when ParallelApply is set to true", func(t *testing.T) {
		projCfg.Name = "Apply Comment"
		When(mockCommentBuilder.BuildPlanComment(projRepoRelDir, projWorkspace, "", []string{})).ThenReturn(expectedPlanCmt)
		When(mockCommentBuilder.BuildApplyComment(projRepoRelDir, projWorkspace, "", false)).ThenReturn(expectedApplyCmt)
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

		result := subject.BuildProjectContext(commandCtx, models.PlanCommand, projCfg, []string{}, "some/dir", false, false, true, false, false)

		assert.True(t, result[0].ParallelApplyEnabled)
		assert.False(t, result[0].ParallelPlanEnabled)
	})
}
