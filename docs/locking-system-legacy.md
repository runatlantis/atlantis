# Atlantis Legacy Locking System Documentation

## Overview

The current Atlantis locking system manages project-level locks to prevent concurrent execution of Terraform operations on the same project/workspace combination. This document details the legacy system architecture, data structures, and implementation patterns.

## System Architecture

### Core Components

1. **Locking Interface (`Backend`)**: Defines the contract for lock storage implementations
2. **Client Layer (`locking.Client`)**: Provides high-level locking operations and key management
3. **Storage Backends**:
   - **BoltDB** (`server/core/db/boltdb.go`) - File-based embedded database
   - **Redis** (`server/core/redis/redis.go`) - Remote Redis storage
4. **Apply Locking** (`apply_locking.go`) - Global apply command locking mechanism

### Backend Interface

```go
type Backend interface {
    TryLock(lock models.ProjectLock) (bool, models.ProjectLock, error)
    Unlock(project models.Project, workspace string) (*models.ProjectLock, error)
    List() ([]models.ProjectLock, error)
    GetLock(project models.Project, workspace string) (*models.ProjectLock, error)
    UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error)
    UpdateProjectStatus(pull models.PullRequest, workspace string, repoRelDir string, newStatus models.ProjectPlanStatus) error
    GetPullStatus(pull models.PullRequest) (*models.PullStatus, error)
    DeletePullStatus(pull models.PullRequest) error
    UpdatePullWithResults(pull models.PullRequest, newResults []command.ProjectResult) (models.PullStatus, error)

    // Global command locking
    LockCommand(cmdName command.Name, lockTime time.Time) (*command.Lock, error)
    UnlockCommand(cmdName command.Name) error
    CheckCommandLock(cmdName command.Name) (*command.Lock, error)
}
```

## Data Models

### ProjectLock Structure

```go
type ProjectLock struct {
    // Project is the project that is being locked.
    Project Project
    // Pull is the pull request from which the command was run that created this lock.
    Pull PullRequest
    // User is the username of the user that ran the command that created this lock.
    User User
    // Workspace is the Terraform workspace that this lock is being held against.
    Workspace string
    // Time is the time at which the lock was first created.
    Time time.Time
}
```

### Project Structure

```go
type Project struct {
    // ProjectName of the project
    ProjectName string
    // RepoFullName is the owner and repo name, ex. "runatlantis/atlantis"
    RepoFullName string
    // Path to project root in the repo.
    // If "." then project is at root.
    // Never ends in "/".
    Path string
}
```

## Lock Key Generation

### Key Format Algorithm

The lock key uniquely identifies a project/workspace combination:

```go
func GenerateLockKey(project Project, workspace string) string {
    return fmt.Sprintf("%s/%s/%s", project.RepoFullName, project.Path, workspace)
}
```

**Key Format**: `{repoFullName}/{path}/{workspace}`

**Examples**:
- `runatlantis/atlantis/./default` - Root project, default workspace
- `runatlantis/atlantis/modules/vpc/production` - VPC module, production workspace
- `company/infrastructure/terraform/aws/dev` - AWS infrastructure, dev workspace

### Key Parsing

```go
// keyRegex matches and captures {repoFullName}/{path}/{workspace}
var keyRegex = regexp.MustCompile(`^(.*?\/.*?)\/(.*)\/(.*)$`)

func (c *Client) lockKeyToProjectWorkspace(key string) (models.Project, string, error) {
    matches := keyRegex.FindStringSubmatch(key)
    if len(matches) != 4 {
        return models.Project{}, "", errors.New("invalid key format")
    }

    return models.Project{RepoFullName: matches[1], Path: matches[2]}, matches[3], nil
}
```

## Storage Backend Implementations

### BoltDB Backend

**File Location**: `{dataDir}/atlantis.db`

**Bucket Structure**:
- `runLocks` - Project locks storage
- `pulls` - Pull request status tracking
- `globalLocks` - Global command locks (apply, etc.)

**Key Storage**:
- **Project locks**: Direct key mapping using `GenerateLockKey()`
- **Global locks**: `{commandName}/lock` format
- **Pull status**: `{hostname}::{repoFullName}::{pullNumber}` format

**Data Format**: JSON serialization of Go structs

```go
func (b *BoltDB) lockKey(p models.Project, workspace string) string {
    return models.GenerateLockKey(p, workspace)
}

func (b *BoltDB) commandLockKey(cmdName command.Name) string {
    return fmt.Sprintf("%s/lock", cmdName)
}
```

