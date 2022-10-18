package terraform

import (
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
)

type Request struct {
	Root         terraform.Root
	Repo         github.Repo
	DeploymentID string
	Revision     string
}
