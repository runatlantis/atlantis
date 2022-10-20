package deployment

import (
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
)

const InfoSchemaVersion = "1.0.0"

type Info struct {
	Version        string
	ID             string
	CheckRunID     int64
	Revision       string
	InitiatingUser github.User
	Repo           github.Repo
	Root           terraform.Root
	Tags           map[string]string
}
