package auth

import (
	"testing"
)

func TestNewPermissionChecker(t *testing.T) {
	customRoles := map[string]Role{
		"custom-role": {
			Name:        "custom-role",
			Permissions: []Permission{PermissionRepoRead, PermissionLockRead},
			Description: "Custom role for testing",
		},
	}

	checker := NewPermissionChecker(customRoles)

	// Test that default roles are included
	if !checker.HasPermission(&User{Roles: []string{"user"}}, PermissionRepoRead) {
		t.Error("Default user role should have repo:read permission")
	}

	// Test that custom roles are included
	if !checker.HasPermission(&User{Roles: []string{"custom-role"}}, PermissionRepoRead) {
		t.Error("Custom role should have repo:read permission")
	}
}

func TestHasPermission(t *testing.T) {
	checker := NewPermissionChecker(nil)

	user := &User{
		ID:         "test-user",
		Email:      "test@example.com",
		Roles:      []string{"user", "developer"},
		Permissions: []Permission{PermissionLockForce}, // Direct permission
	}

	// Test direct permissions
	if !checker.HasPermission(user, PermissionLockForce) {
		t.Error("User should have direct lock:force permission")
	}

	// Test role-based permissions
	if !checker.HasPermission(user, PermissionRepoRead) {
		t.Error("User should have repo:read permission from user role")
	}

	if !checker.HasPermission(user, PermissionRepoWrite) {
		t.Error("User should have repo:write permission from developer role")
	}

	// Test non-existent permissions
	if checker.HasPermission(user, PermissionUserDelete) {
		t.Error("User should not have user:delete permission")
	}

	// Test with nil user
	if checker.HasPermission(nil, PermissionRepoRead) {
		t.Error("Nil user should not have any permissions")
	}
}

func TestHasAnyPermission(t *testing.T) {
	checker := NewPermissionChecker(nil)

	user := &User{
		ID:         "test-user",
		Email:      "test@example.com",
		Roles:      []string{"user"},
		Permissions: []Permission{},
	}

	// Test with permissions user has
	permissions := []Permission{PermissionRepoRead, PermissionUserDelete}
	if !checker.HasAnyPermission(user, permissions) {
		t.Error("User should have at least one of the permissions")
	}

	// Test with permissions user doesn't have
	permissions = []Permission{PermissionUserDelete, PermissionUserWrite}
	if checker.HasAnyPermission(user, permissions) {
		t.Error("User should not have any of these permissions")
	}

	// Test with empty permissions list
	if checker.HasAnyPermission(user, []Permission{}) {
		t.Error("Empty permissions list should return false")
	}
}

func TestHasAllPermissions(t *testing.T) {
	checker := NewPermissionChecker(nil)

	user := &User{
		ID:         "test-user",
		Email:      "test@example.com",
		Roles:      []string{"user", "developer"},
		Permissions: []Permission{},
	}

	// Test with permissions user has
	permissions := []Permission{PermissionRepoRead, PermissionRepoWrite}
	if !checker.HasAllPermissions(user, permissions) {
		t.Error("User should have all of these permissions")
	}

	// Test with permissions user doesn't have all of
	permissions = []Permission{PermissionRepoRead, PermissionUserDelete}
	if checker.HasAllPermissions(user, permissions) {
		t.Error("User should not have all of these permissions")
	}

	// Test with empty permissions list
	if !checker.HasAllPermissions(user, []Permission{}) {
		t.Error("Empty permissions list should return true")
	}
}

