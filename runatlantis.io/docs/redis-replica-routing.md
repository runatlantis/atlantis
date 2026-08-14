# Redis Replica Routing

Redis replica routing lets multiple homogeneous Atlantis replicas accept VCS webhooks while preserving pull request affinity. Atlantis assigns each pull request to one live replica. Requests that arrive elsewhere are forwarded to that owner, so every workspace in the pull request executes on the same replica.

Routing is activated when a routing-specific setting is configured. Redis locking by itself does not activate routing. Plan files can remain on the owner or use the existing external PlanStore for recovery after owner loss.

## Architecture

For each `{VCS host, repository, pull request}` key:

1. The first replica to process an actionable webhook atomically creates a Redis ownership lease.
2. If that replica owns the exact process claim, it executes the command locally.
3. Otherwise, the ingress replica forwards a credential-free command envelope to the owner's advertised URL.
4. The owner authenticates the request and checks Redis plus its process-local claim before accepting it.
5. Before resetting local state or scheduling work, the owner atomically verifies and renews the exact serialized claim in Redis.
6. Project work that waited for local locks verifies the claim again before cloning or merging the working directory, and once more after acquiring its Git read lock and before starting workflow steps.
7. The owner also renews all held leases every one-third of the configured TTL.

The ownership key covers the whole pull request, not an individual project or workspace. Different pull requests can be assigned to different replicas. Assignment follows first ingress and is not actively rebalanced.

| State | Location |
| --- | --- |
| Project and global locks | Redis |
| Pull status and shared metadata | Redis |
| PR ownership lease | Redis with TTL |
| Working copy | Owner replica's local `--data-dir` |
| `.tfplan` files | Owner replica's local `--data-dir`, optionally backed by the configured external PlanStore |
| VCS credentials | Rehydrated from the owner replica's local configuration |

## Required Configuration

Every replica must use the same Redis backend, internal token, VCS configuration, repository allowlist, and server-side repo configuration. Each replica needs a directly reachable advertise URL. The replica ID defaults to the process hostname and must be unique among live replicas.

```bash
ATLANTIS_LOCKING_DB_TYPE=redis
ATLANTIS_REDIS_HOST=redis-primary.redis.svc.cluster.local
ATLANTIS_REDIS_PORT=6379
ATLANTIS_REPLICA_ADVERTISE_URL=http://atlantis-0.atlantis-headless.atlantis.svc.cluster.local:4141
ATLANTIS_OWNERSHIP_TTL_SECONDS=30
ATLANTIS_INTERNAL_COMMAND_TOKEN=<shared-secret>
```

Setting `--replica-advertise-url`, `--internal-command-token`, or the optional `--replica-id` override expresses routing intent. If any one is configured, Atlantis requires Redis locking, a Redis endpoint, a valid advertise URL, and a non-empty internal token. Partial routing configuration fails startup.

By default, Atlantis uses the hostname returned by the operating system as the replica ID. Use `--replica-id` only when that hostname is not stable and unique. In Kubernetes, the container hostname is normally the pod name; do not use the Kubernetes worker-node hostname.

The ownership TTL defaults to 30 seconds and must be at least 10 seconds. Use a TTL long enough to tolerate routine scheduling and Redis latency, but short enough for the desired failover time.

## Plan Storage

Ownership routing and plan storage are independent. Both modes keep a pull request on one owner while its lease is live.

### Local Plans

Without `--enable-external-stores`, plan files stay beneath the owner's local `--data-dir`. A new owner clears stale local state and requires a new `plan` before `apply`.

### External Plans

With `--enable-external-stores` and a valid server-side `external_stores.plan_store` configuration, Atlantis saves plans through the external PlanStore. After takeover, the new owner clears its local state, ensures the default checkout and every plan-bearing project workspace exist, and restores plans before discovery. Targeted applies also ensure the selected project's resolved workspace exists before loading its plan. Recovery runs for a new local ownership generation even when a pre-workflow hook already recreated the pull directory. Missing, stale, or unavailable external plans fail the command and require a new plan.

## Failure Behavior

Atlantis fails closed when it cannot resolve or reach the owner:

