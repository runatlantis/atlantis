package models

import (
	"path/filepath"
	"regexp"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
)

// buildCtx is a helper method that handles constructing the ProjectCommandContext.
func NewProjectCommandContext(ctx *CommandContext,
	cmd CommandName,
	applyCmd string,
	planCmd string,
	projCfg valid.MergedProjectCfg,
	commentArgs []string,
	automergeEnabled bool,
	parallelApplyEnabled bool,
	parallelPlanEnabled bool,
	verbose bool,
	absRepoDir string) ProjectCommandContext {

	var policySets PolicySets
	var steps []valid.Step
	switch cmd {
	case PlanCommand:
		steps = projCfg.Workflow.Plan.Steps
	case ApplyCommand:
		steps = projCfg.Workflow.Apply.Steps
	case PolicyCheckCommand:
		steps = projCfg.Workflow.PolicyCheck.Steps
	}

	// If TerraformVersion not defined in config file look for a
	// terraform.require_version block.
	if projCfg.TerraformVersion == nil {
		projCfg.TerraformVersion = getTfVersion(ctx, filepath.Join(absRepoDir, projCfg.RepoRelDir))
	}

	return ProjectCommandContext{
		CommandName:          cmd,
		ApplyCmd:             applyCmd,
		BaseRepo:             ctx.Pull.BaseRepo,
		EscapedCommentArgs:   escapeArgs(commentArgs),
		AutomergeEnabled:     automergeEnabled,
		ParallelApplyEnabled: parallelApplyEnabled,
		ParallelPlanEnabled:  parallelPlanEnabled,
		AutoplanEnabled:      projCfg.AutoplanEnabled,
		Steps:                steps,
		HeadRepo:             ctx.HeadRepo,
		Log:                  ctx.Log,
		PullMergeable:        ctx.PullMergeable,
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
