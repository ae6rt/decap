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
