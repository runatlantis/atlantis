# API Endpoints

Aside from interacting via pull request comments, Atlantis could respond to a limited number of API endpoints.

## Main Endpoints

The API endpoints in this section are disabled by default, since these API endpoints could change the infrastructure directly.
To enable the API endpoints, `api-secret` should be configured.

:::tip Prerequisites

* Set `api-secret` as part of the [Server Configuration](server-configuration.html#api-secret)
* Pass `X-Atlantis-Token` with the same secret in the request header
  :::

### POST /api/plan

#### Description

Execute [atlantis plan](using-atlantis.html#atlantis-plan) on the specified repository.

#### Parameters

| Name       | Type                                | Required | Description                              |
|------------|-------------------------------------|----------|------------------------------------------|
| Repository | string                              | Yes      | Name of the Terraform repository         |
| Ref        | string                              | Yes      | Git reference, like a branch name        |
| Type       | string                              | Yes      | Type of the VCS provider (Github/Gitlab) |
| Paths      | [ [Path](api-endpoints.html#path) ] | Yes      | Paths to the projects to run the plan    |

##### Path

Similar to the [Options](using-atlantis.html#options) of `atlantis plan`. Path specifies which directory/workspace
within the repository to run the plan.  
At least one of `Directory` or `Workspace` should be specified.

| Name      | Type   | Required | Description                                                                                                                                               |
|-----------|--------|----------|-----------------------------------------------------------------------------------------------------------------------------------------------------------|
| Directory | string | No       | Which directory to run plan in relative to root of repo                                                                                                   |
| Workspace | string | No       | [Terraform workspace](https://developer.hashicorp.com/terraform/language/state/workspaces) of the plan. Use `default` if Terraform workspaces are unused. |

#### Sample Request

```shell
curl --request POST 'https://<ATLANTIS_HOST_NAME>/api/plan' \
--header 'X-Atlantis-Token: <ATLANTIS_API_SECRET>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "Repository": "repo-name",
    "Ref": "main",
    "Type": "Github",
    "Paths": [{
      "Directory": ".",
      "Workspace": "default"
    }]
}'
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

| Name       | Type                                  | Required | Description                              |
|------------|---------------------------------------|----------|------------------------------------------|
| Repository | string                                | Yes      | Name of the Terraform repository         |
| Ref        | string                                | Yes      | Git reference, like a branch name        |
| Type       | string                                | Yes      | Type of the VCS provider (Github/Gitlab) |
| Paths      | [ [Path](api-endpoints.html#path-1) ] | Yes      | Paths to the projects to run the apply   |

##### Path

Similar to the [Options](using-atlantis.html#options-1) of `atlantis apply`. Path specifies which directory/workspace
within the repository to run the apply.  
At least one of `Directory` or `Workspace` should be specified.

| Name      | Type   | Required | Description                                                                                                                                               |
|-----------|--------|----------|-----------------------------------------------------------------------------------------------------------------------------------------------------------|
| Directory | string | No       | Which directory to run apply in relative to root of repo                                                                                                  |
| Workspace | string | No       | [Terraform workspace](https://developer.hashicorp.com/terraform/language/state/workspaces) of the plan. Use `default` if Terraform workspaces are unused. |

#### Sample Request

```shell
curl --request POST 'https://<ATLANTIS_HOST_NAME>/api/apply' \
--header 'X-Atlantis-Token: <ATLANTIS_API_SECRET>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "Repository": "repo-name",
    "Ref": "main",
    "Type": "Github",
    "Paths": [{
      "Directory": ".",
      "Workspace": "default"
    }]
}'
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

## Other Endpoints

The endpoints listed in this section are non-destructive and therefore don't require authentication nor special secret token.

### GET /status

#### Description

Return the status of the Atlantis server.

#### Sample Request

```shell
curl --request GET 'https://<ATLANTIS_HOST_NAME>/status'
```

#### Sample Response

```json
{
  "shutting_down": false,
  "in_progress_operations": 0,
  "version": "0.22.3"
}
```

### GET /healthz

#### Description

Serves as the health-check endpoint for a containerized Atlantis server.

#### Sample Request

```shell
curl --request GET 'https://<ATLANTIS_HOST_NAME>/healthz'
```

#### Sample Response

```json
{
  "status": "ok"
}
```
