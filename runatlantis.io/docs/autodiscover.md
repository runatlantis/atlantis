# Autodiscover
Atlantis can be configured to automatically discover Terraform projects in a repository.


## How To Enable
By default the autodiscover is enabled, but you can "enforce it":
1. Server Side:
    ```yaml
    repo:
        ...
        autodiscover:
          enabled: true
        ...
    ```
1. Repo side `atlantis.yaml` file:
    ```yaml
    version: 3
    autodiscover:
        enabled: true
    projects:
    - dir: .
    ```
    :::tip NOTE
    You need allow override in the server side configuration if you would like to disable it for a specific repo in the repo side configuration:
    ```yaml
    repo:
        ...
        allowed_overrides: [autodiscover]
        ...
    ```
    ::::

## How to Disable
You can disable it globally for each repo in the server side.
1. Server Side:
    ```yaml
    repo:
        ...
        autodiscover:
          enabled: false
        ...
    ```
or for a specific repo in the repo side `atlantis.yaml` file:
1. Repo side `atlantis.yaml` file:
    ```yaml
    version: 3
    autodiscover:
        enabled: false
    projects:
    - dir: .
    ```
    :::tip NOTE
    You need allow override in the server side configuration if you would like to disable it for a specific repo in the repo side configuration:
    ```yaml
    repo:
        ...
        allowed_overrides: [autodiscover]
        ...
    ```
    ::::
## Cases
### Multiple Projects
Atlantis will discover all Terraform projects in a repository. For example, if you have a repository with the following structure:
```
|── prod
│   └── project1
│          └── main.tf
|   └── project2
│          └── main.tf
```
Atlantis will discover both `project1` and `project2` and create a plan for each one.

### Multi Server
Image that you have a mono repository with multiples stages, and for each stage you have a different Atlantis server. You can configure each server to discover only the projects that are related to it. For example, if you have a repository with the following structure:
```
|── prod
│   └── project1
│   └── main.tf
|── stage
|── test
└── project1
└── main.tf
```
You can create the atlantis.yaml dynamically in each server to discover only the projects that are related to it. For example, the `prod` server side configuration will have the:
```yaml
repos:
    - id: /.*/
    apply_requirements: [approved, mergeable]
    repo_config_file: atlantis-prod.yaml
``` 

and you can use a [pre_workflow_hooks](./pre-workflow-hooks.md) to create the `atlantis-prod.yaml` file dynamically.