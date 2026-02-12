// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/runatlantis/atlantis/server/logging"
)

// DriftHttpWebhook sends drift notifications to an HTTP endpoint.
type DriftHttpWebhook struct {
	Client *HttpClient
	URL    string
}

// Send sends the drift result to the configured HTTP endpoint.
func (h *DriftHttpWebhook) Send(_ logging.SimpleLogging, result DriftResult) error {
	body, err := json.Marshal(result)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", h.URL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for header, values := range h.Client.Headers {
		for _, value := range values {
			req.Header.Add(header, value)
		}
	}
	resp, err := h.Client.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("drift webhook to %q returned status %d: %s", h.URL, resp.StatusCode, respBody)
	}
	return nil
}
