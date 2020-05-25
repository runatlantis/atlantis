package events_test

import (
	"context"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/events"
	. "github.com/runatlantis/atlantis/testing"
)

// Test starting and completing ops.
func TestDrainer(t *testing.T) {
	d := events.Drainer{}

	// Starts at 0.
	Equals(t, 0, d.GetStatus().InProgressOps)

	// Add 1.
	d.StartOp()
	Equals(t, 1, d.GetStatus().InProgressOps)

	// Remove 1.
	d.OpDone()
	Equals(t, 0, d.GetStatus().InProgressOps)

	// Add 2.
	d.StartOp()
	d.StartOp()
	Equals(t, 2, d.GetStatus().InProgressOps)

	// Remove 1.
	d.OpDone()
	Equals(t, 1, d.GetStatus().InProgressOps)
}

func TestDrainer_Shutdown(t *testing.T) {
	d := events.Drainer{}
	d.StartOp()

	shutdown := make(chan bool)
	go func() {
		d.ShutdownBlocking()
		close(shutdown)
	}()

	// Sleep to ensure that ShutdownBlocking has been called.
	time.Sleep(300 * time.Millisecond)

	// Starting another op should fail.
	Equals(t, false, d.StartOp())

	// Status should be shutting down.
	Equals(t, events.DrainStatus{
		ShuttingDown:  true,
		InProgressOps: 1,
	}, d.GetStatus())

	// Stop the final operation and wait for shutdown to exit.
	d.OpDone()
	timer, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	select {
	case <-shutdown:
	case <-timer.Done():
		Assert(t, false, "Timer reached without shutdown")

	}
}
