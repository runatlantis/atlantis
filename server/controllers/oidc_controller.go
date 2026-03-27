// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"fmt"
	"net/http"

	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/oidc"
)

// OIDCController handles the OIDC authentication flow for the Atlantis web UI.
type OIDCController struct {
	Provider       *oidc.Provider
	SessionManager *oidc.SessionManager
	Logger         logging.SimpleLogging
	BasePath       string
}

func (o *OIDCController) homeURL() string {
	if o.BasePath == "" || o.BasePath == "/" {
		return "/"
	}
	return o.BasePath + "/"
}

// Login initiates the OIDC authorization code flow by redirecting the user
// to the identity provider's authorization endpoint.
func (o *OIDCController) Login(w http.ResponseWriter, r *http.Request) {
	state, err := o.SessionManager.CreateState(w)
	if err != nil {
		o.Logger.Err("creating OIDC state - %s", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	authURL := o.Provider.AuthCodeURL(state)
	o.Logger.Debug("OIDC login - redirecting to %q", authURL)
	http.Redirect(w, r, authURL, http.StatusFound)
}

// Callback handles the OIDC callback from the identity provider after user
// authentication. It exchanges the authorization code for tokens, verifies
// the ID token, and establishes a session cookie.
func (o *OIDCController) Callback(w http.ResponseWriter, r *http.Request) {
	// Check for errors from the IDP.
	if errParam := r.URL.Query().Get("error"); errParam != "" {
		errDesc := r.URL.Query().Get("error_description")
		o.Logger.Err("OIDC callback error from IDP - %s - %s", errParam, errDesc)
		http.Error(w, fmt.Sprintf("Authentication error: %s - %s", errParam, errDesc), http.StatusBadRequest)
		return
	}

	// Verify state parameter.
	state := r.URL.Query().Get("state")
	if state == "" {
		http.Error(w, "Missing state parameter", http.StatusBadRequest)
		return
	}
	if err := o.SessionManager.VerifyState(r, state); err != nil {
		o.Logger.Err("verifying OIDC state - %s", err)
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Exchange authorization code for tokens.
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	_, rawIDToken, err := o.Provider.Exchange(r.Context(), code)
	if err != nil {
		o.Logger.Err("exchanging OIDC token - %s", err)
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
	}

	// Extract the email (or preferred_username) from the ID token to store
	// in the session cookie. This keeps the cookie small instead of
	// embedding the full raw ID token.
	email := oidc.ExtractEmail(rawIDToken)
	if email == "" {
		o.Logger.Err("OIDC callback - no email or preferred_username in ID token")
		http.Error(w, "Authentication failed: no user identity in token", http.StatusBadRequest)
		return
	}

	// Set session cookie with only the email claim.
	if err := o.SessionManager.SetSession(w, email); err != nil {
		o.Logger.Err("setting OIDC session - %s", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Clear the state cookie.
	o.SessionManager.ClearState(w)

	o.Logger.Info("OIDC login successful, redirecting to %q", o.homeURL())
	http.Redirect(w, r, o.homeURL(), http.StatusFound)
}

// Logout clears the OIDC session cookie and redirects to the home page.
func (o *OIDCController) Logout(w http.ResponseWriter, r *http.Request) {
	o.SessionManager.ClearSession(w)

	o.Logger.Info("OIDC logout, redirecting to %q", o.homeURL())
	http.Redirect(w, r, o.homeURL(), http.StatusFound)
}
