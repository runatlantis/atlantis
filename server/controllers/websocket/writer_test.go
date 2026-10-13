// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package websocket_test

import (
	"bufio"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	ws "github.com/runatlantis/atlantis/server/controllers/websocket"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

type writerFixture struct {
	client *websocket.Conn
	input  chan string
	done   chan struct{}
	err    error
	finish func()
}

func newWriterFixture(t *testing.T) *writerFixture {
	t.Helper()
	f := &writerFixture{input: make(chan string, 10), done: make(chan struct{})}
	f.finish = sync.OnceFunc(func() { close(f.input) })
	w := ws.NewWriter(logging.NewNoopLogger(t), false)
	s := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		f.err = w.Write(rw, r, f.input)
		close(f.done)
	}))
	t.Cleanup(s.Close)
	var err error
	f.client, _, err = websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(s.URL, "http"), nil)
	Ok(t, err)
	t.Cleanup(func() {
		f.finish()
		_ = f.client.Close()
		select {
		case <-f.done:
		case <-time.After(5 * time.Second):
			t.Error("writer did not exit during cleanup")
		}
	})
	Ok(t, f.client.SetReadDeadline(time.Now().Add(5*time.Second)))
	return f
}

func (f *writerFixture) await(t *testing.T) error {
	t.Helper()
	select {
	case <-f.done:
		return f.err
	case <-time.After(5 * time.Second):
		t.Fatal("writer did not exit")
		return nil
	}
}

func TestWriterServerCloseWaitsForReply(t *testing.T) {
	f := newWriterFixture(t)
	// Delay the reply explicitly to prove that receiving a Close does not
	// immediately tear down the transport on the server.
	f.client.SetCloseHandler(func(int, string) error { return nil })
	f.input <- "plan output"
	f.finish()
	kind, msg, err := f.client.ReadMessage()
	Ok(t, err)
	Equals(t, websocket.BinaryMessage, kind)
	Equals(t, "\rplan output\n", string(msg))
	_, _, err = f.client.ReadMessage()
	Assert(t, websocket.IsCloseError(err, websocket.CloseNormalClosure), "expected Close(1000), got %v", err)
	select {
	case <-f.done:
		t.Fatal("server closed before receiving the peer's Close")
	case <-time.After(50 * time.Millisecond):
	}
	Ok(t, f.client.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(1000, ""), time.Now().Add(time.Second)))
	Ok(t, f.await(t))
}

func TestWriterClientCloseWithIdleInput(t *testing.T) {
	f := newWriterFixture(t)
	Ok(t, f.client.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(1001, "leaving"), time.Now().Add(time.Second)))
	_, _, err := f.client.ReadMessage()
	Assert(t, websocket.IsCloseError(err, 1001), "expected echoed Close(1001), got %v", err)
	Ok(t, f.await(t))
	// The producer is still running; the writer must not close its channel.
	f.input <- "later output"
	Equals(t, 1, len(f.input))
}

func TestWriterClientCloseWithQueuedOutput(t *testing.T) {
	f := newWriterFixture(t)
	for i := 0; i < cap(f.input); i++ {
		f.input <- "queued output"
	}
	Ok(t, f.client.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(1000, ""), time.Now().Add(time.Second)))
	for {
		_, _, err := f.client.ReadMessage()
		if err != nil {
			Assert(t, websocket.IsCloseError(err, 1000), "expected Close(1000), got %v", err)
			break
		}
	}
	Ok(t, f.await(t))
	// No data frames may follow the server's Close response on the wire.
	buf := make([]byte, 1)
	n, err := f.client.UnderlyingConn().Read(buf)
	Equals(t, 0, n)
	Assert(t, errors.Is(err, io.EOF), "expected TCP EOF after Close reply, got %v", err)
}

