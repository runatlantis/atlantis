package events

import (
	"sync"
)

// Drainer is used to gracefully shut down atlantis by waiting for in-progress
// operations to complete.
type Drainer struct {
	status DrainStatus
	mutex  sync.Mutex
	wg     sync.WaitGroup
}

type DrainStatus struct {
	// ShuttingDown is whether we are in the progress of shutting down.
	ShuttingDown bool
	// InProgressOps is the number of operations currently in progress.
	InProgressOps int
}

// StartOp tries to start a new operation. It returns false if Atlantis is
// shutting down.
func (d *Drainer) StartOp() bool {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if d.status.ShuttingDown {
		return false
	}
	d.status.InProgressOps++
	d.wg.Add(1)
	return true
}

// OpDone marks an operation as complete.
func (d *Drainer) OpDone() {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.status.InProgressOps--
	d.wg.Done()
	if d.status.InProgressOps < 0 {
		// This would be a bug.
		d.status.InProgressOps = 0
	}
}

// ShutdownBlocking sets "shutting down" to true and blocks until there are no
// in progress operations.
func (d *Drainer) ShutdownBlocking() {
	// Set the shutdown status.
	d.mutex.Lock()
	d.status.ShuttingDown = true
	d.mutex.Unlock()

	// Block until there are no in-progress ops.
	d.wg.Wait()
}

func (d *Drainer) GetStatus() DrainStatus {
	return d.status
}
