package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/logging"
	"github.com/urfave/negroni/v3"
)

// MockManager implements the Manager interface for testing
type MockManager struct {
	loginRequired bool
	user          *User
	userError     error
	redirectError error
}

func (m *MockManager) AuthenticateUser(ctx context.Context, user *User) (*Session, error) {
	return &Session{
		ID:        "test-session",
		UserID:    user.ID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}, nil
}

func (m *MockManager) ValidateSession(ctx context.Context, sessionID string) (*User, error) {
	if m.user != nil {
		return m.user, nil
	}
	return nil, m.userError
}

func (m *MockManager) InvalidateSession(ctx context.Context, sessionID string) error {
	return nil
}

func (m *MockManager) GetUserFromRequest(r *http.Request) (*User, error) {
	return m.user, m.userError
}

func (m *MockManager) LoginRequired(r *http.Request) bool {
	return m.loginRequired
}

func (m *MockManager) RedirectToLogin(w http.ResponseWriter, r *http.Request) error {
	return m.redirectError
}

func (m *MockManager) GetProvider(providerID string) (Provider, error) {
	return nil, nil
}

func (m *MockManager) GetEnabledProviders() []Provider {
	return nil
}

func (m *MockManager) GetPermissionChecker() PermissionChecker {
	return NewPermissionChecker(nil)
}

func TestNewAuthMiddleware(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	mockManager := &MockManager{}
	
	middleware := NewAuthMiddleware(mockManager, logger)
	
	if middleware == nil {
		t.Fatal("Middleware should not be nil")
	}
	
	if middleware.authManager != mockManager {
		t.Error("Auth manager should be set correctly")
	}
	
	if middleware.logger != logger {
		t.Error("Logger should be set correctly")
	}
}

func TestAuthMiddleware_ServeHTTP_NoAuthRequired(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	mockManager := &MockManager{
		loginRequired: false,
	}
	
	middleware := NewAuthMiddleware(mockManager, logger)
	
	req := httptest.NewRequest("GET", "/public", nil)
	w := httptest.NewRecorder()
	
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})
	
	middleware.ServeHTTP(w, req, next)
	
	if !called {
		t.Error("Next handler should be called when auth is not required")
	}
}

func TestAuthMiddleware_ServeHTTP_AuthRequired_ValidUser(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	testUser := &User{
		ID:    "test-user",
		Email: "test@example.com",
		Name:  "Test User",
	}
	
	mockManager := &MockManager{
		loginRequired: true,
		user:          testUser,
	}
	
	middleware := NewAuthMiddleware(mockManager, logger)
	
	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		// Check that user is in context
		user, ok := UserFromContext(r.Context())
		if !ok {
			t.Error("User should be in request context")
		}
		if user != testUser {
			t.Error("User in context should match test user")
		}
	})
	
	middleware.ServeHTTP(w, req, next)
	
	if !called {
		t.Error("Next handler should be called when user is authenticated")
	}
}

func TestAuthMiddleware_ServeHTTP_AuthRequired_NoUser(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	mockManager := &MockManager{
		loginRequired: true,
		userError:     http.ErrNoCookie,
	}
	
	middleware := NewAuthMiddleware(mockManager, logger)
	
	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})
	
	middleware.ServeHTTP(w, req, next)
	
	if called {
		t.Error("Next handler should not be called when user is not authenticated")
	}
}

func TestAuthMiddleware_ServeHTTP_AuthRequired_RedirectError(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	mockManager := &MockManager{
		loginRequired: true,
		userError:     http.ErrNoCookie,
		redirectError: http.ErrServerClosed,
	}
	
	middleware := NewAuthMiddleware(mockManager, logger)
	
	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})
	
	middleware.ServeHTTP(w, req, next)
	
	if called {
		t.Error("Next handler should not be called when redirect fails")
	}
	
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestContextWithUser(t *testing.T) {
	ctx := context.Background()
	testUser := &User{
		ID:    "test-user",
		Email: "test@example.com",
	}
	
	newCtx := contextWithUser(ctx, testUser)
	
	user, ok := UserFromContext(newCtx)
	if !ok {
		t.Fatal("User should be retrievable from context")
	}
	
	if user != testUser {
		t.Error("User from context should match test user")
	}
}

