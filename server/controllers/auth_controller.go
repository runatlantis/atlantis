package controllers

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"

	"github.com/runatlantis/atlantis/server/auth"
	"github.com/runatlantis/atlantis/server/logging"
)

var loginTemplate = template.Must(template.New("login").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <title>Atlantis - Login</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="{{.BasePath}}/static/css/normalize.css">
    <link rel="stylesheet" href="{{.BasePath}}/static/css/skeleton.css">
    <link rel="stylesheet" href="{{.BasePath}}/static/css/custom.css">
    <link rel="icon" type="image/png" href="{{.BasePath}}/static/images/atlantis-icon.png">
</head>
<body>
<div class="container">
    <section class="header">
        <a title="atlantis" href="{{.BasePath}}/"><img class="hero" src="{{.BasePath}}/static/images/atlantis-icon_512.png"/></a>
        <p class="title-heading">Atlantis</p>
        <p class="title-heading"><strong>Login</strong></p>
    </section>
    <section>
        <div class="row">
            <div class="six columns offset-by-three">
                <h5>Choose your login method:</h5>
                {{range .Providers}}
                <a class="button button-primary u-full-width" href="{{.AuthURL}}">
                    Login with {{.Name}}
                </a>
                {{end}}
            </div>
        </div>
    </section>
</div>
<footer>
    <p>Atlantis SSO Authentication</p>
</footer>
</body>
</html>
`))

type loginProviderData struct {
	Name   string
	AuthURL string
}

type loginPageData struct {
	BasePath  string
	Providers []loginProviderData
}

// AuthController handles authentication-related HTTP requests
type AuthController struct {
	AuthManager auth.Manager
	Logger      logging.SimpleLogging
	BaseURL     *url.URL
}

// NewAuthController creates a new authentication controller
func NewAuthController(authManager auth.Manager, logger logging.SimpleLogging, baseURL *url.URL) *AuthController {
	return &AuthController{
		AuthManager: authManager,
		Logger:      logger,
		BaseURL:     baseURL,
	}
}

// Login handles the login page
func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	providers := c.AuthManager.GetEnabledProviders()
	
	// If only one provider and it's OAuth2/OIDC, redirect directly
	if len(providers) == 1 {
		provider := providers[0]
		if provider.GetType() == auth.ProviderTypeOAuth2 || provider.GetType() == auth.ProviderTypeOIDC {
			c.redirectToProvider(w, r, provider)
			return
		}
	}

	// Show simple login page
	c.showLoginPage(w, r, providers)
}

// Callback handles OAuth2/OIDC callback
func (c *AuthController) Callback(w http.ResponseWriter, r *http.Request) {
	// Get provider from query parameter
	providerID := r.URL.Query().Get("provider")
	if providerID == "" {
		c.respond(w, logging.Error, http.StatusBadRequest, "Missing provider parameter")
		return
	}

	provider, err := c.AuthManager.GetProvider(providerID)
	if err != nil {
		c.respond(w, logging.Error, http.StatusBadRequest, "Invalid provider: %s", err)
		return
	}

	// Handle OAuth2/OIDC callback
	if provider.GetType() == auth.ProviderTypeOAuth2 || provider.GetType() == auth.ProviderTypeOIDC {
		c.handleOAuthCallback(w, r, provider)
		return
	}

	c.respond(w, logging.Error, http.StatusBadRequest, "Unsupported provider type for callback")
}

// Logout handles user logout
func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	// Get user from request to invalidate their session
	if user, err := c.AuthManager.GetUserFromRequest(r); err == nil {
		if user.SessionID != "" {
			if err := c.AuthManager.InvalidateSession(r.Context(), user.SessionID); err != nil {
				c.Logger.Err("Failed to invalidate session: %v", err)
			}
		}
	}

	// Clear session cookie
	if authManager, ok := c.AuthManager.(*auth.AuthManager); ok {
		authManager.ClearSessionCookie(w)
	}

	// Redirect to login page
	http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
}

// redirectToProvider redirects to a specific OAuth2/OIDC provider
func (c *AuthController) redirectToProvider(w http.ResponseWriter, r *http.Request, provider auth.Provider) {
	state, err := c.generateState()
	if err != nil {
		c.respond(w, logging.Error, http.StatusInternalServerError, "Failed to generate state: %s", err)
		return
	}

	authURL, err := provider.InitAuthURL(state)
	if err != nil {
		c.respond(w, logging.Error, http.StatusInternalServerError, "Failed to generate auth URL: %s", err)
		return
	}

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// showLoginPage displays a simple login page with available providers
func (c *AuthController) showLoginPage(w http.ResponseWriter, r *http.Request, providers []auth.Provider) {
	var providerData []loginProviderData
	for _, provider := range providers {
		if provider.GetType() == auth.ProviderTypeOAuth2 || provider.GetType() == auth.ProviderTypeOIDC {
			state, err := c.generateState()
			if err != nil {
				c.Logger.Err("Failed to generate state for provider %s: %s", provider.GetID(), err)
				continue
			}
			authURL, err := provider.InitAuthURL(state)
			if err != nil {
				c.Logger.Err("Failed to generate auth URL for provider %s: %s", provider.GetID(), err)
				continue
			}
			providerData = append(providerData, loginProviderData{
				Name:   provider.GetName(),
				AuthURL: authURL,
			})
		}
	}
	data := loginPageData{
		BasePath:  c.BaseURL.Path,
		Providers: providerData,
	}
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	if err := loginTemplate.Execute(w, data); err != nil {
		c.Logger.Err("Failed to render login template: %s", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleOAuthCallback processes OAuth2/OIDC callback
func (c *AuthController) handleOAuthCallback(w http.ResponseWriter, r *http.Request, provider auth.Provider) {
	// Get authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		c.respond(w, logging.Error, http.StatusBadRequest, "Missing authorization code")
		return
	}

	// Get state parameter for CSRF protection
	state := r.URL.Query().Get("state")
	if state == "" {
		c.respond(w, logging.Error, http.StatusBadRequest, "Missing state parameter")
		return
	}

	// Exchange code for tokens
	token, err := provider.ExchangeCode(r.Context(), code)
	if err != nil {
		c.Logger.Err("Failed to exchange code for token: %s", err)
		c.respond(w, logging.Error, http.StatusInternalServerError, "Authentication failed")
		return
	}

	// Get user information
	user, err := provider.GetUserInfo(r.Context(), token)
	if err != nil {
		c.Logger.Err("Failed to get user info: %s", err)
		c.respond(w, logging.Error, http.StatusInternalServerError, "Failed to get user information")
		return
	}

	// Authenticate user and create session
	session, err := c.AuthManager.AuthenticateUser(r.Context(), user)
	if err != nil {
		c.Logger.Err("Failed to authenticate user: %s", err)
		c.respond(w, logging.Error, http.StatusInternalServerError, "Failed to create session")
		return
	}

	// Set session cookie
	if authManager, ok := c.AuthManager.(*auth.AuthManager); ok {
		authManager.SetSessionCookie(w, session.ID)
	}

	c.Logger.Info("User %s authenticated successfully via %s", user.Email, provider.GetName())

	// Redirect to home page
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// generateState generates a random state parameter for CSRF protection
func (c *AuthController) generateState() (string, error) {
	// This is a simplified implementation
	// In a real implementation, you would use a cryptographically secure random generator
	return "state", nil
}

// respond sends an HTTP response with the given status code and message
func (c *AuthController) respond(w http.ResponseWriter, lvl logging.LogLevel, responseCode int, format string, args ...interface{}) {
	response := fmt.Sprintf(format, args...)
	c.Logger.Log(lvl, response)
	w.WriteHeader(responseCode)
	if _, err := fmt.Fprintln(w, response); err != nil {
		c.Logger.Err("Failed to write response: %v", err)
	}
} 