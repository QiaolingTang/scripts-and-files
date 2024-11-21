## Configure the following environment variables
```
export CLUSTER_NAME=$(oc get infrastructure cluster -o=jsonpath="{.status.infrastructureName}"  | sed 's/-[a-z0-9]\+$//')

export LOKISTACK_NAME="logging-loki-${CLUSTER_NAME}"

export LOKISTACK_NS="openshift-logging"

export REGION=$(oc get infrastructures.config.openshift.io cluster -ojsonpath={.status.platformStatus.aws.region})

export OIDC_ENDPOINT=$(oc get authentication.config.openshift.io cluster -o json | jq -r .spec.serviceAccountIssuer | sed  's|^https://||')

export AWS_ACCOUNT_ID=`aws sts get-caller-identity --query Account --output text`

export SCRATCH="/tmp/${CLUSTER_NAME}/lokistack-sts"

mkdir -p ${SCRATCH}

echo "Cluster: ${CLUSTER_NAME}, Region: ${REGION}, OIDC Endpoint: ${OIDC_ENDPOINT}, AWS Account ID: ${AWS_ACCOUNT_ID}"
```

## Prepare AWS Account
### Create an IAM Policy
```
POLICY_ARN=$(aws iam list-policies --query "Policies[?PolicyName=='LoggingLokiS3Bucket'].{ARN:Arn}" --output text)
if [[ -z "${POLICY_ARN}" ]]; then
cat << EOF > ${SCRATCH}/policy.json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:CreateBucket",
        "s3:DeleteBucket",
        "s3:PutBucketTagging",
        "s3:GetBucketTagging",
        "s3:PutBucketPublicAccessBlock",
        "s3:GetBucketPublicAccessBlock",
        "s3:PutEncryptionConfiguration",
        "s3:GetEncryptionConfiguration",
        "s3:PutLifecycleConfiguration",
        "s3:GetLifecycleConfiguration",
        "s3:GetBucketLocation",
        "s3:ListBucket",
        "s3:GetObject",
        "s3:PutObject",
        "s3:DeleteObject",
        "s3:ListBucketMultipartUploads",
        "s3:AbortMultipartUpload",
        "s3:ListMultipartUploadParts"
      ],
      "Resource": "arn:aws:s3:*:*:*"
    }
  ]
}
EOF
POLICY_ARN=$(aws iam create-policy --policy-name "LoggingLokiS3Bucket" \
--policy-document file:///${SCRATCH}/policy.json --query Policy.Arn --output text)
fi
echo ${POLICY_ARN}
```

### Create an IAM Role trust policy
```
cat <<EOF > ${SCRATCH}/trust-policy.json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::${AWS_ACCOUNT_ID}:oidc-provider/${OIDC_ENDPOINT}"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "${OIDC_ENDPOINT}:sub": [
            "system:serviceaccount:${LOKISTACK_NS}:${LOKISTACK_NAME}",
            "system:serviceaccount:${LOKISTACK_NS}:${LOKISTACK_NAME}-ruler"
          ]
        }
      }
    }
  ]
}
EOF
ROLE_ARN=$(aws iam create-role --role-name "${CLUSTER_NAME}-LoggingLokiS3Bucket" \
   --assume-role-policy-document file://${SCRATCH}/trust-policy.json \
   --query Role.Arn --output text)
echo ${ROLE_ARN}
```

### Attach the IAM Policy to the IAM Role
```
aws iam attach-role-policy --role-name "${CLUSTER_NAME}-LoggingLokiS3Bucket" \
--policy-arn ${POLICY_ARN}
```

## Create Bucket
```
BUCKETNAME="logging-loki-qitang"
SECRETNAME="logging-storage-secret"
STORAGECLASS=$(oc get sc -ojsonpath={.items[?(@.metadata.annotations.storageclass\\.kubernetes\\.io/is-default-class == \"true\")].metadata.name})

aws s3api create-bucket --bucket ${BUCKETNAME} --region us-east-2 --create-bucket-configuration LocationConstraint=us-east-2
```

## Create Secret for LokiStack
```
oc -n ${lokistack_ns} create secret generic ${SECRETNAME} \
  --from-literal=bucketnames="${BUCKETNAME}" \
  --from-literal=role_arn="${ROLE_ARN}" \
  --from-literal=region="us-east-2"
```

# Deploy Lokistack
```
cat << EOF | oc apply -f -
apiVersion: loki.grafana.com/v1
kind: LokiStack
metadata:
  name: ${LOKISTACK_NAME}
  namespace: openshift-logging
spec:
  managementState: Managed
  size: 1x.demo
  storage:
    secret:
      name: ${SECRETNAME}
      type: s3
  storageClassName: ${STORAGECLASS}
  tenants:
    mode: openshift-logging
  rules:
    enabled: true
    selector:
      matchLabels:
        openshift.io/cluster-monitoring: "true"
    namespaceSelector:
      matchLabels:
        openshift.io/cluster-monitoring: "true"
EOF
```

# Create clusterlogging
```
cat << EOF | oc create -f -
apiVersion: logging.openshift.io/v1
kind: ClusterLogging
metadata:
  name: instance
  namespace: openshift-logging
spec:
  collection:
    type: vector
  logStore:
    lokistack:
      name: ${LOKISTACK_NAME}
    type: lokistack
  managementState: Managed
  visualization:
    type: ocp-console
EOF
```

# Cleanup Resources
```
aws iam detach-role-policy --role-name "${CLUSTER_NAME}-LoggingLokiS3Bucket" --policy-arn "${POLICY_ARN}"
aws iam delete-role --role-name "${CLUSTER_NAME}-LoggingLokiS3Bucket"
aws iam delete-policy --policy-arn "${POLICY_ARN}"
```