func TestUserFromContext(t *testing.T) {
	ctx := context.Background()
	testUser := &User{
		ID:    "test-user",
		Email: "test@example.com",
	}
	
	// Test with no user in context
	user, ok := UserFromContext(ctx)
	if ok {
		t.Error("Should return false when no user in context")
	}
	if user != nil {
		t.Error("Should return nil user when no user in context")
	}
	
	// Test with user in context
	ctxWithUser := contextWithUser(ctx, testUser)
	user, ok = UserFromContext(ctxWithUser)
	if !ok {
		t.Error("Should return true when user in context")
	}
	if user != testUser {
		t.Error("Should return correct user from context")
	}
}

func TestNewLegacyAuthMiddleware(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	
	middleware := NewLegacyAuthMiddleware(true, "testuser", "testpass", logger)
	
	if middleware == nil {
		t.Fatal("Middleware should not be nil")
	}
	
	if !middleware.WebAuthentication {
		t.Error("WebAuthentication should be true")
	}
	
	if middleware.WebUsername != "testuser" {
		t.Errorf("WebUsername = %s, want testuser", middleware.WebUsername)
	}
	
	if middleware.WebPassword != "testpass" {
		t.Errorf("WebPassword = %s, want testpass", middleware.WebPassword)
	}
	
	if middleware.logger != logger {
		t.Error("Logger should be set correctly")
	}
}

func TestLegacyAuthMiddleware_ServeHTTP_NoAuthRequired(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	middleware := NewLegacyAuthMiddleware(false, "testuser", "testpass", logger)
	
	req := httptest.NewRequest("GET", "/public", nil)
	
	// Create a proper negroni.ResponseWriter for status logging
	recorder := httptest.NewRecorder()
	responseWriter := negroni.NewResponseWriter(recorder)
	
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})
	
	middleware.ServeHTTP(responseWriter, req, next)
	
	if !called {
		t.Error("Next handler should be called when auth is not required")
	}
}

func TestLegacyAuthMiddleware_ServeHTTP_ExemptPaths(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	middleware := NewLegacyAuthMiddleware(true, "testuser", "testpass", logger)
	
	exemptPaths := []string{"/events", "/healthz", "/status", "/api/test"}
	
	for _, path := range exemptPaths {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest("GET", path, nil)
			
			// Create a proper negroni.ResponseWriter for status logging
			recorder := httptest.NewRecorder()
			responseWriter := negroni.NewResponseWriter(recorder)
			
			called := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
			})
			
			middleware.ServeHTTP(responseWriter, req, next)
			
			if !called {
				t.Errorf("Next handler should be called for exempt path: %s", path)
			}
		})
	}
}

func TestLegacyAuthMiddleware_ServeHTTP_ValidCredentials(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	middleware := NewLegacyAuthMiddleware(true, "testuser", "testpass", logger)
	
	req := httptest.NewRequest("GET", "/protected", nil)
	req.SetBasicAuth("testuser", "testpass")
	
	// Create a proper negroni.ResponseWriter for status logging
	recorder := httptest.NewRecorder()
	responseWriter := negroni.NewResponseWriter(recorder)
	
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})
	
	middleware.ServeHTTP(responseWriter, req, next)
	
	if !called {
		t.Error("Next handler should be called with valid credentials")
	}
}

func TestLegacyAuthMiddleware_ServeHTTP_InvalidCredentials(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	middleware := NewLegacyAuthMiddleware(true, "testuser", "testpass", logger)
	
	req := httptest.NewRequest("GET", "/protected", nil)
	req.SetBasicAuth("testuser", "wrongpass")
	
	// Create a proper negroni.ResponseWriter for status logging
	recorder := httptest.NewRecorder()
	responseWriter := negroni.NewResponseWriter(recorder)
	
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})
	
	middleware.ServeHTTP(responseWriter, req, next)
	
	if called {
		t.Error("Next handler should not be called with invalid credentials")
	}
	
	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
	
	authHeader := recorder.Header().Get("WWW-Authenticate")
	if authHeader == "" {
		t.Error("WWW-Authenticate header should be set")
	}
}

func TestLegacyAuthMiddleware_ServeHTTP_NoAuthHeader(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	middleware := NewLegacyAuthMiddleware(true, "testuser", "testpass", logger)
	
	req := httptest.NewRequest("GET", "/protected", nil)
	
	// Create a proper negroni.ResponseWriter for status logging
	recorder := httptest.NewRecorder()
	responseWriter := negroni.NewResponseWriter(recorder)
	
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})
	
	middleware.ServeHTTP(responseWriter, req, next)
	
	if called {
		t.Error("Next handler should not be called without auth header")
	}
	
	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestRequirePermission(t *testing.T) {
	mockManager := &MockManager{}
	
	// Create middleware that adds auth manager to context
	authMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "auth_manager", mockManager)
			next(w, r.WithContext(ctx))
		}
	}
	
	permissionMiddleware := RequirePermission(PermissionRepoRead)
	
	// Test with user having permission
	testUser := &User{
		ID:         "test-user",
		Email:      "test@example.com",
		Roles:      []string{"user"}, // user role has repo:read permission
		Permissions: []Permission{},
	}
	
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := contextWithUser(req.Context(), testUser)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})
	
	// Apply both middlewares
	finalHandler := authMiddleware(permissionMiddleware(handler))
	finalHandler.ServeHTTP(w, req)
	
	if !called {
		t.Error("Handler should be called when user has permission")
	}
}

