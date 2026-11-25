# Permisos de Repo y Project

A veces puede ser necesario limitar quién puede ejecutar cuáles comandos, como
restringir quién puede apply cambios a producción, mientras se permite más
libertad para los entornos de dev y test.

## Workflow de autorización

Atlantis realiza dos verificaciones de autorización para verificar que un usuario tiene los
permisos necesarios para ejecutar un comando:

1. Después de que un comando ha sido validado, antes de que los archivos var, los metadatos del repo, o los
   estados del pull request sean revisados y validados.
2. Después de que los hooks pre workflow se han ejecutado, la configuración del repo se ha procesado, y los
   projects afectados se han determinado.

::: tip Nota
La primera verificación debe considerarse como una validación del usuario para un repositorio
como un todo, mientras que la segunda verificación es para validar un usuario para un project específico
en ese repo.
:::

### ¿Por qué verificar permisos dos veces?

De la forma en que Atlantis está diseñado actualmente, no toda la información relevante puede estar
disponible cuando ocurre la primera verificación. En particular, los projects afectados
no se conocen porque los hooks pre workflow todavía no se han ejecutado, por lo que los repositorios
que usan hooks para generar o modificar configuraciones de repo no sabrán cuáles
projects verificar permisos para.

## Configuración de permisos

Atlantis tiene dos opciones para permitir a los administradores de la instancia configurar
permisos.

### Opción del servidor [`--gh-team-allowlist`](server-configuration.md#gh-team-allowlist)

La opción `--gh-team-allowlist` permite a los administradores configurar un conjunto global
de permisos que se aplican a todos los repositorios. Para la mayoría de los casos de uso, esto
debería ser suficiente.

::: warning
Si está usando [policy checking](policy-checking.md), también debe permitir en allowlist el comando `policy_check`:

```bash
--gh-team-allowlist="*:plan,*:policy_check,myteam:apply"
```

