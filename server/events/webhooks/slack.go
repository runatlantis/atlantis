package webhooks

import (
	"regexp"

	"fmt"

	"github.com/atlantisnorth/atlantis/server/logging"
	"github.com/pkg/errors"
)

// SlackWebhook sends webhooks to Slack.
type SlackWebhook struct {
	Client         SlackClient
	WorkspaceRegex *regexp.Regexp
	Channel        string
}

func NewSlack(r *regexp.Regexp, channel string, client SlackClient) (*SlackWebhook, error) {
	if err := client.AuthTest(); err != nil {
		return nil, fmt.Errorf("testing slack authentication: %s. Verify your slack-token is valid", err)
	}

	channelExists, err := client.ChannelExists(channel)
	if err != nil {
		return nil, err
	}
	if !channelExists {
		return nil, errors.Errorf("slack channel %q doesn't exist", channel)
	}

	return &SlackWebhook{
		Client:         client,
		WorkspaceRegex: r,
		Channel:        channel,
	}, nil
}

// Send sends the webhook to Slack if the workspace matches the regex.
func (s *SlackWebhook) Send(log *logging.SimpleLogger, applyResult ApplyResult) error {
	if !s.WorkspaceRegex.MatchString(applyResult.Workspace) {
		return nil
	}
	return s.Client.PostMessage(s.Channel, applyResult)
}
