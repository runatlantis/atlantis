package webhooks_test

import (
	"regexp"
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/webhooks"
	"github.com/runatlantis/atlantis/server/events/webhooks/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/petergtz/pegomock/v4"
	. "github.com/runatlantis/atlantis/testing"
)

func TestMSTeamsWebhook_Send(t *testing.T) {
	t.Run("should send when workspace and branch match", func(t *testing.T) {
		pegomock.RegisterMockTestingT(t)
		client := mocks.NewMockMSTeamsClient()
		
		workspaceRegex, _ := regexp.Compile(".*")
		branchRegex, _ := regexp.Compile(".*")
		
		webhook, err := webhooks.NewMSTeams(workspaceRegex, branchRegex, "https://example.com/webhook", client)
		Ok(t, err)

		applyResult := webhooks.ApplyResult{
			Workspace: "production",
			Repo: models.Repo{
				FullName: "owner/repo",
			},
			Pull: models.PullRequest{
				BaseBranch: "main",
				Num:        1,
				URL:        "https://github.com/owner/repo/pull/1",
			},
			User: models.User{
				Username: "atlantis",
			},
			Success:     true,
			Directory:   ".",
			ProjectName: "test-project",
		}

		err = webhook.Send(logging.NewNoopLogger(t), applyResult)
		Ok(t, err)

		client.VerifyWasCalledOnce().PostMessage("https://example.com/webhook", applyResult)
	})

	t.Run("should not send when workspace doesn't match", func(t *testing.T) {
		pegomock.RegisterMockTestingT(t)
		client := mocks.NewMockMSTeamsClient()
		
		workspaceRegex, _ := regexp.Compile("staging")
		branchRegex, _ := regexp.Compile(".*")
		
		webhook, err := webhooks.NewMSTeams(workspaceRegex, branchRegex, "https://example.com/webhook", client)
		Ok(t, err)

		applyResult := webhooks.ApplyResult{
			Workspace: "production",
			Pull: models.PullRequest{
				BaseBranch: "main",
			},
		}

		err = webhook.Send(logging.NewNoopLogger(t), applyResult)
		Ok(t, err)

		client.VerifyWasCalled(pegomock.Never()).PostMessage(pegomock.AnyString(), pegomock.Any[webhooks.ApplyResult]())
	})

	t.Run("should not send when branch doesn't match", func(t *testing.T) {
		pegomock.RegisterMockTestingT(t)
		client := mocks.NewMockMSTeamsClient()
		
		workspaceRegex, _ := regexp.Compile(".*")
		branchRegex, _ := regexp.Compile("main")
		
		webhook, err := webhooks.NewMSTeams(workspaceRegex, branchRegex, "https://example.com/webhook", client)
		Ok(t, err)

		applyResult := webhooks.ApplyResult{
			Workspace: "production",
			Pull: models.PullRequest{
				BaseBranch: "feature-branch",
			},
		}

		err = webhook.Send(logging.NewNoopLogger(t), applyResult)
		Ok(t, err)

		client.VerifyWasCalled(pegomock.Never()).PostMessage(pegomock.AnyString(), pegomock.Any[webhooks.ApplyResult]())
	})
}
