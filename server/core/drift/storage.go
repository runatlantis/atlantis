// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package drift

import (
	"sync"
	"time"

	"github.com/runatlantis/atlantis/server/events/models"
)

// Storage defines the interface for drift status persistence.
// Implementations can store drift data in memory, database, or external services.
//
//go:generate pegomock generate --package mocks -o mocks/mock_drift_storage.go Storage
type Storage interface {
	// Store saves a drift result for a project.
	Store(repository string, drift models.ProjectDrift) error

	// Get retrieves drift results for a repository.
	// Optional filters can limit results by project name, path, or workspace.
	Get(repository string, opts GetOptions) ([]models.ProjectDrift, error)

	// Delete removes drift results for a repository.
	// If projectName is empty, all drift results for the repository are removed.
	Delete(repository string, projectName string) error

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
	// MaxAge filters out drift results older than this duration.
	// If zero, no age filtering is applied.
	MaxAge time.Duration
}

// driftKey creates a unique key for a project drift entry.
func driftKey(drift models.ProjectDrift) string {
	return drift.ProjectName + ":" + drift.Path + ":" + drift.Workspace
}

// InMemoryStorage is an in-memory implementation of drift Storage.
// Drift results are lost on server restart.
type InMemoryStorage struct {
	mu sync.RWMutex
	// data maps repository -> (project key -> ProjectDrift)
	data map[string]map[string]models.ProjectDrift
}

// NewInMemoryStorage creates a new in-memory drift storage.
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		data: make(map[string]map[string]models.ProjectDrift),
	}
}

// Store saves a drift result for a project.
func (s *InMemoryStorage) Store(repository string, drift models.ProjectDrift) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.data[repository] == nil {
		s.data[repository] = make(map[string]models.ProjectDrift)
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
		// Apply filters
		if opts.ProjectName != "" && drift.ProjectName != opts.ProjectName {
			continue
		}
		if opts.Path != "" && drift.Path != opts.Path {
			continue
		}
		if opts.Workspace != "" && drift.Workspace != opts.Workspace {
			continue
		}
		if opts.MaxAge > 0 && now.Sub(drift.LastChecked) > opts.MaxAge {
			continue
		}

		result = append(result, drift)
	}

	return result, nil
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
