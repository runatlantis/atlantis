// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package webhooks

import (
	"fmt"
	"strings"

	"github.com/rivo/uniseg"
	"github.com/slack-go/slack"
)

const (
	slackSuccessColour = "good"
	slackFailureColour = "danger"
	// maxDescriptionGraphemeClusters is a readability cap for the pull request
	// description, counted in user-visible grapheme clusters. It includes the
	// trailing ellipsis when the description is truncated.
	maxDescriptionGraphemeClusters = 1000
)

//go:generate go tool pegomock generate --package mocks -o mocks/mock_slack_client.go SlackClient

// SlackClient handles making API calls to Slack.
type SlackClient interface {
	AuthTest() error
	TokenIsSet() bool
	PostMessage(channel string, eventResult EventResult) error
}

//go:generate go tool pegomock generate --package mocks -o mocks/mock_underlying_slack_client.go UnderlyingSlackClient

// UnderlyingSlackClient wraps the nlopes/slack.Client implementation so
// we can mock it during tests.
type UnderlyingSlackClient interface {
	AuthTest() (response *slack.AuthTestResponse, error error)
	GetConversations(conversationParams *slack.GetConversationsParameters) (channels []slack.Channel, nextCursor string, err error)
	PostMessage(channelID string, options ...slack.MsgOption) (string, string, error)
}

type DefaultSlackClient struct {
	Slack UnderlyingSlackClient
	Token string
}

func NewSlackClient(token string) SlackClient {
	return &DefaultSlackClient{
		Slack: slack.New(token),
		Token: token,
	}
}

func (d *DefaultSlackClient) AuthTest() error {
	_, err := d.Slack.AuthTest()
	return err
}

func (d *DefaultSlackClient) TokenIsSet() bool {
	return d.Token != ""
}

func (d *DefaultSlackClient) PostMessage(channel string, eventResult EventResult) error {
	attachments := d.createAttachments(eventResult)
	_, _, err := d.Slack.PostMessage(
		channel,
		slack.MsgOptionAsUser(true),
		slack.MsgOptionText("", false),
		slack.MsgOptionAttachments(attachments[0]),
	)
	return err
}

func (d *DefaultSlackClient) createAttachments(eventResult EventResult) []slack.Attachment {
	var colour string
	var successWord string
	if eventResult.Success {
		colour = slackSuccessColour
		successWord = "succeeded"
	} else {
		colour = slackFailureColour
		successWord = "failed"
	}

	eventName := string(eventResult.Event)
	if eventName == "" {
		eventName = "event"
	}
	eventName = strings.ToUpper(eventName[:1]) + eventName[1:]
	text := fmt.Sprintf("%s %s for <%s|%s>", eventName, successWord, eventResult.Pull.URL, eventResult.Repo.FullName)

	directory := eventResult.Directory
	// Since "." looks weird, replace it with "/" to make it clear this is the root.
	if directory == "." {
		directory = "/"
	}

	attachment := slack.Attachment{
		Color: colour,
		Text:  text,
		Fields: []slack.AttachmentField{
			{
				Title: "Workspace",
				Value: eventResult.Workspace,
				Short: true,
			},
			{
				Title: "Branch",
				Value: eventResult.Pull.BaseBranch,
				Short: true,
			},
			{
				Title: "User",
				Value: eventResult.User.Username,
				Short: true,
			},
			{
				Title: "Directory",
				Value: directory,
				Short: true,
			},
		},
	}

	// Include the pull request description when present, rendered as a
	// full-width field and truncated so a long description can't exceed Slack's
	// message size limits.
	if applyResult.Pull.Body != "" {
		attachment.Fields = append(attachment.Fields, slack.AttachmentField{
			Title: "Description",
			Value: truncateGraphemeClusters(applyResult.Pull.Body, maxDescriptionGraphemeClusters),
			Short: false,
		})
	}

	return []slack.Attachment{attachment}
}

// truncateGraphemeClusters returns s unchanged when it has at most max
// grapheme clusters. Otherwise it is cut at a cluster boundary to max-1
// clusters with an ellipsis appended. It walks the string without allocating a
// slice for the whole input.
func truncateGraphemeClusters(s string, max int) string {
	if max <= 0 {
		return ""
	}

	count := 0
	keepUntil := 0
	offset := 0
	state := -1
	rest := s
	for len(rest) > 0 {
		cluster, remaining, _, nextState := uniseg.FirstGraphemeClusterInString(rest, state)
		if count == max-1 {
			keepUntil = offset
		}
		if count == max {
			return s[:keepUntil] + "…"
		}
		offset += len(cluster)
		rest = remaining
		state = nextState
		count++
	}
	return s
}
