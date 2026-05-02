// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package webhooks

import (
	"fmt"

	"github.com/slack-go/slack"
)

const (
	slackSuccessColour = "good"
	slackFailureColour = "danger"
)

//go:generate go tool pegomock generate --package mocks -o mocks/mock_slack_client.go SlackClient

// SlackClient handles making API calls to Slack.
type SlackClient interface {
	AuthTest() error
	TokenIsSet() bool
	PostMessage(channel string, applyResult ApplyResult) error
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

func (d *DefaultSlackClient) PostMessage(channel string, applyResult ApplyResult) error {
	attachments := d.createAttachments(applyResult)
	_, _, err := d.Slack.PostMessage(
		channel,
		slack.MsgOptionAsUser(true),
		slack.MsgOptionText("", false),
		slack.MsgOptionAttachments(attachments[0]),
	)
	return err
}

func (d *DefaultSlackClient) createAttachments(applyResult ApplyResult) []slack.Attachment {
	var colour string
	var successWord string
	if applyResult.Success {
		colour = slackSuccessColour
		successWord = "succeeded"
	} else {
		colour = slackFailureColour
		successWord = "failed"
	}

	text := fmt.Sprintf("Apply %s for <%s|%s>", successWord, applyResult.Pull.URL, applyResult.Repo.FullName)
	directory := applyResult.Directory
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
				Value: applyResult.Workspace,
				Short: true,
			},
			{
				Title: "Branch",
				Value: applyResult.Pull.BaseBranch,
				Short: true,
			},
			{
				Title: "User",
				Value: applyResult.User.Username,
				Short: true,
			},
			{
				Title: "Directory",
				Value: directory,
				Short: true,
			},
		},
	}
	return []slack.Attachment{attachment}
}
