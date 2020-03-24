package events

import (
	"sync"

	"github.com/runatlantis/atlantis/server/logging"
)

type Drainer interface {
	TryAddNewOngoingOperation() bool
	RemoveOngoingOperation()
	StartDrain()
	GetStatus() DrainStatus
}

type SimpleDrainer struct {
	Logger logging.SimpleLogging
	Status DrainStatus
	mutex  sync.Mutex
}

type DrainStatus struct {
	DrainStarted             bool
	DrainCompleted           bool
	OngoingOperationsCounter int
}

// Try to add an operation as ongoing. Return true if the operation is allowed to start, false if it should be rejected.
func (d *SimpleDrainer) TryAddNewOngoingOperation() bool {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if d.Status.DrainStarted {
		return false
	}
	d.Status.OngoingOperationsCounter++
	return true
}

// Consider an operation as completed.
func (d *SimpleDrainer) RemoveOngoingOperation() {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.Status.OngoingOperationsCounter--
	if d.Status.OngoingOperationsCounter < 0 {
		d.Logger.Log(logging.Warn, "Drain OngoingOperationsCounter became below 0, this is a bug")
		d.Status.OngoingOperationsCounter = 0
	}
	if d.Status.DrainStarted && d.Status.OngoingOperationsCounter == 0 {
		d.Status.DrainCompleted = true
	}
}

// Start to drain the server.
func (d *SimpleDrainer) StartDrain() {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.Status.DrainStarted = true
	if d.Status.OngoingOperationsCounter == 0 {
		d.Status.DrainCompleted = true
	}
}

func (d *SimpleDrainer) GetStatus() DrainStatus {
	return d.Status
}
