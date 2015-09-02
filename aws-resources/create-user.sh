#!/bin/bash

set -u

. common.rc

check

USER_NAME="$APPLICATION_NAME"

echo "===Creating user $USER_NAME"

aws --profile $AWS_PROFILE iam create-user --user-name $USER_NAME >> aws.log
