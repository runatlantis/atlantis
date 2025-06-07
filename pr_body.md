## 🚀 Phase 4: Complete Migration to VCS Plugin System

This PR completes the migration to the VCS plugin system by removing all legacy VCS code and making the plugin system the sole VCS interface. This is the final phase of the migration strategy outlined in issue #5574.

**Migration Phase**: 4/4 - Complete Migration
**Breaking Changes**: ⚠️ **YES** - Legacy VCS flags and configurations are no longer supported
**Rollback Strategy**: Available via feature flags for emergency rollback

## 📋 Changes Summary

### ✅ Removed (Legacy Code)
- **Legacy VCS Flags**: All deprecated `--gh-*` and `--gitlab-*` flags
- **Legacy Configuration**: VCS-specific configuration structs and validation
- **Legacy VCS Clients**: Direct GitHub/GitLab client implementations
- **Legacy Event Handlers**: VCS-specific event handling code
- **Legacy Permissions**: Provider-specific permission checking

### 🆕 Added (Plugin System)
- **Production VCS Plugins**: Complete GitHub, GitLab, Bitbucket, and Azure DevOps plugins
- **Plugin Auto-Discovery**: Automatic plugin registration and configuration
- **Enhanced Capabilities**: Advanced feature detection and validation
- **Monitoring & Metrics**: VCS plugin performance and health metrics
- **Configuration Migration**: Automatic conversion of legacy configurations

## 🚨 Breaking Changes

### Removed Command Line Flags
All legacy VCS flags are **permanently removed**:
```bash
# ❌ NO LONGER SUPPORTED
--gh-allow-mergeable-bypass-apply
--gh-team-allowlist
--gh-app-id / --gh-app-key-file
--gitlab-hostname / --gitlab-token
--gitlab-group-allowlist
# ... and all other legacy VCS flags
```

### New Required Configuration
Plugin-based configuration is now **mandatory**:
```yaml
vcs:
  plugins:
    enabled: true
    default_provider: github
  github:
    plugin: github-v1
    app_id: "${ATLANTIS_GH_APP_ID}"
    capabilities:
      mergeable_bypass: true
      team_allowlist: true
```

## 📖 Migration Guide

### Automatic Migration
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

## 🧪 Testing & Deployment Strategy

### Canary Deployment (3 weeks)
- **Week 1**: Internal testing (0% rollout)
- **Week 2**: Limited rollout (10% of traffic)
- **Week 3**: Full production (100% rollout)

### Emergency Rollback
```bash
# Enable legacy compatibility mode
export ATLANTIS_VCS_LEGACY_FALLBACK=true
export ATLANTIS_VCS_PLUGINS_ENABLED=false
```

## 📊 Success Metrics
- **Error Rate**: VCS operation error rate < 1%
- **Latency**: P95 VCS operation latency < 2s
- **Migration Success**: >95% of configurations migrate automatically
- **Support Impact**: <10 migration-related support tickets

## 🔗 Related Issues
- Closes #5574: VCS-specific features support
- Implements complete VCS plugin architecture
- Enables multi-provider support with provider-specific features
- Provides clean migration path from legacy system

## ✅ Pre-merge Checklist
- [ ] All automated tests pass
- [ ] Manual testing completed for all VCS providers
- [ ] Migration tools tested with sample configurations
- [ ] Documentation reviewed and approved
- [ ] Security review completed
- [ ] Performance benchmarks met
- [ ] Rollback plan validated
- [ ] Support team trained

---

**🎯 This PR represents the completion of our VCS modernization journey. After this merge, Atlantis will have a clean, extensible, and maintainable VCS integration architecture that supports multiple providers with provider-specific features while maintaining clean abstractions.**

**⚠️ IMPORTANT: This is a breaking change. Please ensure you have completed the migration process and have a rollback plan before merging.** 