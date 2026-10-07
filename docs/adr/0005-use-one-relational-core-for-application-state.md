# 5. Use durable storage with one data model

Date: 2026-09-23

## Status

Proposed.

## Context

Every feature in Atlantis keeps state its own way. Locks and PR status are JSON in BoltDB or Redis. Plans live in working directories that PR cleanup deletes. Job output, the working-directory lock, and cancellation live in memory.

Nothing shares identifiers, so each new feature has to work out where its data goes, how it links to a repo or PR, and how it survives a restart. Drift detection, the API, and job history all need the same answers.

## Decision

### One relational database

Application state lives in 1 relational database, accessed through GORM behind Atlantis's own interfaces. We support SQLite (embedded, single instance) and PostgreSQL (shared, multi-instance). Other engines need testing first.

Schema changes are versioned SQL migrations run with Goose. Transactions stay short and never wrap a Terraform run or VCS call.

Records link by foreign key, and each has 1 identity:

| Record | Unique by |
| --- | --- |
| VCS instance | Host, port, and base path; also records the provider |
| Repository | VCS instance and full name |
| Pull request | Repository and number |
| Project | Repository, directory, name, workspace |
| Job | Its own ID, linked to PR, project, command, and revision |
| Artifact | Its own ID, linked to its job |

`platform/infrastructure` on `github.com` and on `github.corp.example` are 2 repos. The provider lives on the instance because 1 instance only ever serves 1 VCS. Each plan run is a new job, even at the same commit, and apply uses the specific plan recorded as pending.

### Coordination

If Atlantis needs something to stay correct or recover, it goes in the database or artifact storage. Memory can cache it but can't own it.

Project locks and the working-directory lock both move to the database. Execution claims record their owner, and every change checks that owner atomically so a stale instance can't overwrite a newer one. Recovery assumes Terraform may still be running after a claim expires.

SQLite is single-instance only. We call multi-instance supported once tests cover competing commands, duplicate webhooks, cancellation, PR close, and instance failure.

### Artifacts

Plans, showfiles, and logs go to a separately configured store: a directory, object storage like S3, or the database. The database records each artifact's job, kind, location, size, and checksum.

Working directories add the instance, so same-named repos on different hosts stop sharing a clone: `repos/<instance>/<owner>/<repo>/<pr>/<workspace>`. Artifacts follow the same layout, plus the project directory and job ID:

```text
artifacts/github.com/platform/infrastructure/6920/production/envs/prod/
  9c1f497e-285a-4f13-81ed-6c35570746db/
    networking.tfplan
    networking.json
    networking.log
```

Files take the project name, or the workspace if there's no name. The job ID keeps paths unique. Nothing parses paths, and cleanup goes by database record.

Before a command runs, Atlantis copies the artifacts it needs into the working directory, so `$PLANFILE` and `$SHOWFILE` still work.

Content is written before the database references it, and writes and deletes can be retried. Logs are saved as they stream. Closing a PR invalidates its plans but keeps jobs, logs, and artifacts until retention removes them.

### Deployment

Multi-instance installs need PostgreSQL and artifact storage every instance can reach. We document that; there's no HA flag. The whole thing ships behind the `durable-storage` experiment.

## Consequences

New features get one place to store state and one set of records to link to. Plans and logs survive restarts and PR cleanup, and multi-instance becomes something we can test.

The costs: schema changes need coordination, writes span 2 systems, retained output needs cleanup, and the database can't kill a Terraform process on another host.

## Transition

Existing installs don't change until operators opt in. We recommend a clean cutover: drain, start on an empty store, and regenerate plans before applying. Old logs and plans don't carry over, so keep the old data if you need it. Migration tooling is optional.

Delivery:

1. Locks, status, SQLite, migrations.
2. Jobs, claims, cancellation, filesystem artifacts, durable logs, retention.
3. PostgreSQL, more artifact backends, multi-instance testing.
4. Graduation.

Changing the default or retiring BoltDB and Redis are separate decisions.

## Alternatives considered

- A pluggable store per domain: more combinations, no shared constraints.
- PostgreSQL only: every install needs an external database.
- Everything in SQL: artifacts drive database size.
- Leave it as is: every feature keeps solving storage alone.

## References

- [Pluggable-backends draft (#6829, closed in favor of this ADR)](https://github.com/runatlantis/atlantis/pull/6829)
- [Feature lifecycle proposal (#6923)](https://github.com/runatlantis/atlantis/pull/6923)
- [`working_dir_locker.go`](../../server/events/working_dir_locker.go)
- [Goose migrations](https://github.com/pressly/goose)
- [Issue #6694](https://github.com/runatlantis/atlantis/issues/6694)
