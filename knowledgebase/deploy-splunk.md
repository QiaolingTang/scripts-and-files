### Install splunk-operator
```
oc apply -f https://github.com/splunk/splunk-operator/releases/download/2.4.0/splunk-operator-cluster.yaml --server-side  --force-conflicts

oc project splunk-operator


oc adm policy add-scc-to-user nonroot-v2 -z splunk-operator-controller-manager
oc adm policy add-scc-to-user nonroot-v2 -z default
```

Might also need `oc annotate ns splunk-operator openshift.io/sa.scc.supplemental-groups="40000/10000" openshift.io/sa.scc.uid-range="40000/10000" --overwrite`

### Create splunk standalone
```
oc adm policy add-scc-to-user nonroot -z default

cat <<EOF | kubectl apply -n splunk-operator -f -
apiVersion: enterprise.splunk.com/v4
kind: Standalone
metadata:
  name: standalone-sample
  finalizers:
  - enterprise.splunk.com/delete-pvc
EOF
```

### Expose route
```
oc expose svc/splunk-standalone-sample-standalone-headless

oc get route splunk-standalone-sample-standalone-headless -ojsonpath={.spec.host}

password=$(oc get secret splunk-standalone-sample-standalone-secret-v1 -ojsonpath={.data.password} | base64 -d)
echo $password
```
Login to console with `admin:$password`


### Forward logs to splunk
```
oc create route passthrough --service splunk-standalone-sample-standalone-service --port 8088

splunk_host=$(oc get route splunk-standalone-sample-standalone-service -ojsonpath='{.spec.host}')

oc create secret generic clf-splunk-secret --from-literal=hecToken=$(oc get secret splunk-standalone-sample-standalone-secret-v1 -ojsonpath={.data.hec_token} | base64 -d)

oc create sa clf-splunk
oc adm policy add-cluster-role-to-user collect-application-logs -z clf-splunk
oc adm policy add-cluster-role-to-user collect-infrastructure-logs -z clf-splunk
oc adm policy add-cluster-role-to-user collect-audit-logs -z clf-splunk


cat << EOF | oc create -f -
  apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: clf-splunk
  spec:
    outputs:
    - name: splunk
      type: splunk
      splunk:
        url: https://${splunk_host}
        authentication:
          token:
            key: hecToken
            secretName: clf-splunk-secret
        index: main
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
    serviceAccount:
      name: clf-splunk
EOF
```


Refs:
- https://splunk.github.io/splunk-operator/OpenShift.html
- https://splunk.github.io/splunk-operator/
- https://github.com/splunk/splunk-operator
