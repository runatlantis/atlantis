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
				"default.tfplan": nil,
			},
			[]events.PendingPlan{
				{
					RepoDir:    "???",
					RepoRelDir: ".",
					Workspace:  "default",
				},
			},
		},
		{
			"root dir project plan",
			map[string]interface{}{
				"projectname::default.tfplan": nil,
			},
			[]events.PendingPlan{
				{
					RepoDir:     "???",
					RepoRelDir:  ".",
					Workspace:   "default",
					ProjectName: "projectname",
				},
			},
		},
		{
			"root dir project plan with slashes",
			map[string]interface{}{
				"project::name::default.tfplan": nil,
			},
			[]events.PendingPlan{
				{
					RepoDir:     "???",
					RepoRelDir:  ".",
					Workspace:   "default",
					ProjectName: "project/name",
				},
			},
		},
		{
			"multiple directories in single workspace",
			map[string]interface{}{
				"dir1": map[string]interface{}{
					"default.tfplan": nil,
				},
				"dir2": map[string]interface{}{
					"default.tfplan": nil,
				},
			},
			[]events.PendingPlan{
				{
					RepoDir:    "???",
					RepoRelDir: "dir1",
					Workspace:  "default",
				},
				{
					RepoDir:    "???",
					RepoRelDir: "dir2",
					Workspace:  "default",
				},
			},
		},
		{
			"multiple directories nested within each other",
			map[string]interface{}{
				"dir1": map[string]interface{}{
					"default.tfplan": nil,
				},
				"default.tfplan": nil,
			},
			[]events.PendingPlan{
				{
					RepoDir:    "???",
					RepoRelDir: ".",
					Workspace:  "default",
				},
				{
					RepoDir:    "???",
					RepoRelDir: "dir1",
					Workspace:  "default",
				},
			},
		},
		{
			"multiple workspaces",
			map[string]interface{}{
				"default.tfplan":    nil,
				"staging.tfplan":    nil,
				"production.tfplan": nil,
			},
			[]events.PendingPlan{
				{
					RepoDir:    "???",
					RepoRelDir: ".",
					Workspace:  "default",
				},
				{
					RepoDir:    "???",
					RepoRelDir: ".",
					Workspace:  "production",
				},
				{
					RepoDir:    "???",
					RepoRelDir: ".",
					Workspace:  "staging",
				},
			},
		},
		{
			".terragrunt-cache",
			map[string]interface{}{
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
			[]events.PendingPlan{
				{
					RepoDir:    "???",
					RepoRelDir: ".",
					Workspace:  "default",
				},
			},
		},
	}

	pf := &events.DefaultPendingPlanFinder{}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			tmpDir := DirStructure(t, c.files)

			runCmd(t, tmpDir, "git", "init")

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
	tmpDir := DirStructure(t, map[string]interface{}{
		"default.tfplan": nil,
	})

	// Add that file to git.
	repoDir := tmpDir
	runCmd(t, repoDir, "git", "init")
	runCmd(t, repoDir, "touch", ".gitkeep")
	runCmd(t, repoDir, "git", "add", ".")
	runCmd(t, repoDir, "git", "config", "--local", "user.email", "atlantisbot@runatlantis.io")
	runCmd(t, repoDir, "git", "config", "--local", "user.name", "atlantisbot")
	runCmd(t, repoDir, "git", "config", "--local", "commit.gpgsign", "false")
	runCmd(t, repoDir, "git", "commit", "-m", "initial commit")

	pf := &events.DefaultPendingPlanFinder{}
	actPlans, err := pf.Find(tmpDir)
	Ok(t, err)
	Equals(t, 0, len(actPlans))
}

// Test that it deletes pending plans.
func TestPendingPlanFinder_DeletePlans(t *testing.T) {
	files := map[string]interface{}{
		"dir1": map[string]interface{}{
			"default.tfplan": nil,
		},
		"dir2": map[string]interface{}{
			"default.tfplan": nil,
		},
	}
	tmp := DirStructure(t, files)

	runCmd(t, tmp, "git", "init")

	pf := &events.DefaultPendingPlanFinder{}
	Ok(t, pf.DeletePlans(tmp))

	// First, check the files were deleted.
	for _, plan := range []string{
		"dir1/default.tfplan",
		"dir2/default.tfplan",
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
