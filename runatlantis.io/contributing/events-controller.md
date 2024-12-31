# Events Controller

Webhooks are the primary interaction between the Version Control System (VCS)
and Atlantis. Each VCS sends the requests to the `/events` endpoint. The
implementation of this endpoint can be found in the
[events_controller.go](https://github.com/runatlantis/atlantis/blob/main/server/controllers/events/events_controller.go)
file. This file contains the Post function `func (e *VCSEventsController)
Post(w http.ResponseWriter, r *http.Request`)` that parses the request
according to the configured VCS.

Atlantis currently handles one of the following events:

- Comment Event
- Pull Request Event

All the other events are ignored.

```mermaid
---
title: events controller flowchart
---
flowchart LR
    events(/events - Endpoint) --> Comment_Event(Comment - Event)
    events --> Pull_Request_Event(Pull Request - Event)

    Comment_Event --> pre_workflow(pre-workflow - Hook)
    pre_workflow --> plan(plan - command)
    pre_workflow --> apply(apply - command)
    pre_workflow --> approve_policies(approve policies - command)
    pre_workflow --> unlock(unlock - command)
    pre_workflow --> version(version - command)
    pre_workflow --> import(import - command)
    pre_workflow --> state(state - command)

    plan --> post_workflow(post-workflow - Hook)
    apply --> post_workflow
    approve_policies --> post_workflow
    unlock --> post_workflow
    version --> post_workflow
    import --> post_workflow
    state --> post_workflow

    Pull_Request_Event --> Open_Update_PR(Open / Update Pull Request)
    Pull_Request_Event --> Close_PR(Close Pull Request)

    Open_Update_PR --> pre_workflow(pre-workflow - Hook)
    Close_PR --> plan(plan - command)

    pre_workflow --> plan
    plan --> post_workflow(post-workflow - Hook)

    Close_PR --> CleanUpPull(CleanUpPull)
    CleanUpPull --> post_workflow(post-workflow - Hook)
```

## Comment Event

This event is triggered whenever a user enters a comment on the Pull Request,
Merge Request, or whatever it's called for the respective VCS. After parsing the
VCS-specific request, the code calls the `handleCommentEvent` function, which
then passes the processing to the `handleCommentEvent` function in the
[command_runner.go](https://github.com/runatlantis/atlantis/blob/main/server/events/command_runner.go)
file. This function first calls the pre-workflow hooks, then executes one of the
below-listed commands and, at last, the post-workflow hooks.

- [plan_command_runner.go](https://github.com/runatlantis/atlantis/blob/main/server/events/plan_command_runner.go)
- [apply_command_runner.go](https://github.com/runatlantis/atlantis/blob/main/server/events/apply_command_runner.go)
- [approve_policies_command_runner.go](https://github.com/runatlantis/atlantis/blob/main/server/events/approve_policies_command_runner.go)
- [unlock_command_runner.go](https://github.com/runatlantis/atlantis/blob/main/server/events/unlock_command_runner.go)
- [version_command_runner.go](https://github.com/runatlantis/atlantis/blob/main/server/events/version_command_runner.go)
- [import_command_runner.go](https://github.com/runatlantis/atlantis/blob/main/server/events/import_command_runner.go)
- [state_command_runner.go](https://github.com/runatlantis/atlantis/blob/main/server/events/state_command_runner.go)

## Pull Request Event

To handle comment events on Pull Requests, they must be created first. Atlantis
also allows the running of commands for certain Pull Requests events.

<details>
  <summary>Pull Request Webhooks</summary>

The list below links to the supported VCSs and their Pull Request Webhook
documentation.

- [Azure DevOps Pull Request Created](https://learn.microsoft.com/en-us/azure/devops/service-hooks/events?view=azure-devops#pull-request-created)
- [BitBucket Pull Request](https://support.atlassian.com/bitbucket-cloud/docs/event-payloads/#Pull-request-events)
- [GitHub Pull Request](https://docs.github.com/en/webhooks/webhook-events-and-payloads#pull_request)
- [GitLab Merge Request](https://docs.gitlab.com/ee/user/project/integrations/webhook_events.html#merge-request-events)
- [Gitea Webhooks](https://docs.gitea.com/next/usage/webhooks)

</details>

The following list shows the supported events:

- Opened Pull Request
- Updated Pull Request
- Closed Pull Request
- Other Pull Request event

The `RunAutoPlanCommand` function in the
[command_runner.go](https://github.com/runatlantis/atlantis/blob/main/server/events/command_runner.go)
file is called for the _Open_ and _Update_ Pull Request events. When enabled on
the project, this automatically runs the `plan` for the specific repository.

Whenever a Pull Request is closed, the `CleanUpPull` function in the
[instrumented_pull_closed_executor.go](https://github.com/runatlantis/atlantis/blob/main/server/events/instrumented_pull_closed_executor.go)
file is called. This function cleans up all the closed Pull Request files,
locks, and other related information.
