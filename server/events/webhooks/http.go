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
	Client         *HttpClient
	WorkspaceRegex *regexp.Regexp
	BranchRegex    *regexp.Regexp
	URL            string
}

// SendApplyResult sends the apply webhook to URL if workspace and branch matches their respective regex.
func (h *HttpWebhook) SendApplyResult(_ logging.SimpleLogging, applyResult ApplyResult) error {
	if !h.WorkspaceRegex.MatchString(applyResult.Workspace) || !h.BranchRegex.MatchString(applyResult.Pull.BaseBranch) {
		return nil
	}
	if err := h.doSendApplyResult(applyResult); err != nil {
		return errors.Wrap(err, fmt.Sprintf("sending apply webhook to %q", h.URL))
	}
	return nil
}

// SendPlanResult sends the plan webhook to URL if workspace and branch matches their respective regex.
func (h *HttpWebhook) SendPlanResult(_ logging.SimpleLogging, planResult PlanResult) error {
	if !h.WorkspaceRegex.MatchString(planResult.Workspace) || !h.BranchRegex.MatchString(planResult.Pull.BaseBranch) {
		return nil
	}
	if err := h.doSendPlanResult(planResult); err != nil {
		return errors.Wrap(err, fmt.Sprintf("sending plan webhook to %q", h.URL))
	}
	return nil
}

// Send is kept for backward compatibility.
// Deprecated: Use SendApplyResult instead.
func (h *HttpWebhook) Send(log logging.SimpleLogging, applyResult ApplyResult) error {
	return h.SendApplyResult(log, applyResult)
}

func (h *HttpWebhook) doSendApplyResult(applyResult ApplyResult) error {
	body, err := json.Marshal(applyResult)
	if err != nil {
		return err
	}
	return h.sendRequest(body)
}

func (h *HttpWebhook) doSendPlanResult(planResult PlanResult) error {
	body, err := json.Marshal(planResult)
	if err != nil {
		return err
	}
	return h.sendRequest(body)
}

func (h *HttpWebhook) sendRequest(body []byte) error {
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
