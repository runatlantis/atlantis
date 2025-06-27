package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/runatlantis/atlantis/server/logging"
)

// AuthManager handles authentication, session, and provider management for Atlantis.
type AuthManager struct {
	config           *Config
	providers        map[string]Provider
	sessions         map[string]*Session
	sessionMux       sync.RWMutex
	logger           logging.SimpleLogging
	permissionChecker PermissionChecker
}

// NewManager creates a new Manager instance
func NewManager(config *Config, logger logging.SimpleLogging) (Manager, error) {
	// Create permission checker
	permissionChecker := NewPermissionChecker(config.Roles)

	manager := &AuthManager{
		config:            config,
		providers:         make(map[string]Provider),
		sessions:          make(map[string]*Session),
		logger:            logger,
		permissionChecker: permissionChecker,
	}

	// Initialize providers
	for _, providerConfig := range config.Providers {
		if !providerConfig.Enabled {
			continue
		}

		var provider Provider
		var err error

		switch providerConfig.Type {
		case ProviderTypeOAuth2, ProviderTypeOIDC:
			provider, err = NewOAuth2Provider(&providerConfig)
		case ProviderTypeBasicAuth:
			provider, err = NewBasicAuthProvider(&providerConfig)
		default:
			logger.Warn("unsupported provider type: %s", providerConfig.Type)
			continue
		}

		if err != nil {
			return nil, fmt.Errorf("failed to create provider %s: %w", providerConfig.ID, err)
		}

		manager.providers[providerConfig.ID] = provider
		logger.Info("initialized auth provider: %s (%s)", providerConfig.Name, providerConfig.Type)
	}

	// Add basic auth provider if enabled
	if config.EnableBasicAuth {
		basicConfig := ProviderConfig{
			Type:         ProviderTypeBasicAuth,
			ID:           "basic",
			Name:         "Basic Authentication",
			ClientID:     config.BasicAuthUser,
			ClientSecret: config.BasicAuthPass,
			Enabled:      true,
			DefaultRoles: config.DefaultRoles,
		}

		basicProvider, err := NewBasicAuthProvider(&basicConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create basic auth provider: %w", err)
		}

		manager.providers["basic"] = basicProvider
		logger.Info("initialized basic auth provider")
	}

	return manager, nil
}

// GetProvider returns a provider by ID
func (m *AuthManager) GetProvider(id string) (Provider, error) {
	provider, exists := m.providers[id]
	if !exists {
		return nil, fmt.Errorf("provider not found: %s", id)
	}
	return provider, nil
}

// GetEnabledProviders returns all enabled providers
func (m *AuthManager) GetEnabledProviders() []Provider {
	providers := make([]Provider, 0, len(m.providers))
	for _, provider := range m.providers {
		if provider.IsEnabled() {
			providers = append(providers, provider)
		}
	}
	return providers
}

// GetPermissionChecker returns the permission checker
func (m *AuthManager) GetPermissionChecker() PermissionChecker {
	return m.permissionChecker
}

// AuthenticateUser authenticates a user and creates a session
func (m *AuthManager) AuthenticateUser(_ context.Context, user *User) (*Session, error) {
	// Generate session ID
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	// Map user roles and permissions
	m.mapUserRolesAndPermissions(user)

	// Create session
	session := &Session{
		ID:        sessionID,
		UserID:    user.ID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(m.config.SessionDuration),
	}

	// Store session
	m.sessionMux.Lock()
	m.sessions[sessionID] = session
	m.sessionMux.Unlock()

	// Set session ID in user
	user.SessionID = sessionID

	m.logger.Debug("created session for user %s: %s", user.Email, sessionID)
	return session, nil
}

// mapUserRolesAndPermissions maps user attributes to roles and permissions
func (m *AuthManager) mapUserRolesAndPermissions(user *User) {
	// Start with default roles
	user.Roles = make([]string, 0)
	user.Roles = append(user.Roles, m.config.DefaultRoles...)

	// Check if user is in admin groups
	for _, group := range user.Groups {
		for _, adminGroup := range m.config.AdminGroups {
			if group == adminGroup {
				user.Roles = append(user.Roles, "admin")
				break
			}
		}
	}

	// Check if user email is in admin emails
	for _, adminEmail := range m.config.AdminEmails {
		if user.Email == adminEmail {
			user.Roles = append(user.Roles, "admin")
			break
		}
	}

	// Get permissions from roles
	user.Permissions = m.permissionChecker.GetUserPermissions(user)

	m.logger.Debug("mapped user %s to roles: %v, permissions: %v", user.Email, user.Roles, user.Permissions)
}

