// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package azuredevops

import (
	"fmt"
	"os"
	"strings"
)

// Credentials handles supplying the token used to authenticate with Azure
// DevOps. Implementations may return a static token or re-read it from a file
// on every call so that an externally-rotated token is picked up without
// restarting Atlantis.
type Credentials interface {
	// GetToken returns the token used to authenticate API and git requests.
	GetToken() (string, error)
	// GetUser returns the username associated with the token.
	GetUser() (string, error)
}

// PATCredentials implements Credentials for the personal access token flow. When
// TokenFile is set the token is re-read from disk on every call, allowing an
// external process to rotate the token without an Atlantis restart.
type PATCredentials struct {
	User      string
	Token     string
	TokenFile string
}

// GetUser returns the username for these credentials.
func (c *PATCredentials) GetUser() (string, error) {
	return c.User, nil
}

// GetToken returns the token, reading it from TokenFile when configured.
func (c *PATCredentials) GetToken() (string, error) {
	if c.TokenFile != "" {
		content, err := os.ReadFile(c.TokenFile)
		if err != nil {
			return "", fmt.Errorf("reading azure devops token file %q: %w", c.TokenFile, err)
		}
		return strings.TrimSpace(string(content)), nil
	}
	return c.Token, nil
}
