package events_test

import (
	"testing"

	"github.com/atlantisnorth/atlantis/server/events"
	"github.com/atlantisnorth/atlantis/server/logging"
	. "github.com/atlantisnorth/atlantis/testing"
)

var noopLogger = logging.NewNoopLogger()
var modifiedRepo = "owner/repo"
var m = events.DefaultProjectFinder{}

func TestGetModified(t *testing.T) {
	cases := []struct {
		description     string
		files           []string
		expProjectPaths []string
	}{
		{
			"If no files were modified then should return an empty list",
			nil,
			nil,
		},
		{
			"Should ignore non .tf files and return an empty list",
			[]string{"non-tf"},
			nil,
		},
		{
			"Should plan in the parent directory from modules",
			[]string{"modules/file.tf"},
			[]string{"."},
		},
		{
			"Should plan in the parent directory from modules when module is in a subdir",
			[]string{"modules/subdir/file.tf"},
			[]string{"."},
		},
		{
			"Should plan in the parent directory from modules when project is in its own dir",
			[]string{"projectdir/modules/file.tf"},
			[]string{"projectdir"},
		},
		{
			"Should ignore tfstate files and return an empty list",
			[]string{"terraform.tfstate", "terraform.tfstate.backup", "parent/terraform.tfstate", "parent/terraform.tfstate.backup"},
			nil,
		},
		{
			"Should ignore tfstate files and return an empty list",
			[]string{"terraform.tfstate", "terraform.tfstate.backup", "parent/terraform.tfstate", "parent/terraform.tfstate.backup"},
			nil,
		},
		{
			"Should return '.' when changed file is at root",
			[]string{"a.tf"},
			[]string{"."},
		},
		{
			"Should return directory when changed file is in a dir",
			[]string{"parent/a.tf"},
			[]string{"parent"},
		},
		{
			"Should return parent dir when changed file is in an env/ dir",
			[]string{"env/a.tfvars"},
			[]string{"."},
		},
		{
			"Should de-duplicate when multiple files changed in the same dir",
			[]string{"root.tf", "env/env.tfvars", "parent/parent.tf", "parent/parent2.tf", "parent/child/child.tf", "parent/child/env/env.tfvars"},
			[]string{".", "parent", "parent/child"},
		},
	}
	for _, c := range cases {
		t.Log(c.description)
		projects := m.FindModified(noopLogger, c.files, modifiedRepo)

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
