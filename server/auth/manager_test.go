package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/logging"
)

func TestNewManager(t *testing.T) {
	config := &Config{
		SessionSecret:     "test-secret",
		SessionDuration:   24 * time.Hour,
		SessionCookieName: "atlantis_session",
		SecureCookies:     false,
		CSRFSecret:        "csrf-secret",
		EnableBasicAuth:   true,
		BasicAuthUser:     "atlantis",
		BasicAuthPass:     "atlantis",
		DefaultRoles:      []string{"user"},
		AdminGroups:       []string{"admin"},
		AdminEmails:       []string{"admin@example.com"},
		Roles:             make(map[string]Role),
		Providers:         []ProviderConfig{},
	}

	logger := logging.NewNoopLogger(t)
	manager, err := NewManager(config, logger)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	if manager == nil {
		t.Fatal("Manager should not be nil")
	}
}

func TestManager_AuthenticateUser(t *testing.T) {
	config := &Config{
		SessionSecret:     "test-secret",
		SessionDuration:   24 * time.Hour,
		SessionCookieName: "atlantis_session",
		SecureCookies:     false,
		CSRFSecret:        "csrf-secret",
		EnableBasicAuth:   true,
		BasicAuthUser:     "atlantis",
		BasicAuthPass:     "atlantis",
		DefaultRoles:      []string{"user"},
		AdminGroups:       []string{"admin"},
		AdminEmails:       []string{"admin@example.com"},
		Roles:             make(map[string]Role),
		Providers:         []ProviderConfig{},
	}

	logger := logging.NewNoopLogger(t)
	manager, err := NewManager(config, logger)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	user := &User{
		ID:         "test-user",
		Email:      "test@example.com",
		Name:       "Test User",
		Groups:     []string{"developers"},
		Roles:      []string{"user"},
		Permissions: []Permission{},
		Provider:   "google",
		LastLogin:  time.Now(),
	}

	ctx := context.Background()
	session, err := manager.AuthenticateUser(ctx, user)
	if err != nil {
		t.Errorf("Failed to authenticate user: %v", err)
	}

	if session == nil {
		t.Fatal("Session should not be nil")
	}

	if session.UserID != user.ID {
		t.Errorf("Session UserID = %s, want %s", session.UserID, user.ID)
	}

	if session.CreatedAt.IsZero() {
		t.Error("Session CreatedAt should not be zero")
	}

	if session.ExpiresAt.Before(time.Now()) {
		t.Error("Session ExpiresAt should be in the future")
	}
}

func TestManager_ValidateSession(t *testing.T) {
	config := &Config{
		SessionSecret:     "test-secret",
		SessionDuration:   24 * time.Hour,
		SessionCookieName: "atlantis_session",
		SecureCookies:     false,
		CSRFSecret:        "csrf-secret",
		EnableBasicAuth:   true,
		BasicAuthUser:     "atlantis",
		BasicAuthPass:     "atlantis",
		DefaultRoles:      []string{"user"},
		AdminGroups:       []string{"admin"},
		AdminEmails:       []string{"admin@example.com"},
		Roles:             make(map[string]Role),
		Providers:         []ProviderConfig{},
	}

	logger := logging.NewNoopLogger(t)
	manager, err := NewManager(config, logger)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test validating non-existent session
	ctx := context.Background()
	user, err := manager.ValidateSession(ctx, "non-existent")
	if err == nil {
		t.Error("Should return error for non-existent session")
	}
	if user != nil {
		t.Error("Should return nil user for non-existent session")
	}

	// Test validating existing session
	testUser := &User{
		ID:         "test-user",
		Email:      "test@example.com",
		Name:       "Test User",
		Groups:     []string{"developers"},
		Roles:      []string{"user"},
		Permissions: []Permission{},
		Provider:   "google",
		LastLogin:  time.Now(),
	}

	session, err := manager.AuthenticateUser(ctx, testUser)
	if err != nil {
		t.Fatalf("Failed to authenticate user: %v", err)
	}

	validatedUser, err := manager.ValidateSession(ctx, session.ID)
	if err != nil {
		t.Errorf("Failed to validate session: %v", err)
	}

	if validatedUser == nil {
		t.Error("Should return user for valid session")
	}

	if validatedUser.ID != testUser.ID {
		t.Errorf("Validated user ID = %s, want %s", validatedUser.ID, testUser.ID)
	}
}

