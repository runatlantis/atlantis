// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"sync"
	"testing"
	"testing/synctest"

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
	synctest.Test(t, func(t *testing.T) {
		d := events.Drainer{}
		Assert(t, d.StartOp(), "operation should start before shutdown")
		finishOp := sync.OnceFunc(d.OpDone)
		defer finishOp()

		shutdown := make(chan struct{})
		go func() {
			d.ShutdownBlocking()
			close(shutdown)
		}()

		// Wait until shutdown is blocked on the outstanding operation.
		synctest.Wait()
		Equals(t, false, d.StartOp())
		Equals(t, events.DrainStatus{
			ShuttingDown:  true,
			InProgressOps: 1,
		}, d.GetStatus())
		select {
		case <-shutdown:
			t.Error("shutdown returned before the outstanding operation completed")
		default:
		}

		finishOp()
		<-shutdown
		Equals(t, events.DrainStatus{ShuttingDown: true}, d.GetStatus())
		Equals(t, false, d.StartOp())
	})
}

func TestDrainer_ConcurrentStatus(t *testing.T) {
	var d events.Drainer
	var wg sync.WaitGroup
	start := make(chan struct{})
	const workers = 4
	for range workers {
		wg.Go(func() {
			<-start
			for range 1000 {
				if !d.StartOp() {
					t.Error("operation rejected before shutdown")
					return
				}
				d.OpDone()
			}
		})
	}
	wg.Go(func() {
		<-start
		for range 1000 {
			status := d.GetStatus()
			if status.ShuttingDown || status.InProgressOps < 0 || status.InProgressOps > workers {
				t.Errorf("invalid concurrent status: %+v", status)
			}
		}
	})
	close(start)
	wg.Wait()
	Equals(t, events.DrainStatus{}, d.GetStatus())
}
