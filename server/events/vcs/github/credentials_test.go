// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package github_test

import (
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"

	gogithub "github.com/google/go-github/v88/github"
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

func TestUserCredentials_TokenFileRotationTrimsWhitespace(t *testing.T) {
	tokenFile := filepath.Join(t.TempDir(), "token")
	Ok(t, os.WriteFile(tokenFile, []byte("initial-token\n"), 0o600))

	credentials := &github.UserCredentials{
		User:      " user \n",
		TokenFile: tokenFile,
	}
	client, err := credentials.Client()
	Ok(t, err)

	transport, ok := client.Transport.(*github.UserTransport)
	if !ok {
		t.Fatalf("expected *github.UserTransport, got %T", client.Transport)
	}

	var gotUser string
	var gotToken string
	transport.Transport.Transport = userTransportRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		gotUser, gotToken, _ = req.BasicAuth()
		return userTransportResponse(req), nil
	})

	Ok(t, os.WriteFile(tokenFile, []byte(" rotated-token \n"), 0o600))
	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/user", nil)
	Ok(t, err)
	resp, err := client.Do(req)
	Ok(t, err)
	Ok(t, resp.Body.Close())

	Equals(t, "user", gotUser)
	Equals(t, "rotated-token", gotToken)
}

func TestUserTransport_RoundTripConcurrentDoesNotMutateBaseTransport(t *testing.T) {
	const requests = 32

	observedTokens := make(chan string, requests)
	release := make(chan struct{})
	baseTransport := &gogithub.BasicAuthTransport{
		Username: "user",
		Password: "initial-token",
		Transport: userTransportRoundTripFunc(func(req *http.Request) (*http.Response, error) {
			_, token, _ := req.BasicAuth()
			observedTokens <- token
			<-release
			return userTransportResponse(req), nil
		}),
	}
	transport := &github.UserTransport{
		Credentials: &github.UserCredentials{Token: "rotated-token"},
		Transport:   baseTransport,
	}

	start := make(chan struct{})
	errs := make(chan error, requests)
	var wg sync.WaitGroup
	for i := 0; i < requests; i++ {
		req, err := http.NewRequest(http.MethodGet, "https://api.github.com/user", nil)
		Ok(t, err)
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			resp, err := transport.RoundTrip(req)
			if err == nil {
				err = resp.Body.Close()
			}
			errs <- err
		}()
	}

	close(start)
	for i := 0; i < requests; i++ {
		Equals(t, "rotated-token", <-observedTokens)
	}
	close(release)
	wg.Wait()
	for i := 0; i < requests; i++ {
		Ok(t, <-errs)
	}

	Equals(t, "initial-token", baseTransport.Password)
}

type userTransportRoundTripFunc func(*http.Request) (*http.Response, error)

func (f userTransportRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func userTransportResponse(req *http.Request) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       http.NoBody,
		Request:    req,
	}
}
