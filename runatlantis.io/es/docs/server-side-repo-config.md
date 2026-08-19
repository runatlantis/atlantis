# Configuración de repositorio del lado del servidor

Un archivo de configuración del lado del servidor se usa para más grupos de configuración del servidor que no pueden expresarse razonablemente mediante flags.

Un caso de uso de este tipo es controlar el comportamiento por repositorio
y lo que los usuarios pueden hacer en archivos `atlantis.yaml` a nivel de repositorio.

## ¿Necesito un archivo de configuración del lado del servidor?

No necesita un archivo de configuración de repositorio del lado del servidor a menos que quiera personalizar
algún aspecto de Atlantis por repositorio.

Lea los [casos de uso](#use-cases) para determinar si lo necesita.

## Habilitar configuración del lado del servidor

Para usar configuración de repositorio del lado del servidor, cree un archivo de configuración, p. ej. `repos.yaml`, y páselo al
comando `atlantis server` mediante el flag `--repo-config`, p. ej. `--repo-config=path/to/repos.yaml`.

Si no desea escribir un archivo de configuración en disco, puede usar el
flag `--repo-config-json` o la variable de entorno `ATLANTIS_REPO_CONFIG_JSON`
para especificar su configuración como JSON. Vea [--repo-config-json](server-configuration.md#repo-config-json)
para un ejemplo.

## Ejemplo de repositorio del lado del servidor

```yaml
# repos lists the config for specific repos.
repos:
  # id can either be an exact repo ID or a regex.
  # If using a regex, it must start and end with a slash.
  # Repo ID's are of the form {VCS hostname}/{org}/{repo name}, ex.
  # github.com/runatlantis/atlantis.
- id: /.*/
  # branch is a regex matching pull requests by base branch
  # (the branch the pull request is getting merged into).
  # By default, all branches are matched
  branch: /.*/

  # repo_config_file specifies which repo config file to use for this repo.
  # By default, atlantis.yaml is used.
  repo_config_file: path/to/atlantis.yaml

  # plan_requirements sets the Plan Requirements for all repos that match.
  plan_requirements: [approved, mergeable, undiverged]

  # apply_requirements sets the Apply Requirements for all repos that match.
  apply_requirements: [approved, mergeable, undiverged]

  # import_requirements sets the Import Requirements for all repos that match.
  import_requirements: [approved, mergeable, undiverged]

  # workflow sets the workflow for all repos that match.
  # This workflow must be defined in the workflows section.
  workflow: custom

  # allowed_overrides specifies which keys can be overridden by this repo in
  # its atlantis.yaml file.
  allowed_overrides: [apply_requirements, workflow, delete_source_branch_on_merge, repo_locking, repo_locks, custom_policy_check]

  # allowed_workflows specifies which workflows the repos that match
  # are allowed to select.
  allowed_workflows: [custom]

  # allow_custom_workflows defines whether this repo can define its own
  # workflows. If false (default), the repo can only use server-side defined
  # workflows.
  allow_custom_workflows: true

  # delete_source_branch_on_merge defines whether the source branch would be deleted on merge
  # If false (default), the source branch won't be deleted on merge
  delete_source_branch_on_merge: true

  # repo_locking defines whether lock repository when planning.
  # If true (default), atlantis try to get a lock.
  # deprecated: use repo_locks instead
  repo_locking: true

  # repo_locks defines whether the repository would be locked on apply instead of plan, or disabled
  # Valid values are on_plan (default), on_apply or disabled.
  repo_locks:
    mode: on_plan

  # custom_policy_check defines whether policy checking tools besides Conftest are enabled in checks
  # If false (default), only Conftest JSON output is allowed
  custom_policy_check: false

  # pre_workflow_hooks defines arbitrary list of scripts to execute before workflow execution.
  pre_workflow_hooks:
    - run: my-pre-workflow-hook-command arg1

  # post_workflow_hooks defines arbitrary list of scripts to execute after workflow execution.
  post_workflow_hooks:
    - run: my-post-workflow-hook-command arg1

  # policy_check defines if policy checking should be enabled on this repository.
  policy_check: false

  # autodiscover defines how atlantis should automatically discover projects in this repository.
  # If any part of this setting is set here, it overrides the entire setting in the repo config.
  autodiscover:
    mode: auto
    # Optionally ignore some paths for autodiscovery by a glob path.
    # When autodiscovery is enabled, also applies to all targeted -d commands
    # (plan, apply, import, etc.) when the path has no explicit project configuration.
    ignore_paths:
      - foo/*

  # id can also be an exact match.
- id: github.com/myorg/specific-repo

# workflows lists server-side custom workflows
workflows:
  custom:
    plan:
      steps:
      - run: my-custom-command arg1 arg2
      - init
      - plan:
          extra_args: ["-lock", "false"]
      - run: my-custom-command arg1 arg2
    apply:
      steps:
      - run: echo hi
      - apply
 ```

## Casos de uso

Aquí hay algunas de las razones por las que podría querer usar una configuración de repositorio.

### Requerir que el PR esté aprobado antes de un subcomando aplicable

Si quiere requerir que todos los repositorios (o repositorios específicos) deban tener pull requests
aprobados antes de que Atlantis permita ejecutar `apply` o `import`, use las claves `plan_requirements`, `apply_requirements` o `import_requirements`.

Para todos los repositorios:

```yaml
# repos.yaml
repos:
- id: /.*/
  plan_requirements: [approved]
  apply_requirements: [approved]
  import_requirements: [approved]
```

Para un repositorio específico:

```yaml
# repos.yaml
repos:
- id: github.com/myorg/myrepo
  plan_requirements: [approved]
  apply_requirements: [approved]
  import_requirements: [approved]
```

Vea [Requisitos de comandos](command-requirements.md) para más detalles.

### Requerir que el PR sea "mergeable" antes de Apply o Import

Si quiere requerir que todos los repositorios (o repositorios específicos) deban tener pull requests
en un estado mergeable antes de que Atlantis permita ejecutar `apply` o `import`, use las claves `plan_requirements`, `apply_requirements` o `import_requirements`.

Para todos los repositorios:

```yaml
# repos.yaml
repos:
- id: /.*/
  plan_requirements: [mergeable]
  apply_requirements: [mergeable]
  import_requirements: [mergeable]
```

Para un repositorio específico:

```yaml
# repos.yaml
repos:
- id: github.com/myorg/myrepo
  plan_requirements: [mergeable]
  apply_requirements: [mergeable]
  import_requirements: [mergeable]
```

Vea [Requisitos de comandos](command-requirements.md) para más detalles.

### Los repositorios pueden establecer sus propios requisitos de Apply

Si quiere que todos los repositorios (o repositorios específicos) puedan sobrescribir los requisitos predeterminados de apply, use
la clave `allowed_overrides`.

Para permitir que todos los repositorios sobrescriban el valor predeterminado:

```yaml
# repos.yaml
repos:
- id: /.*/
  # The default will be approved.
  plan_requirements: [approved]
  apply_requirements: [approved]
  import_requirements: [approved]

  # But all repos can set their own using atlantis.yaml
  allowed_overrides: [plan_requirements, apply_requirements, import_requirements]
```

Para permitir que solo un repositorio específico sobrescriba el valor predeterminado:

```yaml
# repos.yaml
repos:
# Set a default for all repos.
- id: /.*/
  plan_requirements: [approved]
  apply_requirements: [approved]
  import_requirements: [approved]

# Allow a specific repo to override.
- id: github.com/myorg/myrepo
  allowed_overrides: [plan_requirements, apply_requirements, import_requirements]
```

Luego, cada repositorio permitido puede tener un archivo `atlantis.yaml` que
establezca `plan_requirements`, `apply_requirements` o `import_requirements` en un arreglo vacío (deshabilitando el requisito).

```yaml
# atlantis.yaml in the repo root or set repo_config_file in repos.yaml
version: 3
projects:
- dir: .
  plan_requirements: []
  apply_requirements: []
  import_requirements: []
```

### Ejecutar scripts antes de los workflows de Atlantis

Si quiere ejecutar scripts que se ejecutarían antes de que Atlantis pueda ejecutar workflows predeterminados o
personalizados, puede crear un `pre-workflow-hooks`:

```yaml
repos:
  - id: /.*/
    pre_workflow_hooks:
      - run: my custom command
      - run: |
          my bash script inline
```

Vea [Pre Workflow Hooks](pre-workflow-hooks.md) para más detalles sobre cómo escribir
hooks de pre workflow.

### Ejecutar scripts después de los workflows de Atlantis

Si quiere ejecutar scripts que se ejecutarían después de que Atlantis ejecute workflows predeterminados o
personalizados, puede crear un `post-workflow-hooks`:

```yaml
repos:
  - id: /.*/
    post_workflow_hooks:
      - run: my custom command
      - run: |
          my bash script inline
```

Vea [Post Workflow Hooks](post-workflow-hooks.md) para más detalles sobre cómo escribir
hooks de post workflow.

### Cambiar el workflow predeterminado de Atlantis

Si quiere cambiar los comandos predeterminados que Atlantis ejecuta durante las fases `plan` e `apply`,
puede crear un nuevo `workflow`.

Si quiere usar ese workflow de forma predeterminada para todos los repositorios, use la
clave de workflow `default`:

```yaml
# repos.yaml
# NOTE: the repos key is not required.
workflows:
  # It's important that this is "default".
  default:
    plan:
      steps:
      - init
      - run: my custom plan command
    apply:
      steps:
      - run: my custom apply command
```

Vea [Custom Workflows](custom-workflows.md) para más detalles sobre cómo escribir
workflows personalizados.

### Permitir que los repositorios elijan un workflow del lado del servidor

Si quiere que los repositorios puedan elegir sus propios workflows que están definidos
en la configuración de repositorio del lado del servidor, necesita crear los workflows
del lado del servidor y luego permitir que cada repositorio sobrescriba la clave `workflow`:

```yaml
# repos.yaml
# Allow repos to override the workflow key.
repos:
- id: /.*/
  allowed_overrides: [workflow]

# Define your custom workflows.
workflows:
  custom1:
    plan:
      steps:
      - init
      - run: my custom plan command
    apply:
      steps:
      - run: my custom apply command

  custom2:
    plan:
      steps:
      - run: another custom command
    apply:
      steps:
      - run: another custom command
```

O, si quiere restringir a qué workflows tiene acceso cada repositorio, use la
clave `allowed_workflows`:

```yaml
# repos.yaml
# Restrict which workflows repos can select.
repos:
- id: /.*/
  allowed_overrides: [workflow]

- id: /my_repo/
  allowed_overrides: [workflow]
  allowed_workflows: [custom1]

# Define your custom workflows.
workflows:
  custom1:
    plan:
      steps:
      - init
      - run: my custom plan command
    apply:
      steps:
      - run: my custom apply command

  custom2:
    plan:
      steps:
      - run: another custom command
    apply:
      steps:
      - run: another custom command
```

Luego, cada repositorio permitido puede elegir uno de los workflows en sus archivos `atlantis.yaml`:

```yaml
# atlantis.yaml
version: 3
projects:
- dir: .
  workflow: custom1 # could also be custom2 OR default
```

:::tip NOTA
Siempre hay un workflow llamado `default` que corresponde al workflow predeterminado de Atlantis
a menos que haya creado su propio workflow del lado del servidor con esa clave (sobrescribiéndolo).
:::

Vea [Custom Workflows](custom-workflows.md) para más detalles sobre cómo escribir
workflows personalizados.

### Permitir el uso de herramientas de políticas personalizadas

Conftest es la aplicación estándar de verificación de políticas integrada con Atlantis, pero herramientas personalizadas aún pueden ejecutarse en workflows personalizados cuando la opción `custom_policy_check` está establecida. Vea la [página de Custom Policy Checks](custom-policy-checks.md) para ejemplos detallados.

### Permitir que los repositorios definan sus propios workflows

Si quiere que los repositorios puedan definir sus propios workflows necesita
permitirles sobrescribir la clave `workflow` y establecer `allow_custom_workflows` en `true`.

::: danger
Si los repositorios pueden definir sus propios workflows, entonces cualquiera que pueda crear un pull
request para ese repositorio puede esencialmente ejecutar código arbitrario en su servidor Atlantis.
:::

```yaml
# repos.yaml
repos:
- id: /.*/

  # With just allowed_overrides: [workflow], repos can only
  # choose workflows defined server-side.
  allowed_overrides: [workflow]

  # By setting allow_custom_workflows to true, we allow repos to also
  # define their own workflows.
  allow_custom_workflows: true
```

Luego, cada repositorio permitido puede definir y usar un workflow personalizado en sus archivos `atlantis.yaml`:

```yaml
# atlantis.yaml
version: 3
projects:
- dir: .
  workflow: custom1
workflows:
  custom1:
    plan:
      steps:
      - init
      - run: my custom plan command
    apply:
      steps:
      - run: my custom apply command
```

Vea [Custom Workflows](custom-workflows.md) para más detalles sobre cómo escribir
workflows personalizados.

### Varios servidores Atlantis manejan el mismo repositorio

Ejecutar varios servidores Atlantis para manejar el mismo repositorio puede hacerse para separar permisos para cada servidor Atlantis.
En este caso, puede usarse un archivo de configuración de repositorio [atlantis.yaml](repo-level-atlantis-yaml.md) diferente usando diferentes archivos `repos.yaml`.

Por ejemplo, considere una situación donde Atlantis de `production-server` usa la configuración de repositorio `atlantis-production.yaml` y Atlantis de `staging-server` usa la configuración de repositorio `atlantis-staging.yaml`.

Primero, despliegue 2 servidores Atlantis, `production-server` e `staging-server`.
Cada servidor tiene permisos diferentes y un archivo `repos.yaml` diferente.
El `repos.yaml` contiene la clave `repo_config_file` para especificar la ruta del archivo de configuración atlantis del repositorio.

```yaml
# repos.yaml
repos:
- id: /.*/
  # for production-server
  repo_config_file: atlantis-production.yaml
  # for staging-server
  # repo_config_file: atlantis-staging.yaml
```

Luego, cree los archivos `atlantis-production.yaml` e `atlantis-staging.yaml` en el repositorio.
Vea los ejemplos de configuración en [atlantis.yaml](repo-level-atlantis-yaml.md).

```yaml
# atlantis-production.yaml
version: 3
projects:
- name: project
  branch: /production/
  dir: infrastructure/production
---
# atlantis-staging.yaml
version: 3
projects:
  - name: project
    branch: /staging/
    dir: infrastructure/staging
```

Ahora, pueden configurarse 2 URLs de webhook para el repositorio, que envían eventos a `production-server` e `staging-server` respectivamente.
Cada servidor maneja diferentes archivos de configuración del repositorio.

:::tip Notas

* Si los comentarios `no projects` son molestos, establezca [--silence-no-projects](server-configuration.md#silence-no-projects).
* El nombre del ejecutable disparador de comandos puede reconfigurarse de `atlantis` a otra cosa configurando [Executable Name](server-configuration.md#executable-name).
* Al usar diferentes usuarios vcs del servidor atlantis como `@atlantis-staging`, el comentario `@atlantis-staging plan` puede usarse en lugar de `atlantis plan` para llamar solo a `staging-server`.
:::

## Referencia

### Claves de nivel superior

| Key        | Type                                                  | Default   | Required | Description                                                                           |
|------------|-------------------------------------------------------|-----------|----------|---------------------------------------------------------------------------------------|
| repos      | array[[Repo](#repo)]                                  | see below | no       | Lista de repositorios a los que aplicar la configuración.                             |
| workflows  | map[string: [Workflow](custom-workflows.md#workflow)] | see below | no       | Mapa del nombre del workflow al workflow. Los workflows sobrescriben los comandos predeterminados de Atlantis. |
| policies   | Policies.                                             | none      | no       | Lista de conjuntos de políticas a ejecutar y metadatos asociados                      |
| metrics    | Metrics.                                              | none      | no       | Mapa de configuración de métricas                                                     |
| team_authz | [TeamAuthz](#teamauthz)                               | none      | no       | Configuración de verificación de permisos de equipo                                   |

::: tip Una nota sobre los valores predeterminados

#### `repos`

`repos` siempre contiene un primer elemento con la configuración predeterminada de Atlantis:

```yaml
repos:
- id: /.*/
  branch: /.*/
  plan_requirements: []
  apply_requirements: []
  import_requirements: []
  workflow: default
  allowed_overrides: []
  allow_custom_workflows: false
```

#### `workflows`

`workflows` siempre contiene el workflow predeterminado de Atlantis bajo la clave `default`:

```yaml
workflows:
  default:
    plan:
      steps: [init, plan]
    apply:
      steps: [apply]
```

Esto se fusiona con cualquier configuración que escriba.
Si establece un workflow con la clave `default`, esto lo sobrescribirá.
:::

### Repo

| Key | Type | Default | Required | Description |
| --- | --- | --- | --- | --- |
| id | string | none | yes | El valor puede ser una expresión regular cuando se especifica como /&lt;regex&gt;/ o una coincidencia exacta de cadena. Los ID de repositorio tienen la forma `{vcs hostname}/{org}/{name}`, p. ej. `github.com/owner/repo`. El hostname se especifica sin esquema ni puerto. Para Bitbucket Server, {org} es el **nombre** del proyecto, no la clave. |
| branch | string | none | no | Una regex que coincide con pull requests por rama base (la rama en la que se va a fusionar el pull request). De forma predeterminada, todas las ramas coinciden |
| repo_config_file | string | none | no | Ruta del archivo de configuración del repositorio en este repo. De forma predeterminada, usa `atlantis.yaml` que está ubicado en la raíz del repositorio. Cuando varios servidores atlantis trabajan con el mismo repo, establezca nombres de archivo diferentes. |
| workflow | string | none | no | Un workflow personalizado. |
| plan_requirements | []string | none | no | Requisitos que deben satisfacerse antes de que pueda ejecutarse `atlantis plan`. Actualmente los únicos requisitos compatibles son `approved`, `mergeable` e `undiverged`. Vea [Command Requirements](command-requirements.md) para más detalles. |
| apply_requirements | []string | none | no | Requisitos que deben satisfacerse antes de que pueda ejecutarse `atlantis apply`. Actualmente los únicos requisitos compatibles son `approved`, `mergeable` e `undiverged`. Vea [Command Requirements](command-requirements.md) para más detalles. |
| import_requirements | []string | none | no | Requisitos que deben satisfacerse antes de que pueda ejecutarse `atlantis import`. Actualmente los únicos requisitos compatibles son `approved`, `mergeable` e `undiverged`. Vea [Command Requirements](command-requirements.md) para más detalles. |
| allowed_overrides | []string | none | no | Una lista de claves restringidas que los archivos `atlantis.yaml` pueden sobrescribir. Las únicas claves compatibles son `apply_requirements`, `workflow`, `delete_source_branch_on_merge`,`repo_locking`, `repo_locks` y `custom_policy_check` |
| allowed_workflows | []string | none | no | Una lista de workflows de los que los archivos `atlantis.yaml` pueden seleccionar. |
| allow_custom_workflows | bool | false | no | Si se permite o no [Custom Workflows](custom-workflows.md). |
| delete_source_branch_on_merge | bool | false | no | Si se elimina o no la rama de origen al fusionar. |
| repo_locking | bool | false | no | (obsoleto) Si se obtiene o no un lock. |
| repo_locks | [RepoLocks](#repolocks) | `mode: on_plan` | no | Si los locks de repositorio están habilitados o no para este proyecto en plan o apply. Vea [RepoLocks](#repolocks) para más detalles. |
| policy_check | bool | false | no | Si se ejecutan o no verificaciones de políticas en este repositorio. |
| custom_policy_check | bool | false | no | Si se habilitan o no herramientas personalizadas de verificación de políticas fuera de Conftest en este repositorio. |
| autodiscover | AutoDiscover | none | no | Configuración de autodiscover para este repo |
| silence_pr_comments | []string | none | no | Silencia los comentarios de PR de las etapas definidas mientras preserva las verificaciones de estado del PR. Útil en entornos grandes con muchas instancias de Atlantis y/o proyectos, cuando los comentarios son demasiado grandes y demasiados, por lo tanto es preferible depender únicamente de las verificaciones de estado del PR. Los valores compatibles son: `plan`, `apply`. |

:::tip Notas

* Si varios repositorios coinciden, se aplicará la última coincidencia.
* Si una clave no está definida, no sobrescribirá una clave que coincidió arriba.
  Por ejemplo, dado un ID de repositorio `github.com/owner/repo` y una configuración:

  ```yaml
  repos:
  - id: /.*/
    allow_custom_workflows: true
    apply_requirements: [approved]
  - id: github.com/owner/repo
    apply_requirements: []
  ```

  La configuración final se verá así:

  ```yaml
  apply_requirements: []
  workflow: default
  allowed_overrides: []
  allow_custom_workflows: true
  ```

  Donde
  * `apply_requirements` se establece desde la configuración `id: github.com/owner/repo` porque
    sobrescribe la configuración coincidente anterior de `id: /.*/`.
  * `workflow` se establece desde la configuración predeterminada que siempre
    existe.
  * `allowed_overrides` se establece desde la configuración predeterminada que siempre
    existe.
  * `allow_custom_workflows` se establece desde la configuración `id: /.*/` y no se elimina
    por la configuración `id: github.com/owner/repo` porque no definió esa clave.
:::

### RepoLocks

```yaml
mode: on_apply
```

| Key  | Type   | Default   | Required | Description                                                                                                                           |
|------|--------|-----------|----------|---------------------------------------------------------------------------------------------------------------------------------------|
| mode | `Mode` | `on_plan` | no       | Si los locks de repositorio están habilitados o no para este proyecto en plan o apply. Los valores válidos son `disabled`, `on_plan` e `on_apply`. |

### Policies

| Key | Type | Default | Required | Description |
| --- | --- | --- | --- | --- |
| conftest_version | string | none | no | versión de conftest para ejecutar todos los conjuntos de políticas |
| owners | Owners(#Owners) | none | yes | owners que pueden aprobar políticas fallidas |
| approve_count | int | 1 | no | número de aprobaciones requeridas para omitir políticas fallidas. |
| sticky_policy_approvals | bool | false | no | cuando es true, las aprobaciones de políticas sobreviven a re-plans siempre que no se introduzcan nuevos elementos de salida de políticas (por `policy_item_regex`). Vea [Sticky Policy Approvals](policy-checking.md#sticky-policy-approvals). |
| policy_item_regex | string | `(?s).+` | no | regex para extraer elementos comparables de la salida de políticas para el seguimiento de aprobaciones persistentes. El valor predeterminado coincide con toda la salida como un solo elemento. Vea [Sticky Policy Approvals](policy-checking.md#sticky-policy-approvals). |
| policy_sets | []PolicySet | none | yes | conjunto de políticas a ejecutar sobre una salida de plan |

### Owners

| Key         | Type              | Default | Required   | Description                                             |
|-------------|-------------------|---------|------------|---------------------------------------------------------|
| users       | []string          | none    | no         | lista de usuarios de github que pueden aprobar políticas fallidas |
| teams       | []string          | none    | no         | lista de equipos de github que pueden aprobar políticas fallidas |
|

### PolicySet

| Key                      | Type   | Default   | Required | Description                                                                                                                                               |
|--------------------------|--------|-----------|----------|-----------------------------------------------------------------------------------------------------------------------------------------------------------|
| name                     | string | none      | yes      | nombre único para el conjunto de políticas                                                                                                                |
| path                     | string | none      | yes      | ruta al directorio de políticas rego                                                                                                                      |
| source                   | string | none      | yes      | solo se admite `local` en este momento                                                                                                                     |
| owners                   | Owners | none      | no       | owners que pueden aprobar este conjunto de políticas específico (fusionado con los owners de nivel superior)                                            |
| approve_count            | int    | inherited | no       | número de aprobaciones requeridas. El valor predeterminado es el valor de nivel superior `approve_count`                                                       |
| prevent_self_approve     | bool   | false     | no       | si el autor del PR puede aprobar políticas. El valor predeterminado es `false` (el autor también debe estar en owners)                                |
| sticky_policy_approvals  | bool   | inherited | no       | sobrescribe el valor de nivel superior `sticky_policy_approvals` para este conjunto de políticas. Vea [Sticky Policy Approvals](policy-checking.md#sticky-policy-approvals). |
| policy_item_regex        | string | inherited | no       | sobrescribe el valor de nivel superior `policy_item_regex` para este conjunto de políticas. Vea [Sticky Policy Approvals](policy-checking.md#sticky-policy-approvals). |
|

### Metrics

| Key                    | Type                      | Default | Required  | Description                              |
|------------------------|---------------------------|---------|-----------|------------------------------------------|
| statsd                 | [Statsd](#statsd)         | none    | no        | proveedor de métricas Statsd             |
| prometheus             | [Prometheus](#prometheus) | none    | no        | proveedor de métricas Prometheus         |

### Statsd

| Key    | Type   | Default | Required | Description                            |
| ------ | ------ | ------- | -------- | -------------------------------------- |
| host   | string | none    | yes      | dirección IP del host de statsd        |
| port   | string | none    | yes      | puerto de statsd                       |

### Prometheus

| Key      | Type   | Default | Required | Description                            |
| -------- | ------ | ------- | -------- | -------------------------------------- |
| endpoint | string | none    | yes      | ruta al endpoint de métricas           |

### TeamAuthz

| Key     | Type     | Default | Required | Description                                 |
|---------|----------|---------|----------|---------------------------------------------|
| command | string   | none    | yes      | ruta completa al comando externo de autorización |
| args    | []string | none    | no       | argumentos opcionales para pasar a `command`     |
