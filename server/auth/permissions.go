package auth

import (
	"fmt"
	"strings"
)

// PermissionCheckerImpl implements the PermissionChecker interface
type PermissionCheckerImpl struct {
	roles map[string]Role
}

// NewPermissionChecker creates a new permission checker
func NewPermissionChecker(roles map[string]Role) *PermissionCheckerImpl {
	// Merge with default roles
	allRoles := make(map[string]Role)
	for name, role := range DefaultRoles {
		allRoles[name] = role
	}
	for name, role := range roles {
		allRoles[name] = role
	}

	return &PermissionCheckerImpl{
		roles: allRoles,
	}
}

// HasPermission checks if a user has a specific permission
func (pc *PermissionCheckerImpl) HasPermission(user *User, permission Permission) bool {
	if user == nil {
		return false
	}

	// Check direct permissions first
	for _, userPerm := range user.Permissions {
		if userPerm == permission {
			return true
		}
	}

	// Check permissions from roles
	for _, roleName := range user.Roles {
		if role, exists := pc.roles[roleName]; exists {
			for _, rolePerm := range role.Permissions {
				if rolePerm == permission {
					return true
				}
			}
		}
	}

	return false
}

// HasAnyPermission checks if a user has any of the specified permissions
func (pc *PermissionCheckerImpl) HasAnyPermission(user *User, permissions []Permission) bool {
	for _, permission := range permissions {
		if pc.HasPermission(user, permission) {
			return true
		}
	}
	return false
}

// HasAllPermissions checks if a user has all of the specified permissions
func (pc *PermissionCheckerImpl) HasAllPermissions(user *User, permissions []Permission) bool {
	for _, permission := range permissions {
		if !pc.HasPermission(user, permission) {
			return false
		}
	}
	return true
}

// GetUserPermissions returns all permissions for a user
func (pc *PermissionCheckerImpl) GetUserPermissions(user *User) []Permission {
	if user == nil {
		return []Permission{}
	}

	permissionSet := make(map[Permission]bool)

	// Add direct permissions
	for _, perm := range user.Permissions {
		permissionSet[perm] = true
	}

	// Add permissions from roles
	for _, roleName := range user.Roles {
		if role, exists := pc.roles[roleName]; exists {
			for _, perm := range role.Permissions {
				permissionSet[perm] = true
			}
		}
	}

	// Convert set to slice
	permissions := make([]Permission, 0, len(permissionSet))
	for perm := range permissionSet {
		permissions = append(permissions, perm)
	}

	return permissions
}

// GetUserRoles returns all roles for a user
func (pc *PermissionCheckerImpl) GetUserRoles(user *User) []Role {
	if user == nil {
		return []Role{}
	}

	var roles []Role
	for _, roleName := range user.Roles {
		if role, exists := pc.roles[roleName]; exists {
			roles = append(roles, role)
		}
	}

	return roles
}

// AddRole adds a custom role
func (pc *PermissionCheckerImpl) AddRole(role Role) error {
	if role.Name == "" {
		return fmt.Errorf("role name cannot be empty")
	}

	pc.roles[role.Name] = role
	return nil
}

// RemoveRole removes a role
func (pc *PermissionCheckerImpl) RemoveRole(roleName string) error {
	if _, exists := DefaultRoles[roleName]; exists {
		return fmt.Errorf("cannot remove default role: %s", roleName)
	}

	if _, exists := pc.roles[roleName]; !exists {
		return fmt.Errorf("role does not exist: %s", roleName)
	}

	delete(pc.roles, roleName)
	return nil
}

// GetRole returns a role by name
func (pc *PermissionCheckerImpl) GetRole(roleName string) (Role, bool) {
	role, exists := pc.roles[roleName]
	return role, exists
}

// ListRoles returns all available roles
func (pc *PermissionCheckerImpl) ListRoles() map[string]Role {
	result := make(map[string]Role)
	for name, role := range pc.roles {
		result[name] = role
	}
	return result
}

// Helper functions for common permission checks

// CanDeleteLock checks if user can delete locks
func (pc *PermissionCheckerImpl) CanDeleteLock(user *User) bool {
	return pc.HasPermission(user, PermissionLockDelete)
}

