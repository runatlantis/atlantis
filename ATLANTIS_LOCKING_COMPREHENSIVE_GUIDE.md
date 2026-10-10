# Atlantis Locking Mechanisms - Comprehensive Guide

> **Document Status**: Authoritative Reference
> **Last Updated**: 2025-10-31
> **Atlantis Version**: v0.35.1+
> **Authors**: Hive Mind Collective Intelligence Analysis

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Quick Reference](#quick-reference)
3. [Lock Types Overview](#lock-types-overview)
4. [Project Locks (Persistent)](#project-locks-persistent)
5. [Working Directory Locks (In-Memory)](#working-directory-locks-in-memory)
6. [Apply Command Locks (Global)](#apply-command-locks-global)
7. [Lock Interactions & Workflows](#lock-interactions--workflows)
8. [Architecture Deep Dive](#architecture-deep-dive)
9. [Historical Context & Evolution](#historical-context--evolution)
10. [Troubleshooting Guide](#troubleshooting-guide)
11. [Best Practices](#best-practices)
12. [Developer Reference](#developer-reference)

---

## Executive Summary

Atlantis implements a **three-tiered hierarchical locking system** to prevent race conditions and ensure safe concurrent Terraform operations:

1. **Global Command Locks** - System-wide apply command control
2. **Project Locks** - Database-persisted locks preventing concurrent plan/apply on same project+workspace
3. **Working Directory Locks** - In-memory locks protecting filesystem operations

These mechanisms work together to enable **parallel execution** while maintaining **data integrity** across repositories, projects, workspaces, and pull requests.

---

## Quick Reference

### Common Lock Error Messages

| Error Message | Lock Type | Meaning | Resolution |
|--------------|-----------|---------|------------|
| "workspace is currently locked by another command" | Working Directory | Another command executing in same workspace | Wait for completion (~seconds to minutes) |
| "project locked by pull #X" | Project Lock | Another PR has unapplied plan | Apply PR #X or run `atlantis unlock` |
| "Apply commands are locked globally" | Apply Command Lock | Global apply disabled | Admin must unlock or remove `--disable-apply` |

### Quick Commands

```bash
# View all locks
GET /locks

# Unlock specific project
atlantis unlock -r owner/repo -p path -w workspace

# Delete lock by ID (UI)
DELETE /locks/{lockID}

# Check apply lock status
GET /lock/apply/status
```

---

## Lock Types Overview

### Comparison Matrix

| Attribute | Project Lock | Working Dir Lock | Apply Command Lock |
|-----------|--------------|------------------|-------------------|
| **Storage** | Database (BoltDB/Redis) | In-memory (sync.Map) | Database |
| **Scope** | Project + Workspace | Repo + PR + Workspace + Path | Global (all repos) |
| **Persistence** | Survives restarts | Lost on restart | Survives restarts |
| **Acquisition** | Explicit (plan/apply) | Automatic (command start) | Manual (admin) |
| **Release** | Manual/Auto (apply) | Automatic (command end) | Manual |
| **Duration** | Hours to days | Seconds to minutes | Indefinite |
| **Distributed** | Yes (with Redis) | No (single instance) | Yes (with Redis) |

### Lock Hierarchy

```
┌─────────────────────────────────────────────────────┐
│     LAYER 1: GLOBAL COMMAND LOCK                    │
│     Scope: All apply commands system-wide           │
│     Storage: Database                               │
└──────────────────┬──────────────────────────────────┘
                   ↓
┌─────────────────────────────────────────────────────┐
│     LAYER 2: PROJECT LOCK                           │
│     Scope: Specific project + workspace             │
│     Storage: Database                               │
│     Key: {repo}/{path}/{workspace}                  │
└──────────────────┬──────────────────────────────────┘
                   ↓
┌─────────────────────────────────────────────────────┐
│     LAYER 3: WORKING DIRECTORY LOCK                 │
│     Scope: PR + workspace + path (filesystem)       │
│     Storage: In-memory                              │
│     Key: {repo}/{pullNum}/{workspace}/{path}        │
└─────────────────────────────────────────────────────┘
```

---

## Project Locks (Persistent)

### Purpose

Project locks prevent **concurrent plan/apply operations** on the same Terraform project and workspace, ensuring only one pull request can modify infrastructure state at a time.

### Location

- **Interface**: `/server/core/locking/locking.go`
- **BoltDB Implementation**: `/server/core/boltdb/boltdb.go`
- **Redis Implementation**: `/server/core/redis/redis.go`

### Lock Structure

```go
type ProjectLock struct {
    Project   Project       // Project metadata (repo, path, name)
    Pull      PullRequest   // Which PR holds the lock
    User      User          // Who created the lock
    Workspace string        // Terraform workspace
    Time      time.Time     // Lock creation timestamp
}

type Project struct {
    RepoFullName string  // "owner/repo"
    Path         string  // Relative path from repo root
    ProjectName  string  // Optional project name
}
```

### Lock Key Format

**Format**: `{repoFullName}/{path}/{workspace}`

**Examples**:
```
runatlantis/atlantis/terraform/staging/default
myorg/infrastructure/modules/networking/production
acme/app/.deployments/dev
```

**Key Generation** (`models.go:287-289`):
```go
func GenerateLockKey(project Project, workspace string) string {
    return fmt.Sprintf("%s/%s/%s",
        project.RepoFullName,
        project.Path,
        workspace)
}
```

### Lifecycle

```
┌──────────────┐
│   UNLOCKED   │  ← No database entry exists
└──────┬───────┘
       │ atlantis plan (with repo_locking: true)
       ↓
┌──────────────┐
│    LOCKED    │  ← ProjectLock record in database
│  (Owner: PR) │
└──────┬───────┘
       │
       ├─→ Same PR plans again → Lock reacquired (idempotent)
       ├─→ Different PR tries → Lock denied with error
       │
       ├─→ atlantis apply (success) → UNLOCKED
       ├─→ atlantis unlock → UNLOCKED
       └─→ PR closed/merged → UNLOCKED
```

### Acquisition Behavior

**Scenario 1: No Existing Lock**
```go
resp, _ := locker.TryLock(project, workspace, pull, user)
// resp.LockAcquired == true
// resp.LockKey == "owner/repo/path/workspace"
```

**Scenario 2: Same PR Re-locks**
```go
// PR #100 already holds lock
resp, _ := locker.TryLock(project, workspace, pull100, user)
// resp.LockAcquired == true (idempotent)
```

**Scenario 3: Different PR Attempts Lock**
```go
// PR #100 holds lock, PR #200 tries
resp, _ := locker.TryLock(project, workspace, pull200, user)
// resp.LockAcquired == false
// resp.LockFailureReason == "Project locked by pull #100"
```

### Database Storage

**BoltDB Bucket**: `runLocks`

**Key**: `{repoFullName}/{path}/{workspace}`
**Value**: JSON-serialized `ProjectLock`

**Example**:
```json
{
  "Project": {
    "RepoFullName": "runatlantis/atlantis",
    "Path": "terraform/staging",
    "ProjectName": ""
  },
  "Pull": {
    "Num": 123,
    "URL": "https://github.com/runatlantis/atlantis/pull/123",
    "State": "open"
  },
  "User": {
    "Username": "alice"
  },
  "Workspace": "default",
  "Time": "2025-10-31T10:30:00Z"
}
```

### API Operations

**List All Locks**:
```bash
GET /locks
```

**Get Specific Lock**:
```go
lock, err := locker.GetLock(lockKey)
```

**Unlock Project**:
```go
deletedLock, err := locker.Unlock(lockKey)
```

**Unlock All for PR**:
```go
locks, err := locker.UnlockByPull(repoFullName, pullNum)
```

### Configuration

```yaml
# atlantis.yaml
repos:
  - id: /.*/
    repo_locking: true  # Enable project locks (default)
```

**Disable per-repo**:
```yaml
repos:
  - id: /.*-automation$/
    repo_locking: false  # Use NoOpLocker
```

---

## Working Directory Locks (In-Memory)

### Purpose

Working directory locks prevent **concurrent filesystem operations** in the same working directory, protecting against:
- Git clone/checkout race conditions
- Terraform state file corruption
- Concurrent plan file access

### Location

**Implementation**: `/server/events/working_dir_locker.go`

### Lock Structure

```go
type DefaultWorkingDirLocker struct {
    mutex sync.Mutex           // Protects map operations
    locks map[string]struct{}  // Set of locked workspace keys
}
```

### Lock Key Format

**Format**: `{repo}/{pullNum}/{workspace}/{path}`

**Examples**:
```
runatlantis/atlantis/123/default/.
myorg/infra/456/production/terraform/region/us-west-2
acme/app/789/staging/.
```

**Key Generation** (`working_dir_locker.go:75-77`):
```go
func (d *DefaultWorkingDirLocker) workspaceKey(
    repoFullName string,
    pullNum int,
    workspace string,
    path string,
) string {
    return fmt.Sprintf("%s/%d/%s/%s",
        repoFullName, pullNum, workspace, path)
}
```

### Lifecycle

```
┌──────────────┐
│  AVAILABLE   │  ← Key not in locks map
└──────┬───────┘
       │ TryLock() called
       ↓
┌──────────────┐
│    LOCKED    │  ← Key added to locks map
└──────┬───────┘
       │
       ├─→ Another TryLock() → Error: "workspace locked"
       │
       └─→ unlockFn() called (defer) → AVAILABLE
```

### Acquisition Pattern

```go
unlockFn, err := workingDirLocker.TryLock(repo, pullNum, workspace, path)
if err != nil {
    return err  // Already locked
}
defer unlockFn()  // Guaranteed release

// Protected section: filesystem operations
workingDir.Clone(...)
terraform.Plan(...)
```

### Thread Safety

**Mutex Protection**:
```go
func (d *DefaultWorkingDirLocker) TryLock(...) (func(), error) {
    d.mutex.Lock()
    defer d.mutex.Unlock()  // Critical section

    key := d.workspaceKey(...)

    // Atomic check-and-set
    if _, exists := d.locks[key]; exists {
        return func() {}, errors.New("locked")
    }

    d.locks[key] = struct{}{}

    return func() {
        d.unlock(...)  // Deferred cleanup
    }, nil
}
```

### Ephemeral Nature

**Characteristics**:
- ✅ Fast (in-memory, no I/O)
- ✅ Automatic release via defer
- ❌ Lost on server restart (acceptable)
- ❌ Not distributed (single server only)

**Impact of Restart**:
- All working directory locks cleared
- In-progress commands fail (expected)
- New commands can execute immediately
- No orphaned locks possible

---

## Apply Command Locks (Global)

### Purpose

Global apply locks provide **system-wide control** over apply commands for:
- Maintenance windows
- Emergency freezes
- Change control policies

### Location

**Interface**: `/server/core/locking/apply_locking.go`
**Database**: `/server/core/boltdb/boltdb.go` (lines 175-221)

### Lock Structure

```go
type Lock struct {
    LockMetadata LockMetadata  // Contains UnixTime timestamp
    CommandName  Name          // Command being locked (e.g., "apply")
}

type ApplyCommandLock struct {
    Locked     bool
    Time       time.Time
}
```

### Lock Key Format

**Format**: `command:{commandName}`

**Example**: `command:apply`

### Precedence Rules

```
Priority 1 (Highest): DisableApply config flag
Priority 2: Global apply command lock (database)
Priority 3: Project locks
```

**Check Order** (`apply_locking.go:94-116`):
```go
func (c *ApplyClient) CheckApplyLock() (ApplyCommandLock, error) {
    // 1. Check config flag
    if c.disableApply {
        return ApplyCommandLock{Locked: true}, nil
    }

    // 2. Check global lock flag
    if c.disableGlobalApplyLock {
        return ApplyCommandLock{Locked: false}, nil
    }

    // 3. Check database lock
    lock, err := c.database.CheckCommandLock(command.Apply)
    if err != nil {
        return ApplyCommandLock{}, err
    }

    if lock == nil {
        return ApplyCommandLock{Locked: false}, nil
    }

    return ApplyCommandLock{Locked: true, Time: lock.Time()}, nil
}
```

### Operations

**Enable Apply Lock**:
```bash
# Via API
POST /lock/apply

# Response
{
  "Locked": true,
  "Time": "2025-10-31T15:00:00Z"
}
```

**Disable Apply Lock**:
```bash
# Via API
DELETE /lock/apply

# Via CLI flag (requires restart)
atlantis server --disable-apply=false
```

**Check Status**:
```bash
GET /lock/apply/status
```

---

## Lock Interactions & Workflows

### Plan Command Flow

```
User: atlantis plan
        │
        ▼
┌──────────────────────────────────────┐
│ 1. BUILD COMMAND CONTEXT            │
│    Parse repo, PR, user info        │
└──────────────────────────────────────┘
        │
        ▼
┌──────────────────────────────────────┐
│ 2. ACQUIRE WORKING DIR LOCK          │
│    TryLock(repo, pull, ws, path)     │
│    → Returns unlockFn                │
│    → defer unlockFn()                │
└──────────────────────────────────────┘
        │
        ├─ [LOCKED] → Error: "workspace locked"
        │
        ▼ [ACQUIRED]
┌──────────────────────────────────────┐
│ 3. CLONE REPOSITORY                  │
│    (Protected by working dir lock)   │
└──────────────────────────────────────┘
        │
        ▼
┌──────────────────────────────────────┐
│ 4. BUILD PROJECT CONTEXTS            │
│    Parse atlantis.yaml               │
│    Detect modified projects          │
└──────────────────────────────────────┘
        │
        ▼
┌──────────────────────────────────────┐
│ 5. ACQUIRE PROJECT LOCK              │
│    TryLock(project, workspace, pull) │
│    → Check if locked by other PR     │
└──────────────────────────────────────┘
        │
        ├─ [LOCKED BY DIFFERENT PR] → Error: "Project locked by pull #X"
        ├─ [LOCKED BY SAME PR] → Continue (idempotent)
        │
        ▼ [ACQUIRED]
┌──────────────────────────────────────┐
│ 6. EXECUTE TERRAFORM PLAN            │
│    (Both locks held)                 │
└──────────────────────────────────────┘
        │
        ▼
┌──────────────────────────────────────┐
│ 7. RELEASE WORKING DIR LOCK          │
│    unlockFn() via defer              │
│    (Immediate)                       │
└──────────────────────────────────────┘
        │
        ▼
┌──────────────────────────────────────┐
│ 8. PROJECT LOCK PERSISTS             │
│    Until: apply, unlock, or PR close │
└──────────────────────────────────────┘
```

### Apply Command Flow

```
User: atlantis apply
        │
        ▼
┌──────────────────────────────────────┐
│ 1. CHECK GLOBAL APPLY LOCK           │
│    CheckApplyLock()                  │
└──────────────────────────────────────┘
        │
        ├─ [LOCKED] → Error: "Apply commands are locked globally"
        │
        ▼ [UNLOCKED]
┌──────────────────────────────────────┐
│ 2. ACQUIRE WORKING DIR LOCK          │
│    (Same as plan flow)               │
└──────────────────────────────────────┘
        │
        ▼
┌──────────────────────────────────────┐
│ 3. VERIFY PROJECT LOCK EXISTS        │
│    Must have prior plan              │
│    Check lock held by this PR        │
└──────────────────────────────────────┘
        │
        ├─ [NO LOCK] → Error: "No plan exists"
        ├─ [LOCKED BY DIFFERENT PR] → Error: "Plan from different PR"
        │
        ▼ [VERIFIED]
┌──────────────────────────────────────┐
│ 4. EXECUTE TERRAFORM APPLY           │
│    (All locks verified/held)         │
└──────────────────────────────────────┘
        │
        ▼
┌──────────────────────────────────────┐
│ 5. RELEASE WORKING DIR LOCK          │
│    unlockFn() via defer              │
└──────────────────────────────────────┘
        │
        ▼
┌──────────────────────────────────────┐
│ 6. RELEASE PROJECT LOCK              │
│    On successful apply:              │
│    - Delete plan file                │
│    - Unlock project                  │
│    - Update pull status              │
└──────────────────────────────────────┘
```

### Lock Acquisition Order (Deadlock Prevention)

**REQUIRED ORDER**:
```
1. Global Command Lock (check only)
   ↓
2. Working Directory Lock
   ↓
3. Project Lock
```

**Rationale**:
- Consistent ordering prevents circular wait
- Working dir lock is always acquired first in execution flow
- Project lock checked/acquired after working dir secured
- Global lock is checked before any local locks

---

## Architecture Deep Dive

### Database Layer

#### BoltDB Implementation

**File**: `/server/core/boltdb/boltdb.go`

**Buckets**:
```go
const (
    locksBucketName       = "runLocks"       // Project locks
    pullsBucketName       = "pulls"          // Pull request status
    globalLocksBucketName = "globalLocks"    // Command locks
)
```

**Transaction Pattern** (lines 82-117):
```go
func (b *BoltDB) TryLock(newLock models.ProjectLock) (bool, models.ProjectLock, error) {
    var lockAcquired bool
    var currLock models.ProjectLock
    key := b.lockKey(newLock.Project, newLock.Workspace)
    newLockSerialized, _ := json.Marshal(newLock)

    transactionErr := b.db.Update(func(tx *bolt.Tx) error {
        bucket := tx.Bucket(b.locksBucketName)

        // Atomic check
        currLockSerialized := bucket.Get([]byte(key))
        if currLockSerialized == nil {
            // Lock available - acquire atomically
            bucket.Put([]byte(key), newLockSerialized)
            lockAcquired = true
            currLock = newLock
            return nil
        }

        // Lock exists - deserialize current lock
        if err := json.Unmarshal(currLockSerialized, &currLock); err != nil {
            return errors.Wrap(err, "failed to deserialize lock")
        }
        lockAcquired = false
        return nil
    })

    return lockAcquired, currLock, transactionErr
}
```

**ACID Guarantees**:
- ✅ Atomicity: Check-and-set in single transaction
- ✅ Consistency: JSON schema validation
- ✅ Isolation: Serializable (BoltDB file lock)
- ✅ Durability: Fsync on commit

#### Redis Implementation

**File**: `/server/core/redis/redis.go`

**Key Pattern**: `locks:{repoFullName}/{path}/{workspace}`

**Atomic Operations**:
```go
func (r *RedisDB) TryLock(lock models.ProjectLock) (bool, models.ProjectLock, error) {
    key := r.lockKey(lock.Project, lock.Workspace)

    // Check if lock exists
    val, err := r.client.Get(ctx, key).Result()
    if err == redis.Nil {
        // Lock doesn't exist - create it
        serialized, _ := json.Marshal(lock)
        err := r.client.Set(ctx, key, serialized, 0).Err()
        return true, lock, err
    }

    // Lock exists - return current lock
    var currLock models.ProjectLock
    json.Unmarshal([]byte(val), &currLock)
    return false, currLock, nil
}
```

**Distributed Characteristics**:
- ✅ Multi-server support
- ✅ Network-based coordination
- ✅ Optional TLS encryption
- ⚠️ Network latency added

### Concurrency Patterns

#### Mutex Guard Pattern (Working Dir Locks)

```go
type DefaultWorkingDirLocker struct {
    mutex sync.Mutex
    locks map[string]struct{}
}

func (d *DefaultWorkingDirLocker) TryLock(...) (func(), error) {
    d.mutex.Lock()
    defer d.mutex.Unlock()  // Protects critical section

    // Atomic check-and-set
    key := d.workspaceKey(...)
    if _, exists := d.locks[key]; exists {
        return func() {}, errors.New("locked")
    }

    d.locks[key] = struct{}{}

    // Return closure for cleanup
    return func() {
        d.unlock(...)
    }, nil
}
```

**Best Practices Demonstrated**:
- ✅ Defer unlock in critical section
- ✅ Minimal lock hold time
- ✅ Atomic operations
- ✅ Closure-based cleanup

#### Database Transaction Pattern (Project Locks)

```go
// Optimistic locking with check-then-set
err := db.Update(func(tx *bolt.Tx) error {
    // Read current state
    currentValue := bucket.Get(key)

    // Make decision based on current state
    if currentValue == nil {
        // Modify state atomically
        bucket.Put(key, newValue)
    }

    return nil
})
```

**Advantages**:
- ✅ No external coordination needed
- ✅ Database ensures serialization
- ✅ ACID guarantees
- ✅ Works across process restarts

---

## Historical Context & Evolution

### PR #3345: Architecture Decision Record

**Status**: Open (Draft)
**Created**: April 21, 2023
**Author**: GenPage

**Problem Statement**:
> Current locking is "not scalable for repos with many modules and concurrent users"

**Root Cause**:
- System clones repositories **3+ times per command execution**
- Two separate lock types with insufficient coordination
- Workspace-level locking prevents true parallel execution

**Proposed Solution**:
1. **Clone Once Strategy**: Clone repo only once per PR
2. **TF_DATA_DIR Separation**: Use Terraform's `TF_DATA_DIR` for workspace isolation
3. **Lock Simplification**: Reduce to repository/PR level locking

**Status**: Architecture decision pending community consensus

### Known Issues

#### Issue #5722: Project Name Lock Isolation

**Problem**: Projects with same path but different names share locks

**Example**:
```yaml
projects:
  - name: app-dev
    dir: .
    workflow: dev

  - name: app-prod
    dir: .
    workflow: prod
```

Both share lock key: `owner/repo/./default`

**Workaround**: Use different directories or workspaces

**Long-term Fix**: Include project name in lock key (breaking change)

#### Issue #2200: Default Workspace Bottleneck

**Problem**: All projects without explicit workspace use `default`, creating serial execution

**Impact**: 35 👍 reactions (highest engagement)

**Lock Collision**:
```
terraform/region/dev → owner/repo/terraform/region/dev/default
terraform/apps/api  → owner/repo/terraform/apps/api/default
```

Both locked by workspace name `default`.

**Workaround**: Use explicit workspace names per project

### Recent Improvements

#### PR #4192: Add ProjectName to Lock Metadata

**Change**: Added `ProjectName` field to `ProjectLock` struct

**Impact**:
- Lock still uses same key format (backward compatible)
- Plan file naming improved
- Unlock operations can match by name
- Full isolation not achieved (awaiting ADR #3345)

#### PR #5790: Working Directory Lock Path Fix

**Change**: Added `path` parameter to working directory lock key

**Before**: `{repo}/{pullNum}/{workspace}`
**After**: `{repo}/{pullNum}/{workspace}/{path}`

**Impact**:
- Enables parallel execution of different projects
- Prevents false lock collisions
- Maintains filesystem safety

---

## Troubleshooting Guide

### Common Error Messages

#### 1. "Workspace is currently locked by another command"

**Lock Type**: Working Directory Lock
**Meaning**: Another command is executing in the same workspace
**Duration**: Seconds to minutes (in-memory)

**Troubleshooting**:
```bash
# Check if command is still running
ps aux | grep terraform

# Wait for command to complete (typical: <5 min)
# OR restart Atlantis (clears in-memory locks)
systemctl restart atlantis
```

**Causes**:
- User ran multiple commands quickly
- Long-running plan/apply
- Stuck command (rare)

**Resolution**:
- ✅ Wait for completion
- ✅ Restart Atlantis (if stuck >10 min)
- ❌ DON'T disable locking

---

#### 2. "Project locked by pull #X"

**Lock Type**: Project Lock
**Meaning**: Another PR has an unapplied plan for this project
**Duration**: Hours to days (database-persisted)

**Troubleshooting**:
```bash
# View all locks
curl https://atlantis.example.com/locks

# Check specific lock
curl https://atlantis.example.com/locks?repo=owner/repo

# Unlock via API
curl -X DELETE https://atlantis.example.com/locks/{lockID}
```

**Causes**:
- PR #X has pending plan
- PR #X merged but plan not applied
- Orphaned lock from closed PR

**Resolution**:
1. **Apply the blocking PR**: `atlantis apply` on PR #X
2. **Unlock manually**: `atlantis unlock -r owner/repo -p path -w workspace`
3. **Delete via UI**: Navigate to locks page, click delete

---

#### 3. "Apply commands are locked globally"

**Lock Type**: Apply Command Lock
**Meaning**: Global apply lock is enabled
**Duration**: Until manually unlocked

**Troubleshooting**:
```bash
# Check apply lock status
curl https://atlantis.example.com/lock/apply/status

# Unlock
curl -X DELETE https://atlantis.example.com/lock/apply
```

**Causes**:
- Admin enabled via `--disable-apply` flag
- Global lock set for maintenance
- Configuration: `AllowCommands` excludes "apply"

**Resolution**:
1. **Remove flag**: Restart without `--disable-apply`
2. **Unlock API**: `DELETE /lock/apply`
3. **Update config**: Ensure `apply` in `AllowCommands`

---

### Diagnostic Queries

#### BoltDB Lock Inspection

```bash
# Install bolt CLI
go install go.etcd.io/bbolt/cmd/bolt@latest

# List all locks
bolt get /var/atlantis/data/atlantis.db runLocks

# Find locks for specific repo
bolt get /var/atlantis/data/atlantis.db runLocks | grep "owner/repo"
```

#### Redis Lock Inspection

```bash
# List all lock keys
redis-cli KEYS "locks:*"

# Get specific lock
redis-cli GET "locks:owner/repo/path/workspace"

# Count total locks
redis-cli KEYS "locks:*" | wc -l
```

#### Programmatic Lock Listing

```go
// List all locks via API
locker := locking.NewClient(database)
locks, err := locker.List()

for key, lock := range locks {
    fmt.Printf("Lock: %s\n", key)
    fmt.Printf("  PR: #%d (%s)\n", lock.Pull.Num, lock.Pull.URL)
    fmt.Printf("  User: %s\n", lock.User.Username)
    fmt.Printf("  Time: %s\n", lock.Time)
}
```

---

## Best Practices

### DO: Use Defer for Working Dir Locks

```go
✅ CORRECT:
unlockFn, err := workingDirLocker.TryLock(repo, pull, workspace, path)
if err != nil {
    return err
}
defer unlockFn()  // Guaranteed release

// Protected operations
workingDir.Clone(...)
terraform.Plan(...)
```

```go
❌ INCORRECT:
unlockFn, err := workingDirLocker.TryLock(repo, pull, workspace, path)
// ... do work ...
unlockFn()  // May not execute if error occurs
```

**Rationale**: Defer ensures unlock even on panic or early return

---

### DO: Check Lock Acquisition Before Proceeding

```go
✅ CORRECT:
resp, err := projectLocker.TryLock(log, pull, user, workspace, project, repoLocking)
if err != nil {
    return err
}
if !resp.LockAcquired {
    return errors.New(resp.LockFailureReason)
}
// Proceed with locked operation
```

```go
❌ INCORRECT:
resp, _ := projectLocker.TryLock(...)
// Proceed without checking LockAcquired
```

**Rationale**: Prevents execution without proper lock protection

---

### DO: Provide Context in Errors

```go
✅ CORRECT:
return errors.Wrapf(err,
    "failed to acquire project lock for %s/%s",
    project.RepoFullName, workspace)
```

```go
❌ INCORRECT:
return err
```

**Rationale**: Easier debugging, better user experience

---

### DON'T: Hold Locks Across Network Calls

```go
✅ CORRECT:
// Acquire lock
unlockFn, _ := locker.TryLock(...)
defer unlockFn()

// Do local work
localComputation()

// Release lock (defer)
// THEN make network call
makeAPICall()
```

```go
❌ INCORRECT:
unlockFn, _ := locker.TryLock(...)
defer unlockFn()

// Slow network call while holding lock
makeSlowAPICall()  // Blocks other requests

localComputation()
```

**Rationale**: Minimizes lock hold time, reduces contention

---

### DON'T: Disable Locking Without Understanding Impact

```yaml
❌ INCORRECT:
repos:
  - id: /.*/
    repo_locking: false  # Disables ALL project locks
```

**Consequences**:
- Multiple PRs can plan same project simultaneously
- "Last plan wins" scenario
- Potential state corruption
- Lost plan context

**Correct Use Case**:
```yaml
✅ CORRECT (for automation repos only):
repos:
  - id: /.*-automation$/
    repo_locking: false  # Safe: automation handles locking externally
```

---

### DO: Clean Up Locks on PR Close

**Automatic** (recommended):
```yaml
# Ensure PR close webhook is configured
webhooks:
  - event: pull_request
    actions: [closed]
```

**Manual** (if webhooks unavailable):
```bash
# Periodic cleanup script
atlantis unlock --all-closed-prs
```

**Rationale**: Prevents orphaned locks from blocking future PRs

---

## Developer Reference

### Code Navigation

#### Core Interfaces

**Locker Interface** (`/server/core/locking/locking.go:45-51`):
```go
type Locker interface {
    TryLock(p models.Project, workspace string, pull models.PullRequest, user models.User) (TryLockResponse, error)
    Unlock(key string) (*models.ProjectLock, error)
    List() (map[string]models.ProjectLock, error)
    UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error)
    GetLock(key string) (*models.ProjectLock, error)
}
```

**WorkingDirLocker Interface** (`/server/events/working_dir_locker.go:28-34`):
```go
type WorkingDirLocker interface {
    TryLock(repoFullName string, pullNum int, workspace string, path string) (func(), error)
}
```

**ApplyLocker Interface** (`/server/core/locking/apply_locking.go:22-30`):
```go
type ApplyLocker interface {
    LockApply() (ApplyCommandLock, error)
    UnlockApply() error
    ApplyLockChecker
}
```

#### Database Interface

**Database Lock Methods** (`/server/core/db/db.go:28-45`):
```go
type Database interface {
    // Project Locks
    TryLock(lock models.ProjectLock) (bool, models.ProjectLock, error)
    Unlock(project models.Project, workspace string) (*models.ProjectLock, error)
    GetLock(project models.Project, workspace string) (*models.ProjectLock, error)
    UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error)
    List() ([]models.ProjectLock, error)

    // Command Locks
    LockCommand(cmdName command.Name, lockTime time.Time) (*command.Lock, error)
    UnlockCommand(cmdName command.Name) error
    CheckCommandLock(cmdName command.Name) (*command.Lock, error)
}
```

### Testing Patterns

#### Unit Test Example

```go
func TestTryLock_Success(t *testing.T) {
    locker := events.NewDefaultWorkingDirLocker()

    // First lock should succeed
    unlockFn, err := locker.TryLock("owner/repo", 1, "default", ".")
    assert.NoError(t, err)

    // Second attempt should fail
    _, err = locker.TryLock("owner/repo", 1, "default", ".")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "currently locked")

    // After unlock, should succeed again
    unlockFn()
    _, err = locker.TryLock("owner/repo", 1, "default", ".")
    assert.NoError(t, err)
}
```

#### Integration Test Example

```go
func TestBoltDB_TryLock_Concurrent(t *testing.T) {
    db := newTestBoltDB(t)
    defer db.Close()

    project := models.NewProject("owner/repo", "path")

    // Acquire lock
    acquired, _, err := db.TryLock(models.ProjectLock{
        Project: project,
        Workspace: "default",
        Pull: pull1,
    })
    assert.NoError(t, err)
    assert.True(t, acquired)

    // Different PR attempts lock
    acquired, currLock, err := db.TryLock(models.ProjectLock{
        Project: project,
        Workspace: "default",
        Pull: pull2,
    })
    assert.NoError(t, err)
    assert.False(t, acquired)
    assert.Equal(t, pull1.Num, currLock.Pull.Num)
}
```

### Adding Custom Locking Logic

```go
// Custom locker wrapper
type MyCustomLocker struct {
    underlying locking.Locker
    metrics    MetricsCollector
}

func (m *MyCustomLocker) TryLock(
    p models.Project,
    workspace string,
    pull models.PullRequest,
    user models.User,
) (locking.TryLockResponse, error) {
    start := time.Now()

    // Call underlying locker
    resp, err := m.underlying.TryLock(p, workspace, pull, user)

    // Collect metrics
    m.metrics.RecordLockAttempt(
        time.Since(start),
        resp.LockAcquired,
        p.RepoFullName,
    )

    // Custom logic
    if !resp.LockAcquired {
        m.sendLockContentionAlert(p, workspace, resp.CurrLock)
    }

    return resp, err
}
```

### Configuration Example

```yaml
# atlantis.yaml
version: 3
projects:
  - name: infrastructure
    dir: terraform/infra
    workspace: production
    autoplan:
      when_modified: ["*.tf"]

  - name: application
    dir: terraform/app
    workspace: production
    autoplan:
      when_modified: ["*.tf"]

# Both projects will have separate locks:
# Lock 1: owner/repo/terraform/infra/production
# Lock 2: owner/repo/terraform/app/production
```

---

## Appendix

### Lock File Locations

| Component | File Path | Lines |
|-----------|-----------|-------|
| Locker Interface | `/server/core/locking/locking.go` | 45-51 |
| Client Implementation | `/server/core/locking/locking.go` | 39-139 |
| NoOpLocker | `/server/core/locking/locking.go` | 141-183 |
| ApplyLocker | `/server/core/locking/apply_locking.go` | 22-122 |
| ProjectLocker | `/server/events/project_locker.go` | 29-96 |
| WorkingDirLocker | `/server/events/working_dir_locker.go` | 28-78 |
| BoltDB Implementation | `/server/core/boltdb/boltdb.go` | 82-313 |
| Redis Implementation | `/server/core/redis/redis.go` | Similar |
| ProjectLock Model | `/server/events/models/models.go` | 236-289 |

### Related PRs and Issues

| Type | Number | Title | Status |
|------|--------|-------|--------|
| ADR | #3345 | Locking Architecture Decision | Open (Draft) |
| PR | #4192 | Add ProjectName to locks | Merged (v0.35.0) |
| PR | #5790 | Add path to working dir locks | Merged |
| Issue | #5722 | Lock isolation by project name | Open |
| Issue | #2200 | Default workspace bottleneck | Open |

---

## Summary

The Atlantis locking system successfully balances **concurrent execution** with **data integrity** through a three-tiered architecture:

1. **Global Command Locks** - Emergency controls and maintenance windows
2. **Project Locks** - Cross-PR coordination and plan protection
3. **Working Directory Locks** - Filesystem safety and race prevention

**Key Strengths**:
- ✅ Multiple granularity levels for different use cases
- ✅ Database persistence for critical locks
- ✅ In-memory speed for transient locks
- ✅ Flexible backends (BoltDB, Redis)
- ✅ Clear error messages with actionable guidance

**Known Limitations**:
- ⚠️ Project name not in lock key (#5722)
- ⚠️ Default workspace creates bottleneck (#2200)
- ⚠️ Multiple clones impact performance (ADR #3345)

**Recommended Actions**:
1. Configure explicit workspaces to avoid default bottleneck
2. Monitor lock metrics and clean up orphaned locks
3. Use Redis for multi-server deployments
4. Follow PR #3345 ADR for future improvements

This documentation serves as the authoritative reference for understanding, operating, and extending the Atlantis locking mechanisms.