func TestWriterPingAndInboundData(t *testing.T) {
	f := newWriterFixture(t)
	pong := ""
	f.client.SetPongHandler(func(payload string) error { pong = payload; return nil })
	Ok(t, f.client.WriteMessage(websocket.TextMessage, []byte(strings.Repeat("x", 64*1024))))
	Ok(t, f.client.WriteControl(websocket.PingMessage, []byte("probe"), time.Now().Add(time.Second)))
	// A client Close follows the Ping on the wire, so the read below must
	// process the Pong before it sees the server's Close response.
	Ok(t, f.client.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(1000, ""), time.Now().Add(time.Second)))
	_, _, err := f.client.ReadMessage()
	Assert(t, websocket.IsCloseError(err, 1000), "expected Close(1000), got %v", err)
	Equals(t, "probe", pong)
	Ok(t, f.await(t))
}

func TestWriterAbruptDisconnect(t *testing.T) {
	f := newWriterFixture(t)
	Ok(t, f.client.Close())
	Assert(t, f.await(t) != nil, "transport failure must not be reported as a clean close")
}

func TestWriterCloseTimeout(t *testing.T) {
	f := newWriterFixture(t)
	f.client.SetCloseHandler(func(int, string) error { return nil })
	f.finish()
	_, _, err := f.client.ReadMessage()
	Assert(t, websocket.IsCloseError(err, 1000), "expected Close(1000), got %v", err)
	Assert(t, f.await(t) != nil, "missing Close reply must not be reported as a clean close")
	// Read the underlying transport because Gorilla remembers Close errors.
	buf := make([]byte, 1)
	_, err = f.client.UnderlyingConn().Read(buf)
	Assert(t, errors.Is(err, io.EOF), "expected TCP EOF after handshake timeout, got %v", err)
}

func TestWriterUpgradeFailure(t *testing.T) {
	w := ws.NewWriter(logging.NewNoopLogger(t), false)
	err := w.Write(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil), make(chan string))
	ErrContains(t, "upgrading websocket connection", err)
}

// failingConn injects a transport write failure after the successful upgrade.
type failingConn struct {
	net.Conn
	fail   atomic.Bool
	closed chan struct{}
	once   sync.Once
}

var errInjectedWrite = errors.New("injected transport write failure")

func (c *failingConn) Write(p []byte) (int, error) {
	if c.fail.Load() {
		return 0, errInjectedWrite
	}
	return c.Conn.Write(p)
}

func (c *failingConn) Close() error {
	c.once.Do(func() { close(c.closed) })
	return c.Conn.Close()
}

type failingResponseWriter struct {
	http.ResponseWriter
	conn *failingConn
}

func (w failingResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	c, rw, err := w.ResponseWriter.(http.Hijacker).Hijack()
	if err != nil {
		return nil, nil, err
	}
	w.conn.Conn = c
	return w.conn, rw, nil
}

func TestWriterWriteFailureClosesTransport(t *testing.T) {
	for _, mode := range []string{"data", "server close", "client close"} {
		t.Run(mode, func(t *testing.T) { testWriterWriteFailure(t, mode) })
	}
}

func testWriterWriteFailure(t *testing.T, mode string) {
	t.Helper()
	conn := &failingConn{closed: make(chan struct{})}
	input := make(chan string)
	done := make(chan error, 1)
	w := ws.NewWriter(logging.NewNoopLogger(t), false)
	s := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		done <- w.Write(failingResponseWriter{rw, conn}, r, input)
	}))
	defer s.Close()
	client, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(s.URL, "http"), nil)
	Ok(t, err)
	defer client.Close()
	conn.fail.Store(true)
	switch mode {
	case "data":
		input <- "output"
	case "server close":
		close(input)
	case "client close":
		Ok(t, client.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(1000, ""), time.Now().Add(time.Second)))
	}
	select {
	case err := <-done:
		Assert(t, errors.Is(err, errInjectedWrite), "expected write failure, got %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("writer did not exit after write failure")
	}
	select {
	case <-conn.closed:
	default:
		t.Error("write failure leaked the transport")
		_ = conn.Close()
	}
}
