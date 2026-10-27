# Configuración del servidor

Esta página explica cómo configurar el comando `atlantis server`.

La configuración de `atlantis server` puede especificarse mediante flags de línea de comandos,
variables de entorno, un archivo de configuración o una mezcla de los tres.

## Variables de entorno

Todos los flags pueden especificarse como variables de entorno.

1. Tome el nombre del flag, p. ej. `--gh-user`
1. Ignore el primer `--` => `gh-user`
1. Convierta los `-` en `_` => `gh_user`
1. Ponga todas las letras en mayúsculas => `GH_USER`
1. Añada el prefijo `ATLANTIS_` => `ATLANTIS_GH_USER`

::: warning NOTE
Para establecer un flag booleano use `true` o `false` como valor.
:::

::: warning NOTE
El flag `--atlantis-url` se establece mediante la variable de entorno `ATLANTIS_ATLANTIS_URL` **NO** `ATLANTIS_URL`.
:::

## Archivo de configuración

Todos los flags también pueden especificarse mediante un archivo de configuración YAML.

Para usar un archivo de configuración YAML, ejecute `atlantis server --config /path/to/config.yaml`.

Las claves de su archivo de configuración deben ser las mismas que los nombres de los flags, p. ej.

```yaml
gh-token: ...
log-level: ...
```

::: warning
El archivo de configuración que pasa a `--config` es diferente del archivo `--repo-config`.
El archivo de configuración de `--config` solo se usa como una forma alternativa de establecer flags de `atlantis server`.
:::

## Precedencia

Los valores se eligen en este orden:

1. Flags
1. Variables de entorno
1. Archivo de configuración

## Flags

### `--allow-commands`

```bash
atlantis server --allow-commands=version,plan,apply,unlock,approve_policies,cancel
# or
ATLANTIS_ALLOW_COMMANDS='version,plan,apply,unlock,approve_policies,cancel'
```

Lista de comandos permitidos para ejecutarse en el servidor Atlantis, por defecto es `version,plan,apply,unlock,approve_policies,cancel`

Notas:

