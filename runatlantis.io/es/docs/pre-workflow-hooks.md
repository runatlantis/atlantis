# Hooks pre workflow

Los hooks pre workflow pueden definirse para ejecutar scripts antes de que se ejecuten los workflows predeterminados o personalizados. Los hooks pre workflow difieren de los [custom
workflows](custom-workflows.md#custom-run-command) de varias maneras.

1. Los hooks pre workflow no requieren que la configuración del repositorio esté presente. Esto puede utilizarse para [generar dinámicamente configuraciones de repositorio](pre-workflow-hooks.md#dynamic-repo-config-generation).
2. Los hooks pre workflow se ejecutan fuera de los comandos de Atlantis. Esto significa
   que no muestran su salida de vuelta al PR como un comentario.

## Uso

Los hooks pre workflow solo pueden especificarse en la Server-Side Repo Config bajo la
`repos` key.

::: tip Nota
De forma predeterminada, `pre-workflow-hooks` no impiden que Atlantis ejecute sus
workflows(`plan`, `apply`) incluso si un comando `run` termina con un error. Este
comportamiento puede cambiarse estableciendo la bandera [fail-on-pre-workflow-hook-error](server-configuration.md#fail-on-pre-workflow-hook-error)
en la configuración del servidor de Atlantis.
:::

## Segmentación de comandos de Atlantis

De forma predeterminada, el workflow hook se ejecutará cuando Atlantis procese cualquier comando.
Esto puede modificarse especificando la key `commands` en el workflow hook que contiene una lista delimitada por comas
de comandos de Atlantis para los cuales debe ejecutarse el hook. El detalle de los comandos de Atlantis
puede encontrarse en [Using Atlantis](using-atlantis.md).

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

## Personalización del shell

De forma predeterminada, el comando se ejecutará usando el shell 'sh' con un argumento de '-c'. Esto
puede personalizarse usando las keys `shell` y `shellArgs`.

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

Esto es muy similar al [custom workflow run
command](custom-workflows.md#custom-run-command).

```yaml
- run: custom-command
```

| Key         | Type   | Default | Required | Description                                                     |
| ----------- | ------ | ------- | -------- | --------------------------------------------------------------- |
| run         | string | none    | no       | Ejecuta un comando personalizado                                |
| description | string | none    | no       | Descripción del pre hook                                        |
| shell       | string | 'sh'    | no       | El shell que se usará para ejecutar el comando                  |
| shellArgs   | string | '-c'    | no       | Los argumentos del shell que se usarán para ejecutar el comando |

::: tip Notas

* Los comandos `run` se ejecutan con las siguientes variables de entorno:
  * `BASE_REPO_NAME` - Nombre del repositorio en el que se fusionará el pull request, p. ej. `atlantis`.
  * `BASE_REPO_OWNER` - Propietario del repositorio en el que se fusionará el pull request, p. ej. `runatlantis`.
  * `HEAD_REPO_NAME` - Nombre del repositorio que se está fusionando en el repositorio base, p. ej. `atlantis`.
  * `HEAD_REPO_OWNER` - Propietario del repositorio que se está fusionando en el repositorio base, p. ej. `acme-corp`.
  * `HEAD_BRANCH_NAME` - Nombre de la rama head del pull request (la rama que se está fusionando en la base)
  * `HEAD_COMMIT` - El sha256 que apunta al head de la rama que está siendo enviada como pull request a la base. Si el pull request es de Bitbucket Cloud, la cadena tendrá solo 12 caracteres porque Bitbucket Cloud trunca sus IDs de commit.
  * `BASE_BRANCH_NAME` - Nombre de la rama base del pull request (la rama en la que se está fusionando el pull request)
  * `PULL_NUM` - Número o ID del pull request, p. ej. `2`.
  * `PULL_URL` - URL del pull request, p. ej. `https://github.com/runatlantis/atlantis/pull/2`.
  * `PULL_AUTHOR` - Nombre de usuario del autor del pull request, p. ej. `acme-user`.
  * `DIR` - La ruta absoluta a la raíz del repositorio clonado.
  * `USER_NAME` - Nombre de usuario del usuario de VCS que ejecuta el comando, p. ej. `acme-user`. Durante un autoplan, el usuario será el usuario de la API de Atlantis, p. ej. `atlantis`.
  * `COMMENT_ARGS` - Cualquier bandera adicional pasada en el comentario del pull request. Las banderas están separadas por comas y
      cada carácter se escapa, p. ej. `atlantis plan -- arg1 arg2` resultará en `COMMENT_ARGS=\a\r\g\1,\a\r\g\2`.
  * `COMMAND_NAME` - El nombre del comando que se está ejecutando, es decir `plan`, `apply` etc.
  * `OUTPUT_STATUS_FILE` - Un archivo de salida para personalizar el estado de éxito o fallo. p. ej. `echo 'failure' > $OUTPUT_STATUS_FILE`.
  * `PROJECT_NAME` - Nombre del proyecto pasado por la opción `-p`. Si no se proporciona `-p`, este valor está vacío.

:::
