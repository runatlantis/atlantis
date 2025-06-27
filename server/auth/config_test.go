package auth

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestLoadConfigFromFile(t *testing.T) {
	// Create a temporary config file
	configData := map[string]interface{}{
		"session_secret":      "test-secret",
		"session_duration":    "2h",
		"session_cookie_name": "test_session",
		"secure_cookies":      true,
		"csrf_secret":         "csrf-secret",
		"enable_basic_auth":   true,
		"basic_auth_user":     "testuser",
		"basic_auth_pass":     "testpass",
		"default_roles":       []string{"user", "developer"},
		"admin_groups":        []string{"admin"},
		"admin_emails":        []string{"admin@example.com"},
		"providers": []map[string]interface{}{
			{
				"type":          "oauth2",
				"id":            "google",
				"name":          "Google",
				"client_id":     "test-client-id",
				"client_secret": "test-client-secret",
				"redirect_url":  "https://example.com/callback",
				"enabled":       true,
				"default_roles": []string{"user"},
			},
		},
	}

	configJSON, err := json.Marshal(configData)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	// Write to temporary file
	tmpfile, err := os.CreateTemp("", "auth_config_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(configJSON); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	tmpfile.Close()

	// Load config from file
	config, err := LoadConfigFromFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify config values
	if config.SessionSecret != "test-secret" {
		t.Errorf("SessionSecret = %s, want test-secret", config.SessionSecret)
	}

	if config.SessionDuration != 2*time.Hour {
		t.Errorf("SessionDuration = %v, want 2h", config.SessionDuration)
	}

	if config.SessionCookieName != "test_session" {
		t.Errorf("SessionCookieName = %s, want test_session", config.SessionCookieName)
	}

	if !config.SecureCookies {
		t.Error("SecureCookies should be true")
	}

	if config.CSRFSecret != "csrf-secret" {
		t.Errorf("CSRFSecret = %s, want csrf-secret", config.CSRFSecret)
	}

	if !config.EnableBasicAuth {
		t.Error("EnableBasicAuth should be true")
	}

	if config.BasicAuthUser != "testuser" {
		t.Errorf("BasicAuthUser = %s, want testuser", config.BasicAuthUser)
	}

	if config.BasicAuthPass != "testpass" {
		t.Errorf("BasicAuthPass = %s, want testpass", config.BasicAuthPass)
	}

	if len(config.DefaultRoles) != 2 {
		t.Errorf("DefaultRoles = %v, want 2 roles", config.DefaultRoles)
	}

	if len(config.AdminGroups) != 1 {
		t.Errorf("AdminGroups = %v, want 1 group", config.AdminGroups)
	}

	if len(config.Providers) != 1 {
		t.Errorf("Providers = %v, want 1 provider", config.Providers)
	}

	provider := config.Providers[0]
	if provider.ID != "google" {
		t.Errorf("Provider ID = %s, want google", provider.ID)
	}

	if provider.Name != "Google" {
		t.Errorf("Provider Name = %s, want Google", provider.Name)
	}
}

func TestLoadConfigFromFile_Defaults(t *testing.T) {
	// Create a minimal config file
	configData := map[string]interface{}{
		"session_secret": "test-secret",
	}

	configJSON, err := json.Marshal(configData)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	// Write to temporary file
	tmpfile, err := os.CreateTemp("", "auth_config_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(configJSON); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	tmpfile.Close()

	// Load config from file
	config, err := LoadConfigFromFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify default values are set
	if config.SessionDuration != 24*time.Hour {
		t.Errorf("SessionDuration = %v, want 24h", config.SessionDuration)
	}

	if config.SessionCookieName != "atlantis_session" {
		t.Errorf("SessionCookieName = %s, want atlantis_session", config.SessionCookieName)
	}
}

func TestLoadConfigFromFile_InvalidJSON(t *testing.T) {
	// Create a temporary file with invalid JSON
	tmpfile, err := os.CreateTemp("", "auth_config_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte("invalid json")); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	tmpfile.Close()

	// Load config from file should fail
	_, err = LoadConfigFromFile(tmpfile.Name())
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestLoadConfigFromFile_FileNotFound(t *testing.T) {
	// Try to load non-existent file
	_, err := LoadConfigFromFile("/non/existent/file.json")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("ATLANTIS_SESSION_SECRET", "env-secret")
	os.Setenv("ATLANTIS_SECURE_COOKIES", "true")
	os.Setenv("ATLANTIS_CSRF_SECRET", "env-csrf")
	os.Setenv("ATLANTIS_ENABLE_BASIC_AUTH", "true")
	os.Setenv("ATLANTIS_BASIC_AUTH_USER", "envuser")
	os.Setenv("ATLANTIS_BASIC_AUTH_PASS", "envpass")
	defer func() {
		os.Unsetenv("ATLANTIS_SESSION_SECRET")
		os.Unsetenv("ATLANTIS_SECURE_COOKIES")
		os.Unsetenv("ATLANTIS_CSRF_SECRET")
		os.Unsetenv("ATLANTIS_ENABLE_BASIC_AUTH")
		os.Unsetenv("ATLANTIS_BASIC_AUTH_USER")
		os.Unsetenv("ATLANTIS_BASIC_AUTH_PASS")
	}()

	config, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("Failed to load config from env: %v", err)
	}

	if config.SessionSecret != "env-secret" {
		t.Errorf("SessionSecret = %s, want env-secret", config.SessionSecret)
	}

	if !config.SecureCookies {
		t.Error("SecureCookies should be true")
	}

	if config.CSRFSecret != "env-csrf" {
		t.Errorf("CSRFSecret = %s, want env-csrf", config.CSRFSecret)
	}

	if !config.EnableBasicAuth {
		t.Error("EnableBasicAuth should be true")
	}

	if config.BasicAuthUser != "envuser" {
		t.Errorf("BasicAuthUser = %s, want envuser", config.BasicAuthUser)
	}

	if config.BasicAuthPass != "envpass" {
		t.Errorf("BasicAuthPass = %s, want envpass", config.BasicAuthPass)
	}
}

func TestLoadConfigFromEnv_Defaults(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("ATLANTIS_SESSION_SECRET")
	os.Unsetenv("ATLANTIS_SECURE_COOKIES")
	os.Unsetenv("ATLANTIS_CSRF_SECRET")
	os.Unsetenv("ATLANTIS_ENABLE_BASIC_AUTH")
	os.Unsetenv("ATLANTIS_BASIC_AUTH_USER")
	os.Unsetenv("ATLANTIS_BASIC_AUTH_PASS")

	config, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("Failed to load config from env: %v", err)
	}

	if config.SessionSecret != "change-me-in-production" {
		t.Errorf("SessionSecret = %s, want change-me-in-production", config.SessionSecret)
	}

	if config.SecureCookies {
		t.Error("SecureCookies should be false by default")
	}

	if config.CSRFSecret != "change-me-in-production" {
		t.Errorf("CSRFSecret = %s, want change-me-in-production", config.CSRFSecret)
	}

	if config.EnableBasicAuth {
		t.Error("EnableBasicAuth should be false by default")
	}

	if config.BasicAuthUser != "atlantis" {
		t.Errorf("BasicAuthUser = %s, want atlantis", config.BasicAuthUser)
	}

	if config.BasicAuthPass != "atlantis" {
		t.Errorf("BasicAuthPass = %s, want atlantis", config.BasicAuthPass)
	}
}

func TestLoadConfigFromEnv_GoogleProvider(t *testing.T) {
	// Set Google OAuth2 environment variables
	os.Setenv("ATLANTIS_GOOGLE_CLIENT_ID", "google-client-id")
	os.Setenv("ATLANTIS_GOOGLE_CLIENT_SECRET", "google-client-secret")
	os.Setenv("ATLANTIS_GOOGLE_REDIRECT_URL", "https://example.com/google/callback")
	defer func() {
		os.Unsetenv("ATLANTIS_GOOGLE_CLIENT_ID")
		os.Unsetenv("ATLANTIS_GOOGLE_CLIENT_SECRET")
		os.Unsetenv("ATLANTIS_GOOGLE_REDIRECT_URL")
	}()

	config, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("Failed to load config from env: %v", err)
	}

	if len(config.Providers) != 1 {
		t.Fatalf("Expected 1 provider, got %d", len(config.Providers))
	}

	provider := config.Providers[0]
	if provider.ID != "google" {
		t.Errorf("Provider ID = %s, want google", provider.ID)
	}

	if provider.Name != "Google" {
		t.Errorf("Provider Name = %s, want Google", provider.Name)
	}

	if provider.ClientID != "google-client-id" {
		t.Errorf("Provider ClientID = %s, want google-client-id", provider.ClientID)
	}

	if provider.ClientSecret != "google-client-secret" {
		t.Errorf("Provider ClientSecret = %s, want google-client-secret", provider.ClientSecret)
	}

	if provider.RedirectURL != "https://example.com/google/callback" {
		t.Errorf("Provider RedirectURL = %s, want https://example.com/google/callback", provider.RedirectURL)
	}

	if provider.Type != ProviderTypeOAuth2 {
		t.Errorf("Provider Type = %s, want oauth2", provider.Type)
	}
}

func TestLoadConfigFromEnv_OktaProvider(t *testing.T) {
	// Set Okta OIDC environment variables
	os.Setenv("ATLANTIS_OKTA_CLIENT_ID", "okta-client-id")
	os.Setenv("ATLANTIS_OKTA_CLIENT_SECRET", "okta-client-secret")
	os.Setenv("ATLANTIS_OKTA_REDIRECT_URL", "https://example.com/okta/callback")
	os.Setenv("ATLANTIS_OKTA_ISSUER_URL", "https://example.okta.com")
	defer func() {
		os.Unsetenv("ATLANTIS_OKTA_CLIENT_ID")
		os.Unsetenv("ATLANTIS_OKTA_CLIENT_SECRET")
		os.Unsetenv("ATLANTIS_OKTA_REDIRECT_URL")
		os.Unsetenv("ATLANTIS_OKTA_ISSUER_URL")
	}()

	config, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("Failed to load config from env: %v", err)
	}

	if len(config.Providers) != 1 {
		t.Fatalf("Expected 1 provider, got %d", len(config.Providers))
	}

	provider := config.Providers[0]
	if provider.ID != "okta" {
		t.Errorf("Provider ID = %s, want okta", provider.ID)
	}

	if provider.Name != "Okta" {
		t.Errorf("Provider Name = %s, want Okta", provider.Name)
	}

	if provider.Type != ProviderTypeOIDC {
		t.Errorf("Provider Type = %s, want oidc", provider.Type)
	}

	if provider.IssuerURL != "https://example.okta.com" {
		t.Errorf("Provider IssuerURL = %s, want https://example.okta.com", provider.IssuerURL)
	}
}

