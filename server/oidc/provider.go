// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"github.com/runatlantis/atlantis/server/logging"
)

// OIDCAuthProvider represents a supported OIDC authentication provider.
type OIDCAuthProvider string

const (
	// OIDCAuthEntraID enables Entra ID (Azure AD) as the OIDC provider.
	OIDCAuthEntraID OIDCAuthProvider = "entraid"
)

// ValidOIDCAuthProviders contains all supported provider values.
var ValidOIDCAuthProviders = []OIDCAuthProvider{OIDCAuthEntraID}

// IsValidOIDCAuthProvider returns true if the given string is a supported provider.
func IsValidOIDCAuthProvider(s string) bool {
	return slices.Contains(ValidOIDCAuthProviders, OIDCAuthProvider(s))
}

const (
	// clientAssertionType is the OAuth2 parameter value for JWT bearer
	// client assertion, used with Azure workload identity.
	clientAssertionType = "urn:ietf:params:oauth:client-assertion-type:jwt-bearer"
)

// Provider wraps the OIDC discovery provider and OAuth2 configuration needed
// to perform the authorization code flow.
type Provider struct {
	oidcProvider *gooidc.Provider
	oauth2Config oauth2.Config
	verifier     *gooidc.IDTokenVerifier
	azureWI      *AzureWorkloadIdentity
	logger       logging.SimpleLogging
}

// NewProvider performs OIDC discovery against the issuer URL and returns a
// configured Provider ready for authorization code flow.
func NewProvider(ctx context.Context, issuerURL, clientID, clientSecret, redirectURL string, scopes []string, azureWI *AzureWorkloadIdentity, logger logging.SimpleLogging) (*Provider, error) {
	oidcProvider, err := gooidc.NewProvider(ctx, issuerURL)
	if err != nil {
		return nil, fmt.Errorf("running OIDC discovery for %q: %w", issuerURL, err)
	}

	oauth2Config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     oidcProvider.Endpoint(),
		Scopes:       scopes,
	}

	verifier := oidcProvider.Verifier(&gooidc.Config{
		ClientID: clientID,
	})

	logger.Info("OIDC provider initialized - issuer=%q clientID=%q", issuerURL, clientID)

	return &Provider{
		oidcProvider: oidcProvider,
		oauth2Config: oauth2Config,
		verifier:     verifier,
		azureWI:      azureWI,
		logger:       logger,
	}, nil
}

// AuthCodeURL generates the authorization URL that redirects the user to the
// identity provider for authentication.
func (p *Provider) AuthCodeURL(state string) string {
	return p.oauth2Config.AuthCodeURL(state)
}

// Exchange trades an authorization code for tokens. When Azure workload
// identity is configured, it passes the federated service account token as
// a client_assertion instead of using a client_secret.
//
// Returns the verified ID token, the raw ID token string, and any error.
func (p *Provider) Exchange(ctx context.Context, code string) (*gooidc.IDToken, string, error) {
	var options []oauth2.AuthCodeOption

	if p.azureWI != nil {
		clientAssertion, err := p.azureWI.GetFederatedToken()
		if err != nil {
			return nil, "", fmt.Errorf("getting federated token: %w", err)
		}
		options = []oauth2.AuthCodeOption{
			oauth2.SetAuthURLParam("client_assertion_type", clientAssertionType),
			oauth2.SetAuthURLParam("client_assertion", clientAssertion),
		}
	}

	token, err := p.oauth2Config.Exchange(ctx, code, options...)
	if err != nil {
		return nil, "", fmt.Errorf("exchanging authorization code: %w", err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, "", fmt.Errorf("no id_token in token response")
	}

	idToken, err := p.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, "", fmt.Errorf("verifying id_token: %w", err)
	}

	return idToken, rawIDToken, nil
}

// Verify validates a raw ID token string and returns the parsed token.
// Used for verifying session tokens on subsequent requests.
func (p *Provider) Verify(ctx context.Context, rawIDToken string) (*gooidc.IDToken, error) {
	return p.verifier.Verify(ctx, rawIDToken)
}

// ExtractEmail extracts the email claim from a raw JWT ID token by decoding
// the payload without signature verification (the token was already verified
// when the session was established). Falls back to preferred_username for
// Entra ID tokens that may not include an email claim.
func ExtractEmail(rawIDToken string) string {
	parts := strings.Split(rawIDToken, ".")
	if len(parts) != 3 {
		return ""
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}
	var claims struct {
		Email             string `json:"email"`
		PreferredUsername string `json:"preferred_username"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return ""
	}
	if claims.Email != "" {
		return claims.Email
	}
	return claims.PreferredUsername
}
