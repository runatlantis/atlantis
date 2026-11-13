# Configuración de Webhooks

Atlantis necesita recibir Webhooks de tu host Git para que pueda responder a eventos de pull request.

:::tip Requisitos previos

* Has creado una [access credential](access-credentials.md)
* Has creado un [webhook secret](webhook-secrets.md)
* Has [deployed](deployment.md) Atlantis y tienes una url para él
:::

Consulta las instrucciones para tu proveedor específico a continuación.

## GitHub/GitHub Enterprise

Puedes instalar tu webhook en el nivel de [organization](https://docs.github.com/en/get-started/learning-about-github/types-of-github-accounts), o para cada repositorio individual.

::: tip NOTE
Si solo algunos de los repos de tu organization serán administrados por Atlantis, entonces
puede que quieras instalarlo solo en repos específicos por ahora.
:::

Al autenticarte como una GitHub App, los Webhooks se crean automáticamente y no necesitan configuración adicional, más allá de ser instalados en tu cuenta de organization/usuario después de la creación. Consulta la sección [GitHub App setup](access-credentials.md#github-app) para obtener instrucciones sobre cómo hacerlo.

Si estás instalando en la organization, navega a la página de tu organization y haz clic en **Settings**.
Si estás instalando en un solo repositorio, navega a la página principal del repositorio y haz clic en **Settings**.

* Selecciona **Webhooks** o **Hooks** en la barra lateral
* Haz clic en **Add webhook**
* configura **Payload URL** como `http://$URL/events` (o `https://$URL/events` si estás usando SSL) donde `$URL` es donde Atlantis está alojado. **Asegúrate de agregar `/events`**
* verifica de nuevo que agregaste `/events` al final de tu URL.
* configura **Content type** como `application/json`
* configura **Secret** con el Webhook Secret que generaste previamente
  * **NOTE** Si estás agregando un webhook a múltiples repositorios, cada repositorio necesitará usar el **mismo** secret.
* selecciona **Let me select individual events**
* marca las casillas
  * **Pull request reviews**
  * **Pushes**
  * **Issue comments**
  * **Pull requests**
* deja **Active** marcado
* haz clic en **Add webhook**
* Consulta [Next Steps](#next-steps)

## GitLab

Si estás usando GitLab, navega a la página principal de tu proyecto en GitLab

* Haz clic en **Settings > Webhooks** en la barra lateral
* configura **URL** como `http://$URL/events` (o `https://$URL/events` si estás usando SSL) donde `$URL` es donde Atlantis está alojado. **Asegúrate de agregar `/events`**
* verifica de nuevo que agregaste `/events` al final de tu URL.
* configura **Secret Token** con el Webhook Secret que generaste previamente
  * **NOTE** Si estás agregando un webhook a múltiples repositorios, cada repositorio necesitará usar el **mismo** secret.
* marca las casillas
  * **Push events**
  * **Comments**
  * **Merge Request events**
* deja **Enable SSL verification** marcado
* haz clic en **Add webhook**
* Consulta [Next Steps](#next-steps)

## Gitea

Si estás usando Gitea, navega a la página principal de tu proyecto en Gitea

* Haz clic en **Settings > Webhooks** en la barra superior y luego en la barra lateral
* Haz clic en **Add webhook > Gitea** (los webhooks de Gitea son específicos del servicio, pero esto funciona)
* configura **Target URL** como `http://$URL/events` (o `https://$URL/events` si estás usando SSL) donde `$URL` es donde Atlantis está alojado. **Asegúrate de agregar `/events`**
* verifica de nuevo que agregaste `/events` al final de tu URL.
* configura **Secret** con el Webhook Secret que generaste previamente
  * **NOTE** Si estás agregando un webhook a múltiples repositorios, cada repositorio necesitará usar el **mismo** secret.
* Selecciona **Custom Events...**
* Marca las casillas
  * **Repository events > Push**
  * **Issue events > Issue Comment**
  * **Pull Request events > Pull Request**
  * **Pull Request events > Pull Request Comment**
  * **Pull Request events > Pull Request Reviewed**
  * **Pull Request events > Pull Request Synchronized**
* Deja **Active** marcado
* Haz clic en **Add Webhook**
* Consulta [Next Steps](#next-steps)

## Bitbucket Cloud (bitbucket.org)

* Ve a la página principal de tu repo
* Haz clic en **Settings** en la barra lateral
* Haz clic en **Webhooks** en la sección **WORKFLOW**
* Haz clic en **Add webhook**
* Ingresa "Atlantis" en **Title**
* configura **URL** como `http://$URL/events` (o `https://$URL/events` si estás usando SSL) donde `$URL` es donde Atlantis está alojado. **Asegúrate de agregar `/events`**
* verifica de nuevo que agregaste `/events` al final de tu URL.
* Mantén **Status** como Active
* No marques **Skip certificate validation** porque NGROK tiene un certificado válido.
* Selecciona **Choose from a full list of triggers**
* En **Repository** desmarca todo
* En **Issues** deja todo desmarcado
* En **Pull Request**, selecciona: Created, Updated, Merged, Declined y Comment created
* Haz clic en **Save**
<img src="../../guide/images/bitbucket-webhook.png" alt="Bitbucket Webhook" style="max-height: 500px">
* Consulta [Next Steps](#next-steps)

## Bitbucket Server (aka Stash)

* Ve a la página principal de tu repo
* Haz clic en **Settings** en la barra lateral
* Haz clic en **Webhooks** en la sección **WORKFLOW**
* Haz clic en **Create webhook**
* Ingresa "Atlantis" en **Name**
* configura **URL** como `http://$URL/events` (o `https://$URL/events` si estás usando SSL) donde `$URL` es donde Atlantis está alojado. **Asegúrate de agregar `/events`**
* Verifica de nuevo que agregaste `/events` al final de tu URL.
* Configura **Secret** con el Webhook Secret que generaste previamente
  * **NOTE** Si estás agregando un webhook a múltiples repositorios, cada repositorio necesitará usar el **mismo** secret.
* En **Pull Request**, selecciona: Opened, Source branch updated, Merged, Declined, Deleted y Comment added
* Haz clic en **Save**<img src="../../guide/images/bitbucket-server-webhook.png" alt="Bitbucket Webhook" style="max-height: 600px;">
* Consulta [Next Steps](#next-steps)

## Azure DevOps

Los Webhooks se instalan en el nivel de [team project](https://docs.microsoft.com/en-us/azure/devops/organizations/projects/about-projects?view=azure-devops), pero pueden restringirse para activarse solo según eventos pertenecientes a [specific repos](https://docs.microsoft.com/en-us/azure/devops/service-hooks/services/webhooks?view=azure-devops) dentro del team project.

* Navega a cualquier lugar dentro de un team project, por ejemplo: `https://dev.azure.com/orgName/projectName/_git/repoName`
* Selecciona **Project settings** en la esquina inferior izquierda
* Selecciona **Service hooks**
  * Si ves el mensaje "You do not have sufficient permissions to view or configure subscriptions." necesitas asegurarte de que tu usuario sea miembro ya sea del grupo "Project Collection Administrators" de la organization o del grupo "Project Administrators" del proyecto.
  * Para agregar tu usuario al grupo Project Collection Build Administrators, navega al nivel de organization, haz clic en **Organization Settings** y luego haz clic en **Permissions**. Deberías estar en `https://dev.azure.com/<organization>/_settings/groups`. Ahora haz clic en el grupo **\<organization\>/Project Collection Administrators** y agrega tu usuario como miembro.
  * Para agregar tu usuario al grupo Project Administrators, navega al nivel del proyecto, haz clic en **Project Settings** y luego haz clic en **Permissions**. Deberías estar en `https://dev.azure.com/<organization>/<project>/_settings/permissions`. Ahora haz clic en el grupo **\<project\>/Project Administrators** y agrega tu usuario como miembro.
* Haz clic en **Create subscription** o en el ícono verde de más para agregar un nuevo webhook
* Desplázate hasta la parte inferior de la lista y selecciona **Web Hooks**
* Haz clic en **Next**
* En "Trigger on this type of event", selecciona **Pull request created**
  * Opcionalmente, selecciona un repositorio en **Filters** para restringir el alcance de esta suscripción de webhook a un repositorio específico
* Haz clic en **Next**
* Configura **URL** como `http://$URL/events` donde `$URL` es donde Atlantis está alojado. Ten en cuenta que SSL, o `https://$URL/events`, es requerido si configuras un nombre de usuario y contraseña Basic para el webhook. **Asegúrate de agregar `/events`**
* Se recomienda fuertemente configurar un Basic Username y Password para todos los webhooks
* Deja los tres menús desplegables para `...to send` configurados en **All**
* Resource version debe configurarse en **1.0** para los tipos de evento `Pull request created` e `Pull request updated` y en **2.0** para `Pull request commented on`
* **NOTE** Si estás agregando un webhook a múltiples team projects o repositorios (usando filtros), cada repositorio necesitará usar el **mismo** nombre de usuario y contraseña Basic.
* Haz clic en **Finish**

Repite el proceso anterior hasta que tengas suscripciones de webhook para los siguientes tipos de evento que se activarán en todos los repositorios que Atlantis administrará:

* Pull request created (acabas de agregar este)
* Pull request updated
* Pull request commented on

* Consulta [Next Steps](#next-steps)

## Next Steps

* Para verificar que Atlantis está recibiendo tus webhooks, crea un pull request de prueba a tu repo.
* Deberías ver la solicitud aparecer en los logs de Atlantis en un nivel `INFO`.
* Ahora necesitarás configurar Atlantis para agregar tus [Provider Credentials](provider-credentials.md)
