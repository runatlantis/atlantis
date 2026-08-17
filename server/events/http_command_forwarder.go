// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/runatlantis/atlantis/server/core/ownership"
)

const (
	InternalCommentCommandPath    = "/internal/commands/comment"
	InternalAutoplanCommandPath   = "/internal/commands/autoplan"
	InternalPullClosedCommandPath = "/internal/commands/pull-closed"
	InternalCommandTokenHeader    = "X-Atlantis-Internal-Token"
	OwnershipClaimHeader          = "X-Atlantis-Ownership-Claim"

	internalForwardTimeout = 5 * time.Second
	maxInternalErrorBody   = 4 * 1024
)

// HTTPCommandForwarder sends credential-free commands to an owning replica.
type HTTPCommandForwarder struct {
	token  string
	client *http.Client
}

// NewHTTPCommandForwarder creates an internal command forwarder.
func NewHTTPCommandForwarder(token string) *HTTPCommandForwarder {
	return &HTTPCommandForwarder{
		token: token,
		client: &http.Client{
			Timeout: internalForwardTimeout,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func (f *HTTPCommandForwarder) ForwardComment(owner ownership.Record, command CommentDispatch) error {
	return f.forward(owner, InternalCommentCommandPath, command)
}

func (f *HTTPCommandForwarder) ForwardAutoplan(owner ownership.Record, command AutoplanDispatch) error {
	return f.forward(owner, InternalAutoplanCommandPath, command)
}

func (f *HTTPCommandForwarder) ForwardPullClosed(owner ownership.Record, command PullClosedDispatch) error {
	return f.forward(owner, InternalPullClosedCommandPath, command)
}

func (f *HTTPCommandForwarder) forward(owner ownership.Record, endpoint string, command any) error {
	target, err := internalCommandURL(owner.AdvertiseURL, endpoint)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(command)
	if err != nil {
		return fmt.Errorf("serializing internal command: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, target, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("creating internal command request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(InternalCommandTokenHeader, f.token)
	req.Header.Set(OwnershipClaimHeader, owner.ClaimID)

	resp, err := f.client.Do(req)
	if err != nil {
		return fmt.Errorf("sending internal command: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusAccepted {
		return nil
	}
	if resp.StatusCode == http.StatusConflict {
		return ErrOwnershipChanged
	}
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxInternalErrorBody))
	if readErr != nil {
		return fmt.Errorf("internal command returned HTTP %d (reading response: %w)", resp.StatusCode, readErr)
	}
	message := strings.TrimSpace(string(body))
	if message == "" {
		return fmt.Errorf("internal command returned HTTP %d", resp.StatusCode)
	}
	return fmt.Errorf("internal command returned HTTP %d: %s", resp.StatusCode, message)
}

func internalCommandURL(advertiseURL, endpoint string) (string, error) {
	base, err := url.Parse(advertiseURL)
	if err != nil {
		return "", fmt.Errorf("parsing owner advertise URL: %w", err)
	}
	if base.Scheme != "http" && base.Scheme != "https" {
		return "", errors.New("owner advertise URL must use HTTP or HTTPS")
	}
	if base.Host == "" || base.User != nil || base.RawQuery != "" || base.Fragment != "" || base.Opaque != "" {
		return "", errors.New("owner advertise URL contains unsupported components")
	}
	targetString, err := url.JoinPath(base.String(), strings.TrimPrefix(endpoint, "/"))
	if err != nil {
		return "", fmt.Errorf("joining internal command URL: %w", err)
	}
	target, err := url.Parse(targetString)
	if err != nil {
		return "", fmt.Errorf("parsing internal command URL: %w", err)
	}
	if !strings.EqualFold(target.Scheme, base.Scheme) || !strings.EqualFold(target.Host, base.Host) {
		return "", errors.New("internal command URL changed origin")
	}
	return target.String(), nil
}
