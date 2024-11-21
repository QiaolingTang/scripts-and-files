### Install splunk-operator
```
oc apply -f https://github.com/splunk/splunk-operator/releases/download/2.4.0/splunk-operator-cluster.yaml --server-side  --force-conflicts

oc project splunk-operator

oc adm policy add-scc-to-user nonroot -z splunk-operator-controller-manager
```

### Create splunk standalone
```
oc adm policy add-scc-to-user nonroot -z default

cat <<EOF | kubectl apply -n splunk-operator -f -
apiVersion: enterprise.splunk.com/v4
kind: Standalone
metadata:
  name: single
  finalizers:
  - enterprise.splunk.com/delete-pvc
EOF
```

### Expose route
```
oc expose svc/splunk-single-standalone-headless

oc get route splunk-single-standalone-headless -ojsonpath={.spec.host}

password=$(oc get secret splunk-single-standalone-secret-v1 -ojsonpath={.data.password} | base64 -d)
echo $password
```
Login to console with `admin:$password`


### Forward logs to splunk
```
oc create secret generic clf-splunk-secret --from-literal=hecToken=$(oc get secret splunk-single-standalone-secret-v1 -ojsonpath={.data.hec_token} | base64 -d)
oc create sa clf-splunk
oc adm policy add-cluster-role-to-user collect-application-logs -z clf-splunk
oc adm policy add-cluster-role-to-user collect-infrastructure-logs -z clf-splunk
oc adm policy add-cluster-role-to-user collect-audit-logs -z clf-splunk

cat << EOF | oc create -f -
  apiVersion: logging.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: clf-splunk
  spec:
    outputs:
    - name: splunk
      type: splunk
      url: https://splunk-single-standalone-service.splunk-operator.svc:8088
      secret:
        name: clf-splunk-secret
      tls:
        insecureSkipVerify: true
    pipelines:
    - name: pipeline-splunk
      inputRefs:
      - application
      - infrastructure
      - audit
      outputRefs:
      - splunk
    serviceAccountName: clf-splunk
EOF
```


Refs:
- https://splunk.github.io/splunk-operator/OpenShift.html
- https://splunk.github.io/splunk-operator/
- https://github.com/splunk/splunk-operator
