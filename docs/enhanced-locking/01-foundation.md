# Enhanced Locking System - Foundation

This document describes the foundation layer of Atlantis's enhanced locking system, introduced as part of the modernization effort to improve scalability, reliability, and observability.

## Overview

The enhanced locking system is designed to replace the current BoltDB-based locking mechanism with a more scalable and feature-rich solution. The foundation layer (PR #1) establishes the core architecture and interfaces without breaking existing functionality.

## Architecture

### Core Components

```
┌─────────────────────────────────────────────────────────┐
│                   Enhanced Locking System               │
├─────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐     │
│  │   Config    │  │   Factory   │  │  Migration  │     │
│  │ Management  │  │    & DI     │  │   Support   │     │
│  └─────────────┘  └─────────────┘  └─────────────┘     │
├─────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐     │
│  │ Enhanced    │  │   Backend   │  │   Event     │     │
│  │   Locker    │  │ Interface   │  │  Handling   │     │
│  └─────────────┘  └─────────────┘  └─────────────┘     │
├─────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐     │
│  │   Redis     │  │   BoltDB    │  │   Legacy    │     │
│  │ Backend     │  │  Enhanced   │  │ Fallback    │     │
│  │(Future PR)  │  │(Future PR)  │  │             │     │
│  └─────────────┘  └─────────────┘  └─────────────┘     │
└─────────────────────────────────────────────────────────┘
```

### Key Interfaces

#### EnhancedLocker
The main interface for enhanced locking operations:
- Context-aware lock operations
- Priority-based queuing support
- Advanced timeout handling
- Batch operations
- Administrative operations

#### Backend
The backend interface for different storage implementations:
- Redis backend (future)
- Enhanced BoltDB backend (future)
- Legacy compatibility layer

#### ConfigProvider
Configuration management interface:
- Feature flag control
- Migration settings
- Runtime configuration updates

## Features

### Foundation Layer (PR #1)

- ✅ **Core Types & Interfaces**: Complete type system for enhanced locking
- ✅ **Configuration System**: Flexible configuration with feature flags
- ✅ **Dependency Injection**: Factory pattern for component creation
- ✅ **Migration Support**: Gradual rollout with percentage-based routing
- ✅ **Legacy Compatibility**: Seamless fallback to existing system
- ✅ **Feature Flags**: Runtime control of enhanced features

### Planned Features (Future PRs)

- 🔄 **Priority Queuing**: Fair scheduling with anti-starvation
- 🔄 **Redis Backend**: Distributed locking with Redis
- 🔄 **Deadlock Detection**: Automatic deadlock detection and resolution
- 🔄 **Event Streaming**: Real-time lock events
- 🔄 **Metrics & Monitoring**: Comprehensive observability
- 🔄 **Performance Optimization**: Connection pooling and caching

## Configuration

### Environment Variables

```bash
# Enable enhanced locking system
ATLANTIS_ENHANCED_LOCKING_ENABLED=false

# Backend selection
ATLANTIS_ENHANCED_LOCKING_BACKEND=boltdb  # or 'redis'

# Feature flags
ATLANTIS_ENHANCED_LOCKING_PRIORITY_QUEUE=false
ATLANTIS_ENHANCED_LOCKING_RETRIES=false
ATLANTIS_ENHANCED_LOCKING_METRICS=true

# Redis configuration (when backend=redis)
ATLANTIS_ENHANCED_LOCKING_REDIS_ADDR=localhost:6379
ATLANTIS_ENHANCED_LOCKING_REDIS_PASSWORD=""
ATLANTIS_ENHANCED_LOCKING_REDIS_DB=0

# Timeouts
ATLANTIS_ENHANCED_LOCKING_DEFAULT_TIMEOUT=30s
ATLANTIS_ENHANCED_LOCKING_MAX_TIMEOUT=5m

# Queue configuration
ATLANTIS_ENHANCED_LOCKING_MAX_QUEUE_SIZE=1000
ATLANTIS_ENHANCED_LOCKING_STARVATION_THRESHOLD=2m
ATLANTIS_ENHANCED_LOCKING_MAX_PRIORITY_BOOST=3
```

### Command Line Flags

```bash
# Enable enhanced locking
--enhanced-locking-enabled

# Backend selection
--enhanced-locking-backend=boltdb

# Redis connection
--enhanced-locking-redis-addr=localhost:6379
--enhanced-locking-redis-password=secret
--enhanced-locking-redis-db=0

# Feature toggles
--enhanced-locking-priority-queue
--enhanced-locking-retries
--enhanced-locking-metrics
```

### YAML Configuration

```yaml
# atlantis.yaml server configuration
enhanced-locking-enabled: false
enhanced-locking-backend: boltdb
enhanced-locking-priority-queue: false
enhanced-locking-retries: false
enhanced-locking-metrics: true

# Redis configuration
enhanced-locking-redis-addr: localhost:6379
enhanced-locking-redis-password: ""
enhanced-locking-redis-db: 0
```

## Migration Strategy

The enhanced locking system uses a gradual migration approach:

### Phase 1: Foundation (Current PR)
- ✅ Core interfaces and types
- ✅ Configuration framework
- ✅ Feature flags infrastructure
- ✅ Legacy compatibility layer
- ❌ **Enhanced features disabled by default**

### Phase 2: Implementation (Future PRs)
- Redis backend implementation
- Priority queue implementation
- Enhanced BoltDB backend
- Event streaming system

### Phase 3: Migration (Future PRs)
- Gradual traffic routing
- A/B testing framework
- Data migration tools
- Performance comparison

### Phase 4: Optimization (Future PRs)
- Performance optimizations
- Advanced features
- Monitoring and alerting

## Safety Measures

### Backward Compatibility
- All existing APIs remain unchanged
- Legacy locking behavior is preserved
- Fallback mechanisms prevent service disruption
- Configuration is opt-in only

### Feature Flags
- Enhanced locking disabled by default
- Individual features can be toggled
- Runtime configuration updates
- Emergency rollback capability

### Migration Controls
- Percentage-based traffic routing
- Automatic fallback on errors
- Validation mode for comparison
- Comprehensive logging

## Code Organization

```
server/core/locking/enhanced/
├── types.go           # Core types and interfaces
├── config.go          # Configuration management
├── factory.go         # Dependency injection
├── placeholders.go    # Placeholder implementations
└── README.md          # Package documentation

docs/enhanced-locking/
├── 01-foundation.md   # This document
├── 02-migration.md    # Migration guide (future)
├── 03-redis.md        # Redis backend (future)
└── 04-monitoring.md   # Monitoring guide (future)
```

## Testing Strategy

### Foundation Tests
- Configuration validation
- Feature flag management
- Factory creation patterns
- Migration routing logic

### Integration Tests
- Legacy compatibility verification
- Fallback behavior testing
- Configuration loading
- Service locator functionality

### Future Test Categories
- Backend implementation tests
- Performance benchmarks
- Chaos engineering tests
- Migration validation tests

## Performance Considerations

### Current Impact
- **Zero performance impact** when disabled (default)
- Minimal overhead for configuration management
- No changes to existing lock paths
- Memory footprint under 1MB for types

### Future Optimizations
- Connection pooling for Redis
- Async lock operations
- Batch lock management
- Memory-efficient data structures

## Security Considerations

### Foundation Security
- Configuration validation
- Input sanitization
- No credential exposure in logs
- Secure defaults

### Future Security Features
- Redis authentication support
- TLS encryption for Redis
- Audit logging
- Access control lists

## Monitoring and Observability

### Current Capabilities
- Feature flag status logging
- Configuration validation errors
- Migration routing decisions
- Legacy fallback events

### Future Monitoring
- Lock acquisition metrics
- Queue depth monitoring
- Backend health checks
- Performance dashboards

## Development Guidelines

### Adding New Features
1. Define interfaces in `types.go`
2. Update configuration in `config.go`
3. Add factory methods in `factory.go`
4. Implement in separate PR
5. Update documentation

### Testing New Components
1. Unit tests for interfaces
2. Integration tests for factories
3. End-to-end tests for workflows
4. Performance benchmarks
5. Chaos testing

### Configuration Changes
1. Add environment variable support
2. Update user config structure
3. Add validation logic
4. Update documentation
5. Test migration scenarios

## FAQ

### Q: Why is enhanced locking disabled by default?
A: To ensure zero impact on existing installations. Enhanced features will be enabled gradually as they mature.

### Q: Will existing locks be migrated automatically?
A: No automatic migration in the foundation layer. Future PRs will include migration tools and strategies.

### Q: What happens if Redis is unavailable?
A: The system will automatically fall back to the legacy BoltDB implementation with logging.

### Q: Can I enable enhanced locking in production?
A: The foundation layer is safe but provides no enhanced features. Wait for implementation PRs.

### Q: How do I contribute to the enhanced locking system?
A: See the main project contribution guidelines and the enhanced locking design documents.

## Related Documents

- [Enhanced Locking Design Document](../design/enhanced-locking.md)
- [Migration Strategy](02-migration.md) (coming in future PR)
- [Redis Backend Implementation](03-redis.md) (coming in future PR)
- [Monitoring and Observability](04-monitoring.md) (coming in future PR)

## Changelog

### v1.0.0 - Foundation Release
- Initial foundation layer implementation
- Core types and interfaces
- Configuration framework
- Feature flags system
- Legacy compatibility layer
- Placeholder implementations