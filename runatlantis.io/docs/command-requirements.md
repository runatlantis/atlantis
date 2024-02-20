# Command Requirements
[[toc]]

## Intro
Atlantis requires certain conditions be satisfied **before** `atlantis apply` and `atlantis import`
commands can be run:

* [Approved](#approved) – requires pull requests to be approved by at least one user other than the author
* [Mergeable](#mergeable) – requires pull requests to be able to be merged
* [UnDiverged](#undiverged) - requires pull requests to be ahead of the base branch

## What Happens If The Requirement Is Not Met?
If the requirement is not met, users will see an error if they try to run `atlantis apply`:
![Mergeable Apply Requirement](./images/apply-requirement.png)

## Supported Requirements
### Approved
The `approved` requirement will prevent applies unless the pull request is approved
by at least one person other than the author.

#### Usage
The `approved` requirement by:
1. Creating a `repos.yaml` file with the `apply_requirements` key:
   ```yaml
   repos:
   - id: /.*/
     apply_requirements: [approved]
   ```
1. Or by allowing an `atlantis.yaml` file to specify the `apply_requirements` key in the `repos.yaml` config:
    #### repos.yaml
    ```yaml
    repos:
    - id: /.*/
      allowed_overrides: [apply_requirements]
    ```

    #### atlantis.yaml
    ```yaml
    version: 3
    projects:
    - dir: .
      apply_requirements: [approved]
    ```

#### Meaning
Each VCS provider has different rules around who can approve:
* **GitHub** – **Any user with read permissions** to the repo can approve a pull request
* **GitLab** – The user who can approve can be set in the [repo settings](https://docs.gitlab.com/ee/user/project/merge_requests/approvals/)
* **Bitbucket Cloud (bitbucket.org)** – A user can approve their own pull request but
  Atlantis does not count that as an approval and requires an approval from at least one user that
  is not the author of the pull request
* **Azure DevOps** – **All builtin groups include the "Contribute to pull requests"** permission and can approve a pull request

:::tip Tip
To require **certain people** to approve the pull request, look at the
[mergeable](#mergeable) requirement.
:::

### Mergeable
The `mergeable` requirement will prevent applies unless a pull request is able to be merged.

#### Usage
Set the `mergeable` requirement by:
1. Creating a `repos.yaml` file with the `apply_requirements` key:
   ```yaml
   repos:
   - id: /.*/
     apply_requirements: [mergeable]
   ```

1. Or by allowing an `atlantis.yaml` file to specify `plan_requirements`, `apply_requirements` and `import_requirements` keys in the `repos.yaml` config:
    #### repos.yaml
    ```yaml
    repos:
    - id: /.*/
      allowed_overrides: [plan_requirements, apply_requirements, import_requirements]
    ```

    #### atlantis.yaml
    ```yaml
    version: 3
    projects:
    - dir: .
      plan_requirements: [mergeable]
      apply_requirements: [mergeable]
      import_requirements: [mergeable]
    ```

#### Meaning
Each VCS provider has a different concept of "mergeability":

::: warning
Some VCS providers have a feature for branch protection to control "mergeability". To use it,
limit the base branch so to not bypass the branch protection.
See also the `branch` keyword in [Server Side Repo Config](server-side-repo-config.html#reference) for more details.
:::

#### GitHub
In GitHub, if you're not using [Protected Branches](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/defining-the-mergeability-of-pull-requests/about-protected-branches) then
all pull requests are mergeable unless there is a conflict.

If you set up Protected Branches then you can enforce:
* Requiring certain status checks to be passing
* Requiring certain people to have reviewed and approved the pull request
* Requiring `CODEOWNERS` to have reviewed and approved the pull request
* Requiring that the branch is up to date with `main`

See [https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/defining-the-mergeability-of-pull-requests/about-protected-branches](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/defining-the-mergeability-of-pull-requests/about-protected-branches)
for more details.

::: warning
If you have the **Restrict who can push to this branch** requirement, then
the Atlantis user needs to be part of that list in order for it to consider
a pull request mergeable.
:::

::: warning
If you set `atlantis/apply` to the mergeable requirement, use the `--gh-allow-mergeable-bypass-apply` flag or set the `ATLANTIS_GH_ALLOW_MERGEABLE_BYPASS_APPLY=true` environment variable. This flag and environment variable allow the mergeable check before executing `atlantis apply` to skip checking the status of `atlantis/apply`.
:::

#### GitLab
For GitLab, a merge request will be merged if there are no conflicts, no unresolved discussions if it is a project requirement and if all necessary approvers have approved the pull request.

For pipelines, if the project requires that pipelines must succeed, all builds except the apply command status will be checked.

For Jobs with allow_failure setting set to true, will be ignored. If the pipeline has been skipped and the project allows merging, it will be marked as mergeable.

#### Bitbucket.org (Bitbucket Cloud) and Bitbucket Server (Stash)
For Bitbucket, we just check if there is a conflict that is preventing a
merge. We don't check anything else because Bitbucket's API doesn't support it.

If you need a specific check, please
[open an issue](https://github.com/runatlantis/atlantis/issues/new).

#### Azure DevOps
In Azure DevOps, all pull requests are mergeable unless there is a conflict. You can set a pull request to "Complete" right away, or set "Auto-Complete", which will merge after all branch policies are met. See [Review code with pull requests](https://docs.microsoft.com/en-us/azure/devops/repos/git/pull-requests?view=azure-devops).

[Branch policies](https://docs.microsoft.com/en-us/azure/devops/repos/git/branch-policies?view=azure-devops) can:
* Require a minimum number of reviewers
* Allow users to approve their own changes
* Allow completion even if some reviewers vote "Waiting" or "Reject"
* Reset code reviewer votes when there are new changes
* Require a specific merge strategy (squash, rebase, etc.)

::: warning
At this time, the Azure DevOps client only supports merging using the default 'no fast-forward' strategy. Make sure your branch policies permit this type of merge.
:::

### UnDiverged
Prevent applies if there are any changes on the base branch since the most recent plan.
Applies to `merge` checkout strategy only which you need to set via `--checkout-strategy` flag.

#### Usage
You can set the `undiverged` requirement by:
1. Creating a `repos.yaml` file with `plan_requirements`, `apply_requirements` and `import_requirements` keys:
   ```yaml
   repos:
   - id: /.*/
     plan_requirements: [undiverged]
     apply_requirements: [undiverged]
     import_requirements: [undiverged]
   ```
1. Or by allowing an `atlantis.yaml` file to specify the `plan_requirements`, `apply_requirements` and `import_requirements` keys in your `repos.yaml` config:
   #### repos.yaml
    ```yaml
    repos:
    - id: /.*/
      allowed_overrides: [plan_requirements, apply_requirements, import_requirements]
    ```

   #### atlantis.yaml
    ```yaml
    version: 3
    projects:
    - dir: .
      plan_requirements: [undiverged]
      apply_requirements: [undiverged]
      import_requirements: [undiverged]
    ```
#### Meaning
The `merge` checkout strategy creates a temporary merge commit and runs the `plan` on the Atlantis local version of the PR
source and destination branch. The local destination branch can become out of date since changes to the destination branch are not fetched
if there are no changes to the source branch. `undiverged` enforces that Atlantis local version of main is up to date
with remote so that the state of the source during the `apply` is identical to that if you were to merge the PR at that
time.

## Setting Command Requirements
As mentioned above, you can set command requirements via flags, in `repos.yaml`, or in `atlantis.yaml` if `repos.yaml`
allows the override.

### Flags Override
Flags **override** any `repos.yaml` or `atlantis.yaml` settings so they are equivalent to always
having that apply requirement set.

### Project-Specific Settings
If you only want some projects/repos to have apply requirements, then you must
1. Specifying which repos have which requirements via the `repos.yaml` file.
   ```yaml
   repos:
   - id: /.*/
     plan_requirements: [approved]
     apply_requirements: [approved]
     import_requirements: [approved]
   # Regex that defaults all repos to requiring approval
   - id: /github.com/runatlantis/.*/
     # Regex to match any repo under the atlantis namespace, and not require approval
     # except for repos that might match later in the chain
     plan_requirements: []
     apply_requirements: []
     import_requirements: []
   - id: github.com/runatlantis/atlantis
     plan_requirements: [approved]
     apply_requirements: [approved]
     import_requirements: [approved]
     # Exact string match of the github.com/runatlantis/atlantis repo
     # that sets apply_requirements to approved
   ```

1. Specify which projects have which requirements via an `atlantis.yaml` file, and allowing
   `plan_requirements`, `apply_requirements` and `import_requirements` to be set in `atlantis.yaml` by the server side `repos.yaml`
   config.

   For example if I have two directories, `staging` and `production`, I might use:
   #### repos.yaml
   ```yaml
   repos:
   - id: /.*/
     allowed_overrides: [plan_requirements, apply_requirements, import_requirements]
     # Allow any repo to specify apply_requirements in atlantis.yaml
   ```

   #### atlantis.yaml
   ```yaml
   version: 3
   projects:
   - dir: staging
     # By default, plan_requirements, apply_requirements and import_requirements are empty so this
     # isn't strictly necessary.
     plan_requirements: []
     apply_requirements: []
     import_requirements: []
   - dir: production
     # This requirement will only apply to the
     # production directory.
     plan_requirements: [mergeable]
     apply_requirements: [mergeable]
     import_requirements: [mergeable]
   ```

### Multiple Requirements
You can set any or all of `approved`, `mergeable`, and `undiverged` requirements.

## Who Can Apply?
Once the apply requirement is satisfied, **anyone** that can comment on the pull
request can run the actual `atlantis apply` command.

## Next Steps
* For more information on GitHub pull request reviews and approvals see: [https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/reviewing-changes-in-pull-requests/about-pull-request-reviews](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/reviewing-changes-in-pull-requests/about-pull-request-reviews)
* For more information on GitLab merge request reviews and approvals (only supported on GitLab Enterprise) see: [https://docs.gitlab.com/ee/user/project/merge_requests/approvals/](https://docs.gitlab.com/ee/user/project/merge_requests/approvals/).
* For more information on Bitbucket pull request reviews and approvals see: [https://confluence.atlassian.com/bitbucket/pull-requests-and-code-review-223220593.html](https://confluence.atlassian.com/bitbucket/pull-requests-and-code-review-223220593.html)
* For more information on Azure DevOps pull request reviews and approvals see: [https://docs.microsoft.com/en-us/azure/devops/repos/git/pull-requests?view=azure-devops&tabs=browser](https://docs.microsoft.com/en-us/azure/devops/repos/git/pull-requests?view=azure-devops&tabs=browser)
