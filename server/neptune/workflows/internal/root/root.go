package root

import (
	"path/filepath"

	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/job"
)

type Root struct {
	Name      string
	Path      string
	TfVersion string
	Apply     job.Job
	Plan      job.Job
}

// Root Instance is a root at a certain commit with the repo info
type RootInstance struct {
	Root Root
	Repo github.RepoInstance
}

func (r *RootInstance) RelativePathFromRepo() (string, error) {
	return filepath.Rel(r.Repo.Path, r.Root.Path)
}

func BuildRootInstanceFrom(root Root, repo github.RepoInstance) *RootInstance {
	return &RootInstance{
		Root: root,
		Repo: repo,
	}
}