- Acepta una lista separada por comas, p. ej. `command1,command2`.
- `version`, `plan`, `apply`, `unlock`, `approve_policies`, `cancel`, `import`, `state`, `policy_check` y `all` están disponibles.
- `policy_check` es un comando interno que se ejecuta automáticamente después de `plan` cuando [policy checking](policy-checking.md) está habilitado. Debe incluirse explícitamente en la allowlist al usar [`--gh-team-allowlist`](#gh-team-allowlist).
- `all` es una palabra clave especial que permite todos los comandos. Si pasa `all` entonces todos los demás comandos serán ignorados.

### `--allow-draft-prs`

```bash
atlantis server --allow-draft-prs
# or
ATLANTIS_ALLOW_DRAFT_PRS=true
```

Responder a pull requests de draft prs. El valor predeterminado es `false`.

### `--allow-fork-prs`

```bash
atlantis server --allow-fork-prs
# or
ATLANTIS_ALLOW_FORK_PRS=true
```

Responder a pull requests desde forks. El valor predeterminado es `false`.

:::warning SECURITY WARNING
Potencialmente peligroso de habilitar
porque si atacantes pueden crear un pull request a su repo entonces pueden hacer que Atlantis
ejecute código arbitrario. Esto puede ocurrir porque
Atlantis ejecutará automáticamente `terraform plan`
que puede ejecutar código arbitrario si recibe una configuración Terraform maliciosa.
:::

Independientemente de este flag, Atlantis solo ejecuta comandos para un pull request cuyo repositorio
head está en el mismo host que su repositorio base.

::: warning Cambio de comportamiento
Anteriormente, Atlantis no comparaba los hosts de los repositorios head y base. Ahora, si el host de la clone URL del repositorio head difiere del del repositorio base (sin distinguir mayúsculas de minúsculas), o no tiene host, Atlantis no ejecuta el comando y no publica ningún comentario. Busque la línea del log del servidor `head repo host "<head>" does not match base repo host "<base>", not running command`. Para GitHub, vea también [`--gh-hostname`](#gh-hostname).
:::

### `--api-secret`

```bash
atlantis server --api-secret="secret"
# or (recommended)
ATLANTIS_API_SECRET="secret"
```

Secreto requerido usado para validar solicitudes realizadas a los endpoints de [`/api/*`](api-endpoints.md).

Cuando está configurado, también es requerido por el endpoint de solo lectura [`GET /api/locks`](api-endpoints.md#get-api-locks).

### `--atlantis-url`

```bash
atlantis server --atlantis-url="https://my-domain.com:9090/basepath"
# or
ATLANTIS_ATLANTIS_URL=https://my-domain.com:9090/basepath
```

Especifique la URL desde la que Atlantis es accesible. Se usa en la UI de Atlantis
y en enlaces desde comentarios de pull request. El valor predeterminado es `http://$(hostname):$port`
donde `$port` proviene del flag [`--port`](#port). Soporta un basepath si aloja Atlantis bajo una ruta.

Notas:

- Si se usa un balanceador de carga con un puerto no http/https (no el definido en el flag `--port`), actualice la URL para incluir el puerto como en el ejemplo anterior.
- Esta URL se usa como el enlace `details` junto a cada trabajo de atlantis para ver los logs del trabajo.

### `--autodiscover-mode`

```bash
atlantis server --autodiscover-mode="<auto|enabled|disabled>"
# or
ATLANTIS_AUTODISCOVER_MODE="<auto|enabled|disabled>"
```

Establece el modo auto discover; el valor predeterminado es `auto`. Cuando se establece en `auto`, los proyectos en un repo serán descubiertos por
Atlantis cuando no haya proyectos configurados en la config del repo. Si uno o más proyectos están definidos
en la config del repo, entonces el auto discovery quedará completamente deshabilitado.

Cuando se establece en `enabled` los proyectos serán descubiertos incondicionalmente. Si un proyecto auto descubierto ya está
definido en la sección projects de la config del repo, el proyecto de la config del repo tendrá precedencia sobre
el proyecto auto descubierto.

Cuando se establece en `disabled` los proyectos nunca serán descubiertos, incluso si no hay proyectos configurados en la config del repo.

### `--automerge`

```bash
atlantis server --automerge
# or
ATLANTIS_AUTOMERGE=true
```

Fusiona automáticamente pull requests después de que todos los plans se hayan aplicado correctamente.
El valor predeterminado es `false`. Vea [Automerging](automerging.md) para más detalles.

### `--automerge-method`

```bash
atlantis server --automerge-method="squash"
# or
ATLANTIS_AUTOMERGE_METHOD="squash"
```

Método de merge predeterminado a usar al hacer automerging de pull requests. Los valores válidos son
`merge`, `rebase` y `squash`. Cuando no está configurado, se usa el método de merge
predeterminado del proveedor VCS. Esto puede sobrescribirse por comando con el flag de comentario `--auto-merge-method`.
Actualmente solo está implementado para GitHub.

### `--autoplan-file-list`

```bash
# NOTE: Use single quotes to avoid shell expansion of *.
atlantis server --autoplan-file-list='**/*.tf,project1/*.pkr.hcl'
# or
ATLANTIS_AUTOPLAN_FILE_LIST='**/*.tf,project1/*.pkr.hcl'
```

Lista de patrones de archivos que Atlantis usará para verificar si un directorio contiene archivos modificados que deben disparar el planning del proyecto.

Notas:

- Acepta una lista separada por comas, p. ej. `pattern1,pattern2`.
- Los patrones usan la [sintaxis de `.dockerignore`](https://docs.docker.com/engine/reference/builder/#dockerignore-file)
- La lista de patrones de archivos será usada tanto por plans automáticos como por plans ejecutados manualmente.
- Cuando no está configurado, el valor predeterminado es todos los archivos `.tf`, `.tf.json`, `.tfvars`, `.tfvars.json`, `.tofu`, `.tofu.json`, `terragrunt.hcl` e `.terraform.lock.hcl`
   (`--autoplan-file-list='**/*.tf,**/*.tf.json,**/*.tfvars,**/*.tfvars.json,**/*.tofu,**/*.tofu.json,**/terragrunt.hcl,**/.terraform.lock.hcl'`).
- Configurar `--autoplan-file-list` sobrescribirá los valores predeterminados. **Debe** agregar `**/*.tf` y otros valores predeterminados si quiere incluirlos.
- El valor predeterminado es global (no distingue distribución). Tanto las instalaciones de Terraform como de OpenTofu coincidirán con cambios en `.tofu`. Si no usa OpenTofu, puede sobrescribir el valor predeterminado para excluir los patrones `.tofu`.
- Un [Workflow](repo-level-atlantis-yaml.md#configuring-planning) personalizado que use autoplan `when_modified` ignorará este valor.

Ejemplos:

- Autoplan cuando cualquier archivo `*.tf` o `*.tfvars` es modificado.
  - `--autoplan-file-list='**/*.tf,**/*.tfvars'`
- Autoplan cuando cualquier archivo `*.tf` es modificado excepto en el directorio `project2/`
  - `--autoplan-file-list='**/*.tf,!project2'`
- Autoplan cuando cualquier archivo `*.tf` o archivos `.yml` en una subcarpeta de `project1` es modificado.
  - `--autoplan-file-list='**/*.tf,project1/**/*.yml'`

::: warning NOTA
De forma predeterminada, los cambios a módulos no dispararán autoplanning. Vea los flags a continuación.
:::

### `--autoplan-modules`

```bash
atlantis server --autoplan-modules
# or
ATLANTIS_AUTOPLAN_MODULES=true
```

El valor predeterminado es `false`. Cuando se establece en `true`, Atlantis rastreará los módulos locales de los proyectos incluidos.
Los proyectos incluidos son proyectos con archivos incluidos por `--autoplan-file-list`.
Después del rastreo, Atlantis hará plan de cualquier proyecto que incluya un módulo cambiado. Esto es equivalente a establecer
`--autoplan-modules-from-projects` al valor de `--autoplan-file-list`. Vea abajo.

::: tip NOTA
La indexación de dependencias de módulos usa inspección de configuración de Terraform y puede no soportar completamente archivos `.tofu` / `.tofu.json`. Las llamadas a módulos definidas solo en archivos `.tofu` o directorios de módulos compartidos que contienen solo archivos `.tofu` pueden no ser rastreadas. El autoplanning directo por cambio de archivo para proyectos `.tofu` funciona independientemente. Use patrones `autoplan.when_modified` explícitos como solución alternativa. Vea [soporte de archivos OpenTofu .tofu](terraform-versions.md#opentofu-tofu-file-support) para detalles.
:::

### `--autoplan-modules-from-projects`

```bash
atlantis server --autoplan-modules-from-projects='**/init.tf'
# or
ATLANTIS_AUTOPLAN_MODULES_FROM_PROJECTS='**/init.tf'
```

Habilita auto-planing de proyectos cuando una dependencia de módulo en el mismo repositorio ha cambiado.
Esta es una lista de patrones de archivos como `autoplan-file-list`.

Estos patrones seleccionan **proyectos** para indexar según los archivos coincidentes. El índice mapea módulos a los proyectos que dependen de ellos,
incluyendo proyectos que incluyen el módulo a través de otros módulos. Cuando cambia un archivo de módulo que coincide con `autoplan-file-list`,
se hará plan de todos los proyectos indexados.

El valor predeterminado actual es "" (deshabilitado).

Ejemplos:

- `**/*.tf` - indexará todos los proyectos que tengan un archivo `.tf` en su directorio, y hará plan de ellos siempre que una dependencia de módulo dentro del repo haya cambiado.
- `**/*.tf,!foo,!bar` - indexará todos los proyectos que contengan `.tf` excepto `foo` y `bar` y hará plan de ellos siempre que una dependencia de módulo dentro del repo haya cambiado.
   Esto permite que los proyectos opten por no participar en auto-planning cuando cambia una dependencia de módulo.

::: warning NOTA
Los módulos que no son seleccionados por autoplan-file-list no serán indexados y no se hará plan de los proyectos dependientes. Este
flag permite seleccionar los _proyectos_ a indexar, pero el disparador para un plan debe ser un archivo en `autoplan-file-list`.
:::

::: warning NOTA
Este flag sobrescribe `--autoplan-modules`. Si desea deshabilitar auto-planning de módulos, establezca este flag a una cadena vacía,
y establezca `--autoplan-modules` en `false`.
:::

### `--azuredevops-hostname`

```bash
atlantis server --azuredevops-hostname="dev.azure.com"
# or
ATLANTIS_AZUREDEVOPS_HOSTNAME="dev.azure.com"
```

Hostname de Azure DevOps para soportar instancias cloud y self-hosted. El valor predeterminado es `dev.azure.com`.

::: warning ADVERTENCIA DE COMPATIBILIDAD
Si le afecta este cambio [docs](https://learn.microsoft.com/en-us/azure/devops/release-notes/2018/sep-10-azure-devops-launch#administration)
o este [issue](https://github.com/runatlantis/atlantis/issues/5595)
ambos Service Hooks (v1 y v2) convertirán el nombre de la organización AD a minúsculas:
Ejemplos:
`https://dev.azure.com/MYCompany/` e `https://mycompany.visualstudio.com/` serán convertidos a `mycompany`
`https://dev.azure.com/MYCOMPANY/` e `https://myCOMPANY.visualstudio.com/` serán convertidos a `mycompany`

Este [cambio](https://github.com/runatlantis/atlantis/pull/5596) será aplicado desde la versión v0.35.0

¿Qué hacer si tiene plans pendientes que fueron generados con una versión anterior?
Ejecutar un atlantis unlock desde v0.35.0 en sus PR actuales ignorará los archivos en la carpeta `MYCompany`. En el siguiente atlantis plan se usará la carpeta `mycompany` y se generará todo en el nuevo nombre de carpeta
:::

### `--azuredevops-token`

```bash
atlantis server --azuredevops-token="RandomStringProducedByAzureDevOps"
# or (recommended)
ATLANTIS_AZUREDEVOPS_TOKEN="RandomStringProducedByAzureDevOps"
```

Token de Azure DevOps del usuario de API.

### `--azuredevops-user`

```bash
atlantis server --azuredevops-user="username@example.com"
# or
ATLANTIS_AZUREDEVOPS_USER="username@example.com"
```

Nombre de usuario de Azure DevOps del usuario de API.

### `--azuredevops-webhook-password`

```bash
atlantis server --azuredevops-webhook-password="password123"
# or (recommended)
ATLANTIS_AZUREDEVOPS_WEBHOOK_PASSWORD="password123"
```

Contraseña de autenticación básica de Azure DevOps para webhooks entrantes (vea
[docs](https://docs.microsoft.com/en-us/azure/devops/service-hooks/authorize?view=azure-devops)).

::: warning ADVERTENCIA DE SEGURIDAD
Si no se especifica, Atlantis no podrá validar que la
llamada webhook entrante provino de su organización de Azure DevOps. Esto significa que un
atacante podría falsificar llamadas a Atlantis y hacer que realice acciones maliciosas.
Debe especificarse mediante la variable de entorno `ATLANTIS_AZUREDEVOPS_WEBHOOK_PASSWORD`.
:::

### `--azuredevops-webhook-user`

```bash
atlantis server --azuredevops-webhook-user="username@example.com"
# or
ATLANTIS_AZUREDEVOPS_WEBHOOK_USER="username@example.com"
```

Nombre de usuario de autenticación básica de Azure DevOps para webhooks entrantes.

### `--bitbucket-api-user`

```bash
atlantis server --bitbucket-api-user="apiuser@example.com"
# or
ATLANTIS_BITBUCKET_API_USER="apiuser@example.com"
```

Nombre de usuario de Bitbucket (normalmente un email) usado para autenticación de API con Bitbucket Cloud. Esto se usa solo para llamadas de API. Si no se especifica, Atlantis usará el valor de `--bitbucket-user` para autenticación de API para mantener compatibilidad hacia atrás.

**Nota:**

- La compatibilidad hacia atrás es para soportar los Bitbucket APP Passwords existentes que siguen siendo válidos hasta junio de 2026 (vea [el aviso de deprecación de app passwords de Bitbucket de Atlassian](https://www.atlassian.com/blog/bitbucket/bitbucket-cloud-transitions-to-api-tokens-enhancing-security-with-app-password-deprecation)).

**Clave del archivo de configuración:**

```yaml
bitbucket-api-user: apiuser@example.com
```

**Variable de entorno:** `ATLANTIS_BITBUCKET_API_USER`

**Nota:** Este flag solo es relevante para integraciones con Bitbucket Cloud (bitbucket.org).

### `--bitbucket-base-url`

```bash
atlantis server --bitbucket-base-url="http://bitbucket.corp:7990/basepath"
# or
ATLANTIS_BITBUCKET_BASE_URL="http://bitbucket.corp:7990/basepath"
```

URL base de la instalación de Bitbucket Server (también conocido como Stash). Debe incluir
`http://` o `https://`. Si usa Bitbucket Cloud (bitbucket.org), no lo configure. El valor predeterminado es
`https://api.bitbucket.org`.

### `--bitbucket-token`

```bash
atlantis server --bitbucket-token="token"
# or (recommended)
ATLANTIS_BITBUCKET_TOKEN="token"
```

App password de Bitbucket del usuario de API.

### `--bitbucket-user`

```bash
atlantis server --bitbucket-user="myuser"
# or
ATLANTIS_BITBUCKET_USER="myuser"
```

Nombre de usuario de Bitbucket usado para operaciones git. Para Bitbucket Cloud, si `--bitbucket-api-user` no está especificado, este valor también se usará para autenticación de API.

### `--bitbucket-webhook-secret`

```bash
atlantis server --bitbucket-webhook-secret="secret"
# or (recommended)
ATLANTIS_BITBUCKET_WEBHOOK_SECRET="secret"
```

Secreto usado para validar webhooks de Bitbucket.

::: warning ADVERTENCIA DE SEGURIDAD
Si no se especifica, Atlantis no podrá validar que la llamada webhook entrante provino de Bitbucket.
Esto significa que un atacante podría falsificar llamadas a Atlantis y hacer que realice acciones maliciosas.
:::

### `--blocked-extra-args`

```bash
atlantis server --blocked-extra-args="-lock-timeout,-parallelism"
# or
ATLANTIS_BLOCKED_EXTRA_ARGS='-lock-timeout,-parallelism'
```

Lista separada por comas de flags adicionales de Terraform CLI que no están permitidos en argumentos extra de comentarios (los flags después de `--`).
Estos flags siempre están bloqueados, esté o no configurado este flag: `-chdir,-plugin-dir,-out,-generate-config-out,-state,-state-out,-backup`.

Notas:

- `-chdir` e `-plugin-dir` están bloqueados porque mueven el directorio de trabajo o la fuente del proveedor de Terraform a una ubicación que el autor del comentario controla.
- `-out`, `-generate-config-out`, `-state`, `-state-out` e `-backup` están bloqueados porque hacen que Terraform escriba un archivo, o lea state, en una ruta que el autor del comentario elige. Atlantis gestiona por sí mismo las ubicaciones de los archivos de plan y state.
- Configurar este flag **agrega** a la lista siempre bloqueada; no puede desbloquear ninguno de ellos.
- Un comentario que usa un flag bloqueado es rechazado con `flag "<arg>" is not allowed in extra args`.
- Para generar configuración para bloques `import`, que necesita `-generate-config-out`, defina un [custom workflow](custom-workflows.md) del lado del servidor cuyo paso `plan` pase el flag en `extra_args`, por ejemplo `- plan: {extra_args: ["-generate-config-out=generated.tf"]}`, y selecciónelo para el proyecto solo mientras genera configuración.
- Un flag se hace coincidir por nombre, por lo que listar `-lock-timeout` bloquea `-lock-timeout`, `--lock-timeout`, `-lock-timeout=VALUE` e `--lock-timeout=VALUE`. No necesita listar ambas grafías con guion.
- Solo se verifican los argumentos extra de comentarios. Los flags establecidos por el operador del servidor, como `extra_args` en un workflow del lado del servidor o en un workflow a nivel de repo que el servidor permite, se pasan a Terraform sin cambios.
- Acepta una lista separada por comas, p. ej. `-flag1,-flag2`.

::: warning Cambio de comportamiento
Anteriormente, solo `-chdir` e `-plugin-dir` estaban bloqueados por defecto, una entrada coincidía solo con la grafía con guion listada (`-foo` no bloqueaba `--foo`), y configurar `--blocked-extra-args` reemplazaba los valores predeterminados. Ahora `-out`, `-generate-config-out`, `-state`, `-state-out` e `-backup` también siempre están bloqueados, ambas grafías con guion coinciden, e `--blocked-extra-args` solo agrega flags. Comentarios como `atlantis plan -- -generate-config-out=generated.tf` o `atlantis state rm ADDRESS -- -backup=FILE` ahora fallan con `flag "<arg>" is not allowed in extra args`; mueva los flags que aún necesite al `extra_args` de un workflow del lado del servidor. Una lista personalizada existente sigue funcionando y ahora también obtiene las protecciones integradas.
:::

### `--checkout-depth`

```bash
atlantis server --checkout-depth=0
# or
ATLANTIS_CHECKOUT_DEPTH=0
```

El número de commits a obtener de la rama. Se usa si `--checkout-strategy=merge`, ya que la estrategia de checkout `--checkout-strategy=branch` (predeterminada) siempre usa por defecto un clone superficial con una profundidad de 1.
El valor predeterminado es `0`. Vea [Checkout Strategy](checkout-strategy.md) para más detalles.

### `--checkout-strategy`

```bash
atlantis server --checkout-strategy="<branch|merge>"
# or
ATLANTIS_CHECKOUT_STRATEGY="<branch|merge>"
```

Cómo hacer checkout de pull requests. Use `branch` o `merge`.
El valor predeterminado es `branch`. Vea [Checkout Strategy](checkout-strategy.md) para más detalles.

### `--config`

```bash
atlantis server --config="my/config/file.yaml"
# or
ATLANTIS_CONFIG="my/config/file.yaml"
```

Archivo de configuración YAML donde también se pueden establecer flags. Vea [Config File](#config-file) para más detalles.

### `--data-dir`

```bash
atlantis server --data-dir="path/to/data/dir"
# or
ATLANTIS_DATA_DIR="path/to/data/dir"
```

Directorio donde Atlantis almacenará sus datos. Se creará si no existe.
El valor predeterminado es `~/.atlantis`. Atlantis almacenará aquí su base de datos, repos checkout, plans de Terraform por defecto y binarios de
Terraform descargados. Si Atlantis pierde este directorio, los [locks](locking.md)
se perderán y los plans no aplicados se perderán.

Tenga en cuenta que el usuario atlantis está restringido a `~/.atlantis`.
Si establece el flag `--data-dir` en una ruta fuera del directorio home de Atlantis, asegúrese de otorgar al usuario atlantis los permisos correctos.

### `--default-tf-distribution`

```bash
atlantis server --default-tf-distribution="terraform"
# or
ATLANTIS_DEFAULT_TF_DISTRIBUTION="terraform"
```

Qué distribución TF usar. Puede establecerse en `terraform` o `opentofu`.

### `--default-tf-version`

```bash
atlantis server --default-tf-version="v0.12.31"
# or
ATLANTIS_DEFAULT_TF_VERSION="v0.12.31"
```

Versión de Terraform a usar por defecto. Se descargará a `<data-dir>/bin/terraform<version>`
si no está en `PATH`. Vea [Terraform Versions](terraform-versions.md) para más detalles.

### `--disable-apply-all`

```bash
atlantis server --disable-apply-all
# or
ATLANTIS_DISABLE_APPLY_ALL=true
```

Deshabilita el comando `atlantis apply` para que deba
especificarse un proyecto/workspace/directorio específico para los apply.

### `--disable-automerge-label`

```bash
atlantis server --disable-automerge-label="no-auto-merge"
# or
ATLANTIS_DISABLE_AUTOMERGE_LABEL="no-auto-merge"
```

Deshabilita atlantis automerge solo en pull requests con la etiqueta especificada.
El valor predeterminado es una cadena vacía, por lo que ninguna etiqueta deshabilita automerge por defecto.
Este flag no tiene efecto a menos que automerge esté habilitado con `--automerge` o
`automerge: true` a nivel de repo.

### `--disable-autoplan`

```bash
atlantis server --disable-autoplan
# or
ATLANTIS_DISABLE_AUTOPLAN=true
```

Deshabilita auto planning de atlantis.

### `--disable-autoplan-label`

```bash
atlantis server --disable-autoplan-label="no-autoplan"
# or
ATLANTIS_DISABLE_AUTOPLAN_LABEL="no-autoplan"
```

Deshabilita auto planning de atlantis solo en pull requests con la etiqueta especificada.

Si la propiedad `disable-autoplan` es `true`, este flag no tiene efecto.

### `--disable-global-apply-lock`

```bash
atlantis server --disable-global-apply-lock
# or
ATLANTIS_DISABLE_GLOBAL_APPLY_LOCK=true
```

Si es true, elimina el botón en la UI que permite a los usuarios deshabilitar globalmente comandos apply.

### `--disable-markdown-folding`

```bash
atlantis server --disable-markdown-folding
# or
ATLANTIS_DISABLE_MARKDOWN_FOLDING=true
```

Deshabilita el plegado en la salida markdown usando la etiqueta html `<details>`.

### `--disable-repo-locking`

```bash
atlantis server --disable-repo-locking
# or
ATLANTIS_DISABLE_REPO_LOCKING=true
```

Evita que atlantis bloquee proyectos y/o workspaces al ejecutar terraform.

### `--disable-unlock-label`

```bash
atlantis server --disable-unlock-label do-not-unlock
# or
ATLANTIS_DISABLE_UNLOCK_LABEL="do-not-unlock"
```

Evita que atlantis desbloquee un pull request con esta etiqueta. El valor predeterminado es "" (funcionalidad deshabilitada).

### `--discard-approval-on-plan`

```bash
atlantis server --discard-approval-on-plan
# or
ATLANTIS_DISCARD_APPROVAL_ON_PLAN=true
```

Si está configurado, descarta la aprobación si se ha ejecutado un nuevo plan. Actualmente solo soportado en GitHub y GitLab. Para GitLab se requiere un bot, group o project token para esta funcionalidad.
 Referencia: [reset-approvals-of-a-merge-request](https://docs.gitlab.com/api/merge_request_approvals/#reset-approvals-of-a-merge-request)

### `--emoji-reaction`

```bash
atlantis server --emoji-reaction eyes
# or
ATLANTIS_EMOJI_REACTION=eyes
```

La reacción emoji a usar para marcar comentarios procesados. Actualmente soportado en Gitea, GitHub y GitLab. Si no se especifica, Atlantis no usará una reacción emoji.
El valor predeterminado es "" (cadena vacía).

::: warning NOTA
Cada proveedor VCS soporta una lista diferente de emojis:

- [GitHub](https://docs.github.com/en/rest/reactions/reactions?apiVersion=2022-11-28#about-reactions)
- [GitLab](https://gitlab.com/gitlab-org/gitlab/-/blob/master/fixtures/emojis/digests.json)
- [Gitea](https://docs.gitea.com/administration/customizing-gitea#reactions)

   :::

### `--enable-diff-markdown-format`

```bash
atlantis server --enable-diff-markdown-format
# or
ATLANTIS_ENABLE_DIFF_MARKDOWN_FORMAT=true
```

Habilita Atlantis para formatear la salida de Terraform plan en un formato amigable con markdown-diff para propósitos de codificación por colores.

Útil para habilitar al usar GitHub. Las líneas cambiadas dentro de diffs de heredoc y cadenas multilínea de Terraform también se formatean para que los renderizadores markdown con reconocimiento de diff puedan colorearlas.

### `--enable-drift-detection`

```bash
atlantis server --enable-drift-detection
# or
ATLANTIS_ENABLE_DRIFT_DETECTION=true
```

Habilita endpoints de API de detección de drift. La detección de drift no ejecuta Terraform apply, pero
sí ejecuta el ciclo de vida normal de plan, incluidos pre-workflow hooks configurados,
custom workflows, pasos custom de plan y comandos Terraform plan. Cuando está habilitado, Atlantis
inicializará almacenamiento en memoria para resultados de detección de drift y un servicio de remediación,
haciendo funcionales los endpoints de detección de drift, estado y remediación solo plan. Si se configuran [webhooks](sending-notifications-via-webhooks.md#drift-detection-webhooks) de drift
(`event: drift`), las ejecuciones exitosas de detección envían notificaciones a Slack o endpoints HTTP,
incluyendo resultados heartbeat sin drift. La detección de drift no omite allowlists de equipos ni
`plan_requirements` de estado de PR como `approved` o `mergeable`; esas verificaciones fallan de forma cerrada cuando
no pueden evaluarse fuera de un pull request. Las acciones destructivas de apply de remediación de drift también requieren
`--enable-drift-remediation`. El valor predeterminado es `false`.

### `--enable-drift-remediation`

```bash
atlantis server --enable-drift-detection --enable-drift-remediation
# or
ATLANTIS_ENABLE_DRIFT_DETECTION=true
ATLANTIS_ENABLE_DRIFT_REMEDIATION=true
```

Habilita acciones destructivas de apply para remediación de drift en el endpoint `/api/drift/remediate`.
Este flag requiere `--enable-drift-detection`; sin él, las solicitudes `action: "apply"` son
rechazadas mientras la detección de drift de solo lectura permanece disponible. Este flag no omite
`apply_requirements` del repositorio; los requisitos que necesitan estado de pull request fallan de forma cerrada para
solicitudes de remediación que no son de PR. El valor predeterminado es `false`.

### `--enable-external-stores`

```bash
atlantis server --enable-external-stores
# or
ATLANTIS_ENABLE_EXTERNAL_STORES=true
```

Habilita backends de almacenamiento externo configurados en la config del repo del lado del servidor (bloque `external_stores`). Cuando está configurado, Atlantis lee la sección `external_stores` del YAML de config del repo para inicializar backends como S3 para persistencia de archivos de plan.

### `--enable-policy-checks`

```bash
atlantis server --enable-policy-checks
# or
ATLANTIS_ENABLE_POLICY_CHECKS=true
```

Habilita a atlantis para ejecutar políticas del lado del servidor sobre el resultado de un terraform plan. Las políticas están definidas en [server side repo config](server-side-repo-config.md#reference).

### `--enable-profiling-api`

```bash
atlantis server --enable-profiling-api
# or
ATLANTIS_ENABLE_PROFILING_API=true
```

Habilita endpoints de [`net/http/pprof`](https://pkg.go.dev/net/http/pprof) para [continuous profiling](https://grafana.com/docs/pyroscope/latest/introduction/continuous-profiling/) de recursos usados por el servidor. Vea [profiling Go programs](https://go.dev/blog/pprof) para más información.

### `--enable-regexp-cmd`

```bash
atlantis server --enable-regexp-cmd
# or
ATLANTIS_ENABLE_REGEXP_CMD=true
```

Habilita Atlantis para usar expresiones regulares para ejecutar comandos plan/apply contra nombres de proyectos definidos cuando se pasa el flag `-p` con él.

Esto puede usarse para ejecutar todos los proyectos definidos (con la clave `name`) en `atlantis.yaml` usando `atlantis plan -p .*`.

El flag solo permitirá las regex listadas en la clave [`allowed_regexp_prefixes`](repo-level-atlantis-yaml.md#reference) definida en el archivo `atlantis.yaml` del repo. Si la clave no está definida, su valor predeterminado es `[]`, lo que permitirá cualquier regex.

Esto todavía no funcionará con `-d` y para usar `-p` los proyectos del repo deben estar definidos en el archivo `atlantis.yaml` del repo.

Cuando `--restrict-file-list` está habilitado, los plans de proyectos por regex se limitan a coincidir con proyectos con archivos modificados en el pull request. Sin `--restrict-file-list`, los comandos de proyectos por regex todavía pueden ejecutarse contra todos los proyectos coincidentes.

::: warning ADVERTENCIA DE SEGURIDAD
No se supone que se use con `--disable-apply-all`.
El comando `atlantis apply -p .*` omitirá la restricción y ejecutará apply en cada proyecto.
:::

### `--executable-name`

```bash
atlantis server --executable-name="atlantis"
# or
ATLANTIS_EXECUTABLE_NAME="atlantis"
```

Nombre del ejecutable disparador del comando de comentario. El valor predeterminado es `atlantis`.

Esto es útil cuando se ejecutan múltiples servidores Atlantis contra un solo repositorio.

Nota: la advertencia de error ortográfico "did you mean" solo aplica al nombre ejecutable predeterminado (`atlantis`). Si establece un `--executable-name` personalizado, Atlantis no advertirá sobre comentarios que sean cercanos a él según Levenshtein — esto evita advertencias espurias cuando múltiples servidores con nombres similares (p. ej. `atlantis-dev` e `atlantis-prod`) reciben el mismo comentario.

### `--fail-on-pre-workflow-hook-error`

```bash
atlantis server --fail-on-pre-workflow-hook-error
# or
ATLANTIS_FAIL_ON_PRE_WORKFLOW_HOOK_ERROR=true
```

Falla y no ejecuta el comando solicitado de Atlantis si alguno de los pre workflow hooks produce un error.

### `--gh-allow-mergeable-bypass-apply`

```bash
atlantis server --gh-allow-mergeable-bypass-apply
# or
ATLANTIS_GH_ALLOW_MERGEABLE_BYPASS_APPLY=true
```

Feature flag para habilitar la capacidad de usar el modo `mergeable` con required apply status check.

### `--gh-app-id`

```bash
atlantis server --gh-app-id="00000"
# or
ATLANTIS_GH_APP_ID="00000"
```

ID de GitHub app. Si está configurado, la autenticación de GitHub se realizará como [an installation](https://docs.github.com/en/rest/apps/installations).

::: tip
Un GitHub app puede crearse iniciando primero Atlantis, luego apuntando su navegador a

```shell
$(hostname)/github-app/setup
```

Será redirigido a GitHub para crear una nueva app, y luego será redirigido a

```shell
$(hostname)/github-app/exchange-code?code=some-code
```

Después de lo cual Atlantis mostrará las credenciales de su nueva app: el ID de su app, su `--gh-webhook-secret` generado y el contenido del archivo para `--gh-app-key-file`. Actualice su configuración de Atlantis en consecuencia y reinicie el servidor.
:::

### `--gh-app-installation-id`

```bash
atlantis server --gh-app-installation-id="123"
# or
ATLANTIS_GH_APP_INSTALLATION_ID="123"
```

El installation ID de una instancia específica de una aplicación GitHub. Normalmente este valor se
deriva consultando a GitHub la lista de installations del ID suministrado mediante `--gh-app-id` y seleccionando
la primera encontrada, y múltiples resultados de installations producen un error. Use este flag si tiene múltiples
instancias de Atlantis pero quiere usar una sola GitHub app ya instalada para todas ellas. Normalmente haría esto si
está ejecutando un proxy como su única aplicación GitHub que hará proxy a una instancia de Atlantis apropiada
según la organización o usuario que disparó el webhook.

### `--gh-app-key`

```bash
atlantis server --gh-app-key="-----BEGIN RSA PRIVATE KEY-----(...)"
# or
ATLANTIS_GH_APP_KEY="-----BEGIN RSA PRIVATE KEY-----(...)"
```

La private key codificada en PEM para la GitHub App.

::: warning ADVERTENCIA DE SEGURIDAD
El contenido de la private key será visible para cualquiera que pueda ejecutar `ps` o mirar el historial del shell de la máquina donde se está ejecutando Atlantis. Use `--gh-app-key-file` para mitigar ese riesgo.
:::

### `--gh-app-key-file`

```bash
atlantis server --gh-app-key-file="path/to/app-key.pem"
# or
ATLANTIS_GH_APP_KEY_FILE="path/to/app-key.pem"
```

Ruta a un archivo de private key codificada en PEM de GitHub App. Si está configurado, la autenticación de GitHub se realizará como [an installation](https://docs.github.com/en/rest/apps/installations).

### `--gh-app-slug`

```bash
atlantis server --gh-app-slug="myappslug"
# or
ATLANTIS_GH_APP_SLUG="myappslug"
```

Una versión slug del nombre de GitHub app mostrada en comentarios de pull requests, etc. (no `Atlantis App` sino algo como `atlantis-app`). Atlantis usa el valor de este parámetro para identificar los comentarios que ha dejado en los pull requests de GitHub. Esto se usa para funciones como `--hide-prev-plan-comments`. Necesita obtener este valor de su GitHub app; una forma es ir a la configuración de su App y abrir "Public page" desde la barra lateral izquierda. Su valor `--gh-app-slug` será la última parte de la URL, p. ej. `https://github.com/apps/<slug>`.

### `--gh-hostname`

```bash
atlantis server --gh-hostname="my.github.enterprise.com"
# or
ATLANTIS_GH_HOSTNAME="my.github.enterprise.com"
```

Hostname de su instalación GitHub Enterprise. Si usa [GitHub.com](https://github.com),
no lo configure. El valor predeterminado es `github.com`.

Para GitHub Enterprise Cloud, use el hostname del tenant, por ejemplo `tenant.ghe.com`. No incluya un scheme ni un prefijo `api.`; Atlantis deriva los endpoints de API REST y GraphQL a partir del hostname.

Atlantis requiere que la clone URL de cada repositorio GitHub que recibe, en webhooks
o desde la API de GitHub, esté en este host. La comparación ignora mayúsculas/minúsculas y puertos.

::: warning Cambio de comportamiento
Anteriormente, Atlantis verificaba solo la ruta de una clone URL de GitHub (`/owner/repo.git`) y aceptaba cualquier host. Ahora el host de la clone URL tanto del repositorio base como del head debe ser igual a `--gh-hostname`, y una clone URL sin host es rechazada. Un webhook de pull request con host no coincidente se responde con HTTP 400 y el body contiene `expected clone url host "<gh-hostname>" but had "<host>"`; no se hace ningún plan. Si ejecuta GitHub Enterprise, configure `--gh-hostname` con el host que aparece en las clone URLs de sus repositorios (el campo `clone_url` en payloads de webhook), y verifique las entregas de webhook en GitHub por respuestas 400 después de actualizar.
:::

### `--gh-org`

```bash
atlantis server --gh-org="myorgname"
# or
ATLANTIS_GH_ORG="myorgname"
```

Nombre de organización de GitHub. Configúrelo para habilitar la creación de una GitHub app privada para esta organización.

### `--gh-team-allowlist`

```bash
atlantis server --gh-team-allowlist="myteam:plan, secteam:apply, devops-team:apply, devops-team:import"
# or
ATLANTIS_GH_TEAM_ALLOWLIST="myteam:plan, secteam:apply, devops-team:apply, devops-team:import"
```

En las versiones v0.35.0 y posteriores, el nombre del equipo de GitHub solo puede ser un slug porque es inmutable.

En las versiones entre v0.21.0 y v0.34.0, el nombre del equipo de GitHub puede ser un nombre o un slug.

En las versiones v0.20.1 y anteriores, el nombre del equipo de GitHub requería el nombre del equipo sensible a mayúsculas y minúsculas.

Lista separada por comas de equipos de GitHub y pares de permisos.

Por defecto, cualquier equipo puede plan y apply.

Se respeta la jerarquía de equipos de GitHub. Si un equipo en allowlist tiene equipos hijo, los miembros de esos equipos hijo heredan los comandos permitidos del equipo padre.

::: tip
Si está usando [policy checking](policy-checking.md), también debe incluir el comando `policy_check` en la allowlist para que funcione en comandos manuales `atlantis plan`:

```bash
atlantis server --gh-team-allowlist="*:plan,*:policy_check,myteam:apply"
```

Vea [Policy Checking documentation](policy-checking.md#step-1-enable-the-workflow) para más detalles.
:::

### `--gh-token`

```bash
atlantis server --gh-token="token"
# or (recommended)
ATLANTIS_GH_TOKEN="token"
```

Token de GitHub del usuario de API.

### `--gh-token-file`

```bash
atlantis server --gh-token-file="/path/to/token"
# or
ATLANTIS_GH_TOKEN_FILE="/path/to/token"
```

Token de GitHub del usuario de API. El token se carga regularmente desde disco para permitir la rotación del token sin necesidad de reiniciar el servidor Atlantis.

### `--gh-user`

```bash
atlantis server --gh-user="myuser"
# or
ATLANTIS_GH_USER="myuser"
```

Nombre de usuario de GitHub del usuario de API. Este usuario también es usado por el flag `--hide-user-plan-comments` y necesitará actualizarse si migra a github EMU.

### `--gh-webhook-secret`

```bash
atlantis server --gh-webhook-secret="secret"
# or (recommended)
ATLANTIS_GH_WEBHOOK_SECRET="secret"
```

Secreto usado para validar webhooks de GitHub (vea [GitHub: Validating webhook deliveries](https://docs.github.com/en/webhooks/using-webhooks/validating-webhook-deliveries)).

::: warning ADVERTENCIA DE SEGURIDAD
Si no se especifica, Atlantis no podrá validar que la llamada webhook entrante provino de GitHub.
Esto significa que un atacante podría falsificar llamadas a Atlantis y hacer que realice acciones maliciosas.
:::

### `--gitea-base-url`

```bash
atlantis server --gitea-base-url="http://your-gitea.corp:7990/basepath"
# or
ATLANTIS_GITEA_BASE_URL="http://your-gitea.corp:7990/basepath"
```

URL base de la instalación de Gitea. Debe incluir `http://` o `https://`. El valor predeterminado es `https://gitea.com` si se deja vacío/ausente.

### `--gitea-page-size`

```bash
atlantis server --gitea-page-size=30
# or (recommended)
ATLANTIS_GITEA_PAGE_SIZE=30
```

Número de elementos en una sola página en respuestas paginadas de Gitea.

::: warning Dependiente de la configuración
El valor predeterminado cumple con la configuración estándar del servidor Gitea: DEFAULT_PAGING_NUM
El valor válido más alto depende de la configuración del servidor Gitea: MAX_RESPONSE_ITEMS
:::

### `--gitea-token`

```bash
atlantis server --gitea-token="token"
# or (recommended)
ATLANTIS_GITEA_TOKEN="token"
```

App password de Gitea del usuario de API.

### `--gitea-user`

```bash
atlantis server --gitea-user="myuser"
# or
ATLANTIS_GITEA_USER="myuser"
```

Nombre de usuario de Gitea del usuario de API.

### `--gitea-webhook-secret`

```bash
atlantis server --gitea-webhook-secret="secret"
# or (recommended)
ATLANTIS_GITEA_WEBHOOK_SECRET="secret"
```

Secreto usado para validar webhooks de Gitea.

::: warning ADVERTENCIA DE SEGURIDAD
Si no se especifica, Atlantis no podrá validar que la llamada webhook entrante provino de Gitea.
Esto significa que un atacante podría falsificar llamadas a Atlantis y hacer que realice acciones maliciosas.
:::

### `--gitlab-group-allowlist`

```bash
atlantis server --gitlab-group-allowlist="myorg/mygroup:plan, myorg/secteam:apply, myorg/devops:apply, myorg/devops:import"
# or
ATLANTIS_GITLAB_GROUP_ALLOWLIST="myorg/mygroup:plan, myorg/secteam:apply, myorg/devops:apply, myorg/devops:import"
```

Lista separada por comas de grupos de GitLab y pares de permisos.

Por defecto, cualquier grupo puede plan y apply.

::: warning NOTA
Atlantis necesita poder ver los miembros del grupo listado; los grupos inaccesibles o inexistentes se ignoran silenciosamente.
:::

### `--gitlab-hostname`

```bash
atlantis server --gitlab-hostname="my.gitlab.enterprise.com"
# or
ATLANTIS_GITLAB_HOSTNAME="my.gitlab.enterprise.com"
```

Hostname de su instalación GitLab Enterprise. Si usa [GitLab.com](https://gitlab.com),
no lo configure. El valor predeterminado es `gitlab.com`.

### `--gitlab-status-retry-enabled`

```bash
atlantis server --gitlab-status-retry-enabled
# or
ATLANTIS_GITLAB_STATUS_RETRY_ENABLED=true
```

Habilita lógica mejorada de reintentos para actualizaciones de estado de pipelines de GitLab con exponential backoff.

El valor predeterminado es `false`.

### `--gitlab-token`

```bash
atlantis server --gitlab-token="token"
# or (recommended)
ATLANTIS_GITLAB_TOKEN="token"
```

Token de GitLab del usuario de API.

### `--gitlab-user`

```bash
atlantis server --gitlab-user="myuser"
# or
ATLANTIS_GITLAB_USER="myuser"
```

Nombre de usuario de GitLab del usuario de API.

### `--gitlab-webhook-secret`

```bash
atlantis server --gitlab-webhook-secret="secret"
# or (recommended)
ATLANTIS_GITLAB_WEBHOOK_SECRET="secret"
```

Secreto usado para validar webhooks de GitLab.

::: warning ADVERTENCIA DE SEGURIDAD
Si no se especifica, Atlantis no podrá validar que la llamada webhook entrante provino de GitLab.
Esto significa que un atacante podría falsificar llamadas a Atlantis y hacer que realice acciones maliciosas.
:::

### `--help`

```bash
atlantis server --help
```

Ver ayuda.

### `--hide-prev-plan-comments`

```bash
atlantis server --hide-prev-plan-comments
# or
ATLANTIS_HIDE_PREV_PLAN_COMMENTS=true
```

Oculta comentarios previos de plan para reducir el desorden en los PR. Actualmente esto solo está soportado en
GitHub, GitLab y Bitbucket y no está habilitado por defecto.

Para Bitbucket, los comentarios se eliminan en lugar de ocultarse ya que Bitbucket no soporta ocultar comentarios.

Para GitHub, asegúrese de que `--gh-user` esté configurado apropiadamente o los comentarios no se ocultarán.

Al usar GitHub App, necesita establecer `--gh-app-slug` para habilitar esta funcionalidad.

### `--hide-unchanged-plan-comments`

```bash
atlantis server --hide-unchanged-plan-comments
# or
ATLANTIS_HIDE_UNCHANGED_PLAN_COMMENTS=true
```

Elimina comentarios de plan sin cambios del pull request.

Esto es útil cuando tiene muchos proyectos y quiere mantener el pull request limpio de comentarios inútiles.

### `--ignore-vcs-status-names`

```bash
atlantis server --ignore-vcs-status-names="status1,status2"
# or
ATLANTIS_IGNORE_VCS_STATUS_NAMES=status1,status2
```

Lista separada por comas de nombres de estado VCS de otros servicios atlantis.
Cuando `gh-allow-mergeable-bypass-apply` es true, ignorará verificaciones de estado
(p. ej. `status1/plan`, `status1/apply`, `status2/plan`, `status2/apply`)
de otros servicios Atlantis al verificar si el PR es mergeable.
Actualmente solo implementado para GitHub.

### `--include-git-untracked-files`

```bash
atlantis server --include-git-untracked-files
# or
ATLANTIS_INCLUDE_GIT_UNTRACKED_FILES=true
```

Incluye archivos git no rastreados en la lista de archivos modificados de Atlantis.
Usado por ejemplo con pre-workflow hooks de [CDK Terrain](custom-workflows.md#cdk-terrain-cdktn) que
generan dinámicamente archivos Terraform.

### `--language`

```bash
atlantis server --language="en"
# or
ATLANTIS_LANGUAGE="en"
```

Idioma usado para comentarios de pull request de Atlantis. El valor predeterminado es `en`.

Valores soportados:

- `en` (Inglés)
- `es` (Español)

Atlantis normaliza valores de estilo locale (por ejemplo `es-MX` se resuelve a `es`).
Si se configura un idioma no soportado y `--language-config-file` no está configurado,
Atlantis devuelve un error de validación al iniciar.

Las cadenas de idioma integradas se cargan desde:

- `server/i18n/locales/en.yaml`
- `server/i18n/locales/es.yaml`

### `--language-config-file`

```bash
atlantis server --language="de" --language-config-file="/etc/atlantis/language.yaml"
# or
ATLANTIS_LANGUAGE_CONFIG_FILE="/etc/atlantis/language.yaml"
```

Ruta opcional a un catálogo de idioma YAML personalizado. Los valores en este archivo sobrescriben
el idioma integrado seleccionado, y se soportan sobrescrituras parciales.

Cuando `--language-config-file` está configurado, se permiten valores `--language` no soportados
y Atlantis vuelve al inglés integrado para cadenas no especificadas.

Esquema YAML esperado:

```yaml
pull_request_label: Pull Request (custom)
merge_request_label: Merge Request (custom)
command_titles:
  plan: Plan (custom)
  apply: Apply (custom)
```

Para una personalización completa del texto markdown, continúe usando
`--markdown-template-overrides-dir`.

### `--locking-db-type`

```bash
atlantis server --locking-db-type="<boltdb|redis>"
# or
ATLANTIS_LOCKING_DB_TYPE="<boltdb|redis>"
```

El tipo de base de datos de locking a usar para almacenar locks de plan y apply. El valor predeterminado es `boltdb`.

Notas:

- Si se establece en `boltdb`, solo un proceso puede tener acceso a la instancia boltdb.
- Si se establece en `redis`, use `--redis-host` e `--redis-port` para modo de nodo único, o `--redis-cluster-addresses` para modo Redis Cluster. Use `--redis-password` y (opcionalmente) `--redis-username` solo si su despliegue Redis requiere autenticación.

### `--log-level`

```bash
atlantis server --log-level="<debug|info|warn|error>"
# or
ATLANTIS_LOG_LEVEL="<debug|info|warn|error>"
```

Nivel de log. El valor predeterminado es `info`.

### `--markdown-template-overrides-dir`

```bash
atlantis server --markdown-template-overrides-dir="path/to/templates/"
# or
ATLANTIS_MARKDOWN_TEMPLATE_OVERRIDES_DIR="path/to/templates/"
```

Esto estará disponible en v0.21.0.

Directorio donde Atlantis leerá sobrescrituras para plantillas markdown usadas para renderizar comentarios en pull requests.
Las sobrescrituras de plantillas markdown pueden especificarse ya sea en archivos individuales, o todas juntas en un solo archivo. Todos los archivos de
sobrescritura de plantillas _deben_ tener la extensión `.tmpl`, de lo contrario no serán parseados.

Las plantillas markdown que pueden tener sobrescrituras se pueden encontrar en el [directorio de plantillas markdown](https://github.com/runatlantis/atlantis/tree/main/server/events/templates)

Tenga presente que configuraciones como `--enable-diff-markdown-format` dependen de la lógica definida en las plantillas. Es
posible divergir del comportamiento esperado si no se tiene cuidado al sobrescribir las plantillas predeterminadas.

El valor predeterminado es el directorio home de atlantis `/home/atlantis/.markdown_templates/` en `/$HOME/.markdown_templates`.

### `--max-comments-per-command`

```bash
atlantis server --max-comments-per-command=100
# or
ATLANTIS_MAX_COMMENTS_PER_COMMAND=100
```

Limita el número de comentarios publicados después de ejecutar un comando, para evitar hacer spam a su VCS y que Atlantis sea limitado como resultado. El valor predeterminado es `100`. Establezca esta opción en `0` para deshabilitar el truncado de logs. Tenga en cuenta que el truncado ocurrirá en la parte superior de la salida del comando, para preservar las partes más importantes de la salida, a menudo mostradas al final.

Cuando la salida del comando excede el límite de tamaño de comentarios del VCS (o cuando aplica este límite), Atlantis divide la salida en múltiples comentarios usando **división inteligente de comentarios**. Los puntos de división se eligen para que la estructura markdown se preserve: el divisor detecta si está dentro de un bloque de código (``

 ``` ``), a `<details>` block, or inline code (`` ` ``), and inserts appropriate closing and continuation markers so that each comment renders correctly. Continuation comments are labeled with the command name (e.g. "Continued plan output from previous comment") when available.

### `--parallel-apply`

```bash
atlantis server --parallel-apply
# or
ATLANTIS_PARALLEL_APPLY=true
```

Si ejecutar operaciones apply en paralelo. El valor predeterminado es `false`. La declaración explícita en [repo config](repo-level-atlantis-yaml.md#run-plans-and-applies-in-parallel) tiene precedencia.

### `--parallel-plan`

```bash
atlantis server --parallel-plan
# or
ATLANTIS_PARALLEL_PLAN=true
```

Si ejecutar operaciones plan en paralelo. El valor predeterminado es `false`. La declaración explícita en [repo config](repo-level-atlantis-yaml.md#run-plans-and-applies-in-parallel) tiene precedencia.

### `--parallel-pool-size`

```bash
atlantis server --parallel-pool-size=100
# or
ATLANTIS_PARALLEL_POOL_SIZE=100
```

Tamaño máximo del wait group que ejecuta plans y applies en paralelo (si está habilitado). El valor predeterminado es `15`

El límite es por comando: también limita los proyectos de una solicitud de API plan o detección de drift que hacen plan al mismo tiempo.

### `--pending-apply-status`

```bash
atlantis server --pending-apply-status
# or (recommended)
ATLANTIS_PENDING_APPLY_STATUS=true
```

Establece el estado del commit en pending cuando hay cambios planificados que no se han aplicado.
Esto evita que los merge requests se fusionen hasta que todos los Terraform apply se completen si tiene `Pipelines must succeed` habilitado en su repositorio.

Cuando está habilitado, después de ejecutar `atlantis plan`, el estado del MR se mostrará como pending si hay cambios
para aplicar. Una vez que todos los proyectos se aplican correctamente (o no muestran cambios), el estado se actualizará a success.
Los proyectos sin cambios de Terraform se cuentan como actualizados en lugar de aplicados. Si un pull request tiene tanto
proyectos actualizados como proyectos todavía esperando apply, el estado de commit apply de Atlantis permanece pending
hasta que todos los proyectos cambiados se apliquen.

El valor predeterminado es `false`.

Solo soportado en GitLab

### `--port`

```bash
atlantis server --port=4141
# or
ATLANTIS_PORT=4141
```

Puerto al que enlazarse. El valor predeterminado es `4141`.

### `--quiet-policy-checks`

```bash
atlantis server --quiet-policy-checks
# or
ATLANTIS_QUIET_POLICY_CHECKS=true
```

Excluye comentarios de policy check de pull requests a menos que haya un error real de conftest. Esto también excluye advertencias. El valor predeterminado es `false`.

### `--redis-cluster-addresses`

```bash
atlantis server --redis-cluster-addresses="redis-node-0:6379,redis-node-1:6379,redis-node-2:6379"
# or
ATLANTIS_REDIS_CLUSTER_ADDRESSES="redis-node-0:6379,redis-node-1:6379,redis-node-2:6379"
```

Lista delimitada por comas de direcciones de nodos del clúster Redis en el formato `host:port`. Cuando está configurado, Atlantis usa modo Redis Cluster en lugar de modo de nodo único. Esto es mutuamente excluyente con `--redis-host`/`--redis-port` (que se usan para modo de nodo único).

### `--redis-db`

```bash
atlantis server --redis-db=0
# or
ATLANTIS_REDIS_DB=0
```

La base de datos Redis a usar cuando se usa un tipo de Locking DB de `redis`. El valor predeterminado es `0`.

### `--redis-host`

```bash
atlantis server --redis-host="localhost"
# or
ATLANTIS_REDIS_HOST="localhost"
```

El hostname Redis para cuando se usa un tipo de Locking DB de `redis`.

### `--redis-insecure-skip-verify`

```bash
atlantis server --redis-insecure-skip-verify=false
# or
ATLANTIS_REDIS_INSECURE_SKIP_VERIFY=false
```

Controla si el cliente Redis verifica la cadena de certificados y el nombre de host del servidor Redis. Si es true, acepta cualquier certificado presentado por el servidor y cualquier nombre de host en ese certificado. El valor predeterminado es `false`.

::: warning ADVERTENCIA DE SEGURIDAD
Si esto está habilitado, TLS es susceptible a ataques de intermediario a menos que se use verificación personalizada.
:::

### `--redis-password`

```bash
atlantis server --redis-password="password123"
# or (recommended)
ATLANTIS_REDIS_PASSWORD="password123"
```

La contraseña Redis para cuando se usa un tipo de Locking DB de `redis`.

### `--redis-port`

```bash
atlantis server --redis-port=6379
# or
ATLANTIS_REDIS_PORT=6379
```

El puerto Redis para cuando se usa un tipo de Locking DB de `redis`. El valor predeterminado es `6379`.

### `--redis-tls-enabled`

```bash
atlantis server --redis-tls-enabled=false
# or
ATLANTIS_REDIS_TLS_ENABLED=false
```

Habilita una conexión TLS, con versión mínima 1.2, a Redis cuando se usa un tipo de Locking DB de `redis`. El valor predeterminado es `false`.

### `--redis-username`

```bash
atlantis server --redis-username="myuser"
# or
ATLANTIS_REDIS_USERNAME="myuser"
```

El nombre de usuario Redis para cuando se usa un tipo de Locking DB de `redis`. Útil cuando Redis está configurado con autenticación basada en ACL.

### `--repo-allowlist`

```bash
# NOTE: Use single quotes to avoid shell expansion of *.
atlantis server --repo-allowlist='github.com/myorg/*'
# or
ATLANTIS_REPO_ALLOWLIST='github.com/myorg/*'
```

Atlantis requiere que especifique una allowlist de repositorios de los que aceptará webhooks.

Notas:

- Acepta una lista separada por comas, p. ej. `definition1,definition2`
- El formato es `{hostname}/{owner}/{repo}`, p. ej. `github.com/runatlantis/atlantis`
- `*` coincide con cualquier carácter, p. ej. `github.com/runatlantis/*` coincidirá con todos los repos en la organización runatlantis
- Una entrada que comienza con `!` la niega, p. ej. `github.com/foo/*,!github.com/foo/bar` coincidirá con todos los repos github en el owner `foo` _excepto_ `bar`.
- Para Bitbucket Server: `{hostname}` es el dominio sin scheme ni puerto, `{owner}` es el nombre del proyecto (no la key), e `{repo}` es el nombre del repo
  - Los repositorios de usuario (no proyecto) tienen el formato: `{hostname}/{full name}/{repo}` (p. ej., `bitbucket.example.com/Jane Doe/myatlantis` para el nombre de usuario `jdoe` y nombre completo `Jane Doe`, lo cual no es muy intuitivo)
- Para Azure DevOps la allowlist toma una de dos formas: `{owner}.visualstudio.com/{owner}/{project}/{repo}` o `dev.azure.com/{owner}/{project}/{repo}`
- Microsoft está en proceso de cambiar Azure DevOps a la segunda forma, por lo que puede ser más seguro especificar siempre ambos formatos en su allowlist de repos para cada repositorio hasta que el cambio esté completo.

Ejemplos:

- Allowlist `myorg/repo1` e `myorg/repo2` en `github.com`
  - `--repo-allowlist=github.com/myorg/repo1,github.com/myorg/repo2`
- Allowlist de todos los repos bajo `myorg` en `github.com`
  - `--repo-allowlist='github.com/myorg/*'`
- Allowlist de todos los repos bajo `myorg` en `github.com`, excluyendo `myorg/untrusted-repo`
  - `--repo-allowlist='github.com/myorg/*,!github.com/myorg/untrusted-repo'`
- Allowlist de todos los repos en mi instalación GitHub Enterprise
  - `--repo-allowlist='github.yourcompany.com/*'`
- Allowlist de todos los repositorios bajo el proyecto `myproject` de `myorg` en Azure DevOps
  - `--repo-allowlist='myorg.visualstudio.com/myorg/myproject/*,dev.azure.com/myorg/myproject/*'`
- Allowlist de todos los repositorios
  - `--repo-allowlist='*'`

### `--repo-config`

```bash
atlantis server --repo-config="path/to/repos.yaml"
# or
ATLANTIS_REPO_CONFIG="path/to/repos.yaml"
```

Ruta a un archivo de configuración YAML de repo del lado del servidor. Vea [Server Side Repo Config](server-side-repo-config.md).

### `--repo-config-json`

```bash
atlantis server --repo-config-json='{"repos":[{"id":"/.*/", "apply_requirements":["mergeable"]}]}'
# or
ATLANTIS_REPO_CONFIG_JSON='{"repos":[{"id":"/.*/", "apply_requirements":["mergeable"]}]}'
```

Especifique la config de repo del lado del servidor como una cadena JSON. Útil si no quiere escribir un archivo de configuración en disco.
Vea [Server Side Repo Config](server-side-repo-config.md) para más detalles.

::: tip
Si especifica un [Workflow](custom-workflows.md#reference), los [step](custom-workflows.md#step)
pueden especificarse como sigue:

```json
{
   "repos": [],
   "workflows": {
      "custom": {
         "plan": {
            "steps": [
               "init",
               {
                  "plan": {
                     "extra_args": ["extra", "args"]
                  }
               },
               {
                  "run": "my custom command"
               }
            ]
         }
      }
   }
}
```

:::

### `--restrict-file-list`

```bash
atlantis server --restrict-file-list
# or (recommended)
ATLANTIS_RESTRICT_FILE_LIST=true
```

`--restrict-file-list` bloqueará solicitudes de plan de proyectos fuera de los archivos modificados en el pull request.
Cuando `--enable-regexp-cmd` también está habilitado, los plans de proyectos por regex como `atlantis plan -p .*` se limitan a proyectos coincidentes con archivos modificados en el pull request.
El valor predeterminado es `false`.

### `--share-plan-dir`

```bash
atlantis server --share-plan-dir="path/to/local/plan/dir"
# or
ATLANTIS_SHARE_PLAN_DIR="path/to/local/plan/dir"
```

Directorio donde Atlantis almacenará archivos de plan local de Terraform. Si no está configurado, esto usa por defecto el valor resuelto de `--data-dir` para que las instalaciones existentes mantengan el mismo layout en disco. Cuando se configura a un directorio diferente, los repos checkout permanecen bajo `--data-dir` y los archivos `.tfplan` generados usan el mismo layout de ruta de repo, pull request, workspace y proyecto bajo `--share-plan-dir`.

### `--silence-allowlist-errors`

```bash
atlantis server --silence-allowlist-errors
# or
ATLANTIS_SILENCE_ALLOWLIST_ERRORS=true
```

Algunos usuarios usan el flag `--repo-allowlist` para controlar a qué repos Atlantis
responde. Normalmente, si Atlantis recibe un webhook de pull request de un repo no listado
en la allowlist, responderá con un comentario de error. Este flag deshabilita ese comentario.

Algunos usuarios encuentran esto útil porque prefieren agregar el webhook de Atlantis
a nivel de organización en lugar de en cada repo.

### `--silence-fork-pr-errors`

```bash
atlantis server --silence-fork-pr-errors
# or
ATLANTIS_SILENCE_FORK_PR_ERRORS=true
```

Normalmente, si Atlantis recibe un webhook de pull request de un fork y --allow-fork-prs no está configurado,
responderá con un comentario de error. Este flag deshabilita ese comentario.

### `--silence-no-projects`

```bash
atlantis server --silence-no-projects
# or
ATLANTIS_SILENCE_NO_PROJECTS=true
```

`--silence-no-projects` le dirá a Atlantis que ignore PRs si ninguno de los archivos modificados es parte de un proyecto definido en el archivo `atlantis.yaml`.
Este flag asegura que un servidor Atlantis solo responda a sus proyectos declarados explícitamente.
Esto no tiene efecto si los proyectos no están definidos en `atlantis.yaml` a nivel de repo.
Esto también silencia comandos dirigidos (p. ej. `atlantis plan -d mydir` o `atlantis apply -p myproj`), por lo que si el proyecto no está en el archivo de configuración del repo `atlantis.yaml`, estos comandos no se ejecutarán ni reportarán en un comentario.

Esto es útil cuando se ejecutan múltiples servidores Atlantis contra un solo repositorio, para que pueda
delegar trabajo a cada servidor Atlantis. También es útil cuando se usa con pre_workflow_hooks para generar dinámicamente un archivo `atlantis.yaml`.

### `--silence-vcs-status-no-plans`

```bash
atlantis server --silence-vcs-status-no-plans
# or
ATLANTIS_SILENCE_VCS_STATUS_NO_PLANS=true
```

`--silence-vcs-status-no-plans` le dirá a Atlantis que ignore establecer estado VCS en plans si ninguno de los archivos modificados es parte de un proyecto definido en el archivo `atlantis.yaml`.

### `--silence-vcs-status-no-projects`

```bash
atlantis server --silence-vcs-status-no-projects
# or
ATLANTIS_SILENCE_VCS_STATUS_NO_PROJECTS=true
```

`--silence-vcs-status-no-projects` le dirá a Atlantis que ignore establecer estado VCS en cualquier comando si ninguno de los archivos modificados es parte de un proyecto definido en el archivo `atlantis.yaml`.

### `--skip-clone-no-changes`

```bash
atlantis server --skip-clone-no-changes
# or
ATLANTIS_SKIP_CLONE_NO_CHANGES=true
```

`--skip-clone-no-changes` omitirá clonar el repo durante autoplan si no hay cambios en proyectos Terraform. Esto solo se aplicará para GitHub y GitLab y solo para repos que tengan archivo `atlantis.yaml`. El valor predeterminado es `false`.

### `--slack-token`

```bash
atlantis server --slack-token=token
# or (recommended)
ATLANTIS_SLACK_TOKEN='token'
```

Token de API para notificaciones de Slack. Vea [Using Slack hooks](sending-notifications-via-webhooks.md#using-slack-hooks).

### `--ssl-cert-file`

```bash
atlantis server --ssl-cert-file="/etc/ssl/certs/my-cert.crt"
# or
ATLANTIS_SSL_CERT_FILE="/etc/ssl/certs/my-cert.crt"
```

Archivo que contiene el certificado x509 usado para servir HTTPS.
Si el cert está firmado por una CA, el archivo debe ser la concatenación
del certificado del servidor, cualquier intermedio y el certificado de la CA.

### `--ssl-key-file`

```bash
atlantis server --ssl-key-file="/etc/ssl/private/my-cert.key"
# or
ATLANTIS_SSL_KEY_FILE="/etc/ssl/private/my-cert.key"
```

Archivo que contiene la private key x509 que coincide con `--ssl-cert-file`.

### `--stats-namespace`

```bash
atlantis server --stats-namespace="myatlantis"
# or
ATLANTIS_STATS_NAMESPACE="myatlantis"
```

Namespace para emitir stats/metrics. Vea la sección [stats](stats.md).

### `--tf-distribution`

  <Badge text="Deprecated" type="warn"/>
  Obsoleto en favor de `--default-tf-distribution`.

### `--tf-download`

```bash
atlantis server --tf-download=false
# or
ATLANTIS_TF_DOWNLOAD=false
```

El valor predeterminado es `true`. Permite a Atlantis listar y descargar versiones adicionales de Terraform.
Establecer esto en `false` puede ser útil en un entorno air-gapped donde un mirror de descarga no está disponible.

### `--tf-download-url`

```bash
atlantis server --tf-download-url="https://releases.company.com"
# or
ATLANTIS_TF_DOWNLOAD_URL="https://releases.company.com"
```

Una URL alternativa para descargar versiones de Terraform si faltan. Útil en un entorno airgapped
donde releases.hashicorp.com no está disponible. La estructura de directorio del endpoint
personalizado debe coincidir con la de releases.hashicorp.com.

Esto no tiene impacto si `--tf-download` está configurado en `false`.

Esta configuración todavía no está soportada cuando `--tf-distribution` está configurado en `opentofu`.

### `--tfe-hostname`

```bash
atlantis server --tfe-hostname="my-terraform-enterprise.company.com"
# or
ATLANTIS_TFE_HOSTNAME="my-terraform-enterprise.company.com"
```

Hostname de su instalación Terraform Enterprise para usar junto con
`--tfe-token`. Vea [Terraform Cloud](terraform-cloud.md) para más detalles.
Si usa Terraform Cloud (es decir, no tiene su propia instalación Terraform Enterprise)
no necesita configurarlo ya que el valor predeterminado es `app.terraform.io`.

### `--tfe-local-execution-mode`

```bash
atlantis server --tfe-local-execution-mode
# or
ATLANTIS_TFE_LOCAL_EXECUTION_MODE=true
```

Habilite esto si está usando modo de ejecución local (en lugar del modo de ejecución remota de TFE/C). Vea [Terraform Cloud](terraform-cloud.md) para más detalles.

### `--tfe-token`

```bash
atlantis server --tfe-token="xxx.atlasv1.yyy"
# or (recommended)
ATLANTIS_TFE_TOKEN='xxx.atlasv1.yyy'
```

Un token para integración con Terraform Cloud/Terraform Enterprise. Vea [Terraform Cloud](terraform-cloud.md) para más detalles.

### `--use-tf-plugin-cache`

```bash
atlantis server --use-tf-plugin-cache=false
```

Establézcalo en false si quiere deshabilitar la caché de plugins de terraform.

Este flag es útil cuando hay múltiples proyectos que necesitan ejecutar plan y apply en el mismo PR para evitar la condición de carrera de `plugin_cache_dir` concurrentemente; este es un issue conocido de terraform, más información:

- [discusión concurrente de plugin_cache_dir](https://github.com/hashicorp/terraform/issues/31964)
- [PR para mejorar la situación](https://github.com/hashicorp/terraform/pull/33479)

El efecto de la condición de carrera es más evidente al usar configuración paralela para ejecutar plan y apply. Deshabilitar el uso de caché de plugins impactará el rendimiento al iniciar un nuevo plan o apply, pero en despliegues grandes de Atlantis con múltiples proyectos y módulos compartidos el uso de `--parallel_plan` e `--parallel_apply` es obligatorio para una gestión eficiente de los PRs.

### `--var-file-allowlist`

```bash
atlantis server --var-file-allowlist='/path/to/tfvars/dir'
# or
ATLANTIS_VAR_FILE_ALLOWLIST='/path/to/tfvars/dir'
```

Lista separada por comas de rutas de directorio adicionales desde las que se pueden leer [archivos de definición de variables](https://developer.hashicorp.com/terraform/language/values/variables#variable-definitions-tfvars-files).
Las rutas en este argumento deben ser rutas absolutas. Las rutas relativas y el globbing actualmente no están soportados.
Si este argumento no se proporciona, el valor predeterminado es el directorio de datos de Atlantis, determinado por el argumento `--data-dir`.

Notas:

- La allowlist se aplica para cada comando de comentario que pasa `-var-file` a Terraform (`plan`, `apply`, `import` e `state rm`), y para ambas grafías `-var-file` e `--var-file`.
- Solo se verifican los valores `-var-file` dados en argumentos extra de comentarios. Los archivos var nombrados en `extra_args` del workflow no.
- Una ruta dentro del repositorio siempre está permitida tal como se escribe, pero una vez que el repositorio es clonado el archivo se resuelve siguiendo enlaces simbólicos. Un enlace confirmado en el repositorio que apunta fuera tanto del repositorio como de los directorios en allowlist es rechazado, y también lo es un enlace cuyo destino no existe. Una ruta de archivo var no puede contener un elemento `..`, porque el sistema operativo sigue un enlace simbólico antes de aplicar `..`.

::: warning Cambio de comportamiento
Anteriormente, la allowlist se verificaba solo para `atlantis plan`, solo para la grafía de guion simple `-var-file`, y solo contra la ruta tal como estaba escrita. Ahora se verifica para `plan`, `apply`, `import` e `state rm`, para cada grafía (`-var-file`, `--var-file`, con un espacio o `=`), y de nuevo después del clone con enlaces simbólicos resueltos. Las rutas rechazadas fallan con `var file path <path> is not allowed by the current allowlist: [<dirs>]`, `var file path <path> resolves to <target>, which is not allowed by the current allowlist: [<dirs>]`, `var file path <path> is a symbolic link whose target does not exist` o `var file path <path> must not contain '..' elements`. Mueva los archivos var compartidos al repositorio o a un directorio en allowlist, y reemplace los enlaces que apuntan fuera de ellos por archivos reales.
:::

### `--vcs-status-name`

```bash
atlantis server --vcs-status-name="atlantis-dev"
# or
ATLANTIS_VCS_STATUS_NAME="atlantis-dev"
```

Nombre usado para identificar Atlantis al actualizar un estado de pull request. El valor predeterminado es `atlantis`.

Esto es útil cuando se ejecutan múltiples servidores Atlantis contra un solo repositorio para que pueda
dar a cada servidor Atlantis su propio nombre único para evitar que los estados colisionen.

### `--web-basic-auth`

```bash
atlantis server --web-basic-auth
# or
ATLANTIS_WEB_BASIC_AUTH=true
```

Habilita Basic Authentication en el servicio web de Atlantis.

### `--web-password`

```bash
atlantis server --web-password="atlantis"
# or
ATLANTIS_WEB_PASSWORD="atlantis"
```

Contraseña usada para Basic Authentication en el servicio web de Atlantis. El valor predeterminado es `atlantis`.

### `--web-username`

```bash
atlantis server --web-username="atlantis"
# or
ATLANTIS_WEB_USERNAME="atlantis"
```

Nombre de usuario usado para Basic Authentication en el servicio web de Atlantis. El valor predeterminado es `atlantis`.

### `--webhook-http-headers`

```bash
atlantis server --webhook-http-headers='{"Authorization":"Bearer some-token","X-Custom-Header":["value1","value2"]}'
# or
ATLANTIS_WEBHOOK_HTTP_HEADERS='{"Authorization":"Bearer some-token","X-Custom-Header":["value1","value2"]}'
```

Headers adicionales agregados a cada payload HTTP POST al usar [http webhooks](sending-notifications-via-webhooks.md#using-http-webhooks)
proporcionados como una cadena JSON. La key del mapa es el nombre del header y el valor es el valor del header
(cadena) o valores (array de cadenas).

### `--websocket-check-origin`

```bash
atlantis server --websocket-check-origin
# or
ATLANTIS_WEBSOCKET_CHECK_ORIGIN=true
```

Solo permite conexión websockets cuando se originan desde el servidor web Atlantis en ejecución

### `--write-git-creds`

```bash
atlantis server --write-git-creds
# or
ATLANTIS_WRITE_GIT_CREDS=true
```

Escribe un archivo .git-credentials con el usuario y token del proveedor para permitir
clonar módulos privados sobre HTTPS o SSH. Vea [Git Credential Store documentation](https://git-scm.com/docs/git-credential-store) para más información.

Siga el `git::ssh`
