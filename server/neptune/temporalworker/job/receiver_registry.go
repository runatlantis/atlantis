package job

import (
	"context"
	"sync"
)

type ReceiverRegistry interface {
	AddReceiver(jobID string, ch chan string)
	Broadcast(msg OutputLine)
	CleanUp()

	// Activity context
	Close(ctx context.Context, jobID string)
}

type receiverRegistry struct {
	receivers map[string]map[chan string]bool
	lock      sync.RWMutex
}

func NewReceiverRegistry() *receiverRegistry {
	return &receiverRegistry{
		receivers: map[string]map[chan string]bool{},
	}
}

func (r *receiverRegistry) AddReceiver(jobID string, ch chan string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.receivers[jobID] == nil {
		r.receivers[jobID] = map[chan string]bool{}
	}

	r.receivers[jobID][ch] = true
}

func (r *receiverRegistry) Broadcast(msg OutputLine) {
	for ch := range r.getReceivers(msg.JobID) {
		select {
		case ch <- msg.Line:
		default:
			r.removeReceiver(msg.JobID, ch)
		}
	}
}

// Activity context since it's called from within an activity
func (r *receiverRegistry) Close(ctx context.Context, jobID string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	for ch := range r.receivers[jobID] {
		close(ch)
		delete(r.receivers[jobID], ch)
	}

	delete(r.receivers, jobID)
}

// Called on Shutdown
func (r *receiverRegistry) CleanUp() {
	r.lock.Lock()
	defer r.lock.Unlock()

	for jobId, recvMap := range r.receivers {
		for ch := range recvMap {
			close(ch)
			delete(r.receivers[jobId], ch)
		}
		delete(r.receivers, jobId)
	}
}

func (r *receiverRegistry) getReceivers(jobID string) map[chan string]bool {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.receivers[jobID]
}

func (r *receiverRegistry) removeReceiver(jobID string, ch chan string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	delete(r.receivers[jobID], ch)
}
