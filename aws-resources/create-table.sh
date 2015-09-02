#!/bin/bash

set -ux

KEY_SCHEMA=$(cat <<KS
[
            {
                "KeyType": "HASH", 
                "AttributeName": "buildID"
            }
]
KS
)

ATTRIBUTE_DEFINITIONS=$(cat <<ATTRS
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
ATTRS
)

GLOBAL_SECONDARY_INDEXES=$(cat <<GSI
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
GSI
)

aws --profile petrovic --region us-west-1 dynamodb create-table \
     --table-name fosse-build-metadata \
     --attribute-definitions "$(cat table-attrs.json)" \
     --key-schema "$(cat key-schema.json)" \
     --global-secondary-indexes "$(cat gsi.json)" \
     --provisioned-throughput ReadCapacityUnits=1,WriteCapacityUnits=1 
