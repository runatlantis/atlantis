# 3. etcd High-Availability Locking and Coordination Backend

Date: 2026-09-09

## Status

Proposed

**Delivered in phases.** This decision is accepted as the target architecture, but
ships incrementally to reduce risk:

- **Phase 1 — external etcd only (current).** Atlantis connects to an
  operator-provided external etcd cluster (`--etcd-mode=external`). The full
  active-active coordination stack — leased PR ownership, owner-routing, execution
  fencing, migration, and recovery — is delivered against that external cluster.
- **Phase 2 — embedded etcd (future).** The embedded in-process etcd voter, the
  member StatefulSet topology, and the embedded lifecycle/membership/restore
  machinery (decisions 7 and 9 below) are deferred. Until phase 2 lands,
  `--etcd-mode=embedded` is rejected at configuration validation.

The remaining decisions below hold for both phases; the embedded-specific parts
apply when phase 2 is delivered.

Supersedes nothing. Relates to
[ADR 0002 (API Enhancement and Drift Detection)](0002-api-enhancement-drift-detection.md).

Companion documents:

- Proposal / RFC: [`../proposals/2026-09-etcd-ha-locking.md`](../proposals/2026-09-etcd-ha-locking.md)
- The detailed design, implementation plan, and reference deployment manifests
  are provided in the implementation pull request (under `docs/superpowers/`),
  per the ADR process: this record captures the decision; implementation detail
  lives with the code.

## Context

Atlantis persists its locks, pull/project statuses, and the global apply lock in a
single-writer store — BoltDB by default (an on-disk file), or Redis. Both assume
exactly one Atlantis process is authoritative:

- **BoltDB** is a local file; it cannot be shared, so a BoltDB Atlantis is a
  single point of failure. A crash or node loss stalls all Terraform automation
  until the process is rescheduled and its volume reattached.
- **Redis** allows a shared store, but Atlantis has no coordination layer above
  it: nothing prevents two Atlantis replicas from planning/applying the *same*
  pull request concurrently. Redis alone does not make Atlantis safely
  active-active.

Teams that run Terraform automation as critical infrastructure need Atlantis to
survive a node failure without a manual failover and without risking two replicas
acting on the same pull request. They also want to scale ingress and Terraform
execution horizontally.

etcd provides a linearizable, quorum-replicated key/value store with leases,
watches, and multi-key transactions — the primitives needed for distributed
locking, leader-free ownership, and fencing. It can be embedded in the Atlantis
binary or run as an external cluster.

The core tension is safety versus availability: making Atlantis active-active must
not weaken the single-writer guarantee that protects Terraform state. Terraform
apply is destructive and not idempotent, so the design must **fail closed** and
must not promise exactly-once execution it cannot deliver.

## Decision

Add an **opt-in etcd locking and coordination backend**, selected with
`--locking-db-type=etcd`, in a new `server/core/etcd` package. The following
decisions are locked:

1. **BoltDB stays the default; there is no fallback.** BoltDB and Redis keys and
   behavior are unchanged. When etcd is selected and unavailable, Atlantis fails
   closed rather than falling back to another backend.

2. **Two runtime modes, one client path — delivered in phases.**
   `--etcd-mode=external` connects to an existing etcd cluster; `--etcd-mode=embedded`
   runs one embedded etcd voter inside the Atlantis process. Both construct the same
   long-lived `clientv3` client and use the same database, ownership, routing, and
   fencing code. A `Backend` owns the client (and, in embedded mode, the server); the
   database is its designated idempotent close delegate. **Phase 1 ships external mode
   only; embedded mode (the in-process server path) is deferred to phase 2 and rejected
   at validation until then.** The single-client-path design keeps embedded an additive
   change: no ownership, routing, or fencing code differs between the two modes.

3. **Active-active with fenced PR ownership — no leader role.** Every replica
   accepts ingress and can execute Terraform. For each pull request, exactly one
   process wins a **leased ownership claim**; commands are resolved to the owner
   and either executed locally or forwarded over a secure internal transport
   (202 admitted / 409 stale-claim-reroute-once / 503 unavailable). Selecting
   etcd always activates ownership and routing; there is no supported
   etcd-backed multi-replica mode without fencing.

4. **Persistent execution barriers enforce same-PR safety over automatic
   takeover.** Before any side-effecting step, the owner writes a persistent
   barrier bound to its exact claim generation (coordination epoch + key creation
   revision). A new owner may claim after a lease expiry but cannot start work
   while an older generation's barrier is unresolved; such a barrier is cleared
   only by an authenticated completion, otherwise the PR is marked **uncertain**
   and requires operator resolution. Atlantis does **not** promise exactly-once
   destructive execution; Terraform backend locking and provider idempotency
   remain necessary.

5. **Host-scoped project locking.** etcd uses a supplemental
   `ScopedProjectLockStore` keyed by the full `{VCS hostname, repo, path,
   project, workspace}` scope, because the legacy lock key
   (`models.GenerateLockKey`) omits the VCS hostname and would collide across
   hosts. BoltDB/Redis keep their existing keys; the legacy lock key is **not**
   retrofitted.

