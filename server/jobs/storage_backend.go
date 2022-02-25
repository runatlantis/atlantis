package jobs

import (
	"fmt"
	"strings"

	"github.com/graymeta/stow"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/logging"
)

const PageSize = 100

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_storage_backend.go StorageBackend

type StorageBackend interface {
	// Read logs from the storage backend. Must close the reader
	Read(key string) ([]string, error)

	// Write logs to the storage backend
	Write(key string, logs []string) (bool, error)
}

type storageBackend struct {
	location      stow.Location
	logger        logging.SimpleLogging
	containerName string
}

func (s *storageBackend) Read(key string) ([]string, error) {
	return []string{}, nil
}

func (s *storageBackend) Write(key string, logs []string) (bool, error) {
	containerFound := false

	logString := strings.Join(logs, "\n")
	size := int64(len(logString))
	reader := strings.NewReader(logString)

	// Function to write to container
	writeFn := func(container stow.Container, err error) error {
		if err != nil {
			return errors.Wrapf(err, "walking containers at location: %s", s.location)
		}

		// Skip if not right container
		if container.Name() != s.containerName {
			return nil
		}

		containerFound = true
		_, err = container.Put(key, reader, size, nil)
		if err != nil {
			return errors.Wrapf(err, "uploading object for job: %s to %s", key, s.location)
		}

		s.logger.Info("successfully uploaded object for job: %s at location: %s", key, s.location)
		return nil
	}

	s.logger.Info("uploading object for job: %s to container: %s at location: %s", key, s.containerName, s.location)
	err := stow.WalkContainers(s.location, s.containerName, PageSize, writeFn)
	if err != nil {
		return false, err
	}

	if !containerFound {
		return false, fmt.Errorf("container: %s not found at location: %s", s.containerName, s.location)
	}
	return true, nil
}

func NewStorageBackend(jobs valid.Jobs, logger logging.SimpleLogging) (StorageBackend, error) {

	if jobs.StorageBackend == nil {
		return &NoopStorageBackend{}, nil
	}

	config := jobs.StorageBackend.BackendConfig.GetConfigMap()
	backend := jobs.StorageBackend.BackendConfig.GetConfiguredBackend()
	containerName := jobs.StorageBackend.BackendConfig.GetContainerName()

	location, err := stow.Dial(backend, config)
	if err != nil {
		return nil, err
	}

	return &storageBackend{
		location:      location,
		logger:        logger,
		containerName: containerName,
	}, nil
}

// Used when log persistence is not configured
type NoopStorageBackend struct{}

func (s *NoopStorageBackend) Read(key string) ([]string, error) {
	return []string{}, nil
}

func (s *NoopStorageBackend) Write(key string, logs []string) (bool, error) {
	return false, nil
}
