package events_test

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

var logger *logging.SimpleLogger

// Test that we write the file as expected
func TestWriteGitCreds_WriteFile(t *testing.T) {
	tmp, cleanup := TempDir(t)
	defer cleanup()

	err := events.WriteGitCreds("user", "token", "hostname", tmp, logger, false)
	Ok(t, err)

	expContents := `https://user:token@hostname`

	actContents, err := ioutil.ReadFile(filepath.Join(tmp, ".git-credentials"))
	Ok(t, err)
	Equals(t, expContents, string(actContents))
}

// Test that if the file already exists and it doesn't have the line we would
// have written, we write it.
func TestWriteGitCreds_Appends(t *testing.T) {
	tmp, cleanup := TempDir(t)
	defer cleanup()

	credsFile := filepath.Join(tmp, ".git-credentials")
	err := ioutil.WriteFile(credsFile, []byte("contents"), 0600)
	Ok(t, err)

	err = events.WriteGitCreds("user", "token", "hostname", tmp, logger, false)
	Ok(t, err)

	expContents := "contents\nhttps://user:token@hostname"
	actContents, err := ioutil.ReadFile(filepath.Join(tmp, ".git-credentials"))
	Ok(t, err)
	Equals(t, expContents, string(actContents))
}

// Test that if the file already exists and it already has the line expected
// we do nothing.
func TestWriteGitCreds_NoModification(t *testing.T) {
	tmp, cleanup := TempDir(t)
	defer cleanup()

	credsFile := filepath.Join(tmp, ".git-credentials")
	contents := "line1\nhttps://user:token@hostname\nline2"
	err := ioutil.WriteFile(credsFile, []byte(contents), 0600)
	Ok(t, err)

	err = events.WriteGitCreds("user", "token", "hostname", tmp, logger, false)
	Ok(t, err)
	actContents, err := ioutil.ReadFile(filepath.Join(tmp, ".git-credentials"))
	Ok(t, err)
	Equals(t, contents, string(actContents))
}

// Test that the github app credentials get replaced.
func TestWriteGitCreds_ReplaceApp(t *testing.T) {
	tmp, cleanup := TempDir(t)
	defer cleanup()

	credsFile := filepath.Join(tmp, ".git-credentials")
	contents := "line1\nhttps://x-access-token:v1.87dddddddddddddddd@github.com\nline2"
	err := ioutil.WriteFile(credsFile, []byte(contents), 0600)
	Ok(t, err)

	err = events.WriteGitCreds("x-access-token", "token", "github.com", tmp, logger, true)
	Ok(t, err)
	expContets := "line1\nhttps://x-access-token:token@github.com\nline2"
	actContents, err := ioutil.ReadFile(filepath.Join(tmp, ".git-credentials"))
	Ok(t, err)
	Equals(t, expContets, string(actContents))
}

// Test that if we can't read the existing file to see if the contents will be
// the same that we just error out.
func TestWriteGitCreds_ErrIfCannotRead(t *testing.T) {
	tmp, cleanup := TempDir(t)
	defer cleanup()

	credsFile := filepath.Join(tmp, ".git-credentials")
	err := ioutil.WriteFile(credsFile, []byte("can't see me!"), 0000)
	Ok(t, err)

	expErr := fmt.Sprintf("open %s: permission denied", credsFile)
	actErr := events.WriteGitCreds("user", "token", "hostname", tmp, logger, false)
	ErrContains(t, expErr, actErr)
}

// Test that if we can't write, we error out.
func TestWriteGitCreds_ErrIfCannotWrite(t *testing.T) {
	credsFile := "/this/dir/does/not/exist/.git-credentials"
	expErr := fmt.Sprintf("writing generated .git-credentials file with user, token and hostname to %s: open %s: no such file or directory", credsFile, credsFile)
	actErr := events.WriteGitCreds("user", "token", "hostname", "/this/dir/does/not/exist", logger, false)
	ErrEquals(t, expErr, actErr)
}

// Test that git is actually configured to use the credentials
func TestWriteGitCreds_ConfigureGitCredentialHelper(t *testing.T) {
	tmp, cleanup := TempDir(t)
	defer cleanup()

	err := events.WriteGitCreds("user", "token", "hostname", tmp, logger, false)
	Ok(t, err)

	expOutput := `store`
	actOutput, err := exec.Command("git", "config", "--global", "credential.helper").Output()
	Ok(t, err)
	Equals(t, expOutput+"\n", string(actOutput))
}

// Test that git is configured to use https instead of ssh
func TestWriteGitCreds_ConfigureGitUrlOverride(t *testing.T) {
	tmp, cleanup := TempDir(t)
	defer cleanup()

	err := events.WriteGitCreds("user", "token", "hostname", tmp, logger, false)
	Ok(t, err)

	expOutput := `ssh://git@hostname`
	actOutput, err := exec.Command("git", "config", "--global", "url.https://user@hostname.insteadof").Output()
	Ok(t, err)
	Equals(t, expOutput+"\n", string(actOutput))
}
