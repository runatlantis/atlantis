package events_test

import (
	"testing"

	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/mocks/matchers"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/models/fixtures"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
)

func TestApproveCommandRunner_IsOwner(t *testing.T) {
	RegisterMockTestingT(t)

	cases := []struct {
		Description string
		OwnerUsers  []string
		ExpComment  string
	}{
		{
			Description: "When user is not an owner, approval fails",
			OwnerUsers:  []string{},
			ExpComment:  "**Approve Policies Error**\n```\ncontact policy owners to approve failing policies\n```\n",
		},
		{
			Description: "When user is an owner, approval succeeds",
			OwnerUsers:  []string{fixtures.User.Username},
			ExpComment:  "Approved Policies for 1 projects:\n\n1. dir: `` workspace: ``\n\n\n",
		},
	}

	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			vcsClient := setup(t)

			scopeNull, _, _ := metrics.NewLoggingScope(logger, "atlantis")

			modelPull := models.PullRequest{BaseRepo: fixtures.GithubRepo, State: models.OpenPullState, Num: fixtures.Pull.Num}

			ctx := &command.Context{
				User:     fixtures.User,
				Log:      logging.NewNoopLogger(t),
				Scope:    scopeNull,
				Pull:     modelPull,
				HeadRepo: fixtures.GithubRepo,
				Trigger:  command.CommentTrigger,
			}

			When(projectCommandBuilder.BuildApprovePoliciesCommands(matchers.AnyPtrToCommandContext(), matchers.AnyPtrToEventsCommentCommand())).ThenReturn([]command.ProjectContext{
				{
					CommandName: command.ApprovePolicies,
					PolicySets: valid.PolicySets{
						Owners: valid.PolicyOwners{
							Users: c.OwnerUsers,
						},
					},
				},
			}, nil)

			approvePoliciesCommandRunner.Run(ctx, &events.CommentCommand{Name: command.ApprovePolicies})

			vcsClient.VerifyWasCalledOnce().CreateComment(fixtures.GithubRepo, modelPull.Num, c.ExpComment, "approve_policies")
		})
	}
}
