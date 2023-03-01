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

package command_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/events/vcs/types"
	. "github.com/runatlantis/atlantis/testing"
)

func TestUpdateCombined(t *testing.T) {
	cases := []struct {
		status     models.VCSStatus
		command    command.Name
		expDescrip string
	}{
		{
			status:     models.QueuedVCSStatus,
			command:    command.Plan,
			expDescrip: "Plan queued.",
		},
		{
			status:     models.PendingVCSStatus,
			command:    command.Plan,
			expDescrip: "Plan in progress...",
		},
		{
			status:     models.FailedVCSStatus,
			command:    command.Plan,
			expDescrip: "Plan failed.",
		},
		{
			status:     models.SuccessVCSStatus,
			command:    command.Plan,
			expDescrip: "Plan succeeded.",
		},
		{
			status:     models.PendingVCSStatus,
			command:    command.Apply,
			expDescrip: "Apply in progress...",
		},
		{
			status:     models.FailedVCSStatus,
			command:    command.Apply,
			expDescrip: "Apply failed.",
		},
		{
			status:     models.SuccessVCSStatus,
			command:    command.Apply,
			expDescrip: "Apply succeeded.",
		},
	}

	for _, c := range cases {
		t.Run(c.expDescrip, func(t *testing.T) {
			RegisterMockTestingT(t)
			client := mocks.NewMockClient()

			titleBuilder := vcs.StatusTitleBuilder{TitlePrefix: "atlantis"}
			s := command.VCSStatusUpdater{Client: client, TitleBuilder: titleBuilder}
			ctx := context.Background()

			_, err := s.UpdateCombined(ctx, models.Repo{}, models.PullRequest{}, c.status, c.command, "", "")
			Ok(t, err)

			expSrc := fmt.Sprintf("atlantis/%s", c.command)
			client.VerifyWasCalledOnce().UpdateStatus(ctx, types.UpdateStatusRequest{
				Repo:        models.Repo{},
				PullNum:     0,
				State:       c.status,
				StatusName:  expSrc,
				Description: c.expDescrip,
				CommandName: c.command.TitleString(),
			})
		})
	}
}

func TestUpdateCombinedCount(t *testing.T) {
	cases := []struct {
		status     models.VCSStatus
		command    command.Name
		numSuccess int
		numTotal   int
		expDescrip string
	}{
		{
			status:     models.QueuedVCSStatus,
			command:    command.Plan,
			numSuccess: 0,
			numTotal:   2,
			expDescrip: "0/2 projects planned successfully.",
		},
		{
			status:     models.PendingVCSStatus,
			command:    command.Plan,
			numSuccess: 0,
			numTotal:   2,
			expDescrip: "0/2 projects planned successfully.",
		},
		{
			status:     models.FailedVCSStatus,
			command:    command.Plan,
			numSuccess: 1,
			numTotal:   2,
			expDescrip: "1/2 projects planned successfully.",
		},
		{
			status:     models.SuccessVCSStatus,
			command:    command.Plan,
			numSuccess: 2,
			numTotal:   2,
			expDescrip: "2/2 projects planned successfully.",
		},
		{
			status:     models.FailedVCSStatus,
			command:    command.Apply,
			numSuccess: 0,
			numTotal:   2,
			expDescrip: "0/2 projects applied successfully.",
		},
		{
			status:     models.PendingVCSStatus,
			command:    command.Apply,
			numSuccess: 1,
			numTotal:   2,
			expDescrip: "1/2 projects applied successfully.",
		},
		{
			status:     models.SuccessVCSStatus,
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
			titleBuilder := vcs.StatusTitleBuilder{TitlePrefix: "atlantis-test"}
			s := command.VCSStatusUpdater{Client: client, TitleBuilder: titleBuilder}
			ctx := context.Background()
			_, err := s.UpdateCombinedCount(ctx, models.Repo{}, models.PullRequest{}, c.status, c.command, c.numSuccess, c.numTotal, "")
			Ok(t, err)

			expSrc := fmt.Sprintf("%s/%s", titleBuilder.TitlePrefix, c.command)
			client.VerifyWasCalledOnce().UpdateStatus(ctx, types.UpdateStatusRequest{
				Repo:        models.Repo{},
				PullNum:     0,
				State:       c.status,
				StatusName:  expSrc,
				Description: c.expDescrip,
				CommandName: c.command.TitleString(),
				NumSuccess:  strconv.FormatInt(int64(c.numSuccess), 10),
				NumTotal:    strconv.FormatInt(int64(c.numTotal), 10),
			})
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
			titleBuilder := vcs.StatusTitleBuilder{TitlePrefix: "atlantis"}
			s := command.VCSStatusUpdater{Client: client, TitleBuilder: titleBuilder}
			ctx := context.Background()
			_, err := s.UpdateProject(ctx, command.ProjectContext{
				ProjectName: c.projectName,
				RepoRelDir:  c.repoRelDir,
				Workspace:   c.workspace,
			},
				command.Plan,
				models.PendingVCSStatus,
				"url", "")
			Ok(t, err)
			client.VerifyWasCalledOnce().UpdateStatus(ctx, types.UpdateStatusRequest{
				Repo:        models.Repo{},
				PullNum:     0,
				State:       models.PendingVCSStatus,
				StatusName:  c.expSrc,
				Description: "Plan in progress...",
				DetailsURL:  "url",

				CommandName: "Plan",
				Workspace:   c.workspace,
				Directory:   c.repoRelDir,
				Project:     c.projectName,
			})
		})
	}
}

