package controllers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/auth"
	"github.com/runatlantis/atlantis/server/logging"
)

// MockAuthManager implements the auth.Manager interface for testing
type MockAuthManager struct {
	enabledProviders []auth.Provider
	provider         auth.Provider
	providerError    error
	user             *auth.User
	userError        error
	authenticateUser func(ctx context.Context, user *auth.User) (*auth.Session, error)
}

func (m *MockAuthManager) AuthenticateUser(ctx context.Context, user *auth.User) (*auth.Session, error) {
	if m.authenticateUser != nil {
		return m.authenticateUser(ctx, user)
	}
	return &auth.Session{
		ID:        "test-session",
		UserID:    user.ID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}, nil
}

func (m *MockAuthManager) ValidateSession(ctx context.Context, sessionID string) (*auth.User, error) {
	return m.user, m.userError
}

func (m *MockAuthManager) InvalidateSession(ctx context.Context, sessionID string) error {
	return nil
}

func (m *MockAuthManager) GetUserFromRequest(r *http.Request) (*auth.User, error) {
	return m.user, m.userError
}

func (m *MockAuthManager) LoginRequired(r *http.Request) bool {
	return true
}

func (m *MockAuthManager) RedirectToLogin(w http.ResponseWriter, r *http.Request) error {
	http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
	return nil
}

func (m *MockAuthManager) GetProvider(providerID string) (auth.Provider, error) {
	return m.provider, m.providerError
}

func (m *MockAuthManager) GetEnabledProviders() []auth.Provider {
	return m.enabledProviders
}

func (m *MockAuthManager) GetPermissionChecker() auth.PermissionChecker {
	return auth.NewPermissionChecker(nil)
}

// MockProvider implements the auth.Provider interface for testing
type MockProvider struct {
	id       string
	name     string
	providerType auth.ProviderType
	enabled  bool
	authURL  string
	authError error
	user     *auth.User
	userError error
}

func (m *MockProvider) GetID() string {
	return m.id
}

func (m *MockProvider) GetName() string {
	return m.name
}

func (m *MockProvider) GetType() auth.ProviderType {
	return m.providerType
}

func (m *MockProvider) IsEnabled() bool {
	return m.enabled
}

func (m *MockProvider) InitAuthURL(state string) (string, error) {
	return m.authURL, m.authError
}

func (m *MockProvider) InitiateLogin(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (m *MockProvider) ExchangeCode(ctx context.Context, code string) (*auth.TokenResponse, error) {
	return nil, nil
}

func (m *MockProvider) GetUserInfo(ctx context.Context, token *auth.TokenResponse) (*auth.User, error) {
	return m.user, m.userError
}

func (m *MockProvider) ValidateToken(ctx context.Context, token string) (*auth.User, error) {
	return m.user, m.userError
}

func (m *MockProvider) ProcessSAMLResponse(w http.ResponseWriter, r *http.Request) (*auth.User, error) {
	return nil, nil
}

func TestNewAuthController(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	mockManager := &MockAuthManager{}
	baseURL, _ := url.Parse("https://example.com")
	
	controller := NewAuthController(mockManager, logger, baseURL)
	
	if controller == nil {
		t.Fatal("Controller should not be nil")
	}
	
	if controller.AuthManager != mockManager {
		t.Error("Auth manager should be set correctly")
	}
	
	if controller.Logger != logger {
		t.Error("Logger should be set correctly")
	}
	
	if controller.BaseURL != baseURL {
		t.Error("Base URL should be set correctly")
	}
}

func TestAuthController_Login_SingleOAuthProvider(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	mockProvider := &MockProvider{
		id:       "google",
		name:     "Google",
		providerType: auth.ProviderTypeOAuth2,
		enabled:  true,
		authURL:  "https://accounts.google.com/oauth/authorize",
	}
	
	mockManager := &MockAuthManager{
		enabledProviders: []auth.Provider{mockProvider},
	}
	
	baseURL, _ := url.Parse("https://example.com")
	controller := NewAuthController(mockManager, logger, baseURL)
	
	req := httptest.NewRequest("GET", "/auth/login", nil)
	w := httptest.NewRecorder()
	
	controller.Login(w, req)
	
	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected status %d, got %d", http.StatusTemporaryRedirect, w.Code)
	}
	
	location := w.Header().Get("Location")
	if location == "" {
		t.Error("Location header should be set for redirect")
	}
}

