#!/bin/bash

set -u

. common.rc

check

TABLE_NAME="${APPLICATION_NAME}-build-metadata"

KEY_SCHEMA=$(cat <<EOF
[
            {
                "KeyType": "HASH", 
                "AttributeName": "buildID"
            }
]
EOF
)

ATTRIBUTE_DEFINITIONS=$(cat <<EOF
[
            {
                "AttributeName": "buildID",
                "AttributeType": "S"
            },
            {
                "AttributeName": "buildTime",
                "AttributeType": "N"
            },
            {
                "AttributeName": "isBuilding",
                "AttributeType": "N"
            },
            {
                "AttributeName": "projectKey",
                "AttributeType": "S"
            }
]
EOF
)

GLOBAL_SECONDARY_INDEXES=$(cat <<EOF
[
{
                "IndexName": "projectKey-buildTime-index", 
                "Projection": {
                    "ProjectionType": "ALL"
                }, 
                "ProvisionedThroughput": {
                    "WriteCapacityUnits": 1, 
                    "ReadCapacityUnits": 1
                }, 
                "KeySchema": [
                    {
                        "KeyType": "HASH", 
                        "AttributeName": "projectKey"
                    }, 
                    {
                        "KeyType": "RANGE", 
                        "AttributeName": "buildTime"
                    }
                ] 
}, 
{
                "IndexName": "isBuilding-index", 
                "Projection": {
                    "ProjectionType": "ALL"
                }, 
                "ProvisionedThroughput": {
                    "WriteCapacityUnits": 2, 
                    "ReadCapacityUnits": 1
                }, 
                "KeySchema": [
                    {
                        "KeyType": "HASH", 
                        "AttributeName": "isBuilding"
                    }
                ]
}
]
EOF
)

echo "===Creating DynamoDb Table"

aws --profile $AWS_PROFILE dynamodb create-table \
     --table-name $TABLE_NAME \
     --attribute-definitions "$ATTRIBUTE_DEFINITIONS" \
     --key-schema "$KEY_SCHEMA" \
     --global-secondary-indexes "$GLOBAL_SECONDARY_INDEXES" \
     --provisioned-throughput ReadCapacityUnits=1,WriteCapacityUnits=1 > aws.log
