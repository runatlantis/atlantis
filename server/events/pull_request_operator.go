package events

import (
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

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_pull_request_operator.go PullRequestOperator

type PullRequestOperator interface {
	Autoplan(ctx *CommandContext) CommandResponse
	PlanViaComment(ctx *CommandContext) CommandResponse
	ApplyViaComment(ctx *CommandContext) CommandResponse
}

type DefaultPullRequestOperator struct {
	TerraformExecutor TerraformExec
	DefaultTFVersion  *version.Version
	ParserValidator   *yaml.ParserValidator
	ProjectFinder     ProjectFinder
	VCSClient         vcs.ClientProxy
	Workspace         AtlantisWorkspace
	ProjectOperator   ProjectOperator
}

type TerraformExec interface {
	RunCommandWithVersion(log *logging.SimpleLogger, path string, args []string, v *version.Version, workspace string) (string, error)
}

func (p *DefaultPullRequestOperator) Autoplan(ctx *CommandContext) CommandResponse {
	// check out repo to parse atlantis.yaml
	// this will check out the repo to a * dir
	repoDir, err := p.Workspace.Clone(ctx.Log, ctx.BaseRepo, ctx.HeadRepo, ctx.Pull, ctx.Command.Workspace)
	if err != nil {
		return CommandResponse{Error: err}
	}

	// Parse config file if it exists.
	ctx.Log.Debug("parsing config file")
	config, err := p.ParserValidator.ReadConfig(repoDir)
	if err != nil && !os.IsNotExist(err) {
		return CommandResponse{Error: err}
	}
	noAtlantisYAML := os.IsNotExist(err)
	if noAtlantisYAML {
		ctx.Log.Info("found no %s file", yaml.AtlantisYAMLFilename)
	} else {
		ctx.Log.Info("successfully parsed %s file", yaml.AtlantisYAMLFilename)
	}

	// We'll need the list of modified files.
	modifiedFiles, err := p.VCSClient.GetModifiedFiles(ctx.BaseRepo, ctx.Pull)
	if err != nil {
		return CommandResponse{Error: err}
	}
	ctx.Log.Debug("%d files were modified in this pull request", len(modifiedFiles))

	// Prepare the project contexts so the ProjectOperator can execute.
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
				RepoRelPath:   mp.Path,
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
			return CommandResponse{Error: err}
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
				RepoRelPath:   mp.Dir,
				ProjectConfig: &mp,
				GlobalConfig:  &config,
			})
		}
	}

	// Execute the operations.
	var results []ProjectResult
	for _, pCtx := range projCtxs {
		res := p.ProjectOperator.Plan(pCtx, nil)
		res.Path = pCtx.RepoRelPath
		res.Workspace = pCtx.Workspace
		results = append(results, res)
	}
	return CommandResponse{ProjectResults: results}
}

func (p *DefaultPullRequestOperator) PlanViaComment(ctx *CommandContext) CommandResponse {
	repoDir, err := p.Workspace.Clone(ctx.Log, ctx.BaseRepo, ctx.HeadRepo, ctx.Pull, ctx.Command.Workspace)
	if err != nil {
		return CommandResponse{Error: err}
	}

	var projCfg *valid.Project
	var globalCfg *valid.Spec

	// Parse config file if it exists.
	config, err := p.ParserValidator.ReadConfig(repoDir)
	if err != nil && !os.IsNotExist(err) {
		return CommandResponse{Error: err}
	}
	if !os.IsNotExist(err) {
		projCfg = config.FindProject(ctx.Command.Dir, ctx.Command.Workspace)
		globalCfg = &config
	}

	projCtx := models.ProjectCommandContext{
		BaseRepo:      ctx.BaseRepo,
		HeadRepo:      ctx.HeadRepo,
		Pull:          ctx.Pull,
		User:          ctx.User,
		Log:           ctx.Log,
		CommentArgs:   ctx.Command.Flags,
		Workspace:     ctx.Command.Workspace,
		RepoRelPath:   ctx.Command.Dir,
		ProjectConfig: projCfg,
		GlobalConfig:  globalCfg,
	}
	projAbsPath := filepath.Join(repoDir, ctx.Command.Dir)
	res := p.ProjectOperator.Plan(projCtx, &projAbsPath)
	res.Workspace = projCtx.Workspace
	res.Path = projCtx.RepoRelPath
	return CommandResponse{
		ProjectResults: []ProjectResult{
			res,
		},
	}
}

func (p *DefaultPullRequestOperator) ApplyViaComment(ctx *CommandContext) CommandResponse {
	repoDir, err := p.Workspace.GetWorkspace(ctx.BaseRepo, ctx.Pull, ctx.Command.Workspace)
	if err != nil {
		return CommandResponse{Failure: "No workspace found. Did you run plan?"}
	}

	// todo: can deduplicate this between PlanViaComment
	var projCfg *valid.Project
	var globalCfg *valid.Spec

	// Parse config file if it exists.
	config, err := p.ParserValidator.ReadConfig(repoDir)
	if err != nil && !os.IsNotExist(err) {
		return CommandResponse{Error: err}
	}
	if !os.IsNotExist(err) {
		projCfg = config.FindProject(ctx.Command.Dir, ctx.Command.Workspace)
		globalCfg = &config
	}

	projCtx := models.ProjectCommandContext{
		BaseRepo:      ctx.BaseRepo,
		HeadRepo:      ctx.HeadRepo,
		Pull:          ctx.Pull,
		User:          ctx.User,
		Log:           ctx.Log,
		CommentArgs:   ctx.Command.Flags,
		Workspace:     ctx.Command.Workspace,
		RepoRelPath:   ctx.Command.Dir,
		ProjectConfig: projCfg,
		GlobalConfig:  globalCfg,
	}
	res := p.ProjectOperator.Apply(projCtx, filepath.Join(repoDir, ctx.Command.Dir))
	res.Workspace = projCtx.Workspace
	res.Path = projCtx.RepoRelPath
	return CommandResponse{
		ProjectResults: []ProjectResult{
			res,
		},
	}
}

// matchingProjects returns the list of projects whose WhenModified fields match
// any of the modifiedFiles.
func (p *DefaultPullRequestOperator) matchingProjects(log *logging.SimpleLogger, modifiedFiles []string, config valid.Spec) ([]valid.Project, error) {
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
