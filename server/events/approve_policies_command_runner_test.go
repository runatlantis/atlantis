package events_test

import (
	"testing"

	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/mocks/matchers"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/models/testdata"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
)

func TestApproveCommandRunner_IsOwner(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)

	cases := []struct {
		Description string
		OwnerUsers  []string
		OwnerTeams  []string // Teams configured as owners
		UserTeams   []string // Teams the user is a member of
		ExpComment  string
	}{
		{
			Description: "When user is not an owner, approval fails",
			OwnerUsers:  []string{},
			OwnerTeams:  []string{},
			ExpComment:  "**Approve Policies Error**\n```\ncontact policy owners to approve failing policies\n```",
		},
		{
			Description: "When user is an owner, approval succeeds",
			OwnerUsers:  []string{testdata.User.Username},
			OwnerTeams:  []string{},
			ExpComment:  "Approved Policies for 1 projects:\n\n1. dir: `` workspace: ``",
		},
		{
			Description: "When user is an owner via team membership, approval succeeds",
			OwnerUsers:  []string{},
			OwnerTeams:  []string{"SomeTeam"},
			UserTeams:   []string{"SomeTeam"},
			ExpComment:  "Approved Policies for 1 projects:\n\n1. dir: `` workspace: ``",
		},
		{
			Description: "When user belongs to a team not configured as a owner, approval fails",
			OwnerUsers:  []string{},
			OwnerTeams:  []string{"SomeTeam"},
			UserTeams:   []string{"SomeOtherTeam}"},
			ExpComment:  "**Approve Policies Error**\n```\ncontact policy owners to approve failing policies\n```",
		},
		{
			Description: "When user is an owner but not a team member, approval succeeds",
			OwnerUsers:  []string{testdata.User.Username},
			OwnerTeams:  []string{"SomeTeam"},
			UserTeams:   []string{"SomeOtherTeam"},
			ExpComment:  "Approved Policies for 1 projects:\n\n1. dir: `` workspace: ``",
		},
	}

	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			vcsClient := setup(t)

			scopeNull, _, _ := metrics.NewLoggingScope(logger, "atlantis")

			modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}

			ctx := &command.Context{
				User:     testdata.User,
				Log:      logging.NewNoopLogger(t),
				Scope:    scopeNull,
				Pull:     modelPull,
				HeadRepo: testdata.GithubRepo,
				Trigger:  command.CommentTrigger,
			}

			When(projectCommandBuilder.BuildApprovePoliciesCommands(matchers.AnyPtrToCommandContext(), matchers.AnyPtrToEventsCommentCommand())).ThenReturn([]command.ProjectContext{
				{
					CommandName: command.ApprovePolicies,
					PolicySets: valid.PolicySets{
						Owners: valid.PolicyOwners{
							Users: c.OwnerUsers,
							Teams: c.OwnerTeams,
						},
					},
				},
			}, nil)
			When(vcsClient.GetTeamNamesForUser(testdata.GithubRepo, testdata.User)).ThenReturn(c.UserTeams, nil)

			approvePoliciesCommandRunner.Run(ctx, &events.CommentCommand{Name: command.ApprovePolicies})

			vcsClient.VerifyWasCalledOnce().CreateComment(testdata.GithubRepo, modelPull.Num, c.ExpComment, "approve_policies")
		})
	}
}
