package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// LoadConfigFromFile loads authentication configuration from a JSON file
func LoadConfigFromFile(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if config.SessionDuration == 0 {
		config.SessionDuration = 24 * time.Hour
	}
	if config.SessionCookieName == "" {
		config.SessionCookieName = "atlantis_session"
	}
	if config.SessionSecret == "" {
		config.SessionSecret = "change-me-in-production"
	}

	return &config, nil
}

// LoadConfigFromEnv loads authentication configuration from environment variables
func LoadConfigFromEnv() (*Config, error) {
	config := &Config{
		SessionDuration:   24 * time.Hour,
		SessionCookieName: "atlantis_session",
		SessionSecret:     getEnvOrDefault("ATLANTIS_SESSION_SECRET", "change-me-in-production"),
		SecureCookies:     getEnvBoolOrDefault("ATLANTIS_SECURE_COOKIES", false),
		CSRFSecret:        getEnvOrDefault("ATLANTIS_CSRF_SECRET", "change-me-in-production"),
		EnableBasicAuth:   getEnvBoolOrDefault("ATLANTIS_ENABLE_BASIC_AUTH", false),
		BasicAuthUser:     getEnvOrDefault("ATLANTIS_BASIC_AUTH_USER", "atlantis"),
		BasicAuthPass:     getEnvOrDefault("ATLANTIS_BASIC_AUTH_PASS", "atlantis"),
	}

	// Load providers from environment
	providers, err := loadProvidersFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to load providers from environment: %w", err)
	}
	config.Providers = providers

	return config, nil
}

// loadProvidersFromEnv loads provider configurations from environment variables
func loadProvidersFromEnv() ([]ProviderConfig, error) {
	var providers []ProviderConfig

	// Google OAuth2
	if clientID := os.Getenv("ATLANTIS_GOOGLE_CLIENT_ID"); clientID != "" {
		providers = append(providers, ProviderConfig{
			Type:         ProviderTypeOAuth2,
			ID:           "google",
			Name:         "Google",
			ClientID:     clientID,
			ClientSecret: os.Getenv("ATLANTIS_GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("ATLANTIS_GOOGLE_REDIRECT_URL"),
			Scopes:       []string{"openid", "email", "profile"},
			Enabled:      true,
			DefaultRoles: []string{"user"},
		})
	}

	// Okta OIDC
	if clientID := os.Getenv("ATLANTIS_OKTA_CLIENT_ID"); clientID != "" {
		issuerURL := os.Getenv("ATLANTIS_OKTA_ISSUER_URL")
		if issuerURL == "" {
			return nil, fmt.Errorf("ATLANTIS_OKTA_ISSUER_URL is required for Okta provider")
		}

		providers = append(providers, ProviderConfig{
			Type:         ProviderTypeOIDC,
			ID:           "okta",
			Name:         "Okta",
			ClientID:     clientID,
			ClientSecret: os.Getenv("ATLANTIS_OKTA_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("ATLANTIS_OKTA_REDIRECT_URL"),
			IssuerURL:    issuerURL,
			Scopes:       []string{"openid", "email", "profile", "groups"},
			Enabled:      true,
			DefaultRoles: []string{"user"},
		})
	}

	// Azure AD
	if clientID := os.Getenv("ATLANTIS_AZURE_CLIENT_ID"); clientID != "" {
		tenantID := os.Getenv("ATLANTIS_AZURE_TENANT_ID")
		if tenantID == "" {
			return nil, fmt.Errorf("ATLANTIS_AZURE_TENANT_ID is required for Azure provider")
		}

		issuerURL := fmt.Sprintf("https://login.microsoftonline.com/%s", tenantID)
		providers = append(providers, ProviderConfig{
			Type:         ProviderTypeOIDC,
			ID:           "azure",
			Name:         "Azure AD",
			ClientID:     clientID,
			ClientSecret: os.Getenv("ATLANTIS_AZURE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("ATLANTIS_AZURE_REDIRECT_URL"),
			IssuerURL:    issuerURL,
			Scopes:       []string{"openid", "email", "profile", "User.Read"},
			Enabled:      true,
			DefaultRoles: []string{"user"},
		})
	}

	// Auth0 OIDC
	if clientID := os.Getenv("ATLANTIS_AUTH0_CLIENT_ID"); clientID != "" {
		domain := os.Getenv("ATLANTIS_AUTH0_DOMAIN")
		if domain == "" {
			return nil, fmt.Errorf("ATLANTIS_AUTH0_DOMAIN is required for Auth0 provider")
		}

		issuerURL := fmt.Sprintf("https://%s", domain)
		providers = append(providers, ProviderConfig{
			Type:         ProviderTypeOIDC,
			ID:           "auth0",
			Name:         "Auth0",
			ClientID:     clientID,
			ClientSecret: os.Getenv("ATLANTIS_AUTH0_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("ATLANTIS_AUTH0_REDIRECT_URL"),
			IssuerURL:    issuerURL,
			Scopes:       []string{"openid", "email", "profile"},
			Enabled:      true,
			DefaultRoles: []string{"user"},
		})
	}

	return providers, nil
}

// Helper functions
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}

// CreateDefaultConfig creates a default configuration for development
func CreateDefaultConfig() *Config {
	return &Config{
		SessionDuration:   24 * time.Hour,
		SessionCookieName: "atlantis_session",
		SessionSecret:     "dev-secret-change-in-production",
		SecureCookies:     false,
		CSRFSecret:        "dev-csrf-secret-change-in-production",
		EnableBasicAuth:   true,
		BasicAuthUser:     "atlantis",
		BasicAuthPass:     "atlantis",
		DefaultRoles:      []string{"user"},
		AdminGroups:       []string{"admin", "administrators"},
		AdminEmails:       []string{"admin@example.com"},
		Providers: []ProviderConfig{
			{
				Type:         ProviderTypeBasicAuth,
				ID:           "basic",
				Name:         "Basic Authentication",
				ClientID:     "atlantis",
				ClientSecret: "atlantis",
				Enabled:      true,
				DefaultRoles: []string{"user"},
			},
		},
	}
} 