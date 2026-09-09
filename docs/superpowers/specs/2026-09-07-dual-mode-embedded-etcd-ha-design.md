# Dual-Mode Embedded etcd High-Availability Design

Date: 2026-09-07

Status: approved architecture; implementation not started

## Summary

Atlantis will gain an opt-in etcd locking and coordination backend with two
runtime modes:

- **Embedded mode:** an Atlantis process starts and owns one embedded etcd
  voting member.
- **External mode:** an Atlantis process connects to an existing etcd cluster,
  including a cluster hosted by embedded-mode Atlantis replicas.

Both modes construct the same long-lived etcd client and use the same Atlantis
database, ownership, routing, and fencing implementations. BoltDB remains the
default backend and its behavior is unchanged.

Production embedded deployments use two logical replica types built from the
same Atlantis binary:

- A fixed StatefulSet of 3, 5, or 7 **member replicas**. Every member replica
  runs the complete Atlantis ingress and execution stack and embeds one etcd
  voter.
- Zero or more independently scalable **client replicas**. Every client replica
  runs the same Atlantis ingress and execution stack in external etcd mode.

There are no etcd-only followers and no leader-only Atlantis role. The etcd
leader has no special Atlantis execution responsibility. Every Atlantis
replica can accept ingress, own pull requests, and execute Terraform work.

## Goals

1. Preserve BoltDB as the unchanged default for single-replica deployments.
2. Support a dedicated external etcd cluster without embedding a server.
3. Make embedded etcd production-supported with durable storage, TLS,
   controlled membership, backup, recovery, and rolling-upgrade procedures.
4. Scale Atlantis ingress and computation independently of etcd voting
   membership.
5. Keep every Atlantis replica active while enforcing one authoritative owner
   generation for all commands belonging to the same pull request, including a
   takeover barrier whenever earlier execution has an uncertain outcome.
6. Preserve existing project and workspace concurrency controlled by
   `parallel-plan`, `parallel-apply`, and `parallel-pool-size`.
7. Fail closed when ownership or lock authority cannot be established.
8. Keep Terraform plan artifacts and working directories outside etcd.
9. Provide explicit migration and rollback procedures without dual-writing
   BoltDB and etcd.

## Non-goals

- Replacing or rewriting the existing BoltDB backend.
- Using `hashicorp/raft-boltdb`.
- Storing `.tfplan` files, repository clones, command logs, or Terraform state
  in etcd.
- Running an etcd voter in every autoscaled Atlantis replica.
- Automatically changing etcd membership in response to an HPA or an ordinary
  StatefulSet replica-count change.
- Promising exactly-once Terraform or provider-side effects.
- Using a Kubernetes control-plane etcd cluster as application storage.
- Adding a distributed command queue, dedicated worker role, or idle Atlantis
  follower role.

## Locked architecture decisions

### Backend compatibility

`--locking-db-type` gains the value `etcd`. Its existing default remains
`boltdb`. Existing BoltDB and Redis construction and behavior remain unchanged.
Atlantis does not fall back to another backend if etcd is unavailable.

The existing `server/core/db.Database` interface remains the compatibility
contract for BoltDB, Redis, pull/project status, global command locks, `Ping`,
and `Close`. Its legacy project-lock methods omit VCS hostname, so the etcd HA
path additionally uses a host-aware scoped-lock interface. That interface takes
an exact project scope {VCS hostname, repository, path, project, workspace} and
pull scope {VCS hostname, repository, pull number} for acquisition, lookup,
conditional unlock, and pull cleanup. Etcd implements both contracts; all etcd
project-lock call sites use the scoped contract. BoltDB and Redis keep their
existing keys and behavior unchanged.

Plan persistence continues through the independent
`server/core/planstore.PlanStore` abstraction.

### Dual runtime, single client path

Both etcd modes return one shared backend object that owns one long-lived
`*clientv3.Client`:

```text
External runtime ----+
                     +--> shared etcd client --> EtcdDatabase
Embedded runtime ----+                      +--> EtcdOwnershipStore
                                            +--> readiness probe
```

The ownership store borrows the client and closes before the shared backend.
`EtcdDatabase.Close` is the designated idempotent delegate to
`Backend.Close`; the backend is the only component that closes the client.
In embedded mode, the same backend closes the embedded server after all
client-side users and the client have stopped. Partial-startup error paths call
the same idempotent backend close operation.

### Production topology

```text
                               Atlantis Service
                                      |
                 +--------------------+--------------------+
                 |                    |                    |
       member replicas          member replicas       client replicas
       Atlantis + voter         Atlantis + voter      Atlantis + client
       stable identity/PVC      stable identity/PVC   independently scaled
                 \                    |                    /
                  +--------- one logical etcd cluster ----+
                                      |
                         locks, status, PR ownership

                         PlanStore remains separate
```

Member replicas and client replicas use the same image and expose the same
Atlantis endpoints. A Service or load balancer selects both groups. The logical
type is derived from `etcd-mode`; a separate Atlantis leader or worker flag is
not introduced.

The total active Atlantis capacity is:

```text
voter count + client replica count
```

Only client replica count is eligible for ordinary autoscaling. Member replicas
must use stable network identities, durable volumes, anti-affinity, and a pod
disruption budget that preserves quorum.

## Configuration contract

### Common etcd settings

```text
--locking-db-type=etcd
--etcd-mode=external|embedded
--etcd-deployment-id=<stable UUID>
--etcd-namespace=/atlantis
--etcd-ca-file=<path>
--etcd-cert-file=<path>
--etcd-key-file=<path>
--etcd-server-name=<name>
--etcd-username=<name>
--etcd-password-file=<path>
--etcd-request-timeout=5s
--etcd-startup-timeout=5m
--etcd-allow-insecure-dev=false
```

The deployment ID is generated once per logical Atlantis installation and
remains stable across process restarts and database migrations. The namespace is
normalized once and all Atlantis keys are written below a versioned child
prefix.

