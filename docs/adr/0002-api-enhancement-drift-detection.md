# 2. API Enhancement and Drift Detection

Date: 2025-01-21

## Status

Proposed

## Context

The current Atlantis API (`/api/plan`, `/api/apply`, `/api/locks`) provides basic functionality for triggering Terraform operations but has several issues that need addressing:

### Current Issues

1. **Error Serialization Bug**: Go `error` types serialize as `{}` or `null` in JSON responses instead of the actual error message
2. **Non-PR Workflow Limitations**: API calls without PR context fail `approved` and `mergeable` requirement checks, preventing drift detection use cases
3. **No Drift Detection**: Cannot detect infrastructure drift outside of PR workflows
4. **API Stability**: Current endpoints are marked as ALPHA and subject to change
5. **Limited Features**: Missing endpoints for projects, runs, and operational tasks

### Analysis of PRs #6093 and #6095

**PR #6093 (Error Serialization):**
- Goal: Fix error fields serializing as `{}` instead of error messages
- Issue: The implementation changes `ProjectResult` from embedding `ProjectCommandOutput` to a named field
- Problem: This **breaks the API response structure** by wrapping fields in `"ProjectCommandOutput": {...}`
- Correct approach: Use custom `MarshalJSON` on `ProjectResult` that handles the embedded struct while converting errors to strings

**PR #6095 (API Non-PR Support):**
- Goal: Allow API applies without PR to bypass `approved`/`mergeable` requirements
- Approach: Add `API` flag to `ProjectContext` and skip PR-specific requirements when `ctx.API && ctx.Pull.Num == 0`
- Status: Good approach, but should be integrated into a larger design

### Current API Endpoints

| Endpoint | Method | Auth Required | Description |
|----------|--------|---------------|-------------|
| `/api/plan` | POST | Yes | Trigger terraform plan |
| `/api/apply` | POST | Yes | Trigger terraform apply |
| `/api/locks` | GET | No | List active locks |
| `/status` | GET | No | Server status |
| `/healthz` | GET | No | Health check |

### Current API Response Structure (must be preserved)

```json
{
  "Error": null,
  "Failure": "",
  "ProjectResults": [
    {
      "Error": null,           // BUG: Shows {} when error exists
      "Failure": "",
      "PlanSuccess": {...},
      "PolicyCheckResults": null,
      "ApplySuccess": "",
      "VersionSuccess": "",
      "ImportSuccess": null,
      "StateRmSuccess": null,
      "Command": 1,
      "SubCommand": "",
      "RepoRelDir": ".",
      "Workspace": "default",
      "ProjectName": "",
      "SilencePRComments": null
    }
  ],
  "PlansDeleted": false
}
```

## Decision

We will enhance the Atlantis API in five phases, implementing the fixes from PRs #6093 and #6095 properly as part of a comprehensive design.

---

### Phase 1: Error Serialization Fix (Backwards Compatible)

**Goal**: Fix error serialization WITHOUT changing the API response structure.

**Problem Analysis**:
The current `ProjectResult` struct embeds `ProjectCommandOutput`:
```go
type ProjectResult struct {
    ProjectCommandOutput  // embedded
    Command           Name
    // ...
}
```

PR #6093's approach changes this to a named field, which changes JSON output from flat fields to nested `ProjectCommandOutput: {...}` - a breaking change.

**Correct Implementation**:

Add `MarshalJSON` to `ProjectResult` that:
1. Keeps the embedded struct (no structural changes)
2. Manually serializes fields with error converted to string
3. Maintains exact same JSON structure

