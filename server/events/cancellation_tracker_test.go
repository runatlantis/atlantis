// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/stretchr/testify/assert"
)

func TestCancellationTracker_CancelReachesOnlyEarlierMarks(t *testing.T) {
	tracker := events.NewCancellationTracker()
	pull := models.PullRequest{Num: 1, BaseRepo: models.Repo{FullName: "owner/repo"}}
	otherPull := models.PullRequest{Num: 2, BaseRepo: models.Repo{FullName: "owner/repo"}}

	before := tracker.Mark()
	tracker.Cancel(pull)
	after := tracker.Mark()

	assert.True(t, tracker.IsCancelled(pull, before))
	assert.False(t, tracker.IsCancelled(pull, after))
	assert.False(t, tracker.IsCancelled(otherPull, before))
}

func TestCancellationTracker_ClearForgetsCancels(t *testing.T) {
	tracker := events.NewCancellationTracker()
	pull := models.PullRequest{Num: 1, BaseRepo: models.Repo{FullName: "owner/repo"}}

	before := tracker.Mark()
	tracker.Cancel(pull)
	tracker.Clear(pull)
	assert.False(t, tracker.IsCancelled(pull, before))

	tracker.Cancel(pull)
	assert.True(t, tracker.IsCancelled(pull, before), "a cancel after Clear must still reach commands that started before it")
}
