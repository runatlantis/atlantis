package auth

import (
	"context"
	"encoding/base64"
	"net/http/httptest"
	"testing"
)

func TestNewBasicAuthProvider(t *testing.T) {
	config := &ProviderConfig{
		Type:         ProviderTypeBasicAuth,
		ID:           "basic",
		Name:         "Basic Authentication",
		ClientID:     "atlantis",
		ClientSecret: "atlantis",
		Enabled:      true,
		DefaultRoles: []string{"user"},
	}

	provider, err := NewBasicAuthProvider(config)
	if err != nil {
		t.Fatalf("Failed to create basic auth provider: %v", err)
	}

	if provider == nil {
		t.Fatal("Provider should not be nil")
	}

	if provider.GetType() != ProviderTypeBasicAuth {
		t.Errorf("Provider type = %s, want %s", provider.GetType(), ProviderTypeBasicAuth)
	}

	if provider.GetID() != "basic" {
		t.Errorf("Provider ID = %s, want basic", provider.GetID())
	}

	if provider.GetName() != "Basic Authentication" {
		t.Errorf("Provider name = %s, want Basic Authentication", provider.GetName())
	}

	if !provider.IsEnabled() {
		t.Error("Provider should be enabled")
	}
}

func TestBasicAuthProvider_InitAuthURL(t *testing.T) {
	config := &ProviderConfig{
		Type:         ProviderTypeBasicAuth,
		ID:           "basic",
		Name:         "Basic Authentication",
		ClientID:     "atlantis",
		ClientSecret: "atlantis",
		Enabled:      true,
		DefaultRoles: []string{"user"},
	}

	provider, err := NewBasicAuthProvider(config)
	if err != nil {
		t.Fatalf("Failed to create basic auth provider: %v", err)
	}

	state := "test-state"
	authURL, err := provider.InitAuthURL(state)
	if err == nil {
		t.Error("Basic auth provider should not support auth URL generation")
	}
	if authURL != "" {
		t.Error("Auth URL should be empty for basic auth")
	}
}

func TestBasicAuthProvider_InitiateLogin(t *testing.T) {
	config := &ProviderConfig{
		Type:         ProviderTypeBasicAuth,
		ID:           "basic",
		Name:         "Basic Authentication",
		ClientID:     "atlantis",
		ClientSecret: "atlantis",
		Enabled:      true,
		DefaultRoles: []string{"user"},
	}

	provider, err := NewBasicAuthProvider(config)
	if err != nil {
		t.Fatalf("Failed to create basic auth provider: %v", err)
	}

	req := httptest.NewRequest("GET", "/login", nil)
	w := httptest.NewRecorder()

	err = provider.InitiateLogin(w, req)
	if err == nil {
		t.Error("Basic auth provider should not support initiate login")
	}
}

func TestBasicAuthProvider_ExchangeCode(t *testing.T) {
	config := &ProviderConfig{
		Type:         ProviderTypeBasicAuth,
		ID:           "basic",
		Name:         "Basic Authentication",
		ClientID:     "atlantis",
		ClientSecret: "atlantis",
		Enabled:      true,
		DefaultRoles: []string{"user"},
	}

	provider, err := NewBasicAuthProvider(config)
	if err != nil {
		t.Fatalf("Failed to create basic auth provider: %v", err)
	}

	ctx := context.Background()
	token, err := provider.ExchangeCode(ctx, "test-code")
	if err == nil {
		t.Error("Basic auth provider should not support code exchange")
	}
	if token != nil {
		t.Error("Should return nil token for unsupported code exchange")
	}
}

func TestBasicAuthProvider_GetUserInfo(t *testing.T) {
	config := &ProviderConfig{
		Type:         ProviderTypeBasicAuth,
		ID:           "basic",
		Name:         "Basic Authentication",
		ClientID:     "atlantis",
		ClientSecret: "atlantis",
		Enabled:      true,
		DefaultRoles: []string{"user"},
	}

	provider, err := NewBasicAuthProvider(config)
	if err != nil {
		t.Fatalf("Failed to create basic auth provider: %v", err)
	}

	token := &TokenResponse{
		AccessToken: "test-access-token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
	}

	ctx := context.Background()
	user, err := provider.GetUserInfo(ctx, token)
	if err == nil {
		t.Error("Basic auth provider should not support token-based user info")
	}
	if user != nil {
		t.Error("Should return nil user for unsupported token-based user info")
	}
}

