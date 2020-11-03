package events

import (
	"fmt"

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
	ctx *models.CommandContext,
	cmdName models.CommandName,
	commentCmd *CommentCommand,
	commentBuilder CommentBuilder,
	parser *yaml.ParserValidator,
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
			parser,
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
	ctx *models.CommandContext,
	cmd models.CommandName,
	parser *yaml.ParserValidator,
	globalCfg valid.GlobalCfg,
	commentBuilder CommentBuilder,
	projectName string,
	commentFlags []string,
	repoDir string,
	repoRelDir string,
	workspace string,
	verbose bool) (models.ProjectCommandContext, error) {

	projCfgPtr, repoCfgPtr, err := getCfg(ctx, parser, globalCfg, projectName, repoRelDir, workspace, repoDir)
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

	applyCmd, planCmd := buildRePlanAndApplyComments(
		commentBuilder,
		projCfg.RepoRelDir,
		projCfg.Workspace,
		projCfg.Name,
		commentFlags...)

	return models.NewProjectCommandContext(
		ctx,
		cmd,
		applyCmd,
		planCmd,
		projCfg,
		commentFlags,
		automerge,
		parallelApply,
		parallelPlan,
		verbose,
		repoDir,
	), nil
}

func buildRePlanAndApplyComments(commentBuilder CommentBuilder, repoRelDir string, workspace string, project string, commentArgs ...string) (applyCmd string, planCmd string) {
	applyCmd = commentBuilder.BuildApplyComment(repoRelDir, workspace, project)
	planCmd = commentBuilder.BuildPlanComment(repoRelDir, workspace, project, commentArgs)
	return
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

	return repoCfg.ValidateWorkspaceAllowed(repoRelDir, workspace)
}

// getCfg returns the atlantis.yaml config (if it exists) for this project. If
// there is no config, then projectCfg and repoCfg will be nil.
func getCfg(
	ctx *models.CommandContext,
	parserValidator *yaml.ParserValidator,
	globalCfg valid.GlobalCfg,
	projectName string,
	dir string,
	workspace string,
	repoDir string,
) (projectCfg *valid.Project, repoCfg *valid.RepoCfg, err error) {

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