- If Redis is unavailable, ingress returns HTTP 503 and does not execute locally.
- If forwarding reaches a stale claim, ingress resolves ownership once more and retries once.
- Pull-close forwarding is acknowledged after exact claim admission and scheduling. Cleanup continues on the owner; if it fails, Atlantis logs the error and retains ownership so a redelivered event can retry cleanup.
- If the owner disappears or its process restarts, another process can claim the PR after the lease expires.
- On every new process claim, the new owner waits for commands from an older local claim to finish, then deletes the local working directory for that PR once before executing new work.
- With local plan storage, `apply` on a new owner fails with the existing missing-plan response until a user runs `plan` again.
- With external plan storage, a new owner may restore and apply a plan only when its stored head commit matches the pull request.
- Shared project locks are retained. A lock held by the same PR does not prevent the new owner from re-planning.

Graceful shutdown marks the owner store as draining, stops HTTP traffic, waits for active commands, releases exact claims, and then closes Redis. If HTTP shutdown times out, Atlantis logs the error and still drains active commands, releases claims, and closes Redis; in-progress work is tracked independently of open HTTP connections, so a lingering jobs or SSE stream does not abort the rest of the shutdown sequence.

Internal forwarding is at-least-once. A timeout can leave the ingress replica unsure whether the owner accepted a command, so a provider or manual redelivery can execute it again. Atlantis does not claim exactly-once execution. Ownership or forwarding failures return HTTP 503; monitor failed VCS deliveries and redeliver them when the provider does not retry automatically.

## Redis Requirements

Use a dedicated production Redis deployment or managed service:

- Configure authentication with ACLs and enable TLS.
- Use persistence appropriate for lock and metadata durability.
- Set `maxmemory-policy noeviction`; evicting a live lock or ownership key can allow conflicting work.
- Monitor latency, rejected connections, replication health, failovers, and memory pressure.
- Do not use the Kubernetes control-plane etcd. It is not an application datastore and Atlantis does not integrate with it.

Redis replication and failover are generally asynchronous. A successful ownership or lock write can be lost during failover, and a network partition can briefly expose inconsistent primaries. The implementation uses atomic Redis operations, renewable leases, exact process claim IDs, and fail-closed client behavior, but Redis HA is not a linearizable consensus system. Lease checks fence top-level scheduling and queued project admission. Admission is re-checked at each stage a replica could act on stale ownership: before a mutating command (plan, import, state rm) clones or merges the working directory, and again after the in-process working-directory lock is acquired but immediately before workflow steps run. A replica that loses its lease while queued therefore aborts before touching the working directory, and read-only commands abort before any step executes. What the lease checks cannot stop is Terraform or another workflow step that was already running when its owner lost Redis connectivity. Environments requiring strict partition and execution safety need a consensus-backed ownership design plus end-to-end fencing, such as fencing tokens enforced by the execution target; that design is not implemented by this mode.

## Kubernetes

Use a StatefulSet to provide stable pod hostnames and headless DNS. Send public webhook traffic through a normal ClusterIP Service with `sessionAffinity: None`; any replica can accept it. The advertise URL must use the individual pod's headless DNS name, not the load-balanced Service.

For a Kubernetes deployment, configure:

- Three replicas with stable pod hostnames and per-replica persistent volumes.
- A headless Service for internal forwarding and a normal Service for ingress.
- `/healthz` liveness and lease-aware `/readyz` readiness probes.
- A PodDisruptionBudget, topology spreading, and a 10-minute termination grace period.
- Secret references for Redis, VCS credentials, and the internal forwarding token.
- A starting NetworkPolicy that must be adapted to the cluster's ingress controller labels.

The internal endpoints are served under `/internal/commands/` on the Atlantis HTTP port. Do not publish that path through the public Ingress or Gateway. NetworkPolicy cannot filter HTTP paths, so enforce the path restriction at the L7 proxy. Use service-mesh mTLS or another authenticated transport in addition to the shared token when the pod network is not trusted.

## Scope

Replica routing currently covers actionable VCS comment, pull-open/update autoplan, and pull-close events received through `/events`. Direct `/api` commands, drift operations, job streams, and UI navigation are not owner-routed. Keep those endpoints sticky to one replica or expose them only for operational use until they gain an explicit distributed contract.
