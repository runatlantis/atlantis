# 3. Commit Status Handling in the different VCSs

Date: 2024-04-16

## Status

Draft

## Context

Atlantis sets the commit status in the different version control systems (VCSs)
when it runs plans or applies. However, the implementation details of these
statuses differ from system to system. Therefore, Atlantis has an abstraction
layer that defines the status to set and leaves the implementation details to
the specific VCS client.

Another point is that Atlantis users leverage the status of the various
jobs/checks differently, e.g., simply indicating the status of the changes
or triggering automation when a specific status transitions to success.

## Problem

The current implementation of the commit status feature is not working for all
the supported VCSs. Sometimes, it also leads to confusion, as people are not
aware of what the intended behavior should be. One of the most reported problems
is related to targeted plans or applies. When a user runs commands just for a
subset of projects, it can leave a commit status in the state `pending`. For
some VCS, this is obvious. However, this is not how users expect it to behave in
other systems.

Another problem is that no specification currently defines what users can expect
from the status of the VCS they are using. This leads to confusion and makes it
hard to explain the correct behavior.

The following list shows all the open issues related to the problem:

- https://github.com/runatlantis/atlantis/issues/2125
- https://github.com/runatlantis/atlantis/issues/2685
- https://github.com/runatlantis/atlantis/issues/2971
- https://github.com/runatlantis/atlantis/issues/3852
- https://github.com/runatlantis/atlantis/issues/4272
- https://github.com/runatlantis/atlantis/issues/4372

There are also open issues not directly related to the problem at hand. However,
these give a better understanding of what is also needed to solve the problem:

- https://github.com/runatlantis/atlantis/issues/3007
- https://github.com/runatlantis/atlantis/issues/3096
- https://github.com/runatlantis/atlantis/issues/3340
- https://github.com/runatlantis/atlantis/issues/3722
- https://github.com/runatlantis/atlantis/issues/4023
- https://github.com/runatlantis/atlantis/issues/4313

The last list is features suggested by the community to enhance the capabilities
and allow them a better way of integrating Atlantis into their workflow:

- https://github.com/runatlantis/atlantis/issues/4116
- https://github.com/runatlantis/atlantis/issues/4122
- https://github.com/runatlantis/atlantis/issues/4140
- https://github.com/runatlantis/atlantis/issues/4185
- https://github.com/runatlantis/atlantis/issues/4187
- https://github.com/runatlantis/atlantis/issues/4282

## Decision

TBD - The change that we're or have agreed to implement.

## Consequences

TBD - What becomes easier or more difficult to do and any risks introduced by the change that will need to be mitigated.
