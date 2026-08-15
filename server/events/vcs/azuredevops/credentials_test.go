// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package azuredevops_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/runatlantis/atlantis/server/events/vcs/azuredevops"
	. "github.com/runatlantis/atlantis/testing"
)

func TestPATCredentials_GetUser(t *testing.T) {
	creds := &azuredevops.PATCredentials{User: "atlantis"}
	user, err := creds.GetUser()
	Ok(t, err)
	Equals(t, "atlantis", user)
}

func TestPATCredentials_GetToken_Static(t *testing.T) {
	creds := &azuredevops.PATCredentials{User: "atlantis", Token: "some-token"}
	token, err := creds.GetToken()
	Ok(t, err)
	Equals(t, "some-token", token)
}

func TestPATCredentials_GetToken_FromFile(t *testing.T) {
	tmpDir := t.TempDir()
	tokenFile := filepath.Join(tmpDir, "token")
	Ok(t, os.WriteFile(tokenFile, []byte("first-token\n"), 0600))

	creds := &azuredevops.PATCredentials{User: "atlantis", TokenFile: tokenFile}

	token, err := creds.GetToken()
	Ok(t, err)
	Equals(t, "first-token", token)

	// An externally-rotated token is picked up on the next read without
	// recreating the credentials.
	Ok(t, os.WriteFile(tokenFile, []byte("  second-token  \n"), 0600))
	token, err = creds.GetToken()
	Ok(t, err)
	Equals(t, "second-token", token)
}

func TestPATCredentials_GetToken_FileTakesPrecedence(t *testing.T) {
	tmpDir := t.TempDir()
	tokenFile := filepath.Join(tmpDir, "token")
	Ok(t, os.WriteFile(tokenFile, []byte("file-token"), 0600))

	creds := &azuredevops.PATCredentials{User: "atlantis", Token: "static-token", TokenFile: tokenFile}
	token, err := creds.GetToken()
	Ok(t, err)
	Equals(t, "file-token", token)
}

func TestPATCredentials_GetToken_MissingFile(t *testing.T) {
	creds := &azuredevops.PATCredentials{User: "atlantis", TokenFile: "/does/not/exist"}
	_, err := creds.GetToken()
	Assert(t, err != nil, "expected an error when the token file is missing")
}

func TestPATCredentials_GetToken_EmptyWhenUnset(t *testing.T) {
	creds := &azuredevops.PATCredentials{User: "atlantis"}
	token, err := creds.GetToken()
	Ok(t, err)
	Equals(t, "", token)
}

func TestPATCredentials_GetToken_WhitespaceOnlyFile(t *testing.T) {
	tmpDir := t.TempDir()
	tokenFile := filepath.Join(tmpDir, "token")
	Ok(t, os.WriteFile(tokenFile, []byte("   \n\t"), 0600))

	creds := &azuredevops.PATCredentials{User: "atlantis", TokenFile: tokenFile}
	token, err := creds.GetToken()
	Ok(t, err)
	Equals(t, "", token)
}

// TestPATCredentials_GetToken_KubernetesSecretRotation reproduces how the
// kubelet (and the External Secrets Operator on top of it) exposes and rotates
// a mounted secret: the user-facing file is a symlink to ..data/<key>, ..data
// is itself a symlink to a timestamped directory, and a rotation atomically
// swaps the ..data symlink. Because the token is re-read on every call and
// os.ReadFile follows symlinks, the rotated value is picked up with no restart.
func TestPATCredentials_GetToken_KubernetesSecretRotation(t *testing.T) {
	mount := t.TempDir()

	writeVersion := func(versionDir, content string) {
		dir := filepath.Join(mount, versionDir)
		Ok(t, os.Mkdir(dir, 0755))
		Ok(t, os.WriteFile(filepath.Join(dir, "token"), []byte(content), 0600))
	}

	// Atomically repoint <mount>/..data at the given version directory, exactly
	// as the kubelet does (write a temp symlink, then rename it over ..data).
	swapData := func(versionDir string) {
		tmp := filepath.Join(mount, "..data_tmp")
		_ = os.Remove(tmp)
		Ok(t, os.Symlink(versionDir, tmp))
		Ok(t, os.Rename(tmp, filepath.Join(mount, "..data")))
	}

	writeVersion("..data_v1", "k8s-token-1")
	swapData("..data_v1")
	// The user-facing path Atlantis is configured with is a symlink chain:
	// <mount>/token -> ..data/token -> ..data_v1/token
	Ok(t, os.Symlink(filepath.Join("..data", "token"), filepath.Join(mount, "token")))

	creds := &azuredevops.PATCredentials{User: "atlantis", TokenFile: filepath.Join(mount, "token")}

	token, err := creds.GetToken()
	Ok(t, err)
	Equals(t, "k8s-token-1", token)

	// The secret is rotated: a new version directory is written and the ..data
	// symlink is atomically swapped to point at it.
	writeVersion("..data_v2", "k8s-token-2")
	swapData("..data_v2")

	token, err = creds.GetToken()
	Ok(t, err)
	Equals(t, "k8s-token-2", token)
}
