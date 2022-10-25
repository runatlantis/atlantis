package terraform

import (
	"fmt"

	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/deployment"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"

	"github.com/google/uuid"
)

type DeploymentInfo struct {
	ID             uuid.UUID
	CheckRunID     int64
	Revision       string
	InitiatingUser github.User
	Root           terraform.Root
	Repo           github.Repo
	Tags           map[string]string
}

func (i DeploymentInfo) BuildPersistableInfo() *deployment.Info {
	return &deployment.Info{
		Version:  deployment.InfoSchemaVersion,
		ID:       i.ID.String(),
		Revision: i.Revision,
		Root: deployment.Root{
			Name:    i.Root.Name,
			Trigger: string(i.Root.Trigger),
		},
		Repo: deployment.Repo{
			Name:  i.Repo.Name,
			Owner: i.Repo.Owner,
		},
	}
}

func BuildCheckRunTitle(rootName string) string {
	return fmt.Sprintf("atlantis/deploy: %s", rootName)
}
