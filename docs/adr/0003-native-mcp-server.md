# 3. Native MCP server for AI assistant integration

Date: 2026-07-20

> Re-validated against `main @ c5d4e192` (2026-08-01): updated for the `/readyz` web-auth exemption (#6669), the external plan store and `--enable-external-stores` (#6312), and the Go 1.26.5 toolchain.

## Status

Proposed

> **Numbering note:** ADR [0002](0002-api-enhancement-drift-detection.md) (on `main`) already covers API enhancement and drift detection (#6360). This MCP decision is therefore **0003**.

Related issues:

* [feat: optional MCP server for AI assistant integration #6530](https://github.com/runatlantis/atlantis/issues/6530)
* [Make API a first-class feature #6162](https://github.com/runatlantis/atlantis/issues/6162)
* [Break up events package #5950](https://github.com/runatlantis/atlantis/issues/5950)

Related prior art in-tree:

* HTTP API plan/apply endpoints (`server/controllers/api_controller.go`, PR [#997](https://github.com/runatlantis/atlantis/pull/997))
* `GET /api/locks` (PR [#5328](https://github.com/runatlantis/atlantis/pull/5328))
* Drift detection/remediation APIs, typed API responses, and `APIMiddleware.RequireAuth` (PR [#6360](https://github.com/runatlantis/atlantis/pull/6360); see also ADR 0002)
* External plan store (S3) and `--enable-external-stores` (PR [#6312](https://github.com/runatlantis/atlantis/pull/6312))
* Alpha API documentation (`runatlantis.io/docs/api-endpoints.md`)

## Context

### Problem

Platform engineers increasingly operate infrastructure with AI coding assistants (Claude Code, Cursor, GitHub Copilot, and similar). Those assistants can already open PRs and edit Terraform, but they cannot natively discover or invoke Atlantis operations. Operators today either:

1. Manually copy PR URLs, lock IDs, and job log snippets into chat prompts, or
2. Build ad-hoc wrappers around the alpha `/api/*` HTTP endpoints with bespoke auth and tool schemas.

The [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) is the emerging standard for exposing tools and resources to LLM clients. Atlantis already owns the primitives operators want (plan, apply, locks, job output, server status). Exposing them as MCP tools would let MCP-capable assistants interact with Atlantis without each organization inventing a private bridge.

Issue [#6530](https://github.com/runatlantis/atlantis/issues/6530) proposes shipping an **optional, in-process MCP server** inside the Atlantis binary. The issue author is willing to implement the work. This ADR records the architectural decisions needed before code lands.

### Lineage of this ADR

This ADR consolidates and supersedes:

1. The original feature request in [#6530](https://github.com/runatlantis/atlantis/issues/6530) (tool list, flags, second listener sketch), and
2. The refined design comment on that issue ([comment 4745075464](https://github.com/runatlantis/atlantis/issues/6530#issuecomment-4745075464): single port, read-only first, service layer prerequisite, reuse `--api-secret`).

Appendix A maps both the original issue and the deliberate deltas from that design comment (for example `/mcp/*` → `/api/mcp/*`).

### Current architecture (facts that constrain the design)

Atlantis is a single-process Go monolith:

```text
cmd/server.go
  → server.NewServer(userConfig, …)   // dependency injection
  → Server.Start()                    // one Mux, one http.Server, one shutdown path
```

Relevant HTTP surface today (`server/server.go` on current `main`, post-#6360):

| Route | Handler | Response | Auth today |
|-------|---------|----------|------------|
| `GET /status` | `StatusController.Get` | JSON | None |
| `POST /api/plan` | `APIController.Plan` | JSON | `X-Atlantis-Token` / `--api-secret` (constant-time) |
| `POST /api/apply` | `APIController.Apply` | JSON | Same |
| `GET /api/locks` | `APIController.ListLocks` | JSON (legacy `ListLocksResult`; `ListLocksResultAPI` preferred) | **None** (does not call `RequireAuth`) |
| `GET /api/drift/status` | `APIController.DriftStatus` | JSON | `APIMiddleware.RequireAuth` |
| `POST /api/drift/detect` | `APIController.DetectDrift` | JSON | `RequireAuth` |
| `POST /api/drift/remediate` | `APIController.Remediate` | JSON | `RequireAuth` |
| `GET /api/drift/remediate` | `APIController.ListRemediationResults` | JSON | `RequireAuth` |
| `GET /api/drift/remediate/{id}` | `APIController.GetRemediationResult` | JSON | `RequireAuth` |
| `DELETE /locks?id=` | `LocksController.DeleteLock` | text | None (optional web basic-auth only) |
| `POST /apply/lock` | `LocksController.LockApply` | text | None (optional web basic-auth only) |
| `DELETE /apply/unlock` | `LocksController.UnlockApply` | text | None (optional web basic-auth only) |
| `GET /jobs/{job-id}` | `JobsController.GetProjectJobs` | **HTML** | Optional web basic-auth |
| `GET /jobs/{job-id}/ws` | `JobsController.GetProjectJobsWS` | WebSocket | Optional web basic-auth |

The global apply-lock routes are registered unless `--disable-global-apply-lock` is set. Setting `--api-secret` does not protect these UI routes; without web basic-auth or an ingress restriction, callers with network access can create or remove the global apply lock.

Process and middleware facts:

* There is **one** `http.Server` listening on `--port` (default `4141`). There is no second listener anywhere in the codebase.
* Graceful shutdown is: signal → `Drainer.ShutdownBlocking()` → stats flush → DB close → `http.Server.Shutdown(5s)`.
* Negroni `RequestLogger` enforces web basic-auth when `--web-basic-auth` is enabled, **except** for `/events`, `/healthz`, `/readyz`, `/status`, and paths under `/api/` (`server/middleware.go`).
* Controllers are HTTP-bound (`func(http.ResponseWriter, *http.Request)`). There is no shared service package between controllers for plan/apply/locks (drift has begun typed request/response helpers under `server/controllers/api_*.go`).
* `APIController` private helpers are **not** uniformly “typed service results”:
  * `apiPlan` / `apiApply` return `(*command.Result, error)`.
  * `apiSetup` returns `error` only.
  * `apiParseAndValidate` returns `(*APIRequest, *command.Context, int /* HTTP status */, error)` — the status code is HTTP-layer coupling that **must be translated to domain errors** during service extraction (D3).
* PR [#6360](https://github.com/runatlantis/atlantis/pull/6360) added `APIMiddleware.RequireAuth` in `server/controllers/api_response.go`: empty `--api-secret` → “API is disabled”; otherwise `crypto/subtle.ConstantTimeCompare` on `X-Atlantis-Token`. Drift endpoints use it; plan/apply validation also uses constant-time compare. **MCP and locks hardening must reuse this middleware**, with the shared Bearer extension specified in D4, not reimplement secret checks.
* Typed lock response types already exist: `ListLocksResultAPI` / `LockDetailAPI` in `api_types.go`. `ListLocksResult` / `LockDetail` are **deprecated**. The live `ListLocks` handler still emits the legacy shape — MCP/service work should target the non-deprecated API types.
* `CommandRunner` (webhook/comment path) uses `Drainer.StartOp()` / `OpDone()`. **`APIController` does not.**
* `--allow-commands` is enforced by `CommentParser` for PR comments. **The HTTP API path does not consult it.**
* Global apply lock (`ApplyLocker.CheckApplyLock` / `IsLocked`) and `--disable-apply-all` are enforced on the **comment path** in `ApplyCommandRunner.Run`. **The HTTP `/api/apply` path does not.** There is **no** `--disable-apply` flag; the real knobs are `--disable-apply-all`, `--disable-global-apply-lock` (controls whether global apply lock UI/endpoints exist), and removing `apply` from `--allow-commands`.
* Job log buffers live on `AsyncProjectCommandOutputHandler.GetProjectOutputBuffer(jobID)`. That method is **not** on the `ProjectCommandOutputHandler` interface — only on the concrete type. Buffer contents are log lines plus `OperationComplete`; they do **not** store success/failure outcome.
* Job metadata: `pullToJobMapping` is a `sync.Map` of `PullInfo` → nested `*sync.Map` of `jobID` → `JobIDInfo`. `JobIDInfo` holds `Time` / step / description; `JobInfo` (on the send path) carries `PullInfo` + step but is not the durable lookup table. There is no `jobID →` reverse index today.
* `JobID` is generated on `command.ProjectContext` during command building, but plan/apply API responses are **synchronous** full `command.Result` payloads and do not return job IDs to the caller.
* The public API is documented as **alpha** and subject to change without deprecation (`runatlantis.io/docs/api-endpoints.md`). Alpha status allows breaking changes but does **not** excuse silent breaks without changelog/release notes.

### Why the original issue sketch is insufficient as-is

[#6530](https://github.com/runatlantis/atlantis/issues/6530) suggested mapping MCP tools directly onto existing controllers and a second listener (`--mcp-port=8081`). Review of the codebase shows several hard problems with that sketch:

1. **`JobsController.GetProjectJobs` cannot back `atlantis_get_job_status`.** It renders `ProjectJobsTemplate` (HTML log viewer), not structured job state.
2. **Controllers cannot be called without `httptest` wrappers.** That approach is brittle and fails the project's DRY / typed-API expectations (see PR [#997](https://github.com/runatlantis/atlantis/pull/997) review history).
3. **A second HTTP listener is novel.** It adds shutdown coordination, TLS duplication, Helm/Service/NetworkPolicy surface, and has no in-tree precedent.
4. **A separate `--mcp-token` duplicates `--api-secret`** and diverges from the auth model used by `/api/*` and the drift API (`APIMiddleware.RequireAuth` from #6360).
5. **Unlock is not “call `Locker.Unlock`”.** `LocksController.DeleteLock` also updates project status and comments on the PR after `DeleteLockCommand` deletes the lock and plan file — and, when `--enable-external-stores` is on, the external plan-store object as well (#6312).

### Alternatives considered

| Alternative | Summary | Why rejected (or deferred) |
|-------------|---------|----------------------------|
| **External MCP bridge** (standalone process calling `/api/*`) | Works without core changes | Every install re-implements auth, config, and deployment; still limited by alpha API + missing job JSON endpoint. Remains valid as a community option. |
| **OpenAPI-only for `/api/*`** | Helps LLMs generate HTTP calls | Not recognized by MCP-native clients; no tool/resource discovery model. |
| **Loopback HTTP from in-process MCP to `localhost:4141/api/*`** | Lowest code change | Couples MCP to unstable HTTP schemas; double serialization; poor errors; still needs job JSON endpoint. |
| **`httptest` adapters over controllers** | Avoids extraction | Brittle; fails on redirects/headers/status coupling; hard to test cleanly. |
| **Second listener on `--mcp-port`** | Network isolation by port | Novel dual-server lifecycle; ops complexity; isolation can be done at the ingress/NetworkPolicy layer if needed. |
| **Hand-rolled JSON-RPC without an SDK** | Zero dependency | Spec drift and transport churn (SSE → streamable HTTP) become permanent maintenance. |
| **Write tools in the first MCP PR** | Faster feature completeness | Highest security blast radius before transport/auth are proven. |
| **Per-tool unauthenticated MCP status** | Match public `/status` | Conflicts with a uniform token gate on `/api/mcp`; public `/status` already covers unauthenticated health. |

## Decision

We will add an **optional, experimental, in-process MCP server** to the Atlantis binary, subject to the decisions below. Implementation proceeds as an incremental PR stack. The first PR extracts a typed service layer and adds a JSON job-status API **without any MCP protocol code**, so the work stands alone even if MCP is later deferred.

### D1. Scope and maturity

1. MCP is **opt-in** and **off by default**.
2. MCP is labeled **experimental** in docs and changelog until maintainers promote it (same spirit as the alpha `/api/*` surface).
3. MVP exposes **read-only observability tools only**. Write tools (`plan`, `apply`, `unlock`) are deferred to a later PR that requires explicit maintainer ACK after the read-only surface lands.
4. MCP must not become a backdoor around existing safety mechanisms. When write tools ship, they share the same locking, allowlisting, and (where applicable) policy paths as the existing API — and may add stricter gates where AI-driven invocation warrants them (see D7).

### D2. Single HTTP port (no second listener)

MCP endpoints register on the **existing** Gorilla Mux router and `http.Server` used for webhooks, UI, and `/api/*`. We will **not** introduce `--mcp-port` or a second `ListenAndServe` in MVP.

Rationale:

* Retains the existing listener and server construction in `Server.Start()`.
* TLS, request logging, and shutdown coordination stay shared; MCP transport shutdown is explicitly integrated into that coordinator (D9).
* Job WebSockets provide a precedent for sharing the port, but not for automatic stream shutdown. Go excludes hijacked WebSockets from `http.Server.Shutdown` tracking; an active SSE response remains an HTTP request that must finish before shutdown can complete.
* Port isolation, if an operator wants it, is an **ingress / NetworkPolicy** concern, not an in-process dual-listener requirement.

**Route prefix:** Use paths under `/api/mcp/...` (not bare `/mcp/...`) so that:

* Existing Negroni web basic-auth exemptions for `/api/` continue to apply (`server/middleware.go`).
* MCP clients authenticate with the API token (D4), not HTTP Basic.
* Operators can reason about “API surface” as one prefix.

This is a deliberate delta from the design comment on #6530, which sketched `/mcp/sse` and `/mcp/message`. See Appendix A.

Exact path shapes (classic SSE pair vs streamable HTTP single endpoint) follow the pinned SDK’s recommended transport. The architectural constraint is **same port + `/api/` prefix**, not a specific path string frozen forever.

If maintainers later require network isolation by port, that can be revisited in a superseding ADR; it is not the default.

### D3. Service layer extraction is a prerequisite

MCP tools **must not** call HTTP controllers. Before (or as PR 1 of) MCP work, extract typed services that both HTTP controllers and MCP tools call.

Suggested package layout (names may adjust during implementation, intent must not):

```text
server/
  services/                 # NEW — transport-agnostic application services
    api_service.go          # Plan, Apply, ListLocks (+ validation helpers)
    lock_service.go         # Unlock with full side effects (status, PR comment, plan file + external plan store)
    job_service.go          # Job status / output buffer access
    status_service.go       # Server status (Drainer + version)
  controllers/              # thin HTTP adapters
  mcp/                      # MCP protocol adapters only (PR 2+)
```

Extraction rules:

1. Extract from **`APIController`**, not from `CommandRunner`. Comment/webhook command running and API plan/apply are different code paths today; conflating them in the first cut increases risk without buying correctness.
2. Controllers become thin adapters: parse HTTP → call service → write status/body. Prefer existing `APIResponder` / typed API helpers from #6360 where they fit.
3. Services return typed values and **domain errors**; they do not write to `http.ResponseWriter` and must not return raw HTTP status codes.
4. `APIController` private helpers (`apiPlan`, `apiApply`, `apiSetup`, `apiParseAndValidate`) are the natural extraction starting point, with explicit translation work:
   * Map request-validation and allowlist failures from `apiParseAndValidate`’s `(…, int status, error)` pattern to typed domain errors (e.g. invalid request, forbidden). Authentication and disabled-API responses move to the HTTP middleware (rule 6).
   * Keep `apiSetup`’s clone/lock steps as service methods that return `error` only.
5. Align with post-#6360 typed responses: list locks should use **`ListLocksResultAPI` / `LockDetailAPI`** (not the deprecated `ListLocksResult`). Drift/remediation can remain in their existing controller/service shapes for MVP; MCP need not expose drift tools initially, but extraction must not regress drift’s auth/response patterns.
6. Reuse **`APIMiddleware.RequireAuth`** at the HTTP edge for `/api/*` (and MCP under `/api/mcp`). As part of extraction, move plan/apply authentication out of `apiParseAndValidate` into this middleware and remove its independent `X-Atlantis-Token` check. Preserve disabled-API and invalid-token rejection; do not retain a second header-only check that would reject authenticated Bearer requests.

Example shapes (illustrative):

```go
// server/services/api_service.go (illustrative)
type APIService struct { /* existing deps moved from APIController */ }

func (s *APIService) Plan(ctx context.Context, req *APIRequest) (*command.Result, error)
func (s *APIService) Apply(ctx context.Context, req *APIRequest) (*command.Result, error)
func (s *APIService) ListLocks(ctx context.Context) (*ListLocksResultAPI, error)

// server/services/lock_service.go (illustrative)
// wraps DeleteLockCommand: lock + plan file + external plan-store cleanup (#6312), DB status, VCS comment
// source attributes the discard comment, e.g. "Atlantis UI", "API", "MCP"
func (s *LockService) Unlock(ctx context.Context, lockID string, source string) (*models.ProjectLock, error)

// server/services/job_service.go (illustrative)
func (s *JobService) GetStatus(ctx context.Context, jobID string) (*JobStatus, error)
```

### D4. Authentication reuses `--api-secret` and `APIMiddleware.RequireAuth`

1. **Do not add `--mcp-token`.**
2. MCP authenticates with the existing `--api-secret` value.
3. **Reuse `APIMiddleware.RequireAuth`** (`server/controllers/api_response.go`, added in #6360) for:
   * Plan/apply HTTP endpoints as part of D3 extraction, alongside the drift endpoints that already use it,
   * MCP HTTP endpoints under `/api/mcp`, and
   * Hardening `GET /api/locks` (D4.8).

   Do **not** reimplement secret comparison in MCP or list-locks handlers. Plan/apply and drift already use constant-time comparison; #6360 fixed the historical plain-string compare path. This ADR does **not** propose re-adding constant-time compare as new work — it is already on `main`.

4. Extend the shared middleware once to accept both header forms consistently on plan/apply, drift, hardened list-locks, the new jobs endpoint, and MCP:
   * `X-Atlantis-Token: <secret>` (existing API / `RequireAuth` convention), and
   * `Authorization: Bearer <secret>` (common for MCP / HTTP clients).

   Centralizing plan/apply authentication in PR 1a is a prerequisite; extending `RequireAuth` alone does not reach their current validation path. Bearer support lands in PR 2 for all of these routes together, with the existing header still supported.

5. **Bearer vs Basic:** `Authorization: Bearer` is only for the API/MCP token. It is not confused with `--web-basic-auth`, which uses the HTTP Basic scheme and is already **skipped** for all `/api/*` paths in `server/middleware.go`. MCP under `/api/mcp` therefore never requires web Basic credentials.
6. Startup validation:

   ```text
   if --mcp-enabled and --api-secret is empty → fail NewServer / startup with a clear error
   ```

7. **Uniform token gate for all MCP tools.** The MCP HTTP endpoint(s) under `/api/mcp` require a valid API secret on every request (via `RequireAuth` or equivalent). There is no per-tool auth bypass and no unauthenticated MCP tool.

   | Tool | Auth via MCP |
   |------|----------------|
   | `atlantis_get_status` | **Required** |
   | `atlantis_list_locks` | **Required** |
   | `atlantis_get_job_status` | **Required** |
   | Write tools (PR 3) | **Required** |

   Public `GET /status` (outside MCP) remains unauthenticated for load balancers and probes. Operators who want unauthenticated health checks continue to use `/status` or `/healthz`, not the MCP status tool.

8. **Hardening `GET /api/locks` is a breaking change** (alpha API, but still in use). Today it is unauthenticated and does **not** call `RequireAuth`. This ADR decides:

   * Call **`APIMiddleware.RequireAuth`** at the start of `ListLocks` (typically a one-line guard, same pattern as drift endpoints). When `--api-secret` is unset, respond with “API is disabled”; when set, require a valid token.
   * Prefer migrating the response body to **`ListLocksResultAPI`** in the same series so MCP and HTTP share the non-deprecated schema.
   * Label the auth change as **breaking** in the PR description, changelog/release notes, and `runatlantis.io/docs/api-endpoints.md`.
   * Prefer landing auth hardening in a **dedicated commit or PR** adjacent to service extraction (see D10), not as an unmarked side effect of a “refactor only” PR.
   * Include `DELETE /locks`, `POST /apply/lock`, and `DELETE /apply/unlock` in the security review. These UI routes currently rely on optional web basic-auth. Any hardening must specify how browser callers authenticate and document the migration; simply adding `RequireAuth` would break UI requests, which do not send an API token. Do not expose the API secret to browser JavaScript to make those requests work. UI-route hardening may land in the same series with an explicit compatible auth design, or in a separate security change; it is not an implicit consequence of securing `/api/locks`.

   Migration for existing scrapers:

   * **Deployments with `--api-secret` set:** send `X-Atlantis-Token` (or `Authorization: Bearer` once supported) with that secret on `GET /api/locks`.
   * **Deployments with no `--api-secret`:** today `GET /api/locks` works without any token (unlike plan/apply/drift). After 1b, lock listing is **disabled** the same way as other API endpoints (`API is disabled`). Operators who only used lock listing without enabling the write/drift API must set `--api-secret` (and send the token) to keep lock listing.

### D5. Configuration: two flags for MVP

| Flag | Type | Default | Purpose |
|------|------|---------|---------|
| `--mcp-enabled` | bool | `false` | Master switch |
| `--mcp-read-only` | bool | `true` | When true, write tools are **not registered** (undiscoverable) |

**Read-only behavior (chosen, not optional):**

* When `--mcp-read-only=true` (default), write tools are **not registered** with the MCP server. They do not appear in tool discovery. This is preferred over “register but always reject” because undiscoverable tools reduce prompt-injection / mistaken tool-selection surface.
* When `--mcp-read-only=false`, write tools are registered only if implemented (PR 3+).

**Credential scope:** `--mcp-read-only` controls the MCP tool registry, not the authority of `--api-secret`. The same secret already authorizes HTTP plan/apply even when MCP write tools are absent. Treat MCP clients as holding that authority unless network controls limit them to `/api/mcp` and deny access to other API/UI write routes and the direct backend. Reusing the secret does not provide a read-only credential; scoped credentials are outside this MVP.

**PR 2 interim behavior (write tools do not exist yet):**

* If an operator sets `--mcp-enabled=true` and `--mcp-read-only=false` before write tools ship, **startup must fail** (or, if maintainers prefer softer UX, log a loud warning and force read-only). Silent success is not allowed — `false` must not be a no-op that implies writes are enabled.
* Recommended default for PR 2: fail startup with a clear message, e.g. `--mcp-read-only=false is not supported until write tools are available`.

Wiring follows existing flag patterns in `cmd/server.go` and fields on `server/user_config.go` (`mapstructure` tags, flag parity tests in `cmd/server_test.go`).

**Explicitly omitted from MVP:**

| Omitted | Reason |
|---------|--------|
| `--mcp-port` | Single-port decision (D2) |
| `--mcp-token` | Reuse `--api-secret` (D4) |
| `--mcp-transport` | Server serves the SDK’s HTTP transport; stdio is a later product decision |

If MCP settings grow (allowlists, path prefix, resource exposure), prefer a server-side YAML `mcp:` block with a single enable flag — consistent with other features that moved detail out of CLI sprawl (precedent: the `external_stores` block gated by `--enable-external-stores`, #6312). That migration is out of scope for MVP unless maintainers request it up front.

MCP config must be stored on `Server` (or the MCP handler constructed in `NewServer`). `UserConfig` is consumed at construction time and is not retained as `s.UserConfig` today.

### D6. Tool registry

#### MVP (read-only) — PR 2

| MCP tool | Service method | Inputs | Notes |
|----------|----------------|--------|-------|
| `atlantis_get_status` | `StatusService.Get()` | none | Version, shutting down, in-progress ops (token required via MCP; public `/status` still available) |
| `atlantis_list_locks` | `APIService.ListLocks(filter?)` | optional filter (see below) | `ListLocksResultAPI` shape |
| `atlantis_get_job_status` | `JobService.GetStatus(jobID)` | `job_id` | Structured status + context + logs (see schema below) |

**`atlantis_list_locks` inputs / schema:** MVP may ship with no required inputs (`Locker.List()` returns all locks). Prefer an **optional `repository` (repo full name) filter** server-side so large multi-repo installs do not dump every lock into an LLM context. Filtering is a thin post-`List()` (or locker-supported) step — not a schema blocker, but recommended in PR 1/2 for payload size and injection surface (see D7 untrusted data). Response baseline is **`ListLocksResultAPI`** (`locks`, `total_count`, snake_case fields on `LockDetailAPI`) from #6360 — not the deprecated `ListLocksResult`.

**Job status backend (required new surface in PR 1):**

1. Extend `ProjectCommandOutputHandler` (or a narrow read interface) so buffer and job-metadata access is mockable and not cast-dependent.
2. **Add a reverse `jobID → metadata` index (new capability, not a thin wrapper).** Today the code only supports:
   * `jobID → OutputBuffer` (log lines + `OperationComplete`), and
   * `PullInfo → *sync.Map` of `jobID → JobIDInfo` (`pullToJobMapping`; nested `*sync.Map`, not a plain `map[string]JobIDInfo`).

   There is **no** `jobID →` reverse accessor. Repo, PR, path, and workspace live on the **`PullInfo` map key**; `JobIDInfo` holds job id, description, **timestamps** (`Time`), and step — not full pull context. Serving D6 context fields for a lone `job_id` therefore requires either ranging all pulls on every request (unacceptable) or a **new reverse index** (e.g. `jobID → {PullInfo, step, times, outcome}`). PR 1 must implement O(1) (or equivalent) reverse lookup and keep it updated on send/complete/cleanup. Do not assume “expose existing methods” is enough.
3. Add `GET /api/jobs/{job-id}` returning JSON for programmatic consumers (HTTP and MCP share this contract).
4. Keep existing `GET /jobs/{job-id}` HTML UI and WebSocket endpoints unchanged.
5. **Terminal outcome is not in `OutputBuffer` today.** `operation_complete: true` only means “no more log lines,” not success. PR 1 must **persist a terminal outcome** when a job completes (e.g. record status from the project command result when the job finishes), or document a temporary `unknown` outcome until that wiring lands in the same PR series. Do not invent success by scraping log text.
6. **Durability:** outcome, `completed_at`, context metadata, and log lines live **in-process alongside today’s buffers** (same lifetime as `projectOutputBuffers` / job maps). They are **lost on process restart** and cleaned up with existing pull/job cleanup — consistent with current log behavior. This ADR does **not** require durable job history in DB/Redis/the plan store (the #6312 plan store persists Terraform plan files, not job metadata).

**`status` values (normative boundaries):**

| Value | Meaning | Typical client action |
|-------|---------|------------------------|
| `running` | Job not finished | Poll again |
| `succeeded` | Completed without error or project failure | Done |
| `failed` | Completed with a nonempty `ProjectResult.Failure` and no `ProjectResult.Error`; includes rejection before execution, such as lock contention or unmet requirements | Inspect the failure reason before deciding whether to retry |
| `error` | Completed with a non-nil `ProjectResult.Error`; includes Terraform/configuration/provider errors as well as Atlantis/infrastructure errors | Inspect the error before deciding whether to retry |
| `unknown` | Completion and/or outcome not recorded yet | Treat as incomplete bookkeeping; do not assume success |

If both `Error` and `Failure` are present, `error` takes precedence. Neither status alone establishes retryability or whether Terraform started running. Finer categories require explicit error classification; do not infer them from the result field name or log text.

**Target JSON schema for `GET /api/jobs/{job-id}` / `atlantis_get_job_status`** (field names may use Go/JSON conventions finalized in PR 1; semantics are normative):

| Field | Required | Description |
|-------|----------|-------------|
| `job_id` | yes | Job identifier |
| `exists` | yes | Whether this job is known to the server |
| `status` | yes | See table above |
| `operation_complete` | yes | Whether log streaming finished |
| `repository` / `repo_full_name` | when known | From reverse index / `PullInfo` |
| `pull_num` | when known | PR number |
| `project_name` | when known | Project name if set |
| `project_path` / `path` | when known | Repo-relative project path |
| `workspace` | when known | Terraform workspace |
| `job_step` | when known | e.g. plan/apply step label if available |
| `started_at` | when known | First observation / mapping time |
| `completed_at` | when known | Time completion was recorded (in-memory; lost on restart) |
| `lines` | yes (may be empty) | Buffered log lines; **untrusted data** (D7); document truncation/max size |

Illustrative payload:

```json
{
  "job_id": "…",
  "exists": true,
  "status": "running",
  "operation_complete": false,
  "repo_full_name": "acme/infra",
  "pull_num": 42,
  "project_name": "",
  "path": "stacks/prod",
  "workspace": "default",
  "job_step": "plan",
  "started_at": "2026-07-16T12:00:00Z",
  "completed_at": null,
  "lines": ["…"]
}
```

Large buffers must define truncation or paging; document limits in the API docs. Clients should not assume full terraform output is always retained forever. Log lines are untrusted (D7); truncation also limits prompt-injection surface into LLM contexts.

#### Deferred (write) — PR 3, after read-only ACK

| MCP tool | Service method | Extra gates beyond token + `mcp-read-only=false` |
|----------|----------------|--------------------------------------------------|
| `atlantis_plan` | `APIService.Plan(req)` | Repo allowlist; working-dir lock; Drainer (D7) |
| `atlantis_apply` | `APIService.Apply(req)` | Same as plan; apply uses existing API semantics (plan then apply). **Also enforce global apply lock and `--disable-apply-all` in `APIService`** (see D7 — this is intentional **new** shared behavior, not today’s HTTP API parity). There is no `--disable-apply` flag. |
| `atlantis_unlock_project` | `LockService.Unlock(id, source)` | Full side effects: `DeleteLockCommand` (unlock + delete plan file + external plan-store object when `--enable-external-stores` is on, per #6312), DB project status update, VCS comment with accurate source attribution |

Plan/apply tool inputs mirror `APIRequest`:

| Field | Required | Description |
|-------|----------|-------------|
| `repository` | yes | Repository identity understood by VCS client |
| `ref` | yes | Git ref |
| `type` | yes | VCS host type (`Github`, `Gitlab`, …) |
| `pr` | no | Positive pull request number; omitted or non-positive values select non-PR API semantics and skip PR approval/mergeability requirements (D7) |
| `projects` | no* | Project names |
| `paths` | no* | `{directory, workspace}` pairs |

\* At least one of `projects` or `paths` is required (same rule as the HTTP API).

### D7. Safety, concurrency, and policy

#### Always (any MCP tool registration)

* Opt-in only; experimental labeling.
* Uniform token auth as in D4 (all tools).
* Repo allowlist checks for any tool that targets a repository (plan/apply path already does this in `apiParseAndValidate`).
* **Observability (PR 2 minimum, also useful on HTTP services):**
  * One structured log line per MCP tool invocation: tool name, success/error class, duration, and target identifiers when present (job id, lock id, repository) — clearly attributable as MCP-sourced (e.g. logger field `source=mcp`).
  * Metrics under a `mcp` (or equivalent) `tally` scope: invocation count, auth failures, latency by tool. Reuse existing stats flush on shutdown.
  * Do not log secrets or full plan bodies.
* **Untrusted tool results / prompt injection (LLM-specific):**
  * Job log lines returned by `atlantis_get_job_status` (and any future tool that surfaces plan diffs, terraform output, or hook output) are **untrusted, PR- and provider-influenced data**. They are not Atlantis instructions.
  * A malicious or compromised PR can emit log text that **impersonates instructions** to an assistant (classic indirect prompt injection), e.g. urging `atlantis_apply` or other write tools.
  * MCP tool descriptions and docs must state that log/`lines` content is **data to display or analyze**, not policy to obey.
  * Assistants and client integrations should treat these fields as **quoted untrusted content** (same discipline as browsing untrusted web pages).
  * Operators enabling MCP accept that **PR-controlled text enters the LLM context**. Mitigations are primarily client/operator discipline, truncation limits, optional repository filters on list tools, and keeping write tools unregistered unless explicitly approved — not a perfect server-side “LLM firewall” (out of scope for MVP).
  * This is independent of secret leakage in logs (also real; see Appendix B).

#### Drainer integration (deliberate HTTP API behavior change)

Today `APIController` plan/apply do **not** call `Drainer.StartOp()` / `OpDone()`. Folding Drainer into `APIService` means existing `POST /api/plan` and `POST /api/apply` clients will:

* Be refused (or fail fast) when Atlantis is shutting down,
* Increment `/status` `in_progress_operations` while running, and
* Hold shutdown open until admitted operations finish. `Drainer.ShutdownBlocking()` has no timeout; the existing five-second HTTP shutdown timeout begins only after draining and does not bound it.

This extends the existing comment-command drain policy to API operations. An operation whose remaining runtime exceeds the deployment's termination grace period can still be killed by the supervisor; a hung operation can block draining indefinitely. The current API's lack of Drainer tracking does not protect it from interruption: it can instead be cut off during HTTP shutdown.

That is a **behavior change for the current HTTP API**, not only an MCP concern. This ADR decides:

1. **Implement Drainer in `APIService` in PR 1** (service extraction), so HTTP and future MCP share correct shutdown semantics and accurate status counts.
2. Call it out in the PR description and **changelog/release notes**, including the unbounded drain wait and deployment termination-grace implications.
3. Refuse new write operations (HTTP and MCP) when `Drainer` reports shutting down.
4. Preserve the existing wait-for-completion policy for admitted operations. Document that operators must budget termination grace for expected plan/apply duration; it cannot guarantee completion of a hung operation. A bounded drain or forced Terraform cancellation requires a separate, explicit policy decision. MCP transport shutdown must not implicitly cancel admitted work (D9).

If maintainers insist PR 1 be behavior-preserving for HTTP, Drainer may move to PR 3 — but then PR 1 must **not** claim shutdown safety, and PR 3 must restate the HTTP impact. Default plan: **PR 1**.

#### When write tools ship

| Concern | Requirement |
|---------|-------------|
| Graceful shutdown | Use Drainer-wrapped service methods; preserve admitted work independently of MCP connection lifetime and quiesce handlers before closing shared dependencies (D9). |
| Git / workspace fencing | Use the same `WorkingDirLocker` and `Locker` paths as `APIController` today. |
| Read-only mode | `--mcp-read-only=true` (default) → write tools **not registered**. |
| Unlock completeness | Unlock must not call raw `Locker.Unlock` alone; use `DeleteLockCommand` plus controller side effects (status + PR comment). With `--enable-external-stores`, `DeleteLockCommand` also deletes the external plan-store object (#6312) — `LockService` must preserve that wiring. |
| Unlock comment attribution | Comment text must accurately attribute the discard source. `LockService` accepts a `source` (e.g. `Atlantis UI`, `API`, `MCP`) used by **all** callers of the service — UI, HTTP, and MCP — so wording stays consistent and tests/docs can update fixed English strings. No i18n framework depends on these strings today. |
| No policy backdoor | Do not skip apply requirements, policy checks, or VCS permission checks that the underlying service path already enforces. |

#### Policy matrix (explicit product decision for PR 3)

Today, **comment path** and **API path** enforce different policy stacks:

| Control | PR comments | HTTP `/api/plan\|apply` today |
|---------|-------------|-------------------------------|
| `--api-secret` | N/A | Yes (plan/apply; locks not yet) |
| Repo allowlist | Yes | Yes |
| `--allow-commands` (e.g. remove `apply`) | Yes (`CommentParser`) | **No** |
| Team allowlists | Yes (when configured) | **No** (not on API path) |
| Global apply lock (`ApplyLocker.CheckApplyLock` / `IsLocked` in `ApplyCommandRunner`) | Yes | **No** |
| `--disable-apply-all` | Yes | **No** |
| `--disable-global-apply-lock` | Affects whether global apply lock UI/endpoints are offered | N/A to API apply logic directly |
| `apply_requirements` / approval flow | Yes (command runners) | PR approval/mergeability requirements are checked for positive PR numbers; omitted/non-positive `pr` sets `SkipPRRequirements` and skips these checks |
| `Drainer` | Yes | **No** (becomes Yes after PR 1 per above) |

The optional `pr` input deliberately retains the API's non-PR behavior. MCP callers must not assume approval or mergeability gates apply when no positive PR number is supplied. Other policy checks remain in the shared path. Requiring a PR for every MCP write would be a separate policy change, not an existing API guarantee.

**There is no `--disable-apply` server flag.** Do not document or implement against that name.

For MCP write tools / `APIService`, this ADR decides:

1. **Baseline:** MCP uses the same `APIService` as HTTP (no separate backdoor path), plus Drainer and `--mcp-read-only`.
2. **Intentional apply-policy hardening (not “API parity as of today”):** `APIService.Apply` **must** check:
   * global apply lock (`ApplyLocker` / same `IsLocked` semantics as comment path), and
   * `--disable-apply-all` (reject non-targeted apply-all when configured),

   so HTTP `/api/apply` and MCP `atlantis_apply` gain these guards together. Call this out in changelog as a **behavior change for the HTTP API**, not silent MCP-only policy.
3. **Optional stricter gates** (`allow-commands`, team allowlists) may be added for MCP (and ideally HTTP) in PR 3 if maintainers want AI-driven invocation locked down further — document as intentional, not silent.
4. Document the full matrix in `runatlantis.io` when write tools ship so operators are not surprised.

#### Long-running plan/apply and job polling

HTTP API plan/apply are **synchronous**. First write-tool implementation may keep that behavior and document client timeouts.

**Do not block the first write PR on a full async redesign.** Optional follow-ups:

* Return job IDs (already allocated on `ProjectContext`) in API/MCP responses and poll `atlantis_get_job_status`.
* MCP progress notifications.
* Link to existing job log UI; human streaming remains on `/jobs/{id}/ws`.

**Polling guidance (document in API + MCP docs; PR 1/2):**

* Clients (including AI agents) should poll job status on the order of **1–5 seconds**, with backoff after repeated identical `running` responses, and **stop when `operation_complete` is true** (or `status` is terminal).
* Do not spin in a tight loop; job buffers are in-memory and high-frequency polling wastes CPU without improving freshness.
* Hard server-side rate limits are **not** required for MVP (no existing general rate-limit framework for this path). If abuse appears, follow up with soft limits and/or 429s. Emit metrics (D7) so operators can detect hot poll loops.
* MCP clients are not required to speak the existing WebSocket protocol; WebSocket remains the UI/streaming option outside MCP.

Any async redesign should improve HTTP API and MCP together via the service layer.

### D8. SDK and protocol

1. Use a maintained Go MCP SDK; pin an exact version in `go.mod`.
2. Candidate libraries (choice finalized in implementation PR):
   * [`mark3labs/mcp-go`](https://github.com/mark3labs/mcp-go) — widely adopted
   * [`modelcontextprotocol/go-sdk`](https://github.com/modelcontextprotocol/go-sdk) — official
3. **Decision criteria** for the pick (document the comparison in the PR):
   * Streamable HTTP and/or SSE **server** support matching current MCP remote transport guidance
   * **Builds and tests with Atlantis’s pinned Go toolchain**, using the repository's `go.mod` at implementation time as the source of truth. An SDK or transitive module that requires a newer Go blocks PR 2 until compatible versions are selected or Atlantis's toolchain is updated separately.
   * Active maintenance / release cadence
   * Stable Go module path (low rename risk)
   * Acceptable license and transitive dependency surface for Renovate/review
   * Security posture (recent CVEs, response practice)
   * How tightly the SDK couples Atlantis to one transport (prefer ability to keep `server/mcp` adapters thin so the SDK can be swapped later)
4. Prefer the SDK’s current recommended remote transport. Do not invent a private wire protocol. Do not freeze SSE-specific path names in this ADR.
5. **Stdio transport is out of MVP.** A future `atlantis mcp` subcommand or stdio mode needs a separate process-model decision (in-process stdin conflicts with a long-running server). External bridges remain valid for local-only setups.
6. When MCP is disabled, no MCP routes are registered and the SDK should not affect the default runtime path beyond module dependency cost.

### D9. Package and lifecycle integration

```text
MCP client
   │  HTTPS + token (all tools)
   ▼
:4141  Negroni + Mux
   │
   ├─ /api/plan|apply|locks|jobs/{id}  → controllers → services
   └─ /api/mcp/…                       → server/mcp  → services
                                              │
                                              ▼
                                    events / locking / jobs / VCS / Terraform
```

Lifecycle:

1. `NewServer` constructs services and, if enabled, an MCP handler with injected services + secrets + read-only flag.
2. `Start` registers MCP routes only when `--mcp-enabled`.
3. On signal, mark the server as draining and reject new service operations and MCP sessions/requests. This admission gate must also prevent new read requests from racing shared-dependency teardown.
4. Explicitly stop MCP transport streams/sessions through the SDK's shutdown mechanism or a handler-observed shutdown context. An open SSE response can otherwise exhaust the HTTP shutdown timeout; `http.Server.Shutdown` does not force-close an active handler when its deadline expires. Transport disconnection or session closure must not implicitly cancel an admitted plan/apply: its execution lifetime is owned by the service and tracked by the Drainer.
5. Await admitted operations under D7's drain policy and finish HTTP shutdown, including any remaining read handlers, **before** closing the database and flushing/closing stats. This changes today's database-before-HTTP ordering. If using `RegisterOnShutdown`, it must trigger transport termination early enough to let handlers finish; merely adding a callback to the current shutdown sequence does not fix the dependency ordering.

The implementation PR must document the chosen SDK shutdown hook and how handler completion is awaited. Single-port deployment retains one listener and one shutdown coordinator, but does not eliminate MCP lifecycle work. Test both an idle open stream and admitted work that outlives its client connection (D11).

### D10. Implementation PR stack

| PR | Scope | Contains MCP protocol? | Merge value without later PRs |
|----|-------|------------------------|-------------------------------|
| **1a** | Extract services; thin controllers; domain-error translation from HTTP status helpers; move plan/apply auth to `RequireAuth`; Drainer in `APIService` (changelog includes unbounded drain); admission and dependency-close ordering (D9); **reverse `jobID → metadata` index**; in-memory terminal outcome; `GET /api/jobs/{id}` JSON; optional locks filter; migrate list locks toward `ListLocksResultAPI`; tests | No | Yes — API quality (#6162) |
| **1b** (same PR series or immediate follow-up) | **Breaking:** `RequireAuth` on `GET /api/locks`; review project/global lock UI routes with an explicit browser-compatible hardening design (D4.8); changelog + docs for **token required** and **no-secret deployments lose free lock listing** | No | Yes — security consistency |
| **2** | `--mcp-enabled`, `--mcp-read-only`; shared `RequireAuth` with Bearer support for HTTP API and MCP; pinned SDK compatible with `go.mod` (D8); `/api/mcp` routes; explicit transport shutdown (D9); three read-only tools; audit log + metrics; untrusted-log and credential-scope docs; tests | Yes | Yes — useful observability |
| **3** | Write tools; **apply-policy hardening** (global apply lock + `--disable-apply-all` in `APIService` for HTTP+MCP); policy matrix docs; unlock `source` attribution; concurrency tests | Yes | Only after PR 2 ACK |
| **4** | Full docs: server config, security, client examples (Cursor / Claude Code), ops notes | No | Yes |

PR 1 commit messages and descriptions should lead with API quality benefits and cite #6162 / #5950; MCP enablement is a consumer of the layer, not the only justification. Mark **1b** and Drainer shutdown behavior explicitly as breaking/behavior changes even though the API is alpha.

### D11. Testing and docs expectations

Minimum bar:

* Unit tests for each service method and each MCP tool handler (auth success/failure, missing job ID, empty locks, job schema fields including outcome). Cover Terraform execution errors, pre-execution failures, and `Error` precedence when both result fields are populated.
* HTTP auth tests for both token headers across plan/apply, drift, hardened list-locks, jobs, and MCP; preserve missing-secret and invalid-token rejection. Any UI-route hardening must test existing browser workflows without exposing the API secret.
* Lifecycle tests: an idle open SSE stream terminates cleanly; new work is rejected during draining; admitted operations balance `StartOp`/`OpDone` on success/error paths and survive client disconnection; shared dependencies remain open until operations and read handlers finish. Verify the operation-drain wait is independent of the HTTP shutdown timeout.
* Write-tool tests distinguish PR and non-PR requests: approval/mergeability checks for a positive `pr`, and the documented `SkipPRRequirements` behavior otherwise, while preserving other policy checks.
* Flag registration parity tests (`cmd/server_test.go` patterns).
* Startup failure tests: `--mcp-enabled` without `--api-secret`; PR 2 `--mcp-read-only=false` when writes unavailable.
* Default CI and Docker paths keep MCP **disabled**.
* Docs: experimental banner; auth and shared-secret write authority even with read-only MCP; flags; tool table; security considerations (including untrusted job logs / prompt injection); job polling guidance; non-PR approval/mergeability semantics; locks auth migration (token + no-secret cases); no claim of stable MCP schema until status changes.
* Changelog/release notes for: locks auth break (including no-secret loss of free listing), Drainer-on-API's unbounded wait and deployment termination-grace implications, explicit MCP stream shutdown and dependency-close ordering, apply-policy hardening on `/api/apply` when PR 3 lands, new `/api/jobs/{id}`, experimental MCP.
* If AI assistance is used to author PRs, follow `AI_USAGE_POLICY.md` and project contribution norms.

### D12. Non-goals (this ADR)

* Replacing PR-comment-driven Atlantis workflows.
* Stabilizing the alpha HTTP API schema beyond what the service layer naturally requires.
* Multi-tenant MCP auth (per-user OAuth, per-team tokens) — future work.
* MCP resources/prompts beyond tools in MVP (may follow later).
* Bundling a full OpenAPI publication as a substitute for MCP.
* Enabling MCP in the default Helm/compose examples without explicit operator action.
* Hard server-side rate limiting infrastructure for job polling (MVP documents client backoff; metrics detect abuse).
* Per-tool unauthenticated access under `/api/mcp`.

## Consequences

### Positive

* AI assistants gain a standard, discoverable tool surface for Atlantis observability without each org shipping a private bridge.
* Service-layer extraction improves HTTP API testability and supports #6162 regardless of MCP adoption.
* JSON job status with outcome + context fills a real gap for all automation, not only MCP.
* Single-port design avoids novel process architecture and extra ops knobs.
* Reusing `--api-secret` with a uniform MCP token gate avoids credential sprawl and auth special cases.
* Read-only-first rollout limits operations exposed through MCP; its shared API credential still grants HTTP write authority (D5).
* Explicit breaking-change notes for locks auth and Drainer-on-API reduce surprise for existing automation.
* Experimental labeling sets correct stability expectations while the MCP ecosystem and Atlantis API both evolve.

### Negative / costs

* New Go module dependency (MCP SDK) and ongoing pin/upgrade burden (Renovate, security advisories).
* Additional attack surface when enabled: even read-only tools can leak lock and job log details; operators must treat `--api-secret` as sensitive and restrict ingress.
* MCP surfaces PR-controlled log text into LLM contexts (indirect prompt injection risk); cannot be fully eliminated server-side.
* Service extraction is non-trivial refactoring of `APIController` and related tests (~days of careful work).
* Reverse job index + terminal outcome bookkeeping is new in-memory state (same lifetime as today’s buffers; not durable across restart).
* Auth hardening on `GET /api/locks` breaks unauthenticated scrapers and no-secret lock listing (mitigated by alpha status + changelog); code change is small (`RequireAuth`) but behavior is not.
* Drainer extends the existing unbounded shutdown wait to HTTP API plan/apply; operations can outlast deployment termination grace and be killed externally. MCP streams need explicit termination, and HTTP handlers must finish before shared dependencies close (D7, D9).
* Enforcing global apply lock / `--disable-apply-all` on `/api/apply` (with MCP) is a deliberate policy upgrade beyond today’s API path.
* Auth and policy differences between comment path, HTTP API, and MCP must be documented carefully to avoid a false sense of identical security.
* Same-port deployment means MCP shares availability and TLS config with webhooks; operators who wanted a firewalled-only AI port must enforce that at the mesh/ingress layer.
* Experimental MCP schemas may change; early adopters may need client config updates.
* Review bandwidth: feature work competes with 1.0 and other roadmap items; the PR stack is designed so PR 1 remains valuable if MCP is paused.

### Risks and mitigations

| Risk | Mitigation |
|------|------------|
| AI agent triggers destructive apply | Write tools deferred; read-only MCP default; shared token still authorizes HTTP writes, so restrict MCP clients' route access where read-only access is intended (D5); PR 3 apply lock + `--disable-apply-all` in `APIService` |
| Indirect prompt injection via job logs | Document untrusted `lines`; tool descriptions; client discipline; truncation; optional lock filters; keep writes unregistered by default (D7) |
| Job tool backed by HTML handler | Explicit JSON API + reverse index + interface extension (D6) |
| Assistant cannot tell success vs mere completion | Persist `status` with error vs failed semantics; do not infer from logs alone (D6) |
| Context fields missing for job_id | New reverse `jobID → metadata` index in PR 1 (D6) |
| Schema drift between HTTP and MCP | Shared services; prefer `ListLocksResultAPI` and shared job types |
| Duplicated/weak API auth | Reuse `APIMiddleware.RequireAuth` from #6360 (D4) |
| Shutdown cuts in-flight API/MCP plans | Drainer in `APIService`; admitted work survives transport disconnects; document unbounded draining and external termination-grace limits (D7, D9) |
| Open MCP stream exhausts HTTP shutdown timeout | Explicit transport shutdown; await handlers before closing database/stats (D9) |
| HTTP API semantics change unnoticed | Changelog for Drainer, locks auth, apply-policy hardening (D4, D7, D10) |
| False “parity” with comment-path apply gates | Matrix documents gaps; PR 3 hardens apply explicitly (D7) |
| Web basic-auth breaks MCP clients | Serve under `/api/mcp` (D2) |
| Spec/transport churn in MCP | Pin SDK; experimental status; thin `server/mcp` adapter layer; SDK criteria (D8) |
| SDK requires newer Go than repo pin | D8 criterion: must build and test with the toolchain pinned in `go.mod` at implementation time |
| Unlock leaves plans behind | Always go through `DeleteLockCommand` + status/comment side effects (incl. external plan-store cleanup from #6312) |
| Unauthenticated lock listing on HTTP | `RequireAuth` + migration note (D4) |
| UI caller can remove the global apply lock without an API token | Document optional web basic-auth/ingress protection; review project/global lock UI auth with browser-compatible migration (D4.8) |
| Caller assumes every apply requires PR approval | Explicit non-PR `SkipPRRequirements` semantics in tool inputs and policy docs (D6, D7) |
| No-secret deploy loses free lock listing | Documented in D4 migration; set `--api-secret` to restore |
| Agent poll-storms job status | Documented backoff; metrics; hard rate limits only if needed later (D7) |
| Huge lock list into LLM context | Optional repository filter on list locks (D6) |
| `read-only=false` silent no-op before writes exist | Startup fail/warn in PR 2 (D5) |
| No audit trail for MCP | Structured log + metrics per invocation (D7) |
| Reviewer assumes durable job store | Outcome/logs in-memory only; lost on restart (D6) |

### Follow-up work (not decided here)

* Stdio / `atlantis mcp` subcommand process model.
* Async plan/apply with job ID responses for both HTTP and MCP.
* MCP resources (e.g. lock list as a resource) and progress notifications.
* Server-side YAML `mcp:` configuration block.
* Dual-listener mode if operators demand in-process port isolation.
* Hard rate limiting or MCP-native streaming if polling proves insufficient.
* Stronger server-side log redaction / injection mitigations beyond truncation and docs.
* Durable job history (DB-backed status) if operators need post-restart inspection.
* Promoting MCP (and/or HTTP API) out of experimental/alpha after production use.

## Appendix A — Mapping from issue #6530 and design comment to this ADR

**Lineage:** rows cover the original issue body. Rows marked *comment → ADR* document refinements from [comment 4745075464](https://github.com/runatlantis/atlantis/issues/6530#issuecomment-4745075464) that this ADR further hardens.

| Proposal source | Proposal | This ADR |
|-----------------|----------|----------|
| Issue | `--mcp-enabled` | Keep (D5) |
| Issue | `--mcp-port` default 8081 | **Reject** for MVP; single port (D2) |
| Issue | `--mcp-token` | **Reject**; reuse `--api-secret` (D4) |
| Issue | Tools wrap controllers directly | **Reject**; service layer first (D3) |
| Issue | `atlantis_get_job_status` → `JobsController.GetProjectJobs` | **Reject**; JSON via `JobService` + interface extension (D6) |
| Issue | All six tools in first delivery | **Split**: read-only MVP, writes later (D1, D6) |
| Issue | `mark3labs/mcp-go` | Candidate; finalize with criteria in D8 |
| Issue | Incremental PRs | Adopted and expanded (D10) |
| *Comment → ADR* | Single port; sketch `/mcp/sse` + `/mcp/message` | Single port **and** prefix **`/api/mcp/...`** so web basic-auth exemptions apply and the surface reads as API; exact path shapes follow the SDK (not frozen SSE names) (D2) |
| *Comment → ADR* | Read-only first; reuse `api-secret` | Adopted; **all** MCP tools token-required (no optional status) (D4) |
| *Comment → ADR* | Service layer + `GET /api/jobs/{id}` | Adopted; schema expanded with outcome + context (D6) |
| *Comment → ADR* | Two flags | Adopted; clarify not-registered + PR 2 `read-only=false` (D5) |

## Appendix B — Operator security checklist (when enabling MCP)

1. Set a strong `--api-secret`; rotate if previously exposed.
2. Leave `--mcp-read-only=true` unless write tools are implemented **and** organizationally approved.
3. Restrict network access to trusted clients (VPN, mesh, IP allowlist, ingress auth). For clients intended to have read-only MCP access, allow only `/api/mcp` routes and block other API/UI write routes and direct backend access. Allowing all of `/api/` exposes HTTP plan/apply to the same secret, regardless of `--mcp-read-only`.
4. Do not expose Atlantis to the public internet with only a static API secret.
5. Monitor MCP structured logs and `mcp` metrics (and `/status`) for unexpected plan/apply when write tools exist.
6. Treat job log output as potentially **sensitive** (secrets in Terraform output, plan diffs).
7. Treat job log output (and plan diffs returned via tools) as **untrusted, PR-controlled data** relative to any LLM or assistant: they can contain prompt-injection text that tries to trigger write tools or ignore policy. Configure assistants to treat tool `lines` as data, not instructions; prefer human confirmation before any write tool even if later enabled.
8. After upgrading through locks-auth hardening:
   * Update scrapers that called `GET /api/locks` without a token to send the API secret.
   * If you never set `--api-secret` and relied on free lock listing, set a secret (and send it) — listing is no longer available with API disabled.
9. Protect the project/global lock UI routes with web basic-auth or ingress controls; `--api-secret` alone does not protect them. Follow the documented browser-auth migration if these routes are hardened.
10. When write tools ship, supply a positive `pr` to use PR approval/mergeability requirements. Non-PR requests skip those requirements under the existing API semantics.

## Appendix C — Open questions for maintainers

These do not block accepting the bulk of this ADR; answers refine PR sequencing:

1. Confirm single-port `/api/mcp` vs. a desire for a second listener for network isolation.
2. Prefer `mark3labs/mcp-go` or `modelcontextprotocol/go-sdk` after applying the D8 criteria.
3. Confirm **1a + 1b** sequencing: ship locks auth hardening in the same PR series as service extraction vs. a separate security PR (breaking change is intentional either way).
4. Confirm Drainer-on-`APIService` in **PR 1** (recommended) vs. delay to PR 3 for a behavior-preserving extraction.
5. Confirm apply-policy hardening in PR 3: global apply lock + `--disable-apply-all` on `APIService` (recommended) vs leave HTTP API without those gates.
6. Additionally enforce `--allow-commands` / team allowlists on MCP (and possibly HTTP API), or leave that for a later hardening pass.
7. Whether ADR acceptance is required before PR 1 (service layer only), or only before PR 2 (MCP protocol).
8. Timing relative to the 1.0 focus: ship PR 1 now as API work; hold PR 2+ for post-1.0 if needed.
9. PR 2 interim: **fail startup** on `--mcp-read-only=false` (recommended) vs. warn-and-force-read-only.

## References

* Michael Nygard, [Documenting Architecture Decisions](http://thinkrelevance.com/blog/2011/11/15/documenting-architecture-decisions)
* [Model Context Protocol specification](https://modelcontextprotocol.io/)
* Atlantis ADR process: [docs/adr/0001-record-architecture-decisions.md](0001-record-architecture-decisions.md)
* API enhancement / drift ADR: [docs/adr/0002-api-enhancement-drift-detection.md](0002-api-enhancement-drift-detection.md) (on `main`)
* Design discussion on #6530: [comment 4745075464](https://github.com/runatlantis/atlantis/issues/6530#issuecomment-4745075464)
* API docs: [runatlantis.io/docs/api-endpoints.md](../../runatlantis.io/docs/api-endpoints.md)
* PRs: [#997](https://github.com/runatlantis/atlantis/pull/997) (plan/apply), [#5328](https://github.com/runatlantis/atlantis/pull/5328) (list locks), [#6360](https://github.com/runatlantis/atlantis/pull/6360) (drift + `APIMiddleware` + typed API responses)
