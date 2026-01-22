# API Endpoints

Aside from interacting via pull request comments, Atlantis could respond to a limited number of API endpoints.

:::warning ALPHA API - SUBJECT TO CHANGE
The API endpoints documented on this page are currently in **alpha state** and are **not considered stable**. The request and response schemas may change at any time without prior notice or deprecation period.

If you build integrations against these endpoints, when upgrading Atlantis you should review the release notes carefully and be prepared to update your code.
:::

## Response Format

All API responses use a consistent envelope format:

```json
{
  "success": true,
  "data": { ... },
  "error": null,
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2025-01-21T10:30:00Z"
}
```

### Response Fields

| Field      | Type    | Description                                                    |
|------------|---------|----------------------------------------------------------------|
| success    | boolean | `true` if the request succeeded, `false` otherwise             |
| data       | object  | The response payload (present on success)                      |
| error      | object  | Error details (present on failure, `null` on success)          |
| request_id | string  | Unique identifier for request tracing                          |
| timestamp  | string  | ISO 8601 timestamp of when the response was generated          |

### Error Response Format

When an error occurs, the response includes structured error information:

```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "missing required parameter: repository"
  },
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2025-01-21T10:30:00Z"
}
```

### Error Codes

| Code                 | HTTP Status | Description                                      |
|----------------------|-------------|--------------------------------------------------|
| VALIDATION_ERROR     | 400         | Invalid request parameters or body               |
| UNAUTHORIZED         | 401         | Invalid or missing authentication token          |
| FORBIDDEN            | 403         | Access denied (e.g., repository not allowed)     |
| NOT_FOUND            | 404         | Requested resource not found                     |
| INTERNAL_ERROR       | 500         | Internal server error                            |
| SERVICE_UNAVAILABLE  | 503         | Feature not enabled or service unavailable       |

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

* PR-specific requirements (`approved`, `mergeable`) are automatically skipped
* Security requirements (`policies_passed`, `undiverged`) are still enforced
* This enables drift detection and other automation workflows without an associated pull request

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
  "success": true,
  "data": {
    "command": "plan",
    "projects": [
      {
        "project_name": "",
        "directory": ".",
        "workspace": "default",
        "command": "plan",
        "status": "success",
        "output": "<terraform plan output>",
        "plan": {
          "has_changes": true,
          "to_add": 2,
          "to_change": 1,
          "to_destroy": 0,
          "to_import": 0,
          "summary": "Plan: 2 to add, 1 to change, 0 to destroy."
        }
      }
    ],
    "total_projects": 1,
    "success_count": 1,
    "failure_count": 0,
    "executed_at": "2025-01-21T10:30:00Z"
  },
  "error": null,
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2025-01-21T10:30:00Z"
}
```

#### Sample Response (Error)

When an error occurs, the response includes structured error information:

```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "at least one of 'Projects' or 'Paths' must be specified"
  },
  "request_id": "550e8400-e29b-41d4-a716-446655440001",
  "timestamp": "2025-01-21T10:30:00Z"
}
```

#### Sample Response (Project Error)

When a project-level error occurs:

```json
{
  "success": true,
  "data": {
    "command": "plan",
    "projects": [
      {
        "project_name": "vpc",
        "directory": "modules/vpc",
        "workspace": "production",
        "command": "plan",
        "status": "error",
        "error": "terraform plan failed: resource not found"
      }
    ],
    "total_projects": 1,
    "success_count": 0,
    "failure_count": 1,
    "executed_at": "2025-01-21T10:30:00Z"
  },
  "error": null,
  "request_id": "550e8400-e29b-41d4-a716-446655440002",
  "timestamp": "2025-01-21T10:30:00Z"
}
```

::: tip Project Status Values

* `success`: Project command completed successfully
* `error`: An error occurred (check `error` field)
* `failed`: A failure occurred (check `failure` field)

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

* PR-specific requirements (`approved`, `mergeable`) are automatically skipped
* Security requirements (`policies_passed`, `undiverged`) are still enforced
* This enables drift detection and other automation workflows without an associated pull request

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
  "success": true,
  "data": {
    "command": "apply",
    "projects": [
      {
        "project_name": "",
        "directory": ".",
        "workspace": "default",
        "command": "apply",
        "status": "success",
        "output": "Apply complete! Resources: 2 added, 1 changed, 0 destroyed."
      }
    ],
    "total_projects": 1,
    "success_count": 1,
    "failure_count": 0,
    "executed_at": "2025-01-21T10:30:00Z"
  },
  "error": null,
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2025-01-21T10:30:00Z"
}
```

