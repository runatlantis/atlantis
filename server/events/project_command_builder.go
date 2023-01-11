package events

import (
	"fmt"
	"os"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/logging"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/vcs"
)

const (
	// DefaultRepoRelDir is the default directory we run commands in, relative
	// to the root of the repo.
	DefaultRepoRelDir = "."
	// DefaultWorkspace is the default Terraform workspace we run commands in.
	// This is also Terraform's default workspace.
	DefaultWorkspace = "default"
	// DefaultParallelApplyEnabled is the default for the parallel apply setting.
	DefaultParallelApplyEnabled = false
	// DefaultParallelPlanEnabled is the default for the parallel plan setting.
	DefaultParallelPlanEnabled = false
	// InfiniteProjectLimitPerPR is the default setting for number of projects per PR.
	// this is set to -1 to signify no limit.
	InfiniteProjectsPerPR = -1
)

func NewProjectCommandBuilder(
	projectContextBuilder ProjectCommandContextBuilder,
	parserValidator *config.ParserValidator,
	projectFinder ProjectFinder,
	vcsClient vcs.Client,
	workingDir WorkingDir,
	workingDirLocker WorkingDirLocker,
	globalCfg valid.GlobalCfg,
	pendingPlanFinder *DefaultPendingPlanFinder,
	EnableRegExpCmd bool,
	AutoplanFileList string,
	logger logging.Logger,
	limit int,
) ProjectCommandBuilder {
	var projectCommandBuilder ProjectCommandBuilder = &DefaultProjectCommandBuilder{
		ParserValidator:              parserValidator,
		ProjectFinder:                projectFinder,
		VCSClient:                    vcsClient,
		WorkingDir:                   workingDir,
		WorkingDirLocker:             workingDirLocker,
		GlobalCfg:                    globalCfg,
		PendingPlanFinder:            pendingPlanFinder,
		EnableRegExpCmd:              EnableRegExpCmd,
		AutoplanFileList:             AutoplanFileList,
		ProjectCommandContextBuilder: projectContextBuilder,
	}

	projectCommandBuilder = &SizeLimitedProjectCommandBuilder{
		Limit:                 limit,
		ProjectCommandBuilder: projectCommandBuilder,
	}

	return &InstrumentedProjectCommandBuilder{
		ProjectCommandBuilder: projectCommandBuilder,
		Logger:                logger,
	}
}

type ProjectPlanCommandBuilder interface {
	// BuildAutoplanCommands builds project commands that will run plan on
	// the projects determined to be modified.
	BuildAutoplanCommands(ctx *command.Context) ([]command.ProjectContext, error)
	// BuildPlanCommands builds project plan commands for this ctx and comment. If
	// comment doesn't specify one project then there may be multiple commands
	// to be run.
	BuildPlanCommands(ctx *command.Context, comment *command.Comment) ([]command.ProjectContext, error)
}

type ProjectApplyCommandBuilder interface {
	// BuildApplyCommands builds project Apply commands for this ctx and comment. If
	// comment doesn't specify one project then there may be multiple commands
	// to be run.
	BuildApplyCommands(ctx *command.Context, comment *command.Comment) ([]command.ProjectContext, error)
}

type ProjectApprovePoliciesCommandBuilder interface {
	// BuildApprovePoliciesCommands builds project PolicyCheck commands for this ctx and comment.
	BuildApprovePoliciesCommands(ctx *command.Context, comment *command.Comment) ([]command.ProjectContext, error)
}

type ProjectVersionCommandBuilder interface {
	// BuildVersionCommands builds project Version commands for this ctx and comment. If
	// comment doesn't specify one project then there may be multiple commands
	// to be run.
	BuildVersionCommands(ctx *command.Context, comment *command.Comment) ([]command.ProjectContext, error)
}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_project_command_builder.go ProjectCommandBuilder

