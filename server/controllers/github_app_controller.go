// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/runatlantis/atlantis/server/controllers/web_templates"
	"github.com/runatlantis/atlantis/server/events/vcs/github"
	"github.com/runatlantis/atlantis/server/logging"
)

// githubAppStateTTL bounds how long a setup flow stays valid. The admin has
// to click "Setup", create the app on GitHub and be redirected back within
// this window.
const githubAppStateTTL = 10 * time.Minute

// githubAppMaxPendingFlows caps how many setup pages can be waiting for their
// GitHub redirect at once, so repeated visits to /github-app/setup cannot
// grow memory without bound. Further visits are refused rather than evicting
// a flow an admin may be in the middle of.
const githubAppMaxPendingFlows = 16

// GithubAppSetupCookieName is the cookie that binds a pending setup flow to
// the browser that opened /github-app/setup.
const GithubAppSetupCookieName = "atlantis_github_app_setup"

// GithubAppCodeExchanger exchanges a GitHub App manifest code for the new
// app's credentials. It is satisfied by *github.Client and faked in tests.
type GithubAppCodeExchanger interface {
	ExchangeCode(logger logging.SimpleLogging, code string) (*github.GithubAppTemporarySecrets, error)
}

// GithubAppController handles the creation and setup of a new GitHub app.
//
// The manifest flow is a small state machine guarded by mu:
//
//   - idle: no exchange is running. New adds a pending flow keyed by a random
//     state and sets a cookie holding a separate random binding value. Several
//     flows may be pending (up to githubAppMaxPendingFlows), each expiring
//     after githubAppStateTTL.
//   - exchanging: a callback presented an unexpired pending state together
//     with that flow's binding cookie. The flow is claimed atomically, every
//     pending flow is dropped, and the code is exchanged with GitHub outside
//     the lock. While exchanging, New and every other callback get 409.
//   - complete: the exchange succeeded (or the server started with GitHub App
//     credentials). New and every callback get 400 for the rest of the process.
//
// A failed exchange returns to idle with no pending flows, so the admin starts
// again from /github-app/setup and gets a fresh state. A callback whose state
// is missing, unknown or whose cookie does not match gets 400 and leaves the
// pending flows untouched; only an expired flow is dropped.
type GithubAppController struct {
	AtlantisURL         *url.URL              `validate:"required"`
	Logger              logging.SimpleLogging `validate:"required"`
	GithubSetupComplete bool
	GithubHostname      string `validate:"required"`
	GithubOrg           string
	// Exchanger performs the manifest code conversion. When nil an anonymous
	// GitHub client for GithubHostname is created on demand.
	Exchanger GithubAppCodeExchanger

	// mu guards GithubSetupComplete, exchanging and flows once the controller
	// is serving requests.
	mu sync.Mutex
	// exchanging is true while a claimed flow's code is being exchanged.
	exchanging bool
	// flows holds the pending setup flows keyed by their state nonce.
	flows map[string]githubAppFlow
	// now is overridable in tests.
	now func() time.Time
}

// githubAppFlow is one pending setup started by a browser.
type githubAppFlow struct {
	// binding is the value of the cookie set on the browser that started the
	// flow; the callback must present it.
	binding string
	expiry  time.Time
}

// githubAppFlowError is a refused setup step and the HTTP status to answer.
type githubAppFlowError struct {
	status int
	msg    string
}

func (e *githubAppFlowError) Error() string { return e.msg }

