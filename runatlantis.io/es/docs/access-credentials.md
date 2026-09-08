# Credenciales de acceso al host Git

Esta página describe cómo crear credenciales para tu host Git (GitHub, GitLab, Gitea, Bitbucket o Azure DevOps)

que Atlantis usará para hacer llamadas API.

## Crear un usuario de Atlantis (opcional)

Recomendamos crear un nuevo usuario llamado **@atlantis** (o algo parecido) o usar un usuario de CI dedicado.

Esto no es obligatorio (puedes usar un usuario existente o credenciales de github app), sin embargo todos los comentarios que Atlantis escribe
vendrán de ese usuario, por lo que podría ser confuso si provienen de una cuenta personal.

![Example Comment](../../docs/images/example-comment.png)

<p align="center"><i>Un comentario de ejemplo proveniente del usuario @atlantisbot</i></p>

## Generar un Access Token

Una vez que hayas creado un nuevo usuario (o decidido usar uno existente), necesitas
generar un access token. Sigue leyendo para ver las instrucciones para tu host Git específico:

* [GitHub](#github-user)
* [GitHub app](#github-app)
* [GitLab](#gitlab)
* [Gitea](#gitea)
* [Bitbucket Cloud (bitbucket.org)](#bitbucket-cloud-bitbucket-org)
* [Bitbucket Server (aka Stash)](#bitbucket-server-aka-stash)
* [Azure DevOps](#azure-devops)

### GitHub user

* Crea un [Personal Access Token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token#creating-a-fine-grained-personal-access-token)
* Crea el token con alcance **repo**
  * Los siguientes permisos de repositorio son el mínimo requerido:
    * Commit statuses: lectura y escritura (para actualizar el PR con indicadores de estado de los trabajos de plan/apply/policy)
    * Contents: solo lectura (para obtener los archivos cambiados y clonar el repositorio)
    * Metadata: solo lectura (esto se seleccionará automáticamente como obligatorio cuando Contents se establezca en solo lectura)
    * Pull requests: lectura y escritura (para comentar y reaccionar en el PR)
* Guarda el access token
::: warning
Tu usuario de Atlantis también debe tener "Write permissions" (para repos en una organización) o ser un "Collaborator" (para repos en una cuenta de usuario) para poder establecer commit statuses:

![Atlantis status](../../docs/images/status.png)

:::

### GitHub app

#### Crear la GitHub App usando Atlantis

::: warning
Disponible en versiones de Atlantis **posteriores** a 0.13.0.
:::

* Inicia Atlantis con un nombre de usuario y token de github falsos (`atlantis server --gh-user fake --gh-token fake --repo-allowlist 'github.com/your-org/*' --atlantis-url https://$ATLANTIS_HOST`). Si instalas como una **Organization**, recuerda agregar `--gh-org your-github-org` a este comando.
* Visita `https://$ATLANTIS_HOST/github-app/setup` y haz clic en **Setup** para crear la app en GitHub. Serás redirigido de vuelta a Atlantis
* En la pantalla se mostrará un enlace para instalar tu app, junto con sus secretos. Guarda las credenciales de tu app e instálala para tu usuario/org siguiendo dicho enlace.
* Crea un archivo con el contenido de la GitHub App Key, por ejemplo `atlantis-app-key.pem`
* Reinicia Atlantis con nuevas flags: `atlantis server --gh-app-id <your id> --gh-app-key-file atlantis-app-key.pem --gh-webhook-secret <your secret> --write-git-creds --repo-allowlist 'github.com/your-org/*' --atlantis-url https://$ATLANTIS_HOST`.

  NOTE: En lugar de usar un archivo para la GitHub App Key, también puedes pasar el valor de la clave directamente usando `--gh-app-key`. También puedes crear un archivo de configuración en lugar de usar flags. Consulta [Server Configuration](server-configuration.md#config-file).

::: warning
Por el momento solo se admite una única instalación por GitHub App.
:::

::: tip NOTE
GitHub App maneja las llamadas de webhook por sí misma, por lo tanto no es necesario crear webhooks por separado. Si los webhooks se crearon manualmente, se pueden eliminar al usar GitHub App. De lo contrario, habría 2 llamadas a Atlantis que darían como resultado errores de bloqueo en path/workspace.

Los webhooks pueden crearse manualmente o ser administrados por la GitHub App para los repositorios que disparan Atlantis. Si los creas manualmente (consulta la [sección de abajo](access-credentials.md#manually-creating-the-github-app)), no especifiques detalles de webhook en la configuración de la GitHub app. En ambos casos se recomienda encarecidamente proteger los webhooks usando un secreto. Consulta [Webhook Secrets](webhook-secrets.md#webhook-secrets)
:::

#### Crear manualmente la GitHub app

* Crea la GitHub app como un Administrator
  * Asegúrate de que la app esté registrada / instalada en la organización / usuario
  * Consulta la [documentation](https://docs.github.com/en/apps/creating-github-apps/about-creating-github-apps/about-creating-github-apps) de GitHub app
* Crea un archivo con el contenido de la GitHub App Key, por ejemplo `atlantis-app-key.pem`
* Inicia Atlantis con las siguientes flags: `atlantis server --gh-app-id <your id> --gh-installation-id <installation id> --gh-app-key-file atlantis-app-key.pem --gh-webhook-secret <your secret> --write-git-creds --repo-allowlist 'github.com/your-org/*' --atlantis-url https://$ATLANTIS_HOST`.

  NOTE: En lugar de usar un archivo para la GitHub App Key, también puedes pasar el valor de la clave directamente usando `--gh-app-key`. También puedes crear un archivo de configuración en lugar de usar flags. Consulta [Server Configuration](server-configuration.md#config-file).

::: tip NOTE
Instalar manualmente la GitHub app significa que las credenciales pueden ser compartidas por muchas instalaciones de Atlantis. Esto tiene el beneficio de centralizar el acceso al repositorio para módulos / código compartidos.
:::

::: tip NOTE
Los repositorios deben registrarse manualmente con la GitHub app creada para permitir que Atlantis interactúe con Pull Requests.
:::

::: tip NOTE
Pasar la flag adicional `--gh-app-slug` modificará el nombre de la App al publicar comentarios en un Pull Request.
:::

#### Permisos

GitHub App necesita estos permisos. Estos se establecen automáticamente cuando se crea una GitHub app.

::: tip NOTE
Desde v0.19.7, se ha agregado un nuevo permiso para `Administration`. Si ya has creado una GitHub app, actualizar Atlantis a v0.19.7 no agregará automáticamente este permiso, por lo que tendrás que establecerlo manualmente.

Desde v0.22.3, se ha agregado un nuevo permiso para `Members`, que es necesario para funciones que aplican permisos a los miembros de un equipo de organización en lugar de a usuarios individuales. Al igual que el permiso `Administration` anterior, actualizar Atlantis no agregará automáticamente este permiso, por lo que si deseas usar funciones que dependen de verificar la pertenencia a equipos, tendrás que agregarlo manualmente.

Desde v0.30.0, se ha agregado un nuevo permiso para `Actions`, que es necesario para comprobar si un pull request se puede fusionar mientras se omite la verificación de apply. Actualizar Atlantis no agregará automáticamente este permiso, por lo que tendrás que agregarlo manualmente.
:::

| Tipo            | Acceso              |
| --------------- | ------------------- |
| Administration  | Solo lectura        |
| Checks          | Lectura y escritura |
| Commit statuses | Lectura y escritura |
| Contents        | Lectura y escritura |
| Issues          | Lectura y escritura |
| Metadata        | Solo lectura (predeterminado) |
| Pull requests   | Lectura y escritura |
| Webhooks        | Lectura y escritura |
| Members         | Solo lectura        |
| Actions         | Solo lectura        |

### GitLab

* Sigue: [GitLab: Create a personal access token](https://docs.gitlab.com/user/profile/personal_access_tokens/#create-a-personal-access-token)
* Crea un token con alcance **api**
* Guarda el access token

### Gitea

* Ve a "Profile and Settings" > "Settings" en Gitea (arriba a la derecha)
* Ve a "Applications" bajo "User Settings" en Gitea
* Crea un token bajo "Manage Access Tokens" con los siguientes permisos:
  * issue: Read and Write
  * repository: Read and Write
  * user: Read
* Guarda el access token

### Bitbucket Cloud (bitbucket.org)

* Crea un App Password siguiendo [BitBucket Cloud: Create an app password](https://support.atlassian.com/bitbucket-cloud/docs/create-an-app-password/)
* Etiqueta la contraseña como "atlantis"
* Selecciona **Pull requests**: **Read** y **Write** para que Atlantis pueda leer tus pull requests y escribir comentarios en ellos. Si quieres habilitar la función [hide-prev-plan-comments](server-configuration.md#hide-prev-plan-comments) y por lo tanto eliminar comentarios antiguos, agrega también **Account**: **Read**.
* Guarda el access token

### Bitbucket Server (aka Stash)

* Haz clic en tu avatar en la parte superior derecha y selecciona **Manage account**
* Haz clic en **Personal access tokens** en la barra lateral
* Haz clic en **Create a token**
* Nombra el token **atlantis**
* Dale al token permisos de Project **Read** y permisos de Pull request **Write**
* Haz clic en **Create** y guarda el access token

  NOTE: Atlantis enviará el token como un [Bearer Auth to the Bitbucket API](https://confluence.atlassian.com/bitbucketserver/http-access-tokens-939515499.html#HTTPaccesstokens-UsingHTTPaccesstokens) en lugar de usar Basic Auth.

### Azure DevOps

* Crea un Personal access token siguiendo [Azure DevOps: Use personal access tokens to authenticate](https://docs.microsoft.com/en-us/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate?view=azure-devops)
* Etiqueta la contraseña como "atlantis"
* Los alcances mínimos requeridos para este token son:
  * Code (Read & Write)
  * Code (Status)
  * Member Entitlement Management (Read)
* Guarda el access token

## Siguientes pasos

Una vez que tengas tu usuario y access token, estarás listo para crear un secreto de webhook. Consulta [Creating a Webhook Secret](webhook-secrets.md).
