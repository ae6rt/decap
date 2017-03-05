#!/bin/bash

set -u

. common.rc

check

USER=$APPLICATION_NAME

ACCOUNT_ID=$(aws --profile $AWS_PROFILE iam get-user | jq -r ".User.UserId")

LOCKS_POLICY=$(cat <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "dynamodb:*"
            ],
            "Resource": [
                "arn:aws:dynamodb:${AWS_REGION}:${ACCOUNT_ID}:table/${APPLICATION_NAME}-buildlocks"
            ]
        }
    ]
}
EOF
)

INDEX_POLICY=$(cat <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "dynamodb:*"
            ],
            "Resource": [
                "arn:aws:dynamodb:${AWS_REGION}:${ACCOUNT_ID}:table/${APPLICATION_NAME}-buildlocks/index/expiresunixtime-index"
            ]
        }
    ]
}
EOF
)

echo "===Creating policies for Dynamodb buildlocks table"

BASE_POLICY=$(aws --profile $AWS_PROFILE iam create-policy --policy-name ${APPLICATION_NAME}-locks-base --policy-document "$LOCKS_POLICY" --description "Give r/w to $USER user on buildlocks table")

KEY_POLICY=$(aws --profile $AWS_PROFILE iam create-policy --policy-name ${APPLICATION_NAME}-locks-expires-index --policy-document "$INDEX_POLICY" --description "Give r/w to $USER user on expiresunixtime index")

for i in "$BASE_POLICY" "$KEY_POLICY"; do 
	aws --profile $AWS_PROFILE iam attach-user-policy --user-name $USER --policy-arn "$(echo "$i" | jq -r ".Policy.Arn")"
done
