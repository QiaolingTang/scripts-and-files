```
aws s3api create-bucket --bucket qitang-test-logging-bucket --region us-east-2 --create-bucket-configuration LocationConstraint=us-east-2

oc create secret generic logging-s3-output --from-file=aws_access_key_id=aws_access_key_id --from-file=aws_secret_access_key=aws_secret_access_key

oc create sa s3-collector
oc adm policy add-cluster-role-to-user collect-application-logs -z s3-collector
oc adm policy add-cluster-role-to-user collect-infrastructure-logs -z s3-collector
oc adm policy add-cluster-role-to-user collect-audit-logs -z s3-collector

cat << EOF | oc apply -f -
apiVersion: observability.openshift.io/v1
kind: ClusterLogForwarder
metadata:
  name: clf-s3-output
spec:
  collector:
    networkPolicy:
      ruleSet: RestrictIngressEgress
  managementState: Managed
  outputs:
  - s3:
      authentication:
        awsAccessKey:
          keyId:
            key: aws_access_key_id
            secretName: logging-s3-output
          keySecret:
            key: aws_secret_access_key
            secretName: logging-s3-output
        type: awsAccessKey
      bucket: qitang-test-logging-bucket
      keyPrefix: qitang-test-proxy.{.log_type||"none-typed-logs"}
      region: us-east-2
      tuning:
        compression: none
        deliveryMode: AtMostOnce
        maxRetryDuration: 20
        maxWrite: 10M
        minRetryDuration: 5
    name: s3-output
    type: s3
  pipelines:
  - inputRefs:
    - infrastructure
    - audit
    - application
    name: to-s3
    outputRefs:
    - s3-output
  serviceAccount:
    name: s3-collector
EOF
```


# CloudWatch

```
oc create secret generic logging-cloudwatch-output --from-file=aws_access_key_id=aws_access_key_id --from-file=aws_secret_access_key=aws_secret_access_key

oc create sa cloudwatch-collector
oc adm policy add-cluster-role-to-user collect-application-logs -z cloudwatch-collector
oc adm policy add-cluster-role-to-user collect-infrastructure-logs -z cloudwatch-collector
oc adm policy add-cluster-role-to-user collect-audit-logs -z cloudwatch-collector

cat << EOF | oc apply -f -
apiVersion: observability.openshift.io/v1
kind: ClusterLogForwarder
metadata:
  name: clf-cloudwatch-output
spec:
  collector:
    networkPolicy:
      ruleSet: RestrictIngressEgress
  managementState: Managed
  outputs:
  - cloudwatch:
      authentication:
        awsAccessKey:
          keyId:
            key: aws_access_key_id
            secretName: logging-cloudwatch-output
          keySecret:
            key: aws_secret_access_key
            secretName: logging-cloudwatch-output
        type: awsAccessKey
      groupName: logging-qitang-ipv6.{.log_type||"none-typed-logs"}
      region: us-east-2
      url: https://logs.us-east-2.amazonaws.com
    name: cloudwatch
    type: cloudwatch
  pipelines:
  - inputRefs:
    - infrastructure
    - audit
    - application
    name: to-cloudwatch
    outputRefs:
    - cloudwatch
  serviceAccount:
    name: cloudwatch-collector
EOF
```
