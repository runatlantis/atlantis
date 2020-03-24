package events

import (
	"sync"

	"github.com/runatlantis/atlantis/server/logging"
)

type Drainer struct {
	Logger                   logging.SimpleLogging
	DrainStarted             bool
	DrainCompleted           bool
	OngoingOperationsCounter int
	mutex                    sync.Mutex
}

// Try to add an operation as ongoing. Return true if the operation is allowed to start, false if it should be rejected.
func (d *Drainer) TryAddNewOngoingOperation() bool {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if d.DrainStarted {
		return false
	}
	d.OngoingOperationsCounter++
	return true
}

// Consider an operation as completed.
func (d *Drainer) RemoveOngoingOperation() {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.OngoingOperationsCounter--
	if d.OngoingOperationsCounter < 0 {
		d.Logger.Log(logging.Warn, "Drain OngoingOperationsCounter became below 0, this is a bug")
		d.OngoingOperationsCounter = 0
	}
	if d.DrainStarted && d.OngoingOperationsCounter == 0 {
		d.DrainCompleted = true
	}
}

// Start to drain the server.
func (d *Drainer) StartDrain() {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.DrainStarted = true
	if d.OngoingOperationsCounter == 0 {
		d.DrainCompleted = true
	}
}
