package vcs_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

// Test that we write the file as expected
func TestWriteGitCreds_WriteFile(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	err := vcs.WriteGitCreds("user", "token", "hostname", tmp, logger, false)
	Ok(t, err)

	expContents := `https://user:token@hostname`

	actContents, err := os.ReadFile(filepath.Join(tmp, ".git-credentials"))
	Ok(t, err)
	Equals(t, expContents, string(actContents))
}

// Test that if the file already exists and it doesn't have the line we would
// have written, we write it.
func TestWriteGitCreds_Appends(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	credsFile := filepath.Join(tmp, ".git-credentials")
	err := os.WriteFile(credsFile, []byte("contents"), 0600)
	Ok(t, err)

	err = vcs.WriteGitCreds("user", "token", "hostname", tmp, logger, false)
	Ok(t, err)

	expContents := "contents\nhttps://user:token@hostname"
	actContents, err := os.ReadFile(filepath.Join(tmp, ".git-credentials"))
	Ok(t, err)
	Equals(t, expContents, string(actContents))
}

// Test that if the file already exists and it already has the line expected
// we do nothing.
func TestWriteGitCreds_NoModification(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	credsFile := filepath.Join(tmp, ".git-credentials")
	contents := "line1\nhttps://user:token@hostname\nline2"
	err := os.WriteFile(credsFile, []byte(contents), 0600)
	Ok(t, err)

	err = vcs.WriteGitCreds("user", "token", "hostname", tmp, logger, false)
	Ok(t, err)
	actContents, err := os.ReadFile(filepath.Join(tmp, ".git-credentials"))
	Ok(t, err)
	Equals(t, contents, string(actContents))
}

// Test that the github app credentials get replaced.
func TestWriteGitCreds_ReplaceApp(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	credsFile := filepath.Join(tmp, ".git-credentials")
	contents := "line1\nhttps://x-access-token:v1.87dddddddddddddddd@github.com\nline2"
	err := os.WriteFile(credsFile, []byte(contents), 0600)
	Ok(t, err)

	err = vcs.WriteGitCreds("x-access-token", "token", "github.com", tmp, logger, true)
	Ok(t, err)
	expContets := "line1\nhttps://x-access-token:token@github.com\nline2"
	actContents, err := os.ReadFile(filepath.Join(tmp, ".git-credentials"))
	Ok(t, err)
	Equals(t, expContets, string(actContents))
}

// Test that the github app credentials get updated when cred file is empty.
func TestWriteGitCreds_AppendApp(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	credsFile := filepath.Join(tmp, ".git-credentials")
	contents := ""
	err := os.WriteFile(credsFile, []byte(contents), 0600)
	Ok(t, err)

	err = vcs.WriteGitCreds("x-access-token", "token", "github.com", tmp, logger, true)
	Ok(t, err)
	expContets := "https://x-access-token:token@github.com"
	actContents, err := os.ReadFile(filepath.Join(tmp, ".git-credentials"))
	Ok(t, err)
	Equals(t, expContets, string(actContents))
}

// Test that if we can't read the existing file to see if the contents will be
// the same that we just error out.
func TestWriteGitCreds_ErrIfCannotRead(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	credsFile := filepath.Join(tmp, ".git-credentials")
	err := os.WriteFile(credsFile, []byte("can't see me!"), 0000)
	Ok(t, err)

	expErr := fmt.Sprintf("open %s: permission denied", credsFile)
	actErr := vcs.WriteGitCreds("user", "token", "hostname", tmp, logger, false)
	ErrContains(t, expErr, actErr)
}

// Test that if we can't write, we error out.
func TestWriteGitCreds_ErrIfCannotWrite(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	credsFile := "/this/dir/does/not/exist/.git-credentials" // nolint: gosec
	expErr := fmt.Sprintf("writing generated .git-credentials file with user, token and hostname to %s: open %s: no such file or directory", credsFile, credsFile)
	actErr := vcs.WriteGitCreds("user", "token", "hostname", "/this/dir/does/not/exist", logger, false)
	ErrEquals(t, expErr, actErr)
}

// Test that git is actually configured to use the credentials
func TestWriteGitCreds_ConfigureGitCredentialHelper(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	err := vcs.WriteGitCreds("user", "token", "hostname", tmp, logger, false)
	Ok(t, err)

	expOutput := `store`
	actOutput, err := exec.Command("git", "config", "--global", "credential.helper").Output()
	Ok(t, err)
	Equals(t, expOutput+"\n", string(actOutput))
}

// Test that git is configured to use https instead of ssh
func TestWriteGitCreds_ConfigureGitUrlOverride(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	err := vcs.WriteGitCreds("user", "token", "hostname", tmp, logger, false)
	Ok(t, err)

	expOutput := `ssh://git@hostname`
	actOutput, err := exec.Command("git", "config", "--global", "url.https://user@hostname.insteadof").Output()
	Ok(t, err)
	Equals(t, expOutput+"\n", string(actOutput))
}
