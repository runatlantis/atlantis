// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package etcd_test

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/core/etcd"
	. "github.com/runatlantis/atlantis/testing"
)

// TestCoordinator_CheckAdmissible_Quarantine proves the non-owner-routed admission
// gate rejects execution while recovery quarantine is active and permits it
// otherwise (design §788).
func TestCoordinator_CheckAdmissible_Quarantine(t *testing.T) {
	backend := startEmbeddedEtcd(t)
	ctx := context.Background()
	rt, err := etcd.NewRuntime(ctx, runtimeConfig(t, backend, "A", "http://127.0.0.1:4142"))
	Ok(t, err)
	t.Cleanup(func() { _ = rt.Close() })
	coord := etcd.NewRuntimeCoordinator(rt)

	Ok(t, coord.CheckAdmissible(ctx))

	Ok(t, rt.Quarantine().Set(ctx, "rec-1", "post-restore"))
	ErrContains(t, "recovery quarantine active", coord.CheckAdmissible(ctx))

	Ok(t, rt.Quarantine().Clear(ctx, "rec-1"))
	Ok(t, coord.CheckAdmissible(ctx))
}

// TestCoordinator_ExecuteUnderQuarantineMarksUncertain proves that if quarantine
// trips between admission and execution, Execute does not run the command and
// advances the admission record to uncertain (terminal) rather than stranding it
// in scheduled, so a later redelivery is not silently short-circuited.
func TestCoordinator_ExecuteUnderQuarantineMarksUncertain(t *testing.T) {
	backend := startEmbeddedEtcd(t)
	ctx := context.Background()
	rt, err := etcd.NewRuntime(ctx, runtimeConfig(t, backend, "A", "http://127.0.0.1:4142"))
	Ok(t, err)
	t.Cleanup(func() { _ = rt.Close() })
	coord := etcd.NewRuntimeCoordinator(rt)

	// Admit a command (reserved -> scheduled) without executing it: the executor
	// only captures the admitted command.
	captured := make(chan etcd.Command, 1)
	rt.AttachExecutor(&inlineExecutor{fn: func(c etcd.Command) { captured <- c }})
	_, err = rt.Route(ctx, cmd("d1"))
	Ok(t, err)

	var admitted etcd.Command
	select {
	case admitted = <-captured:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for admission")
	}

	// Quarantine trips after admission but before execution.
	Ok(t, rt.Quarantine().Set(ctx, "rec-1", "post-restore"))

	ran := false
	outcome := coord.Execute(admitted, func() bool { ran = true; return true })
	Equals(t, etcd.ExecuteUnavailable, outcome)
	Assert(t, !ran, "a quarantined command must not run")

	rec, err := rt.Admission().Get(ctx, admitted.Identity)
	Ok(t, err)
	Assert(t, rec != nil, "admission record should exist")
	Equals(t, etcd.AdmissionUncertain, rec.State())
}

// inlineExecutor runs an arbitrary function on Register, for tests that need to
// drive coordinator.Execute with a custom run closure.
type inlineExecutor struct {
	fn func(cmd etcd.Command)
}

func (e *inlineExecutor) Register(_ context.Context, cmd etcd.Command) error {
	go e.fn(cmd)
	return nil
}

// coordExecutor mimics the real command router's executor: on Register it drives
// the coordinator's Execute in a goroutine, running a recorded closure.
type coordExecutor struct {
	coord   *etcd.RuntimeCoordinator
	mu      sync.Mutex
	ran     map[string]int
	outcome map[string]etcd.ExecuteOutcome
	done    chan string
	success bool
}

func newCoordExecutor(coord *etcd.RuntimeCoordinator, success bool) *coordExecutor {
	return &coordExecutor{
		coord:   coord,
		ran:     map[string]int{},
		outcome: map[string]etcd.ExecuteOutcome{},
		done:    make(chan string, 8),
		success: success,
	}
}

func (e *coordExecutor) Register(_ context.Context, cmd etcd.Command) error {
	go func() {
		outcome := e.coord.Execute(cmd, func() bool {
			e.mu.Lock()
			e.ran[cmd.Identity.DeliveryID]++
			e.mu.Unlock()
			return e.success
		})
		e.mu.Lock()
		e.outcome[cmd.Identity.DeliveryID] = outcome
		e.mu.Unlock()
		e.done <- cmd.Identity.DeliveryID
	}()
	return nil
}

