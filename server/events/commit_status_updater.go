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

package events

import (
	"fmt"
	"strings"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_commit_status_updater.go CommitStatusUpdater

// CommitStatusUpdater updates the status of a commit with the VCS host. We set
// the status to signify whether the plan/apply succeeds.
type CommitStatusUpdater interface {
	// UpdateCombined updates the combined status of the head commit of pull.
	// A combined status represents all the projects modified in the pull.
	UpdateCombined(repo models.Repo, pull models.PullRequest, status models.CommitStatus, cmdName models.CommandName) error
	// UpdateCombinedCount updates the combined status to reflect the
	// numSuccess out of numTotal.
	UpdateCombinedCount(repo models.Repo, pull models.PullRequest, status models.CommitStatus, cmdName models.CommandName, numSuccess int, numTotal int) error
	// UpdateProject sets the commit status for the project represented by
	// ctx.
	UpdateProject(ctx models.ProjectCommandContext, cmdName models.CommandName, status models.CommitStatus, url string) error
}

// DefaultCommitStatusUpdater implements CommitStatusUpdater.
type DefaultCommitStatusUpdater struct {
	Client       vcs.Client
	TitleBuilder vcs.StatusTitleBuilder
}

func (d *DefaultCommitStatusUpdater) UpdateCombined(repo models.Repo, pull models.PullRequest, status models.CommitStatus, cmdName models.CommandName) error {
	src := d.TitleBuilder.Build(cmdName.String())
	var descripWords string
	switch status {
	case models.PendingCommitStatus:
		descripWords = "in progress..."
	case models.FailedCommitStatus:
		descripWords = "failed."
	case models.SuccessCommitStatus:
		descripWords = "succeeded."
	}
	descrip := fmt.Sprintf("%s %s", strings.Title(cmdName.String()), descripWords)
	return d.Client.UpdateStatus(repo, pull, status, src, descrip, "")
}

func (d *DefaultCommitStatusUpdater) UpdateCombinedCount(repo models.Repo, pull models.PullRequest, status models.CommitStatus, cmdName models.CommandName, numSuccess int, numTotal int) error {
	src := d.TitleBuilder.Build(cmdName.String())
	cmdVerb := "unknown"

	switch cmdName {
	case models.PlanCommand:
		cmdVerb = "planned"
	case models.PolicyCheckCommand:
		cmdVerb = "policies checked"
	case models.ApplyCommand:
		cmdVerb = "applied"
	}

	return d.Client.UpdateStatus(repo, pull, status, src, fmt.Sprintf("%d/%d projects %s successfully.", numSuccess, numTotal, cmdVerb), "")
}

func (d *DefaultCommitStatusUpdater) UpdateProject(ctx models.ProjectCommandContext, cmdName models.CommandName, status models.CommitStatus, url string) error {
	projectID := ctx.ProjectName
	if projectID == "" {
		projectID = fmt.Sprintf("%s/%s", ctx.RepoRelDir, ctx.Workspace)
	}

	src := d.TitleBuilder.Build(cmdName.String(), vcs.StatusTitleOptions{
		ProjectName: projectID,
	})
	var descripWords string
	switch status {
	case models.PendingCommitStatus:
		descripWords = "in progress..."
	case models.FailedCommitStatus:
		descripWords = "failed."
	case models.SuccessCommitStatus:
		descripWords = "succeeded."
	}
	descrip := fmt.Sprintf("%s %s", strings.Title(cmdName.String()), descripWords)
	return d.Client.UpdateStatus(ctx.BaseRepo, ctx.Pull, status, src, descrip, url)
}