func TestAuthController_Login_MultipleProviders(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	mockProvider1 := &MockProvider{
		id:       "google",
		name:     "Google",
		providerType: auth.ProviderTypeOAuth2,
		enabled:  true,
		authURL:  "https://accounts.google.com/oauth/authorize",
	}
	
	mockProvider2 := &MockProvider{
		id:       "basic",
		name:     "Basic Auth",
		providerType: auth.ProviderTypeBasicAuth,
		enabled:  true,
	}
	
	mockManager := &MockAuthManager{
		enabledProviders: []auth.Provider{mockProvider1, mockProvider2},
	}
	
	baseURL, _ := url.Parse("https://example.com")
	controller := NewAuthController(mockManager, logger, baseURL)
	
	req := httptest.NewRequest("GET", "/auth/login", nil)
	w := httptest.NewRecorder()
	
	controller.Login(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	contentType := w.Header().Get("Content-Type")
	if contentType != "text/html" {
		t.Errorf("Expected Content-Type text/html, got %s", contentType)
	}
}

func TestAuthController_Login_NoProviders(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	mockManager := &MockAuthManager{
		enabledProviders: []auth.Provider{},
	}
	
	baseURL, _ := url.Parse("https://example.com")
	controller := NewAuthController(mockManager, logger, baseURL)
	
	req := httptest.NewRequest("GET", "/auth/login", nil)
	w := httptest.NewRecorder()
	
	controller.Login(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	contentType := w.Header().Get("Content-Type")
	if contentType != "text/html" {
		t.Errorf("Expected Content-Type text/html, got %s", contentType)
	}
}

func TestAuthController_Callback_MissingProvider(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	mockManager := &MockAuthManager{}
	baseURL, _ := url.Parse("https://example.com")
	controller := NewAuthController(mockManager, logger, baseURL)
	
	req := httptest.NewRequest("GET", "/auth/callback", nil)
	w := httptest.NewRecorder()
	
	controller.Callback(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAuthController_Callback_InvalidProvider(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	mockManager := &MockAuthManager{
		providerError: errors.New("provider not found"),
	}
	baseURL, _ := url.Parse("https://example.com")
	controller := NewAuthController(mockManager, logger, baseURL)
	
	req := httptest.NewRequest("GET", "/auth/callback?provider=invalid", nil)
	w := httptest.NewRecorder()
	
	controller.Callback(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAuthController_Callback_OAuthProvider(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	testUser := &auth.User{
		ID:    "test-user",
		Email: "test@example.com",
		Name:  "Test User",
	}
	
	mockProvider := &MockProvider{
		id:       "google",
		name:     "Google",
		providerType: auth.ProviderTypeOAuth2,
		enabled:  true,
		user:     testUser,
	}
	
	mockManager := &MockAuthManager{
		provider: mockProvider,
		authenticateUser: func(ctx context.Context, user *auth.User) (*auth.Session, error) {
			return &auth.Session{
				ID:        "test-session",
				UserID:    user.ID,
				CreatedAt: time.Now(),
				ExpiresAt: time.Now().Add(24 * time.Hour),
			}, nil
		},
	}
	
	baseURL, _ := url.Parse("https://example.com")
	controller := NewAuthController(mockManager, logger, baseURL)
	
	req := httptest.NewRequest("GET", "/auth/callback?provider=google&code=test-code&state=test-state", nil)
	w := httptest.NewRecorder()
	
	controller.Callback(w, req)
	
	// The callback should process the OAuth response
	// Since we're using a mock provider, it should handle the callback
	if w.Code != http.StatusOK && w.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected status 200 or 302, got %d", w.Code)
	}
}

func TestAuthController_Callback_UnsupportedProvider(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	mockProvider := &MockProvider{
		id:       "saml",
		name:     "SAML",
		providerType: auth.ProviderTypeSAML,
		enabled:  true,
	}
	
	mockManager := &MockAuthManager{
		provider: mockProvider,
	}
	
	baseURL, _ := url.Parse("https://example.com")
	controller := NewAuthController(mockManager, logger, baseURL)
	
	req := httptest.NewRequest("GET", "/auth/callback?provider=saml", nil)
	w := httptest.NewRecorder()
	
	controller.Callback(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAuthController_Logout(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	testUser := &auth.User{
		ID:        "test-user",
		Email:     "test@example.com",
		SessionID: "test-session",
	}
	
	mockManager := &MockAuthManager{
		user: testUser,
	}
	
	baseURL, _ := url.Parse("https://example.com")
	controller := NewAuthController(mockManager, logger, baseURL)
	
	req := httptest.NewRequest("GET", "/auth/logout", nil)
	w := httptest.NewRecorder()
	
	controller.Logout(w, req)
	
	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected status %d, got %d", http.StatusTemporaryRedirect, w.Code)
	}
	
	location := w.Header().Get("Location")
	if location != "/auth/login" {
		t.Errorf("Expected redirect to /auth/login, got %s", location)
	}
}

func TestAuthController_Logout_NoUser(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	mockManager := &MockAuthManager{
		userError: http.ErrNoCookie,
	}
	
	baseURL, _ := url.Parse("https://example.com")
	controller := NewAuthController(mockManager, logger, baseURL)
	
	req := httptest.NewRequest("GET", "/auth/logout", nil)
	w := httptest.NewRecorder()
	
	controller.Logout(w, req)
	
	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected status %d, got %d", http.StatusTemporaryRedirect, w.Code)
	}
	
	location := w.Header().Get("Location")
	if location != "/auth/login" {
		t.Errorf("Expected redirect to /auth/login, got %s", location)
	}
}

func TestAuthController_RedirectToProvider(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	mockProvider := &MockProvider{
		id:       "google",
		name:     "Google",
		providerType: auth.ProviderTypeOAuth2,
		enabled:  true,
		authURL:  "https://accounts.google.com/oauth/authorize?client_id=test&redirect_uri=test&scope=openid+email+profile&state=test-state&response_type=code",
	}
	
	mockManager := &MockAuthManager{}
	baseURL, _ := url.Parse("https://example.com")
	controller := NewAuthController(mockManager, logger, baseURL)
	
	req := httptest.NewRequest("GET", "/auth/login", nil)
	w := httptest.NewRecorder()
	
	controller.redirectToProvider(w, req, mockProvider)
	
	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected status %d, got %d", http.StatusTemporaryRedirect, w.Code)
	}
	
	location := w.Header().Get("Location")
	if location == "" {
		t.Error("Location header should be set for redirect")
	}
}

func TestAuthController_RedirectToProvider_Error(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	mockProvider := &MockProvider{
		id:       "google",
		name:     "Google",
		providerType: auth.ProviderTypeOAuth2,
		enabled:  true,
		authError: errors.New("provider not configured"),
	}
	
	mockManager := &MockAuthManager{}
	baseURL, _ := url.Parse("https://example.com")
	controller := NewAuthController(mockManager, logger, baseURL)
	
	req := httptest.NewRequest("GET", "/auth/login", nil)
	w := httptest.NewRecorder()
	
	controller.redirectToProvider(w, req, mockProvider)
	
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestAuthController_ShowLoginPage(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	mockProvider1 := &MockProvider{
		id:       "google",
		name:     "Google",
		providerType: auth.ProviderTypeOAuth2,
		enabled:  true,
		authURL:  "https://accounts.google.com/oauth/authorize",
	}
	
	mockProvider2 := &MockProvider{
		id:       "okta",
		name:     "Okta",
		providerType: auth.ProviderTypeOIDC,
		enabled:  true,
		authURL:  "https://example.okta.com/oauth/authorize",
	}
	
	mockManager := &MockAuthManager{
		enabledProviders: []auth.Provider{mockProvider1, mockProvider2},
	}
	
	baseURL, _ := url.Parse("https://example.com")
	controller := NewAuthController(mockManager, logger, baseURL)
	
	req := httptest.NewRequest("GET", "/auth/login", nil)
	w := httptest.NewRecorder()
	
	controller.showLoginPage(w, req, []auth.Provider{mockProvider1, mockProvider2})
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	contentType := w.Header().Get("Content-Type")
	if contentType != "text/html" {
		t.Errorf("Expected Content-Type text/html, got %s", contentType)
	}
}

func TestAuthController_ShowLoginPage_ProviderError(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	mockProvider := &MockProvider{
		id:       "google",
		name:     "Google",
		providerType: auth.ProviderTypeOAuth2,
		enabled:  true,
		authError: errors.New("provider not configured"),
	}
	
	mockManager := &MockAuthManager{
		enabledProviders: []auth.Provider{mockProvider},
	}
	
	baseURL, _ := url.Parse("https://example.com")
	controller := NewAuthController(mockManager, logger, baseURL)
	
	req := httptest.NewRequest("GET", "/auth/login", nil)
	w := httptest.NewRecorder()
	
	controller.showLoginPage(w, req, []auth.Provider{mockProvider})
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	// Should still render the page even if one provider has an error
	contentType := w.Header().Get("Content-Type")
	if contentType != "text/html" {
		t.Errorf("Expected Content-Type text/html, got %s", contentType)
	}
}

func TestAuthController_HandleOAuthCallback_MissingCode(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	mockProvider := &MockProvider{
		id:       "google",
		name:     "Google",
		providerType: auth.ProviderTypeOAuth2,
		enabled:  true,
	}
	
	mockManager := &MockAuthManager{
		provider: mockProvider,
	}
	
	baseURL, _ := url.Parse("https://example.com")
	controller := NewAuthController(mockManager, logger, baseURL)
	
	req := httptest.NewRequest("GET", "/auth/callback?provider=google&state=test-state", nil)
	w := httptest.NewRecorder()
	
	controller.handleOAuthCallback(w, req, mockProvider)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAuthController_HandleOAuthCallback_MissingState(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	mockProvider := &MockProvider{
		id:       "google",
		name:     "Google",
		providerType: auth.ProviderTypeOAuth2,
		enabled:  true,
	}
	
	mockManager := &MockAuthManager{
		provider: mockProvider,
	}
	
	baseURL, _ := url.Parse("https://example.com")
	controller := NewAuthController(mockManager, logger, baseURL)
	
	req := httptest.NewRequest("GET", "/auth/callback?provider=google&code=test-code", nil)
	w := httptest.NewRecorder()
	
	controller.handleOAuthCallback(w, req, mockProvider)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAuthController_HandleOAuthCallback_ValidRequest(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	testUser := &auth.User{
		ID:    "test-user",
		Email: "test@example.com",
		Name:  "Test User",
	}
	
	mockProvider := &MockProvider{
		id:       "google",
		name:     "Google",
		providerType: auth.ProviderTypeOAuth2,
		enabled:  true,
		user:     testUser,
	}
	
	mockManager := &MockAuthManager{
		provider: mockProvider,
	}
	
	baseURL, _ := url.Parse("https://example.com")
	controller := NewAuthController(mockManager, logger, baseURL)
	
	req := httptest.NewRequest("GET", "/auth/callback?provider=google&code=test-code&state=test-state", nil)
	w := httptest.NewRecorder()
	
	controller.handleOAuthCallback(w, req, mockProvider)
	
	// Should redirect to home page after successful authentication
	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected status %d, got %d", http.StatusTemporaryRedirect, w.Code)
	}
	
	location := w.Header().Get("Location")
	if location != "/" {
		t.Errorf("Expected redirect to /, got %s", location)
	}
}

func TestAuthController_HandleOAuthCallback_ProviderError(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	mockProvider := &MockProvider{
		id:       "google",
		name:     "Google",
		providerType: auth.ProviderTypeOAuth2,
		enabled:  true,
		userError: errors.New("invalid credentials"),
	}
	
	mockManager := &MockAuthManager{
		provider: mockProvider,
	}
	
	baseURL, _ := url.Parse("https://example.com")
	controller := NewAuthController(mockManager, logger, baseURL)
	
	req := httptest.NewRequest("GET", "/auth/callback?provider=google&code=test-code&state=test-state", nil)
	w := httptest.NewRecorder()
	
	controller.handleOAuthCallback(w, req, mockProvider)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAuthController_GenerateState(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	mockManager := &MockAuthManager{}
	baseURL, _ := url.Parse("https://example.com")
	controller := NewAuthController(mockManager, logger, baseURL)
	
	state1, err1 := controller.generateState()
	if err1 != nil {
		t.Errorf("Failed to generate state: %v", err1)
	}
	
	state2, err2 := controller.generateState()
	if err2 != nil {
		t.Errorf("Failed to generate state: %v", err2)
	}
	
	if state1 == state2 {
		t.Error("Generated states should be different")
	}
	
	if len(state1) == 0 {
		t.Error("Generated state should not be empty")
	}
	
	if len(state2) == 0 {
		t.Error("Generated state should not be empty")
	}
}

func TestAuthController_Respond(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	mockManager := &MockAuthManager{}
	baseURL, _ := url.Parse("https://example.com")
	controller := NewAuthController(mockManager, logger, baseURL)
	
	w := httptest.NewRecorder()
	
	controller.respond(w, logging.Error, http.StatusBadRequest, "Test error message")
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	
	body := w.Body.String()
	if body != "Test error message" {
		t.Errorf("Expected body 'Test error message', got '%s'", body)
	}
}

func TestAuthController_Respond_WithFormatting(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	mockManager := &MockAuthManager{}
	baseURL, _ := url.Parse("https://example.com")
	controller := NewAuthController(mockManager, logger, baseURL)
	
	w := httptest.NewRecorder()
	
	controller.respond(w, logging.Error, http.StatusBadRequest, "Error: %s", "test error")
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	
	body := w.Body.String()
	if body != "Error: test error" {
		t.Errorf("Expected body 'Error: test error', got '%s'", body)
	}
} 