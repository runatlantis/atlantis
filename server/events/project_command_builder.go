package events

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/runatlantis/atlantis/server/events/yaml/valid"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/yaml"
)

const (
	// DefaultRepoRelDir is the default directory we run commands in, relative
	// to the root of the repo.
	DefaultRepoRelDir = "."
	// DefaultWorkspace is the default Terraform workspace we run commands in.
	// This is also Terraform's default workspace.
	DefaultWorkspace = "default"
	// DefaultAutomergeEnabled is the default for the automerge setting.
	DefaultAutomergeEnabled = false
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_project_command_builder.go ProjectCommandBuilder

// ProjectCommandBuilder builds commands that run on individual projects.
type ProjectCommandBuilder interface {
	// BuildAutoplanCommands builds project commands that will run plan on
	// the projects determined to be modified.
	BuildAutoplanCommands(ctx *CommandContext) ([]models.ProjectCommandContext, error)
	// BuildPlanCommands builds project plan commands for this ctx and comment. If
	// comment doesn't specify one project then there may be multiple commands
	// to be run.
	BuildPlanCommands(ctx *CommandContext, comment *CommentCommand) ([]models.ProjectCommandContext, error)
	// BuildApplyCommands builds project apply commands for ctx and comment. If
	// comment doesn't specify one project then there may be multiple commands
	// to be run.
	BuildApplyCommands(ctx *CommandContext, comment *CommentCommand) ([]models.ProjectCommandContext, error)
}

// DefaultProjectCommandBuilder implements ProjectCommandBuilder.
// This class combines the data from the comment and any atlantis.yaml file or
// Atlantis server config and then generates a set of contexts.
type DefaultProjectCommandBuilder struct {
	ParserValidator   *yaml.ParserValidator
	ProjectFinder     ProjectFinder
	VCSClient         vcs.Client
	WorkingDir        WorkingDir
	WorkingDirLocker  WorkingDirLocker
	GlobalCfg         valid.GlobalCfg
	PendingPlanFinder *DefaultPendingPlanFinder
	CommentBuilder    CommentBuilder
}

// See ProjectCommandBuilder.BuildAutoplanCommands.
func (p *DefaultProjectCommandBuilder) BuildAutoplanCommands(ctx *CommandContext) ([]models.ProjectCommandContext, error) {
	projCtxs, err := p.buildPlanAllCommands(ctx, nil, false)
	if err != nil {
		return nil, err
	}
	var autoplanEnabled []models.ProjectCommandContext
	for _, projCtx := range projCtxs {
		if !projCtx.AutoplanEnabled {
			ctx.Log.Debug("ignoring project at dir %q, workspace: %q because autoplan is disabled", projCtx.RepoRelDir, projCtx.Workspace)
			continue
		}
		autoplanEnabled = append(autoplanEnabled, projCtx)
	}
	return autoplanEnabled, nil
}

// See ProjectCommandBuilder.BuildPlanCommands.
func (p *DefaultProjectCommandBuilder) BuildPlanCommands(ctx *CommandContext, cmd *CommentCommand) ([]models.ProjectCommandContext, error) {
	if !cmd.IsForSpecificProject() {
		return p.buildPlanAllCommands(ctx, cmd.Flags, cmd.Verbose)
	}
	pcc, err := p.buildProjectPlanCommand(ctx, cmd)
	return []models.ProjectCommandContext{pcc}, err
}

// See ProjectCommandBuilder.BuildApplyCommands.
func (p *DefaultProjectCommandBuilder) BuildApplyCommands(ctx *CommandContext, cmd *CommentCommand) ([]models.ProjectCommandContext, error) {
	if !cmd.IsForSpecificProject() {
		return p.buildApplyAllCommands(ctx, cmd)
	}
	pac, err := p.buildProjectApplyCommand(ctx, cmd)
	return []models.ProjectCommandContext{pac}, err
}

// buildPlanAllCommands builds plan contexts for all projects we determine were
// modified in this ctx.
func (p *DefaultProjectCommandBuilder) buildPlanAllCommands(ctx *CommandContext, commentFlags []string, verbose bool) ([]models.ProjectCommandContext, error) {
	// Need to lock the workspace we're about to clone to.
	workspace := DefaultWorkspace
	unlockFn, err := p.WorkingDirLocker.TryLock(ctx.BaseRepo.FullName, ctx.Pull.Num, workspace)
	if err != nil {
		ctx.Log.Warn("workspace was locked")
		return nil, err
	}
	ctx.Log.Debug("got workspace lock")
	defer unlockFn()

	// We'll need the list of modified files.
	modifiedFiles, err := p.VCSClient.GetModifiedFiles(ctx.BaseRepo, ctx.Pull)
	if err != nil {
		return nil, err
	}
	ctx.Log.Debug("%d files were modified in this pull request", len(modifiedFiles))

	repoDir, err := p.prepareWorkspace(ctx, workspace)
	if err != nil {
		return nil, err
	}

	// Parse config file if it exists.
	hasRepoCfg, err := p.ParserValidator.HasRepoCfg(repoDir)
	if err != nil {
		return nil, errors.Wrapf(err, "looking for %s file in %q", yaml.AtlantisYAMLFilename, repoDir)
	}

	var projCtxs []models.ProjectCommandContext
	if hasRepoCfg {
		// If there's a repo cfg then we'll use it to figure out which projects
		// should be planed.
		repoCfg, err := p.ParserValidator.ParseRepoCfg(repoDir, p.GlobalCfg, ctx.BaseRepo.ID())
		if err != nil {
			return nil, errors.Wrapf(err, "parsing %s", yaml.AtlantisYAMLFilename)
		}
		ctx.Log.Info("successfully parsed %s file", yaml.AtlantisYAMLFilename)
		matchingProjects, err := p.ProjectFinder.DetermineProjectsViaConfig(ctx.Log, modifiedFiles, repoCfg, repoDir)
		if err != nil {
			return nil, err
		}
		ctx.Log.Info("%d projects are to be planned based on their when_modified config", len(matchingProjects))
		for _, mp := range matchingProjects {
			ctx.Log.Debug("determining config for project at dir: %q workspace: %q", mp.Dir, mp.Workspace)
			mergedCfg := p.GlobalCfg.MergeProjectCfg(ctx.Log, ctx.BaseRepo.ID(), mp, repoCfg)
			projCtxs = append(projCtxs, p.buildCtx(ctx, models.PlanCommand, mergedCfg, commentFlags, repoCfg.Automerge, verbose))
		}
	} else {
		// If there is no config file, then we'll plan each project that
		// our algorithm determines was modified.
		ctx.Log.Info("found no %s file", yaml.AtlantisYAMLFilename)
		modifiedProjects := p.ProjectFinder.DetermineProjects(ctx.Log, modifiedFiles, ctx.BaseRepo.FullName, repoDir)
		ctx.Log.Info("automatically determined that there were %d projects modified in this pull request: %s", len(modifiedProjects), modifiedProjects)
		for _, mp := range modifiedProjects {
			ctx.Log.Debug("determining config for project at dir: %q", mp.Path)
			pCfg := p.GlobalCfg.DefaultProjCfg(ctx.Log, ctx.BaseRepo.ID(), mp.Path, DefaultWorkspace)
			projCtxs = append(projCtxs, p.buildCtx(ctx, models.PlanCommand, pCfg, commentFlags, DefaultAutomergeEnabled, verbose))
		}
	}

	return projCtxs, nil
}

// buildProjectPlanCommand builds a plan context for a single project.
// cmd must be for only one project.
func (p *DefaultProjectCommandBuilder) buildProjectPlanCommand(ctx *CommandContext, cmd *CommentCommand) (models.ProjectCommandContext, error) {
	workspace := DefaultWorkspace
	if cmd.Workspace != "" {
		workspace = cmd.Workspace
	}

	var pcc models.ProjectCommandContext
	ctx.Log.Debug("building plan command")
	unlockFn, err := p.WorkingDirLocker.TryLock(ctx.BaseRepo.FullName, ctx.Pull.Num, workspace)
	if err != nil {
		return pcc, err
	}
	defer unlockFn()

	ctx.Log.Debug("cloning repository")
	repoDir, err := p.prepareWorkspace(ctx, workspace)
	if err != nil {
		return pcc, err
	}

	repoRelDir := DefaultRepoRelDir
	if cmd.RepoRelDir != "" {
		repoRelDir = cmd.RepoRelDir
	}

	return p.buildProjectCommandCtx(ctx, models.PlanCommand, cmd.ProjectName, cmd.Flags, repoDir, repoRelDir, workspace, cmd.Verbose)
}

// prepareWorkspace clones new changes into our repository and checks out the appropriate
// version of atlantis.yaml.
func (p *DefaultProjectCommandBuilder) prepareWorkspace(ctx *CommandContext, workspace string) (string, error) {
	// If we have no repo that matches in our global config then that is ok, as we'll get a
	// nil pointer back that will allow us to ignore the config source branch
	_, _, _, _, configSourceBranch := p.GlobalCfg.GetMatchingCfg(ctx.Log, ctx.BaseRepo.ID())

	// If we need to access another branch, ensure we fetch it during our initial clone
	additionalBranches := []string{}
	if configSourceBranch != nil {
		additionalBranches = append(additionalBranches, *configSourceBranch)
	}

	repoDir, err := p.WorkingDir.Clone(ctx.Log, ctx.BaseRepo, ctx.HeadRepo, ctx.Pull, workspace, additionalBranches)
	if err != nil {
		return repoDir, err
	}

	// If we've specified a source branch for our atlantis.yaml, checkout the file from that
	// branch before continuing with validation.
	if configSourceBranch != nil {
		ctx.Log.Debug("checking out %s from repos config source branch %s", yaml.AtlantisYAMLFilename, *configSourceBranch)
		checkoutCmd := exec.Command("git", "checkout", fmt.Sprintf("origin/%s", *configSourceBranch), "--", yaml.AtlantisYAMLFilename)
		checkoutCmd.Dir = repoDir
		output, err := checkoutCmd.CombinedOutput()
		if err != nil {
			return repoDir, errors.Wrapf(err, "failed to checkout %s from branch %s in %s: %s", yaml.AtlantisYAMLFilename, *configSourceBranch, repoDir, string(output))
		}
	}

	return repoDir, nil
}

// buildApplyAllCommands builds apply contexts for every project that has
// pending plans in this ctx.
func (p *DefaultProjectCommandBuilder) buildApplyAllCommands(ctx *CommandContext, commentCmd *CommentCommand) ([]models.ProjectCommandContext, error) {
	// Lock all dirs in this pull request (instead of a single dir) because we
	// don't know how many dirs we'll need to apply in.
	unlockFn, err := p.WorkingDirLocker.TryLockPull(ctx.BaseRepo.FullName, ctx.Pull.Num)
	if err != nil {
		return nil, err
	}
	defer unlockFn()

	pullDir, err := p.WorkingDir.GetPullDir(ctx.BaseRepo, ctx.Pull)
	if err != nil {
		return nil, err
	}

	plans, err := p.PendingPlanFinder.Find(pullDir)
	if err != nil {
		return nil, err
	}

	var cmds []models.ProjectCommandContext
	for _, plan := range plans {
		cmd, err := p.buildProjectCommandCtx(ctx, models.ApplyCommand, plan.ProjectName, commentCmd.Flags, plan.RepoDir, plan.RepoRelDir, plan.Workspace, commentCmd.Verbose)
		if err != nil {
			return nil, errors.Wrapf(err, "building command for dir %q", plan.RepoRelDir)
		}
		cmds = append(cmds, cmd)
	}
	return cmds, nil
}

// buildProjectApplyCommand builds an apply command for the single project
// identified by cmd.
func (p *DefaultProjectCommandBuilder) buildProjectApplyCommand(ctx *CommandContext, cmd *CommentCommand) (models.ProjectCommandContext, error) {
	workspace := DefaultWorkspace
	if cmd.Workspace != "" {
		workspace = cmd.Workspace
	}

	var projCtx models.ProjectCommandContext
	unlockFn, err := p.WorkingDirLocker.TryLock(ctx.BaseRepo.FullName, ctx.Pull.Num, workspace)
	if err != nil {
		return projCtx, err
	}
	defer unlockFn()

	repoDir, err := p.WorkingDir.GetWorkingDir(ctx.BaseRepo, ctx.Pull, workspace)
	if err != nil {
		return projCtx, err
	}

	repoRelDir := DefaultRepoRelDir
	if cmd.RepoRelDir != "" {
		repoRelDir = cmd.RepoRelDir
	}

	return p.buildProjectCommandCtx(ctx, models.ApplyCommand, cmd.ProjectName, cmd.Flags, repoDir, repoRelDir, workspace, cmd.Verbose)
}

// buildProjectCommandCtx builds a context for a single project identified
// by the parameters.
func (p *DefaultProjectCommandBuilder) buildProjectCommandCtx(
	ctx *CommandContext,
	cmd models.CommandName,
	projectName string,
	commentFlags []string,
	repoDir string,
	repoRelDir string,
	workspace string,
	verbose bool) (models.ProjectCommandContext, error) {

	projCfgPtr, repoCfgPtr, err := p.getCfg(ctx, projectName, repoRelDir, workspace, repoDir)
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
		projCfg = p.GlobalCfg.MergeProjectCfg(ctx.Log, ctx.BaseRepo.ID(), *projCfgPtr, *repoCfgPtr)
	} else {
		projCfg = p.GlobalCfg.DefaultProjCfg(ctx.Log, ctx.BaseRepo.ID(), repoRelDir, workspace)
	}

	if err := p.validateWorkspaceAllowed(repoCfgPtr, repoRelDir, workspace); err != nil {
		return models.ProjectCommandContext{}, err
	}

	automerge := DefaultAutomergeEnabled
	if repoCfgPtr != nil {
		automerge = repoCfgPtr.Automerge
	}
	return p.buildCtx(ctx, cmd, projCfg, commentFlags, automerge, verbose), nil
}

