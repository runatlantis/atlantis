package events

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/yaml"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
	"github.com/runatlantis/atlantis/server/fetchers"
	"github.com/runatlantis/atlantis/server/logging"
)

const (
	// DefaultRepoRelDir is the default directory we run commands in, relative
	// to the root of the repo.
	DefaultRepoRelDir = "."
	// DefaultWorkspace is the default Terraform workspace we run commands in.
	// This is also Terraform's default workspace.
	DefaultWorkspace = "default"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_project_command_builder.go ProjectCommandBuilder

// ProjectCommandBuilder builds commands that run on individual projects.
type ProjectCommandBuilder interface {
	// BuildAutoplanCommands builds project commands that will run plan on
	// the projects determined to be modified.
	BuildAutoplanCommands(ctx *CommandContext) ([]models.ProjectCommandContext, error)
	// BuildPlanCommands builds project plan commands for this comment. If the
	// comment doesn't specify one project then there may be multiple commands
	// to be run.
	BuildPlanCommands(ctx *CommandContext, commentCommand *CommentCommand) ([]models.ProjectCommandContext, error)
	// BuildApplyCommands builds project apply commands for this comment. If the
	// comment doesn't specify one project then there may be multiple commands
	// to be run.
	BuildApplyCommands(ctx *CommandContext, commentCommand *CommentCommand) ([]models.ProjectCommandContext, error)
}

// DefaultProjectCommandBuilder implements ProjectCommandBuilder.
// This class combines the data from the comment and any repo config file or
// Atlantis server config and then generates a set of contexts.
type DefaultProjectCommandBuilder struct {
	ParserValidator     *yaml.ParserValidator
	ProjectFinder       ProjectFinder
	VCSClient           vcs.Client
	WorkingDir          WorkingDir
	WorkingDirLocker    WorkingDirLocker
	AllowRepoConfig     bool
	AllowRepoConfigFlag string
	PendingPlanFinder   *DefaultPendingPlanFinder
	CommentBuilder      CommentBuilder
	FetcherConfig       fetchers.FetcherConfig
}

// TFCommandRunner runs Terraform commands.
type TFCommandRunner interface {
	// RunCommandWithVersion runs a Terraform command using the version v.
	RunCommandWithVersion(log *logging.SimpleLogger, path string, args []string, v *version.Version, workspace string) (string, error)
}

// BuildAutoplanCommands builds project commands that will run plan on
// the projects determined to be modified.
func (p *DefaultProjectCommandBuilder) BuildAutoplanCommands(ctx *CommandContext) ([]models.ProjectCommandContext, error) {
	cmds, err := p.buildPlanAllCommands(ctx, nil, false)
	if err != nil {
		return nil, err
	}
	// Filter out projects where autoplanning is specifically disabled.
	var autoplanEnabled []models.ProjectCommandContext
	for _, cmd := range cmds {
		if cmd.ProjectConfig != nil && !cmd.ProjectConfig.Autoplan.Enabled {
			ctx.Log.Debug("ignoring project at dir %q, workspace: %q because autoplan is disabled", cmd.RepoRelDir, cmd.Workspace)
			continue
		}
		autoplanEnabled = append(autoplanEnabled, cmd)
	}
	return autoplanEnabled, nil
}

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

	repoDir, err := p.WorkingDir.Clone(ctx.Log, ctx.BaseRepo, ctx.HeadRepo, ctx.Pull, workspace)
	if err != nil {
		return nil, err
	}

	switch p.FetcherConfig.ConfigType {
	case fetchers.Github:
		ctx.Log.Debug("fetching github config...")
		configString, err := p.FetcherConfig.GithubConfig.FetchConfig()
		if err != nil {
			return nil, errors.Wrap(err, "could not fetch github config")
		}

		configFilePath := filepath.Join(repoDir, yaml.AtlantisYAMLFilename)
		if configString != "" {
			err = ioutil.WriteFile(configFilePath, []byte(configString), 0644)
		} else {
			return nil, errors.Wrap(nil, "empty config file retrieved")
		}
		if err != nil {
			return nil, errors.Wrap(err, "could not write config file")
		}
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
				CommentArgs:   commentFlags,
				Workspace:     DefaultWorkspace,
				Verbose:       verbose,
				RePlanCmd:     p.CommentBuilder.BuildPlanComment(mp.Path, DefaultWorkspace, "", commentFlags),
				ApplyCmd:      p.CommentBuilder.BuildApplyComment(mp.Path, DefaultWorkspace, ""),
				PullMergeable: ctx.PullMergeable,
			})
		}
	} else {
		// Otherwise, we use the projects that match the WhenModified fields
		// in the config file.
		matchingProjects, err := p.ProjectFinder.DetermineProjectsViaConfig(ctx.Log, modifiedFiles, config, repoDir)
		if err != nil {
			return nil, err
		}
		ctx.Log.Info("%d projects are to be planned based on their when_modified config", len(matchingProjects))

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
				CommentArgs:   commentFlags,
				Workspace:     mp.Workspace,
				RepoRelDir:    mp.Dir,
				ProjectConfig: &mp,
				GlobalConfig:  &config,
				Verbose:       verbose,
				RePlanCmd:     p.CommentBuilder.BuildPlanComment(mp.Dir, mp.Workspace, mp.GetName(), commentFlags),
				ApplyCmd:      p.CommentBuilder.BuildApplyComment(mp.Dir, mp.Workspace, mp.GetName()),
				PullMergeable: ctx.PullMergeable,
			})
		}
	}
	return projCtxs, nil
}

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
	repoDir, err := p.WorkingDir.Clone(ctx.Log, ctx.BaseRepo, ctx.HeadRepo, ctx.Pull, workspace)
	if err != nil {
		return pcc, err
	}

	repoRelDir := DefaultRepoRelDir
	if cmd.RepoRelDir != "" {
		repoRelDir = cmd.RepoRelDir
	}

	return p.buildProjectCommandCtx(ctx, cmd.ProjectName, cmd.Flags, repoDir, repoRelDir, workspace)
}

