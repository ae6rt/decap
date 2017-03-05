#!/bin/bash

set -ux

. common.rc

set +e

for i in ${APPLICATION_NAME}-build-metadata ${APPLICATION_NAME}-console-logs; do 
  aws --profile ${AWS_PROFILE} dynamodb delete-table --table-name $i
done

for i in ${APPLICATION_NAME}-db-base ${APPLICATION_NAME}-db-project-key ${APPLICATION_NAME}-locks-base ${APPLICATION_NAME}-locks-expires-index; do 
  ARN=$(aws --profile ${AWS_PROFILE} iam list-policies --scope Local  | jq --arg pname $i -r '.Policies[] | select(.PolicyName==$pname) | .Arn')
  aws --profile ${AWS_PROFILE} iam detach-user-policy --user-name ${APPLICATION_NAME} --policy-arn $ARN
  aws --profile ${AWS_PROFILE} iam delete-policy --policy-arn $ARN
done

