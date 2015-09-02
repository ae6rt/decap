#!/bin/bash

set -u

. common.rc

checkprofile

echo "===Creating user policies for table and buckets"

USER=$APPLICATION_NAME

ACCOUNT_ID=$(aws --profile $AWS_PROFILE iam get-user --user-name $USER | jq -r ".User.UserId")

DB_POLICY_DOCUMENT=$(cat <<DBPOLICY
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "dynamodb:*"
            ],
            "Resource": [
                "arn:aws:dynamodb:${AWS_REGION}:${ACCOUNT_ID}:table/${APPLICATION_NAME}-build-metadata"
            ]
        }
    ]
}
DBPOLICY
)

aws --profile $AWS_PROFILE iam create-policy --policy-name ${APPLICATION_NAME}-db --policy-document "$DB_POLICY_DOCUMENT" --description "Give r/w to $USER user on metadata table"
