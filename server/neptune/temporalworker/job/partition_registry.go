package job

import (
	"fmt"

	"github.com/runatlantis/atlantis/server/logging"
)

type receiverAdder interface {
	AddReceiver(jobID string, ch chan string)
}

type storeGetter interface {
	Get(jobID string) (*Job, error)
}

type PartitionRegistry struct {
	ReceiverRegistry receiverAdder
	Store            storeGetter
	Logger           logging.Logger
}

func (p PartitionRegistry) Register(key string, buffer chan string) {
	job, err := p.Store.Get(key)
	if err != nil || job == nil {
		p.Logger.Error(fmt.Sprintf("getting key partition: %s, err: %v", key, err))
		return
	}

	for _, line := range job.Output {
		buffer <- line
	}

	if job.Status == Complete {
		close(buffer)
		return
	}

	p.ReceiverRegistry.AddReceiver(key, buffer)
}
