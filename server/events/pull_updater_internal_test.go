// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"errors"
	"strings"
	"testing"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/models/testdata"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestPullUpdater_SilencedCommentsStillPostFailures(t *testing.T) {
	lockFailure := "This project is currently locked by an unapplied plan from pull #2."
	silenced := []string{"plan"}
	planned := command.ProjectResult{
		Command:           command.Plan,
		RepoRelDir:        "planned",
		Workspace:         "default",
		SilencePRComments: silenced,
		ProjectCommandOutput: command.ProjectCommandOutput{
			PlanSuccess: &models.PlanSuccess{TerraformOutput: "very long plan output"},
		},
	}
	locked := command.ProjectResult{
		Command:              command.Plan,
		RepoRelDir:           "locked",
		Workspace:            "default",
		SilencePRComments:    silenced,
		ProjectCommandOutput: command.ProjectCommandOutput{Failure: lockFailure},
	}
	errored := command.ProjectResult{
		Command:              command.Plan,
		RepoRelDir:           "errored",
		Workspace:            "default",
		SilencePRComments:    silenced,
		ProjectCommandOutput: command.ProjectCommandOutput{Error: errors.New("very long terraform error")},
	}

	cases := map[string]struct {
		results     []command.ProjectResult
		expComments int
	}{
		"successful plan is silenced":       {results: []command.ProjectResult{planned}, expComments: 0},
		"plan error is silenced":            {results: []command.ProjectResult{errored}, expComments: 0},
		"lock failure is still posted":      {results: []command.ProjectResult{locked}, expComments: 1},
		"only the lock failure is rendered": {results: []command.ProjectResult{planned, locked, errored}, expComments: 1},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			RegisterMockTestingT(t)
			vcsClient := vcsmocks.NewMockClient()
			updater := &PullUpdater{
				VCSClient:        vcsClient,
				MarkdownRenderer: NewMarkdownRenderer(false, false, false, false, false, false, "", "atlantis", false, false),
			}
			ctx := &command.Context{Log: logging.NewNoopLogger(t), Pull: testdata.Pull}

			updater.updatePull(ctx, &CommentCommand{Name: command.Plan}, command.Result{ProjectResults: c.results})

			_, _, _, comments, _ := vcsClient.VerifyWasCalled(Times(c.expComments)).CreateComment(
				Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]()).GetAllCapturedArguments()
			for _, comment := range comments {
				Assert(t, strings.Contains(comment, lockFailure), "got: %s", comment)
				Assert(t, !strings.Contains(comment, "very long"), "got: %s", comment)
			}
		})
	}
}
