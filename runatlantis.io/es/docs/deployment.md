# Despliegue

Esta página cubre poner Atlantis en funcionamiento en su infraestructura.

::: tip Prerrequisitos

* Ha creado [credenciales de acceso](access-credentials.md) para su usuario de Atlantis
* Ha creado un [secreto de webhook](webhook-secrets.md)
:::

## Resumen de arquitectura

### Tiempo de ejecución

Atlantis es una aplicación simple de [Go](https://golang.org/). Recibe webhooks de
su host de Git y ejecuta comandos de Terraform localmente. Hay una imagen oficial de
Atlantis [Docker image](https://ghcr.io/runatlantis/atlantis).

### Enrutamiento

Atlantis y su host de Git necesitan poder enrutar y comunicarse entre sí. Su host de Git necesita poder enviar webhooks a Atlantis y Atlantis necesita poder hacer llamadas a la API a su host de Git.
Si está usando
un host de Git público como github.com, gitlab.com, gitea.com, bitbucket.org o dev.azure.com, entonces necesitará
exponer Atlantis a internet.

Si está usando un host de Git privado como GitHub Enterprise, GitLab Enterprise, Gitea autoalojado o
Bitbucket Server, entonces Atlantis necesita ser enrutable desde el host privado y Atlantis necesitará poder enrutar al host privado.

### Datos

Atlantis no tiene base de datos externa. Atlantis almacena archivos de plan de Terraform en disco.
Si Atlantis pierde esos datos entre un ciclo de `plan` y `apply`, entonces los usuarios tendrán que
volver a ejecutar `plan`. Debido a esto, puede querer aprovisionar un disco persistente
para Atlantis.

## Despliegue

Elija su tipo de despliegue:

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

1. Agregue el repositorio del helm chart de runatlantis a helm

    ```bash
    helm repo add runatlantis https://runatlantis.github.io/helm-charts
    ```

1. `cd` en un directorio donde va a configurar su Atlantis Helm chart
1. Cree un archivo `values.yaml` ejecutando

    ```bash
    helm inspect values runatlantis/atlantis > values.yaml
    ```

1. Edite `values.yaml` y agregue sus credenciales de acceso y secreto de webhook

    ```yaml
    # for example
    github:
      user: foo
      token: bar
      secret: baz
    ```

1. Edite `values.yaml` y configure su `orgAllowlist` (vea [Repo Allowlist](server-configuration.md#repo-allowlist) para más información)

    ```yaml
    orgAllowlist: github.com/runatlantis/*
    ```

    **Nota**: Para la versión del helm chart < `4.0.2`, se debe usar `orgWhitelist` en su lugar.
1. Configure cualquier otra variable (vea [Atlantis Helm Chart: Customization](https://github.com/runatlantis/helm-charts#customization)
    para la documentación)
1. Ejecute

    ```sh
    helm install atlantis runatlantis/atlantis -f values.yaml
    ```

    Si está usando helm v2, ejecute:

    ```sh
    helm install -f values.yaml runatlantis/atlantis
    ```

¡Atlantis debería estar en funcionamiento en minutos! Vea [Next Steps](#next-steps) para
qué hacer a continuación.

### Kubernetes Manifests

Si desea usar un manifest de Kubernetes sin procesar, ofrecemos ya sea un
[Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
o un [Statefulset](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/) con almacenamiento persistente.

Se recomienda StatefulSet porque Atlantis almacena sus datos en disco y así, si su Pod muere
o actualiza Atlantis, no perderá los plans que no se hayan aplicado. Si
aun así pierde esos datos, solo necesita ejecutar `atlantis plan` de nuevo, por lo que no es el fin del mundo.

Independientemente de si elige un Deployment o StatefulSet, primero cree un Secret con el secreto de webhook y el token de acceso:

```bash
echo -n "yourtoken" > token
echo -n "yoursecret" > webhook-secret
kubectl create secret generic atlantis-vcs --from-file=token --from-file=webhook-secret
```

A continuación, edite los manifests siguientes de la siguiente manera:

1. Reemplace `<VERSION>` en `image: ghcr.io/runatlantis/atlantis:<VERSION>` con la versión más reciente de [GitHub: Atlantis latest release](https://github.com/runatlantis/atlantis/releases/latest).
    * NOTA: Nunca querrá ejecutar con `:latest` porque si su Pod se mueve a un nuevo nodo, Kubernetes descargará la imagen más reciente y podría terminar
actualizando Atlantis por accidente.
2. Reemplace `value: github.com/yourorg/*` bajo `name: ATLANTIS_REPO_ALLOWLIST` con el patrón de allowlist
para sus repos de Terraform. Vea [--repo-allowlist](server-configuration.md#repo-allowlist) para más detalles.
3. Si está usando GitHub:
    1. Reemplace `<YOUR_GITHUB_USER>` con el nombre de usuario de su usuario Atlantis de GitHub sin el `@`.
    2. Elimine todas las variables de entorno `ATLANTIS_GITLAB_*`, `ATLANTIS_GITEA_*`, `ATLANTIS_BITBUCKET_*` y `ATLANTIS_AZUREDEVOPS_*`.
4. Si está usando GitLab:
    1. Reemplace `<YOUR_GITLAB_USER>` con el nombre de usuario de su usuario Atlantis de GitLab sin el `@`.
    2. Elimine todas las variables de entorno `ATLANTIS_GH_*`, `ATLANTIS_GITEA_*`, `ATLANTIS_BITBUCKET_*` y `ATLANTIS_AZUREDEVOPS_*`.
5. Si está usando Gitea:
    1. Reemplace `<YOUR_GITEA_USER>` con el nombre de usuario de su usuario Atlantis de Gitea sin el `@`.
    2. Elimine todas las variables de entorno `ATLANTIS_GH_*`, `ATLANTIS_GITLAB_*`, `ATLANTIS_BITBUCKET_*` y `ATLANTIS_AZUREDEVOPS_*`.
6. Si está usando Bitbucket:
    1. Reemplace `<YOUR_BITBUCKET_USER>` con el nombre de usuario de su usuario Atlantis de Bitbucket sin el `@`.
    2. Elimine todas las variables de entorno `ATLANTIS_GH_*`, `ATLANTIS_GITLAB_*`, `ATLANTIS_GITEA_*` y `ATLANTIS_AZUREDEVOPS_*`.
7. Si está usando Azure DevOps:
    1. Reemplace `<YOUR_AZUREDEVOPS_USER>` con el nombre de usuario de su usuario Atlantis de Azure DevOps sin el `@`.
    2. Elimine todas las variables de entorno `ATLANTIS_GH_*`, `ATLANTIS_GITLAB_*`, `ATLANTIS_GITEA_*` y `ATLANTIS_BITBUCKET_*`.

#### StatefulSet Manifest

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

#### Deployment Manifest

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

Los manifests anteriores crean un Kubernetes `Service` de `type: ClusterIP` que no es accesible fuera de su clúster.
Dependiendo de cómo esté haciendo el enrutamiento hacia Kubernetes, puede querer usar un Service de `type: LoadBalancer` para que Atlantis sea accesible
para GitHub/GitLab y sus usuarios internos.

Si desea agregar SSL, puede usar algo como [cert-manager](https://github.com/cert-manager/cert-manager) para generar certificados SSL
y montarlos en el Pod. Luego configure las variables de entorno `ATLANTIS_SSL_CERT_FILE` e `ATLANTIS_SSL_KEY_FILE` para habilitar SSL.
También podría configurar SSL en su LoadBalancer.

**¡Ha terminado! Vea [Next Steps](#next-steps) para qué hacer a continuación.**

### Kubernetes Kustomize

Se proporciona un archivo `kustomization.yaml` en el directorio `kustomize/`, por lo que puede usar este repositorio como una base remota para desplegar Atlantis con Kustomize.

Necesitará proporcionar un secret (con el nombre predeterminado de `atlantis-vcs`) para configurar Atlantis con credenciales de acceso para sus repositorios remotos.

Ejemplo:

```yaml
bases:
- github.com/runatlantis/atlantis//kustomize

resources:
- secrets.yaml
```

**Importante:** Debe asegurarse de aplicar parches a los manifests proporcionados con las variables de entorno correctas para su instalación. Puede crear parches inline desde su archivo `kustomization.yaml` como se muestra a continuación:

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

El Helm chart y los manifests de Kubernetes anteriores son compatibles con OpenShift; sin embargo, necesita ejecutar
con una variable de entorno adicional: `HOME=/home/atlantis`. Esto es necesario porque
OpenShift ejecuta imágenes de Docker con id de usuario aleatorios que usan `/` como su directorio home.

### AWS Fargate

Si desea ejecutar Atlantis en [AWS Fargate](https://aws.amazon.com/fargate/)
 revise el módulo de Atlantis en el [Terraform Module Registry](https://registry.terraform.io/modules/terraform-aws-modules/atlantis/aws/latest)
 y luego revise [Next Steps](#next-steps).

### Google Kubernetes Engine (GKE)

Puede ejecutar Atlantis en GKE usando el [Helm chart](#kubernetes-helm-chart) o los [manifests](#kubernetes-manifests).

También hay un conjunto completo de configuraciones de Terraform que crean un clúster de GKE,
Cloud Storage Backend y certificados TLS: [sethvargo atlantis-on-gke](https://github.com/sethvargo/atlantis-on-gke).

Una vez que haya terminado, vea [Next Steps](#next-steps).

### Google Compute Engine (GCE)

Atlantis puede ejecutarse en Google Compute Engine usando un módulo de Terraform que lo despliega como un contenedor de Docker en una instancia administrada de Compute Engine.

Este [módulo de Terraform](https://registry.terraform.io/modules/runatlantis/atlantis/gce/latest) incluye la creación de un balanceador de carga de Cloud, una VM basada en Container-Optimized OS, un disco de datos persistente y un grupo de instancias administradas.

Después de que esté desplegado, vea [Next Steps](#next-steps).

### Docker

Atlantis tiene una imagen oficial de Docker [official](https://ghcr.io/runatlantis/atlantis): `ghcr.io/runatlantis/atlantis`.

#### Variantes de imagen

Cada release se publica en cuatro variantes. La tag sin sufijo (por ejemplo `v0.47.1` o `latest`) es la imagen Alpine.

| Sufijo de tag     | Base   | Terraform y OpenTofu incluidos |
|----------------|--------|--------------------------------|
| `-alpine`      | Alpine | sí                            |
| `-debian`      | Debian | sí                            |
| `-alpine-slim` | Alpine | no                             |
| `-debian-slim` | Debian | no                             |

Las imágenes completas incluyen las últimas versiones menores de Terraform y la versión actual de OpenTofu, e `terraform` en `PATH` apunta a la más nueva de ellas.

Las imágenes slim se distribuyen sin ninguno de los dos binarios, por lo que los escáneres de vulnerabilidades no informan avisos sobre versiones de Terraform u OpenTofu que puede que ni siquiera use. Todo lo demás (`conftest`, `git-lfs`, `git`, `curl`, `dumb-init`) es igual que en la imagen completa. Atlantis descarga la versión de Terraform que necesita en el primer uso, por lo que a la imagen slim se le debe indicar cuál es esa versión:

* Configure [`--default-tf-version`](server-configuration.md#default-tf-version) como una flag, como `ATLANTIS_DEFAULT_TF_VERSION` o en el archivo de configuración del servidor. Sin ello, el servidor se niega a iniciar con `terraform not found in $PATH`. La imagen slim deliberadamente no establece ningún valor predeterminado propio, porque una variable de entorno incorporada en la imagen tendría precedencia sobre una versión fijada en su archivo de configuración.
* `terraform_version` por proyecto en `atlantis.yaml` e `--tf-download-url` funcionan como de costumbre.
* Para OpenTofu, configure `ATLANTIS_TF_DISTRIBUTION=opentofu` y dé una versión de OpenTofu como la predeterminada.
* Si las descargas salientes no están permitidas desde su host de Atlantis (`--tf-download=false`), monte o copie en la imagen los binarios que necesita en su lugar. Vea [Customization](#customization) a continuación.

#### Customization

Si necesita modificar la imagen de Docker que proporcionamos, por ejemplo para agregar el binario de terragrunt, puede hacer algo como esto:

1. Cree un archivo docker personalizado

    ```dockerfile
    FROM ghcr.io/runatlantis/atlantis:{latest version}

    # copy a terraform binary of the version you need
    USER root
    COPY terragrunt /usr/local/bin/terragrunt
    USER atlantis
    ```

A partir de la versión 0.26.0, la imagen de Atlantis ha sido actualizada para ejecutarse bajo el usuario atlantis, reemplazando la configuración anterior del usuario root. Este cambio requiere ajustes en las definiciones de contenedor y scripts existentes para adaptarse a la nueva configuración de usuario. En escenarios donde se requieren paquetes adicionales de otras imágenes, los usuarios pueden cambiar temporalmente al usuario root insertando USER root en el Dockerfile. Después de la instalación de los paquetes necesarios, es aconsejable volver al usuario atlantis para iniciar el servicio de Atlantis.
Adicionalmente, el directorio /docker-entrypoint.d/ ofrece una opción flexible para introducir scripts extra que se ejecutarán antes del lanzamiento del servidor Atlantis. Esta característica es particularmente beneficiosa para usuarios que buscan personalizar su instancia de Atlantis sin la necesidad de desarrollar un pipeline dedicado.
**Aviso importante**: Hay una actualización crítica con respecto al directorio de datos en Atlantis. En versiones anteriores a 0.26.0, el directorio estaba configurado para ser accesible por el usuario root. Sin embargo, con la transición al usuario atlantis en las versiones más nuevas, es imperativo actualizar los permisos del directorio en su despliegue actual al actualizar a una versión posterior a 0.26.0. Este paso asegura acceso y funcionalidad sin problemas para el usuario atlantis.

1. Construya su imagen de Docker

    ```bash
    docker build -t {YOUR_DOCKER_ORG}/atlantis-custom .
    ```

1. Ejecute su imagen

    ```bash
    docker run {YOUR_DOCKER_ORG}/atlantis-custom server --gh-user=GITHUB_USERNAME --gh-token=GITHUB_TOKEN
    ```

### Microsoft Azure

El [Kubernetes Helm Chart](#kubernetes-helm-chart) estándar debería funcionar bien en [Azure Kubernetes Service](https://docs.microsoft.com/en-us/azure/aks/intro-kubernetes).

Otra opción es [Azure Container Instances](https://docs.microsoft.com/en-us/azure/container-instances/). Vea el [repo](https://github.com/jplane/atlantis-on-aci) de este miembro de la comunidad o el [módulo de Terraform](https://github.com/getindata/terraform-azurerm-atlantis) nuevo y más actualizado para scripts de instalación y más información sobre ejecutar Atlantis en ACI.

**Nota sobre el despliegue en ACI:** Debido a un bug en releases anteriores de Docker, se requiere Docker v23.0.0 o posterior para un despliegue directo. Alternativamente, la imagen de Docker de Atlantis puede enviarse a un registry privado como ACR y luego usarse.

### Roll Your Own

Si quiere crear su propia instalación de Atlantis, puede obtener el binario `atlantis`
desde [GitHub](https://github.com/runatlantis/atlantis/releases)
o usar la [imagen oficial de Docker](https://ghcr.io/runatlantis/atlantis).

#### Comando de inicio

Las flags exactas para `atlantis server` dependen de su host de Git:

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

Para GitHub Enterprise Cloud, configure `--gh-hostname` con el hostname del tenant, como `tenant.ghe.com`, sin `https://` ni un prefijo `api.`.

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

* `$URL` es la URL en la que se puede alcanzar Atlantis
* `$USERNAME` es el nombre de usuario de GitHub/GitLab/Gitea/Bitbucket/AzureDevops para el que generó el token
* `$TOKEN` es el token de acceso que creó. Si no quiere que esto se pase
  como un argumento por razones de seguridad, puede especificarlo en un archivo de configuración
   (vea [Configuration](server-configuration.md#environment-variables))
    o como una variable de entorno: `ATLANTIS_GH_TOKEN` o `ATLANTIS_GITLAB_TOKEN` o `ATLANTIS_GITEA_TOKEN`
     o `ATLANTIS_BITBUCKET_TOKEN` o `ATLANTIS_AZUREDEVOPS_TOKEN`
* `$SECRET` es la clave aleatoria que usó para el secreto de webhook.
   Si no quiere que esto se pase como un argumento por razones de seguridad,
    puede especificarlo en un archivo de configuración
     (vea [Configuration](server-configuration.md#environment-variables))
      o como una variable de entorno: `ATLANTIS_GH_WEBHOOK_SECRET` o `ATLANTIS_GITLAB_WEBHOOK_SECRET` o
  `ATLANTIS_GITEA_WEBHOOK_SECRET`
* `$REPO_ALLOWLIST` es en qué repos puede ejecutarse Atlantis, p. ej.
 `github.com/runatlantis/*` o `github.enterprise.corp.com/*`.
  Vea [--repo-allowlist](server-configuration.md#repo-allowlist) para más detalles.

¡Atlantis ahora está en ejecución!
::: tip
Recomendamos ejecutarlo bajo algo como Systemd o Supervisord que lo
reiniciará en caso de fallo.
:::

## Próximos pasos

* Para asegurarse de que Atlantis está ejecutándose, cargue su UI. De forma predeterminada Atlantis se ejecuta en el puerto `4141`.
* Ahora está listo para agregar Webhooks a sus repos. Vea [Configuring Webhooks](configuring-webhooks.md).
