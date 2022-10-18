package deployment

import (
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/root"
)

type Info struct {
	Version    string
	ID         string
	CheckRunID int64
	Revision   string
	Repo       github.Repo
	Root       root.Root
}
