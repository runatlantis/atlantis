// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package webhooks

import (
	"fmt"
	"regexp"

	"errors"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

const SlackKind = "slack"
const HttpKind = "http"
const MSTeamsKind = "msteams"
const ApplyEvent = "apply"

//go:generate go tool pegomock generate --package mocks -o mocks/mock_sender.go Sender

// Sender sends webhooks.
type Sender interface {
	// Send sends the webhook (if the implementation thinks it should).
	Send(log logging.SimpleLogging, applyResult ApplyResult) error
}

// ApplyResult is the result of a terraform apply.
type ApplyResult struct {
	Workspace   string
	Repo        models.Repo
	Pull        models.PullRequest
	User        models.User
	Success     bool
	Directory   string
	ProjectName string
}

// MultiWebhookSender sends multiple webhooks for each one it's configured for.
type MultiWebhookSender struct {
	Webhooks []Sender
}

type Config struct {
	Event          string
	WorkspaceRegex string
	BranchRegex    string
	Kind           string
	Channel        string
	URL            string
}

type Clients struct {
	Slack   SlackClient
	Http    *HttpClient
	MSTeams MSTeamsClient
}

func NewMultiWebhookSender(configs []Config, clients Clients) (*MultiWebhookSender, error) {
	var webhooks []Sender
	for _, c := range configs {
		wr, err := regexp.Compile(c.WorkspaceRegex)
		if err != nil {
			return nil, err
		}
		br, err := regexp.Compile(c.BranchRegex)
		if err != nil {
			return nil, err
		}
		if c.Kind == "" || c.Event == "" {
			return nil, errors.New("must specify \"kind\" and \"event\" keys for webhooks")
		}
		if c.Event != ApplyEvent {
			return nil, fmt.Errorf("\"event: %s\" not supported. Only \"event: %s\" is supported right now", c.Event, ApplyEvent)
		}
		switch c.Kind {
		case SlackKind:
			if !clients.Slack.TokenIsSet() {
				return nil, errors.New("must specify top-level \"slack-token\" if using a webhook of \"kind: slack\"")
			}
			if c.Channel == "" {
				return nil, errors.New("must specify \"channel\" if using a webhook of \"kind: slack\"")
			}
			slack, err := NewSlack(wr, br, c.Channel, clients.Slack)
			if err != nil {
				return nil, err
			}
			webhooks = append(webhooks, slack)
		case HttpKind:
			if c.URL == "" {
				return nil, errors.New("must specify \"url\" if using a webhook of \"kind: http\"")
			}
			httpWebhook := &HttpWebhook{
				Client:         clients.Http,
				WorkspaceRegex: wr,
				BranchRegex:    br,
				URL:            c.URL,
			}
			webhooks = append(webhooks, httpWebhook)
		case MSTeamsKind:
			if c.URL == "" {
				return nil, errors.New("must specify \"url\" if using a webhook of \"kind: msteams\"")
			}
			teamsWebhook, err := NewMSTeams(wr, br, c.URL, clients.MSTeams)
			if err != nil {
				return nil, err
			}
			webhooks = append(webhooks, teamsWebhook)
		default:
			return nil, fmt.Errorf("\"kind: %s\" not supported. Only \"kind: %s\", \"kind: %s\", and \"kind: %s\" are supported right now", c.Kind, SlackKind, HttpKind, MSTeamsKind)
		}
	}

	return &MultiWebhookSender{
		Webhooks: webhooks,
	}, nil
}

// Send sends the webhook using its Webhooks.
func (w *MultiWebhookSender) Send(log logging.SimpleLogging, result ApplyResult) error {
	for _, w := range w.Webhooks {
		if err := w.Send(log, result); err != nil {
			log.Warn("error sending webhook: %s", err)
		}
	}
	return nil
}
