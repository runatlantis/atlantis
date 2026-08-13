# Terraform Versions

Puede personalizar qué versión de Terraform usa Atlantis por defecto configurando
la flag `--default-tf-version` (p. ej. `--default-tf-version=v1.3.7`).

## Mediante `atlantis.yaml`

Si desea usar una versión diferente de la predeterminada para un repo o proyecto específico, necesita
crear un archivo `atlantis.yaml` y configurar la clave `terraform_version`:

```yaml
version: 3
projects:
- dir: .
  terraform_version: v1.1.5
```

Vea [atlantis.yaml Use Cases](repo-level-atlantis-yaml.md#terraform-versions) para más detalles.

## Mediante terraform config

Alternativamente, se puede usar la clave `required_version` del bloque de configuración de terraform para especificar una versión exacta (`x.y.z` o `= x.y.z`), o a partir de [atlantis v0.21.0](https://github.com/runatlantis/atlantis/releases/tag/v0.21.0), una [version constraint](https://developer.hashicorp.com/terraform/language/expressions/version-constraints#version-constraint-syntax) de comparación o pesimista:

### Exactamente la versión 1.2.9

```tf
terraform {
  required_version = "= 1.2.9"
}
```

### Cualquier versión patch/tiny de la versión minor 1.2 (1.2.z)

```tf
terraform {
  required_version = "~> 1.2.0"
}
```

### Cualquier versión minor de la versión major 1 (1.y.z)

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

Vea [Terraform `required_version`](https://developer.hashicorp.com/terraform/language/terraform#terraform-required_version) como referencia.

::: tip NOTE
Atlantis descargará automáticamente la última versión que cumpla la restricción especificada.
Un `terraform_version` especificado en el archivo `atlantis.yaml` tiene precedencia sobre tanto la flag [`--default-tf-version`](server-configuration.md#default-tf-version) como `required_version` en el hcl de terraform.
Cuando un proyecto configura `terraform_distribution`, Atlantis resuelve la
restricción `required_version` contra esa distribución. Por ejemplo, un proyecto OpenTofu se resuelve a una
versión de OpenTofu en lugar de una versión de Terraform.
:::

## Soporte de archivos `.tofu` de OpenTofu

Cuando la distribución efectiva es OpenTofu, Atlantis lee `required_version`
de los archivos `.tofu` e `.tofu.json` además de `.tf` e `.tf.json`. La distribución
efectiva es OpenTofu cuando se cumple cualquiera de estas condiciones:

- Un proyecto configura `terraform_distribution: opentofu` en `atlantis.yaml`
- El valor predeterminado del servidor es `--default-tf-distribution=opentofu` y el proyecto no lo sobrescribe

Si un proyecto configura explícitamente `terraform_distribution: terraform`, Atlantis usa la
ruta de detección de versión de Terraform (solo `.tf` / `.tf.json`) incluso si el valor predeterminado del servidor es OpenTofu.

Se respeta la precedencia de archivos de OpenTofu: un archivo `.tofu` sobrescribe un archivo `.tf`
con el mismo nombre base, e `.tofu.json` sobrescribe `.tf.json` con el mismo nombre base. Los archivos con diferentes
nombres base contribuyen ambos restricciones de manera independiente.

La distribución Terraform no se ve afectada y continúa leyendo solo los archivos `.tf` / `.tf.json`.

::: warning Known limitation
La indexación de dependencias de module autoplanning (`--autoplan-modules`) aún depende de
`terraform-config-inspect` y no entiende completamente `.tofu` / `.tofu.json`:

1. Los bloques source de módulos definidos solo en archivos `.tofu` / `.tofu.json` no se indexan.
   Los proyectos que usan estos archivos para bloques `module {}` no serán planificados cuando cambien
   módulos compartidos.
2. Es posible que los directorios de módulos compartidos que contienen solo archivos `.tofu` / `.tofu.json` no sean
   reconocidos como módulos por el índice de dependencias.

El autoplanning por cambio directo de archivo (sin `--autoplan-modules`) es totalmente compatible para
proyectos `.tofu`. Como solución alternativa para dependencias de módulos, incluya rutas de módulos compartidos
en patrones explícitos `autoplan.when_modified`, o mantenga las declaraciones source de módulos en
archivos `.tf` hasta que se implemente la indexación completa de módulos `.tofu`.

La detección de workspace de Terraform Cloud (`cloud { workspaces { ... } }`) admite archivos `.tf`,
`.tf.json`, `.tofu` e `.tofu.json` para **autodiscovered projects**. `.tofu` e
`.tofu.json` solo se escanean cuando la distribución predeterminada del servidor
(`--default-tf-distribution`) es OpenTofu. La precedencia con el mismo nombre base se aplica en modo OpenTofu:
`.tofu` sobrescribe `main.tf`, `.tofu.json` sobrescribe `main.tf.json`. Los proyectos con
el valor predeterminado del servidor en Terraform leen `.tf` e `.tf.json` pero ignoran `.tofu` / `.tofu.json`.

Para proyectos configurados explícitamente en `atlantis.yaml`, configure el campo `workspace:` directamente
— el escaneo HCL del workspace no se usa para proyectos configurados independientemente de la distribución.
:::

::: tip NOTE
La [latest docker image](https://github.com/runatlantis/atlantis/pkgs/container/atlantis/9854680?tag=latest) de Atlantis tiende a tener versiones recientes de Terraform, pero puede haber un retraso a medida que se publican nuevas versiones. La versión más alta de Terraform permitida en su código es la versión especificada por `DEFAULT_TERRAFORM_VERSION` en la imagen que está ejecutando su servidor.
:::
