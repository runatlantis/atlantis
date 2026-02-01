// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package websocket

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/runatlantis/atlantis/server/logging"
)

func NewWriter(log logging.SimpleLogging, checkOrigin bool) *Writer {
	upgrader := websocket.Upgrader{
		CheckOrigin: checkOriginFunc(checkOrigin),
	}
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	return &Writer{
		upgrader: upgrader,
		log:      log,
	}
}

type Writer struct {
	upgrader websocket.Upgrader
	log      logging.SimpleLogging
}

func (w *Writer) Write(rw http.ResponseWriter, r *http.Request, input chan string) error {
	conn, err := w.upgrader.Upgrade(rw, r, nil)

	if err != nil {
		return fmt.Errorf("upgrading websocket connection: %w", err)
	}

	// Use defer to ensure Close Frame is always sent (RFC 6455 compliance)
	defer func() {
		closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
		deadline := time.Now().Add(time.Second)
		if err := conn.WriteControl(websocket.CloseMessage, closeMsg, deadline); err != nil {
			w.log.Warn("Failed to send ws close frame: %s", err)
		}
		if err := conn.Close(); err != nil {
			w.log.Warn("Failed to close ws connection: %s", err)
		}
	}()

	// Block on reading our input channel.
	// Add small delay between messages to work around GCP HTTP(S) LB
	// buffering issues with rapid WebSocket frames.
	for msg := range input {
		if err := conn.WriteMessage(websocket.BinaryMessage, []byte("\r"+msg+"\n")); err != nil {
			w.log.Warn("Failed to write ws message: %s", err)
			return err
		}
		time.Sleep(time.Millisecond)
	}

	return nil
}
