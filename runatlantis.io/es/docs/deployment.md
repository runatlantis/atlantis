# Despliegue

Esta página cubre poner Atlantis en funcionamiento en tu infraestructura.

::: tip Requisitos previos

* Has creado [credenciales de acceso](access-credentials.md) para tu usuario de Atlantis
* Has creado un [secreto de webhook](webhook-secrets.md)
:::

## Resumen de la arquitectura

### Runtime

Atlantis es una aplicación simple de [Go](https://golang.org/). Recibe webhooks de
tu host Git y ejecuta comandos de Terraform localmente. Hay una [imagen Docker](https://ghcr.io/runatlantis/atlantis) oficial de
Atlantis.

### Enrutamiento

Atlantis y tu host Git necesitan poder enrutar y comunicarse entre sí. Tu host Git necesita poder enviar webhooks a Atlantis y Atlantis necesita poder hacer llamadas a la API a tu host Git.
Si estás usando
un host Git público como github.com, gitlab.com, gitea.com, bitbucket.org o dev.azure.com, entonces necesitarás
exponer Atlantis a internet.

Si estás usando un host Git privado como GitHub Enterprise, GitLab Enterprise, Gitea autoalojado o
Bitbucket Server, entonces Atlantis necesita ser enrutable desde el host privado y Atlantis necesitará poder enrutar hacia el host privado.

### Datos

Atlantis no tiene base de datos externa. Atlantis almacena archivos de plan de Terraform en disco.
Si Atlantis pierde esos datos entre un ciclo de `plan` e `apply`, entonces los usuarios tendrán
que volver a ejecutar `plan`. Debido a esto, puede que quieras aprovisionar un disco persistente
para Atlantis.

## Despliegue

Elige tu tipo de despliegue:

* [Kubernetes Helm Chart](#kubernetes-helm-chart)
* [Kubernetes Manifests](#kubernetes-manifests)
* [Kubernetes Kustomize](#kubernetes-kustomize)
* [OpenShift](#openshift)
* [AWS Fargate](#aws-fargate)
* [Google Kubernetes Engine (GKE)](#google-kubernetes-engine-gke)
* [Docker](#docker)
* [Roll Your Own](#roll-your-own)

### Kubernetes Helm Chart

Atlantis tiene un [Helm chart oficial](https://github.com/runatlantis/helm-charts/tree/main/charts/atlantis)

Para instalar:

1. Agrega el repositorio del helm chart de runatlantis a helm

    ```bash
    helm repo add runatlantis https://runatlantis.github.io/helm-charts
    ```

1. Haz `cd` dentro de un directorio donde vas a configurar tu Helm chart de Atlantis
1. Crea un archivo `values.yaml` ejecutando

    ```bash
    helm inspect values runatlantis/atlantis > values.yaml
    ```

1. Edita `values.yaml` y agrega tus credenciales de acceso y secreto de webhook

    ```yaml
    # for example
    github:
      user: foo
      token: bar
      secret: baz
    ```

1. Edita `values.yaml` y configura tu `orgAllowlist` (consulta [Repo Allowlist](server-configuration.md#repo-allowlist) para más información)

    ```yaml
    orgAllowlist: github.com/runatlantis/*
    ```

    **Nota**: Para la versión del helm chart < `4.0.2`, se debe usar `orgWhitelist` en su lugar.
1. Configura cualquier otra variable (consulta [Atlantis Helm Chart: Customization](https://github.com/runatlantis/helm-charts#customization)
    para la documentación)
1. Ejecuta

    ```sh
    helm install atlantis runatlantis/atlantis -f values.yaml
    ```

    Si estás usando helm v2, ejecuta:

    ```sh
    helm install -f values.yaml runatlantis/atlantis
    ```

¡Atlantis debería estar en funcionamiento en minutos! Consulta [Next Steps](#next-steps) para
qué hacer después.

### Kubernetes Manifests

Si quieres usar un manifiesto de Kubernetes sin procesar, ofrecemos ya sea un
[Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
o un [Statefulset](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/) con almacenamiento persistente.

Se recomienda StatefulSet porque Atlantis almacena sus datos en disco y así, si tu Pod muere
o actualizas Atlantis, no perderás planes que no han sido aplicados. Si
pierdes esos datos, solo necesitas ejecutar `atlantis plan` de nuevo, así que no es el fin del mundo.

Independientemente de si eliges un Deployment o StatefulSet, primero crea un Secret con el secreto de webhook y el token de acceso:

```bash
echo -n "yourtoken" > token
echo -n "yoursecret" > webhook-secret
kubectl create secret generic atlantis-vcs --from-file=token --from-file=webhook-secret
```

A continuación, edita los manifiestos siguientes de la siguiente manera:

1. Reemplaza `<VERSION>` en `image: ghcr.io/runatlantis/atlantis:<VERSION>` con la versión más reciente de [GitHub: Atlantis latest release](https://github.com/runatlantis/atlantis/releases/latest).
    * NOTA: Nunca querrás ejecutar con `:latest` porque si tu Pod se mueve a un nodo nuevo, Kubernetes descargará la imagen más reciente y podrías terminar
actualizando Atlantis por accidente.
2. Reemplaza `value: github.com/yourorg/*` bajo `name: ATLANTIS_REPO_ALLOWLIST` con el patrón de allowlist
para tus repos de Terraform. Consulta [--repo-allowlist](server-configuration.md#repo-allowlist) para más detalles.
3. Si estás usando GitHub:
    1. Reemplaza `<YOUR_GITHUB_USER>` con el nombre de usuario de tu usuario GitHub de Atlantis sin el `@`.
    2. Elimina todas las variables de entorno `ATLANTIS_GITLAB_*`, `ATLANTIS_GITEA_*`, `ATLANTIS_BITBUCKET_*` e `ATLANTIS_AZUREDEVOPS_*`.
4. Si estás usando GitLab:
    1. Reemplaza `<YOUR_GITLAB_USER>` con el nombre de usuario de tu usuario GitLab de Atlantis sin el `@`.
    2. Elimina todas las variables de entorno `ATLANTIS_GH_*`, `ATLANTIS_GITEA_*`, `ATLANTIS_BITBUCKET_*` e `ATLANTIS_AZUREDEVOPS_*`.
5. Si estás usando Gitea:
    1. Reemplaza `<YOUR_GITEA_USER>` con el nombre de usuario de tu usuario Gitea de Atlantis sin el `@`.
    2. Elimina todas las variables de entorno `ATLANTIS_GH_*`, `ATLANTIS_GITLAB_*`, `ATLANTIS_BITBUCKET_*` e `ATLANTIS_AZUREDEVOPS_*`.
6. Si estás usando Bitbucket:
    1. Reemplaza `<YOUR_BITBUCKET_USER>` con el nombre de usuario de tu usuario Bitbucket de Atlantis sin el `@`.
    2. Elimina todas las variables de entorno `ATLANTIS_GH_*`, `ATLANTIS_GITLAB_*`, `ATLANTIS_GITEA_*` e `ATLANTIS_AZUREDEVOPS_*`.
7. Si estás usando Azure DevOps:
    1. Reemplaza `<YOUR_AZUREDEVOPS_USER>` con el nombre de usuario de tu usuario Azure DevOps de Atlantis sin el `@`.
    2. Elimina todas las variables de entorno `ATLANTIS_GH_*`, `ATLANTIS_GITLAB_*`, `ATLANTIS_GITEA_*` e `ATLANTIS_BITBUCKET_*`.

#### Manifest de StatefulSet

<details>
 <summary>Mostrar...</summary>

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: atlantis
spec:
  serviceName: atlantis
  replicas: 1
  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      partition: 0
  selector:
    matchLabels:
      app.kubernetes.io/name: atlantis
  template:
    metadata:
      labels:
        app.kubernetes.io/name: atlantis
    spec:
      securityContext:
        fsGroup: 1000 # Atlantis group (1000) read/write access to volumes.
      containers:
      - name: atlantis
        image: ghcr.io/runatlantis/atlantis:v<VERSION> # 1. Replace <VERSION> with the most recent release.
        env:
        - name: ATLANTIS_REPO_ALLOWLIST
          value: github.com/yourorg/* # 2. Replace this with your own repo allowlist.

        ### GitHub Config ###
        - name: ATLANTIS_GH_USER
          value: <YOUR_GITHUB_USER> # 3i. If you're using GitHub replace <YOUR_GITHUB_USER> with the username of your Atlantis GitHub user without the `@`.
        - name: ATLANTIS_GH_TOKEN
          valueFrom:
            secretKeyRef:
              name: atlantis-vcs
              key: token
        - name: ATLANTIS_GH_WEBHOOK_SECRET
          valueFrom:
            secretKeyRef:
              name: atlantis-vcs
              key: webhook-secret
        ### End GitHub Config ###

        ### GitLab Config ###
        - name: ATLANTIS_GITLAB_USER
          value: <YOUR_GITLAB_USER> # 4i. If you're using GitLab replace <YOUR_GITLAB_USER> with the username of your Atlantis GitLab user without the `@`.
        - name: ATLANTIS_GITLAB_TOKEN
          valueFrom:
            secretKeyRef:
              name: atlantis-vcs
              key: token
        - name: ATLANTIS_GITLAB_WEBHOOK_SECRET
          valueFrom:
            secretKeyRef:
              name: atlantis-vcs
              key: webhook-secret
        ### End GitLab Config ###

        ### Gitea Config ###
        - name: ATLANTIS_GITEA_USER
          value: <YOUR_GITEA_USER> # 4i. If you're using Gitea replace <YOUR_GITEA_USER> with the username of your Atlantis Gitea user without the `@`.
        - name: ATLANTIS_GITEA_TOKEN
          valueFrom:
            secretKeyRef:
              name: atlantis-vcs
              key: token
        - name: ATLANTIS_GITEA_WEBHOOK_SECRET
          valueFrom:
            secretKeyRef:
              name: atlantis-vcs
              key: webhook-secret
        ### End Gitea Config ###

        ### Bitbucket Config ###
        - name: ATLANTIS_BITBUCKET_USER
          value: <YOUR_BITBUCKET_USER> # 5i. If you're using Bitbucket replace <YOUR_BITBUCKET_USER> with the username of your Atlantis Bitbucket user without the `@`.
        - name: ATLANTIS_BITBUCKET_TOKEN
          valueFrom:
            secretKeyRef:
              name: atlantis-vcs
              key: token
        - name: ATLANTIS_BITBUCKET_WEBHOOK_SECRET
          valueFrom:
            secretKeyRef:
              name: atlantis-vcs
              key: webhook-secret
        ### End Bitbucket Config ###

        ### Azure DevOps Config ###
        - name: ATLANTIS_AZUREDEVOPS_USER
          value: <YOUR_AZUREDEVOPS_USER> # 6i. If you're using Azure DevOps replace <YOUR_AZUREDEVOPS_USER> with the username of your Atlantis Azure DevOps user without the `@`.
        - name: ATLANTIS_AZUREDEVOPS_TOKEN
          valueFrom:
            secretKeyRef:
              name: atlantis-vcs
              key: token
        - name: ATLANTIS_AZUREDEVOPS_WEBHOOK_USER
          valueFrom:
            secretKeyRef:
              name: atlantis-vcs
              key: basic-user
        - name: ATLANTIS_AZUREDEVOPS_WEBHOOK_PASSWORD
          valueFrom:
            secretKeyRef:
              name: atlantis-vcs
              key: basic-password
        ### End Azure DevOps Config ###

        - name: ATLANTIS_DATA_DIR
          value: /atlantis
        - name: ATLANTIS_PORT
          value: "4141" # Kubernetes sets an ATLANTIS_PORT variable so we need to override.
        volumeMounts:
        - name: atlantis-data
          mountPath: /atlantis
        ports:
        - name: atlantis
          containerPort: 4141
        resources:
          requests:
            memory: 256Mi
            cpu: 100m
          limits:
            memory: 256Mi
            cpu: 100m
        livenessProbe:
          # We only need to check every 60s since Atlantis is not a
          # high-throughput service.
          periodSeconds: 60
          httpGet:
            path: /healthz
            port: 4141
            # If using https, change this to HTTPS
            scheme: HTTP
        readinessProbe:
          periodSeconds: 60
          httpGet:
            path: /healthz
            port: 4141
            # If using https, change this to HTTPS
            scheme: HTTP
  volumeClaimTemplates:
  - metadata:
      name: atlantis-data
    spec:
      accessModes: ["ReadWriteOnce"] # Volume should not be shared by multiple nodes.
      resources:
        requests:
          # The biggest thing Atlantis stores is the Git repo when it checks it out.
          # It deletes the repo after the pull request is merged.
          storage: 5Gi
---
apiVersion: v1
kind: Service
metadata:
  name: atlantis
spec:
  type: ClusterIP
  ports:
  - name: atlantis
    port: 80
    targetPort: 4141
  selector:
    app.kubernetes.io/name: atlantis
```

</details>

#### Manifest de Deployment

<details>
 <summary>Mostrar...</summary>

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: atlantis
  labels:
    app.kubernetes.io/name: atlantis
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: atlantis
  template:
    metadata:
      labels:
        app.kubernetes.io/name: atlantis
    spec:
      containers:
      - name: atlantis
        image: ghcr.io/runatlantis/atlantis:v<VERSION> # 1. Replace <VERSION> with the most recent release.
        env:
        - name: ATLANTIS_REPO_ALLOWLIST
          value: github.com/yourorg/* # 2. Replace this with your own repo allowlist.

        ### GitHub Config ###
        - name: ATLANTIS_GH_USER
          value: <YOUR_GITHUB_USER> # 3i. If you're using GitHub replace <YOUR_GITHUB_USER> with the username of your Atlantis GitHub user without the `@`.
        - name: ATLANTIS_GH_TOKEN
          valueFrom:
            secretKeyRef:
              name: atlantis-vcs
              key: token
        - name: ATLANTIS_GH_WEBHOOK_SECRET
          valueFrom:
            secretKeyRef:
              name: atlantis-vcs
              key: webhook-secret
        ### End GitHub Config ###

        ### GitLab Config ###
        - name: ATLANTIS_GITLAB_USER
          value: <YOUR_GITLAB_USER> # 4i. If you're using GitLab replace <YOUR_GITLAB_USER> with the username of your Atlantis GitLab user without the `@`.
        - name: ATLANTIS_GITLAB_TOKEN
          valueFrom:
            secretKeyRef:
              name: atlantis-vcs
              key: token
        - name: ATLANTIS_GITLAB_WEBHOOK_SECRET
          valueFrom:
            secretKeyRef:
              name: atlantis-vcs
              key: webhook-secret
        ### End GitLab Config ###

        ### Gitea Config ###
        - name: ATLANTIS_GITEA_USER
          value: <YOUR_GITEA_USER> # 4i. If you're using Gitea replace <YOUR_GITEA_USER> with the username of your Atlantis Gitea user without the `@`.
        - name: ATLANTIS_GITEA_TOKEN
          valueFrom:
            secretKeyRef:
              name: atlantis-vcs
              key: token
        - name: ATLANTIS_GITEA_WEBHOOK_SECRET
          valueFrom:
            secretKeyRef:
              name: atlantis-vcs
              key: webhook-secret
        ### End Gitea Config ###

        ### Bitbucket Config ###
        - name: ATLANTIS_BITBUCKET_USER
          value: <YOUR_BITBUCKET_USER> # 5i. If you're using Bitbucket replace <YOUR_BITBUCKET_USER> with the username of your Atlantis Bitbucket user without the `@`.
        - name: ATLANTIS_BITBUCKET_TOKEN
          valueFrom:
            secretKeyRef:
              name: atlantis-vcs
              key: token
        ### End Bitbucket Config ###

        ### Azure DevOps Config ###
        - name: ATLANTIS_AZUREDEVOPS_USER
          value: <YOUR_AZUREDEVOPS_USER> # 6i. If you're using Azure DevOps replace <YOUR_AZUREDEVOPS_USER> with the username of your Atlantis Azure DevOps user without the `@`.
        - name: ATLANTIS_AZUREDEVOPS_TOKEN
          valueFrom:
            secretKeyRef:
              name: atlantis-vcs
              key: token
        - name: ATLANTIS_AZUREDEVOPS_WEBHOOK_USER
          valueFrom:
            secretKeyRef:
              name: atlantis-vcs
              key: basic-user
        - name: ATLANTIS_AZUREDEVOPS_WEBHOOK_PASSWORD
          valueFrom:
            secretKeyRef:
              name: atlantis-vcs
              key: basic-password
        ### End Azure DevOps Config ###

        - name: ATLANTIS_PORT
          value: "4141" # Kubernetes sets an ATLANTIS_PORT variable so we need to override.
        ports:
        - name: atlantis
          containerPort: 4141
        resources:
          requests:
            memory: 256Mi
            cpu: 100m
          limits:
            memory: 256Mi
            cpu: 100m
        livenessProbe:
          # We only need to check every 60s since Atlantis is not a
          # high-throughput service.
          periodSeconds: 60
          httpGet:
            path: /healthz
            port: 4141
            # If using https, change this to HTTPS
            scheme: HTTP
        readinessProbe:
          periodSeconds: 60
          httpGet:
            path: /healthz
            port: 4141
            # If using https, change this to HTTPS
            scheme: HTTP
---
apiVersion: v1
kind: Service
metadata:
  name: atlantis
spec:
  type: ClusterIP
  ports:
  - name: atlantis
    port: 80
    targetPort: 4141
  selector:
    app.kubernetes.io/name: atlantis
```

</details>

#### Enrutamiento y SSL

Los manifiestos anteriores crean un `Service` de Kubernetes de tipo `type: ClusterIP` que no es accesible fuera de tu clúster.
Dependiendo de cómo estés haciendo el enrutamiento hacia Kubernetes, puede que quieras usar un Service de tipo `type: LoadBalancer` para que Atlantis sea accesible
para GitHub/GitLab y tus usuarios internos.

Si quieres agregar SSL puedes usar algo como [cert-manager](https://github.com/cert-manager/cert-manager) para generar certificados SSL
y montarlos en el Pod. Luego configura las variables de entorno `ATLANTIS_SSL_CERT_FILE` e `ATLANTIS_SSL_KEY_FILE` para habilitar SSL.
También podrías configurar SSL en tu LoadBalancer.

**¡Ya terminaste! Consulta [Next Steps](#next-steps) para qué hacer después.**

### Kubernetes Kustomize

Se proporciona un archivo `kustomization.yaml` en el directorio `kustomize/`, por lo que puedes usar este repositorio como una base remota para desplegar Atlantis con Kustomize.

Necesitarás proporcionar un secreto (con el nombre predeterminado de `atlantis-vcs`) para configurar Atlantis con credenciales de acceso para tus repositorios remotos.

Ejemplo:

```yaml
bases:
- github.com/runatlantis/atlantis//kustomize

resources:
- secrets.yaml
```

**Importante:** Debes asegurarte de aplicar parches a los manifiestos proporcionados con las variables de entorno correctas para tu instalación. Puedes crear parches inline desde tu archivo `kustomization.yaml` como se muestra a continuación:

```yaml
patchesStrategicMerge:
- |-
  apiVersion: apps/v1
  kind: StatefulSet
  metadata:
    name: atlantis
  spec:
    template:
      spec:
        ...
```

#### Requerido

```yaml
...
 containers:
  - name: atlantis
    env:
      - name: ATLANTIS_REPO_ALLOWLIST
        value: github.com/yourorg/* # 2. Replace this with your own repo allowlist.
```

#### GitLab

```yaml
...
containers:
- name: atlantis
  env:
    - name: ATLANTIS_GITLAB_USER
      value: <YOUR_GITLAB_USER> # 4i. If you're using GitLab replace <YOUR_GITLAB_USER> with the username of your Atlantis GitLab user without the `@`.
    - name: ATLANTIS_GITLAB_TOKEN
      valueFrom:
        secretKeyRef:
          name: atlantis-vcs
          key: token
    - name: ATLANTIS_GITLAB_WEBHOOK_SECRET
      valueFrom:
        secretKeyRef:
          name: atlantis-vcs
          key: webhook-secret
```

#### Gitea

```yaml
containers:
- name: atlantis
  env:
    - name: ATLANTIS_GITEA_USER
      value: <YOUR_GITEA_USER> # 4i. If you're using Gitea replace <YOUR_GITEA_USER> with the username of your Atlantis Gitea user without the `@`.
    - name: ATLANTIS_GITEA_TOKEN
      valueFrom:
        secretKeyRef:
          name: atlantis-vcs
          key: token
    - name: ATLANTIS_GITEA_WEBHOOK_SECRET
      valueFrom:
        secretKeyRef:
          name: atlantis-vcs
          key: webhook-secret
```

#### GitHub

```yaml
...
containers:
- name: atlantis
  env:
    - name: ATLANTIS_GH_USER
      value: <YOUR_GITHUB_USER> # 3i. If you're using GitHub replace <YOUR_GITHUB_USER> with the username of your Atlantis GitHub user without the `@`.
    - name: ATLANTIS_GH_TOKEN
      valueFrom:
        secretKeyRef:
          name: atlantis-vcs
          key: token
    - name: ATLANTIS_GH_WEBHOOK_SECRET
      valueFrom:
        secretKeyRef:
          name: atlantis-vcs
          key: webhook-secret
```

#### BitBucket

```yaml
...
containers:
- name: atlantis
  env:
    - name: ATLANTIS_BITBUCKET_USER
      value: <YOUR_BITBUCKET_USER> # 5i. If you're using Bitbucket replace <YOUR_BITBUCKET_USER> with the username of your Atlantis Bitbucket user without the `@`.
    - name: ATLANTIS_BITBUCKET_TOKEN
      valueFrom:
        secretKeyRef:
          name: atlantis-vcs
          key: token
```

### OpenShift

El Helm chart y los manifiestos de Kubernetes anteriores son compatibles con OpenShift, sin embargo necesitas ejecutar
con una variable de entorno adicional: `HOME=/home/atlantis`. Esto es necesario porque
OpenShift ejecuta imágenes Docker con id de usuario aleatorios que usan `/` como su directorio home.

### AWS Fargate

Si quieres ejecutar Atlantis en [AWS Fargate](https://aws.amazon.com/fargate/)
 revisa el módulo de Atlantis en el [Terraform Module Registry](https://registry.terraform.io/modules/terraform-aws-modules/atlantis/aws/latest)
 y luego revisa [Next Steps](#next-steps).

### Google Kubernetes Engine (GKE)

Puedes ejecutar Atlantis en GKE usando el [Helm chart](#kubernetes-helm-chart) o los [manifests](#kubernetes-manifests).

También hay un conjunto completo de configuraciones de Terraform que crean un clúster GKE,
Cloud Storage Backend y certificados TLS: [sethvargo atlantis-on-gke](https://github.com/sethvargo/atlantis-on-gke).

Una vez que hayas terminado, consulta [Next Steps](#next-steps).

### Google Compute Engine (GCE)

Atlantis se puede ejecutar en Google Compute Engine usando un módulo de Terraform que lo despliega como un contenedor Docker en una instancia administrada de Compute Engine.

Este [módulo de Terraform](https://registry.terraform.io/modules/runatlantis/atlantis/gce/latest) incluye la creación de un balanceador de carga de Cloud, una VM basada en Container-Optimized OS, un disco de datos persistente y un grupo de instancias administradas.

Después de que esté desplegado, consulta [Next Steps](#next-steps).

### Docker

Atlantis tiene una imagen Docker [oficial](https://ghcr.io/runatlantis/atlantis): `ghcr.io/runatlantis/atlantis`.

#### Customization

Si necesitas modificar la imagen Docker que proporcionamos, por ejemplo para agregar el binario de terragrunt, puedes hacer algo como esto:

1. Crea un archivo docker personalizado

    ```dockerfile
    FROM ghcr.io/runatlantis/atlantis:{latest version}

    # copy a terraform binary of the version you need
    USER root
    COPY terragrunt /usr/local/bin/terragrunt
    USER atlantis
    ```

A partir de la versión 0.26.0, la imagen de Atlantis ha sido actualizada para ejecutarse bajo el usuario atlantis, reemplazando la configuración anterior del usuario root. Este cambio requiere ajustes en definiciones de contenedores y scripts existentes para acomodar la nueva configuración de usuario. En escenarios donde se requieren paquetes adicionales de otras imágenes, los usuarios pueden cambiar temporalmente al usuario root insertando USER root en el Dockerfile. Después de la instalación de los paquetes necesarios, es recomendable volver al usuario atlantis para iniciar el servicio de Atlantis.
Además, el directorio /docker-entrypoint.d/ ofrece una opción flexible para introducir scripts extra que se ejecutarán antes del inicio del servidor Atlantis. Esta funcionalidad es particularmente beneficiosa para los usuarios que buscan personalizar su instancia de Atlantis sin la necesidad de desarrollar un pipeline dedicado.
**Aviso importante**: Hay una actualización crítica con respecto al directorio de datos en Atlantis. En versiones anteriores a 0.26.0, el directorio estaba configurado para ser accesible por el usuario root. Sin embargo, con la transición al usuario atlantis en versiones más nuevas, es imperativo actualizar los permisos del directorio en tu despliegue actual al actualizar a una versión posterior a 0.26.0. Este paso asegura acceso y funcionalidad sin problemas para el usuario atlantis.

1. Construye tu imagen Docker

    ```bash
    docker build -t {YOUR_DOCKER_ORG}/atlantis-custom .
    ```

1. Ejecuta tu imagen

    ```bash
    docker run {YOUR_DOCKER_ORG}/atlantis-custom server --gh-user=GITHUB_USERNAME --gh-token=GITHUB_TOKEN
    ```

### Microsoft Azure

El [Kubernetes Helm Chart](#kubernetes-helm-chart) estándar debería funcionar bien en [Azure Kubernetes Service](https://docs.microsoft.com/en-us/azure/aks/intro-kubernetes).

Otra opción es [Azure Container Instances](https://docs.microsoft.com/en-us/azure/container-instances/). Consulta el [repo](https://github.com/jplane/atlantis-on-aci) de este miembro de la comunidad o el [módulo de Terraform](https://github.com/getindata/terraform-azurerm-atlantis) nuevo y más actualizado para scripts de instalación y más información sobre ejecutar Atlantis en ACI.

**Nota sobre el despliegue en ACI:** Debido a un bug en versiones anteriores de Docker, se requiere Docker v23.0.0 o posterior para un despliegue sencillo. Alternativamente, la imagen Docker de Atlantis se puede enviar a un registro privado como ACR y luego usarse.

### Roll Your Own

Si quieres hacer tu propia instalación de Atlantis, puedes obtener el binario `atlantis`
de [GitHub](https://github.com/runatlantis/atlantis/releases)
o usar la [imagen Docker oficial](https://ghcr.io/runatlantis/atlantis).

#### Comando de inicio

Los flags exactos para `atlantis server` dependen de tu host Git:

##### GitHub

```bash
atlantis server \
--atlantis-url="$URL" \
--gh-user="$USERNAME" \
--gh-token="$TOKEN" \
--gh-webhook-secret="$SECRET" \
--repo-allowlist="$REPO_ALLOWLIST"
```

##### GitHub Enterprise

```bash
HOSTNAME=YOUR_GITHUB_ENTERPRISE_HOSTNAME # ex. github.runatlantis.io or tenant.ghe.com
atlantis server \
--atlantis-url="$URL" \
--gh-user="$USERNAME" \
--gh-token="$TOKEN" \
--gh-webhook-secret="$SECRET" \
--gh-hostname="$HOSTNAME" \
--repo-allowlist="$REPO_ALLOWLIST"
```

Para GitHub Enterprise Cloud, configura `--gh-hostname` con el hostname del tenant, como `tenant.ghe.com`, sin `https://` ni un prefijo `api.`.

##### GitLab

```bash
atlantis server \
--atlantis-url="$URL" \
--gitlab-user="$USERNAME" \
--gitlab-token="$TOKEN" \
--gitlab-webhook-secret="$SECRET" \
--repo-allowlist="$REPO_ALLOWLIST"
```

##### GitLab Enterprise

```bash
HOSTNAME=YOUR_GITLAB_ENTERPRISE_HOSTNAME # ex. gitlab.runatlantis.io
atlantis server \
--atlantis-url="$URL" \
--gitlab-user="$USERNAME" \
--gitlab-token="$TOKEN" \
--gitlab-webhook-secret="$SECRET" \
--gitlab-hostname="$HOSTNAME" \
--repo-allowlist="$REPO_ALLOWLIST"
```

##### Gitea

```bash
GITEA_BASE_URL=YOUR_GITEA_BASE_URL # ex. https://gitea.example.com:3000
atlantis server \
--atlantis-url="$URL" \
--gitea-user="$USERNAME" \
--gitea-token="$TOKEN" \
--gitea-base-url="$GITEA_BASE_URL" \
--gitea-webhook-secret="$SECRET" \
--gitea-page-size=30 \
--repo-allowlist="$REPO_ALLOWLIST"
```

##### Bitbucket Cloud (bitbucket.org)

```bash
atlantis server \
--atlantis-url="$URL" \
--bitbucket-user="$USERNAME" \
--bitbucket-token="$TOKEN" \
--bitbucket-webhook-secret="$SECRET" \
--repo-allowlist="$REPO_ALLOWLIST"
```

##### Bitbucket Server (aka Stash)

```bash
BASE_URL=YOUR_BITBUCKET_SERVER_URL # ex. http://bitbucket.mycorp:7990
atlantis server \
--atlantis-url="$URL" \
--bitbucket-user="$USERNAME" \
--bitbucket-token="$TOKEN" \
--bitbucket-webhook-secret="$SECRET" \
--bitbucket-base-url="$BASE_URL" \
--repo-allowlist="$REPO_ALLOWLIST"
```

##### Azure DevOps

Se requiere un certificado y una clave privada si se usa autenticación Basic para webhooks.

```bash
atlantis server \
--atlantis-url="$URL" \
--azuredevops-user="$USERNAME" \
--azuredevops-token="$TOKEN" \
--azuredevops-webhook-user="$ATLANTIS_AZUREDEVOPS_WEBHOOK_USER" \
--azuredevops-webhook-password="$ATLANTIS_AZUREDEVOPS_WEBHOOK_PASSWORD" \
--repo-allowlist="$REPO_ALLOWLIST"
--ssl-cert-file=file.crt
--ssl-key-file=file.key
```

Donde

* `$URL` es la URL en la que se puede acceder a Atlantis
* `$USERNAME` es el nombre de usuario de GitHub/GitLab/Gitea/Bitbucket/AzureDevops para el que generaste el token
* `$TOKEN` es el token de acceso que creaste. Si no quieres que esto se pase
  como un argumento por razones de seguridad, puedes especificarlo en un archivo de configuración
   (consulta [Configuration](server-configuration.md#environment-variables))
    o como una variable de entorno: `ATLANTIS_GH_TOKEN` o `ATLANTIS_GITLAB_TOKEN` o `ATLANTIS_GITEA_TOKEN`
     o `ATLANTIS_BITBUCKET_TOKEN` o `ATLANTIS_AZUREDEVOPS_TOKEN`
* `$SECRET` es la clave aleatoria que usaste para el secreto de webhook.
   Si no quieres que esto se pase como un argumento por razones de seguridad,
    puedes especificarlo en un archivo de configuración
     (consulta [Configuration](server-configuration.md#environment-variables))
      o como una variable de entorno: `ATLANTIS_GH_WEBHOOK_SECRET` o `ATLANTIS_GITLAB_WEBHOOK_SECRET` o
  `ATLANTIS_GITEA_WEBHOOK_SECRET`
* `$REPO_ALLOWLIST` es en qué repos puede ejecutarse Atlantis, por ej.
 `github.com/runatlantis/*` o `github.enterprise.corp.com/*`.
  Consulta [--repo-allowlist](server-configuration.md#repo-allowlist) para más detalles.

¡Atlantis ahora está ejecutándose!
::: tip
Recomendamos ejecutarlo bajo algo como Systemd o Supervisord que lo
reiniciará en caso de fallo.
:::

## Next Steps

* Para asegurar que Atlantis se está ejecutando, carga su UI. De forma predeterminada Atlantis se ejecuta en el puerto `4141`.
* Ahora estás listo para agregar Webhooks a tus repos. Consulta [Configuring Webhooks](configuring-webhooks.md).
