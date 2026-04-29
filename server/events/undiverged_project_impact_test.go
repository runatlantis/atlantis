// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/config"
	"github.com/runatlantis/atlantis/server/core/config/raw"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

const defaultAutoplanFileList = "**/*.tf,**/*.tfvars,**/*.tfvars.json,**/terragrunt.hcl,**/.terraform.lock.hcl"

func TestUndivergedProjectImpactResolver_ConfiguredProjectModuleChange(t *testing.T) {
	repoDir := configuredProjectRepo(t)
	resolver := newTestUndivergedProjectImpactResolver("**/*.tf", defaultAutoplanFileList, "auto")
	ctx := newTestUndivergedProjectContext(t, "project1")

	target, err := resolver.resolveTarget(ctx, repoDir)
	Ok(t, err)
	Equals(t, undivergedProjectImpactModeConfigured, target.mode)

	impacted, err := resolver.impactedByModifiedFiles(ctx, repoDir, target, []string{"modules/database/main.tf"})
	Ok(t, err)
	Equals(t, true, impacted)
}

func TestUndivergedProjectImpactResolver_ConfiguredProjectIgnoresUnrelatedChanges(t *testing.T) {
	repoDir := configuredProjectRepo(t)
	resolver := newTestUndivergedProjectImpactResolver("**/*.tf", defaultAutoplanFileList, "auto")
	ctx := newTestUndivergedProjectContext(t, "project1")

	target, err := resolver.resolveTarget(ctx, repoDir)
	Ok(t, err)
	Equals(t, undivergedProjectImpactModeConfigured, target.mode)

	impacted, err := resolver.impactedByModifiedFiles(ctx, repoDir, target, []string{"project2/main.tf"})
	Ok(t, err)
	Equals(t, false, impacted)
}

func TestUndivergedProjectImpactResolver_AutoDiscoveredProjectUsesModuleAutoplanning(t *testing.T) {
	repoDir := autoDiscoveredRepo(t)
	resolver := newTestUndivergedProjectImpactResolver("**/*.tf", defaultAutoplanFileList, "auto")
	ctx := newTestUndivergedProjectContext(t, "project1")

	target, err := resolver.resolveTarget(ctx, repoDir)
	Ok(t, err)
	Equals(t, undivergedProjectImpactModeAutoDiscovered, target.mode)

	impacted, err := resolver.impactedByModifiedFiles(ctx, repoDir, target, []string{"modules/database/main.tf"})
	Ok(t, err)
	Equals(t, true, impacted)
}

func TestUndivergedProjectImpactResolver_AutoDiscoveredProjectIgnoresUnrelatedChanges(t *testing.T) {
	repoDir := autoDiscoveredRepo(t)
	resolver := newTestUndivergedProjectImpactResolver("**/*.tf", defaultAutoplanFileList, "auto")
	ctx := newTestUndivergedProjectContext(t, "project1")

	target, err := resolver.resolveTarget(ctx, repoDir)
	Ok(t, err)
	Equals(t, undivergedProjectImpactModeAutoDiscovered, target.mode)

	impacted, err := resolver.impactedByModifiedFiles(ctx, repoDir, target, []string{"project2/main.tf"})
	Ok(t, err)
	Equals(t, false, impacted)
}

func TestUndivergedProjectImpactResolver_NoTargetWhenGlobalAutoDiscoverDisabled(t *testing.T) {
	repoDir := autoDiscoveredRepo(t)
	resolver := newTestUndivergedProjectImpactResolver("**/*.tf", defaultAutoplanFileList, "disabled")
	resolver.GlobalCfg.Repos[0].AutoDiscover = &valid.AutoDiscover{Mode: valid.AutoDiscoverDisabledMode}
	ctx := newTestUndivergedProjectContext(t, "project1")

	target, err := resolver.resolveTarget(ctx, repoDir)
	Ok(t, err)
	Equals(t, undivergedProjectImpactModeNone, target.mode)
}

func TestUndivergedProjectImpactResolver_NoTargetWhenRepoAutoDiscoverDisabled(t *testing.T) {
	repoDir := autoDiscoveredRepoWithAutoDiscoverMode(t, valid.AutoDiscoverDisabledMode)
	resolver := newTestUndivergedProjectImpactResolver("**/*.tf", defaultAutoplanFileList, "auto")
	ctx := newTestUndivergedProjectContext(t, "project1")

	target, err := resolver.resolveTarget(ctx, repoDir)
	Ok(t, err)
	Equals(t, undivergedProjectImpactModeNone, target.mode)
}

