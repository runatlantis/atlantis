package cloud

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// rcFileContents is a format string to be used with Sprintf that can be used
// to generate the contents of a ~/.terraformrc file for authenticating with
// Terraform Enterprise.
var rcFileContents = `credentials "%s" {
  token = %q
}`

const rcFilename = ".terraformrc"

// GenerateConfigFile generates a .terraformrc file containing config for tfeToken
// and hostname tfeHostname.
// It will create the file in home/.terraformrc.
func GenerateConfigFile(tfeToken string, tfeHostname string, home string) error {
	config := fmt.Sprintf(rcFileContents, tfeHostname, tfeToken)
	rcFile := filepath.Join(home, rcFilename)

	// If there is already a .terraformrc file and its contents aren't exactly
	// what we would have written to it, then we error out because we don't
	// want to overwrite anything.
	if _, err := os.Stat(rcFile); err == nil {
		currContents, err := os.ReadFile(rcFile) // nolint: gosec
		if err != nil {
			return errors.Wrapf(err, "trying to read %s to ensure we're not overwriting it", rcFile)
		}
		if config != string(currContents) {
			return fmt.Errorf("can't write TFE token to %s because that file has contents that would be overwritten", rcFile)
		}
		// Otherwise we don't need to write the file because it already has
		// what we need.
		return nil
	}

	if err := os.WriteFile(rcFile, []byte(config), 0600); err != nil {
		return errors.Wrapf(err, "writing generated %s file with TFE token to %s", rcFilename, rcFile)
	}
	return nil
}