`policy_check` es un comando interno que se ejecuta automáticamente después de `plan`. Sin permitirlo en allowlist, los comandos manuales `atlantis plan` omitirán las verificaciones de policy (aunque los autoplans seguirán funcionando). Vea [Policy Checking](policy-checking.md#step-1-enable-the-workflow) para detalles.
:::

### Comando externo

Para administradores que requieren definiciones de permisos más granulares y específicas,
se puede definir un comando externo en la [configuración de repo del lado del servidor](server-side-repo-config.md#teamauthz). Este comando recibirá
información sobre el comando, repo, project, y equipos de GitHub de los que el usuario es
miembro, permitiendo a los administradores integrar la validación de permisos
con otros sistemas o requisitos de negocio. Un ejemplo sería permitir a los
usuarios apply cambios a entornos inferiores como los entornos dev y test
mientras se restringen cambios a producción u otros entornos sensibles.

::: warning
Estas opciones son mutuamente excluyentes. Si se define un comando externo,
la opción `--gh-team-allowlist` se ignora.
:::

## Ejemplo

### Restringir cambios de producción

Este ejemplo muestra un ejemplo simple de cómo se podría usar un script para restringir
cambios de producción a un equipo específico, mientras se permite a cualquiera trabajar en otros
entornos. Para brevedad, este ejemplo asume que cada usuario es miembro de un
solo equipo.

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

Los comandos externos se ejecutan en cada verificación de autorización con argumentos y
variables de entorno que contienen contexto sobre el comando que se está verificando. El
comando se ejecuta usando el siguiente formato:

```shell
external_command [external_args...] atlantis_command repo [teams...]
```

| Key                | Optional | Description                                                                                      |
| ------------------ | -------- | ------------------------------------------------------------------------------------------------ |
| `external_command` | no       | Comando definido en [server side repo configuration](server-side-repo-config.md)                 |
| `external_args`    | yes      | Argumentos del comando definidos en [server side repo configuration](server-side-repo-config.md) |
| `atlantis_command` | no       | El comando de atlantis que se está ejecutando (`plan`, `apply`, etc)                             |
| `repo`             | no       | El nombre completo del repo que se está ejecutando (formato: `owner/repo_name`)                  |
| `teams`            | yes      | Una lista de cero o más equipos del usuario que ejecuta el comando                               |

Las siguientes variables de entorno se pasan al comando en cada ejecución:

| Key               | Description                                                                                                                                                    |
| ----------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `BASE_REPO_NAME`  | Nombre del repositorio en el que se hará merge del pull request, ej. `atlantis`.                                                                               |
| `BASE_REPO_OWNER` | Propietario del repositorio en el que se hará merge del pull request, ej. `runatlantis`.                                                                       |
| `COMMAND_NAME`    | El nombre del comando que se está ejecutando, es decir `plan`, `apply` etc.                                                                                    |
| `USER_NAME`       | Nombre de usuario del usuario de VCS que ejecuta el comando, ej. `acme-user`. Durante un autoplan, el usuario será el usuario API de Atlantis, ej. `atlantis`. |

Las siguientes variables de entorno también se pasan al comando cuando se verifica la autorización del project:

| Key                | Description                                                                                                                                                                                                                  |
| ------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `BASE_BRANCH_NAME` | Nombre de la rama base del pull request (la rama en la que se está haciendo merge del pull request)                                                                                                                          |
| `COMMENT_ARGS`     | Cualquier flag adicional pasado en el comentario del pull request. Los flags están separados por comas y cada carácter está escapado, ej. `atlantis plan -- arg1 arg2` resultará en `COMMENT_ARGS=\a\r\g\1,\a\r\g\2`.        |
| `HEAD_REPO_NAME`   | Nombre del repositorio que se está haciendo merge en el repositorio base, ej. `atlantis`.                                                                                                                                    |
| `HEAD_REPO_OWNER`  | Propietario del repositorio que se está haciendo merge en el repositorio base, ej. `acme-corp`.                                                                                                                              |
| `HEAD_BRANCH_NAME` | Nombre de la rama head del pull request (la rama que se está haciendo merge en la base)                                                                                                                                      |
| `HEAD_COMMIT`      | El sha256 que apunta al head de la rama que está siendo enviada en pull request hacia la base. Si el pull request es de Bitbucket Cloud la cadena solo tendrá 12 caracteres porque Bitbucket Cloud trunca sus IDs de commit. |
| `PROJECT_NAME`     | Nombre del project en el que se está ejecutando el comando                                                                                                                                                                   |
| `PULL_NUM`         | Número o ID del pull request, ej. `2`.                                                                                                                                                                                       |
| `PULL_URL`         | URL del pull request, ej. `https://github.com/runatlantis/atlantis/pull/2`.                                                                                                                                                  |
| `PULL_AUTHOR`      | Nombre de usuario del autor del pull request, ej. `acme-user`.                                                                                                                                                               |
| `REPO_ROOT`        | La ruta absoluta a la raíz del repositorio clonado.                                                                                                                                                                          |
| `REPO_REL_PATH`    | Ruta al project relativa a `REPO_ROOT`                                                                                                                                                                                       |

### Manejo del resultado de comando externo

Atlantis determina si un usuario está autorizado para ejecutar el comando solicitado
verificando si el comando externo salió con código `0` y si la última línea
de salida es `pass`.

```text
# Pseudo-code of Atlantis evaluation of external commands

user_authorized =
  external_command.exit_code == 0
  && external_command.output.last_line == 'pass'
```

::: tip

* Un código de salida distinto de cero significa que el comando no pudo evaluar la solicitud por
alguna razón (configuración incorrecta, dependencias faltantes, tormentas solares, etc).
* Si el comando pudo ejecutarse correctamente, pero determinó que el usuario no está
autorizado, aun así debe salir con código `0`.
  * La salida del comando podría contener el razonamiento de la falla de autorización.
:::
