// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"fmt"
	"path/filepath"

	"github.com/runatlantis/atlantis/server/core/config"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/command"
)

type undivergedProjectImpactMode int

const (
	undivergedProjectImpactModeNone undivergedProjectImpactMode = iota
	undivergedProjectImpactModeConfigured
	undivergedProjectImpactModeAutoDiscovered
)

type undivergedProjectImpactTarget struct {
	mode    undivergedProjectImpactMode
	repoCfg valid.RepoCfg
}

type undivergedProjectImpactResolver struct {
	ParserValidator       *config.ParserValidator
	ProjectFinder         ProjectFinder
	GlobalCfg             valid.GlobalCfg
	AutoDetectModuleFiles string
	AutoplanFileList      string
	AutoDiscoverMode      string
}

// NewUndivergedProjectImpactResolver builds the resolver used by targeted
// undiverged checks to mirror Atlantis project selection.
func NewUndivergedProjectImpactResolver(
	parserValidator *config.ParserValidator,
	projectFinder ProjectFinder,
	globalCfg valid.GlobalCfg,
	autoDetectModuleFiles string,
	autoplanFileList string,
	autoDiscoverMode string,
) *undivergedProjectImpactResolver {
	return &undivergedProjectImpactResolver{
		ParserValidator:       parserValidator,
		ProjectFinder:         projectFinder,
		GlobalCfg:             globalCfg,
		AutoDetectModuleFiles: autoDetectModuleFiles,
		AutoplanFileList:      autoplanFileList,
		AutoDiscoverMode:      autoDiscoverMode,
	}
}

func (r *undivergedProjectImpactResolver) HasUndivergedImpact(
	ctx command.ProjectContext,
	repoDir string,
	workingDir WorkingDir,
) (handled bool, impacted bool, err error) {
	target, err := r.resolveTarget(ctx, repoDir)
	if err != nil {
		return false, false, err
	}
	if target.mode == undivergedProjectImpactModeNone {
		return false, false, nil
	}
	if target.mode == undivergedProjectImpactModeConfigured && len(ctx.AutoplanWhenModified) == 0 {
		// Preserve the previous empty-pattern behavior for configured projects:
		// HasDiverged treats an empty list as a full divergence check.
		return false, false, nil
	}

	divergedFiles, err := workingDir.GetDivergedFiles(ctx.Log, repoDir, ctx.Pull)
	if err != nil {
		return true, false, err
	}
	if len(divergedFiles) == 0 {
		return true, false, nil
	}

	impacted, err = r.impactedByModifiedFiles(ctx, repoDir, target, divergedFiles)
	if err != nil {
		return true, false, err
	}

	return true, impacted, nil
}

func (r *undivergedProjectImpactResolver) resolveTarget(ctx command.ProjectContext, repoDir string) (undivergedProjectImpactTarget, error) {
	if r == nil {
		return undivergedProjectImpactTarget{}, nil
	}

	repoCfgFile := r.GlobalCfg.RepoConfigFile(ctx.Pull.BaseRepo.ID())
	hasRepoCfg, err := r.ParserValidator.HasRepoCfg(repoDir, repoCfgFile)
	if err != nil {
		return undivergedProjectImpactTarget{}, fmt.Errorf("looking for %q in %q: %w", repoCfgFile, repoDir, err)
	}

	var repoCfg valid.RepoCfg
	if hasRepoCfg {
		repoCfg, err = r.ParserValidator.ParseRepoCfg(repoDir, r.GlobalCfg, ctx.Pull.BaseRepo.ID(), ctx.Pull.BaseBranch)
		if err != nil {
			return undivergedProjectImpactTarget{}, fmt.Errorf("parsing %s: %w", repoCfgFile, err)
		}
	} else if ctx.RepoConfigVersion > 0 {
		// The command context came from repo config, but this workspace clone
		// does not have the config file. That can happen with pre-workflow
		// generated config in non-default workspaces. Fall back to the existing
		// HasDiverged path using the context's stored when_modified patterns.
		return undivergedProjectImpactTarget{
			mode: undivergedProjectImpactModeNone,
		}, nil
	}

	for _, project := range repoCfg.Projects {
		if matchesConfiguredProjectContext(project, ctx) {
			return undivergedProjectImpactTarget{
				mode:    undivergedProjectImpactModeConfigured,
				repoCfg: repoCfg,
			}, nil
		}
	}

	if r.shouldUseAutoDiscoveredTargeting(ctx, repoCfg) {
		return undivergedProjectImpactTarget{
			mode:    undivergedProjectImpactModeAutoDiscovered,
			repoCfg: repoCfg,
		}, nil
	}

	return undivergedProjectImpactTarget{
		mode:    undivergedProjectImpactModeNone,
		repoCfg: repoCfg,
	}, nil
}

