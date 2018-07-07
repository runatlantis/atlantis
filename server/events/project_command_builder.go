package events

import (
	"fmt"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/yaml"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
	"github.com/runatlantis/atlantis/server/logging"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_project_command_builder.go ProjectCommandBuilder

type ProjectCommandBuilder interface {
	BuildAutoplanCommands(ctx *CommandContext) ([]models.ProjectCommandContext, error)
	BuildPlanCommand(ctx *CommandContext, commentCommand *CommentCommand) (models.ProjectCommandContext, error)
	BuildApplyCommand(ctx *CommandContext, commentCommand *CommentCommand) (models.ProjectCommandContext, error)
}

type DefaultProjectCommandBuilder struct {
	ParserValidator     *yaml.ParserValidator
	ProjectFinder       ProjectFinder
	VCSClient           vcs.ClientProxy
	WorkingDir          WorkingDir
	WorkingDirLocker    WorkingDirLocker
	AllowRepoConfig     bool
	AllowRepoConfigFlag string
}

type TerraformExec interface {
	RunCommandWithVersion(log *logging.SimpleLogger, path string, args []string, v *version.Version, workspace string) (string, error)
}

func (p *DefaultProjectCommandBuilder) BuildAutoplanCommands(ctx *CommandContext) ([]models.ProjectCommandContext, error) {
	// Need to lock the workspace we're about to clone to.
	workspace := "default"
	unlockFn, err := p.WorkingDirLocker.TryLock(ctx.BaseRepo.FullName, workspace, ctx.Pull.Num)
	if err != nil {
		ctx.Log.Warn("workspace was locked")
		return nil, err
	}
	ctx.Log.Debug("got workspace lock")
	defer unlockFn()

	repoDir, err := p.WorkingDir.Clone(ctx.Log, ctx.BaseRepo, ctx.HeadRepo, ctx.Pull, workspace)
	if err != nil {
		return nil, err
	}

	// Parse config file if it exists.
	var config valid.Config
	hasConfigFile, err := p.ParserValidator.HasConfigFile(repoDir)
	if err != nil {
		return nil, errors.Wrapf(err, "looking for %s file in %q", yaml.AtlantisYAMLFilename, repoDir)
	}
	if hasConfigFile {
		if !p.AllowRepoConfig {
			return nil, fmt.Errorf("%s files not allowed because Atlantis is not running with --%s", yaml.AtlantisYAMLFilename, p.AllowRepoConfigFlag)
		}
		config, err = p.ParserValidator.ReadConfig(repoDir)
		if err != nil {
			return nil, err
		}
		ctx.Log.Info("successfully parsed %s file", yaml.AtlantisYAMLFilename)
	} else {
		ctx.Log.Info("found no %s file", yaml.AtlantisYAMLFilename)
	}

	// We'll need the list of modified files.
	modifiedFiles, err := p.VCSClient.GetModifiedFiles(ctx.BaseRepo, ctx.Pull)
	if err != nil {
		return nil, err
	}
	ctx.Log.Debug("%d files were modified in this pull request", len(modifiedFiles))

	// Prepare the project contexts so the ProjectCommandRunner can execute.
	var projCtxs []models.ProjectCommandContext

	// If there is no config file, then we try to plan for each project that
	// was modified in the pull request.
	if !hasConfigFile {
		modifiedProjects := p.ProjectFinder.DetermineProjects(ctx.Log, modifiedFiles, ctx.BaseRepo.FullName, repoDir)
		ctx.Log.Info("automatically determined that there were %d projects modified in this pull request: %s", len(modifiedProjects), modifiedProjects)
		for _, mp := range modifiedProjects {
			projCtxs = append(projCtxs, models.ProjectCommandContext{
				BaseRepo:      ctx.BaseRepo,
				HeadRepo:      ctx.HeadRepo,
				Pull:          ctx.Pull,
				User:          ctx.User,
				Log:           ctx.Log,
				RepoRelDir:    mp.Path,
				ProjectConfig: nil,
				GlobalConfig:  nil,
				CommentArgs:   nil,
				Workspace:     "default",
			})
		}
	} else {
		// Otherwise, we use the projects that match the WhenModified fields
		// in the config file.
		matchingProjects, err := p.ProjectFinder.DetermineProjectsViaConfig(ctx.Log, modifiedFiles, config, repoDir)
		if err != nil {
			return nil, err
		}
		ctx.Log.Info("%d projects are to be autoplanned based on their when_modified config", len(matchingProjects))

		// Use for i instead of range because need to get the pointer to the
		// project config.
		for i := 0; i < len(matchingProjects); i++ {
			mp := matchingProjects[i]
			projCtxs = append(projCtxs, models.ProjectCommandContext{
				BaseRepo:      ctx.BaseRepo,
				HeadRepo:      ctx.HeadRepo,
				Pull:          ctx.Pull,
				User:          ctx.User,
				Log:           ctx.Log,
				CommentArgs:   nil,
				Workspace:     mp.Workspace,
				RepoRelDir:    mp.Dir,
				ProjectConfig: &mp,
				GlobalConfig:  &config,
			})
		}
	}
	return projCtxs, nil
}

func (p *DefaultProjectCommandBuilder) BuildPlanCommand(ctx *CommandContext, cmd *CommentCommand) (models.ProjectCommandContext, error) {
	var projCtx models.ProjectCommandContext

	ctx.Log.Debug("building plan command")
	unlockFn, err := p.WorkingDirLocker.TryLock(ctx.BaseRepo.FullName, cmd.Workspace, ctx.Pull.Num)
	if err != nil {
		return projCtx, err
	}
	defer unlockFn()

	ctx.Log.Debug("cloning repository")
	repoDir, err := p.WorkingDir.Clone(ctx.Log, ctx.BaseRepo, ctx.HeadRepo, ctx.Pull, cmd.Workspace)
	if err != nil {
		return projCtx, err
	}

	return p.buildProjectCommandCtx(ctx, cmd, repoDir)
}

func (p *DefaultProjectCommandBuilder) BuildApplyCommand(ctx *CommandContext, cmd *CommentCommand) (models.ProjectCommandContext, error) {
	var projCtx models.ProjectCommandContext

	unlockFn, err := p.WorkingDirLocker.TryLock(ctx.BaseRepo.FullName, cmd.Workspace, ctx.Pull.Num)
	if err != nil {
		return projCtx, err
	}
	defer unlockFn()

	repoDir, err := p.WorkingDir.GetWorkingDir(ctx.BaseRepo, ctx.Pull, cmd.Workspace)
	if err != nil {
		return projCtx, err
	}

	return p.buildProjectCommandCtx(ctx, cmd, repoDir)
}

func (p *DefaultProjectCommandBuilder) buildProjectCommandCtx(ctx *CommandContext, cmd *CommentCommand, repoDir string) (models.ProjectCommandContext, error) {
	projCfg, globalCfg, err := p.getCfg(cmd.ProjectName, cmd.RepoRelDir, cmd.Workspace, repoDir)
	if err != nil {
		return models.ProjectCommandContext{}, err
	}

	// Override any dir/workspace defined on the comment with what was
	// defined in config. This shouldn't matter since we don't allow comments
	// with both project name and dir/workspace.
	dir := cmd.RepoRelDir
	workspace := cmd.Workspace
	if projCfg != nil {
		dir = projCfg.Dir
		workspace = projCfg.Workspace
	}

	return models.ProjectCommandContext{
		BaseRepo:      ctx.BaseRepo,
		HeadRepo:      ctx.HeadRepo,
		Pull:          ctx.Pull,
		User:          ctx.User,
		Log:           ctx.Log,
		CommentArgs:   cmd.Flags,
		Workspace:     workspace,
		RepoRelDir:    dir,
		ProjectConfig: projCfg,
		GlobalConfig:  globalCfg,
	}, nil
}

func (p *DefaultProjectCommandBuilder) getCfg(projectName string, dir string, workspace string, repoDir string) (*valid.Project, *valid.Config, error) {
	hasConfigFile, err := p.ParserValidator.HasConfigFile(repoDir)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "looking for %s file in %q", yaml.AtlantisYAMLFilename, repoDir)
	}
	if !hasConfigFile {
		if projectName != "" {
			return nil, nil, fmt.Errorf("cannot specify a project name unless an %s file exists to configure projects", yaml.AtlantisYAMLFilename)
		}
		return nil, nil, nil
	}

	if !p.AllowRepoConfig {
		return nil, nil, fmt.Errorf("%s files not allowed because Atlantis is not running with --%s", yaml.AtlantisYAMLFilename, p.AllowRepoConfigFlag)
	}

	globalCfg, err := p.ParserValidator.ReadConfig(repoDir)
	if err != nil {
		return nil, nil, err
	}

	// If they've specified a project by name we look it up. Otherwise we
	// use the dir and workspace.
	if projectName != "" {
		projCfg := globalCfg.FindProjectByName(projectName)
		if projCfg == nil {
			return nil, nil, fmt.Errorf("no project with name %q is defined in %s", projectName, yaml.AtlantisYAMLFilename)
		}
		return projCfg, &globalCfg, nil
	}

	projCfgs := globalCfg.FindProjectsByDirWorkspace(dir, workspace)
	if len(projCfgs) == 0 {
		return nil, nil, nil
	}
	if len(projCfgs) > 1 {
		return nil, nil, fmt.Errorf("must specify project name: more than one project defined in %s matched dir: %q workspace: %q", yaml.AtlantisYAMLFilename, dir, workspace)
	}
	return &projCfgs[0], &globalCfg, nil
}
