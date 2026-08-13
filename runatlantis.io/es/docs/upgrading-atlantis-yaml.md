# Actualizar atlantis.yaml

## Actualizar de v2 a v3

La versión de Atlantis `v0.7.0` introdujo una nueva versión 3 de `atlantis.yaml`.

**Si no estás usando pasos de [custom `run`](custom-workflows.md#custom-run-command),
 entonces puedes actualizar de `version: 2` a `version: 3` sin ningún cambio.**

**NOTA:** La versión 2 **no se está deprecando** y no hay necesidad de actualizar tu versión
si no deseas hacerlo.

El único cambio de v2 a v3 es que estamos analizando los pasos custom `run` de manera diferente.

```yaml
# atlantis.yaml
workflows:
  custom:
    plan:
      steps:
      - run: my custom command
```

<center><i>Un ejemplo de workflow que usa un paso custom run</i></center>

Anteriormente, usábamos una biblioteca que analizaba el paso custom antes de ejecutarlo.
Ahora, simplemente ejecutamos el paso directamente. Esto solo afectará a tus pasos si estaban usando escape de shell de algún tipo.
Por ejemplo, si tu paso era anteriormente:

```yaml
# version: 2
- run: "printf \'print me\'"
```

Ahora puedes escribir esto en la versión 3 como:

```yaml
# version: 3
- run: "printf 'print me'"
```

## Actualizar de V1 a V3

Si estás actualizando desde una versión **antigua** de Atlantis `<=v0.3.10` (de antes del 4 de julio de 2018)
necesitarás seguir los siguientes pasos.

### Un solo atlantis.yaml

Si tenías múltiples archivos `atlantis.yaml` por directorio, entonces necesitarás
consolidarlos en un solo archivo `atlantis.yaml` en la raíz del repo.

Por ejemplo, si tenías una estructura de directorios:

```plain
.
├── project1
│   └── atlantis.yaml
└── project2
    └── atlantis.yaml
```

Entonces tu nueva estructura se vería así:

```plain
.
├── atlantis.yaml
├── project1
└── project2
```

Y tu `atlantis.yaml` se vería algo así:

```yaml
version: 2
projects:
- dir: project1
  terraform_version: my-version
  workflow: project1-workflow
- dir: project2
  terraform_version: my-version
  workflow: project2-workflow
workflows:
  project1-workflow:
    ...
  project2-workflow:
    ...
```

Hablaremos más sobre `workflows` abajo.

### Versión de Terraform

La clave `terraform_version` pasó de ser una clave de nivel superior a ser por `project`,
así que si antes tu `atlantis.yaml` estaba en el directorio `mydir` y se veía así:

```yaml
terraform_version: 0.11.0
```

Entonces tu nueva configuración sería:

```yaml
version: 2
projects:
- dir: mydir
  terraform_version: 0.11.0
```

### Workflows

Los workflows son la nueva manera de establecer todos los `pre_*`, `post_*` y `extra_arguments`.

Cada `project` puede tener un workflow custom mediante la clave `workflow`.

```yaml
version: 2
projects:
- dir: .
  workflow: myworkflow
```

Los workflows se definen como una clave de nivel superior:

```yaml
version: 2
projects:
...

workflows:
  myworkflow:
  ...
```

Para comenzar, determina si estás personalizando comandos que ocurren durante
`plan` o `apply`. Luego estableces esa clave bajo el nombre del workflow:

```yaml
...
workflows:
  myworkflow:
    plan:
      steps:
      ...
    apply:
      steps:
      ...
```

Si no estás personalizando una etapa específica, entonces puedes omitir esa clave. Por ejemplo,
si solo estás personalizando los comandos que ocurren durante `plan`, entonces tu configuración
se verá así:

```yaml
...
workflows:
  myworkflow:
    plan:
      steps:
      ...
```

#### Argumentos extra

`extra_arguments` ahora se especifica de la siguiente manera. Dada una configuración previa:

```yaml
extra_arguments:
  - command_name: init
    arguments:
    - "-lock=false"
  - command_name: plan
    arguments:
    - "-lock=false"
  - command_name: apply
    arguments:
    - "-lock=false"
```

Tu configuración ahora se vería así:

```yaml
...
workflows:
  myworkflow:
    plan:
      steps:
      - init:
          extra_args: ["-lock=false"]
      - plan:
          extra_args: ["-lock=false"]
    apply:
      steps:
      - apply:
          extra_args: ["-lock=false"]
```

#### Comandos pre/post

En lugar de usar `pre_*` o `post_*`, ahora puedes insertar tus comandos custom
antes/después de los comandos integrados. Dada una configuración previa:

```yaml
pre_init:
  commands:
  - "curl http://example.com"
# pre_get commands are run when the Terraform version is < 0.9.0
pre_get:
  commands:
  - "curl http://example.com"
pre_plan:
  commands:
  - "curl http://example.com"
post_plan:
  commands:
  - "curl http://example.com"
pre_apply:
  commands:
  - "curl http://example.com"
post_apply:
  commands:
  - "curl http://example.com"
```

Tu configuración ahora se vería así:

```yaml
...
workflows:
  myworkflow:
    plan:
      steps:
      - run: curl http://example.com
      - init
      - plan
      - run: curl http://example.com
    apply:
      steps:
      - run: curl http://example.com
      - apply
      - run: curl http://example.com
```

::: tip
Es importante incluir los comandos integrados: `init`, `plan` y `apply`.
De lo contrario, Atlantis no ejecutará los comandos necesarios para realmente hacer plan/apply.
:::
