#!/bin/bash

set -u

. common.rc

checkprofile

USER_NAME="fosse"

echo "===Creating user $USER_NAME"

aws --profile $AWS_PROFILE iam create-user --user-name $USER_NAME >> aws.log
