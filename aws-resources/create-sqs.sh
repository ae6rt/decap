#!/bin/bash

set -u

. common.rc

check

QUEUE_NAME=${APPLICATION_NAME}-build-deferrals

echo "===Creating SQS deferral queue" 

aws --profile ${AWS_PROFILE} sqs create-queue --queue-name ${QUEUE_NAME}

