# Architecture Decision Records

Architecture Decision Records (ADRs) preserve important design choices and their tradeoffs. Atlantis adopted them in [ADR 0001](0001-record-architecture-decisions.md).

## Index

| ADR | Title | Status |
| --- | --- | --- |
| [0001](0001-record-architecture-decisions.md) | Record architecture decisions | Accepted |
| [0002](0002-api-enhancement-drift-detection.md) | API enhancement and drift detection | Proposed |

Update this table when adding or changing an ADR.

## Which process should I use?

| Change | Process |
| --- | --- |
| Bug fixes, docs, tests, CI, dependency updates, and behavior-preserving refactors | Open a pull request. |
| New user-visible behavior, flags, or commands | Agree on the change in an issue, then open a pull request. |
| A change matching the criteria below | Open an issue, submit an ADR, then implement it after the ADR is accepted. |

The decision matters more than the diff size.

## When is an ADR required?

Write an ADR when a change:

- changes persisted data, migrations, or on-disk layout;
- makes a breaking change to an API, config schema, server flag, default, or webhook contract;
- changes concurrency or locking behavior;
- adds a required runtime service, plugin, or extension point;
- changes authentication, authorization, secret handling, command execution, or webhook validation;
- is irreversible or reverses an accepted ADR.

An ADR is not required for bug fixes, additive optional fields, default-off features, performance work with unchanged behavior, refactors, tests, docs, tooling, CI, or dependency updates.

If unsure, open an issue before starting a large implementation.

## Submit an ADR

1. Agree on the problem in an issue.
2. Copy [template.md](template.md) to `docs/adr/NNNN-short-title.md`.
3. Use the next number across merged ADRs and open ADR pull requests. Renumber if another ADR merges first.
4. Open a pull request with status `Proposed` and update the index above.

Keep the ADR focused on one decision. Put implementation details in the issue or implementation pull request.

## Statuses

- **Proposed:** under discussion.
- **Accepted:** approved by the project.
- **Rejected:** considered and declined.
- **Superseded by NNNN:** replaced by a later ADR.
- **Deprecated:** no longer applies.

Merge accepted and rejected ADRs so both decisions remain discoverable. Do not rewrite them later; supersede them with a new ADR.

ADRs follow normal pull request review rules. Link accepted ADRs from their implementation pull requests.
