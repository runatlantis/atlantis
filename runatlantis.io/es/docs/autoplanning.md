# Planificación automática

En cualquier pull request **nuevo** o **nuevo commit** en un pull request existente, Atlantis intentará
executar `terraform plan` en los directorios que crea que contienen proyectos de Terraform modificados.

El algoritmo que usa es el siguiente:

1. Obtener la lista de todos los archivos modificados en el pull request
1. Filtrar a aquellos que coinciden con el [`--autoplan-file-list`](server-configuration.md#autoplan-file-list) configurado (por defecto incluye `.tf`, `.tf.json`, `.tfvars`, `.tofu`, `.tofu.json`, etc.)
1. Obtener los directorios en los que están esos archivos
1. Si la ruta del directorio no contiene `modules/`, entonces intentar ejecutar `plan` en ese directorio
1. Si sí contiene `modules/`, mirar el directorio un nivel por encima de `modules/`. Si
contiene un archivo indicador de proyecto compatible (`*.tf`, `*.tf.json`, `*.tofu`, `*.tofu.json`, o `terragrunt.hcl`) ejecutar plan en ese directorio; de lo contrario, ignorar el cambio (vea abajo las excepciones).

## Ejemplo

Dada la estructura de directorios:

```plain
.
├── modules
│   └── module1
│       └── main.tf
└── project1
    ├── main.tf
    └── modules
        └── module1
            └── main.tf
```

* Si se modificara `project1/main.tf`, ejecutaríamos `plan` en `project1`
* Si se modificara `modules/module1/main.tf`, no ejecutaríamos automáticamente `plan` porque no podríamos determinar la ubicación del proyecto de Terraform
  * Podría usar un archivo [atlantis.yaml](repo-level-atlantis-yaml.md#configuring-planning) para especificar qué proyectos hacer plan cuando este módulo cambie
  * Podría habilitar la [planificación automática de módulos](server-configuration.md#autoplan-modules), que indexa los proyectos según sus dependencias de módulos locales.
  * O podría hacer plan manualmente con `atlantis plan -d <dir>`
* Si se modificara `project1/modules/module1/main.tf`, miraríamos un nivel por encima de `project1/modules`,
en `project1/`, veríamos que había un archivo `main.tf` y por lo tanto ejecutaríamos plan en `project1/`

## Notas específicas de Bitbucket

Bitbucket no tiene un webhook que se active solo cuando hay un PR nuevo o un commit nuevo. Para solucionar esto, almacenamos en caché el último commit para ver si ha cambiado. Si la caché se vacía, Atlantis pensará que su commit es nuevo y puede que vea plans adicionales.
Este escenario puede ocurrir si:

* Atlantis se reinicia
* Está ejecutando múltiples instancias de Atlantis detrás de un balanceador de carga

## Personalización

Si desea personalizar cómo Atlantis determina en qué directorio ejecutar
o deshabilitarlo por completo, necesita crear un archivo `atlantis.yaml`.
Vea

* [Deshabilitar la planificación automática](repo-level-atlantis-yaml.md#disabling-autoplanning)
* [Configurar Planning](repo-level-atlantis-yaml.md#configuring-planning)
