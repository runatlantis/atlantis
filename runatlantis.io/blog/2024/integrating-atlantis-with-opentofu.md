---
title: Integrating Atlantis with Opentofu
lang: en-US
---

# Integrating Atlantis with Opentofu

::: info
This post was originally written on May 27nd, 2024
Original post: <https://dev.to/jmateusousa/integrating-atlantis-with-opentofu-lnd>
:::

## What was our motivation?

Due to the Terraform license change, many companies are migrating their IAC processes to OpenTofu, with this in mind and knowing that many of them use Atlantis and Terraform as infrastructure delivery automation, I created this documentation showing what to do to integrate Atlantis with OpenTofu.

Stack: Atlantis, Terragrunt, OpenTofu, Github, ALB, EKS.

We will implement it with your [Helm chart](https://www.runatlantis.io/docs/deployment.html#kubernetes-helm-chart):

**1** - Add the runatlantis repository.
```
helm repo add runatlantis https://runatlantis.github.io/helm-charts
```
**2** - Create file values.yaml and run:
```
helm inspect values runatlantis/atlantis > values.yaml
```
**3** - Edit the file values.yaml and add your credentials access and secret which will be used in the Atlantis webhook configuration:
See as create a [GitHubApp](https://docs.github.com/pt/apps/creating-github-apps/about-creating-github-apps).

```
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
**4** - Enter the org and repository from github that Atlantis will interact in orgAllowlist:
```
# All repositories the org
orgAllowlist: github.com/MY-ORG/*

or
# Just one repository
orgAllowlist: github.com/MY-ORG/MY-REPO-IAC

or
# All repositories that start with MY-REPO-IAC-
orgAllowlist: github.com/MY-ORG/MY-REPO-IAC-*
```
**5** - Now let’s configure the script that will be executed upon startup of the Atlantis init pod. In this step we download and install Terragrunt and OpenTofu, as well as include their binaries in the shared dir ```/plugins```.
```
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
**6** - Here we configure the envs to avoid downloading alternative versions of Terraform and indicate to Terragrunt where it should fetch the OpenTofu binary.
```
# envs
environment:
  ATLANTIS_TF_DOWNLOAD: false
  TERRAGRUNT_TFPATH: /plugins/tofu
```
**7** - Last but not least, here we specify which Atlantis-side configurations we will have for the repositories.
```
# repository config
repoConfig: |
  ---
  repos:
  - id: /.*/
    apply_requirements: [approved, mergeable]
    allow_custom_workflows: true
    allowed_overrides: [workflow, apply_requirements, delete_source_branch_on_merge]
```
**8** - Configure Atlantis webhook ingress, in the example below we are using the AWS ALB.
```
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
Save all changes made to ```values.yaml```

**9** - Using one of the Atlantis options custom workflows, we can create a file ```atlantis.yaml``` in the root folder of your repository, the example below should meet most scenarios, adapt as needed.
```
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
**10** - Now let’s go to the installation itself, search for the available versions of Atlantis:
```
helm search repo runatlantis
```
Replace ```CHART-VERSION``` with the version you want to install and run the command below:

```
helm upgrade -i atlantis runatlantis/atlantis --version CHART-VERSION -f values.yaml --create-namespace atlantis
```

Now, see as configure Atlantis [webhook on github](../../docs/configuring-webhooks.md) repository.

See as Atlantis [work](../../docs/using-atlantis.md).

Find out more at:

https://www.runatlantis.io/guide.html.
https://opentofu.org/docs/.
https://github.com/runatlantis/atlantis/issues/3741.

Share it with your friends =)
