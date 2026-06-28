// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package drift

import (
	"fmt"
	"sync"
	"time"

	"github.com/runatlantis/atlantis/server/events/models"
)

// Storage defines the interface for drift status persistence.
// Implementations can store drift data in memory, database, or external services.
//
//go:generate go tool pegomock generate --package mocks -o mocks/mock_drift_storage.go Storage
type Storage interface {
	// Store saves a drift result for a project.
	Store(repository string, drift models.ProjectDrift) error

	// Get retrieves drift results for a repository.
	// Optional filters can limit results by project name, path, or workspace.
	Get(repository string, opts GetOptions) ([]models.ProjectDrift, error)

	// Delete removes drift results for a repository.
	// If projectName is empty, all drift results for the repository are removed.
	Delete(repository string, projectName string) error

	// DeleteMatching removes drift results for a repository that match the given
	// filters. At least one filter must be set; use Delete(repository, "") to
	// clear an entire repository.
	DeleteMatching(repository string, opts GetOptions) error

	// GetAll retrieves all stored drift results across all repositories.
	GetAll() (map[string][]models.ProjectDrift, error)
}

// GetOptions defines optional filters for retrieving drift results.
type GetOptions struct {
	// ProjectName filters by project name (exact match).
	ProjectName string
	// Path filters by project path (exact match).
	Path string
	// Workspace filters by Terraform workspace (exact match).
	Workspace string
	// Ref filters by git reference (exact match). Drift records are keyed
	// by ref, so callers that want to act on a specific branch/commit
	// should set this to avoid mixing data across refs.
	Ref string
	// BaseBranch filters by the branch context used for repo-config branch
	// filters and undiverged checks.
	BaseBranch string
	// MaxAge filters out drift results older than this duration.
	// If zero, no age filtering is applied.
	MaxAge time.Duration
	// Exact treats empty project/path/workspace/ref/base_branch fields as exact
	// values instead of wildcards. This is useful for deleting a specific drift
	// record whose identity legitimately contains empty fields.
	Exact bool
}

type driftCacheKey struct {
	ProjectName string
	Path        string
	Workspace   string
	Ref         string
	BaseBranch  string
}

// driftKey creates a unique key for a project drift entry.
// Includes the git ref so the same project on different branches
// does not overwrite each other's drift data.
func driftKey(drift models.ProjectDrift) driftCacheKey {
	return driftCacheKey{
		ProjectName: drift.ProjectName,
		Path:        drift.Path,
		Workspace:   drift.Workspace,
		Ref:         drift.Ref,
		BaseBranch:  drift.BaseBranch,
	}
}

// InMemoryStorage is an in-memory implementation of drift Storage.
// Drift results are lost on server restart.
type InMemoryStorage struct {
	mu sync.RWMutex
	// data maps repository -> (project identity -> ProjectDrift)
	data map[string]map[driftCacheKey]models.ProjectDrift
}

// NewInMemoryStorage creates a new in-memory drift storage.
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		data: make(map[string]map[driftCacheKey]models.ProjectDrift),
	}
}

// Store saves a drift result for a project.
func (s *InMemoryStorage) Store(repository string, drift models.ProjectDrift) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.data[repository] == nil {
		s.data[repository] = make(map[driftCacheKey]models.ProjectDrift)
	}

	key := driftKey(drift)
	s.data[repository][key] = drift
	return nil
}

// Get retrieves drift results for a repository with optional filtering.
func (s *InMemoryStorage) Get(repository string, opts GetOptions) ([]models.ProjectDrift, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	repoData, ok := s.data[repository]
	if !ok {
		return []models.ProjectDrift{}, nil
	}

	now := time.Now()
	result := make([]models.ProjectDrift, 0)

	for _, drift := range repoData {
		if !matchesGetOptions(drift, opts, now) {
			continue
		}

		result = append(result, drift)
	}

	return result, nil
}

func matchesGetOptions(drift models.ProjectDrift, opts GetOptions, now time.Time) bool {
	if opts.ProjectName != "" && drift.ProjectName != opts.ProjectName {
		return false
	}
	if opts.Path != "" && drift.Path != opts.Path {
		return false
	}
	if opts.Workspace != "" && drift.Workspace != opts.Workspace {
		return false
	}
	if opts.Ref != "" && drift.Ref != opts.Ref {
		return false
	}
	if opts.BaseBranch != "" && drift.BaseBranch != opts.BaseBranch {
		return false
	}
	if opts.MaxAge > 0 && now.Sub(drift.LastChecked) > opts.MaxAge {
		return false
	}
	return true
}

// Delete removes drift results for a repository.
func (s *InMemoryStorage) Delete(repository string, projectName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if projectName == "" {
		// Delete all drift results for the repository
		delete(s.data, repository)
		return nil
	}

	// Delete specific project's drift results
	repoData, ok := s.data[repository]
	if !ok {
		return nil
	}

	for key, drift := range repoData {
		if drift.ProjectName == projectName {
			delete(repoData, key)
		}
	}

	return nil
}

// DeleteMatching removes drift results for a repository that match the given filters.
func (s *InMemoryStorage) DeleteMatching(repository string, opts GetOptions) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if opts == (GetOptions{}) {
		return fmt.Errorf("at least one drift delete filter is required")
	}

	repoData, ok := s.data[repository]
	if !ok {
		return nil
	}

	for key, drift := range repoData {
		if matchesDeleteOptions(drift, opts, time.Now()) {
			delete(repoData, key)
		}
	}

	return nil
}

func matchesDeleteOptions(drift models.ProjectDrift, opts GetOptions, now time.Time) bool {
	if !opts.Exact {
		return matchesGetOptions(drift, opts, now)
	}
	if drift.ProjectName != opts.ProjectName {
		return false
	}
	if drift.Path != opts.Path {
		return false
	}
	if drift.Workspace != opts.Workspace {
		return false
	}
	if drift.Ref != opts.Ref {
		return false
	}
	if drift.BaseBranch != opts.BaseBranch {
		return false
	}
	if opts.MaxAge > 0 && now.Sub(drift.LastChecked) > opts.MaxAge {
		return false
	}
	return true
}

// GetAll retrieves all stored drift results across all repositories.
func (s *InMemoryStorage) GetAll() (map[string][]models.ProjectDrift, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string][]models.ProjectDrift)
	for repo, repoData := range s.data {
		drifts := make([]models.ProjectDrift, 0, len(repoData))
		for _, drift := range repoData {
			drifts = append(drifts, drift)
		}
		result[repo] = drifts
	}

	return result, nil
}