func TestLoadConfigFromEnv_OktaProvider_MissingIssuerURL(t *testing.T) {
	// Set Okta OIDC environment variables without issuer URL
	os.Setenv("ATLANTIS_OKTA_CLIENT_ID", "okta-client-id")
	os.Setenv("ATLANTIS_OKTA_CLIENT_SECRET", "okta-client-secret")
	os.Setenv("ATLANTIS_OKTA_REDIRECT_URL", "https://example.com/okta/callback")
	defer func() {
		os.Unsetenv("ATLANTIS_OKTA_CLIENT_ID")
		os.Unsetenv("ATLANTIS_OKTA_CLIENT_SECRET")
		os.Unsetenv("ATLANTIS_OKTA_REDIRECT_URL")
	}()

	_, err := LoadConfigFromEnv()
	if err == nil {
		t.Error("Expected error for missing ATLANTIS_OKTA_ISSUER_URL")
	}
}

func TestLoadConfigFromEnv_AzureProvider(t *testing.T) {
	// Set Azure AD environment variables
	os.Setenv("ATLANTIS_AZURE_CLIENT_ID", "azure-client-id")
	os.Setenv("ATLANTIS_AZURE_CLIENT_SECRET", "azure-client-secret")
	os.Setenv("ATLANTIS_AZURE_REDIRECT_URL", "https://example.com/azure/callback")
	os.Setenv("ATLANTIS_AZURE_TENANT_ID", "tenant-123")
	defer func() {
		os.Unsetenv("ATLANTIS_AZURE_CLIENT_ID")
		os.Unsetenv("ATLANTIS_AZURE_CLIENT_SECRET")
		os.Unsetenv("ATLANTIS_AZURE_REDIRECT_URL")
		os.Unsetenv("ATLANTIS_AZURE_TENANT_ID")
	}()

	config, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("Failed to load config from env: %v", err)
	}

	if len(config.Providers) != 1 {
		t.Fatalf("Expected 1 provider, got %d", len(config.Providers))
	}

	provider := config.Providers[0]
	if provider.ID != "azure" {
		t.Errorf("Provider ID = %s, want azure", provider.ID)
	}

	if provider.Name != "Azure AD" {
		t.Errorf("Provider Name = %s, want Azure AD", provider.Name)
	}

	if provider.Type != ProviderTypeOIDC {
		t.Errorf("Provider Type = %s, want oidc", provider.Type)
	}

	expectedIssuerURL := "https://login.microsoftonline.com/tenant-123"
	if provider.IssuerURL != expectedIssuerURL {
		t.Errorf("Provider IssuerURL = %s, want %s", provider.IssuerURL, expectedIssuerURL)
	}
}

