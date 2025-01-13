# Set env vars
```
BUCKETNAME="logging-loki-qitang"
SECRETNAME="logging-storage-secret"
NAMESPACE="openshift-logging"
STORAGECLASS=$(oc get sc -ojsonpath="{.items[?(@.metadata.annotations.storageclass\\.kubernetes\\.io/is-default-class == \"true\")].metadata.name}")
LOKISTACK_NAME="logging-loki"
```

# Create bucket
### AWS
```
STORAGETYPE="s3"

aws s3api create-bucket --bucket ${BUCKETNAME} --region us-east-2 --create-bucket-configuration LocationConstraint=us-east-2

oc extract secret/aws-creds -n kube-system --confirm

oc -n ${NAMESPACE} create secret generic ${SECRETNAME} --from-file=access_key_id=aws_access_key_id --from-file=access_key_secret=aws_secret_access_key --from-literal=region=us-east-2 --from-literal=bucketnames="${BUCKETNAME}" --from-literal=endpoint=https://s3.us-east-2.amazonaws.com
```

### GCP
```
STORAGETYPE="gcs"

gcloud alpha storage buckets create gs://${BUCKETNAME}
oc extract secret/gcp-credentials -n kube-system --confirm
oc -n ${NAMESPACE} delete secret ${SECRETNAME} || :
oc -n ${NAMESPACE} create secret generic ${SECRETNAME} --from-literal=bucketname="${BUCKETNAME}" --from-file="key.json"="service_account.json"
```

# Create lokistack
```
cat << EOF | oc apply -f -
apiVersion: loki.grafana.com/v1
kind: LokiStack
metadata:
  name: ${LOKISTACK_NAME}
  namespace: openshift-logging
spec:
  managementState: Managed
  size: 1x.extra-small
  storage:
    secret:
      name: ${SECRETNAME}
      type: ${STORAGETYPE}
  storageClassName: ${STORAGECLASS}
  tenants:
    mode: openshift-logging
  rules:
    enabled: true
    selector:
      matchLabels:
        openshift.io/cluster-monitoring: "true"
    namespaceSelector:
      matchLabels:
        openshift.io/cluster-monitoring: "true"
EOF
```

```
cat << EOF | oc apply -f -
apiVersion: loki.grafana.com/v1
kind: LokiStack
metadata:
  name: ${LOKISTACK_NAME}
  namespace: openshift-logging
spec:
  managementState: Managed
  size: 1x.demo
  storage:
    schemas:
    - effectiveDate: "2025-01-01"
      version: v13
    secret:
      name: ${SECRETNAME}
      type: ${STORAGETYPE}
  storageClassName: ${STORAGECLASS}
  tenants:
    mode: openshift-logging
  rules:
    enabled: true
    selector:
      matchLabels:
        openshift.io/cluster-monitoring: "true"
    namespaceSelector:
      matchLabels:
        openshift.io/cluster-monitoring: "true"
EOF
```

```
cat << EOF | oc apply -f -
apiVersion: loki.grafana.com/v1
kind: LokiStack
metadata:
  name: ${LOKISTACK_NAME}
  namespace: openshift-logging
spec:
  managementState: Managed
  size: 1x.pico
  storage:
    schemas:
    - effectiveDate: "2025-01-01"
      version: v13
    secret:
      name: ${SECRETNAME}
      type: ${STORAGETYPE}
  storageClassName: ${STORAGECLASS}
  tenants:
    mode: openshift-logging
  rules:
    enabled: true
    selector:
      matchLabels:
        openshift.io/cluster-monitoring: "true"
    namespaceSelector:
      matchLabels:
        openshift.io/cluster-monitoring: "true"
EOF
```

# Create clusterlogging
```
cat << EOF | oc create -f -
apiVersion: logging.openshift.io/v1
kind: ClusterLogging
metadata:
  name: instance
  namespace: openshift-logging
spec:
  collection:
    type: vector
  logStore:
    lokistack:
      name: ${LOKISTACK_NAME}
    type: lokistack
  managementState: Managed
  visualization:
    type: ocp-console
EOF
```

```
cat << EOF | oc create -f -
apiVersion: logging.openshift.io/v1
kind: ClusterLogging
metadata:
  name: instance
  namespace: openshift-logging
spec:
  collection:
    type: vector
  logStore:
    lokistack:
      name: ${LOKISTACK_NAME}
    type: lokistack
  managementState: Managed
EOF
```


# Logging 6.x
```
oc -n openshift-logging create sa log-collector

oc -n openshift-logging adm policy add-cluster-role-to-user logging-collector-logs-writer -z log-collector

oc -n openshift-logging adm policy add-cluster-role-to-user collect-application-logs -z log-collector

oc -n openshift-logging adm policy add-cluster-role-to-user collect-infrastructure-logs -z log-collector

oc -n openshift-logging adm policy add-cluster-role-to-user collect-audit-logs -z log-collector
```

```
cat << EOF | oc create -f -
apiVersion: observability.openshift.io/v1
kind: ClusterLogForwarder
metadata:
  name: collector-lokistack
  namespace: openshift-logging
spec:
  managementState: Managed
  outputs:
  - lokiStack:
      authentication:
        token:
          from: serviceAccount
      target:
        name: ${LOKISTACK_NAME}
        namespace: ${NAMESPACE}
    name: lokistack
    tls:
      ca:
        configMapName: openshift-service-ca.crt
        key: service-ca.crt
    type: lokiStack
  pipelines:
  - inputRefs:
    - application
    - audit
    - infrastructure
    name: logs-to-loki
    outputRefs:
    - lokistack
  serviceAccount:
    name: log-collector
EOF
```
