package revision

import (
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/revision/queue"
	"go.temporal.io/sdk/workflow"
)

type NewRevisionRequest struct {
	Revision string
}

type Queue interface {
	Push(queue.Message)
}

func NewReceiver(ctx workflow.Context, queue Queue) *Receiver {
	return &Receiver{
		queue: queue,
		ctx:   ctx,
	}
}

type Receiver struct {
	queue Queue
	ctx   workflow.Context
}

func (n *Receiver) Receive(c workflow.ReceiveChannel, more bool) {
	// more is false when the channel is closed, so let's just return right away
	if !more {
		return
	}

	var request NewRevisionRequest
	c.Receive(n.ctx, &request)

	n.queue.Push(queue.Message{
		Revision: request.Revision,
	})

}
