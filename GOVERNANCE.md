# Atlantis Governance

This document defines the project governance for Atlantis.

## Overview

**Atlantis** is committed to building an open, inclusive, productive and self-governing open source
community focused on building a high-quality infrastructure orchestration system. The
community is governed by this document with the goal of defining how community
should work together to achieve this goal.

## Code Repositories

The following code repositories are governed by Atlantis community and
maintained under the `runatlantis` organization.

* **[atlantis](https://github.com/runatlantis/atlantis):** Main Atlantis codebase.
* **[atlantis-helm-charts](https://github.com/runatlantis/helm-charts):** Helm chart for easy deployment of Atlantis.
* **[atlantis-tests](https://github.com/runatlantis/atlantis-tests):** A set of terraform projects that atlantis e2e tests run on.
* **[atlantis-example](https://github.com/runatlantis/atlantis-example):** A simple terraform project to use along with atlantis bootstrap mode.

## Community Roles

* **Users:** Members that engage with the Atlantis community via any medium (Slack, GitHub, mailing lists, etc.).
* **Contributors:** Regular contributions to projects (documentation, code reviews, responding to issues, participation in proposal discussions, contributing code, etc.). 
* **Core Contributors:** Contributors who drive certain subprojects within Atlantis. They are responsible for the direction and work done within that subproject, providing enhancements and support for the Atlantis project as a whole. Core Contributors are expected to contribute code and documentation, review PRs including ensuring quality of code, triage issues, proactively fix bugs, and perform maintenance tasks for the subprojects they are responsible for.
* **Maintainers:** The Atlantis project leaders. They are responsible for the overall health and direction of the project; final reviewers of PRs and responsible for releases. Some Maintainers are responsible for one or more components within a project, acting as technical leads for that component. Maintainers are expected to contribute code and documentation, review PRs including ensuring quality of code, triage issues, proactively fix bugs, and perform maintenance tasks for these components.

### Maintainers

New maintainers and subproject maintainers must be nominated by an existing maintainer and must be elected by a supermajority of existing maintainers. Likewise, maintainers can be removed by a supermajority of the existing maintainers or can resign by notifying one of the maintainers.

### Supermajority

A supermajority is defined as two-thirds of members in the group.
A supermajority of [Maintainers](#maintainers) is required for certain
decisions as outlined above. Voting on decisions can happen on the mailing list, GitHub, Slack, email, or via a voting service, when appropriate. Maintainers can either vote "agree, yes, +1", "disagree, no, -1", or "abstain". A vote passes when supermajority is met. An abstain vote equals not voting at all.

### Decision Making

Ideally, all project decisions are resolved by consensus. If impossible, any
maintainer may call a vote. Unless otherwise specified in this document, any
vote will be decided by a supermajority of maintainers.

Votes by maintainers belonging to the same company
will count as one vote; e.g., 4 maintainers employed by fictional company **Fictiousum** will
only have **one** combined vote. If voting members from a given company do not
agree, the company's vote is determined by a supermajority of voters from that
company. If no supermajority is achieved, the company is considered to have
abstained.

## Proposal Process

One of the most important aspects in any open source community is the concept
of proposals. Large changes to the codebase and/or new features should be
preceded by a proposal as an ADR or GH issue in the main Atlantis repo. This process allows for all
members of the community to weigh in on the concept (including the technical
details), share their comments and ideas, and offer to help. It also ensures
that members are not duplicating work or inadvertently stepping on toes by
making large conflicting changes.

The project roadmap is defined by accepted proposals.

Proposals should cover the high-level objectives, use cases, and technical
recommendations on how to implement. In general, the community member(s)
interested in implementing the proposal should be either deeply engaged in the
proposal process or be an author of the proposal.

The proposal should be documented as a separated markdown file pushed to the root of the 
`docs/adr` folder in the [atlantis](https://github.com/runatlantis/atlantis)
repository via PR. The name of the file should follow the name pattern set by the ADR process `<####-short
meaningful words joined by '-'>.md`, e.g:
`0002-adr-proposal.md`.

Use the [ADR Tools](https://github.com/npryce/adr-tools) and run `adr new <title>`

### Proposal Lifecycle

The proposal PR can be marked with different status labels to represent the
status of the proposal:

* **New**: Proposal is just created.
* **Reviewing**: Proposal is under review and discussion.
* **Accepted**: Proposal is reviewed and accepted (either by consensus or vote).
* **Rejected**: Proposal is reviewed and rejected (either by consensus or vote).

## Lazy Consensus

To maintain velocity in a project as busy as Atlantis, the concept of [Lazy
Consensus](http://en.osswiki.info/concepts/lazy_consensus) is practiced. Ideas
and/or proposals should be shared by maintainers via
GitHub with the appropriate maintainer groups (e.g.,
`@atlantis/all-maintainers`) tagged. Out of respect for other contributors,
major changes should also be accompanied by a ping on Slack or a note on the
Atlantis google mailing list as appropriate. Author(s) of proposal, Pull Requests,
issues, etc.  will give a time period of no less than five (5) working days for
comment and remain cognizant of popular observed world holidays.

Other maintainers may chime in and request additional time for review, but
should remain cognizant of blocking progress and abstain from delaying
progress unless absolutely needed. The expectation is that blocking progress
is accompanied by a guarantee to review and respond to the relevant action(s)
(proposals, PRs, issues, etc.) in short order.

Lazy Consensus is practiced for all projects in the `runatlantis` org, including
the main project repository, community-driven sub-projects, and the community
repo that includes proposals and governing documents.

Lazy consensus does _not_ apply to the process of:

* Removal of maintainers from Atlantis

## Updating Governance

All substantive changes in Governance require a supermajority agreement by all maintainers.