func TestLoadConfigFromEnv_AzureProvider_MissingTenantID(t *testing.T) {
	// Set Azure AD environment variables without tenant ID
	os.Setenv("ATLANTIS_AZURE_CLIENT_ID", "azure-client-id")
	os.Setenv("ATLANTIS_AZURE_CLIENT_SECRET", "azure-client-secret")
	os.Setenv("ATLANTIS_AZURE_REDIRECT_URL", "https://example.com/azure/callback")
	defer func() {
		os.Unsetenv("ATLANTIS_AZURE_CLIENT_ID")
		os.Unsetenv("ATLANTIS_AZURE_CLIENT_SECRET")
		os.Unsetenv("ATLANTIS_AZURE_REDIRECT_URL")
	}()

	_, err := LoadConfigFromEnv()
	if err == nil {
		t.Error("Expected error for missing ATLANTIS_AZURE_TENANT_ID")
	}
}

func TestLoadConfigFromEnv_Auth0Provider(t *testing.T) {
	// Set Auth0 environment variables
	os.Setenv("ATLANTIS_AUTH0_CLIENT_ID", "auth0-client-id")
	os.Setenv("ATLANTIS_AUTH0_CLIENT_SECRET", "auth0-client-secret")
	os.Setenv("ATLANTIS_AUTH0_REDIRECT_URL", "https://example.com/auth0/callback")
	os.Setenv("ATLANTIS_AUTH0_DOMAIN", "example.auth0.com")
	defer func() {
		os.Unsetenv("ATLANTIS_AUTH0_CLIENT_ID")
		os.Unsetenv("ATLANTIS_AUTH0_CLIENT_SECRET")
		os.Unsetenv("ATLANTIS_AUTH0_REDIRECT_URL")
		os.Unsetenv("ATLANTIS_AUTH0_DOMAIN")
	}()

	config, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("Failed to load config from env: %v", err)
	}

	if len(config.Providers) != 1 {
		t.Fatalf("Expected 1 provider, got %d", len(config.Providers))
	}

	provider := config.Providers[0]
	if provider.ID != "auth0" {
		t.Errorf("Provider ID = %s, want auth0", provider.ID)
	}

	if provider.Name != "Auth0" {
		t.Errorf("Provider Name = %s, want Auth0", provider.Name)
	}

	if provider.Type != ProviderTypeOIDC {
		t.Errorf("Provider Type = %s, want oidc", provider.Type)
	}

	expectedIssuerURL := "https://example.auth0.com"
	if provider.IssuerURL != expectedIssuerURL {
		t.Errorf("Provider IssuerURL = %s, want %s", provider.IssuerURL, expectedIssuerURL)
	}
}

