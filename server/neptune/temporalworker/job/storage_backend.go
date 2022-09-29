package job

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/graymeta/stow"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/uber-go/tally/v4"
)

const OutputPrefix = "output"
const PageSize = 100

type StorageBackend interface {
	Read(key string) ([]string, error)
	Write(ctx context.Context, key string, logs []string) (bool, error)
}

func NewStorageBackend(jobs valid.Jobs, logger logging.Logger, scope tally.Scope) (StorageBackend, error) {
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

	storageBackend := &storageBackend{
		location:      location,
		containerName: containerName,
		logger:        logger,
	}

	return &InstrumentedStorageBackend{
		StorageBackend: storageBackend,
		scope:          scope.SubScope("storage_backend"),
	}, nil
}

type storageBackend struct {
	location      stow.Location
	containerName string
	logger        logging.Logger
}

func (s storageBackend) Read(key string) (logs []string, err error) {
	s.logger.Info(fmt.Sprintf("reading object for job: %s in container: %s", key, s.containerName))

	container, err := s.location.Container(s.containerName)
	if err != nil {
		return []string{}, errors.Wrap(err, "resolving container")
	}

	key = fmt.Sprintf("%s/%s", OutputPrefix, key)
	item, err := container.Item(key)
	if err != nil {
		return []string{}, errors.Wrap(err, "getting item")
	}

	r, err := item.Open()
	if err != nil {
		return []string{}, errors.Wrap(err, "reading item")
	}

	buf := new(strings.Builder)
	_, err = io.Copy(buf, r)
	if err != nil {
		return []string{}, errors.Wrapf(err, "copying to buffer")
	}

	logs = strings.Split(buf.String(), "\n")
	return
}

// Activity context since it's called from within an activity
func (s storageBackend) Write(ctx context.Context, key string, logs []string) (bool, error) {
	container, err := s.location.Container(s.containerName)
	if err != nil {
		return false, errors.Wrap(err, "resolving container")
	}

	logString := strings.Join(logs, "\n")
	size := int64(len(logString))
	reader := strings.NewReader(logString)

	key = fmt.Sprintf("%s/%s", OutputPrefix, key)
	_, err = container.Put(key, reader, size, nil)
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

func (i *InstrumentedStorageBackend) Read(key string) ([]string, error) {
	failureCount := i.scope.Counter("read_failure")
	latency := i.scope.Timer("read_latency")
	span := latency.Start()
	defer span.Stop()
	logs, err := i.StorageBackend.Read(key)
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

func (s *NoopStorageBackend) Read(key string) ([]string, error) {
	return []string{}, nil
}

func (s *NoopStorageBackend) Write(ctx context.Context, key string, logs []string) (bool, error) {
	return false, nil
}
