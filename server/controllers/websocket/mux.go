// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package websocket

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/runatlantis/atlantis/server/logging"
)

// PartitionKeyGenerator generates partition keys for the multiplexor
type PartitionKeyGenerator interface {
	Generate(r *http.Request) (string, error)
}

// PartitionRegistry is the registry holding each partition
// and is responsible for registering/deregistering new buffers
type PartitionRegistry interface {
	Register(key string, buffer chan string, done <-chan struct{})
	Deregister(key string, buffer chan string)
	IsKeyExists(key string) bool
}

// Multiplexor is responsible for handling the data transfer between the storage layer
// and the registry. Note this is still a WIP as right now the registry is assumed to handle
// everything.
type Multiplexor struct {
	writer       *Writer
	keyGenerator PartitionKeyGenerator
	registry     PartitionRegistry
}

func checkOriginFunc(checkOrigin bool) func(r *http.Request) bool {
	if checkOrigin {
		return nil // use Gorilla websocket's checkSameOrigin
	}
	return func(r *http.Request) bool {
		return true
	}
}

func NewMultiplexor(log logging.SimpleLogging, keyGenerator PartitionKeyGenerator, registry PartitionRegistry, checkOrigin bool) *Multiplexor {
	upgrader := websocket.Upgrader{
		CheckOrigin: checkOriginFunc(checkOrigin),
	}
	return &Multiplexor{
		writer: &Writer{
			upgrader: upgrader,
			log:      log,
		},
		keyGenerator: keyGenerator,
		registry:     registry,
	}
}

// responseCommittedError marks a failure that happened after the Writer took
// over the HTTP response, either by writing the upgrade error itself or by
// hijacking the connection.
type responseCommittedError struct{ err error }

func (e *responseCommittedError) Error() string { return e.err.Error() }
func (e *responseCommittedError) Unwrap() error { return e.err }

// IsResponseCommitted reports whether err came after the HTTP response was
// already written or hijacked, so the caller must not write a response.
func IsResponseCommitted(err error) bool {
	var committed *responseCommittedError
	return errors.As(err, &committed)
}

// Handle should be called for a given websocket request. It blocks
// while writing to the websocket until the buffer is closed.
func (m *Multiplexor) Handle(w http.ResponseWriter, r *http.Request) error {
	key, err := m.keyGenerator.Generate(r)

	if err != nil {
		return fmt.Errorf("generating partition key: %w", err)
	}

	// check if the job ID exists before registering receiver
	if !m.registry.IsKeyExists(key) {
		return fmt.Errorf("invalid key: %s", key)
	}

	// Buffer size set to 1000 to ensure messages get queued.
	// TODO: make buffer size configurable
	buffer := make(chan string, 1000)

	// spinning up a goroutine for this since we are attempting to block on the read side.
	// Write can return before the buffer is drained (for example when the
	// client leaves mid-backfill), so done lets Register stop instead of
	// blocking on a buffer nobody reads. Defers run LIFO: done closes first,
	// then Deregister removes a receiver registered before that.
	done := make(chan struct{})
	go m.registry.Register(key, buffer, done)
	defer m.registry.Deregister(key, buffer)
	defer close(done)

	err = m.writer.Write(w, r, buffer)
	if err != nil {
		return &responseCommittedError{fmt.Errorf("writing to ws %s: %w", key, err)}
	}
	return nil
}
