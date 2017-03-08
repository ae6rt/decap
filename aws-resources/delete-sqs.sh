#!/bin/bash

set -ux

. common.rc

set +e

aws --profile ${AWS_PROFILE} sqs delete-queue --queue-name ${APPLICATION_NAME}-build-deferrals 

for i in ${APPLICATION_NAME}-build-deferrals; do 
  ARN=$(aws --profile ${AWS_PROFILE} iam list-policies --scope Local  | jq --arg pname $i -r '.Policies[] | select(.PolicyName==$pname) | .Arn')
  aws --profile ${AWS_PROFILE} iam detach-user-policy --user-name ${APPLICATION_NAME} --policy-arn $ARN
  aws --profile ${AWS_PROFILE} iam delete-policy --policy-arn $ARN
done

