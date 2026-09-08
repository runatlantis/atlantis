# Permisos de Repo y Proyecto

A veces puede ser necesario limitar quién puede ejecutar qué comandos, como
restringir quién puede apply cambios a producción, mientras se permite más
libertad para los entornos de dev y test.

## Workflow de autorización

Atlantis realiza dos comprobaciones de autorización para verificar que un usuario tiene los permisos necesarios
para ejecutar un comando:

1. Después de que un comando ha sido validado, antes de que se comprueben y validen los archivos var, los metadatos del repo o los estados del
   pull request.
2. Después de que los hooks pre workflow se hayan ejecutado, se haya procesado la configuración del repo y se hayan determinado los proyectos
   afectados.

::: tip Nota
La primera comprobación debe considerarse como la validación del usuario para un repositorio
como un todo, mientras que la segunda comprobación es para validar a un usuario para un proyecto
específico en ese repo.
:::

### ¿Por qué comprobar los permisos dos veces?

Por la manera en que Atlantis está diseñado actualmente, no toda la información relevante puede estar
disponible cuando ocurre la primera comprobación. En particular, los proyectos afectados
no se conocen porque los hooks pre workflow todavía no se han ejecutado, por lo que los repositorios
que usan hooks para generar o modificar configuraciones de repo no sabrán para qué
proyectos comprobar permisos.

## Configuración de permisos

Atlantis tiene dos opciones para permitir que los administradores de la instancia configuren
permisos.

### Opción del servidor [`--gh-team-allowlist`](server-configuration.md#gh-team-allowlist)

La opción `--gh-team-allowlist` permite a los administradores configurar un conjunto global
de permisos que se aplican a todos los repositorios. Para la mayoría de los casos de uso, esto
debería ser suficiente.

::: warning
Si está usando [policy checking](policy-checking.md), también debe allowlist el comando `policy_check`:

```bash
--gh-team-allowlist="*:plan,*:policy_check,myteam:apply"
```

