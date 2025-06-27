package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/runatlantis/atlantis/server/logging"
	"github.com/urfave/negroni/v3"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	userContextKey contextKey = "user"
)

// AuthMiddleware handles authentication for HTTP requests
type AuthMiddleware struct {
	authManager Manager
	logger      logging.SimpleLogging
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(authManager Manager, logger logging.SimpleLogging) *AuthMiddleware {
	return &AuthMiddleware{
		authManager: authManager,
		logger:      logger,
	}
}

// ServeHTTP implements the middleware function
func (m *AuthMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	m.logger.Debug("%s %s – from %s", r.Method, r.URL.RequestURI(), r.RemoteAddr)

	// Check if authentication is required for this request
	if !m.authManager.LoginRequired(r) {
		next(rw, r)
		return
	}

	// Try to get user from request
	user, err := m.authManager.GetUserFromRequest(r)
	if err != nil {
		m.logger.Debug("[AUTH] No valid authentication found for: %s", r.URL.RequestURI())
		
		// Redirect to login page
		if err := m.authManager.RedirectToLogin(rw, r); err != nil {
			m.logger.Err("Failed to redirect to login: %s", err)
			http.Error(rw, "Authentication required", http.StatusUnauthorized)
		}
		return
	}

	// User is authenticated, add user info to request context
	ctx := r.Context()
	ctx = WithUser(ctx, user)
	r = r.WithContext(ctx)

	m.logger.Debug("[AUTH] User %s authenticated for: %s", user.Email, r.URL.RequestURI())
	next(rw, r)
}

// WithUser adds a user to the request context
func WithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// GetUserFromContext retrieves a user from the request context
func GetUserFromContext(ctx context.Context) (*User, bool) {
	user, ok := ctx.Value(userContextKey).(*User)
	return user, ok
}

// LegacyAuthMiddleware provides backward compatibility with the old basic auth
type LegacyAuthMiddleware struct {
	WebAuthentication bool
	WebUsername       string
	WebPassword       string
	logger            logging.SimpleLogging
}

// NewLegacyAuthMiddleware creates a legacy authentication middleware
func NewLegacyAuthMiddleware(webAuth bool, username, password string, logger logging.SimpleLogging) *LegacyAuthMiddleware {
	return &LegacyAuthMiddleware{
		WebAuthentication: webAuth,
		WebUsername:       username,
		WebPassword:       password,
		logger:            logger,
	}
}

// ServeHTTP implements the legacy middleware function
func (l *LegacyAuthMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	l.logger.Debug("%s %s – from %s", r.Method, r.URL.RequestURI(), r.RemoteAddr)
	
	allowed := false
	if !l.WebAuthentication ||
		r.URL.Path == "/events" ||
		r.URL.Path == "/healthz" ||
		r.URL.Path == "/status" ||
		strings.HasPrefix(r.URL.Path, "/api/") {
		allowed = true
	} else {
		user, pass, ok := r.BasicAuth()
		if ok {
			r.SetBasicAuth(user, pass)
			if user == l.WebUsername && pass == l.WebPassword {
				l.logger.Debug("[VALID] log in: >> url: %s", r.URL.RequestURI())
				allowed = true
			} else {
				allowed = false
				l.logger.Info("[INVALID] log in attempt: >> url: %s", r.URL.RequestURI())
			}
		}
	}
	
	if !allowed {
		rw.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(rw, "Unauthorized", http.StatusUnauthorized)
	} else {
		next(rw, r)
	}
	
	l.logger.Debug("%s %s – respond HTTP %d", r.Method, r.URL.RequestURI(), rw.(negroni.ResponseWriter).Status())
}

// RequirePermission middleware requires a specific permission
func RequirePermission(permission Permission) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			user, ok := GetUserFromContext(r.Context())
			if !ok {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			// Get permission checker from auth manager
			authManager, ok := r.Context().Value("auth_manager").(Manager)
			if !ok {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if authManagerWithChecker, ok := authManager.(*AuthManager); ok {
				permissionChecker := authManagerWithChecker.GetPermissionChecker()
				if !permissionChecker.HasPermission(user, permission) {
					http.Error(w, "Insufficient permissions", http.StatusForbidden)
					return
				}
			}

			next(w, r)
		}
	}
}

// RequireAnyPermission middleware requires any of the specified permissions
func RequireAnyPermission(permissions []Permission) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			user, ok := GetUserFromContext(r.Context())
			if !ok {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			// Get permission checker from auth manager
			authManager, ok := r.Context().Value("auth_manager").(Manager)
			if !ok {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if authManagerWithChecker, ok := authManager.(*AuthManager); ok {
				permissionChecker := authManagerWithChecker.GetPermissionChecker()
				if !permissionChecker.HasAnyPermission(user, permissions) {
					http.Error(w, "Insufficient permissions", http.StatusForbidden)
					return
				}
			}

			next(w, r)
		}
	}
}

// RequireAllPermissions middleware requires all of the specified permissions
func RequireAllPermissions(permissions []Permission) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			user, ok := GetUserFromContext(r.Context())
			if !ok {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			// Get permission checker from auth manager
			authManager, ok := r.Context().Value("auth_manager").(Manager)
			if !ok {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if authManagerWithChecker, ok := authManager.(*AuthManager); ok {
				permissionChecker := authManagerWithChecker.GetPermissionChecker()
				if !permissionChecker.HasAllPermissions(user, permissions) {
					http.Error(w, "Insufficient permissions", http.StatusForbidden)
					return
				}
			}

			next(w, r)
		}
	}
}

// RequireAdmin middleware requires admin privileges
func RequireAdmin() func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			user, ok := GetUserFromContext(r.Context())
			if !ok {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			// Get permission checker from auth manager
			authManager, ok := r.Context().Value("auth_manager").(Manager)
			if !ok {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if authManagerWithChecker, ok := authManager.(*AuthManager); ok {
				permissionChecker := authManagerWithChecker.GetPermissionChecker()
				if !permissionChecker.IsAdmin(user) {
					http.Error(w, "Admin privileges required", http.StatusForbidden)
					return
				}
			}

			next(w, r)
		}
	}
}

// Helper middleware functions for common permission checks

// RequireLockDeletePermission middleware for lock deletion
func RequireLockDeletePermission() func(http.HandlerFunc) http.HandlerFunc {
	return RequirePermission(PermissionLockDelete)
}

// RequireLockForcePermission middleware for force lock deletion
func RequireLockForcePermission() func(http.HandlerFunc) http.HandlerFunc {
	return RequirePermission(PermissionLockForce)
}

// RequirePlanApplyPermission middleware for plan application
func RequirePlanApplyPermission() func(http.HandlerFunc) http.HandlerFunc {
	return RequirePermission(PermissionPlanApply)
}

// RequirePlanDeletePermission middleware for plan deletion
func RequirePlanDeletePermission() func(http.HandlerFunc) http.HandlerFunc {
	return RequirePermission(PermissionPlanDelete)
}

// RequirePolicyWritePermission middleware for policy writing
func RequirePolicyWritePermission() func(http.HandlerFunc) http.HandlerFunc {
	return RequirePermission(PermissionPolicyWrite)
}

// RequireUserManagementPermission middleware for user management
func RequireUserManagementPermission() func(http.HandlerFunc) http.HandlerFunc {
	return RequireAnyPermission([]Permission{
		PermissionUserWrite,
		PermissionUserDelete,
	})
} 