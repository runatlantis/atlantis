package job

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/neptune/storage"
)

const PageSize = 100

type StorageBackend interface {
	Read(ctx context.Context, key string) ([]string, error)
	Write(ctx context.Context, key string, logs []string) (bool, error)
}

func NewStorageBackend(stowClient *storage.Client, logger logging.Logger) (StorageBackend, error) {
	return &storageBackend{
		client: stowClient,
		logger: logger,
	}, nil
}

type storageBackend struct {
	client *storage.Client
	logger logging.Logger
}

func (s storageBackend) Read(ctx context.Context, key string) (logs []string, err error) {
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

	logs = strings.Split(buf.String(), "\n")
	return
}

// Activity context since it's called from within an activity
func (s storageBackend) Write(ctx context.Context, key string, logs []string) (bool, error) {
	logString := strings.Join(logs, "\n")
	object := []byte(logString)

	err := s.client.Set(ctx, key, object)
	if err != nil {
		return false, errors.Wrapf(err, "uploading object for job: %s", key)
	}

	s.logger.Info(fmt.Sprintf("successfully uploaded object for job: %s", key))
	return true, nil
}

// Used when log persistence is not configured
type NoopStorageBackend struct{}

func (s *NoopStorageBackend) Read(ctx context.Context, key string) ([]string, error) {
	return []string{}, nil
}

func (s *NoopStorageBackend) Write(ctx context.Context, key string, logs []string) (bool, error) {
	return false, nil
}
