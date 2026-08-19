# Requisitos

Atlantis funciona con la mayoría de los hosts de Git y configuraciones de Terraform. Continúe leyendo para confirmar
que funciona con la suya.

## Host de Git

Atlantis se integra con los siguientes hosts de Git:

* GitHub (público, privado o enterprise)
* GitLab (público, privado o enterprise)
* Gitea (público, privado y forks compatibles como Forgejo)
* Bitbucket Cloud también conocido como bitbucket.org (público o privado)
* Bitbucket Server también conocido como Stash
* Azure DevOps

### Versión de GitLab

Atlantis soporta las versiones de GitLab mantenidas activamente. GitLab 15.6 introdujo
`detailed_merge_status`, que Atlantis usa para verificaciones más precisas del
[requisito de apply fusionable](apply-requirements.md) . En versiones anteriores de GitLab,
Atlantis recurre al campo heredado `merge_status`, que puede
informar una merge request como fusionable incluso cuando la rama necesita un rebase.

## Estado de Terraform

Atlantis soporta todos los tipos de backend **excepto local state**. No soportamos local state
porque Atlantis no tiene almacenamiento permanente y no hace commit del nuevo
statefile de vuelta al control de versiones.

:::tip
Si está buscando una solución sencilla de estado remoto, consulte el almacenamiento de
[free remote state](https://app.terraform.io) de Terraform Cloud. Esto es totalmente soportado por Atlantis.
:::

## Estructura del repositorio

Atlantis soporta cualquier estructura de repositorio de Terraform, por ejemplo:

### Proyecto único de Terraform en la raíz del repo

```plain
.
├── main.tf
└── ...
```

### Múltiples carpetas de proyecto

```plain
.
├── project1
│   ├── main.tf
|   └── ...
└── project2
    ├── main.tf
    └── ...
```

### Módulos

```plain
.
├── project1
│   ├── main.tf
|   └── ...
└── modules
    └── module1
        ├── main.tf
        └── ...
```

Con módulos, si quiere `project1` planificado automáticamente cuando `module1` es modificado
necesita crear un archivo `atlantis.yaml`. Vea [Casos de uso de atlantis.yaml](repo-level-atlantis-yaml.md#configuring-planning) para más detalles.

### Workspaces de Terraform

*Vea [la documentación de Terraform](https://developer.hashicorp.com/terraform/language/state/workspaces) si no está familiarizado con los workspaces.*

Si está usando Terraform `>= 0.9.0`, Atlantis soporta workspaces mediante un
archivo `atlantis.yaml` que le indica a Atlantis los nombres de sus workspaces
(vea [Casos de uso de atlantis.yaml](repo-level-atlantis-yaml.md#supporting-terraform-workspaces) para más detalles)

### Archivos .tfvars

```plain
.
├── production.tfvars
│── staging.tfvars
└── main.tf
```

Atlantis soporta archivos `.tfvars` de dos maneras:

#### Archivos automáticos env/{workspace}.tfvars

Atlantis incluye automáticamente archivos de variables específicos del workspace si existen en un directorio `env/`:

```plain
.
├── main.tf
├── variables.tf
└── env/
    ├── default.tfvars
    ├── staging.tfvars
    └── production.tfvars
```

Al usar esta estructura, Atlantis incluirá automáticamente el archivo apropiado según el workspace:

* `atlantis plan` incluye `env/default.tfvars`
* `atlantis plan -w staging` incluye `env/staging.tfvars`  
* `atlantis plan -w production` incluye `env/production.tfvars`

Esto no requiere configuración adicional y funciona automáticamente.

#### Archivos .tfvars personalizados con atlantis.yaml

Para otras ubicaciones o estructuras de archivos `.tfvars`, necesita crear
un archivo `atlantis.yaml` para indicarle a Atlantis que use `-var-file={YOUR_FILE}`.
Vea [Casos de uso de atlantis.yaml](custom-workflows.md#tfvars-files) para más detalles.

### Múltiples repos

Atlantis también soporta múltiples repos, siempre que haya un webhook configurado
para cada repo.

## Versiones de Terraform

Atlantis soporta todas las versiones de Terraform (incluida 0.12) y puede configurarse
para usar diferentes versiones para diferentes repositorios/proyectos. Vea [Versiones de Terraform](terraform-versions.md).

## Próximos pasos

* Si su configuración de Terraform cumple con los requisitos de Atlantis, continúe con la guía de instalación
  y configure sus [Credenciales de acceso del host de Git](access-credentials.md)
