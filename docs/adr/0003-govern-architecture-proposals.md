# 3. Govern architecture proposals

Date: 2026-10-02

## Status

Proposed

Related issue: TODO before publication.

Draft number is provisional; assign the next number after the latest merged ADR when this proposal lands.

This ADR extends [ADR 0001](0001-record-architecture-decisions.md), which remains Accepted. Its adoption follows the process in force at acceptance. Until then, the [current ADR process](https://github.com/runatlantis/atlantis/blob/main/docs/adr/README.md) applies.

## Context

Atlantis records architecture decisions but does not distinguish approval to publish a proposal from approval to adopt it. Proposed records lack a defined acceptance authority or final review period. Contributors need a public path from discussion to a decision that permits asynchronous participation.

Feature maturity and compatibility are separate policy questions, covered by [ADR 0004](0004-govern-feature-lifecycles.md). This ADR governs architecture proposals only.

## Decision

### Publication and discussion

Every ADR links to a primary public issue before publication. Reuse an existing issue when appropriate. The issue links back to the ADR and records open questions, discussion outcomes, and agreed next steps.

| Step | Record |
| --- | --- |
| Discuss | Describe the problem and alternatives in the primary issue |
| Publish | Review the document for clarity, scope, context, and alternatives; merge it as Proposed |
| Deliberate | Revise the proposal through pull requests and discuss unresolved questions in the issue |
| Decide | Review the exact final text in a separate status-change pull request |
| Record | Merge the status, decision date, rationale, and discussion links; update the primary issue |

Publication does not establish acceptance. Meetings are optional; their outcomes inform public review and are recorded in the issue. Attendance is not required. Material changes discussed in a meeting must be published for asynchronous review.

Contributors may request renewed review in the issue. Maintainers record requested revisions, further review, or deferral with a reason when responding. Review checkpoints may be agreed, but there is no mandatory meeting schedule or response deadline. Inactivity changes neither status nor delivery priority.

### Acceptance rule

Acceptance requires all of the following on the status-change pull request:

- Approval of the final text by two distinct maintainers whose role is Maintainer in [MAINTAINERS.md](https://github.com/runatlantis/atlantis/blob/main/MAINTAINERS.md). A listing as Core Contributor alone does not establish this authority.
- A final comment period of 14 calendar days after the final text is posted. A maintainer announces the revision being reviewed and the start and end times in UTC on the pull request, with a link from the primary issue.
- No outstanding objection from a maintainer. Each objection must be resolved with the objecting maintainer's agreement or explicitly withdrawn by that maintainer. A response or an approving review from someone else does not resolve it.

Material revisions restart the comment period. Editorial corrections may retain it when maintainers record that they do not change the decision. Both required approvals must cover the final text.

Silence after the period means no additional objection; it does not supply either required approval. Reactions, approval of the publication PR, and meeting attendance do not count as acceptance. An outstanding maintainer objection cannot be overruled under this process; introducing an overrule mechanism requires a separate ADR.

The decision record links the approvals and comment-period announcement, identifies resolved or withdrawn objections, and states the rationale. Non-maintainer feedback must receive a recorded disposition; maintainers remain responsible for the decision.

### Dispositions and implementation

ADRs retain the existing statuses: Proposed, Accepted, Rejected, Superseded, and Deprecated. Rejection records the reasons in a reviewed status-change PR. Withdrawal is recorded as Rejected with the reason "withdrawn by proposer"; it does not reject the design's merits.

After a decision, close the primary issue or identify remaining implementation work. Exploratory implementation can inform discussion, but merging an architecture change requires acceptance of the applicable ADR. Acceptance does not promise implementation priority or delivery.

An ADR that adds to an earlier decision links to it and leaves it Accepted. An accepted replacement marks the earlier ADR Superseded and links both records. Existing ADR statuses do not change through adoption of this process.

### Adoption and publication tooling

On acceptance, update the ADR README, template, and contributor guidance. The template requires a Related issue field. The README distinguishes discussion in issues, proposal and decision text in ADRs, and document review in pull requests. Keep the index current. Assign numbers in merge order; open-proposal numbers are provisional. Renumber the ADR and its references before merge if another proposal lands first.

Rendering canonical `docs/adr/` records and their statuses on the website, and adding website checks for that path, are separate future work. They do not condition adoption of this process. Website publication must distinguish Proposed records from accepted decisions and retain one editable source.

### Minimal example

The following illustrates the record format; it does not accept or release the API gate.

```markdown
# N. Require explicit permission for the API

Date: YYYY-MM-DD

## Status

Proposed

Related issue: <primary issue URL>

## Context

Operators need to explicitly opt into the experimental API.

## Decision

Require the server experiment permission `api` before configuring `api-secret`.
Keep the existing authentication and plan/apply safety requirements.

## Consequences

Existing API installations must add `experiments: [api]` before upgrading
to the first gated release. Permission alone does not activate the API.
```

Publishing this record leaves it Proposed. A later status-change PR sets Accepted and records the decision date, rationale, two approving review links, the final revision and 14-day comment period, and the disposition of objections. The implementation and release-specific lifecycle documentation follow separately.

## Consequences

Contributors and maintainers can distinguish publication from acceptance and assess a recorded decision without attending a meeting. Two approvals and a final review period establish a concrete acceptance threshold.

Decisions require explicit maintainer participation and may remain blocked by an unresolved objection. Issue records, review periods, and the ADR index require upkeep. The process sets no response deadline or implementation commitment.

## Alternatives considered

| Alternative | Reason not selected |
| --- | --- |
| Accept through the publication PR | Conflates making a proposal discoverable with adopting it |
| Accept through silence alone | Does not establish the required maintainer approvals |
| Require a meeting for each decision | Limits asynchronous participation and adds scheduling work |
| Permit overriding maintainer objections | Requires a separately agreed authority and decision rule |
| Bundle governance, feature policy, and runtime tooling | Combines independently reviewable decisions |

## References

- [ADR 0001](0001-record-architecture-decisions.md)
- [Current ADR process](https://github.com/runatlantis/atlantis/blob/main/docs/adr/README.md)
- [PR 6923](https://github.com/runatlantis/atlantis/pull/6923)
- [Review requesting the split](https://github.com/runatlantis/atlantis/pull/6923#issuecomment-5942874427)