Production mode requires HTTPS endpoints, a trusted CA, and a complete client
certificate/key pair. Authorization uses either a client-certificate identity
mapped to an etcd RBAC user or a username with a password file. Supplying a
username requires a password file; when supplied, username authentication is
used while mTLS remains mandatory. Secrets and private-key material use
secret-backed environment or file inputs and are never logged or rendered into
pod arguments.

The request timeout bounds individual database, ownership, and readiness RPCs.
The startup timeout bounds initial external connectivity and embedded quorum
formation; deployment startup probes must allow at least the same interval.
The insecure-development flag permits HTTP only on loopback addresses and is
rejected for any non-loopback listen or advertise URL.

### External mode

External mode requires:

```text
--etcd-endpoints=https://member-0:2379,https://member-1:2379,...
```

It rejects embedded-only fields. Every production endpoint must use HTTPS and
pass hostname verification. `clientv3.New` is nonblocking, so startup and
readiness perform a bounded linearizable operation rather than treating
successful client construction as connectivity proof.

### Embedded mode

Embedded mode requires:

```text
--etcd-embedded-config-file=<path>
--etcd-embedded-voter-count=3
--etcd-embedded-lifecycle=bootstrap|restart|join-existing|restore
--etcd-embedded-startup-purpose=serve|maintenance
--etcd-embedded-identity-file=<path>
--etcd-embedded-join-endpoints=https://member-0:2379,...  # join-existing only
--etcd-embedded-membership-ticket-file=<path>             # join-existing only
--etcd-embedded-restore-manifest-file=<path>              # restore only
```

`etcd-embedded-voter-count` has these rules:

- Default: 3.
- Minimum: 3.
- Supported values: 3, 5, or 7.
- Even values and values above 7 are rejected.
- It is the desired final number of voting members and applies only to
  embedded member replicas. Temporary learners do not count toward it.

The lifecycle mode is mandatory:

- `bootstrap` requires an empty data directory and an identity manifest in
  the new state. The manifest contains the deployment ID, bootstrap generation,
  and complete original member set. Bootstrap is refused after the manifest has
  been activated or when storage is nonempty. Production bootstrap permits only
  the `maintenance` startup purpose.
- `restart` is the steady-state mode. It requires nonempty storage and an
  active identity manifest whose cluster ID, member name, and member ID match
  the data directory. Empty or mismatched storage is refused.
- `join-existing` requires empty storage, an active cluster identity, and a
  one-time membership ticket produced by an explicit learner-add operation.
  The ticket binds the expected cluster ID, member ID, name, and peer URL.
  Missing, reused, or mismatched tickets are refused.
- `restore` requires nonempty storage produced by the documented offline
  snapshot-restore tool plus a pending recovery manifest outside the member
  PVC. The manifest binds the snapshot hash, deployment ID, fresh recovery
  generation UUID, new cluster/member IDs, member name, and complete peer set.
  Any mismatch, reused manifest, or ordinary pre-restore data directory is
  refused. Restore mode permits only the `maintenance` startup purpose.

The join endpoints and membership-ticket fields are required only in
`join-existing` mode and are rejected in the other modes. An operator identity
with membership privileges creates the learner, then writes a one-time ticket
record below the Atlantis namespace. The ticket file carries an opaque nonce;
the record binds its hash to the deployment ID, cluster ID, member ID, name, peer
URL, and transition generation.

Before starting the embedded server, the joiner constructs its shared long-lived
client against the join endpoints, validates the active deployment and ticket
record, and atomically consumes the ticket. The runtime Atlantis identity does
not receive membership-administration privileges. A failed start does not make
the ticket reusable automatically; recovery requires explicit operator
reconciliation against live membership.

The startup purpose defaults to `serve`, but production configuration accepts
only `bootstrap`/`maintenance`, `restore`/`maintenance`,
`restart`/`serve`, or `join-existing`/`serve`. Maintenance starts and probes the
embedded cluster but does not initialize or consume the Atlantis namespace,
construct database/ownership adapters, establish claims, or serve public
executable ingress. It is used for every initial production bootstrap, for
BoltDB import into a new embedded cluster, and for safe restore initialization.

After initial bootstrap quorum forms, deployment tooling verifies every observed
cluster/member ID, initializes the empty namespace for a fresh installation or
completes the migration, atomically activates the external identity manifest,
and only then restarts members in `restart`/`serve`. No bootstrap process creates
ownership adapters or becomes ready for Atlantis traffic. After restore quorum
forms, recovery tooling validates every pending manifest and atomically installs
new active identity manifests before switching members to restart mode. Identity
manifests are retained outside individual member PVCs so accidental PVC loss
cannot be mistaken for a new cluster.

The embedded configuration contains the member name, data directory, listen
and advertise URLs, the complete initial peer set, cluster token and state, and
separate client and peer TLS settings. Atlantis parses the configuration
without using helpers that can terminate the host process, validates it, and
then constructs `embed.Config`.

In bootstrap mode, the number of unique members in the initial peer set must
equal the configured voter count. Each original member receives the same peer
set and cluster token, a unique member name, and its own advertised peer URL.
Production peer and client listeners use separate TLS identities and require
client-certificate authentication. Configuration validation hard-rejects
`force-new-cluster`, `unsafe-no-fsync`, disabled strict reconfiguration checks,
automatic self-signed certificates, discovery-based bootstrap, and any
`initial-cluster-state` inconsistent with the selected lifecycle. Initial
corruption checking is required for production member startup.

Atlantis never derives lifecycle mode from network reachability or directory
emptiness. It never reuses another member data directory and never starts a
removed member from its old storage. If initial bootstrap fails after writing
member data but before the bootstrap generation becomes active, the incomplete
cluster is not resumed automatically; the operator must follow the documented
whole-bootstrap recovery procedure.

