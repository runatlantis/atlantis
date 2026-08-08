// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"crypto/subtle"
	"encoding/json"
	"io"
	"net/http"

	"github.com/runatlantis/atlantis/server/core/ownership"
	"github.com/runatlantis/atlantis/server/events"
)

const maxInternalCommandBody = 1 << 20

// InternalCommandController accepts commands only for this process's exact claim.
type InternalCommandController struct {
	Token     string
	ReplicaID string
	Owners    ownership.Store
	Executor  events.CommandExecutor
}

// Comment accepts a forwarded pull request comment command.
func (c *InternalCommandController) Comment(w http.ResponseWriter, r *http.Request) {
	var command events.CommentDispatch
	c.handle(w, r, &command, func() ownership.Key { return command.OwnershipKey() }, func(claimID string) error {
		return c.Executor.ExecuteComment(command, claimID)
	})
}

// Autoplan accepts a forwarded pull request autoplan command.
func (c *InternalCommandController) Autoplan(w http.ResponseWriter, r *http.Request) {
	var command events.AutoplanDispatch
	c.handle(w, r, &command, func() ownership.Key { return command.OwnershipKey() }, func(claimID string) error {
		return c.Executor.ExecuteAutoplan(command, claimID)
	})
}

// PullClosed accepts forwarded owner-local pull request cleanup.
func (c *InternalCommandController) PullClosed(w http.ResponseWriter, r *http.Request) {
	var command events.PullClosedDispatch
	c.handle(w, r, &command, func() ownership.Key { return command.OwnershipKey() }, func(claimID string) error {
		return c.Executor.ExecutePullClosed(command, claimID)
	})
}

func (c *InternalCommandController) handle(
	w http.ResponseWriter,
	r *http.Request,
	command any,
	key func() ownership.Key,
	execute func(string) error,
) {
	providedToken := r.Header.Get(events.InternalCommandTokenHeader)
	if c.Token == "" || subtle.ConstantTimeCompare([]byte(providedToken), []byte(c.Token)) != 1 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxInternalCommandBody)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(command); err != nil {
		http.Error(w, "invalid command payload", http.StatusBadRequest)
		return
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		http.Error(w, "invalid command payload", http.StatusBadRequest)
		return
	}

	if c.Owners == nil || c.Executor == nil {
		http.Error(w, "command routing unavailable", http.StatusServiceUnavailable)
		return
	}
	commandKey := key()
	if _, err := commandKey.Canonical(); err != nil {
		http.Error(w, "invalid command payload", http.StatusBadRequest)
		return
	}
	if err := c.Owners.Ready(r.Context()); err != nil {
		http.Error(w, "command routing unavailable", http.StatusServiceUnavailable)
		return
	}
	current, found, err := c.Owners.Current(r.Context(), commandKey)
	if err != nil {
		http.Error(w, "command routing unavailable", http.StatusServiceUnavailable)
		return
	}
	claimID := r.Header.Get(events.OwnershipClaimHeader)
	if !found || claimID == "" || current.ReplicaID != c.ReplicaID || current.ClaimID != claimID || !c.Owners.Owns(commandKey, claimID) {
		http.Error(w, "pull request ownership changed", http.StatusConflict)
		return
	}
	if err := execute(claimID); err != nil {
		http.Error(w, "command execution unavailable", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
