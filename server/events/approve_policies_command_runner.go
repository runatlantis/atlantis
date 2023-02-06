package events

import (
	"fmt"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/hashicorp/go-multierror"
	"encoding/json"
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

	if err := a.commitStatusUpdater.UpdateCombined(baseRepo, pull, models.PendingCommitStatus, command.PolicyCheck); err != nil {
		ctx.Log.Warn("unable to update commit status: %s", err)
	}

	projectCmds, err := a.prjCmdBuilder.BuildApprovePoliciesCommands(ctx, cmd)
	if err != nil {
		if statusErr := a.commitStatusUpdater.UpdateCombined(ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, command.PolicyCheck); statusErr != nil {
			ctx.Log.Warn("unable to update commit status: %s", statusErr)
		}
		a.pullUpdater.updatePull(ctx, cmd, command.Result{Error: err})
		return
	}

	if len(projectCmds) == 0 && a.SilenceNoProjects {
		ctx.Log.Info("determined there was no project to run approve_policies in")
		if !a.silenceVCSStatusNoProjects {
			// If there were no projects modified, we set successful commit statuses
			// with 0/0 projects approve_policies successfully because some users require
			// the Atlantis status to be passing for all pull requests.
			ctx.Log.Debug("setting VCS status to success with no projects found")
			if err := a.commitStatusUpdater.UpdateCombinedCount(ctx.Pull.BaseRepo, ctx.Pull, models.SuccessCommitStatus, command.PolicyCheck, 0, 0); err != nil {
				ctx.Log.Warn("unable to update commit status: %s", err)
			}
		}
		return
	}

	result := a.buildApprovePolicyCommandResults(ctx, projectCmds)
	r, _ := json.Marshal(result)
	ctx.Log.Err(string(r))

	a.pullUpdater.updatePull(
		ctx,
		cmd,
		result,
	)

	pullStatus, err := a.dbUpdater.updateDB(ctx, pull, result.ProjectResults)
	if err != nil {
		ctx.Log.Err("writing results: %s", err)
		return
	}

	a.updateCommitStatus(ctx, pullStatus)
}

func (a *ApprovePoliciesCommandRunner) buildApprovePolicyCommandResults(ctx *command.Context, prjCmds []command.ProjectContext) (result command.Result) {
	// Check if vcs user is in the top-level owner list of the PolicySets. All projects
	// share the same Owners list at this time so no reason to iterate over each
	// project.
    var prjResults []command.ProjectResult
	if len(prjCmds) > 0 {
		teams := []string{}

		// Only query the users team membership if any teams have been configured as owners.
		if prjCmds[0].PolicySets.HasTeamOwners() {
			userTeams, err := a.vcsClient.GetTeamNamesForUser(ctx.Pull.BaseRepo, ctx.User)
			if err != nil {
				ctx.Log.Err("unable to get team membership for user: %s", err)
				return
			}
			teams = append(teams, userTeams...)
		}
		isAdmin := prjCmds[0].PolicySets.Owners.IsOwner(ctx.User.Username, teams)

    	for _, prjCmd := range prjCmds {
    	    var prjErrs error
    	    var prjPolicyStatus []models.PolicySetApproval
    	    // Grab policy set status for project
    	    for _, prjPullStatus := range ctx.PullStatus.Projects {
    	        if prjCmd.Workspace == prjPullStatus.Workspace &&
    	            prjCmd.RepoRelDir == prjPullStatus.RepoRelDir &&
    	            prjCmd.ProjectName == prjPullStatus.ProjectName {
    	            prjPolicyStatus = prjPullStatus.PolicyStatus
    	        }
    	    }

    	    for _, policySet := range prjCmd.PolicySets.PolicySets {
    	        isOwner := policySet.Owners.IsOwner(ctx.User.Username, teams) || isAdmin
    	        for i, policyStatus := range prjPolicyStatus {
    	            if policySet.Name == policyStatus.PolicySetName {
    	                if policyStatus.Approvals == 0 {
    	                    continue
    	                }
    	                if isOwner {
    	                    prjPolicyStatus[i].Approvals = policyStatus.Approvals + 1
    	                } else {
    	                    prjErrs = multierror.Append(fmt.Errorf("Policy set: %s user %s is not a policy owner. Please contact policy owners to approve failing policies", policySet.Name, ctx.User.Username))
    	                }
    	                if prjPolicyStatus[i].Approvals != 0 {
    	                    prjErrs = multierror.Append(prjErrs, fmt.Errorf("Policy set: %s requires %d approvals, have %d.", policySet.Name, policySet.ReviewCount, (0-prjPolicyStatus[i].Approvals)))
    	                }
    	            }
    	        }
    	    }

    		//prjResult := a.prjCmdRunner.ApprovePolicies(prjCmd)
    		prjResult := command.ProjectResult{
                         		Command:              command.PolicyCheck,
                         		Failure:              "",
                         		Error:                prjErrs,
                         		PolicyCheckSuccess:   nil,
                         		PolicyCheckApprovals: prjPolicyStatus,
                         		RepoRelDir:           prjCmd.RepoRelDir,
                         		Workspace:            prjCmd.Workspace,
                         		ProjectName:          prjCmd.ProjectName,
                         	}
    	    prjResults = append(prjResults, prjResult)
    	    // TESTING. REMOVE ME.
            j, _ := json.Marshal(prjResult)
            ctx.Log.Err(string(j))
    	}
	}
    result.ProjectResults = prjResults
	return
}

func (a *ApprovePoliciesCommandRunner) updateCommitStatus(ctx *command.Context, pullStatus models.PullStatus) {
	var numSuccess int
	var numErrored int
	status := models.SuccessCommitStatus

	numSuccess = pullStatus.StatusCount(models.PassedPolicyCheckStatus)
	numErrored = pullStatus.StatusCount(models.ErroredPolicyCheckStatus)

	if numErrored > 0 {
		status = models.FailedCommitStatus
	}

	if err := a.commitStatusUpdater.UpdateCombinedCount(ctx.Pull.BaseRepo, ctx.Pull, status, command.PolicyCheck, numSuccess, len(pullStatus.Projects)); err != nil {
		ctx.Log.Warn("unable to update commit status: %s", err)
	}
}
