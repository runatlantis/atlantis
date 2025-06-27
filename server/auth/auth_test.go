package auth

import (
	"testing"
	"time"
)

func TestPermissionConstants(t *testing.T) {
	tests := []struct {
		name       string
		permission Permission
		expected   string
	}{
		{"RepoRead", PermissionRepoRead, "repo:read"},
		{"RepoWrite", PermissionRepoWrite, "repo:write"},
		{"RepoDelete", PermissionRepoDelete, "repo:delete"},
		{"LockRead", PermissionLockRead, "lock:read"},
		{"LockCreate", PermissionLockCreate, "lock:create"},
		{"LockDelete", PermissionLockDelete, "lock:delete"},
		{"LockForce", PermissionLockForce, "lock:force"},
		{"PlanRead", PermissionPlanRead, "plan:read"},
		{"PlanCreate", PermissionPlanCreate, "plan:create"},
		{"PlanApply", PermissionPlanApply, "plan:apply"},
		{"PlanDelete", PermissionPlanDelete, "plan:delete"},
		{"PolicyRead", PermissionPolicyRead, "policy:read"},
		{"PolicyWrite", PermissionPolicyWrite, "policy:write"},
		{"AdminRead", PermissionAdminRead, "admin:read"},
		{"AdminWrite", PermissionAdminWrite, "admin:write"},
		{"AdminDelete", PermissionAdminDelete, "admin:delete"},
		{"UserRead", PermissionUserRead, "user:read"},
		{"UserWrite", PermissionUserWrite, "user:write"},
		{"UserDelete", PermissionUserDelete, "user:delete"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.permission) != tt.expected {
				t.Errorf("Permission %s = %s, want %s", tt.name, string(tt.permission), tt.expected)
			}
		})
	}
}

func TestProviderTypeConstants(t *testing.T) {
	tests := []struct {
		name         string
		providerType ProviderType
		expected     string
	}{
		{"OAuth2", ProviderTypeOAuth2, "oauth2"},
		{"OIDC", ProviderTypeOIDC, "oidc"},
		{"SAML", ProviderTypeSAML, "saml"},
		{"BasicAuth", ProviderTypeBasicAuth, "basic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.providerType) != tt.expected {
				t.Errorf("ProviderType %s = %s, want %s", tt.name, string(tt.providerType), tt.expected)
			}
		})
	}
}

func TestUserCreation(t *testing.T) {
	user := &User{
		ID:         "test-user",
		Email:      "test@example.com",
		Name:       "Test User",
		Groups:     []string{"developers"},
		Roles:      []string{"user"},
		Permissions: []Permission{PermissionRepoRead},
		Attributes:  map[string]string{"department": "engineering"},
		Provider:   "google",
		LastLogin:  time.Now(),
	}

	if user.ID != "test-user" {
		t.Errorf("User ID = %s, want test-user", user.ID)
	}

	if user.Email != "test@example.com" {
		t.Errorf("User Email = %s, want test@example.com", user.Email)
	}

	if len(user.Groups) != 1 || user.Groups[0] != "developers" {
		t.Errorf("User Groups = %v, want [developers]", user.Groups)
	}

	if len(user.Roles) != 1 || user.Roles[0] != "user" {
		t.Errorf("User Roles = %v, want [user]", user.Roles)
	}

	if len(user.Permissions) != 1 || user.Permissions[0] != PermissionRepoRead {
		t.Errorf("User Permissions = %v, want [repo:read]", user.Permissions)
	}
}

func TestSessionCreation(t *testing.T) {
	now := time.Now()
	session := &Session{
		ID:        "test-session",
		UserID:    "test-user",
		CreatedAt: now,
		ExpiresAt: now.Add(24 * time.Hour),
	}

	if session.ID != "test-session" {
		t.Errorf("Session ID = %s, want test-session", session.ID)
	}

	if session.UserID != "test-user" {
		t.Errorf("Session UserID = %s, want test-user", session.UserID)
	}

	if session.CreatedAt != now {
		t.Errorf("Session CreatedAt = %v, want %v", session.CreatedAt, now)
	}

	if session.ExpiresAt != now.Add(24*time.Hour) {
		t.Errorf("Session ExpiresAt = %v, want %v", session.ExpiresAt, now.Add(24*time.Hour))
	}
}

