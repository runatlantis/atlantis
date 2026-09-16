# Post Workflow Hooks

Los post workflow hooks pueden definirse para ejecutar scripts después de que se ejecuten workflows predeterminados o personalizados. Los post workflow hooks difieren de los [custom
workflows](custom-workflows.md#custom-run-command) en que se ejecutan
fuera de los comandos de Atlantis. Lo que significa que no muestran su salida
de vuelta al PR como un comentario.

## Uso

Los post workflow hooks solo pueden especificarse en la Server-Side Repo Config bajo
la clave `repos`.

## Atlantis Command Targeting

De forma predeterminada, el workflow hook se ejecutará cuando Atlantis procese cualquier comando.
Esto puede modificarse especificando la clave `commands` en el workflow hook que contiene una lista delimitada por comas
de comandos de Atlantis para los que debe ejecutarse el hook. Los detalles de los comandos de Atlantis
pueden encontrarse en [Using Atlantis](using-atlantis.md).

### Ejemplo

```yaml
repos:
    - id: /.*/
      post_workflow_hooks:
        - run: ./plan-hook.sh
          description: Plan Hook
          commands: plan
        - run: ./plan-apply-hook.sh
          description: Plan & Apply Hook
          commands: plan, apply
```

## Casos de uso

### Reporte de estimación de costos

Puede agregar un post workflow hook para realizar reportes personalizados después de que todos los workflows
hayan terminado.

En este ejemplo usamos un custom workflow para generar estimaciones de costo para cada
workflow usando [Infracost](https://www.infracost.io/docs/integrations/cicd/#cicd-integrations), luego crear un reporte de resumen después de que todos los workflows se hayan completado.

```yaml
# repos.yaml
workflows:
  myworkflow:
    plan:
      steps:
      - init
      - plan
      - run: infracost breakdown --path=$PLANFILE --format=json --out-file=/tmp/$BASE_REPO_OWNER-$BASE_REPO_NAME-$PULL_NUM-$WORKSPACE-$REPO_REL_DIR-infracost.json
repos:
  - id: /.*/
    workflow: myworkflow
    post_workflow_hooks:
      - run: infracost output --path=/tmp/$BASE_REPO_OWNER-$BASE_REPO_NAME-$PULL_NUM-*-infracost.json --format=github-comment --out-file=/tmp/infracost-comment.md
        description: Running infracost
      # Now report the output as desired, e.g. post to GitHub as a comment.
      # ...
```

## Personalizar el shell

De forma predeterminada, los comandos se ejecutarán usando el shell 'sh' con un argumento de '-c'. Esto
puede personalizarse usando las claves `shell` e `shellArgs`.

Ejemplo:

```yaml
repos:
    - id: /.*/
      post_workflow_hooks:
        - run: |
            echo 'atlantis.yaml config:'
            cat atlantis.yaml
          description: atlantis.yaml report
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
| run         | string | none    | no       | Ejecutar un comando personalizado                               |
| description | string | none    | no       | Descripción del post hook                                       |
| shell       | string | 'sh'    | no       | El shell que se usará para ejecutar el comando                  |
| shellArgs   | string | '-c'    | no       | Los argumentos del shell que se usarán para ejecutar el comando |

::: tip Notas

* Los comandos `run` se ejecutan con las siguientes variables de entorno:
  * `BASE_REPO_NAME` - Nombre del repositorio en el que se fusionará el pull request, p. ej. `atlantis`.
  * `BASE_REPO_OWNER` - Propietario del repositorio en el que se fusionará el pull request, p. ej. `runatlantis`.
  * `HEAD_REPO_NAME` - Nombre del repositorio que se está fusionando en el repositorio base, p. ej. `atlantis`.
  * `HEAD_REPO_OWNER` - Propietario del repositorio que se está fusionando en el repositorio base, p. ej. `acme-corp`.
  * `HEAD_BRANCH_NAME` - Nombre de la rama head del pull request (la rama que se está fusionando en la base)
  * `HEAD_COMMIT` - El sha256 que apunta al head de la rama que está siendo enviada como pull request a la base. Si el pull request es de Bitbucket Cloud, la cadena solo tendrá 12 caracteres de longitud porque Bitbucket Cloud trunca sus IDs de commit.
  * `BASE_BRANCH_NAME` - Nombre de la rama base del pull request (la rama en la que el pull request se está fusionando)
  * `PULL_NUM` - Número o ID del pull request, p. ej. `2`.
  * `PULL_URL` - URL del pull request, p. ej. `https://github.com/runatlantis/atlantis/pull/2`.
  * `PULL_AUTHOR` - Nombre de usuario del autor del pull request, p. ej. `acme-user`.
  * `DIR` - La ruta absoluta a la raíz del repositorio clonado.
  * `USER_NAME` - Nombre de usuario del usuario de VCS que ejecuta el comando, p. ej. `acme-user`. Durante un autoplan, el usuario será el usuario de la API de Atlantis, p. ej. `atlantis`.
  * `COMMENT_ARGS` - Cualquier flag adicional pasado en el comentario del pull request. Los flags están separados por comas y
    cada carácter está escapado, p. ej. `atlantis plan -- arg1 arg2` dará como resultado `COMMENT_ARGS=\a\r\g\1,\a\r\g\2`.
  * `COMMAND_NAME` - El nombre del comando que se está ejecutando, es decir `plan`, `apply` etc.
  * `COMMAND_HAS_ERRORS` - Indica si ocurrió algún error durante la ejecución del comando (`plan`, `apply`). Si se establece en `true`, se encontró al menos un error; de lo contrario, es `false`.
  * `OUTPUT_STATUS_FILE` - Un archivo de salida para personalizar el estado de éxito o fallo. p. ej. `echo 'failure' > $OUTPUT_STATUS_FILE`.
  * `PROJECT_NAME` - Nombre del proyecto pasado por la opción `-p`. Si `-p` no se proporciona, este valor está vacío.
:::