### Ownership and routing settings

Selecting `locking-db-type=etcd` always activates active-active PR ownership
and routing. There is no separate routing enable flag and no supported
etcd-backed multi-replica mode without ownership fencing.

```text
--replica-id=<stable replica name>
--replica-advertise-url=https://<internal address>
--replica-advertise-allowlist=<host-or-CIDR>[,...]
--internal-command-token-file=<path>
--internal-command-ca-file=<path>
--ownership-ttl-seconds=30
```

Replica ID defaults to the pod hostname but must resolve to a stable, unique
value. Advertise URL, destination allowlist, token file, and internal CA are
mandatory in production. The URL must use HTTPS, match the allowlist, and pass
hostname verification. The development-only insecure flag permits HTTP only
when both source and destination are loopback.

Every process also generates a unique instance ID so a restarted process cannot
inherit previous claims. Ownership TTL defaults to 30 seconds and has a minimum
of 10 seconds. Any missing, duplicate, malformed, or insecure resolved setting
is a startup error.

## Voter-count lifecycle

Voter count is a deployment and quorum setting, not a computation-scaling
setting.

### Initial bootstrap

For a new embedded cluster, deployment tooling renders exactly the configured
number of stable member identities and peer endpoints. Kubernetes deployments
use:

- A headless peer Service with not-ready addresses published.
- A StatefulSet with `podManagementPolicy: Parallel` and
  `updateStrategy: OnDelete`.
- One persistent volume claim per member.
- Stable ordinal DNS names for peer and client URLs.
- Pod anti-affinity and topology spreading.
- A disruption budget with `maxUnavailable: 1` for every supported voter count;
  runbook preflight also permits only one manual member deletion at a time.

Parallel pod startup is required because ordered readiness can wait for the
first member to become ready before starting the peers needed to form quorum.

### Scaling Atlantis computation

Client replicas may scale from zero to any operationally supported count. They
connect to the complete member endpoint list and do not participate in Raft.
Adding or removing client replicas does not change etcd membership.

### Changing voter count

`voterCount` is the desired final voting membership. A durable membership
transition record contains source count, target count, direction, generation,
and next permitted ordinal. Only that explicit record permits live membership to
differ from the configured target while a learner or even voter count exists;
those intermediate states are never accepted as steady state.

Changing voter count is never an ordinary rollout. For `3 -> 5`, the operator:

1. Verifies cluster health and creates a `3 -> 5` transition record before
   changing the configured target to 5. `OnDelete` prevents that template change
   from automatically restarting existing voters.
2. Adds exactly one target member through the live cluster as a learner.
3. Issues a one-time membership ticket bound to the transition, returned
   cluster/member identity, name, and peer URL.
4. Grows the StatefulSet by exactly one ordinal. Only that new ordinal starts in
   `join-existing` mode with fresh storage.
5. Verifies identity, health, and Raft-log catch-up, then promotes it to voter.
6. Records completion of that ordinal and repeats steps 2-5 for the next ordinal.
7. While the transition remains active, rolls retained voters one at a time if
   their rendered configuration must converge. It verifies that live membership,
   every running member's configured target, and the StatefulSet all equal 5
   before atomically marking the transition complete.

Scale-down creates the matching transition first, then repeats this safety order:
verify quorum, remove the highest target ordinal from membership, reduce the
StatefulSet by exactly that ordinal, and retire its PVC. Removed member storage is
never restarted. Retained voters converge one at a time while the transition
remains active. It completes only when live membership, every running member's
configured target, StatefulSet replica count, and retained identities all agree.

The first implementation supplies a documented and tested operational runbook
using the etcd membership API or `etcdctl`; it does not introduce an Atlantis
membership controller. Deployment admission and preflight tooling reject a
member replica/template change unless an active transition authorizes its exact
target and next ordinal. The membership-administration identity is separate from
the runtime Atlantis identity.

## Components and boundaries

### Etcd runtime/backend

A new `server/core/etcd` package owns common configuration, key construction,
client construction, health probes, and shutdown.

Its runtime boundary provides:

```go
type Backend interface {
    Client() *clientv3.Client
    Ready(context.Context) error
    Close() error
}
```

The backend owns the client and embedded runtime. The database is its designated
close delegate, and the ownership adapter borrows the client. Consumers do not
close the client directly.

The external runtime constructs and probes the client. The embedded runtime
validates configuration, starts `embed.Etcd`, waits for server readiness,
constructs the same client type, and probes the cluster.

### Etcd database adapter

`EtcdDatabase` implements the common `db.Database` status, global-lock, health,
and lifecycle behavior without changing BoltDB. Etcd project locking is wired
through a supplemental `ScopedProjectLockStore` because the legacy project
model and `UnlockByPull` API omit VCS hostname.

The scoped store uses two explicit value types:

```text
ProjectScope = {VCS hostname, repository, path, project, workspace}
PullScope    = {VCS hostname, repository, pull number}
```

The etcd locking client carries these scopes from `PullRequest.BaseRepo.VCSHost`
through acquire, get, UI lock ID, conditional unlock, and pull cleanup. Etcd
mode never invokes an ambiguous legacy project-lock operation. BoltDB and Redis
continue using the existing lock client, serialized records, and key format.
Etcd UI lock IDs are versioned opaque encodings that include VCS hostname;
their decoder cannot fall back to the legacy parser.

All etcd operations use bounded contexts internally because the existing common
database interface does not carry caller contexts.

Required transaction semantics include:

- Scoped project and global lock acquisition compares `CreateRevision(key) == 0`.
- Conditional unlock compares the exact owner value and observed revision.
- Pull-status and project-status merges use revision-based compare-and-swap
  retry loops to avoid lost concurrent updates while preserving the existing
  base-branch, head-commit, and corrupt-status replacement behavior.
