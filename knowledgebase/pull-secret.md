# Allowing pods to reference images from other secured registries

```
oc create secret generic <secret_name> --from-file=.dockerconfigjson=<path-to-file> --type=kubernetes.io/dockerconfigjson

oc secrets link <sa_name> <secret_name> --for=pull
```

Ref: https://docs.openshift.com/rosa/openshift_images/managing_images/using-image-pull-secrets.html#images-allow-pods-to-reference-images-from-secure-registries_using-image-pull-secrets


# Update pull-secret in HCP

## Management Cluster

- Update secret/${hcp-cluster-name}-pull-secret in ns/clusters
- Update secret/pull-secret in ns/clusters-${hcp-cluster-name}

## Hosted Cluster

- Update secret/pull-secret in ns/openshift-config
