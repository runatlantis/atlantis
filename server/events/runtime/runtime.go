// Package runtime holds code for actually running commands vs. preparing
// and constructing.
package runtime

import (
	"fmt"
	"github.com/runatlantis/atlantis/server/events/models"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
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
	// output and error channel that callers can use to follow the progress
	// of the command. 1 or 0 errors will be sent on the error channel.
	// When the command is complete, both channels will be closed.
	RunCommandAsync(log *logging.SimpleLogger, path string, args []string, v *version.Version, workspace string) (<-chan string, <-chan error)
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
// given a workspace and maybe a project's config.
func GetPlanFilename(workspace string, maybeCfg *valid.Project) string {
	var unescapedFilename string
	if maybeCfg == nil || maybeCfg.Name == nil {
		unescapedFilename = fmt.Sprintf("%s.tfplan", workspace)
	} else {
		unescapedFilename = fmt.Sprintf("%s-%s.tfplan", *maybeCfg.Name, workspace)
	}
	return invalidFilenameChars.ReplaceAllLiteralString(unescapedFilename, "-")
}

// runRemoteOp runs a terraform command that utilizes the remote operations
// backend. It watches the command output for the run url to be printed, and
// then updates the commit status with a link to the run url.
// The run url is a link to the Terraform Enterprise UI where the output
// from the in-progress command can be viewed.
// cmdArgs is the args to terraform to execute.
// path is the path to where we need to execute.
func runRemoteOp(
	ctx models.ProjectCommandContext,
	cmd models.CommandName,
	cmdArgs []string,
	path string,
	tfVersion *version.Version,
	executor AsyncTFExec,
	statusUpdater StatusUpdater) (string, error) {

	// updateStatus will update the commit status and log any error.
	updateStatus := func(status models.CommitStatus, url string) {
		if err := statusUpdater.UpdateProject(ctx, cmd, status, url); err != nil {
			ctx.Log.Err("unable to update status: %s", err)
		}
	}

	// Start the async command execution.
	ctx.Log.Debug("starting async tf remote operation")
	outputChan, errChan := executor.RunCommandAsync(ctx.Log, filepath.Clean(path), cmdArgs, tfVersion, ctx.Workspace)
	var lines []string
	nextLineIsRunURL := false
	var runURL string
	var err error

	// This code reads off of both the output and error channels. When each
	// channel is closed, we set it to nil so the select won't block but we
	// can still read from the other channel if it isn't closed yet.
	for {

		// Inside the select, we're setting our channels to nil when they close
		// so this will be our exit condition when they're both closed.
		if outputChan == nil && errChan == nil {
			break
		}

		select {
		case line, ok := <-outputChan:
			if ok {
				lines = append(lines, line)
				// Here we're checking for the run url and updating the status
				// if found.
				if line == lineBeforeRunURL {
					nextLineIsRunURL = true
				} else if nextLineIsRunURL {
					runURL = strings.TrimSpace(line)
					ctx.Log.Debug("remote run url found, updating commit status")
					updateStatus(models.PendingCommitStatus, runURL)
					nextLineIsRunURL = false
				}
			} else {
				// outputChan has closed. Set it to nil so the select doesn't
				// block (select won't block on nil channels).
				outputChan = nil
			}
		case e, ok := <-errChan:
			if ok {
				err = e
			} else {
				// errChan has closed. Set it to nil so the select doesn't
				// block (select won't block on nil channels).
				errChan = nil
			}
		}
	}

	ctx.Log.Debug("async tf remote operation complete")
	output := strings.Join(lines, "\n")
	if err != nil {
		updateStatus(models.FailedCommitStatus, runURL)
	} else {
		updateStatus(models.SuccessCommitStatus, runURL)
	}
	return output, err
}
