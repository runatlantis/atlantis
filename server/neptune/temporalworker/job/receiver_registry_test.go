package job_test

import (
	"context"
	"testing"

	"github.com/runatlantis/atlantis/server/neptune/temporalworker/job"
	"github.com/stretchr/testify/assert"
)

type testReceiverRegistry struct {
	t     *testing.T
	JobID string
	Ch    chan string
	Msg   job.OutputLine
}

func (t *testReceiverRegistry) AddReceiver(jobID string, ch chan string) {
	assert.Equal(t.t, t.JobID, jobID)
	assert.Equal(t.t, t.Ch, ch)
}

func (t *testReceiverRegistry) Broadcast(msg job.OutputLine) {
	assert.Equal(t.t, t.Msg, msg)
}

func (t *testReceiverRegistry) Close(ctx context.Context, jobID string) {
}

func (t *testReceiverRegistry) CleanUp() {
}

type strictTestReceiverRegistry struct {
	t           *testing.T
	addReceiver struct {
		runners []*testReceiverRegistry
		count   int
	}
	broadcast struct {
		runners []*testReceiverRegistry
		count   int
	}
	close struct {
		runners []*testReceiverRegistry
		count   int
	}
	cleanup struct {
		runners []*testReceiverRegistry
		count   int
	}
}

func (t strictTestReceiverRegistry) AddReceiver(jobID string, ch chan string) {
	if t.addReceiver.count > len(t.addReceiver.runners)-1 {
		t.t.FailNow()
	}
	t.addReceiver.runners[t.addReceiver.count].AddReceiver(jobID, ch)
	t.addReceiver.count++
}

func (t strictTestReceiverRegistry) Broadcast(msg job.OutputLine) {
	if t.broadcast.count > len(t.broadcast.runners)-1 {
		t.t.FailNow()
	}
	t.broadcast.runners[t.broadcast.count].Broadcast(msg)
	t.broadcast.count++
}

func (t strictTestReceiverRegistry) Close(ctx context.Context, jobID string) {
	if t.close.count > len(t.close.runners)-1 {
		t.t.FailNow()
	}
	t.close.runners[t.close.count].Close(ctx, jobID)
	t.close.count++
}

func (t *strictTestReceiverRegistry) CleanUp() {
	if t.cleanup.count > len(t.cleanup.runners)-1 {
		t.t.FailNow()
	}
	t.cleanup.runners[t.cleanup.count].CleanUp()
	t.cleanup.count++
}

func TestReceiverRegistry(t *testing.T) {
	jobID := "1234"
	outputMsg := "a"

	t.Run("adds a receiver and broadcast", func(t *testing.T) {
		recvRegistry := job.NewReceiverRegistry()

		ch := make(chan string)
		recvRegistry.AddReceiver(jobID, ch)

		go recvRegistry.Broadcast(job.OutputLine{
			JobID: jobID,
			Line:  outputMsg,
		})

		assert.Equal(t, outputMsg, <-ch)
	})

	t.Run("removes receiver when close", func(t *testing.T) {
		recvRegistry := job.NewReceiverRegistry()

		ch := make(chan string)
		recvRegistry.AddReceiver(jobID, ch)

		recvRegistry.Close(context.TODO(), jobID)

		for range ch {
		}
	})

	t.Run("removes receiver if blocking", func(t *testing.T) {
		recvRegistry := job.NewReceiverRegistry()

		ch := make(chan string)
		recvRegistry.AddReceiver(jobID, ch)

		// this call is blocking if receiver is not removed since we are not listening to it
		recvRegistry.Broadcast(job.OutputLine{
			JobID: jobID,
			Line:  outputMsg,
		})
	})
}
