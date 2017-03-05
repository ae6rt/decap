#!/bin/bash

set -u

. common.rc

check

TABLE_NAME="${APPLICATION_NAME}-buildlocks"

KEY_SCHEMA=$(cat <<EOF
[
            {
                "KeyType": "HASH", 
                "AttributeName": "lockname"
            }
]
EOF
)

ATTRIBUTE_DEFINITIONS=$(cat <<EOF
[
            {
                "AttributeName": "lockname",
                "AttributeType": "S"
            },
            {
                "AttributeName": "expiresunixtime",
                "AttributeType": "N"
            }
]
EOF
)

GLOBAL_SECONDARY_INDEXES=$(cat <<EOF
[
{
                "IndexName": "expiresunixtime-index", 
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
                        "AttributeName": "lockname"
                    }, 
                    {
                        "KeyType": "RANGE", 
                        "AttributeName": "expiresunixtime"
                    }
                ] 
} 
]
EOF
)

echo "===Creating DynamoDb buildlocks table"

aws --profile $AWS_PROFILE dynamodb create-table \
     --table-name $TABLE_NAME \
     --attribute-definitions "$ATTRIBUTE_DEFINITIONS" \
     --key-schema "$KEY_SCHEMA" \
     --global-secondary-indexes "$GLOBAL_SECONDARY_INDEXES" \
     --provisioned-throughput ReadCapacityUnits=1,WriteCapacityUnits=1 >> aws.log
