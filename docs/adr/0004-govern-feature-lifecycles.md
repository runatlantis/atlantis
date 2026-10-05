# 4. Govern feature lifecycles

Date: 2026-10-02

## Status

Proposed

Related issue: [#6953](https://github.com/runatlantis/atlantis/issues/6953).

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

Existing Atlantis documentation that calls a feature alpha corresponds to Experimental maturity. This terminology alone introduces no new permission requirement.

The tracks are independent. Accepting an ADR does not release or graduate a feature. A Stable feature can remain optional or have a Deprecated setting. Graduation does not imply a default change, state migration, or predecessor deprecation. Experimental status does not exempt changes from the ADR criteria or relax correctness requirements.

Maintain a canonical Markdown table in the repository, published on the documentation site. Entries record the identifier, description, maturity, compatibility, experiment name, release milestones, instructions, decision links, and feedback location. Deprecated interfaces record replacements and migration guidance. Retain historical records and do not reuse shipped names.

The server contains experiment-name and permission checks plus feature-specific prerequisite checks. Lifecycle metadata remains in documentation. This decision does not introduce a compiled lifecycle registry, general validator framework, alias framework, or generated runtime catalog.

### Experiment permission

Operators opt in through a server-only list:

```yaml
experiments:
  - example-feature
```

Equivalent interfaces are `--experiments=example-feature` and `ATLANTIS_EXPERIMENTS=example-feature`. The list is empty by default and fixed at startup. Multiple experiments can be selected. Parsing, environment-list encoding, precedence, repeated flags, and clearing behavior belong in the implementation proposal and operator documentation.

Names are exact, stable, and lowercase. There is no wildcard or `all`. Repository configuration cannot grant permission. Permission allows experimental functionality to be configured; feature settings still decide activation and scope. An unused permission has no operational effect. Dependencies are not enabled automatically, and adding or removing permission performs no migration or fallback.

Unknown names reject startup with supported names and documentation. Known permission names for graduated features remain recognized, warn operators to remove them, and are ignored; a harmless retired permission must not prevent startup. Withdrawal of an experimental feature requires its own documented treatment of the permission name and recovery instructions. Settings that activate an experiment without permission reject startup, as do unmet prerequisites. Feature-specific checks prevent repository configuration or later requests from bypassing the permission or existing safety requirements. Diagnostics identify the setting and scope without revealing secrets.

### Graduation checklists

Every experiment documents Configuration, Limitations, Graduation criteria, Graduation decision, and Transitions. Instructions identify applicable releases, supported combinations, and behavior on activation, upgrade, disabling, rollback, and removal.

Graduation criteria use Markdown checkboxes and stable identifiers. Each criterion describes a verifiable outcome. Checked items link evidence supporting the whole criterion, such as tests, merged changes, documented contracts, or operational validation. Update progress with the change supplying that evidence. Criteria changes require review; identifiers are not renumbered or reused.

Reviewers check the required sections, criterion identifiers, and whether linked evidence satisfies each completed criterion. Format enforcement tooling is separate implementation work. This policy does not require documentation to be cross-checked against server names or binary metadata.

Completing the checklist makes an experiment eligible for a public graduation decision. If the decision meets the ADR criteria, propose acceptance through that process. Otherwise, use a dedicated graduation PR requiring two eligible maintainer approvals, a 14-day final comment period announced by the proposer, and no unresolved maintainer objection, following the governance ADR once accepted. The announcement identifies the final revision and UTC start and end times, with a link from the feedback issue. Experimental publication and checklist completion start no clock. Record the decision and evidence before removing the permission requirement; until approval, the decision remains Pending.

### Transitions and compatibility

Lifecycle decisions follow the applicable ADR process and need an ADR when they meet its criteria. A removal already specified by an accepted ADR does not require another ADR. Update runtime behavior, documentation, lifecycle records, release notes, and migration instructions together for each transition.

- **Graduation:** remove the permission requirement, record the first Stable release, and retain the experiment name as a Deprecated no-op. Warn operators to remove it; continue recognizing and ignoring it for compatibility. Graduation alone changes no settings, defaults, or stored state.
- **Default changes:** separately record the old and new defaults, effective release, impact on installations omitting the setting, and whether and how previous behavior can be retained.
- **Deprecation:** record the reason, replacement if any, operator action, first warning release, and earliest removal release. Supported deprecated interfaces continue to work with actionable warnings.
- **Removal:** require an explicit reviewed code change, tests, release notes, and cleanup or migration guidance. Dates and version numbers never trigger removal automatically. Removed interfaces receive actionable diagnostics. Retired permissions for graduated features continue to be recognized and ignored.
- **Experimental changes or withdrawal:** document interface changes, cleanup, and recovery. Honor published transition commitments; experimental payloads do not acquire Stable compatibility guarantees merely by shipping.

Stable interfaces receive at least two subsequent stable minor releases and 60 calendar days of notice before removal, whichever ends later. For notice first shipped in N, retain support through N+1; N+2 is the earliest removal, provided 60 days have elapsed since the first warning shipped. Patch releases and prereleases do not shorten this window. For Atlantis's current versioning, a stable minor release advances the minor version without a prerelease suffix; a future major release may serve as a subsequent release boundary. Announce the earliest removal when the notice first ships. Longer promised windows remain binding.

An emergency security fix may shorten notice only with a recorded rationale, the affected interfaces, and operator recovery instructions.

### Website and release availability

Website publication from `main` does not establish release availability. Label forthcoming behavior Unreleased, identify the releases to which instructions apply, and retain usable instructions for current releases. An unreleased requirement must not silently replace the configuration instructions for the latest release.

Assign release milestones when behavior ships and retain historical records in release tags. ADRs remain historical design records; the feature table and instructions describe availability. Preserve historical and migration links. This policy does not require versioning the whole website; rendering, validation scripts, and release-publication automation are separate implementation work.

### Adoption

Document features incrementally. Drift detection is the initial candidate for an experiment permission. Its normal enable setting controls activation; destructive remediation retains its separate setting and safety requirements. The existing API is not gated as a whole by this policy. New or substantially changed API capabilities may warrant their own experiment when their contracts remain unsettled; implementation size alone does not establish instability.

Introducing a mandatory permission for an existing feature is a breaking server-configuration change. An applicable accepted ADR must cover that migration, the first affected release, and upgrade and rollback instructions. This policy establishes the model but does not authorize those feature-specific changes. A documented breaking release can introduce the requirement once that decision and operator guidance exist. Existing feature ADR statuses remain unchanged.

### Minimal example

The following illustrates a table entry and its linked instructions. The checklist is abbreviated to show the format; it is not a complete graduation plan or a release announcement.

```markdown
| Identifier | Description | Maturity | Compatibility | Experiment | First gated release | First stable release | Instructions |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `drift-detection` | Detect infrastructure drift | Experimental | Supported | `drift-detection` | Unreleased | Not scheduled | [Drift detection](experiments.md#drift-detection) |
```

```markdown
## Drift detection

### Configuration

Applies to the unreleased gate change. Permission: `drift-detection`.
Server YAML: `experiments: [drift-detection]`; activation also requires
the drift enable setting and existing API authentication prerequisites.
Feedback: <primary issue URL>. Design: <ADR URL>.

### Limitations

Drift request and response contracts may change. Destructive remediation
requires its separate enable setting and existing safety checks.

### Graduation criteria

- [x] **DRIFT-01** Document drift endpoint requests and responses. Evidence: [API reference](api-endpoints.md).
- [ ] **DRIFT-02** Agree on stable drift request/response contracts.

### Graduation decision

Pending. Checklist completion does not remove the gate.

### Transitions

- **Activate:** configure permission, drift activation, and prerequisites.
- **Upgrade:** configure permission as part of upgrading to a supporting release.
- **Disable:** turn off detection, remediation, and dependent drift webhooks.
- **Rollback:** follow the target release instructions; remove unsupported CLI flags.
- **Removal:** not scheduled.
```

An implementation PR supplies evidence and updates the relevant checkbox. After all criteria are satisfied and maintainers approve graduation, record the decision link and first Stable release, remove the gate requirement, and retain `drift-detection` as a recognized, ignored permission with a warning to remove it. Drift activation still requires its setting; graduation does not change that default.

## Consequences

Operators receive explicit opt-in, visible graduation progress, and release-specific compatibility guidance. Contributors can update documentary records without maintaining server lifecycle metadata. Default changes remain independently reviewable.

Maintainers must review evidence and keep the table, checklists, release milestones, and runtime checks consistent. Document review cannot establish runtime correctness or replace a graduation decision. Existing users may need configuration changes when a feature first receives a gate.

## Alternatives considered

| Alternative | Reason not selected |
| --- | --- |
| Compiled lifecycle registry and validation framework | Adds runtime metadata and upkeep beyond current needs |
| One global experimental switch | Grants permission for experiments operators did not select |
| A separate maturity flag for each feature | Both approaches can work; the shared list provides one consistent declaration of experimental consent, while ordinary feature settings retain activation and scope |
| Alpha off, Beta on, GA locked | Maturity does not establish a safe default for every installation; automatic enablement requires a separate impact and compatibility decision |
| Automatic graduation after checklist completion | Evidence and compatibility decisions require maintainer review |
| Treat website publication as release availability | `main` can describe behavior absent from released binaries |

The shared list adds configuration for features already disabled by default. Its purpose is to record acceptance of unstable behavior separately from operational settings. On graduation, operators retain those settings and no longer need experimental permission. Per-feature enable flags are simpler for a single feature, but combine consent with activation unless another setting is introduced.

Experimental covers features whose contracts are unsettled, including those described as alpha or beta in other projects. A Stable feature may remain optional, and GA does not lock activation or scope settings. Any default change needs its own reviewed compatibility decision.

## References

- [Governance proposal](0003-govern-architecture-proposals.md)
- [ADR 0002: API enhancement and drift detection](0002-api-enhancement-drift-detection.md)
- [PR 6923 and split review](https://github.com/runatlantis/atlantis/pull/6923#issuecomment-5942874427)
- [Current ADR process](https://github.com/runatlantis/atlantis/blob/main/docs/adr/README.md)
