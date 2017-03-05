#!/bin/bash

set -ux

. common.rc

set +e

for i in $(aws --profile ${AWS_PROFILE} iam list-access-keys --user-name ${APPLICATION_NAME} | jq -r ".AccessKeyMetadata[] | .AccessKeyId"); do
   aws --profile ${AWS_PROFILE} iam delete-access-key --user-name ${APPLICATION_NAME} --access-key $i
done

aws --profile ${AWS_PROFILE} iam delete-user --user-name ${APPLICATION_NAME}

