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
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs/mocks"
	. "github.com/runatlantis/atlantis/testing"
)

func TestUpdateCombined(t *testing.T) {
	cases := []struct {
		status     models.CommitStatus
		command    command.Name
		expDescrip string
	}{
		{
			status:     models.PendingCommitStatus,
			command:    command.Plan,
			expDescrip: "Plan in progress...",
		},
		{
			status:     models.FailedCommitStatus,
			command:    command.Plan,
			expDescrip: "Plan failed.",
		},
		{
			status:     models.SuccessCommitStatus,
			command:    command.Plan,
			expDescrip: "Plan succeeded.",
		},
		{
			status:     models.PendingCommitStatus,
			command:    command.Apply,
			expDescrip: "Apply in progress...",
		},
		{
			status:     models.FailedCommitStatus,
			command:    command.Apply,
			expDescrip: "Apply failed.",
		},
		{
			status:     models.SuccessCommitStatus,
			command:    command.Apply,
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
		command    command.Name
		numSuccess int
		numTotal   int
		expDescrip string
	}{
		{
			status:     models.PendingCommitStatus,
			command:    command.Plan,
			numSuccess: 0,
			numTotal:   2,
			expDescrip: "0/2 projects planned successfully.",
		},
		{
			status:     models.FailedCommitStatus,
			command:    command.Plan,
			numSuccess: 1,
			numTotal:   2,
			expDescrip: "1/2 projects planned successfully.",
		},
		{
			status:     models.SuccessCommitStatus,
			command:    command.Plan,
			numSuccess: 2,
			numTotal:   2,
			expDescrip: "2/2 projects planned successfully.",
		},
		{
			status:     models.FailedCommitStatus,
			command:    command.Apply,
			numSuccess: 0,
			numTotal:   2,
			expDescrip: "0/2 projects applied successfully.",
		},
		{
			status:     models.PendingCommitStatus,
			command:    command.Apply,
			numSuccess: 1,
			numTotal:   2,
			expDescrip: "1/2 projects applied successfully.",
		},
		{
			status:     models.SuccessCommitStatus,
			command:    command.Apply,
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
			err := s.UpdateProject(command.ProjectContext{
				ProjectName: c.projectName,
				RepoRelDir:  c.repoRelDir,
				Workspace:   c.workspace,
			}, command.Plan, models.PendingCommitStatus, "url", nil)
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
		cmd        command.Name
		result     *command.ProjectResult
		expDescrip string
	}{
		{
			status:     models.PendingCommitStatus,
			cmd:        command.Plan,
			expDescrip: "Plan in progress...",
		},
		{
			status:     models.FailedCommitStatus,
			cmd:        command.Plan,
			expDescrip: "Plan failed.",
		},
		{
			status: models.SuccessCommitStatus,
			cmd:    command.Plan,
			result: &command.ProjectResult{
				PlanSuccess: &models.PlanSuccess{
					TerraformOutput: "aaa\nNote: Objects have changed outside of Terraform\nbbb\nPlan: 1 to add, 2 to change, 3 to destroy.\nbbb",
				},
			},
			expDescrip: "Plan: 1 to add, 2 to change, 3 to destroy.",
		},
		{
			status:     models.PendingCommitStatus,
			cmd:        command.Apply,
			expDescrip: "Apply in progress...",
		},
		{
			status:     models.FailedCommitStatus,
			cmd:        command.Apply,
			expDescrip: "Apply failed.",
		},
		{
			status: models.SuccessCommitStatus,
			cmd:    command.Apply,
			result: &command.ProjectResult{
				ApplySuccess: "success",
			},
			expDescrip: "Apply succeeded.",
		},
	}

	for _, c := range cases {
		t.Run(c.expDescrip, func(t *testing.T) {
			client := mocks.NewMockClient()
			s := events.DefaultCommitStatusUpdater{Client: client, StatusName: "atlantis"}
			err := s.UpdateProject(command.ProjectContext{
				RepoRelDir: ".",
				Workspace:  "default",
			}, c.cmd, c.status, "url", c.result)
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
	err := s.UpdateProject(command.ProjectContext{
		RepoRelDir: ".",
		Workspace:  "default",
	}, command.Apply, models.SuccessCommitStatus, "url", nil)
	Ok(t, err)
	client.VerifyWasCalledOnce().UpdateStatus(models.Repo{}, models.PullRequest{},
		models.SuccessCommitStatus, "custom/apply: ./default", "Apply succeeded.", "url")
}
