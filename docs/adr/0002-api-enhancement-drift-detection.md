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
| ---------- | -------- | --------------- | ------------- |
| `/api/plan` | POST | Yes | Trigger terraform plan |
| `/api/apply` | POST | Yes | Trigger terraform apply |
| `/api/locks` | GET | No | List active locks |
| `/status` | GET | No | Server status |
| `/readyz` | GET | No | Readiness check |
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

1. **Propagate API flag in command builder**:

```go
// server/events/project_command_context_builder.go
func newProjectCommandContext(...) command.ProjectContext {
    return command.ProjectContext{
        // ... existing fields ...
        API: ctx.API,
    }
}
```

1. **Skip PR-specific requirements for non-PR API calls**:

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

**Dependency Checks for Non-PR API Applies**:

Non-PR API applies must also handle `depends_on` without relying on pull
request status from a VCS event. Before building apply project contexts for
`ctx.API && ctx.Pull.Num == 0`, seed `ctx.PullStatus` with a synthetic
`models.PullStatus` containing every selected project and its dependencies.
Each entry should use the project config's `ProjectName`, `RepoRelDir`, and
`Workspace`, and start as `models.PlannedPlanStatus` unless the current
operation already knows it has no changes.

`APIController.apiApply` currently runs projects in a simple loop instead of
using `project_command_pool_executor`, so the API apply path must update
`ctx.PullStatus.Projects` immediately after each project apply result. The
update should use the same `Workspace`, `RepoRelDir`, and `ProjectName` match
as the pool executor and assign `result.PlanStatus()` so later project contexts
see dependencies as `AppliedPlanStatus` or `PlannedNoChangesPlanStatus` after
earlier applies succeed. Alternatively, route non-PR API applies through the
pool executor so it owns the same status update behavior.

`ValidateProjectDependencies` must also return a clear failure when
`ctx.PullStatus` is nil for a project with dependencies, or when a declared
dependency is absent from the synthetic status. Non-PR API apply requests should
therefore include dependent projects in the same request or fail before apply
instead of panicking or treating the dependency as satisfied.

**Files to Modify**:

- `server/events/command/project_context.go` - Add `API` field
- `server/events/project_command_context_builder.go` - Propagate `API` flag
- `server/events/project_command_builder.go` - Seed synthetic `PullStatus` for non-PR API applies
- `server/controllers/api_controller.go` - Update synthetic `PullStatus` after each API apply result or route applies through the pool executor
- `server/events/command_requirement_handler.go` - Skip PR-specific requirements
- `server/events/command_requirement_handler_test.go` - Add test cases for PR-specific requirements and dependency failures
- `server/events/project_command_builder_test.go` - Add non-PR API apply dependency status coverage

---

### Phase 3: Drift Detection Endpoints

**Goal**: Enable infrastructure drift detection outside of PR workflows.

**New Endpoints**:

| Endpoint                | Method   | Description                                      |
| ----------------------- | -------- | ------------------------------------------------ |
| `POST /api/plan`        | Enhanced | Support `PR: 0` or omitted for drift detection   |
| `GET /api/drift/status` | New      | Get drift status for repository/projects         |

**API Authentication**:

All new `/api/drift/*` endpoints must enforce the same API authentication as
`/api/plan` and `/api/apply` before parsing request input or returning drift
data. Implementations should reuse the `X-Atlantis-Token` validation currently
performed by `APIController.apiParseAndValidate`, or extract that check into a
shared API-auth helper for handlers that do not accept the same request body.
The handler must reject requests when the API secret is unset or the
`X-Atlantis-Token` header does not match the configured secret. Do not rely on
web basic auth for these endpoints, because `/api/*` routes are API routes and
their handlers are responsible for API-token enforcement.

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

Add drift detection helpers to expose plan output as structured drift status.
These helpers must reuse `models.NewPlanSuccessStats` (or extend it if
Terraform/OpenTofu adds new counters) instead of adding a separate regex parser,
so drift API responses stay consistent with PR plan comments and existing
summary rendering.

