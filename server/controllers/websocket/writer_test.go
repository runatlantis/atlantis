// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package websocket

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/stretchr/testify/assert"
)

func TestWriter_Write(t *testing.T) {
	log := logging.NewNoopLogger(t)

	t.Run("client disconnects", func(t *testing.T) {
		writer := NewWriter(log, true)
		input := make(chan string, 1)

		// Test server that uses the writer
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := writer.Write(w, r, input)
			assert.NoError(t, err, "writer.Write should not return an error on client disconnect")
		}))
		defer server.Close()

		// Client connects to the server
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
		ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		assert.NoError(t, err, "failed to connect to websocket")

		// Simulate client disconnect
		ws.Close()

		// Try to send a message, this should not block or panic
		input <- "test"

		// Give it some time to process
		time.Sleep(100 * time.Millisecond)
	})
}