func TestGetUserPermissions(t *testing.T) {
	checker := NewPermissionChecker(nil)

	user := &User{
		ID:         "test-user",
		Email:      "test@example.com",
		Roles:      []string{"user", "developer"},
		Permissions: []Permission{PermissionLockForce}, // Direct permission
	}

	permissions := checker.GetUserPermissions(user)

	// Check that permissions from roles are included
	expectedFromRoles := []Permission{
		PermissionRepoRead,  // from user role
		PermissionRepoWrite, // from developer role
		PermissionLockRead,  // from user role
		PermissionLockCreate, // from developer role
		PermissionPlanRead,  // from user role
		PermissionPlanCreate, // from developer role
		PermissionPolicyRead, // from user role
	}

	for _, expected := range expectedFromRoles {
		found := false
		for _, perm := range permissions {
			if perm == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Permission %s should be included in user permissions", expected)
		}
	}

	// Check that direct permissions are included
	found := false
	for _, perm := range permissions {
		if perm == PermissionLockForce {
			found = true
			break
		}
	}
	if !found {
		t.Error("Direct permission lock:force should be included in user permissions")
	}

	// Test with nil user
	permissions = checker.GetUserPermissions(nil)
	if len(permissions) != 0 {
		t.Error("Nil user should have no permissions")
	}
}

func TestGetUserRoles(t *testing.T) {
	checker := NewPermissionChecker(nil)

	user := &User{
		ID:         "test-user",
		Email:      "test@example.com",
		Roles:      []string{"user", "developer"},
		Permissions: []Permission{},
	}

	roles := checker.GetUserRoles(user)

	if len(roles) != 2 {
		t.Errorf("Expected 2 roles, got %d", len(roles))
	}

	// Check that roles are correctly retrieved
	roleNames := make(map[string]bool)
	for _, role := range roles {
		roleNames[role.Name] = true
	}

	if !roleNames["user"] {
		t.Error("User role should be included")
	}

	if !roleNames["developer"] {
		t.Error("Developer role should be included")
	}

	// Test with nil user
	roles = checker.GetUserRoles(nil)
	if len(roles) != 0 {
		t.Error("Nil user should have no roles")
	}
}

func TestAddRole(t *testing.T) {
	checker := NewPermissionChecker(nil)

	newRole := Role{
		Name:        "test-role",
		Permissions: []Permission{PermissionRepoRead, PermissionLockRead},
		Description: "Test role",
	}

	err := checker.AddRole(newRole)
	if err != nil {
		t.Errorf("Failed to add role: %v", err)
	}

	// Test that the role was added
	role, exists := checker.GetRole("test-role")
	if !exists {
		t.Error("Added role should exist")
	}

	if role.Name != "test-role" {
		t.Errorf("Role name = %s, want test-role", role.Name)
	}

	if len(role.Permissions) != 2 {
		t.Errorf("Role permissions = %d, want 2", len(role.Permissions))
	}

	// Test adding role with empty name
	err = checker.AddRole(Role{Name: ""})
	if err == nil {
		t.Error("Should fail to add role with empty name")
	}
}

func TestRemoveRole(t *testing.T) {
	customRoles := map[string]Role{
		"custom-role": {
			Name:        "custom-role",
			Permissions: []Permission{PermissionRepoRead},
			Description: "Custom role",
		},
	}

	checker := NewPermissionChecker(customRoles)

	// Test removing custom role
	err := checker.RemoveRole("custom-role")
	if err != nil {
		t.Errorf("Failed to remove role: %v", err)
	}

	// Test that the role was removed
	_, exists := checker.GetRole("custom-role")
	if exists {
		t.Error("Removed role should not exist")
	}

	// Test removing default role (should fail)
	err = checker.RemoveRole("user")
	if err == nil {
		t.Error("Should fail to remove default role")
	}

	// Test removing non-existent role
	err = checker.RemoveRole("non-existent")
	if err == nil {
		t.Error("Should fail to remove non-existent role")
	}
}

