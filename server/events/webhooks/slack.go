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
	ProjectRegex   *regexp.Regexp
	DirectoryRegex *regexp.Regexp
	Channel        string
}

func NewSlack(wr *regexp.Regexp, br *regexp.Regexp, pr *regexp.Regexp, dr *regexp.Regexp, channel string, client SlackClient) (*SlackWebhook, error) {
	if err := client.AuthTest(); err != nil {
		return nil, fmt.Errorf("testing slack authentication: %s. Verify your slack-token is valid", err)
	}

	return &SlackWebhook{
		Client:         client,
		WorkspaceRegex: wr,
		BranchRegex:    br,
		ProjectRegex:   pr,
		DirectoryRegex: dr,
		Channel:        channel,
	}, nil
}

// Send sends the webhook to Slack if workspace and branch matches their respective regex.
func (s *SlackWebhook) Send(_ logging.SimpleLogging, applyResult ApplyResult) error {
	if !s.WorkspaceRegex.MatchString(applyResult.Workspace) || !s.BranchRegex.MatchString(applyResult.Pull.BaseBranch) || !s.ProjectRegex.MatchString(applyResult.ProjectName) || !s.DirectoryRegex.MatchString(applyResult.Directory) {
		return nil
	}
	return s.Client.PostMessage(s.Channel, applyResult)
}