// CanForceDeleteLock checks if user can force delete locks
func (pc *PermissionCheckerImpl) CanForceDeleteLock(user *User) bool {
	return pc.HasPermission(user, PermissionLockForce)
}

// CanApplyPlan checks if user can apply plans
func (pc *PermissionCheckerImpl) CanApplyPlan(user *User) bool {
	return pc.HasPermission(user, PermissionPlanApply)
}

// CanDeletePlan checks if user can delete plans
func (pc *PermissionCheckerImpl) CanDeletePlan(user *User) bool {
	return pc.HasPermission(user, PermissionPlanDelete)
}

// CanWritePolicy checks if user can write policies
func (pc *PermissionCheckerImpl) CanWritePolicy(user *User) bool {
	return pc.HasPermission(user, PermissionPolicyWrite)
}

// CanManageUsers checks if user can manage other users
func (pc *PermissionCheckerImpl) CanManageUsers(user *User) bool {
	return pc.HasAnyPermission(user, []Permission{
		PermissionUserWrite,
		PermissionUserDelete,
	})
}

// IsAdmin checks if user has admin privileges
func (pc *PermissionCheckerImpl) IsAdmin(user *User) bool {
	return pc.HasAnyPermission(user, []Permission{
		PermissionAdminRead,
		PermissionAdminWrite,
		PermissionAdminDelete,
	})
}

// IsSuperAdmin checks if user has super admin privileges
func (pc *PermissionCheckerImpl) IsSuperAdmin(user *User) bool {
	return pc.HasPermission(user, PermissionUserDelete)
}

// ParsePermission parses a permission string into a Permission type
func ParsePermission(permissionStr string) (Permission, error) {
	permission := Permission(permissionStr)
	
	// Validate the permission
	validPermissions := []Permission{
		PermissionRepoRead, PermissionRepoWrite, PermissionRepoDelete,
		PermissionLockRead, PermissionLockCreate, PermissionLockDelete, PermissionLockForce,
		PermissionPlanRead, PermissionPlanCreate, PermissionPlanApply, PermissionPlanDelete,
		PermissionPolicyRead, PermissionPolicyWrite,
		PermissionAdminRead, PermissionAdminWrite, PermissionAdminDelete,
		PermissionUserRead, PermissionUserWrite, PermissionUserDelete,
	}

	for _, valid := range validPermissions {
		if permission == valid {
			return permission, nil
		}
	}

	return "", fmt.Errorf("invalid permission: %s", permissionStr)
}

// ParsePermissions parses a slice of permission strings
func ParsePermissions(permissionStrs []string) ([]Permission, error) {
	var permissions []Permission
	for _, permStr := range permissionStrs {
		perm, err := ParsePermission(permStr)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, perm)
	}
	return permissions, nil
}

// PermissionToString converts a permission to string
func PermissionToString(permission Permission) string {
	return string(permission)
}

// GetPermissionDescription returns a human-readable description of a permission
func GetPermissionDescription(permission Permission) string {
	descriptions := map[Permission]string{
		PermissionRepoRead:    "Read repository information",
		PermissionRepoWrite:   "Write to repositories",
		PermissionRepoDelete:  "Delete repositories",
		PermissionLockRead:    "Read locks",
		PermissionLockCreate:  "Create locks",
		PermissionLockDelete:  "Delete locks",
		PermissionLockForce:   "Force delete locks",
		PermissionPlanRead:    "Read plans",
		PermissionPlanCreate:  "Create plans",
		PermissionPlanApply:   "Apply plans",
		PermissionPlanDelete:  "Delete plans",
		PermissionPolicyRead:  "Read policies",
		PermissionPolicyWrite: "Write policies",
		PermissionAdminRead:   "Read admin information",
		PermissionAdminWrite:  "Write admin settings",
		PermissionAdminDelete: "Delete admin resources",
		PermissionUserRead:    "Read user information",
		PermissionUserWrite:   "Write user information",
		PermissionUserDelete:  "Delete users",
	}

	if desc, exists := descriptions[permission]; exists {
		return desc
	}
	return "Unknown permission"
}

// GetPermissionCategory returns the category of a permission
func GetPermissionCategory(permission Permission) string {
	parts := strings.Split(string(permission), ":")
	if len(parts) >= 2 {
		return parts[0]
	}
	return "unknown"
} 