`policy_check` es un comando interno que se ejecuta automáticamente después de `plan`. Sin hacer allowlist de él, los comandos manuales `atlantis plan` omitirán las comprobaciones de policy (aunque los autoplans seguirán funcionando). Vea [Policy Checking](policy-checking.md#step-1-enable-the-workflow) para más detalles.
:::

### Comando externo

Para administradores que requieren definiciones de permisos más granulares y específicas,
se puede definir un comando externo en la [configuración del repo del lado del servidor
](server-side-repo-config.md#teamauthz). Este comando recibirá
información sobre el comando, repo, proyecto y equipos de GitHub de los que el usuario es
miembro, lo que permite a los administradores integrar la validación de permisos
con otros sistemas o requisitos del negocio. Un ejemplo sería permitir que los
usuarios hagan apply de cambios a entornos inferiores como los entornos dev y test
mientras se restringen cambios a producción u otros entornos sensibles.

::: warning
Estas opciones son mutuamente excluyentes. Si se define un comando externo,
se ignora la opción `--gh-team-allowlist`.
:::

## Ejemplo

### Restringir cambios en producción

Este ejemplo muestra un ejemplo simple de cómo se podría usar un script para restringir
los cambios en producción a un equipo específico, mientras se permite que cualquiera trabaje en otros
entornos. Por brevedad, este ejemplo asume que cada usuario es miembro de un
único equipo.

`server-side-repo-config.yaml`

```yaml
team_authz:
  command: "/scripts/example.sh"
```

`example.sh`

```shell
#!/bin/bash

# Define name of team allowed to make production changes
PROD_TEAM="example-org/prod-deployers"

# Set variables from command-line arguments for convenience
COMMAND="$1"
REPO="$2"
TEAM="$3"

# Check if we are running the 'apply' command on prod
if [ "${COMMAND}" == "apply" -a "${PROJECT_NAME}" == "prod" ]
then
   # Only the prod team can make this change
   if [ "${TEAM}" == "${PROD_TEAM}" ]
   then
      echo "pass"
      exit 0
   fi

   # Print reason for failing and exit
   echo "user \"${USER_NAME}\" must be a member of \"${PROD_TEAM}\" to apply changes to production."
   exit 0
fi

# Any other command and environment is okay
echo "pass"
exit 0
```

## Referencia

### Ejecución de comando externo

Los comandos externos se ejecutan en cada comprobación de autorización con argumentos y
variables de entorno que contienen contexto sobre el comando que se está comprobando. El
comando se ejecuta usando el siguiente formato:

```shell
external_command [external_args...] atlantis_command repo [teams...]
```

| Key                | Optional | Description                                                                               |
|--------------------|----------|-------------------------------------------------------------------------------------------|
| `external_command` | no       | Comando definido en la [configuración del repo del lado del servidor](server-side-repo-config.md)           |
| `external_args`    | yes      | Argumentos del comando definidos en la [configuración del repo del lado del servidor](server-side-repo-config.md) |
| `atlantis_command` | no       | El comando de atlantis que se está ejecutando (`plan`, `apply`, etc.)                                     |
| `repo`             | no       | El nombre completo del repo que se está ejecutando (formato: `owner/repo_name`)                      |
| `teams`            | yes      | Una lista de cero o más equipos del usuario que ejecuta el comando                            |

Las siguientes variables de entorno se pasan al comando en cada ejecución:

| Key                  | Description                                                                                                                                                                                                                           |
|----------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `BASE_REPO_NAME`     | Nombre del repositorio en el que se hará merge del pull request, p. ej. `atlantis`.                                                                                                                                                     |
| `BASE_REPO_OWNER`    | Propietario del repositorio en el que se hará merge del pull request, p. ej. `runatlantis`.                                                                                                                                                 |
| `COMMAND_NAME`       | El nombre del comando que se está ejecutando, es decir, `plan`, `apply`, etc.                                                                                                                                                             |
| `USER_NAME`          | Nombre de usuario del usuario de VCS que ejecuta el comando, p. ej. `acme-user`. Durante un autoplan, el usuario será el usuario de la API de Atlantis, p. ej. `atlantis`.                                                                                                |

Las siguientes variables de entorno también se pasan al comando cuando se comprueba la autorización del proyecto:

| Key                  | Description                                                                                                                                                                                                                           |
|----------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `BASE_BRANCH_NAME`   | Nombre de la rama base del pull request (la rama en la que se hace merge del pull request)                                                                                                                                 |
| `COMMENT_ARGS`       | Cualquier flag adicional pasado en el comentario del pull request. Los flags están separados por comas y cada carácter está escapado, p. ej. `atlantis plan -- arg1 arg2` dará como resultado `COMMENT_ARGS=\a\r\g\1,\a\r\g\2`.                       |
| `HEAD_REPO_NAME`     | Nombre del repositorio que se está haciendo merge en el repositorio base, p. ej. `atlantis`.                                                                                                                                               |
| `HEAD_REPO_OWNER`    | Propietario del repositorio que se está haciendo merge en el repositorio base, p. ej. `acme-corp`.                                                                                                                                             |
| `HEAD_BRANCH_NAME`   | Nombre de la rama head del pull request (la rama que se está haciendo merge en la base)                                                                                                                                         |
| `HEAD_COMMIT`        | El sha256 que apunta a la head de la rama que está siendo enviada en pull request a la base. Si el pull request es de Bitbucket Cloud, la cadena solo tendrá 12 caracteres porque Bitbucket Cloud trunca sus IDs de commit. |
| `PROJECT_NAME`       | Nombre del proyecto sobre el que se está ejecutando el comando                                                                                                                                                                                  |
| `PULL_NUM`           | Número o ID del pull request, p. ej. `2`.                                                                                                                                                                                                   |
| `PULL_URL`           | URL del pull request, p. ej. `https://github.com/runatlantis/atlantis/pull/2`.                                                                                                                                                               |
| `PULL_AUTHOR`        | Nombre de usuario del autor del pull request, p. ej. `acme-user`.                                                                                                                                                                                 |
| `REPO_ROOT`          | La ruta absoluta a la raíz del repositorio clonado.                                                                                                                                                                               |
| `REPO_REL_PATH`      | Ruta al proyecto relativa a `REPO_ROOT`                                                                                                                                                                                           |

### Manejo del resultado del comando externo

Atlantis determina si un usuario está autorizado para ejecutar el comando solicitado
comprobando si el comando externo terminó con el código `0` y si la última línea
de salida es `pass`.

```text
# Pseudo-code of Atlantis evaluation of external commands

user_authorized =
  external_command.exit_code == 0
  && external_command.output.last_line == 'pass'
```

::: tip

* Un código de salida distinto de cero significa que el comando no pudo evaluar la solicitud por
alguna razón (mala configuración, dependencias faltantes, tormentas solares, etc.).
* Si el comando pudo ejecutarse correctamente, pero determinó que el usuario no está
autorizado, aun así debe salir con el código `0`.
  * La salida del comando podría contener la razón de la falla de autorización.
:::
