# Versiones de Terraform

Puede personalizar qué versión de Terraform usa Atlantis por defecto configurando
la bandera `--default-tf-version` (p. ej. `--default-tf-version=v1.3.7`).

## Mediante `atlantis.yaml`

Si desea usar una versión diferente de la predeterminada para un repo o proyecto específico, necesita
crear un archivo `atlantis.yaml` y establecer la clave `terraform_version`:

```yaml
version: 3
projects:
- dir: .
  terraform_version: v1.1.5
```

Consulte [casos de uso de atlantis.yaml](repo-level-atlantis-yaml.md#terraform-versions) para más detalles.

## Mediante terraform config

Alternativamente, se puede usar la clave `required_version` del bloque de configuración de terraform para especificar una versión exacta (`x.y.z` o `= x.y.z`), o a partir de [atlantis v0.21.0](https://github.com/runatlantis/atlantis/releases/tag/v0.21.0), una [restricción de versión](https://developer.hashicorp.com/terraform/language/expressions/version-constraints#version-constraint-syntax) comparativa o pesimista:

### Exactamente la versión 1.2.9

```tf
terraform {
  required_version = "= 1.2.9"
}
```

### Cualquier versión patch/tiny de la versión menor 1.2 (1.2.z)

```tf
terraform {
  required_version = "~> 1.2.0"
}
```

### Cualquier versión menor de la versión mayor 1 (1.y.z)

```tf
terraform {
  required_version = "~> 1.2"
}
```

### Cualquier versión que sea al menos 1.2.0

```tf
terraform {
  required_version = ">= 1.2.0"
}
```

Consulte [Terraform `required_version`](https://developer.hashicorp.com/terraform/language/terraform#terraform-required_version) como referencia.

::: tip NOTE
Atlantis descargará automáticamente la versión más reciente que cumpla la restricción especificada.
Un `terraform_version` especificado en el archivo `atlantis.yaml` tiene precedencia sobre tanto la bandera [`--default-tf-version`](server-configuration.md#default-tf-version) como `required_version` en el terraform hcl.
Cuando un proyecto establece `terraform_distribution`, Atlantis resuelve la restricción `required_version`
contra esa distribución. Por ejemplo, un proyecto OpenTofu se resuelve a una
versión de OpenTofu en lugar de una versión de Terraform.
:::

## Soporte de archivos `.tofu` de OpenTofu

Cuando la distribución efectiva es OpenTofu, Atlantis lee `required_version`
de los archivos `.tofu` e `.tofu.json` además de `.tf` e `.tf.json`. La distribución
efectiva es OpenTofu cuando se cumple cualquiera de estas condiciones:

- Un proyecto establece `terraform_distribution: opentofu` en `atlantis.yaml`
- El valor predeterminado del servidor es `--default-tf-distribution=opentofu` y el proyecto no lo sobrescribe

Si un proyecto establece explícitamente `terraform_distribution: terraform`, Atlantis usa la
ruta de detección de versiones de Terraform (solo `.tf` / `.tf.json`) incluso si el valor predeterminado del servidor es OpenTofu.

Se respeta la precedencia de archivos de OpenTofu: un archivo `.tofu` sobrescribe un archivo `.tf`
con el mismo nombre base, e `.tofu.json` sobrescribe `.tf.json` con el mismo nombre base. Los archivos con diferentes
nombres base contribuyen ambos restricciones de forma independiente.

La distribución de Terraform no se ve afectada y sigue leyendo solo los archivos `.tf` / `.tf.json`.

::: warning Limitación conocida
La indexación de dependencias para autoplanning de módulos (`--autoplan-modules`) todavía depende de
`terraform-config-inspect` y no entiende completamente `.tofu` / `.tofu.json`:

1. Los bloques source de módulo definidos solo en archivos `.tofu` / `.tofu.json` no se indexan.
   Los proyectos que usan estos archivos para bloques `module {}` no se planificarán cuando cambien
   los módulos compartidos.
2. Los directorios de módulos compartidos que contienen solo archivos `.tofu` / `.tofu.json` pueden no ser
   reconocidos como módulos por el índice de dependencias.

El autoplanning por cambio directo de archivos (sin `--autoplan-modules`) es totalmente compatible para
proyectos `.tofu`. Como solución temporal para dependencias de módulos, incluya rutas de módulos compartidos
en patrones `autoplan.when_modified` explícitos, o mantenga las declaraciones source de módulos en
archivos `.tf` hasta que se implemente la indexación completa de módulos `.tofu`.

La detección de workspace de Terraform Cloud (`cloud { workspaces { ... } }`) admite archivos `.tf`,
`.tf.json`, `.tofu` e `.tofu.json` para **proyectos autodiscovered**. `.tofu` e
`.tofu.json` solo se escanean cuando la distribución predeterminada del servidor
(`--default-tf-distribution`) es OpenTofu. La precedencia de mismo nombre base se aplica en modo OpenTofu:
`.tofu` sobrescribe `main.tf`, `.tofu.json` sobrescribe `main.tf.json`. Los proyectos con
valor predeterminado del servidor Terraform leen `.tf` e `.tf.json` pero ignoran `.tofu` / `.tofu.json`.

Para proyectos configurados explícitamente en `atlantis.yaml`, establezca el campo `workspace:` directamente
— el escaneo HCL del workspace no se usa para proyectos configurados independientemente de la distribución.
:::

::: tip NOTE
La [imagen docker latest](https://github.com/runatlantis/atlantis/pkgs/container/atlantis/9854680?tag=latest) de Atlantis tiende a tener versiones recientes de Terraform, pero puede haber una demora a medida que se publican nuevas versiones. La versión más alta de Terraform permitida en su código es la versión especificada por `DEFAULT_TERRAFORM_VERSION` en la imagen que está ejecutando su servidor.
:::
