// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"errors"
	"strings"
	"testing"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestStateShowCommandRunner_DispatchesWorkflow(t *testing.T) {
	name := command.StateShow
	for _, outcome := range []string{"success", "builder error", "no projects"} {
		t.Run(name.String()+"/"+outcome, func(t *testing.T) {
			RegisterMockTestingT(t)
			builder := mocks.NewMockProjectCommandBuilder()
			projectRunner := mocks.NewMockProjectCommandRunner()
			vcs := vcsmocks.NewMockClient()
			updater := &events.PullUpdater{VCSClient: vcs, MarkdownRenderer: events.NewMarkdownRenderer(false, false, false, false, false, false, "", "atlantis", false, false)}
			runner := events.NewStateShowCommandRunner(updater, builder, projectRunner, true)
			build := builder.BuildStateShowCommands
			run := projectRunner.StateShow
			output := command.ProjectCommandOutput{StateShowSuccess: &models.StateShowSuccess{Output: "value"}}
			var buildErr error
			cmds := []command.ProjectContext{{CommandName: name, RepoRelDir: ".", Workspace: "default"}}
			if outcome == "builder error" {
				buildErr = errors.New("building workflow")
				cmds = nil
			}
			if outcome == "no projects" {
				cmds = nil
			}
			When(build(Any[*command.Context](), Any[*events.CommentCommand]())).ThenReturn(cmds, buildErr)
			When(run(Any[command.ProjectContext]())).ThenReturn(output)
			When(vcs.CreateComment(Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]())).Then(func(params []Param) ReturnValues {
				comment := params[3].(string)
				if outcome == "success" {
					Assert(t, strings.Contains(comment, "value"), "missing workflow output: %s", comment)
				}
				if outcome == "builder error" {
					Assert(t, strings.Contains(comment, "building workflow"), "missing build error: %s", comment)
				}
				return ReturnValues{nil}
			})
			ctx := &command.Context{Log: logging.NewNoopLogger(t).WithHistory(), Pull: models.PullRequest{Num: 1}}
			runner.Run(ctx, &events.CommentCommand{Name: name})
			expected := Never()
			if outcome == "success" {
				expected = Once()
			}
			projectRunner.VerifyWasCalled(expected).StateShow(Any[command.ProjectContext]())
			commentCount := Once()
			if outcome == "no projects" {
				commentCount = Never()
			}
			vcs.VerifyWasCalled(commentCount).CreateComment(Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Eq(name.String()))
		})
	}
}
