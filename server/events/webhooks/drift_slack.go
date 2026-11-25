// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"github.com/runatlantis/atlantis/server/logging"
)

// DriftSlackWebhook sends drift notifications to Slack.
type DriftSlackWebhook struct {
	Client  SlackClient
	Channel string
}

// Send sends the drift result to Slack.
func (s *DriftSlackWebhook) Send(_ logging.SimpleLogging, result DriftResult) error {
	return s.Client.PostDriftMessage(s.Channel, result)
}
