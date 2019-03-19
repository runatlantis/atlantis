// Package runtime holds code for actually running commands vs. preparing
// and constructing.
package runtime

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/terraform"
	"github.com/runatlantis/atlantis/server/logging"
)

// lineBeforeRunURL is the line output during a remote operation right before
// a link to the run url will be output.
const lineBeforeRunURL = "To view this run in a browser, visit:"

// TerraformExec brings the interface from TerraformClient into this package
// without causing circular imports.
type TerraformExec interface {
	RunCommandWithVersion(log *logging.SimpleLogger, path string, args []string, v *version.Version, workspace string) (string, error)
}

// AsyncTFExec brings the interface from TerraformClient into this package
// without causing circular imports.
// It's split from TerraformExec because due to a bug in pegomock with channels,
// we can't generate a mock for it so we hand-write it for this specific method.
type AsyncTFExec interface {
	// RunCommandAsync runs terraform with args. It immediately returns an
	// input and output channel. Callers can use the output channel to
	// get the realtime output from the command.
	// Callers can use the input channel to pass stdin input to the command.
	// If any error is passed on the out channel, there will be no
	// further output (so callers are free to exit).
	RunCommandAsync(log *logging.SimpleLogger, path string, args []string, v *version.Version, workspace string) (chan<- string, <-chan terraform.Line)
}

// StatusUpdater brings the interface from CommitStatusUpdater into this package
// without causing circular imports.
type StatusUpdater interface {
	UpdateProject(ctx models.ProjectCommandContext, cmdName models.CommandName, status models.CommitStatus, url string) error
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

// GetPlanFilename returns the filename (not the path) of the generated tf plan
// given a workspace and project name.
func GetPlanFilename(workspace string, projName string) string {
	var unescapedFilename string
	if projName == "" {
		unescapedFilename = fmt.Sprintf("%s.tfplan", workspace)
	} else {
		unescapedFilename = fmt.Sprintf("%s-%s.tfplan", projName, workspace)
	}
	return invalidFilenameChars.ReplaceAllLiteralString(unescapedFilename, "-")
}
