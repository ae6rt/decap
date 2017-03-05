#!/bin/bash

set -u

. common.rc

check

USER=$APPLICATION_NAME

ACCOUNT_ID=$(aws --profile $AWS_PROFILE iam get-user | jq -r ".User.UserId")

DB_POLICY=$(cat <<EOF
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
EOF
)

PROJECT_KEY_INDEX_POLICY=$(cat <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "dynamodb:*"
            ],
            "Resource": [
                "arn:aws:dynamodb:${AWS_REGION}:${ACCOUNT_ID}:table/${APPLICATION_NAME}-build-metadata/index/project-key-build-start-time-index"
            ]
        }
    ]
}
EOF
)

echo "===Creating policies for Dynamodb table"

BASE_POLICY=$(aws --profile $AWS_PROFILE iam create-policy --policy-name ${APPLICATION_NAME}-db-base --policy-document "$DB_POLICY" --description "Give r/w to $USER user on metadata table")

PROJECT_KEY_POLICY=$(aws --profile $AWS_PROFILE iam create-policy --policy-name ${APPLICATION_NAME}-db-project-key --policy-document "$PROJECT_KEY_INDEX_POLICY" --description "Give r/w to $USER user on project-key index")

for i in "$BASE_POLICY" "$PROJECT_KEY_POLICY"; do 
	aws --profile $AWS_PROFILE iam attach-user-policy --user-name $USER --policy-arn "$(echo "$i" | jq -r ".Policy.Arn")"
done
