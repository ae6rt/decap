#!/bin/bash

. common.rc

check

sh create-user.sh
sh create-buckets.sh
sh create-bucket-policies.sh
sh create-table.sh
sh create-table-policies.sh
