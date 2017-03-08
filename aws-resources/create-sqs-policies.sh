#!/bin/bash

set -u

. common.rc

check

USER=$APPLICATION_NAME

ACCOUNT_ID=$(aws --profile $AWS_PROFILE iam get-user | jq -r ".User.UserId")
QUEUE_NAME=${APPLICATION_NAME}-build-deferrals

QUEUE_POLICY=$(cat <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": [
                "sqs:*"
            ],
            "Effect": "Allow",
            "Resource": "arn:aws:sqs:us-west-1:${ACCOUNT_ID}:${QUEUE_NAME}"
        }
    ]
}
EOF
)

echo "===Creating policies for SQS Queue"

POLICY_ARN=$(aws --profile $AWS_PROFILE iam create-policy --policy-name ${QUEUE_NAME} --policy-document "$QUEUE_POLICY" --description "Give r/w to $USER user on SQS deferral queue")

aws --profile $AWS_PROFILE iam attach-user-policy --user-name $USER --policy-arn "$(echo "${POLICY_ARN}" | jq -r ".Policy.Arn")"
