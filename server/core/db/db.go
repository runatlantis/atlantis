// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.
//
// Package db defines the database interface for Atlantis.
package db

import (
	"time"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

//go:generate mockgen -package mocks -destination mocks/mock_database.go . Database

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
