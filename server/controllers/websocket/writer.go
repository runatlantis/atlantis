// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package websocket

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/runatlantis/atlantis/server/logging"
)

const (
	writeTimeout          = 10 * time.Second
	closeHandshakeTimeout = 2 * time.Second
	// maxInboundMessageSize bounds a single data message from the client.
	// Job logs are output-only; the only data a browser sends is what
	// xterm's AttachAddon forwards from keystrokes and small pastes, which
	// the server discards. Control frames are capped at 125 bytes by
	// RFC 6455 5.5 and are not counted against this limit.
	maxInboundMessageSize = 512
	// defaultPongWait is how long the peer may stay silent before the
	// connection is treated as gone. Pings are sent every 9/10 of it, which
	// also keeps idle-timeout proxies such as AWS ALB or nginx (60s by
	// default) from dropping a connection while terraform prints nothing.
	// The values follow Gorilla's chat example.
	defaultPongWait = 60 * time.Second
)

func NewWriter(log logging.SimpleLogging, checkOrigin bool) *Writer {
	upgrader := websocket.Upgrader{
		CheckOrigin: checkOriginFunc(checkOrigin),
	}
	return &Writer{
		upgrader: upgrader,
		log:      log,
		pongWait: defaultPongWait,
	}
}

type Writer struct {
	upgrader websocket.Upgrader
	log      logging.SimpleLogging
	pongWait time.Duration
}

func (w *Writer) Write(rw http.ResponseWriter, r *http.Request, input chan string) error {
	conn, err := w.upgrader.Upgrade(rw, r, nil)

	if err != nil {
		return fmt.Errorf("upgrading websocket connection: %w", err)
	}
	// Gorilla replies Close(1009) and fails the read once a data message
	// exceeds the limit, so an oversized message ends the connection.
	conn.SetReadLimit(maxInboundMessageSize)
	// Browsers answer pings automatically. Every pong extends the read
	// deadline, so a peer that vanished without FIN fails the read instead
	// of holding the connection until the job ends.
	if err := conn.SetReadDeadline(time.Now().Add(w.pongWait)); err != nil {
		_ = conn.Close()
		return fmt.Errorf("setting websocket read deadline: %w", err)
	}
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(w.pongWait))
	})
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
	ticker := time.NewTicker(w.pongWait * 9 / 10)
	defer ticker.Stop()

	for {
		select {
		case <-readDone:
			return closeReadError(readErr)
		case <-ticker.C:
			if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(writeTimeout)); err != nil {
				return resolveWriteError(conn, readDone, &readErr, fmt.Errorf("sending websocket ping: %w", err))
			}
		case msg, ok := <-input:
			if !ok {
				readerDone, err := w.sendClose(conn, readDone)
				if err != nil {
					return resolveWriteError(conn, readDone, &readErr, err)
				}
				if !readerDone {
					return nil // missing Close reply, already logged
				}
				return closeReadError(readErr)
			}
			if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
				return resolveWriteError(conn, readDone, &readErr, fmt.Errorf("setting websocket write deadline: %w", err))
			}
			if err := conn.WriteMessage(websocket.BinaryMessage, []byte("\r"+msg+"\n")); err != nil {
				return resolveWriteError(conn, readDone, &readErr, fmt.Errorf("writing websocket message: %w", err))
			}
		}
	}
}

// resolveWriteError reports what a failed write actually means. When the
// client drops the connection, the reader closes it and a concurrent write
// fails with "use of closed network connection", possibly before readDone
// is closed. Waiting for the reader and preferring its result makes the
// outcome the client's disconnect instead of depending on scheduling.
// readErr is read only after readDone is closed.
func resolveWriteError(conn *websocket.Conn, readDone <-chan struct{}, readErr *error, writeErr error) error {
	// ErrCloseSent: Close was already exchanged on the reader's side, so the
	// reader is finishing on its own. Otherwise close the connection so a
	// reader still blocked on a live peer returns.
	if !errors.Is(writeErr, websocket.ErrCloseSent) {
		_ = conn.Close()
	}
	<-readDone
	readResult := closeReadError(*readErr)
	if errors.Is(writeErr, websocket.ErrCloseSent) || readResult == nil {
		return readResult
	}
	return writeErr
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
// Gorilla reports an unexpected EOF as CloseAbnormalClosure (1006). An
// expired read deadline means the peer stopped answering pings, which is
// the same routine disappearance without a FIN.
func closeReadError(err error) error {
	if websocket.IsCloseError(err,
		websocket.CloseNormalClosure,
		websocket.CloseGoingAway,
		websocket.CloseNoStatusReceived,
		websocket.CloseAbnormalClosure,
	) || errors.Is(err, syscall.ECONNRESET) || isReadTimeout(err) {
		return nil
	}
	return fmt.Errorf("reading websocket message: %w", err)
}

// isReadTimeout reports an expired read deadline. Gorilla replaces timeout
// errors with its own net.Error that does not unwrap to
// os.ErrDeadlineExceeded, so check the interface instead.
func isReadTimeout(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}
