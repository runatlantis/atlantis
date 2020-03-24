package server_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/runatlantis/atlantis/server"
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
			d := &server.DrainController{
				Logger:                   logger,
				DrainStarted:             tt.fields.DrainStarted,
				DrainCompleted:           tt.fields.DrainCompleted,
				OngoingOperationsCounter: tt.fields.OngoingOperationsCounter,
			}
			d.Get(w, r)

			var result server.DrainReponse
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
			d := &server.DrainController{
				Logger:                   logger,
				DrainStarted:             tt.fields.DrainStarted,
				DrainCompleted:           tt.fields.DrainCompleted,
				OngoingOperationsCounter: tt.fields.OngoingOperationsCounter,
			}
			d.Post(w, r)

			var result server.DrainReponse
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

func TestDrainController_TryAddNewOngoingOperation(t *testing.T) {
	type fields struct {
		DrainStarted             bool
		DrainCompleted           bool
		OngoingOperationsCounter int
	}
	type wants struct {
		DrainStarted             bool
		DrainCompleted           bool
		OngoingOperationsCounter int
		Result                   bool
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
				OngoingOperationsCounter: 1,
				Result:                   true,
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
				Result:                   false,
			},
		},
		{
			name: "already completed",
			fields: fields{
				DrainStarted:             true,
				DrainCompleted:           true,
				OngoingOperationsCounter: 1,
			},
			wants: wants{
				DrainStarted:             true,
				DrainCompleted:           true,
				OngoingOperationsCounter: 1,
				Result:                   false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logging.NewNoopLogger()
			d := &server.DrainController{
				Logger:                   logger,
				DrainStarted:             tt.fields.DrainStarted,
				DrainCompleted:           tt.fields.DrainCompleted,
				OngoingOperationsCounter: tt.fields.OngoingOperationsCounter,
			}

			result := d.TryAddNewOngoingOperation()

			t.Helper()
			myTests.Assert(t, tt.wants.Result == result, "exp %d got %d", tt.wants.Result, result)
			myTests.Assert(t, tt.wants.DrainStarted == d.DrainStarted, "exp %s got %s in DrainStarted", tt.wants.DrainStarted, d.DrainStarted)
			myTests.Assert(t, tt.wants.DrainCompleted == d.DrainCompleted, "exp %s got %s in DrainCompleted", tt.wants.DrainCompleted, d.DrainCompleted)
			myTests.Assert(t, tt.wants.OngoingOperationsCounter == d.OngoingOperationsCounter, "exp %s got %s in OngoingOperationsCounter", tt.wants.OngoingOperationsCounter, d.OngoingOperationsCounter)
		})
	}
}

func TestDrainController_RemoveOngoingOperation(t *testing.T) {
	type fields struct {
		DrainStarted             bool
		DrainCompleted           bool
		OngoingOperationsCounter int
	}
	type wants struct {
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
				OngoingOperationsCounter: 1,
			},
			wants: wants{
				DrainStarted:             false,
				DrainCompleted:           false,
				OngoingOperationsCounter: 0,
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
				DrainCompleted:           true,
				OngoingOperationsCounter: 0,
			},
		},
		{
			name: "going negative - not started",
			fields: fields{
				DrainStarted:             false,
				DrainCompleted:           false,
				OngoingOperationsCounter: 0,
			},
			wants: wants{
				DrainStarted:             false,
				DrainCompleted:           false,
				OngoingOperationsCounter: 0,
			},
		},
		{
			name: "going negative - started",
			fields: fields{
				DrainStarted:             true,
				DrainCompleted:           true,
				OngoingOperationsCounter: 0,
			},
			wants: wants{
				DrainStarted:             true,
				DrainCompleted:           true,
				OngoingOperationsCounter: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logging.NewNoopLogger()
			d := &server.DrainController{
				Logger:                   logger,
				DrainStarted:             tt.fields.DrainStarted,
				DrainCompleted:           tt.fields.DrainCompleted,
				OngoingOperationsCounter: tt.fields.OngoingOperationsCounter,
			}

			d.RemoveOngoingOperation()

			t.Helper()
			myTests.Assert(t, tt.wants.DrainStarted == d.DrainStarted, "exp %s got %s in DrainStarted", tt.wants.DrainStarted, d.DrainStarted)
			myTests.Assert(t, tt.wants.DrainCompleted == d.DrainCompleted, "exp %s got %s in DrainCompleted", tt.wants.DrainCompleted, d.DrainCompleted)
			myTests.Assert(t, tt.wants.OngoingOperationsCounter == d.OngoingOperationsCounter, "exp %s got %s in OngoingOperationsCounter", tt.wants.OngoingOperationsCounter, d.OngoingOperationsCounter)
		})
	}
}