6. **Plan artifacts stay out of etcd.** `.tfplan` files, repository clones, and
   command logs remain in the existing `PlanStore`. etcd stores only locks,
   statuses, ownership, admission records, and execution barriers.

7. **Fixed voting membership, independently scalable compute.** *(Phase 2 — embedded.)*
   Production embedded deployments run a fixed StatefulSet of 3/5/7 **member replicas**
   (Atlantis + one etcd voter each, stable identity + PVC) plus zero-or-more
   independently autoscaled **client replicas** (Atlantis in external mode).
   Voter count is a quorum setting, never an HPA target. Membership changes are a
   documented operator runbook (add-learner → ticket → join → promote), not an
   Atlantis membership controller.

8. **Offline migration and recovery, fenced by a coordination epoch.** There is
   no live mixed BoltDB/etcd fleet and no dual-write. Migration is offline with a
   manifest, create-only import, count + checksum verification, and an atomic
   completion marker. A fresh coordination epoch is minted on init/migration/
   restore and stamped into every claim, admission, and barrier, so a restored or
   rolled-back snapshot cannot reuse a discarded generation. After a snapshot
   restore, a **recovery-quarantine** marker rejects every executable command
   until an operator reconciles and explicitly clears it.

9. **Explicit embedded lifecycle; unsafe options rejected.** *(Phase 2 — embedded.)*
   The embedded server
   requires an explicit lifecycle (`bootstrap|restart|join-existing|restore`) and
   startup purpose (`serve|maintenance`); it never derives lifecycle from
   directory emptiness or reachability. Unsafe native etcd options
   (force-new-cluster, disabled fsync, disabled strict reconfiguration, discovery
   bootstrap, auto-TLS) are hard-rejected at validation.

Production requires mTLS on every etcd, peer, and internal endpoint; HTTP is
permitted only for loopback under an explicit development flag.

### Alternatives considered and rejected

- **Redis plus an application-level lock/lease/ownership layer.** Reimplements,
  with weaker guarantees, the leases, fencing, and membership etcd already
  provides linearizably.
- **`hashicorp/raft` (raft-boltdb).** Heavier to operate for our key/value +
  lease/watch needs; an explicit non-goal.
- **A single leader that owns all execution.** Concentrates all Terraform work on
  one node (no compute scale-out) and makes the leader a failover bottleneck;
  rejected in favor of leader-free per-PR ownership.
- **SQL (e.g. Postgres advisory locks).** Adequate for locks but lacks first-class
  leases/watches for liveness-based ownership; a weaker fit for the coordination
  layer.
- **Status quo (document HA as unsupported).** Leaves BoltDB a single point of
  failure and Redis unsafe active-active — the gap this decision closes.

See the [proposal](../proposals/2026-09-etcd-ha-locking.md) for the full
alternatives analysis.

## Consequences

### Positive

- Atlantis survives the loss of a minority of voters with no manual failover, and
  ingress/execution scale horizontally via client replicas.
- The single-writer guarantee is preserved and strengthened: exactly one
  authoritative owner per PR, enforced by linearizable claims and execution
  barriers, failing closed on quorum or lease loss.
- Cross-VCS-host correctness is fixed for the etcd path (host-scoped locks).
- No new coupling for existing users: BoltDB remains the untouched default, and
  its tests pass unchanged.

### Negative / costs

- Operational complexity rises sharply: quorum, durable per-member volumes, TLS
  material, backups, restore drills, and a membership runbook. Embedded etcd needs
  firm CPU/memory reservations and low-latency disk so Terraform cannot starve it.
- **No exactly-once destructive execution.** When an owner is lost mid-apply, the
  outcome is reported as uncertain and requires operator reconciliation rather than
  automatic replay. This is an explicit, accepted trade-off.
- Availability now depends on etcd quorum: losing quorum makes all replicas
  unready and fails new commands closed (by design, but a real dependency).
- New dependency surface: `go.etcd.io/etcd` client/server pulls in gRPC and
  related modules.

### Neutral / scope boundaries

- **Embedded etcd (decisions 7 and 9) is a phase-2 deliverable.** Phase 1 supports
  external etcd only and rejects `--etcd-mode=embedded` at validation; the embedded
  runtime, member StatefulSet topology, and lifecycle/membership/restore tooling land
  in a later phase. The dependency surface is correspondingly smaller in phase 1: the
  shipped binary does not link the embedded etcd server.
- Some capabilities are deferred behind the same fail-closed contract: drift
  detection is rejected in combination with active-active etcd until it gets a
  distributed exclusion design; per-project-step barriers, in-flight subprocess
  reaping, and full owner-routing of the lock-UI delete are tracked residuals that
  do not weaken the cross-generation invariant. See the implementation plan for
  the live status and residual list.
- The first release ships a tested operator runbook for membership and recovery,
  not automated controllers.

### Verification gate

Production support is gated on a route-inventory test proving every executable
entry point passes through owner resolution, plus the consistency,
failure-injection, embedded-lifecycle, and migration/restore suites enumerated in
the design document's acceptance criteria. BoltDB and Redis behavior must remain
unchanged.
