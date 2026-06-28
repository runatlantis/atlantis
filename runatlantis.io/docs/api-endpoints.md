# API Endpoints

Aside from interacting via pull request comments, Atlantis could respond to a limited number of API endpoints.

:::warning ALPHA API - SUBJECT TO CHANGE
The API endpoints documented on this page are currently in **alpha state** and are **not considered stable**. The request and response schemas may change at any time without prior notice or deprecation period.

If you build integrations against these endpoints, when upgrading Atlantis you should review the release notes carefully and be prepared to update your code.
:::

## Response Format

The newer drift API endpoints use a consistent envelope format:

```json
{
  "success": true,
  "data": { ... },
  "error": null,
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2025-01-21T10:30:00Z"
}
```

The command and lock endpoints currently return their original top-level bodies rather than the drift API envelope. This includes `POST /api/plan`, `POST /api/apply`, and the existing lock endpoints. Command endpoints return `command.Result` at the top level on success or project failure, and return a top-level `{ "error": "..." }` body for request/auth/setup errors.

### Envelope Response Fields

| Field      | Type    | Description                                                    |
|------------|---------|----------------------------------------------------------------|
| success    | boolean | `true` if the request succeeded, `false` otherwise             |
| data       | object  | The response payload (present on success)                      |
| error      | object  | Error details (present on failure, `null` on success)          |
| request_id | string  | Unique identifier for request tracing                          |
| timestamp  | string  | ISO 8601 timestamp of when the response was generated          |

### Envelope Error Response Format

When an error occurs on an endpoint that uses the envelope, the response includes structured error information:

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

::: tip No-PR API Requests
When `PR` is omitted or set to `0`, Atlantis runs the request as an isolated synthetic non-PR workflow. Synthetic API workflows use generated pull identities for working directories and locks, perform hardened branch-qualified checkout, skip pull-request modified-file lookups, fail closed on team allowlist denial, and sort selected projects by configured execution order.

For tag, commit SHA, or ambiguous non-branch refs, provide `base_branch` so Atlantis can verify the checked ref is reachable from the intended base branch. Branch names such as `main` or `feature/foo` are fetched as `refs/heads/<branch>`.
:::

::: tip Policy Checks
When policy checks are enabled and project selection generates `policy_check` contexts, API plan/apply requests run policy checks after successful plan contexts. API apply requests with `apply_requirements: [policies_passed]` require successful policy status before applying.
:::

::: tip API Apply Plan Phase
The API apply endpoint runs a plan before apply. Project-level errors in that pre-apply phase do not by themselves skip the apply phase for other eligible pending plans. Drift remediation auto-apply is stricter and fails closed when its pre-apply plan has errors.
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
      "Error": null,
      "Failure": "",
      "PlanSuccess": {
        "TerraformOutput": "<terraform plan output>"
      },
      "RepoRelDir": ".",
      "Workspace": "default",
      "ProjectName": ""
    }
  ],
  "PlansDeleted": false
}
```

#### Sample Response (Error)

When a request/auth/setup error occurs, the legacy endpoint returns a top-level error body:

```json
{
  "error": "request \"{}\" is missing fields"
}
```

#### Sample Response (Project Error)

When a project-level error occurs:

```json
{
  "Error": null,
  "Failure": "",
  "ProjectResults": [
    {
      "Error": {},
      "Failure": "",
      "RepoRelDir": "modules/vpc",
      "Workspace": "production",
      "ProjectName": "vpc"
    }
  ],
  "PlansDeleted": false
}
```

::: tip Project Status Values

* `success`: Project command completed successfully
* `error`: An error occurred. Legacy plan/apply responses preserve the historical Go error JSON shape: non-nil project `Error` values are encoded as `{}`, and nil values are encoded as `null`.
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
      "Error": null,
      "Failure": "",
      "ApplySuccess": "Apply complete! Resources: 2 added, 1 changed, 0 destroyed.",
      "RepoRelDir": ".",
      "Workspace": "default",
      "ProjectName": ""
    }
  ],
  "PlansDeleted": false
}
```