func TestBasicAuthProvider_ValidateToken(t *testing.T) {
	config := &ProviderConfig{
		Type:         ProviderTypeBasicAuth,
		ID:           "basic",
		Name:         "Basic Authentication",
		ClientID:     "atlantis",
		ClientSecret: "atlantis",
		Enabled:      true,
		DefaultRoles: []string{"user"},
	}

	provider, err := NewBasicAuthProvider(config)
	if err != nil {
		t.Fatalf("Failed to create basic auth provider: %v", err)
	}

	ctx := context.Background()
	credentials := base64.StdEncoding.EncodeToString([]byte("atlantis:atlantis"))
	user, err := provider.ValidateToken(ctx, credentials)
	if err != nil {
		t.Errorf("Failed to validate token: %v", err)
	}

	if user == nil {
		t.Fatal("User should not be nil")
	}

	if user.ID != "atlantis" {
		t.Errorf("User ID = %s, want atlantis", user.ID)
	}

	if user.Email != "atlantis" {
		t.Errorf("User email = %s, want atlantis", user.Email)
	}

	if user.Name != "atlantis" {
		t.Errorf("User name = %s, want atlantis", user.Name)
	}

	if user.Provider != "basic" {
		t.Errorf("User provider = %s, want basic", user.Provider)
	}

	if user.LastLogin.IsZero() {
		t.Error("User last login should not be zero")
	}

	// Check that default roles are assigned
	if len(user.Roles) != 1 || user.Roles[0] != "user" {
		t.Errorf("User roles = %v, want [user]", user.Roles)
	}
}

func TestBasicAuthProvider_ProcessSAMLResponse(t *testing.T) {
	config := &ProviderConfig{
		Type:         ProviderTypeBasicAuth,
		ID:           "basic",
		Name:         "Basic Authentication",
		ClientID:     "atlantis",
		ClientSecret: "atlantis",
		Enabled:      true,
		DefaultRoles: []string{"user"},
	}

	provider, err := NewBasicAuthProvider(config)
	if err != nil {
		t.Fatalf("Failed to create basic auth provider: %v", err)
	}

	req := httptest.NewRequest("POST", "/saml", nil)
	w := httptest.NewRecorder()

	user, err := provider.ProcessSAMLResponse(w, req)
	if err == nil {
		t.Error("Basic auth provider should not support SAML responses")
	}
	if user != nil {
		t.Error("Should return nil user for unsupported SAML response")
	}
}

func TestBasicAuthProvider_ValidateBasicAuth(t *testing.T) {
	config := &ProviderConfig{
		Type:         ProviderTypeBasicAuth,
		ID:           "basic",
		Name:         "Basic Authentication",
		ClientID:     "atlantis",
		ClientSecret: "atlantis",
		Enabled:      true,
		DefaultRoles: []string{"user"},
	}

	provider, err := NewBasicAuthProvider(config)
	if err != nil {
		t.Fatalf("Failed to create basic auth provider: %v", err)
	}

	// Test with valid credentials
	req := httptest.NewRequest("GET", "/", nil)
	req.SetBasicAuth("atlantis", "atlantis")

	user, err := provider.ValidateBasicAuth(req)
	if err != nil {
		t.Errorf("Failed to validate basic auth: %v", err)
	}

	if user == nil {
		t.Fatal("User should not be nil")
	}

	if user.ID != "atlantis" {
		t.Errorf("User ID = %s, want atlantis", user.ID)
	}

	if user.Email != "atlantis" {
		t.Errorf("User email = %s, want atlantis", user.Email)
	}

	if user.Name != "atlantis" {
		t.Errorf("User name = %s, want atlantis", user.Name)
	}

	if user.Provider != "basic" {
		t.Errorf("User provider = %s, want basic", user.Provider)
	}

	if user.LastLogin.IsZero() {
		t.Error("User last login should not be zero")
	}

	// Check that default roles are assigned
	if len(user.Roles) != 1 || user.Roles[0] != "user" {
		t.Errorf("User roles = %v, want [user]", user.Roles)
	}
}

func TestBasicAuthProvider_ValidateBasicAuth_InvalidCredentials(t *testing.T) {
	config := &ProviderConfig{
		Type:         ProviderTypeBasicAuth,
		ID:           "basic",
		Name:         "Basic Authentication",
		ClientID:     "atlantis",
		ClientSecret: "atlantis",
		Enabled:      true,
		DefaultRoles: []string{"user"},
	}

	provider, err := NewBasicAuthProvider(config)
	if err != nil {
		t.Fatalf("Failed to create basic auth provider: %v", err)
	}

	// Test with invalid credentials
	req := httptest.NewRequest("GET", "/", nil)
	req.SetBasicAuth("atlantis", "wrongpassword")

	user, err := provider.ValidateBasicAuth(req)
	if err == nil {
		t.Error("Should return error for invalid credentials")
	}
	if user != nil {
		t.Error("Should return nil user for invalid credentials")
	}
}

