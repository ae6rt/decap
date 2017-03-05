#!/bin/bash

set -ux

. common.rc

set +e

for i in ${APPLICATION_NAME}-console-logs ${APPLICATION_NAME}-build-artifacts; do
   aws --profile ${AWS_PROFILE} s3 rm s3://$i --recursive
   aws --profile ${AWS_PROFILE} s3api delete-bucket --bucket $i
done

for i in ${APPLICATION_NAME}-s3-build-artifacts ${APPLICATION_NAME}-s3-console-logs; do 
  ARN=$(aws --profile ${AWS_PROFILE} iam list-policies --scope Local  | jq --arg pname $i -r '.Policies[] | select(.PolicyName==$pname) | .Arn')
  aws --profile ${AWS_PROFILE} iam detach-user-policy --user-name ${APPLICATION_NAME} --policy-arn $ARN
  aws --profile ${AWS_PROFILE} iam delete-policy --policy-arn $ARN
done

