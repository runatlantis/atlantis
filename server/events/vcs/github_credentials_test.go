package vcs_test

import (
	"fmt"
	"testing"

	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/vcs/fixtures"
	. "github.com/runatlantis/atlantis/testing"
)

func TestGithubClient_AppAuthentication(t *testing.T) {
	defer disableSSLVerification()()
	testServer, err := fixtures.GithubAppTestServer(t)
	Ok(t, err)

	anonCreds := &vcs.GithubAnonymousCredentials{}
	anonClient, err := vcs.NewGithubClient(testServer, anonCreds, nil)
	Ok(t, err)
	tempSecrets, err := anonClient.ExchangeCode("good-code")
	Ok(t, err)

	tmpDir, cleanup := DirStructure(t, map[string]interface{}{
		"key.pem": tempSecrets.Key,
	})
	defer cleanup()
	keyPath := fmt.Sprintf("%v/key.pem", tmpDir)

	appCreds := &vcs.GithubAppCredentials{
		AppID:    tempSecrets.ID,
		KeyPath:  keyPath,
		Hostname: testServer,
	}
	_, err = vcs.NewGithubClient(testServer, appCreds, nil)
	Ok(t, err)

	token, err := appCreds.GetToken()
	Ok(t, err)

	newToken, err := appCreds.GetToken()
	Ok(t, err)

	if token != newToken {
		t.Errorf("app token was not cached: %q != %q", token, newToken)
	}
}
