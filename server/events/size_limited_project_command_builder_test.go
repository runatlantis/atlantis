package events_test

import (
	"testing"

	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	. "github.com/runatlantis/atlantis/testing"
)

func TestSizeLimitedProjectCommandBuilder_autoplan(t *testing.T) {
	RegisterMockTestingT(t)

	ctx := &command.Context{}

	project1 := command.ProjectContext{
		ProjectName: "test1",
		CommandName: command.Plan,
	}

	project2 := command.ProjectContext{
		ProjectName: "test2",
		CommandName: command.Plan,
	}

	project3 := command.ProjectContext{
		ProjectName: "test1",
		CommandName: command.PolicyCheck,
	}

	expectedResult := []command.ProjectContext{project1, project2}
	delegate := mockProjectCommandBuilder{
		commands: expectedResult,
	}

	t.Run("Limit Defined and Breached", func(t *testing.T) {
		subject := &events.SizeLimitedProjectCommandBuilder{
			Limit:                 1,
			ProjectCommandBuilder: delegate,
		}
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
		result, err := subject.BuildAutoplanCommands(ctx)

		Ok(t, err)

		Assert(t, len(result) == len(expectedResult), "size is expected")
	})

	t.Run("Limit not defined", func(t *testing.T) {
		subject := &events.SizeLimitedProjectCommandBuilder{
			Limit:                 events.InfiniteProjectsPerPR,
			ProjectCommandBuilder: delegate,
		}
		result, err := subject.BuildAutoplanCommands(ctx)

		Ok(t, err)

		Assert(t, len(result) == len(expectedResult), "size is expected")
	})

	t.Run("Only plan commands counted in limit", func(t *testing.T) {
		resultWithPolicyCheckCommand := []command.ProjectContext{project1, project2, project3}
		delegate = mockProjectCommandBuilder{
			commands: resultWithPolicyCheckCommand,
		}
		subject := &events.SizeLimitedProjectCommandBuilder{
			Limit:                 2,
			ProjectCommandBuilder: delegate,
		}
		result, err := subject.BuildAutoplanCommands(ctx)

		Ok(t, err)

		Assert(t, len(result) == len(resultWithPolicyCheckCommand), "size is expected")
	})
}

func TestSizeLimitedProjectCommandBuilder_planComment(t *testing.T) {
	RegisterMockTestingT(t)
	ctx := &command.Context{}

	comment := &command.Comment{}

	project1 := command.ProjectContext{
		ProjectName: "test1",
		CommandName: command.Plan,
	}

	project2 := command.ProjectContext{
		ProjectName: "test2",
		CommandName: command.Plan,
	}

	expectedResult := []command.ProjectContext{project1, project2}
	delegate := mockProjectCommandBuilder{
		commands: expectedResult,
	}

	t.Run("Limit Defined and Breached", func(t *testing.T) {
		subject := &events.SizeLimitedProjectCommandBuilder{
			Limit:                 1,
			ProjectCommandBuilder: delegate,
		}
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
		result, err := subject.BuildPlanCommands(ctx, comment)

		Ok(t, err)

		Assert(t, len(result) == len(expectedResult), "size is expected")
	})

	t.Run("Limit not defined", func(t *testing.T) {
		subject := &events.SizeLimitedProjectCommandBuilder{
			Limit:                 events.InfiniteProjectsPerPR,
			ProjectCommandBuilder: delegate,
		}
		result, err := subject.BuildPlanCommands(ctx, comment)

		Ok(t, err)

		Assert(t, len(result) == len(expectedResult), "size is expected")
	})
}

type mockProjectCommandBuilder struct {
	commands []command.ProjectContext
	error    error
}

func (m mockProjectCommandBuilder) BuildAutoplanCommands(ctx *command.Context) ([]command.ProjectContext, error) {
	return m.commands, m.error
}

func (m mockProjectCommandBuilder) BuildPlanCommands(ctx *command.Context, comment *command.Comment) ([]command.ProjectContext, error) {
	return m.commands, m.error
}

func (m mockProjectCommandBuilder) BuildPolicyCheckCommands(ctx *command.Context) ([]command.ProjectContext, error) {
	return m.commands, m.error
}

func (m mockProjectCommandBuilder) BuildApplyCommands(ctx *command.Context, comment *command.Comment) ([]command.ProjectContext, error) {
	return m.commands, m.error
}

func (m mockProjectCommandBuilder) BuildApprovePoliciesCommands(ctx *command.Context, comment *command.Comment) ([]command.ProjectContext, error) {
	return m.commands, m.error
}

func (m mockProjectCommandBuilder) BuildVersionCommands(ctx *command.Context, comment *command.Comment) ([]command.ProjectContext, error) {
	return m.commands, m.error
}
