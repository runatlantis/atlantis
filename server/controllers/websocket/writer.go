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

	// block on reading our input channel
	for msg := range input {
		if err := conn.WriteMessage(websocket.BinaryMessage, []byte("\r"+msg+"\n")); err != nil {
			w.log.Warn("Failed to write ws message: %s", err)
			return err
		}
	}

	// Send WebSocket Close Frame (RFC 6455) before closing the TCP connection.
	// GCP HTTP(S) LB workaround: the GFE prematurely terminates the connection
	// when WebSocket FIN and TCP FIN arrive together. Separating them with a
	// short delay prevents the client from receiving a TCP FIN before it has
	// finished reading all buffered data.
	closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
	if err = conn.WriteControl(websocket.CloseMessage, closeMsg, time.Now().Add(time.Second)); err != nil {
		w.log.Warn("Failed to send ws close frame: %s", err)
	}
	time.Sleep(50 * time.Millisecond)

	if err = conn.Close(); err != nil {
		w.log.Warn("Failed to close ws connection: %s", err)
	}
	return nil
}
