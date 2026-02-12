// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package webhooks_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/runatlantis/atlantis/server/events/webhooks"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestDriftHttpWebhook_Send(t *testing.T) {
	var receivedBody webhooks.DriftResult
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Equals(t, "POST", r.Method)
		Equals(t, "application/json", r.Header.Get("Content-Type"))
		body, err := io.ReadAll(r.Body)
		Ok(t, err)
		err = json.Unmarshal(body, &receivedBody)
		Ok(t, err)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	hook := webhooks.DriftHttpWebhook{
		Client: &webhooks.HttpClient{Client: http.DefaultClient},
		URL:    server.URL,
	}

	err := hook.Send(logging.NewNoopLogger(t), driftResult)
	Ok(t, err)

	Equals(t, driftResult.Repository, receivedBody.Repository)
	Equals(t, driftResult.DetectionID, receivedBody.DetectionID)
	Equals(t, driftResult.ProjectsWithDrift, receivedBody.ProjectsWithDrift)
	Equals(t, driftResult.TotalProjects, receivedBody.TotalProjects)
	Equals(t, len(driftResult.Projects), len(receivedBody.Projects))
}

func TestDriftHttpWebhook_SendWithHeaders(t *testing.T) {
	expectedHeaders := map[string][]string{
		"Authorization": {"Bearer drift-token"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Equals(t, "Bearer drift-token", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	hook := webhooks.DriftHttpWebhook{
		Client: &webhooks.HttpClient{Client: http.DefaultClient, Headers: expectedHeaders},
		URL:    server.URL,
	}

	err := hook.Send(logging.NewNoopLogger(t), driftResult)
	Ok(t, err)
}

func TestDriftHttpWebhook_Send500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	hook := webhooks.DriftHttpWebhook{
		Client: &webhooks.HttpClient{Client: http.DefaultClient},
		URL:    server.URL,
	}

	err := hook.Send(logging.NewNoopLogger(t), driftResult)
	Assert(t, err != nil, "expected error on 500")
	ErrContains(t, "status 500", err)
}
