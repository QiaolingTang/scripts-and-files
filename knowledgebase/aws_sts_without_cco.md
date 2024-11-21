# Forward to CloudWatch

## Configure the following environment variables
```
export CLUSTER_NAME=$(oc get infrastructure cluster -o=jsonpath="{.status.infrastructureName}"  | sed 's/-[a-z0-9]\+$//')

export REGION=$(oc get infrastructures.config.openshift.io cluster -ojsonpath={.status.platformStatus.aws.region})

export OIDC_ENDPOINT=$(oc get authentication.config.openshift.io cluster -o json | jq -r .spec.serviceAccountIssuer | sed  's|^https://||')

export AWS_ACCOUNT_ID=`aws sts get-caller-identity --query Account --output text`

export SCRATCH="/tmp/${CLUSTER_NAME}/clf-cloudwatch-sts"

mkdir -p ${SCRATCH}

echo "Cluster: ${CLUSTER_NAME}, Region: ${REGION}, OIDC Endpoint: ${OIDC_ENDPOINT}, AWS Account ID: ${AWS_ACCOUNT_ID}"
```

## Prepare AWS Account
### Create an IAM Policy
```
POLICY_ARN=$(aws iam list-policies --query "Policies[?PolicyName=='LoggingCloudWatch'].{ARN:Arn}" --output text)
if [[ -z "${POLICY_ARN}" ]]; then
cat << EOF > ${SCRATCH}/policy.json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:DescribeLogGroups",
        "logs:DescribeLogStreams",
        "logs:PutLogEvents",
        "logs:PutRetentionPolicy"
      ],
      "Resource": "arn:aws:logs:*:*:*"
    }
  ]
}
EOF
POLICY_ARN=$(aws iam create-policy --policy-name "LoggingCloudWatch" \
--policy-document file:///${SCRATCH}/policy.json --query Policy.Arn --output text)
fi
echo ${POLICY_ARN}
```

### Create an IAM Role trust policy
```
cat <<EOF > ${SCRATCH}/trust-policy.json
{
   "Version": "2012-10-17",
   "Statement": [{
     "Effect": "Allow",
     "Principal": {
       "Federated": "arn:aws:iam::${AWS_ACCOUNT_ID}:oidc-provider/${OIDC_ENDPOINT}"
     },
     "Action": "sts:AssumeRoleWithWebIdentity",
     "Condition": {
       "StringEquals": {
         "${OIDC_ENDPOINT}:sub": "system:serviceaccount:openshift-logging:logcollector"
       }
     }
   }]
}
EOF
ROLE_ARN=$(aws iam create-role --role-name "${CLUSTER_NAME}-LoggingCloudWatch" \
   --assume-role-policy-document file://${SCRATCH}/trust-policy.json \
   --query Role.Arn --output text)
echo ${ROLE_ARN}
```

### Attach the IAM Policy to the IAM Role
```
aws iam attach-role-policy --role-name "${CLUSTER_NAME}-LoggingCloudWatch" \
--policy-arn ${POLICY_ARN}
```

## Create Secret for Log Collector
```
cat << EOF | oc apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: cloudwatch-credentials
  namespace: openshift-logging
stringData:
  role_arn: $ROLE_ARN
EOF
```

## Create CLF
```
cat << EOF | oc apply -f -
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
          groupBy: namespaceName
          groupPrefix: ${CLUSTER_NAME}
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
EOF
```


# Cleanup Resources
```
aws iam detach-role-policy --role-name "${CLUSTER_NAME}-LoggingCloudWatch" --policy-arn "${POLICY_ARN}"
aws iam delete-role --role-name "${CLUSTER_NAME}-LoggingCloudWatch"
aws iam delete-policy --policy-arn "${POLICY_ARN}"
```

Ref: https://mobb.ninja/docs/rosa/clf-cloudwatch-sts/
