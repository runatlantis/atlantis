package jobs

import "sync"

type receiverRegistry interface {
	AddReceiver(jobID string, ch chan string)
	RemoveReceiver(jobID string, ch chan string)
	GetReceivers(jobID string) map[chan string]bool
	CloseAndRemoveReceiversForJob(jobID string)
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

func (r *ReceiverRegistry) RemoveReceiver(jobID string, ch chan string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	delete(r.receivers[jobID], ch)
}

func (r *ReceiverRegistry) GetReceivers(jobID string) map[chan string]bool {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.receivers[jobID]
}

func (r *ReceiverRegistry) CloseAndRemoveReceiversForJob(jobID string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	for ch := range r.receivers[jobID] {
		close(ch)
		delete(r.receivers[jobID], ch)
	}

	delete(r.receivers, jobID)
}
