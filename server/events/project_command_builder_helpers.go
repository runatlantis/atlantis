package events

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/yaml"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
)

// ProjectCommandBuilder helper functions

// buildProjectCommands builds project command for a provided command name based
// on existing plan files(apply, policy_checks). This helper only works after
// atlantis plan already ran.
func buildProjectCommands(
	ctx *CommandContext,
	cmdName models.CommandName,
	commentCmd *CommentCommand,
	commentBuilder CommentBuilder,
	globalCfg valid.GlobalCfg,
	workingDirLocker WorkingDirLocker,
	workingDir WorkingDir,
) ([]models.ProjectCommandContext, error) {
	planFinder := &DefaultPendingPlanFinder{}
	// Lock all dirs in this pull request (instead of a single dir) because we
	// don't know how many dirs we'll need to apply in.
	unlockFn, err := workingDirLocker.TryLockPull(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num)
	if err != nil {
		return nil, err
	}
	defer unlockFn()

	pullDir, err := workingDir.GetPullDir(ctx.Pull.BaseRepo, ctx.Pull)
	if err != nil {
		return nil, err
	}

	plans, err := planFinder.Find(pullDir)
	if err != nil {
		return nil, err
	}

	var cmds []models.ProjectCommandContext
	for _, plan := range plans {
		cmd, err := buildProjectCommandCtx(
			ctx,
			cmdName,
			globalCfg,
			commentBuilder,
			plan.ProjectName,
			commentCmd.Flags,
			plan.RepoDir,
			plan.RepoRelDir,
			plan.Workspace,
			commentCmd.Verbose,
		)
		if err != nil {
			return nil, errors.Wrapf(err, "building command for dir %q", plan.RepoRelDir)
		}
		cmds = append(cmds, cmd)
	}
	return cmds, nil
}

func buildProjectCommandCtx(
	ctx *CommandContext,
	cmd models.CommandName,
	globalCfg valid.GlobalCfg,
	commentBuilder CommentBuilder,
	projectName string,
	commentFlags []string,
	repoDir string,
	repoRelDir string,
	workspace string,
	verbose bool) (models.ProjectCommandContext, error) {

	projCfgPtr, repoCfgPtr, err := getCfg(ctx, globalCfg, projectName, repoRelDir, workspace, repoDir)
	if err != nil {
		return models.ProjectCommandContext{}, err
	}

	var projCfg valid.MergedProjectCfg
	if projCfgPtr != nil {
		// Override any dir/workspace defined on the comment with what was
		// defined in config. This shouldn't matter since we don't allow comments
		// with both project name and dir/workspace.
		repoRelDir = projCfg.RepoRelDir
		workspace = projCfg.Workspace
		projCfg = globalCfg.MergeProjectCfg(ctx.Log, ctx.Pull.BaseRepo.ID(), *projCfgPtr, *repoCfgPtr)
	} else {
		projCfg = globalCfg.DefaultProjCfg(ctx.Log, ctx.Pull.BaseRepo.ID(), repoRelDir, workspace)
	}

	if err := validateWorkspaceAllowed(repoCfgPtr, repoRelDir, workspace); err != nil {
		return models.ProjectCommandContext{}, err
	}

	automerge := DefaultAutomergeEnabled
	parallelApply := DefaultParallelApplyEnabled
	parallelPlan := DefaultParallelPlanEnabled
	if repoCfgPtr != nil {
		automerge = repoCfgPtr.Automerge
		parallelApply = repoCfgPtr.ParallelApply
		parallelPlan = repoCfgPtr.ParallelPlan
	}

	return buildCtx(
		ctx,
		cmd,
		commentBuilder,
		projCfg,
		commentFlags,
		automerge,
		parallelApply,
		parallelPlan,
		verbose,
		repoDir,
	), nil
}

