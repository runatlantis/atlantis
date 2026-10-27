# Config de atlantis.yaml a nivel de repo

Un archivo `atlantis.yaml` especificado en la raíz de un repo de Terraform te permite
indicar a Atlantis la estructura de tu repo y establecer workflows personalizados.

## ¿Necesito un archivo atlantis.yaml?

Los archivos `atlantis.yaml` solo se requieren si deseas personalizar algún aspecto de Atlantis.
La config predeterminada de Atlantis funciona para muchos usuarios sin cambios.

Lee los [casos de uso](#use-cases) para determinar si lo necesitas.

## Habilitar atlantis.yaml

De forma predeterminada, todos los repos pueden tener un archivo `atlantis.yaml`,
pero algunas de las keys están restringidas de forma predeterminada.

Las keys restringidas se pueden establecer en el archivo de config del repo `repos.yaml` del lado del servidor.
Puedes habilitar `atlantis.yaml` para sobrescribir
keys restringidas estableciendo allí la key `allowed_overrides`. Consulta [Server Side Repo Config](server-side-repo-config.md) para
más detalles.

**Notas:**

- De forma predeterminada, se usa el archivo `atlantis.yaml` en la raíz del repo.
- Puedes cambiar este comportamiento configurando [Server Side Repo Config](server-side-repo-config.md)

::: danger DANGER
Atlantis usa la versión `atlantis.yaml` del pull request, similar a otros
sistemas CI/CD. Si estás permitiendo a los usuarios [crear workflows personalizados](server-side-repo-config.md#allow-repos-to-define-their-own-workflows)
entonces esto significa que
cualquiera que pueda crear un pull request a tu repo puede ejecutar código arbitrario en el
servidor de Atlantis.

De forma predeterminada, esto no está permitido.
:::

::: warning
Una vez que existe un archivo `atlantis.yaml` en un repo y uno o más `projects` están configurados,
Atlantis no intentará determinar automáticamente dónde ejecutar plan. En su lugar, solo
seguirá la configuración del proyecto. Esto significa que necesitarás definir cada proyecto
en tu repo.

Si tienes muchos directorios con configuración de Terraform, cada directorio
tendrá que ser definido.

Este comportamiento puede sobrescribirse estableciendo `autodiscover.mode` en
`enabled`, en cuyo caso Atlantis seguirá intentando descubrir proyectos que no fueron
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

## Ejemplo de hacer DRY de proyectos usando YAML anchors

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

Ejecuta esto en la raíz de tu repositorio. Esto usará gnu `grep` para buscar archivos terraform por un backend S3 (directorio terraform), recuperar la ruta del directorio, recuperar las entradas únicas, y luego usar `yq` para devolver el YAML de una configuración simple de directorio de proyecto que luego puede modificarse según tus preferencias.

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

### Deshabilitar autoplanning

```yaml
version: 3
projects:
   - dir: project1
     autoplan:
        enabled: false
```

Esto hará que Atlantis deje de ejecutar automáticamente plan cuando `project1/` se actualiza
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

Los plans y applies en paralelo funcionan tanto entre múltiples directorios como entre múltiples workspaces.

`parallel_plan` también se aplica a los plans y policy checks iniciados a través de la [API](api-endpoints.md), como `POST /api/plan` e `POST /api/drift/detect`. Los applies iniciados a través de la API todavía se ejecutan un proyecto a la vez.

### Configuración de Planning

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

Si quieres que Atlantis haga plan de `project1/` cada vez que cambie cualquier archivo `.tf` bajo `module1/` o cualquier archivo `.tf` o `.tfvars` bajo `project1/`, podrías usar la siguiente configuración:

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
- `when_modified` será usado tanto por los plans automáticos como por los ejecutados manualmente.
- `when_modified` seguirá funcionando para los plans ejecutados manualmente incluso cuando autoplan esté deshabilitado.
- El valor predeterminado de `when_modified` incluye `**/*.tf*`, `**/*.tofu`, `**/*.tofu.json`, `**/terragrunt.hcl` e `**/.terraform.lock.hcl`. Un `when_modified` personalizado reemplaza completamente estos valores predeterminados.

### Soporte para Terraform Workspaces

```yaml
version: 3
projects:
   - dir: project1
     workspace: staging
   - dir: project1
     workspace: production
```

Con la configuración anterior, cuando Atlantis determina que la configuración para el directorio `project1` ha cambiado,
ejecutará plan para ambos workspaces `staging` e `production`.

Si quieres `plan` o `apply` para un workspace específico, puedes usar

```shell
atlantis plan -w staging -d project1
```

y

```shell
atlantis apply -w staging -d project1
```

### Uso de archivos .tfvars

Consulta [Custom Workflow Use Cases: Using .tfvars files](custom-workflows.md#tfvars-files)

### Agregar argumentos extra a los comandos de Terraform

Consulta [Custom Workflow Use Cases: Adding extra arguments to Terraform commands](custom-workflows.md#adding-extra-arguments-to-terraform-commands)

### Comandos init/plan/apply personalizados

Consulta [Custom Workflow Use Cases: Custom init/plan/apply Commands](custom-workflows.md#custom-init-plan-apply-commands)

### Terragrunt

Consulta [Custom Workflow Use Cases: Terragrunt](custom-workflows.md#terragrunt)

### Ejecución de comandos personalizados

Consulta [Custom Workflow Use Cases: Running custom commands](custom-workflows.md#running-custom-commands)

### Distribuciones de Terraform

Si quieres usar una distribución de Terraform diferente de la que está establecida
por la bandera `--default-tf-version`, entonces establece la clave `terraform_distribution`:

```yaml
version: 3
projects:
   - dir: project1
     terraform_distribution: opentofu
```

Atlantis descargará y usará automáticamente esta distribución. Los valores válidos son `terraform` e `opentofu`.
Si se omite `terraform_version` y el proyecto usa una restricción `required_version`, Atlantis resuelve esa
restricción contra la distribución seleccionada.

### Versiones de Terraform

Si quieres usar una versión de Terraform diferente de la que está en `PATH` de Atlantis
o está establecida por la bandera `--default-tf-version`, entonces establece la clave `terraform_version`:

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
`plan_requirements`, `apply_requirements` e `import_requirements` son claves restringidas, por lo que este repo necesitará estar configurado
para que se le permita establecer esta clave. Consulta [Server-Side Repo Config Use Cases](server-side-repo-config.md#repos-can-set-their-own-apply-requirements).
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

Si algún plan/apply falla e `abort_on_execution_order_fail` está establecido en true a nivel de repo, todos los
grupos siguientes serán abortados. Para este ejemplo, si project2 falla entonces project1 no se ejecutará.

Los grupos de orden de ejecución son útiles cuando tienes dependencias entre proyectos. Sin embargo, solo son aplicables en el caso en que
inicies un apply global para todos tus proyectos, es decir `atlantis apply`. Si inicias un apply en un solo proyecto, entonces los grupos de orden de ejecución se ignoran.
Por lo tanto, la clave `depends_on` es más útil en este caso. y puede usarse en conjunto con los grupos de orden de ejecución.

La siguiente configuración es un ejemplo de cómo usar juntos los grupos de orden de ejecución y depends_on para imponer dependencias entre proyectos.

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

la funcionalidad `depends_on` se asegurará de que `production` no se aplique antes que `staging`, por ejemplo.

::: tip
¿Qué sucede si una o más dependencias de un proyecto no están aplicadas?

Si hay uno o más proyectos en la lista de dependencias que no están en estado applied, los usuarios verán un mensaje de error como este:
`Can't apply your project unless you apply its dependencies`
:::

### Configuración de Autodiscovery

```yaml
autodiscover:
   mode: "auto"
```

Lo anterior es la configuración predeterminada para `autodiscover.mode`. Cuando `autodiscover.mode` es auto,
los proyectos serán descubiertos solo si el repo no tiene `projects` configurado.

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

Con la configuración anterior, Atlantis intentará incondicionalmente descubrir proyectos basados en modified_files,
incluso cuando falte el directorio del proyecto en los `projects` configurados en la configuración del repo.
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

Autodiscover también puede configurarse para omitir directorios que coincidan con un path glob (como lo define el [paquete de coincidencia de rutas doublestar](https://pkg.go.dev/github.com/bmatcuk/doublestar/v4)).

Cuando `ignore_paths` está establecido, se aplica a:

- Descubrimiento automático de proyectos durante autoplan e `atlantis plan` (sin `-d`)
- `atlantis apply` (sin `-d`) al filtrar plans pendientes
- Todos los comandos `-d` dirigidos (`plan`, `apply`, `import`, `state rm`, etc.) cuando autodiscovery está habilitado, si la ruta no tiene configuración de proyecto explícita

Esto hace que `ignore_paths` sea útil para **configuraciones multi-instance** donde cada instancia de Atlantis administra un subárbol de directorios diferente. Por ejemplo, una instancia puede ignorar `environments/prod/**` mientras otra ignora `environments/nonprod/**`, evitando la interferencia entre instancias en comandos dirigidos.

### Configuración de backend personalizada

Consulta [Custom Workflow Use Cases: Custom Backend Config](custom-workflows.md#custom-backend-config)

## Referencia

### Claves de nivel superior

```yaml
version: 3
automerge: false
delete_source_branch_on_merge: false
projects:
workflows:
allowed_regexp_prefixes:
```

| Key                           | Type                                                   | Default | Required | Description                                                                                                                                       |
| ----------------------------- | ------------------------------------------------------ | ------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| version                       | int                                                    | none    | **yes**  | Esta clave es requerida y debe establecerse en `3`.                                                                                               |
| automerge                     | bool                                                   | `false` | no       | Fusiona automáticamente el pull request cuando todos los plans son applied.                                                                       |
| delete_source_branch_on_merge | bool                                                   | `false` | no       | Elimina automáticamente la rama fuente al fusionar.                                                                                               |
| projects                      | array[[Project](repo-level-atlantis-yaml.md#project)]  | `[]`    | no       | Enumera los proyectos en este repo.                                                                                                               |
| workflows<br />_(restricted)_ | map[string: [Workflow](custom-workflows.md#reference)] | `{}`    | no       | Workflows personalizados.                                                                                                                         |
| allowed_regexp_prefixes       | array\[string\]                                        | `[]`    | no       | Enumera los prefijos regexp permitidos para usar cuando se utiliza la bandera [`--enable-regexp-cmd`](server-configuration.md#enable-regexp-cmd). |

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

| Key                                     | Type                    | Default         | Required | Description                                                                                                                                                                                                                                                                 |
| --------------------------------------- | ----------------------- | --------------- | -------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| name                                    | string                  | none            | maybe    | Requerido si hay más de un proyecto con el mismo `dir` e `workspace`. Este nombre de proyecto puede usarse con la bandera `-p`.                                                                                                                                             |
| branch                                  | string                  | none            | no       | Regex que coincide con proyectos por la rama base del pull request (la rama en la que se va a fusionar el pull request). Solo se considerarán los proyectos que coincidan con la rama del PR. De forma predeterminada, todas las ramas coinciden.                           |
| dir                                     | string                  | none            | **yes**  | El directorio de este proyecto relativo a la raíz del repo. Por ejemplo, si el proyecto estuviera bajo `./project1` entonces usa `project1`. Usa `.` para indicar la raíz del repo.                                                                                         |
| workspace                               | string                  | `"default"`     | no       | El [Terraform workspace](https://developer.hashicorp.com/terraform/language/state/workspaces) del proyecto. Atlantis cambia a él en plan/apply. No debe contener `/`, `\\`, `..`, `$`, espacios en blanco ni caracteres de control, ni comenzar con `-` o `~`.              |
| execution_order_group                   | int                     | `0`             | no       | Índice del grupo de orden de ejecución. Los proyectos se ordenarán por este campo antes de planning/applying.                                                                                                                                                               |
| delete_source_branch_on_merge           | bool                    | `false`         | no       | Elimina automáticamente la rama fuente al fusionar.                                                                                                                                                                                                                         |
| repo_locking                            | bool                    | `true`          | no       | (deprecated) Obtiene un bloqueo del repositorio en este proyecto cuando se hace plan.                                                                                                                                                                                       |
| repo_locks                              | [RepoLocks](#repolocks) | `mode: on_plan` | no       | Obtiene un bloqueo del repositorio en este proyecto en plan o apply. Consulta [RepoLocks](#repolocks) para más detalles.                                                                                                                                                    |
| custom_policy_check                     | bool                    | `false`         | no       | Habilita el uso de herramientas de policy check distintas de Conftest                                                                                                                                                                                                       |
| autoplan                                | [Autoplan](#autoplan)   | none            | no       | Una configuración autoplan personalizada. Si no se especifica, usará la configuración de autoplan. Consulta [Autoplanning](autoplanning.md).                                                                                                                                |
| terraform_version                       | string                  | none            | no       | Una versión específica de Terraform para usar al ejecutar comandos para este proyecto. Debe ser [compatible con Semver](https://semver.org/), p. ej. `v0.11.0`, `0.12.0-beta1`.                                                                                             |
| plan_requirements<br />_(restricted)_   | array\[string\]         | none            | no       | Requisitos que deben cumplirse antes de que `atlantis plan` pueda ejecutarse. Actualmente los únicos requisitos soportados son `approved`, `mergeable` e `undiverged`. Consulta [Command Requirements](command-requirements.md) para más detalles.                          |
| apply_requirements<br />_(restricted)_  | array\[string\]         | none            | no       | Requisitos que deben cumplirse antes de que `atlantis apply` e `atlantis state rm` puedan ejecutarse. Actualmente los únicos requisitos soportados son `approved`, `mergeable` e `undiverged`. Consulta [Command Requirements](command-requirements.md) para más detalles.  |
| import_requirements<br />_(restricted)_ | array\[string\]         | none            | no       | Requisitos que deben cumplirse antes de que `atlantis import` e `atlantis state rm` puedan ejecutarse. Actualmente los únicos requisitos soportados son `approved`, `mergeable` e `undiverged`. Consulta [Command Requirements](command-requirements.md) para más detalles. |
| silence_pr_comments                     | array\[string\]         | none            | no       | Silencia los comentarios del PR de las etapas definidas mientras preserva las verificaciones de estado del PR. Los valores soportados son: `plan`, `apply`.                                                                                                                 |
| workflow <br />_(restricted)_           | string                  | none            | no       | Un workflow personalizado. Si no se especifica, Atlantis usará su workflow predeterminado.                                                                                                                                                                                  |

::: tip
Un proyecto representa un estado de Terraform. Típicamente, hay un estado por directorio y workspace; sin embargo, es posible
tener múltiples estados en el mismo directorio usando `terraform init -backend-config=custom-config.tfvars`.
Atlantis soporta esto pero requiere que se especifique la clave `name`. Consulta [Custom Backend Config](custom-workflows.md#custom-backend-config) para más detalles.
:::

#### Directorios de proyecto con symlink

Atlantis resuelve symlinks en el `dir` de un proyecto antes de ejecutar cualquier comando en
él. Un `dir` que es, o pasa por, un symlink a otro directorio dentro
del repo funciona como antes. Un `dir` que se resuelve a una ubicación fuera del
repo clonado es rechazado con `project path traversal detected: path escapes base directory`.
Un `dir` cuyo destino del symlink no existe falla con `dir "<dir>" does not exist`,
y un `dir` que no puede resolverse por cualquier otra razón falla con
`resolving project path: <error>`. Si una ruta de archivo de plan dentro del directorio del proyecto
es un symlink o cualquier cosa distinta de un archivo regular, Atlantis lo elimina
antes de escribir el nuevo plan.

::: warning Cambio de comportamiento
Anteriormente, Atlantis verificaba el `dir` del proyecto solo como texto, por lo que un symlink
en el repo podía apuntar el proyecto a un directorio fuera del clon, y se seguía un
symlink en la ruta del archivo de plan cuando se escribía el plan. Ahora
los symlinks se resuelven primero, un `dir` fuera del clon se rechaza con
`project path traversal detected: path escapes base directory`, un `dir` que
no puede resolverse ahora falla en lugar de que el error sea ignorado, y un archivo de plan con symlink
se reemplaza con un archivo regular. Los repos cuyos `dir` de proyecto enlazan fuera del
repo deben mover ese código dentro del repo o usar un directorio real.
:::

### Autoplan

```yaml
enabled: true
when_modified: ["*.tf", "terragrunt.hcl", ".terraform.lock.hcl"]
```

| Key           | Type            | Default   | Required | Description                                                                                                                                                                                                                                                                                 |
| ------------- | --------------- | --------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| enabled       | boolean         | `true`    | no       | Si autoplanning está habilitado para este proyecto.                                                                                                                                                                                                                                         |
| when_modified | array\[string\] | see below | no       | Usa la sintaxis de [.dockerignore](https://docs.docker.com/engine/reference/builder/#dockerignore-file). Si cualquier archivo modificado en el pull request coincide, se hará plan de este proyecto. Consulta [Autoplanning](autoplanning.md). Las rutas son relativas al dir del proyecto. |

Patrones predeterminados de `when_modified`: `["**/*.tf*", "**/*.tofu", "**/*.tofu.json", "**/terragrunt.hcl", "**/.terraform.lock.hcl"]`. Los valores personalizados de `when_modified` reemplazan completamente estos valores predeterminados. El valor predeterminado es global (no depende de la distribución).

### RepoLocks

```yaml
mode: on_apply
```

| Key  | Type   | Default   | Required | Description                                                                                                                                           |
| ---- | ------ | --------- | -------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| mode | `Mode` | `on_plan` | no       | Si los bloqueos de repositorio están habilitados o no para este proyecto en plan o apply. Los valores válidos son `disabled`, `on_plan` e `on_apply`. |
