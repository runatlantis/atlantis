# Terraform Cloud/Enterprise

::: tip NOTE
Terraform Enterprise fue [renombrado recientemente](https://www.hashicorp.com/blog/introducing-terraform-cloud-remote-state-management) a Terraform Cloud
y Private Terraform Enterprise fue renombrado a Terraform Enterprise.
:::

Atlantis se integra perfectamente con Terraform Cloud y Terraform Enterprise, ya sea que estés usando:

* [Free Remote State Management](https://app.terraform.io)
* Terraform Cloud Paid Tiers
* Una instalación privada de Terraform Enterprise

Lee la documentación a continuación :point_down: según tu caso de uso.

## Usando Atlantis con Free Remote State Storage

Para usar Atlantis con Free Remote State Storage, necesitas:

1. Migrar tu state a Terraform Cloud. Consulta [Migrating State from Local Terraform](https://developer.hashicorp.com/terraform/cloud-docs/migrate)
1. Actualizar cualquier proyecto que esté referenciando el state que migraste para usar la nueva ubicación
1. [Generar un token de Terraform Cloud/Enterprise](#generating-a-terraform-cloud-enterprise-token)
1. [Pasar el token a Atlantis](#passing-the-token-to-atlantis)

¡Eso es todo! Atlantis se ejecutará con normalidad y tu state se almacenará en Terraform
Cloud.

## Usando Atlantis con Terraform Cloud Remote Operations o Terraform Enterprise

Atlantis se integra con la versión completa de Terraform Cloud y Terraform Enterprise
a través del [remote backend](https://developer.hashicorp.com/terraform/language/settings/backends/remote).

Atlantis ejecutará `terraform` comandos como de costumbre, sin embargo esos comandos
realmente se ejecutarán *remotamente* en Terraform Cloud o Terraform Enterprise.

### ¿Por qué?

Usar Atlantis con Terraform Cloud o Terraform Enterprise te da acceso a características como:

* Salida en streaming en tiempo real
* Capacidad de cancelar comandos en progreso
* Variables secretas
* [Sentinel](https://www.hashicorp.com/sentinel)

**Sin** tener que cambiar tu workflow de pull request.

### Primeros pasos

Para usar Atlantis con Terraform Cloud Remote Operations o Terraform Enterprise, necesitas:

1. Migrar tu state a Terraform Cloud/Enterprise. Consulta [Migrating State from Local Terraform](https://developer.hashicorp.com/terraform/cloud-docs/migrate)
1. Actualizar cualquier proyecto que esté referenciando el state que migraste para usar la nueva ubicación
1. [Generar un token de Terraform Cloud/Enterprise](#generating-a-terraform-cloud-enterprise-token)
1. [Pasar el token a Atlantis](#passing-the-token-to-atlantis)

## Generando un token de Terraform Cloud/Enterprise

Atlantis necesita un token de Terraform Cloud/Enterprise que usará para acceder a la API.
Se recomienda **usar un Team Token**, sin embargo también puedes usar un User Token.

### Team Token

Para generar un team token, haz clic en **Settings** en la barra superior, luego en **Teams** en
la barra lateral.
Elige un team existente o crea uno nuevo.
Habilita el permiso **Manage Workspaces**, luego desplázate hacia abajo hasta **Team API Token**.

### User Token

Para generar un user token, haz clic en tu avatar, luego en **User Settings**, después en
**Tokens** en la barra lateral.
Asegúrate de que el permiso **Manage Workspaces** esté habilitado para el team de este usuario.

## Pasando el token a Atlantis

El token puede pasarse a Atlantis mediante la variable de entorno `ATLANTIS_TFE_TOKEN`.

También puedes usar la flag `--tfe-token`, sin embargo entonces tu token sería fácilmente
visible en la lista de procesos.

Si estás alojando tu propia instalación de Terraform Enterprise, establece la
flag `--tfe-hostname` con su hostname.

¡Eso es todo! Atlantis ahora debería poder realizar operaciones de Terraform usando el
remote state backend de Terraform Cloud/Enterprise.

:::warning
Si estás usando el modo de ejecución local para tus workspaces, recuerda establecer
`--tfe-local-execution-mode`. De lo contrario, no verás los logs en Atlantis.

:::warning
La integración de Terraform Cloud/Enterprise solo funciona con los pasos incorporados
`plan` e `apply`. No funciona con pasos `run` personalizados que reemplacen
plan o apply.
:::

:::tip NOTE
Internamente, Atlantis está generando un archivo `~/.terraformrc`.
Si ya tenías un archivo `~/.terraformrc` donde Atlantis se está ejecutando,
 entonces necesitarás
agregar manualmente el bloque de credenciales a ese archivo:

```hcl
...
credentials "app.terraform.io" {
  token = "xxxx"
}
```

en lugar de usar la variable de entorno `ATLANTIS_TFE_TOKEN`, ya que Atlantis
no sobrescribirá tu archivo `.terraformrc`.
:::
