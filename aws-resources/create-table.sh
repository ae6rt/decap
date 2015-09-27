#!/bin/bash

set -u

. common.rc

check

TABLE_NAME="${APPLICATION_NAME}-build-metadata"

KEY_SCHEMA=$(cat <<EOF
[
            {
                "KeyType": "HASH", 
                "AttributeName": "build-id"
            }
]
EOF
)

ATTRIBUTE_DEFINITIONS=$(cat <<EOF
[
            {
                "AttributeName": "build-id",
                "AttributeType": "S"
            },
            {
                "AttributeName": "build-start-time",
                "AttributeType": "N"
            },
            {
                "AttributeName": "project-key",
                "AttributeType": "S"
            }
]
EOF
)

GLOBAL_SECONDARY_INDEXES=$(cat <<EOF
[
{
                "IndexName": "project-key-build-start-time-index", 
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
                        "AttributeName": "project-key"
                    }, 
                    {
                        "KeyType": "RANGE", 
                        "AttributeName": "build-start-time"
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
