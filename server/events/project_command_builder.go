package events

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	tally "github.com/uber-go/tally/v4"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"

	"github.com/pkg/errors"

	"github.com/runatlantis/atlantis/server/core/config"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
)

const (
	// DefaultRepoRelDir is the default directory we run commands in, relative
	// to the root of the repo.
	DefaultRepoRelDir = "."
	// DefaultWorkspace is the default Terraform workspace we run commands in.
	// This is also Terraform's default workspace.
	DefaultWorkspace = "default"
	// DefaultDeleteSourceBranchOnMerge being false is the default setting whether or not to remove a source branch on merge
	DefaultDeleteSourceBranchOnMerge = false
	// DefaultAbortOnExcecutionOrderFail being false is the default setting for abort on execution group failiures
	DefaultAbortOnExcecutionOrderFail = false
)

func NewInstrumentedProjectCommandBuilder(
	logger logging.SimpleLogging,
	policyChecksSupported bool,
	parserValidator *config.ParserValidator,
	projectFinder ProjectFinder,
	vcsClient vcs.Client,
	workingDir WorkingDir,
	workingDirLocker WorkingDirLocker,
	globalCfg valid.GlobalCfg,
	pendingPlanFinder *DefaultPendingPlanFinder,
	commentBuilder CommentBuilder,
	skipCloneNoChanges bool,
	EnableRegExpCmd bool,
	EnableAutoMerge bool,
	EnableParallelPlan bool,
	EnableParallelApply bool,
	AutoDetectModuleFiles string,
	AutoplanFileList string,
	RestrictFileList bool,
	SilenceNoProjects bool,
	IncludeGitUntrackedFiles bool,
	AutoDiscoverMode string,
	scope tally.Scope,
	terraformClient terraform.Client,
) *InstrumentedProjectCommandBuilder {
	scope = scope.SubScope("builder")

	for _, m := range []string{metrics.ExecutionSuccessMetric, metrics.ExecutionErrorMetric} {
		metrics.InitCounter(scope, m)
	}

	return &InstrumentedProjectCommandBuilder{
		ProjectCommandBuilder: NewProjectCommandBuilder(
			policyChecksSupported,
			parserValidator,
			projectFinder,
			vcsClient,
			workingDir,
			workingDirLocker,
			globalCfg,
			pendingPlanFinder,
			commentBuilder,
			skipCloneNoChanges,
			EnableRegExpCmd,
			EnableAutoMerge,
			EnableParallelPlan,
			EnableParallelApply,
			AutoDetectModuleFiles,
			AutoplanFileList,
			RestrictFileList,
			SilenceNoProjects,
			IncludeGitUntrackedFiles,
			AutoDiscoverMode,
			scope,
			terraformClient,
		),
		Logger: logger,
		scope:  scope,
	}
}

func NewProjectCommandBuilder(
	policyChecksSupported bool,
	parserValidator *config.ParserValidator,
	projectFinder ProjectFinder,
	vcsClient vcs.Client,
	workingDir WorkingDir,
	workingDirLocker WorkingDirLocker,
	globalCfg valid.GlobalCfg,
	pendingPlanFinder *DefaultPendingPlanFinder,
	commentBuilder CommentBuilder,
	skipCloneNoChanges bool,
	EnableRegExpCmd bool,
	EnableAutoMerge bool,
	EnableParallelPlan bool,
	EnableParallelApply bool,
	AutoDetectModuleFiles string,
	AutoplanFileList string,
	RestrictFileList bool,
	SilenceNoProjects bool,
	IncludeGitUntrackedFiles bool,
	AutoDiscoverMode string,
	scope tally.Scope,
	terraformClient terraform.Client,
) *DefaultProjectCommandBuilder {
	return &DefaultProjectCommandBuilder{
		ParserValidator:          parserValidator,
		ProjectFinder:            projectFinder,
		VCSClient:                vcsClient,
		WorkingDir:               workingDir,
		WorkingDirLocker:         workingDirLocker,
		GlobalCfg:                globalCfg,
		PendingPlanFinder:        pendingPlanFinder,
		SkipCloneNoChanges:       skipCloneNoChanges,
		EnableRegExpCmd:          EnableRegExpCmd,
		EnableAutoMerge:          EnableAutoMerge,
		EnableParallelPlan:       EnableParallelPlan,
		EnableParallelApply:      EnableParallelApply,
		AutoDetectModuleFiles:    AutoDetectModuleFiles,
		AutoplanFileList:         AutoplanFileList,
		RestrictFileList:         RestrictFileList,
		SilenceNoProjects:        SilenceNoProjects,
		IncludeGitUntrackedFiles: IncludeGitUntrackedFiles,
		AutoDiscoverMode:         AutoDiscoverMode,
		ProjectCommandContextBuilder: NewProjectCommandContextBuilder(
			policyChecksSupported,
			commentBuilder,
			scope,
		),
		TerraformExecutor: terraformClient,
	}
}

