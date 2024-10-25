package events

import (
	"path/filepath"

	"github.com/google/uuid"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	tally "github.com/uber-go/tally/v4"
)

func NewProjectCommandContextBuilder(policyCheckEnabled bool, commentBuilder CommentBuilder, scope tally.Scope) ProjectCommandContextBuilder {
	projectCommandContextBuilder := &DefaultProjectCommandContextBuilder{
		CommentBuilder: commentBuilder,
	}

	if policyCheckEnabled {
		return &PolicyCheckProjectCommandContextBuilder{
			CommentBuilder:               commentBuilder,
			ProjectCommandContextBuilder: projectCommandContextBuilder,
		}
	}

	return &CommandScopedStatsProjectCommandContextBuilder{
		ProjectCommandContextBuilder: projectCommandContextBuilder,
		ProjectCounter:               scope.Counter("projects"),
	}
}

type ProjectCommandContextBuilder interface {
	// BuildProjectContext builds project command contexts for atlantis commands
	BuildProjectContext(
		ctx *command.Context,
		cmdName command.Name,
		subCmdName string,
		prjCfg valid.MergedProjectCfg,
		commentFlags []string,
		repoDir string,
		automerge, parallelApply, parallelPlan, verbose, abortOnExcecutionOrderFail bool, terraformClient terraform.Client,
	) []command.ProjectContext
}

// CommandScopedStatsProjectCommandContextBuilder ensures that project command context contains a scoped stats
// object relevant to the command it applies to.
type CommandScopedStatsProjectCommandContextBuilder struct {
	ProjectCommandContextBuilder
	// Consciously making this global since it gets flushed periodically anyways
	ProjectCounter tally.Counter
}

// BuildProjectContext builds the context and injects the appropriate command level scope after the fact.
func (cb *CommandScopedStatsProjectCommandContextBuilder) BuildProjectContext(
	ctx *command.Context,
	cmdName command.Name,
	subCmdName string,
	prjCfg valid.MergedProjectCfg,
	commentFlags []string,
	repoDir string,
	automerge, parallelApply, parallelPlan, verbose, abortOnExcecutionOrderFail bool,
	terraformClient terraform.Client,
) (projectCmds []command.ProjectContext) {
	cb.ProjectCounter.Inc(1)

	cmds := cb.ProjectCommandContextBuilder.BuildProjectContext(
		ctx, cmdName, subCmdName, prjCfg, commentFlags, repoDir, automerge, parallelApply, parallelPlan, verbose, abortOnExcecutionOrderFail, terraformClient,
	)

	projectCmds = []command.ProjectContext{}

	for _, cmd := range cmds {

		// specifically use the command name in the context instead of the arg
		// since we can return multiple commands worth of contexts for a given command name arg
		// to effectively pipeline them.
		cmd.Scope = cmd.SetProjectScopeTags(cmd.Scope)
		projectCmds = append(projectCmds, cmd)
	}

	return
}

type DefaultProjectCommandContextBuilder struct {
	CommentBuilder CommentBuilder
}