// ProjectCommandBuilder builds commands that run on individual projects.
type ProjectCommandBuilder interface {
	ProjectPlanCommandBuilder
	ProjectPolicyCheckCommandBuilder
	ProjectApplyCommandBuilder
	ProjectApprovePoliciesCommandBuilder
	ProjectVersionCommandBuilder
}

// DefaultProjectCommandBuilder implements ProjectCommandBuilder.
// This class combines the data from the comment and any atlantis.yaml file or
// Atlantis server config and then generates a set of contexts.
type DefaultProjectCommandBuilder struct {
	ParserValidator              *config.ParserValidator
	ProjectFinder                ProjectFinder
	VCSClient                    vcs.Client
	WorkingDir                   WorkingDir
	WorkingDirLocker             WorkingDirLocker
	GlobalCfg                    valid.GlobalCfg
	PendingPlanFinder            *DefaultPendingPlanFinder
	ProjectCommandContextBuilder ProjectCommandContextBuilder
	EnableRegExpCmd              bool
	AutoplanFileList             string
	EnableDiffMarkdownFormat     bool
}

// See ProjectCommandBuilder.BuildAutoplanCommands.
func (p *DefaultProjectCommandBuilder) BuildAutoplanCommands(ctx *command.Context) ([]command.ProjectContext, error) {
	projCtxs, err := p.buildPlanAllCommands(ctx, nil, false, "")
	if err != nil {
		return nil, err
	}
	var autoplanEnabled []command.ProjectContext
	for _, projCtx := range projCtxs {
		if !projCtx.AutoplanEnabled {
			continue
		}
		autoplanEnabled = append(autoplanEnabled, projCtx)
	}
	return autoplanEnabled, nil
}

// See ProjectCommandBuilder.BuildPlanCommands.
func (p *DefaultProjectCommandBuilder) BuildPlanCommands(ctx *command.Context, cmd *command.Comment) ([]command.ProjectContext, error) {
	if !cmd.IsForSpecificProject() {
		return p.buildPlanAllCommands(ctx, cmd.Flags, cmd.ForceApply, cmd.LogLevel)
	}
	pcc, err := p.buildProjectPlanCommand(ctx, cmd)
	return pcc, err
}

func (p *DefaultProjectCommandBuilder) BuildPolicyCheckCommands(ctx *command.Context) ([]command.ProjectContext, error) {
	unlockFn, err := p.WorkingDirLocker.TryLockPull(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num)
	if err != nil {
		return nil, err
	}
	defer unlockFn()

	pullDir, err := p.WorkingDir.GetPullDir(ctx.Pull.BaseRepo, ctx.Pull)
	if err != nil {
		return nil, err
	}

	plans, err := p.PendingPlanFinder.Find(pullDir)
	if err != nil {
		return nil, err
	}

	// use the default repository workspace because it is the only one guaranteed to have an atlantis.yaml,
	// other workspaces will not have the file if they are using pre_workflow_hooks to generate it dynamically
	defaultRepoDir, err := p.WorkingDir.GetWorkingDir(ctx.Pull.BaseRepo, ctx.Pull, DefaultWorkspace)
	if err != nil {
		return nil, err
	}

	var cmds []command.ProjectContext
	for _, plan := range plans {
		commands, err := p.buildProjectCommandCtx(ctx, command.PolicyCheck, plan.ProjectName, []string{}, defaultRepoDir, plan.RepoRelDir, plan.Workspace, false, "")
		if err != nil {
			return nil, errors.Wrapf(err, "building command for dir %q", plan.RepoRelDir)
		}
		cmds = append(cmds, commands...)
	}
	return cmds, nil
}

// See ProjectCommandBuilder.BuildApplyCommands.
func (p *DefaultProjectCommandBuilder) BuildApplyCommands(ctx *command.Context, cmd *command.Comment) ([]command.ProjectContext, error) {
	if !cmd.IsForSpecificProject() {
		return p.buildAllProjectCommands(ctx, cmd)
	}
	pac, err := p.buildProjectApplyCommand(ctx, cmd)
	return pac, err
}

