package jobs

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/lyft/feature"
	"github.com/runatlantis/atlantis/server/neptune/storage"
	"github.com/uber-go/tally/v4"
)

const PageSize = 100
const OutputPrefix = "output"

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_storage_backend.go StorageBackend

type StorageBackend interface {
	// Read logs from the storage backend. Must close the reader
	Read(ctx context.Context, key string) ([]string, error)

	// Write logs to the storage backend
	Write(ctx context.Context, key string, logs []string) (bool, error)
}

func NewStorageBackend(client *storage.Client, logger logging.Logger, featureAllocator feature.Allocator, scope tally.Scope) (StorageBackend, error) {
	return &InstrumentedStorageBackend{
		StorageBackend: &storageBackend{
			client: client,
			logger: logger,
		},
		scope: scope.SubScope("storage_backend"),
	}, nil
}

type storageBackend struct {
	client *storage.Client
	logger logging.Logger
}

func (s *storageBackend) Read(ctx context.Context, key string) ([]string, error) {
	s.logger.Info(fmt.Sprintf("reading object for job: %s", key))
	reader, err := s.client.Get(ctx, key)
	if err != nil {
		return nil, errors.Wrap(err, "getting item")
	}
	defer reader.Close()

	buf := new(strings.Builder)
	_, err = io.Copy(buf, reader)
	if err != nil {
		return []string{}, errors.Wrapf(err, "copying to buffer")
	}

	logs := strings.Split(buf.String(), "\n")
	return logs, nil
}

func (s *storageBackend) Write(ctx context.Context, key string, logs []string) (bool, error) {
	logString := strings.Join(logs, "\n")
	object := []byte(logString)

	err := s.client.Set(ctx, key, object)
	if err != nil {
		return false, errors.Wrapf(err, "uploading object for job: %s", key)
	}

	s.logger.Info(fmt.Sprintf("successfully uploaded object for job: %s", key))
	return true, nil
}

// Adds instrumentation to storage backend
type InstrumentedStorageBackend struct {
	StorageBackend

	scope tally.Scope
}

func (i *InstrumentedStorageBackend) Read(ctx context.Context, key string) ([]string, error) {
	failureCount := i.scope.Counter("read_failure")
	latency := i.scope.Timer("read_latency")
	span := latency.Start()
	defer span.Stop()
	logs, err := i.StorageBackend.Read(ctx, key)
	if err != nil {
		failureCount.Inc(1)
	}
	return logs, err
}

func (i *InstrumentedStorageBackend) Write(ctx context.Context, key string, logs []string) (bool, error) {
	failureCount := i.scope.Counter("write_failure")
	successCount := i.scope.Counter("write_success")
	latency := i.scope.Timer("write_latency")
	span := latency.Start()
	defer span.Stop()
	ok, err := i.StorageBackend.Write(ctx, key, logs)
	if err != nil {
		failureCount.Inc(1)
		return ok, err
	}
	successCount.Inc(1)
	return ok, err
}

// Used when log persistence is not configured
type NoopStorageBackend struct{}

func (s *NoopStorageBackend) Read(ctx context.Context, key string) ([]string, error) {
	return []string{}, nil
}

func (s *NoopStorageBackend) Write(ctx context.Context, key string, logs []string) (bool, error) {
	return false, nil
}
