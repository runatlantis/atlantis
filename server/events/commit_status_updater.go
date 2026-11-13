// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package events

import (
	"crypto/sha256"
	"fmt"
	"unicode/utf8"

	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

//go:generate go tool pegomock generate github.com/runatlantis/atlantis/server/events --package mocks -o mocks/mock_commit_status_updater.go CommitStatusUpdater

// CommitStatusUpdater updates the status of a commit with the VCS host. We set
// the status to signify whether the plan/apply succeeds.
type CommitStatusUpdater interface {
	// UpdateCombined updates the combined status of the head commit of pull.
	// A combined status represents all the projects modified in the pull.
	UpdateCombined(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest, status models.CommitStatus, cmdName command.Name) error
	// UpdateCombinedCount updates the combined status to reflect the counts
	// of project outcomes. counts.NoChanges is only meaningful for command.Apply.
	UpdateCombinedCount(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest, status models.CommitStatus, cmdName command.Name, counts models.ProjectCounts) error

	UpdatePreWorkflowHook(logger logging.SimpleLogging, pull models.PullRequest, status models.CommitStatus, hookDescription string, runtimeDescription string, url string) error
	UpdatePostWorkflowHook(logger logging.SimpleLogging, pull models.PullRequest, status models.CommitStatus, hookDescription string, runtimeDescription string, url string) error
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

func (d *DefaultCommitStatusUpdater) UpdateCombined(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest, status models.CommitStatus, cmdName command.Name) error {
	src := truncateContext(fmt.Sprintf("%s/%s", d.StatusName, cmdName.String()))
	var descripWords string
	switch status {
	case models.PendingCommitStatus:
		descripWords = genProjectStatusDescription(cmdName.String(), "in progress...")
	case models.FailedCommitStatus:
		descripWords = genProjectStatusDescription(cmdName.String(), "failed.")
	case models.SuccessCommitStatus:
		descripWords = genProjectStatusDescription(cmdName.String(), "succeeded.")
	}
	return d.Client.UpdateStatus(logger, repo, pull, status, src, descripWords, "")
}

func (d *DefaultCommitStatusUpdater) UpdateCombinedCount(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest, status models.CommitStatus, cmdName command.Name, counts models.ProjectCounts) error {
	src := truncateContext(fmt.Sprintf("%s/%s", d.StatusName, cmdName.String()))

	var descrip string
	switch cmdName {
	case command.Apply:
		switch {
		case status == models.FailedCommitStatus:
			descrip = fmt.Sprintf("%d/%d projects failed to apply.", counts.Errored, counts.Total)
		case status == models.PendingCommitStatus:
			applied := counts.Success - counts.NoChanges
			if counts.NoChanges > 0 {
				descrip = fmt.Sprintf("%d/%d projects applied (%d up to date).", applied, counts.Total, counts.NoChanges)
			} else {
				descrip = fmt.Sprintf("%d/%d projects applied.", applied, counts.Total)
			}
		case counts.NoChanges > 0 && counts.NoChanges == counts.Total:
			descrip = fmt.Sprintf("%d/%d projects up to date.", counts.NoChanges, counts.Total)
		case counts.NoChanges > 0 && counts.Success > counts.NoChanges:
			applied := counts.Success - counts.NoChanges
			descrip = fmt.Sprintf("%d/%d projects applied successfully (%d up to date).", applied, counts.Total, counts.NoChanges)
		default:
			descrip = fmt.Sprintf("%d/%d projects applied successfully.", counts.Success, counts.Total)
		}
	case command.Plan:
		switch status {
		case models.FailedCommitStatus:
			descrip = fmt.Sprintf("%d/%d projects failed to plan.", counts.Total-counts.Success, counts.Total)
		case models.PendingCommitStatus:
			descrip = fmt.Sprintf("%d/%d projects planned.", counts.Success, counts.Total)
		default:
			descrip = fmt.Sprintf("%d/%d projects planned successfully.", counts.Success, counts.Total)
		}
	case command.PolicyCheck:
		switch status {
		case models.FailedCommitStatus:
			descrip = fmt.Sprintf("%d/%d projects failed policy checks.", counts.Errored, counts.Total)
		case models.PendingCommitStatus:
			descrip = fmt.Sprintf("%d/%d projects had policies checked.", counts.Success, counts.Total)
		default:
			descrip = fmt.Sprintf("%d/%d projects had policies checked successfully.", counts.Success, counts.Total)
		}
	default:
		switch status {
		case models.FailedCommitStatus:
			descrip = fmt.Sprintf("%d/%d projects failed.", counts.Total-counts.Success, counts.Total)
		case models.PendingCommitStatus:
			descrip = fmt.Sprintf("%d/%d projects completed.", counts.Success, counts.Total)
		default:
			descrip = fmt.Sprintf("%d/%d projects succeeded.", counts.Success, counts.Total)
		}
	}

	return d.Client.UpdateStatus(logger, repo, pull, status, src, descrip, "")
}

func (d *DefaultCommitStatusUpdater) UpdateProject(ctx command.ProjectContext, cmdName command.Name, status models.CommitStatus, url string, result *command.ProjectCommandOutput) error {
	if ctx.SuppressVCSStatus {
		return nil
	}

	src := truncateContext(fmt.Sprintf("%s/%s: %s", d.StatusName, cmdName.String(), ctx.ProjectID()))
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
	return d.Client.UpdateStatus(ctx.Log, ctx.BaseRepo, ctx.Pull, status, src, descripWords, url)
}

func genProjectStatusDescription(cmdName, description string) string {
	return fmt.Sprintf("%s %s", cases.Title(language.English).String(cmdName), description)
}

// maxStatusContext is the maximum number of characters allowed by the GitHub
// Statuses API for the "context" field. Exceeding this limit causes a 422.
// See https://docs.github.com/en/rest/commits/statuses
const (
	maxStatusContext            = 255
	statusContextHashLength     = 12
	statusContextHashPrefixSize = statusContextHashLength / 2
)

// truncateContext shortens s to maxStatusContext characters if needed while
// preserving uniqueness for contexts that share the same long prefix.
func truncateContext(s string) string {
	if utf8.RuneCountInString(s) <= maxStatusContext {
		return s
	}
	hash := sha256.Sum256([]byte(s))
	suffix := fmt.Sprintf("-%x", hash[:statusContextHashPrefixSize])
	prefixLength := maxStatusContext - utf8.RuneCountInString(suffix)
	return string([]rune(s)[:prefixLength]) + suffix
}

func (d *DefaultCommitStatusUpdater) UpdatePreWorkflowHook(log logging.SimpleLogging, pull models.PullRequest, status models.CommitStatus, hookDescription string, runtimeDescription string, url string) error {
	return d.updateWorkflowHook(log, pull, status, hookDescription, runtimeDescription, "pre_workflow_hook", url)
}

func (d *DefaultCommitStatusUpdater) UpdatePostWorkflowHook(log logging.SimpleLogging, pull models.PullRequest, status models.CommitStatus, hookDescription string, runtimeDescription string, url string) error {
	return d.updateWorkflowHook(log, pull, status, hookDescription, runtimeDescription, "post_workflow_hook", url)
}

func (d *DefaultCommitStatusUpdater) updateWorkflowHook(log logging.SimpleLogging, pull models.PullRequest, status models.CommitStatus, hookDescription string, runtimeDescription string, workflowType string, url string) error {
	src := truncateContext(fmt.Sprintf("%s/%s: %s", d.StatusName, workflowType, hookDescription))

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

	return d.Client.UpdateStatus(log, pull.BaseRepo, pull, status, src, descripWords, url)
}
