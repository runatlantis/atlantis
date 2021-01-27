# Pre Workflow Hooks

Pre workflow hooks can be defined to run scripts right before default or custom
workflows are executed. 

::: tip Note
Defined pre workflow hooks will be executed only once when a new pull 
request is opened, and later only if this pull request has been updated.
The pre workflow hooks are not run on manual `atlantis plan` or
`atlantis apply` commands.
::: 

[[toc]]

## Usage
Pre workflow hooks can only be specified in the Server-Side Repo Config under
`repos` key. 
::: tip Note
`pre-workflow-hooks` do not prevent Atlantis from executing its
workflows(`plan`, `apply`) even if a `run` command exits with an error.
::: 

## Use Cases
### Dynamic Repo Config Generation
If you want generate your `atlantis.yaml` before Atlantis can parse it. You
can add a `run` command to `pre_workflow_hooks`. Your Repo config will be generated
right before Atlantis can parse it.

```yaml
repos:
    - id: /.*/
      pre_workflow_hooks:
        - run: ./repo-config-genarator.sh
```
### Reference
#### Custom `run` Command
This is very similar to [custom workflow run
command](custom-workflows.html#custom-run-command). 
```yaml
- run: custom-command
```
| Key | Type   | Default | Required | Description          |
| --- | ------ | ------- | -------- | -------------------- |
| run | string | none    | no       | Run a custom command |

::: tip Notes
* `run` commands are executed with the following environment variables:
  * `BASE_REPO_NAME` - Name of the repository that the pull request will be merged into, ex. `atlantis`.
  * `BASE_REPO_OWNER` - Owner of the repository that the pull request will be merged into, ex. `runatlantis`.
  * `HEAD_REPO_NAME` - Name of the repository that is getting merged into the base repository, ex. `atlantis`.
  * `HEAD_REPO_OWNER` - Owner of the repository that is getting merged into the base repository, ex. `acme-corp`.
  * `HEAD_BRANCH_NAME` - Name of the head branch of the pull request (the branch that is getting merged into the base)
  * `BASE_BRANCH_NAME` - Name of the base branch of the pull request (the branch that the pull request is getting merged into)
  * `PULL_NUM` - Pull request number or ID, ex. `2`.
  * `PULL_AUTHOR` - Username of the pull request author, ex. `acme-user`.
  * `DIR` - The absolute path to the root of the cloned repository. 
  * `USER_NAME` - Username of the VCS user running command, ex. `acme-user`. During an autoplan, the user will be the Atlantis API user, ex. `atlantis`.
:::

