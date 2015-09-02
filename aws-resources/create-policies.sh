#!/bin/bash

set -u

. common.rc

check

echo "===Creating user policies for table and buckets"

USER=$APPLICATION_NAME

ACCOUNT_ID=$(aws --profile $AWS_PROFILE iam get-user | jq -r ".User.UserId")

DB_POLICY=$(cat <<DBPOLICY
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

PROJECT_KEY_INDEX_POLICY=$(cat <<PKEY
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "dynamodb:*"
            ],
            "Resource": [
                "arn:aws:dynamodb:${AWS_REGION}:${ACCOUNT_ID}:table/${APPLICATION_NAME}-build-metadata/index/projectKey-buildTime-index"
            ]
        }
    ]
}
PKEY
)

IS_BUILDING__INDEX_POLICY=$(cat <<ISBUILDING
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "dynamodb:*"
            ],
            "Resource": [
                "arn:aws:dynamodb:${AWS_REGION}:${ACCOUNT_ID}:table/${APPLICATION_NAME}-build-metadata/index/isBuilding-index"
            ]
        }
    ]
}
ISBUILDING
)

BASE_POLICY=$(aws --profile $AWS_PROFILE iam create-policy --policy-name ${APPLICATION_NAME}-db-base --policy-document "$DB_POLICY" --description "Give r/w to $USER user on metadata table")

PROJECT_KEY_POLICY=$(aws --profile $AWS_PROFILE iam create-policy --policy-name ${APPLICATION_NAME}-db-projectKey --policy-document "$PROJECT_KEY_INDEX_POLICY" --description "Give r/w to $USER user on projectKey index")

IS_BUILDING_POLICY=$(aws --profile $AWS_PROFILE iam create-policy --policy-name ${APPLICATION_NAME}-db-isBuilding --policy-document "$IS_BUILDING__INDEX_POLICY" --description "Give r/w to $USER user on isBuilding index")

for i in "$BASE_POLICY" "$PROJECT_KEY_POLICY" "$IS_BUILDING_POLICY"; do 
	aws --profile $AWS_PROFILE iam attach-user-policy --user-name $USER --policy-arn "$(echo "$i" | jq -r ".Policy.Arn")"
done
