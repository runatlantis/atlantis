package events

import (
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// PendingPlanFinder finds unapplied plans.
type PendingPlanFinder struct{}

// PendingPlan is a plan that has not been applied.
type PendingPlan struct {
	// RepoDir is the absolute path to the root of the repo that holds this
	// plan.
	RepoDir string
	// RepoRelDir is the relative path from the repo to the project that
	// the plan is for.
	RepoRelDir string
	// Workspace is the workspace this plan should execute in.
	Workspace string
}

// Find finds all pending plans in pullDir. pullDir should be the working
// directory where Atlantis will operate on this pull request. It's one level
// up from where Atlantis clones the repo for each workspace.
func (p *PendingPlanFinder) Find(pullDir string) ([]PendingPlan, error) {
	workspaceDirs, err := ioutil.ReadDir(pullDir)
	if err != nil {
		return nil, err
	}
	var plans []PendingPlan
	for _, workspaceDir := range workspaceDirs {
		workspace := workspaceDir.Name()
		repoDir := filepath.Join(pullDir, workspace)

		// Any generated plans should be untracked by git since Atlantis created
		// them.
		lsCmd := exec.Command("git", "ls-files", ".", "--others") // nolint: gosec
		lsCmd.Dir = repoDir
		lsOut, err := lsCmd.CombinedOutput()
		if err != nil {
			return nil, errors.Wrapf(err, "running git ls-files . "+
				"--others: %s", string(lsOut))
		}
		for _, file := range strings.Split(string(lsOut), "\n") {
			if filepath.Ext(file) == ".tfplan" {
				repoRelDir := filepath.Dir(file)
				plans = append(plans, PendingPlan{
					RepoDir:    repoDir,
					RepoRelDir: repoRelDir,
					Workspace:  workspace,
				})
			}
		}
	}
	return plans, nil
}

// FindErrors finds all failed init or plans in pullDir. pullDir should be the
// working directory where Atlantis will operate on this pull request. It's one
// level up from where Atlantis clones the repo for each workspace.
func (p *PendingPlanFinder) FindErrors(pullDir string) ([]PendingPlan, error) {
	workspaceDirs, err := ioutil.ReadDir(pullDir)
	if err != nil {
		return nil, err
	}
	var plans []PendingPlan
	for _, workspaceDir := range workspaceDirs {
		workspace := workspaceDir.Name()
		repoDir := filepath.Join(pullDir, workspace)

		// Any generated plans should be untracked by git since Atlantis created
		// them.
		lsCmd := exec.Command("git", "ls-files", ".", "--others") // nolint: gosec
		lsCmd.Dir = repoDir
		lsOut, err := lsCmd.CombinedOutput()
		if err != nil {
			return nil, errors.Wrapf(err, "running git ls-files . "+
				"--others: %s", string(lsOut))
		}
		for _, file := range strings.Split(string(lsOut), "\n") {
			if filepath.Ext(file) == ".tfinit-error" || filepath.Ext(file) == ".tfplan-error" {
				repoRelDir := filepath.Dir(file)
				plans = append(plans, PendingPlan{
					RepoDir:    repoDir,
					RepoRelDir: repoRelDir,
					Workspace:  workspace,
				})
			}
		}
	}
	return plans, nil
}
