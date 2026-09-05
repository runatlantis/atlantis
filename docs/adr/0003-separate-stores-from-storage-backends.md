# 3. Separate stores from storage backends

Date: 2026-08-25

## Status

Proposed

## Context

Atlantis uses separate configuration and lifecycle mechanisms for its stores:

- `coordination` uses BoltDB or Redis selected through `--locking-db-type` and `--redis-*` flags.
- `plans` uses filesystem or S3 storage selected through `--enable-external-stores` and `external_stores.plan_store`.
- `jobs` uses process memory.

Feature-specific storage configuration duplicates connection management and leaves backend lifecycle without a single owner.

Existing configuration and physical layouts for coordination and plans are compatibility requirements.

## Decision

Atlantis will separate stores from backends.

### Stores

A store represents the purpose and behavior of Atlantis-owned data.

Atlantis currently has three stores:

- `coordination`: project locks, command locks, and pull status.
- `plans`: Terraform plan artifacts.
- `jobs`: command-output buffers, receiver state, pull-to-job lookup, and websocket delivery state.

Each store owns:

- Its domain operations and interfaces.
- Its consistency, atomicity, durability, retention, and failure requirements.
- Its physical data layout.
- A conformance suite shared by its drivers.

#### Coordination store

Coordination currently exposes project locks, command locks, and pull status through the broad `db.Database` interface. This decision replaces that boundary with three narrow interfaces:

- `LockStore` for project locks.
- `CommandLockStore` for global command locks.
- `PullStatusStore` for project plan and apply status.

The aggregate coordination `Store` implements all three. Consumers depend on the narrowest interface they require. Services such as `Locker` and `ApplyLocker` remain wrappers over these interfaces.

Coordination operations remain atomic relative to competing operations on the same record. Every coordination driver must pass the same conformance suite, including concurrent lock and unlock cases.

A coordination backend failure is fail-closed: Atlantis does not continue a lock-dependent operation.

### Backends

A backend represents configured access to a storage technology.

The current backend types are:

- `boltdb`
- `filesystem`
- `memory`
- `redis`
- `s3`

A backend encapsulates technology-specific configuration and operations for constructing, checking, and closing its client or handle. It does not define the meaning or layout of store data.

Backend definitions and store bindings are global server configuration. Repository-controlled `atlantis.yaml` cannot define or override them, and they are not eligible for `allowed_overrides`.

Configuration is keyed by backend type and permits one configured backend of each type:

```yaml
backends:
  redis:
    host: redis.internal

stores:
  coordination:
    backend: redis
```

Here, `redis` identifies both the backend type and the configured backend.

Unknown fields, missing backend references, and unsupported store/backend pairs fail at startup.

### Drivers and defaults

Drivers are compiled into Atlantis. This is a configurable built-in driver model, not a runtime plugin system.

| Store | Current backends | Default |
| --- | --- | --- |
| `coordination` | BoltDB, Redis | BoltDB under `--data-dir` |
| `plans` | Filesystem, S3 | Filesystem under `--data-dir` |
| `jobs` | Memory | Process memory |

Defaults apply independently to omitted store bindings and do not require backend declarations.

This table records current storage, not permanent compatibility. Each new pair requires defined behavior, a namespace layout, and conformance tests.

### Backend lifecycle

A registry constructs each referenced backend at most once, reuses it across stores, runs its health checks, and closes it exactly once. Store drivers receive constructed clients or handles, not raw backend configuration.

### Data ownership and namespaces

Each driver implements its store's namespace and physical layout for a backend. Stores may share a backend client, but their records and namespaces remain separate. Store identifiers are fixed in code and cannot be configured by repositories or operators.

New store/backend pairs use these namespace encodings from their first release:

| Backend | Encoding |
| --- | --- |
| S3 | `<configured-prefix>/<store-id>/...` |
| Filesystem | `<configured-root>/<store-id>/...` |
| Memory | Store-owned keyspace `<store-id>` |
| Redis | `atlantis:<store-id>:<record>` |
| BoltDB | Top-level bucket `<store-id>` |
| Relational database with schemas | Schema `atlantis_<store-id>` |
| Relational database without schemas | Table prefix `atlantis_<store-id>_` |

> A relational driver may use normalized tables, foreign keys, indexes, and transactions within its store namespace. Cross-store foreign keys and transactions are not permitted because
> they couple store lifecycles and migrations.

