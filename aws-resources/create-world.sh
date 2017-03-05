#!/bin/bash

. common.rc

check

echo "===Creating AWS resource for $APPLICATION_NAME"

sh create-user.sh
sh create-access-secret.sh
sh create-buckets.sh
sh create-bucket-policies.sh
sh create-artifact-table.sh
sh create-artifact-table-policies.sh
sh create-buildlocks-table.sh
sh create-buildlocks-table-policies.sh