func TestHelperFunctions(t *testing.T) {
	checker := NewPermissionChecker(nil)

	user := &User{
		ID:         "test-user",
		Email:      "test@example.com",
		Roles:      []string{"admin"},
		Permissions: []Permission{},
	}

	// Test CanDeleteLock
	if !checker.CanDeleteLock(user) {
		t.Error("Admin should be able to delete locks")
	}

	// Test CanForceDeleteLock
	if !checker.CanForceDeleteLock(user) {
		t.Error("Admin should be able to force delete locks")
	}

	// Test CanApplyPlan
	if !checker.CanApplyPlan(user) {
		t.Error("Admin should be able to apply plans")
	}

	// Test CanDeletePlan
	if !checker.CanDeletePlan(user) {
		t.Error("Admin should be able to delete plans")
	}

	// Test CanWritePolicy
	if !checker.CanWritePolicy(user) {
		t.Error("Admin should be able to write policies")
	}

	// Test CanManageUsers (should be superadmin)
	superadmin := &User{
		ID:         "superadmin-user",
		Email:      "superadmin@example.com",
		Roles:      []string{"superadmin"},
		Permissions: []Permission{},
	}
	if !checker.CanManageUsers(superadmin) {
		t.Error("Superadmin should be able to manage users")
	}

	// Test IsAdmin
	if !checker.IsAdmin(user) {
		t.Error("User with admin role should be considered admin")
	}

	// Test IsSuperAdmin (should be superadmin)
	if !checker.IsSuperAdmin(superadmin) {
		t.Error("Superadmin should be considered super admin")
	}

	// Test with regular user
	regularUser := &User{
		ID:         "regular-user",
		Email:      "regular@example.com",
		Roles:      []string{"user"},
		Permissions: []Permission{},
	}

	if checker.CanDeleteLock(regularUser) {
		t.Error("Regular user should not be able to delete locks")
	}

	if checker.IsAdmin(regularUser) {
		t.Error("Regular user should not be considered admin")
	}
}

func TestParsePermission(t *testing.T) {
	tests := []struct {
		name        string
		permission  string
		expectError bool
	}{
		{"Valid repo:read", "repo:read", false},
		{"Valid lock:delete", "lock:delete", false},
		{"Valid admin:write", "admin:write", false},
		{"Invalid permission", "invalid:permission", true},
		{"Empty permission", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perm, err := ParsePermission(tt.permission)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if string(perm) != tt.permission {
					t.Errorf("Parsed permission = %s, want %s", string(perm), tt.permission)
				}
			}
		})
	}
}

func TestParsePermissions(t *testing.T) {
	permissions := []string{"repo:read", "lock:delete", "admin:write"}
	parsed, err := ParsePermissions(permissions)
	if err != nil {
		t.Errorf("Failed to parse permissions: %v", err)
	}

	if len(parsed) != 3 {
		t.Errorf("Expected 3 permissions, got %d", len(parsed))
	}

	// Test with invalid permission
	permissions = []string{"repo:read", "invalid:permission"}
	_, err = ParsePermissions(permissions)
	if err == nil {
		t.Error("Expected error for invalid permission")
	}
}

func TestPermissionToString(t *testing.T) {
	perm := PermissionRepoRead
	str := PermissionToString(perm)
	if str != "repo:read" {
		t.Errorf("PermissionToString = %s, want repo:read", str)
	}
}

func TestGetPermissionDescription(t *testing.T) {
	tests := []struct {
		permission Permission
		expected   string
	}{
		{PermissionRepoRead, "Read repository information"},
		{PermissionLockDelete, "Delete locks"},
		{PermissionAdminWrite, "Write admin settings"},
		{Permission("invalid"), "Unknown permission"},
	}

	for _, tt := range tests {
		desc := GetPermissionDescription(tt.permission)
		if desc != tt.expected {
			t.Errorf("Description for %s = %s, want %s", tt.permission, desc, tt.expected)
		}
	}
}

func TestGetPermissionCategory(t *testing.T) {
	tests := []struct {
		permission Permission
		expected   string
	}{
		{PermissionRepoRead, "repo"},
		{PermissionLockDelete, "lock"},
		{PermissionAdminWrite, "admin"},
		{Permission("invalid"), "unknown"},
	}

	for _, tt := range tests {
		category := GetPermissionCategory(tt.permission)
		if category != tt.expected {
			t.Errorf("Category for %s = %s, want %s", tt.permission, category, tt.expected)
		}
	}
} 