type ProjectPlanCommandBuilder interface {
	// BuildAutoplanCommands builds project commands that will run plan on
	// the projects determined to be modified.
	BuildAutoplanCommands(ctx *command.Context) ([]command.ProjectContext, error)
	// BuildPlanCommands builds project plan commands for this ctx and comment. If
	// comment doesn't specify one project then there may be multiple commands
	// to be run.
	BuildPlanCommands(ctx *command.Context, comment *CommentCommand) ([]command.ProjectContext, error)
}

type ProjectApplyCommandBuilder interface {
	// BuildApplyCommands builds project Apply commands for this ctx and comment. If
	// comment doesn't specify one project then there may be multiple commands
	// to be run.
	BuildApplyCommands(ctx *command.Context, comment *CommentCommand) ([]command.ProjectContext, error)
}

type ProjectApprovePoliciesCommandBuilder interface {
	// BuildApprovePoliciesCommands builds project PolicyCheck commands for this ctx and comment.
	BuildApprovePoliciesCommands(ctx *command.Context, comment *CommentCommand) ([]command.ProjectContext, error)
}

type ProjectVersionCommandBuilder interface {
	// BuildVersionCommands builds project Version commands for this ctx and comment. If
	// comment doesn't specify one project then there may be multiple commands
	// to be run.
	BuildVersionCommands(ctx *command.Context, comment *CommentCommand) ([]command.ProjectContext, error)
}

type ProjectImportCommandBuilder interface {
	// BuildImportCommands builds project Import commands for this ctx and comment. If
	// comment doesn't specify one project then there may be multiple commands
	// to be run.
	BuildImportCommands(ctx *command.Context, comment *CommentCommand) ([]command.ProjectContext, error)
}

type ProjectStateCommandBuilder interface {
	// BuildStateRmCommands builds project state rm commands for this ctx and comment. If
	// comment doesn't specify one project then there may be multiple commands
	// to be run.
	BuildStateRmCommands(ctx *command.Context, comment *CommentCommand) ([]command.ProjectContext, error)
}

//go:generate pegomock generate github.com/runatlantis/atlantis/server/events --package mocks -o mocks/mock_project_command_builder.go ProjectCommandBuilder

// ProjectCommandBuilder builds commands that run on individual projects.
type ProjectCommandBuilder interface {
	ProjectPlanCommandBuilder
	ProjectApplyCommandBuilder
	ProjectApprovePoliciesCommandBuilder
	ProjectVersionCommandBuilder
	ProjectImportCommandBuilder
	ProjectStateCommandBuilder
}

// DefaultProjectCommandBuilder implements ProjectCommandBuilder.
// This class combines the data from the comment and any atlantis.yaml file or
// Atlantis server config and then generates a set of contexts.
type DefaultProjectCommandBuilder struct {
	// Parses and validates server-side repo config files and repo-level atlantis.yaml files.
	ParserValidator *config.ParserValidator
	// Determines which projects were modified in a given pull request.
	ProjectFinder ProjectFinder
	// Used to make API calls to a VCS host like GitHub or GitLab.
	VCSClient vcs.Client
	// Handles the workspace on disk for running commands.
	WorkingDir WorkingDir
	// Used to prevent multiple commands from executing at the same time for a single repo, pull, and workspace.
	WorkingDirLocker WorkingDirLocker
	// The final parsed version of the server-side repo config.
	GlobalCfg valid.GlobalCfg
	// Finds unapplied plans.
	PendingPlanFinder *DefaultPendingPlanFinder
	// Builds project command contexts for Atlantis commands.
	ProjectCommandContextBuilder ProjectCommandContextBuilder
	// User config option: Skip cloning the repo during autoplan if there are no changes to Terraform projects.
	SkipCloneNoChanges bool
	// User config option: Enable the use of regular expressions to run plan/apply commands against defined project names.
	EnableRegExpCmd bool
	// User config option: Automatically merge pull requests after all plans have been successfully applied.
	EnableAutoMerge bool
	// User config option: Whether to run plan operations in parallel.
	EnableParallelPlan bool
	// User config option: Whether to run apply operations in parallel.
	EnableParallelApply bool
	// User config option: Enables auto-planning of projects when a module dependency in the same repository has changed.
	AutoDetectModuleFiles string
	// User config option: List of file patterns to to to check if a directory contains modified files.
	AutoplanFileList string
	// User config option: Format Terraform plan output into a markdown-diff friendy format for color-coding purposes.
	EnableDiffMarkdownFormat bool
	// User config option: Block plan requests from projects outside the files modified in the pull request.
	RestrictFileList bool
	// User config option: Ignore PR if none of the modified files are part of a project.
	SilenceNoProjects bool
	// User config option: Include git untracked files in the modified file list.
	IncludeGitUntrackedFiles bool
	// User config option: Controls auto-discovery of projects in a repository.
	AutoDiscoverMode string
	// Handles the actual running of Terraform commands.
	TerraformExecutor terraform.Client
}