func TestUndivergedProjectImpactResolver_GlobalAutoDiscoverModeOverridesRepoMode(t *testing.T) {
	repoDir := autoDiscoveredRepoWithAutoDiscoverMode(t, valid.AutoDiscoverDisabledMode)
	resolver := newTestUndivergedProjectImpactResolver("**/*.tf", defaultAutoplanFileList, "auto")
	resolver.GlobalCfg.Repos[0].AutoDiscover = &valid.AutoDiscover{Mode: valid.AutoDiscoverEnabledMode}
	ctx := newTestUndivergedProjectContext(t, "project1")

	target, err := resolver.resolveTarget(ctx, repoDir)
	Ok(t, err)
	Equals(t, undivergedProjectImpactModeAutoDiscovered, target.mode)
}

func TestUndivergedProjectImpactResolver_InheritedGlobalAutoDiscoverModeOverridesRepoMode(t *testing.T) {
	repoDir := autoDiscoveredRepoWithAutoDiscoverMode(t, valid.AutoDiscoverDisabledMode)
	resolver := newTestUndivergedProjectImpactResolver("**/*.tf", defaultAutoplanFileList, "auto")
	resolver.GlobalCfg.Repos = []valid.Repo{
		{
			IDRegex:      regexp.MustCompile(".*"),
			AutoDiscover: &valid.AutoDiscover{Mode: valid.AutoDiscoverEnabledMode},
		},
		{
			IDRegex: regexp.MustCompile(".*"),
		},
	}
	ctx := newTestUndivergedProjectContext(t, "project1")

	target, err := resolver.resolveTarget(ctx, repoDir)
	Ok(t, err)
	Equals(t, undivergedProjectImpactModeAutoDiscovered, target.mode)
}

func TestUndivergedProjectImpactResolver_GlobalAutoDiscoverWithoutIgnorePathsUsesRepoIgnorePaths(t *testing.T) {
	repoDir := autoDiscoveredRepoWithIgnorePaths(t)
	resolver := newTestUndivergedProjectImpactResolver("**/*.tf", defaultAutoplanFileList, "auto")
	resolver.GlobalCfg.Repos[0].AutoDiscover = &valid.AutoDiscover{Mode: valid.AutoDiscoverEnabledMode}
	ctx := newTestUndivergedProjectContext(t, "project1")

	target, err := resolver.resolveTarget(ctx, repoDir)
	Ok(t, err)
	Equals(t, undivergedProjectImpactModeNone, target.mode)
}

func TestDefaultCommandRequirementHandler_TargetedUndivergedFailsForImpactedConfiguredProject(t *testing.T) {
	RegisterMockTestingT(t)

	repoDir := configuredProjectRepo(t)
	resolver := newTestUndivergedProjectImpactResolver("**/*.tf", defaultAutoplanFileList, "auto")
	workingDir := NewMockWorkingDir()
	When(workingDir.GetDivergedFiles(Any[logging.SimpleLogging](), Any[string](), Any[models.PullRequest]())).ThenReturn([]string{"modules/database/main.tf"}, nil)

	handler := &DefaultCommandRequirementHandler{
		WorkingDir:            workingDir,
		ProjectImpactResolver: resolver,
	}

	ctx := newTestUndivergedProjectContext(t, "project1")
	ctx.ApplyRequirements = []string{raw.UnDivergedRequirement}
	ctx.AutoplanWhenModified = raw.DefaultAutoPlanWhenModified

	failure, err := handler.ValidateApplyProject(repoDir, ctx)
	Ok(t, err)
	Equals(t, "Default branch must be rebased onto pull request before running apply.", failure)
}

func TestDefaultCommandRequirementHandler_TargetedUndivergedPassesForUnrelatedConfiguredChange(t *testing.T) {
	RegisterMockTestingT(t)

	repoDir := configuredProjectRepo(t)
	resolver := newTestUndivergedProjectImpactResolver("**/*.tf", defaultAutoplanFileList, "auto")
	workingDir := NewMockWorkingDir()
	When(workingDir.GetDivergedFiles(Any[logging.SimpleLogging](), Any[string](), Any[models.PullRequest]())).ThenReturn([]string{"project2/main.tf"}, nil)

	handler := &DefaultCommandRequirementHandler{
		WorkingDir:            workingDir,
		ProjectImpactResolver: resolver,
	}

	ctx := newTestUndivergedProjectContext(t, "project1")
	ctx.ApplyRequirements = []string{raw.UnDivergedRequirement}
	ctx.AutoplanWhenModified = raw.DefaultAutoPlanWhenModified

	failure, err := handler.ValidateApplyProject(repoDir, ctx)
	Ok(t, err)
	Equals(t, "", failure)
}

