// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/runatlantis/atlantis/server/logging"
)

// HttpWebhook sends webhooks to any HTTP destination.
type HttpWebhook struct {
	Client         *HttpClient
	WorkspaceRegex *regexp.Regexp
	BranchRegex    *regexp.Regexp
	URL            string
}

// Send sends the webhook to URL if workspace and branch matches their respective regex.
func (h *HttpWebhook) Send(_ logging.SimpleLogging, applyResult ApplyResult) error {
	if !h.WorkspaceRegex.MatchString(applyResult.Workspace) || !h.BranchRegex.MatchString(applyResult.Pull.BaseBranch) {
		return nil
	}
	if err := h.doSend(applyResult); err != nil {
		return fmt.Errorf("sending webhook to %q: %w", h.URL, err)
	}
	return nil
}

func (h *HttpWebhook) doSend(applyResult ApplyResult) error {
	body, err := json.Marshal(applyResult)
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
		return fmt.Errorf("returned status code %d with response %q", resp.StatusCode, respBody)
	}
	return nil
}

// HttpClient wraps http.Client allowing to add arbitrary Headers to a request.
type HttpClient struct {
	Client  *http.Client
	Headers map[string][]string
}