func TestManager_InvalidateSession(t *testing.T) {
	config := &Config{
		SessionSecret:     "test-secret",
		SessionDuration:   24 * time.Hour,
		SessionCookieName: "atlantis_session",
		SecureCookies:     false,
		CSRFSecret:        "csrf-secret",
		EnableBasicAuth:   true,
		BasicAuthUser:     "atlantis",
		BasicAuthPass:     "atlantis",
		DefaultRoles:      []string{"user"},
		AdminGroups:       []string{"admin"},
		AdminEmails:       []string{"admin@example.com"},
		Roles:             make(map[string]Role),
		Providers:         []ProviderConfig{},
	}

	logger := logging.NewNoopLogger(t)
	manager, err := NewManager(config, logger)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	user := &User{
		ID:         "test-user",
		Email:      "test@example.com",
		Name:       "Test User",
		Groups:     []string{"developers"},
		Roles:      []string{"user"},
		Permissions: []Permission{},
		Provider:   "google",
		LastLogin:  time.Now(),
	}

	ctx := context.Background()
	session, err := manager.AuthenticateUser(ctx, user)
	if err != nil {
		t.Fatalf("Failed to authenticate user: %v", err)
	}

	// Test invalidate session
	err = manager.InvalidateSession(ctx, session.ID)
	if err != nil {
		t.Errorf("Failed to invalidate session: %v", err)
	}

	// Verify session is invalidated
	_, err = manager.ValidateSession(ctx, session.ID)
	if err == nil {
		t.Error("Session should be invalidated")
	}

	// Test invalidate non-existent session
	err = manager.InvalidateSession(ctx, "non-existent")
	if err != nil {
		t.Errorf("Invalidating non-existent session should not error: %v", err)
	}
}

func TestManager_GetUserFromRequest(t *testing.T) {
	config := &Config{
		SessionSecret:     "test-secret",
		SessionDuration:   24 * time.Hour,
		SessionCookieName: "atlantis_session",
		SecureCookies:     false,
		CSRFSecret:        "csrf-secret",
		EnableBasicAuth:   true,
		BasicAuthUser:     "atlantis",
		BasicAuthPass:     "atlantis",
		DefaultRoles:      []string{"user"},
		AdminGroups:       []string{"admin"},
		AdminEmails:       []string{"admin@example.com"},
		Roles:             make(map[string]Role),
		Providers:         []ProviderConfig{},
	}

	logger := logging.NewNoopLogger(t)
	manager, err := NewManager(config, logger)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test request without session cookie
	req := httptest.NewRequest("GET", "/", nil)
	user, err := manager.GetUserFromRequest(req)
	if err == nil {
		t.Error("Should return error for request without session cookie")
	}
	if user != nil {
		t.Error("Should return nil user for request without session cookie")
	}

	// Test request with session cookie
	testUser := &User{
		ID:         "test-user",
		Email:      "test@example.com",
		Name:       "Test User",
		Groups:     []string{"developers"},
		Roles:      []string{"user"},
		Permissions: []Permission{},
		Provider:   "google",
		LastLogin:  time.Now(),
	}

	ctx := context.Background()
	session, err := manager.AuthenticateUser(ctx, testUser)
	if err != nil {
		t.Fatalf("Failed to authenticate user: %v", err)
	}

	req = httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  config.SessionCookieName,
		Value: session.ID,
	})

	retrievedUser, err := manager.GetUserFromRequest(req)
	if err != nil {
		t.Errorf("Failed to get user from request: %v", err)
	}

	if retrievedUser == nil {
		t.Error("Should return user for request with valid session cookie")
	}

	if retrievedUser.ID != testUser.ID {
		t.Errorf("Retrieved user ID = %s, want %s", retrievedUser.ID, testUser.ID)
	}
}

func TestManager_LoginRequired(t *testing.T) {
	config := &Config{
		SessionSecret:     "test-secret",
		SessionDuration:   24 * time.Hour,
		SessionCookieName: "atlantis_session",
		SecureCookies:     false,
		CSRFSecret:        "csrf-secret",
		EnableBasicAuth:   true,
		BasicAuthUser:     "atlantis",
		BasicAuthPass:     "atlantis",
		DefaultRoles:      []string{"user"},
		AdminGroups:       []string{"admin"},
		AdminEmails:       []string{"admin@example.com"},
		Roles:             make(map[string]Role),
		Providers:         []ProviderConfig{},
	}

	logger := logging.NewNoopLogger(t)
	manager, err := NewManager(config, logger)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test request without session cookie
	req := httptest.NewRequest("GET", "/", nil)
	if !manager.LoginRequired(req) {
		t.Error("Should require login for request without session")
	}

	// Test request with valid session cookie
	user := &User{
		ID:         "test-user",
		Email:      "test@example.com",
		Name:       "Test User",
		Groups:     []string{"developers"},
		Roles:      []string{"user"},
		Permissions: []Permission{},
		Provider:   "google",
		LastLogin:  time.Now(),
	}

	ctx := context.Background()
	session, err := manager.AuthenticateUser(ctx, user)
	if err != nil {
		t.Fatalf("Failed to authenticate user: %v", err)
	}

	req = httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  config.SessionCookieName,
		Value: session.ID,
	})

	if manager.LoginRequired(req) {
		t.Error("Should not require login for request with valid session")
	}
}