var (
	errGithubSetupComplete = &githubAppFlowError{http.StatusBadRequest, "Atlantis already has GitHub credentials"}
	errGithubSetupBusy     = &githubAppFlowError{http.StatusConflict, "a GitHub app setup is already being completed, wait for it to finish"}
	errGithubSetupFull     = &githubAppFlowError{http.StatusTooManyRequests, "too many GitHub app setups in progress, try again later"}
	errGithubStateMissing  = &githubAppFlowError{http.StatusBadRequest, "missing state query parameter"}
	errGithubStateUnknown  = &githubAppFlowError{http.StatusBadRequest, "GitHub app setup state is not valid for this browser, start again from /github-app/setup"}
	errGithubStateExpired  = &githubAppFlowError{http.StatusBadRequest, "GitHub app setup state expired, start again from /github-app/setup"}
)

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
	query := r.URL.Query()
	code := query.Get("code")
	if code == "" {
		g.respond(w, logging.Warn, http.StatusBadRequest, "Ignoring callback, missing code query parameter")
		return
	}

	binding := ""
	if cookie, err := r.Cookie(GithubAppSetupCookieName); err == nil {
		binding = cookie.Value
	}
	if err := g.claimFlow(query.Get("state"), binding); err != nil {
		g.respond(w, logging.Error, err.status, "%s", err)
		return
	}
	// The flow is spent whatever the exchange outcome.
	g.clearBindingCookie(w, r)

	g.Logger.Debug("Exchanging GitHub app code for app credentials")
	app, err := g.exchangeClaimed(code)
	if err != nil {
		g.respond(w, logging.Error, http.StatusInternalServerError, "Failed to exchange code for github app: %s", err)
		return
	}
	g.Logger.Debug("Found credentials for GitHub app %q with id %d", app.Name, app.ID)

	err = web_templates.GithubAppSetupTemplate.Execute(w, web_templates.GithubSetupData{
		Target:          "",
		Manifest:        "",
		ID:              app.ID,
		Key:             app.Key,
		WebhookSecret:   app.WebhookSecret,
		URL:             app.URL,
		CleanedBasePath: g.AtlantisURL.Path,
	})
	if err != nil {
		g.Logger.Err("%s", err.Error())
	}
}

// New redirects the user to create a new GitHub app
func (g *GithubAppController) New(w http.ResponseWriter, r *http.Request) {
	state, binding, flowErr := g.issueFlow()
	if flowErr != nil {
		g.respond(w, logging.Error, flowErr.status, "%s", flowErr)
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
			"administration":   "read",
			"members":          "read",
			"actions":          "read",
		},
	}

	// GitHub echoes the state query parameter back on the redirect to
	// redirect_url, which lets ExchangeCode tie the callback to this page.
	// https://docs.github.com/en/apps/sharing-github-apps/registering-a-github-app-from-a-manifest#implementing-the-github-app-manifest-flow
	target := &url.URL{
		Scheme:   "https",
		Host:     g.GithubHostname,
		Path:     "/settings/apps/new",
		RawQuery: url.Values{"state": {state}}.Encode(),
	}

	// https://developer.github.com/apps/building-github-apps/creating-github-apps-using-url-parameters/#about-github-app-url-parameters
	if g.GithubOrg != "" {
		target.Path = fmt.Sprintf("organizations/%s%s", g.GithubOrg, target.Path)
	}

	jsonManifest, err := json.MarshalIndent(manifest, "", " ")
	if err != nil {
		g.respond(w, logging.Error, http.StatusBadRequest, "Failed to serialize manifest: %s", err)
		return
	}

	g.setBindingCookie(w, r, binding, int(githubAppStateTTL.Seconds()))
	err = web_templates.GithubAppSetupTemplate.Execute(w, web_templates.GithubSetupData{
		Target:   target.String(),
		Manifest: string(jsonManifest),
	})
	if err != nil {
		g.Logger.Err("%s", err.Error())
	}
}

// SetupComplete reports whether GitHub App credentials have been obtained.
func (g *GithubAppController) SetupComplete() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.GithubSetupComplete
}

// issueFlow records a new pending flow and returns its state and binding.
func (g *GithubAppController) issueFlow() (string, string, *githubAppFlowError) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if err := g.busyLocked(); err != nil {
		return "", "", err
	}
	g.pruneExpiredLocked()
	if len(g.flows) >= githubAppMaxPendingFlows {
		return "", "", errGithubSetupFull
	}

	state, err := randomToken()
	if err != nil {
		return "", "", &githubAppFlowError{http.StatusInternalServerError, fmt.Sprintf("Failed to generate setup state: %s", err)}
	}
	binding, err := randomToken()
	if err != nil {
		return "", "", &githubAppFlowError{http.StatusInternalServerError, fmt.Sprintf("Failed to generate setup binding: %s", err)}
	}
	if g.flows == nil {
		g.flows = map[string]githubAppFlow{}
	}
	g.flows[state] = githubAppFlow{binding: binding, expiry: g.clock().Add(githubAppStateTTL)}
	return state, binding, nil
}

