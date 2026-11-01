// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package websocket

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/runatlantis/atlantis/server/logging"
)

const (
	writeTimeout          = 10 * time.Second
	closeHandshakeTimeout = 2 * time.Second
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
	// Unlike Gorilla's default handler, preserve a failed Close reply so a
	// transport failure is not mistaken for a completed closing handshake.
	conn.SetCloseHandler(func(code int, _ string) error {
		err := conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(code, ""), time.Now().Add(writeTimeout))
		if errors.Is(err, websocket.ErrCloseSent) {
			return nil
		}
		return err
	})

	// Gorilla only processes Close/Ping/Pong handlers while reading. A single
	// reader also detects disconnects while the producer has no output.
	readDone := make(chan struct{})
	var readErr error
	go func() {
		defer close(readDone)
		readErr = readMessages(conn)
		// The close handler has attempted the reply to a peer's Close. RFC 6455
		// 5.5.1 requires the server to close TCP immediately after the exchange.
		// Read/protocol failures must also release the transport.
		_ = conn.Close()
	}()
	defer func() {
		_ = conn.Close()
		<-readDone // Closing TCP unblocks the reader on every exit path.
	}()

	for {
		select {
		case <-readDone:
			return closeReadError(readErr)
		case msg, ok := <-input:
			if !ok {
				readerDone, err := w.sendClose(conn, readDone)
				if err != nil || !readerDone {
					// readErr is only safe to read once the reader has exited.
					return err
				}
				return closeReadError(readErr)
			}
			if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
				return fmt.Errorf("setting websocket write deadline: %w", err)
			}
			if err := conn.WriteMessage(websocket.BinaryMessage, []byte("\r"+msg+"\n")); err != nil {
				// WriteControl serializes Close against data writes and prevents
				// further data frames once Close is sent, including a peer reply.
				if errors.Is(err, websocket.ErrCloseSent) {
					<-readDone
					return closeReadError(readErr)
				}
				return fmt.Errorf("writing websocket message: %w", err)
			}
		}
	}
}

// readMessages discards client data without buffering whole messages in memory.
func readMessages(conn *websocket.Conn) error {
	for {
		_, reader, err := conn.NextReader()
		if err != nil {
			return err
		}
		if _, err := io.Copy(io.Discard, reader); err != nil {
			return err
		}
	}
}

// sendClose starts a normal closing handshake and bounds the wait for a reply.
// It reports whether the reader exited, which indicates reader termination,
// not necessarily a successful handshake; the caller must then inspect the
// read error as well.
// All output was written before the handshake starts, so a missing reply is
// logged rather than returned: the server has nothing left to recover.
func (w *Writer) sendClose(conn *websocket.Conn, readDone <-chan struct{}) (bool, error) {
	msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
	err := conn.WriteControl(websocket.CloseMessage, msg, time.Now().Add(writeTimeout))
	if err != nil && !errors.Is(err, websocket.ErrCloseSent) {
		return false, fmt.Errorf("sending websocket close: %w", err)
	}
	timer := time.NewTimer(closeHandshakeTimeout)
	defer timer.Stop()
	select {
	case <-readDone:
		return true, nil
	case <-timer.C:
		w.log.Warn("websocket close reply not received within %s, closing connection", closeHandshakeTimeout)
		return false, nil
	}
}

// closeReadError reports only failures the server should act on. A client
// that leaves cleanly, or vanishes without a Close frame (closed lid, lost
// network, a proxy cutting the socket), is routine browser behavior.
// Gorilla reports an unexpected EOF as CloseAbnormalClosure (1006).
func closeReadError(err error) error {
	if websocket.IsCloseError(err,
		websocket.CloseNormalClosure,
		websocket.CloseGoingAway,
		websocket.CloseNoStatusReceived,
		websocket.CloseAbnormalClosure,
	) || errors.Is(err, syscall.ECONNRESET) {
		return nil
	}
	return fmt.Errorf("reading websocket message: %w", err)
}
