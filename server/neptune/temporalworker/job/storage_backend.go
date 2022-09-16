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
	"github.com/runatlantis/atlantis/server/neptune/logger"
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

	return &InstrumenetedStorageBackend{
		StorageBackend: storageBackend,
		readFailures:   scope.SubScope("storage_backend").Counter("read_failure"),
		writeFailures:  scope.SubScope("storage_backend").Counter("write_failure"),
		writeSuccesses: scope.SubScope("storage_backend").Counter("write_success"),
	}, nil
}

type storageBackend struct {
	location      stow.Location
	containerName string
	logger        logging.Logger
}

func (s storageBackend) Read(key string) (logs []string, err error) {

	// Read from  /output directory
	key = fmt.Sprintf("%s/%s", OutputPrefix, key)
	readContainerFn := func(item stow.Item, err error) error {
		if err != nil {
			return errors.Wrapf(err, "reading item: %s", item.Name())
		}

		// Skip if not right item
		if item.Name() != key {
			return nil
		}

		r, err := item.Open()
		if err != nil {
			return errors.Wrapf(err, "building reader for item: %s", item.Name())
		}

		buf := new(strings.Builder)
		_, err = io.Copy(buf, r)
		if err != nil {
			return errors.Wrapf(err, "building buffer for item: %s", item.Name())
		}

		logs = strings.Split(buf.String(), "\n")
		return nil
	}

	readLocationFn := func(container stow.Container, err error) error {
		if err != nil {
			return errors.Wrap(err, "reading containers")
		}

		// Skip if not right container
		if container.Name() != s.containerName {
			return nil
		}

		return stow.Walk(container, key, PageSize, readContainerFn)
	}

	s.logger.Info(fmt.Sprintf("reading object for job: %s in container: %s", key, s.containerName))
	err = stow.WalkContainers(s.location, s.containerName, PageSize, readLocationFn)
	return
}

// Activity context since it's called from within an activity
func (s storageBackend) Write(ctx context.Context, key string, logs []string) (bool, error) {
	// Write to /output directory
	key = fmt.Sprintf("%s/%s", OutputPrefix, key)

	containerFound := false
	logString := strings.Join(logs, "\n")
	size := int64(len(logString))
	reader := strings.NewReader(logString)

	// Function to write to container
	writeFn := func(container stow.Container, err error) error {
		if err != nil {
			return errors.Wrap(err, "walking containers")
		}

		// Skip if not right container
		if container.Name() != s.containerName {
			return nil
		}

		containerFound = true
		_, err = container.Put(key, reader, size, nil)
		if err != nil {
			return errors.Wrapf(err, "uploading object for job: %s", key)
		}

		logger.Info(ctx, fmt.Sprintf("successfully uploaded object for job: %s", key))
		return nil
	}

	logger.Info(ctx, fmt.Sprintf("uploading object for job: %s to container: %s", key, s.containerName))
	err := stow.WalkContainers(s.location, s.containerName, PageSize, writeFn)
	if err != nil {
		return false, err
	}

	if !containerFound {
		return false, fmt.Errorf("container: %s not found", s.containerName)
	}
	return true, nil
}

// Adds instrumentation to storage backend
type InstrumenetedStorageBackend struct {
	StorageBackend

	readFailures   tally.Counter
	writeFailures  tally.Counter
	writeSuccesses tally.Counter
}

func (i *InstrumenetedStorageBackend) Read(key string) ([]string, error) {
	logs, err := i.StorageBackend.Read(key)
	if err != nil {
		i.readFailures.Inc(1)
	}
	return logs, err
}

func (i *InstrumenetedStorageBackend) Write(ctx context.Context, key string, logs []string) (bool, error) {
	ok, err := i.StorageBackend.Write(ctx, key, logs)
	if err != nil {
		i.writeFailures.Inc(1)
		return ok, err
	}
	i.writeSuccesses.Inc(1)
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
