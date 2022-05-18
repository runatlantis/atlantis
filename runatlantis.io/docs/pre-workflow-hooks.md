# Pre Workflow Hooks

Pre workflow hooks can be defined to run scripts before default or custom
workflows are executed. Pre workflow hooks differ from [custom
workflows](custom-workflows.html#custom-run-command) in several ways.

1. Pre workflow hooks do not require for repository configuration to be
   present. This be utilized to [dynamically generate repo configs](pre-workflow-hooks.html#dynamic-repo-config-generation).
2. Pre workflow hooks are run outside of Atlantis commands. Which means
   they do not surface their output back to the PR as a comment.

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
        - run: ./repo-config-generator.sh
```
## Reference
### Custom `run` Command
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
  * `HEAD_COMMIT` - The sha256 that points to the head of the branch that is being pull requested into the base. If the pull request is from Bitbucket Cloud the string will only be 12 characters long because Bitbucket Cloud truncates its commit IDs.
  * `BASE_BRANCH_NAME` - Name of the base branch of the pull request (the branch that the pull request is getting merged into)
  * `PULL_NUM` - Pull request number or ID, ex. `2`.
  * `PULL_AUTHOR` - Username of the pull request author, ex. `acme-user`.
  * `DIR` - The absolute path to the root of the cloned repository. 
  * `USER_NAME` - Username of the VCS user running command, ex. `acme-user`. During an autoplan, the user will be the Atlantis API user, ex. `atlantis`.
:::

