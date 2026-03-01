// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

package webhooks

import (
	"regexp"

	"github.com/runatlantis/atlantis/server/logging"
)

// MSTeamsWebhook sends webhooks to Microsoft Teams.
type MSTeamsWebhook struct {
	Client         MSTeamsClient
	WorkspaceRegex *regexp.Regexp
	BranchRegex    *regexp.Regexp
	WebhookURL     string
}

// NewMSTeams creates a new MS Teams webhook.
func NewMSTeams(wr *regexp.Regexp, br *regexp.Regexp, webhookURL string, client MSTeamsClient) (*MSTeamsWebhook, error) {
	return &MSTeamsWebhook{
		Client:         client,
		WorkspaceRegex: wr,
		BranchRegex:    br,
		WebhookURL:     webhookURL,
	}, nil
}

// Send sends the webhook to MS Teams if workspace and branch matches their respective regex.
func (m *MSTeamsWebhook) Send(_ logging.SimpleLogging, applyResult ApplyResult) error {
	if !m.WorkspaceRegex.MatchString(applyResult.Workspace) || !m.BranchRegex.MatchString(applyResult.Pull.BaseBranch) {
		return nil
	}
	return m.Client.PostMessage(m.WebhookURL, applyResult)
}
