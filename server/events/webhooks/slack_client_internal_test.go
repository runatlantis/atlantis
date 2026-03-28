package webhooks

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/slack-go/slack"

	. "github.com/runatlantis/atlantis/testing"
)

func TestCreateAttachments_HeadBranch(t *testing.T) {
	t.Log("The Branch field should use the PR head branch")
	client := DefaultSlackClient{}
	result := ApplyResult{
		Workspace: "default",
		Repo:      models.Repo{FullName: "runatlantis/atlantis"},
		Pull: models.PullRequest{
			Num:        1,
			URL:        "https://github.com/runatlantis/atlantis/pull/1",
			BaseBranch: "main",
			HeadBranch: "feature-branch",
		},
		User:    models.User{Username: "testuser"},
		Success: true,
	}

	attachments := client.createAttachments(result)
	Equals(t, 1, len(attachments))

	fields := attachments[0].Fields
	Equals(t, 4, len(fields))

	branchField := fields[1]
	Equals(t, slack.AttachmentField{
		Title: "Branch",
		Value: "feature-branch",
		Short: true,
	}, branchField)
}
