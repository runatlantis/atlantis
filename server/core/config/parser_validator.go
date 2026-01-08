// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	validation "github.com/go-ozzo/ozzo-validation"
	shlex "github.com/google/shlex"

	"github.com/runatlantis/atlantis/server/core/config/raw"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	yaml "go.yaml.in/yaml/v4"
)

// ParserValidator parses and validates server-side repo config files and
// repo-level atlantis.yaml files.
type ParserValidator struct{}

// HasRepoCfg returns true if there is a repo config (atlantis.yaml) file
// for the repo at absRepoDir.
// Returns an error if for some reason it can't read that directory.
func (p *ParserValidator) HasRepoCfg(absRepoDir, repoConfigFile string) (bool, error) {
	// Checks for a config file with an invalid extension (atlantis.yml)
	const invalidExtensionFilename = "atlantis.yml"
	_, err := os.Stat(p.repoCfgPath(absRepoDir, invalidExtensionFilename))
	if err == nil {
		return false, fmt.Errorf("found %q as config file; rename using the .yaml extension", invalidExtensionFilename)
	}

	_, err = os.Stat(p.repoCfgPath(absRepoDir, repoConfigFile))
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return err == nil, err
}

// ParseRepoCfg returns the parsed and validated atlantis.yaml config for the
// repo at absRepoDir.
// If there was no config file, it will return an os.IsNotExist(error).
func (p *ParserValidator) ParseRepoCfg(absRepoDir string, globalCfg valid.GlobalCfg, repoID string, branch string) (valid.RepoCfg, error) {
	repoConfigFile := globalCfg.RepoConfigFile(repoID)
	configFile := p.repoCfgPath(absRepoDir, repoConfigFile)
	configData, err := os.ReadFile(configFile) // nolint: gosec

	if err != nil {
		return valid.RepoCfg{}, fmt.Errorf("unable to read %s file: %w", repoConfigFile, err)
	}

	// Parse YAML first to expand glob patterns before validation
	var rawConfig raw.RepoCfg
	decoder := yaml.NewDecoder(bytes.NewReader(configData))
	decoder.KnownFields(true)
	err = decoder.Decode(&rawConfig)
	if err != nil && !errors.Is(err, io.EOF) {
		return valid.RepoCfg{}, err
	}

	// Expand glob patterns in project dirs
	expandedProjects, err := p.expandProjectGlobs(absRepoDir, rawConfig.Projects)
	if err != nil {
		return valid.RepoCfg{}, err
	}
	rawConfig.Projects = expandedProjects

	return p.parseRawRepoCfg(rawConfig, globalCfg, repoID, branch)
}

// ParseRepoCfgData parses repo config from raw YAML bytes. Note that glob patterns
// in project dirs are NOT expanded here because we don't have access to the repo
// directory. This method is primarily used for skip-clone scenarios.
func (p *ParserValidator) ParseRepoCfgData(repoCfgData []byte, globalCfg valid.GlobalCfg, repoID string, branch string) (valid.RepoCfg, error) {
	var rawConfig raw.RepoCfg

	decoder := yaml.NewDecoder(bytes.NewReader(repoCfgData))
	decoder.KnownFields(true)

	err := decoder.Decode(&rawConfig)
	if err != nil && !errors.Is(err, io.EOF) {
		return valid.RepoCfg{}, err
	}

	return p.parseRawRepoCfg(rawConfig, globalCfg, repoID, branch)
}

// parseRawRepoCfg validates and processes a raw config into a valid config.
// This is the shared logic between ParseRepoCfg and ParseRepoCfgData.
func (p *ParserValidator) parseRawRepoCfg(rawConfig raw.RepoCfg, globalCfg valid.GlobalCfg, repoID string, branch string) (valid.RepoCfg, error) {
	// Set ErrorTag to yaml so it uses the YAML field names in error messages.
	validation.ErrorTag = "yaml"
	if err := rawConfig.Validate(); err != nil {
		return valid.RepoCfg{}, err
	}

	validConfig := rawConfig.ToValid()

	// Filter the repo config's projects based on pull request's branch. Only
	// keep projects that either:
	//
	//   - Have no branch regex defined at all (i.e. match all branches), or
	//   - Those that have branch regex matching the PR's base branch.
	//
	i := 0
	for _, p := range validConfig.Projects {
		if branch == "" || p.BranchRegex == nil || p.BranchRegex.MatchString(branch) {
			validConfig.Projects[i] = p
			i++
		}
	}
	validConfig.Projects = validConfig.Projects[:i]

	// We do the project name validation after we get the valid config because
	// we need the defaults of dir and workspace to be populated.
	if err := p.validateProjectNames(validConfig); err != nil {
		return valid.RepoCfg{}, err
	}
	if validConfig.Version == 2 {
		// The only difference between v2 and v3 is how we parse custom run
		// commands.
		if err := p.applyLegacyShellParsing(&validConfig); err != nil {
			return validConfig, err
		}
	}

	err := globalCfg.ValidateRepoCfg(validConfig, repoID)
	return validConfig, err
}