- `UnlockByPullScope` first creates a pull-scoped lifecycle record in `cleaning`
  state that every etcd project-lock acquisition transaction checks. It ranges
  only over the project-lock namespace, decodes the exact identity, and
  conditionally deletes each matching revision before a final linearizable empty
  check. Manual cleanup returns the lifecycle to `open`; pull-close cleanup
  advances it to a persistent `closed` generation.
- A reopen event follows the same owner-routed path, refetches authoritative VCS
  state, and compare-and-swap advances `closed` to a new `open` generation.
  Project-lock acquisition records and compares that generation. Delayed close
  or reopen deliveries must revalidate current VCS state and cannot overwrite a
  newer lifecycle generation.
- Read-only paginated scans pin every page to the first response revision. Large
  deletes respect transaction operation limits and never infer identity from key
  substrings.
- `Ping` performs a bounded linearizable range operation.
- `Close` is idempotent.

Persistent database locks and statuses do not use leases. They must survive an
Atlantis process crash so plan/apply safety state is retained.

### Ownership store

The ownership identity is:

```text
{normalized VCS hostname, repository full name, pull-request number}
```

Each process creates one renewable session lease. Every PR ownership key won by
that process is attached to the session lease. The stored record contains:

- Stable replica ID.
- Unique process instance ID.
- Internal advertise URL.
- Random claim ID.
- Claim timestamp.
- Coordination epoch UUID.

The key's etcd creation revision and coordination epoch UUID form the fencing
generation. Claiming uses one transaction that succeeds only when the ownership
key does not exist. Admission and release compare the exact record and revision.
Reusing a replica ID never adopts a claim from an older process.

Watches are only wake-up and cache-invalidation mechanisms. They are not an
authorization source. A compacted or canceled watch performs a linearizable
resynchronization and restarts. Every execution admission uses a linearizable
read or transaction.

### Command dispatch and local fencing

Every replica accepts public ingress. A command handler resolves or creates the
PR owner claim:

1. If the local process owns the exact claim, it proceeds locally.
2. If another process owns the claim, it forwards a credential-free envelope
   to that process's internal advertise URL.
3. The receiving process validates internal authentication and the exact claim
   before admission.
4. A stale claim returns HTTP 409. The ingress replica refreshes ownership and
   reroutes once.
5. Backend or execution unavailability returns HTTP 503.

The internal transport carries an immutable command identity, limits body and
error sizes, does not follow cross-origin redirects with credentials, and allows
only configured internal destinations. The canonical identity includes source
kind, normalized VCS hostname, and source delivery or request ID. Webhooks retain
their provider delivery ID; API retries require an explicit idempotency key, while
requests without one receive a new ingress-generated ID. Internal retries always
preserve the identity. VCS credentials are never forwarded.

Internal forwarding may be duplicated or lost by transport failures. The
receiving owner atomically validates its exact claim and creates a persistent
`reserved` admission record keyed by immutable command identity. It then
registers work idempotently in its instance-local drainer, compare-and-swap
transitions the record to `scheduled`, and only then returns HTTP 202. Runner
start transitions it to `running`; completion writes a terminal result.

A duplicate under the same exact claim and process instance reconciles a
`reserved` record with the idempotent local registry: it may retry registration
only when absence is proven, otherwise it transitions the record to `uncertain`.
A different instance or owner generation never schedules the reservation.
Failure between local registration and the `scheduled` transition is likewise
reconciled from the local registry or marked uncertain, never blindly duplicated.
Duplicates of `scheduled`, `running`, uncertain, or terminal records return the
stored admission state or result and never create a second local registration.

The admission record is an audit and idempotency record, not a distributed
queue: no other replica automatically consumes it. If the owner dies after 202
but before execution, the next reconciliation marks the record uncertain and
requires a new user/VCS action. An ambiguous timeout for a destructive command
is never automatically replayed.

### Execution admission boundaries

The exact claim generation is propagated in the command and project contexts.
Atlantis checks it:

- Before resetting stale local pull state.
- Before asynchronous scheduling.
- After every local working-directory or project-lock wait.
- Before cloning, merging, restoring, or loading plans.
- Immediately before each project workflow begins.
- Before plan, apply, import, state removal, policy, unlock, cancellation, and
  version execution paths.

Before starting every workflow subprocess or side-effecting step, the owner uses
one transaction to compare its exact claim and create a persistent execution
barrier. Barriers are keyed by pull scope, owner generation, and execution ID so
parallel projects within the same valid owner generation remain supported. A
barrier is cleared only after the child process has exited and the step result is
recorded.

Claim loss prevents subsequent steps, cancels queued work, and cancels or reaps
subprocesses where possible. It also makes the replica unready until a new
authoritative process session is established. A newly acquired project lock is
unwound if admission fails before execution.

A new owner may acquire the PR claim after lease expiry, but it cannot start any
step while an unresolved execution barrier from an older generation exists.
The old owner may clear the barrier only through an authenticated completion
acknowledgement after it has stopped the process. Otherwise the PR is marked
uncertain and requires explicit operator resolution. Other PRs continue
normally. This barrier favors same-PR safety over automatic takeover when an
earlier execution cannot be positively fenced.

The barrier prevents two Atlantis generations from starting overlapping work;
it cannot undo a Terraform provider request that has already crossed the
process boundary or prove that an external system stopped after process exit.
Terraform backend locking, provider idempotency, and operator reconciliation
remain necessary. Atlantis therefore does not promise exactly-once destructive
execution.

### Route coverage

The final HA route inventory includes:

- All actionable comment commands, including plan, apply, unlock, cancel,
  version, policy approval, import, and state operations.
- Autoplan.
- Pull-close cleanup and reopen lifecycle transition.
- Positive-PR `/api/plan` and `/api/apply` requests.
- Web lock deletion when it affects owner-local plan state.
- Review-triggered executable work when enabled.

