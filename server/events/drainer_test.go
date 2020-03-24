package events_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/logging"
	myTests "github.com/runatlantis/atlantis/testing"
)

func TestDrainer_TryAddNewOngoingOperation(t *testing.T) {
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
			d := &events.SimpleDrainer{
				Logger: logger,
				Status: events.DrainStatus{
					DrainStarted:             tt.fields.DrainStarted,
					DrainCompleted:           tt.fields.DrainCompleted,
					OngoingOperationsCounter: tt.fields.OngoingOperationsCounter,
				},
			}

			result := d.TryAddNewOngoingOperation()

			t.Helper()
			myTests.Assert(t, tt.wants.Result == result, "exp %d got %d", tt.wants.Result, result)
			myTests.Assert(t, tt.wants.DrainStarted == d.Status.DrainStarted, "exp %s got %s in DrainStarted", tt.wants.DrainStarted, d.Status.DrainStarted)
			myTests.Assert(t, tt.wants.DrainCompleted == d.Status.DrainCompleted, "exp %s got %s in DrainCompleted", tt.wants.DrainCompleted, d.Status.DrainCompleted)
			myTests.Assert(t, tt.wants.OngoingOperationsCounter == d.Status.OngoingOperationsCounter, "exp %s got %s in OngoingOperationsCounter", tt.wants.OngoingOperationsCounter, d.Status.OngoingOperationsCounter)
		})
	}
}

func TestDrainer_RemoveOngoingOperation(t *testing.T) {
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
			d := &events.SimpleDrainer{
				Logger: logger,
				Status: events.DrainStatus{
					DrainStarted:             tt.fields.DrainStarted,
					DrainCompleted:           tt.fields.DrainCompleted,
					OngoingOperationsCounter: tt.fields.OngoingOperationsCounter,
				},
			}

			d.RemoveOngoingOperation()

			t.Helper()
			myTests.Assert(t, tt.wants.DrainStarted == d.Status.DrainStarted, "exp %s got %s in DrainStarted", tt.wants.DrainStarted, d.Status.DrainStarted)
			myTests.Assert(t, tt.wants.DrainCompleted == d.Status.DrainCompleted, "exp %s got %s in DrainCompleted", tt.wants.DrainCompleted, d.Status.DrainCompleted)
			myTests.Assert(t, tt.wants.OngoingOperationsCounter == d.Status.OngoingOperationsCounter, "exp %s got %s in OngoingOperationsCounter", tt.wants.OngoingOperationsCounter, d.Status.OngoingOperationsCounter)
		})
	}
}

func TestDrainer_StartDrain(t *testing.T) {
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
			name: "simple with no ongoing operation",
			fields: fields{
				DrainStarted:             false,
				DrainCompleted:           false,
				OngoingOperationsCounter: 0,
			},
			wants: wants{
				DrainStarted:             true,
				DrainCompleted:           true,
				OngoingOperationsCounter: 0,
			},
		},
		{
			name: "simple with one ongoing operation",
			fields: fields{
				DrainStarted:             false,
				DrainCompleted:           false,
				OngoingOperationsCounter: 1,
			},
			wants: wants{
				DrainStarted:             true,
				DrainCompleted:           false,
				OngoingOperationsCounter: 1,
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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logging.NewNoopLogger()
			d := &events.SimpleDrainer{
				Logger: logger,
				Status: events.DrainStatus{
					DrainStarted:             tt.fields.DrainStarted,
					DrainCompleted:           tt.fields.DrainCompleted,
					OngoingOperationsCounter: tt.fields.OngoingOperationsCounter,
				},
			}

			d.StartDrain()

			t.Helper()
			myTests.Assert(t, tt.wants.DrainStarted == d.Status.DrainStarted, "exp %s got %s in DrainStarted", tt.wants.DrainStarted, d.Status.DrainStarted)
			myTests.Assert(t, tt.wants.DrainCompleted == d.Status.DrainCompleted, "exp %s got %s in DrainCompleted", tt.wants.DrainCompleted, d.Status.DrainCompleted)
			myTests.Assert(t, tt.wants.OngoingOperationsCounter == d.Status.OngoingOperationsCounter, "exp %s got %s in OngoingOperationsCounter", tt.wants.OngoingOperationsCounter, d.Status.OngoingOperationsCounter)
		})
	}
}

func TestDrainer_GetStatus(t *testing.T) {
	d := &events.SimpleDrainer{
		Status: events.DrainStatus{
			DrainStarted:             true,
			DrainCompleted:           true,
			OngoingOperationsCounter: 12,
		},
	}
	status := d.GetStatus()

	myTests.Assert(t, d.Status.DrainStarted == status.DrainStarted, "exp %s got %s in DrainStarted", d.Status.DrainStarted, status.DrainStarted)
	myTests.Assert(t, d.Status.DrainCompleted == status.DrainCompleted, "exp %s got %s in DrainCompleted", d.Status.DrainCompleted, status.DrainCompleted)
	myTests.Assert(t, d.Status.OngoingOperationsCounter == status.OngoingOperationsCounter, "exp %s got %s in OngoingOperationsCounter", d.Status.OngoingOperationsCounter, status.OngoingOperationsCounter)

}
