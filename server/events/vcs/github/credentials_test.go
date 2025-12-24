// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package github_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events/vcs/github"
	"github.com/runatlantis/atlantis/server/events/vcs/github/testdata"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestClient_GetUser_AppSlug(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	defer disableSSLVerification()()
	testServer, err := testdata.GithubAppTestServer(t)
	Ok(t, err)

	anonCreds := &github.AnonymousCredentials{}
	anonClient, err := github.New(testServer, anonCreds, github.Config{}, 0, logging.NewNoopLogger(t))
	Ok(t, err)
	tempSecrets, err := anonClient.ExchangeCode(logger, "good-code")
	Ok(t, err)

	appCreds := &github.AppCredentials{
		AppID:    tempSecrets.ID,
		Key:      []byte(testdata.PrivateKey),
		Hostname: testServer,
		AppSlug:  "some-app",
	}

	user, err := appCreds.GetUser()
	Ok(t, err)

	Assert(t, user == "octoapp[bot]", "user should not be empty")
}

func TestClient_AppAuthentication(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	defer disableSSLVerification()()
	testServer, err := testdata.GithubAppTestServer(t)
	Ok(t, err)

	anonCreds := &github.AnonymousCredentials{}
	anonClient, err := github.New(testServer, anonCreds, github.Config{}, 0, logging.NewNoopLogger(t))
	Ok(t, err)
	tempSecrets, err := anonClient.ExchangeCode(logger, "good-code")
	Ok(t, err)

	appCreds := &github.AppCredentials{
		AppID:    tempSecrets.ID,
		Key:      []byte(testdata.PrivateKey),
		Hostname: testServer,
	}
	_, err = github.New(testServer, appCreds, github.Config{}, 0, logging.NewNoopLogger(t))
	Ok(t, err)

	token, err := appCreds.GetToken()
	Ok(t, err)

	newToken, err := appCreds.GetToken()
	Ok(t, err)

	user, err := appCreds.GetUser()
	Ok(t, err)

	Assert(t, user == "", "user should be empty")

	if token != newToken {
		t.Errorf("app token was not cached: %q != %q", token, newToken)
	}
}

func TestClient_MultipleAppAuthentication(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	defer disableSSLVerification()()
	testServer, err := testdata.GithubMultipleAppTestServer(t)
	Ok(t, err)

	anonCreds := &github.AnonymousCredentials{}
	anonClient, err := github.New(testServer, anonCreds, github.Config{}, 0, logging.NewNoopLogger(t))
	Ok(t, err)
	tempSecrets, err := anonClient.ExchangeCode(logger, "good-code")
	Ok(t, err)

	appCreds := &github.AppCredentials{
		AppID:          tempSecrets.ID,
		InstallationID: 1,
		Key:            []byte(testdata.PrivateKey),
		Hostname:       testServer,
	}
	_, err = github.New(testServer, appCreds, github.Config{}, 0, logging.NewNoopLogger(t))
	Ok(t, err)

	token, err := appCreds.GetToken()
	Ok(t, err)

	newToken, err := appCreds.GetToken()
	Ok(t, err)

	user, err := appCreds.GetUser()
	Ok(t, err)

	Assert(t, user == "", "user should be empty")

	if token != newToken {
		t.Errorf("app token was not cached: %q != %q", token, newToken)
	}
}
