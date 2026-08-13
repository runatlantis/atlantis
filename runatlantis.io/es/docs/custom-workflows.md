# Workflows personalizados

Se pueden definir workflows personalizados para sobrescribir los comandos predeterminados que Atlantis
ejecuta.

## Uso

Los workflows personalizados se pueden especificar en la Configuración de Repositorio del Lado del Servidor o en los archivos `atlantis.yaml` a Nivel de Repositorio.

**Notas:**

* Si desea permitir que los repos seleccionen sus propios workflows, deben tener la configuración `allowed_overrides: [workflow]`. Consulte [casos de uso de la configuración de repositorio del lado del servidor](server-side-repo-config.md#allow-repos-to-choose-a-server-side-workflow) para más detalles.
* Si además también desea permitir que los repos definan sus propios workflows, deben tener la configuración `allow_custom_workflows: true`. Consulte [casos de uso de la configuración de repositorio del lado del servidor](server-side-repo-config.md#allow-repos-to-define-their-own-workflows) para más detalles.

## Casos de uso

### Archivos .tfvars

::: tip
Antes de crear workflows personalizados para archivos `.tfvars`, considere usar la funcionalidad automática de Atlantis `env/{workspace}.tfvars`. Si estructura sus archivos como `env/staging.tfvars`, `env/production.tfvars`, etc., Atlantis los incluirá automáticamente según el workspace sin ninguna configuración. Consulte [Usar Atlantis - Archivos automáticos de variables de entorno](using-atlantis.md#automatic-environment-variable-files) para obtener detalles.
:::

Dada la estructura:

```plain
.
└── project1
    ├── main.tf
    ├── production.tfvars
    └── staging.tfvars
```

Si quisiera que Atlantis ejecutara automáticamente plan con `-var-file staging.tfvars` y `-var-file production.tfvars`
podría definir dos workflows:

```yaml
# repos.yaml or atlantis.yaml
workflows:
  staging:
    plan:
      steps:
      - init
      - plan:
          extra_args: ["-var-file", "staging.tfvars"]
    # NOTE: no need to define the apply stage because it will default
    # to the normal apply stage.

  production:
    plan:
      steps:
      - init
      - plan:
          extra_args: ["-var-file", "production.tfvars"]
    apply:
      steps:
        - apply:
            extra_args: ["-var-file", "production.tfvars"]
    import:
      steps:
        - init
        - import:
            extra_args: ["-var-file", "production.tfvars"]
    state_rm:
      steps:
        - init
        - state_rm:
            extra_args: ["-lock=false"]
```

Luego, en su archivo `atlantis.yaml` a nivel de repositorio, haría referencia a los workflows:

```yaml
# atlantis.yaml
version: 3
projects:
# If two or more projects have the same dir and workspace, they must also have
# a 'name' key to differentiate them.
- name: project1-staging
  dir: project1
  workflow: staging
- name: project1-production
  dir: project1
  workflow: production

workflows:
  # If you didn't define the workflows in your server-side repos.yaml config,
  # you would define them here instead.
```

Cuando quiera aplicar los plans, puede comentar

```shell
atlantis apply -p project1-staging
```

y

```shell
atlantis apply -p project1-production
```

Donde `-p` se refiere al nombre del proyecto.

### Agregar argumentos extra a comandos de Terraform

Si necesita anexar flags a `terraform plan` o `apply` temporalmente, puede
anexar flags en un comentario después de `--`, por ejemplo comentando:

```shell
atlantis plan -- -lock=false
```

Si siempre necesita hacer esto para los comandos `init`, `plan` o `apply` de un proyecto,
entonces debe definir un workflow personalizado y establecer la clave `extra_args` para el comando que necesite modificar.

```yaml
# atlantis.yaml or repos.yaml
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

Si la [verificación de políticas](policy-checking.md#how-it-works) está habilitada, `extra_args` también se puede usar para cambiar el comportamiento predeterminado de conftest.

```yaml
workflows:
  myworkflow:
    policy_check:
      steps:
      - show
      - policy_check:
          extra_args: ["--all-namespaces"]
```

### Comandos init/plan/apply personalizados

Si desea personalizar `terraform init`, `plan` o `apply` de formas que no están admitidas por `extra_args`, puede sobrescribir completamente esos comandos.

En este ejemplo, no estamos usando ninguno de los comandos integrados y en su lugar
estamos usando los nuestros.

```yaml
# atlantis.yaml or repos.yaml
workflows:
  myworkflow:
    plan:
      steps:
      # If you want to hide command output from Atlantis's PR comment, use
      # the output option on the run step's expanded form.
      - run:
          command: terraform init -input=false
          output: hide

      # If you're using workspaces you need to select the workspace using the
      # $WORKSPACE environment variable.
      - run: terraform workspace select $WORKSPACE

      # You MUST output the plan using -out $PLANFILE because Atlantis expects
      # plans to be in a specific location.
      - run: terraform plan -input=false -refresh -out $PLANFILE
    apply:
      steps:
      # Again, you must use the $PLANFILE environment variable.
      - run: terraform apply $PLANFILE
```

### CDKTF

Estos son los requisitos para habilitar [CDKTF](https://developer.hashicorp.com/terraform/cdktf)

* Una imagen personalizada con `CDKTF` instalado
* Agregue `**/cdk.tf.json` a la lista de archivos de autoplan de Atlantis.
* Establezca el flag `atlantis-include-git-untracked-files` para que los archivos Terraform generados dinámicamente
por CDKTF se agreguen a la lista de archivos modificados de Atlantis.
* Use `pre_workflow_hooks` para ejecutar `cdktf synth`
* Opcional: No hay un requisito de usar un archivo de repositorio `atlantis.yaml`, pero se puede aprovechar si es necesario.

#### Imagen personalizada

```dockerfile
# Dockerfile
FROM ghcr.io/runatlantis/atlantis:v0.19.7

USER root
RUN apk add npm && npm i -g cdktf-cli
```

#### Configuración del servidor

```bash
# env variables
ATLANTIS_AUTOPLAN_FILE_LIST="**/*.tf,**/*.tfvars,**/*.tfvars.json,**/cdk.tf.json"
ATLANTIS_INCLUDE_GIT_UNTRACKED_FILES=true
```

O

`atlantis server --config config.yaml`

```yaml
# config.yaml
autoplan-file-list: "**/*.tf,**/*.tfvars,**/*.tfvars.json,**/cdk.tf.json"
include-git-untracked-files: true
```

#### Configuración de repositorio del servidor

Use `pre_workflow_hooks`

`atlantis server --repo-config="repos.yaml"`

```yaml
# repos.yaml
repos:
  - id: /.*cdktf.*/
    pre_workflow_hooks:
      - run: npm i && cdktf get && cdktf synth --output ci-cdktf.out
```

**Nota:** no use el directorio predeterminado `cdktf.out` que usa CDKTF, ya que este debe estar en la lista `.gitignore` del repositorio, para que los archivos generados localmente no se hagan check-in.

#### Estructura del repositorio

Esta es la estructura del repositorio git después de ejecutar `cdktf synth`. Los archivos `cdk.tf.json` contienen la configuración Terraform
que atlantis puede ejecutar.

```bash
$ tree --gitignore
.
├── cdktf.json
├── ci-cdktf.out
│   ├── manifest.json
│   └── stacks
│       └── eks
│           └── cdk.tf.json
```

#### Workflow

1. El orquestador de contenedores (k8s/fargate/ecs/etc) usa la imagen docker personalizada de atlantis con `cdktf` instalado con
`--autoplan-file-list` para activarse en archivos `cdk.tf.json` y `--include-git-untracked-files` establecido para incluir los
archivos Terraform generados dinámicamente por CDKTF en el plan de Atlantis.
1. Se hace push de la rama del PR que contiene cambios de código `cdktf`.
1. Atlantis hace checkout de la rama en el repositorio.
1. Atlantis ejecuta el comando `npm i && cdktf get && cdktf synth` en la raíz del repositorio como un paso en `pre_workflow_hooks`,
generando los archivos Terraform `cdk.tf.json`.
1. Atlantis detecta los archivos no rastreados `cdk.tf.json` en varios directorios.
1. Atlantis luego ejecuta workflows `terraform` en los directorios respectivos como de costumbre.

### Terragrunt

Atlantis admite ejecutar comandos personalizados en lugar de los comandos predeterminados de Atlantis. Podemos usar esta funcionalidad para habilitar
[Terragrunt](https://github.com/gruntwork-io/terragrunt).

Puede usar el archivo `atlantis.yaml` de su repositorio o el archivo `repos.yaml` del servidor Atlantis.

Atlantis selecciona la distribución y versión de Terraform de cada proyecto. En los workflows a continuación,
`ATLANTIS_TERRAFORM_DISTRIBUTION` se expande al prefijo ejecutable (`terraform` o `tofu`), y combinarlo con
`ATLANTIS_TERRAFORM_VERSION` hace que Terragrunt apunte al mismo binario versionado. Esto también admite repositorios que contienen
tanto proyectos Terraform como OpenTofu.

Dada una estructura de directorios:

```plain
.
└── live
    ├── prod
    │   └── terragrunt.hcl
    └── staging
        └── terragrunt.hcl
```

Si usa el archivo `repos.yaml` del servidor, usaría la siguiente configuración:

```yaml
# repos.yaml
# Generate json plan via terragrunt for policy checks
repos:
- id: "/.*/"
  workflow: terragrunt
workflows:
  terragrunt:
    plan:
      steps:
      - env:
          name: TG_TF_PATH
          command: 'echo "${ATLANTIS_TERRAFORM_DISTRIBUTION}${ATLANTIS_TERRAFORM_VERSION}"'
      - env:
          # Reduce Terraform suggestion output
          name: TF_IN_AUTOMATION
          value: 'true'
      - run:
          # Allow for targeted plans/applies as not supported for Terraform wrappers by default
          command: terragrunt plan -input=false $(printf '%s' $COMMENT_ARGS | sed 's/,/ /g' | tr -d '\\') -no-color -out $PLANFILE
          output: hide
      - run: |
          terragrunt show $PLANFILE
    apply:
      steps:
      - env:
          name: TG_TF_PATH
          command: 'echo "${ATLANTIS_TERRAFORM_DISTRIBUTION}${ATLANTIS_TERRAFORM_VERSION}"'
      - env:
          # Reduce Terraform suggestion output
          name: TF_IN_AUTOMATION
          value: 'true'
      - run: terragrunt apply -input=false $PLANFILE
    import:
      steps:
      - env:
          name: TG_TF_PATH
          command: 'echo "${ATLANTIS_TERRAFORM_DISTRIBUTION}${ATLANTIS_TERRAFORM_VERSION}"'
      - env:
          name: TF_VAR_author
          command: 'git show -s --format="%ae" $HEAD_COMMIT'
      # Allow for imports as not supported for Terraform wrappers by default
      - run: terragrunt import -input=false $(printf '%s' $COMMENT_ARGS | sed 's/,/ /' | tr -d '\\')
    state_rm:
      steps:
      - env:
          name: TG_TF_PATH
          command: 'echo "${ATLANTIS_TERRAFORM_DISTRIBUTION}${ATLANTIS_TERRAFORM_VERSION}"'
      # Allow for state removals as not supported for Terraform wrappers by default
      - run: terragrunt state rm $(printf '%s' $COMMENT_ARGS | sed 's/,/ /' | tr -d '\\')
```

Si usa el archivo `atlantis.yaml` del repositorio, usaría la siguiente configuración:

```yaml
version: 3
projects:
- dir: live/staging
  workflow: terragrunt
- dir: live/prod
  workflow: terragrunt
workflows:
  terragrunt:
    plan:
      steps:
      - env:
          name: TG_TF_PATH
          command: 'echo "${ATLANTIS_TERRAFORM_DISTRIBUTION}${ATLANTIS_TERRAFORM_VERSION}"'
      - env:
          # Reduce Terraform suggestion output
          name: TF_IN_AUTOMATION
          value: 'true'
      - run:
          command: terragrunt plan -input=false -out=$PLANFILE
          output: strip_refreshing
    apply:
      steps:
      - env:
          name: TG_TF_PATH
          command: 'echo "${ATLANTIS_TERRAFORM_DISTRIBUTION}${ATLANTIS_TERRAFORM_VERSION}"'
      - env:
          # Reduce Terraform suggestion output
          name: TF_IN_AUTOMATION
          value: 'true'
      - run: terragrunt apply $PLANFILE
```

**NOTA:** Si usa el archivo `atlantis.yaml` del repositorio, necesitará especificar cada directorio que sea un proyecto Terragrunt.

::: warning
Atlantis necesitará tener el binario `terragrunt` en su PATH.
Si está usando Docker puede construir su propia imagen, consulte [Personalización](deployment.md#customization).
:::

Si no desea crear/gestionar usted mismo el archivo `atlantis.yaml` del repositorio, puede usar la herramienta [terragrunt-atlantis-config](https://github.com/transcend-io/terragrunt-atlantis-config) para generarlo.

La herramienta `terragrunt-atlantis-config` es un proyecto de la comunidad y no es mantenida por el equipo de Atlantis.

### Ejecutar comandos personalizados

Atlantis admite ejecutar comandos completamente personalizados. En este ejemplo, queremos ejecutar
un script después de cada `apply`:

```yaml
# repos.yaml or atlantis.yaml
workflows:
  myworkflow:
    apply:
      steps:
      - apply
      - run: ./my-custom-script.sh
```

::: tip Notas

* No necesitamos escribir una clave `plan` bajo `myworkflow`. Si `plan`
no está establecido, Atlantis usará el workflow plan predeterminado, que es lo que queremos en este caso.
* Un comando personalizado solo terminará si todos los descriptores de archivo de salida están cerrados.
Por lo tanto, un comando personalizado solo puede enviarse al background (p. ej., para un túnel SSH durante
la ejecución de terraform) cuando su salida se redirige a una ubicación diferente. Por ejemplo, Atlantis
ejecutará correctamente un script personalizado que contenga el siguiente código para crear un túnel SSH:
`ssh -f -M -S /tmp/ssh_tunnel -L 3306:database:3306 -N bastion 1>/dev/null 2>&1`. Sin
la redirección, el script bloquearía el workflow de Atlantis.
:::

### Configuración de backend personalizada

Si necesita especificar el flag `-backend-config` para `terraform init` necesitará usar un workflow personalizado.
En este ejemplo, estamos usando archivos backend personalizados para configurar dos estados remotos, uno para cada entorno.
Luego usamos archivos `.tfvars` para cargar variables diferentes para cada entorno.

```yaml
# repos.yaml or atlantis.yaml
workflows:
  staging:
    plan:
      steps:
      - run: rm -rf .terraform
      - init:
          extra_args: [-backend-config=staging.backend.tfvars]
      - plan:
          extra_args: [-var-file=staging.tfvars]
  production:
    plan:
      steps:
      - run: rm -rf .terraform
      - init:
          extra_args: [-backend-config=production.backend.tfvars]
      - plan:
          extra_args: [-var-file=production.tfvars]
```

::: warning NOTE
Tenemos que usar un paso personalizado `run` para `rm -rf .terraform` porque de lo contrario Terraform
se quejará entre comandos ya que la configuración del backend ha cambiado.
:::

Luego haría referencia a los workflows en su archivo `atlantis.yaml` a nivel de repositorio:

```yaml
version: 3
projects:
- name: staging
  dir: .
  workflow: staging
- name: production
  dir: .
  workflow: production
```

### Agregar contexto de directorio y repositorio para recursos aws usando default tags

Esto solo está disponible en la versión del provider AWS [5.62.0](https://github.com/hashicorp/terraform-provider-aws/releases/tag/v5.62.0) y superiores.

Esta configuración creará las siguientes tags

* `repository` igual a `github.com/<owner>/<repo>` que se puede cambiar para gitlab u otro VCS
* `repository_dir` igual al directorio relativo

Se pueden agregar otras variables predeterminadas, como para workspace. Vea abajo más variables de entorno disponibles.

```yaml
workflows:
  terraform:
    plan:
      steps:
        # These env vars TF_AWS_DEFAULT_TAGS_ will work for aws provider 5.62.0+
        # https://github.com/hashicorp/terraform-provider-aws/releases/tag/v5.62.0
        - &env_default_tags_repository
          env:
            name: TF_AWS_DEFAULT_TAGS_repository
            command: 'echo "github.com/${BASE_REPO_OWNER}/${BASE_REPO_NAME}"'
        - &env_default_tags_repository_dir
          env:
            name: TF_AWS_DEFAULT_TAGS_repository_dir
            command: 'echo "${REPO_REL_DIR}"'
    apply:
      steps:
        - *env_default_tags_repository
        - *env_default_tags_repository_dir
```

NOTA:

* Anexar tags a cada recurso puede regenerar fuentes de datos como `aws_iam_policy_document` lo que causará que muchos recursos sean modificados. Vea el problema conocido en el provider aws [#29421](https://github.com/hashicorp/terraform-provider-aws/issues/29421).

* Para ejecutar un plan local fuera de terraform, será necesario crear las mismas variables de entorno.

    ```bash
    tfvars () {
      export terraform_repository=$(git config --get remote.origin.url | sed 's,^git@,,g' | tr ':' '/' | sed 's,.git$,,g')
      export terraform_repository_dir=$(git rev-parse --show-prefix | sed 's,\/$,,g')
    }
    export TF_AWS_DEFAULT_TAGS_repository=$terraform_repository
    export TF_AWS_DEFAULT_TAGS_repository_dir=$terraform_repository_dir
    tfvars
    terraform plan
    ```

    Si se usa dos puntos en el nombre de la tag, use el comando `env` en lugar de `export`.

    ```bash
    tfvars
    env \
      TF_AWS_DEFAULT_TAGS_org:repository=$terraform_repository \
      TF_AWS_DEFAULT_TAGS_org:repository_dir=$terraform_repository_dir \
      terraform plan
    ```

## Referencia

### Workflow

```yaml
plan:
apply:
import:
state_rm:
```

| Key      | Type            | Default                   | Required | Description                           |
|----------|-----------------|---------------------------|----------|---------------------------------------|
| plan     | [Stage](#stage) | `steps: [init, plan]`     | no       | Cómo hacer plan para este proyecto.         |
| apply    | [Stage](#stage) | `steps: [apply]`          | no       | Cómo hacer apply para este proyecto.        |
| import   | [Stage](#stage) | `steps: [init, import]`   | no       | Cómo hacer import para este proyecto.       |
| state_rm | [Stage](#stage) | `steps: [init, state_rm]` | no       | Cómo ejecutar state rm para este proyecto. |

### Stage

```yaml
steps:
- run: custom-command
- init
- plan:
    extra_args: [-lock=false]
```

| Key   | Type                 | Default | Required | Description                                                                                   |
|-------|----------------------|---------|----------|-----------------------------------------------------------------------------------------------|
| steps | array[[Step](#step)] | `[]`    | no       | Lista de pasos para esta etapa. Si la clave steps está vacía, no se ejecutará ningún paso para esta etapa. |

### Step

#### Comandos integrados

Los pasos pueden ser una sola cadena para un comando integrado.

```yaml
- init
- plan
- apply
- import
- state_rm
```

| Key                             | Type   | Default | Required | Description                                                                                                                  |
|---------------------------------|--------|---------|----------|------------------------------------------------------------------------------------------------------------------------------|
| init/plan/apply/import/state_rm | string | none    | no       | Use un comando integrado sin configuración adicional. Solo se admiten `init`, `plan`, `apply`, `import` y `state_rm` |

#### Comando integrado con args extra

Un map de string a `extra_args` para un comando integrado con argumentos extra.

```yaml
- init:
    extra_args: [arg1, arg2]
- plan:
    extra_args: [arg1, arg2]
- apply:
    extra_args: [arg1, arg2]
- import:
    extra_args: [arg1, arg2]
- state_rm:
    extra_args: [arg1, arg2]
```

| Key | Type | Default | Required | Description |
| --- | --- | --- | --- | --- |
| init/plan/apply/import/state_rm | map\[`extra_args` -> array\[string\]\] | none | no | Use un comando integrado y anexe `extra_args`. Solo se admiten `init`, `plan`, `apply`, `import` y `state_rm` como claves y solo se admite `extra_args` como valor |

#### Comando `run` personalizado

Un comando personalizado puede escribirse de 2 maneras

Compacto:

```yaml
- run: custom-command arg1 arg2
```

| Key | Type   | Default | Required | Description          |
|-----|--------|---------|----------|----------------------|
| run | string | none    | no       | Ejecutar un comando personalizado |

Ejemplo completo:

```yaml
- run:
    command: custom-command arg1 arg2
    shell: sh
    shellArgs:
     - "--debug"
     - "-c"
    output: show
```

Ejemplo completo, filtrando la salida y enmascarando el texto coincidente (`mySecret: "foo"` -> `mySecret: "<redacted>"`):

```yaml
- run:
    command: custom-command arg1 arg2
    shell: sh
    shellArgs:
     - "--debug"
     - "-c"
    output:
      - strip_refreshing
      - filter_regex: "((?i)secret:\\s\")[^\"]*"
```

| Key | Type | Default | Required | Description |
| ----- | ----- | ----- | ----- | ----- |
| run | map\[string -> string\] | none | no | Ejecutar un comando personalizado |
| run.command | string | none | yes | Comando de shell a ejecutar |
| run.shell | string | "sh" | no | Nombre del shell que se usará para la ejecución del comando |
| run.shellArgs | string or []string | "-c" | no | Argumentos de línea de comandos que se pasarán al shell. No se puede establecer sin `shell` |
| run.output | string or []string or []any | "show" | no | Cómo postprocesar la salida de este comando cuando se publique en el comentario del PR. Las opciones son:<br/>*`show` - preservar la salida completa<br/>* `hide` - ocultar la salida del comentario (sigue siendo visible en la salida de streaming en tiempo real)<br/> `strip_refreshing` - ocultar toda la salida hasta e incluyendo la última línea que contenga "Refreshing...". Esto coincide con el comportamiento del comando integrado `plan` <br/> `filter_regex: "<regex_pattern>"` - enmascara texto sensible en comentarios de Atlantis reemplazando coincidencias regex con &lt;redacted&gt;. Se puede usar varias veces (procesadas en orden). Solo filtra comentarios inline: los enlaces al plan completo todavía muestran resultados sin filtrar. |

#### Variables de entorno nativas

* Los pasos `run` en el `workflow` principal se ejecutan con las siguientes variables de entorno:
  nota: estas variables no están disponibles para workflows `pre` o `post`
  * `WORKSPACE` - El workspace de Terraform usado para este proyecto, ej. `default`.
      NOTA: si el paso se ejecuta antes de `init` entonces Atlantis todavía no habrá cambiado a este workspace.
  * `ATLANTIS_TERRAFORM_VERSION` - La versión de Terraform usada para este proyecto, ej. `0.11.0`.
  * `DIR` - Ruta absoluta al directorio actual.
  * `PLANFILE` - Ruta absoluta a la ubicación donde Atlantis espera que el plan
      sea generado (por plan) o ya exista (si se ejecuta apply). Se puede usar para
      sobrescribir los comandos integrados `plan`/`apply`, ej. `run: terraform plan -out $PLANFILE`.
  * `SHOWFILE` - Ruta absoluta a la ubicación donde Atlantis espera que el plan en formato json
      sea generado (por show) o ya exista (si se ejecutan verificaciones de políticas). Se puede usar para
      sobrescribir los comandos integrados `plan`/`apply`, ej. `run: terraform show -json $PLANFILE > $SHOWFILE`.
  * `POLICYCHECKFILE` - Ruta absoluta a la ubicación de la salida de verificación de políticas si Atlantis ejecuta verificaciones de políticas.
      Consulte [verificación de políticas](policy-checking.md#data-for-custom-run-steps) para información sobre la estructura de datos.
  * `BASE_REPO_NAME` - Nombre del repositorio en el que se fusionará el pull request, ej. `atlantis`.
  * `BASE_REPO_OWNER` - Propietario del repositorio en el que se fusionará el pull request, ej. `runatlantis`.
  * `HEAD_REPO_NAME` - Nombre del repositorio que se está fusionando en el repositorio base, ej. `atlantis`.
  * `HEAD_REPO_OWNER` - Propietario del repositorio que se está fusionando en el repositorio base, ej. `acme-corp`.
  * `HEAD_BRANCH_NAME` - Nombre de la rama head del pull request (la rama que se está fusionando en la base)
  * `HEAD_COMMIT` - El sha256 que apunta al head de la rama sobre la que se hace pull request hacia la base. Si el pull request es de Bitbucket Cloud la cadena solo tendrá 12 caracteres porque Bitbucket Cloud trunca sus IDs de commit.
  * `BASE_BRANCH_NAME` - Nombre de la rama base del pull request (la rama en la que se está fusionando el pull request)
  * `PROJECT_NAME` - Nombre del proyecto configurado en `atlantis.yaml`. Si no se configura ningún nombre de proyecto, esto será una cadena vacía.
  * `PULL_NUM` - Número o ID del pull request, ej. `2`.
  * `PULL_URL` - URL del pull request, ej. `https://github.com/runatlantis/atlantis/pull/2`.
  * `PULL_AUTHOR` - Nombre de usuario del autor del pull request, ej. `acme-user`.
  * `REPO_REL_DIR` - La ruta relativa del proyecto en el repositorio. Por ejemplo, si su proyecto está en `dir1/dir2/` entonces esto se establecerá en `"dir1/dir2"`. Si su proyecto está en la raíz esto será `"."`.
  * `USER_NAME` - Nombre de usuario del usuario del VCS que ejecuta el comando, ej. `acme-user`. Durante un autoplan, el usuario será el usuario de la API de Atlantis, ej. `atlantis`.
  * `COMMENT_ARGS` - Cualquier flag adicional pasado en el comentario del pull request. Los flags se separan por comas y
      cada carácter es escapado, ej. `atlantis plan -- arg1 arg2` resultará en `COMMENT_ARGS=\a\r\g\1,\a\r\g\2`.
  * `ATLANTIS_PR_APPROVED` - "true" si el PR está aprobado
  * `ATLANTIS_PR_MERGEABLE` - "true" si el PR se puede fusionar

* Un comando personalizado solo terminará si todos los descriptores de archivo de salida están cerrados.
Por lo tanto, un comando personalizado solo puede enviarse al background (p. ej., para un túnel SSH durante
la ejecución de terraform) cuando su salida se redirige a una ubicación diferente. Por ejemplo, Atlantis
ejecutará correctamente un script personalizado que contenga el siguiente código para crear un túnel SSH:
`ssh -f -M -S /tmp/ssh_tunnel -L 3306:database:3306 -N bastion 1>/dev/null 2>&1`. Sin
la redirección, el script bloquearía el workflow de Atlantis.
* Si un paso del workflow devuelve un código de salida distinto de cero, el workflow se detendrá.
:::

#### Comando de variable de entorno `env`

El comando `env` le permite establecer variables de entorno que estarán disponibles
para todos los pasos definidos **debajo** del paso `env`.

Puede establecer valores codificados mediante la clave `value`, o establecer valores dinámicos mediante
la clave `command` que le permite ejecutar cualquier comando y usa la salida
como valor de la variable de entorno.

```yaml
- env:
    name: ENV_NAME
    value: hard-coded-value
- env:
    name: ENV_NAME_2
    command: 'echo "dynamic-value-$(date)"'
- env:
    name: ENV_NAME_3
    command: echo ${DIR%$REPO_REL_DIR}
    shell: bash
    shellArgs:
      - "--verbose"
      - "-c"
```

| Key | Type | Default | Required | Description |
| ----------------- | ----------------------- | --------- | ---------- | ----------------------------------------------------------------------------------------------------------------- |
| env | map\[string -> string\] | none | no | Establecer variables de entorno para pasos subsiguientes |
| env.name | string | none | yes | Nombre de la variable de entorno |
| env.value | string | none | no | Establecer el valor de la variable de entorno en una cadena codificada. No se puede establecer al mismo tiempo que `command` |
| env.command | string | none | no | Establecer el valor de la variable de entorno en la salida de un comando. No se puede establecer al mismo tiempo que `value` |
| env.shell | string | "sh" | no | Nombre del shell que se usará para la ejecución del comando. No se puede establecer sin `command` |
| env.shellArgs | string or []string | "-c" | no | Argumentos de línea de comandos que se pasarán al shell. No se puede establecer sin `shell` |

::: tip Notas

* Los `command` de `env` pueden usar cualquiera de las variables de entorno integradas disponibles
  para comandos `run`.
:::

#### Comando de múltiples variables de entorno `multienv`

El comando `multienv` le permite establecer una cantidad dinámica de múltiples variables de entorno que estarán disponibles
para todos los pasos definidos **debajo** del paso `multienv`.

Compacto:

```yaml
- multienv: custom-command
```

| Key      | Type   | Default | Required | Description                                                |
|----------|--------|---------|----------|------------------------------------------------------------|
| multienv | string | none    | no       | Ejecutar un comando personalizado y agregar variables de entorno impresas |

Completo:

```yaml
- multienv:
    command: custom-command
    shell: bash
    shellArgs:
      - "--verbose"
      - "-c"
    output: show
```

| Key | Type | Default | Required | Description |
| --- | --- | --- | --- | --- |
| multienv | map[string -> string] | none | no | Ejecutar un comando personalizado y agregar variables de entorno impresas |
| multienv.command | string | none | yes | Nombre del script personalizado a ejecutar |
| multienv.shell | string | "sh" | no | Nombre del shell que se usará para la ejecución del comando |
| multienv.shellArgs | string or []string | "-c" | no | Argumentos de línea de comandos que se pasarán al shell. No se puede establecer sin `shell` |
| multienv.output | string | "show" | no | Establecer la salida en "hide" suprimirá el mensaje sobre variables de entorno agregadas |

La salida de la ejecución del comando debe tener el siguiente formato:
`EnvVar1Name=value1,EnvVar2Name=value2,EnvVar3Name=value3`

Los pares nombre-valor en la salida se agregan como variables de entorno si la ejecución del comando es exitosa; de lo contrario, la ejecución del workflow se interrumpe con un error y se devuelve el errorMessage.

::: tip Notas

* Los `command` de `multienv` pueden usar cualquiera de las variables de entorno integradas disponibles
  para comandos `run`.
:::
