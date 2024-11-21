#!/bin/bash

set +x

STORAGETYPE="$1"
BUCKETNAME="${2:-logging-loki-qitang}"
SECRETNAME="${3:-logging-storage-secret}"
NAMESPACE="openshift-logging"
#STORAGECLASS=$(oc get sc -ojsonpath={.items[?(@.metadata.annotations.storageclass\\.kubernetes\\.io/is-default-class == \"true\")].metadata.name})


create_gcs_bucket() {
    gcloud alpha storage ls gs://${BUCKETNAME}
    if [ $? -eq 0 ]
    then
        echo "bucket ${BUCKETNAME} already exists"
    else
        echo "creating gs://${BUCKETNAME}"
        gcloud alpha storage buckets create gs://${BUCKETNAME}
    fi

    oc extract secret/gcp-credentials -n kube-system --confirm
    oc -n ${NAMESPACE} delete secret ${SECRETNAME} || :
    oc -n ${NAMESPACE} create secret generic ${SECRETNAME} --from-literal=bucketname="${BUCKETNAME}" --from-file="key.json"="service_account.json"
}

delete_gcs_bucket() {
    gcloud alpha storage rm --recursive gs://${BUCKETNAME}/
}

create_aws_s3_bucket() {
    aws s3api head-bucket --bucket ${BUCKETNAME}
    if [ $? -eq 0 ]
    then
        echo "bucket ${BUCKETNAME} already exists"
    else
        aws s3api create-bucket --bucket ${BUCKETNAME} --region us-east-2 --create-bucket-configuration LocationConstraint=us-east-2
    fi
    oc extract secret/aws-creds -n kube-system --confirm
    oc -n ${NAMESPACE} create secret generic ${SECRETNAME} --from-file=access_key_id=aws_access_key_id --from-file=access_key_secret=aws_secret_access_key --from-literal=region=us-east-2 --from-literal=bucketnames="${BUCKETNAME}" --from-literal=endpoint=https://s3.us-east-2.amazonaws.com
}

delete_aws_s3_bucket() {
    aws s3 rb s3://${BUCKETNAME} --force
}

deploy_loki() {
    case $STORAGETYPE in
        "gcs")
        create_gcs_bucket
        ;;
        "s3")
        create_aws_s3_bucket
        ;;
        "*")
        ;;
    esac

    #get default storage class
    default_sc=""
    scs=$(oc get sc -o jsonpath='{range .items[*]}{.metadata.name} ')
    for sc in "${scs[@]}"
    do
        if [[ $(oc get sc ${sc} -ojsonpath='{.metadata.annotations.storageclass\.kubernetes\.io/is-default-class}') == "true" ]]; then
            default_sc=${sc}
        fi
    done
    if [ -z $default_sc ]; then
        default_sc=$(oc get sc -ojsonpath='{.items[0].metadata.name}')
    fi
    api_version="v1"
    channel=$(oc get sub -n openshift-operators-redhat loki-operator -ojsonpath='{.spec.channel}')
    if [[ $channel == "candidate" ]]; then
        cat << EOF | oc apply -f -
apiVersion: loki.grafana.com/v1beta1
kind: LokiStack
metadata:
  name: lokistack-sample
  namespace: openshift-logging
spec:
  managementState: Managed
  size: 1x.demo
  storage:
    secret:
      name: ${SECRETNAME}
      type: ${STORAGETYPE}
  storageClassName: ${default_sc}
  tenants:
    mode: openshift-logging
EOF
    else
        cat << EOF | oc apply -f -
apiVersion: loki.grafana.com/v1
kind: LokiStack
metadata:
  name: lokistack-sample
  namespace: openshift-logging
spec:
  managementState: Managed
  size: 1x.demo
  storage:
    secret:
      name: ${SECRETNAME}
      type: ${STORAGETYPE}
  storageClassName: ${default_sc}
  tenants:
    mode: openshift-logging
  rules:
    enabled: true
    namespaceSelector:
      matchLabels:
        openshift.io/cluster-monitoring: "true"
    selector:
      matchLabels:
        openshift.io/cluster-monitoring: "true"
EOF
    fi
}

deploy_loki
