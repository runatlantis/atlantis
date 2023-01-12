package events_test

import (
	"errors"
	"testing"

	"github.com/google/go-github/v49/github"
	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/mocks/matchers"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/models/testdata"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
)

func TestApplyCommandRunner_IsLocked(t *testing.T) {
	RegisterMockTestingT(t)

	cases := []struct {
		Description    string
		ApplyLocked    bool
		ApplyLockError error
		ExpComment     string
	}{
		{
			Description:    "When global apply lock is present IsDisabled returns true",
			ApplyLocked:    true,
			ApplyLockError: nil,
			ExpComment:     "**Error:** Running `atlantis apply` is disabled.",
		},
		{
			Description:    "When no global apply lock is present and DisableApply flag is false IsDisabled returns false",
			ApplyLocked:    false,
			ApplyLockError: nil,
			ExpComment:     "Ran Apply for 0 projects:",
		},
		{
			Description:    "If ApplyLockChecker returns an error IsDisabled return value of DisableApply flag",
			ApplyLockError: errors.New("error"),
			ApplyLocked:    false,
			ExpComment:     "Ran Apply for 0 projects:",
		},
	}

	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			vcsClient := setup(t)

			scopeNull, _, _ := metrics.NewLoggingScope(logger, "atlantis")

			pull := &github.PullRequest{
				State: github.String("open"),
			}
			modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}
			When(githubGetter.GetPullRequest(testdata.GithubRepo, testdata.Pull.Num)).ThenReturn(pull, nil)
			When(eventParsing.ParseGithubPull(pull)).ThenReturn(modelPull, modelPull.BaseRepo, testdata.GithubRepo, nil)

			ctx := &command.Context{
				User:     testdata.User,
				Log:      logging.NewNoopLogger(t),
				Scope:    scopeNull,
				Pull:     modelPull,
				HeadRepo: testdata.GithubRepo,
				Trigger:  command.CommentTrigger,
			}

			When(applyLockChecker.CheckApplyLock()).ThenReturn(locking.ApplyCommandLock{Locked: c.ApplyLocked}, c.ApplyLockError)
			applyCommandRunner.Run(ctx, &events.CommentCommand{Name: command.Apply})

			vcsClient.VerifyWasCalledOnce().CreateComment(testdata.GithubRepo, modelPull.Num, c.ExpComment, "apply")
		})
	}
}

func TestApplyCommandRunner_IsSilenced(t *testing.T) {
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
			Description:  "When applying, don't comment but set the 0/0 VCS status",
			ExpVCSZeroed: true,
			ExpSilenced:  true,
		},
		{
			Description:  "When applying with unmatched target, don't comment or set the 0/0 VCS status",
			Targeted:     true,
			ExpVCSZeroed: false,
			ExpSilenced:  true,
		},
		{
			Description:      "When applying with silenced VCS status, don't do anything",
			VCSStatusSilence: true,
			ExpVCSZeroed:     false,
			ExpSilenced:      true,
		},
		{
			Description:  "When applying with matching projects, comment as usual",
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

			pull := &github.PullRequest{
				State: github.String("open"),
			}
			modelPull := models.PullRequest{BaseRepo: fixtures.GithubRepo, State: models.OpenPullState, Num: fixtures.Pull.Num}
			When(githubGetter.GetPullRequest(fixtures.GithubRepo, fixtures.Pull.Num)).ThenReturn(pull, nil)
			When(eventParsing.ParseGithubPull(pull)).ThenReturn(modelPull, modelPull.BaseRepo, fixtures.GithubRepo, nil)

			cmd := &events.CommentCommand{Name: command.Apply}
			if c.Targeted {
				cmd.RepoRelDir = "mydir"
			}

			ctx := &command.Context{
				User:     fixtures.User,
				Log:      logging.NewNoopLogger(t),
				Scope:    scopeNull,
				Pull:     modelPull,
				HeadRepo: fixtures.GithubRepo,
				Trigger:  command.CommentTrigger,
			}

			When(projectCommandBuilder.BuildApplyCommands(ctx, cmd)).Then(func(args []Param) ReturnValues {
				if c.Matched {
					return ReturnValues{[]command.ProjectContext{{CommandName: command.Apply}}, nil}
				}
				return ReturnValues{[]command.ProjectContext{}, nil}
			})

			applyCommandRunner.Run(ctx, cmd)

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
				matchers.EqCommandName(command.Apply),
				EqInt(0),
				EqInt(0),
			)
		})
	}
}
