# Atlantis in Kubernetes

Atlantis can be deployed as a Kubernetes
[Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
or as a [Statefulset](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/).

## Statefulset vs deployment

For production it is recommended to deploy it as a statefulset
with a persistent disk. See [atlantis-persistent-storage.yaml](atlantis-persistent-storage.yaml)
for an example.

If you do not want persistent storage,
have a look at [atlantis-deployment.yaml](atlantis-deployment.yaml).
It configures a deployment with a single replica, and no persistent disk.

## Creating secrets

Add the Github token and the webhook secret as a Kubernetes secret.

```
echo -n "yourtoken" > token
echo -n "yoursecret" > webhook-secret
kubectl create secret generic atlantis-github --from-file=token --from-file=webhook-secret
```

## DNS/SSL

[atlantis-deployment-dns-ssl.yaml](atlantis-deployment-dns-ssl.yaml) shows an example of how to configure
Atlantis with SSL enabled. In addition, it adds a DNS entry
for the service.

Dependencies:
- [external-dns](https://github.com/kubernetes-incubator/external-dns)
- [kubernetes-letsencrypt](https://github.com/tazjin/kubernetes-letsencrypt)
