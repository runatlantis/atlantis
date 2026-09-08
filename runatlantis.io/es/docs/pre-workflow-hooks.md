# Hooks de pre workflow

Los hooks de pre workflow se pueden definir para ejecutar scripts antes de que se ejecuten los workflows predeterminados o personalizados. Los hooks de pre workflow difieren de los [custom
workflows](custom-workflows.md#custom-run-command) de varias maneras.

1. Los hooks de pre workflow no requieren que la configuración del repositorio esté
   presente. Esto se puede utilizar para [generar dinámicamente configuraciones de repositorio](pre-workflow-hooks.md#dynamic-repo-config-generation).
2. Los hooks de pre workflow se ejecutan fuera de los comandos de Atlantis. Lo que significa
   que no devuelven su salida al PR como un comentario.

## Uso

Los hooks de pre workflow solo se pueden especificar en la configuración de repositorio del lado del servidor bajo la
`repos` key.

::: tip Nota
De forma predeterminada, `pre-workflow-hooks` no impiden que Atlantis ejecute sus
workflows(`plan`, `apply`) incluso si un comando `run` termina con un error. Este
comportamiento se puede cambiar configurando el flag [fail-on-pre-workflow-hook-error](server-configuration.md#fail-on-pre-workflow-hook-error)
en la configuración del servidor Atlantis.
:::

## Segmentación de comandos de Atlantis

De forma predeterminada, el hook de workflow se ejecutará cuando Atlantis procese cualquier comando.
Esto se puede modificar especificando la key `commands` en el hook de workflow que contiene una lista delimitada por comas
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

### Generación dinámica de configuración de repositorio

Para generar el repositorio `atlantis.yaml` antes de que Atlantis pueda analizarlo,
agregue un comando `run` a `pre_workflow_hooks`. La configuración de su repositorio se generará
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
se puede personalizar usando las keys `shell` y `shellArgs`.

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

Esto es muy similar al [comando run de workflow personalizado
](custom-workflows.md#custom-run-command).

```yaml
- run: custom-command
```

| Key | Type | Default | Required | Description |
| --- | --- | --- | --- | --- |
| run | string | none | no | Ejecuta un comando personalizado |
| description | string | none | no | Descripción del pre hook |
| shell | string | 'sh' | no | El shell que se usará para ejecutar el comando |
| shellArgs | string | '-c' | no | Los argumentos del shell que se usarán para ejecutar el comando |

::: tip Notas

* Los comandos `run` se ejecutan con las siguientes variables de entorno:
  * `BASE_REPO_NAME` - Nombre del repositorio en el que se fusionará el pull request, ej. `atlantis`.
  * `BASE_REPO_OWNER` - Propietario del repositorio en el que se fusionará el pull request, ej. `runatlantis`.
  * `HEAD_REPO_NAME` - Nombre del repositorio que se está fusionando en el repositorio base, ej. `atlantis`.
  * `HEAD_REPO_OWNER` - Propietario del repositorio que se está fusionando en el repositorio base, ej. `acme-corp`.
  * `HEAD_BRANCH_NAME` - Nombre de la rama head del pull request (la rama que se está fusionando en la base)
  * `HEAD_COMMIT` - El sha256 que apunta al head de la rama que se está enviando como pull request a la base. Si el pull request es de Bitbucket Cloud, la cadena solo tendrá 12 caracteres porque Bitbucket Cloud trunca sus IDs de commit.
  * `BASE_BRANCH_NAME` - Nombre de la rama base del pull request (la rama en la que se está fusionando el pull request)
  * `PULL_NUM` - Número o ID del pull request, ej. `2`.
  * `PULL_URL` - URL del pull request, ej. `https://github.com/runatlantis/atlantis/pull/2`.
  * `PULL_AUTHOR` - Nombre de usuario del autor del pull request, ej. `acme-user`.
  * `DIR` - La ruta absoluta a la raíz del repositorio clonado.
  * `USER_NAME` - Nombre de usuario del usuario de VCS que ejecuta el comando, ej. `acme-user`. Durante un autoplan, el usuario será el usuario de la API de Atlantis, ej. `atlantis`.
  * `COMMENT_ARGS` - Cualquier flag adicional pasado en el comentario del pull request. Los flags están separados por comas y
      cada carácter está escapado, ej. `atlantis plan -- arg1 arg2` resultará en `COMMENT_ARGS=\a\r\g\1,\a\r\g\2`.
  * `COMMAND_NAME` - El nombre del comando que se está ejecutando, es decir `plan`, `apply`, etc.
  * `OUTPUT_STATUS_FILE` - Un archivo de salida para personalizar el estado de éxito o fallo. ej. `echo 'failure' > $OUTPUT_STATUS_FILE`.
  * `PROJECT_NAME` - Nombre del proyecto pasado por la opción `-p`. Si no se proporciona `-p`, este valor está vacío.

:::
