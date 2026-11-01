// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/runatlantis/atlantis/server/controllers"
	ws "github.com/runatlantis/atlantis/server/controllers/websocket"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
	tally "github.com/uber-go/tally/v4"
)

type staticJobKey struct{}

func (staticJobKey) Generate(*http.Request) (string, error) { return "job", nil }

// finishedJobRegistry behaves like a job whose output has already completed.
type finishedJobRegistry struct{ exists bool }

func (r finishedJobRegistry) IsKeyExists(string) bool                            { return r.exists }
func (finishedJobRegistry) Register(_ string, ch chan string, _ <-chan struct{}) { close(ch) }
func (finishedJobRegistry) Deregister(string, chan string)                       {}

func newJobsWSController(t *testing.T, jobExists bool) *controllers.JobsController {
	t.Helper()
	registry := finishedJobRegistry{exists: jobExists}
	return &controllers.JobsController{
		Logger:     logging.NewNoopLogger(t),
		WsMux:      ws.NewMultiplexor(logging.NewNoopLogger(t), staticJobKey{}, registry, false),
		StatsScope: tally.NewTestScope("test", nil),
	}
}

func TestGetProjectJobsWS_UnknownJobResponds500(t *testing.T) {
	rec := httptest.NewRecorder()
	newJobsWSController(t, false).GetProjectJobsWS(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	Equals(t, http.StatusInternalServerError, rec.Code)
	Assert(t, strings.Contains(rec.Body.String(), "invalid key"), "body %q", rec.Body.String())
}

// The upgrader writes its own 400 response; the controller must not append a
// second response to it.
func TestGetProjectJobsWS_UpgradeFailureDoesNotRespondAgain(t *testing.T) {
	rec := httptest.NewRecorder()
	newJobsWSController(t, true).GetProjectJobsWS(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	Equals(t, http.StatusBadRequest, rec.Code)
	Assert(t, !strings.Contains(rec.Body.String(), "upgrading websocket connection"),
		"controller wrote a second response: %q", rec.Body.String())
}

// After the upgrade the connection is hijacked, so writing an HTTP response
// only makes net/http log "response.WriteHeader on hijacked connection".
func TestGetProjectJobsWS_ErrorAfterUpgradeDoesNotTouchHijackedWriter(t *testing.T) {
	var serverLog bytes.Buffer
	done := make(chan struct{})
	c := newJobsWSController(t, true)
	scope := c.StatsScope.(tally.TestScope)
	s := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer close(done)
		c.GetProjectJobsWS(w, r)
	}))
	s.Config.ErrorLog = log.New(&serverLog, "", 0)
	s.Start()
	defer s.Close()

	client, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(s.URL, "http"), nil)
	Ok(t, err)
	defer client.Close()
	// Clients must mask their frames (RFC 6455 5.1). An unmasked frame is a
	// protocol error, so Handle fails after the connection was hijacked.
	_, err = client.UnderlyingConn().Write([]byte{0x82, 0x01, 'x'})
	Ok(t, err)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("handler did not return")
	}
	Equals(t, int64(1), scope.Snapshot().Counters()["test.getprojectjobs.execution_error+"].Value())
	Assert(t, !strings.Contains(serverLog.String(), "hijacked"), "server log: %q", serverLog.String())
}
