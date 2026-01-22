# API Endpoints

Aside from interacting via pull request comments, Atlantis could respond to a limited number of API endpoints.

:::warning ALPHA API - SUBJECT TO CHANGE
The API endpoints documented on this page are currently in **alpha state** and are **not considered stable**. The request and response schemas may change at any time without prior notice or deprecation period.

If you build integrations against these endpoints, when upgrading Atlantis you should review the release notes carefully and be prepared to update your code.
:::

## Main Endpoints

The API endpoints in this section are disabled by default, since these API endpoints could change the infrastructure directly.
To enable the API endpoints, `api-secret` should be configured.

:::tip Prerequisites

* Set `api-secret` as part of the [Server Configuration](server-configuration.md#api-secret)
* Pass `X-Atlantis-Token` with the same secret in the request header
  :::

### POST /api/plan

#### Description

Execute [atlantis plan](using-atlantis.md#atlantis-plan) on the specified repository.

#### Parameters

| Name       | Type     | Required | Description                              |
|------------|----------|----------|------------------------------------------|
| Repository | string   | Yes      | Name of the Terraform repository         |
| Ref        | string   | Yes      | Git reference, like a branch name        |
| Type       | string   | Yes      | Type of the VCS provider (Github/Gitlab) |
| Projects   | []string | No       | List of project names to run the plan    |
| Paths      | []Path   | No       | Paths to the projects to run the plan    |
| PR         | int      | No       | Pull Request number                      |

::: tip NOTE
At least one of `Projects` or `Paths` must be specified.
:::

::: tip Non-PR Workflows (Drift Detection)
When `PR` is omitted or set to `0`, Atlantis operates in non-PR mode. In this mode:
- PR-specific requirements (`approved`, `mergeable`) are automatically skipped
- Security requirements (`policies_passed`, `undiverged`) are still enforced
- This enables drift detection and other automation workflows without an associated pull request
:::

#### Path

Similar to the [Options](using-atlantis.md#options) of `atlantis plan`. Path specifies which directory/workspace
within the repository to run the plan.
At least one of `Directory` or `Workspace` should be specified.

| Name      | Type   | Required | Description                                                                                                                                               |
|-----------|--------|----------|-----------------------------------------------------------------------------------------------------------------------------------------------------------|
| Directory | string | No       | Which directory to run plan in relative to root of repo                                                                                                   |
| Workspace | string | No       | [Terraform workspace](https://developer.hashicorp.com/terraform/language/state/workspaces) of the plan. Use `default` if Terraform workspaces are unused. |

#### Sample Request (with PR)

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
    }],
    "PR": 2
}'
```

#### Sample Request (Drift Detection - No PR)

For drift detection workflows, omit the `PR` parameter:

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

#### Sample Response (Success)

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

#### Sample Response (Error)

When an error occurs, the `Error` field contains the error message as a string:

```json
{
  "Error": null,
  "Failure": "",
  "ProjectResults": [
    {
      "Command": 1,
      "RepoRelDir": "modules/vpc",
      "Workspace": "production",
      "Error": "terraform plan failed: resource not found",
      "Failure": "plan execution error",
      "PlanSuccess": null,
      "PolicyCheckSuccess": null,
      "ApplySuccess": "",
      "VersionSuccess": "",
      "ProjectName": "vpc"
    }
  ],
  "PlansDeleted": false
}
```

::: tip Error Handling
- `Error`: Contains the error message as a string when an error occurs, `null` on success
- `Failure`: Contains a human-readable failure message, empty string on success
- Both top-level and per-project `Error`/`Failure` fields follow this format
:::

### POST /api/apply

#### Description

Execute [atlantis apply](using-atlantis.md#atlantis-apply) on the specified repository.

#### Parameters

| Name       | Type     | Required | Description                              |
|------------|----------|----------|------------------------------------------|
| Repository | string   | Yes      | Name of the Terraform repository         |
| Ref        | string   | Yes      | Git reference, like a branch name        |
| Type       | string   | Yes      | Type of the VCS provider (Github/Gitlab) |
| Projects   | []string | No       | List of project names to run the apply   |
| Paths      | []Path   | No       | Paths to the projects to run the apply   |
| PR         | int      | No       | Pull Request number                      |

::: tip NOTE
At least one of `Projects` or `Paths` must be specified.
:::

::: tip Non-PR Workflows (Drift Detection)
When `PR` is omitted or set to `0`, Atlantis operates in non-PR mode. In this mode:
- PR-specific requirements (`approved`, `mergeable`) are automatically skipped
- Security requirements (`policies_passed`, `undiverged`) are still enforced
- This enables drift detection and other automation workflows without an associated pull request
:::

#### Path

Similar to the [Options](using-atlantis.md#options-1) of `atlantis apply`. Path specifies which directory/workspace
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
    }],
    "PR": 2
}'
```

#### Sample Response (Success)

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

::: tip Error Response Format
Error responses follow the same format as the plan endpoint. See the [plan error response example](#sample-response-error) for details.
:::

## Other Endpoints

The endpoints listed in this section are non-destructive and therefore don't require authentication nor special secret token.

### GET /api/locks

#### Description

List the currently held project locks.

#### Sample Request

```shell
curl --request GET 'https://<ATLANTIS_HOST_NAME>/api/locks'
```

#### Sample Response

```json
{
  "Locks": [
    {
      "Name": "lock-id",
      "ProjectName": "terraform",
      "ProjectRepo": "owner/repo",
      "ProjectRepoPath": "/path",
      "PullID": "123",
      "PullURL": "url",
      "User": "jdoe",
      "Workspace": "default",
      "Time": "2025-02-13T16:47:42.040856-08:00"
    }
  ]
}
```

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

### GET /debug/pprof

If `--enable-profiling-api` is set to true, it adds endpoints under this path to expose server's profiling data. See [profiling Go programs](https://go.dev/blog/pprof) for more information.
