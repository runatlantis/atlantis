package events

import (
	"fmt"
	"github.com/runatlantis/atlantis/server/lyft/feature"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

type commandOutputGenerator interface {
	GeneratePolicyCheckOutputStore(ctx *command.Context, cmd *command.Comment) (command.PolicyCheckOutputStore, error)
}

// TODO: delete entire approve policies workflow when policy v2 rollout is complete
func NewApprovePoliciesCommandRunner(
	vcsStatusUpdater VCSStatusUpdater,
	prjCommandBuilder ProjectApprovePoliciesCommandBuilder,
	prjCommandRunner ProjectApprovePoliciesCommandRunner,
	outputUpdater OutputUpdater,
	dbUpdater *DBUpdater,
	policyCheckOutputGenerator commandOutputGenerator,
	allocator feature.Allocator) *ApprovePoliciesCommandRunner {
	return &ApprovePoliciesCommandRunner{
		vcsStatusUpdater:           vcsStatusUpdater,
		prjCmdBuilder:              prjCommandBuilder,
		prjCmdRunner:               prjCommandRunner,
		outputUpdater:              outputUpdater,
		dbUpdater:                  dbUpdater,
		policyCheckOutputGenerator: policyCheckOutputGenerator,
		allocator:                  allocator,
	}
}

type ApprovePoliciesCommandRunner struct {
	vcsStatusUpdater           VCSStatusUpdater
	outputUpdater              OutputUpdater
	dbUpdater                  *DBUpdater
	prjCmdBuilder              ProjectApprovePoliciesCommandBuilder
	prjCmdRunner               ProjectApprovePoliciesCommandRunner
	policyCheckOutputGenerator commandOutputGenerator
	allocator                  feature.Allocator
}

func (a *ApprovePoliciesCommandRunner) Run(ctx *command.Context, cmd *command.Comment) {
	baseRepo := ctx.Pull.BaseRepo
	pull := ctx.Pull
	shouldAllocate, err := a.allocator.ShouldAllocate(feature.PolicyV2, feature.FeatureContext{
		RepoName: baseRepo.FullName,
	})
	if err != nil {
		ctx.Log.ErrorContext(ctx.RequestCtx, "unable to allocate policy v2, continuing with legacy mode")
	}
	if shouldAllocate {
		ctx.Log.ErrorContext(ctx.RequestCtx, "policy v2 mode doesn't support atlantis approve_policies command")
		return
	}

	statusID, err := a.vcsStatusUpdater.UpdateCombined(ctx.RequestCtx, baseRepo, pull, models.PendingVCSStatus, command.PolicyCheck, "", "")
	if err != nil {
		ctx.Log.WarnContext(ctx.RequestCtx, fmt.Sprintf("unable to update commit status: %s", err))
	}

	projectCmds, err := a.prjCmdBuilder.BuildApprovePoliciesCommands(ctx, cmd)
	if err != nil {
		if _, statusErr := a.vcsStatusUpdater.UpdateCombined(ctx.RequestCtx, ctx.Pull.BaseRepo, ctx.Pull, models.FailedVCSStatus, command.PolicyCheck, statusID, ""); statusErr != nil {
			ctx.Log.WarnContext(ctx.RequestCtx, fmt.Sprintf("unable to update commit status: %s", statusErr))
		}
		a.outputUpdater.UpdateOutput(ctx, cmd, command.Result{Error: err})
		return
	}

	if len(projectCmds) == 0 {
		ctx.Log.InfoContext(ctx.RequestCtx, "determined there was no project to run approve_policies in")
		// If there were no projects modified, we set successful commit statuses
		// with 0/0 projects approve_policies successfully because some users require
		// the Atlantis status to be passing for all pull requests.
		if _, err := a.vcsStatusUpdater.UpdateCombinedCount(ctx.RequestCtx, ctx.Pull.BaseRepo, ctx.Pull, models.SuccessVCSStatus, command.PolicyCheck, 0, 0, statusID); err != nil {
			ctx.Log.WarnContext(ctx.RequestCtx, fmt.Sprintf("unable to update commit status: %s", err))
		}
		return
	}

	result := a.buildApprovePolicyCommandResults(ctx, projectCmds)

	// Adds the policy check output for failing policies which needs to be populated when using github checks
	// Noop when github checks is not enabled.
	policyCheckOutputStore, err := a.policyCheckOutputGenerator.GeneratePolicyCheckOutputStore(ctx, cmd)
	if err != nil {
		if _, statusErr := a.vcsStatusUpdater.UpdateCombined(ctx.RequestCtx, ctx.Pull.BaseRepo, ctx.Pull, models.FailedVCSStatus, command.PolicyCheck, statusID, ""); statusErr != nil {
			ctx.Log.WarnContext(ctx.RequestCtx, fmt.Sprintf("unable to update commit status: %s", statusErr))
		}
		a.outputUpdater.UpdateOutput(ctx, cmd, command.Result{Error: err})
		return
	}

	for i, prjResult := range result.ProjectResults {
		policyCheckOutput := policyCheckOutputStore.Get(prjResult.ProjectName, prjResult.Workspace)
		if policyCheckOutput != nil {
			result.ProjectResults[i].PolicyCheckSuccess = policyCheckOutput
		}
	}

	a.outputUpdater.UpdateOutput(
		ctx,
		cmd,
		result,
	)

	pullStatus, err := a.dbUpdater.updateDB(ctx, pull, result.ProjectResults)
	if err != nil {
		ctx.Log.ErrorContext(ctx.RequestCtx, fmt.Sprintf("writing results: %s", err))
		return
	}

	a.updateVcsStatus(ctx, pullStatus, statusID)
}

func (a *ApprovePoliciesCommandRunner) buildApprovePolicyCommandResults(ctx *command.Context, prjCmds []command.ProjectContext) (result command.Result) {
	// Check if vcs user is in the owner list of the PolicySets. All projects
	// share the same Owners list at this time so no reason to iterate over each
	// project.
	if len(prjCmds) > 0 && !prjCmds[0].PolicySets.IsOwner(ctx.User.Username) {
		result.Error = fmt.Errorf("contact policy owners to approve failing policies")
		return
	}

	var prjResults []command.ProjectResult

	for _, prjCmd := range prjCmds {
		prjResult := a.prjCmdRunner.ApprovePolicies(prjCmd)
		prjResults = append(prjResults, prjResult)
	}
	result.ProjectResults = prjResults
	return
}

func (a *ApprovePoliciesCommandRunner) updateVcsStatus(ctx *command.Context, pullStatus models.PullStatus, statusID string) {
	var numSuccess int
	var numErrored int
	status := models.SuccessVCSStatus

	numSuccess = pullStatus.StatusCount(models.PassedPolicyCheckStatus)
	numErrored = pullStatus.StatusCount(models.ErroredPolicyCheckStatus)

	if numErrored > 0 {
		status = models.FailedVCSStatus
	}

	if _, err := a.vcsStatusUpdater.UpdateCombinedCount(ctx.RequestCtx, ctx.Pull.BaseRepo, ctx.Pull, status, command.PolicyCheck, numSuccess, len(pullStatus.Projects), statusID); err != nil {
		ctx.Log.WarnContext(ctx.RequestCtx, fmt.Sprintf("unable to update commit status: %s", err))
	}
}
