package events_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

var noopLogger = logging.NewNoopLogger()
var modifiedRepo = "owner/repo"
var m = events.DefaultProjectFinder{}
var nestedModules1 string
var nestedModules2 string
var topLevelModules string

func setupTmpRepos(t *testing.T) {
	// Create different repo structures for testing.

	// 1. Nested modules directory inside a project
	// project1/
	//   main.tf
	//   modules/
	//     main.tf
	var err error
	nestedModules1, err = ioutil.TempDir("", "")
	Ok(t, err)
	err = os.MkdirAll(filepath.Join(nestedModules1, "project1/modules"), 0700)
	Ok(t, err)
	_, err = os.Create(filepath.Join(nestedModules1, "project1/main.tf"))
	Ok(t, err)
	_, err = os.Create(filepath.Join(nestedModules1, "project1/modules/main.tf"))
	Ok(t, err)

	// 2. Nested modules dir inside top-level project
	// main.tf
	//  modules/
	//    main.tf
	// We can just re-use part of the previous dir structure.
	nestedModules2 = filepath.Join(nestedModules1, "project1")

	// 3. Top-level modules
	//  modules/
	//    main.tf
	//  project1/
	//    main.tf
	//  project2/
	//    main.tf
	topLevelModules, err = ioutil.TempDir("", "")
	Ok(t, err)
	for _, path := range []string{"modules", "project1", "project2"} {
		err = os.MkdirAll(filepath.Join(topLevelModules, path), 0700)
		Ok(t, err)
		_, err = os.Create(filepath.Join(topLevelModules, path, "main.tf"))
		Ok(t, err)
	}
}

func TestDetermineProjects(t *testing.T) {
	setupTmpRepos(t)

	cases := []struct {
		description     string
		files           []string
		expProjectPaths []string
		repoDir         string
	}{
		{
			"If no files were modified then should return an empty list",
			nil,
			nil,
			"",
		},
		{
			"Should ignore non .tf files and return an empty list",
			[]string{"non-tf"},
			nil,
			"",
		},
		{
			"Should plan in the parent directory from modules if that dir has a main.tf",
			[]string{"project1/modules/main.tf"},
			[]string{"project1"},
			nestedModules1,
		},
		{
			"Should plan in the parent directory from modules if that dir has a main.tf",
			[]string{"modules/main.tf"},
			[]string{"."},
			nestedModules2,
		},
		{
			"Should plan in the parent directory from modules when module is in a subdir if that dir has a main.tf",
			[]string{"modules/subdir/main.tf"},
			[]string{"."},
			nestedModules2,
		},
		{
			"Should not plan in the parent directory from modules if that dir does not have a main.tf",
			[]string{"modules/main.tf"},
			[]string{},
			topLevelModules,
		},
		{
			"Should not plan in the parent directory from modules if that dir does not have a main.tf",
			[]string{"modules/main.tf", "project1/main.tf"},
			[]string{"project1"},
			topLevelModules,
		},
		{
			"Should ignore tfstate files and return an empty list",
			[]string{"terraform.tfstate", "terraform.tfstate.backup", "parent/terraform.tfstate", "parent/terraform.tfstate.backup"},
			nil,
			"",
		},
		{
			"Should ignore tfstate files and return an empty list",
			[]string{"terraform.tfstate", "terraform.tfstate.backup", "parent/terraform.tfstate", "parent/terraform.tfstate.backup"},
			nil,
			"",
		},
		{
			"Should return '.' when changed file is at root",
			[]string{"a.tf"},
			[]string{"."},
			"",
		},
		{
			"Should return directory when changed file is in a dir",
			[]string{"parent/a.tf"},
			[]string{"parent"},
			"",
		},
		{
			"Should return parent dir when changed file is in an env/ dir",
			[]string{"env/a.tfvars"},
			[]string{"."},
			"",
		},
		{
			"Should de-duplicate when multiple files changed in the same dir",
			[]string{"root.tf", "env/env.tfvars", "parent/parent.tf", "parent/parent2.tf", "parent/child/child.tf", "parent/child/env/env.tfvars"},
			[]string{".", "parent", "parent/child"},
			"",
		},
	}
	for _, c := range cases {
		t.Log(c.description)
		projects := m.DetermineProjects(noopLogger, c.files, modifiedRepo, c.repoDir)

		// Extract the paths from the projects. We use a slice here instead of a
		// map so we can test whether there are duplicates returned.
		var paths []string
		for _, project := range projects {
			paths = append(paths, project.Path)
			// Check that the project object has the repo set properly.
			Equals(t, modifiedRepo, project.RepoFullName)
		}
		Assert(t, len(c.expProjectPaths) == len(paths),
			"exp %d paths but found %d. They were %v", len(c.expProjectPaths), len(paths), paths)

		for _, expPath := range c.expProjectPaths {
			found := false
			for _, actPath := range paths {
				if expPath == actPath {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("exp %q but was not in paths %v", expPath, paths)
			}
		}
	}
}
