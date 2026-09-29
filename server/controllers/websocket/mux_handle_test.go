// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package websocket_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	ws "github.com/runatlantis/atlantis/server/controllers/websocket"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

type jobKey struct{ err error }

func (k jobKey) Generate(*http.Request) (string, error) { return "job", k.err }

type completedJob struct {
	exists       bool
	deregistered chan struct{}
}

func (j completedJob) IsKeyExists(string) bool { return j.exists }

func (j completedJob) Register(_ string, output chan string) {
	output <- "hook output"
	close(output)
}

func (j completedJob) Deregister(_ string, _ chan string) { close(j.deregistered) }

func TestMultiplexorCompletedJobClose(t *testing.T) {
	registry := completedJob{exists: true, deregistered: make(chan struct{})}
	mux := ws.NewMultiplexor(logging.NewNoopLogger(t), jobKey{}, registry, false)
	done := make(chan error, 1)
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		done <- mux.Handle(rw, r)
	}))
	defer server.Close()
	client, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), nil)
	Ok(t, err)
	defer client.Close()
	Ok(t, client.SetReadDeadline(time.Now().Add(5*time.Second)))
	kind, msg, err := client.ReadMessage()
	Ok(t, err)
	Equals(t, websocket.BinaryMessage, kind)
	Equals(t, "\rhook output\n", string(msg))
	_, _, err = client.ReadMessage()
	Assert(t, websocket.IsCloseError(err, 1000), "expected Close(1000), got %v", err)
	select {
	case err := <-done:
		Ok(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("multiplexor did not finish")
	}
	select {
	case <-registry.deregistered:
	default:
		t.Fatal("receiver was not deregistered")
	}
}

func TestMultiplexorInvalidJob(t *testing.T) {
	for _, tc := range []struct {
		name   string
		keyErr error
		want   string
	}{
		{"unknown job", nil, "invalid key"},
		{"invalid key", errors.New("missing job id"), "generating partition key"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mux := ws.NewMultiplexor(logging.NewNoopLogger(t), jobKey{tc.keyErr}, completedJob{}, false)
			err := mux.Handle(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
			ErrContains(t, tc.want, err)
		})
	}
}
