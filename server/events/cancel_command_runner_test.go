package events

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/stretchr/testify/assert"
)

func TestCancelCommandRunner_OnlyStopsQueuedOperations(t *testing.T) {
	// Setup
	tracker := NewCancellationTracker()

	// Create a mock pull request
	pull := models.PullRequest{
		BaseRepo: models.Repo{FullName: "owner/repo"},
		Num:      123,
	}

	// Create operation keys
	opKey1 := NewOperationKey(pull, "project1", "default", "dir1", "job1")
	opKey2 := NewOperationKey(pull, "project2", "default", "dir2", "job2")
	opKey3 := NewOperationKey(pull, "project3", "default", "dir3", "job3")

	// Track operations with different states
	tracker.TrackOperation(opKey1, OperationQueued)  // Should be cancelled
	tracker.TrackOperation(opKey2, OperationRunning) // Should continue running
	tracker.TrackOperation(opKey3, OperationQueued)  // Should be cancelled

	// Cancel queued operations
	tracker.CancelQueuedOperations(pull)

	// Verify results
	assert.True(t, tracker.IsOperationCancelled(opKey1), "Queued operation should be cancelled")
	assert.False(t, tracker.IsOperationCancelled(opKey2), "Running operation should not be cancelled")
	assert.True(t, tracker.IsOperationCancelled(opKey3), "Queued operation should be cancelled")

	// Verify states directly
	assert.Equal(t, OperationCancelled, tracker.GetOperationState(opKey1))
	assert.Equal(t, OperationRunning, tracker.GetOperationState(opKey2))
	assert.Equal(t, OperationCancelled, tracker.GetOperationState(opKey3))
}

func TestCancelCommandRunner_OperationLifecycle(t *testing.T) {
	// Setup
	tracker := NewCancellationTracker()

	pull := models.PullRequest{
		BaseRepo: models.Repo{FullName: "owner/repo"},
		Num:      456,
	}

	opKey := NewOperationKey(pull, "project1", "default", "dir1", "job1")

	// Test lifecycle: queued -> running -> completed
	tracker.TrackOperation(opKey, OperationQueued)
	assert.Equal(t, OperationQueued, tracker.GetOperationState(opKey))
	assert.False(t, tracker.IsOperationCancelled(opKey))

	// Start running - should not be cancellable anymore
	tracker.UpdateOperationState(opKey, OperationRunning)
	tracker.CancelQueuedOperations(pull)
	assert.False(t, tracker.IsOperationCancelled(opKey), "Running operation should not be cancelled")
	assert.Equal(t, OperationRunning, tracker.GetOperationState(opKey))

	// Complete the operation
	tracker.UpdateOperationState(opKey, OperationCompleted)
	assert.Equal(t, OperationCompleted, tracker.GetOperationState(opKey))
}

func TestCancelCommandRunner_ClearOperations(t *testing.T) {
	// Setup
	tracker := NewCancellationTracker()

	pull := models.PullRequest{
		BaseRepo: models.Repo{FullName: "owner/repo"},
		Num:      789,
	}

	opKey1 := NewOperationKey(pull, "project1", "default", "dir1", "job1")
	opKey2 := NewOperationKey(pull, "project2", "default", "dir2", "job2")

	// Track some operations
	tracker.TrackOperation(opKey1, OperationQueued)
	tracker.TrackOperation(opKey2, OperationRunning)

	// Clear all operations for the pull request
	tracker.ClearOperations(pull)

	// Verify operations are cleared (should return default state)
	assert.Equal(t, OperationQueued, tracker.GetOperationState(opKey1), "Should return default state after clearing")
	assert.Equal(t, OperationQueued, tracker.GetOperationState(opKey2), "Should return default state after clearing")
}

func TestCancelCommandRunner_PullRequestLevelCancellation(t *testing.T) {
	// Setup
	tracker := NewCancellationTracker()

	pull := models.PullRequest{
		BaseRepo: models.Repo{FullName: "owner/repo"},
		Num:      999,
	}

	opKey1 := NewOperationKey(pull, "project1", "default", "dir1", "job1")
	opKey2 := NewOperationKey(pull, "project2", "default", "dir2", "job2")
	opKey3 := NewOperationKey(pull, "project3", "default", "dir3", "job3")

	// Track operations
	tracker.TrackOperation(opKey1, OperationQueued)
	tracker.TrackOperation(opKey2, OperationRunning)
	tracker.TrackOperation(opKey3, OperationQueued)

	// Cancel the entire pull request
	tracker.CancelPullRequest(pull)

	// Verify PR is marked as cancelled
	assert.True(t, tracker.IsPullRequestCancelled(pull), "Pull request should be cancelled")

	// Verify queued operations are cancelled but running ones continue
	assert.True(t, tracker.IsOperationCancelled(opKey1), "Queued operation should be cancelled")
	assert.False(t, tracker.IsOperationCancelled(opKey2), "Running operation should not be cancelled")
	assert.True(t, tracker.IsOperationCancelled(opKey3), "Queued operation should be cancelled")

	// Verify states
	assert.Equal(t, OperationCancelled, tracker.GetOperationState(opKey1))
	assert.Equal(t, OperationRunning, tracker.GetOperationState(opKey2))
	assert.Equal(t, OperationCancelled, tracker.GetOperationState(opKey3))
}
