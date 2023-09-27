// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

package events

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/utils"

	"github.com/moby/patternmatcher"
	"github.com/pkg/errors"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
)

// ProjectFinder determines which projects were modified in a given pull
// request.
type ProjectFinder interface {
	// DetermineProjects returns the list of projects that were modified based on
	// the modifiedFiles. The list will be de-duplicated.
	// absRepoDir is the path to the cloned repo on disk.
	DetermineProjects(log logging.SimpleLogging, modifiedFiles []string, repoFullName string, absRepoDir string, autoplanFileList string, moduleInfo ModuleProjects) []models.Project
	// DetermineProjectsViaConfig returns the list of projects that were modified
	// based on modifiedFiles and the repo's config.
	// absRepoDir is the path to the cloned repo on disk.
	DetermineProjectsViaConfig(log logging.SimpleLogging, modifiedFiles []string, config valid.RepoCfg, absRepoDir string, moduleInfo ModuleProjects) ([]valid.Project, error)

	DetermineWorkspaceFromHCL(log logging.SimpleLogging, absRepoDir string) (string, error)
}

var rootBlockSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       "terraform",
			LabelNames: nil,
		},
	},
}

var terraformBlockSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: "cloud",
		},
	},
}

var cloudBlockSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: "workspaces",
		},
	},
}

func (p *DefaultProjectFinder) DetermineWorkspaceFromHCL(log logging.SimpleLogging, absRepoDir string) (string, error) {
	log.Info("looking for Terraform Cloud workspace from configuration in %q", absRepoDir)
	infos, err := os.ReadDir(absRepoDir)
	if err != nil {
		return "", err
	}
	parser := hclparse.NewParser()
	for _, info := range infos {
		if info.IsDir() {
			continue
		}

		name := info.Name()
		if strings.HasSuffix(name, ".tf") {
			fullPath := filepath.Join(absRepoDir, name)
			file, _ := parser.ParseHCLFile(fullPath)
			workspace, err := findTFCloudWorkspaceFromFile(file)
			if err != nil {
				log.Warn(err.Error())
				return DefaultWorkspace, nil
			}

			if len(workspace) > 0 {
				log.Debug("found configured Terraform Cloud workspace with name %q", workspace)
				return workspace, nil
			}
		}
	}

	log.Debug("no Terraform Cloud workspace explicitly configured in Terraform codes. Use default workspace (%q)", DefaultWorkspace)
	return DefaultWorkspace, nil
}

func findTFCloudWorkspaceFromFile(file *hcl.File) (string, error) {
	content, _, _ := file.Body.PartialContent(rootBlockSchema)
	workspace := ""

	if len(content.Blocks) == 1 {
		content, _, _ = content.Blocks[0].Body.PartialContent(terraformBlockSchema)
		if len(content.Blocks) == 1 {
			content, _, _ = content.Blocks[0].Body.PartialContent(cloudBlockSchema)
			if len(content.Blocks) == 1 {
				attrs, _ := content.Blocks[0].Body.JustAttributes()
				if nameAttr, defined := attrs["name"]; defined {
					diags := gohcl.DecodeExpression(nameAttr.Expr, nil, &workspace)
					if diags.HasErrors() {
						return "", fmt.Errorf("unable to parse workspace configuration: %q", diags.Error())
					}
				}
			}
		}
	}
	return workspace, nil
}

// ignoredFilenameFragments contains filename fragments to ignore while looking at changes
var ignoredFilenameFragments = []string{"terraform.tfstate", "terraform.tfstate.backup", "tflint.hcl"}

// DefaultProjectFinder implements ProjectFinder.
type DefaultProjectFinder struct{}