```go
// server/events/models/drift.go

// DriftSummary represents detected infrastructure drift
type DriftSummary struct {
    HasDrift       bool   `json:"has_drift"`
    ToImport       int    `json:"to_import"`
    ToAdd          int    `json:"to_add"`
    ToChange       int    `json:"to_change"`
    ToDestroy      int    `json:"to_destroy"`
    ToForget       int    `json:"to_forget"`
    ChangesOutside bool   `json:"changes_outside"`
    Summary        string `json:"summary"` // e.g., "2 to import, 2 to add, 1 to change, 0 to destroy, 1 to forget"
}

// DriftRepositoryKey identifies the VCS repository that owns drift status.
// VCSHostType should be Repo.VCSHost.Type.String().
type DriftRepositoryKey struct {
    VCSHostType string `json:"type"`
    Hostname    string `json:"hostname"`
    Repository  string `json:"repository"`
}

// DriftProjectKey identifies a drift status record for an Atlantis project on
// a specific ref. ProjectName is optional in atlantis.yaml, so Path and
// Workspace must be part of the storage key to avoid collisions between
// unnamed projects in the same repository. Ref must also be part of the key so
// drift checks against different branches do not overwrite each other.
type DriftProjectKey struct {
    Ref         string `json:"ref"`
    ProjectName string `json:"project_name,omitempty"`
    Path        string `json:"path"`
    Workspace   string `json:"workspace"`
}

// ParseDriftFromPlan extracts drift information from terraform plan output.
func ParseDriftFromPlan(planOutput string) DriftSummary {
    stats := NewPlanSuccessStats(planOutput)

    return DriftSummary{
        HasDrift:       stats.Changes || stats.ChangesOutside,
        ToImport:       stats.Import,
        ToAdd:          stats.Add,
        ToChange:       stats.Change,
        ToDestroy:      stats.Destroy,
        ToForget:       stats.Forget,
        ChangesOutside: stats.ChangesOutside,
        Summary:        (&PlanSuccess{TerraformOutput: planOutput}).Summary(),
    }
}
```

**New Drift Status Endpoint**:

```go
// GET /api/drift/status?type=Github&hostname=github.com&repository=org/repo&ref=main
//   &project=myproject&path=.&workspace=default
// type must use the canonical models.VCSHostType.String() value. Type,
// hostname, repository, and ref identify the VCS source. Project, path, and
// workspace are optional filters; path and workspace identify unnamed Atlantis
// projects. Ref is required so callers can distinguish drift runs for the same
// project on different branches.
type DriftStatusResponse struct {
    Repository DriftRepositoryKey `json:"repository"`
    Projects   []ProjectDrift    `json:"projects"`
    CheckedAt  time.Time         `json:"checked_at"`
}

type ProjectDrift struct {
    Repository  DriftRepositoryKey `json:"repository"`
    ProjectName string             `json:"project_name,omitempty"`
    Path        string             `json:"path"`
    Workspace   string             `json:"workspace"`
    Ref         string             `json:"ref"`
    Drift       DriftSummary       `json:"drift"`
    LastChecked time.Time          `json:"last_checked"`
}
```

**Drift Storage**:

The status endpoint is a read endpoint, not a recomputation endpoint. Each
successful drift-detection `/api/plan` run must parse every project plan result
with `ParseDriftFromPlan` and persist one `ProjectDrift` record per project.
That record must include the VCS repository identity (VCS type, hostname, and
repository full name), the request source metadata (`Ref`), the project
identity (`ProjectName`, `Path`, `Workspace`), the parsed drift summary, and
`LastChecked`. `GET /api/drift/status` then filters and returns the persisted
records for the requested VCS repository/ref/project key.
The repository key must not include a separate derived `RepoID`; derive
`models.Repo.ID()` from the same hostname and repository fields only when a log
message or downstream API call needs that string.

If the implementation instead chooses to compute drift on demand, the status
endpoint request shape must change to include the required `Ref`, VCS type,
hostname, repository, and `Paths` inputs and must run a plan synchronously.
With the endpoint shape above, storage is required.

```go
// server/core/drift/storage.go

type DriftStorage interface {
    StoreDriftStatus(repo models.DriftRepositoryKey, status models.ProjectDrift) error
    GetDriftStatus(repo models.DriftRepositoryKey, key models.DriftProjectKey) (*models.ProjectDrift, error)
    ListDriftStatus(repo models.DriftRepositoryKey, ref string) ([]models.ProjectDrift, error)
}

// In-memory implementation for initial version
type InMemoryDriftStorage struct {
    mu     sync.RWMutex
    status map[models.DriftRepositoryKey]map[models.DriftProjectKey]models.ProjectDrift
}
```

**Files to Create/Modify**:

