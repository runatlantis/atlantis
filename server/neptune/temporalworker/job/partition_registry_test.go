package job_test

import (
	"context"
	"sync"
	"testing"

	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/neptune/temporalworker/job"
	"github.com/stretchr/testify/assert"
)

type testStore struct {
	t      *testing.T
	JobID  string
	Output string
	Err    error
	Job    job.Job
	Status job.JobStatus
}

func (t *testStore) Get(ctx context.Context, jobID string) (*job.Job, error) {
	assert.Equal(t.t, t.JobID, jobID)
	return &t.Job, t.Err
}

func (t *testStore) Write(ctx context.Context, jobID string, output string) error {
	assert.Equal(t.t, t.JobID, jobID)
	assert.Equal(t.t, t.Output, output)
	return t.Err
}

func (t *testStore) Remove(jobID string) {
	assert.Equal(t.t, t.JobID, jobID)
}

func (t *testStore) Close(ctx context.Context, jobID string, status job.JobStatus) error {
	assert.Equal(t.t, t.JobID, jobID)
	assert.Equal(t.t, t.Status, status)
	return t.Err
}

func (t *testStore) Cleanup(ctx context.Context) error {
	return t.Err
}

type strictTestStore struct {
	t   *testing.T
	get struct {
		runners []*testStore
		count   int
	}
	write struct {
		runners []*testStore
		count   int
	}
	remove struct {
		runners []*testStore
		count   int
	}
	close struct {
		runners []*testStore
		count   int
	}
	cleanup struct {
		runners []*testStore
		count   int
	}
}

func (t strictTestStore) Get(ctx context.Context, jobID string) (*job.Job, error) {
	if t.get.count > len(t.get.runners)-1 {
		t.t.FailNow()
	}
	job, err := t.get.runners[t.get.count].Get(ctx, jobID)
	t.get.count++
	return job, err
}

func (t strictTestStore) Write(ctx context.Context, jobID string, output string) error {
	if t.write.count > len(t.write.runners)-1 {
		t.t.FailNow()
	}
	err := t.write.runners[t.write.count].Write(ctx, jobID, output)
	t.write.count++
	return err
}

func (t strictTestStore) Remove(jobID string) {
	if t.remove.count > len(t.remove.runners)-1 {
		t.t.FailNow()
	}
	t.remove.runners[t.remove.count].Remove(jobID)
	t.remove.count++
}

func (t strictTestStore) Close(ctx context.Context, jobID string, status job.JobStatus) error {
	if t.close.count > len(t.close.runners)-1 {
		t.t.FailNow()
	}
	err := t.close.runners[t.close.count].Close(ctx, jobID, status)
	t.close.count++
	return err
}

func (t strictTestStore) Cleanup(ctx context.Context) error {
	if t.cleanup.count > len(t.cleanup.runners)-1 {
		t.t.FailNow()
	}
	err := t.cleanup.runners[t.cleanup.count].Cleanup(ctx)
	t.cleanup.count++
	return err
}

func TestPartitionRegistry_Register(t *testing.T) {
	logs := []string{"a", "b"}
	jobID := "1234"

	t.Run("streams job output", func(t *testing.T) {
		testStore := &testStore{
			t:     t,
			JobID: jobID,
			Job: job.Job{
				Status: job.Complete,
				Output: logs,
			},
		}
		partitionRegistry := job.PartitionRegistry{
			ReceiverRegistry: &testReceiverRegistry{},
			Store:            testStore,
			Logger:           logging.NewNoopCtxLogger(t),
		}

		buffer := make(chan string, 100)
		go partitionRegistry.Register(context.Background(), jobID, buffer)

		receivedLogs := []string{}
		for line := range buffer {
			receivedLogs = append(receivedLogs, line)
		}

		assert.Equal(t, logs, receivedLogs)

	})

	t.Run("add to receiver registry when job is in progress", func(t *testing.T) {
		buffer := make(chan string)
		testStore := &strictTestStore{
			t: t,
			get: struct {
				runners []*testStore
				count   int
			}{
				runners: []*testStore{
					{
						t:     t,
						JobID: jobID,
						Job: job.Job{
							Status: job.Processing,
							Output: logs,
						},
					},
				},
			},
		}
		receiverRegistry := &strictTestReceiverRegistry{
			t: t,
			addReceiver: struct {
				runners []*testReceiverRegistry
				count   int
			}{
				runners: []*testReceiverRegistry{
					{
						t:     t,
						JobID: jobID,
						Ch:    buffer,
					},
				},
			},
		}

		partitionRegistry := job.PartitionRegistry{
			ReceiverRegistry: receiverRegistry,
			Store:            testStore,
			Logger:           logging.NewNoopCtxLogger(t),
		}

		go func() {
			for range buffer {
			}
		}()
		partitionRegistry.Register(context.Background(), jobID, buffer)
	})

	t.Run("closes receiver after streaming complete job", func(t *testing.T) {
		buffer := make(chan string)
		testStore := &strictTestStore{
			t: t,
			get: struct {
				runners []*testStore
				count   int
			}{
				runners: []*testStore{
					{
						t:     t,
						JobID: jobID,
						Job: job.Job{
							Status: job.Complete,
							Output: logs,
						},
					},
				},
			},
		}
		receiverRegistry := &strictTestReceiverRegistry{
			t: t,
			addReceiver: struct {
				runners []*testReceiverRegistry
				count   int
			}{
				runners: []*testReceiverRegistry{
					{
						t:     t,
						JobID: jobID,
						Ch:    buffer,
					},
				},
			},
		}

		partitionRegistry := job.PartitionRegistry{
			ReceiverRegistry: receiverRegistry,
			Store:            testStore,
			Logger:           logging.NewNoopCtxLogger(t),
		}

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			for range buffer {
			}
			wg.Done()
		}()
		partitionRegistry.Register(context.Background(), jobID, buffer)

		// Read goroutine exits only when the buffer is closed
		wg.Wait()
	})
}
