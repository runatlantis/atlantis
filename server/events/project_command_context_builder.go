package events

import (
	"path/filepath"
	"regexp"

	"github.com/google/uuid"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/uber-go/tally"
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
		prjCfg valid.MergedProjectCfg,
		commentFlags []string,
		repoDir string,
		automerge, deleteSourceBranchOnMerge, parallelApply, parallelPlan, verbose bool,
	) []command.ProjectContext
}

// CommandScopedStatsProjectCommandContextBuilder ensures that project command context contains a scoped stats
// object relevant to the command it applies to.
type CommandScopedStatsProjectCommandContextBuilder struct {
	ProjectCommandContextBuilder
	// Conciously making this global since it gets flushed periodically anyways
	ProjectCounter tally.Counter
}

// BuildProjectContext builds the context and injects the appropriate command level scope after the fact.
func (cb *CommandScopedStatsProjectCommandContextBuilder) BuildProjectContext(
	ctx *command.Context,
	cmdName command.Name,
	prjCfg valid.MergedProjectCfg,
	commentFlags []string,
	repoDir string,
	automerge, deleteSourceBranchOnMerge, parallelApply, parallelPlan, verbose bool,
) (projectCmds []command.ProjectContext) {
	cb.ProjectCounter.Inc(1)

	cmds := cb.ProjectCommandContextBuilder.BuildProjectContext(
		ctx, cmdName, prjCfg, commentFlags, repoDir, automerge, deleteSourceBranchOnMerge, parallelApply, parallelPlan, verbose,
	)

	projectCmds = []command.ProjectContext{}

	for _, cmd := range cmds {

		// specifically use the command name in the context instead of the arg
		// since we can return multiple commands worth of contexts for a given command name arg
		// to effectively pipeline them.
		cmd.SetScope(cmd.CommandName.String())
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
	prjCfg valid.MergedProjectCfg,
	commentFlags []string,
	repoDir string,
	automerge, deleteSourceBranchOnMerge, parallelApply, parallelPlan, verbose bool,
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
	}

	// If TerraformVersion not defined in config file look for a
	// terraform.require_version block.
	if prjCfg.TerraformVersion == nil {
		prjCfg.TerraformVersion = getTfVersion(ctx, filepath.Join(repoDir, prjCfg.RepoRelDir))
	}

	projectCmdContext := newProjectCommandContext(
		ctx,
		cmdName,
		cb.CommentBuilder.BuildApplyComment(prjCfg.RepoRelDir, prjCfg.Workspace, prjCfg.Name, prjCfg.AutoMergeDisabled),
		cb.CommentBuilder.BuildPlanComment(prjCfg.RepoRelDir, prjCfg.Workspace, prjCfg.Name, commentFlags),
		cb.CommentBuilder.BuildVersionComment(prjCfg.RepoRelDir, prjCfg.Workspace, prjCfg.Name),
		prjCfg,
		steps,
		prjCfg.PolicySets,
		escapeArgs(commentFlags),
		automerge,
		deleteSourceBranchOnMerge,
		parallelApply,
		parallelPlan,
		verbose,
		ctx.Scope,
		ctx.PullRequestStatus,
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
	prjCfg valid.MergedProjectCfg,
	commentFlags []string,
	repoDir string,
	automerge, deleteSourceBranchOnMerge, parallelApply, parallelPlan, verbose bool,
) (projectCmds []command.ProjectContext) {
	ctx.Log.Debug("PolicyChecks are enabled")

	// If TerraformVersion not defined in config file look for a
	// terraform.require_version block.
	if prjCfg.TerraformVersion == nil {
		prjCfg.TerraformVersion = getTfVersion(ctx, filepath.Join(repoDir, prjCfg.RepoRelDir))
	}

	projectCmds = cb.ProjectCommandContextBuilder.BuildProjectContext(
		ctx,
		cmdName,
		prjCfg,
		commentFlags,
		repoDir,
		automerge,
		deleteSourceBranchOnMerge,
		parallelApply,
		parallelPlan,
		verbose,
	)

	if cmdName == command.Plan || cmdName == command.PlanAll {
		ctx.Log.Debug("Building project command context for %s", command.PolicyCheck)
		steps := prjCfg.Workflow.PolicyCheck.Steps

		projectCmds = append(projectCmds, newProjectCommandContext(
			ctx,
			command.PolicyCheck,
			cb.CommentBuilder.BuildApplyComment(prjCfg.RepoRelDir, prjCfg.Workspace, prjCfg.Name, prjCfg.AutoMergeDisabled),
			cb.CommentBuilder.BuildPlanComment(prjCfg.RepoRelDir, prjCfg.Workspace, prjCfg.Name, commentFlags),
			cb.CommentBuilder.BuildVersionComment(prjCfg.RepoRelDir, prjCfg.Workspace, prjCfg.Name),
			prjCfg,
			steps,
			prjCfg.PolicySets,
			escapeArgs(commentFlags),
			automerge,
			deleteSourceBranchOnMerge,
			parallelApply,
			parallelPlan,
			verbose,
			ctx.Scope,
			ctx.PullRequestStatus,
		))
	}

	return
}

// newProjectCommandContext is a initializer method that handles constructing the
// ProjectCommandContext.
func newProjectCommandContext(ctx *command.Context,
	cmd command.Name,
	applyCmd string,
	planCmd string,
	versionCmd string,
	projCfg valid.MergedProjectCfg,
	steps []valid.Step,
	policySets valid.PolicySets,
	escapedCommentArgs []string,
	automergeEnabled bool,
	deleteSourceBranchOnMerge,
	parallelApplyEnabled bool,
	parallelPlanEnabled bool,
	verbose bool,
	scope tally.Scope,
	pullStatus models.PullReqStatus,
) command.ProjectContext {

	var projectPlanStatus models.ProjectPlanStatus

	if ctx.PullStatus != nil {
		for _, project := range ctx.PullStatus.Projects {

			// if name is not used, let's match the directory
			if projCfg.Name == "" && project.RepoRelDir == projCfg.RepoRelDir {
				projectPlanStatus = project.Status
				break
			}

			if projCfg.Name != "" && project.ProjectName == projCfg.Name {
				projectPlanStatus = project.Status
				break
			}
		}
	}

	return command.ProjectContext{
		CommandName:                cmd,
		ApplyCmd:                   applyCmd,
		BaseRepo:                   ctx.Pull.BaseRepo,
		EscapedCommentArgs:         escapedCommentArgs,
		AutomergeEnabled:           automergeEnabled,
		DeleteSourceBranchOnMerge:  deleteSourceBranchOnMerge,
		ParallelApplyEnabled:       parallelApplyEnabled,
		ParallelPlanEnabled:        parallelPlanEnabled,
		ParallelPolicyCheckEnabled: parallelPlanEnabled,
		AutoplanEnabled:            projCfg.AutoplanEnabled,
		Steps:                      steps,
		HeadRepo:                   ctx.HeadRepo,
		Log:                        ctx.Log,
		Scope:                      scope,
		ProjectPlanStatus:          projectPlanStatus,
		Pull:                       ctx.Pull,
		ProjectName:                projCfg.Name,
		ApplyRequirements:          projCfg.ApplyRequirements,
		RePlanCmd:                  planCmd,
		RepoRelDir:                 projCfg.RepoRelDir,
		RepoConfigVersion:          projCfg.RepoCfgVersion,
		TerraformVersion:           projCfg.TerraformVersion,
		User:                       ctx.User,
		Verbose:                    verbose,
		Workspace:                  projCfg.Workspace,
		PolicySets:                 policySets,
		PullReqStatus:              pullStatus,
		JobID:                      uuid.New().String(),
		ExecutionOrderGroup:        projCfg.ExecutionOrderGroup,
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

// Extracts required_version from Terraform configuration.
// Returns nil if unable to determine version from configuration.
func getTfVersion(ctx *command.Context, absProjDir string) *version.Version {
	module, diags := tfconfig.LoadModule(absProjDir)
	if diags.HasErrors() {
		ctx.Log.Err("trying to detect required version: %s", diags.Error())
		return nil
	}

	if len(module.RequiredCore) != 1 {
		ctx.Log.Info("cannot determine which version to use from terraform configuration, detected %d possibilities.", len(module.RequiredCore))
		return nil
	}
	requiredVersionSetting := module.RequiredCore[0]

	// We allow `= x.y.z`, `=x.y.z` or `x.y.z` where `x`, `y` and `z` are integers.
	re := regexp.MustCompile(`^=?\s*([^\s]+)\s*$`)
	matched := re.FindStringSubmatch(requiredVersionSetting)
	if len(matched) == 0 {
		ctx.Log.Debug("did not specify exact version in terraform configuration, found %q", requiredVersionSetting)
		return nil
	}
	ctx.Log.Debug("found required_version setting of %q", requiredVersionSetting)
	version, err := version.NewVersion(matched[1])
	if err != nil {
		ctx.Log.Debug(err.Error())
		return nil
	}

	ctx.Log.Info("detected module requires version: %q", version.String())
	return version
}
