# Pre Workflow Hooks

Los hooks de pre workflow se pueden definir para ejecutar scripts antes de que se ejecuten los workflows predeterminados o personalizados. Los hooks de pre workflow difieren de los [custom
workflows](custom-workflows.md#custom-run-command) de varias maneras.

1. Los hooks de pre workflow no requieren que la configuración del repositorio esté presente. Esto se puede utilizar para [generar dinámicamente configuraciones de repositorio](pre-workflow-hooks.md#dynamic-repo-config-generation).
2. Los hooks de pre workflow se ejecutan fuera de los comandos de Atlantis. Esto significa que no muestran su salida de vuelta en el PR como un comentario.

## Uso

Los hooks de pre workflow solo se pueden especificar en la Server-Side Repo Config bajo la clave
`repos`.

::: tip Nota
De forma predeterminada, `pre-workflow-hooks` no impiden que Atlantis ejecute sus
workflows(`plan`, `apply`) incluso si un comando `run` termina con un error. Este
comportamiento se puede cambiar estableciendo la opción [fail-on-pre-workflow-hook-error](server-configuration.md#fail-on-pre-workflow-hook-error)
en la configuración del servidor de Atlantis.
:::

## Targeting de comandos de Atlantis

De forma predeterminada, el workflow hook se ejecutará cuando cualquier comando sea procesado por Atlantis.
Esto se puede modificar especificando la clave `commands` en el workflow hook que contiene una lista delimitada por comas
de comandos de Atlantis para los que se debe ejecutar el hook. Los detalles de los comandos de Atlantis
se pueden encontrar en [Using Atlantis](using-atlantis.md).

### Ejemplo

```yaml
repos:
    - id: /.*/
      pre_workflow_hooks:
        - run: ./plan-hook.sh
          description: Plan Hook
          commands: plan
        - run: ./plan-apply-hook.sh
          description: Plan & Apply Hook
          commands: plan, apply
```

## Casos de uso

### Generación dinámica de Repo Config

Para generar el repo `atlantis.yaml` antes de que Atlantis pueda analizarlo,
agregue un comando `run` a `pre_workflow_hooks`. Su Repo config se generará
justo antes de que Atlantis la analice.

```yaml
repos:
    - id: /.*/
      pre_workflow_hooks:
        - run: ./repo-config-generator.sh
          description: Generating configs
```

## Personalizar el shell

De forma predeterminada, el comando se ejecutará usando el shell 'sh' con un argumento de '-c'. Esto
se puede personalizar usando las claves `shell` e `shellArgs`.

Ejemplo:

```yaml
repos:
    - id: /.*/
      pre_workflow_hooks:
        - run: |
            echo "generating atlantis.yaml"
            terragrunt-atlantis-config generate --output atlantis.yaml --autoplan --parallel
          description: Generating atlantis.yaml
          shell: bash
          shellArgs: -cv
```

## Referencia

### Comando `run` personalizado

Esto es muy similar al [comando run de custom workflow](custom-workflows.md#custom-run-command).

```yaml
- run: custom-command
```

| Key | Type | Default | Required | Description |
| --- | --- | --- | --- | --- |
| run | string | none | no | Ejecuta un comando personalizado |
| description | string | none | no | Descripción del pre hook |
| shell | string | 'sh' | no | El shell a usar para ejecutar el comando |
| shellArgs | string | '-c' | no | Los argumentos del shell a usar para ejecutar el comando |

::: tip Notas

* Los comandos `run` se ejecutan con las siguientes variables de entorno:
  * `BASE_REPO_NAME` - Nombre del repositorio en el que se fusionará el pull request, p. ej. `atlantis`.
  * `BASE_REPO_OWNER` - Propietario del repositorio en el que se fusionará el pull request, p. ej. `runatlantis`.
  * `HEAD_REPO_NAME` - Nombre del repositorio que se está fusionando en el repositorio base, p. ej. `atlantis`.
  * `HEAD_REPO_OWNER` - Propietario del repositorio que se está fusionando en el repositorio base, p. ej. `acme-corp`.
  * `HEAD_BRANCH_NAME` - Nombre de la rama head del pull request (la rama que se está fusionando en la base)
  * `HEAD_COMMIT` - El sha256 que apunta al head de la rama que se está enviando como pull request a la base. Si el pull request es de Bitbucket Cloud, la cadena solo tendrá 12 caracteres porque Bitbucket Cloud trunca sus IDs de commit.
  * `BASE_BRANCH_NAME` - Nombre de la rama base del pull request (la rama en la que se está fusionando el pull request)
  * `PULL_NUM` - Número o ID del pull request, p. ej. `2`.
  * `PULL_URL` - URL del pull request, p. ej. `https://github.com/runatlantis/atlantis/pull/2`.
  * `PULL_AUTHOR` - Nombre de usuario del autor del pull request, p. ej. `acme-user`.
  * `DIR` - La ruta absoluta a la raíz del repositorio clonado.
  * `USER_NAME` - Nombre de usuario del usuario de VCS que ejecuta el comando, p. ej. `acme-user`. Durante un autoplan, el usuario será el usuario de API de Atlantis, p. ej. `atlantis`.
  * `COMMENT_ARGS` - Cualquier opción adicional pasada en el comentario del pull request. Las opciones están separadas por comas y
      cada carácter es escapado, p. ej. `atlantis plan -- arg1 arg2` dará como resultado `COMMENT_ARGS=\a\r\g\1,\a\r\g\2`.
  * `COMMAND_NAME` - El nombre del comando que se está ejecutando, es decir `plan`, `apply`, etc.
  * `OUTPUT_STATUS_FILE` - Un archivo de salida para personalizar el estado de éxito o fallo. p. ej. `echo 'failure' > $OUTPUT_STATUS_FILE`.
  * `PROJECT_NAME` - Nombre del proyecto pasado por la opción `-p`. Si `-p` no se proporciona, este valor está vacío.

:::