Positive-PR API calls use synchronous response proxying rather than the
asynchronous command transport. Complete routing for every entry point in this
inventory is a production release gate for `locking-db-type=etcd`; no endpoint
may execute directly on an arbitrary replica or be documented as supported
before its owner-aware path is implemented.

Drift detection uses repository/ref identities and process-local storage in the
current implementation. It requires a separate distributed exclusion and
storage design. Until that work is complete, configuration validation must
reject drift enablement with active-active etcd routing rather than silently
permit duplicate remediation.

## Data model

All keys are below a normalized namespace and version prefix:

```text
<namespace>/v1/meta/schema
<namespace>/v1/meta/deployment
<namespace>/v1/meta/migrations/<migration-id>
<namespace>/v1/meta/recoveries/<recovery-id>
<namespace>/v1/meta/recovery-quarantine
<namespace>/v1/meta/capabilities/<instance-id>
<namespace>/v1/meta/schema-upgrades/<upgrade-id>
<namespace>/v1/meta/membership-transitions/<transition-id>
<namespace>/v1/meta/membership-tickets/<ticket-id>
<namespace>/v1/db/project-locks/<encoded-project-scope>
<namespace>/v1/db/pull-lifecycle/<encoded-pull-scope>
<namespace>/v1/db/pulls/<encoded-pull-scope>
<namespace>/v1/db/global-locks/<encoded-lock-name>
<namespace>/v1/ownership/pulls/<encoded-pull-scope>
<namespace>/v1/commands/<epoch>/<encoded-delivery-id>
<namespace>/v1/execution/pulls/<encoded-pull-scope>/<generation>/<execution-id>
```

Tuple identities are canonically serialized and encoded with unpadded URL-safe
base64. Code must not depend on path escaping, substring matching, or glob
semantics. Values use versioned JSON envelopes. Decoding is strict for a
recognized version; a binary must explicitly implement every older and newer
record version it claims to read.

The schema marker identifies Atlantis and the schema version. The deployment
record contains the configured deployment ID and an opaque coordination epoch
UUID. A cryptographically random epoch is created at initial namespace setup and
rotated by every migration and restore; recovery tooling supplies a UUID created
outside the snapshot so rolling back or restoring the same snapshot cannot reuse
a discarded generation. Every claim, command admission, and execution barrier
includes that epoch. Startup refuses an incompatible schema, mismatched
deployment ID, or incomplete migration or restore marker. Operators assign a
distinct namespace to each logical Atlantis deployment.

A namespace is genuinely empty only when no keys exist below its normalized
prefix. A fresh empty namespace may be initialized atomically. If any data key
exists without the matching schema, deployment record, and completed
migration-or-restore state, startup refuses it; absence is never interpreted as
successful completion of a partially started operation.

Only ephemeral coordination keys—PR ownership claims and replica capability
records—carry the process session lease. Project locks, pull/project status,
global command locks, command-admission records, execution barriers, and
migration metadata remain persistent until their explicit lifecycle completes.
Completed command records are compare-and-swap deleted after a fixed 24-hour
deduplication window. Successful execution barriers are deleted immediately;
uncertain barriers persist until explicit resolution.

### Schema and rolling-version compatibility

Every Atlantis build declares minimum/maximum readable schema versions and its
supported write-version set. The stored schema phase selects exactly one active
write version. Schema changes use an expand/migrate/contract protocol:

1. Deployment tooling writes an upgrade manifest that pins eligible image
   revisions and freezes replica scaling for the schema-phase transition.
2. The expanding release reads both N and N+1 records but continues writing N.
3. Each serving member and client replica atomically publishes a leased
   capability record that includes its observed schema-phase generation.
   Operator preflight verifies the pinned replica set and that every ready
   replica can read N+1, then writes a capability certificate containing the
   expected-set hash and acknowledged capability revisions. The phase transaction
   compares only the immutable upgrade manifest and certificate revisions before
   enabling N+1 writes, so activation does not exceed etcd transaction limits as
   client replicas scale. Deployment admission prevents an N image from starting
   in this interval.
4. Every database read, write, command admission, and ownership transaction
   compares the caller's observed schema-phase generation. On mismatch, the
   replica stops admission, becomes unready, reloads the phase, and resumes only
   if its declared capabilities cover the active read and write versions.
5. An idempotent, manifest-backed migration rewrites old records while both
   versions remain readable.
6. Contracting away N is deferred to a later release after migration verification
   and the documented rollback window; scaling resumes after the phase is
   committed.

Capability publication itself compares the current phase, so publication and
phase advancement are ordered by etcd. A stale process cannot successfully
publish for one phase and transact in another. The first release that introduces
this protocol makes no schema change; every later N baseline must already
implement the phase comparison before N+1 activation is permitted. A binary
outside the stored schema's readable range refuses readiness. Schema phase
changes are serialized with a migration lock and never coincide with
coordination-backend cutover or snapshot recovery.

## Plan ownership and takeover

Plan artifacts remain in the existing PlanStore:

- With local PlanStore, owner loss requires a new plan before apply.
- With external S3-compatible PlanStore, a new owner may restore plans only
  after cloning the default and plan-bearing workspaces and verifying the plan's
  recorded head commit.

On every new local ownership generation, the new owner waits for older local
generation work, clears stale pull-local state exactly once, clones the required
workspaces, restores external plans, and then uses the existing plan validation
path. Etcd never stores plan binaries.

## Startup and readiness

### External runtime

1. Validate mode, deployment ID, endpoint, namespace, authentication, and TLS
   configuration.
2. Construct one client using all configured endpoints.
3. Perform a bounded linearizable probe.
4. Atomically initialize a genuinely empty namespace, or validate its schema,
   deployment ID, coordination epoch, and completed migration/restore state.
5. Create the database and ownership adapters.
6. Establish the process ownership session for the validated epoch.
7. Mark `/readyz` successful only when the database and ownership session are
   authoritative.

