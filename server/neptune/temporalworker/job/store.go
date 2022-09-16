package job

import (
	"context"
	"fmt"
	"sync"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/uber-go/tally/v4"
)

type JobStatus int

const (
	Processing JobStatus = iota
	Complete
)

type Job struct {
	Output []string
	Status JobStatus
}

type Store interface {
	Get(jobID string) (*Job, error)
	Write(jobID string, output string) error
	Remove(jobID string)
	Close(ctx context.Context, jobID string, status JobStatus) error
}

func NewStorageBackedStore(config valid.Jobs, logger logging.Logger, scope tally.Scope) (*StorageBackendJobStore, error) {
	storageBackend, err := NewStorageBackend(config, logger, scope)
	if err != nil {
		return nil, errors.Wrapf(err, "initializing storage backend")
	}

	return &StorageBackendJobStore{
		Store: &InMemoryStore{
			jobs: map[string]*Job{},
		},
		storageBackend: storageBackend,
	}, nil
}

func NewTestStorageBackedStore(logger logging.Logger, storageBackend StorageBackend, jobs map[string]*Job) *StorageBackendJobStore {
	return &StorageBackendJobStore{
		Store: &InMemoryStore{
			jobs: jobs,
		},
		storageBackend: storageBackend,
	}
}

// Setup job store for testing
func NewTestJobStore(storageBackend StorageBackend, jobs map[string]*Job) *StorageBackendJobStore {
	return &StorageBackendJobStore{
		Store: &InMemoryStore{
			jobs: jobs,
		},
		storageBackend: storageBackend,
	}
}

// Memory Job store deals with handling jobs in memory
type InMemoryStore struct {
	jobs map[string]*Job
	lock sync.RWMutex
}

func (m *InMemoryStore) Get(jobID string) (*Job, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	if m.jobs[jobID] == nil {
		return nil, nil
	}
	return m.jobs[jobID], nil
}

func (m *InMemoryStore) Write(jobID string, output string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	// Create new job if job dne
	if m.jobs[jobID] == nil {
		m.jobs[jobID] = &Job{}
	}

	if m.jobs[jobID].Status == Complete {
		return fmt.Errorf("cannot append to a complete job")
	}

	updatedOutput := append(m.jobs[jobID].Output, output)
	m.jobs[jobID].Output = updatedOutput
	return nil
}

// Activity context since it's called from within an activity
func (m *InMemoryStore) Close(ctx context.Context, jobID string, status JobStatus) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	// Error out when job dne
	if m.jobs[jobID] == nil {
		return fmt.Errorf("job: %s does not exist", jobID)
	}

	// Error when job is already set to complete
	if job := m.jobs[jobID]; job.Status == Complete {
		return fmt.Errorf("job: %s is already complete", jobID)
	}

	job := m.jobs[jobID]
	job.Status = Complete
	return nil
}

func (m *InMemoryStore) Remove(jobID string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	delete(m.jobs, jobID)
}

// Storage backend job store deals with handling jobs in backend storage
type StorageBackendJobStore struct {
	Store
	storageBackend StorageBackend
}

func (s *StorageBackendJobStore) Get(jobID string) (*Job, error) {
	// Get job from memory
	if jobInMem, _ := s.Store.Get(jobID); jobInMem != nil {
		return jobInMem, nil
	}

	// Get from storage backend if not in memory
	logs, err := s.storageBackend.Read(jobID)
	if err != nil {
		return nil, errors.Wrap(err, "reading from backend storage")
	}

	return &Job{
		Output: logs,
		Status: Complete,
	}, nil
}

func (s StorageBackendJobStore) Write(jobID string, output string) error {
	return s.Store.Write(jobID, output)
}

// Activity context since it's called from within an activity
func (s *StorageBackendJobStore) Close(ctx context.Context, jobID string, status JobStatus) error {
	if err := s.Store.Close(ctx, jobID, status); err != nil {
		return err
	}

	job, err := s.Store.Get(jobID)
	if err != nil || job == nil {
		return errors.Wrapf(err, "retrieving job: %s from memory store", jobID)
	}

	ok, err := s.storageBackend.Write(ctx, jobID, job.Output)
	if err != nil {
		return errors.Wrapf(err, "persisting job: %s", jobID)
	}

	// Remove from memory if successfully persisted
	if ok {
		s.Store.Remove(jobID)
	}
	return nil
}

func (s *StorageBackendJobStore) Remove(jobID string) {
	s.Store.Remove(jobID)
}
