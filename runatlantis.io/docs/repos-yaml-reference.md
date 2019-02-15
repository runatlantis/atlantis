# repos.yaml Reference
[[toc]]

::: tip Do I need a repos.yaml file?
A `repos.yaml` file is only required if you wish to customize some aspect of Atlantis.
You can provide even more customizations by combining a server side `repos.yaml` file with an
`atlantis.yaml` file in a repository.  See the [atlantis.yaml reference](atlantis-yaml-reference.html) for
more information on additional customizations.
:::

## Enabling repos.yaml
The atlantis server must be running with the `--allow-restricted-repo-config` option to allow Atlantis to use
`repos.yaml` and `atlantis.yaml` files.  The location of the `repos.yaml` file on the Atlantis server must be provided
to the `--repos-config` option.

## Overview
The `repos.yaml` file lets you provide customized settings to be applied globally to repositories that match a 
regular expression or an exact string.  An `atlantis.yaml` file allows you to provide more fine grained configuration
about how atlantis should act on a specific repository.

Some of the settings in `repos.yaml` are sensitive in nature, and would allow users to run arbitrary code on the
Atlantis server if they were allowed in the `atlantis.yaml` file.  These settings are restricted to the `repos.yaml`
file by default, but the `repos.yaml` file can allow them to be set in an `atlantis.yaml` file.

The `repos.yaml` file allows a list of repos to be defined, using either a regex or exact string to match a repository
path.  If a repository matches multiple entries, the settings from the last entry in the list take precedence.

## Example Using All Keys
```yaml
repos:
- id: /.*/
  apply_requirements: [approved, mergeable]
  workflow: repoworkflow
  allowed_overrides: [apply_requirements, workflow]
  allow_custom_workflows: true
 
workflows:
  repoworkflow:
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

## Reference


### Top-Level Keys
| Key       | Type                                          | Default | Required | Description                       |
| --------- | --------------------------------------------- | ------- | -------- | --------------------------------- |
| repos     | array[[Repo](repos-yaml-reference.html#repo)] |[]       | no       | List of repos to apply settings to|
| workflows | map[string -> [Workflow](repos-yaml-reference.html#workflow)]

### Repo
| Key                | Type     | Default | Required | Description                                                                                                                                                                                                           |
| ------------------ | -------- | ------- | -------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| id                 | string   | none    | yes      | Value can be a regular expersion when specified as /&lt;regex&gt;/ or an exact string match                                                                                                                           |
| workflow           | string   | none    | no       | A custom workflow.  If not specified, Atlantis will use its default workflow                                                                                                                                          |
| apply_requirements | []string | none    | no       | Requirements that must be satisfied before `atlantis apply` can be run. Currently the only supported requirements are `approved` and `mergeable`. See [Apply Requirements](apply-requirements.html) for more details. |
| allowed_overrides  | []string | none    | no       | A list of restricted keys to allow the `atlantis.yaml` file to override settings from the `repos.yaml` file.  <br /><br />Restricted Keys Include:<br /><ul><li>apply_requirements</li><li>workflow</li></ul>         |

### Workflow
```yaml
plan:
apply:
```

| Key   | Type                                        | Default               | Required |  Description                    |
| ----- | ------------------------------------------- | --------------------- | -------- |  ------------------------------ |
| plan  | [Stage](atlantis-yaml-reference.html#stage) | `steps: [init, plan]` | no       | How to plan for this project.  |
| apply | [Stage](atlantis-yaml-reference.html#stage) | `steps: [apply]`      | no       | How to apply for this project. |

### Stage
```yaml
steps:
- run: custom-command
- init
- plan:
    extra_args: [-lock=false]
```

| Key   | Type                                             | Default | Required | Description                                                                                   |
| ----- | ------------------------------------------------ | ------- | -------- | --------------------------------------------------------------------------------------------- |
| steps | array[[Step](atlantis-yaml-reference.html#step)] | `[]`    | no       | List of steps for this stage. If the steps key is empty, no steps will be run for this stage. |

### Step
#### Built-In Commands: init, plan, apply
Steps can be a single string for a built-in command.
```yaml
- init
- plan
- apply
```
| Key             | Type   | Default | Required | Description                                                                                            |
| --------------- | ------ | ------- | -------- | ------------------------------------------------------------------------------------------------------ |
| init/plan/apply | string | none    | no       | Use a built-in command without additional configuration. Only `init`, `plan` and `apply` are supported |

#### Built-In Command With Extra Args
A map from string to `extra_args` for a built-in command with extra arguments.
```yaml
- init:
    extra_args: [arg1, arg2]
- plan:
    extra_args: [arg1, arg2]
- apply:
    extra_args: [arg1, arg2]
```
| Key             | Type                               | Default | Required | Description                                                                                                                                         |
| --------------- | ---------------------------------- | ------- | -------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| init/plan/apply | map[`extra_args` -> array[string]] | none    | no       | Use a built-in command and append `extra_args`. Only `init`, `plan` and `apply` are supported as keys and only `extra_args` is supported as a value |
#### Custom `run` Command
Or a custom command
```yaml
- run: custom-command
```
| Key | Type   | Default | Required | Description          |
| --- | ------ | ------- | -------- | -------------------- |
| run | string | none    | no       | Run a custom command |