func TestLoadConfigFromEnv_Auth0Provider_MissingDomain(t *testing.T) {
	// Set Auth0 environment variables without domain
	os.Setenv("ATLANTIS_AUTH0_CLIENT_ID", "auth0-client-id")
	os.Setenv("ATLANTIS_AUTH0_CLIENT_SECRET", "auth0-client-secret")
	os.Setenv("ATLANTIS_AUTH0_REDIRECT_URL", "https://example.com/auth0/callback")
	defer func() {
		os.Unsetenv("ATLANTIS_AUTH0_CLIENT_ID")
		os.Unsetenv("ATLANTIS_AUTH0_CLIENT_SECRET")
		os.Unsetenv("ATLANTIS_AUTH0_REDIRECT_URL")
	}()

	_, err := LoadConfigFromEnv()
	if err == nil {
		t.Error("Expected error for missing ATLANTIS_AUTH0_DOMAIN")
	}
}

func TestLoadConfigFromEnv_MultipleProviders(t *testing.T) {
	// Set multiple provider environment variables
	os.Setenv("ATLANTIS_GOOGLE_CLIENT_ID", "google-client-id")
	os.Setenv("ATLANTIS_GOOGLE_CLIENT_SECRET", "google-client-secret")
	os.Setenv("ATLANTIS_GOOGLE_REDIRECT_URL", "https://example.com/google/callback")
	
	os.Setenv("ATLANTIS_OKTA_CLIENT_ID", "okta-client-id")
	os.Setenv("ATLANTIS_OKTA_CLIENT_SECRET", "okta-client-secret")
	os.Setenv("ATLANTIS_OKTA_REDIRECT_URL", "https://example.com/okta/callback")
	os.Setenv("ATLANTIS_OKTA_ISSUER_URL", "https://example.okta.com")
	
	defer func() {
		os.Unsetenv("ATLANTIS_GOOGLE_CLIENT_ID")
		os.Unsetenv("ATLANTIS_GOOGLE_CLIENT_SECRET")
		os.Unsetenv("ATLANTIS_GOOGLE_REDIRECT_URL")
		os.Unsetenv("ATLANTIS_OKTA_CLIENT_ID")
		os.Unsetenv("ATLANTIS_OKTA_CLIENT_SECRET")
		os.Unsetenv("ATLANTIS_OKTA_REDIRECT_URL")
		os.Unsetenv("ATLANTIS_OKTA_ISSUER_URL")
	}()

	config, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("Failed to load config from env: %v", err)
	}

	if len(config.Providers) != 2 {
		t.Fatalf("Expected 2 providers, got %d", len(config.Providers))
	}

	// Check that both providers are present
	providerIDs := make(map[string]bool)
	for _, provider := range config.Providers {
		providerIDs[provider.ID] = true
	}

	if !providerIDs["google"] {
		t.Error("Google provider should be present")
	}

	if !providerIDs["okta"] {
		t.Error("Okta provider should be present")
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	// Test with environment variable set
	os.Setenv("TEST_VAR", "test-value")
	defer os.Unsetenv("TEST_VAR")

	result := getEnvOrDefault("TEST_VAR", "default-value")
	if result != "test-value" {
		t.Errorf("getEnvOrDefault = %s, want test-value", result)
	}

	// Test with environment variable not set
	result = getEnvOrDefault("NONEXISTENT_VAR", "default-value")
	if result != "default-value" {
		t.Errorf("getEnvOrDefault = %s, want default-value", result)
	}
}

