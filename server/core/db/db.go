// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.
//
// Package db defines the database interface for Atlantis.
package db

import (
	"time"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

//go:generate go tool mockgen -package mocks -destination mocks/mock_database.go . Database

// Database is an implementation of the database API we require.
type Database interface {
	TryLock(lock models.ProjectLock) (bool, models.ProjectLock, error)
	Unlock(project models.Project, workspace string) (*models.ProjectLock, error)
	List() ([]models.ProjectLock, error)
	GetLock(project models.Project, workspace string) (*models.ProjectLock, error)
	UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error)
	UpdateProjectStatus(pull models.PullRequest, workspace string, repoRelDir string, newStatus models.ProjectPlanStatus) error
	GetPullStatus(pull models.PullRequest) (*models.PullStatus, error)
	DeletePullStatus(pull models.PullRequest) error
	UpdatePullWithResults(pull models.PullRequest, newResults []command.ProjectResult) (models.PullStatus, error)

	LockCommand(cmdName command.Name, lockTime time.Time) (*command.Lock, error)
	UnlockCommand(cmdName command.Name) error
	CheckCommandLock(cmdName command.Name) (*command.Lock, error)

	Close() error
}
