package queue_test

import (
	"testing"

	activity "github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/revision/queue"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/test"
	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/workflow"
)

func noopCallback(ctx workflow.Context, q *queue.Deploy) {}

func TestQueue(t *testing.T) {
	t.Run("priority", func(t *testing.T) {
		q := queue.NewQueue(nil)

		msg1 := wrap("1", activity.MergeTrigger)
		q.Push(msg1)
		msg2 := wrap("2", activity.ManualTrigger)
		q.Push(msg2)

		info, err := q.Pop()
		assert.NoError(t, err)
		assert.Equal(t, msg2, info)

		info, err = q.Pop()
		assert.NoError(t, err)
		assert.Equal(t, msg1, info)
	})
	t.Run("test lock state callback", func(t *testing.T) {
		var called bool
		q := queue.NewQueue(func(ctx workflow.Context, d *queue.Deploy) {
			called = true
		})
		q.SetLockForMergedItems(test.Background(), queue.LockState{
			Status: queue.LockedStatus,
		})

		assert.True(t, called)
	})

	t.Run("can pop empty queue unlocked", func(t *testing.T) {
		q := queue.NewQueue(nil)
		assert.Equal(t, false, q.CanPop())
	})

	t.Run("can pop empty queue locked", func(t *testing.T) {
		q := queue.NewQueue(noopCallback)
		q.SetLockForMergedItems(test.Background(), queue.LockState{
			Status: queue.LockedStatus,
		})
		assert.Equal(t, false, q.CanPop())
	})
	t.Run("can pop manual trigger locked", func(t *testing.T) {
		q := queue.NewQueue(noopCallback)
		msg1 := wrap("1", activity.ManualTrigger)
		q.Push(msg1)
		q.SetLockForMergedItems(test.Background(), queue.LockState{
			Status: queue.LockedStatus,
		})
		assert.Equal(t, true, q.CanPop())
	})
	t.Run("can pop manual trigger unlocked", func(t *testing.T) {
		q := queue.NewQueue(nil)
		msg1 := wrap("1", activity.ManualTrigger)
		q.Push(msg1)
		assert.Equal(t, true, q.CanPop())
	})
	t.Run("can pop merge trigger locked", func(t *testing.T) {
		q := queue.NewQueue(noopCallback)
		msg1 := wrap("1", activity.MergeTrigger)
		q.Push(msg1)
		q.SetLockForMergedItems(test.Background(), queue.LockState{
			Status: queue.LockedStatus,
		})
		assert.Equal(t, false, q.CanPop())
	})
	t.Run("can pop merge trigger unlocked", func(t *testing.T) {
		q := queue.NewQueue(nil)
		msg1 := wrap("1", activity.MergeTrigger)
		q.Push(msg1)
		assert.Equal(t, true, q.CanPop())
	})
}

func wrap(msg string, trigger activity.Trigger) terraform.DeploymentInfo {
	return terraform.DeploymentInfo{Revision: msg, Root: activity.Root{
		Trigger: trigger,
	}}

}