### Redis Backend

**Key Prefixes**:
- **Project locks**: `pr/{GenerateLockKey()}`
- **Global locks**: `global/{commandName}/lock`

**Examples**:
- `pr/runatlantis/atlantis/./default`
- `global/apply/lock`

**Data Format**: JSON serialization with Redis string storage

```go
func (r *RedisDB) lockKey(p models.Project, workspace string) string {
    return fmt.Sprintf("pr/%s", models.GenerateLockKey(p, workspace))
}

func (r *RedisDB) commandLockKey(cmdName command.Name) string {
    return fmt.Sprintf("global/%s/lock", cmdName)
}
```

**Redis Operations**:
- **Locking**: `GET` to check, `SET` to acquire
- **Listing**: `SCAN` with pattern `pr*`
- **UnlockByPull**: `SCAN` with pattern `pr/{repoFullName}*`

## Database Schema Differences

### BoltDB vs Redis Key Differences

| Operation | BoltDB Key | Redis Key |
|-----------|------------|-----------|
| Project Lock | `runatlantis/atlantis/./default` | `pr/runatlantis/atlantis/./default` |
| Global Apply Lock | `apply/lock` | `global/apply/lock` |
| Pull Status | `github.com::runatlantis/atlantis::123` | N/A (same bucket) |

**Key Insight**: Redis adds prefixes (`pr/`, `global/`) for namespace organization, while BoltDB uses separate buckets.

## Integration with Atlantis Configuration

### atlantis.yaml Integration

The locking system integrates with `atlantis.yaml` project configurations:

```yaml
version: 3
projects:
  - name: "vpc"
    dir: "terraform/vpc"
    workspace: "production"
  - name: "app"
    dir: "terraform/app"
    workspace: "staging"
```

**Lock Key Generation**:
- Project "vpc": `{repoFullName}/terraform/vpc/production`
- Project "app": `{repoFullName}/terraform/app/staging`

### Path Normalization

```go
func NewProject(repoFullName string, path string, projectName string) Project {
    path = paths.Clean(path)
    if path == "/" {
        path = "."
    }
    // ... rest of constructor
}
```

**Path Rules**:
- Root directory becomes `"."`
- Paths are cleaned and normalized
- No trailing slashes
- Relative paths maintained

## Locking Workflow

### Lock Acquisition Flow

1. **Plan/Apply Command Initiated**
   - Extract project info (repo, path, workspace)
   - Generate lock key using `GenerateLockKey()`

2. **TryLock Operation**
   ```go
   lock := models.ProjectLock{
       Workspace: workspace,
       Time:      time.Now().Local(),
       Project:   project,
       User:      user,
       Pull:      pullRequest,
   }
   lockAcquired, currLock, err := backend.TryLock(lock)
   ```

3. **Backend-Specific Logic**
   - **BoltDB**: Atomic transaction, check existence, serialize JSON
   - **Redis**: `GET` to check, `SET` if free

4. **Lock Response**
   - Success: Return lock details and unique key
   - Failure: Return current lock holder info

### Lock Release Flow

1. **Command Completion/Cancellation**
2. **Unlock by Key**: `client.Unlock(lockKey)`
3. **Key Parsing**: Extract project and workspace from key
4. **Backend Deletion**: Remove from BoltDB bucket or Redis key
5. **Return Deleted Lock**: For audit/logging purposes

## Limitations and Pain Points

### Current System Issues

1. **Key Format Coupling**:
   - Lock keys tightly coupled to internal project structure
   - Changes to key format break existing locks
   - Different backends use different key patterns

2. **Path Handling Complexity**:
   - Path normalization logic scattered across codebase
   - Special case handling for root directory (`"."`)
   - Inconsistent path representations

3. **Backend Inconsistencies**:
   - BoltDB and Redis have different key prefixing strategies
   - Different scanning patterns for bulk operations
   - JSON serialization format differences

4. **Concurrency Limitations**:
   - No distributed locking coordination
   - Race conditions possible during high concurrency
   - Limited lock timeout/TTL support

5. **Error Handling**:
   - Limited error differentiation between lock types
   - Missing lock acquisition retry mechanisms
   - Poor error context for debugging

6. **Scalability Issues**:
   - BoltDB single-file limitation
   - Redis memory usage for large lock volumes
   - No lock partitioning or sharding

### Lock Key Collision Risks