// validateWorkspaceAllowed returns an error if repoCfg defines projects in
// repoRelDir but none of them use workspace. We want this to be an error
// because if users have gone to the trouble of defining projects in repoRelDir
// then it's likely that if we're running a command for a workspace that isn't
// defined then they probably just typed the workspace name wrong.
func validateWorkspaceAllowed(repoCfg *valid.RepoCfg, repoRelDir string, workspace string) error {
	if repoCfg == nil {
		return nil
	}

	projects := repoCfg.FindProjectsByDir(repoRelDir)

	// If that directory doesn't have any projects configured then we don't
	// enforce workspace names.
	if len(projects) == 0 {
		return nil
	}

	var configuredSpaces []string
	for _, p := range projects {
		if p.Workspace == workspace {
			return nil
		}
		configuredSpaces = append(configuredSpaces, p.Workspace)
	}

	return fmt.Errorf(
		"running commands in workspace %q is not allowed because this"+
			" directory is only configured for the following workspaces: %s",
		workspace,
		strings.Join(configuredSpaces, ", "),
	)
}

// buildCtx is a helper method that handles constructing the ProjectCommandContext.
func buildCtx(ctx *CommandContext,
	cmd models.CommandName,
	commentBuilder CommentBuilder,
	projCfg valid.MergedProjectCfg,
	commentArgs []string,
	automergeEnabled bool,
	parallelApplyEnabled bool,
	parallelPlanEnabled bool,
	verbose bool,
	absRepoDir string) models.ProjectCommandContext {

	var policySets models.PolicySets
	var steps []valid.Step
	switch cmd {
	case models.PlanCommand:
		steps = projCfg.Workflow.Plan.Steps
	case models.ApplyCommand:
		steps = projCfg.Workflow.Apply.Steps
	case models.PolicyCheckCommand:
		steps = projCfg.Workflow.PolicyCheck.Steps
	}

	// If TerraformVersion not defined in config file look for a
	// terraform.require_version block.
	if projCfg.TerraformVersion == nil {
		projCfg.TerraformVersion = getTfVersion(ctx, filepath.Join(absRepoDir, projCfg.RepoRelDir))
	}

	return models.ProjectCommandContext{
		CommandName:          cmd,
		ApplyCmd:             commentBuilder.BuildApplyComment(projCfg.RepoRelDir, projCfg.Workspace, projCfg.Name),
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
		RePlanCmd:            commentBuilder.BuildPlanComment(projCfg.RepoRelDir, projCfg.Workspace, projCfg.Name, commentArgs),
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

// getCfg returns the atlantis.yaml config (if it exists) for this project. If
// there is no config, then projectCfg and repoCfg will be nil.
func getCfg(
	ctx *CommandContext,
	globalCfg valid.GlobalCfg,
	projectName string,
	dir string,
	workspace string,
	repoDir string,
) (projectCfg *valid.Project, repoCfg *valid.RepoCfg, err error) {
	parserValidator := &yaml.ParserValidator{}

	hasConfigFile, err := parserValidator.HasRepoCfg(repoDir)
	if err != nil {
		err = errors.Wrapf(err, "looking for %s file in %q", yaml.AtlantisYAMLFilename, repoDir)
		return
	}
	if !hasConfigFile {
		if projectName != "" {
			err = fmt.Errorf("cannot specify a project name unless an %s file exists to configure projects", yaml.AtlantisYAMLFilename)
			return
		}
		return
	}

	var repoConfig valid.RepoCfg
	repoConfig, err = parserValidator.ParseRepoCfg(repoDir, globalCfg, ctx.Pull.BaseRepo.ID())
	if err != nil {
		return
	}
	repoCfg = &repoConfig

	// If they've specified a project by name we look it up. Otherwise we
	// use the dir and workspace.
	if projectName != "" {
		projectCfg = repoCfg.FindProjectByName(projectName)
		if projectCfg == nil {
			err = fmt.Errorf("no project with name %q is defined in %s", projectName, yaml.AtlantisYAMLFilename)
			return
		}
		return
	}

	projCfgs := repoCfg.FindProjectsByDirWorkspace(dir, workspace)
	if len(projCfgs) == 0 {
		return
	}
	if len(projCfgs) > 1 {
		err = fmt.Errorf("must specify project name: more than one project defined in %s matched dir: %q workspace: %q", yaml.AtlantisYAMLFilename, dir, workspace)
		return
	}
	projectCfg = &projCfgs[0]
	return
}
