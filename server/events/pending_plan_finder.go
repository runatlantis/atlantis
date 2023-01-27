package events

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/events/models"
)

//go:generate pegomock generate -m --package mocks -o mocks/mock_pending_plan_finder.go PendingPlanFinder

type PendingPlanFinder interface {
	Find(pullRequest models.PullRequest) ([]PendingPlan, error)
	DeletePlans(pullRequest models.PullRequest) error
}

// DefaultPendingPlanFinder finds unapplied plans.
type DefaultPendingPlanFinder struct {
	Backend    locking.Backend
	WorkingDir WorkingDir
}

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
func (p *DefaultPendingPlanFinder) Find(pullRequest models.PullRequest) ([]PendingPlan, error) {
	plans, err := p.findWithAbsPaths(pullRequest)
	return plans, err
}

// deletePlans deletes all plans in pullDir.
func (p *DefaultPendingPlanFinder) DeletePlans(pullRequest models.PullRequest) error {
	plans, err := p.findWithAbsPaths(pullRequest)
	if err != nil {
		return err
	}

	for _, plan := range plans {
		path := fmt.Sprintf(
			"%s/%s/%s",
			plan.RepoDir,
			plan.RepoRelDir,
			runtime.GetPlanFilename(plan.Workspace, plan.ProjectName),
		)

		if err := os.Remove(path); err != nil {
			return errors.Wrapf(err, "delete plan at %s", path)
		}

	}

	return nil
}

func (p *DefaultPendingPlanFinder) findWithAbsPaths(pullRequest models.PullRequest) ([]PendingPlan, error) {
	var plans []PendingPlan

	pullDir, err := p.WorkingDir.GetPullDir(pullRequest.BaseRepo, pullRequest)
	if err != nil {
		return plans, err
	}

	status, err := p.Backend.GetPullStatus(pullRequest)
	if err != nil || status == nil {
		return plans, err
	}

	for _, project := range status.Projects {
		if !project.Status.IsPlanned() {
			continue
		}

		plans = append(plans, PendingPlan{
			RepoDir:     pullDir,
			RepoRelDir:  project.RepoRelDir,
			Workspace:   project.Workspace,
			ProjectName: project.ProjectName,
		})
	}

	return plans, nil
}