func TestBasicAuthProvider_ValidateBasicAuth_NoAuthHeader(t *testing.T) {
	config := &ProviderConfig{
		Type:         ProviderTypeBasicAuth,
		ID:           "basic",
		Name:         "Basic Authentication",
		ClientID:     "atlantis",
		ClientSecret: "atlantis",
		Enabled:      true,
		DefaultRoles: []string{"user"},
	}

	provider, err := NewBasicAuthProvider(config)
	if err != nil {
		t.Fatalf("Failed to create basic auth provider: %v", err)
	}

	// Test with no authorization header
	req := httptest.NewRequest("GET", "/", nil)

	user, err := provider.ValidateBasicAuth(req)
	if err == nil {
		t.Error("Should return error for missing authorization header")
	}
	if user != nil {
		t.Error("Should return nil user for missing authorization header")
	}
}

func TestBasicAuthProvider_ValidateToken_InvalidCredentials(t *testing.T) {
	config := &ProviderConfig{
		Type:         ProviderTypeBasicAuth,
		ID:           "basic",
		Name:         "Basic Authentication",
		ClientID:     "atlantis",
		ClientSecret: "atlantis",
		Enabled:      true,
		DefaultRoles: []string{"user"},
	}

	provider, err := NewBasicAuthProvider(config)
	if err != nil {
		t.Fatalf("Failed to create basic auth provider: %v", err)
	}

	// Test with invalid credentials
	ctx := context.Background()
	credentials := base64.StdEncoding.EncodeToString([]byte("atlantis:wrongpassword"))
	user, err := provider.ValidateToken(ctx, credentials)
	if err == nil {
		t.Error("Should return error for invalid credentials")
	}
	if user != nil {
		t.Error("Should return nil user for invalid credentials")
	}
}

func TestBasicAuthProvider_ValidateToken_InvalidFormat(t *testing.T) {
	config := &ProviderConfig{
		Type:         ProviderTypeBasicAuth,
		ID:           "basic",
		Name:         "Basic Authentication",
		ClientID:     "atlantis",
		ClientSecret: "atlantis",
		Enabled:      true,
		DefaultRoles: []string{"user"},
	}

	provider, err := NewBasicAuthProvider(config)
	if err != nil {
		t.Fatalf("Failed to create basic auth provider: %v", err)
	}

	// Test with invalid credentials format (no colon)
	ctx := context.Background()
	credentials := base64.StdEncoding.EncodeToString([]byte("atlantis"))
	user, err := provider.ValidateToken(ctx, credentials)
	if err == nil {
		t.Error("Should return error for invalid credentials format")
	}
	if user != nil {
		t.Error("Should return nil user for invalid credentials format")
	}
}

func TestBasicAuthProvider_ErrorHandling(t *testing.T) {
	// Test with invalid configuration
	config := &ProviderConfig{
		Type:         ProviderTypeBasicAuth,
		ID:           "",
		Name:         "",
		ClientID:     "",
		ClientSecret: "",
		Enabled:      true,
	}

	provider, err := NewBasicAuthProvider(config)
	if err == nil {
		t.Error("Should return error for invalid configuration")
	}
	if provider != nil {
		t.Error("Should return nil provider for invalid configuration")
	}

	// Test with missing credentials
	config = &ProviderConfig{
		Type:         ProviderTypeBasicAuth,
		ID:           "basic",
		Name:         "Basic Authentication",
		ClientID:     "",
		ClientSecret: "",
		Enabled:      true,
	}

	provider, err = NewBasicAuthProvider(config)
	if err == nil {
		t.Error("Should return error for missing credentials")
	}
	if provider != nil {
		t.Error("Should return nil provider for missing credentials")
	}
}

func TestBasicAuthProvider_CustomRoles(t *testing.T) {
	config := &ProviderConfig{
		Type:         ProviderTypeBasicAuth,
		ID:           "basic",
		Name:         "Basic Authentication",
		ClientID:     "atlantis",
		ClientSecret: "atlantis",
		Enabled:      true,
		DefaultRoles: []string{"admin", "user"},
	}

	provider, err := NewBasicAuthProvider(config)
	if err != nil {
		t.Fatalf("Failed to create basic auth provider: %v", err)
	}

	req := httptest.NewRequest("GET", "/", nil)
	req.SetBasicAuth("atlantis", "atlantis")

	user, err := provider.ValidateBasicAuth(req)
	if err != nil {
		t.Errorf("Failed to validate basic auth: %v", err)
	}

	if user == nil {
		t.Fatal("User should not be nil")
	}

	// Check that custom roles are assigned
	expectedRoles := []string{"admin", "user"}
	if len(user.Roles) != len(expectedRoles) {
		t.Errorf("User roles = %v, want %v", user.Roles, expectedRoles)
	}

	for _, expectedRole := range expectedRoles {
		found := false
		for _, role := range user.Roles {
			if role == expectedRole {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("User missing role: %s", expectedRole)
		}
	}
} 