func (r *undivergedProjectImpactResolver) impactedByModifiedFiles(
	ctx command.ProjectContext,
	repoDir string,
	target undivergedProjectImpactTarget,
	modifiedFiles []string,
) (bool, error) {
	if target.mode == undivergedProjectImpactModeNone {
		return false, nil
	}

	moduleInfo, err := FindModuleProjects(repoDir, r.AutoDetectModuleFiles)
	if err != nil {
		return false, fmt.Errorf("loading project module dependencies: %w", err)
	}

	switch target.mode {
	case undivergedProjectImpactModeConfigured:
		return r.configuredProjectImpacted(ctx, repoDir, target.repoCfg, modifiedFiles, moduleInfo)
	case undivergedProjectImpactModeAutoDiscovered:
		return r.autoDiscoveredProjectImpacted(ctx, repoDir, target.repoCfg, modifiedFiles, moduleInfo)
	default:
		return false, nil
	}
}

func (r *undivergedProjectImpactResolver) configuredProjectImpacted(
	ctx command.ProjectContext,
	repoDir string,
	repoCfg valid.RepoCfg,
	modifiedFiles []string,
	moduleInfo ModuleProjects,
) (bool, error) {
	projects, err := r.ProjectFinder.DetermineProjectsViaConfig(ctx.Log, modifiedFiles, repoCfg, repoDir, moduleInfo)
	if err != nil {
		return false, err
	}

	for _, project := range projects {
		if matchesConfiguredProjectContext(project, ctx) {
			return true, nil
		}
	}

	return false, nil
}

func (r *undivergedProjectImpactResolver) autoDiscoveredProjectImpacted(
	ctx command.ProjectContext,
	repoDir string,
	repoCfg valid.RepoCfg,
	modifiedFiles []string,
	moduleInfo ModuleProjects,
) (bool, error) {
	modifiedProjects := r.ProjectFinder.DetermineProjects(
		ctx.Log,
		modifiedFiles,
		ctx.Pull.BaseRepo.FullName,
		repoDir,
		r.AutoplanFileList,
		moduleInfo,
	)

	configuredProjDirs := make(map[string]bool)
	for _, configProj := range repoCfg.Projects {
		configuredProjDirs[filepath.Clean(configProj.Dir)] = true
	}

	currentDir := filepath.Clean(ctx.RepoRelDir)
	for _, project := range modifiedProjects {
		projectDir := filepath.Clean(project.Path)
		if r.isAutoDiscoverPathIgnored(ctx, repoCfg, projectDir) {
			continue
		}
		if configuredProjDirs[projectDir] {
			continue
		}
		if projectDir == currentDir {
			return true, nil
		}
	}

	return false, nil
}

func (r *undivergedProjectImpactResolver) shouldUseAutoDiscoveredTargeting(ctx command.ProjectContext, repoCfg valid.RepoCfg) bool {
	if ctx.ProjectName != "" {
		return false
	}

	if !r.autoDiscoverModeEnabled(ctx, repoCfg) {
		return false
	}

	currentDir := filepath.Clean(ctx.RepoRelDir)
	if len(repoCfg.FindProjectsByDir(currentDir)) > 0 {
		return false
	}

	return !r.isAutoDiscoverPathIgnored(ctx, repoCfg, currentDir)
}

func (r *undivergedProjectImpactResolver) autoDiscoverModeEnabled(ctx command.ProjectContext, repoCfg valid.RepoCfg) bool {
	if global := r.GlobalCfg.RepoAutoDiscoverCfg(ctx.Pull.BaseRepo.ID()); global != nil {
		return repoCfg.AutoDiscoverEnabled(global.Mode)
	}
	if repoCfg.AutoDiscover != nil {
		return repoCfg.AutoDiscoverEnabled(repoCfg.AutoDiscover.Mode)
	}

	defaultAutoDiscoverMode := valid.AutoDiscoverMode(r.AutoDiscoverMode)
	if defaultAutoDiscoverMode == "" {
		defaultAutoDiscoverMode = valid.AutoDiscoverAutoMode
	}

	return repoCfg.AutoDiscoverEnabled(defaultAutoDiscoverMode)
}

func (r *undivergedProjectImpactResolver) isAutoDiscoverPathIgnored(ctx command.ProjectContext, repoCfg valid.RepoCfg, path string) bool {
	fromGlobalAutoDiscover := r.GlobalCfg.RepoAutoDiscoverCfg(ctx.Pull.BaseRepo.ID())
	if fromGlobalAutoDiscover != nil && fromGlobalAutoDiscover.IgnorePaths != nil {
		return fromGlobalAutoDiscover.IsPathIgnored(path)
	}
	if repoCfg.AutoDiscover != nil {
		return repoCfg.AutoDiscover.IsPathIgnored(path)
	}

	return false
}

func matchesConfiguredProjectContext(project valid.Project, ctx command.ProjectContext) bool {
	return filepath.Clean(project.Dir) == filepath.Clean(ctx.RepoRelDir) &&
		project.Workspace == ctx.Workspace &&
		project.GetName() == ctx.ProjectName
}
