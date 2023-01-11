package events

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/runtime"
)

//go:generate pegomock generate -m --package mocks -o mocks/mock_pending_plan_finder.go PendingPlanFinder

type PendingPlanFinder interface {
	Find(pullDir string) ([]PendingPlan, error)
	DeletePlans(pullDir string) error
}

// DefaultPendingPlanFinder finds unapplied plans.
type DefaultPendingPlanFinder struct{}

// PendingPlan is a plan that has not been applied.
type PendingPlan struct {
	// RepoDir is the absolute path to the root of the repo that holds this
	// plan.
	RepoDir string
	// RepoRelDir is the relative path from the repo to the project that
	// the plan is for.
	RepoRelDir string
	// Workspace is the workspace this plan should execute in.
	Workspace   string
	ProjectName string
}

// Find finds all pending plans in pullDir. pullDir should be the working
// directory where Atlantis will operate on this pull request. It's one level
// up from where Atlantis clones the repo for each workspace.
func (p *DefaultPendingPlanFinder) Find(pullDir string) ([]PendingPlan, error) {
	plans, _, err := p.findWithAbsPaths(pullDir)
	return plans, err
}

// deletePlans deletes all plans in pullDir.
func (p *DefaultPendingPlanFinder) DeletePlans(pullDir string) error {
	_, absPaths, err := p.findWithAbsPaths(pullDir)
	if err != nil {
		return err
	}
	for _, path := range absPaths {
		if err := os.Remove(path); err != nil {
			return errors.Wrapf(err, "delete plan at %s", path)
		}
	}
	return nil
}

func (p *DefaultPendingPlanFinder) findWithAbsPaths(pullDir string) ([]PendingPlan, []string, error) {
	var plans []PendingPlan
	var absPaths []string

	_, err := os.ReadDir(pullDir)
	if err != nil {
		return nil, nil, err
	}

	lsCmd := exec.Command("git", "ls-files", ".", "--others") // nolint: gosec
	lsCmd.Dir = pullDir
	lsOut, err := lsCmd.CombinedOutput()
	if err != nil {
		return nil, nil, errors.Wrapf(err, "running git ls-files . "+
			"--others: %s", string(lsOut))
	}

	for _, file := range strings.Split(string(lsOut), "\n") {
		if filepath.Ext(file) == ".tfplan" {
			// Ignore .terragrunt-cache dirs (#487)
			if strings.Contains(file, ".terragrunt-cache/") {
				continue
			}

			// Fix: This will return an invalid workspace if
			// either the project name or workspace has a '-'.
			workspace := strings.Split(filepath.Base(file), ".")[0]

			projectName, err := runtime.ProjectNameFromPlanfile(workspace, filepath.Base(file))
			if err != nil {
				return nil, nil, err
			}
			plans = append(plans, PendingPlan{
				RepoDir:     pullDir,
				RepoRelDir:  filepath.Dir(file),
				Workspace:   workspace,
				ProjectName: projectName,
			})
			absPaths = append(absPaths, filepath.Join(pullDir, file))
		}
	}

	return plans, absPaths, nil
}