func TestTokenResponseCreation(t *testing.T) {
	token := &TokenResponse{
		AccessToken:  "access-token",
		TokenType:    "Bearer",
		RefreshToken: "refresh-token",
		ExpiresIn:    3600,
	}

	if token.AccessToken != "access-token" {
		t.Errorf("Token AccessToken = %s, want access-token", token.AccessToken)
	}

	if token.TokenType != "Bearer" {
		t.Errorf("Token TokenType = %s, want Bearer", token.TokenType)
	}

	if token.RefreshToken != "refresh-token" {
		t.Errorf("Token RefreshToken = %s, want refresh-token", token.RefreshToken)
	}

	if token.ExpiresIn != 3600 {
		t.Errorf("Token ExpiresIn = %d, want 3600", token.ExpiresIn)
	}
}

func TestProviderConfigCreation(t *testing.T) {
	config := &ProviderConfig{
		Type:         ProviderTypeOAuth2,
		ID:           "google",
		Name:         "Google",
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "https://example.com/callback",
		Scopes:       []string{"openid", "email", "profile"},
		Enabled:      true,
		DefaultRoles: []string{"user"},
		AllowedGroups: []string{"admin", "devops"},
		AllowedEmails: []string{"admin@example.com"},
	}

	if config.Type != ProviderTypeOAuth2 {
		t.Errorf("Config Type = %s, want oauth2", config.Type)
	}

	if config.ID != "google" {
		t.Errorf("Config ID = %s, want google", config.ID)
	}

	if config.Name != "Google" {
		t.Errorf("Config Name = %s, want Google", config.Name)
	}

	if !config.Enabled {
		t.Errorf("Config Enabled = %v, want true", config.Enabled)
	}

	if len(config.Scopes) != 3 {
		t.Errorf("Config Scopes = %v, want 3 scopes", config.Scopes)
	}

	if len(config.AllowedGroups) != 2 {
		t.Errorf("Config AllowedGroups = %v, want 2 groups", config.AllowedGroups)
	}
}

func TestDefaultRoles(t *testing.T) {
	// Test that default roles exist
	expectedRoles := []string{"user", "developer", "admin", "superadmin"}
	for _, roleName := range expectedRoles {
		if role, exists := DefaultRoles[roleName]; !exists {
			t.Errorf("Default role %s does not exist", roleName)
		} else if role.Name != roleName {
			t.Errorf("Default role %s has wrong name: %s", roleName, role.Name)
		}
	}

	// Test user role permissions
	userRole := DefaultRoles["user"]
	expectedUserPerms := []Permission{
		PermissionRepoRead,
		PermissionLockRead,
		PermissionPlanRead,
		PermissionPolicyRead,
	}

	for _, expectedPerm := range expectedUserPerms {
		found := false
		for _, perm := range userRole.Permissions {
			if perm == expectedPerm {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("User role missing permission: %s", expectedPerm)
		}
	}

	// Test admin role has more permissions than user
	adminRole := DefaultRoles["admin"]
	if len(adminRole.Permissions) <= len(userRole.Permissions) {
		t.Errorf("Admin role should have more permissions than user role")
	}

	// Test superadmin role has the most permissions
	superadminRole := DefaultRoles["superadmin"]
	if len(superadminRole.Permissions) <= len(adminRole.Permissions) {
		t.Errorf("Superadmin role should have more permissions than admin role")
	}
}

func TestConfigCreation(t *testing.T) {
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
		AdminGroups:       []string{"admin", "devops"},
		AdminEmails:       []string{"admin@example.com"},
		Roles:             make(map[string]Role),
		Providers:         []ProviderConfig{},
	}

	if config.SessionSecret != "test-secret" {
		t.Errorf("Config SessionSecret = %s, want test-secret", config.SessionSecret)
	}

	if config.SessionDuration != 24*time.Hour {
		t.Errorf("Config SessionDuration = %v, want 24h", config.SessionDuration)
	}

	if config.SessionCookieName != "atlantis_session" {
		t.Errorf("Config SessionCookieName = %s, want atlantis_session", config.SessionCookieName)
	}

	if config.SecureCookies {
		t.Errorf("Config SecureCookies = %v, want false", config.SecureCookies)
	}

	if !config.EnableBasicAuth {
		t.Errorf("Config EnableBasicAuth = %v, want true", config.EnableBasicAuth)
	}

	if len(config.DefaultRoles) != 1 || config.DefaultRoles[0] != "user" {
		t.Errorf("Config DefaultRoles = %v, want [user]", config.DefaultRoles)
	}

	if len(config.AdminGroups) != 2 {
		t.Errorf("Config AdminGroups = %v, want 2 groups", config.AdminGroups)
	}
} 