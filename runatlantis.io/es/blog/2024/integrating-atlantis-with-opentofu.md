---
title: Integrating Atlantis with OpenTofu
lang: en-US
---

# Integrar Atlantis con OpenTofu

::: info
Esta publicación fue escrita originalmente el 27 de mayo de 2024
Publicación original: <https://dev.to/jmateusousa/integrating-atlantis-with-opentofu-lnd>
:::

## ¿Cuál fue nuestra motivación?

Debido al cambio de licencia de Terraform, muchas empresas están migrando sus procesos de IAC a OpenTofu; con esto en mente y sabiendo que muchas de ellas usan Atlantis y Terraform como automatización de entrega de infraestructura, creé esta documentación mostrando qué hacer para integrar Atlantis con OpenTofu.

Stack: Atlantis, Terragrunt, OpenTofu, Github, ALB, EKS.

Lo implementaremos con su [Helm chart](https://www.runatlantis.io/docs/deployment.html#kubernetes-helm-chart):

**1** - Agregue el repositorio runatlantis.

```sh
helm repo add runatlantis https://runatlantis.github.io/helm-charts
```

**2** - Cree el archivo values.yaml y ejecute:

```sh
helm inspect values runatlantis/atlantis > values.yaml
```

**3** - Edite el archivo values.yaml y agregue sus credenciales access y secret que se usarán en la configuración del webhook de Atlantis:
Vea cómo crear una [GitHubApp](https://docs.github.com/pt/apps/creating-github-apps/about-creating-github-apps).

```yaml
githubApp:
  id: "CHANGE ME"
  key: |
    -----BEGIN RSA PRIVATE KEY-----
            "CHANGE ME"
    -----END RSA PRIVATE KEY-----
  slug: atlantis
# secret webhook Atlantis
  secret: "CHANGE ME"
```

**4** - Ingrese la organización y el repositorio de github con los que Atlantis interactuará en orgAllowlist:

```yaml
# All repositories the org
orgAllowlist: github.com/MY-ORG/*

or
# Just one repository
orgAllowlist: github.com/MY-ORG/MY-REPO-IAC

or
# All repositories that start with MY-REPO-IAC-
orgAllowlist: github.com/MY-ORG/MY-REPO-IAC-*
```

**5** - Ahora configuremos el script que se ejecutará al iniciar el pod init de Atlantis. En este paso descargamos e instalamos Terragrunt y OpenTofu, así como incluimos sus binarios en el dir compartido

 ```/plugins```.

```yaml
initConfig:
  enabled: true
  image: alpine:latest
  imagePullPolicy: IfNotPresent
  # sharedDir is set as env var INIT_SHARED_DIR
  sharedDir: /plugins
  workDir: /tmp
  sizeLimit: 250Mi
  # example of how the script can be configured to install tools/providers required by the atlantis pod
  script: |
    #!/bin/sh
    set -eoux pipefail# terragrunt
    TG_VERSION="0.55.10"
    TG_SHA256_SUM="1ad609399352348a41bb5ea96fdff5c7a18ac223742f60603a557a54fc8c6cff"
    TG_FILE="${INIT_SHARED_DIR}/terragrunt"
    wget https://github.com/gruntwork-io/terragrunt/releases/download/v${TG_VERSION}/terragrunt_linux_amd64 -O "${TG_FILE}"
    echo "${TG_SHA256_SUM} ${TG_FILE}" | sha256sum -c
    chmod 755 "${TG_FILE}"
    terragrunt -v

    # OpenTofu
    TF_VERSION="1.6.2"
    TF_FILE="${INIT_SHARED_DIR}/tofu"
    wget https://github.com/opentofu/opentofu/releases/download/v${TF_VERSION}/tofu_${TF_VERSION}_linux_amd64.zip
    unzip tofu_${TF_VERSION}_linux_amd64.zip
    mv tofu ${INIT_SHARED_DIR}
    chmod 755 "${TF_FILE}"
    tofu -v
```

**6** - Aquí configuramos las envs para evitar descargar versiones alternativas de Terraform e indicar a Terragrunt dónde debe obtener el binario de OpenTofu.

```yaml
# envs
environment:
  ATLANTIS_TF_DOWNLOAD: false
  TERRAGRUNT_TFPATH: /plugins/tofu
```

**7** - Por último, pero no menos importante, aquí especificamos qué configuraciones del lado de Atlantis tendremos para los repositorios.

```yaml
# repository config
repoConfig: |
  ---
  repos:
  - id: /.*/
    apply_requirements: [approved, mergeable]
    allow_custom_workflows: true
    allowed_overrides: [workflow, apply_requirements, delete_source_branch_on_merge]
```

**8** - Configure el ingress del webhook de Atlantis; en el siguiente ejemplo estamos usando AWS ALB.

```yaml
# ingress config
ingress:
  annotations:
    alb.ingress.kubernetes.io/backend-protocol: HTTP
    alb.ingress.kubernetes.io/certificate-arn: arn:aws:acm:certificate
    alb.ingress.kubernetes.io/group.name: external-atlantis
    alb.ingress.kubernetes.io/healthcheck-path: /healthz
    alb.ingress.kubernetes.io/healthcheck-port: "80"
    alb.ingress.kubernetes.io/healthcheck-protocol: HTTP
    alb.ingress.kubernetes.io/listen-ports: '[{"HTTPS":443}]'
    alb.ingress.kubernetes.io/scheme: internet-facing
    alb.ingress.kubernetes.io/ssl-redirect: "443"
    alb.ingress.kubernetes.io/success-codes: "200"
    alb.ingress.kubernetes.io/target-type: ip
  apiVersion: networking.k8s.io/v1
  enabled: true
  host: atlantis.your.domain
  ingressClassName: aws-ingress-class-name
  path: /*
  pathType: ImplementationSpecific
```

Guarde todos los cambios realizados en

 ```values.yaml```

**9** - Using one of the Atlantis options custom workflows, we can create a file ```atlantis.yaml``` in the root folder of your repository, the example below should meet most scenarios, adapt as needed.

```yaml
version: 3
automerge: true
parallel_plan: true
parallel_apply: false
projects:
- name: terragrunt
  dir: .
  workspace: terragrunt
  delete_source_branch_on_merge: true
  autoplan:
    enabled: false
  apply_requirements: [mergeable, approved]
  workflow: terragrunt
workflows:
  terragrunt:
    plan:
      steps:
      - env:
          name: TF_IN_AUTOMATION
          value: 'true'
      - run: find . -name '.terragrunt-cache' | xargs rm -rf
      - run: terragrunt init -reconfigure
      - run:
          command: terragrunt plan -input=false -out=$PLANFILE
          output: strip_refreshing
    apply:
      steps:
        - run: terragrunt apply $PLANFILE
```

**10** - Ahora vayamos a la instalación en sí; busque las versiones disponibles de Atlantis:

```sh
helm search repo runatlantis
```

Reemplace

 ```CHART-VERSION``` with the version you want to install and run the command below:

```sh
helm upgrade -i atlantis runatlantis/atlantis --version CHART-VERSION -f values.yaml --create-namespace atlantis
```

Ahora, vea cómo configurar el [webhook en github](../../docs/configuring-webhooks.md) del repositorio para Atlantis.

Vea cómo Atlantis [funciona](../../docs/using-atlantis.md).

Descubra más en:

- <https://www.runatlantis.io/guide.html>.
- <https://opentofu.org/docs/>.
- <https://github.com/runatlantis/atlantis/issues/3741>.

Compártalo con sus amigos =)
