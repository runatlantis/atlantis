package events

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/logging"
)

// WriteGitCreds generates a .git-credentials file containing the username and token
// used for authenticating with git over HTTPS
// It will create the file in home/.git-credentials
func WriteGitCreds(gitUser string, gitToken string, gitHostname string, home string, logger *logging.SimpleLogger) error {
	const credsFilename = ".git-credentials"
	credsFile := filepath.Join(home, credsFilename)
	credsFileContents := `https://%s:%s@%s`
	config := fmt.Sprintf(credsFileContents, gitUser, gitToken, gitHostname)

	// If there is already a .git-credentials file and its contents aren't exactly
	// what we would have written to it, then we error out because we don't
	// want to overwrite anything
	if _, err := os.Stat(credsFile); err == nil {
		currContents, err := ioutil.ReadFile(credsFile) // nolint: gosec
		if err != nil {
			return errors.Wrapf(err, "trying to read %s to ensure we're not overwriting it", credsFile)
		}
		if config != string(currContents) {
			return fmt.Errorf("can't write git-credentials to %s because that file has contents that would be overwritten", credsFile)
		}
		// Otherwise we don't need to write the file because it already has
		// what we need.
		return nil
	}

	if err := ioutil.WriteFile(credsFile, []byte(config), 0600); err != nil {
		return errors.Wrapf(err, "writing generated %s file with user, token and hostname to %s", credsFilename, credsFile)
	}

	logger.Info("wrote git credentials to %s", credsFile)

	credentialCmd := exec.Command("git", "config", "--global", "credential.helper", "store")
	if out, err := credentialCmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "There was an error running %s: %s", strings.Join(credentialCmd.Args, " "), string(out))
	}
	logger.Info("successfully ran %s", strings.Join(credentialCmd.Args, " "))

	urlCmd := exec.Command("git", "config", "--global", fmt.Sprintf("url.https://%s@%s.insteadOf", gitUser, gitHostname), fmt.Sprintf("ssh://git@%s", gitHostname)) // nolint: gosec
	if out, err := urlCmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "There was an error running %s: %s", strings.Join(urlCmd.Args, " "), string(out))
	}
	logger.Info("successfully ran %s", strings.Join(urlCmd.Args, " "))
	return nil
}
