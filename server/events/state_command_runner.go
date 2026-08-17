// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

func NewStateCommandRunner(
	pullUpdater *PullUpdater,
	dbUpdater *DBUpdater,
	prjCmdBuilder ProjectStateCommandBuilder,
	prjCmdRunner ProjectStateCommandRunner,
) *StateCommandRunner {
	return &StateCommandRunner{
		pullUpdater:   pullUpdater,
		dbUpdater:     dbUpdater,
		prjCmdBuilder: prjCmdBuilder,
		prjCmdRunner:  prjCmdRunner,
	}
}

type StateCommandRunner struct {
	pullUpdater   *PullUpdater
	dbUpdater     *DBUpdater
	prjCmdBuilder ProjectStateCommandBuilder
	prjCmdRunner  ProjectStateCommandRunner
}

func (v *StateCommandRunner) Run(ctx *command.Context, cmd *CommentCommand) {
	publicationToken, err := v.dbUpdater.acquirePlanPublicationClaim(ctx.Pull)
	if err != nil {
		ctx.CommandHasErrors = true
		ctx.Log.Err("acquiring plan publication claim for state command: %s", err)
		return
	}
	claimActive := true
	finishClaim := func(publicationErr error) {
		if !claimActive || publicationErr != nil {
			return
		}
		if releaseErr := v.dbUpdater.releasePlanPublicationClaim(ctx.Pull, publicationToken); releaseErr != nil {
			ctx.CommandHasErrors = true
			ctx.Log.Err("releasing plan publication claim for state command: %s", releaseErr)
			return
		}
		claimActive = false
	}
	defer func() {
		if claimActive {
			ctx.Log.Warn("retaining plan publication claim after ambiguous state publication")
		}
	}()

	currentPullStatus, statusErr := v.dbUpdater.Database.GetPullStatus(ctx.Pull)
	if statusErr != nil {
		ctx.CommandHasErrors = true
		ctx.Log.Err("reading durable plan status before state command: %s", statusErr)
		finishClaim(nil)
		return
	}
	if ctx.PullStatus != nil && !reflect.DeepEqual(ctx.PullStatus, currentPullStatus) {
		ctx.CommandHasErrors = true
		ctx.Log.Info("state command belongs to an obsolete plan generation, skipping execution")
		finishClaim(nil)
		return
	}
	ctx.PullStatus = currentPullStatus

	var result command.Result
	switch cmd.SubName {
	case "rm":
		result = v.runRm(ctx, cmd)
	default:
		result = command.Result{
			Failure: fmt.Sprintf("unknown state subcommand %s", cmd.SubName),
		}
	}
	if ctx.CommandSkipped {
		finishClaim(nil)
		return
	}
	pullStatus := models.PullStatus{Pull: ctx.Pull}
	if ctx.PullStatus != nil {
		pullStatus = *ctx.PullStatus
	}
	var persistenceErr error
	successfulResults := successfulDiscardMutationProjectResults(result)
	discardResults := discardTargetsInPullStatus(ctx.PullStatus, successfulResults)
	if len(discardResults) > 0 {
		pullStatus, persistenceErr = v.dbUpdater.updateDiscardResultsForPlanGeneration(ctx, ctx.Pull, discardResults, publicationToken)
	}
	if persistenceErr != nil {
		err = persistenceErr
		if db.IsPlanGenerationObsolete(err) {
			ctx.CommandHasErrors = true
			ctx.Log.Warn("state result was superseded by a newer plan generation; durable plan state was not changed")
			finishClaim(nil)
			return
		}
		result.Error = fmt.Errorf("writing discarded plan status: %w", err)
		ctx.CommandHasErrors = true
	}
	ctx.PullStatus = &pullStatus
	publicationErr := v.pullUpdater.updatePull(ctx, cmd, result)
	if publicationErr != nil {
		ctx.CommandHasErrors = true
		ctx.Log.Warn("unable to publish state result: %s", publicationErr)
	}
	finishClaim(errors.Join(publicationErr, persistenceErr))
}

func (v *StateCommandRunner) runRm(ctx *command.Context, cmd *CommentCommand) command.Result {
	projectCmds, err := v.prjCmdBuilder.BuildStateRmCommands(ctx, cmd)
	if MarkCommandSkippedIfIgnoredTargetedDir(ctx, cmd.CommandName(), err) {
		return command.Result{}
	}
	if err != nil {
		ctx.Log.Warn("Error %s", err)
	}
	if project, active := activePlanGenerationTarget(ctx.PullStatus, projectCmds); active {
		ctx.CommandHasErrors = true
		return command.Result{Failure: fmt.Sprintf(
			"cannot run state rm while plan generation is in progress for project %q in dir %q and workspace %q; wait for the plan to finish and try again",
			project.ProjectName,
			project.RepoRelDir,
			project.Workspace,
		)}
	}
	return runProjectCmds(projectCmds, v.prjCmdRunner.StateRm)
}

func (v *StateCommandRunner) ShouldSkipPreWorkflowHooks(ctx *command.Context, cmd *CommentCommand) bool {
	return MarkCommandSkippedIfIgnoredTarget(ctx, cmd.CommandName(), cmd, v.prjCmdBuilder)
}
