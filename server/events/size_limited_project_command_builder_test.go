package events_test

import (
	"testing"

	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

func TestSizeLimitedProjectCommandBuilder_autoplan(t *testing.T) {
	RegisterMockTestingT(t)

	delegate := mocks.NewMockProjectCommandBuilder()

	ctx := &events.CommandContext{}

	project1 := models.ProjectCommandContext{
		ProjectName: "test1",
	}

	project2 := models.ProjectCommandContext{
		ProjectName: "test2",
	}

	expectedResult := []models.ProjectCommandContext{project1, project2}

	t.Run("Limit Defined and Breached", func(t *testing.T) {
		subject := &events.SizeLimitedProjectCommandBuilder{
			Limit:                 1,
			ProjectCommandBuilder: delegate,
		}

		When(delegate.BuildAutoplanCommands(ctx)).ThenReturn(expectedResult, nil)

		_, err := subject.BuildAutoplanCommands(ctx)

		ErrEquals(t, `Number of projects cannot exceed 1.  This can either be caused by:
1) GH failure in recognizing the diff
2) Pull Request batch is too large for the given Atlantis instance

Please break this pull request into smaller batches and try again.`, err)
	})

	t.Run("Limit defined and not breached", func(t *testing.T) {
		subject := &events.SizeLimitedProjectCommandBuilder{
			Limit:                 2,
			ProjectCommandBuilder: delegate,
		}

		When(delegate.BuildAutoplanCommands(ctx)).ThenReturn(expectedResult, nil)

		result, err := subject.BuildAutoplanCommands(ctx)

		Ok(t, err)

		Assert(t, len(result) == len(expectedResult), "size is expected")
	})

	t.Run("Limit not defined", func(t *testing.T) {
		subject := &events.SizeLimitedProjectCommandBuilder{
			Limit:                 events.InfiniteProjectsPerPR,
			ProjectCommandBuilder: delegate,
		}

		When(delegate.BuildAutoplanCommands(ctx)).ThenReturn(expectedResult, nil)

		result, err := subject.BuildAutoplanCommands(ctx)

		Ok(t, err)

		Assert(t, len(result) == len(expectedResult), "size is expected")
	})
}

func TestSizeLimitedProjectCommandBuilder_planComment(t *testing.T) {
	RegisterMockTestingT(t)

	delegate := mocks.NewMockProjectCommandBuilder()

	ctx := &events.CommandContext{}

	comment := &events.CommentCommand{}

	project1 := models.ProjectCommandContext{
		ProjectName: "test1",
	}

	project2 := models.ProjectCommandContext{
		ProjectName: "test2",
	}

	expectedResult := []models.ProjectCommandContext{project1, project2}

	t.Run("Limit Defined and Breached", func(t *testing.T) {
		subject := &events.SizeLimitedProjectCommandBuilder{
			Limit:                 1,
			ProjectCommandBuilder: delegate,
		}

		When(delegate.BuildPlanCommands(ctx, comment)).ThenReturn(expectedResult, nil)

		_, err := subject.BuildPlanCommands(ctx, comment)

		ErrEquals(t, `Number of projects cannot exceed 1.  This can either be caused by:
1) GH failure in recognizing the diff
2) Pull Request batch is too large for the given Atlantis instance

Please break this pull request into smaller batches and try again.`, err)
	})

	t.Run("Limit defined and not breached", func(t *testing.T) {
		subject := &events.SizeLimitedProjectCommandBuilder{
			Limit:                 2,
			ProjectCommandBuilder: delegate,
		}

		When(delegate.BuildPlanCommands(ctx, comment)).ThenReturn(expectedResult, nil)

		result, err := subject.BuildPlanCommands(ctx, comment)

		Ok(t, err)

		Assert(t, len(result) == len(expectedResult), "size is expected")
	})

	t.Run("Limit not defined", func(t *testing.T) {
		subject := &events.SizeLimitedProjectCommandBuilder{
			Limit:                 events.InfiniteProjectsPerPR,
			ProjectCommandBuilder: delegate,
		}

		When(delegate.BuildPlanCommands(ctx, comment)).ThenReturn(expectedResult, nil)

		result, err := subject.BuildPlanCommands(ctx, comment)

		Ok(t, err)

		Assert(t, len(result) == len(expectedResult), "size is expected")
	})
}