::: tip Error Response Format
Error responses follow the same envelope format as the plan endpoint. See the [plan error response example](#sample-response-error) for details.
:::

### POST /api/drift/remediate

#### Description

Execute drift remediation on the specified repository. This endpoint allows you to run plan-only (to preview remediation) or auto-apply (to automatically fix drift) operations for projects with detected drift.

::: tip Prerequisites

* Drift detection storage must be enabled on the Atlantis server
* The repository must be in the allowed repository list (if configured)

:::

#### Parameters

| Name       | Type     | Required | Description                                                              |
|------------|----------|----------|--------------------------------------------------------------------------|
| repository | string   | Yes      | Full repository name (e.g., `owner/repo`)                                |
| ref        | string   | Yes      | Git reference (branch/tag/commit) to use for remediation                 |
| type       | string   | Yes      | Type of the VCS provider (`Github`/`Gitlab`)                             |
| action     | string   | No       | Remediation action: `plan` (default) or `apply`                          |
| projects   | []string | No       | List of project names to remediate. If empty, uses drift detection data  |
| workspaces | []string | No       | Filter remediation to specific workspaces                                |
| drift_only | boolean  | No       | If true, only remediate projects with detected drift                     |

::: tip Actions

* `plan`: Runs a plan to preview what would change (default, non-destructive)
* `apply`: Runs both plan and apply to automatically fix drift (destructive)

:::

#### Sample Request (Plan Only)

```shell
curl --request POST 'https://<ATLANTIS_HOST_NAME>/api/drift/remediate' \
--header 'X-Atlantis-Token: <ATLANTIS_API_SECRET>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "repository": "owner/repo",
    "ref": "main",
    "type": "Github",
    "action": "plan",
    "drift_only": true
}'
```

#### Sample Request (Auto-Apply Specific Projects)

```shell
curl --request POST 'https://<ATLANTIS_HOST_NAME>/api/drift/remediate' \
--header 'X-Atlantis-Token: <ATLANTIS_API_SECRET>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "repository": "owner/repo",
    "ref": "main",
    "type": "Github",
    "action": "apply",
    "projects": ["vpc", "ec2"],
    "workspaces": ["production"]
}'
```

#### Sample Response (Success)

```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "repository": "owner/repo",
    "ref": "main",
    "action": "plan",
    "status": "success",
    "started_at": "2025-01-21T10:30:00Z",
    "completed_at": "2025-01-21T10:31:00Z",
    "projects": [
      {
        "project_name": "vpc",
        "directory": "modules/vpc",
        "workspace": "production",
        "status": "success",
        "plan_output": "Terraform will perform the following actions:\n  # aws_vpc.main will be updated...",
        "drift_before": {
          "to_add": 0,
          "to_change": 1,
          "to_destroy": 0,
          "to_import": 0,
          "total_changes": 1,
          "summary": "Plan: 0 to add, 1 to change, 0 to destroy.",
          "changes_outside": false
        },
        "drift_after": {
          "to_add": 0,
          "to_change": 1,
          "to_destroy": 0,
          "to_import": 0,
          "total_changes": 1,
          "summary": "Plan: 0 to add, 1 to change, 0 to destroy.",
          "changes_outside": false
        }
      },
      {
        "project_name": "ec2",
        "directory": "modules/ec2",
        "workspace": "production",
        "status": "success",
        "plan_output": "No changes. Infrastructure is up-to-date."
      }
    ],
    "summary": {
      "total_projects": 2,
      "success_count": 2,
      "failure_count": 0
    }
  },
  "error": null,
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2025-01-21T10:30:00Z"
}
```

#### Sample Response (Auto-Apply Success)

```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "repository": "owner/repo",
    "ref": "main",
    "action": "apply",
    "status": "success",
    "started_at": "2025-01-21T10:30:00Z",
    "completed_at": "2025-01-21T10:32:00Z",
    "projects": [
      {
        "project_name": "vpc",
        "directory": "modules/vpc",
        "workspace": "production",
        "status": "success",
        "plan_output": "Terraform will perform the following actions:\n  # aws_vpc.main will be updated...",
        "apply_output": "Apply complete! Resources: 0 added, 1 changed, 0 destroyed.",
        "drift_before": {
          "to_add": 0,
          "to_change": 1,
          "to_destroy": 0,
          "to_import": 0,
          "total_changes": 1,
          "summary": "Plan: 0 to add, 1 to change, 0 to destroy.",
          "changes_outside": false
        },
        "drift_after": {
          "to_add": 0,
          "to_change": 0,
          "to_destroy": 0,
          "to_import": 0,
          "total_changes": 0,
          "summary": "Apply completed successfully",
          "changes_outside": false
        }
      }
    ],
    "summary": {
      "total_projects": 1,
      "success_count": 1,
      "failure_count": 0
    }
  },
  "error": null,
  "request_id": "550e8400-e29b-41d4-a716-446655440001",
  "timestamp": "2025-01-21T10:32:00Z"
}
```

#### Sample Response (Partial Failure)

```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440002",
    "repository": "owner/repo",
    "ref": "main",
    "action": "plan",
    "status": "partial",
    "started_at": "2025-01-21T10:30:00Z",
    "completed_at": "2025-01-21T10:31:00Z",
    "projects": [
      {
        "project_name": "vpc",
        "directory": "modules/vpc",
        "workspace": "production",
        "status": "success",
        "plan_output": "No changes. Infrastructure is up-to-date."
      },
      {
        "project_name": "ec2",
        "directory": "modules/ec2",
        "workspace": "production",
        "status": "failed",
        "error": "terraform plan failed: Error acquiring state lock"
      }
    ],
    "summary": {
      "total_projects": 2,
      "success_count": 1,
      "failure_count": 1
    }
  },
  "error": null,
  "request_id": "550e8400-e29b-41d4-a716-446655440002",
  "timestamp": "2025-01-21T10:31:00Z"
}
```

#### Status Values

| Status    | Description                                                    |
|-----------|----------------------------------------------------------------|
| `pending` | Remediation is queued but not yet started                      |
| `running` | Remediation is currently in progress                           |
| `success` | All projects were remediated successfully                      |
| `failed`  | All projects failed remediation                                |
| `partial` | Some projects succeeded, some failed                           |

#### Error Responses

| Status Code | Description                                                                |
|-------------|----------------------------------------------------------------------------|
| 400         | Invalid request (missing required fields, invalid action, or API disabled) |
| 401         | Invalid or missing `X-Atlantis-Token` header                               |
| 403         | Repository not in allowed list                                             |
| 503         | Drift remediation is not enabled on the server                             |
| 500         | Internal error during remediation                                          |

### POST /api/drift/detect

#### Description

Trigger drift detection for projects in a repository. This endpoint initiates a plan operation to detect infrastructure drift without requiring a pull request. Results are stored for later retrieval via the drift status endpoints.

::: tip Prerequisites

* Drift detection storage must be enabled on the Atlantis server

:::

#### Parameters

| Name       | Type                 | Required | Description                                                |
|------------|----------------------|----------|------------------------------------------------------------|
| repository | string               | Yes      | Full repository name (e.g., `owner/repo`)                  |
| ref        | string               | Yes      | Git reference (branch/tag/commit) to check for drift       |
| type       | string               | Yes      | Type of the VCS provider (`Github`/`Gitlab`)               |
| projects   | []string             | No       | List of project names to check. If empty, all are checked  |
| paths      | []DriftDetectionPath | No       | List of paths to check. If empty, project names are used   |

#### DriftDetectionPath

| Name      | Type   | Required | Description                                                     |
|-----------|--------|----------|-----------------------------------------------------------------|
| directory | string | No       | Relative path to the Terraform directory                        |
| workspace | string | No       | Terraform workspace (optional)                                  |

::: tip NOTE
At least one of `projects` or `paths` should be specified for targeted detection. If both are empty, drift detection may scan all discovered projects.
:::

#### Sample Request

```shell
curl --request POST 'https://<ATLANTIS_HOST_NAME>/api/drift/detect' \
--header 'X-Atlantis-Token: <ATLANTIS_API_SECRET>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "repository": "owner/repo",
    "ref": "main",
    "type": "Github",
    "projects": ["vpc", "ec2"]
}'
```

#### Sample Request (with paths)

```shell
curl --request POST 'https://<ATLANTIS_HOST_NAME>/api/drift/detect' \
--header 'X-Atlantis-Token: <ATLANTIS_API_SECRET>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "repository": "owner/repo",
    "ref": "main",
    "type": "Github",
    "paths": [
        {"directory": "modules/vpc", "workspace": "production"},
        {"directory": "modules/ec2", "workspace": "production"}
    ]
}'
```

#### Sample Response (Success)

```json
{
  "success": true,
  "data": {
    "repository": "owner/repo",
    "projects": [
      {
        "project_name": "vpc",
        "directory": "modules/vpc",
        "workspace": "production",
        "ref": "main",
        "has_drift": true,
        "drift": {
          "to_add": 1,
          "to_change": 2,
          "to_destroy": 0,
          "to_import": 0,
          "total_changes": 3,
          "summary": "Plan: 1 to add, 2 to change, 0 to destroy.",
          "changes_outside": false
        },
        "last_checked": "2025-01-21T10:30:00Z"
      },
      {
        "project_name": "ec2",
        "directory": "modules/ec2",
        "workspace": "production",
        "ref": "main",
        "has_drift": false,
        "last_checked": "2025-01-21T10:30:00Z"
      }
    ],
    "detected_at": "2025-01-21T10:30:00Z",
    "summary": {
      "total_projects": 2,
      "projects_with_drift": 1,
      "projects_without_drift": 1,
      "projects_with_errors": 0
    }
  },
  "error": null,
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2025-01-21T10:30:00Z"
}
```

#### Error Responses

| Status Code | Error Code          | Description                                          |
|-------------|---------------------|------------------------------------------------------|
| 400         | VALIDATION_ERROR    | Invalid request (missing required fields)            |
| 401         | UNAUTHORIZED        | Invalid or missing `X-Atlantis-Token` header         |
| 503         | SERVICE_UNAVAILABLE | Drift detection storage is not enabled on the server |
| 500         | INTERNAL_ERROR      | Internal error during drift detection                |

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
  "success": true,
  "data": {
    "locks": [
      {
        "id": "owner/repo/./default",
        "project_name": "terraform",
        "repository": "owner/repo",
        "path": ".",
        "workspace": "default",
        "pull_request_id": 123,
        "pull_request_url": "https://github.com/owner/repo/pull/123",
        "locked_by": "jdoe",
        "locked_at": "2025-02-13T16:47:42.040856-08:00"
      }
    ],
    "total_count": 1
  },
  "error": null,
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2025-01-21T10:30:00Z"
}
```

#### Sample Response (No Locks)

```json
{
  "success": true,
  "data": {
    "locks": [],
    "total_count": 0
  },
  "error": null,
  "request_id": "550e8400-e29b-41d4-a716-446655440001",
  "timestamp": "2025-01-21T10:30:00Z"
}
```

### GET /api/drift/status

#### Description

Return the drift status for a repository. This endpoint provides cached drift detection results from previous plan executions. Drift detection must be enabled on the server for this endpoint to work.

::: tip Prerequisites
Drift detection storage must be enabled on the Atlantis server. If not enabled, this endpoint returns a `503 Service Unavailable` error.
:::

#### Query Parameters

| Name       | Type   | Required | Description                                                  |
|------------|--------|----------|--------------------------------------------------------------|
| repository | string | Yes      | Full repository name (e.g., `owner/repo`)                    |
| project    | string | No       | Filter by project name                                       |
| workspace  | string | No       | Filter by Terraform workspace                                |

#### Sample Request

```shell
curl --request GET 'https://<ATLANTIS_HOST_NAME>/api/drift/status?repository=owner/repo'
```

#### Sample Request (with filters)

```shell
curl --request GET 'https://<ATLANTIS_HOST_NAME>/api/drift/status?repository=owner/repo&project=vpc&workspace=production'
```

#### Sample Response (with drift)

```json
{
  "success": true,
  "data": {
    "repository": "owner/repo",
    "projects": [
      {
        "project_name": "vpc",
        "directory": "modules/vpc",
        "workspace": "production",
        "ref": "main",
        "has_drift": true,
        "drift": {
          "to_add": 2,
          "to_change": 1,
          "to_destroy": 0,
          "to_import": 0,
          "total_changes": 3,
          "summary": "Plan: 2 to add, 1 to change, 0 to destroy.",
          "changes_outside": false
        },
        "last_checked": "2025-01-21T10:30:00Z"
      },
      {
        "project_name": "ec2",
        "directory": "modules/ec2",
        "workspace": "production",
        "ref": "main",
        "has_drift": false,
        "last_checked": "2025-01-21T10:25:00Z"
      }
    ],
    "checked_at": "2025-01-21T10:30:00Z",
    "summary": {
      "total_projects": 2,
      "projects_with_drift": 1,
      "projects_without_drift": 1,
      "projects_with_errors": 0
    }
  },
  "error": null,
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2025-01-21T10:30:00Z"
}
```

#### Sample Response (no drift data)

```json
{
  "success": true,
  "data": {
    "repository": "owner/repo",
    "projects": [],
    "checked_at": "2025-01-21T10:30:00Z",
    "summary": {
      "total_projects": 0,
      "projects_with_drift": 0,
      "projects_without_drift": 0,
      "projects_with_errors": 0
    }
  },
  "error": null,
  "request_id": "550e8400-e29b-41d4-a716-446655440001",
  "timestamp": "2025-01-21T10:30:00Z"
}
```

#### Error Responses

| Status Code | Error Code          | Description                                  |
|-------------|---------------------|----------------------------------------------|
| 400         | VALIDATION_ERROR    | Missing required `repository` parameter      |
| 503         | SERVICE_UNAVAILABLE | Drift detection is not enabled on the server |
| 500         | INTERNAL_ERROR      | Internal error retrieving drift data         |

### GET /api/drift/remediate

#### Description

List remediation results for a repository. Returns a paginated list of past remediation operations. This endpoint does not require authentication.

::: tip Prerequisites
Drift remediation must be enabled on the Atlantis server. If not enabled, this endpoint returns a `503 Service Unavailable` error.
:::

#### Query Parameters

| Name       | Type   | Required | Description                                                  |
|------------|--------|----------|--------------------------------------------------------------|
| repository | string | Yes      | Full repository name (e.g., `owner/repo`)                    |
| limit      | int    | No       | Maximum number of results to return (default: 50)            |

#### Sample Request

```shell
curl --request GET 'https://<ATLANTIS_HOST_NAME>/api/drift/remediate?repository=owner/repo'
```

#### Sample Request (with limit)

```shell
curl --request GET 'https://<ATLANTIS_HOST_NAME>/api/drift/remediate?repository=owner/repo&limit=10'
```

#### Sample Response

```json
{
  "success": true,
  "data": {
    "repository": "owner/repo",
    "count": 2,
    "results": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "repository": "owner/repo",
        "ref": "main",
        "action": "plan",
        "status": "success",
        "started_at": "2025-01-21T10:30:00Z",
        "completed_at": "2025-01-21T10:31:00Z",
        "projects": [],
        "summary": {
          "total_projects": 2,
          "success_count": 2,
          "failure_count": 0
        }
      },
      {
        "id": "550e8400-e29b-41d4-a716-446655440001",
        "repository": "owner/repo",
        "ref": "main",
        "action": "apply",
        "status": "partial",
        "started_at": "2025-01-21T09:00:00Z",
        "completed_at": "2025-01-21T09:05:00Z",
        "projects": [],
        "summary": {
          "total_projects": 3,
          "success_count": 2,
          "failure_count": 1
        }
      }
    ]
  },
  "error": null,
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2025-01-21T10:30:00Z"
}
```

#### Sample Response (no results)

```json
{
  "success": true,
  "data": {
    "repository": "owner/repo",
    "count": 0,
    "results": []
  },
  "error": null,
  "request_id": "550e8400-e29b-41d4-a716-446655440001",
  "timestamp": "2025-01-21T10:30:00Z"
}
```

#### Error Responses

| Status Code | Error Code          | Description                                    |
|-------------|---------------------|------------------------------------------------|
| 400         | VALIDATION_ERROR    | Missing required `repository` parameter        |
| 503         | SERVICE_UNAVAILABLE | Drift remediation is not enabled on the server |
| 500         | INTERNAL_ERROR      | Internal error retrieving remediation data     |

### GET /api/drift/remediate/{id}

#### Description

Get a specific remediation result by ID. Returns detailed information about a past remediation operation including per-project results. This endpoint does not require authentication.

::: tip Prerequisites
Drift remediation must be enabled on the Atlantis server. If not enabled, this endpoint returns a `503 Service Unavailable` error.
:::

#### Path Parameters

| Name | Type   | Required | Description                                |
|------|--------|----------|--------------------------------------------|
| id   | string | Yes      | The unique identifier of the remediation   |

#### Sample Request

```shell
curl --request GET 'https://<ATLANTIS_HOST_NAME>/api/drift/remediate/550e8400-e29b-41d4-a716-446655440000'
```

#### Sample Response

```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "repository": "owner/repo",
    "ref": "main",
    "action": "plan",
    "status": "success",
    "started_at": "2025-01-21T10:30:00Z",
    "completed_at": "2025-01-21T10:31:00Z",
    "projects": [
      {
        "project_name": "vpc",
        "directory": "modules/vpc",
        "workspace": "production",
        "status": "success",
        "plan_output": "Terraform will perform the following actions:\n  # aws_vpc.main will be updated...",
        "drift_before": {
          "to_add": 0,
          "to_change": 1,
          "to_destroy": 0,
          "to_import": 0,
          "total_changes": 1,
          "summary": "Plan: 0 to add, 1 to change, 0 to destroy.",
          "changes_outside": false
        }
      },
      {
        "project_name": "ec2",
        "directory": "modules/ec2",
        "workspace": "production",
        "status": "success",
        "plan_output": "No changes. Infrastructure is up-to-date."
      }
    ],
    "summary": {
      "total_projects": 2,
      "success_count": 2,
      "failure_count": 0
    }
  },
  "error": null,
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2025-01-21T10:31:00Z"
}
```

#### Error Responses

| Status Code | Error Code          | Description                                           |
|-------------|---------------------|-------------------------------------------------------|
| 400         | VALIDATION_ERROR    | Missing required `id` parameter                       |
| 404         | NOT_FOUND           | Remediation result not found                          |
| 503         | SERVICE_UNAVAILABLE | Drift remediation is not enabled on the server        |
| 500         | INTERNAL_ERROR      | Internal error retrieving remediation data            |

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
