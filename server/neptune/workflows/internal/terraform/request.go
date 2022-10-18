package terraform

import (
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities/terraform"
)

type Request struct {
	Root         terraform.Root
	Repo         github.Repo
	DeploymentID string
	Revision     string
}
