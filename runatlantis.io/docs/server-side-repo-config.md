# Server Side Repo Config

A Server-Side Config file is used for more groups of server config that can't reasonably be expressed through flags.

One such usecase is to control per-repo behaviour
and what users can do in repo-level `atlantis.yaml` files.

## Do I Need A Server-Side Config File?

You do not need a server-side repo config file unless you want to customize
some aspect of Atlantis on a per-repo basis.

Read through the [use-cases](#use-cases) to determine if you need it.

## Enabling Server Side Config

To use server side repo config create a config file, ex. `repos.yaml`, and pass it to
the `atlantis server` command via the `--repo-config` flag, ex. `--repo-config=path/to/repos.yaml`.

If you don't wish to write a config file to disk, you can use the
`--repo-config-json` flag or `ATLANTIS_REPO_CONFIG_JSON` environment variable
to specify your config as JSON. See [--repo-config-json](server-configuration.md#repo-config-json)
for an example.

## Example Server Side Repo

```yaml
# repos lists the config for specific repos.
repos:
  # id can either be an exact repo ID or a regex.
  # If using a regex, it must start and end with a slash.
  # Repo ID's are of the form {VCS hostname}/{org}/{repo name}, ex.
  # github.com/runatlantis/atlantis.
- id: /.*/
  # branch is an regex matching pull requests by base branch
  # (the branch the pull request is getting merged into).
  # By default, all branches are matched
  branch: /.*/

  # repo_config_file specifies which repo config file to use for this repo.
  # By default, atlantis.yaml is used.
  repo_config_file: path/to/atlantis.yaml

  # plan_requirements sets the Plan Requirements for all repos that match.
  plan_requirements: [approved, mergeable, undiverged]

  # apply_requirements sets the Apply Requirements for all repos that match.
  apply_requirements: [approved, mergeable, undiverged]

  # import_requirements sets the Import Requirements for all repos that match.
  import_requirements: [approved, mergeable, undiverged]

  # workflow sets the workflow for all repos that match.
  # This workflow must be defined in the workflows section.
  workflow: custom

  # allowed_overrides specifies which keys can be overridden by this repo in
  # its atlantis.yaml file.
  allowed_overrides: [apply_requirements, workflow, delete_source_branch_on_merge, repo_locking, repo_locks, custom_policy_check]

  # allowed_workflows specifies which workflows the repos that match
  # are allowed to select.
  allowed_workflows: [custom]

  # allow_custom_workflows defines whether this repo can define its own
  # workflows. If false (default), the repo can only use server-side defined
  # workflows.
  allow_custom_workflows: true

  # delete_source_branch_on_merge defines whether the source branch would be deleted on merge
  # If false (default), the source branch won't be deleted on merge
  delete_source_branch_on_merge: true

  # repo_locking defines whether lock repository when planning.
  # If true (default), atlantis try to get a lock.
  # deprecated: use repo_locks instead
  repo_locking: true

  # repo_locks defines whether the repository would be locked on apply instead of plan, or disabled
  # Valid values are on_plan (default), on_apply or disabled.
  repo_locks:
    mode: on_plan

  # custom_policy_check defines whether policy checking tools besides Conftest are enabled in checks
  # If false (default), only Conftest JSON output is allowed
  custom_policy_check: false

  # pre_workflow_hooks defines arbitrary list of scripts to execute before workflow execution.
  pre_workflow_hooks:
    - run: my-pre-workflow-hook-command arg1

  # post_workflow_hooks defines arbitrary list of scripts to execute after workflow execution.
  post_workflow_hooks:
    - run: my-post-workflow-hook-command arg1

  # policy_check defines if policy checking should be enable on this repository.
  policy_check: false

  # autodiscover defines how atlantis should automatically discover projects in this repository.
  autodiscover:
    mode: auto
    # Optionally ignore some paths for autodiscovery by a glob path
    ignore_paths:
      - foo/*

  # id can also be an exact match.
- id: github.com/myorg/specific-repo

# workflows lists server-side custom workflows
workflows:
  custom:
    plan:
      steps:
      - run: my-custom-command arg1 arg2
      - init
      - plan:
          extra_args: ["-lock", "false"]
      - run: my-custom-command arg1 arg2
    apply:
      steps:
      - run: echo hi
      - apply
 ```

## Use Cases

Here are some of the reasons you might want to use a repo config.

### Requiring PR Is Approved Before an applicable subcommand

If you want to require that all (or specific) repos must have pull requests
approved before Atlantis will allow running `apply` or `import`, use the `plan_requirements`, `apply_requirements` or `import_requirements` keys.

For all repos:

```yaml
# repos.yaml
repos:
- id: /.*/
  plan_requirements: [approved]
  apply_requirements: [approved]
  import_requirements: [approved]
```

For a specific repo:

```yaml
# repos.yaml
repos:
- id: github.com/myorg/myrepo
  plan_requirements: [approved]
  apply_requirements: [approved]
  import_requirements: [approved]
```

See [Command Requirements](command-requirements.md) for more details.

### Requiring PR Is "Mergeable" Before Apply or Import

If you want to require that all (or specific) repos must have pull requests
in a mergeable state before Atlantis will allow running `apply` or `import`, use the `plan_requirements`, `apply_requirements` or `import_requirements` keys.

For all repos:

```yaml
# repos.yaml
repos:
- id: /.*/
  plan_requirements: [mergeable]
  apply_requirements: [mergeable]
  import_requirements: [mergeable]
```

For a specific repo:

```yaml
# repos.yaml
repos:
- id: github.com/myorg/myrepo
  plan_requirements: [mergeable]
  apply_requirements: [mergeable]
  import_requirements: [mergeable]
```

See [Command Requirements](command-requirements.md) for more details.

### Repos Can Set Their Own Apply an applicable subcommand

If you want all (or specific) repos to be able to override the default apply requirements, use
the `allowed_overrides` key.

To allow all repos to override the default:

```yaml
# repos.yaml
repos:
- id: /.*/
  # The default will be approved.
  plan_requirements: [approved]
  apply_requirements: [approved]
  import_requirements: [approved]

  # But all repos can set their own using atlantis.yaml
  allowed_overrides: [plan_requirements, apply_requirements, import_requirements]
```

To allow only a specific repo to override the default:

```yaml
# repos.yaml
repos:
# Set a default for all repos.
- id: /.*/
  plan_requirements: [approved]
  apply_requirements: [approved]
  import_requirements: [approved]

# Allow a specific repo to override.
- id: github.com/myorg/myrepo
  allowed_overrides: [plan_requirements, apply_requirements, import_requirements]
```

Then each allowed repo can have an `atlantis.yaml` file that
sets `plan_requirements`, `apply_requirements` or `import_requirements` to an empty array (disabling the requirement).

```yaml
# atlantis.yaml in the repo root or set repo_config_file in repos.yaml
version: 3
projects:
- dir: .
  plan_requirements: []
  apply_requirements: []
  import_requirements: []
```

### Running Scripts Before Atlantis Workflows

If you want to run scripts that would execute before Atlantis can run default or
custom workflows, you can create a `pre-workflow-hooks`:

```yaml
repos:
  - id: /.*/
    pre_workflow_hooks:
      - run: my custom command
      - run: |
          my bash script inline
```

See [Pre Workflow Hooks](pre-workflow-hooks.md) for more details on writing
pre workflow hooks.

### Running Scripts After Atlantis Workflows

If you want to run scripts that would execute after Atlantis runs default or
custom workflows, you can create a `post-workflow-hooks`:

```yaml
repos:
  - id: /.*/
    post_workflow_hooks:
      - run: my custom command
      - run: |
          my bash script inline
```

See [Post Workflow Hooks](post-workflow-hooks.md) for more details on writing
post workflow hooks.

### Change The Default Atlantis Workflow

If you want to change the default commands that Atlantis runs during `plan` and `apply`
phases, you can create a new `workflow`.

If you want to use that workflow by default for all repos, use the workflow
key `default`:

```yaml
# repos.yaml
# NOTE: the repos key is not required.
workflows:
  # It's important that this is "default".
  default:
    plan:
      steps:
      - init
      - run: my custom plan command
    apply:
      steps:
      - run: my custom apply command
```

See [Custom Workflows](custom-workflows.md) for more details on writing
custom workflows.

### Allow Repos To Choose A Server-Side Workflow

If you want repos to be able to choose their own workflows that are defined
in the server-side repo config, you need to create the workflows
server-side and then allow each repo to override the `workflow` key:

```yaml
# repos.yaml
# Allow repos to override the workflow key.
repos:
- id: /.*/
  allowed_overrides: [workflow]

# Define your custom workflows.
workflows:
  custom1:
    plan:
      steps:
      - init
      - run: my custom plan command
    apply:
      steps:
      - run: my custom apply command

  custom2:
    plan:
      steps:
      - run: another custom command
    apply:
      steps:
      - run: another custom command
```

Or, if you want to restrict what workflows each repo has access to, use the `allowed_workflows`
key:

```yaml
# repos.yaml
# Restrict which workflows repos can select.
repos:
- id: /.*/
  allowed_overrides: [workflow]

- id: /my_repo/
  allowed_overrides: [workflow]
  allowed_workflows: [custom1]

# Define your custom workflows.
workflows:
  custom1:
    plan:
      steps:
      - init
      - run: my custom plan command
    apply:
      steps:
      - run: my custom apply command

  custom2:
    plan:
      steps:
      - run: another custom command
    apply:
      steps:
      - run: another custom command
```

Then each allowed repo can choose one of the workflows in their `atlantis.yaml`
files:

```yaml
# atlantis.yaml
version: 3
projects:
- dir: .
  workflow: custom1 # could also be custom2 OR default
```

:::tip NOTE
There is always a workflow named `default` that corresponds to Atlantis' default workflow
unless you've created your own server-side workflow with that key (overriding it).
:::

See [Custom Workflows](custom-workflows.md) for more details on writing
custom workflows.

### Allow Using Custom Policy Tools

Conftest is the standard policy check application integrated with Atlantis, but custom tools can still be run in custom workflows when the `custom_policy_check` option is set.  See the [Custom Policy Checks page](custom-policy-checks.md) for detailed examples.

### Allow Repos To Define Their Own Workflows

If you want repos to be able to define their own workflows you need to
allow them to override the `workflow` key and set `allow_custom_workflows` to `true`.

::: danger
If repos can define their own workflows, then anyone that can create a pull
request to that repo can essentially run arbitrary code on your Atlantis server.
:::

```yaml
# repos.yaml
repos:
- id: /.*/

  # With just allowed_overrides: [workflow], repos can only
  # choose workflows defined server-side.
  allowed_overrides: [workflow]

  # By setting allow_custom_workflows to true, we allow repos to also
  # define their own workflows.
  allow_custom_workflows: true
```

Then each allowed repo can define and use a custom workflow in their `atlantis.yaml` files:

```yaml
# atlantis.yaml
version: 3
projects:
- dir: .
  workflow: custom1
workflows:
  custom1:
    plan:
      steps:
      - init
      - run: my custom plan command
    apply:
      steps:
      - run: my custom apply command
```

See [Custom Workflows](custom-workflows.md) for more details on writing
custom workflows.

### Multiple Atlantis Servers Handle The Same Repository

Running multiple Atlantis servers to handle the same repository can be done to separate permissions for each Atlantis server.
In this case, a different [atlantis.yaml](repo-level-atlantis-yaml.md) repository config file can be used by using different `repos.yaml` files.

For example, consider a situation where a separate `production-server` atlantis uses repo config `atlantis-production.yaml` and `staging-server` atlantis uses repo config `atlantis-staging.yaml`.

Firstly, deploy 2 Atlantis servers, `production-server` and `staging-server`.
Each server has different permissions and a different `repos.yaml` file.
The `repos.yaml` contains `repo_config_file` key to specify the repository atlantis config file path.

```yaml
# repos.yaml
repos:
- id: /.*/
  # for production-server
  repo_config_file: atlantis-production.yaml
  # for staging-server
  # repo_config_file: atlantis-staging.yaml
```

Then, create `atlantis-production.yaml` and `atlantis-staging.yaml` files in the repository.
See the configuration examples in [atlantis.yaml](repo-level-atlantis-yaml.md).

```yaml
# atlantis-production.yaml
version: 3
projects:
- name: project
  branch: /production/
  dir: infrastructure/production
---
# atlantis-staging.yaml
version: 3
projects:
  - name: project
    branch: /staging/
    dir: infrastructure/staging
```

Now, 2 webhook URLs can be setup for the repository, which send events to `production-server` and `staging-server` respectively.
Each servers handle different repository config files.

:::tip Notes

* If `no projects` comments are annoying, set [--silence-no-projects](server-configuration.md#silence-no-projects).
* The command trigger executable name can be reconfigured from `atlantis` to something else by setting [Executable Name](server-configuration.md#executable-name).
* When using different atlantis server vcs users such as `@atlantis-staging`, the comment `@atlantis-staging plan` can be used instead `atlantis plan` to call `staging-server` only.
:::

## Reference

### Top-Level Keys

| Key        | Type                                                  | Default   | Required | Description                                                                           |
|------------|-------------------------------------------------------|-----------|----------|---------------------------------------------------------------------------------------|
| repos      | array[[Repo](#repo)]                                  | see below | no       | List of repos to apply settings to.                                                   |
| workflows  | map[string: [Workflow](custom-workflows.md#workflow)] | see below | no       | Map from workflow name to workflow. Workflows override the default Atlantis commands. |
| policies   | Policies.                                             | none      | no       | List of policy sets to run and associated metadata                                    |
| metrics    | Metrics.                                              | none      | no       | Map of metric configuration                                                           |
| team_authz | [TeamAuthz](#teamauthz)                               | none      | no       | Configuration of team permission checking                                             |

::: tip A Note On Defaults

#### `repos`

`repos` always contains a first element with the Atlantis default config:

```yaml
repos:
- id: /.*/
  branch: /.*/
  plan_requirements: []
  apply_requirements: []
  import_requirements: []
  workflow: default
  allowed_overrides: []
  allow_custom_workflows: false
```

#### `workflows`

`workflows` always contains the Atlantis default workflow under the key `default`:

```yaml
workflows:
  default:
    plan:
      steps: [init, plan]
    apply:
      steps: [apply]
```

This gets merged with whatever config you write.
If you set a workflow with the key `default`, it will override this.
:::

### Repo

| Key                           | Type                    | Default         | Required | Description                                                                                                                                                                                                                                                                                               |
|-------------------------------|-------------------------|-----------------|----------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| id                            | string                  | none            | yes      | Value can be a regular expression when specified as /&lt;regex&gt;/ or an exact string match. Repo IDs are of the form `{vcs hostname}/{org}/{name}`, ex. `github.com/owner/repo`. Hostname is specified without scheme or port. For Bitbucket Server, {org} is the **name** of the project, not the key. |
| branch                        | string                  | none            | no       | An regex matching pull requests by base branch (the branch the pull request is getting merged into). By default, all branches are matched                                                                                                                                                                 |
| repo_config_file              | string                  | none            | no       | Repo config file path in this repo. By default, use `atlantis.yaml` which is located on repository root. When multiple atlantis servers work with the same repo, please set different file names.                                                                                                         |
| workflow                      | string                  | none            | no       | A custom workflow.                                                                                                                                                                                                                                                                                        |
| plan_requirements             | []string                | none            | no       | Requirements that must be satisfied before `atlantis plan` can be run. Currently the only supported requirements are `approved`, `mergeable`, and `undiverged`. See [Command Requirements](command-requirements.md) for more details.                                                                   |
| apply_requirements            | []string                | none            | no       | Requirements that must be satisfied before `atlantis apply` can be run. Currently the only supported requirements are `approved`, `mergeable`, and `undiverged`. See [Command Requirements](command-requirements.md) for more details.                                                                  |
| import_requirements           | []string                | none            | no       | Requirements that must be satisfied before `atlantis import` can be run. Currently the only supported requirements are `approved`, `mergeable`, and `undiverged`. See [Command Requirements](command-requirements.md) for more details.                                                                 |
| allowed_overrides             | []string                | none            | no       | A list of restricted keys that `atlantis.yaml` files can override. The only supported keys are `apply_requirements`, `workflow`, `delete_source_branch_on_merge`,`repo_locking`, `repo_locks`, and `custom_policy_check`                                                                                  |
| allowed_workflows             | []string                | none            | no       | A list of workflows that `atlantis.yaml` files can select from.                                                                                                                                                                                                                                           |
| allow_custom_workflows        | bool                    | false           | no       | Whether or not to allow [Custom Workflows](custom-workflows.md).                                                                                                                                                                                                                                        |
| delete_source_branch_on_merge | bool                    | false           | no       | Whether or not to delete the source branch on merge.                                                                                                                                                                                                                                                      |
| repo_locking                  | bool                    | false           | no       | (deprecated) Whether or not to get a lock.                                                                                                                                                                                                                                                                |
| repo_locks                    | [RepoLocks](#repolocks) | `mode: on_plan` | no       | Whether or not repository locks are enabled for this project on plan or apply. See [RepoLocks](#repolocks) for more details.                                                                                                                                                                              |
| policy_check                  | bool                    | false           | no       | Whether or not to run policy checks on this repository.                                                                                                                                                                                                                                                   |
| custom_policy_check           | bool                    | false           | no       | Whether or not to enable custom policy check tools outside of Conftest on this repository.                                                                                                                                                                                                                |
| autodiscover                  | AutoDiscover            | none            | no       | Auto discover settings for this repo                                                                                                                                                                                                                                                                      |
| silence_pr_comments           | []string                | none            | no       | Silence PR comments from defined stages while preserving PR status checks. Useful in large environments with many Atlantis instances and/or projects, when the comments are too big and too many, therefore it is preferable to rely solely on PR status checks. Supported values are: `plan`, `apply`.   |

:::tip Notes

* If multiple repos match, the last match will apply.
* If a key isn't defined, it won't override a key that matched from above.
  For example, given a repo ID `github.com/owner/repo` and a config:

  ```yaml
  repos:
  - id: /.*/
    allow_custom_workflows: true
    apply_requirements: [approved]
  - id: github.com/owner/repo
    apply_requirements: []
  ```

  The final config will look like:

  ```yaml
  apply_requirements: []
  workflow: default
  allowed_overrides: []
  allow_custom_workflows: true
  ```

  Where
  * `apply_requirements` is set from the `id: github.com/owner/repo` config because
    it overrides the previous matching config from `id: /.*/`.
  * `workflow` is set from the default config that always
    exists.
  * `allowed_overrides` is set from the default config that always
    exists.
  * `allow_custom_workflows` is set from the `id: /.*/` config and isn't unset
    by the `id: github.com/owner/repo` config because it didn't define that key.
:::

### RepoLocks

```yaml
mode: on_apply
```

| Key  | Type   | Default   | Required | Description                                                                                                                           |
|------|--------|-----------|----------|---------------------------------------------------------------------------------------------------------------------------------------|
| mode | `Mode` | `on_plan` | no       | Whether or not repository locks are enabled for this project on plan or apply. Valid values are `disabled`, `on_plan` and `on_apply`. |

### Policies

| Key                    | Type            | Default | Required  | Description                                              |
|------------------------|-----------------|---------|-----------|----------------------------------------------------------|
| conftest_version       | string          | none    | no        | conftest version to run all policy sets                  |
| owners                 | Owners(#Owners) | none    | yes       | owners that can approve failing policies                 |
| approve_count          | int             | 1       | no        | number of approvals required to bypass failing policies. |
| policy_sets            | []PolicySet     | none    | yes       | set of policies to run on a plan output                  |

### Owners

| Key         | Type              | Default | Required   | Description                                             |
|-------------|-------------------|---------|------------|---------------------------------------------------------|
| users       | []string          | none    | no         | list of github users that can approve failing policies  |
| teams       | []string          | none    | no         | list of github teams that can approve failing policies  |

### PolicySet

| Key                  | Type   | Default | Required | Description                                                                                                   |
| ------               | ------ | ------- | -------- | --------------------------------------------------------------------------------------------------------------|
| name                 | string | none    | yes      | unique name for the policy set                                                                                |
| path                 | string | none    | yes      | path to the rego policies directory                                                                           |
| source               | string | none    | yes      | only `local` is supported at this time                                                                        |
| prevent_self_approve | bool   | false   | no       | Whether or not the author of PR can approve policies. Defaults to `false` (the author must also be in owners) |

### Metrics

| Key                    | Type                      | Default | Required  | Description                              |
|------------------------|---------------------------|---------|-----------|------------------------------------------|
| statsd                 | [Statsd](#statsd)         | none    | no        | Statsd metrics provider                  |
| prometheus             | [Prometheus](#prometheus) | none    | no        | Prometheus metrics provider              |

### Statsd

| Key    | Type   | Default | Required | Description                            |
| ------ | ------ | ------- | -------- | -------------------------------------- |
| host   | string | none    | yes      | statsd host ip address                 |
| port   | string | none    | yes      | statsd port                            |

### Prometheus

| Key      | Type   | Default | Required | Description                            |
| -------- | ------ | ------- | -------- | -------------------------------------- |
| endpoint | string | none    | yes      | path to metrics endpoint               |

### TeamAuthz

| Key     | Type     | Default | Required | Description                                 |
|---------|----------|---------|----------|---------------------------------------------|
| command | string   | none    | yes      | full path to external authorization command |
| args    | []string | none    | no       | optional arguments to pass to `command`     |
