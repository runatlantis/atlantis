package jobs

import (
	"context"
	"fmt"
	"sync"

	"github.com/uber-go/tally/v4"

	"github.com/pkg/errors"
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

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_job_store.go JobStore

type JobStore interface {
	// Gets the job from the in memory buffer, if available and if not, reaches to the storage backend
	Get(ctx context.Context, jobID string) (*Job, error)

	// Appends a given string to a job's output if the job is not complete yet
	AppendOutput(jobID string, output string) error

	// Sets a job status to complete and triggers any associated workflow,
	// e.g: if the status is complete, the job is flushed to the associated storage backend
	SetJobCompleteStatus(ctx context.Context, jobID string, status JobStatus) error

	// Removes a job from the store
	RemoveJob(jobID string)
}

func NewJobStore(storageBackend StorageBackend, scope tally.Scope) JobStore {
	return &StorageBackendJobStore{
		JobStore: &InMemoryJobStore{
			jobs: map[string]*Job{},
		},
		storageBackend: storageBackend,
		scope:          scope,
	}
}

// Setup job store for testing
func NewTestJobStore(storageBackend StorageBackend, jobs map[string]*Job) JobStore {
	return &StorageBackendJobStore{
		JobStore: &InMemoryJobStore{
			jobs: jobs,
		},
		storageBackend: storageBackend,
		scope:          tally.NewTestScope("test_jobstore", map[string]string{}),
	}
}

// Memory Job store deals with handling jobs in memory
type InMemoryJobStore struct {
	jobs map[string]*Job
	lock sync.RWMutex
}

func (m *InMemoryJobStore) Get(ctx context.Context, jobID string) (*Job, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	if m.jobs[jobID] == nil {
		return nil, nil
	}
	return m.jobs[jobID], nil
}

func (m *InMemoryJobStore) AppendOutput(jobID string, output string) error {
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

func (m *InMemoryJobStore) SetJobCompleteStatus(ctx context.Context, jobID string, status JobStatus) error {
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

func (m *InMemoryJobStore) RemoveJob(jobID string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	delete(m.jobs, jobID)
}

// Storage backend job store deals with handling jobs in backend storage
type StorageBackendJobStore struct {
	JobStore
	storageBackend StorageBackend
	scope          tally.Scope
}

func (s *StorageBackendJobStore) Get(ctx context.Context, jobID string) (*Job, error) {
	// Get job from memory
	if jobInMem, _ := s.JobStore.Get(ctx, jobID); jobInMem != nil {
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

func (s StorageBackendJobStore) AppendOutput(jobID string, output string) error {
	return s.JobStore.AppendOutput(jobID, output)
}

func (s *StorageBackendJobStore) SetJobCompleteStatus(ctx context.Context, jobID string, status JobStatus) error {
	if err := s.JobStore.SetJobCompleteStatus(ctx, jobID, status); err != nil {
		return err
	}

	job, err := s.JobStore.Get(ctx, jobID)
	if err != nil || job == nil {
		return errors.Wrapf(err, "retrieving job: %s from memory store", jobID)
	}
	subScope := s.scope.SubScope("set_job_complete_status")
	subScope.Counter("write_attempt").Inc(1)
	ok, err := s.storageBackend.Write(ctx, jobID, job.Output)
	if err != nil {
		return errors.Wrapf(err, "persisting job: %s", jobID)
	}

	// Remove from memory if successfully persisted
	if ok {
		s.JobStore.RemoveJob(jobID)
	}
	return nil
}

func (s *StorageBackendJobStore) RemoveJob(jobID string) {
	s.JobStore.RemoveJob(jobID)
}