func (p *DefaultProjectCommandBuilder) BuildApprovePoliciesCommands(ctx *command.Context, cmd *command.Comment) ([]command.ProjectContext, error) {
	return p.buildAllProjectCommands(ctx, cmd)
}

func (p *DefaultProjectCommandBuilder) BuildVersionCommands(ctx *command.Context, cmd *command.Comment) ([]command.ProjectContext, error) {
	if !cmd.IsForSpecificProject() {
		return p.buildAllProjectCommands(ctx, cmd)
	}
	pac, err := p.buildProjectVersionCommand(ctx, cmd)
	return pac, err
}

// buildPlanAllCommands builds plan contexts for all projects we determine were
// modified in this ctx.
func (p *DefaultProjectCommandBuilder) buildPlanAllCommands(ctx *command.Context, commentFlags []string, forceApply bool, logLevel string) ([]command.ProjectContext, error) {
	// We'll need the list of modified files.
	modifiedFiles, err := p.VCSClient.GetModifiedFiles(ctx.Pull.BaseRepo, ctx.Pull)
	if err != nil {
		return nil, err
	}

	// Need to lock the workspace we're about to clone to.
	workspace := DefaultWorkspace

	unlockFn, err := p.WorkingDirLocker.TryLock(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, workspace)
	if err != nil {
		ctx.Log.WarnContext(ctx.RequestCtx, "workspace was locked")
		return nil, err
	}
	defer unlockFn()

	repoDir, _, err := p.WorkingDir.Clone(ctx.Log, ctx.HeadRepo, ctx.Pull, workspace)
	if err != nil {
		return nil, err
	}

	// Parse config file if it exists.
	hasRepoCfg, err := p.ParserValidator.HasRepoCfg(repoDir)
	if err != nil {
		return nil, errors.Wrapf(err, "looking for %s file in %q", config.AtlantisYAMLFilename, repoDir)
	}

	var projCtxs []command.ProjectContext
	if hasRepoCfg {
		// If there's a repo cfg then we'll use it to figure out which projects
		// should be planed.
		repoCfg, err := p.ParserValidator.ParseRepoCfg(repoDir, p.GlobalCfg, ctx.Pull.BaseRepo.ID())
		if err != nil {
			return nil, errors.Wrapf(err, "parsing %s", config.AtlantisYAMLFilename)
		}
		ctx.Log.InfoContext(ctx.RequestCtx, fmt.Sprintf("successfully parsed %s file", config.AtlantisYAMLFilename))
		matchingProjects, err := p.ProjectFinder.DetermineProjectsViaConfig(ctx.Log, modifiedFiles, repoCfg, repoDir)
		if err != nil {
			return nil, err
		}
		ctx.Log.InfoContext(ctx.RequestCtx, fmt.Sprintf("%d projects are to be planned based on their when_modified config", len(matchingProjects)))

		for _, mp := range matchingProjects {
			mergedCfg := p.GlobalCfg.MergeProjectCfg(ctx.Pull.BaseRepo.ID(), mp, repoCfg)
			contextFlags := &command.ContextFlags{
				ForceApply:    forceApply,
				ParallelApply: repoCfg.ParallelApply,
				ParallelPlan:  repoCfg.ParallelPlan,
				LogLevel:      logLevel,
			}
			projCtxs = append(projCtxs,
				p.ProjectCommandContextBuilder.BuildProjectContext(
					ctx,
					command.Plan,
					mergedCfg,
					commentFlags,
					repoDir,
					contextFlags,
				)...)
		}
	} else {
		// If there is no config file, then we'll plan each project that
		// our algorithm determines was modified.
		ctx.Log.InfoContext(ctx.RequestCtx, fmt.Sprintf("found no %s file", config.AtlantisYAMLFilename))
		modifiedProjects := p.ProjectFinder.DetermineProjects(ctx.RequestCtx, ctx.Log, modifiedFiles, ctx.Pull.BaseRepo.FullName, repoDir, p.AutoplanFileList)
		if err != nil {
			return nil, errors.Wrapf(err, "finding modified projects: %s", modifiedFiles)
		}
		ctx.Log.InfoContext(ctx.RequestCtx, fmt.Sprintf("automatically determined that there were %d projects modified in this pull request: %s", len(modifiedProjects), modifiedProjects))
		for _, mp := range modifiedProjects {
			pCfg := p.GlobalCfg.DefaultProjCfg(ctx.Log, ctx.Pull.BaseRepo.ID(), mp.Path, DefaultWorkspace)

			contextFlags := &command.ContextFlags{
				ForceApply:    forceApply,
				ParallelApply: DefaultParallelApplyEnabled,
				ParallelPlan:  DefaultParallelPlanEnabled,
				LogLevel:      logLevel,
			}
			projCtxs = append(projCtxs,
				p.ProjectCommandContextBuilder.BuildProjectContext(
					ctx,
					command.Plan,
					pCfg,
					commentFlags,
					repoDir,
					contextFlags,
				)...)
		}
	}

	return projCtxs, nil
}