### Embedded runtime

1. Validate embedded configuration, lifecycle, startup purpose, identity or
   pending recovery manifest, and data-directory state.
2. In join-existing mode, build the shared client against the join endpoints,
   perform a bounded linearizable probe, validate cluster/deployment identity,
   and atomically consume the matching active-transition ticket.
3. Call `embed.StartEtcd`.
4. Wait on `Server.ReadyNotify`, `Err`, cancellation, and the five-minute
   default startup timeout.
5. In bootstrap, restart, and restore modes, build the shared client; in join
   mode, reuse the pre-start client. Perform a bounded linearizable probe.
6. For `maintenance`, keep Atlantis `/readyz` false, expose only process
   liveness/metrics, skip namespace initialization and all adapters/claims, and
   continue monitoring `Err` while operator tooling initializes a fresh
   namespace, completes migration, or performs recovery.
7. For `serve`, require restart or join mode and validate schema, deployment ID,
   coordination epoch, completed initialization/migration/restore state, active
   identity manifest, and recovery-quarantine state. Bootstrap and restore with
   `serve` are invalid.
8. Create the database and ownership adapters and establish the process session
   for the validated epoch.
9. Mark `/readyz` successful; a recovery quarantine still rejects every
   executable command at admission.
10. Continue monitoring `Err` for the full process lifetime.

An embedded member may remain alive but unready while its initial peers start;
deployment startup probes must allow enough time for quorum formation. A fatal
embedded-server error makes Atlantis unready and terminates the process after
bounded cleanup.

`/healthz` remains a process-liveness endpoint. `/readyz` reflects etcd
authority, ownership-session health, and embedded-server health.

## Shutdown

Shutdown order is fixed:

1. Mark the replica unready and begin ownership drain.
2. Stop accepting new public and internal command admission.
3. Wait for admitted work and the existing drainer within the shutdown grace
   period; cancel remaining work on expiry.
4. Release or revoke exact ownership claims while the client is available.
5. Stop ownership watches and session renewal.
6. In serve mode, call `EtcdDatabase.Close`, which idempotently delegates to
   `Backend.Close`. In maintenance mode or any partial-startup path where the
   database was never constructed, call `Backend.Close` directly. One shared
   close guard makes these paths mutually idempotent. The backend closes the
   shared client and then, in embedded mode, calls `Etcd.Close` and waits for the
   server to stop.

Ordinary process termination never removes the member from etcd membership. A
member with its retained identity and PVC simply restarts.

## Failure semantics

| Failure | Required behavior |
| --- | --- |
| One voter unavailable in a three-member cluster | Backend authority and unrelated-PR admission continue through two healthy voters; killed-owner work follows the uncertain-outcome policy. |
| Quorum unavailable | All replicas become unready and reject new executable work; no backend fallback. |
| Ownership lease lost | Become unready, reject new steps, cancel queued work, and cancel active subprocesses where possible. |
| Older-generation execution barrier remains | Let a new owner claim, but block that PR until authenticated completion or explicit resolution. |
| Owner dies after HTTP 202 but before execution | Mark its durable admission uncertain after claim loss; never auto-replay it. |
| Forwarding returns stale-claim 409 | Refresh ownership and reroute once. |
| Destructive forwarding response is ambiguous | Report outcome as uncertain; do not automatically replay. |
| Watch compacted or canceled | Perform a linearizable resync and restart the watch. |
| Etcd quota/NOSPACE or persistent write failure | Fail the operation closed and surface an actionable error. |
| Embedded server exits unexpectedly | Become unready, drain/cancel, and terminate the Atlantis process. |
| Restored plan head commit differs | Reject apply and require a new plan. |
| Embedded bootstrap configuration is inconsistent | Fail startup before serving public traffic. |
| Restart mode finds an empty/lost PVC | Refuse startup; require the controlled replacement-member procedure. |
| A nonempty migration or restore target lacks its completion marker | Refuse normal Atlantis startup and executable traffic. |
| Recovery quarantine is active | Allow health and operator inspection, but reject every executable command until audited reconciliation clears it. |

Read and write RPCs use bounded contexts. Transient errors may be retried only
when the operation is proven idempotent or reconciled by reading committed
state.

## Security

Production embedded and external modes require:

- Separate peer and client TLS configurations.
- Peer certificate authentication.
- Atlantis client certificates trusted by the client listener CA.
- Stable advertised DNS names included in certificate SANs.
- HTTPS for every production peer/client/internal endpoint; HTTP is accepted
  only for loopback endpoints with the explicit development flag.
- No `InsecureSkipVerify` path in production configuration.
- A provisioned runtime etcd RBAC identity, selected by the documented
  certificate-CN or username/password-file rule, restricted to the Atlantis
  namespace.
- A separate operator identity for backup, restore, and membership operations;
  those privileges are never granted to the Atlantis runtime identity.
- Constant-time validation of the mandatory internal command token over HTTPS
  using the configured internal CA.
- Network policy separating peer, client, internal command, and public ingress
  traffic.
- An allowlist for replica advertise destinations to prevent SSRF and token
  disclosure.
- File permissions that protect private keys and embedded data directories.
- Hard rejection of unsafe native embedded-etcd settings, including forced new
  cluster, disabled fsync or strict reconfiguration, discovery bootstrap,
  automatic certificates, and lifecycle-inconsistent cluster state.

Atlantis documentation must explicitly prohibit connecting to Kubernetes
control-plane etcd.

## Operations and observability

Member replicas need firm CPU and memory reservations and low-latency durable
storage. Terraform execution must not be allowed to starve embedded etcd of CPU
or disk I/O. Production guidance should prefer spreading voters across distinct
nodes and failure domains while keeping consensus latency within tested bounds.

Required metrics and alerts include:

