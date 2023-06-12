package controllers

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/runatlantis/atlantis/server/logging"
)

// OIDCController handles OIDC requests.
type OIDCController struct {
	AtlantisURL *url.URL
	Logger      logging.SimpleLogging
}

// https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderConfig
func (l *OIDCController) GetOpenIDConfiguration(w http.ResponseWriter, r *http.Request) {
	l.respond(w, logging.Info, http.StatusOK, `{
		"issuer": "%s",
		"authorization_endpoint": "%s",
		"token_endpoint": "%s",
		"userinfo_endpoint": "%s",
		"jwks_uri": "%s",
		"response_types_supported": [
			"code",
			"id_token",
			"code id_token",
			"code token",
			"id_token token",
			"code id_token token"
		],
		"subject_types_supported": [
			"public"
		],
		"id_token_signing_alg_values_supported": [
			"RS256"
		],
		"scopes_supported": [
			"openid",
			"profile",
			"email"
		],
		"token_endpoint_auth_methods_supported": [
			"client_secret_post",
			"client_secret_basic"
		],
		"claims_supported": [
			"aud",
			"email",
			"email_verified",
			"exp",
			"iat",
			"iss",
			"name",
			"sub"
		]
	}`, l.AtlantisURL.String(), l.AtlantisURL.String()+"/oidc/auth", l.AtlantisURL.String()+"/oidc/token", l.AtlantisURL.String()+"/oidc/userinfo", l.AtlantisURL.String()+"/oidc/certs")
}

// respond is a helper function to respond and log the response. lvl is the log
// level to log at, code is the HTTP response code.
func (l *OIDCController) respond(w http.ResponseWriter, lvl logging.LogLevel, responseCode int, format string, args ...interface{}) {
	response := fmt.Sprintf(format, args...)
	l.Logger.Log(lvl, response)
	w.WriteHeader(responseCode)
	fmt.Fprintln(w, response)
}
