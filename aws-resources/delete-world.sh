#!/bin/sh

set -u

# Remove this exit statement when you want this script to work.  It is called as a safety against accidental invocation.
exit -1

sh delete-s3.sh
sh delete-dynamodb.sh
sh delete-user.sh
