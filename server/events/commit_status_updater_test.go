// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

package events_test

import (
	"fmt"
	"testing"

	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs/mocks"
	. "github.com/runatlantis/atlantis/testing"
)

func TestUpdateCombined(t *testing.T) {
	cases := []struct {
		status     models.CommitStatus
		command    models.CommandName
		expDescrip string
	}{
		{
			status:     models.PendingCommitStatus,
			command:    models.PlanCommand,
			expDescrip: "Plan in progress...",
		},
		{
			status:     models.FailedCommitStatus,
			command:    models.PlanCommand,
			expDescrip: "Plan failed.",
		},
		{
			status:     models.SuccessCommitStatus,
			command:    models.PlanCommand,
			expDescrip: "Plan succeeded.",
		},
		{
			status:     models.PendingCommitStatus,
			command:    models.ApplyCommand,
			expDescrip: "Apply in progress...",
		},
		{
			status:     models.FailedCommitStatus,
			command:    models.ApplyCommand,
			expDescrip: "Apply failed.",
		},
		{
			status:     models.SuccessCommitStatus,
			command:    models.ApplyCommand,
			expDescrip: "Apply succeeded.",
		},
	}

	for _, c := range cases {
		t.Run(c.expDescrip, func(t *testing.T) {
			RegisterMockTestingT(t)
			client := mocks.NewMockClient()
			s := events.DefaultCommitStatusUpdater{Client: client, StatusName: "atlantis"}
			err := s.UpdateCombined(models.Repo{}, models.PullRequest{}, c.status, c.command)
			Ok(t, err)

			expSrc := fmt.Sprintf("atlantis/%s", c.command)
			client.VerifyWasCalledOnce().UpdateStatus(models.Repo{}, models.PullRequest{}, c.status, expSrc, c.expDescrip, "")
		})
	}
}

func TestUpdateCombinedCount(t *testing.T) {
	cases := []struct {
		status     models.CommitStatus
		command    models.CommandName
		numSuccess int
		numTotal   int
		expDescrip string
	}{
		{
			status:     models.PendingCommitStatus,
			command:    models.PlanCommand,
			numSuccess: 0,
			numTotal:   2,
			expDescrip: "0/2 projects planned successfully.",
		},
		{
			status:     models.FailedCommitStatus,
			command:    models.PlanCommand,
			numSuccess: 1,
			numTotal:   2,
			expDescrip: "1/2 projects planned successfully.",
		},
		{
			status:     models.SuccessCommitStatus,
			command:    models.PlanCommand,
			numSuccess: 2,
			numTotal:   2,
			expDescrip: "2/2 projects planned successfully.",
		},
		{
			status:     models.FailedCommitStatus,
			command:    models.ApplyCommand,
			numSuccess: 0,
			numTotal:   2,
			expDescrip: "0/2 projects applied successfully.",
		},
		{
			status:     models.PendingCommitStatus,
			command:    models.ApplyCommand,
			numSuccess: 1,
			numTotal:   2,
			expDescrip: "1/2 projects applied successfully.",
		},
		{
			status:     models.SuccessCommitStatus,
			command:    models.ApplyCommand,
			numSuccess: 2,
			numTotal:   2,
			expDescrip: "2/2 projects applied successfully.",
		},
	}

	for _, c := range cases {
		t.Run(c.expDescrip, func(t *testing.T) {
			RegisterMockTestingT(t)
			client := mocks.NewMockClient()
			s := events.DefaultCommitStatusUpdater{Client: client, StatusName: "atlantis-test"}
			err := s.UpdateCombinedCount(models.Repo{}, models.PullRequest{}, c.status, c.command, c.numSuccess, c.numTotal)
			Ok(t, err)

			expSrc := fmt.Sprintf("%s/%s", s.StatusName, c.command)
			client.VerifyWasCalledOnce().UpdateStatus(models.Repo{}, models.PullRequest{}, c.status, expSrc, c.expDescrip, "")
		})
	}
}

// Test that it sets the "source" properly depending on if the project is
// named or not.
func TestDefaultCommitStatusUpdater_UpdateProjectSrc(t *testing.T) {
	RegisterMockTestingT(t)
	cases := []struct {
		projectName string
		repoRelDir  string
		workspace   string
		expSrc      string
	}{
		{
			projectName: "name",
			repoRelDir:  ".",
			workspace:   "default",
			expSrc:      "atlantis/plan: name",
		},
		{
			projectName: "",
			repoRelDir:  "dir1/dir2",
			workspace:   "workspace",
			expSrc:      "atlantis/plan: dir1/dir2/workspace",
		},
	}

	for _, c := range cases {
		t.Run(c.expSrc, func(t *testing.T) {
			client := mocks.NewMockClient()
			s := events.DefaultCommitStatusUpdater{Client: client, StatusName: "atlantis"}
			err := s.UpdateProject(models.ProjectCommandContext{
				ProjectName: c.projectName,
				RepoRelDir:  c.repoRelDir,
				Workspace:   c.workspace,
			},
				models.PlanCommand,
				models.PendingCommitStatus,
				"url")
			Ok(t, err)
			client.VerifyWasCalledOnce().UpdateStatus(models.Repo{}, models.PullRequest{}, models.PendingCommitStatus, c.expSrc, "Plan in progress...", "url")
		})
	}
}

// Test that it uses the right words in the description.
func TestDefaultCommitStatusUpdater_UpdateProject(t *testing.T) {
	RegisterMockTestingT(t)
	cases := []struct {
		status     models.CommitStatus
		cmd        models.CommandName
		expDescrip string
	}{
		{
			models.PendingCommitStatus,
			models.PlanCommand,
			"Plan in progress...",
		},
		{
			models.FailedCommitStatus,
			models.PlanCommand,
			"Plan failed.",
		},
		{
			models.SuccessCommitStatus,
			models.PlanCommand,
			"Plan succeeded.",
		},
		{
			models.PendingCommitStatus,
			models.ApplyCommand,
			"Apply in progress...",
		},
		{
			models.FailedCommitStatus,
			models.ApplyCommand,
			"Apply failed.",
		},
		{
			models.SuccessCommitStatus,
			models.ApplyCommand,
			"Apply succeeded.",
		},
	}

	for _, c := range cases {
		t.Run(c.expDescrip, func(t *testing.T) {
			client := mocks.NewMockClient()
			s := events.DefaultCommitStatusUpdater{Client: client, StatusName: "atlantis"}
			err := s.UpdateProject(models.ProjectCommandContext{
				RepoRelDir: ".",
				Workspace:  "default",
			},
				c.cmd,
				c.status,
				"url")
			Ok(t, err)
			client.VerifyWasCalledOnce().UpdateStatus(models.Repo{}, models.PullRequest{}, c.status, fmt.Sprintf("atlantis/%s: ./default", c.cmd.String()), c.expDescrip, "url")
		})
	}
}

// Test that we can set the status name.
func TestDefaultCommitStatusUpdater_UpdateProjectCustomStatusName(t *testing.T) {
	RegisterMockTestingT(t)
	client := mocks.NewMockClient()
	s := events.DefaultCommitStatusUpdater{Client: client, StatusName: "custom"}
	err := s.UpdateProject(models.ProjectCommandContext{
		RepoRelDir: ".",
		Workspace:  "default",
	},
		models.ApplyCommand,
		models.SuccessCommitStatus,
		"url")
	Ok(t, err)
	client.VerifyWasCalledOnce().UpdateStatus(models.Repo{}, models.PullRequest{},
		models.SuccessCommitStatus, "custom/apply: ./default", "Apply succeeded.", "url")
}