// buildProjectPlanCommand builds a plan context for a single project.
// cmd must be for only one project.
func (p *DefaultProjectCommandBuilder) buildProjectPlanCommand(ctx *command.Context, cmd *command.Comment) ([]command.ProjectContext, error) {
	workspace := DefaultWorkspace
	if cmd.Workspace != "" {
		workspace = cmd.Workspace
	}

	var pcc []command.ProjectContext
	unlockFn, err := p.WorkingDirLocker.TryLock(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, workspace)
	if err != nil {
		return pcc, err
	}
	defer unlockFn()

	_, _, err = p.WorkingDir.Clone(ctx.Log, ctx.HeadRepo, ctx.Pull, workspace)
	if err != nil {
		return pcc, err
	}

	repoRelDir := DefaultRepoRelDir
	if cmd.RepoRelDir != "" {
		repoRelDir = cmd.RepoRelDir
	}

	// use the default repository workspace because it is the only one guaranteed to have an atlantis.yaml,
	// other workspaces will not have the file if they are using pre_workflow_hooks to generate it dynamically
	defaultRepoDir, err := p.WorkingDir.GetWorkingDir(ctx.Pull.BaseRepo, ctx.Pull, DefaultWorkspace)
	if err != nil {
		return pcc, err
	}

	return p.buildProjectCommandCtx(
		ctx,
		command.Plan,
		cmd.ProjectName,
		cmd.Flags,
		defaultRepoDir,
		repoRelDir,
		workspace,
		cmd.ForceApply,
		cmd.LogLevel,
	)
}

// getCfg returns the atlantis.yaml config (if it exists) for this project. If
// there is no config, then projectCfg and repoCfg will be nil.
func (p *DefaultProjectCommandBuilder) getCfg(ctx *command.Context, projectName string, dir string, workspace string, repoDir string) (projectsCfg []valid.Project, repoCfg *valid.RepoCfg, err error) {
	hasConfigFile, err := p.ParserValidator.HasRepoCfg(repoDir)
	if err != nil {
		err = errors.Wrapf(err, "looking for %s file in %q", config.AtlantisYAMLFilename, repoDir)
		return
	}
	if !hasConfigFile {
		if projectName != "" {
			err = fmt.Errorf("cannot specify a project name unless an %s file exists to configure projects", config.AtlantisYAMLFilename)
			return
		}
		return
	}

	var repoConfig valid.RepoCfg
	repoConfig, err = p.ParserValidator.ParseRepoCfg(repoDir, p.GlobalCfg, ctx.Pull.BaseRepo.ID())
	if err != nil {
		return
	}
	repoCfg = &repoConfig

	// If they've specified a project by name we look it up. Otherwise we
	// use the dir and workspace.
	if projectName != "" {
		if p.EnableRegExpCmd {
			projectsCfg = repoCfg.FindProjectsByName(projectName)
		} else {
			if p := repoCfg.FindProjectByName(projectName); p != nil {
				projectsCfg = append(projectsCfg, *p)
			}
		}
		if len(projectsCfg) == 0 {
			err = fmt.Errorf("no project with name %q is defined in %s", projectName, config.AtlantisYAMLFilename)
			return
		}
		return
	}

	projCfgs := repoCfg.FindProjectsByDirWorkspace(dir, workspace)
	if len(projCfgs) == 0 {
		return
	}
	if len(projCfgs) > 1 {
		err = fmt.Errorf("must specify project name: more than one project defined in %s matched dir: %q workspace: %q", config.AtlantisYAMLFilename, dir, workspace)
		return
	}
	projectsCfg = projCfgs
	return
}

