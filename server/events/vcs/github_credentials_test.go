package vcs_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/vcs/testdata"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestGithubClient_GetUser_AppSlug(t *testing.T) {
	defer disableSSLVerification()()
	testServer, err := testdata.GithubAppTestServer(t)
	Ok(t, err)

	anonCreds := &vcs.GithubAnonymousCredentials{}
	anonClient, err := vcs.NewGithubClient(testServer, anonCreds, vcs.GithubConfig{}, logging.NewNoopLogger(t))
	Ok(t, err)
	tempSecrets, err := anonClient.ExchangeCode("good-code")
	Ok(t, err)

	appCreds := &vcs.GithubAppCredentials{
		AppID:    tempSecrets.ID,
		Key:      []byte(testdata.GithubPrivateKey),
		Hostname: testServer,
		AppSlug:  "some-app",
	}

	user, err := appCreds.GetUser()
	Ok(t, err)

	Assert(t, user == "octoapp[bot]", "user should not be empty")
}

func TestGithubClient_AppAuthentication(t *testing.T) {
	defer disableSSLVerification()()
	testServer, err := testdata.GithubAppTestServer(t)
	Ok(t, err)

	anonCreds := &vcs.GithubAnonymousCredentials{}
	anonClient, err := vcs.NewGithubClient(testServer, anonCreds, vcs.GithubConfig{}, logging.NewNoopLogger(t))
	Ok(t, err)
	tempSecrets, err := anonClient.ExchangeCode("good-code")
	Ok(t, err)

	appCreds := &vcs.GithubAppCredentials{
		AppID:    tempSecrets.ID,
		Key:      []byte(testdata.GithubPrivateKey),
		Hostname: testServer,
	}
	_, err = vcs.NewGithubClient(testServer, appCreds, vcs.GithubConfig{}, logging.NewNoopLogger(t))
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
