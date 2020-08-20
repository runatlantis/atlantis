package datadog

import (
	"bytes"
	"io"
	"net"
	"runtime"
	"time"
)

// Handler defines the interface that types must satisfy to process metrics
// received by a dogstatsd server.
type Handler interface {
	// HandleMetric is called when a dogstatsd server receives a metric.
	// The method receives the metric and the address from which it was sent.
	HandleMetric(Metric, net.Addr)

	// HandleEvent is called when a dogstatsd server receives an event.
	// The method receives the metric and the address from which it was sent.
	HandleEvent(Event, net.Addr)
}

// HandlerFunc makes it possible for function types to be used as metric
// handlers on dogstatsd servers.
type HandlerFunc func(Metric, net.Addr)

// HandleMetric calls f(m, a).
func (f HandlerFunc) HandleMetric(m Metric, a net.Addr) {
	f(m, a)
}

// HandleEvent is a no-op for backwards compatibility.
func (f HandlerFunc) HandleEvent(e Event, a net.Addr) {
	return
}

// ListenAndServe starts a new dogstatsd server, listening for UDP datagrams on
// addr and forwarding the metrics to handler.
func ListenAndServe(addr string, handler Handler) (err error) {
	var conn net.PacketConn

	if conn, err = net.ListenPacket("udp", addr); err != nil {
		return
	}

	err = Serve(conn, handler)
	return
}

// Serve runs a dogstatsd server, listening for datagrams on conn and forwarding
// the metrics to handler.
func Serve(conn net.PacketConn, handler Handler) (err error) {
	defer conn.Close()

	concurrency := runtime.GOMAXPROCS(-1)
	if concurrency <= 0 {
		concurrency = 1
	}

	done := make(chan error, concurrency)
	conn.SetDeadline(time.Time{})

	for i := 0; i != concurrency; i++ {
		go serve(conn, handler, done)
	}

	for i := 0; i != concurrency; i++ {
		switch e := <-done; e {
		case nil, io.EOF, io.ErrClosedPipe, io.ErrUnexpectedEOF:
		default:
			err = e
		}
		conn.Close()
	}

	return
}

func serve(conn net.PacketConn, handler Handler, done chan<- error) {
	b := make([]byte, 65536)

	for {
		n, a, err := conn.ReadFrom(b)
		if err != nil {
			done <- err
			return
		}

		for s := b[:n]; len(s) != 0; {
			var ln []byte
			var off int

			if off = bytes.IndexByte(s, '\n'); off < 0 {
				off = len(s)
			} else {
				off++
			}

			ln, s = s[:off], s[off:]

			if bytes.HasPrefix(ln, []byte("_e")) {
				e, err := parseEvent(string(ln))
				if err != nil {
					continue
				}

				handler.HandleEvent(e, a)
			} else {
				m, err := parseMetric(string(ln))
				if err != nil {
					continue
				}

				handler.HandleMetric(m, a)
			}
		}
	}
}
