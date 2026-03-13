# Phase 4: Complete Migration to VCS Plugin System

## 🚀 Overview

This PR completes the migration to the VCS plugin system by removing all legacy VCS code and making the plugin system the sole VCS interface. This is the final phase of the migration strategy outlined in issue #5574.

**Migration Phase**: 4/4 - **Complete Migration**
**Breaking Changes**: ⚠️ **YES** - Legacy VCS flags and configurations are no longer supported
**Rollback Strategy**: Available via feature flags for emergency rollback

## 📋 Changes Summary

### ✅ **Removed (Legacy Code)**

- [ ] **Legacy VCS Flags**: All deprecated `--gh-*` and `--gitlab-*` flags
- [ ] **Legacy Configuration**: VCS-specific configuration structs and validation
- [ ] **Legacy VCS Clients**: Direct GitHub/GitLab client implementations
- [ ] **Legacy Event Handlers**: VCS-specific event handling code
- [ ] **Legacy Permissions**: Provider-specific permission checking
- [ ] **Legacy Tests**: Tests for removed legacy functionality

### 🆕 **Added/Finalized (Plugin System)**

- [ ] **Production VCS Plugins**: Complete GitHub, GitLab, Bitbucket, and Azure DevOps plugins
- [ ] **Plugin Auto-Discovery**: Automatic plugin registration and configuration
- [ ] **Enhanced Capabilities**: Advanced feature detection and validation
- [ ] **Monitoring & Metrics**: VCS plugin performance and health metrics
- [ ] **Configuration Migration**: Automatic conversion of legacy configurations
- [ ] **Admin UI**: Web interface for VCS plugin management

### 🔄 **Modified (Existing Code)**

- [ ] **Server Initialization**: Complete integration of VCS service
- [ ] **Event Processing**: All events routed through plugin system
- [ ] **Configuration Loading**: Plugin-based configuration only
- [ ] **Documentation**: Updated to reflect plugin-only architecture
- [ ] **CLI Commands**: Migration commands and plugin management

## 🗂️ File Changes

### **Removed Files** (Legacy VCS)
```
server/events/vcs/
├── github_client.go              # ❌ REMOVED
├── gitlab_client.go              # ❌ REMOVED  
├── bitbucket_client.go          # ❌ REMOVED
├── azuredevops_client.go        # ❌ REMOVED
├── github_credentials.go        # ❌ REMOVED
├── gitlab_credentials.go        # ❌ REMOVED
└── legacy_vcs_wrapper.go        # ❌ REMOVED

server/events/
├── github_app_working_dir.go    # ❌ REMOVED
├── gitlab_request_parser.go     # ❌ REMOVED
├── bitbucket_request_parser.go  # ❌ REMOVED
└── legacy_event_handlers.go     # ❌ REMOVED

server/
├── legacy_server_config.go      # ❌ REMOVED
├── github_app_setup.go         # ❌ REMOVED
└── gitlab_setup.go             # ❌ REMOVED
```

### **Added Files** (Production Plugins)
```
internal/infrastructure/vcs/plugins/
├── github/
│   ├── plugin.go                # ✅ Production GitHub plugin
│   ├── client.go               # ✅ GitHub API client
│   ├── auth.go                 # ✅ GitHub authentication
│   ├── webhooks.go             # ✅ GitHub webhook handling
│   └── capabilities.go         # ✅ GitHub capabilities
├── gitlab/
│   ├── plugin.go               # ✅ Production GitLab plugin
│   ├── client.go              # ✅ GitLab API client
│   ├── auth.go                # ✅ GitLab authentication
│   └── capabilities.go        # ✅ GitLab capabilities
├── bitbucket/
│   └── plugin.go              # ✅ Production Bitbucket plugin
└── azuredevops/
    └── plugin.go              # ✅ Production Azure DevOps plugin

internal/infrastructure/vcs/
├── plugin_loader.go            # ✅ Dynamic plugin loading
├── config_converter.go         # ✅ Legacy config conversion
├── metrics.go                  # ✅ VCS plugin metrics
└── admin_api.go               # ✅ Admin API for plugin management
```

### **Modified Files** (Integration)
```
server/
├── server.go                   # 🔄 Complete VCS service integration
├── events_controller.go        # 🔄 Plugin-only event routing
└── user_config.go             # 🔄 Plugin configuration only

cmd/server/
└── main.go                     # 🔄 Plugin system initialization

docs/
├── README.md                   # 🔄 Updated for plugin system
├── configuration.md            # 🔄 Plugin configuration docs
└── vcs-plugins.md             # 🔄 Plugin usage guide
```

## 🚨 Breaking Changes

### **Removed Command Line Flags**

All legacy VCS flags are **permanently removed**:

