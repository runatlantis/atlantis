# 5. Use durable storage with one data model

Date: 2026-10-22

## Status

Proposed

Related issue: [#7031](https://github.com/runatlantis/atlantis/issues/7031).

## Context

Atlantis keeps state in several places. Locks and PR status are JSON in BoltDB or Redis. Plans and workflow files live in working directories, unless `--enable-external-stores` sends plans to S3. Job output, the working-directory lock, and cancellation live in memory.

Features don't share identifiers, so each one has to solve storage, repo and PR relationships, and restart recovery. Drift detection, the API, and job history all need the same answers. Clone paths also leave out the VCS host, so same-named repos on different hosts share a clone.

## Decision

Atlantis will use one relational database for application records and coordination, plus a configurable store for workflow artifacts. Repositories, pull requests, projects, and jobs share one data model. The database links jobs to their inputs and results. Durable storage replaces BoltDB and Redis.

### Database

Atlantis will support SQLite for single-instance installs and PostgreSQL for shared deployments, accessed through GORM behind Atlantis's own interfaces. SQLite must use a pure-Go driver compatible with `CGO_ENABLED=0`. Goose migrations are the only source of schema changes; Atlantis never calls GORM `AutoMigrate`. Every durable-storage database runs the same migrations, including experimental tables and columns. Experiment settings control whether Atlantis uses those additions.

The model defines the logical identities and relationships below. Identity fields are non-null, with uniqueness enforced by the database. Unnamed projects use an empty name, and related records link by foreign key. Tables, columns, and indexes are implementation details.

| Record | Identity and relationships |
| --- | --- |
| VCS instance | Host, port, and base path |
| Repository | VCS instance and full name |
| Pull request | Repository and number |
| Project | Repository, directory, workspace, and name |
| Job | UUID, linked to its PR, project, command, and revision |

### Coordination

State needed for correctness or recovery lives in the database or artifact storage. Memory can cache it but can't own it. Project locks, the working-directory lock, and execution claims move to the database. Each change to execution state checks the claim's owner atomically, preventing an instance that loses its claim from overwriting newer work. The PR claim covers plan and policy checks and remains held until the required artifacts and completion state are saved.

Recovery must assume Terraform may still be running after a claim expires, including on another host. Database transactions stay short and never wrap Terraform execution or VCS calls.

Multi-instance deployments require PostgreSQL and artifact storage accessible to every instance. Atlantis won't call multi-instance supported until tests cover competing commands, duplicate webhooks, cancellation, PR closure, and instance failure.

### Artifacts

Artifacts include plans, showfiles, job logs, and files passed between workflow stages. They live in a directory, S3-compatible object storage, or the database. In every case, the database records which stored artifacts each job used. The database backend gives small installs one store to back up. Larger installs can keep plans and logs in a directory or object storage to limit database growth.

Every plan run creates a new job, even at the same commit. Policy check and apply jobs link to that plan job. Apply uses the artifacts recorded for the specific pending plan; a successful apply consumes it. Dependent jobs can use artifacts only after they are fully written. Atlantis records checksums when saving artifacts and verifies them before using those artifacts as inputs.

Every backend confines reads, writes, and deletes to the job's assigned repository and PR namespace, including paths derived from PR-controlled `atlantis.yaml`. Working directories and artifact paths include the VCS instance to prevent collisions between repos on different hosts. Pending-plan discovery and cleanup use database records.

### Security and retention

Plans, showfiles, and logs can contain plaintext secrets. Saved plans include sensitive values even when Terraform redacts them from terminal output. Operators configure access controls and encryption at rest for the database and artifact storage. Backend documentation will state the minimum requirements.

By default, plans, showfiles, and workflow files are deleted after their plan is applied or invalidated. PR closure invalidates plans and cancels dependent work; cleanup waits for running commands to stop. Operators may configure a retention period for obsolete files.

Logs stay with their originating job and are saved during execution. They have a bounded default retention period starting at PR closure, or job completion when there is no PR. Operators can shorten or extend that period. Job records survive artifact cleanup, and retention never removes anything a pending plan or running job still needs. Cleanup also removes stored artifacts that no record references, once no job can still be writing them.

## Transition

Durable storage ships behind the `durable-storage` experiment; existing installs keep their configured stores unless operators opt in. The experiment will include the documentation page and graduation checklist required by [ADR 0004](0004-govern-feature-lifecycles.md), with implementation details and open questions tracked in [#7031](https://github.com/runatlantis/atlantis/issues/7031). The existing S3 plan store becomes the artifact store's object storage backend, with new configuration replacing `--enable-external-stores`. This ADR also covers the durable job output proposed in [#6928](https://github.com/runatlantis/atlantis/pull/6928) and replaces that proposal.

BoltDB, Redis, and `--enable-external-stores` are deprecated in the release that introduces the experiment. If the experiment is withdrawn, those deprecations are withdrawn too. After graduation and the required ADR 0004 notice period, reviewed PRs make durable storage the default and remove the old stores and configuration. Those changes may ship in the same release.

Cutover requires draining active work, starting with an empty store, and replanning open PRs. No migration tooling is planned for the old locks, PR status, and plans; replanning is already the documented [recovery procedure](https://www.runatlantis.io/docs/faq.html). Experiment data, including logs and job history, may not survive disabling the experiment or resetting its storage before graduation. Operators who need that data must preserve it separately.

## Consequences

New features share records, identities, and storage rules. Plans survive restarts, logs and job records survive PR cleanup, and multi-instance behavior has explicit correctness requirements.

Schema changes go through shared migrations. Saving an artifact writes to the artifact store and the database without one transaction covering both, so recovery and cleanup have to handle partial writes. Retained output needs protection and cleanup. Database claims can't stop Terraform processes on another host. Multi-instance installs using Redis will need PostgreSQL, and existing installs will receive deprecation warnings while the replacement is experimental.

## Alternatives considered

- A pluggable store per domain: more combinations, no shared constraints.
- PostgreSQL only: every install needs an external database.
- The database as the only artifact store: plans and logs drive database size and backup time on every install.
- Leave it as is: every feature keeps solving storage alone.

## References

- [ADR 0003: Govern architecture proposals](0003-govern-architecture-proposals.md)
- [ADR 0004: Manage feature lifecycles](0004-govern-feature-lifecycles.md)
- [Pluggable-backends draft (#6829, closed in favor of this ADR)](https://github.com/runatlantis/atlantis/pull/6829)
- [Pluggable-backends issue (#6694, closed)](https://github.com/runatlantis/atlantis/issues/6694)
- [Job output persistence (#6928)](https://github.com/runatlantis/atlantis/pull/6928)
- [Plan store object storage (#6312)](https://github.com/runatlantis/atlantis/pull/6312)
- [Goose migrations](https://github.com/pressly/goose)