func TestRequirePermission_NoUser(t *testing.T) {
	permissionMiddleware := RequirePermission(PermissionRepoRead)
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})
	
	permissionMiddleware(handler).ServeHTTP(w, req)
	
	if called {
		t.Error("Handler should not be called when no user in context")
	}
	
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestRequirePermission_InsufficientPermissions(t *testing.T) {
	mockManager := &MockManager{}
	
	// Create middleware that adds auth manager to context
	authMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "auth_manager", mockManager)
			next(w, r.WithContext(ctx))
		}
	}
	
	permissionMiddleware := RequirePermission(PermissionUserDelete) // user role doesn't have this
	
	// Test with user not having permission
	testUser := &User{
		ID:         "test-user",
		Email:      "test@example.com",
		Roles:      []string{"user"}, // user role doesn't have user:delete
		Permissions: []Permission{},
	}
	
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := contextWithUser(req.Context(), testUser)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})
	
	// Apply both middlewares
	finalHandler := authMiddleware(permissionMiddleware(handler))
	finalHandler.ServeHTTP(w, req)
	
	if called {
		t.Error("Handler should not be called when user lacks permission")
	}
	
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestRequireAnyPermission(t *testing.T) {
	mockManager := &MockManager{}
	
	// Create middleware that adds auth manager to context
	authMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "auth_manager", mockManager)
			next(w, r.WithContext(ctx))
		}
	}
	
	permissions := []Permission{PermissionRepoRead, PermissionUserDelete}
	permissionMiddleware := RequireAnyPermission(permissions)
	
	// Test with user having one of the permissions
	testUser := &User{
		ID:         "test-user",
		Email:      "test@example.com",
		Roles:      []string{"user"}, // user role has repo:read but not user:delete
		Permissions: []Permission{},
	}
	
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := contextWithUser(req.Context(), testUser)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})
	
	// Apply both middlewares
	finalHandler := authMiddleware(permissionMiddleware(handler))
	finalHandler.ServeHTTP(w, req)
	
	if !called {
		t.Error("Handler should be called when user has any of the permissions")
	}
}

func TestRequireAnyPermission_NoPermissions(t *testing.T) {
	mockManager := &MockManager{}
	
	// Create middleware that adds auth manager to context
	authMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "auth_manager", mockManager)
			next(w, r.WithContext(ctx))
		}
	}
	
	permissions := []Permission{PermissionUserDelete, PermissionUserWrite}
	permissionMiddleware := RequireAnyPermission(permissions)
	
	// Test with user having none of the permissions
	testUser := &User{
		ID:         "test-user",
		Email:      "test@example.com",
		Roles:      []string{"user"}, // user role doesn't have these permissions
		Permissions: []Permission{},
	}
	
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := contextWithUser(req.Context(), testUser)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})
	
	// Apply both middlewares
	finalHandler := authMiddleware(permissionMiddleware(handler))
	finalHandler.ServeHTTP(w, req)
	
	if called {
		t.Error("Handler should not be called when user has none of the permissions")
	}
	
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestRequireAllPermissions(t *testing.T) {
	mockManager := &MockManager{}
	
	// Create middleware that adds auth manager to context
	authMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "auth_manager", mockManager)
			next(w, r.WithContext(ctx))
		}
	}
	
	permissions := []Permission{PermissionRepoRead, PermissionLockRead}
	permissionMiddleware := RequireAllPermissions(permissions)
	
	// Test with user having all permissions
	testUser := &User{
		ID:         "test-user",
		Email:      "test@example.com",
		Roles:      []string{"user"}, // user role has both permissions
		Permissions: []Permission{},
	}
	
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := contextWithUser(req.Context(), testUser)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})
	
	// Apply both middlewares
	finalHandler := authMiddleware(permissionMiddleware(handler))
	finalHandler.ServeHTTP(w, req)
	
	if !called {
		t.Error("Handler should be called when user has all permissions")
	}
}