```bash
# ❌ NO LONGER SUPPORTED
--gh-allow-mergeable-bypass-apply
--gh-team-allowlist
--gh-org-allowlist  
--gh-app-id
--gh-app-key-file
--gh-app-slug
--gh-hostname
--gh-token
--gh-user
--gh-webhook-secret

--gitlab-hostname
--gitlab-token
--gitlab-user
--gitlab-webhook-secret
--gitlab-group-allowlist

--bitbucket-user
--bitbucket-token
--bitbucket-webhook-secret
--bitbucket-base-url

--azuredevops-hostname
--azuredevops-token
--azuredevops-user
--azuredevops-webhook-user
--azuredevops-webhook-password
```

### **New Required Configuration**

Plugin-based configuration is now **mandatory**:

```yaml
# ✅ REQUIRED: Plugin configuration
vcs:
  plugins:
    enabled: true
    default_provider: github
    
  github:
    plugin: github-v1
    app_id: "${ATLANTIS_GH_APP_ID}"
    app_key_file: "${ATLANTIS_GH_APP_KEY_FILE}"
    hostname: "${ATLANTIS_GH_HOSTNAME}"
    capabilities:
      mergeable_bypass: true
      team_allowlist: true
      
  gitlab:
    plugin: gitlab-v1  
    hostname: "${ATLANTIS_GITLAB_HOSTNAME}"
    token: "${ATLANTIS_GITLAB_TOKEN}"
    capabilities:
      group_allowlist: true
      merge_request_approvals: true
```

### **Environment Variable Changes**

```bash
# ❌ REMOVED
ATLANTIS_GH_TOKEN
ATLANTIS_GH_USER
ATLANTIS_GH_APP_ID
ATLANTIS_GITLAB_TOKEN
ATLANTIS_GITLAB_USER

# ✅ NEW REQUIRED
ATLANTIS_VCS_PLUGINS_ENABLED=true
ATLANTIS_VCS_PLUGINS_CONFIG_FILE=/path/to/vcs-config.yaml
ATLANTIS_VCS_DEFAULT_PROVIDER=github

# ✅ PLUGIN-SPECIFIC
ATLANTIS_VCS_GITHUB_APP_ID
ATLANTIS_VCS_GITHUB_APP_KEY_FILE  
ATLANTIS_VCS_GITLAB_TOKEN
ATLANTIS_VCS_GITLAB_HOSTNAME
```

## 📖 Migration Guide

### **Automatic Migration**

The migration process is **automated** for most configurations:

```bash
# 1. Run migration detector
atlantis migrate detect

# 2. Generate new configuration
atlantis migrate convert --output vcs-config.yaml

# 3. Validate new configuration  
atlantis migrate validate --config vcs-config.yaml

# 4. Test plugin functionality
atlantis migrate test --config vcs-config.yaml --dry-run
```

### **Manual Migration Steps**

For complex configurations requiring manual intervention:

#### **Step 1: Convert GitHub App Configuration**

```yaml
# Before (command line flags)
# --gh-app-id=12345 --gh-app-key-file=/app/key.pem

# After (plugin configuration)
vcs:
  github:
    plugin: github-v1
    app_id: 12345
    app_key_file: /app/key.pem
    capabilities:
      mergeable_bypass: true
      team_allowlist: true
```

#### **Step 2: Convert GitLab Configuration**

```yaml
# Before (command line flags)  
# --gitlab-token=glpat-xxx --gitlab-hostname=gitlab.company.com

# After (plugin configuration)
vcs:
  gitlab:
    plugin: gitlab-v1
    hostname: gitlab.company.com
    token: glpat-xxx
    capabilities:
      group_allowlist: true
      merge_request_approvals: true
```

#### **Step 3: Update Environment Variables**

```bash
# Convert existing environment variables
export ATLANTIS_VCS_PLUGINS_ENABLED=true
export ATLANTIS_VCS_DEFAULT_PROVIDER=github
export ATLANTIS_VCS_GITHUB_APP_ID="${ATLANTIS_GH_APP_ID}"
export ATLANTIS_VCS_GITHUB_APP_KEY_FILE="${ATLANTIS_GH_APP_KEY_FILE}"
```

### **Configuration Validation**

```bash
# Validate plugin configuration
atlantis config validate --vcs-plugins

# Test VCS connectivity
atlantis plugins test --provider github
atlantis plugins test --provider gitlab

# Check capability detection
atlantis plugins capabilities --provider github
```

## 🧪 Testing Strategy

### **Pre-Deployment Testing**

- [ ] **Unit Tests**: All plugin implementations have >90% coverage
- [ ] **Integration Tests**: End-to-end VCS operations for each provider
- [ ] **Load Tests**: Plugin performance under typical Atlantis workloads
- [ ] **Compatibility Tests**: Existing repositories work without changes
- [ ] **Migration Tests**: Legacy configurations convert correctly

### **Canary Deployment**

