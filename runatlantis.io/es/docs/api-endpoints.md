# API Endpoints

Aparte de interactuar mediante comentarios de pull request, Atlantis podría responder a un número limitado de API endpoints.

:::warning API ALPHA - SUJETA A CAMBIOS
Los API endpoints documentados en esta página están actualmente en **estado alpha** y **no se consideran estables**. Los esquemas de solicitud y respuesta pueden cambiar en cualquier momento sin aviso previo ni período de deprecación.

Si construyes integraciones contra estos endpoints, al actualizar Atlantis deberías revisar las notas de la versión cuidadosamente y estar preparado para actualizar tu código.
:::

## Formato de respuesta

Los API endpoints más nuevos de drift usan un formato de envoltura consistente:

```json
{
  "success": true,
  "data": { ... },
  "error": null,
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2025-01-21T10:30:00Z"
}
```

Los endpoints de command y lock actualmente devuelven sus cuerpos originales de nivel superior en lugar de la envoltura de la API de drift. Esto incluye `POST /api/plan`, `POST /api/apply` y los endpoints de lock existentes. Los endpoints de command devuelven `command.Result` en el nivel superior en caso de éxito o fallo de proyecto, y devuelven un cuerpo `{ "error": "..." }` de nivel superior para errores de request/auth/setup.

### Campos de respuesta de la envoltura

| Field      | Type    | Description                                                    |
|------------|---------|----------------------------------------------------------------|
| success    | boolean | `true` si la solicitud tuvo éxito, `false` en caso contrario             |
| data       | object  | La carga útil de la respuesta (presente en caso de éxito)      |
| error      | object  | Detalles del error (presente en caso de fallo, `null` en caso de éxito)          |
| request_id | string  | Identificador único para el trazado de solicitudes             |
| timestamp  | string  | Marca de tiempo ISO 8601 de cuándo se generó la respuesta      |

### Formato de respuesta de error de la envoltura

Cuando ocurre un error en un endpoint que usa la envoltura, la respuesta incluye información de error estructurada:

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

### Códigos de error

| Code                 | HTTP Status | Description                                      |
|----------------------|-------------|--------------------------------------------------|
| VALIDATION_ERROR     | 400         | Parámetros de solicitud o cuerpo no válidos      |
| UNAUTHORIZED         | 401         | Token de autenticación no válido o ausente       |
| FORBIDDEN            | 403         | Acceso denegado (p. ej., repositorio no permitido)     |
| NOT_FOUND            | 404         | Recurso solicitado no encontrado                  |
| INTERNAL_ERROR       | 500         | Error interno del servidor                        |
| SERVICE_UNAVAILABLE  | 503         | Funcionalidad no habilitada o servicio no disponible       |

## Endpoints principales

Los API endpoints de esta sección están deshabilitados de forma predeterminada, ya que estos API endpoints podrían cambiar la infraestructura directamente.
Para habilitar los API endpoints, se debe configurar `api-secret`.

:::tip Requisitos previos