// Test that it uses the right words in the description.
func TestDefaultCommitStatusUpdater_UpdateProject(t *testing.T) {
	RegisterMockTestingT(t)
	cases := []struct {
		status     models.VCSStatus
		cmd        command.Name
		expDescrip string
	}{
		{
			models.PendingVCSStatus,
			command.Plan,
			"Plan in progress...",
		},
		{
			models.FailedVCSStatus,
			command.Plan,
			"Plan failed.",
		},
		{
			models.SuccessVCSStatus,
			command.Plan,
			"Plan succeeded.",
		},
		{
			models.PendingVCSStatus,
			command.Apply,
			"Apply in progress...",
		},
		{
			models.FailedVCSStatus,
			command.Apply,
			"Apply failed.",
		},
		{
			models.SuccessVCSStatus,
			command.Apply,
			"Apply succeeded.",
		},
	}

	for _, c := range cases {
		t.Run(c.expDescrip, func(t *testing.T) {
			client := mocks.NewMockClient()
			titleBuilder := vcs.StatusTitleBuilder{TitlePrefix: "atlantis"}
			s := command.VCSStatusUpdater{Client: client, TitleBuilder: titleBuilder}
			ctx := context.Background()
			_, err := s.UpdateProject(ctx, command.ProjectContext{
				RepoRelDir: ".",
				Workspace:  "default",
			},
				c.cmd,
				c.status,
				"url", "")
			Ok(t, err)
			client.VerifyWasCalledOnce().UpdateStatus(ctx, types.UpdateStatusRequest{
				Repo:        models.Repo{},
				PullNum:     0,
				State:       c.status,
				StatusName:  fmt.Sprintf("atlantis/%s: ./default", c.cmd.String()),
				Description: c.expDescrip,
				DetailsURL:  "url",
				CommandName: c.cmd.TitleString(),
				Workspace:   "default",
				Directory:   ".",
			})
		})
	}
}

// Test that we can set the status name.
func TestDefaultCommitStatusUpdater_UpdateProjectCustomStatusName(t *testing.T) {
	RegisterMockTestingT(t)
	client := mocks.NewMockClient()
	titleBuilder := vcs.StatusTitleBuilder{TitlePrefix: "custom"}
	s := command.VCSStatusUpdater{Client: client, TitleBuilder: titleBuilder}
	ctx := context.Background()
	_, err := s.UpdateProject(ctx, command.ProjectContext{
		RepoRelDir: ".",
		Workspace:  "default",
	},
		command.Apply,
		models.SuccessVCSStatus,
		"url", "")
	Ok(t, err)
	client.VerifyWasCalledOnce().UpdateStatus(ctx, types.UpdateStatusRequest{
		Repo:        models.Repo{},
		PullNum:     0,
		State:       models.SuccessVCSStatus,
		StatusName:  "custom/apply: ./default",
		Description: "Apply succeeded.",
		DetailsURL:  "url",
		Workspace:   "default",
		CommandName: "Apply",
		Directory:   ".",
	})
}