// ParseGlobalCfg returns the parsed and validated global repo config file at
// configFile. defaultCfg will be merged into the parsed config.
// If there is no file at configFile it will return an error.
func (p *ParserValidator) ParseGlobalCfg(configFile string, defaultCfg valid.GlobalCfg) (valid.GlobalCfg, error) {
	configData, err := os.ReadFile(configFile) // nolint: gosec
	if err != nil {
		return valid.GlobalCfg{}, fmt.Errorf("unable to read %s file: %w", configFile, err)
	}
	if len(configData) == 0 {
		return valid.GlobalCfg{}, fmt.Errorf("file %s was empty", configFile)
	}

	var rawCfg raw.GlobalCfg

	decoder := yaml.NewDecoder(bytes.NewReader(configData))
	decoder.KnownFields(true)

	err = decoder.Decode(&rawCfg)
	if err != nil && !errors.Is(err, io.EOF) {
		return valid.GlobalCfg{}, err
	}

	return p.validateRawGlobalCfg(rawCfg, defaultCfg, "yaml")
}

// ParseGlobalCfgJSON parses a json string cfgJSON into global config.
func (p *ParserValidator) ParseGlobalCfgJSON(cfgJSON string, defaultCfg valid.GlobalCfg) (valid.GlobalCfg, error) {
	var rawCfg raw.GlobalCfg
	err := json.Unmarshal([]byte(cfgJSON), &rawCfg)
	if err != nil {
		return valid.GlobalCfg{}, err
	}
	return p.validateRawGlobalCfg(rawCfg, defaultCfg, "json")
}

func (p *ParserValidator) validateRawGlobalCfg(rawCfg raw.GlobalCfg, defaultCfg valid.GlobalCfg, errTag string) (valid.GlobalCfg, error) {
	// Setting ErrorTag means our errors will use the field names defined in
	// the struct tags for yaml/json.
	validation.ErrorTag = errTag
	if err := rawCfg.Validate(); err != nil {
		return valid.GlobalCfg{}, err
	}

	validCfg := rawCfg.ToValid(defaultCfg)
	return validCfg, nil
}

func (p *ParserValidator) repoCfgPath(repoDir, cfgFilename string) string {
	return filepath.Join(repoDir, cfgFilename)
}

func (p *ParserValidator) validateProjectNames(config valid.RepoCfg) error {
	// First, validate that all names are unique.
	seen := make(map[string]bool)
	for _, project := range config.Projects {
		if project.Name != nil {
			name := *project.Name
			exists := seen[name]
			if exists {
				return fmt.Errorf("found two or more projects with name %q; project names must be unique", name)
			}
			seen[name] = true
		}
	}

	// Next, validate that all dir/workspace combos are named.
	// This map's keys will be 'dir/workspace' and the values are the names for
	// that project.
	dirWorkspaceToNames := make(map[string][]string)
	for _, project := range config.Projects {
		key := fmt.Sprintf("%s/%s", project.Dir, project.Workspace)
		names := dirWorkspaceToNames[key]

		// If there is already a project with this dir/workspace then this
		// project must have a name.
		if len(names) > 0 && project.Name == nil {
			return fmt.Errorf("there are two or more projects with dir: %q workspace: %q that are not all named; they must have a 'name' key so they can be targeted for apply's separately", project.Dir, project.Workspace)
		}
		var name string
		if project.Name != nil {
			name = *project.Name
		}
		dirWorkspaceToNames[key] = append(dirWorkspaceToNames[key], name)
	}

	return nil
}