```go
// server/events/command/project_result.go

// MarshalJSON implements custom JSON marshaling to properly serialize the Error field
// while maintaining backwards compatibility with the existing API response structure.
func (p ProjectResult) MarshalJSON() ([]byte, error) {
    // Convert error to string pointer for proper serialization
    var errMsg *string
    if p.Error != nil {
        msg := p.Error.Error()
        errMsg = &msg
    }

    // Use anonymous struct to control JSON output
    // This preserves the flat structure (no ProjectCommandOutput wrapper)
    return json.Marshal(&struct {
        Error              *string                    `json:"Error"`
        Failure            string                     `json:"Failure"`
        PlanSuccess        *models.PlanSuccess        `json:"PlanSuccess"`
        PolicyCheckResults *models.PolicyCheckResults `json:"PolicyCheckResults"`
        ApplySuccess       string                     `json:"ApplySuccess"`
        VersionSuccess     string                     `json:"VersionSuccess"`
        ImportSuccess      *models.ImportSuccess      `json:"ImportSuccess"`
        StateRmSuccess     *models.StateRmSuccess     `json:"StateRmSuccess"`
        Command            Name                       `json:"Command"`
        SubCommand         string                     `json:"SubCommand"`
        RepoRelDir         string                     `json:"RepoRelDir"`
        Workspace          string                     `json:"Workspace"`
        ProjectName        string                     `json:"ProjectName"`
        SilencePRComments  []string                   `json:"SilencePRComments"`
    }{
        Error:              errMsg,
        Failure:            p.Failure,
        PlanSuccess:        p.PlanSuccess,
        PolicyCheckResults: p.PolicyCheckResults,
        ApplySuccess:       p.ApplySuccess,
        VersionSuccess:     p.VersionSuccess,
        ImportSuccess:      p.ImportSuccess,
        StateRmSuccess:     p.StateRmSuccess,
        Command:            p.Command,
        SubCommand:         p.SubCommand,
        RepoRelDir:         p.RepoRelDir,
        Workspace:          p.Workspace,
        ProjectName:        p.ProjectName,
        SilencePRComments:  p.SilencePRComments,
    })
}

// MarshalJSON for Result to properly serialize top-level Error
func (c Result) MarshalJSON() ([]byte, error) {
    type Alias Result
    var errMsg *string
    if c.Error != nil {
        msg := c.Error.Error()
        errMsg = &msg
    }
    return json.Marshal(&struct {
        Error *string `json:"Error"`
        *Alias
    }{
        Error: errMsg,
        Alias: (*Alias)(&c),
    })
}
```

**Why This Approach**:
1. **No structural changes**: Keeps embedded struct, no changes to how code accesses fields
2. **Backwards compatible**: JSON output structure remains identical
3. **Explicit field listing**: The reviewer's concern about "manually listing all 14 fields" is valid - add a comment warning future maintainers
4. **Testable**: Easy to verify JSON output matches expected format

**Files to Modify**:
- `server/events/command/project_result.go` - Add `MarshalJSON` methods
- `server/events/command/result.go` - Add `MarshalJSON` method
- `server/events/command/result_test.go` - Add comprehensive tests

---

### Phase 2: Non-PR Workflow Support (Foundation for Drift Detection)

**Goal**: Allow API operations without PR context for drift detection use cases.

