# Proposal: High-Availability Atlantis via an etcd Locking and Coordination Backend

- **Status:** Proposed
- **Date:** 2026-09-09
- **Authors:** Atlantis maintainers
- **Decision record:** [ADR 0003](../adr/0003-etcd-high-availability-locking-backend.md)
- **Detailed design and implementation:** provided with the code in the
  implementation pull request.

## 1. Summary

Add an opt-in etcd backend (`--locking-db-type=etcd`) that lets Atlantis run as a
fault-tolerant, active-active service without weakening its single-writer safety
guarantees. etcd can be embedded in the Atlantis binary or connected to as an
external cluster. BoltDB remains the default and is untouched.

**Delivered in phases.** Phase 1 (current) ships **external etcd only** — the full
active-active coordination stack against an operator-provided external cluster;
`--etcd-mode=embedded` is rejected until phase 2. Phase 2 (future) adds the embedded
in-process voter and its lifecycle/topology tooling.

## 2. Motivation

Atlantis today is effectively a singleton for the pull requests it manages:

- **BoltDB** (default) is a local file — a single point of failure. A node loss
  stalls all automation until the pod is rescheduled and its volume reattached.
- **Redis** can be shared, but Atlantis has no coordination above it, so two
  replicas could plan/apply the *same* PR at once. Redis alone is not safe
  active-active.

Teams running Terraform automation as critical infrastructure want:

1. **No single point of failure** — survive a node/pod loss with no manual
   failover.
2. **Horizontal scale** — add ingress/execution capacity for large fleets.
3. **Preserved safety** — never let two replicas act on the same PR; never
   silently double-apply.

No current backend delivers (1)+(2) without sacrificing (3). etcd's linearizable
transactions, leases, and watches are the right primitives to close that gap.

## 3. Goals

1. Preserve BoltDB as the unchanged default for single-replica deployments.
2. Support a dedicated external etcd cluster, and a production-grade embedded etcd
   (durable storage, TLS, controlled membership, backup/recovery, rolling upgrade).
3. Scale Atlantis ingress and computation independently of etcd voting membership.
4. Keep every replica active while enforcing exactly one authoritative owner per
   PR, with a takeover barrier whenever earlier execution has an uncertain outcome.
5. Fail closed when ownership or lock authority cannot be established.
6. Keep Terraform plan artifacts and working directories out of etcd.
7. Provide explicit migration and rollback without dual-writing BoltDB and etcd.

## 4. Non-goals

- Replacing or rewriting BoltDB/Redis, or dual-writing between backends.
- Storing `.tfplan` files, clones, logs, or Terraform state in etcd.
- Running an etcd voter in every autoscaled replica.
- Promising exactly-once Terraform or provider side effects.
- A distributed command queue, a dedicated worker role, or an idle follower role.
- Using a Kubernetes control-plane etcd as application storage.

## 5. Proposed solution (overview)

Selecting `--locking-db-type=etcd` activates a new `server/core/etcd` backend and,
always, active-active PR ownership and routing. See the ADR for the locked
decisions; the shape is:

- **Two modes, one client path.** `external` connects to an etcd cluster;
  `embedded` runs one etcd voter in-process. Both share the same client and the
  same database/ownership/fencing code.
- **Fenced PR ownership.** Each process holds a renewable session lease; each PR is
  owned by exactly one process via a key-absent claim transaction. A command
  handler resolves the owner and executes locally or forwards a credential-free,
  mTLS-authenticated envelope to the owner (202/409/503 semantics, reroute once).
- **Execution barriers.** Before any side-effecting step, the owner writes a
  persistent barrier bound to its exact claim generation (coordination epoch + key
  creation revision). A new generation cannot start work while an older barrier is
  unresolved; an unresolvable one marks the PR **uncertain** for operator action.
- **Host-scoped locks.** A supplemental scoped-lock store keys project locks by the
  full VCS-host-qualified scope (fixing a cross-host collision the legacy key has).
- **Topology.** A fixed StatefulSet of 3/5/7 member replicas (Atlantis + one voter,
  stable identity + PVC) plus zero-or-more autoscaled client replicas (external
  mode). A Service fronts both. Voter count is a quorum setting, not an HPA target.
