// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package metrics_test

import (
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/metrics"
	. "github.com/runatlantis/atlantis/testing"
	tally "github.com/uber-go/tally/v4"
)

func TestWithoutTags_DropsKeysFromDerivedScopes(t *testing.T) {
	root := tally.NewTestScope("atlantis", nil)
	scope := metrics.WithoutTags(root, metrics.PullRequestNumberTag)

	scope = scope.SubScope("github").Tagged(map[string]string{
		"base_repo":                  "owner/repo",
		metrics.PullRequestNumberTag: "1",
	})
	scope = scope.SubScope("create_comment").Tagged(map[string]string{
		metrics.PullRequestNumberTag: "2",
		"project":                    "proj",
	})
	scope.Counter(metrics.ExecutionSuccessMetric).Inc(1)
	scope.Timer(metrics.ExecutionTimeMetric).Record(time.Second)

	snapshot := root.Snapshot()
	Equals(t, 1, len(snapshot.Counters()))
	Equals(t, 1, len(snapshot.Timers()))
	want := map[string]string{"base_repo": "owner/repo", "project": "proj"}
	for _, c := range snapshot.Counters() {
		Equals(t, "atlantis.github.create_comment."+metrics.ExecutionSuccessMetric, c.Name())
		Equals(t, want, c.Tags())
	}
	for _, tm := range snapshot.Timers() {
		Equals(t, want, tm.Tags())
	}
}

func TestWithoutTags_NoKeysReturnsSameScope(t *testing.T) {
	root := tally.NewTestScope("atlantis", nil)
	Assert(t, metrics.WithoutTags(root) == root, "expected the original scope when no keys are given")
}