// See ProjectCommandBuilder.BuildAutoplanCommands.
func (p *DefaultProjectCommandBuilder) BuildAutoplanCommands(ctx *command.Context) ([]command.ProjectContext, error) {
	projCtxs, err := p.buildAllCommandsByCfg(ctx, command.Plan, "", nil, false)
	if err != nil {
		return nil, err
	}
	var autoplanEnabled []command.ProjectContext
	for _, projCtx := range projCtxs {
		if !projCtx.AutoplanEnabled {
			ctx.Log.Debug("ignoring project at dir '%s', workspace: '%s' because autoplan is disabled", projCtx.RepoRelDir, projCtx.Workspace)
			continue
		}
		autoplanEnabled = append(autoplanEnabled, projCtx)
	}
	return autoplanEnabled, nil
}

// See ProjectCommandBuilder.BuildPlanCommands.
func (p *DefaultProjectCommandBuilder) BuildPlanCommands(ctx *command.Context, cmd *CommentCommand) ([]command.ProjectContext, error) {
	if !cmd.IsForSpecificProject() {
		ctx.Log.Debug("Building plan command for all affected projects")
		return p.buildAllCommandsByCfg(ctx, cmd.CommandName(), cmd.SubName, cmd.Flags, cmd.Verbose)
	}
	ctx.Log.Debug("Building plan command for specific project with directory: '%v', workspace: '%v', project: '%v'",
		cmd.RepoRelDir, cmd.Workspace, cmd.ProjectName)
	return p.buildProjectPlanCommand(ctx, cmd)
}

// See ProjectCommandBuilder.BuildApplyCommands.
func (p *DefaultProjectCommandBuilder) BuildApplyCommands(ctx *command.Context, cmd *CommentCommand) ([]command.ProjectContext, error) {
	if !cmd.IsForSpecificProject() {
		return p.buildAllProjectCommandsByPlan(ctx, cmd)
	}
	return p.buildProjectCommand(ctx, cmd)
}

func (p *DefaultProjectCommandBuilder) BuildApprovePoliciesCommands(ctx *command.Context, cmd *CommentCommand) ([]command.ProjectContext, error) {
	if !cmd.IsForSpecificProject() {
		return p.buildAllProjectCommandsByPlan(ctx, cmd)
	}
	return p.buildProjectCommand(ctx, cmd)
}

func (p *DefaultProjectCommandBuilder) BuildVersionCommands(ctx *command.Context, cmd *CommentCommand) ([]command.ProjectContext, error) {
	if !cmd.IsForSpecificProject() {
		return p.buildAllProjectCommandsByPlan(ctx, cmd)
	}
	return p.buildProjectCommand(ctx, cmd)
}

func (p *DefaultProjectCommandBuilder) BuildImportCommands(ctx *command.Context, cmd *CommentCommand) ([]command.ProjectContext, error) {
	if !cmd.IsForSpecificProject() {
		// import discard a plan file, so use buildAllCommandsByCfg instead buildAllProjectCommandsByPlan.
		return p.buildAllCommandsByCfg(ctx, cmd.CommandName(), cmd.SubName, cmd.Flags, cmd.Verbose)
	}
	return p.buildProjectCommand(ctx, cmd)
}

func (p *DefaultProjectCommandBuilder) BuildStateRmCommands(ctx *command.Context, cmd *CommentCommand) ([]command.ProjectContext, error) {
	if !cmd.IsForSpecificProject() {
		// state rm discard a plan file, so use buildAllCommandsByCfg instead buildAllProjectCommandsByPlan.
		return p.buildAllCommandsByCfg(ctx, cmd.CommandName(), cmd.SubName, cmd.Flags, cmd.Verbose)
	}
	return p.buildProjectCommand(ctx, cmd)
}

