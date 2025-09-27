# Enhanced Locking System - PR Cross-Reference Guide

## CRITICAL: Correct PR Structure

This document establishes the **correct** cross-reference structure for the Enhanced Locking System. All documentation and communications must use these exact PR numbers.

## 📋 Official PR Mapping

### Documentation Hub
- **PR #0**: **#5845** - 📚 Documentation Hub (ONLY .md files)
  - Location: `docs/enhanced-locking/*.md`
  - Contains: All documentation, migration guides, examples
  - **NO Go code** - Documentation ONLY

### Go Implementation PRs
- **PR #1**: **#5842** - 🏗️ Foundation (Go implementation)
  - Location: `server/core/locking/enhanced/`
  - Contains: Core types, interfaces, configuration

- **PR #2**: **#5836** - 🔄 Compatibility (Go implementation)
  - Location: `server/core/locking/enhanced/`
  - Contains: Backward compatibility layer

- **PR #3**: **#5840** - ⚡ Redis Backend (Go implementation)
  - Location: `server/core/locking/enhanced/`
  - Contains: Redis backend implementation

- **PR #4**: **#5843** - 📊 Manager Events (Go implementation)
  - Location: `server/core/locking/enhanced/`
  - Contains: Central manager and event system

- **PR #5**: **#5841** - 📋 Priority Queue (Go implementation)
  - Location: `server/core/locking/enhanced/`
  - Contains: Priority queuing system

## 🚨 CRITICAL CORRECTIONS

### ❌ NEVER Reference These
- `https://github.com/runatlantis/atlantis/pull/1` (WRONG - does not exist)
- `PR #0` referring to a real GitHub PR #0 (WRONG - use #5845)
- "Future PR" for implemented features (WRONG - update to actual PR numbers)

### ✅ ALWAYS Use These
- **#5845** for documentation-only PR #0
- **#5842, #5836, #5840, #5843, #5841** for Go implementation PRs
- Specific GitHub issue numbers with # prefix
- "Future" only for genuinely unimplemented features

## 📊 Architecture Diagram References

```
PR #0 (#5845): 📚 Documentation Hub (ONLY .md files)
├── docs/enhanced-locking/*.md
├── docs/enhanced-locking/migration/*.md
└── docs/enhanced-locking/examples/*.md

PR #1 (#5842): 🏗️ Foundation (Go implementation)
├── Core types and interfaces
├── Configuration framework
└── Feature flags infrastructure

PR #2 (#5836): 🔄 Compatibility (Go implementation)
├── Backward compatibility layer
├── Legacy fallback mechanisms
└── Migration support

PR #3 (#5840): ⚡ Redis Backend (Go implementation)
├── Redis cluster support
├── Distributed locking
└── Connection pooling

PR #4 (#5843): 📊 Manager Events (Go implementation)
├── Central lock manager
├── Event system
└── Metrics collection

PR #5 (#5841): 📋 Priority Queue (Go implementation)
├── Priority-based queuing
├── Anti-starvation logic
└── Timeout management
```

## 🎯 Team Separation

### Core Developers (Go Implementation)
Focus on PRs **#1-5** (#5842, #5836, #5840, #5843, #5841):
- Review Go source code changes
- Validate implementation architecture
- Test functional requirements
- Performance and integration testing

### Documentation Reviewers
Focus on PR **#0** (#5845):
- Review documentation accuracy
- Validate migration guides
- Test configuration examples
- Ensure clarity and completeness

## 📝 Usage in Documentation

### ✅ Correct Examples
```markdown
- Foundation layer (#5842) provides core interfaces
- Redis backend (#5840) enables distributed locking
- Documentation is maintained in #5845
- See PR #3 (#5840) for Redis implementation details
```

### ❌ Incorrect Examples
```markdown
- Foundation layer (PR #1) provides core interfaces  // Missing actual number
- Redis backend (Future PR) enables distributed locking  // Implemented, not future
- See https://github.com/runatlantis/atlantis/pull/1  // Link doesn't exist
- Documentation is in PR #0  // Ambiguous, missing actual number
```

## 🔍 Quick Reference Table

| Feature | PR Label | GitHub PR | Status | File Location |
|---------|----------|-----------|--------|---------------|
| Documentation | PR #0 | #5845 | ✅ Complete | `docs/enhanced-locking/` |
| Foundation | PR #1 | #5842 | ✅ Complete | `server/core/locking/enhanced/` |
| Compatibility | PR #2 | #5836 | ✅ Complete | `server/core/locking/enhanced/` |
| Redis Backend | PR #3 | #5840 | ✅ Complete | `server/core/locking/enhanced/` |
| Manager Events | PR #4 | #5843 | ✅ Complete | `server/core/locking/enhanced/` |
| Priority Queue | PR #5 | #5841 | ✅ Complete | `server/core/locking/enhanced/` |

## 🚀 Status Updates

All core features have been implemented:
- ✅ **Foundation** (#5842) - Core architecture complete
- ✅ **Compatibility** (#5836) - Legacy support complete
- ✅ **Redis Backend** (#5840) - Distributed locking complete
- ✅ **Manager Events** (#5843) - Orchestration complete
- ✅ **Priority Queue** (#5841) - Advanced queuing complete
- ✅ **Documentation** (#5845) - Comprehensive docs complete

Future features:
- ⏳ Advanced deadlock detection
- ⏳ Machine learning integration
- ⏳ Cross-region support

---

**Last Updated**: September 2025
**Maintained by**: Hive Mind Cross-Reference Fix Agent
**Authority**: This document is the official source of truth for PR cross-references