// buildAllProjectCommands builds contexts for a command for every project that has
// pending plans in this ctx.
func (p *DefaultProjectCommandBuilder) buildAllProjectCommands(ctx *command.Context, commentCmd *command.Comment) ([]command.ProjectContext, error) {
	// Lock all dirs in this pull request (instead of a single dir) because we
	// don't know how many dirs we'll need to run the command in.
	unlockFn, err := p.WorkingDirLocker.TryLockPull(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num)
	if err != nil {
		return nil, err
	}
	defer unlockFn()

	pullDir, err := p.WorkingDir.GetPullDir(ctx.Pull.BaseRepo, ctx.Pull)
	if err != nil {
		return nil, err
	}

	plans, err := p.PendingPlanFinder.Find(pullDir)
	if err != nil {
		return nil, err
	}

	// use the default repository workspace because it is the only one guaranteed to have an atlantis.yaml,
	// other workspaces will not have the file if they are using pre_workflow_hooks to generate it dynamically
	defaultRepoDir, err := p.WorkingDir.GetWorkingDir(ctx.Pull.BaseRepo, ctx.Pull, DefaultWorkspace)
	if err != nil {
		return nil, err
	}

	var cmds []command.ProjectContext
	for _, plan := range plans {
		commentCmds, err := p.buildProjectCommandCtx(ctx, commentCmd.CommandName(), plan.ProjectName, commentCmd.Flags, defaultRepoDir, plan.RepoRelDir, plan.Workspace, commentCmd.ForceApply, commentCmd.LogLevel)
		if err != nil {
			return nil, errors.Wrapf(err, "building command for dir %q", plan.RepoRelDir)
		}
		cmds = append(cmds, commentCmds...)
	}
	return cmds, nil
}

// buildProjectApplyCommand builds an apply command for the single project
// identified by cmd.
func (p *DefaultProjectCommandBuilder) buildProjectApplyCommand(ctx *command.Context, cmd *command.Comment) ([]command.ProjectContext, error) {
	workspace := DefaultWorkspace
	if cmd.Workspace != "" {
		workspace = cmd.Workspace
	}

	var projCtx []command.ProjectContext
	unlockFn, err := p.WorkingDirLocker.TryLock(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, workspace)
	if err != nil {
		return projCtx, err
	}
	defer unlockFn()

	// use the default repository workspace because it is the only one guaranteed to have an atlantis.yaml,
	// other workspaces will not have the file if they are using pre_workflow_hooks to generate it dynamically
	repoDir, err := p.WorkingDir.GetWorkingDir(ctx.Pull.BaseRepo, ctx.Pull, DefaultWorkspace)
	if os.IsNotExist(errors.Cause(err)) {
		return projCtx, errors.New("no working directory found–did you run plan?")
	} else if err != nil {
		return projCtx, err
	}

	repoRelDir := DefaultRepoRelDir
	if cmd.RepoRelDir != "" {
		repoRelDir = cmd.RepoRelDir
	}

	return p.buildProjectCommandCtx(
		ctx,
		command.Apply,
		cmd.ProjectName,
		cmd.Flags,
		repoDir,
		repoRelDir,
		workspace,
		cmd.ForceApply,
		cmd.LogLevel,
	)
}

