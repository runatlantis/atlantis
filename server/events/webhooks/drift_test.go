// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package webhooks_test

import (
	"net/http"
	"testing"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/events/webhooks"
	"github.com/runatlantis/atlantis/server/events/webhooks/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

var driftResult = webhooks.DriftResult{
	Repository:        "owner/repo",
	Ref:               "main",
	DetectionID:       "det-123",
	ProjectsWithDrift: 1,
	TotalProjects:     2,
	Projects: []webhooks.DriftProjectResult{
		{
			ProjectName: "project1",
			Path:        "infra/project1",
			Workspace:   "default",
			HasDrift:    true,
			ToAdd:       1,
			ToChange:    2,
			ToDestroy:   0,
			Summary:     "Plan: 1 to add, 2 to change, 0 to destroy.",
		},
		{
			ProjectName: "project2",
			Path:        "infra/project2",
			Workspace:   "default",
			HasDrift:    false,
		},
	},
}

func TestDriftWebhookSender_Send_SingleSuccess(t *testing.T) {
	RegisterMockTestingT(t)
	sender := mocks.NewMockDriftSender()
	manager := webhooks.DriftWebhookSender{
		Webhooks: []webhooks.DriftSender{sender},
	}
	logger := logging.NewNoopLogger(t)
	err := manager.Send(logger, driftResult)
	Ok(t, err)
	sender.VerifyWasCalledOnce().Send(logger, driftResult)
}

func TestDriftWebhookSender_Send_MultipleSuccess(t *testing.T) {
	RegisterMockTestingT(t)
	senders := []*mocks.MockDriftSender{
		mocks.NewMockDriftSender(),
		mocks.NewMockDriftSender(),
	}
	manager := webhooks.DriftWebhookSender{
		Webhooks: []webhooks.DriftSender{senders[0], senders[1]},
	}
	logger := logging.NewNoopLogger(t)
	err := manager.Send(logger, driftResult)
	Ok(t, err)
	for _, s := range senders {
		s.VerifyWasCalledOnce().Send(logger, driftResult)
	}
}

func TestDriftWebhookSender_Send_NoWebhooksNoOp(t *testing.T) {
	manager := webhooks.DriftWebhookSender{}
	logger := logging.NewNoopLogger(t)
	err := manager.Send(logger, driftResult)
	Ok(t, err)
}

func TestNewDriftWebhookSender_NoConfigs(t *testing.T) {
	sender, err := webhooks.NewDriftWebhookSender(nil, webhooks.Clients{})
	Ok(t, err)
	Equals(t, 0, len(sender.Webhooks))
}

func TestNewDriftWebhookSender_SkipsApplyConfigs(t *testing.T) {
	configs := []webhooks.Config{
		{Event: webhooks.ApplyEvent, Kind: webhooks.SlackKind, Channel: "ch"},
	}
	sender, err := webhooks.NewDriftWebhookSender(configs, webhooks.Clients{})
	Ok(t, err)
	Equals(t, 0, len(sender.Webhooks))
}

func TestNewDriftWebhookSender_SlackSuccess(t *testing.T) {
	RegisterMockTestingT(t)
	slackClient := mocks.NewMockSlackClient()
	When(slackClient.TokenIsSet()).ThenReturn(true)

	configs := []webhooks.Config{
		{Event: webhooks.DriftEvent, Kind: webhooks.SlackKind, Channel: "drift-alerts"},
	}
	clients := webhooks.Clients{Slack: slackClient}
	sender, err := webhooks.NewDriftWebhookSender(configs, clients)
	Ok(t, err)
	Equals(t, 1, len(sender.Webhooks))
}

func TestNewDriftWebhookSender_SlackNoToken(t *testing.T) {
	RegisterMockTestingT(t)
	slackClient := mocks.NewMockSlackClient()
	When(slackClient.TokenIsSet()).ThenReturn(false)

	configs := []webhooks.Config{
		{Event: webhooks.DriftEvent, Kind: webhooks.SlackKind, Channel: "drift-alerts"},
	}
	clients := webhooks.Clients{Slack: slackClient}
	_, err := webhooks.NewDriftWebhookSender(configs, clients)
	Assert(t, err != nil, "expected error when slack token is not set")
	ErrContains(t, "slack-token", err)
}

func TestNewDriftWebhookSender_SlackNoChannel(t *testing.T) {
	RegisterMockTestingT(t)
	slackClient := mocks.NewMockSlackClient()
	When(slackClient.TokenIsSet()).ThenReturn(true)

	configs := []webhooks.Config{
		{Event: webhooks.DriftEvent, Kind: webhooks.SlackKind, Channel: ""},
	}
	clients := webhooks.Clients{Slack: slackClient}
	_, err := webhooks.NewDriftWebhookSender(configs, clients)
	Assert(t, err != nil, "expected error when channel is empty")
	ErrContains(t, "channel", err)
}

func TestNewDriftWebhookSender_HttpSuccess(t *testing.T) {
	configs := []webhooks.Config{
		{Event: webhooks.DriftEvent, Kind: webhooks.HttpKind, URL: "http://example.com/webhook"},
	}
	clients := webhooks.Clients{
		Http: &webhooks.HttpClient{Client: http.DefaultClient},
	}
	sender, err := webhooks.NewDriftWebhookSender(configs, clients)
	Ok(t, err)
	Equals(t, 1, len(sender.Webhooks))
}

func TestNewDriftWebhookSender_HttpNoURL(t *testing.T) {
	configs := []webhooks.Config{
		{Event: webhooks.DriftEvent, Kind: webhooks.HttpKind, URL: ""},
	}
	clients := webhooks.Clients{
		Http: &webhooks.HttpClient{Client: http.DefaultClient},
	}
	_, err := webhooks.NewDriftWebhookSender(configs, clients)
	Assert(t, err != nil, "expected error when URL is empty")
	ErrContains(t, "url", err)
}

func TestNewDriftWebhookSender_UnsupportedKind(t *testing.T) {
	configs := []webhooks.Config{
		{Event: webhooks.DriftEvent, Kind: "unsupported"},
	}
	_, err := webhooks.NewDriftWebhookSender(configs, webhooks.Clients{})
	Assert(t, err != nil, "expected error for unsupported kind")
	ErrContains(t, "unsupported", err)
}

func TestNewDriftWebhookSender_NoKind(t *testing.T) {
	configs := []webhooks.Config{
		{Event: webhooks.DriftEvent, Kind: ""},
	}
	_, err := webhooks.NewDriftWebhookSender(configs, webhooks.Clients{})
	Assert(t, err != nil, "expected error when kind is empty")
	ErrContains(t, "kind", err)
}

func TestNewDriftWebhookSender_MixedConfigs(t *testing.T) {
	RegisterMockTestingT(t)
	slackClient := mocks.NewMockSlackClient()
	When(slackClient.TokenIsSet()).ThenReturn(true)

	configs := []webhooks.Config{
		{Event: webhooks.ApplyEvent, Kind: webhooks.SlackKind, Channel: "apply-ch", WorkspaceRegex: ".*", BranchRegex: ".*"},
		{Event: webhooks.DriftEvent, Kind: webhooks.SlackKind, Channel: "drift-ch"},
		{Event: webhooks.DriftEvent, Kind: webhooks.HttpKind, URL: "http://example.com/drift"},
	}
	clients := webhooks.Clients{
		Slack: slackClient,
		Http:  &webhooks.HttpClient{Client: http.DefaultClient},
	}
	sender, err := webhooks.NewDriftWebhookSender(configs, clients)
	Ok(t, err)
	Equals(t, 2, len(sender.Webhooks))
}
