// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
	. "github.com/runatlantis/atlantis/testing"
	tally "github.com/uber-go/tally/v4"
)

func TestRunAndEmitStats_PullRequestNumberTag(t *testing.T) {
	cases := []struct {
		name        string
		dropPRTag   bool
		wantPRTag   bool
		wantPRValue string
	}{
		{name: "included by default", wantPRTag: true, wantPRValue: "123"},
		{name: "dropped when filtered", dropPRTag: true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := tally.NewTestScope("atlantis", nil)
			var scope tally.Scope = root
			if c.dropPRTag {
				scope = metrics.WithoutTags(scope, metrics.PullRequestNumberTag)
			}

			ctx := command.ProjectContext{
				CommandName: command.Plan,
				BaseRepo:    models.Repo{FullName: "owner/repo"},
				Pull:        models.PullRequest{Num: 123},
				ProjectName: "project",
				RepoRelDir:  "dir",
				Workspace:   "default",
				Log:         logging.NewNoopLogger(t),
			}
			events.RunAndEmitStats(ctx, func(command.ProjectContext) command.ProjectCommandOutput {
				return command.ProjectCommandOutput{}
			}, scope.SubScope("project"))

			var success tally.CounterSnapshot
			for _, cs := range root.Snapshot().Counters() {
				if cs.Name() == "atlantis.project.plan."+metrics.ExecutionSuccessMetric {
					success = cs
				}
			}
			Assert(t, success != nil, "plan success counter not found")

			tags := success.Tags()
			v, ok := tags[metrics.PullRequestNumberTag]
			Equals(t, c.wantPRTag, ok)
			Equals(t, c.wantPRValue, v)
			Equals(t, "owner/repo", tags["base_repo"])
			Equals(t, "project", tags["project"])
		})
	}
}
