// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package azuredevops_test

import (
	"errors"
	"testing"

	azuredevopsclient "github.com/runatlantis/atlantis/server/events/vcs/azuredevops"
	. "github.com/runatlantis/atlantis/testing"
)

// fakeCredentials lets us exercise constructor error paths that PATCredentials
// itself never triggers.
type fakeCredentials struct {
	user     string
	userErr  error
	token    string
	tokenErr error
}

func (c fakeCredentials) GetUser() (string, error)  { return c.user, c.userErr }
func (c fakeCredentials) GetToken() (string, error) { return c.token, c.tokenErr }

func TestNew_SetsUserName(t *testing.T) {
	client, err := azuredevopsclient.New("dev.azure.com", "some-user", "some-token")
	Ok(t, err)
	Equals(t, "some-user", client.UserName)
}

func TestNewClientForCredentials_DefaultHostname(t *testing.T) {
	client, err := azuredevopsclient.NewClientForCredentials(
		"dev.azure.com",
		&azuredevopsclient.PATCredentials{User: "user", Token: "token"},
	)
	Ok(t, err)
	Equals(t, "user", client.UserName)
	Equals(t, "https://dev.azure.com/", client.Client.BaseURL.String())
}

func TestNewClientForCredentials_CustomHostname(t *testing.T) {
	client, err := azuredevopsclient.NewClientForCredentials(
		"myorg.visualstudio.com",
		&azuredevopsclient.PATCredentials{User: "user", Token: "token"},
	)
	Ok(t, err)
	Equals(t, "https://myorg.visualstudio.com/", client.Client.BaseURL.String())
}

func TestNewClientForCredentials_PropagatesGetUserError(t *testing.T) {
	_, err := azuredevopsclient.NewClientForCredentials(
		"dev.azure.com",
		fakeCredentials{userErr: errors.New("boom")},
	)
	Assert(t, err != nil, "expected an error when GetUser fails")
	ErrContains(t, "getting azure devops user", err)
	ErrContains(t, "boom", err)
}
