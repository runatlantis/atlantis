
# ADR-0002: Project Locks

- Date: 2023-05-09
- Status: Draft 

## Context and Problem Statement

### Problem 

There is a long-standing regression introduced by a PR to allow parallel plans to happen for projects within the same repository that also belongs in the same workspace. The error prompting users when attempting to plan is:
```
The default workspace at path . is currently locked by another command that is running for this pull request.
Wait until the previous command is complete and try again.
```
### Context

There are multiple locks that occur during the lifecycle of a Pull Request event in Atlantis. The two main locks that pertain to this are:

- Project locks (overall project lock - stored in DB)
- Working Dir locks (file operations - not stored in DB)

#### Project Locks + Atlantis Models
There are four main model classes that pertain to this issue:

- [Repo](https://github.com/runatlantis/atlantis/blob/main/server/events/models/models.go#L40-L62)
- [PullRequest](https://github.com/runatlantis/atlantis/blob/main/server/events/models/models.go#L155-L180)
- [Project](https://github.com/runatlantis/atlantis/blob/main/server/events/models/models.go#L245-L255)
	- Project represents a Terraform project. Since there may be multiple Terraform projects in a single repo we also include Path to the project root relative to the repo root.
- [ProjectLock](https://github.com/runatlantis/atlantis/blob/main/server/events/models/models.go#L225-L240)
	- ProjectLock represents a lock on a project.

Each Repo can have many Pull Requests (One to Many)
Each Pull Request can have many Projects (Many to Many)
Each ProjectLock has one Project (One to One)

#### Working Dir Loccks
[Working Dir locks](https://github.com/runatlantis/atlantis/blob/main/server/events/working_dir_locker.go#L29-L52) are not part of the backend DB and thus do not have a model, instead its in-memory in `working_dir_locker`

Currently can lock the entire PR or per workspace + path
- [TryLockPull](https://github.com/runatlantis/atlantis/blob/f4fa3138d7a9dfdb494105597dce88366effff9e/server/events/working_dir_locker.go#L59-L75): Repo + PR + workspace
- [TryLock](https://github.com/runatlantis/atlantis/blob/f4fa3138d7a9dfdb494105597dce88366effff9e/server/events/working_dir_locker.go#L77-L94): Repo + PR + workspace + path

#### Stack Walk Overview

Here is a high-level view of what happens when a Pull Request is opened with an auto-plan:

1. Events Controller accepts POST webhook and determines which VCS client to handle the request
2. VCSEventsController determines what type of VCS event (opened) and calls RunPlanCommand in CommandRunner
3. CommandRunner validates the context and runs pre-hooks (if they exist)
    1. WorkingDirLocker locks - `w.WorkingDirLocker.TryLock(baseRepo.FullName, pull.Num, DefaultWorkspace, DefaultRepoRelDir)`
    2. Git Repo is cloned - `w.WorkingDir.Clone(log, headRepo, pull, DefaultWorkspace)`
    3. run hooks
    4. WorkingDir lock is released when the pre-hook function returns
4. CommandRunner determines which command runner to use (Plan)
5. PlanCommandRunner determines projects affected by the Pull Request by calling projectCmdBuilder 
    1. VCS client returns modified files
    2. returns here/skips cloning in some cases `--skip-clone-no-changes`
    3. WorkingDirLocker locks - `p.WorkingDirLocker.TryLock(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, workspace, DefaultRepoRelDir)`
    4. Git Repo is cloned - `p.WorkingDir.Clone(ctx.Log, ctx.HeadRepo, ctx.Pull, workspace)`
    5. parses server and repo configs
    6. determines projects and returns them
    7. WorkingDir lock is released
6. PlanCommandRunner cleans up previous plans and Project locks
7. PlanCommandRunner passes ProjectCommandRunner and a list of projects to ProjectCommandPoolExecutor which executes  `ProjectCommandRunner.doPlan`
8. ProjectCommandRunner.doPlan 
	1. acquires Project lock - `p.Locker.TryLock(ctx.Log, ctx.Pull, ctx.User, ctx.Workspace, models.NewProject(ctx.Pull.BaseRepo.FullName, ctx.RepoRelDir), ctx.RepoLocking)`
	2. acquires Working Dir lock - `p.WorkingDirLocker.TryLock(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, ctx.Workspace, ctx.RepoRelDir)`
	3. Git Repo is cloned - `p.WorkingDir.Clone(ctx.Log, ctx.HeadRepo, ctx.Pull, ctx.Workspace)`
	4. Validates Plan Requirements (if any)
	5. runs project plan steps (runtime.<step>) 
	6. return Plan output
9. update Pull Request w/ comment, update backend DB and update commit status
10. fin

Note: This skips over and summarizes some function calls, it's a rough high-level stack walk.

We call Working Dir lock at least `Θ(n) + 2` as well as Project locks  `Θ(n)`. Originally, Atlantis only locked based on BaseRepo name, Pull Request number, and workspace. This was introduced in v0.13.0. 

#### Previous attempts to fix

To allow parallel plans for projects in the same TF workspace, PR [#2131](https://github.com/runatlantis/atlantis/pull/2131) introduced `path` to the `working_dir_locker` and `locks_controller`.  

An additional PR was made PR [#2180](https://github.com/runatlantis/atlantis/pull/2180) because the original PR unblocked `working_dir` that had directory locks at different paths, causing collision because `working_dir` was unaware of different project paths within the same repository. The attempt was to clone even more times at different paths inside each workspace by appending a base32 encoded string of the project path. This was reverted due to another issue [#2239](https://github.com/runatlantis/atlantis/issues/2239).

This is due to a combination of different directories passed to Working Dir when cloning the repository during the pre-workflow hook and during a plan.

 ## Goals & Non-Goals
 
 ### Goals
 - Alignment on what Atlantis should or shouldn't be locking
 - Focus on small scope changes related only to locking
 - Supports all workflow use-cases 
	 - Terraform w/o workspaces
	 - Terraform w/ workspaces
	 - Terragrunt monorepo
 
### Non-Goals
- Making changes/improvements to other packages/sub-systems
  - Moving plan data storage into backend state
- Avoid massive refactoring if possible
- Focus on a singular workflow

## Previous attempts
 There have been a couple of PRs submitted that have either been reverted, or grown stale:

- add path to WorkingDir [#2180](https://github.com/runatlantis/atlantis/pull/2180)
	- base32 encodes the path provided
	- Uses the unique base32 string to additional clone the repo to unique directories to avoid plan file overlap per project + workspace
- add non-default workspaces to workflow hooks [#2882](https://github.com/runatlantis/atlantis/pull/2882)
    - reintroduces #2180 changes to working_dir
    - run hooks on the default workspace
    - also run hooks on every project found
    - significant increase in execution time, especially for Terragrunt users relying on terragrunt-atlantis-config
- reduce the amount of git clone operations [#2921](https://github.com/runatlantis/atlantis/pull/2921)
    - started to try to reduce execution time on #2882 implementation
    - attempts to utilize [TF_DATA_DIR](https://developer.hashicorp.com/terraform/cli/config/environment-variables#tf_data_dir) for workspaces and remove the workspace for the lock. 
    - Clones only at BaseRepo/PR# and moves workspaces from Atlantis clones to TF workspace managed

This is not to suggest the revival of these PRs in their current state but to act as a reference for additional focused solutions. 

## Solution: Clone once + TF_DATA_DIR

Take PR #2921 and re-implement locks on the terraform DATA_DIR only and the entire base repo + PR for git clones. This will focus solely on git operations and Terraform Data Directories. 

### Clone once

We should be avoiding altogether re-cloning the repo unless absolutely necessary. We currently attempt to clone 3 times in a single command execution.

1) Pre-Workflow Hooks
2) <Import/Plan/Apply>CommandRunner
	a) determintes projects to run plans on
3) ProjectCommandRunner
	a) clones before actual plan execution

We should clone one initially for the entire repo/PR. This will empower multiple workspaces/plans alongside TF_DATA_DIR without needing to re-clone.

Here are the cases that will trigger a clone:

	1) Initial clone of repo
 	2) Force Reclone 
  		a) Error when attempting fetch + pull to update existing local git repo
    		b) File operation on a path that doesn't exist (data loss/accidental deletes)

In all other situations, we should be utilizing Git for its delta compressions and performing less intensive fetch + pull operations.

### TF_DATA_DIR

Terraform currently contains its working directory data at `TF_DATA_DIR` which by default is `.terraform`. Utilizing this environmental override, we can store information about the individual project plans and backend state in separate data directories. This would allow for parallel plans not only per project, but across workspaces, in a single PR.

The proposed new structure would pass through `.terraform-$workspace`. Workspace in this case refers to the terraform workspace specified on a project/repo. If one is not provided, it is set to `default`.

### Locking

There has also been a history of issues with colliding locks that disrupts parallel commands running. In this case, locking happens to avoid collison on file operations. We should be locking during Git operations but might not need to during workspace/plan file creation.

*TODO:* dig into locking more thoroughly. 

## Links

- https://github.com/runatlantis/atlantis/issues/1914
- https://github.com/runatlantis/atlantis/pull/2131
- https://github.com/runatlantis/atlantis/pull/2180
- https://github.com/runatlantis/atlantis/pull/2882
- https://github.com/runatlantis/atlantis/pull/2921
- https://developer.hashicorp.com/terraform/cli/config/environment-variables#tf_data_dir
- https://github.com/runatlantis/atlantis/blob/main/server/events/models/models.go
- https://github.com/runatlantis/atlantis/blob/main/server/events/working_dir.go
- https://github.com/runatlantis/atlantis/blob/main/server/events/working_dir_locker.go
