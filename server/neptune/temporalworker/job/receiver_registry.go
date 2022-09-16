package job

import (
	"context"
	"sync"
)

type receiverRegistry interface {
	AddReceiver(jobID string, ch chan string)
	Broadcast(msg OutputLine)

	// Activity context
	Close(ctx context.Context, jobID string)
}

type ReceiverRegistry struct {
	receivers map[string]map[chan string]bool
	lock      sync.RWMutex
}

func NewReceiverRegistry() *ReceiverRegistry {
	return &ReceiverRegistry{
		receivers: map[string]map[chan string]bool{},
	}
}

func (r *ReceiverRegistry) AddReceiver(jobID string, ch chan string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.receivers[jobID] == nil {
		r.receivers[jobID] = map[chan string]bool{}
	}

	r.receivers[jobID][ch] = true
}

func (r *ReceiverRegistry) Broadcast(msg OutputLine) {
	for ch := range r.getReceivers(msg.JobID) {
		select {
		case ch <- msg.Line:
		default:
			r.removeReceiver(msg.JobID, ch)
		}
	}
}

// Activity context since it's called from within an activity
func (r *ReceiverRegistry) Close(ctx context.Context, jobID string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	for ch := range r.receivers[jobID] {
		close(ch)
		delete(r.receivers[jobID], ch)
	}

	delete(r.receivers, jobID)
}

func (r *ReceiverRegistry) getReceivers(jobID string) map[chan string]bool {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.receivers[jobID]
}

func (r *ReceiverRegistry) removeReceiver(jobID string, ch chan string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	delete(r.receivers[jobID], ch)
}
