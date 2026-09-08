# Endpoints de API

Además de interactuar mediante comentarios de pull request, Atlantis puede responder a un número limitado de endpoints de API.

:::warning API ALPHA - SUJETA A CAMBIOS
Los endpoints de API documentados en esta página están actualmente en **estado alpha** y **no se consideran estables**. Los esquemas de solicitud y respuesta pueden cambiar en cualquier momento sin aviso previo ni período de deprecación.

Si construyes integraciones contra estos endpoints, al actualizar Atlantis debes revisar cuidadosamente las notas de la versión y estar preparado para actualizar tu código.
:::

## Formato de Respuesta

Los endpoints más nuevos de la API de drift usan un formato de envoltura consistente:

```json
{
  "success": true,
  "data": { ... },
  "error": null,
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2025-01-21T10:30:00Z"
}
```

Los endpoints de command y lock actualmente devuelven sus cuerpos originales de nivel superior en lugar de la envoltura de la API de drift. Esto incluye `POST /api/plan`, `POST /api/apply` y los endpoints de lock existentes. Los endpoints de command devuelven `command.Result` en el nivel superior en caso de éxito o fallo del proyecto, y devuelven un cuerpo `{ "error": "..." }` de nivel superior para errores de solicitud/auth/setup.

### Campos de la Respuesta de Envoltura

| Field      | Type    | Description                                                    |
|------------|---------|----------------------------------------------------------------|
| success    | boolean | `true` si la solicitud tuvo éxito, `false` en caso contrario             |
| data       | object  | La carga útil de la respuesta (presente en caso de éxito)      |
| error      | object  | Detalles del error (presente en caso de fallo, `null` en caso de éxito)          |
| request_id | string  | Identificador único para el rastreo de la solicitud            |
| timestamp  | string  | Marca de tiempo ISO 8601 de cuándo se generó la respuesta      |

### Formato de Respuesta de Error de Envoltura

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

### Códigos de Error

| Code                 | HTTP Status | Description                                      |
|----------------------|-------------|--------------------------------------------------|
| VALIDATION_ERROR     | 400         | Parámetros o cuerpo de solicitud no válidos      |
| UNAUTHORIZED         | 401         | Token de autenticación no válido o ausente       |
| FORBIDDEN            | 403         | Acceso denegado (p. ej., repositorio no permitido) |
| NOT_FOUND            | 404         | Recurso solicitado no encontrado                 |
| INTERNAL_ERROR       | 500         | Error interno del servidor                       |
| SERVICE_UNAVAILABLE  | 503         | Funcionalidad no habilitada o servicio no disponible       |

## Endpoints Principales

Los endpoints de API en esta sección están deshabilitados por defecto, ya que estos endpoints de API podrían cambiar la infraestructura directamente.
Para habilitar los endpoints de API, se debe configurar `api-secret`.

:::tip Requisitos previos

