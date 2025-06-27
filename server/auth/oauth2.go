package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

// OAuth2Provider implements OAuth2/OIDC authentication
type OAuth2Provider struct {
	config     *ProviderConfig
	oauth2Config *oauth2.Config
	issuerURL  string
	userInfoURL string
}

// NewOAuth2Provider creates a new OAuth2 provider
func NewOAuth2Provider(config *ProviderConfig) (*OAuth2Provider, error) {
	if config.Type != ProviderTypeOAuth2 && config.Type != ProviderTypeOIDC {
		return nil, fmt.Errorf("invalid provider type for OAuth2: %s", config.Type)
	}

	// Validate required fields
	if config.ID == "" {
		return nil, fmt.Errorf("provider ID is required")
	}
	if config.Name == "" {
		return nil, fmt.Errorf("provider name is required")
	}
	if config.ClientID == "" {
		return nil, fmt.Errorf("client ID is required")
	}
	if config.ClientSecret == "" {
		return nil, fmt.Errorf("client secret is required")
	}

	oauth2Config := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       config.Scopes,
	}

	// Set endpoints based on provider type
	switch config.ID {
	case "google":
		oauth2Config.Endpoint = oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		}
		if config.UserInfoURL == "" {
			config.UserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
		}
	case "okta":
		if config.IssuerURL == "" {
			return nil, fmt.Errorf("issuer URL is required for Okta provider")
		}
		oauth2Config.Endpoint = oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/oauth2/v1/authorize", config.IssuerURL),
			TokenURL: fmt.Sprintf("%s/oauth2/v1/token", config.IssuerURL),
		}
		if config.UserInfoURL == "" {
			config.UserInfoURL = fmt.Sprintf("%s/oauth2/v1/userinfo", config.IssuerURL)
		}
	case "azure":
		if config.IssuerURL == "" {
			return nil, fmt.Errorf("issuer URL is required for Azure provider")
		}
		oauth2Config.Endpoint = oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/oauth2/v2.0/authorize", config.IssuerURL),
			TokenURL: fmt.Sprintf("%s/oauth2/v2.0/token", config.IssuerURL),
		}
		if config.UserInfoURL == "" {
			config.UserInfoURL = "https://graph.microsoft.com/v1.0/me"
		}
	default:
		// Custom provider
		if config.AuthURL == "" || config.TokenURL == "" {
			return nil, fmt.Errorf("custom OAuth2 provider must specify auth_url and token_url")
		}
		oauth2Config.Endpoint = oauth2.Endpoint{
			AuthURL:  config.AuthURL,
			TokenURL: config.TokenURL,
		}
	}

	return &OAuth2Provider{
		config:       config,
		oauth2Config: oauth2Config,
		issuerURL:    config.IssuerURL,
		userInfoURL:  config.UserInfoURL,
	}, nil
}

// GetType returns the provider type
func (p *OAuth2Provider) GetType() ProviderType {
	return p.config.Type
}

// GetID returns the provider ID
func (p *OAuth2Provider) GetID() string {
	return p.config.ID
}

// GetName returns the provider name
func (p *OAuth2Provider) GetName() string {
	return p.config.Name
}

// IsEnabled returns whether the provider is enabled
func (p *OAuth2Provider) IsEnabled() bool {
	return p.config.Enabled
}

// InitAuthURL generates the authorization URL
func (p *OAuth2Provider) InitAuthURL(state string) (string, error) {
	authURL := p.oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return authURL, nil
}

// ExchangeCode exchanges authorization code for tokens
func (p *OAuth2Provider) ExchangeCode(ctx context.Context, code string) (*TokenResponse, error) {
	token, err := p.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	return &TokenResponse{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		RefreshToken: token.RefreshToken,
		ExpiresIn:    int(time.Until(token.Expiry).Seconds()),
	}, nil
}

