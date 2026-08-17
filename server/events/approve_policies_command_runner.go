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
	baseRepo := ctx.Pull.BaseRepo
	pull := ctx.Pull
	publicationToken, err := a.acquirePlanPublicationClaim(ctx)
	if err != nil {
		return
	}

	projectCmds, err := a.prjCmdBuilder.BuildApprovePoliciesCommands(ctx, cmd)
	if MarkCommandSkippedIfIgnoredTargetedDir(ctx, cmd.CommandName(), err) {
		a.finishPlanPublicationClaim(ctx, publicationToken, nil)
		return
	}
	if err != nil {
		a.finishPlanPublicationClaim(ctx, publicationToken, errors.Join(
			a.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, command.PolicyCheck),
			a.pullUpdater.updatePull(ctx, cmd, command.Result{Error: err}),
		))
		return
	}

	if len(projectCmds) == 0 && a.SilenceNoProjects {
		ctx.Log.Info("determined there was no project to run approve_policies in")
		if !a.silenceVCSStatusNoProjects {
			// If there were no projects modified, we set successful commit statuses
			// with 0/0 projects approve_policies successfully because some users require
			// the Atlantis status to be passing for all pull requests.
			ctx.Log.Debug("setting VCS status to success with no projects found")
			if err := a.commitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.SuccessCommitStatus, command.PolicyCheck, models.ProjectCounts{}); err != nil {
				ctx.Log.Warn("unable to update commit status: %s", err)
				a.finishPlanPublicationClaim(ctx, publicationToken, err)
				return
			}
		}
		a.finishPlanPublicationClaim(ctx, publicationToken, nil)
		return
	}

	result := runProjectCmds(projectCmds, a.prjCmdRunner.ApprovePolicies)

	pullStatus, err := a.dbUpdater.updatePolicyResultsForPlanGeneration(ctx, pull, result.ProjectResults, publicationToken)
	if err != nil {
		if db.IsPlanGenerationObsolete(err) {
			ctx.Log.Info("approve_policies results were superseded by a newer plan generation; ignoring obsolete results")
			ctx.CommandHasErrors = false
			a.finishPlanPublicationClaim(ctx, publicationToken, nil)
			return
		}
		ctx.Log.Err("writing results: %s", err)
		ctx.CommandHasErrors = true
		result.Error = fmt.Errorf("writing approve_policies results: %w", err)
		publicationErr := errors.Join(
			a.commitStatusUpdater.UpdateCombined(ctx.Log, baseRepo, pull, models.FailedCommitStatus, command.PolicyCheck),
			a.pullUpdater.updatePull(ctx, cmd, result),
		)
		a.finishPlanPublicationClaim(ctx, publicationToken, errors.Join(publicationErr, err))
		return
	}
	ctx.PullStatus = &pullStatus

	a.finishPlanPublicationClaim(ctx, publicationToken, errors.Join(
		a.pullUpdater.updatePull(ctx, cmd, result),
		a.updateCommitStatus(ctx, pullStatus),
	))
}

func (a *ApprovePoliciesCommandRunner) acquirePlanPublicationClaim(ctx *command.Context) (string, error) {
	token, err := a.dbUpdater.acquirePlanPublicationClaim(ctx.Pull)
	if err == nil {
		return token, nil
	}
	ctx.CommandHasErrors = true
	ctx.Log.Err("acquiring plan publication claim for policy approval: %s", err)
	return "", err
}

func (a *ApprovePoliciesCommandRunner) finishPlanPublicationClaim(ctx *command.Context, token string, publicationErr error) bool {
	if publicationErr != nil {
		ctx.CommandHasErrors = true
		ctx.Log.Err("publishing policy approval state; publication claim retained for offline recovery: %s", publicationErr)
		return false
	}
	if err := a.dbUpdater.releasePlanPublicationClaim(ctx.Pull, token); err != nil {
		ctx.CommandHasErrors = true
		ctx.Log.Err("releasing plan publication claim after policy approval: %s", err)
		return false
	}
	return true
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

	if err := a.commitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, status, command.PolicyCheck, models.ProjectCounts{Success: numSuccess, Total: countActivePullStatusProjects(pullStatus), Errored: numErrored}); err != nil {
		ctx.Log.Warn("unable to update commit status: %s", err)
		return err
	}
	return nil
}