- Configured mode and embedded etcd version.
- Client endpoint health and linearizable probe latency.
- Cluster leader changes and member health.
- WAL fsync latency, pending/failed proposals, database size, and quota alarms.
- Ownership lease renewal health and lost sessions.
- Claim wins, conflicts, takeovers, and stale-admission failures.
- Forwarding latency, 409 reroutes, 503 failures, and uncertain outcomes.
- Readiness transitions and shutdown-drain time.

Operational procedures include periodic compaction, one-member-at-a-time
defragmentation, snapshots stored outside member PVCs, snapshot validation, and
regular restore drills. Restore is performed offline into new data directories
and creates new etcd cluster/member identities. The offline
`etcdutl snapshot restore` step applies the documented revision bump and marks
the bumped revision compacted while creating every member directory. It emits
per-member pending manifests bound to the snapshot hash, bumped revision, and one fresh
recovery-generation UUID created outside the snapshot. Embedded members first
start in `restore`/`maintenance`; external clusters remain isolated from
Atlantis.

Recovery tooling then verifies the configured deployment ID, restored member
identities, and offline revision-bump receipts, replaces the coordination epoch
with the fresh UUID, deletes restored ownership/session keys, marks every
restored nonterminal command record and execution barrier uncertain, and
atomically writes both the restore-complete marker and an active
recovery-quarantine marker. It installs
the new active identity manifests before embedded members restart in
`restart`/`serve`. Normal startup refuses a restored namespace without the
completion marker.

A snapshot cannot describe locks, plans, or external side effects created after
its revision. Therefore recovery quarantine rejects all executable Atlantis
commands across the deployment until an operator reconciles Terraform state, VCS
heads, PlanStore artifacts, persistent locks, and uncertain executions. Clearing
quarantine is an explicit authenticated, audited compare-and-swap operation; it
is never timer-driven or automatic.

Embedded client/server etcd modules are pinned to the same tested patch release
and upgraded one member at a time. Because the image also contains Atlantis, an
application schema change first rolls the expand-compatible, old-writing binary
across every member and client replica, verifies leased capability records, and
only then enables new writes. Etcd protocol/version upgrade gates remain
one-member-at-a-time operations distinct from the application schema phase. The
embedded version and schema capabilities are exposed in build information.

## Migration and rollback

There is no live mixed BoltDB/etcd fleet and no dual-write mode.

A production cutover follows this sequence:

1. Stop executable ingress and drain Atlantis commands.
2. Back up the BoltDB file or source backend.
3. Export persistent project locks, pull/project statuses, and global locks.
4. For an embedded target, form the new cluster with
   `bootstrap`/`maintenance`; it exposes etcd to operator tooling but cannot
   initialize the namespace or serve Atlantis traffic.
5. Validate that the target etcd namespace is genuinely empty.
6. Create an in-progress migration manifest containing migration ID, deployment
   ID, source identity, expected counts/checksum, and a fresh coordination epoch
   UUID.
7. Import records with create-only transactions associated with that migration
   ID; retries are idempotent and conflicting records abort the migration.
8. Verify exact counts and a deterministic checksum.
9. In one transaction, compare the in-progress manifest, write the schema and
   deployment/epoch records, and mark the migration complete.
10. For embedded mode, activate the recorded identities and restart members in
    `restart`/`serve`; then start the complete Atlantis fleet with
    `locking-db-type=etcd`.
11. Re-plan work that depended on local plan artifacts.

Ownership and session keys are never migrated. An interrupted import leaves the
manifest in progress, and normal startup refuses the namespace until the same
migration is safely resumed or the target namespace is discarded and recreated.
Source data remains read-only until validation is complete.

Rollback is simple only before etcd receives new writes. After new writes,
rollback requires another drain and a reverse export; merely starting the old
binary would reactivate stale lock state.

Moving from embedded to a separately managed external cluster is a cluster
migration using snapshot/restore or controlled member replacement. It is not a
runtime flag flip. Client replicas may already use external mode against the
embedded cluster without moving data.

## Delivery decomposition

This is an umbrella architecture specification. Each workstream below receives
a separate implementation plan and review; no workstream may weaken the locked
cross-workstream safety contracts:

1. **External etcd backend:** common client/keyspace, common database behavior,
   host-scoped project locking, configuration, readiness, and concurrency tests.
2. **Embedded runtime:** explicit bootstrap/restart/join/restore lifecycle,
   maintenance-only startup, identity manifests, unsafe-option rejection,
   voter-count validation, shutdown, dependency pinning, and tests.
3. **Ownership and transport:** leased PR claims, durable admission/idempotency
   records, secure internal forwarding, local claim guard, and readiness.
4. **Execution fencing and plan takeover:** persistent execution barriers,
   admission at every workflow boundary, lease-loss cancellation, uncertain
   outcome resolution, and external PlanStore recovery.
5. **Complete entry-point coverage:** positive-PR API proxying, pull close/reopen
   lifecycle, lock-UI routing, review-triggered work, and explicit drift gating
   or distributed drift work.
   This complete inventory is a production release gate.
6. **Migration and operations:** offline migration manifests, recovery epoch and
   quarantine initialization, backup/recovery, membership and upgrade runbooks,
   and failure drills.
7. **Deployment integration:** Helm support for the fixed member StatefulSet,
   scalable client replicas, identity/ticket material, services, PVCs, TLS
   secrets, probes, PDB, and topology constraints.

External mode is implemented first because it validates key schema and
consistency behavior without embedded lifecycle concerns. Production embedded
support is not declared complete until all seven workstreams and their release
gates pass.

## Verification and acceptance criteria

### Compatibility

- Existing BoltDB tests pass without changes to BoltDB behavior.
- Redis behavior remains unchanged.
- BoltDB remains the default when no new flags are supplied.
- External and embedded modes pass the same etcd database contract suite.
- Mixed N/N+1 fleets remain old-writing until every pinned serving replica
  advertises N+1 readability. Per-operation phase comparisons fence a stale
  binary at activation, and incompatible binaries fail readiness.