func TestRequireAllPermissions_MissingPermission(t *testing.T) {
	mockManager := &MockManager{}
	
	// Create middleware that adds auth manager to context
	authMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "auth_manager", mockManager)
			next(w, r.WithContext(ctx))
		}
	}
	
	permissions := []Permission{PermissionRepoRead, PermissionUserDelete}
	permissionMiddleware := RequireAllPermissions(permissions)
	
	// Test with user missing one permission
	testUser := &User{
		ID:         "test-user",
		Email:      "test@example.com",
		Roles:      []string{"user"}, // user role has repo:read but not user:delete
		Permissions: []Permission{},
	}
	
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := contextWithUser(req.Context(), testUser)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})
	
	// Apply both middlewares
	finalHandler := authMiddleware(permissionMiddleware(handler))
	finalHandler.ServeHTTP(w, req)
	
	if called {
		t.Error("Handler should not be called when user lacks any permission")
	}
	
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestRequireAdmin(t *testing.T) {
	mockManager := &MockManager{}
	
	// Create middleware that adds auth manager to context
	authMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "auth_manager", mockManager)
			next(w, r.WithContext(ctx))
		}
	}
	
	adminMiddleware := RequireAdmin()
	
	// Test with admin user
	testUser := &User{
		ID:         "admin-user",
		Email:      "admin@example.com",
		Roles:      []string{"admin"}, // admin role has admin permissions
		Permissions: []Permission{},
	}
	
	req := httptest.NewRequest("GET", "/admin", nil)
	ctx := contextWithUser(req.Context(), testUser)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})
	
	// Apply both middlewares
	finalHandler := authMiddleware(adminMiddleware(handler))
	finalHandler.ServeHTTP(w, req)
	
	if !called {
		t.Error("Handler should be called when user is admin")
	}
}

func TestRequireAdmin_NonAdmin(t *testing.T) {
	mockManager := &MockManager{}
	
	// Create middleware that adds auth manager to context
	authMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "auth_manager", mockManager)
			next(w, r.WithContext(ctx))
		}
	}
	
	adminMiddleware := RequireAdmin()
	
	// Test with non-admin user
	testUser := &User{
		ID:         "user",
		Email:      "user@example.com",
		Roles:      []string{"user"}, // user role doesn't have admin permissions
		Permissions: []Permission{},
	}
	
	req := httptest.NewRequest("GET", "/admin", nil)
	ctx := contextWithUser(req.Context(), testUser)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})
	
	// Apply both middlewares
	finalHandler := authMiddleware(adminMiddleware(handler))
	finalHandler.ServeHTTP(w, req)
	
	if called {
		t.Error("Handler should not be called when user is not admin")
	}
	
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestSpecificPermissionMiddlewares(t *testing.T) {
	mockManager := &MockManager{}
	
	// Create middleware that adds auth manager to context
	authMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "auth_manager", mockManager)
			next(w, r.WithContext(ctx))
		}
	}
	
	// Test specific permission middlewares
	middlewares := []struct {
		name     string
		middleware func(http.HandlerFunc) http.HandlerFunc
		requiredPermission Permission
	}{
		{"RequireLockDeletePermission", RequireLockDeletePermission(), PermissionLockDelete},
		{"RequireLockForcePermission", RequireLockForcePermission(), PermissionLockForce},
		{"RequirePlanApplyPermission", RequirePlanApplyPermission(), PermissionPlanApply},
		{"RequirePlanDeletePermission", RequirePlanDeletePermission(), PermissionPlanDelete},
		{"RequirePolicyWritePermission", RequirePolicyWritePermission(), PermissionPolicyWrite},
		{"RequireUserManagementPermission", RequireUserManagementPermission(), PermissionUserWrite},
	}
	
	for _, tt := range middlewares {
		t.Run(tt.name, func(t *testing.T) {
			// Test with user having the required permission
			testUser := &User{
				ID:         "test-user",
				Email:      "test@example.com",
				Roles:      []string{"admin"}, // admin role has all permissions
				Permissions: []Permission{},
			}
			
			req := httptest.NewRequest("GET", "/test", nil)
			ctx := contextWithUser(req.Context(), testUser)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()
			
			called := false
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
			})
			
			// Apply both middlewares
			finalHandler := authMiddleware(tt.middleware(handler))
			finalHandler.ServeHTTP(w, req)
			
			if !called {
				t.Errorf("Handler should be called when user has %s permission", tt.requiredPermission)
			}
		})
	}
} 