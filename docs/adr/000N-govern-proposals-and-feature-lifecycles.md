# N. Govern proposals and feature lifecycles

Date: 2026-09-24

## Status

Proposed

Related issue: required before publication.

This ADR extends [ADR 0001](0001-record-architecture-decisions.md), which remains
accepted. Until this ADR is accepted, proposals follow the current process in
[`docs/adr/README.md`](README.md).

## Context

ADR 0001 adopted ADRs, and [PR 6820](https://github.com/runatlantis/atlantis/pull/6820)
documented when one is required. The README defines `Accepted` as "approved by
the project" and relies on normal pull request review without defining consensus
or a separate publication review. ADR 0002, dated January 2025 and merged in
June 2026, still has status `Proposed`. These gaps make the state of a decision
difficult to establish.

Features lack a shared lifecycle model too. Server flags control availability
without recording maturity, while `--tf-distribution` deprecation is handled
explicitly in `deprecationWarnings`. Operators need a consistent account of
stability, compatibility, and the notice given before removal.

Contributors need to know how a proposal reaches a decision. Operators need to
know what is experimental or deprecated in the release they run, and what to do
about it.

## Decision

Publish proposals for community discussion, and record ADR acceptance and feature
lifecycle decisions through public review and maintainer consensus. Publish
status and release availability from canonical records shared by the project
process, runtime, and website. Individual feature proposals define behavior,
prerequisites, and migrations.

### States

| Track | States | Applies to |
| --- | --- | --- |
| ADR status | Proposed, accepted, rejected, superseded, deprecated | Designs |
| Maturity | Experimental, stable | Implemented features |
| Compatibility | Supported, deprecated, removed | Features, settings, and experiment names |

The tracks are independent. Accepting an ADR does not release a feature, a
stable feature can have a deprecated setting, and an experiment can be withdrawn
without graduating. Experimental interfaces may change with documented migration
guidance, but experimental status does not exempt a change from the ADR criteria
or relax correctness requirements.

### Proposal and decision process

Anyone can propose an ADR or lifecycle change in a public issue or pull request.
Every ADR links to a primary public issue before publication; reuse an existing
issue where appropriate. The issue links back, stays open while the ADR is
`Proposed`, and records open questions, meeting outcomes, and any agreed next
steps. Asynchronous participation remains available throughout; meeting
attendance is not required to contribute.

| Step | What happens |
| --- | --- |
| Discuss | The issue describes the problem and possible approaches; a community meeting can help but is optional |
| Publish | The ADR merges as `Proposed` after review for clarity, scope, context, and alternatives; this makes it discoverable without accepting it |
| Deliberate | Discussion continues in the issue, and the ADR is revised through pull requests |
| Decide | Maintainers seek consensus through public review; unresolved questions may receive further asynchronous discussion or a follow-up maintainer meeting |
| Record | Maintainers endorse the exact final text in a separate status-change pull request, which records the status, date, and rationale and is linked from the issue |

Meeting outcomes inform the status-change pull request; they do not establish
acceptance on their own. The final text must be available for asynchronous
feedback before acceptance. Material changes require renewed review of the
changed text. Silence, reactions, and approval of the publication pull request
do not establish acceptance; reactions and rankings can help prioritize
discussion.

The decision record distinguishes objections resolved through revisions,
withdrawn by their authors, and still outstanding. Answering an objection does
not itself resolve it. Who may approve, the required participation, and whether
an outstanding objection can be overruled remain open questions below.

Contributors and maintainers are encouraged to revisit proposed ADRs during
community syncs or asynchronous review as availability permits. A review date or
follow-up meeting may be agreed when useful, but is not required. Record outcomes
and any agreed next steps in the issue. Inactivity does not imply acceptance or
rejection. When a proposer withdraws an ADR, a status-change PR marks it
`Rejected` with the reason "withdrawn by proposer" and links the withdrawal from
the issue. This records withdrawal, not a rejection of the design's merits.

Contributors may request renewed maintainer review in the primary issue. When
maintainers respond, they record the next step, such as needs revision, awaiting
further review, or deferred with a reason. These are discussion dispositions;
the ADR remains `Proposed`. No response deadline or delivery commitment is
implied.

After a decision, close the issue or list the remaining implementation work.
Exploratory work can inform discussion, but merging an implementation of the
architecture requires acceptance, and acceptance does not commit anyone to
priority or delivery. Implementation pull requests link the ADR. An ADR that adds
to earlier decisions links to them and leaves them accepted. When a replacement
ADR is accepted, the earlier ADR is marked superseded and links to its successor.

Graduation, default changes, deprecation, and removal use the same consensus
rule. Changes meeting the README criteria must be covered by an accepted ADR;
removing a stable setting is a breaking change. Carrying out a removal already
specified by an accepted ADR does not require another one. No dedicated feature
owner or team is required.

### Registry and validation

A registry in the Atlantis binary records each feature's identifier,
description, maturity, compatibility status, experiment gate, release
milestones, lifecycle decision links, documentation and feedback links, and
validators. Setting entries reference their feature and existing option
definition, adding aliases, replacements, compatibility milestones, and links to
default-change records without duplicating the configuration schema.
The following validation rules apply to registered features and settings.

Features supply prerequisite, conflict, and transition validators that check
actual configuration and capabilities. The framework runs them and reports
failures. It does not enable dependencies, define a dependency language or
plugin system, or require network access.

Operators opt into experiments through a server-only `experiments` list:

```yaml
experiments:
  - example-feature
```

The flag is `--experiments=example-feature` and the environment variable is
`ATLANTIS_EXPERIMENTS=example-feature`. The list is empty by default and fixed at
startup. Normal precedence replaces lists rather than merging them, and
duplicates are normalized. Names are exact, stable, and lowercase, with no
wildcard or `all`. Repository configuration cannot opt into experiments.

Opting in permits experimental functionality to be used; its settings decide
activation and scope. An unused permission has no operational effect. Adding or
removing an experiment entry performs no migration and authorizes no fallback.

Atlantis validates whenever it accepts effective configuration, including
repository configuration read after startup. Invalid configuration blocks the
affected operation before it runs, so components receive only validated
configuration.

| Condition | Result |
| --- | --- |
| Unknown experiment | Reject, listing supported names and documentation |
| Configuration enables or uses experimental functionality without opt-in | Reject, naming the required experiment |
| Failed prerequisite, conflict, or transition check | Reject with the feature's correction or migration guidance |
| Supported deprecated setting | Continue with a warning naming the replacement, if any, and migration guidance |
| Equivalent deprecated alias | Normalize and warn; conflicting alias and canonical values are an error |
| Graduated experiment name within its compatibility period | Accept as a deprecated no-op |
| Removed feature, setting, or experiment name | Reject with migration or cleanup guidance |

Aliases translate only equivalent behavior and cannot bypass experiment opt-in.
Diagnostics name the setting and scope, link to stable documentation URLs, and
omit secret values. Emit server warnings once per
startup. Within a server run, emit each repository warning once per repository
and effective configuration.

### Transitions

Each experiment documents its limitations, supported combinations, graduation
criteria, and behavior on activation, upgrade, disabling, removal, and rollback.
The feature supplies any transition validators and migration procedures.

- **Graduation** requires evidence against documented criteria in a public
  issue and recorded maintainer consensus. A graduation PR removes the gate
  requirement, retains the experiment name as a deprecated no-op for a documented
  period, and updates maturity, documentation, and release notes. The registry
  links to the decision and records the first stable release; until released,
  graduation is shown as unreleased.
- **Default changes** are compatibility decisions, not maturity states. They may
  accompany graduation or occur later, as may migrations and predecessor
  deprecations; graduation implies none of them. Record the old and new defaults,
  effective release, impact on installations that omit the setting, and whether
  and how operators can retain previous behavior. A stable feature may remain
  optional indefinitely.
- **Deprecation** names the interface, reason, replacement if any, operator action,
  first warning release, and migration guidance. Removal can start unscheduled,
  but the earliest removal release must be announced in at least one earlier
  published release. A feature's migration plan can promise a longer window.
  Whether this baseline provides sufficient notice for stable interfaces remains
  an open question below.
- **Removal** is a reviewed code change with tests and release notes, never
  triggered by a date or version check. Shipped names stay in the registry for
  actionable errors and are never reused. Withdrawn experiments follow the same
  notice rules.

Emergency security fixes can shorten notice if they document the exception and
the recovery path. Every lifecycle change, including a change to an experimental
interface, updates registry metadata and documentation in the same pull request,
with release notes and migration guidance ready before the release.

For example, with placeholder releases and separately recorded decisions:

| Release | Change | Operator sees |
| --- | --- | --- |
| N | `example-feature` ships as experimental | Configuration requires explicit experiment opt-in |
| N+2 | Feature graduates; a separate accepted ADR deprecates `--old-flag`. Both obsolete names announce earliest removal in N+4 | Experiment entry warns as a no-op; old flag warns with migration guidance. Defaults remain unchanged |
| N+4 | Reviewed changes remove the old flag and obsolete experiment entry as announced | Both names are rejected with cleanup or migration guidance |

### Publication

The documentation site publishes three linked views:

| View | Navigation | Content |
| --- | --- | --- |
| Architecture decisions | Contributing | Decision text, status, rationale, related issue, and supersession links |
| Experiments | Docs | Availability, opt-in, prerequisites, limitations, graduation criteria and decisions, transition guidance, and feedback links |
| Deprecations and migrations | Docs | Affected interfaces, replacements, release milestones, and migration procedures |

ADR pages render from the Markdown in `docs/adr/`, with an index built from ADR
metadata. Proposed records are clearly labeled, related issues are prominent,
and edit links point to the source; there is no second editable copy. Feature
documentation describes current behavior and leaves ADRs as historical records.

The registry generates deterministic, machine-readable catalogs for the
experiment and deprecation indexes and for configuration-reference labels.
Entries link to Markdown instructions, ADRs, and feedback locations. Each
release keeps its catalog, unreleased changes are labeled, and instructions name
the releases they apply to, without versioning the whole site. Historical and
migration URLs stay reachable, with redirects where needed.

Changes to ADRs, the registry, the generator, or documentation run website CI,
which checks catalog consistency, required metadata and pages, links,
navigation, and the site build. Catalog delivery and rendering tools are
implementation details.

### Adoption

On acceptance, update the ADR README, template, and contributor guidance to
match. The template gains a required `Related issue` field. The README states
where each part lives: discussion in the issue, the proposal and decision in the
ADR, and review of the document in pull requests. Existing ADRs keep their
recorded status.

Adoption is incremental. This ADR does not reclassify existing flags, choose
initial features, or retire anything. A feature registers when its
implementation, metadata, and documentation are ready together.

Framework tests cover permission and name validation, alias conflicts,
diagnostics, graduation compatibility, and propagation of validator failures.
Features own their behavior and compatibility tests, and CI checks that
published catalogs match the registry.

## Consequences

Contributors get a public path from proposal to decision, with explicit
consensus and recorded discussion outcomes. Operators get opt-in experiments,
consistent diagnostics, and documentation for the release they run. The binary
and website share one source of lifecycle facts.

Discussion outcomes need to be recorded, and deferred proposals may need renewed
community attention. Publishing and accepting an ADR take separate reviews. The
registry, catalogs, and CI checks need upkeep, and each feature must define its
transition contract and validators before it registers.

## Alternatives considered

| Alternative | Reason not selected |
| --- | --- |
| Keep proposals in open pull requests until acceptance | Proposals are hard to find, and publication review is mixed with approval |
| Binding votes or automatic acceptance deadlines | Neither shows that technical objections were resolved |
| A meeting for every decision | Makes attendance a barrier and delays proposals that already have public consensus |
| Independent enable flags and prose warnings | Today's approach; lifecycle and compatibility expectations stay inconsistent |
| A global experimental mode or automatic dependencies | Can opt operators into behavior they did not choose |
| All feature settings nested under `experiments` | Forces a configuration move on graduation |
| Remote or dynamically changing feature flags | Adds infrastructure and live-transition semantics |
| Separately maintained website catalogs | Drift from runtime and release behavior |
| ADR acceptance as release status | Conflates design approval with availability |

## Questions for review

The following policy choices must be resolved before this ADR is accepted, or
explicitly delegated through a defined decision process:

1. **Acceptance authority:** Which maintainers are eligible to approve, and what
   minimum participation is required? How are abstentions handled? Must all
   outstanding objections be resolved, or can they be overruled? If overruling
   is allowed, who decides and what rationale must be recorded?
2. **Final review opportunity:** What opportunity for asynchronous feedback is
   required before merging the status-change PR, including after material
   revisions? Should this use a minimum review period or another explicit rule?
   This requirement is separate from optional community meeting schedules.
3. **Deprecation notice:** Should stable interfaces have a project-wide minimum
   measured in both releases and elapsed time, or should each deprecation
   decision specify an earliest removal release and date? What shorter
   commitments, if any, apply to experimental interfaces? The documented
   security exception remains available.

## References

- [ADR 0001: record architecture decisions](0001-record-architecture-decisions.md)
- [ADR process and criteria](README.md)
- [PR 6820: define when an ADR is required](https://github.com/runatlantis/atlantis/pull/6820)
- [ADR 0002: API enhancement and drift detection](0002-api-enhancement-drift-detection.md)
- [PR 6590: ADR 0002 publication](https://github.com/runatlantis/atlantis/pull/6590)
- [Server flags and `deprecationWarnings`](../../cmd/server.go)
- [Server configuration docs](../../runatlantis.io/docs/server-configuration.md)
- [Website configuration](../../runatlantis.io/.vitepress/config.ts)
- [Website CI](../../.github/workflows/website.yml)
