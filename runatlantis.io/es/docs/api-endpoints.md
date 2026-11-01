# API Endpoints

Aparte de interactuar mediante comentarios de pull request, Atlantis podría responder a un número limitado de endpoints de API.

:::warning API ALPHA - SUJETA A CAMBIOS
Los endpoints de API documentados en esta página están actualmente en **estado alpha** y **no se consideran estables**. Los esquemas de solicitud y respuesta pueden cambiar en cualquier momento sin aviso previo ni período de deprecación.

Si construye integraciones contra estos endpoints, al actualizar Atlantis debe revisar las notas de la versión cuidadosamente y estar preparado para actualizar su código.
:::

## Formato de respuesta

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

Los endpoints de comando y lock actualmente devuelven sus cuerpos originales de nivel superior en lugar de la envoltura de la API de drift. Esto incluye `POST /api/plan`, `POST /api/apply` y los endpoints de lock existentes. Los endpoints de comando devuelven `command.Result` en el nivel superior en caso de éxito o fallo del proyecto, y devuelven un cuerpo `{ "error": "..." }` de nivel superior para errores de solicitud/autenticación/configuración.

### Campos de la respuesta de envoltura

| Field      | Type    | Description                                                             |
| ---------- | ------- | ----------------------------------------------------------------------- |
| success    | boolean | `true` si la solicitud tuvo éxito, `false` en caso contrario            |
| data       | object  | La carga útil de la respuesta (presente en caso de éxito)               |
| error      | object  | Detalles del error (presente en caso de fallo, `null` en caso de éxito) |
| request_id | string  | Identificador único para el rastreo de solicitudes                      |
| timestamp  | string  | Marca de tiempo ISO 8601 de cuándo se generó la respuesta               |

### Formato de respuesta de error de envoltura

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

| Code                | HTTP Status | Description                                        |
| ------------------- | ----------- | -------------------------------------------------- |
| VALIDATION_ERROR    | 400         | Parámetros o cuerpo de solicitud no válidos        |
| UNAUTHORIZED        | 401         | Token de autenticación no válido o faltante        |
| FORBIDDEN           | 403         | Acceso denegado (p. ej., repositorio no permitido) |
| NOT_FOUND           | 404         | Recurso solicitado no encontrado                   |
| INTERNAL_ERROR      | 500         | Error interno del servidor                         |
| SERVICE_UNAVAILABLE | 503         | Función no habilitada o servicio no disponible     |

## Endpoints principales

Los endpoints de API en esta sección están deshabilitados por defecto, ya que estos endpoints de API podrían cambiar la infraestructura directamente.
Para habilitar los endpoints de API, se debe configurar `api-secret`.

:::tip Prerrequisitos

