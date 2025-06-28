# SSO Authentication for Atlantis

Atlantis now supports Single Sign-On (SSO) authentication with multiple identity providers, providing a flexible and secure authentication layer for the web UI.

## Features

-  **Multiple SSO Providers**: Support for Google OAuth2, Okta OIDC, Azure AD, and custom OAuth2/OIDC providers
-  **Session Management**: Secure session-based authentication with configurable session duration
-  **Role-Based Access Control (RBAC)**: User roles and permissions based on groups and attributes
-  **Backward Compatibility**: Maintains support for existing basic authentication
-  **Flexible Configuration**: Environment variables and JSON configuration files

## Supported Providers

### Google OAuth2

-  Uses Google's OAuth2 implementation
-  Supports OpenID Connect
-  Automatic user info retrieval

### Okta OIDC

-  Full OpenID Connect support
-  Group membership integration
-  Customizable scopes

### Azure AD (Entra ID)

-  Microsoft Entra ID (formerly Azure AD) integration
-  Enterprise SSO support
-  Group-based role assignment

### Auth0

-  Auth0 as an OIDC provider
-  Support for multiple social connections
-  Custom rules and claims
-  Enterprise SSO capabilities

### Custom OAuth2/OIDC

-  Support for any OAuth2 or OpenID Connect provider
-  Configurable endpoints and scopes
-  Flexible attribute mapping

### Basic Authentication

-  Legacy basic HTTP authentication
-  Backward compatibility
-  Simple username/password validation

## Configuration

### Environment Variables

#### General Settings

```bash
# Session configuration
export ATLANTIS_SESSION_SECRET="your-secret-key"
export ATLANTIS_SESSION_DURATION="24h"
export ATLANTIS_SESSION_COOKIE_NAME="atlantis_session"
export ATLANTIS_SECURE_COOKIES="true"
export ATLANTIS_CSRF_SECRET="your-csrf-secret"

# Basic auth (legacy)
export ATLANTIS_ENABLE_BASIC_AUTH="true"
export ATLANTIS_BASIC_AUTH_USER="atlantis"
export ATLANTIS_BASIC_AUTH_PASS="atlantis"
```

#### Google OAuth2

```bash
export ATLANTIS_GOOGLE_CLIENT_ID="your-google-client-id"
export ATLANTIS_GOOGLE_CLIENT_SECRET="your-google-client-secret"
export ATLANTIS_GOOGLE_REDIRECT_URL="https://your-atlantis-domain/auth/callback?provider=google"
```

#### Okta OIDC

```bash
export ATLANTIS_OKTA_CLIENT_ID="your-okta-client-id"
export ATLANTIS_OKTA_CLIENT_SECRET="your-okta-client-secret"
export ATLANTIS_OKTA_ISSUER_URL="https://your-org.okta.com"
export ATLANTIS_OKTA_REDIRECT_URL="https://your-atlantis-domain/auth/callback?provider=okta"
```

#### Azure AD

```bash
export ATLANTIS_AZURE_CLIENT_ID="your-azure-client-id"
export ATLANTIS_AZURE_CLIENT_SECRET="your-azure-client-secret"
export ATLANTIS_AZURE_TENANT_ID="your-tenant-id"
export ATLANTIS_AZURE_REDIRECT_URL="https://your-atlantis-domain/auth/callback?provider=azure"
```

#### Auth0

```bash
export ATLANTIS_AUTH0_CLIENT_ID="your-auth0-client-id"
export ATLANTIS_AUTH0_CLIENT_SECRET="your-auth0-client-secret"
export ATLANTIS_AUTH0_DOMAIN="your-tenant.auth0.com"
export ATLANTIS_AUTH0_REDIRECT_URL="https://your-atlantis-domain/auth/callback?provider=auth0"
```

### JSON Configuration File

Create a `auth-config.json` file:

