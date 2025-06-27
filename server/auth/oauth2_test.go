package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewOAuth2Provider(t *testing.T) {
	config := &ProviderConfig{
		Type:         ProviderTypeOAuth2,
		ID:           "google",
		Name:         "Google",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "https://example.com/callback",
		AuthURL:      "https://accounts.google.com/oauth/authorize",
		TokenURL:     "https://oauth2.googleapis.com/token",
		UserInfoURL:  "https://www.googleapis.com/oauth2/v2/userinfo",
		Scopes:       []string{"openid", "email", "profile"},
		Enabled:      true,
		DefaultRoles: []string{"user"},
	}

	provider, err := NewOAuth2Provider(config)
	if err != nil {
		t.Fatalf("Failed to create OAuth2 provider: %v", err)
	}

	if provider == nil {
		t.Fatal("Provider should not be nil")
	}

	if provider.GetType() != ProviderTypeOAuth2 {
		t.Errorf("Provider type = %s, want %s", provider.GetType(), ProviderTypeOAuth2)
	}

	if provider.GetID() != "google" {
		t.Errorf("Provider ID = %s, want google", provider.GetID())
	}

	if provider.GetName() != "Google" {
		t.Errorf("Provider name = %s, want Google", provider.GetName())
	}

	if !provider.IsEnabled() {
		t.Error("Provider should be enabled")
	}
}

func TestOAuth2Provider_InitAuthURL(t *testing.T) {
	config := &ProviderConfig{
		Type:         ProviderTypeOAuth2,
		ID:           "google",
		Name:         "Google",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "https://example.com/callback",
		AuthURL:      "https://accounts.google.com/oauth/authorize",
		TokenURL:     "https://oauth2.googleapis.com/token",
		UserInfoURL:  "https://www.googleapis.com/oauth2/v2/userinfo",
		Scopes:       []string{"openid", "email", "profile"},
		Enabled:      true,
		DefaultRoles: []string{"user"},
	}

	provider, err := NewOAuth2Provider(config)
	if err != nil {
		t.Fatalf("Failed to create OAuth2 provider: %v", err)
	}

	state := "test-state"
	authURL, err := provider.InitAuthURL(state)
	if err != nil {
		t.Errorf("Failed to generate auth URL: %v", err)
	}

	if authURL == "" {
		t.Error("Auth URL should not be empty")
	}

	// Check that the auth URL contains expected parameters
	if !contains(authURL, "client_id=test-client-id") {
		t.Error("Auth URL should contain client_id parameter")
	}

	if !contains(authURL, "redirect_uri=https%3A//example.com/callback") {
		t.Error("Auth URL should contain redirect_uri parameter")
	}

	if !contains(authURL, "scope=openid+email+profile") {
		t.Error("Auth URL should contain scope parameter")
	}

	if !contains(authURL, "state=test-state") {
		t.Error("Auth URL should contain state parameter")
	}

	if !contains(authURL, "response_type=code") {
		t.Error("Auth URL should contain response_type parameter")
	}
}

func TestOAuth2Provider_InitiateLogin(t *testing.T) {
	config := &ProviderConfig{
		Type:         ProviderTypeOAuth2,
		ID:           "google",
		Name:         "Google",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "https://example.com/callback",
		AuthURL:      "https://accounts.google.com/oauth/authorize",
		TokenURL:     "https://oauth2.googleapis.com/token",
		UserInfoURL:  "https://www.googleapis.com/oauth2/v2/userinfo",
		Scopes:       []string{"openid", "email", "profile"},
		Enabled:      true,
		DefaultRoles: []string{"user"},
	}

	provider, err := NewOAuth2Provider(config)
	if err != nil {
		t.Fatalf("Failed to create OAuth2 provider: %v", err)
	}

	req := httptest.NewRequest("GET", "/login", nil)
	w := httptest.NewRecorder()

	err = provider.InitiateLogin(w, req)
	if err != nil {
		t.Errorf("Failed to initiate login: %v", err)
	}

	if w.Code != http.StatusFound {
		t.Errorf("Expected status %d, got %d", http.StatusFound, w.Code)
	}

	location := w.Header().Get("Location")
	if location == "" {
		t.Error("Location header should not be empty")
	}

	if !contains(location, "accounts.google.com") {
		t.Error("Location should redirect to Google")
	}
}

