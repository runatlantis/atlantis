package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
)

// GithubAppController handles the creation and setup of a new GitHub app
type GithubAppController struct {
	AtlantisURL         *url.URL
	Logger              *logging.SimpleLogger
	GithubSetupComplete bool
	GithubHostname      string
	GithubOrg           string
}

type githubWebhook struct {
	URL    string `json:"url"`
	Active bool   `json:"active"`
}

// githubAppRequest contains the query parameters for
// https://developer.github.com/apps/building-github-apps/creating-github-apps-from-a-manifest
type githubAppRequest struct {
	Description string            `json:"description"`
	Events      []string          `json:"default_events"`
	Name        string            `json:"name"`
	Permissions map[string]string `json:"default_permissions"`
	Public      bool              `json:"public"`
	RedirectURL string            `json:"redirect_url"`
	URL         string            `json:"url"`
	Webhook     *githubWebhook    `json:"hook_attributes"`
}

// ExchangeCode handles the user coming back from creating their app
// A code query parameter is exchanged for this app's ID, key, and webhook_secret
// Implements https://developer.github.com/apps/building-github-apps/creating-github-apps-from-a-manifest/#implementing-the-github-app-manifest-flow
func (g *GithubAppController) ExchangeCode(w http.ResponseWriter, r *http.Request) {

	if g.GithubSetupComplete {
		g.respond(w, logging.Error, http.StatusBadRequest, "Atlantis already has GitHub credentials")
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		g.respond(w, logging.Debug, http.StatusOK, "Ignoring callback, missing code query parameter")
	}

	g.Logger.Debug("Exchanging GitHub app code for app credentials")
	creds := &vcs.GithubAnonymousCredentials{}
	client, err := vcs.NewGithubClient(g.GithubHostname, creds, g.Logger)
	if err != nil {
		g.respond(w, logging.Error, http.StatusInternalServerError, "Failed to exchange code for github app: %s", err)
		return
	}

	app, err := client.ExchangeCode(code)
	if err != nil {
		g.respond(w, logging.Error, http.StatusInternalServerError, "Failed to exchange code for github app: %s", err)
		return
	}

	g.Logger.Debug("Found credentials for GitHub app %q with id %d", app.Name, app.ID)

	err = githubAppSetupTemplate.Execute(w, GithubSetupData{
		Target:        "",
		Manifest:      "",
		ID:            app.ID,
		Key:           app.Key,
		WebhookSecret: app.WebhookSecret,
		URL:           app.URL,
	})
	if err != nil {
		g.Logger.Err(err.Error())
	}
}

// New redirects the user to create a new GitHub app
func (g *GithubAppController) New(w http.ResponseWriter, r *http.Request) {

	if g.GithubSetupComplete {
		g.respond(w, logging.Error, http.StatusBadRequest, "Atlantis already has GitHub credentials")
		return
	}

	manifest := &githubAppRequest{
		Name:        fmt.Sprintf("Atlantis for %s", g.AtlantisURL.Hostname()),
		Description: fmt.Sprintf("Terraform Pull Request Automation at %s", g.AtlantisURL),
		URL:         g.AtlantisURL.String(),
		RedirectURL: fmt.Sprintf("%s/github-app/exchange-code", g.AtlantisURL),
		Public:      false,
		Webhook: &githubWebhook{
			Active: true,
			URL:    fmt.Sprintf("%s/events", g.AtlantisURL),
		},
		Events: []string{
			"check_run",
			"create",
			"delete",
			"issue_comment",
			"issues",
			"pull_request_review_comment",
			"pull_request_review",
			"pull_request",
			"push",
		},
		Permissions: map[string]string{
			"checks":           "write",
			"contents":         "write",
			"issues":           "write",
			"pull_requests":    "write",
			"repository_hooks": "write",
			"statuses":         "write",
		},
	}

	url := &url.URL{
		Scheme: "https",
		Host:   g.GithubHostname,
		Path:   "/settings/apps/new",
	}

	// https://developer.github.com/apps/building-github-apps/creating-github-apps-using-url-parameters/#about-github-app-url-parameters
	if g.GithubOrg != "" {
		url.Path = fmt.Sprintf("organizations/%s%s", g.GithubOrg, url.Path)
	}

	jsonManifest, err := json.MarshalIndent(manifest, "", " ")
	if err != nil {
		g.respond(w, logging.Error, http.StatusBadRequest, "Failed to serialize manifest: %s", err)
		return
	}

	err = githubAppSetupTemplate.Execute(w, GithubSetupData{
		Target:   url.String(),
		Manifest: string(jsonManifest),
	})
	if err != nil {
		g.Logger.Err(err.Error())
	}
}

func (g *GithubAppController) respond(w http.ResponseWriter, lvl logging.LogLevel, code int, format string, args ...interface{}) {
	response := fmt.Sprintf(format, args...)
	g.Logger.Log(lvl, response)
	w.WriteHeader(code)
	fmt.Fprintln(w, response)
}
