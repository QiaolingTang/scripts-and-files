# Preparations

1. Have docker or podman installed
2. Have opm installed

# Build bundle

1. Download code from upstream
```
git clone git@github.com:grafana/loki.git
git clone git@github.com:openshift/cluster-logging-operator.git
```

2. Copy manifests to a directory
```
mkdir bundles
cp -r loki/operator/bundle/manifests loki-operator
cp -r cluster-logging-operator/bundle/manifests cluster-logging
```

3. Modify manifests if needed, for loki, the CSV version is v0.0.1, if you want to do upgrade test, you have to change it to v5.x

4. Build bundle image
```
export TAG="6.0"

opm alpha bundle build -b podman -c stable -e stable -d ./cluster-logging/ -p cluster-logging -t quay.io/logging/cluster-logging-operator-bundle:${TAG} --overwrite
podman push quay.io/logging/cluster-logging-operator-bundle:${TAG}

opm alpha bundle build  -b podman -c stable -e stable -d ./loki-operator/ -p loki-operator -t quay.io/logging/loki-operator-bundle:${TAG} --overwrite
podman push quay.io/logging/loki-operator-bundle:${TAG}
```

Note: run `opm alpha bundle build --help` to check it's usage


# Build index image
We have 2 types of catalogs: file-based catalog and SQLite-based catalog.  Start from OCP 4.11, only the file-based catalog is supported.

## SQLite-based catalog
```
opm index add -b quay.io/logging/cluster-logging-operator-bundle:${TAG},quay.io/logging/loki-operator-bundle:${TAG} -t quay.io/logging/logging-index:${TAG} -c podman

podman push quay.io/logging/logging-index:${TAG}
```

## File-based catalog

### Build from existing index
```
mkdir catalog-test
opm render quay.io/logging/logging-index:${TAG} -o yaml > catalog-test/index.yaml
opm generate dockerfile  catalog-test
podman build . -f catalog-test.Dockerfile -t quay.io/logging/logging-index:${TAG}
podman push quay.io/logging/logging-index:${TAG}
```

### Build from bundles
```
mkdir logging-index
opm generate dockerfile logging-index
touch logging-index/index.yaml

opm init cluster-logging --default-channel=stable --output yaml >> logging-index/index.yaml
echo "---
entries:
- name: cluster-logging.v6.0.0
  skipRange: '>=5.8.0-0 <6.0.0'
name: stable
package: cluster-logging
schema: olm.channel" >> logging-index/index.yaml
opm render quay.io/logging/cluster-logging-operator-bundle:${TAG} --output=yaml >> logging-index/index.yaml

opm init loki-operator --default-channel=stable --output yaml >> logging-index/index.yaml
echo "---
entries:
- name: loki-operator.v6.0.0
  skipRange: '>=5.8.0-0 <6.0.0'
name: stable
package: loki-operator
schema: olm.channel" >> logging-index/index.yaml
opm render quay.io/logging/loki-operator-bundle:${TAG} --output=yaml >> logging-index/index.yaml

opm validate logging-index
podman build . -f logging-index.Dockerfile -t quay.io/logging/logging-index:${TAG}
```