// buildAllCommandsByCfg builds init contexts for all projects we determine were
// modified in this ctx.
func (p *DefaultProjectCommandBuilder) buildAllCommandsByCfg(ctx *command.Context, cmdName command.Name, subCmdName string, commentFlags []string, verbose bool) ([]command.ProjectContext, error) {
	// We'll need the list of modified files.
	modifiedFiles, err := p.VCSClient.GetModifiedFiles(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull)
	if err != nil {
		return nil, err
	}

	if p.IncludeGitUntrackedFiles {
		ctx.Log.Debug(("'include-git-untracked-files' option is set, getting untracked files"))
		untrackedFiles, err := p.WorkingDir.GetGitUntrackedFiles(ctx.Log, ctx.HeadRepo, ctx.Pull, DefaultWorkspace)
		if err != nil {
			return nil, err
		}
		modifiedFiles = append(modifiedFiles, untrackedFiles...)
	}

	ctx.Log.Debug("%d files were modified in this pull request. Modified files: %v", len(modifiedFiles), modifiedFiles)

	// Get default AutoDiscoverMode from userConfig/globalConfig
	defaultAutoDiscoverMode := valid.AutoDiscoverMode(p.AutoDiscoverMode)
	globalAutoDiscover := p.GlobalCfg.RepoAutoDiscoverCfg(ctx.Pull.BaseRepo.ID())
	if globalAutoDiscover != nil {
		defaultAutoDiscoverMode = globalAutoDiscover.Mode
	}

	if p.SkipCloneNoChanges && p.VCSClient.SupportsSingleFileDownload(ctx.Pull.BaseRepo) {
		repoCfgFile := p.GlobalCfg.RepoConfigFile(ctx.Pull.BaseRepo.ID())

		var hasRepoCfg bool
		var repoCfgData []byte
		var err error
		ctx.Log.Debug("Getting file content for pull request %d", ctx.Pull)
		if ctx.HeadRepo.Owner != ctx.Pull.BaseRepo.Owner {
			hasRepoCfg, repoCfgData, err = p.VCSClient.GetFileContent(ctx.Log, ctx.HeadRepo, ctx.Pull.HeadBranch, repoCfgFile)
		} else {
			hasRepoCfg, repoCfgData, err = p.VCSClient.GetFileContent(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull.HeadBranch, repoCfgFile)
		}

		if err != nil {
			return nil, errors.Wrapf(err, "downloading %s", repoCfgFile)
		}

		if hasRepoCfg {
			repoCfg, err := p.ParserValidator.ParseRepoCfgData(repoCfgData, p.GlobalCfg, ctx.Pull.BaseRepo.ID(), ctx.Pull.BaseBranch)
			if err != nil {
				return nil, errors.Wrapf(err, "parsing %s", repoCfgFile)
			}
			ctx.Log.Info("successfully parsed remote %s file", repoCfgFile)

			if repoCfg.AutoDiscover != nil {
				defaultAutoDiscoverMode = repoCfg.AutoDiscover.Mode
			}
			// If auto discover is enabled, we never want to skip cloning
			if !repoCfg.AutoDiscoverEnabled(defaultAutoDiscoverMode) {
				if len(repoCfg.Projects) > 0 {
					matchingProjects, err := p.ProjectFinder.DetermineProjectsViaConfig(ctx.Log, modifiedFiles, repoCfg, "", nil)
					if err != nil {
						return nil, err
					}
					ctx.Log.Info("%d projects are changed on MR %d based on their when_modified config", len(matchingProjects), ctx.Pull.Num)
					if len(matchingProjects) == 0 {
						ctx.Log.Info("skipping repo clone since no project was modified")
						return []command.ProjectContext{}, nil
					}
				} else {
					ctx.Log.Info("no projects are defined in %s. Will resume automatic detection", repoCfgFile)
				}
			} else {
				ctx.Log.Info("automatic project discovery enabled. Will resume automatic detection")
			}
			// NOTE: We discard this work here and end up doing it again after
			// cloning to ensure all the return values are set properly with
			// the actual clone directory.
		}
	}

	// Need to lock the workspace we're about to clone to.
	workspace := DefaultWorkspace

	unlockFn, err := p.WorkingDirLocker.TryLock(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, workspace, DefaultRepoRelDir)
	if err != nil {
		ctx.Log.Warn("workspace was locked")
		return nil, err
	}
	ctx.Log.Debug("got workspace lock")
	defer unlockFn()

	repoDir, _, err := p.WorkingDir.Clone(ctx.Log, ctx.HeadRepo, ctx.Pull, workspace)
	if err != nil {
		return nil, err
	}

	// Parse config file if it exists.
	repoCfgFile := p.GlobalCfg.RepoConfigFile(ctx.Pull.BaseRepo.ID())
	hasRepoCfg, err := p.ParserValidator.HasRepoCfg(repoDir, repoCfgFile)
	if err != nil {
		return nil, errors.Wrapf(err, "looking for '%s' file in '%s'", repoCfgFile, repoDir)
	}

	var projCtxs []command.ProjectContext
	var repoCfg valid.RepoCfg

	if hasRepoCfg {
		// If there's a repo cfg with projects then we'll use it to figure out which projects
		// should be planed.
		repoCfg, err = p.ParserValidator.ParseRepoCfg(repoDir, p.GlobalCfg, ctx.Pull.BaseRepo.ID(), ctx.Pull.BaseBranch)
		if err != nil {
			return nil, errors.Wrapf(err, "parsing %s", repoCfgFile)
		}
		ctx.Log.Info("successfully parsed %s file", repoCfgFile)
		// It's possible we've already set defaultAutoDiscoverMode
		// from the config file while checking whether we can skip
		// cloning. We still need to set it here in the case that
		// we were not able to check whether we can skip cloning
		// and thus were not able to previously fetch the repo
		// config.
		if repoCfg.AutoDiscover != nil {
			defaultAutoDiscoverMode = repoCfg.AutoDiscover.Mode
		}
	}

	moduleInfo, err := FindModuleProjects(repoDir, p.AutoDetectModuleFiles)
	if err != nil {
		ctx.Log.Warn("error(s) loading project module dependencies: %s", err)
	}
	ctx.Log.Debug("moduleInfo for '%s' (matching '%s') = %v", repoDir, p.AutoDetectModuleFiles, moduleInfo)

	automerge := p.EnableAutoMerge
	parallelApply := p.EnableParallelApply
	parallelPlan := p.EnableParallelPlan
	abortOnExcecutionOrderFail := DefaultAbortOnExcecutionOrderFail
	if hasRepoCfg {
		if repoCfg.Automerge != nil {
			automerge = *repoCfg.Automerge
		}
		if repoCfg.ParallelApply != nil {
			parallelApply = *repoCfg.ParallelApply
		}
		if repoCfg.ParallelPlan != nil {
			parallelPlan = *repoCfg.ParallelPlan
		}
		abortOnExcecutionOrderFail = repoCfg.AbortOnExcecutionOrderFail
	}

	if len(repoCfg.Projects) > 0 {
		matchingProjects, err := p.ProjectFinder.DetermineProjectsViaConfig(ctx.Log, modifiedFiles, repoCfg, repoDir, moduleInfo)
		if err != nil {
			return nil, err
		}
		ctx.Log.Info("%d projects are to be planned based on their when_modified config", len(matchingProjects))

		for _, mp := range matchingProjects {
			ctx.Log.Debug("determining config for project at dir: '%s' workspace: '%s'", mp.Dir, mp.Workspace)
			mergedCfg := p.GlobalCfg.MergeProjectCfg(ctx.Log, ctx.Pull.BaseRepo.ID(), mp, repoCfg)

			projCtxs = append(projCtxs,
				p.ProjectCommandContextBuilder.BuildProjectContext(
					ctx,
					cmdName,
					subCmdName,
					mergedCfg,
					commentFlags,
					repoDir,
					automerge,
					parallelApply,
					parallelPlan,
					verbose,
					abortOnExcecutionOrderFail,
					p.TerraformExecutor,
				)...)
		}
	}

	if repoCfg.AutoDiscoverEnabled(defaultAutoDiscoverMode) {
		// If there is no config file or it specified no projects, then we'll plan each project that
		// our algorithm determines was modified.
		if hasRepoCfg {
			if len(repoCfg.Projects) == 0 {
				ctx.Log.Info("no projects are defined in %s. Will resume automatic detection", repoCfgFile)
			} else {
				ctx.Log.Info("automatic project discovery enabled. Will resume automatic detection")
			}
		} else {
			ctx.Log.Info("found no %s file", repoCfgFile)
		}
		// build a module index for projects that are explicitly included
		allModifiedProjects := p.ProjectFinder.DetermineProjects(
			ctx.Log, modifiedFiles, ctx.Pull.BaseRepo.FullName, repoDir, p.AutoplanFileList, moduleInfo)
		// If a project is already manually configured with the same dir as a discovered project, the manually configured
		// project should take precedence
		modifiedProjects := make([]models.Project, 0)
		configuredProjDirs := make(map[string]bool)
		// We compare against all configured projects instead of projects which match the modified files in case a
		// project is being specifically excluded (ex: when_modified doesn't match). We don't want to accidentally
		// "discover" it again.
		for _, configProj := range repoCfg.Projects {
			// Clean the path to make sure ./rel_path is equivalent to rel_path, etc
			configuredProjDirs[filepath.Clean(configProj.Dir)] = true
		}
		for _, mp := range allModifiedProjects {
			_, dirExists := configuredProjDirs[filepath.Clean(mp.Path)]
			if !dirExists {
				modifiedProjects = append(modifiedProjects, mp)
			}
		}
		ctx.Log.Info("automatically determined that there were %d additional projects modified in this pull request: %s",
			len(modifiedProjects), modifiedProjects)
		for _, mp := range modifiedProjects {
			ctx.Log.Debug("determining config for project at dir: '%s'", mp.Path)
			absProjectDir := filepath.Join(repoDir, mp.Path)
			pWorkspace, err := p.ProjectFinder.DetermineWorkspaceFromHCL(ctx.Log, absProjectDir)
			if err != nil {
				return nil, errors.Wrapf(err, "looking for Terraform Cloud workspace from configuration %s", absProjectDir)
			}

			pCfg := p.GlobalCfg.DefaultProjCfg(ctx.Log, ctx.Pull.BaseRepo.ID(), mp.Path, pWorkspace)

			projCtxs = append(projCtxs,
				p.ProjectCommandContextBuilder.BuildProjectContext(
					ctx,
					cmdName,
					subCmdName,
					pCfg,
					commentFlags,
					repoDir,
					automerge,
					parallelApply,
					parallelPlan,
					verbose,
					abortOnExcecutionOrderFail,
					p.TerraformExecutor,
				)...)
		}
	}

	sort.Slice(projCtxs, func(i, j int) bool {
		return projCtxs[i].ExecutionOrderGroup < projCtxs[j].ExecutionOrderGroup
	})

	// Filter projects to only include ones the user is authorized for
	projCtxs = slices.DeleteFunc(projCtxs, func(projCtx command.ProjectContext) bool {
		if projCtx.TeamAllowlistChecker == nil || !projCtx.TeamAllowlistChecker.HasRules() {
			// allowlist restriction is not enabled
			return false
		}
		ctx := models.TeamAllowlistCheckerContext{
			BaseRepo:           projCtx.BaseRepo,
			CommandName:        projCtx.CommandName.String(),
			EscapedCommentArgs: projCtx.EscapedCommentArgs,
			HeadRepo:           projCtx.HeadRepo,
			Log:                projCtx.Log,
			Pull:               projCtx.Pull,
			ProjectName:        projCtx.ProjectName,
			RepoDir:            repoDir,
			RepoRelDir:         projCtx.RepoRelDir,
			User:               projCtx.User,
			Verbose:            projCtx.Verbose,
			Workspace:          projCtx.Workspace,
			API:                false,
		}
		return !projCtx.TeamAllowlistChecker.IsCommandAllowedForAnyTeam(ctx, projCtx.User.Teams, projCtx.CommandName.String())
	})

	return projCtxs, nil
}

