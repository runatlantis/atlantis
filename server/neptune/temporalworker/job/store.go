package job

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/neptune/storage"
	"github.com/uber-go/tally/v4"
)

type JobStatus int //nolint:revive // avoiding refactor while adding linter action

const (
	Processing JobStatus = iota
	Complete
)

type Job struct {
	Output []string
	Status JobStatus
}

type Store interface {
	Get(ctx context.Context, jobID string) (*Job, error)
	Write(ctx context.Context, jobID string, output string) error
	Remove(jobID string)
	Close(ctx context.Context, jobID string, status JobStatus) error
	Cleanup(ctx context.Context) error
}

func NewStorageBackendStore(jobStoreConfig valid.StoreConfig, scope tally.Scope, logger logging.Logger) (*StorageBackendJobStore, error) {
	stowClient, err := storage.NewClient(jobStoreConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "initializing stow client")
	}

	storageBackend, err := NewStorageBackend(stowClient, scope, logger)
	if err != nil {
		return nil, errors.Wrapf(err, "initializing storage backend")
	}

	return &StorageBackendJobStore{
		InMemoryStore: &InMemoryStore{
			jobs: map[string]*Job{},
		},
		storageBackend: storageBackend,
		logger:         logger,
	}, nil
}

func NewTestStorageBackedStore(logger logging.Logger, storageBackend StorageBackend, jobs map[string]*Job) *StorageBackendJobStore {
	return &StorageBackendJobStore{
		InMemoryStore: &InMemoryStore{
			jobs: jobs,
		},
		storageBackend: storageBackend,
		logger:         logger,
	}
}

// Setup job store for testing
func NewTestJobStore(storageBackend StorageBackend, jobs map[string]*Job) *StorageBackendJobStore {
	return &StorageBackendJobStore{
		InMemoryStore: &InMemoryStore{
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

func (m *InMemoryStore) Get(ctx context.Context, jobID string) (*Job, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	if _, ok := m.jobs[jobID]; !ok {
		return nil, nil
	}
	return m.jobs[jobID], nil
}

func (m *InMemoryStore) Write(ctx context.Context, jobID string, output string) error {
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

	// No need to close job if it DNE
	// It is possible a job DNE in memory since we call CloseJob() in a separate activity which can be executed after a worker restart
	if _, ok := m.jobs[jobID]; !ok {
		return nil
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

func (m *InMemoryStore) GetJobs() map[string]*Job {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return m.jobs
}

// Storage backend job store deals with handling jobs in backend storage
type StorageBackendJobStore struct {
	*InMemoryStore
	storageBackend StorageBackend
	logger         logging.Logger
}

func (s *StorageBackendJobStore) Get(ctx context.Context, jobID string) (*Job, error) {
	// Get job from memory
	if jobInMem, _ := s.InMemoryStore.Get(ctx, jobID); jobInMem != nil {
		return jobInMem, nil
	}

	// Get from storage backend if not in memory
	logs, err := s.storageBackend.Read(ctx, jobID)
	if err != nil {
		return nil, errors.Wrap(err, "reading from backend storage")
	}

	return &Job{
		Output: logs,
		Status: Complete,
	}, nil
}

func (s StorageBackendJobStore) Write(ctx context.Context, jobID string, output string) error {
	return s.InMemoryStore.Write(ctx, jobID, output)
}

// Activity context since it's called from within an activity
func (s *StorageBackendJobStore) Close(ctx context.Context, jobID string, status JobStatus) error {
	if err := s.InMemoryStore.Close(ctx, jobID, status); err != nil {
		return err
	}

	job, err := s.InMemoryStore.Get(ctx, jobID)
	if err != nil {
		return errors.Wrapf(err, "retrieving job: %s from memory store", jobID)
	}

	// Since we close the job in a different activity than when it's created, it is possible that we try closing a non existent job
	// after the worker has been restarted. So, instead of hard failing, let's return since we close all in progress jobs  during shutdown
	if job == nil {
		s.logger.WarnContext(ctx, fmt.Sprintf("job: %s does not exist in memory", jobID))
		return nil
	}

	ok, err := s.storageBackend.Write(ctx, jobID, job.Output)
	if err != nil {
		return errors.Wrapf(err, "persisting job: %s", jobID)
	}

	// Remove from memory if successfully persisted
	if ok {
		s.InMemoryStore.Remove(jobID)
	}
	return nil
}

func (s *StorageBackendJobStore) Remove(jobID string) {
	s.InMemoryStore.Remove(jobID)
}

// Persist all jobs in memory
func (s *StorageBackendJobStore) Cleanup(ctx context.Context) error {
	failedJobs := []string{}
	for jobID, job := range s.InMemoryStore.GetJobs() {
		_, err := s.storageBackend.Write(ctx, jobID, job.Output)

		// Track failed jobs, log errors and continue with other jobs
		if err != nil {
			s.logger.ErrorContext(ctx, fmt.Sprintf("failed to persist job %s on cleanup: %s", jobID, err))
			failedJobs = append(failedJobs, jobID)
		}
	}

	if len(failedJobs) > 0 {
		return errors.Errorf("failed to persist jobs: %s", strings.Join(failedJobs, ","))
	}
	return nil
}
