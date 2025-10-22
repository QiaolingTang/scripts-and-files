# Build Catalog based on FBC image

## Create catalog.yaml

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

## Create new catalog

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
echo $indexImage
podman build -t "$indexImage" -f "$name.Dockerfile" .
podman push "$indexImage"
```

## Create catalogsource

```
cat << EOF | oc apply -f -
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
  image: quay.io/logging/logging-operators:latest
EOF
```

# Create a Catalog of operators

## Create the Catalog hierarchy and Dockerfile for generating the image

```
mkdir -p logging-operators/cluster-logging-operator

mkdir logging-operators/loki-operator

opm generate dockerfile logging-operators
```

## Organize the Bundles into Channels

```
cat << EOF >> cluster-logging-operator-template.yaml
Schema: olm.semver
GenerateMajorChannels: true
GenerateMinorChannels: false
Stable:
  Bundles:
  - Image: quay.io/xxxx/cluster-logging-operator-bundle-v6-4@sha256:xxxx
EOF

cat << EOF >> loki-operator-template.yaml
Schema: olm.semver
GenerateMajorChannels: true
GenerateMinorChannels: false
Stable:
  Bundles:
  - Image: quay.io/xxx/loki-operator-bundle-v6-4@sha256:xxxx
EOF
```

## Generating the Catalog

```
opm alpha render-template semver -o yaml  < cluster-logging-operator-template.yaml > logging-operators/catalog.yaml

opm alpha render-template semver -o yaml  < loki-operator-template.yaml >> logging-operators/catalog.yaml

sed -i '' -e "s/stable-v6/stable-6.4/g" logging-operators/catalog.yaml
```


## Build and push the catalog image

```
podman build -t quay.io/logging/logging-operators:latest -f logging-operators.Dockerfile .

podman push quay.io/logging/logging-operators:latest
```

## Create catalogsource

```
cat << EOF | oc apply -f -
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: qe-app-registry
  namespace: openshift-marketplace
spec:
  sourceType: grpc
  grpcPodConfig:
    extractContent:
      cacheDir: /tmp/cache
      catalogDir: /configs
    memoryTarget: 30Mi
  image: quay.io/logging/logging-operators:latest
EOF
```


# Update a Catalog Image

```
opm render registry.redhat.io/redhat/redhat-operator-index:v4.20 -oyaml > index.yaml
mkdir logging-operators
opm generate dockerfile logging-operators -i registry.redhat.io/openshift4/ose-operator-registry-rhel9:v4.20

opm init cluster-logging --default-channel=stable-6.4 --output yaml > logging-operators/index.yaml
opm render quay.io/logging/cluster-logging-operator-bundle:6.4  -oyaml >> logging-operators/index.yaml

opm init loki-operator --default-channel=stable-6.4 --output yaml >> logging-operators/index.yaml
opm render quay.io/logging/loki-operator-bundle:6.4  -oyaml >> logging-operators/index.yaml
```

add a channel entry to logging-operators/index.yaml:

```
---
schema: olm.channel
package: cluster-logging
name: stable-6.4
entries:
  - name: cluster-logging.v6.4.0
    skipRange: '>=6.2.0-0 <6.4.0'
---
schema: olm.channel
package: loki-operator
name: stable-6.4
entries:
  - name: loki-operator.v6.4.0
    skipRange: '>=6.2.0-0 <6.4.0'
```

merge the context in `index.yaml` and `logging-operators/index.yaml` into `logging-operators/index.yaml`, then validate package:

```
# opm validate logging-operators
# echo $?
0
```

build image:

```
podman build . -f logging-operators.Dockerfile -t quay.io/logging/logging-operators:latest

podman push quay.io/logging/logging-operators:latest
```


# Refs:

- https://docs.redhat.com/en/documentation/openshift_container_platform/4.18/html/extensions/catalogs#olm-fb-catalogs-package-reqd-prop_fbc
- https://olm.operatorframework.io/docs/tasks/creating-a-catalog/
- https://docs.redhat.com/en/documentation/openshift_container_platform/4.19/html/operators/administrator-tasks#olm-filtering-fbc_olm-managing-custom-catalogs