// buildProjectPlanCommand builds a plan context for a single project.
// cmd must be for only one project.
func (p *DefaultProjectCommandBuilder) buildProjectPlanCommand(ctx *command.Context, cmd *CommentCommand) ([]command.ProjectContext, error) {
	workspace := DefaultWorkspace
	if cmd.Workspace != "" {
		workspace = cmd.Workspace
	}

	var pcc []command.ProjectContext

	ctx.Log.Debug("building plan command")
	unlockFn, err := p.WorkingDirLocker.TryLock(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, workspace, DefaultRepoRelDir)
	if err != nil {
		return pcc, err
	}
	defer unlockFn()

	ctx.Log.Debug("cloning repository")
	_, _, err = p.WorkingDir.Clone(ctx.Log, ctx.HeadRepo, ctx.Pull, DefaultWorkspace)
	if err != nil {
		return pcc, err
	}

	// use the default repository workspace because it is the only one guaranteed to have an atlantis.yaml,
	// other workspaces will not have the file if they are using pre_workflow_hooks to generate it dynamically
	defaultRepoDir, err := p.WorkingDir.GetWorkingDir(ctx.Pull.BaseRepo, ctx.Pull, DefaultWorkspace)
	if err != nil {
		return pcc, err
	}

	if p.RestrictFileList {
		ctx.Log.Debug("'restrict-file-list' option is set, checking modified files")
		modifiedFiles, err := p.VCSClient.GetModifiedFiles(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull)
		if err != nil {
			return nil, err
		}

		if p.IncludeGitUntrackedFiles {
			ctx.Log.Debug(("'include-git-untracked-files' option is set, getting untracked files"))
			untrackedFiles, err := p.WorkingDir.GetGitUntrackedFiles(ctx.Log, ctx.HeadRepo, ctx.Pull, workspace)
			if err != nil {
				return nil, err
			}
			modifiedFiles = append(modifiedFiles, untrackedFiles...)
		}

		ctx.Log.Debug("%d files were modified in this pull request. Modified files: %v", len(modifiedFiles), modifiedFiles)

		if cmd.RepoRelDir != "" {
			ctx.Log.Debug("Command directory specified: %s", cmd.RepoRelDir)
			foundDir := false

			for _, f := range modifiedFiles {
				if filepath.Dir(f) == cmd.RepoRelDir {
					foundDir = true
				}
			}

			if !foundDir {
				return pcc, fmt.Errorf("the dir \"%s\" is not in the plan list of this pull request", cmd.RepoRelDir)
			}
		}

		if cmd.ProjectName != "" {
			ctx.Log.Debug("Command project name specified: %s", cmd.ProjectName)
			var notFoundFiles = []string{}
			var repoConfig valid.RepoCfg

			repoConfig, err = p.ParserValidator.ParseRepoCfg(defaultRepoDir, p.GlobalCfg, ctx.Pull.BaseRepo.ID(), ctx.Pull.BaseBranch)
			if err != nil {
				return pcc, err
			}
			repoCfgProjects := repoConfig.FindProjectsByName(cmd.ProjectName)

			for _, f := range modifiedFiles {
				foundDir := false

				for _, p := range repoCfgProjects {
					if filepath.Dir(f) == p.Dir {
						foundDir = true
					}
				}

				if !foundDir {
					notFoundFiles = append(notFoundFiles, filepath.Dir(f))
				}
			}

			if len(notFoundFiles) > 0 {
				return pcc, fmt.Errorf("the following directories are present in the pull request but not in the requested project:\n%s", strings.Join(notFoundFiles, "\n"))
			}
		}
	}

	if DefaultWorkspace != workspace {
		ctx.Log.Debug("cloning repository with workspace %s", workspace)
		_, _, err = p.WorkingDir.Clone(ctx.Log, ctx.HeadRepo, ctx.Pull, workspace)
		if err != nil {
			return pcc, err
		}
	}

	repoRelDir := DefaultRepoRelDir
	if cmd.RepoRelDir != "" {
		repoRelDir = cmd.RepoRelDir
	}

	return p.buildProjectCommandCtx(
		ctx,
		command.Plan,
		"",
		cmd.ProjectName,
		cmd.Flags,
		defaultRepoDir,
		repoRelDir,
		workspace,
		cmd.Verbose,
	)
}

