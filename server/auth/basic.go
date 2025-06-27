package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// BasicAuthProvider implements basic HTTP authentication
type BasicAuthProvider struct {
	config *ProviderConfig
}

// NewBasicAuthProvider creates a new basic auth provider
func NewBasicAuthProvider(config *ProviderConfig) (*BasicAuthProvider, error) {
	if config.Type != ProviderTypeBasicAuth {
		return nil, fmt.Errorf("invalid provider type for basic auth: %s", config.Type)
	}

	return &BasicAuthProvider{
		config: config,
	}, nil
}

// GetType returns the provider type
func (p *BasicAuthProvider) GetType() ProviderType {
	return p.config.Type
}

// GetID returns the provider ID
func (p *BasicAuthProvider) GetID() string {
	return p.config.ID
}

// GetName returns the provider name
func (p *BasicAuthProvider) GetName() string {
	return p.config.Name
}

// IsEnabled returns whether the provider is enabled
func (p *BasicAuthProvider) IsEnabled() bool {
	return p.config.Enabled
}

// InitAuthURL is not used for basic auth
func (p *BasicAuthProvider) InitAuthURL(state string) (string, error) {
	return "", fmt.Errorf("auth URL not supported for basic auth")
}

// ExchangeCode is not used for basic auth
func (p *BasicAuthProvider) ExchangeCode(ctx context.Context, code string) (*TokenResponse, error) {
	return nil, fmt.Errorf("code exchange not supported for basic auth")
}

// GetUserInfo is not used for basic auth
func (p *BasicAuthProvider) GetUserInfo(ctx context.Context, token *TokenResponse) (*User, error) {
	return nil, fmt.Errorf("user info not supported for basic auth")
}

// ValidateToken validates basic auth credentials
func (p *BasicAuthProvider) ValidateToken(ctx context.Context, tokenString string) (*User, error) {
	// For basic auth, the token is the base64 encoded credentials
	// Format: "username:password"
	credentials, err := decodeBasicAuth(tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid basic auth token: %w", err)
	}

	parts := strings.SplitN(credentials, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid credentials format")
	}

	username := parts[0]
	password := parts[1]

	// Validate against configured credentials
	if username == p.config.ClientID && password == p.config.ClientSecret {
		user := &User{
			ID:         username,
			Email:      username, // Basic auth doesn't provide email
			Name:       username,
			Provider:   p.config.ID,
			LastLogin:  time.Now(),
			Attributes: make(map[string]string),
			Groups:     []string{},
			Roles:      p.config.DefaultRoles,
		}

		// Check if user is in admin groups
		for _, adminGroup := range p.config.AllowedGroups {
			if username == adminGroup {
				user.Roles = append(user.Roles, "admin")
				break
			}
		}

		// Check if user email is in admin emails
		for _, adminEmail := range p.config.AllowedEmails {
			if username == adminEmail {
				user.Roles = append(user.Roles, "admin")
				break
			}
		}

		return user, nil
	}

	return nil, fmt.Errorf("invalid credentials")
}

// InitiateLogin is not used for basic auth
func (p *BasicAuthProvider) InitiateLogin(w http.ResponseWriter, r *http.Request) error {
	return fmt.Errorf("initiate login not supported for basic auth")
}

// ProcessSAMLResponse is not used for basic auth
func (p *BasicAuthProvider) ProcessSAMLResponse(w http.ResponseWriter, r *http.Request) (*User, error) {
	return nil, fmt.Errorf("SAML response processing not supported for basic auth")
}

// ValidateBasicAuth validates basic auth from HTTP request
func (p *BasicAuthProvider) ValidateBasicAuth(r *http.Request) (*User, error) {
	username, password, ok := r.BasicAuth()
	if !ok {
		return nil, fmt.Errorf("no basic auth credentials provided")
	}

	// Validate against configured credentials
	if username == p.config.ClientID && password == p.config.ClientSecret {
		user := &User{
			ID:         username,
			Email:      username, // Basic auth doesn't provide email
			Name:       username,
			Provider:   p.config.ID,
			LastLogin:  time.Now(),
			Attributes: make(map[string]string),
			Groups:     []string{},
			Roles:      p.config.DefaultRoles,
		}

		// Check if user is in admin groups
		for _, adminGroup := range p.config.AllowedGroups {
			if username == adminGroup {
				user.Roles = append(user.Roles, "admin")
				break
			}
		}

		// Check if user email is in admin emails
		for _, adminEmail := range p.config.AllowedEmails {
			if username == adminEmail {
				user.Roles = append(user.Roles, "admin")
				break
			}
		}

		return user, nil
	}

	return nil, fmt.Errorf("invalid credentials")
}

// Helper function to decode basic auth token
func decodeBasicAuth(token string) (string, error) {
	// Decode base64 encoded credentials
	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}
	return string(decoded), nil
} 