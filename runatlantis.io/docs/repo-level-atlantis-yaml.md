# Repo Level atlantis.yaml Config

An `atlantis.yaml` file specified at the root of a Terraform repo allows you
to instruct Atlantis on the structure of your repo and set custom workflows.

## Do I need an atlantis.yaml file?

`atlantis.yaml` files are only required if you wish to customize some aspect of Atlantis.
The default Atlantis config works for many users without changes.

Read through the [use-cases](#use-cases) to determine if you need it.

## Enabling atlantis.yaml

By default, all repos are allowed to have an `atlantis.yaml` file,
but some of the keys are restricted by default.

Restricted keys can be set in the server-side `repos.yaml` repo config file.
You can enable `atlantis.yaml` to override restricted
keys by setting the `allowed_overrides` key there. See the [Server Side Repo Config](server-side-repo-config.md) for
more details.

**Notes:**

- By default, repo root `atlantis.yaml` file is used.
- You can change this behaviour by setting [Server Side Repo Config](server-side-repo-config.md)

::: danger DANGER
Atlantis uses the `atlantis.yaml` version from the pull request, similar to other
CI/CD systems. If you're allowing users to [create custom workflows](server-side-repo-config.md#allow-repos-to-define-their-own-workflows)
then this means
anyone that can create a pull request to your repo can run arbitrary code on the
Atlantis server.

By default, this is not allowed.
:::

::: warning
Once an `atlantis.yaml` file exists in a repo and one or more `projects` are configured,
Atlantis won't try to determine where to run plan automatically. Instead it will just
follow the project configuration. This means that you'll need to define each project
in your repo.

If you have many directories with Terraform configuration, each directory will
need to be defined.

This behavior can be overridden by setting `autodiscover.mode` to
`enabled` in which case Atlantis will still try to discover projects which were not
explicitly configured. If the directory of any discovered project conflicts with a
manually configured project, the manually configured project will take precedence.
:::

## Example Using All Keys

```yaml
version: 3 # Available since v0.1.0
automerge: true # Available since v0.15.0
autodiscover: # Available since v0.18.0
  mode: auto
  ignore_paths:
  - some/path
delete_source_branch_on_merge: true # Available since v0.15.0
parallel_plan: true # Available since v0.17.0
parallel_apply: true # Available since v0.17.0
abort_on_execution_order_fail: true # Available since v0.17.0
projects:
- name: my-project-name # Available since v0.1.0
  branch: /main/ # Available since v0.21.0
  dir: . # Available since v0.1.0
  workspace: default # Available since v0.1.0
  terraform_distribution: terraform # Available since v0.25.0
  terraform_version: v0.11.0 # Available since v0.1.0
  delete_source_branch_on_merge: true # Available since v0.17.0
  repo_locking: true # deprecated: use repo_locks instead, Available since v0.17.0
  repo_locks: # Available since v0.17.0
    mode: on_plan
  custom_policy_check: false # Available since v0.17.0
  autoplan: # Available since v0.1.0
    when_modified: ["*.tf", "../modules/**/*.tf", ".terraform.lock.hcl"]
    enabled: true
  plan_requirements: [mergeable, approved, undiverged] # Available since v0.17.0
  apply_requirements: [mergeable, approved, undiverged] # Available since v0.17.0
  import_requirements: [mergeable, approved, undiverged] # Available since v0.17.0
  silence_pr_comments: ["apply"] # Available since v0.17.0
  execution_order_group: 1 # Available since v0.17.0
  depends_on: # Available since v0.20.0
    - project-1
  workflow: myworkflow # Available since v0.17.0
workflows: # Available since v0.1.0
  myworkflow:
    plan:
      steps:
      - run: my-custom-command arg1 arg2
      - run:
          command: my-custom-command arg1 arg2
          output: hide
      - init
      - plan:
          extra_args: ["-lock", "false"]
      - run: my-custom-command arg1 arg2
    apply:
      steps:
      - run: echo hi
      - apply
allowed_regexp_prefixes: # Available since v0.19.0
- dev/
- staging/
```

## Example of DRYing up projects using YAML anchors

```yaml
projects:
   - &template
     name: template
     dir: template
     workflow: custom
     autoplan:
        enabled: true
        when_modified:
           - "./terraform/modules/**/*.tf"
           - "**/*.tf"
           - ".terraform.lock.hcl"

   - <<: *template
     name: ue1-prod-titan
     dir: ./terraform/titan
     workspace: ue1-prod

   - <<: *template
     name: ue1-stage-titan
     dir: ./terraform/titan
     workspace: ue1-stage

   - <<: *template
     name: ue1-dev-titan
     dir: ./terraform/titan
     workspace: ue1-dev
```

## Auto generate projects

This is useful if you have many projects in a repository. This assumes the `default` workspace (or no workspace).

Run this in the root of your repository. This will use gnu `grep` to search terraform files for an S3 backend (terraform dir), retrieve the directory path, retrieve the unique entries, and then use `yq` to return the YAML of a simple project dir setup which can then be modified to your liking.

```sh
grep -P 'backend[\s]+"s3"' **/*.tf |
  rev | cut -d'/' -f2- | rev |
  sort |
  uniq |
  while read d; do \
    echo '[ {"name": "'"$d"'","dir": "'"$d"'", "autoplan": {"when_modified": ["**/*.tf.*"] }} ]' | yq -PM; \
  done
```

## Use Cases

### Disabling Autoplanning

```yaml
version: 3
projects:
   - dir: project1
     autoplan:
        enabled: false
```

This will stop Atlantis automatically running plan when `project1/` is updated
in a pull request.

### Run plans and applies in parallel

```yaml
version: 3
parallel_plan: true
parallel_apply: true
```

This will run plans and applies for all of your projects in parallel.

Enabling these options can significantly reduce the duration of plans and applies, especially for repositories with many projects.

Use the `--parallel-pool-size` to configure the max number of plans and applies that can run in parallel. The default is 15.

Parallel plans and applies work across both multiple directories and multiple workspaces.

### Configuring Planning

Given the directory structure:

```plain
.
├── modules
│   └── module1
│       ├── main.tf
│       ├── outputs.tf
│       └── submodule
│           ├── main.tf
│           └── outputs.tf
└── project1
    └── main.tf
```

If you want Atlantis to plan `project1/` whenever any `.tf` files under `module1/` change or any `.tf` or `.tfvars` files under `project1/` change you could use the following configuration:

```yaml
version: 3
projects:
   - dir: project1
     autoplan:
        when_modified: ["../modules/**/*.tf", "*.tf*", ".terraform.lock.hcl"]
```

Note:

- `when_modified` uses the [`.dockerignore` syntax](https://docs.docker.com/engine/reference/builder/#dockerignore-file)
- The paths are relative to the project's directory.
- `when_modified` will be used by both automatic and manually run plans.
- `when_modified` will continue to work for manually run plans even when autoplan is disabled.

### Supporting Terraform Workspaces

```yaml
version: 3
projects:
   - dir: project1
     workspace: staging
   - dir: project1
     workspace: production
```

With the above config, when Atlantis determines that the configuration for the `project1` dir has changed,
it will run plan for both the `staging` and `production` workspaces.

If you want to `plan` or `apply` for a specific workspace you can use

```shell
atlantis plan -w staging -d project1
```

and

```shell
atlantis apply -w staging -d project1
```

### Using .tfvars files

See [Custom Workflow Use Cases: Using .tfvars files](custom-workflows.md#tfvars-files)

### Adding extra arguments to Terraform commands

See [Custom Workflow Use Cases: Adding extra arguments to Terraform commands](custom-workflows.md#adding-extra-arguments-to-terraform-commands)

### Custom init/plan/apply Commands

See [Custom Workflow Use Cases: Custom init/plan/apply Commands](custom-workflows.md#custom-init-plan-apply-commands)

### Terragrunt

See [Custom Workflow Use Cases: Terragrunt](custom-workflows.md#terragrunt)

### Running custom commands

See [Custom Workflow Use Cases: Running custom commands](custom-workflows.md#running-custom-commands)

### Terraform Distributions

If you'd like to use a different distribution of Terraform than what is set
by the `--default-tf-version` flag, then set the `terraform_distribution` key:

```yaml
version: 3
projects:
   - dir: project1
     terraform_distribution: opentofu
```

Atlantis will automatically download and use this distribution. Valid values are `terraform` and `opentofu`.

### Terraform Versions

If you'd like to use a different version of Terraform than what is in Atlantis'
`PATH` or is set by the `--default-tf-version` flag, then set the `terraform_version` key:

```yaml
version: 3
projects:
   - dir: project1
     terraform_version: 0.10.0
```

Atlantis will automatically download and use this version.

### Requiring Approvals For Production

In this example, we only want to require `apply` approvals for the `production` directory.

```yaml
version: 3
projects:
   - dir: staging
   - dir: production
     plan_requirements: [approved]
     apply_requirements: [approved]
     import_requirements: [approved]
```

:::warning
`plan_requirements`, `apply_requirements` and `import_requirements` are restricted keys so this repo will need to be configured
to be allowed to set this key. See [Server-Side Repo Config Use Cases](server-side-repo-config.md#repos-can-set-their-own-apply-an-applicable-subcommand).
:::

### Order of planning/applying

```yaml
version: 3
abort_on_execution_order_fail: true
projects:
   - dir: project1
     execution_order_group: 2
   - dir: project2
     execution_order_group: 1
```

With this config above, Atlantis runs planning/applying for project2 first, then for project1.
Several projects can have same `execution_order_group`. Any order in one group isn't guaranteed.
`parallel_plan` and `parallel_apply` respect these order groups, so parallel planning/applying works
in each group one by one.

If any plan/apply fails and `abort_on_execution_order_fail` is set to true on a repo level, all the
following groups will be aborted. For this example, if project2 fails then project1 will not run.

Execution order groups are useful when you have dependencies between projects. However, they are only applicable in the case where
you initiate a global apply for all of your projects, i.e `atlantis apply`. If you initiate an apply on a single project, then the execution order groups are ignored.
Thus, the `depends_on` key is more useful in this case. and can be used in conjunction with execution order groups.

The following configuration is an example of how to use execution order groups and depends_on together to enforce dependencies between projects.

```yaml
version: 3
projects:
   - name: development
     dir: .
     autoplan:
        when_modified: ["*.tf", "vars/development.tfvars"]
     execution_order_group: 1
     workspace: development
     workflow: infra
   - name: staging
     dir: .
     autoplan:
        when_modified: ["*.tf", "vars/staging.tfvars"]
     depends_on: ["development"]
     execution_order_group: 2
     workspace: staging
     workflow: infra
   - name: production
     dir: .
     autoplan:
        when_modified: ["*.tf", "vars/production.tfvars"]
     depends_on: ["staging"]
     execution_order_group: 3
     workspace: production
     workflow: infra
```

the `depends_on` feature will make sure that `production` is not applied before `staging` for example.

::: tip
What Happens if one or more project's dependencies are not applied?

If there's one or more projects in the dependency list which is not in applied status, users will see an error message like this:
`Can't apply your project unless you apply its dependencies`
:::

### Order of destroying

```yaml
version: 3
abort_on_execution_order_fail: true
projects:
  - dir: project1
    destroy_execution_order_group: 2
  - dir: project2
    destroy_execution_order_group: 1
  - dir: project3
    destroy_execution_order_group: 0
---
version: 3
abort_on_execution_order_fail: true
projects:
  - dir: project1
    execution_order_group: 0
    destroy_execution_order_group: 2
  - dir: project2
    execution_order_group: 1
    destroy_execution_order_group: 1
  - dir: project3
    execution_order_group: 2
    destroy_execution_order_group: 0
```

Notes:

- It’s necessary to configure a workflow to handle the destructive commands.
- To orchestrate multiple modules within a single root module, use a Terraform-based or OpenTofu-based workflow.
- To orchestrate multiple modules distributed across separate paths, use a Terragrunt-based workflow.

::: tip

#### `terraform`

```yaml
# repos.yaml
workflows:
  terraform:
    plan:
      steps:
      # Use the -destroy flag in plan. e.g. atlantis plan -- -destroy
      - run:
          command: terraform plan $(printf '%s' $COMMENT_ARGS | sed 's/,/ /g' | tr -d '\\') -out=$PLANFILE
          output: strip_refreshing
    apply:
      steps:
      # Use the -destroy flag in apply. e.g. atlantis apply -- -destroy
      - run:
          command: terraform apply $(printf '%s' $COMMENT_ARGS | sed 's/,/ /g' | tr -d '\\') -auto-approve $PLANFILE
          output: strip_refreshing
```

#### `opentofu`

```yaml
# repos.yaml
workflows:
  opentofu:
    plan:
      steps:
      # Use the -destroy flag in plan. e.g. atlantis plan -- -destroy
      - run:
          command: tofu plan $(printf '%s' $COMMENT_ARGS | sed 's/,/ /g' | tr -d '\\') -out=$PLANFILE
          output: strip_refreshing
    apply:
      steps:
      # Use the -destroy flag in apply. e.g. atlantis apply -- -destroy
      - run:
          command: tofu apply $(printf '%s' $COMMENT_ARGS | sed 's/,/ /g' | tr -d '\\') -auto-approve $PLANFILE
          output: strip_refreshing
```

#### `terragrunt`

```yaml
# repos.yaml
workflows:
  terragrunt:
    plan:
      steps:
      # Use the -destroy flag in plan. e.g. atlantis plan -- -destroy
      - run:
          command: terragrunt plan $(printf '%s' $COMMENT_ARGS | sed 's/,/ /g' | tr -d '\\') -out=$PLANFILE
          output: strip_refreshing
    apply:
      steps:
      # Use the -destroy flag in apply. e.g. atlantis apply -- -destroy
      - run:
          command: terragrunt apply $(printf '%s' $COMMENT_ARGS | sed 's/,/ /g' | tr -d '\\') -auto-approve
          output: strip_refreshing
```

- The use of the native environment variable `COMMENT_ARGS` is required using `-destroy` flag.
  This variable contains any additional flags passed in the pull-request comment.

- The use of the native environment variable `PLANFILE` isn’t required when Terragrunt modules have dependencies.
  This variable can be used to override the built-in `plan`/`apply` commands.
:::

After use `atlantis plan -- -destroy` and `atlantis apply -- -destroy`, Atlantis will execute destroy operations for project3 first, then project2, and finally project1. Multiple projects can share the same `destroy_execution_order_group`, but the order within a single group is not guaranteed. Both `parallel_plan` and `parallel_apply` respect these order groups, so parallel destroying operations will run group by group.

If any destructive plan/apply fails and `abort_on_execution_order_fail` is set to true at the repo level, all subsequent groups will be aborted. In this example, if project3 fails, then project2 and project1 will not run. At any time, it is possible to switch between destructive and non-destructive `plan`/`apply` operations within the same PR.

Rules:

- If all projects define `destroy_execution_order_group`, use only those values.
- If no project defines `destroy_execution_order_group`, use `execution_order_group`.
- If some projects define `destroy_execution_order_group`, they will use that; the others will use `execution_order_group`.
- Non-destructive plan/apply always use `execution_order_group`.

::: warning
The default use of `execution_order_group` for destructive plan/apply operations is applied when `destroy_execution_order_group` is not specified. However, defining it explicitly is required, as demonstrated in the destroying order.
:::

### Autodiscovery Config

```yaml
autodiscover:
   mode: "auto"
```

The above is the default configuration for `autodiscover.mode`. When `autodiscover.mode` is auto,
projects will be discovered only if the repo has no `projects` configured.

```yaml
autodiscover:
   mode: "disabled"
```

With the config above, Atlantis will never try to discover projects, even when there are no
`projects` configured. This is useful if dynamically generating Atlantis config in pre_workflow hooks.
See [Dynamic Repo Config Generation](pre-workflow-hooks.md#dynamic-repo-config-generation).

```yaml
autodiscover:
   mode: "enabled"
```

With the config above, Atlantis will unconditionally try to discover projects based on modified_files,
even when the directory of the project is missing from the configured `projects` in the repo configuration.
If a discovered project has the same directory as a project which was manually configured in `projects`,
the manual configuration will take precedence.

Use this feature when some projects require specific configuration in a repo with many projects yet
it's still desirable for Atlantis to plan/apply for projects not enumerated in the config.

This setting is ignored if it is configured on the server, see [Server Side Repo Config](server-side-repo-config.md#repo)

```yaml
autodiscover:
   mode: "enabled"
   ignore_paths:
      - dir/*
```

Autodiscover can also be configured to skip over directories that match a path glob (as defined [here](https://pkg.go.dev/github.com/bmatcuk/doublestar/v4))

### Custom Backend Config

See [Custom Workflow Use Cases: Custom Backend Config](custom-workflows.md#custom-backend-config)

## Reference

### Top-Level Keys

```yaml
version: 3
automerge: false
delete_source_branch_on_merge: false
projects:
workflows:
allowed_regexp_prefixes:
```

| Key                           | Type                                                   | Default | Required | Description                                                                                                                        |
| ----------------------------- | ------------------------------------------------------ | ------- | -------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| version                       | int                                                    | none    | **yes**  | This key is required and must be set to `3`.                                                                                       |
| automerge                     | bool                                                   | `false` | no       | Automatically merges pull request when all plans are applied.                                                                      |
| delete_source_branch_on_merge | bool                                                   | `false` | no       | Automatically deletes the source branch on merge.                                                                                  |
| projects                      | array[[Project](repo-level-atlantis-yaml.md#project)]  | `[]`    | no       | Lists the projects in this repo.                                                                                                   |
| workflows<br />_(restricted)_ | map[string: [Workflow](custom-workflows.md#reference)] | `{}`    | no       | Custom workflows.                                                                                                                  |
| allowed_regexp_prefixes       | array\[string\]                                        | `[]`    | no       | Lists the allowed regexp prefixes to use when the [`--enable-regexp-cmd`](server-configuration.md#enable-regexp-cmd) flag is used. |

### Project

```yaml
name: myname
branch: /mybranch/
dir: mydir
workspace: myworkspace
execution_order_group: 0
delete_source_branch_on_merge: false
repo_locking: true # deprecated: use repo_locks instead
repo_locks:
   mode: on_plan
custom_policy_check: false
autoplan:
terraform_version: 0.11.0
plan_requirements: ["approved"]
apply_requirements: ["approved"]
import_requirements: ["approved"]
silence_pr_comments: ["apply"]
workflow: myworkflow
```

| Key                                     | Type                    | Default         | Required | Description                                                                                                                                                                                                                             |
| --------------------------------------- | ----------------------- | --------------- | -------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| name                                    | string                  | none            | maybe    | Required if there is more than one project with the same `dir` and `workspace`. This project name can be used with the `-p` flag.                                                                                                       |
| branch                                  | string                  | none            | no       | Regex matching projects by the base branch of pull request (the branch the pull request is getting merged into). Only projects that match the PR's branch will be considered. By default, all branches are matched.                     |
| dir                                     | string                  | none            | **yes**  | The directory of this project relative to the repo root. For example if the project was under `./project1` then use `project1`. Use `.` to indicate the repo root.                                                                      |
| workspace                               | string                  | `"default"`     | no       | The [Terraform workspace](https://developer.hashicorp.com/terraform/language/state/workspaces) for this project. Atlantis will switch to this workplace when planning/applying and will create it if it doesn't exist.                  |
| execution_order_group                   | int                     | `0`             | no       | Index of execution order group. Projects will be sort by this field before planning/applying.                                                                                                                                           |
| destroy_execution_order_group           | int                     | none            | no       | Index of destroy execution order group. Projects will be sort by this field before destroying.                                                                                                                                          |
| delete_source_branch_on_merge           | bool                    | `false`         | no       | Automatically deletes the source branch on merge.                                                                                                                                                                                       |
| repo_locking                            | bool                    | `true`          | no       | (deprecated) Get a repository lock in this project when plan.                                                                                                                                                                           |
| repo_locks                              | [RepoLocks](#repolocks) | `mode: on_plan` | no       | Get a repository lock in this project on plan or apply. See [RepoLocks](#repolocks) for more details.                                                                                                                                   |
| custom_policy_check                     | bool                    | `false`         | no       | Enable using policy check tools other than Conftest                                                                                                                                                                                     |
| autoplan                                | [Autoplan](#autoplan)   | none            | no       | A custom autoplan configuration. If not specified, will use the autoplan config. See [Autoplanning](autoplanning.md).                                                                                                                   |
| terraform_version                       | string                  | none            | no       | A specific Terraform version to use when running commands for this project. Must be [Semver compatible](https://semver.org/), ex. `v0.11.0`, `0.12.0-beta1`.                                                                            |
| plan_requirements<br />_(restricted)_   | array\[string\]         | none            | no       | Requirements that must be satisfied before `atlantis plan` can be run. Currently the only supported requirements are `approved`, `mergeable`, and `undiverged`. See [Command Requirements](command-requirements.md) for more details.   |
| apply_requirements<br />_(restricted)_  | array\[string\]         | none            | no       | Requirements that must be satisfied before `atlantis apply` can be run. Currently the only supported requirements are `approved`, `mergeable`, and `undiverged`. See [Command Requirements](command-requirements.md) for more details.  |
| import_requirements<br />_(restricted)_ | array\[string\]         | none            | no       | Requirements that must be satisfied before `atlantis import` can be run. Currently the only supported requirements are `approved`, `mergeable`, and `undiverged`. See [Command Requirements](command-requirements.md) for more details. |
| silence_pr_comments                     | array\[string\]         | none            | no       | Silence PR comments from defined stages while preserving PR status checks. Supported values are: `plan`, `apply`.                                                                                                                       |
| workflow <br />_(restricted)_           | string                  | none            | no       | A custom workflow. If not specified, Atlantis will use its default workflow.                                                                                                                                                            |

::: tip
A project represents a Terraform state. Typically, there is one state per directory and workspace however it's possible to
have multiple states in the same directory using `terraform init -backend-config=custom-config.tfvars`.
Atlantis supports this but requires the `name` key to be specified. See [Custom Backend Config](custom-workflows.md#custom-backend-config) for more details.
:::

### Autoplan

```yaml
enabled: true
when_modified: ["*.tf", "terragrunt.hcl", ".terraform.lock.hcl"]
```

| Key           | Type            | Default        | Required | Description                                                                                                                                                                                                                                                     |
| ------------- | --------------- | -------------- | -------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| enabled       | boolean         | `true`         | no       | Whether autoplanning is enabled for this project.                                                                                                                                                                                                               |
| when_modified | array\[string\] | `["**/*.tf*"]` | no       | Uses [.dockerignore](https://docs.docker.com/engine/reference/builder/#dockerignore-file) syntax. If any modified file in the pull request matches, this project will be planned. See [Autoplanning](autoplanning.md). Paths are relative to the project's dir. |

### RepoLocks

```yaml
mode: on_apply
```

| Key  | Type   | Default   | Required | Description                                                                                                                           |
| ---- | ------ | --------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------- |
| mode | `Mode` | `on_plan` | no       | Whether or not repository locks are enabled for this project on plan or apply. Valid values are `disabled`, `on_plan` and `on_apply`. |