### Consistency and concurrency

- Multi-process contention produces exactly one winner for project and global
  locks.
- Two VCS hosts with identical repository names and pull numbers have distinct
  project locks, pull cleanup, statuses, and ownership; neither can delete or
  read the other's scoped records.
- Conditional unlock never removes another owner's lock.
- Parallel pull/project status updates do not lose results.
- Multi-page list and pull-cleanup tests pin reads to one revision; concurrent
  same-pull lock creation is blocked by the lifecycle record and no lock is
  missed.
- Close/reopen races refetch authoritative VCS state, advance lifecycle
  generations with compare-and-swap, and never leave a reopened PR permanently
  unable to acquire a lock.
- At most one authoritative PR owner exists during races and partitions; with
  quorum, a successful claim race eventually commits one owner.
- A stale ownership generation starts no subsequent project workflow step.
- An unresolved older-generation execution barrier blocks new work for that PR
  while unrelated PRs and same-generation parallel projects continue.

### Routing and execution

- Any replica can accept each executable PR-scoped entry point.
- A route-inventory test proves that every executable entry point passes through
  owner resolution; no etcd-mode handler can execute directly on an arbitrary
  replica.
- Atlantis starts work only under the current owner generation. An unresolved
  barrier from an older generation blocks takeover execution for that PR.
- Different PRs distribute across replicas.
- Existing local project/workspace parallelism remains intact.
- Actual HTTP requests assert 202, 409, and 503 behavior rather than testing
  router matching alone.
- Fault injection before and after reservation, local registration, the
  `scheduled` transition, HTTP 202, and runner start proves that each command is
  either scheduled once or durably marked uncertain, never blindly replayed.
- Killing an owner after HTTP 202 but before runner start leaves one durable
  admission record marked uncertain and does not automatically replay it.
- Local plans require re-plan after takeover; external plans restore only after
  clone and head-commit validation.

### Embedded lifecycle

- A new 3-, 5-, or 7-member cluster bootstraps from consistent static
  configuration.
- Unsupported voter counts and inconsistent peer sets fail validation.
- Production bootstrap remains non-serving until all member identities and the
  namespace initialization are atomically activated, followed by restart mode.
- Startup waits for embedded readiness and a linearizable client probe.
- Early errors, timeouts, cancellation, double close, and partial construction
  leak no listeners or goroutines.
- Normal restart with retained PVC and identity is not treated as a membership
  change.
- Bootstrap, restart, join-existing, and restore reject every invalid
  data-directory and identity combination, including an empty/lost PVC in
  restart mode, nonempty storage in bootstrap or join mode, reuse of a removed
  member's storage, and any restore manifest/data mismatch.
- Maintenance startup forms and probes the cluster without initializing the
  Atlantis namespace, claiming PRs, or serving executable ingress.
- Join-existing rejects absent, reused, forged, cluster-mismatched, and
  member-mismatched tickets and starts only a live learner with fresh storage.
- A 3-to-5 transition is tested as two separately authorized
  add-learner/ticket/start/verify/promote sequences. Missing or stale transition
  records, unauthorized ordinals, direct multi-member StatefulSet scaling, and
  automatic template rollouts are rejected.

### Failure and operations

- Killing a follower and killing the leader separately preserve backend
  availability and unrelated-PR admission when quorum remains; work owned by the
  killed Atlantis process follows the defined uncertain-outcome policy.
- A 1/2 network partition in a three-member cluster permits only the majority
  side to admit work.
- Quorum loss makes all replicas unready and fails new commands closed.
- Owner death, lease expiry, forwarding timeouts, watch compaction, slow disk,
  quota exhaustion, and recovery are covered by fault tests.
- Client replicas scale up and down without changing etcd membership.
- PDB and runbook tests for 3, 5, and 7 voters permit at most one voluntary
  member disruption or manual restart at a time.
- Rolling-upgrade tests cover expand, write-version activation, migration,
  rollback within the compatibility window, and later contract across both
  member and client replicas.
- Snapshot restore, migration, and rollback procedures are rehearsed and
  checksum-verified.
- A partial migration or restore without its atomic completion marker blocks
  startup and executable traffic.
- Restore tests apply revision bump and mark-compacted during every offline
  member restore, then install a fresh out-of-snapshot coordination epoch UUID,
  remove restored ownership/session keys, mark nonterminal commands and barriers
  uncertain, and keep every executable command blocked by recovery quarantine
  until audited reconciliation.
- HTTP etcd endpoints, missing or partial TLS identity, invalid authentication
  combinations, non-loopback insecure mode, unsafe advertise destinations,
  missing internal transport credentials, forced-new-cluster, no-fsync, disabled
  strict reconfiguration, discovery bootstrap, automatic certificates, and
  lifecycle-inconsistent cluster state all fail configuration validation.

## References

- [Embedding etcd in a Go application](https://etcd.io/docs/v3.7/dev-guide/golang_embed_pkg/)
- [Embedded server package](https://pkg.go.dev/go.etcd.io/etcd/server/v3/embed)
- [Etcd v3 API](https://etcd.io/docs/v3.7/learning/api/)
- [Etcd API guarantees](https://etcd.io/docs/v3.7/learning/api_guarantees/)
- [Etcd FAQ and quorum guidance](https://etcd.io/docs/v3.7/faq/)
- [Runtime membership reconfiguration](https://etcd.io/docs/v3.7/op-guide/runtime-configuration/)
- [Etcd on Kubernetes](https://etcd.io/docs/v3.7/op-guide/kubernetes/)
- [Etcd transport security](https://etcd.io/docs/v3.7/op-guide/security/)
- [Etcd maintenance](https://etcd.io/docs/v3.7/op-guide/maintenance/)
- [Etcd disaster recovery](https://etcd.io/docs/v3.7/op-guide/recovery/)
