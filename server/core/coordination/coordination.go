// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

// Package coordination provides the store for Atlantis coordination
// data: project locks, the global apply lock, and pull status.
package coordination

import (
	"fmt"
	"time"

	"github.com/runatlantis/atlantis/server/core/backends"
	"github.com/runatlantis/atlantis/server/core/coordination/boltdb"
	"github.com/runatlantis/atlantis/server/core/coordination/redis"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

//go:generate go tool mockgen -package mocks -destination mocks/mock_store.go . Store

// LockStore persists project locks.
type LockStore interface {
	TryLock(lock models.ProjectLock) (bool, models.ProjectLock, error)
	Unlock(project models.Project, workspace string) (*models.ProjectLock, error)
	UnlockIfOwnedByPull(project models.Project, workspace string, pullNum int) (*models.ProjectLock, error)
	List() ([]models.ProjectLock, error)
	GetLock(project models.Project, workspace string) (*models.ProjectLock, error)
	UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error)
}

// PullStatusStore persists the plan/apply status of each project in a
// pull request.
type PullStatusStore interface {
	UpdateProjectStatus(pull models.PullRequest, workspace string, repoRelDir string, newStatus models.ProjectPlanStatus) error
	GetPullStatus(pull models.PullRequest) (*models.PullStatus, error)
	DeletePullStatus(pull models.PullRequest) error
	UpdatePullWithResults(pull models.PullRequest, newResults []command.ProjectResult) (models.PullStatus, error)
}

// CommandLockStore persists global command locks (the apply lock).
type CommandLockStore interface {
	LockCommand(cmdName command.Name, lockTime time.Time) (*command.Lock, error)
	UnlockCommand(cmdName command.Name) error
	CheckCommandLock(cmdName command.Name) (*command.Lock, error)
}

// Store persists all coordination data. Consumers should depend on the
// narrowest sub-interface that covers what they touch.
type Store interface {
	LockStore
	PullStatusStore
	CommandLockStore

	// Ping checks the store's connection health.
	Ping() error
	Close() error
}

// NewStore builds the coordination store on the given backend.
func NewStore(logger logging.SimpleLogging, backend backends.Backend) (Store, error) {
	switch b := backend.(type) {
	case *backends.BoltDBBackend:
		logger.Info("using BoltDB coordination store")
		return boltdb.NewStore(b.DB)
	case *backends.RedisBackend:
		logger.Info("using Redis coordination store")
		return redis.NewStore(b.Client)
	default:
		return nil, fmt.Errorf("the coordination store has no %s driver; supported backends: boltdb, redis", backend.Kind())
	}
}
