package terraform

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/root"
)

type DeploymentInfo struct {
	ID         uuid.UUID
	CheckRunID int64
	Revision   string
	Root       root.Root
}

func BuildCheckRunTitle(rootName string) string {
	return fmt.Sprintf("atlantis/deploy: %s", rootName)
}
