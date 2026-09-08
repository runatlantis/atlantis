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
	Publication         *PublicationCoordinator
	LivePullHeadFetcher LivePullHeadFetcher
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
		err := p.Publication.RunWithObservedStatus(ctx, p.dbUpdater.Database, p.LivePullHeadFetcher, func() error {
			ctx.Log.Info("no projects to run policy_check in")
			if p.silenceVCSStatusNoProjects {
				return nil
			}
			return publishTerminal(ctx, func() error {
				return p.commitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.SuccessCommitStatus, command.PolicyCheck, models.ProjectCounts{})
			})
		})
		if err != nil {
			p.publicationFailed(ctx, err)
		}
		return
	}

	startErr := p.Publication.RunCommand(ctx, func() error {
		if err := p.refreshPublicationIdentity(ctx, cmds); err != nil {
			return err
		}

		return publishTerminal(ctx, func() error {
			return p.commitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.PendingCommitStatus, command.PolicyCheck)
		})
	})
	if startErr != nil {
		p.publicationFailed(ctx, startErr)
		return
	}
	// Policy execution, like Terraform, owns no publication lease.
	var result command.Result
	if p.isParallelEnabled(cmds) {
		ctx.Log.Info("Running policy_checks in parallel")
		result = runProjectCmdsParallel(cmds, p.prjCmdRunner.PolicyCheck, p.parallelPoolSize, nil, ctx.Pull)
	} else {
		result = runProjectCmds(cmds, p.prjCmdRunner.PolicyCheck)
	}

	err := p.Publication.RunCommand(ctx, func() error {
		if err := p.refreshPublicationIdentity(ctx, cmds); err != nil {
			return err
		}
		pullStatus, err := p.dbUpdater.updateDB(ctx, ctx.Pull, result.ProjectResults)
		if err != nil {
			return fmt.Errorf("writing policy results: %w", err)
		}
		// Durable generation validation must precede both comments and statuses.
		if result.HasErrors() || !p.quietPolicyChecks {
			if err := p.pullUpdater.updatePull(ctx, PolicyCheckCommand{}, result); err != nil {
				return err
			}
		}
		return p.updateCommitStatus(ctx, pullStatus)
	})
	if err != nil {
		p.publicationFailed(ctx, err)
	}
}

func (p *PolicyCheckCommandRunner) publicationFailed(ctx *command.Context, err error) {
	ctx.CommandHasErrors = true
	if errors.Is(err, db.ErrPlanGenerationSuperseded) {
		ctx.Log.Warn("suppressing obsolete policy publication %v", err)
		return
	}
	ctx.Log.Err("publishing policy results %v", err)
	if reportErr := p.pullUpdater.updatePull(ctx, PolicyCheckCommand{}, command.Result{Error: err}); reportErr != nil {
		ctx.Log.Err("reporting command result: %s", reportErr)
	}
}

func (p *PolicyCheckCommandRunner) refreshPublicationIdentity(ctx *command.Context, projects []command.ProjectContext) error {
	if ctx.TerminalPublisher == nil {
		return nil
	}
	if p.LivePullHeadFetcher != nil {
		live, err := p.LivePullHeadFetcher.GetLivePullIdentity(command.ProjectContext{Log: ctx.Log, Pull: ctx.Pull, PullStatus: ctx.PullStatus, API: ctx.API})
		if err != nil {
			return err
		}
		if err := validateCommandStartIdentity(command.ProjectContext{Pull: ctx.Pull}, live); err != nil {
			return fmt.Errorf("%w: %w", db.ErrPlanGenerationSuperseded, err)
		}
	}
	status, err := p.dbUpdater.Database.GetPullStatus(ctx.Pull)
	if err != nil {
		return err
	}
	for _, project := range projects {
		stored := findProjectInPullStatus(status, project.Workspace, project.RepoRelDir, project.ProjectName)
		if stored == nil || !pullStatusFreshForPull(ctx.Pull, status.Pull) || stored.PlanGeneration != project.PlanGeneration || stored.PlanGenerationActive || stored.AcceptedPlanGeneration != project.AcceptedPlanGeneration {
			return db.ErrPlanGenerationSuperseded
		}
	}
	return nil
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

	return publishTerminal(ctx, func() error {
		return p.commitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, status, command.PolicyCheck, models.ProjectCounts{Success: numSuccess, Total: len(pullStatus.Projects), Errored: numErrored})
	})
}

func (p *PolicyCheckCommandRunner) isParallelEnabled(cmds []command.ProjectContext) bool {
	return len(cmds) > 0 && cmds[0].ParallelPolicyCheckEnabled
}