A new backend type must define an equivalent namespace encoding before its first store driver is released.

Data-derived identifiers must be encoded so they cannot escape the store namespace.

Redis Cluster requires keys in one multi-key atomic operation to occupy the same hash slot. Such keys share a cluster hash tag:

```text
atlantis:<store-id>:{<atomic-domain>}:<record>
```

The driver uses the smallest atomic domain that preserves correct behavior rather than placing the entire store in one slot.

Existing coordination Redis keys and BoltDB buckets, and plans filesystem paths and S3 object keys, retain their released layouts.

Changing a released layout requires an accepted migration plan covering mixed Atlantis versions, rollback, cleanup, and data present in both layouts. Coordination migrations must also preserve active locks and atomicity; fallback reads across old and new lock keys are insufficient.

### Compatibility and precedence

Legacy locking, Redis, and external plan-store configuration remain supported but are deprecated.

New and legacy configuration may coexist only when they define independent store/backend paths. Atlantis fails at startup when:

- An explicit backend and legacy configuration define or select the same backend type.
- An explicit coordination binding and legacy locking configuration are both present.
- An explicit plans binding and legacy external plan-store configuration are both present.

Conflicts fail even when the values match.

An explicit store binding must reference an explicit backend entry. After rejecting conflicts and validating explicit references, Atlantis translates legacy backend settings and store bindings as complete paths, then applies defaults to unbound stores. Legacy paths never supply values to explicit configuration.

Translations of released Redis and S3 fields must be lossless. Values that cannot be represented fail at startup rather than being ignored.

When legacy configuration is used, Atlantis emits a startup warning that identifies the deprecated configuration, its replacement, and that support may be removed in a future release. Warnings must not include secret values.

User-facing documentation must identify legacy configuration as deprecated and show the equivalent new configuration.

### Secrets

Secrets are backend configuration. Store drivers receive constructed clients or handles, not credentials or secret-source configuration.

Static secret sources must be explicit and mutually exclusive. Atlantis validates environment-variable and file references at startup and never logs secret values or file contents. Errors may identify the source but not its resolved value.

Provider-native credential chains may manage and refresh credentials at runtime.

#### Redis

An explicit Redis backend may set at most one of:

- `password` for a directly configured value.
- `password_env` for the name of an environment variable.
- `password_file` for a file read by Atlantis.

If none is configured, Redis uses no password.

`ATLANTIS_REDIS_PASSWORD` belongs to legacy Redis configuration and conflicts with an explicit Redis backend.

#### S3

S3 uses the AWS SDK credential chain, including environment variables, shared profiles, and workload identity.

## Consequences

### Positive

- Storage configuration and backend lifecycle follow one model.
- Stores can share backend clients while retaining implicit defaults.
- Narrow interfaces and conformance suites preserve store behavior across drivers.
- New stores and backends extend the same model without feature-specific configuration.

### Negative

- Atlantis must support legacy and new configuration during migration.
- Each store/backend pair requires its own driver and conformance tests.
- Only one backend of each type can be configured.
- Released and new drivers may use different namespace conventions.

## Alternatives considered

### Keep feature-specific configuration

Rejected because each additional store would require another configuration and lifecycle mechanism.

### Put backend configuration inside each store

Rejected because stores using the same backend would duplicate configuration and clients.

### Use named backend instances

Deferred because the current stores do not require more than one backend of each type. Named instances can be introduced if Atlantis later needs, for example, separate Redis connections for coordination and jobs.

### Use a shared normalized relational schema

A shared schema could represent repositories, pull requests, locks, plans, and jobs as related tables. It could enforce relationships through foreign keys, support cross-domain queries
and transactions, and provide one migration and operational model. This would make sense if Atlantis required a relational database and its data domains shared transactional
boundaries.

Rejected because the stores have different consistency, durability, access, and scaling requirements, and Atlantis already supports non-relational storage. A shared schema would make
relational storage a system-wide requirement and couple store lifecycles and migrations. A relational driver may still use a normalized schema within its store-owned namespace.

## References

- [Issue #6694](https://github.com/runatlantis/atlantis/issues/6694)
- [Pull request #6695](https://github.com/runatlantis/atlantis/pull/6695)
- [ADR convention](https://github.com/runatlantis/atlantis/blob/main/docs/adr/0001-record-architecture-decisions.md)