- `server/events/models/drift.go` - New drift response, parser, and key models
- `server/events/models/models.go` - Reuse or extend `NewPlanSuccessStats` for any new drift counters
- `server/core/drift/storage.go` - Required drift status storage for `/api/drift/status`
- `server/controllers/api_controller.go` - Add drift status endpoint
- `server/controllers/api_controller_test.go` - Add status endpoint API-token coverage

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

func (g *Client) CreatePullRequest(
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

func (g *Client) CreateBranch(
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
    _, _, err = g.client.Git.CreateRef(g.ctx, owner, repoName, github.CreateRef{
        Ref: "refs/heads/" + branchName,
        SHA: ref.Object.GetSHA(),
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

func (g *Client) CreatePullRequest(
    logger logging.SimpleLogging,
    repo models.Repo,
    opts vcs.CreatePullRequestOptions,
) (models.PullRequest, error) {
    labels := gitlab.LabelOptions(opts.Labels)
    title := opts.Title
    if opts.Draft {
        // GitLab's create-MR API uses a title prefix to create draft MRs.
        title = "Draft: " + title
    }

    mr, _, err := g.Client.MergeRequests.CreateMergeRequest(repo.FullName, &gitlab.CreateMergeRequestOptions{
        Title:        gitlab.String(title),
        Description:  gitlab.String(opts.Body),
        SourceBranch: gitlab.String(opts.HeadBranch),
        TargetBranch: gitlab.String(opts.BaseBranch),
        Labels:       &labels,
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
type DriftRemediatePath struct {
    Directory string `json:"directory" validate:"required"`
    Workspace string `json:"workspace,omitempty"`
}

type DriftRemediateRequest struct {
    Repository  string                `json:"repository" validate:"required"`
    Ref         string                `json:"ref" validate:"required"`  // Base branch
    Type        string                `json:"type" validate:"required"` // VCS type
    Projects    []string              `json:"projects,omitempty"`       // Projects to remediate
    Paths       []DriftRemediatePath  `json:"paths,omitempty"`
    Title       string                `json:"title,omitempty"`       // PR title
    Description string                `json:"description,omitempty"` // PR body
    Draft       bool                  `json:"draft,omitempty"`
    AutoMerge   bool                  `json:"auto_merge,omitempty"` // Enable auto-merge after approval
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

The remediation handler must enforce the shared API-token authentication before
creating branches, commits, or PR/MRs. It must reject unauthenticated requests
before running a plan or exposing drift data.

**Remediation Workflow**:

1. Run plan for each specified project (detect what needs to be applied)
2. Create a new branch (e.g., `atlantis-drift-remediation-{timestamp}`)
3. Materialize a non-empty branch change before opening a PR/MR. This can be a
   generated drift-remediation manifest/report, proposed configuration edits
   from a remediation engine, or another reviewable artifact committed to the
   branch. If Atlantis cannot create a commit that makes the head differ from
   the base, return an error with the drift summary instead of calling
   `CreatePullRequest`.
4. Create PR/MR with drift summary in description
5. Return PR details for user to review and merge

**Files to Create/Modify**:

- `server/events/vcs/client.go` - Add `CreatePullRequest`, `CreateBranch` interface methods
- `server/events/vcs/proxy.go` - Proxy new methods to the concrete VCS client
- `server/events/vcs/not_configured_vcs_client.go` - Return explicit unsupported errors for unconfigured providers
- `server/events/vcs/github/client.go` - Implement for GitHub
- `server/events/vcs/gitlab/client.go` - Implement for GitLab
- `server/events/vcs/bitbucketcloud/client.go` - Implement for Bitbucket Cloud
- `server/events/vcs/bitbucketserver/client.go` - Implement for Bitbucket Server or return explicit unsupported errors
- `server/events/vcs/azuredevops/client.go` - Implement for Azure DevOps
- `server/events/vcs/gitea/client.go` - Implement for Gitea or return explicit unsupported errors
- `server/events/vcs/mocks/mock_client.go` - Regenerate after changing `vcs.Client`
- `server/controllers/api_controller.go` - Add remediation endpoint
- `server/controllers/api_controller_test.go` - Add remediation endpoint API-token coverage

---

### Phase 5: Additional API Enhancements

**Goal**: Add useful operational endpoints and improve API stability.

**New Endpoints**:

| Endpoint | Method | Auth | Description |
| ---------- | -------- | ------ | ------------- |
| `GET /api/projects` | GET | Yes | List configured projects |
| `GET /api/runs` | GET | Yes | List recent plan/apply runs |
| `DELETE /api/locks?id={lockID}` | DELETE | Yes | Force unlock a specific lock |

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
// DELETE /api/locks?id=<url-escaped lock key from GET /api/locks>
// Requires X-Atlantis-Token authentication
func (a *APIController) ForceUnlock(w http.ResponseWriter, r *http.Request) {
    // Validate authentication
    // Extract lock ID from the id query parameter. Atlantis lock keys contain
    // slash separators, so they must not be modeled as a single path segment.
    // Call a.Locker.Unlock(lockID)
    // Return success/failure
}
```

The force-unlock endpoint must accept the exact lock key returned by
`GET /api/locks` as `LockDetail.Name`. Those keys are generated as
`<repo>/<path>/<workspace>/<project>` and contain `/`, so implementations must
use a query parameter, request body field, or explicit catch-all route that
preserves the full URL-escaped value before calling `Locker.Unlock(lockID)`.

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

- [ ] Create drift response models that reuse or extend `models.NewPlanSuccessStats`
- [ ] Add `DriftSummary` to plan response (optional enhancement)
- [ ] Implement required drift status storage (in-memory)
- [ ] Persist parsed drift status from each successful drift-detection `/api/plan` project result
- [ ] Add `/api/drift/status` endpoint
- [ ] Add documentation for drift detection usage

### Stage 4: Drift Remediation (3-4 weeks)

- [ ] Add `CreatePullRequest` and `CreateBranch` to VCS Client interface
- [ ] Update `ClientProxy`, `NotConfiguredVCSClient`, and generated VCS mocks
- [ ] Implement for GitHub
- [ ] Implement for GitLab
- [ ] Implement for Bitbucket Cloud
- [ ] Implement or explicitly return unsupported errors for Bitbucket Server
- [ ] Implement for Azure DevOps
- [ ] Implement or explicitly return unsupported errors for Gitea
- [ ] Define and test the branch materialization step that creates at least one commit before PR/MR creation
- [ ] Add `/api/drift/remediate` endpoint
- [ ] Add comprehensive tests

### Stage 5: Additional Features (1-2 weeks)

- [ ] Add `/api/projects` endpoint
- [ ] Add `/api/runs` endpoint (if run history available)
- [ ] Add `DELETE /api/locks?id={lockID}` endpoint
- [ ] Update API documentation
- [ ] Consider API versioning for future changes

---

## File Changes Summary

### Modified Files

| File | Phase | Changes |
| ------ | ------- | --------- |
| `server/events/command/project_result.go` | 1 | Add `MarshalJSON` method |
| `server/events/command/result.go` | 1 | Add `MarshalJSON` method |
| `server/events/command/result_test.go` | 1 | Add serialization tests |
| `server/events/command/project_context.go` | 2 | Add `API` field |
| `server/events/project_command_context_builder.go` | 2 | Propagate `API` flag |
| `server/events/command_requirement_handler.go` | 2 | Skip PR-specific requirements |
| `server/events/command_requirement_handler_test.go` | 2 | Add test cases |
| `server/events/models/models.go` | 3 | Reuse or extend `NewPlanSuccessStats` for drift counters |
| `server/events/vcs/client.go` | 4 | Add `CreatePullRequest`, `CreateBranch` and shared option types |
| `server/events/vcs/proxy.go` | 4 | Proxy branch and PR creation calls |
| `server/events/vcs/not_configured_vcs_client.go` | 4 | Return unsupported errors for unconfigured clients |
| `server/events/vcs/github/client.go` | 4 | Implement PR creation |
| `server/events/vcs/gitlab/client.go` | 4 | Implement MR creation |
| `server/events/vcs/bitbucketcloud/client.go` | 4 | Implement PR creation |
| `server/events/vcs/bitbucketserver/client.go` | 4 | Implement PR creation or explicit unsupported errors |
| `server/events/vcs/azuredevops/client.go` | 4 | Implement PR creation |
| `server/events/vcs/gitea/client.go` | 4 | Implement PR creation or explicit unsupported errors |
| `server/events/vcs/mocks/mock_client.go` | 4 | Regenerate after `vcs.Client` changes |
| `server/controllers/api_controller.go` | 3,4,5 | Add new endpoints |
| `server/server.go` | 3,4,5 | Register new routes |

### New Files

| File | Phase | Purpose |
| ------ | ------- | --------- |
| `server/events/models/drift.go` | 3 | Drift models |
| `server/core/drift/storage.go` | 3 | Required drift status storage |
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
| ------ | ------------ |
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