func TestOAuth2Provider_ExchangeCode(t *testing.T) {
	// Create a mock server to simulate the OAuth2 token endpoint
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check that the request contains the expected parameters
		if err := r.ParseForm(); err != nil {
			t.Errorf("Failed to parse form: %v", err)
		}

		if r.FormValue("grant_type") != "authorization_code" {
			t.Error("Expected grant_type=authorization_code")
		}

		if r.FormValue("code") != "test-code" {
			t.Error("Expected code=test-code")
		}

		if r.FormValue("client_id") != "test-client-id" {
			t.Error("Expected client_id=test-client-id")
		}

		if r.FormValue("client_secret") != "test-client-secret" {
			t.Error("Expected client_secret=test-client-secret")
		}

		if r.FormValue("redirect_uri") != "https://example.com/callback" {
			t.Error("Expected redirect_uri=https://example.com/callback")
		}

		// Return a mock token response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{
			"access_token": "test-access-token",
			"token_type": "Bearer",
			"refresh_token": "test-refresh-token",
			"expires_in": 3600
		}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer mockServer.Close()

	config := &ProviderConfig{
		Type:         ProviderTypeOAuth2,
		ID:           "google",
		Name:         "Google",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "https://example.com/callback",
		AuthURL:      "https://accounts.google.com/oauth/authorize",
		TokenURL:     mockServer.URL,
		UserInfoURL:  "https://www.googleapis.com/oauth2/v2/userinfo",
		Scopes:       []string{"openid", "email", "profile"},
		Enabled:      true,
		DefaultRoles: []string{"user"},
	}

	provider, err := NewOAuth2Provider(config)
	if err != nil {
		t.Fatalf("Failed to create OAuth2 provider: %v", err)
	}

	ctx := context.Background()
	token, err := provider.ExchangeCode(ctx, "test-code")
	if err != nil {
		t.Errorf("Failed to exchange code: %v", err)
	}

	if token == nil {
		t.Fatal("Token should not be nil")
	}

	if token.AccessToken != "test-access-token" {
		t.Errorf("Access token = %s, want test-access-token", token.AccessToken)
	}

	if token.TokenType != "Bearer" {
		t.Errorf("Token type = %s, want Bearer", token.TokenType)
	}

	if token.RefreshToken != "test-refresh-token" {
		t.Errorf("Refresh token = %s, want test-refresh-token", token.RefreshToken)
	}

	if token.ExpiresIn != 3600 {
		t.Errorf("Expires in = %d, want 3600", token.ExpiresIn)
	}
}

func TestOAuth2Provider_GetUserInfo(t *testing.T) {
	// Create a mock server to simulate the user info endpoint
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check that the request contains the authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-access-token" {
			t.Errorf("Expected Authorization: Bearer test-access-token, got %s", authHeader)
		}

		// Return a mock user info response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{
			"id": "123456789",
			"email": "test@example.com",
			"name": "Test User",
			"given_name": "Test",
			"family_name": "User",
			"picture": "https://example.com/avatar.jpg",
			"locale": "en"
		}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer mockServer.Close()

	config := &ProviderConfig{
		Type:         ProviderTypeOAuth2,
		ID:           "google",
		Name:         "Google",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "https://example.com/callback",
		AuthURL:      "https://accounts.google.com/oauth/authorize",
		TokenURL:     "https://oauth2.googleapis.com/token",
		UserInfoURL:  mockServer.URL,
		Scopes:       []string{"openid", "email", "profile"},
		Enabled:      true,
		DefaultRoles: []string{"user"},
	}

	provider, err := NewOAuth2Provider(config)
	if err != nil {
		t.Fatalf("Failed to create OAuth2 provider: %v", err)
	}

	token := &TokenResponse{
		AccessToken: "test-access-token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
	}

	ctx := context.Background()
	user, err := provider.GetUserInfo(ctx, token)
	if err != nil {
		t.Errorf("Failed to get user info: %v", err)
	}

	if user == nil {
		t.Fatal("User should not be nil")
	}

	if user.ID != "123456789" {
		t.Errorf("User ID = %s, want 123456789", user.ID)
	}

	if user.Email != "test@example.com" {
		t.Errorf("User email = %s, want test@example.com", user.Email)
	}

	if user.Name != "Test User" {
		t.Errorf("User name = %s, want Test User", user.Name)
	}

	if user.Provider != "google" {
		t.Errorf("User provider = %s, want google", user.Provider)
	}

	if user.LastLogin.IsZero() {
		t.Error("User last login should not be zero")
	}
}

