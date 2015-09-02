#!/bin/bash

set -u

. common.rc

checkprofile

BUCKET_NAME_PREFIX=$APPLICATION_NAME
REGION="us-west-1"

echo "===Creating buckets" >> aws.log

for i in console-logs build-artifacts; do 
	aws --profile $AWS_PROFILE s3api create-bucket --bucket ${BUCKET_NAME_PREFIX}-$i --create-bucket-configuration LocationConstraint=$REGION
done
