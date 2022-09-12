package queue

import (
	"container/list"

	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/terraform"
)

// Queue is a standard queue implementation
type Queue struct {
	queue *list.List
}

func NewQueue() *Queue {
	return &Queue{
		queue: list.New(),
	}
}

func (q *Queue) IsEmpty() bool {
	return q.queue.Len() == 0
}

func (q *Queue) Push(msg terraform.DeploymentInfo) {
	q.queue.PushBack(msg)
}

func (q *Queue) Peek() terraform.DeploymentInfo {
	result := q.queue.Front()
	return result.Value.(terraform.DeploymentInfo)
}

func (q *Queue) Pop() terraform.DeploymentInfo {
	result := q.queue.Remove(q.queue.Front())

	// naughty casting
	return result.(terraform.DeploymentInfo)
}
