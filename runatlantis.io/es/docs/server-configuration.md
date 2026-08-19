# Configuración del servidor

Esta página explica cómo configurar el comando `atlantis server`.

La configuración para `atlantis server` se puede especificar mediante flags de línea de comandos,
variables de entorno, un archivo de configuración o una mezcla de los tres.

## Variables de entorno

Todos los flags se pueden especificar como variables de entorno.

1. Tome el nombre del flag, ej. `--gh-user`
1. Ignore el primer `--` => `gh-user`
1. Convierta los `-` en `_` => `gh_user`
1. Ponga todas las letras en mayúsculas => `GH_USER`
1. Prefije con `ATLANTIS_` => `ATLANTIS_GH_USER`

::: warning NOTE
Para establecer un flag booleano use `true` o `false` como valor.
:::

::: warning NOTE
El flag `--atlantis-url` se establece mediante la variable de entorno `ATLANTIS_ATLANTIS_URL` **NO** `ATLANTIS_URL`.
:::

## Archivo de configuración

Todos los flags también se pueden especificar mediante un archivo de configuración YAML.

Para usar un archivo de configuración YAML, ejecute `atlantis server --config /path/to/config.yaml`.

Las claves de su archivo de configuración deben ser las mismas que los nombres de los flags, ej.

```yaml
gh-token: ...
log-level: ...
```

::: warning
El archivo de configuración que pasa a `--config` es diferente del archivo `--repo-config`.
El archivo de configuración de `--config` solo se usa como una forma alternativa de establecer los flags de `atlantis server`.
:::

## Precedencia

Los valores se eligen en este orden:

1. Flags
1. Variables de entorno
1. Archivo de configuración

## Flags

### `--allow-commands` <Badge text="v0.27.0+" type="info"/>

```bash
atlantis server --allow-commands=version,plan,apply,unlock,approve_policies,cancel
# or
ATLANTIS_ALLOW_COMMANDS='version,plan,apply,unlock,approve_policies,cancel'
```

Lista de comandos permitidos para ejecutarse en el servidor Atlantis, por defecto es `version,plan,apply,unlock,approve_policies,cancel`

Notas:

