package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ProviderType represents the type of SSO provider
type ProviderType string

const (
	ProviderTypeOAuth2    ProviderType = "oauth2"
	ProviderTypeOIDC      ProviderType = "oidc"
	ProviderTypeSAML      ProviderType = "saml"
	ProviderTypeBasicAuth ProviderType = "basic"
)

// Permission represents a specific action that can be performed
type Permission string

const (
	// Repository permissions
	PermissionRepoRead    Permission = "repo:read"
	PermissionRepoWrite   Permission = "repo:write"
	PermissionRepoDelete  Permission = "repo:delete"
	
	// Lock permissions
	PermissionLockRead    Permission = "lock:read"
	PermissionLockCreate  Permission = "lock:create"
	PermissionLockDelete  Permission = "lock:delete"
	PermissionLockForce   Permission = "lock:force"
	
	// Plan permissions
	PermissionPlanRead    Permission = "plan:read"
	PermissionPlanCreate  Permission = "plan:create"
	PermissionPlanApply   Permission = "plan:apply"
	PermissionPlanDelete  Permission = "plan:delete"
	
	// Policy permissions
	PermissionPolicyRead  Permission = "policy:read"
	PermissionPolicyWrite Permission = "policy:write"
	
	// Admin permissions
	PermissionAdminRead   Permission = "admin:read"
	PermissionAdminWrite  Permission = "admin:write"
	PermissionAdminDelete Permission = "admin:delete"
	
	// User management permissions
	PermissionUserRead    Permission = "user:read"
	PermissionUserWrite   Permission = "user:write"
	PermissionUserDelete  Permission = "user:delete"
)

// Role represents a collection of permissions
type Role struct {
	Name        string       `json:"name"`
	Permissions []Permission `json:"permissions"`
	Description string       `json:"description"`
}

// User represents an authenticated user
type User struct {
	ID          string            `json:"id"`
	Email       string            `json:"email"`
	Name        string            `json:"name"`
	Groups      []string          `json:"groups"`
	Roles       []string          `json:"roles"`
	Permissions []Permission      `json:"permissions"`
	Attributes  map[string]string `json:"attributes"`
	Provider    string            `json:"provider"`
	LastLogin   time.Time         `json:"last_login"`
	SessionID   string            `json:"session_id,omitempty"`
}

// Session represents a user session
type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// TokenResponse represents an OAuth2 token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
}

// ProviderConfig represents the configuration for an authentication provider
type ProviderConfig struct {
	Type         ProviderType `json:"type"`
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	ClientID     string       `json:"client_id"`
	ClientSecret string       `json:"client_secret"`
	RedirectURL  string       `json:"redirect_url,omitempty"`
	IssuerURL    string       `json:"issuer_url,omitempty"`
	AuthURL      string       `json:"auth_url,omitempty"`
	TokenURL     string       `json:"token_url,omitempty"`
	UserInfoURL  string       `json:"user_info_url,omitempty"`
	Scopes       []string     `json:"scopes,omitempty"`
	Enabled      bool         `json:"enabled"`
	DefaultRoles []string     `json:"default_roles,omitempty"`
	AllowedGroups []string     `json:"allowed_groups,omitempty"`
	AllowedEmails []string     `json:"allowed_emails,omitempty"`
}

// Config represents the authentication configuration
type Config struct {
	SessionSecret     string           `json:"session_secret"`
	SessionDuration   time.Duration    `json:"session_duration"`
	SessionCookieName string           `json:"session_cookie_name"`
	SecureCookies     bool             `json:"secure_cookies"`
	CSRFSecret        string           `json:"csrf_secret"`
	EnableBasicAuth   bool             `json:"enable_basic_auth"`
	BasicAuthUser     string           `json:"basic_auth_user"`
	BasicAuthPass     string           `json:"basic_auth_pass"`
	DefaultRoles      []string         `json:"default_roles"`
	AdminGroups       []string         `json:"admin_groups"`
	AdminEmails       []string         `json:"admin_emails"`
	Roles             map[string]Role  `json:"roles"`
	Providers         []ProviderConfig `json:"providers"`
}

