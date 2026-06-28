// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/runatlantis/atlantis/server/logging"
)

var (
	driftWebhookURLRE           = regexp.MustCompile(`https?://[^\s'"]+`)
	driftWebhookURLCredentialRE = regexp.MustCompile(`(?i)(https?://)([^\s/@]+:)?[^\s/@]+@`)
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
		return fmt.Errorf("creating drift webhook request for %q: %s", sanitizeDriftWebhookURL(h.URL), sanitizeDriftWebhookError(err.Error()))
	}
	req.Header.Set("Content-Type", "application/json")
	for header, values := range h.Client.Headers {
		for _, value := range values {
			req.Header.Add(header, value)
		}
	}
	resp, err := h.Client.Client.Do(req)
	if err != nil {
		return fmt.Errorf("sending drift webhook to %q: %s", sanitizeDriftWebhookURL(h.URL), sanitizeDriftWebhookError(err.Error()))
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("drift webhook to %q returned status %d: %s", sanitizeDriftWebhookURL(h.URL), resp.StatusCode, sanitizeDriftWebhookError(string(respBody)))
	}
	return nil
}

func sanitizeDriftWebhookURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "<invalid-url>"
	}
	return parsed.Scheme + "://" + parsed.Host
}

func sanitizeDriftWebhookError(message string) string {
	message = driftWebhookURLRE.ReplaceAllStringFunc(message, sanitizeDriftWebhookURL)
	message = driftWebhookURLCredentialRE.ReplaceAllString(message, "$1<redacted>@")
	for _, key := range []string{"token", "access_token", "key", "secret", "signature"} {
		message = redactQueryValue(message, key)
	}
	return message
}

func redactQueryValue(message, key string) string {
	lower := strings.ToLower(message)
	needle := strings.ToLower(key) + "="
	var builder strings.Builder
	last := 0
	searchFrom := 0
	for {
		idxRel := strings.Index(lower[searchFrom:], needle)
		if idxRel == -1 {
			if last == 0 {
				return message
			}
			builder.WriteString(message[last:])
			return builder.String()
		}
		idx := searchFrom + idxRel
		start := idx + len(key) + 1
		end := start
		for end < len(message) && message[end] != '&' && message[end] != ' ' && message[end] != '\'' && message[end] != '"' {
			end++
		}
		builder.WriteString(message[last:start])
		builder.WriteString("<redacted>")
		last = end
		searchFrom = end
	}
}
