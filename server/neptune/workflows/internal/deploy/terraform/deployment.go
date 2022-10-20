package terraform

import (
	"fmt"

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

func BuildCheckRunTitle(rootName string) string {
	return fmt.Sprintf("atlantis/deploy: %s", rootName)
}
