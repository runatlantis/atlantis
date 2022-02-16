package jobs

import (
	"github.com/runatlantis/atlantis/server/core/config/valid"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_storage_backend.go StorageBackend

type StorageBackend interface {
	// Read logs from the storage backend. Must close the reader
	Read(key string) ([]string, error)

	// Write logs to the storage backend
	Write(key string, logs []string) (success bool, err error)
}

func NewStorageBackend(jobs valid.Jobs) StorageBackend {
	// No storage backend configured, return Noop for now
	return &NoopStorageBackend{}
}

// Used when log persistence is not configured
type NoopStorageBackend struct{}

func (s *NoopStorageBackend) Read(key string) ([]string, error) {
	return []string{}, nil
}

func (s *NoopStorageBackend) Write(key string, logs []string) (success bool, err error) {
	return false, nil
}