func TestGetEnvBoolOrDefault(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		defaultValue bool
		expected bool
	}{
		{"true string", "true", false, true},
		{"1", "1", false, true},
		{"yes", "yes", false, true},
		{"false string", "false", true, false},
		{"0", "0", true, false},
		{"no", "no", true, false},
		{"empty", "", true, true},
		{"not set", "", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv("TEST_BOOL_VAR", tt.envValue)
				defer os.Unsetenv("TEST_BOOL_VAR")
			}

			result := getEnvBoolOrDefault("TEST_BOOL_VAR", tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvBoolOrDefault = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCreateDefaultConfig(t *testing.T) {
	config := CreateDefaultConfig()

	if config.SessionDuration != 24*time.Hour {
		t.Errorf("SessionDuration = %v, want 24h", config.SessionDuration)
	}

	if config.SessionCookieName != "atlantis_session" {
		t.Errorf("SessionCookieName = %s, want atlantis_session", config.SessionCookieName)
	}

	if config.SessionSecret != "dev-secret-change-in-production" {
		t.Errorf("SessionSecret = %s, want dev-secret-change-in-production", config.SessionSecret)
	}

	if config.SecureCookies {
		t.Error("SecureCookies should be false for default config")
	}

	if config.CSRFSecret != "dev-csrf-secret-change-in-production" {
		t.Errorf("CSRFSecret = %s, want dev-csrf-secret-change-in-production", config.CSRFSecret)
	}

	if !config.EnableBasicAuth {
		t.Error("EnableBasicAuth should be true for default config")
	}

	if config.BasicAuthUser != "atlantis" {
		t.Errorf("BasicAuthUser = %s, want atlantis", config.BasicAuthUser)
	}

	if config.BasicAuthPass != "atlantis" {
		t.Errorf("BasicAuthPass = %s, want atlantis", config.BasicAuthPass)
	}

	if len(config.DefaultRoles) != 1 || config.DefaultRoles[0] != "user" {
		t.Errorf("DefaultRoles = %v, want [user]", config.DefaultRoles)
	}

	if len(config.AdminGroups) != 2 {
		t.Errorf("AdminGroups = %v, want 2 groups", config.AdminGroups)
	}

	if len(config.AdminEmails) != 1 || config.AdminEmails[0] != "admin@example.com" {
		t.Errorf("AdminEmails = %v, want [admin@example.com]", config.AdminEmails)
	}

	if len(config.Providers) != 1 {
		t.Errorf("Providers = %v, want 1 provider", config.Providers)
	}

	provider := config.Providers[0]
	if provider.ID != "basic" {
		t.Errorf("Provider ID = %s, want basic", provider.ID)
	}

	if provider.Name != "Basic Authentication" {
		t.Errorf("Provider Name = %s, want Basic Authentication", provider.Name)
	}

	if provider.Type != ProviderTypeBasicAuth {
		t.Errorf("Provider Type = %s, want basic", provider.Type)
	}
} 