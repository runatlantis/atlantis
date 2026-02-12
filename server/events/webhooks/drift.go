// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"errors"
	"fmt"

	"github.com/runatlantis/atlantis/server/logging"
)

const DriftEvent = "drift"

// DriftResult is the payload sent to drift webhooks.
type DriftResult struct {
	Repository        string               `json:"repository"`
	Ref               string               `json:"ref"`
	DetectionID       string               `json:"detection_id"`
	ProjectsWithDrift int                   `json:"projects_with_drift"`
	TotalProjects     int                   `json:"total_projects"`
	Projects          []DriftProjectResult  `json:"projects"`
}

// DriftProjectResult describes drift for a single project.
type DriftProjectResult struct {
	ProjectName string `json:"project_name"`
	Path        string `json:"path"`
	Workspace   string `json:"workspace"`
	HasDrift    bool   `json:"has_drift"`
	ToAdd       int    `json:"to_add"`
	ToChange    int    `json:"to_change"`
	ToDestroy   int    `json:"to_destroy"`
	Summary     string `json:"summary"`
	Error       string `json:"error,omitempty"`
}

//go:generate pegomock generate --package mocks -o mocks/mock_drift_sender.go DriftSender

// DriftSender sends drift webhooks.
type DriftSender interface {
	Send(log logging.SimpleLogging, result DriftResult) error
}

// DriftWebhookSender distributes drift notifications to all configured senders.
type DriftWebhookSender struct {
	Webhooks []DriftSender
}

// Send sends drift results to all configured webhook senders.
// Errors are logged but not propagated (fire-and-forget).
func (d *DriftWebhookSender) Send(log logging.SimpleLogging, result DriftResult) error {
	for _, w := range d.Webhooks {
		if err := w.Send(log, result); err != nil {
			log.Warn("error sending drift webhook: %s", err)
		}
	}
	return nil
}

// NewDriftWebhookSender creates a DriftWebhookSender from webhook configs.
// It filters configs to only those with event: drift and validates kind/channel/url.
// Returns a sender with no webhooks (no-op) if no drift configs are found.
func NewDriftWebhookSender(configs []Config, clients Clients) (*DriftWebhookSender, error) {
	var senders []DriftSender
	for _, c := range configs {
		if c.Event != DriftEvent {
			continue
		}
		if c.Kind == "" {
			return nil, errors.New("must specify \"kind\" key for drift webhooks")
		}
		switch c.Kind {
		case SlackKind:
			if !clients.Slack.TokenIsSet() {
				return nil, errors.New("must specify top-level \"slack-token\" if using a drift webhook of \"kind: slack\"")
			}
			if c.Channel == "" {
				return nil, errors.New("must specify \"channel\" for drift webhook of \"kind: slack\"")
			}
			senders = append(senders, &DriftSlackWebhook{
				Client:  clients.Slack,
				Channel: c.Channel,
			})
		case HttpKind:
			if c.URL == "" {
				return nil, errors.New("must specify \"url\" for drift webhook of \"kind: http\"")
			}
			senders = append(senders, &DriftHttpWebhook{
				Client: clients.Http,
				URL:    c.URL,
			})
		default:
			return nil, fmt.Errorf("\"kind: %s\" not supported for drift webhooks. Only \"kind: %s\" and \"kind: %s\" are supported", c.Kind, SlackKind, HttpKind)
		}
	}
	return &DriftWebhookSender{Webhooks: senders}, nil
}