// getCfg returns the atlantis.yaml config (if it exists) for this project. If
// there is no config, then projectCfg and repoCfg will be nil.
func (p *DefaultProjectCommandBuilder) getCfg(ctx *command.Context, projectName string, dir string, workspace string, repoDir string) (projectsCfg []valid.Project, repoCfg *valid.RepoCfg, err error) {
	repoCfgFile := p.GlobalCfg.RepoConfigFile(ctx.Pull.BaseRepo.ID())
	hasRepoCfg, err := p.ParserValidator.HasRepoCfg(repoDir, repoCfgFile)
	if err != nil {
		err = errors.Wrapf(err, "looking for '%s' file in '%s'", repoCfgFile, repoDir)
		return
	}
	if !hasRepoCfg {
		if projectName != "" {
			err = fmt.Errorf("cannot specify a project name unless an %s file exists to configure projects", repoCfgFile)
			return
		}
		return
	}

	var repoConfig valid.RepoCfg
	repoConfig, err = p.ParserValidator.ParseRepoCfg(repoDir, p.GlobalCfg, ctx.Pull.BaseRepo.ID(), ctx.Pull.BaseBranch)
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
			if p.SilenceNoProjects && len(repoConfig.Projects) > 0 {
				ctx.Log.Debug("no project with name '%s' found but silencing the error", projectName)
			} else {
				err = fmt.Errorf("no project with name '%s' is defined in '%s'", projectName, repoCfgFile)
			}
			return
		}
		return
	}

	projCfgs := repoCfg.FindProjectsByDirWorkspace(dir, workspace)
	if len(projCfgs) == 0 {
		return
	}
	if len(projCfgs) > 1 {
		err = fmt.Errorf("must specify project name: more than one project defined in '%s' matched dir: '%s' workspace: '%s'", repoCfgFile, dir, workspace)
		return
	}
	projectsCfg = projCfgs
	return
}