// See ProjectFinder.DetermineProjects.
func (p *DefaultProjectFinder) DetermineProjects(log logging.SimpleLogging, modifiedFiles []string, repoFullName string, absRepoDir string, autoplanFileList string, moduleInfo ModuleProjects) []models.Project {
	var projects []models.Project

	modifiedTerraformFiles := p.filterToFileList(log, modifiedFiles, autoplanFileList)
	if len(modifiedTerraformFiles) == 0 {
		return projects
	}
	log.Info("filtered modified files to %d file(s) in the autoplan file list: %v",
		len(modifiedTerraformFiles), modifiedTerraformFiles)

	var dirs []string
	for _, modifiedFile := range modifiedTerraformFiles {
		projectDir := getProjectDir(log, modifiedFile, absRepoDir)
		if projectDir != "" {
			dirs = append(dirs, projectDir)
		} else if moduleInfo != nil {
			downstreamProjects := moduleInfo.DependentProjects(path.Dir(modifiedFile))
			log.Debug("found downstream projects for %q: %v", modifiedFile, downstreamProjects)
			dirs = append(dirs, downstreamProjects...)
		}
	}
	uniqueDirs := p.unique(dirs)

	// The list of modified files will include files that were deleted. We still
	// want to run plan if a file was deleted since that often results in a
	// change however we want to remove directories that have been completely
	// deleted.
	exists := p.removeNonExistingDirs(uniqueDirs, absRepoDir)

	for _, p := range exists {
		projects = append(projects, models.NewProject(repoFullName, p))
	}
	log.Info("there are %d modified project(s) at path(s): %v",
		len(projects), strings.Join(exists, ", "))
	return projects
}

// See ProjectFinder.DetermineProjectsViaConfig.
func (p *DefaultProjectFinder) DetermineProjectsViaConfig(log logging.SimpleLogging, modifiedFiles []string, config valid.RepoCfg, absRepoDir string, moduleInfo ModuleProjects) ([]valid.Project, error) {

	// Check moduleInfo for downstream project dependencies
	var dependentProjects []string
	for _, file := range modifiedFiles {
		if moduleInfo != nil {
			downstreamProjects := moduleInfo.DependentProjects(path.Dir(file))
			log.Debug("found downstream projects for %q: %v", file, downstreamProjects)
			dependentProjects = append(dependentProjects, downstreamProjects...)
		}
	}

	var projects []valid.Project
	for _, project := range config.Projects {
		log.Debug("checking if project at dir %q workspace %q was modified", project.Dir, project.Workspace)

		if utils.SlicesContains(dependentProjects, project.Dir) {
			projects = append(projects, project)
			continue
		}

		var whenModifiedRelToRepoRoot []string
		for _, wm := range project.Autoplan.WhenModified {
			wm = strings.TrimSpace(wm)
			// An exclusion uses a '!' at the beginning. If it's there, we need
			// to remove it, then add in the project path, then add it back.
			exclusion := false
			if wm != "" && wm[0] == '!' {
				wm = wm[1:]
				exclusion = true
			}

			// Prepend project dir to when modified patterns because the patterns
			// are relative to the project dirs but our list of modified files is
			// relative to the repo root.
			wmRelPath := filepath.Join(project.Dir, wm)
			if exclusion {
				wmRelPath = "!" + wmRelPath
			}
			whenModifiedRelToRepoRoot = append(whenModifiedRelToRepoRoot, wmRelPath)
		}
		pm, err := patternmatcher.New(whenModifiedRelToRepoRoot)
		if err != nil {
			return nil, errors.Wrapf(err, "matching modified files with patterns: %v", project.Autoplan.WhenModified)
		}

		// If any of the modified files matches the pattern then this project is
		// considered modified.
		for _, file := range modifiedFiles {
			match, err := pm.MatchesOrParentMatches(file)
			if err != nil {
				log.Debug("match err for file %q: %s", file, err)
				continue
			}
			if match {
				log.Debug("file %q matched pattern", file)
				// If we're checking using an atlantis.yaml file we downloaded
				// directly from the repo (when doing a no-clone check) then
				// absRepoDir will be empty. Since we didn't clone the repo
				// yet we can't do this check. If there was a file modified
				// in a deleted directory then when we finally do clone the repo
				// we'll call this function again and then we'll detect the
				// directory was deleted.
				if absRepoDir != "" {
					_, err := os.Stat(filepath.Join(absRepoDir, project.Dir))
					if err == nil {
						projects = append(projects, project)
					} else {
						log.Debug("project at dir %q not included because dir does not exist", project.Dir)
					}
				} else {
					projects = append(projects, project)
				}
				break
			}
		}
	}
	return projects, nil
}