```json
{
   "session_secret": "your-secret-key",
   "session_duration": "24h",
   "session_cookie_name": "atlantis_session",
   "secure_cookies": true,
   "csrf_secret": "your-csrf-secret",
   "enable_basic_auth": false,
   "basic_auth_user": "atlantis",
   "basic_auth_pass": "atlantis",
   "default_roles": ["user"],
   "admin_groups": ["admin", "administrators"],
   "admin_emails": ["admin@example.com"],
   "providers": [
      {
         "type": "oauth2",
         "id": "google",
         "name": "Google",
         "client_id": "your-google-client-id",
         "client_secret": "your-google-client-secret",
         "redirect_url": "https://your-atlantis-domain/auth/callback?provider=google",
         "scopes": ["openid", "email", "profile"],
         "enabled": true,
         "default_roles": ["user"],
         "allowed_groups": ["admin", "devops"],
         "allowed_emails": ["admin@example.com"]
      },
      {
         "type": "oidc",
         "id": "okta",
         "name": "Okta",
         "client_id": "your-okta-client-id",
         "client_secret": "your-okta-client-secret",
         "redirect_url": "https://your-atlantis-domain/auth/callback?provider=okta",
         "issuer_url": "https://your-org.okta.com",
         "scopes": ["openid", "email", "profile", "groups"],
         "enabled": true,
         "default_roles": ["user"],
         "allowed_groups": ["admin", "devops"],
         "allowed_emails": ["admin@example.com"]
      },
      {
         "type": "oidc",
         "id": "azure",
         "name": "Azure AD",
         "client_id": "your-azure-client-id",
         "client_secret": "your-azure-client-secret",
         "redirect_url": "https://your-atlantis-domain/auth/callback?provider=azure",
         "issuer_url": "https://login.microsoftonline.com/your-tenant-id",
         "scopes": ["openid", "email", "profile", "User.Read"],
         "enabled": true,
         "default_roles": ["user"],
         "allowed_groups": ["admin", "devops"],
         "allowed_emails": ["admin@example.com"]
      },
      {
         "type": "oidc",
         "id": "auth0",
         "name": "Auth0",
         "client_id": "your-auth0-client-id",
         "client_secret": "your-auth0-client-secret",
         "redirect_url": "https://your-atlantis-domain/auth/callback?provider=auth0",
         "issuer_url": "https://your-tenant.auth0.com",
         "scopes": ["openid", "email", "profile"],
         "enabled": true,
         "default_roles": ["user"],
         "allowed_groups": ["admin", "devops"],
         "allowed_emails": ["admin@example.com"]
      }
   ]
}
```

## Setup Instructions

### 1. Google OAuth2 Setup

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the Google+ API
4. Go to "Credentials" and create an OAuth 2.0 Client ID
5. Set the authorized redirect URI to: `https://your-atlantis-domain/auth/callback?provider=google`
6. Copy the Client ID and Client Secret to your environment variables

### 2. Okta OIDC Setup

1. Log in to your Okta Admin Console
2. Go to "Applications" → "Applications"
3. Click "Create App Integration"
4. Choose "OIDC - OpenID Connect" and "Web application"
5. Configure the app:
   -  Base URIs: `https://your-atlantis-domain`
   -  Login redirect URIs: `https://your-atlantis-domain/auth/callback?provider=okta`
   -  Grant type: Authorization Code
6. Copy the Client ID and Client Secret
7. Note your Okta domain (e.g., `https://your-org.okta.com`)

### 3. Azure AD Setup

