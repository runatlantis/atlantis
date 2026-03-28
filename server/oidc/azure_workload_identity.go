// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/runatlantis/atlantis/server/logging"
)

const (
	// azureFederatedTokenFileEnv is the environment variable set by the AKS
	// workload identity mutation webhook pointing to the projected service
	// account token file.
	azureFederatedTokenFileEnv = "AZURE_FEDERATED_TOKEN_FILE"

	// tokenCacheTTL controls how long a federated token is cached before
	// re-reading from disk. Kubernetes rotates projected tokens at 80% of
	// their TTL (minimum 1 hour), so 10 minutes gives comfortable margin.
	tokenCacheTTL = 10 * time.Minute
)

// AzureWorkloadIdentity reads the projected service account token from disk
// and caches it for use as a client_assertion during OIDC token exchange.
type AzureWorkloadIdentity struct {
	mu              sync.RWMutex
	cachedToken     string
	cacheExpiration time.Time
	tokenFile       string
	logger          logging.SimpleLogging
}

// NewAzureWorkloadIdentity creates a new AzureWorkloadIdentity. It resolves
// the token file path from AZURE_FEDERATED_TOKEN_FILE at construction time
// and returns an error if the environment variable is not set.
func NewAzureWorkloadIdentity(logger logging.SimpleLogging) (*AzureWorkloadIdentity, error) {
	file, ok := os.LookupEnv(azureFederatedTokenFileEnv)
	if !ok || file == "" {
		return nil, fmt.Errorf("AZURE_FEDERATED_TOKEN_FILE environment variable is not set; " +
			"ensure the pod has the azure.workload.identity/use=true label and the " +
			"service account is annotated with azure.workload.identity/client-id")
	}
	return &AzureWorkloadIdentity{
		tokenFile: file,
		logger:    logger,
	}, nil
}

// GetFederatedToken returns the federated service account token for use as a
// client_assertion in OAuth2 token exchange. It reads from the file resolved
// at construction time and caches the result with double-check locking for
// thread safety.
func (a *AzureWorkloadIdentity) GetFederatedToken() (string, error) {
	// Fast path: return cached token if still valid.
	a.mu.RLock()
	if time.Now().Before(a.cacheExpiration) {
		token := a.cachedToken
		a.mu.RUnlock()
		return token, nil
	}
	a.mu.RUnlock()

	// Slow path: acquire write lock and re-check (double-check locking).
	a.mu.Lock()
	defer a.mu.Unlock()

	if time.Now().Before(a.cacheExpiration) {
		return a.cachedToken, nil
	}

	content, err := os.ReadFile(a.tokenFile)
	if err != nil {
		return "", fmt.Errorf("reading federated token from %s: %w", a.tokenFile, err)
	}

	a.cachedToken = string(content)
	a.cacheExpiration = time.Now().Add(tokenCacheTTL)
	a.logger.Debug("refreshed azure federated token from %q", a.tokenFile)

	return a.cachedToken, nil
}
