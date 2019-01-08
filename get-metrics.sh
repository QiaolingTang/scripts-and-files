#!/bin/bash
curl -k -H "Authorization: Bearer `oc sa get-token prometheus-k8s -n openshift-monitoring`"   -H "Content-type: application/json" https://`oc get pods -n openshift-logging --selector component=elasticsearch -o jsonpath={.items[?\(@.status.phase==\"Running\"\)].metadata.name} | cut -d" " -f1 | xargs oc get pod -n openshift-logging -o jsonpath={.status.podIP}`:9200/_prometheus/metrics

curl -k -H "Authorization: Bearer `oc sa get-token prometheus-k8s -n openshift-monitoring`"   -H "Content-type: application/json" https://`oc get pods -n openshift-logging --selector component=fluentd -o jsonpath={.items[?\(@.status.phase==\"Running\"\)].metadata.name} | cut -d" " -f1 | xargs oc get pod -n openshift-logging -o jsonpath={.status.podIP}`:24231/metrics