// claimFlow moves the controller from idle to exchanging when state names an
// unexpired pending flow whose binding matches. Any other request leaves the
// pending flows as they are, except that an expired flow is dropped.
func (g *GithubAppController) claimFlow(state, binding string) *githubAppFlowError {
	g.mu.Lock()
	defer g.mu.Unlock()
	if err := g.busyLocked(); err != nil {
		return err
	}
	if state == "" {
		return errGithubStateMissing
	}
	flow, ok := g.flows[state]
	if !ok {
		return errGithubStateUnknown
	}
	if g.clock().After(flow.expiry) {
		delete(g.flows, state)
		return errGithubStateExpired
	}
	if subtle.ConstantTimeCompare([]byte(binding), []byte(flow.binding)) != 1 {
		return errGithubStateUnknown
	}

	g.exchanging = true
	g.flows = nil
	return nil
}

// exchangeClaimed exchanges the code of the claimed flow and always leaves
// the exchanging state, even if the exchanger panics.
func (g *GithubAppController) exchangeClaimed(code string) (app *github.GithubAppTemporarySecrets, err error) {
	succeeded := false
	defer func() { g.finishExchange(succeeded) }()
	app, err = g.exchange(code)
	succeeded = err == nil
	return app, err
}

// finishExchange leaves the exchanging state: to complete on success, back
// to idle on failure.
func (g *GithubAppController) finishExchange(succeeded bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.exchanging = false
	if succeeded {
		g.GithubSetupComplete = true
	}
}

// busyLocked returns the error for the complete and exchanging states.
func (g *GithubAppController) busyLocked() *githubAppFlowError {
	switch {
	case g.GithubSetupComplete:
		return errGithubSetupComplete
	case g.exchanging:
		return errGithubSetupBusy
	}
	return nil
}

func (g *GithubAppController) pruneExpiredLocked() {
	now := g.clock()
	for state, flow := range g.flows {
		if now.After(flow.expiry) {
			delete(g.flows, state)
		}
	}
}

func (g *GithubAppController) exchange(code string) (*github.GithubAppTemporarySecrets, error) {
	exchanger := g.Exchanger
	if exchanger == nil {
		creds := &github.AnonymousCredentials{}
		config := github.Config{}
		// This client does not post comments, so we don't need to configure it with maxCommentsPerCommand.
		client, err := github.New(g.GithubHostname, creds, config, 0, g.Logger)
		if err != nil {
			return nil, err
		}
		exchanger = client
	}
	return exchanger.ExchangeCode(g.Logger, code)
}

// setBindingCookie writes the flow-binding cookie. SameSite=Lax is required
// because GitHub sends the browser back with a top-level cross-site GET.
func (g *GithubAppController) setBindingCookie(w http.ResponseWriter, r *http.Request, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     GithubAppSetupCookieName,
		Value:    value,
		Path:     strings.TrimSuffix(g.AtlantisURL.Path, "/") + "/github-app/",
		MaxAge:   maxAge,
		Secure:   r.TLS != nil || strings.EqualFold(g.AtlantisURL.Scheme, "https"),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (g *GithubAppController) clearBindingCookie(w http.ResponseWriter, r *http.Request) {
	g.setBindingCookie(w, r, "", -1)
}

func randomToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func (g *GithubAppController) clock() time.Time {
	if g.now != nil {
		return g.now()
	}
	return time.Now()
}

func (g *GithubAppController) respond(w http.ResponseWriter, lvl logging.LogLevel, code int, format string, args ...any) {
	response := fmt.Sprintf(format, args...)
	g.Logger.Log(lvl, response)
	w.WriteHeader(code)
	fmt.Fprintln(w, response)
}
