package queue

import (
	"container/list"
)

type Message struct {
	Revision string
}

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

func (q *Queue) Push(msg Message) {
	q.queue.PushBack(msg)
}

func (q *Queue) Peek() Message {
	result := q.queue.Front()
	return result.Value.(Message)
}

func (q *Queue) Pop() Message {
	result := q.queue.Remove(q.queue.Front())

	// naughty casting
	return result.(Message)
}
