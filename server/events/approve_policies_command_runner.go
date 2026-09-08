// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"errors"
	"fmt"

	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
)

func NewApprovePoliciesCommandRunner(
	commitStatusUpdater CommitStatusUpdater,
	prjCommandBuilder ProjectApprovePoliciesCommandBuilder,
	prjCommandRunner ProjectApprovePoliciesCommandRunner,
	pullUpdater *PullUpdater,
	dbUpdater *DBUpdater,
	SilenceNoProjects bool,
	silenceVCSStatusNoProjects bool,
	vcsClient vcs.Client,
) *ApprovePoliciesCommandRunner {
	return &ApprovePoliciesCommandRunner{
		commitStatusUpdater:        commitStatusUpdater,
		prjCmdBuilder:              prjCommandBuilder,
		prjCmdRunner:               prjCommandRunner,
		pullUpdater:                pullUpdater,
		dbUpdater:                  dbUpdater,
		SilenceNoProjects:          SilenceNoProjects,
		silenceVCSStatusNoProjects: silenceVCSStatusNoProjects,
		vcsClient:                  vcsClient,
	}
}

type ApprovePoliciesCommandRunner struct {
	Publication         *PublicationCoordinator
	LivePullHeadFetcher LivePullHeadFetcher
	commitStatusUpdater CommitStatusUpdater
	pullUpdater         *PullUpdater
	dbUpdater           *DBUpdater
	prjCmdBuilder       ProjectApprovePoliciesCommandBuilder
	prjCmdRunner        ProjectApprovePoliciesCommandRunner
	// SilenceNoProjects is whether Atlantis should respond to PRs if no projects
	// are found
	SilenceNoProjects          bool
	silenceVCSStatusNoProjects bool
	vcsClient                  vcs.Client
}

func (a *ApprovePoliciesCommandRunner) Run(ctx *command.Context, cmd *CommentCommand) {
	// Approval checks read ownership and update policy approvals; they do not
	// execute Terraform. Keep their durable read/modify/publication together.
	err := a.Publication.RunWithObservedStatus(ctx, a.dbUpdater.Database, a.LivePullHeadFetcher, func() error { return a.approve(ctx, cmd) })
	if err != nil {
		ctx.CommandHasErrors = true
		if errors.Is(err, db.ErrPlanGenerationSuperseded) {
			ctx.Log.Warn("suppressing obsolete policy approval publication %v", err)
			return
		}
		if reportErr := a.pullUpdater.updatePull(ctx, cmd, command.Result{Error: err}); reportErr != nil {
			ctx.Log.Err("reporting command result: %s", reportErr)
		}
	}
}

func (a *ApprovePoliciesCommandRunner) approve(ctx *command.Context, cmd *CommentCommand) error {
	projectCmds, err := a.prjCmdBuilder.BuildApprovePoliciesCommands(ctx, cmd)
	if MarkCommandSkippedIfIgnoredTargetedDir(ctx, cmd.CommandName(), err) {
		return nil
	}
	if err != nil {
		if statusErr := publishTerminal(ctx, func() error {
			return a.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, command.PolicyCheck)
		}); statusErr != nil {
			return errors.Join(err, statusErr)
		}
		return err
	}
	if err := publishTerminal(ctx, func() error {
		return a.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.PendingCommitStatus, command.PolicyCheck)
	}); err != nil {
		return err
	}
	if len(projectCmds) == 0 && a.SilenceNoProjects {
		ctx.Log.Info("determined there was no project to run approve_policies in")
		if a.silenceVCSStatusNoProjects {
			return nil
		}
		return publishTerminal(ctx, func() error {
			return a.commitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.SuccessCommitStatus, command.PolicyCheck, models.ProjectCounts{})
		})
	}
	result := runProjectCmds(projectCmds, a.prjCmdRunner.ApprovePolicies)
	pullStatus, err := a.dbUpdater.updateDB(ctx, ctx.Pull, result.ProjectResults)
	if err != nil {
		return fmt.Errorf("writing policy approvals: %w", err)
	}
	if err := a.pullUpdater.updatePull(ctx, cmd, result); err != nil {
		return err
	}
	return a.updateCommitStatus(ctx, pullStatus)
}

func (a *ApprovePoliciesCommandRunner) ShouldSkipPreWorkflowHooks(ctx *command.Context, cmd *CommentCommand) bool {
	return MarkCommandSkippedIfIgnoredTarget(ctx, cmd.CommandName(), cmd, a.prjCmdBuilder)
}

func (a *ApprovePoliciesCommandRunner) updateCommitStatus(ctx *command.Context, pullStatus models.PullStatus) error {
	var numSuccess int
	var numErrored int
	status := models.SuccessCommitStatus

	numSuccess = pullStatus.StatusCount(models.PassedPolicyCheckStatus)
	numErrored = pullStatus.StatusCount(models.ErroredPolicyCheckStatus)

	if numErrored > 0 {
		status = models.FailedCommitStatus
	}

	return publishTerminal(ctx, func() error {
		return a.commitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, status, command.PolicyCheck, models.ProjectCounts{Success: numSuccess, Total: len(pullStatus.Projects), Errored: numErrored})
	})
}
