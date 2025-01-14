# Get token
```
token='xxxxx'
```

# Get loki route
```
lokistack_route=$(oc get route $(oc get lokistack -n openshift-logging|grep -v NAME | awk '{print $1}') -n openshift-logging -o json |jq '.spec.host' -r)
```
or
```
lokistack_route=$(oc get route -A -l app.kubernetes.io/managed-by=lokistack-controller -ojsonpath='{.items[0].spec.host}')
```

# Query Application logs
```
logcli -o raw --tls-skip-verify --bearer-token="${token}" --addr="https://${lokistack_route}/api/logs/v1/application" query '{log_type="application"}'
```

# Query Infrastructure logs
```
logcli -o raw --tls-skip-verify --bearer-token="${token}" --addr="https://${lokistack_route}/api/logs/v1/infrastructure" query '{log_type="infrastructure"}'
```

# Query Audit logs
```
logcli -o jsonl --tls-skip-verify --bearer-token="${token}" --addr="https://${lokistack_route}/api/logs/v1/audit" query '{log_type="audit"}'
```