// buildAllProjectCommandsByPlan builds contexts for a command for every project that has
// pending plans in this ctx.
func (p *DefaultProjectCommandBuilder) buildAllProjectCommandsByPlan(ctx *command.Context, commentCmd *CommentCommand) ([]command.ProjectContext, error) {
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
		commentCmds, err := p.buildProjectCommandCtx(ctx, commentCmd.CommandName(), commentCmd.SubName, plan.ProjectName, commentCmd.Flags, defaultRepoDir, plan.RepoRelDir, plan.Workspace, commentCmd.Verbose)
		if err != nil {
			return nil, errors.Wrapf(err, "building command for dir '%s'", plan.RepoRelDir)
		}
		cmds = append(cmds, commentCmds...)
	}

	sort.Slice(cmds, func(i, j int) bool {
		return cmds[i].ExecutionOrderGroup < cmds[j].ExecutionOrderGroup
	})

	return cmds, nil
}

// buildProjectCommand builds an command for the single project
// identified by cmd except plan.
func (p *DefaultProjectCommandBuilder) buildProjectCommand(ctx *command.Context, cmd *CommentCommand) ([]command.ProjectContext, error) {
	workspace := DefaultWorkspace
	if cmd.Workspace != "" {
		workspace = cmd.Workspace
	}

	var projCtx []command.ProjectContext
	unlockFn, err := p.WorkingDirLocker.TryLock(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, workspace, DefaultRepoRelDir)
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
		cmd.Name,
		cmd.SubName,
		cmd.ProjectName,
		cmd.Flags,
		repoDir,
		repoRelDir,
		workspace,
		cmd.Verbose,
	)
}