// UnmarshalJSON implements custom JSON unmarshaling for Config
func (c *Config) UnmarshalJSON(data []byte) error {
	// Create a temporary struct to handle the raw JSON
	type configAlias Config
	aux := &struct {
		SessionDuration string `json:"session_duration"`
		*configAlias
	}{
		configAlias: (*configAlias)(c),
	}
	
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	
	// Parse session duration if it's a string
	if aux.SessionDuration != "" {
		duration, err := time.ParseDuration(aux.SessionDuration)
		if err != nil {
			return fmt.Errorf("invalid session_duration: %w", err)
		}
		c.SessionDuration = duration
	}
	
	return nil
}

// Provider interface for authentication providers
type Provider interface {
	GetType() ProviderType
	GetID() string
	GetName() string
	IsEnabled() bool
	InitAuthURL(state string) (string, error)
	ExchangeCode(ctx context.Context, code string) (*TokenResponse, error)
	GetUserInfo(ctx context.Context, token *TokenResponse) (*User, error)
	ValidateToken(ctx context.Context, tokenString string) (*User, error)
	InitiateLogin(w http.ResponseWriter, r *http.Request) error
	ProcessSAMLResponse(w http.ResponseWriter, r *http.Request) (*User, error)
}

// Manager interface for authentication management
type Manager interface {
	GetProvider(id string) (Provider, error)
	GetEnabledProviders() []Provider
	AuthenticateUser(ctx context.Context, user *User) (*Session, error)
	ValidateSession(ctx context.Context, sessionID string) (*User, error)
	InvalidateSession(ctx context.Context, sessionID string) error
	GetUserFromRequest(r *http.Request) (*User, error)
	LoginRequired(r *http.Request) bool
	RedirectToLogin(w http.ResponseWriter, r *http.Request) error
	GetPermissionChecker() PermissionChecker
}

// PermissionChecker interface for checking user permissions
type PermissionChecker interface {
	HasPermission(user *User, permission Permission) bool
	HasAnyPermission(user *User, permissions []Permission) bool
	HasAllPermissions(user *User, permissions []Permission) bool
	GetUserPermissions(user *User) []Permission
	IsAdmin(user *User) bool
	IsSuperAdmin(user *User) bool
	CanDeleteLock(user *User) bool
	CanForceDeleteLock(user *User) bool
	CanApplyPlan(user *User) bool
	CanDeletePlan(user *User) bool
	CanWritePolicy(user *User) bool
	CanManageUsers(user *User) bool
}

// Default roles with permissions
var DefaultRoles = map[string]Role{
	"user": {
		Name: "user",
		Permissions: []Permission{
			PermissionRepoRead,
			PermissionLockRead,
			PermissionPlanRead,
			PermissionPolicyRead,
		},
		Description: "Basic user with read access",
	},
	"developer": {
		Name: "developer",
		Permissions: []Permission{
			PermissionRepoRead,
			PermissionRepoWrite,
			PermissionLockRead,
			PermissionLockCreate,
			PermissionPlanRead,
			PermissionPlanCreate,
			PermissionPolicyRead,
		},
		Description: "Developer with write access",
	},
	"admin": {
		Name: "admin",
		Permissions: []Permission{
			PermissionRepoRead,
			PermissionRepoWrite,
			PermissionRepoDelete,
			PermissionLockRead,
			PermissionLockCreate,
			PermissionLockDelete,
			PermissionLockForce,
			PermissionPlanRead,
			PermissionPlanCreate,
			PermissionPlanApply,
			PermissionPlanDelete,
			PermissionPolicyRead,
			PermissionPolicyWrite,
			PermissionAdminRead,
			PermissionAdminWrite,
			PermissionUserRead,
		},
		Description: "Administrator with full access",
	},
	"superadmin": {
		Name: "superadmin",
		Permissions: []Permission{
			PermissionRepoRead,
			PermissionRepoWrite,
			PermissionRepoDelete,
			PermissionLockRead,
			PermissionLockCreate,
			PermissionLockDelete,
			PermissionLockForce,
			PermissionPlanRead,
			PermissionPlanCreate,
			PermissionPlanApply,
			PermissionPlanDelete,
			PermissionPolicyRead,
			PermissionPolicyWrite,
			PermissionAdminRead,
			PermissionAdminWrite,
			PermissionAdminDelete,
			PermissionUserRead,
			PermissionUserWrite,
			PermissionUserDelete,
		},
		Description: "Super administrator with all permissions",
	},
} 