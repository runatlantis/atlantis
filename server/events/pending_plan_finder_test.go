// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

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

// Non-git subdirectories in the pull dir should be silently skipped.
func TestPendingPlanFinder_FindIncludingNotGitDir(t *testing.T) {
	gitDirName := "default"
	notGitDirName := "reviews"
	tmpDir := DirStructure(t, map[string]any{
		gitDirName: map[string]any{
			"default.tfplan": nil,
		},
		notGitDirName: map[string]any{
			"some_file.tfplan": nil,
		},
	})
	// Initialize git only in the workspace directory, not in the stray directory.
	gitDir := filepath.Join(tmpDir, gitDirName)
	runCmd(t, gitDir, "git", "init")
	pf := &events.DefaultPendingPlanFinder{}

	plans, err := pf.Find(tmpDir)
	Ok(t, err)
	// Only the plan from the git workspace should be found; the reviews dir is skipped.
	Equals(t, 1, len(plans))
	Equals(t, gitDirName, plans[0].Workspace)
}

// Non-directory entries (files, symlinks) in the pull dir should be silently skipped.
func TestPendingPlanFinder_FindSkipsNonDirEntries(t *testing.T) {
	tmpDir := DirStructure(t, map[string]any{
		"default": map[string]any{
			"default.tfplan": nil,
		},
	})
	runCmd(t, filepath.Join(tmpDir, "default"), "git", "init")

	// Create a plain file at the top level of the pull dir (not a workspace clone).
	if err := os.WriteFile(filepath.Join(tmpDir, "somefile.txt"), []byte("data"), 0600); err != nil {
		t.Fatal(err)
	}

	pf := &events.DefaultPendingPlanFinder{}
	plans, err := pf.Find(tmpDir)
	Ok(t, err)
	Equals(t, 1, len(plans))
	Equals(t, "default", plans[0].Workspace)
}

// Directories that are only inside a parent git repo are not workspace clone roots.
func TestPendingPlanFinder_FindSkipsDirInsideParentGitRepo(t *testing.T) {
	gitDirName := "default"
	notGitDirName := "reviews"
	tmpDir := DirStructure(t, map[string]any{
		gitDirName: map[string]any{
			"default.tfplan": nil,
		},
		notGitDirName: map[string]any{
			"some_file.tfplan": nil,
		},
	})
	runCmd(t, tmpDir, "git", "init")
	runCmd(t, filepath.Join(tmpDir, gitDirName), "git", "init")

	pf := &events.DefaultPendingPlanFinder{}
	plans, err := pf.Find(tmpDir)

	Ok(t, err)
	Equals(t, 1, len(plans))
	Equals(t, gitDirName, plans[0].Workspace)
}

// Test different directory structures.
func TestPendingPlanFinder_Find(t *testing.T) {
	cases := []struct {
		description string
		files       map[string]any
		expPlans    []events.PendingPlan
	}{
		{
			"no plans",
			nil,
			nil,
		},
		{
			"root directory",
			map[string]any{
				"default": map[string]any{
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
			map[string]any{
				"default": map[string]any{
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
			map[string]any{
				"default": map[string]any{
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
			map[string]any{
				"default": map[string]any{
					"dir1": map[string]any{
						"default.tfplan": nil,
					},
					"dir2": map[string]any{
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
			map[string]any{
				"default": map[string]any{
					"dir1": map[string]any{
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
			map[string]any{
				"default": map[string]any{
					"default.tfplan": nil,
				},
				"staging": map[string]any{
					"staging.tfplan": nil,
				},
				"production": map[string]any{
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
			map[string]any{
				"default": map[string]any{
					".terragrunt-cache": map[string]any{
						"N6lY9xk7PivbOAzdsjDL6VUFVYk": map[string]any{
							"K4xpUZI6HgUF-ip6E1eib4L8mwQ": map[string]any{
								"app": map[string]any{
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
			tmpDir := DirStructure(t, c.files)

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
				p.RepoDir = strings.ReplaceAll(p.RepoDir, tmpDir, "???")
				actPlansComparable = append(actPlansComparable, p)
			}
			Equals(t, c.expPlans, actPlansComparable)
		})
	}
}

// If a planfile is checked in to git, we shouldn't use it.
func TestPendingPlanFinder_FindPlanCheckedIn(t *testing.T) {
	tmpDir := DirStructure(t, map[string]any{
		"default": map[string]any{
			"default.tfplan": nil,
		},
	})

	// Add that file to git.
	repoDir := filepath.Join(tmpDir, "default")
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

func runCmdErrCode(t *testing.T, dir string, errCode int, name string, args ...string) string {
	t.Helper()
	cpCmd := exec.Command(name, args...)
	cpCmd.Dir = dir
	cpOut, err := cpCmd.CombinedOutput()
	cmd := strings.Join(append([]string{name}, args...), " ")
	if err != nil {
		if eerr, ok := err.(*exec.ExitError); ok {
			Assert(t, errCode == eerr.ExitCode(), "unexpected exit code: want %v, got %v, running %q: %s", errCode, eerr.ExitCode(), cmd, cpCmd)
			return string(cpOut)
		}
	}
	Assert(t, false, "invalid exit code, running %q: %s", cmd, cpOut)
	return string(cpOut)
}

// Test that it deletes pending plans.
func TestPendingPlanFinder_DeletePlans(t *testing.T) {
	files := map[string]any{
		"default": map[string]any{
			"dir1": map[string]any{
				"default.tfplan": nil,
			},
			"dir2": map[string]any{
				"default.tfplan": nil,
			},
		},
	}
	tmp := DirStructure(t, files)

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
