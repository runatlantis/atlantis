package events

import (
	"path/filepath"
	"regexp"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/command"
)

type ProjectCommandContextBuilder interface {
	// BuildProjectContext builds project command contexts for atlantis commands
	BuildProjectContext(
		ctx *command.Context,
		cmdName command.Name,
		prjCfg valid.MergedProjectCfg,
		commentArgs []string,
		repoDir string,
		contextFlags *command.ContextFlags,
	) []command.ProjectContext
}

func NewProjectCommandContextBuilder(
	commentBuilder CommentBuilder,
) ProjectCommandContextBuilder {
	return &projectCommandContextBuilder{
		CommentBuilder: commentBuilder,
	}
}

type projectCommandContextBuilder struct {
	CommentBuilder CommentBuilder
}

func (cb *projectCommandContextBuilder) BuildProjectContext(
	ctx *command.Context,
	cmdName command.Name,
	prjCfg valid.MergedProjectCfg,
	commentArgs []string,
	repoDir string,
	contextFlags *command.ContextFlags,
) []command.ProjectContext {
	return buildContext(
		ctx,
		cmdName,
		getSteps(cmdName, prjCfg.Workflow),
		cb.CommentBuilder,
		prjCfg,
		commentArgs,
		repoDir,
		contextFlags,
	)
}

func buildContext(
	ctx *command.Context,
	cmdName command.Name,
	steps []valid.Step,
	commentBuilder CommentBuilder,
	prjCfg valid.MergedProjectCfg,
	commentArgs []string,
	repoDir string,
	contextFlags *command.ContextFlags,
) []command.ProjectContext {
	ctx.Log.Debugf("Building project command context for %s", cmdName)
	projectCmds := make([]command.ProjectContext, 0)

	// If TerraformVersion not defined in config file look for a
	// terraform.require_version block.
	if prjCfg.TerraformVersion == nil {
		prjCfg.TerraformVersion = getTfVersion(ctx, filepath.Join(repoDir, prjCfg.RepoRelDir))
	}

	projectCmds = append(projectCmds,
		command.NewProjectContext(
			ctx,
			cmdName,
			commentBuilder.BuildApplyComment(prjCfg.RepoRelDir, prjCfg.Workspace, prjCfg.Name),
			commentBuilder.BuildPlanComment(prjCfg.RepoRelDir, prjCfg.Workspace, prjCfg.Name, commentArgs),
			prjCfg,
			steps,
			prjCfg.PolicySets,
			escapeArgs(commentArgs),
			contextFlags,
			ctx.Scope,
			ctx.PullRequestStatus,
		),
	)

	return projectCmds

}

func getSteps(
	cmdName command.Name,
	workflow valid.Workflow,
) (steps []valid.Step) {
	switch cmdName {
	case command.Plan:
		steps = workflow.Plan.Steps
	case command.Apply:
		steps = workflow.Apply.Steps
	case command.Version:
		steps = []valid.Step{{
			StepName: "version",
		}}
	case command.PolicyCheck:
		steps = workflow.PolicyCheck.Steps
	}
	return steps
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
		ctx.Log.Errorf("trying to detect required version: %s", diags.Error())
		return nil
	}
	if len(module.RequiredCore) != 1 {
		ctx.Log.Infof("cannot determine which version to use from terraform configuration, detected %d possibilities.", len(module.RequiredCore))
		return nil
	}
	requiredVersionSetting := module.RequiredCore[0]

	// We allow `= x.y.z`, `=x.y.z` or `x.y.z` where `x`, `y` and `z` are integers.
	re := regexp.MustCompile(`^=?\s*([^\s]+)\s*$`)
	matched := re.FindStringSubmatch(requiredVersionSetting)
	if len(matched) == 0 {
		ctx.Log.Debugf("did not specify exact version in terraform configuration, found %q", requiredVersionSetting)
		return nil
	}
	ctx.Log.Debugf("found required_version setting of %q", requiredVersionSetting)
	version, err := version.NewVersion(matched[1])
	if err != nil {
		ctx.Log.Debugf(err.Error())
		return nil
	}

	ctx.Log.Infof("detected module requires version: %q", version.String())
	return version
}
