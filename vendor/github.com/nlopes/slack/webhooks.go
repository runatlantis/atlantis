package slack

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

type WebhookMessage struct {
	Text        string       `json:"text,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

func PostWebhook(url string, msg *WebhookMessage) error {
	raw, err := json.Marshal(msg)

	if err != nil {
		return errors.Wrap(err, "marshal failed")
	}

	response, err := http.Post(url, "application/json", bytes.NewReader(raw))

	if err != nil {
		return errors.Wrap(err, "failed to post webhook")
	}

	if response.StatusCode != http.StatusOK {
		return statusCodeError{Code: response.StatusCode, Status: response.Status}
	}

	return nil
}