// filterToFileList filters out files not included in the file list
func (p *DefaultProjectFinder) filterToFileList(log logging.SimpleLogging, files []string, fileList string) []string {
	var filtered []string
	patterns := strings.Split(fileList, ",")
	// Ignore pattern matcher error here as it was checked for errors in server validation
	patternMatcher, _ := patternmatcher.New(patterns)

	for _, fileName := range files {
		if p.shouldIgnore(fileName) {
			continue
		}
		match, err := patternMatcher.MatchesOrParentMatches(fileName)
		if err != nil {
			log.Debug("filter err for file %q: %s", fileName, err)
			continue
		}
		if match {
			filtered = append(filtered, fileName)
		}
	}

	return filtered
}

// shouldIgnore returns true if we shouldn't trigger a plan on changes to this file.
func (p *DefaultProjectFinder) shouldIgnore(fileName string) bool {
	for _, s := range ignoredFilenameFragments {
		if strings.Contains(fileName, s) {
			return true
		}
	}
	return false
}

// getProjectDir determines which project should be planned based on the path to a changed file.
// - If the directory containing the changed file contains ".tf" files and is not a module, return that
// - If not, check the parent directory. Repeat until the root of the repo is reached
// - If the root of the repo has ".tf" files, return that. If not:
//   - If the directory containing the changed files is not a module, return that
//   - If the directory containing the changed files _is_ a module, return "" (nothing is planned)
//
// File paths are all relative to the root of the repository
func getProjectDir(log logging.SimpleLogging, modifiedFilePath string, repoDir string) string {
	return getProjectDirFromFs(log, os.DirFS(repoDir), modifiedFilePath)
}

func getProjectDirFromFs(log logging.SimpleLogging, files fs.FS, modifiedFilePath string) string {
	dir := path.Dir(modifiedFilePath)

	searchDir := dir
	nonRootDir, err := isNonRoot(searchDir)
	if err != nil {
		log.Warn("Found error checking relative path of %q to repo directory: %s", searchDir, err)
		return dir
	}

	for nonRootDir {
		dirHasTfFiles, err := hasTerraformFiles(files, searchDir)
		if err != nil {
			log.Warn("Found error checking for Terraform files in dir %q: %s", searchDir, err)
			return dir
		}
		if dirHasTfFiles && !isModule(searchDir) {
			return searchDir
		}
		searchDir = path.Dir(searchDir)
		nonRootDir, err = isNonRoot(searchDir)
		if err != nil {
			log.Warn("Found error checking relative path of %q to repo directory: %s", searchDir, err)
			return dir
		}
	}

	rootHasTerraformFiles, err := hasTerraformFiles(files, ".")
	if err != nil {
		log.Warn("Found error checking for Terraform files in repo directory: %s", err)
		return dir
	}
	if rootHasTerraformFiles {
		return "."
	}

	if isModule(dir) {
		return ""
	}
	return dir
}

// Check if a given directory in a filesystem contains any files with the ".tf" extension
func hasTerraformFiles(files fs.FS, dir string) (bool, error) {
	tfFiles, err := fs.Glob(files, dir+"/*.tf")
	if err != nil {
		return false, err
	}
	if len(tfFiles) > 0 {
		return true, nil
	}
	return false, nil
}

func isModule(dir string) bool {
	return strings.Contains("/"+dir+"/", "/modules/")
}

// Check if a given directory is a child of the current working directory
// Directory paths are all relative to the root of the repository
func isNonRoot(dir string) (bool, error) {
	up := "../"

	rel, err := filepath.Rel(".", dir)
	if err != nil {
		return false, err
	}
	if !strings.HasPrefix(rel, up) && rel != ".." && rel != "." {
		return true, nil
	}
	return false, nil
}

// unique de-duplicates strs.
func (p *DefaultProjectFinder) unique(strs []string) []string {
	hash := make(map[string]bool)
	var unique []string
	for _, s := range strs {
		if !hash[s] {
			unique = append(unique, s)
			hash[s] = true
		}
	}
	return unique
}

// removeNonExistingDirs removes paths from relativePaths that don't exist.
// relativePaths is a list of paths relative to absRepoDir.
func (p *DefaultProjectFinder) removeNonExistingDirs(relativePaths []string, absRepoDir string) []string {
	var filtered []string
	for _, pth := range relativePaths {
		absPath := filepath.Join(absRepoDir, pth)
		if _, err := os.Stat(absPath); !os.IsNotExist(err) {
			filtered = append(filtered, pth)
		}
	}
	return filtered
}
