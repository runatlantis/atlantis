package events

import (
	"path"
	"strings"

	"github.com/hootsuite/atlantis/server/events/models"
	"github.com/hootsuite/atlantis/server/logging"
)

//go:generate pegomock generate --use-experimental-model-gen --package mocks -o mocks/mock_modified_project_finder.go ModifiedProjectFinder

type ModifiedProjectFinder interface {
	// FindModified returns the list of projects that were modified based on
	// the modifiedFiles. The list will be de-duplicated.
	FindModified(log *logging.SimpleLogger, modifiedFiles []string, repoFullName string) []models.Project
}

// ProjectFinder identifies projects in a repo.
type ProjectFinder struct{}

var excludeList = []string{"terraform.tfstate", "terraform.tfstate.backup", "_modules", "modules"}

// FindModified returns the list of projects that were modified based on
// the modifiedFiles. The list will be de-duplicated.
func (p *ProjectFinder) FindModified(log *logging.SimpleLogger, modifiedFiles []string, repoFullName string) []models.Project {
	var projects []models.Project

	modifiedTerraformFiles := p.filterToTerraform(modifiedFiles)
	if len(modifiedTerraformFiles) == 0 {
		return projects
	}
	log.Info("filtered modified files to %d non-module .tf files: %v",
		len(modifiedTerraformFiles), modifiedTerraformFiles)

	var paths []string
	for _, modifiedFile := range modifiedFiles {
		paths = append(paths, p.getProjectPath(modifiedFile))
	}
	uniquePaths := p.unique(paths)
	for _, uniquePath := range uniquePaths {
		projects = append(projects, models.NewProject(repoFullName, uniquePath))
	}
	log.Info("there are %d modified project(s) at path(s): %v",
		len(projects), strings.Join(uniquePaths, ", "))
	return projects
}

func (p *ProjectFinder) filterToTerraform(files []string) []string {
	var filtered []string
	for _, fileName := range files {
		if !p.isInExcludeList(fileName) && strings.Contains(fileName, ".tf") {
			filtered = append(filtered, fileName)
		}
	}
	return filtered
}

func (p *ProjectFinder) isInExcludeList(fileName string) bool {
	for _, s := range excludeList {
		if strings.Contains(fileName, s) {
			return true
		}
	}
	return false
}

// getProjectPath returns the path to the project relative to the repo root
// if the project is at the root returns "."
func (p *ProjectFinder) getProjectPath(modifiedFilePath string) string {
	dir := path.Dir(modifiedFilePath)
	if path.Base(dir) == "env" {
		// if the modified file was inside an env/ directory, we treat this specially and
		// run plan one level up
		return path.Dir(dir)
	}
	return dir
}

func (p *ProjectFinder) unique(strs []string) []string {
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
