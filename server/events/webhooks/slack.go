// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package webhooks

import (
	"regexp"

	"fmt"

	"github.com/runatlantis/atlantis/server/logging"
)

// SlackWebhook sends webhooks to Slack.
type SlackWebhook struct {
	Client         SlackClient
	WorkspaceRegex *regexp.Regexp
	BranchRegex    *regexp.Regexp
	Channel        string
}

func NewSlack(wr *regexp.Regexp, br *regexp.Regexp, channel string, client SlackClient) (*SlackWebhook, error) {
	if err := client.AuthTest(); err != nil {
		return nil, fmt.Errorf("testing slack authentication: %s. Verify your slack-token is valid", err)
	}

	return &SlackWebhook{
		Client:         client,
		WorkspaceRegex: wr,
		BranchRegex:    br,
		Channel:        channel,
	}, nil
}

// Send sends the webhook to Slack if workspace and branch matches their respective regex.
func (s *SlackWebhook) Send(_ logging.SimpleLogging, eventResult EventResult) error {
	if !s.WorkspaceRegex.MatchString(eventResult.Workspace) || !s.BranchRegex.MatchString(eventResult.Pull.BaseBranch) {
		return nil
	}
	return s.Client.PostMessage(s.Channel, eventResult)
}