**Potential Issues**:
- Repository names with special characters
- Path separators in repo names
- Unicode handling inconsistencies
- Case sensitivity differences across VCS systems

**Example Collision**:
```
Repo: "company/project/sub"
Path: "terraform"
vs
Repo: "company/project"
Path: "sub/terraform"
```
Both generate similar key patterns that could be confused during parsing.

## Global Apply Locking

### ApplyCommandLock System

**Purpose**: Prevent concurrent apply operations across all projects

```go
type ApplyCommandLock struct {
    Locked                 bool
    GlobalApplyLockEnabled bool
    Time                   time.Time
    Failure                string
}
```

**Key Features**:
- Server-wide apply prevention
- Override via `--disable-apply` flag
- Separate from project-level locks
- Time-based lock tracking

**Storage**:
- **BoltDB**: `globalLocks` bucket, key `apply/lock`
- **Redis**: Key `global/apply/lock`

## Memory Usage and Performance

### BoltDB Performance Characteristics

**Advantages**:
- Local file storage, no network overhead
- ACID transaction support
- Automatic persistence
- Single binary deployment

**Limitations**:
- Single writer limitation
- File size growth over time
- No horizontal scaling
- Backup complexity

### Redis Performance Characteristics

**Advantages**:
- High-performance in-memory storage
- Pub/sub capabilities for real-time updates
- Horizontal scaling support
- Rich data structure support

**Limitations**:
- Network latency dependency
- Memory usage scaling
- Persistence configuration complexity
- Additional infrastructure requirement

## Migration Considerations

### Data Migration Challenges

1. **Key Format Changes**:
   - Existing locks using old key formats
   - Need backward compatibility during transition
   - Risk of orphaned locks

2. **Backend Migration**:
   - BoltDB to Redis data export/import
   - Lock state consistency during migration
   - Minimal downtime requirements

3. **Schema Evolution**:
   - Adding new lock metadata fields
   - Changing serialization formats
   - Maintaining API compatibility

### Recommended Migration Strategy

1. **Dual Write Phase**: Write to both old and new systems
2. **Validation Phase**: Verify data consistency between systems
3. **Read Migration**: Switch reads to new system
4. **Cleanup Phase**: Remove old system dependencies

## Security Considerations

### Access Control

**Current Limitations**:
- No user-based lock permissions
- Missing audit trail for lock operations
- Lack of lock ownership validation

**Potential Improvements**:
- Role-based lock access control
- Lock operation audit logging
- User authorization validation
- Lock hijacking prevention

### Data Security

**Current State**:
- Plain JSON storage in both backends
- No encryption at rest
- Limited access logging

**Recommendations**:
- Encrypt sensitive lock metadata
- Add comprehensive audit logging
- Implement lock access rate limiting
- Add lock tampering detection

## Testing and Debugging

### Current Test Coverage

**Test Files**:
- `server/core/locking/locking_test.go` - Client layer tests
- `server/core/db/boltdb_test.go` - BoltDB backend tests
- `server/core/redis/redis_test.go` - Redis backend tests

**Test Patterns**:
- Mock backend implementations for unit tests
- Integration tests with real database instances
- Race condition testing for concurrent operations

### Debugging Tools

**Lock Inspection**:
- `/locks` API endpoint for active locks
- Admin commands for lock manipulation
- Database direct query capabilities

**Common Issues**:
- Orphaned locks after process crashes
- Key format parsing errors
- Backend connection failures

## Future Enhancement Opportunities

### System Improvements

1. **Enhanced Lock Metadata**:
   - Lock expiration/TTL support
   - Lock priority levels
   - Lock dependency tracking

2. **Distributed Coordination**:
   - Consensus-based locking
   - Cross-instance lock coordination
   - Lock leader election

3. **Performance Optimization**:
   - Lock operation batching
   - Connection pooling improvements
   - Caching layer implementation

4. **Observability**:
   - Metrics for lock operations
   - Lock contention monitoring
   - Performance profiling tools

### API Enhancements

1. **REST API Improvements**:
   - Pagination for lock listings
   - Advanced filtering options
   - Bulk lock operations

2. **Event System**:
   - Lock event notifications
   - Webhook integration
   - Real-time lock updates

3. **CLI Improvements**:
   - Interactive lock management
   - Lock debugging commands
   - Batch operation support

---

*This documentation represents the current state of the Atlantis locking system as of the analysis date. For the most up-to-date information, refer to the source code and recent commit history.*