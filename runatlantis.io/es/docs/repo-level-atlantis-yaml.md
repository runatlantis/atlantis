# Configuración atlantis.yaml a nivel de repo

Un archivo `atlantis.yaml` especificado en la raíz de un repo de Terraform te permite
indicar a Atlantis la estructura de tu repo y establecer workflows personalizados.

## ¿Necesito un archivo atlantis.yaml?

Los archivos `atlantis.yaml` solo son necesarios si deseas personalizar algún aspecto de Atlantis.
La configuración predeterminada de Atlantis funciona para muchos usuarios sin cambios.

Lee los [casos de uso](#use-cases) para determinar si lo necesitas.

## Habilitar atlantis.yaml

De forma predeterminada, a todos los repos se les permite tener un archivo `atlantis.yaml`,
pero algunas de las keys están restringidas de forma predeterminada.

Las keys restringidas pueden establecerse en el archivo de configuración de repo `repos.yaml` del lado del servidor.
Puedes habilitar `atlantis.yaml` para sobrescribir
keys restringidas estableciendo allí la key `allowed_overrides`. Consulta [Server Side Repo Config](server-side-repo-config.md) para
más detalles.

**Notas:**

- De forma predeterminada, se usa el archivo `atlantis.yaml` en la raíz del repo.
- Puedes cambiar este comportamiento configurando [Server Side Repo Config](server-side-repo-config.md)

::: danger DANGER
Atlantis usa la versión de `atlantis.yaml` del pull request, similar a otros
sistemas de CI/CD. Si estás permitiendo a los usuarios [crear workflows personalizados](server-side-repo-config.md#allow-repos-to-define-their-own-workflows)
entonces esto significa que
cualquiera que pueda crear un pull request a tu repo puede ejecutar código arbitrario en el
servidor de Atlantis.

De forma predeterminada, esto no está permitido.
:::

::: warning
Una vez que existe un archivo `atlantis.yaml` en un repo y se configura uno o más `projects`,
Atlantis no intentará determinar automáticamente dónde ejecutar plan. En su lugar, solo
seguirá la configuración del proyecto. Esto significa que necesitarás definir cada proyecto
en tu repo.

Si tienes muchos directorios con configuración de Terraform, cada directorio
necesitará ser definido.

Este comportamiento puede sobrescribirse configurando `autodiscover.mode` en
`enabled`, en cuyo caso Atlantis todavía intentará descubrir proyectos que no fueron
configurados explícitamente. Si el directorio de cualquier proyecto descubierto entra en conflicto con un
proyecto configurado manualmente, el proyecto configurado manualmente tendrá prioridad.
:::

## Ejemplo usando todas las keys

```yaml
version: 3 # Available since v0.1.0
automerge: true # Available since v0.15.0
autodiscover: # Available since v0.18.0
  mode: auto
  ignore_paths:
  - some/path
delete_source_branch_on_merge: true # Available since v0.15.0
parallel_plan: true # Available since v0.17.0
parallel_apply: true # Available since v0.17.0
abort_on_execution_order_fail: true # Available since v0.17.0
projects:
- name: my-project-name # Available since v0.1.0
  branch: /main/ # Available since v0.21.0
  dir: . # Available since v0.1.0
  workspace: default # Available since v0.1.0
  terraform_distribution: terraform # Available since v0.33.0
  terraform_version: v0.11.0 # Available since v0.1.0
  delete_source_branch_on_merge: true # Available since v0.17.0
  repo_locking: true # deprecated: use repo_locks instead, Available since v0.17.0
  repo_locks: # Available since v0.17.0
    mode: on_plan
  custom_policy_check: false # Available since v0.17.0
  autoplan: # Available since v0.1.0
    when_modified: ["*.tf", "../modules/**/*.tf", ".terraform.lock.hcl"]
    enabled: true
  plan_requirements: [mergeable, approved, undiverged] # Available since v0.17.0
  apply_requirements: [mergeable, approved, undiverged] # Available since v0.17.0
  import_requirements: [mergeable, approved, undiverged] # Available since v0.17.0
  silence_pr_comments: ["apply"] # Available since v0.17.0
  execution_order_group: 1 # Available since v0.17.0
  depends_on: # Available since v0.20.0
    - project-1
  workflow: myworkflow # Available since v0.17.0
workflows: # Available since v0.1.0
  myworkflow:
    plan:
      steps:
      - run: my-custom-command arg1 arg2
      - run:
          command: my-custom-command arg1 arg2
          output: hide
      - init
      - plan:
          extra_args: ["-lock", "false"]
      - run: my-custom-command arg1 arg2
    apply:
      steps:
      - run: echo hi
      - apply
allowed_regexp_prefixes: # Available since v0.19.0
- dev/
- staging/
```

## Ejemplo de DRY de proyectos usando YAML anchors

```yaml
projects:
   - &template
     name: template
     dir: template
     workflow: custom
     autoplan:
        enabled: true
        when_modified:
           - "./terraform/modules/**/*.tf"
           - "**/*.tf"
           - ".terraform.lock.hcl"

   - <<: *template
     name: ue1-prod-titan
     dir: ./terraform/titan
     workspace: ue1-prod

   - <<: *template
     name: ue1-stage-titan
     dir: ./terraform/titan
     workspace: ue1-stage

   - <<: *template
     name: ue1-dev-titan
     dir: ./terraform/titan
     workspace: ue1-dev
```

## Generar proyectos automáticamente

Esto es útil si tienes muchos proyectos en un repositorio. Esto asume el workspace `default` (o ningún workspace).

Ejecuta esto en la raíz de tu repositorio. Esto usará gnu `grep` para buscar archivos terraform para un backend S3 (terraform dir), recuperar la ruta del directorio, recuperar las entradas únicas, y luego usar `yq` para devolver el YAML de una configuración simple de directorio de proyecto que luego puede modificarse a tu gusto.

```sh
grep -P 'backend[\s]+"s3"' **/*.tf |
  rev | cut -d'/' -f2- | rev |
  sort |
  uniq |
  while read d; do \
    echo '[ {"name": "'"$d"'","dir": "'"$d"'", "autoplan": {"when_modified": ["**/*.tf.*"] }} ]' | yq -PM; \
  done
```

## Casos de uso

### Deshabilitar Autoplanning

```yaml
version: 3
projects:
   - dir: project1
     autoplan:
        enabled: false
```

Esto hará que Atlantis deje de ejecutar automáticamente plan cuando `project1/` se actualice
en un pull request.

### Ejecutar plans y applies en paralelo

```yaml
version: 3
parallel_plan: true
parallel_apply: true
```

Esto ejecutará plans y applies para todos tus proyectos en paralelo.

Habilitar estas opciones puede reducir significativamente la duración de los plans y applies, especialmente para repositorios con muchos proyectos.

Usa `--parallel-pool-size` para configurar el número máximo de plans y applies que pueden ejecutarse en paralelo. El valor predeterminado es 15.

Los plans y applies en paralelo funcionan tanto en múltiples directorios como en múltiples workspaces.

### Configurar Planning

Dada la estructura de directorios:

```plain
.
├── modules
│   └── module1
│       ├── main.tf
│       ├── outputs.tf
│       └── submodule
│           ├── main.tf
│           └── outputs.tf
└── project1
    └── main.tf
```

Si quieres que Atlantis haga plan de `project1/` cada vez que cualquier archivo `.tf` bajo `module1/` cambie o cualquier archivo `.tf` o `.tfvars` bajo `project1/` cambie, podrías usar la siguiente configuración:

```yaml
version: 3
projects:
   - dir: project1
     autoplan:
        when_modified: ["../modules/**/*.tf", "*.tf*", ".terraform.lock.hcl"]
```

Nota:

- `when_modified` usa la [sintaxis de `.dockerignore`](https://docs.docker.com/engine/reference/builder/#dockerignore-file)
- Las rutas son relativas al directorio del proyecto.
- `when_modified` será usado tanto por plans automáticos como por los ejecutados manualmente.
- `when_modified` seguirá funcionando para plans ejecutados manualmente incluso cuando autoplan esté deshabilitado.
- El valor predeterminado de `when_modified` incluye `**/*.tf*`, `**/*.tofu`, `**/*.tofu.json`, `**/terragrunt.hcl` y `**/.terraform.lock.hcl`. Los valores personalizados de `when_modified` sobrescriben por completo estos valores predeterminados.

### Soporte para Terraform Workspaces

```yaml
version: 3
projects:
   - dir: project1
     workspace: staging
   - dir: project1
     workspace: production
```

Con la configuración anterior, cuando Atlantis determina que la configuración del directorio `project1` ha cambiado,
ejecutará plan para los workspaces `staging` e `production`.

Si quieres `plan` o `apply` para un workspace específico puedes usar

```shell
atlantis plan -w staging -d project1
```

y

```shell
atlantis apply -w staging -d project1
```

### Usar archivos .tfvars

Consulta [Custom Workflow Use Cases: Using .tfvars files](custom-workflows.md#tfvars-files)

### Agregar argumentos extra a comandos de Terraform

Consulta [Custom Workflow Use Cases: Adding extra arguments to Terraform commands](custom-workflows.md#adding-extra-arguments-to-terraform-commands)

### Comandos personalizados de init/plan/apply

Consulta [Custom Workflow Use Cases: Custom init/plan/apply Commands](custom-workflows.md#custom-init-plan-apply-commands)

### Terragrunt

Consulta [Custom Workflow Use Cases: Terragrunt](custom-workflows.md#terragrunt)

### Ejecutar comandos personalizados

Consulta [Custom Workflow Use Cases: Running custom commands](custom-workflows.md#running-custom-commands)

### Distribuciones de Terraform

Si quieres usar una distribución diferente de Terraform a la que está configurada
por la flag `--default-tf-version`, entonces establece la key `terraform_distribution`:

```yaml
version: 3
projects:
   - dir: project1
     terraform_distribution: opentofu
```

Atlantis descargará y usará automáticamente esta distribución. Los valores válidos son `terraform` e `opentofu`.
Si `terraform_version` se omite y el proyecto usa una restricción `required_version`, Atlantis resuelve esa
restricción contra la distribución seleccionada.

### Versiones de Terraform

Si quieres usar una versión diferente de Terraform a la que está en `PATH` de Atlantis
o está configurada por la flag `--default-tf-version`, entonces establece la key `terraform_version`:

```yaml
version: 3
projects:
   - dir: project1
     terraform_version: 0.10.0
```

Atlantis descargará y usará automáticamente esta versión.

### Requerir aprobaciones para producción

En este ejemplo, solo queremos requerir aprobaciones `apply` para el directorio `production`.

```yaml
version: 3
projects:
   - dir: staging
   - dir: production
     plan_requirements: [approved]
     apply_requirements: [approved]
     import_requirements: [approved]
```

:::warning
`plan_requirements`, `apply_requirements` e `import_requirements` son keys restringidas, por lo que este repo necesitará estar configurado
para que se le permita establecer esta key. Consulta [Server-Side Repo Config Use Cases](server-side-repo-config.md#repos-can-set-their-own-apply-requirements).
:::

### Orden de planning/applying

```yaml
version: 3
abort_on_execution_order_fail: true
projects:
   - dir: project1
     execution_order_group: 2
   - dir: project2
     execution_order_group: 1
```

Con esta configuración anterior, Atlantis ejecuta planning/applying para project2 primero, luego para project1.
Varios proyectos pueden tener el mismo `execution_order_group`. No se garantiza ningún orden dentro de un grupo.
`parallel_plan` e `parallel_apply` respetan estos grupos de orden, por lo que el planning/applying en paralelo funciona
en cada grupo uno por uno.

Si cualquier plan/apply falla y `abort_on_execution_order_fail` está establecido en true a nivel de repo, todos los
grupos siguientes serán abortados. Para este ejemplo, si project2 falla entonces project1 no se ejecutará.

Los grupos de orden de ejecución son útiles cuando tienes dependencias entre proyectos. Sin embargo, solo son aplicables en el caso en que
inicies un apply global para todos tus proyectos, es decir `atlantis apply`. Si inicias un apply en un solo proyecto, entonces los grupos de orden de ejecución se ignoran.
Por lo tanto, la key `depends_on` es más útil en este caso. y puede usarse junto con los grupos de orden de ejecución.

La siguiente configuración es un ejemplo de cómo usar juntos grupos de orden de ejecución y depends_on para imponer dependencias entre proyectos.

```yaml
version: 3
projects:
   - name: development
     dir: .
     autoplan:
        when_modified: ["*.tf", "vars/development.tfvars"]
     execution_order_group: 1
     workspace: development
     workflow: infra
   - name: staging
     dir: .
     autoplan:
        when_modified: ["*.tf", "vars/staging.tfvars"]
     depends_on: ["development"]
     execution_order_group: 2
     workspace: staging
     workflow: infra
   - name: production
     dir: .
     autoplan:
        when_modified: ["*.tf", "vars/production.tfvars"]
     depends_on: ["staging"]
     execution_order_group: 3
     workspace: production
     workflow: infra
```

la funcionalidad `depends_on` se asegurará de que `production` no se aplique antes de `staging`, por ejemplo.

::: tip
¿Qué sucede si una o más dependencias de un proyecto no están aplicadas?

Si hay uno o más proyectos en la lista de dependencias que no están en estado applied, los usuarios verán un mensaje de error como este:
`Can't apply your project unless you apply its dependencies`
:::

### Configuración de autodiscovery

```yaml
autodiscover:
   mode: "auto"
```

Lo anterior es la configuración predeterminada para `autodiscover.mode`. Cuando `autodiscover.mode` es auto,
los proyectos se descubrirán solo si el repo no tiene ningún `projects` configurado.

```yaml
autodiscover:
   mode: "disabled"
```

Con la configuración anterior, Atlantis nunca intentará descubrir proyectos, incluso cuando no haya
`projects` configurados. Esto es útil si generas dinámicamente la configuración de Atlantis en hooks pre_workflow.
Consulta [Dynamic Repo Config Generation](pre-workflow-hooks.md#dynamic-repo-config-generation).

```yaml
autodiscover:
   mode: "enabled"
```

Con la configuración anterior, Atlantis intentará incondicionalmente descubrir proyectos basándose en modified_files,
incluso cuando el directorio del proyecto falta en los `projects` configurados en la configuración del repo.
Si un proyecto descubierto tiene el mismo directorio que un proyecto que fue configurado manualmente en `projects`,
la configuración manual tendrá prioridad.

Usa esta funcionalidad cuando algunos proyectos requieren configuración específica en un repo con muchos proyectos y aun así
sigue siendo deseable que Atlantis haga plan/apply para proyectos no enumerados en la configuración.

Esta configuración se ignora si está configurada en el servidor, consulta [Server Side Repo Config](server-side-repo-config.md#repo)

```yaml
autodiscover:
   mode: "enabled"
   ignore_paths:
      - dir/*
```

Autodiscover también puede configurarse para omitir directorios que coincidan con un path glob (como se define en el [paquete de coincidencia de paths doublestar](https://pkg.go.dev/github.com/bmatcuk/doublestar/v4)).

Cuando `ignore_paths` está configurado, se aplica a:

- Descubrimiento automático de proyectos durante autoplan e `atlantis plan` (sin `-d`)
- `atlantis apply` (sin `-d`) al filtrar plans pendientes
- Todos los comandos `-d` dirigidos (`plan`, `apply`, `import`, `state rm`, etc.) cuando autodiscovery está habilitado, si el path no tiene configuración explícita de proyecto

Esto hace que `ignore_paths` sea útil para **configuraciones de múltiples instancias** donde cada instancia de Atlantis administra un subárbol de directorios diferente. Por ejemplo, una instancia puede ignorar `environments/prod/**` mientras otra ignora `environments/nonprod/**`, evitando interferencia entre instancias en comandos dirigidos.

### Configuración personalizada de backend

Consulta [Custom Workflow Use Cases: Custom Backend Config](custom-workflows.md#custom-backend-config)

## Referencia

### Keys de nivel superior

```yaml
version: 3
automerge: false
delete_source_branch_on_merge: false
projects:
workflows:
allowed_regexp_prefixes:
```

| Key                           | Type                                                   | Default | Required | Description                                                                                                                        |
| ----------------------------- | ------------------------------------------------------ | ------- | -------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| version                       | int                                                    | none    | **yes**  | Esta key es requerida y debe establecerse en `3`.                                                                                       |
| automerge                     | bool                                                   | `false` | no       | Fusiona automáticamente el pull request cuando todos los plans están aplicados.                                                                      |
| delete_source_branch_on_merge | bool                                                   | `false` | no       | Elimina automáticamente la rama de origen al hacer merge.                                                                                  |
| projects                      | array[[Project](repo-level-atlantis-yaml.md#project)]  | `[]`    | no       | Lista los proyectos en este repo.                                                                                                   |
| workflows<br />_(restricted)_ | map[string: [Workflow](custom-workflows.md#reference)] | `{}`    | no       | Workflows personalizados.                                                                                                                  |
| allowed_regexp_prefixes       | array\[string\]                                        | `[]`    | no       | Lista los prefijos regexp permitidos para usar cuando se usa la flag [`--enable-regexp-cmd`](server-configuration.md#enable-regexp-cmd). |

### Project

```yaml
name: myname
branch: /mybranch/
dir: mydir
workspace: myworkspace
execution_order_group: 0
delete_source_branch_on_merge: false
repo_locking: true # deprecated: use repo_locks instead
repo_locks:
   mode: on_plan
custom_policy_check: false
autoplan:
terraform_version: 0.11.0
plan_requirements: ["approved"]
apply_requirements: ["approved"]
import_requirements: ["approved"]
silence_pr_comments: ["apply"]
workflow: myworkflow
```

| Key                                     | Type                    | Default         | Required | Description                                                                                                                                                                                                                             |
| --------------------------------------- | ----------------------- | --------------- | -------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| name                                    | string                  | none            | maybe    | Requerida si hay más de un proyecto con el mismo `dir` e `workspace`. Este nombre de proyecto puede usarse con la flag `-p`.                                                                                                       |
| branch                                  | string                  | none            | no       | Regex que hace match con proyectos por la rama base del pull request (la rama en la que se va a fusionar el pull request). Solo se considerarán los proyectos que coincidan con la rama del PR. De forma predeterminada, coinciden todas las ramas.                     |
| dir                                     | string                  | none            | **yes**  | El directorio de este proyecto relativo a la raíz del repo. Por ejemplo, si el proyecto estaba bajo `./project1` entonces usa `project1`. Usa `.` para indicar la raíz del repo.                                                                      |
| workspace                               | string                  | `"default"`     | no       | El [Terraform workspace](https://developer.hashicorp.com/terraform/language/state/workspaces) para este proyecto. Atlantis cambiará a este workspace al hacer planning/applying y lo creará si no existe.                  |
| execution_order_group                   | int                     | `0`             | no       | Índice del grupo de orden de ejecución. Los proyectos se ordenarán por este campo antes de planning/applying.                                                                                                                                         |
| delete_source_branch_on_merge           | bool                    | `false`         | no       | Elimina automáticamente la rama de origen al hacer merge.                                                                                                                                                                                       |
| repo_locking                            | bool                    | `true`          | no       | (deprecated) Obtiene un bloqueo de repositorio en este proyecto al hacer plan.                                                                                                                                                                           |
| repo_locks                              | [RepoLocks](#repolocks) | `mode: on_plan` | no       | Obtiene un bloqueo de repositorio en este proyecto en plan o apply. Consulta [RepoLocks](#repolocks) para más detalles.                                                                                                                                   |
| custom_policy_check                     | bool                    | `false`         | no       | Habilita el uso de herramientas de policy check distintas de Conftest                                                                                                                                                                                     |
| autoplan                                | [Autoplan](#autoplan)   | none            | no       | Una configuración personalizada de autoplan. Si no se especifica, usará la configuración de autoplan. Consulta [Autoplanning](autoplanning.md).                                                                                                                   |
| terraform_version                       | string                  | none            | no       | Una versión específica de Terraform para usar al ejecutar comandos para este proyecto. Debe ser [compatible con Semver](https://semver.org/), ej. `v0.11.0`, `0.12.0-beta1`.                                                                            |
| plan_requirements<br />_(restricted)_   | array\[string\]         | none            | no       | Requisitos que deben satisfacerse antes de que pueda ejecutarse `atlantis plan`. Actualmente, los únicos requisitos admitidos son `approved`, `mergeable` e `undiverged`. Consulta [Command Requirements](command-requirements.md) para más detalles.   |
| apply_requirements<br />_(restricted)_  | array\[string\]         | none            | no       | Requisitos que deben satisfacerse antes de que pueda ejecutarse `atlantis apply`. Actualmente, los únicos requisitos admitidos son `approved`, `mergeable` e `undiverged`. Consulta [Command Requirements](command-requirements.md) para más detalles.  |
| import_requirements<br />_(restricted)_ | array\[string\]         | none            | no       | Requisitos que deben satisfacerse antes de que pueda ejecutarse `atlantis import`. Actualmente, los únicos requisitos admitidos son `approved`, `mergeable` e `undiverged`. Consulta [Command Requirements](command-requirements.md) para más detalles. |
| silence_pr_comments                     | array\[string\]         | none            | no       | Silencia los comentarios del PR de las etapas definidas mientras preserva las verificaciones de estado del PR. Los valores admitidos son: `plan`, `apply`.                                                                                                                       |
| workflow <br />_(restricted)_           | string                  | none            | no       | Un workflow personalizado. Si no se especifica, Atlantis usará su workflow predeterminado.                                                                                                                                                            |

::: tip
Un proyecto representa un estado de Terraform. Típicamente, hay un estado por directorio y workspace; sin embargo, es posible
tener múltiples estados en el mismo directorio usando `terraform init -backend-config=custom-config.tfvars`.
Atlantis soporta esto, pero requiere que se especifique la key `name`. Consulta [Custom Backend Config](custom-workflows.md#custom-backend-config) para más detalles.
:::

### Autoplan

```yaml
enabled: true
when_modified: ["*.tf", "terragrunt.hcl", ".terraform.lock.hcl"]
```

| Key           | Type            | Default        | Required | Description                                                                                                                                                                                                                                                     |
| ------------- | --------------- | -------------- | -------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| enabled       | boolean         | `true`         | no       | Si autoplanning está habilitado para este proyecto.                                                                                                                                                                                                               |
| when_modified | array\[string\] | see below      | no       | Usa sintaxis de [.dockerignore](https://docs.docker.com/engine/reference/builder/#dockerignore-file). Si cualquier archivo modificado en el pull request coincide, se hará plan de este proyecto. Consulta [Autoplanning](autoplanning.md). Las rutas son relativas al dir del proyecto. |

Patrones predeterminados de `when_modified`: `["**/*.tf*", "**/*.tofu", "**/*.tofu.json", "**/terragrunt.hcl", "**/.terraform.lock.hcl"]`. Los valores personalizados de `when_modified` sobrescriben por completo estos valores predeterminados. El valor predeterminado es global (no depende de la distribución).

### RepoLocks

```yaml
mode: on_apply
```

| Key  | Type   | Default   | Required | Description                                                                                                                           |
| ---- | ------ | --------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------- |
| mode | `Mode` | `on_plan` | no       | Si los bloqueos de repositorio están habilitados o no para este proyecto en plan o apply. Los valores válidos son `disabled`, `on_plan` e `on_apply`. |
