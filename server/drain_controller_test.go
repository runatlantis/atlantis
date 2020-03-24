package server_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/runatlantis/atlantis/server"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/logging"
	myTests "github.com/runatlantis/atlantis/testing"
)

func TestDrainController_Get(t *testing.T) {
	type fields struct {
		DrainStarted             bool
		DrainCompleted           bool
		OngoingOperationsCounter int
	}
	type wants struct {
		Status                   int
		DrainStarted             bool
		DrainCompleted           bool
		OngoingOperationsCounter int
	}
	tests := []struct {
		name   string
		fields fields
		wants  wants
	}{
		{
			name: "simple",
			fields: fields{
				DrainStarted:             false,
				DrainCompleted:           false,
				OngoingOperationsCounter: 0,
			},
			wants: wants{
				DrainStarted:             false,
				DrainCompleted:           false,
				OngoingOperationsCounter: 0,
				Status:                   http.StatusOK,
			},
		},
		{
			name: "on ongoing",
			fields: fields{
				DrainStarted:             false,
				DrainCompleted:           false,
				OngoingOperationsCounter: 1,
			},
			wants: wants{
				DrainStarted:             false,
				DrainCompleted:           false,
				OngoingOperationsCounter: 1,
				Status:                   http.StatusOK,
			},
		},
		{
			name: "started",
			fields: fields{
				DrainStarted:             true,
				DrainCompleted:           false,
				OngoingOperationsCounter: 0,
			},
			wants: wants{
				DrainStarted:             true,
				DrainCompleted:           false,
				OngoingOperationsCounter: 0,
				Status:                   http.StatusOK,
			},
		},
		{
			name: "started and completed",
			fields: fields{
				DrainStarted:             true,
				DrainCompleted:           true,
				OngoingOperationsCounter: 0,
			},
			wants: wants{
				DrainStarted:             true,
				DrainCompleted:           true,
				OngoingOperationsCounter: 0,
				Status:                   http.StatusOK,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logging.NewNoopLogger()
			r, _ := http.NewRequest("GET", "/drain", bytes.NewBuffer(nil))
			w := httptest.NewRecorder()
			dr := &events.Drainer{
				Logger:                   logger,
				DrainStarted:             tt.fields.DrainStarted,
				DrainCompleted:           tt.fields.DrainCompleted,
				OngoingOperationsCounter: tt.fields.OngoingOperationsCounter,
			}
			d := &server.DrainController{
				Logger:  logger,
				Drainer: dr,
			}
			d.Get(w, r)

			var result server.DrainResponse
			t.Helper()
			body, err := ioutil.ReadAll(w.Result().Body)
			myTests.Ok(t, err)
			myTests.Assert(t, tt.wants.Status == w.Result().StatusCode, "exp %d got %d, body: %s", tt.wants.Status, w.Result().StatusCode, string(body))
			err = json.Unmarshal(body, &result)
			myTests.Ok(t, err)
			myTests.Assert(t, tt.wants.DrainStarted == result.DrainStarted, "exp %s got %s in DrainStarted of %s", tt.wants.DrainStarted, result.DrainStarted, string(body))
			myTests.Assert(t, tt.wants.DrainCompleted == result.DrainCompleted, "exp %s got %s in DrainCompleted of %s", tt.wants.DrainCompleted, result.DrainCompleted, string(body))
			myTests.Assert(t, tt.wants.OngoingOperationsCounter == result.OngoingOperationsCounter, "exp %s got %s in OngoingOperationsCounter of %s", tt.wants.OngoingOperationsCounter, result.OngoingOperationsCounter, string(body))
		})
	}
}

func TestDrainController_Post(t *testing.T) {
	type fields struct {
		DrainStarted             bool
		DrainCompleted           bool
		OngoingOperationsCounter int
	}
	type wants struct {
		Status                   int
		DrainStarted             bool
		DrainCompleted           bool
		OngoingOperationsCounter int
	}
	tests := []struct {
		name   string
		fields fields
		wants  wants
	}{
		{
			name: "simple",
			fields: fields{
				DrainStarted:             false,
				DrainCompleted:           false,
				OngoingOperationsCounter: 0,
			},
			wants: wants{
				DrainStarted:             true,
				DrainCompleted:           true,
				OngoingOperationsCounter: 0,
				Status:                   http.StatusCreated,
			},
		},
		{
			name: "on ongoing",
			fields: fields{
				DrainStarted:             false,
				DrainCompleted:           false,
				OngoingOperationsCounter: 1,
			},
			wants: wants{
				DrainStarted:             true,
				DrainCompleted:           false,
				OngoingOperationsCounter: 1,
				Status:                   http.StatusCreated,
			},
		},
		{
			name: "already started",
			fields: fields{
				DrainStarted:             true,
				DrainCompleted:           false,
				OngoingOperationsCounter: 1,
			},
			wants: wants{
				DrainStarted:             true,
				DrainCompleted:           false,
				OngoingOperationsCounter: 1,
				Status:                   http.StatusCreated,
			},
		},
		{
			name: "already started and completed",
			fields: fields{
				DrainStarted:             true,
				DrainCompleted:           true,
				OngoingOperationsCounter: 0,
			},
			wants: wants{
				DrainStarted:             true,
				DrainCompleted:           true,
				OngoingOperationsCounter: 0,
				Status:                   http.StatusCreated,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logging.NewNoopLogger()
			r, _ := http.NewRequest("GET", "/drain", bytes.NewBuffer(nil))
			w := httptest.NewRecorder()
			dr := &events.Drainer{
				Logger:                   logger,
				DrainStarted:             tt.fields.DrainStarted,
				DrainCompleted:           tt.fields.DrainCompleted,
				OngoingOperationsCounter: tt.fields.OngoingOperationsCounter,
			}
			d := &server.DrainController{
				Logger:  logger,
				Drainer: dr,
			}
			d.Post(w, r)

			var result server.DrainResponse
			t.Helper()
			body, err := ioutil.ReadAll(w.Result().Body)
			myTests.Ok(t, err)
			myTests.Assert(t, tt.wants.Status == w.Result().StatusCode, "exp %d got %d, body: %s", tt.wants.Status, w.Result().StatusCode, string(body))
			err = json.Unmarshal(body, &result)
			myTests.Ok(t, err)
			myTests.Assert(t, tt.wants.DrainStarted == result.DrainStarted, "exp %s got %s in DrainStarted of %s", tt.wants.DrainStarted, result.DrainStarted, string(body))
			myTests.Assert(t, tt.wants.DrainCompleted == result.DrainCompleted, "exp %s got %s in DrainCompleted of %s", tt.wants.DrainCompleted, result.DrainCompleted, string(body))
			myTests.Assert(t, tt.wants.OngoingOperationsCounter == result.OngoingOperationsCounter, "exp %s got %s in OngoingOperationsCounter of %s", tt.wants.OngoingOperationsCounter, result.OngoingOperationsCounter, string(body))
		})
	}
}
