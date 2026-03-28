// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/runatlantis/atlantis/server/oidc"
	. "github.com/runatlantis/atlantis/testing"
)

func setupAzureIdentity(t *testing.T, tokenContent string) {
	t.Helper()
	tempDir := t.TempDir()
	tokenFilePath := filepath.Join(tempDir, "token.txt")
	Ok(t, os.WriteFile(tokenFilePath, []byte(tokenContent), 0600))
	t.Setenv("AZURE_FEDERATED_TOKEN_FILE", tokenFilePath)
}

func TestAzureWorkloadIdentity_GetFederatedToken(t *testing.T) {
	logger := newTestLogger(t)
	setupAzureIdentity(t, "serviceAccountToken")

	awi, err := oidc.NewAzureWorkloadIdentity(logger)
	Ok(t, err)

	token, err := awi.GetFederatedToken()
	Ok(t, err)
	Equals(t, "serviceAccountToken", token)
}

func TestAzureWorkloadIdentity_CachesToken(t *testing.T) {
	logger := newTestLogger(t)
	setupAzureIdentity(t, "initial-token")

	awi, err := oidc.NewAzureWorkloadIdentity(logger)
	Ok(t, err)

	token1, err := awi.GetFederatedToken()
	Ok(t, err)

	// Overwrite the file with different content.
	file := os.Getenv("AZURE_FEDERATED_TOKEN_FILE")
	Ok(t, os.WriteFile(file, []byte("updated-token"), 0600))

	// Second call should return cached token (not the new file content).
	token2, err := awi.GetFederatedToken()
	Ok(t, err)
	Equals(t, token1, token2)
}

func TestAzureWorkloadIdentity_MissingEnvVar(t *testing.T) {
	logger := newTestLogger(t)
	t.Setenv("AZURE_FEDERATED_TOKEN_FILE", "")

	_, err := oidc.NewAzureWorkloadIdentity(logger)
	Assert(t, err != nil, "expected error when AZURE_FEDERATED_TOKEN_FILE is empty")
}

func TestAzureWorkloadIdentity_NonExistentFile(t *testing.T) {
	logger := newTestLogger(t)
	t.Setenv("AZURE_FEDERATED_TOKEN_FILE", "/nonexistent/path/token.txt")

	awi, err := oidc.NewAzureWorkloadIdentity(logger)
	Ok(t, err)

	_, err = awi.GetFederatedToken()
	Assert(t, err != nil, "expected error when token file does not exist")
}