// ValidateSession validates a session and returns the user
func (m *AuthManager) ValidateSession(_ context.Context, sessionID string) (*User, error) {
	m.sessionMux.RLock()
	session, exists := m.sessions[sessionID]
	m.sessionMux.RUnlock()

	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	if time.Now().After(session.ExpiresAt) {
		// Clean up expired session
		m.sessionMux.Lock()
		delete(m.sessions, sessionID)
		m.sessionMux.Unlock()
		return nil, fmt.Errorf("session expired")
	}

	// In a real implementation, you would store user information in a database
	// For now, we'll return a basic user object
	user := &User{
		ID:        session.UserID,
		SessionID: sessionID,
		LastLogin: session.CreatedAt,
	}

	// Map roles and permissions
	m.mapUserRolesAndPermissions(user)

	return user, nil
}

// InvalidateSession invalidates a session
func (m *AuthManager) InvalidateSession(_ context.Context, sessionID string) error {
	m.sessionMux.Lock()
	defer m.sessionMux.Unlock()

	if _, exists := m.sessions[sessionID]; !exists {
		// Session doesn't exist, which is fine - no error
		return nil
	}

	delete(m.sessions, sessionID)
	m.logger.Debug("invalidated session: %s", sessionID)
	return nil
}

// GetUserFromRequest extracts user from HTTP request
func (m *AuthManager) GetUserFromRequest(r *http.Request) (*User, error) {
	// First, try to get user from session cookie
	if user, err := m.getUserFromSession(r); err == nil {
		return user, nil
	}

	// Then, try basic auth
	if user, err := m.getUserFromBasicAuth(r); err == nil {
		return user, nil
	}

	return nil, fmt.Errorf("no valid authentication found")
}

// LoginRequired checks if login is required for the request
func (m *AuthManager) LoginRequired(r *http.Request) bool {
	// Skip authentication for certain paths
	skipPaths := []string{
		"/events",
		"/healthz",
		"/status",
		"/auth/callback",
		"/auth/login",
		"/static/",
	}

	path := r.URL.Path
	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return false
		}
	}

	// Skip authentication for API endpoints if they have valid API secret
	if strings.HasPrefix(path, "/api/") {
		// API endpoints use their own authentication
		return false
	}

	// Check if user is already authenticated
	if _, err := m.GetUserFromRequest(r); err == nil {
		return false
	}

	return true
}

// RedirectToLogin redirects to the appropriate login provider
func (m *AuthManager) RedirectToLogin(w http.ResponseWriter, r *http.Request) error {
	providers := m.GetEnabledProviders()
	if len(providers) == 0 {
		return fmt.Errorf("no authentication providers available")
	}

	// If only one provider, redirect directly
	if len(providers) == 1 {
		provider := providers[0]
		if provider.GetType() == ProviderTypeOAuth2 || provider.GetType() == ProviderTypeOIDC {
			state, err := generateState()
			if err != nil {
				return fmt.Errorf("failed to generate state: %w", err)
			}

			authURL, err := provider.InitAuthURL(state)
			if err != nil {
				return fmt.Errorf("failed to generate auth URL: %w", err)
			}

			http.Redirect(w, r, authURL, http.StatusFound)
			return nil
		}
	}

	// Multiple providers - show login page
	http.Redirect(w, r, "/login", http.StatusFound)
	return nil
}

// getUserFromSession extracts user from session cookie
func (m *AuthManager) getUserFromSession(r *http.Request) (*User, error) {
	cookie, err := r.Cookie(m.config.SessionCookieName)
	if err != nil {
		return nil, fmt.Errorf("no session cookie: %w", err)
	}

	user, err := m.ValidateSession(r.Context(), cookie.Value)
	if err != nil {
		return nil, fmt.Errorf("invalid session: %w", err)
	}

	return user, nil
}

// getUserFromBasicAuth extracts user from basic auth
func (m *AuthManager) getUserFromBasicAuth(r *http.Request) (*User, error) {
	basicProvider, exists := m.providers["basic"]
	if !exists {
		return nil, fmt.Errorf("basic auth provider not available")
	}

	basicAuthProvider, ok := basicProvider.(*BasicAuthProvider)
	if !ok {
		return nil, fmt.Errorf("invalid basic auth provider")
	}

	user, err := basicAuthProvider.ValidateBasicAuth(r)
	if err != nil {
		return nil, fmt.Errorf("basic auth validation failed: %w", err)
	}

	return user, nil
}

// SetSessionCookie sets the session cookie
func (m *AuthManager) SetSessionCookie(w http.ResponseWriter, sessionID string) {
	cookie := &http.Cookie{
		Name:     m.config.SessionCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   m.config.SecureCookies,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(m.config.SessionDuration),
	}

	http.SetCookie(w, cookie)
}

// ClearSessionCookie clears the session cookie
func (m *AuthManager) ClearSessionCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     m.config.SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   m.config.SecureCookies,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(-24 * time.Hour), // Expire in the past
	}

	http.SetCookie(w, cookie)
}

// Helper functions
func generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func generateState() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
} 