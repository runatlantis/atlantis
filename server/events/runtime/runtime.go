// Package runtime holds code for actually running commands vs. preparing
// and constructing.
package runtime

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
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

// invalidFilenameChars matches chars that are invalid for linux and windows
// filenames.
// From https://www.oreilly.com/library/view/regular-expressions-cookbook/9781449327453/ch08s25.html
var invalidFilenameChars = regexp.MustCompile(`[\\/:"*?<>|]`)

// GetProjectFilenamePrefix returns the filename prefix given a workspace and
// maybe a project's config.
func GetProjectFilenamePrefix(workspace string, maybeCfg *valid.Project) string {
	var unescapedFilename string
	if maybeCfg == nil || maybeCfg.Name == nil {
		unescapedFilename = workspace
	} else {
		unescapedFilename = fmt.Sprintf("%s-%s", *maybeCfg.Name, workspace)
	}
	return invalidFilenameChars.ReplaceAllLiteralString(unescapedFilename, "-")
}

// GetPlanFilename returns the filename (not the path) of the generated tf plan
// given a workspace and maybe a project's config.
func GetPlanFilename(workspace string, maybeCfg *valid.Project) string {
	return GetProjectFilenamePrefix(workspace, maybeCfg) + ".tfplan"
}