// buildProjectVersionCommand builds a version command for the single project
// identified by cmd.
func (p *DefaultProjectCommandBuilder) buildProjectVersionCommand(ctx *command.Context, cmd *command.Comment) ([]command.ProjectContext, error) {
	workspace := DefaultWorkspace
	if cmd.Workspace != "" {
		workspace = cmd.Workspace
	}

	var projCtx []command.ProjectContext
	unlockFn, err := p.WorkingDirLocker.TryLock(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, workspace)
	if err != nil {
		return projCtx, err
	}
	defer unlockFn()

	// use the default repository workspace because it is the only one guaranteed to have an atlantis.yaml,
	// other workspaces will not have the file if they are using pre_workflow_hooks to generate it dynamically
	repoDir, err := p.WorkingDir.GetWorkingDir(ctx.Pull.BaseRepo, ctx.Pull, DefaultWorkspace)
	if os.IsNotExist(errors.Cause(err)) {
		return projCtx, errors.New("no working directory found–did you run plan?")
	} else if err != nil {
		return projCtx, err
	}

	repoRelDir := DefaultRepoRelDir
	if cmd.RepoRelDir != "" {
		repoRelDir = cmd.RepoRelDir
	}

	return p.buildProjectCommandCtx(
		ctx,
		command.Version,
		cmd.ProjectName,
		cmd.Flags,
		repoDir,
		repoRelDir,
		workspace,
		cmd.ForceApply,
		cmd.LogLevel,
	)
}

// buildProjectCommandCtx builds a context for a single or several projects identified
// by the parameters.
func (p *DefaultProjectCommandBuilder) buildProjectCommandCtx(ctx *command.Context,
	cmd command.Name,
	projectName string,
	commentFlags []string,
	repoDir string,
	repoRelDir string,
	workspace string,
	forceApply bool,
	logLevel string) ([]command.ProjectContext, error) {

	matchingProjects, repoCfgPtr, err := p.getCfg(ctx, projectName, repoRelDir, workspace, repoDir)
	if err != nil {
		return []command.ProjectContext{}, err
	}
	var projCtxs []command.ProjectContext
	var projCfg valid.MergedProjectCfg
	parallelApply := DefaultParallelApplyEnabled
	parallelPlan := DefaultParallelPlanEnabled
	if repoCfgPtr != nil {
		parallelApply = repoCfgPtr.ParallelApply
		parallelPlan = repoCfgPtr.ParallelPlan
	}

	contextFlags := &command.ContextFlags{
		ForceApply:    forceApply,
		ParallelApply: parallelApply,
		ParallelPlan:  parallelPlan,
		LogLevel:      logLevel,
	}

	if len(matchingProjects) > 0 {
		// Override any dir/workspace defined on the comment with what was
		// defined in config. This shouldn't matter since we don't allow comments
		// with both project name and dir/workspace.
		repoRelDir = projCfg.RepoRelDir
		workspace = projCfg.Workspace
		for _, mp := range matchingProjects {
			projCfg = p.GlobalCfg.MergeProjectCfg(ctx.Pull.BaseRepo.ID(), mp, *repoCfgPtr)

			projCtxs = append(projCtxs,
				p.ProjectCommandContextBuilder.BuildProjectContext(
					ctx,
					cmd,
					projCfg,
					commentFlags,
					repoDir,
					contextFlags,
				)...)
		}
	} else {
		projCfg = p.GlobalCfg.DefaultProjCfg(
			ctx.Log,
			ctx.Pull.BaseRepo.ID(),
			repoRelDir,
			workspace,
		)

		projCtxs = append(projCtxs,
			p.ProjectCommandContextBuilder.BuildProjectContext(
				ctx,
				cmd,
				projCfg,
				commentFlags,
				repoDir,
				contextFlags,
			)...)
	}

	if err := p.validateWorkspaceAllowed(repoCfgPtr, repoRelDir, workspace); err != nil {
		return []command.ProjectContext{}, err
	}

	return projCtxs, nil
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

	return repoCfg.ValidateWorkspaceAllowed(repoRelDir, workspace)
}
