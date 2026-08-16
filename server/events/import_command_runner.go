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
	"github.com/runatlantis/atlantis/server/events/vcs"
)

func NewImportCommandRunner(
	pullUpdater *PullUpdater,
	dbUpdater *DBUpdater,
	pullReqStatusFetcher vcs.PullReqStatusFetcher,
	prjCmdBuilder ProjectImportCommandBuilder,
	prjCmdRunner ProjectImportCommandRunner,
	SilenceNoProjects bool,
) *ImportCommandRunner {
	return &ImportCommandRunner{
		pullUpdater:          pullUpdater,
		dbUpdater:            dbUpdater,
		pullReqStatusFetcher: pullReqStatusFetcher,
		prjCmdBuilder:        prjCmdBuilder,
		prjCmdRunner:         prjCmdRunner,
		SilenceNoProjects:    SilenceNoProjects,
	}
}

type ImportCommandRunner struct {
	pullUpdater          *PullUpdater
	dbUpdater            *DBUpdater
	pullReqStatusFetcher vcs.PullReqStatusFetcher
	prjCmdBuilder        ProjectImportCommandBuilder
	prjCmdRunner         ProjectImportCommandRunner
	SilenceNoProjects    bool
}

func (v *ImportCommandRunner) Run(ctx *command.Context, cmd *CommentCommand) {
	var err error
	// Get the mergeable status before we set any build statuses of our own.
	// We do this here because when we set a "Pending" status, if users have
	// required the Atlantis status checks to pass, then we've now changed
	// the mergeability status of the pull request.
	// This sets the approved, mergeable, and sqlocked status in the context.
	ctx.PullRequestStatus, err = v.pullReqStatusFetcher.FetchPullStatus(ctx.Log, ctx.Pull)
	if err != nil {
		// On error we continue the request with mergeable assumed false.
		// We want to continue because not all import will need this status,
		// only if they rely on the mergeability requirement.
		// All PullRequestStatus fields are set to false by default when error.
		ctx.Log.Warn("unable to get pull request status: %s. Continuing with mergeable and approved assumed false", err)
	}
	publicationToken, err := v.dbUpdater.acquirePlanPublicationClaim(ctx.Pull)
	if err != nil {
		ctx.CommandHasErrors = true
		ctx.Log.Err("acquiring plan publication claim for import: %s", err)
		return
	}
	claimActive := true
	finishClaim := func(publicationErr error) {
		if !claimActive || publicationErr != nil {
			return
		}
		if releaseErr := v.dbUpdater.releasePlanPublicationClaim(ctx.Pull, publicationToken); releaseErr != nil {
			ctx.CommandHasErrors = true
			ctx.Log.Err("releasing plan publication claim for import: %s", releaseErr)
			return
		}
		claimActive = false
	}
	defer func() {
		if claimActive {
			ctx.Log.Warn("retaining plan publication claim after ambiguous import publication")
		}
	}()

	currentPullStatus, statusErr := v.dbUpdater.Database.GetPullStatus(ctx.Pull)
	if statusErr != nil {
		ctx.CommandHasErrors = true
		ctx.Log.Err("reading durable plan status before import: %s", statusErr)
		finishClaim(nil)
		return
	}
	if ctx.PullStatus != nil && !reflect.DeepEqual(ctx.PullStatus, currentPullStatus) {
		ctx.CommandHasErrors = true
		ctx.Log.Info("import command belongs to an obsolete plan generation, skipping execution")
		finishClaim(nil)
		return
	}
	ctx.PullStatus = currentPullStatus

	var projectCmds []command.ProjectContext
	projectCmds, err = v.prjCmdBuilder.BuildImportCommands(ctx, cmd)
	if MarkCommandSkippedIfIgnoredTargetedDir(ctx, cmd.CommandName(), err) {
		finishClaim(nil)
		return
	}
	if err != nil {
		ctx.Log.Warn("Error %s", err)
	}
	if project, active := activePlanGenerationTarget(ctx.PullStatus, projectCmds); active {
		ctx.CommandHasErrors = true
		result := command.Result{Failure: fmt.Sprintf(
			"cannot run import while plan generation is in progress for project %q in dir %q and workspace %q; wait for the plan to finish and try again",
			project.ProjectName,
			project.RepoRelDir,
			project.Workspace,
		)}
		publicationErr := v.pullUpdater.updatePull(ctx, cmd, result)
		if publicationErr != nil {
			ctx.CommandHasErrors = true
			ctx.Log.Warn("unable to publish import result: %s", publicationErr)
		}
		finishClaim(publicationErr)
		return
	}

	if len(projectCmds) == 0 && v.SilenceNoProjects {
		ctx.Log.Info("determined there was no project to run import in.")
		finishClaim(nil)
		return
	}
	var result command.Result
	if len(projectCmds) > 1 {
		// There is no usecase to kick terraform import into multiple projects.
		// To avoid incorrect import, suppress to execute terraform import in multiple projects.
		result = command.Result{
			Failure: "import cannot run on multiple projects. please specify one project.",
		}
	} else {
		result = runProjectCmds(projectCmds, v.prjCmdRunner.Import)
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
			ctx.Log.Warn("import result was superseded by a newer plan generation; durable plan state was not changed")
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
		ctx.Log.Warn("unable to publish import result: %s", publicationErr)
	}
	finishClaim(errors.Join(publicationErr, persistenceErr))
}

func (v *ImportCommandRunner) ShouldSkipPreWorkflowHooks(ctx *command.Context, cmd *CommentCommand) bool {
	return MarkCommandSkippedIfIgnoredTarget(ctx, cmd.CommandName(), cmd, v.prjCmdBuilder)
}