**Implementation** (based on PR #6095 with enhancements):

1. **Add `API` flag to `ProjectContext`**:
```go
// server/events/command/project_context.go
type ProjectContext struct {
    // ... existing fields ...

    // API is true if this command was triggered via API endpoint rather than PR comment.
    // When true and Pull.Num == 0, PR-specific requirements (approved, mergeable) are skipped.
    API bool
}
```

2. **Propagate API flag in command builder**:
```go
// server/events/project_command_context_builder.go
func newProjectCommandContext(...) command.ProjectContext {
    return command.ProjectContext{
        // ... existing fields ...
        API: ctx.API,
    }
}
```

3. **Skip PR-specific requirements for non-PR API calls**:
```go
// server/events/command_requirement_handler.go
func (a *DefaultCommandRequirementHandler) validateCommandRequirement(...) (string, error) {
    for _, req := range requirements {
        switch req {
        case raw.ApprovedRequirement:
            // Skip for API calls without PR - approval is PR-specific
            if ctx.API && ctx.Pull.Num == 0 {
                ctx.Log.Info("skipping approval requirement for API call without PR number")
                continue
            }
            // ... existing approval check ...

        case raw.MergeableRequirement:
            // Skip for API calls without PR - mergeability is PR-specific
            if ctx.API && ctx.Pull.Num == 0 {
                ctx.Log.Info("skipping mergeable requirement for API call without PR number")
                continue
            }
            // ... existing mergeable check ...
        }
    }
}
```

**Design Rationale**:
- `approved` and `mergeable` requirements are **PR-specific checks** - they have no meaning outside a PR context
- Non-PR API calls (drift detection, emergency applies) should still respect other requirements like `undiverged`
- API calls WITH a PR number should still enforce all requirements (maintains security for PR-based workflows)

**Files to Modify**:
- `server/events/command/project_context.go` - Add `API` field
- `server/events/project_command_context_builder.go` - Propagate `API` flag
- `server/events/command_requirement_handler.go` - Skip PR-specific requirements
- `server/events/command_requirement_handler_test.go` - Add test cases

---

### Phase 3: Drift Detection Endpoints

**Goal**: Enable infrastructure drift detection outside of PR workflows.

**New Endpoints**:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `POST /api/plan` | Enhanced | Support `PR: 0` or omitted for drift detection |
| `GET /api/drift/status` | New | Get drift status for repository/projects |

**Enhanced Plan Endpoint for Drift Detection**:

The existing `/api/plan` endpoint already supports omitting the `PR` field. With Phase 2's non-PR workflow support, it becomes usable for drift detection:

```bash
# Drift detection request (no PR)
curl -X POST 'https://atlantis/api/plan' \
  -H 'X-Atlantis-Token: <secret>' \
  -H 'Content-Type: application/json' \
  -d '{
    "Repository": "org/repo",
    "Ref": "main",
    "Type": "Github",
    "Paths": [{"Directory": ".", "Workspace": "default"}]
  }'
```

**Response Enhancement for Drift Detection**:

Add drift detection helpers to parse plan output:

```go
// server/events/models/drift.go

// DriftSummary represents detected infrastructure drift
type DriftSummary struct {
    HasDrift    bool   `json:"has_drift"`
    ToAdd       int    `json:"to_add"`
    ToChange    int    `json:"to_change"`
    ToDestroy   int    `json:"to_destroy"`
    Summary     string `json:"summary"`       // e.g., "2 to add, 1 to change, 0 to destroy"
}

// ParseDriftFromPlan extracts drift information from terraform plan output
func ParseDriftFromPlan(planOutput string) DriftSummary {
    // Parse lines like:
    // "Plan: 2 to add, 1 to change, 0 to destroy."
    // or "No changes. Your infrastructure matches the configuration."
    // ...
}
```

**New Drift Status Endpoint**:

```go
// GET /api/drift/status?repository=org/repo&project=myproject
type DriftStatusResponse struct {
    Repository string        `json:"repository"`
    Projects   []ProjectDrift `json:"projects"`
    CheckedAt  time.Time     `json:"checked_at"`
}

type ProjectDrift struct {
    ProjectName string       `json:"project_name"`
    Path        string       `json:"path"`
    Workspace   string       `json:"workspace"`
    Drift       DriftSummary `json:"drift"`
    LastChecked time.Time    `json:"last_checked"`
}
```

**Drift Storage** (optional, for caching drift status):

```go
// server/core/drift/storage.go

type DriftStorage interface {
    StoreDriftStatus(repo string, project string, drift DriftSummary) error
    GetDriftStatus(repo string, project string) (*ProjectDrift, error)
    ListDriftStatus(repo string) ([]ProjectDrift, error)
}

// In-memory implementation for initial version
type InMemoryDriftStorage struct {
    mu     sync.RWMutex
    status map[string]map[string]ProjectDrift // repo -> project -> drift
}
```

**Files to Create/Modify**:
- `server/events/models/drift.go` - New drift models
- `server/core/drift/parser.go` - Plan output parser for drift detection
- `server/core/drift/storage.go` - Optional drift status storage
- `server/controllers/api_controller.go` - Add drift status endpoint

---

### Phase 4: Drift Remediation (PR/MR Creation)

**Goal**: Enable automated PR/MR creation to consolidate drift fixes.

**Prerequisites**:
- VCS Client interface must support PR/MR creation
- Currently, PR creation only exists in E2E test code

**VCS Client Extension**:

```go
// server/events/vcs/client.go

type Client interface {
    // ... existing methods ...

    // CreatePullRequest creates a new PR/MR for the repository.
    // Returns the created pull request or an error.
    CreatePullRequest(
        logger logging.SimpleLogging,
        repo models.Repo,
        opts CreatePullRequestOptions,
    ) (models.PullRequest, error)

    // CreateBranch creates a new branch from the given ref.
    CreateBranch(
        logger logging.SimpleLogging,
        repo models.Repo,
        branchName string,
        fromRef string,
    ) error
}

type CreatePullRequestOptions struct {
    Title       string
    Body        string
    HeadBranch  string   // Source branch (e.g., "drift-remediation-2025-01-21")
    BaseBranch  string   // Target branch (e.g., "main")
    Draft       bool
    Labels      []string
    Assignees   []string
}
```

**GitHub Implementation**:

```go
// server/events/vcs/github/client.go

func (g *GithubClient) CreatePullRequest(
    logger logging.SimpleLogging,
    repo models.Repo,
    opts vcs.CreatePullRequestOptions,
) (models.PullRequest, error) {
    owner, repoName := repo.Owner, repo.Name

    pr, _, err := g.client.PullRequests.Create(g.ctx, owner, repoName, &github.NewPullRequest{
        Title: github.String(opts.Title),
        Body:  github.String(opts.Body),
        Head:  github.String(opts.HeadBranch),
        Base:  github.String(opts.BaseBranch),
        Draft: github.Bool(opts.Draft),
    })
    if err != nil {
        return models.PullRequest{}, err
    }

    return models.PullRequest{
        Num:        pr.GetNumber(),
        HeadBranch: pr.GetHead().GetRef(),
        BaseBranch: pr.GetBase().GetRef(),
        URL:        pr.GetHTMLURL(),
        State:      models.OpenPullState,
        // ... other fields
    }, nil
}

func (g *GithubClient) CreateBranch(
    logger logging.SimpleLogging,
    repo models.Repo,
    branchName string,
    fromRef string,
) error {
    owner, repoName := repo.Owner, repo.Name

    // Get the SHA of the source ref
    ref, _, err := g.client.Git.GetRef(g.ctx, owner, repoName, "refs/heads/"+fromRef)
    if err != nil {
        return fmt.Errorf("failed to get ref %s: %w", fromRef, err)
    }

    // Create new branch
    _, _, err = g.client.Git.CreateRef(g.ctx, owner, repoName, &github.Reference{
        Ref:    github.String("refs/heads/" + branchName),
        Object: &github.GitObject{SHA: ref.Object.SHA},
    })
    if err != nil {
        return fmt.Errorf("failed to create branch %s: %w", branchName, err)
    }

    return nil
}
```

**GitLab Implementation**:

```go
// server/events/vcs/gitlab/client.go

func (g *GitlabClient) CreatePullRequest(
    logger logging.SimpleLogging,
    repo models.Repo,
    opts vcs.CreatePullRequestOptions,
) (models.PullRequest, error) {
    mr, _, err := g.Client.MergeRequests.CreateMergeRequest(repo.FullName, &gitlab.CreateMergeRequestOptions{
        Title:        gitlab.String(opts.Title),
        Description:  gitlab.String(opts.Body),
        SourceBranch: gitlab.String(opts.HeadBranch),
        TargetBranch: gitlab.String(opts.BaseBranch),
        Labels:       &gitlab.Labels{opts.Labels...},
    })
    if err != nil {
        return models.PullRequest{}, err
    }

    return models.PullRequest{
        Num:        mr.IID,
        HeadBranch: mr.SourceBranch,
        BaseBranch: mr.TargetBranch,
        URL:        mr.WebURL,
        State:      models.OpenPullState,
    }, nil
}
```

**Drift Remediation Endpoint**:

```go
// POST /api/drift/remediate
type DriftRemediateRequest struct {
    Repository  string   `json:"repository" validate:"required"`
    Ref         string   `json:"ref" validate:"required"`          // Base branch
    Type        string   `json:"type" validate:"required"`         // VCS type
    Projects    []string `json:"projects,omitempty"`               // Projects to remediate
    Paths       []Path   `json:"paths,omitempty"`
    Title       string   `json:"title,omitempty"`                  // PR title
    Description string   `json:"description,omitempty"`            // PR body
    Draft       bool     `json:"draft,omitempty"`
    AutoMerge   bool     `json:"auto_merge,omitempty"`             // Enable auto-merge after approval
}

type DriftRemediateResponse struct {
    Success     bool               `json:"success"`
    PullRequest *PullRequestInfo   `json:"pull_request,omitempty"`
    Error       string             `json:"error,omitempty"`
}

type PullRequestInfo struct {
    Number  int    `json:"number"`
    URL     string `json:"url"`
    Branch  string `json:"branch"`
    Title   string `json:"title"`
}
```

**Remediation Workflow**:
1. Run plan for each specified project (detect what needs to be applied)
2. Create a new branch (e.g., `atlantis-drift-remediation-{timestamp}`)
3. Create PR/MR with drift summary in description
4. Return PR details for user to review and merge

**Files to Create/Modify**:
- `server/events/vcs/client.go` - Add `CreatePullRequest`, `CreateBranch` interface methods
- `server/events/vcs/github/client.go` - Implement for GitHub
- `server/events/vcs/gitlab/client.go` - Implement for GitLab
- `server/events/vcs/bitbucketcloud/client.go` - Implement for Bitbucket
- `server/events/vcs/azuredevops/client.go` - Implement for Azure DevOps
- `server/controllers/api_controller.go` - Add remediation endpoint

---

### Phase 5: Additional API Enhancements

**Goal**: Add useful operational endpoints and improve API stability.

**New Endpoints**:

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `GET /api/projects` | GET | Yes | List configured projects |
| `GET /api/runs` | GET | Yes | List recent plan/apply runs |
| `DELETE /api/locks/{id}` | DELETE | Yes | Force unlock a specific lock |

**Projects Endpoint**:

```go
// GET /api/projects?repository=org/repo
type ProjectsResponse struct {
    Projects []ProjectInfo `json:"projects"`
}

type ProjectInfo struct {
    Name         string `json:"name"`
    Repository   string `json:"repository"`
    Path         string `json:"path"`
    Workspace    string `json:"workspace"`
    Workflow     string `json:"workflow,omitempty"`
    AutoplanEnabled bool `json:"autoplan_enabled"`
}
```

**Force Unlock Endpoint**:

```go
// DELETE /api/locks/{id}
// Requires X-Atlantis-Token authentication
func (a *APIController) ForceUnlock(w http.ResponseWriter, r *http.Request) {
    // Extract lock ID from path
    // Validate authentication
    // Call a.Locker.Unlock(lockID)
    // Return success/failure
}
```

---

## Implementation Plan

### Stage 1: Error Serialization (1-2 weeks)
- [ ] Add `MarshalJSON` to `ProjectResult` with explicit field listing
- [ ] Add `MarshalJSON` to `Result`
- [ ] Add comprehensive unit tests for JSON serialization
- [ ] Add integration tests verifying API response format unchanged
- [ ] Add maintainer warning comment about field synchronization

### Stage 2: Non-PR Workflow Support (1-2 weeks)
- [ ] Add `API` field to `ProjectContext`
- [ ] Propagate `API` flag in command builder
- [ ] Modify requirement handler to skip PR-specific requirements
- [ ] Add unit tests for all combinations (API/non-API, with/without PR)
- [ ] Add integration tests for drift detection workflow

### Stage 3: Drift Detection (2-3 weeks)
- [ ] Create drift models and parser
- [ ] Add `DriftSummary` to plan response (optional enhancement)
- [ ] Implement drift status storage (in-memory)
- [ ] Add `/api/drift/status` endpoint
- [ ] Add documentation for drift detection usage

### Stage 4: Drift Remediation (3-4 weeks)
- [ ] Add `CreatePullRequest` and `CreateBranch` to VCS Client interface
- [ ] Implement for GitHub
- [ ] Implement for GitLab
- [ ] Implement for Bitbucket Cloud
- [ ] Implement for Azure DevOps
- [ ] Add `/api/drift/remediate` endpoint
- [ ] Add comprehensive tests

### Stage 5: Additional Features (1-2 weeks)
- [ ] Add `/api/projects` endpoint
- [ ] Add `/api/runs` endpoint (if run history available)
- [ ] Add `DELETE /api/locks/{id}` endpoint
- [ ] Update API documentation
- [ ] Consider API versioning for future changes

---

## File Changes Summary

### Modified Files

| File | Phase | Changes |
|------|-------|---------|
| `server/events/command/project_result.go` | 1 | Add `MarshalJSON` method |
| `server/events/command/result.go` | 1 | Add `MarshalJSON` method |
| `server/events/command/result_test.go` | 1 | Add serialization tests |
| `server/events/command/project_context.go` | 2 | Add `API` field |
| `server/events/project_command_context_builder.go` | 2 | Propagate `API` flag |
| `server/events/command_requirement_handler.go` | 2 | Skip PR-specific requirements |
| `server/events/command_requirement_handler_test.go` | 2 | Add test cases |
| `server/events/vcs/client.go` | 4 | Add `CreatePullRequest`, `CreateBranch` |
| `server/events/vcs/github/client.go` | 4 | Implement PR creation |
| `server/events/vcs/gitlab/client.go` | 4 | Implement MR creation |
| `server/events/vcs/bitbucketcloud/client.go` | 4 | Implement PR creation |
| `server/events/vcs/azuredevops/client.go` | 4 | Implement PR creation |
| `server/controllers/api_controller.go` | 3,4,5 | Add new endpoints |
| `server/server.go` | 3,4,5 | Register new routes |

### New Files

| File | Phase | Purpose |
|------|-------|---------|
| `server/events/models/drift.go` | 3 | Drift models |
| `server/core/drift/parser.go` | 3 | Plan output parser |
| `server/core/drift/storage.go` | 3 | Drift status storage |
| `runatlantis.io/docs/api-endpoints.md` | 5 | Documentation updates |

---

## Consequences

### Benefits
1. **Error Visibility**: Proper error messages in API responses
2. **Drift Detection**: Proactive infrastructure monitoring capability
3. **Backwards Compatible**: No breaking changes to existing API
4. **Operational Flexibility**: Non-PR workflows for emergency situations
5. **Automation Ready**: PR creation API enables GitOps automation

### Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Field synchronization in MarshalJSON | Add maintainer warning comment, add linter check |
| VCS provider differences | Abstract through interface, comprehensive testing |
| Security (non-PR applies) | Keep other requirements enforced, audit logging |
| API stability | Consider versioning in future, maintain compatibility |

---

## References

- [PR #6093 - Error Response Fix](https://github.com/runatlantis/atlantis/pull/6093)
- [PR #6095 - API Non-PR Support](https://github.com/runatlantis/atlantis/pull/6095)
- [Current API Documentation](https://www.runatlantis.io/docs/api-endpoints.html)
- [ADR Process](http://thinkrelevance.com/blog/2011/11/15/documenting-architecture-decisions)