::: tip Error Response Format
Error responses follow the same legacy format as the plan endpoint. See the [plan error response example](#sample-response-error) for details.
:::

## Drift Detection And Remediation (Alpha)

:::warning ALPHA FEATURE - SUBJECT TO CHANGE
The drift detection, drift status, remediation, remediation history, and drift webhook APIs are alpha features. Their request fields, response schemas, storage semantics, webhook payloads, and safety gates may change before the feature is promoted to stable.

Drift detection runs Terraform plan workflows and can execute configured hooks or custom plan steps. Destructive remediation apply requires both `--enable-drift-detection` and `--enable-drift-remediation`.
:::

### POST /api/drift/remediate

#### Description

Execute drift remediation on the specified repository. This endpoint allows you to run plan-only (to preview remediation) or auto-apply (to automatically fix drift) operations for projects with detected drift.

::: tip Prerequisites

* Drift detection storage must be enabled on the Atlantis server
* The repository must be in the allowed repository list (if configured)

:::

#### Parameters

| Name        | Type                 | Required    | Description                                                             |
|-------------|----------------------|-------------|-------------------------------------------------------------------------|
| repository  | string               | Yes         | Full repository name (e.g., `owner/repo`)                               |
| ref         | string               | Yes         | Git reference (branch/tag/commit) to use for remediation                |
| base_branch | string               | Conditional | Branch context for repo-config branch filters and undiverged checks     |
| type        | string               | Yes         | Type of the VCS provider (`Github`/`Gitlab`/`Gitea`)                    |
| action      | string               | No          | Remediation action: `plan` (default) or `apply`                         |
| projects    | []string             | No          | List of project names to remediate. If empty, uses drift detection data |
| paths       | []DriftDetectionPath | No          | List of repo-relative directories/workspaces to remediate               |
| workspaces  | []string             | No          | Filter remediation to specific workspaces                               |
| drift_only  | boolean              | No          | If true, only remediate projects with detected drift                    |

The `paths` field uses the same `DriftDetectionPath` object described under `POST /api/drift/detect`.
For remediation, a path selector without `workspace` targets the default Terraform workspace only.
Use the top-level `workspaces` field or path-level `workspace` values to remediate non-default workspaces.
Project selectors for remediation are exact project names. Regular expression project selectors are not supported for remediation; use explicit project names or path selectors when targeting multiple projects.
API path selectors are literal normalized repo-relative paths; glob patterns such as `envs/*` are not supported.

::: tip Actions

* `plan`: Runs a plan to preview what would change (default, non-destructive)
* `apply`: Runs both plan and apply to automatically fix drift (destructive). This action requires both `--enable-drift-detection` and `--enable-drift-remediation`, plus cached drift with `has_drift: true` from a previous detection run for each targeted project/path/workspace.

:::

::: warning Apply Requirements
Drift remediation apply does not bypass repository `apply_requirements`. Requirements that need pull request state, such as `approved` or `mergeable`, fail closed for non-PR remediation requests. Use plan-only remediation or normal PR workflows for projects guarded by those requirements.
:::

::: tip Webhooks
Drift remediation apply does not trigger legacy `event: apply` webhooks. Use drift webhooks for drift workflow notifications.
:::

::: warning Plan Requirements
Drift remediation plan-only actions and drift detection do not bypass PR-state `plan_requirements`. Requirements such as `approved` or `mergeable` cannot be satisfied without a pull request and fail closed.
:::

::: tip Ref Safety
When remediation uses cached drift for a moving ref such as `main`, Atlantis compares the current checkout commit with the commit that produced the cached drift record. If the ref has moved, rerun drift detection before using `action: "apply"`.
:::

::: tip Cached Drift Required
Remediation `action: "apply"` only applies cached drift records with `has_drift: true` for the same repository, ref, `base_branch`, project/path, and workspace. Use `action: "plan"` for uncached previews, then run drift detection before applying.
:::

::: tip Branch Context
For branch refs such as `main`, `feature/foo`, or `refs/heads/feature/foo`, Atlantis uses `ref` as the branch context and fetches the branch namespace explicitly. Ambiguous bare refs such as `prod`, `latest`, `stable`, or `v1.2.3` require `base_branch` but are still fetched as branch names. For raw commit SHAs and explicit `refs/tags/...` refs, provide `base_branch` so repo-config branch filters and undiverged checks are evaluated against the intended branch. Use the explicit `refs/tags/...` form for tags.
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
    "workspaces": ["production"],
    "drift_only": true
}'
```

#### Sample Request (Specific Paths)

```shell
curl --request POST 'https://<ATLANTIS_HOST_NAME>/api/drift/remediate' \
--header 'X-Atlantis-Token: <ATLANTIS_API_SECRET>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "repository": "owner/repo",
    "ref": "main",
    "type": "Github",
    "action": "plan",
    "paths": [
        {"directory": "modules/vpc", "workspace": "production"}
    ]
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
          "to_forget": 0,
          "total_changes": 1,
          "summary": "Plan: 0 to add, 1 to change, 0 to destroy.",
          "changes_outside": false
        },
        "drift_after": {
          "to_add": 0,
          "to_change": 1,
          "to_destroy": 0,
          "to_import": 0,
          "to_forget": 0,
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
          "to_forget": 0,
          "total_changes": 1,
          "summary": "Plan: 0 to add, 1 to change, 0 to destroy.",
          "changes_outside": false
        },
        "drift_after": {
          "to_add": 0,
          "to_change": 0,
          "to_destroy": 0,
          "to_import": 0,
          "to_forget": 0,
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
| 400         | Invalid request (missing required fields or invalid action)                |
| 401         | Invalid or missing `X-Atlantis-Token` header                               |
| 403         | Repository not in allowed list                                             |
| 409         | Remediation ran but all targeted projects failed                           |
| 503         | API, drift remediation, or remediation apply is not enabled on the server  |
| 500         | Internal error during remediation                                          |

### POST /api/drift/detect

#### Description

Trigger drift detection for projects in a repository. This endpoint initiates a plan operation to detect infrastructure drift without requiring a pull request. Results are stored for later retrieval via the drift status endpoints.

When [drift webhooks](sending-notifications-via-webhooks.md#drift-detection-webhooks) are configured (`event: drift`), successful detection runs send webhook notifications automatically to Slack channels and/or HTTP endpoints, including no-drift heartbeat results.

::: tip Prerequisites

* Drift detection storage must be enabled on the Atlantis server (`--enable-drift-detection`)

:::

#### Parameters

| Name        | Type                 | Required    | Description                                                         |
|-------------|----------------------|-------------|---------------------------------------------------------------------|
| repository  | string               | Yes         | Full repository name (e.g., `owner/repo`)                           |
| ref         | string               | Yes         | Git reference (branch/tag/commit) to check for drift                |
| base_branch | string               | Conditional | Branch context for repo-config branch filters and undiverged checks |
| type        | string               | Yes         | Type of the VCS provider (`Github`/`Gitlab`/`Gitea`)                |
| projects    | []string             | No          | List of project names to check. If empty, all are checked           |
| paths       | []DriftDetectionPath | No          | List of paths to check. If empty, project names are used            |

#### DriftDetectionPath

| Name      | Type   | Required | Description                                                     |
|-----------|--------|----------|-----------------------------------------------------------------|
| directory | string | Yes      | Relative path to the Terraform directory                        |
| workspace | string | No       | Terraform workspace. If omitted, the default workspace is used. |

Path selectors are literal normalized repo-relative paths. Glob patterns such as `envs/*` are not supported.

::: tip NOTE
At least one of `projects` or `paths` should be specified for targeted detection. If both are empty, drift detection may scan all discovered projects. `projects` and `paths` are mutually exclusive for drift detection; use one selector type per request.
:::

::: tip Status Side Effects
Drift detection suppresses normal Atlantis plan, policy check, apply, and hook commit statuses. Drift-specific webhook notifications can still be sent for successful detection runs, including no-drift heartbeat results, when drift webhooks are configured.

Drift detection does not run Terraform apply, but it does execute the normal plan lifecycle. Configured pre-workflow hooks, custom workflows, custom plan steps, and Terraform plan commands can run server-side outside a pull request context.

Drift detection does not bypass team allowlists. If a configured team allowlist cannot authorize the API request, the request fails instead of scanning or reconciling an empty project set. PR-state `plan_requirements` such as `approved` or `mergeable` also fail closed for non-PR drift detection.
:::

::: tip Branch Context
For branch refs such as `main`, `feature/foo`, or `refs/heads/feature/foo`, Atlantis uses `ref` as the branch context and fetches the branch namespace explicitly. Ambiguous bare refs such as `prod`, `latest`, `stable`, or `v1.2.3` require `base_branch` but are still fetched as branch names. For raw commit SHAs and explicit `refs/tags/...` refs, provide `base_branch` so repo-config branch filters and undiverged checks are evaluated against the intended branch. Use the explicit `refs/tags/...` form for tags.
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
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "repository": "owner/repo",
    "projects": [
      {
        "project_name": "vpc",
        "directory": "modules/vpc",
        "workspace": "production",
        "ref": "main",
        "detection_id": "550e8400-e29b-41d4-a716-446655440000",
        "has_drift": true,
        "drift": {
          "to_add": 1,
          "to_change": 2,
          "to_destroy": 0,
          "to_import": 0,
          "to_forget": 0,
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

### GET /api/drift/remediate

#### Description

List remediation results for a repository. Returns a paginated list of past remediation operations. This is an authenticated endpoint that requires the API secret.

::: tip Prerequisites
Drift detection must be enabled on the Atlantis server. Destructive remediation apply additionally requires `--enable-drift-remediation`.
:::

#### Query Parameters

| Name       | Type   | Required | Description                                                  |
|------------|--------|----------|--------------------------------------------------------------|
| repository | string | Yes      | Full repository name (e.g., `owner/repo`)                    |
| type       | string | Yes      | VCS provider type (e.g., `Github`, `Gitlab`, `Gitea`)        |
| limit      | int    | No       | Maximum number of results to return (default: 10, max: 100)  |

#### Sample Request

```shell
curl --request GET 'https://<ATLANTIS_HOST_NAME>/api/drift/remediate?repository=owner/repo&type=Github' \
--header 'X-Atlantis-Token: <ATLANTIS_API_SECRET>'
```

#### Sample Request (with limit)

```shell
curl --request GET 'https://<ATLANTIS_HOST_NAME>/api/drift/remediate?repository=owner/repo&type=Github&limit=10' \
--header 'X-Atlantis-Token: <ATLANTIS_API_SECRET>'
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

| Status Code | Error Code          | Description                                             |
|-------------|---------------------|---------------------------------------------------------|
| 400         | VALIDATION_ERROR    | Missing required `repository` parameter                 |
| 401         | UNAUTHORIZED        | Invalid or missing `X-Atlantis-Token` header            |
| 503         | SERVICE_UNAVAILABLE | Drift detection storage is not enabled on the server    |
| 500         | INTERNAL_ERROR      | Internal error retrieving remediation data              |

### GET /api/drift/remediate/{id}

#### Description

Get a specific remediation result by ID. Returns detailed information about a past remediation operation including per-project results. This is an authenticated endpoint that requires the API secret.

::: tip Prerequisites
Drift detection must be enabled on the Atlantis server. Destructive remediation apply additionally requires `--enable-drift-remediation`.
:::

#### Path Parameters

| Name | Type   | Required | Description                                |
|------|--------|----------|--------------------------------------------|
| id   | string | Yes      | The unique identifier of the remediation   |

#### Query Parameters

| Name       | Type   | Required | Description                                                 |
|------------|--------|----------|-------------------------------------------------------------|
| repository | string | Yes      | Full repository name (e.g., `owner/repo`)                   |
| type       | string | Yes      | Type of the VCS provider (`Github`/`Gitlab`/`Gitea`)        |

#### Sample Request

```shell
curl --request GET 'https://<ATLANTIS_HOST_NAME>/api/drift/remediate/550e8400-e29b-41d4-a716-446655440000?repository=owner/repo&type=Github' \
--header 'X-Atlantis-Token: <ATLANTIS_API_SECRET>'
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
          "to_forget": 0,
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

| Status Code | Error Code          | Description                                                  |
|-------------|---------------------|--------------------------------------------------------------|
| 400         | VALIDATION_ERROR    | Missing required parameter                                   |
| 401         | UNAUTHORIZED        | Invalid or missing `X-Atlantis-Token` header                 |
| 403         | FORBIDDEN           | Repository is not in the allowlist                           |
| 404         | NOT_FOUND           | Remediation result not found                                 |
| 503         | SERVICE_UNAVAILABLE | Drift detection storage is not enabled on the server         |
| 500         | INTERNAL_ERROR      | Internal error retrieving remediation data                   |

## Other Endpoints

Most endpoints listed in this section are non-destructive and therefore don't require authentication nor a special secret token. `GET /api/drift/status` is an authenticated drift API read endpoint and requires `X-Atlantis-Token`.

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
      "Name": "owner/repo/./default/terraform",
      "ProjectName": "terraform",
      "ProjectRepo": "owner/repo",
      "ProjectRepoPath": ".",
      "PullID": "123",
      "PullURL": "https://github.com/owner/repo/pull/123",
      "User": "jdoe",
      "Workspace": "default",
      "Time": "2025-02-13T16:47:42.040856-08:00"
    }
  ]
}
```

#### Sample Response (No Locks)

```json
{
  "Locks": []
}
```

### GET /api/drift/status

#### Description

Return the drift status for a repository. This endpoint provides cached drift detection results from previous plan executions. Drift detection must be enabled on the server for this endpoint to work and requires the configured API token.

::: tip Prerequisites
Drift detection storage must be enabled on the Atlantis server. If not enabled, this endpoint returns a `503 Service Unavailable` error.
:::

#### Query Parameters

| Name        | Type   | Required | Description                                                   |
|-------------|--------|----------|---------------------------------------------------------------|
| repository  | string | Yes      | Full repository name (e.g., `owner/repo`)                     |
| type        | string | Yes      | VCS provider type (e.g., `Github`, `Gitlab`, `Gitea`)         |
| project     | string | No       | Filter by project name                                        |
| path        | string | No       | Filter by literal normalized repository-relative project path |
| workspace   | string | No       | Filter by Terraform workspace                                 |
| ref         | string | No       | Filter by git reference                                       |
| base_branch | string | No       | Filter by branch context used when drift was detected         |

#### Sample Request

```shell
curl --request GET 'https://<ATLANTIS_HOST_NAME>/api/drift/status?repository=owner/repo&type=Github' \
  --header 'X-Atlantis-Token: <API_TOKEN>'
```

#### Sample Request (with filters)

```shell
curl --request GET 'https://<ATLANTIS_HOST_NAME>/api/drift/status?repository=owner/repo&type=Github&project=vpc&path=modules/vpc&workspace=production&ref=main&base_branch=main' \
  --header 'X-Atlantis-Token: <API_TOKEN>'
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
          "to_forget": 0,
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
| 401         | UNAUTHORIZED        | Invalid or missing `X-Atlantis-Token` header |
| 403         | FORBIDDEN           | Repository is not in the allowlist           |
| 503         | SERVICE_UNAVAILABLE | Drift detection is not enabled on the server |
| 500         | INTERNAL_ERROR      | Internal error retrieving drift data         |

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

Liveness endpoint. Returns 200 if the Atlantis process is running. Does not check external dependencies. Suitable for Kubernetes liveness probes.

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

### GET /readyz

#### Description

Readiness endpoint. Returns 200 if the server is ready to handle requests, including connectivity to external dependencies (e.g. Redis). Returns 503 if any dependency is unreachable. Suitable for Kubernetes readiness probes.

#### Sample Request

```shell
curl --request GET 'https://<ATLANTIS_HOST_NAME>/readyz'
```

#### Sample Response (healthy)

```json
{
  "status": "ok"
}
```

#### Sample Response (unhealthy)

Returns HTTP 503:

```json
{
  "status": "error",
  "error": "failed to ping redis: ..."
}
```

### GET /debug/pprof

If `--enable-profiling-api` is set to true, it adds endpoints under this path to expose server's profiling data. See [profiling Go programs](https://go.dev/blog/pprof) for more information.
