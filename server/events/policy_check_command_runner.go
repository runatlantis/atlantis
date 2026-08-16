// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"errors"
	"fmt"

	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

func NewPolicyCheckCommandRunner(
	dbUpdater *DBUpdater,
	pullUpdater *PullUpdater,
	commitStatusUpdater CommitStatusUpdater,
	projectCommandRunner ProjectPolicyCheckCommandRunner,
	parallelPoolSize int,
	silenceVCSStatusNoProjects bool,
	quietPolicyChecks bool,
) *PolicyCheckCommandRunner {
	return &PolicyCheckCommandRunner{
		dbUpdater:                  dbUpdater,
		pullUpdater:                pullUpdater,
		commitStatusUpdater:        commitStatusUpdater,
		prjCmdRunner:               projectCommandRunner,
		parallelPoolSize:           parallelPoolSize,
		silenceVCSStatusNoProjects: silenceVCSStatusNoProjects,
		quietPolicyChecks:          quietPolicyChecks,
	}
}

type PolicyCheckCommandRunner struct {
	dbUpdater           *DBUpdater
	pullUpdater         *PullUpdater
	commitStatusUpdater CommitStatusUpdater
	prjCmdRunner        ProjectPolicyCheckCommandRunner
	parallelPoolSize    int
	// SilenceVCSStatusNoProjects is whether any plan should set commit status if no projects
	// are found
	silenceVCSStatusNoProjects bool
	quietPolicyChecks          bool
}

func (p *PolicyCheckCommandRunner) Run(ctx *command.Context, cmds []command.ProjectContext) {
	if len(cmds) == 0 {
		ctx.Log.Info("no projects to run policy_check in")
		return
	}

	var result command.Result
	if p.isParallelEnabled(cmds) {
		ctx.Log.Info("Running policy_checks in parallel")
		result = runProjectCmdsParallel(cmds, p.prjCmdRunner.PolicyCheck, p.parallelPoolSize, nil, ctx.Pull)
	} else {
		result = runProjectCmds(cmds, p.prjCmdRunner.PolicyCheck)
	}

	publicationToken, err := p.waitForPlanPublicationClaim(ctx)
	if err != nil {
		return
	}
	pullStatus, err := p.dbUpdater.updatePolicyResultsForPlanGeneration(ctx, ctx.Pull, result.ProjectResults, publicationToken)
	if err != nil {
		if db.IsPlanGenerationObsolete(err) {
			ctx.Log.Info("policy check was superseded by a newer plan generation; ignoring obsolete results")
			ctx.CommandHasErrors = false
			p.finishPlanPublicationClaim(ctx, publicationToken, nil)
			return
		}
		ctx.Log.Err("writing results: %s", err)
		ctx.CommandHasErrors = true
		result.Error = fmt.Errorf("writing policy check results: %w", err)
		publicationErr := errors.Join(
			p.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, command.PolicyCheck),
			p.pullUpdater.updatePull(ctx, PolicyCheckCommand{}, result),
		)
		p.finishPlanPublicationClaim(ctx, publicationToken, errors.Join(publicationErr, err))
		return
	}
	ctx.PullStatus = &pullStatus
	// Quiet policy checks unless there's an error.
	var commentErr error
	if result.HasErrors() || !p.quietPolicyChecks {
		commentErr = p.pullUpdater.updatePull(ctx, PolicyCheckCommand{}, result)
	}
	p.finishPlanPublicationClaim(ctx, publicationToken, errors.Join(commentErr, p.updateCommitStatus(ctx, pullStatus)))
}

func (p *PolicyCheckCommandRunner) waitForPlanPublicationClaim(ctx *command.Context) (string, error) {
	token, err := p.dbUpdater.waitForPlanPublicationClaim(ctx.Pull)
	if err == nil {
		return token, nil
	}
	ctx.CommandHasErrors = true
	ctx.Log.Err("waiting for plan publication claim for policy results: %s", err)
	return "", err
}

func (p *PolicyCheckCommandRunner) finishPlanPublicationClaim(ctx *command.Context, token string, publicationErr error) bool {
	if publicationErr != nil {
		ctx.CommandHasErrors = true
		ctx.Log.Err("publishing policy state; publication claim retained for offline recovery: %s", publicationErr)
		return false
	}
	if err := p.dbUpdater.releasePlanPublicationClaim(ctx.Pull, token); err != nil {
		ctx.CommandHasErrors = true
		ctx.Log.Err("releasing plan publication claim after policy results: %s", err)
		return false
	}
	return true
}

func (p *PolicyCheckCommandRunner) updateCommitStatus(ctx *command.Context, pullStatus models.PullStatus) error {
	var numSuccess int
	var numErrored int
	status := models.SuccessCommitStatus

	numSuccess = pullStatus.StatusCount(models.PassedPolicyCheckStatus)
	numErrored = pullStatus.StatusCount(models.ErroredPolicyCheckStatus)

	if numErrored > 0 {
		status = models.FailedCommitStatus
	}

	if err := p.commitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, status, command.PolicyCheck, models.ProjectCounts{Success: numSuccess, Total: countActivePullStatusProjects(pullStatus), Errored: numErrored}); err != nil {
		ctx.Log.Warn("unable to update commit status: %s", err)
		return err
	}
	return nil
}

func (p *PolicyCheckCommandRunner) isParallelEnabled(cmds []command.ProjectContext) bool {
	return len(cmds) > 0 && cmds[0].ParallelPolicyCheckEnabled
}
