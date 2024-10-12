package cmd

import (
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

// CheckSkipCI checks if the latest commit message contains the "[skip ci]" flag.
func CheckSkipCI() bool {
	// Fetch the latest commit message
	cmd := exec.Command("git", "log", "-1", "--pretty=%B")
	output, err := cmd.Output()
	if err != nil {
		errors.Wrap(err, "failed to fetch the latest commit message")
	}
	commitMessage := string(output)

	// Check if the commit message contains "[skip ci]"
	return strings.Contains(commitMessage, "[skip ci]")
}
