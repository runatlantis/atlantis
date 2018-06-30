package events

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/pkg/fileutils"
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
	RequireApproval     bool
	AllowRepoConfig     bool
	AllowRepoConfigFlag string
}

type TerraformExec interface {
	RunCommandWithVersion(log *logging.SimpleLogger, path string, args []string, v *version.Version, workspace string) (string, error)
}

func (p *DefaultProjectCommandBuilder) BuildAutoplanCommands(ctx *CommandContext) ([]models.ProjectCommandContext, error) {
	// Need to lock the workspace we're about to clone to.
	workspace := DefaultWorkspace
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
	ctx.Log.Debug("parsing config file")
	config, err := p.ParserValidator.ReadConfig(repoDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	noAtlantisYAML := os.IsNotExist(err)
	if noAtlantisYAML {
		ctx.Log.Info("found no %s file", yaml.AtlantisYAMLFilename)
	} else {
		ctx.Log.Info("successfully parsed %s file", yaml.AtlantisYAMLFilename)
		if !p.AllowRepoConfig {
			return nil, fmt.Errorf("%s files not allowed because Atlantis is not running with --%s", yaml.AtlantisYAMLFilename, p.AllowRepoConfigFlag)
		}
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
	if noAtlantisYAML {
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
				Workspace:     DefaultWorkspace,
			})
		}
	} else {
		// Otherwise, we use the projects that match the WhenModified fields
		// in the config file.
		matchingProjects, err := p.matchingProjects(ctx.Log, modifiedFiles, config)
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

	unlockFn, err := p.WorkingDirLocker.TryLock(ctx.BaseRepo.FullName, cmd.Workspace, ctx.Pull.Num)
	if err != nil {
		return projCtx, err
	}
	defer unlockFn()

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
		BaseRepo:                ctx.BaseRepo,
		HeadRepo:                ctx.HeadRepo,
		Pull:                    ctx.Pull,
		User:                    ctx.User,
		Log:                     ctx.Log,
		CommentArgs:             cmd.Flags,
		Workspace:               workspace,
		RepoRelDir:              dir,
		ProjectConfig:           projCfg,
		GlobalConfig:            globalCfg,
		RequireApprovalOverride: p.RequireApproval,
	}, nil
}

func (p *DefaultProjectCommandBuilder) getCfg(projectName string, dir string, workspace string, repoDir string) (*valid.Project, *valid.Config, error) {
	globalCfg, err := p.ParserValidator.ReadConfig(repoDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, nil, err
	}
	hasAtlantisYAML := !os.IsNotExist(err)
	if !hasAtlantisYAML && projectName != "" {
		return nil, nil, fmt.Errorf("cannot specify a project name unless an %s file exists to configure projects", yaml.AtlantisYAMLFilename)
	}
	if !hasAtlantisYAML {
		return nil, nil, nil
	}
	if !p.AllowRepoConfig {
		return nil, nil, fmt.Errorf("%s files not allowed because Atlantis is not running with --%s", yaml.AtlantisYAMLFilename, p.AllowRepoConfigFlag)
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

// matchingProjects returns the list of projects whose WhenModified fields match
// any of the modifiedFiles.
func (p *DefaultProjectCommandBuilder) matchingProjects(log *logging.SimpleLogger, modifiedFiles []string, config valid.Config) ([]valid.Project, error) {
	var projects []valid.Project
	for _, project := range config.Projects {
		log.Debug("checking if project at dir %q workspace %q was modified", project.Dir, project.Workspace)
		if !project.Autoplan.Enabled {
			log.Debug("autoplan disabled, ignoring")
			continue
		}
		// Prepend project dir to when modified patterns because the patterns
		// are relative to the project dirs but our list of modified files is
		// relative to the repo root.
		var whenModifiedRelToRepoRoot []string
		for _, wm := range project.Autoplan.WhenModified {
			whenModifiedRelToRepoRoot = append(whenModifiedRelToRepoRoot, filepath.Join(project.Dir, wm))
		}
		pm, err := fileutils.NewPatternMatcher(whenModifiedRelToRepoRoot)
		if err != nil {
			return nil, errors.Wrapf(err, "matching modified files with patterns: %v", project.Autoplan.WhenModified)
		}

		// If any of the modified files matches the pattern then this project is
		// considered modified.
		for _, file := range modifiedFiles {
			match, err := pm.Matches(file)
			if err != nil {
				log.Debug("match err for file %q: %s", file, err)
				continue
			}
			if match {
				log.Debug("file %q matched pattern", file)
				projects = append(projects, project)
				break
			}
		}
	}
	// todo: check if dir is deleted though
	return projects, nil
}