func (cb *DefaultProjectCommandContextBuilder) BuildProjectContext(
	ctx *command.Context,
	cmdName command.Name,
	subName string,
	prjCfg valid.MergedProjectCfg,
	commentFlags []string,
	repoDir string,
	automerge, parallelApply, parallelPlan, verbose, abortOnExcecutionOrderFail bool,
	terraformClient terraform.Client,
) (projectCmds []command.ProjectContext) {
	ctx.Log.Debug("Building project command context for %s", cmdName)

	var steps []valid.Step
	switch cmdName {
	case command.Plan:
		steps = prjCfg.Workflow.Plan.Steps
	case command.Apply:
		steps = prjCfg.Workflow.Apply.Steps
	case command.Version:
		// Setting statically since there will only be one step
		steps = []valid.Step{{
			StepName: "version",
		}}
	case command.Import:
		steps = prjCfg.Workflow.Import.Steps
	case command.State:
		switch subName {
		case "rm":
			steps = prjCfg.Workflow.StateRm.Steps
		default:
			// comment_parser prevent invalid subcommand, so not need to handle this.
			// if comes here, state_command_runner will respond on PR, so it's enough to do log only.
			ctx.Log.Err("unknown state subcommand: %s", subName)
		}
	}

	// If TerraformVersion not defined in config file look for a
	// terraform.require_version block.
	if prjCfg.TerraformVersion == nil {
		prjCfg.TerraformVersion = terraformClient.DetectVersion(ctx.Log, filepath.Join(repoDir, prjCfg.RepoRelDir))
	}

	projectCmdContext := newProjectCommandContext(
		ctx,
		cmdName,
		cb.CommentBuilder.BuildApplyComment(prjCfg.RepoRelDir, prjCfg.Workspace, prjCfg.Name, prjCfg.AutoMergeDisabled, prjCfg.MergeMethod),
		cb.CommentBuilder.BuildApprovePoliciesComment(prjCfg.RepoRelDir, prjCfg.Workspace, prjCfg.Name),
		cb.CommentBuilder.BuildPlanComment(prjCfg.RepoRelDir, prjCfg.Workspace, prjCfg.Name, commentFlags),
		prjCfg,
		steps,
		prjCfg.PolicySets,
		escapeArgs(commentFlags),
		automerge,
		parallelApply,
		parallelPlan,
		verbose,
		abortOnExcecutionOrderFail,
		ctx.Scope,
		ctx.PullRequestStatus,
		ctx.PullStatus,
		ctx.TeamAllowlistChecker,
	)

	projectCmds = append(projectCmds, projectCmdContext)

	return
}

type PolicyCheckProjectCommandContextBuilder struct {
	ProjectCommandContextBuilder *DefaultProjectCommandContextBuilder
	CommentBuilder               CommentBuilder
}

func (cb *PolicyCheckProjectCommandContextBuilder) BuildProjectContext(
	ctx *command.Context,
	cmdName command.Name,
	subCmdName string,
	prjCfg valid.MergedProjectCfg,
	commentFlags []string,
	repoDir string,
	automerge, parallelApply, parallelPlan, verbose, abortOnExcecutionOrderFail bool,
	terraformClient terraform.Client,
) (projectCmds []command.ProjectContext) {
	if prjCfg.PolicyCheck {
		ctx.Log.Debug("PolicyChecks are enabled")
	} else {
		// PolicyCheck is disabled at repository level
		ctx.Log.Debug("PolicyChecks are disabled on this repository")
	}

	// If TerraformVersion not defined in config file look for a
	// terraform.require_version block.
	if prjCfg.TerraformVersion == nil {
		prjCfg.TerraformVersion = terraformClient.DetectVersion(ctx.Log, filepath.Join(repoDir, prjCfg.RepoRelDir))
	}

	projectCmds = cb.ProjectCommandContextBuilder.BuildProjectContext(
		ctx,
		cmdName,
		subCmdName,
		prjCfg,
		commentFlags,
		repoDir,
		automerge,
		parallelApply,
		parallelPlan,
		verbose,
		abortOnExcecutionOrderFail,
		terraformClient,
	)

	if cmdName == command.Plan && prjCfg.PolicyCheck {
		ctx.Log.Debug("Building project command context for %s", command.PolicyCheck)
		steps := prjCfg.Workflow.PolicyCheck.Steps

		projectCmds = append(projectCmds, newProjectCommandContext(
			ctx,
			command.PolicyCheck,
			cb.CommentBuilder.BuildApplyComment(prjCfg.RepoRelDir, prjCfg.Workspace, prjCfg.Name, prjCfg.AutoMergeDisabled, prjCfg.MergeMethod),
			cb.CommentBuilder.BuildApprovePoliciesComment(prjCfg.RepoRelDir, prjCfg.Workspace, prjCfg.Name),
			cb.CommentBuilder.BuildPlanComment(prjCfg.RepoRelDir, prjCfg.Workspace, prjCfg.Name, commentFlags),
			prjCfg,
			steps,
			prjCfg.PolicySets,
			escapeArgs(commentFlags),
			automerge,
			parallelApply,
			parallelPlan,
			verbose,
			abortOnExcecutionOrderFail,
			ctx.Scope,
			ctx.PullRequestStatus,
			ctx.PullStatus,
			ctx.TeamAllowlistChecker,
		))
	}

	return
}

