# Credenciales del proveedor

Atlantis ejecuta Terraform simplemente ejecutando `terraform plan` e `apply` comandos
en el servidor en el que Atlantis está alojado.
Igual que cuando ejecutas Terraform localmente, Atlantis necesita credenciales para tu
proveedor específico.

Depende de ti cómo proporcionas credenciales para tu proveedor específico a Atlantis:

* El [Helm Chart](deployment.md#kubernetes-helm-chart) y
    [AWS Fargate Module](deployment.md#aws-fargate) de Atlantis tienen sus propios mecanismos para las credenciales del proveedor.
    Lee su documentación.
* Si estás ejecutando Atlantis en una nube, entonces muchas nubes tienen formas de dar acceso a la API de la nube
  a las aplicaciones que se ejecutan en ellas, p. ej.:
  * [AWS EC2 Roles](https://registry.terraform.io/providers/hashicorp/aws/latest/docs) (Busca "EC2 Role")
  * [GCE Instance Service Accounts](https://registry.terraform.io/providers/hashicorp/google/latest/docs/guides/provider_reference)
* Muchos usuarios establecen variables de entorno, p. ej. `AWS_ACCESS_KEY`, donde Atlantis se está ejecutando.
* Otros crean los archivos de configuración necesarios, p. ej. `~/.aws/credentials`, donde Atlantis se está ejecutando.
* Usa el [HashiCorp Vault Provider](https://registry.terraform.io/providers/hashicorp/vault/latest/docs)
  para obtener credenciales del proveedor.

:::tip
Como regla general, si puedes `ssh` o `exec` en el servidor donde Atlantis se está
ejecutando y ejecutar `terraform` comandos como lo harías localmente, entonces Atlantis funcionará.
:::

## Información específica de AWS

### Múltiples cuentas de AWS

Atlantis soporta múltiples cuentas de AWS mediante el uso de la
[AWS Authentication](https://registry.terraform.io/providers/hashicorp/aws/latest/docs) de Terraform (Busca "Authentication").

Si estás usando el [Shared Credentials file](https://registry.terraform.io/providers/hashicorp/aws/latest/docs) (Busca "Shared Credentials file")
tendrás que asegurarte de que el servidor en el que Atlantis se está ejecutando tenga el archivo de credenciales correspondiente.

Si estás usando [Assume role](https://registry.terraform.io/providers/hashicorp/aws/latest/docs) (Busca "Assume role")
tendrás que asegurarte de que el archivo de credenciales tenga un perfil `default` que sea capaz
de asumir todos los roles requeridos.

Usar múltiples [Environment variables](https://registry.terraform.io/providers/hashicorp/aws/latest/docs) (Busca "Environment variables")
no funcionará para múltiples cuentas porque Atlantis no sabría con qué variables de entorno ejecutar
Terraform.

### Nombres de sesión de Assume Role

Si estás usando Terraform < 0.12, Atlantis inyecta 5 variables de Terraform que pueden usarse para nombrar dinámicamente el nombre de la sesión assume role.
Establecer `session_name` te permite rastrear las llamadas a la API realizadas a través de Atlantis hasta un
usuario y repo específicos mediante CloudWatch:

```bash
provider "aws" {
  assume_role {
    role_arn     = "arn:aws:iam::ACCOUNT_ID:role/ROLE_NAME"
    session_name = "${var.atlantis_user}-${var.atlantis_repo_owner}-${var.atlantis_repo_name}-${var.atlantis_pull_num}"
  }
}
```

Atlantis ejecuta `terraform` con las siguientes variables:

| `-var` Argumento                    | Descripción                                                                                                                            |
|--------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------|
| `atlantis_user=lkysow`               | El nombre de usuario del VCS de quien está ejecutando el comando plan.                                                                 |
| `atlantis_repo=runatlantis/atlantis` | El nombre completo del repo en el que está el pull request. NOTA: Esta variable no puede usarse en el nombre de sesión de AWS porque contiene un `/`. |
| `atlantis_repo_owner=runatlantis`    | El nombre del **owner** del repo en el que está el pull request.                                                                       |
| `atlantis_repo_name=atlantis`        | El nombre del repo en el que está el pull request.                                                                                     |
| `atlantis_pull_num=200`              | El número del pull request.                                                                                                            |

Si quieres usar `assume_role` con Atlantis y también estás usando el [S3 Backend](https://developer.hashicorp.com/terraform/language/settings/backends/s3),
asegúrate de agregar la opción `role_arn`:

```bash
terraform {
  backend "s3" {
    bucket   = "mybucket"
    key      = "path/to/my/key"
    region   = "us-east-1"
    role_arn = "arn:aws:iam::ACCOUNT_ID:role/ROLE_NAME"
    # can't use var.atlantis_user as the session name because
    # interpolations are not allowed in backend configuration
    # session_name = "${var.atlantis_user}" WON'T WORK
  }
}
```

:::tip ¿Por qué esto no funciona en TF >= 0.12?
En Terraform >= 0.12, no se te permite establecer ninguna bandera `-var` si esas variables
no están siendo usadas. Como no podemos saber si estás usando estas variables `atlantis_*`,
no podemos establecer la bandera `-var`.

Todavía puedes establecer estas variables tú mismo usando la configuración `extra_args`.
:::

## Siguientes pasos

* Si quieres configurar Atlantis más a fondo, lee [Configuring Atlantis](configuring-atlantis.md)
* Si estás listo para usar Atlantis, lee [Using Atlantis](using-atlantis.md)