1. Go to the [Azure Portal](https://portal.azure.com/)
2. Navigate to "Azure Active Directory" → "App registrations"
3. Click "New registration"
4. Configure the app:
   -  Name: Atlantis
   -  Redirect URI: Web → `https://your-atlantis-domain/auth/callback?provider=azure`
5. Go to "Certificates & secrets" and create a new client secret
6. Copy the Application (client) ID, Directory (tenant) ID, and client secret

### 4. Auth0 Setup

1. Log in to your [Auth0 Dashboard](https://manage.auth0.com/)
2. Go to "Applications" → "Applications"
3. Click "Create Application"
4. Choose "Regular Web Application"
5. Configure the application:
   -  **Allowed Callback URLs**: `https://your-atlantis-domain/auth/callback?provider=auth0`
   -  **Allowed Logout URLs**: `https://your-atlantis-domain`
   -  **Allowed Web Origins**: `https://your-atlantis-domain`
6. Go to "Settings" tab and copy the Client ID and Client Secret
7. Note your Domain (e.g., `your-tenant.auth0.com`)

### 5. Atlantis Configuration

1. Set the required environment variables for your chosen provider(s)
2. Ensure your Atlantis URL is accessible from the internet (for OAuth callbacks)
3. Start Atlantis with the new authentication system

## Usage

### Starting Atlantis with SSO

```bash
# Using environment variables
export ATLANTIS_GOOGLE_CLIENT_ID="your-client-id"
export ATLANTIS_GOOGLE_CLIENT_SECRET="your-client-secret"
export ATLANTIS_GOOGLE_REDIRECT_URL="https://your-domain/auth/callback?provider=google"
export ATLANTIS_SESSION_SECRET="your-session-secret"

atlantis server --gh-user=your-user --gh-token=your-token --repo-allowlist=github.com/your-org
```

### Using Configuration File

```bash
# Create auth-config.json with your settings
atlantis server --gh-user=your-user --gh-token=your-token --repo-allowlist=github.com/your-org --auth-config=auth-config.json
```

## Security Considerations

### Session Security

-  Use strong, unique session secrets
-  Enable secure cookies in production
-  Set appropriate session duration
-  Use HTTPS in production

### Provider Security

-  Keep client secrets secure
-  Use environment variables for sensitive data
-  Regularly rotate secrets
-  Monitor authentication logs

### Network Security

-  Use HTTPS for all OAuth callbacks
-  Configure proper firewall rules
-  Monitor for suspicious authentication attempts

## Troubleshooting

### Common Issues

1. **Callback URL Mismatch**

   -  Ensure the redirect URL in your provider matches exactly
   -  Check for trailing slashes and protocol differences

2. **Session Issues**

   -  Verify session secret is set correctly
   -  Check cookie settings and domain configuration
   -  Ensure HTTPS is used in production

3. **Provider Configuration**
   -  Verify client ID and secret are correct
   -  Check that required scopes are configured
   -  Ensure the provider is enabled

### Debugging

Enable debug logging to troubleshoot authentication issues:

```bash
export ATLANTIS_LOG_LEVEL=debug
```

## Migration from Basic Auth

To migrate from basic authentication to SSO:

1. Configure your SSO provider(s)
2. Set environment variables or create configuration file
3. Test authentication with a small group of users
4. Gradually migrate users to SSO
5. Keep basic auth enabled during transition
6. Disable basic auth once migration is complete

## API Authentication

The SSO authentication system only affects the web UI. API endpoints continue to use their existing authentication mechanisms (API tokens, etc.).

## Support

For issues and questions:

-  Check the troubleshooting section above
-  Review provider-specific documentation
-  Open an issue on the Atlantis GitHub repository

## Role-Based Access Control (RBAC)

The SSO authentication system includes a comprehensive Role-Based Access Control (RBAC) system that allows fine-grained permission management.

### Permission System

The system defines granular permissions for different actions:

#### Repository Permissions

-  `repo:read` - Read repository information
-  `repo:write` - Write to repositories
-  `repo:delete` - Delete repositories

#### Lock Permissions

-  `lock:read` - Read locks
-  `lock:create` - Create locks
-  `lock:delete` - Delete locks
-  `lock:force` - Force delete locks (admin only)

#### Plan Permissions

-  `plan:read` - Read plans
-  `plan:create` - Create plans
-  `plan:apply` - Apply plans
-  `plan:delete` - Delete plans

#### Policy Permissions

-  `policy:read` - Read policies
-  `policy:write` - Write policies

#### Admin Permissions

-  `admin:read` - Read admin information
-  `admin:write` - Write admin settings
-  `admin:delete` - Delete admin resources

#### User Management Permissions

-  `user:read` - Read user information
-  `user:write` - Write user information
-  `user:delete` - Delete users

### Default Roles

The system comes with predefined roles:

#### User Role

-  Basic read access to repositories, locks, plans, and policies
-  Suitable for team members who need to view information

#### Developer Role

-  Read and write access to repositories
-  Can create locks and plans
-  Suitable for developers working on infrastructure

#### Admin Role

-  Full access to most resources
-  Can manage locks, plans, policies, and admin settings
-  Cannot delete users (super admin only)

#### Super Admin Role

-  Complete access to all resources
-  Can manage users and delete any resource
-  Highest privilege level

### Custom Roles

You can define custom roles with specific permissions:

```json
{
   "roles": {
      "lock-manager": {
         "name": "lock-manager",
         "permissions": [
            "lock:read",
            "lock:create",
            "lock:delete",
            "lock:force",
            "repo:read",
            "plan:read"
         ],
         "description": "Can manage locks and read repositories/plans"
      },
      "plan-approver": {
         "name": "plan-approver",
         "permissions": [
            "repo:read",
            "lock:read",
            "plan:read",
            "plan:create",
            "plan:apply",
            "plan:delete"
         ],
         "description": "Can approve and apply plans"
      }
   }
}
```

### Group-Based Role Assignment

Roles can be automatically assigned based on user groups from your SSO provider:

```json
{
   "admin_groups": ["admin", "administrators", "devops"],
   "admin_emails": ["admin@example.com", "devops@example.com"],
   "providers": [
      {
         "allowed_groups": ["admin", "devops", "engineers", "lock-managers"],
         "allowed_emails": ["admin@example.com", "devops@example.com"]
      }
   ]
}
```

### Permission Checking

Use the permission checker to verify user permissions:

```go
// Check if user can delete locks
if permissionChecker.CanDeleteLock(user) {
    // Allow lock deletion
}

// Check if user can force delete locks
if permissionChecker.CanForceDeleteLock(user) {
    // Allow force lock deletion
}

// Check if user can apply plans
if permissionChecker.CanApplyPlan(user) {
    // Allow plan application
}

// Check if user is admin
if permissionChecker.IsAdmin(user) {
    // Allow admin actions
}
```

### Middleware for Permission Protection

Use middleware to protect endpoints:

```go
// Require specific permission
router.HandleFunc("/locks/{id}", RequireLockDeletePermission(deleteLockHandler))

// Require any of multiple permissions
router.HandleFunc("/plans/{id}", RequireAnyPermission([]Permission{
    PermissionPlanApply,
    PermissionPlanDelete,
})(planHandler))

// Require admin privileges
router.HandleFunc("/admin", RequireAdmin(adminHandler))
```

### Common Permission Patterns

#### Lock Management

-  **Read locks**: `lock:read` (default for all users)
-  **Create locks**: `lock:create` (developers and above)
-  **Delete locks**: `lock:delete` (lock managers and above)
-  **Force delete locks**: `lock:force` (emergency admins only)

#### Plan Management

-  **Read plans**: `plan:read` (default for all users)
-  **Create plans**: `plan:create` (developers and above)
-  **Apply plans**: `plan:apply` (plan approvers and above)
-  **Delete plans**: `plan:delete` (plan approvers and above)

#### Policy Management

-  **Read policies**: `policy:read` (default for all users)
-  **Write policies**: `policy:write` (policy admins and above)

### Best Practices

1. **Principle of Least Privilege**: Assign only the permissions users need
2. **Role-Based Assignment**: Use groups from your SSO provider for automatic role assignment
3. **Emergency Access**: Create emergency admin roles for critical situations
4. **Regular Review**: Periodically review and update role assignments
5. **Audit Logging**: Log permission checks and access attempts

### Example Use Cases

#### DevOps Team

-  Role: `developer` or `admin`
-  Can create/delete locks, apply plans, manage policies
-  Full access to infrastructure management

#### Security Team

-  Role: `policy-admin`
-  Can read all resources and manage policies
-  Cannot modify infrastructure directly

#### Emergency Response Team

-  Role: `emergency-admin`
-  Can force delete locks and override normal restrictions
-  Used only in critical situations

#### Read-Only Observers

-  Role: `read-only` or `user`
-  Can view all resources but cannot make changes
-  Suitable for auditors and stakeholders