// getCfg returns the atlantis.yaml config (if it exists) for this project. If
// there is no config, then projectCfg and repoCfg will be nil.
func (p *DefaultProjectCommandBuilder) getCfg(ctx *CommandContext, projectName string, dir string, workspace string, repoDir string) (projectCfg *valid.Project, repoCfg *valid.RepoCfg, err error) {
	hasConfigFile, err := p.ParserValidator.HasRepoCfg(repoDir)
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
	repoConfig, err = p.ParserValidator.ParseRepoCfg(repoDir, p.GlobalCfg, ctx.BaseRepo.ID())
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

// validateWorkspaceAllowed returns an error if repoCfg defines projects in
// repoRelDir but none of them use workspace. We want this to be an error
// because if users have gone to the trouble of defining projects in repoRelDir
// then it's likely that if we're running a command for a workspace that isn't
// defined then they probably just typed the workspace name wrong.
func (p *DefaultProjectCommandBuilder) validateWorkspaceAllowed(repoCfg *valid.RepoCfg, repoRelDir string, workspace string) error {
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
func (p *DefaultProjectCommandBuilder) buildCtx(ctx *CommandContext,
	cmd models.CommandName,
	projCfg valid.MergedProjectCfg,
	commentArgs []string,
	automergeEnabled bool,
	verbose bool) models.ProjectCommandContext {

	var steps []valid.Step
	switch cmd {
	case models.PlanCommand:
		steps = projCfg.Workflow.Plan.Steps
	case models.ApplyCommand:
		steps = projCfg.Workflow.Apply.Steps
	}

	return models.ProjectCommandContext{
		ApplyCmd:           p.CommentBuilder.BuildApplyComment(projCfg.RepoRelDir, projCfg.Workspace, projCfg.Name),
		BaseRepo:           ctx.BaseRepo,
		EscapedCommentArgs: p.escapeArgs(commentArgs),
		AutomergeEnabled:   automergeEnabled,
		AutoplanEnabled:    projCfg.AutoplanEnabled,
		Steps:              steps,
		HeadRepo:           ctx.HeadRepo,
		Log:                ctx.Log,
		PullMergeable:      ctx.PullMergeable,
		Pull:               ctx.Pull,
		ProjectName:        projCfg.Name,
		ApplyRequirements:  projCfg.ApplyRequirements,
		RePlanCmd:          p.CommentBuilder.BuildPlanComment(projCfg.RepoRelDir, projCfg.Workspace, projCfg.Name, commentArgs),
		RepoRelDir:         projCfg.RepoRelDir,
		RepoConfigVersion:  projCfg.RepoCfgVersion,
		TerraformVersion:   projCfg.TerraformVersion,
		User:               ctx.User,
		Verbose:            verbose,
		Workspace:          projCfg.Workspace,
	}
}

func (p *DefaultProjectCommandBuilder) escapeArgs(args []string) []string {
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
