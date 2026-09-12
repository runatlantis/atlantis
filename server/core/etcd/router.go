// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.
//
// This file implements the local claim guard and owner-routed dispatch (design
// §"Command dispatch and local fencing", §490). Every replica accepts ingress; a
// command either executes on the local owner or is forwarded once to the owning
// replica. A stale claim returns 409 and the ingress replica refreshes ownership
// and reroutes exactly once; backend or execution unavailability returns 503.
package etcd

import (
	"context"
	"errors"
	"net/http"
)

// Executor idempotently registers an admitted command for local execution under
// a validated owner generation. It must be fast and idempotent; the asynchronous
// run advances the admission record to running and then a terminal state itself.
type Executor interface {
	Register(ctx context.Context, cmd Command) error
}

// Router resolves PR ownership and dispatches a command locally or by forwarding.
type Router struct {
	ownership  *OwnershipStore
	admission  *AdmissionStore
	client     *InternalClient
	executor   Executor
	quarantine *QuarantineStore
}

// NewRouter wires the ownership store, admission store, forwarding client, local
// executor, and recovery-quarantine store into the dispatch guard.
func NewRouter(ownership *OwnershipStore, admission *AdmissionStore, client *InternalClient, executor Executor, quarantine *QuarantineStore) *Router {
	return &Router{ownership: ownership, admission: admission, client: client, executor: executor, quarantine: quarantine}
}

// Route is the ingress entry point. It resolves or creates the PR owner claim
// and either admits locally or forwards to the owning replica (design §490).
func (r *Router) Route(ctx context.Context, cmd Command) Result {
	claim, won, err := r.ownership.Claim(ctx, cmd.Scope)
	if err != nil {
		return result503("ownership backend unavailable")
	}
	if won || claim.Record.InstanceID == r.ownership.InstanceID() {
		cmd.Generation = claim.Generation()
		return r.admitLocal(ctx, cmd, claim)
	}
	return r.forward(ctx, cmd, claim, true)
}

// HandleForwarded is the receiver entry point. It validates that this process
// still owns the exact claim before admitting; otherwise it returns 409 so the
// ingress replica refreshes and reroutes (design §490 steps 3-4).
func (r *Router) HandleForwarded(ctx context.Context, cmd Command) Result {
	owned, err := r.ownership.OwnedLocally(ctx, cmd.Scope, cmd.Generation)
	if err != nil {
		return result503("ownership backend unavailable")
	}
	if !owned {
		return Result{Status: http.StatusConflict, Message: "stale claim"}
	}
	claim, err := r.ownership.Get(ctx, cmd.Scope)
	if err != nil || claim == nil {
		return Result{Status: http.StatusConflict, Message: "claim disappeared"}
	}
	return r.admitLocal(ctx, cmd, *claim)
}

// forward sends the command to the owning replica. On a 409 stale-claim response
// it refreshes ownership and reroutes exactly once (design §501).
func (r *Router) forward(ctx context.Context, cmd Command, owner Claim, allowReroute bool) Result {
	cmd.Generation = owner.Generation()
	res, err := r.client.Send(ctx, owner.AdvertiseURL(), cmd)
	if err != nil {
		return result503("forwarding failed")
	}
	if res.Status == http.StatusConflict && allowReroute {
		// Reroute exactly once: re-resolve ownership with a single claim attempt
		// and make at most one more hop, never resetting the reroute budget
		// (design §501 "reroute once").
		fresh, won, gerr := r.ownership.Claim(ctx, cmd.Scope)
		if gerr != nil {
			return result503("ownership backend unavailable")
		}
		if won || fresh.Record.InstanceID == r.ownership.InstanceID() {
			cmd.Generation = fresh.Generation()
			return r.admitLocal(ctx, cmd, fresh)
		}
		return r.forward(ctx, cmd, fresh, false)
	}
	return res
}

// admitLocal reserves the durable admission record bound to the exact claim,
// idempotently registers the work, and advances the record to scheduled before
// returning 202. Duplicates return the stored state without a second
// registration (design §514, §526).
func (r *Router) admitLocal(ctx context.Context, cmd Command, claim Claim) Result {
	// Recovery quarantine rejects every executable command at admission until an
	// operator reconciles and clears it (design §788, §864). Fail closed on both an
	// active marker and an inability to check it.
	if blocked := r.quarantineBlocks(ctx); blocked != nil {
		return *blocked
	}

	rec, created, err := r.admission.Reserve(ctx, cmd.Identity, claim)
	if errors.Is(err, errStaleClaim) {
		// Ownership moved between resolution and reservation.
		return Result{Status: http.StatusConflict, Message: "stale claim at reservation"}
	}
	if err != nil {
		return result503("admission backend unavailable")
	}

	// Idempotency: a duplicate of an already-advanced record returns its stored
	// state and never creates a second local registration.
	switch rec.State() {
	case AdmissionScheduled, AdmissionRunning, AdmissionSucceeded, AdmissionFailed, AdmissionUncertain:
		return Result{Status: http.StatusAccepted, State: rec.State(), Message: "already admitted"}
	}

	// rec is reserved (fresh, or a duplicate whose prior registration may not
	// have completed). Register idempotently, then advance to scheduled.
	if err := r.executor.Register(ctx, cmd); err != nil {
		// Leave the record reserved for reconciliation rather than duplicating.
		return result503("local execution unavailable")
	}
	scheduled, err := r.admission.Transition(ctx, rec, AdmissionScheduled)
	if err != nil {
		if errors.Is(err, errAdmissionConflict) {
			if cur, gerr := r.admission.Get(ctx, cmd.Identity); gerr == nil && cur != nil {
				return Result{Status: http.StatusAccepted, State: cur.State(), Message: "reconciled"}
			}
		}
		return result503("admission transition failed")
	}
	_ = created
	return Result{Status: http.StatusAccepted, State: scheduled.State(), Message: "admitted"}
}

func result503(msg string) Result {
	return Result{Status: http.StatusServiceUnavailable, Message: msg}
}

// quarantineBlocks returns a fail-closed 503 Result when recovery quarantine is
// active or cannot be checked, and nil otherwise. A nil store (never expected on
// a serving runtime) does not block.
func (r *Router) quarantineBlocks(ctx context.Context) *Result {
	if r.quarantine == nil {
		return nil
	}
	active, err := r.quarantine.IsActive(ctx)
	if err != nil {
		res := result503("recovery quarantine status could not be verified")
		return &res
	}
	if active {
		res := Result{Status: http.StatusServiceUnavailable, Message: "recovery quarantine active; executable commands are rejected until reconciled"}
		return &res
	}
	return nil
}
