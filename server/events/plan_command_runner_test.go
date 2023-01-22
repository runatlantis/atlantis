package events_test

import (
	"testing"

	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/mocks/matchers"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/models/testdata"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
)

func TestPlanCommandRunner_IsSilenced(t *testing.T) {
	RegisterMockTestingT(t)

	cases := []struct {
		Description      string
		Matched          bool
		Targeted         bool
		VCSStatusSilence bool
		ExpVCSZeroed     bool
		ExpSilenced      bool
	}{
		{
			Description:  "When planning, don't comment but set the 0/0 VCS status",
			ExpVCSZeroed: true,
			ExpSilenced:  true,
		},
		{
			Description:  "When planning with unmatched target, don't comment or set the 0/0 VCS status",
			Targeted:     true,
			ExpVCSZeroed: false,
			ExpSilenced:  true,
		},
		{
			Description:      "When planning with silenced VCS status, don't do anything",
			VCSStatusSilence: true,
			ExpVCSZeroed:     false,
			ExpSilenced:      true,
		},
		{
			Description:  "When planning with matching projects, comment as usual",
			Matched:      true,
			ExpVCSZeroed: false,
			ExpSilenced:  false,
		},
	}

	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			vcsClient := setup(t, func(tc *TestConfig) {
				tc.SilenceNoProjects = true
				tc.silenceVCSStatusNoProjects = c.VCSStatusSilence
			})

			scopeNull, _, _ := metrics.NewLoggingScope(logger, "atlantis")
			modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}

			cmd := &events.CommentCommand{Name: command.Plan}
			if c.Targeted {
				cmd.RepoRelDir = "mydir"
			}

			ctx := &command.Context{
				User:     testdata.User,
				Log:      logging.NewNoopLogger(t),
				Scope:    scopeNull,
				Pull:     modelPull,
				HeadRepo: testdata.GithubRepo,
				Trigger:  command.CommentTrigger,
			}

			When(projectCommandBuilder.BuildPlanCommands(ctx, cmd)).Then(func(args []Param) ReturnValues {
				if c.Matched {
					return ReturnValues{[]command.ProjectContext{{CommandName: command.Plan}}, nil}
				}
				return ReturnValues{[]command.ProjectContext{}, nil}
			})

			planCommandRunner.Run(ctx, cmd)

			timesComment, timesVCS := 1, 0
			if c.ExpSilenced {
				timesComment = 0
			}
			if c.ExpVCSZeroed {
				timesVCS = 1
			}

			vcsClient.VerifyWasCalled(Times(timesComment)).CreateComment(AnyRepo(), AnyInt(), AnyString(), AnyString())
			commitUpdater.VerifyWasCalled(Times(timesVCS)).UpdateCombinedCount(
				matchers.AnyModelsRepo(),
				matchers.AnyModelsPullRequest(),
				matchers.EqModelsCommitStatus(models.SuccessCommitStatus),
				matchers.EqCommandName(command.Plan),
				EqInt(0),
				EqInt(0),
			)
		})
	}
}
