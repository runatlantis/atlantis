package vcs

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v28/github"
)

func githubAPIURL(hostname string) *url.URL {
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

// GithubCredentials handles creating http.Clients that authenticate
type GithubCredentials interface {
	Client(hostname string) (*http.Client, error)
}

// GithubUserCredentials implements GithubCredentials for the personal auth token flow
type GithubUserCredentials struct {
	User  string
	Token string
}

// Client returns a client for basic auth user credentials
func (c *GithubUserCredentials) Client(hostname string) (*http.Client, error) {
	tr := &github.BasicAuthTransport{
		Username: strings.TrimSpace(c.User),
		Password: strings.TrimSpace(c.Token),
	}
	return tr.Client(), nil
}

// GithubAppCredentials implements GithubCredentials for github app installation token flow
type GithubAppCredentials struct {
	AppID   int64
	KeyPath string
}

func (c *GithubAppCredentials) getInstallationID(apiURL string) (int64, error) {
	tr := http.DefaultTransport
	// A non-installation transport
	t, err := ghinstallation.NewAppsTransportKeyFromFile(tr, c.AppID, c.KeyPath)
	if err != nil {
		return 0, err
	}

	// Query github with the app's JWT
	client := github.NewClient(&http.Client{Transport: t})
	ctx := context.Background()
	app := &struct {
		ID int64 `json:"id"`
	}{}

	url := fmt.Sprintf("%sapp", apiURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	_, err = client.Do(ctx, req, app)
	if err != nil {
		return 0, err
	}

	return app.ID, nil
}

// Client returns a github app installation client
func (c *GithubAppCredentials) Client(hostname string) (*http.Client, error) {
	installationID, err := c.getInstallationID(githubAPIURL(hostname).String())
	if err != nil {
		return nil, err
	}

	tr := http.DefaultTransport
	itr, err := ghinstallation.NewKeyFromFile(tr, c.AppID, installationID, c.KeyPath)
	if err != nil {
		return nil, err
	}

	return &http.Client{Transport: itr}, nil
}
