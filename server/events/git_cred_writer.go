package events

import (
	"fmt"
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
// If ghAccessToken is true we will look for a line starting with https://x-access-token and ending with gitHostname and replace it.
func WriteGitCreds(gitUser string, gitToken string, gitHostname string, home string, logger logging.SimpleLogging, ghAccessToken bool) error {
	const credsFilename = ".git-credentials"
	credsFile := filepath.Join(home, credsFilename)
	credsFileContentsPattern := `https://%s:%s@%s`
	config := fmt.Sprintf(credsFileContentsPattern, gitUser, gitToken, gitHostname)

	// If the file doesn't exist, write it.
	if _, err := os.Stat(credsFile); err != nil {
		if err := os.WriteFile(credsFile, []byte(config), 0600); err != nil {
			return errors.Wrapf(err, "writing generated %s file with user, token and hostname to %s", credsFilename, credsFile)
		}
		logger.Info("wrote git credentials to %s", credsFile)
	} else {
		hasLine, err := fileHasLine(config, credsFile)
		if err != nil {
			return err
		}
		if hasLine {
			logger.Debug("git credentials file has expected contents, not modifying")
			return nil
		}

		if ghAccessToken {
			// Need to replace the line.
			if err := fileLineReplace(config, gitUser, gitHostname, credsFile); err != nil {
				return errors.Wrap(err, "replacing git credentials line for github app")
			}
			logger.Info("updated git app credentials in %s", credsFile)
		} else {
			// Otherwise we need to append the line.
			if err := fileAppend(config, credsFile); err != nil {
				return err
			}
			logger.Info("wrote git credentials to %s", credsFile)
		}
	}

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

func fileHasLine(line string, filename string) (bool, error) {
	currContents, err := os.ReadFile(filename) // nolint: gosec
	if err != nil {
		return false, errors.Wrapf(err, "reading %s", filename)
	}
	for _, l := range strings.Split(string(currContents), "\n") {
		if l == line {
			return true, nil
		}
	}
	return false, nil
}

func fileAppend(line string, filename string) error {
	currContents, err := os.ReadFile(filename) // nolint: gosec
	if err != nil {
		return err
	}
	if len(currContents) > 0 && !strings.HasSuffix(string(currContents), "\n") {
		line = "\n" + line
	}
	return os.WriteFile(filename, []byte(string(currContents)+line), 0600)
}

func fileLineReplace(line, user, host, filename string) error {
	currContents, err := os.ReadFile(filename) // nolint: gosec
	if err != nil {
		return err
	}
	prevLines := strings.Split(string(currContents), "\n")
	var newLines []string
	for _, l := range prevLines {
		if strings.HasPrefix(l, "https://"+user) && strings.HasSuffix(l, host) {
			newLines = append(newLines, line)
		} else {
			newLines = append(newLines, l)
		}
	}
	toWrite := strings.Join(newLines, "\n")

	// there was nothing to replace so we need to append the creds
	if toWrite == "" {
		return fileAppend(line, filename)
	}

	return os.WriteFile(filename, []byte(toWrite), 0600)
}
