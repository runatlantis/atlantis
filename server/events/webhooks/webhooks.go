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

type Event string

const (
	ApplyEvent Event = "apply"
	PlanEvent  Event = "plan"
)

//go:generate go tool pegomock generate --package mocks -o mocks/mock_sender.go Sender

// Sender sends webhooks.
type Sender interface {
	// Send sends the webhook (if the implementation thinks it should).
	Send(log logging.SimpleLogging, eventResult EventResult) error
}

// EventResult is the result of a terraform apply.
type EventResult struct {
	Event       Event
	Workspace   string
	Repo        models.Repo
	Pull        models.PullRequest
	User        models.User
	Success     bool
	Directory   string
	ProjectName string
}

// ConfiguredSender couples a sender with the event it should handle.
type ConfiguredSender struct {
	Event  Event
	Sender Sender
}

// MultiWebhookSender sends multiple webhooks for each one it's configured for.
type MultiWebhookSender struct {
	Webhooks []ConfiguredSender
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
	Slack SlackClient
	Http  *HttpClient
}

func NewMultiWebhookSender(configs []Config, clients Clients) (*MultiWebhookSender, error) {
	var webhooks []ConfiguredSender
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
		evt := Event(c.Event)
		switch evt {
		case ApplyEvent, PlanEvent:
			// ok
		default:
			return nil, fmt.Errorf("\"event: %s\" not supported. Only \"events: %s, %s\" are supported right now", c.Event, ApplyEvent, PlanEvent)
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
			webhooks = append(webhooks, ConfiguredSender{
				Event:  evt,
				Sender: slack,
			})
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
			webhooks = append(webhooks, ConfiguredSender{
				Event:  evt,
				Sender: httpWebhook,
			})
		default:
			return nil, fmt.Errorf("\"kind: %s\" not supported. Only \"kind: %s\" and \"kind: %s\" are supported right now", c.Kind, SlackKind, HttpKind)
		}
	}

	return &MultiWebhookSender{
		Webhooks: webhooks,
	}, nil
}

// Send sends the webhook using its Webhooks.
func (w *MultiWebhookSender) Send(log logging.SimpleLogging, result EventResult) error {
	for _, w := range w.Webhooks {
		if w.Event != result.Event {
			continue
		}
		if err := w.Sender.Send(log, result); err != nil {
			log.Warn("error sending webhook: %s", err)
		}
	}
	return nil
}
