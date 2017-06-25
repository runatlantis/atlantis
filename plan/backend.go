package plan

import (
	"github.com/hootsuite/atlantis/models"
)

type Backend interface {
	SavePlan(path string, project models.Project, env string, pullNum int) error
	CopyPlans(dstRepoPath string, repoFullName string, env string, pullNum int) ([]Plan, error)
	DeletePlan(project models.Project, env string, pullNum int) error
	DeletePlansByPull(repoFullName string, pullNum int) error
}

type Plan struct {
	Project models.Project
	// LocalPath is the path to the plan on disk
	LocalPath string
}
