// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"strings"
	"testing"

	"github.com/rivo/uniseg"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
	"github.com/slack-go/slack"
)

func attachmentField(attachments []slack.Attachment, title string) (slack.AttachmentField, bool) {
	if len(attachments) == 0 {
		return slack.AttachmentField{}, false
	}
	for _, f := range attachments[0].Fields {
		if f.Title == title {
			return f, true
		}
	}
	return slack.AttachmentField{}, false
}

func descriptionField(attachments []slack.Attachment) (slack.AttachmentField, bool) {
	return attachmentField(attachments, "Description")
}

func applyResultWithBody(body string) ApplyResult {
	return ApplyResult{
		Workspace: "default",
		Repo:      models.Repo{FullName: "owner/repo"},
		Pull: models.PullRequest{
			Num:        1,
			URL:        "url",
			BaseBranch: "main",
			HeadBranch: "feature-branch",
			Body:       body,
		},
		User:      models.User{Username: "user"},
		Success:   true,
		Directory: "dir",
	}
}

func TestCreateAttachments_UsesHeadBranch(t *testing.T) {
	c := DefaultSlackClient{}
	attachments := c.createAttachments(applyResultWithBody(""))
	field, ok := attachmentField(attachments, "Branch")
	Assert(t, ok, "expected a Branch field")
	Equals(t, slack.AttachmentField{
		Title: "Branch",
		Value: "feature-branch",
		Short: true,
	}, field)
}

func TestCreateAttachments_NoDescriptionWhenBodyEmpty(t *testing.T) {
	c := DefaultSlackClient{}
	attachments := c.createAttachments(applyResultWithBody(""))
	_, ok := descriptionField(attachments)
	Equals(t, false, ok)
}

func TestCreateAttachments_IncludesDescription(t *testing.T) {
	c := DefaultSlackClient{}
	attachments := c.createAttachments(applyResultWithBody("a pull request description"))
	field, ok := descriptionField(attachments)
	Assert(t, ok, "expected a Description field")
	Equals(t, "a pull request description", field.Value)
	Equals(t, false, field.Short)
}

func TestCreateAttachments_TruncatesLongDescription(t *testing.T) {
	c := DefaultSlackClient{}
	attachments := c.createAttachments(applyResultWithBody(strings.Repeat("a", 1500)))
	field, ok := descriptionField(attachments)
	Assert(t, ok, "expected a Description field")
	// The result is capped at maxDescriptionGraphemeClusters, ellipsis included.
	Equals(t, maxDescriptionGraphemeClusters, uniseg.GraphemeClusterCount(field.Value))
	Assert(t, strings.HasSuffix(field.Value, "…"), "expected truncated body to end with an ellipsis")
}

func TestCreateAttachments_DoesNotTruncateAtLimit(t *testing.T) {
	c := DefaultSlackClient{}
	body := strings.Repeat("a", maxDescriptionGraphemeClusters)
	attachments := c.createAttachments(applyResultWithBody(body))
	field, ok := descriptionField(attachments)
	Assert(t, ok, "expected a Description field")
	Equals(t, body, field.Value)
}

func TestCreateAttachments_TruncatesDescriptionAtGraphemeBoundary(t *testing.T) {
	c := DefaultSlackClient{}
	body := strings.Repeat("a", maxDescriptionGraphemeClusters-2) + "🧑‍💻bc"
	attachments := c.createAttachments(applyResultWithBody(body))
	field, ok := descriptionField(attachments)
	Assert(t, ok, "expected a Description field")
	Equals(t, maxDescriptionGraphemeClusters, uniseg.GraphemeClusterCount(field.Value))
	Assert(t, strings.HasSuffix(field.Value, "🧑‍💻…"), "expected truncation to preserve the final emoji grapheme cluster")
	Assert(t, !strings.Contains(field.Value, "b"), "expected truncation before the next grapheme cluster")
}