func (e *coordExecutor) wait(t *testing.T) string {
	t.Helper()
	select {
	case d := <-e.done:
		return d
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for coordinator Execute")
		return ""
	}
}

func (e *coordExecutor) runCount(d string) int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.ran[d]
}

func (e *coordExecutor) outcomeFor(d string) etcd.ExecuteOutcome {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.outcome[d]
}

// TestCoordinator_ExecuteRunsAndCompletes proves that a locally-admitted command
// is fenced, run exactly once, and its admission record reaches a terminal state.
func TestCoordinator_ExecuteRunsAndCompletes(t *testing.T) {
	backend := startEmbeddedEtcd(t)
	ctx := context.Background()
	rt, err := etcd.NewRuntime(ctx, runtimeConfig(t, backend, "A", "http://127.0.0.1:4142"))
	Ok(t, err)
	t.Cleanup(func() { _ = rt.Close() })

	coord := etcd.NewRuntimeCoordinator(rt)
	exec := newCoordExecutor(coord, true)
	rt.AttachExecutor(exec)

	res, err := rt.Route(ctx, cmd("d1"))
	Ok(t, err)
	Equals(t, http.StatusAccepted, res.Status)

	d := exec.wait(t)
	Equals(t, "d1", d)
	Equals(t, 1, exec.runCount("d1"))
	Equals(t, etcd.ExecuteRan, exec.outcomeFor("d1"))

	// The admission record should reach succeeded.
	id := cmd("d1").Identity
	deadline := time.Now().Add(5 * time.Second)
	var state etcd.AdmissionState
	for time.Now().Before(deadline) {
		rec, gerr := rt.Admission().Get(ctx, id)
		Ok(t, gerr)
		if rec != nil {
			state = rec.State()
			if state == etcd.AdmissionSucceeded {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	Equals(t, etcd.AdmissionSucceeded, state)
}

// TestCoordinator_ExecuteRecordsFailure proves a run returning false lands the
// admission record in the failed terminal state.
func TestCoordinator_ExecuteRecordsFailure(t *testing.T) {
	backend := startEmbeddedEtcd(t)
	ctx := context.Background()
	rt, err := etcd.NewRuntime(ctx, runtimeConfig(t, backend, "A", "http://127.0.0.1:4142"))
	Ok(t, err)
	t.Cleanup(func() { _ = rt.Close() })

	coord := etcd.NewRuntimeCoordinator(rt)
	exec := newCoordExecutor(coord, false)
	rt.AttachExecutor(exec)

	res, err := rt.Route(ctx, cmd("d1"))
	Ok(t, err)
	Equals(t, http.StatusAccepted, res.Status)
	exec.wait(t)

	id := cmd("d1").Identity
	deadline := time.Now().Add(5 * time.Second)
	var state etcd.AdmissionState
	for time.Now().Before(deadline) {
		rec, gerr := rt.Admission().Get(ctx, id)
		Ok(t, gerr)
		if rec != nil {
			state = rec.State()
			if state == etcd.AdmissionFailed {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	Equals(t, etcd.AdmissionFailed, state)
}

// TestCoordinator_ExecuteBlockedByOlderGeneration proves that an unresolved
// barrier from an earlier owner generation blocks a new generation's execution:
// the run function is never invoked and the outcome is ExecuteBlocked.
func TestCoordinator_ExecuteBlockedByOlderGeneration(t *testing.T) {
	backend := startEmbeddedEtcd(t)
	keys := etcd.NewKeyspace("/atlantis")
	ctx := context.Background()
	epoch, err := etcd.InitOrValidateNamespace(ctx, backend.Client().KV, keys, "dep-1")
	Ok(t, err)

	rt, err := etcd.NewRuntime(ctx, runtimeConfig(t, backend, "A", "http://127.0.0.1:4142"))
	Ok(t, err)
	t.Cleanup(func() { _ = rt.Close() })

	scope := etcd.PullScope{VCSHostname: "github.com", Repository: "o/r", PullNum: 1}

	// Establish an active barrier for an older generation and leave it in place.
	own, err := etcd.NewOwnershipStore(backend, keys, epoch, "old", "http://127.0.0.1:4999", 10*time.Second, 5*time.Second)
	Ok(t, err)
	t.Cleanup(func() { _ = own.Close() })
	oldClaim, won, err := own.Claim(ctx, scope)
	Ok(t, err)
	Assert(t, won, "expected to win the pull claim")
	barriers := etcd.NewExecutionBarrierStore(backend, keys, epoch, own.InstanceID())
	_, err = barriers.StartStep(ctx, oldClaim, "old-exec")
	Ok(t, err)

	// A new generation (different create revision) attempts to execute.
	coord := etcd.NewRuntimeCoordinator(rt)
	newGen := etcd.Generation{Epoch: epoch, CreateRevision: oldClaim.Generation().CreateRevision + 1000}
	blockedCmd := etcd.Command{
		Identity:   etcd.CommandIdentity{SourceKind: "webhook", VCSHostname: "github.com", DeliveryID: "blocked"},
		Scope:      scope,
		Generation: newGen,
	}

	ran := false
	outcome := coord.Execute(blockedCmd, func() bool { ran = true; return true })
	Equals(t, etcd.ExecuteBlocked, outcome)
	Assert(t, !ran, "a blocked execution must not run the command")
}

// TestCoordinator_LeaseLostDuringRunFencesUncertain proves that if the ownership
// lease is lost while a command runs, Execute leaves the barrier in place
// (blocking a new owner generation) and marks the command uncertain rather than
// clearing the fence.
func TestCoordinator_LeaseLostDuringRunFencesUncertain(t *testing.T) {
	backend := startEmbeddedEtcd(t)
	keys := etcd.NewKeyspace("/atlantis")
	ctx := context.Background()
	epoch, err := etcd.InitOrValidateNamespace(ctx, backend.Client().KV, keys, "dep-1")
	Ok(t, err)

	rt, err := etcd.NewRuntime(ctx, runtimeConfig(t, backend, "A", "http://127.0.0.1:4142"))
	Ok(t, err)
	t.Cleanup(func() { _ = rt.Close() })
	coord := etcd.NewRuntimeCoordinator(rt)

	outcomeCh := make(chan etcd.ExecuteOutcome, 1)
	exec := &inlineExecutor{fn: func(cmd etcd.Command) {
		outcomeCh <- coord.Execute(cmd, func() bool {
			// Simulate losing the ownership lease mid-run.
			_ = rt.Ownership().Close()
			select {
			case <-rt.Ownership().Done():
			case <-time.After(5 * time.Second):
			}
			return true
		})
	}}
	rt.AttachExecutor(exec)

	_, err = rt.Route(ctx, cmd("d1"))
	Ok(t, err)

	select {
	case o := <-outcomeCh:
		Equals(t, etcd.ExecuteUncertain, o)
	case <-time.After(15 * time.Second):
		t.Fatal("timed out waiting for Execute")
	}

	// The admission record must be uncertain, never terminal-success.
	rec, err := rt.Admission().Get(ctx, cmd("d1").Identity)
	Ok(t, err)
	Assert(t, rec != nil && rec.State() == etcd.AdmissionUncertain, "admission should be uncertain after lease loss")

	// The barrier remained as a fence: a new owner generation is blocked.
	own2, err := etcd.NewOwnershipStore(backend, keys, epoch, "B", "http://127.0.0.1:4143", 10*time.Second, 5*time.Second)
	Ok(t, err)
	t.Cleanup(func() { _ = own2.Close() })
	scope := etcd.PullScope{VCSHostname: "github.com", Repository: "o/r", PullNum: 1}
	newClaim, won, err := own2.Claim(ctx, scope)
	Ok(t, err)
	Assert(t, won, "B should win the claim after A's lease dropped")
	barriers2 := etcd.NewExecutionBarrierStore(backend, keys, epoch, own2.InstanceID())
	_, err = barriers2.StartStep(ctx, newClaim, "new-exec")
	Assert(t, errors.Is(err, etcd.ErrBlockedByOlderGeneration()), "a new owner must be blocked by the uncertain barrier")
}