```yaml
# Phase 4A: Internal testing (1 week)
deployment:
  canary_percentage: 0%
  internal_testing: true
  rollback_triggers:
    - vcs_error_rate > 1%
    - plugin_health_failures > 5%

# Phase 4B: Limited rollout (1 week)  
deployment:
  canary_percentage: 10%
  monitoring:
    - vcs_operation_latency
    - plugin_success_rate
    - configuration_errors

# Phase 4C: Full deployment (1 week)
deployment: 
  canary_percentage: 100%
  legacy_system: disabled
```

### **Monitoring & Alerting**

```yaml
alerts:
  - name: VCS Plugin Failure Rate
    condition: vcs_plugin_error_rate > 5%
    action: page_oncall
    
  - name: Plugin Health Check Failure
    condition: vcs_plugin_health_check_failures > 3
    action: rollback_deployment
    
  - name: Configuration Migration Errors
    condition: config_migration_errors > 10
    action: notify_team
```

## 🔄 Rollback Strategy

### **Emergency Rollback**

If critical issues are discovered, emergency rollback is available:

```bash
# 1. Enable legacy compatibility mode
export ATLANTIS_VCS_LEGACY_FALLBACK=true
export ATLANTIS_VCS_PLUGINS_ENABLED=false

# 2. Restart Atlantis with legacy flags
atlantis server \
  --gh-app-id=12345 \
  --gh-app-key-file=/app/key.pem \
  --gitlab-token=glpat-xxx

# 3. Monitor for stability
atlantis health --legacy-mode
```

### **Gradual Rollback**

For non-critical issues, gradual rollback per provider:

```yaml
vcs:
  plugins:
    enabled: true
    fallback_providers:
      github: legacy  # Rollback GitHub to legacy
      gitlab: plugin  # Keep GitLab on plugin
```

## 📊 Success Metrics

### **Technical Metrics**

- [ ] **Error Rate**: VCS operation error rate < 1%
- [ ] **Latency**: P95 VCS operation latency < 2s
- [ ] **Availability**: Plugin system uptime > 99.9%
- [ ] **Memory Usage**: Memory usage reduction > 20%
- [ ] **Plugin Health**: All plugins passing health checks

### **Operational Metrics**

- [ ] **Migration Success**: >95% of configurations migrate automatically
- [ ] **Support Tickets**: <10 migration-related support tickets
- [ ] **Documentation**: Plugin system documentation complete
- [ ] **Team Adoption**: Development team comfortable with plugin system

## 🚀 Deployment Plan

### **Timeline: 3 Weeks**

#### **Week 1: Pre-deployment**
- [ ] **Code Review**: Comprehensive review of all changes
- [ ] **Testing**: Complete testing suite execution
- [ ] **Documentation**: Finalize migration guides and plugin docs
- [ ] **Team Training**: Train support team on plugin system

#### **Week 2: Canary Deployment**  
- [ ] **Internal Deployment**: Deploy to internal Atlantis instances
- [ ] **Migration Testing**: Test configuration migration tools
- [ ] **Performance Validation**: Validate performance improvements
- [ ] **Feedback Collection**: Gather feedback from internal users

#### **Week 3: Production Rollout**
- [ ] **Staged Rollout**: 10% → 50% → 100% over 3 days
- [ ] **Real-time Monitoring**: 24/7 monitoring during rollout
- [ ] **Support Readiness**: Support team on standby
- [ ] **Legacy Cleanup**: Remove legacy code after successful rollout

## 🔗 Related Issues

- Closes #5574: VCS-specific features support
- Closes #XXXX: Plugin system architecture
- Closes #XXXX: Migration tooling
- Closes #XXXX: Legacy code cleanup

## 📚 Documentation Updates

- [ ] **README.md**: Updated with plugin system overview
- [ ] **configuration.md**: Plugin configuration guide
- [ ] **vcs-plugins.md**: Detailed plugin usage documentation
- [ ] **migration-guide.md**: Step-by-step migration instructions
- [ ] **troubleshooting.md**: Plugin system troubleshooting guide

## ✅ Pre-merge Checklist

- [ ] All automated tests pass
- [ ] Manual testing completed for all VCS providers
- [ ] Migration tools tested with sample configurations
- [ ] Documentation reviewed and approved
- [ ] Security review completed
- [ ] Performance benchmarks met
- [ ] Rollback plan validated
- [ ] Support team trained
- [ ] Monitoring dashboards configured
- [ ] Emergency contacts notified

---

**🎯 This PR represents the completion of our VCS modernization journey. After this merge, Atlantis will have a clean, extensible, and maintainable VCS integration architecture that supports multiple providers with provider-specific features while maintaining clean abstractions.**

**⚠️ IMPORTANT: This is a breaking change. Please ensure you have completed the migration process and have a rollback plan before merging.** 