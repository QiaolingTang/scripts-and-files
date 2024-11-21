# Check credentialsMode

```
oc get cloudcredentials cluster -o=jsonpath={.spec.credentialsMode}
```

```
oc get secret $secret_name -n kube-system -o jsonpath --template '{ .metadata.annotations }'
```
where `secret_name` is `aws-creds` for AWS or `gcp-credentials` for GCP.

AWS or GCP clusters that use `manual` mode only: To determine whether the cluster is configured to create and manage cloud credentials from outside of the cluster, run the following command:
```
oc get authentication cluster -o jsonpath --template='{ .spec.serviceAccountIssuer }'
```

Ref: https://docs.openshift.com/container-platform/4.13/authentication/managing_cloud_provider_credentials/about-cloud-credential-operator.html

# Create CredentialsRequest

When the cluster uses `mint` or `passthrough` mode, users can get cloud credentials with specific permissions by creating a `CredentialsRequest` CR. For example:
```
apiVersion: cloudcredential.openshift.io/v1
kind: CredentialsRequest
metadata:
  annotations:
    capability.openshift.io/name: Logging
    include.release.openshift.io/self-managed-high-availability: "true"
    include.release.openshift.io/single-node-developer: "true"
  name: cluster-logging-operator
  namespace: openshift-cloud-credential-operator
spec:
  providerSpec:
    apiVersion: cloudcredential.openshift.io/v1
    kind: AWSProviderSpec
    statementEntries:
    - action:
      - s3:*
      effect: Allow
      resource: "arn:aws:s3:::*"
  secretRef:
    name: aws-s3-credentials
    namespace: openshift-logging
  serviceAccountNames:
  - cluster-logging-operator
  - logcollector
```
With above file, there will be a `secret/aws-s3-credentials` created in `openshift-logging` project, and the secret contains the `aws_access_key_id` and `aws_secret_access_key` which can be used to manage s3 buckets. More actions can be found here: https://docs.aws.amazon.com/AmazonS3/latest/userguide/using-with-s3-actions.html.


Below CredentialsRequest contains all the requested permissions for s3 bucket:
```
$ oc get credentialsrequests.cloudcredential.openshift.io -n openshift-cloud-credential-operator   openshift-image-registry   -oyaml
apiVersion: cloudcredential.openshift.io/v1
kind: CredentialsRequest
metadata:
  annotations:
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
    include.release.openshift.io/single-node-developer: "true"
  labels:
    controller-tools.k8s.io: "1.0"
  name: $name
  namespace: openshift-cloud-credential-operator
spec:
  providerSpec:
    apiVersion: cloudcredential.openshift.io/v1
    kind: AWSProviderSpec
    statementEntries:
    - action:
      - s3:CreateBucket
      - s3:DeleteBucket
      - s3:PutBucketTagging
      - s3:GetBucketTagging
      - s3:PutBucketPublicAccessBlock
      - s3:GetBucketPublicAccessBlock
      - s3:PutEncryptionConfiguration
      - s3:GetEncryptionConfiguration
      - s3:PutLifecycleConfiguration
      - s3:GetLifecycleConfiguration
      - s3:GetBucketLocation
      - s3:ListBucket
      - s3:GetObject
      - s3:PutObject
      - s3:DeleteObject
      - s3:ListBucketMultipartUploads
      - s3:AbortMultipartUpload
      - s3:ListMultipartUploadParts
      effect: Allow
      resource: '*'
  secretRef:
    name: $secret-name
    namespace: $namespace
  serviceAccountNames:
  - $serviceaccount
```

# Forward Logs to CloudWatch on STS Cluster

## Create CredentialsRequest
```
cat << EOF >> cluster-logging-operator_01_cloudwatch_request_aws.yaml
apiVersion: cloudcredential.openshift.io/v1
kind: CredentialsRequest
metadata:
  name: cloudwatch-credentials-credrequest
  namespace: openshift-cloud-credential-operator
spec:
  providerSpec:
    apiVersion: cloudcredential.openshift.io/v1
    kind: AWSProviderSpec
    statementEntries:
      - action:
          - logs:PutLogEvents
          - logs:CreateLogGroup
          - logs:PutRetentionPolicy
          - logs:CreateLogStream
          - logs:DescribeLogGroups
          - logs:DescribeLogStreams
        effect: Allow
        resource: arn:aws:logs:*:*:*
  secretRef:
    name: cloudwatch-credentials
    namespace: openshift-logging
  serviceAccountNames:
    - logcollector
EOF

oc create -f cluster-logging-operator_01_cloudwatch_request_aws.yaml
```
Note: the CredentialsRequest yaml file must exist in the ${credentials-requests-dir}, otherwise running the ccoctl command won't get any files, and no roles can be created.

## Create Role in AWS
```
export OIDC_ENDPOINT=$(oc get authentication.config.openshift.io cluster -o json | jq -r .spec.serviceAccountIssuer | sed  's|^https://||')

export AWS_ACCOUNT_ID=`aws sts get-caller-identity --query Account --output text`

export CLUSTER_NAME=${OIDC_ENDPOINT%-oidc*}

export REGION=$(oc get infrastructures.config.openshift.io cluster -ojsonpath={.status.platformStatus.aws.region})


ccoctl aws create-iam-roles --name=${CLUSTER_NAME} --region=${REGION} --credentials-requests-dir=./ --identity-provider-arn=arn:aws:iam::${AWS_ACCOUNT_ID}:oidc-provider/${OIDC_ENDPOINT}
```


There will be a directory named `manifests` created:
```
# ls -l
total 4
drwx------. 2 root root  62 Jun  9 09:06 manifests

# ls manifests/ -l
total 4
-rw-------. 1 root root 374 Jun  9 09:06 openshift-logging-cloudwatch-credentials-credentials.yaml
```

## Apply the Secret
```
# cat manifests/openshift-logging-cloudwatch-credentials-credentials.yaml
apiVersion: v1
stringData:
  credentials: |-
    [default]
    sts_regional_endpoints = regional
    role_arn = arn:aws:iam::${AWS_ACCOUNT_ID}:role/${role_name}
    web_identity_token_file = /var/run/secrets/openshift/serviceaccount/token
kind: Secret
metadata:
  name: cloudwatch-credentials
  namespace: openshift-logging
type: Opaque

oc apply -f manifests/openshift-logging-cloudwatch-credentials-credentials.yaml
```

## Create CLF
```
apiVersion: "logging.openshift.io/v1"
kind: ClusterLogForwarder
metadata:
  name: instance
  namespace: openshift-logging
spec:
  outputs:
   - name: cw
     type: cloudwatch
     cloudwatch:
       groupBy: logType
       groupPrefix: ${prefix}
       region: ${REGION}
     secret:
        name: cloudwatch-credentials
  pipelines:
    - name: to-cloudwatch
      inputRefs:
        - infrastructure
        - audit
        - application
      outputRefs:
        - cw
```

Ref: https://docs.openshift.com/container-platform/4.13/logging/cluster-logging-external.html#cluster-logging-collector-log-forward-sts-cloudwatch_cluster-logging-external
