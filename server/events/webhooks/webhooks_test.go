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

package webhooks_test

import (
	"net/http"
	"strings"
	"testing"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/events/webhooks"
	"github.com/runatlantis/atlantis/server/events/webhooks/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

const (
	validEvent   = webhooks.ApplyEvent
	validRegex   = ".*"
	validKind    = webhooks.SlackKind
	validChannel = "validchannel"
)

var validConfig = webhooks.Config{
	Event:          validEvent,
	WorkspaceRegex: validRegex,
	BranchRegex:    validRegex,
	Kind:           validKind,
	Channel:        validChannel,
}

func validConfigs() []webhooks.Config {
	return []webhooks.Config{validConfig}
}

func validClients() webhooks.Clients {
	return webhooks.Clients{
		Slack: mocks.NewMockSlackClient(),
		Http:  &webhooks.HttpClient{Client: http.DefaultClient},
	}
}

func TestNewWebhooksManager_InvalidWorkspaceRegex(t *testing.T) {
	t.Log("When given an invalid workspace regex in a config, an error is returned")
	RegisterMockTestingT(t)
	clients := validClients()

	invalidRegex := "("
	configs := validConfigs()
	configs[0].WorkspaceRegex = invalidRegex
	_, err := webhooks.NewMultiWebhookSender(configs, clients)
	Assert(t, err != nil, "expected error")
	Assert(t, strings.Contains(err.Error(), "error parsing regexp"), "expected regex error")
}

func TestNewWebhooksManager_InvalidBranchRegex(t *testing.T) {
	t.Log("When given an invalid branch regex in a config, an error is returned")
	RegisterMockTestingT(t)
	clients := validClients()

	invalidRegex := "("
	configs := validConfigs()
	configs[0].BranchRegex = invalidRegex
	_, err := webhooks.NewMultiWebhookSender(configs, clients)
	Assert(t, err != nil, "expected error")
	Assert(t, strings.Contains(err.Error(), "error parsing regexp"), "expected regex error")
}

func TestNewWebhooksManager_InvalidBranchAndWorkspaceRegex(t *testing.T) {
	t.Log("When given an invalid branch and invalid workspace regex in a config, an error is returned")
	RegisterMockTestingT(t)
	clients := validClients()

	invalidRegex := "("
	configs := validConfigs()
	configs[0].WorkspaceRegex = invalidRegex
	configs[0].BranchRegex = invalidRegex
	_, err := webhooks.NewMultiWebhookSender(configs, clients)
	Assert(t, err != nil, "expected error")
	Assert(t, strings.Contains(err.Error(), "error parsing regexp"), "expected regex error")
}

func TestNewWebhooksManager_NoEvent(t *testing.T) {
	t.Log("When the event key is not specified in a config, an error is returned")
	RegisterMockTestingT(t)
	clients := validClients()
	configs := validConfigs()
	configs[0].Event = ""
	_, err := webhooks.NewMultiWebhookSender(configs, clients)
	Assert(t, err != nil, "expected error")
	Equals(t, "must specify \"kind\" and \"event\" keys for webhooks", err.Error())
}

func TestNewWebhooksManager_UnsupportedEvent(t *testing.T) {
	t.Log("When given an unsupported event in a config, an error is returned")
	RegisterMockTestingT(t)
	clients := validClients()

	unsupportedEvent := "badevent"
	configs := validConfigs()
	configs[0].Event = unsupportedEvent
	_, err := webhooks.NewMultiWebhookSender(configs, clients)
	Assert(t, err != nil, "expected error")
	Equals(t, "\"event: badevent\" not supported. Only \"event: apply\" is supported right now", err.Error())
}

func TestNewWebhooksManager_NoKind(t *testing.T) {
	t.Log("When the kind key is not specified in a config, an error is returned")
	RegisterMockTestingT(t)
	clients := validClients()
	configs := validConfigs()
	configs[0].Kind = ""
	_, err := webhooks.NewMultiWebhookSender(configs, clients)
	Assert(t, err != nil, "expected error")
	Equals(t, "must specify \"kind\" and \"event\" keys for webhooks", err.Error())
}

func TestNewWebhooksManager_UnsupportedKind(t *testing.T) {
	t.Log("When given an unsupported kind in a config, an error is returned")
	RegisterMockTestingT(t)
	clients := validClients()

	unsupportedKind := "badkind"
	configs := validConfigs()
	configs[0].Kind = unsupportedKind
	_, err := webhooks.NewMultiWebhookSender(configs, clients)
	Assert(t, err != nil, "expected error")
	Equals(t, "\"kind: badkind\" not supported. Only \"kind: slack\", \"kind: http\", and \"kind: msteams\" are supported right now", err.Error())
}

func TestNewWebhooksManager_NoConfigSuccess(t *testing.T) {
	t.Log("When there are no configs, function should succeed")
	t.Log("passing any client should succeed")
	var emptyConfigs []webhooks.Config
	emptyToken := ""
	anyClients := webhooks.Clients{
		Slack: webhooks.NewSlackClient(emptyToken),
		Http:  &webhooks.HttpClient{Client: http.DefaultClient},
	}
	m, err := webhooks.NewMultiWebhookSender(emptyConfigs, anyClients)
	Ok(t, err)
	Equals(t, 0, len(m.Webhooks)) // nolint: staticcheck

	t.Log("passing nil client should succeed")
	m, err = webhooks.NewMultiWebhookSender(emptyConfigs, webhooks.Clients{})
	Ok(t, err)
	Equals(t, 0, len(m.Webhooks)) // nolint: staticcheck
}
func TestNewWebhooksManager_SingleConfigSuccess(t *testing.T) {
	t.Log("When there is one valid config, function should succeed")
	RegisterMockTestingT(t)
	clients := validClients()
	When(clients.Slack.TokenIsSet()).ThenReturn(true)

	configs := validConfigs()
	m, err := webhooks.NewMultiWebhookSender(configs, clients)
	Ok(t, err)
	Equals(t, 1, len(m.Webhooks)) // nolint: staticcheck
}

func TestNewWebhooksManager_MultipleConfigSuccess(t *testing.T) {
	t.Log("When there are multiple valid configs, function should succeed")
	RegisterMockTestingT(t)
	clients := validClients()
	When(clients.Slack.TokenIsSet()).ThenReturn(true)

	var configs []webhooks.Config
	nConfigs := 5
	for range nConfigs {
		configs = append(configs, validConfig)
	}
	m, err := webhooks.NewMultiWebhookSender(configs, clients)
	Ok(t, err)
	Equals(t, nConfigs, len(m.Webhooks)) // nolint: staticcheck
}

func TestSend_SingleSuccess(t *testing.T) {
	t.Log("Sending one webhook should succeed")
	RegisterMockTestingT(t)
	sender := mocks.NewMockSender()
	manager := webhooks.MultiWebhookSender{
		Webhooks: []webhooks.Sender{sender},
	}
	logger := logging.NewNoopLogger(t)
	result := webhooks.ApplyResult{}
	manager.Send(logger, result) // nolint: errcheck
	sender.VerifyWasCalledOnce().Send(logger, result)
}

func TestSend_MultipleSuccess(t *testing.T) {
	t.Log("Sending multiple webhooks should succeed")
	RegisterMockTestingT(t)
	senders := []*mocks.MockSender{
		mocks.NewMockSender(),
		mocks.NewMockSender(),
		mocks.NewMockSender(),
	}
	manager := webhooks.MultiWebhookSender{
		Webhooks: []webhooks.Sender{senders[0], senders[1], senders[2]},
	}
	logger := logging.NewNoopLogger(t)
	result := webhooks.ApplyResult{}
	err := manager.Send(logger, result)
	Ok(t, err)
	for _, s := range senders {
		s.VerifyWasCalledOnce().Send(logger, result)
	}
}

func TestNewMultiWebhookSender_MSTeams(t *testing.T) {
	t.Log("Should be able to create a MS Teams webhook")
	configs := []webhooks.Config{
		{
			Event:          "apply",
			WorkspaceRegex: ".*",
			BranchRegex:    ".*",
			Kind:           "msteams",
			URL:            "https://outlook.office.com/webhook/test",
		},
	}
	clients := webhooks.Clients{
		MSTeams: webhooks.NewMSTeamsClient(),
	}
	m, err := webhooks.NewMultiWebhookSender(configs, clients)
	Ok(t, err)
	Equals(t, 1, len(m.Webhooks))
}

func TestNewMultiWebhookSender_MSTeamsNoURL(t *testing.T) {
	t.Log("Should return error if no URL is specified for MS Teams webhook")
	configs := []webhooks.Config{
		{
			Event:          "apply",
			WorkspaceRegex: ".*",
			BranchRegex:    ".*",
			Kind:           "msteams",
		},
	}
	clients := webhooks.Clients{
		MSTeams: webhooks.NewMSTeamsClient(),
	}
	_, err := webhooks.NewMultiWebhookSender(configs, clients)
	ErrContains(t, "must specify \"url\" if using a webhook of \"kind: msteams\"", err)
}
