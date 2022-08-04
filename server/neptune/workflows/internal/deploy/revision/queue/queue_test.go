package queue_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/revision/queue"
	"github.com/stretchr/testify/assert"
)

func TestQueue(t *testing.T) {
	q := queue.NewQueue()

	assert.True(t, q.IsEmpty())

	q.Push(wrap("1"))
	q.Push(wrap("2"))
	q.Push(wrap("3"))

	assert.False(t, q.IsEmpty())

	assert.Equal(t, "1", unwrap(q.Pop()))
	assert.Equal(t, "2", unwrap(q.Pop()))
	assert.Equal(t, "3", unwrap(q.Pop()))

	assert.True(t, q.IsEmpty())
}

func wrap(msg string) queue.Message {
	return queue.Message{Revision: msg}

}

func unwrap(msg queue.Message) string {
	return msg.Revision
}
