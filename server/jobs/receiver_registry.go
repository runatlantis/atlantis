package jobs

import "sync"

type ReceiverRegistry interface {
	AddReceiver(jobID string, ch chan string)
	RemoveReceiver(jobID string, ch chan string)
	GetReceivers(jobID string) map[chan string]bool
	CloseAndRemoveReceiversForJob(jobID string)
}

type InMemoryReceiverRegistry struct {
	receivers map[string]map[chan string]bool
	lock      sync.RWMutex
}

func NewReceiverRegistry() *InMemoryReceiverRegistry {
	return &InMemoryReceiverRegistry{
		receivers: map[string]map[chan string]bool{},
	}
}

func (r *InMemoryReceiverRegistry) AddReceiver(jobID string, ch chan string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.receivers[jobID] == nil {
		r.receivers[jobID] = map[chan string]bool{}
	}

	r.receivers[jobID][ch] = true
}

func (r *InMemoryReceiverRegistry) RemoveReceiver(jobID string, ch chan string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	delete(r.receivers[jobID], ch)
}

func (r *InMemoryReceiverRegistry) GetReceivers(jobID string) map[chan string]bool {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.receivers[jobID]
}

func (r *InMemoryReceiverRegistry) CloseAndRemoveReceiversForJob(jobID string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	for ch := range r.receivers[jobID] {
		close(ch)
		delete(r.receivers[jobID], ch)
	}

	delete(r.receivers, jobID)
}
