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
	log    logging.SimpleLogging
}

func newWriterFixture(t *testing.T, opts ...func(*ws.Writer)) *writerFixture {
	t.Helper()
	f := &writerFixture{
		input: make(chan string, 10),
		done:  make(chan struct{}),
		log:   logging.NewNoopLogger(t).WithHistory(),
	}
	f.finish = sync.OnceFunc(func() { close(f.input) })
	w := ws.NewWriter(f.log, false)
	for _, opt := range opts {
		opt(w)
	}
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
	// xterm's AttachAddon forwards keystrokes and small pastes; they are
	// discarded, but must not end the stream.
	Ok(t, f.client.WriteMessage(websocket.TextMessage, []byte(strings.Repeat("x", 512))))
	Ok(t, f.client.WriteControl(websocket.PingMessage, []byte("probe"), time.Now().Add(time.Second)))
	// A client Close follows the Ping on the wire, so the read below must
	// process the Pong before it sees the server's Close response.
	Ok(t, f.client.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(1000, ""), time.Now().Add(time.Second)))
	_, _, err := f.client.ReadMessage()
	Assert(t, websocket.IsCloseError(err, 1000), "expected Close(1000), got %v", err)
	Equals(t, "probe", pong)
	Ok(t, f.await(t))
}

const testPongWait = 200 * time.Millisecond

func withTestPongWait(w *ws.Writer) { w.SetPongWait(testPongWait) }

// A terraform phase with no output must not leave the connection silent:
// idle-timeout load balancers would drop it while the job is still running.
func TestWriterKeepsIdleConnectionAlive(t *testing.T) {
	f := newWriterFixture(t, withTestPongWait)
	var pings atomic.Int32
	f.client.SetPingHandler(func(payload string) error {
		pings.Add(1)
		return f.client.WriteControl(websocket.PongMessage, []byte(payload), time.Now().Add(time.Second))
	})
	// Control frames are only handled while the client reads.
	readErr := make(chan error, 1)
	go func() {
		_, _, err := f.client.ReadMessage()
		readErr <- err
	}()

	select {
	case <-f.done:
		t.Fatalf("writer exited on an idle but responsive connection: %v", f.err)
	case <-time.After(3 * testPongWait):
	}
	Assert(t, pings.Load() >= 2, "expected periodic pings, got %d", pings.Load())

	f.finish()
	Assert(t, websocket.IsCloseError(<-readErr, websocket.CloseNormalClosure), "expected a normal close")
	Ok(t, f.await(t))
}

// A peer that vanishes without FIN never answers pings; the read deadline
// must release the connection instead of holding it until the job ends.
func TestWriterReleasesUnresponsivePeer(t *testing.T) {
	f := newWriterFixture(t, withTestPongWait)
	// The client never reads, so it never sees a ping or sends a pong.
	select {
	case <-f.done:
	case <-time.After(5 * testPongWait):
		t.Fatal("writer kept an unresponsive connection open")
	}
	Ok(t, f.err)
	// The job is still running; the writer must not have consumed or closed input.
	f.input <- "later output"
	Equals(t, 1, len(f.input))
}

// Job logs are output-only, so an oversized inbound message ends the
// connection instead of being read and discarded.
func TestWriterRejectsOversizedInboundMessage(t *testing.T) {
	f := newWriterFixture(t)
	Ok(t, f.client.WriteMessage(websocket.TextMessage, []byte(strings.Repeat("x", 513))))
	_, _, err := f.client.ReadMessage()
	Assert(t, websocket.IsCloseError(err, websocket.CloseMessageTooBig), "expected Close(1009), got %v", err)
	err = f.await(t)
	Assert(t, errors.Is(err, websocket.ErrReadLimit), "expected read limit error, got %v", err)
}

// A browser that loses its network or closes the lid drops TCP without a
// Close frame. That is routine client behavior, not a server failure.
func TestWriterAbruptDisconnect(t *testing.T) {
	f := newWriterFixture(t)
	Ok(t, f.client.Close())
	Ok(t, f.await(t))
}

func TestWriterConnectionReset(t *testing.T) {
	f := newWriterFixture(t)
	tcp, ok := f.client.UnderlyingConn().(*net.TCPConn)
	Assert(t, ok, "expected a TCP connection, got %T", f.client.UnderlyingConn())
	// Linger 0 makes Close send RST, like a proxy or LB cutting the socket.
	Ok(t, tcp.SetLinger(0))
	Ok(t, f.client.Close())
	Ok(t, f.await(t))
}

// Every line was already written when the server starts the closing
// handshake, so a missing reply leaves nothing to act on: warn, don't fail.
func TestWriterCloseTimeout(t *testing.T) {
	f := newWriterFixture(t)
	f.client.SetCloseHandler(func(int, string) error { return nil })
	f.finish()
	_, _, err := f.client.ReadMessage()
	Assert(t, websocket.IsCloseError(err, 1000), "expected Close(1000), got %v", err)
	Ok(t, f.await(t))
	Assert(t, strings.Contains(f.log.GetHistory(), "[WARN] websocket close reply not received"),
		"expected a warning, got log %q", f.log.GetHistory())
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

// The client drops TCP just as the job finishes or prints a line. The
// reader has already closed the connection, so the writer may see both its
// input and readDone ready, or fail a write before readDone is closed. The
// result must be the client's disconnect, never a scheduling-dependent
// "use of closed network connection".
func TestWriterClientDropRacesWithOutput(t *testing.T) {
	for _, mode := range []string{"job done", "data"} {
		t.Run(mode, func(t *testing.T) {
			for range 50 {
				testWriterClientDropRace(t, mode)
			}
		})
	}
}

func testWriterClientDropRace(t *testing.T, mode string) {
	t.Helper()
	conn := &failingConn{closed: make(chan struct{})}
	input := make(chan string, 1)
	done := make(chan error, 1)
	w := ws.NewWriter(logging.NewNoopLogger(t), false)
	s := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		done <- w.Write(failingResponseWriter{rw, conn}, r, input)
	}))
	defer s.Close()
	client, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(s.URL, "http"), nil)
	Ok(t, err)
	Ok(t, client.Close()) // drop TCP without a Close frame
	select {
	case <-conn.closed: // the reader saw the drop and closed the server side
	case <-time.After(5 * time.Second):
		t.Fatal("reader did not close the connection after the client dropped")
	}
	if mode == "data" {
		input <- "output"
	} else {
		close(input)
	}
	select {
	case err := <-done:
		Ok(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("writer did not exit")
	}
}

// NewWriter must honor checkOrigin; NewMultiplexor builds the shipped
// Writer through it, so the writer tests exercise the production upgrader.
func TestWriterCheckOriginRejectsCrossSiteRequest(t *testing.T) {
	w := ws.NewWriter(logging.NewNoopLogger(t), true)
	done := make(chan error, 1)
	s := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		done <- w.Write(rw, r, make(chan string))
	}))
	defer s.Close()
	header := http.Header{"Origin": []string{"https://evil.example"}}
	client, resp, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(s.URL, "http"), header)
	if client != nil {
		_ = client.Close()
	}
	Assert(t, errors.Is(err, websocket.ErrBadHandshake), "expected the handshake to be rejected, got %v", err)
	Equals(t, http.StatusForbidden, resp.StatusCode)
	ErrContains(t, "upgrading websocket connection", <-done)
}
