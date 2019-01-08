#!/bin/bash

oc get csv --all-namespaces |grep clusterlogging.v0.0.1 |awk '{print $1}' > ns

while IFS='' read -r line || [[ -n "$line" ]]; do
    oc delete csv clusterlogging.v0.0.1 -n $line
done < ns

