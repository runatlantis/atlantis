package webhooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/logging"
)

// HttpWebhook sends webhooks to any HTTP destination.
type HttpWebhook struct {
	Client         *http.Client
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
		return errors.Wrap(err, fmt.Sprintf("sending webhook to %q", h.URL))
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
	resp, err := h.Client.Do(req)
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

// NewHttpClient creates a new HTTP client that will add arbitrary headers to every request.
func NewHttpClient(headers map[string][]string) *http.Client {
	return &http.Client{
		Transport: &AuthedTransport{
			Base:    http.DefaultTransport,
			Headers: headers,
		},
	}
}

// AuthedTransport is a http.RoundTripper which wraps Base
// adding arbitrary Headers to each request.
type AuthedTransport struct {
	Base    http.RoundTripper
	Headers map[string][]string
}

// RoundTrip handles each http request.
func (t *AuthedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for header, values := range t.Headers {
		for _, value := range values {
			req.Header.Add(header, value)
		}
	}
	return t.Base.RoundTrip(req)
}