* Establezca `api-secret` como parte de la [Configuración del servidor](server-configuration.md#api-secret)
* Pase `X-Atlantis-Token` con el mismo secreto en el encabezado de la solicitud
  :::

### POST /api/plan

#### Descripción

Ejecute [atlantis plan](using-atlantis.md#atlantis-plan) en el repositorio especificado.

Los proyectos con [`parallel_plan`](repo-level-atlantis-yaml.md#run-plans-and-applies-in-parallel) habilitado se planifican en paralelo, hasta [`--parallel-pool-size`](server-configuration.md#parallel-pool-size) a la vez. En ese caso, los [pre workflow hooks](pre-workflow-hooks.md) de cada proyecto se ejecutan antes de que comience el primer plan, y los [post workflow hooks](post-workflow-hooks.md) se ejecutan después de que termine el último plan. De lo contrario, los hooks de cada proyecto se ejecutan justo antes y después de su propio plan.

#### Parámetros

| Name       | Type     | Required | Description                                         |
| ---------- | -------- | -------- | --------------------------------------------------- |
| Repository | string   | Yes      | Nombre del repositorio de Terraform                 |
| Ref        | string   | Yes      | Referencia de Git, como un nombre de rama           |
| Type       | string   | Yes      | Tipo del proveedor VCS (Github/Gitlab)              |
| Projects   | []string | No       | Lista de nombres de proyectos para ejecutar el plan |
| Paths      | []Path   | No       | Rutas a los proyectos para ejecutar el plan         |
| PR         | int      | No       | Número del Pull Request                             |

::: tip NOTE
Se debe especificar al menos uno de `Projects` o `Paths`.
:::

::: tip Solicitudes API sin PR
Cuando `PR` se omite o se establece en `0`, Atlantis ejecuta la solicitud como un workflow sintético aislado sin PR. Los workflows sintéticos de API usan identidades de pull generadas para directorios de trabajo y locks, realizan checkout reforzado calificado por rama, omiten búsquedas de archivos modificados del pull request, fallan de forma cerrada en caso de denegación de allowlist de equipos, y ordenan los proyectos seleccionados por el orden de ejecución configurado.

Para refs de tag, commit SHA o refs no ambiguas que no son de rama, proporcione `base_branch` para que Atlantis pueda verificar que la ref comprobada es alcanzable desde la rama base prevista. Los nombres de rama como `main` o `feature/foo` se obtienen como `refs/heads/<branch>`.
:::

::: tip Verificaciones de políticas
Cuando las verificaciones de políticas están habilitadas y la selección de proyectos genera contextos `policy_check`, las solicitudes API de plan/apply ejecutan verificaciones de políticas después de contextos de plan exitosos. Las solicitudes API de apply con `apply_requirements: [policies_passed]` requieren un estado de política exitoso antes de aplicar.
:::

::: tip Fase de plan de API apply
El endpoint API apply ejecuta un plan antes de apply. Los errores a nivel de proyecto en esa fase previa a apply no omiten por sí mismos la fase de apply para otros planes pendientes elegibles. El auto-apply de remediación de drift es más estricto y falla de forma cerrada cuando su plan previo a apply tiene errores.
:::

#### Path

Similar a las [Options](using-atlantis.md#options) de `atlantis plan`. Path especifica qué directorio/workspace
 dentro del repositorio ejecutar el plan.
Se debe especificar al menos uno de `Directory` o `Workspace`.

| Name      | Type   | Required | Description                                                                                                                                                               |
| --------- | ------ | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Directory | string | No       | En qué directorio ejecutar plan relativo a la raíz del repositorio. Debe ser una ruta literal relativa al repositorio; vea [Directory validation](#directory-validation). |
| Workspace | string | No       | [Terraform workspace](https://developer.hashicorp.com/terraform/language/state/workspaces) del plan. Use `default` si no se usan Terraform workspaces.                    |

##### Validación de directorio

`Directory` debe ser una ruta literal relativa a la raíz del repositorio, como `.` o `modules/vpc`. Se limpia antes de usarla, por lo que `./modules/vpc/` se convierte en `modules/vpc`. Un `Directory` vacío es lo mismo que omitirlo. La solicitud se rechaza con HTTP 400 y un cuerpo `{ "error": "..." }` de nivel superior si `Directory` es absoluta (incluyendo un `/` solo), se resuelve fuera del repositorio mediante `..`, contiene una barra invertida, o contiene un carácter glob (`*`, `?`, `[`, `]`, `{`, `}`). La misma regla aplica a `POST /api/apply`.

```json
{
  "error": "invalid directory \"/modules/vpc\": directory must be a clean repo-relative path"
}
```

::: warning Cambio de comportamiento
Anteriormente, a `Directory` solo se le quitaban las barras finales. Una barra inicial no se rechazaba: `/modules/vpc` se ejecutaba en `modules/vpc` bajo el clone pero no coincidía con el proyecto `atlantis.yaml` de ese directorio, por lo que se ejecutaba con la configuración predeterminada; un `/` solo significaba la raíz del repositorio; los patrones glob coincidían con proyectos configurados por patrón; y las rutas que escapaban del repositorio fallaban más tarde durante la ejecución. Ahora todo esto se rechaza por adelantado con HTTP 400 e `invalid directory "<value>": directory must be a clean repo-relative path`. Quite la barra inicial (`modules/vpc`), use `.` para la raíz del repositorio, y enumere explícitamente cada directorio de proyecto en lugar de usar un glob.
:::

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

#### Solicitud de ejemplo (Detección de Drift - Sin PR)

Para workflows de detección de drift, omita el parámetro `PR`:

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

#### Respuesta de ejemplo (Éxito)

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

#### Respuesta de ejemplo (Error)

Cuando ocurre un error de solicitud/auth/setup, el endpoint heredado devuelve un cuerpo de error de nivel superior:

```json
{
  "error": "request \"{}\" is missing fields"
}
```

#### Respuesta de ejemplo (Error de proyecto)

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

* `success`: El comando del proyecto se completó exitosamente
* `error`: Ocurrió un error. Las respuestas heredadas de plan/apply conservan la forma histórica del JSON de error de Go: los valores `Error` de proyecto no nulos se codifican como `{}`, y los valores nulos se codifican como `null`.
* `failed`: Ocurrió una falla (revise el campo `failure`)

:::

### POST /api/apply

#### Descripción

Ejecute [atlantis apply](using-atlantis.md#atlantis-apply) en el repositorio especificado.

El endpoint planifica primero los proyectos, de la misma manera que [`POST /api/plan`](#post-api-plan), luego los aplica un proyecto a la vez, incluso cuando [`parallel_apply`](repo-level-atlantis-yaml.md#run-plans-and-applies-in-parallel) está habilitado.

#### Parámetros

| Name       | Type     | Required | Description                                          |
| ---------- | -------- | -------- | ---------------------------------------------------- |
| Repository | string   | Yes      | Nombre del repositorio de Terraform                  |
| Ref        | string   | Yes      | Referencia de Git, como un nombre de rama            |
| Type       | string   | Yes      | Tipo del proveedor VCS (Github/Gitlab)               |
| Projects   | []string | No       | Lista de nombres de proyectos para ejecutar el apply |
| Paths      | []Path   | No       | Rutas a los proyectos para ejecutar el apply         |
| PR         | int      | No       | Número del Pull Request                              |

::: tip NOTE
Se debe especificar al menos uno de `Projects` o `Paths`.
:::

#### Path

Similar a las [Options](using-atlantis.md#options-1) de `atlantis apply`. Path especifica qué directorio/workspace
 dentro del repositorio ejecutar el apply.
Se debe especificar al menos uno de `Directory` o `Workspace`.

| Name      | Type   | Required | Description                                                                                                                                                                |
| --------- | ------ | -------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Directory | string | No       | En qué directorio ejecutar apply relativo a la raíz del repositorio. Debe ser una ruta literal relativa al repositorio; vea [Directory validation](#directory-validation). |
| Workspace | string | No       | [Terraform workspace](https://developer.hashicorp.com/terraform/language/state/workspaces) del plan. Use `default` si no se usan Terraform workspaces.                     |

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

#### Respuesta de ejemplo (Éxito)

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
Las respuestas de error siguen el mismo formato heredado que el endpoint de plan. Vea el [ejemplo de respuesta de error de plan](#sample-response-error) para más detalles.
:::

## Detección y remediación de drift (Alpha)

:::warning ALPHA FEATURE - SUBJECT TO CHANGE
Las APIs de detección de drift, estado de drift, remediación, historial de remediación y webhook de drift son características alpha. Sus campos de solicitud, esquemas de respuesta, semántica de almacenamiento, payloads de webhook y puertas de seguridad pueden cambiar antes de que la característica sea promovida a estable.

La detección de drift ejecuta workflows de Terraform plan y puede ejecutar hooks configurados o pasos de plan personalizados. El apply destructivo de remediación requiere tanto `--enable-drift-detection` como `--enable-drift-remediation`.
:::

### POST /api/drift/remediate

#### Descripción

Ejecute remediación de drift en el repositorio especificado. Este endpoint le permite ejecutar operaciones solo de plan (para previsualizar la remediación) o auto-apply (para corregir automáticamente el drift) para proyectos con drift detectado.

Los proyectos con [`parallel_plan`](repo-level-atlantis-yaml.md#run-plans-and-applies-in-parallel) habilitado se planifican en paralelo, hasta [`--parallel-pool-size`](server-configuration.md#parallel-pool-size) a la vez. El auto-apply sigue aplicando un proyecto a la vez.

::: tip Requisitos previos

* El almacenamiento de detección de drift debe estar habilitado en el servidor Atlantis
* El repositorio debe estar en la lista de repositorios permitidos (si está configurada)

:::

::: tip Orden del workflow
`POST /api/drift/remediate` con `action: "plan"` (el predeterminado) ejecuta un plan nuevo incluso sin datos de drift en caché — no requiere una llamada previa a `POST /api/drift/detect` — cuando se especifican `projects` o `paths`; una remediación sin alcance (ninguno establecido) todavía toma sus objetivos del drift en caché. Solo `action: "apply"` (o `drift_only: true`) requiere un registro de drift en caché para cada proyecto/path/workspace objetivo. Vea el tip "Cached Drift Required" más abajo.
:::

#### Parámetros

| Name        | Type                 | Required    | Description                                                                              |
| ----------- | -------------------- | ----------- | ---------------------------------------------------------------------------------------- |
| repository  | string               | Yes         | Nombre completo del repositorio (p. ej., `owner/repo`)                                   |
| ref         | string               | Yes         | Referencia de Git (rama/tag/commit) a usar para la remediación                           |
| base_branch | string               | Conditional | Contexto de rama para filtros de rama de repo-config y comprobaciones de no divergencia  |
| type        | string               | Yes         | Tipo del proveedor VCS (`Github`/`Gitlab`/`Gitea`)                                       |
| action      | string               | No          | Acción de remediación: `plan` (predeterminada) o `apply`                                 |
| projects    | []string             | No          | Lista de nombres de proyectos a remediar. Si está vacía, usa datos de detección de drift |
| paths       | []DriftDetectionPath | No          | Lista de directorios/workspaces relativos al repositorio a remediar                      |
| workspaces  | []string             | No          | Filtra la remediación a workspaces específicos                                           |
| drift_only  | boolean              | No          | Si es true, solo remedia proyectos con drift detectado                                   |

El campo `paths` usa el mismo objeto `DriftDetectionPath` descrito en `POST /api/drift/detect`.
Para remediación, un selector de path sin `workspace` apunta solo al workspace predeterminado de Terraform.
Use el campo de nivel superior `workspaces` o los valores `workspace` a nivel de path para remediar workspaces no predeterminados.
Los selectores de proyecto para remediación son nombres exactos de proyecto. Los selectores de proyecto con expresiones regulares no son compatibles para remediación; use nombres de proyecto explícitos o selectores de path cuando apunte a varios proyectos.
Los selectores de path de API son rutas literales normalizadas relativas al repositorio; los patrones glob como `envs/*` no son compatibles.

::: tip Acciones

* `plan`: Ejecuta un plan para previsualizar qué cambiaría (predeterminado, no destructivo)
* `apply`: Ejecuta tanto plan como apply para corregir automáticamente el drift (destructivo). Esta acción requiere tanto `--enable-drift-detection` como `--enable-drift-remediation`, además de drift en caché con `has_drift: true` de una ejecución de detección anterior para cada proyecto/path/workspace objetivo.

:::

::: warning Requisitos de apply
El apply de remediación de drift no omite `apply_requirements` del repositorio. Los requisitos que necesitan estado de pull request, como `approved` o `mergeable`, fallan de forma cerrada para solicitudes de remediación sin PR. Use remediación solo de plan o workflows normales de PR para proyectos protegidos por esos requisitos.
:::

::: tip Webhooks
El apply de remediación de drift no dispara webhooks heredados de `event: apply`. Use webhooks de drift para notificaciones de workflow de drift.
:::

::: warning Requisitos de plan
Las acciones solo de plan de remediación de drift y la detección de drift no omiten `plan_requirements` de estado de PR. Los requisitos como `approved` o `mergeable` no pueden satisfacerse sin un pull request y fallan de forma cerrada.
:::

::: tip Seguridad de ref
Cuando la remediación usa drift en caché para una ref móvil como `main`, Atlantis compara el commit actual del checkout con el commit que produjo el registro de drift en caché. Si la ref se movió, vuelva a ejecutar la detección de drift antes de usar `action: "apply"`.
:::

::: tip Se requiere drift en caché
La remediación `action: "apply"` solo aplica registros de drift en caché con `has_drift: true` para el mismo repositorio, ref, `base_branch`, proyecto/path y workspace. Use `action: "plan"` para vistas previas sin caché, luego ejecute detección de drift antes de aplicar.
:::

::: tip Contexto de rama
Para refs de rama como `main`, `feature/foo` o `refs/heads/feature/foo`, Atlantis usa `ref` como contexto de rama y obtiene explícitamente el namespace de rama. Las refs ambiguas desnudas como `prod`, `latest`, `stable` o `v1.2.3` requieren `base_branch` pero aun así se obtienen como nombres de rama. Para commit SHAs sin procesar y refs `refs/tags/...` explícitas, proporcione `base_branch` para que los filtros de rama de repo-config y las comprobaciones de no divergencia se evalúen contra la rama prevista. Use la forma explícita `refs/tags/...` para tags.
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

#### Respuesta de ejemplo (Éxito)

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

#### Respuesta de ejemplo (Éxito de auto-apply)

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

#### Respuesta de ejemplo (Falla parcial)

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

| Status    | Description                                          |
| --------- | ---------------------------------------------------- |
| `pending` | La remediación está en cola pero aún no ha comenzado |
| `running` | La remediación está actualmente en progreso          |
| `success` | Todos los proyectos se remediaron exitosamente       |
| `failed`  | Todos los proyectos fallaron en la remediación       |
| `partial` | Algunos proyectos tuvieron éxito, algunos fallaron   |

#### Respuestas de error

| Status Code | Description                                                                     |
| ----------- | ------------------------------------------------------------------------------- |
| 400         | Solicitud inválida (faltan campos requeridos o acción inválida)                 |
| 401         | Encabezado `X-Atlantis-Token` inválido o faltante                               |
| 403         | El repositorio no está en la lista permitida                                    |
| 409         | La remediación se ejecutó pero todos los proyectos objetivo fallaron            |
| 503         | API, remediación de drift o remediation apply no está habilitado en el servidor |
| 500         | Error interno durante la remediación                                            |

### POST /api/drift/detect

#### Descripción

Dispare la detección de drift para proyectos en un repositorio. Este endpoint inicia una operación de plan para detectar drift de infraestructura sin requerir un pull request. Los resultados se almacenan para recuperación posterior a través de los endpoints de estado de drift.

Los proyectos con [`parallel_plan`](repo-level-atlantis-yaml.md#run-plans-and-applies-in-parallel) habilitado se planifican en paralelo, hasta [`--parallel-pool-size`](server-configuration.md#parallel-pool-size) a la vez.

Cuando los [drift webhooks](sending-notifications-via-webhooks.md#drift-detection-webhooks) están configurados (`event: drift`), las ejecuciones de detección exitosas envían automáticamente notificaciones por webhook a canales de Slack y/o endpoints HTTP, incluyendo resultados heartbeat sin drift.

::: tip Requisitos previos

* El almacenamiento de detección de drift debe estar habilitado en el servidor Atlantis (`--enable-drift-detection`)

:::

#### Parámetros

| Name                | Type                 | Required    | Description                                                                                              |
| ------------------- | -------------------- | ----------- | -------------------------------------------------------------------------------------------------------- |
| repository          | string               | Yes         | Nombre completo del repositorio (p. ej., `owner/repo`)                                                   |
| ref                 | string               | Yes         | Referencia de Git (rama/tag/commit) para verificar drift                                                 |
| base_branch         | string               | Conditional | Contexto de rama para filtros de rama de repo-config y comprobaciones de no divergencia                  |
| type                | string               | Yes         | Tipo del proveedor VCS (`Github`/`Gitlab`/`Gitea`)                                                       |
| projects            | []string             | No          | Lista de nombres de proyecto a verificar. Si está vacía, se verifican todos                              |
| paths               | []DriftDetectionPath | No          | Lista de paths a verificar. Si está vacía, se usan nombres de proyecto                                   |
| include_plan_output | boolean              | No          | Si es true, incluye `plan_output` para cada proyecto en la respuesta. El valor predeterminado es `false` |

#### DriftDetectionPath

| Name      | Type   | Required | Description                                                           |
| --------- | ------ | -------- | --------------------------------------------------------------------- |
| directory | string | Yes      | Ruta relativa al directorio de Terraform                              |
| workspace | string | No       | Terraform workspace. Si se omite, se usa el workspace predeterminado. |

Los selectores de path son rutas literales normalizadas relativas al repositorio. Los patrones glob como `envs/*` no son compatibles.

::: tip NOTE
Se debe especificar al menos uno de `projects` o `paths` para detección dirigida. Si ambos están vacíos, la detección de drift puede escanear todos los proyectos descubiertos. `projects` y `paths` son mutuamente excluyentes para la detección de drift; use un tipo de selector por solicitud.
:::

::: tip Efectos secundarios de estado
La detección de drift suprime los estados normales de commit de plan, verificación de políticas, apply y hooks de Atlantis. Las notificaciones por webhook específicas de drift aún pueden enviarse para ejecuciones de detección exitosas, incluyendo resultados heartbeat sin drift, cuando los drift webhooks están configurados.

La detección de drift no ejecuta Terraform apply, pero sí ejecuta el ciclo de vida normal de plan. Los pre-workflow hooks configurados, workflows personalizados, pasos personalizados de plan y comandos Terraform plan pueden ejecutarse del lado del servidor fuera del contexto de un pull request.

La detección de drift no omite las allowlists de equipo. Si una allowlist de equipo configurada no puede autorizar la solicitud API, la solicitud falla en lugar de escanear o reconciliar un conjunto vacío de proyectos. `plan_requirements` de estado de PR como `approved` o `mergeable` también fallan de forma cerrada para detección de drift sin PR.
:::

::: tip Contexto de rama
Para refs de rama como `main`, `feature/foo` o `refs/heads/feature/foo`, Atlantis usa `ref` como contexto de rama y obtiene explícitamente el namespace de rama. Las refs ambiguas desnudas como `prod`, `latest`, `stable` o `v1.2.3` requieren `base_branch` pero aun así se obtienen como nombres de rama. Para commit SHAs sin procesar y refs `refs/tags/...` explícitas, proporcione `base_branch` para que los filtros de rama de repo-config y las comprobaciones de no divergencia se evalúen contra la rama prevista. Use la forma explícita `refs/tags/...` para tags.
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

::: tip Salida de plan
Establezca `include_plan_output: true` en la solicitud para que la respuesta incluya `plan_output` para cada proyecto — el texto de Terraform plan para ese proyecto. Para el paso de plan incorporado esto normalmente se normaliza para renderizado de diff; el contenido exacto depende del workflow configurado, ya que un paso personalizado `run` puede producir en su lugar salida arbitraria no normalizada. El valor predeterminado es `false`, ya que el texto del plan puede ser grande; cuando se omite o es `false`, `plan_output` no se incluye incluso para proyectos con un plan exitoso. También se omite cuando no hay salida de plan (por ejemplo, si el proyecto tuvo un error antes de que se ejecutara un plan). `plan_output` solo se devuelve alguna vez en esta respuesta de detect; nunca se incluye en `GET /api/drift/status`, ya que nunca se persiste en el almacenamiento de drift.
:::

::: warning La salida de plan puede contener datos sensibles
Antes de que existiera este campo, `POST /api/drift/detect` solo devolvía conteos numéricos de drift. Con `include_plan_output: true`, las respuestas pueden incluir valores de atributos de recursos y, para workflows con paso `run` personalizado, salida arbitraria de comandos. El límite de autenticación del endpoint no cambia (el mismo token API que otros endpoints de drift/remediation), por lo que esto no es una nueva brecha de autorización, pero la sensibilidad de los datos de la respuesta cambia materialmente cuando este campo está habilitado. Solo la redacción filter-regex del paso `run` (si está configurada) se aplica a la salida del plan; por lo demás no se depura.
:::

#### Respuesta de ejemplo (Éxito)

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

| Status Code | Error Code          | Description                                                               |
| ----------- | ------------------- | ------------------------------------------------------------------------- |
| 400         | VALIDATION_ERROR    | Solicitud inválida (faltan campos requeridos)                             |
| 401         | UNAUTHORIZED        | Encabezado `X-Atlantis-Token` inválido o faltante                         |
| 503         | SERVICE_UNAVAILABLE | El almacenamiento de detección de drift no está habilitado en el servidor |
| 500         | INTERNAL_ERROR      | Error interno durante la detección de drift                               |

### GET /api/drift/remediate

#### Descripción

Liste los resultados de remediación para un repositorio. Devuelve una lista paginada de operaciones de remediación pasadas. Este es un endpoint autenticado que requiere el secreto de API.

::: tip Requisitos previos
La detección de drift debe estar habilitada en el servidor Atlantis. El apply destructivo de remediación adicionalmente requiere `--enable-drift-remediation`.
:::

#### Parámetros de consulta

| Name       | Type   | Required | Description                                                              |
| ---------- | ------ | -------- | ------------------------------------------------------------------------ |
| repository | string | Yes      | Nombre completo del repositorio (p. ej., `owner/repo`)                   |
| type       | string | Yes      | Tipo del proveedor VCS (p. ej., `Github`, `Gitlab`, `Gitea`)             |
| limit      | int    | No       | Número máximo de resultados a devolver (predeterminado: 10, máximo: 100) |

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

| Status Code | Error Code          | Description                                                               |
| ----------- | ------------------- | ------------------------------------------------------------------------- |
| 400         | VALIDATION_ERROR    | Falta el parámetro requerido `repository`                                 |
| 401         | UNAUTHORIZED        | Encabezado `X-Atlantis-Token` inválido o faltante                         |
| 503         | SERVICE_UNAVAILABLE | El almacenamiento de detección de drift no está habilitado en el servidor |
| 500         | INTERNAL_ERROR      | Error interno al recuperar datos de remediación                           |

### GET /api/drift/remediate/{id}

#### Descripción

Obtenga un resultado de remediación específico por ID. Devuelve información detallada sobre una operación de remediación pasada, incluidos resultados por proyecto. Este es un endpoint autenticado que requiere el secreto de API.

::: tip Requisitos previos
La detección de drift debe estar habilitada en el servidor Atlantis. El apply destructivo de remediación adicionalmente requiere `--enable-drift-remediation`.
:::

#### Parámetros de path

| Name | Type   | Required | Description                              |
| ---- | ------ | -------- | ---------------------------------------- |
| id   | string | Yes      | El identificador único de la remediación |

::: tip ¿Qué ID?
El `id` aquí es el campo `id` devuelto por una llamada previa a `POST /api/drift/remediate` — no el `detection_id`/`id` devuelto por `POST /api/drift/detect`. Las ejecuciones de detección y remediación se rastrean por separado, cada una con su propio espacio de IDs. Para inspeccionar la salida del plan para una remediación, llame primero a `POST /api/drift/remediate` y use el `id` de su respuesta.
:::

#### Parámetros de consulta

| Name       | Type   | Required | Description                                            |
| ---------- | ------ | -------- | ------------------------------------------------------ |
| repository | string | Yes      | Nombre completo del repositorio (p. ej., `owner/repo`) |
| type       | string | Yes      | Tipo del proveedor VCS (`Github`/`Gitlab`/`Gitea`)     |

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

| Status Code | Error Code          | Description                                                               |
| ----------- | ------------------- | ------------------------------------------------------------------------- |
| 400         | VALIDATION_ERROR    | Falta un parámetro requerido                                              |
| 401         | UNAUTHORIZED        | Encabezado `X-Atlantis-Token` inválido o faltante                         |
| 403         | FORBIDDEN           | El repositorio no está en la allowlist                                    |
| 404         | NOT_FOUND           | No se encontró el resultado de remediación                                |
| 503         | SERVICE_UNAVAILABLE | El almacenamiento de detección de drift no está habilitado en el servidor |
| 500         | INTERNAL_ERROR      | Error interno al recuperar datos de remediación                           |

## Otros endpoints

Los endpoints listados en esta sección no son destructivos, pero no están exentos de autenticación. `GET /api/locks` requiere `X-Atlantis-Token` siempre que un secreto API esté configurado en el servidor y permanece abierto solo cuando no hay ningún secreto configurado, porque el prefijo `/api/` no está cubierto por web basic auth. `GET /api/drift/status` es un endpoint autenticado de lectura de la API de drift y siempre requiere `X-Atlantis-Token`.

### GET /api/locks

#### Descripción

Liste los locks de proyecto actualmente retenidos.

::: warning Cambio de comportamiento
Anteriormente, `GET /api/locks` devolvía la lista de locks a cualquier llamador, incluso cuando `--api-secret` estaba establecido. Ahora, cuando un secreto API está configurado, la solicitud debe enviarlo en el encabezado `X-Atlantis-Token`; sin secreto API configurado el endpoint permanece abierto. Actualice los scripts y dashboards que consultan este endpoint para enviar el encabezado. Un token faltante o incorrecto devuelve HTTP 401 con código de error `UNAUTHORIZED` y mensaje `invalid or missing API token`.
:::

#### Solicitud de ejemplo

```shell
curl --request GET 'https://<ATLANTIS_HOST_NAME>/api/locks' --header 'X-Atlantis-Token: <ATLANTIS_API_SECRET>'
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

#### Respuesta de ejemplo (Sin locks)

```json
{
  "Locks": []
}
```

### GET /api/drift/status

#### Descripción

Devuelva el estado de drift para un repositorio. Este endpoint proporciona resultados en caché de detección de drift de ejecuciones de plan previas. La detección de drift debe estar habilitada en el servidor para que este endpoint funcione y requiere el token API configurado.

::: tip Requisitos previos
El almacenamiento de detección de drift debe estar habilitado en el servidor Atlantis. Si no está habilitado, este endpoint devuelve un error `503 Service Unavailable`.
:::

#### Parámetros de consulta

| Name        | Type   | Required | Description                                                               |
| ----------- | ------ | -------- | ------------------------------------------------------------------------- |
| repository  | string | Yes      | Nombre completo del repositorio (p. ej., `owner/repo`)                    |
| type        | string | Yes      | Tipo del proveedor VCS (p. ej., `Github`, `Gitlab`, `Gitea`)              |
| project     | string | No       | Filtrar por nombre de proyecto                                            |
| path        | string | No       | Filtrar por ruta literal normalizada del proyecto relativa al repositorio |
| workspace   | string | No       | Filtrar por Terraform workspace                                           |
| ref         | string | No       | Filtrar por referencia de git                                             |
| base_branch | string | No       | Filtrar por el contexto de rama usado cuando se detectó drift             |

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

| Status Code | Error Code          | Description                                             |
| ----------- | ------------------- | ------------------------------------------------------- |
| 400         | VALIDATION_ERROR    | Falta el parámetro requerido `repository`               |
| 401         | UNAUTHORIZED        | Encabezado `X-Atlantis-Token` inválido o faltante       |
| 403         | FORBIDDEN           | El repositorio no está en la allowlist                  |
| 503         | SERVICE_UNAVAILABLE | La detección de drift no está habilitada en el servidor |
| 500         | INTERNAL_ERROR      | Error interno al recuperar datos de drift               |

### GET /status

#### Descripción

Devuelva el estado del servidor Atlantis.

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

Endpoint de liveness. Devuelve 200 si el proceso de Atlantis está en ejecución. No verifica dependencias externas. Adecuado para probes de liveness de Kubernetes.

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

Endpoint de readiness. Devuelve 200 si el servidor está listo para manejar solicitudes, incluida la conectividad con dependencias externas (p. ej. Redis). Devuelve 503 si alguna dependencia es inalcanzable. Adecuado para probes de readiness de Kubernetes.

#### Solicitud de ejemplo

```shell
curl --request GET 'https://<ATLANTIS_HOST_NAME>/readyz'
```

#### Respuesta de ejemplo (healthy)

```json
{
  "status": "ok"
}
```

#### Respuesta de ejemplo (unhealthy)

Devuelve HTTP 503:

```json
{
  "status": "error",
  "error": "failed to ping redis: ..."
}
```

### GET /debug/pprof

Si `--enable-profiling-api` se establece en true, agrega endpoints bajo esta ruta para exponer los datos de profiling del servidor. Consulte [profiling Go programs](https://go.dev/blog/pprof) para más información.
