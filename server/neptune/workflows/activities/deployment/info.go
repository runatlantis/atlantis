package deployment

import (
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
)

type Info struct {
	Version    string
	ID         string
	CheckRunID int64
	Revision   string
	Repo       github.Repo
	Root       terraform.Root
}