* Establece `api-secret` como parte de la [Configuración del Servidor](server-configuration.md#api-secret)
* Pasa `X-Atlantis-Token` con el mismo secreto en el encabezado de la solicitud
  :::

### POST /api/plan

#### Descripción

Ejecuta [atlantis plan](using-atlantis.md#atlantis-plan) en el repositorio especificado.

#### Parámetros

| Name       | Type     | Required | Description                              |
|------------|----------|----------|------------------------------------------|
| Repository | string   | Yes      | Nombre del repositorio de Terraform      |
| Ref        | string   | Yes      | Referencia de Git, como un nombre de rama |
| Type       | string   | Yes      | Tipo del proveedor de VCS (Github/Gitlab) |
| Projects   | []string | No       | Lista de nombres de proyectos para ejecutar el plan    |
| Paths      | []Path   | No       | Rutas a los proyectos para ejecutar el plan    |
| PR         | int      | No       | Número de Pull Request                      |

::: tip NOTE
Se debe especificar al menos uno de `Projects` o `Paths`.
:::

::: tip Solicitudes de API sin PR
Cuando `PR` se omite o se establece en `0`, Atlantis ejecuta la solicitud como un workflow sintético aislado sin PR. Los workflows sintéticos de API usan identidades de pull generadas para directorios de trabajo y locks, realizan un checkout reforzado calificado por rama, omiten las búsquedas de archivos modificados del pull request, fallan de forma cerrada en caso de denegación de allowlist de equipos, y ordenan los proyectos seleccionados según el orden de ejecución configurado.

Para refs de tag, SHA de commit o refs no de rama ambiguas, proporciona `base_branch` para que Atlantis pueda verificar que la ref checkouted sea alcanzable desde la rama base prevista. Los nombres de rama como `main` o `feature/foo` se obtienen como `refs/heads/<branch>`.
:::

::: tip Verificaciones de Políticas
Cuando las verificaciones de políticas están habilitadas y la selección de proyectos genera contextos `policy_check`, las solicitudes API plan/apply ejecutan verificaciones de políticas después de contextos de plan exitosos. Las solicitudes API apply con `apply_requirements: [policies_passed]` requieren un estado de política exitoso antes de aplicar.
:::

::: tip Fase de Plan de API Apply
El endpoint API apply ejecuta un plan antes de apply. Los errores a nivel de proyecto en esa fase previa a apply no omiten por sí mismos la fase apply para otros planes pendientes elegibles. El auto-apply de remediación de drift es más estricto y falla de forma cerrada cuando su plan previo a apply tiene errores.
:::

#### Path

Similar a las [Options](using-atlantis.md#options) de `atlantis plan`. Path especifica qué directorio/workspace
dentro del repositorio ejecutar el plan.
Se debe especificar al menos uno de `Directory` o `Workspace`.

| Name      | Type   | Required | Description                                                                                                                                               |
|-----------|--------|----------|-----------------------------------------------------------------------------------------------------------------------------------------------------------|
| Directory | string | No       | En qué directorio ejecutar plan relativo a la raíz del repo                                                                                                   |
| Workspace | string | No       | [Terraform workspace](https://developer.hashicorp.com/terraform/language/state/workspaces) del plan. Usa `default` si no se usan Terraform workspaces. |

#### Solicitud de Ejemplo (con PR)

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

#### Solicitud de Ejemplo (Detección de Drift - Sin PR)

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

#### Respuesta de Ejemplo (Éxito)

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

#### Respuesta de Ejemplo (Error)

Cuando ocurre un error de solicitud/auth/setup, el endpoint heredado devuelve un cuerpo de error de nivel superior:

```json
{
  "error": "request \"{}\" is missing fields"
}
```

#### Respuesta de Ejemplo (Error de Proyecto)

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

::: tip Valores de Estado del Proyecto

* `success`: El command del proyecto se completó exitosamente
* `error`: Ocurrió un error. Las respuestas heredadas de plan/apply preservan la forma JSON histórica del error de Go: los valores `Error` del proyecto no nil se codifican como `{}`, y los valores nil se codifican como `null`.
* `failed`: Ocurrió un fallo (revisa el campo `failure`)

:::

### POST /api/apply

#### Descripción

Ejecuta [atlantis apply](using-atlantis.md#atlantis-apply) en el repositorio especificado.

#### Parámetros

| Name       | Type     | Required | Description                              |
|------------|----------|----------|------------------------------------------|
| Repository | string   | Yes      | Nombre del repositorio de Terraform      |
| Ref        | string   | Yes      | Referencia de Git, como un nombre de rama |
| Type       | string   | Yes      | Tipo del proveedor de VCS (Github/Gitlab) |
| Projects   | []string | No       | Lista de nombres de proyectos para ejecutar el apply   |
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

#### Solicitud de Ejemplo

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

#### Respuesta de Ejemplo (Éxito)

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

::: tip Formato de Respuesta de Error
Las respuestas de error siguen el mismo formato heredado que el endpoint plan. Consulta el [ejemplo de respuesta de error de plan](#sample-response-error) para más detalles.
:::

## Detección y Remediación de Drift (Alpha)

:::warning FUNCIONALIDAD ALPHA - SUJETA A CAMBIOS
Las API de detección de drift, estado de drift, remediación, historial de remediación y webhook de drift son funcionalidades alpha. Sus campos de solicitud, esquemas de respuesta, semántica de almacenamiento, payloads de webhook y compuertas de seguridad pueden cambiar antes de que la funcionalidad sea promovida a estable.

La detección de drift ejecuta workflows de Terraform plan y puede ejecutar hooks configurados o pasos plan personalizados. El apply destructivo de remediación requiere tanto `--enable-drift-detection` como `--enable-drift-remediation`.
:::

### POST /api/drift/remediate

#### Descripción

Ejecuta remediación de drift en el repositorio especificado. Este endpoint te permite ejecutar operaciones de solo plan (para previsualizar la remediación) o auto-apply (para corregir automáticamente el drift) para proyectos con drift detectado.

::: tip Requisitos previos

* El almacenamiento de detección de drift debe estar habilitado en el servidor Atlantis
* El repositorio debe estar en la lista de repositorios permitidos (si está configurada)

:::

::: tip Orden del Workflow
`POST /api/drift/remediate` con `action: "plan"` (el valor por defecto) ejecuta un plan nuevo incluso sin datos de drift en caché — no requiere una llamada previa a `POST /api/drift/detect` — cuando se especifican `projects` o `paths`; una remediación no acotada (sin ninguno de los dos establecidos) aún toma sus objetivos del drift en caché. Solo `action: "apply"` (o `drift_only: true`) requiere un registro de drift en caché para cada proyecto/ruta/workspace objetivo. Consulta el consejo "Cached Drift Required" a continuación.
:::

#### Parámetros

| Name        | Type                 | Required    | Description                                                             |
|-------------|----------------------|-------------|-------------------------------------------------------------------------|
| repository  | string               | Yes         | Nombre completo del repositorio (p. ej., `owner/repo`)                               |
| ref         | string               | Yes         | Referencia de Git (rama/tag/commit) para usar en la remediación                |
| base_branch | string               | Conditional | Contexto de rama para filtros de rama de repo-config y verificaciones de no divergencia     |
| type        | string               | Yes         | Tipo del proveedor de VCS (`Github`/`Gitlab`/`Gitea`)                    |
| action      | string               | No          | Acción de remediación: `plan` (por defecto) o `apply`                         |
| projects    | []string             | No          | Lista de nombres de proyectos para remediar. Si está vacía, usa datos de detección de drift |
| paths       | []DriftDetectionPath | No          | Lista de directorios/workspaces relativos al repo para remediar               |
| workspaces  | []string             | No          | Filtra la remediación a workspaces específicos                               |
| drift_only  | boolean              | No          | Si es true, solo remedia proyectos con drift detectado                    |

El campo `paths` usa el mismo objeto `DriftDetectionPath` descrito en `POST /api/drift/detect`.
Para remediación, un selector de ruta sin `workspace` apunta solo al Terraform workspace por defecto.
Usa el campo de nivel superior `workspaces` o los valores `workspace` a nivel de ruta para remediar workspaces no predeterminados.
Los selectores de proyecto para remediación son nombres exactos de proyecto. Los selectores de proyecto con expresiones regulares no son compatibles para remediación; usa nombres de proyecto explícitos o selectores de ruta cuando apuntes a múltiples proyectos.
Los selectores de ruta de API son rutas literales normalizadas relativas al repo; los patrones glob como `envs/*` no son compatibles.

::: tip Acciones

* `plan`: Ejecuta un plan para previsualizar qué cambiaría (por defecto, no destructivo)
* `apply`: Ejecuta tanto plan como apply para corregir automáticamente el drift (destructivo). Esta acción requiere tanto `--enable-drift-detection` como `--enable-drift-remediation`, además de drift en caché con `has_drift: true` de una ejecución de detección previa para cada proyecto/ruta/workspace objetivo.

:::

::: warning Requisitos de Apply
El apply de remediación de drift no omite los `apply_requirements` del repositorio. Los requisitos que necesitan estado de pull request, como `approved` o `mergeable`, fallan de forma cerrada para solicitudes de remediación sin PR. Usa remediación de solo plan o workflows normales de PR para proyectos protegidos por esos requisitos.
:::

::: tip Webhooks
El apply de remediación de drift no activa los webhooks heredados `event: apply`. Usa webhooks de drift para notificaciones del workflow de drift.
:::

::: warning Requisitos de Plan
Las acciones de solo plan de remediación de drift y la detección de drift no omiten `plan_requirements` de estado de PR. Los requisitos como `approved` o `mergeable` no pueden satisfacerse sin un pull request y fallan de forma cerrada.
:::

::: tip Seguridad de Ref
Cuando la remediación usa drift en caché para una ref móvil como `main`, Atlantis compara el commit del checkout actual con el commit que produjo el registro de drift en caché. Si la ref se ha movido, vuelve a ejecutar la detección de drift antes de usar `action: "apply"`.
:::

::: tip Drift en Caché Requerido
La remediación `action: "apply"` solo aplica registros de drift en caché con `has_drift: true` para el mismo repositorio, ref, `base_branch`, proyecto/ruta y workspace. Usa `action: "plan"` para previsualizaciones sin caché, luego ejecuta la detección de drift antes de aplicar.
:::

::: tip Contexto de Rama
Para refs de rama como `main`, `feature/foo` o `refs/heads/feature/foo`, Atlantis usa `ref` como contexto de rama y obtiene explícitamente el espacio de nombres de la rama. Las refs ambiguas sin prefijo como `prod`, `latest`, `stable` o `v1.2.3` requieren `base_branch` pero aun así se obtienen como nombres de rama. Para SHAs de commit sin procesar y refs `refs/tags/...` explícitas, proporciona `base_branch` para que los filtros de rama de repo-config y las verificaciones de no divergencia se evalúen contra la rama prevista. Usa la forma explícita `refs/tags/...` para tags.
:::

#### Solicitud de Ejemplo (Solo Plan)

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

#### Solicitud de Ejemplo (Auto-Apply de Proyectos Específicos)

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

#### Solicitud de Ejemplo (Rutas Específicas)

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

#### Respuesta de Ejemplo (Éxito)

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

#### Respuesta de Ejemplo (Éxito de Auto-Apply)

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

#### Respuesta de Ejemplo (Fallo Parcial)

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

#### Valores de Estado

| Status    | Description                                                    |
|-----------|----------------------------------------------------------------|
| `pending` | La remediación está en cola pero aún no ha comenzado                      |
| `running` | La remediación está actualmente en progreso                           |
| `success` | Todos los proyectos fueron remediados exitosamente                      |
| `failed`  | Todos los proyectos fallaron la remediación                                |
| `partial` | Algunos proyectos tuvieron éxito, algunos fallaron                           |

#### Respuestas de Error

| Status Code | Description                                                                |
|-------------|----------------------------------------------------------------------------|
| 400         | Solicitud no válida (faltan campos requeridos o la acción no es válida)                |
| 401         | Encabezado `X-Atlantis-Token` no válido o ausente                               |
| 403         | Repositorio no está en la lista permitida                                             |
| 409         | La remediación se ejecutó pero todos los proyectos objetivo fallaron                           |
| 503         | API, remediación de drift o remediation apply no está habilitado en el servidor  |
| 500         | Error interno durante la remediación                                          |

### POST /api/drift/detect

#### Descripción

Activa la detección de drift para proyectos en un repositorio. Este endpoint inicia una operación plan para detectar drift de infraestructura sin requerir un pull request. Los resultados se almacenan para su recuperación posterior mediante los endpoints de estado de drift.

Cuando los [webhooks de drift](sending-notifications-via-webhooks.md#drift-detection-webhooks) están configurados (`event: drift`), las ejecuciones de detección exitosas envían notificaciones de webhook automáticamente a canales de Slack y/o endpoints HTTP, incluidos resultados heartbeat sin drift.

::: tip Requisitos previos

* El almacenamiento de detección de drift debe estar habilitado en el servidor Atlantis (`--enable-drift-detection`)

:::

#### Parámetros

| Name                 | Type                 | Required    | Description                                                                          |
|----------------------|----------------------|-------------|--------------------------------------------------------------------------------------|
| repository           | string               | Yes         | Nombre completo del repositorio (p. ej., `owner/repo`)                                            |
| ref                  | string               | Yes         | Referencia de Git (rama/tag/commit) para verificar drift                                 |
| base_branch          | string               | Conditional | Contexto de rama para filtros de rama de repo-config y verificaciones de no divergencia                  |
| type                 | string               | Yes         | Tipo del proveedor de VCS (`Github`/`Gitlab`/`Gitea`)                                 |
| projects             | []string             | No          | Lista de nombres de proyectos a verificar. Si está vacía, se verifican todos                            |
| paths                | []DriftDetectionPath | No          | Lista de rutas a verificar. Si está vacía, se usan nombres de proyecto                             |
| include_plan_output  | boolean              | No          | Si es true, incluye `plan_output` para cada proyecto en la respuesta. Por defecto es `false` |

#### DriftDetectionPath

| Name      | Type   | Required | Description                                                     |
|-----------|--------|----------|-----------------------------------------------------------------|
| directory | string | Yes      | Ruta relativa al directorio de Terraform                        |
| workspace | string | No       | Terraform workspace. Si se omite, se usa el workspace por defecto. |

Los selectores de ruta son rutas literales normalizadas relativas al repo. Los patrones glob como `envs/*` no son compatibles.

::: tip NOTE
Se debe especificar al menos uno de `projects` o `paths` para una detección dirigida. Si ambos están vacíos, la detección de drift puede escanear todos los proyectos descubiertos. `projects` y `paths` son mutuamente excluyentes para detección de drift; usa un tipo de selector por solicitud.
:::

::: tip Efectos Secundarios de Estado
La detección de drift suprime los estados normales de commit de Atlantis para plan, verificación de políticas, apply y hooks. Las notificaciones de webhook específicas de drift aún pueden enviarse para ejecuciones de detección exitosas, incluidos resultados heartbeat sin drift, cuando los webhooks de drift están configurados.

La detección de drift no ejecuta Terraform apply, pero sí ejecuta el ciclo de vida normal de plan. Los hooks pre-workflow configurados, workflows personalizados, pasos plan personalizados y comandos Terraform plan pueden ejecutarse del lado del servidor fuera del contexto de un pull request.

La detección de drift no omite las allowlists de equipos. Si una allowlist de equipos configurada no puede autorizar la solicitud de API, la solicitud falla en lugar de escanear o reconciliar un conjunto vacío de proyectos. `plan_requirements` de estado de PR como `approved` o `mergeable` también fallan de forma cerrada para la detección de drift sin PR.
:::

::: tip Contexto de Rama
Para refs de rama como `main`, `feature/foo` o `refs/heads/feature/foo`, Atlantis usa `ref` como contexto de rama y obtiene explícitamente el espacio de nombres de la rama. Las refs ambiguas sin prefijo como `prod`, `latest`, `stable` o `v1.2.3` requieren `base_branch` pero aun así se obtienen como nombres de rama. Para SHAs de commit sin procesar y refs `refs/tags/...` explícitas, proporciona `base_branch` para que los filtros de rama de repo-config y las verificaciones de no divergencia se evalúen contra la rama prevista. Usa la forma explícita `refs/tags/...` para tags.
:::

#### Solicitud de Ejemplo

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

#### Solicitud de Ejemplo (con rutas)

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

::: tip Salida de Plan
Establece `include_plan_output: true` en la solicitud para que la respuesta incluya `plan_output` para cada proyecto — el texto del Terraform plan para ese proyecto. Para el paso plan incorporado esto normalmente se normaliza para renderizado diff; el contenido exacto depende del workflow configurado, ya que un paso `run` personalizado puede producir salida arbitraria y no normalizada. Por defecto es `false`, ya que el texto del plan puede ser grande; cuando se omite o es `false`, `plan_output` no se incluye incluso para proyectos con un plan exitoso. También se omite cuando no hay salida de plan (por ejemplo, si el proyecto tuvo un error antes de que se ejecutara un plan). `plan_output` solo se devuelve mediante esta respuesta detect; nunca se incluye en `GET /api/drift/status`, ya que nunca se persiste en el almacenamiento de drift.
:::

::: warning La Salida de Plan Puede Contener Datos Sensibles
Antes de que existiera este campo, `POST /api/drift/detect` solo devolvía conteos numéricos de drift. Con `include_plan_output: true`, las respuestas pueden incluir valores de atributos de recursos y, para workflows de pasos `run` personalizados, salida de comandos arbitraria. El límite de autenticación del endpoint no cambia (el mismo token de API que otros endpoints de drift/remediación), así que esto no es una nueva brecha de autorización, pero la sensibilidad de los datos de la respuesta cambia materialmente cuando este campo está habilitado. Solo la redacción filter-regex del paso `run` (si está configurada) se aplica a la salida del plan; por lo demás no se depura.
:::

#### Respuesta de Ejemplo (Éxito)

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

#### Respuestas de Error

| Status Code | Error Code          | Description                                          |
|-------------|---------------------|------------------------------------------------------|
| 400         | VALIDATION_ERROR    | Solicitud no válida (faltan campos requeridos)            |
| 401         | UNAUTHORIZED        | Encabezado `X-Atlantis-Token` no válido o ausente         |
| 503         | SERVICE_UNAVAILABLE | El almacenamiento de detección de drift no está habilitado en el servidor |
| 500         | INTERNAL_ERROR      | Error interno durante la detección de drift                |

### GET /api/drift/remediate

#### Descripción

Lista resultados de remediación para un repositorio. Devuelve una lista paginada de operaciones de remediación pasadas. Este es un endpoint autenticado que requiere el secreto de API.

::: tip Requisitos previos
La detección de drift debe estar habilitada en el servidor Atlantis. El apply destructivo de remediación además requiere `--enable-drift-remediation`.
:::

#### Parámetros de Consulta

| Name       | Type   | Required | Description                                                  |
|------------|--------|----------|--------------------------------------------------------------|
| repository | string | Yes      | Nombre completo del repositorio (p. ej., `owner/repo`)                    |
| type       | string | Yes      | Tipo del proveedor de VCS (p. ej., `Github`, `Gitlab`, `Gitea`)        |
| limit      | int    | No       | Número máximo de resultados a devolver (por defecto: 10, máximo: 100)  |

#### Solicitud de Ejemplo

```shell
curl --request GET 'https://<ATLANTIS_HOST_NAME>/api/drift/remediate?repository=owner/repo&type=Github' \
--header 'X-Atlantis-Token: <ATLANTIS_API_SECRET>'
```

#### Solicitud de Ejemplo (con limit)

```shell
curl --request GET 'https://<ATLANTIS_HOST_NAME>/api/drift/remediate?repository=owner/repo&type=Github&limit=10' \
--header 'X-Atlantis-Token: <ATLANTIS_API_SECRET>'
```

#### Respuesta de Ejemplo

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

#### Respuesta de Ejemplo (sin resultados)

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

#### Respuestas de Error

| Status Code | Error Code          | Description                                             |
|-------------|---------------------|---------------------------------------------------------|
| 400         | VALIDATION_ERROR    | Falta el parámetro requerido `repository`                 |
| 401         | UNAUTHORIZED        | Encabezado `X-Atlantis-Token` no válido o ausente            |
| 503         | SERVICE_UNAVAILABLE | El almacenamiento de detección de drift no está habilitado en el servidor    |
| 500         | INTERNAL_ERROR      | Error interno al recuperar datos de remediación              |

### GET /api/drift/remediate/{id}

#### Descripción

Obtén un resultado de remediación específico por ID. Devuelve información detallada sobre una operación de remediación pasada, incluidos resultados por proyecto. Este es un endpoint autenticado que requiere el secreto de API.

::: tip Requisitos previos
La detección de drift debe estar habilitada en el servidor Atlantis. El apply destructivo de remediación además requiere `--enable-drift-remediation`.
:::

#### Parámetros de Ruta

| Name | Type   | Required | Description                                |
|------|--------|----------|--------------------------------------------|
| id   | string | Yes      | El identificador único de la remediación   |

::: tip ¿Qué ID?
El `id` aquí es el campo `id` devuelto por una llamada previa a `POST /api/drift/remediate` — no el `detection_id`/`id` devuelto por `POST /api/drift/detect`. Las ejecuciones de detección y remediación se rastrean por separado, cada una con su propio espacio de IDs. Para inspeccionar la salida del plan de una remediación, llama primero a `POST /api/drift/remediate` y usa el `id` de su respuesta.
:::

#### Parámetros de Consulta

| Name       | Type   | Required | Description                                                 |
|------------|--------|----------|-------------------------------------------------------------|
| repository | string | Yes      | Nombre completo del repositorio (p. ej., `owner/repo`)                   |
| type       | string | Yes      | Tipo del proveedor de VCS (`Github`/`Gitlab`/`Gitea`)        |

#### Solicitud de Ejemplo

```shell
curl --request GET 'https://<ATLANTIS_HOST_NAME>/api/drift/remediate/550e8400-e29b-41d4-a716-446655440000?repository=owner/repo&type=Github' \
--header 'X-Atlantis-Token: <ATLANTIS_API_SECRET>'
```

#### Respuesta de Ejemplo

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

#### Respuestas de Error

| Status Code | Error Code          | Description                                                  |
|-------------|---------------------|--------------------------------------------------------------|
| 400         | VALIDATION_ERROR    | Falta un parámetro requerido                                   |
| 401         | UNAUTHORIZED        | Encabezado `X-Atlantis-Token` no válido o ausente                 |
| 403         | FORBIDDEN           | El repositorio no está en la allowlist                           |
| 404         | NOT_FOUND           | Resultado de remediación no encontrado                                 |
| 503         | SERVICE_UNAVAILABLE | El almacenamiento de detección de drift no está habilitado en el servidor         |
| 500         | INTERNAL_ERROR      | Error interno al recuperar datos de remediación                   |

## Otros Endpoints

La mayoría de los endpoints listados en esta sección son no destructivos y por lo tanto no requieren autenticación ni un token secreto especial. `GET /api/drift/status` es un endpoint autenticado de lectura de la API de drift y requiere `X-Atlantis-Token`.

### GET /api/locks

#### Descripción

Lista los locks de proyecto actualmente retenidos.

#### Solicitud de Ejemplo

```shell
curl --request GET 'https://<ATLANTIS_HOST_NAME>/api/locks'
```

#### Respuesta de Ejemplo

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

#### Respuesta de Ejemplo (Sin Locks)

```json
{
  "Locks": []
}
```

### GET /api/drift/status

#### Descripción

Devuelve el estado de drift para un repositorio. Este endpoint proporciona resultados en caché de detección de drift de ejecuciones plan previas. La detección de drift debe estar habilitada en el servidor para que este endpoint funcione y requiere el token de API configurado.

::: tip Requisitos previos
El almacenamiento de detección de drift debe estar habilitado en el servidor Atlantis. Si no está habilitado, este endpoint devuelve un error `503 Service Unavailable`.
:::

#### Parámetros de Consulta

| Name        | Type   | Required | Description                                                   |
|-------------|--------|----------|---------------------------------------------------------------|
| repository  | string | Yes      | Nombre completo del repositorio (p. ej., `owner/repo`)                     |
| type        | string | Yes      | Tipo del proveedor de VCS (p. ej., `Github`, `Gitlab`, `Gitea`)         |
| project     | string | No       | Filtrar por nombre de proyecto                                        |
| path        | string | No       | Filtrar por ruta literal normalizada del proyecto relativa al repositorio |
| workspace   | string | No       | Filtrar por Terraform workspace                                 |
| ref         | string | No       | Filtrar por referencia de git                                       |
| base_branch | string | No       | Filtrar por contexto de rama usado cuando se detectó drift         |

#### Solicitud de Ejemplo

```shell
curl --request GET 'https://<ATLANTIS_HOST_NAME>/api/drift/status?repository=owner/repo&type=Github' \
  --header 'X-Atlantis-Token: <API_TOKEN>'
```

#### Solicitud de Ejemplo (con filtros)

```shell
curl --request GET 'https://<ATLANTIS_HOST_NAME>/api/drift/status?repository=owner/repo&type=Github&project=vpc&path=modules/vpc&workspace=production&ref=main&base_branch=main' \
  --header 'X-Atlantis-Token: <API_TOKEN>'
```

#### Respuesta de Ejemplo (con drift)

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

#### Respuesta de Ejemplo (sin datos de drift)

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

#### Respuestas de Error

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

#### Solicitud de Ejemplo

```shell
curl --request GET 'https://<ATLANTIS_HOST_NAME>/status'
```

#### Respuesta de Ejemplo

```json
{
  "shutting_down": false,
  "in_progress_operations": 0,
  "version": "0.22.3"
}
```

### GET /healthz

#### Descripción

Endpoint de vivacidad. Devuelve 200 si el proceso de Atlantis está en ejecución. No verifica dependencias externas. Adecuado para probes de vivacidad de Kubernetes.

#### Solicitud de Ejemplo

```shell
curl --request GET 'https://<ATLANTIS_HOST_NAME>/healthz'
```

#### Respuesta de Ejemplo

```json
{
  "status": "ok"
}
```

### GET /readyz

#### Descripción

Endpoint de preparación. Devuelve 200 si el servidor está listo para manejar solicitudes, incluida la conectividad a dependencias externas (p. ej. Redis). Devuelve 503 si alguna dependencia es inalcanzable. Adecuado para probes de preparación de Kubernetes.

#### Solicitud de Ejemplo

```shell
curl --request GET 'https://<ATLANTIS_HOST_NAME>/readyz'
```

#### Respuesta de Ejemplo (saludable)

```json
{
  "status": "ok"
}
```

#### Respuesta de Ejemplo (no saludable)

Devuelve HTTP 503:

```json
{
  "status": "error",
  "error": "failed to ping redis: ..."
}
```

### GET /debug/pprof

Si `--enable-profiling-api` está establecido en true, agrega endpoints bajo esta ruta para exponer los datos de profiling del servidor. Consulta [profiling Go programs](https://go.dev/blog/pprof) para más información.