// BuildPlanCommands builds project plan commands for this comment. If the
// comment doesn't specify one project then there may be multiple commands
// to be run.
func (p *DefaultProjectCommandBuilder) BuildPlanCommands(ctx *CommandContext, cmd *CommentCommand) ([]models.ProjectCommandContext, error) {
	if !cmd.IsForSpecificProject() {
		return p.buildPlanAllCommands(ctx, cmd.Flags, cmd.Verbose)
	}
	pcc, err := p.buildProjectPlanCommand(ctx, cmd)
	if err != nil {
		return nil, err
	}
	return []models.ProjectCommandContext{pcc}, nil
}

func (p *DefaultProjectCommandBuilder) buildApplyAllCommands(ctx *CommandContext, commentCmd *CommentCommand) ([]models.ProjectCommandContext, error) {
	// lock all dirs in this pull request
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
		cmd, err := p.buildProjectCommandCtx(ctx, commentCmd.ProjectName, commentCmd.Flags, plan.RepoDir, plan.RepoRelDir, plan.Workspace)
		if err != nil {
			return nil, errors.Wrapf(err, "building command for dir %q", plan.RepoRelDir)
		}
		cmds = append(cmds, cmd)
	}
	return cmds, nil
}

// BuildApplyCommands builds project apply commands for this comment. If the
// comment doesn't specify one project then there may be multiple commands
// to be run.
func (p *DefaultProjectCommandBuilder) BuildApplyCommands(ctx *CommandContext, cmd *CommentCommand) ([]models.ProjectCommandContext, error) {
	if !cmd.IsForSpecificProject() {
		return p.buildApplyAllCommands(ctx, cmd)
	}
	pac, err := p.buildProjectApplyCommand(ctx, cmd)
	if err != nil {
		return nil, err
	}
	return []models.ProjectCommandContext{pac}, nil
}

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

	return p.buildProjectCommandCtx(ctx, cmd.ProjectName, cmd.Flags, repoDir, repoRelDir, workspace)
}

func (p *DefaultProjectCommandBuilder) buildProjectCommandCtx(ctx *CommandContext, projectName string, commentFlags []string, repoDir string, repoRelDir string, workspace string) (models.ProjectCommandContext, error) {
	projCfg, globalCfg, err := p.getCfg(projectName, repoRelDir, workspace, repoDir)
	if err != nil {
		return models.ProjectCommandContext{}, err
	}

	// Override any dir/workspace defined on the comment with what was
	// defined in config. This shouldn't matter since we don't allow comments
	// with both project name and dir/workspace.
	if projCfg != nil {
		repoRelDir = projCfg.Dir
		workspace = projCfg.Workspace
	}

	if err := p.validateWorkspaceAllowed(globalCfg, repoRelDir, workspace); err != nil {
		return models.ProjectCommandContext{}, err
	}

	return models.ProjectCommandContext{
		BaseRepo:      ctx.BaseRepo,
		HeadRepo:      ctx.HeadRepo,
		Pull:          ctx.Pull,
		User:          ctx.User,
		Log:           ctx.Log,
		CommentArgs:   commentFlags,
		Workspace:     workspace,
		RepoRelDir:    repoRelDir,
		ProjectConfig: projCfg,
		GlobalConfig:  globalCfg,
		RePlanCmd:     p.CommentBuilder.BuildPlanComment(repoRelDir, workspace, projectName, commentFlags),
		ApplyCmd:      p.CommentBuilder.BuildApplyComment(repoRelDir, workspace, projectName),
		PullMergeable: ctx.PullMergeable,
	}, nil
}

func (p *DefaultProjectCommandBuilder) getCfg(projectName string, dir string, workspace string, repoDir string) (projectCfg *valid.Project, globalCfg *valid.Config, err error) {
	hasConfigFile, err := p.ParserValidator.HasConfigFile(repoDir)
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

	if !p.AllowRepoConfig {
		err = fmt.Errorf("%s files not allowed because Atlantis is not running with --%s", yaml.AtlantisYAMLFilename, p.AllowRepoConfigFlag)
		return
	}

	globalCfgStruct, err := p.ParserValidator.ReadConfig(repoDir)
	if err != nil {
		return
	}
	globalCfg = &globalCfgStruct

	// If they've specified a project by name we look it up. Otherwise we
	// use the dir and workspace.
	if projectName != "" {
		projectCfg = globalCfg.FindProjectByName(projectName)
		if projectCfg == nil {
			err = fmt.Errorf("no project with name %q is defined in %s", projectName, yaml.AtlantisYAMLFilename)
			return
		}
		return
	}

	projCfgs := globalCfg.FindProjectsByDirWorkspace(dir, workspace)
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

// validateWorkspaceAllowed returns an error if there are projects configured
// in globalCfg for repoRelDir and none of those projects use workspace.
func (p *DefaultProjectCommandBuilder) validateWorkspaceAllowed(globalCfg *valid.Config, repoRelDir string, workspace string) error {
	if globalCfg == nil {
		return nil
	}

	projects := globalCfg.FindProjectsByDir(repoRelDir)

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
