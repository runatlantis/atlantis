# 4. Manage feature lifecycles

Date: 2026-10-02

## Status

Proposed

Related issue: [#6953](https://github.com/runatlantis/atlantis/issues/6953). The number is provisional. Adoption follows [ADR 0003](0003-govern-architecture-proposals.md) if it's accepted, otherwise the [current ADR process](README.md).

## Context

Atlantis has no shared rules for shipping a feature before it's finished, calling it stable, or removing it. Contributors don't know what graduation requires. Operators can't tell what their release supports or how much warning they'll get before something breaks.

## Decision

A feature is experimental, then stable. Later it may be deprecated and removed. It can stay stable forever, and an experiment can be withdrawn without graduating. Features that meet the [ADR criteria](README.md#when-is-an-adr-required) need an accepted ADR before they ship as experiments.

### Experiments

Operators opt in on the server:

```yaml
experiments:
  - example-feature
```

`--experiments` and `ATLANTIS_EXPERIMENTS` work too. The list is empty by default, read at startup, and matched exactly, with no wildcards. Names are lowercase and never reused. Repo config can't opt in.

Opting in means accepting that the feature may change. The feature's own settings still decide whether it runs. Atlantis refuses to start if a name is unknown or an enabled feature is missing its opt-in or prerequisites. Errors name the setting, say how to fix it, and link to docs.

Each experiment has a tracking issue and a docs page in the repo. The page covers configuration, limitations, how to enable, disable, and roll back, and a checklist of requirements to become stable. Each item must be verifiable and link to evidence, usually a test. Checklist changes go through normal PR review.

Experiments still have to be correct and safe. They don't get compatibility guarantees, but changing or withdrawing one needs cleanup instructions. Existing alpha features count as experimental; relabeling them adds no opt-in.

### Graduation

When the checklist is done, a normal PR removes the opt-in requirement. No new ADR. Reviewers check the evidence and the docs.

Graduation doesn't change defaults or migrate state. Atlantis keeps accepting the old experiment name, ignores it, and warns, so old configs still start.

### Deprecation and removal

A deprecation notice says why, what replaces it, how to migrate, the first release with a warning, and the earliest removal release. The feature keeps working, with warnings, until it's removed.

For stable features, settings, and APIs, notice lasts at least 2 stable minor releases and 60 days, whichever is longer. If the warning ships in N, N+1 still supports it, and removal can land in N+2 once 60 days have passed. Patches and prereleases don't count.

Removal is a reviewed PR with tests, release notes, and migration steps. Nothing is removed automatically on a date or version. An emergency security fix can shorten the notice if the PR explains why and how operators recover.

### Defaults

Changing a default is its own reviewed change, separate from graduation. The PR names the release, shows the old and new values, and explains how to keep the old behavior.

### Documentation

The website has a feature table with each feature's name, experiment name, stage, releases, and links to its design, docs, and tracking issue.

The website has a version selector. Each stable release tag publishes its own docs, and the latest is the default. ADRs and experiment pages build from `main`, and each experiment page records the release where a change shipped.

Website and release tooling are separate work. The server only checks experiment names, opt-ins, and prerequisites.

### Adoption

The policy applies one feature at a time, starting with drift detection. Adding an opt-in to an existing feature breaks configs, so that needs its own ADR with migration steps. This ADR changes no existing feature or ADR.

## Consequences

Contributors get a defined path to stable. Operators see what they're opting into and get predictable notice before removals.

Maintainers have to keep the checklists and feature table current, and publish 1 doc build per release. Doc fixes for a released version wait for the next patch release. Operators get one more config step per experiment.

## Alternatives considered

- One switch for all experiments: turns on things the operator didn't choose.
- Enable flags only: one setting would mean both "allow this experiment" and "turn it on."
- Auto-graduate when the checklist is done: skips review.
- Auto-enable stable features: some installs shouldn't run every stable feature.
- Lifecycle records in the server: code to maintain for data the docs can hold.

## References

- [ADR 0002: API enhancement and drift detection](0002-api-enhancement-drift-detection.md)
- [ADR 0003: Govern architecture proposals](0003-govern-architecture-proposals.md)
- [PR #6923 split review](https://github.com/runatlantis/atlantis/pull/6923#issuecomment-5942874427)
