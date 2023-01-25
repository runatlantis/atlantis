package events_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/petergtz/pegomock"
	lockingMocks "github.com/runatlantis/atlantis/server/core/locking/mocks"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/models/testdata"
	. "github.com/runatlantis/atlantis/testing"
)

// Test different directory structures.
func TestPendingPlanFinder_Find(t *testing.T) {
	cases := []struct {
		description   string
		projectStatus []models.ProjectStatus
		expPlans      []events.PendingPlan
	}{
		{
			"no plans",
			nil,
			nil,
		},
		{
			"root directory",
			[]models.ProjectStatus{
				{
					RepoRelDir: ".",
					Workspace:  "default",
					Status:     models.PlannedPlanStatus,
				},
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
			[]models.ProjectStatus{
				{
					RepoRelDir:  ".",
					Workspace:   "default",
					ProjectName: "projectname",
					Status:      models.PlannedPlanStatus,
				},
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
			[]models.ProjectStatus{
				{
					RepoRelDir:  ".",
					Workspace:   "default",
					ProjectName: "project/name",
					Status:      models.PlannedPlanStatus,
				},
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
			[]models.ProjectStatus{
				{
					RepoRelDir: "dir1",
					Workspace:  "default",
					Status:     models.PlannedPlanStatus,
				},
				{
					RepoRelDir: "dir2",
					Workspace:  "default",
					Status:     models.PlannedPlanStatus,
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
			[]models.ProjectStatus{
				{
					RepoRelDir: ".",
					Workspace:  "default",
					Status:     models.PlannedPlanStatus,
				},
				{
					RepoRelDir: "dir1",
					Workspace:  "default",
					Status:     models.PlannedPlanStatus,
				},
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
			[]models.ProjectStatus{
				{
					RepoRelDir: ".",
					Workspace:  "default",
					Status:     models.PlannedPlanStatus,
				},
				{
					RepoRelDir: ".",
					Workspace:  "production",
					Status:     models.PlannedPlanStatus,
				},
				{
					RepoRelDir: ".",
					Workspace:  "staging",
					Status:     models.PlannedPlanStatus,
				},
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
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			tmpDir := t.TempDir()

			backend := lockingMocks.NewMockBackend()
			workingDir := mocks.NewMockWorkingDir()

			modelPull := models.PullRequest{
				BaseRepo: testdata.GithubRepo,
				State:    models.OpenPullState,
				Num:      testdata.Pull.Num,
			}

			pullStatusMock := models.PullStatus{
				Projects: c.projectStatus,
				Pull:     modelPull,
			}

			When(backend.GetPullStatus(modelPull)).ThenReturn(&pullStatusMock, nil)
			When(workingDir.GetPullDir(modelPull.BaseRepo, modelPull)).ThenReturn(tmpDir, nil)

			pf := &events.DefaultPendingPlanFinder{
				backend,
				workingDir,
			}

			actPlans, err := pf.Find(modelPull)
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

	modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}

	backend := lockingMocks.NewMockBackend()
	workingDir := mocks.NewMockWorkingDir()

	pendingPlans := models.PullStatus{
		Projects: []models.ProjectStatus{
			{
				RepoRelDir:  "dir1",
				Workspace:   "default",
				ProjectName: "",
				Status:      models.PlannedPlanStatus,
			},
			{
				RepoRelDir:  "dir2",
				Workspace:   "default",
				ProjectName: "",
				Status:      models.PlannedPlanStatus,
			},
		},
		Pull: modelPull,
	}

	When(backend.GetPullStatus(modelPull)).ThenReturn(&pendingPlans, nil)
	When(workingDir.GetPullDir(modelPull.BaseRepo, modelPull)).ThenReturn(tmp, nil)

	pf := &events.DefaultPendingPlanFinder{
		Backend:    backend,
		WorkingDir: workingDir,
	}
	Ok(t, pf.DeletePlans(modelPull))

	// First, check the files were deleted.
	for _, plan := range []string{
		"dir1/default.tfplan",
		"dir2/default.tfplan",
	} {
		absPath := filepath.Join(tmp, plan)
		_, err := os.Stat(absPath)
		ErrContains(t, "no such file or directory", err)
	}
}

func runCmd(t *testing.T, dir string, name string, args ...string) string {
	t.Helper()
	cpCmd := exec.Command(name, args...)
	cpCmd.Dir = dir
	cpOut, err := cpCmd.CombinedOutput()
	Assert(t, err == nil, "err running %q: %s", strings.Join(append([]string{name}, args...), " "), cpOut)
	return string(cpOut)
}