func TestDefaultCommandRequirementHandler_TargetedUndivergedSkipsImpactResolutionWhenNoDivergedFiles(t *testing.T) {
	RegisterMockTestingT(t)

	repoDir := configuredProjectRepo(t)
	writeTestFile(t, filepath.Join(repoDir, "project1", "main.tf"), "invalid terraform")
	resolver := newTestUndivergedProjectImpactResolver("**/*.tf", defaultAutoplanFileList, "auto")
	workingDir := NewMockWorkingDir()
	When(workingDir.GetDivergedFiles(Any[logging.SimpleLogging](), Any[string](), Any[models.PullRequest]())).ThenReturn([]string{}, nil)

	handler := &DefaultCommandRequirementHandler{
		WorkingDir:            workingDir,
		ProjectImpactResolver: resolver,
	}

	ctx := newTestUndivergedProjectContext(t, "project1")
	ctx.ApplyRequirements = []string{raw.UnDivergedRequirement}
	ctx.AutoplanWhenModified = raw.DefaultAutoPlanWhenModified

	failure, err := handler.ValidateApplyProject(repoDir, ctx)
	Ok(t, err)
	Equals(t, "", failure)
}

func TestDefaultCommandRequirementHandler_TargetedUndivergedFallsBackWhenGeneratedConfigMissing(t *testing.T) {
	RegisterMockTestingT(t)

	repoDir := autoDiscoveredRepo(t)
	resolver := newTestUndivergedProjectImpactResolver("**/*.tf", defaultAutoplanFileList, "auto")
	workingDir := NewMockWorkingDir()
	When(workingDir.HasDiverged(Any[logging.SimpleLogging](), Eq(repoDir), Eq("project1"), Eq([]string{"generated-only/**"}), Any[models.PullRequest]())).ThenReturn(true)

	handler := &DefaultCommandRequirementHandler{
		WorkingDir:            workingDir,
		ProjectImpactResolver: resolver,
	}

	ctx := newTestUndivergedProjectContext(t, "project1")
	ctx.ApplyRequirements = []string{raw.UnDivergedRequirement}
	ctx.AutoplanWhenModified = []string{"generated-only/**"}
	ctx.RepoConfigVersion = 3
	ctx.Workspace = "nondefault"

	failure, err := handler.ValidateApplyProject(repoDir, ctx)
	Ok(t, err)
	Equals(t, "Default branch must be rebased onto pull request before running apply.", failure)
}

func TestDefaultCommandRequirementHandler_TargetedUndivergedFallsBackForEmptyConfiguredWhenModified(t *testing.T) {
	RegisterMockTestingT(t)

	repoDir := configuredProjectRepoWithEmptyWhenModified(t)
	resolver := newTestUndivergedProjectImpactResolver("**/*.tf", defaultAutoplanFileList, "auto")
	workingDir := NewMockWorkingDir()
	When(workingDir.HasDiverged(Any[logging.SimpleLogging](), Eq(repoDir), Eq("project1"), Eq([]string{}), Any[models.PullRequest]())).ThenReturn(true)

	handler := &DefaultCommandRequirementHandler{
		WorkingDir:            workingDir,
		ProjectImpactResolver: resolver,
	}

	ctx := newTestUndivergedProjectContext(t, "project1")
	ctx.ApplyRequirements = []string{raw.UnDivergedRequirement}
	ctx.AutoplanWhenModified = []string{}
	ctx.RepoConfigVersion = 3

	failure, err := handler.ValidateApplyProject(repoDir, ctx)
	Ok(t, err)
	Equals(t, "Default branch must be rebased onto pull request before running apply.", failure)
}

func TestDefaultCommandRequirementHandler_TargetedUndivergedFallsBackToFullCheckOnResolverError(t *testing.T) {
	RegisterMockTestingT(t)

	repoDir := configuredProjectRepo(t)
	workingDir := NewMockWorkingDir()
	When(workingDir.HasDiverged(Any[logging.SimpleLogging](), Eq(repoDir), Eq("project1"), Eq([]string{"modules/database/**"}), Any[models.PullRequest]())).ThenReturn(true)
	When(workingDir.HasDiverged(Any[logging.SimpleLogging](), Eq(repoDir), Eq("project1"), Eq([]string(nil)), Any[models.PullRequest]())).ThenReturn(false)

	handler := &DefaultCommandRequirementHandler{
		WorkingDir:            workingDir,
		ProjectImpactResolver: stubUndivergedProjectImpactResolver{err: os.ErrInvalid},
	}

	ctx := newTestUndivergedProjectContext(t, "project1")
	ctx.ApplyRequirements = []string{raw.UnDivergedRequirement}
	ctx.AutoplanWhenModified = []string{"modules/database/**"}

	failure, err := handler.ValidateApplyProject(repoDir, ctx)
	Ok(t, err)
	Equals(t, "", failure)
}