// newProjectCommandContext is a initializer method that handles constructing the
// ProjectCommandContext.
func newProjectCommandContext(ctx *command.Context,
	cmd command.Name,
	applyCmd string,
	approvePoliciesCmd string,
	planCmd string,
	projCfg valid.MergedProjectCfg,
	steps []valid.Step,
	policySets valid.PolicySets,
	escapedCommentArgs []string,
	automergeEnabled bool,
	parallelApplyEnabled bool,
	parallelPlanEnabled bool,
	verbose bool,
	abortOnExcecutionOrderFail bool,
	scope tally.Scope,
	pullReqStatus models.PullReqStatus,
	pullStatus *models.PullStatus,
	teamAllowlistChecker command.TeamAllowlistChecker,
) command.ProjectContext {

	var projectPlanStatus models.ProjectPlanStatus
	var projectPolicyStatus []models.PolicySetStatus

	if ctx.PullStatus != nil {
		for _, project := range ctx.PullStatus.Projects {

			// if name is not used, let's match the directory
			if projCfg.Name == "" && project.RepoRelDir == projCfg.RepoRelDir {
				projectPlanStatus = project.Status
				projectPolicyStatus = project.PolicyStatus
				break
			}

			if projCfg.Name != "" && project.ProjectName == projCfg.Name {
				projectPlanStatus = project.Status
				projectPolicyStatus = project.PolicyStatus
				break
			}
		}
	}

	return command.ProjectContext{
		CommandName:                cmd,
		ApplyCmd:                   applyCmd,
		ApprovePoliciesCmd:         approvePoliciesCmd,
		BaseRepo:                   ctx.Pull.BaseRepo,
		EscapedCommentArgs:         escapedCommentArgs,
		AutomergeEnabled:           automergeEnabled,
		DeleteSourceBranchOnMerge:  projCfg.DeleteSourceBranchOnMerge,
		RepoLocksMode:              projCfg.RepoLocks.Mode,
		CustomPolicyCheck:          projCfg.CustomPolicyCheck,
		ParallelApplyEnabled:       parallelApplyEnabled,
		ParallelPlanEnabled:        parallelPlanEnabled,
		ParallelPolicyCheckEnabled: parallelPlanEnabled,
		DependsOn:                  projCfg.DependsOn,
		AutoplanEnabled:            projCfg.AutoplanEnabled,
		Steps:                      steps,
		HeadRepo:                   ctx.HeadRepo,
		Log:                        ctx.Log,
		Scope:                      scope,
		ProjectPlanStatus:          projectPlanStatus,
		ProjectPolicyStatus:        projectPolicyStatus,
		Pull:                       ctx.Pull,
		ProjectName:                projCfg.Name,
		PlanRequirements:           projCfg.PlanRequirements,
		ApplyRequirements:          projCfg.ApplyRequirements,
		ImportRequirements:         projCfg.ImportRequirements,
		RePlanCmd:                  planCmd,
		RepoRelDir:                 projCfg.RepoRelDir,
		RepoConfigVersion:          projCfg.RepoCfgVersion,
		TerraformVersion:           projCfg.TerraformVersion,
		User:                       ctx.User,
		Verbose:                    verbose,
		Workspace:                  projCfg.Workspace,
		PolicySets:                 policySets,
		PolicySetTarget:            ctx.PolicySet,
		ClearPolicyApproval:        ctx.ClearPolicyApproval,
		PullReqStatus:              pullReqStatus,
		PullStatus:                 pullStatus,
		JobID:                      uuid.New().String(),
		ExecutionOrderGroup:        projCfg.ExecutionOrderGroup,
		AbortOnExcecutionOrderFail: abortOnExcecutionOrderFail,
		SilencePRComments:          projCfg.SilencePRComments,
		TeamAllowlistChecker:       teamAllowlistChecker,
	}
}

func escapeArgs(args []string) []string {
	var escaped []string
	for _, arg := range args {
		var escapedArg string
		for i := range arg {
			escapedArg += "\\" + string(arg[i])
		}
		escaped = append(escaped, escapedArg)
	}
	return escaped
}
