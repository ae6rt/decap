#!/bin/bash

. common.rc

check

USER_NAME="$APPLICATION_NAME"

echo "===Creating user AWS access-key"

result=$(aws --profile $AWS_PROFILE iam create-access-key --user-name $USER_NAME)

KEY=$(echo "$result"  | jq -j -r ".AccessKey.AccessKeyId")
SECRET=$(echo "$result" | jq -j -r ".AccessKey.SecretAccessKey")

cat <<EOF > aws.credentials
[decap]
aws_access_key_id = $KEY
aws_secret_access_key = $SECRET
EOF

KEY_BASE64=$(/bin/echo -n "$KEY" | openssl base64)
SECRET_BASE64=$(/bin/echo -n "$SECRET" | openssl base64)
AWS_REGION_BASE64=$(/bin/echo -n "$AWS_REGION" | openssl base64)

cat <<XXX > k8s-decap-secret.yaml
apiVersion: v1
data:
  aws-key: $KEY_BASE64
  aws-secret: $SECRET_BASE64
  region: $AWS_REGION_BASE64
kind: Secret
metadata:
     name: decap-aws-credentials
type: Opaque
XXX
