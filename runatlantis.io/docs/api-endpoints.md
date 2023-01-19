# API Endpoints

Aside from interacting via pull request comments, Atlantis could respond to a limited number of API endpoints.
The API endpoints are disabled by default, since the API endpoint could change the infrastructure directly.
To enable API endpoints, `api-secret` should be configured.

:::tip Prerequisites

* You have set `api-secret` as part of the [Server Configuration](server-configuration.html#api-secret)
* You have pass `X-Atlantis-Token` with the same secret in the request header
  :::

## Endpoints

### POST /api/plan

#### Description

Execute [atlantis plan](using-atlantis.html#atlantis-plan) on the specified repository.

#### Parameters

| Name       | Type              | Required | Description                              |
|------------|-------------------|----------|------------------------------------------|
| Repository | string            | Yes      | Name of the Terraform repository         |
| Ref        | string            | Yes      | Git reference, like a branch name        |
| Type       | string            | Yes      | Type of the VCS provider (Github/Gitlab) |
| Paths      | [ [Path](#Path) ] | Yes      | Paths to the projects to run the plan    |

##### Path

Similar to the [Options](using-atlantis.html#options) of `atlantis plan`. Path specifies which directory/workspace
within the repository to run the plan.  
At least one of `Directory` or `Workspace` should be specified.

| Name      | Type   | Required | Description                                                                                                                                               |
|-----------|--------|----------|-----------------------------------------------------------------------------------------------------------------------------------------------------------|
| Directory | string | No       | Which directory to run plan in relative to root of repo                                                                                                   |
| Workspace | string | No       | [Terraform workspace](https://developer.hashicorp.com/terraform/language/state/workspaces) of the plan. Use `default` if Terraform workspaces are unused. |

#### Sample Request Body

```json
{
  "Repository": "repo-name",
  "Ref": "main",
  "Type": "Github",
  "Paths": [
    {
      "Directory": ".",
      "Workspace": "default"
    }
  ]
}
```

#### Sample Response

```json
{
  "Error": null,
  "Failure": "",
  "ProjectResults": [
    {
      "Command": 1,
      "RepoRelDir": ".",
      "Workspace": "default",
      "Error": null,
      "Failure": "",
      "PlanSuccess": {
        "TerraformOutput": "<redacted>",
        "LockURL": "<redacted>",
        "RePlanCmd": "atlantis plan -d .",
        "ApplyCmd": "atlantis apply -d .",
        "HasDiverged": false
      },
      "PolicyCheckSuccess": null,
      "ApplySuccess": "",
      "VersionSuccess": "",
      "ProjectName": ""
    }
  ],
  "PlansDeleted": false
}
```

### POST /api/apply

#### Description

Execute [atlantis apply](using-atlantis.html#atlantis-apply) on the specified repository.

#### Parameters

| Name       | Type              | Required | Description                              |
|------------|-------------------|----------|------------------------------------------|
| Repository | string            | Yes      | Name of the Terraform repository         |
| Ref        | string            | Yes      | Git reference, like a branch name        |
| Type       | string            | Yes      | Type of the VCS provider (Github/Gitlab) |
| Paths      | [ [Path](#Path) ] | Yes      | Paths to the projects to run the apply   |

##### Path

Similar to the [Options](using-atlantis.html#options-1) of `atlantis apply`. Path specifies which directory/workspace
within the repository to run the apply.  
At least one of `Directory` or `Workspace` should be specified.

| Name      | Type   | Required | Description                                                                                                                                               |
|-----------|--------|----------|-----------------------------------------------------------------------------------------------------------------------------------------------------------|
| Directory | string | No       | Which directory to run apply in relative to root of repo                                                                                                  |
| Workspace | string | No       | [Terraform workspace](https://developer.hashicorp.com/terraform/language/state/workspaces) of the plan. Use `default` if Terraform workspaces are unused. |

#### Sample Request Body

```json
{
  "Repository": "repo-name",
  "Ref": "main",
  "Type": "Github",
  "Paths": [
    {
      "Directory": ".",
      "Workspace": "default"
    }
  ]
}
```

#### Sample Response

```json
{
  "Error": null,
  "Failure": "",
  "ProjectResults": [
    {
      "Command": 0,
      "RepoRelDir": ".",
      "Workspace": "default",
      "Error": null,
      "Failure": "",
      "PlanSuccess": null,
      "PolicyCheckSuccess": null,
      "ApplySuccess": "<redacted>",
      "VersionSuccess": "",
      "ProjectName": ""
    }
  ],
  "PlansDeleted": false
}
```
