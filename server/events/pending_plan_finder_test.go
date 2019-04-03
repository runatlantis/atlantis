package events_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/events"
	. "github.com/runatlantis/atlantis/testing"
)

// If the dir doesn't exist should get an error.
func TestPendingPlanFinder_FindNoDir(t *testing.T) {
	pf := &events.DefaultPendingPlanFinder{}
	_, err := pf.Find("/doesntexist")
	ErrEquals(t, "open /doesntexist: no such file or directory", err)
}

// Test different directory structures.
func TestPendingPlanFinder_Find(t *testing.T) {
	cases := []struct {
		description string
		files       map[string]interface{}
		expPlans    []events.PendingPlan
	}{
		{
			"no plans",
			nil,
			nil,
		},
		{
			"root directory",
			map[string]interface{}{
				"default": map[string]interface{}{
					"default.tfplan": nil,
				},
			},
			[]events.PendingPlan{
				{
					RepoDir:    "???/default",
					RepoRelDir: ".",
					Workspace:  "default",
				},
			},
		},
		{
			"root dir project plan",
			map[string]interface{}{
				"default": map[string]interface{}{
					"projectname-default.tfplan": nil,
				},
			},
			[]events.PendingPlan{
				{
					RepoDir:     "???/default",
					RepoRelDir:  ".",
					Workspace:   "default",
					ProjectName: "projectname",
				},
			},
		},
		{
			"root dir project plan with slashes",
			map[string]interface{}{
				"default": map[string]interface{}{
					"project::name-default.tfplan": nil,
				},
			},
			[]events.PendingPlan{
				{
					RepoDir:     "???/default",
					RepoRelDir:  ".",
					Workspace:   "default",
					ProjectName: "project/name",
				},
			},
		},
		{
			"multiple directories in single workspace",
			map[string]interface{}{
				"default": map[string]interface{}{
					"dir1": map[string]interface{}{
						"default.tfplan": nil,
					},
					"dir2": map[string]interface{}{
						"default.tfplan": nil,
					},
				},
			},
			[]events.PendingPlan{
				{
					RepoDir:    "???/default",
					RepoRelDir: "dir1",
					Workspace:  "default",
				},
				{
					RepoDir:    "???/default",
					RepoRelDir: "dir2",
					Workspace:  "default",
				},
			},
		},
		{
			"multiple directories nested within each other",
			map[string]interface{}{
				"default": map[string]interface{}{
					"dir1": map[string]interface{}{
						"default.tfplan": nil,
					},
					"default.tfplan": nil,
				},
			},
			[]events.PendingPlan{
				{
					RepoDir:    "???/default",
					RepoRelDir: ".",
					Workspace:  "default",
				},
				{
					RepoDir:    "???/default",
					RepoRelDir: "dir1",
					Workspace:  "default",
				},
			},
		},
		{
			"multiple workspaces",
			map[string]interface{}{
				"default": map[string]interface{}{
					"default.tfplan": nil,
				},
				"staging": map[string]interface{}{
					"staging.tfplan": nil,
				},
				"production": map[string]interface{}{
					"production.tfplan": nil,
				},
			},
			[]events.PendingPlan{
				{
					RepoDir:    "???/default",
					RepoRelDir: ".",
					Workspace:  "default",
				},
				{
					RepoDir:    "???/production",
					RepoRelDir: ".",
					Workspace:  "production",
				},
				{
					RepoDir:    "???/staging",
					RepoRelDir: ".",
					Workspace:  "staging",
				},
			},
		},
		{
			".terragrunt-cache",
			map[string]interface{}{
				"default": map[string]interface{}{
					".terragrunt-cache": map[string]interface{}{
						"N6lY9xk7PivbOAzdsjDL6VUFVYk": map[string]interface{}{
							"K4xpUZI6HgUF-ip6E1eib4L8mwQ": map[string]interface{}{
								"app": map[string]interface{}{
									"default.tfplan": nil,
								},
							},
						},
					},
					"default.tfplan": nil,
				},
			},
			[]events.PendingPlan{
				{
					RepoDir:    "???/default",
					RepoRelDir: ".",
					Workspace:  "default",
				},
			},
		},
	}

	pf := &events.DefaultPendingPlanFinder{}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			tmpDir, cleanup := DirStructure(t, c.files)
			defer cleanup()

			// Create a git repo in each workspace directory.
			for dirname, contents := range c.files {
				// If contents is nil then this isn't a directory.
				if contents != nil {
					runCmd(t, filepath.Join(tmpDir, dirname), "git", "init")
				}
			}

			actPlans, err := pf.Find(tmpDir)
			Ok(t, err)

			// Replace the actual dir with ??? to allow for comparison.
			var actPlansComparable []events.PendingPlan
			for _, p := range actPlans {
				p.RepoDir = strings.Replace(p.RepoDir, tmpDir, "???", -1)
				actPlansComparable = append(actPlansComparable, p)
			}
			Equals(t, c.expPlans, actPlansComparable)
		})
	}
}

// If a planfile is checked in to git, we shouldn't use it.
func TestPendingPlanFinder_FindPlanCheckedIn(t *testing.T) {
	tmpDir, cleanup := DirStructure(t, map[string]interface{}{
		"default": map[string]interface{}{
			"default.tfplan": nil,
		},
	})
	defer cleanup()

	// Add that file to git.
	repoDir := filepath.Join(tmpDir, "default")
	runCmd(t, repoDir, "git", "init")
	runCmd(t, repoDir, "touch", ".gitkeep")
	runCmd(t, repoDir, "git", "add", ".")
	runCmd(t, repoDir, "git", "config", "--local", "user.email", "atlantisbot@runatlantis.io")
	runCmd(t, repoDir, "git", "config", "--local", "user.name", "atlantisbot")
	runCmd(t, repoDir, "git", "commit", "--no-gpg-sign", "-m", "initial commit")

	pf := &events.DefaultPendingPlanFinder{}
	actPlans, err := pf.Find(tmpDir)
	Ok(t, err)
	Equals(t, 0, len(actPlans))
}

// Test that it deletes pending plans.
func TestPendingPlanFinder_DeletePlans(t *testing.T) {
	files := map[string]interface{}{
		"default": map[string]interface{}{
			"dir1": map[string]interface{}{
				"default.tfplan": nil,
			},
			"dir2": map[string]interface{}{
				"default.tfplan": nil,
			},
		},
	}
	tmp, cleanup := DirStructure(t,
		files)
	defer cleanup()

	// Create a git repo in each workspace directory.
	for dirname, contents := range files {
		// If contents is nil then this isn't a directory.
		if contents != nil {
			runCmd(t, filepath.Join(tmp, dirname), "git", "init")
		}
	}

	pf := &events.DefaultPendingPlanFinder{}
	Ok(t, pf.DeletePlans(tmp))

	// First, check the files were deleted.
	for _, plan := range []string{
		"default/dir1/default.tfplan",
		"default/dir2/default.tfplan",
	} {
		absPath := filepath.Join(tmp, plan)
		_, err := os.Stat(absPath)
		ErrContains(t, "no such file or directory", err)
	}

	// Double check by using Find().
	foundPlans, err := pf.Find(tmp)
	Ok(t, err)
	Equals(t, 0, len(foundPlans))
}

func runCmd(t *testing.T, dir string, name string, args ...string) string {
	t.Helper()
	cpCmd := exec.Command(name, args...)
	cpCmd.Dir = dir
	cpOut, err := cpCmd.CombinedOutput()
	Assert(t, err == nil, "err running %q: %s", strings.Join(append([]string{name}, args...), " "), cpOut)
	return string(cpOut)
}
