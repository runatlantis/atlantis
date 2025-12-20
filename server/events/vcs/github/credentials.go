// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v71/github"
)

//go:generate pegomock generate --package mocks -o mocks/mock_credentials.go Credentials

// GithubCredentials handles creating http.Clients that authenticate.
type Credentials interface {
	Client() (*http.Client, error)
	GetToken() (string, error)
	GetUser() (string, error)
}

// GithubAnonymousCredentials expose no credentials.
type AnonymousCredentials struct{}

// Client returns a client with no credentials.
func (c *AnonymousCredentials) Client() (*http.Client, error) {
	tr := http.DefaultTransport
	return &http.Client{Transport: tr}, nil
}

// GetUser returns the username for these credentials.
func (c *AnonymousCredentials) GetUser() (string, error) {
	return "anonymous", nil
}

// GetToken returns an empty token.
func (c *AnonymousCredentials) GetToken() (string, error) {
	return "", nil
}

// GithubUserCredentials implements GithubCredentials for the personal auth token flow.
type UserCredentials struct {
	User      string
	Token     string
	TokenFile string
}

type UserTransport struct {
	Credentials *UserCredentials
	Transport   *github.BasicAuthTransport
}

func (t *UserTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// update token
	token, err := t.Credentials.GetToken()
	if err != nil {
		return nil, err
	}
	t.Transport.Password = token

	// defer to the underlying transport
	return t.Transport.RoundTrip(req)
}

// Client returns a client for basic auth user credentials.
func (c *UserCredentials) Client() (*http.Client, error) {
	password, err := c.GetToken()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &UserTransport{
			Credentials: c,
			Transport: &github.BasicAuthTransport{
				Username: strings.TrimSpace(c.User),
				Password: strings.TrimSpace(password),
			},
		},
	}
	return client, nil
}

// GetUser returns the username for these credentials.
func (c *UserCredentials) GetUser() (string, error) {
	return c.User, nil
}

// GetToken returns the user token.
func (c *UserCredentials) GetToken() (string, error) {
	if c.TokenFile != "" {
		content, err := os.ReadFile(c.TokenFile)
		if err != nil {
			return "", fmt.Errorf("failed reading github token file: %w", err)
		}

		return string(content), nil
	}

	return c.Token, nil
}

// GithubAppCredentials implements GithubCredentials for github app installation token flow.
type AppCredentials struct {
	AppID          int64
	Key            []byte
	Hostname       string
	apiURL         *url.URL
	InstallationID int64
	tr             *ghinstallation.Transport
	AppSlug        string
}

// Client returns a github app installation client.
func (c *AppCredentials) Client() (*http.Client, error) {
	itr, err := c.transport()
	if err != nil {
		return nil, err
	}
	return &http.Client{Transport: itr}, nil
}

// GetUser returns the username for these credentials.
func (c *AppCredentials) GetUser() (string, error) {
	// Keeping backwards compatibility since this flag is optional
	if c.AppSlug == "" {
		return "", nil
	}
	client, err := c.Client()

	if err != nil {
		return "", fmt.Errorf("initializing client: %w", err)
	}

	ghClient := github.NewClient(client)
	ghClient.BaseURL = c.getAPIURL()
	ctx := context.Background()

	app, _, err := ghClient.Apps.Get(ctx, c.AppSlug)

	if err != nil {
		return "", fmt.Errorf("getting app details: %w", err)
	}
	// Currently there is no way to get the bot's login info, so this is a
	// hack until Github exposes that.
	return fmt.Sprintf("%s[bot]", app.GetSlug()), nil
}

// GetToken returns a fresh installation token.
func (c *AppCredentials) GetToken() (string, error) {
	tr, err := c.transport()
	if err != nil {
		return "", fmt.Errorf("transport failed: %w", err)
	}

	return tr.Token(context.Background())
}

func (c *AppCredentials) getInstallationID() (int64, error) {
	if c.InstallationID != 0 {
		return c.InstallationID, nil
	}

	tr := http.DefaultTransport
	// A non-installation transport
	t, err := ghinstallation.NewAppsTransport(tr, c.AppID, c.Key)
	if err != nil {
		return 0, err
	}
	t.BaseURL = c.getAPIURL().String()

	// Query github with the app's JWT
	client := github.NewClient(&http.Client{Transport: t})
	client.BaseURL = c.getAPIURL()
	ctx := context.Background()

	installations, _, err := client.Apps.ListInstallations(ctx, nil)
	if err != nil {
		return 0, err
	}

	if len(installations) != 1 {
		return 0, fmt.Errorf("wrong number of installations, expected 1, found %d", len(installations))
	}

	c.InstallationID = installations[0].GetID()
	return c.InstallationID, nil
}

func (c *AppCredentials) transport() (*ghinstallation.Transport, error) {
	if c.tr != nil {
		return c.tr, nil
	}

	installationID, err := c.getInstallationID()
	if err != nil {
		return nil, err
	}

	tr := http.DefaultTransport
	itr, err := ghinstallation.New(tr, c.AppID, installationID, c.Key)
	if err == nil {
		apiURL := c.getAPIURL()
		itr.BaseURL = strings.TrimSuffix(apiURL.String(), "/")
		c.tr = itr
	}
	return itr, err
}

func (c *AppCredentials) getAPIURL() *url.URL {
	if c.apiURL != nil {
		return c.apiURL
	}

	c.apiURL = resolveGithubAPIURL(c.Hostname)
	return c.apiURL
}

func resolveGithubAPIURL(hostname string) *url.URL {
	// If we're using github.com then we don't need to do any additional configuration
	// for the client. It we're using Github Enterprise, then we need to manually
	// set the base url for the API.
	baseURL := &url.URL{
		Scheme: "https",
		Host:   "api.github.com",
		Path:   "/",
	}

	if hostname != "github.com" {
		baseURL.Host = hostname
		baseURL.Path = "/api/v3/"
	}

	return baseURL
}
