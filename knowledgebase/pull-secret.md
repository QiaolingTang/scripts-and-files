# Allowing pods to reference images from other secured registries

```
oc create secret generic <secret_name> --from-file=.dockerconfigjson=<path-to-file> --type=kubernetes.io/dockerconfigjson

oc secrets link <sa_name> <secret_name> --for=pull
```

Ref: https://docs.openshift.com/rosa/openshift_images/managing_images/using-image-pull-secrets.html#images-allow-pods-to-reference-images-from-secure-registries_using-image-pull-secrets


# Update pull-secret in HCP

## Management Cluster

- Create new secret with the new credentials in `clusters` namespace:

```
oc create secret generic <cluster-name>-pull-secret-new -n clusters --from-file=.dockerconfigjson
```

- Update `hostedclusters.hypershift.openshift.io/<cluster-name>` in `clusters` namespace:

```
oc patch -n clusters hostedclusters.hypershift.openshift.io/<cluster-name> -p '{"spec": {"pullSecret": {"name": "<cluster-name>-pull-secret-new"}}}' --type=merge
```

- Wait for the reconciliation completed on the hostedcluster

