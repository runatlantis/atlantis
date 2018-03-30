# Atlantis in Kubernetes

TODO: Different options, statefulset recommended.

## Statefulset vs deployment

TODO

## Creating secrets

```
echo -n "yourtoken" > token
echo -n "yoursecret" > webhook-secret
kubectl create secret generic atlantis-github --from-file=token --from-file=webhook-secret
```

## DNS/SSL

TODO
