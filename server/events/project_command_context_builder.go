package events

import (
	"path/filepath"
	"regexp"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	stats "github.com/lyft/gostats"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
)

func NewProjectCommandContextBulder(policyCheckEnabled bool, commentBuilder CommentBuilder, scope stats.Scope) ProjectCommandContextBuilder {
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
		ProjectCounter:               scope.NewCounter("projects"),
	}
}

type ProjectCommandContextBuilder interface {
	// BuildProjectContext builds project command contexts for atlantis commands
	BuildProjectContext(
		ctx *CommandContext,
		cmdName models.CommandName,
		prjCfg valid.MergedProjectCfg,
		commentFlags []string,
		repoDir string,
		automerge, parallelPlan, parallelApply, verbose bool,
	) []models.ProjectCommandContext
}

// CommandScopedStatsProjectCommandContextBuilder ensures that project command context contains a scoped stats
// object relevant to the command it applies to.
type CommandScopedStatsProjectCommandContextBuilder struct {
	ProjectCommandContextBuilder
	// Conciously making this global since it gets flushed periodically anyways
	ProjectCounter stats.Counter
}

// BuildProjectContext builds the context and injects the appropriate command level scope after the fact.
func (cb *CommandScopedStatsProjectCommandContextBuilder) BuildProjectContext(
	ctx *CommandContext,
	cmdName models.CommandName,
	prjCfg valid.MergedProjectCfg,
	commentFlags []string,
	repoDir string,
	automerge, parallelApply, parallelPlan, verbose bool,
) (projectCmds []models.ProjectCommandContext) {
	cb.ProjectCounter.Inc()

	cmds := cb.ProjectCommandContextBuilder.BuildProjectContext(
		ctx, cmdName, prjCfg, commentFlags, repoDir, automerge, parallelApply, parallelPlan, verbose,
	)

	projectCmds = []models.ProjectCommandContext{}

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
	ctx *CommandContext,
	cmdName models.CommandName,
	prjCfg valid.MergedProjectCfg,
	commentFlags []string,
	repoDir string,
	automerge, parallelApply, parallelPlan, verbose bool,
) (projectCmds []models.ProjectCommandContext) {
	ctx.Log.Debug("Building project command context for %s", cmdName)

	var steps []valid.Step
	switch cmdName {
	case models.PlanCommand:
		steps = prjCfg.Workflow.Plan.Steps
	case models.ApplyCommand:
		steps = prjCfg.Workflow.Apply.Steps
	}

	// If TerraformVersion not defined in config file look for a
	// terraform.require_version block.
	if prjCfg.TerraformVersion == nil {
		prjCfg.TerraformVersion = getTfVersion(ctx, filepath.Join(repoDir, prjCfg.RepoRelDir))
	}

	projectCmds = append(projectCmds, newProjectCommandContext(
		ctx,
		cmdName,
		cb.CommentBuilder.BuildApplyComment(prjCfg.RepoRelDir, prjCfg.Workspace, prjCfg.Name),
		cb.CommentBuilder.BuildPlanComment(prjCfg.RepoRelDir, prjCfg.Workspace, prjCfg.Name, commentFlags),
		prjCfg,
		steps,
		prjCfg.PolicySets,
		escapeArgs(commentFlags),
		automerge,
		parallelApply,
		parallelPlan,
		verbose,
		ctx.Scope,
	))

	return
}

type PolicyCheckProjectCommandContextBuilder struct {
	ProjectCommandContextBuilder *DefaultProjectCommandContextBuilder
	CommentBuilder               CommentBuilder
}

func (cb *PolicyCheckProjectCommandContextBuilder) BuildProjectContext(
	ctx *CommandContext,
	cmdName models.CommandName,
	prjCfg valid.MergedProjectCfg,
	commentFlags []string,
	repoDir string,
	automerge, parallelApply, parallelPlan, verbose bool,
) (projectCmds []models.ProjectCommandContext) {
	ctx.Log.Debug("PolicyChecks are enabled")
	projectCmds = cb.ProjectCommandContextBuilder.BuildProjectContext(
		ctx,
		cmdName,
		prjCfg,
		commentFlags,
		repoDir,
		automerge,
		parallelApply,
		parallelPlan,
		verbose,
	)

	if cmdName == models.PlanCommand {
		ctx.Log.Debug("Building project command context for %s", models.PolicyCheckCommand)
		steps := prjCfg.Workflow.PolicyCheck.Steps

		projectCmds = append(projectCmds, newProjectCommandContext(
			ctx,
			models.PolicyCheckCommand,
			cb.CommentBuilder.BuildApplyComment(prjCfg.RepoRelDir, prjCfg.Workspace, prjCfg.Name),
			cb.CommentBuilder.BuildPlanComment(prjCfg.RepoRelDir, prjCfg.Workspace, prjCfg.Name, commentFlags),
			prjCfg,
			steps,
			prjCfg.PolicySets,
			escapeArgs(commentFlags),
			automerge,
			parallelApply,
			parallelPlan,
			verbose,
			ctx.Scope,
		))
	}

	return
}

// newProjectCommandContext is a initializer method that handles constructing the
// ProjectCommandContext.
func newProjectCommandContext(ctx *CommandContext,
	cmd models.CommandName,
	applyCmd string,
	planCmd string,
	projCfg valid.MergedProjectCfg,
	steps []valid.Step,
	policySets valid.PolicySets,
	escapedCommentArgs []string,
	automergeEnabled bool,
	parallelApplyEnabled bool,
	parallelPlanEnabled bool,
	verbose bool,
	scope stats.Scope,
) models.ProjectCommandContext {

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

	return models.ProjectCommandContext{
		CommandName:          cmd,
		ApplyCmd:             applyCmd,
		BaseRepo:             ctx.Pull.BaseRepo,
		EscapedCommentArgs:   escapedCommentArgs,
		AutomergeEnabled:     automergeEnabled,
		ParallelApplyEnabled: parallelApplyEnabled,
		ParallelPlanEnabled:  parallelPlanEnabled,
		AutoplanEnabled:      projCfg.AutoplanEnabled,
		Steps:                steps,
		HeadRepo:             ctx.HeadRepo,
		Log:                  ctx.Log,
		Scope:                scope,
		PullMergeable:        ctx.PullMergeable,
		ProjectPlanStatus:    projectPlanStatus,
		Pull:                 ctx.Pull,
		ProjectName:          projCfg.Name,
		ApplyRequirements:    projCfg.ApplyRequirements,
		RePlanCmd:            planCmd,
		RepoRelDir:           projCfg.RepoRelDir,
		RepoConfigVersion:    projCfg.RepoCfgVersion,
		TerraformVersion:     projCfg.TerraformVersion,
		User:                 ctx.User,
		Verbose:              verbose,
		Workspace:            projCfg.Workspace,
		PolicySets:           policySets,
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
func getTfVersion(ctx *CommandContext, absProjDir string) *version.Version {
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
