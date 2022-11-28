# Intro

Neptune is the project code name for a new user workflow in Atlantis.  This workflow does not take place within a PR like upstream atlantis and instead introduces the concept of a deploy which occurs when the PR is merged.   This new deployment workflow is written entirely on top of a Temporal backend.

Our goal is to eventually have all of Atlantis written as Temporal Workflows.

Note: Both legacy and neptune workflows can be toggled based on a repo's config, however, this document goes into the Neptune workflow and will not talk about the legacy one since it's being phased out.

# Pre-Reading

[Temporal Docs](https://docs.temporal.io/)

# System Architecture

Currently, Atlantis can operate in 3 modes based on the configuration passed in:
* Gateway
* Legacy Worker
* Temporal Worker

In order for Neptune to work correctly, all of these must exist.  

## Gateway
Receives webhook events from github and acts on them accordingly.  Gateway is stateless however each request does spin up a go routine that clones a repository to disk.  This is the primary bottleneck here.

### Events

#### Push
* Clone repository on disk and fetch root configuration.
* Get the latest diff and determine if any roots have changed.
* Start/signal the deploy temporal workflow for each root that requires processing.

#### Pull
* Clone repository on disk and fetch root configuration.
* Get the latest diff of the entire PR and determine if any roots have changed.
* Proxy event to legacy worker if any roots require processing.

#### Check Run
* Determine which action is being triggered.  We have custom buttons that are thrown up on the check run depending on the situation.  As of now there are 2 types of events we are looking for:  `unlocked` and `plan_review`.  When we receive both these events, we signal our deploy workflow and terraform workflow respectively.
* Listens for the GH provided `Re-run failed checkruns` button selection and signals our deploy workflow that we are attempting to add the previously failing revision back into the deploy queue for a rerun attempt. This is similar to the check suite event described below, but the key difference here is this request only reruns attempts where a checkrun has failed. 

#### Check Suite
* Listens for GH provided `Re-run all checkruns` button selection events and signals our deploy workflows that we are attempting add all off the modified roots within a revision back into their respective deploy queues for a rerun attempt, regardless of success or failure status. Note that we only support rerun attempts if a revision meets the following criteria:
  - check run request comes from a revision on the default branch of a repo (force applies are not allowed)
  - revision is identical to the latest persisted deployment attempt for a specific root
## Legacy Worker
Responsible for speculative planning and policy checking within the PR.  This code is relatively untouched from upstream atlantis and should eventually be nuked in favor of Temporal workflows. 

## Temporal Worker
Responsible for running 3 primary processes:
* HTTP Server for viewing realtime logs
* Terraform Task Queue Worker for running Terraform Workflow tasks
* Deploy Task Queue Worker for running Deploy Workflow tasks

# Temporal Workflows

## Deploy

Deploy workflows are run on the granularity of a single repository root.  It follows the ID pattern below:
```
<OWNER/REPO>||<ROOT_NAME>
```

The following is a high level diagram of how this workflow is structured:

```
                                           ┌─────────────┐
                                           │             │
                                           │  deployment │
                                           │    store    │
┌─────────────────┐                        │             │
│     select      │                        └────▲───┬────┘
│                 │                             │   │
│                 │                             │   │
│                 │                             │   │
├┬───────────────┬┤                        ┌────┴───▼──────┐
││revision signal││    ┌──────────────┐    │               │
││   channel     │┼────►priority queue├────► queue worker  │
│┼───────────────┼│    └──────────────┘    │               │
│┼───────────────┼│                        └──────┬────────┘
││ timeout timer ││                               │
│┼───────────────┼│                               │
└─────────────────┘                       ┌───────▼─────────┐
                                          │     select      │
                                          │                 │
                                          │                 │
                                          │                 │
                                          ├┬───────────────┬┤
                                          ││ queue::CanPop │┼───────────────┐
                                          ││               ││               │
                                          │┼───────────────┼│      ┌────────▼─────────┐
                                          │┼───────────────┼│      │      select      │
                                          ││ unlock signal ││      │                  │         ┌─────────┐
                                          │┼───────────────┼│      │                  │         │ Github  │
                                          └─────────────────┘      │                  │         └▲────────┘
                                                                   ├┬────────────────┬┤          │
                                                                   ││ state change   ││          │
                                                                   ││ signal channel │┼──────────┴───┐
                                                                   │┼────────────────┼│              │
                                                                   │┼────────────────┼│              │
                                                                   ││ child workflow ││           ┌──▼──┐
                                                                   │┼────────────────┼│           │ SNS │
                                                                   └──────────────────┘           └─────┘
```

The deploy workflow is responsible for a few things:
* Receiving revisions to deploy from gateway.
* Queueing up successive terraform workflows.
* Updating github check run state based on the in-progress terraform workflow

In order to receive revisions our main workflow thread listens to a dedicated channel.  If we haven't received a new revision in 60 minutes and our queue is empty, we instigate a shutdown of the workflow in its entirety.

### Queue

The queue is modeled as a priority queue where manually triggered deployments always happen before merge triggered deployments.  This queue can be in a locked state if the last deployment that happened was triggered manually.  The queue lock applies only to deployments that have been triggered via merging and can be unlocked through the corresponding check run of an item that is blocked.  

Items can only be popped of the queue if the queue is unlocked OR if the queue is locked but contains a manually triggered deployment.  

By default, a new workflow starts up in an unlocked state.

### Queue Worker
Upon workflow startup, we start a queue worker go routine which is responsible for popping off items from our queue when it can, and listening for unlock signals from gateway.

The worker also maintains local state on the latest deployment that has happened. This is used for validating that new revisions intended for deploy are ahead of the latest deployed revisions.  Once each deployment is complete, the worker persists this information in the configured blob store.  The worker only fetches from this blob store on workflow startup and maintains the information locally for lifetime of its execution.

A deploy consists of executing a terraform workflow.  The worker blocks on execution of this "child" workflow and listens for state changes via a dedicated signal channel.  Once the child is complete, we stop listening to this signal channel and move on.

### State Signal

State changes are reflected in the github checks UI (ie. plan in progress, plan failed, apply in progress etc.).  A single check run is used to represent the deployment state.  The check run state is indicative of the completion state of the deployment and the details of the deployment itself are rendered in the check run details section.

State changes for apply jobs specifically are sent to SNS for internal auditing purposes.

## Terraform

The Terraform workflow runs on the granularity of a single deployment.  It's identifier is the deployment's identifier which is randomly generated in the Deploy Workflow.  Note: this means a single revision can be tied to multiple deployments.

The terraform workflow is stateful due to the fact that it keeps data on disk and references it throughout the workflow.  Cleanup of that data only happens when that workflow is complete.  In order to ensure this statefulness, the terraform workflow is aware of the worker it's running on and fetches this information as part of the first activity.  Each successive activity task takes place on the same task queue.  

Following this:
* The workflow clones a repository and stores it on disk
* Generates a Job ID for the first set of Terraform Operations
* Runs Terraform Init followed by Terraform Plan

Before and after this job, the workflow signals it's parent execution with a state object.  At this point, the workflow either blocks on a dedicated plan review channel, or proceeds to the apply under some criteria. Atm this is only if there are no changes in the plan.

Plan review signals are received directly by this workflow from gateway which pulls the workflow ID from the check run's `External ID` field.  

If the plan is approved, the workflow proceeds with the apply, all the while updating the parent execution with the status, before exiting the workflow.
If the plan is rejected or times out (1 week timeout on plan reviews), the parent is notified and the workflow exits.

### Retries

The workflow itself has no retries configured.  All activities use the default retry policy except for Terraform Activities.  Terraform Activities throw up a `TerraformClientError` if there is an error from the binary itself.  This error is configured to be non-retryable since most of the time this is a user error.  

For Terraform Applies, timeouts are not retried. Timeouts can happen from exceeding the ScheduleToClose threshold or from lack of heartbeat for over a minute.  Instead of retrying the apply, which can have unpredictable results, we signal our parent that there has been a timeout and this is surfaced to the user.

### Heartbeats

Since Terraform activities can run long, we send hearbeats at 5 second intervals. If 1 minute goes by without receiving a hearbeat, temporal will assume the worker node is down and the configured retry policy will be run.

### Job Logs

Terraform operation logs are streamed to the local server process using go channels.  Once the operation is complete, the channel is closed and the receiving process persists the logs to the configured job store.  


