package terraform

import (
	"github.com/google/uuid"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/root"
)

type DeploymentInfo struct {
	ID         uuid.UUID
	CheckRunID int64
	Revision   string
	Root       root.Root
}
