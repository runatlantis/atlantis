# Overview




## Locking
When `plan` is run, the [project](#project) and [workspace](#workspaceenvironment) (**but not the whole repo**) are **Locked** until an `apply` succeeds **and** the pull request/merge request is merged.
This protects against concurrent modifications to the same set of infrastructure and prevents
users from seeing a `plan` that will be invalid if another pull request is merged.

If you have multiple directories inside a single repository, only the directory will be locked. Not the whole repo.

To unlock the project and workspace without completing an `apply` and merging, click the link
at the bottom of the plan comment to discard the plan and delete the lock.
Once a plan is discarded, you'll need to run `plan` again prior to running `apply` when you go back to that pull request.





