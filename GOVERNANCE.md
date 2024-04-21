# Atlantis Governance

**Atlantis** is committed to building an open, inclusive, productive, and self-governing community focused on building a high-quality infrastructure orchestration system. The
community is governed by this document with the goal of defining how the community should work together to achieve this goal.

Atlantis follows a two-tier governance model. The higher tier is made up of the Atlantis Steering Committee, which is responsible for the project's overall health. Maintainers and Members make up the lower tier, and they are the main contributors to one or more repositories within the overall project.

The governance policies defined here apply to all repositories in the runatlantis GitHub organization.

## Project Steering Committee

The Atlantis project has a project steering committee consisting of 5 members, with a maximum of 1 member from any single organization. The steering committee in Atlantis has a final say in any decision concerning the Atlantis project, with the exceptions of deciding steering committee membership, and changes to project governance. See Changes in Project Steering Committee Membership and Changes in Project Governance.

The initial steering committee will be nominated by the current maintainers of the project in order to ensure continuity of the project charter and vision as additional members come on board. Once the terms of the initial committee members expire, new members will be selected through the election process outlined below.

A list of the steering committee will be published here once decided and voted on.

Any decision made must not conflict with CNCF policy.

The maximum term length of each steering committee member is two year, with no term limit restriction.

Voting for steering committee members is open to all current steering committee members and maintainers.

## Repository Governance 

The Atlantis project consists of multiple repositories that are published and maintained on GitHub. Each repository will be subject to the same overall governance model, but will be allowed to have different teams of people (“maintainers”) with permissions and access to the repository. This increases diversity of maintainers in the Atlantis organization, and also increases the velocity of code changes.

### Maintainer

Each repository in the Atlantis organization are allowed their own unique set of maintainers. Maintainers have the most experience with the given repo and are expected to have the knowledge and insight to lead its growth and improvement.

New maintainers and subproject maintainers must be nominated by an existing maintainer and must be elected by a supermajority of existing maintainers. Likewise, maintainers can be removed by a supermajority of the existing maintainers or can resign by notifying one of the maintainers.

If a Maintainer feels she/he can not fulfill the "Expectations from Maintainers", they are free to step down.

In general, adding and removing maintainers for a given repo is the responsibility of the existing maintainer team for that repo and therefore does not require approval from the steering committee. However, in rare cases, the steering committee can veto the addition of a new maintainer by following the conflict resolution process.

Responsibilities include:

- Strong commitment to the project
- Participate in design and technical discussions
- Participate in the conflict resolution and voting process at the repository scope when necessary
- Seek review and obtain approval from the steering committee when making a change to central architecture that will have broad impact across multiple repositories
- Contribute non-trivial pull requests
- Perform code reviews on other's pull requests
- Ensure that proposed changes to your repository adhere to the established standards, best practices, and guidelines, and that the overall quality and integrity of the code base is upheld.
- Add and remove maintainers to the repository as described below
- Approve and merge pull requests into the code base
- Regularly triage GitHub issues. The areas of specialization possibly listed in OWNERS.md can be used to help with routing an issue/question to the right person.
- Make sure that ongoing PRs are moving forward at the right pace or closing them
- Monitor Atlantis Slack (delayed response is perfectly acceptable), particularly for the area of your repository
- Regularly attend the recurring community meetings
- Periodically attend the recurring steering committee meetings to provide input
- In general, continue to be willing to spend at least 25% of their time working on Atlantis (~1.25 business days per week)

The current list of maintainers for each repository is published and updated in each repo’s OWNERS.md file.

#### Removing a maintainer

If a maintainer is no longer interested or cannot perform the maintainer duties listed above, they should volunteer to be moved to emeritus status. In extreme cases this can also occur by a vote of the maintainers per the voting process below.

## Conflict resolution and voting
In general, it is preferred that technical issues and maintainer membership are amicably worked out between the persons involved. If a dispute cannot be decided independently, the leadership at the appropriate scope can be called in to decide an issue. If that group cannot decide an issue themselves, the issue will be resolved by voting.

### Issue Voting Scopes
Issues can be resolved or voted on at different scopes:

* **Repository**: When an issue or conflict only affects a single repository, then the maintainer team for that repository should resolve or vote on the issue. This includes technical decisions as well as maintainer team membership.
* **Organization**: If an issue or conflict affects multiple repositories or the Crossplane organizations and community at large, the steering committee should resolve or vote on the issue.

### Issue Voting Process

The issue voting process is usually a simple majority in which each entity within the voting scope gets a single vote. The following decisions require a super majority (at least 2/3 of votes), all other decisions and changes require only a simple majority:

* Updates to governance by the steering committee
* Additions and removals of maintainers by the repository’s current maintainer team
* Vetoes of maintainer additions by the steering committee

For organization scope voting, repository maintainers do not have a vote in this process, although steering committee members should consider their input.

For formal votes, a specific statement of what is being voted on should be added to the relevant GitHub issue or PR. Voting entities should indicate their yes/no vote on that issue or PR.

After a suitable period of time (goal is by 5 business days), the votes will be tallied and the outcome noted. If any voting entities are unreachable during the voting period, postponing the completion of the voting process should be considered.

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

## Updating Governance

This governance will likely be a living document and its policies will therefore need to be updated over time as the community grows. The steering committee has full ownership of this governance and only the committee may make updates to it. Changes can be made at any time, but a super majority (at least 2/3 of votes) is required to approve any updates.

## Credits

Sections of these documents have been borrowed from (Argoproj)[https://github.com/argoproj/argoproj/blob/main/community/GOVERNANCE.md] and (Crossplane)[https://github.com/crossplane/crossplane/blob/master/GOVERNANCE.md] projects.