func TestOAuth2Provider_ValidateToken(t *testing.T) {
	// Create a mock server to simulate the user info endpoint
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check that the request contains the authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-token" {
			t.Errorf("Expected Authorization: Bearer test-token, got %s", authHeader)
		}

		// Return a mock user info response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{
			"id": "123456789",
			"email": "test@example.com",
			"name": "Test User",
			"given_name": "Test",
			"family_name": "User",
			"picture": "https://example.com/avatar.jpg",
			"locale": "en"
		}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer mockServer.Close()

	config := &ProviderConfig{
		Type:         ProviderTypeOAuth2,
		ID:           "google",
		Name:         "Google",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "https://example.com/callback",
		AuthURL:      "https://accounts.google.com/oauth/authorize",
		TokenURL:     "https://oauth2.googleapis.com/token",
		UserInfoURL:  mockServer.URL,
		Scopes:       []string{"openid", "email", "profile"},
		Enabled:      true,
		DefaultRoles: []string{"user"},
	}

	provider, err := NewOAuth2Provider(config)
	if err != nil {
		t.Fatalf("Failed to create OAuth2 provider: %v", err)
	}

	ctx := context.Background()
	user, err := provider.ValidateToken(ctx, "test-token")
	if err != nil {
		t.Errorf("Failed to validate token: %v", err)
	}

	if user == nil {
		t.Fatal("User should not be nil")
	}

	if user.ID != "123456789" {
		t.Errorf("User ID = %s, want 123456789", user.ID)
	}

	if user.Email != "test@example.com" {
		t.Errorf("User email = %s, want test@example.com", user.Email)
	}

	if user.Name != "Test User" {
		t.Errorf("User name = %s, want Test User", user.Name)
	}

	if user.Provider != "google" {
		t.Errorf("User provider = %s, want google", user.Provider)
	}
}

func TestOAuth2Provider_ProcessSAMLResponse(t *testing.T) {
	config := &ProviderConfig{
		Type:         ProviderTypeOAuth2,
		ID:           "google",
		Name:         "Google",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "https://example.com/callback",
		AuthURL:      "https://accounts.google.com/oauth/authorize",
		TokenURL:     "https://oauth2.googleapis.com/token",
		UserInfoURL:  "https://www.googleapis.com/oauth2/v2/userinfo",
		Scopes:       []string{"openid", "email", "profile"},
		Enabled:      true,
		DefaultRoles: []string{"user"},
	}

	provider, err := NewOAuth2Provider(config)
	if err != nil {
		t.Fatalf("Failed to create OAuth2 provider: %v", err)
	}

	req := httptest.NewRequest("POST", "/saml", nil)
	w := httptest.NewRecorder()

	user, err := provider.ProcessSAMLResponse(w, req)
	if err == nil {
		t.Error("OAuth2 provider should not support SAML responses")
	}
	if user != nil {
		t.Error("Should return nil user for unsupported SAML response")
	}
}

func TestOAuth2Provider_ErrorHandling(t *testing.T) {
	// Test with invalid configuration
	config := &ProviderConfig{
		Type:         ProviderTypeOAuth2,
		ID:           "",
		Name:         "",
		ClientID:     "",
		ClientSecret: "",
		Enabled:      true,
	}

	provider, err := NewOAuth2Provider(config)
	if err == nil {
		t.Error("Should return error for invalid configuration")
	}
	if provider != nil {
		t.Error("Should return nil provider for invalid configuration")
	}

	// Test with missing required URLs
	config = &ProviderConfig{
		Type:         ProviderTypeOAuth2,
		ID:           "google",
		Name:         "Google",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		Enabled:      true,
	}

	provider, err = NewOAuth2Provider(config)
	if err == nil {
		t.Error("Should return error for missing required URLs")
	}
	if provider != nil {
		t.Error("Should return nil provider for missing required URLs")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
} 