- Acepta una lista separada por comas, ej. `command1,command2`.
- `version`, `plan`, `apply`, `unlock`, `approve_policies`, `cancel`, `import`, `state`, `policy_check` y `all` están disponibles.
- `policy_check` es un comando interno que se ejecuta automáticamente después de `plan` cuando [policy checking](policy-checking.md) está habilitado. Debe permitirse explícitamente al usar [`--gh-team-allowlist`](#gh-team-allowlist).
- `all` es una palabra clave especial que permite todos los comandos. Si pasa `all`, entonces todos los demás comandos serán ignorados.

### `--allow-draft-prs` <Badge text="v0.13.0" type="info"/>

```bash
atlantis server --allow-draft-prs
# or
ATLANTIS_ALLOW_DRAFT_PRS=true
```

Responder a pull requests desde draft prs. Por defecto es `false`.

### `--allow-fork-prs` <Badge text="v0.3.1+" type="info"/>

```bash
atlantis server --allow-fork-prs
# or
ATLANTIS_ALLOW_FORK_PRS=true
```

Responder a pull requests desde forks. Por defecto es `false`.

:::warning SECURITY WARNING
Potencialmente peligroso de habilitar
porque si atacantes pueden crear un pull request a su repo entonces pueden hacer que Atlantis
ejecute código arbitrario. Esto puede ocurrir porque
Atlantis ejecutará automáticamente `terraform plan`
que puede ejecutar código arbitrario si se le da una configuración de Terraform maliciosa.
:::

### `--api-secret` <Badge text="v0.22.2+" type="info"/>

```bash
atlantis server --api-secret="secret"
# or (recommended)
ATLANTIS_API_SECRET="secret"
```

Secreto requerido usado para validar solicitudes realizadas a los endpoints de [`/api/*`](api-endpoints.md).

### `--atlantis-url` <Badge text="v0.1.3+" type="info"/>

```bash
atlantis server --atlantis-url="https://my-domain.com:9090/basepath"
# or
ATLANTIS_ATLANTIS_URL=https://my-domain.com:9090/basepath
```

Especifique la URL desde la que Atlantis es accesible. Se usa en la UI de Atlantis
y en enlaces desde comentarios de pull request. Por defecto es `http://$(hostname):$port`
donde `$port` proviene del flag [`--port`](#port). Soporta un basepath si está alojando Atlantis bajo una ruta.

Notas:

- Si se usa un load balancer con un puerto no http/https (no el definido en el flag `--port`), actualice la URL para incluir el puerto como en el ejemplo anterior.
- Esta URL se usa como el enlace `details` junto a cada job de atlantis para ver los logs del job.

### `--autodiscover-mode` <Badge text="v0.27.0+" type="info"/>

```bash
atlantis server --autodiscover-mode="<auto|enabled|disabled>"
# or
ATLANTIS_AUTODISCOVER_MODE="<auto|enabled|disabled>"
```

Establece el modo de auto discover, el valor por defecto es `auto`. Cuando se establece en `auto`, los proyectos en un repo serán descubiertos por
Atlantis cuando no haya proyectos configurados en la configuración del repo. Si uno o más proyectos están definidos
en la configuración del repo entonces el auto discovery será deshabilitado completamente.

Cuando se establece en `enabled` los proyectos serán descubiertos incondicionalmente. Si un proyecto auto descubierto ya está
definido en la sección projects de la configuración del repo, el proyecto de la configuración del repo tendrá precedencia sobre
el proyecto auto descubierto.

Cuando se establece en `disabled` los proyectos nunca serán descubiertos, incluso si no hay proyectos configurados en la configuración del repo.

### `--automerge` <Badge text="v0.17.0" type="info"/>

```bash
atlantis server --automerge
# or
ATLANTIS_AUTOMERGE=true
```

Fusionar automáticamente pull requests después de que todos los plans se hayan aplicado correctamente.
Por defecto es `false`. Vea [Automerging](automerging.md) para más detalles.

### `--automerge-method`

```bash
atlantis server --automerge-method="squash"
# or
ATLANTIS_AUTOMERGE_METHOD="squash"
```

Método de merge predeterminado a usar al hacer automerge de pull requests. Los valores válidos son
`merge`, `rebase` y `squash`. Cuando no se establece, se usa el método de merge
predeterminado del proveedor VCS. Esto puede sobrescribirse por comando con el flag de comentario `--auto-merge-method`.
Actualmente solo implementado para GitHub.

### `--autoplan-file-list` <Badge text="v0.15.0+" type="info"/>

```bash
# NOTE: Use single quotes to avoid shell expansion of *.
atlantis server --autoplan-file-list='**/*.tf,project1/*.pkr.hcl'
# or
ATLANTIS_AUTOPLAN_FILE_LIST='**/*.tf,project1/*.pkr.hcl'
```

Lista de patrones de archivos que Atlantis usará para verificar si un directorio contiene archivos modificados que deberían activar el planning del proyecto.

Notas:

- Acepta una lista separada por comas, ej. `pattern1,pattern2`.
- Los patrones usan la [sintaxis de `.dockerignore`](https://docs.docker.com/engine/reference/builder/#dockerignore-file)
- La lista de patrones de archivos será usada tanto por plans automáticos como ejecutados manualmente.
- Cuando no se establece, por defecto son todos los archivos `.tf`, `.tf.json`, `.tfvars`, `.tfvars.json`, `.tofu`, `.tofu.json`, `terragrunt.hcl` y `.terraform.lock.hcl`
   (`--autoplan-file-list='**/*.tf,**/*.tf.json,**/*.tfvars,**/*.tfvars.json,**/*.tofu,**/*.tofu.json,**/terragrunt.hcl,**/.terraform.lock.hcl'`).
- Establecer `--autoplan-file-list` sobrescribirá los valores predeterminados. **Debe** agregar `**/*.tf` y otros valores predeterminados si quiere incluirlos.
- El valor predeterminado es global (no consciente de la distribución). Tanto las instalaciones de Terraform como de OpenTofu coincidirán con cambios en `.tofu`. Si no usa OpenTofu, puede sobrescribir el valor predeterminado para excluir patrones `.tofu`.
- Un [Workflow](repo-level-atlantis-yaml.md#configuring-planning) personalizado que usa autoplan `when_modified` ignorará este valor.

Ejemplos:

- Autoplan cuando cualquier archivo `*.tf` o `*.tfvars` es modificado.
  - `--autoplan-file-list='**/*.tf,**/*.tfvars'`
- Autoplan cuando cualquier archivo `*.tf` es modificado excepto en el directorio `project2/`
  - `--autoplan-file-list='**/*.tf,!project2'`
- Autoplan cuando cualquier archivo `*.tf` o archivos `.yml` en subcarpeta de `project1` es modificado.
  - `--autoplan-file-list='**/*.tf,project1/**/*.yml'`

::: warning NOTE
Por defecto, los cambios en módulos no activarán autoplanning. Vea los flags abajo.
:::

### `--autoplan-modules` <Badge text="v0.26.0+" type="info"/>

```bash
atlantis server --autoplan-modules
# or
ATLANTIS_AUTOPLAN_MODULES=true
```

Por defecto es `false`. Cuando se establece en `true`, Atlantis rastreará los módulos locales de los proyectos incluidos.
Los proyectos incluidos son proyectos con archivos incluidos por `--autoplan-file-list`.
Después del rastreo, Atlantis hará plan de cualquier proyecto que incluya un módulo cambiado. Esto es equivalente a establecer
`--autoplan-modules-from-projects` al valor de `--autoplan-file-list`. Vea abajo.

::: tip NOTE
La indexación de dependencias de módulos usa inspección de configuración de Terraform y puede no soportar completamente archivos `.tofu` / `.tofu.json`. Las llamadas a módulos definidas solo en archivos `.tofu` o directorios de módulos compartidos que contienen solo archivos `.tofu` pueden no ser rastreadas. El autoplanning directo por cambio de archivo para proyectos `.tofu` funciona de todos modos. Use patrones explícitos `autoplan.when_modified` como solución temporal. Vea [OpenTofu .tofu file support](terraform-versions.md#opentofu-tofu-file-support) para detalles.
:::

### `--autoplan-modules-from-projects` <Badge text="v0.26.0+" type="info"/>

```bash
atlantis server --autoplan-modules-from-projects='**/init.tf'
# or
ATLANTIS_AUTOPLAN_MODULES_FROM_PROJECTS='**/init.tf'
```

Habilita el auto-planing de proyectos cuando una dependencia de módulo en el mismo repositorio ha cambiado.
Esta es una lista de patrones de archivos como `autoplan-file-list`.

Estos patrones seleccionan **proyectos** para indexar basándose en los archivos coincidentes. El índice mapea módulos a los proyectos que dependen de ellos,
incluyendo proyectos que incluyen el módulo mediante otros módulos. Cuando cambia un archivo de módulo que coincide con `autoplan-file-list`,
se hará plan de todos los proyectos indexados.

El valor predeterminado actual es "" (deshabilitado).

Ejemplos:

- `**/*.tf` - indexará todos los proyectos que tengan un archivo `.tf` en su directorio, y hará plan de ellos siempre que una dependencia de módulo dentro del repo haya cambiado.
- `**/*.tf,!foo,!bar` - indexará todos los proyectos que contengan `.tf` excepto `foo` e `bar` y hará plan de ellos siempre que una dependencia de módulo dentro del repo haya cambiado.
   Esto permite que los proyectos hagan opt-out del auto-planning cuando una dependencia de módulo cambia.

::: warning NOTE
Los módulos que no son seleccionados por autoplan-file-list no serán indexados y no se hará plan de los proyectos dependientes. Este
flag permite seleccionar los _projects_ a indexar, pero el disparador para un plan debe ser un archivo en `autoplan-file-list`.
:::

::: warning NOTE
Este flag sobrescribe `--autoplan-modules`. Si desea deshabilitar el auto-planning de módulos, establezca este flag en una cadena vacía,
y establezca `--autoplan-modules` en `false`.
:::

### `--azuredevops-hostname` <Badge text="v0.9.0+" type="info"/>

```bash
atlantis server --azuredevops-hostname="dev.azure.com"
# or
ATLANTIS_AZUREDEVOPS_HOSTNAME="dev.azure.com"
```

Nombre de host de Azure DevOps para soportar instancias cloud y self-hosted. Por defecto es `dev.azure.com`.

::: warning COMPATIBILITY WARNING
Si este cambio le afecta [docs](https://learn.microsoft.com/en-us/azure/devops/release-notes/2018/sep-10-azure-devops-launch#administration)
o este [issue](https://github.com/runatlantis/atlantis/issues/5595)
ambos Service Hooks (v1 y v2) convertirán el nombre de la organización de AD a minúsculas:
Ejemplos:
`https://dev.azure.com/MYCompany/` e `https://mycompany.visualstudio.com/` se convertirán en `mycompany`
`https://dev.azure.com/MYCOMPANY/` e `https://myCOMPANY.visualstudio.com/` se convertirán en `mycompany`

Este [cambio](https://github.com/runatlantis/atlantis/pull/5596) se aplicará desde la versión v0.35.0

¿Qué hacer si tiene plans pendientes que fueron generados con una versión anterior?
Ejecutar un atlantis unlock desde v0.35.0 en sus PR actuales ignorará los archivos en la carpeta `MYCompany`. En el siguiente atlantis plan usará la carpeta `mycompany` y generará todo en el nuevo nombre de carpeta
:::

### `--azuredevops-token` <Badge text="v0.9.0+" type="info"/>

```bash
atlantis server --azuredevops-token="RandomStringProducedByAzureDevOps"
# or (recommended)
ATLANTIS_AZUREDEVOPS_TOKEN="RandomStringProducedByAzureDevOps"
```

Token de Azure DevOps del usuario de API.

### `--azuredevops-user` <Badge text="v0.9.0+" type="info"/>

```bash
atlantis server --azuredevops-user="username@example.com"
# or
ATLANTIS_AZUREDEVOPS_USER="username@example.com"
```

Nombre de usuario de Azure DevOps del usuario de API.

### `--azuredevops-webhook-password` <Badge text="v0.9.0+" type="info"/>

```bash
atlantis server --azuredevops-webhook-password="password123"
# or (recommended)
ATLANTIS_AZUREDEVOPS_WEBHOOK_PASSWORD="password123"
```

Contraseña de autenticación básica de Azure DevOps para webhooks entrantes (vea
[docs](https://docs.microsoft.com/en-us/azure/devops/service-hooks/authorize?view=azure-devops)).

::: warning SECURITY WARNING
Si no se especifica, Atlantis no podrá validar que la
llamada webhook entrante provino de su organización de Azure DevOps. Esto significa que un
atacante podría falsificar llamadas a Atlantis y hacer que realice acciones
maliciosas. Debe especificarse mediante la variable de entorno `ATLANTIS_AZUREDEVOPS_WEBHOOK_PASSWORD`.
:::

### `--azuredevops-webhook-user` <Badge text="v0.9.0+" type="info"/>

```bash
atlantis server --azuredevops-webhook-user="username@example.com"
# or
ATLANTIS_AZUREDEVOPS_WEBHOOK_USER="username@example.com"
```

Nombre de usuario de autenticación básica de Azure DevOps para webhooks entrantes.

### `--bitbucket-api-user` <Badge text="v0.36.0+" type="info"/>

```bash
atlantis server --bitbucket-api-user="apiuser@example.com"
# or
ATLANTIS_BITBUCKET_API_USER="apiuser@example.com"
```

Nombre de usuario de Bitbucket (normalmente un email) usado para autenticación de API con Bitbucket Cloud. Esto se usa solo para llamadas de API. Si no se especifica, Atlantis usará el valor de `--bitbucket-user` para autenticación de API para mantener compatibilidad hacia atrás.

**Nota:**

- La compatibilidad hacia atrás es para soportar las APP Passwords existentes de Bitbucket que siguen siendo válidas hasta junio de 2026 (vea [Atlassian's Bitbucket app password deprecation notice](https://www.atlassian.com/blog/bitbucket/bitbucket-cloud-transitions-to-api-tokens-enhancing-security-with-app-password-deprecation)).

**Clave del archivo de configuración:**

```yaml
bitbucket-api-user: apiuser@example.com
```

**Variable de entorno:** `ATLANTIS_BITBUCKET_API_USER`

**Nota:** Este flag solo es relevante para integraciones de Bitbucket Cloud (bitbucket.org).

### `--bitbucket-base-url` <Badge text="v0.36.0+" type="info"/>

```bash
atlantis server --bitbucket-base-url="http://bitbucket.corp:7990/basepath"
# or
ATLANTIS_BITBUCKET_BASE_URL="http://bitbucket.corp:7990/basepath"
```

URL base de la instalación de Bitbucket Server (también conocido como Stash). Debe incluir
`http://` o `https://`. Si usa Bitbucket Cloud (bitbucket.org), no lo establezca. Por defecto es
`https://api.bitbucket.org`.

### `--bitbucket-token` <Badge text="v0.36.0+" type="info"/>

```bash
atlantis server --bitbucket-token="token"
# or (recommended)
ATLANTIS_BITBUCKET_TOKEN="token"
```

App password de Bitbucket del usuario de API.

### `--bitbucket-user` <Badge text="v0.36.0+" type="info"/>

```bash
atlantis server --bitbucket-user="myuser"
# or
ATLANTIS_BITBUCKET_USER="myuser"
```

Nombre de usuario de Bitbucket usado para operaciones git. Para Bitbucket Cloud, si `--bitbucket-api-user` no se especifica, este valor también se usará para autenticación de API.

### `--bitbucket-webhook-secret` <Badge text="v0.36.0+" type="info"/>

```bash
atlantis server --bitbucket-webhook-secret="secret"
# or (recommended)
ATLANTIS_BITBUCKET_WEBHOOK_SECRET="secret"
```

Secreto usado para validar webhooks de Bitbucket.

::: warning SECURITY WARNING
Si no se especifica, Atlantis no podrá validar que la llamada webhook entrante vino de Bitbucket.
Esto significa que un atacante podría falsificar llamadas a Atlantis y hacer que realice acciones maliciosas.
:::

### `--blocked-extra-args`

```bash
atlantis server --blocked-extra-args="-chdir,--chdir,-plugin-dir,--plugin-dir"
# or
ATLANTIS_BLOCKED_EXTRA_ARGS='-chdir,--chdir,-plugin-dir,--plugin-dir'
```

Lista separada por comas de prefijos de flags de Terraform CLI que no están permitidos en args extra de comentarios (los flags después de `--`).
Por defecto es `-chdir,--chdir,-plugin-dir,--plugin-dir`.

Notas:

- Estos flags se bloquean para prevenir problemas de seguridad tales como traversal de directorio de trabajo (`-chdir`) o cargar providers maliciosos (`-plugin-dir`).
- Establecer este flag **reemplaza** por completo la lista predeterminada. Para extender los valores predeterminados, inclúyalos junto con sus flags personalizados, p. ej. `-chdir,--chdir,-plugin-dir,--plugin-dir,-my-flag`.
- Acepta una lista separada por comas, ej. `-flag1,-flag2`.

### `--checkout-depth` <Badge text="v0.28.0+" type="info"/>

```bash
atlantis server --checkout-depth=0
# or
ATLANTIS_CHECKOUT_DEPTH=0
```

El número de commits a obtener de la rama. Usado si `--checkout-strategy=merge` ya que la estrategia de checkout `--checkout-strategy=branch` (predeterminada) siempre usa por defecto un shallow clone con una profundidad de 1.
Por defecto es `0`. Vea [Checkout Strategy](checkout-strategy.md) para más detalles.

### `--checkout-strategy` <Badge text="v0.9.0+" type="info"/>

```bash
atlantis server --checkout-strategy="<branch|merge>"
# or
ATLANTIS_CHECKOUT_STRATEGY="<branch|merge>"
```

Cómo hacer checkout de pull requests. Use `branch` o `merge`.
Por defecto es `branch`. Vea [Checkout Strategy](checkout-strategy.md) para más detalles.

### `--config` <Badge text="v0.1.3+" type="info"/>

```bash
atlantis server --config="my/config/file.yaml"
# or
ATLANTIS_CONFIG="my/config/file.yaml"
```

Archivo de configuración YAML donde también se pueden establecer flags. Vea [Config File](#config-file) para más detalles.

### `--data-dir` <Badge text="v0.1.3+" type="info"/>

```bash
atlantis server --data-dir="path/to/data/dir"
# or
ATLANTIS_DATA_DIR="path/to/data/dir"
```

Directorio donde Atlantis almacenará sus datos. Será creado si no existe.
Por defecto es `~/.atlantis`. Atlantis almacenará aquí su base de datos, repos con checkout, plans de Terraform y binarios de
Terraform descargados. Si Atlantis pierde este directorio, los [locks](locking.md)
se perderán y los plans no aplicados se perderán.

Tenga en cuenta que el usuario atlantis está restringido a `~/.atlantis`.
Si establece el flag `--data-dir` a una ruta fuera del directorio home de Atlantis, asegúrese de otorgar al usuario atlantis los permisos correctos.

### `--default-tf-distribution` <Badge text="v0.24.0+" type="info"/>

```bash
atlantis server --default-tf-distribution="terraform"
# or
ATLANTIS_DEFAULT_TF_DISTRIBUTION="terraform"
```

Qué distribución de TF usar. Puede establecerse en `terraform` o `opentofu`.

### `--default-tf-version` <Badge text="v0.13.0" type="info"/>

```bash
atlantis server --default-tf-version="v0.12.31"
# or
ATLANTIS_DEFAULT_TF_VERSION="v0.12.31"
```

Versión de Terraform a usar por defecto. Se descargará a `<data-dir>/bin/terraform<version>`
si no está en `PATH`. Vea [Terraform Versions](terraform-versions.md) para más detalles.

### `--disable-apply-all` <Badge text="v0.9.0+" type="info"/>

```bash
atlantis server --disable-apply-all
# or
ATLANTIS_DISABLE_APPLY_ALL=true
```

Deshabilitar el comando `atlantis apply` para que tenga que
especificarse un project/workspace/directory específico para applies.

### `--disable-automerge-label` <Badge text="v0.45.0+" type="info"/>

```bash
atlantis server --disable-automerge-label="no-auto-merge"
# or
ATLANTIS_DISABLE_AUTOMERGE_LABEL="no-auto-merge"
```

Deshabilitar atlantis automerge solo en pull requests con la label especificada.
Por defecto es una cadena vacía, por lo que ninguna label deshabilita automerge de forma predeterminada.
Este flag no tiene efecto a menos que automerge esté habilitado con `--automerge` o
`automerge: true` a nivel repo.

### `--disable-autoplan` <Badge text="v0.15.0+" type="info"/>

```bash
atlantis server --disable-autoplan
# or
ATLANTIS_DISABLE_AUTOPLAN=true
```

Deshabilitar auto planning de atlantis.

### `--disable-autoplan-label` <Badge text="v0.33.0+" type="info"/>

```bash
atlantis server --disable-autoplan-label="no-autoplan"
# or
ATLANTIS_DISABLE_AUTOPLAN_LABEL="no-autoplan"
```

Deshabilitar auto planning de atlantis solo en pull requests con la label especificada.

Si la propiedad `disable-autoplan` es `true`, este flag no tiene efecto.

### `--disable-global-apply-lock` <Badge text="v0.17.0" type="info"/>

```bash
atlantis server --disable-global-apply-lock
# or
ATLANTIS_DISABLE_GLOBAL_APPLY_LOCK=true
```

Si es true, elimina el botón en la UI que permite a los usuarios deshabilitar globalmente comandos apply.

### `--disable-markdown-folding` <Badge text="v0.31.0+" type="info"/>

```bash
atlantis server --disable-markdown-folding
# or
ATLANTIS_DISABLE_MARKDOWN_FOLDING=true
```

Deshabilitar folding en la salida markdown usando la etiqueta html `<details>`.

### `--disable-repo-locking` <Badge text="v0.16.1" type="info"/>

```bash
atlantis server --disable-repo-locking
# or
ATLANTIS_DISABLE_REPO_LOCKING=true
```

Evita que atlantis bloquee projects y/o workspaces al ejecutar terraform.

### `--disable-unlock-label` <Badge text="v0.33.0+" type="info"/>

```bash
atlantis server --disable-unlock-label do-not-unlock
# or
ATLANTIS_DISABLE_UNLOCK_LABEL="do-not-unlock"
```

Evita que atlantis desbloquee un pull request con esta label. Por defecto es "" (feature deshabilitada).

### `--discard-approval-on-plan` <Badge text="v0.29.0+" type="info"/>

```bash
atlantis server --discard-approval-on-plan
# or
ATLANTIS_DISCARD_APPROVAL_ON_PLAN=true
```

Si se establece, descarta la aprobación si se ha ejecutado un nuevo plan. Actualmente solo soportado en GitHub y GitLab. Para GitLab se requiere un token de bot, grupo o proyecto para esta feature.
 Referencia: [reset-approvals-of-a-merge-request](https://docs.gitlab.com/api/merge_request_approvals/#reset-approvals-of-a-merge-request)

### `--emoji-reaction` <Badge text="v0.29.0+" type="info"/>

```bash
atlantis server --emoji-reaction eyes
# or
ATLANTIS_EMOJI_REACTION=eyes
```

La reacción emoji a usar para marcar comentarios procesados. Actualmente soportado en Gitea, GitHub y GitLab. Si no se especifica, Atlantis no usará una reacción emoji.
Por defecto es "" (cadena vacía).

::: warning NOTE
Cada proveedor VCS soporta una lista diferente de emojis:

- [GitHub](https://docs.github.com/en/rest/reactions/reactions?apiVersion=2022-11-28#about-reactions)
- [GitLab](https://gitlab.com/gitlab-org/gitlab/-/blob/master/fixtures/emojis/digests.json)
- [Gitea](https://docs.gitea.com/administration/customizing-gitea#reactions)

   :::

### `--enable-diff-markdown-format` <Badge text="v0.25.0+" type="info"/>

```bash
atlantis server --enable-diff-markdown-format
# or
ATLANTIS_ENABLE_DIFF_MARKDOWN_FORMAT=true
```

Habilitar Atlantis para formatear la salida de Terraform plan en un formato amigable con markdown-diff para propósitos de color-coding.

Útil de habilitar para usar con GitHub. Las líneas cambiadas dentro de diffs de heredoc y multiline-string de Terraform también se formatean para que los renderizadores markdown conscientes de diff puedan colorearlas.

### `--enable-drift-detection`

```bash
atlantis server --enable-drift-detection
# or
ATLANTIS_ENABLE_DRIFT_DETECTION=true
```

Habilitar endpoints de API de detección de drift. La detección de drift no ejecuta Terraform apply, pero
sí ejecuta el ciclo de vida normal de plan, incluyendo hooks pre-workflow configurados,
workflows personalizados, pasos de plan personalizados y comandos Terraform plan. Cuando está habilitado, Atlantis
inicializará almacenamiento en memoria para resultados de detección de drift y un servicio de remediación,
haciendo funcionales los endpoints de drift detection, status y remediación solo-plan. Si los [webhooks](sending-notifications-via-webhooks.md#drift-detection-webhooks) de drift
están configurados (`event: drift`), las ejecuciones de detección exitosas envían notificaciones a Slack o endpoints HTTP,
incluyendo resultados heartbeat de no-drift. La detección de drift no omite allowlists de equipos ni requisitos de estado de PR
`plan_requirements` como `approved` o `mergeable`; esas comprobaciones fallan cerradas cuando
no pueden evaluarse fuera de un pull request. Las acciones apply destructivas de remediación de drift también requieren
`--enable-drift-remediation`. Por defecto es `false`.

### `--enable-drift-remediation`

```bash
atlantis server --enable-drift-detection --enable-drift-remediation
# or
ATLANTIS_ENABLE_DRIFT_DETECTION=true
ATLANTIS_ENABLE_DRIFT_REMEDIATION=true
```

Habilitar acciones apply destructivas de remediación de drift en el endpoint `/api/drift/remediate`.
Este flag requiere `--enable-drift-detection`; sin él, las solicitudes `action: "apply"` son
rechazadas mientras la detección de drift de solo lectura sigue disponible. Este flag no omite
`apply_requirements` del repositorio; los requisitos que necesitan estado de pull request fallan cerrados para
solicitudes de remediación que no son PR. Por defecto es `false`.

### `--enable-external-stores`

```bash
atlantis server --enable-external-stores
# or
ATLANTIS_ENABLE_EXTERNAL_STORES=true
```

Habilitar backends de almacenamiento externo configurados en la configuración de repo del lado del servidor (bloque `external_stores`). Cuando se establece, Atlantis lee la sección `external_stores` del YAML de configuración del repo para inicializar backends como S3 para persistencia de archivos plan.

### `--enable-policy-checks` <Badge text="v0.17.0" type="info"/>

```bash
atlantis server --enable-policy-checks
# or
ATLANTIS_ENABLE_POLICY_CHECKS=true
```

Habilita atlantis para ejecutar policies del lado del servidor sobre el resultado de un terraform plan. Las policies están definidas en [server side repo config](server-side-repo-config.md#reference).

### `--enable-profiling-api` <Badge text="v0.25.0+" type="info"/>

```bash
atlantis server --enable-profiling-api
# or
ATLANTIS_ENABLE_PROFILING_API=true
```

Habilitar endpoints [`net/http/pprof`](https://pkg.go.dev/net/http/pprof) para [continuous profiling](https://grafana.com/docs/pyroscope/latest/introduction/continuous-profiling/) de recursos usados por el servidor. Vea [profiling Go programs](https://go.dev/blog/pprof) para más información.

### `--enable-regexp-cmd` <Badge text="v0.17.0" type="info"/>

```bash
atlantis server --enable-regexp-cmd
# or
ATLANTIS_ENABLE_REGEXP_CMD=true
```

Habilitar Atlantis para usar expresiones regulares para ejecutar comandos plan/apply contra nombres de proyecto definidos cuando se pasa con ello el flag `-p`.

Esto se puede usar para ejecutar todos los proyectos definidos (con la clave `name`) en `atlantis.yaml` usando `atlantis plan -p .*`.

El flag solo permitirá las regex listadas en la clave [`allowed_regexp_prefixes`](repo-level-atlantis-yaml.md#reference) definida en el archivo repo `atlantis.yaml`. Si la clave no está definida, su valor por defecto es `[]` que permitirá cualquier regex.

Esto todavía no funcionará con `-d` y para usar `-p` los proyectos del repo deben estar definidos en el archivo repo `atlantis.yaml`.

Cuando `--restrict-file-list` está habilitado, los plans de proyecto regex se limitan a coincidir con proyectos con archivos modificados en el pull request. Sin `--restrict-file-list`, los comandos de proyecto regex todavía pueden ejecutarse contra todos los proyectos coincidentes.

::: warning SECURITY WARNING
No se supone que se use con `--disable-apply-all`.
El comando `atlantis apply -p .*` omitirá la restricción y ejecutará apply en cada proyecto.
:::

### `--executable-name` <Badge text="v0.42.0+" type="info"/>

```bash
atlantis server --executable-name="atlantis"
# or
ATLANTIS_EXECUTABLE_NAME="atlantis"
```

Nombre ejecutable disparador del comando de comentario. Por defecto es `atlantis`.

Esto es útil al ejecutar múltiples servidores Atlantis contra un solo repositorio.

### `--fail-on-pre-workflow-hook-error` <Badge text="v0.27.0+" type="info"/>

```bash
atlantis server --fail-on-pre-workflow-hook-error
# or
ATLANTIS_FAIL_ON_PRE_WORKFLOW_HOOK_ERROR=true
```

Fallar y no ejecutar el comando Atlantis solicitado si cualquiera de los hooks pre workflow devuelve error.

### `--gh-allow-mergeable-bypass-apply` <Badge text="v0.30.0+" type="info"/>

```bash
atlantis server --gh-allow-mergeable-bypass-apply
# or
ATLANTIS_GH_ALLOW_MERGEABLE_BYPASS_APPLY=true
```

Feature flag para habilitar la capacidad de usar el modo `mergeable` con required apply status check.

### `--gh-app-id` <Badge text="v0.20.0+" type="info"/>

```bash
atlantis server --gh-app-id="00000"
# or
ATLANTIS_GH_APP_ID="00000"
```

ID de app de GitHub. Si se establece, la autenticación de GitHub se realizará como [an installation](https://docs.github.com/en/rest/apps/installations).

::: tip
Una app de GitHub puede crearse iniciando Atlantis primero, luego apuntando su navegador a

```shell
$(hostname)/github-app/setup
```

Será redirigido a GitHub para crear una nueva app, y luego será redirigido a

```shell
$(hostname)/github-app/exchange-code?code=some-code
```

Después de lo cual Atlantis mostrará las credenciales de su nueva app: el ID de su app, su `--gh-webhook-secret` generado y el contenido del archivo para `--gh-app-key-file`. Actualice su configuración de Atlantis en consecuencia y reinicie el servidor.
:::

### `--gh-app-installation-id` <Badge text="v0.20.0+" type="info"/>

```bash
atlantis server --gh-app-installation-id="123"
# or
ATLANTIS_GH_APP_INSTALLATION_ID="123"
```

El installation ID de una instancia específica de una aplicación de GitHub. Normalmente este valor se
deriva consultando GitHub para la lista de installations del ID suministrado mediante `--gh-app-id` y seleccionando
la primera encontrada y donde múltiples resultados de installations producen un error. Use este flag si tiene múltiples
instancias de Atlantis pero quiere usar una sola app de GitHub ya instalada para todas ellas. Normalmente haría esto si
está ejecutando un proxy como su única aplicación de GitHub que hará proxy a una instancia apropiada de Atlantis
en función de la organización o usuario que disparó el webhook.

### `--gh-app-key` <Badge text="v0.20.0+" type="info"/>

```bash
atlantis server --gh-app-key="-----BEGIN RSA PRIVATE KEY-----(...)"
# or
ATLANTIS_GH_APP_KEY="-----BEGIN RSA PRIVATE KEY-----(...)"
```

La clave privada codificada en PEM para la GitHub App.

::: warning SECURITY WARNING
El contenido de la clave privada será visible para cualquiera que pueda ejecutar `ps` o mirar el historial del shell de la máquina donde Atlantis se está ejecutando. Use `--gh-app-key-file` para mitigar ese riesgo.
:::

### `--gh-app-key-file` <Badge text="v0.20.0+" type="info"/>

```bash
atlantis server --gh-app-key-file="path/to/app-key.pem"
# or
ATLANTIS_GH_APP_KEY_FILE="path/to/app-key.pem"
```

Ruta a un archivo de clave privada codificada en PEM de GitHub App. Si se establece, la autenticación de GitHub se realizará como [an installation](https://docs.github.com/en/rest/apps/installations).

### `--gh-app-slug` <Badge text="v0.16.1" type="info"/>

```bash
atlantis server --gh-app-slug="myappslug"
# or
ATLANTIS_GH_APP_SLUG="myappslug"
```

Una versión slugged del nombre de la app de GitHub mostrada en comentarios de pull requests, etc. (no `Atlantis App` sino algo como `atlantis-app`). Atlantis usa el valor de este parámetro para identificar los comentarios que ha dejado en pull requests de GitHub. Esto se usa para funciones como `--hide-prev-plan-comments`. Necesita obtener este valor de su app de GitHub; una manera es ir a la configuración de su App y abrir "Public page" desde la barra lateral izquierda. Su valor de `--gh-app-slug` será la última parte de la URL, p. ej. `https://github.com/apps/<slug>`.

### `--gh-hostname` <Badge text="v0.1.3+" type="info"/>

```bash
atlantis server --gh-hostname="my.github.enterprise.com"
# or
ATLANTIS_GH_HOSTNAME="my.github.enterprise.com"
```

Nombre de host de su instalación de GitHub Enterprise. Si usa [GitHub.com](https://github.com),
no lo establezca. Por defecto es `github.com`.

Para GitHub Enterprise Cloud, use el nombre de host del tenant, por ejemplo `tenant.ghe.com`. No incluya un esquema ni un prefijo `api.`; Atlantis deriva los endpoints de API REST y GraphQL a partir del nombre de host.

### `--gh-org` <Badge text="v0.1.3+" type="info"/>

```bash
atlantis server --gh-org="myorgname"
# or
ATLANTIS_GH_ORG="myorgname"
```

Nombre de organización de GitHub. Establézcalo para habilitar la creación de una app privada de GitHub para esta organización.

### `--gh-team-allowlist` <Badge text="v0.41.0+" type="info"/>

```bash
atlantis server --gh-team-allowlist="myteam:plan, secteam:apply, devops-team:apply, devops-team:import"
# or
ATLANTIS_GH_TEAM_ALLOWLIST="myteam:plan, secteam:apply, devops-team:apply, devops-team:import"
```

En las versiones v0.35.0 y posteriores, el nombre del equipo de GitHub solo puede ser un slug porque es inmutable.

En las versiones entre v0.21.0 y v0.34.0, el nombre del equipo de GitHub puede ser un nombre o un slug.

En las versiones v0.20.1 e inferiores, el nombre del equipo de GitHub requería el nombre del equipo sensible a mayúsculas y minúsculas.

Lista separada por comas de pares de equipos de GitHub y permisos.

Por defecto, cualquier equipo puede plan y apply.

Se respeta la jerarquía de equipos de GitHub. Si un equipo en allowlist tiene equipos hijo, los miembros de esos equipos hijo heredan los comandos permitidos del equipo padre.

::: tip
Si está usando [policy checking](policy-checking.md), también debe permitir el comando `policy_check` para que funcione en comandos manuales `atlantis plan`:

```bash
atlantis server --gh-team-allowlist="*:plan,*:policy_check,myteam:apply"
```

Vea [Policy Checking documentation](policy-checking.md#step-1-enable-the-workflow) para más detalles.
:::

### `--gh-token` <Badge text="v0.1.3+" type="info"/>

```bash
atlantis server --gh-token="token"
# or (recommended)
ATLANTIS_GH_TOKEN="token"
```

Token de GitHub del usuario de API.

### `--gh-token-file` <Badge text="v0.41.0+" type="info"/>

```bash
atlantis server --gh-token-file="/path/to/token"
# or
ATLANTIS_GH_TOKEN_FILE="/path/to/token"
```

Token de GitHub del usuario de API. El token se carga regularmente desde disco para permitir la rotación del token sin necesidad de reiniciar el servidor Atlantis.

### `--gh-user` <Badge text="v0.1.3+" type="info"/>

```bash
atlantis server --gh-user="myuser"
# or
ATLANTIS_GH_USER="myuser"
```

Nombre de usuario de GitHub del usuario de API. Este usuario también es usado por el flag `--hide-user-plan-comments` y deberá actualizarse si migra a github EMU.

### `--gh-webhook-secret` <Badge text="v0.1.3+" type="info"/>

```bash
atlantis server --gh-webhook-secret="secret"
# or (recommended)
ATLANTIS_GH_WEBHOOK_SECRET="secret"
```

Secreto usado para validar webhooks de GitHub (vea [GitHub: Validating webhook deliveries](https://docs.github.com/en/webhooks/using-webhooks/validating-webhook-deliveries)).

::: warning SECURITY WARNING
Si no se especifica, Atlantis no podrá validar que la llamada webhook entrante vino de GitHub.
Esto significa que un atacante podría falsificar llamadas a Atlantis y hacer que realice acciones maliciosas.
:::

### `--gitea-base-url` <Badge text="v0.28.0+" type="info"/>

```bash
atlantis server --gitea-base-url="http://your-gitea.corp:7990/basepath"
# or
ATLANTIS_GITEA_BASE_URL="http://your-gitea.corp:7990/basepath"
```

URL base de la instalación de Gitea. Debe incluir `http://` o `https://`. Por defecto es `https://gitea.com` si se deja vacío/ausente.

### `--gitea-page-size` <Badge text="v0.28.0+" type="info"/>

```bash
atlantis server --gitea-page-size=30
# or (recommended)
ATLANTIS_GITEA_PAGE_SIZE=30
```

Número de elementos en una sola página en respuestas paginadas de Gitea.

::: warning Configuration dependent
El valor predeterminado se ajusta a la configuración estándar del servidor Gitea: DEFAULT_PAGING_NUM
El valor válido más alto depende de la configuración del servidor Gitea: MAX_RESPONSE_ITEMS
:::

### `--gitea-token` <Badge text="v0.28.0+" type="info"/>

```bash
atlantis server --gitea-token="token"
# or (recommended)
ATLANTIS_GITEA_TOKEN="token"
```

App password de Gitea del usuario de API.

### `--gitea-user` <Badge text="v0.28.0+" type="info"/>

```bash
atlantis server --gitea-user="myuser"
# or
ATLANTIS_GITEA_USER="myuser"
```

Nombre de usuario de Gitea del usuario de API.

### `--gitea-webhook-secret` <Badge text="v0.28.0+" type="info"/>

```bash
atlantis server --gitea-webhook-secret="secret"
# or (recommended)
ATLANTIS_GITEA_WEBHOOK_SECRET="secret"
```

Secreto usado para validar webhooks de Gitea.

::: warning SECURITY WARNING
Si no se especifica, Atlantis no podrá validar que la llamada webhook entrante vino de Gitea.
Esto significa que un atacante podría falsificar llamadas a Atlantis y hacer que realice acciones maliciosas.
:::

### `--gitlab-group-allowlist` <Badge text="v0.13.0+" type="info"/>

```bash
atlantis server --gitlab-group-allowlist="myorg/mygroup:plan, myorg/secteam:apply, myorg/devops:apply, myorg/devops:import"
# or
ATLANTIS_GITLAB_GROUP_ALLOWLIST="myorg/mygroup:plan, myorg/secteam:apply, myorg/devops:apply, myorg/devops:import"
```

Lista separada por comas de grupos de GitLab y pares de permisos.

Por defecto, cualquier grupo puede plan y apply.

::: warning NOTE
Atlantis necesita poder ver los miembros del grupo listado; los grupos inaccesibles o no existentes se ignoran silenciosamente.
:::

### `--gitlab-hostname` <Badge text="v0.2.0+" type="info"/>

```bash
atlantis server --gitlab-hostname="my.gitlab.enterprise.com"
# or
ATLANTIS_GITLAB_HOSTNAME="my.gitlab.enterprise.com"
```

Nombre de host de su instalación de GitLab Enterprise. Si usa [GitLab.com](https://gitlab.com),
no lo establezca. Por defecto es `gitlab.com`.

### `--gitlab-status-retry-enabled`

```bash
atlantis server --gitlab-status-retry-enabled
# or
ATLANTIS_GITLAB_STATUS_RETRY_ENABLED=true
```

Habilitar lógica de reintento mejorada para actualizaciones de estado de pipeline de GitLab con exponential backoff.

Por defecto es `false`.

### `--gitlab-token` <Badge text="v0.2.0+" type="info"/>

```bash
atlantis server --gitlab-token="token"
# or (recommended)
ATLANTIS_GITLAB_TOKEN="token"
```

Token de GitLab del usuario de API.

### `--gitlab-user` <Badge text="v0.2.0+" type="info"/>

```bash
atlantis server --gitlab-user="myuser"
# or
ATLANTIS_GITLAB_USER="myuser"
```

Nombre de usuario de GitLab del usuario de API.

### `--gitlab-webhook-secret` <Badge text="v0.2.0+" type="info"/>

```bash
atlantis server --gitlab-webhook-secret="secret"
# or (recommended)
ATLANTIS_GITLAB_WEBHOOK_SECRET="secret"
```

Secreto usado para validar webhooks de GitLab.

::: warning SECURITY WARNING
Si no se especifica, Atlantis no podrá validar que la llamada webhook entrante vino de GitLab.
Esto significa que un atacante podría falsificar llamadas a Atlantis y hacer que realice acciones maliciosas.
:::

### `--help` <Badge text="v0.1.3+" type="info"/>

```bash
atlantis server --help
```

Ver ayuda.

### `--hide-prev-plan-comments` <Badge text="v0.19.0" type="info"/>

```bash
atlantis server --hide-prev-plan-comments
# or
ATLANTIS_HIDE_PREV_PLAN_COMMENTS=true
```

Ocultar comentarios previos de plan para reducir el desorden en los PRs. Esto actualmente solo es soportado en
GitHub, GitLab y Bitbucket, y no está habilitado por defecto.

Para Bitbucket, los comentarios se eliminan en lugar de ocultarse ya que Bitbucket no soporta ocultar comentarios.

Para GitHub, asegúrese de que `--gh-user` esté establecido apropiadamente o los comentarios no se ocultarán.

Al usar GitHub App, necesita establecer `--gh-app-slug` para habilitar esta feature.

### `--hide-unchanged-plan-comments` <Badge text="v0.29.0+" type="info"/>

```bash
atlantis server --hide-unchanged-plan-comments
# or
ATLANTIS_HIDE_UNCHANGED_PLAN_COMMENTS=true
```

Eliminar comentarios de plan sin cambios del pull request.

Esto es útil cuando tiene muchos proyectos y quiere mantener el pull request limpio de comentarios inútiles.

### `--ignore-vcs-status-names` <Badge text="v0.30.0+" type="info"/>

```bash
atlantis server --ignore-vcs-status-names="status1,status2"
# or
ATLANTIS_IGNORE_VCS_STATUS_NAMES=status1,status2
```

Lista separada por comas de nombres de estado VCS de otros servicios atlantis.
Cuando `gh-allow-mergeable-bypass-apply` es true, ignorará las comprobaciones de estado
(p. ej. `status1/plan`, `status1/apply`, `status2/plan`, `status2/apply`)
de otros servicios Atlantis al comprobar si el PR es mergeable.
Actualmente solo implementado para GitHub.

### `--include-git-untracked-files` <Badge text="v0.27.0+" type="info"/>

```bash
atlantis server --include-git-untracked-files
# or
ATLANTIS_INCLUDE_GIT_UNTRACKED_FILES=true
```

Incluir archivos git untracked en la lista de archivos modificados de Atlantis.
Usado por ejemplo con hooks pre-workflow de CDKTF que generan dinámicamente
archivos Terraform.

### `--language` <Badge text="v0.45.0+" type="info"/>

```bash
atlantis server --language="en"
# or
ATLANTIS_LANGUAGE="en"
```

Idioma usado para los comentarios de pull request de Atlantis. Por defecto es `en`.

Valores soportados:

- `en` (Inglés)
- `es` (Español)

Atlantis normaliza valores tipo locale (por ejemplo `es-MX` resuelve a `es`).
Si se configura un idioma no soportado y `--language-config-file` no está establecido,
Atlantis devuelve un error de validación al iniciar.

Las cadenas de idioma integradas se cargan desde:

- `server/i18n/locales/en.yaml`
- `server/i18n/locales/es.yaml`

### `--language-config-file` <Badge text="v0.45.0+" type="info"/>

```bash
atlantis server --language="de" --language-config-file="/etc/atlantis/language.yaml"
# or
ATLANTIS_LANGUAGE_CONFIG_FILE="/etc/atlantis/language.yaml"
```

Ruta opcional a un catálogo de idioma YAML personalizado. Los valores en este archivo sobrescriben
el idioma integrado seleccionado, y se soportan sobrescrituras parciales.

Cuando `--language-config-file` está establecido, se permiten valores `--language` no soportados
y Atlantis vuelve al inglés integrado para cadenas no especificadas.

Esquema YAML esperado:

```yaml
pull_request_label: Pull Request (custom)
merge_request_label: Merge Request (custom)
command_titles:
  plan: Plan (custom)
  apply: Apply (custom)
```

Para una personalización completa del texto markdown, siga usando
`--markdown-template-overrides-dir`.

### `--locking-db-type` <Badge text="v0.19.9+" type="info"/>

```bash
atlantis server --locking-db-type="<boltdb|redis>"
# or
ATLANTIS_LOCKING_DB_TYPE="<boltdb|redis>"
```

El tipo de base de datos de locking a usar para almacenar locks de plan y apply. Por defecto es `boltdb`.

Notas:

- Si se establece en `boltdb`, solo un proceso puede tener acceso a la instancia boltdb.
- Si se establece en `redis`, use `--redis-host` y `--redis-port` para modo de nodo único, o `--redis-cluster-addresses` para modo Redis Cluster. Use `--redis-password` y (opcionalmente) `--redis-username` solo si su despliegue Redis requiere autenticación.

### `--log-level` <Badge text="v0.1.3+" type="info"/>

```bash
atlantis server --log-level="<debug|info|warn|error>"
# or
ATLANTIS_LOG_LEVEL="<debug|info|warn|error>"
```

Nivel de log. Por defecto es `info`.

### `--markdown-template-overrides-dir` <Badge text="v0.21.0" type="info"/>

```bash
atlantis server --markdown-template-overrides-dir="path/to/templates/"
# or
ATLANTIS_MARKDOWN_TEMPLATE_OVERRIDES_DIR="path/to/templates/"
```

Esto estará disponible en v0.21.0.

Directorio donde Atlantis leerá sobrescrituras para plantillas markdown usadas para renderizar comentarios en pull requests.
Las sobrescrituras de plantillas markdown pueden especificarse en archivos individuales, o todas juntas en un solo archivo. Todos los archivos de
sobrescritura de plantilla _deben_ tener la extensión `.tmpl`, de lo contrario no serán parseados.

Las plantillas markdown que pueden tener sobrescrituras se pueden encontrar en [markdown templates directory](https://github.com/runatlantis/atlantis/tree/main/server/events/templates)

Tenga en cuenta que configuraciones como `--enable-diff-markdown-format` dependen de lógica definida en las plantillas. Es
posible divergir del comportamiento esperado, si no se tiene cuidado al sobrescribir las plantillas predeterminadas.

Por defecto es el directorio home de atlantis `/home/atlantis/.markdown_templates/` en `/$HOME/.markdown_templates`.

### `--max-comments-per-command` <Badge text="v0.32.0+" type="info"/>

```bash
atlantis server --max-comments-per-command=100
# or
ATLANTIS_MAX_COMMENTS_PER_COMMAND=100
```

Limitar el número de comentarios publicados después de ejecutar un comando, para prevenir spam a su VCS y que Atlantis sea throttled como resultado. Por defecto es `100`. Establezca esta opción en `0` para deshabilitar la truncación de logs. Tenga en cuenta que la truncación ocurrirá en la parte superior de la salida del comando, para preservar las partes más importantes de la salida, a menudo mostradas al final.

Cuando la salida del comando excede el límite de tamaño de comentario del VCS (o cuando este límite se aplica), Atlantis divide la salida en múltiples comentarios usando **división inteligente de comentarios**. Los puntos de división se eligen para que la estructura markdown se preserve: el divisor detecta si está dentro de un bloque de código (``

 ``` ``), a `<details>` block, or inline code (`` ` ``), and inserts appropriate closing and continuation markers so that each comment renders correctly. Continuation comments are labeled with the command name (e.g. "Continued plan output from previous comment") when available.

### `--parallel-apply` <Badge text="v0.22.0+" type="info"/>

```bash
atlantis server --parallel-apply
# or
ATLANTIS_PARALLEL_APPLY=true
```

Si ejecutar operaciones apply en paralelo. Por defecto es `false`. La declaración explícita en [repo config](repo-level-atlantis-yaml.md#run-plans-and-applies-in-parallel) tiene precedencia.

### `--parallel-plan` <Badge text="v0.22.0+" type="info"/>

```bash
atlantis server --parallel-plan
# or
ATLANTIS_PARALLEL_PLAN=true
```

Si ejecutar operaciones plan en paralelo. Por defecto es `false`. La declaración explícita en [repo config](repo-level-atlantis-yaml.md#run-plans-and-applies-in-parallel) tiene precedencia.

### `--parallel-pool-size` <Badge text="v0.16.0" type="info"/>

```bash
atlantis server --parallel-pool-size=100
# or
ATLANTIS_PARALLEL_POOL_SIZE=100
```

Tamaño máximo del wait group que ejecuta plans y applies en paralelo (si están habilitados). Por defecto es `15`

### `--pending-apply-status` <Badge text="v0.36.0+" type="info"/>

```bash
atlantis server --pending-apply-status
# or (recommended)
ATLANTIS_PENDING_APPLY_STATUS=true
```

Establecer el estado del commit en pending cuando hay cambios planeados que no han sido aplicados.
Esto evita que merge requests se fusionen hasta que todos los Terraform applies se completen si tiene `Pipelines must succeed` habilitado en su repositorio.

Cuando está habilitado, después de ejecutar `atlantis plan`, el estado del MR se mostrará como pending si hay cambios
por aplicar. Una vez que todos los proyectos se apliquen correctamente (o muestren sin cambios), el estado se actualizará a success.
Los proyectos sin cambios de Terraform se cuentan como actualizados en lugar de aplicados. Si un pull request tiene ambos,
proyectos actualizados y proyectos aún esperando apply, el estado de commit apply de Atlantis permanece pending
hasta que todos los proyectos cambiados sean aplicados.

Por defecto es `false`.

Solo soportado en GitLab

### `--port` <Badge text="v0.1.3+" type="info"/>

```bash
atlantis server --port=4141
# or
ATLANTIS_PORT=4141
```

Puerto al que hacer bind. Por defecto es `4141`.

### `--quiet-policy-checks` <Badge text="v0.32.0+" type="info"/>

```bash
atlantis server --quiet-policy-checks
# or
ATLANTIS_QUIET_POLICY_CHECKS=true
```

Excluir comentarios de policy check de pull requests a menos que haya un error real de conftest. Esto también excluye warnings. Por defecto es `false`.

### `--redis-cluster-addresses`

```bash
atlantis server --redis-cluster-addresses="redis-node-0:6379,redis-node-1:6379,redis-node-2:6379"
# or
ATLANTIS_REDIS_CLUSTER_ADDRESSES="redis-node-0:6379,redis-node-1:6379,redis-node-2:6379"
```

Lista delimitada por comas de direcciones de nodos del cluster Redis en el formato `host:port`. Cuando se establece, Atlantis usa modo Redis Cluster en lugar de modo de nodo único. Esto es mutuamente excluyente con `--redis-host`/`--redis-port` (que se usan para modo de nodo único).

### `--redis-db` <Badge text="v0.19.9+" type="info"/>

```bash
atlantis server --redis-db=0
# or
ATLANTIS_REDIS_DB=0
```

La base de datos Redis a usar cuando se usa un Locking DB type de `redis`. Por defecto es `0`.

### `--redis-host` <Badge text="v0.19.9+" type="info"/>

```bash
atlantis server --redis-host="localhost"
# or
ATLANTIS_REDIS_HOST="localhost"
```

El nombre de host de Redis para cuando se usa un Locking DB type de `redis`.

### `--redis-insecure-skip-verify` <Badge text="v0.19.9+" type="info"/>

```bash
atlantis server --redis-insecure-skip-verify=false
# or
ATLANTIS_REDIS_INSECURE_SKIP_VERIFY=false
```

Controla si el cliente Redis verifica la cadena de certificados y el nombre de host del servidor Redis. Si es true, acepta cualquier certificado presentado por el servidor y cualquier nombre de host en ese certificado. Por defecto es `false`.

::: warning SECURITY WARNING
Si esto está habilitado, TLS es susceptible a ataques machine-in-the-middle a menos que se use verificación personalizada.
:::

### `--redis-password` <Badge text="v0.19.9+" type="info"/>

```bash
atlantis server --redis-password="password123"
# or (recommended)
ATLANTIS_REDIS_PASSWORD="password123"
```

La contraseña Redis para cuando se usa un Locking DB type de `redis`.

### `--redis-port` <Badge text="v0.19.9+" type="info"/>

```bash
atlantis server --redis-port=6379
# or
ATLANTIS_REDIS_PORT=6379
```

El puerto Redis para cuando se usa un Locking DB type de `redis`. Por defecto es `6379`.

### `--redis-tls-enabled` <Badge text="v0.19.9+" type="info"/>

```bash
atlantis server --redis-tls-enabled=false
# or
ATLANTIS_REDIS_TLS_ENABLED=false
```

Habilita una conexión TLS, con versión mínima 1.2, a Redis cuando se usa un Locking DB type de `redis`. Por defecto es `false`.

### `--redis-username`

```bash
atlantis server --redis-username="myuser"
# or
ATLANTIS_REDIS_USERNAME="myuser"
```

El nombre de usuario Redis para cuando se usa un Locking DB type de `redis`. Útil cuando Redis está configurado con autenticación basada en ACL.

### `--repo-allowlist` <Badge text="v0.13.0" type="info"/>

```bash
# NOTE: Use single quotes to avoid shell expansion of *.
atlantis server --repo-allowlist='github.com/myorg/*'
# or
ATLANTIS_REPO_ALLOWLIST='github.com/myorg/*'
```

Atlantis requiere que especifique una allowlist de repositorios de los que aceptará webhooks.

Notas:

- Acepta una lista separada por comas, ej. `definition1,definition2`
- El formato es `{hostname}/{owner}/{repo}`, ej. `github.com/runatlantis/atlantis`
- `*` coincide con cualquier carácter, ej. `github.com/runatlantis/*` coincidirá con todos los repos en la organización runatlantis
- Una entrada que comienza con `!` la niega, ej. `github.com/foo/*,!github.com/foo/bar` coincidirá con todos los repos github en el owner `foo` _excepto_ `bar`.
- Para Bitbucket Server: `{hostname}` es el dominio sin esquema ni puerto, `{owner}` es el nombre del proyecto (no la clave), e `{repo}` es el nombre del repo
  - Los repositorios de usuario (no de proyecto) tienen el formato: `{hostname}/{full name}/{repo}` (p. ej., `bitbucket.example.com/Jane Doe/myatlantis` para nombre de usuario `jdoe` y nombre completo `Jane Doe`, lo cual no es muy intuitivo)
- Para Azure DevOps la allowlist toma una de dos formas: `{owner}.visualstudio.com/{owner}/{project}/{repo}` o `dev.azure.com/{owner}/{project}/{repo}`
- Microsoft está en proceso de cambiar Azure DevOps a la segunda forma, así que puede ser más seguro especificar siempre ambos formatos en su allowlist de repo para cada repositorio hasta que el cambio se complete.

Ejemplos:

- Allowlist `myorg/repo1` e `myorg/repo2` en `github.com`
  - `--repo-allowlist=github.com/myorg/repo1,github.com/myorg/repo2`
- Allowlist de todos los repos bajo `myorg` en `github.com`
  - `--repo-allowlist='github.com/myorg/*'`
- Allowlist de todos los repos bajo `myorg` en `github.com`, excluyendo `myorg/untrusted-repo`
  - `--repo-allowlist='github.com/myorg/*,!github.com/myorg/untrusted-repo'`
- Allowlist de todos los repos en mi instalación GitHub Enterprise
  - `--repo-allowlist='github.yourcompany.com/*'`
- Allowlist de todos los repos bajo el proyecto `myorg` `myproject` en Azure DevOps
  - `--repo-allowlist='myorg.visualstudio.com/myorg/myproject/*,dev.azure.com/myorg/myproject/*'`
- Allowlist de todos los repositorios
  - `--repo-allowlist='*'`

### `--repo-config` <Badge text="v0.5.0+" type="info"/>

```bash
atlantis server --repo-config="path/to/repos.yaml"
# or
ATLANTIS_REPO_CONFIG="path/to/repos.yaml"
```

Ruta a un archivo de configuración YAML de repo del lado del servidor. Vea [Server Side Repo Config](server-side-repo-config.md).

### `--repo-config-json` <Badge text="v0.5.0+" type="info"/>

```bash
atlantis server --repo-config-json='{"repos":[{"id":"/.*/", "apply_requirements":["mergeable"]}]}'
# or
ATLANTIS_REPO_CONFIG_JSON='{"repos":[{"id":"/.*/", "apply_requirements":["mergeable"]}]}'
```

Especifique la configuración de repo del lado del servidor como una cadena JSON. Útil si no quiere escribir un archivo de configuración al disco.
Vea [Server Side Repo Config](server-side-repo-config.md) para más detalles.

::: tip
Si especifica un [Workflow](custom-workflows.md#reference), los [step](custom-workflows.md#step)'s
se pueden especificar como sigue:

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

### `--restrict-file-list` <Badge text="v0.28.0+" type="info"/>

```bash
atlantis server --restrict-file-list
# or (recommended)
ATLANTIS_RESTRICT_FILE_LIST=true
```

`--restrict-file-list` bloqueará solicitudes de plan de proyectos fuera de los archivos modificados en el pull request.
Cuando `--enable-regexp-cmd` también está habilitado, los plans de proyecto regex tales como `atlantis plan -p .*` se limitan a proyectos coincidentes con archivos modificados en el pull request.
Por defecto es `false`.

### `--silence-allowlist-errors` <Badge text="v0.28.0+" type="info"/>

```bash
atlantis server --silence-allowlist-errors
# or
ATLANTIS_SILENCE_ALLOWLIST_ERRORS=true
```

Algunos usuarios usan el flag `--repo-allowlist` para controlar a qué repos responde Atlantis. Normalmente, si Atlantis recibe un webhook de pull request de un repo no listado
en la allowlist, comentará de vuelta con un error. Este flag deshabilita ese comentario.

Algunos usuarios encuentran esto útil porque prefieren agregar el webhook de Atlantis
a nivel de organización en lugar de en cada repo.

### `--silence-fork-pr-errors` <Badge text="v0.28.0+" type="info"/>

```bash
atlantis server --silence-fork-pr-errors
# or
ATLANTIS_SILENCE_FORK_PR_ERRORS=true
```

Normalmente, si Atlantis recibe un webhook de pull request desde un fork y --allow-fork-prs no está establecido,
comentará de vuelta con un error. Este flag deshabilita ese comentario.

### `--silence-no-projects` <Badge text="v0.17.0" type="info"/>

```bash
atlantis server --silence-no-projects
# or
ATLANTIS_SILENCE_NO_PROJECTS=true
```

`--silence-no-projects` le dirá a Atlantis que ignore PRs si ninguno de los archivos modificados es parte de un proyecto definido en el archivo `atlantis.yaml`.
Este flag asegura que un servidor Atlantis solo responda a sus proyectos declarados explícitamente.
Esto no tiene efecto si los proyectos no están definidos en `atlantis.yaml` a nivel repo.
Esto también silencia comandos dirigidos (p. ej. `atlantis plan -d mydir` o `atlantis apply -p myproj`), así que si el proyecto no está en la configuración del repo `atlantis.yaml`, estos comandos no se ejecutarán ni reportarán en un comentario.

Esto es útil al ejecutar múltiples servidores Atlantis contra un solo repositorio para que pueda
delegar trabajo a cada servidor Atlantis. También es útil cuando se usa con pre_workflow_hooks para generar dinámicamente un archivo `atlantis.yaml`.

### `--silence-vcs-status-no-plans` <Badge text="v0.28.0+" type="info"/>

```bash
atlantis server --silence-vcs-status-no-plans
# or
ATLANTIS_SILENCE_VCS_STATUS_NO_PLANS=true
```

`--silence-vcs-status-no-plans` le dirá a Atlantis que ignore establecer el estado VCS en plans si ninguno de los archivos modificados es parte de un proyecto definido en el archivo `atlantis.yaml`.

### `--silence-vcs-status-no-projects` <Badge text="v0.28.0+" type="info"/>

```bash
atlantis server --silence-vcs-status-no-projects
# or
ATLANTIS_SILENCE_VCS_STATUS_NO_PROJECTS=true
```

`--silence-vcs-status-no-projects` le dirá a Atlantis que ignore establecer el estado VCS en cualquier comando si ninguno de los archivos modificados es parte de un proyecto definido en el archivo `atlantis.yaml`.

### `--skip-clone-no-changes` <Badge text="v0.15.0" type="info"/>

```bash
atlantis server --skip-clone-no-changes
# or
ATLANTIS_SKIP_CLONE_NO_CHANGES=true
```

`--skip-clone-no-changes` omitirá clonar el repo durante autoplan si no hay cambios en proyectos Terraform. Esto solo aplicará para GitHub y GitLab y solo para repos que tengan archivo `atlantis.yaml`. Por defecto es `false`.

### `--slack-token` <Badge text="v0.43.0+" type="info"/>

```bash
atlantis server --slack-token=token
# or (recommended)
ATLANTIS_SLACK_TOKEN='token'
```

Token de API para notificaciones de Slack. Vea [Using Slack hooks](sending-notifications-via-webhooks.md#using-slack-hooks).

### `--ssl-cert-file` <Badge text="v0.2.4+" type="info"/>

```bash
atlantis server --ssl-cert-file="/etc/ssl/certs/my-cert.crt"
# or
ATLANTIS_SSL_CERT_FILE="/etc/ssl/certs/my-cert.crt"
```

Archivo que contiene el certificado x509 usado para servir HTTPS.
Si el cert está firmado por una CA, el archivo debe ser la concatenación
del certificado del servidor, cualquier intermedio y el certificado de la CA.

### `--ssl-key-file` <Badge text="v0.2.4+" type="info"/>

```bash
atlantis server --ssl-key-file="/etc/ssl/private/my-cert.key"
# or
ATLANTIS_SSL_KEY_FILE="/etc/ssl/private/my-cert.key"
```

Archivo que contiene la clave privada x509 que coincide con `--ssl-cert-file`.

### `--stats-namespace` <Badge text="v0.43.0+" type="info"/>

```bash
atlantis server --stats-namespace="myatlantis"
# or
ATLANTIS_STATS_NAMESPACE="myatlantis"
```

Namespace para emitir stats/metrics. Vea la sección [stats](stats.md).

### `--tf-distribution` <Badge text="v0.24.0+" type="info"/>

  <Badge text="Deprecated" type="warn"/>
  Deprecated para `--default-tf-distribution`.

### `--tf-download` <Badge text="v0.18.0+" type="info"/>

```bash
atlantis server --tf-download=false
# or
ATLANTIS_TF_DOWNLOAD=false
```

Por defecto es `true`. Permitir que Atlantis liste y descargue versiones adicionales de Terraform.
Establecer esto en `false` puede ser útil en un entorno air-gapped donde un mirror de descarga no está disponible.

### `--tf-download-url` <Badge text="v0.18.0+" type="info"/>

```bash
atlantis server --tf-download-url="https://releases.company.com"
# or
ATLANTIS_TF_DOWNLOAD_URL="https://releases.company.com"
```

Una URL alternativa para descargar versiones de Terraform si faltan. Útil en un entorno airgapped
donde releases.hashicorp.com no está disponible. La estructura de directorios del endpoint personalizado
debe coincidir con la de releases.hashicorp.com.

Esto no tiene impacto si `--tf-download` está establecido en `false`.

Esta configuración aún no es soportada cuando `--tf-distribution` está establecido en `opentofu`.

### `--tfe-hostname` <Badge text="v0.8.3+" type="info"/>

```bash
atlantis server --tfe-hostname="my-terraform-enterprise.company.com"
# or
ATLANTIS_TFE_HOSTNAME="my-terraform-enterprise.company.com"
```

Nombre de host de su instalación de Terraform Enterprise para ser usado junto con
`--tfe-token`. Vea [Terraform Cloud](terraform-cloud.md) para más detalles.
Si usa Terraform Cloud (es decir, no tiene su propia instalación de Terraform Enterprise)
no necesita establecerlo ya que por defecto es `app.terraform.io`.

### `--tfe-local-execution-mode` <Badge text="v0.8.3+" type="info"/>

```bash
atlantis server --tfe-local-execution-mode
# or
ATLANTIS_TFE_LOCAL_EXECUTION_MODE=true
```

Habilite esto si usa modo de ejecución local (en lugar del modo de ejecución remota de TFE/C). Vea [Terraform Cloud](terraform-cloud.md) para más detalles.

### `--tfe-token` <Badge text="v0.8.3+" type="info"/>

```bash
atlantis server --tfe-token="xxx.atlasv1.yyy"
# or (recommended)
ATLANTIS_TFE_TOKEN='xxx.atlasv1.yyy'
```

Un token para integración de Terraform Cloud/Terraform Enterprise. Vea [Terraform Cloud](terraform-cloud.md) para más detalles.

### `--use-tf-plugin-cache` <Badge text="v0.26.0+" type="info"/>

```bash
atlantis server --use-tf-plugin-cache=false
```

Establezca en false si quiere deshabilitar la caché de plugins de terraform.

Este flag es útil cuando se tienen múltiples proyectos que necesitan ejecutar un plan y apply en el mismo PR para evitar la condición de carrera de `plugin_cache_dir` concurrentemente, este es un problema conocido de terraform, más info:

- [plugin_cache_dir concurrently discussion](https://github.com/hashicorp/terraform/issues/31964)
- [PR to improve the situation](https://github.com/hashicorp/terraform/pull/33479)

El efecto de la condición de carrera es más evidente al usar configuración en paralelo para ejecutar plan y apply. Deshabilitar el uso de la caché de plugins impactará el rendimiento al iniciar un nuevo plan o apply, pero en despliegues grandes de Atlantis con múltiples proyectos y módulos compartidos el uso de `--parallel_plan` e `--parallel_apply` es obligatorio para una gestión eficiente de los PRs.

### `--var-file-allowlist` <Badge text="v0.19.5" type="info"/>

```bash
atlantis server --var-file-allowlist='/path/to/tfvars/dir'
# or
ATLANTIS_VAR_FILE_ALLOWLIST='/path/to/tfvars/dir'
```

Lista separada por comas de rutas de directorio adicionales desde las que se pueden leer [variable definition files](https://developer.hashicorp.com/terraform/language/values/variables#variable-definitions-tfvars-files).
Las rutas en este argumento deben ser rutas absolutas. Actualmente no se soportan rutas relativas ni globbing.
Si este argumento no se proporciona, por defecto es el directorio de datos de Atlantis, determinado por el argumento `--data-dir`.

### `--vcs-status-name` <Badge text="v0.42.0+" type="info"/>

```bash
atlantis server --vcs-status-name="atlantis-dev"
# or
ATLANTIS_VCS_STATUS_NAME="atlantis-dev"
```

Nombre usado para identificar Atlantis al actualizar el estado de un pull request. Por defecto es `atlantis`.

Esto es útil al ejecutar múltiples servidores Atlantis contra un solo repositorio para que pueda
dar a cada servidor Atlantis su propio nombre único para evitar que los estados entren en conflicto.

### `--web-basic-auth` <Badge text="v0.1.0+" type="info"/>

```bash
atlantis server --web-basic-auth
# or
ATLANTIS_WEB_BASIC_AUTH=true
```

Habilitar Basic Authentication en el servicio web Atlantis.

### `--web-password` <Badge text="v0.1.0+" type="info"/>

```bash
atlantis server --web-password="atlantis"
# or
ATLANTIS_WEB_PASSWORD="atlantis"
```

Contraseña usada para Basic Authentication en el servicio web Atlantis. Por defecto es `atlantis`.

### `--web-username` <Badge text="v0.1.0+" type="info"/>

```bash
atlantis server --web-username="atlantis"
# or
ATLANTIS_WEB_USERNAME="atlantis"
```

Nombre de usuario usado para Basic Authentication en el servicio web Atlantis. Por defecto es `atlantis`.

### `--webhook-http-headers` <Badge text="v0.35.0+" type="info"/>

```bash
atlantis server --webhook-http-headers='{"Authorization":"Bearer some-token","X-Custom-Header":["value1","value2"]}'
# or
ATLANTIS_WEBHOOK_HTTP_HEADERS='{"Authorization":"Bearer some-token","X-Custom-Header":["value1","value2"]}'
```

Headers adicionales agregados a cada payload HTTP POST al usar [http webhooks](sending-notifications-via-webhooks.md#using-http-webhooks)
proporcionados como una cadena JSON. La clave del map es el nombre del header y el valor es el valor del header
(cadena) o valores (array de cadenas).

### `--websocket-check-origin` <Badge text="v0.19.0+" type="info"/>

```bash
atlantis server --websocket-check-origin
# or
ATLANTIS_WEBSOCKET_CHECK_ORIGIN=true
```

Permitir conexión websockets solo cuando se originan desde el servidor web Atlantis en ejecución

### `--write-git-creds` <Badge text="v0.11.0+" type="info"/>

```bash
atlantis server --write-git-creds
# or
ATLANTIS_WRITE_GIT_CREDS=true
```

Escribir un archivo .git-credentials con el usuario y token del provider para permitir
clonar módulos privados a través de HTTPS o SSH. Vea [Git Credential Store documentation](https://git-scm.com/docs/git-credential-store) para más información.

Siga el `git::ssh`