type stubUndivergedProjectImpactResolver struct {
	handled  bool
	impacted bool
	err      error
}

func (s stubUndivergedProjectImpactResolver) HasUndivergedImpact(command.ProjectContext, string, WorkingDir) (bool, bool, error) {
	return s.handled, s.impacted, s.err
}

func configuredProjectRepo(t *testing.T) string {
	t.Helper()

	repoDir := DirStructure(t, map[string]any{
		"atlantis.yaml": nil,
		"project1": map[string]any{
			"main.tf": nil,
		},
		"modules": map[string]any{
			"database": map[string]any{
				"main.tf": nil,
			},
		},
	})

	writeTestFile(t, filepath.Join(repoDir, "atlantis.yaml"), `version: 3
projects:
- dir: project1
  workspace: default
`)
	writeTestFile(t, filepath.Join(repoDir, "project1", "main.tf"), `module "database" {
  source = "../modules/database"
}
`)
	writeTestFile(t, filepath.Join(repoDir, "modules", "database", "main.tf"), `output "name" {
  value = "database"
}
`)

	return repoDir
}

func configuredProjectRepoWithEmptyWhenModified(t *testing.T) string {
	t.Helper()

	repoDir := configuredProjectRepo(t)
	writeTestFile(t, filepath.Join(repoDir, "atlantis.yaml"), `version: 3
projects:
- dir: project1
  workspace: default
  autoplan:
    when_modified: []
`)

	return repoDir
}

func autoDiscoveredRepo(t *testing.T) string {
	t.Helper()

	repoDir := DirStructure(t, map[string]any{
		"project1": map[string]any{
			"main.tf": nil,
		},
		"project2": map[string]any{
			"main.tf": nil,
		},
		"modules": map[string]any{
			"database": map[string]any{
				"main.tf": nil,
			},
		},
	})

	writeTestFile(t, filepath.Join(repoDir, "project1", "main.tf"), `module "database" {
  source = "../modules/database"
}
`)
	writeTestFile(t, filepath.Join(repoDir, "project2", "main.tf"), `output "name" {
  value = "project2"
}
`)
	writeTestFile(t, filepath.Join(repoDir, "modules", "database", "main.tf"), `output "name" {
  value = "database"
}
`)

	return repoDir
}

func autoDiscoveredRepoWithAutoDiscoverMode(t *testing.T, mode valid.AutoDiscoverMode) string {
	t.Helper()

	repoDir := autoDiscoveredRepo(t)
	writeTestFile(t, filepath.Join(repoDir, "atlantis.yaml"), `version: 3
autodiscover:
  mode: `+string(mode)+`
`)

	return repoDir
}

func autoDiscoveredRepoWithIgnorePaths(t *testing.T) string {
	t.Helper()

	repoDir := autoDiscoveredRepo(t)
	writeTestFile(t, filepath.Join(repoDir, "atlantis.yaml"), `version: 3
autodiscover:
  mode: auto
  ignore_paths:
  - project1/**
`)

	return repoDir
}

func newTestUndivergedProjectImpactResolver(autoDetectModuleFiles string, autoplanFileList string, autoDiscoverMode string) *undivergedProjectImpactResolver {
	parserValidator := &config.ParserValidator{}
	globalCfg := valid.NewGlobalCfgFromArgs(valid.GlobalCfgArgs{})

	return NewUndivergedProjectImpactResolver(
		parserValidator,
		&DefaultProjectFinder{},
		globalCfg,
		autoDetectModuleFiles,
		autoplanFileList,
		autoDiscoverMode,
	)
}

func newTestUndivergedProjectContext(t *testing.T, repoRelDir string) command.ProjectContext {
	t.Helper()

	return command.ProjectContext{
		Log:        logging.NewNoopLogger(t),
		RepoRelDir: repoRelDir,
		Workspace:  DefaultWorkspace,
		Pull: models.PullRequest{
			BaseBranch: "main",
			BaseRepo: models.Repo{
				FullName: "owner/repo",
				VCSHost: models.VCSHost{
					Hostname: "github.com",
				},
			},
		},
	}
}

func writeTestFile(t *testing.T, path string, contents string) {
	t.Helper()

	Ok(t, os.WriteFile(path, []byte(contents), 0600))
}
