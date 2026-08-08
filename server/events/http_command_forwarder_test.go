// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/core/ownership"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/stretchr/testify/require"
)

func TestHTTPCommandForwarder_ForwardsAuthenticatedCredentialFreePayload(t *testing.T) {
	dispatch := testCommentDispatch()
	var received events.CommentDispatch
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/atlantis/internal/commands/comment", r.URL.Path)
		require.Equal(t, "internal-secret", r.Header.Get(events.InternalCommandTokenHeader))
		require.Equal(t, "claim-1", r.Header.Get(events.OwnershipClaimHeader))
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))
		require.NoError(t, json.NewDecoder(r.Body).Decode(&received))
		w.WriteHeader(http.StatusAccepted)
	}))
	t.Cleanup(server.Close)
	forwarder := events.NewHTTPCommandForwarder("internal-secret")

	err := forwarder.ForwardComment(ownership.Record{
		ReplicaID:    "replica-1",
		ClaimID:      "claim-1",
		AdvertiseURL: server.URL + "/atlantis",
	}, dispatch)
	require.NoError(t, err)
	require.Equal(t, dispatch, received)

	payload, err := json.Marshal(received)
	require.NoError(t, err)
	require.NotContains(t, string(payload), "internal-secret")
}

func TestHTTPCommandForwarder_MapsConflictToOwnershipChanged(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "stale ownership", http.StatusConflict)
	}))
	t.Cleanup(server.Close)
	forwarder := events.NewHTTPCommandForwarder("internal-secret")

	err := forwarder.ForwardComment(ownership.Record{ClaimID: "claim-1", AdvertiseURL: server.URL}, testCommentDispatch())
	require.ErrorIs(t, err, events.ErrOwnershipChanged)
}

func TestHTTPCommandForwarder_DoesNotForwardTokenAcrossRedirect(t *testing.T) {
	leakedToken := make(chan string, 1)
	redirectTarget := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		leakedToken <- r.Header.Get(events.InternalCommandTokenHeader)
	}))
	t.Cleanup(redirectTarget.Close)
	redirector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, redirectTarget.URL, http.StatusTemporaryRedirect)
	}))
	t.Cleanup(redirector.Close)
	forwarder := events.NewHTTPCommandForwarder("internal-secret")

	err := forwarder.ForwardComment(ownership.Record{ClaimID: "claim-1", AdvertiseURL: redirector.URL}, testCommentDispatch())
	require.Error(t, err)
	select {
	case token := <-leakedToken:
		require.Empty(t, token)
	default:
	}
}

func TestHTTPCommandForwarder_RejectsUnsafeAdvertiseURL(t *testing.T) {
	forwarder := events.NewHTTPCommandForwarder("internal-secret")
	tests := []string{
		"http://user:password@replica-1",
		"http://replica-1?redirect=http://attacker",
		"file:///tmp/socket",
		"//replica-1",
	}
	for _, advertiseURL := range tests {
		err := forwarder.ForwardComment(ownership.Record{ClaimID: "claim-1", AdvertiseURL: advertiseURL}, testCommentDispatch())
		require.Error(t, err, advertiseURL)
	}
}

func TestHTTPCommandForwarder_CapsErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = io.WriteString(w, strings.Repeat("x", 16*1024))
	}))
	t.Cleanup(server.Close)
	forwarder := events.NewHTTPCommandForwarder("internal-secret")

	err := forwarder.ForwardComment(ownership.Record{ClaimID: "claim-1", AdvertiseURL: server.URL}, testCommentDispatch())
	require.Error(t, err)
	require.LessOrEqual(t, len(err.Error()), 4*1024+128)
}

func TestHTTPCommandForwarder_ReturnsTransportErrors(t *testing.T) {
	forwarder := events.NewHTTPCommandForwarder("internal-secret")
	err := forwarder.ForwardComment(ownership.Record{ClaimID: "claim-1", AdvertiseURL: "http://127.0.0.1:1"}, testCommentDispatch())
	require.Error(t, err)
	require.False(t, errors.Is(err, events.ErrOwnershipChanged))
}
