// Package runtime handles constructing an execution graph for each action
// based on configuration and defaults. The handlers can then execute this
// graph.
package runtime

import (
	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/logging"
)

type TerraformExec interface {
	RunCommandWithVersion(log *logging.SimpleLogger, path string, args []string, v *version.Version, workspace string) (string, error)
}

// MustConstraint returns a constraint. It panics on error.
func MustConstraint(constraint string) version.Constraints {
	c, err := version.NewConstraint(constraint)
	if err != nil {
		panic(err)
	}
	return c
}
