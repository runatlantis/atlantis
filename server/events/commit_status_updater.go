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

	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

//go:generate pegomock generate -m --package mocks -o mocks/mock_commit_status_updater.go CommitStatusUpdater

// CommitStatusUpdater updates the status of a commit with the VCS host. We set
// the status to signify whether the plan/apply succeeds.
type CommitStatusUpdater interface {
	// UpdateCombined updates the combined status of the head commit of pull.
	// A combined status represents all the projects modified in the pull.
	UpdateCombined(repo models.Repo, pull models.PullRequest, status models.CommitStatus, cmdName command.Name) error
	// UpdateCombinedCount updates the combined status to reflect the
	// numSuccess out of numTotal.
	UpdateCombinedCount(repo models.Repo, pull models.PullRequest, status models.CommitStatus, cmdName command.Name, numSuccess int, numTotal int) error

	UpdatePreWorkflowHook(pull models.PullRequest, status models.CommitStatus, hookDescription string, runtimeDescription string, url string) error
	UpdatePostWorkflowHook(pull models.PullRequest, status models.CommitStatus, hookDescription string, runtimeDescription string, url string) error
}

// DefaultCommitStatusUpdater implements CommitStatusUpdater.
type DefaultCommitStatusUpdater struct {
	Client vcs.Client
	// StatusName is the name used to identify Atlantis when creating PR statuses.
	StatusName string
}

// ensure DefaultCommitStatusUpdater implements runtime.StatusUpdater interface
// cause runtime.StatusUpdater is extracted for resolving circular dependency
var _ runtime.StatusUpdater = (*DefaultCommitStatusUpdater)(nil)

func (d *DefaultCommitStatusUpdater) UpdateCombined(repo models.Repo, pull models.PullRequest, status models.CommitStatus, cmdName command.Name) error {
	src := fmt.Sprintf("%s/%s", d.StatusName, cmdName.String())
	var descripWords string
	switch status {
	case models.PendingCommitStatus:
		descripWords = genProjectStatusDescription(cmdName.String(), "in progress...")
	case models.FailedCommitStatus:
		descripWords = genProjectStatusDescription(cmdName.String(), "failed.")
	case models.SuccessCommitStatus:
		descripWords = genProjectStatusDescription(cmdName.String(), "succeeded.")
	}
	return d.Client.UpdateStatus(repo, pull, status, src, descripWords, "")
}

func (d *DefaultCommitStatusUpdater) UpdateCombinedCount(repo models.Repo, pull models.PullRequest, status models.CommitStatus, cmdName command.Name, numSuccess int, numTotal int) error {
	src := fmt.Sprintf("%s/%s", d.StatusName, cmdName.String())
	cmdVerb := "unknown"

	switch cmdName {
	case command.Plan:
		cmdVerb = "planned"
	case command.PolicyCheck:
		cmdVerb = "policies checked"
	case command.Apply:
		cmdVerb = "applied"
	}

	return d.Client.UpdateStatus(repo, pull, status, src, fmt.Sprintf("%d/%d projects %s successfully.", numSuccess, numTotal, cmdVerb), "")
}

func (d *DefaultCommitStatusUpdater) UpdateProject(ctx command.ProjectContext, cmdName command.Name, status models.CommitStatus, url string, result *command.ProjectResult) error {
	projectID := ctx.ProjectName
	if projectID == "" {
		projectID = fmt.Sprintf("%s/%s", ctx.RepoRelDir, ctx.Workspace)
	}
	src := fmt.Sprintf("%s/%s: %s", d.StatusName, cmdName.String(), projectID)
	var descripWords string
	switch status {
	case models.PendingCommitStatus:
		descripWords = genProjectStatusDescription(cmdName.String(), "in progress...")
	case models.FailedCommitStatus:
		descripWords = genProjectStatusDescription(cmdName.String(), "failed.")
	case models.SuccessCommitStatus:
		if result != nil && result.PlanSuccess != nil {
			descripWords = result.PlanSuccess.DiffSummary()
		} else {
			descripWords = genProjectStatusDescription(cmdName.String(), "succeeded.")
		}
	}
	return d.Client.UpdateStatus(ctx.BaseRepo, ctx.Pull, status, src, descripWords, url)
}

func genProjectStatusDescription(cmdName, description string) string {
	return fmt.Sprintf("%s %s", cases.Title(language.English).String(cmdName), description)
}

func (d *DefaultCommitStatusUpdater) UpdatePreWorkflowHook(pull models.PullRequest, status models.CommitStatus, hookDescription string, runtimeDescription string, url string) error {
	return d.updateWorkflowHook(pull, status, hookDescription, runtimeDescription, "pre_workflow_hook", url)
}

func (d *DefaultCommitStatusUpdater) UpdatePostWorkflowHook(pull models.PullRequest, status models.CommitStatus, hookDescription string, runtimeDescription string, url string) error {
	return d.updateWorkflowHook(pull, status, hookDescription, runtimeDescription, "post_workflow_hook", url)
}

func (d *DefaultCommitStatusUpdater) updateWorkflowHook(pull models.PullRequest, status models.CommitStatus, hookDescription string, runtimeDescription string, workflowType string, url string) error {
	src := fmt.Sprintf("%s/%s: %s", d.StatusName, workflowType, hookDescription)

	var descripWords string
	if runtimeDescription != "" {
		descripWords = runtimeDescription
	} else {
		switch status {
		case models.PendingCommitStatus:
			descripWords = "in progress..."
		case models.FailedCommitStatus:
			descripWords = "failed."
		case models.SuccessCommitStatus:
			descripWords = "succeeded."
		}
	}

	return d.Client.UpdateStatus(pull.BaseRepo, pull, status, src, descripWords, url)
}
