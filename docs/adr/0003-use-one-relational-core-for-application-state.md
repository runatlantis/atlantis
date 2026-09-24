# Use one relational core for application state

Date: 2026-09-23

## Status

Proposed alternative to ADR 0003, "Separate stores from storage backends."
This is a sibling discussion draft, not a replacement for that proposal.

## Context

Atlantis stores project locks, global command locks, and pull/project status in
BoltDB or Redis. These records describe related repositories, pull requests, and
projects, but their relationships are largely encoded in keys, JSON, and application
logic. Persistence behavior is implemented separately for each backend, and changes
to stored formats require bespoke conversion logic. Both backend constructors, for
example, contain logic to detect and rewrite older lock keys.

BoltDB already provides transactions, and Redis can support atomic coordination.
The opportunity is to give Atlantis a common way to compose related changes,
enforce relationships, and evolve its schema across deployment options.

The pluggable-backends proposal keeps each domain's persistence independently
selectable and prohibits cross-store foreign keys and transactions. A single
relational core would allow those domains to share a data model and transaction
boundary while retaining separate application interfaces.

## Decision

Use one authoritative relational database for core application state, accessed
through GORM. SQLite and PostgreSQL are the initial implementations:

- **SQLite:** an embedded database under `--data-dir`, preserving a simple
  installation without an external database service.
- **PostgreSQL:** an external database for operators who prefer managed or
  separately operated storage.

MySQL and other GORM-compatible databases are possible future implementations.
Each supported engine must pass the same behavioral tests for locking,
transactions, durability, and migrations; a GORM driver alone does not establish
support.

Keep domain interfaces such as `LockStore`, `CommandLockStore`, and
`PullStatusStore`. Share database selection, connection lifecycle, schema ownership,
and explicit transaction boundaries across those interfaces. Keep GORM types
inside the persistence layer.

Use foreign keys, uniqueness constraints, and transactions where they enforce
application invariants. Keep transactions short and outside Terraform execution
and VCS requests. Manage schema changes through reviewed, versioned migrations
with a recorded schema version, rather than relying solely on GORM's automatic
schema reconciliation.

The initial implementation replaces persistence for project locks, global command
locks, and pull/project status while preserving existing behavior, including
disabled repository locking. Execution scheduling, durable output, and HA are
future work.

Artifact content remains independently configurable. Existing filesystem/S3 plan
storage and layouts remain supported. Optional SQL artifact content storage is a
follow-up that would reuse the selected core database.

## Benefits

- **Transactions across related state:** lock and status changes can commit or
  roll back together through a shared transaction boundary when a workflow
  requires it.
- **Database-enforced locking invariants:** uniqueness constraints and atomic
  conditional writes can enforce lock acquisition and ownership checks, reducing
  application-level coordination while preserving project-lock semantics.
- **Easier schema evolution:** an ordered migration history provides a consistent
  place to review, test, and apply schema changes and data backfills, replacing
  format detection and conversion embedded in backend startup logic.
- **A richer data model:** foreign keys, indexes, and joins can express relationships
  among repositories, pulls, projects, and locks. Future execution and artifact
  metadata can build on that model.
- **A foundation for first-class HA:** shared relational state and transactions
  provide a basis for coordinating execution across instances. Distributed
  execution and recovery still need their own design.
- **Common persistence across deployment options:** SQLite and PostgreSQL share
  domain logic and GORM-based data access, with a path to additional engines.
  Testing focuses on each engine's behavior instead of combinations of
  independently selected stores.

## Tradeoffs

- Existing BoltDB and Redis installations must migrate, even when their current
  setup works. Maintaining both implementations during transition adds temporary
  complexity.
- A shared schema requires coordinated evolution across domains and deliberate
  transaction and retention boundaries.
- Each supported engine needs validation and may require specific queries,
  concurrency handling, or migrations. GORM adds a dependency and abstraction to
  maintain; it does not remove those differences.
- Versioned migrations still require deliberate data transformation code and
  operational procedures for upgrades and recovery.

## Transition

Introduce SQLite and PostgreSQL as opt-in choices while preserving existing
installations until explicitly migrated. Provide offline migration tooling for
both BoltDB and Redis, validating locks, statuses, and pending-plan associations
before cutover.

Each deployment uses one authoritative core store at a time, without dual writes
or automatic fallback. After the migration path is validated, make SQLite the
default for new installations and retire BoltDB/Redis core storage after a
documented deprecation period.

The implementation design will define release timing, configuration, supported
database versions, and detailed cutover, backup, and rollback procedures.

## Alternatives considered

- **Independently pluggable stores:** preserves per-domain backend choice, but
  expands the supported combinations and prevents shared relational invariants.
- **PostgreSQL only:** reduces engine-specific work, but requires an external
  service for every installation.
- **Keep BoltDB and Redis:** avoids migration and could improve transactions and
  migration tooling in place, but retains separate persistence implementations
  and application-managed relationships.

## References

- [pluggable-backends draft](https://github.com/runatlantis/atlantis/pull/6829)
- [Current BoltDB implementation](../server/core/boltdb/boltdb.go)
- [Current Redis implementation](../server/core/redis/redis.go)
- [GORM database support and drivers](https://gorm.io/docs/connecting_to_the_database.html)
- [GORM migrations](https://gorm.io/docs/migration.html)
- [Issue #6694](https://github.com/runatlantis/atlantis/issues/6694)
- [Pull request #6695](https://github.com/runatlantis/atlantis/pull/6695)