// applyLegacyShellParsing changes any custom run commands in cfg to use the old
// parsing method with shlex.Split().
func (p *ParserValidator) applyLegacyShellParsing(cfg *valid.RepoCfg) error {
	legacyParseF := func(s *valid.Step) error {
		if s.StepName == "run" {
			split, err := shlex.Split(s.RunCommand)
			if err != nil {
				return fmt.Errorf("unable to parse %q: %w", s.RunCommand, err)
			}
			s.RunCommand = strings.Join(split, " ")
		}
		return nil
	}

	for k := range cfg.Workflows {
		w := cfg.Workflows[k]
		for i := range w.Plan.Steps {
			s := &w.Plan.Steps[i]
			if err := legacyParseF(s); err != nil {
				return err
			}
		}
		for i := range w.Apply.Steps {
			s := &w.Apply.Steps[i]
			if err := legacyParseF(s); err != nil {
				return err
			}
		}
		cfg.Workflows[k] = w
	}
	return nil
}

// expandProjectGlobs expands projects with glob patterns in their dir field
// into multiple projects, one for each matching directory that contains
// Terraform files (.tf).
func (p *ParserValidator) expandProjectGlobs(absRepoDir string, projects []raw.Project) ([]raw.Project, error) {
	var expandedProjects []raw.Project

	for _, project := range projects {
		// If dir is nil or doesn't contain glob patterns, keep the project as-is
		if project.Dir == nil || !raw.ContainsGlobPattern(*project.Dir) {
			expandedProjects = append(expandedProjects, project)
			continue
		}

		// Expand the glob pattern
		pattern := filepath.Join(absRepoDir, *project.Dir)
		matches, err := doublestar.FilepathGlob(pattern)
		if err != nil {
			return nil, fmt.Errorf("error expanding glob pattern %q: %w", *project.Dir, err)
		}

		// Filter matches to only include directories with Terraform files
		for _, match := range matches {
			// Check if it's a directory
			info, err := os.Stat(match)
			if err != nil || !info.IsDir() {
				continue
			}

			// Check if the directory contains any .tf files
			hasTerraformFiles, err := p.dirContainsTerraformFiles(match)
			if err != nil {
				return nil, fmt.Errorf("error checking for Terraform files in %q: %w", match, err)
			}
			if !hasTerraformFiles {
				continue
			}

			// Create a new project for this matched directory
			// Calculate the relative path from the repo root
			relDir, err := filepath.Rel(absRepoDir, match)
			if err != nil {
				return nil, fmt.Errorf("error getting relative path for %q: %w", match, err)
			}

			// Copy the project and set the expanded directory
			expandedProject := p.copyProjectWithDir(project, relDir)
			expandedProjects = append(expandedProjects, expandedProject)
		}
	}

	return expandedProjects, nil
}

// dirContainsTerraformFiles returns true if the directory contains at least one .tf file.
func (p *ParserValidator) dirContainsTerraformFiles(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".tf") {
			return true, nil
		}
	}
	return false, nil
}

// copyProjectWithDir creates a copy of a project with a new directory value.
// All other fields are copied from the original project.
func (p *ParserValidator) copyProjectWithDir(original raw.Project, newDir string) raw.Project {
	// Create a new project with the expanded directory
	dirCopy := newDir
	newProject := raw.Project{
		Dir:                       &dirCopy,
		Branch:                    original.Branch,
		Workspace:                 original.Workspace,
		Workflow:                  original.Workflow,
		TerraformDistribution:     original.TerraformDistribution,
		TerraformVersion:          original.TerraformVersion,
		Autoplan:                  original.Autoplan,
		PlanRequirements:          original.PlanRequirements,
		ApplyRequirements:         original.ApplyRequirements,
		ImportRequirements:        original.ImportRequirements,
		DependsOn:                 original.DependsOn,
		DeleteSourceBranchOnMerge: original.DeleteSourceBranchOnMerge,
		RepoLocking:               original.RepoLocking,
		RepoLocks:                 original.RepoLocks,
		ExecutionOrderGroup:       original.ExecutionOrderGroup,
		PolicyCheck:               original.PolicyCheck,
		CustomPolicyCheck:         original.CustomPolicyCheck,
		SilencePRComments:         original.SilencePRComments,
	}

	// Note: We intentionally do NOT copy the Name field.
	// Each expanded project should be identified by its dir+workspace combination.
	// If users need unique names, they should not use glob patterns for that project.

	return newProject
}