func TestManager_RedirectToLogin(t *testing.T) {
	config := &Config{
		SessionSecret:     "test-secret",
		SessionDuration:   24 * time.Hour,
		SessionCookieName: "atlantis_session",
		SecureCookies:     false,
		CSRFSecret:        "csrf-secret",
		EnableBasicAuth:   true,
		BasicAuthUser:     "atlantis",
		BasicAuthPass:     "atlantis",
		DefaultRoles:      []string{"user"},
		AdminGroups:       []string{"admin"},
		AdminEmails:       []string{"admin@example.com"},
		Roles:             make(map[string]Role),
		Providers:         []ProviderConfig{},
	}

	logger := logging.NewNoopLogger(t)
	manager, err := NewManager(config, logger)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()

	err = manager.RedirectToLogin(w, req)
	if err != nil {
		t.Errorf("Failed to redirect to login: %v", err)
	}

	if w.Code != http.StatusFound {
		t.Errorf("Expected status %d, got %d", http.StatusFound, w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/login" {
		t.Errorf("Expected location /login, got %s", location)
	}
}

func TestManager_GetProvider(t *testing.T) {
	config := &Config{
		SessionSecret:     "test-secret",
		SessionDuration:   24 * time.Hour,
		SessionCookieName: "atlantis_session",
		SecureCookies:     false,
		CSRFSecret:        "csrf-secret",
		EnableBasicAuth:   true,
		BasicAuthUser:     "atlantis",
		BasicAuthPass:     "atlantis",
		DefaultRoles:      []string{"user"},
		AdminGroups:       []string{"admin"},
		AdminEmails:       []string{"admin@example.com"},
		Roles:             make(map[string]Role),
		Providers: []ProviderConfig{
			{
				Type:         ProviderTypeOAuth2,
				ID:           "google",
				Name:         "Google",
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				RedirectURL:  "https://example.com/callback",
				Scopes:       []string{"openid", "email", "profile"},
				Enabled:      true,
				DefaultRoles: []string{"user"},
			},
		},
	}

	logger := logging.NewNoopLogger(t)
	manager, err := NewManager(config, logger)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test getting existing provider
	provider, err := manager.GetProvider("google")
	if err != nil {
		t.Errorf("Failed to get provider: %v", err)
	}

	if provider == nil {
		t.Error("Should return provider for existing provider ID")
	}

	// Test getting non-existent provider
	provider, err = manager.GetProvider("non-existent")
	if err == nil {
		t.Error("Should return error for non-existent provider")
	}
	if provider != nil {
		t.Error("Should return nil for non-existent provider")
	}
}

func TestManager_GetEnabledProviders(t *testing.T) {
	config := &Config{
		SessionSecret:     "test-secret",
		SessionDuration:   24 * time.Hour,
		SessionCookieName: "atlantis_session",
		SecureCookies:     false,
		CSRFSecret:        "csrf-secret",
		EnableBasicAuth:   true,
		BasicAuthUser:     "atlantis",
		BasicAuthPass:     "atlantis",
		DefaultRoles:      []string{"user"},
		AdminGroups:       []string{"admin"},
		AdminEmails:       []string{"admin@example.com"},
		Roles:             make(map[string]Role),
		Providers: []ProviderConfig{
			{
				Type:         ProviderTypeOAuth2,
				ID:           "google",
				Name:         "Google",
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				RedirectURL:  "https://example.com/callback",
				Scopes:       []string{"openid", "email", "profile"},
				Enabled:      true,
				DefaultRoles: []string{"user"},
			},
			{
				Type:         ProviderTypeOAuth2,
				ID:           "disabled",
				Name:         "Disabled",
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				RedirectURL:  "https://example.com/callback",
				Scopes:       []string{"openid", "email", "profile"},
				Enabled:      false,
				DefaultRoles: []string{"user"},
			},
		},
	}

	logger := logging.NewNoopLogger(t)
	manager, err := NewManager(config, logger)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	providers := manager.GetEnabledProviders()
	if len(providers) != 2 { // 1 OAuth2 + 1 Basic Auth
		t.Errorf("Expected 2 enabled providers, got %d", len(providers))
	}
} 