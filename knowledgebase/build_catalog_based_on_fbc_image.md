# Create catalog.yaml

```
cat << EOF > catalog.yaml
name: logging-operators
repo: quay.io/logging/logging-operators
tag: latest
references:
- name: loki-operator
  image: quay.io/redhat-user-workloads/obs-logging-tenant/loki-operator-fbc-v4-18@sha256:xxx
- name: cluster-logging-operator
  image: quay.io/redhat-user-workloads/obs-logging-tenant/cluster-logging-fbc-v4-18@sha256:xxx
EOF
```

# Create new catalog

```
name=$(yq eval '.name' catalog.yaml)
mkdir "$name"
yq eval '.name + "/" + .references[].name' catalog.yaml | xargs mkdir
for l in $(yq e '.name as $catalog | .references[] | .image + "|" + $catalog + "/" + .name + "/index.yaml"' catalog.yaml); do
  image=$(echo $l | cut -d'|' -f1)
  file=$(echo $l | cut -d'|' -f2)
  opm render "$image" > "$file"
done
opm generate dockerfile "$name"
indexImage=$(yq eval '.repo + ":" + .tag' catalog.yaml)
podman build -t "$indexImage" -f "$name.Dockerfile" .
podman push "$indexImage"
```

# Create catalogsource

```
cat << EOF | oc create -f -
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: logging-operators
  namespace: openshift-marketplace
spec:
  sourceType: grpc
  grpcPodConfig:
    extractContent:
      cacheDir: /tmp/cache
      catalogDir: /configs
    memoryTarget: 30Mi
  image: ${indexImage}
EOF
```



# Refs:

- https://docs.redhat.com/en/documentation/openshift_container_platform/4.18/html/extensions/catalogs#olm-fb-catalogs-package-reqd-prop_fbc