* Establece `api-secret` como parte de la [Configuración del servidor](server-configuration.md#api-secret)
* Pasa `X-Atlantis-Token` con el mismo secreto en el encabezado de la solicitud
  :::

### POST /api/plan

#### Descripción

Ejecuta [atlantis plan](using-atlantis.md#atlantis-plan) en el repositorio especificado.

#### Parámetros

| Name       | Type     | Required | Description                              |
|------------|----------|----------|------------------------------------------|
| Repository | string   | Yes      | Nombre del repositorio de Terraform         |
| Ref        | string   | Yes      | Referencia Git, como un nombre de rama        |
| Type       | string   | Yes      | Tipo del proveedor VCS (Github/Gitlab) |
| Projects   | []string | No       | Lista de nombres de proyecto para ejecutar el plan    |
| Paths      | []Path   | No       | Rutas a los proyectos para ejecutar el plan    |
| PR         | int      | No       | Número de Pull Request                      |

::: tip NOTE
Se debe especificar al menos uno de `Projects` o `Paths`.
:::

::: tip Solicitudes API sin PR
Cuando `PR` se omite o se establece en `0`, Atlantis ejecuta la solicitud como un workflow sintético aislado sin PR. Los workflows sintéticos de API usan identidades de pull generadas para directorios de trabajo y locks, realizan checkout reforzado calificado por rama, omiten las búsquedas de archivos modificados del pull request, fallan de manera cerrada en caso de denegación de la allowlist de equipo y ordenan los proyectos seleccionados según el orden de ejecución configurado.

Para refs de tag, SHA de commit o refs no ambiguas que no son ramas, proporciona `base_branch` para que Atlantis pueda verificar que el ref obtenido es alcanzable desde la rama base prevista. Nombres de rama como `main` o `feature/foo` se obtienen como `refs/heads/<branch>`.
:::

::: tip Verificaciones de políticas
Cuando las verificaciones de políticas están habilitadas y la selección de proyectos genera contextos `policy_check`, las solicitudes API de plan/apply ejecutan verificaciones de políticas después de contextos plan exitosos. Las solicitudes API de apply con `apply_requirements: [policies_passed]` requieren un estado de política exitoso antes de aplicar.
:::

::: tip Fase de plan de API apply
El endpoint API apply ejecuta un plan antes de apply. Los errores a nivel de proyecto en esa fase previa a apply no omiten por sí mismos la fase apply para otros planes pendientes elegibles. El auto-apply de remediación de drift es más estricto y falla de manera cerrada cuando su plan previo a apply tiene errores.
:::

#### Path

Similar a las [Options](using-atlantis.md#options) de `atlantis plan`. Path especifica qué directorio/workspace
 dentro del repositorio ejecutar el plan.
Se debe especificar al menos uno de `Directory` o `Workspace`.

| Name      | Type   | Required | Description                                                                                                                                               |
|-----------|--------|----------|-----------------------------------------------------------------------------------------------------------------------------------------------------------|
| Directory | string | No       | En qué directorio ejecutar plan relativo a la raíz del repo                                                                                                   |
| Workspace | string | No       | [Terraform workspace](https://developer.hashicorp.com/terraform/language/state/workspaces) del plan. Usa `default` si no se usan Terraform workspaces. |

#### Solicitud de ejemplo (con PR)

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

#### Solicitud de ejemplo (detección de drift - sin PR)

Para workflows de detección de drift, omite el parámetro `PR`:

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

#### Respuesta de ejemplo (éxito)

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

#### Respuesta de ejemplo (error)

Cuando ocurre un error de request/auth/setup, el endpoint heredado devuelve un cuerpo de error de nivel superior:

```json
{
  "error": "request \"{}\" is missing fields"
}
```

#### Respuesta de ejemplo (error de proyecto)

Cuando ocurre un error a nivel de proyecto:

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

::: tip Valores de estado del proyecto

* `success`: El comando del proyecto se completó con éxito
* `error`: Ocurrió un error. Las respuestas heredadas de plan/apply preservan la forma histórica del JSON de error de Go: los valores `Error` de proyecto no nulos se codifican como `{}`, y los valores nil se codifican como `null`.
* `failed`: Ocurrió un fallo (revisa el campo `failure`)

:::

### POST /api/apply

#### Descripción

Ejecuta [atlantis apply](using-atlantis.md#atlantis-apply) en el repositorio especificado.

#### Parámetros

| Name       | Type     | Required | Description                              |
|------------|----------|----------|------------------------------------------|
| Repository | string   | Yes      | Nombre del repositorio de Terraform         |
| Ref        | string   | Yes      | Referencia Git, como un nombre de rama        |
| Type       | string   | Yes      | Tipo del proveedor VCS (Github/Gitlab) |
| Projects   | []string | No       | Lista de nombres de proyecto para ejecutar el apply   |
| Paths      | []Path   | No       | Rutas a los proyectos para ejecutar el apply   |
| PR         | int      | No       | Número de Pull Request                      |

::: tip NOTE
Se debe especificar al menos uno de `Projects` o `Paths`.
:::

#### Path

Similar a las [Options](using-atlantis.md#options-1) de `atlantis apply`. Path especifica qué directorio/workspace
 dentro del repositorio ejecutar el apply.
Se debe especificar al menos uno de `Directory` o `Workspace`.

| Name      | Type   | Required | Description                                                                                                                                               |
|-----------|--------|----------|-----------------------------------------------------------------------------------------------------------------------------------------------------------|
| Directory | string | No       | En qué directorio ejecutar apply relativo a la raíz del repo                                                                                                  |
| Workspace | string | No       | [Terraform workspace](https://developer.hashicorp.com/terraform/language/state/workspaces) del plan. Usa `default` si no se usan Terraform workspaces. |

#### Solicitud de ejemplo

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

#### Respuesta de ejemplo (éxito)

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

::: tip Formato de respuesta de error
Las respuestas de error siguen el mismo formato heredado que el endpoint plan. Consulta el [ejemplo de respuesta de error de plan](#sample-response-error) para más detalles.
:::

## Detección y remediación de drift (Alpha)

:::warning FUNCIONALIDAD ALPHA - SUJETA A CAMBIOS
Las APIs de detección de drift, estado de drift, remediación, historial de remediación y webhook de drift son funcionalidades alpha. Sus campos de solicitud, esquemas de respuesta, semántica de almacenamiento, payloads de webhook y controles de seguridad pueden cambiar antes de que la funcionalidad sea promovida a estable.

La detección de drift ejecuta workflows de Terraform plan y puede ejecutar hooks configurados o pasos plan personalizados. La remediación apply destructiva requiere tanto `--enable-drift-detection` como `--enable-drift-remediation`.
:::

### POST /api/drift/remediate

#### Descripción

Ejecuta remediación de drift en el repositorio especificado. Este endpoint te permite ejecutar operaciones de solo plan (para previsualizar la remediación) o auto-apply (para corregir automáticamente el drift) para proyectos con drift detectado.

::: tip Requisitos previos

* El almacenamiento de detección de drift debe estar habilitado en el servidor Atlantis
* El repositorio debe estar en la lista de repositorios permitidos (si está configurada)

:::

::: tip Orden del workflow
`POST /api/drift/remediate` con `action: "plan"` (el valor predeterminado) ejecuta un plan nuevo incluso sin datos de drift en caché — no requiere una llamada previa a `POST /api/drift/detect` — cuando se especifican `projects` o `paths`; una remediación no acotada (ninguno establecido) aún toma sus objetivos del drift en caché. Solo `action: "apply"` (o `drift_only: true`) requiere un registro de drift en caché para cada proyecto/ruta/workspace objetivo. Consulta el tip "Cached Drift Required" a continuación.
:::

#### Parámetros

| Name        | Type                 | Required    | Description                                                             |
|-------------|----------------------|-------------|-------------------------------------------------------------------------|
| repository  | string               | Yes         | Nombre completo del repositorio (p. ej., `owner/repo`)                               |
| ref         | string               | Yes         | Referencia Git (branch/tag/commit) a usar para la remediación                |
| base_branch | string               | Conditional | Contexto de rama para filtros de rama de repo-config y verificaciones de no divergencia     |
| type        | string               | Yes         | Tipo del proveedor VCS (`Github`/`Gitlab`/`Gitea`)                    |
| action      | string               | No          | Acción de remediación: `plan` (predeterminada) o `apply`                         |
| projects    | []string             | No          | Lista de nombres de proyecto a remediar. Si está vacía, usa datos de detección de drift |
| paths       | []DriftDetectionPath | No          | Lista de directorios/workspaces relativos al repo a remediar               |
| workspaces  | []string             | No          | Filtra la remediación a workspaces específicos                               |
| drift_only  | boolean              | No          | Si es true, remedia solo proyectos con drift detectado                    |

El campo `paths` usa el mismo objeto `DriftDetectionPath` descrito en `POST /api/drift/detect`.
Para la remediación, un selector de path sin `workspace` apunta solo al workspace predeterminado de Terraform.
Usa el campo de nivel superior `workspaces` o valores `workspace` a nivel de path para remediar workspaces no predeterminados.
Los selectores de proyecto para remediación son nombres exactos de proyecto. Los selectores de proyecto mediante expresiones regulares no son compatibles para remediación; usa nombres explícitos de proyecto o selectores de path cuando apuntes a múltiples proyectos.
Los selectores de path de la API son rutas literales normalizadas relativas al repo; patrones glob como `envs/*` no son compatibles.

::: tip Acciones

* `plan`: Ejecuta un plan para previsualizar qué cambiaría (predeterminado, no destructivo)
* `apply`: Ejecuta tanto plan como apply para corregir automáticamente el drift (destructivo). Esta acción requiere tanto `--enable-drift-detection` como `--enable-drift-remediation`, además de drift en caché con `has_drift: true` de una ejecución previa de detección para cada proyecto/path/workspace objetivo.

:::

::: warning Requisitos de apply
La remediación de drift apply no omite los `apply_requirements` del repositorio. Los requisitos que necesitan estado de pull request, como `approved` o `mergeable`, fallan de manera cerrada para solicitudes de remediación sin PR. Usa remediación de solo plan o workflows PR normales para proyectos protegidos por esos requisitos.
:::

::: tip Webhooks
La remediación de drift apply no activa los webhooks heredados `event: apply`. Usa webhooks de drift para notificaciones de workflow de drift.
:::

::: warning Requisitos de plan
Las acciones de solo plan de remediación de drift y la detección de drift no omiten `plan_requirements` de estado de PR. Requisitos como `approved` o `mergeable` no pueden satisfacerse sin un pull request y fallan de manera cerrada.
:::

::: tip Seguridad de ref
Cuando la remediación usa drift en caché para un ref móvil como `main`, Atlantis compara el commit del checkout actual con el commit que produjo el registro de drift en caché. Si el ref se ha movido, vuelve a ejecutar la detección de drift antes de usar `action: "apply"`.
:::

::: tip Se requiere drift en caché
La remediación `action: "apply"` solo aplica registros de drift en caché con `has_drift: true` para el mismo repositorio, ref, `base_branch`, proyecto/path y workspace. Usa `action: "plan"` para previsualizaciones sin caché, luego ejecuta la detección de drift antes de aplicar.
:::

::: tip Contexto de rama
Para refs de rama como `main`, `feature/foo` o `refs/heads/feature/foo`, Atlantis usa `ref` como el contexto de rama y obtiene explícitamente el espacio de nombres de la rama. Los refs bare ambiguos como `prod`, `latest`, `stable` o `v1.2.3` requieren `base_branch` pero aun así se obtienen como nombres de rama. Para SHAs de commit sin procesar y refs explícitos `refs/tags/...`, proporciona `base_branch` para que los filtros de rama de repo-config y las verificaciones de no divergencia se evalúen contra la rama prevista. Usa la forma explícita `refs/tags/...` para tags.
:::

#### Solicitud de ejemplo (solo plan)

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

#### Solicitud de ejemplo (auto-apply de proyectos específicos)

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

#### Solicitud de ejemplo (paths específicos)

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

#### Respuesta de ejemplo (éxito)

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

#### Respuesta de ejemplo (éxito de auto-apply)

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

#### Respuesta de ejemplo (fallo parcial)

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

#### Valores de estado

| Status    | Description                                                    |
|-----------|----------------------------------------------------------------|
| `pending` | La remediación está en cola pero todavía no ha comenzado                      |
| `running` | La remediación está actualmente en progreso                           |
| `success` | Todos los proyectos fueron remediados con éxito                      |
| `failed`  | Todos los proyectos fallaron en la remediación                                |
| `partial` | Algunos proyectos tuvieron éxito, algunos fallaron                           |

#### Respuestas de error

| Status Code | Description                                                                |
|-------------|----------------------------------------------------------------------------|
| 400         | Solicitud no válida (faltan campos requeridos o la acción no es válida)                |
| 401         | Encabezado `X-Atlantis-Token` no válido o ausente                               |
| 403         | El repositorio no está en la lista permitida                                             |
| 409         | La remediación se ejecutó pero todos los proyectos objetivo fallaron                           |
| 503         | La API, la remediación de drift o remediation apply no está habilitada en el servidor  |
| 500         | Error interno durante la remediación                                          |

### POST /api/drift/detect

#### Descripción

Activa la detección de drift para proyectos en un repositorio. Este endpoint inicia una operación plan para detectar drift de infraestructura sin requerir un pull request. Los resultados se almacenan para recuperación posterior mediante los endpoints de estado de drift.

Cuando los [webhooks de drift](sending-notifications-via-webhooks.md#drift-detection-webhooks) están configurados (`event: drift`), las ejecuciones de detección exitosas envían automáticamente notificaciones webhook a canales de Slack y/o endpoints HTTP, incluidos resultados de heartbeat sin drift.

::: tip Requisitos previos

* El almacenamiento de detección de drift debe estar habilitado en el servidor Atlantis (`--enable-drift-detection`)

:::

#### Parámetros

| Name                 | Type                 | Required    | Description                                                                          |
|----------------------|----------------------|-------------|--------------------------------------------------------------------------------------|
| repository           | string               | Yes         | Nombre completo del repositorio (p. ej., `owner/repo`)                                            |
| ref                  | string               | Yes         | Referencia Git (branch/tag/commit) para comprobar drift                                 |
| base_branch          | string               | Conditional | Contexto de rama para filtros de rama de repo-config y verificaciones de no divergencia                  |
| type                 | string               | Yes         | Tipo del proveedor VCS (`Github`/`Gitlab`/`Gitea`)                                 |
| projects             | []string             | No          | Lista de nombres de proyecto para comprobar. Si está vacía, se comprueban todos                            |
| paths                | []DriftDetectionPath | No          | Lista de paths para comprobar. Si está vacía, se usan nombres de proyecto                             |
| include_plan_output  | boolean              | No          | Si es true, incluye `plan_output` para cada proyecto en la respuesta. El valor predeterminado es `false` |

#### DriftDetectionPath

| Name      | Type   | Required | Description                                                     |
|-----------|--------|----------|-----------------------------------------------------------------|
| directory | string | Yes      | Path relativo al directorio de Terraform                        |
| workspace | string | No       | Workspace de Terraform. Si se omite, se usa el workspace predeterminado. |

Los selectores de path son rutas literales normalizadas relativas al repo. Los patrones glob como `envs/*` no son compatibles.

::: tip NOTE
Se debe especificar al menos uno de `projects` o `paths` para detección dirigida. Si ambos están vacíos, la detección de drift puede escanear todos los proyectos descubiertos. `projects` e `paths` son mutuamente excluyentes para la detección de drift; usa un tipo de selector por solicitud.
:::

::: tip Efectos secundarios de estado
La detección de drift suprime los estados de commit normales de Atlantis para plan, verificación de políticas, apply y hooks. Las notificaciones webhook específicas de drift aún pueden enviarse para ejecuciones de detección exitosas, incluidos resultados de heartbeat sin drift, cuando los webhooks de drift están configurados.

La detección de drift no ejecuta Terraform apply, pero sí ejecuta el ciclo de vida normal de plan. Los hooks pre-workflow configurados, workflows personalizados, pasos plan personalizados y comandos Terraform plan pueden ejecutarse del lado del servidor fuera del contexto de un pull request.

La detección de drift no omite las allowlists de equipo. Si una allowlist de equipo configurada no puede autorizar la solicitud API, la solicitud falla en lugar de escanear o reconciliar un conjunto de proyectos vacío. Los `plan_requirements` de estado de PR como `approved` o `mergeable` también fallan de manera cerrada para la detección de drift sin PR.
:::

::: tip Contexto de rama
Para refs de rama como `main`, `feature/foo` o `refs/heads/feature/foo`, Atlantis usa `ref` como el contexto de rama y obtiene explícitamente el espacio de nombres de la rama. Los refs bare ambiguos como `prod`, `latest`, `stable` o `v1.2.3` requieren `base_branch` pero aun así se obtienen como nombres de rama. Para SHAs de commit sin procesar y refs explícitos `refs/tags/...`, proporciona `base_branch` para que los filtros de rama de repo-config y las verificaciones de no divergencia se evalúen contra la rama prevista. Usa la forma explícita `refs/tags/...` para tags.
:::

#### Solicitud de ejemplo

```shell
curl --request POST 'https://<ATLANTIS_HOST_NAME>/api/drift/detect' \
--header 'X-Atlantis-Token: <ATLANTIS_API_SECRET>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "repository": "owner/repo",
    "ref": "main",
    "type": "Github",
    "projects": ["vpc", "ec2"],
    "include_plan_output": true
}'
```

#### Solicitud de ejemplo (con paths)

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
    ],
    "include_plan_output": true
}'
```

::: tip Salida del plan
Establece `include_plan_output: true` en la solicitud para que la respuesta incluya `plan_output` para cada proyecto — el texto del plan de Terraform para ese proyecto. Para el paso plan incorporado esto normalmente se normaliza para el renderizado de diff; el contenido exacto depende del workflow configurado, ya que un paso `run` personalizado puede producir en su lugar salida arbitraria, no normalizada. El valor predeterminado es `false`, ya que el texto del plan puede ser grande; cuando se omite o es `false`, `plan_output` no se incluye incluso para proyectos con un plan exitoso. También se omite cuando no hay salida de plan (por ejemplo, si el proyecto tuvo un error antes de que se ejecutara un plan). `plan_output` solo es devuelto por esta respuesta detect; nunca se incluye en `GET /api/drift/status`, ya que nunca se persiste en el almacenamiento de drift.
:::

::: warning La salida del plan puede contener datos sensibles
Antes de que este campo existiera, `POST /api/drift/detect` solo devolvía conteos numéricos de drift. Con `include_plan_output: true`, las respuestas pueden incluir valores de atributos de recursos y, para workflows de pasos `run` personalizados, salida arbitraria de comandos. El límite de autenticación del endpoint no cambia (el mismo token API que otros endpoints de drift/remediation), por lo que esto no es una nueva brecha de autorización, pero la sensibilidad de los datos de la respuesta cambia materialmente cuando este campo está habilitado. Solo la redacción filter-regex del paso `run` (si está configurada) se aplica a la salida del plan; por lo demás no se limpia.
:::

#### Respuesta de ejemplo (éxito)

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
        "plan_output": "Terraform will perform the following actions:\n  # aws_vpc.main will be updated in-place\n\nPlan: 1 to add, 2 to change, 0 to destroy.",
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

#### Respuestas de error

| Status Code | Error Code          | Description                                          |
|-------------|---------------------|------------------------------------------------------|
| 400         | VALIDATION_ERROR    | Solicitud no válida (faltan campos requeridos)            |
| 401         | UNAUTHORIZED        | Encabezado `X-Atlantis-Token` no válido o ausente         |
| 503         | SERVICE_UNAVAILABLE | El almacenamiento de detección de drift no está habilitado en el servidor |
| 500         | INTERNAL_ERROR      | Error interno durante la detección de drift                |

### GET /api/drift/remediate

#### Descripción

Lista resultados de remediación para un repositorio. Devuelve una lista paginada de operaciones de remediación pasadas. Este es un endpoint autenticado que requiere el secreto API.

::: tip Requisitos previos
La detección de drift debe estar habilitada en el servidor Atlantis. La remediación apply destructiva además requiere `--enable-drift-remediation`.
:::

#### Parámetros de consulta

| Name       | Type   | Required | Description                                                  |
|------------|--------|----------|--------------------------------------------------------------|
| repository | string | Yes      | Nombre completo del repositorio (p. ej., `owner/repo`)                    |
| type       | string | Yes      | Tipo de proveedor VCS (p. ej., `Github`, `Gitlab`, `Gitea`)        |
| limit      | int    | No       | Número máximo de resultados a devolver (predeterminado: 10, máx.: 100)  |

#### Solicitud de ejemplo

```shell
curl --request GET 'https://<ATLANTIS_HOST_NAME>/api/drift/remediate?repository=owner/repo&type=Github' \
--header 'X-Atlantis-Token: <ATLANTIS_API_SECRET>'
```

#### Solicitud de ejemplo (con limit)

```shell
curl --request GET 'https://<ATLANTIS_HOST_NAME>/api/drift/remediate?repository=owner/repo&type=Github&limit=10' \
--header 'X-Atlantis-Token: <ATLANTIS_API_SECRET>'
```

#### Respuesta de ejemplo

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

#### Respuesta de ejemplo (sin resultados)

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

#### Respuestas de error

| Status Code | Error Code          | Description                                             |
|-------------|---------------------|---------------------------------------------------------|
| 400         | VALIDATION_ERROR    | Falta el parámetro requerido `repository`                 |
| 401         | UNAUTHORIZED        | Encabezado `X-Atlantis-Token` no válido o ausente            |
| 503         | SERVICE_UNAVAILABLE | El almacenamiento de detección de drift no está habilitado en el servidor    |
| 500         | INTERNAL_ERROR      | Error interno al recuperar datos de remediación              |

### GET /api/drift/remediate/{id}

#### Descripción

Obtén un resultado de remediación específico por ID. Devuelve información detallada sobre una operación de remediación pasada, incluidos resultados por proyecto. Este es un endpoint autenticado que requiere el secreto API.

::: tip Requisitos previos
La detección de drift debe estar habilitada en el servidor Atlantis. La remediación apply destructiva además requiere `--enable-drift-remediation`.
:::

#### Parámetros de ruta

| Name | Type   | Required | Description                                |
|------|--------|----------|--------------------------------------------|
| id   | string | Yes      | El identificador único de la remediación   |

::: tip ¿Qué ID?
El `id` aquí es el campo `id` devuelto por una llamada previa a `POST /api/drift/remediate` — no el `detection_id`/`id` devuelto por `POST /api/drift/detect`. Las ejecuciones de detección y remediación se rastrean por separado, cada una con su propio espacio de IDs. Para inspeccionar la salida del plan de una remediación, llama primero a `POST /api/drift/remediate` y usa el `id` de su respuesta.
:::

#### Parámetros de consulta

| Name       | Type   | Required | Description                                                 |
|------------|--------|----------|-------------------------------------------------------------|
| repository | string | Yes      | Nombre completo del repositorio (p. ej., `owner/repo`)                   |
| type       | string | Yes      | Tipo del proveedor VCS (`Github`/`Gitlab`/`Gitea`)        |

#### Solicitud de ejemplo

```shell
curl --request GET 'https://<ATLANTIS_HOST_NAME>/api/drift/remediate/550e8400-e29b-41d4-a716-446655440000?repository=owner/repo&type=Github' \
--header 'X-Atlantis-Token: <ATLANTIS_API_SECRET>'
```

#### Respuesta de ejemplo

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

#### Respuestas de error

| Status Code | Error Code          | Description                                                  |
|-------------|---------------------|--------------------------------------------------------------|
| 400         | VALIDATION_ERROR    | Falta un parámetro requerido                                   |
| 401         | UNAUTHORIZED        | Encabezado `X-Atlantis-Token` no válido o ausente                 |
| 403         | FORBIDDEN           | El repositorio no está en la allowlist                           |
| 404         | NOT_FOUND           | Resultado de remediación no encontrado                                 |
| 503         | SERVICE_UNAVAILABLE | El almacenamiento de detección de drift no está habilitado en el servidor         |
| 500         | INTERNAL_ERROR      | Error interno al recuperar datos de remediación                   |

## Otros endpoints

La mayoría de los endpoints listados en esta sección no son destructivos y por lo tanto no requieren autenticación ni un token secreto especial. `GET /api/drift/status` es un endpoint autenticado de lectura de la API de drift y requiere `X-Atlantis-Token`.

### GET /api/locks

#### Descripción

Lista los locks de proyecto retenidos actualmente.

#### Solicitud de ejemplo

```shell
curl --request GET 'https://<ATLANTIS_HOST_NAME>/api/locks'
```

#### Respuesta de ejemplo

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

#### Respuesta de ejemplo (sin locks)

```json
{
  "Locks": []
}
```

### GET /api/drift/status

#### Descripción

Devuelve el estado de drift para un repositorio. Este endpoint proporciona resultados en caché de detección de drift de ejecuciones plan anteriores. La detección de drift debe estar habilitada en el servidor para que este endpoint funcione y requiere el token API configurado.

::: tip Requisitos previos
El almacenamiento de detección de drift debe estar habilitado en el servidor Atlantis. Si no está habilitado, este endpoint devuelve un error `503 Service Unavailable`.
:::

#### Parámetros de consulta

| Name        | Type   | Required | Description                                                   |
|-------------|--------|----------|---------------------------------------------------------------|
| repository  | string | Yes      | Nombre completo del repositorio (p. ej., `owner/repo`)                     |
| type        | string | Yes      | Tipo de proveedor VCS (p. ej., `Github`, `Gitlab`, `Gitea`)         |
| project     | string | No       | Filtrar por nombre de proyecto                                        |
| path        | string | No       | Filtrar por path literal normalizado de proyecto relativo al repositorio |
| workspace   | string | No       | Filtrar por Terraform workspace                                 |
| ref         | string | No       | Filtrar por referencia git                                       |
| base_branch | string | No       | Filtrar por el contexto de rama usado cuando se detectó drift         |

#### Solicitud de ejemplo

```shell
curl --request GET 'https://<ATLANTIS_HOST_NAME>/api/drift/status?repository=owner/repo&type=Github' \
  --header 'X-Atlantis-Token: <API_TOKEN>'
```

#### Solicitud de ejemplo (con filtros)

```shell
curl --request GET 'https://<ATLANTIS_HOST_NAME>/api/drift/status?repository=owner/repo&type=Github&project=vpc&path=modules/vpc&workspace=production&ref=main&base_branch=main' \
  --header 'X-Atlantis-Token: <API_TOKEN>'
```

#### Respuesta de ejemplo (con drift)

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

#### Respuesta de ejemplo (sin datos de drift)

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

#### Respuestas de error

| Status Code | Error Code          | Description                                  |
|-------------|---------------------|----------------------------------------------|
| 400         | VALIDATION_ERROR    | Falta el parámetro requerido `repository`      |
| 401         | UNAUTHORIZED        | Encabezado `X-Atlantis-Token` no válido o ausente |
| 403         | FORBIDDEN           | El repositorio no está en la allowlist           |
| 503         | SERVICE_UNAVAILABLE | La detección de drift no está habilitada en el servidor |
| 500         | INTERNAL_ERROR      | Error interno al recuperar datos de drift         |

### GET /status

#### Descripción

Devuelve el estado del servidor Atlantis.

#### Solicitud de ejemplo

```shell
curl --request GET 'https://<ATLANTIS_HOST_NAME>/status'
```

#### Respuesta de ejemplo

```json
{
  "shutting_down": false,
  "in_progress_operations": 0,
  "version": "0.22.3"
}
```

### GET /healthz

#### Descripción

Endpoint de liveness. Devuelve 200 si el proceso Atlantis está en ejecución. No comprueba dependencias externas. Adecuado para sondas de liveness de Kubernetes.

#### Solicitud de ejemplo

```shell
curl --request GET 'https://<ATLANTIS_HOST_NAME>/healthz'
```

#### Respuesta de ejemplo

```json
{
  "status": "ok"
}
```

### GET /readyz

#### Descripción

Endpoint de readiness. Devuelve 200 si el servidor está listo para manejar solicitudes, incluida la conectividad con dependencias externas (p. ej. Redis). Devuelve 503 si cualquier dependencia es inalcanzable. Adecuado para sondas de readiness de Kubernetes.

#### Solicitud de ejemplo

```shell
curl --request GET 'https://<ATLANTIS_HOST_NAME>/readyz'
```

#### Respuesta de ejemplo (saludable)

```json
{
  "status": "ok"
}
```

#### Respuesta de ejemplo (no saludable)

Devuelve HTTP 503:

```json
{
  "status": "error",
  "error": "failed to ping redis: ..."
}
```

### GET /debug/pprof

Si `--enable-profiling-api` se establece en true, agrega endpoints bajo esta ruta para exponer los datos de profiling del servidor. Consulta [profiling Go programs](https://go.dev/blog/pprof) para más información.
