# Elasticsearch

## Installation

### Install from elasticsearch repo

- ES8
```
cat <<EOF >> /etc/yum.repos.d/es.repo
[elasticsearch]
name=Elasticsearch repository for 8.x packages
baseurl=https://artifacts.elastic.co/packages/8.x/yum
gpgcheck=1
gpgkey=https://artifacts.elastic.co/GPG-KEY-elasticsearch
enabled=1
autorefresh=1
type=rpm-md
EOF

sudo dnf install elasticsearch -y
```

- ES9
```
cat <<EOF >> /etc/yum.repos.d/es.repo
[elasticsearch]
name=Elasticsearch repository for 9.x packages
baseurl=https://artifacts.elastic.co/packages/9.x/yum
gpgcheck=1
gpgkey=https://artifacts.elastic.co/GPG-KEY-elasticsearch
enabled=1
autorefresh=1
type=rpm-md
EOF

sudo dnf install elasticsearch -y
```

### Install from rpm package
```
wget https://artifacts.elastic.co/downloads/elasticsearch/elasticsearch-8.18.1-x86_64.rpm
wget https://artifacts.elastic.co/downloads/elasticsearch/elasticsearch-8.18.1-x86_64.rpm.sha512
shasum -a 512 -c elasticsearch-8.18.1-x86_64.rpm.sha512
sudo rpm --install elasticsearch-8.18.1-x86_64.rpm
```

## Configuration

### Start service
```
sudo systemctl daemon-reload
sudo systemctl enable elasticsearch.service
sudo systemctl start elasticsearch.service
```

### Get password for user elastic
```
/usr/share/elasticsearch/bin/elasticsearch-reset-password -u elastic
```

### Connect to Elasticsearch
```
curl --cacert /etc/elasticsearch/certs/http_ca.crt -u elastic:$ELASTIC_PASSWORD https://localhost:9200

# in collector pod:
curl -u elastic:$ELASTIC_PASSWORD --cacert /var/run/ocp-collector/secrets/es-https/ca-bundle.crt 'https://xxxx.com:9200/_cat/indices?v' -vv
```

## Forward to ES
```
oc project openshift-logging
oc create sa logcollector
oc adm policy add-cluster-role-to-user collect-infrastructure-logs -z logcollector
oc adm policy add-cluster-role-to-user collect-audit-logs -z logcollector
oc adm policy add-cluster-role-to-user collect-application-logs -z logcollector
oc create secret generic es-https --from-file=ca-bundle.crt=http_ca.crt --from-literal=username=elastic --from-literal=password=$ELASTIC_PASSWORD

cat << EOF | oc create -f -
  apiVersion: "observability.openshift.io/v1"
  kind: ClusterLogForwarder
  metadata:
    name: es8
    namespace: openshift-logging
  spec:
    managementState: Managed
    outputs:
    - name: es8
      type: elasticsearch
      elasticsearch:
        authentication:
          password:
            key: password
            secretName: es-https
          username:
            key: username
            secretName: es-https
        index: "{.log_type||\"none-typed-logs\"}"
        url: https://xxxx.com:9200
        version: 8
      tls:
        ca:
          key: ca-bundle.crt
          secretName: es-https
    pipelines:
    - name: forward-to-external-es
      inputRefs:
      - infrastructure
      - audit
      outputRefs:
      - es8
    serviceAccount:
      name: logcollector
EOF
```


# Splunk

## Installation

### Install splunk-operator
```
oc apply -f https://github.com/splunk/splunk-operator/releases/download/2.8.0/splunk-operator-cluster.yaml --server-side

oc project splunk-operator


oc adm policy add-scc-to-user nonroot-v2 -z splunk-operator-controller-manager
oc adm policy add-scc-to-user nonroot-v2 -z default
```


### Create splunk standalone
```
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


## Forward logs to splunk
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

Or

```
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
        url: https://splunk-standalone-sample-standalone-service.splunk-operator.svc:8088
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

### Refs:
- https://splunk.github.io/splunk-operator/OpenShift.html
- https://splunk.github.io/splunk-operator/
- https://github.com/splunk/splunk-operator


# Rsyslog

Ref: https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/9/html/security_hardening/assembly_configuring-a-remote-logging-solution_security-hardening#configuring-a-server-for-remote-logging-over-tcp_assembly_configuring-a-remote-logging-solution
