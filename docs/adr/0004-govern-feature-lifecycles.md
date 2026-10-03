# 4. Govern feature lifecycles

Date: 2026-10-02

## Status

Proposed

Related issue: TODO before publication.

Draft number is provisional; assign the next number after the latest merged ADR when this proposal lands.

This proposal uses the decision process in [ADR 0003](0003-govern-architecture-proposals.md) once that process is accepted. Until then, the [current ADR process](https://github.com/runatlantis/atlantis/blob/main/docs/adr/README.md) applies. Publication of either proposal does not put its policy into effect.

## Context

Feature settings do not consistently communicate maturity, graduation requirements, or compatibility timelines. Operators need to distinguish upcoming behavior from the behavior of their release. Contributors need evidence-based graduation criteria and explicit decisions about defaults and removal.

A documentation table and simple experiment checks can provide these guarantees without a lifecycle registry compiled into the server or a general validation framework.

## Decision

### States and records

| Track | States | Applies to |
| --- | --- | --- |
| Maturity | Experimental, Stable | Implemented features |
| Compatibility | Supported, Deprecated, Removed | Features, settings, and experiment names |

The tracks are independent. Accepting an ADR does not release or graduate a feature. A Stable feature can remain optional or have a Deprecated setting. Graduation does not imply a default change, state migration, or predecessor deprecation. Experimental status does not exempt changes from the ADR criteria or relax correctness requirements.

Maintain a canonical Markdown table in the repository, published on the documentation site. Entries record the identifier, description, maturity, compatibility, experiment name, release milestones, instructions, decision links, and feedback location. Deprecated interfaces record replacements and migration guidance. Retain historical records and do not reuse shipped names.

The server contains experiment-name and permission checks plus feature-specific prerequisite checks. Lifecycle metadata remains in documentation. This decision does not introduce a compiled lifecycle registry, general validator framework, alias framework, or generated runtime catalog.

### Experiment permission

Operators opt in through a server-only list:

```yaml
experiments:
  - example-feature
```

Equivalent interfaces are `--experiments=example-feature` and `ATLANTIS_EXPERIMENTS=example-feature`. The list is empty by default and fixed at startup. Flags override environment variables, which override server YAML; lists replace across sources. Repeated flags append within the flag list, duplicates are normalized, and `--experiments=` clears it.

Names are exact, stable, and lowercase. There is no wildcard or `all`. Repository configuration cannot grant permission. Permission allows experimental functionality to be configured; feature settings still decide activation and scope. An unused permission has no operational effect. Dependencies are not enabled automatically, and adding or removing permission performs no migration or fallback.

Unknown names reject startup with supported names and documentation. Settings that activate an experiment without permission reject startup, as do unmet prerequisites. Feature-specific checks prevent repository configuration or later requests from bypassing the permission or existing safety requirements. Diagnostics identify the setting and scope without revealing secrets.

### Graduation checklists

Every experiment documents Configuration, Limitations, Graduation criteria, Graduation decision, and Transitions. Instructions identify applicable releases, supported combinations, and behavior on activation, upgrade, disabling, rollback, and removal.

Graduation criteria use Markdown checkboxes and stable identifiers. Each criterion describes a verifiable outcome. Checked items link evidence supporting the whole criterion, such as tests, merged changes, documented contracts, or operational validation. Update progress with the change supplying that evidence. Criteria changes require review; identifiers are not renumbered or reused.

Lightweight documentation checks enforce required sections, checklist syntax, identifiers, and evidence references, and compare documented experiment names with supported server names. Reviewers assess whether evidence satisfies each criterion. The checks neither infer completion nor graduate a feature.

Completing the checklist makes an experiment eligible for a public graduation decision under the applicable ADR process. Record that decision and its evidence before removing the permission requirement. Until approval, the graduation decision remains Pending even if all criteria are checked.

### Transitions and compatibility

Lifecycle decisions follow the applicable ADR process and need an ADR when they meet its criteria. A removal already specified by an accepted ADR does not require another ADR. Update runtime behavior, documentation, lifecycle records, release notes, and migration instructions together for each transition.

- **Graduation:** remove the permission requirement, record the first Stable release, and retain the experiment name as a Deprecated no-op. Warn operators to remove it and announce its earliest removal. Graduation alone changes no settings, defaults, or stored state.
- **Default changes:** separately record the old and new defaults, effective release, impact on installations omitting the setting, and whether and how previous behavior can be retained.
- **Deprecation:** record the reason, replacement if any, operator action, first warning release, and earliest removal release. Supported deprecated interfaces continue to work with actionable warnings.
- **Removal:** require an explicit reviewed code change, tests, release notes, and cleanup or migration guidance. Dates and version numbers never trigger removal automatically. Removed names receive actionable diagnostics.
- **Experimental changes or withdrawal:** document interface changes, cleanup, and recovery. Honor published transition commitments; experimental payloads do not acquire Stable compatibility guarantees merely by shipping.

Stable interfaces and graduated experiment names receive at least two stable minor releases of notice before removal. For notice first shipped in N, retain support through N+1; N+2 is the earliest removal. Patch releases and prereleases do not shorten this window. For Atlantis's current versioning, a stable minor release advances the minor version without a prerelease suffix; a future major release may serve as a subsequent release boundary. Announce the earliest removal when the notice first ships. Longer promised windows remain binding.

An emergency security fix may shorten notice only with a recorded rationale, the affected interfaces, and operator recovery instructions.

### Website and release availability

Website publication from `main` does not establish release availability. Label forthcoming behavior Unreleased, identify the releases to which instructions apply, and retain usable instructions for current releases. An unreleased requirement must not silently replace the configuration instructions for the latest release.

Assign release milestones when behavior ships and retain historical records in release tags. ADRs remain historical design records; the feature table and instructions describe availability. Preserve historical and migration links. This policy does not require versioning the whole website; rendering, validation scripts, and release-publication automation are separate implementation work.

### Adoption

Register features incrementally as their implementation and documentation become ready. Initial candidates are `api` and `drift-detection`. The API permission allows API activation through `api-secret`. Drift detection additionally requires its own permission and enable setting; destructive remediation retains its separate enable setting and safety checks.

Introducing these requirements for existing installations needs explicit upgrade instructions and a first gated release. This proposal does not immediately reclassify existing features, change their defaults, or retire interfaces. Existing feature ADR statuses remain unchanged.

### Minimal example

The following illustrates a table entry and its linked instructions. The checklist is abbreviated to show the format; it is not a complete graduation plan or a release announcement.

```markdown
| Identifier | Description | Maturity | Compatibility | Experiment | First gated release | First stable release | Instructions |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `api` | Authenticated plan, apply, and lock endpoints | Experimental | Supported | `api` | Unreleased | Not scheduled | [API](experiments.md#api) |
```

```markdown
## API

### Configuration

Applies to the unreleased gate change. Permission: `api`.
Server YAML: `experiments: [api]`; activation also requires `api-secret`.
Feedback: <primary issue URL>. Design: <ADR URL>.

### Limitations

Request and response schemas may change. Drift endpoints require
the separate drift experiment and its settings.

### Graduation criteria

- [x] **API-01** Document endpoint requests and responses. Evidence: [API reference](api-endpoints.md).
- [ ] **API-02** Agree on stable request/response contracts.

### Graduation decision

Pending. Checklist completion does not remove the gate.

### Transitions

- **Activate:** configure both permission and secret, then restart.
- **Upgrade:** add permission before installing the first gated release.
- **Disable:** remove the secret and dependent drift settings/webhooks, then restart.
- **Rollback:** remove `experiments` for a release predating experiment support.
- **Removal:** not scheduled.
```

An implementation PR supplies evidence and updates the relevant checkbox. After all criteria are satisfied and maintainers approve graduation, record the decision link and first Stable release, remove the gate requirement, and retain `api` as a Deprecated no-op until its announced removal. API activation still requires its setting; graduation does not change that default.

## Consequences

Operators receive explicit opt-in, visible graduation progress, and release-specific compatibility guidance. Contributors can update documentary records without maintaining server lifecycle metadata. Default changes remain independently reviewable.

Maintainers must review evidence and keep the table, checklists, release milestones, and runtime checks consistent. Format checks cannot establish correctness or replace a graduation decision. Existing users may need configuration changes when a feature first receives a gate.

## Alternatives considered

| Alternative | Reason not selected |
| --- | --- |
| Compiled lifecycle registry and validation framework | Adds runtime metadata and upkeep beyond current needs |
| One global experimental switch | Grants permission for experiments operators did not select |
| A separate maturity flag for each feature | The shared list provides explicit opt-in without multiplying gate flags |
| Alpha off, Beta on, GA locked | Couples maturity to defaults; Stable features may remain optional |
| Automatic graduation after checklist completion | Evidence and compatibility decisions require maintainer review |
| Treat website publication as release availability | `main` can describe behavior absent from released binaries |

## References

- [Governance proposal](0003-govern-architecture-proposals.md)
- [ADR 0002: API enhancement and drift detection](0002-api-enhancement-drift-detection.md)
- [PR 6923 and split review](https://github.com/runatlantis/atlantis/pull/6923#issuecomment-5942874427)
- [Current ADR process](https://github.com/runatlantis/atlantis/blob/main/docs/adr/README.md)
