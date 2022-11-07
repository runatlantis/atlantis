package queue

import (
	"testing"

	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/terraform"
	"github.com/stretchr/testify/assert"
)

func TestPriorityQueue(t *testing.T) {
	t.Run("priority order", func(t *testing.T) {
		q := newPriorityQueue()

		q.Push(wrap("1"), Low)
		q.Push(wrap("2"), High)
		q.Push(wrap("3"), High)
		q.Push(wrap("4"), Low)

		assert.Equal(t, 4, q.Size())

		highQueue := q.Scan(High)
		assert.Equal(t, []terraform.DeploymentInfo{
			{
				Revision: "2",
			},
			{
				Revision: "3",
			}}, highQueue)

		lowQueue := q.Scan(Low)
		assert.Equal(t, []terraform.DeploymentInfo{
			{
				Revision: "1",
			},
			{
				Revision: "4",
			}}, lowQueue)

		info, err := q.Pop()
		assert.Equal(t, 3, q.Size())
		assert.NoError(t, err)
		assert.Equal(t, "2", unwrap(info))

		info, err = q.Pop()
		assert.Equal(t, 2, q.Size())
		assert.NoError(t, err)
		assert.Equal(t, "3", unwrap(info))

		info, err = q.Pop()
		assert.Equal(t, 1, q.Size())
		assert.NoError(t, err)
		assert.Equal(t, "1", unwrap(info))

		info, err = q.Pop()
		assert.Equal(t, 0, q.Size())
		assert.NoError(t, err)
		assert.Equal(t, "4", unwrap(info))
	})

	t.Run("pop empty queue", func(t *testing.T) {
		q := newPriorityQueue()
		assert.Equal(t, 0, q.Size())

		_, err := q.Pop()
		assert.Error(t, err)
	})

	priorities := []priorityType{
		Low,
		High,
	}

	for _, p := range priorities {
		t.Run("has items of priority", func(t *testing.T) {
			q := newPriorityQueue()
			assert.False(t, q.HasItemsOfPriority(p))

			q.Push(wrap("1"), p)
			assert.True(t, q.HasItemsOfPriority(p))

		})

		t.Run("empty", func(t *testing.T) {
			q := newPriorityQueue()

			assert.True(t, q.IsEmpty())
			q.Push(wrap("1"), p)
			assert.False(t, q.IsEmpty())
		})
	}
}

func wrap(msg string) terraform.DeploymentInfo {
	return terraform.DeploymentInfo{Revision: msg}

}

func unwrap(msg terraform.DeploymentInfo) string {
	return msg.Revision
}
