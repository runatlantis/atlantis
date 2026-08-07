// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package azuredevops_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/events/vcs/azuredevops"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func Test_azuredevopsTokenRotator_GenerateJob(t *testing.T) {
	tests := []struct {
		name             string
		credentials      azuredevops.Credentials
		credsFileWritten bool
		wantErr          bool
	}{
		{
			name:             "should write .git-credentials file on start",
			credentials:      &azuredevops.PATCredentials{User: "atlantis", Token: "some-token"},
			credsFileWritten: true,
			wantErr:          false,
		},
		{
			name:             "should return an error if the token file is missing",
			credentials:      &azuredevops.PATCredentials{User: "atlantis", TokenFile: "/does/not/exist"},
			credsFileWritten: false,
			wantErr:          true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			t.Setenv("HOME", tmpDir)
			r := azuredevops.NewTokenRotator(logging.NewNoopLogger(t), tt.credentials, "dev.azure.com", "atlantis", tmpDir)
			got, err := r.GenerateJob()
			if (err != nil) != tt.wantErr {
				t.Errorf("azuredevopsTokenRotator.GenerateJob() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.credsFileWritten {
				credsFileContent := fmt.Sprintf(`https://atlantis:some-token@%s`, "dev.azure.com")
				actContents, err := os.ReadFile(filepath.Join(tmpDir, ".git-credentials"))
				Ok(t, err)
				Equals(t, credsFileContent, string(actContents))
				Equals(t, 30*time.Second, got.Period)
			}
		})
	}
}

func Test_azuredevopsTokenRotator_RefreshesFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	tokenFile := filepath.Join(tmpDir, "token")
	Ok(t, os.WriteFile(tokenFile, []byte("first-token"), 0600))

	creds := &azuredevops.PATCredentials{User: "atlantis", TokenFile: tokenFile}
	r := azuredevops.NewTokenRotator(logging.NewNoopLogger(t), creds, "dev.azure.com", "atlantis", tmpDir)

	_, err := r.GenerateJob()
	Ok(t, err)
	actContents, err := os.ReadFile(filepath.Join(tmpDir, ".git-credentials"))
	Ok(t, err)
	Equals(t, "https://atlantis:first-token@dev.azure.com", string(actContents))

	// Simulate an external rotation of the token file and run the rotator
	// again. The credentials line is replaced in place rather than appended.
	Ok(t, os.WriteFile(tokenFile, []byte("second-token"), 0600))
	r.Run()
	actContents, err = os.ReadFile(filepath.Join(tmpDir, ".git-credentials"))
	Ok(t, err)
	Equals(t, "https://atlantis:second-token@dev.azure.com", string(actContents))
}

// Test_azuredevopsTokenRotator_PreservesCredsOnFailure ensures that a transient
// failure to read the rotated token (e.g. the file is briefly unavailable)
// leaves the previously-written ~/.git-credentials intact instead of clobbering
// it, so in-flight git operations keep working.
func Test_azuredevopsTokenRotator_PreservesCredsOnFailure(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	tokenFile := filepath.Join(tmpDir, "token")
	Ok(t, os.WriteFile(tokenFile, []byte("good-token"), 0600))

	creds := &azuredevops.PATCredentials{User: "atlantis", TokenFile: tokenFile}
	r := azuredevops.NewTokenRotator(logging.NewNoopLogger(t), creds, "dev.azure.com", "atlantis", tmpDir)

	_, err := r.GenerateJob()
	Ok(t, err)
	credsPath := filepath.Join(tmpDir, ".git-credentials")
	actContents, err := os.ReadFile(credsPath)
	Ok(t, err)
	Equals(t, "https://atlantis:good-token@dev.azure.com", string(actContents))

	// The token file disappears; the next rotation must not overwrite the
	// last-good credentials. Run() swallows and logs the error.
	Ok(t, os.Remove(tokenFile))
	r.Run()
	actContents, err = os.ReadFile(credsPath)
	Ok(t, err)
	Equals(t, "https://atlantis:good-token@dev.azure.com", string(actContents))
}

// Test_azuredevopsTokenRotator_PreservesCredsOnEmptyToken ensures that a
// transiently empty token (e.g. the backing secret was briefly empty mid-sync
// or accidentally wiped) does not clobber the last-good ~/.git-credentials with
// an empty credential, which would break git operations.
func Test_azuredevopsTokenRotator_PreservesCredsOnEmptyToken(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	tokenFile := filepath.Join(tmpDir, "token")
	Ok(t, os.WriteFile(tokenFile, []byte("good-token"), 0600))

	creds := &azuredevops.PATCredentials{User: "atlantis", TokenFile: tokenFile}
	r := azuredevops.NewTokenRotator(logging.NewNoopLogger(t), creds, "dev.azure.com", "atlantis", tmpDir)

	_, err := r.GenerateJob()
	Ok(t, err)
	credsPath := filepath.Join(tmpDir, ".git-credentials")
	actContents, err := os.ReadFile(credsPath)
	Ok(t, err)
	Equals(t, "https://atlantis:good-token@dev.azure.com", string(actContents))

	// The secret is wiped to empty (whitespace-only here). The rotation must be
	// skipped without error and the last-good credential left untouched.
	Ok(t, os.WriteFile(tokenFile, []byte("  \n"), 0600))
	r.Run()
	actContents, err = os.ReadFile(credsPath)
	Ok(t, err)
	Equals(t, "https://atlantis:good-token@dev.azure.com", string(actContents))
}
