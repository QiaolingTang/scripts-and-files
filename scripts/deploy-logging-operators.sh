#/bin/bash

set -e

source="qe-app-registry"
channel="stable"
SOURCE=${1:-$source}
CHANNEL=${2:-$channel}

declare -a projects=(openshift-logging openshift-operators-redhat)
for i in "${projects[@]}"
do
  oc delete ns $i || true
done

oc delete crd alertingrules.loki.grafana.com  lokistacks.loki.grafana.com  recordingrules.loki.grafana.com  rulerconfigs.loki.grafana.com || true


# create namespace
cat << EOF | oc apply -f -
kind: Namespace
apiVersion: v1
metadata:
  name: openshift-logging
  labels:
    openshift.io/cluster-monitoring: "true"
---
kind: Namespace
apiVersion: v1
metadata:
  name: openshift-operators-redhat
  labels:
    openshift.io/cluster-monitoring: "true"
EOF


cat << EOF | oc apply -f -
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: openshift-logging
  namespace: openshift-logging
spec: {}
EOF

cat << EOF | oc apply -f -
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: openshift-operators-redhat
  namespace: openshift-operators-redhat
spec: {}
EOF


cat << EOF | oc apply -f -
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: cluster-logging
  namespace: openshift-logging
spec:
  channel: "$CHANNEL"
  installPlanApproval: Automatic
  name: cluster-logging
  source: $SOURCE
  sourceNamespace: openshift-marketplace
EOF

cat << EOF | oc apply -f -
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: loki-operator
  namespace: openshift-operators-redhat
spec:
  channel: "$CHANNEL"
  installPlanApproval: Automatic
  name: loki-operator
  source: $SOURCE
  sourceNamespace: openshift-marketplace
EOF