// buildProjectCommandCtx builds a context for a single or several projects identified
// by the parameters.
func (p *DefaultProjectCommandBuilder) buildProjectCommandCtx(ctx *command.Context,
	cmd command.Name,
	subCmd string,
	projectName string,
	commentFlags []string,
	repoDir string,
	repoRelDir string,
	workspace string,
	verbose bool) ([]command.ProjectContext, error) {

	matchingProjects, repoCfgPtr, err := p.getCfg(ctx, projectName, repoRelDir, workspace, repoDir)
	if err != nil {
		return []command.ProjectContext{}, err
	}
	var projCtxs []command.ProjectContext
	var projCfg valid.MergedProjectCfg
	automerge := p.EnableAutoMerge
	parallelApply := p.EnableParallelApply
	parallelPlan := p.EnableParallelPlan
	abortOnExcecutionOrderFail := DefaultAbortOnExcecutionOrderFail
	if repoCfgPtr != nil {
		if repoCfgPtr.Automerge != nil {
			automerge = *repoCfgPtr.Automerge
		}
		if repoCfgPtr.ParallelApply != nil {
			parallelApply = *repoCfgPtr.ParallelApply
		}
		if repoCfgPtr.ParallelPlan != nil {
			parallelPlan = *repoCfgPtr.ParallelPlan
		}
		abortOnExcecutionOrderFail = repoCfgPtr.AbortOnExcecutionOrderFail
	}

	if len(matchingProjects) > 0 {
		// Override any dir/workspace defined on the comment with what was
		// defined in config. This shouldn't matter since we don't allow comments
		// with both project name and dir/workspace.
		repoRelDir = projCfg.RepoRelDir
		workspace = projCfg.Workspace
		for _, mp := range matchingProjects {
			ctx.Log.Debug("Merging config for project at dir: '%s' workspace: '%s'", mp.Dir, mp.Workspace)
			projCfg = p.GlobalCfg.MergeProjectCfg(ctx.Log, ctx.Pull.BaseRepo.ID(), mp, *repoCfgPtr)

			projCtxs = append(projCtxs,
				p.ProjectCommandContextBuilder.BuildProjectContext(
					ctx,
					cmd,
					subCmd,
					projCfg,
					commentFlags,
					repoDir,
					automerge,
					parallelApply,
					parallelPlan,
					verbose,
					abortOnExcecutionOrderFail,
					p.TerraformExecutor,
				)...)
		}
	} else {
		// Ignore the project if silenced with projects set in the repo config
		if p.SilenceNoProjects && repoCfgPtr != nil && len(repoCfgPtr.Projects) > 0 {
			ctx.Log.Debug("silencing is in effect, project will be ignored")
			return []command.ProjectContext{}, nil
		}

		projCfg = p.GlobalCfg.DefaultProjCfg(ctx.Log, ctx.Pull.BaseRepo.ID(), repoRelDir, workspace)
		projCtxs = append(projCtxs,
			p.ProjectCommandContextBuilder.BuildProjectContext(
				ctx,
				cmd,
				subCmd,
				projCfg,
				commentFlags,
				repoDir,
				automerge,
				parallelApply,
				parallelPlan,
				verbose,
				abortOnExcecutionOrderFail,
				p.TerraformExecutor,
			)...)
	}

	if err := p.validateWorkspaceAllowed(repoCfgPtr, repoRelDir, workspace); err != nil {
		return []command.ProjectContext{}, err
	}

	// Filter projects to only include ones the user is authorized for
	projCtxs = slices.DeleteFunc(projCtxs, func(projCtx command.ProjectContext) bool {
		if projCtx.TeamAllowlistChecker == nil || !projCtx.TeamAllowlistChecker.HasRules() {
			// allowlist restriction is not enabled
			return false
		}
		ctx := models.TeamAllowlistCheckerContext{
			BaseRepo:           projCtx.BaseRepo,
			CommandName:        projCtx.CommandName.String(),
			EscapedCommentArgs: projCtx.EscapedCommentArgs,
			HeadRepo:           projCtx.HeadRepo,
			Log:                projCtx.Log,
			Pull:               projCtx.Pull,
			ProjectName:        projCtx.ProjectName,
			RepoDir:            repoDir,
			RepoRelDir:         projCtx.RepoRelDir,
			User:               projCtx.User,
			Verbose:            projCtx.Verbose,
			Workspace:          projCtx.Workspace,
			API:                false,
		}
		return !projCtx.TeamAllowlistChecker.IsCommandAllowedForAnyTeam(ctx, projCtx.User.Teams, projCtx.CommandName.String())
	})

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
