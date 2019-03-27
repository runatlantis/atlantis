# Deployment
This doc covers getting Atlantis up and running in your infrastructure.

::: tip Pre-Requisites
* You have created an [access credential](access-credentials.html)
* You have created a [webhook secret](webhook-secrets.html)
:::

[[toc]]

## Architecture Overview
### Runtime
Atlantis is a simple [Go](https://golang.org/) app. It receives webhooks from
your Git host and executes Terraform commands locally. There is an official
Atlantis [Docker image](https://hub.docker.com/r/runatlantis/atlantis/).

### Routing
Atlantis and your Git host need to be able to route and communicate with one another. Your Git host needs to be able to send webhooks to Atlantis and Atlantis needs to be able to make API calls to your Git host.
If you're using 
a public Git host like GitHub.com, GitLab.com or Bitbucket.org then you'll need to
expose Atlantis to the internet.

If you're using a private Git host like GitHub Enterprise, GitLab Enterprise or
Bitbucket Server, then you just need Atlantis to be routable to and from there.

### Data
Atlantis has no external database. Atlantis stores Terraform plan files on disk.
If Atlantis loses that data in between a `plan` and `apply` cycle, then users will have
to re-run `plan`. Because of this, you may want to provision a persistent disk
for Atlantis.

## Deployment

Pick your deployment type:
* [Kubernetes Helm Chart](#kubernetes-helm-chart)
* [Kubernetes Manifests](#kubernetes-manifests)
* [OpenShift](#openshift)
* [AWS Fargate](#aws-fargate)
* [Google Kubernetes Engine (GKE)](#google-kubernetes-engine-gke)
* [Docker](#docker)
* [Roll Your Own](#roll-your-own)


### Kubernetes Helm Chart
Atlantis has an [official Helm chart](https://hub.kubeapps.com/charts/stable/atlantis).

To install:
1. `cd` into a directory where you're going to configure your Atlantis Helm chart
1. Create a `values.yaml` file by running
    ```bash
    helm inspect values stable/atlantis > values.yaml
    ```
1. Edit `values.yaml` and add your access credentials and webhook secret
    ```yaml
    # for example
    github:
      user: foo
      token: bar
      secret: baz
    ```
1. Edit `values.yaml` and set your `orgWhitelist` (see [Repo Whitelist](server-configuration.html#repo-whitelist) for more information)
    ```yaml
    orgWhitelist: github.com/runatlantis/*
    ```
1. Configure any other variables (see [https://github.com/helm/charts/tree/master/stable/atlantis#customization](https://github.com/helm/charts/tree/master/stable/atlantis#customization)
    for documentation)
1. Run
    ```sh
    helm install -f values.yaml stable/atlantis
    ```


Atlantis should be up and running in minutes! See [Next Steps](#next-steps) for
what to do next.

### Kubernetes Manifests
If you'd like to use a raw Kubernetes manifest, we offer either a
[Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
or a [Statefulset](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/) with persistent storage.

StatefulSet is recommended because Atlantis stores its data on disk and so if your Pod dies
or you upgrade Atlantis, you won't lose plans that haven't been applied. If
you do lose that data, you just need to run `atlantis plan` again so it's not the end of the world.

Regardless of whether you choose a Deployment or StatefulSet, first create a Secret with the webhook secret and access token:
```bash
echo -n "yourtoken" > token
echo -n "yoursecret" > webhook-secret
kubectl create secret generic atlantis-vcs --from-file=token --from-file=webhook-secret
```
::: tip Note
If you're using Bitbucket Cloud then there is no webhook secret since it's not supported.
:::

Next, edit the manifests below as follows:
1. Replace `<VERSION>` in `image: runatlantis/atlantis:<VERSION>` with the most recent version from [https://github.com/runatlantis/atlantis/releases/latest](https://github.com/runatlantis/atlantis/releases/latest).
    * NOTE: You never want to run with `:latest` because if your Pod moves to a new node, Kubernetes will pull the latest image and you might end
up upgrading Atlantis by accident!
2. Replace `value: github.com/yourorg/*` under `name: ATLANTIS_REPO_WHITELIST` with the whitelist pattern
for your Terraform repos. See [Repo Whitelist](server-configuration.html#repo-whitelist) for more details.
3. If you're using GitHub:
    1. Replace `<YOUR_GITHUB_USER>` with the username of your Atlantis GitHub user without the `@`.
    2. Delete all the `ATLANTIS_GITLAB_*` and `ATLANTIS_BITBUCKET_*` environment variables.
4. If you're using GitLab:
    1. Replace `<YOUR_GITLAB_USER>` with the username of your Atlantis GitLab user without the `@`.
    2. Delete all the `ATLANTIS_GH_*` and `ATLANTIS_BITBUCKET_*` environment variables.
5. If you're using Bitbucket:
    1. Replace `<YOUR_BITBUCKET_USER>` with the username of your Atlantis Bitbucket user without the `@`.
    2. Delete all the `ATLANTIS_GH_*` and `ATLANTIS_GITLAB_*` environment variables.

#### StatefulSet Manifest
<details>
 <summary>Show...</summary>

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
      app: atlantis
  template:
    metadata:
      labels:
        app: atlantis
    spec:
      securityContext:
        fsGroup: 1000 # Atlantis group (1000) read/write access to volumes.
      containers:
      - name: atlantis
        image: runatlantis/atlantis:v<VERSION> # 1. Replace <VERSION> with the most recent release.
        env:
        - name: ATLANTIS_REPO_WHITELIST
          value: github.com/yourorg/* # 2. Replace this with your own repo whitelist.

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

        ### Bitbucket Config ###
        - name: ATLANTIS_BITBUCKET_USER
          value: <YOUR_BITBUCKET_USER> # 5i. If you're using Bitbucket replace <YOUR_BITBUCKET_USER> with the username of your Atlantis Bitbucket user without the `@`.
        - name: ATLANTIS_BITBUCKET_TOKEN
          valueFrom:
            secretKeyRef:
              name: atlantis-vcs
              key: token
        ### End Bitbucket Config ###

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
  ports:
  - name: atlantis
    port: 80
    targetPort: 4141
  selector:
    app: atlantis
```
</details>


#### Deployment Manifest
<details>
 <summary>Show...</summary>

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: atlantis
  labels:
    app: atlantis
spec:
  replicas: 1
  selector:
    matchLabels:
      app: atlantis
  template:
    metadata:
      labels:
        app: atlantis
    spec:
      containers:
      - name: atlantis
        image: runatlantis/atlantis:v<VERSION> # 1. Replace <VERSION> with the most recent release.
        env:
        - name: ATLANTIS_REPO_WHITELIST
          value: github.com/yourorg/* # 2. Replace this with your own repo whitelist.

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

        ### Bitbucket Config ###
        - name: ATLANTIS_BITBUCKET_USER
          value: <YOUR_BITBUCKET_USER> # 5i. If you're using Bitbucket replace <YOUR_BITBUCKET_USER> with the username of your Atlantis Bitbucket user without the `@`.
        - name: ATLANTIS_BITBUCKET_TOKEN
          valueFrom:
            secretKeyRef:
              name: atlantis-vcs
              key: token
        ### End Bitbucket Config ###

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
  ports:
  - name: atlantis
    port: 80
    targetPort: 4141
  selector:
    app: atlantis
```
</details>

#### Routing and SSL
The manifests above create a Kubernetes `Service` of type `ClusterIP` which isn't accessible outside your cluster.
Depending on how you're doing routing into Kubernetes, you may want to use a `LoadBalancer` so that Atlantis is accessible
to GitHub/GitLab and your internal users.

If you want to add SSL you can use something like [https://github.com/jetstack/cert-manager](https://github.com/jetstack/cert-manager) to generate SSL
certs and mount them into the Pod. Then set the `ATLANTIS_SSL_CERT_FILE` and `ATLANTIS_SSL_KEY_FILE` environment variables to enable SSL.
You could also set up SSL at your LoadBalancer.

**You're done! See [Next Steps](#next-steps) for what to do next.**

### OpenShift
The Helm chart and Kubernetes manifests above are compatible with OpenShift, however you need to run
with an additional environment variable: `ATLANTIS_DATA_DIR=/home/atlantis`. This is required because
OpenShift runs Docker images with random user id's that use `/` as their home directory.

### AWS Fargate
If you'd like to run Atlantis on [AWS Fargate](https://aws.amazon.com/fargate/)
 check out the Atlantis module on the Terraform Module Registry: [https://registry.terraform.io/modules/terraform-aws-modules/atlantis/aws](https://registry.terraform.io/modules/terraform-aws-modules/atlantis/aws)
 and then check out the [Next Steps](#next-steps).

### Google Kubernetes Engine (GKE)
You can run Atlantis on GKE using the [Helm chart](#kubernetes-helm-chart) or the [manifests](#kubernetes-manifests).

There is also a set of full Terraform configurations that create a GKE Cluster,
Cloud Storage Backend and TLS certs: [https://github.com/sethvargo/atlantis-on-gke](https://github.com/sethvargo/atlantis-on-gke).

Once you're done, see [Next Steps](#next-steps).

### Docker
Atlantis has an [official](https://hub.docker.com/r/runatlantis/atlantis/) Docker image: `runatlantis/atlantis`.

#### Customization
If you need to modify the Docker image that we provide, for instance to add the terragrunt binary, you can do something like this:

1. Create a custom docker file
    ```dockerfile
    FROM runatlantis/atlantis:{latest version}

    # copy a terraform binary of the version you need
    COPY terragrunt /usr/local/bin/terrgrunt
    ```

1. Build your Docker image
    ```bash
    docker build -t {YOUR_DOCKER_ORG}/atlantis-custom .
    ```

1. Run your image
    ```bash
    docker run {YOUR_DOCKER_ORG}/atlantis-custom server --gh-user=GITHUB_USERNAME --gh-token=GITHUB_TOKEN
    ```

### Roll Your Own
If you're deploying Atlantis into infrastructure not listed above, here's what
Atlantis needs in its environment.

#### Terraform
The `terraform` binary needs to be in the `$PATH` for Atlantis.
Download from https://www.terraform.io/downloads.html
```bash
unzip path/to/terraform_*.zip -d /usr/local/bin
```
Check that it's in your `$PATH`
```
$ terraform version
Terraform v0.10.0
```
If you want to use a different version of Terraform see [Terraform Versions](requirements.html#terraform-versions)

#### Atlantis Binary
Get the latest release from [https://github.com/runatlantis/atlantis/releases](https://github.com/runatlantis/atlantis/releases) and unpackage it.

#### Startup Command
The exact flags to `atlantis server` depends on your Git host:

##### GitHub
```bash
atlantis server \
--atlantis-url="$URL" \
--gh-user="$USERNAME" \
--gh-token="$TOKEN" \
--gh-webhook-secret="$SECRET" \
--repo-whitelist="$REPO_WHITELIST"
```

##### GitHub Enterprise
```bash
HOSTNAME=YOUR_GITHUB_ENTERPRISE_HOSTNAME # ex. github.runatlantis.io
atlantis server \
--atlantis-url="$URL" \
--gh-user="$USERNAME" \
--gh-token="$TOKEN" \
--gh-webhook-secret="$SECRET" \
--gh-hostname="$HOSTNAME" \
--repo-whitelist="$REPO_WHITELIST"
```

##### GitLab
```bash
atlantis server \
--atlantis-url="$URL" \
--gitlab-user="$USERNAME" \
--gitlab-token="$TOKEN" \
--gitlab-webhook-secret="$SECRET" \
--repo-whitelist="$REPO_WHITELIST"
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
--repo-whitelist="$REPO_WHITELIST"
```

##### Bitbucket Cloud (bitbucket.org)
```bash
atlantis server \
--atlantis-url="$URL" \
--bitbucket-user="$USERNAME" \
--bitbucket-token="$TOKEN" \
--repo-whitelist="$REPO_WHITELIST"
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
--repo-whitelist="$REPO_WHITELIST"
```

Where
- `$URL` is the URL that Atlantis can be reached at
- `$USERNAME` is the GitHub/GitLab/Bitbucket username you generated the token for
- `$TOKEN` is the access token you created. If you don't want this to be passed
  in as an argument for security reasons you can specify it in a config file
   (see [Configuration](/docs/server-configuration.html#environment-variables))
    or as an environment variable: `ATLANTIS_GH_TOKEN` or `ATLANTIS_GITLAB_TOKEN`
     or `ATLANTIS_BITBUCKET_TOKEN`
- `$SECRET` is the random key you used for the webhook secret.
   If you don't want this to be passed in as an argument for security reasons
    you can specify it in a config file
     (see [Configuration](/docs/server-configuration.html#environment-variables))
      or as an environment variable: `ATLANTIS_GH_WEBHOOK_SECRET` or `ATLANTIS_GITLAB_WEBHOOK_SECRET`
- `$REPO_WHITELIST` is which repos Atlantis can run on, ex.
 `github.com/runatlantis/*` or `github.enterprise.corp.com/*`.
  See [Repo Whitelist](server-configuration.html#repo-whitelist) for more details.

Atlantis is now running!
::: tip
We recommend running it under something like Systemd or Supervisord. That will
restart it in case of failure.
:::

## Next Steps
* To ensure Atlantis is running, load its UI. By default Atlantis runs on port `4141`.
* Now you're ready to add Webhooks to your repos. See [Configuring Webhooks](configuring-webhooks.html).