// GetUserInfo retrieves user information from the provider
func (p *OAuth2Provider) GetUserInfo(ctx context.Context, token *TokenResponse) (*User, error) {
	client := p.oauth2Config.Client(ctx, &oauth2.Token{
		AccessToken: token.AccessToken,
		TokenType:   token.TokenType,
	})

	resp, err := client.Get(p.userInfoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer func() {
		_ = resp.Body.Close() // ignore error, as defer cannot return values
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user info request failed with status: %d", resp.StatusCode)
	}

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	user := &User{
		ID:         extractString(userInfo, "id", "sub"),
		Email:      extractString(userInfo, "email"),
		Name:       extractString(userInfo, "name", "display_name"),
		Provider:   p.config.ID,
		LastLogin:  time.Now(),
		Attributes: make(map[string]string),
		Groups:     extractGroups(userInfo),
		Roles:      p.config.DefaultRoles,
	}

	// Extract additional attributes
	for key, value := range userInfo {
		if str, ok := value.(string); ok {
			user.Attributes[key] = str
		}
	}

	// Apply role mapping based on groups/attributes
	user.Roles = p.mapUserRoles(user)

	return user, nil
}

// ValidateToken validates a token and returns user info
func (p *OAuth2Provider) ValidateToken(ctx context.Context, tokenString string) (*User, error) {
	// For OIDC, we can validate the ID token
	if p.config.Type == ProviderTypeOIDC {
		return p.validateIDToken(ctx, tokenString)
	}

	// For OAuth2, we need to use the token to get user info
	token := &oauth2.Token{AccessToken: tokenString}
	client := p.oauth2Config.Client(ctx, token)
	
	resp, err := client.Get(p.userInfoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}
	defer func() {
		_ = resp.Body.Close() // ignore error, as defer cannot return values
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token validation failed with status: %d", resp.StatusCode)
	}

	return p.GetUserInfo(ctx, &TokenResponse{AccessToken: tokenString})
}

// InitiateLogin redirects to the OAuth2 authorization URL
func (p *OAuth2Provider) InitiateLogin(w http.ResponseWriter, r *http.Request) error {
	state, err := generateState()
	if err != nil {
		return fmt.Errorf("failed to generate state: %w", err)
	}

	authURL, err := p.InitAuthURL(state)
	if err != nil {
		return fmt.Errorf("failed to generate auth URL: %w", err)
	}

	http.Redirect(w, r, authURL, http.StatusFound)
	return nil
}

// ProcessSAMLResponse is not used for OAuth2/OIDC
func (p *OAuth2Provider) ProcessSAMLResponse(w http.ResponseWriter, r *http.Request) (*User, error) {
	return nil, fmt.Errorf("SAML response processing not supported for OAuth2/OIDC")
}

// validateIDToken validates an OIDC ID token
func (p *OAuth2Provider) validateIDToken(ctx context.Context, tokenString string) (*User, error) {
	// Parse and validate the ID token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// In a real implementation, you would fetch the public key from the issuer
		// For now, we'll skip signature validation
		return []byte(""), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse ID token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		user := &User{
			ID:         extractStringFromClaims(claims, "sub"),
			Email:      extractStringFromClaims(claims, "email"),
			Name:       extractStringFromClaims(claims, "name"),
			Provider:   p.config.ID,
			LastLogin:  time.Now(),
			Attributes: make(map[string]string),
			Groups:     extractGroupsFromClaims(claims),
			Roles:      p.config.DefaultRoles,
		}

		// Extract additional claims
		for key, value := range claims {
			if str, ok := value.(string); ok {
				user.Attributes[key] = str
			}
		}

		user.Roles = p.mapUserRoles(user)
		return user, nil
	}

	return nil, fmt.Errorf("invalid ID token")
}

// mapUserRoles maps user attributes to roles
func (p *OAuth2Provider) mapUserRoles(user *User) []string {
	roles := make([]string, 0)
	roles = append(roles, p.config.DefaultRoles...)

	// Check if user is in admin groups
	for _, group := range user.Groups {
		for _, adminGroup := range p.config.AllowedGroups {
			if group == adminGroup {
				roles = append(roles, "admin")
				break
			}
		}
	}

	// Check if user email is in admin emails
	for _, adminEmail := range p.config.AllowedEmails {
		if user.Email == adminEmail {
			roles = append(roles, "admin")
			break
		}
	}

	return roles
}

// Helper functions
func extractString(data map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if value, ok := data[key]; ok {
			if str, ok := value.(string); ok {
				return str
			}
		}
	}
	return ""
}

func extractGroups(data map[string]interface{}) []string {
	groups := make([]string, 0)
	
	// Try different possible group fields
	groupFields := []string{"groups", "roles", "memberOf", "groups"}
	
	for _, field := range groupFields {
		if value, ok := data[field]; ok {
			switch v := value.(type) {
			case []interface{}:
				for _, group := range v {
					if str, ok := group.(string); ok {
						groups = append(groups, str)
					}
				}
			case []string:
				groups = append(groups, v...)
			}
		}
	}
	
	return groups
}

func extractStringFromClaims(claims jwt.MapClaims, key string) string {
	if value, ok := claims[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

func extractGroupsFromClaims(claims jwt.MapClaims) []string {
	groups := make([]string, 0)
	
	groupFields := []string{"groups", "roles", "memberOf"}
	
	for _, field := range groupFields {
		if value, ok := claims[field]; ok {
			switch v := value.(type) {
			case []interface{}:
				for _, group := range v {
					if str, ok := group.(string); ok {
						groups = append(groups, str)
					}
				}
			case []string:
				groups = append(groups, v...)
			}
		}
	}
	
	return groups
} 