- **Migration & recovery.** Offline, manifested migration (no dual-write); a fresh
  coordination epoch minted on init/migrate/restore; a recovery-quarantine that
  blocks all execution after a snapshot restore until an operator reconciles.

### Failure behavior (fail closed)

| Situation | Behavior |
| --- | --- |
| A minority of voters down | Backend and unrelated-PR work continue via quorum |
| Quorum lost | All replicas unready; new work rejected; no fallback |
| Ownership lease lost | Replica unready; queued work cancelled; subprocess cancelled where possible |
| Older-generation barrier unresolved | New owner may claim, but that PR is blocked until resolved |
| Owner dies after 202, before execution | Admission marked uncertain; never auto-replayed |
| Snapshot restored | Recovery quarantine rejects all commands until an operator clears it |

## 6. Alternatives considered

- **Redis + an application lock layer.** Would require building leases, fencing,
  ownership, and membership semantics on top of Redis — reimplementing what etcd
  already provides linearizably, with weaker guarantees. Rejected.
- **`hashicorp/raft` (raft-boltdb).** A raft log embedded per replica; heavier to
  operate for our key/value + lease/watch needs and less mature for this shape than
  etcd's client/server. Explicitly a non-goal.
- **A single leader that owns all execution.** Simpler ownership, but concentrates
  all Terraform work on one node (no compute scale-out) and makes the leader a hot
  spot and failover bottleneck. Rejected in favor of leader-free per-PR ownership.
- **Postgres/SQL with advisory locks.** Viable for locks, but lacks first-class
  leases/watches for liveness-based ownership and adds a second stateful system;
  weaker fit than etcd for the coordination layer.
- **Do nothing / document HA as unsupported.** Leaves BoltDB as a SPOF and Redis as
  unsafe-active-active — the status quo this proposal exists to fix.

## 7. Rollout and migration

1. **Phase 1:** Ship external mode first (validates schema/consistency without
   embedded lifecycle risk). Single-replica external etcd is usable as a drop-in
   backend.
2. **Phase 1:** Add ownership/transport, execution fencing, and complete entry-point
   coverage (a route-inventory is the production release gate).
3. **Phase 2 (future):** Add the embedded runtime, then migration/operations and
   Helm/deployment integration. `--etcd-mode=embedded` is rejected at validation
   until this phase lands.

**Cutover** (per the design's runbook): stop executable ingress and drain; back up
the source; export locks/statuses/global locks; form the target cluster in
`bootstrap`/`maintenance`; validate the namespace is empty; run the offline,
manifested, checksum-verified import; atomically write schema + deployment/epoch +
completion marker; restart members in `restart`/`serve`; start the fleet with
`--locking-db-type=etcd`; re-plan work that depended on local plan artifacts.
Rollback is trivial before etcd receives new writes and, after, requires a reverse
export.

## 8. Risks and mitigations

| Risk | Mitigation |
| --- | --- |
| No exactly-once destructive execution | Explicit uncertain-outcome policy + operator reconciliation; rely on Terraform backend locking and provider idempotency |
| etcd quorum becomes an availability dependency | Fail closed by design; PDB `maxUnavailable: 1`; anti-affinity/topology spread; tested consensus latency bounds |
| Terraform starves embedded etcd of CPU/disk | Firm reservations, low-latency durable storage, spread voters across nodes/failure domains |
| Operational complexity (membership, backup, restore) | Documented, tested runbooks; recovery quarantine; restore drills; a separate operator identity distinct from the runtime identity |
| Security exposure of the internal transport | mTLS + constant-time token, an advertise-destination allowlist (SSRF guard), network policy separating traffic classes |
| Data-model or schema drift on upgrade | Versioned record envelopes + an expand/migrate/contract schema-phase protocol; a coordination epoch fences stale generations |

## 9. Open questions / deferred work

- Distributed drift detection under active-active etcd (currently rejected at
  config validation until designed).
- Finer per-project-step execution barriers, in-flight subprocess reaping on lease
  loss, and owner-routing of the lock-UI delete (tracked residuals; the
  whole-command barrier already holds the cross-generation invariant).
- Provider-delivery-ID dedup for VCS providers that do not expose a stable comment
  ID (Bitbucket, Azure DevOps).

See the implementation plan for the authoritative, live status of each item.
