package events

import (
	"os"
	"path/filepath"

	"github.com/hashicorp/go-version"
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
	config, err := p.ParserValidator.ReadConfig(repoDir)
	if err != nil && !os.IsNotExist(err) {
		return CommandResponse{Error: err}
	}
	noAtlantisYAML := os.IsNotExist(err)

	// We'll need the list of modified files.
	modifiedFiles, err := p.VCSClient.GetModifiedFiles(ctx.BaseRepo, ctx.Pull)
	if err != nil {
		return CommandResponse{Error: err}
	}

	// Prepare the project contexts so the ProjectOperator can execute.
	var projCtxs []models.ProjectCommandContext

	// If there is no config file, then we try to plan for each project that
	// was modified in the pull request.
	if noAtlantisYAML {
		modifiedProjects := p.ProjectFinder.DetermineProjects(ctx.Log, modifiedFiles, ctx.BaseRepo.FullName, repoDir)
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
		matchingProjects := p.matchingProjects(modifiedFiles, config)
		for _, mp := range matchingProjects {
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
func (p *DefaultPullRequestOperator) matchingProjects(modifiedFiles []string, config valid.Spec) []valid.Project {
	//todo
	// match the modified files against the config
	// remember the modified_files paths are relative to the project paths
	return nil
}
